package catalog_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"model-relay/internal/access"
	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/fault"
	"model-relay/internal/routing"
	"model-relay/internal/secure"
	"model-relay/internal/upstream"
)

func ptr[T any](v T) *T { return &v }
func fixture(t *testing.T, endpoint string) (*catalog.Service, *sql.DB, *configstate.Coordinator, *routing.Engine) {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "catalog.db")+"?_foreign_keys=on&_busy_timeout=2000")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(`CREATE TABLE meta(config_revision INTEGER);INSERT INTO meta VALUES(1);
 CREATE TABLE upstream(id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT UNIQUE,url TEXT,token BLOB,enabled INTEGER,revision INTEGER,execution_revision INTEGER);
 CREATE TABLE upstream_checks(id INTEGER PRIMARY KEY AUTOINCREMENT,upstream_id INTEGER REFERENCES upstream(id),execution_revision INTEGER,state TEXT,started_at TEXT,finished_at TEXT,duration_ms INTEGER,error_code TEXT,upstream_status INTEGER);
 CREATE TABLE service_profiles(upstream_id INTEGER PRIMARY KEY REFERENCES upstream(id),source_kind TEXT);
 CREATE TABLE models(alias TEXT PRIMARY KEY,enabled INTEGER,revision INTEGER,legacy_remote_model TEXT);
 CREATE TABLE model_route_policies(alias TEXT PRIMARY KEY REFERENCES models(alias),strategy TEXT);
 CREATE TABLE route_candidates(alias TEXT REFERENCES models(alias),upstream_id INTEGER REFERENCES upstream(id),remote_model TEXT,enabled INTEGER,priority INTEGER,weight INTEGER,execution_revision INTEGER,PRIMARY KEY(alias,upstream_id));
 CREATE TABLE tenants(id TEXT PRIMARY KEY,active INTEGER);INSERT INTO tenants VALUES('active',1),('paused',0);
 CREATE TABLE tenant_model_grants(tenant_id TEXT REFERENCES tenants(id),alias TEXT REFERENCES models(alias),PRIMARY KEY(tenant_id,alias));
 CREATE TABLE console_audit(id TEXT,time TEXT,actor TEXT,tenant_id TEXT,action TEXT,object_id TEXT);`)
	if err != nil {
		t.Fatal(err)
	}
	authority := "example.test"
	if endpoint != "" {
		value, _ := url.Parse(endpoint)
		authority = value.Host
	}
	policy := upstream.Policy{TestAuthority: authority}
	state, router := configstate.New(), routing.New()
	client := upstream.NewClient(upstream.NewHTTPClient(policy, time.Second))
	return catalog.New(db, make([]byte, 32), 3, state, router, client, policy), db, state, router
}
func serviceInput(endpoint string) catalog.UpstreamInput {
	return catalog.UpstreamInput{Name: "supplier", URL: endpoint, Token: "initial", Enabled: ptr(true), SourceKind: "provider"}
}
func modelInput(id int64, remote string) catalog.ModelInput {
	values := []catalog.CandidateInput{{UpstreamID: id, RemoteModel: remote, Enabled: ptr(true), Priority: ptr(1), Weight: ptr(2)}}
	return catalog.ModelInput{Alias: "alpha", Enabled: ptr(true), Candidates: &values}
}
func requireCode(t *testing.T, err error, kind fault.Kind, code string) {
	t.Helper()
	var value *fault.Error
	if !errors.As(err, &value) || value.Kind != kind || value.Code != code {
		t.Fatalf("got %v, want %v/%s", err, kind, code)
	}
}
func createService(t *testing.T, service *catalog.Service, endpoint string) catalog.Upstream {
	t.Helper()
	value, err := service.CreateUpstream(context.Background(), "actor", serviceInput(endpoint))
	if err != nil {
		t.Fatal(err)
	}
	return value
}
func TestUpstreamMutationOwnsRevisionCredentialAndAtomicRuntimePublication(t *testing.T) {
	remote := httptest.NewServer(http.NotFoundHandler())
	defer remote.Close()
	service, db, state, _ := fixture(t, remote.URL)
	ctx := context.Background()
	value := createService(t, service, remote.URL)
	if value.Revision != 1 || value.ExecutionRevision != 1 || !value.CredentialConfigured || value.SourceKind != "provider" {
		t.Fatalf("unexpected creation: %+v", value)
	}
	token, err := secure.Unseal(make([]byte, 32), value.Encrypted, secure.UpstreamAAD(value.ID, value.URL))
	if err != nil || token != "initial" {
		t.Fatalf("credential not sealed by persisted identity: %q %v", token, err)
	}
	input := serviceInput(remote.URL)
	input.Expected = ptr(value.Revision)
	input.Name = "renamed"
	input.Token = ""
	renamed, err := service.UpdateUpstream(ctx, "actor", value.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Revision != 2 || renamed.ExecutionRevision != 1 {
		t.Fatalf("public edit invalidated execution: %+v", renamed)
	}
	_, err = service.UpdateUpstream(ctx, "actor", value.ID, input)
	var conflict *catalog.RevisionConflict
	if !errors.As(err, &conflict) || conflict.CurrentRevision != 2 {
		t.Fatalf("missing current revision: %v", err)
	}
	requireCode(t, err, fault.Conflict, "configuration_conflict")
	state.RLock()
	epoch := state.Epoch()
	state.RUnlock()
	if _, err = db.Exec(`CREATE TRIGGER reject_audit BEFORE INSERT ON console_audit BEGIN SELECT RAISE(ABORT,'controlled audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	input.Expected = ptr(renamed.Revision)
	input.Token = "replacement"
	_, err = service.UpdateUpstream(ctx, "actor", value.ID, input)
	requireCode(t, err, fault.Unavailable, "storage_unavailable")
	current, err := service.ReadUpstream(ctx, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Revision != 2 || current.ExecutionRevision != 1 {
		t.Fatalf("failed transaction partially committed: %+v", current)
	}
	state.RLock()
	if state.Epoch() != epoch {
		t.Error("uncommitted mutation advanced runtime epoch")
	}
	state.RUnlock()
	var revision int
	if err = db.QueryRow("SELECT config_revision FROM meta").Scan(&revision); err != nil || revision != 3 {
		t.Fatalf("persisted configuration drift: %d %v", revision, err)
	}
}
func TestModelReplacementAndGrantValidationKeepExecutionEvidence(t *testing.T) {
	remote := httptest.NewServer(http.NotFoundHandler())
	defer remote.Close()
	service, db, state, router := fixture(t, remote.URL)
	ctx := context.Background()
	source := createService(t, service, remote.URL)
	model, err := service.CreateModel(ctx, "actor", modelInput(source.ID, "one"))
	if err != nil {
		t.Fatal(err)
	}
	candidate := routing.Candidate{UpstreamID: source.ID, RemoteModel: "one", Enabled: true, Priority: 1, Weight: 2, ExecutionRevision: 1, UpstreamExecution: 1}
	router.Observe(routing.Snapshot{Alias: "alpha", Candidates: []routing.Candidate{candidate}}, candidate, routing.Observation{State: "success", Sequence: 1, ObservedAt: time.Now()})
	changed := modelInput(source.ID, "one")
	changed.Alias = ""
	changed.Expected = ptr(model.Revision)
	(*changed.Candidates)[0].Weight = ptr(7)
	model, err = service.UpdateModel(ctx, "actor", "alpha", changed)
	if err != nil {
		t.Fatal(err)
	}
	if model.Candidates[0].ExecutionRevision != 1 || router.Health("alpha", candidate) == nil {
		t.Fatal("non-execution edit erased compatible generation evidence")
	}
	changed.Expected = ptr(model.Revision)
	(*changed.Candidates)[0].RemoteModel = "two"
	model, err = service.UpdateModel(ctx, "actor", "alpha", changed)
	if err != nil {
		t.Fatal(err)
	}
	if model.Candidates[0].ExecutionRevision != 2 || router.Health("alpha", candidate) != nil {
		t.Fatal("remote-model replacement retained old generation evidence")
	}
	principal := access.Principal{ID: "actor", Platform: true}
	if err = service.SetGrants(ctx, principal, "alpha", []string{"active"}); err != nil {
		t.Fatal(err)
	}
	err = service.SetGrants(ctx, principal, "alpha", []string{"active", "paused"})
	requireCode(t, err, fault.Invalid, "invalid_grants")
	grants, err := service.Grants(ctx, principal, "alpha")
	if err != nil || fmt.Sprint(grants) != "[active]" {
		t.Fatalf("invalid replacement changed grants: %v %v", grants, err)
	}
	requireCode(t, service.SetGrants(ctx, access.Principal{ID: "other"}, "alpha", nil), fault.Forbidden, "forbidden")
	state.RLock()
	snapshot, err := service.ModelSnapshot(ctx, "alpha")
	state.RUnlock()
	if err != nil || snapshot.Model.Candidates[0].RemoteModel != "two" || snapshot.Upstreams[source.ID].ExecutionRevision != 1 {
		t.Fatalf("inconsistent snapshot: %+v %v", snapshot, err)
	}
	var audits int
	if err = db.QueryRow("SELECT count(*) FROM console_audit WHERE actor='actor'").Scan(&audits); err != nil || audits != 5 {
		t.Fatalf("audit actor or atomicity changed: %d %v", audits, err)
	}
}
func TestChecksAreExclusiveCancelableAndLateResultsRemainHistorical(t *testing.T) {
	started, release := make(chan struct{}, 2), make(chan struct{}, 2)
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/models" {
			t.Errorf("check generated or used wrong path: %s %s", r.Method, r.URL.Path)
		}
		started <- struct{}{}
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"list","data":[{"id":"one"}]}`)
	}))
	defer remote.Close()
	service, _, state, _ := fixture(t, remote.URL)
	source := createService(t, service, remote.URL)
	done := make(chan catalog.CheckResult, 1)
	errs := make(chan error, 1)
	go func() { value, err := service.Check(context.Background(), source.ID, 1); done <- value; errs <- err }()
	<-started
	_, err := service.Check(context.Background(), source.ID, 1)
	requireCode(t, err, fault.Conflict, "check_in_progress")
	input := serviceInput(remote.URL)
	input.Expected = ptr(source.Revision)
	input.Token = "new"
	if _, err = service.UpdateUpstream(context.Background(), "actor", source.ID, input); err != nil {
		t.Fatal(err)
	}
	release <- struct{}{}
	checked := <-done
	if err = <-errs; err != nil || checked.State != "success" || checked.Applicable {
		t.Fatalf("late check promoted: %+v %v", checked, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { value, err := service.Check(ctx, source.ID, 2); done <- value; errs <- err }()
	<-started
	cancel()
	checked = <-done
	if err = <-errs; err != nil || checked.State != "canceled" || checked.Code != "client_canceled" {
		t.Fatalf("cancellation lost: %+v %v", checked, err)
	}
	state.RLock()
	checking := state.Checking(source.ID)
	state.RUnlock()
	if checking {
		t.Fatal("canceled check retained admission lock")
	}
	current, err := service.ReadUpstream(context.Background(), source.ID)
	if err != nil || current.Check == nil || current.Check.State != "canceled" || !current.Check.Applicable {
		t.Fatalf("final check not persisted: %+v %v", current, err)
	}
}

func TestFailedCheckPersistenceReleasesAdmissionAndDistinguishesDegradedStorage(t *testing.T) {
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"list","data":[]}`)
	}))
	defer remote.Close()
	service, db, state, _ := fixture(t, remote.URL)
	source := createService(t, service, remote.URL)
	if _, err := db.Exec(`CREATE TRIGGER reject_check_result BEFORE UPDATE ON upstream_checks BEGIN SELECT RAISE(ABORT,'controlled result failure'); END`); err != nil {
		t.Fatal(err)
	}
	_, err := service.Check(context.Background(), source.ID, 1)
	var persistence *catalog.CheckPersistenceError
	if !errors.As(err, &persistence) {
		t.Fatalf("cannot distinguish failed final persistence: %v", err)
	}
	requireCode(t, err, fault.Unavailable, "storage_unavailable")
	state.RLock()
	checking := state.Checking(source.ID)
	state.RUnlock()
	if checking {
		t.Fatal("failed final persistence retained registry slot")
	}
	var status string
	if err = db.QueryRow("SELECT state FROM upstream_checks WHERE upstream_id=?", source.ID).Scan(&status); err != nil || status != "checking" {
		t.Fatalf("failed persistence invented final evidence: %s %v", status, err)
	}
}

func TestModelValidationRejectsMissingOrDuplicateCandidateFactsWithoutMutation(t *testing.T) {
	remote := httptest.NewServer(http.NotFoundHandler())
	defer remote.Close()
	service, _, state, _ := fixture(t, remote.URL)
	source := createService(t, service, remote.URL)
	cases := []struct {
		name   string
		change func(*catalog.ModelInput)
		code   string
	}{
		{"missing enabled", func(input *catalog.ModelInput) { input.Enabled = nil }, "invalid_request"},
		{"missing candidates", func(input *catalog.ModelInput) { input.Candidates = nil }, "invalid_request"},
		{"unknown strategy", func(input *catalog.ModelInput) { input.Strategy = ptr("random") }, "invalid_routing_strategy"},
		{"duplicate upstream", func(input *catalog.ModelInput) { *input.Candidates = append(*input.Candidates, (*input.Candidates)[0]) }, "invalid_route_candidate"},
		{"missing enabled candidate", func(input *catalog.ModelInput) { (*input.Candidates)[0].Enabled = nil }, "invalid_route_candidate"},
		{"missing weight", func(input *catalog.ModelInput) { (*input.Candidates)[0].Weight = nil }, "invalid_route_candidate"},
		{"zero weight", func(input *catalog.ModelInput) { (*input.Candidates)[0].Weight = ptr(0) }, "invalid_route_candidate"},
		{"excess priority", func(input *catalog.ModelInput) { (*input.Candidates)[0].Priority = ptr(101) }, "invalid_route_candidate"},
		{"missing upstream", func(input *catalog.ModelInput) { (*input.Candidates)[0].UpstreamID = source.ID + 1 }, "invalid_route_candidate"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := modelInput(source.ID, "remote")
			test.change(&input)
			_, err := service.CreateModel(context.Background(), "actor", input)
			requireCode(t, err, fault.Invalid, test.code)
		})
	}
	state.RLock()
	snapshotEpoch := state.Epoch()
	state.RUnlock()
	if snapshotEpoch != 1 {
		t.Fatalf("invalid candidates changed epoch: %d", snapshotEpoch)
	}
	values, err := service.ListModels(context.Background(), "", 10)
	if err != nil || len(values) != 0 {
		t.Fatalf("invalid candidates created model: %v %v", values, err)
	}
}

func TestCredentialReplacementInvalidatesCheckApplicabilityAndRequiresNewCredentialForAddress(t *testing.T) {
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"list","data":[]}`)
	}))
	defer remote.Close()
	service, _, _, _ := fixture(t, remote.URL)
	ctx := context.Background()
	source := createService(t, service, remote.URL)
	checked, err := service.Check(ctx, source.ID, 1)
	if err != nil || !checked.Applicable {
		t.Fatalf("check did not apply to saved configuration: %+v %v", checked, err)
	}
	input := serviceInput(remote.URL)
	input.Expected = ptr(source.Revision)
	input.Token = "rotated"
	replaced, err := service.UpdateUpstream(ctx, "actor", source.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	if replaced.ExecutionRevision != 2 || replaced.Check == nil || replaced.Check.Applicable {
		t.Fatalf("credential edit kept old check applicable: %+v", replaced)
	}
	input.Expected = ptr(replaced.Revision)
	input.Token = ""
	input.URL = remote.URL + "/different"
	_, err = service.UpdateUpstream(ctx, "actor", source.ID, input)
	requireCode(t, err, fault.Invalid, "new_credential_required")
	current, err := service.ReadUpstream(ctx, source.ID)
	if err != nil || current.URL != source.URL || current.Revision != 2 {
		t.Fatalf("rejected address changed configuration: %+v %v", current, err)
	}
}

