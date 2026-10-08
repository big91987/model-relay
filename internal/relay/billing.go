package relay

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
)

func validPeriod(period string) bool {
	parsed, err := time.Parse("2006-01", period)
	return err == nil && parsed.Format("2006-01") == period
}

func (server *Server) tenantBills(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, tenantID, role string) {
	if role == "developer" {
		failure(writer, 403, "forbidden", "")
		return
	}
	if request.Method == http.MethodGet {
		values, err := queryMaps(transaction, `WITH periods AS (
SELECT period FROM admissions WHERE tenant_id=? UNION SELECT period FROM legacy_pending WHERE tenant_id=?
UNION SELECT period FROM bills WHERE tenant_id=? UNION SELECT c.period FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=?)
SELECT p.period,
(SELECT coalesce(sum(c.cost),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=p.period) AS cost_micro_usd,
(SELECT coalesce(sum(c.fee),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=p.period) AS usage_fee_micro_usd,
(SELECT coalesce(sum(c.amount),0) FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=p.period) AS adjustment_micro_usd,
(SELECT coalesce(sum(c.fee),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=p.period)+(SELECT coalesce(sum(c.amount),0) FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=p.period) AS fee_micro_usd,
(SELECT coalesce(sum(a.pending),0) FROM admissions a WHERE a.tenant_id=? AND a.period=p.period)+(SELECT count(*) FROM legacy_pending l WHERE l.tenant_id=? AND l.period=p.period AND l.resolved=0) AS pending,
b.id AS bill_id,b.closed,b.original_fee,b.original_cost FROM periods p LEFT JOIN bills b ON b.tenant_id=? AND b.period=p.period ORDER BY p.period DESC`, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		replyCollection(writer, request, values, map[string]any{"currency": "USD", "payment": "not_connected"})
		return
	}
	if request.Method != http.MethodPost || !principal.Platform {
		failure(writer, 403, "forbidden", "")
		return
	}
	var input struct {
		Period string `json:"period"`
	}
	if jsonInput(writer, request, &input) != nil || !validPeriod(input.Period) || input.Period >= server.Store.now().Format("2006-01") {
		failure(writer, 409, "period_not_ended", "")
		return
	}
	operation, prior, fingerprint, err := operationStart(transaction, principal.ID, request, "bill.close", input)
	if err != nil {
		failure(writer, 409, err.Error(), "")
		return
	}
	if prior != "" {
		reply(writer, 200, map[string]any{"id": prior, "already_applied": true})
		return
	}
	var existingBill string
	if err := transaction.QueryRow("SELECT id FROM bills WHERE tenant_id=? AND period=?", tenantID, input.Period).Scan(&existingBill); err == nil {
		if operationFinish(transaction, principal.ID, operation, fingerprint, "bill.close", existingBill, tenantID) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]any{"id": existingBill, "already_closed": true})
		return
	} else if err != sql.ErrNoRows {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	var pending, running int64
	if transaction.QueryRow("SELECT coalesce(sum(pending),0),coalesce(sum(running),0) FROM admissions WHERE tenant_id=? AND period=?", tenantID, input.Period).Scan(&pending, &running) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if pending != 0 || running != 0 {
		failure(writer, 409, "pending_usage", "")
		return
	}
	if transaction.QueryRow("SELECT count(*) FROM legacy_pending WHERE tenant_id=? AND period=? AND resolved=0", tenantID, input.Period).Scan(&pending) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if pending > 0 {
		failure(writer, 409, "legacy_price_pending", "")
		return
	}
	var fee, cost int64
	if transaction.QueryRow("SELECT coalesce(sum(c.fee),0),coalesce(sum(c.cost),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=?", tenantID, input.Period).Scan(&fee, &cost) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	var adjustment int64
	if transaction.QueryRow("SELECT coalesce(sum(c.amount),0) FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=?", tenantID, input.Period).Scan(&adjustment) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	fee += adjustment
	id := randomSecret(16)
	if _, err := transaction.Exec("INSERT INTO bills VALUES(?,?,?,?,?,?)", id, tenantID, input.Period, cost, fee, server.Store.now().Format(time.RFC3339Nano)); err != nil {
		failure(writer, 409, "bill_already_closed", "")
		return
	}
	if operationFinish(transaction, principal.ID, operation, fingerprint, "bill.close", id, tenantID) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 201, map[string]any{"id": id, "original_fee": strconv.FormatInt(fee, 10), "original_cost": strconv.FormatInt(cost, 10), "currency": "USD"})
}

