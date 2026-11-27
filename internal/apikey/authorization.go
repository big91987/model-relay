package apikey

import (
	"database/sql"
	"strings"
	"time"

	"model-relay/internal/secure"
)

// Identify validates an application credential against current key/generation state.
func (service *Service) Identify(credential string) (string, bool) {
	if !strings.HasPrefix(credential, "mr_") || len(credential) != 46 {
		return "", false
	}
	return GenerationIdentity(service.db, credential, service.now())
}

type identityQuery interface{ QueryRow(string, ...any) *sql.Row }

// AuthorizeModel checks the current tenant and key grants through the service's
// persistence dependency. Admission uses the transaction-level function below.
func (service *Service) AuthorizeModel(keyID, alias string) (bool, string) {
	return AuthorizeModel(service.db, keyID, alias)
}

func GenerationIdentity(query identityQuery, credential string, now time.Time) (string, bool) {
	var keyID string
	var generationExpiry, keyExpiry *string
	err := query.QueryRow("SELECT k.id,g.expires,c.expires FROM key_generations g JOIN app_keys k ON k.id=g.key_id JOIN key_controls c ON c.key_id=k.id WHERE g.digest=? AND g.retired=0 AND k.enabled=1 AND c.revoked=0", secure.Digest(credential)).Scan(&keyID, &generationExpiry, &keyExpiry)
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

// ModelAccess records the tenant and both authorization facts without deciding
// whether pause, grant or an existing logical call takes precedence.
type ModelAccess struct {
	TenantID string
	Active   bool
	Allowed  bool
}

// LookupModelAccess uses the caller's query or transaction so authorization and
// admission can share facts while retaining their distinct decision order.
func LookupModelAccess(query identityQuery, keyID, alias string) (ModelAccess, error) {
	var access ModelAccess
	var allowed int
	err := query.QueryRow("SELECT c.tenant_id,t.active,(SELECT count(*) FROM key_model_grants k JOIN tenant_model_grants g ON g.alias=k.alias AND g.tenant_id=c.tenant_id WHERE k.key_id=c.key_id AND k.alias=?) FROM key_controls c JOIN tenants t ON t.id=c.tenant_id WHERE c.key_id=?", alias, keyID).Scan(&access.TenantID, &access.Active, &allowed)
	access.Allowed = allowed == 1
	return access, err
}

func AuthorizeModel(query identityQuery, keyID, alias string) (bool, string) {
	access, err := LookupModelAccess(query, keyID, alias)
	if err != nil {
		return false, "storage_unavailable"
	}
	if !access.Active {
		return false, "tenant_paused"
	}
	if !access.Allowed {
		return false, "model_forbidden"
	}
	return true, ""
}
