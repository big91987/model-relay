package reporting

import (
	"context"
	"database/sql"

	"model-relay/internal/fault"
	"model-relay/internal/records"
)

// LegacyOverview reads the platform-wide inventory and verifies the caller's
// current-run success evidence in one snapshot. clearSuccess instructs the caller
// to invalidate a known stale ID. Unassessed storage failures leave it false;
// a commit failure retains an already determined invalidation instruction.
func (service *Service) LegacyOverview(ctx context.Context, successID string, degraded bool) (overview map[string]any, clearSuccess bool, err error) {
	transaction, err := service.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, false, fault.New(fault.Unavailable, "storage_unavailable")
	}
	defer transaction.Rollback()
	var enabled, credential bool
	var revision int64
	var upstreamTotal, upstreamEnabled, ready int64
	err = transaction.QueryRowContext(ctx, "SELECT count(*),coalesce(sum(enabled),0),coalesce(max(length(token)>0),0) FROM upstream").Scan(&upstreamTotal, &upstreamEnabled, &credential)
	configured := upstreamTotal > 0
	enabled = upstreamEnabled > 0
	if err != nil {
		return nil, false, fault.New(fault.Unavailable, "storage_unavailable")
	}
	if transaction.QueryRow("SELECT config_revision FROM meta").Scan(&revision) != nil || transaction.QueryRow("SELECT count(*) FROM models WHERE enabled=1 AND EXISTS(SELECT 1 FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id WHERE c.alias=models.alias AND c.enabled=1 AND u.enabled=1 AND length(u.token)>0)").Scan(&ready) != nil {
		return nil, false, fault.New(fault.Unavailable, "storage_unavailable")
	}
	type counts struct {
		Total   int64 `json:"total"`
		Enabled int64 `json:"enabled"`
	}
	var models, keys counts
	for _, query := range []struct {
		table string
		value *counts
	}{{"models", &models}, {"app_keys", &keys}} {
		if transaction.QueryRowContext(ctx, "SELECT count(*),coalesce(sum(enabled),0) FROM "+query.table).Scan(&query.value.Total, &query.value.Enabled) != nil {
			return nil, false, fault.New(fault.Unavailable, "storage_unavailable")
		}
	}
	var evidence any
	if successID != "" {
		value, queryErr := records.Scan(transaction.QueryRowContext(ctx, "SELECT "+records.Columns+" FROM requests WHERE id=?", successID))
		var matches bool
		if queryErr == nil {
			queryErr = transaction.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM requests r JOIN models m ON m.alias=r.alias JOIN app_keys k ON k.id=r.key_id JOIN route_candidates c ON c.alias=m.alias AND c.upstream_id=r.successful_upstream_id JOIN upstream u ON u.id=c.upstream_id WHERE r.id=? AND m.enabled=1 AND k.enabled=1 AND c.enabled=1 AND u.enabled=1 AND r.config_revision=? AND c.remote_model=r.model AND u.revision=r.revision)", value.ID, revision).Scan(&matches)
		}
		if queryErr != nil && queryErr != sql.ErrNoRows {
			return nil, false, fault.New(fault.Unavailable, "storage_unavailable")
		}
		if queryErr == nil && value.State == "success" && enabled && matches {
			evidence = map[string]any{"request_id": value.ID, "time": value.Time, "alias": value.Alias, "model": value.Model, "key_id": value.KeyID, "upstream_revision": value.Revision, "scope": "本次服务启动、当前配置的一次已记录成功"}
		} else {
			clearSuccess = true
		}
	}
	if transaction.Commit() != nil {
		return nil, clearSuccess, fault.New(fault.Unavailable, "storage_unavailable")
	}
	return map[string]any{"upstream": map[string]any{"configured": configured, "enabled": enabled, "credential_configured": credential, "revision": revision, "total": upstreamTotal, "enabled_count": upstreamEnabled}, "models": models, "keys": keys, "callable_models": ready, "call_evidence": evidence, "degraded": degraded}, clearSuccess, nil
}
