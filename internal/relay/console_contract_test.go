package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type governedFixture struct {
	test           *testing.T
	store          *Store
	server         *Server
	admin          consoleClient
	tenant         string
	key            string
	keyID          string
	origin         string
	clock          atomic.Int64
	hits           atomic.Int64
	block          chan struct{}
	entered        chan struct{}
	allowUnbounded bool
}

func governedInstallation(test *testing.T, actual bool) *governedFixture {
	test.Helper()
	store, password := consoleInstallation(test)
	fixture := &governedFixture{test: test, store: store, origin: "http://127.0.0.1:5544"}
	fixture.clock.Store(time.Date(2026, 9, 30, 23, 59, 0, 0, time.UTC).UnixNano())
	upstream := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fixture.hits.Add(1)
		var input map[string]any
		if json.NewDecoder(request.Body).Decode(&input) != nil {
			test.Error("controlled upstream invalid payload")
			writer.WriteHeader(400)
			return
		}
		if input["max_tokens"] != float64(100) && !(fixture.allowUnbounded && input["max_tokens"] == nil) {
			test.Error("bounded output not sent")
		}
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(request.URL.Path, "/first/"):
			writer.WriteHeader(500)
			fmt.Fprint(writer, `{"error":{"code":"controlled"},"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`)
		case strings.Contains(request.URL.Path, "/second/"):
			writer.WriteHeader(500)
			fmt.Fprint(writer, `{"error":{"code":"controlled_unknown"}}`)
		default:
			if fixture.block != nil {
				select {
				case fixture.entered <- struct{}{}:
				default:
				}
				select {
				case <-fixture.block:
				case <-request.Context().Done():
					return
				}
			}
			fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled third"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`)
		}
	})
	upstreamURL := "https://controlled.example"
	policy := Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}
	if actual {
		upstreamService := localServer(test, upstream)
		upstreamURL = upstreamService.URL
		policy = Policy{TestAuthority: strings.TrimPrefix(upstreamURL, "http://")}
	}
	fixture.server = New(store, Config{Authority: "127.0.0.1:5544", Clock: func() time.Time { return time.Unix(0, fixture.clock.Load()) }, Policy: policy})
	if actual {
		fixture.origin = localServer(test, fixture.server).URL
	} else {
		fixture.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
			writer := httptest.NewRecorder()
			upstream.ServeHTTP(writer, request)
			return writer.Result(), nil
		})
	}
	fixture.admin = consoleClient{test: test, server: fixture.server}
	if status, _ := fixture.admin.call("POST", "login", map[string]any{"password": password}, ""); status != 200 {
		test.Fatal("governed login", status)
	}
	status, tenant := fixture.admin.call("POST", "tenants", map[string]any{"name": "Governed team"}, "")
	if status != 201 {
		test.Fatal("tenant", status)
	}
	fixture.tenant = tenant["id"].(string)
	candidates := []map[string]any{}
	prices := []map[string]any{}
	for index, name := range []string{"first", "second", "third"} {
		if status, _ := fixture.admin.call("POST", "upstreams", map[string]any{"name": name, "url": upstreamURL + "/" + name + "/v1", "token": "controlled-only", "enabled": true, "source_kind": "self_hosted"}, ""); status != 201 {
			test.Fatal("controlled source", status)
		}
		candidates = append(candidates, map[string]any{"upstream_id": index + 1, "remote_model": "controlled", "enabled": true, "priority": 3 - index, "weight": 1})
		prices = append(prices, map[string]any{"upstream_id": index + 1, "cost_input_rate": (index + 1) * 1000000, "cost_output_rate": (index + 1) * 2000000, "tenant_input_rate": 1000000, "tenant_output_rate": 1000000})
	}
	if status, _ := fixture.admin.call("POST", "models", map[string]any{"alias": "governed", "enabled": true, "candidates": candidates}, ""); status != 201 {
		test.Fatal("model", status)
	}
	if status, _ := fixture.admin.call("PUT", "models/governed/tenant-grants", map[string]any{"tenants": []string{fixture.tenant}}, ""); status != 200 {
		test.Fatal("grant", status)
	}
	for index := 1; index <= 3; index++ {
		if status, _ := fixture.admin.call("PUT", "models/governed/bounds", map[string]any{"upstream_id": index, "max_input": 100, "max_output": 100, "evidence": "controlled contractual cap, not supplier validation"}, ""); status != 200 {
			test.Fatal("bound", status)
		}
	}
	if status, _ := fixture.admin.call("POST", "models/governed/prices", map[string]any{"prices": prices}, "initial-prices"); status != 201 {
		test.Fatal("price", status)
	}
	if status, _ := fixture.admin.call("PUT", "tenants/"+fixture.tenant+"/limits", ResourceLimits{TPM: ptrInt64(10000), Budget: ptrInt64(10000), Hard: true}, ""); status != 200 {
		test.Fatal("limits", status)
	}
	status, key := fixture.admin.call("POST", "tenants/"+fixture.tenant+"/keys", KeyInput{Name: "governed app", Application: "tests", Enabled: true, Models: []string{"governed"}}, "initial-key")
	if status != 201 {
		test.Fatal("issue", status)
	}
	fixture.key = key["key"].(string)
	fixture.keyID = key["id"].(string)
	return fixture
}

