package installation

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newInstallation(t *testing.T, console bool) (*Store, string, string, string) {
	t.Helper()
	root := t.TempDir()
	directory, keyFile := filepath.Join(root, "data"), filepath.Join(root, "material")
	initialize := Init
	if console {
		initialize = InitConsole
	}
	password, err := initialize(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	store, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store, password, directory, keyFile
}

func TestInstallationExclusiveLockAndAdminReset(t *testing.T) {
	store, password, directory, keyFile := newInstallation(t, true)
	if !store.Password(password) || store.Password("wrong") {
		t.Fatal("initial administrator credential")
	}
	if opened, err := Open(directory, keyFile); err == nil {
		opened.Close()
		t.Fatal("active installation opened twice")
	}
	if err := ResetAdminPassword(directory, keyFile, "replacement-password"); err == nil {
		t.Fatal("active installation password changed")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ResetAdminPassword(directory, keyFile, "replacement-password"); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if reopened.Password(password) || !reopened.Password("replacement-password") {
		t.Fatal("password reset did not replace administrator credential")
	}
	var admin, platformSalt, platformHash []byte
	if err := reopened.DB.QueryRow("SELECT hash FROM admin").Scan(&admin); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB.QueryRow("SELECT salt,hash FROM users WHERE id='platform-admin'").Scan(&platformSalt, &platformHash); err != nil || !bytes.Equal(admin, platformHash) || len(platformSalt) != 16 {
		t.Fatal("console administrator password not synchronized", err)
	}
}

func TestBackupRestorePreservesKeyAndDurableClock(t *testing.T) {
	store, password, directory, keyFile := newInstallation(t, true)
	deadline := time.Date(2026, 6, 1, 4, 5, 6, 7, time.UTC)
	store.SetClock(func() time.Time { return deadline })
	if err := store.AdvanceClock(); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "backup.tar")
	if err := Backup(directory, archive); err != nil {
		t.Fatal(err)
	}
	if err := Backup(directory, archive); err == nil {
		t.Fatal("backup overwrote existing destination")
	}
	restored := filepath.Join(t.TempDir(), "restored")
	if err := Restore(archive, restored, keyFile); err != nil {
		t.Fatal(err)
	}
	opened, err := Open(restored, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	if !opened.Now().Equal(deadline) || !opened.Password(password) || !bytes.Equal(opened.Key, store.Key) {
		t.Fatal("restored identity or UTC high water changed")
	}
	opened.SetClock(func() time.Time { return deadline.Add(-time.Hour) })
	if err := opened.AdvanceClock(); err != nil || !opened.Now().Equal(deadline) {
		t.Fatal("restore revived expired time", err)
	}
	if err := Restore(archive, restored, keyFile); err == nil {
		t.Fatal("restore overwrote existing installation")
	}
}

func TestClockObservesWallTimeAfterDatabaseQueue(t *testing.T) {
	store, _, _, _ := newInstallation(t, true)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var wall atomic.Int64
	wall.Store(start.UnixNano())
	observed := make(chan struct{}, 1)
	store.SetClock(func() time.Time {
		select {
		case observed <- struct{}{}:
		default:
		}
		return time.Unix(0, wall.Load())
	})
	transaction, err := store.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	done := make(chan error, 1)
	go func() { done <- store.AdvanceClock() }()
	select {
	case <-observed:
		t.Fatal("wall time sampled before waiting for database")
	case <-time.After(30 * time.Millisecond):
	}
	deadline := start.Add(time.Hour)
	wall.Store(deadline.UnixNano())
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("clock did not resume")
	}
	if !store.Now().Equal(deadline) {
		t.Fatal("queued advance cached stale wall time")
	}
}

func TestFailedClockWriteDoesNotPublishTimeOrWarning(t *testing.T) {
	store, _, _, _ := newInstallation(t, true)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store.SetClock(func() time.Time { return start })
	if err := store.AdvanceClock(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.DB.Exec(`CREATE TRIGGER reject_clock BEFORE UPDATE ON console_clock BEGIN SELECT RAISE(FAIL,'controlled'); END;`); err != nil {
		t.Fatal(err)
	}
	store.SetClock(func() time.Time { return start.Add(-time.Hour) })
	if err := store.AdvanceClock(); err == nil {
		t.Fatal("failed clock update accepted")
	}
	if !store.Now().Equal(start) || store.clockWarned {
		t.Fatal("uncommitted clock/warning published")
	}
	if _, err := store.DB.Exec("DROP TRIGGER reject_clock"); err != nil {
		t.Fatal(err)
	}
	if err := store.AdvanceClock(); err != nil || !store.Now().Equal(start) || !store.clockWarned {
		t.Fatal("retry lost high water or rollback warning", err)
	}
}

func TestConsoleUpgradeRetainsDataAndLegacyPending(t *testing.T) {
	store, password, directory, keyFile := newInstallation(t, false)
	if _, err := store.DB.Exec(`INSERT INTO app_keys VALUES('key','Application',x'1234','hint',1,'2026-01-01T00:00:00Z'); INSERT INTO requests(id,time,key_id,alias,model,stream,state,error_code) VALUES('old','2026-01-01T00:00:00Z','key','alpha','remote',0,'success','');`); err != nil {
		t.Fatal(err)
	}
	store.Close()
	archive := filepath.Join(t.TempDir(), "before-console.tar")
	result, err := UpgradeConsole(directory, keyFile, archive)
	if err != nil || result.FromSchema != 2 || result.Schema != 3 || result.Result != "migrated" {
		t.Fatal("console upgrade", result, err)
	}
	opened, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	var pending, legacy int
	if opened.DB.QueryRow("SELECT count(*) FROM legacy_pending WHERE request_id='old' AND resolved=0").Scan(&pending) != nil || pending != 1 {
		t.Fatal("historical accounting gap lost")
	}
	if opened.DB.QueryRow("SELECT legacy FROM key_controls WHERE key_id='key'").Scan(&legacy) != nil || legacy != 1 || !opened.Password(password) {
		t.Fatal("legacy key or administrator changed")
	}
}

func TestRestoreRejectsWrongKeyAndCorruptionWithoutPublishingTarget(t *testing.T) {
	store, _, directory, keyFile := newInstallation(t, false)
	store.Close()
	archive := filepath.Join(t.TempDir(), "backup.tar")
	if err := Backup(directory, archive); err != nil {
		t.Fatal(err)
	}
	wrongKey := filepath.Join(t.TempDir(), "wrong")
	if err := os.WriteFile(wrongKey, bytes.Repeat([]byte{1}, 32), 0600); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"key", "corrupt"} {
		t.Run(scenario, func(t *testing.T) {
			input, key := archive, wrongKey
			if scenario == "corrupt" {
				input = filepath.Join(t.TempDir(), "invalid.tar")
				key = keyFile
				if err := os.WriteFile(input, []byte(strings.Repeat("bad", 100)), 0600); err != nil {
					t.Fatal(err)
				}
			}
			target := filepath.Join(t.TempDir(), "restored")
			if err := Restore(input, target, key); err == nil {
				t.Fatal("invalid restore accepted")
			}
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				t.Fatal("failed restore published target", err)
			}
		})
	}
}
