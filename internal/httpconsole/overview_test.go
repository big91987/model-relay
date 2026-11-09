package httpconsole

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"

	"model-relay/internal/apikey"
	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/reporting"
	"model-relay/internal/routing"
	"model-relay/internal/secure"
	"model-relay/internal/sessions"
	"model-relay/internal/upstream"
)

func overviewControllerFixture(t *testing.T) (*Controller, *installation.Store, *http.Cookie, string) {
	t.Helper()
	root := t.TempDir()
	directory, keyFile := filepath.Join(root, "data"), filepath.Join(root, "material")
	if _, err := installation.Init(directory, keyFile); err != nil {
		t.Fatal(err)
	}
	store, err := installation.Open(directory, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	state, router := configstate.New(), routing.New()
	models := catalog.New(store.DB, store.Key, store.Schema, state, router, nil, upstream.Policy{})
	enabled, priority, weight := true, 0, 1
	source, err := models.CreateUpstream(context.Background(), "", catalog.UpstreamInput{Name: "Overview provider", URL: "https://8.8.8.8/v1", Token: "controlled-token", Enabled: &enabled})
	if err != nil {
		t.Fatal(err)
	}
	candidates := []catalog.CandidateInput{{UpstreamID: source.ID, RemoteModel: "remote", Enabled: &enabled, Priority: &priority, Weight: &weight}}
	if _, err = models.CreateModel(context.Background(), "", catalog.ModelInput{Alias: "alpha", Enabled: &enabled, Candidates: &candidates}); err != nil {
		t.Fatal(err)
	}
	issued, err := apikey.New(store.DB, store.Now, state, router).CreateLegacy(context.Background(), "Overview application")
	if err != nil {
		t.Fatal(err)
	}
	state.Lock()
	snapshot, err := models.ModelSnapshot(context.Background(), "alpha")
	state.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	journal := records.New(store.DB, store.Schema)
	id := secure.Random(16)
	record := records.Record{ID: id, Time: store.Now().Format(time.RFC3339Nano), KeyID: issued.Metadata.ID, Alias: "alpha", Model: "remote", Revision: source.Revision, ConfigRevision: snapshot.ConfigRevision, ModelRevision: snapshot.Model.Revision}
	if err = journal.Start(record); err != nil {
		t.Fatal(err)
	}
	record.State = "success"
	record.SuccessfulUpstream = &source.ID
	if err = journal.Finish(record); err != nil {
		t.Fatal(err)
	}
	state.Lock()
	state.CompleteRequest(id, state.Epoch(), true)
	state.Unlock()
	sessionStore := sessions.New()
	secret := secure.Random(32)
	sessionStore.Create(time.Now(), secret, "", secure.Random(32))
	controller := &Controller{Schema: store.Schema, Authority: func() string { return "overview.test" }, Records: journal, Reports: reporting.New(store.DB, store.Now), State: state, Degraded: &atomic.Bool{}, Sessions: sessionStore}
	return controller, store, &http.Cookie{Name: "mr_session", Value: secret}, id
}

func overviewRequest(t *testing.T, controller *Controller, cookie *http.Cookie) (*httptest.ResponseRecorder, map[string]json.RawMessage) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "http://overview.test/admin/api/overview", nil)
	request.AddCookie(cookie)
	writer := httptest.NewRecorder()
	controller.ServeHTTP(writer, request)
	var value map[string]json.RawMessage
	if err := json.Unmarshal(writer.Body.Bytes(), &value); err != nil {
		t.Fatal(err)
	}
	return writer, value
}

