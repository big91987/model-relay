package apikey

import (
	"context"
	"database/sql"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/configstate"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
	"model-relay/internal/routing"
	"model-relay/internal/secure"
	"model-relay/internal/sqlread"
)

type Service struct {
	db     *sql.DB
	now    func() time.Time
	state  *configstate.Coordinator
	router *routing.Engine
}

// New shares configuration coordination and routing with other configuration
// services. Legacy writes own that coordination; tenant-scoped operations
// retain their existing caller-held coordination contract.
func New(db *sql.DB, now func() time.Time, state *configstate.Coordinator, router *routing.Engine) *Service {
	return &Service{db: db, now: now, state: state, router: router}
}

func unavailable() error { return fault.New(fault.Unavailable, "storage_unavailable") }

func (service *Service) begin(ctx context.Context, principal access.Principal, tenantID string, write bool) (*sql.Tx, string, error) {
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", unavailable()
	}
	role, err := access.Role(transaction, principal, tenantID)
	if err != nil {
		transaction.Rollback()
		return nil, "", fault.New(fault.NotFound, "not_found")
	}
	if write && role != "administrator" {
		transaction.Rollback()
		return nil, role, fault.New(fault.Forbidden, "forbidden")
	}
	return transaction, role, nil
}

func scopedKey(transaction *sql.Tx, principal access.Principal, tenantID, keyID string, write bool) (ControlledKey, error) {
	key, _, err := Scope(transaction, principal, keyID, write)
	if err != nil || key.Tenant != tenantID {
		return key, fault.New(fault.NotFound, "not_found")
	}
	return key, nil
}

