package reporting

import (
	"context"
	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/identity"
)

// Tenant adds the authorized management summary to the tenant facts.
type Tenant struct {
	identity.Tenant
	Summary map[string]any `json:"summary,omitempty"`
}

func (service *Service) Tenants(ctx context.Context, principal access.Principal) ([]Tenant, error) {
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	defer transaction.Rollback()
	query := "SELECT t.id,t.name,t.active,t.created,'administrator' FROM tenants t ORDER BY t.created,t.id"
	arguments := []any{}
	if !principal.Platform {
		query = "SELECT t.id,t.name,t.active,t.created,m.role FROM tenants t JOIN memberships m ON m.tenant_id=t.id WHERE m.user_id=? ORDER BY t.created,t.id"
		arguments = append(arguments, principal.ID)
	}
	rows, err := transaction.Query(query, arguments...)
	if err != nil {
		return nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	defer rows.Close()
	values := []Tenant{}
	for rows.Next() {
		var value Tenant
		if rows.Scan(&value.ID, &value.Name, &value.Active, &value.Created, &value.Role) != nil {
			return nil, fault.New(fault.Unavailable, "storage_unavailable")
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	rows.Close()
	for index := range values {
		if !principal.Platform && values[index].Role != "administrator" {
			continue
		}
		values[index].Summary, err = TenantSummary(transaction, values[index].ID, service.now())
		if err != nil {
			return nil, fault.New(fault.Unavailable, "storage_unavailable")
		}
	}
	return values, nil
}
