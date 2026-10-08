package relay

import (
	"database/sql"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func replyCollection[Value any](writer http.ResponseWriter, request *http.Request, values []Value, fields ...map[string]any) {
	page, size := 1, 25
	for _, parameter := range []struct {
		name    string
		target  *int
		maximum int
	}{{"page", &page, 100000}, {"page_size", &size, 100}} {
		if raw := request.URL.Query().Get(parameter.name); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value < 1 || value > parameter.maximum {
				failure(writer, 400, "invalid_pagination", "")
				return
			}
			*parameter.target = value
		}
	}
	start := min((page-1)*size, len(values))
	end := min(start+size, len(values))
	result := map[string]any{"items": values[start:end], "total": len(values), "page": page, "page_size": size}
	for _, extra := range fields {
		for name, value := range extra {
			result[name] = value
		}
	}
	reply(writer, 200, result)
}

func reportScope(transaction *sql.Tx, principal Principal, tenantID string) (string, []any, int) {
	if tenantID == "" && principal.Platform {
		return "1=1", nil, 200
	}
	if tenantID == "" {
		return "", nil, 400
	}
	role, err := scopedRole(transaction, principal, tenantID)
	if err != nil {
		return "", nil, 404
	}
	where := "c.tenant_id=?"
	args := []any{tenantID}
	if role == "developer" {
		where += " AND EXISTS(SELECT 1 FROM key_diagnostic_members d WHERE d.key_id=r.key_id AND d.tenant_id=c.tenant_id AND d.user_id=?)"
		args = append(args, principal.ID)
	}
	return where, args, 200
}

