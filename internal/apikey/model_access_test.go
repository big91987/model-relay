package apikey_test

import (
	"database/sql"
	"errors"
	"testing"

	"model-relay/internal/apikey"
)

func TestModelAccessFactsRetainCallerTransactionAndSeparatePauseFromGrants(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	allowed, err := apikey.LookupModelAccess(f.db, created.ID, "alpha")
	if err != nil || allowed.TenantID != "tenant" || !allowed.Active || !allowed.Allowed {
		t.Fatal("granted model facts", allowed, err)
	}
	denied, err := apikey.LookupModelAccess(f.db, created.ID, "beta")
	if err != nil || denied.TenantID != "tenant" || !denied.Active || denied.Allowed {
		t.Fatal("ungranted model facts", denied, err)
	}
	transaction, err := f.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.Exec("UPDATE tenants SET active=0 WHERE id='tenant'; DELETE FROM key_model_grants WHERE key_id=?", created.ID); err != nil {
		t.Fatal(err)
	}
	paused, err := apikey.LookupModelAccess(transaction, created.ID, "alpha")
	if err != nil || paused.TenantID != "tenant" || paused.Active || paused.Allowed {
		t.Fatal("transaction-local pause/grant facts", paused, err)
	}
	if allowed, code := apikey.AuthorizeModel(transaction, created.ID, "alpha"); allowed || code != "tenant_paused" {
		t.Fatal("direct authorization lost pause precedence", code)
	}
	if err = transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	restored, err := apikey.LookupModelAccess(f.db, created.ID, "alpha")
	if err != nil || restored != allowed {
		t.Fatal("rolled-back access changes leaked", restored, err)
	}
}

func TestModelAccessLookupPreservesMissingAndStorageErrors(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	if _, err := apikey.LookupModelAccess(f.db, "missing", "alpha"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("missing key fact error", err)
	}
	if _, err := f.db.Exec("DROP TABLE key_model_grants"); err != nil {
		t.Fatal(err)
	}
	if _, err := apikey.LookupModelAccess(f.db, created.ID, "alpha"); err == nil {
		t.Fatal("failed grant query reported valid facts")
	}
	if allowed, code := apikey.AuthorizeModel(f.db, created.ID, "alpha"); allowed || code != "storage_unavailable" {
		t.Fatal("failed grant query lost storage precedence", code)
	}
}
