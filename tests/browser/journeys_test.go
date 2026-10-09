package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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

func TestContinuousUpstreamStreamContract(test *testing.T) {
	var dispatched atomic.Int64
	request := httptest.NewRequest("POST", "http://controlled.test/v1/chat/completions", strings.NewReader(`{"stream":true,"messages":[{"role":"user","content":"controlled journey input"}]}`))
	request.Header.Set("Authorization", "Bearer controlled-journey-upstream")
	writer := httptest.NewRecorder()
	continuousUpstream(&dispatched).ServeHTTP(writer, request)
	if writer.Code != 200 || writer.Header().Get("Content-Type") != "text/event-stream" || !writer.Flushed {
		test.Fatal("controlled SSE requires successful stream headers and an actually flushed first frame")
	}
	frames := strings.Split(strings.TrimSpace(writer.Body.String()), "\n\n")
	if len(frames) != 3 || frames[2] != "data: [DONE]" {
		test.Fatal("controlled SSE requires two structured frames and a terminal frame")
	}
	for index, frame := range frames[:2] {
		var chunk struct {
			Choices []struct {
				Delta struct{ Role, Content string }
			}
			Usage *struct {
				Input  int64 `json:"prompt_tokens"`
				Output int64 `json:"completion_tokens"`
				Total  int64 `json:"total_tokens"`
			}
		}
		if !strings.HasPrefix(frame, "data: ") || json.Unmarshal([]byte(strings.TrimPrefix(frame, "data: ")), &chunk) != nil || len(chunk.Choices) != 1 {
			test.Fatal("controlled SSE structured frame mismatch")
		}
		if index == 0 && (chunk.Choices[0].Delta.Role != "assistant" || chunk.Choices[0].Delta.Content == "" || chunk.Usage != nil) {
			test.Fatal("controlled SSE first frame must carry assistant text, not invented usage")
		}
		if index == 1 && (chunk.Usage == nil || chunk.Usage.Input != 3 || chunk.Usage.Output != 4 || chunk.Usage.Total != 7) {
			test.Fatal("controlled SSE terminal usage changed")
		}
	}
	if dispatched.Load() != 1 {
		test.Fatal("controlled SSE must count exactly one completion dispatch")
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
		stream        bool
	}{
		{"known", "controlled journey input", 0, false},
		{"unknown", "controlled unknown usage", 1, false},
		{"known-stream", "controlled journey input", 0, true},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			data, _ := json.Marshal(map[string]any{"model": "controlled", "messages": []map[string]string{{"role": "user", "content": scenario.message}}, "stream": scenario.stream})
			request := httptest.NewRequest("POST", "http://console.test/v1/chat/completions", bytes.NewReader(data))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+key)
			writer := httptest.NewRecorder()
			server.ServeHTTP(writer, request)
			if writer.Code != 200 {
				test.Fatalf("actual controlled HTTP relay status=%d want=200", writer.Code)
			}
			if scenario.stream && (writer.Header().Get("Content-Type") != "text/event-stream" || !strings.Contains(writer.Body.String(), "data: [DONE]")) {
				test.Fatal("controlled stream response must be actual SSE with a terminal frame")
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
	if dispatched.Load() != 3 {
		test.Fatal("actual controlled HTTP dispatch count mismatch")
	}
}

func TestSelfHostedControlledHTTPBoundary(test *testing.T) {
	var dispatched atomic.Int64
	server := httptest.NewServer(selfHostedUpstream(&dispatched))
	defer server.Close()
	client := &http.Client{Timeout: 2 * time.Second}
	for _, scenario := range []struct {
		path, token string
		status      int
	}{
		{selfHostedPath + "/models", "controlled-journey-upstream", 200},
		{selfHostedPath + "/models", "controlled-rejected-credential", 401},
	} {
		request, err := http.NewRequest("GET", server.URL+scenario.path, nil)
		if err != nil {
			test.Fatal("controlled request creation failed")
		}
		request.Header.Set("Authorization", "Bearer "+scenario.token)
		response, err := client.Do(request)
		if err != nil {
			test.Fatal("controlled check transport failed")
		}
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode != scenario.status {
			test.Fatal("controlled check HTTP status mismatch")
		}
	}
	response, err := client.Get(server.URL + "/unreachable/models")
	if response != nil {
		response.Body.Close()
	}
	if err == nil {
		test.Fatal("closed test connection must fail at transport boundary, not return an HTTP error")
	}
	if dispatched.Load() != 0 {
		test.Fatal("check-only fixture must not count a completion")
	}
}
