package apikey_test

import (
	"context"
	"testing"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/apikey"
	"model-relay/internal/fault"
	"model-relay/internal/secure"
)

func TestIssueReplayGrantsAndOneTimeSecret(t *testing.T) {
	f := newFixture(t)
	result := f.create(t)
	if result.ID == "" || len(result.Secret) != 46 || result.AlreadyApplied {
		t.Fatal("issued identity/secret invalid", result)
	}
	key, err := f.service.Read(context.Background(), f.admin, "tenant", result.ID)
	if err != nil || key.ID != result.ID || key.Hint != result.Secret[:11] || !key.Enabled || key.Revision != 1 || len(key.Models) != 1 || key.Models[0] != "alpha" || len(key.Diagnostics) != 1 || key.Diagnostics[0] != "developer" || len(key.Generations) != 1 {
		t.Fatal("persisted key metadata/grants", key, err)
	}
	var saved []byte
	if err = f.db.QueryRow("SELECT digest FROM app_keys WHERE id=?", result.ID).Scan(&saved); err != nil || string(saved) != string(secure.Digest(result.Secret)) {
		t.Fatal("secret digest not persisted", err)
	}
	replayed, err := f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "create", Path: "/admin/api/tenants/tenant/keys"}, f.input())
	if err != nil || !replayed.AlreadyApplied || replayed.ID != result.ID || replayed.Secret != "" {
		t.Fatal("replay exposed or reissued secret", replayed, err)
	}
	var audits int
	if f.db.QueryRow("SELECT count(*) FROM console_audit").Scan(&audits) != nil || audits != 1 {
		t.Fatal("create/replay audit duplication", audits)
	}
	changed := f.input()
	changed.Application = "Changed"
	_, err = f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "create", Path: "/admin/api/tenants/tenant/keys"}, changed)
	expectError(t, err, fault.Conflict, "idempotency_conflict")
}

func TestDiagnosticReadScopeAndTenantIsolation(t *testing.T) {
	f := newFixture(t)
	first := f.create(t)
	input := f.input()
	input.Diagnostics = nil
	second, err := f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "second", Path: "/keys"}, input)
	if err != nil {
		t.Fatal(err)
	}
	developer := access.Principal{ID: "developer"}
	keys, err := f.service.List(context.Background(), developer, "tenant")
	if err != nil || len(keys) != 1 || keys[0].ID != first.ID {
		t.Fatal("developer list escaped diagnostic assignment", keys, err)
	}
	_, err = f.service.Read(context.Background(), developer, "tenant", second.ID)
	expectError(t, err, fault.NotFound, "not_found")
	_, err = f.service.Read(context.Background(), f.admin, "other", first.ID)
	expectError(t, err, fault.NotFound, "not_found")
	_, err = f.service.Update(context.Background(), developer, "tenant", "missing", apikey.Operation{}, input)
	expectError(t, err, fault.Forbidden, "forbidden")
	keys, err = f.service.List(context.Background(), access.Principal{ID: "finance"}, "tenant")
	if err != nil || len(keys) != 2 {
		t.Fatal("finance read scope changed", keys, err)
	}
}

func TestInvalidInputAndOperationPrecedence(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	input := f.input()
	input.Name = ""
	input.Revision = 99
	_, err := f.service.Update(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{}, input)
	expectError(t, err, fault.Conflict, "idempotency_key_required")
	_, err = f.service.Update(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "invalid", Path: "/keys/" + created.ID}, input)
	expectError(t, err, fault.Invalid, "invalid_key")
	input = f.input()
	input.Models = []string{"alpha", "alpha"}
	_, err = f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "dupe", Path: "/keys"}, input)
	expectError(t, err, fault.Invalid, "invalid_key")
	input = f.input()
	input.Diagnostics = []string{"finance"}
	_, err = f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "wrong-role", Path: "/keys"}, input)
	expectError(t, err, fault.Invalid, "invalid_key")
	_, err = f.service.Rotate(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{}, &apikey.Transition{Hours: 25, Revision: 99})
	expectError(t, err, fault.Invalid, "invalid_request")
}