func (fixture *governedFixture) generate(body map[string]any, operation string) (int, string) {
	fixture.test.Helper()
	data, _ := json.Marshal(body)
	request, _ := http.NewRequest("POST", fixture.origin+"/v1/chat/completions", bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+fixture.key)
	request.Header.Set("Idempotency-Key", operation)
	var response *http.Response
	if fixture.server.Config.Authority == "127.0.0.1:5544" {
		writer := httptest.NewRecorder()
		fixture.server.ServeHTTP(writer, request)
		response = writer.Result()
	} else {
		var err error
		response, err = http.DefaultClient.Do(request)
		if err != nil {
			fixture.test.Fatal("controlled generation HTTP", err)
		}
	}
	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)
	return response.StatusCode, response.Header.Get("X-Request-ID")
}

func TestConsoleGovernanceActualHTTP(test *testing.T) { governanceScenario(test, true) }
func TestConsoleGovernanceInProcess(test *testing.T)  { governanceScenario(test, false) }

func governanceScenario(test *testing.T, actual bool) {
	fixture := governedInstallation(test, actual)
	admin := &fixture.admin
	body := map[string]any{"model": "governed", "messages": []map[string]string{{"role": "user", "content": "controlled input"}}}
	for _, field := range []string{"n", "max_tokens", "tools"} {
		invalid := map[string]any{"model": "governed", "messages": body["messages"]}
		switch field {
		case "n":
			invalid[field] = 2
		case "max_tokens":
			invalid[field] = 101
		default:
			invalid[field] = []any{}
		}
		if status, _ := fixture.generate(invalid, ""); status != 400 {
			test.Fatal("unsupported bounded request", field, status)
		}
	}
	if fixture.hits.Load() != 0 {
		test.Fatal("invalid dispatch")
	}
	status, id := fixture.generate(body, "three-attempts")
	if status != 200 || fixture.hits.Load() != 3 {
		test.Fatal("three attempt response", status, fixture.hits.Load())
	}
	if status, repeated := fixture.generate(body, "three-attempts"); status != 409 || repeated != id || fixture.hits.Load() != 3 {
		test.Fatal("logical replay identity", status, repeated == id)
	}
	status, detail := admin.call("GET", "workspace/requests/"+id, nil, "")
	if status != 200 || detail["pending"] != float64(1) || detail["fee_reserved"] != "200" || len(detail["attempts"].([]any)) != 3 {
		test.Fatal("unknown attempt reserve", status, detail["pending"], detail["fee_reserved"])
	}
	entries := detail["charge_entries"].([]any)
	priceSnapshots := detail["price_snapshots"]
	unknownEntry := entries[1].(map[string]any)
	if unknownEntry["ordinal"] != float64(2) || unknownEntry["revision"] != float64(0) || unknownEntry["fee"] != "0" || unknownEntry["cost"] != "0" {
		test.Fatal("unknown initial ledger revision", unknownEntry["ordinal"], unknownEntry["revision"], unknownEntry["fee"], unknownEntry["cost"])
	}
	var fee, cost int64
	for _, entry := range entries {
		item := entry.(map[string]any)
		var amount int64
		fmt.Sscan(item["fee"].(string), &amount)
		fee += amount
		fmt.Sscan(item["cost"].(string), &amount)
		cost += amount
	}
	if fee != 14 || cost != 44 {
		test.Fatal("all dispatched known attempts", fee, cost)
	}
	if status, _ := admin.call("POST", "tenants/"+fixture.tenant+"/bills", map[string]any{"period": "2026-09"}, "close-current"); status != 409 {
		test.Fatal("current close")
	}
	fixture.clock.Store(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).UnixNano())
	if status, _ := admin.call("POST", "tenants/"+fixture.tenant+"/bills", map[string]any{"period": "2026-09"}, "close-pending"); status != 409 {
		test.Fatal("pending close")
	}
	status, limits := admin.call("GET", "tenants/"+fixture.tenant+"/limits", nil, "")
	usage := limits["usage"].(map[string]any)
	if status != 200 || usage["reserved_micro_usd"] != "200" || usage["calls"] != float64(0) {
		test.Fatal("cross month pending carry", status)
	}
	fixture.clock.Store(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).UnixNano())
	status, limits = admin.call("GET", "tenants/"+fixture.tenant+"/limits", nil, "")
	if status != 200 || limits["period"] != "2026-09" || limits["usage"].(map[string]any)["calls"] != float64(1) {
		test.Fatal("clock rollback conserves", status)
	}
	fixture.clock.Store(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).UnixNano())
	changed := []map[string]any{}
	for index := 1; index <= 3; index++ {
		changed = append(changed, map[string]any{"upstream_id": index, "cost_input_rate": 9000000, "cost_output_rate": 9000000, "tenant_input_rate": 9000000, "tenant_output_rate": 9000000})
	}
	if status, _ := admin.call("POST", "models/governed/prices", map[string]any{"prices": changed[:1]}, "partial-price"); status != 400 {
		test.Fatal("incomplete model price accepted", status)
	}
	if status, _ := admin.call("POST", "models/governed/prices", map[string]any{"prices": changed}, "changed-price"); status != 201 {
		test.Fatal("change price", status)
	}
	reconciliation := map[string]any{"input_tokens": 2, "output_tokens": 3, "reason": "controlled receipt", "evidence": "test-only-authoritative-receipt", "expected_revision": 1}
	path := "workspace/requests/" + id + "/reconciliations/2"
	if status, _ := admin.call("POST", path, reconciliation, "reconcile-one"); status != 200 {
		test.Fatal("reconcile", status)
	}
	if status, _ := admin.call("POST", path, reconciliation, "reconcile-one"); status != 200 {
		test.Fatal("reconcile replay", status)
	}
	if status, _ := admin.call("POST", path, reconciliation, "reconcile-stale"); status != 409 {
		test.Fatal("stale revision accepted", status)
	}
	reconciliation["expected_revision"] = 2
	reconciliation["input_tokens"] = 1
	if status, _ := admin.call("POST", path, reconciliation, "decrease"); status != 409 {
		test.Fatal("usage decreased", status)
	}
	status, detail = admin.call("GET", "workspace/requests/"+id, nil, "")
	if status != 200 || detail["pending"] != float64(0) || detail["fee_reserved"] != "0" || len(detail["charge_entries"].([]any)) != 4 {
		test.Fatal("settled append-only", status, detail["pending"])
	}
	settledEntries := detail["charge_entries"].([]any)
	for index, previous := range entries {
		settledIndex := index
		if index == 2 {
			settledIndex = 3
		}
		if !reflect.DeepEqual(previous, settledEntries[settledIndex]) {
			test.Fatal("reconciliation mutated prior ledger revision", index)
		}
	}
	if !reflect.DeepEqual(priceSnapshots, detail["price_snapshots"]) {
		test.Fatal("reconciliation mutated request price snapshots")
	}
	delta := settledEntries[2].(map[string]any)
	if delta["ordinal"] != float64(2) || delta["revision"] != float64(2) || delta["fee"] != "5" || delta["cost"] != "16" || delta["input_tokens"] != float64(2) || delta["output_tokens"] != float64(3) {
		test.Fatal("reconciliation append delta", delta["ordinal"], delta["revision"], delta["fee"], delta["cost"])
	}
	status, bill := admin.call("POST", "tenants/"+fixture.tenant+"/bills", map[string]any{"period": "2026-09"}, "closed")
	if status != 201 || bill["original_fee"] != "19" || bill["original_cost"] != "60" {
		test.Fatal("snapshot financial total", status, bill["original_fee"], bill["original_cost"])
	}
	status, repeated := admin.call("POST", "tenants/"+fixture.tenant+"/bills", map[string]any{"period": "2026-09"}, "closed-other-operation")
	if status != 200 || repeated["id"] != bill["id"] {
		test.Fatal("close duplicate", status)
	}
	if status, _ := admin.call("POST", "bills/"+bill["id"].(string)+"/corrections", map[string]any{"amount_micro_usd": -1, "reason": "controlled credit", "adjustment_period": "2026-10"}, "credit"); status != 201 {
		test.Fatal("credit", status)
	}
	status, periods := admin.call("GET", "tenants/"+fixture.tenant+"/bills", nil, "")
	if status != 200 || len(periods["items"].([]any)) != 2 {
		test.Fatal("adjustment-only period missing", status)
	}
	first := periods["items"].([]any)[0].(map[string]any)
	if first["fee_micro_usd"] != "-1" || first["period"] != "2026-10" {
		test.Fatal("period adjustment counted once")
	}
	status, trace := admin.call("GET", "bills/"+bill["id"].(string), nil, "")
	if status != 200 || trace["original_fee"] != "19" || trace["current_fee_micro_usd"] != "18" {
		test.Fatal("original/current split", status)
	}
}

