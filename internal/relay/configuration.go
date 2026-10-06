package relay

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Upstream struct {
	ID                   int64        `json:"id"`
	Name                 string       `json:"name"`
	URL                  string       `json:"url"`
	Enabled              bool         `json:"enabled"`
	Revision             int64        `json:"revision"`
	ExecutionRevision    int64        `json:"execution_revision"`
	CredentialConfigured bool         `json:"credential_configured"`
	Check                *CheckResult `json:"check"`
	Encrypted            []byte       `json:"-"`
}

type CheckResult struct {
	ID                int64   `json:"id"`
	ExecutionRevision int64   `json:"execution_revision"`
	State             string  `json:"state"`
	Started           string  `json:"started_at"`
	Finished          *string `json:"finished_at"`
	Duration          *int64  `json:"duration_ms"`
	Code              string  `json:"error_code"`
	Status            *int    `json:"upstream_status"`
	Applicable        bool    `json:"applicable"`
}

func (store *Store) upstreamByID(id int64) (Upstream, error) {
	var value Upstream
	err := store.DB.QueryRow("SELECT id,name,url,enabled,revision,execution_revision,token FROM upstream WHERE id=?", id).Scan(&value.ID, &value.Name, &value.URL, &value.Enabled, &value.Revision, &value.ExecutionRevision, &value.Encrypted)
	value.CredentialConfigured = len(value.Encrypted) > 0
	if err != nil {
		return value, err
	}
	var check CheckResult
	err = store.DB.QueryRow("SELECT id,execution_revision,state,started_at,finished_at,duration_ms,error_code,upstream_status FROM upstream_checks WHERE upstream_id=? ORDER BY id DESC LIMIT 1", id).Scan(&check.ID, &check.ExecutionRevision, &check.State, &check.Started, &check.Finished, &check.Duration, &check.Code, &check.Status)
	if err == nil {
		check.Applicable = check.ExecutionRevision == value.ExecutionRevision
		value.Check = &check
	}
	if err == sql.ErrNoRows {
		err = nil
	}
	return value, err
}

