package relay

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

const identitySchema = `CREATE TABLE users(id TEXT PRIMARY KEY,account TEXT NOT NULL UNIQUE,salt BLOB NOT NULL,hash BLOB NOT NULL,platform INTEGER NOT NULL CHECK(platform IN(0,1)),active INTEGER NOT NULL CHECK(active IN(0,1)));
CREATE TABLE tenants(id TEXT PRIMARY KEY,name TEXT NOT NULL,active INTEGER NOT NULL CHECK(active IN(0,1)),created TEXT NOT NULL);
CREATE TABLE memberships(tenant_id TEXT NOT NULL REFERENCES tenants(id),user_id TEXT NOT NULL REFERENCES users(id),role TEXT NOT NULL CHECK(role IN('administrator','developer','finance')),PRIMARY KEY(tenant_id,user_id));
CREATE TABLE invitations(id TEXT PRIMARY KEY,tenant_id TEXT NOT NULL REFERENCES tenants(id),account TEXT NOT NULL,role TEXT NOT NULL CHECK(role IN('administrator','developer','finance')),digest BLOB NOT NULL UNIQUE,expires TEXT NOT NULL,state TEXT NOT NULL CHECK(state IN('pending','accepted','revoked')),accepted_by TEXT REFERENCES users(id));
CREATE TABLE console_audit(id TEXT PRIMARY KEY,time TEXT NOT NULL,actor TEXT NOT NULL,tenant_id TEXT NOT NULL,action TEXT NOT NULL,object_id TEXT NOT NULL);
CREATE INDEX memberships_user ON memberships(user_id,tenant_id);
CREATE INDEX invitations_tenant ON invitations(tenant_id,id);
CREATE INDEX console_audit_tenant ON console_audit(tenant_id,time,id);`

type Principal struct {
	ID       string `json:"id"`
	Account  string `json:"account"`
	Platform bool   `json:"platform"`
}

type Tenant struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Active  bool           `json:"active"`
	Created string         `json:"created"`
	Role    string         `json:"role,omitempty"`
	Summary map[string]any `json:"summary,omitempty"`
}

type Invitation struct {
	ID      string `json:"id"`
	Tenant  string `json:"tenant_id"`
	Account string `json:"account"`
	Role    string `json:"role"`
	Expires string `json:"expires_at"`
	State   string `json:"state"`
}

func validRole(role string) bool {
	return role == "administrator" || role == "developer" || role == "finance"
}

func invitationAccount(value string) (string, bool) {
	account := strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(account)
	return account, err == nil && parsed.Address == account && len(account) <= 254
}

func validPassword(value string) bool {
	return len(value) >= 12 && len(value) <= 1024 && !strings.ContainsAny(value, "\r\n\x00")
}

func passwordMaterial(password string) ([]byte, []byte) {
	salt := digest(randomSecret(16))[:16]
	return salt, argon2.IDKey([]byte(password), salt, 3, 64*1024, 1, 32)
}

func createIdentity(transaction *sql.Tx) error {
	if _, err := transaction.Exec(identitySchema); err != nil {
		return err
	}
	_, err := transaction.Exec("INSERT INTO users SELECT 'platform-admin','admin',salt,hash,1,1 FROM admin; INSERT INTO tenants VALUES('legacy','原有应用',1,strftime('%Y-%m-%dT%H:%M:%fZ','now')); INSERT INTO memberships VALUES('legacy','platform-admin','administrator')")
	return err
}

func (store *Store) authenticate(account, password string) (Principal, bool) {
	if account == "" {
		account = "admin"
	}
	if account != "admin" {
		account = strings.ToLower(strings.TrimSpace(account))
	}
	var principal Principal
	var salt, hash []byte
	err := store.DB.QueryRow("SELECT id,account,platform,salt,hash FROM users WHERE account=? AND active=1", account).Scan(&principal.ID, &principal.Account, &principal.Platform, &salt, &hash)
	if err != nil {
		salt = digest("model-relay-missing-account")[:16]
		hash = make([]byte, 32)
	}
	actual := argon2.IDKey([]byte(password), salt, 3, 64*1024, 1, 32)
	return principal, err == nil && subtle.ConstantTimeCompare(actual, hash) == 1
}

func (store *Store) principal(userID string) (Principal, error) {
	var principal Principal
	err := store.DB.QueryRow("SELECT id,account,platform FROM users WHERE id=? AND active=1", userID).Scan(&principal.ID, &principal.Account, &principal.Platform)
	return principal, err
}

func scopedRole(transaction *sql.Tx, principal Principal, tenantID string) (string, error) {
	var role string
	if principal.Platform {
		err := transaction.QueryRow("SELECT 'administrator' FROM tenants WHERE id=?", tenantID).Scan(&role)
		return role, err
	}
	err := transaction.QueryRow("SELECT m.role FROM memberships m JOIN users u ON u.id=m.user_id WHERE tenant_id=? AND user_id=? AND u.active=1", tenantID, principal.ID).Scan(&role)
	return role, err
}

