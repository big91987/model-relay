// Package relay owns the synchronous inference request and attempt lifecycle.
package relay

import (
	"context"
	"log"
	"model-relay/internal/admission"
	"model-relay/internal/apikey"
	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/httpapi"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/routing"
	"model-relay/internal/upstream"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type Config struct {
	Clock                                                              func() time.Time
	LatencyClock                                                       func() time.Time
	HTTPClient                                                         *http.Client
	Authority                                                          string
	Version                                                            string
	Policy                                                             upstream.Policy
	TotalTimeout, IdleTimeout, HeaderTimeout, AttemptTimeout, Cooldown time.Duration
}

func (config *Config) Defaults() {
	if config.TotalTimeout == 0 {
		config.TotalTimeout = 300 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 60 * time.Second
	}
	if config.HeaderTimeout == 0 {
		config.HeaderTimeout = 30 * time.Second
	}
	if config.AttemptTimeout == 0 {
		config.AttemptTimeout = 60 * time.Second
	}
	if config.Cooldown == 0 {
		config.Cooldown = 30 * time.Second
	}
}

type Gateway struct {
	Store     *installation.Store
	Config    *Config
	State     *configstate.Coordinator
	Router    *routing.Engine
	Upstream  *upstream.Client
	Catalog   *catalog.Service
	Keys      *apikey.Service
	Records   *records.Store
	Admission *admission.Service
	Degraded  *atomic.Bool
}

func (server *Gateway) storageFailure() {
	server.Degraded.Store(true)
	log.Print("storage_write_failed")
}
func (server *Gateway) appID(request *http.Request) (string, bool) {
	credential := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	if server.Store.Schema == 3 {
		return server.Keys.Identify(credential)
	}
	if !strings.HasPrefix(request.Header.Get("Authorization"), "Bearer mr_") {
		return "", false
	}
	return server.Keys.IdentifyLegacy(credential)
}
func (server *Gateway) identifyKey(request *http.Request) (string, bool) {
	if !strings.HasPrefix(request.Header.Get("Authorization"), "Bearer mr_") {
		return "", false
	}
	return server.Keys.Observe(strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer "), server.Store.Schema)
}
func (server *Gateway) snapshot(alias string) (catalog.Snapshot, error) {
	return server.Catalog.ModelSnapshot(context.Background(), alias)
}
func (server *Gateway) ListModels(writer http.ResponseWriter, request *http.Request) {
	server.State.RLock()
	defer server.State.RUnlock()
	if server.Store.AdvanceClock() != nil {
		server.storageFailure()
		httpapi.Failure(writer, 503, "storage_unavailable", "")
		return
	}
	key, ok := server.appID(request)
	if !ok {
		httpapi.Failure(writer, 401, "unauthorized", "")
		return
	}
	values, err := server.Keys.ListModels(request.Context(), key, server.Store.Schema)
	if err != nil {
		code := "storage_unavailable"
		status := 503
		if err.Error() == "tenant_paused" {
			code = "tenant_paused"
			status = 403
		}
		httpapi.Failure(writer, status, code, "")
		return
	}
	httpapi.Reply(writer, 200, map[string]any{"object": "list", "data": values})
}
