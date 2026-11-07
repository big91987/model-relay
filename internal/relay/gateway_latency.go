package relay

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"model-relay/internal/routing"
)

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

func (server *Server) routingNow() time.Time {
	if server.latencyClock != nil {
		return server.latencyClock()
	}
	return time.Now()
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
	sample := routing.Sample{Snapshot: routingSnapshot(snapshot, time.Time{}), Candidate: routingCandidate(candidate, snapshot.Upstreams[candidate.UpstreamID]), Ordinal: attempt.Ordinal, Sequence: attempt.Sequence, State: attempt.State, Finished: attempt.Finished, CurrentEpoch: server.configEpoch, Degraded: server.degraded.Load(), Stream: stream, EvaluatedAt: server.routingNow(), Measurement: routing.Measurement{Measured: measurement.Measured, Value: measurement.Value}}
	err := server.router.RecordLatency(ctx, sample, func() bool {
		current, err := server.Store.routeModel(snapshot.Model.Alias)
		upstream, upstreamErr := server.Store.upstreamByID(candidate.UpstreamID)
		return err == nil && upstreamErr == nil && current.Enabled && upstream.Enabled && current.Revision == snapshot.Model.Revision && upstream.ExecutionRevision == attempt.UpstreamExecution
	}, func(result routing.LatencyResult) error { return server.Store.writeLatencyEvidence(id, result) })
	if err != nil {
		server.storageFailure()
	}
}