func auditIdentity(transaction *sql.Tx, actor, tenant, action, objectID string) error {
	_, err := transaction.Exec("INSERT INTO console_audit VALUES(?,?,?,?,?,?)", randomSecret(16), time.Now().UTC().Format(time.RFC3339Nano), actor, tenant, action, objectID)
	return err
}

func (server *Server) identityAPI(writer http.ResponseWriter, request *http.Request, current session) bool {
	path := strings.TrimPrefix(request.URL.Path, "/admin/api/")
	parts := strings.Split(path, "/")
	identityPath := path == "identity" || path == "tenants" || (len(parts) == 2 && parts[0] == "tenants") || (len(parts) >= 3 && parts[0] == "tenants" && (parts[2] == "members" || parts[2] == "invitations"))
	if !identityPath {
		return false
	}
	principal, err := server.Store.principal(current.UserID)
	if err != nil {
		failure(writer, 401, "unauthorized", "")
		return true
	}
	if path == "identity" && request.Method == http.MethodGet {
		reply(writer, 200, principal)
		return true
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	transaction, err := server.Store.DB.BeginTx(request.Context(), nil)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return true
	}
	defer transaction.Rollback()
	if path == "tenants" && request.Method == http.MethodGet {
		query := "SELECT t.id,t.name,t.active,t.created,'administrator' FROM tenants t ORDER BY t.created,t.id"
		arguments := []any{}
		if !principal.Platform {
			query = "SELECT t.id,t.name,t.active,t.created,m.role FROM tenants t JOIN memberships m ON m.tenant_id=t.id WHERE m.user_id=? ORDER BY t.created,t.id"
			arguments = append(arguments, principal.ID)
		}
		rows, err := transaction.Query(query, arguments...)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return true
		}
		defer rows.Close()
		values := []Tenant{}
		for rows.Next() {
			var value Tenant
			if rows.Scan(&value.ID, &value.Name, &value.Active, &value.Created, &value.Role) != nil {
				failure(writer, 503, "storage_unavailable", "")
				return true
			}
			values = append(values, value)
		}
		if rows.Err() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return true
		}
		rows.Close()
		for index := range values {
			if !principal.Platform && values[index].Role != "administrator" {
				continue
			}
			values[index].Summary, err = tenantSummary(transaction, values[index].ID, server.Store.now())
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return true
			}
		}
		replyCollection(writer, request, values)
		return true
	}
	if path == "tenants" && request.Method == http.MethodPost {
		if !principal.Platform {
			failure(writer, 403, "forbidden", "")
			return true
		}
		var input struct {
			Name string `json:"name"`
		}
		if jsonInput(writer, request, &input) != nil || !validName(input.Name) {
			failure(writer, 400, "invalid_name", "")
			return true
		}
		value := Tenant{ID: randomSecret(16), Name: input.Name, Active: true, Created: time.Now().UTC().Format(time.RFC3339Nano)}
		if _, err := transaction.Exec("INSERT INTO tenants VALUES(?,?,1,?)", value.ID, value.Name, value.Created); err != nil || auditIdentity(transaction, principal.ID, value.ID, "tenant.create", value.ID) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return true
		}
		reply(writer, 201, value)
		return true
	}
	if len(parts) < 2 || parts[0] != "tenants" {
		failure(writer, 404, "not_found", "")
		return true
	}
	tenantID := parts[1]
	role, err := scopedRole(transaction, principal, tenantID)
	if err != nil {
		failure(writer, 404, "not_found", "")
		return true
	}
	if request.Method != http.MethodGet && role != "administrator" && !(len(parts) == 4 && parts[2] == "members" && parts[3] == principal.ID && request.Method == http.MethodDelete) {
		failure(writer, 403, "forbidden", "")
		return true
	}
	if len(parts) == 2 && request.Method == http.MethodPatch {
		if !principal.Platform {
			failure(writer, 403, "forbidden", "")
			return true
		}
		var input struct {
			Active bool `json:"active"`
		}
		if jsonInput(writer, request, &input) != nil {
			failure(writer, 400, "invalid_request", "")
			return true
		}
		if _, err := transaction.Exec("UPDATE tenants SET active=? WHERE id=?", input.Active, tenantID); err != nil || auditIdentity(transaction, principal.ID, tenantID, "tenant.status", tenantID) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return true
		}
		reply(writer, 200, map[string]bool{"ok": true})
		return true
	}
	if len(parts) >= 3 && parts[2] == "invitations" {
		server.tenantInvitations(writer, request, transaction, principal, tenantID, role, parts[3:])
		return true
	}
	if len(parts) >= 3 && parts[2] == "members" {
		server.tenantMembers(writer, request, transaction, principal, tenantID, role, parts[3:])
		return true
	}
	failure(writer, 404, "not_found", "")
	return true
}

