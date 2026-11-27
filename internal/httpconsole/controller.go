// Package httpconsole maps management HTTP requests to capability operations.
package httpconsole

import (
	"errors"
	"log"
	"model-relay/internal/apikey"
	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/fault"
	"model-relay/internal/finance"
	"model-relay/internal/httpapi"
	"model-relay/internal/identity"
	"model-relay/internal/operations"
	"model-relay/internal/records"
	"model-relay/internal/reporting"
	"model-relay/internal/sessions"
	"net/http"
	"sync/atomic"
	"time"
)

type Controller struct {
	Schema       int
	Authority    func() string
	Password     func(string) bool
	Now          func() time.Time
	AdvanceClock func() error
	Degraded     *atomic.Bool
	State        *configstate.Coordinator
	Sessions     *sessions.Manager
	Identity     *identity.Service
	Catalog      *catalog.Service
	Keys         *apikey.Service
	Finance      *finance.Service
	Reports      *reporting.Service
	Records      *records.Store
}

func (controller *Controller) storageFailure() {
	controller.Degraded.Store(true)
	log.Print("storage_write_failed")
}
func (controller *Controller) advanceClock(writer http.ResponseWriter) bool {
	if controller.AdvanceClock() != nil {
		controller.storageFailure()
		httpapi.Failure(writer, 503, "storage_unavailable", "")
		return false
	}
	return true
}
func (controller *Controller) fail(writer http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var persistence *catalog.CheckPersistenceError
	if errors.As(err, &persistence) {
		controller.storageFailure()
	}
	var revision *catalog.RevisionConflict
	if errors.As(err, &revision) {
		httpapi.Reply(writer, 409, map[string]any{"error": map[string]string{"message": "Configuration changed; reload before editing", "type": "relay_error", "code": "configuration_conflict"}, "current_revision": revision.CurrentRevision})
		return
	}
	var failure *fault.Error
	status := 503
	code := "storage_unavailable"
	if errors.As(err, &failure) {
		code = failure.Code
		switch failure.Kind {
		case fault.Invalid:
			status = 400
		case fault.Unauthenticated:
			status = 401
		case fault.Forbidden:
			status = 403
		case fault.NotFound:
			status = 404
		case fault.Conflict:
			status = 409
		case fault.Unavailable:
			status = 503
		case fault.RateLimited:
			status = 429
		case fault.Unsupported:
			status = 405
		case fault.TooLarge:
			status = 413
		}
	}
	httpapi.Failure(writer, status, code, "")
}
func (controller *Controller) result(writer http.ResponseWriter, result operations.Result) {
	status := 200
	if result.Created {
		status = 201
	}
	httpapi.Reply(writer, status, result.Value)
}
func (controller *Controller) actor(request *http.Request) string {
	current, _ := controller.getSession(request)
	return current.UserID
}
