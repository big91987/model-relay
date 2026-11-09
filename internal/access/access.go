package access

import (
	"database/sql"
)

type Principal struct {
	ID       string `json:"id"`
	Account  string `json:"account"`
	Platform bool   `json:"platform"`
}

func Role(transaction *sql.Tx, principal Principal, tenantID string) (string, error) {
	var role string
	if principal.Platform {
		err := transaction.QueryRow("SELECT 'administrator' FROM tenants WHERE id=?", tenantID).Scan(&role)
		return role, err
	}
	err := transaction.QueryRow("SELECT m.role FROM memberships m JOIN users u ON u.id=m.user_id WHERE tenant_id=? AND user_id=? AND u.active=1", tenantID, principal.ID).Scan(&role)
	return role, err
}

func ValidRole(role string) bool {
	return role == "administrator" || role == "developer" || role == "finance"
}
