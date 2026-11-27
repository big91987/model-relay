package reporting

import (
	"context"
	"database/sql"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/finance"
	"model-relay/internal/sqlread"
)

func (service *Service) Catalog(ctx context.Context, principal access.Principal, input CatalogInput) ([]CatalogItem, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	if input.Tenant == "" && !principal.Platform {
		return nil, fault.New(fault.Invalid, "tenant_required")
	}
	if input.Tenant != "" {
		if _, err = access.Role(transaction, principal, input.Tenant); err != nil {
			return nil, fault.New(fault.NotFound, "not_found")
		}
	}
	query := "SELECT alias FROM models ORDER BY alias"
	args := []any{}
	if input.IncludeUngranted && !principal.Platform {
		return nil, fault.New(fault.Forbidden, "forbidden")
	}
	if input.Tenant != "" && !input.IncludeUngranted {
		query = "SELECT m.alias FROM models m JOIN tenant_model_grants g ON g.alias=m.alias WHERE g.tenant_id=? ORDER BY m.alias"
		args = append(args, input.Tenant)
	}
	ids, err := sqlread.Strings(transaction, query, args...)
	if err != nil {
		return nil, storage()
	}
	values := []CatalogItem{}
	for _, alias := range ids {
		value := CatalogItem{Alias: alias, Currency: "USD", Unit: "micro_usd_per_million_tokens", Protocol: "chat.completions", Stream: true}
		if transaction.QueryRow("SELECT m.enabled,m.revision,coalesce(p.strategy,'weighted') FROM models m LEFT JOIN model_route_policies p ON p.alias=m.alias WHERE m.alias=?", alias).Scan(&value.Enabled, &value.Revision, &value.Strategy) != nil {
			return nil, storage()
		}
		value.TenantInput, value.TenantOutput, value.Effective, err = finance.CurrentModelRate(transaction, alias, service.now())
		if err != nil && err != sql.ErrNoRows {
			return nil, storage()
		}
		value.Summary, err = ModelSummary(transaction, principal, input.Tenant, alias, service.now())
		if err != nil {
			return nil, storage()
		}
		values = append(values, value)
	}
	return values, nil
}
