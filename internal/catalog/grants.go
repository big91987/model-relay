package catalog

import (
	"context"
	"database/sql"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
	"model-relay/internal/sqlread"
)

func grantModel(transaction *sql.Tx, principal access.Principal, alias string) error {
	if !principal.Platform {
		return fault.New(fault.Forbidden, "forbidden")
	}
	var count int
	if transaction.QueryRow("SELECT count(*) FROM models WHERE alias=?", alias).Scan(&count) != nil || count != 1 {
		return fault.New(fault.NotFound, "not_found")
	}
	return nil
}
func (service *Service) Grants(ctx context.Context, principal access.Principal, alias string) ([]string, error) {
	service.state.RLock()
	defer service.state.RUnlock()
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, unavailable()
	}
	defer transaction.Rollback()
	if err = grantModel(transaction, principal, alias); err != nil {
		return nil, err
	}
	values, err := sqlread.Strings(transaction, "SELECT tenant_id FROM tenant_model_grants WHERE alias=? ORDER BY tenant_id", alias)
	if err != nil {
		return nil, unavailable()
	}
	return values, nil
}
func (service *Service) SetGrants(ctx context.Context, principal access.Principal, alias string, tenants []string) error {
	service.state.Lock()
	defer service.state.Unlock()
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return unavailable()
	}
	defer transaction.Rollback()
	if err = grantModel(transaction, principal, alias); err != nil {
		return err
	}
	if len(tenants) > 1000 {
		return fault.New(fault.Invalid, "invalid_grants")
	}
	seen := map[string]bool{}
	for _, tenant := range tenants {
		var active bool
		if seen[tenant] || transaction.QueryRowContext(ctx, "SELECT active FROM tenants WHERE id=?", tenant).Scan(&active) != nil || !active {
			return fault.New(fault.Invalid, "invalid_grants")
		}
		seen[tenant] = true
	}
	if _, err = transaction.ExecContext(ctx, "DELETE FROM tenant_model_grants WHERE alias=?", alias); err != nil {
		return unavailable()
	}
	for _, tenant := range tenants {
		if _, err = transaction.ExecContext(ctx, "INSERT INTO tenant_model_grants VALUES(?,?)", tenant, alias); err != nil {
			return unavailable()
		}
	}
	if operations.Audit(transaction, principal.ID, "", "model.grants", alias) != nil || transaction.Commit() != nil {
		return unavailable()
	}
	return nil
}
