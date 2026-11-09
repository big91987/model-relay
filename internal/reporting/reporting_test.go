package reporting_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/reporting"
)

type reportFixture struct {
	db                   *sql.DB
	service              *reporting.Service
	ctx                  context.Context
	platform, admin, dev access.Principal
}

func reports(t *testing.T) reportFixture {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "reports.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	f := reportFixture{db: db, service: reporting.New(db, func() time.Time { return time.Date(2030, 1, 2, 12, 0, 0, 0, time.UTC) }), ctx: context.Background(), platform: access.Principal{ID: "platform", Platform: true}, admin: access.Principal{ID: "admin"}, dev: access.Principal{ID: "dev"}}
	f.exec(t, reportSchema)
	f.exec(t, `INSERT INTO users VALUES('admin','admin@example.com',1),('dev','dev@example.com',1);
INSERT INTO tenants VALUES('tenant','Tenant'),('other','Other'); INSERT INTO memberships VALUES('tenant','admin','administrator'),('tenant','dev','developer');
INSERT INTO key_controls VALUES('assigned','tenant'),('unassigned','tenant'),('other-key','other'); INSERT INTO key_diagnostic_members VALUES('assigned','tenant','dev');
INSERT INTO requests VALUES('one','2030-01-02T10:00:00Z','assigned','model','remote','success','',0,200,10,NULL,NULL,NULL),('two','2030-01-02T11:00:00Z','unassigned','model','remote','upstream_failed','upstream_http_error',0,500,30,1,2,3),('three','2030-01-02T12:00:00Z','other-key','model','remote','success','',0,200,50,1,1,2);
INSERT INTO admissions VALUES('one','tenant','assigned','2030-01',1,20,100,1),('two','tenant','unassigned','2030-01',0,0,0,1),('three','other','other-key','2030-01',0,0,0,1);
INSERT INTO charge_entries VALUES('one',1,0,'2030-01',3,4,11,18,'',''),('one',1,1,'2030-01',1,2,3,5,'trace','receipt'),('two',1,0,'2030-01',1,2,2,4,'',''),('three',1,0,'2030-01',1,1,9,7,'','');
INSERT INTO models VALUES('model',1,7),('ungranted',1,1); INSERT INTO tenant_model_grants VALUES('tenant','model');
INSERT INTO upstream VALUES(1,'source',1,9); INSERT INTO route_candidates VALUES('model',1,1,10,3,11); INSERT INTO service_profiles VALUES(1,'provider');
INSERT INTO prices VALUES('price','model',1,'2030-01-01T00:00:00Z',0,0,100,200,9,11);
INSERT INTO console_audit VALUES('audit-tenant','2030-01-02T10:00:00Z','admin','tenant','member.patch','dev'),('audit-other','2030-01-02T11:00:00Z','platform','other','tenant.status','other');
INSERT INTO console_operations VALUES('admin','own','one','usage.reconcile','tenant'),('admin','removed','three','usage.reconcile','other'),('dev','global','model','price.publish','');`)
	return f
}
func (f reportFixture) exec(t *testing.T, statement string, args ...any) {
	t.Helper()
	if _, err := f.db.Exec(statement, args...); err != nil {
		t.Fatal(err)
	}
}
func expect(t *testing.T, err error, kind fault.Kind, code string) {
	t.Helper()
	var result *fault.Error
	if !errors.As(err, &result) || result.Kind != kind || result.Code != code {
		t.Fatalf("want %v/%s got %v", kind, code, err)
	}
}

func TestRequestListsKeepAssignedKeyIsolationAndConfirmedLedgerTotals(t *testing.T) {
	f := reports(t)
	page, err := f.service.Requests(f.ctx, f.dev, reporting.Filters{Tenant: "tenant"}, reporting.PageInput{})
	if err != nil || page.Total != 1 || len(page.Items) != 1 || page.Items[0]["id"] != "one" || page.Items[0]["confirmed_input_tokens"] != "4" || page.Items[0]["confirmed_output_tokens"] != "6" || page.Items[0]["prompt_tokens"] != nil || page.Items[0]["fee_micro_usd"] != "23" {
		t.Fatal(page, err)
	}
	page, err = f.service.Requests(f.ctx, f.admin, reporting.Filters{Tenant: "tenant"}, reporting.PageInput{Page: "2", Size: "1"})
	if err != nil || page.Total != 2 || len(page.Items) != 1 || page.Items[0]["id"] != "one" {
		t.Fatal("page", page, err)
	}
	page, err = f.service.Requests(f.ctx, f.platform, reporting.Filters{From: "2030-01-02T11:00:00Z", To: "2030-01-02T11:00:00Z"}, reporting.PageInput{})
	if err != nil || page.Total != 1 || page.Items[0]["id"] != "two" {
		t.Fatal("inclusive time filter", page, err)
	}
}

