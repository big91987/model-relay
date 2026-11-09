package httpconsole

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/identity"
	"model-relay/internal/sessions"
)

func identityController(t *testing.T) (*Controller, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "identity-http.db")+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(`CREATE TABLE users(id TEXT PRIMARY KEY,account TEXT,platform INTEGER,active INTEGER,salt BLOB,hash BLOB);
 CREATE TABLE tenants(id TEXT PRIMARY KEY,name TEXT,active INTEGER,created TEXT);
 CREATE TABLE memberships(tenant_id TEXT,user_id TEXT,role TEXT,PRIMARY KEY(tenant_id,user_id));
 CREATE TABLE invitations(id TEXT PRIMARY KEY,tenant_id TEXT,account TEXT,role TEXT,digest BLOB,expires TEXT,state TEXT,accepted_by TEXT);
 CREATE TABLE console_audit(id TEXT,time TEXT,actor TEXT,tenant_id TEXT,action TEXT,object_id TEXT);
 INSERT INTO users VALUES('admin','admin',1,1,x'00',x'00'),('developer','developer',0,1,x'00',x'00');
 INSERT INTO tenants VALUES('tenant','Tenant',1,'2026-01-01');
 INSERT INTO memberships VALUES('tenant','admin','administrator'),('tenant','developer','developer');
 INSERT INTO invitations VALUES('invitation','tenant','invitee@example.test','finance',x'00','2027-01-01T00:00:00Z','pending',NULL);`); err != nil {
		t.Fatal(err)
	}
	manager := sessions.New()
	for _, id := range []string{"admin", "developer"} {
		if _, ok := manager.Create(time.Now(), "session-"+id, id, "csrf"); !ok {
			t.Fatal("session fixture rejected")
		}
	}
	controller := &Controller{Schema: 3, Authority: func() string { return "controlled.test" }, Now: time.Now, AdvanceClock: func() error { return nil }, Degraded: &atomic.Bool{}, State: configstate.New(), Sessions: manager, Identity: identity.New(db, time.Now)}
	return controller, db
}
func facadeRequest(controller *Controller, method, path, body, user, origin, csrf string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if user != "" {
		request.AddCookie(&http.Cookie{Name: "mr_session", Value: "session-" + user})
	}
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	if csrf != "" {
		request.Header.Set("X-CSRF-Token", csrf)
	}
	response := httptest.NewRecorder()
	controller.ServeHTTP(response, request)
	return response
}
func assertHTTPError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	var value struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil || response.Code != status || value.Error.Code != code {
		t.Fatalf("HTTP got %d %s, want %d/%s (%v)", response.Code, response.Body.String(), status, code, err)
	}
}
func TestFacadePreservesOriginSessionCSRFAndIdentityAuthorizationPrecedence(t *testing.T) {
	controller, _ := identityController(t)
	for _, test := range []struct {
		method, path, user, origin, csrf string
		status                           int
		code                             string
	}{
		{"POST", "/admin/api/tenants", "", "", "", 403, "untrusted_origin"},
		{"POST", "/admin/api/tenants", "", "http://controlled.test", "", 401, "unauthorized"},
		{"POST", "/admin/api/tenants", "developer", "http://controlled.test", "wrong", 403, "invalid_csrf"},
		{"POST", "/admin/api/tenants", "developer", "http://controlled.test", "csrf", 403, "forbidden"},
		{"POST", "/admin/api/tenants/tenant/invitations", "developer", "http://controlled.test", "csrf", 403, "forbidden"},
		{"PATCH", "/admin/api/tenants/missing/members/developer", "admin", "http://controlled.test", "csrf", 404, "not_found"},
		{"PATCH", "/admin/api/tenants/tenant/members/missing", "admin", "http://controlled.test", "csrf", 404, "not_found"},
		{"PATCH", "/admin/api/tenants/tenant/members/developer", "admin", "http://controlled.test", "csrf", 400, "invalid_role"},
	} {
		t.Run(test.method+test.path+test.user+test.csrf, func(t *testing.T) {
			assertHTTPError(t, facadeRequest(controller, test.method, test.path, "{", test.user, test.origin, test.csrf), test.status, test.code)
		})
	}
}
func TestIdentityUnsupportedRoutesRetainTheirOriginalScopedAuthorization(t *testing.T) {
	controller, _ := identityController(t)
	for _, test := range []struct {
		method, path, user string
		status             int
		code               string
	}{
		{"DELETE", "/admin/api/tenants/tenant", "developer", 403, "forbidden"},
		{"GET", "/admin/api/tenants/tenant/invitations/unknown", "developer", 403, "forbidden"},
		{"POST", "/admin/api/tenants/tenant/invitations/invitation", "developer", 403, "forbidden"},
		{"GET", "/admin/api/tenants/tenant/members/unknown", "developer", 404, "not_found"},
		{"POST", "/admin/api/tenants/tenant/members/unknown/extra", "developer", 403, "forbidden"},
		{"POST", "/admin/api/tenants/tenant/invitations/invitation", "admin", 404, "not_found"},
		{"GET", "/admin/api/tenants/missing/invitations/unknown", "admin", 404, "not_found"},
	} {
		t.Run(test.method+test.path+test.user, func(t *testing.T) {
			assertHTTPError(t, facadeRequest(controller, test.method, test.path, "{}", test.user, "http://controlled.test", "csrf"), test.status, test.code)
		})
	}
}
func TestFacadeKeepsTheSupportedInvitationRevokeRouteAndEmptyBody(t *testing.T) {
	controller, db := identityController(t)
	path := "/admin/api/tenants/tenant/invitations/invitation/revoke"
	assertHTTPError(t, facadeRequest(controller, "POST", path, "{", "admin", "http://controlled.test", "csrf"), 400, "invalid_request")
	response := facadeRequest(controller, "POST", path, "{}", "admin", "http://controlled.test", "csrf")
	if response.Code != 200 || strings.TrimSpace(response.Body.String()) != `{"ok":true}` {
		t.Fatalf("revoke result changed: %d %s", response.Code, response.Body.String())
	}
	var state string
	var audits int
	if err := db.QueryRow("SELECT state FROM invitations WHERE id='invitation'").Scan(&state); err != nil || state != "revoked" {
		t.Fatalf("revocation did not persist: %q %v", state, err)
	}
	if err := db.QueryRow("SELECT count(*) FROM console_audit WHERE action='invitation.revoke' AND actor='admin'").Scan(&audits); err != nil || audits != 1 {
		t.Fatalf("revocation audit mismatch: %d %v", audits, err)
	}
}
func TestRevisionConflictRetainsOriginalResponseShape(t *testing.T) {
	response := httptest.NewRecorder()
	(&Controller{}).fail(response, &catalog.RevisionConflict{CurrentRevision: 7})
	var value map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
		t.Fatal(err)
	}
	if response.Code != 409 || len(value) != 2 || value["current_revision"] != float64(7) {
		t.Fatalf("conflict response shape changed: %d %s", response.Code, response.Body.String())
	}
	failure := value["error"].(map[string]any)
	if len(failure) != 3 || failure["type"] != "relay_error" || failure["code"] != "configuration_conflict" || failure["message"] != "Configuration changed; reload before editing" {
		t.Fatalf("conflict response changed: %v", failure)
	}
}
