package relay

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (server *Server) consoleAPI(writer http.ResponseWriter, request *http.Request, current session) bool {
	path := strings.TrimPrefix(request.URL.Path, "/admin/api/")
	workspace := strings.HasPrefix(path, "workspace/")
	path = strings.TrimPrefix(path, "workspace/")
	parts := strings.Split(path, "/")
	if parts[0] == "requests" && !workspace && request.URL.Query().Get("tenant") == "" {
		return false
	}
	known := (workspace && parts[0] == "overview") || parts[0] == "catalog" || parts[0] == "usage" || parts[0] == "audit" || parts[0] == "requests" || parts[0] == "bills" || parts[0] == "operations" || (parts[0] == "models" && len(parts) == 3 && (parts[2] == "prices" || parts[2] == "bounds" || parts[2] == "tenant-grants")) || (parts[0] == "tenants" && len(parts) >= 3 && (parts[2] == "keys" || parts[2] == "limits" || parts[2] == "bills"))
	if !known {
		return false
	}
	principal, err := server.Store.principal(current.UserID)
	if err != nil {
		failure(writer, 401, "unauthorized", "")
		return true
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	if !server.advanceClock(writer) {
		return true
	}
	transaction, err := server.Store.DB.BeginTx(request.Context(), nil)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return true
	}
	defer transaction.Rollback()
	switch parts[0] {
	case "overview":
		server.consoleOverview(writer, request, transaction, principal)
	case "operations":
		if len(parts) != 2 || request.Method != http.MethodGet {
			failure(writer, 404, "not_found", "")
			return true
		}
		var objectID, action, tenantID string
		if transaction.QueryRow("SELECT object_id,action,tenant_id FROM console_operations WHERE actor=? AND operation=?", principal.ID, parts[1]).Scan(&objectID, &action, &tenantID) != nil {
			failure(writer, 404, "not_found", "")
			return true
		}
		if tenantID != "" {
			if _, err := scopedRole(transaction, principal, tenantID); err != nil {
				failure(writer, 404, "not_found", "")
				return true
			}
		} else if !principal.Platform {
			failure(writer, 403, "forbidden", "")
			return true
		}
		reply(writer, 200, map[string]any{"object_id": objectID, "action": action, "secret_recoverable": false})
	case "tenants":
		tenantID := parts[1]
		role, err := scopedRole(transaction, principal, tenantID)
		if err != nil {
			failure(writer, 404, "not_found", "")
			return true
		}
		switch parts[2] {
		case "keys":
			server.controlledKeys(writer, request, transaction, principal, tenantID, parts[3:])
		case "limits":
			server.manageLimits(writer, request, transaction, principal, tenantID, role, "tenant", tenantID)
		case "bills":
			server.tenantBills(writer, request, transaction, principal, tenantID, role)
		}
	case "models":
		server.manageModelFinance(writer, request, transaction, principal, parts[1], parts[2])
	case "catalog":
		server.catalog(writer, request, transaction, principal)
	case "bills":
		server.manageBill(writer, request, transaction, principal, parts[1:])
	case "requests", "usage", "audit":
		server.consoleReports(writer, request, transaction, principal, parts)
	}
	return true
}

