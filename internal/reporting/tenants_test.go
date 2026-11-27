package reporting_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/identity"
	"model-relay/internal/installation"
	"model-relay/internal/reporting"
)

type tenantReadFixture struct {
	store     *installation.Store
	people    *identity.Service
	reports   *reporting.Service
	directory string
	ctx       context.Context
	platform  access.Principal
	admin     access.Principal
	dev       access.Principal
}

func tenantReads(t *testing.T) tenantReadFixture {
	t.Helper()
	directory, keyFile := filepath.Join(t.TempDir(), "data"), filepath.Join(t.TempDir(), "key")
	if _, err := installation.InitConsole(directory, keyFile); err != nil {
		t.Fatal(err)
	}
	store, err := installation.Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	store.SetClock(func() time.Time { return time.Date(2030, 1, 2, 12, 0, 0, 0, time.UTC) })
	if err = store.AdvanceClock(); err != nil {
		t.Fatal(err)
	}
	if _, err = store.DB.Exec(`INSERT INTO tenants VALUES('tenant','Tenant',1,'2026-01-01T00:00:00Z'),('other','Other',1,'2026-01-02T00:00:00Z');
INSERT INTO users VALUES('admin','admin@example.com',x'00',x'00',0,1),('dev','dev@example.com',x'00',x'00',0,1),('finance','finance@example.com',x'00',x'00',0,1),('backup','backup@example.com',x'00',x'00',0,1);
INSERT INTO memberships VALUES('tenant','admin','administrator'),('tenant','dev','developer'),('tenant','finance','finance'),('tenant','backup','administrator');`); err != nil {
		t.Fatal(err)
	}
	return tenantReadFixture{store: store, people: identity.New(store.DB, store.Now), reports: reporting.New(store.DB, store.Now), directory: directory, ctx: context.Background(), platform: access.Principal{ID: "platform-admin", Platform: true}, admin: access.Principal{ID: "admin"}, dev: access.Principal{ID: "dev"}}
}

func TestTenantListingScopeAndPrivilegedSummaries(t *testing.T) {
	f := tenantReads(t)
	created, err := f.people.CreateTenant(f.ctx, f.platform, &identity.TenantInput{Name: "Created"})
	if err != nil {
		t.Fatal(err)
	}
	visible, err := f.reports.Tenants(f.ctx, f.dev)
	if err != nil || len(visible) != 1 || visible[0].Role != "developer" || visible[0].Summary != nil {
		t.Fatal(visible, err)
	}
	all, err := f.reports.Tenants(f.ctx, f.platform)
	if err != nil || len(all) != 3 || all[0].Summary == nil {
		t.Fatal(all, err)
	}
	finance, err := f.reports.Tenants(f.ctx, access.Principal{ID: "finance"})
	if err != nil || len(finance) != 1 || finance[0].Summary != nil {
		t.Fatal("finance member acquired privileged summary", finance, err)
	}
	for _, value := range []any{created, visible[0]} {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err = json.Unmarshal(encoded, &fields); err != nil {
			t.Fatal(err)
		}
		if fields["id"] == nil || fields["name"] == nil || fields["summary"] != nil || fields["Tenant"] != nil {
			t.Fatal("tenant response fields changed", fields)
		}
	}
}

func TestTenantListFactsAndSummaryUseOneSnapshot(t *testing.T) {
	f := tenantReads(t)
	writer, err := sql.Open("sqlite3", filepath.Join(f.directory, "relay.db")+"?_foreign_keys=on&_busy_timeout=2000")
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	writer.SetMaxOpenConns(1)
	people := identity.New(writer, f.store.Now)
	var once sync.Once
	reader := reporting.New(f.store.DB, func() time.Time {
		once.Do(func() {
			if _, err := people.ChangeMemberRole(f.ctx, f.platform, "tenant", "admin", &identity.MemberInput{Role: "finance"}); err != nil {
				t.Fatal(err)
			}
		})
		return f.store.Now()
	})
	listed, err := reader.Tenants(f.ctx, f.admin)
	if err != nil || len(listed) != 1 || listed[0].Role != "administrator" || listed[0].Summary == nil {
		t.Fatal("list facts changed during snapshot", listed, err)
	}
	if owners := listed[0].Summary["owners"]; !reflect.DeepEqual(owners, []string{"admin@example.com", "backup@example.com"}) {
		t.Fatal("summary observed a later membership transaction", owners)
	}
	current, err := f.reports.Tenants(f.ctx, f.admin)
	if err != nil || len(current) != 1 || current[0].Role != "finance" || current[0].Summary != nil {
		t.Fatal("next list did not observe committed role change", current, err)
	}
}
