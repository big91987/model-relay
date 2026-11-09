package identity_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/identity"
	"model-relay/internal/secure"
)

type identityFixture struct {
	db                   *sql.DB
	service              *identity.Service
	ctx                  context.Context
	clock                *time.Time
	platform, admin, dev access.Principal
}

func installation(t *testing.T) identityFixture {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "identity.db")+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	now := time.Date(2030, 1, 2, 12, 0, 0, 0, time.UTC)
	f := identityFixture{db: db, service: identity.New(db, func() time.Time { return now }), ctx: context.Background(), clock: &now, platform: access.Principal{ID: "platform", Platform: true}, admin: access.Principal{ID: "admin"}, dev: access.Principal{ID: "dev"}}
	f.exec(t, identitySchema)
	f.exec(t, `INSERT INTO tenants VALUES('tenant','Tenant',1,'2026-01-01T00:00:00Z'),('other','Other',1,'2026-01-02T00:00:00Z');
INSERT INTO users VALUES('admin','admin@example.com',x'00',x'00',0,1),('dev','dev@example.com',x'00',x'00',0,1),('finance','finance@example.com',x'00',x'00',0,1),('inactive','inactive@example.com',x'00',x'00',0,0);
INSERT INTO memberships VALUES('tenant','admin','administrator'),('tenant','dev','developer'),('tenant','finance','finance'),('tenant','inactive','administrator');
INSERT INTO key_diagnostic_members VALUES('key','tenant','dev');`)
	return f
}

func (f identityFixture) exec(t *testing.T, statement string, args ...any) {
	t.Helper()
	if _, err := f.db.Exec(statement, args...); err != nil {
		t.Fatal(err)
	}
}
func (f identityFixture) count(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var result int64
	if err := f.db.QueryRow(query, args...).Scan(&result); err != nil {
		t.Fatal(err)
	}
	return result
}
func failure(t *testing.T, err error, kind fault.Kind, code string) {
	t.Helper()
	var value *fault.Error
	if !errors.As(err, &value) || value.Kind != kind || value.Code != code {
		t.Fatalf("want %v/%s, got %v", kind, code, err)
	}
}

func TestMemberChangesPreserveLastActiveAdministratorAndDiagnosticCleanup(t *testing.T) {
	f := installation(t)
	_, err := f.service.ChangeMemberRole(f.ctx, f.admin, "tenant", "admin", &identity.MemberInput{Role: "developer"})
	failure(t, err, fault.Conflict, "last_administrator")
	_, err = f.service.RemoveMember(f.ctx, f.admin, "tenant", "admin")
	failure(t, err, fault.Conflict, "last_administrator")
	if _, err = f.service.ChangeMemberRole(f.ctx, f.admin, "tenant", "dev", &identity.MemberInput{Role: "administrator"}); err != nil {
		t.Fatal(err)
	}
	if f.count(t, "SELECT count(*) FROM key_diagnostic_members") != 0 {
		t.Fatal("changed role retained diagnostic assignment")
	}
	if _, err = f.service.ChangeMemberRole(f.ctx, f.admin, "tenant", "admin", &identity.MemberInput{Role: "finance"}); err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.RemoveMember(f.ctx, f.admin, "tenant", "admin"); err != nil {
		t.Fatal("self leave", err)
	}
	if f.count(t, "SELECT count(*) FROM memberships WHERE user_id='admin'") != 0 {
		t.Fatal("self leave did not remove membership")
	}
	if f.count(t, "SELECT count(*) FROM console_audit WHERE action IN('member.patch','member.delete')") != 3 {
		t.Fatal("member operations audit count")
	}
}

