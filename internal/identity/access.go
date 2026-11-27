package identity

import (
	"context"
	"database/sql"

	"model-relay/internal/access"
)

type TenantAccess struct {
	Write          bool
	Invitations    bool
	RemoveMemberID string
}

// CheckTenantAccess enforces scoped management access, including self removal.
func (service *Service) CheckTenantAccess(ctx context.Context, principal access.Principal, tenantID string, intent TenantAccess) error {
	transaction, err := service.begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	return checkTenantAccess(transaction, principal, tenantID, intent)
}

func checkTenantAccess(transaction *sql.Tx, principal access.Principal, tenantID string, intent TenantAccess) error {
	role, err := access.Role(transaction, principal, tenantID)
	if err != nil {
		return missing()
	}
	selfRemoval := intent.RemoveMemberID != "" && intent.RemoveMemberID == principal.ID
	if role != "administrator" && ((intent.Write && !selfRemoval) || intent.Invitations) {
		return forbidden()
	}
	return nil
}
