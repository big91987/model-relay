// Package installation owns stopped-installation operations and persistent database lifecycle.
package installation

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/argon2"

	"model-relay/internal/secure"
)

type Store struct {
	DB            *sql.DB
	Key           []byte
	Schema        int
	lock          *os.File
	clock         func() time.Time
	clockMu       sync.RWMutex
	effectiveTime time.Time
	clockWarned   bool
}

func (store *Store) Now() time.Time {
	if store.Schema == 3 {
		store.clockMu.RLock()
		defer store.clockMu.RUnlock()
		return store.effectiveTime
	}
	return store.WallTime()
}

func (store *Store) WallTime() time.Time {
	store.clockMu.RLock()
	clock := store.clock
	store.clockMu.RUnlock()
	if clock != nil {
		return clock().UTC()
	}
	return time.Now().UTC()
}

// SetClock replaces the observed wall clock without altering persisted effective time.
func (store *Store) SetClock(clock func() time.Time) {
	store.clockMu.Lock()
	store.clock = clock
	store.clockMu.Unlock()
}

func lockDirectory(directory string) (*os.File, error) {
	file, err := os.OpenFile(filepath.Join(directory, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		return nil, errors.New("installation is in use")
	}
	return file, nil
}
func database(directory string) (*sql.DB, error) {
	filename, err := filepath.Abs(filepath.Join(directory, "relay.db"))
	if err != nil {
		return nil, err
	}
	destination := url.URL{Scheme: "file", Path: filepath.ToSlash(filename), RawQuery: "_busy_timeout=2000&_foreign_keys=on&_journal_mode=WAL&_synchronous=FULL"}
	db, err := sql.Open("sqlite3", destination.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func Init(directory, keyFile string) (string, error) {
	return initSchema(directory, keyFile, 2)
}

func InitConsole(directory, keyFile string) (string, error) {
	return initSchema(directory, keyFile, 3)
}

func initSchema(directory, keyFile string, version int) (string, error) {
	if filepath.Clean(keyFile) == filepath.Join(filepath.Clean(directory), "relay.db") {
		return "", errors.New("key must be separate")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		return "", errors.New("data directory must be empty")
	}
	if err = os.Chmod(directory, 0700); err != nil {
		return "", err
	}
	lock, err := lockDirectory(directory)
	if err != nil {
		return "", err
	}
	defer lock.Close()
	key := make([]byte, 32)
	secret := secure.Random(32)
	copy(key, secure.Digest(secret))
	keyHandle, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", errors.New("key file must be new")
	}
	if _, err = keyHandle.Write(key); err != nil {
		keyHandle.Close()
		return "", err
	}
	if err = keyHandle.Close(); err != nil {
		return "", err
	}
	db, err := database(directory)
	if err != nil {
		return "", err
	}
	defer db.Close()
	transaction, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer transaction.Rollback()
	if _, err = transaction.Exec(schema); err != nil {
		return "", err
	}
	sentinel, err := secure.Seal(key, "model-relay-v1", "installation")
	if err != nil {
		return "", err
	}
	password := secure.Random(32)
	salt := secure.Digest(secure.Random(16))[:16]
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 1, 32)
	if _, err = transaction.Exec("INSERT INTO meta(version,sentinel) VALUES(?,?); INSERT INTO admin VALUES(?,?)", version, sentinel, salt, hash); err != nil {
		return "", err
	}
	if version == 3 {
		if err = createConsole(transaction); err != nil {
			return "", err
		}
		if _, err := transaction.Exec("DELETE FROM resource_limits WHERE owner_type='tenant' AND owner_id='legacy'; DELETE FROM memberships WHERE tenant_id='legacy'; DELETE FROM tenants WHERE id='legacy'"); err != nil {
			return "", err
		}
	}
	if err = routingExtension(transaction, true); err != nil {
		return "", err
	}
	if err = transaction.Commit(); err != nil {
		return "", err
	}
	if err = os.Chmod(filepath.Join(directory, "relay.db"), 0600); err != nil {
		return "", err
	}
	return password, nil
}
func Open(directory, keyFile string) (*Store, error) {
	return openInstallation(directory, keyFile, false)
}

func openInstallation(directory, keyFile string, allowLegacy bool) (*Store, error) {
	if _, err := os.Stat(filepath.Join(directory, "relay.db")); err != nil {
		return nil, errors.New("installation missing")
	}
	key, err := os.ReadFile(keyFile)
	if err != nil || len(key) != 32 {
		return nil, errors.New("key file invalid")
	}
	lock, err := lockDirectory(directory)
	if err != nil {
		return nil, err
	}
	db, err := database(directory)
	if err != nil {
		lock.Close()
		return nil, err
	}
	store := &Store{DB: db, Key: key, lock: lock}
	var version int
	var sentinel []byte
	if err = db.QueryRow("SELECT version,sentinel FROM meta").Scan(&version, &sentinel); err != nil || (version != 1 && version != 2 && version != 3) {
		store.Close()
		return nil, errors.New("unsupported or invalid database version")
	}
	store.Schema = version
	if version == 1 && !allowLegacy {
		store.Close()
		return nil, errors.New("storage schema 1 requires explicit upgrade --data-dir --key-file --output before serve")
	}
	plain, err := secure.Unseal(key, sentinel, "installation")
	if err != nil || plain != "model-relay-v1" {
		store.Close()
		return nil, errors.New("incorrect encryption key")
	}
	if err = validateMaterial(db, key, version); err != nil {
		store.Close()
		return nil, err
	}
	if err = validateStorage(db, version); err != nil {
		store.Close()
		return nil, err
	}
	if err = store.ensureRouting(); err != nil {
		store.Close()
		return nil, err
	}
	if allowLegacy {
		return store, nil
	}
	if _, err = db.Exec("UPDATE requests SET state='process_interrupted',error_code='process_interrupted',duration_ms=NULL,attempts_complete=0 WHERE state='started'; UPDATE request_attempts SET state='process_interrupted',error_code='process_interrupted' WHERE finished_at IS NULL; UPDATE upstream_checks SET state='unknown',error_code='process_interrupted' WHERE finished_at IS NULL"); err != nil {
		store.Close()
		return nil, err
	}
	if version == 3 {
		if err = store.recoverConsole(); err != nil {
			store.Close()
			return nil, err
		}
	}
	return store, nil
}
func (store *Store) Close() error {
	err := store.DB.Close()
	lockErr := store.lock.Close()
	if err != nil {
		return err
	}
	return lockErr
}

// ResetAdminPassword requires the stopped installation's exclusive lock and
// encryption key. Restarting the service also discards existing login sessions.
func ResetAdminPassword(directory, keyFile, password string) error {
	if len(password) > 1024 || strings.TrimSpace(password) == "" || strings.ContainsAny(password, "\r\n\x00") || !utf8.ValidString(password) {
		return errors.New("password must be one nonempty UTF-8 line of at most 1024 bytes")
	}
	store, err := Open(directory, keyFile)
	if err != nil {
		return err
	}
	defer store.Close()
	transaction, err := store.DB.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	salt := secure.Digest(secure.Random(16))[:16]
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 1, 32)
	result, err := transaction.Exec("UPDATE admin SET salt=?,hash=?", salt, hash)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return errors.New("installation must contain exactly one administrator")
	}
	if store.Schema == 3 {
		if _, err := transaction.Exec("UPDATE users SET salt=?,hash=? WHERE id='platform-admin'", salt, hash); err != nil {
			return err
		}
	}
	return transaction.Commit()
}

func (store *Store) Password(value string) bool {
	var salt, hash []byte
	if store.DB.QueryRow("SELECT salt,hash FROM admin").Scan(&salt, &hash) != nil {
		return false
	}
	actual := argon2.IDKey([]byte(value), salt, 3, 64*1024, 1, 32)
	return subtle.ConstantTimeCompare(hash, actual) == 1
}
