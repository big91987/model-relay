package relay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (function roundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
func TestInProcessContract(t *testing.T) {
	store, password, _, _ := installation(t)
	server := New(store, Config{Authority: "127.0.0.1:5544", Policy: Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}})
	var cookie *http.Cookie
	csrf := ""
	key := ""
	admin := func(method, path string, body any) *httptest.ResponseRecorder {
		data, _ := json.Marshal(body)
		request := httptest.NewRequest(method, "http://127.0.0.1:5544/admin/api/"+path, bytes.NewReader(data))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "http://127.0.0.1:5544")
		request.Header.Set("X-CSRF-Token", csrf)
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}
	login := admin("POST", "login", map[string]string{"password": password})
	if login.Code != 200 {
		t.Fatal("login contract", login.Code)
	}
	cookie = login.Result().Cookies()[0]
	var signedIn struct{ CSRF string }
	json.Unmarshal(login.Body.Bytes(), &signedIn)
	csrf = signedIn.CSRF
	for _, step := range []struct {
		method, path string
		body         any
	}{
		{"PUT", "upstream", map[string]any{"url": "https://public.example/v1", "token": "upstream-sentinel", "enabled": true}},
		{"POST", "models", map[string]any{"alias": "alpha", "model": "real-model", "enabled": true}},
		{"POST", "models", map[string]any{"alias": "beta", "model": "second-model", "enabled": true}},
	} {
		if response := admin(step.method, step.path, step.body); response.Code != 200 {
			t.Fatal("configuration contract", response.Code, response.Body.String())
		}
	}
	issued := admin("POST", "keys", map[string]string{"name": "first"})
	if issued.Code != 201 {
		t.Fatal("key contract")
	}
	var secret struct {
		Key      string
		Metadata AppKey
	}
	json.Unmarshal(issued.Body.Bytes(), &secret)
	key = secret.Key
	var count int
	server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		count++
		if request.Header.Get("Authorization") != "Bearer upstream-sentinel" {
			t.Fatal("upstream identity")
		}
		data, _ := io.ReadAll(request.Body)
		if !bytes.Contains(data, []byte(`"model":"real-model"`)) {
			t.Fatal("mapped model")
		}
		response := httptest.NewRecorder()
		if strings.Contains(string(data), `"stream":true`) {
			response.Header().Set("Content-Type", "text/event-stream")
			response.WriteString("data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\ndata: {\"choices\":[],\"usage\":{\"prompt_tokens\":0}}\n\ndata: {\"choices\":[{\"delta\":{}}],\"usage\":null}\n\ndata: [DONE]\n\n")
		} else {
			normal(response, request)
		}
		return response.Result(), nil
	})
	call := func(body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest("POST", "http://127.0.0.1:5544/v1/chat/completions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+key)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}
	if response := call(ordinary); response.Code != 200 {
		t.Fatal("ordinary contract", response.Code, response.Body.String())
	}
	if response := call(streaming); response.Code != 200 || !bytes.Contains(response.Body.Bytes(), []byte("[DONE]")) {
		t.Fatal("stream contract", response.Code, response.Body.String())
	}
	records := admin("GET", "requests", nil)
	var data struct{ Data []Record }
	if json.Unmarshal(records.Body.Bytes(), &data) != nil || len(data.Data) != 2 || data.Data[0].State != "success" || data.Data[0].Prompt == nil || *data.Data[0].Prompt != 0 || data.Data[0].Total != nil {
		t.Fatal("record contract")
	}
	for _, path := range []string{"upstream", "keys", "requests"} {
		response := admin("GET", path, nil)
		for _, secret := range []string{key, password, "upstream-sentinel", "prompt-sentinel", "answer-sentinel"} {
			if bytes.Contains(response.Body.Bytes(), []byte(secret)) {
				t.Fatal("query leaks protected data")
			}
		}
	}
	if response := admin("PATCH", "keys/"+secret.Metadata.ID, map[string]bool{"enabled": false}); response.Code != 200 {
		t.Fatal("disable contract")
	}
	if response := call(ordinary); response.Code != 401 || count != 2 {
		t.Fatal("disabled key forwarded")
	}
	if response := admin("POST", "logout", map[string]any{}); response.Code != 200 {
		t.Fatal("logout contract")
	}
	if response := admin("GET", "keys", nil); response.Code != 401 {
		t.Fatal("session survived logout")
	}
}
