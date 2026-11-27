package finance

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
	"model-relay/internal/secure"
)

func (service *Service) ListBills(ctx context.Context, principal access.Principal, tenantID string) ([]BillPeriod, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	role, err := access.Role(transaction, principal, tenantID)
	if err != nil {
		return nil, missing()
	}
	if role == "developer" {
		return nil, forbidden()
	}
	rows, err := transaction.Query(billPeriodsQuery, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID)
	if err != nil {
		return nil, storage()
	}
	defer rows.Close()
	values := []BillPeriod{}
	for rows.Next() {
		var value BillPeriod
		var cost, usageFee, adjustment, fee int64
		var billID, closed sql.NullString
		var originalFee, originalCost sql.NullInt64
		if rows.Scan(&value.Period, &cost, &usageFee, &adjustment, &fee, &value.Pending, &billID, &closed, &originalFee, &originalCost) != nil {
			return nil, storage()
		}
		value.Cost = strconv.FormatInt(cost, 10)
		value.UsageFee = strconv.FormatInt(usageFee, 10)
		value.Adjustment = strconv.FormatInt(adjustment, 10)
		value.Fee = strconv.FormatInt(fee, 10)
		value.BillID = nullableText(billID)
		value.Closed = nullableText(closed)
		value.OriginalFee = nullableMoney(originalFee)
		value.OriginalCost = nullableMoney(originalCost)
		values = append(values, value)
	}
	if rows.Err() != nil {
		return nil, storage()
	}
	return values, nil
}

func (service *Service) CloseBill(ctx context.Context, principal access.Principal, tenantID, key, path string, input *CloseBillInput) (operations.Result, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return operations.Result{}, err
	}
	defer transaction.Rollback()
	role, err := access.Role(transaction, principal, tenantID)
	if err != nil {
		return operations.Result{}, missing()
	}
	if role == "developer" || !principal.Platform {
		return operations.Result{}, forbidden()
	}
	if input == nil || !validPeriod(input.Period) || input.Period >= service.now().Format("2006-01") {
		return operations.Result{}, fault.New(fault.Conflict, "period_not_ended")
	}
	operation, prior, fingerprint, err := beginOperation(transaction, principal, key, path, "bill.close", input)
	if err != nil {
		return operations.Result{}, err
	}
	if prior != "" {
		return replay(prior), nil
	}
	var existing string
	if err := transaction.QueryRow("SELECT id FROM bills WHERE tenant_id=? AND period=?", tenantID, input.Period).Scan(&existing); err == nil {
		if err = completeOperation(transaction, principal, operation, fingerprint, "bill.close", existing, tenantID); err != nil {
			return operations.Result{}, err
		}
		return operations.Result{Value: struct {
			ID            string `json:"id"`
			AlreadyClosed bool   `json:"already_closed"`
		}{existing, true}}, nil
	} else if err != sql.ErrNoRows {
		return operations.Result{}, storage()
	}
	var pending, running int64
	if transaction.QueryRow("SELECT coalesce(sum(pending),0),coalesce(sum(running),0) FROM admissions WHERE tenant_id=? AND period=?", tenantID, input.Period).Scan(&pending, &running) != nil {
		return operations.Result{}, storage()
	}
	if pending != 0 || running != 0 {
		return operations.Result{}, fault.New(fault.Conflict, "pending_usage")
	}
	if transaction.QueryRow("SELECT count(*) FROM legacy_pending WHERE tenant_id=? AND period=? AND resolved=0", tenantID, input.Period).Scan(&pending) != nil {
		return operations.Result{}, storage()
	}
	if pending > 0 {
		return operations.Result{}, fault.New(fault.Conflict, "legacy_price_pending")
	}
	var fee, cost, adjustment int64
	if transaction.QueryRow("SELECT coalesce(sum(c.fee),0),coalesce(sum(c.cost),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=?", tenantID, input.Period).Scan(&fee, &cost) != nil {
		return operations.Result{}, storage()
	}
	if transaction.QueryRow("SELECT coalesce(sum(c.amount),0) FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=?", tenantID, input.Period).Scan(&adjustment) != nil {
		return operations.Result{}, storage()
	}
	fee += adjustment
	id := secure.Random(16)
	if _, err := transaction.Exec("INSERT INTO bills VALUES(?,?,?,?,?,?)", id, tenantID, input.Period, cost, fee, service.now().Format(time.RFC3339Nano)); err != nil {
		return operations.Result{}, fault.New(fault.Conflict, "bill_already_closed")
	}
	if err = completeOperation(transaction, principal, operation, fingerprint, "bill.close", id, tenantID); err != nil {
		return operations.Result{}, err
	}
	return operations.Result{Created: true, Value: struct {
		ID           string `json:"id"`
		OriginalFee  string `json:"original_fee"`
		OriginalCost string `json:"original_cost"`
		Currency     string `json:"currency"`
	}{id, strconv.FormatInt(fee, 10), strconv.FormatInt(cost, 10), "USD"}}, nil
}

