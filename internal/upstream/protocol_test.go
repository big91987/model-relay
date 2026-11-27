package upstream_test

import (
	"errors"
	"io"

	"model-relay/internal/upstream"
	"net/http"
	"strings"
	"testing"
)

const ordinary = `{"model":"alpha","messages":[{"role":"user","content":"controlled"}]}`

func TestRequestContract(t *testing.T) {
	for _, payload := range []string{
		`{"model":"a","model":"b","messages":[]}`, ordinary + ` {}`,
		`{"model":"alpha","messages":[{"role":"user","content":[{"type":"image"}]}]}`,
		`{"model":"alpha","messages":[{"role":"user","content":"text"}],"base_url":"http://evil"}`,
		`{"model":"alpha","messages":[{"role":"user","content":"text"}],"stream":"true"}`,
		`{"model":"alpha","messages":[{"role":"user","content":"text"}],"n":0}`,
	} {
		if _, _, _, code := upstream.ValidateRequest([]byte(payload)); code == "" {
			t.Fatalf("invalid request accepted: %s", payload)
		}
	}
	value, model, stream, code := upstream.ValidateRequest([]byte(ordinary))
	if code != "" || model != "alpha" || stream || string(value["model"]) != `"alpha"` {
		t.Fatal("ordinary request changed", model, stream, code)
	}
	for _, payload := range []string{
		`{"model":"alpha","messages":[{"role":"assistant","content":null,"tool_calls":[{"id":"c","type":"function","function":{"name":"f","arguments":"{}"}}]},{"role":"tool","tool_call_id":"c","content":"ok"}],"stream":true,"stream_options":{"include_usage":true}}`,
		`{"model":"alpha","messages":[{"role":"developer","content":"controlled"}],"tools":[{"type":"function","function":{"name":"f","parameters":{},"strict":true}}],"tool_choice":{"type":"function","function":{"name":"f"}},"response_format":{"type":"json_schema","json_schema":{"name":"a","schema":{},"strict":true}}}`,
	} {
		if _, _, _, code := upstream.ValidateRequest([]byte(payload)); code != "" {
			t.Fatal("supported tool/structured request rejected", code)
		}
	}
}

func completion(body, contentType string) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body))}
}

func TestClientCompletionAssistantRoleContract(t *testing.T) {
	for _, scenario := range []struct {
		name, body   string
		valid, known bool
	}{
		{"missing_role_known", `{"choices":[{"message":{"content":"controlled"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`, false, true},
		{"missing_role_unknown", `{"choices":[{"message":{"content":"controlled"}}]}`, false, false},
		{"wrong_role", `{"choices":[{"message":{"role":"user","content":"controlled"}}]}`, false, false},
		{"assistant_known", `{"choices":[{"message":{"role":"assistant","content":"controlled"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`, true, true},
		{"assistant_unknown", `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`, true, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			result, err := upstream.ReadCompletion(completion(scenario.body, "application/json; charset=utf-8"))
			if (err == nil) != scenario.valid {
				t.Fatal("completion assistant role contract mismatch", err)
			}
			if scenario.valid && ((result.Usage.Prompt != nil && result.Usage.Completion != nil) != scenario.known || string(result.Data) != scenario.body) {
				t.Fatal("completion bytes or usage changed")
			}
		})
	}
}

func TestCompletionRejectsMalformedOrAmbiguousProviderData(t *testing.T) {
	for _, body := range []string{
		`{"choices":[]}`, `{"choices":[{"message":{"role":"assistant","content":2}}]}`,
		`{"choices":[{"message":{"role":"assistant","content":"a","content":"b"}}]}`,
		`{"choices":[{"message":{"role":"assistant","content":"a"}}],"error":null}`,
		`{"choices":[{"message":{"role":"assistant","content":"a"}}]} {}`,
		"{\"choices\":[{\"message\":{\"role\":\"assistant\",\"content\":\"\xff\"}}]}",
		strings.Repeat(" ", 8<<20) + `{"choices":[]}`,
	} {
		if _, err := upstream.ReadCompletion(completion(body, "application/json")); !errors.Is(err, upstream.ErrProtocol) {
			t.Fatalf("invalid completion did not produce protocol error: %v", err)
		}
	}
	if _, err := upstream.ReadCompletion(completion(`{}`, "text/html")); !errors.Is(err, upstream.ErrProtocol) {
		t.Fatal("wrong content type accepted", err)
	}
}

func TestUnknownUsageIsNotInvented(t *testing.T) {
	usage := upstream.ReadErrorUsage(strings.NewReader(`{"usage":{"prompt_tokens":0,"completion_tokens":null,"total_tokens":-2}}`))
	if usage.Prompt == nil || *usage.Prompt != 0 || usage.Completion != nil || usage.Total != nil {
		t.Fatal("unknown usage fabricated")
	}
	for _, body := range []string{`{"usage":{"prompt_tokens":1,"prompt_tokens":2}}`, `{"usage":{"prompt_tokens":1}} {}`, "{\"usage\":{\"prompt_tokens\":1},\"bad\":\"\xff\"}", strings.Repeat(" ", 1<<20) + `{}`} {
		usage := upstream.ReadErrorUsage(strings.NewReader(body))
		if usage.Prompt != nil || usage.Completion != nil || usage.Total != nil {
			t.Fatal("invalid error body yielded usage")
		}
	}
}

