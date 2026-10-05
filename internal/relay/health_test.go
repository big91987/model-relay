package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthMethodsHTTP(t *testing.T) {
	testHealthMethods(t, true)
}

func TestHealthMethodsHandler(t *testing.T) {
	testHealthMethods(t, false)
}

func testHealthMethods(t *testing.T, realHTTP bool) {
	t.Helper()
	for _, state := range []string{"healthy", "degraded", "database_closed"} {
		t.Run(state, func(t *testing.T) {
			store, _, directory, keyFile := installation(t)
			server := New(store, Config{Authority: "127.0.0.1:5544", Version: "health-regression"})
			address := "http://" + server.Config.Authority
			client := &http.Client{Timeout: 5 * time.Second}
			if realHTTP {
				address = localServer(t, server).URL
			}
			if state == "degraded" {
				server.degraded.Store(true)
			}
			if state == "database_closed" {
				if err := store.DB.Close(); err != nil {
					t.Fatal(err)
				}
			}
			check := func(t *testing.T, method string, status int) {
				t.Helper()
				request, err := http.NewRequest(method, address+"/healthz", nil)
				if err != nil {
					t.Fatal(err)
				}
				var response *http.Response
				if realHTTP {
					response, err = client.Do(request)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					writer := httptest.NewRecorder()
					server.ServeHTTP(writer, request)
					response = writer.Result()
				}
				defer response.Body.Close()
				body, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatal(err)
				}
				if response.StatusCode != status {
					t.Errorf("%s status = %d, want %d", method, response.StatusCode, status)
				}
				if status == http.StatusMethodNotAllowed {
					if allow := response.Header.Get("Allow"); allow != "GET, HEAD" {
						t.Errorf("%s Allow = %q, want GET, HEAD", method, allow)
					}
					return
				}
				if response.Header.Get("Content-Type") != "application/json" {
					t.Error("health content type must remain application/json")
				}
				if method == http.MethodHead {
					if len(body) != 0 {
						t.Errorf("HEAD body has %d bytes, want zero", len(body))
					}
					return
				}
				if status == http.StatusOK {
					var payload map[string]string
					if err := json.Unmarshal(body, &payload); err != nil {
						t.Fatal(err)
					}
					if len(payload) != 2 || payload["status"] != "ok" || payload["version"] != "health-regression" {
						t.Error("GET status/version contract changed")
					}
				} else {
					var payload struct {
						Error struct{ Code string }
					}
					if err := json.Unmarshal(body, &payload); err != nil {
						t.Fatal(err)
					}
					if payload.Error.Code != "storage_unavailable" {
						t.Error("GET storage failure contract changed")
					}
				}
			}
			status := http.StatusOK
			if state != "healthy" {
				status = http.StatusServiceUnavailable
			}
			for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodTrace, "CUSTOM"} {
				t.Run(method, func(t *testing.T) {
					expected := status
					if method != http.MethodGet && method != http.MethodHead {
						expected = http.StatusMethodNotAllowed
					}
					check(t, method, expected)
				})
			}
			if state == "database_closed" {
				if err := store.Close(); err != nil {
					t.Fatal(err)
				}
				reopened, err := Open(directory, keyFile)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { reopened.Close() })
				server.Store = reopened
			}
			server.degraded.Store(false)
			t.Run("recovered_GET", func(t *testing.T) { check(t, http.MethodGet, http.StatusOK) })
			t.Run("recovered_HEAD", func(t *testing.T) { check(t, http.MethodHead, http.StatusOK) })
		})
	}
}
