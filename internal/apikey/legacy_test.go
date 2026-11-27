package apikey_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"model-relay/internal/access"
	"model-relay/internal/apikey"
	"model-relay/internal/fault"
)

func TestLegacyKeysPreserveOneTimeSecretAndCommittedRevision(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if _, err := f.db.Exec("CREATE TABLE meta(config_revision INTEGER);INSERT INTO meta VALUES(1)"); err != nil {
		t.Fatal(err)
	}
	result, err := f.service.CreateLegacy(ctx, "Legacy application")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Secret) != 46 || result.Metadata.ID == "" || result.Metadata.Hint != result.Secret[:11] {
		t.Fatalf("invalid legacy issuance: %+v", result.Metadata)
	}
	if id, ok := f.service.IdentifyLegacy(result.Secret); !ok || id != result.Metadata.ID {
		t.Fatal("issued legacy credential rejected")
	}
	if err = f.service.SetLegacyEnabled(ctx, result.Metadata.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.service.IdentifyLegacy(result.Secret); ok {
		t.Fatal("disabled legacy credential authorized")
	}
	if id, ok := f.service.Observe(result.Secret, 2); id != result.Metadata.ID || ok {
		t.Fatal("disabled legacy identity disappeared")
	}
	values, err := f.service.ListLegacy(ctx)
	if err != nil || len(values) != 1 || values[0].Enabled {
		t.Fatalf("global legacy list changed: %v %v", values, err)
	}
	encoded, _ := json.Marshal(values)
	if string(encoded) == "" || strings.Contains(string(encoded), result.Secret) {
		t.Fatal("metadata list exposed secret")
	}
	var revision int
	if err = f.db.QueryRow("SELECT config_revision FROM meta").Scan(&revision); err != nil || revision != 3 {
		t.Fatalf("legacy config revision drift: %d %v", revision, err)
	}
	if _, err = f.db.Exec(`CREATE TRIGGER reject_legacy_revision BEFORE UPDATE ON meta BEGIN SELECT RAISE(ABORT,'controlled commit failure'); END`); err != nil {
		t.Fatal(err)
	}
	err = f.service.SetLegacyEnabled(ctx, result.Metadata.ID, true)
	expectError(t, err, fault.Unavailable, "storage_unavailable")
	if _, ok := f.service.IdentifyLegacy(result.Secret); ok {
		t.Fatal("failed transaction enabled credential")
	}
}
func TestNullableInputsPreserveAuthorizationAndOperationPrecedence(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	created := f.create(t)
	_, err := f.service.Create(ctx, access.Principal{ID: "developer"}, "tenant", apikey.Operation{}, nil)
	expectError(t, err, fault.Forbidden, "forbidden")
	_, err = f.service.Create(ctx, f.admin, "tenant", apikey.Operation{}, nil)
	expectError(t, err, fault.Invalid, "invalid_key")
	_, err = f.service.Update(ctx, f.admin, "other", created.ID, apikey.Operation{}, nil)
	expectError(t, err, fault.NotFound, "not_found")
	_, err = f.service.Update(ctx, f.admin, "tenant", created.ID, apikey.Operation{}, nil)
	expectError(t, err, fault.Invalid, "invalid_key")
	for _, transition := range []func(context.Context, access.Principal, string, string, apikey.Operation, *apikey.Transition) (apikey.Result, error){f.service.Rotate, f.service.Revoke, f.service.RetirePrevious} {
		_, err = transition(ctx, access.Principal{ID: "developer"}, "tenant", created.ID, apikey.Operation{}, nil)
		expectError(t, err, fault.Forbidden, "forbidden")
		_, err = transition(ctx, f.admin, "other", created.ID, apikey.Operation{}, nil)
		expectError(t, err, fault.NotFound, "not_found")
		_, err = transition(ctx, f.admin, "tenant", created.ID, apikey.Operation{}, nil)
		expectError(t, err, fault.Invalid, "invalid_request")
	}
}

func TestGenerationObservationKeepsRejectedIdentityAndGlobalMetadata(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	values, err := f.service.ListLegacy(context.Background())
	if err != nil || len(values) != 1 || values[0].ID != created.ID {
		t.Fatalf("global schema3 metadata list lost controlled key: %v %v", values, err)
	}
	if _, err = f.db.Exec("UPDATE key_generations SET retired=1 WHERE key_id=?", created.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.service.Identify(created.Secret); ok {
		t.Fatal("retired generation authorized")
	}
	if id, ok := f.service.Observe(created.Secret, 3); id != created.ID || ok {
		t.Fatal("retired generation evidence lost identity")
	}
	if id, ok := f.service.Observe("mr_unknown", 3); id != "" || ok {
		t.Fatal("unknown key acquired identity")
	}
}

func TestGenerationModelListUsesCallableRoutesAndBothGrants(t *testing.T) {
	f := newFixture(t)
	created := f.create(t)
	if _, err := f.db.Exec(`ALTER TABLE models ADD COLUMN enabled INTEGER NOT NULL DEFAULT 1;
 CREATE TABLE upstream(id INTEGER PRIMARY KEY,enabled INTEGER,token BLOB);
 CREATE TABLE route_candidates(alias TEXT,upstream_id INTEGER,enabled INTEGER);
 INSERT INTO upstream VALUES(1,1,x'00');INSERT INTO route_candidates VALUES('alpha',1,1),('beta',1,1);`); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	models, err := f.service.ListModels(ctx, created.ID, 3)
	if err != nil || len(models) != 1 || models[0]["id"] != "alpha" || models[0]["object"] != "model" || models[0]["owned_by"] != "model-relay" {
		t.Fatalf("schema3 list escaped key grant: %v %v", models, err)
	}
	models, err = f.service.ListModels(ctx, created.ID, 2)
	if err != nil || len(models) != 2 {
		t.Fatalf("legacy callable model list changed: %v %v", models, err)
	}
	if _, err = f.db.Exec("DELETE FROM tenant_model_grants WHERE tenant_id='tenant' AND alias='alpha'"); err != nil {
		t.Fatal(err)
	}
	models, err = f.service.ListModels(ctx, created.ID, 3)
	if err != nil || len(models) != 0 {
		t.Fatalf("removed tenant grant remained callable: %v %v", models, err)
	}
	if _, err = f.db.Exec("UPDATE tenants SET active=0 WHERE id='tenant'"); err != nil {
		t.Fatal(err)
	}
	_, err = f.service.ListModels(ctx, created.ID, 3)
	expectError(t, err, fault.Forbidden, "tenant_paused")
	if _, err = f.db.Exec("UPDATE upstream SET token=x''"); err != nil {
		t.Fatal(err)
	}
	models, err = f.service.ListModels(ctx, created.ID, 2)
	if err != nil || len(models) != 0 {
		t.Fatalf("credentialless route remained callable: %v %v", models, err)
	}
}
