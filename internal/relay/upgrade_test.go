package relay

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func legacyInstallation(t *testing.T, upstream bool) (string, string) {
	t.Helper()
	root := t.TempDir()
	directory := filepath.Join(root, "legacy")
	keyFile := filepath.Join(root, "material")
	key := digest("isolated migration material")
	if os.Mkdir(directory, 0700) != nil || os.WriteFile(keyFile, key, 0600) != nil {
		t.Fatal("fixture material")
	}
	db, err := database(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}
	sentinel, err := seal(key, "model-relay-v1", "installation")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO meta VALUES(1,?); INSERT INTO admin VALUES(?,?); INSERT INTO models VALUES('alpha','old-a',1),('beta','old-b',0); INSERT INTO app_keys VALUES('key1','old enabled',?,'hint1',1,'old-time'),('key2','old disabled',?,'hint2',0,'old-time'); INSERT INTO requests VALUES('old-id','old-time','key1','alpha','old-a',0,7,'success','',200,19,0,NULL,12)", sentinel, []byte("old-salt"), []byte("old-hash"), digest("old-key1"), digest("old-key2")); err != nil {
		t.Fatal(err)
	}
	if upstream {
		encrypted, err := seal(key, "isolated-old-token", "upstream:https://public.example/v1")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec("INSERT INTO upstream VALUES(1,'https://public.example/v1',?,0,7)", encrypted); err != nil {
			t.Fatal(err)
		}
	}
	return directory, keyFile
}

func TestExplicitUpgradeConservation(t *testing.T) {
	for _, hasUpstream := range []bool{true, false} {
		t.Run(map[bool]string{true: "mapped", false: "orphan"}[hasUpstream], func(t *testing.T) {
			directory, keyFile := legacyInstallation(t, hasUpstream)
			if store, err := Open(directory, keyFile); err == nil {
				store.Close()
				t.Fatal("serve silently opened schema1")
			}
			archive := filepath.Join(t.TempDir(), "before.tar")
			result, err := Upgrade(directory, keyFile, archive)
			if err != nil || result.Result != "migrated" {
				t.Fatal("upgrade", result, err)
			}
			store, err := openInstallation(directory, keyFile, true)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			var count, revision int
			var salt, hash []byte
			if store.DB.QueryRow("SELECT salt,hash FROM admin").Scan(&salt, &hash) != nil || !bytes.Equal(salt, []byte("old-salt")) || !bytes.Equal(hash, []byte("old-hash")) {
				t.Fatal("admin changed")
			}
			if store.DB.QueryRow("SELECT count(*) FROM app_keys WHERE (id='key1' AND enabled=1) OR (id='key2' AND enabled=0)").Scan(&count) != nil || count != 2 {
				t.Fatal("keys changed")
			}
			if store.DB.QueryRow("SELECT count(*) FROM requests WHERE rowid=1 AND id='old-id' AND time='old-time' AND revision=7 AND duration_ms=19 AND prompt_tokens=0 AND completion_tokens IS NULL AND total_tokens=12 AND evidence_version=1").Scan(&count) != nil || count != 1 {
				t.Fatal("historical values changed")
			}
			if hasUpstream {
				value, err := store.upstreamByID(1)
				if err != nil || value.Enabled || value.Revision != 7 || value.Name != "原有上游" {
					t.Fatal("upstream changed")
				}
				plain, err := unseal(store.Key, value.Encrypted, upstreamAAD(1, value.URL))
				if err != nil || plain != "isolated-old-token" {
					t.Fatal("reseal failed")
				}
				if _, err = unseal(store.Key, value.Encrypted, "upstream:"+value.URL); err == nil {
					t.Fatal("old aad still valid")
				}
			}
			model, err := store.routeModel("alpha")
			if err != nil {
				t.Fatal(err)
			}
			if hasUpstream && (len(model.Candidates) != 1 || model.Candidates[0].RemoteModel != "old-a" || model.Candidates[0].Weight != 1) || !hasUpstream && (len(model.Candidates) != 0 || model.Legacy == nil || *model.Legacy != "old-a") {
				t.Fatal("mapping changed")
			}
			if err = store.DB.QueryRow("SELECT config_revision FROM meta").Scan(&revision); err != nil || revision != 1 {
				t.Fatal("initial config revision")
			}
			store.Close()
			repeat, err := Upgrade(directory, keyFile, archive)
			if err != nil || repeat.Result != "already_current" {
				t.Fatal("repeat", repeat, err)
			}
			restored := filepath.Join(t.TempDir(), "restored")
			if err = Restore(archive, restored, keyFile); err != nil {
				t.Fatal("restore", err)
			}
			old, err := openInstallation(restored, keyFile, true)
			if err != nil || old.Schema != 1 {
				t.Fatal("restored upgrade implicitly", err)
			}
			old.Close()
		})
	}
}

func TestUpgradeRefusalIsNonMigrating(t *testing.T) {
	for _, condition := range []string{"wrong_key", "existing_output", "locked", "unknown_version"} {
		t.Run(condition, func(t *testing.T) {
			directory, keyFile := legacyInstallation(t, true)
			output := filepath.Join(t.TempDir(), "backup.tar")
			switch condition {
			case "wrong_key":
				keyFile = filepath.Join(t.TempDir(), "wrong")
				os.WriteFile(keyFile, digest("wrong"), 0600)
			case "existing_output":
				os.WriteFile(output, []byte("do not overwrite"), 0600)
			case "locked":
				lock, err := lockDirectory(directory)
				if err != nil {
					t.Fatal(err)
				}
				defer lock.Close()
			case "unknown_version":
				db, err := database(directory)
				if err != nil {
					t.Fatal(err)
				}
				db.Exec("UPDATE meta SET version=9")
				db.Close()
			}
			if _, err := Upgrade(directory, keyFile, output); err == nil {
				t.Fatal("unsafe upgrade accepted")
			}
			db, err := database(directory)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			var version int
			db.QueryRow("SELECT version FROM meta").Scan(&version)
			want := 1
			if condition == "unknown_version" {
				want = 9
			}
			if version != want {
				t.Fatal("failed upgrade mutated schema")
			}
			var count int
			if db.QueryRow("SELECT count(*) FROM models WHERE model='old-a'").Scan(&count) != nil || count != 1 {
				t.Fatal("failed upgrade mutated mappings")
			}
		})
	}
}

func TestUpgradeAtomicFailureAndSchemaValidation(t *testing.T) {
	directory, keyFile := legacyInstallation(t, true)
	db, err := database(directory)
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("UPDATE models SET enabled=2 WHERE alias='alpha'")
	db.Close()
	output := filepath.Join(t.TempDir(), "before.tar")
	if _, err = Upgrade(directory, keyFile, output); err == nil {
		t.Fatal("invalid legacy state migrated")
	}
	db, err = database(directory)
	if err != nil {
		t.Fatal(err)
	}
	var version int
	db.QueryRow("SELECT version FROM meta").Scan(&version)
	if version != 1 {
		t.Fatal("non-atomic upgrade")
	}
	db.Close()
	if _, err = os.Stat(output); err != nil {
		t.Fatal("rollback snapshot lost")
	}
	store, _, directory, keyFile := installation(t)
	store.DB.Exec("DROP TABLE request_route_decisions")
	store.Close()
	if opened, err := Open(directory, keyFile); err == nil {
		opened.Close()
		t.Fatal("invalid schema2 opened")
	}
}
