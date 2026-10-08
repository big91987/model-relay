package relay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type fixture struct {
	store     *Store
	server    *Server
	http      *httptest.Server
	upstream  *httptest.Server
	password  string
	cookie    *http.Cookie
	csrf      string
	key       string
	keyID     string
	directory string
	keyFile   string
}

func localServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("real HTTP listener unavailable: %v", err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener.Close()
	server.Listener = listener
	if relayServer, ok := handler.(*Server); ok {
		relayServer.Config.Authority = listener.Addr().String()
	}
	server.Start()
	t.Cleanup(server.Close)
	return server
}
func installation(t *testing.T) (*Store, string, string, string) {
	t.Helper()
	root := t.TempDir()
	directory := filepath.Join(root, "data")
	keyFile := filepath.Join(root, "material.key")
	password, err := Init(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	store, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store, password, directory, keyFile
}
func newFixture(t *testing.T, handler http.HandlerFunc) *fixture {
	t.Helper()
	store, password, directory, keyFile := installation(t)
	upstream := localServer(t, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		data, err := io.ReadAll(io.LimitReader(request.Body, (1<<20)+1))
		if err != nil || len(data) > 1<<20 {
			http.Error(writer, "invalid controlled request", 400)
			return
		}
		request.Body = io.NopCloser(bytes.NewReader(data))
		handler(writer, request)
	}))
	policy := Policy{TestAuthority: strings.TrimPrefix(upstream.URL, "http://")}
	server := New(store, Config{Policy: policy, TotalTimeout: 2 * time.Second, IdleTimeout: 250 * time.Millisecond, HeaderTimeout: 250 * time.Millisecond})
	service := localServer(t, server)
	value := &fixture{store: store, server: server, http: service, upstream: upstream, password: password, directory: directory, keyFile: keyFile}
	response := value.admin(t, "POST", "login", map[string]any{"password": password})
	if response.StatusCode != 200 {
		t.Fatal("login failed", response.StatusCode)
	}
	var login struct{ CSRF string }
	decode(t, response, &login)
	value.csrf = login.CSRF
	for _, cookie := range response.Cookies() {
		if cookie.Name == "mr_session" {
			value.cookie = cookie
		}
	}
	response = value.admin(t, "PUT", "upstream", map[string]any{"url": upstream.URL + "/v1", "token": "upstream-secret-sentinel", "enabled": true})
	expect(t, response, 201)
	response = value.admin(t, "POST", "models", map[string]any{"alias": "alpha", "model": "provider-model", "enabled": true})
	expect(t, response, 201)
	response = value.admin(t, "POST", "keys", map[string]any{"name": "first"})
	var issued struct {
		Key      string
		Metadata AppKey
	}
	decode(t, response, &issued)
	value.key = issued.Key
	value.keyID = issued.Metadata.ID
	return value
}
func (value *fixture) admin(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()
	method, path, body = singleFixtureInput(value.store, method, path, body)
	data, _ := json.Marshal(body)
	request, err := http.NewRequest(method, value.http.URL+"/admin/api/"+path, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", value.http.URL)
	request.Header.Set("X-CSRF-Token", value.csrf)
	if value.cookie != nil {
		request.AddCookie(value.cookie)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}
func (value *fixture) call(t *testing.T, body string) *http.Response {
	t.Helper()
	request, _ := http.NewRequest("POST", value.http.URL+"/v1/chat/completions", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+value.key)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}
func expect(t *testing.T, response *http.Response, status int) {
	t.Helper()
	data, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status {
		t.Fatalf("status %d, wanted %d; %s", response.StatusCode, status, data)
	}
}
func decode(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}
func normal(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprint(writer, `{"model":"provider-model","choices":[{"message":{"role":"assistant","content":"answer-sentinel"}}],"usage":{"prompt_tokens":0,"completion_tokens":2}}`)
}

const ordinary = `{"model":"alpha","messages":[{"role":"user","content":"prompt-sentinel"}]}`
const streaming = `{"model":"alpha","messages":[{"role":"user","content":"prompt-sentinel"}],"stream":true}`

func lastRecord(t *testing.T, value *fixture) Record {
	t.Helper()
	response := value.admin(t, "GET", "requests", nil)
	var records struct{ Data []Record }
	decode(t, response, &records)
	if len(records.Data) == 0 {
		t.Fatal("record missing")
	}
	return records.Data[0]
}
func waitState(t *testing.T, value *fixture, state string) Record {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		record := lastRecord(t, value)
		if record.State == state {
			return record
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("state did not reach %s", state)
	return Record{}
}

func TestInstallationRecovery(t *testing.T) {
	store, password, directory, keyFile := installation(t)
	if _, err := Init(directory, keyFile); err == nil {
		t.Fatal("repeat init accepted")
	}
	if !store.Password(password) || store.Password("wrong") {
		t.Fatal("password isolation")
	}
	if _, err := Open(directory, keyFile); err == nil {
		t.Fatal("parallel serve accepted")
	}
	archive := filepath.Join(t.TempDir(), "installation.backup")
	if err := Backup(directory, archive); err == nil {
		t.Fatal("live backup accepted")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := Backup(directory, archive); err != nil {
		t.Fatal(err)
	}
	restored := filepath.Join(t.TempDir(), "restore")
	wrong := filepath.Join(t.TempDir(), "wrong.key")
	os.WriteFile(wrong, bytes.Repeat([]byte{1}, 32), 0600)
	if Restore(archive, restored, wrong) == nil {
		t.Fatal("wrong key restored")
	}
	if _, err := os.Stat(restored); !os.IsNotExist(err) {
		t.Fatal("failed restore installed data")
	}
	if err := Restore(archive, restored, keyFile); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(restored, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if !reopened.Password(password) {
		t.Fatal("restored login lost")
	}
	if Restore(archive, restored, keyFile) == nil {
		t.Fatal("restore overwrote installation")
	}
	corrupt := filepath.Join(t.TempDir(), "corrupt.backup")
	os.WriteFile(corrupt, []byte("not archive"), 0600)
	if Restore(corrupt, filepath.Join(t.TempDir(), "bad"), keyFile) == nil {
		t.Fatal("corrupt archive accepted")
	}
	for _, file := range []string{keyFile, archive, filepath.Join(directory, "relay.db")} {
		info, err := os.Stat(file)
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatal("unsafe permission")
		}
	}
}
func TestInterruptedAndVersion(t *testing.T) {
	store, _, directory, keyFile := installation(t)
	if err := store.Start(Record{ID: "interrupted", Time: time.Now().Format(time.RFC3339)}); err != nil {
		t.Fatal(err)
	}
	store.Close()
	reopened, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	var state string
	var duration *int64
	reopened.DB.QueryRow("SELECT state,duration_ms FROM requests WHERE id='interrupted'").Scan(&state, &duration)
	if state != "process_interrupted" || duration != nil {
		t.Fatal("interrupted falsely finalized")
	}
	reopened.DB.Exec("UPDATE meta SET version=999")
	reopened.Close()
	if _, err = Open(directory, keyFile); err == nil {
		t.Fatal("future schema accepted")
	}
}
func TestSourceAndLoginLimits(t *testing.T) {
	store, password, _, _ := installation(t)
	server := New(store, Config{Authority: "127.0.0.1:5544"})
	request := func(origin, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "http://127.0.0.1:5544/admin/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, req)
		return recorder
	}
	for _, origin := range []string{"", "null", "http://evil.example"} {
		if request(origin, `{"password":"bad"}`).Code != 403 {
			t.Fatal("origin bypass")
		}
	}
	for count := 0; count < 10; count++ {
		if request("http://127.0.0.1:5544", `{"password":"bad"}`).Code != 401 {
			t.Fatal("unexpected login result")
		}
	}
	payload, _ := json.Marshal(map[string]string{"password": password})
	if request("http://127.0.0.1:5544", string(payload)).Code != 429 {
		t.Fatal("login not rate limited")
	}
}

type fixedResolver struct{ Addresses []netip.Addr }

func (resolver fixedResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return resolver.Addresses, nil
}
func TestDestinationAndProtocol(t *testing.T) {
	policy := Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}
	for _, raw := range []string{"http://example.com/v1", "https://u:p@example.com", "https://example.com/?x=1", "https://example.com/#x", "https://example.com/a/../b", "https://example.com/a%2fb", "https://127.0.0.1/v1", "https://10.0.0.1", "https://[::ffff:127.0.0.1]", "https://100.64.0.1"} {
		if _, err := policy.URL(context.Background(), raw); err == nil {
			t.Fatalf("unsafe URL accepted %s", raw)
		}
	}
	if _, err := policy.URL(context.Background(), "https://public.example/v1"); err != nil {
		t.Fatal(err)
	}
	policy.Resolver = fixedResolver{[]netip.Addr{netip.MustParseAddr("127.0.0.1")}}
	if _, err := policy.Dial(context.Background(), "tcp", "public.example:443"); err == nil {
		t.Fatal("rebound private DNS accepted")
	}
	policy.Resolver = fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("10.0.0.1")}}
	if _, err := policy.URL(context.Background(), "https://public.example"); err == nil {
		t.Fatal("mixed private DNS accepted")
	}
	for _, payload := range []string{`{"model":"a","model":"b","messages":[]}`, ordinary + ` {}`, `{"model":"alpha","messages":[{"role":"user","content":[{"type":"image"}]}]}`, `{"model":"alpha","messages":[{"role":"user","content":"text"}],"base_url":"http://evil"}`, `{"model":"alpha","messages":[{"role":"user","content":"text"}],"stream":"true"}`, `{"model":"alpha","messages":[{"role":"user","content":"text"}],"n":0}`} {
		if _, _, _, code := validateRequest([]byte(payload)); code == "" {
			t.Fatal("invalid request accepted")
		}
	}
	if _, _, _, code := validateRequest([]byte(ordinary)); code != "" {
		t.Fatal(code)
	}
	usage := usageFrom(map[string]json.RawMessage{"usage": json.RawMessage(`{"prompt_tokens":0,"completion_tokens":null,"total_tokens":-2}`)})
	if usage.Prompt == nil || *usage.Prompt != 0 || usage.Completion != nil || usage.Total != nil {
		t.Fatal("unknown usage fabricated")
	}
}
func TestHTTPManagementAndOrdinary(t *testing.T) {
	var calls atomic.Int32
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		if request.Header.Get("Authorization") != "Bearer upstream-secret-sentinel" || request.Header.Get("Cookie") != "" {
			t.Error("credential boundary")
		}
		var payload map[string]json.RawMessage
		json.NewDecoder(request.Body).Decode(&payload)
		if string(payload["model"]) != `"provider-model"` {
			t.Error("mapping failed")
		}
		normal(writer, request)
	})
	response := value.call(t, ordinary)
	expect(t, response, 200)
	record := lastRecord(t, value)
	if record.State != "success" || record.Prompt == nil || *record.Prompt != 0 || record.Total != nil || record.Duration == nil || record.KeyID != value.keyID {
		t.Fatal("record incorrect")
	}
	for _, path := range []string{"upstream", "keys", "requests"} {
		response = value.admin(t, "GET", path, nil)
		data, _ := io.ReadAll(response.Body)
		response.Body.Close()
		for _, secret := range []string{value.key, "upstream-secret-sentinel", value.password, "prompt-sentinel", "answer-sentinel"} {
			if bytes.Contains(data, []byte(secret)) {
				t.Fatal("management leaked secret")
			}
		}
	}
	response = value.admin(t, "POST", "models", map[string]any{"alias": "alpha", "model": "other", "enabled": true})
	expect(t, response, 409)
	response = value.admin(t, "PUT", "upstream", map[string]any{"url": value.upstream.URL + "/other", "token": "", "enabled": true})
	expect(t, response, 400)
	response = value.admin(t, "PUT", "upstream", map[string]any{"url": value.upstream.URL + "/v1", "token": "", "enabled": true})
	expect(t, response, 200)
	expect(t, value.call(t, ordinary), 200)
	response = value.admin(t, "PATCH", "models/alpha", map[string]any{"model": "provider-model", "enabled": false})
	expect(t, response, 200)
	expect(t, value.call(t, ordinary), 404)
	if calls.Load() != 2 {
		t.Fatal("rejection forwarded")
	}
	response = value.admin(t, "POST", "keys", map[string]any{"name": "second"})
	var second struct {
		Key      string
		Metadata AppKey
	}
	decode(t, response, &second)
	response = value.admin(t, "PATCH", "keys/"+value.keyID, map[string]bool{"enabled": false})
	expect(t, response, 200)
	expect(t, value.call(t, ordinary), 401)
	value.key = second.Key
	response = value.admin(t, "PATCH", "models/alpha", map[string]any{"model": "provider-model", "enabled": true})
	expect(t, response, 200)
	expect(t, value.call(t, ordinary), 200)
	response = value.admin(t, "PUT", "upstream", map[string]any{"url": value.upstream.URL + "/v1", "enabled": false})
	expect(t, response, 200)
	expect(t, value.call(t, ordinary), 503)
	if calls.Load() != 3 {
		t.Fatal("disable forwarded")
	}
	request, _ := http.NewRequest("GET", value.http.URL+"/v1/models", nil)
	request.Header.Set("Authorization", "Bearer "+value.key)
	response, _ = http.DefaultClient.Do(request)
	var models struct{ Data []any }
	decode(t, response, &models)
	if len(models.Data) != 0 {
		t.Fatal("disabled route visible")
	}
	request, _ = http.NewRequest("GET", value.http.URL+"/v1/models", nil)
	request.AddCookie(value.cookie)
	response, _ = http.DefaultClient.Do(request)
	expect(t, response, 401)
	request, _ = http.NewRequest("GET", value.http.URL+"/admin/api/keys", nil)
	request.Header.Set("Authorization", "Bearer "+value.key)
	response, _ = http.DefaultClient.Do(request)
	expect(t, response, 401)
	request, _ = http.NewRequest("PATCH", value.http.URL+"/admin/api/keys/"+value.keyID, strings.NewReader(`{"enabled":true}`))
	request.AddCookie(value.cookie)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", value.csrf)
	response, _ = http.DefaultClient.Do(request)
	expect(t, response, 403)
	request.Header.Set("Origin", value.http.URL)
	request.Header.Del("X-CSRF-Token")
	response, _ = http.DefaultClient.Do(request)
	expect(t, response, 403)
	response = value.admin(t, "POST", "logout", map[string]any{})
	expect(t, response, 200)
	response = value.admin(t, "GET", "keys", nil)
	expect(t, response, 401)
	files, err := filepath.Glob(filepath.Join(value.directory, "relay.db*"))
	if err != nil || len(files) == 0 {
		t.Fatal("database inventory unavailable")
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal("database secret scan failed")
		}
		for _, secret := range []string{value.key, second.Key, "upstream-secret-sentinel", value.password, "prompt-sentinel", "answer-sentinel"} {
			if bytes.Contains(data, []byte(secret)) {
				t.Fatal("database or WAL plaintext secret")
			}
		}
	}
}
func TestTransparentParametersAndLocalReject(t *testing.T) {
	var calls atomic.Int32
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		data, _ := io.ReadAll(request.Body)
		for _, field := range []string{`"temperature":0.25`, `"tool_choice":"auto"`, `"tools"`, `9007199254740993`} {
			if !bytes.Contains(data, []byte(field)) {
				t.Error("field not transparent", field)
			}
		}
		normal(writer, request)
	})
	payload := `{"model":"alpha","messages":[{"role":"user","content":"text"}],"temperature":0.25,"seed":9007199254740993,"tools":[{"type":"function","function":{"name":"look","parameters":{"type":"object"}}}],"tool_choice":"auto"}`
	expect(t, value.call(t, payload), 200)
	for _, bad := range []string{`{`, strings.Repeat(" ", 1<<20) + ordinary, `{"model":"alpha","messages":[],"headers":{}}`, strings.Replace(ordinary, "alpha", "missing", 1)} {
		response := value.call(t, bad)
		if response.StatusCode < 400 {
			t.Fatal("invalid request succeeded")
		}
		response.Body.Close()
	}
	if calls.Load() != 1 {
		t.Fatal("invalid request forwarded")
	}
	if lastRecord(t, value).State != "rejected" {
		t.Fatal("rejected record absent")
	}
}
func TestSSEBarrierAndUsage(t *testing.T) {
	release := make(chan struct{})
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(writer, ": hello\r\ndata: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\r\n\r\n")
		writer.(http.Flusher).Flush()
		select {
		case <-release:
		case <-request.Context().Done():
			return
		}
		fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"{}\"}}]}}]}\n\ndata: {\"choices\":[],\"usage\":{\"prompt_tokens\":3,\"total_tokens\":5}}\n\ndata: [DONE]\n\n")
	})
	response := value.call(t, streaming)
	reader := bufio.NewReader(response.Body)
	first, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(first, "first") {
		t.Fatal("first chunk not received before barrier")
	}
	close(release)
	rest, err := io.ReadAll(reader)
	response.Body.Close()
	if err != nil || !bytes.Contains(rest, []byte("[DONE]")) || !bytes.Contains(rest, []byte("tool_calls")) {
		t.Fatal("SSE not complete")
	}
	record := lastRecord(t, value)
	if record.State != "success" || record.Prompt == nil || *record.Prompt != 3 || record.Completion != nil || record.Total == nil || *record.Total != 5 {
		t.Fatal("SSE usage incorrect")
	}
}
func TestCancellation(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			seen := make(chan struct{})
			entered := make(chan struct{})
			var count atomic.Int32
			value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
				count.Add(1)
				close(entered)
				if stream {
					writer.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\n")
					writer.(http.Flusher).Flush()
				}
				<-request.Context().Done()
				close(seen)
			})
			ctx, cancel := context.WithCancel(context.Background())
			body := ordinary
			if stream {
				body = streaming
			}
			request, _ := http.NewRequestWithContext(ctx, "POST", value.http.URL+"/v1/chat/completions", strings.NewReader(body))
			request.Header.Set("Authorization", "Bearer "+value.key)
			request.Header.Set("Content-Type", "application/json")
			completed := make(chan struct{})
			go func() {
				response, _ := http.DefaultClient.Do(request)
				if response != nil {
					io.Copy(io.Discard, response.Body)
					response.Body.Close()
				}
				close(completed)
			}()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("upstream not reached")
			}
			cancel()
			select {
			case <-seen:
			case <-time.After(5 * time.Second):
				t.Fatal("upstream not canceled")
			}
			<-completed
			waitState(t, value, "canceled")
			if count.Load() != 1 {
				t.Fatal("cancellation replayed")
			}
		})
	}
}
func TestUpstreamFailures(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		body       string
		kind       string
		wantStatus int
		state      string
	}{
		{"auth", 401, "sensitive provider diagnostic", "application/json", 502, "upstream_rejected"},
		{"rate", 429, "sensitive provider diagnostic", "application/json", 502, "upstream_failed"},
		{"server", 500, "sensitive provider diagnostic", "application/json", 502, "upstream_failed"},
		{"bad_type", 200, "x", "text/html", 502, "upstream_failed"},
		{"bad_json", 200, `{"error":{"message":"secret"}}`, "application/json", 502, "upstream_failed"},
		{"bad_stream_type", 200, "x", "application/json", 502, "upstream_failed"},
		{"stream_error", 200, "data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\ndata: {\"error\":{\"message\":\"sensitive provider diagnostic\"}}\n\n", "text/event-stream", 200, "stream_interrupted"},
		{"stream_eof", 200, "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":0}}\n\n", "text/event-stream", 200, "stream_interrupted"},
		{"stream_extension", 200, "event: vendor\ndata: {}\n\n", "text/event-stream", 502, "stream_interrupted"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var count atomic.Int32
			value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
				count.Add(1)
				writer.Header().Set("Content-Type", test.kind)
				writer.WriteHeader(test.status)
				fmt.Fprint(writer, test.body)
			})
			body := ordinary
			if strings.HasPrefix(test.name, "stream_") || test.name == "bad_stream_type" {
				body = streaming
			}
			response := value.call(t, body)
			data, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil || response.StatusCode != test.wantStatus || bytes.Contains(data, []byte("sensitive provider diagnostic")) || bytes.Contains(data, []byte("[DONE]")) {
				t.Fatal("failure response incorrect")
			}
			record := lastRecord(t, value)
			if record.State != test.state || count.Load() != 1 {
				t.Fatal("failure state or replay incorrect", record.State)
			}
			if test.name == "stream_eof" && (record.Prompt == nil || *record.Prompt != 0) {
				t.Fatal("partial usage lost")
			}
		})
	}
}
func TestTimeoutRedirectAndNetwork(t *testing.T) {
	for _, mode := range []string{"header_timeout", "idle_timeout", "redirect", "network"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			var targetCount atomic.Int32
			target := localServer(t, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) { targetCount.Add(1) }))
			var count atomic.Int32
			value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
				count.Add(1)
				switch mode {
				case "header_timeout":
					<-request.Context().Done()
				case "idle_timeout":
					writer.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\n")
					writer.(http.Flusher).Flush()
					<-request.Context().Done()
				case "redirect":
					http.Redirect(writer, request, target.URL, 307)
				}
			})
			body := ordinary
			if mode == "idle_timeout" {
				body = streaming
			}
			if mode == "network" {
				value.upstream.Close()
			}
			response := value.call(t, body)
			data, _ := io.ReadAll(response.Body)
			response.Body.Close()
			record := lastRecord(t, value)
			if mode == "idle_timeout" {
				if record.State != "timeout" {
					t.Fatal("timeout state", record.State)
				}
			} else if record.State != "upstream_failed" {
				t.Fatal("failure state")
			}
			if count.Load() > 1 || targetCount.Load() != 0 || bytes.Contains(data, []byte("[DONE]")) {
				t.Fatal("replayed or redirect followed")
			}
		})
	}
}
func TestStorageFailClosed(t *testing.T) {
	for _, phase := range []string{"start", "finish", "stream_finish"} {
		t.Run(phase, func(t *testing.T) {
			var count atomic.Int32
			value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
				count.Add(1)
				if phase == "stream_finish" {
					writer.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\ndata: [DONE]\n\n")
					return
				}
				normal(writer, request)
			})
			var statement string
			switch phase {
			case "start":
				statement = `CREATE TRIGGER fail_request BEFORE INSERT ON requests BEGIN SELECT RAISE(FAIL,'test record failure'); END`
			default:
				statement = `CREATE TRIGGER fail_request BEFORE UPDATE ON requests BEGIN SELECT RAISE(FAIL,'test record failure'); END`
			}
			if _, err := value.store.DB.Exec(statement); err != nil {
				t.Fatal(err)
			}
			body := ordinary
			if phase == "stream_finish" {
				body = streaming
			}
			response := value.call(t, body)
			data, _ := io.ReadAll(response.Body)
			response.Body.Close()
			if phase == "stream_finish" {
				if bytes.Contains(data, []byte("[DONE]")) || !bytes.Contains(data, []byte("storage_unavailable")) {
					t.Fatal("false stream success")
				}
			} else if response.StatusCode != 503 {
				t.Fatal("false ordinary success")
			}
			before := count.Load()
			expect(t, value.call(t, ordinary), 503)
			response, _ = http.Get(value.http.URL + "/healthz")
			expect(t, response, 503)
			if count.Load() != before || phase == "start" && before != 0 {
				t.Fatal("degraded service forwarded")
			}
		})
	}
}
