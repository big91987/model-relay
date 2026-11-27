package identity

import (
	"context"
	"database/sql"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
)

func (service *Service) Members(ctx context.Context, principal access.Principal, tenantID string) ([]Member, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	if err = administrator(transaction, principal, tenantID); err != nil {
		return nil, err
	}
	rows, err := transaction.Query("SELECT u.id,u.account,m.role,u.active FROM memberships m JOIN users u ON u.id=m.user_id WHERE m.tenant_id=? ORDER BY u.account,u.id", tenantID)
	if err != nil {
		return nil, storage()
	}
	defer rows.Close()
	values := []Member{}
	for rows.Next() {
		var value Member
		if rows.Scan(&value.ID, &value.Account, &value.Role, &value.Active) != nil {
			return nil, storage()
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return nil, storage()
	}
	return values, nil
}

func (service *Service) ChangeMemberRole(ctx context.Context, principal access.Principal, tenantID, userID string, input *MemberInput) (map[string]bool, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	previous, err := memberForChange(transaction, principal, tenantID, userID, false)
	if err != nil {
		return nil, err
	}
	if input == nil || !access.ValidRole(input.Role) {
		return nil, fault.New(fault.Invalid, "invalid_role")
	}
	if err = protectLastAdministrator(transaction, tenantID, previous, input.Role); err != nil {
		return nil, err
	}
	if _, err := transaction.Exec("UPDATE memberships SET role=? WHERE tenant_id=? AND user_id=?", input.Role, tenantID, userID); err != nil {
		return nil, storage()
	}
	if operations.Audit(transaction, principal.ID, tenantID, "member.patch", userID) != nil || transaction.Commit() != nil {
		return nil, storage()
	}
	return map[string]bool{"ok": true}, nil
}

func (service *Service) RemoveMember(ctx context.Context, principal access.Principal, tenantID, userID string) (map[string]bool, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	previous, err := memberForChange(transaction, principal, tenantID, userID, true)
	if err != nil {
		return nil, err
	}
	if err = protectLastAdministrator(transaction, tenantID, previous, ""); err != nil {
		return nil, err
	}
	if _, err := transaction.Exec("DELETE FROM memberships WHERE tenant_id=? AND user_id=?", tenantID, userID); err != nil {
		return nil, storage()
	}
	if operations.Audit(transaction, principal.ID, tenantID, "member.delete", userID) != nil || transaction.Commit() != nil {
		return nil, storage()
	}
	return map[string]bool{"ok": true}, nil
}

func memberForChange(transaction *sql.Tx, principal access.Principal, tenantID, userID string, removing bool) (string, error) {
	intent := TenantAccess{Write: true}
	if removing {
		intent.RemoveMemberID = userID
	}
	if err := checkTenantAccess(transaction, principal, tenantID, intent); err != nil {
		return "", err
	}
	var previous string
	if transaction.QueryRow("SELECT role FROM memberships WHERE tenant_id=? AND user_id=?", tenantID, userID).Scan(&previous) != nil {
		return "", missing()
	}
	return previous, nil
}

func protectLastAdministrator(transaction *sql.Tx, tenantID, previous, next string) error {
	if previous != "administrator" || next == "administrator" {
		return nil
	}
	var count int
	if transaction.QueryRow("SELECT count(*) FROM memberships m JOIN users u ON u.id=m.user_id WHERE tenant_id=? AND role='administrator' AND u.active=1", tenantID).Scan(&count) != nil {
		return storage()
	}
	if count <= 1 {
		return fault.New(fault.Conflict, "last_administrator")
	}
	return nil
}
