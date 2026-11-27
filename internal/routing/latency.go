package routing

import (
	"context"
	"fmt"
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

func latencyMetric(stream bool) string {
	if stream {
		return "first_output_delta_ms"
	}
	return "response_body_first_byte_ms"
}

func latencyKey(snapshot Snapshot, candidate Candidate, stream bool) string {
	return fmt.Sprintf("%s/%d/%d/%d/%t/%d", snapshot.Alias, snapshot.Revision, candidate.UpstreamID, candidate.UpstreamExecution, stream, candidate.ExecutionRevision)
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

func (engine *Engine) chooseLatency(snapshot Snapshot, runtime *routeRuntime, group []Candidate, stream bool, round int, weighted *Candidate, reasons []RouteDecision, now time.Time) (*Candidate, []SelectionObservation) {
	if runtime.Exploration == nil {
		runtime.Exploration = map[bool]*explorationState{}
	}
	state := runtime.Exploration[stream]
	if state == nil {
		state = &explorationState{Last: map[int64]uint64{}}
		runtime.Exploration[stream] = state
	}
	observations := make([]SelectionObservation, 0, len(reasons))
	for _, reason := range reasons {
		candidate := Candidate{UpstreamID: reason.UpstreamID}
		for _, value := range snapshot.Candidates {
			if value.UpstreamID == reason.UpstreamID {
				candidate = value
				break
			}
		}
		status, count, average, measured, age := engine.latency[latencyKey(snapshot, candidate, stream)].describe(now)
		observations = append(observations, SelectionObservation{Round: round, UpstreamID: candidate.UpstreamID, Strategy: snapshot.Strategy, Metric: latencyMetric(stream), Algorithm: "ewma-explore-v2", Generation: snapshot.Epoch, Reason: reason.Reason, State: status, Count: count, EWMA: average, Measured: measured, Age: age, TTL: 60, Bound: 5 * len(group)})
	}
	selected := weighted
	selectionReason := "weighted"
	if len(group) > 0 {
		state.Counter++
		if snapshot.Strategy == "low_latency" {
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
					_, _, score, _, _ := engine.latency[latencyKey(snapshot, *candidate, stream)].describe(now)
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

// RecordLatency saves the decision before making a sample visible to selection.
// Configuration validation and durable evidence are adapter operations; neither
// HTTP requests nor SQL handles cross this interface.
func (engine *Engine) RecordLatency(ctx context.Context, sample Sample, isCurrent func() bool, save func(LatencyResult) error) error {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	now := sample.EvaluatedAt
	snapshot, candidate, measurement := sample.Snapshot, sample.Candidate, sample.Measurement
	result := LatencyResult{Ordinal: sample.Ordinal, UpstreamID: candidate.UpstreamID, Metric: latencyMetric(sample.Stream), Acceptance: "not_success", Generation: snapshot.Epoch, AcceptedAt: now.UTC().Format(time.RFC3339Nano)}
	if sample.Finished != nil {
		result.Finished = *sample.Finished
	}
	if !measurement.Measured.IsZero() {
		measured := measurement.Measured.UTC().Format(time.RFC3339Nano)
		result.Measured = &measured
		result.Value = &measurement.Value
	}
	key := latencyKey(snapshot, candidate, sample.Stream)
	pool := engine.latency[key]
	switch {
	case sample.State != "success" || ctx.Err() != nil:
	case sample.Degraded:
		result.Acceptance = "storage_unavailable"
	case measurement.Measured.IsZero():
		result.Acceptance = "no_output_delta"
	case now.Before(measurement.Measured) || measurement.Value < 0:
		result.Acceptance = "clock_invalid"
	case now.Sub(measurement.Measured) > latencyTTL:
		result.Acceptance = "expired"
	case snapshot.Epoch != sample.CurrentEpoch:
		result.Acceptance = "configuration_changed"
	default:
		health := engine.health[healthKey(snapshot.Alias, candidate)]
		if !isCurrent() {
			result.Acceptance = "configuration_changed"
		} else if health != nil && sample.Sequence <= health.InvalidatedThrough {
			result.Acceptance = "invalidated_by_failure"
		} else if pool != nil && (sample.Sequence <= pool.Sequence || len(pool.Samples) > 0 && measurement.Measured.Before(pool.Samples[len(pool.Samples)-1].Measured)) {
			result.Acceptance = "out_of_order"
		} else {
			result.Acceptance = "accepted"
		}
	}
	if err := save(result); err != nil {
		return err
	}
	if result.Acceptance != "accepted" {
		return nil
	}
	if pool == nil {
		pool = &latencyPool{}
		engine.latency[key] = pool
	}
	pool.describe(now)
	pool.Samples = append(pool.Samples, latencySample{Measured: measurement.Measured, Milliseconds: measurement.Value, Sequence: sample.Sequence})
	if len(pool.Samples) > 32 {
		pool.Samples = pool.Samples[len(pool.Samples)-32:]
	}
	pool.Sequence = sample.Sequence
	return nil
}
