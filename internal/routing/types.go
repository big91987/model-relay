// Package routing owns selection policy and its in-memory state.
// Callers provide configuration facts; HTTP and SQL belong to their adapters.
package routing

import "time"

type Candidate struct {
	UpstreamID        int64
	RemoteModel       string
	Enabled           bool
	Priority          int
	Weight            int
	ExecutionRevision int64
	UpstreamExecution int64
	Unavailable       string
}

type Snapshot struct {
	ObservedAt     time.Time
	Alias          string
	Strategy       string
	Revision       int64
	ConfigRevision int64
	Epoch          uint64
	Candidates     []Candidate
}

type Selection struct {
	Candidate    *Candidate
	Decisions    []RouteDecision
	Observations []SelectionObservation
}

type Measurement struct {
	Measured time.Time
	Value    float64
}

type Sample struct {
	EvaluatedAt  time.Time
	Snapshot     Snapshot
	Candidate    Candidate
	Ordinal      int
	Sequence     int64
	State        string
	Finished     *string
	CurrentEpoch uint64
	Degraded     bool
	Stream       bool
	Measurement  Measurement
}

type Observation struct {
	Cooldown   time.Duration
	ObservedAt time.Time
	Sequence   int64
	State      string
	Temporary  bool
}

// CacheStats describes resident state without exposing mutable maps.
type CacheStats struct {
	Samples         int
	HealthEntries   int
	CooldownEntries int
}

type Health struct {
	Sequence           int64
	InvalidatedThrough int64
	Cooldown           time.Time
	State              string
	Observed           time.Time
	CandidateRevision  int64
}

type RouteDecision struct {
	Round      int     `json:"selection_round"`
	UpstreamID int64   `json:"upstream_id"`
	Enabled    bool    `json:"snapshot_enabled"`
	Priority   int     `json:"snapshot_priority"`
	Weight     int     `json:"snapshot_weight"`
	Reason     string  `json:"reason"`
	Observed   string  `json:"observed_at"`
	Cooldown   *string `json:"cooldown_until"`
}

type SelectionObservation struct {
	Round       int      `json:"selection_round"`
	UpstreamID  int64    `json:"upstream_id"`
	Strategy    string   `json:"strategy"`
	Metric      string   `json:"metric_name"`
	Algorithm   string   `json:"algorithm"`
	Generation  uint64   `json:"generation"`
	Reason      string   `json:"selection_reason"`
	State       string   `json:"sample_state"`
	Count       int      `json:"sample_count"`
	EWMA        *float64 `json:"ewma_ms"`
	Measured    *string  `json:"measured_at"`
	Age         *float64 `json:"age_ms"`
	TTL         int      `json:"ttl_seconds"`
	Exploration uint64   `json:"exploration_sequence"`
	Bound       int      `json:"exploration_bound"`
}

type LatencyResult struct {
	Ordinal    int      `json:"ordinal"`
	UpstreamID int64    `json:"upstream_id"`
	Metric     string   `json:"metric_name"`
	Value      *float64 `json:"latency_ms"`
	Measured   *string  `json:"measured_at"`
	Finished   string   `json:"finished_at"`
	AcceptedAt string   `json:"evaluated_at"`
	Acceptance string   `json:"sample_acceptance"`
	Generation uint64   `json:"generation"`
}
