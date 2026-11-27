package apikey_test

import (
	"context"
	"testing"

	"model-relay/internal/fault"
	"model-relay/internal/routing"
)

func seedLegacyRuntime(t *testing.T, f *fixture, successID string) uint64 {
	t.Helper()
	f.state.Lock()
	epoch := f.state.Epoch()
	f.state.CompleteRequest(successID, epoch, true)
	f.state.Unlock()
	candidate := routing.Candidate{UpstreamID: 1, RemoteModel: "remote", Enabled: true, ExecutionRevision: 1, UpstreamExecution: 1}
	sample := routing.Sample{Snapshot: routing.Snapshot{Alias: "alpha", Epoch: epoch}, Candidate: candidate, Sequence: 1, State: "success", CurrentEpoch: epoch, EvaluatedAt: f.now, Measurement: routing.Measurement{Measured: f.now, Value: 1}}
	if err := f.router.RecordLatency(context.Background(), sample, func() bool { return true }, func(routing.LatencyResult) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if f.router.Stats().Samples != 1 {
		t.Fatal("runtime fixture did not contain a routing sample")
	}
	return epoch
}

func assertLegacyRuntime(t *testing.T, f *fixture, epoch uint64, successID string, samples int) {
	t.Helper()
	f.state.RLock()
	defer f.state.RUnlock()
	if f.state.Epoch() != epoch || f.state.SuccessID() != successID || f.router.Stats().Samples != samples {
		t.Fatalf("runtime epoch=%d success=%q samples=%d; want %d/%q/%d", f.state.Epoch(), f.state.SuccessID(), f.router.Stats().Samples, epoch, successID, samples)
	}
}

func TestLegacyMutationsPublishCommittedRuntimeWithoutHTTPAdapter(t *testing.T) {
	f := newFixture(t)
	if _, err := f.db.Exec("CREATE TABLE meta(config_revision INTEGER); INSERT INTO meta VALUES(1)"); err != nil {
		t.Fatal(err)
	}
	epoch := seedLegacyRuntime(t, f, "before-create")
	issued, err := f.service.CreateLegacy(context.Background(), "Application")
	if err != nil {
		t.Fatal(err)
	}
	assertLegacyRuntime(t, f, epoch+1, "", 0)
	epoch = seedLegacyRuntime(t, f, "before-disable")
	if err := f.service.SetLegacyEnabled(context.Background(), issued.Metadata.ID, false); err != nil {
		t.Fatal(err)
	}
	assertLegacyRuntime(t, f, epoch+1, "", 0)
}

func TestFailedLegacyRevisionKeepsCommittedRuntime(t *testing.T) {
	for _, action := range []string{"create", "disable"} {
		t.Run(action, func(t *testing.T) {
			f := newFixture(t)
			if _, err := f.db.Exec("CREATE TABLE meta(config_revision INTEGER); INSERT INTO meta VALUES(1)"); err != nil {
				t.Fatal(err)
			}
			var keyID string
			if action == "disable" {
				issued, err := f.service.CreateLegacy(context.Background(), "Application")
				if err != nil {
					t.Fatal(err)
				}
				keyID = issued.Metadata.ID
			}
			epoch := seedLegacyRuntime(t, f, "committed-evidence")
			if _, err := f.db.Exec("CREATE TRIGGER reject_legacy_revision BEFORE UPDATE ON meta BEGIN SELECT RAISE(ABORT,'controlled revision failure'); END"); err != nil {
				t.Fatal(err)
			}
			var err error
			if action == "create" {
				_, err = f.service.CreateLegacy(context.Background(), "Rejected application")
			} else {
				err = f.service.SetLegacyEnabled(context.Background(), keyID, false)
			}
			expectError(t, err, fault.Unavailable, "storage_unavailable")
			assertLegacyRuntime(t, f, epoch, "committed-evidence", 1)
			var keys, revision int
			if err := f.db.QueryRow("SELECT count(*) FROM app_keys WHERE enabled=1").Scan(&keys); err != nil {
				t.Fatal(err)
			}
			if err := f.db.QueryRow("SELECT config_revision FROM meta").Scan(&revision); err != nil {
				t.Fatal(err)
			}
			wantKeys, wantRevision := 0, 1
			if action == "disable" {
				wantKeys, wantRevision = 1, 2
			}
			if keys != wantKeys || revision != wantRevision {
				t.Fatal("failed mutation changed persisted key or revision", keys, revision)
			}
		})
	}
}
