package sqlread

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

func Maps(transaction *sql.Tx, query string, args ...any) ([]map[string]any, error) {
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

func Strings(transaction *sql.Tx, query string, arguments ...any) ([]string, error) {
	rows, err := transaction.Query(query, arguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []string{}
	for rows.Next() {
		var value string
		if rows.Scan(&value) != nil {
			return nil, errors.New("storage scan failed")
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
