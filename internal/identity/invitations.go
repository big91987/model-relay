package identity

import (
	"context"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
	"model-relay/internal/secure"
)

func (service *Service) Invitations(ctx context.Context, principal access.Principal, tenantID string) ([]Invitation, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	if err = administrator(transaction, principal, tenantID); err != nil {
		return nil, err
	}
	rows, err := transaction.Query("SELECT id,tenant_id,account,role,expires,state FROM invitations WHERE tenant_id=? ORDER BY expires,id", tenantID)
	if err != nil {
		return nil, storage()
	}
	defer rows.Close()
	values := []Invitation{}
	for rows.Next() {
		var value Invitation
		if rows.Scan(&value.ID, &value.Tenant, &value.Account, &value.Role, &value.Expires, &value.State) != nil {
			return nil, storage()
		}
		// Listing projects display state using observed wall UTC; acceptance uses persisted effective UTC.
		if value.State == "pending" && value.Expires <= time.Now().UTC().Format(time.RFC3339Nano) {
			value.State = "expired"
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return nil, storage()
	}
	return values, nil
}

func (service *Service) Invite(ctx context.Context, principal access.Principal, tenantID string, command *InvitationInput) (*IssuedInvitation, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	if err = administrator(transaction, principal, tenantID); err != nil {
		return nil, err
	}
	if command == nil {
		return nil, fault.New(fault.Invalid, "invalid_request")
	}
	input := *command
	account, valid := invitationAccount(input.Account)
	if input.Hours == 0 {
		input.Hours = 24
	}
	if !valid || !access.ValidRole(input.Role) || input.Hours < 1 || input.Hours > 168 {
		return nil, fault.New(fault.Invalid, "invalid_invitation")
	}
	secret := secure.Random(32)
	invitation := Invitation{ID: secure.Random(16), Tenant: tenantID, Account: account, Role: input.Role, Expires: service.now().Add(time.Duration(input.Hours) * time.Hour).Format(time.RFC3339Nano), State: "pending"}
	if _, err := transaction.Exec("INSERT INTO invitations VALUES(?,?,?,?,?,?,'pending',NULL)", invitation.ID, tenantID, account, input.Role, secure.Digest(secret), invitation.Expires); err != nil || operations.Audit(transaction, principal.ID, tenantID, "invitation.create", invitation.ID) != nil || transaction.Commit() != nil {
		return nil, storage()
	}
	return &IssuedInvitation{Invitation: invitation, AcceptanceToken: secret, Delivery: "manual_not_emailed"}, nil
}

func (service *Service) RevokeInvitation(ctx context.Context, principal access.Principal, tenantID, id string, input *RevokeInput) (map[string]bool, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	if err = administrator(transaction, principal, tenantID); err != nil {
		return nil, err
	}
	if input == nil {
		return nil, fault.New(fault.Invalid, "invalid_request")
	}
	result, err := transaction.Exec("UPDATE invitations SET state='revoked' WHERE id=? AND tenant_id=? AND state='pending'", id, tenantID)
	if err != nil {
		return nil, storage()
	}
	changed, err := result.RowsAffected()
	if err != nil || changed != 1 {
		return nil, fault.New(fault.Conflict, "invitation_not_pending")
	}
	if operations.Audit(transaction, principal.ID, tenantID, "invitation.revoke", id) != nil || transaction.Commit() != nil {
		return nil, storage()
	}
	return map[string]bool{"ok": true}, nil
}
