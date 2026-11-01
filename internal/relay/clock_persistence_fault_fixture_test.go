package relay

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
)

type clockFaultStage uint8

const (
	clockFaultNone clockFaultStage = iota
	clockFaultBegin
	clockFaultRead
	clockFaultUpdate
	clockFaultCommit
)

type clockFaultCounts struct {
	begin  atomic.Int64
	read   atomic.Int64
	update atomic.Int64
	commit atomic.Int64
}

type clockFaultState struct {
	stage          clockFaultStage
	counts         clockFaultCounts
	commitHookHits atomic.Int64
	err            error
}

type clockFaultConnector struct {
	dsn   string
	state *clockFaultState
}

func (connector *clockFaultConnector) Connect(context.Context) (driver.Conn, error) {
	base := &sqlite3.SQLiteDriver{}
	connection, err := base.Open(connector.dsn)
	if err != nil {
		return nil, err
	}
	sqliteConnection := connection.(*sqlite3.SQLiteConn)
	if connector.state.stage == clockFaultCommit {
		sqliteConnection.RegisterCommitHook(func() int {
			connector.state.commitHookHits.Add(1)
			return 1
		})
	}
	return &clockFaultConn{Conn: sqliteConnection, state: connector.state}, nil
}

func (connector *clockFaultConnector) Driver() driver.Driver { return &sqlite3.SQLiteDriver{} }

type clockFaultConn struct {
	driver.Conn
	state *clockFaultState
}

func (connection *clockFaultConn) Begin() (driver.Tx, error) {
	return connection.begin(func() (driver.Tx, error) { return connection.Conn.Begin() })
}

func (connection *clockFaultConn) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	return connection.begin(func() (driver.Tx, error) {
		return connection.Conn.(driver.ConnBeginTx).BeginTx(ctx, options)
	})
}

func (connection *clockFaultConn) begin(delegate func() (driver.Tx, error)) (driver.Tx, error) {
	connection.state.counts.begin.Add(1)
	if connection.state.stage == clockFaultBegin {
		return nil, connection.state.err
	}
	transaction, err := delegate()
	if err != nil {
		return nil, err
	}
	return &clockFaultTx{Tx: transaction, state: connection.state}, nil
}

func (connection *clockFaultConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if strings.HasPrefix(strings.TrimSpace(query), "SELECT effective FROM console_clock") {
		connection.state.counts.read.Add(1)
		if connection.state.stage == clockFaultRead {
			return nil, connection.state.err
		}
	}
	return connection.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
}

func (connection *clockFaultConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if strings.HasPrefix(strings.TrimSpace(query), "UPDATE console_clock SET effective=") {
		connection.state.counts.update.Add(1)
		if connection.state.stage == clockFaultUpdate {
			return nil, connection.state.err
		}
	}
	return connection.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
}

type clockFaultTx struct {
	driver.Tx
	state *clockFaultState
}

func (transaction *clockFaultTx) Commit() error {
	transaction.state.counts.commit.Add(1)
	return transaction.Tx.Commit()
}

type clockFaultFixture struct {
	store *Store
	state *clockFaultState
}

func newClockFaultFixture(t *testing.T, stage clockFaultStage) *clockFaultFixture {
	t.Helper()
	root := t.TempDir()
	directory := filepath.Join(root, "data")
	keyFile := filepath.Join(root, "material.key")
	if _, err := InitConsole(directory, keyFile); err != nil {
		t.Fatal(err)
	}
	store, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DB.Close(); err != nil {
		store.lock.Close()
		t.Fatal(err)
	}
	state := &clockFaultState{stage: stage, err: fmt.Errorf("clock persistence %d: %w", stage, errClockFault)}
	filename, err := filepath.Abs(filepath.Join(directory, "relay.db"))
	if err != nil {
		store.lock.Close()
		t.Fatal(err)
	}
	destination := url.URL{Scheme: "file", Path: filepath.ToSlash(filename), RawQuery: "_busy_timeout=2000&_foreign_keys=on&_journal_mode=WAL&_synchronous=FULL"}
	store.DB = sql.OpenDB(&clockFaultConnector{dsn: destination.String(), state: state})
	store.DB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = store.Close() })
	return &clockFaultFixture{store: store, state: state}
}

var errClockFault = errors.New("deterministic clock persistence fault")

func TestClockPersistenceFaultFixture(t *testing.T) {
	tests := []struct {
		name   string
		stage  clockFaultStage
		counts [4]int64
	}{
		{name: "normal delegation", stage: clockFaultNone, counts: [4]int64{1, 1, 1, 1}},
		{name: "begin", stage: clockFaultBegin, counts: [4]int64{1, 0, 0, 0}},
		{name: "read console_clock", stage: clockFaultRead, counts: [4]int64{1, 1, 0, 0}},
		{name: "update console_clock", stage: clockFaultUpdate, counts: [4]int64{1, 1, 1, 0}},
		{name: "commit", stage: clockFaultCommit, counts: [4]int64{1, 1, 1, 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newClockFaultFixture(t, test.stage)
			observed := time.Date(2026, 11, 1, 12, 0, int(test.stage), 0, time.UTC)
			fixture.store.clock = func() time.Time { return observed }

			err := fixture.store.advanceClock()
			got := [4]int64{
				fixture.state.counts.begin.Load(),
				fixture.state.counts.read.Load(),
				fixture.state.counts.update.Load(),
				fixture.state.counts.commit.Load(),
			}
			if test.stage == clockFaultNone {
				if err != nil {
					t.Fatalf("normal advanceClock: %v", err)
				}
				var persisted string
				if err := fixture.store.DB.QueryRow("SELECT effective FROM console_clock WHERE id=1").Scan(&persisted); err != nil {
					t.Fatal(err)
				}
				if persisted != observed.Format(time.RFC3339Nano) {
					t.Fatalf("persisted effective = %q, want %q", persisted, observed.Format(time.RFC3339Nano))
				}
			} else if test.stage == clockFaultCommit {
				var sqliteError sqlite3.Error
				if !errors.As(err, &sqliteError) || sqliteError.ExtendedCode != sqlite3.ErrConstraintCommitHook {
					t.Fatalf("commit error = %#v, want SQLITE_CONSTRAINT_COMMITHOOK", err)
				}
				if hits := fixture.state.commitHookHits.Load(); hits != 1 {
					t.Fatalf("SQLite commit hook hits = %d, want 1", hits)
				}
			} else if !errors.Is(err, errClockFault) {
				t.Fatalf("stage %d error = %v, want sentinel", test.stage, err)
			}

			if got != test.counts {
				t.Fatalf("begin/read/update/commit counts = %v, want %v", got, test.counts)
			}
			if test.stage != clockFaultCommit && fixture.state.commitHookHits.Load() != 0 {
				t.Fatal("SQLite commit hook unexpectedly installed or invoked")
			}
		})
	}
}
