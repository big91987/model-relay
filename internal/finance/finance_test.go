package finance_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"model-relay/internal/access"
	"model-relay/internal/accounting"
	"model-relay/internal/fault"
	"model-relay/internal/finance"
	"model-relay/internal/operations"
)

type fixture struct {
	db        *sql.DB
	service   *finance.Service
	ctx       context.Context
	platform  access.Principal
	admin     access.Principal
	developer access.Principal
}

func install(t *testing.T) fixture {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "finance.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	f := fixture{db: db, ctx: context.Background(), platform: access.Principal{ID: "platform", Platform: true}, admin: access.Principal{ID: "admin"}, developer: access.Principal{ID: "dev"}}
	f.exec(t, fixtureSchema)
	f.exec(t, `INSERT INTO tenants VALUES('tenant'); INSERT INTO users VALUES('admin',1),('dev',1),('finance',1); INSERT INTO memberships VALUES('tenant','admin','administrator'),('tenant','dev','developer'),('tenant','finance','finance');
INSERT INTO app_keys VALUES('key'); INSERT INTO key_controls VALUES('key','tenant'); INSERT INTO key_diagnostic_members VALUES('key','tenant','dev');
INSERT INTO models VALUES('model'); INSERT INTO upstream VALUES(1,7),(2,8); INSERT INTO route_candidates VALUES('model',1,11),('model',2,12);`)
	f.service = finance.New(db, func() time.Time { return time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC) })
	return f
}

func (f fixture) exec(t *testing.T, statement string, args ...any) {
	t.Helper()
	if _, err := f.db.Exec(statement, args...); err != nil {
		t.Fatal(err)
	}
}

func (f fixture) scalar(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var value int64
	if err := f.db.QueryRow(query, args...).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func expectFault(t *testing.T, err error, kind fault.Kind, code string) {
	t.Helper()
	var value *fault.Error
	if !errors.As(err, &value) || value.Kind != kind || value.Code != code {
		t.Fatalf("want fault %v/%s, got %v", kind, code, err)
	}
}

func result(t *testing.T, value operations.Result) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(value.Value)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	return fields
}

func (f fixture) seedCharge(t *testing.T) {
	f.exec(t, `INSERT INTO requests VALUES('request','2026-10-01T12:00:00Z','key','model','success');
INSERT INTO admissions VALUES('request','tenant','key','2026-10','2026-10-01T12:00','2026-10-01T12:00:00Z',0,0,0,0,1);
INSERT INTO charge_entries VALUES('charge','request',1,0,'2026-10',3,4,11,18,'','');`)
}

