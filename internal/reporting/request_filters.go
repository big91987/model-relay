package reporting

import (
	"database/sql"
	"strconv"
	"time"

	"model-relay/internal/access"
	"model-relay/internal/fault"
)

func requestScope(transaction *sql.Tx, principal access.Principal, input Filters) (string, []any, error) {
	where, args, err := reportScope(transaction, principal, input.Tenant)
	if err != nil {
		return "", nil, err
	}
	for _, filter := range []struct{ value, column string }{{input.Model, "r.alias"}, {input.Key, "r.key_id"}, {input.State, "r.state"}, {input.RequestID, "r.id"}} {
		if filter.value != "" {
			where += " AND " + filter.column + "=?"
			args = append(args, filter.value)
		}
	}
	for _, filter := range []struct{ value, operator string }{{input.From, ">="}, {input.To, "<="}} {
		if filter.value != "" {
			if _, err := time.Parse(time.RFC3339Nano, filter.value); err != nil {
				return "", nil, fault.New(fault.Invalid, "invalid_time_filter")
			}
			where += " AND julianday(r.time)" + filter.operator + "julianday(?)"
			args = append(args, filter.value)
		}
	}
	return where, args, nil
}

func pages(input PageInput) (int, int, error) {
	page, size := 1, 25
	if input.Page != "" {
		parsed, err := strconv.Atoi(input.Page)
		if err != nil || parsed < 1 || parsed > 100000 {
			return 0, 0, fault.New(fault.Invalid, "invalid_page")
		}
		page = parsed
	}
	if input.Size != "" {
		parsed, err := strconv.Atoi(input.Size)
		if err != nil || parsed < 1 || parsed > 100 {
			return 0, 0, fault.New(fault.Invalid, "invalid_page_size")
		}
		size = parsed
	}
	return page, size, nil
}

const requestsFrom = " FROM requests r LEFT JOIN key_controls c ON c.key_id=r.key_id LEFT JOIN admissions a ON a.request_id=r.id WHERE "
