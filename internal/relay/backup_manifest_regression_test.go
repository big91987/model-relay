package relay

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func backupParts(t *testing.T, archivePath string) map[string][]byte {
	t.Helper()
	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	parts := map[string][]byte{}
	reader := tar.NewReader(file)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		parts[header.Name], err = io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
	}
	return parts
}

func backupWithManifest(t *testing.T, parts map[string][]byte, metadata string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "modified.backup")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	writer := tar.NewWriter(file)
	for _, entry := range []struct {
		name string
		data []byte
	}{{"manifest.json", []byte(metadata)}, {"relay.db", parts["relay.db"]}} {
		if err = writer.WriteHeader(&tar.Header{Name: entry.name, Mode: 0600, Size: int64(len(entry.data))}); err == nil {
			_, err = writer.Write(entry.data)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRestoreStrictManifest(t *testing.T) {
	store, _, directory, keyFile := installation(t)
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(t.TempDir(), "valid.backup")
	if err := Backup(directory, archivePath); err != nil {
		t.Fatal(err)
	}
	parts := backupParts(t, archivePath)
	valid := string(parts["manifest.json"])
	restored := filepath.Join(t.TempDir(), "valid-restore")
	if err := Restore(archivePath, restored, keyFile); err != nil {
		t.Fatalf("project backup no longer restores: %v", err)
	}

	checksum := strings.Repeat("0", 64)
	cases := map[string]string{
		"duplicate version":  `{"version":1,"version":1,"sha256":"` + checksum + `"}`,
		"duplicate sha256":   `{"version":1,"sha256":"` + checksum + `","sha256":"` + checksum + `"}`,
		"unknown field":      `{"version":1,"sha256":"` + checksum + `","extra":false}`,
		"trailing value":     valid + ` {}`,
		"wrong version type": `{"version":"1","sha256":"` + checksum + `"}`,
		"wrong digest type":  `{"version":1,"sha256":64}`,
		"empty digest":       `{"version":1,"sha256":""}`,
		"short digest":       `{"version":1,"sha256":"00"}`,
		"uppercase digest":   `{"version":1,"sha256":"` + strings.Repeat("A", 64) + `"}`,
		"non-hex digest":     `{"version":1,"sha256":"` + strings.Repeat("z", 64) + `"}`,
	}
	for name, metadata := range cases {
		t.Run(name, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "restore")
			if err := Restore(backupWithManifest(t, parts, metadata), target, keyFile); err == nil {
				t.Fatal("invalid manifest restored")
			}
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				t.Fatalf("failed restore partially installed target: %v", err)
			}
		})
	}
}

func TestDecodeManifestRequiresExactObject(t *testing.T) {
	checksum := strings.Repeat("0", 64)
	for _, input := range []string{
		`[]`, `null`, `1`, `"manifest"`,
		`{"version":1}`,
		`{"sha256":"` + checksum + `"}`,
	} {
		if _, err := decodeManifest([]byte(input)); err == nil {
			t.Fatalf("accepted %s", bytes.TrimSpace([]byte(input)))
		}
	}
}
