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
	"sync/atomic"
	"testing"
	"time"
)

type consoleFixture struct {
	server *Server
	cookie *http.Cookie
	csrf   string
	key    string
	keyID  string
}

func newConsoleFixture(t *testing.T) *consoleFixture {
	t.Helper()
	store, password, _, _ := installation(t)
	value := &consoleFixture{server: New(store, Config{Authority: "console.test", Policy: Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}})}
	response := value.admin("POST", "login", map[string]string{"password": password})
	if response.Code != 200 {
		t.Fatal("login", response.Code)
	}
	value.cookie = response.Result().Cookies()[0]
	var login struct {
		CSRF      string `json:"csrf"`
		ExpiresAt string `json:"expires_at"`
	}
	if json.Unmarshal(response.Body.Bytes(), &login) != nil {
		t.Fatal("login response")
	}
	if expires, err := time.Parse(time.RFC3339, login.ExpiresAt); err != nil || time.Until(expires) < 7*time.Hour {
		t.Fatal("expiry contract")
	}
	value.csrf = login.CSRF
	return value
}

func (value *consoleFixture) admin(method, path string, body any) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	request := httptest.NewRequest(method, "http://console.test/admin/api/"+path, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://console.test")
	request.Header.Set("X-CSRF-Token", value.csrf)
	if value.cookie != nil {
		request.AddCookie(value.cookie)
	}
	response := httptest.NewRecorder()
	value.server.ServeHTTP(response, request)
	return response
}

func (value *consoleFixture) configure(t *testing.T) {
	t.Helper()
	for _, step := range []struct {
		method, path string
		body         any
	}{
		{"PUT", "upstream", map[string]any{"url": "https://public.example/v1", "token": "isolated-upstream", "enabled": true}},
		{"POST", "models", map[string]any{"alias": "alpha", "model": "provider-model", "enabled": true}},
	} {
		if response := value.admin(step.method, step.path, step.body); response.Code != 200 {
			t.Fatal("configure", response.Code)
		}
	}
	var issued struct {
		Key      string
		Metadata AppKey
	}
	response := value.admin("POST", "keys", map[string]string{"name": "isolated"})
	if response.Code != 201 || json.Unmarshal(response.Body.Bytes(), &issued) != nil {
		t.Fatal("issue")
	}
	value.key, value.keyID = issued.Key, issued.Metadata.ID
}

func (value *consoleFixture) call(key, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest("POST", "http://console.test/v1/chat/completions", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	value.server.ServeHTTP(response, request)
	return response
}

func (value *consoleFixture) overview(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	response := value.admin("GET", "overview", nil)
	var result map[string]json.RawMessage
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &result) != nil {
		t.Fatal("overview", response.Code)
	}
	return result
}

