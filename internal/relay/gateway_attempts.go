package relay

import (
	"context"
	"model-relay/internal/catalog"
	"model-relay/internal/records"
	"model-relay/internal/routing"
	"net/http"
	"time"
)

func (server *Gateway) beginAttempt(ctx context.Context, request *http.Request, record *records.Record, snapshot catalog.Snapshot, tried map[int64]bool, round int) (*catalog.Candidate, *records.Attempt, string) {
	server.State.Lock()
	defer server.State.Unlock()
	if server.Store.AdvanceClock() != nil {
		return nil, nil, "storage_unavailable"
	}
	if ctx.Err() != nil {
		return nil, nil, "request_timeout"
	}
	if server.Degraded.Load() {
		return nil, nil, "storage_unavailable"
	}
	if _, ok := server.appID(request); !ok {
		return nil, nil, "unauthorized"
	}
	if server.Store.Schema == 3 {
		if allowed, code := server.Keys.AuthorizeModel(record.KeyID, record.Alias); !allowed {
			return nil, nil, code
		}
	}
	current, err := server.Catalog.ModelFacts(context.Background(), snapshot.Model.Alias)
	if err != nil {
		return nil, nil, "storage_unavailable"
	}
	now := server.routingNow()
	policy := routingSnapshot(snapshot, now)
	for index, candidate := range snapshot.Model.Candidates {
		upstream := snapshot.Upstreams[candidate.UpstreamID]
		latest, err := server.Catalog.UpstreamFacts(context.Background(), candidate.UpstreamID)
		if err != nil {
			return nil, nil, "storage_unavailable"
		}
		present := false
		for _, value := range current.Candidates {
			if value.UpstreamID == candidate.UpstreamID && value.Enabled {
				present = true
			}
		}
		reason := ""
		switch {
		case !current.Enabled || !snapshot.Model.Enabled:
			reason = "disabled_model"
		case snapshot.Epoch != server.State.Epoch() || current.Revision != snapshot.Model.Revision || upstream.ExecutionRevision != latest.ExecutionRevision:
			reason = "configuration_changed"
		case !upstream.Enabled || !latest.Enabled:
			reason = "disabled_upstream"
		case !candidate.Enabled || !present:
			reason = "disabled_candidate"
		case !upstream.CredentialConfigured:
			reason = "credential_missing"
		}
		policy.Candidates[index].Unavailable = reason
	}
	selection := server.Router.Select(policy, record.Stream, round, tried)
	decisions, observations := selection.Decisions, selection.Observations
	var selected *catalog.Candidate
	if selection.Candidate != nil {
		for _, candidate := range snapshot.Model.Candidates {
			if candidate.UpstreamID == selection.Candidate.UpstreamID {
				copy := candidate
				selected = &copy
				break
			}
		}
	}
	var attempt *records.Attempt
	if selected != nil {
		upstream := snapshot.Upstreams[selected.UpstreamID]
		attempt = &records.Attempt{Ordinal: len(tried) + 1, UpstreamID: upstream.ID, UpstreamName: upstream.Name, RemoteModel: selected.RemoteModel, UpstreamRevision: upstream.Revision, UpstreamExecution: upstream.ExecutionRevision, CandidateExecution: selected.ExecutionRevision, Started: now.UTC().Format(time.RFC3339Nano), State: "dispatching"}
	}
	if err := server.Records.BeginAttempt(record.ID, round, decisions, observations, attempt, server.Router.NextSequence); err != nil {
		return nil, nil, "storage_unavailable"
	}
	return selected, attempt, ""
}

func (server *Gateway) observe(snapshot catalog.Snapshot, candidate catalog.Candidate, attempt records.Attempt, temporary bool) {
	server.State.Lock()
	defer server.State.Unlock()
	current, err := server.Catalog.ModelFacts(context.Background(), snapshot.Model.Alias)
	if err != nil {
		return
	}
	upstream, err := server.Catalog.UpstreamFacts(context.Background(), candidate.UpstreamID)
	if err != nil || upstream.ExecutionRevision != attempt.UpstreamExecution {
		return
	}
	valid := false
	for _, value := range current.Candidates {
		if value.UpstreamID == candidate.UpstreamID && value.RemoteModel == candidate.RemoteModel && value.ExecutionRevision == candidate.ExecutionRevision {
			valid = true
		}
	}
	if !valid {
		return
	}
	server.Router.Observe(routingSnapshot(snapshot, time.Time{}), routingCandidate(candidate, upstream), routing.Observation{Cooldown: server.Config.Cooldown, Sequence: attempt.Sequence, State: attempt.State, Temporary: temporary, ObservedAt: server.routingNow()})
}

func temporaryStatus(status int) bool {
	return status == 408 || status == 429 || status == 500 || status == 502 || status == 503 || status == 504
}

func routingCandidate(candidate catalog.Candidate, upstream catalog.Upstream) routing.Candidate {
	return routing.Candidate{UpstreamID: candidate.UpstreamID, RemoteModel: candidate.RemoteModel, Enabled: candidate.Enabled, Priority: candidate.Priority, Weight: candidate.Weight, ExecutionRevision: candidate.ExecutionRevision, UpstreamExecution: upstream.ExecutionRevision}
}

func routingSnapshot(snapshot catalog.Snapshot, observed time.Time) routing.Snapshot {
	policy := routing.Snapshot{Alias: snapshot.Model.Alias, Strategy: snapshot.Model.Strategy, Revision: snapshot.Model.Revision, ConfigRevision: snapshot.ConfigRevision, Epoch: snapshot.Epoch, ObservedAt: observed}
	for _, candidate := range snapshot.Model.Candidates {
		policy.Candidates = append(policy.Candidates, routingCandidate(candidate, snapshot.Upstreams[candidate.UpstreamID]))
	}
	return policy
}
