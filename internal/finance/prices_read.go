package finance

import (
	"database/sql"
	"time"

	"model-relay/internal/accounting"
)

// PriceAt selects a price for the supplied frozen execution versions. It does
// not consult the current route, which may have changed since admission began.
// Missing prices retain the upstream identity and return sql.ErrNoRows so the
// caller can apply its existing legacy-unpriced policy.
func PriceAt(transaction *sql.Tx, alias string, upstreamID, upstreamExecution, candidateExecution int64, at time.Time) (accounting.PriceSnapshot, string, error) {
	value := accounting.PriceSnapshot{Upstream: upstreamID}
	var effective string
	err := transaction.QueryRow("SELECT id,cost_input,cost_output,tenant_input,tenant_output,effective FROM prices WHERE alias=? AND upstream_id=? AND upstream_execution=? AND candidate_execution=? AND julianday(effective)<=julianday(?) ORDER BY julianday(effective) DESC,id DESC LIMIT 1", alias, upstreamID, upstreamExecution, candidateExecution, at.Format(time.RFC3339Nano)).Scan(&value.ID, &value.CostIn, &value.CostOut, &value.FeeIn, &value.FeeOut, &effective)
	return value, effective, err
}

// CurrentModelRate returns the latest price among the model's current route
// executions. Its cross-candidate ordering is distinct from frozen PriceAt.
func CurrentModelRate(transaction *sql.Tx, alias string, at time.Time) (*int64, *int64, *string, error) {
	var input, output *int64
	var effective *string
	err := transaction.QueryRow("SELECT p.tenant_input,p.tenant_output,p.effective FROM prices p JOIN upstream u ON u.id=p.upstream_id JOIN route_candidates c ON c.alias=p.alias AND c.upstream_id=p.upstream_id WHERE p.alias=? AND p.upstream_execution=u.execution_revision AND p.candidate_execution=c.execution_revision AND julianday(p.effective)<=julianday(?) ORDER BY julianday(p.effective) DESC,p.id DESC LIMIT 1", alias, at.Format(time.RFC3339Nano)).Scan(&input, &output, &effective)
	return input, output, effective, err
}