func TestPausedTenantOnlyAllowsTightening(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	if _, err := f.db.Exec("UPDATE tenants SET active=0 WHERE id='tenant'"); err != nil {
		t.Fatal(err)
	}
	_, err := f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "paused-create", Path: "/keys"}, f.input())
	expectError(t, err, fault.Conflict, "tenant_paused")
	input := f.input()
	input.Revision = 1
	input.Models = []string{"alpha", "beta"}
	_, err = f.service.Update(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "expand", Path: "/keys/key"}, input)
	expectError(t, err, fault.Conflict, "tenant_paused")
	input = f.input()
	input.Revision = 1
	input.Enabled = false
	input.Models = nil
	_, err = f.service.Update(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "tighten", Path: "/keys/key"}, input)
	if err != nil {
		t.Fatal("paused tightening rejected", err)
	}
	input.Enabled = true
	input.Revision = 2
	_, err = f.service.Update(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "enable", Path: "/keys/key"}, input)
	expectError(t, err, fault.Conflict, "tenant_paused")
	_, err = f.service.Rotate(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "rotate", Path: "/keys/key/rotations"}, &apikey.Transition{Revision: 2})
	expectError(t, err, fault.Conflict, "tenant_paused")
	_, err = f.service.Revoke(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "revoke", Path: "/keys/key/revoke"}, &apikey.Transition{Revision: 2})
	if err != nil {
		t.Fatal("paused revocation rejected", err)
	}
}

func TestRotationTransitionExpiryAndRevocation(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	operation := apikey.Operation{Key: "rotate", Path: "/keys/key/rotations"}
	transition := &apikey.Transition{Hours: 1, Revision: 1}
	rotated, err := f.service.Rotate(context.Background(), f.admin, "tenant", created.ID, operation, transition)
	if err != nil || len(rotated.Secret) != 46 || rotated.Secret == created.Secret {
		t.Fatal("rotation failed", rotated, err)
	}
	for _, secret := range []string{created.Secret, rotated.Secret} {
		if id, ok := apikey.GenerationIdentity(f.db, secret, f.now); !ok || id != created.ID {
			t.Fatal("active transition credential rejected", id, ok)
		}
	}
	replay, err := f.service.Rotate(context.Background(), f.admin, "tenant", created.ID, operation, transition)
	if err != nil || !replay.AlreadyApplied || replay.Secret != "" {
		t.Fatal("rotation replay secret recovered", replay, err)
	}
	_, err = f.service.Rotate(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "second", Path: operation.Path}, &apikey.Transition{Revision: 2, Hours: 1})
	expectError(t, err, fault.Conflict, "transition_already_active")
	f.now = f.now.Add(time.Hour)
	if id, ok := apikey.GenerationIdentity(f.db, created.Secret, f.now); ok || id != created.ID {
		t.Fatal("transition expiry boundary changed", id, ok)
	}
	latest, err := f.service.Rotate(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "later", Path: operation.Path}, &apikey.Transition{Revision: 2, Hours: 0})
	if err != nil {
		t.Fatal("expired transition prevented new rotation", err)
	}
	if _, ok := apikey.GenerationIdentity(f.db, rotated.Secret, f.now); ok {
		t.Fatal("zero-hour rotation retained old generation")
	}
	_, err = f.service.Revoke(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "revoke", Path: "/keys/key/revoke"}, &apikey.Transition{Revision: 3})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := apikey.GenerationIdentity(f.db, latest.Secret, f.now); ok {
		t.Fatal("revoked generation authorized")
	}
}

func TestRetirePreviousKeepsCurrentGeneration(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	rotated, err := f.service.Rotate(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "rotate", Path: "/keys/key/rotations"}, &apikey.Transition{Revision: 1, Hours: 24})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.RetirePrevious(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "retire", Path: "/keys/key/retire-previous"}, &apikey.Transition{Revision: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := apikey.GenerationIdentity(f.db, created.Secret, f.now); ok {
		t.Fatal("previous credential survived retirement")
	}
	if _, ok := apikey.GenerationIdentity(f.db, rotated.Secret, f.now); !ok {
		t.Fatal("current credential retired")
	}
}