type billRecord struct {
	tenantID, period, closed  string
	originalFee, originalCost int64
}

func billFor(transaction *sql.Tx, principal access.Principal, id string) (billRecord, error) {
	var value billRecord
	if transaction.QueryRow("SELECT tenant_id,period,original_fee,original_cost,closed FROM bills WHERE id=?", id).Scan(&value.tenantID, &value.period, &value.originalFee, &value.originalCost, &value.closed) != nil {
		return value, missing()
	}
	role, err := access.Role(transaction, principal, value.tenantID)
	if err != nil {
		return value, missing()
	}
	if role == "developer" {
		return value, forbidden()
	}
	return value, nil
}

func billCorrections(transaction *sql.Tx, id string) ([]BillCorrection, error) {
	rows, err := transaction.Query("SELECT id,period,amount,reason,actor,created FROM bill_corrections WHERE bill_id=? ORDER BY created,id", id)
	if err != nil {
		return nil, storage()
	}
	defer rows.Close()
	values := []BillCorrection{}
	for rows.Next() {
		var value BillCorrection
		var amount int64
		if rows.Scan(&value.ID, &value.Period, &amount, &value.Reason, &value.Actor, &value.Created) != nil {
			return nil, storage()
		}
		value.Amount = strconv.FormatInt(amount, 10)
		values = append(values, value)
	}
	if rows.Err() != nil {
		return nil, storage()
	}
	return values, nil
}

func (service *Service) Bill(ctx context.Context, principal access.Principal, id string) (BillDetail, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return BillDetail{}, err
	}
	defer transaction.Rollback()
	value, err := billFor(transaction, principal, id)
	if err != nil {
		return BillDetail{}, err
	}
	corrections, err := billCorrections(transaction, id)
	if err != nil {
		return BillDetail{}, err
	}
	var adjustments int64
	if transaction.QueryRow("SELECT coalesce(sum(amount),0) FROM bill_corrections WHERE bill_id=?", id).Scan(&adjustments) != nil {
		return BillDetail{}, storage()
	}
	return BillDetail{ID: id, TenantID: value.tenantID, Period: value.period, OriginalFee: strconv.FormatInt(value.originalFee, 10), OriginalCost: strconv.FormatInt(value.originalCost, 10), CurrentFee: strconv.FormatInt(value.originalFee+adjustments, 10), Adjustment: strconv.FormatInt(adjustments, 10), Closed: value.closed, Corrections: corrections, Currency: "USD", Payment: "not_connected"}, nil
}

func (service *Service) ExportBill(ctx context.Context, principal access.Principal, id string) ([]BillEntry, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	value, err := billFor(transaction, principal, id)
	if err != nil {
		return nil, err
	}
	if _, err = billCorrections(transaction, id); err != nil {
		return nil, err
	}
	rows, err := transaction.Query(`SELECT 'usage' AS kind,c.id,c.period,c.request_id,c.ordinal,c.cost AS cost_micro_usd,c.fee AS fee_micro_usd,c.reason,c.evidence FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=? UNION ALL SELECT 'adjustment',c.id,c.period,c.bill_id,NULL,NULL,c.amount,c.reason,c.actor FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=? ORDER BY kind,id`, value.tenantID, value.period, value.tenantID, value.period)
	if err != nil {
		return nil, storage()
	}
	defer rows.Close()
	entries := []BillEntry{}
	for rows.Next() {
		var entry BillEntry
		var ordinal, cost, fee sql.NullInt64
		if rows.Scan(&entry.Kind, &entry.ID, &entry.Period, &entry.RequestID, &ordinal, &cost, &fee, &entry.Reason, &entry.Evidence) != nil {
			return nil, storage()
		}
		if ordinal.Valid {
			entry.Ordinal = &ordinal.Int64
		}
		entry.Cost = nullableMoney(cost)
		entry.Fee = nullableMoney(fee)
		entries = append(entries, entry)
	}
	if rows.Err() != nil {
		return nil, storage()
	}
	return entries, nil
}

