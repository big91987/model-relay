// Package reporting owns authorized management read models. Read operations
// retain a single database snapshot and the caller's explicit effective clock.
package reporting

import (
	"context"
	"database/sql"
	"time"

	"model-relay/internal/fault"
)

type Service struct {
	db  *sql.DB
	now func() time.Time
}

func New(db *sql.DB, now func() time.Time) *Service { return &Service{db: db, now: now} }
func (service *Service) begin(ctx context.Context) (*sql.Tx, error) {
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, storage()
	}
	return transaction, nil
}
func storage() error { return fault.New(fault.Unavailable, "storage_unavailable") }

type Filters struct{ Tenant, Model, Key, State, RequestID, From, To string }

// PageInput retains unparsed values so scope and time-filter errors can retain
// precedence over page errors. Export validates these values before its cap.
type PageInput struct{ Page, Size string }
type RequestPage struct {
	Items []map[string]any `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"page_size"`
}
type UsageSummary struct {
	Items            []map[string]any `json:"items"`
	Currency         string           `json:"currency"`
	UnknownIsNotZero bool             `json:"unknown_is_not_zero"`
}
type AuditPage struct {
	Items              []map[string]any `json:"items"`
	Size               int              `json:"page_size"`
	MoreRequiresFilter bool             `json:"more_requires_filter"`
}
type OperationView struct {
	ObjectID          string `json:"object_id"`
	Action            string `json:"action"`
	SecretRecoverable bool   `json:"secret_recoverable"`
}
type Overview struct {
	Today            string           `json:"today"`
	Period           string           `json:"period"`
	Timezone         string           `json:"timezone"`
	RequestsToday    int64            `json:"requests_today"`
	SuccessRate      *float64         `json:"success_rate_percent"`
	AverageDuration  *float64         `json:"average_duration_ms"`
	Fee              string           `json:"confirmed_fee_micro_usd"`
	Pending          int64            `json:"pending_attempts"`
	LegacyPending    int64            `json:"legacy_pending_requests"`
	Reserved         string           `json:"reserved_micro_usd"`
	Trend            []map[string]any `json:"trend"`
	Distribution     []map[string]any `json:"tenant_distribution"`
	Currency         string           `json:"currency"`
	UnknownIsNotZero bool             `json:"unknown_is_not_zero"`
	ObservedAt       string           `json:"observed_at"`
}
type CatalogInput struct {
	Tenant           string
	IncludeUngranted bool
}
type CatalogItem struct {
	Alias        string         `json:"alias"`
	Enabled      bool           `json:"enabled"`
	Revision     int64          `json:"revision"`
	Strategy     string         `json:"routing_strategy"`
	TenantInput  *int64         `json:"tenant_input_rate"`
	TenantOutput *int64         `json:"tenant_output_rate"`
	Effective    *string        `json:"effective_at"`
	Currency     string         `json:"currency"`
	Unit         string         `json:"unit"`
	Protocol     string         `json:"protocol"`
	Stream       bool           `json:"stream"`
	Summary      map[string]any `json:"summary"`
}
