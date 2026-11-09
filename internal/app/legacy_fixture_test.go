package app

import (
	"model-relay/internal/secure"

	"os"
	"path/filepath"
	"testing"
)

const legacyFixtureSchema = `CREATE TABLE meta(version INTEGER NOT NULL, sentinel BLOB NOT NULL);
CREATE TABLE admin(salt BLOB NOT NULL, hash BLOB NOT NULL);
CREATE TABLE upstream(id INTEGER PRIMARY KEY CHECK(id=1), url TEXT NOT NULL, token BLOB NOT NULL, enabled INTEGER NOT NULL, revision INTEGER NOT NULL);
CREATE TABLE models(alias TEXT PRIMARY KEY, model TEXT NOT NULL, enabled INTEGER NOT NULL);
CREATE TABLE app_keys(id TEXT PRIMARY KEY, name TEXT NOT NULL, digest BLOB UNIQUE NOT NULL, hint TEXT NOT NULL, enabled INTEGER NOT NULL, created TEXT NOT NULL);
CREATE TABLE requests(id TEXT PRIMARY KEY, time TEXT NOT NULL, key_id TEXT, alias TEXT, model TEXT, stream INTEGER NOT NULL, revision INTEGER, state TEXT NOT NULL, error_code TEXT NOT NULL, upstream_status INTEGER, duration_ms INTEGER, prompt_tokens INTEGER, completion_tokens INTEGER, total_tokens INTEGER);`

func legacyInstallation(t *testing.T, upstream bool) (string, string) {
	t.Helper()
	root := t.TempDir()
	directory := filepath.Join(root, "legacy")
	keyFile := filepath.Join(root, "material")
	key := secure.Digest("isolated migration material")
	if os.Mkdir(directory, 0700) != nil || os.WriteFile(keyFile, key, 0600) != nil {
		t.Fatal("fixture material")
	}
	db, err := testDatabase(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(legacyFixtureSchema); err != nil {
		t.Fatal(err)
	}
	sentinel, err := secure.Seal(key, "model-relay-v1", "installation")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO meta VALUES(1,?); INSERT INTO admin VALUES(?,?); INSERT INTO models VALUES('alpha','old-a',1),('beta','old-b',0); INSERT INTO app_keys VALUES('key1','old enabled',?,'hint1',1,'old-time'),('key2','old disabled',?,'hint2',0,'old-time'); INSERT INTO requests VALUES('old-id','old-time','key1','alpha','old-a',0,7,'success','',200,19,0,NULL,12)", sentinel, []byte("old-salt"), []byte("old-hash"), secure.Digest("old-key1"), secure.Digest("old-key2")); err != nil {
		t.Fatal(err)
	}
	if upstream {
		encrypted, err := secure.Seal(key, "isolated-old-token", "upstream:https://public.example/v1")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec("INSERT INTO upstream VALUES(1,'https://public.example/v1',?,0,7)", encrypted); err != nil {
			t.Fatal(err)
		}
	}
	return directory, keyFile
}
