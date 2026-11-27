package identity

import (
	"crypto/subtle"
	"golang.org/x/crypto/argon2"
	"model-relay/internal/access"
	"model-relay/internal/secure"
	"strings"
)

func (service *Service) Authenticate(account, password string) (access.Principal, bool) {
	if account == "" {
		account = "admin"
	}
	if account != "admin" {
		account = strings.ToLower(strings.TrimSpace(account))
	}
	var principal access.Principal
	var salt, hash []byte
	err := service.db.QueryRow("SELECT id,account,platform,salt,hash FROM users WHERE account=? AND active=1", account).Scan(&principal.ID, &principal.Account, &principal.Platform, &salt, &hash)
	if err != nil {
		salt = secure.Digest("model-relay-missing-account")[:16]
		hash = make([]byte, 32)
	}
	actual := argon2.IDKey([]byte(password), salt, 3, 64*1024, 1, 32)
	return principal, err == nil && subtle.ConstantTimeCompare(actual, hash) == 1
}

func (service *Service) Principal(userID string) (access.Principal, error) {
	var principal access.Principal
	err := service.db.QueryRow("SELECT id,account,platform FROM users WHERE id=? AND active=1", userID).Scan(&principal.ID, &principal.Account, &principal.Platform)
	return principal, err
}
