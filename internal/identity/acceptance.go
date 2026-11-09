package identity

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"model-relay/internal/fault"
	"model-relay/internal/operations"
	"model-relay/internal/secure"
)

// AcceptInvitation calls verifyExisting only if the invited account exists.
// The proof callback inspects session identity and CSRF at that point; it must
// not use the service's database, whose transaction is already held. Password
// concurrency and attempt windows belong to the adapter's sessions owner.
func (service *Service) AcceptInvitation(ctx context.Context, input *AcceptInput, verifyExisting func(expectedUserID string) bool) (Acceptance, error) {
	if input == nil || len(input.Token) != 43 || len(input.Password) > 1024 {
		return Acceptance{}, fault.New(fault.Invalid, "invalid_invitation")
	}
	transaction, err := service.begin(ctx)
	if err != nil {
		return Acceptance{}, err
	}
	defer transaction.Rollback()
	var invitation Invitation
	var acceptedBy sql.NullString
	var active bool
	err = transaction.QueryRow("SELECT i.id,i.tenant_id,i.account,i.role,i.expires,i.state,i.accepted_by,t.active FROM invitations i JOIN tenants t ON t.id=i.tenant_id WHERE i.digest=?", secure.Digest(input.Token)).Scan(&invitation.ID, &invitation.Tenant, &invitation.Account, &invitation.Role, &invitation.Expires, &invitation.State, &acceptedBy, &active)
	expiry, parseErr := time.Parse(time.RFC3339Nano, invitation.Expires)
	if err == nil && invitation.State == "accepted" {
		return Acceptance{Accepted: true, AlreadyApplied: true}, nil
	}
	if err != nil || parseErr != nil || !service.now().Before(expiry) || invitation.State != "pending" || !active {
		return Acceptance{}, fault.New(fault.Invalid, "invitation_unavailable")
	}
	var userID string
	err = transaction.QueryRow("SELECT id FROM users WHERE account=?", invitation.Account).Scan(&userID)
	if err == nil {
		if verifyExisting == nil || !verifyExisting(userID) {
			return Acceptance{}, fault.New(fault.Unauthenticated, "existing_account_login_required")
		}
		var enabled bool
		if transaction.QueryRow("SELECT active FROM users WHERE id=?", userID).Scan(&enabled) != nil || !enabled {
			return Acceptance{}, fault.New(fault.Unauthenticated, "unauthorized")
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		if !validPassword(input.Password) {
			return Acceptance{}, fault.New(fault.Invalid, "password_requires_12_characters")
		}
		userID = secure.Random(16)
		salt, hash := passwordMaterial(input.Password)
		if _, err = transaction.Exec("INSERT INTO users VALUES(?,?,?,?,0,1)", userID, invitation.Account, salt, hash); err != nil {
			return Acceptance{}, storage()
		}
	} else {
		return Acceptance{}, storage()
	}
	var membershipCount int
	if transaction.QueryRow("SELECT count(*) FROM memberships WHERE tenant_id=? AND user_id=?", invitation.Tenant, userID).Scan(&membershipCount) != nil {
		return Acceptance{}, storage()
	}
	if membershipCount != 0 {
		return Acceptance{}, fault.New(fault.Conflict, "already_member")
	}
	if _, err = transaction.Exec("INSERT INTO memberships VALUES(?,?,?); UPDATE invitations SET state='accepted',accepted_by=? WHERE id=? AND state='pending'", invitation.Tenant, userID, invitation.Role, userID, invitation.ID); err != nil || operations.Audit(transaction, userID, invitation.Tenant, "invitation.accept", invitation.ID) != nil || transaction.Commit() != nil {
		return Acceptance{}, storage()
	}
	verified := false
	return Acceptance{Accepted: true, Account: invitation.Account, TenantID: invitation.Tenant, EmailVerified: &verified}, nil
}
