package identity_test

import (
	"testing"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/identity"
)

func TestTenantAccessKeepsScopeAdminAndSelfRemovalPrecedence(t *testing.T) {
	f := installation(t)
	for _, scenario := range []struct {
		name      string
		principal access.Principal
		tenant    string
		intent    identity.TenantAccess
		code      string
		kind      fault.Kind
	}{
		{name: "member read", principal: f.dev, tenant: "tenant"},
		{name: "administrator write", principal: f.admin, tenant: "tenant", intent: identity.TenantAccess{Write: true}},
		{name: "platform invitations", principal: f.platform, tenant: "tenant", intent: identity.TenantAccess{Invitations: true}},
		{name: "member write", principal: f.dev, tenant: "tenant", intent: identity.TenantAccess{Write: true}, code: "forbidden", kind: fault.Forbidden},
		{name: "member invitations read", principal: f.dev, tenant: "tenant", intent: identity.TenantAccess{Invitations: true}, code: "forbidden", kind: fault.Forbidden},
		{name: "self removal", principal: f.dev, tenant: "tenant", intent: identity.TenantAccess{Write: true, RemoveMemberID: "dev"}},
		{name: "other removal", principal: f.dev, tenant: "tenant", intent: identity.TenantAccess{Write: true, RemoveMemberID: "admin"}, code: "forbidden", kind: fault.Forbidden},
		{name: "invitations ignore self exemption", principal: f.dev, tenant: "tenant", intent: identity.TenantAccess{Write: true, Invitations: true, RemoveMemberID: "dev"}, code: "forbidden", kind: fault.Forbidden},
		{name: "unrelated tenant before write", principal: f.dev, tenant: "other", intent: identity.TenantAccess{Write: true}, code: "not_found", kind: fault.NotFound},
		{name: "missing tenant before admin", principal: f.platform, tenant: "missing", intent: identity.TenantAccess{Write: true, Invitations: true}, code: "not_found", kind: fault.NotFound},
		{name: "inactive member", principal: access.Principal{ID: "inactive"}, tenant: "tenant", code: "not_found", kind: fault.NotFound},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			err := f.service.CheckTenantAccess(f.ctx, scenario.principal, scenario.tenant, scenario.intent)
			if scenario.code == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				failure(t, err, scenario.kind, scenario.code)
			}
		})
	}
}
