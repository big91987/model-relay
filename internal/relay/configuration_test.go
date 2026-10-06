package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestConnectionCheckDoesNotGenerateAndScopesLateResults(t *testing.T) {
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 2, nil)
	upstream, err := value.server.Store.upstreamByID(1)
	if err != nil {
		t.Fatal(err)
	}
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		if request.Method != "GET" || request.URL.Path != "/a/v1/models" || request.Body != nil {
			t.Error("connection check generated or used another target")
		}
		response := value.admin("PATCH", "upstreams/1", map[string]any{"expected_revision": upstream.Revision, "name": upstream.Name, "url": upstream.URL, "enabled": true, "token": "replacement-controlled-token"})
		if response.Code != 200 {
			t.Error("configuration edit during check", response.Code)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"object":"list","data":[]}`))}, nil
	})
	response := value.admin("POST", "upstreams/1/check", map[string]any{"expected_execution_revision": upstream.ExecutionRevision})
	var result CheckResult
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &result) != nil || result.State != "success" || result.Applicable {
		t.Fatal("late check incorrectly applicable", response.Code, response.Body.String())
	}
	latest, err := value.server.Store.upstreamByID(1)
	if err != nil || latest.Check == nil || latest.Check.Applicable || latest.ExecutionRevision == upstream.ExecutionRevision {
		t.Fatal("late check did not retain its execution scope")
	}
	var attempts int
	if value.server.Store.DB.QueryRow("SELECT count(*) FROM request_attempts").Scan(&attempts) != nil || attempts != 0 {
		t.Fatal("connection check recorded generation")
	}
}

func TestCandidateWriteRejectsDuplicatesAtomically(t *testing.T) {
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 2, nil)
	model, err := value.server.Store.routeModel("alpha")
	if err != nil {
		t.Fatal(err)
	}
	candidates := cleanTestCandidates(model.Candidates)
	candidates[1]["upstream_id"] = 1
	response := value.admin("PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": candidates})
	if response.Code != 400 {
		t.Fatal("duplicate accepted", response.Code)
	}
	after, err := value.server.Store.routeModel("alpha")
	if err != nil || after.Revision != model.Revision || len(after.Candidates) != 2 || after.Candidates[1].UpstreamID != 2 {
		t.Fatal("rejected candidate write partially committed")
	}
}
