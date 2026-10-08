package relay

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"
)

const consoleClockTable = "CREATE TABLE IF NOT EXISTS console_clock(id INTEGER PRIMARY KEY CHECK(id=1), effective TEXT NOT NULL);"

func (server *Server) advanceClock(writer http.ResponseWriter) bool {
	if server.Store.advanceClock() == nil {
		return true
	}
	server.storageFailure()
	failure(writer, 503, "storage_unavailable", "")
	return false
}

func initializeConsoleClock(transaction *sql.Tx) error {
	if _, err := transaction.Exec(consoleClockTable); err != nil {
		return err
	}
	var count int
	if err := transaction.QueryRow("SELECT count(*) FROM console_clock").Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return nil
	}
	latest := time.Time{}
	rows, err := transaction.Query("SELECT accepted FROM admissions")
	if err != nil {
		return err
	}
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			rows.Close()
			return err
		}
		observed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			rows.Close()
			return errors.New("invalid persisted admission time")
		}
		if observed.After(latest) {
			latest = observed
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	_, err = transaction.Exec("INSERT INTO console_clock VALUES(1,?)", latest.UTC().Format(time.RFC3339Nano))
	return err
}

func (store *Store) advanceClock() error {
	if store.Schema != 3 {
		return nil
	}
	transaction, err := store.DB.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	var text string
	if err := transaction.QueryRow("SELECT effective FROM console_clock WHERE id=1").Scan(&text); err != nil {
		return err
	}
	previous, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return errors.New("invalid effective clock")
	}
	observed := store.wallTime()
	drift := previous.Sub(observed)
	if previous.After(observed) {
		observed = previous
	}
	if _, err := transaction.Exec("UPDATE console_clock SET effective=? WHERE id=1", observed.UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	store.clockMu.Lock()
	warn := drift > time.Minute && !store.clockWarned
	if warn {
		store.clockWarned = true
	}
	if observed.After(store.effectiveTime) {
		store.effectiveTime = observed
	}
	store.clockMu.Unlock()
	if warn {
		log.Print("effective_clock_rollback: UTC high water retained; verify host time synchronization; do not reset database windows")
	}
	return nil
}
