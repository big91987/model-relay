package catalog

import (
	"context"
	"database/sql"
	"time"

	"model-relay/internal/fault"
	"model-relay/internal/secure"
	"model-relay/internal/upstream"
)

// Check registers exactly one model-list probe under configuration coordination,
// releases that lock during provider IO, and persists even a canceled outcome.
// Its ten-second total budget includes body reads; the supplied upstream client
// retains its transport/header budget. It never generates or retries a request.
func (service *Service) Check(ctx context.Context, id, expected int64) (CheckResult, error) {
	service.state.Lock()
	source, err := service.UpstreamFacts(ctx, id)
	checking := service.state.Checking(id)
	if err != nil || source.ExecutionRevision != expected || checking || !source.Enabled || !source.CredentialConfigured {
		service.state.Unlock()
		switch {
		case err == sql.ErrNoRows:
			return CheckResult{}, fault.New(fault.NotFound, "not_found")
		case err != nil:
			return CheckResult{}, unavailable()
		case source.ExecutionRevision != expected:
			return CheckResult{}, fault.New(fault.Conflict, "configuration_conflict")
		case checking:
			return CheckResult{}, fault.New(fault.Conflict, "check_in_progress")
		default:
			return CheckResult{}, fault.New(fault.Invalid, "upstream_unavailable")
		}
	}
	started := time.Now()
	check := CheckResult{ExecutionRevision: source.ExecutionRevision, State: "checking", Started: started.UTC().Format(time.RFC3339Nano)}
	result, err := service.db.ExecContext(ctx, "INSERT INTO upstream_checks(upstream_id,execution_revision,state,started_at,error_code) VALUES(?,?,'checking',?,'')", id, source.ExecutionRevision, check.Started)
	if err != nil {
		service.state.Unlock()
		return CheckResult{}, unavailable()
	}
	check.ID, err = result.LastInsertId()
	if err != nil {
		service.state.Unlock()
		return CheckResult{}, unavailable()
	}
	service.state.BeginCheck(id)
	service.state.Unlock()
	defer func() { service.state.Lock(); service.state.EndCheck(id); service.state.Unlock() }()
	probe, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	check.State = "failed"
	check.Code = "upstream_connection_failed"
	token, err := secure.Unseal(service.key, source.Encrypted, secure.UpstreamAAD(id, source.URL))
	if err == nil {
		result := service.client.Check(probe, upstream.Endpoint{URL: source.URL, Token: token})
		check.Status = result.Status
		check.Code = result.Code
		if check.Code == "" {
			check.State = "success"
		}
	}
	if ctx.Err() != nil {
		check.State = "canceled"
		check.Code = "client_canceled"
	} else if probe.Err() != nil {
		check.State = "failed"
		check.Code = "upstream_timeout"
	}
	finished := time.Now().UTC().Format(time.RFC3339Nano)
	duration := time.Since(started).Milliseconds()
	check.Finished = &finished
	check.Duration = &duration
	cleanup, release := context.WithTimeout(context.Background(), 2*time.Second)
	defer release()
	if _, err = service.db.ExecContext(cleanup, "UPDATE upstream_checks SET state=?,finished_at=?,duration_ms=?,error_code=?,upstream_status=? WHERE id=?", check.State, finished, duration, check.Code, check.Status, check.ID); err != nil {
		return CheckResult{}, &CheckPersistenceError{}
	}
	service.state.RLock()
	var execution int64
	err = service.db.QueryRowContext(cleanup, "SELECT execution_revision FROM upstream WHERE id=?", id).Scan(&execution)
	service.state.RUnlock()
	check.Applicable = err == nil && execution == check.ExecutionRevision
	return check, nil
}