func (server *Server) manageBill(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, parts []string) {
	if len(parts) < 1 {
		failure(writer, 404, "not_found", "")
		return
	}
	var tenantID, period, closed string
	var originalFee, originalCost int64
	if transaction.QueryRow("SELECT tenant_id,period,original_fee,original_cost,closed FROM bills WHERE id=?", parts[0]).Scan(&tenantID, &period, &originalFee, &originalCost, &closed) != nil {
		failure(writer, 404, "not_found", "")
		return
	}
	role, err := scopedRole(transaction, principal, tenantID)
	if err != nil {
		failure(writer, 404, "not_found", "")
		return
	}
	if role == "developer" {
		failure(writer, 403, "forbidden", "")
		return
	}
	if request.Method == http.MethodGet {
		corrections, err := queryMaps(transaction, "SELECT id,period,amount,reason,actor,created FROM bill_corrections WHERE bill_id=? ORDER BY created,id", parts[0])
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if len(parts) == 2 && parts[1] == "export" {
			entries, err := queryMaps(transaction, `SELECT 'usage' AS kind,c.id,c.period,c.request_id,c.ordinal,c.cost AS cost_micro_usd,c.fee AS fee_micro_usd,c.reason,c.evidence FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=? UNION ALL SELECT 'adjustment',c.id,c.period,c.bill_id,NULL,NULL,c.amount,c.reason,c.actor FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=? ORDER BY kind,id`, tenantID, period, tenantID, period)
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			writeCSV(writer, []string{"kind", "id", "period", "request_id", "ordinal", "cost_micro_usd", "fee_micro_usd", "reason", "evidence"}, entries)
			return
		}
		var adjustments int64
		if transaction.QueryRow("SELECT coalesce(sum(amount),0) FROM bill_corrections WHERE bill_id=?", parts[0]).Scan(&adjustments) != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]any{"id": parts[0], "tenant_id": tenantID, "period": period, "original_fee": strconv.FormatInt(originalFee, 10), "original_cost": strconv.FormatInt(originalCost, 10), "current_fee_micro_usd": strconv.FormatInt(originalFee+adjustments, 10), "adjustment_micro_usd": strconv.FormatInt(adjustments, 10), "closed_at": closed, "corrections": corrections, "currency": "USD", "payment": "not_connected"})
		return
	}
	if !principal.Platform {
		failure(writer, 403, "forbidden", "")
		return
	}
	if len(parts) != 2 || parts[1] != "corrections" || request.Method != http.MethodPost {
		failure(writer, 404, "not_found", "")
		return
	}
	var input struct {
		Period string `json:"adjustment_period"`
		Amount int64  `json:"amount_micro_usd"`
		Reason string `json:"reason"`
	}
	if jsonInput(writer, request, &input) != nil || !validName(input.Reason) || input.Amount == 0 || input.Amount < -1000000000000 || input.Amount > 1000000000000 {
		failure(writer, 400, "invalid_correction", "")
		return
	}
	operation, prior, fingerprint, err := operationStart(transaction, principal.ID, request, "bill.correct", input)
	if err != nil {
		failure(writer, 409, err.Error(), "")
		return
	}
	if prior != "" {
		reply(writer, 200, map[string]any{"id": prior, "already_applied": true})
		return
	}
	if input.Period == "" {
		input.Period = server.Store.now().Format("2006-01")
	}
	var closedAdjustment int
	if !validPeriod(input.Period) || input.Period < server.Store.now().Format("2006-01") || transaction.QueryRow("SELECT count(*) FROM bills WHERE tenant_id=? AND period=?", tenantID, input.Period).Scan(&closedAdjustment) != nil || closedAdjustment != 0 {
		failure(writer, 400, "adjustment_period_not_open", "")
		return
	}
	var sum int64
	if transaction.QueryRow("SELECT coalesce(sum(amount),0) FROM bill_corrections WHERE bill_id=?", parts[0]).Scan(&sum) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if originalFee+sum+input.Amount < 0 {
		failure(writer, 409, "negative_bill_total", "")
		return
	}
	id := randomSecret(16)
	if _, err := transaction.Exec("INSERT INTO bill_corrections VALUES(?,?,?,?,?,?,?)", id, parts[0], input.Period, input.Amount, input.Reason, principal.ID, server.Store.now().Format(time.RFC3339Nano)); err != nil || operationFinish(transaction, principal.ID, operation, fingerprint, "bill.correct", id, tenantID) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 201, map[string]any{"id": id, "original_fee": strconv.FormatInt(originalFee, 10), "adjustment": strconv.FormatInt(input.Amount, 10), "adjustment_period": input.Period})
}

