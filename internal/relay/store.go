package relay

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/argon2"
)

type Store struct {
	DB   *sql.DB
	Key  []byte
	lock *os.File
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

const schema = `CREATE TABLE meta(version INTEGER NOT NULL, sentinel BLOB NOT NULL);
CREATE TABLE admin(salt BLOB NOT NULL, hash BLOB NOT NULL);
CREATE TABLE upstream(id INTEGER PRIMARY KEY CHECK(id=1), url TEXT NOT NULL, token BLOB NOT NULL, enabled INTEGER NOT NULL, revision INTEGER NOT NULL);
CREATE TABLE models(alias TEXT PRIMARY KEY, model TEXT NOT NULL, enabled INTEGER NOT NULL);
CREATE TABLE app_keys(id TEXT PRIMARY KEY, name TEXT NOT NULL, digest BLOB UNIQUE NOT NULL, hint TEXT NOT NULL, enabled INTEGER NOT NULL, created TEXT NOT NULL);
CREATE TABLE requests(id TEXT PRIMARY KEY, time TEXT NOT NULL, key_id TEXT, alias TEXT, model TEXT, stream INTEGER NOT NULL, revision INTEGER, state TEXT NOT NULL, error_code TEXT NOT NULL, upstream_status INTEGER, duration_ms INTEGER, prompt_tokens INTEGER, completion_tokens INTEGER, total_tokens INTEGER);`

func Init(directory, keyFile string) (string, error) {
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
	if _, err = transaction.Exec("INSERT INTO meta VALUES(1,?); INSERT INTO admin VALUES(?,?)", sentinel, salt, hash); err != nil {
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
	if err = db.QueryRow("SELECT version,sentinel FROM meta").Scan(&version, &sentinel); err != nil || version != 1 {
		store.Close()
		return nil, errors.New("unsupported or invalid database version")
	}
	plain, err := unseal(key, sentinel, "installation")
	if err != nil || plain != "model-relay-v1" {
		store.Close()
		return nil, errors.New("incorrect encryption key")
	}
	var rawURL string
	var token []byte
	err = db.QueryRow("SELECT url,token FROM upstream WHERE id=1").Scan(&rawURL, &token)
	if err != nil && err != sql.ErrNoRows {
		store.Close()
		return nil, err
	}
	if err == nil {
		if _, err = unseal(key, token, "upstream:"+rawURL); err != nil {
			store.Close()
			return nil, errors.New("invalid upstream material")
		}
	}
	if _, err = db.Exec("UPDATE requests SET state='process_interrupted',error_code='process_interrupted',duration_ms=NULL WHERE state='started'"); err != nil {
		store.Close()
		return nil, err
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
}

func (store *Store) Start(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := store.DB.ExecContext(ctx, "INSERT INTO requests(id,time,key_id,alias,model,stream,revision,state,error_code) VALUES(?,?,?,?,?,?,?,'started','')", record.ID, record.Time, record.KeyID, record.Alias, record.Model, record.Stream, record.Revision)
	return err
}
func (store *Store) Finish(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := store.DB.ExecContext(ctx, "UPDATE requests SET state=?,error_code=?,upstream_status=?,duration_ms=?,prompt_tokens=?,completion_tokens=?,total_tokens=? WHERE id=?", record.State, record.Code, record.Status, record.Duration, record.Prompt, record.Completion, record.Total, record.ID)
	return err
}
