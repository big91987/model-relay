package catalog

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"model-relay/internal/routing"
)

// ErrModelMissing identifies an unavailable configured route without exposing storage errors.
var ErrModelMissing = errors.New("model unavailable")

// UpstreamFacts reads persistence without acquiring coordination. The caller
// must hold the shared coordinator's RLock or Lock through its whole operation.
// It returns sql.ErrNoRows unchanged for gateway and maintenance decisions.
func (service *Service) UpstreamFacts(ctx context.Context, id int64) (Upstream, error) {
	var value Upstream
	err := service.db.QueryRowContext(ctx, "SELECT id,name,url,enabled,revision,execution_revision,token FROM upstream WHERE id=?", id).Scan(&value.ID, &value.Name, &value.URL, &value.Enabled, &value.Revision, &value.ExecutionRevision, &value.Encrypted)
	value.CredentialConfigured = len(value.Encrypted) > 0
	if err != nil {
		return value, err
	}
	if service.schema == 3 {
		if err = service.db.QueryRowContext(ctx, "SELECT source_kind FROM service_profiles WHERE upstream_id=?", id).Scan(&value.SourceKind); err != nil {
			return value, err
		}
	}
	var check CheckResult
	err = service.db.QueryRowContext(ctx, "SELECT id,execution_revision,state,started_at,finished_at,duration_ms,error_code,upstream_status FROM upstream_checks WHERE upstream_id=? ORDER BY id DESC LIMIT 1", id).Scan(&check.ID, &check.ExecutionRevision, &check.State, &check.Started, &check.Finished, &check.Duration, &check.Code, &check.Status)
	if err == nil {
		check.Applicable = check.ExecutionRevision == value.ExecutionRevision
		value.Check = &check
	}
	if err == sql.ErrNoRows {
		err = nil
	}
	return value, err
}

// ModelFacts has the same caller-held coordination requirement as UpstreamFacts.
// It returns configuration facts without presentation-only health descriptions.
func (service *Service) ModelFacts(ctx context.Context, alias string) (RouteModel, error) {
	value := RouteModel{Candidates: []Candidate{}}
	err := service.db.QueryRowContext(ctx, "SELECT m.alias,m.enabled,m.revision,m.legacy_remote_model,coalesce(p.strategy,'weighted') FROM models m LEFT JOIN model_route_policies p ON p.alias=m.alias WHERE m.alias=?", alias).Scan(&value.Alias, &value.Enabled, &value.Revision, &value.Legacy, &value.Strategy)
	if err != nil {
		return value, err
	}
	rows, err := service.db.QueryContext(ctx, "SELECT upstream_id,remote_model,enabled,priority,weight,execution_revision FROM route_candidates WHERE alias=? ORDER BY upstream_id", alias)
	if err != nil {
		return value, err
	}
	defer rows.Close()
	for rows.Next() {
		var candidate Candidate
		if err = rows.Scan(&candidate.UpstreamID, &candidate.RemoteModel, &candidate.Enabled, &candidate.Priority, &candidate.Weight, &candidate.ExecutionRevision); err != nil {
			return value, err
		}
		value.Candidates = append(value.Candidates, candidate)
	}
	return value, rows.Err()
}

// ModelSnapshot captures configuration and its runtime epoch under the caller's
// shared coordination lock. Keep that lock through gateway admission/persistence
// so a snapshot cannot be admitted against another configuration generation.
func (service *Service) ModelSnapshot(ctx context.Context, alias string) (Snapshot, error) {
	model, err := service.ModelFacts(ctx, alias)
	if err != nil {
		return Snapshot{}, snapshotError(err)
	}
	snapshot := Snapshot{Model: model, Upstreams: map[int64]Upstream{}, Epoch: service.state.Epoch()}
	if err = service.db.QueryRowContext(ctx, "SELECT config_revision FROM meta").Scan(&snapshot.ConfigRevision); err != nil {
		return snapshot, snapshotError(err)
	}
	for _, candidate := range model.Candidates {
		value, err := service.UpstreamFacts(ctx, candidate.UpstreamID)
		if err != nil {
			return snapshot, snapshotError(err)
		}
		snapshot.Upstreams[candidate.UpstreamID] = value
	}
	return snapshot, nil
}

