package reporting

import (
	"context"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/sqlread"
)

func (service *Service) Audit(ctx context.Context, principal access.Principal, tenant string) (AuditPage, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return AuditPage{}, err
	}
	defer transaction.Rollback()
	if _, _, err = reportScope(transaction, principal, tenant); err != nil {
		return AuditPage{}, err
	}
	if tenant != "" {
		role, _ := access.Role(transaction, principal, tenant)
		if role != "administrator" {
			return AuditPage{}, fault.New(fault.Forbidden, "forbidden")
		}
	}
	query := "SELECT id,time,actor,tenant_id,action,object_id FROM console_audit"
	args := []any{}
	if tenant != "" {
		query += " WHERE tenant_id=?"
		args = append(args, tenant)
	}
	query += " ORDER BY time DESC,id DESC LIMIT 100"
	values, err := sqlread.Maps(transaction, query, args...)
	if err != nil {
		return AuditPage{}, storage()
	}
	return AuditPage{Items: values, Size: 100, MoreRequiresFilter: len(values) == 100}, nil
}

func (service *Service) Operation(ctx context.Context, principal access.Principal, key string) (OperationView, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return OperationView{}, err
	}
	defer transaction.Rollback()
	var value OperationView
	var tenant string
	if transaction.QueryRow("SELECT object_id,action,tenant_id FROM console_operations WHERE actor=? AND operation=?", principal.ID, key).Scan(&value.ObjectID, &value.Action, &tenant) != nil {
		return OperationView{}, fault.New(fault.NotFound, "not_found")
	}
	if tenant != "" {
		if _, err = access.Role(transaction, principal, tenant); err != nil {
			return OperationView{}, fault.New(fault.NotFound, "not_found")
		}
	} else if !principal.Platform {
		return OperationView{}, fault.New(fault.Forbidden, "forbidden")
	}
	return value, nil
}
