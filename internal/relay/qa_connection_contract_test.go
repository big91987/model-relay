package relay

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestQAConnectionCheckTenSecondBudget(t *testing.T) {
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 1, nil)
	upstream, err := value.server.Store.upstreamByID(1)
	if err != nil {
		t.Fatal(err)
	}
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		deadline, present := request.Context().Deadline()
		remaining := time.Until(deadline)
		if !present || remaining > 10*time.Second || remaining < 9*time.Second {
			t.Errorf("connection check budget = %s, want at most 10 seconds", remaining.Round(time.Millisecond))
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"object":"list","data":[]}`))}, nil
	})
	response := value.admin("POST", "upstreams/1/check", map[string]any{"expected_execution_revision": upstream.ExecutionRevision})
	if response.Code != 200 {
		t.Fatal("connection check endpoint did not complete")
	}
}

func TestQAConnectionCheckProtocolAndFailureStates(t *testing.T) {
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 1, nil)
	upstream, err := value.server.Store.upstreamByID(1)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name        string
		status      int
		contentType string
		body        string
		network     bool
		want        string
	}{
		{"valid_empty_list", 200, "application/json", `{"object":"list","data":[]}`, false, "success"},
		{"wrong_content_type", 200, "text/html", `{"object":"list","data":[]}`, false, "failed"},
		{"missing_list_object", 200, "application/json", `{"data":[]}`, false, "failed"},
		{"wrong_list_object", 200, "application/json", `{"object":"chat.completion","data":[]}`, false, "failed"},
		{"missing_model_id", 200, "application/json", `{"object":"list","data":[{}]}`, false, "failed"},
		{"empty_model_id", 200, "application/json", `{"object":"list","data":[{"id":""}]}`, false, "failed"},
		{"malformed_json", 200, "application/json", `{`, false, "failed"},
		{"authentication_failure", 401, "application/json", `{}`, false, "failed"},
		{"rate_limit_failure", 429, "application/json", `{}`, false, "failed"},
		{"server_failure", 503, "application/json", `{}`, false, "failed"},
		{"network_failure", 0, "", "", true, "failed"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
				if request.Method != http.MethodGet || request.Body != nil {
					t.Error("connection check must not generate")
				}
				if scenario.network {
					return nil, errors.New("controlled network failure")
				}
				return &http.Response{StatusCode: scenario.status, Header: http.Header{"Content-Type": {scenario.contentType}}, Body: io.NopCloser(strings.NewReader(scenario.body))}, nil
			})
			response := value.admin("POST", "upstreams/1/check", map[string]any{"expected_execution_revision": upstream.ExecutionRevision})
			var result CheckResult
			if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &result) != nil {
				t.Fatal("connection check result unavailable")
			}
			if result.State != scenario.want {
				t.Errorf("connection check state = %q, want %q", result.State, scenario.want)
			}
			stored, err := value.server.Store.upstreamByID(1)
			if err != nil || stored.Check == nil || stored.Check.State != scenario.want {
				t.Error("persisted connection check does not satisfy the contract")
			}
		})
	}
}