func TestBillsCloseReplayCorrectionsAndExportPreservePeriodTotals(t *testing.T) {
	f := install(t)
	f.seedCharge(t)
	closed, err := f.service.CloseBill(f.ctx, f.platform, "tenant", "close", "/admin/api/tenants/tenant/bills", &finance.CloseBillInput{Period: "2026-10"})
	if err != nil || !closed.Created {
		t.Fatal("close", closed, err)
	}
	fields := result(t, closed)
	id := fields["id"].(string)
	if fields["original_fee"] != "18" || fields["original_cost"] != "11" {
		t.Fatal(fields)
	}
	replay, err := f.service.CloseBill(f.ctx, f.platform, "tenant", "close", "/admin/api/tenants/tenant/bills", &finance.CloseBillInput{Period: "2026-10"})
	if err != nil || replay.Created || result(t, replay)["already_applied"] != true {
		t.Fatal("replay", replay, err)
	}
	again, err := f.service.CloseBill(f.ctx, f.platform, "tenant", "close-again", "/admin/api/tenants/tenant/bills", &finance.CloseBillInput{Period: "2026-10"})
	if err != nil || result(t, again)["already_closed"] != true {
		t.Fatal("already closed", again, err)
	}
	correction := &finance.CorrectionInput{Amount: -1, Reason: "credit"}
	corrected, err := f.service.CorrectBill(f.ctx, f.platform, id, "credit", "/admin/api/bills/"+id+"/corrections", correction)
	if err != nil || !corrected.Created || correction.Period != "" {
		t.Fatal("correction", corrected, err)
	}
	replayed, err := f.service.CorrectBill(f.ctx, f.platform, id, "credit", "/admin/api/bills/"+id+"/corrections", correction)
	if err != nil || result(t, replayed)["already_applied"] != true {
		t.Fatal("correction replay", replayed, err)
	}
	detail, err := f.service.Bill(f.ctx, f.admin, id)
	if err != nil || detail.OriginalFee != "18" || detail.CurrentFee != "17" || detail.Adjustment != "-1" {
		t.Fatal("bill detail", detail, err)
	}
	periods, err := f.service.ListBills(f.ctx, f.admin, "tenant")
	if err != nil || len(periods) != 2 || periods[0].Period != "2026-11" || periods[0].Fee != "-1" || periods[0].BillID != nil {
		t.Fatal("adjustment-only period", periods, err)
	}
	entries, err := f.service.ExportBill(f.ctx, f.admin, id)
	if err != nil || len(entries) != 1 || entries[0].Fee == nil || *entries[0].Fee != "18" || entries[0].Ordinal == nil || *entries[0].Ordinal != 1 {
		t.Fatal("bill export", entries, err)
	}
	if f.scalar(t, "SELECT count(*) FROM console_audit") != 3 {
		t.Fatal("replays duplicated audits")
	}
}

func TestBillCloseAndCorrectionKeepFailurePrecedenceAndRollback(t *testing.T) {
	f := install(t)
	f.seedCharge(t)
	_, err := f.service.CloseBill(f.ctx, f.developer, "tenant", "", "", nil)
	expectFault(t, err, fault.Forbidden, "forbidden")
	_, err = f.service.CloseBill(f.ctx, f.platform, "tenant", "", "", nil)
	expectFault(t, err, fault.Conflict, "period_not_ended")
	f.exec(t, "UPDATE admissions SET pending=1,running=1")
	_, err = f.service.CloseBill(f.ctx, f.platform, "tenant", "close", "path", &finance.CloseBillInput{Period: "2026-10"})
	expectFault(t, err, fault.Conflict, "pending_usage")
	if f.scalar(t, "SELECT count(*) FROM console_operations") != 0 {
		t.Fatal("rejected close recorded operation")
	}
	f.exec(t, "UPDATE admissions SET pending=0,running=0")
	closed, err := f.service.CloseBill(f.ctx, f.platform, "tenant", "close", "path", &finance.CloseBillInput{Period: "2026-10"})
	if err != nil {
		t.Fatal(err)
	}
	id := result(t, closed)["id"].(string)
	_, err = f.service.CorrectBill(f.ctx, f.platform, id, "negative", "path", &finance.CorrectionInput{Amount: -19, Reason: "too large"})
	expectFault(t, err, fault.Conflict, "negative_bill_total")
	f.exec(t, "CREATE TRIGGER reject_audit BEFORE INSERT ON console_audit BEGIN SELECT RAISE(ABORT,'test audit failure'); END")
	_, err = f.service.CorrectBill(f.ctx, f.platform, id, "credit", "path", &finance.CorrectionInput{Amount: -1, Reason: "credit"})
	expectFault(t, err, fault.Unavailable, "storage_unavailable")
	if f.scalar(t, "SELECT count(*) FROM bill_corrections") != 0 {
		t.Fatal("audit failure left correction committed")
	}
}

