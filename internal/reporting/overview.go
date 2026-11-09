package reporting

import (
	"context"
	"strconv"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/sqlread"
)

func (service *Service) Overview(ctx context.Context, principal access.Principal, tenant string) (Overview, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return Overview{}, err
	}
	defer transaction.Rollback()
	where, args, err := reportScope(transaction, principal, tenant)
	if err != nil {
		return Overview{}, err
	}
	now := service.now()
	today, month := now.Format("2006-01-02"), now.Format("2006-01")
	base := requestsFrom + where
	var count, success, pending, reserved, legacyPending, fee int64
	var latency *float64
	arguments := append([]any{today, today}, args...)
	if transaction.QueryRow("SELECT coalesce(sum(CASE WHEN substr(r.time,1,10)=? THEN 1 ELSE 0 END),0),coalesce(sum(CASE WHEN substr(r.time,1,10)=? AND r.state='success' THEN 1 ELSE 0 END),0),avg(r.duration_ms),coalesce(sum(a.pending),0),coalesce(sum(a.fee_reserved),0)"+base, arguments...).Scan(&count, &success, &latency, &pending, &reserved) != nil {
		return Overview{}, storage()
	}
	if transaction.QueryRow("SELECT count(*)"+base+" AND EXISTS(SELECT 1 FROM legacy_pending l WHERE l.request_id=r.id AND l.resolved=0)", args...).Scan(&legacyPending) != nil {
		return Overview{}, storage()
	}
	arguments = append([]any{month}, args...)
	if transaction.QueryRow("SELECT coalesce(sum((SELECT coalesce(sum(e.fee),0) FROM charge_entries e WHERE e.request_id=r.id AND e.period=?)),0)"+base, arguments...).Scan(&fee) != nil {
		return Overview{}, storage()
	}
	trend, err := sqlread.Maps(transaction, "SELECT substr(r.time,1,10) AS day,count(*) AS requests,sum(CASE WHEN r.state='success' THEN 1 ELSE 0 END) AS successes"+base+" AND substr(r.time,1,10)>=? GROUP BY substr(r.time,1,10) ORDER BY day", append(append([]any{}, args...), now.AddDate(0, 0, -6).Format("2006-01-02"))...)
	if err != nil {
		return Overview{}, storage()
	}
	arguments = append([]any{month}, args...)
	distribution, err := sqlread.Maps(transaction, "SELECT c.tenant_id,t.name,coalesce(sum((SELECT coalesce(sum(e.fee),0) FROM charge_entries e WHERE e.request_id=r.id AND e.period=?)),0) AS fee_micro_usd,coalesce(sum(a.pending),0) AS pending FROM requests r LEFT JOIN key_controls c ON c.key_id=r.key_id LEFT JOIN admissions a ON a.request_id=r.id LEFT JOIN tenants t ON t.id=c.tenant_id WHERE "+where+" AND c.tenant_id IS NOT NULL GROUP BY c.tenant_id,t.name ORDER BY fee_micro_usd DESC,c.tenant_id", arguments...)
	if err != nil {
		return Overview{}, storage()
	}
	var successRate *float64
	if count > 0 {
		rate := float64(success) * 100 / float64(count)
		successRate = &rate
	}
	return Overview{Today: today, Period: month, Timezone: "UTC", RequestsToday: count, SuccessRate: successRate, AverageDuration: latency, Fee: strconv.FormatInt(fee, 10), Pending: pending, LegacyPending: legacyPending, Reserved: strconv.FormatInt(reserved, 10), Trend: trend, Distribution: distribution, Currency: "USD", UnknownIsNotZero: true, ObservedAt: now.Format(time.RFC3339Nano)}, nil
}
