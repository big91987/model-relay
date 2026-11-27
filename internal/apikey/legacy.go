package apikey

import (
	"context"
	"time"

	"model-relay/internal/fault"
	"model-relay/internal/names"
	"model-relay/internal/secure"
)

type LegacyResult struct {
	Metadata AppKey `json:"metadata"`
	Secret   string `json:"key"`
}

// ListLegacy is the supported platform-wide metadata list, including schema3
// keys. It deliberately does not expose digests, generations or issued secrets.
func (service *Service) ListLegacy(ctx context.Context) ([]AppKey, error) {
	rows, err := service.db.QueryContext(ctx, "SELECT id,name,hint,enabled,created FROM app_keys ORDER BY created DESC LIMIT 1000")
	if err != nil {
		return nil, unavailable()
	}
	defer rows.Close()
	values := []AppKey{}
	for rows.Next() {
		var value AppKey
		if err = rows.Scan(&value.ID, &value.Name, &value.Hint, &value.Enabled, &value.Created); err != nil {
			return nil, unavailable()
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return nil, unavailable()
	}
	return values, nil
}

// CreateLegacy is used only after the adapter has rejected schema3 issuance.
// It retains configuration coordination through the persisted revision commit
// and publication of the new runtime generation.
func (service *Service) CreateLegacy(ctx context.Context, name string) (LegacyResult, error) {
	service.state.Lock()
	defer service.state.Unlock()
	if !names.Valid(name) {
		return LegacyResult{}, fault.New(fault.Invalid, "invalid_name")
	}
	secret := "mr_" + secure.Random(32)
	key := AppKey{ID: secure.Random(16), Name: name, Hint: secret[:11], Enabled: true, Created: service.now().UTC().Format(time.RFC3339Nano)}
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return LegacyResult{}, unavailable()
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, "INSERT INTO app_keys VALUES(?,?,?,?,1,?)", key.ID, key.Name, secure.Digest(secret), key.Hint, key.Created); err != nil {
		return LegacyResult{}, unavailable()
	}
	if _, err = transaction.ExecContext(ctx, "UPDATE meta SET config_revision=config_revision+1"); err != nil {
		return LegacyResult{}, unavailable()
	}
	if transaction.Commit() != nil {
		return LegacyResult{}, unavailable()
	}
	service.state.Changed(service.router)
	return LegacyResult{Metadata: key, Secret: secret}, nil
}

// SetLegacyEnabled also supports the existing global schema3 endpoint. The
// runtime change is published under coordination only after commit succeeds.
func (service *Service) SetLegacyEnabled(ctx context.Context, id string, enabled bool) error {
	service.state.Lock()
	defer service.state.Unlock()
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return unavailable()
	}
	defer transaction.Rollback()
	result, err := transaction.ExecContext(ctx, "UPDATE app_keys SET enabled=? WHERE id=?", enabled, id)
	if err != nil {
		return unavailable()
	}
	count, err := result.RowsAffected()
	if err != nil {
		return unavailable()
	}
	if count == 0 {
		return fault.New(fault.NotFound, "not_found")
	}
	if _, err = transaction.ExecContext(ctx, "UPDATE meta SET config_revision=config_revision+1"); err != nil {
		return unavailable()
	}
	if transaction.Commit() != nil {
		return unavailable()
	}
	service.state.Changed(service.router)
	return nil
}
