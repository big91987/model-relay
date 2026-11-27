package upstream_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"model-relay/internal/upstream"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type fixedResolver struct{ addresses []netip.Addr }

func (resolver fixedResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return resolver.addresses, nil
}

func TestDestinationPolicy(t *testing.T) {
	policy := upstream.Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}
	for _, raw := range []string{"http://example.com/v1", "https://u:p@example.com", "https://example.com/?x=1", "https://example.com/#x", "https://example.com/a/../b", "https://example.com/a%2fb", "https://127.0.0.1/v1", "https://10.0.0.1", "https://[::ffff:127.0.0.1]", "https://100.64.0.1", "https://example.com:0", "https://example.com:65536", "https://example.com?", "https://example.com/a//b", "https://example.com/a\\b"} {
		if _, err := policy.URL(context.Background(), raw); err == nil {
			t.Fatalf("unsafe URL accepted %s", raw)
		}
	}
	if normalized, err := policy.URL(context.Background(), "https://public.example/v1/"); err != nil || normalized != "https://public.example/v1" {
		t.Fatal("safe URL rejected", normalized, err)
	}
	policy.Resolver = fixedResolver{[]netip.Addr{netip.MustParseAddr("127.0.0.1")}}
	if _, err := policy.Dial(context.Background(), "tcp", "public.example:443"); !errors.Is(err, upstream.ErrDestinationForbidden) {
		t.Fatal("rebound private DNS accepted", err)
	}
	policy.Resolver = fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("10.0.0.1")}}
	if _, err := policy.URL(context.Background(), "https://public.example"); !errors.Is(err, upstream.ErrDestinationForbidden) {
		t.Fatal("mixed private DNS accepted", err)
	}
	for _, ip := range []string{"169.254.169.254", "192.0.2.1", "198.18.0.1", "203.0.113.1", "::1", "2001:db8::1", "64:ff9b::127.0.0.1"} {
		policy.Resolver = fixedResolver{[]netip.Addr{netip.MustParseAddr(ip)}}
		if _, err := policy.URL(context.Background(), "https://public.example"); !errors.Is(err, upstream.ErrDestinationForbidden) {
			t.Fatal("reserved address accepted", ip, err)
		}
	}
}

func TestPrivateAuthorityIsExactAndExcludesLoopback(t *testing.T) {
	policy := upstream.Policy{PrivateAuthorities: []string{"private.example:8080"}, Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("10.0.0.5")}}}
	if _, err := policy.URL(context.Background(), "http://private.example:8080/v1"); err != nil {
		t.Fatal(err)
	}
	if _, err := policy.URL(context.Background(), "http://private.example:8081/v1"); err == nil {
		t.Fatal("private authority widened to another port")
	}
	policy.Resolver = fixedResolver{[]netip.Addr{netip.MustParseAddr("127.0.0.1")}}
	if _, err := policy.URL(context.Background(), "http://private.example:8080/v1"); !errors.Is(err, upstream.ErrDestinationForbidden) {
		t.Fatal("private authority allowed loopback", err)
	}
}

