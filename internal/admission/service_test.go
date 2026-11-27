package admission_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/admission"
	"model-relay/internal/apikey"
	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/routing"
)

type fixture struct {
	store         *installation.Store
	service       *admission.Service
	now           time.Time
	keyID, secret string
	snapshot      catalog.Snapshot
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	root := t.TempDir()
	directory, keyFile := filepath.Join(root, "data"), filepath.Join(root, "material")
	if _, err := installation.InitConsole(directory, keyFile); err != nil {
		t.Fatal(err)
	}
	store, err := installation.Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	f := &fixture{store: store, now: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}
	store.SetClock(func() time.Time { return f.now })
	if err = store.AdvanceClock(); err != nil {
		t.Fatal(err)
	}
	if _, err = store.DB.Exec(`INSERT INTO tenants VALUES('tenant','Tenant',1,'2026-01-01T00:00:00Z');
INSERT INTO models VALUES('alpha',1,1,NULL);
INSERT INTO upstream VALUES(1,'Provider','https://public.example/v1',x'',1,1,1);
INSERT INTO tenant_model_grants VALUES('tenant','alpha');
INSERT INTO resource_limits VALUES('tenant','tenant',NULL,NULL,NULL,NULL,NULL,1,1);
INSERT INTO prices VALUES('price','alpha',1,'2026-01-01T00:00:00Z',1000000,1000000,2000000,3000000,1,1);
INSERT INTO model_bounds VALUES('alpha',1,10,20,'trusted evidence',1,1);`); err != nil {
		t.Fatal(err)
	}
	key, err := apikey.New(store.DB, store.Now, configstate.New(), routing.New()).Create(context.Background(), access.Principal{ID: "admin", Platform: true}, "tenant", apikey.Operation{Key: "create", Path: "/keys"}, &apikey.Input{Name: "Application", Application: "Application", Enabled: true, Models: []string{"alpha"}})
	if err != nil {
		t.Fatal(err)
	}
	f.keyID, f.secret = key.ID, key.Secret
	f.snapshot = catalog.Snapshot{Model: catalog.RouteModel{Alias: "alpha", Enabled: true, Revision: 1, Candidates: []catalog.Candidate{{UpstreamID: 1, RemoteModel: "remote", Enabled: true, ExecutionRevision: 1}}}, Upstreams: map[int64]catalog.Upstream{1: {ID: 1, Enabled: true, ExecutionRevision: 1}}}
	f.service = admission.New(store.DB, store.Now, store.AdvanceClock)
	return f
}

func payload(extra string) map[string]json.RawMessage {
	var value map[string]json.RawMessage
	json.Unmarshal([]byte(`{"model":"alpha","messages":[{"role":"user","content":"controlled"}]`+extra+`}`), &value)
	return value
}

func (f *fixture) request(t *testing.T, id string) records.Record {
	t.Helper()
	record := records.Record{ID: id, KeyID: f.keyID, Alias: "alpha", Time: f.now.Format(time.RFC3339Nano)}
	if _, err := f.store.DB.Exec("INSERT INTO requests(id,time,key_id,alias,stream,state,error_code) VALUES(?,?,?,?,0,'started','')", id, record.Time, f.keyID, "alpha"); err != nil {
		t.Fatal(err)
	}
	return record
}

func expect(t *testing.T, code string, status int, prior, wantCode string, wantStatus int, wantPrior string) {
	t.Helper()
	if code != wantCode || status != wantStatus || prior != wantPrior {
		t.Fatalf("admission=(%s,%d,%s), want (%s,%d,%s)", code, status, prior, wantCode, wantStatus, wantPrior)
	}
}

func TestAdmissionFreezesPriceBoundsAndLogicalCall(t *testing.T) {
	f := newFixture(t)
	if _, err := f.store.DB.Exec("UPDATE resource_limits SET tpm=100,budget=200 WHERE owner_type='tenant'"); err != nil {
		t.Fatal(err)
	}
	data := payload("")
	record := f.request(t, "first")
	code, status, prior := f.service.Admit(record, f.snapshot, data, f.secret, "logical-call")
	expect(t, code, status, prior, "", 200, "")
	if string(data["max_tokens"]) != "20" {
		t.Fatal("bounded output default not inserted", string(data["max_tokens"]))
	}
	var tokens, fees, input, output, rate int64
	if f.store.DB.QueryRow("SELECT tokens_reserved,fee_reserved FROM admissions WHERE request_id='first'").Scan(&tokens, &fees) != nil || tokens != 30 || fees != 80 {
		t.Fatal("reservation changed", tokens, fees)
	}
	if f.store.DB.QueryRow("SELECT max_input,max_output,tenant_output FROM request_prices WHERE request_id='first'").Scan(&input, &output, &rate) != nil || input != 10 || output != 20 || rate != 3000000 {
		t.Fatal("price/bounds snapshot changed")
	}
	code, status, prior = f.service.Admit(f.request(t, "repeat"), f.snapshot, payload(""), f.secret, "logical-call")
	expect(t, code, status, prior, "logical_call_already_exists", 409, "first")
	code, status, prior = f.service.Admit(f.request(t, "different"), f.snapshot, payload(`,"temperature":0.5`), f.secret, "logical-call")
	expect(t, code, status, prior, "idempotency_conflict", 409, "")
	var count int
	if f.store.DB.QueryRow("SELECT count(*) FROM admissions").Scan(&count) != nil || count != 1 {
		t.Fatal("logical call replay reserved twice")
	}
}

