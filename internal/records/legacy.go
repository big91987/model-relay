package records

import (
	"context"
	"database/sql"
	"strconv"

	"model-relay/internal/fault"
	"model-relay/internal/names"
	"model-relay/internal/routing"
)

func (store *Store) LegacyList(ctx context.Context, cursor int64, limit int) ([]Record, any, error) {
	rows, err := store.DB.QueryContext(ctx, "SELECT rowid,"+Columns+" FROM requests WHERE rowid<? ORDER BY rowid DESC LIMIT ?", cursor, limit+1)
	if err != nil {
		return nil, nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	defer rows.Close()
	list := []Record{}
	var next any
	var lastRow int64
	for rows.Next() {
		var value Record
		var rowID int64
		if err = rows.Scan(&rowID, &value.ID, &value.Time, &value.KeyID, &value.Alias, &value.Model, &value.Stream, &value.Revision, &value.State, &value.Code, &value.Status, &value.Duration, &value.Prompt, &value.Completion, &value.Total); err != nil {
			return nil, nil, fault.New(fault.Unavailable, "storage_unavailable")
		}
		if len(list) == limit {
			next = strconv.FormatInt(lastRow, 10)
			break
		}
		list = append(list, value)
		lastRow = rowID
	}
	if rows.Err() != nil {
		return nil, nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	return list, next, nil
}

func (store *Store) Detail(ctx context.Context, id string) (Record, error) {
	if len(id) != 22 || !names.Alias(id) {
		return Record{}, fault.New(fault.NotFound, "not_found")
	}
	value, err := Scan(store.DB.QueryRowContext(ctx, "SELECT "+Columns+" FROM requests WHERE id=?", id))
	if err == sql.ErrNoRows {
		return Record{}, fault.New(fault.NotFound, "not_found")
	}
	if err != nil {
		return Record{}, fault.New(fault.Unavailable, "storage_unavailable")
	}
	if err = store.DB.QueryRow("SELECT evidence_version,coalesce(config_revision,0),coalesce(model_revision,0),successful_upstream_id,committed_at,attempts_complete FROM requests WHERE id=?", id).Scan(&value.EvidenceVersion, &value.ConfigRevision, &value.ModelRevision, &value.SuccessfulUpstream, &value.CommittedAt, &value.AttemptsComplete); err != nil {
		return Record{}, fault.New(fault.Unavailable, "storage_unavailable")
	}
	if value.EvidenceVersion != 2 {
		return value, nil
	}
	value.Attempts, err = ReadAttempts(store.DB, id)
	if err != nil {
		return Record{}, fault.New(fault.Unavailable, "storage_unavailable")
	}
	rows, err := store.DB.Query("SELECT selection_round,upstream_id,snapshot_enabled,snapshot_priority,snapshot_weight,reason,observed_at,cooldown_until FROM request_route_decisions WHERE request_id=? ORDER BY selection_round,upstream_id", id)
	if err != nil {
		return Record{}, fault.New(fault.Unavailable, "storage_unavailable")
	}
	value.Decisions = []routing.RouteDecision{}
	for rows.Next() {
		var decision routing.RouteDecision
		if rows.Scan(&decision.Round, &decision.UpstreamID, &decision.Enabled, &decision.Priority, &decision.Weight, &decision.Reason, &decision.Observed, &decision.Cooldown) != nil {
			rows.Close()
			return Record{}, fault.New(fault.Unavailable, "storage_unavailable")
		}
		value.Decisions = append(value.Decisions, decision)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return Record{}, fault.New(fault.Unavailable, "storage_unavailable")
	}
	value.RoutingEvidence, value.LatencyEvidence, err = ReadRoutingEvidence(store.DB, id)
	if err != nil {
		return Record{}, fault.New(fault.Unavailable, "storage_unavailable")
	}
	value.UsageSummary = UsageParts(value.Attempts, value.AttemptsComplete)
	return value, nil
}
