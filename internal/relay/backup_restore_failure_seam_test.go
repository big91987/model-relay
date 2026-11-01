package relay

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type restoreTargetSnapshot struct {
	info    os.FileInfo
	mode    os.FileMode
	entries []string
	digest  [sha256.Size]byte
}

func validRestoreBackup(t *testing.T) (string, string) {
	t.Helper()
	store, _, directory, keyFile := installation(t)
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "valid.backup")
	if err := Backup(directory, archive); err != nil {
		t.Fatal(err)
	}
	return archive, keyFile
}

func snapshotRestoreTarget(t *testing.T, target string) restoreTargetSnapshot {
	t.Helper()
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(entries))
	hash := sha256.New()
	for index, entry := range entries {
		names[index] = entry.Name()
		fmt.Fprintf(hash, "%s\x00", entry.Name())
	}
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return restoreTargetSnapshot{info: info, mode: info.Mode(), entries: names, digest: digest}
}

func assertNoRestoreTemporary(t *testing.T, parent string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(parent, ".restore-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Errorf("temporary restore directories remain: %v", matches)
	}
}

func TestRestoreFinalRenameFailureSeam(t *testing.T) {
	archive, keyFile := validRestoreBackup(t)
	sentinel := errors.New("restore final rename sentinel")

	for _, existing := range []bool{false, true} {
		name := "absent-target"
		if existing {
			name = "existing-target"
		}
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			target := filepath.Join(parent, "target")
			var before restoreTargetSnapshot
			if existing {
				if err := os.Mkdir(target, 0750); err != nil {
					t.Fatal(err)
				}
				before = snapshotRestoreTarget(t, target)
			}

			calls := 0
			var sourceArgument, targetArgument string
			err := restoreWithRename(archive, target, keyFile, func(source, destination string) error {
				calls++
				sourceArgument, targetArgument = source, destination
				if _, statErr := os.Stat(filepath.Join(source, "relay.db")); statErr != nil {
					t.Errorf("rename injection ran before the validated temporary installation existed: %v", statErr)
				}
				return sentinel
			})
			if !errors.Is(err, sentinel) {
				t.Errorf("restore error %v does not preserve sentinel", err)
			}
			if calls != 1 {
				t.Errorf("rename called %d times, want 1", calls)
			}
			if targetArgument != target {
				t.Errorf("rename target %q, want %q", targetArgument, target)
			}
			if filepath.Dir(sourceArgument) != parent || filepath.Dir(targetArgument) != parent {
				t.Errorf("rename escaped temporary parent: source=%q target=%q parent=%q", sourceArgument, targetArgument, parent)
			}
			if filepath.Base(sourceArgument) == filepath.Base(targetArgument) || len(filepath.Base(sourceArgument)) < len(".restore-") || filepath.Base(sourceArgument)[:len(".restore-")] != ".restore-" {
				t.Errorf("rename source is not the restore temporary directory: %q", sourceArgument)
			}

			if existing {
				after, statErr := os.Stat(target)
				if statErr != nil {
					t.Errorf("existing target was removed after failed rename: %v", statErr)
				} else {
					afterSnapshot := snapshotRestoreTarget(t, target)
					if !os.SameFile(before.info, after) {
						t.Error("existing target identity changed")
					}
					if afterSnapshot.mode != before.mode {
						t.Errorf("existing target mode changed: got %v want %v", afterSnapshot.mode, before.mode)
					}
					if !reflect.DeepEqual(afterSnapshot.entries, before.entries) || afterSnapshot.digest != before.digest {
						t.Errorf("existing target sentinel listing changed: got %v/%x want %v/%x", afterSnapshot.entries, afterSnapshot.digest, before.entries, before.digest)
					}
					if _, relayErr := os.Stat(filepath.Join(target, "relay.db")); !os.IsNotExist(relayErr) {
						t.Errorf("failed restore partially installed relay.db: %v", relayErr)
					}
				}
			} else if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
				t.Errorf("absent target was installed after failed rename: %v", statErr)
			}
			assertNoRestoreTemporary(t, parent)
		})
	}
}

func TestRestorePublicPathStillInstallsValidBackup(t *testing.T) {
	archive, keyFile := validRestoreBackup(t)
	parent := t.TempDir()
	target := filepath.Join(parent, "target")
	if err := Restore(archive, target, keyFile); err != nil {
		t.Fatalf("public Restore failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "relay.db")); err != nil {
		t.Fatalf("public Restore did not install relay.db: %v", err)
	}
	store, err := Open(target, keyFile)
	if err != nil {
		t.Fatalf("public Restore result cannot be opened: %v", err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	assertNoRestoreTemporary(t, parent)
}
