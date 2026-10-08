package relay

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type manifest struct {
	Version int    `json:"version"`
	SHA256  string `json:"sha256"`
}

func Backup(directory, output string) error {
	lock, err := lockDirectory(directory)
	if err != nil {
		return err
	}
	defer lock.Close()
	return backupLocked(directory, output)
}

func backupLocked(directory, output string) error {
	if _, err := os.Stat(filepath.Join(directory, "relay.db")); err != nil {
		return errors.New("installation missing")
	}
	db, err := database(directory)
	if err != nil {
		return err
	}
	var version int
	if err = db.QueryRow("SELECT version FROM meta").Scan(&version); err != nil || (version != 1 && version != 2 && version != 3) {
		db.Close()
		return errors.New("unsupported database version")
	}
	var busy, logFrames, checkpointed int
	if err = db.QueryRow("PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &logFrames, &checkpointed); err != nil || busy != 0 {
		db.Close()
		return errors.New("backup checkpoint incomplete")
	}
	if err = db.Close(); err != nil {
		return err
	}
	source, err := os.Open(filepath.Join(directory, "relay.db"))
	if err != nil {
		return err
	}
	defer source.Close()
	info, err := source.Stat()
	if err != nil {
		return err
	}
	hash := sha256.New()
	if _, err = io.Copy(hash, source); err != nil {
		return err
	}
	if _, err = source.Seek(0, io.SeekStart); err != nil {
		return err
	}
	metadata, _ := json.Marshal(manifest{Version: version, SHA256: hex.EncodeToString(hash.Sum(nil))})
	destination, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("backup destination must be new")
	}
	successful := false
	defer func() {
		destination.Close()
		if !successful {
			os.Remove(output)
		}
	}()
	archive := tar.NewWriter(destination)
	if err = archive.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: int64(len(metadata))}); err != nil {
		return err
	}
	if _, err = archive.Write(metadata); err != nil {
		return err
	}
	if err = archive.WriteHeader(&tar.Header{Name: "relay.db", Mode: 0600, Size: info.Size()}); err != nil {
		return err
	}
	if _, err = io.Copy(archive, source); err != nil {
		return err
	}
	if err = archive.Close(); err != nil {
		return err
	}
	if err = destination.Sync(); err != nil {
		return err
	}
	successful = true
	return destination.Close()
}
func Restore(input, directory, keyFile string) error {
	if entries, err := os.ReadDir(directory); err == nil && len(entries) > 0 {
		return errors.New("restore directory must be empty")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(directory), 0700); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp(filepath.Dir(directory), ".restore-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	source, err := os.Open(input)
	if err != nil {
		return err
	}
	defer source.Close()
	archive := tar.NewReader(source)
	seen := map[string]bool{}
	var metadata manifest
	var checksum string
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New("corrupt backup")
		}
		if seen[header.Name] || header.Typeflag != tar.TypeReg || header.Size < 0 || header.Size > 1<<30 {
			return errors.New("invalid backup entry")
		}
		seen[header.Name] = true
		switch header.Name {
		case "manifest.json":
			if header.Size > 4096 {
				return errors.New("manifest too large")
			}
			data, err := io.ReadAll(archive)
			if err != nil || json.Unmarshal(data, &metadata) != nil {
				return errors.New("invalid manifest")
			}
		case "relay.db":
			destination, err := os.OpenFile(filepath.Join(temporary, "relay.db"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				return err
			}
			hash := sha256.New()
			_, err = io.Copy(io.MultiWriter(destination, hash), archive)
			closeErr := destination.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
			checksum = hex.EncodeToString(hash.Sum(nil))
		default:
			return errors.New("unexpected backup entry")
		}
	}
	if !seen["manifest.json"] || !seen["relay.db"] || (metadata.Version != 1 && metadata.Version != 2 && metadata.Version != 3) || checksum != metadata.SHA256 {
		return errors.New("backup checksum or version invalid")
	}
	store, err := openInstallation(temporary, keyFile, true)
	if err != nil {
		return err
	}
	if store.Schema != metadata.Version {
		store.Close()
		return errors.New("backup schema mismatch")
	}
	if err = store.Close(); err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(temporary, ".lock"))
	if entries, err := os.ReadDir(directory); err == nil {
		if len(entries) != 0 {
			return errors.New("restore target changed")
		}
		if err = os.Remove(directory); err != nil {
			return err
		}
	}
	return os.Rename(temporary, directory)
}
