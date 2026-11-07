package relay

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"time"

	"model-relay/internal/routing"
)

type routeSnapshot struct {
	Model          RouteModel
	Upstreams      map[int64]Upstream
	ConfigRevision int64
	Epoch          uint64
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
	return snapshot, nil
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
	now := server.routingNow()
	policy := routingSnapshot(snapshot, now)
	for index, candidate := range snapshot.Model.Candidates {
		upstream := snapshot.Upstreams[candidate.UpstreamID]
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
		reason := ""
		switch {
		case !current.Enabled || !snapshot.Model.Enabled:
			reason = "disabled_model"
		case snapshot.Epoch != server.configEpoch || current.Revision != snapshot.Model.Revision || upstream.ExecutionRevision != latest.ExecutionRevision:
			reason = "configuration_changed"
		case !upstream.Enabled || !latest.Enabled:
			reason = "disabled_upstream"
		case !candidate.Enabled || !present:
			reason = "disabled_candidate"
		case !upstream.CredentialConfigured:
			reason = "credential_missing"
		}
		policy.Candidates[index].Unavailable = reason
	}
	selection := server.router.Select(policy, record.Stream, round, tried)
	decisions, observations := selection.Decisions, selection.Observations
	var selected *Candidate
	if selection.Candidate != nil {
		for _, candidate := range snapshot.Model.Candidates {
			if candidate.UpstreamID == selection.Candidate.UpstreamID {
				copy := candidate
				selected = &copy
				break
			}
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
	for _, observation := range observations {
		raw, marshalErr := json.Marshal(observation)
		if marshalErr != nil {
			return nil, nil, "storage_unavailable"
		}
		if _, err = transaction.ExecContext(cleanup, "INSERT INTO route_selection_evidence VALUES(?,?,?,?)", record.ID, round, observation.UpstreamID, string(raw)); err != nil {
			return nil, nil, "storage_unavailable"
		}
	}
	var attempt *Attempt
	if selected != nil {
		upstream := snapshot.Upstreams[selected.UpstreamID]
		sequence := server.router.NextSequence()
		attempt = &Attempt{Ordinal: len(tried) + 1, UpstreamID: upstream.ID, UpstreamName: upstream.Name, RemoteModel: selected.RemoteModel, UpstreamRevision: upstream.Revision, UpstreamExecution: upstream.ExecutionRevision, CandidateExecution: selected.ExecutionRevision, Started: now.UTC().Format(time.RFC3339Nano), State: "dispatching", Sequence: sequence}
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
	server.router.Observe(routingSnapshot(snapshot, time.Time{}), routingCandidate(candidate, upstream), routing.Observation{Cooldown: server.Config.Cooldown, Sequence: attempt.Sequence, State: attempt.State, Temporary: temporary, ObservedAt: server.routingNow()})
}

func temporaryStatus(status int) bool {
	return status == 408 || status == 429 || status == 500 || status == 502 || status == 503 || status == 504
}

func routingCandidate(candidate Candidate, upstream Upstream) routing.Candidate {
	return routing.Candidate{UpstreamID: candidate.UpstreamID, RemoteModel: candidate.RemoteModel, Enabled: candidate.Enabled, Priority: candidate.Priority, Weight: candidate.Weight, ExecutionRevision: candidate.ExecutionRevision, UpstreamExecution: upstream.ExecutionRevision}
}

func routingSnapshot(snapshot routeSnapshot, observed time.Time) routing.Snapshot {
	policy := routing.Snapshot{Alias: snapshot.Model.Alias, Strategy: snapshot.Model.Strategy, Revision: snapshot.Model.Revision, ConfigRevision: snapshot.ConfigRevision, Epoch: snapshot.Epoch, ObservedAt: observed}
	for _, candidate := range snapshot.Model.Candidates {
		policy.Candidates = append(policy.Candidates, routingCandidate(candidate, snapshot.Upstreams[candidate.UpstreamID]))
	}
	return policy
}
