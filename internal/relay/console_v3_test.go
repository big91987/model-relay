package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

type consoleClient struct {
	test   *testing.T
	server *Server
	cookie *http.Cookie
	csrf   string
}

func (client *consoleClient) call(method, path string, input any, operation string) (int, map[string]any) {
	client.test.Helper()
	data, _ := json.Marshal(input)
	request := httptest.NewRequest(method, "http://"+client.server.Config.Authority+"/admin/api/"+path, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://"+client.server.Config.Authority)
	request.Header.Set("X-CSRF-Token", client.csrf)
	request.Header.Set("Idempotency-Key", operation)
	if client.cookie != nil {
		request.AddCookie(client.cookie)
	}
	writer := httptest.NewRecorder()
	if client.server.Config.Authority == "127.0.0.1:5544" {
		client.server.ServeHTTP(writer, request)
	} else {
		request.RequestURI = ""
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			client.test.Fatal("actual management HTTP failed", err)
		}
		defer response.Body.Close()
		writer.Code = response.StatusCode
		_, _ = writer.Body.ReadFrom(response.Body)
		for name, values := range response.Header {
			writer.Header()[name] = values
		}
	}
	var result map[string]any
	_ = json.Unmarshal(writer.Body.Bytes(), &result)
	if method == "POST" && path == "login" && writer.Code == 200 {
		client.cookie = writer.Result().Cookies()[0]
		client.csrf = result["csrf"].(string)
	}
	return writer.Code, result
}

