package finance

import (
	"context"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
	"model-relay/internal/secure"
)

func (service *Service) Prices(ctx context.Context, principal access.Principal, alias string) ([]Price, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	if err = modelExists(transaction, principal, alias); err != nil {
		return nil, err
	}
	rows, err := transaction.Query("SELECT id,upstream_id,effective,cost_input,cost_output,tenant_input,tenant_output FROM prices WHERE alias=? ORDER BY effective DESC,id DESC", alias)
	if err != nil {
		return nil, storage()
	}
	defer rows.Close()
	values := []Price{}
	for rows.Next() {
		value := Price{Currency: "USD", Unit: "micro_usd_per_million_tokens"}
		if rows.Scan(&value.Price.ID, &value.Price.Upstream, &value.Effective, &value.Price.CostIn, &value.Price.CostOut, &value.Price.FeeIn, &value.Price.FeeOut) != nil {
			return nil, storage()
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return nil, storage()
	}
	return values, nil
}

func (service *Service) PublishPrices(ctx context.Context, principal access.Principal, alias, key, path string, input *PublishPricesInput) (operations.Result, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return operations.Result{}, err
	}
	defer transaction.Rollback()
	if err = modelExists(transaction, principal, alias); err != nil {
		return operations.Result{}, err
	}
	if input == nil || len(input.Prices) == 0 || len(input.Prices) > 100 {
		return operations.Result{}, fault.New(fault.Invalid, "invalid_prices")
	}
	effectiveText := input.Effective
	if effectiveText == "" {
		effectiveText = service.now().Format(time.RFC3339Nano)
	}
	effective, err := time.Parse(time.RFC3339Nano, effectiveText)
	if err != nil {
		return operations.Result{}, fault.New(fault.Invalid, "invalid_effective_at")
	}
	operation, prior, fingerprint, err := beginOperation(transaction, principal, key, path, "price.publish", input)
	if err != nil {
		return operations.Result{}, err
	}
	if prior != "" {
		return replay(prior), nil
	}
	if effective.Before(service.now().Add(-time.Minute)) {
		return operations.Result{}, fault.New(fault.Invalid, "invalid_effective_at")
	}
	seen := map[int64]bool{}
	var candidates int
	if transaction.QueryRow("SELECT count(*) FROM route_candidates WHERE alias=?", alias).Scan(&candidates) != nil {
		return operations.Result{}, storage()
	}
	if candidates != len(input.Prices) {
		return operations.Result{}, fault.New(fault.Invalid, "all_candidate_prices_required")
	}
	for _, price := range input.Prices {
		if price.FeeIn != input.Prices[0].FeeIn || price.FeeOut != input.Prices[0].FeeOut {
			return operations.Result{}, fault.New(fault.Invalid, "tenant_rates_must_match")
		}
		if seen[price.Upstream] {
			return operations.Result{}, fault.New(fault.Invalid, "duplicate_price")
		}
		seen[price.Upstream] = true
		for _, rate := range []int64{price.CostIn, price.CostOut, price.FeeIn, price.FeeOut} {
			if rate < 0 || rate > 1000000000000 {
				return operations.Result{}, fault.New(fault.Invalid, "invalid_rate")
			}
		}
		price.ID = secure.Random(16)
		var sourceExecution, candidateExecution int64
		if transaction.QueryRow("SELECT u.execution_revision,c.execution_revision FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id WHERE c.alias=? AND c.upstream_id=?", alias, price.Upstream).Scan(&sourceExecution, &candidateExecution) != nil {
			return operations.Result{}, fault.New(fault.Invalid, "route_candidate_required")
		}
		if _, err := transaction.Exec("INSERT INTO prices VALUES(?,?,?,?,?,?,?,?,?,?)", price.ID, alias, price.Upstream, effective.UTC().Format(time.RFC3339Nano), price.CostIn, price.CostOut, price.FeeIn, price.FeeOut, sourceExecution, candidateExecution); err != nil {
			return operations.Result{}, storage()
		}
	}
	if err = completeOperation(transaction, principal, operation, fingerprint, "price.publish", alias, ""); err != nil {
		return operations.Result{}, err
	}
	return operations.Result{Created: true, Value: struct {
		ID       string `json:"id"`
		Currency string `json:"currency"`
		Unit     string `json:"unit"`
	}{alias, "USD", "micro_usd_per_million_tokens"}}, nil
}

func (service *Service) Bounds(ctx context.Context, principal access.Principal, alias string) ([]Bound, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	if err = modelExists(transaction, principal, alias); err != nil {
		return nil, err
	}
	rows, err := transaction.Query("SELECT b.upstream_id,b.max_input,b.max_output,b.evidence,b.upstream_execution,b.candidate_execution,(b.upstream_execution=u.execution_revision AND b.candidate_execution=c.execution_revision) AS applicable FROM model_bounds b JOIN upstream u ON u.id=b.upstream_id JOIN route_candidates c ON c.alias=b.alias AND c.upstream_id=b.upstream_id WHERE b.alias=? ORDER BY b.upstream_id", alias)
	if err != nil {
		return nil, storage()
	}
	defer rows.Close()
	values := []Bound{}
	for rows.Next() {
		var value Bound
		if rows.Scan(&value.Upstream, &value.Input, &value.Output, &value.Evidence, &value.UpstreamExecution, &value.CandidateExecution, &value.Applicable) != nil {
			return nil, storage()
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return nil, storage()
	}
	return values, nil
}

func (service *Service) DeclareBound(ctx context.Context, principal access.Principal, alias string, input *BoundInput) (operations.Result, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return operations.Result{}, err
	}
	defer transaction.Rollback()
	if err = modelExists(transaction, principal, alias); err != nil {
		return operations.Result{}, err
	}
	if input == nil || !validBound(*input) {
		return operations.Result{}, fault.New(fault.Invalid, "invalid_bound")
	}
	var sourceExecution, candidateExecution int64
	if transaction.QueryRow("SELECT u.execution_revision,c.execution_revision FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id WHERE c.alias=? AND c.upstream_id=?", alias, input.Upstream).Scan(&sourceExecution, &candidateExecution) != nil {
		return operations.Result{}, fault.New(fault.Invalid, "route_candidate_required")
	}
	if _, err := transaction.Exec("INSERT INTO model_bounds VALUES(?,?,?,?,?,?,?) ON CONFLICT(alias,upstream_id) DO UPDATE SET max_input=excluded.max_input,max_output=excluded.max_output,evidence=excluded.evidence,upstream_execution=excluded.upstream_execution,candidate_execution=excluded.candidate_execution", alias, input.Upstream, input.Input, input.Output, input.Evidence, sourceExecution, candidateExecution); err != nil {
		return operations.Result{}, storage()
	}
	if operations.Audit(transaction, principal.ID, "", "model.bound", alias) != nil || transaction.Commit() != nil {
		return operations.Result{}, storage()
	}
	return ok(), nil
}
