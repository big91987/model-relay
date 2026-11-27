// Package admission owns atomic authorization, quota checks and request reservations.
package admission

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"model-relay/internal/accounting"
	"model-relay/internal/apikey"
	"model-relay/internal/catalog"
	"model-relay/internal/finance"
	"model-relay/internal/names"
	"model-relay/internal/records"
	"model-relay/internal/secure"
)

type Service struct {
	db           *sql.DB
	now          func() time.Time
	advanceClock func() error
}

// New uses the installation's effective-time operations without owning that clock.
func New(db *sql.DB, now func() time.Time, advanceClock func() error) *Service {
	return &Service{db: db, now: now, advanceClock: advanceClock}
}

func (service *Service) Admit(record records.Record, snapshot catalog.Snapshot, payload map[string]json.RawMessage, credential, operation string) (string, int, string) {
	if service.advanceClock() != nil {
		return "storage_unavailable", 503, ""
	}
	transaction, err := service.db.Begin()
	if err != nil {
		return "storage_unavailable", 503, ""
	}
	defer transaction.Rollback()
	if keyID, valid := apikey.GenerationIdentity(transaction, credential, service.now()); !valid || keyID != record.KeyID {
		return "unauthorized", 401, ""
	}
	access, err := apikey.LookupModelAccess(transaction, record.KeyID, record.Alias)
	if err != nil {
		return "storage_unavailable", 503, ""
	}
	if !access.Allowed {
		return "model_forbidden", 403, ""
	}
	var fingerprint []byte
	if operation != "" {
		if !names.Valid(operation) {
			return "invalid_idempotency_key", 400, ""
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "invalid_request", 400, ""
		}
		fingerprint = secure.Digest(string(encoded))
		var prior string
		var saved []byte
		err = transaction.QueryRow("SELECT fingerprint,request_id FROM generation_operations WHERE key_id=? AND operation=?", record.KeyID, operation).Scan(&saved, &prior)
		if err == nil {
			if !bytes.Equal(saved, fingerprint) {
				return "idempotency_conflict", 409, ""
			}
			return "logical_call_already_exists", 409, prior
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "storage_unavailable", 503, ""
		}
	}
	tenantID := access.TenantID
	if !access.Active {
		return "tenant_paused", 403, ""
	}
	tenantLimit, err := finance.LimitsFor(transaction, "tenant", tenantID)
	if err != nil {
		return "storage_unavailable", 503, ""
	}
	keyLimit, err := finance.LimitsFor(transaction, "key", record.KeyID)
	if err != nil {
		return "storage_unavailable", 503, ""
	}
	bounded := tenantLimit.RequiresBounds() || keyLimit.RequiresBounds()
	if bounded {
		var pending int
		if transaction.QueryRow("SELECT count(*) FROM legacy_pending l WHERE l.tenant_id=? AND l.resolved=0", tenantID).Scan(&pending) != nil {
			return "storage_unavailable", 503, ""
		}
		if pending > 0 {
			return "historical_usage_pending", 503, ""
		}
		if transaction.QueryRow("SELECT count(*) FROM admissions a JOIN request_prices p ON p.request_id=a.request_id WHERE a.tenant_id=? AND a.pending>0 AND (p.max_input=0 OR p.max_output=0)", tenantID).Scan(&pending) != nil {
			return "storage_unavailable", 503, ""
		}
		if pending > 0 {
			return "unbounded_usage_pending", 503, ""
		}
	}
	now := finance.AdmissionClock(transaction, service.now())
	prices := []accounting.PriceSnapshot{}
	for _, candidate := range snapshot.Model.Candidates {
		upstream := snapshot.Upstreams[candidate.UpstreamID]
		if !candidate.Enabled || !upstream.Enabled {
			continue
		}
		value, _, err := finance.PriceAt(transaction, record.Alias, candidate.UpstreamID, upstream.ExecutionRevision, candidate.ExecutionRevision, now)
		var legacy bool
		if transaction.QueryRow("SELECT legacy FROM key_controls WHERE key_id=?", record.KeyID).Scan(&legacy) != nil {
			return "storage_unavailable", 503, ""
		}
		if errors.Is(err, sql.ErrNoRows) && legacy && !bounded {
			value.ID = "legacy_unpriced"
		} else if err != nil {
			return "price_not_ready", 503, ""
		}
		var evidence string
		err = transaction.QueryRow("SELECT max_input,max_output,evidence FROM model_bounds WHERE alias=? AND upstream_id=? AND upstream_execution=? AND candidate_execution=?", record.Alias, candidate.UpstreamID, upstream.ExecutionRevision, candidate.ExecutionRevision).Scan(&value.Input, &value.Output, &evidence)
		if err != nil && (!errors.Is(err, sql.ErrNoRows) || bounded) {
			return "trusted_bound_required", 503, ""
		}
		if bounded {
			if evidence == "" || value.Input <= 0 || value.Output <= 0 {
				return "trusted_bound_required", 503, ""
			}
			var alternatives int64 = 1
			if raw := payload["n"]; raw != nil {
				if json.Unmarshal(raw, &alternatives) != nil || alternatives != 1 {
					return "unsupported_bounded_request", 400, ""
				}
			}
		}
		prices = append(prices, value)
	}
	if len(prices) == 0 {
		return "upstream_unavailable", 503, ""
	}
	if bounded {
		maximumOutput := prices[0].Output
		for _, price := range prices {
			maximumOutput = min(maximumOutput, price.Output)
		}
		if !boundedPayload(payload, maximumOutput) {
			return "unsupported_bounded_request", 400, ""
		}
	}
	tokensReserved, feeReserved, err := accounting.ReserveAttempts(prices)
	if err != nil {
		return "amount_overflow", 503, ""
	}
	if operation != "" {
		if _, err := transaction.Exec("INSERT INTO generation_operations VALUES(?,?,?,?)", record.KeyID, operation, fingerprint, record.ID); err != nil {
			return "storage_unavailable", 503, ""
		}
	}
	period, minute := now.Format("2006-01"), now.Format("2006-01-02T15:04")
	for _, owner := range []struct {
		kind, id string
		limits   accounting.ResourceLimits
	}{{"tenant", tenantID, tenantLimit}, {"key", record.KeyID, keyLimit}} {
		used, err := finance.UsageFor(transaction, owner.kind, owner.id, period, minute)
		if err != nil {
			return "storage_unavailable", 503, ""
		}
		if code := owner.limits.CheckAdmission(used, tokensReserved, feeReserved); code != "" {
			return owner.kind + "_" + code, 429, ""
		}
	}
	if _, err := transaction.Exec("INSERT INTO admissions VALUES(?,?,?,?,?,?,1,?,?,0,1)", record.ID, tenantID, record.KeyID, period, minute, now.Format(time.RFC3339Nano), tokensReserved, feeReserved); err != nil {
		return "storage_unavailable", 503, ""
	}
	for _, price := range prices {
		if _, err := transaction.Exec("INSERT INTO request_prices VALUES(?,?,?,?,?,?,?,?,?)", record.ID, price.Upstream, price.ID, price.CostIn, price.CostOut, price.FeeIn, price.FeeOut, price.Input, price.Output); err != nil {
			return "storage_unavailable", 503, ""
		}
	}
	if transaction.Commit() != nil {
		return "storage_unavailable", 503, ""
	}
	return "", 200, ""
}