func (service *Service) CorrectBill(ctx context.Context, principal access.Principal, id, key, path string, command *CorrectionInput) (operations.Result, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return operations.Result{}, err
	}
	defer transaction.Rollback()
	value, err := billFor(transaction, principal, id)
	if err != nil {
		return operations.Result{}, err
	}
	if !principal.Platform {
		return operations.Result{}, forbidden()
	}
	if command == nil || !validCorrection(*command) {
		return operations.Result{}, fault.New(fault.Invalid, "invalid_correction")
	}
	input := *command
	operation, prior, fingerprint, err := beginOperation(transaction, principal, key, path, "bill.correct", input)
	if err != nil {
		return operations.Result{}, err
	}
	if prior != "" {
		return replay(prior), nil
	}
	if input.Period == "" {
		input.Period = service.now().Format("2006-01")
	}
	var closed int
	if !validPeriod(input.Period) || input.Period < service.now().Format("2006-01") || transaction.QueryRow("SELECT count(*) FROM bills WHERE tenant_id=? AND period=?", value.tenantID, input.Period).Scan(&closed) != nil || closed != 0 {
		return operations.Result{}, fault.New(fault.Invalid, "adjustment_period_not_open")
	}
	var sum int64
	if transaction.QueryRow("SELECT coalesce(sum(amount),0) FROM bill_corrections WHERE bill_id=?", id).Scan(&sum) != nil {
		return operations.Result{}, storage()
	}
	if value.originalFee+sum+input.Amount < 0 {
		return operations.Result{}, fault.New(fault.Conflict, "negative_bill_total")
	}
	correctionID := secure.Random(16)
	if _, err := transaction.Exec("INSERT INTO bill_corrections VALUES(?,?,?,?,?,?,?)", correctionID, id, input.Period, input.Amount, input.Reason, principal.ID, service.now().Format(time.RFC3339Nano)); err != nil {
		return operations.Result{}, storage()
	}
	if err = completeOperation(transaction, principal, operation, fingerprint, "bill.correct", correctionID, value.tenantID); err != nil {
		return operations.Result{}, err
	}
	return operations.Result{Created: true, Value: struct {
		ID          string `json:"id"`
		OriginalFee string `json:"original_fee"`
		Adjustment  string `json:"adjustment"`
		Period      string `json:"adjustment_period"`
	}{correctionID, strconv.FormatInt(value.originalFee, 10), strconv.FormatInt(input.Amount, 10), input.Period}}, nil
}

func nullableText(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
func nullableMoney(value sql.NullInt64) *string {
	if !value.Valid {
		return nil
	}
	text := strconv.FormatInt(value.Int64, 10)
	return &text
}

const billPeriodsQuery = `WITH periods AS (
SELECT period FROM admissions WHERE tenant_id=? UNION SELECT period FROM legacy_pending WHERE tenant_id=?
UNION SELECT period FROM bills WHERE tenant_id=? UNION SELECT c.period FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=?)
SELECT p.period,
(SELECT coalesce(sum(c.cost),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=p.period) AS cost_micro_usd,
(SELECT coalesce(sum(c.fee),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=p.period) AS usage_fee_micro_usd,
(SELECT coalesce(sum(c.amount),0) FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=p.period) AS adjustment_micro_usd,
(SELECT coalesce(sum(c.fee),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a.tenant_id=? AND c.period=p.period)+(SELECT coalesce(sum(c.amount),0) FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=p.period) AS fee_micro_usd,
(SELECT coalesce(sum(a.pending),0) FROM admissions a WHERE a.tenant_id=? AND a.period=p.period)+(SELECT count(*) FROM legacy_pending l WHERE l.tenant_id=? AND l.period=p.period AND l.resolved=0) AS pending,
b.id AS bill_id,b.closed,b.original_fee,b.original_cost FROM periods p LEFT JOIN bills b ON b.tenant_id=? AND b.period=p.period ORDER BY p.period DESC`
