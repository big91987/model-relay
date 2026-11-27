package reporting

import (
	"context"
	"database/sql"
	"strconv"

	"model-relay/internal/access"
	"model-relay/internal/fault"
	"model-relay/internal/records"
	"model-relay/internal/sqlread"
)

func (service *Service) Requests(ctx context.Context, principal access.Principal, filters Filters, input PageInput) (RequestPage, error) {
	return service.requests(ctx, principal, filters, input, false)
}

func (service *Service) ExportRequests(ctx context.Context, principal access.Principal, filters Filters, input PageInput) ([]map[string]any, error) {
	page, err := service.requests(ctx, principal, filters, input, true)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (service *Service) requests(ctx context.Context, principal access.Principal, filters Filters, input PageInput, export bool) (RequestPage, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return RequestPage{}, err
	}
	defer transaction.Rollback()
	where, args, err := requestScope(transaction, principal, filters)
	if err != nil {
		return RequestPage{}, err
	}
	page, size, err := pages(input)
	if err != nil {
		return RequestPage{}, err
	}
	base := requestsFrom + where
	var total int64
	if transaction.QueryRow("SELECT count(*)"+base, args...).Scan(&total) != nil {
		return RequestPage{}, storage()
	}
	if export {
		if total > 10000 {
			return RequestPage{}, fault.New(fault.Invalid, "export_requires_narrower_filter")
		}
		size, page = 10000, 1
	}
	arguments := append(append([]any{}, args...), size, (page-1)*size)
	values, err := sqlread.Maps(transaction, "SELECT r.id,r.time,c.tenant_id,r.key_id,r.alias,r.state,r.error_code,r.duration_ms,r.prompt_tokens,r.completion_tokens,r.total_tokens,a.period,a.pending,a.fee_reserved,(SELECT coalesce(sum(e.cost),0) FROM charge_entries e WHERE e.request_id=r.id) AS cost_micro_usd,(SELECT coalesce(sum(e.fee),0) FROM charge_entries e WHERE e.request_id=r.id) AS fee_micro_usd,(SELECT count(*) FROM legacy_pending l WHERE l.request_id=r.id AND l.resolved=0) AS legacy_pending"+base+" ORDER BY r.time DESC,r.id DESC LIMIT ? OFFSET ?", arguments...)
	if err != nil {
		return RequestPage{}, storage()
	}
	if addConfirmedTokens(transaction, values) != nil {
		return RequestPage{}, storage()
	}
	return RequestPage{Items: values, Total: total, Page: page, Size: size}, nil
}

func (service *Service) Request(ctx context.Context, principal access.Principal, id string, filters Filters) (map[string]any, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	where, args, err := requestScope(transaction, principal, filters)
	if err != nil {
		return nil, err
	}
	base := requestsFrom + where + " AND r.id=?"
	args = append(args, id)
	values, err := sqlread.Maps(transaction, "SELECT r.id,r.time,r.key_id,c.tenant_id,r.alias,r.model,r.state,r.error_code,r.stream,r.upstream_status,r.duration_ms,r.prompt_tokens,r.completion_tokens,r.total_tokens,a.period,a.pending,a.tokens_reserved,a.fee_reserved,a.revision,(SELECT count(*) FROM legacy_pending l WHERE l.request_id=r.id AND l.resolved=0) AS legacy_pending"+base, args...)
	if err != nil {
		return nil, storage()
	}
	if len(values) != 1 {
		return nil, fault.New(fault.NotFound, "not_found")
	}
	attempts, err := sqlread.Maps(transaction, "SELECT a.ordinal,a.upstream_id,a.upstream_name,a.remote_model,a.started_at,a.finished_at,a.state,a.error_code,a.upstream_status,a.prompt_tokens,a.completion_tokens,a.total_tokens,d.possible AS dispatch_possible FROM request_attempts a LEFT JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal WHERE a.request_id=? ORDER BY a.ordinal", id)
	if err != nil {
		return nil, storage()
	}
	prices, err := sqlread.Maps(transaction, "SELECT upstream_id,price_id,cost_input,cost_output,tenant_input,tenant_output,max_input,max_output FROM request_prices WHERE request_id=? ORDER BY upstream_id", id)
	if err != nil {
		return nil, storage()
	}
	entries, err := sqlread.Maps(transaction, "SELECT ordinal,revision,period,input_tokens,output_tokens,cost,fee,reason,evidence FROM charge_entries WHERE request_id=? ORDER BY ordinal,revision", id)
	if err != nil {
		return nil, storage()
	}
	decisions, observations, err := records.ReadRoutingEvidence(transaction, id)
	if err != nil {
		return nil, storage()
	}
	values[0]["routing_decisions"], values[0]["latency_observations"] = decisions, observations
	values[0]["attempts"], values[0]["price_snapshots"], values[0]["charge_entries"] = attempts, prices, entries
	if addConfirmedTokens(transaction, values) != nil {
		return nil, storage()
	}
	history, err := sqlread.Maps(transaction, "SELECT upstream_id,price_reference,cost_input,cost_output,tenant_input,tenant_output,reason,evidence FROM historical_price_evidence WHERE request_id=? ORDER BY upstream_id", id)
	if err != nil {
		return nil, storage()
	}
	values[0]["historical_price_evidence"] = history
	return values[0], nil
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
