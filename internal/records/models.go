package records

import (
	"model-relay/internal/routing"
)

const Columns = "id,time,coalesce(key_id,''),coalesce(alias,''),coalesce(model,''),stream,coalesce(revision,0),state,error_code,upstream_status,duration_ms,prompt_tokens,completion_tokens,total_tokens"

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
	ConfigRevision     int64                          `json:"config_revision"`
	ModelRevision      int64                          `json:"model_revision"`
	SuccessfulUpstream *int64                         `json:"successful_upstream_id"`
	CommittedAt        *string                        `json:"committed_at"`
	AttemptsComplete   bool                           `json:"attempts_complete"`
	EvidenceVersion    int                            `json:"evidence_version"`
	Attempts           []Attempt                      `json:"attempts"`
	Decisions          []routing.RouteDecision        `json:"selection_decisions"`
	UsageSummary       map[string]UsagePart           `json:"usage_summary"`
	RoutingEvidence    []routing.SelectionObservation `json:"routing_decisions"`
	LatencyEvidence    []routing.LatencyResult        `json:"latency_observations"`
}

type Attempt struct {
	Ordinal            int     `json:"ordinal"`
	UpstreamID         int64   `json:"upstream_id"`
	UpstreamName       string  `json:"upstream_name"`
	RemoteModel        string  `json:"remote_model"`
	UpstreamRevision   int64   `json:"upstream_revision"`
	UpstreamExecution  int64   `json:"upstream_execution_revision"`
	CandidateExecution int64   `json:"candidate_execution_revision"`
	Started            string  `json:"started_at"`
	Finished           *string `json:"finished_at"`
	Duration           *int64  `json:"duration_ms"`
	State              string  `json:"state"`
	Code               string  `json:"error_code"`
	Status             *int    `json:"upstream_status"`
	Usage
	Sequence int64 `json:"-"`
}

type UsagePart struct {
	KnownSum        int64  `json:"known_sum"`
	UnknownAttempts int    `json:"unknown_attempts"`
	Completeness    string `json:"completeness"`
	Total           *int64 `json:"total"`
}

type rowScanner interface {
	Scan(...any) error
}

func Scan(row rowScanner) (Record, error) {
	var value Record
	err := row.Scan(&value.ID, &value.Time, &value.KeyID, &value.Alias, &value.Model, &value.Stream, &value.Revision, &value.State, &value.Code, &value.Status, &value.Duration, &value.Prompt, &value.Completion, &value.Total)
	return value, err
}