func TestChatOwnsWireRequestAndDoesNotMutateSnapshot(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary", true: "stream"}[stream], func(t *testing.T) {
			var calls atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				calls.Add(1)
				wantAccept := "application/json"
				if stream {
					wantAccept = "text/event-stream"
				}
				if request.Method != "POST" || request.URL.Path != "/v1/chat/completions" || request.Header.Get("Authorization") != "Bearer controlled-secret" || request.Header.Get("Content-Type") != "application/json" || request.Header.Get("Accept") != wantAccept || request.Header.Get("X-Request-ID") != "request-id" || request.Header.Get("Cookie") != "" || request.Header.Get("Idempotency-Key") != "" {
					t.Error("upstream wire request changed")
				}
				var payload map[string]json.RawMessage
				if json.NewDecoder(request.Body).Decode(&payload) != nil || string(payload["model"]) != `"remote"` || string(payload["messages"]) != `[{"role":"user","content":"controlled"}]` {
					t.Error("request model/body changed")
				}
				writer.Header().Set("Content-Type", "application/json")
				io.WriteString(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`)
			}))
			defer server.Close()
			policy := upstream.Policy{TestAuthority: strings.TrimPrefix(server.URL, "http://")}
			client := upstream.NewClient(upstream.NewHTTPClient(policy, time.Second))
			payload, _, _, _ := upstream.ValidateRequest([]byte(ordinary))
			prepared, err := upstream.PrepareChat(context.Background(), upstream.ChatInput{Endpoint: upstream.Endpoint{URL: server.URL + "/v1", Token: "controlled-secret"}, Payload: payload, Model: "remote", RequestID: "request-id", Stream: stream})

			if err != nil {
				t.Fatal(err)
			}

			response, err := client.Send(prepared)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != 200 || calls.Load() != 1 || string(payload["model"]) != `"alpha"` {
				t.Fatal("snapshot mutation or replay")
			}
		})
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (trip roundTrip) RoundTrip(request *http.Request) (*http.Response, error) { return trip(request) }

func TestChatDisablesAutomaticBodyReplay(t *testing.T) {
	client := upstream.NewClient(&http.Client{Transport: roundTrip(func(request *http.Request) (*http.Response, error) {
		if request.GetBody != nil {
			t.Error("automatic request replay enabled")
		}
		return completion(`{}`, "application/json"), nil
	})})
	payload, _, _, _ := upstream.ValidateRequest([]byte(ordinary))
	prepared, err := upstream.PrepareChat(context.Background(), upstream.ChatInput{Endpoint: upstream.Endpoint{URL: "https://public.example", Token: "token"}, Payload: payload, Model: "remote"})

	if err != nil {
		t.Fatal(err)
	}

	response, err := client.Send(prepared)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
}

func TestSafeClientDoesNotFollowRedirectOrEnvironmentProxy(t *testing.T) {
	var redirects, proxyCalls atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { redirects.Add(1) }))
	defer target.Close()
	proxy := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { proxyCalls.Add(1) }))
	defer proxy.Close()
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")
	source := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	// A named destination matters: Go bypasses environment proxies for a literal
	// loopback host even when a transport accidentally enables ProxyFromEnvironment.
	authority := "provider.example:" + source.URL[strings.LastIndex(source.URL, ":")+1:]
	client := upstream.NewClient(upstream.NewHTTPClient(upstream.Policy{TestAuthority: authority, Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("127.0.0.1")}}}, time.Second))
	payload, _, _, _ := upstream.ValidateRequest([]byte(ordinary))
	prepared, err := upstream.PrepareChat(context.Background(), upstream.ChatInput{Endpoint: upstream.Endpoint{URL: "http://" + authority, Token: "secret"}, Payload: payload, Model: "remote"})

	if err != nil {
		t.Fatal(err)
	}

	response, err := client.Send(prepared)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 307 || redirects.Load() != 0 || proxyCalls.Load() != 0 {
		t.Fatal("redirect/proxy leaked credentials")
	}
}

func TestHTTPHeaderBudgetCancelsProvider(t *testing.T) {
	canceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		io.Copy(io.Discard, request.Body)
		<-request.Context().Done()
		close(canceled)
	}))
	defer server.Close()
	client := upstream.NewClient(upstream.NewHTTPClient(upstream.Policy{TestAuthority: strings.TrimPrefix(server.URL, "http://")}, 40*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	started := time.Now()
	prepared, err := upstream.PrepareChat(ctx, upstream.ChatInput{Endpoint: upstream.Endpoint{URL: server.URL}, Payload: map[string]json.RawMessage{}, Model: "remote"})

	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Send(prepared)
	if err == nil || ctx.Err() != nil || time.Since(started) > time.Second {
		t.Fatal("provider header timeout was not enforced", err)
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("provider request survived header timeout")
	}
}

func TestCheckOwnsModelListingProtocol(t *testing.T) {
	for _, scenario := range []struct {
		name                    string
		status                  int
		contentType, body, code string
	}{
		{"valid_empty_list", 200, "application/json", `{"object":"list","data":[]}`, ""},
		{"valid_models", 201, "application/json; charset=utf-8", `{"object":"list","data":[{"id":"model"}]}`, ""},
		{"wrong_content_type", 200, "text/html", `{"object":"list","data":[]}`, "upstream_protocol_error"},
		{"missing_list_object", 200, "application/json", `{"data":[]}`, "upstream_protocol_error"},
		{"wrong_list_object", 200, "application/json", `{"object":"chat.completion","data":[]}`, "upstream_protocol_error"},
		{"null_list", 200, "application/json", `{"object":"list","data":null}`, "upstream_protocol_error"},
		{"missing_model_id", 200, "application/json", `{"object":"list","data":[{}]}`, "upstream_protocol_error"},
		{"blank_model_id", 200, "application/json", `{"object":"list","data":[{"id":" "}]}`, "upstream_protocol_error"},
		{"malformed_json", 200, "application/json", `{`, "upstream_protocol_error"},
		{"oversize", 200, "application/json", strings.Repeat(" ", 1<<20) + `{}`, "upstream_protocol_error"},
		{"rejected", 401, "application/json", `{}`, "upstream_rejected"},
		{"forbidden", 403, "application/json", `{}`, "upstream_rejected"},
		{"rate_limited", 429, "application/json", `{}`, "upstream_http_error"},
		{"server_failure", 503, "application/json", `{}`, "upstream_http_error"},
		// The original listing endpoint deliberately uses encoding/json decoding;
		// its acceptance of duplicate listing metadata is preserved by extraction.
		{"existing_duplicate_behavior", 200, "application/json", `{"object":"ignored","object":"list","data":[]}`, ""},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			client := upstream.NewClient(&http.Client{Transport: roundTrip(func(request *http.Request) (*http.Response, error) {
				if request.Method != "GET" || request.URL.String() != "https://public.example/v1/models" || request.Body != nil || request.Header.Get("Authorization") != "Bearer secret" {
					t.Error("connection check generated or changed destination")
				}
				response := completion(scenario.body, scenario.contentType)
				response.StatusCode = scenario.status
				return response, nil
			})})
			result := client.Check(context.Background(), upstream.Endpoint{URL: "https://public.example/v1", Token: "secret"})
			if result.Code != scenario.code || result.Status == nil || *result.Status != scenario.status {
				t.Fatal("check result changed", result)
			}
		})
	}
}

func TestCheckReportsNetworkFailureWithoutInventedStatus(t *testing.T) {
	client := upstream.NewClient(&http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) { return nil, errors.New("controlled network failure") })})
	result := client.Check(context.Background(), upstream.Endpoint{URL: "https://public.example", Token: "token"})
	if result.Code != "upstream_connection_failed" || result.Status != nil {
		t.Fatal("network result changed", result)
	}
}

func TestClientPreservesCallerCancellation(t *testing.T) {
	entered := make(chan struct{})
	client := upstream.NewClient(&http.Client{Transport: roundTrip(func(request *http.Request) (*http.Response, error) {
		close(entered)
		<-request.Context().Done()
		return nil, request.Context().Err()
	})})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		prepared, err := upstream.PrepareChat(ctx, upstream.ChatInput{Endpoint: upstream.Endpoint{URL: "https://public.example"}, Payload: map[string]json.RawMessage{}, Model: "remote"})
		if err != nil {
			done <- err
			return
		}
		_, err = client.Send(prepared)
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request not dispatched")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("cancellation lost", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not propagate")
	}
}

func TestPreparedChatFreezesBodyBeforeAnyDispatch(t *testing.T) {
	var calls int
	client := upstream.NewClient(&http.Client{Transport: roundTrip(func(request *http.Request) (*http.Response, error) {
		calls++
		data, err := io.ReadAll(request.Body)
		var value map[string]json.RawMessage
		if err != nil || json.Unmarshal(data, &value) != nil || string(value["model"]) != `"remote"` || string(value["messages"]) != `[{"role":"user","content":"controlled"}]` {
			t.Error("prepared body changed after caller mutation", string(data), err)
		}
		return completion(`{}`, "application/json"), nil
	})})
	payload, _, _, _ := upstream.ValidateRequest([]byte(ordinary))
	prepared, err := upstream.PrepareChat(context.Background(), upstream.ChatInput{Endpoint: upstream.Endpoint{URL: "https://public.example"}, Payload: payload, Model: "remote"})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("preparation dispatched the request")
	}
	payload["messages"] = json.RawMessage(`[]`)
	payload["model"] = json.RawMessage(`"changed"`)
	response, err := client.Send(prepared)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if calls != 1 {
		t.Fatal("prepared request dispatched more than once")
	}
}

func TestPreparationAttributesBodyAndDestinationFailures(t *testing.T) {
	for _, scenario := range []struct {
		url     string
		payload map[string]json.RawMessage
		code    string
	}{
		{"https://public.example", map[string]json.RawMessage{"messages": json.RawMessage(`{`)}, "invalid_request"},
		{"http://[", map[string]json.RawMessage{}, "upstream_connection_failed"},
	} {
		_, err := upstream.PrepareChat(context.Background(), upstream.ChatInput{Endpoint: upstream.Endpoint{URL: scenario.url}, Payload: scenario.payload, Model: "remote"})
		var failure *upstream.PreparationError
		if !errors.As(err, &failure) || failure.Code != scenario.code {
			t.Fatal("pre-dispatch failure attribution changed", err)
		}
	}
}
