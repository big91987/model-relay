package relay

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"
)

type routeRuntime struct {
	Fingerprint string
	Current     map[int64]int
}
type routeHealth struct {
	Sequence          int64
	Cooldown          time.Time
	State             string
	Observed          time.Time
	CandidateRevision int64
}
type routeSnapshot struct {
	Model          RouteModel
	Upstreams      map[int64]Upstream
	ConfigRevision int64
	Epoch          uint64
	Runtime        *routeRuntime
}

type Attempt struct {
	Ordinal            int     `json:"ordinal"`
	UpstreamID         int64   `json:"upstream_id"`
	UpstreamName       string  `json:"upstream_name"`
	RemoteModel        string  `json:"remote_model"`
	UpstreamRevision   int64   `json:"upstream_revision"`
	UpstreamExecution  int64   `json:"upstream_execution_revision"`
	CandidateExecution int64   `json:"candidate_execution_revision"`
	Started            string  `json:"started_at"`
	Finished           *string `json:"finished_at"`
	Duration           *int64  `json:"duration_ms"`
	State              string  `json:"state"`
	Code               string  `json:"error_code"`
	Status             *int    `json:"upstream_status"`
	Usage
	Sequence int64 `json:"-"`
}
type RouteDecision struct {
	Round      int     `json:"selection_round"`
	UpstreamID int64   `json:"upstream_id"`
	Enabled    bool    `json:"snapshot_enabled"`
	Priority   int     `json:"snapshot_priority"`
	Weight     int     `json:"snapshot_weight"`
	Reason     string  `json:"reason"`
	Observed   string  `json:"observed_at"`
	Cooldown   *string `json:"cooldown_until"`
}
type UsagePart struct {
	KnownSum        int64  `json:"known_sum"`
	UnknownAttempts int    `json:"unknown_attempts"`
	Completeness    string `json:"completeness"`
	Total           *int64 `json:"total"`
}

func usageParts(attempts []Attempt, complete bool) map[string]UsagePart {
	parts := map[string]UsagePart{}
	for _, field := range []string{"prompt_tokens", "completion_tokens", "total_tokens"} {
		part := UsagePart{Completeness: "complete"}
		overflow := false
		for _, attempt := range attempts {
			value := attempt.Prompt
			if field == "completion_tokens" {
				value = attempt.Completion
			}
			if field == "total_tokens" {
				value = attempt.Total
			}
			if value == nil {
				part.UnknownAttempts++
				continue
			}
			if *value < 0 || *value > math.MaxInt64-part.KnownSum {
				overflow = true
				continue
			}
			part.KnownSum += *value
		}
		switch {
		case len(attempts) == 0:
			part.Completeness = "not_applicable"
		case overflow:
			part.Completeness = "overflow"
		case !complete:
			part.Completeness = "incomplete"
		case part.UnknownAttempts > 0:
			part.Completeness = "unknown"
		default:
			sum := part.KnownSum
			part.Total = &sum
		}
		parts[field] = part
	}
	return parts
}

