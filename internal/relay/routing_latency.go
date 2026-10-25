package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const latencyTTL = 60 * time.Second

type latencySample struct {
	Measured     time.Time
	Milliseconds float64
	Sequence     int64
}
type latencyPool struct {
	Samples  []latencySample
	Sequence int64
	Expired  bool
}
type explorationState struct {
	Counter uint64
	Last    map[int64]uint64
}
type SelectionObservation struct {
	Round       int      `json:"selection_round"`
	UpstreamID  int64    `json:"upstream_id"`
	Strategy    string   `json:"strategy"`
	Metric      string   `json:"metric_name"`
	Algorithm   string   `json:"algorithm"`
	Generation  uint64   `json:"generation"`
	Reason      string   `json:"selection_reason"`
	State       string   `json:"sample_state"`
	Count       int      `json:"sample_count"`
	EWMA        *float64 `json:"ewma_ms"`
	Measured    *string  `json:"measured_at"`
	Age         *float64 `json:"age_ms"`
	TTL         int      `json:"ttl_seconds"`
	Exploration uint64   `json:"exploration_sequence"`
	Bound       int      `json:"exploration_bound"`
}
type LatencyResult struct {
	Ordinal    int      `json:"ordinal"`
	UpstreamID int64    `json:"upstream_id"`
	Metric     string   `json:"metric_name"`
	Value      *float64 `json:"latency_ms"`
	Measured   *string  `json:"measured_at"`
	Finished   string   `json:"finished_at"`
	AcceptedAt string   `json:"evaluated_at"`
	Acceptance string   `json:"sample_acceptance"`
	Generation uint64   `json:"generation"`
}
type latencyMeasurement struct {
	Started  time.Time
	Measured time.Time
	Value    float64
}
type latencyReader struct {
	io.ReadCloser
	Measurement *latencyMeasurement
	Now         func() time.Time
}