func TestPricePublicationIsAtomicAndBindsIdempotencyToOriginalPath(t *testing.T) {
	f := install(t)
	command := &finance.PublishPricesInput{Prices: []accounting.PriceSnapshot{{Upstream: 1, CostIn: 1, FeeIn: 2}, {Upstream: 2, CostIn: 3, FeeIn: 4}}}
	_, err := f.service.PublishPrices(f.ctx, f.platform, "model", "price", "/admin/api/models/model/prices", command)
	expectFault(t, err, fault.Invalid, "tenant_rates_must_match")
	if f.scalar(t, "SELECT count(*) FROM prices") != 0 || f.scalar(t, "SELECT count(*) FROM console_operations") != 0 {
		t.Fatal("invalid second price left partial publication")
	}
	command.Prices[1].FeeIn = 2
	published, err := f.service.PublishPrices(f.ctx, f.platform, "model", "price", "/admin/api/models/model/prices", command)
	if err != nil || !published.Created {
		t.Fatal("publish", published, err)
	}
	replay, err := f.service.PublishPrices(f.ctx, f.platform, "model", "price", "/admin/api/models/model/prices", command)
	if err != nil || replay.Created || result(t, replay)["already_applied"] != true {
		t.Fatal("publish replay", replay, err)
	}
	_, err = f.service.PublishPrices(f.ctx, f.platform, "model", "price", "/different-path", command)
	expectFault(t, err, fault.Conflict, "idempotency_conflict")
	prices, err := f.service.Prices(f.ctx, f.platform, "model")
	if err != nil || len(prices) != 2 || prices[0].Effective != "2026-11-09T12:00:00Z" {
		t.Fatal(prices, err)
	}
	_, err = f.service.PublishPrices(f.ctx, f.developer, "missing", "", "", nil)
	expectFault(t, err, fault.Forbidden, "forbidden")
	_, err = f.service.PublishPrices(f.ctx, f.platform, "missing", "", "", nil)
	expectFault(t, err, fault.NotFound, "not_found")
}

func TestBoundsCaptureBothExecutionRevisions(t *testing.T) {
	f := install(t)
	_, err := f.service.DeclareBound(f.ctx, f.platform, "model", &finance.BoundInput{Upstream: 1, Input: 10, Output: 100, Evidence: "operator evidence"})
	if err != nil {
		t.Fatal(err)
	}
	bounds, err := f.service.Bounds(f.ctx, f.platform, "model")
	if err != nil || len(bounds) != 1 || bounds[0].Applicable != 1 || bounds[0].UpstreamExecution != 7 || bounds[0].CandidateExecution != 11 {
		t.Fatal(bounds, err)
	}
	f.exec(t, "UPDATE upstream SET execution_revision=8 WHERE id=1")
	bounds, err = f.service.Bounds(f.ctx, f.platform, "model")
	if err != nil || bounds[0].Applicable != 0 {
		t.Fatal("stale bound applicable", bounds, err)
	}
}

