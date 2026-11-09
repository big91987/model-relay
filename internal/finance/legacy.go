package finance

import (
	"context"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/accounting"
	"model-relay/internal/fault"
	"model-relay/internal/names"
	"model-relay/internal/operations"
	"model-relay/internal/secure"
)

func (service *Service) ReconcileLegacy(ctx context.Context, principal access.Principal, requestID, key, path string, input *LegacyInput) (operations.Result, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return operations.Result{}, err
	}
	defer transaction.Rollback()
	if !principal.Platform {
		return operations.Result{}, forbidden()
	}
	if input == nil || input.Revision < 0 || !names.Valid(input.Reason) || !names.Valid(input.Evidence) || len(input.Attempts) < 1 || len(input.Attempts) > 3 {
		return operations.Result{}, fault.New(fault.Invalid, "historical_evidence_required")
	}
	operation, prior, fingerprint, err := beginOperation(transaction, principal, key, path, "legacy.reconcile", input)
	if err != nil {
		return operations.Result{}, err
	}
	if prior != "" {
		return replay(prior), nil
	}
	var tenantID, period, keyID, accepted string
	var resolved bool
	var revision int64
	var running bool
	existing := transaction.QueryRow("SELECT tenant_id,period,key_id,accepted,revision,running FROM admissions WHERE request_id=?", requestID).Scan(&tenantID, &period, &keyID, &accepted, &revision, &running) == nil
	if existing {
		var unpriced int
		if transaction.QueryRow("SELECT count(*) FROM request_prices p LEFT JOIN historical_price_evidence h ON h.request_id=p.request_id AND h.upstream_id=p.upstream_id WHERE p.request_id=? AND p.price_id='legacy_unpriced' AND h.request_id IS NULL", requestID).Scan(&unpriced) != nil || unpriced == 0 || running || revision != input.Revision {
			return operations.Result{}, fault.New(fault.Conflict, "historical_revision_conflict_or_resolved")
		}
	} else {
		if input.Revision != 0 || transaction.QueryRow("SELECT l.tenant_id,l.period,l.resolved,r.key_id,r.time FROM legacy_pending l JOIN requests r ON r.id=l.request_id WHERE l.request_id=?", requestID).Scan(&tenantID, &period, &resolved, &keyID, &accepted) != nil {
			return operations.Result{}, missing()
		}
	}
	if resolved {
		return operations.Result{}, fault.New(fault.Conflict, "legacy_already_reconciled")
	}
	parsed, err := time.Parse(time.RFC3339Nano, accepted)
	if err != nil {
		return operations.Result{}, fault.New(fault.Conflict, "legacy_time_requires_supported_recovery")
	}
	var actualAttempts int
	if transaction.QueryRow("SELECT count(*) FROM request_attempts a JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal WHERE a.request_id=? AND d.possible=1", requestID).Scan(&actualAttempts) != nil {
		return operations.Result{}, storage()
	}
	if actualAttempts > 0 && actualAttempts != len(input.Attempts) || actualAttempts == 0 && len(input.Attempts) != 1 {
		return operations.Result{}, fault.New(fault.Invalid, "all_dispatch_evidence_required")
	}
	var closed int
	if transaction.QueryRow("SELECT count(*) FROM bills WHERE tenant_id=? AND period=?", tenantID, period).Scan(&closed) != nil || closed != 0 {
		return operations.Result{}, fault.New(fault.Conflict, "closed_bill_requires_correction")
	}
	if !existing {
		if _, err := transaction.Exec("INSERT INTO admissions VALUES(?,?,?,?,?,?,0,0,0,0,1)", requestID, tenantID, keyID, period, parsed.UTC().Format("2006-01-02T15:04"), accepted); err != nil {
			return operations.Result{}, fault.New(fault.Conflict, "legacy_admission_conflict")
		}
	}
	seen := map[int]bool{}
	for _, attempt := range input.Attempts {
		if attempt.Ordinal < 1 || attempt.Ordinal > 3 || seen[attempt.Ordinal] || !names.Valid(attempt.PriceID) || attempt.Input < 0 || attempt.Output < 0 || attempt.Input > 1000000000 || attempt.Output > 1000000000 {
			return operations.Result{}, fault.New(fault.Invalid, "invalid_historical_attempt")
		}
		seen[attempt.Ordinal] = true
		if actualAttempts > 0 {
			var count int
			if transaction.QueryRow("SELECT count(*) FROM request_attempts a JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal WHERE a.request_id=? AND a.ordinal=? AND a.upstream_id=? AND d.possible=1", requestID, attempt.Ordinal, attempt.Upstream).Scan(&count) != nil || count != 1 {
				return operations.Result{}, fault.New(fault.Invalid, "attempt_source_mismatch")
			}
		}
		cost, err := accounting.Charge(attempt.Input, attempt.Output, attempt.CostIn, attempt.CostOut)
		if err != nil {
			return operations.Result{}, fault.New(fault.Invalid, "invalid_amount")
		}
		fee, err := accounting.Charge(attempt.Input, attempt.Output, attempt.FeeIn, attempt.FeeOut)
		if err != nil {
			return operations.Result{}, fault.New(fault.Invalid, "invalid_amount")
		}
		if !existing {
			if _, err := transaction.Exec("INSERT INTO request_prices VALUES(?,?,?,?,?,?,?,?,?)", requestID, attempt.Upstream, "legacy_unpriced", 0, 0, 0, 0, 0, 0); err != nil {
				return operations.Result{}, storage()
			}
		} else {
			var priceID string
			var previousInput, previousOutput int64
			if transaction.QueryRow("SELECT price_id FROM request_prices WHERE request_id=? AND upstream_id=?", requestID, attempt.Upstream).Scan(&priceID) != nil || priceID != "legacy_unpriced" || transaction.QueryRow("SELECT coalesce(sum(input_tokens),0),coalesce(sum(output_tokens),0) FROM charge_entries WHERE request_id=? AND ordinal=?", requestID, attempt.Ordinal).Scan(&previousInput, &previousOutput) != nil || attempt.Input < previousInput || attempt.Output < previousOutput {
				return operations.Result{}, fault.New(fault.Conflict, "historical_evidence_conflict")
			}
		}
		if _, err := transaction.Exec("INSERT INTO historical_price_evidence VALUES(?,?,?,?,?,?,?,?,?)", requestID, attempt.Upstream, "legacy_evidence:"+attempt.PriceID, attempt.CostIn, attempt.CostOut, attempt.FeeIn, attempt.FeeOut, input.Reason, input.Evidence); err != nil {
			return operations.Result{}, storage()
		}
		if actualAttempts == 0 {
			_, err = transaction.Exec("INSERT INTO charge_entries VALUES(?,?,?,?,?,?,?,?,?,?,?)", secure.Random(16), requestID, attempt.Ordinal, revision+1, period, attempt.Input, attempt.Output, cost, fee, input.Reason, input.Evidence)
		} else {
			_, err = transaction.Exec("INSERT INTO usage_reconciliations VALUES(?,?,?,?,?,?,?)", requestID, attempt.Ordinal, revision+1, attempt.Input, attempt.Output, input.Reason, input.Evidence)
		}
		if err != nil {
			return operations.Result{}, storage()
		}
	}
	if _, err := transaction.Exec("UPDATE admissions SET revision=revision+1 WHERE request_id=?; UPDATE legacy_pending SET resolved=1 WHERE request_id=?", requestID, requestID); err != nil || Settle(transaction, requestID) != nil {
		return operations.Result{}, storage()
	}
	if err = completeOperation(transaction, principal, operation, fingerprint, "legacy.reconcile", requestID, tenantID); err != nil {
		return operations.Result{}, err
	}
	return operations.Result{Value: struct {
		RequestID                 string `json:"request_id"`
		HistoricalEvidence        string `json:"historical_evidence"`
		ExecutionHistoryUnchanged bool   `json:"execution_history_unchanged"`
	}{requestID, "operator_confirmed_not_live_supplier_validation", true}}, nil
}
