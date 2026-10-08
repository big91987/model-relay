package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSpecialInstallationPath(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "literal?mode=memory#data")
	keyFile := filepath.Join(root, "key")
	password, err := Init(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	store, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if !store.Password(password) {
		t.Fatal("encoded installation path lost identity")
	}
	if _, err = os.Stat(filepath.Join(directory, "relay.db")); err != nil {
		t.Fatal("database not persisted in requested directory")
	}
}
func TestHeaderBoundaryAndFailureContract(t *testing.T) {
	store, password, _, _ := installation(t)
	server := New(store, Config{Authority: "127.0.0.1:5544", Policy: Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}, TotalTimeout: time.Second})
	session := randomSecret(32)
	csrf := randomSecret(32)
	server.sessions[string(digest(session))] = struct {
		CSRF   string
		Expiry time.Time
		UserID string
	}{CSRF: csrf, Expiry: time.Now().Add(time.Hour)}
	admin := func(method, path, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, "http://127.0.0.1:5544/admin/api/"+path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "http://127.0.0.1:5544")
		request.Header.Set("X-CSRF-Token", csrf)
		request.AddCookie(&http.Cookie{Name: "mr_session", Value: session})
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}
	if !store.Password(password) {
		t.Fatal("identity invalid")
	}
	if response := admin("POST", "upstreams", `{"name":"boundary fixture","url":"https://public.example/v1","token":"secret-boundary","enabled":true}`); response.Code != 201 {
		t.Fatal("config failed")
	}
	if response := admin("POST", "models", `{"alias":"alpha","candidates":[{"upstream_id":1,"remote_model":"real","enabled":true,"priority":0,"weight":1}],"enabled":true}`); response.Code != 201 {
		t.Fatal("model failed")
	}
	var issued struct {
		Key      string
		Metadata AppKey
	}
	json.Unmarshal(admin("POST", "keys", `{"name":"first"}`).Body.Bytes(), &issued)
	for _, bad := range []string{`null`, `{}`, `{"enabled":null}`, `{"enabled":false,"enabled":true}`} {
		if response := admin("PATCH", "keys/"+issued.Metadata.ID, bad); response.Code != 400 {
			t.Fatal("invalid state update accepted")
		}
	}
	call := func(body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest("POST", "http://127.0.0.1:5544/v1/chat/completions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+issued.Key)
		request.Header.Set("Cookie", "client-secret")
		request.Header.Set("X-Forwarded-For", "client-ip")
		request.Header.Set("Idempotency-Key", "client-retry")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}
	calls := 0
	server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.Header.Get("Authorization") != "Bearer secret-boundary" || request.Header.Get("Cookie") != "" || request.Header.Get("X-Forwarded-For") != "" || request.Header.Get("Idempotency-Key") != "" || request.GetBody != nil {
			t.Fatal("client headers/replay leaked")
		}
		response := httptest.NewRecorder()
		response.Header().Set("Set-Cookie", "provider-secret")
		response.Header().Set("X-Provider-Secret", "provider-secret")
		normal(response, request)
		return response.Result(), nil
	})
	response := call(ordinary)
	if response.Code != 200 || response.Header().Get("Set-Cookie") != "" || response.Header().Get("X-Provider-Secret") != "" {
		t.Fatal("upstream headers leaked")
	}
	server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		calls++
		response := httptest.NewRecorder()
		response.Header().Set("Content-Type", "text/event-stream")
		response.WriteString("data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\ndata: {\"error\":{\"message\":\"upstream-secret-diagnostic\"}}\n\n")
		return response.Result(), nil
	})
	response = call(streaming)
	if response.Code != 200 || !bytes.Contains(response.Body.Bytes(), []byte("event: error")) || bytes.Contains(response.Body.Bytes(), []byte("[DONE]")) || bytes.Contains(response.Body.Bytes(), []byte("upstream-secret-diagnostic")) {
		t.Fatal("incomplete stream advertised success")
	}
	server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		calls++
		<-request.Context().Done()
		return nil, request.Context().Err()
	})
	server.Config.TotalTimeout = 10 * time.Millisecond
	response = call(ordinary)
	if response.Code != 504 {
		t.Fatal("timeout incorrectly classified")
	}
	records := admin("GET", "requests", "")
	var list struct{ Data []Record }
	if json.Unmarshal(records.Body.Bytes(), &list) != nil || len(list.Data) != 3 || list.Data[0].State != "timeout" || list.Data[1].State != "stream_interrupted" || calls != 3 {
		t.Fatal("terminal state incorrect")
	}
	if _, err := store.DB.Exec(`CREATE TRIGGER fail_request BEFORE INSERT ON requests BEGIN SELECT RAISE(FAIL,'test record failure'); END`); err != nil {
		t.Fatal(err)
	}
	if response = call(ordinary); response.Code != 503 || calls != 3 || !server.degraded.Load() {
		t.Fatal("admission storage fault forwarded")
	}
	request := httptest.NewRequest("GET", "http://127.0.0.1:5544/healthz", nil)
	health := httptest.NewRecorder()
	server.ServeHTTP(health, request)
	if health.Code != 503 {
		t.Fatal("degraded health not reported")
	}
	request = httptest.NewRequest("GET", "http://evil.example/admin/", nil)
	health = httptest.NewRecorder()
	server.ServeHTTP(health, request)
	if health.Code != 403 {
		t.Fatal("Host spoof accepted")
	}
	if _, _, code := classify(context.Background(), io.ErrUnexpectedEOF, true); code != "stream_interrupted" {
		t.Fatal("unexpected EOF classification")
	}
}
