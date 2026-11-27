package identity

import (
	"context"
	"database/sql"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/fault"
)

type Service struct {
	db  *sql.DB
	now func() time.Time
}

func New(db *sql.DB, now func() time.Time) *Service { return &Service{db: db, now: now} }

func (service *Service) begin(ctx context.Context) (*sql.Tx, error) {
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, storage()
	}
	return transaction, nil
}

func storage() error   { return fault.New(fault.Unavailable, "storage_unavailable") }
func missing() error   { return fault.New(fault.NotFound, "not_found") }
func forbidden() error { return fault.New(fault.Forbidden, "forbidden") }

func administrator(transaction *sql.Tx, principal access.Principal, tenantID string) error {
	role, err := access.Role(transaction, principal, tenantID)
	if err != nil {
		return missing()
	}
	if role != "administrator" {
		return forbidden()
	}
	return nil
}

type TenantInput struct {
	Name string `json:"name"`
}
type StatusInput struct {
	Active bool `json:"active"`
}
type InvitationInput struct {
	Account string `json:"account"`
	Role    string `json:"role"`
	Hours   int    `json:"hours"`
}
type MemberInput struct {
	Role string `json:"role"`
}
type AcceptInput struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// RevokeInput is the empty command body accepted by invitation revocation.
// A nil pointer represents a request that could not be decoded by the adapter.
type RevokeInput struct{}

type Member struct {
	ID      string `json:"id"`
	Account string `json:"account"`
	Role    string `json:"role"`
	Active  bool   `json:"active"`
}

type IssuedInvitation struct {
	Invitation      Invitation `json:"invitation"`
	AcceptanceToken string     `json:"acceptance_token"`
	Delivery        string     `json:"delivery"`
}

type Acceptance struct {
	Accepted       bool   `json:"accepted"`
	AlreadyApplied bool   `json:"already_applied,omitempty"`
	Account        string `json:"account,omitempty"`
	TenantID       string `json:"tenant_id,omitempty"`
	EmailVerified  *bool  `json:"email_verified,omitempty"`
}