func (server *Server) tenantInvitations(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, tenantID, role string, remaining []string) {
	if role != "administrator" {
		failure(writer, 403, "forbidden", "")
		return
	}
	if len(remaining) == 0 && request.Method == http.MethodGet {
		rows, err := transaction.Query("SELECT id,tenant_id,account,role,expires,state FROM invitations WHERE tenant_id=? ORDER BY expires,id", tenantID)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		defer rows.Close()
		values := []Invitation{}
		for rows.Next() {
			var value Invitation
			if rows.Scan(&value.ID, &value.Tenant, &value.Account, &value.Role, &value.Expires, &value.State) != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			if value.State == "pending" && value.Expires <= time.Now().UTC().Format(time.RFC3339Nano) {
				value.State = "expired"
			}
			values = append(values, value)
		}
		if rows.Err() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		replyCollection(writer, request, values)
		return
	}
	if len(remaining) == 0 && request.Method == http.MethodPost {
		var input struct {
			Account string `json:"account"`
			Role    string `json:"role"`
			Hours   int    `json:"hours"`
		}
		if jsonInput(writer, request, &input) != nil {
			failure(writer, 400, "invalid_request", "")
			return
		}
		account, valid := invitationAccount(input.Account)
		if input.Hours == 0 {
			input.Hours = 24
		}
		if !valid || !validRole(input.Role) || input.Hours < 1 || input.Hours > 168 {
			failure(writer, 400, "invalid_invitation", "")
			return
		}
		secret := randomSecret(32)
		invitation := Invitation{ID: randomSecret(16), Tenant: tenantID, Account: account, Role: input.Role, Expires: server.Store.now().Add(time.Duration(input.Hours) * time.Hour).Format(time.RFC3339Nano), State: "pending"}
		if _, err := transaction.Exec("INSERT INTO invitations VALUES(?,?,?,?,?,?,'pending',NULL)", invitation.ID, tenantID, account, input.Role, digest(secret), invitation.Expires); err != nil || auditIdentity(transaction, principal.ID, tenantID, "invitation.create", invitation.ID) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 201, map[string]any{"invitation": invitation, "acceptance_token": secret, "delivery": "manual_not_emailed"})
		return
	}
	if len(remaining) == 2 && remaining[1] == "revoke" && request.Method == http.MethodPost {
		var input struct{}
		if jsonInput(writer, request, &input) != nil {
			failure(writer, 400, "invalid_request", "")
			return
		}
		result, err := transaction.Exec("UPDATE invitations SET state='revoked' WHERE id=? AND tenant_id=? AND state='pending'", remaining[0], tenantID)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		changed, err := result.RowsAffected()
		if err != nil || changed != 1 {
			failure(writer, 409, "invitation_not_pending", "")
			return
		}
		if auditIdentity(transaction, principal.ID, tenantID, "invitation.revoke", remaining[0]) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	failure(writer, 404, "not_found", "")
}

func (server *Server) tenantMembers(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, tenantID, role string, remaining []string) {
	if len(remaining) == 0 && request.Method == http.MethodGet {
		if role != "administrator" {
			failure(writer, 403, "forbidden", "")
			return
		}
		rows, err := transaction.Query("SELECT u.id,u.account,m.role,u.active FROM memberships m JOIN users u ON u.id=m.user_id WHERE m.tenant_id=? ORDER BY u.account,u.id", tenantID)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		defer rows.Close()
		values := []map[string]any{}
		for rows.Next() {
			var userID, account, memberRole string
			var active bool
			if rows.Scan(&userID, &account, &memberRole, &active) != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			values = append(values, map[string]any{"id": userID, "account": account, "role": memberRole, "active": active})
		}
		if rows.Err() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		replyCollection(writer, request, values)
		return
	}
	if len(remaining) != 1 || (request.Method != http.MethodPatch && request.Method != http.MethodDelete) {
		failure(writer, 404, "not_found", "")
		return
	}
	userID := remaining[0]
	if role != "administrator" && !(userID == principal.ID && request.Method == http.MethodDelete) {
		failure(writer, 403, "forbidden", "")
		return
	}
	var previous string
	if transaction.QueryRow("SELECT role FROM memberships WHERE tenant_id=? AND user_id=?", tenantID, userID).Scan(&previous) != nil {
		failure(writer, 404, "not_found", "")
		return
	}
	var input struct {
		Role string `json:"role"`
	}
	if request.Method == http.MethodPatch && (jsonInput(writer, request, &input) != nil || !validRole(input.Role)) {
		failure(writer, 400, "invalid_role", "")
		return
	}
	if previous == "administrator" && input.Role != "administrator" {
		var administrators int
		if transaction.QueryRow("SELECT count(*) FROM memberships m JOIN users u ON u.id=m.user_id WHERE tenant_id=? AND role='administrator' AND u.active=1", tenantID).Scan(&administrators) != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if administrators <= 1 {
			failure(writer, 409, "last_administrator", "")
			return
		}
	}
	var err error
	if request.Method == http.MethodDelete {
		_, err = transaction.Exec("DELETE FROM memberships WHERE tenant_id=? AND user_id=?", tenantID, userID)
	} else {
		_, err = transaction.Exec("UPDATE memberships SET role=? WHERE tenant_id=? AND user_id=?", input.Role, tenantID, userID)
	}
	if err != nil || auditIdentity(transaction, principal.ID, tenantID, "member."+strings.ToLower(request.Method), userID) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]bool{"ok": true})
}

