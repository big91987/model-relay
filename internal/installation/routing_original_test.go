package installation

import "testing"

func TestLatencyMigrationOldSchemaAndCorruption(t *testing.T) {
	store, _, directory, keyFile := newInstallation(t, false)
	if _, err := store.DB.Exec("DROP TABLE route_latency_evidence;DROP TABLE route_selection_evidence;DROP TABLE model_route_policies;DROP TABLE routing_extension_meta;"); err != nil {
		t.Fatal("prepare old fixture", err)
	}
	store.Close()
	reopened, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal("old format migration", err)
	}
	defer reopened.Close()
	if err = reopened.ensureRouting(); err != nil {
		t.Fatal("idempotent migration", err)
	}
	if _, err = reopened.DB.Exec("DROP TABLE route_latency_evidence"); err != nil {
		t.Fatal(err)
	}
	if err = reopened.ensureRouting(); err == nil {
		t.Fatal("partial extension silently repaired")
	}
}

func TestLatencyPolicyStorageValidation(t *testing.T) {
	store, _, _, _ := newInstallation(t, false)
	transaction, err := store.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err = routingExtension(transaction, false); err != nil {
		t.Fatal(err)
	}
	transaction.Rollback()
}
