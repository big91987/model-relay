package latencyfixture

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"model-relay/internal/relay"
)

const Credential = "controlled-latency-only"

type Scenario struct {
	Swap      atomic.Bool
	Heartbeat bool
	Tail      bool
}
type Timing struct {
	Upstream   int     `json:"upstream_id"`
	BodyMS     float64 `json:"body_ms"`
	OutputMS   float64 `json:"output_ms"`
	CompleteMS float64 `json:"complete_ms"`
}
type Fixture struct {
	Root      string
	Password  string
	Store     *relay.Store
	Gateway   *httptest.Server
	Upstream  *httptest.Server
	Client    *http.Client
	CSRF      string
	Tenant    string
	Key       string
	mu        sync.Mutex
	scenarios map[string]*Scenario
	timings   map[string]Timing
}

func New() (*Fixture, error) {
	root, err := os.MkdirTemp("", "relay-latency-controlled-")
	if err != nil {
		return nil, err
	}
	fixture := &Fixture{Root: root, scenarios: map[string]*Scenario{}, timings: map[string]Timing{}}
	fixture.Upstream = httptest.NewServer(http.HandlerFunc(fixture.upstream))
	fixture.Password, err = relay.InitConsole(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err == nil {
		err = fixture.start()
	}
	if err != nil {
		fixture.Close()
		return nil, err
	}
	return fixture, nil
}
func (fixture *Fixture) start() error {
	store, err := relay.Open(filepath.Join(fixture.Root, "data"), filepath.Join(fixture.Root, "key"))
	if err != nil {
		return err
	}
	fixture.Store = store
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		store.Close()
		return err
	}
	server := relay.New(store, relay.Config{Authority: listener.Addr().String(), Policy: relay.Policy{TestAuthority: strings.TrimPrefix(fixture.Upstream.URL, "http://")}})
	gateway := httptest.NewUnstartedServer(server)
	gateway.Listener.Close()
	gateway.Listener = listener
	gateway.Start()
	fixture.Gateway = gateway
	jar, _ := cookiejar.New(nil)
	fixture.Client = &http.Client{Jar: jar, Timeout: 10 * time.Second}
	result, err := fixture.Admin("POST", "login", map[string]any{"account": "admin", "password": fixture.Password}, 200)
	if err != nil {
		return err
	}
	fixture.CSRF, _ = result["csrf"].(string)
	if fixture.CSRF == "" {
		return errors.New("controlled login missing CSRF")
	}
	return nil
}
func (fixture *Fixture) Close() {
	if fixture.Gateway != nil {
		fixture.Gateway.Close()
	}
	if fixture.Store != nil {
		fixture.Store.Close()
	}
	if fixture.Upstream != nil {
		fixture.Upstream.Close()
	}
	if fixture.Root != "" {
		os.RemoveAll(fixture.Root)
	}
}
func (fixture *Fixture) Restart() error {
	fixture.Gateway.Close()
	fixture.Store.Close()
	return fixture.start()
}
func (fixture *Fixture) Admin(method, path string, input any, status int) (map[string]any, error) {
	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(method, fixture.Gateway.URL+"/admin/api/"+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", fixture.Gateway.URL)
	request.Header.Set("X-CSRF-Token", fixture.CSRF)
	request.Header.Set("Idempotency-Key", fmt.Sprintf("controlled-%d", time.Now().UnixNano()))
	response, err := fixture.Client.Do(request)
	if err != nil {
		return nil, errors.New("controlled management transport failure")
	}
	defer response.Body.Close()
	var result map[string]any
	if json.NewDecoder(response.Body).Decode(&result) != nil {
		return nil, fmt.Errorf("controlled management invalid JSON status=%d", response.StatusCode)
	}
	if response.StatusCode != status {
		code := "unknown"
		if value, ok := result["error"].(map[string]any); ok {
			code, _ = value["code"].(string)
		}
		return nil, fmt.Errorf("controlled management %s status=%d expected=%d code=%s", path, response.StatusCode, status, code)
	}
	return result, nil
}
func (fixture *Fixture) Register(alias string, scenario *Scenario) {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	fixture.scenarios[alias] = scenario
}
func (fixture *Fixture) Timing(id string) Timing {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	return fixture.timings[id]
}
func (fixture *Fixture) SetupUpstreams() error {
	for index, label := range []string{"A", "B"} {
		_, err := fixture.Admin("POST", "upstreams", map[string]any{"name": "Controlled " + label, "url": fixture.Upstream.URL + fmt.Sprintf("/%d/v1", index+1), "token": Credential, "enabled": true, "source_kind": "self_hosted"}, 201)
		if err != nil {
			return err
		}
	}
	result, err := fixture.Admin("POST", "tenants", map[string]any{"name": "Controlled experiment team"}, 201)
	if err != nil {
		return err
	}
	fixture.Tenant, _ = result["id"].(string)
	return nil
}
func (fixture *Fixture) Configure(alias, strategy string) error {
	candidates := []map[string]any{}
	for index := 1; index <= 2; index++ {
		candidates = append(candidates, map[string]any{"upstream_id": index, "remote_model": alias, "enabled": true, "priority": 0, "weight": 1})
	}
	if _, err := fixture.Admin("POST", "models", map[string]any{"alias": alias, "enabled": true, "routing_strategy": strategy, "candidates": candidates}, 201); err != nil {
		return err
	}
	return fixture.PrepareModel(alias)
}
func (fixture *Fixture) PrepareModel(alias string) error {
	if _, err := fixture.Admin("PUT", "models/"+alias+"/tenant-grants", map[string]any{"tenants": []string{fixture.Tenant}}, 200); err != nil {
		return err
	}
	prices := []map[string]any{}
	for index := 1; index <= 2; index++ {
		if _, err := fixture.Admin("PUT", "models/"+alias+"/bounds", map[string]any{"upstream_id": index, "max_input": 100, "max_output": 100, "evidence": "controlled HTTP contract, not supplier validation"}, 200); err != nil {
			return err
		}
		prices = append(prices, map[string]any{"upstream_id": index, "cost_input_rate": 1000000, "cost_output_rate": 1000000, "tenant_input_rate": 2000000, "tenant_output_rate": 2000000})
	}
	_, err := fixture.Admin("POST", "models/"+alias+"/prices", map[string]any{"prices": prices}, 201)
	return err
}
func (fixture *Fixture) IssueKey(aliases []string) error {
	result, err := fixture.Admin("POST", "tenants/"+fixture.Tenant+"/keys", map[string]any{"name": "Controlled experiment key", "application": "controlled HTTP experiment", "enabled": true, "models": aliases, "diagnostic_members": []string{}}, 201)
	if err != nil {
		return err
	}
	fixture.Key, _ = result["key"].(string)
	if fixture.Key == "" {
		return errors.New("controlled key missing")
	}
	return nil
}
func (fixture *Fixture) upstream(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Authorization") != "Bearer "+Credential {
		writer.WriteHeader(401)
		return
	}
	if request.Method == "GET" {
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{"data":[]}`)
		return
	}
	var input struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if json.NewDecoder(io.LimitReader(request.Body, 1<<20)).Decode(&input) != nil {
		writer.WriteHeader(400)
		return
	}
	fixture.mu.Lock()
	scenario := fixture.scenarios[input.Model]
	fixture.mu.Unlock()
	if scenario == nil {
		writer.WriteHeader(400)
		return
	}
	source := 1
	if strings.HasPrefix(request.URL.Path, "/2/") {
		source = 2
	}
	fast := source == 1
	if scenario.Swap.Load() {
		fast = !fast
	}
	outputDelay := 20 * time.Millisecond
	if !fast {
		outputDelay = 100 * time.Millisecond
	}
	tail := 5 * time.Millisecond
	if scenario.Tail {
		outputDelay = 20 * time.Millisecond
		tail = 20 * time.Millisecond
		if source == 2 {
			tail = 200 * time.Millisecond
		}
	}
	started := time.Now()
	wait := func(delay time.Duration) bool {
		timer := time.NewTimer(max(0, time.Until(started.Add(delay))))
		defer timer.Stop()
		select {
		case <-timer.C:
			return true
		case <-request.Context().Done():
			return false
		}
	}
	timing := Timing{Upstream: source}
	record := func() {
		fixture.mu.Lock()
		fixture.timings[request.Header.Get("X-Request-ID")] = timing
		fixture.mu.Unlock()
	}
	defer record()
	if input.Stream {
		writer.Header().Set("Content-Type", "text/event-stream")
	} else {
		writer.Header().Set("Content-Type", "application/json")
	}
	writer.WriteHeader(200)
	writer.(http.Flusher).Flush()
	if input.Stream && scenario.Heartbeat && source == 2 {
		if !wait(time.Millisecond) {
			return
		}
		fmt.Fprint(writer, ": controlled heartbeat\n\n")
		writer.(http.Flusher).Flush()
		timing.BodyMS = float64(time.Since(started)) / float64(time.Millisecond)
		if !wait(3 * time.Millisecond) {
			return
		}
		fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\"}}]}\n\n")
		writer.(http.Flusher).Flush()
		if !wait(5 * time.Millisecond) {
			return
		}
		fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"\"}}]}\n\n")
		writer.(http.Flusher).Flush()
	}
	if !wait(outputDelay) {
		return
	}
	timing.OutputMS = float64(time.Since(started)) / float64(time.Millisecond)
	if timing.BodyMS == 0 {
		timing.BodyMS = timing.OutputMS
	}
	if input.Stream {
		fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\n")
	} else {
		fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"`)
	}
	writer.(http.Flusher).Flush()
	if !wait(outputDelay + tail) {
		return
	}
	if input.Stream {
		fmt.Fprint(writer, "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":4,\"total_tokens\":7}}\n\ndata: [DONE]\n\n")
	} else {
		fmt.Fprint(writer, `controlled"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`)
	}
	timing.CompleteMS = float64(time.Since(started)) / float64(time.Millisecond)
}
