package httpconsole

import (
	"crypto/subtle"
	"model-relay/internal/httpapi"
	"model-relay/internal/secure"
	"model-relay/internal/sessions"
	"net/http"
	"strings"
	"time"
)

func (controller *Controller) getSession(request *http.Request) (sessions.Session, bool) {
	cookie, err := request.Cookie("mr_session")
	if err != nil {
		return sessions.Session{}, false
	}
	return controller.Sessions.Lookup(time.Now(), cookie.Value)
}
func (controller *Controller) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	write := request.Method != http.MethodGet
	if write && request.Header.Get("Origin") != "http://"+controller.Authority() {
		httpapi.Failure(writer, 403, "untrusted_origin", "")
		return
	}
	if request.URL.Path == "/admin/api/login" && request.Method == http.MethodPost {
		controller.login(writer, request)
		return
	}
	if controller.Schema == 3 && request.URL.Path == "/admin/api/invitations/accept" && request.Method == http.MethodPost {
		controller.acceptInvitation(writer, request)
		return
	}
	current, ok := controller.getSession(request)
	if !ok {
		httpapi.Failure(writer, 401, "unauthorized", "")
		return
	}
	if write && subtle.ConstantTimeCompare([]byte(current.CSRF), []byte(request.Header.Get("X-CSRF-Token"))) != 1 {
		httpapi.Failure(writer, 403, "invalid_csrf", "")
		return
	}
	if request.URL.Path == "/admin/api/session" && request.Method == http.MethodGet {
		value := map[string]any{"csrf": current.CSRF, "degraded": controller.Degraded.Load(), "expires_at": current.Expiry.UTC().Format(time.RFC3339)}
		if controller.Schema == 3 {
			p, err := controller.Identity.Principal(current.UserID)
			if err != nil {
				httpapi.Failure(writer, 401, "unauthorized", "")
				return
			}
			value["user"] = p
		}
		httpapi.Reply(writer, 200, value)
		return
	}
	if request.URL.Path == "/admin/api/logout" && request.Method == http.MethodPost {
		var body struct{}
		if httpapi.Input(writer, request, &body) != nil {
			httpapi.Failure(writer, 400, "invalid_request", "")
			return
		}
		cookie, _ := request.Cookie("mr_session")
		controller.Sessions.Delete(cookie.Value)
		http.SetCookie(writer, &http.Cookie{Name: "mr_session", Value: "", Path: "/admin", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
		httpapi.Reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	if controller.Schema == 3 {
		if controller.identityAPI(writer, request, current) || controller.consoleAPI(writer, request, current) {
			return
		}
		p, err := controller.Identity.Principal(current.UserID)
		if err != nil {
			httpapi.Failure(writer, 401, "unauthorized", "")
			return
		}
		if !p.Platform {
			httpapi.Failure(writer, 403, "forbidden", "")
			return
		}
	}
	if request.URL.Path == "/admin/api/upstream" {
		httpapi.Failure(writer, 410, "management_api_changed", "")
		return
	}
	response := newAdminResponse(writer)
	defer response.flush()
	writer = response
	if request.URL.Path == "/admin/api/upstreams" || strings.HasPrefix(request.URL.Path, "/admin/api/upstreams/") {
		controller.upstreams(writer, request)
		return
	}
	if request.URL.Path == "/admin/api/models" || strings.HasPrefix(request.URL.Path, "/admin/api/models/") {
		controller.routeModels(writer, request)
		return
	}
	if request.URL.Path == "/admin/api/keys" || strings.HasPrefix(request.URL.Path, "/admin/api/keys/") {
		controller.keys(writer, request)
		return
	}
	if request.URL.Path == "/admin/api/overview" && request.Method == http.MethodGet {
		controller.State.Lock()
		defer controller.State.Unlock()
	}
	switch {
	case request.URL.Path == "/admin/api/overview" && request.Method == http.MethodGet:
		controller.overview(writer, request)
	case request.URL.Path == "/admin/api/requests" && request.Method == http.MethodGet:
		controller.records(writer, request)
	case strings.HasPrefix(request.URL.Path, "/admin/api/requests/"):
		controller.record(writer, request)
	default:
		httpapi.Failure(writer, 404, "not_found", "")
	}
}
func (controller *Controller) login(writer http.ResponseWriter, request *http.Request) {
	now := time.Now()
	release, rejection := controller.Sessions.BeginLogin(now)
	if rejection != sessions.None {
		if rejection == sessions.AttemptLimit {
			writer.Header().Set("Retry-After", "60")
		}
		httpapi.Failure(writer, 429, "login_rate_limited", "")
		return
	}
	defer release()
	var input struct {
		Password string `json:"password"`
		Account  string `json:"account"`
	}
	if httpapi.Input(writer, request, &input) != nil {
		httpapi.Failure(writer, 400, "invalid_request", "")
		return
	}
	userID := ""
	valid := false
	if controller.Schema == 3 {
		p, ok := controller.Identity.Authenticate(input.Account, input.Password)
		userID, valid = p.ID, ok
	} else {
		valid = input.Account == "" && controller.Password(input.Password)
	}
	if !valid {
		httpapi.Failure(writer, 401, "unauthorized", "")
		return
	}
	secret, csrf := secure.Random(32), secure.Random(32)
	current, ok := controller.Sessions.Create(now, secret, userID, csrf)
	if !ok {
		httpapi.Failure(writer, 429, "session_limit", "")
		return
	}
	http.SetCookie(writer, &http.Cookie{Name: "mr_session", Value: secret, Path: "/admin", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 3600})
	httpapi.Reply(writer, 200, map[string]string{"csrf": csrf, "expires_at": current.Expiry.UTC().Format(time.RFC3339)})
}