func (service *Service) List(ctx context.Context, principal access.Principal, tenantID string) ([]ControlledKey, error) {
	transaction, role, err := service.begin(ctx, principal, tenantID, false)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	query := "SELECT key_id FROM key_controls WHERE tenant_id=? ORDER BY key_id"
	arguments := []any{tenantID}
	if role == "developer" {
		query = "SELECT c.key_id FROM key_controls c JOIN key_diagnostic_members d ON d.key_id=c.key_id WHERE c.tenant_id=? AND d.user_id=? ORDER BY c.key_id"
		arguments = append(arguments, principal.ID)
	}
	ids, err := sqlread.Strings(transaction, query, arguments...)
	if err != nil {
		return nil, unavailable()
	}
	keys := []ControlledKey{}
	for _, id := range ids {
		key, err := scan(transaction, id)
		if err != nil {
			return nil, unavailable()
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (service *Service) Read(ctx context.Context, principal access.Principal, tenantID, keyID string) (ControlledKey, error) {
	transaction, _, err := service.begin(ctx, principal, tenantID, false)
	if err != nil {
		return ControlledKey{}, err
	}
	defer transaction.Rollback()
	return scopedKey(transaction, principal, tenantID, keyID, false)
}

// validGrants checks current tenant model grants and active developer membership.
// Shape and transition rules remain independent of persistence in rules.go.
func validGrants(transaction *sql.Tx, tenantID string, input Input) bool {
	for _, alias := range input.Models {
		var count int
		if transaction.QueryRow("SELECT count(*) FROM tenant_model_grants WHERE tenant_id=? AND alias=?", tenantID, alias).Scan(&count) != nil || count != 1 {
			return false
		}
	}
	for _, userID := range input.Diagnostics {
		var count int
		if transaction.QueryRow("SELECT count(*) FROM memberships m JOIN users u ON u.id=m.user_id WHERE tenant_id=? AND user_id=? AND role='developer' AND u.active=1", tenantID, userID).Scan(&count) != nil || count != 1 {
			return false
		}
	}
	return true
}

func (service *Service) Create(ctx context.Context, principal access.Principal, tenantID string, operation Operation, input *Input) (Result, error) {
	transaction, _, err := service.begin(ctx, principal, tenantID, true)
	if err != nil {
		return Result{}, err
	}
	defer transaction.Rollback()
	if input == nil {
		return Result{}, fault.New(fault.Invalid, "invalid_key")
	}
	key, prior, fingerprint, err := operations.Begin(transaction, principal.ID, operation.Key, operation.Path, "key.create", input)
	if err != nil {
		return Result{}, fault.New(fault.Conflict, err.Error())
	}
	if prior != "" {
		return Result{ID: prior, AlreadyApplied: true}, nil
	}
	if !validInput(*input, service.now()) || !validGrants(transaction, tenantID, *input) {
		return Result{}, fault.New(fault.Invalid, "invalid_key")
	}
	var active bool
	if transaction.QueryRow("SELECT active FROM tenants WHERE id=?", tenantID).Scan(&active) != nil || !active {
		return Result{}, fault.New(fault.Conflict, "tenant_paused")
	}
	keyID, secret := secure.Random(16), "mr_"+secure.Random(32)
	created := service.now().Format(time.RFC3339Nano)
	if _, err := transaction.Exec("INSERT INTO app_keys VALUES(?,?,?,?,?,?); INSERT INTO key_controls VALUES(?,?,?,?,0,1,0); INSERT INTO key_generations VALUES(?,?,?,?,?,NULL,0); INSERT INTO resource_limits VALUES('key',?,NULL,NULL,NULL,NULL,NULL,1,1)", keyID, input.Name, secure.Digest(secret), secret[:11], input.Enabled, created, keyID, tenantID, input.Application, input.Expires, secure.Random(16), keyID, secure.Digest(secret), secret[:11], created, keyID); err != nil || setKeyGrants(transaction, keyID, tenantID, *input) != nil || operations.Complete(transaction, principal.ID, key, fingerprint, "key.create", keyID, tenantID) != nil || transaction.Commit() != nil {
		return Result{}, unavailable()
	}
	return Result{ID: keyID, Secret: secret}, nil
}

func (service *Service) Update(ctx context.Context, principal access.Principal, tenantID, keyID string, operation Operation, input *Input) (Result, error) {
	transaction, _, err := service.begin(ctx, principal, tenantID, true)
	if err != nil {
		return Result{}, err
	}
	defer transaction.Rollback()
	key, err := scopedKey(transaction, principal, tenantID, keyID, true)
	if err != nil {
		return Result{}, err
	}
	if input == nil {
		return Result{}, fault.New(fault.Invalid, "invalid_key")
	}
	operationKey, prior, fingerprint, err := operations.Begin(transaction, principal.ID, operation.Key, operation.Path, "key.update", input)
	if err != nil {
		return Result{}, fault.New(fault.Conflict, err.Error())
	}
	if prior != "" {
		return Result{ID: prior, AlreadyApplied: true}, nil
	}
	if !validInput(*input, service.now()) || !validGrants(transaction, tenantID, *input) {
		return Result{}, fault.New(fault.Invalid, "invalid_key")
	}
	if input.Revision != key.Revision || key.Revoked {
		return Result{}, fault.New(fault.Conflict, "revision_conflict_or_revoked")
	}
	var active bool
	if transaction.QueryRow("SELECT active FROM tenants WHERE id=?", tenantID).Scan(&active) != nil {
		return Result{}, unavailable()
	}
	if !active && pausedExpansion(*input, key) {
		return Result{}, fault.New(fault.Conflict, "tenant_paused")
	}
	if _, err := transaction.Exec("UPDATE app_keys SET name=?,enabled=? WHERE id=?; UPDATE key_controls SET application=?,expires=?,revision=revision+1,legacy=0 WHERE key_id=?", input.Name, input.Enabled, keyID, input.Application, input.Expires, keyID); err != nil || setKeyGrants(transaction, keyID, tenantID, *input) != nil || operations.Complete(transaction, principal.ID, operationKey, fingerprint, "key.update", keyID, tenantID) != nil || transaction.Commit() != nil {
		return Result{}, unavailable()
	}
	return Result{ID: keyID}, nil
}

func (service *Service) Rotate(ctx context.Context, principal access.Principal, tenantID, keyID string, operation Operation, input *Transition) (Result, error) {
	return service.transition(ctx, principal, tenantID, keyID, operation, input, "rotations")
}

func (service *Service) Revoke(ctx context.Context, principal access.Principal, tenantID, keyID string, operation Operation, input *Transition) (Result, error) {
	return service.transition(ctx, principal, tenantID, keyID, operation, input, "revoke")
}

func (service *Service) RetirePrevious(ctx context.Context, principal access.Principal, tenantID, keyID string, operation Operation, input *Transition) (Result, error) {
	return service.transition(ctx, principal, tenantID, keyID, operation, input, "retire-previous")
}

func (service *Service) transition(ctx context.Context, principal access.Principal, tenantID, keyID string, operation Operation, input *Transition, action string) (Result, error) {
	transaction, _, err := service.begin(ctx, principal, tenantID, true)
	if err != nil {
		return Result{}, err
	}
	defer transaction.Rollback()
	key, err := scopedKey(transaction, principal, tenantID, keyID, true)
	if err != nil {
		return Result{}, err
	}
	if input == nil || input.Hours < 0 || input.Hours > 24 {
		return Result{}, fault.New(fault.Invalid, "invalid_request")
	}
	operationKey, prior, fingerprint, err := operations.Begin(transaction, principal.ID, operation.Key, operation.Path, "key."+action, input)
	if err != nil {
		return Result{}, fault.New(fault.Conflict, err.Error())
	}
	if prior != "" {
		return Result{ID: prior, AlreadyApplied: true}, nil
	}
	if input.Revision != key.Revision || key.Revoked {
		return Result{}, fault.New(fault.Conflict, "revision_conflict_or_revoked")
	}
	secret := ""
	switch action {
	case "revoke":
		_, err = transaction.Exec("UPDATE key_controls SET revoked=1,revision=revision+1 WHERE key_id=?; UPDATE app_keys SET enabled=0 WHERE id=?; UPDATE key_generations SET retired=1 WHERE key_id=?", keyID, keyID, keyID)
	case "retire-previous":
		_, err = transaction.Exec("UPDATE key_generations SET retired=1 WHERE key_id=? AND expires IS NOT NULL; UPDATE key_controls SET revision=revision+1 WHERE key_id=?", keyID, keyID)
	case "rotations":
		secret, err = service.rotate(transaction, tenantID, keyID, input.Hours)
		if err != nil {
			return Result{}, err
		}
	}
	if err != nil || operations.Complete(transaction, principal.ID, operationKey, fingerprint, "key."+action, keyID, tenantID) != nil || transaction.Commit() != nil {
		return Result{}, unavailable()
	}
	return Result{ID: keyID, Secret: secret}, nil
}

func (service *Service) rotate(transaction *sql.Tx, tenantID, keyID string, hours int) (string, error) {
	var active bool
	if transaction.QueryRow("SELECT active FROM tenants WHERE id=?", tenantID).Scan(&active) != nil || !active {
		return "", fault.New(fault.Conflict, "tenant_paused")
	}
	now := service.now()
	previous, err := sqlread.Maps(transaction, "SELECT id,expires FROM key_generations WHERE key_id=? AND retired=0 AND expires IS NOT NULL", keyID)
	if err != nil {
		return "", unavailable()
	}
	for _, generation := range previous {
		expiry, err := time.Parse(time.RFC3339Nano, generation["expires"].(string))
		if err != nil || now.Before(expiry) {
			return "", fault.New(fault.Conflict, "transition_already_active")
		}
		if _, err := transaction.Exec("UPDATE key_generations SET retired=1 WHERE id=?", generation["id"]); err != nil {
			return "", unavailable()
		}
	}
	secret := "mr_" + secure.Random(32)
	_, err = transaction.Exec("UPDATE key_generations SET expires=?,retired=? WHERE key_id=? AND retired=0 AND expires IS NULL; INSERT INTO key_generations VALUES(?,?,?,?,?,NULL,0); UPDATE app_keys SET digest=?,hint=? WHERE id=?; UPDATE key_controls SET revision=revision+1 WHERE key_id=?", now.Add(time.Duration(hours)*time.Hour).Format(time.RFC3339Nano), hours == 0, keyID, secure.Random(16), keyID, secure.Digest(secret), secret[:11], now.Format(time.RFC3339Nano), secure.Digest(secret), secret[:11], keyID, keyID)
	if err != nil {
		return "", unavailable()
	}
	return secret, nil
}
