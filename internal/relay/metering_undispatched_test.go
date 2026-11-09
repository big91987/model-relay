package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"model-relay/internal/access"
	"model-relay/internal/accounting"
	"model-relay/internal/admission"
	"model-relay/internal/apikey"
	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/finance"
	"model-relay/internal/httpconsole"
	"model-relay/internal/identity"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/reporting"
	"model-relay/internal/routing"
	"model-relay/internal/secure"
	"model-relay/internal/sessions"
	"model-relay/internal/upstream"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

type undispatchedTransport struct{ hits *atomic.Int64 }

func (transport undispatchedTransport) RoundTrip(*http.Request) (*http.Response, error) {
	transport.hits.Add(1)
	return nil, errors.New("undispatched request reached upstream")
}

func TestMeteringUndispatchedInProcess(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	directory, keyFile := filepath.Join(root, "data"), filepath.Join(root, "key")
	if _, err := installation.InitConsole(directory, keyFile); err != nil {
		test.Fatal(err)
	}
	store, err := installation.Open(directory, keyFile)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()
	store.SetClock(func() time.Time { return time.Date(2026, 9, 30, 23, 59, 0, 0, time.UTC) })
	if err = store.AdvanceClock(); err != nil {
		test.Fatal(err)
	}
	state, router := configstate.New(), routing.New()
	var hits atomic.Int64
	var degraded atomic.Bool
	client := upstream.NewClient(&http.Client{Transport: undispatchedTransport{&hits}})
	configuration := Config{Authority: "undispatched.test"}
	configuration.Defaults()
	models := catalog.New(store.DB, store.Key, store.Schema, state, router, client, upstream.Policy{})
	keys := apikey.New(store.DB, store.Now, state, router)
	journal := records.New(store.DB, store.Schema)
	gateway := &Gateway{Store: store, Config: &configuration, State: state, Router: router, Upstream: client, Catalog: models, Keys: keys, Records: journal, Admission: admission.New(store.DB, store.Now, store.AdvanceClock), Degraded: &degraded}
	ctx := context.Background()
	principal := access.Principal{ID: "platform-admin", Platform: true}
	people := identity.New(store.DB, store.Now)
	tenant, err := people.CreateTenant(ctx, principal, &identity.TenantInput{Name: "Undispatched metering"})
	if err != nil {
		test.Fatal(err)
	}
	enabled := true
	priority, weight := 0, 1
	source, err := models.CreateUpstream(ctx, principal.ID, catalog.UpstreamInput{Name: "Isolated metering source", URL: "https://8.8.8.8/v1", Token: "controlled-only", Enabled: &enabled})
	if err != nil {
		test.Fatal(err)
	}
	candidates := []catalog.CandidateInput{{UpstreamID: source.ID, RemoteModel: "real", Enabled: &enabled, Priority: &priority, Weight: &weight}}
	if _, err = models.CreateModel(ctx, principal.ID, catalog.ModelInput{Alias: "governed", Enabled: &enabled, Candidates: &candidates}); err != nil {
		test.Fatal(err)
	}
	if err = models.SetGrants(ctx, principal, "governed", []string{tenant.ID}); err != nil {
		test.Fatal(err)
	}
	issued, err := keys.Create(ctx, principal, tenant.ID, apikey.Operation{Key: "key", Path: "/keys"}, &apikey.Input{Name: "Application", Application: "test", Enabled: true, Models: []string{"governed"}})
	if err != nil {
		test.Fatal(err)
	}
	money := finance.New(store.DB, store.Now)
	if _, err = money.PublishPrices(ctx, principal, "governed", "price", "/prices", &finance.PublishPricesInput{Prices: []accounting.PriceSnapshot{{Upstream: source.ID, CostIn: 1000000, CostOut: 1000000, FeeIn: 1000000, FeeOut: 1000000}}}); err != nil {
		test.Fatal(err)
	}
	if _, err = money.DeclareBound(ctx, principal, "governed", &finance.BoundInput{Upstream: source.ID, Input: 100, Output: 100, Evidence: "isolated bound"}); err != nil {
		test.Fatal(err)
	}
	tokens := int64(1000)
	if _, err = money.UpdateLimits(ctx, principal, finance.LimitScope{TenantID: tenant.ID, OwnerType: "tenant", OwnerID: tenant.ID}, &accounting.ResourceLimits{TPM: &tokens, Hard: true}); err != nil {
		test.Fatal(err)
	}
	assertUsage := func(phase string, want accounting.ResourceUsage) {
		test.Helper()
		for _, scope := range []finance.LimitScope{{TenantID: tenant.ID, OwnerType: "tenant", OwnerID: tenant.ID}, {TenantID: tenant.ID, OwnerType: "key", OwnerID: issued.ID}} {
			view, err := money.Limits(ctx, principal, scope)
			if err != nil || view.Usage != want {
				test.Fatalf("%s %s usage=%+v want=%+v error=%v", phase, scope.OwnerType, view.Usage, want, err)
			}
		}
	}
	record := records.Record{ID: secure.Random(16), Time: store.Now().Format(time.RFC3339Nano), KeyID: issued.ID, Alias: "governed"}
	encoded, err := json.Marshal(map[string]any{"model": "governed", "messages": []map[string]string{{"role": "user", "content": "controlled"}}, "max_tokens": 100})
	if err != nil {
		test.Fatal(err)
	}
	payload, _, _, code := upstream.ValidateRequest(encoded)
	if code != "" {
		test.Fatal("undispatched input", code)
	}
	requestContext, cancel := context.WithCancel(ctx)
	defer cancel()
	request := httptest.NewRequest("POST", "http://undispatched.test/v1/chat/completions", bytes.NewReader(encoded)).WithContext(requestContext)
	request.Header.Set("Authorization", "Bearer "+issued.Secret)
	state.Lock()
	snapshot, err := gateway.snapshot("governed")
	state.Unlock()
	if err != nil || journal.Start(record) != nil {
		test.Fatal("undispatched snapshot/start", err)
	}
	if code, status, _ := gateway.Admission.Admit(record, snapshot, payload, issued.Secret, ""); status != 200 || code != "" {
		test.Fatal("undispatched admission", status, code)
	}
	assertUsage("admitted_before_dispatch", accounting.ResourceUsage{Calls: 1, RPM: 1, TPM: 200, Concurrency: 1, Reserved: 200})
	cancel()
	gateway.routedGenerate(httptest.NewRecorder(), request, requestContext, &record, payload, snapshot, func(state, code string) bool {
		record.State, record.Code = state, code
		return journal.Finish(record) == nil
	})
	if record.State != "canceled" || record.Code != "client_canceled" || hits.Load() != 0 {
		test.Fatal("undispatched cancel must not call upstream", record.State, record.Code)
	}
	assertUsage("undispatched_settled", accounting.ResourceUsage{Calls: 1, RPM: 1})
	record.State, record.Code = "success", ""
	if journal.Finish(record) != nil {
		test.Fatal("duplicate/late undispatched finish")
	}
	assertUsage("late_finish_no_second_release", accounting.ResourceUsage{Calls: 1, RPM: 1})
	sessionStore := sessions.New()
	secret, csrf := secure.Random(32), secure.Random(32)
	sessionStore.Create(time.Now(), secret, principal.ID, csrf)
	controller := &httpconsole.Controller{Schema: store.Schema, Authority: func() string { return configuration.Authority }, Password: store.Password, Now: store.Now, AdvanceClock: store.AdvanceClock, Degraded: &degraded, State: state, Sessions: sessionStore, Identity: people, Catalog: models, Keys: keys, Finance: money, Reports: reporting.New(store.DB, store.Now), Records: journal}
	read := httptest.NewRequest("GET", "http://undispatched.test/admin/api/workspace/requests/"+record.ID+"?tenant="+tenant.ID, nil)
	read.AddCookie(&http.Cookie{Name: "mr_session", Value: secret})
	response := httptest.NewRecorder()
	controller.ServeHTTP(response, read)
	status := response.Code
	var detail map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &detail); err != nil {
		test.Fatal(err)
	}
	if status != 200 || detail["state"] != "canceled" || len(detail["attempts"].([]any)) != 0 {
		test.Fatal("late result replaced canceled terminal")
	}
	test.Logf("METERING_UNDISPATCHED request_id=%s calls=1 rpm=1 tpm=0 concurrency=0 fee=0 pending=0 attempts=0 layer=admission_dispatch_contract_not_http", record.ID)
}
