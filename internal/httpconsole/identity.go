package httpconsole

import (
	"crypto/subtle"
	"model-relay/internal/access"
	"model-relay/internal/httpapi"
	"model-relay/internal/identity"
	"model-relay/internal/sessions"
	"net/http"
	"strings"
)

func command[T any](writer http.ResponseWriter, request *http.Request) *T {
	var value T
	if httpapi.Input(writer, request, &value) != nil {
		return nil
	}
	return &value
}
func (controller *Controller) identityAPI(writer http.ResponseWriter, request *http.Request, current sessions.Session) bool {
	path := strings.TrimPrefix(request.URL.Path, "/admin/api/")
	parts := strings.Split(path, "/")
	known := path == "identity" || path == "tenants" || (len(parts) == 2 && parts[0] == "tenants") || (len(parts) >= 3 && parts[0] == "tenants" && (parts[2] == "members" || parts[2] == "invitations"))
	if !known {
		return false
	}
	p, err := controller.Identity.Principal(current.UserID)
	if err != nil {
		httpapi.Failure(writer, 401, "unauthorized", "")
		return true
	}
	if path == "identity" && request.Method == http.MethodGet {
		httpapi.Reply(writer, 200, p)
		return true
	}
	if request.Method != http.MethodGet {
		controller.State.Lock()
		defer controller.State.Unlock()
	}
	if !controller.advanceClock(writer) {
		return true
	}
	switch {
	case path == "tenants" && request.Method == http.MethodGet:
		value, err := controller.Reports.Tenants(request.Context(), p)
		if err != nil {
			controller.fail(writer, err)
		} else {
			httpapi.Collection(writer, request, value)
		}
	case path == "tenants" && request.Method == http.MethodPost:
		value, err := controller.Identity.CreateTenant(request.Context(), p, command[identity.TenantInput](writer, request))
		if err != nil {
			controller.fail(writer, err)
		} else {
			httpapi.Reply(writer, 201, value)
		}
	case len(parts) == 2 && parts[0] == "tenants" && request.Method == http.MethodPatch:
		value, err := controller.Identity.SetTenantActive(request.Context(), p, parts[1], command[identity.StatusInput](writer, request))
		if err != nil {
			controller.fail(writer, err)
		} else {
			httpapi.Reply(writer, 200, value)
		}
	case len(parts) >= 3 && parts[2] == "invitations":
		switch {
		case len(parts) == 3 && request.Method == http.MethodGet:
			value, e := controller.Identity.Invitations(request.Context(), p, parts[1])
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Collection(writer, request, value)
			}
		case len(parts) == 3 && request.Method == http.MethodPost:
			value, e := controller.Identity.Invite(request.Context(), p, parts[1], command[identity.InvitationInput](writer, request))
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Reply(writer, 201, value)
			}
		case len(parts) == 5 && parts[4] == "revoke" && request.Method == http.MethodPost:
			value, e := controller.Identity.RevokeInvitation(request.Context(), p, parts[1], parts[3], command[identity.RevokeInput](writer, request))
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Reply(writer, 200, value)
			}
		default:
			controller.identityFallback(writer, request, p, parts)
		}
	case len(parts) >= 3 && parts[2] == "members":
		switch {
		case len(parts) == 3 && request.Method == http.MethodGet:
			value, e := controller.Identity.Members(request.Context(), p, parts[1])
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Collection(writer, request, value)
			}
		case len(parts) == 4 && request.Method == http.MethodPatch:
			value, e := controller.Identity.ChangeMemberRole(request.Context(), p, parts[1], parts[3], command[identity.MemberInput](writer, request))
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Reply(writer, 200, value)
			}
		case len(parts) == 4 && request.Method == http.MethodDelete:
			value, e := controller.Identity.RemoveMember(request.Context(), p, parts[1], parts[3])
			if e != nil {
				controller.fail(writer, e)
			} else {
				httpapi.Reply(writer, 200, value)
			}
		default:
			controller.identityFallback(writer, request, p, parts)
		}
	default:
		controller.identityFallback(writer, request, p, parts)
	}
	return true
}

// Unsupported identity routes retain the same tenant-scope checks as supported
// operations before reporting that no such operation exists.
func (controller *Controller) identityFallback(writer http.ResponseWriter, request *http.Request, principal access.Principal, parts []string) {
	if len(parts) < 2 || parts[0] != "tenants" {
		httpapi.Failure(writer, 404, "not_found", "")
		return
	}
	intent := identity.TenantAccess{Write: request.Method != http.MethodGet, Invitations: len(parts) >= 3 && parts[2] == "invitations"}
	if len(parts) == 4 && parts[2] == "members" && request.Method == http.MethodDelete {
		intent.RemoveMemberID = parts[3]
	}
	if err := controller.Identity.CheckTenantAccess(request.Context(), principal, parts[1], intent); err != nil {
		controller.fail(writer, err)
		return
	}
	httpapi.Failure(writer, 404, "not_found", "")
}

func (controller *Controller) acceptInvitation(writer http.ResponseWriter, request *http.Request) {
	release, rejection := controller.Sessions.BeginInvitation(controller.Now())
	if rejection != sessions.None {
		code := "login_rate_limited"
		if rejection == sessions.AttemptLimit {
			writer.Header().Set("Retry-After", "60")
			code = "invitation_rate_limited"
		}
		httpapi.Failure(writer, 429, code, "")
		return
	}
	defer release()
	input := command[identity.AcceptInput](writer, request)
	if input == nil || len(input.Token) != 43 || len(input.Password) > 1024 {
		httpapi.Failure(writer, 400, "invalid_invitation", "")
		return
	}
	controller.State.Lock()
	defer controller.State.Unlock()
	if !controller.advanceClock(writer) {
		return
	}
	value, err := controller.Identity.AcceptInvitation(request.Context(), input, func(userID string) bool {
		current, ok := controller.getSession(request)
		return ok && current.UserID == userID && subtle.ConstantTimeCompare([]byte(current.CSRF), []byte(request.Header.Get("X-CSRF-Token"))) == 1
	})
	if err != nil {
		controller.fail(writer, err)
	} else {
		httpapi.Reply(writer, 200, value)
	}
}
