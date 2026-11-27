package httpconsole

import (
	"errors"
	"net/http"
	"strconv"

	"model-relay/internal/access"
	"model-relay/internal/accounting"
	"model-relay/internal/fault"
	"model-relay/internal/finance"
	"model-relay/internal/httpapi"
)

// inputCommand retains malformed input as a nil command so the operation can
// enforce authorization and entity visibility before returning validation errors.
func inputCommand[Input any](writer http.ResponseWriter, request *http.Request) *Input {
	var value Input
	if httpapi.Input(writer, request, &value) != nil {
		return nil
	}
	return &value
}

func (controller *Controller) manageLimits(writer http.ResponseWriter, request *http.Request, principal access.Principal, tenantID, ownerType, ownerID string) {
	scope := finance.LimitScope{TenantID: tenantID, OwnerType: ownerType, OwnerID: ownerID}
	if request.Method == http.MethodGet {
		value, err := controller.Finance.Limits(request.Context(), principal, scope)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Reply(writer, 200, value)
		return
	}
	if request.Method != http.MethodPut {
		// A nil command performs no write and retains scoped rejection precedence.
		_, err := controller.Finance.UpdateLimits(request.Context(), principal, scope, nil)
		var failure *fault.Error
		if errors.As(err, &failure) && failure.Kind == fault.Invalid && failure.Code == "invalid_request" {
			httpapi.Failure(writer, 403, "forbidden", "")
		} else {
			controller.fail(writer, err)
		}
		return
	}
	result, err := controller.Finance.UpdateLimits(request.Context(), principal, scope, inputCommand[accounting.ResourceLimits](writer, request))
	if err != nil {
		controller.fail(writer, err)
		return
	}
	controller.result(writer, result)
}

