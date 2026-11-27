package finance

import "model-relay/internal/accounting"

type CloseBillInput struct {
	Period string `json:"period"`
}
type CorrectionInput struct {
	Period string `json:"adjustment_period"`
	Amount int64  `json:"amount_micro_usd"`
	Reason string `json:"reason"`
}
type ReconcileInput struct {
	Input    int64  `json:"input_tokens"`
	Output   int64  `json:"output_tokens"`
	Reason   string `json:"reason"`
	Evidence string `json:"evidence"`
	Revision int64  `json:"expected_revision"`
}
type LegacyInput struct {
	Revision int64           `json:"expected_revision"`
	Reason   string          `json:"reason"`
	Evidence string          `json:"evidence"`
	Attempts []LegacyAttempt `json:"attempts"`
}
type LegacyAttempt struct {
	Ordinal  int    `json:"ordinal"`
	Upstream int64  `json:"upstream_id"`
	PriceID  string `json:"historical_price_reference"`
	Input    int64  `json:"input_tokens"`
	Output   int64  `json:"output_tokens"`
	CostIn   int64  `json:"cost_input_rate"`
	CostOut  int64  `json:"cost_output_rate"`
	FeeIn    int64  `json:"tenant_input_rate"`
	FeeOut   int64  `json:"tenant_output_rate"`
}
type PublishPricesInput struct {
	Effective string                     `json:"effective_at"`
	Prices    []accounting.PriceSnapshot `json:"prices"`
}
type BoundInput struct {
	Upstream int64  `json:"upstream_id"`
	Input    int64  `json:"max_input"`
	Output   int64  `json:"max_output"`
	Evidence string `json:"evidence"`
}
type LimitScope struct{ TenantID, OwnerType, OwnerID string }
type LimitsView struct {
	Limits             accounting.ResourceLimits  `json:"limits"`
	Usage              accounting.ResourceUsage   `json:"usage"`
	Period             string                     `json:"period"`
	Currency           string                     `json:"currency"`
	UnknownCarry       bool                       `json:"unknown_carry"`
	BudgetRemaining    *string                    `json:"budget_remaining_micro_usd"`
	SoftBudgetExceeded bool                       `json:"soft_budget_exceeded"`
	Platform           *accounting.ResourceLimits `json:"platform_limits"`
}
type Price struct {
	Price     accounting.PriceSnapshot `json:"price"`
	Effective string                   `json:"effective_at"`
	Currency  string                   `json:"currency"`
	Unit      string                   `json:"unit"`
}
type Bound struct {
	Upstream           int64  `json:"upstream_id"`
	Input              int64  `json:"max_input"`
	Output             int64  `json:"max_output"`
	Evidence           string `json:"evidence"`
	UpstreamExecution  int64  `json:"upstream_execution"`
	CandidateExecution int64  `json:"candidate_execution"`
	Applicable         int64  `json:"applicable"`
}
type BillPeriod struct {
	Period       string  `json:"period"`
	Cost         string  `json:"cost_micro_usd"`
	UsageFee     string  `json:"usage_fee_micro_usd"`
	Adjustment   string  `json:"adjustment_micro_usd"`
	Fee          string  `json:"fee_micro_usd"`
	Pending      int64   `json:"pending"`
	BillID       *string `json:"bill_id"`
	Closed       *string `json:"closed"`
	OriginalFee  *string `json:"original_fee"`
	OriginalCost *string `json:"original_cost"`
}
type BillCorrection struct {
	ID      string `json:"id"`
	Period  string `json:"period"`
	Amount  string `json:"amount"`
	Reason  string `json:"reason"`
	Actor   string `json:"actor"`
	Created string `json:"created"`
}
type BillDetail struct {
	ID           string           `json:"id"`
	TenantID     string           `json:"tenant_id"`
	Period       string           `json:"period"`
	OriginalFee  string           `json:"original_fee"`
	OriginalCost string           `json:"original_cost"`
	CurrentFee   string           `json:"current_fee_micro_usd"`
	Adjustment   string           `json:"adjustment_micro_usd"`
	Closed       string           `json:"closed_at"`
	Corrections  []BillCorrection `json:"corrections"`
	Currency     string           `json:"currency"`
	Payment      string           `json:"payment"`
}
type BillEntry struct {
	Kind      string  `json:"kind"`
	ID        string  `json:"id"`
	Period    string  `json:"period"`
	RequestID string  `json:"request_id"`
	Ordinal   *int64  `json:"ordinal"`
	Cost      *string `json:"cost_micro_usd"`
	Fee       *string `json:"fee_micro_usd"`
	Reason    string  `json:"reason"`
	Evidence  string  `json:"evidence"`
}
