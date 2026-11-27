package finance

import (
	"model-relay/internal/names"
	"time"
)

func validPeriod(period string) bool {
	parsed, err := time.Parse("2006-01", period)
	return err == nil && parsed.Format("2006-01") == period
}

func validCorrection(input CorrectionInput) bool {
	return names.Valid(input.Reason) && input.Amount != 0 && input.Amount >= -1000000000000 && input.Amount <= 1000000000000
}

func validReconciliation(input ReconcileInput) bool {
	return input.Input >= 0 && input.Output >= 0 && input.Input <= 1000000000 && input.Output <= 1000000000 && names.Valid(input.Reason) && names.Valid(input.Evidence)
}

func validBound(input BoundInput) bool {
	return input.Input > 0 && input.Output > 0 && input.Input <= 1000000000 && input.Output <= 1000000000 && names.Valid(input.Evidence)
}