func (server *Server) reconcileUsage(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, requestID, ordinalText string) {
	if !principal.Platform {
		failure(writer, 403, "forbidden", "")
		return
	}
	ordinal, err := strconv.Atoi(ordinalText)
	if err != nil || ordinal < 1 || ordinal > 3 {
		failure(writer, 400, "invalid_attempt", "")
		return
	}
	var input struct {
		Input    int64  `json:"input_tokens"`
		Output   int64  `json:"output_tokens"`
		Reason   string `json:"reason"`
		Evidence string `json:"evidence"`
		Revision int64  `json:"expected_revision"`
	}
	if jsonInput(writer, request, &input) != nil || input.Input < 0 || input.Output < 0 || input.Input > 1000000000 || input.Output > 1000000000 || !validName(input.Reason) || !validName(input.Evidence) {
		failure(writer, 400, "evidence_required", "")
		return
	}
	operation, prior, fingerprint, err := operationStart(transaction, principal.ID, request, "usage.reconcile", input)
	if err != nil {
		failure(writer, 409, err.Error(), "")
		return
	}
	if prior != "" {
		reply(writer, 200, map[string]any{"id": prior, "already_applied": true})
		return
	}
	var tenantID, period string
	var running bool
	var revision int64
	if transaction.QueryRow("SELECT tenant_id,period,running,revision FROM admissions WHERE request_id=?", requestID).Scan(&tenantID, &period, &running, &revision) != nil {
		failure(writer, 404, "not_found", "")
		return
	}
	if running || revision != input.Revision {
		failure(writer, 409, "revision_conflict_or_running", "")
		return
	}
	var previousInput, previousOutput int64
	if transaction.QueryRow("SELECT coalesce(sum(input_tokens),0),coalesce(sum(output_tokens),0) FROM charge_entries WHERE request_id=? AND ordinal=?", requestID, ordinal).Scan(&previousInput, &previousOutput) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if input.Input < previousInput || input.Output < previousOutput {
		failure(writer, 409, "cumulative_usage_cannot_decrease", "")
		return
	}
	var count int
	if transaction.QueryRow("SELECT count(*) FROM bills WHERE tenant_id=? AND period=?", tenantID, period).Scan(&count) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if count > 0 {
		failure(writer, 409, "closed_bill_requires_correction", "")
		return
	}
	if transaction.QueryRow("SELECT count(*) FROM request_attempts a JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal WHERE a.request_id=? AND a.ordinal=? AND d.possible=1", requestID, ordinal).Scan(&count) != nil || count != 1 {
		failure(writer, 404, "not_found", "")
		return
	}
	if _, err := transaction.Exec("INSERT INTO usage_reconciliations VALUES(?,?,?,?,?,?,?); UPDATE admissions SET revision=revision+1 WHERE request_id=?", requestID, ordinal, revision+1, input.Input, input.Output, input.Reason, input.Evidence, requestID); err != nil || settleAdmission(transaction, requestID) != nil || operationFinish(transaction, principal.ID, operation, fingerprint, "usage.reconcile", requestID, tenantID) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"request_id": requestID, "revision": revision + 1, "execution_state_unchanged": true})
}
