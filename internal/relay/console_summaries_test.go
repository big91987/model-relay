package relay

import (
	"encoding/csv"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestConsoleSummariesInProcess(test *testing.T)  { summariesScenario(test, false) }
func TestConsoleSummariesActualHTTP(test *testing.T) { summariesScenario(test, true) }

func summariesScenario(test *testing.T, actual bool) {
	fixture := governedInstallation(test, actual)
	admin := &fixture.admin
	var developer *consoleClient
	status, catalog := admin.call("GET", "catalog?tenant="+fixture.tenant, nil, "")
	if status != 200 {
		test.Fatal("catalog summary", status)
	}
	model := catalog["items"].([]any)[0].(map[string]any)
	summary := model["summary"].(map[string]any)
	sources := summary["sources"].([]any)
	if summary["average_duration_ms"] != nil || len(sources) != 3 {
		test.Fatal("unmeasured model must not invent latency or omit sources")
	}
	for index, name := range []string{"first", "second", "third"} {
		source := sources[index].(map[string]any)
		if source["name"] != name || source["source_kind"] != "self_hosted" || source["priority"] != float64(3-index) || source["tenant_input_rate"] != float64(1000000) || source["tenant_output_rate"] != float64(1000000) {
			test.Fatal("source or current candidate price summary mismatch", index)
		}
	}
	status, tenants := admin.call("GET", "tenants", nil, "")
	if status != 200 {
		test.Fatal("tenant summary", status)
	}
	first := tenants["items"].([]any)[0].(map[string]any)["summary"].(map[string]any)
	if len(first["owners"].([]any)) != 0 || first["member_count"] != float64(0) || first["key_count"] != float64(1) || !reflect.DeepEqual(first["models"], []any{"governed"}) {
		test.Fatal("caller is not automatically a tenant owner")
	}
	for _, role := range []string{"administrator", "developer", "finance"} {
		account := role + "@summary.example.test"
		status, invitation := admin.call("POST", "tenants/"+fixture.tenant+"/invitations", map[string]any{"account": account, "role": role}, "")
		if status != 201 {
			test.Fatal("summary invitation", status)
		}
		person := consoleClient{test: test, server: fixture.server}
		if status, _ := person.call("POST", "invitations/accept", map[string]any{"token": invitation["acceptance_token"], "password": "controlled-summary-password"}, ""); status != 200 {
			test.Fatal("summary acceptance", status)
		}
		if status, _ := person.call("POST", "login", map[string]any{"account": account, "password": "controlled-summary-password"}, ""); status != 200 {
			test.Fatal("summary login", status)
		}
		status, visible := person.call("GET", "tenants", nil, "")
		if status != 200 || len(visible["items"].([]any)) != 1 {
			test.Fatal("summary tenant isolation", status)
		}
		value := visible["items"].([]any)[0].(map[string]any)
		if role == "administrator" {
			owners := value["summary"].(map[string]any)["owners"]
			if !reflect.DeepEqual(owners, []any{account}) {
				test.Fatal("owner must be an active real administrator")
			}
		} else if _, present := value["summary"]; present {
			test.Fatal("aggregate personnel/key/budget metadata leaked to restricted role")
		}
		if status, _ := person.call("GET", "catalog?tenant="+fixture.tenant+"&include_ungranted=true", nil, ""); status != 403 {
			test.Fatal("nonplatform catalog expansion", status)
		}
		if status, _ := person.call("GET", "catalog?tenant=other", nil, ""); status != 404 {
			test.Fatal("cross tenant summary", status)
		}
		if role == "developer" {
			developer = &person
		}
	}
	status, id := fixture.generate(map[string]any{"model": "governed", "messages": []map[string]string{{"role": "user", "content": "controlled summary accounting"}}}, "summary-call")
	if status != 200 {
		test.Fatal("summary dispatch", status)
	}
	status, before := admin.call("GET", "workspace/requests/"+id, nil, "")
	if status != 200 || before["confirmed_input_tokens"] != "6" || before["confirmed_output_tokens"] != "8" || before["pending"] != float64(1) {
		test.Fatal("all known attempted usage must be counted before reconciliation", status)
	}
	reconciliation := map[string]any{"input_tokens": 3, "output_tokens": 4, "expected_revision": 1, "reason": "controlled summary receipt", "evidence": "test-only authority"}
	for iteration := 0; iteration < 2; iteration++ {
		if status, _ := admin.call("POST", "workspace/requests/"+id+"/reconciliations/2", reconciliation, "summary-reconcile"); status != 200 {
			test.Fatal("summary reconciliation or replay", status)
		}
	}
	status, after := admin.call("GET", "workspace/requests/"+id, nil, "")
	if status != 200 || after["confirmed_input_tokens"] != "9" || after["confirmed_output_tokens"] != "12" || after["pending"] != float64(0) || !reflect.DeepEqual(after["price_snapshots"], before["price_snapshots"]) || !reflect.DeepEqual(after["attempts"], before["attempts"]) || after["total_tokens"] != before["total_tokens"] {
		test.Fatal("confirmed summary must not mutate original attempts/response/prices or double count")
	}
	var fee int64
	for _, entry := range after["charge_entries"].([]any) {
		amount, err := strconv.ParseInt(entry.(map[string]any)["fee"].(string), 10, 64)
		if err != nil {
			test.Fatal("exact charge amount", err)
		}
		fee += amount
	}
	if fee != 21 {
		test.Fatal("reconciliation replay double-charged")
	}
	for _, original := range before["charge_entries"].([]any) {
		found := false
		for _, entry := range after["charge_entries"].([]any) {
			found = found || reflect.DeepEqual(original, entry)
		}
		if !found {
			test.Fatal("original charge entry modified")
		}
	}
	status, records := admin.call("GET", "workspace/requests?tenant="+fixture.tenant+"&request_id="+id, nil, "")
	if status != 200 || len(records["items"].([]any)) != 1 {
		test.Fatal("summary filtered list", status)
	}
	record := records["items"].([]any)[0].(map[string]any)
	if record["confirmed_input_tokens"] != "9" || record["confirmed_output_tokens"] != "12" {
		test.Fatal("confirmed list totals")
	}
	status, usage := admin.call("GET", "usage?tenant="+fixture.tenant+"&request_id="+id, nil, "")
	if status != 200 || usage["items"].([]any)[0].(map[string]any)["confirmed_input_tokens"] != float64(9) {
		test.Fatal("confirmed analysis totals")
	}
	request := httptest.NewRequest("GET", fixture.origin+"/admin/api/workspace/requests/export?tenant="+fixture.tenant+"&request_id="+id, nil)
	request.AddCookie(admin.cookie)
	var response *http.Response
	if actual {
		request.RequestURI = ""
		var err error
		response, err = http.DefaultClient.Do(request)
		if err != nil {
			test.Fatal("summary CSV HTTP", err)
		}
	} else {
		writer := httptest.NewRecorder()
		fixture.server.ServeHTTP(writer, request)
		response = writer.Result()
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	rows, csvErr := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if response.StatusCode != 200 || err != nil || csvErr != nil || len(rows) != 2 {
		test.Fatal("summary CSV", response.StatusCode)
	}
	columns := 0
	for index, name := range rows[0] {
		if name == "confirmed_input_tokens" || name == "confirmed_output_tokens" {
			columns++
		}
		if (name == "confirmed_input_tokens" && rows[1][index] != "9") || (name == "confirmed_output_tokens" && rows[1][index] != "12") {
			test.Fatal("confirmed CSV mismatch")
		}
	}
	if columns != 2 {
		test.Fatal("confirmed CSV columns absent")
	}
	status, catalog = admin.call("GET", "catalog?tenant="+fixture.tenant, nil, "")
	if status != 200 || catalog["items"].([]any)[0].(map[string]any)["summary"].(map[string]any)["average_duration_ms"] == nil {
		test.Fatal("finished latency evidence missing", status)
	}
	status, restricted := developer.call("GET", "catalog?tenant="+fixture.tenant, nil, "")
	if status != 200 || restricted["items"].([]any)[0].(map[string]any)["summary"].(map[string]any)["average_duration_ms"] != nil {
		test.Fatal("unassigned application latency leaked into developer summary", status)
	}
	status, tenants = admin.call("GET", "tenants", nil, "")
	if status != 200 {
		test.Fatal("final summary", status)
	}
	final := tenants["items"].([]any)[0].(map[string]any)["summary"].(map[string]any)
	status, limits := admin.call("GET", "tenants/"+fixture.tenant+"/limits", nil, "")
	if status != 200 || !reflect.DeepEqual(final["usage"], limits["usage"]) || !reflect.DeepEqual(final["limits"], limits["limits"]) || final["member_count"] != float64(3) {
		test.Fatal("summary must match authoritative live limits and memberships")
	}
	service, err := fixture.store.upstreamByID(3)
	if err != nil {
		test.Fatal("current service for invalidation", err)
	}
	if status, _ := admin.call("PATCH", "upstreams/3", map[string]any{"name": service.Name, "url": service.URL, "enabled": true, "source_kind": "self_hosted", "token": "controlled-new-summary-credential", "expected_revision": service.Revision}, ""); status != 200 {
		test.Fatal("service execution edit", status)
	}
	status, catalog = admin.call("GET", "catalog?tenant="+fixture.tenant, nil, "")
	if status != 200 {
		test.Fatal("invalidated catalog summary", status)
	}
	invalidated := catalog["items"].([]any)[0].(map[string]any)["summary"].(map[string]any)["sources"].([]any)[2].(map[string]any)
	if invalidated["tenant_input_rate"] != nil || invalidated["tenant_output_rate"] != nil {
		test.Fatal("prior execution price must be shown as unknown, not current")
	}
}
