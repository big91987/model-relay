// Package upstream owns provider HTTP transport and the supported chat protocol.
// Request deadlines, retries, persistence and downstream commitment belong to its caller.
package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

type Endpoint struct {
	URL   string
	Token string
}

type ChatInput struct {
	Endpoint  Endpoint
	Payload   map[string]json.RawMessage
	Model     string
	RequestID string
	Stream    bool
}

type Client struct {
	http *http.Client
}

// NewHTTPClient constructs the production transport. Both URL validation and
// every actual connection use Policy; environment proxies and redirects are disabled.
func NewHTTPClient(policy Policy, headerTimeout time.Duration) *http.Client {
	transport := &http.Transport{Proxy: nil, DialContext: policy.Dial, DisableKeepAlives: true, ForceAttemptHTTP2: false, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: headerTimeout}
	return &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

// NewClient accepts an HTTP dependency so controlled transports can exercise the
// same provider operations. Production callers use NewHTTPClient.
func NewClient(client *http.Client) *Client { return &Client{http: client} }

// PreparedChat holds the completed wire request before durable dispatch is
// marked possible. Its transport request remains private to this adapter.
type PreparedChat struct {
	request *http.Request
}

// PreparationError preserves whether preparation rejected the JSON body or the
// destination, so the gateway can retain its pre-dispatch failure attribution.
type PreparationError struct {
	Code string
	err  error
}

func (failure *PreparationError) Error() string { return failure.err.Error() }
func (failure *PreparationError) Unwrap() error { return failure.err }

func PrepareChat(ctx context.Context, input ChatInput) (PreparedChat, error) {
	payload := make(map[string]json.RawMessage, len(input.Payload)+1)
	for name, value := range input.Payload {
		payload[name] = value
	}
	payload["model"], _ = json.Marshal(input.Model)
	body, err := json.Marshal(payload)
	if err != nil {
		return PreparedChat{}, &PreparationError{Code: "invalid_request", err: err}
	}
	outgoing, err := http.NewRequestWithContext(ctx, http.MethodPost, input.Endpoint.URL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return PreparedChat{}, &PreparationError{Code: "upstream_connection_failed", err: err}
	}
	outgoing.GetBody = nil
	outgoing.Header.Set("Content-Type", "application/json")
	outgoing.Header.Set("Authorization", "Bearer "+input.Endpoint.Token)
	outgoing.Header.Set("X-Request-ID", input.RequestID)
	outgoing.Header.Set("Accept", "application/json")
	if input.Stream {
		outgoing.Header.Set("Accept", "text/event-stream")
	}
	return PreparedChat{request: outgoing}, nil
}

// Send performs the prepared attempt once. The caller owns its response body,
// deadlines and retry decisions; no incoming headers or replay body are copied.
func (client *Client) Send(prepared PreparedChat) (*http.Response, error) {
	return client.http.Do(prepared.request)
}

type CheckResult struct {
	Status *int
	Code   string
}

// Check performs the non-generating model-list probe. Its caller supplies the
// entire time budget and attributes cancellation after it returns.
func (client *Client) Check(ctx context.Context, endpoint Endpoint) CheckResult {
	result := CheckResult{Code: "upstream_connection_failed"}
	outgoing, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.URL+"/models", nil)
	if err != nil {
		return result
	}
	outgoing.Header.Set("Authorization", "Bearer "+endpoint.Token)
	response, err := client.http.Do(outgoing)
	if err != nil {
		return result
	}
	defer response.Body.Close()
	result.Status = &response.StatusCode
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		data, readErr := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
		var listing struct {
			Object string `json:"object"`
			Data   []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		mediaType, _, typeErr := mime.ParseMediaType(response.Header.Get("Content-Type"))
		valid := typeErr == nil && mediaType == "application/json" && readErr == nil && len(data) <= 1<<20 && json.Unmarshal(data, &listing) == nil && listing.Object == "list" && listing.Data != nil
		for _, model := range listing.Data {
			valid = valid && strings.TrimSpace(model.ID) != ""
		}
		result.Code = "upstream_protocol_error"
		if valid {
			result.Code = ""
		}
	} else if response.StatusCode == 401 || response.StatusCode == 403 {
		result.Code = "upstream_rejected"
	} else {
		result.Code = "upstream_http_error"
	}
	return result
}
