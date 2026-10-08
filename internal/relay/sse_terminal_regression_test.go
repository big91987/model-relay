package relay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const terminalChunk = "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\n"
const terminalRequest = `{"model":"alpha","messages":[{"role":"user","content":"controlled"}],"stream":true}`

type terminalWriter struct {
	*httptest.ResponseRecorder
	onDone func() error
}

func TestSSERealHTTPOverlongLineNeverSwitchesOrCools(testContext *testing.T) {
	for _, committed := range []bool{false, true} {
		testContext.Run(map[bool]string{false: "before_commit", true: "after_commit"}[committed], func(testContext *testing.T) {
			var primary, secondary atomic.Int64
			value := newFixture(testContext, func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "text/event-stream")
				if strings.HasPrefix(request.URL.Path, "/secondary/") {
					secondary.Add(1)
					io.WriteString(writer, terminalChunk+"data: [DONE]\n\n")
					return
				}
				primary.Add(1)
				if committed {
					io.WriteString(writer, terminalChunk)
					writer.(http.Flusher).Flush()
				}
				io.WriteString(writer, "data: "+strings.Repeat("x", 256*1024)+"\n\n")
			})
			configureHTTPPair(testContext, value, 1)
			response := value.call(testContext, streaming)
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
			var record Record
			decode(testContext, value.admin(testContext, "GET", "requests/"+response.Header.Get("X-Request-ID"), nil), &record)
			wantStatus := 502
			if committed {
				wantStatus = 200
			}
			if response.StatusCode != wantStatus || primary.Load() != 1 || secondary.Load() != 0 || len(record.Attempts) != 1 || record.Code != "upstream_protocol_error" || record.State != "stream_interrupted" || record.SuccessfulUpstream != nil || (record.CommittedAt != nil) != committed {
				testContext.Fatal("real HTTP protocol stop boundary", response.StatusCode, record.State, record.Code, primary.Load(), secondary.Load())
			}
			if record.Attempts[0].Code != record.Code || record.Attempts[0].State != record.State {
				testContext.Fatal("real HTTP attempt/request mismatch")
			}
			value.server.mu.RLock()
			defer value.server.mu.RUnlock()
			for _, health := range value.server.health {
				if !health.Cooldown.IsZero() {
					testContext.Fatal("real HTTP protocol error cooled candidate")
				}
			}
		})
	}
}

func (writer *terminalWriter) Write(data []byte) (int, error) {
	if strings.Contains(string(data), "data: [DONE]") && writer.onDone != nil {
		if err := writer.onDone(); err != nil {
			return 0, err
		}
	}
	return writer.ResponseRecorder.Write(data)
}

func terminalDetails(testContext *testing.T, value *consoleFixture, id string) Record {
	testContext.Helper()
	response := value.admin("GET", "requests/"+id, nil)
	var record Record
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &record) != nil {
		testContext.Fatal("persisted detail unavailable", response.Code)
	}
	return record
}

func TestSSEOverlongLineNeverSwitchesOrCools(testContext *testing.T) {
	for _, committed := range []bool{false, true} {
		testContext.Run(map[bool]string{false: "before_commit", true: "after_commit"}[committed], func(testContext *testing.T) {
			value := newConsoleFixture(testContext)
			multiConfiguration(testContext, value, 2, nil)
			calls := 0
			value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
				calls++
				body := "data: " + strings.Repeat("x", 256*1024) + "\n\n"
				if committed {
					body = terminalChunk + body
				}
				if calls > 1 {
					body = terminalChunk + "data: [DONE]\n\n"
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			response := value.call(value.key, terminalRequest)
			record := terminalDetails(testContext, value, response.Header().Get("X-Request-ID"))
			if calls != 1 || record.State != "stream_interrupted" || record.Code != "upstream_protocol_error" || record.SuccessfulUpstream != nil || len(record.Attempts) != 1 {
				testContext.Fatalf("protocol boundary calls=%d state=%s code=%s success=%v attempts=%d", calls, record.State, record.Code, record.SuccessfulUpstream, len(record.Attempts))
			}
			attempt := record.Attempts[0]
			if attempt.State != record.State || attempt.Code != record.Code || (record.CommittedAt != nil) != committed {
				testContext.Fatal("attempt/commit differs from request terminal")
			}
			wantStatus := 502
			if committed {
				wantStatus = 200
			}
			if response.Code != wantStatus || string(value.overview(testContext)["call_evidence"]) != "null" {
				testContext.Fatal("HTTP/overview protocol terminal mismatch")
			}
			for _, health := range value.server.health {
				if !health.Cooldown.IsZero() {
					testContext.Fatal("protocol error cooled candidate")
				}
			}
		})
	}
}

func TestSSEDoneTerminalAttribution(testContext *testing.T) {
	for _, mode := range []string{"success", "write_failure", "cancel_with_error", "cancel_after_write"} {
		testContext.Run(mode, func(testContext *testing.T) {
			testContext.Parallel()
			value := newConsoleFixture(testContext)
			multiConfiguration(testContext, value, 2, nil)
			calls := 0
			value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(terminalChunk + "data: [DONE]\n\n"))}, nil
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			request := httptest.NewRequest("POST", "http://console.test/v1/chat/completions", strings.NewReader(terminalRequest)).WithContext(ctx)
			request.Header.Set("Authorization", "Bearer "+value.key)
			request.Header.Set("Content-Type", "application/json")
			writer := &terminalWriter{ResponseRecorder: httptest.NewRecorder(), onDone: func() error {
				if strings.HasPrefix(mode, "cancel_") {
					cancel()
				}
				if mode == "write_failure" || mode == "cancel_with_error" {
					return errors.New("controlled downstream failure")
				}
				return nil
			}}
			value.server.ServeHTTP(writer, request)
			record := terminalDetails(testContext, value, writer.Header().Get("X-Request-ID"))
			wantState, wantCode := "success", ""
			if mode != "success" {
				wantState, wantCode = "canceled", "downstream_write_failed"
				if strings.HasPrefix(mode, "cancel_") {
					wantCode = "client_canceled"
				}
			}
			if calls != 1 || record.State != wantState || record.Code != wantCode || len(record.Attempts) != 1 || record.CommittedAt == nil {
				testContext.Fatalf("DONE boundary calls=%d state=%s code=%s attempts=%d", calls, record.State, record.Code, len(record.Attempts))
			}
			attempt := record.Attempts[0]
			if attempt.State != wantState || attempt.Code != wantCode {
				testContext.Fatal("attempt and request disagree")
			}
			evidence := string(value.overview(testContext)["call_evidence"])
			if mode == "success" {
				if record.SuccessfulUpstream == nil || *record.SuccessfulUpstream != attempt.UpstreamID || !strings.Contains(evidence, record.ID) {
					testContext.Fatal("success attribution/evidence missing")
				}
			} else if record.SuccessfulUpstream != nil || evidence != "null" {
				testContext.Fatal("non-success retained successful upstream or overview evidence")
			}
		})
	}
}
