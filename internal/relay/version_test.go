package relay

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthVersionContract(t *testing.T) {
	for _, version := range []string{"dev", "v1.2.3-build", "  raw-version  "} {
		t.Run(version, func(t *testing.T) {
			store, _, _, _ := installation(t)
			server := New(store, Config{Version: version, Authority: "example.com"})
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, httptest.NewRequest("GET", "/healthz", nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("health status %d", recorder.Code)
			}
			var body struct {
				Status, Version string
				Schema          int `json:"storage_schema"`
			}
			decode(t, recorder.Result(), &body)
			if body.Status != "ok" || body.Version != version || body.Schema != 2 {
				t.Fatalf("health contract mismatch: %#v", body)
			}
		})
	}
}

func TestHTTPHealthVersion(t *testing.T) {
	for _, version := range []string{"dev", "v1.2.3-build"} {
		t.Run(version, func(t *testing.T) {
			store, _, _, _ := installation(t)
			service := localServer(t, New(store, Config{Version: version}))
			response, err := http.Get(service.URL + "/healthz")
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != http.StatusOK {
				t.Fatalf("health status %d", response.StatusCode)
			}
			var body struct{ Version string }
			decode(t, response, &body)
			if body.Version != version {
				t.Fatalf("health version mismatch: %q", body.Version)
			}
		})
	}
}
