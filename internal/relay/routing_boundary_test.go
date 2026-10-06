package relay

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type unlockedAdminWriter struct {
	*httptest.ResponseRecorder
	server  *Server
	testing *testing.T
}

func (writer unlockedAdminWriter) WriteHeader(status int) {
	if !writer.server.mu.TryLock() {
		writer.testing.Error("management network write holds configuration lock")
	} else {
		writer.server.mu.Unlock()
	}
	writer.ResponseRecorder.WriteHeader(status)
}

func TestManagementResponseReleasesConfigurationLock(t *testing.T) {
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 2, nil)
	for _, path := range []string{"upstreams", "models", "overview", "keys"} {
		request := httptest.NewRequest("GET", "http://console.test/admin/api/"+path, nil)
		request.AddCookie(value.cookie)
		response := httptest.NewRecorder()
		value.server.ServeHTTP(unlockedAdminWriter{response, value.server, t}, request)
		if response.Code != 200 {
			t.Fatal("management read", path, response.Code)
		}
	}
}

func TestRoutingDeadlineAndCancellationNeverRestart(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		value := newConsoleFixture(t)
		multiConfiguration(t, value, 2, nil)
		value.server.Config.TotalTimeout = 20 * time.Millisecond
		value.server.Config.AttemptTimeout = time.Second
		ctx, cancel := context.WithCancel(context.Background())
		var calls atomic.Int64
		value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
			calls.Add(1)
			if canceled {
				cancel()
			}
			<-request.Context().Done()
			return nil, request.Context().Err()
		})
		request := httptest.NewRequest("POST", "http://console.test/v1/chat/completions", strings.NewReader(ordinary)).WithContext(ctx)
		request.Header.Set("Authorization", "Bearer "+value.key)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		value.server.ServeHTTP(response, request)
		cancel()
		if calls.Load() != 1 {
			t.Fatal("deadline/cancel restarted generation", calls.Load())
		}
		if !canceled && response.Code != 504 {
			t.Fatal("overall deadline result", response.Code)
		}
	}
}

func TestRoutingPrecommitTimeoutCanSwitchOnce(t *testing.T) {
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 2, nil)
	value.server.Config.AttemptTimeout = 20 * time.Millisecond
	value.server.Config.TotalTimeout = time.Second
	var calls atomic.Int64
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		calls.Add(1)
		if strings.Contains(request.URL.Path, "/a/") {
			<-request.Context().Done()
			return nil, request.Context().Err()
		}
		return upstreamJSON(), nil
	})
	response := value.call(value.key, ordinary)
	if response.Code != 200 || calls.Load() != 2 {
		t.Fatal("precommit timeout did not switch", response.Code, calls.Load())
	}
}

func TestAttemptOutputStopsTimerOnFirstCommit(t *testing.T) {
	response := httptest.NewRecorder()
	output := newAttemptOutput(response, context.Background(), 20*time.Millisecond)
	defer output.stop()
	output.Header().Set("Content-Type", "text/event-stream")
	if _, err := io.WriteString(output, "data: first\n\n"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)
	if output.timeout() || output.child.Err() != nil || !output.isCommitted() {
		t.Fatal("attempt timer canceled an already committed stream")
	}
}
