package app

import (
	"database/sql"

	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/routing"
	"model-relay/internal/upstream"
	"net/http"
	"net/url"
	"path/filepath"
)

// catalogFor supplies a persistence-only fixture reader when a test has no
// running composition. Runtime/configuration mutations use the actual server.
func catalogFor(store *installation.Store) *catalog.Service {
	return catalog.New(store.DB, store.Key, store.Schema, configstate.New(), routing.New(), upstream.NewClient(&http.Client{}), upstream.Policy{})
}

func testDatabase(directory string) (*sql.DB, error) {
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

func testRoutingCandidate(candidate catalog.Candidate, source catalog.Upstream) routing.Candidate {
	return routing.Candidate{UpstreamID: candidate.UpstreamID, RemoteModel: candidate.RemoteModel, Enabled: candidate.Enabled, Priority: candidate.Priority, Weight: candidate.Weight, ExecutionRevision: candidate.ExecutionRevision, UpstreamExecution: source.ExecutionRevision}
}

func testRoutingSnapshot(snapshot catalog.Snapshot) routing.Snapshot {
	value := routing.Snapshot{Alias: snapshot.Model.Alias, Strategy: snapshot.Model.Strategy, Revision: snapshot.Model.Revision, ConfigRevision: snapshot.ConfigRevision, Epoch: snapshot.Epoch}
	for _, candidate := range snapshot.Model.Candidates {
		value.Candidates = append(value.Candidates, testRoutingCandidate(candidate, snapshot.Upstreams[candidate.UpstreamID]))
	}
	return value
}

func observeTestRouting(server *Server, snapshot catalog.Snapshot, candidate catalog.Candidate, attempt records.Attempt, temporary bool) {
	server.router.Observe(testRoutingSnapshot(snapshot), testRoutingCandidate(candidate, snapshot.Upstreams[candidate.UpstreamID]), routing.Observation{Cooldown: server.Config.Cooldown, Sequence: attempt.Sequence, State: attempt.State, Temporary: temporary, ObservedAt: server.Store.WallTime()})
}