func TestOverviewStorageFailureRetainsSuccessfulRuntimeID(t *testing.T) {
	for _, phase := range []string{"closed_database", "failed_query"} {
		t.Run(phase, func(t *testing.T) {
			controller, store, cookie, id := overviewControllerFixture(t)
			writer, value := overviewRequest(t, controller, cookie)
			if writer.Code != 200 || string(value["call_evidence"]) == "null" {
				t.Fatal("successful runtime control unavailable", writer.Code, writer.Body.String())
			}
			if phase == "closed_database" {
				if err := store.DB.Close(); err != nil {
					t.Fatal(err)
				}
			} else if _, err := store.DB.Exec("DROP TABLE meta"); err != nil {
				t.Fatal(err)
			}
			writer, value = overviewRequest(t, controller, cookie)
			var failure struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(value["error"], &failure); err != nil {
				t.Fatal(err)
			}
			controller.State.RLock()
			current := controller.State.SuccessID()
			controller.State.RUnlock()
			if writer.Code != 503 || failure.Code != "storage_unavailable" || current != id {
				t.Fatal("unassessed storage failure cleared successful runtime evidence", writer.Code, failure.Code, current)
			}
		})
	}
}

func TestOverviewClearsSuccessfullyAssessedStaleRuntimeID(t *testing.T) {
	controller, store, cookie, _ := overviewControllerFixture(t)
	if _, err := store.DB.Exec("UPDATE upstream SET revision=revision+1"); err != nil {
		t.Fatal(err)
	}
	writer, value := overviewRequest(t, controller, cookie)
	controller.State.RLock()
	current := controller.State.SuccessID()
	controller.State.RUnlock()
	if writer.Code != 200 || string(value["call_evidence"]) != "null" || current != "" {
		t.Fatal("assessed stale runtime evidence was retained", writer.Code, current, writer.Body.String())
	}
}

type overviewCommitFailureDriver struct{}
type overviewCommitFailureConn struct{ driver.Conn }
type overviewCommitFailureTx struct{ driver.Tx }

func (overviewCommitFailureDriver) Open(name string) (driver.Conn, error) {
	connection, err := (&sqlite3.SQLiteDriver{}).Open(name)
	if err != nil {
		return nil, err
	}
	return overviewCommitFailureConn{connection}, nil
}

func (connection overviewCommitFailureConn) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	transaction, err := connection.Conn.(driver.ConnBeginTx).BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return overviewCommitFailureTx{transaction}, nil
}

func (transaction overviewCommitFailureTx) Commit() error {
	// Queries and transaction state are real SQLite; only the final commit
	// boundary fails so this test can observe the already determined decision.
	if err := transaction.Tx.Rollback(); err != nil {
		return err
	}
	return errors.New("controlled overview commit failure")
}

var registerOverviewCommitFailure sync.Once

func overviewWithFailingCommit(t *testing.T, controller *Controller, store *installation.Store) {
	t.Helper()
	var sequence int
	var name, filename string
	if err := store.DB.QueryRow("PRAGMA database_list").Scan(&sequence, &name, &filename); err != nil {
		t.Fatal(err)
	}
	if err := store.DB.Close(); err != nil {
		t.Fatal(err)
	}
	const driverName = "overview-commit-failure-sqlite3"
	registerOverviewCommitFailure.Do(func() { sql.Register(driverName, overviewCommitFailureDriver{}) })
	database, err := sql.Open(driverName, filename+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	store.DB = database
	controller.Records = records.New(database, store.Schema)
	controller.Reports = reporting.New(database, store.Now)
}

func TestOverviewCommitFailurePreservesAssessedInvalidationDecision(t *testing.T) {
	for _, stale := range []bool{false, true} {
		t.Run(map[bool]string{false: "valid_success", true: "known_stale_success"}[stale], func(t *testing.T) {
			controller, store, cookie, id := overviewControllerFixture(t)
			if stale {
				if _, err := store.DB.Exec("UPDATE upstream SET revision=revision+1"); err != nil {
					t.Fatal(err)
				}
			}
			overviewWithFailingCommit(t, controller, store)
			value, clearSuccess, err := controller.Reports.LegacyOverview(context.Background(), id, false)
			if err == nil || value != nil || clearSuccess != stale {
				t.Fatal("commit failure lost assessed decision", value, clearSuccess, err)
			}
			writer, _ := overviewRequest(t, controller, cookie)
			controller.State.RLock()
			current := controller.State.SuccessID()
			controller.State.RUnlock()
			if writer.Code != 503 || stale && current != "" || !stale && current != id {
				t.Fatal("commit failure changed assessed invalidation", writer.Code, current)
			}
		})
	}
}
