package apikey

import (
	"context"
	"crypto/subtle"
	"strings"

	"model-relay/internal/fault"
	"model-relay/internal/secure"
)

// IdentifyLegacy admits enabled pre-generation credentials. Its caller has
// already parsed the transport's authorization header; credential is raw text.
func (service *Service) IdentifyLegacy(credential string) (string, bool) {
	if !strings.HasPrefix(credential, "mr_") {
		return "", false
	}
	var id string
	var saved []byte
	digest := secure.Digest(credential)
	err := service.db.QueryRow("SELECT id,digest FROM app_keys WHERE digest=? AND enabled=1", digest).Scan(&id, &saved)
	return id, err == nil && subtle.ConstantTimeCompare(saved, digest) == 1
}

// Observe retains the recognized identity for rejected-request evidence, even
// when its key is disabled or its generation is retired/expired. Authorization
// is evaluated separately against the supported schema's current rules.
func (service *Service) Observe(credential string, schema int) (string, bool) {
	if !strings.HasPrefix(credential, "mr_") {
		return "", false
	}
	var id string
	var enabled bool
	if schema == 3 {
		if service.db.QueryRow("SELECT k.id,k.enabled FROM key_generations g JOIN app_keys k ON k.id=g.key_id WHERE g.digest=?", secure.Digest(credential)).Scan(&id, &enabled) != nil {
			return "", false
		}
		_, authorized := service.Identify(credential)
		return id, authorized
	}
	if service.db.QueryRow("SELECT id,enabled FROM app_keys WHERE digest=?", secure.Digest(credential)).Scan(&id, &enabled) != nil {
		return "", false
	}
	return id, enabled
}

// ListModels reads the callable aliases granted to a previously authenticated
// application. Its caller holds configuration coordination through admission.
// Schema1 requires the existing explicit storage upgrade before serving.
func (service *Service) ListModels(ctx context.Context, keyID string, schema int) ([]map[string]string, error) {
	query := "SELECT alias FROM models WHERE enabled=1 AND EXISTS(SELECT 1 FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id WHERE c.alias=models.alias AND c.enabled=1 AND u.enabled=1 AND length(u.token)>0)"
	arguments := []any{}
	if schema == 3 {
		var active bool
		var tenantID string
		if service.db.QueryRowContext(ctx, "SELECT t.active,c.tenant_id FROM key_controls c JOIN tenants t ON t.id=c.tenant_id WHERE c.key_id=?", keyID).Scan(&active, &tenantID) != nil {
			return nil, unavailable()
		}
		if !active {
			return nil, fault.New(fault.Forbidden, "tenant_paused")
		}
		query += " AND EXISTS(SELECT 1 FROM tenant_model_grants g JOIN key_model_grants k ON k.alias=g.alias WHERE g.tenant_id=? AND k.key_id=? AND g.alias=models.alias)"
		arguments = append(arguments, tenantID, keyID)
	}
	rows, err := service.db.QueryContext(ctx, query+" ORDER BY alias LIMIT 1000", arguments...)
	if err != nil {
		return nil, unavailable()
	}
	defer rows.Close()
	values := []map[string]string{}
	for rows.Next() {
		var alias string
		if err = rows.Scan(&alias); err != nil {
			return nil, unavailable()
		}
		values = append(values, map[string]string{"id": alias, "object": "model", "owned_by": "model-relay"})
	}
	if rows.Err() != nil {
		return nil, unavailable()
	}
	return values, nil
}
