package upstream

import (
	"bytes"
	"encoding/json"
	"strings"

	"model-relay/internal/jsoncodec"
	"model-relay/internal/names"
)

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
	value, ok := jsoncodec.Object(raw)
	if !ok || !onlyKeys(value, "name description parameters strict") {
		return false
	}
	name, ok := stringValue(value["name"])
	if !ok || !names.Valid(name) {
		return false
	}
	if description, exists := value["description"]; exists {
		if _, ok := stringValue(description); !ok {
			return false
		}
	}
	if parameters, exists := value["parameters"]; exists {
		if _, ok := jsoncodec.Object(parameters); !ok {
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
		value, ok := jsoncodec.Object(call)
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
		function, ok := jsoncodec.Object(value["function"])
		if !ok || !onlyKeys(function, "name arguments") {
			return false
		}
		name, ok := stringValue(function["name"])
		if !ok || !names.Valid(name) {
			return false
		}
		if _, ok = stringValue(function["arguments"]); !ok {
			return false
		}
	}
	return true
}
func ValidateRequest(data []byte) (map[string]json.RawMessage, string, bool, string) {
	if !jsoncodec.Unique(data) {
		return nil, "", false, "invalid_request"
	}
	value, ok := jsoncodec.Object(data)
	if !ok {
		return nil, "", false, "invalid_request"
	}
	if !onlyKeys(value, "model messages stream temperature top_p max_tokens max_completion_tokens n stop presence_penalty frequency_penalty seed logit_bias logprobs top_logprobs user tools tool_choice parallel_tool_calls response_format stream_options") {
		return nil, "", false, "unsupported_field"
	}
	alias, ok := stringValue(value["model"])
	if !ok || !names.Alias(alias) {
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
		message, ok := jsoncodec.Object(raw)
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
			if value, ok := stringValue(name); !ok || !names.Valid(value) {
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
			members, ok := jsoncodec.Object(raw)
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
				entry, ok := jsoncodec.Object(tool)
				kind, _ := stringValue(entry["type"])
				if !ok || !onlyKeys(entry, "type function") || kind != "function" || !functionDefinition(entry["function"]) {
					valid = false
				}
			}
		case "tool_choice":
			choice, ok := stringValue(raw)
			valid = ok && (choice == "auto" || choice == "none" || choice == "required")
			if !ok {
				entry, ok := jsoncodec.Object(raw)
				kind, _ := stringValue(entry["type"])
				function, functionOK := jsoncodec.Object(entry["function"])
				name, nameOK := stringValue(function["name"])
				valid = ok && onlyKeys(entry, "type function") && kind == "function" && functionOK && onlyKeys(function, "name") && nameOK && names.Valid(name)
			}
		case "response_format":
			entry, ok := jsoncodec.Object(raw)
			kind, _ := stringValue(entry["type"])
			valid = ok && onlyKeys(entry, "type json_schema") && (kind == "text" || kind == "json_object" || kind == "json_schema")
			if kind == "json_schema" {
				schema, ok := jsoncodec.Object(entry["json_schema"])
				name, nameOK := stringValue(schema["name"])
				_, schemaOK := jsoncodec.Object(schema["schema"])
				valid = valid && ok && onlyKeys(schema, "name description strict schema") && nameOK && names.Valid(name) && schemaOK
				if raw, exists := schema["strict"]; exists {
					_, ok := boolValue(raw)
					valid = valid && ok
				}
			} else if _, exists := entry["json_schema"]; exists {
				valid = false
			}
		case "stream_options":
			options, ok := jsoncodec.Object(raw)
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
	usage, ok := jsoncodec.Object(value["usage"])
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
	if !jsoncodec.Unique(data) {
		return nil, false
	}
	value, ok := jsoncodec.Object(data)
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
		if _, ok := jsoncodec.Object(value["usage"]); !ok {
			return nil, false
		}
	}
	for _, raw := range choices {
		choice, ok := jsoncodec.Object(raw)
		if !ok {
			return nil, false
		}
		field := "message"
		if stream {
			field = "delta"
		}
		if _, ok := jsoncodec.Object(choice[field]); !ok {
			return nil, false
		}
		message, _ := jsoncodec.Object(choice[field])
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

func outputDelta(value map[string]json.RawMessage) bool {
	choices, ok := array(value["choices"])
	if !ok {
		return false
	}
	for _, raw := range choices {
		choice, _ := jsoncodec.Object(raw)
		delta, _ := jsoncodec.Object(choice["delta"])
		for _, field := range []string{"content", "refusal"} {
			if text, ok := stringValue(delta[field]); ok && text != "" {
				return true
			}
		}
		calls, _ := array(delta["tool_calls"])
		for _, callRaw := range calls {
			call, _ := jsoncodec.Object(callRaw)
			function, _ := jsoncodec.Object(call["function"])
			for _, field := range []string{"name", "arguments"} {
				if text, ok := stringValue(function[field]); ok && text != "" {
					return true
				}
			}
		}
	}
	return false
}
