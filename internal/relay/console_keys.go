package relay

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type ControlledKey struct {
	AppKey
	Tenant      string           `json:"tenant_id"`
	Application string           `json:"application"`
	Expires     *string          `json:"expires_at"`
	Revoked     bool             `json:"revoked"`
	Revision    int64            `json:"revision"`
	Models      []string         `json:"models"`
	Diagnostics []string         `json:"diagnostic_members"`
	Generations []map[string]any `json:"generations"`
}

type KeyInput struct {
	Name        string   `json:"name"`
	Application string   `json:"application"`
	Expires     *string  `json:"expires_at"`
	Enabled     bool     `json:"enabled"`
	Models      []string `json:"models"`
	Diagnostics []string `json:"diagnostic_members"`
	Revision    int64    `json:"expected_revision"`
}

func operationStart(transaction *sql.Tx, actor string, request *http.Request, action string, input any) (string, string, []byte, error) {
	operation := request.Header.Get("Idempotency-Key")
	if !validName(operation) || len(operation) > 128 {
		return "", "", nil, errors.New("idempotency_key_required")
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		return "", "", nil, errors.New("invalid_request")
	}
	fingerprint := digest(action + ":" + request.URL.Path + ":" + string(encoded))
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

func operationFinish(transaction *sql.Tx, actor, operation string, fingerprint []byte, action, objectID, tenant string) error {
	if _, err := transaction.Exec("INSERT INTO console_operations VALUES(?,?,?,?,?,?)", actor, operation, fingerprint, objectID, action, tenant); err != nil {
		return err
	}
	return auditIdentity(transaction, actor, tenant, action, objectID)
}

func scanControlledKey(transaction *sql.Tx, keyID string) (ControlledKey, error) {
	var value ControlledKey
	err := transaction.QueryRow("SELECT k.id,k.name,k.hint,k.enabled,k.created,c.tenant_id,c.application,c.expires,c.revoked,c.revision FROM app_keys k JOIN key_controls c ON c.key_id=k.id WHERE k.id=?", keyID).Scan(&value.ID, &value.Name, &value.Hint, &value.Enabled, &value.Created, &value.Tenant, &value.Application, &value.Expires, &value.Revoked, &value.Revision)
	if err != nil {
		return value, err
	}
	value.Models, err = stringList(transaction, "SELECT alias FROM key_model_grants WHERE key_id=? ORDER BY alias", keyID)
	if err != nil {
		return value, err
	}
	value.Diagnostics, err = stringList(transaction, "SELECT user_id FROM key_diagnostic_members WHERE key_id=? ORDER BY user_id", keyID)
	if err == nil {
		value.Generations, err = queryMaps(transaction, "SELECT id,hint,created,expires,retired FROM key_generations WHERE key_id=? ORDER BY created,id", keyID)
	}
	return value, err
}

func stringList(transaction *sql.Tx, query string, arguments ...any) ([]string, error) {
	rows, err := transaction.Query(query, arguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []string{}
	for rows.Next() {
		var value string
		if rows.Scan(&value) != nil {
			return nil, errors.New("storage scan failed")
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func keyScope(transaction *sql.Tx, principal Principal, keyID string, write bool) (ControlledKey, string, int) {
	value, err := scanControlledKey(transaction, keyID)
	if err != nil {
		return value, "", 404
	}
	role, err := scopedRole(transaction, principal, value.Tenant)
	if err != nil {
		return value, "", 404
	}
	if write && role != "administrator" {
		return value, role, 403
	}
	if role == "developer" {
		var assigned int
		if transaction.QueryRow("SELECT count(*) FROM key_diagnostic_members WHERE key_id=? AND tenant_id=? AND user_id=?", keyID, value.Tenant, principal.ID).Scan(&assigned) != nil || assigned != 1 {
			return value, role, 404
		}
	}
	return value, role, 200
}

func validateKeyInput(transaction *sql.Tx, tenantID string, input KeyInput, now time.Time) bool {
	if !validName(input.Name) || !validName(input.Application) || len(input.Models) > 1000 || len(input.Diagnostics) > 1000 {
		return false
	}
	if input.Expires != nil {
		expires, err := time.Parse(time.RFC3339Nano, *input.Expires)
		if err != nil || input.Enabled && !now.Before(expires) {
			return false
		}
	}
	seen := map[string]bool{}
	for _, alias := range input.Models {
		var count int
		if seen[alias] || transaction.QueryRow("SELECT count(*) FROM tenant_model_grants WHERE tenant_id=? AND alias=?", tenantID, alias).Scan(&count) != nil || count != 1 {
			return false
		}
		seen[alias] = true
	}
	seen = map[string]bool{}
	for _, userID := range input.Diagnostics {
		var count int
		if seen[userID] || transaction.QueryRow("SELECT count(*) FROM memberships m JOIN users u ON u.id=m.user_id WHERE tenant_id=? AND user_id=? AND role='developer' AND u.active=1", tenantID, userID).Scan(&count) != nil || count != 1 {
			return false
		}
		seen[userID] = true
	}
	return true
}

func setKeyGrants(transaction *sql.Tx, keyID, tenantID string, input KeyInput) error {
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

func (server *Server) controlledKeys(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, tenantID string, parts []string) {
	role, err := scopedRole(transaction, principal, tenantID)
	if err != nil {
		failure(writer, 404, "not_found", "")
		return
	}
	if len(parts) == 2 && parts[1] == "limits" {
		key, _, status := keyScope(transaction, principal, parts[0], request.Method != http.MethodGet)
		if status != 200 || key.Tenant != tenantID {
			failure(writer, 404, "not_found", "")
			return
		}
		server.manageLimits(writer, request, transaction, principal, tenantID, role, "key", parts[0])
		return
	}
	if len(parts) == 0 && request.Method == http.MethodGet {
		query := "SELECT key_id FROM key_controls WHERE tenant_id=? ORDER BY key_id"
		arguments := []any{tenantID}
		if role == "developer" {
			query = "SELECT c.key_id FROM key_controls c JOIN key_diagnostic_members d ON d.key_id=c.key_id WHERE c.tenant_id=? AND d.user_id=? ORDER BY c.key_id"
			arguments = append(arguments, principal.ID)
		}
		ids, err := stringList(transaction, query, arguments...)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		values := []ControlledKey{}
		for _, id := range ids {
			value, err := scanControlledKey(transaction, id)
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			values = append(values, value)
		}
		replyCollection(writer, request, values)
		return
	}
	if role != "administrator" && request.Method != http.MethodGet {
		failure(writer, 403, "forbidden", "")
		return
	}
	if len(parts) == 0 && request.Method == http.MethodPost {
		var input KeyInput
		if jsonInput(writer, request, &input) != nil {
			failure(writer, 400, "invalid_key", "")
			return
		}
		operation, prior, fingerprint, err := operationStart(transaction, principal.ID, request, "key.create", input)
		if err != nil {
			failure(writer, 409, err.Error(), "")
			return
		}
		if prior != "" {
			reply(writer, 200, map[string]any{"id": prior, "already_applied": true, "secret_recoverable": false})
			return
		}
		if !validateKeyInput(transaction, tenantID, input, server.Store.now()) {
			failure(writer, 400, "invalid_key", "")
			return
		}
		var active bool
		if transaction.QueryRow("SELECT active FROM tenants WHERE id=?", tenantID).Scan(&active) != nil || !active {
			failure(writer, 409, "tenant_paused", "")
			return
		}
		keyID, secret := randomSecret(16), "mr_"+randomSecret(32)
		created := server.Store.now().Format(time.RFC3339Nano)
		if _, err := transaction.Exec("INSERT INTO app_keys VALUES(?,?,?,?,?,?); INSERT INTO key_controls VALUES(?,?,?,?,0,1,0); INSERT INTO key_generations VALUES(?,?,?,?,?,NULL,0); INSERT INTO resource_limits VALUES('key',?,NULL,NULL,NULL,NULL,NULL,1,1)", keyID, input.Name, digest(secret), secret[:11], input.Enabled, created, keyID, tenantID, input.Application, input.Expires, randomSecret(16), keyID, digest(secret), secret[:11], created, keyID); err != nil || setKeyGrants(transaction, keyID, tenantID, input) != nil || operationFinish(transaction, principal.ID, operation, fingerprint, "key.create", keyID, tenantID) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 201, map[string]any{"id": keyID, "key": secret, "secret_recoverable": false})
		return
	}
	if len(parts) == 0 {
		failure(writer, 405, "method_not_allowed", "")
		return
	}
	keyID := parts[0]
	key, _, status := keyScope(transaction, principal, keyID, request.Method != http.MethodGet)
	if status != 200 || key.Tenant != tenantID {
		failure(writer, 404, "not_found", "")
		return
	}
	if len(parts) == 1 && request.Method == http.MethodGet {
		reply(writer, 200, key)
		return
	}
	if len(parts) == 1 && request.Method == http.MethodPatch {
		var input KeyInput
		if jsonInput(writer, request, &input) != nil {
			failure(writer, 400, "invalid_key", "")
			return
		}
		operation, prior, fingerprint, err := operationStart(transaction, principal.ID, request, "key.update", input)
		if err != nil {
			failure(writer, 409, err.Error(), "")
			return
		}
		if prior != "" {
			reply(writer, 200, map[string]any{"id": prior, "already_applied": true})
			return
		}
		if !validateKeyInput(transaction, tenantID, input, server.Store.now()) {
			failure(writer, 400, "invalid_key", "")
			return
		}
		if input.Revision != key.Revision || key.Revoked {
			failure(writer, 409, "revision_conflict_or_revoked", "")
			return
		}
		var active bool
		if transaction.QueryRow("SELECT active FROM tenants WHERE id=?", tenantID).Scan(&active) != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if !active && (input.Enabled && !key.Enabled || !subsetStrings(input.Models, key.Models) || expiryExpands(input.Expires, key.Expires)) {
			failure(writer, 409, "tenant_paused", "")
			return
		}
		if _, err := transaction.Exec("UPDATE app_keys SET name=?,enabled=? WHERE id=?; UPDATE key_controls SET application=?,expires=?,revision=revision+1,legacy=0 WHERE key_id=?", input.Name, input.Enabled, keyID, input.Application, input.Expires, keyID); err != nil || setKeyGrants(transaction, keyID, tenantID, input) != nil || operationFinish(transaction, principal.ID, operation, fingerprint, "key.update", keyID, tenantID) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	if len(parts) != 2 || request.Method != http.MethodPost {
		failure(writer, 404, "not_found", "")
		return
	}
	action := parts[1]
	var input struct {
		Hours    int   `json:"transition_hours"`
		Revision int64 `json:"expected_revision"`
	}
	if jsonInput(writer, request, &input) != nil || input.Hours < 0 || input.Hours > 24 {
		failure(writer, 400, "invalid_request", "")
		return
	}
	operation, prior, fingerprint, err := operationStart(transaction, principal.ID, request, "key."+action, input)
	if err != nil {
		failure(writer, 409, err.Error(), "")
		return
	}
	if prior != "" {
		reply(writer, 200, map[string]any{"id": prior, "already_applied": true, "secret_recoverable": false})
		return
	}
	if input.Revision != key.Revision || key.Revoked {
		failure(writer, 409, "revision_conflict_or_revoked", "")
		return
	}
	secret := ""
	switch action {
	case "revoke":
		_, err = transaction.Exec("UPDATE key_controls SET revoked=1,revision=revision+1 WHERE key_id=?; UPDATE app_keys SET enabled=0 WHERE id=?; UPDATE key_generations SET retired=1 WHERE key_id=?", keyID, keyID, keyID)
	case "retire-previous":
		_, err = transaction.Exec("UPDATE key_generations SET retired=1 WHERE key_id=? AND expires IS NOT NULL; UPDATE key_controls SET revision=revision+1 WHERE key_id=?", keyID, keyID)
	case "rotations":
		var active bool
		if transaction.QueryRow("SELECT active FROM tenants WHERE id=?", tenantID).Scan(&active) != nil || !active {
			failure(writer, 409, "tenant_paused", "")
			return
		}
		now := server.Store.now()
		previous, err := queryMaps(transaction, "SELECT id,expires FROM key_generations WHERE key_id=? AND retired=0 AND expires IS NOT NULL", keyID)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		for _, generation := range previous {
			expiry, err := time.Parse(time.RFC3339Nano, generation["expires"].(string))
			if err != nil || now.Before(expiry) {
				failure(writer, 409, "transition_already_active", "")
				return
			}
			if _, err := transaction.Exec("UPDATE key_generations SET retired=1 WHERE id=?", generation["id"]); err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
		}
		secret = "mr_" + randomSecret(32)
		_, err = transaction.Exec("UPDATE key_generations SET expires=?,retired=? WHERE key_id=? AND retired=0 AND expires IS NULL; INSERT INTO key_generations VALUES(?,?,?,?,?,NULL,0); UPDATE app_keys SET digest=?,hint=? WHERE id=?; UPDATE key_controls SET revision=revision+1 WHERE key_id=?", now.Add(time.Duration(input.Hours)*time.Hour).Format(time.RFC3339Nano), input.Hours == 0, keyID, randomSecret(16), keyID, digest(secret), secret[:11], now.Format(time.RFC3339Nano), digest(secret), secret[:11], keyID, keyID)
	default:
		failure(writer, 404, "not_found", "")
		return
	}
	if err != nil || operationFinish(transaction, principal.ID, operation, fingerprint, "key."+action, keyID, tenantID) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"id": keyID, "key": secret, "secret_recoverable": false})
}

func subsetStrings(values, existing []string) bool {
	set := map[string]bool{}
	for _, value := range existing {
		set[value] = true
	}
	for _, value := range values {
		if !set[value] {
			return false
		}
	}
	return true
}

func expiryExpands(next, previous *string) bool {
	if previous == nil {
		return false
	}
	if next == nil {
		return true
	}
	nextTime, nextErr := time.Parse(time.RFC3339Nano, *next)
	previousTime, previousErr := time.Parse(time.RFC3339Nano, *previous)
	return nextErr != nil || previousErr != nil || nextTime.After(previousTime)
}

func (server *Server) controlledAppID(request *http.Request) (string, bool) {
	credential := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	if !strings.HasPrefix(credential, "mr_") || len(credential) != 46 {
		return "", false
	}
	return generationIdentity(server.Store.DB, credential, server.Store.now())
}

type identityQuery interface{ QueryRow(string, ...any) *sql.Row }

func generationIdentity(query identityQuery, credential string, now time.Time) (string, bool) {
	var keyID string
	var generationExpiry, keyExpiry *string
	err := query.QueryRow("SELECT k.id,g.expires,c.expires FROM key_generations g JOIN app_keys k ON k.id=g.key_id JOIN key_controls c ON c.key_id=k.id WHERE g.digest=? AND g.retired=0 AND k.enabled=1 AND c.revoked=0", digest(credential)).Scan(&keyID, &generationExpiry, &keyExpiry)
	if err != nil {
		return "", false
	}
	for _, expiry := range []*string{generationExpiry, keyExpiry} {
		if expiry != nil {
			parsed, err := time.Parse(time.RFC3339Nano, *expiry)
			if err != nil || !now.Before(parsed) {
				return keyID, false
			}
		}
	}
	return keyID, true
}

func (store *Store) authorizeModel(keyID, alias string) (bool, string) {
	var active bool
	var allowed int
	err := store.DB.QueryRow("SELECT t.active,(SELECT count(*) FROM key_model_grants k JOIN tenant_model_grants g ON g.alias=k.alias AND g.tenant_id=c.tenant_id WHERE k.key_id=c.key_id AND k.alias=?) FROM key_controls c JOIN tenants t ON t.id=c.tenant_id WHERE c.key_id=?", alias, keyID).Scan(&active, &allowed)
	if err != nil {
		return false, "storage_unavailable"
	}
	if !active {
		return false, "tenant_paused"
	}
	if allowed != 1 {
		return false, "model_forbidden"
	}
	return true, ""
}
