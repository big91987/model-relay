package httpapi

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"model-relay/internal/jsoncodec"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Reply(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func Failure(writer http.ResponseWriter, status int, code, id string) {
	Reply(writer, status, map[string]any{"error": map[string]string{"message": "Request could not be completed: " + code, "type": "relay_error", "code": code}, "request_id": id})
}

func Input(writer http.ResponseWriter, request *http.Request, target any) error {
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
	if !jsoncodec.Unique(data) {
		return errors.New("invalid JSON")
	}
	if _, ok := jsoncodec.Object(data); !ok {
		return errors.New("JSON object required")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func Collection[Value any](writer http.ResponseWriter, request *http.Request, values []Value, fields ...map[string]any) {
	page, size := 1, 25
	for _, parameter := range []struct {
		name    string
		target  *int
		maximum int
	}{{"page", &page, 100000}, {"page_size", &size, 100}} {
		if raw := request.URL.Query().Get(parameter.name); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value < 1 || value > parameter.maximum {
				Failure(writer, 400, "invalid_pagination", "")
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
	Reply(writer, 200, result)
}

func CSV(writer http.ResponseWriter, columns []string, values []map[string]any) {
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