func readAttempts(query materialQuery, id string) ([]Attempt, error) {
	rows, err := query.Query("SELECT ordinal,upstream_id,upstream_name,remote_model,upstream_revision,upstream_execution_revision,candidate_execution_revision,started_at,finished_at,duration_ms,state,error_code,upstream_status,prompt_tokens,completion_tokens,total_tokens FROM request_attempts WHERE request_id=? ORDER BY ordinal", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Attempt{}
	for rows.Next() {
		var value Attempt
		if err = rows.Scan(&value.Ordinal, &value.UpstreamID, &value.UpstreamName, &value.RemoteModel, &value.UpstreamRevision, &value.UpstreamExecution, &value.CandidateExecution, &value.Started, &value.Finished, &value.Duration, &value.State, &value.Code, &value.Status, &value.Prompt, &value.Completion, &value.Total); err != nil {
			return nil, err
		}
		list = append(list, value)
	}
	return list, rows.Err()
}
func aggregateUsage(transaction *sql.Tx, id string) (Usage, bool, error) {
	attempts, err := readAttempts(transaction, id)
	if err != nil {
		return Usage{}, false, err
	}
	complete := true
	for _, attempt := range attempts {
		if attempt.Finished == nil {
			complete = false
		}
	}
	parts := usageParts(attempts, complete)
	return Usage{Prompt: parts["prompt_tokens"].Total, Completion: parts["completion_tokens"].Total, Total: parts["total_tokens"].Total}, complete, nil
}

func (server *Server) snapshot(alias string) (routeSnapshot, error) {
	model, err := server.Store.routeModel(alias)
	if err != nil {
		return routeSnapshot{}, err
	}
	snapshot := routeSnapshot{Model: model, Upstreams: map[int64]Upstream{}, Epoch: server.configEpoch}
	if err = server.Store.DB.QueryRow("SELECT config_revision FROM meta").Scan(&snapshot.ConfigRevision); err != nil {
		return snapshot, err
	}
	for _, candidate := range model.Candidates {
		value, err := server.Store.upstreamByID(candidate.UpstreamID)
		if err != nil {
			return snapshot, err
		}
		snapshot.Upstreams[candidate.UpstreamID] = value
	}
	key := fmt.Sprintf("%s:%d:%d", alias, snapshot.ConfigRevision, model.Revision)
	if server.routing[key] == nil {
		server.routing[key] = &routeRuntime{Current: map[int64]int{}}
	}
	snapshot.Runtime = server.routing[key]
	return snapshot, nil
}

func healthKey(alias string, candidate Candidate, upstream Upstream) string {
	return fmt.Sprintf("%s:%d:%d:%s", alias, upstream.ID, upstream.ExecutionRevision, candidate.RemoteModel)
}

func (server *Server) beginAttempt(ctx context.Context, request *http.Request, record *Record, snapshot routeSnapshot, tried map[int64]bool, round int) (*Candidate, *Attempt, string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.Store.advanceClock() != nil {
		return nil, nil, "storage_unavailable"
	}
	if ctx.Err() != nil {
		return nil, nil, "request_timeout"
	}
	if server.degraded.Load() {
		return nil, nil, "storage_unavailable"
	}
	if _, ok := server.appID(request); !ok {
		return nil, nil, "unauthorized"
	}
	if server.Store.Schema == 3 {
		if allowed, code := server.Store.authorizeModel(record.KeyID, record.Alias); !allowed {
			return nil, nil, code
		}
	}
	current, err := server.Store.routeModel(snapshot.Model.Alias)
	if err != nil {
		return nil, nil, "storage_unavailable"
	}
	now := time.Now()
	decisions := []RouteDecision{}
	eligible := []Candidate{}
	priority := -1
	for _, candidate := range snapshot.Model.Candidates {
		upstream := snapshot.Upstreams[candidate.UpstreamID]
		decision := RouteDecision{Round: round, UpstreamID: candidate.UpstreamID, Enabled: candidate.Enabled, Priority: candidate.Priority, Weight: candidate.Weight, Observed: now.UTC().Format(time.RFC3339Nano)}
		latest, err := server.Store.upstreamByID(candidate.UpstreamID)
		if err != nil {
			return nil, nil, "storage_unavailable"
		}
		present := false
		for _, value := range current.Candidates {
			if value.UpstreamID == candidate.UpstreamID && value.Enabled {
				present = true
			}
		}
		health := server.health[healthKey(snapshot.Model.Alias, candidate, upstream)]
		switch {
		case !current.Enabled || !snapshot.Model.Enabled:
			decision.Reason = "disabled_model"
		case !upstream.Enabled || !latest.Enabled:
			decision.Reason = "disabled_upstream"
		case !candidate.Enabled || !present:
			decision.Reason = "disabled_candidate"
		case !upstream.CredentialConfigured:
			decision.Reason = "credential_missing"
		case health != nil && health.Cooldown.After(now):
			decision.Reason = "cooling"
			until := health.Cooldown.UTC().Format(time.RFC3339Nano)
			decision.Cooldown = &until
		case tried[candidate.UpstreamID]:
			decision.Reason = "already_attempted"
		default:
			eligible = append(eligible, candidate)
			if candidate.Priority > priority {
				priority = candidate.Priority
			}
		}
		decisions = append(decisions, decision)
	}
	group := []Candidate{}
	for _, candidate := range eligible {
		if candidate.Priority == priority {
			group = append(group, candidate)
		}
	}
	sort.Slice(group, func(left, right int) bool { return group[left].UpstreamID < group[right].UpstreamID })
	var selected *Candidate
	fingerprint := ""
	for _, candidate := range group {
		fingerprint += fmt.Sprintf("%d/%d/%d/%d;", candidate.UpstreamID, candidate.Priority, candidate.Weight, candidate.ExecutionRevision)
	}
	if fingerprint != snapshot.Runtime.Fingerprint {
		snapshot.Runtime.Fingerprint = fingerprint
		snapshot.Runtime.Current = map[int64]int{}
	}
	total := 0
	for _, candidate := range group {
		snapshot.Runtime.Current[candidate.UpstreamID] += candidate.Weight
		total += candidate.Weight
		if selected == nil || snapshot.Runtime.Current[candidate.UpstreamID] > snapshot.Runtime.Current[selected.UpstreamID] {
			copy := candidate
			selected = &copy
		}
	}
	if selected != nil {
		snapshot.Runtime.Current[selected.UpstreamID] -= total
	}
	for index := range decisions {
		if decisions[index].Reason != "" {
			continue
		}
		switch {
		case selected != nil && decisions[index].UpstreamID == selected.UpstreamID:
			decisions[index].Reason = "selected"
		case decisions[index].Priority < priority:
			decisions[index].Reason = "lower_priority"
		default:
			decisions[index].Reason = "not_selected_same_priority"
		}
	}
	cleanup, release := context.WithTimeout(context.Background(), 2*time.Second)
	defer release()
	transaction, err := server.Store.DB.BeginTx(cleanup, nil)
	if err != nil {
		return nil, nil, "storage_unavailable"
	}
	defer transaction.Rollback()
	for _, decision := range decisions {
		if _, err = transaction.ExecContext(cleanup, "INSERT INTO request_route_decisions VALUES(?,?,?,?,?,?,?,?,?)", record.ID, round, decision.UpstreamID, decision.Enabled, decision.Priority, decision.Weight, decision.Reason, decision.Observed, decision.Cooldown); err != nil {
			return nil, nil, "storage_unavailable"
		}
	}
	var attempt *Attempt
	if selected != nil {
		upstream := snapshot.Upstreams[selected.UpstreamID]
		server.dispatchSequence++
		attempt = &Attempt{Ordinal: len(tried) + 1, UpstreamID: upstream.ID, UpstreamName: upstream.Name, RemoteModel: selected.RemoteModel, UpstreamRevision: upstream.Revision, UpstreamExecution: upstream.ExecutionRevision, CandidateExecution: selected.ExecutionRevision, Started: now.UTC().Format(time.RFC3339Nano), State: "dispatching", Sequence: server.dispatchSequence}
		if _, err = transaction.ExecContext(cleanup, "INSERT INTO request_attempts(request_id,ordinal,upstream_id,upstream_name,remote_model,upstream_revision,upstream_execution_revision,candidate_execution_revision,started_at,state,error_code) VALUES(?,?,?,?,?,?,?,?,?,'dispatching','')", record.ID, attempt.Ordinal, attempt.UpstreamID, attempt.UpstreamName, attempt.RemoteModel, attempt.UpstreamRevision, attempt.UpstreamExecution, attempt.CandidateExecution, attempt.Started); err != nil {
			return nil, nil, "storage_unavailable"
		}
	}
	if err = transaction.Commit(); err != nil {
		return nil, nil, "storage_unavailable"
	}
	return selected, attempt, ""
}

func (store *Store) finishAttempt(id string, attempt *Attempt) error {
	cleanup, release := context.WithTimeout(context.Background(), 2*time.Second)
	defer release()
	query := "UPDATE request_attempts SET finished_at=?,duration_ms=?,state=?,error_code=?,upstream_status=?,prompt_tokens=?,completion_tokens=?,total_tokens=? WHERE request_id=? AND ordinal=?"
	if store.Schema == 3 {
		query += " AND finished_at IS NULL AND EXISTS(SELECT 1 FROM requests WHERE id=request_id AND state='started')"
	}
	_, err := store.DB.ExecContext(cleanup, query, attempt.Finished, attempt.Duration, attempt.State, attempt.Code, attempt.Status, attempt.Prompt, attempt.Completion, attempt.Total, id, attempt.Ordinal)
	return err
}

func (server *Server) observe(snapshot routeSnapshot, candidate Candidate, attempt Attempt, temporary bool) {
	server.mu.Lock()
	defer server.mu.Unlock()
	current, err := server.Store.routeModel(snapshot.Model.Alias)
	if err != nil {
		return
	}
	upstream, err := server.Store.upstreamByID(candidate.UpstreamID)
	if err != nil || upstream.ExecutionRevision != attempt.UpstreamExecution {
		return
	}
	valid := false
	for _, value := range current.Candidates {
		if value.UpstreamID == candidate.UpstreamID && value.RemoteModel == candidate.RemoteModel && value.ExecutionRevision == candidate.ExecutionRevision {
			valid = true
		}
	}
	if !valid {
		return
	}
	key := healthKey(current.Alias, candidate, upstream)
	previous := server.health[key]
	if previous != nil && previous.Sequence > attempt.Sequence {
		return
	}
	health := &routeHealth{Sequence: attempt.Sequence, State: attempt.State, Observed: time.Now(), CandidateRevision: candidate.ExecutionRevision}
	if temporary {
		health.Cooldown = health.Observed.Add(server.Config.Cooldown)
	}
	server.health[key] = health
}

func temporaryStatus(status int) bool {
	return status == 408 || status == 429 || status == 500 || status == 502 || status == 503 || status == 504
}

type attemptOutput struct {
	writer      http.ResponseWriter
	parent      context.Context
	child       context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	timer       *time.Timer
	timedOut    bool
	committed   bool
	committedAt string
	header      http.Header
}

func newAttemptOutput(writer http.ResponseWriter, parent context.Context, duration time.Duration) *attemptOutput {
	child, cancel := context.WithCancel(parent)
	output := &attemptOutput{writer: writer, parent: parent, child: child, cancel: cancel, header: make(http.Header)}
	output.timer = time.AfterFunc(duration, func() {
		output.mu.Lock()
		defer output.mu.Unlock()
		if !output.committed {
			output.timedOut = true
			cancel()
		}
	})
	return output
}
func (output *attemptOutput) Header() http.Header    { return output.header }
func (output *attemptOutput) WriteHeader(status int) {}
func (output *attemptOutput) Write(data []byte) (int, error) {
	output.mu.Lock()
	defer output.mu.Unlock()
	if output.header.Get("Content-Type") != "text/event-stream" {
		return len(data), nil
	}
	if output.parent.Err() != nil {
		return 0, output.parent.Err()
	}
	if output.timedOut {
		return 0, context.DeadlineExceeded
	}
	if !output.committed {
		if !output.timer.Stop() {
			return 0, context.DeadlineExceeded
		}
		output.committed = true
		output.committedAt = time.Now().UTC().Format(time.RFC3339Nano)
		output.writer.Header().Set("Content-Type", "text/event-stream")
	}
	return output.writer.Write(data)
}
func (output *attemptOutput) Flush() {
	if flusher, ok := output.writer.(http.Flusher); ok && output.isCommitted() {
		flusher.Flush()
	}
}
func (output *attemptOutput) Unwrap() http.ResponseWriter { return output.writer }
func (output *attemptOutput) isCommitted() bool {
	output.mu.Lock()
	defer output.mu.Unlock()
	return output.committed
}
func (output *attemptOutput) timeout() bool {
	output.mu.Lock()
	defer output.mu.Unlock()
	return output.timedOut
}
func (output *attemptOutput) stop() {
	output.mu.Lock()
	output.timer.Stop()
	output.mu.Unlock()
	output.cancel()
}
func (output *attemptOutput) commitNormal() error {
	output.mu.Lock()
	defer output.mu.Unlock()
	if output.parent.Err() != nil {
		return output.parent.Err()
	}
	if output.timedOut || !output.timer.Stop() {
		return context.DeadlineExceeded
	}
	output.committed = true
	output.committedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

func transportTemporary(err error) bool {
	return err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, errDestinationForbidden)
}
