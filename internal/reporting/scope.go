package reporting

import (
	"database/sql"

	"model-relay/internal/access"
	"model-relay/internal/fault"
)

// reportScope applies tenant visibility and assigned-key diagnostics to report
// queries. Authentication and membership roles remain owned by access.
func reportScope(transaction *sql.Tx, principal access.Principal, tenantID string) (string, []any, error) {
	if tenantID == "" && principal.Platform {
		return "1=1", nil, nil
	}
	if tenantID == "" {
		return "", nil, fault.New(fault.Invalid, "scope_unavailable")
	}
	role, err := access.Role(transaction, principal, tenantID)
	if err != nil {
		return "", nil, fault.New(fault.NotFound, "scope_unavailable")
	}
	where := "c.tenant_id=?"
	args := []any{tenantID}
	if role == "developer" {
		where += " AND EXISTS(SELECT 1 FROM key_diagnostic_members d WHERE d.key_id=r.key_id AND d.tenant_id=c.tenant_id AND d.user_id=?)"
		args = append(args, principal.ID)
	}
	return where, args, nil
}
