package relay

import (
	"database/sql"
	"errors"
	"fmt"
)

type materialQuery interface {
	Query(string, ...any) (*sql.Rows, error)
}

func upstreamAAD(id int64, address string) string {
	return fmt.Sprintf("upstream:v2:%d:%s", id, address)
}

func validateMaterial(query materialQuery, key []byte, version int) error {
	rows, err := query.Query("SELECT id,url,token FROM upstream")
	if err != nil {
		return errors.New("invalid upstream storage")
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var address string
		var encrypted []byte
		if rows.Scan(&id, &address, &encrypted) != nil {
			return errors.New("invalid upstream material")
		}
		associated := upstreamAAD(id, address)
		if version == 1 {
			associated = "upstream:" + address
		}
		if len(encrypted) > 0 {
			if _, err := unseal(key, encrypted, associated); err != nil {
				return errors.New("invalid upstream material")
			}
		}
	}
	return rows.Err()
}

type UpgradeResult struct {
	Result     string `json:"result"`
	FromSchema int    `json:"from_schema"`
	Schema     int    `json:"schema"`
}

func integrity(transaction *sql.Tx) error {
	var result string
	if err := transaction.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil || result != "ok" {
		return errors.New("database integrity check failed")
	}
	rows, err := transaction.Query("PRAGMA foreign_key_check")
	if err != nil {
		return errors.New("database foreign key check failed")
	}
	defer rows.Close()
	if rows.Next() || rows.Err() != nil {
		return errors.New("database foreign key check failed")
	}
	return nil
}

func validateStorage(db *sql.DB, version int) error {
	transaction, err := db.Begin()
	if err != nil {
		return errors.New("storage validation failed")
	}
	defer transaction.Rollback()
	var count int
	if transaction.QueryRow("SELECT count(*) FROM meta").Scan(&count) != nil || count != 1 {
		return errors.New("installation metadata invalid")
	}
	if transaction.QueryRow("SELECT count(*) FROM admin").Scan(&count) != nil || count != 1 {
		return errors.New("administrator metadata invalid")
	}
	queries := []string{"SELECT id,name,digest,hint,enabled,created FROM app_keys LIMIT 0", "SELECT " + recordColumns + " FROM requests LIMIT 0"}
	if version >= 2 {
		queries = append(queries, "SELECT config_revision FROM meta LIMIT 0", "SELECT name,execution_revision FROM upstream LIMIT 0", "SELECT alias,enabled,revision,legacy_remote_model FROM models LIMIT 0", "SELECT alias,upstream_id,remote_model,enabled,priority,weight,execution_revision FROM route_candidates LIMIT 0", "SELECT evidence_version,config_revision,model_revision,committed_at,successful_upstream_id,attempts_complete FROM requests LIMIT 0", "SELECT request_id,ordinal,upstream_id,upstream_name,remote_model,upstream_revision,upstream_execution_revision,candidate_execution_revision,started_at,finished_at,duration_ms,state,error_code,upstream_status,prompt_tokens,completion_tokens,total_tokens FROM request_attempts LIMIT 0", "SELECT request_id,selection_round,upstream_id,snapshot_enabled,snapshot_priority,snapshot_weight,reason,observed_at,cooldown_until FROM request_route_decisions LIMIT 0", "SELECT upstream_id,execution_revision,state,started_at,finished_at,duration_ms,error_code,upstream_status FROM upstream_checks LIMIT 0")
	}
	for _, query := range queries {
		rows, err := transaction.Query(query)
		if err != nil {
			return errors.New("storage schema contract invalid")
		}
		rows.Close()
	}
	if version == 3 {
		if err := validateConsole(transaction); err != nil {
			return err
		}
	}
	return integrity(transaction)
}

func Upgrade(directory, keyFile, output string) (UpgradeResult, error) {
	return upgradeTo(directory, keyFile, output, 2)
}

func UpgradeConsole(directory, keyFile, output string) (UpgradeResult, error) {
	return upgradeTo(directory, keyFile, output, 3)
}

