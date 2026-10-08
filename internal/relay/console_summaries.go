package relay

import (
	"database/sql"
	"time"
)

func tenantSummary(transaction *sql.Tx, tenantID string, now time.Time) (map[string]any, error) {
	owners, err := stringList(transaction, "SELECT u.account FROM memberships m JOIN users u ON u.id=m.user_id WHERE m.tenant_id=? AND m.role='administrator' AND u.active=1 ORDER BY u.account", tenantID)
	if err != nil {
		return nil, err
	}
	models, err := stringList(transaction, "SELECT alias FROM tenant_model_grants WHERE tenant_id=? ORDER BY alias", tenantID)
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
	limits, err := limitsFor(transaction, "tenant", tenantID)
	if err != nil {
		return nil, err
	}
	now = admissionClock(transaction, now)
	usage, err := usageFor(transaction, "tenant", tenantID, now.Format("2006-01"), now.Format("2006-01-02T15:04"))
	if err != nil {
		return nil, err
	}
	var platform ResourceLimits
	err = transaction.QueryRow("SELECT calls,rpm,tpm,concurrency,budget,hard,revision FROM platform_limits WHERE tenant_id=?", tenantID).Scan(&platform.Calls, &platform.RPM, &platform.TPM, &platform.Concurrency, &platform.Budget, &platform.Hard, &platform.Revision)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	var cap *ResourceLimits
	if err == nil {
		cap = &platform
	}
	return map[string]any{"owners": owners, "member_count": members, "models": models, "key_count": keys, "limits": limits, "usage": usage, "platform_limits": cap, "period": now.Format("2006-01")}, nil
}

func modelSummary(transaction *sql.Tx, principal Principal, tenantID, alias string, now time.Time) (map[string]any, error) {
	sources, err := queryMaps(transaction, `SELECT c.upstream_id,u.name,coalesce(s.source_kind,'unspecified') AS source_kind,c.enabled,c.priority,c.weight,
(SELECT p.tenant_input FROM prices p WHERE p.alias=c.alias AND p.upstream_id=c.upstream_id AND p.upstream_execution=u.execution_revision AND p.candidate_execution=c.execution_revision AND julianday(p.effective)<=julianday(?) ORDER BY julianday(p.effective) DESC,p.id DESC LIMIT 1) AS tenant_input_rate,
(SELECT p.tenant_output FROM prices p WHERE p.alias=c.alias AND p.upstream_id=c.upstream_id AND p.upstream_execution=u.execution_revision AND p.candidate_execution=c.execution_revision AND julianday(p.effective)<=julianday(?) ORDER BY julianday(p.effective) DESC,p.id DESC LIMIT 1) AS tenant_output_rate
FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id LEFT JOIN service_profiles s ON s.upstream_id=u.id WHERE c.alias=? ORDER BY c.priority DESC,c.upstream_id`, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), alias)
	if err != nil {
		return nil, err
	}
	where, args, _ := reportScope(transaction, principal, tenantID)
	var latency *float64
	args = append(args, alias)
	err = transaction.QueryRow("SELECT avg(r.duration_ms) FROM requests r LEFT JOIN key_controls c ON c.key_id=r.key_id WHERE "+where+" AND r.alias=? AND r.state!='started' AND r.duration_ms IS NOT NULL AND EXISTS(SELECT 1 FROM request_attempts ra WHERE ra.request_id=r.id)", args...).Scan(&latency)
	if err != nil {
		return nil, err
	}
	return map[string]any{"sources": sources, "average_duration_ms": latency, "latency_scope": "visible_finished_requests_not_service_check"}, nil
}
