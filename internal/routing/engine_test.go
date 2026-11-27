package routing

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLatencyAdoptionFollowsDurableEvidence(t *testing.T) {
	now := time.Unix(100, 0)
	engine := New()
	candidate := Candidate{UpstreamID: 1, RemoteModel: "remote", Enabled: true, Weight: 1, ExecutionRevision: 1, UpstreamExecution: 1}
	snapshot := Snapshot{Alias: "alpha", Revision: 1, ConfigRevision: 1, Epoch: 1, Strategy: "low_latency", Candidates: []Candidate{candidate}}
	finished := now.Format(time.RFC3339Nano)
	sample := Sample{EvaluatedAt: now, Snapshot: snapshot, Candidate: candidate, Ordinal: 1, Sequence: engine.NextSequence(), State: "success", Finished: &finished, CurrentEpoch: 1, Measurement: Measurement{Measured: now, Value: 10}}
	storageFailure := errors.New("controlled evidence failure")
	if err := engine.RecordLatency(context.Background(), sample, func() bool { return true }, func(LatencyResult) error { return storageFailure }); !errors.Is(err, storageFailure) {
		t.Fatal("evidence failure not propagated", err)
	}
	if engine.Stats().Samples != 0 {
		t.Fatal("unretained evidence influenced routing")
	}
	var saved LatencyResult
	if err := engine.RecordLatency(context.Background(), sample, func() bool { return true }, func(result LatencyResult) error { saved = result; return nil }); err != nil {
		t.Fatal(err)
	}
	if saved.Acceptance != "accepted" || engine.Stats().Samples != 1 {
		t.Fatal("retained measurement was not adopted", saved.Acceptance)
	}
}
