package httpconsole

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/routing"
	"model-relay/internal/upstream"
)

func TestCatalogRoutesRejectInvalidPathsAndPayloadsBeforeServiceAccess(t *testing.T) {
	controller := &Controller{}
	for _, test := range []struct {
		method, path, body string
		status             int
		code               string
	}{
		{"GET", "/admin/api/upstreams?limit=0", "", 400, "invalid_cursor"},
		{"GET", "/admin/api/upstreams?cursor=0", "", 400, "invalid_cursor"},
		{"GET", "/admin/api/upstreams/0", "", 404, "not_found"},
		{"GET", "/admin/api/upstreams/1/check", "", 405, "method_not_allowed"},
		{"POST", "/admin/api/upstreams/1/check", "{}", 400, "invalid_request"},
		{"POST", "/admin/api/upstreams", "{", 400, "invalid_request"},
		{"GET", "/admin/api/models/", "", 404, "not_found"},
		{"GET", "/admin/api/models?cursor=bad%2Falias", "", 400, "invalid_cursor"},
		{"GET", "/admin/api/models?limit=101", "", 400, "invalid_cursor"},
		{"POST", "/admin/api/models/alpha", "{}", 405, "method_not_allowed"},
		{"POST", "/admin/api/models", "{", 400, "invalid_request"},
	} {
		t.Run(test.method+test.path, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			if strings.HasPrefix(request.URL.Path, "/admin/api/upstreams") {
				controller.upstreams(recorder, request)
			} else {
				controller.routeModels(recorder, request)
			}
			if recorder.Code != test.status {
				t.Fatalf("status %d, want %d: %s", recorder.Code, test.status, recorder.Body.String())
			}
			var value struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil || value.Error.Code != test.code {
				t.Fatalf("error %s, want %s: %v", value.Error.Code, test.code, err)
			}
		})
	}
}
func TestCatalogAdaptersPreserveOrderedPaginationAndNullExhaustion(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "catalog-http.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(`CREATE TABLE upstream(id INTEGER PRIMARY KEY,name TEXT,url TEXT,enabled INTEGER,revision INTEGER,execution_revision INTEGER,token BLOB);
 CREATE TABLE upstream_checks(id INTEGER PRIMARY KEY,upstream_id INTEGER,execution_revision INTEGER,state TEXT,started_at TEXT,finished_at TEXT,duration_ms INTEGER,error_code TEXT,upstream_status INTEGER);
 CREATE TABLE models(alias TEXT PRIMARY KEY,enabled INTEGER,revision INTEGER,legacy_remote_model TEXT);
 CREATE TABLE model_route_policies(alias TEXT PRIMARY KEY,strategy TEXT);
 CREATE TABLE route_candidates(alias TEXT,upstream_id INTEGER,remote_model TEXT,enabled INTEGER,priority INTEGER,weight INTEGER,execution_revision INTEGER);
 INSERT INTO upstream VALUES(1,'one','http://example.test',1,1,1,x'01'),(2,'two','http://example.test',1,1,1,x'01');
 INSERT INTO models VALUES('alpha',1,1,NULL),('beta',1,1,NULL);`); err != nil {
		t.Fatal(err)
	}
	controller := &Controller{Catalog: catalog.New(db, make([]byte, 32), 2, configstate.New(), routing.New(), nil, upstream.Policy{})}
	for _, test := range []struct {
		path, next string
		field      string
		want       string
	}{
		{"/admin/api/upstreams?limit=1", "1", "name", "one"}, {"/admin/api/upstreams?limit=1&cursor=1", "", "name", "two"},
		{"/admin/api/models?limit=1", "alpha", "alias", "alpha"}, {"/admin/api/models?limit=1&cursor=alpha", "", "alias", "beta"},
	} {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			recorder := httptest.NewRecorder()
			if strings.HasPrefix(request.URL.Path, "/admin/api/upstreams") {
				controller.upstreams(recorder, request)
			} else {
				controller.routeModels(recorder, request)
			}
			if recorder.Code != 200 {
				t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
			}
			var value struct {
				Data []map[string]any `json:"data"`
				Next *string          `json:"next_cursor"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
				t.Fatal(err)
			}
			if len(value.Data) != 1 || value.Data[0][test.field] != test.want {
				t.Fatalf("page facts changed: %+v", value)
			}
			if test.next == "" {
				if value.Next != nil {
					t.Fatal("exhausted cursor must remain null")
				}
			} else if value.Next == nil || *value.Next != test.next {
				t.Fatalf("wrong cursor: %+v", value.Next)
			}
		})
	}
}
