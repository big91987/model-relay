package admission

import "encoding/json"

func boundedPayload(payload map[string]json.RawMessage, maximumOutput int64) bool {
	allowed := map[string]bool{"model": true, "messages": true, "stream": true, "stream_options": true, "max_tokens": true, "max_completion_tokens": true, "n": true, "temperature": true, "top_p": true, "seed": true, "stop": true}
	for field := range payload {
		if !allowed[field] {
			return false
		}
	}
	var messages []map[string]json.RawMessage
	if json.Unmarshal(payload["messages"], &messages) != nil || len(messages) == 0 {
		return false
	}
	for _, message := range messages {
		var role, content string
		if len(message) != 2 || json.Unmarshal(message["role"], &role) != nil || json.Unmarshal(message["content"], &content) != nil || (role != "system" && role != "user" && role != "assistant") {
			return false
		}
	}
	if payload["max_tokens"] != nil && payload["max_completion_tokens"] != nil {
		return false
	}
	field := "max_tokens"
	if payload["max_completion_tokens"] != nil {
		field = "max_completion_tokens"
	}
	if raw := payload[field]; raw != nil {
		var requested int64
		return json.Unmarshal(raw, &requested) == nil && requested > 0 && requested <= maximumOutput
	}
	payload[field], _ = json.Marshal(maximumOutput)
	return true
}