func upgradeTo(directory, keyFile, output string, target int) (UpgradeResult, error) {
	if output == "" {
		return UpgradeResult{}, errors.New("--output required")
	}
	store, err := openInstallation(directory, keyFile, true)
	if err != nil {
		return UpgradeResult{}, errors.New("upgrade installation/key/lock validation failed")
	}
	defer store.Close()
	if store.Schema >= target {
		transaction, err := store.DB.Begin()
		if err != nil {
			return UpgradeResult{}, errors.New("upgrade validation failed")
		}
		defer transaction.Rollback()
		if err = integrity(transaction); err != nil {
			return UpgradeResult{}, err
		}
		return UpgradeResult{Result: "already_current", FromSchema: store.Schema, Schema: store.Schema}, nil
	}
	if err = backupLocked(directory, output); err != nil {
		return UpgradeResult{}, errors.New("upgrade backup failed; no migration performed")
	}
	transaction, err := store.DB.Begin()
	if err != nil {
		return UpgradeResult{}, errors.New("upgrade transaction failed")
	}
	defer transaction.Rollback()
	if store.Schema == 2 && target == 3 {
		if err := createConsole(transaction); err != nil {
			return UpgradeResult{}, errors.New("console migration failed; source retained")
		}
		if _, err := transaction.Exec("UPDATE meta SET version=3"); err != nil {
			return UpgradeResult{}, err
		}
		if err := validateConsole(transaction); err != nil {
			return UpgradeResult{}, err
		}
		if err := integrity(transaction); err != nil {
			return UpgradeResult{}, err
		}
		if err := transaction.Commit(); err != nil {
			return UpgradeResult{}, err
		}
		return UpgradeResult{Result: "migrated", FromSchema: 2, Schema: 3}, nil
	}
	var address string
	var encrypted []byte
	err = transaction.QueryRow("SELECT url,token FROM upstream WHERE id=1").Scan(&address, &encrypted)
	hasUpstream := err == nil
	if err != nil && err != sql.ErrNoRows {
		return UpgradeResult{}, errors.New("upgrade source invalid")
	}
	if hasUpstream && len(encrypted) > 0 {
		plain, err := unseal(store.Key, encrypted, "upstream:"+address)
		if err != nil {
			return UpgradeResult{}, errors.New("invalid upstream material")
		}
		encrypted, err = seal(store.Key, plain, upstreamAAD(1, address))
		if err != nil {
			return UpgradeResult{}, errors.New("upgrade encryption failed")
		}
	}
	if _, err = transaction.Exec("ALTER TABLE upstream RENAME TO legacy_upstream; ALTER TABLE models RENAME TO legacy_models;" + configurationSchema); err != nil {
		return UpgradeResult{}, errors.New("upgrade configuration failed")
	}
	if hasUpstream {
		if _, err = transaction.Exec("INSERT INTO upstream SELECT id,'原有上游',url,?,enabled,revision,1 FROM legacy_upstream", encrypted); err != nil {
			return UpgradeResult{}, errors.New("upgrade upstream failed")
		}
	}
	if _, err = transaction.Exec("INSERT INTO models SELECT alias,enabled,1,CASE WHEN ? THEN NULL ELSE model END FROM legacy_models", hasUpstream); err != nil {
		return UpgradeResult{}, errors.New("upgrade models failed")
	}
	if hasUpstream {
		if _, err = transaction.Exec("INSERT INTO route_candidates SELECT alias,1,model,1,0,1,1 FROM legacy_models"); err != nil {
			return UpgradeResult{}, errors.New("upgrade candidates failed")
		}
	}
	var beforeModels, afterModels, beforeUpstreams, afterUpstreams int
	if transaction.QueryRow("SELECT count(*) FROM legacy_models").Scan(&beforeModels) != nil || transaction.QueryRow("SELECT count(*) FROM models").Scan(&afterModels) != nil || transaction.QueryRow("SELECT count(*) FROM legacy_upstream").Scan(&beforeUpstreams) != nil || transaction.QueryRow("SELECT count(*) FROM upstream").Scan(&afterUpstreams) != nil || beforeModels != afterModels || beforeUpstreams != afterUpstreams {
		return UpgradeResult{}, errors.New("upgrade object conservation failed")
	}
	if _, err = transaction.Exec(requestUpgrade + evidenceSchema + "ALTER TABLE meta ADD COLUMN config_revision INTEGER NOT NULL DEFAULT 1; DROP TABLE legacy_models; DROP TABLE legacy_upstream; UPDATE meta SET version=2;"); err != nil {
		return UpgradeResult{}, errors.New("upgrade evidence failed")
	}
	if err = validateMaterial(transaction, store.Key, 2); err != nil {
		return UpgradeResult{}, err
	}
	if target == 3 {
		if err := createConsole(transaction); err != nil {
			return UpgradeResult{}, err
		}
		if _, err := transaction.Exec("UPDATE meta SET version=3"); err != nil {
			return UpgradeResult{}, err
		}
		if err := validateConsole(transaction); err != nil {
			return UpgradeResult{}, err
		}
	}
	if err = integrity(transaction); err != nil {
		return UpgradeResult{}, err
	}
	if err = transaction.Commit(); err != nil {
		return UpgradeResult{}, errors.New("upgrade commit failed")
	}
	return UpgradeResult{Result: "migrated", FromSchema: 1, Schema: target}, nil
}
