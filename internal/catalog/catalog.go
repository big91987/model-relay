// Package catalog owns endpoint configuration, model routes, grants and
// connection-check lifetimes. HTTP adapters own decoding and pagination cursors.
package catalog

import (
	"database/sql"

	"model-relay/internal/configstate"
	"model-relay/internal/fault"
	"model-relay/internal/routing"
	"model-relay/internal/upstream"
)

type Upstream struct {
	ID                   int64        `json:"id"`
	Name                 string       `json:"name"`
	URL                  string       `json:"url"`
	Enabled              bool         `json:"enabled"`
	Revision             int64        `json:"revision"`
	ExecutionRevision    int64        `json:"execution_revision"`
	CredentialConfigured bool         `json:"credential_configured"`
	Check                *CheckResult `json:"check"`
	Encrypted            []byte       `json:"-"`
	SourceKind           string       `json:"source_kind,omitempty"`
}

type CheckResult struct {
	ID                int64   `json:"id"`
	ExecutionRevision int64   `json:"execution_revision"`
	State             string  `json:"state"`
	Started           string  `json:"started_at"`
	Finished          *string `json:"finished_at"`
	Duration          *int64  `json:"duration_ms"`
	Code              string  `json:"error_code"`
	Status            *int    `json:"upstream_status"`
	Applicable        bool    `json:"applicable"`
}

type Candidate struct {
	UpstreamID        int64   `json:"upstream_id"`
	RemoteModel       string  `json:"remote_model"`
	Enabled           bool    `json:"enabled"`
	Priority          int     `json:"priority"`
	Weight            int     `json:"weight"`
	ExecutionRevision int64   `json:"execution_revision,omitempty"`
	GenerationHealth  string  `json:"generation_health,omitempty"`
	CooldownUntil     *string `json:"cooldown_until,omitempty"`
}

type RouteModel struct {
	Strategy   string      `json:"routing_strategy"`
	Alias      string      `json:"alias"`
	Enabled    bool        `json:"enabled"`
	Revision   int64       `json:"revision"`
	Legacy     *string     `json:"legacy_remote_model"`
	Candidates []Candidate `json:"candidates"`
}

type UpstreamInput struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Token      string `json:"token"`
	Enabled    *bool  `json:"enabled"`
	Expected   *int64 `json:"expected_revision"`
	SourceKind string `json:"source_kind"`
}

type CandidateInput struct {
	UpstreamID  int64  `json:"upstream_id"`
	RemoteModel string `json:"remote_model"`
	Enabled     *bool  `json:"enabled"`
	Priority    *int   `json:"priority"`
	Weight      *int   `json:"weight"`
}

type ModelInput struct {
	Alias      string            `json:"alias"`
	Strategy   *string           `json:"routing_strategy"`
	Enabled    *bool             `json:"enabled"`
	Expected   *int64            `json:"expected_revision"`
	Candidates *[]CandidateInput `json:"candidates"`
}

// RevisionConflict preserves the revision needed for an explicit reload.
// Unwrap retains the ordinary operation error classification for adapters.
type RevisionConflict struct{ CurrentRevision int64 }

func (conflict *RevisionConflict) Error() string { return "configuration_conflict" }
func (conflict *RevisionConflict) Unwrap() error {
	return fault.New(fault.Conflict, "configuration_conflict")
}

// CheckPersistenceError identifies a check that ran but whose final evidence
// could not be persisted. The adapter must mark storage degraded, preserving the
// existing gateway safety response without giving this module gateway state.
type CheckPersistenceError struct{}

func (*CheckPersistenceError) Error() string { return "storage_unavailable" }
func (*CheckPersistenceError) Unwrap() error {
	return fault.New(fault.Unavailable, "storage_unavailable")
}

type Snapshot struct {
	Model          RouteModel
	Upstreams      map[int64]Upstream
	ConfigRevision int64
	Epoch          uint64
}

type Service struct {
	db     *sql.DB
	key    []byte
	schema int
	state  *configstate.Coordinator
	router *routing.Engine
	client *upstream.Client
	policy upstream.Policy
}

// New shares the same coordination owner and routing engine as gateway admission.
// Callers must not hold that coordination lock when invoking ordinary methods.
func New(db *sql.DB, key []byte, schema int, state *configstate.Coordinator, router *routing.Engine, client *upstream.Client, policy upstream.Policy) *Service {
	return &Service{db: db, key: append([]byte(nil), key...), schema: schema, state: state, router: router, client: client, policy: policy}
}
func unavailable() error { return fault.New(fault.Unavailable, "storage_unavailable") }
func readError(err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return fault.New(fault.NotFound, "not_found")
	}
	return unavailable()
}
func configChanged(transaction *sql.Tx) error {
	_, err := transaction.Exec("UPDATE meta SET config_revision=config_revision+1")
	return err
}
func (service *Service) committed() { service.state.Changed(service.router) }
