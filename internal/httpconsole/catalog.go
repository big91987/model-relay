package httpconsole

import (
	"net/http"
	"strconv"
	"strings"

	"model-relay/internal/catalog"
	"model-relay/internal/httpapi"
	"model-relay/internal/names"
)

func catalogPageLimit(request *http.Request) (int, bool) {
	limit := 50
	if value := request.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			return 0, false
		}
		limit = parsed
	}
	return limit, true
}

func (controller *Controller) upstreams(writer http.ResponseWriter, request *http.Request) {
	path := strings.TrimPrefix(request.URL.Path, "/admin/api/upstreams")
	if path == "" && request.Method == http.MethodGet {
		limit, ok := catalogPageLimit(request)
		cursor := int64(0)
		if raw := request.URL.Query().Get("cursor"); raw != "" {
			var err error
			cursor, err = strconv.ParseInt(raw, 10, 64)
			ok = ok && err == nil && cursor > 0
		}
		if !ok {
			httpapi.Failure(writer, 400, "invalid_cursor", "")
			return
		}
		values, err := controller.Catalog.ListUpstreams(request.Context(), cursor, limit+1)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		var next any
		if len(values) > limit {
			values = values[:limit]
			next = strconv.FormatInt(values[len(values)-1].ID, 10)
		}
		httpapi.Reply(writer, 200, map[string]any{"data": values, "next_cursor": next})
		return
	}
	id := int64(0)
	if path != "" {
		raw := strings.TrimPrefix(strings.TrimSuffix(path, "/check"), "/")
		var err error
		id, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || id < 1 {
			httpapi.Failure(writer, 404, "not_found", "")
			return
		}
	}
	if strings.HasSuffix(path, "/check") {
		controller.checkUpstream(writer, request, id)
		return
	}
	if request.Method == http.MethodGet && id > 0 {
		value, err := controller.Catalog.ReadUpstream(request.Context(), id)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Reply(writer, 200, value)
		return
	}
	creating := path == "" && request.Method == http.MethodPost
	if !creating && !(id > 0 && request.Method == http.MethodPatch) {
		httpapi.Failure(writer, 405, "method_not_allowed", "")
		return
	}
	var input catalog.UpstreamInput
	if httpapi.Input(writer, request, &input) != nil {
		httpapi.Failure(writer, 400, "invalid_request", "")
		return
	}
	var value catalog.Upstream
	var err error
	if creating {
		value, err = controller.Catalog.CreateUpstream(request.Context(), controller.actor(request), input)
	} else {
		value, err = controller.Catalog.UpdateUpstream(request.Context(), controller.actor(request), id, input)
	}
	if err != nil {
		controller.fail(writer, err)
		return
	}
	status := 200
	if creating {
		status = 201
	}
	httpapi.Reply(writer, status, value)
}

func (controller *Controller) routeModels(writer http.ResponseWriter, request *http.Request) {
	alias := strings.TrimPrefix(request.URL.Path, "/admin/api/models/")
	collection := request.URL.Path == "/admin/api/models"
	if !collection && !names.Alias(alias) {
		httpapi.Failure(writer, 404, "not_found", "")
		return
	}
	if request.Method == http.MethodGet {
		if !collection {
			value, err := controller.Catalog.ReadModel(request.Context(), alias)
			if err != nil {
				controller.fail(writer, err)
				return
			}
			httpapi.Reply(writer, 200, value)
			return
		}
		limit, ok := catalogPageLimit(request)
		cursor := request.URL.Query().Get("cursor")
		if !ok || cursor != "" && !names.Alias(cursor) {
			httpapi.Failure(writer, 400, "invalid_cursor", "")
			return
		}
		values, err := controller.Catalog.ListModels(request.Context(), cursor, limit+1)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		var next any
		if len(values) > limit {
			values = values[:limit]
			next = values[len(values)-1].Alias
		}
		httpapi.Reply(writer, 200, map[string]any{"data": values, "next_cursor": next})
		return
	}
	creating := collection && request.Method == http.MethodPost
	if !creating && !(request.Method == http.MethodPatch && !collection) {
		httpapi.Failure(writer, 405, "method_not_allowed", "")
		return
	}
	var input catalog.ModelInput
	if httpapi.Input(writer, request, &input) != nil {
		httpapi.Failure(writer, 400, "invalid_request", "")
		return
	}
	var value catalog.RouteModel
	var err error
	if creating {
		value, err = controller.Catalog.CreateModel(request.Context(), controller.actor(request), input)
	} else {
		value, err = controller.Catalog.UpdateModel(request.Context(), controller.actor(request), alias, input)
	}
	if err != nil {
		controller.fail(writer, err)
		return
	}
	status := 200
	if creating {
		status = 201
	}
	httpapi.Reply(writer, status, value)
}

func (controller *Controller) checkUpstream(writer http.ResponseWriter, request *http.Request, id int64) {
	if request.Method != http.MethodPost {
		httpapi.Failure(writer, 405, "method_not_allowed", "")
		return
	}
	var input struct {
		Expected *int64 `json:"expected_execution_revision"`
	}
	if httpapi.Input(writer, request, &input) != nil || input.Expected == nil {
		httpapi.Failure(writer, 400, "invalid_request", "")
		return
	}
	value, err := controller.Catalog.Check(request.Context(), id, *input.Expected)
	if err != nil {
		controller.fail(writer, err)
		return
	}
	httpapi.Reply(writer, 200, value)
}