func (controller *Controller) manageModelFinance(writer http.ResponseWriter, request *http.Request, principal access.Principal, alias, resource string) {
	if resource == "tenant-grants" {
		if request.Method == http.MethodGet {
			values, err := controller.Catalog.Grants(request.Context(), principal, alias)
			if err != nil {
				controller.fail(writer, err)
				return
			}
			httpapi.Collection(writer, request, values)
			return
		}
		// Empty grants are valid. Check visibility before mapping malformed HTTP
		// input, rather than treating an absent command as a valid empty list.
		if _, err := controller.Catalog.Grants(request.Context(), principal, alias); err != nil {
			controller.fail(writer, err)
			return
		}
		var input *struct {
			Tenants []string `json:"tenants"`
		}
		if request.Method == http.MethodPut {
			input = inputCommand[struct {
				Tenants []string `json:"tenants"`
			}](writer, request)
		}
		if input == nil {
			httpapi.Failure(writer, 400, "invalid_grants", "")
			return
		}
		if err := controller.Catalog.SetGrants(request.Context(), principal, alias, input.Tenants); err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	if resource == "bounds" {
		if request.Method == http.MethodGet {
			values, err := controller.Finance.Bounds(request.Context(), principal, alias)
			if err != nil {
				controller.fail(writer, err)
				return
			}
			httpapi.Collection(writer, request, values, map[string]any{"evidence_kind": "operator_declared_not_supplier_validation"})
			return
		}
		var input *finance.BoundInput
		if request.Method == http.MethodPut {
			input = inputCommand[finance.BoundInput](writer, request)
		}
		result, err := controller.Finance.DeclareBound(request.Context(), principal, alias, input)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		controller.result(writer, result)
		return
	}
	if request.Method == http.MethodGet {
		values, err := controller.Finance.Prices(request.Context(), principal, alias)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Collection(writer, request, values)
		return
	}
	var input *finance.PublishPricesInput
	if request.Method == http.MethodPost {
		input = inputCommand[finance.PublishPricesInput](writer, request)
	}
	result, err := controller.Finance.PublishPrices(request.Context(), principal, alias, request.Header.Get("Idempotency-Key"), request.URL.Path, input)
	if err != nil {
		controller.fail(writer, err)
		return
	}
	controller.result(writer, result)
}

func (controller *Controller) tenantBills(writer http.ResponseWriter, request *http.Request, principal access.Principal, tenantID string) {
	if request.Method == http.MethodGet {
		values, err := controller.Finance.ListBills(request.Context(), principal, tenantID)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Collection(writer, request, values, map[string]any{"currency": "USD", "payment": "not_connected"})
		return
	}
	if request.Method != http.MethodPost {
		_, err := controller.Finance.CloseBill(request.Context(), principal, tenantID, request.Header.Get("Idempotency-Key"), request.URL.Path, nil)
		var failure *fault.Error
		if errors.As(err, &failure) && failure.Kind == fault.Conflict && failure.Code == "period_not_ended" {
			httpapi.Failure(writer, 403, "forbidden", "")
		} else {
			controller.fail(writer, err)
		}
		return
	}
	result, err := controller.Finance.CloseBill(request.Context(), principal, tenantID, request.Header.Get("Idempotency-Key"), request.URL.Path, inputCommand[finance.CloseBillInput](writer, request))
	if err != nil {
		controller.fail(writer, err)
		return
	}
	controller.result(writer, result)
}

func (controller *Controller) manageBill(writer http.ResponseWriter, request *http.Request, principal access.Principal, parts []string) {
	if len(parts) < 1 {
		httpapi.Failure(writer, 404, "not_found", "")
		return
	}
	if request.Method == http.MethodGet {
		if len(parts) == 2 && parts[1] == "export" {
			entries, err := controller.Finance.ExportBill(request.Context(), principal, parts[0])
			if err != nil {
				controller.fail(writer, err)
				return
			}
			values := make([]map[string]any, 0, len(entries))
			for _, entry := range entries {
				value := map[string]any{"kind": entry.Kind, "id": entry.ID, "period": entry.Period, "request_id": entry.RequestID, "reason": entry.Reason, "evidence": entry.Evidence}
				if entry.Ordinal != nil {
					value["ordinal"] = *entry.Ordinal
				}
				if entry.Cost != nil {
					value["cost_micro_usd"] = *entry.Cost
				}
				if entry.Fee != nil {
					value["fee_micro_usd"] = *entry.Fee
				}
				values = append(values, value)
			}
			httpapi.CSV(writer, []string{"kind", "id", "period", "request_id", "ordinal", "cost_micro_usd", "fee_micro_usd", "reason", "evidence"}, values)
			return
		}
		value, err := controller.Finance.Bill(request.Context(), principal, parts[0])
		if err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Reply(writer, 200, value)
		return
	}
	if len(parts) == 2 && parts[1] == "corrections" && request.Method == http.MethodPost {
		result, err := controller.Finance.CorrectBill(request.Context(), principal, parts[0], request.Header.Get("Idempotency-Key"), request.URL.Path, inputCommand[finance.CorrectionInput](writer, request))
		if err != nil {
			controller.fail(writer, err)
			return
		}
		controller.result(writer, result)
		return
	}
	_, err := controller.Finance.CorrectBill(request.Context(), principal, parts[0], request.Header.Get("Idempotency-Key"), request.URL.Path, nil)
	var failure *fault.Error
	if errors.As(err, &failure) && failure.Kind == fault.Invalid && failure.Code == "invalid_correction" {
		httpapi.Failure(writer, 404, "not_found", "")
	} else {
		controller.fail(writer, err)
	}
}

func (controller *Controller) reconcileUsage(writer http.ResponseWriter, request *http.Request, principal access.Principal, requestID, ordinalText string) {
	ordinal, err := strconv.Atoi(ordinalText)
	if err != nil {
		ordinal = 0
	}
	result, err := controller.Finance.ReconcileUsage(request.Context(), principal, requestID, ordinal, request.Header.Get("Idempotency-Key"), request.URL.Path, inputCommand[finance.ReconcileInput](writer, request))
	if err != nil {
		controller.fail(writer, err)
		return
	}
	controller.result(writer, result)
}

func (controller *Controller) reconcileLegacy(writer http.ResponseWriter, request *http.Request, principal access.Principal, requestID string) {
	result, err := controller.Finance.ReconcileLegacy(request.Context(), principal, requestID, request.Header.Get("Idempotency-Key"), request.URL.Path, inputCommand[finance.LegacyInput](writer, request))
	if err != nil {
		controller.fail(writer, err)
		return
	}
	controller.result(writer, result)
}