func TestConsoleInvitationRateLimit(test *testing.T) {
	store, _ := consoleInstallation(test)
	server := New(store, Config{Authority: "127.0.0.1:5544"})
	client := consoleClient{test: test, server: server}
	for index := 0; index < 11; index++ {
		status, _ := client.call("POST", "invitations/accept", map[string]any{"token": strings.Repeat("x", 43), "password": "controlled-password"}, "")
		want := 400
		if index == 10 {
			want = 429
		}
		if status != want {
			test.Fatal("acceptance rate boundary", index, status)
		}
	}
}

func TestConsoleGenerationExactDeadline(test *testing.T) {
	fixture := governedInstallation(test, false)
	start := time.Unix(0, fixture.clock.Load())
	old := fixture.key
	path := "tenants/" + fixture.tenant + "/keys/" + fixture.keyID + "/rotations"
	status, rotation := fixture.admin.call("POST", path, map[string]any{"expected_revision": 1, "transition_hours": 24}, "grace")
	if status != 200 {
		test.Fatal("24 hour grace", status)
	}
	deadline := start.Add(24 * time.Hour)
	fixture.clock.Store(deadline.Add(-time.Nanosecond).UnixNano())
	if _, valid := generationIdentity(fixture.store.DB, old, fixture.store.now()); !valid {
		test.Fatal("early expiry")
	}
	if status, _ := fixture.admin.call("POST", path, map[string]any{"expected_revision": 2, "transition_hours": 1}, "before-deadline"); status != 409 {
		test.Fatal("nanosecond grace overlap", status)
	}
	fixture.clock.Store(deadline.UnixNano())
	if _, valid := generationIdentity(fixture.store.DB, old, fixture.store.now()); valid {
		test.Fatal("deadline inclusive auth")
	}
	if status, _ := fixture.admin.call("POST", path, map[string]any{"expected_revision": 2, "transition_hours": 1}, "after-deadline"); status != 200 {
		test.Fatal("following rotation", status)
	}
	if _, valid := generationIdentity(fixture.store.DB, old, fixture.store.now()); valid {
		test.Fatal("expired generation resurrected")
	}
	if _, valid := generationIdentity(fixture.store.DB, rotation["key"].(string), fixture.store.now()); !valid {
		test.Fatal("new previous generation grace lost")
	}
}