func (server *Server) manageLimits(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, tenantID, role, ownerType, ownerID string) {
	if role == "developer" && ownerType == "tenant" {
		failure(writer, 403, "assigned_key_scope_required", "")
		return
	}
	if request.Method == http.MethodGet {
		limits, err := limitsFor(transaction, ownerType, ownerID)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		now := admissionClock(transaction, server.Store.now())
		usage, err := usageFor(transaction, ownerType, ownerID, now.Format("2006-01"), now.Format("2006-01-02T15:04"))
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		var budgetRemaining *string
		if limits.Budget != nil {
			remaining := strconv.FormatInt(*limits.Budget-usage.Fee-usage.Reserved, 10)
			budgetRemaining = &remaining
		}
		var platform *ResourceLimits
		if ownerType == "tenant" {
			var cap ResourceLimits
			err := transaction.QueryRow("SELECT calls,rpm,tpm,concurrency,budget,hard,revision FROM platform_limits WHERE tenant_id=?", tenantID).Scan(&cap.Calls, &cap.RPM, &cap.TPM, &cap.Concurrency, &cap.Budget, &cap.Hard, &cap.Revision)
			if err != nil && err != sql.ErrNoRows {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			if err == nil {
				platform = &cap
			}
		}
		reply(writer, 200, map[string]any{"limits": limits, "usage": usage, "period": now.Format("2006-01"), "currency": "USD", "unknown_carry": true, "budget_remaining_micro_usd": budgetRemaining, "soft_budget_exceeded": limits.Budget != nil && !limits.Hard && usage.Fee+usage.Reserved >= *limits.Budget, "platform_limits": platform})
		return
	}
	if request.Method != http.MethodPut || role != "administrator" {
		failure(writer, 403, "forbidden", "")
		return
	}
	var input ResourceLimits
	if jsonInput(writer, request, &input) != nil {
		failure(writer, 400, "invalid_request", "")
		return
	}
	for _, limit := range []*int64{input.Calls, input.RPM, input.TPM, input.Concurrency, input.Budget} {
		if limit != nil && (*limit < 0 || *limit > 1000000000000) {
			failure(writer, 400, "invalid_limit", "")
			return
		}
	}
	previous, err := limitsFor(transaction, ownerType, ownerID)
	if err != nil || previous.Revision != input.Revision {
		failure(writer, 409, "revision_conflict", "")
		return
	}
	if !principal.Platform && ownerType == "tenant" {
		var cap ResourceLimits
		err := transaction.QueryRow("SELECT calls,rpm,tpm,concurrency,budget,hard,revision FROM platform_limits WHERE tenant_id=?", tenantID).Scan(&cap.Calls, &cap.RPM, &cap.TPM, &cap.Concurrency, &cap.Budget, &cap.Hard, &cap.Revision)
		if err != nil && err != sql.ErrNoRows {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		for index, requested := range []*int64{input.Calls, input.RPM, input.TPM, input.Concurrency, input.Budget} {
			upper := []*int64{cap.Calls, cap.RPM, cap.TPM, cap.Concurrency, cap.Budget}[index]
			if upper != nil && (requested == nil || *requested > *upper) {
				failure(writer, 403, "platform_limit_exceeded", "")
				return
			}
		}
		if cap.Budget != nil && cap.Hard && !input.Hard {
			failure(writer, 403, "platform_limit_exceeded", "")
			return
		}
	}
	if principal.Platform && ownerType == "tenant" {
		if _, err := transaction.Exec("INSERT INTO platform_limits VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(tenant_id) DO UPDATE SET calls=excluded.calls,rpm=excluded.rpm,tpm=excluded.tpm,concurrency=excluded.concurrency,budget=excluded.budget,hard=excluded.hard,revision=excluded.revision", tenantID, input.Calls, input.RPM, input.TPM, input.Concurrency, input.Budget, input.Hard, previous.Revision+1); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
	}
	if _, err := transaction.Exec("INSERT INTO resource_limits VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(owner_type,owner_id) DO UPDATE SET calls=excluded.calls,rpm=excluded.rpm,tpm=excluded.tpm,concurrency=excluded.concurrency,budget=excluded.budget,hard=excluded.hard,revision=excluded.revision", ownerType, ownerID, input.Calls, input.RPM, input.TPM, input.Concurrency, input.Budget, input.Hard, previous.Revision+1); err != nil || auditIdentity(transaction, principal.ID, tenantID, "limits.update", ownerID) != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]bool{"ok": true})
}

