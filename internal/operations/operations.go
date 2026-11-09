package operations

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"model-relay/internal/names"
	"model-relay/internal/secure"
	"time"
)

func Begin(transaction *sql.Tx, actor string, key, path string, action string, input any) (string, string, []byte, error) {
	operation := key
	if !names.Valid(operation) || len(operation) > 128 {
		return "", "", nil, errors.New("idempotency_key_required")
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		return "", "", nil, errors.New("invalid_request")
	}
	fingerprint := secure.Digest(action + ":" + path + ":" + string(encoded))
	var saved []byte
	var objectID, savedAction string
	err = transaction.QueryRow("SELECT fingerprint,object_id,action FROM console_operations WHERE actor=? AND operation=?", actor, operation).Scan(&saved, &objectID, &savedAction)
	if err == nil {
		if !bytes.Equal(saved, fingerprint) || savedAction != action {
			return "", "", nil, errors.New("idempotency_conflict")
		}
		return operation, objectID, fingerprint, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", nil, errors.New("storage_unavailable")
	}
	return operation, "", fingerprint, nil
}

func Complete(transaction *sql.Tx, actor, operation string, fingerprint []byte, action, objectID, tenant string) error {
	if _, err := transaction.Exec("INSERT INTO console_operations VALUES(?,?,?,?,?,?)", actor, operation, fingerprint, objectID, action, tenant); err != nil {
		return err
	}
	return Audit(transaction, actor, tenant, action, objectID)
}

// Audit timestamps use observed wall UTC, separate from effective eligibility time.
func Audit(transaction *sql.Tx, actor, tenant, action, objectID string) error {
	_, err := transaction.Exec("INSERT INTO console_audit VALUES(?,?,?,?,?,?)", secure.Random(16), time.Now().UTC().Format(time.RFC3339Nano), actor, tenant, action, objectID)
	return err
}

// Result distinguishes a newly created fact from an existing idempotent result.
type Result struct {
	Value   any
	Created bool
}