func TestConsoleResourceDimensionsInProcess(test *testing.T)  { resourceDimensionScenario(test, false) }
func TestConsoleResourceDimensionsActualHTTP(test *testing.T) { resourceDimensionScenario(test, true) }

func resourceDimensionScenario(test *testing.T, actual bool) {
	fixture := governedInstallation(test, actual)
	body := map[string]any{"model": "governed", "messages": []map[string]string{{"role": "user", "content": "controlled quota input"}}}
	limits := []ResourceLimits{{Calls: ptrInt64(0)}, {RPM: ptrInt64(0)}, {TPM: ptrInt64(0)}, {Concurrency: ptrInt64(0)}, {Budget: ptrInt64(0), Hard: true}}
	for index, limit := range limits {
		limit.Revision = int64(index + 1)
		if status, _ := fixture.admin.call("PUT", "tenants/"+fixture.tenant+"/limits", limit, ""); status != 200 {
			test.Fatal("independent limit update", index, status)
		}
		if status, _ := fixture.generate(body, ""); status != 429 {
			test.Fatal("independent zero limit", index, status)
		}
	}
	if fixture.hits.Load() != 0 {
		test.Fatal("zero quota dispatch")
	}
	if status, _ := fixture.admin.call("PUT", "tenants/"+fixture.tenant+"/limits", ResourceLimits{Revision: 6, Hard: false}, ""); status != 200 {
		test.Fatal("unlimited tenant", status)
	}
	if status, _ := fixture.admin.call("PUT", "tenants/"+fixture.tenant+"/keys/"+fixture.keyID+"/limits", ResourceLimits{Calls: ptrInt64(0), Hard: false, Revision: 1}, ""); status != 200 {
		test.Fatal("key limit", status)
	}
	if status, _ := fixture.generate(body, ""); status != 429 {
		test.Fatal("key independent limit", status)
	}
	status, usage := fixture.admin.call("GET", "tenants/"+fixture.tenant+"/limits", nil, "")
	if status != 200 || usage["usage"].(map[string]any)["calls"] != float64(0) {
		test.Fatal("partial tenant debit on key rejection")
	}
	if status, _ := fixture.admin.call("PUT", "tenants/"+fixture.tenant+"/keys/"+fixture.keyID+"/limits", ResourceLimits{Revision: 2, Hard: false}, ""); status != 200 {
		test.Fatal("restore key limit", status)
	}
	if status, _ := fixture.admin.call("PUT", "tenants/"+fixture.tenant+"/limits", ResourceLimits{Revision: 7, Concurrency: ptrInt64(1), Budget: ptrInt64(600), TPM: ptrInt64(600), Hard: true}, ""); status != 200 {
		test.Fatal("exact all-attempt reservation", status)
	}
	fixture.block = make(chan struct{})
	fixture.entered = make(chan struct{}, 1)
	defer func() {
		select {
		case <-fixture.block:
		default:
			close(fixture.block)
		}
	}()
	finished := make(chan int, 1)
	go func() { status, _ := fixture.generate(body, "active-call"); finished <- status }()
	select {
	case <-fixture.entered:
	case <-time.After(5 * time.Second):
		test.Fatal("active controlled dispatch missing")
	}
	status, usage = fixture.admin.call("GET", "tenants/"+fixture.tenant+"/limits", nil, "")
	current := usage["usage"].(map[string]any)
	if status != 200 || current["concurrency"] != float64(1) || current["reserved_micro_usd"] != "600" || current["calls"] != float64(1) {
		test.Fatal("atomic admission visibility", status)
	}
	for index := 0; index < 4; index++ {
		if status, _ := fixture.generate(body, ""); status != 429 {
			test.Fatal("active resource race", status)
		}
	}
	if fixture.hits.Load() != 3 {
		test.Fatal("competition dispatched", fixture.hits.Load())
	}
	close(fixture.block)
	select {
	case status := <-finished:
		if status != 200 {
			test.Fatal("active completion", status)
		}
	case <-time.After(5 * time.Second):
		test.Fatal("active completion missing")
	}
	status, usage = fixture.admin.call("GET", "tenants/"+fixture.tenant+"/limits", nil, "")
	current = usage["usage"].(map[string]any)
	if status != 200 || current["concurrency"] != float64(0) || current["calls"] != float64(1) || current["reserved_micro_usd"] != "200" {
		test.Fatal("release conserves unknown and counts", status)
	}
}

