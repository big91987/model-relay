// Package app is the process composition root and public HTTP boundary.
package app

import (
	"log"
	"model-relay/internal/admission"
	"model-relay/internal/apikey"
	"model-relay/internal/catalog"
	"model-relay/internal/configstate"
	"model-relay/internal/finance"
	"model-relay/internal/httpapi"
	"model-relay/internal/httpconsole"
	"model-relay/internal/identity"
	"model-relay/internal/installation"
	"model-relay/internal/records"
	"model-relay/internal/relay"
	"model-relay/internal/reporting"
	"model-relay/internal/routing"
	"model-relay/internal/sessions"
	"model-relay/internal/upstream"
	"model-relay/internal/webui"
	"net/http"
	"strings"
	"sync/atomic"
)

type Server struct {
	Store    *installation.Store
	Config   relay.Config
	Gateway  *relay.Gateway
	Catalog  *catalog.Service
	Records  *records.Store
	state    *configstate.Coordinator
	sessions *sessions.Manager
	router   *routing.Engine
	client   *http.Client
	degraded atomic.Bool
	console  *httpconsole.Controller
}

func New(store *installation.Store, config relay.Config) *Server {
	if config.Clock != nil {
		store.SetClock(config.Clock)
	}
	config.Defaults()
	server := &Server{Store: store, Config: config, state: configstate.New(), sessions: sessions.New(), router: routing.New()}
	server.client = config.HTTPClient
	if server.client == nil {
		server.client = upstream.NewHTTPClient(config.Policy, config.HeaderTimeout)
	}
	client := upstream.NewClient(server.client)
	server.Catalog = catalog.New(store.DB, store.Key, store.Schema, server.state, server.router, client, config.Policy)
	server.Records = records.New(store.DB, store.Schema)
	keys := apikey.New(store.DB, store.Now, server.state, server.router)
	people := identity.New(store.DB, store.Now)
	money := finance.New(store.DB, store.Now)
	server.Gateway = &relay.Gateway{Store: store, Config: &server.Config, State: server.state, Router: server.router, Upstream: client, Catalog: server.Catalog, Keys: keys, Records: server.Records, Admission: admission.New(store.DB, store.Now, store.AdvanceClock), Degraded: &server.degraded}
	server.console = &httpconsole.Controller{Schema: store.Schema, Authority: func() string { return server.Config.Authority }, Password: store.Password, Now: store.Now, AdvanceClock: store.AdvanceClock, Degraded: &server.degraded, State: server.state, Sessions: server.sessions, Identity: people, Catalog: server.Catalog, Keys: keys, Finance: money, Reports: reporting.New(store.DB, store.Now), Records: server.Records}
	return server
}
func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'none'; form-action 'self'")
	if request.Host != server.Config.Authority {
		httpapi.Failure(writer, 403, "invalid_host", "")
		return
	}
	if request.URL.Path != "/healthz" && server.Store.AdvanceClock() != nil {
		server.degraded.Store(true)
		log.Print("storage_write_failed")
		httpapi.Failure(writer, 503, "storage_unavailable", "")
		return
	}
	switch {
	case request.URL.Path == "/healthz":
		server.health(writer, request)
	case strings.HasPrefix(request.URL.Path, "/admin/api/"):
		server.console.ServeHTTP(writer, request)
	case request.URL.Path == "/v1/models" && request.Method == http.MethodGet:
		server.Gateway.ListModels(writer, request)
	case request.URL.Path == "/v1/chat/completions" && request.Method == http.MethodPost:
		server.Gateway.Generate(writer, request)
	case request.Method == http.MethodGet && (request.URL.Path == "/" || request.URL.Path == "/admin" || strings.HasPrefix(request.URL.Path, "/admin/")):
		webui.Serve(writer, request, server.Store.Schema)
	default:
		httpapi.Failure(writer, 404, "not_found", "")
	}
}
func (server *Server) health(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		writer.Header().Set("Allow", "GET, HEAD")
		httpapi.Failure(writer, 405, "method_not_allowed", "")
		return
	}
	status := http.StatusOK
	if server.degraded.Load() || server.Store.DB.PingContext(request.Context()) != nil {
		status = 503
	}
	if request.Method == http.MethodHead {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		return
	}
	if status == 503 {
		httpapi.Failure(writer, status, "storage_unavailable", "")
		return
	}
	httpapi.Reply(writer, 200, map[string]any{"status": "ok", "version": server.Config.Version, "storage_schema": server.Store.Schema})
}
