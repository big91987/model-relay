package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"model-relay/internal/relay"
)

func TestContinuousUpstreamResponseContract(test *testing.T) {
	var dispatched atomic.Int64
	handler := continuousUpstream(&dispatched)
	for _, scenario := range []struct {
		name, message string
		known         bool
	}{
		{"known", "controlled journey input", true},
		{"unknown", "controlled unknown usage", false},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			body, _ := json.Marshal(map[string]any{"messages": []map[string]string{{"role": "user", "content": scenario.message}}})
			request := httptest.NewRequest("POST", "http://controlled.test/v1/chat/completions", bytes.NewReader(body))
			request.Header.Set("Authorization", "Bearer controlled-journey-upstream")
			writer := httptest.NewRecorder()
			handler.ServeHTTP(writer, request)
			if writer.Code != 200 || writer.Header().Get("Content-Type") != "application/json" {
				test.Fatal("controlled response HTTP contract mismatch")
			}
			var response struct {
				Choices []struct {
					Message struct {
						Role, Content string
					}
				}
				Usage *struct {
					Input  int64 `json:"prompt_tokens"`
					Output int64 `json:"completion_tokens"`
					Total  int64 `json:"total_tokens"`
				}
			}
			if json.Unmarshal(writer.Body.Bytes(), &response) != nil || len(response.Choices) != 1 {
				test.Fatal("controlled completion shape mismatch")
			}
			if response.Choices[0].Message.Role != "assistant" || response.Choices[0].Message.Content == "" {
				test.Fatal("controlled completion requires assistant role and content")
			}
			if (response.Usage != nil) != scenario.known {
				test.Fatal("known/unknown usage contract mismatch")
			}
			if scenario.known && (response.Usage.Input != 3 || response.Usage.Output != 4 || response.Usage.Total != 7) {
				test.Fatal("known usage amounts changed")
			}
		})
	}
	if dispatched.Load() != 2 {
		test.Fatal("completion dispatch counter mismatch")
	}
}

func TestContinuousUpstreamRelayHTTP(test *testing.T) {
	var dispatched atomic.Int64
	upstream := httptest.NewServer(continuousUpstream(&dispatched))
	defer upstream.Close()
	root := test.TempDir()
	password, err := relay.InitConsole(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err != nil {
		test.Fatal("isolated console initialization failed")
	}
	store, err := relay.Open(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err != nil {
		test.Fatal("isolated console open failed")
	}
	defer store.Close()
	server := relay.New(store, relay.Config{Authority: "console.test", Policy: relay.Policy{TestAuthority: strings.TrimPrefix(upstream.URL, "http://")}})
	var cookie *http.Cookie
	var csrf string
	admin := func(method, path string, body any, status int) map[string]any {
		test.Helper()
		data, _ := json.Marshal(body)
		request := httptest.NewRequest(method, "http://console.test/admin/api/"+path, bytes.NewReader(data))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "http://console.test")
		request.Header.Set("X-CSRF-Token", csrf)
		request.Header.Set("Idempotency-Key", path)
		if cookie != nil {
			request.AddCookie(cookie)
		}
		writer := httptest.NewRecorder()
		server.ServeHTTP(writer, request)
		if writer.Code != status {
			test.Fatalf("management contract status=%d want=%d", writer.Code, status)
		}
		var result map[string]any
		if json.Unmarshal(writer.Body.Bytes(), &result) != nil {
			test.Fatal("management response shape mismatch")
		}
		if path == "login" {
			cookie = writer.Result().Cookies()[0]
			csrf = result["csrf"].(string)
		}
		return result
	}
	admin("POST", "login", map[string]string{"account": "admin", "password": password}, 200)
	tenant := admin("POST", "tenants", map[string]string{"name": "Controlled fixture contract"}, 201)["id"].(string)
	service := admin("POST", "upstreams", map[string]any{"name": "Controlled fixture only", "url": upstream.URL + "/v1", "token": "controlled-journey-upstream", "enabled": true}, 201)["id"]
	admin("POST", "models", map[string]any{"alias": "controlled", "enabled": true, "candidates": []map[string]any{{"upstream_id": service, "remote_model": "controlled", "enabled": true, "priority": 0, "weight": 1}}}, 201)
	admin("PUT", "models/controlled/tenant-grants", map[string]any{"tenants": []string{tenant}}, 200)
	admin("PUT", "models/controlled/bounds", map[string]any{"upstream_id": service, "max_input": 100, "max_output": 100, "evidence": "controlled test contract; not supplier validation"}, 200)
	admin("POST", "models/controlled/prices", map[string]any{"prices": []map[string]any{{"upstream_id": service, "cost_input_rate": 1000000, "cost_output_rate": 1000000, "tenant_input_rate": 2000000, "tenant_output_rate": 2000000}}}, 201)
	budget := int64(100000)
	admin("PUT", "tenants/"+tenant+"/limits", relay.ResourceLimits{Budget: &budget, Hard: true}, 200)
	key := admin("POST", "tenants/"+tenant+"/keys", relay.KeyInput{Name: "controlled", Application: "test fixture", Enabled: true, Models: []string{"controlled"}}, 201)["key"].(string)
	for _, scenario := range []struct {
		name, message string
		pending       float64
	}{
		{"known", "controlled journey input", 0},
		{"unknown", "controlled unknown usage", 1},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			data, _ := json.Marshal(map[string]any{"model": "controlled", "messages": []map[string]string{{"role": "user", "content": scenario.message}}})
			request := httptest.NewRequest("POST", "http://console.test/v1/chat/completions", bytes.NewReader(data))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+key)
			writer := httptest.NewRecorder()
			server.ServeHTTP(writer, request)
			if writer.Code != 200 {
				test.Fatalf("actual controlled HTTP relay status=%d want=200", writer.Code)
			}
			detail := admin("GET", "workspace/requests/"+writer.Header().Get("X-Request-ID"), nil, 200)
			if detail["pending"] != scenario.pending || detail["state"] != "success" {
				test.Fatal("known/unknown completion persistence mismatch")
			}
			if scenario.pending == 0 {
				entries, ok := detail["charge_entries"].([]any)
				if !ok || len(entries) != 1 {
					test.Fatal("known charge entry missing")
				}
				charge := entries[0].(map[string]any)
				if charge["fee"] != "14" || charge["cost"] != "7" {
					test.Fatal("controlled fixture known price amounts changed")
				}
			}
		})
	}
	if dispatched.Load() != 2 {
		test.Fatal("actual controlled HTTP dispatch count mismatch")
	}
}
