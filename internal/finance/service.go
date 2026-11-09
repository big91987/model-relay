// Package finance owns resource limits, published prices, usage settlement and
// bill operations. Services own their transaction; Settle and the admission
// reads participate in the caller's existing request or recovery transaction.
package finance

import (
	"context"
	"database/sql"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
)

type Service struct {
	db  *sql.DB
	now func() time.Time
}

func New(db *sql.DB, clock func() time.Time) *Service {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &Service{db: db, now: clock}
}

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

func beginOperation(transaction *sql.Tx, principal access.Principal, key, path, action string, input any) (string, string, []byte, error) {
	operation, prior, fingerprint, err := operations.Begin(transaction, principal.ID, key, path, action, input)
	if err != nil {
		return "", "", nil, fault.New(fault.Conflict, err.Error())
	}
	return operation, prior, fingerprint, nil
}

func completeOperation(transaction *sql.Tx, principal access.Principal, operation string, fingerprint []byte, action, objectID, tenantID string) error {
	if operations.Complete(transaction, principal.ID, operation, fingerprint, action, objectID, tenantID) != nil || transaction.Commit() != nil {
		return storage()
	}
	return nil
}

func replay(id string) operations.Result {
	return operations.Result{Value: struct {
		ID             string `json:"id"`
		AlreadyApplied bool   `json:"already_applied"`
	}{id, true}}
}
func ok() operations.Result {
	return operations.Result{Value: struct {
		OK bool `json:"ok"`
	}{true}}
}

func modelExists(transaction *sql.Tx, principal access.Principal, alias string) error {
	if !principal.Platform {
		return forbidden()
	}
	var count int
	if transaction.QueryRow("SELECT count(*) FROM models WHERE alias=?", alias).Scan(&count) != nil || count != 1 {
		return missing()
	}
	return nil
}
