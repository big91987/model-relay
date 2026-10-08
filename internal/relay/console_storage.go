package relay

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const consoleSchema = `CREATE TABLE service_profiles(upstream_id INTEGER PRIMARY KEY REFERENCES upstream(id),source_kind TEXT NOT NULL CHECK(source_kind IN('provider','self_hosted','unspecified')));
INSERT INTO service_profiles SELECT id,'unspecified' FROM upstream;
CREATE TABLE key_controls(key_id TEXT PRIMARY KEY REFERENCES app_keys(id),tenant_id TEXT NOT NULL REFERENCES tenants(id),application TEXT NOT NULL,expires TEXT,revoked INTEGER NOT NULL DEFAULT 0,revision INTEGER NOT NULL DEFAULT 1,legacy INTEGER NOT NULL DEFAULT 0);
CREATE TABLE key_generations(id TEXT PRIMARY KEY,key_id TEXT NOT NULL REFERENCES app_keys(id),digest BLOB NOT NULL UNIQUE,hint TEXT NOT NULL,created TEXT NOT NULL,expires TEXT,retired INTEGER NOT NULL DEFAULT 0);
CREATE UNIQUE INDEX current_generation ON key_generations(key_id) WHERE expires IS NULL AND retired=0;
CREATE TABLE tenant_model_grants(tenant_id TEXT NOT NULL REFERENCES tenants(id),alias TEXT NOT NULL REFERENCES models(alias),PRIMARY KEY(tenant_id,alias));
CREATE TABLE key_model_grants(key_id TEXT NOT NULL REFERENCES app_keys(id),alias TEXT NOT NULL REFERENCES models(alias),PRIMARY KEY(key_id,alias));
CREATE TABLE key_diagnostic_members(key_id TEXT NOT NULL REFERENCES app_keys(id),tenant_id TEXT NOT NULL,user_id TEXT NOT NULL,PRIMARY KEY(key_id,user_id),FOREIGN KEY(tenant_id,user_id) REFERENCES memberships(tenant_id,user_id) ON DELETE CASCADE);
CREATE TRIGGER diagnostic_role_cleanup AFTER UPDATE OF role ON memberships WHEN NEW.role!='developer' BEGIN DELETE FROM key_diagnostic_members WHERE tenant_id=NEW.tenant_id AND user_id=NEW.user_id; END;
CREATE TABLE console_operations(actor TEXT NOT NULL,operation TEXT NOT NULL,fingerprint BLOB NOT NULL,object_id TEXT NOT NULL,action TEXT NOT NULL,tenant_id TEXT NOT NULL,PRIMARY KEY(actor,operation));
CREATE TABLE generation_operations(key_id TEXT NOT NULL REFERENCES app_keys(id),operation TEXT NOT NULL,fingerprint BLOB NOT NULL,request_id TEXT NOT NULL REFERENCES requests(id),PRIMARY KEY(key_id,operation));
CREATE TABLE attempt_dispatches(request_id TEXT NOT NULL,ordinal INTEGER NOT NULL,possible INTEGER NOT NULL DEFAULT 0,PRIMARY KEY(request_id,ordinal),FOREIGN KEY(request_id,ordinal) REFERENCES request_attempts(request_id,ordinal));
INSERT INTO attempt_dispatches SELECT request_id,ordinal,1 FROM request_attempts;
CREATE TRIGGER register_dispatch AFTER INSERT ON request_attempts BEGIN INSERT INTO attempt_dispatches VALUES(NEW.request_id,NEW.ordinal,0); END;
CREATE TABLE resource_limits(owner_type TEXT NOT NULL CHECK(owner_type IN('tenant','key')),owner_id TEXT NOT NULL,calls INTEGER,rpm INTEGER,tpm INTEGER,concurrency INTEGER,budget INTEGER,hard INTEGER NOT NULL DEFAULT 1,revision INTEGER NOT NULL DEFAULT 1,PRIMARY KEY(owner_type,owner_id),CHECK(calls>=0),CHECK(rpm>=0),CHECK(tpm>=0),CHECK(concurrency>=0),CHECK(budget>=0));
CREATE TABLE platform_limits(tenant_id TEXT PRIMARY KEY REFERENCES tenants(id),calls INTEGER,rpm INTEGER,tpm INTEGER,concurrency INTEGER,budget INTEGER,hard INTEGER NOT NULL,revision INTEGER NOT NULL);
CREATE TABLE model_bounds(alias TEXT NOT NULL REFERENCES models(alias),upstream_id INTEGER NOT NULL REFERENCES upstream(id),max_input INTEGER NOT NULL CHECK(max_input>=0),max_output INTEGER NOT NULL CHECK(max_output>=0),evidence TEXT NOT NULL,upstream_execution INTEGER NOT NULL,candidate_execution INTEGER NOT NULL,PRIMARY KEY(alias,upstream_id));
CREATE TABLE prices(id TEXT PRIMARY KEY,alias TEXT NOT NULL REFERENCES models(alias),upstream_id INTEGER NOT NULL REFERENCES upstream(id),effective TEXT NOT NULL,cost_input INTEGER NOT NULL CHECK(cost_input>=0),cost_output INTEGER NOT NULL CHECK(cost_output>=0),tenant_input INTEGER NOT NULL CHECK(tenant_input>=0),tenant_output INTEGER NOT NULL CHECK(tenant_output>=0),upstream_execution INTEGER NOT NULL,candidate_execution INTEGER NOT NULL);
CREATE INDEX prices_active ON prices(alias,upstream_id,effective,id);
CREATE TABLE admissions(request_id TEXT PRIMARY KEY REFERENCES requests(id),tenant_id TEXT NOT NULL REFERENCES tenants(id),key_id TEXT NOT NULL REFERENCES app_keys(id),period TEXT NOT NULL,minute TEXT NOT NULL,accepted TEXT NOT NULL,running INTEGER NOT NULL DEFAULT 1,tokens_reserved INTEGER NOT NULL,fee_reserved INTEGER NOT NULL,pending INTEGER NOT NULL DEFAULT 0,revision INTEGER NOT NULL DEFAULT 1);
CREATE TABLE request_prices(request_id TEXT NOT NULL REFERENCES admissions(request_id),upstream_id INTEGER NOT NULL,price_id TEXT NOT NULL,cost_input INTEGER NOT NULL,cost_output INTEGER NOT NULL,tenant_input INTEGER NOT NULL,tenant_output INTEGER NOT NULL,max_input INTEGER NOT NULL,max_output INTEGER NOT NULL,PRIMARY KEY(request_id,upstream_id));
CREATE TABLE charge_entries(id TEXT PRIMARY KEY,request_id TEXT NOT NULL REFERENCES admissions(request_id),ordinal INTEGER NOT NULL,revision INTEGER NOT NULL,period TEXT NOT NULL,input_tokens INTEGER NOT NULL,output_tokens INTEGER NOT NULL,cost INTEGER NOT NULL,fee INTEGER NOT NULL,reason TEXT NOT NULL,evidence TEXT NOT NULL,UNIQUE(request_id,ordinal,revision));
CREATE TABLE usage_reconciliations(request_id TEXT NOT NULL,ordinal INTEGER NOT NULL,revision INTEGER NOT NULL,input_tokens INTEGER NOT NULL CHECK(input_tokens>=0),output_tokens INTEGER NOT NULL CHECK(output_tokens>=0),reason TEXT NOT NULL,evidence TEXT NOT NULL,PRIMARY KEY(request_id,ordinal,revision));
CREATE TABLE historical_price_evidence(request_id TEXT NOT NULL REFERENCES admissions(request_id),upstream_id INTEGER NOT NULL,price_reference TEXT NOT NULL,cost_input INTEGER NOT NULL,cost_output INTEGER NOT NULL,tenant_input INTEGER NOT NULL,tenant_output INTEGER NOT NULL,reason TEXT NOT NULL,evidence TEXT NOT NULL,PRIMARY KEY(request_id,upstream_id));
CREATE TABLE bills(id TEXT PRIMARY KEY,tenant_id TEXT NOT NULL REFERENCES tenants(id),period TEXT NOT NULL,original_cost INTEGER NOT NULL,original_fee INTEGER NOT NULL,closed TEXT NOT NULL,UNIQUE(tenant_id,period));
CREATE TABLE bill_corrections(id TEXT PRIMARY KEY,bill_id TEXT NOT NULL REFERENCES bills(id),period TEXT NOT NULL,amount INTEGER NOT NULL,reason TEXT NOT NULL,actor TEXT NOT NULL,created TEXT NOT NULL);
CREATE TABLE legacy_pending(request_id TEXT PRIMARY KEY REFERENCES requests(id),tenant_id TEXT NOT NULL REFERENCES tenants(id),period TEXT NOT NULL,resolved INTEGER NOT NULL DEFAULT 0);
INSERT INTO key_controls SELECT id,'legacy',name,NULL,0,1,1 FROM app_keys;
INSERT INTO key_generations SELECT id,id,digest,hint,created,NULL,0 FROM app_keys;
INSERT INTO tenant_model_grants SELECT 'legacy',alias FROM models;
INSERT INTO key_model_grants SELECT k.id,m.alias FROM app_keys k CROSS JOIN models m;
INSERT INTO resource_limits SELECT 'tenant',id,NULL,NULL,NULL,NULL,NULL,1,1 FROM tenants;
INSERT INTO resource_limits SELECT 'key',id,NULL,NULL,NULL,NULL,NULL,1,1 FROM app_keys;`