func TestConsoleOverviewAndExactRecords(t *testing.T) {
	value := newConsoleFixture(t)
	result := value.overview(t)
	if string(result["call_evidence"]) != "null" || !bytes.Contains(result["upstream"], []byte(`"configured":false`)) {
		t.Fatal("empty facts")
	}
	value.configure(t)
	transaction, err := value.server.Store.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 1005; index++ {
		if _, err = transaction.Exec("INSERT INTO models VALUES(?,?,?)", fmt.Sprintf("bulk-%04d", index), "provider", index%2 == 0); err != nil {
			t.Fatal(err)
		}
		if _, err = transaction.Exec("INSERT INTO app_keys VALUES(?,?,?,?,?,?)", fmt.Sprintf("bulk-%04d", index), "bulk", digest(fmt.Sprintf("isolated-%d", index)), "safe-hint", index%2 == 0, time.Now().UTC().Format(time.RFC3339)); err != nil {
			t.Fatal(err)
		}
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"models", "keys"} {
		var count struct{ Total, Enabled int }
		if json.Unmarshal(value.overview(t)[field], &count) != nil || count.Total != 1006 || count.Enabled != 504 {
			t.Fatal("full aggregate", field, count)
		}
	}
	var posts atomic.Int64
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		posts.Add(1)
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"isolated text"}}],"usage":{"prompt_tokens":0,"completion_tokens":2,"total_tokens":2}}`))}, nil
	})
	response := value.call(value.key, `{"model":"alpha","messages":[{"role":"user","content":"isolated prompt"}]}`)
	if response.Code != 200 || posts.Load() != 1 {
		t.Fatal("controlled success", response.Code)
	}
	id := response.Header().Get("X-Request-ID")
	exact := value.admin("GET", "requests/"+id, nil)
	var record Record
	if exact.Code != 200 || json.Unmarshal(exact.Body.Bytes(), &record) != nil || record.ID != id || record.State != "success" || record.KeyID != value.keyID || record.Prompt == nil || *record.Prompt != 0 || record.Duration == nil {
		t.Fatal("exact metadata")
	}
	if !bytes.Contains(value.overview(t)["call_evidence"], []byte(id)) {
		t.Fatal("persisted success evidence")
	}
	if value.admin("GET", "requests/"+randomSecret(16), nil).Code != 404 || value.admin("GET", "requests/%27%20OR%201=1", nil).Code != 404 || value.admin("POST", "requests/"+id, nil).Code != 405 {
		t.Fatal("exact boundary")
	}
	value.server.mu.Lock()
	value.server.successID = id
	value.server.mu.Unlock()
	record.State = "canceled"
	if value.server.Store.Finish(record) != nil || string(value.overview(t)["call_evidence"]) != "null" {
		t.Fatal("later terminal correction must invalidate")
	}
	value.cookie = nil
	if value.admin("GET", "overview", nil).Code != 401 || value.admin("GET", "requests/"+id, nil).Code != 401 {
		t.Fatal("admin isolation")
	}
}

func TestConsoleEvidenceConfigurationRace(t *testing.T) {
	value := newConsoleFixture(t)
	value.configure(t)
	entered, release := make(chan struct{}), make(chan struct{})
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		close(entered)
		<-release
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"isolated"}}]}`))}, nil
	})
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		done <- value.call(value.key, `{"model":"alpha","messages":[{"role":"user","content":"isolated"}]}`)
	}()
	<-entered
	response := value.admin("PATCH", "models/alpha", map[string]any{"model": "new-model", "enabled": true})
	close(release)
	if response.Code != 200 || (<-done).Code != 200 || string(value.overview(t)["call_evidence"]) != "null" {
		t.Fatal("late old config success")
	}
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"isolated"}}]}`))}, nil
	})
	for _, step := range []struct {
		method, path string
		body         any
	}{
		{"PUT", "upstream", map[string]any{"url": "https://public.example/v1", "token": "", "enabled": true}},
		{"POST", "models", map[string]any{"alias": "beta", "model": "provider", "enabled": true}},
		{"PATCH", "models/alpha", map[string]any{"model": "new-model", "enabled": true}},
		{"POST", "keys", map[string]string{"name": "second"}},
		{"PATCH", "keys/" + value.keyID, map[string]bool{"enabled": true}},
	} {
		if value.call(value.key, `{"model":"alpha","messages":[{"role":"user","content":"isolated"}]}`).Code != 200 || string(value.overview(t)["call_evidence"]) == "null" {
			t.Fatal("success before mutation")
		}
		if response := value.admin(step.method, step.path, step.body); response.Code >= 300 || string(value.overview(t)["call_evidence"]) != "null" {
			t.Fatal("mutation invalidation", step.path)
		}
	}
	if value.call(value.key, `{"model":"alpha","messages":[{"role":"user","content":"isolated"}]}`).Code != 200 {
		t.Fatal("success before restart")
	}
	value.server = New(value.server.Store, value.server.Config)
	value.server.sessions[string(digest(value.cookie.Value))] = session{CSRF: value.csrf, Expiry: time.Now().Add(time.Hour)}
	if string(value.overview(t)["call_evidence"]) != "null" {
		t.Fatal("restart must not inherit success")
	}
}

func TestConsoleDeniedMetadataAndStorage(t *testing.T) {
	value := newConsoleFixture(t)
	value.configure(t)
	var posts atomic.Int64
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		posts.Add(1)
		return nil, fmt.Errorf("must not forward")
	})
	if value.admin("PATCH", "keys/"+value.keyID, map[string]bool{"enabled": false}).Code != 200 {
		t.Fatal("disable")
	}
	for _, testCase := range []struct {
		key, body, expectedKey, alias string
	}{
		{"", `{"model":"alpha"}`, "", ""},
		{"mr_unknown", `{"model":"alpha"}`, "", ""},
		{value.key, `{"model":"alpha","messages":[{"role":"user","content":"isolated"}]}`, value.keyID, "alpha"},
		{value.key, `not JSON`, value.keyID, ""},
		{value.key, strings.Repeat("x", (1<<20)+1), value.keyID, ""},
	} {
		response := value.call(testCase.key, testCase.body)
		if response.Code != 401 || !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
			t.Fatal("unified authorization", response.Code)
		}
		var record Record
		exact := value.admin("GET", "requests/"+response.Header().Get("X-Request-ID"), nil)
		if exact.Code != 200 || json.Unmarshal(exact.Body.Bytes(), &record) != nil || record.State != "rejected" || record.KeyID != testCase.expectedKey || record.Alias != testCase.alias || record.Total != nil {
			t.Fatal("safe rejection metadata")
		}
	}
	if posts.Load() != 0 {
		t.Fatal("authorization must never forward")
	}
	results := make(chan *httptest.ResponseRecorder, 16)
	for index := 0; index < cap(results); index++ {
		go func() { results <- value.call("mr_unknown", `{"model":"alpha"}`) }()
	}
	for index := 0; index < cap(results); index++ {
		response := <-results
		if response.Code != 401 || value.admin("GET", "requests/"+response.Header().Get("X-Request-ID"), nil).Code != 200 {
			t.Fatal("concurrent refusal persistence")
		}
	}
	if posts.Load() != 0 {
		t.Fatal("concurrent refusal forwarded")
	}
	value.server.Store.DB.Close()
	if value.admin("GET", "overview", nil).Code != 503 || value.admin("GET", "requests/"+randomSecret(16), nil).Code != 503 || value.call("mr_unknown", `{}`).Code != 503 {
		t.Fatal("storage failure must not masquerade as empty or authorize")
	}
}

func TestConsoleRealHTTP(t *testing.T) {
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		io.WriteString(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled fixture"}}]}`)
	})
	response := value.call(t, `{"model":"alpha","messages":[{"role":"user","content":"fixture"}]}`)
	id := response.Header.Get("X-Request-ID")
	expect(t, response, 200)
	var record Record
	decode(t, value.admin(t, "GET", "requests/"+id, nil), &record)
	if record.ID != id || record.State != "success" || record.Total != nil {
		t.Fatal("real HTTP exact record")
	}
	var overview map[string]json.RawMessage
	decode(t, value.admin(t, "GET", "overview", nil), &overview)
	if !bytes.Contains(overview["call_evidence"], []byte(id)) {
		t.Fatal("real HTTP evidence")
	}
	value.key = "mr_unknown"
	response = value.call(t, `{"model":"alpha","messages":[{"role":"user","content":"fixture"}]}`)
	id = response.Header.Get("X-Request-ID")
	expect(t, response, 401)
	decode(t, value.admin(t, "GET", "requests/"+id, nil), &record)
	if record.State != "rejected" || record.KeyID != "" || record.Alias != "" {
		t.Fatal("real HTTP safe rejection")
	}
}
