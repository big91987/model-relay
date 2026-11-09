package records_test

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"testing"

	"model-relay/internal/fault"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/routing"
	"model-relay/internal/secure"
)

func legacyFixture(t *testing.T) (*installation.Store, *records.Store) {
	t.Helper()
	root := t.TempDir()
	directory, keyFile := filepath.Join(root, "data"), filepath.Join(root, "key")
	if _, err := installation.Init(directory, keyFile); err != nil {
		t.Fatal(err)
	}
	store, err := installation.Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	encrypted, err := secure.Seal(store.Key, "controlled-token", secure.UpstreamAAD(1, "https://public.example/v1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.DB.Exec(`INSERT INTO upstream VALUES(1,'Provider','https://public.example/v1',?,1,7,1);
INSERT INTO models VALUES('alpha',1,1,NULL);
INSERT INTO route_candidates VALUES('alpha',1,'remote',1,0,1,1);
INSERT INTO app_keys VALUES('key','Application',?,'hint',1,'2026-01-01T00:00:00Z');`, encrypted, secure.Digest("credential")); err != nil {
		t.Fatal(err)
	}
	return store, records.New(store.DB, store.Schema)
}

func completeRecord(t *testing.T, store *records.Store, id string) records.Record {
	t.Helper()
	value := records.Record{ID: id, Time: "2026-01-01T00:00:00Z", KeyID: "key", Alias: "alpha", Model: "remote", Revision: 7, ConfigRevision: 1, ModelRevision: 1}
	if err := store.Start(value); err != nil {
		t.Fatal(err)
	}
	finished := "2026-01-01T00:00:01Z"
	zero, total := int64(0), int64(5)
	attempt := records.Attempt{Ordinal: 1, UpstreamID: 1, UpstreamName: "Provider", RemoteModel: "remote", UpstreamRevision: 7, UpstreamExecution: 1, CandidateExecution: 1, Started: value.Time}
	if err := store.BeginAttempt(id, 1, []routing.RouteDecision{{Round: 1, UpstreamID: 1, Enabled: true, Weight: 1, Reason: "selected", Observed: value.Time}}, []routing.SelectionObservation{{Round: 1, UpstreamID: 1, Strategy: "weighted", Reason: "selected"}}, &attempt, func() int64 { return 1 }); err != nil {
		t.Fatal(err)
	}
	attempt.Finished = &finished
	attempt.State = "success"
	attempt.Usage = records.Usage{Prompt: &zero, Total: &total}
	if err := store.FinishAttempt(id, &attempt); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteLatencyEvidence(id, routing.LatencyResult{Ordinal: 1, UpstreamID: 1, Acceptance: "accepted"}); err != nil {
		t.Fatal(err)
	}
	upstream := int64(1)
	value.SuccessfulUpstream = &upstream
	value.State = "success"
	value.CommittedAt = &finished
	if err := store.Finish(value); err != nil {
		t.Fatal(err)
	}
	return value
}

func TestDetailIncludesAttemptsRoutingAndUnknownUsage(t *testing.T) {
	_, store := legacyFixture(t)
	id := "aaaaaaaaaaaaaaaaaaaaaa"
	completeRecord(t, store, id)
	value, err := store.Detail(context.Background(), id)
	if err != nil || value.ID != id || value.Prompt == nil || *value.Prompt != 0 || value.Completion != nil || value.Total == nil || *value.Total != 5 || len(value.Attempts) != 1 || len(value.Decisions) != 1 || len(value.RoutingEvidence) != 1 || len(value.LatencyEvidence) != 1 || !value.AttemptsComplete || value.SuccessfulUpstream == nil || *value.SuccessfulUpstream != 1 {
		t.Fatal("complete record evidence changed", value, err)
	}
	part := value.UsageSummary["completion_tokens"]
	if part.Completeness != "unknown" || part.UnknownAttempts != 1 || part.Total != nil {
		t.Fatal("unknown attempt usage fabricated", part)
	}
}

func TestLegacyListUsesRowCursorAndKeepsNullableCounters(t *testing.T) {
	installation, store := legacyFixture(t)
	for _, id := range []string{"aaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbb", "cccccccccccccccccccccc"} {
		if err := store.Start(records.Record{ID: id, Time: "2026-01-01T00:00:00Z"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := installation.DB.Exec("UPDATE requests SET prompt_tokens=0 WHERE id='cccccccccccccccccccccc'"); err != nil {
		t.Fatal(err)
	}
	values, next, err := store.LegacyList(context.Background(), 1<<62, 2)
	if err != nil || len(values) != 2 || values[0].ID != "cccccccccccccccccccccc" || values[1].ID != "bbbbbbbbbbbbbbbbbbbbbb" || next != "2" || values[0].Prompt == nil || *values[0].Prompt != 0 || values[0].Completion != nil {
		t.Fatal("list cursor/null contract changed", values, next, err)
	}
	values, next, err = store.LegacyList(context.Background(), 2, 2)
	if err != nil || len(values) != 1 || values[0].ID != "aaaaaaaaaaaaaaaaaaaaaa" || next != nil {
		t.Fatal("next page changed", values, next, err)
	}
}

func TestDetailPreservesHistoricalNullEvidenceAndRejectsCorruption(t *testing.T) {
	installation, store := legacyFixture(t)
	id := "aaaaaaaaaaaaaaaaaaaaaa"
	completeRecord(t, store, id)
	if _, err := installation.DB.Exec("UPDATE requests SET evidence_version=1 WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	value, err := store.Detail(context.Background(), id)
	if err != nil || value.Attempts != nil || value.Decisions != nil || value.UsageSummary != nil || value.RoutingEvidence != nil || value.LatencyEvidence != nil {
		t.Fatal("historical record synthesized later evidence", value, err)
	}
	if _, err := installation.DB.Exec("UPDATE requests SET evidence_version=2 WHERE id=?; UPDATE route_latency_evidence SET evidence='{' WHERE request_id=?", id, id); err != nil {
		t.Fatal(err)
	}
	_, err = store.Detail(context.Background(), id)
	var failure *fault.Error
	if !errors.As(err, &failure) || failure.Kind != fault.Unavailable {
		t.Fatal("corrupt evidence exposed partial detail", err)
	}
	_, err = store.Detail(context.Background(), "missing")
	if !errors.As(err, &failure) || failure.Kind != fault.NotFound {
		t.Fatal("invalid record id exposed storage details", err)
	}
}

func TestAttemptSequenceWaitsForDecisionPersistence(t *testing.T) {
	installation, store := legacyFixture(t)
	id := "aaaaaaaaaaaaaaaaaaaaaa"
	if err := store.Start(records.Record{ID: id, Time: "2026-01-01T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	attempt := records.Attempt{Ordinal: 1, UpstreamID: 1, UpstreamName: "Provider", RemoteModel: "remote", Started: "2026-01-01T00:00:00Z"}
	var sequences int64
	next := func() int64 { sequences++; return sequences }
	nan := math.NaN()
	err := store.BeginAttempt(id, 1, []routing.RouteDecision{{Round: 1, UpstreamID: 1, Weight: 1}}, []routing.SelectionObservation{{Round: 1, UpstreamID: 1, EWMA: &nan}}, &attempt, next)
	if err == nil || sequences != 0 {
		t.Fatal("invalid observation consumed sequence or reported success", sequences, err)
	}
	var persisted int
	if installation.DB.QueryRow("SELECT count(*) FROM request_route_decisions WHERE request_id=?", id).Scan(&persisted) != nil || persisted != 0 {
		t.Fatal("failed decision transaction was not rolled back")
	}
	err = store.BeginAttempt(id, 1, []routing.RouteDecision{{Round: 1, UpstreamID: 1, Weight: 1}}, nil, &attempt, next)
	if err != nil || sequences != 1 || attempt.Sequence != 1 {
		t.Fatal("successful attempt did not reserve sequence once", attempt.Sequence, sequences, err)
	}
}
