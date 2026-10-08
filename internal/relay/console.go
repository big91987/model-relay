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
	server.routing = make(map[string]*routeRuntime)
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
		if err = server.Store.DB.QueryRow("SELECT evidence_version,coalesce(config_revision,0),coalesce(model_revision,0),successful_upstream_id,committed_at,attempts_complete FROM requests WHERE id=?", id).Scan(&value.EvidenceVersion, &value.ConfigRevision, &value.ModelRevision, &value.SuccessfulUpstream, &value.CommittedAt, &value.AttemptsComplete); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if value.EvidenceVersion == 2 {
			value.Attempts, err = readAttempts(server.Store.DB, id)
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			rows, err := server.Store.DB.Query("SELECT selection_round,upstream_id,snapshot_enabled,snapshot_priority,snapshot_weight,reason,observed_at,cooldown_until FROM request_route_decisions WHERE request_id=? ORDER BY selection_round,upstream_id", id)
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			value.Decisions = []RouteDecision{}
			for rows.Next() {
				var decision RouteDecision
				if rows.Scan(&decision.Round, &decision.UpstreamID, &decision.Enabled, &decision.Priority, &decision.Weight, &decision.Reason, &decision.Observed, &decision.Cooldown) != nil {
					rows.Close()
					failure(writer, 503, "storage_unavailable", "")
					return
				}
				value.Decisions = append(value.Decisions, decision)
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			value.UsageSummary = usageParts(value.Attempts, value.AttemptsComplete)
		}
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
	var upstreamTotal, upstreamEnabled, ready int64
	err = transaction.QueryRowContext(request.Context(), "SELECT count(*),coalesce(sum(enabled),0),coalesce(max(length(token)>0),0) FROM upstream").Scan(&upstreamTotal, &upstreamEnabled, &credential)
	configured := upstreamTotal > 0
	enabled = upstreamEnabled > 0
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if transaction.QueryRow("SELECT config_revision FROM meta").Scan(&revision) != nil || transaction.QueryRow("SELECT count(*) FROM models WHERE enabled=1 AND EXISTS(SELECT 1 FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id WHERE c.alias=models.alias AND c.enabled=1 AND u.enabled=1 AND length(u.token)>0)").Scan(&ready) != nil {
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
			queryErr = transaction.QueryRowContext(request.Context(), "SELECT EXISTS(SELECT 1 FROM requests r JOIN models m ON m.alias=r.alias JOIN app_keys k ON k.id=r.key_id JOIN route_candidates c ON c.alias=m.alias AND c.upstream_id=r.successful_upstream_id JOIN upstream u ON u.id=c.upstream_id WHERE r.id=? AND m.enabled=1 AND k.enabled=1 AND c.enabled=1 AND u.enabled=1 AND r.config_revision=? AND c.remote_model=r.model AND u.revision=r.revision)", value.ID, revision).Scan(&matches)
		}
		if queryErr != nil && queryErr != sql.ErrNoRows {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if queryErr == nil && value.State == "success" && enabled && matches {
			evidence = map[string]any{"request_id": value.ID, "time": value.Time, "alias": value.Alias, "model": value.Model, "key_id": value.KeyID, "upstream_revision": value.Revision, "scope": "本次服务启动、当前配置的一次已记录成功"}
		} else {
			server.successID = ""
		}
	}
	if transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"upstream": map[string]any{"configured": configured, "enabled": enabled, "credential_configured": credential, "revision": revision, "total": upstreamTotal, "enabled_count": upstreamEnabled}, "models": models, "keys": keys, "callable_models": ready, "call_evidence": evidence, "degraded": server.degraded.Load()})
}

func (server *Server) identifyKey(request *http.Request) (string, bool) {
	header := request.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer mr_") {
		return "", false
	}
	var id string
	var enabled bool
	if server.Store.Schema == 3 {
		if server.Store.DB.QueryRow("SELECT k.id,k.enabled FROM key_generations g JOIN app_keys k ON k.id=g.key_id WHERE g.digest=?", digest(strings.TrimPrefix(header, "Bearer "))).Scan(&id, &enabled) != nil {
			return "", false
		}
		_, authorized := server.controlledAppID(request)
		return id, authorized
	}
	if server.Store.DB.QueryRow("SELECT id,enabled FROM app_keys WHERE digest=?", digest(strings.TrimPrefix(header, "Bearer "))).Scan(&id, &enabled) != nil {
		return "", false
	}
	return id, enabled
}
