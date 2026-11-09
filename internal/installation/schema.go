package installation

import "database/sql"

const configurationSchema = `CREATE TABLE upstream(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT UNIQUE NOT NULL, url TEXT NOT NULL, token BLOB NOT NULL, enabled INTEGER NOT NULL CHECK(enabled IN(0,1)), revision INTEGER NOT NULL, execution_revision INTEGER NOT NULL);
CREATE TABLE models(alias TEXT PRIMARY KEY, enabled INTEGER NOT NULL CHECK(enabled IN(0,1)), revision INTEGER NOT NULL, legacy_remote_model TEXT);
CREATE TABLE route_candidates(alias TEXT NOT NULL REFERENCES models(alias), upstream_id INTEGER NOT NULL REFERENCES upstream(id), remote_model TEXT NOT NULL, enabled INTEGER NOT NULL CHECK(enabled IN(0,1)), priority INTEGER NOT NULL CHECK(priority BETWEEN 0 AND 100), weight INTEGER NOT NULL CHECK(weight BETWEEN 1 AND 100), execution_revision INTEGER NOT NULL, PRIMARY KEY(alias,upstream_id));`

const evidenceSchema = `CREATE TABLE upstream_checks(id INTEGER PRIMARY KEY AUTOINCREMENT, upstream_id INTEGER NOT NULL REFERENCES upstream(id), execution_revision INTEGER NOT NULL, state TEXT NOT NULL, started_at TEXT NOT NULL, finished_at TEXT, duration_ms INTEGER, error_code TEXT NOT NULL, upstream_status INTEGER);
CREATE TABLE request_attempts(request_id TEXT NOT NULL REFERENCES requests(id), ordinal INTEGER NOT NULL CHECK(ordinal BETWEEN 1 AND 3), upstream_id INTEGER NOT NULL, upstream_name TEXT NOT NULL, remote_model TEXT NOT NULL, upstream_revision INTEGER NOT NULL, upstream_execution_revision INTEGER NOT NULL, candidate_execution_revision INTEGER NOT NULL, started_at TEXT NOT NULL, finished_at TEXT, duration_ms INTEGER, state TEXT NOT NULL, error_code TEXT NOT NULL, upstream_status INTEGER, prompt_tokens INTEGER, completion_tokens INTEGER, total_tokens INTEGER, PRIMARY KEY(request_id,ordinal));
CREATE TABLE request_route_decisions(request_id TEXT NOT NULL REFERENCES requests(id), selection_round INTEGER NOT NULL, upstream_id INTEGER NOT NULL, snapshot_enabled INTEGER NOT NULL, snapshot_priority INTEGER NOT NULL, snapshot_weight INTEGER NOT NULL, reason TEXT NOT NULL, observed_at TEXT NOT NULL, cooldown_until TEXT, PRIMARY KEY(request_id,selection_round,upstream_id));
CREATE INDEX attempts_upstream ON request_attempts(upstream_id,started_at);
CREATE INDEX checks_upstream ON upstream_checks(upstream_id,id);`

const requestUpgrade = `ALTER TABLE requests ADD COLUMN evidence_version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE requests ADD COLUMN config_revision INTEGER;
ALTER TABLE requests ADD COLUMN model_revision INTEGER;
ALTER TABLE requests ADD COLUMN committed_at TEXT;
ALTER TABLE requests ADD COLUMN successful_upstream_id INTEGER;
ALTER TABLE requests ADD COLUMN attempts_complete INTEGER NOT NULL DEFAULT 0;`

const schema = `CREATE TABLE meta(version INTEGER NOT NULL, sentinel BLOB NOT NULL, config_revision INTEGER NOT NULL DEFAULT 1);
CREATE TABLE admin(salt BLOB NOT NULL, hash BLOB NOT NULL);
CREATE TABLE app_keys(id TEXT PRIMARY KEY, name TEXT NOT NULL, digest BLOB UNIQUE NOT NULL, hint TEXT NOT NULL, enabled INTEGER NOT NULL, created TEXT NOT NULL);
CREATE TABLE requests(id TEXT PRIMARY KEY, time TEXT NOT NULL, key_id TEXT, alias TEXT, model TEXT, stream INTEGER NOT NULL, revision INTEGER, state TEXT NOT NULL, error_code TEXT NOT NULL, upstream_status INTEGER, duration_ms INTEGER, prompt_tokens INTEGER, completion_tokens INTEGER, total_tokens INTEGER);` + configurationSchema + requestUpgrade + evidenceSchema

const identitySchema = `CREATE TABLE users(id TEXT PRIMARY KEY,account TEXT NOT NULL UNIQUE,salt BLOB NOT NULL,hash BLOB NOT NULL,platform INTEGER NOT NULL CHECK(platform IN(0,1)),active INTEGER NOT NULL CHECK(active IN(0,1)));
CREATE TABLE tenants(id TEXT PRIMARY KEY,name TEXT NOT NULL,active INTEGER NOT NULL CHECK(active IN(0,1)),created TEXT NOT NULL);
CREATE TABLE memberships(tenant_id TEXT NOT NULL REFERENCES tenants(id),user_id TEXT NOT NULL REFERENCES users(id),role TEXT NOT NULL CHECK(role IN('administrator','developer','finance')),PRIMARY KEY(tenant_id,user_id));
CREATE TABLE invitations(id TEXT PRIMARY KEY,tenant_id TEXT NOT NULL REFERENCES tenants(id),account TEXT NOT NULL,role TEXT NOT NULL CHECK(role IN('administrator','developer','finance')),digest BLOB NOT NULL UNIQUE,expires TEXT NOT NULL,state TEXT NOT NULL CHECK(state IN('pending','accepted','revoked')),accepted_by TEXT REFERENCES users(id));
CREATE TABLE console_audit(id TEXT PRIMARY KEY,time TEXT NOT NULL,actor TEXT NOT NULL,tenant_id TEXT NOT NULL,action TEXT NOT NULL,object_id TEXT NOT NULL);
CREATE INDEX memberships_user ON memberships(user_id,tenant_id);
CREATE INDEX invitations_tenant ON invitations(tenant_id,id);
CREATE INDEX console_audit_tenant ON console_audit(tenant_id,time,id);`

func createIdentity(transaction *sql.Tx) error {
	if _, err := transaction.Exec(identitySchema); err != nil {
		return err
	}
	_, err := transaction.Exec("INSERT INTO users SELECT 'platform-admin','admin',salt,hash,1,1 FROM admin; INSERT INTO tenants VALUES('legacy','原有应用',1,strftime('%Y-%m-%dT%H:%M:%fZ','now')); INSERT INTO memberships VALUES('legacy','platform-admin','administrator')")
	return err
}