func consoleInstallation(test *testing.T) (*Store, string) {
	test.Helper()
	root := test.TempDir()
	password, err := InitConsole(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err != nil {
		test.Fatal("console init", err)
	}
	store, err := Open(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err != nil {
		test.Fatal("console open", err)
	}
	test.Cleanup(func() { store.Close() })
	return store, password
}

func TestConsoleFreshInstallationEmptyCollections(t *testing.T) {
	store, password := consoleInstallation(t)
	server := New(store, Config{Authority: "127.0.0.1:5544"})
	admin := consoleClient{test: t, server: server}
	if status, _ := admin.call("POST", "login", map[string]string{"account": "admin", "password": password}, ""); status != 200 {
		t.Fatal("admin login", status)
	}
	for _, resource := range []struct {
		path  string
		field string
	}{{"upstreams", "data"}, {"models", "data"}, {"tenants", "items"}} {
		t.Run(resource.path, func(t *testing.T) {
			status, result := admin.call("GET", resource.path, nil, "")
			if status != 200 {
				t.Fatal("empty collection status", status)
			}
			values, valid := result[resource.field].([]any)
			if !valid || len(values) != 0 {
				t.Fatal("fresh console must return a non-null empty collection")
			}
		})
	}
	if status, _ := admin.call("GET", "tenants/legacy/keys", nil, ""); status != 404 {
		t.Fatal("fresh installation must not expose a migrated legacy tenant", status)
	}
}

func TestConsoleServiceCheckRevisionContract(t *testing.T) {
	store, password := consoleInstallation(t)
	server := New(store, Config{Authority: "127.0.0.1:5544"})
	admin := consoleClient{test: t, server: server}
	if status, _ := admin.call("POST", "login", map[string]string{"account": "admin", "password": password}, ""); status != 200 {
		t.Fatal("admin login", status)
	}
	status, created := admin.call("POST", "upstreams", map[string]any{"name": "controlled check", "url": "https://8.8.8.8/v1", "token": "controlled-check-only", "enabled": true, "source_kind": "provider"}, "")
	if status != 201 {
		t.Fatal("create service", status)
	}
	upstreamID := int64(created["id"].(float64))
	upstream, err := store.upstreamByID(upstreamID)
	if err != nil {
		t.Fatal(err)
	}
	visits := 0
	server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		visits++
		if request.Method != "GET" || request.URL.Path != "/v1/models" || request.Body != nil {
			t.Error("check must not generate")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"object":"list","data":[]}`))}, nil
	})
	path := "upstreams/" + fmt.Sprint(upstreamID) + "/check"
	if status, _ := admin.call("POST", path, map[string]any{}, ""); status != 400 {
		t.Fatal("missing execution revision", status)
	}
	if status, _ := admin.call("POST", path, map[string]any{"expected_execution_revision": upstream.ExecutionRevision + 1}, ""); status != 409 {
		t.Fatal("stale execution revision", status)
	}
	if visits != 0 {
		t.Fatal("rejected checks dispatched", visits)
	}
	status, result := admin.call("POST", path, map[string]any{"expected_execution_revision": upstream.ExecutionRevision}, "")
	if status != 200 || result["state"] != "success" || result["applicable"] != true || visits != 1 {
		t.Fatal("current revision check", status, visits)
	}
	var attempts int
	if err := store.DB.QueryRow("SELECT count(*) FROM request_attempts").Scan(&attempts); err != nil || attempts != 0 {
		t.Fatal("check created generation attempts", attempts, err)
	}
}

func TestConsoleIdentityAndKeys(t *testing.T) {
	store, password := consoleInstallation(t)
	server := New(store, Config{Authority: "127.0.0.1:5544"})
	admin := consoleClient{test: t, server: server}
	if status, _ := admin.call("POST", "login", map[string]string{"account": "admin", "password": password}, ""); status != 200 {
		t.Fatal("admin login", status)
	}
	status, tenant := admin.call("POST", "tenants", map[string]string{"name": "First team"}, "")
	if status != 201 {
		t.Fatal("create tenant", status)
	}
	tenantID := tenant["id"].(string)
	status, invite := admin.call("POST", "tenants/"+tenantID+"/invitations", map[string]any{"account": "owner@example.com", "role": "administrator"}, "")
	if status != 201 {
		t.Fatal("invite", status)
	}
	person := consoleClient{test: t, server: server}
	status, _ = person.call("POST", "invitations/accept", map[string]any{"token": invite["acceptance_token"], "password": "controlled-password-123"}, "")
	if status != 200 {
		t.Fatal("accept", status)
	}
	status, _ = person.call("POST", "login", map[string]any{"account": "owner@example.com", "password": "controlled-password-123"}, "")
	if status != 200 {
		t.Fatal("person login", status)
	}
	if status, _ := person.call("GET", "upstreams", nil, ""); status != 403 {
		t.Fatal("global leak", status)
	}
	if status, _ := person.call("GET", "tenants/legacy/members", nil, ""); status != 404 {
		t.Fatal("tenant leak", status)
	}
	keyInput := KeyInput{Name: "Application", Application: "test app", Enabled: true, Models: []string{}, Diagnostics: []string{}}
	status, issued := person.call("POST", "tenants/"+tenantID+"/keys", keyInput, "issue-1")
	if status != 201 {
		t.Fatal("issue", status)
	}
	keyID := issued["id"].(string)
	secret := issued["key"].(string)
	status, repeated := person.call("POST", "tenants/"+tenantID+"/keys", keyInput, "issue-1")
	if status != 200 || repeated["secret_recoverable"] != false || repeated["key"] != nil {
		t.Fatal("idempotent issuance leaked/reissued")
	}
	status, metadata := person.call("GET", "tenants/"+tenantID+"/keys/"+keyID, nil, "")
	if status != 200 || metadata["key"] != nil {
		t.Fatal("metadata leak")
	}
	request := httptest.NewRequest("GET", "http://127.0.0.1:5544/v1/models", nil)
	request.Header.Set("Authorization", "Bearer "+secret)
	if actual, ok := server.appID(request); !ok || actual != keyID {
		t.Fatal("key auth")
	}
	status, rotation := person.call("POST", "tenants/"+tenantID+"/keys/"+keyID+"/rotations", map[string]any{"expected_revision": 1, "transition_hours": 1}, "rotate-1")
	if status != 200 {
		t.Fatal("rotation", status)
	}
	newSecret := rotation["key"].(string)
	if _, ok := server.appID(request); !ok {
		t.Fatal("old grace")
	}
	request.Header.Set("Authorization", "Bearer "+newSecret)
	if actual, ok := server.appID(request); !ok || actual != keyID {
		t.Fatal("new generation pool")
	}
	status, _ = person.call("POST", "tenants/"+tenantID+"/keys/"+keyID+"/rotations", map[string]any{"expected_revision": 2, "transition_hours": 1}, "rotate-2")
	if status != 409 {
		t.Fatal("stacked transition", status)
	}
	status, _ = person.call("POST", "tenants/"+tenantID+"/keys/"+keyID+"/revoke", map[string]any{"expected_revision": 2}, "revoke-1")
	if status != 200 {
		t.Fatal("revoke", status)
	}
	if _, ok := server.appID(request); ok {
		t.Fatal("revoked key authorized")
	}
	identityStatus, identity := person.call("GET", "identity", nil, "")
	if identityStatus != 200 {
		t.Fatal("identity")
	}
	status, _ = person.call("DELETE", "tenants/"+tenantID+"/members/"+identity["id"].(string), nil, "")
	if status != 409 {
		t.Fatal("last administrator", status)
	}
}

func TestConsoleKeyCollectionContract(t *testing.T) {
	store, password := consoleInstallation(t)
	admin := consoleClient{test: t, server: New(store, Config{Authority: "127.0.0.1:5544"})}
	if status, _ := admin.call("POST", "login", map[string]string{"account": "admin", "password": password}, ""); status != 200 {
		t.Fatal("login status", status)
	}
	status, tenant := admin.call("POST", "tenants", map[string]string{"name": "Collection contract"}, "")
	if status != 201 {
		t.Fatal("tenant status", status)
	}
	path := "tenants/" + tenant["id"].(string) + "/keys"
	for _, enabled := range []bool{true, false} {
		name := "migration-disabled"
		if enabled {
			name = "migration-enabled"
		}
		if status, _ := admin.call("POST", path, KeyInput{Name: name, Application: "controlled", Enabled: enabled}, name); status != 201 {
			t.Fatal("key create status", status)
		}
	}
	status, body := admin.call("GET", path, nil, "")
	items, valid := body["items"].([]any)
	if status != 200 || !valid || len(items) != 2 || body["total"] != float64(2) || body["page"] != float64(1) || body["page_size"] != float64(25) {
		t.Fatal("Key collection envelope mismatch; values withheld")
	}
	if _, obsolete := body["data"]; obsolete {
		t.Fatal("obsolete Key collection field")
	}
	identities := map[string]bool{}
	for _, item := range items {
		key := item.(map[string]any)
		name, valid := key["name"].(string)
		if !valid || identities[name] || name != "migration-enabled" && name != "migration-disabled" {
			t.Fatal("Key identity mismatch; values withheld")
		}
		identities[name] = true
		if key["enabled"] != (key["name"] == "migration-enabled") {
			t.Fatal("Key enabled state mismatch; values withheld")
		}
	}
}

func TestMicroDollarPrecision(t *testing.T) {
	for _, test := range []struct{ input, output, inputRate, outputRate, want int64 }{{1, 0, 499999, 0, 0}, {1, 0, 500000, 0, 1}, {1, 1, 500000, 500000, 1}, {0, 0, 1000000, 1000000, 0}, {1000000, 1000000, 1, 3, 4}} {
		actual, err := microCharge(test.input, test.output, test.inputRate, test.outputRate)
		if err != nil || actual != test.want {
			t.Fatal("micro precision", actual, err)
		}
	}
	if _, err := microCharge(-1, 0, 0, 0); err == nil {
		t.Fatal("negative accepted")
	}
	if _, err := microCharge(9223372036854775807, 0, 9223372036854775807, 0); err == nil {
		t.Fatal("overflow accepted")
	}
}
