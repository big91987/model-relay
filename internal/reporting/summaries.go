package reporting

import (
	"database/sql"
	"errors"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/accounting"
	"model-relay/internal/finance"
	"model-relay/internal/sqlread"
)

func TenantSummary(transaction *sql.Tx, tenantID string, now time.Time) (map[string]any, error) {
	owners, err := sqlread.Strings(transaction, "SELECT u.account FROM memberships m JOIN users u ON u.id=m.user_id WHERE m.tenant_id=? AND m.role='administrator' AND u.active=1 ORDER BY u.account", tenantID)
	if err != nil {
		return nil, err
	}
	models, err := sqlread.Strings(transaction, "SELECT alias FROM tenant_model_grants WHERE tenant_id=? ORDER BY alias", tenantID)
	if err != nil {
		return nil, err
	}
	var members, keys int64
	if err := transaction.QueryRow("SELECT count(*) FROM memberships m JOIN users u ON u.id=m.user_id WHERE m.tenant_id=? AND u.active=1", tenantID).Scan(&members); err != nil {
		return nil, err
	}
	if err := transaction.QueryRow("SELECT count(*) FROM key_controls WHERE tenant_id=?", tenantID).Scan(&keys); err != nil {
		return nil, err
	}
	limits, err := finance.LimitsFor(transaction, "tenant", tenantID)
	if err != nil {
		return nil, err
	}
	now = finance.AdmissionClock(transaction, now)
	usage, err := finance.UsageFor(transaction, "tenant", tenantID, now.Format("2006-01"), now.Format("2006-01-02T15:04"))
	if err != nil {
		return nil, err
	}
	var platform accounting.ResourceLimits
	err = transaction.QueryRow("SELECT calls,rpm,tpm,concurrency,budget,hard,revision FROM platform_limits WHERE tenant_id=?", tenantID).Scan(&platform.Calls, &platform.RPM, &platform.TPM, &platform.Concurrency, &platform.Budget, &platform.Hard, &platform.Revision)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	var cap *accounting.ResourceLimits
	if err == nil {
		cap = &platform
	}
	return map[string]any{"owners": owners, "member_count": members, "models": models, "key_count": keys, "limits": limits, "usage": usage, "platform_limits": cap, "period": now.Format("2006-01")}, nil
}

func ModelSummary(transaction *sql.Tx, principal access.Principal, tenantID, alias string, now time.Time) (map[string]any, error) {
	rows, err := transaction.Query("SELECT c.upstream_id,u.name,coalesce(s.source_kind,'unspecified'),c.enabled,c.priority,c.weight,u.execution_revision,c.execution_revision FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id LEFT JOIN service_profiles s ON s.upstream_id=u.id WHERE c.alias=? ORDER BY c.priority DESC,c.upstream_id", alias)
	if err != nil {
		return nil, err
	}
	type sourceFact struct {
		upstreamID, enabled, priority, weight, upstreamExecution, candidateExecution int64
		name, kind                                                                   string
	}
	facts := []sourceFact{}
	for rows.Next() {
		var fact sourceFact
		if err = rows.Scan(&fact.upstreamID, &fact.name, &fact.kind, &fact.enabled, &fact.priority, &fact.weight, &fact.upstreamExecution, &fact.candidateExecution); err != nil {
			rows.Close()
			return nil, err
		}
		facts = append(facts, fact)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	sources := []map[string]any{}
	for _, fact := range facts {
		value := map[string]any{"upstream_id": fact.upstreamID, "name": fact.name, "source_kind": fact.kind, "enabled": fact.enabled, "priority": fact.priority, "weight": fact.weight, "tenant_input_rate": nil, "tenant_output_rate": nil}
		price, _, err := finance.PriceAt(transaction, alias, fact.upstreamID, fact.upstreamExecution, fact.candidateExecution, now)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			value["tenant_input_rate"], value["tenant_output_rate"] = price.FeeIn, price.FeeOut
		}
		sources = append(sources, value)
	}
	where, args, err := reportScope(transaction, principal, tenantID)
	if err != nil {
		return nil, errors.New("report scope unavailable")
	}
	var latency *float64
	args = append(args, alias)
	err = transaction.QueryRow("SELECT avg(r.duration_ms) FROM requests r LEFT JOIN key_controls c ON c.key_id=r.key_id WHERE "+where+" AND r.alias=? AND r.state!='started' AND r.duration_ms IS NOT NULL AND EXISTS(SELECT 1 FROM request_attempts ra WHERE ra.request_id=r.id)", args...).Scan(&latency)
	if err != nil {
		return nil, err
	}
	return map[string]any{"sources": sources, "average_duration_ms": latency, "latency_scope": "visible_finished_requests_not_service_check"}, nil
}