func (server *Server) acceptInvitation(writer http.ResponseWriter, request *http.Request) {
	server.authMu.Lock()
	now := server.Store.now()
	filtered := server.invitationAttempts[:0]
	for _, attempt := range server.invitationAttempts {
		if attempt.After(now.Add(-time.Minute)) {
			filtered = append(filtered, attempt)
		}
	}
	server.invitationAttempts = filtered
	if len(filtered) >= 10 {
		server.authMu.Unlock()
		writer.Header().Set("Retry-After", "60")
		failure(writer, 429, "invitation_rate_limited", "")
		return
	}
	server.invitationAttempts = append(server.invitationAttempts, now)
	server.authMu.Unlock()
	select {
	case server.passwordSlots <- struct{}{}:
		defer func() { <-server.passwordSlots }()
	default:
		failure(writer, 429, "login_rate_limited", "")
		return
	}
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if jsonInput(writer, request, &input) != nil || len(input.Token) != 43 || len(input.Password) > 1024 {
		failure(writer, 400, "invalid_invitation", "")
		return
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	transaction, err := server.Store.DB.BeginTx(request.Context(), nil)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	defer transaction.Rollback()
	var invitation Invitation
	var acceptedBy sql.NullString
	var active bool
	err = transaction.QueryRow("SELECT i.id,i.tenant_id,i.account,i.role,i.expires,i.state,i.accepted_by,t.active FROM invitations i JOIN tenants t ON t.id=i.tenant_id WHERE i.digest=?", digest(input.Token)).Scan(&invitation.ID, &invitation.Tenant, &invitation.Account, &invitation.Role, &invitation.Expires, &invitation.State, &acceptedBy, &active)
	expiry, parseErr := time.Parse(time.RFC3339Nano, invitation.Expires)
	if err == nil && invitation.State == "accepted" {
		reply(writer, 200, map[string]any{"accepted": true, "already_applied": true})
		return
	}
	if err != nil || parseErr != nil || !server.Store.now().Before(expiry) || invitation.State != "pending" || !active {
		failure(writer, 400, "invitation_unavailable", "")
		return
	}
	var userID string
	err = transaction.QueryRow("SELECT id FROM users WHERE account=?", invitation.Account).Scan(&userID)
	if err == nil {
		current, ok := server.getSession(request)
		if !ok || current.UserID != userID || subtle.ConstantTimeCompare([]byte(current.CSRF), []byte(request.Header.Get("X-CSRF-Token"))) != 1 {
			failure(writer, 401, "existing_account_login_required", "")
			return
		}
		var enabled bool
		if transaction.QueryRow("SELECT active FROM users WHERE id=?", userID).Scan(&enabled) != nil || !enabled {
			failure(writer, 401, "unauthorized", "")
			return
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		if !validPassword(input.Password) {
			failure(writer, 400, "password_requires_12_characters", "")
			return
		}
		userID = randomSecret(16)
		salt, hash := passwordMaterial(input.Password)
		if _, err = transaction.Exec("INSERT INTO users VALUES(?,?,?,?,0,1)", userID, invitation.Account, salt, hash); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
	} else {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	var membershipCount int
	if transaction.QueryRow("SELECT count(*) FROM memberships WHERE tenant_id=? AND user_id=?", invitation.Tenant, userID).Scan(&membershipCount) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if membershipCount != 0 {
		failure(writer, 409, "already_member", "")
		return
	}
	if _, err = transaction.Exec("INSERT INTO memberships VALUES(?,?,?); UPDATE invitations SET state='accepted',accepted_by=? WHERE id=? AND state='pending'", invitation.Tenant, userID, invitation.Role, userID, invitation.ID); err != nil || auditIdentity(transaction, userID, invitation.Tenant, "invitation.accept", invitation.ID) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"accepted": true, "account": invitation.Account, "tenant_id": invitation.Tenant, "email_verified": false})
}
