package upstream

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"model-relay/internal/jsoncodec"
)

var ErrProtocol = errors.New("upstream protocol error")

// ErrIncomplete distinguishes clean EOF without DONE from malformed frames.
// It remains a protocol error unless the caller closed the body for a deadline.
var ErrIncomplete = fmt.Errorf("%w: missing terminal event", ErrProtocol)

type Usage struct {
	Prompt     *int64
	Completion *int64
	Total      *int64
}

type Completion struct {
	Data  []byte
	Usage Usage
}

// ValidateChatResponse checks the response media type. HTTP status policy is
// intentionally left with the attempt lifecycle and retry decisions.
func ValidateChatResponse(response *http.Response, stream bool) error {
	kind, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	expected := "application/json"
	if stream {
		expected = "text/event-stream"
	}
	if err != nil || kind != expected {
		return ErrProtocol
	}
	return nil
}

// ReadCompletion validates one ordinary response without changing its bytes.
// It leaves closing the body to the caller, including on a protocol failure.
func ReadCompletion(response *http.Response) (Completion, error) {
	if err := ValidateChatResponse(response, false); err != nil {
		return Completion{}, err
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, (8<<20)+1))
	if err != nil {
		return Completion{}, err
	}
	value, valid := validResponse(data, false)
	if !valid || len(data) > 8<<20 {
		return Completion{}, ErrProtocol
	}
	return Completion{Data: data, Usage: usageFrom(value)}, nil
}

// ReadErrorUsage extracts only known, nonnegative integer counters from a
// bounded error body. Invalid, duplicate or unreadable bodies yield unknown use.
func ReadErrorUsage(body io.Reader) Usage {
	data, err := io.ReadAll(io.LimitReader(body, (1<<20)+1))
	if err == nil && len(data) <= 1<<20 && jsoncodec.Unique(data) {
		if value, ok := jsoncodec.Object(data); ok {
			return usageFrom(value)
		}
	}
	return Usage{}
}