func TestQuotaDimensionsAndTenantKeyPrecedence(t *testing.T) {
	for _, scenario := range []struct {
		column string
		limit  int64
		want   string
	}{
		{"calls", 0, "tenant_calls_exceeded"}, {"rpm", 0, "tenant_rpm_exceeded"}, {"tpm", 29, "tenant_tpm_exceeded"}, {"concurrency", 0, "tenant_concurrency_exceeded"}, {"budget", 79, "tenant_budget_exceeded"},
	} {
		t.Run(scenario.column, func(t *testing.T) {
			f := newFixture(t)
			if _, err := f.store.DB.Exec("UPDATE resource_limits SET "+scenario.column+"=? WHERE owner_type='tenant'; UPDATE resource_limits SET calls=0 WHERE owner_type='key'", scenario.limit); err != nil {
				t.Fatal(err)
			}
			code, status, prior := f.service.Admit(f.request(t, "limited"), f.snapshot, payload(""), f.secret, "retryable")
			expect(t, code, status, prior, scenario.want, 429, "")
			var operations int
			if f.store.DB.QueryRow("SELECT count(*) FROM generation_operations").Scan(&operations) != nil || operations != 0 {
				t.Fatal("quota rejection consumed operation key")
			}
		})
	}
}

func TestMissingPriceAndBoundRules(t *testing.T) {
	for _, scenario := range []struct {
		name, setup, extra, want string
		status                   int
	}{
		{"price_revision", `UPDATE prices SET upstream_execution=2`, "", "price_not_ready", 503},
		{"bound_revision", `UPDATE resource_limits SET tpm=100 WHERE owner_type='tenant'; UPDATE model_bounds SET candidate_execution=2`, "", "trusted_bound_required", 503},
		{"unsupported_tools", `UPDATE resource_limits SET tpm=100 WHERE owner_type='tenant'`, `,"tools":[]`, "unsupported_bounded_request", 400},
		{"alternatives", `UPDATE resource_limits SET tpm=100 WHERE owner_type='tenant'`, `,"n":2`, "unsupported_bounded_request", 400},
		{"excess_output", `UPDATE resource_limits SET tpm=100 WHERE owner_type='tenant'`, `,"max_tokens":21`, "unsupported_bounded_request", 400},
		{"both_output_fields", `UPDATE resource_limits SET tpm=100 WHERE owner_type='tenant'`, `,"max_tokens":1,"max_completion_tokens":1`, "unsupported_bounded_request", 400},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			f := newFixture(t)
			if _, err := f.store.DB.Exec(scenario.setup); err != nil {
				t.Fatal(err)
			}
			code, status, prior := f.service.Admit(f.request(t, "request"), f.snapshot, payload(scenario.extra), f.secret, "")
			expect(t, code, status, prior, scenario.want, scenario.status, "")
		})
	}
}

func TestAuthorizationRecheckedBeforeOperationReplay(t *testing.T) {
	f := newFixture(t)
	code, status, prior := f.service.Admit(f.request(t, "first"), f.snapshot, payload(""), f.secret, "operation")
	expect(t, code, status, prior, "", 200, "")
	if _, err := f.store.DB.Exec("DELETE FROM key_model_grants WHERE key_id=?; UPDATE tenants SET active=0 WHERE id='tenant'", f.keyID); err != nil {
		t.Fatal(err)
	}
	code, status, prior = f.service.Admit(f.request(t, "forbidden"), f.snapshot, payload(""), f.secret, "operation")
	expect(t, code, status, prior, "model_forbidden", 403, "")
	if _, err := f.store.DB.Exec("INSERT INTO key_model_grants VALUES(?,'alpha')", f.keyID); err != nil {
		t.Fatal(err)
	}
	code, status, prior = f.service.Admit(f.request(t, "replay"), f.snapshot, payload(""), f.secret, "operation")
	expect(t, code, status, prior, "logical_call_already_exists", 409, "first")
	code, status, prior = f.service.Admit(f.request(t, "paused"), f.snapshot, payload(""), f.secret, "new-operation")
	expect(t, code, status, prior, "tenant_paused", 403, "")
	if _, err := f.store.DB.Exec("UPDATE key_controls SET revoked=1 WHERE key_id=?", f.keyID); err != nil {
		t.Fatal(err)
	}
	code, status, prior = f.service.Admit(f.request(t, "revoked"), f.snapshot, payload(""), f.secret, "operation")
	expect(t, code, status, prior, "unauthorized", 401, "")
}

