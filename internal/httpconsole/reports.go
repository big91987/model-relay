package httpconsole

import (
	"model-relay/internal/access"
	"model-relay/internal/httpapi"
	"model-relay/internal/reporting"
	"model-relay/internal/sessions"
	"net/http"
	"strconv"
	"strings"
)

func (controller *Controller) consoleAPI(writer http.ResponseWriter, request *http.Request, current sessions.Session) bool {
	path := strings.TrimPrefix(request.URL.Path, "/admin/api/")
	workspace := strings.HasPrefix(path, "workspace/")
	path = strings.TrimPrefix(path, "workspace/")
	parts := strings.Split(path, "/")
	if parts[0] == "requests" && !workspace && request.URL.Query().Get("tenant") == "" {
		return false
	}
	known := (workspace && parts[0] == "overview") || parts[0] == "catalog" || parts[0] == "usage" || parts[0] == "audit" || parts[0] == "requests" || parts[0] == "bills" || parts[0] == "operations" || (parts[0] == "models" && len(parts) == 3 && (parts[2] == "prices" || parts[2] == "bounds" || parts[2] == "tenant-grants")) || (parts[0] == "tenants" && len(parts) >= 3 && (parts[2] == "keys" || parts[2] == "limits" || parts[2] == "bills"))
	if !known {
		return false
	}
	p, err := controller.Identity.Principal(current.UserID)
	if err != nil {
		httpapi.Failure(writer, 401, "unauthorized", "")
		return true
	}
	if parts[0] == "models" && len(parts) == 3 && parts[2] == "tenant-grants" {
		if controller.advanceClock(writer) {
			controller.manageModelFinance(writer, request, p, parts[1], parts[2])
		}
		return true
	}
	if request.Method != http.MethodGet {
		controller.State.Lock()
		defer controller.State.Unlock()
	}
	if !controller.advanceClock(writer) {
		return true
	}
	switch parts[0] {
	case "overview":
		if request.Method != http.MethodGet {
			httpapi.Failure(writer, 405, "method_not_allowed", "")
		} else {
			v, e := controller.Reports.Overview(request.Context(), p, request.URL.Query().Get("tenant"))
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Reply(writer, 200, v)
			}
		}
	case "operations":
		if len(parts) != 2 || request.Method != http.MethodGet {
			httpapi.Failure(writer, 404, "not_found", "")
		} else {
			v, e := controller.Reports.Operation(request.Context(), p, parts[1])
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Reply(writer, 200, v)
			}
		}
	case "tenants":
		switch parts[2] {
		case "keys":
			controller.controlledKeys(writer, request, p, parts[1], parts[3:])
		case "limits":
			controller.manageLimits(writer, request, p, parts[1], "tenant", parts[1])
		case "bills":
			controller.tenantBills(writer, request, p, parts[1])
		}
	case "models":
		controller.manageModelFinance(writer, request, p, parts[1], parts[2])
	case "catalog":
		if request.Method != http.MethodGet {
			httpapi.Failure(writer, 405, "method_not_allowed", "")
		} else {
			v, e := controller.Reports.Catalog(request.Context(), p, reporting.CatalogInput{Tenant: request.URL.Query().Get("tenant"), IncludeUngranted: request.URL.Query().Get("include_ungranted") == "true"})
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Collection(writer, request, v)
			}
		}
	case "bills":
		controller.manageBill(writer, request, p, parts[1:])
	case "requests", "usage", "audit":
		controller.consoleReports(writer, request, p, parts)
	}
	return true
}
func (controller *Controller) consoleReports(writer http.ResponseWriter, request *http.Request, p access.Principal, parts []string) {
	if len(parts) == 3 && parts[0] == "requests" && parts[2] == "legacy-reconciliations" && request.Method == http.MethodPost {
		controller.reconcileLegacy(writer, request, p, parts[1])
		return
	}
	if len(parts) == 4 && parts[0] == "requests" && parts[2] == "reconciliations" && request.Method == http.MethodPost {
		controller.reconcileUsage(writer, request, p, parts[1], parts[3])
		return
	}
	if request.Method != http.MethodGet {
		httpapi.Failure(writer, 405, "method_not_allowed", "")
		return
	}
	q := request.URL.Query()
	filters := reporting.Filters{Tenant: q.Get("tenant"), Model: q.Get("model"), Key: q.Get("key"), State: q.Get("state"), RequestID: q.Get("request_id"), From: q.Get("from"), To: q.Get("to")}
	page := reporting.PageInput{Page: q.Get("page"), Size: q.Get("page_size")}
	var value any
	var err error
	switch {
	case parts[0] == "audit":
		value, err = controller.Reports.Audit(request.Context(), p, filters.Tenant)
	case len(parts) == 2 && parts[1] != "export" && parts[0] == "requests":
		value, err = controller.Reports.Request(request.Context(), p, parts[1], filters)
	case parts[0] == "usage" && len(parts) == 1:
		value, err = controller.Reports.Usage(request.Context(), p, filters)
	case len(parts) == 2 && parts[1] == "export":
		values, e := controller.Reports.ExportRequests(request.Context(), p, filters, page)
		if e != nil {
			if e.Error() == "export_requires_narrower_filter" {
				httpapi.Failure(writer, 413, e.Error(), "")
			} else {
				controller.fail(writer, e)
			}
			return
		}
		httpapi.CSV(writer, []string{"id", "time", "tenant_id", "key_id", "alias", "state", "error_code", "duration_ms", "prompt_tokens", "completion_tokens", "total_tokens", "confirmed_input_tokens", "confirmed_output_tokens", "period", "pending", "fee_reserved", "cost_micro_usd", "fee_micro_usd", "legacy_pending"}, values)
		return
	default:
		value, err = controller.Reports.Requests(request.Context(), p, filters, page)
	}
	if err != nil {
		controller.fail(writer, err)
	} else {
		httpapi.Reply(writer, 200, value)
	}
}
func (controller *Controller) records(writer http.ResponseWriter, request *http.Request) {
	limit := 50
	cursor := int64(1 << 62)
	if raw := request.URL.Query().Get("limit"); raw != "" {
		v, e := strconv.Atoi(raw)
		if e != nil || v < 1 || v > 100 {
			httpapi.Failure(writer, 400, "invalid_cursor", "")
			return
		}
		limit = v
	}
	if raw := request.URL.Query().Get("cursor"); raw != "" {
		v, e := strconv.ParseInt(raw, 10, 64)
		if e != nil || v < 1 {
			httpapi.Failure(writer, 400, "invalid_cursor", "")
			return
		}
		cursor = v
	}
	values, next, err := controller.Records.LegacyList(request.Context(), cursor, limit)
	if err != nil {
		controller.fail(writer, err)
	} else {
		httpapi.Reply(writer, 200, map[string]any{"data": values, "next_cursor": next})
	}
}
func (controller *Controller) record(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", "GET")
		httpapi.Failure(writer, 405, "method_not_allowed", "")
		return
	}
	value, err := controller.Records.Detail(request.Context(), strings.TrimPrefix(request.URL.Path, "/admin/api/requests/"))
	if err != nil {
		controller.fail(writer, err)
	} else {
		httpapi.Reply(writer, 200, value)
	}
}
func (controller *Controller) overview(writer http.ResponseWriter, request *http.Request) {
	value, clearSuccess, err := controller.Reports.LegacyOverview(request.Context(), controller.State.SuccessID(), controller.Degraded.Load())
	if clearSuccess {
		controller.State.ClearSuccess()
	}
	if err != nil {
		controller.fail(writer, err)
	} else {
		httpapi.Reply(writer, 200, value)
	}
}