func queryMaps(transaction *sql.Tx, query string, args ...any) ([]map[string]any, error) {
	rows, err := transaction.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	values := []map[string]any{}
	for rows.Next() {
		fields := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for index := range fields {
			pointers[index] = &fields[index]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		value := map[string]any{}
		for index, name := range columns {
			if amount, ok := fields[index].(int64); ok && (name == "cost" || name == "fee" || name == "amount" || name == "original_cost" || name == "original_fee" || name == "fee_reserved" || name == "cost_input" || name == "cost_output" || name == "tenant_input" || name == "tenant_output" || strings.HasSuffix(name, "_micro_usd")) {
				value[name] = strconv.FormatInt(amount, 10)
				continue
			}
			if data, ok := fields[index].([]byte); ok {
				value[name] = string(data)
			} else {
				value[name] = fields[index]
			}
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (server *Server) consoleReports(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, parts []string) {
	if len(parts) == 3 && parts[0] == "requests" && parts[2] == "legacy-reconciliations" && request.Method == http.MethodPost {
		server.reconcileLegacy(writer, request, transaction, principal, parts[1])
		return
	}
	if len(parts) == 4 && parts[0] == "requests" && parts[2] == "reconciliations" && request.Method == http.MethodPost {
		server.reconcileUsage(writer, request, transaction, principal, parts[1], parts[3])
		return
	}
	if request.Method != http.MethodGet {
		failure(writer, 405, "method_not_allowed", "")
		return
	}
	tenantID := request.URL.Query().Get("tenant")
	where, args, status := reportScope(transaction, principal, tenantID)
	if status != 200 {
		failure(writer, status, "scope_unavailable", "")
		return
	}
	if parts[0] == "audit" {
		if tenantID != "" {
			role, _ := scopedRole(transaction, principal, tenantID)
			if role != "administrator" {
				failure(writer, 403, "forbidden", "")
				return
			}
		}
		query := "SELECT id,time,actor,tenant_id,action,object_id FROM console_audit"
		arguments := []any{}
		if tenantID != "" {
			query += " WHERE tenant_id=?"
			arguments = append(arguments, tenantID)
		}
		query += " ORDER BY time DESC,id DESC LIMIT 100"
		values, err := queryMaps(transaction, query, arguments...)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]any{"items": values, "page_size": 100, "more_requires_filter": len(values) == 100})
		return
	}
	for _, filter := range []struct{ parameter, column string }{{"model", "r.alias"}, {"key", "r.key_id"}, {"state", "r.state"}, {"request_id", "r.id"}} {
		if value := request.URL.Query().Get(filter.parameter); value != "" {
			where += " AND " + filter.column + "=?"
			args = append(args, value)
		}
	}
	for _, filter := range []struct{ parameter, operator string }{{"from", ">="}, {"to", "<="}} {
		if value := request.URL.Query().Get(filter.parameter); value != "" {
			if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
				failure(writer, 400, "invalid_time_filter", "")
				return
			}
			where += " AND julianday(r.time)" + filter.operator + "julianday(?)"
			args = append(args, value)
		}
	}
	base := " FROM requests r LEFT JOIN key_controls c ON c.key_id=r.key_id LEFT JOIN admissions a ON a.request_id=r.id WHERE " + where
	if len(parts) == 2 && parts[1] != "export" && parts[0] == "requests" {
		base += " AND r.id=?"
		args = append(args, parts[1])
		values, err := queryMaps(transaction, "SELECT r.id,r.time,r.key_id,c.tenant_id,r.alias,r.model,r.state,r.error_code,r.stream,r.upstream_status,r.duration_ms,r.prompt_tokens,r.completion_tokens,r.total_tokens,a.period,a.pending,a.tokens_reserved,a.fee_reserved,a.revision,(SELECT count(*) FROM legacy_pending l WHERE l.request_id=r.id AND l.resolved=0) AS legacy_pending"+base, args...)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		if len(values) != 1 {
			failure(writer, 404, "not_found", "")
			return
		}
		attempts, err := queryMaps(transaction, "SELECT a.ordinal,a.upstream_id,a.upstream_name,a.remote_model,a.started_at,a.finished_at,a.state,a.error_code,a.upstream_status,a.prompt_tokens,a.completion_tokens,a.total_tokens,d.possible AS dispatch_possible FROM request_attempts a LEFT JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal WHERE a.request_id=? ORDER BY a.ordinal", parts[1])
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		prices, err := queryMaps(transaction, "SELECT upstream_id,price_id,cost_input,cost_output,tenant_input,tenant_output,max_input,max_output FROM request_prices WHERE request_id=? ORDER BY upstream_id", parts[1])
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		entries, err := queryMaps(transaction, "SELECT ordinal,revision,period,input_tokens,output_tokens,cost,fee,reason,evidence FROM charge_entries WHERE request_id=? ORDER BY ordinal,revision", parts[1])
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		values[0]["attempts"], values[0]["price_snapshots"], values[0]["charge_entries"] = attempts, prices, entries
		if err := addConfirmedTokens(transaction, values); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		history, err := queryMaps(transaction, "SELECT upstream_id,price_reference,cost_input,cost_output,tenant_input,tenant_output,reason,evidence FROM historical_price_evidence WHERE request_id=? ORDER BY upstream_id", parts[1])
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		values[0]["historical_price_evidence"] = history
		reply(writer, 200, values[0])
		return
	}
	if parts[0] == "usage" && len(parts) == 1 {
		ledgerSum := func(field string) string {
			return "(SELECT coalesce(sum(e." + field + "),0) FROM charge_entries e JOIN requests er ON er.id=e.request_id LEFT JOIN key_controls ec ON ec.key_id=er.key_id WHERE er.alias=r.alias AND " + strings.ReplaceAll(strings.ReplaceAll(where, "c.", "ec."), "r.", "er.") + ")"
		}
		arguments := []any{}
		for index := 0; index < 5; index++ {
			arguments = append(arguments, args...)
		}
		values, err := queryMaps(transaction, "SELECT r.alias,count(*) AS requests,coalesce(sum(a.pending),0) AS pending,coalesce(sum(a.fee_reserved),0) AS reserved_micro_usd,"+ledgerSum("fee")+" AS fee_micro_usd,"+ledgerSum("cost")+" AS cost_micro_usd,"+ledgerSum("input_tokens")+" AS confirmed_input_tokens,"+ledgerSum("output_tokens")+" AS confirmed_output_tokens"+base+" GROUP BY r.alias ORDER BY r.alias", arguments...)
		if err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		reply(writer, 200, map[string]any{"items": values, "currency": "USD", "unknown_is_not_zero": true})
		return
	}
	page, size := 1, 25
	if value := request.URL.Query().Get("page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100000 {
			failure(writer, 400, "invalid_page", "")
			return
		}
		page = parsed
	}
	if value := request.URL.Query().Get("page_size"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			failure(writer, 400, "invalid_page_size", "")
			return
		}
		size = parsed
	}
	var total int64
	if transaction.QueryRow("SELECT count(*)"+base, args...).Scan(&total) != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	export := len(parts) == 2 && parts[1] == "export"
	if export {
		if total > 10000 {
			failure(writer, 413, "export_requires_narrower_filter", "")
			return
		}
		size, page = 10000, 1
	}
	arguments := append(append([]any{}, args...), size, (page-1)*size)
	values, err := queryMaps(transaction, "SELECT r.id,r.time,c.tenant_id,r.key_id,r.alias,r.state,r.error_code,r.duration_ms,r.prompt_tokens,r.completion_tokens,r.total_tokens,a.period,a.pending,a.fee_reserved,(SELECT coalesce(sum(e.cost),0) FROM charge_entries e WHERE e.request_id=r.id) AS cost_micro_usd,(SELECT coalesce(sum(e.fee),0) FROM charge_entries e WHERE e.request_id=r.id) AS fee_micro_usd,(SELECT count(*) FROM legacy_pending l WHERE l.request_id=r.id AND l.resolved=0) AS legacy_pending"+base+" ORDER BY r.time DESC,r.id DESC LIMIT ? OFFSET ?", arguments...)
	if err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	if export {
		if err := addConfirmedTokens(transaction, values); err != nil {
			failure(writer, 503, "storage_unavailable", "")
			return
		}
		writeCSV(writer, []string{"id", "time", "tenant_id", "key_id", "alias", "state", "error_code", "duration_ms", "prompt_tokens", "completion_tokens", "total_tokens", "confirmed_input_tokens", "confirmed_output_tokens", "period", "pending", "fee_reserved", "cost_micro_usd", "fee_micro_usd", "legacy_pending"}, values)
		return
	}
	if err := addConfirmedTokens(transaction, values); err != nil {
		failure(writer, 503, "storage_unavailable", "")
		return
	}
	reply(writer, 200, map[string]any{"items": values, "total": total, "page": page, "page_size": size})
}

