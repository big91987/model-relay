package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestConnectionCancelDeduplicatesAndPreservesRouting(testContext *testing.T) {
	value := newConsoleFixture(testContext)
	multiConfiguration(testContext, value, 2, nil)
	upstream, _ := value.server.Store.upstreamByID(1)
	started := make(chan struct{})
	finished := make(chan struct{})
	var visits atomic.Int64
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		visits.Add(1)
		close(started)
		<-request.Context().Done()
		return nil, request.Context().Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	data, _ := json.Marshal(map[string]any{"expected_execution_revision": upstream.ExecutionRevision})
	request := httptest.NewRequest("POST", "http://console.test/admin/api/upstreams/1/check", bytes.NewReader(data)).WithContext(ctx)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://console.test")
	request.Header.Set("X-CSRF-Token", value.csrf)
	request.AddCookie(value.cookie)
	response := httptest.NewRecorder()
	go func() { defer close(finished); value.server.ServeHTTP(response, request) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		testContext.Fatal("check did not dispatch")
	}
	duplicate := value.admin("POST", "upstreams/1/check", map[string]any{"expected_execution_revision": upstream.ExecutionRevision})
	if duplicate.Code != 409 || !strings.Contains(duplicate.Body.String(), "check_in_progress") || visits.Load() != 1 {
		testContext.Fatal("duplicate check dispatched")
	}
	cancel()
	select {
	case <-finished:
	case <-time.After(time.Second):
		testContext.Fatal("outbound cancellation did not finish")
	}
	var result CheckResult
	if json.Unmarshal(response.Body.Bytes(), &result) != nil || result.State != "canceled" || result.Code != "client_canceled" {
		testContext.Fatal("cancellation result")
	}
	stored, _ := value.server.Store.upstreamByID(1)
	if !stored.Enabled || stored.ExecutionRevision != upstream.ExecutionRevision || stored.Check.State != "canceled" || len(value.server.health) != 0 {
		testContext.Fatal("check cancellation changed execution or health")
	}
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: io.NopCloser(strings.NewReader(`{"object":"list","data":[{"id":"valid"}]}`))}, nil
	})
	retry := value.admin("POST", "upstreams/1/check", map[string]any{"expected_execution_revision": upstream.ExecutionRevision})
	if retry.Code != 200 || !strings.Contains(retry.Body.String(), `"state":"success"`) {
		testContext.Fatal("check gate did not release")
	}
}

func TestConnectionRealHTTPHeaderAndBodyBudget(testContext *testing.T) {
	for _, phase := range []string{"headers", "body"} {
		testContext.Run(phase, func(testContext *testing.T) {
			testContext.Parallel()
			canceled := make(chan struct{})
			value := newFixture(testContext, func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != "GET" || !strings.HasSuffix(request.URL.Path, "/models") {
					testContext.Error("check generated")
				}
				if phase == "body" {
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(200)
					writer.(http.Flusher).Flush()
				}
				<-request.Context().Done()
				close(canceled)
			})
			transport := value.server.client.Transport.(*http.Transport).Clone()
			transport.ResponseHeaderTimeout = 30 * time.Second
			value.server.client.Transport = transport
			upstream, _ := value.store.upstreamByID(1)
			started := time.Now()
			var result CheckResult
			decode(testContext, value.admin(testContext, "POST", "upstreams/1/check", map[string]any{"expected_execution_revision": upstream.ExecutionRevision}), &result)
			elapsed := time.Since(started)
			if result.State != "failed" || result.Code != "upstream_timeout" || elapsed < 9*time.Second || elapsed > 12*time.Second {
				testContext.Fatal("check total budget or failure classification", phase, elapsed, result.State, result.Code)
			}
			select {
			case <-canceled:
			case <-time.After(time.Second):
				testContext.Fatal("upstream still running after deadline")
			}
			stored, _ := value.store.upstreamByID(1)
			if stored.Check.State != "failed" || stored.Check.Code != "upstream_timeout" {
				testContext.Fatal("persisted timeout differs")
			}
		})
	}
}
