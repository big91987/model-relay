package relay

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRoutingRealHTTPCommittedToolsStopBoundaries(testContext *testing.T) {
	for _, phase := range []string{"eof", "idle", "cancel"} {
		testContext.Run(phase, func(testContext *testing.T) {
			var secondary atomic.Int64
			value := newFixture(testContext, func(writer http.ResponseWriter, request *http.Request) {
				if strings.HasPrefix(request.URL.Path, "/secondary/") {
					secondary.Add(1)
				}
				writer.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(writer, "data: {\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"controlled-tool\",\"type\":\"function\",\"function\":{\"name\":\"controlled\",\"arguments\":\"{\"}}]}}]}\n\n")
				writer.(http.Flusher).Flush()
				if phase != "eof" {
					<-request.Context().Done()
				}
			})
			configureHTTPPair(testContext, value, 1)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			request, _ := http.NewRequestWithContext(ctx, "POST", value.http.URL+"/v1/chat/completions", strings.NewReader(streaming))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+value.key)
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				testContext.Fatal("tool dispatch", err)
			}
			reader := bufio.NewReader(response.Body)
			first, err := reader.ReadString('\n')
			if err != nil || !strings.Contains(first, "tool_calls") {
				testContext.Fatal("tool delta not committed")
			}
			if phase == "cancel" {
				cancel()
			}
			rest, _ := io.ReadAll(reader)
			response.Body.Close()
			if strings.Contains(string(rest), "[DONE]") || secondary.Load() != 0 {
				testContext.Fatal("tool replay or false completion")
			}
			var detail Record
			for poll := 0; poll < 100; poll++ {
				decode(testContext, value.admin(testContext, "GET", "requests/"+response.Header.Get("X-Request-ID"), nil), &detail)
				if detail.State != "started" {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			want := "stream_interrupted"
			if phase == "idle" {
				want = "timeout"
			}
			if phase == "cancel" {
				want = "canceled"
			}
			if len(detail.Attempts) != 1 || detail.State != want || secondary.Load() != 0 {
				testContext.Fatal("tool stop evidence", phase, detail.State)
			}
		})
	}
}

func configureHTTPPair(t *testing.T, value *fixture, weight int) {
	t.Helper()
	response := value.admin(t, "POST", "upstreams", map[string]any{"name": "controlled B", "url": value.upstream.URL + "/secondary/v1", "token": "second-isolated-token", "enabled": true})
	expect(t, response, 201)
	response = value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": 1, "enabled": true, "candidates": []map[string]any{{"upstream_id": 1, "remote_model": "provider-model", "enabled": true, "priority": 0, "weight": weight}, {"upstream_id": 2, "remote_model": "second-remote", "enabled": true, "priority": 0, "weight": 1}}})
	expect(t, response, 200)
}

func TestRoutingRealHTTPSelectionFailoverAndRecovery(t *testing.T) {
	var primary, secondary atomic.Int64
	var failPrimary atomic.Bool
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		if strings.HasPrefix(request.URL.Path, "/secondary/") {
			secondary.Add(1)
			if request.Header.Get("Authorization") != "Bearer second-isolated-token" || !strings.Contains(string(body), `"model":"second-remote"`) {
				t.Error("secondary credential or mapping")
				writer.WriteHeader(400)
				return
			}
		} else {
			primary.Add(1)
			if request.Header.Get("Authorization") != "Bearer upstream-secret-sentinel" || !strings.Contains(string(body), `"model":"provider-model"`) {
				t.Error("primary credential or mapping")
				writer.WriteHeader(400)
				return
			}
			if failPrimary.Load() {
				writer.WriteHeader(503)
				return
			}
		}
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled actual HTTP"}}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`)
	})
	configureHTTPPair(t, value, 3)
	var wait sync.WaitGroup
	for index := 0; index < 40; index++ {
		wait.Add(1)
		go func() { defer wait.Done(); response := value.call(t, ordinary); expect(t, response, 200) }()
	}
	wait.Wait()
	if primary.Load() != 30 || secondary.Load() != 10 {
		t.Fatal("actual SWRR visits", primary.Load(), secondary.Load())
	}
	value.server.Config.Cooldown = 2 * time.Second
	failPrimary.Store(true)
	response := value.call(t, ordinary)
	if response.StatusCode != 200 {
		t.Fatal("actual failover")
	}
	id := response.Header.Get("X-Request-ID")
	response.Body.Close()
	var record Record
	decode(t, value.admin(t, "GET", "requests/"+id, nil), &record)
	if len(record.Attempts) != 2 || record.Attempts[0].UpstreamID != 1 || record.Attempts[1].UpstreamID != 2 || record.Total != nil || record.UsageSummary["total_tokens"].KnownSum != 3 {
		t.Fatal("actual attempt accounting", record)
	}
	before := primary.Load()
	expect(t, value.call(t, ordinary), 200)
	if primary.Load() != before {
		t.Fatal("cooldown did not exclude actual primary")
	}
	failPrimary.Store(false)
	time.Sleep(value.server.Config.Cooldown + 25*time.Millisecond)
	expect(t, value.call(t, ordinary), 200)
	if primary.Load() != before+1 {
		t.Fatal("primary did not passively recover")
	}
}

