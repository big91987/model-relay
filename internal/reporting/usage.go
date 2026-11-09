package reporting

import (
	"context"
	"strings"

	"model-relay/internal/access"
	"model-relay/internal/sqlread"
)

func (service *Service) Usage(ctx context.Context, principal access.Principal, filters Filters) (UsageSummary, error) {
	transaction, err := service.begin(ctx)
	if err != nil {
		return UsageSummary{}, err
	}
	defer transaction.Rollback()
	where, args, err := requestScope(transaction, principal, filters)
	if err != nil {
		return UsageSummary{}, err
	}
	ledgerSum := func(field string) string {
		return "(SELECT coalesce(sum(e." + field + "),0) FROM charge_entries e JOIN requests er ON er.id=e.request_id LEFT JOIN key_controls ec ON ec.key_id=er.key_id WHERE er.alias=r.alias AND " + strings.ReplaceAll(strings.ReplaceAll(where, "c.", "ec."), "r.", "er.") + ")"
	}
	arguments := []any{}
	for range 5 {
		arguments = append(arguments, args...)
	}
	values, err := sqlread.Maps(transaction, "SELECT r.alias,count(*) AS requests,coalesce(sum(a.pending),0) AS pending,coalesce(sum(a.fee_reserved),0) AS reserved_micro_usd,"+ledgerSum("fee")+" AS fee_micro_usd,"+ledgerSum("cost")+" AS cost_micro_usd,"+ledgerSum("input_tokens")+" AS confirmed_input_tokens,"+ledgerSum("output_tokens")+" AS confirmed_output_tokens"+requestsFrom+where+" GROUP BY r.alias ORDER BY r.alias", arguments...)
	if err != nil {
		return UsageSummary{}, storage()
	}
	return UsageSummary{Items: values, Currency: "USD", UnknownIsNotZero: true}, nil
}