func TestReportScopeAndTimeErrorsPrecedePaginationErrors(t *testing.T) {
	f := reports(t)
	_, err := f.service.Requests(f.ctx, f.dev, reporting.Filters{From: "invalid"}, reporting.PageInput{Page: "invalid"})
	expect(t, err, fault.Invalid, "scope_unavailable")
	_, err = f.service.Requests(f.ctx, f.dev, reporting.Filters{Tenant: "other", From: "invalid"}, reporting.PageInput{Page: "invalid"})
	expect(t, err, fault.NotFound, "scope_unavailable")
	_, err = f.service.Requests(f.ctx, f.dev, reporting.Filters{Tenant: "tenant", From: "invalid"}, reporting.PageInput{Page: "invalid"})
	expect(t, err, fault.Invalid, "invalid_time_filter")
	_, err = f.service.Requests(f.ctx, f.dev, reporting.Filters{Tenant: "tenant"}, reporting.PageInput{Page: "0", Size: "invalid"})
	expect(t, err, fault.Invalid, "invalid_page")
	_, err = f.service.Requests(f.ctx, f.dev, reporting.Filters{Tenant: "tenant"}, reporting.PageInput{Size: "101"})
	expect(t, err, fault.Invalid, "invalid_page_size")
}

func TestUsageGroupsOnlyVisibleFilteredLedgerEntries(t *testing.T) {
	f := reports(t)
	usage, err := f.service.Usage(f.ctx, f.dev, reporting.Filters{Tenant: "tenant", State: "success"})
	if err != nil || len(usage.Items) != 1 || usage.Items[0]["requests"] != int64(1) || usage.Items[0]["fee_micro_usd"] != "23" || usage.Items[0]["cost_micro_usd"] != "14" || usage.Items[0]["confirmed_input_tokens"] != int64(4) || usage.Items[0]["pending"] != int64(1) {
		t.Fatal(usage, err)
	}
	usage, err = f.service.Usage(f.ctx, f.admin, reporting.Filters{Tenant: "tenant", Key: "unassigned"})
	if err != nil || usage.Items[0]["fee_micro_usd"] != "4" {
		t.Fatal("key-filtered ledger", usage, err)
	}
}

func TestRequestDetailKeepsEvidenceAndConcealsUnassignedRequest(t *testing.T) {
	f := reports(t)
	f.exec(t, `INSERT INTO request_attempts VALUES('one',1,1,'source','remote','2030-01-02T10:00:00Z',NULL,'canceled','client_canceled',NULL,NULL,NULL,NULL); INSERT INTO attempt_dispatches VALUES('one',1,1);
INSERT INTO request_prices VALUES('one',1,'price',1,2,3,4,10,20); INSERT INTO route_selection_evidence VALUES('one',1,1,'{}'); INSERT INTO route_latency_evidence VALUES('one',1,'{}');`)
	detail, err := f.service.Request(f.ctx, f.dev, "one", reporting.Filters{Tenant: "tenant"})
	if err != nil || detail["id"] != "one" || detail["confirmed_input_tokens"] != "4" || len(detail["attempts"].([]map[string]any)) != 1 || len(detail["charge_entries"].([]map[string]any)) != 2 {
		t.Fatal(detail, err)
	}
	_, err = f.service.Request(f.ctx, f.dev, "two", reporting.Filters{Tenant: "tenant"})
	expect(t, err, fault.NotFound, "not_found")
	f.exec(t, "UPDATE route_latency_evidence SET evidence='not-json'")
	_, err = f.service.Request(f.ctx, f.dev, "one", reporting.Filters{Tenant: "tenant"})
	expect(t, err, fault.Unavailable, "storage_unavailable")
}

func TestExportValidatesPagesAndEnforcesFilteredSizeLimit(t *testing.T) {
	f := reports(t)
	_, err := f.service.ExportRequests(f.ctx, f.dev, reporting.Filters{Tenant: "tenant"}, reporting.PageInput{Page: "invalid"})
	expect(t, err, fault.Invalid, "invalid_page")
	rows, err := f.service.ExportRequests(f.ctx, f.dev, reporting.Filters{Tenant: "tenant"}, reporting.PageInput{Page: "2", Size: "1"})
	if err != nil || len(rows) != 1 || rows[0]["confirmed_output_tokens"] != "6" {
		t.Fatal(rows, err)
	}
	f.exec(t, `WITH RECURSIVE counter(n) AS(SELECT 1 UNION ALL SELECT n+1 FROM counter WHERE n<10001) INSERT INTO requests(id,time,key_id,alias,state,error_code,stream) SELECT 'bulk-'||n,'2030-01-02T13:00:00Z','assigned','bulk','success','',0 FROM counter`)
	_, err = f.service.ExportRequests(f.ctx, f.dev, reporting.Filters{Tenant: "tenant"}, reporting.PageInput{})
	expect(t, err, fault.Invalid, "export_requires_narrower_filter")
	rows, err = f.service.ExportRequests(f.ctx, f.dev, reporting.Filters{Tenant: "tenant", Model: "model"}, reporting.PageInput{})
	if err != nil || len(rows) != 1 {
		t.Fatal("filtered export", len(rows), err)
	}
}

