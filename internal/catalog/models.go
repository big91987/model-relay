package catalog

import (
	"context"
	"database/sql"

	"model-relay/internal/fault"
	"model-relay/internal/names"
	"model-relay/internal/operations"
)

func validateModel(input ModelInput) error {
	if input.Enabled == nil || input.Candidates == nil || len(*input.Candidates) > 100 {
		return fault.New(fault.Invalid, "invalid_request")
	}
	if input.Strategy != nil && *input.Strategy != "weighted" && *input.Strategy != "low_latency" {
		return fault.New(fault.Invalid, "invalid_routing_strategy")
	}
	return nil
}
func (service *Service) CreateModel(ctx context.Context, actor string, input ModelInput) (RouteModel, error) {
	if err := validateModel(input); err != nil {
		return RouteModel{}, err
	}
	if !names.Alias(input.Alias) {
		return RouteModel{}, fault.New(fault.Invalid, "invalid_model")
	}
	service.state.Lock()
	defer service.state.Unlock()
	_, err := service.ModelFacts(ctx, input.Alias)
	if err == nil {
		return RouteModel{}, fault.New(fault.Conflict, "alias_conflict")
	}
	if err != sql.ErrNoRows {
		return RouteModel{}, unavailable()
	}
	if err = service.validateCandidates(ctx, *input.Candidates); err != nil {
		return RouteModel{}, err
	}
	return service.writeModel(ctx, actor, input.Alias, RouteModel{}, input)
}
func (service *Service) UpdateModel(ctx context.Context, actor, alias string, input ModelInput) (RouteModel, error) {
	if err := validateModel(input); err != nil {
		return RouteModel{}, err
	}
	if input.Expected == nil || input.Alias != "" {
		return RouteModel{}, fault.New(fault.Invalid, "invalid_request")
	}
	if !names.Alias(alias) {
		return RouteModel{}, fault.New(fault.Invalid, "invalid_model")
	}
	service.state.Lock()
	defer service.state.Unlock()
	current, err := service.ModelFacts(ctx, alias)
	if err != nil {
		return RouteModel{}, readError(err)
	}
	if current.Revision != *input.Expected {
		return RouteModel{}, &RevisionConflict{CurrentRevision: current.Revision}
	}
	if err = service.validateCandidates(ctx, *input.Candidates); err != nil {
		return RouteModel{}, err
	}
	return service.writeModel(ctx, actor, alias, current, input)
}
func (service *Service) validateCandidates(ctx context.Context, candidates []CandidateInput) error {
	seen := map[int64]bool{}
	for _, candidate := range candidates {
		if candidate.UpstreamID < 1 || seen[candidate.UpstreamID] || !names.Valid(candidate.RemoteModel) || candidate.Enabled == nil || candidate.Priority == nil || candidate.Weight == nil || *candidate.Priority < 0 || *candidate.Priority > 100 || *candidate.Weight < 1 || *candidate.Weight > 100 {
			return fault.New(fault.Invalid, "invalid_route_candidate")
		}
		seen[candidate.UpstreamID] = true
		var exists bool
		if err := service.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM upstream WHERE id=?)", candidate.UpstreamID).Scan(&exists); err != nil {
			return unavailable()
		}
		if !exists {
			return fault.New(fault.Invalid, "invalid_route_candidate")
		}
	}
	return nil
}

// writeModel preserves the per-candidate execution generation independently of
// public routing edits. Only a changed remote model invalidates matching health.
func (service *Service) writeModel(ctx context.Context, actor, alias string, current RouteModel, input ModelInput) (RouteModel, error) {
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return RouteModel{}, unavailable()
	}
	defer transaction.Rollback()
	if current.Alias == "" {
		_, err = transaction.ExecContext(ctx, "INSERT INTO models(alias,enabled,revision) VALUES(?,?,1)", alias, *input.Enabled)
	} else {
		_, err = transaction.ExecContext(ctx, "UPDATE models SET enabled=?,revision=revision+1,legacy_remote_model=CASE WHEN ? THEN NULL ELSE legacy_remote_model END WHERE alias=?", *input.Enabled, len(*input.Candidates) > 0, alias)
	}
	if err != nil {
		return RouteModel{}, unavailable()
	}
	strategy := current.Strategy
	if strategy == "" {
		strategy = "weighted"
	}
	if input.Strategy != nil {
		strategy = *input.Strategy
	}
	if _, err = transaction.ExecContext(ctx, "INSERT INTO model_route_policies VALUES(?,?) ON CONFLICT(alias) DO UPDATE SET strategy=excluded.strategy", alias, strategy); err != nil {
		return RouteModel{}, unavailable()
	}
	if _, err = transaction.ExecContext(ctx, "DELETE FROM route_candidates WHERE alias=?", alias); err != nil {
		return RouteModel{}, unavailable()
	}
	for _, candidate := range *input.Candidates {
		execution := int64(1)
		for _, previous := range current.Candidates {
			if previous.UpstreamID == candidate.UpstreamID {
				execution = previous.ExecutionRevision
				if previous.RemoteModel != candidate.RemoteModel {
					execution++
				}
			}
		}
		if _, err = transaction.ExecContext(ctx, "INSERT INTO route_candidates VALUES(?,?,?,?,?,?,?)", alias, candidate.UpstreamID, candidate.RemoteModel, *candidate.Enabled, *candidate.Priority, *candidate.Weight, execution); err != nil {
			return RouteModel{}, unavailable()
		}
	}
	if service.schema == 3 {
		if err = operations.Audit(transaction, actor, "", "model.write", alias); err != nil {
			return RouteModel{}, unavailable()
		}
	}
	if configChanged(transaction) != nil || transaction.Commit() != nil {
		return RouteModel{}, unavailable()
	}
	for _, candidate := range current.Candidates {
		for _, replacement := range *input.Candidates {
			if candidate.UpstreamID == replacement.UpstreamID && candidate.RemoteModel != replacement.RemoteModel {
				service.router.ForgetCandidate(alias, candidate.UpstreamID)
			}
		}
	}
	service.committed()
	value, err := service.ModelFacts(ctx, alias)
	return value, readError(err)
}
