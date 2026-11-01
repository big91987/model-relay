package relay

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestExternalHTTPRequestAdvancesEffectiveClockOnce(t *testing.T) {
	fixture := singleResourceInstallation(t, false)
	var observations atomic.Int64
	fixture.store.clock = func() time.Time {
		observations.Add(1)
		return time.Unix(0, fixture.clock.Load())
	}

	tests := []struct {
		name string
		call func() int
	}{
		{
			name: "model list",
			call: func() int {
				request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:5544/v1/models", nil)
				request.Header.Set("Authorization", "Bearer "+fixture.key)
				writer := httptest.NewRecorder()
				fixture.server.ServeHTTP(writer, request)
				return writer.Code
			},
		},
		{
			name: "identity tenant list",
			call: func() int {
				status, _ := fixture.admin.call(http.MethodGet, "tenants", nil, "")
				return status
			},
		},
		{
			name: "console overview",
			call: func() int {
				status, _ := fixture.admin.call(http.MethodGet, "workspace/overview?tenant="+fixture.tenant, nil, "")
				return status
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observations.Store(0)
			if status := test.call(); status != http.StatusOK {
				t.Fatalf("request status = %d", status)
			}
			if calls := observations.Load(); calls != 1 {
				t.Fatalf("wall clock observations = %d, want exactly one", calls)
			}
		})
	}
}

func TestSingleRequestClockAdvanceRetainsHighWaterOnRollback(t *testing.T) {
	fixture := singleResourceInstallation(t, false)
	high := time.Date(2026, 10, 1, 0, 1, 0, 0, time.UTC)
	fixture.clock.Store(high.UnixNano())
	if status, _ := fixture.admin.call(http.MethodGet, "tenants", nil, ""); status != http.StatusOK {
		t.Fatal("establish high water", status)
	}

	var observations atomic.Int64
	fixture.store.clock = func() time.Time {
		observations.Add(1)
		return high.Add(-time.Hour)
	}
	if status, _ := fixture.admin.call(http.MethodGet, "workspace/overview?tenant="+fixture.tenant, nil, ""); status != http.StatusOK {
		t.Fatal("rollback request", status)
	}
	if calls := observations.Load(); calls != 1 {
		t.Fatalf("rollback wall clock observations = %d, want exactly one", calls)
	}
	if got := fixture.store.now(); !got.Equal(high) {
		t.Fatalf("effective clock regressed to %s, want %s", got, high)
	}
	var persisted string
	if err := fixture.store.DB.QueryRow("SELECT effective FROM console_clock WHERE id=1").Scan(&persisted); err != nil {
		t.Fatal(err)
	}
	if got, err := time.Parse(time.RFC3339Nano, persisted); err != nil || !got.Equal(high) {
		t.Fatalf("persisted high water = %q (%v), want %s", persisted, err, high)
	}
}

func TestSingleRequestClockAdvanceFailsClosedOnStorageError(t *testing.T) {
	fixture := singleResourceInstallation(t, false)
	var observations atomic.Int64
	fixture.store.clock = func() time.Time {
		observations.Add(1)
		return time.Unix(0, fixture.clock.Load())
	}
	if err := fixture.store.DB.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:5544/v1/models", nil)
	request.Header.Set("Authorization", "Bearer "+fixture.key)
	writer := httptest.NewRecorder()
	fixture.server.ServeHTTP(writer, request)
	if writer.Code != http.StatusServiceUnavailable {
		t.Fatalf("request status = %d, want 503", writer.Code)
	}
	if calls := observations.Load(); calls != 0 {
		t.Fatalf("wall clock observations after begin failure = %d, want zero", calls)
	}
	if !fixture.server.degraded.Load() {
		t.Fatal("storage failure did not degrade server")
	}
}
