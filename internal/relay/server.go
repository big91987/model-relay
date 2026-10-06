package relay

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	Authority      string
	Version        string
	Policy         Policy
	TotalTimeout   time.Duration
	IdleTimeout    time.Duration
	HeaderTimeout  time.Duration
	AttemptTimeout time.Duration
	Cooldown       time.Duration
}
type session struct {
	CSRF   string
	Expiry time.Time
}
type Server struct {
	Store            *Store
	Config           Config
	mu               sync.RWMutex
	authMu           sync.Mutex
	sessions         map[string]session
	attempts         []time.Time
	passwordSlots    chan struct{}
	degraded         atomic.Bool
	client           *http.Client
	configEpoch      uint64
	successID        string
	routing          map[string]*routeRuntime
	checking         map[int64]bool
	health           map[string]*routeHealth
	dispatchSequence int64
}

func New(store *Store, config Config) *Server {
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
	transport := &http.Transport{Proxy: nil, DialContext: config.Policy.Dial, DisableKeepAlives: true, ForceAttemptHTTP2: false, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: config.HeaderTimeout}
	return &Server{Store: store, Config: config, sessions: make(map[string]session), routing: make(map[string]*routeRuntime), health: make(map[string]*routeHealth), checking: make(map[int64]bool), passwordSlots: make(chan struct{}, 2), client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func reply(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
func failure(writer http.ResponseWriter, status int, code, id string) {
	reply(writer, status, map[string]any{"error": map[string]string{"message": "Request could not be completed: " + code, "type": "relay_error", "code": code}, "request_id": id})
}
func (server *Server) storageFailure() {
	server.degraded.Store(true)
	log.Print("storage_write_failed")
}
func jsonInput(writer http.ResponseWriter, request *http.Request, target any) error {
	kind, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || kind != "application/json" {
		return errors.New("JSON required")
	}
	_ = http.NewResponseController(writer).SetReadDeadline(time.Now().Add(10 * time.Second))
	defer http.NewResponseController(writer).SetReadDeadline(time.Time{})
	data, err := io.ReadAll(http.MaxBytesReader(writer, request.Body, 64*1024))
	if err != nil {
		return err
	}
	if !uniqueJSON(data) {
		return errors.New("invalid JSON")
	}
	if _, ok := object(data); !ok {
		return errors.New("JSON object required")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func validName(value string) bool {
	return len(value) > 0 && len(value) <= 128 && !strings.ContainsAny(value, "\r\n\x00")
}
func validAlias(value string) bool {
	if !validName(value) {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("._-", character)) {
			return false
		}
	}
	return true
}
func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'none'; form-action 'self'")
	if request.Host != server.Config.Authority {
		failure(writer, 403, "invalid_host", "")
		return
	}
	switch {
	case request.URL.Path == "/healthz":
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			writer.Header().Set("Allow", "GET, HEAD")
			failure(writer, http.StatusMethodNotAllowed, "method_not_allowed", "")
			return
		}
		status := http.StatusOK
		if server.degraded.Load() || server.Store.DB.PingContext(request.Context()) != nil {
			status = http.StatusServiceUnavailable
		}
		if request.Method == http.MethodHead {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(status)
			return
		}
		if status == http.StatusServiceUnavailable {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]any{"status": "ok", "version": server.Config.Version, "storage_schema": server.Store.Schema})
	case strings.HasPrefix(request.URL.Path, "/admin/api/"):
		server.admin(writer, request)
	case request.URL.Path == "/v1/models" && request.Method == http.MethodGet:
		server.listModels(writer, request)
	case request.URL.Path == "/v1/chat/completions" && request.Method == http.MethodPost:
		server.generate(writer, request)
	case request.Method == http.MethodGet && (request.URL.Path == "/" || request.URL.Path == "/admin" || strings.HasPrefix(request.URL.Path, "/admin/")):
		serveWeb(writer, request)
	default:
		failure(writer, 404, "not_found", "")
	}
}
func (server *Server) getSession(request *http.Request) (session, bool) {
	cookie, err := request.Cookie("mr_session")
	if err != nil {
		return session{}, false
	}
	server.authMu.Lock()
	defer server.authMu.Unlock()
	key := string(digest(cookie.Value))
	value, ok := server.sessions[key]
	if !ok || time.Now().After(value.Expiry) {
		delete(server.sessions, key)
		return session{}, false
	}
	return value, true
}
func (server *Server) admin(writer http.ResponseWriter, request *http.Request) {
	write := request.Method != http.MethodGet
	if write && request.Header.Get("Origin") != "http://"+server.Config.Authority {
		failure(writer, 403, "untrusted_origin", "")
		return
	}
	if request.URL.Path == "/admin/api/login" && request.Method == http.MethodPost {
		server.login(writer, request)
		return
	}
	current, ok := server.getSession(request)
	if !ok {
		failure(writer, 401, "unauthorized", "")
		return
	}
	if write && subtle.ConstantTimeCompare([]byte(current.CSRF), []byte(request.Header.Get("X-CSRF-Token"))) != 1 {
		failure(writer, 403, "invalid_csrf", "")
		return
	}
	if request.URL.Path == "/admin/api/session" && request.Method == http.MethodGet {
		reply(writer, 200, map[string]any{"csrf": current.CSRF, "degraded": server.degraded.Load(), "expires_at": current.Expiry.UTC().Format(time.RFC3339)})
		return
	}
	if request.URL.Path == "/admin/api/logout" && request.Method == http.MethodPost {
		var body struct{}
		if jsonInput(writer, request, &body) != nil {
			failure(writer, 400, "invalid_request", "")
			return
		}
		cookie, _ := request.Cookie("mr_session")
		server.authMu.Lock()
		delete(server.sessions, string(digest(cookie.Value)))
		server.authMu.Unlock()
		http.SetCookie(writer, &http.Cookie{Name: "mr_session", Value: "", Path: "/admin", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
		reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	if request.URL.Path == "/admin/api/upstream" {
		failure(writer, 410, "management_api_changed", "")
		return
	}
	response := newAdminResponse(writer)
	defer response.flush()
	writer = response
	if request.URL.Path == "/admin/api/upstreams" || strings.HasPrefix(request.URL.Path, "/admin/api/upstreams/") {
		server.upstreams(writer, request)
		return
	}
	if request.URL.Path == "/admin/api/models" || strings.HasPrefix(request.URL.Path, "/admin/api/models/") {
		server.routeModels(writer, request)
		return
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	switch {
	case request.URL.Path == "/admin/api/overview" && request.Method == http.MethodGet:
		server.overview(writer, request)
	case request.URL.Path == "/admin/api/keys" || strings.HasPrefix(request.URL.Path, "/admin/api/keys/"):
		server.keys(writer, request)
	case request.URL.Path == "/admin/api/requests" && request.Method == http.MethodGet:
		server.records(writer, request)
	case strings.HasPrefix(request.URL.Path, "/admin/api/requests/"):
		server.record(writer, request)
	default:
		failure(writer, 404, "not_found", "")
	}
}
func (server *Server) login(writer http.ResponseWriter, request *http.Request) {
	now := time.Now()
	server.authMu.Lock()
	cutoff := now.Add(-time.Minute)
	filtered := server.attempts[:0]
	for _, attempt := range server.attempts {
		if attempt.After(cutoff) {
			filtered = append(filtered, attempt)
		}
	}
	server.attempts = filtered
	if len(filtered) >= 10 {
		server.authMu.Unlock()
		writer.Header().Set("Retry-After", "60")
		failure(writer, 429, "login_rate_limited", "")
		return
	}
	server.attempts = append(server.attempts, now)
	server.authMu.Unlock()
	select {
	case server.passwordSlots <- struct{}{}:
		defer func() { <-server.passwordSlots }()
	default:
		failure(writer, 429, "login_rate_limited", "")
		return
	}
	var input struct {
		Password string `json:"password"`
	}
	if jsonInput(writer, request, &input) != nil {
		failure(writer, 400, "invalid_request", "")
		return
	}
	if !server.Store.Password(input.Password) {
		failure(writer, 401, "unauthorized", "")
		return
	}
	secret := randomSecret(32)
	csrf := randomSecret(32)
	server.authMu.Lock()
	for key, value := range server.sessions {
		if now.After(value.Expiry) {
			delete(server.sessions, key)
		}
	}
	if len(server.sessions) >= 128 {
		server.authMu.Unlock()
		failure(writer, 429, "session_limit", "")
		return
	}
	server.sessions[string(digest(secret))] = session{CSRF: csrf, Expiry: now.Add(8 * time.Hour)}
	server.authMu.Unlock()
	http.SetCookie(writer, &http.Cookie{Name: "mr_session", Value: secret, Path: "/admin", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 3600})
	reply(writer, 200, map[string]string{"csrf": csrf, "expires_at": now.Add(8 * time.Hour).UTC().Format(time.RFC3339)})
}

type Model struct {
	Alias   string `json:"alias"`
	Model   string `json:"model"`
	Enabled bool   `json:"enabled"`
}

type AppKey struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Hint    string `json:"hint"`
	Enabled bool   `json:"enabled"`
	Created string `json:"created"`
}

func (server *Server) keys(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet && request.URL.Path == "/admin/api/keys" {
		rows, err := server.Store.DB.Query("SELECT id,name,hint,enabled,created FROM app_keys ORDER BY created DESC LIMIT 1000")
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		defer rows.Close()
		list := []AppKey{}
		for rows.Next() {
			var value AppKey
			if rows.Scan(&value.ID, &value.Name, &value.Hint, &value.Enabled, &value.Created) != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			list = append(list, value)
		}
		if rows.Err() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, list)
		return
	}
	if request.Method == http.MethodPost && request.URL.Path == "/admin/api/keys" {
		var input struct {
			Name string `json:"name"`
		}
		if jsonInput(writer, request, &input) != nil || !validName(input.Name) {
			failure(writer, 400, "invalid_name", "")
			return
		}
		secret := "mr_" + randomSecret(32)
		key := AppKey{ID: randomSecret(16), Name: input.Name, Hint: secret[:11], Enabled: true, Created: time.Now().UTC().Format(time.RFC3339Nano)}
		transaction, err := server.Store.DB.Begin()
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		defer transaction.Rollback()
		if _, err := transaction.Exec("INSERT INTO app_keys VALUES(?,?,?,?,1,?)", key.ID, key.Name, digest(secret), key.Hint, key.Created); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if configChanged(transaction) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		server.configurationChanged()
		reply(writer, 201, map[string]any{"key": secret, "metadata": key})
		return
	}
	if request.Method == http.MethodPatch && strings.HasPrefix(request.URL.Path, "/admin/api/keys/") {
		var input struct {
			Enabled *bool `json:"enabled"`
		}
		if jsonInput(writer, request, &input) != nil || input.Enabled == nil {
			failure(writer, 400, "invalid_request", "")
			return
		}
		transaction, err := server.Store.DB.Begin()
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		defer transaction.Rollback()
		result, err := transaction.Exec("UPDATE app_keys SET enabled=? WHERE id=?", *input.Enabled, strings.TrimPrefix(request.URL.Path, "/admin/api/keys/"))
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			failure(writer, 404, "not_found", "")
			return
		}
		if configChanged(transaction) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		server.configurationChanged()
		reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	failure(writer, 405, "method_not_allowed", "")
}
func (server *Server) records(writer http.ResponseWriter, request *http.Request) {
	limit := 50
	if raw := request.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			failure(writer, 400, "invalid_cursor", "")
			return
		}
		limit = value
	}
	cursor := int64(1 << 62)
	if raw := request.URL.Query().Get("cursor"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 1 {
			failure(writer, 400, "invalid_cursor", "")
			return
		}
		cursor = value
	}
	rows, err := server.Store.DB.Query("SELECT rowid,"+recordColumns+" FROM requests WHERE rowid<? ORDER BY rowid DESC LIMIT ?", cursor, limit+1)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	defer rows.Close()
	list := []Record{}
	var next any
	var lastRow int64
	for rows.Next() {
		var value Record
		var rowID int64
		if err = rows.Scan(&rowID, &value.ID, &value.Time, &value.KeyID, &value.Alias, &value.Model, &value.Stream, &value.Revision, &value.State, &value.Code, &value.Status, &value.Duration, &value.Prompt, &value.Completion, &value.Total); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if len(list) == limit {
			next = strconv.FormatInt(lastRow, 10)
			break
		}
		list = append(list, value)
		lastRow = rowID
	}
	if rows.Err() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"data": list, "next_cursor": next})
}
func (server *Server) appID(request *http.Request) (string, bool) {
	header := request.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer mr_") {
		return "", false
	}
	var id string
	var saved []byte
	err := server.Store.DB.QueryRow("SELECT id,digest FROM app_keys WHERE digest=? AND enabled=1", digest(strings.TrimPrefix(header, "Bearer "))).Scan(&id, &saved)
	return id, err == nil && subtle.ConstantTimeCompare(saved, digest(strings.TrimPrefix(header, "Bearer "))) == 1
}
func (server *Server) listModels(writer http.ResponseWriter, request *http.Request) {
	server.mu.RLock()
	defer server.mu.RUnlock()
	if _, ok := server.appID(request); !ok {
		failure(writer, 401, "unauthorized", "")
		return
	}
	rows, err := server.Store.DB.Query("SELECT alias FROM models WHERE enabled=1 AND EXISTS(SELECT 1 FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id WHERE c.alias=models.alias AND c.enabled=1 AND u.enabled=1 AND length(u.token)>0) ORDER BY alias LIMIT 1000")
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	defer rows.Close()
	list := []map[string]string{}
	for rows.Next() {
		var alias string
		if rows.Scan(&alias) != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		list = append(list, map[string]string{"id": alias, "object": "model", "owned_by": "model-relay"})
	}
	if rows.Err() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"object": "list", "data": list})
}