func TestAuditFailureRollsBackIssueAndRetry(t *testing.T) {
	f := newFixture(t)
	if _, err := f.db.Exec(`CREATE TRIGGER fail_audit BEFORE INSERT ON console_audit BEGIN SELECT RAISE(FAIL,'controlled'); END;`); err != nil {
		t.Fatal(err)
	}
	_, err := f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "create", Path: "/admin/api/tenants/tenant/keys"}, f.input())
	expectError(t, err, fault.Unavailable, "storage_unavailable")
	for _, table := range []string{"app_keys", "key_controls", "key_generations", "key_model_grants", "key_diagnostic_members", "console_operations", "resource_limits"} {
		var count int
		if f.db.QueryRow("SELECT count(*) FROM "+table).Scan(&count) != nil || count != 0 {
			t.Fatal("partial issue escaped rollback", table, count)
		}
	}
	if _, err = f.db.Exec("DROP TRIGGER fail_audit"); err != nil {
		t.Fatal(err)
	}
	if created := f.create(t); created.AlreadyApplied || created.Secret == "" {
		t.Fatal("failed transaction consumed idempotency key")
	}
}

func TestModelAuthorizationAndAtomicScope(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	if allowed, code := apikey.AuthorizeModel(f.db, created.ID, "alpha"); !allowed || code != "" {
		t.Fatal("model grant rejected", code)
	}
	if allowed, code := apikey.AuthorizeModel(f.db, created.ID, "beta"); allowed || code != "model_forbidden" {
		t.Fatal("missing model grant accepted", code)
	}
	transaction, err := f.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.Exec("UPDATE tenants SET active=0 WHERE id='tenant'"); err != nil {
		t.Fatal(err)
	}
	if allowed, code := apikey.AuthorizeModel(transaction, created.ID, "beta"); allowed || code != "tenant_paused" {
		t.Fatal("authorization lost pause precedence", code)
	}
	key, role, err := apikey.Scope(transaction, access.Principal{ID: "developer"}, created.ID, false)
	if err != nil || key.ID != created.ID || role != "developer" {
		t.Fatal("transactional diagnostic scope changed", key, role, err)
	}
}

func TestKeyExpiryAndPausedExpiryExpansion(t *testing.T) {
	f := newFixture(t)
	expires := f.now.Add(time.Hour).Format(time.RFC3339Nano)
	input := f.input()
	input.Expires = &expires
	created, err := f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "expiring", Path: "/keys"}, input)
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := f.service.Identify(created.Secret); !ok || id != created.ID {
		t.Fatal("unexpired key rejected", id, ok)
	}
	if _, err = f.db.Exec("UPDATE tenants SET active=0 WHERE id='tenant'"); err != nil {
		t.Fatal(err)
	}
	input.Revision = 1
	input.Expires = nil
	_, err = f.service.Update(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "unexpire", Path: "/keys/key"}, input)
	expectError(t, err, fault.Conflict, "tenant_paused")
	input.Expires = &expires
	f.now = f.now.Add(time.Hour)
	if id, ok := f.service.Identify(created.Secret); ok || id != created.ID {
		t.Fatal("key expiry equality boundary changed", id, ok)
	}
	_, err = f.service.Update(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "expired-enable", Path: "/keys/key"}, input)
	expectError(t, err, fault.Invalid, "invalid_key")
	input.Enabled = false
	_, err = f.service.Update(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "expired-disable", Path: "/keys/key"}, input)
	if err != nil {
		t.Fatal("disabling expired key rejected", err)
	}
}

func TestRotationAuditFailureRestoresOldCredential(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	if _, err := f.db.Exec(`CREATE TRIGGER fail_rotation_audit BEFORE INSERT ON console_audit WHEN NEW.action='key.rotations' BEGIN SELECT RAISE(FAIL,'controlled'); END;`); err != nil {
		t.Fatal(err)
	}
	_, err := f.service.Rotate(context.Background(), f.admin, "tenant", created.ID, apikey.Operation{Key: "rotate", Path: "/keys/key/rotations"}, &apikey.Transition{Hours: 0, Revision: 1})
	expectError(t, err, fault.Unavailable, "storage_unavailable")
	if _, ok := f.service.Identify(created.Secret); !ok {
		t.Fatal("failed rotation retired current credential")
	}
	key, err := f.service.Read(context.Background(), f.admin, "tenant", created.ID)
	if err != nil || key.Revision != 1 || len(key.Generations) != 1 {
		t.Fatal("failed rotation changed generation/revision", key, err)
	}
}
