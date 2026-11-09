package finance

import (
	"database/sql"
	"errors"
	"model-relay/internal/accounting"
	"time"
)

// LimitsFor reads limits within an existing admission transaction.
func LimitsFor(transaction *sql.Tx, ownerType, ownerID string) (accounting.ResourceLimits, error) {
	var value accounting.ResourceLimits
	err := transaction.QueryRow("SELECT calls,rpm,tpm,concurrency,budget,hard,revision FROM resource_limits WHERE owner_type=? AND owner_id=?", ownerType, ownerID).Scan(&value.Calls, &value.RPM, &value.TPM, &value.Concurrency, &value.Budget, &value.Hard, &value.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return accounting.ResourceLimits{Hard: true}, nil
	}
	return value, err
}

// UsageFor reads usage within an existing admission transaction.
func UsageFor(transaction *sql.Tx, ownerType, ownerID, period, minute string) (accounting.ResourceUsage, error) {
	column := "tenant_id"
	if ownerType == "key" {
		column = "key_id"
	}
	var value accounting.ResourceUsage
	err := transaction.QueryRow("SELECT coalesce(sum(CASE WHEN period=? THEN 1 ELSE 0 END),0),coalesce(sum(CASE WHEN minute=? THEN 1 ELSE 0 END),0),coalesce(sum(CASE WHEN running=1 THEN 1 ELSE 0 END),0),coalesce(sum(tokens_reserved),0),coalesce(sum(fee_reserved),0),coalesce(sum(pending),0) FROM admissions WHERE "+column+"=?", period, minute, ownerID).Scan(&value.Calls, &value.RPM, &value.Concurrency, &value.TPM, &value.Reserved, &value.Pending)
	if err != nil {
		return value, err
	}
	var knownTokens int64
	err = transaction.QueryRow("SELECT coalesce(sum(CASE WHEN a.minute=? THEN c.input_tokens+c.output_tokens ELSE 0 END),0),coalesce(sum(CASE WHEN c.period=? THEN c.fee ELSE 0 END),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a."+column+"=?", minute, period, ownerID).Scan(&knownTokens, &value.Fee)
	if err != nil {
		return value, err
	}
	value.TPM += knownTokens
	var legacyCalls, legacyRPM int64
	if transaction.QueryRow("SELECT coalesce(sum(CASE WHEN l.period=? THEN 1 ELSE 0 END),0),coalesce(sum(CASE WHEN substr(r.time,1,16)=? THEN 1 ELSE 0 END),0) FROM legacy_pending l JOIN requests r ON r.id=l.request_id JOIN key_controls c ON c.key_id=r.key_id WHERE l.resolved=0 AND c."+column+"=?", period, minute, ownerID).Scan(&legacyCalls, &legacyRPM) != nil {
		return value, errors.New("historical usage scan failed")
	}
	value.Calls += legacyCalls
	value.RPM += legacyRPM
	if ownerType == "tenant" {
		var adjustment int64
		err = transaction.QueryRow("SELECT coalesce(sum(c.amount),0) FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=?", ownerID, period).Scan(&adjustment)
		value.Fee += adjustment
	}
	return value, err
}

// AdmissionClock retains the latest accepted time within the transaction.
func AdmissionClock(transaction *sql.Tx, now time.Time) time.Time {
	var latest sql.NullString
	if transaction.QueryRow("SELECT accepted FROM admissions ORDER BY julianday(accepted) DESC LIMIT 1").Scan(&latest) == nil && latest.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, latest.String)
		if err == nil && parsed.After(now) {
			return parsed
		}
	}
	return now
}