func TestSchema2ConfigurationPreservesManagementWithoutConsoleTables(t *testing.T) {
	remote := httptest.NewServer(http.NotFoundHandler())
	defer remote.Close()
	_, db, state, router := fixture(t, remote.URL)
	if _, err := db.Exec("DROP TABLE service_profiles;DROP TABLE console_audit"); err != nil {
		t.Fatal(err)
	}
	endpoint, _ := url.Parse(remote.URL)
	policy := upstream.Policy{TestAuthority: endpoint.Host}
	service := catalog.New(db, make([]byte, 32), 2, state, router, upstream.NewClient(upstream.NewHTTPClient(policy, time.Second)), policy)
	source := createService(t, service, remote.URL)
	if source.SourceKind != "" {
		t.Fatal("schema2 invented source profile")
	}
	if _, err := service.CreateModel(context.Background(), "", modelInput(source.ID, "remote")); err != nil {
		t.Fatal(err)
	}
	sources, err := service.ListUpstreams(context.Background(), 0, 2)
	if err != nil || len(sources) != 1 {
		t.Fatalf("schema2 endpoint list: %v %v", sources, err)
	}
	models, err := service.ListModels(context.Background(), "", 2)
	if err != nil || len(models) != 1 || models[0].Candidates[0].GenerationHealth != "未知（本次启动）" {
		t.Fatalf("schema2 model list: %v %v", models, err)
	}
}