func TestConsolePeopleScopeInProcess(test *testing.T)  { peopleScopeScenario(test, false) }
func TestConsolePeopleScopeActualHTTP(test *testing.T) { peopleScopeScenario(test, true) }

func TestConsolePauseUsageInProcess(test *testing.T)  { pauseUsageScenario(test, false) }
func TestConsolePauseUsageActualHTTP(test *testing.T) { pauseUsageScenario(test, true) }

func pauseUsageScenario(test *testing.T, actual bool) {
	fixture := governedInstallation(test, actual)
	admin := &fixture.admin
	status, other := admin.call("POST", "tenants", map[string]any{"name": "Pause isolation tenant"}, "")
	if status != 201 {
		test.Fatal("other tenant", status)
	}
	otherID := other["id"].(string)
	if status, _ := admin.call("PUT", "models/governed/tenant-grants", map[string]any{"tenants": []string{fixture.tenant, otherID}}, ""); status != 200 {
		test.Fatal("two tenant grants", status)
	}
	if status, _ := admin.call("PUT", "tenants/"+otherID+"/limits", ResourceLimits{Budget: ptrInt64(10000), Hard: true}, ""); status != 200 {
		test.Fatal("other tenant budget", status)
	}
	status, otherKey := admin.call("POST", "tenants/"+otherID+"/keys", KeyInput{Name: "other app", Application: "pause isolation", Enabled: true, Models: []string{"governed"}}, "pause-other-key")
	if status != 201 {
		test.Fatal("other tenant key", status)
	}
	status, standby := admin.call("POST", "tenants/"+fixture.tenant+"/keys", KeyInput{Name: "standby app", Application: "pause conservation", Enabled: true, Models: []string{"governed"}}, "pause-standby-key")
	if status != 201 {
		test.Fatal("standby key", status)
	}
	primary := fixture.key
	body := map[string]any{"model": "governed", "messages": []map[string]string{{"role": "user", "content": "controlled"}}, "stream": false}
	generate := func(key string, stream bool, expected int) {
		test.Helper()
		fixture.key = key
		body["stream"] = stream
		if status, _ := fixture.generate(body, ""); status != expected {
			test.Fatal("pause conservation generation", status, "want", expected)
		}
	}
	usage := func(tenant string) ResourceUsage {
		test.Helper()
		status, result := admin.call("GET", "tenants/"+tenant+"/limits", nil, "")
		if status != 200 {
			test.Fatal("authoritative usage", status)
		}
		data, err := json.Marshal(result["usage"])
		if err != nil {
			test.Fatal("usage encoding", err)
		}
		var value ResourceUsage
		if err := json.Unmarshal(data, &value); err != nil {
			test.Fatal("usage decoding", err)
		}
		return value
	}
	generate(primary, false, 200)
	generate(standby["key"].(string), false, 200)
	before := usage(fixture.tenant)
	if before.Fee <= 0 || before.Calls != 2 {
		test.Fatal("nonzero established usage required", before)
	}
	if status, _ := admin.call("PATCH", "tenants/"+fixture.tenant, map[string]any{"active": false}, ""); status != 200 {
		test.Fatal("pause", status)
	}
	hits := fixture.hits.Load()
	generate(primary, false, 403)
	generate(primary, true, 403)
	generate(standby["key"].(string), false, 403)
	if fixture.hits.Load() != hits {
		test.Fatal("paused keys dispatched")
	}
	generate(otherKey["key"].(string), false, 200)
	if usage(otherID).Fee <= 0 {
		test.Fatal("other tenant must retain independent billing")
	}
	if paused := usage(fixture.tenant); paused != before {
		test.Fatal("pause or other tenant call changed usage", before, paused)
	}
	if status, _ := admin.call("PATCH", "tenants/"+fixture.tenant, map[string]any{"active": true}, ""); status != 200 {
		test.Fatal("resume", status)
	}
	if resumed := usage(fixture.tenant); resumed != before {
		test.Fatal("resume changed usage", before, resumed)
	}
	generate(primary, false, 200)
	generate(standby["key"].(string), false, 200)
	if after := usage(fixture.tenant); after.Calls != before.Calls+2 || after.Fee <= before.Fee {
		test.Fatal("resumed calls must append usage", before, after)
	}
}

