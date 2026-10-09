package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConsoleActualHTTPQuotaBillingAndSSE(testContext *testing.T) {
	testContext.Parallel()
	consoleAccountingScenario(testContext, true)
}

func TestConsoleAccountingInProcess(testContext *testing.T) {
	testContext.Parallel()
	consoleAccountingScenario(testContext, false)
}

func consoleAccountingScenario(testContext *testing.T, actualHTTP bool) {
	store, password := consoleInstallation(testContext)
	var hits atomic.Int64
	upstreamHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		hits.Add(1)
		var input struct {
			Stream bool `json:"stream"`
		}
		_ = json.NewDecoder(request.Body).Decode(&input)
		if input.Stream {
			writer.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(writer, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"controlled\"}}]}\n\n")
			writer.(http.Flusher).Flush()
			fmt.Fprint(writer, "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":4,\"total_tokens\":7}}\n\ndata: [DONE]\n\n")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`)
	})
	upstreamURL := "https://controlled.example"
	policy := Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}
	if actualHTTP {
		upstream := localServer(testContext, upstreamHandler)
		upstreamURL = upstream.URL
		policy = Policy{TestAuthority: strings.TrimPrefix(upstream.URL, "http://")}
	}
	var clock atomic.Int64
	clock.Store(time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC).UnixNano())
	server := New(store, Config{Authority: "127.0.0.1:5544", Clock: func() time.Time { return time.Unix(0, clock.Load()) }, Policy: policy})
	serviceURL := "http://127.0.0.1:5544"
	if actualHTTP {
		service := localServer(testContext, server)
		serviceURL = service.URL
	} else {
		server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
			writer := httptest.NewRecorder()
			upstreamHandler.ServeHTTP(writer, request)
			return writer.Result(), nil
		})
	}
	admin := consoleClient{test: testContext, server: server}
	if status, _ := admin.call("POST", "login", map[string]any{"password": password}, ""); status != 200 {
		testContext.Fatal("network login", status)
	}
	status, tenant := admin.call("POST", "tenants", map[string]any{"name": "Controlled quota team"}, "")
	if status != 201 {
		testContext.Fatal("tenant", status)
	}
	tenantID := tenant["id"].(string)
	status, _ = admin.call("POST", "upstreams", map[string]any{"name": "Controlled test only", "url": upstreamURL + "/v1", "token": "controlled-only", "enabled": true}, "")
	if status != 201 {
		testContext.Fatal("upstream", status)
	}
	status, _ = admin.call("POST", "models", map[string]any{"alias": "controlled", "enabled": true, "candidates": []map[string]any{{"upstream_id": 1, "remote_model": "controlled", "enabled": true, "priority": 0, "weight": 1}}}, "")
	if status != 201 {
		testContext.Fatal("model", status)
	}
	for _, step := range []struct {
		method, path string
		body         any
		operation    string
		want         int
	}{
		{"PUT", "models/controlled/tenant-grants", map[string]any{"tenants": []string{tenantID}}, "", 200},
		{"PUT", "models/controlled/bounds", map[string]any{"upstream_id": 1, "max_input": 100, "max_output": 100, "evidence": "controlled-provider-contract-not-supplier"}, "", 200},
		{"POST", "models/controlled/prices", map[string]any{"prices": []map[string]any{{"upstream_id": 1, "cost_input_rate": 1000000, "cost_output_rate": 2000000, "tenant_input_rate": 3000000, "tenant_output_rate": 4000000}}}, "price-1", 201},
		{"PUT", "tenants/" + tenantID + "/limits", ResourceLimits{Calls: ptrInt64(2), TPM: ptrInt64(1000), Concurrency: ptrInt64(1), Budget: ptrInt64(5000), Hard: true}, "", 200},
	} {
		if status, _ := admin.call(step.method, step.path, step.body, step.operation); status != step.want {
			testContext.Fatal("configuration", step.path, status)
		}
	}
	status, key := admin.call("POST", "tenants/"+tenantID+"/keys", KeyInput{Name: "controlled app", Application: "fixture", Enabled: true, Models: []string{"controlled"}}, "key-1")
	if status != 201 {
		testContext.Fatal("key", status)
	}
	generate := func(stream bool, operation string) (int, string) {
		body, _ := json.Marshal(map[string]any{"model": "controlled", "messages": []map[string]string{{"role": "user", "content": "controlled input"}}, "stream": stream})
		request, _ := http.NewRequest("POST", serviceURL+"/v1/chat/completions", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+key["key"].(string))
		request.Header.Set("Idempotency-Key", operation)
		var response *http.Response
		var err error
		if actualHTTP {
			response, err = http.DefaultClient.Do(request)
		} else {
			writer := httptest.NewRecorder()
			server.ServeHTTP(writer, request)
			response = writer.Result()
		}
		if err != nil {
			testContext.Fatal("generation HTTP", err)
		}
		defer response.Body.Close()
		data, _ := io.ReadAll(response.Body)
		if response.StatusCode == 200 && stream && !bytes.Contains(data, []byte("[DONE]")) {
			testContext.Fatal("SSE terminal")
		}
		return response.StatusCode, response.Header.Get("X-Request-ID")
	}
	status, firstID := generate(false, "logical-1")
	if status != 200 {
		testContext.Fatal("plain", status)
	}
	if status, _ := generate(false, "logical-1"); status != 409 {
		testContext.Fatal("double dispatch", status)
	}
	status, streamID := generate(true, "logical-2")
	if status != 200 {
		testContext.Fatal("stream", status)
	}
	if status, _ := generate(false, "logical-3"); status != 429 {
		testContext.Fatal("count rejected", status)
	}
	if hits.Load() != 2 {
		testContext.Fatal("rejected/duplicate upstream dispatch", hits.Load())
	}
	for _, query := range []string{"tenant=" + tenantID, "tenant=" + tenantID + "&request_id=" + firstID} {
		status, usage := admin.call("GET", "workspace/usage?"+query, nil, "")
		if status != 200 || len(usage["items"].([]any)) != 1 {
			testContext.Fatal("scoped usage", status)
		}
		row := usage["items"].([]any)[0].(map[string]any)
		expectedInput, expectedOutput := float64(6), float64(8)
		if strings.Contains(query, "request_id=") {
			expectedInput, expectedOutput = 3, 4
		}
		if row["confirmed_input_tokens"] != expectedInput || row["confirmed_output_tokens"] != expectedOutput {
			testContext.Fatal("confirmed token totals or request filter")
		}
	}
	for _, requestID := range []string{firstID, streamID} {
		status, detail := admin.call("GET", "workspace/requests/"+requestID+"?tenant="+tenantID, nil, "")
		if status != 200 || detail["pending"] != float64(0) {
			testContext.Fatal("known billing", status, detail["pending"])
		}
		entries := detail["charge_entries"].([]any)
		if len(entries) != 1 || entries[0].(map[string]any)["fee"] != "25" || entries[0].(map[string]any)["cost"] != "11" {
			testContext.Fatal("independent fees")
		}
	}
	if status, _ := admin.call("POST", "tenants/"+tenantID+"/bills", map[string]any{"period": "2026-09"}, "close-current"); status != 409 {
		testContext.Fatal("current close accepted", status)
	}
	clock.Store(time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC).UnixNano())
	status, closed := admin.call("POST", "tenants/"+tenantID+"/bills", map[string]any{"period": "2026-09"}, "close-ended")
	if status != 201 || closed["original_fee"] != "50" {
		testContext.Fatal("close ended", status, closed["original_fee"])
	}
	if status, _ := admin.call("POST", "bills/"+closed["id"].(string)+"/corrections", map[string]any{"adjustment_period": "2026-10", "amount_micro_usd": -10, "reason": "controlled adjustment"}, "correction-1"); status != 201 {
		testContext.Fatal("correction", status)
	}
	if status, _ := admin.call("POST", "bills/"+closed["id"].(string)+"/corrections", map[string]any{"adjustment_period": "2026-10", "amount_micro_usd": -10, "reason": "controlled adjustment"}, "correction-1"); status != 200 {
		testContext.Fatal("correction idempotency", status)
	}
	status, bill := admin.call("GET", "bills/"+closed["id"].(string), nil, "")
	if status != 200 || bill["original_fee"] != "50" || len(bill["corrections"].([]any)) != 1 {
		testContext.Fatal("original overwritten")
	}
	status, _ = admin.call("PUT", "tenants/"+tenantID+"/limits", ResourceLimits{Calls: ptrInt64(1), Concurrency: ptrInt64(1), Hard: false, Revision: 1}, "")
	if status != 200 {
		testContext.Fatal("modify does not reset", status)
	}
	var wait sync.WaitGroup
	var accepted atomic.Int64
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if status, _ := generate(false, ""); status == 200 {
				accepted.Add(1)
			}
		}()
	}
	wait.Wait()
	if accepted.Load() != 1 || hits.Load() != 3 {
		testContext.Fatal("atomic shared quota", accepted.Load(), hits.Load())
	}
}

func ptrInt64(value int64) *int64 { return &value }