func (service *Service) ReadUpstream(ctx context.Context, id int64) (Upstream, error) {
	service.state.RLock()
	defer service.state.RUnlock()
	value, err := service.UpstreamFacts(ctx, id)
	return value, readError(err)
}

// ListUpstreams returns up to limit items after the business key. The HTTP
// adapter requests one extra item when it needs to construct a next cursor.
func (service *Service) ListUpstreams(ctx context.Context, after int64, limit int) ([]Upstream, error) {
	service.state.RLock()
	defer service.state.RUnlock()
	rows, err := service.db.QueryContext(ctx, "SELECT id FROM upstream WHERE id>? ORDER BY id LIMIT ?", after, limit)
	if err != nil {
		return nil, unavailable()
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, unavailable()
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, unavailable()
	}
	values := []Upstream{}
	for _, id := range ids {
		value, err := service.UpstreamFacts(ctx, id)
		if err != nil {
			return nil, unavailable()
		}
		values = append(values, value)
	}
	return values, nil
}
func (service *Service) ReadModel(ctx context.Context, alias string) (RouteModel, error) {
	service.state.RLock()
	defer service.state.RUnlock()
	value, err := service.ModelFacts(ctx, alias)
	if err == nil {
		err = service.describeModel(ctx, &value)
	}
	return value, readError(err)
}

// ListModels uses model aliases as its cursor; the adapter owns cursor encoding.
func (service *Service) ListModels(ctx context.Context, after string, limit int) ([]RouteModel, error) {
	service.state.RLock()
	defer service.state.RUnlock()
	rows, err := service.db.QueryContext(ctx, "SELECT alias FROM models WHERE alias>? ORDER BY alias LIMIT ?", after, limit)
	if err != nil {
		return nil, unavailable()
	}
	aliases := []string{}
	for rows.Next() {
		var alias string
		if err = rows.Scan(&alias); err != nil {
			rows.Close()
			return nil, unavailable()
		}
		aliases = append(aliases, alias)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, unavailable()
	}
	values := []RouteModel{}
	for _, alias := range aliases {
		value, err := service.ModelFacts(ctx, alias)
		if err == nil {
			err = service.describeModel(ctx, &value)
		}
		if err != nil {
			return nil, unavailable()
		}
		values = append(values, value)
	}
	return values, nil
}
func (service *Service) describeModel(ctx context.Context, value *RouteModel) error {
	for index := range value.Candidates {
		candidate := &value.Candidates[index]
		source, err := service.UpstreamFacts(ctx, candidate.UpstreamID)
		if err != nil {
			return err
		}
		candidate.GenerationHealth = "未知（本次启动）"
		health := service.router.Health(value.Alias, routing.Candidate{UpstreamID: candidate.UpstreamID, RemoteModel: candidate.RemoteModel, Enabled: candidate.Enabled, Priority: candidate.Priority, Weight: candidate.Weight, ExecutionRevision: candidate.ExecutionRevision, UpstreamExecution: source.ExecutionRevision})
		if health == nil {
			continue
		}
		if health.Cooldown.After(time.Now()) {
			candidate.GenerationHealth = "暂时冷却"
			until := health.Cooldown.UTC().Format(time.RFC3339Nano)
			candidate.CooldownUntil = &until
		} else if !health.Cooldown.IsZero() {
			candidate.GenerationHealth = "待重新尝试"
		} else if health.State == "success" {
			candidate.GenerationHealth = "最近生成成功（非持续健康）"
		} else {
			candidate.GenerationHealth = "最近生成失败（仍参与）"
		}
	}
	return nil
}

// FirstRemoteModel supplies best-effort rejected-request evidence. Like the
// original observation query, absent rows or storage failures leave it unknown.
// This best-effort read owns its coordination, including on rejection paths.
func (service *Service) FirstRemoteModel(ctx context.Context, alias string) string {
	service.state.RLock()
	defer service.state.RUnlock()
	var remote string
	_ = service.db.QueryRowContext(ctx, "SELECT remote_model FROM route_candidates WHERE alias=? ORDER BY upstream_id LIMIT 1", alias).Scan(&remote)
	return remote
}

func snapshotError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrModelMissing
	}
	return err
}
