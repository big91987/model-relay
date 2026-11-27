package finance

import (
	"context"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
)

func (service *Service) ReconcileUsage(ctx context.Context, principal access.Principal, requestID string, ordinal int, key, path string, input *ReconcileInput) (operations.Result, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return operations.Result{}, err
	}
	defer transaction.Rollback()
	if !principal.Platform {
		return operations.Result{}, forbidden()
	}
	if ordinal < 1 || ordinal > 3 {
		return operations.Result{}, fault.New(fault.Invalid, "invalid_attempt")
	}
	if input == nil || !validReconciliation(*input) {
		return operations.Result{}, fault.New(fault.Invalid, "evidence_required")
	}
	operation, prior, fingerprint, err := beginOperation(transaction, principal, key, path, "usage.reconcile", input)
	if err != nil {
		return operations.Result{}, err
	}
	if prior != "" {
		return replay(prior), nil
	}
	var tenantID, period string
	var running bool
	var revision int64
	if transaction.QueryRow("SELECT tenant_id,period,running,revision FROM admissions WHERE request_id=?", requestID).Scan(&tenantID, &period, &running, &revision) != nil {
		return operations.Result{}, missing()
	}
	if running || revision != input.Revision {
		return operations.Result{}, fault.New(fault.Conflict, "revision_conflict_or_running")
	}
	var previousInput, previousOutput int64
	if transaction.QueryRow("SELECT coalesce(sum(input_tokens),0),coalesce(sum(output_tokens),0) FROM charge_entries WHERE request_id=? AND ordinal=?", requestID, ordinal).Scan(&previousInput, &previousOutput) != nil {
		return operations.Result{}, storage()
	}
	if input.Input < previousInput || input.Output < previousOutput {
		return operations.Result{}, fault.New(fault.Conflict, "cumulative_usage_cannot_decrease")
	}
	var count int
	if transaction.QueryRow("SELECT count(*) FROM bills WHERE tenant_id=? AND period=?", tenantID, period).Scan(&count) != nil {
		return operations.Result{}, storage()
	}
	if count > 0 {
		return operations.Result{}, fault.New(fault.Conflict, "closed_bill_requires_correction")
	}
	if transaction.QueryRow("SELECT count(*) FROM request_attempts a JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal WHERE a.request_id=? AND a.ordinal=? AND d.possible=1", requestID, ordinal).Scan(&count) != nil || count != 1 {
		return operations.Result{}, missing()
	}
	if _, err := transaction.Exec("INSERT INTO usage_reconciliations VALUES(?,?,?,?,?,?,?); UPDATE admissions SET revision=revision+1 WHERE request_id=?", requestID, ordinal, revision+1, input.Input, input.Output, input.Reason, input.Evidence, requestID); err != nil || Settle(transaction, requestID) != nil {
		return operations.Result{}, storage()
	}
	if err = completeOperation(transaction, principal, operation, fingerprint, "usage.reconcile", requestID, tenantID); err != nil {
		return operations.Result{}, err
	}
	return operations.Result{Value: struct {
		RequestID               string `json:"request_id"`
		Revision                int64  `json:"revision"`
		ExecutionStateUnchanged bool   `json:"execution_state_unchanged"`
	}{requestID, revision + 1, true}}, nil
}
