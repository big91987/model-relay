package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func multiConfiguration(t *testing.T, value *consoleFixture, count int, weights []int) {
	t.Helper()
	for index := 0; index < count; index++ {
		input := map[string]any{"name": string(rune('A' + index)), "url": "https://public.example/" + string(rune('a'+index)) + "/v1", "token": "isolated-route-token", "enabled": true}
		if response := value.admin("POST", "upstreams", input); response.Code != 201 {
			t.Fatal("create upstream", response.Code, response.Body.String())
		}
	}
	candidates := []map[string]any{}
	for index := 0; index < count; index++ {
		weight := 1
		if weights != nil {
			weight = weights[index]
		}
		candidates = append(candidates, map[string]any{"upstream_id": index + 1, "remote_model": "remote-" + string(rune('a'+index)), "enabled": true, "priority": 0, "weight": weight})
	}
	if response := value.admin("POST", "models", map[string]any{"alias": "alpha", "enabled": true, "candidates": candidates}); response.Code != 201 {
		t.Fatal("create model", response.Code, response.Body.String())
	}
	response := value.admin("POST", "keys", map[string]string{"name": "routing"})
	var issued struct {
		Key      string
		Metadata AppKey
	}
	if response.Code != 201 || json.Unmarshal(response.Body.Bytes(), &issued) != nil {
		t.Fatal("key")
	}
	value.key = issued.Key
	value.keyID = issued.Metadata.ID
}

func upstreamJSON() *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"controlled"}}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`))}
}

func TestRoutingSelectionAndEvidence(t *testing.T) {
	t.Parallel()
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 2, []int{3, 1})
	counts := map[string]int{}
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		counts[request.URL.Path]++
		body, _ := io.ReadAll(request.Body)
		if strings.Contains(request.URL.Path, "/a/") && !strings.Contains(string(body), `"model":"remote-a"`) || strings.Contains(request.URL.Path, "/b/") && !strings.Contains(string(body), `"model":"remote-b"`) {
			t.Error("mapping not real")
		}
		return upstreamJSON(), nil
	})
	for index := 0; index < 16; index++ {
		response := value.call(value.key, ordinary)
		if response.Code != 200 {
			t.Fatal("call", response.Code, response.Body.String())
		}
		detail := value.admin("GET", "requests/"+response.Header().Get("X-Request-ID"), nil)
		var record Record
		if json.Unmarshal(detail.Body.Bytes(), &record) != nil || record.EvidenceVersion != 2 || len(record.Attempts) != 1 || !record.AttemptsComplete || record.Total == nil || *record.Total != 3 {
			t.Fatal("attempt evidence", detail.Body.String())
		}
	}
	if counts["/a/v1/chat/completions"] != 12 || counts["/b/v1/chat/completions"] != 4 {
		t.Fatal("weighted visits", counts)
	}
	model, _ := value.server.Store.routeModel("alpha")
	candidates := cleanTestCandidates(model.Candidates)
	candidates[1]["priority"] = 100
	if response := value.admin("PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": candidates}); response.Code != 200 {
		t.Fatal("priority update")
	}
	for index := 0; index < 4; index++ {
		if value.call(value.key, ordinary).Code != 200 {
			t.Fatal("priority call")
		}
	}
	if counts["/b/v1/chat/completions"] != 8 {
		t.Fatal("larger priority did not win", counts)
	}
	if response := value.admin("PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": candidates}); response.Code != 409 {
		t.Fatal("lost update accepted")
	}
}

func cleanTestCandidates(values []Candidate) []map[string]any {
	result := []map[string]any{}
	for _, value := range values {
		result = append(result, map[string]any{"upstream_id": value.UpstreamID, "remote_model": value.RemoteModel, "enabled": value.Enabled, "priority": value.Priority, "weight": value.Weight})
	}
	return result
}

func TestRoutingTemporaryAllowlistAndUnknownTotals(t *testing.T) {
	for _, status := range []int{301, 400, 401, 403, 404, 408, 429, 500, 501, 502, 503, 504, 505} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()
			value := newConsoleFixture(t)
			multiConfiguration(t, value, 2, nil)
			calls := 0
			value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
				calls++
				if strings.Contains(request.URL.Path, "/a/") {
					return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("sensitive supplier error"))}, nil
				}
				return upstreamJSON(), nil
			})
			response := value.call(value.key, ordinary)
			want := 1
			if temporaryStatus(status) {
				want = 2
				if response.Code != 200 {
					t.Fatal("did not failover", response.Code, response.Body.String())
				}
			}
			if calls != want {
				t.Fatal("attempt allowlist", status, calls)
			}
			detail := value.admin("GET", "requests/"+response.Header().Get("X-Request-ID"), nil)
			var record Record
			json.Unmarshal(detail.Body.Bytes(), &record)
			if len(record.Attempts) != want || strings.Contains(detail.Body.String(), "sensitive supplier error") {
				t.Fatal("attempt privacy or count")
			}
			if want == 2 && (record.Total != nil || record.UsageSummary["total_tokens"].KnownSum != 3 || record.UsageSummary["total_tokens"].UnknownAttempts != 1) {
				t.Fatal("unknown consumption hidden", detail.Body.String())
			}
		})
	}
}

func TestRoutingBoundedAndStop(t *testing.T) {
	t.Parallel()
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 4, nil)
	var calls atomic.Int64
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 503, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("error"))}, nil
	})
	response := value.call(value.key, ordinary)
	if response.Code != 502 || !strings.Contains(response.Body.String(), "upstream_attempts_exhausted") || calls.Load() != 3 {
		t.Fatal("unbounded attempts", response.Code, calls.Load())
	}
	model, _ := value.server.Store.routeModel("alpha")
	var attempts int
	if value.server.Store.DB.QueryRow("SELECT count(*) FROM request_attempts WHERE request_id=?", response.Header().Get("X-Request-ID")).Scan(&attempts) != nil || attempts != 3 {
		t.Fatal("attempt count")
	}
	_ = model
	value = newConsoleFixture(t)
	multiConfiguration(t, value, 2, nil)
	calls.Store(0)
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader("data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\n"))}, nil
	})
	response = value.call(value.key, streaming)
	if calls.Load() != 1 || !strings.Contains(response.Body.String(), "first") || strings.Contains(response.Body.String(), "[DONE]") {
		t.Fatal("replayed committed stream")
	}
}

func TestRoutingPassiveCooldown(t *testing.T) {
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 2, nil)
	value.server.Config.Cooldown = 10 * time.Millisecond
	var clock atomic.Int64
	clock.Store(time.Now().UnixNano())
	value.server.latencyClock = func() time.Time { return time.Unix(0, clock.Load()) }
	cooldownStarted := clock.Load()
	counts := map[string]int{}
	failed := true
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		counts[request.URL.Path]++
		if failed && strings.Contains(request.URL.Path, "/a/") {
			return &http.Response{StatusCode: 503, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("error"))}, nil
		}
		return upstreamJSON(), nil
	})
	if value.call(value.key, ordinary).Code != 200 {
		t.Fatal("initial recovery")
	}
	failed = false
	clock.Store(cooldownStarted + int64(9*time.Millisecond))
	if value.call(value.key, ordinary).Code != 200 || counts["/a/v1/chat/completions"] != 1 {
		t.Fatal("cooling candidate visited")
	}
	clock.Store(cooldownStarted + int64(10*time.Millisecond))
	if value.call(value.key, ordinary).Code != 200 || counts["/a/v1/chat/completions"] != 2 {
		t.Fatal("passive recovery missing")
	}
}
