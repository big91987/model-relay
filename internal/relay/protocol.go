package relay

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

func uniqueJSON(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var walk func() bool
	walk = func() bool {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return true
		}
		switch delimiter {
		case '{':
			seen := map[string]bool{}
			for decoder.More() {
				name, err := decoder.Token()
				if err != nil {
					return false
				}
				key, ok := name.(string)
				if !ok || seen[key] {
					return false
				}
				seen[key] = true
				if !walk() {
					return false
				}
			}
			end, err := decoder.Token()
			return err == nil && end == json.Delim('}')
		case '[':
			for decoder.More() {
				if !walk() {
					return false
				}
			}
			end, err := decoder.Token()
			return err == nil && end == json.Delim(']')
		default:
			return false
		}
	}
	if !walk() {
		return false
	}
	_, err := decoder.Token()
	return err == io.EOF
}
func stringValue(raw json.RawMessage) (string, bool) {
	var value string
	err := json.Unmarshal(raw, &value)
	return value, err == nil && string(raw) != "null"
}
func boolValue(raw json.RawMessage) (bool, bool) {
	var value bool
	err := json.Unmarshal(raw, &value)
	return value, err == nil && string(raw) != "null"
}
func object(raw json.RawMessage) (map[string]json.RawMessage, bool) {
	var value map[string]json.RawMessage
	err := json.Unmarshal(raw, &value)
	return value, err == nil && value != nil
}
func array(raw json.RawMessage) ([]json.RawMessage, bool) {
	var value []json.RawMessage
	err := json.Unmarshal(raw, &value)
	return value, err == nil && value != nil
}
func onlyKeys(value map[string]json.RawMessage, allowed string) bool {
	for key := range value {
		if !strings.Contains(" "+allowed+" ", " "+key+" ") {
			return false
		}
	}
	return true
}
func positiveInteger(raw json.RawMessage) bool {
	var value int64
	return json.Unmarshal(raw, &value) == nil && value > 0
}
func numberValue(raw json.RawMessage) bool {
	var value json.Number
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var data any
	if decoder.Decode(&data) != nil {
		return false
	}
	value, ok := data.(json.Number)
	if !ok {
		return false
	}
	_, err := value.Float64()
	return err == nil
}
func functionDefinition(raw json.RawMessage) bool {
	value, ok := object(raw)
	if !ok || !onlyKeys(value, "name description parameters strict") {
		return false
	}
	name, ok := stringValue(value["name"])
	if !ok || !validName(name) {
		return false
	}
	if description, exists := value["description"]; exists {
		if _, ok := stringValue(description); !ok {
			return false
		}
	}
	if parameters, exists := value["parameters"]; exists {
		if _, ok := object(parameters); !ok {
			return false
		}
	}
	if strict, exists := value["strict"]; exists {
		if _, ok := boolValue(strict); !ok {
			return false
		}
	}
	return true
}
func toolCalls(raw json.RawMessage) bool {
	calls, ok := array(raw)
	if !ok || len(calls) == 0 {
		return false
	}
	for _, call := range calls {
		value, ok := object(call)
		if !ok || !onlyKeys(value, "id type function") {
			return false
		}
		id, ok := stringValue(value["id"])
		if !ok || id == "" {
			return false
		}
		kind, _ := stringValue(value["type"])
		if kind != "function" {
			return false
		}
		function, ok := object(value["function"])
		if !ok || !onlyKeys(function, "name arguments") {
			return false
		}
		name, ok := stringValue(function["name"])
		if !ok || !validName(name) {
			return false
		}
		if _, ok = stringValue(function["arguments"]); !ok {
			return false
		}
	}
	return true
}
func validateRequest(data []byte) (map[string]json.RawMessage, string, bool, string) {
	if !uniqueJSON(data) {
		return nil, "", false, "invalid_request"
	}
	value, ok := object(data)
	if !ok {
		return nil, "", false, "invalid_request"
	}
	if !onlyKeys(value, "model messages stream temperature top_p max_tokens max_completion_tokens n stop presence_penalty frequency_penalty seed logit_bias logprobs top_logprobs user tools tool_choice parallel_tool_calls response_format stream_options") {
		return nil, "", false, "unsupported_field"
	}
	alias, ok := stringValue(value["model"])
	if !ok || !validAlias(alias) {
		return nil, "", false, "invalid_request"
	}
	stream := false
	if raw, exists := value["stream"]; exists {
		stream, ok = boolValue(raw)
		if !ok {
			return nil, alias, false, "invalid_request"
		}
	}
	messages, ok := array(value["messages"])
	if !ok || len(messages) == 0 {
		return nil, alias, stream, "invalid_request"
	}
	for _, raw := range messages {
		message, ok := object(raw)
		if !ok || !onlyKeys(message, "role content name tool_calls tool_call_id") {
			return nil, alias, stream, "unsupported_field"
		}
		role, _ := stringValue(message["role"])
		if !strings.Contains(" system developer user assistant tool ", " "+role+" ") || role == "" {
			return nil, alias, stream, "invalid_request"
		}
		_, text := stringValue(message["content"])
		calls, hasCalls := message["tool_calls"]
		if !text && !(role == "assistant" && string(message["content"]) == "null" && hasCalls) {
			return nil, alias, stream, "invalid_request"
		}
		if hasCalls && (role != "assistant" || !toolCalls(calls)) {
			return nil, alias, stream, "invalid_request"
		}
		if name, exists := message["name"]; exists {
			if value, ok := stringValue(name); !ok || !validName(value) {
				return nil, alias, stream, "invalid_request"
			}
		}
		if role == "tool" {
			id, ok := stringValue(message["tool_call_id"])
			if !ok || id == "" {
				return nil, alias, stream, "invalid_request"
			}
		} else if _, exists := message["tool_call_id"]; exists {
			return nil, alias, stream, "invalid_request"
		}
	}
	for key, raw := range value {
		valid := true
		switch key {
		case "temperature", "top_p", "presence_penalty", "frequency_penalty":
			valid = numberValue(raw)
		case "max_tokens", "max_completion_tokens", "n":
			valid = positiveInteger(raw)
		case "seed", "top_logprobs":
			var number int64
			valid = json.Unmarshal(raw, &number) == nil && string(raw) != "null"
			if key == "top_logprobs" {
				valid = valid && number >= 0
			}
		case "logprobs", "parallel_tool_calls":
			_, valid = boolValue(raw)
		case "user":
			_, valid = stringValue(raw)
		case "stop":
			if _, valid = stringValue(raw); !valid {
				list, ok := array(raw)
				valid = ok
				for _, item := range list {
					if _, ok = stringValue(item); !ok {
						valid = false
					}
				}
			}
		case "logit_bias":
			members, ok := object(raw)
			valid = ok
			for _, item := range members {
				if !numberValue(item) {
					valid = false
				}
			}
		case "tools":
			tools, ok := array(raw)
			valid = ok
			for _, tool := range tools {
				entry, ok := object(tool)
				kind, _ := stringValue(entry["type"])
				if !ok || !onlyKeys(entry, "type function") || kind != "function" || !functionDefinition(entry["function"]) {
					valid = false
				}
			}
		case "tool_choice":
			choice, ok := stringValue(raw)
			valid = ok && (choice == "auto" || choice == "none" || choice == "required")
			if !ok {
				entry, ok := object(raw)
				kind, _ := stringValue(entry["type"])
				function, functionOK := object(entry["function"])
				name, nameOK := stringValue(function["name"])
				valid = ok && onlyKeys(entry, "type function") && kind == "function" && functionOK && onlyKeys(function, "name") && nameOK && validName(name)
			}
		case "response_format":
			entry, ok := object(raw)
			kind, _ := stringValue(entry["type"])
			valid = ok && onlyKeys(entry, "type json_schema") && (kind == "text" || kind == "json_object" || kind == "json_schema")
			if kind == "json_schema" {
				schema, ok := object(entry["json_schema"])
				name, nameOK := stringValue(schema["name"])
				_, schemaOK := object(schema["schema"])
				valid = valid && ok && onlyKeys(schema, "name description strict schema") && nameOK && validName(name) && schemaOK
				if raw, exists := schema["strict"]; exists {
					_, ok := boolValue(raw)
					valid = valid && ok
				}
			} else if _, exists := entry["json_schema"]; exists {
				valid = false
			}
		case "stream_options":
			options, ok := object(raw)
			valid = stream && ok && onlyKeys(options, "include_usage")
			if include, exists := options["include_usage"]; exists {
				_, ok = boolValue(include)
				valid = valid && ok
			}
		}
		if !valid {
			return nil, alias, stream, "invalid_request"
		}
	}
	return value, alias, stream, ""
}
func usageFrom(value map[string]json.RawMessage) Usage {
	usage, ok := object(value["usage"])
	if !ok {
		return Usage{}
	}
	parse := func(raw json.RawMessage) *int64 {
		var number int64
		if json.Unmarshal(raw, &number) != nil || string(raw) == "null" || len(raw) == 0 || number < 0 {
			return nil
		}
		return &number
	}
	return Usage{Prompt: parse(usage["prompt_tokens"]), Completion: parse(usage["completion_tokens"]), Total: parse(usage["total_tokens"])}
}
func validResponse(data []byte, stream bool) (map[string]json.RawMessage, bool) {
	if !uniqueJSON(data) {
		return nil, false
	}
	value, ok := object(data)
	if !ok {
		return nil, false
	}
	if _, exists := value["error"]; exists {
		return nil, false
	}
	choices, ok := array(value["choices"])
	if !ok || !stream && len(choices) == 0 {
		return nil, false
	}
	if stream && len(choices) == 0 {
		if _, ok := object(value["usage"]); !ok {
			return nil, false
		}
	}
	for _, raw := range choices {
		choice, ok := object(raw)
		if !ok {
			return nil, false
		}
		field := "message"
		if stream {
			field = "delta"
		}
		if _, ok := object(choice[field]); !ok {
			return nil, false
		}
		message, _ := object(choice[field])
		if content, exists := message["content"]; exists && string(content) != "null" {
			if _, ok := stringValue(content); !ok {
				return nil, false
			}
		}
		if !stream {
			role, _ := stringValue(message["role"])
			_, text := stringValue(message["content"])
			calls, tools := message["tool_calls"]
			refusal, refusalOK := stringValue(message["refusal"])
			if role != "assistant" || !text && !(tools && toolCalls(calls)) && !(refusalOK && refusal != "") || tools && !toolCalls(calls) {
				return nil, false
			}
		}
	}
	return value, true
}