const legacyPendingMigration = `INSERT INTO legacy_pending SELECT r.id,'legacy',substr(r.time,1,7),0 FROM requests r WHERE r.key_id IN(SELECT id FROM app_keys) AND (EXISTS(SELECT 1 FROM request_attempts a WHERE a.request_id=r.id) OR r.state IN('success','upstream_failed','timeout','stream_interrupted','canceled','process_interrupted','started'));`

func createConsole(transaction *sql.Tx) error {
	if err := createIdentity(transaction); err != nil {
		return err
	}
	if _, err := transaction.Exec(consoleSchema + legacyPendingMigration); err != nil {
		return err
	}
	return initializeConsoleClock(transaction)
}

func validateConsole(transaction *sql.Tx) error {
	queries := []string{
		"SELECT id,account,salt,hash,platform,active FROM users LIMIT 0",
		"SELECT tenant_id,user_id,role FROM memberships LIMIT 0",
		"SELECT id,tenant_id,digest,expires,state,accepted_by FROM invitations LIMIT 0",
		"SELECT key_id,tenant_id,application,expires,revoked,revision,legacy FROM key_controls LIMIT 0",
		"SELECT id,key_id,digest,expires,retired FROM key_generations LIMIT 0",
		"SELECT owner_type,owner_id,calls,rpm,tpm,concurrency,budget,hard,revision FROM resource_limits LIMIT 0",
		"SELECT request_id,tenant_id,key_id,period,minute,running,tokens_reserved,fee_reserved,pending,revision FROM admissions LIMIT 0",
		"SELECT request_id,ordinal,revision,input_tokens,output_tokens,cost,fee,reason,evidence FROM charge_entries LIMIT 0",
		"SELECT id,tenant_id,period,original_cost,original_fee,closed FROM bills LIMIT 0",
		"SELECT upstream_id,source_kind FROM service_profiles LIMIT 0",
		"SELECT id,name,active,created FROM tenants LIMIT 0",
		"SELECT tenant_id,alias FROM tenant_model_grants LIMIT 0",
		"SELECT key_id,alias FROM key_model_grants LIMIT 0",
		"SELECT key_id,tenant_id,user_id FROM key_diagnostic_members LIMIT 0",
		"SELECT actor,operation,fingerprint,object_id,action,tenant_id FROM console_operations LIMIT 0",
		"SELECT key_id,operation,fingerprint,request_id FROM generation_operations LIMIT 0",
		"SELECT tenant_id,calls,rpm,tpm,concurrency,budget,hard,revision FROM platform_limits LIMIT 0",
		"SELECT request_id,ordinal,possible FROM attempt_dispatches LIMIT 0",
		"SELECT alias,upstream_id,max_input,max_output,evidence,upstream_execution,candidate_execution FROM model_bounds LIMIT 0",
		"SELECT id,alias,upstream_id,effective,cost_input,cost_output,tenant_input,tenant_output,upstream_execution,candidate_execution FROM prices LIMIT 0",
		"SELECT request_id,upstream_id,price_id,cost_input,cost_output,tenant_input,tenant_output,max_input,max_output FROM request_prices LIMIT 0",
		"SELECT request_id,ordinal,revision,input_tokens,output_tokens,reason,evidence FROM usage_reconciliations LIMIT 0",
		"SELECT request_id,upstream_id,price_reference,cost_input,cost_output,tenant_input,tenant_output,reason,evidence FROM historical_price_evidence LIMIT 0",
		"SELECT id,bill_id,period,amount,reason,actor,created FROM bill_corrections LIMIT 0",
		"SELECT request_id,tenant_id,period,resolved FROM legacy_pending LIMIT 0",
	}
	for _, query := range queries {
		rows, err := transaction.Query(query)
		if err != nil {
			return errors.New("console schema contract invalid")
		}
		rows.Close()
	}
	return nil
}

func (store *Store) recoverConsole() error {
	transaction, err := store.DB.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if err := initializeConsoleClock(transaction); err != nil {
		return err
	}
	rows, err := transaction.Query("SELECT request_id,tenant_id FROM admissions WHERE running=1")
	if err != nil {
		return err
	}
	type interrupted struct{ requestID, tenantID string }
	values := []interrupted{}
	for rows.Next() {
		var value interrupted
		if rows.Scan(&value.requestID, &value.tenantID) != nil {
			rows.Close()
			return errors.New("recovery scan failed")
		}
		values = append(values, value)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, value := range values {
		if err := settleAdmission(transaction, value.requestID); err != nil {
			return fmt.Errorf("recovery settlement: %w", err)
		}
		if err := auditIdentity(transaction, "system", value.tenantID, "request.recover", value.requestID); err != nil {
			return err
		}
	}
	var text string
	if err := transaction.QueryRow("SELECT effective FROM console_clock WHERE id=1").Scan(&text); err != nil {
		return err
	}
	observed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return errors.New("invalid effective clock")
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	store.effectiveTime = observed
	return nil
}