func TestMemberAuthorizationEntityAndMalformedInputPrecedence(t *testing.T) {
	f := installation(t)
	_, err := f.service.ChangeMemberRole(f.ctx, f.dev, "tenant", "missing", nil)
	failure(t, err, fault.Forbidden, "forbidden")
	_, err = f.service.ChangeMemberRole(f.ctx, f.admin, "tenant", "missing", nil)
	failure(t, err, fault.NotFound, "not_found")
	_, err = f.service.ChangeMemberRole(f.ctx, f.admin, "tenant", "dev", nil)
	failure(t, err, fault.Invalid, "invalid_role")
	_, err = f.service.Members(f.ctx, f.dev, "tenant")
	failure(t, err, fault.Forbidden, "forbidden")
	_, err = f.service.Members(f.ctx, f.dev, "other")
	failure(t, err, fault.NotFound, "not_found")
	if _, err = f.service.RemoveMember(f.ctx, f.dev, "tenant", "dev"); err != nil {
		t.Fatal("developer self leave", err)
	}
	if f.count(t, "SELECT count(*) FROM key_diagnostic_members") != 0 {
		t.Fatal("self leave retained diagnostics")
	}
	members, err := f.service.Members(f.ctx, f.admin, "tenant")
	if err != nil || len(members) != 3 {
		t.Fatal(members, err)
	}
}

func TestMemberAuditFailureRollsBackRoleAndDiagnosticCleanup(t *testing.T) {
	f := installation(t)
	f.exec(t, "CREATE TRIGGER fail_audit BEFORE INSERT ON console_audit BEGIN SELECT RAISE(ABORT,'test audit failure'); END")
	_, err := f.service.ChangeMemberRole(f.ctx, f.admin, "tenant", "dev", &identity.MemberInput{Role: "finance"})
	failure(t, err, fault.Unavailable, "storage_unavailable")
	if f.count(t, "SELECT count(*) FROM memberships WHERE user_id='dev' AND role='developer'") != 1 || f.count(t, "SELECT count(*) FROM key_diagnostic_members") != 1 {
		t.Fatal("failed audit leaked role or trigger cleanup")
	}
}

func TestInvitationCreationNormalizesAccountAndRevokeDoesNotExposeSecret(t *testing.T) {
	f := installation(t)
	input := &identity.InvitationInput{Account: " New.Person@Example.Com ", Role: "developer"}
	issued, err := f.service.Invite(f.ctx, f.admin, "tenant", input)
	if err != nil || issued.Invitation.Account != "new.person@example.com" || issued.Invitation.Expires != "2030-01-03T12:00:00Z" || len(issued.AcceptanceToken) != 43 || input.Hours != 0 {
		t.Fatal(issued, err)
	}
	var saved []byte
	if err = f.db.QueryRow("SELECT digest FROM invitations WHERE id=?", issued.Invitation.ID).Scan(&saved); err != nil || string(saved) != string(secure.Digest(issued.AcceptanceToken)) {
		t.Fatal("invitation token digest", err)
	}
	listed, err := f.service.Invitations(f.ctx, f.admin, "tenant")
	if err != nil || len(listed) != 1 || listed[0].State != "pending" {
		t.Fatal(listed, err)
	}
	_, err = f.service.RevokeInvitation(f.ctx, f.dev, "tenant", issued.Invitation.ID, nil)
	failure(t, err, fault.Forbidden, "forbidden")
	_, err = f.service.RevokeInvitation(f.ctx, f.admin, "tenant", issued.Invitation.ID, nil)
	failure(t, err, fault.Invalid, "invalid_request")
	if _, err = f.service.RevokeInvitation(f.ctx, f.admin, "tenant", issued.Invitation.ID, &identity.RevokeInput{}); err != nil {
		t.Fatal(err)
	}
	_, err = f.service.RevokeInvitation(f.ctx, f.admin, "tenant", issued.Invitation.ID, &identity.RevokeInput{})
	failure(t, err, fault.Conflict, "invitation_not_pending")
	_, err = f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken, Password: "controlled-password"}, nil)
	failure(t, err, fault.Invalid, "invitation_unavailable")
}