func (server *Server) manageModelFinance(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, alias, resource string) {
	if !principal.Platform {
		failure(writer, 403, "forbidden", "")
		return
	}
	var count int
	if transaction.QueryRow("SELECT count(*) FROM models WHERE alias=?", alias).Scan(&count) != nil || count != 1 {
		failure(writer, 404, "not_found", "")
		return
	}
	if resource == "tenant-grants" {
		if request.Method == http.MethodGet {
			values, err := stringList(transaction, "SELECT tenant_id FROM tenant_model_grants WHERE alias=? ORDER BY tenant_id", alias)
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			replyCollection(writer, request, values)
			return
		}
		var input struct {
			Tenants []string `json:"tenants"`
		}
		if request.Method != http.MethodPut || jsonInput(writer, request, &input) != nil || len(input.Tenants) > 1000 {
			failure(writer, 400, "invalid_grants", "")
			return
		}
		seen := map[string]bool{}
		for _, tenant := range input.Tenants {
			var active bool
			if seen[tenant] || transaction.QueryRow("SELECT active FROM tenants WHERE id=?", tenant).Scan(&active) != nil || !active {
				failure(writer, 400, "invalid_grants", "")
				return
			}
			seen[tenant] = true
		}
		if _, err := transaction.Exec("DELETE FROM tenant_model_grants WHERE alias=?", alias); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		for _, tenant := range input.Tenants {
			if _, err := transaction.Exec("INSERT INTO tenant_model_grants VALUES(?,?)", tenant, alias); err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
		}
		if auditIdentity(transaction, principal.ID, "", "model.grants", alias) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	if resource == "bounds" {
		if request.Method == http.MethodGet {
			values, err := queryMaps(transaction, "SELECT b.upstream_id,b.max_input,b.max_output,b.evidence,b.upstream_execution,b.candidate_execution,(b.upstream_execution=u.execution_revision AND b.candidate_execution=c.execution_revision) AS applicable FROM model_bounds b JOIN upstream u ON u.id=b.upstream_id JOIN route_candidates c ON c.alias=b.alias AND c.upstream_id=b.upstream_id WHERE b.alias=? ORDER BY b.upstream_id", alias)
			if err != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			replyCollection(writer, request, values, map[string]any{"evidence_kind": "operator_declared_not_supplier_validation"})
			return
		}
		var input struct {
			Upstream int64  `json:"upstream_id"`
			Input    int64  `json:"max_input"`
			Output   int64  `json:"max_output"`
			Evidence string `json:"evidence"`
		}
		if request.Method != http.MethodPut || jsonInput(writer, request, &input) != nil || input.Input <= 0 || input.Output <= 0 || input.Input > 1000000000 || input.Output > 1000000000 || !validName(input.Evidence) {
			failure(writer, 400, "invalid_bound", "")
			return
		}
		var sourceExecution, candidateExecution int64
		if transaction.QueryRow("SELECT u.execution_revision,c.execution_revision FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id WHERE c.alias=? AND c.upstream_id=?", alias, input.Upstream).Scan(&sourceExecution, &candidateExecution) != nil {
			failure(writer, 400, "route_candidate_required", "")
			return
		}
		if _, err := transaction.Exec("INSERT INTO model_bounds VALUES(?,?,?,?,?,?,?) ON CONFLICT(alias,upstream_id) DO UPDATE SET max_input=excluded.max_input,max_output=excluded.max_output,evidence=excluded.evidence,upstream_execution=excluded.upstream_execution,candidate_execution=excluded.candidate_execution", alias, input.Upstream, input.Input, input.Output, input.Evidence, sourceExecution, candidateExecution); err != nil || auditIdentity(transaction, principal.ID, "", "model.bound", alias) != nil || transaction.Commit() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]bool{"ok": true})
		return
	}
	if request.Method == http.MethodGet {
		rows, err := transaction.Query("SELECT id,upstream_id,effective,cost_input,cost_output,tenant_input,tenant_output FROM prices WHERE alias=? ORDER BY effective DESC,id DESC", alias)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		defer rows.Close()
		values := []map[string]any{}
		for rows.Next() {
			var price PriceSnapshot
			var effective string
			if rows.Scan(&price.ID, &price.Upstream, &effective, &price.CostIn, &price.CostOut, &price.FeeIn, &price.FeeOut) != nil {
				failure(writer, 503, "storage_unavailable", "")
				return
			}
			values = append(values, map[string]any{"price": price, "effective_at": effective, "currency": "USD", "unit": "micro_usd_per_million_tokens"})
		}
		if rows.Err() != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		replyCollection(writer, request, values)
		return
	}
	var input struct {
		Effective string          `json:"effective_at"`
		Prices    []PriceSnapshot `json:"prices"`
	}
	if request.Method != http.MethodPost || jsonInput(writer, request, &input) != nil || len(input.Prices) == 0 || len(input.Prices) > 100 {
		failure(writer, 400, "invalid_prices", "")
		return
	}
	effectiveText := input.Effective
	if effectiveText == "" {
		effectiveText = server.Store.now().Format(time.RFC3339Nano)
	}
	effective, err := time.Parse(time.RFC3339Nano, effectiveText)
	if err != nil {
		failure(writer, 400, "invalid_effective_at", "")
		return
	}
	operation, prior, fingerprint, err := operationStart(transaction, principal.ID, request, "price.publish", input)
	if err != nil {
		failure(writer, 409, err.Error(), "")
		return
	}
	if prior != "" {
		reply(writer, 200, map[string]any{"id": prior, "already_applied": true})
		return
	}
	if effective.Before(server.Store.now().Add(-time.Minute)) {
		failure(writer, 400, "invalid_effective_at", "")
		return
	}
	seen := map[int64]bool{}
	var candidates int
	if transaction.QueryRow("SELECT count(*) FROM route_candidates WHERE alias=?", alias).Scan(&candidates) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if candidates != len(input.Prices) {
		failure(writer, 400, "all_candidate_prices_required", "")
		return
	}
	for _, price := range input.Prices {
		if price.FeeIn != input.Prices[0].FeeIn || price.FeeOut != input.Prices[0].FeeOut {
			failure(writer, 400, "tenant_rates_must_match", "")
			return
		}
		if seen[price.Upstream] {
			failure(writer, 400, "duplicate_price", "")
			return
		}
		seen[price.Upstream] = true
		for _, rate := range []int64{price.CostIn, price.CostOut, price.FeeIn, price.FeeOut} {
			if rate < 0 || rate > 1000000000000 {
				failure(writer, 400, "invalid_rate", "")
				return
			}
		}
		price.ID = randomSecret(16)
		var sourceExecution, candidateExecution int64
		if transaction.QueryRow("SELECT u.execution_revision,c.execution_revision FROM route_candidates c JOIN upstream u ON u.id=c.upstream_id WHERE c.alias=? AND c.upstream_id=?", alias, price.Upstream).Scan(&sourceExecution, &candidateExecution) != nil {
			failure(writer, 400, "route_candidate_required", "")
			return
		}
		if _, err := transaction.Exec("INSERT INTO prices VALUES(?,?,?,?,?,?,?,?,?,?)", price.ID, alias, price.Upstream, effective.UTC().Format(time.RFC3339Nano), price.CostIn, price.CostOut, price.FeeIn, price.FeeOut, sourceExecution, candidateExecution); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
	}
	if operationFinish(transaction, principal.ID, operation, fingerprint, "price.publish", alias, "") != nil || transaction.Commit() != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 201, map[string]any{"id": alias, "currency": "USD", "unit": "micro_usd_per_million_tokens"})
}

