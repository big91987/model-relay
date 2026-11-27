package identity

import (
	"context"
	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/names"
	"model-relay/internal/operations"
	"model-relay/internal/secure"
	"time"
)

func (service *Service) CreateTenant(ctx context.Context, principal access.Principal, input *TenantInput) (*Tenant, error) {
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	defer transaction.Rollback()
	if !principal.Platform {
		return nil, fault.New(fault.Forbidden, "forbidden")
	}
	if input == nil || !names.Valid(input.Name) {
		return nil, fault.New(fault.Invalid, "invalid_name")
	}
	value := Tenant{ID: secure.Random(16), Name: input.Name, Active: true, Created: time.Now().UTC().Format(time.RFC3339Nano)}
	if _, err := transaction.Exec("INSERT INTO tenants VALUES(?,?,1,?)", value.ID, value.Name, value.Created); err != nil || operations.Audit(transaction, principal.ID, value.ID, "tenant.create", value.ID) != nil || transaction.Commit() != nil {
		return nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	return &value, nil
}
func (service *Service) SetTenantActive(ctx context.Context, principal access.Principal, tenantID string, input *StatusInput) (map[string]bool, error) {
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	defer transaction.Rollback()
	role, err := access.Role(transaction, principal, tenantID)
	if err != nil {
		return nil, fault.New(fault.NotFound, "not_found")
	}
	if role != "administrator" || !principal.Platform {
		return nil, fault.New(fault.Forbidden, "forbidden")
	}
	if input == nil {
		return nil, fault.New(fault.Invalid, "invalid_request")
	}
	if _, err := transaction.Exec("UPDATE tenants SET active=? WHERE id=?", input.Active, tenantID); err != nil || operations.Audit(transaction, principal.ID, tenantID, "tenant.status", tenantID) != nil || transaction.Commit() != nil {
		return nil, fault.New(fault.Unavailable, "storage_unavailable")
	}
	return map[string]bool{"ok": true}, nil
}