func TestLimitsKeepPlatformCeilingRevisionsAndAssignedKeyIsolation(t *testing.T) {
	f := install(t)
	scope := finance.LimitScope{TenantID: "tenant", OwnerType: "tenant", OwnerID: "tenant"}
	calls := int64(5)
	_, err := f.service.UpdateLimits(f.ctx, f.platform, scope, &accounting.ResourceLimits{Calls: &calls, Hard: true})
	if err != nil {
		t.Fatal(err)
	}
	calls = 6
	_, err = f.service.UpdateLimits(f.ctx, f.admin, scope, &accounting.ResourceLimits{Calls: &calls, Hard: true, Revision: 1})
	expectFault(t, err, fault.Forbidden, "platform_limit_exceeded")
	calls = 4
	_, err = f.service.UpdateLimits(f.ctx, f.admin, scope, &accounting.ResourceLimits{Calls: &calls, Hard: true, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	view, err := f.service.Limits(f.ctx, f.admin, scope)
	if err != nil || view.Limits.Revision != 2 || view.Platform == nil || *view.Platform.Calls != 5 || view.BudgetRemaining != nil {
		t.Fatal(view, err)
	}
	_, err = f.service.Limits(f.ctx, f.developer, scope)
	expectFault(t, err, fault.Forbidden, "assigned_key_scope_required")
	keyScope := finance.LimitScope{TenantID: "tenant", OwnerType: "key", OwnerID: "key"}
	if _, err = f.service.Limits(f.ctx, f.developer, keyScope); err != nil {
		t.Fatal("assigned developer key", err)
	}
	_, err = f.service.UpdateLimits(f.ctx, f.developer, keyScope, nil)
	expectFault(t, err, fault.NotFound, "not_found")
	f.exec(t, "DELETE FROM key_diagnostic_members")
	_, err = f.service.Limits(f.ctx, f.developer, keyScope)
	expectFault(t, err, fault.NotFound, "not_found")
}

func (f fixture) seedPending(t *testing.T) {
	f.exec(t, `INSERT INTO requests VALUES('request','2026-10-01T12:00:00Z','key','model','canceled');
INSERT INTO admissions VALUES('request','tenant','key','2026-10','2026-10-01T12:00','2026-10-01T12:00:00Z',1,110,320,0,1);
INSERT INTO request_prices VALUES('request',1,'price',1000000,2000000,2000000,3000000,10,100);
INSERT INTO request_attempts VALUES('request',1,1,3,NULL,NULL); INSERT INTO attempt_dispatches VALUES('request',1,1);`)
	tx, err := f.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err = finance.Settle(tx, "request"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestUsageReconciliationAppendsDeltaAndReplaysWithoutDuplicateCharge(t *testing.T) {
	f := install(t)
	f.seedPending(t)
	if f.scalar(t, "SELECT pending FROM admissions") != 1 || f.scalar(t, "SELECT fee_reserved FROM admissions") != 300 {
		t.Fatal("unknown output lost reservation")
	}
	command := &finance.ReconcileInput{Input: 3, Output: 4, Reason: "supplier trace", Evidence: "trace reference", Revision: 1}
	applied, err := f.service.ReconcileUsage(f.ctx, f.platform, "request", 1, "reconcile", "path", command)
	if err != nil || applied.Created || result(t, applied)["revision"] != float64(2) {
		t.Fatal(applied, err)
	}
	replay, err := f.service.ReconcileUsage(f.ctx, f.platform, "request", 1, "reconcile", "path", command)
	if err != nil || result(t, replay)["already_applied"] != true {
		t.Fatal(replay, err)
	}
	if f.scalar(t, "SELECT sum(cost) FROM charge_entries") != 11 || f.scalar(t, "SELECT sum(fee) FROM charge_entries") != 18 || f.scalar(t, "SELECT count(*) FROM charge_entries") != 2 {
		t.Fatal("settlement changed prior charge or duplicated cumulative usage")
	}
	if f.scalar(t, "SELECT pending+fee_reserved+tokens_reserved FROM admissions") != 0 {
		t.Fatal("confirmed usage retained reservation")
	}
	command.Revision = 2
	command.Input = 2
	_, err = f.service.ReconcileUsage(f.ctx, f.platform, "request", 1, "decrease", "path", command)
	expectFault(t, err, fault.Conflict, "cumulative_usage_cannot_decrease")
}

func TestUsageReconciliationAuditFailureRollsBackEveryFinancialWrite(t *testing.T) {
	f := install(t)
	f.seedPending(t)
	f.exec(t, "CREATE TRIGGER reject_audit BEFORE INSERT ON console_audit BEGIN SELECT RAISE(ABORT,'test audit failure'); END")
	_, err := f.service.ReconcileUsage(f.ctx, f.platform, "request", 1, "reconcile", "path", &finance.ReconcileInput{Input: 3, Output: 4, Reason: "supplier trace", Evidence: "trace", Revision: 1})
	expectFault(t, err, fault.Unavailable, "storage_unavailable")
	if f.scalar(t, "SELECT revision FROM admissions") != 1 || f.scalar(t, "SELECT count(*) FROM usage_reconciliations") != 0 || f.scalar(t, "SELECT count(*) FROM charge_entries") != 1 || f.scalar(t, "SELECT pending FROM admissions") != 1 {
		t.Fatal("failed reconciliation leaked financial writes")
	}
}

func TestLegacyReconciliationUsesHistoricalRatesAndPreservesExecutionHistory(t *testing.T) {
	f := install(t)
	f.exec(t, `INSERT INTO requests VALUES('legacy','2026-10-01T12:00:00Z','key','model','process_interrupted'); INSERT INTO legacy_pending VALUES('legacy','tenant','2026-10',0);`)
	command := &finance.LegacyInput{Reason: "historical receipt", Evidence: "archive", Attempts: []finance.LegacyAttempt{{Ordinal: 0, Upstream: 1, PriceID: "historical", Input: 1, Output: 1, CostIn: 500000, CostOut: 500000, FeeIn: 750000, FeeOut: 750000}}}
	_, err := f.service.ReconcileLegacy(f.ctx, f.platform, "legacy", "legacy-op", "path", command)
	expectFault(t, err, fault.Invalid, "invalid_historical_attempt")
	if f.scalar(t, "SELECT count(*) FROM admissions") != 0 {
		t.Fatal("invalid historical evidence left new admission")
	}
	command.Attempts[0].Ordinal = 1
	_, err = f.service.ReconcileLegacy(f.ctx, f.platform, "legacy", "legacy-op", "path", command)
	if err != nil {
		t.Fatal(err)
	}
	if f.scalar(t, "SELECT sum(cost) FROM charge_entries") != 1 || f.scalar(t, "SELECT sum(fee) FROM charge_entries") != 2 || f.scalar(t, "SELECT resolved FROM legacy_pending") != 1 {
		t.Fatal("historical charge rounding or resolution")
	}
	var state string
	if err = f.db.QueryRow("SELECT state FROM requests WHERE id='legacy'").Scan(&state); err != nil || state != "process_interrupted" {
		t.Fatal("execution history changed", state, err)
	}
	replay, err := f.service.ReconcileLegacy(f.ctx, f.platform, "legacy", "legacy-op", "path", command)
	if err != nil || result(t, replay)["already_applied"] != true {
		t.Fatal(replay, err)
	}
}

func TestCanceledContextCannotPublishPrices(t *testing.T) {
	f := install(t)
	ctx, cancel := context.WithCancel(f.ctx)
	cancel()
	_, err := f.service.PublishPrices(ctx, f.platform, "model", "price", "path", &finance.PublishPricesInput{Prices: []accounting.PriceSnapshot{{Upstream: 1}, {Upstream: 2}}})
	expectFault(t, err, fault.Unavailable, "storage_unavailable")
	if f.scalar(t, "SELECT count(*) FROM prices") != 0 {
		t.Fatal("canceled operation published")
	}
}

const fixtureSchema = `CREATE TABLE tenants(id TEXT PRIMARY KEY);
CREATE TABLE users(id TEXT PRIMARY KEY,active INTEGER);
CREATE TABLE memberships(tenant_id TEXT,user_id TEXT,role TEXT);
CREATE TABLE app_keys(id TEXT PRIMARY KEY);
CREATE TABLE key_controls(key_id TEXT PRIMARY KEY,tenant_id TEXT);
CREATE TABLE key_diagnostic_members(key_id TEXT,tenant_id TEXT,user_id TEXT);
CREATE TABLE models(alias TEXT PRIMARY KEY);
CREATE TABLE upstream(id INTEGER PRIMARY KEY,execution_revision INTEGER);
CREATE TABLE route_candidates(alias TEXT,upstream_id INTEGER,execution_revision INTEGER);
CREATE TABLE resource_limits(owner_type TEXT,owner_id TEXT,calls INTEGER,rpm INTEGER,tpm INTEGER,concurrency INTEGER,budget INTEGER,hard INTEGER,revision INTEGER,PRIMARY KEY(owner_type,owner_id));
CREATE TABLE platform_limits(tenant_id TEXT PRIMARY KEY,calls INTEGER,rpm INTEGER,tpm INTEGER,concurrency INTEGER,budget INTEGER,hard INTEGER,revision INTEGER);
CREATE TABLE model_bounds(alias TEXT,upstream_id INTEGER,max_input INTEGER,max_output INTEGER,evidence TEXT,upstream_execution INTEGER,candidate_execution INTEGER,PRIMARY KEY(alias,upstream_id));
CREATE TABLE prices(id TEXT PRIMARY KEY,alias TEXT,upstream_id INTEGER,effective TEXT,cost_input INTEGER,cost_output INTEGER,tenant_input INTEGER,tenant_output INTEGER,upstream_execution INTEGER,candidate_execution INTEGER);
CREATE TABLE requests(id TEXT PRIMARY KEY,time TEXT,key_id TEXT,alias TEXT,state TEXT);
CREATE TABLE admissions(request_id TEXT PRIMARY KEY,tenant_id TEXT,key_id TEXT,period TEXT,minute TEXT,accepted TEXT,running INTEGER,tokens_reserved INTEGER,fee_reserved INTEGER,pending INTEGER,revision INTEGER);
CREATE TABLE request_prices(request_id TEXT,upstream_id INTEGER,price_id TEXT,cost_input INTEGER,cost_output INTEGER,tenant_input INTEGER,tenant_output INTEGER,max_input INTEGER,max_output INTEGER,PRIMARY KEY(request_id,upstream_id));
CREATE TABLE request_attempts(request_id TEXT,ordinal INTEGER,upstream_id INTEGER,prompt_tokens INTEGER,completion_tokens INTEGER,total_tokens INTEGER);
CREATE TABLE attempt_dispatches(request_id TEXT,ordinal INTEGER,possible INTEGER);
CREATE TABLE charge_entries(id TEXT PRIMARY KEY,request_id TEXT,ordinal INTEGER,revision INTEGER,period TEXT,input_tokens INTEGER,output_tokens INTEGER,cost INTEGER,fee INTEGER,reason TEXT,evidence TEXT,UNIQUE(request_id,ordinal,revision));
CREATE TABLE usage_reconciliations(request_id TEXT,ordinal INTEGER,revision INTEGER,input_tokens INTEGER,output_tokens INTEGER,reason TEXT,evidence TEXT,PRIMARY KEY(request_id,ordinal,revision));
CREATE TABLE historical_price_evidence(request_id TEXT,upstream_id INTEGER,price_reference TEXT,cost_input INTEGER,cost_output INTEGER,tenant_input INTEGER,tenant_output INTEGER,reason TEXT,evidence TEXT,PRIMARY KEY(request_id,upstream_id));
CREATE TABLE bills(id TEXT PRIMARY KEY,tenant_id TEXT,period TEXT,original_cost INTEGER,original_fee INTEGER,closed TEXT,UNIQUE(tenant_id,period));
CREATE TABLE bill_corrections(id TEXT PRIMARY KEY,bill_id TEXT,period TEXT,amount INTEGER,reason TEXT,actor TEXT,created TEXT);
CREATE TABLE legacy_pending(request_id TEXT PRIMARY KEY,tenant_id TEXT,period TEXT,resolved INTEGER);
CREATE TABLE console_operations(actor TEXT,operation TEXT,fingerprint BLOB,object_id TEXT,action TEXT,tenant_id TEXT,PRIMARY KEY(actor,operation));
CREATE TABLE console_audit(id TEXT PRIMARY KEY,time TEXT,actor TEXT,tenant_id TEXT,action TEXT,object_id TEXT);`
