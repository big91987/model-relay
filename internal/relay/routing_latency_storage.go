package relay

import (
	"database/sql"
	"encoding/json"
	"errors"
)

const routingExtensionSchema = `CREATE TABLE routing_extension_meta(version INTEGER NOT NULL CHECK(version=1));
INSERT INTO routing_extension_meta VALUES(1);
CREATE TABLE model_route_policies(alias TEXT PRIMARY KEY REFERENCES models(alias) ON DELETE CASCADE,strategy TEXT NOT NULL CHECK(strategy IN('weighted','low_latency')));
CREATE TABLE route_selection_evidence(request_id TEXT NOT NULL,selection_round INTEGER NOT NULL,upstream_id INTEGER NOT NULL,evidence TEXT NOT NULL,PRIMARY KEY(request_id,selection_round,upstream_id),FOREIGN KEY(request_id,selection_round,upstream_id) REFERENCES request_route_decisions(request_id,selection_round,upstream_id));
CREATE TABLE route_latency_evidence(request_id TEXT NOT NULL,ordinal INTEGER NOT NULL,evidence TEXT NOT NULL,PRIMARY KEY(request_id,ordinal),FOREIGN KEY(request_id,ordinal) REFERENCES request_attempts(request_id,ordinal));`

func routingExtension(transaction *sql.Tx, create bool) error {
	var count int
	if err := transaction.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN('routing_extension_meta','model_route_policies','route_selection_evidence','route_latency_evidence')").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if !create {
			return nil
		}
		if _, err := transaction.Exec(routingExtensionSchema); err != nil {
			return err
		}
	} else if count != 4 {
		return errors.New("routing extension incomplete")
	}
	var version, rows int
	if transaction.QueryRow("SELECT count(*),coalesce(max(version),0) FROM routing_extension_meta").Scan(&rows, &version) != nil || rows != 1 || version != 1 {
		return errors.New("routing extension unsupported")
	}
	for _, query := range []string{"SELECT alias,strategy FROM model_route_policies LIMIT 0", "SELECT request_id,selection_round,upstream_id,evidence FROM route_selection_evidence LIMIT 0", "SELECT request_id,ordinal,evidence FROM route_latency_evidence LIMIT 0"} {
		result, err := transaction.Query(query)
		if err != nil {
			return errors.New("routing extension invalid")
		}
		result.Close()
	}
	if err := transaction.QueryRow("SELECT count(*) FROM model_route_policies WHERE strategy NOT IN('weighted','low_latency')").Scan(&count); err != nil || count != 0 {
		return errors.New("routing policy invalid")
	}
	return nil
}

func (store *Store) ensureRouting() error {
	if store.Schema < 2 {
		return nil
	}
	transaction, err := store.DB.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if err = routingExtension(transaction, true); err != nil {
		return err
	}
	return transaction.Commit()
}

func readRoutingEvidence(query materialQuery, id string) ([]SelectionObservation, []LatencyResult, error) {
	decisions := []SelectionObservation{}
	attempts := []LatencyResult{}
	rows, err := query.Query("SELECT evidence FROM route_selection_evidence WHERE request_id=? ORDER BY selection_round,upstream_id", id)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var raw string
		var value SelectionObservation
		if err = rows.Scan(&raw); err != nil {
			break
		}
		if err = json.Unmarshal([]byte(raw), &value); err != nil {
			break
		}
		decisions = append(decisions, value)
	}
	rowErr := rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, err
	}
	if rowErr != nil {
		return nil, nil, rowErr
	}
	rows, err = query.Query("SELECT evidence FROM route_latency_evidence WHERE request_id=? ORDER BY ordinal", id)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var raw string
		var value LatencyResult
		if err = rows.Scan(&raw); err != nil {
			break
		}
		if err = json.Unmarshal([]byte(raw), &value); err != nil {
			break
		}
		attempts = append(attempts, value)
	}
	rowErr = rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, err
	}
	return decisions, attempts, rowErr
}
