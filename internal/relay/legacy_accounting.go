package relay

import (
	"database/sql"
	"net/http"
	"time"
)

func (server *Server) reconcileLegacy(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, requestID string) {
	if !principal.Platform {
		failure(writer, 403, "forbidden", "")
		return
	}
	var input struct {
		Revision int64  `json:"expected_revision"`
		Reason   string `json:"reason"`
		Evidence string `json:"evidence"`
		Attempts []struct {
			Ordinal  int    `json:"ordinal"`
			Upstream int64  `json:"upstream_id"`
			PriceID  string `json:"historical_price_reference"`
			Input    int64  `json:"input_tokens"`
			Output   int64  `json:"output_tokens"`
			CostIn   int64  `json:"cost_input_rate"`
			CostOut  int64  `json:"cost_output_rate"`
			FeeIn    int64  `json:"tenant_input_rate"`
			FeeOut   int64  `json:"tenant_output_rate"`
		} `json:"attempts"`
	}
	if jsonInput(writer, request, &input) != nil || input.Revision < 0 || !validName(input.Reason) || !validName(input.Evidence) || len(input.Attempts) < 1 || len(input.Attempts) > 3 {
		failure(writer, 400, "historical_evidence_required", "")
		return
	}
	operation, prior, fingerprint, err := operationStart(transaction, principal.ID, request, "legacy.reconcile", input)
	if err != nil {
		failure(writer, 409, err.Error(), "")
		return
	}
	if prior != "" {
		reply(writer, 200, map[string]any{"id": prior, "already_applied": true})
		return
	}
	var tenantID, period, keyID, accepted string
	var resolved bool
	var revision int64
	var running bool
	existing := transaction.QueryRow("SELECT tenant_id,period,key_id,accepted,revision,running FROM admissions WHERE request_id=?", requestID).Scan(&tenantID, &period, &keyID, &accepted, &revision, &running) == nil
	if existing {
		var unpriced int
		if transaction.QueryRow("SELECT count(*) FROM request_prices p LEFT JOIN historical_price_evidence h ON h.request_id=p.request_id AND h.upstream_id=p.upstream_id WHERE p.request_id=? AND p.price_id='legacy_unpriced' AND h.request_id IS NULL", requestID).Scan(&unpriced) != nil || unpriced == 0 || running || revision != input.Revision {
			failure(writer, 409, "historical_revision_conflict_or_resolved", "")
			return
		}
	} else {
		if input.Revision != 0 || transaction.QueryRow("SELECT l.tenant_id,l.period,l.resolved,r.key_id,r.time FROM legacy_pending l JOIN requests r ON r.id=l.request_id WHERE l.request_id=?", requestID).Scan(&tenantID, &period, &resolved, &keyID, &accepted) != nil {
			failure(writer, 404, "not_found", "")
			return
		}
	}
	if resolved {
		failure(writer, 409, "legacy_already_reconciled", "")
		return
	}
	parsed, err := time.Parse(time.RFC3339Nano, accepted)
	if err != nil {
		failure(writer, 409, "legacy_time_requires_supported_recovery", "")
		return
	}
	var actualAttempts int
	if transaction.QueryRow("SELECT count(*) FROM request_attempts a JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal WHERE a.request_id=? AND d.possible=1", requestID).Scan(&actualAttempts) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if actualAttempts > 0 && actualAttempts != len(input.Attempts) || actualAttempts == 0 && len(input.Attempts) != 1 {
		failure(writer, 400, "all_dispatch_evidence_required", "")
		return
	}
	var closed int
	if transaction.QueryRow("SELECT count(*) FROM bills WHERE tenant_id=? AND period=?", tenantID, period).Scan(&closed) != nil || closed != 0 {
		failure(writer, 409, "closed_bill_requires_correction", "")
		return
	}
	if !existing {
		if _, err := transaction.Exec("INSERT INTO admissions VALUES(?,?,?,?,?,?,0,0,0,0,1)", requestID, tenantID, keyID, period, parsed.UTC().Format("2006-01-02T15:04"), accepted); err != nil {
			failure(writer, 409, "legacy_admission_conflict", "")
			return
		}
	}
	seen := map[int]bool{}
	for _, attempt := range input.Attempts {
		if attempt.Ordinal < 1 || attempt.Ordinal > 3 || seen[attempt.Ordinal] || !validName(attempt.PriceID) || attempt.Input < 0 || attempt.Output < 0 || attempt.Input > 1000000000 || attempt.Output > 1000000000 {
			failure(writer, 400, "invalid_historical_attempt", "")
			return
		}
		seen[attempt.Ordinal] = true
		if actualAttempts > 0 {
			var count int
			if transaction.QueryRow("SELECT count(*) FROM request_attempts a JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal WHERE a.request_id=? AND a.ordinal=? AND a.upstream_id=? AND d.possible=1", requestID, attempt.Ordinal, attempt.Upstream).Scan(&count) != nil || count != 1 {
				failure(writer, 400, "attempt_source_mismatch", "")
				return
			}
		}
		cost, err := microCharge(attempt.Input, attempt.Output, attempt.CostIn, attempt.CostOut)
		if err != nil {
			failure(writer, 400, "invalid_amount", "")
			return
		}
		fee, err := microCharge(attempt.Input, attempt.Output, attempt.FeeIn, attempt.FeeOut)
		if err != nil {
			failure(writer, 400, "invalid_amount", "")
			return
		}
		if !existing {
			if _, err := transaction.Exec("INSERT INTO request_prices VALUES(?,?,?,?,?,?,?,?,?)", requestID, attempt.Upstream, "legacy_unpriced", 0, 0, 0, 0, 0, 0); err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
		} else {
			var priceID string
			var previousInput, previousOutput int64
			if transaction.QueryRow("SELECT price_id FROM request_prices WHERE request_id=? AND upstream_id=?", requestID, attempt.Upstream).Scan(&priceID) != nil || priceID != "legacy_unpriced" || transaction.QueryRow("SELECT coalesce(sum(input_tokens),0),coalesce(sum(output_tokens),0) FROM charge_entries WHERE request_id=? AND ordinal=?", requestID, attempt.Ordinal).Scan(&previousInput, &previousOutput) != nil || attempt.Input < previousInput || attempt.Output < previousOutput {
				failure(writer, 409, "historical_evidence_conflict", "")
				return
			}
		}
		if _, err := transaction.Exec("INSERT INTO historical_price_evidence VALUES(?,?,?,?,?,?,?,?,?)", requestID, attempt.Upstream, "legacy_evidence:"+attempt.PriceID, attempt.CostIn, attempt.CostOut, attempt.FeeIn, attempt.FeeOut, input.Reason, input.Evidence); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if actualAttempts == 0 {
			_, err = transaction.Exec("INSERT INTO charge_entries VALUES(?,?,?,?,?,?,?,?,?,?,?)", randomSecret(16), requestID, attempt.Ordinal, revision+1, period, attempt.Input, attempt.Output, cost, fee, input.Reason, input.Evidence)
		} else {
			_, err = transaction.Exec("INSERT INTO usage_reconciliations VALUES(?,?,?,?,?,?,?)", requestID, attempt.Ordinal, revision+1, attempt.Input, attempt.Output, input.Reason, input.Evidence)
		}
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
	}
	if _, err := transaction.Exec("UPDATE admissions SET revision=revision+1 WHERE request_id=?; UPDATE legacy_pending SET resolved=1 WHERE request_id=?", requestID, requestID); err != nil || settleAdmission(transaction, requestID) != nil || operationFinish(transaction, principal.ID, operation, fingerprint, "legacy.reconcile", requestID, tenantID) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"request_id": requestID, "historical_evidence": "operator_confirmed_not_live_supplier_validation", "execution_history_unchanged": true})
}