func (reader *latencyReader) Read(data []byte) (int, error) {
	count, err := reader.ReadCloser.Read(data)
	if count > 0 && reader.Measurement.Measured.IsZero() {
		reader.Measurement.mark(reader.Now())
	}
	return count, err
}
func (measurement *latencyMeasurement) mark(now time.Time) {
	if measurement.Measured.IsZero() {
		measurement.Measured = now
		measurement.Value = float64(now.Sub(measurement.Started)) / float64(time.Millisecond)
	}
}
func latencyMetric(stream bool) string {
	if stream {
		return "first_output_delta_ms"
	}
	return "response_body_first_byte_ms"
}
func (server *Server) routingNow() time.Time {
	if server.latencyClock != nil {
		return server.latencyClock()
	}
	return time.Now()
}
func latencyKey(snapshot routeSnapshot, candidate Candidate, stream bool) string {
	return fmt.Sprintf("%s/%d/%d/%d/%t/%d", snapshot.Model.Alias, snapshot.Model.Revision, candidate.UpstreamID, snapshot.Upstreams[candidate.UpstreamID].ExecutionRevision, stream, candidate.ExecutionRevision)
}
func (pool *latencyPool) describe(now time.Time) (string, int, *float64, *string, *float64) {
	if pool == nil {
		return "unknown", 0, nil, nil, nil
	}
	if len(pool.Samples) == 0 {
		if pool.Expired {
			return "expired", 0, nil, nil, nil
		}
		return "unknown", 0, nil, nil, nil
	}
	recent := pool.Samples[:0]
	for _, sample := range pool.Samples {
		age := now.Sub(sample.Measured)
		if age >= 0 && age <= latencyTTL {
			recent = append(recent, sample)
		}
	}
	pool.Samples = recent
	pool.Expired = len(recent) == 0
	if pool.Expired {
		return "expired", 0, nil, nil, nil
	}
	average := recent[0].Milliseconds
	for _, sample := range recent[1:] {
		average = .25*sample.Milliseconds + .75*average
	}
	last := recent[len(recent)-1].Measured
	measured := last.UTC().Format(time.RFC3339Nano)
	age := float64(now.Sub(last)) / float64(time.Millisecond)
	state := "ready"
	if len(recent) < 3 {
		state = "insufficient"
		return state, len(recent), nil, &measured, &age
	}
	return state, len(recent), &average, &measured, &age
}
func (server *Server) chooseLatency(snapshot routeSnapshot, group []Candidate, stream bool, round int, weighted *Candidate, reasons []RouteDecision) (*Candidate, []SelectionObservation) {
	runtime := snapshot.Runtime
	if runtime.Exploration == nil {
		runtime.Exploration = map[bool]*explorationState{}
	}
	state := runtime.Exploration[stream]
	if state == nil {
		state = &explorationState{Last: map[int64]uint64{}}
		runtime.Exploration[stream] = state
	}
	observations := make([]SelectionObservation, 0, len(reasons))
	now := server.routingNow()
	for _, reason := range reasons {
		candidate := Candidate{UpstreamID: reason.UpstreamID}
		for _, value := range snapshot.Model.Candidates {
			if value.UpstreamID == reason.UpstreamID {
				candidate = value
				break
			}
		}
		status, count, average, measured, age := server.latency[latencyKey(snapshot, candidate, stream)].describe(now)
		observations = append(observations, SelectionObservation{Round: round, UpstreamID: candidate.UpstreamID, Strategy: snapshot.Model.Strategy, Metric: latencyMetric(stream), Algorithm: "ewma-explore-v2", Generation: snapshot.Epoch, Reason: reason.Reason, State: status, Count: count, EWMA: average, Measured: measured, Age: age, TTL: 60, Bound: 5 * len(group)})
	}
	selected := weighted
	selectionReason := "weighted"
	if len(group) > 0 {
		state.Counter++
		if snapshot.Model.Strategy == "low_latency" {
			if state.Counter%5 == 0 {
				selected = &group[0]
				for index := range group {
					candidate := &group[index]
					if state.Last[candidate.UpstreamID] < state.Last[selected.UpstreamID] {
						selected = candidate
					}
				}
				state.Last[selected.UpstreamID] = state.Counter
				selectionReason = "explore"
			} else {
				selectionReason = "weighted_fallback"
				var best *float64
				for index := range group {
					candidate := &group[index]
					_, _, score, _, _ := server.latency[latencyKey(snapshot, *candidate, stream)].describe(now)
					if score != nil && (best == nil || *score < *best) {
						selected = candidate
						best = score
						selectionReason = "lowest_latency"
					}
				}
			}
		}
	}
	for index := range observations {
		observation := &observations[index]
		observation.Exploration = state.Counter
		if selected != nil && observation.UpstreamID == selected.UpstreamID {
			observation.Reason = selectionReason
		} else if observation.Reason == "" {
			observation.Reason = "not_selected"
		}
	}
	return selected, observations
}
func outputDelta(value map[string]json.RawMessage) bool {
	choices, ok := array(value["choices"])
	if !ok {
		return false
	}
	for _, raw := range choices {
		choice, _ := object(raw)
		delta, _ := object(choice["delta"])
		for _, field := range []string{"content", "refusal"} {
			if text, ok := stringValue(delta[field]); ok && text != "" {
				return true
			}
		}
		calls, _ := array(delta["tool_calls"])
		for _, callRaw := range calls {
			call, _ := object(callRaw)
			function, _ := object(call["function"])
			for _, field := range []string{"name", "arguments"} {
				if text, ok := stringValue(function[field]); ok && text != "" {
					return true
				}
			}
		}
	}
	return false
}
func (server *Server) acceptLatency(ctx context.Context, id string, snapshot routeSnapshot, candidate Candidate, attempt Attempt, stream bool, measurement latencyMeasurement) {
	server.mu.Lock()
	defer server.mu.Unlock()
	now := server.routingNow()
	result := LatencyResult{Ordinal: attempt.Ordinal, UpstreamID: candidate.UpstreamID, Metric: latencyMetric(stream), Acceptance: "not_success", Generation: snapshot.Epoch, AcceptedAt: now.UTC().Format(time.RFC3339Nano)}
	if attempt.Finished != nil {
		result.Finished = *attempt.Finished
	}
	if !measurement.Measured.IsZero() {
		measured := measurement.Measured.UTC().Format(time.RFC3339Nano)
		result.Measured = &measured
		result.Value = &measurement.Value
	}
	key := latencyKey(snapshot, candidate, stream)
	pool := server.latency[key]
	switch {
	case attempt.State != "success" || ctx.Err() != nil:
	case server.degraded.Load():
		result.Acceptance = "storage_unavailable"
	case measurement.Measured.IsZero():
		result.Acceptance = "no_output_delta"
	case now.Before(measurement.Measured) || measurement.Value < 0:
		result.Acceptance = "clock_invalid"
	case now.Sub(measurement.Measured) > latencyTTL:
		result.Acceptance = "expired"
	case snapshot.Epoch != server.configEpoch:
		result.Acceptance = "configuration_changed"
	default:
		current, err := server.Store.routeModel(snapshot.Model.Alias)
		upstream, upstreamErr := server.Store.upstreamByID(candidate.UpstreamID)
		health := server.health[healthKey(snapshot.Model.Alias, candidate, upstream)]
		valid := err == nil && upstreamErr == nil && current.Enabled && upstream.Enabled && current.Revision == snapshot.Model.Revision && upstream.ExecutionRevision == attempt.UpstreamExecution
		if !valid {
			result.Acceptance = "configuration_changed"
		} else if health != nil && attempt.Sequence <= health.InvalidatedThrough {
			result.Acceptance = "invalidated_by_failure"
		} else if pool != nil && (attempt.Sequence <= pool.Sequence || len(pool.Samples) > 0 && measurement.Measured.Before(pool.Samples[len(pool.Samples)-1].Measured)) {
			result.Acceptance = "out_of_order"
		} else {
			result.Acceptance = "accepted"
		}
	}
	raw, err := json.Marshal(result)
	if err == nil {
		_, err = server.Store.DB.Exec("INSERT INTO route_latency_evidence VALUES(?,?,?) ON CONFLICT(request_id,ordinal) DO UPDATE SET evidence=excluded.evidence", id, attempt.Ordinal, string(raw))
	}
	if err != nil {
		server.storageFailure()
		return
	}
	if result.Acceptance != "accepted" {
		return
	}
	if pool == nil {
		pool = &latencyPool{}
		server.latency[key] = pool
	}
	pool.describe(now)
	pool.Samples = append(pool.Samples, latencySample{Measured: measurement.Measured, Milliseconds: measurement.Value, Sequence: attempt.Sequence})
	if len(pool.Samples) > 32 {
		pool.Samples = pool.Samples[len(pool.Samples)-32:]
	}
	pool.Sequence = attempt.Sequence
}
