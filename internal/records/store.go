package records

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"model-relay/internal/finance"
	"model-relay/internal/routing"
	"time"
)

type Store struct {
	DB     *sql.DB
	Schema int
}

func New(db *sql.DB, schema int) *Store { return &Store{DB: db, Schema: schema} }

type Query interface {
	Query(string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
}

func (store *Store) Start(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := store.DB.ExecContext(ctx, "INSERT INTO requests(id,time,key_id,alias,model,stream,revision,state,error_code,evidence_version,config_revision,model_revision) VALUES(?,?,?,?,?,?,?,'started','',2,?,?)", record.ID, record.Time, record.KeyID, record.Alias, record.Model, record.Stream, record.Revision, record.ConfigRevision, record.ModelRevision)
	return err
}

func (store *Store) Finish(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	transaction, err := store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	usage, complete, err := aggregateUsage(transaction, record.ID)
	if err != nil {
		return err
	}
	query := "UPDATE requests SET state=?,error_code=?,model=?,revision=?,upstream_status=?,duration_ms=?,prompt_tokens=?,completion_tokens=?,total_tokens=?,successful_upstream_id=?,committed_at=?,attempts_complete=? WHERE id=?"
	if store.Schema == 3 {
		query += " AND state='started'"
	}
	_, err = transaction.ExecContext(ctx, query, record.State, record.Code, record.Model, record.Revision, record.Status, record.Duration, usage.Prompt, usage.Completion, usage.Total, record.SuccessfulUpstream, record.CommittedAt, complete, record.ID)
	if err != nil {
		return err
	}
	if store.Schema == 3 {
		if err := finance.Settle(transaction, record.ID); err != nil {
			return err
		}
	}
	return transaction.Commit()
}

func UsageParts(attempts []Attempt, complete bool) map[string]UsagePart {
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

func ReadAttempts(query Query, id string) ([]Attempt, error) {
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
	attempts, err := ReadAttempts(transaction, id)
	if err != nil {
		return Usage{}, false, err
	}
	complete := true
	for _, attempt := range attempts {
		if attempt.Finished == nil {
			complete = false
		}
	}
	parts := UsageParts(attempts, complete)
	return Usage{Prompt: parts["prompt_tokens"].Total, Completion: parts["completion_tokens"].Total, Total: parts["total_tokens"].Total}, complete, nil
}

func (store *Store) FinishAttempt(id string, attempt *Attempt) error {
	cleanup, release := context.WithTimeout(context.Background(), 2*time.Second)
	defer release()
	query := "UPDATE request_attempts SET finished_at=?,duration_ms=?,state=?,error_code=?,upstream_status=?,prompt_tokens=?,completion_tokens=?,total_tokens=? WHERE request_id=? AND ordinal=?"
	if store.Schema == 3 {
		query += " AND finished_at IS NULL AND EXISTS(SELECT 1 FROM requests WHERE id=request_id AND state='started')"
	}
	_, err := store.DB.ExecContext(cleanup, query, attempt.Finished, attempt.Duration, attempt.State, attempt.Code, attempt.Status, attempt.Prompt, attempt.Completion, attempt.Total, id, attempt.Ordinal)
	return err
}

func ReadRoutingEvidence(query Query, id string) ([]routing.SelectionObservation, []routing.LatencyResult, error) {
	decisions := []routing.SelectionObservation{}
	attempts := []routing.LatencyResult{}
	rows, err := query.Query("SELECT evidence FROM route_selection_evidence WHERE request_id=? ORDER BY selection_round,upstream_id", id)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var raw string
		var value routing.SelectionObservation
		if err = rows.Scan(&raw); err != nil {
			break
		}
		if err = json.Unmarshal([]byte(raw), &value); err != nil {
			break
		}
		decisions = append(decisions, value)
	}
	rowErr := rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, err
	}
	if rowErr != nil {
		return nil, nil, rowErr
	}
	rows, err = query.Query("SELECT evidence FROM route_latency_evidence WHERE request_id=? ORDER BY ordinal", id)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var raw string
		var value routing.LatencyResult
		if err = rows.Scan(&raw); err != nil {
			break
		}
		if err = json.Unmarshal([]byte(raw), &value); err != nil {
			break
		}
		attempts = append(attempts, value)
	}
	rowErr = rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, err
	}
	return decisions, attempts, rowErr
}

func (store *Store) WriteLatencyEvidence(id string, result routing.LatencyResult) error {
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = store.DB.Exec("INSERT INTO route_latency_evidence VALUES(?,?,?) ON CONFLICT(request_id,ordinal) DO UPDATE SET evidence=excluded.evidence", id, result.Ordinal, string(raw))
	return err
}

func (store *Store) BeginAttempt(id string, round int, decisions []routing.RouteDecision, observations []routing.SelectionObservation, attempt *Attempt, nextSequence func() int64) error {
	cleanup, release := context.WithTimeout(context.Background(), 2*time.Second)
	defer release()
	transaction, err := store.DB.BeginTx(cleanup, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	for _, decision := range decisions {
		if _, err = transaction.ExecContext(cleanup, "INSERT INTO request_route_decisions VALUES(?,?,?,?,?,?,?,?,?)", id, round, decision.UpstreamID, decision.Enabled, decision.Priority, decision.Weight, decision.Reason, decision.Observed, decision.Cooldown); err != nil {
			return err
		}
	}
	for _, observation := range observations {
		raw, marshalErr := json.Marshal(observation)
		if marshalErr != nil {
			return marshalErr
		}
		if _, err = transaction.ExecContext(cleanup, "INSERT INTO route_selection_evidence VALUES(?,?,?,?)", id, round, observation.UpstreamID, string(raw)); err != nil {
			return err
		}
	}
	if attempt != nil {
		attempt.Sequence = nextSequence()
		if _, err = transaction.ExecContext(cleanup, "INSERT INTO request_attempts(request_id,ordinal,upstream_id,upstream_name,remote_model,upstream_revision,upstream_execution_revision,candidate_execution_revision,started_at,state,error_code) VALUES(?,?,?,?,?,?,?,?,?,'dispatching','')", id, attempt.Ordinal, attempt.UpstreamID, attempt.UpstreamName, attempt.RemoteModel, attempt.UpstreamRevision, attempt.UpstreamExecution, attempt.CandidateExecution, attempt.Started); err != nil {
			return err
		}
	}
	if err = transaction.Commit(); err != nil {
		return err
	}
	return nil
}
func (store *Store) MarkDispatch(id string, ordinal int) error {
	if store.Schema != 3 {
		return nil
	}
	_, err := store.DB.Exec("UPDATE attempt_dispatches SET possible=1 WHERE request_id=? AND ordinal=?", id, ordinal)
	return err
}