func TestInvitationListingRetainsWallClockExpiryPresentation(t *testing.T) {
	f := installation(t)
	*f.clock = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	f.exec(t, "INSERT INTO invitations VALUES('expired','tenant','old@example.com','developer',?,'2001-01-01T00:00:00Z','pending',NULL)", secure.Digest("expired"))
	listed, err := f.service.Invitations(f.ctx, f.admin, "tenant")
	if err != nil || len(listed) != 1 || listed[0].State != "expired" {
		t.Fatal("wall-clock expiry presentation changed", listed, err)
	}
	var state string
	if err = f.db.QueryRow("SELECT state FROM invitations WHERE id='expired'").Scan(&state); err != nil || state != "pending" {
		t.Fatal("listing changed persisted invitation", state, err)
	}
}

func TestAcceptNewAccountAuthenticatesAndAcceptedReplaySkipsProofAndExpiry(t *testing.T) {
	f := installation(t)
	issued, err := f.service.Invite(f.ctx, f.admin, "tenant", &identity.InvitationInput{Account: "new@example.com", Role: "developer"})
	if err != nil {
		t.Fatal(err)
	}
	proofCalls := 0
	accepted, err := f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken, Password: "controlled-password"}, func(string) bool { proofCalls++; return false })
	if err != nil || !accepted.Accepted || accepted.Account != "new@example.com" || proofCalls != 0 || accepted.EmailVerified == nil || *accepted.EmailVerified {
		t.Fatal(accepted, err, proofCalls)
	}
	principal, authenticated := f.service.Authenticate(" New@Example.com ", "controlled-password")
	if !authenticated || principal.Account != "new@example.com" {
		t.Fatal("new account authentication", principal, authenticated)
	}
	if _, authenticated = f.service.Authenticate("new@example.com", "wrong-password"); authenticated {
		t.Fatal("incorrect password authenticated")
	}
	f.exec(t, "UPDATE tenants SET active=0; UPDATE invitations SET expires='invalid'; UPDATE users SET active=0 WHERE id=?", principal.ID)
	accepted, err = f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken}, func(string) bool { proofCalls++; return false })
	if err != nil || !accepted.AlreadyApplied || accepted.EmailVerified != nil || proofCalls != 0 {
		t.Fatal("accepted replay checked expiry, tenant pause or account proof", accepted, err, proofCalls)
	}
	if _, err = f.service.Principal(principal.ID); err == nil {
		t.Fatal("disabled account retained live principal")
	}
}

func TestExistingAccountProofIsLazyAndDisabledAccountCannotAccept(t *testing.T) {
	f := installation(t)
	issued, err := f.service.Invite(f.ctx, f.platform, "other", &identity.InvitationInput{Account: "dev@example.com", Role: "finance"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken}, nil)
	failure(t, err, fault.Unauthenticated, "existing_account_login_required")
	f.exec(t, "UPDATE users SET active=0 WHERE id='dev'")
	_, err = f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken}, func(expected string) bool { return expected == "dev" })
	failure(t, err, fault.Unauthenticated, "unauthorized")
	f.exec(t, "UPDATE users SET active=1 WHERE id='dev'")
	accepted, err := f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken}, func(expected string) bool { return expected == "dev" })
	if err != nil || !accepted.Accepted || f.count(t, "SELECT count(*) FROM memberships WHERE tenant_id='other' AND user_id='dev' AND role='finance'") != 1 {
		t.Fatal(accepted, err)
	}
}

