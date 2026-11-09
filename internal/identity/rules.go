package identity

import (
	"golang.org/x/crypto/argon2"
	"model-relay/internal/secure"
	"net/mail"
	"strings"
)

type Tenant struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Created string `json:"created"`
	Role    string `json:"role,omitempty"`
}

type Invitation struct {
	ID      string `json:"id"`
	Tenant  string `json:"tenant_id"`
	Account string `json:"account"`
	Role    string `json:"role"`
	Expires string `json:"expires_at"`
	State   string `json:"state"`
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
	salt := secure.Digest(secure.Random(16))[:16]
	return salt, argon2.IDKey([]byte(password), salt, 3, 64*1024, 1, 32)
}
