package apikey_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"model-relay/internal/access"
	"model-relay/internal/apikey"
	"model-relay/internal/configstate"
	"model-relay/internal/fault"
	"model-relay/internal/routing"
)

const schema = `CREATE TABLE models(alias TEXT PRIMARY KEY);
CREATE TABLE app_keys(id TEXT PRIMARY KEY, name TEXT NOT NULL, digest BLOB UNIQUE NOT NULL, hint TEXT NOT NULL, enabled INTEGER NOT NULL, created TEXT NOT NULL);
CREATE TABLE users(id TEXT PRIMARY KEY,account TEXT NOT NULL UNIQUE,salt BLOB NOT NULL,hash BLOB NOT NULL,platform INTEGER NOT NULL CHECK(platform IN(0,1)),active INTEGER NOT NULL CHECK(active IN(0,1)));
CREATE TABLE tenants(id TEXT PRIMARY KEY,name TEXT NOT NULL,active INTEGER NOT NULL CHECK(active IN(0,1)),created TEXT NOT NULL);
CREATE TABLE memberships(tenant_id TEXT NOT NULL REFERENCES tenants(id),user_id TEXT NOT NULL REFERENCES users(id),role TEXT NOT NULL CHECK(role IN('administrator','developer','finance')),PRIMARY KEY(tenant_id,user_id));
CREATE TABLE console_audit(id TEXT PRIMARY KEY,time TEXT NOT NULL,actor TEXT NOT NULL,tenant_id TEXT NOT NULL,action TEXT NOT NULL,object_id TEXT NOT NULL);
CREATE TABLE key_controls(key_id TEXT PRIMARY KEY REFERENCES app_keys(id),tenant_id TEXT NOT NULL REFERENCES tenants(id),application TEXT NOT NULL,expires TEXT,revoked INTEGER NOT NULL DEFAULT 0,revision INTEGER NOT NULL DEFAULT 1,legacy INTEGER NOT NULL DEFAULT 0);
CREATE TABLE key_generations(id TEXT PRIMARY KEY,key_id TEXT NOT NULL REFERENCES app_keys(id),digest BLOB NOT NULL UNIQUE,hint TEXT NOT NULL,created TEXT NOT NULL,expires TEXT,retired INTEGER NOT NULL DEFAULT 0);
CREATE TABLE tenant_model_grants(tenant_id TEXT NOT NULL REFERENCES tenants(id),alias TEXT NOT NULL REFERENCES models(alias),PRIMARY KEY(tenant_id,alias));
CREATE TABLE key_model_grants(key_id TEXT NOT NULL REFERENCES app_keys(id),alias TEXT NOT NULL REFERENCES models(alias),PRIMARY KEY(key_id,alias));
CREATE TABLE key_diagnostic_members(key_id TEXT NOT NULL REFERENCES app_keys(id),tenant_id TEXT NOT NULL,user_id TEXT NOT NULL,PRIMARY KEY(key_id,user_id),FOREIGN KEY(tenant_id,user_id) REFERENCES memberships(tenant_id,user_id) ON DELETE CASCADE);
CREATE TABLE console_operations(actor TEXT NOT NULL,operation TEXT NOT NULL,fingerprint BLOB NOT NULL,object_id TEXT NOT NULL,action TEXT NOT NULL,tenant_id TEXT NOT NULL,PRIMARY KEY(actor,operation));
CREATE TABLE resource_limits(owner_type TEXT NOT NULL CHECK(owner_type IN('tenant','key')),owner_id TEXT NOT NULL,calls INTEGER,rpm INTEGER,tpm INTEGER,concurrency INTEGER,budget INTEGER,hard INTEGER NOT NULL DEFAULT 1,revision INTEGER NOT NULL DEFAULT 1,PRIMARY KEY(owner_type,owner_id),CHECK(calls>=0),CHECK(rpm>=0),CHECK(tpm>=0),CHECK(concurrency>=0),CHECK(budget>=0));
CREATE UNIQUE INDEX current_generation ON key_generations(key_id) WHERE expires IS NULL AND retired=0;`

type fixture struct {
	db      *sql.DB
	service *apikey.Service
	now     time.Time
	admin   access.Principal
	state   *configstate.Coordinator
	router  *routing.Engine
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "keys.db")+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	if _, err = db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO tenants VALUES('tenant','Tenant',1,'2026-01-01'),('other','Other',1,'2026-01-01');
 INSERT INTO users VALUES('developer','developer',x'00',x'00',0,1),('finance','finance',x'00',x'00',0,1);
 INSERT INTO memberships VALUES('tenant','developer','developer'),('tenant','finance','finance');
 INSERT INTO models VALUES('alpha'),('beta');
 INSERT INTO tenant_model_grants VALUES('tenant','alpha'),('tenant','beta');`); err != nil {
		t.Fatal(err)
	}
	f := &fixture{db: db, now: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), admin: access.Principal{ID: "admin", Platform: true}, state: configstate.New(), router: routing.New()}
	f.service = apikey.New(db, func() time.Time { return f.now }, f.state, f.router)
	return f
}
func (f *fixture) input() *apikey.Input {
	return &apikey.Input{Name: "Application key", Application: "Application", Enabled: true, Models: []string{"alpha"}, Diagnostics: []string{"developer"}}
}
func (f *fixture) create(t *testing.T) apikey.Result {
	t.Helper()
	result, err := f.service.Create(context.Background(), f.admin, "tenant", apikey.Operation{Key: "create", Path: "/admin/api/tenants/tenant/keys"}, f.input())
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func expectError(t *testing.T, err error, kind fault.Kind, code string) {
	t.Helper()
	var failure *fault.Error
	if !errors.As(err, &failure) || failure.Kind != kind || failure.Code != code {
		t.Fatalf("failure=%v; want kind %d, code %s", err, kind, code)
	}
}