func TestInvitationAcceptanceExactExpiryPauseAndRollback(t *testing.T) {
	f := installation(t)
	issued, err := f.service.Invite(f.ctx, f.admin, "tenant", &identity.InvitationInput{Account: "new@example.com", Role: "developer", Hours: 1})
	if err != nil {
		t.Fatal(err)
	}
	*f.clock = f.clock.Add(time.Hour)
	_, err = f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken, Password: "controlled-password"}, nil)
	failure(t, err, fault.Invalid, "invitation_unavailable")
	*f.clock = f.clock.Add(-time.Hour)
	f.exec(t, "UPDATE tenants SET active=0 WHERE id='tenant'")
	_, err = f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken, Password: "controlled-password"}, nil)
	failure(t, err, fault.Invalid, "invitation_unavailable")
	f.exec(t, "UPDATE tenants SET active=1; CREATE TRIGGER fail_audit BEFORE INSERT ON console_audit BEGIN SELECT RAISE(ABORT,'test audit failure'); END")
	_, err = f.service.AcceptInvitation(f.ctx, &identity.AcceptInput{Token: issued.AcceptanceToken, Password: "controlled-password"}, nil)
	failure(t, err, fault.Unavailable, "storage_unavailable")
	if f.count(t, "SELECT count(*) FROM users WHERE account='new@example.com'") != 0 || f.count(t, "SELECT count(*) FROM invitations WHERE state='pending'") != 1 {
		t.Fatal("failed accept committed user or consumed invitation")
	}
}

func TestTenantStatusAndAuditRollback(t *testing.T) {
	f := installation(t)
	created, err := f.service.CreateTenant(f.ctx, f.platform, &identity.TenantInput{Name: "Created"})
	if err != nil || !created.Active {
		t.Fatal(created, err)
	}
	_, err = f.service.SetTenantActive(f.ctx, f.admin, "tenant", nil)
	failure(t, err, fault.Forbidden, "forbidden")
	if _, err = f.service.SetTenantActive(f.ctx, f.platform, "tenant", &identity.StatusInput{Active: false}); err != nil {
		t.Fatal(err)
	}
	if f.count(t, "SELECT active FROM tenants WHERE id='tenant'") != 0 {
		t.Fatal("tenant not paused")
	}
	f.exec(t, "CREATE TRIGGER fail_audit BEFORE INSERT ON console_audit BEGIN SELECT RAISE(ABORT,'test audit failure'); END")
	_, err = f.service.SetTenantActive(f.ctx, f.platform, "tenant", &identity.StatusInput{Active: true})
	failure(t, err, fault.Unavailable, "storage_unavailable")
	if f.count(t, "SELECT active FROM tenants WHERE id='tenant'") != 0 {
		t.Fatal("failed audit committed tenant status")
	}
}

const identitySchema = `CREATE TABLE users(id TEXT PRIMARY KEY,account TEXT NOT NULL UNIQUE,salt BLOB NOT NULL,hash BLOB NOT NULL,platform INTEGER NOT NULL,active INTEGER NOT NULL);
CREATE TABLE tenants(id TEXT PRIMARY KEY,name TEXT NOT NULL,active INTEGER NOT NULL,created TEXT NOT NULL);
CREATE TABLE memberships(tenant_id TEXT REFERENCES tenants(id),user_id TEXT REFERENCES users(id),role TEXT NOT NULL,PRIMARY KEY(tenant_id,user_id));
CREATE TABLE invitations(id TEXT PRIMARY KEY,tenant_id TEXT REFERENCES tenants(id),account TEXT NOT NULL,role TEXT NOT NULL,digest BLOB NOT NULL UNIQUE,expires TEXT NOT NULL,state TEXT NOT NULL,accepted_by TEXT REFERENCES users(id));
CREATE TABLE console_audit(id TEXT PRIMARY KEY,time TEXT NOT NULL,actor TEXT NOT NULL,tenant_id TEXT NOT NULL,action TEXT NOT NULL,object_id TEXT NOT NULL);
CREATE TABLE key_diagnostic_members(key_id TEXT,tenant_id TEXT,user_id TEXT,FOREIGN KEY(tenant_id,user_id) REFERENCES memberships(tenant_id,user_id) ON DELETE CASCADE);
CREATE TRIGGER diagnostic_role_cleanup AFTER UPDATE OF role ON memberships WHEN NEW.role!='developer' BEGIN DELETE FROM key_diagnostic_members WHERE tenant_id=NEW.tenant_id AND user_id=NEW.user_id; END;
`
