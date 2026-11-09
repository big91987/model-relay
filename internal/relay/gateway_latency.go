package relay

import (
	"context"
	"io"
	"model-relay/internal/catalog"
	"model-relay/internal/records"
	"model-relay/internal/routing"
	"time"
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

func (server *Gateway) routingNow() time.Time {
	if server.Config.LatencyClock != nil {
		return server.Config.LatencyClock()
	}
	return time.Now()
}

func (server *Gateway) acceptLatency(ctx context.Context, id string, snapshot catalog.Snapshot, candidate catalog.Candidate, attempt records.Attempt, stream bool, measurement latencyMeasurement) {
	server.State.Lock()
	defer server.State.Unlock()
	sample := routing.Sample{Snapshot: routingSnapshot(snapshot, time.Time{}), Candidate: routingCandidate(candidate, snapshot.Upstreams[candidate.UpstreamID]), Ordinal: attempt.Ordinal, Sequence: attempt.Sequence, State: attempt.State, Finished: attempt.Finished, CurrentEpoch: server.State.Epoch(), Degraded: server.Degraded.Load(), Stream: stream, EvaluatedAt: server.routingNow(), Measurement: routing.Measurement{Measured: measurement.Measured, Value: measurement.Value}}
	err := server.Router.RecordLatency(ctx, sample, func() bool {
		current, err := server.Catalog.ModelFacts(context.Background(), snapshot.Model.Alias)
		upstream, upstreamErr := server.Catalog.UpstreamFacts(context.Background(), candidate.UpstreamID)
		return err == nil && upstreamErr == nil && current.Enabled && upstream.Enabled && current.Revision == snapshot.Model.Revision && upstream.ExecutionRevision == attempt.UpstreamExecution
	}, func(result routing.LatencyResult) error { return server.Records.WriteLatencyEvidence(id, result) })
	if err != nil {
		server.storageFailure()
	}
}