func peopleScopeScenario(test *testing.T, actual bool) {
	fixture := governedInstallation(test, actual)
	admin := &fixture.admin
	status, other := admin.call("POST", "tenants", map[string]any{"name": "Other isolated tenant"}, "")
	if status != 201 {
		test.Fatal("other tenant", status)
	}
	people := map[string]*consoleClient{}
	personIDs := map[string]string{}
	for _, role := range []string{"administrator", "developer", "finance"} {
		account := role + "@controlled.example"
		status, invitation := admin.call("POST", "tenants/"+fixture.tenant+"/invitations", map[string]any{"account": account, "role": role}, "")
		if status != 201 {
			test.Fatal("invite role", role, status)
		}
		person := &consoleClient{test: test, server: fixture.server}
		if status, _ := person.call("POST", "invitations/accept", map[string]any{"token": invitation["acceptance_token"], "password": "controlled-password-long"}, ""); status != 200 {
			test.Fatal("accept role", role, status)
		}
		if status, _ := person.call("POST", "login", map[string]any{"account": account, "password": "controlled-password-long"}, ""); status != 200 {
			test.Fatal("role login", role, status)
		}
		status, identity := person.call("GET", "identity", nil, "")
		if status != 200 {
			test.Fatal("role identity")
		}
		personIDs[role] = identity["id"].(string)
		people[role] = person
		for _, suffix := range []string{"keys", "limits", "members", "bills"} {
			if status, _ := person.call("GET", "tenants/"+other["id"].(string)+"/"+suffix, nil, ""); status != 404 {
				test.Fatal("cross tenant scope", role, suffix, status)
			}
		}
	}
	keyInput := KeyInput{Name: "assigned app", Application: "test", Models: []string{"governed"}, Enabled: true, Diagnostics: []string{personIDs["developer"]}, Revision: 1}
	if status, _ := admin.call("PATCH", "tenants/"+fixture.tenant+"/keys/"+fixture.keyID, keyInput, "assign-key"); status != 200 {
		test.Fatal("diagnostic assignment", status)
	}
	developer := people["developer"]
	finance := people["finance"]
	if status, _ := developer.call("GET", "tenants/"+fixture.tenant+"/keys/"+fixture.keyID, nil, ""); status != 200 {
		test.Fatal("assigned key read", status)
	}
	for _, path := range []string{"tenants/" + fixture.tenant + "/limits", "tenants/" + fixture.tenant + "/bills", "audit?tenant=" + fixture.tenant, "upstreams"} {
		if status, _ := developer.call("GET", path, nil, ""); status != 403 {
			test.Fatal("developer aggregate forbidden", path, status)
		}
	}
	if status, _ := finance.call("GET", "tenants/"+fixture.tenant+"/bills", nil, ""); status != 200 {
		test.Fatal("finance reads bills", status)
	}
	if status, _ := finance.call("POST", "tenants/"+fixture.tenant+"/keys", keyInput, "finance-issue"); status != 403 {
		test.Fatal("finance issue forbidden", status)
	}
	if status, _ := admin.call("PATCH", "tenants/"+fixture.tenant+"/members/"+personIDs["developer"], map[string]any{"role": "finance"}, ""); status != 200 {
		test.Fatal("role change", status)
	}
	status, roles := developer.call("GET", "tenants", nil, "")
	if status != 200 || len(roles["items"].([]any)) != 1 || roles["items"].([]any)[0].(map[string]any)["role"] != "finance" {
		test.Fatal("live session role readback", status)
	}
	status, changedKey := admin.call("GET", "tenants/"+fixture.tenant+"/keys/"+fixture.keyID, nil, "")
	if status != 200 || len(changedKey["diagnostic_members"].([]any)) != 0 {
		test.Fatal("role change diagnostic cleanup", status)
	}
	if status, _ := admin.call("PATCH", "tenants/"+fixture.tenant+"/members/"+personIDs["developer"], map[string]any{"role": "developer"}, ""); status != 200 {
		test.Fatal("role restoration", status)
	}
	keyInput.Revision = 2
	if status, _ := admin.call("PATCH", "tenants/"+fixture.tenant+"/keys/"+fixture.keyID, keyInput, "reassign-key"); status != 200 {
		test.Fatal("diagnostic reassignment before removal", status)
	}
	if status, _ := admin.call("DELETE", "tenants/"+fixture.tenant+"/members/"+personIDs["developer"], nil, ""); status != 200 {
		test.Fatal("remove person", status)
	}
	if status, _ := developer.call("GET", "tenants/"+fixture.tenant+"/keys/"+fixture.keyID, nil, ""); status != 404 {
		test.Fatal("removed session retained scope", status)
	}
	status, remaining := developer.call("GET", "tenants", nil, "")
	if status != 200 || len(remaining["items"].([]any)) != 0 {
		test.Fatal("removed session retained membership", status)
	}
	var assignments int
	if fixture.store.DB.QueryRow("SELECT count(*) FROM key_diagnostic_members WHERE key_id=?", fixture.keyID).Scan(&assignments) != nil || assignments != 0 {
		test.Fatal("diagnostic cleanup")
	}
	if status, _ := admin.call("PATCH", "tenants/"+fixture.tenant, map[string]any{"active": false}, ""); status != 200 {
		test.Fatal("pause", status)
	}
	body := map[string]any{"model": "governed", "messages": []map[string]string{{"role": "user", "content": "controlled"}}, "stream": true}
	if status, _ := fixture.generate(body, ""); status != 403 {
		test.Fatal("paused SSE", status)
	}
	body["stream"] = false
	if status, _ := fixture.generate(body, ""); status != 403 {
		test.Fatal("paused ordinary", status)
	}
	if status, _ := admin.call("PATCH", "tenants/"+fixture.tenant, map[string]any{"active": true}, ""); status != 200 {
		test.Fatal("resume", status)
	}
	if status, _ := fixture.generate(body, ""); status != 200 {
		test.Fatal("person exit revoked application", status)
	}
	if status, _ := people["administrator"].call("DELETE", "tenants/"+fixture.tenant+"/members/"+personIDs["administrator"], nil, ""); status != 409 {
		test.Fatal("last admin protection", status)
	}
}