func pageLimit(request *http.Request) (int, bool) {
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

func configChanged(transaction *sql.Tx) error {
	_, err := transaction.Exec("UPDATE meta SET config_revision=config_revision+1")
	return err
}

func configurationConflict(writer http.ResponseWriter, revision int64) {
	reply(writer, 409, map[string]any{"error": map[string]string{"type": "relay_error", "code": "configuration_conflict", "message": "Configuration changed; reload before editing"}, "current_revision": revision})
}

func (server *Server) upstreams(writer http.ResponseWriter, request *http.Request) {
	path := strings.TrimPrefix(request.URL.Path, "/admin/api/upstreams")
	if path == "" && request.Method == http.MethodGet {
		limit, ok := pageLimit(request)
		cursor := int64(0)
		if raw := request.URL.Query().Get("cursor"); raw != "" {
			var err error
			cursor, err = strconv.ParseInt(raw, 10, 64)
			ok = ok && err == nil && cursor > 0
		}
		if !ok {
			failure(writer, 400, "invalid_cursor", "")
			return
		}
		server.mu.RLock()
		defer server.mu.RUnlock()
		rows, err := server.Store.DB.Query("SELECT id FROM upstream WHERE id>? ORDER BY id LIMIT ?", cursor, limit+1)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		ids := []int64{}
		for rows.Next() {
			var id int64
			if rows.Scan(&id) != nil {
				rows.Close()
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			ids = append(ids, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		var next any
		if len(ids) > limit {
			ids = ids[:limit]
			next = strconv.FormatInt(ids[len(ids)-1], 10)
		}
		list := []Upstream{}
		for _, id := range ids {
			value, err := server.Store.upstreamByID(id)
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			list = append(list, value)
		}
		reply(writer, 200, map[string]any{"data": list, "next_cursor": next})
		return
	}
	id := int64(0)
	if path != "" {
		raw := strings.TrimPrefix(strings.TrimSuffix(path, "/check"), "/")
		var err error
		id, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || id < 1 {
			failure(writer, 404, "not_found", "")
			return
		}
	}
	if strings.HasSuffix(path, "/check") {
		server.checkUpstream(writer, request, id)
		return
	}
	if request.Method == http.MethodGet && id > 0 {
		server.mu.RLock()
		defer server.mu.RUnlock()
		value, err := server.Store.upstreamByID(id)
		if err == sql.ErrNoRows {
			failure(writer, 404, "not_found", "")
		} else if err != nil {
			failure(writer, 503, "storage_unavailable", "")
		} else {
			reply(writer, 200, value)
		}
		return
	}
	creating := path == "" && request.Method == http.MethodPost
	if !creating && !(id > 0 && request.Method == http.MethodPatch) {
		failure(writer, 405, "method_not_allowed", "")
		return
	}
	var input struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Token    string `json:"token"`
		Enabled  *bool  `json:"enabled"`
		Expected *int64 `json:"expected_revision"`
	}
	if jsonInput(writer, request, &input) != nil || input.Enabled == nil || !validName(strings.TrimSpace(input.Name)) || len(input.Token) > 8192 || strings.ContainsAny(input.Token, "\r\n\x00") || !creating && input.Expected == nil {
		failure(writer, 400, "invalid_request", "")
		return
	}
	normalized, err := server.Config.Policy.URL(request.Context(), input.URL)
	if err != nil {
		failure(writer, 400, "invalid_destination", "")
		return
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	var current Upstream
	if !creating {
		current, err = server.Store.upstreamByID(id)
		if err == sql.ErrNoRows {
			failure(writer, 404, "not_found", "")
			return
		}
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if *input.Expected != current.Revision {
			configurationConflict(writer, current.Revision)
			return
		}
	}
	if input.Token == "" && (creating || normalized != current.URL || !current.CredentialConfigured) {
		failure(writer, 400, "new_credential_required", "")
		return
	}
	transaction, err := server.Store.DB.Begin()
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	defer transaction.Rollback()
	var duplicate bool
	if transaction.QueryRow("SELECT EXISTS(SELECT 1 FROM upstream WHERE name=? AND id<>?)", strings.TrimSpace(input.Name), id).Scan(&duplicate) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if duplicate {
		failure(writer, 409, "name_conflict", "")
		return
	}
	revision := current.Revision + 1
	execution := current.ExecutionRevision
	if creating {
		result, err := transaction.Exec("INSERT INTO upstream(name,url,token,enabled,revision,execution_revision) VALUES(?,?,?, ?,1,1)", strings.TrimSpace(input.Name), normalized, []byte{}, *input.Enabled)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		id, err = result.LastInsertId()
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		execution = 1
	} else if normalized != current.URL || input.Token != "" {
		execution++
	}
	encrypted := current.Encrypted
	if input.Token != "" {
		encrypted, err = seal(server.Store.Key, input.Token, upstreamAAD(id, normalized))
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
	}
	if _, err = transaction.Exec("UPDATE upstream SET name=?,url=?,token=?,enabled=?,revision=?,execution_revision=? WHERE id=?", strings.TrimSpace(input.Name), normalized, encrypted, *input.Enabled, revision, execution, id); err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if configChanged(transaction) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	server.configurationChanged()
	value, err := server.Store.upstreamByID(id)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	status := 200
	if creating {
		status = 201
	}
	reply(writer, status, value)
}

type Candidate struct {
	UpstreamID        int64   `json:"upstream_id"`
	RemoteModel       string  `json:"remote_model"`
	Enabled           bool    `json:"enabled"`
	Priority          int     `json:"priority"`
	Weight            int     `json:"weight"`
	ExecutionRevision int64   `json:"execution_revision,omitempty"`
	GenerationHealth  string  `json:"generation_health,omitempty"`
	CooldownUntil     *string `json:"cooldown_until,omitempty"`
}

func (server *Server) describeModel(value *RouteModel) error {
	for index := range value.Candidates {
		candidate := &value.Candidates[index]
		upstream, err := server.Store.upstreamByID(candidate.UpstreamID)
		if err != nil {
			return err
		}
		candidate.GenerationHealth = "未知（本次启动）"
		health := server.health[healthKey(value.Alias, *candidate, upstream)]
		if health == nil {
			continue
		}
		if health.Cooldown.After(time.Now()) {
			candidate.GenerationHealth = "暂时冷却"
			until := health.Cooldown.UTC().Format(time.RFC3339Nano)
			candidate.CooldownUntil = &until
		} else if !health.Cooldown.IsZero() {
			candidate.GenerationHealth = "待重新尝试"
		} else if health.State == "success" {
			candidate.GenerationHealth = "最近生成成功（非持续健康）"
		} else {
			candidate.GenerationHealth = "最近生成失败（仍参与）"
		}
	}
	return nil
}

type RouteModel struct {
	Alias      string      `json:"alias"`
	Enabled    bool        `json:"enabled"`
	Revision   int64       `json:"revision"`
	Legacy     *string     `json:"legacy_remote_model"`
	Candidates []Candidate `json:"candidates"`
}

func (store *Store) routeModel(alias string) (RouteModel, error) {
	value := RouteModel{Candidates: []Candidate{}}
	err := store.DB.QueryRow("SELECT alias,enabled,revision,legacy_remote_model FROM models WHERE alias=?", alias).Scan(&value.Alias, &value.Enabled, &value.Revision, &value.Legacy)
	if err != nil {
		return value, err
	}
	rows, err := store.DB.Query("SELECT upstream_id,remote_model,enabled,priority,weight,execution_revision FROM route_candidates WHERE alias=? ORDER BY upstream_id", alias)
	if err != nil {
		return value, err
	}
	defer rows.Close()
	for rows.Next() {
		var candidate Candidate
		if err = rows.Scan(&candidate.UpstreamID, &candidate.RemoteModel, &candidate.Enabled, &candidate.Priority, &candidate.Weight, &candidate.ExecutionRevision); err != nil {
			return value, err
		}
		value.Candidates = append(value.Candidates, candidate)
	}
	return value, rows.Err()
}

func (server *Server) routeModels(writer http.ResponseWriter, request *http.Request) {
	alias := strings.TrimPrefix(request.URL.Path, "/admin/api/models/")
	collection := request.URL.Path == "/admin/api/models"
	if !collection && !validAlias(alias) {
		failure(writer, 404, "not_found", "")
		return
	}
	if request.Method == http.MethodGet {
		server.mu.RLock()
		defer server.mu.RUnlock()
		if !collection {
			value, err := server.Store.routeModel(alias)
			if err == nil {
				err = server.describeModel(&value)
			}
			if err == sql.ErrNoRows {
				failure(writer, 404, "not_found", "")
			} else if err != nil {
				failure(writer, 503, "storage_unavailable", "")
			} else {
				reply(writer, 200, value)
			}
			return
		}
		limit, ok := pageLimit(request)
		cursor := request.URL.Query().Get("cursor")
		if !ok || cursor != "" && !validAlias(cursor) {
			failure(writer, 400, "invalid_cursor", "")
			return
		}
		rows, err := server.Store.DB.Query("SELECT alias FROM models WHERE alias>? ORDER BY alias LIMIT ?", cursor, limit+1)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		aliases := []string{}
		for rows.Next() {
			var name string
			if rows.Scan(&name) != nil {
				rows.Close()
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			aliases = append(aliases, name)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		var next any
		if len(aliases) > limit {
			aliases = aliases[:limit]
			next = aliases[len(aliases)-1]
		}
		list := []RouteModel{}
		for _, name := range aliases {
			value, err := server.Store.routeModel(name)
			if err == nil {
				err = server.describeModel(&value)
			}
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			list = append(list, value)
		}
		reply(writer, 200, map[string]any{"data": list, "next_cursor": next})
		return
	}
	creating := collection && request.Method == http.MethodPost
	if !creating && !(request.Method == http.MethodPatch && !collection) {
		failure(writer, 405, "method_not_allowed", "")
		return
	}
	var input struct {
		Alias      string `json:"alias"`
		Enabled    *bool  `json:"enabled"`
		Expected   *int64 `json:"expected_revision"`
		Candidates *[]struct {
			UpstreamID  int64  `json:"upstream_id"`
			RemoteModel string `json:"remote_model"`
			Enabled     *bool  `json:"enabled"`
			Priority    *int   `json:"priority"`
			Weight      *int   `json:"weight"`
		} `json:"candidates"`
	}
	if jsonInput(writer, request, &input) != nil || input.Enabled == nil || input.Candidates == nil || len(*input.Candidates) > 100 || !creating && (input.Expected == nil || input.Alias != "") {
		failure(writer, 400, "invalid_request", "")
		return
	}
	if creating {
		alias = input.Alias
	}
	if !validAlias(alias) {
		failure(writer, 400, "invalid_model", "")
		return
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	current, err := server.Store.routeModel(alias)
	if creating && err == nil {
		failure(writer, 409, "alias_conflict", "")
		return
	}
	if !creating && err == sql.ErrNoRows {
		failure(writer, 404, "not_found", "")
		return
	}
	if err != nil && err != sql.ErrNoRows {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if !creating && current.Revision != *input.Expected {
		configurationConflict(writer, current.Revision)
		return
	}
	seen := map[int64]bool{}
	for _, candidate := range *input.Candidates {
		if candidate.UpstreamID < 1 || seen[candidate.UpstreamID] || !validName(candidate.RemoteModel) || candidate.Enabled == nil || candidate.Priority == nil || candidate.Weight == nil || *candidate.Priority < 0 || *candidate.Priority > 100 || *candidate.Weight < 1 || *candidate.Weight > 100 {
			failure(writer, 400, "invalid_route_candidate", "")
			return
		}
		seen[candidate.UpstreamID] = true
		var exists bool
		if server.Store.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM upstream WHERE id=?)", candidate.UpstreamID).Scan(&exists) != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if !exists {
			failure(writer, 400, "invalid_route_candidate", "")
			return
		}
	}
	transaction, err := server.Store.DB.Begin()
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	defer transaction.Rollback()
	if creating {
		_, err = transaction.Exec("INSERT INTO models(alias,enabled,revision) VALUES(?,?,1)", alias, *input.Enabled)
	} else {
		_, err = transaction.Exec("UPDATE models SET enabled=?,revision=revision+1,legacy_remote_model=CASE WHEN ? THEN NULL ELSE legacy_remote_model END WHERE alias=?", *input.Enabled, len(*input.Candidates) > 0, alias)
	}
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if _, err = transaction.Exec("DELETE FROM route_candidates WHERE alias=?", alias); err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	for _, candidate := range *input.Candidates {
		execution := int64(1)
		for _, previous := range current.Candidates {
			if previous.UpstreamID == candidate.UpstreamID {
				execution = previous.ExecutionRevision
				if previous.RemoteModel != candidate.RemoteModel {
					execution++
				}
			}
		}
		if _, err = transaction.Exec("INSERT INTO route_candidates VALUES(?,?,?,?,?,?,?)", alias, candidate.UpstreamID, candidate.RemoteModel, *candidate.Enabled, *candidate.Priority, *candidate.Weight, execution); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
	}
	if configChanged(transaction) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	for _, candidate := range current.Candidates {
		for _, replacement := range *input.Candidates {
			if candidate.UpstreamID == replacement.UpstreamID && candidate.RemoteModel != replacement.RemoteModel {
				prefix := fmt.Sprintf("%s:%d:", alias, candidate.UpstreamID)
				for key := range server.health {
					if strings.HasPrefix(key, prefix) {
						delete(server.health, key)
					}
				}
			}
		}
	}
	server.configurationChanged()
	value, err := server.Store.routeModel(alias)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	status := 200
	if creating {
		status = 201
	}
	reply(writer, status, value)
}

func (server *Server) checkUpstream(writer http.ResponseWriter, request *http.Request, id int64) {
	if request.Method != http.MethodPost {
		failure(writer, 405, "method_not_allowed", "")
		return
	}
	var input struct {
		Expected *int64 `json:"expected_execution_revision"`
	}
	if jsonInput(writer, request, &input) != nil || input.Expected == nil {
		failure(writer, 400, "invalid_request", "")
		return
	}
	server.mu.Lock()
	value, err := server.Store.upstreamByID(id)
	if err != nil || value.ExecutionRevision != *input.Expected || server.checking[id] || !value.Enabled || !value.CredentialConfigured {
		checking := server.checking[id]
		server.mu.Unlock()
		switch {
		case err == sql.ErrNoRows:
			failure(writer, 404, "not_found", "")
		case err != nil:
			failure(writer, 503, "storage_unavailable", "")
		case value.ExecutionRevision != *input.Expected:
			failure(writer, 409, "configuration_conflict", "")
		case checking:
			failure(writer, 409, "check_in_progress", "")
		default:
			failure(writer, 400, "upstream_unavailable", "")
		}
		return
	}
	started := time.Now()
	check := CheckResult{ExecutionRevision: value.ExecutionRevision, State: "checking", Started: started.UTC().Format(time.RFC3339Nano)}
	result, err := server.Store.DB.Exec("INSERT INTO upstream_checks(upstream_id,execution_revision,state,started_at,error_code) VALUES(?,?,'checking',?,'')", id, value.ExecutionRevision, check.Started)
	if err != nil {
		server.mu.Unlock()
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	check.ID, err = result.LastInsertId()
	if err != nil {
		server.mu.Unlock()
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	server.checking[id] = true
	server.mu.Unlock()
	defer func() { server.mu.Lock(); delete(server.checking, id); server.mu.Unlock() }()
	ctx, cancel := context.WithTimeout(request.Context(), 10*time.Second)
	defer cancel()
	check.State = "failed"
	check.Code = "upstream_connection_failed"
	token, err := unseal(server.Store.Key, value.Encrypted, upstreamAAD(id, value.URL))
	if err == nil {
		outgoing, err := http.NewRequestWithContext(ctx, http.MethodGet, value.URL+"/models", nil)
		if err == nil {
			outgoing.Header.Set("Authorization", "Bearer "+token)
			response, err := server.client.Do(outgoing)
			if err == nil {
				check.Status = &response.StatusCode
				if response.StatusCode >= 200 && response.StatusCode < 300 {
					data, readErr := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
					var listing struct {
						Object string `json:"object"`
						Data   []struct {
							ID string `json:"id"`
						} `json:"data"`
					}
					mediaType, _, typeErr := mime.ParseMediaType(response.Header.Get("Content-Type"))
					valid := typeErr == nil && mediaType == "application/json" && readErr == nil && len(data) <= 1<<20 && json.Unmarshal(data, &listing) == nil && listing.Object == "list" && listing.Data != nil
					for _, model := range listing.Data {
						valid = valid && strings.TrimSpace(model.ID) != ""
					}
					if valid {
						check.State = "success"
						check.Code = ""
					} else {
						check.Code = "upstream_protocol_error"
					}
				} else if response.StatusCode == 401 || response.StatusCode == 403 {
					check.State = "failed"
					check.Code = "upstream_rejected"
				} else {
					check.Code = "upstream_http_error"
				}
				response.Body.Close()
			}
		}
	}
	if request.Context().Err() != nil {
		check.State = "canceled"
		check.Code = "client_canceled"
	} else if ctx.Err() != nil {
		check.State = "failed"
		check.Code = "upstream_timeout"
	}
	finished := time.Now().UTC().Format(time.RFC3339Nano)
	duration := time.Since(started).Milliseconds()
	check.Finished = &finished
	check.Duration = &duration
	cleanup, release := context.WithTimeout(context.Background(), 2*time.Second)
	defer release()
	if _, err = server.Store.DB.ExecContext(cleanup, "UPDATE upstream_checks SET state=?,finished_at=?,duration_ms=?,error_code=?,upstream_status=? WHERE id=?", check.State, finished, duration, check.Code, check.Status, check.ID); err != nil {
		server.storageFailure()
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	server.mu.RLock()
	var execution int64
	err = server.Store.DB.QueryRow("SELECT execution_revision FROM upstream WHERE id=?", id).Scan(&execution)
	server.mu.RUnlock()
	check.Applicable = err == nil && execution == check.ExecutionRevision
	reply(writer, 200, check)
}