func TestOverviewKeepsScopedPendingReservationsAndUnknownAverages(t *testing.T) {
	f := reports(t)
	view, err := f.service.Overview(f.ctx, f.dev, "tenant")
	if err != nil || view.RequestsToday != 1 || view.SuccessRate == nil || *view.SuccessRate != 100 || view.Fee != "23" || view.Reserved != "100" || view.Pending != 1 || view.AverageDuration == nil || *view.AverageDuration != 10 || len(view.Distribution) != 1 {
		t.Fatal(view, err)
	}
	f.exec(t, "DELETE FROM key_diagnostic_members")
	view, err = f.service.Overview(f.ctx, f.dev, "tenant")
	if err != nil || view.RequestsToday != 0 || view.SuccessRate != nil || view.AverageDuration != nil || view.Fee != "0" {
		t.Fatal("empty scoped overview", view, err)
	}
}

func TestAuditAndOperationLookupsRecheckCurrentMembership(t *testing.T) {
	f := reports(t)
	_, err := f.service.Audit(f.ctx, f.dev, "tenant")
	expect(t, err, fault.Forbidden, "forbidden")
	audit, err := f.service.Audit(f.ctx, f.admin, "tenant")
	if err != nil || len(audit.Items) != 1 || audit.Items[0]["id"] != "audit-tenant" || audit.MoreRequiresFilter {
		t.Fatal(audit, err)
	}
	operation, err := f.service.Operation(f.ctx, f.admin, "own")
	if err != nil || operation.ObjectID != "one" || operation.SecretRecoverable {
		t.Fatal(operation, err)
	}
	_, err = f.service.Operation(f.ctx, f.admin, "removed")
	expect(t, err, fault.NotFound, "not_found")
	_, err = f.service.Operation(f.ctx, f.dev, "global")
	expect(t, err, fault.Forbidden, "forbidden")
}

func TestCatalogKeepsGrantScopeAndLiveExecutionPriceApplicability(t *testing.T) {
	f := reports(t)
	_, err := f.service.Catalog(f.ctx, f.dev, reporting.CatalogInput{})
	expect(t, err, fault.Invalid, "tenant_required")
	_, err = f.service.Catalog(f.ctx, f.dev, reporting.CatalogInput{Tenant: "tenant", IncludeUngranted: true})
	expect(t, err, fault.Forbidden, "forbidden")
	items, err := f.service.Catalog(f.ctx, f.dev, reporting.CatalogInput{Tenant: "tenant"})
	if err != nil || len(items) != 1 || items[0].Alias != "model" || items[0].TenantInput == nil || *items[0].TenantInput != 100 || items[0].Summary["average_duration_ms"] == nil {
		t.Fatal(items, err)
	}
	f.exec(t, "UPDATE upstream SET execution_revision=10")
	items, err = f.service.Catalog(f.ctx, f.dev, reporting.CatalogInput{Tenant: "tenant"})
	if err != nil || items[0].TenantInput != nil || items[0].Effective != nil {
		t.Fatal("stale price advertised", items, err)
	}
	items, err = f.service.Catalog(f.ctx, f.platform, reporting.CatalogInput{Tenant: "tenant", IncludeUngranted: true})
	if err != nil || len(items) != 2 {
		t.Fatal(items, err)
	}
}

func TestModelSummaryKeepsCandidateRatesDistinctFromLatestModelRate(t *testing.T) {
	f := reports(t)
	f.exec(t, `INSERT INTO upstream VALUES(2,'second source',0,2);
INSERT INTO route_candidates VALUES('model',2,0,20,1,3);
INSERT INTO prices VALUES('new-price','model',2,'2030-01-01T12:00:00Z',0,0,300,400,2,3);`)
	items, err := f.service.Catalog(f.ctx, f.dev, reporting.CatalogInput{Tenant: "tenant"})
	if err != nil || len(items) != 1 || items[0].TenantInput == nil || *items[0].TenantInput != 300 {
		t.Fatal("latest applicable rate changed for configured candidate", items, err)
	}
	sources := items[0].Summary["sources"].([]map[string]any)
	if len(sources) != 2 || sources[0]["upstream_id"] != int64(2) || sources[0]["enabled"] != int64(0) || sources[0]["tenant_input_rate"] != int64(300) || sources[0]["tenant_output_rate"] != int64(400) || sources[1]["upstream_id"] != int64(1) || sources[1]["tenant_input_rate"] != int64(100) {
		t.Fatal("per-candidate price, numeric fields or priority order changed", sources)
	}
	f.exec(t, "UPDATE upstream SET execution_revision=4 WHERE id=2")
	items, err = f.service.Catalog(f.ctx, f.dev, reporting.CatalogInput{Tenant: "tenant"})
	if err != nil || items[0].TenantInput == nil || *items[0].TenantInput != 100 {
		t.Fatal("model rate did not fall back to applicable candidate", items, err)
	}
	sources = items[0].Summary["sources"].([]map[string]any)
	if sources[0]["tenant_input_rate"] != nil || sources[0]["tenant_output_rate"] != nil || sources[1]["tenant_input_rate"] != int64(100) {
		t.Fatal("stale candidate rate did not retain null semantics", sources)
	}
}

