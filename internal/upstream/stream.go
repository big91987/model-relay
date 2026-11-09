package upstream

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"model-relay/internal/jsoncodec"
)

type EventKind uint8

const (
	EventData EventKind = iota
	EventComment
	EventDone
)

type Event struct {
	Kind            EventKind
	Wire            string
	Usage           Usage
	HasUsage        bool
	EffectiveOutput bool
}

// Stream parses one provider response. It owns framing and protocol progress;
// the gateway decides which events commit output and when to refresh deadlines.
type Stream struct {
	scanner *bufio.Scanner
	total   int
	data    bool
	done    bool
}

func NewStream(body io.Reader) *Stream {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 4096), 256*1024)
	return &Stream{scanner: scanner}
}

// Next returns a complete validated event. A clean EOF without [DONE], invalid
// JSON, an unsupported SSE field or an exceeded byte limit is ErrProtocol.
// Read failures retain their transport error for caller-owned classification.
func (stream *Stream) Next() (Event, error) {
	if stream.done {
		return Event{}, io.EOF
	}
	var lines []string
	eventBytes := 0
	for stream.scanner.Scan() {
		line := strings.TrimSuffix(stream.scanner.Text(), "\r")
		eventBytes += len(line) + 1
		stream.total += len(line) + 1
		if eventBytes > 256*1024 || stream.total > 32<<20 || !utf8.ValidString(line) {
			return Event{}, ErrProtocol
		}
		if line != "" {
			if !strings.HasPrefix(line, ":") && line != "data" && !strings.HasPrefix(line, "data:") {
				return Event{}, ErrProtocol
			}
			lines = append(lines, line)
			continue
		}
		var dataLines []string
		var comments []string
		for _, entry := range lines {
			if strings.HasPrefix(entry, ":") {
				comments = append(comments, entry)
			} else {
				value := strings.TrimPrefix(entry, "data")
				value = strings.TrimPrefix(value, ":")
				value = strings.TrimPrefix(value, " ")
				dataLines = append(dataLines, value)
			}
		}
		lines = nil
		eventBytes = 0
		if len(dataLines) == 0 {
			if len(comments) > 0 {
				return Event{Kind: EventComment, Wire: strings.Join(comments, "\n") + "\n\n"}, nil
			}
			continue
		}
		payload := strings.Join(dataLines, "\n")
		if payload == "[DONE]" {
			if !stream.data {
				return Event{}, ErrProtocol
			}
			stream.done = true
			return Event{Kind: EventDone, Wire: "data: [DONE]\n\n"}, nil
		}
		value, ok := validResponse([]byte(payload), true)
		if !ok {
			return Event{}, ErrProtocol
		}
		_, hasUsage := jsoncodec.Object(value["usage"])
		stream.data = true
		return Event{Kind: EventData, Wire: "data: " + strings.Join(dataLines, "\ndata: ") + "\n\n", Usage: usageFrom(value), HasUsage: hasUsage, EffectiveOutput: outputDelta(value)}, nil
	}
	err := stream.scanner.Err()
	if err == nil {
		return Event{}, ErrIncomplete
	}
	if errors.Is(err, bufio.ErrTooLong) {
		return Event{}, ErrProtocol
	}
	return Event{}, err
}
