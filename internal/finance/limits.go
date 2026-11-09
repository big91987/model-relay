package finance

import (
	"context"
	"database/sql"
	"strconv"

	"model-relay/internal/access"
	"model-relay/internal/accounting"
	"model-relay/internal/fault"
	"model-relay/internal/operations"
)

// limitRole preserves tenant membership and diagnostic-key visibility before
// exposing either the limits or the validation errors for a write.
func limitRole(transaction *sql.Tx, principal access.Principal, scope LimitScope, write bool) (string, error) {
	role, err := access.Role(transaction, principal, scope.TenantID)
	if err != nil {
		return "", missing()
	}
	switch scope.OwnerType {
	case "key":
		var tenant string
		if transaction.QueryRow("SELECT c.tenant_id FROM app_keys k JOIN key_controls c ON c.key_id=k.id WHERE k.id=?", scope.OwnerID).Scan(&tenant) != nil || tenant != scope.TenantID || write && role != "administrator" {
			return "", missing()
		}
		if role == "developer" {
			var assigned int
			if transaction.QueryRow("SELECT count(*) FROM key_diagnostic_members WHERE key_id=? AND tenant_id=? AND user_id=?", scope.OwnerID, scope.TenantID, principal.ID).Scan(&assigned) != nil || assigned != 1 {
				return "", missing()
			}
		}
	case "tenant":
		if scope.OwnerID != scope.TenantID {
			return "", missing()
		}
		if role == "developer" {
			return "", fault.New(fault.Forbidden, "assigned_key_scope_required")
		}
	default:
		return "", missing()
	}
	return role, nil
}

func (service *Service) Limits(ctx context.Context, principal access.Principal, scope LimitScope) (LimitsView, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return LimitsView{}, err
	}
	defer transaction.Rollback()
	if _, err = limitRole(transaction, principal, scope, false); err != nil {
		return LimitsView{}, err
	}
	limits, err := LimitsFor(transaction, scope.OwnerType, scope.OwnerID)
	if err != nil {
		return LimitsView{}, storage()
	}
	now := AdmissionClock(transaction, service.now())
	usage, err := UsageFor(transaction, scope.OwnerType, scope.OwnerID, now.Format("2006-01"), now.Format("2006-01-02T15:04"))
	if err != nil {
		return LimitsView{}, storage()
	}
	var remaining *string
	if limits.Budget != nil {
		value := strconv.FormatInt(*limits.Budget-usage.Fee-usage.Reserved, 10)
		remaining = &value
	}
	var platform *accounting.ResourceLimits
	if scope.OwnerType == "tenant" {
		cap, err := platformLimit(transaction, scope.TenantID)
		if err != nil && err != sql.ErrNoRows {
			return LimitsView{}, storage()
		}
		if err == nil {
			platform = &cap
		}
	}
	return LimitsView{Limits: limits, Usage: usage, Period: now.Format("2006-01"), Currency: "USD", UnknownCarry: true, BudgetRemaining: remaining, SoftBudgetExceeded: limits.SoftBudgetExceeded(usage), Platform: platform}, nil
}

func (service *Service) UpdateLimits(ctx context.Context, principal access.Principal, scope LimitScope, input *accounting.ResourceLimits) (operations.Result, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return operations.Result{}, err
	}
	defer transaction.Rollback()
	role, err := limitRole(transaction, principal, scope, true)
	if err != nil {
		return operations.Result{}, err
	}
	if role != "administrator" {
		return operations.Result{}, forbidden()
	}
	if input == nil {
		return operations.Result{}, fault.New(fault.Invalid, "invalid_request")
	}
	if !input.Valid() {
		return operations.Result{}, fault.New(fault.Invalid, "invalid_limit")
	}
	previous, err := LimitsFor(transaction, scope.OwnerType, scope.OwnerID)
	if err != nil || previous.Revision != input.Revision {
		return operations.Result{}, fault.New(fault.Conflict, "revision_conflict")
	}
	if !principal.Platform && scope.OwnerType == "tenant" {
		cap, err := platformLimit(transaction, scope.TenantID)
		if err != nil && err != sql.ErrNoRows {
			return operations.Result{}, storage()
		}
		if input.Exceeds(cap) {
			return operations.Result{}, fault.New(fault.Forbidden, "platform_limit_exceeded")
		}
	}
	if principal.Platform && scope.OwnerType == "tenant" {
		if _, err := transaction.Exec("INSERT INTO platform_limits VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(tenant_id) DO UPDATE SET calls=excluded.calls,rpm=excluded.rpm,tpm=excluded.tpm,concurrency=excluded.concurrency,budget=excluded.budget,hard=excluded.hard,revision=excluded.revision", scope.TenantID, input.Calls, input.RPM, input.TPM, input.Concurrency, input.Budget, input.Hard, previous.Revision+1); err != nil {
			return operations.Result{}, storage()
		}
	}
	if _, err := transaction.Exec("INSERT INTO resource_limits VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(owner_type,owner_id) DO UPDATE SET calls=excluded.calls,rpm=excluded.rpm,tpm=excluded.tpm,concurrency=excluded.concurrency,budget=excluded.budget,hard=excluded.hard,revision=excluded.revision", scope.OwnerType, scope.OwnerID, input.Calls, input.RPM, input.TPM, input.Concurrency, input.Budget, input.Hard, previous.Revision+1); err != nil {
		return operations.Result{}, storage()
	}
	if operations.Audit(transaction, principal.ID, scope.TenantID, "limits.update", scope.OwnerID) != nil || transaction.Commit() != nil {
		return operations.Result{}, storage()
	}
	return ok(), nil
}

func platformLimit(transaction *sql.Tx, tenant string) (accounting.ResourceLimits, error) {
	var cap accounting.ResourceLimits
	err := transaction.QueryRow("SELECT calls,rpm,tpm,concurrency,budget,hard,revision FROM platform_limits WHERE tenant_id=?", tenant).Scan(&cap.Calls, &cap.RPM, &cap.TPM, &cap.Concurrency, &cap.Budget, &cap.Hard, &cap.Revision)
	return cap, err
}
