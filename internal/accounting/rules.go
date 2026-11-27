// Package accounting owns resource and money rules independently of HTTP and SQL.
package accounting

import (
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strconv"
)

type ResourceLimits struct {
	Calls       *int64 `json:"calls"`
	RPM         *int64 `json:"rpm"`
	TPM         *int64 `json:"tpm"`
	Concurrency *int64 `json:"concurrency"`
	Budget      *int64 `json:"budget_micro_usd"`
	Hard        bool   `json:"hard"`
	Revision    int64  `json:"revision"`
}

type ResourceUsage struct {
	Calls       int64 `json:"calls"`
	RPM         int64 `json:"rpm"`
	TPM         int64 `json:"tpm"`
	Concurrency int64 `json:"concurrency"`
	Fee         int64 `json:"fee_micro_usd,string"`
	Reserved    int64 `json:"reserved_micro_usd,string"`
	Pending     int64 `json:"pending"`
}

type PriceSnapshot struct {
	Upstream int64  `json:"upstream_id"`
	ID       string `json:"price_id"`
	CostIn   int64  `json:"cost_input_rate"`
	CostOut  int64  `json:"cost_output_rate"`
	FeeIn    int64  `json:"tenant_input_rate"`
	FeeOut   int64  `json:"tenant_output_rate"`
	Input    int64  `json:"max_input"`
	Output   int64  `json:"max_output"`
}

func (price PriceSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"upstream_id": price.Upstream, "price_id": price.ID, "cost_input_rate": strconv.FormatInt(price.CostIn, 10), "cost_output_rate": strconv.FormatInt(price.CostOut, 10), "tenant_input_rate": strconv.FormatInt(price.FeeIn, 10), "tenant_output_rate": strconv.FormatInt(price.FeeOut, 10), "max_input": price.Input, "max_output": price.Output})
}

func Charge(input, output, inputRate, outputRate int64) (int64, error) {
	return microAmount(input, output, inputRate, outputRate, 500000)
}

func Reserve(input, output, inputRate, outputRate int64) (int64, error) {
	return microAmount(input, output, inputRate, outputRate, 999999)
}

func microAmount(input, output, inputRate, outputRate, rounding int64) (int64, error) {
	if input < 0 || output < 0 || inputRate < 0 || outputRate < 0 {
		return 0, errors.New("invalid_amount")
	}
	amount := new(big.Int).Mul(big.NewInt(input), big.NewInt(inputRate))
	amount.Add(amount, new(big.Int).Mul(big.NewInt(output), big.NewInt(outputRate)))
	amount.Add(amount, big.NewInt(rounding))
	amount.Div(amount, big.NewInt(1000000))
	if !amount.IsInt64() {
		return 0, errors.New("amount_overflow")
	}
	return amount.Int64(), nil
}

func overLimit(limit *int64, used, additional int64) bool {
	return limit != nil && (used > *limit || additional > *limit-used)
}

func (limits ResourceLimits) RequiresBounds() bool {
	return limits.TPM != nil || limits.Hard && limits.Budget != nil
}

func (limits ResourceLimits) Valid() bool {
	for _, limit := range []*int64{limits.Calls, limits.RPM, limits.TPM, limits.Concurrency, limits.Budget} {
		if limit != nil && (*limit < 0 || *limit > 1000000000000) {
			return false
		}
	}
	return true
}

func (limits ResourceLimits) Exceeds(cap ResourceLimits) bool {
	for index, requested := range []*int64{limits.Calls, limits.RPM, limits.TPM, limits.Concurrency, limits.Budget} {
		upper := []*int64{cap.Calls, cap.RPM, cap.TPM, cap.Concurrency, cap.Budget}[index]
		if upper != nil && (requested == nil || *requested > *upper) {
			return true
		}
	}
	return cap.Budget != nil && cap.Hard && !limits.Hard
}

func (limits ResourceLimits) SoftBudgetExceeded(used ResourceUsage) bool {
	return limits.Budget != nil && !limits.Hard && used.Fee+used.Reserved >= *limits.Budget
}

// CheckAdmission keeps the existing rejection order; the operation adds owner
// identity and maps the result to a response outside this package.
func (limits ResourceLimits) CheckAdmission(used ResourceUsage, tokensReserved, feeReserved int64) string {
	if overLimit(limits.Calls, used.Calls, 1) {
		return "calls_exceeded"
	}
	if overLimit(limits.RPM, used.RPM, 1) {
		return "rpm_exceeded"
	}
	if overLimit(limits.TPM, used.TPM, tokensReserved) {
		return "tpm_exceeded"
	}
	if overLimit(limits.Concurrency, used.Concurrency, 1) {
		return "concurrency_exceeded"
	}
	if limits.Budget != nil && (*limits.Budget == 0 || limits.Hard && overLimit(limits.Budget, used.Fee+used.Reserved, feeReserved)) {
		return "budget_exceeded"
	}
	return ""
}

func ReserveAttempts(prices []PriceSnapshot) (tokensReserved, feeReserved int64, err error) {
	tokenBounds, feeBounds := []int64{}, []int64{}
	for _, price := range prices {
		fee, err := Reserve(price.Input, price.Output, price.FeeIn, price.FeeOut)
		if err != nil || price.Input > 1000000000 || price.Output > 1000000000 {
			return 0, 0, errors.New("amount_overflow")
		}
		tokenBounds = append(tokenBounds, price.Input+price.Output)
		feeBounds = append(feeBounds, fee)
	}
	sort.Slice(tokenBounds, func(left, right int) bool { return tokenBounds[left] > tokenBounds[right] })
	sort.Slice(feeBounds, func(left, right int) bool { return feeBounds[left] > feeBounds[right] })
	for index := 0; index < len(tokenBounds) && index < 3; index++ {
		tokensReserved += tokenBounds[index]
		feeReserved += feeBounds[index]
	}
	return tokensReserved, feeReserved, nil
}
