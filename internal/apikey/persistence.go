package apikey

import (
	"database/sql"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/sqlread"
)

func scan(transaction *sql.Tx, keyID string) (ControlledKey, error) {
	var value ControlledKey
	err := transaction.QueryRow("SELECT k.id,k.name,k.hint,k.enabled,k.created,c.tenant_id,c.application,c.expires,c.revoked,c.revision FROM app_keys k JOIN key_controls c ON c.key_id=k.id WHERE k.id=?", keyID).Scan(&value.ID, &value.Name, &value.Hint, &value.Enabled, &value.Created, &value.Tenant, &value.Application, &value.Expires, &value.Revoked, &value.Revision)
	if err != nil {
		return value, err
	}
	value.Models, err = sqlread.Strings(transaction, "SELECT alias FROM key_model_grants WHERE key_id=? ORDER BY alias", keyID)
	if err != nil {
		return value, err
	}
	value.Diagnostics, err = sqlread.Strings(transaction, "SELECT user_id FROM key_diagnostic_members WHERE key_id=? ORDER BY user_id", keyID)
	if err == nil {
		value.Generations, err = sqlread.Maps(transaction, "SELECT id,hint,created,expires,retired FROM key_generations WHERE key_id=? ORDER BY created,id", keyID)
	}
	return value, err
}

// Scope checks key permissions within an existing transaction, without committing it.
func Scope(transaction *sql.Tx, principal access.Principal, keyID string, write bool) (ControlledKey, string, error) {
	value, err := scan(transaction, keyID)
	if err != nil {
		return value, "", fault.New(fault.NotFound, "not_found")
	}
	role, err := access.Role(transaction, principal, value.Tenant)
	if err != nil {
		return value, "", fault.New(fault.NotFound, "not_found")
	}
	if write && role != "administrator" {
		return value, role, fault.New(fault.Forbidden, "forbidden")
	}
	if role == "developer" {
		var assigned int
		if transaction.QueryRow("SELECT count(*) FROM key_diagnostic_members WHERE key_id=? AND tenant_id=? AND user_id=?", keyID, value.Tenant, principal.ID).Scan(&assigned) != nil || assigned != 1 {
			return value, role, fault.New(fault.NotFound, "not_found")
		}
	}
	return value, role, nil
}

func setKeyGrants(transaction *sql.Tx, keyID, tenantID string, input Input) error {
	if _, err := transaction.Exec("DELETE FROM key_model_grants WHERE key_id=?; DELETE FROM key_diagnostic_members WHERE key_id=?", keyID, keyID); err != nil {
		return err
	}
	for _, alias := range input.Models {
		if _, err := transaction.Exec("INSERT INTO key_model_grants VALUES(?,?)", keyID, alias); err != nil {
			return err
		}
	}
	for _, userID := range input.Diagnostics {
		if _, err := transaction.Exec("INSERT INTO key_diagnostic_members VALUES(?,?,?)", keyID, tenantID, userID); err != nil {
			return err
		}
	}
	return nil
}