func addConfirmedTokens(transaction *sql.Tx, values []map[string]any) error {
	for _, value := range values {
		var input, output int64
		if err := transaction.QueryRow("SELECT coalesce(sum(input_tokens),0),coalesce(sum(output_tokens),0) FROM charge_entries WHERE request_id=?", value["id"]).Scan(&input, &output); err != nil {
			return err
		}
		value["confirmed_input_tokens"], value["confirmed_output_tokens"] = strconv.FormatInt(input, 10), strconv.FormatInt(output, 10)
	}
	return nil
}

func writeCSV(writer http.ResponseWriter, columns []string, values []map[string]any) {
	writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	writer.Header().Set("Content-Disposition", "attachment; filename=relay-export.csv")
	output := csv.NewWriter(writer)
	_ = output.Write(columns)
	for _, value := range values {
		row := []string{}
		for _, column := range columns {
			text := ""
			switch item := value[column].(type) {
			case string:
				text = item
			case int64:
				text = strconv.FormatInt(item, 10)
			case bool:
				text = strconv.FormatBool(item)
			}
			_, numericError := strconv.ParseInt(text, 10, 64)
			amountColumn := column == "amount" || column == "cost" || column == "fee" || column == "original_fee" || column == "original_cost" || strings.HasSuffix(column, "_micro_usd")
			if len(text) > 0 && strings.ContainsRune("=+-@\t\r", rune(text[0])) && !(amountColumn && numericError == nil) {
				text = "'" + text
			}
			row = append(row, text)
		}
		_ = output.Write(row)
	}
	output.Flush()
}
