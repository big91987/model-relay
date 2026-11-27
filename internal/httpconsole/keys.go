package httpconsole

import (
	"errors"
	"net/http"
	"strings"

	"model-relay/internal/access"
	"model-relay/internal/apikey"
	"model-relay/internal/fault"
	"model-relay/internal/httpapi"
)

func (controller *Controller) controlledKeys(writer http.ResponseWriter, request *http.Request, principal access.Principal, tenantID string, parts []string) {
	if len(parts) == 2 && parts[1] == "limits" {
		controller.manageLimits(writer, request, principal, tenantID, "key", parts[0])
		return
	}
	operation := apikey.Operation{Key: request.Header.Get("Idempotency-Key"), Path: request.URL.Path}
	if len(parts) == 0 {
		if request.Method == http.MethodGet {
			values, err := controller.Keys.List(request.Context(), principal, tenantID)
			if err != nil {
				controller.fail(writer, err)
				return
			}
			httpapi.Collection(writer, request, values)
			return
		}
		if request.Method == http.MethodPost {
			result, err := controller.Keys.Create(request.Context(), principal, tenantID, operation, inputCommand[apikey.Input](writer, request))
			if err != nil {
				controller.fail(writer, err)
				return
			}
			if result.AlreadyApplied {
				httpapi.Reply(writer, 200, map[string]any{"id": result.ID, "already_applied": true, "secret_recoverable": false})
			} else {
				httpapi.Reply(writer, 201, map[string]any{"id": result.ID, "key": result.Secret, "secret_recoverable": false})
			}
			return
		}
		_, err := controller.Keys.Create(request.Context(), principal, tenantID, operation, nil)
		var failure *fault.Error
		if errors.As(err, &failure) && failure.Kind == fault.Invalid && failure.Code == "invalid_key" {
			httpapi.Failure(writer, 405, "method_not_allowed", "")
		} else {
			controller.fail(writer, err)
		}
		return
	}
	if request.Method == http.MethodGet {
		value, err := controller.Keys.Read(request.Context(), principal, tenantID, parts[0])
		if err != nil {
			controller.fail(writer, err)
			return
		}
		if len(parts) != 1 {
			httpapi.Failure(writer, 404, "not_found", "")
			return
		}
		httpapi.Reply(writer, 200, value)
		return
	}
	if len(parts) == 1 && request.Method == http.MethodPatch {
		result, err := controller.Keys.Update(request.Context(), principal, tenantID, parts[0], operation, inputCommand[apikey.Input](writer, request))
		if err != nil {
			controller.fail(writer, err)
			return
		}
		if result.AlreadyApplied {
			httpapi.Reply(writer, 200, map[string]any{"id": result.ID, "already_applied": true})
		} else {
			httpapi.Reply(writer, 200, map[string]bool{"ok": true})
		}
		return
	}
	if len(parts) == 2 && request.Method == http.MethodPost {
		switch parts[1] {
		case "rotations", "revoke", "retire-previous":
			input := inputCommand[apikey.Transition](writer, request)
			var result apikey.Result
			var err error
			switch parts[1] {
			case "rotations":
				result, err = controller.Keys.Rotate(request.Context(), principal, tenantID, parts[0], operation, input)
			case "revoke":
				result, err = controller.Keys.Revoke(request.Context(), principal, tenantID, parts[0], operation, input)
			case "retire-previous":
				result, err = controller.Keys.RetirePrevious(request.Context(), principal, tenantID, parts[0], operation, input)
			}
			if err != nil {
				controller.fail(writer, err)
				return
			}
			if result.AlreadyApplied {
				httpapi.Reply(writer, 200, map[string]any{"id": result.ID, "already_applied": true, "secret_recoverable": false})
			} else {
				httpapi.Reply(writer, 200, map[string]any{"id": result.ID, "key": result.Secret, "secret_recoverable": false})
			}
			return
		}
	}
	_, err := controller.Keys.Update(request.Context(), principal, tenantID, parts[0], operation, nil)
	var failure *fault.Error
	if errors.As(err, &failure) && failure.Kind == fault.Invalid && failure.Code == "invalid_key" {
		httpapi.Failure(writer, 404, "not_found", "")
	} else {
		controller.fail(writer, err)
	}
}

func (controller *Controller) keys(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet && request.URL.Path == "/admin/api/keys" {
		values, err := controller.Keys.ListLegacy(request.Context())
		if err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Reply(writer, 200, values)
		return
	}
	if request.Method == http.MethodPost && request.URL.Path == "/admin/api/keys" {
		if controller.Schema == 3 {
			httpapi.Failure(writer, 410, "use_tenant_scoped_keys", "")
			return
		}
		input := inputCommand[struct {
			Name string `json:"name"`
		}](writer, request)
		if input == nil {
			httpapi.Failure(writer, 400, "invalid_name", "")
			return
		}
		value, err := controller.Keys.CreateLegacy(request.Context(), input.Name)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Reply(writer, 201, value)
		return
	}
	if request.Method == http.MethodPatch && strings.HasPrefix(request.URL.Path, "/admin/api/keys/") {
		input := inputCommand[struct {
			Enabled *bool `json:"enabled"`
		}](writer, request)
		if input == nil || input.Enabled == nil {
			httpapi.Failure(writer, 400, "invalid_request", "")
			return
		}
		err := controller.Keys.SetLegacyEnabled(request.Context(), strings.TrimPrefix(request.URL.Path, "/admin/api/keys/"), *input.Enabled)
		if err != nil {
			controller.fail(writer, err)
			return
		}
		httpapi.Reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	httpapi.Failure(writer, 405, "method_not_allowed", "")
}