func TestModelAuditFailureRollsBackCandidatesAndPreservesRuntimeHealth(t *testing.T) {
	remote := httptest.NewServer(http.NotFoundHandler())
	defer remote.Close()
	service, db, state, router := fixture(t, remote.URL)
	ctx := context.Background()
	source := createService(t, service, remote.URL)
	model, err := service.CreateModel(ctx, "actor", modelInput(source.ID, "remote"))
	if err != nil {
		t.Fatal(err)
	}
	candidate := routing.Candidate{UpstreamID: source.ID, RemoteModel: "remote", Enabled: true, Priority: 1, Weight: 2, ExecutionRevision: 1, UpstreamExecution: 1}
	router.Observe(routing.Snapshot{Alias: "alpha", Candidates: []routing.Candidate{candidate}}, candidate, routing.Observation{State: "success", Sequence: 1, ObservedAt: time.Now()})
	if _, err = db.Exec(`CREATE TRIGGER reject_model_audit BEFORE INSERT ON console_audit BEGIN SELECT RAISE(ABORT,'controlled audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	input := modelInput(source.ID, "other")
	input.Alias = ""
	input.Expected = ptr(model.Revision)
	_, err = service.UpdateModel(ctx, "actor", "alpha", input)
	requireCode(t, err, fault.Unavailable, "storage_unavailable")
	value, err := service.ReadModel(ctx, "alpha")
	if err != nil || value.Revision != 1 || value.Candidates[0].RemoteModel != "remote" {
		t.Fatalf("failed candidate replacement partially committed: %+v %v", value, err)
	}
	if router.Health("alpha", candidate) == nil {
		t.Fatal("uncommitted model update invalidated health")
	}
	state.RLock()
	epoch := state.Epoch()
	state.RUnlock()
	if epoch != 2 {
		t.Fatalf("uncommitted model update changed epoch: %d", epoch)
	}
}

func TestRejectedRequestRemoteEvidenceUsesFirstCandidateAndStaysUnknownOnReadFailure(t *testing.T) {
	service, db, state, _ := fixture(t, "")
	if _, err := db.Exec(`INSERT INTO upstream VALUES(1,'first','https://one.test',x'00',1,1,1),(2,'second','https://two.test',x'00',1,1,1);
 INSERT INTO models VALUES('alpha',1,1,NULL);
 INSERT INTO route_candidates VALUES('alpha',2,'second-remote',1,0,1,1),('alpha',1,'first-remote',1,0,1,1);`); err != nil {
		t.Fatal(err)
	}
	state.RLock()
	value := service.FirstRemoteModel(context.Background(), "alpha")
	missing := service.FirstRemoteModel(context.Background(), "missing")
	state.RUnlock()
	if value != "first-remote" || missing != "" {
		t.Fatalf("rejected-request evidence changed: %q/%q", value, missing)
	}
	if _, err := db.Exec("DROP TABLE route_candidates"); err != nil {
		t.Fatal(err)
	}
	state.RLock()
	value = service.FirstRemoteModel(context.Background(), "alpha")
	state.RUnlock()
	if value != "" {
		t.Fatal("read failure invented remote evidence")
	}
}

func TestGatewaySnapshotDistinguishesMissingRouteFromStorageFailure(t *testing.T) {
	service, db, state, _ := fixture(t, "")
	state.RLock()
	_, err := service.ModelSnapshot(context.Background(), "absent")
	state.RUnlock()
	if !errors.Is(err, catalog.ErrModelMissing) {
		t.Fatalf("missing route: %v", err)
	}
	if _, err := db.Exec("DROP TABLE models"); err != nil {
		t.Fatal(err)
	}
	state.RLock()
	_, err = service.ModelSnapshot(context.Background(), "absent")
	state.RUnlock()
	if err == nil || errors.Is(err, catalog.ErrModelMissing) {
		t.Fatalf("storage failure became missing route: %v", err)
	}
}
