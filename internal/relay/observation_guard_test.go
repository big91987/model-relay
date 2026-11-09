package relay

import (
	"context"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/routing"
	"model-relay/internal/upstream"
)

func observationGuardFixture(t *testing.T) *Gateway {
	t.Helper()
	root := t.TempDir()
	directory, keyFile := filepath.Join(root, "data"), filepath.Join(root, "material")
	if _, err := installation.Init(directory, keyFile); err != nil {
		t.Fatal(err)
	}
	store, err := installation.Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	state, router := configstate.New(), routing.New()
	client := upstream.NewClient(&http.Client{})
	models := catalog.New(store.DB, store.Key, store.Schema, state, router, client, upstream.Policy{})
	now := time.Date(2026, 11, 9, 1, 0, 0, 0, time.UTC)
	configuration := &Config{Cooldown: time.Minute, LatencyClock: func() time.Time { return now }}
	gateway := &Gateway{Store: store, Config: configuration, State: state, Router: router, Catalog: models, Degraded: &atomic.Bool{}}
	enabled, priority, weight := true, 0, 1
	source, err := models.CreateUpstream(context.Background(), "", catalog.UpstreamInput{Name: "Observation source", URL: "https://8.8.8.8/v1", Token: "controlled-token", Enabled: &enabled})
	if err != nil {
		t.Fatal(err)
	}
	candidates := []catalog.CandidateInput{{UpstreamID: source.ID, RemoteModel: "original", Enabled: &enabled, Priority: &priority, Weight: &weight}}
	if _, err = models.CreateModel(context.Background(), "", catalog.ModelInput{Alias: "guarded", Enabled: &enabled, Candidates: &candidates}); err != nil {
		t.Fatal(err)
	}
	return gateway
}

func observationSnapshot(t *testing.T, gateway *Gateway) catalog.Snapshot {
	t.Helper()
	gateway.State.RLock()
	snapshot, err := gateway.snapshot("guarded")
	gateway.State.RUnlock()
	if err != nil || len(snapshot.Model.Candidates) != 1 {
		t.Fatal("observation snapshot", snapshot, err)
	}
	return snapshot
}

func currentObservationControl(t *testing.T, gateway *Gateway, snapshot catalog.Snapshot) routing.Health {
	t.Helper()
	candidate := snapshot.Model.Candidates[0]
	source := snapshot.Upstreams[candidate.UpstreamID]
	gateway.observe(snapshot, candidate, records.Attempt{UpstreamExecution: source.ExecutionRevision, CandidateExecution: candidate.ExecutionRevision, Sequence: 10, State: "success"}, false)
	health := gateway.Router.Health(snapshot.Model.Alias, routingCandidate(candidate, source))
	if health == nil || health.Sequence != 10 || health.State != "success" || !health.Cooldown.IsZero() {
		t.Fatal("current observation control did not adopt success", health)
	}
	return *health
}

func TestGatewayObserveRejectsStaleUpstreamExecution(t *testing.T) {
	gateway := observationGuardFixture(t)
	stale := observationSnapshot(t, gateway)
	candidate := stale.Model.Candidates[0]
	source := stale.Upstreams[candidate.UpstreamID]
	enabled := true
	if _, err := gateway.Catalog.UpdateUpstream(context.Background(), "", source.ID, catalog.UpstreamInput{Name: source.Name, URL: source.URL, Token: "rotated-controlled-token", Enabled: &enabled, Expected: &source.Revision}); err != nil {
		t.Fatal(err)
	}
	current := observationSnapshot(t, gateway)
	latest := current.Upstreams[candidate.UpstreamID]
	if latest.ExecutionRevision == source.ExecutionRevision {
		t.Fatal("credential update did not advance upstream execution")
	}
	control := currentObservationControl(t, gateway, current)
	// Without the upstream execution guard, observe combines this old candidate
	// with latest upstream facts and overwrites the current execution's same key.
	gateway.observe(stale, candidate, records.Attempt{UpstreamExecution: source.ExecutionRevision, CandidateExecution: candidate.ExecutionRevision, Sequence: 100, State: "failed"}, true)
	health := gateway.Router.Health(current.Model.Alias, routingCandidate(current.Model.Candidates[0], latest))
	if health == nil || *health != control {
		t.Fatal("stale upstream observation poisoned current health", health)
	}
}

func TestGatewayObserveRejectsChangedCandidateMapping(t *testing.T) {
	for _, restored := range []bool{false, true} {
		t.Run(map[bool]string{false: "new_remote", true: "restored_remote_new_execution"}[restored], func(t *testing.T) {
			gateway := observationGuardFixture(t)
			stale := observationSnapshot(t, gateway)
			oldCandidate := stale.Model.Candidates[0]
			update := func(remote string) {
				t.Helper()
				before := observationSnapshot(t, gateway)
				enabled, priority, weight := true, 0, 1
				candidates := []catalog.CandidateInput{{UpstreamID: oldCandidate.UpstreamID, RemoteModel: remote, Enabled: &enabled, Priority: &priority, Weight: &weight}}
				if _, err := gateway.Catalog.UpdateModel(context.Background(), "", "guarded", catalog.ModelInput{Enabled: &enabled, Expected: &before.Model.Revision, Candidates: &candidates}); err != nil {
					t.Fatal(err)
				}
			}
			update("replacement")
			if restored {
				update(oldCandidate.RemoteModel)
			}
			current := observationSnapshot(t, gateway)
			candidate := current.Model.Candidates[0]
			source := current.Upstreams[candidate.UpstreamID]
			if candidate.ExecutionRevision <= oldCandidate.ExecutionRevision {
				t.Fatal("mapping update did not advance candidate execution")
			}
			control := currentObservationControl(t, gateway, current)
			// Restoring the remote model intentionally reuses the engine health key;
			// candidate execution, checked by Gateway.observe, still distinguishes it.
			gateway.observe(stale, oldCandidate, records.Attempt{UpstreamExecution: source.ExecutionRevision, CandidateExecution: oldCandidate.ExecutionRevision, Sequence: 100, State: "failed"}, true)
			health := gateway.Router.Health(current.Model.Alias, routingCandidate(candidate, source))
			if health == nil || *health != control {
				t.Fatal("stale candidate observation poisoned current health", health)
			}
			if !restored && gateway.Router.Health(stale.Model.Alias, routingCandidate(oldCandidate, source)) != nil {
				t.Fatal("stale candidate observation recreated obsolete mapping health")
			}
			if gateway.Router.Stats().HealthEntries != 1 {
				t.Fatal("stale candidate observation entered routing state")
			}
		})
	}
}
