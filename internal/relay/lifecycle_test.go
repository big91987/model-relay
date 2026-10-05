package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestHTTPConfigurationChangesAndReenable(t *testing.T) {
	expect := func(t *testing.T, response *http.Response, status int) {
		t.Helper()
		_, readErr := io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if readErr != nil || response.StatusCode != status {
			t.Fatalf("configuration lifecycle status %d, wanted %d (body omitted)", response.StatusCode, status)
		}
	}
	var original, replacement atomic.Int64
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		var input struct{ Model string }
		if json.NewDecoder(request.Body).Decode(&input) != nil {
			t.Error("controlled request invalid")
		}
		switch request.URL.Path {
		case "/v1/chat/completions":
			expectedModel := "provider-model"
			if original.Add(1) > 2 {
				expectedModel = "edited-provider-model"
			}
			if request.Header.Get("Authorization") != "Bearer upstream-secret-sentinel" || input.Model != expectedModel {
				t.Error("original mapping or credential incorrect")
			}
		case "/replacement/chat/completions":
			replacement.Add(1)
			if request.Header.Get("Authorization") != "Bearer replacement-token-sentinel" || input.Model != "edited-provider-model" {
				t.Error("updated mapping or credential not used")
			}
		default:
			t.Error("unexpected controlled destination")
		}
		normal(writer, request)
	})
	expect(t, value.call(t, ordinary), 200)
	expect(t, value.admin(t, "PUT", "upstream", map[string]any{"url": value.upstream.URL + "/replacement", "enabled": true}), 400)
	expect(t, value.call(t, ordinary), 200)
	if original.Load() != 2 || replacement.Load() != 0 {
		t.Fatal("rejected address change modified destination")
	}
	expect(t, value.admin(t, "PATCH", "models/alpha", map[string]any{"model": "edited-provider-model", "enabled": true}), 200)
	expect(t, value.call(t, ordinary), 200)
	if original.Load() != 3 || replacement.Load() != 0 {
		t.Fatal("mapping edit unexpectedly changed destination or duplicated call")
	}
	expect(t, value.admin(t, "PUT", "upstream", map[string]any{"url": value.upstream.URL + "/replacement", "token": "replacement-token-sentinel", "enabled": true}), 200)
	expect(t, value.call(t, ordinary), 200)
	if original.Load() != 3 || replacement.Load() != 1 {
		t.Fatal("updated request did not exclusively use new destination")
	}
	expect(t, value.admin(t, "PUT", "upstream", map[string]any{"url": value.upstream.URL + "/replacement", "enabled": false}), 200)
	expect(t, value.call(t, ordinary), 503)
	models := func() []struct{ ID string } {
		request, _ := http.NewRequest("GET", value.http.URL+"/v1/models", nil)
		request.Header.Set("Authorization", "Bearer "+value.key)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != 200 {
			t.Fatal("model query failed")
		}
		var result struct{ Data []struct{ ID string } }
		decode(t, response, &result)
		return result.Data
	}
	if len(models()) != 0 || original.Load() != 3 || replacement.Load() != 1 {
		t.Fatal("disabled upstream remained available or consumed a call")
	}
	expect(t, value.admin(t, "PUT", "upstream", map[string]any{"url": value.upstream.URL + "/replacement", "enabled": true}), 200)
	available := models()
	if len(available) != 1 || available[0].ID != "alpha" {
		t.Fatal("reenabled model list did not recover")
	}
	expect(t, value.call(t, ordinary), 200)
	if original.Load() != 3 || replacement.Load() != 2 {
		t.Fatal("reenabled request did not use retained new configuration exactly once")
	}
}