func TestSnapshotWriteFailureRollsBackAdmissionAndOperation(t *testing.T) {
	f := newFixture(t)
	if _, err := f.store.DB.Exec(`CREATE TRIGGER reject_price_snapshot BEFORE INSERT ON request_prices BEGIN SELECT RAISE(FAIL,'controlled'); END;`); err != nil {
		t.Fatal(err)
	}
	record := f.request(t, "request")
	code, status, prior := f.service.Admit(record, f.snapshot, payload(""), f.secret, "operation")
	expect(t, code, status, prior, "storage_unavailable", 503, "")
	for _, table := range []string{"admissions", "generation_operations", "request_prices"} {
		var count int
		if f.store.DB.QueryRow("SELECT count(*) FROM "+table).Scan(&count) != nil || count != 0 {
			t.Fatal("partial admission committed", table, count)
		}
	}
	if _, err := f.store.DB.Exec("DROP TRIGGER reject_price_snapshot"); err != nil {
		t.Fatal(err)
	}
	code, status, prior = f.service.Admit(record, f.snapshot, payload(""), f.secret, "operation")
	expect(t, code, status, prior, "", 200, "")
}

func TestDurableClockPreventsExpiredCredentialAdmission(t *testing.T) {
	f := newFixture(t)
	expiry := f.now.Add(time.Hour)
	if _, err := f.store.DB.Exec("UPDATE key_controls SET expires=? WHERE key_id=?", expiry.Format(time.RFC3339Nano), f.keyID); err != nil {
		t.Fatal(err)
	}
	f.now = expiry
	code, status, prior := f.service.Admit(f.request(t, "deadline"), f.snapshot, payload(""), f.secret, "")
	expect(t, code, status, prior, "unauthorized", 401, "")
	f.now = expiry.Add(-time.Hour)
	code, status, prior = f.service.Admit(f.request(t, "rollback"), f.snapshot, payload(""), f.secret, "")
	expect(t, code, status, prior, "unauthorized", 401, "")
}

func TestConcurrentAdmissionsRespectNonzeroCallQuota(t *testing.T) {
	f := newFixture(t)
	if _, err := f.store.DB.Exec("UPDATE resource_limits SET calls=1 WHERE owner_type='tenant'"); err != nil {
		t.Fatal(err)
	}
	first, second := f.request(t, "first"), f.request(t, "second")
	type result struct {
		code   string
		status int
		prior  string
	}
	results := make(chan result, 2)
	for _, record := range []records.Record{first, second} {
		go func(record records.Record) {
			code, status, prior := f.service.Admit(record, f.snapshot, payload(""), f.secret, record.ID)
			results <- result{code, status, prior}
		}(record)
	}
	accepted, rejected := 0, 0
	for range 2 {
		select {
		case result := <-results:
			if result.code == "" && result.status == 200 && result.prior == "" {
				accepted++
			} else if result.code == "tenant_calls_exceeded" && result.status == 429 && result.prior == "" {
				rejected++
			} else {
				t.Fatal("unexpected concurrent result", result)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("concurrent admission did not finish")
		}
	}
	if accepted != 1 || rejected != 1 {
		t.Fatal("nonzero call quota oversubscribed", accepted, rejected)
	}
	var admissions, operations int
	if f.store.DB.QueryRow("SELECT count(*) FROM admissions").Scan(&admissions) != nil || admissions != 1 || f.store.DB.QueryRow("SELECT count(*) FROM generation_operations").Scan(&operations) != nil || operations != 1 {
		t.Fatal("rejected concurrent attempt persisted state", admissions, operations)
	}
}

func TestPendingHistoryBlocksNewBoundedAdmission(t *testing.T) {
	for _, scenario := range []string{"historical", "unbounded"} {
		t.Run(scenario, func(t *testing.T) {
			f := newFixture(t)
			if scenario == "historical" {
				f.request(t, "old")
				if _, err := f.store.DB.Exec("INSERT INTO legacy_pending VALUES('old','tenant','2026-01',0)"); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, err := f.store.DB.Exec("DELETE FROM model_bounds"); err != nil {
					t.Fatal(err)
				}
				code, status, prior := f.service.Admit(f.request(t, "old"), f.snapshot, payload(""), f.secret, "")
				expect(t, code, status, prior, "", 200, "")
				if _, err := f.store.DB.Exec("UPDATE admissions SET running=0,pending=1 WHERE request_id='old'"); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := f.store.DB.Exec("UPDATE resource_limits SET tpm=100 WHERE owner_type='tenant'"); err != nil {
				t.Fatal(err)
			}
			code, status, prior := f.service.Admit(f.request(t, "new"), f.snapshot, payload(""), f.secret, "")
			want := "historical_usage_pending"
			if scenario == "unbounded" {
				want = "unbounded_usage_pending"
			}
			expect(t, code, status, prior, want, 503, "")
		})
	}
}