func (server *Server) catalog(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal) {
	if request.Method != http.MethodGet {
		failure(writer, 405, "method_not_allowed", "")
		return
	}
	tenantID := request.URL.Query().Get("tenant")
	if tenantID == "" && !principal.Platform {
		failure(writer, 400, "tenant_required", "")
		return
	}
	if tenantID != "" {
		if _, err := scopedRole(transaction, principal, tenantID); err != nil {
			failure(writer, 404, "not_found", "")
			return
		}
	}
	query := "SELECT alias FROM models ORDER BY alias"
	args := []any{}
	includeUngrant := request.URL.Query().Get("include_ungranted") == "true"
	if includeUngrant && !principal.Platform {
		failure(writer, 403, "forbidden", "")
		return
	}
	if tenantID != "" && !includeUngrant {
		query = "SELECT m.alias FROM models m JOIN tenant_model_grants g ON g.alias=m.alias WHERE g.tenant_id=? ORDER BY m.alias"
		args = append(args, tenantID)
	}
	ids, err := stringList(transaction, query, args...)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	values := []map[string]any{}
	for _, alias := range ids {
		var enabled bool
		var revision int64
		if transaction.QueryRow("SELECT enabled,revision FROM models WHERE alias=?", alias).Scan(&enabled, &revision) != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		var tenantInput, tenantOutput *int64
		var effective *string
		err := transaction.QueryRow("SELECT p.tenant_input,p.tenant_output,p.effective FROM prices p JOIN upstream u ON u.id=p.upstream_id JOIN route_candidates c ON c.alias=p.alias AND c.upstream_id=p.upstream_id WHERE p.alias=? AND p.upstream_execution=u.execution_revision AND p.candidate_execution=c.execution_revision AND julianday(p.effective)<=julianday(?) ORDER BY julianday(p.effective) DESC,p.id DESC LIMIT 1", alias, server.Store.now().Format(time.RFC3339Nano)).Scan(&tenantInput, &tenantOutput, &effective)
		if err != nil && err != sql.ErrNoRows {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		summary, err := modelSummary(transaction, principal, tenantID, alias, server.Store.now())
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		values = append(values, map[string]any{"alias": alias, "enabled": enabled, "revision": revision, "tenant_input_rate": tenantInput, "tenant_output_rate": tenantOutput, "effective_at": effective, "currency": "USD", "unit": "micro_usd_per_million_tokens", "protocol": "chat.completions", "stream": true, "summary": summary})
	}
	replyCollection(writer, request, values)
}

func marshalSafe(value any) json.RawMessage { data, _ := json.Marshal(value); return data }
