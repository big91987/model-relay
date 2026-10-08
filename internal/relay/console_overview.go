package relay

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
)

func (server *Server) consoleOverview(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal) {
	if request.Method != http.MethodGet {
		failure(writer, 405, "method_not_allowed", "")
		return
	}
	tenantID := request.URL.Query().Get("tenant")
	where, args, status := reportScope(transaction, principal, tenantID)
	if status != 200 {
		failure(writer, status, "scope_unavailable", "")
		return
	}
	now := server.Store.now()
	base := " FROM requests r LEFT JOIN key_controls c ON c.key_id=r.key_id LEFT JOIN admissions a ON a.request_id=r.id WHERE " + where
	today, month := now.Format("2006-01-02"), now.Format("2006-01")
	var count, success, pending, reserved int64
	var legacyPending int64
	var latency *float64
	arguments := append([]any{today, today}, args...)
	if transaction.QueryRow("SELECT coalesce(sum(CASE WHEN substr(r.time,1,10)=? THEN 1 ELSE 0 END),0),coalesce(sum(CASE WHEN substr(r.time,1,10)=? AND r.state='success' THEN 1 ELSE 0 END),0),avg(r.duration_ms),coalesce(sum(a.pending),0),coalesce(sum(a.fee_reserved),0)"+base, arguments...).Scan(&count, &success, &latency, &pending, &reserved) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if transaction.QueryRow("SELECT count(*)"+base+" AND EXISTS(SELECT 1 FROM legacy_pending l WHERE l.request_id=r.id AND l.resolved=0)", args...).Scan(&legacyPending) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	var fee int64
	arguments = append([]any{month}, args...)
	if transaction.QueryRow("SELECT coalesce(sum((SELECT coalesce(sum(e.fee),0) FROM charge_entries e WHERE e.request_id=r.id AND e.period=?)),0)"+base, arguments...).Scan(&fee) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	arguments = append([]any{now.AddDate(0, 0, -6).Format("2006-01-02")}, args...)
	trend, err := queryMaps(transaction, "SELECT substr(r.time,1,10) AS day,count(*) AS requests,sum(CASE WHEN r.state='success' THEN 1 ELSE 0 END) AS successes"+base+" AND substr(r.time,1,10)>=? GROUP BY substr(r.time,1,10) ORDER BY day", append(append([]any{}, args...), now.AddDate(0, 0, -6).Format("2006-01-02"))...)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	arguments = append([]any{month}, args...)
	distribution, err := queryMaps(transaction, "SELECT c.tenant_id,t.name,coalesce(sum((SELECT coalesce(sum(e.fee),0) FROM charge_entries e WHERE e.request_id=r.id AND e.period=?)),0) AS fee_micro_usd,coalesce(sum(a.pending),0) AS pending FROM requests r LEFT JOIN key_controls c ON c.key_id=r.key_id LEFT JOIN admissions a ON a.request_id=r.id LEFT JOIN tenants t ON t.id=c.tenant_id WHERE "+where+" AND c.tenant_id IS NOT NULL GROUP BY c.tenant_id,t.name ORDER BY fee_micro_usd DESC,c.tenant_id", arguments...)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	var successRate *float64
	if count > 0 {
		rate := float64(success) * 100 / float64(count)
		successRate = &rate
	}
	reply(writer, 200, map[string]any{"today": today, "period": month, "timezone": "UTC", "requests_today": count, "success_rate_percent": successRate, "average_duration_ms": latency, "confirmed_fee_micro_usd": strconv.FormatInt(fee, 10), "pending_attempts": pending, "legacy_pending_requests": legacyPending, "reserved_micro_usd": strconv.FormatInt(reserved, 10), "trend": trend, "tenant_distribution": distribution, "currency": "USD", "unknown_is_not_zero": true, "observed_at": now.Format(time.RFC3339Nano)})
}