func TestRoutingRealHTTPTemporaryAllowlist(t *testing.T) {
	for _, status := range []int{301, 400, 401, 403, 404, 408, 429, 500, 501, 502, 503, 504, 505} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()
			var primary, secondary atomic.Int64
			value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
				if strings.HasPrefix(request.URL.Path, "/secondary/") {
					secondary.Add(1)
					writer.Header().Set("Content-Type", "application/json")
					fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`)
					return
				}
				primary.Add(1)
				writer.WriteHeader(status)
			})
			configureHTTPPair(t, value, 1)
			response := value.call(t, ordinary)
			response.Body.Close()
			wanted := int64(0)
			if temporaryStatus(status) {
				wanted = 1
			}
			if primary.Load() != 1 || secondary.Load() != wanted || wanted == 1 && response.StatusCode != 200 {
				t.Fatal("real HTTP switch policy", response.StatusCode, primary.Load(), secondary.Load())
			}
		})
	}
}

func TestRoutingRealHTTPFourCandidatesStopAfterThree(t *testing.T) {
	var calls atomic.Int64
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) { calls.Add(1); writer.WriteHeader(503) })
	candidates := []map[string]any{{"upstream_id": 1, "remote_model": "provider-model", "enabled": true, "priority": 0, "weight": 1}}
	for index := 2; index <= 4; index++ {
		response := value.admin(t, "POST", "upstreams", map[string]any{"name": fmt.Sprintf("controlled %d", index), "url": fmt.Sprintf("%s/candidate-%d/v1", value.upstream.URL, index), "token": "isolated-extra-token", "enabled": true})
		expect(t, response, 201)
		candidates = append(candidates, map[string]any{"upstream_id": index, "remote_model": "different-remote", "enabled": true, "priority": 0, "weight": 1})
	}
	expect(t, value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": 1, "enabled": true, "candidates": candidates}), 200)
	response := value.call(t, ordinary)
	id := response.Header.Get("X-Request-ID")
	expect(t, response, 502)
	var record Record
	decode(t, value.admin(t, "GET", "requests/"+id, nil), &record)
	if calls.Load() != 3 || len(record.Attempts) != 3 || record.Code != "upstream_attempts_exhausted" {
		t.Fatal("real HTTP attempt cap", calls.Load(), len(record.Attempts), record.Code)
	}
	for index, attempt := range record.Attempts {
		if attempt.UpstreamID != int64(index+1) {
			t.Fatal("duplicate or unexpected real dispatch")
		}
	}
}

func TestRoutingRealHTTPDisableBeforeFallback(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var secondary atomic.Int64
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/secondary/") {
			secondary.Add(1)
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprint(writer, `{"choices":[{"message":{"content":"controlled"}}]}`)
			return
		}
		close(entered)
		<-release
		writer.WriteHeader(503)
	})
	configureHTTPPair(t, value, 1)
	done := make(chan int, 1)
	go func() { response := value.call(t, ordinary); response.Body.Close(); done <- response.StatusCode }()
	<-entered
	model, _ := value.store.routeModel("alpha")
	expect(t, value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": false, "candidates": cleanTestCandidates(model.Candidates)}), 200)
	close(release)
	if status := <-done; status != 502 || secondary.Load() != 0 {
		t.Fatal("new fallback crossed confirmed disable", status, secondary.Load())
	}
}

func TestRoutingRealHTTPCommittedStreamNeverReplays(t *testing.T) {
	var secondary atomic.Int64
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/secondary/") {
			secondary.Add(1)
		}
		writer.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"actual first\"}}]}\n\n")
		writer.(http.Flusher).Flush()
	})
	configureHTTPPair(t, value, 1)
	response := value.call(t, streaming)
	data, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != 200 || !strings.Contains(string(data), "actual first") || strings.Contains(string(data), "[DONE]") || secondary.Load() != 0 {
		t.Fatal("committed replay")
	}
	var detail Record
	decode(t, value.admin(t, "GET", "requests/"+response.Header.Get("X-Request-ID"), nil), &detail)
	if len(detail.Attempts) != 1 || detail.State != "stream_interrupted" {
		t.Fatal("committed evidence")
	}
	encoded, _ := json.Marshal(detail)
	for _, secret := range []string{value.key, "upstream-secret-sentinel", "second-isolated-token", value.upstream.URL, "actual first"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatal("sensitive attempt evidence")
		}
	}
}
