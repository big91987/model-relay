package app

import (
	"context"

	"encoding/json"
	"model-relay/internal/records"

	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestCandidateIntegerAndReferenceBoundariesAreAtomic(testContext *testing.T) {
	value := newConsoleFixture(testContext)
	multiConfiguration(testContext, value, 2, nil)
	model, _ := catalogFor(value.server.Store).ModelFacts(context.Background(), "alpha")
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
			after, _ := catalogFor(value.server.Store).ModelFacts(context.Background(), "alpha")
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
	snapshot, err := value.server.Catalog.ModelSnapshot(context.Background(), "alpha")
	if err != nil {
		testContext.Fatal(err)
	}
	for _, candidate := range snapshot.Model.Candidates {
		upstream := snapshot.Upstreams[candidate.UpstreamID]
		attempt := records.Attempt{UpstreamExecution: upstream.ExecutionRevision, CandidateExecution: candidate.ExecutionRevision, Sequence: 2, State: "failed"}
		observeTestRouting(value.server, snapshot, candidate, attempt, true)
		attempt.Sequence, attempt.State = 1, "success"
		observeTestRouting(value.server, snapshot, candidate, attempt, false)
		if value.server.router.Health("alpha", testRoutingCandidate(candidate, upstream)).State != "failed" {
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
		if health := value.server.router.Health("alpha", testRoutingCandidate(snapshot.Model.Candidates[0], upstream)); health == nil || !health.Cooldown.After(time.Now()) {
			testContext.Fatal("manual toggle cleared cooldown")
		}
	}
	response := value.admin("PATCH", "upstreams/1", map[string]any{"expected_revision": upstream.Revision, "name": upstream.Name, "url": upstream.URL, "enabled": true, "token": "changed-controlled-token"})
	if response.Code != 200 {
		testContext.Fatal("credential edit")
	}
	candidate := snapshot.Model.Candidates[0]
	observeTestRouting(value.server, snapshot, candidate, records.Attempt{UpstreamExecution: snapshot.Upstreams[1].ExecutionRevision, Sequence: 99, State: "failed"}, true)
	latest, _ := catalogFor(value.server.Store).UpstreamFacts(context.Background(), 1)
	if value.server.router.Health("alpha", testRoutingCandidate(candidate, latest)) != nil {
		testContext.Fatal("old execution poisoned new health")
	}
	restarted := New(value.server.Store, value.server.Config)
	if restarted.router.Stats().HealthEntries != 0 {
		testContext.Fatal("restart retained live health")
	}
	current, _ := catalogFor(value.server.Store).ModelFacts(context.Background(), "alpha")
	candidates := cleanTestCandidates(current.Candidates)
	candidates[0]["remote_model"] = "new-remote"
	if value.admin("PATCH", "models/alpha", map[string]any{"expected_revision": current.Revision, "enabled": true, "candidates": candidates}).Code != 200 {
		testContext.Fatal("mapping edit")
	}
	observeTestRouting(value.server, snapshot, candidate, records.Attempt{UpstreamExecution: latest.ExecutionRevision, Sequence: 100, State: "failed"}, true)
	current, _ = catalogFor(value.server.Store).ModelFacts(context.Background(), "alpha")
	if value.server.router.Health("alpha", testRoutingCandidate(current.Candidates[0], latest)) != nil {
		testContext.Fatal("old candidate poisoned new mapping")
	}
}
