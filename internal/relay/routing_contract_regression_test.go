package relay

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestCandidateIntegerAndReferenceBoundariesAreAtomic(testContext *testing.T) {
	value := newConsoleFixture(testContext)
	multiConfiguration(testContext, value, 2, nil)
	model, _ := value.server.Store.routeModel("alpha")
	for _, scenario := range []struct {
		name  string
		field string
		input any
	}{
		{"negative_priority", "priority", -1}, {"large_priority", "priority", 101}, {"fraction_priority", "priority", 0.5},
		{"zero_weight", "weight", 0}, {"large_weight", "weight", 101}, {"fraction_weight", "weight", 1.5},
		{"empty_remote", "remote_model", ""}, {"unknown_upstream", "upstream_id", 999}, {"missing_priority", "priority", nil},
	} {
		testContext.Run(scenario.name, func(testContext *testing.T) {
			candidates := cleanTestCandidates(model.Candidates)
			candidates[1][scenario.field] = scenario.input
			response := value.admin("PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": candidates})
			if response.Code != 400 {
				testContext.Fatal("invalid candidate accepted", response.Code)
			}
			after, _ := value.server.Store.routeModel("alpha")
			if !reflect.DeepEqual(after, model) {
				testContext.Fatal("rejected write partially changed model")
			}
		})
	}
	candidates := cleanTestCandidates(model.Candidates)
	candidates[0]["priority"], candidates[0]["weight"] = 0, 1
	candidates[1]["priority"], candidates[1]["weight"] = 100, 100
	if value.admin("PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": candidates}).Code != 200 {
		testContext.Fatal("integer endpoints rejected")
	}
}

func TestRoutingHealthEpochSequenceDisableAndRestart(testContext *testing.T) {
	value := newConsoleFixture(testContext)
	multiConfiguration(testContext, value, 2, nil)
	snapshot, err := value.server.snapshot("alpha")
	if err != nil {
		testContext.Fatal(err)
	}
	for _, candidate := range snapshot.Model.Candidates {
		upstream := snapshot.Upstreams[candidate.UpstreamID]
		attempt := Attempt{UpstreamExecution: upstream.ExecutionRevision, CandidateExecution: candidate.ExecutionRevision, Sequence: 2, State: "failed"}
		value.server.observe(snapshot, candidate, attempt, true)
		attempt.Sequence, attempt.State = 1, "success"
		value.server.observe(snapshot, candidate, attempt, false)
		if value.server.health[healthKey("alpha", candidate, upstream)].State != "failed" {
			testContext.Fatal("older late success erased cooling")
		}
	}
	var visits int
	value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) { visits++; return upstreamJSON(), nil })
	if response := value.call(value.key, ordinary); response.Code != 503 || visits != 0 {
		testContext.Fatal("all cooling visited upstream", response.Code, visits)
	}
	upstream := snapshot.Upstreams[1]
	for _, enabled := range []bool{false, true} {
		response := value.admin("PATCH", "upstreams/1", map[string]any{"expected_revision": upstream.Revision, "name": upstream.Name, "url": upstream.URL, "enabled": enabled})
		if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &upstream) != nil {
			testContext.Fatal("toggle")
		}
		if health := value.server.health[healthKey("alpha", snapshot.Model.Candidates[0], upstream)]; health == nil || !health.Cooldown.After(time.Now()) {
			testContext.Fatal("manual toggle cleared cooldown")
		}
	}
	response := value.admin("PATCH", "upstreams/1", map[string]any{"expected_revision": upstream.Revision, "name": upstream.Name, "url": upstream.URL, "enabled": true, "token": "changed-controlled-token"})
	if response.Code != 200 {
		testContext.Fatal("credential edit")
	}
	candidate := snapshot.Model.Candidates[0]
	value.server.observe(snapshot, candidate, Attempt{UpstreamExecution: snapshot.Upstreams[1].ExecutionRevision, Sequence: 99, State: "failed"}, true)
	latest, _ := value.server.Store.upstreamByID(1)
	if value.server.health[healthKey("alpha", candidate, latest)] != nil {
		testContext.Fatal("old execution poisoned new health")
	}
	restarted := New(value.server.Store, value.server.Config)
	if len(restarted.health) != 0 {
		testContext.Fatal("restart retained live health")
	}
	current, _ := value.server.Store.routeModel("alpha")
	candidates := cleanTestCandidates(current.Candidates)
	candidates[0]["remote_model"] = "new-remote"
	if value.admin("PATCH", "models/alpha", map[string]any{"expected_revision": current.Revision, "enabled": true, "candidates": candidates}).Code != 200 {
		testContext.Fatal("mapping edit")
	}
	value.server.observe(snapshot, candidate, Attempt{UpstreamExecution: latest.ExecutionRevision, Sequence: 100, State: "failed"}, true)
	current, _ = value.server.Store.routeModel("alpha")
	if value.server.health[healthKey("alpha", current.Candidates[0], latest)] != nil {
		testContext.Fatal("old candidate poisoned new mapping")
	}
}