func TestStreamFramesAndTerminal(t *testing.T) {
	data := ": initial\r\n\r\n: discarded\r\ndata: {\"choices\":[{\"delta\":{\"role\":\"assistant\"}}]}\r\n\r\n: keepalive\n\ndata:{\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\ndata: {\"choices\":[],\ndata: \"usage\":{\"prompt_tokens\":3,\"total_tokens\":5}}\n\ndata: [DONE]\n\ndata: invalid-after-terminal\n\n"
	stream := upstream.NewStream(strings.NewReader(data))
	for index, want := range []struct {
		kind          upstream.EventKind
		wire          string
		output, usage bool
	}{
		{upstream.EventComment, ": initial\n\n", false, false},
		{upstream.EventData, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\"}}]}\n\n", false, false},
		{upstream.EventComment, ": keepalive\n\n", false, false},
		{upstream.EventData, "data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\n", true, false},
		{upstream.EventData, "data: {\"choices\":[],\ndata: \"usage\":{\"prompt_tokens\":3,\"total_tokens\":5}}\n\n", false, true},
		{upstream.EventDone, "data: [DONE]\n\n", false, false},
	} {
		event, err := stream.Next()
		if err != nil || event.Kind != want.kind || event.Wire != want.wire || event.EffectiveOutput != want.output || event.HasUsage != want.usage {
			t.Fatalf("frame %d changed: %#v, %v", index, event, err)
		}
		if event.HasUsage && (event.Usage.Prompt == nil || *event.Usage.Prompt != 3 || event.Usage.Completion != nil || event.Usage.Total == nil || *event.Usage.Total != 5) {
			t.Fatal("stream usage changed")
		}
	}
	if _, err := stream.Next(); err != io.EOF {
		t.Fatal("data after DONE was consumed", err)
	}
}

func TestStreamProtocolStopBoundaries(t *testing.T) {
	const chunk = "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\n"
	for name, body := range map[string]string{
		"empty": "", "done_before_data": "data: [DONE]\n\n", "unterminated": chunk,
		"overlong":       "data: " + strings.Repeat("x", 256*1024) + "\n\n",
		"oversize_event": strings.Repeat(": "+strings.Repeat("x", 1024)+"\n", 256) + "\n",
		"invalid_field":  "event: secret\n\n", "invalid_json": "data: {\n\n",
		"duplicate_json": "data: {\"choices\":[],\"usage\":{\"total_tokens\":1,\"total_tokens\":2}}\n\n",
		"provider_error": "data: {\"error\":{\"message\":\"secret\"}}\n\n",
		"invalid_utf8":   ": \xff\n\n", "usage_required": "data: {\"choices\":[]}\n\n",
		"unclosed_frame": "data: {\"choices\":[],\"usage\":{}}\n",
	} {
		t.Run(name, func(t *testing.T) {
			stream := upstream.NewStream(strings.NewReader(body))
			for {
				_, err := stream.Next()
				if err == nil {
					continue
				}
				if !errors.Is(err, upstream.ErrProtocol) {
					t.Fatal("protocol boundary lost", err)
				}
				break
			}
		})
	}
}

func TestClientLatencyOutputDeltaContract(t *testing.T) {
	for _, test := range []struct {
		delta     string
		effective bool
	}{
		{`{"role":"assistant"}`, false}, {`{}`, false}, {`{"content":""}`, false}, {`{"content":null}`, false}, {`{"content":" "}`, true}, {`{"refusal":"no"}`, true},
		{`{"tool_calls":[{"index":0,"id":"id","type":"function","function":{"name":"","arguments":""}}]}`, false},
		{`{"tool_calls":[{"index":0,"function":{"arguments":"{"}}]}`, true}, {`{"tool_calls":[{"function":{"name":"f"}}]}`, true},
	} {
		event, err := upstream.NewStream(strings.NewReader("data: {\"choices\":[{\"delta\":" + test.delta + "}]}\n\n")).Next()
		if err != nil || event.EffectiveOutput != test.effective {
			t.Fatal("delta classification", test.delta, err)
		}
	}
	for _, raw := range []string{`{"choices":[],"usage":{}}`, `{"choices":[{"delta":{},"finish_reason":"stop"}]}`} {
		event, err := upstream.NewStream(strings.NewReader("data: " + raw + "\n\n")).Next()
		if err != nil || event.EffectiveOutput {
			t.Fatal("metadata scored", err)
		}
	}
}

type failingReader struct{ err error }

func (reader failingReader) Read([]byte) (int, error) { return 0, reader.err }

func TestReadErrorsRemainTransportErrors(t *testing.T) {
	want := errors.New("controlled connection reset")
	response := completion("", "application/json")
	response.Body = io.NopCloser(failingReader{want})
	if _, err := upstream.ReadCompletion(response); !errors.Is(err, want) {
		t.Fatal("completion read error lost", err)
	}
	if _, err := upstream.NewStream(failingReader{want}).Next(); !errors.Is(err, want) {
		t.Fatal("stream read error lost", err)
	}
}

func TestStreamBoundsTotalBytesAcrossFrames(t *testing.T) {
	// Each frame is 64 KiB. Individual frames are valid; the 513th exceeds the
	// 32 MiB response budget even though every preceding delimiter reset framing.
	frame := ": " + strings.Repeat("x", 65532) + "\n\n"
	readers := make([]io.Reader, 513)
	for index := range readers {
		readers[index] = strings.NewReader(frame)
	}
	stream := upstream.NewStream(io.MultiReader(readers...))
	for index := 0; index < 512; index++ {
		if event, err := stream.Next(); err != nil || event.Kind != upstream.EventComment {
			t.Fatal("valid frame rejected before response byte limit", index, err)
		}
	}
	if _, err := stream.Next(); !errors.Is(err, upstream.ErrProtocol) {
		t.Fatal("response byte limit reset between frames", err)
	}
}
