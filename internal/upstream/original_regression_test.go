package upstream

import (
	"context"
	"encoding/json"
	"net/netip"
	"testing"
)

func TestDestinationAndProtocol(t *testing.T) {
	const ordinary = `{"model":"alpha","messages":[{"role":"user","content":"prompt-sentinel"}]}`
	policy := Policy{Resolver: originalResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}
	for _, raw := range []string{"http://example.com/v1", "https://u:p@example.com", "https://example.com/?x=1", "https://example.com/#x", "https://example.com/a/../b", "https://example.com/a%2fb", "https://127.0.0.1/v1", "https://10.0.0.1", "https://[::ffff:127.0.0.1]", "https://100.64.0.1"} {
		if _, err := policy.URL(context.Background(), raw); err == nil {
			t.Fatalf("unsafe URL accepted %s", raw)
		}
	}
	if _, err := policy.URL(context.Background(), "https://public.example/v1"); err != nil {
		t.Fatal(err)
	}
	policy.Resolver = originalResolver{[]netip.Addr{netip.MustParseAddr("127.0.0.1")}}
	if _, err := policy.Dial(context.Background(), "tcp", "public.example:443"); err == nil {
		t.Fatal("rebound private DNS accepted")
	}
	policy.Resolver = originalResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("10.0.0.1")}}
	if _, err := policy.URL(context.Background(), "https://public.example"); err == nil {
		t.Fatal("mixed private DNS accepted")
	}
	for _, payload := range []string{`{"model":"a","model":"b","messages":[]}`, ordinary + ` {}`, `{"model":"alpha","messages":[{"role":"user","content":[{"type":"image"}]}]}`, `{"model":"alpha","messages":[{"role":"user","content":"text"}],"base_url":"http://evil"}`, `{"model":"alpha","messages":[{"role":"user","content":"text"}],"stream":"true"}`, `{"model":"alpha","messages":[{"role":"user","content":"text"}],"n":0}`} {
		if _, _, _, code := ValidateRequest([]byte(payload)); code == "" {
			t.Fatal("invalid request accepted")
		}
	}
	if _, _, _, code := ValidateRequest([]byte(ordinary)); code != "" {
		t.Fatal(code)
	}
	usage := usageFrom(map[string]json.RawMessage{"usage": json.RawMessage(`{"prompt_tokens":0,"completion_tokens":null,"total_tokens":-2}`)})
	if usage.Prompt == nil || *usage.Prompt != 0 || usage.Completion != nil || usage.Total != nil {
		t.Fatal("unknown usage fabricated")
	}
}

func TestUsageSnapshots(t *testing.T) {
	for _, raw := range []string{`{}`, `{"usage":null}`, `{"usage":{"prompt_tokens":null,"completion_tokens":-1,"total_tokens":1.5}}`} {
		var value map[string]json.RawMessage
		json.Unmarshal([]byte(raw), &value)
		usage := usageFrom(value)
		if usage.Prompt != nil || usage.Completion != nil || usage.Total != nil {
			t.Fatal("invalid usage invented")
		}
	}
}

func TestLatencyOutputDeltaContract(t *testing.T) {
	for _, test := range []struct {
		delta     string
		effective bool
	}{
		{`{"role":"assistant"}`, false}, {`{}`, false}, {`{"content":""}`, false}, {`{"content":null}`, false}, {`{"content":" "}`, true}, {`{"refusal":"no"}`, true},
		{`{"tool_calls":[{"index":0,"id":"id","type":"function","function":{"name":"","arguments":""}}]}`, false},
		{`{"tool_calls":[{"index":0,"function":{"arguments":"{"}}]}`, true}, {`{"tool_calls":[{"function":{"name":"f"}}]}`, true},
	} {
		var value map[string]json.RawMessage
		if json.Unmarshal([]byte(`{"choices":[{"delta":`+test.delta+`}]}`), &value) != nil {
			t.Fatal("fixture")
		}
		if outputDelta(value) != test.effective {
			t.Fatal("delta classification", test.effective)
		}
	}
	for _, raw := range []string{`{"choices":[],"usage":{}}`, `{"choices":[{"delta":{},"finish_reason":"stop"}]}`} {
		var value map[string]json.RawMessage
		json.Unmarshal([]byte(raw), &value)
		if outputDelta(value) {
			t.Fatal("metadata scored")
		}
	}
}

type originalResolver struct{ Addresses []netip.Addr }

func (resolver originalResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return resolver.Addresses, nil
}