const reportSchema = `CREATE TABLE users(id TEXT PRIMARY KEY,account TEXT,active INTEGER);
CREATE TABLE tenants(id TEXT PRIMARY KEY,name TEXT);
CREATE TABLE memberships(tenant_id TEXT,user_id TEXT,role TEXT);
CREATE TABLE key_controls(key_id TEXT PRIMARY KEY,tenant_id TEXT);
CREATE TABLE key_diagnostic_members(key_id TEXT,tenant_id TEXT,user_id TEXT);
CREATE TABLE requests(id TEXT PRIMARY KEY,time TEXT,key_id TEXT,alias TEXT,model TEXT,state TEXT,error_code TEXT,stream INTEGER,upstream_status INTEGER,duration_ms INTEGER,prompt_tokens INTEGER,completion_tokens INTEGER,total_tokens INTEGER);
CREATE TABLE admissions(request_id TEXT,tenant_id TEXT,key_id TEXT,period TEXT,pending INTEGER,tokens_reserved INTEGER,fee_reserved INTEGER,revision INTEGER);
CREATE TABLE legacy_pending(request_id TEXT,resolved INTEGER);
CREATE TABLE charge_entries(request_id TEXT,ordinal INTEGER,revision INTEGER,period TEXT,input_tokens INTEGER,output_tokens INTEGER,cost INTEGER,fee INTEGER,reason TEXT,evidence TEXT);
CREATE TABLE request_attempts(request_id TEXT,ordinal INTEGER,upstream_id INTEGER,upstream_name TEXT,remote_model TEXT,started_at TEXT,finished_at TEXT,state TEXT,error_code TEXT,upstream_status INTEGER,prompt_tokens INTEGER,completion_tokens INTEGER,total_tokens INTEGER);
CREATE TABLE attempt_dispatches(request_id TEXT,ordinal INTEGER,possible INTEGER);
CREATE TABLE request_prices(request_id TEXT,upstream_id INTEGER,price_id TEXT,cost_input INTEGER,cost_output INTEGER,tenant_input INTEGER,tenant_output INTEGER,max_input INTEGER,max_output INTEGER);
CREATE TABLE historical_price_evidence(request_id TEXT,upstream_id INTEGER,price_reference TEXT,cost_input INTEGER,cost_output INTEGER,tenant_input INTEGER,tenant_output INTEGER,reason TEXT,evidence TEXT);
CREATE TABLE route_selection_evidence(request_id TEXT,selection_round INTEGER,upstream_id INTEGER,evidence TEXT);
CREATE TABLE route_latency_evidence(request_id TEXT,ordinal INTEGER,evidence TEXT);
CREATE TABLE console_audit(id TEXT,time TEXT,actor TEXT,tenant_id TEXT,action TEXT,object_id TEXT);
CREATE TABLE console_operations(actor TEXT,operation TEXT,object_id TEXT,action TEXT,tenant_id TEXT);
CREATE TABLE models(alias TEXT,enabled INTEGER,revision INTEGER);
CREATE TABLE model_route_policies(alias TEXT,strategy TEXT);
CREATE TABLE tenant_model_grants(tenant_id TEXT,alias TEXT);
CREATE TABLE upstream(id INTEGER,name TEXT,enabled INTEGER,execution_revision INTEGER);
CREATE TABLE route_candidates(alias TEXT,upstream_id INTEGER,enabled INTEGER,priority INTEGER,weight INTEGER,execution_revision INTEGER);
CREATE TABLE service_profiles(upstream_id INTEGER,source_kind TEXT);
CREATE TABLE prices(id TEXT,alias TEXT,upstream_id INTEGER,effective TEXT,cost_input INTEGER,cost_output INTEGER,tenant_input INTEGER,tenant_output INTEGER,upstream_execution INTEGER,candidate_execution INTEGER);`
