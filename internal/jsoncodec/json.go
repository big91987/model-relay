// Package jsoncodec implements strict JSON syntax shared by HTTP input and provider protocols.
package jsoncodec

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

// Unique requires valid UTF-8, one complete JSON value and unique object member names at every depth.
func Unique(data []byte) bool {
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

// Object decodes a non-null JSON object. Call Unique first when duplicate names must be rejected.
func Object(raw json.RawMessage) (map[string]json.RawMessage, bool) {
	var value map[string]json.RawMessage
	err := json.Unmarshal(raw, &value)
	return value, err == nil && value != nil
}
