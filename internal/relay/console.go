package relay

import (
	"database/sql"
	"net/http"
	"strings"
)

const recordColumns = "id,time,coalesce(key_id,''),coalesce(alias,''),coalesce(model,''),stream,coalesce(revision,0),state,error_code,upstream_status,duration_ms,prompt_tokens,completion_tokens,total_tokens"

type rowScanner interface {
	Scan(...any) error
}

func scanRecord(row rowScanner) (Record, error) {
	var value Record
	err := row.Scan(&value.ID, &value.Time, &value.KeyID, &value.Alias, &value.Model, &value.Stream, &value.Revision, &value.State, &value.Code, &value.Status, &value.Duration, &value.Prompt, &value.Completion, &value.Total)
	return value, err
}

func (server *Server) configurationChanged() {
	server.configEpoch++
	server.successID = ""
}

func (server *Server) record(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", "GET")
		failure(writer, 405, "method_not_allowed", "")
		return
	}
	id := strings.TrimPrefix(request.URL.Path, "/admin/api/requests/")
	if len(id) != 22 || !validAlias(id) {
		failure(writer, 404, "not_found", "")
		return
	}
	value, err := scanRecord(server.Store.DB.QueryRowContext(request.Context(), "SELECT "+recordColumns+" FROM requests WHERE id=?", id))
	if err == sql.ErrNoRows {
		failure(writer, 404, "not_found", "")
	} else if err != nil {
		failure(writer, 503, "storage_unavailable", "")
	} else {
		reply(writer, 200, value)
	}
}

func (server *Server) overview(writer http.ResponseWriter, request *http.Request) {
	transaction, err := server.Store.DB.BeginTx(request.Context(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	defer transaction.Rollback()
	var enabled, credential bool
	var revision int64
	err = transaction.QueryRowContext(request.Context(), "SELECT enabled,length(token)>0,revision FROM upstream WHERE id=1").Scan(&enabled, &credential, &revision)
	configured := err == nil
	if err != nil && err != sql.ErrNoRows {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	type counts struct {
		Total   int64 `json:"total"`
		Enabled int64 `json:"enabled"`
	}
	var models, keys counts
	for _, query := range []struct {
		table string
		value *counts
	}{{"models", &models}, {"app_keys", &keys}} {
		if transaction.QueryRowContext(request.Context(), "SELECT count(*),coalesce(sum(enabled),0) FROM "+query.table).Scan(&query.value.Total, &query.value.Enabled) != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
	}
	var evidence any
	if server.successID != "" {
		value, queryErr := scanRecord(transaction.QueryRowContext(request.Context(), "SELECT "+recordColumns+" FROM requests WHERE id=?", server.successID))
		var matches bool
		if queryErr == nil {
			queryErr = transaction.QueryRowContext(request.Context(), "SELECT EXISTS(SELECT 1 FROM models JOIN app_keys ON app_keys.id=? WHERE models.alias=? AND models.model=? AND models.enabled=1 AND app_keys.enabled=1)", value.KeyID, value.Alias, value.Model).Scan(&matches)
		}
		if queryErr != nil && queryErr != sql.ErrNoRows {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if queryErr == nil && value.State == "success" && enabled && matches && value.Revision == revision {
			evidence = map[string]any{"request_id": value.ID, "time": value.Time, "alias": value.Alias, "model": value.Model, "key_id": value.KeyID, "upstream_revision": value.Revision, "scope": "本次服务启动、当前配置的一次已记录成功"}
		} else {
			server.successID = ""
		}
	}
	if transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"upstream": map[string]any{"configured": configured, "enabled": enabled, "credential_configured": credential, "revision": revision}, "models": models, "keys": keys, "call_evidence": evidence, "degraded": server.degraded.Load()})
}

func (server *Server) identifyKey(request *http.Request) (string, bool) {
	header := request.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer mr_") {
		return "", false
	}
	var id string
	var enabled bool
	if server.Store.DB.QueryRow("SELECT id,enabled FROM app_keys WHERE digest=?", digest(strings.TrimPrefix(header, "Bearer "))).Scan(&id, &enabled) != nil {
		return "", false
	}
	return id, enabled
}
