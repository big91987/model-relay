package reporting_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"model-relay/internal/fault"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/reporting"
	"model-relay/internal/secure"
)

func legacyOverviewFixture(t *testing.T) (*installation.Store, *reporting.Service) {
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
	return store, reporting.New(store.DB, store.Now)
}

func completeOverviewRecord(t *testing.T, installation *installation.Store, id string) {
	t.Helper()
	journal := records.New(installation.DB, installation.Schema)
	value := records.Record{ID: id, Time: "2026-01-01T00:00:00Z", KeyID: "key", Alias: "alpha", Model: "remote", Revision: 7, ConfigRevision: 1, ModelRevision: 1}
	if err := journal.Start(value); err != nil {
		t.Fatal(err)
	}
	upstream := int64(1)
	value.SuccessfulUpstream = &upstream
	value.State = "success"
	if err := journal.Finish(value); err != nil {
		t.Fatal(err)
	}
}

func TestOverviewInvalidatesOnlyStaleSuccessEvidence(t *testing.T) {
	installation, service := legacyOverviewFixture(t)
	id := "aaaaaaaaaaaaaaaaaaaaaa"
	completeOverviewRecord(t, installation, id)
	value, clearSuccess, err := service.LegacyOverview(context.Background(), id, false)
	if err != nil || clearSuccess || value["call_evidence"] == nil || value["callable_models"] != int64(1) || value["degraded"] != false {
		t.Fatal("valid success evidence missing or invalidated", value, clearSuccess, err)
	}
	upstream := value["upstream"].(map[string]any)
	if upstream["total"] != int64(1) || upstream["configured"] != true || upstream["credential_configured"] != true {
		t.Fatal("overview count/credential shape changed", upstream)
	}
	if _, err := installation.DB.Exec("UPDATE upstream SET revision=8 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	value, clearSuccess, err = service.LegacyOverview(context.Background(), id, true)
	if err != nil || !clearSuccess || value["call_evidence"] != nil || value["degraded"] != true {
		t.Fatal("stale success stayed visible or was not invalidated", value, clearSuccess, err)
	}
}

func TestOverviewStorageFailureDoesNotRequestSuccessInvalidation(t *testing.T) {
	for _, phase := range []string{"begin", "upstream_count", "config_read", "key_count", "record_read", "matching_read"} {
		t.Run(phase, func(t *testing.T) {
			installation, service := legacyOverviewFixture(t)
			id := "aaaaaaaaaaaaaaaaaaaaaa"
			completeOverviewRecord(t, installation, id)
			if phase == "begin" {
				if err := installation.DB.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				statements := map[string]string{
					"upstream_count": "ALTER TABLE upstream RENAME COLUMN enabled TO unavailable_enabled",
					"config_read":    "DROP TABLE meta",
					"key_count":      "ALTER TABLE app_keys RENAME COLUMN enabled TO unavailable_enabled",
					"record_read":    "ALTER TABLE requests RENAME COLUMN duration_ms TO unavailable_duration",
					"matching_read":  "ALTER TABLE route_candidates RENAME COLUMN remote_model TO unavailable_model",
				}
				if _, err := installation.DB.Exec(statements[phase]); err != nil {
					t.Fatal(err)
				}
			}
			value, clearSuccess, err := service.LegacyOverview(context.Background(), id, false)
			var failure *fault.Error
			if !errors.As(err, &failure) || failure.Kind != fault.Unavailable || failure.Code != "storage_unavailable" || value != nil || clearSuccess {
				t.Fatal("unassessed storage failure invalidated success", value, clearSuccess, err)
			}
		})
	}
}

func TestOverviewRequestsInvalidationOnlyForKnownStaleID(t *testing.T) {
	_, service := legacyOverviewFixture(t)
	value, clearSuccess, err := service.LegacyOverview(context.Background(), "", false)
	if err != nil || clearSuccess || value["call_evidence"] != nil {
		t.Fatal("empty success ID requested invalidation", value, clearSuccess, err)
	}
	value, clearSuccess, err = service.LegacyOverview(context.Background(), "missing-success", false)
	if err != nil || !clearSuccess || value["call_evidence"] != nil {
		t.Fatal("known missing success was not invalidated", value, clearSuccess, err)
	}
}
