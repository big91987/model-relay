package relay

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/argon2"
)

type Store struct {
	DB     *sql.DB
	Key    []byte
	Schema int
	lock   *os.File
	clock  func() time.Time
}

func (store *Store) now() time.Time {
	if store.clock != nil {
		return store.clock().UTC()
	}
	return time.Now().UTC()
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

const legacySchema = `CREATE TABLE meta(version INTEGER NOT NULL, sentinel BLOB NOT NULL);
CREATE TABLE admin(salt BLOB NOT NULL, hash BLOB NOT NULL);
CREATE TABLE upstream(id INTEGER PRIMARY KEY CHECK(id=1), url TEXT NOT NULL, token BLOB NOT NULL, enabled INTEGER NOT NULL, revision INTEGER NOT NULL);
CREATE TABLE models(alias TEXT PRIMARY KEY, model TEXT NOT NULL, enabled INTEGER NOT NULL);
CREATE TABLE app_keys(id TEXT PRIMARY KEY, name TEXT NOT NULL, digest BLOB UNIQUE NOT NULL, hint TEXT NOT NULL, enabled INTEGER NOT NULL, created TEXT NOT NULL);
CREATE TABLE requests(id TEXT PRIMARY KEY, time TEXT NOT NULL, key_id TEXT, alias TEXT, model TEXT, stream INTEGER NOT NULL, revision INTEGER, state TEXT NOT NULL, error_code TEXT NOT NULL, upstream_status INTEGER, duration_ms INTEGER, prompt_tokens INTEGER, completion_tokens INTEGER, total_tokens INTEGER);`

const configurationSchema = `CREATE TABLE upstream(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT UNIQUE NOT NULL, url TEXT NOT NULL, token BLOB NOT NULL, enabled INTEGER NOT NULL CHECK(enabled IN(0,1)), revision INTEGER NOT NULL, execution_revision INTEGER NOT NULL);
CREATE TABLE models(alias TEXT PRIMARY KEY, enabled INTEGER NOT NULL CHECK(enabled IN(0,1)), revision INTEGER NOT NULL, legacy_remote_model TEXT);
CREATE TABLE route_candidates(alias TEXT NOT NULL REFERENCES models(alias), upstream_id INTEGER NOT NULL REFERENCES upstream(id), remote_model TEXT NOT NULL, enabled INTEGER NOT NULL CHECK(enabled IN(0,1)), priority INTEGER NOT NULL CHECK(priority BETWEEN 0 AND 100), weight INTEGER NOT NULL CHECK(weight BETWEEN 1 AND 100), execution_revision INTEGER NOT NULL, PRIMARY KEY(alias,upstream_id));`

const evidenceSchema = `CREATE TABLE upstream_checks(id INTEGER PRIMARY KEY AUTOINCREMENT, upstream_id INTEGER NOT NULL REFERENCES upstream(id), execution_revision INTEGER NOT NULL, state TEXT NOT NULL, started_at TEXT NOT NULL, finished_at TEXT, duration_ms INTEGER, error_code TEXT NOT NULL, upstream_status INTEGER);
CREATE TABLE request_attempts(request_id TEXT NOT NULL REFERENCES requests(id), ordinal INTEGER NOT NULL CHECK(ordinal BETWEEN 1 AND 3), upstream_id INTEGER NOT NULL, upstream_name TEXT NOT NULL, remote_model TEXT NOT NULL, upstream_revision INTEGER NOT NULL, upstream_execution_revision INTEGER NOT NULL, candidate_execution_revision INTEGER NOT NULL, started_at TEXT NOT NULL, finished_at TEXT, duration_ms INTEGER, state TEXT NOT NULL, error_code TEXT NOT NULL, upstream_status INTEGER, prompt_tokens INTEGER, completion_tokens INTEGER, total_tokens INTEGER, PRIMARY KEY(request_id,ordinal));
CREATE TABLE request_route_decisions(request_id TEXT NOT NULL REFERENCES requests(id), selection_round INTEGER NOT NULL, upstream_id INTEGER NOT NULL, snapshot_enabled INTEGER NOT NULL, snapshot_priority INTEGER NOT NULL, snapshot_weight INTEGER NOT NULL, reason TEXT NOT NULL, observed_at TEXT NOT NULL, cooldown_until TEXT, PRIMARY KEY(request_id,selection_round,upstream_id));
CREATE INDEX attempts_upstream ON request_attempts(upstream_id,started_at);
CREATE INDEX checks_upstream ON upstream_checks(upstream_id,id);`

const requestUpgrade = `ALTER TABLE requests ADD COLUMN evidence_version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE requests ADD COLUMN config_revision INTEGER;
ALTER TABLE requests ADD COLUMN model_revision INTEGER;
ALTER TABLE requests ADD COLUMN committed_at TEXT;
ALTER TABLE requests ADD COLUMN successful_upstream_id INTEGER;
ALTER TABLE requests ADD COLUMN attempts_complete INTEGER NOT NULL DEFAULT 0;`

const schema = `CREATE TABLE meta(version INTEGER NOT NULL, sentinel BLOB NOT NULL, config_revision INTEGER NOT NULL DEFAULT 1);
CREATE TABLE admin(salt BLOB NOT NULL, hash BLOB NOT NULL);
CREATE TABLE app_keys(id TEXT PRIMARY KEY, name TEXT NOT NULL, digest BLOB UNIQUE NOT NULL, hint TEXT NOT NULL, enabled INTEGER NOT NULL, created TEXT NOT NULL);
CREATE TABLE requests(id TEXT PRIMARY KEY, time TEXT NOT NULL, key_id TEXT, alias TEXT, model TEXT, stream INTEGER NOT NULL, revision INTEGER, state TEXT NOT NULL, error_code TEXT NOT NULL, upstream_status INTEGER, duration_ms INTEGER, prompt_tokens INTEGER, completion_tokens INTEGER, total_tokens INTEGER);` + configurationSchema + requestUpgrade + evidenceSchema

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
	secret := randomSecret(32)
	copy(key, digest(secret))
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
	sentinel, err := seal(key, "model-relay-v1", "installation")
	if err != nil {
		return "", err
	}
	password := randomSecret(32)
	salt := digest(randomSecret(16))[:16]
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
	plain, err := unseal(key, sentinel, "installation")
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
	salt := digest(randomSecret(16))[:16]
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

type Usage struct {
	Prompt     *int64 `json:"prompt_tokens"`
	Completion *int64 `json:"completion_tokens"`
	Total      *int64 `json:"total_tokens"`
}
type Record struct {
	ID       string `json:"id"`
	Time     string `json:"time"`
	KeyID    string `json:"key_id"`
	Alias    string `json:"alias"`
	Model    string `json:"model"`
	Stream   bool   `json:"stream"`
	Revision int64  `json:"revision"`
	State    string `json:"state"`
	Code     string `json:"error_code"`
	Status   *int   `json:"upstream_status"`
	Duration *int64 `json:"duration_ms"`
	Usage
	ConfigRevision     int64                `json:"config_revision"`
	ModelRevision      int64                `json:"model_revision"`
	SuccessfulUpstream *int64               `json:"successful_upstream_id"`
	CommittedAt        *string              `json:"committed_at"`
	AttemptsComplete   bool                 `json:"attempts_complete"`
	EvidenceVersion    int                  `json:"evidence_version"`
	Attempts           []Attempt            `json:"attempts"`
	Decisions          []RouteDecision      `json:"selection_decisions"`
	UsageSummary       map[string]UsagePart `json:"usage_summary"`
}

func (store *Store) Start(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := store.DB.ExecContext(ctx, "INSERT INTO requests(id,time,key_id,alias,model,stream,revision,state,error_code,evidence_version,config_revision,model_revision) VALUES(?,?,?,?,?,?,?,'started','',2,?,?)", record.ID, record.Time, record.KeyID, record.Alias, record.Model, record.Stream, record.Revision, record.ConfigRevision, record.ModelRevision)
	return err
}
func (store *Store) Finish(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	transaction, err := store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	usage, complete, err := aggregateUsage(transaction, record.ID)
	if err != nil {
		return err
	}
	query := "UPDATE requests SET state=?,error_code=?,model=?,revision=?,upstream_status=?,duration_ms=?,prompt_tokens=?,completion_tokens=?,total_tokens=?,successful_upstream_id=?,committed_at=?,attempts_complete=? WHERE id=?"
	if store.Schema == 3 {
		query += " AND state='started'"
	}
	_, err = transaction.ExecContext(ctx, query, record.State, record.Code, record.Model, record.Revision, record.Status, record.Duration, usage.Prompt, usage.Completion, usage.Total, record.SuccessfulUpstream, record.CommittedAt, complete, record.ID)
	if err != nil {
		return err
	}
	if store.Schema == 3 {
		if err := settleAdmission(transaction, record.ID); err != nil {
			return err
		}
	}
	return transaction.Commit()
}
