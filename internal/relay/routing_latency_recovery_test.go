package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLatencyRealHTTPFailureWatermark(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, failureStream := range []bool{false, true} {
			for _, adopted := range []bool{false, true} {
				t.Run(fmt.Sprintf("stream=%t/failure_stream=%t/prior_success=%t", stream, failureStream, adopted), func(t *testing.T) {
					var dispatched atomic.Int64
					var clock atomic.Int64
					clock.Store(time.Now().UnixNano())
					oldEntered := make(chan struct{})
					failureEntered := make(chan struct{})
					newerEntered := make(chan struct{})
					oldRelease := make(chan struct{})
					failureRelease := make(chan struct{})
					newerRelease := make(chan struct{})
					var oldOnce, failureOnce, newerOnce sync.Once
					failureOrdinal := int64(2)
					if adopted {
						failureOrdinal++
					}
					value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
						ordinal := dispatched.Add(1)
						if ordinal == failureOrdinal {
							close(failureEntered)
							select {
							case <-failureRelease:
							case <-request.Context().Done():
								return
							}
							writer.WriteHeader(503)
							return
						}
						var input struct {
							Stream bool `json:"stream"`
						}
						if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
							t.Error("invalid controlled request")
							return
						}
						isStream := input.Stream
						if isStream {
							writer.Header().Set("Content-Type", "text/event-stream")
							fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\n")
						} else {
							writer.Header().Set("Content-Type", "application/json")
							fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]`)
						}
						writer.(http.Flusher).Flush()
						var release <-chan struct{}
						switch ordinal {
						case 1:
							close(oldEntered)
							release = oldRelease
						case failureOrdinal + 1:
							close(newerEntered)
							release = newerRelease
						}
						if release != nil {
							select {
							case <-release:
							case <-request.Context().Done():
								return
							}
						}
						if isStream {
							fmt.Fprint(writer, "data: [DONE]\n\n")
						} else {
							fmt.Fprint(writer, `}`)
						}
					})
					defer oldOnce.Do(func() { close(oldRelease) })
					defer failureOnce.Do(func() { close(failureRelease) })
					defer newerOnce.Do(func() { close(newerRelease) })
					value.server.Config.TotalTimeout = 10 * time.Second
					value.server.latencyClock = func() time.Time { return time.Unix(0, clock.Load()) }
					setLatencyStrategy(t, value, "low_latency")
					call := func(mode bool) string {
						body := ordinary
						if mode {
							body = streaming
						}
						response := value.call(t, body)
						io.Copy(io.Discard, response.Body)
						response.Body.Close()
						return response.Header.Get("X-Request-ID")
					}
					start := func(mode bool) <-chan string {
						completed := make(chan string, 1)
						go func() { completed <- call(mode) }()
						return completed
					}
					wait := func(entered <-chan struct{}) {
						select {
						case <-entered:
						case <-time.After(3 * time.Second):
							t.Fatal("controlled upstream barrier timed out")
						}
					}
					check := func(id, acceptance string) {
						deadline := time.Now().Add(3 * time.Second)
						for time.Now().Before(deadline) {
							_, observations, err := readRoutingEvidence(value.store.DB, id)
							if err != nil {
								t.Fatal(err)
							}
							if len(observations) == 1 {
								if observations[0].Acceptance != acceptance {
									t.Fatalf("request_id=%s acceptance=%s want=%s", id, observations[0].Acceptance, acceptance)
								}
								t.Logf("request_id=%s acceptance=%s", id, acceptance)
								return
							}
							time.Sleep(time.Millisecond)
						}
						t.Fatal("missing persisted latency evidence")
					}
					checkSamples := func(expected int) {
						value.server.mu.Lock()
						defer value.server.mu.Unlock()
						count := 0
						for _, pool := range value.server.latency {
							count += len(pool.Samples)
						}
						if count != expected {
							t.Fatalf("sample_count=%d want=%d", count, expected)
						}
					}
					older := start(stream)
					wait(oldEntered)
					if adopted {
						check(call(stream), "accepted")
					}
					failure := start(failureStream)
					wait(failureEntered)
					newer := start(stream)
					wait(newerEntered)
					failureOnce.Do(func() { close(failureRelease) })
					check(<-failure, "not_success")
					checkSamples(0)
					newerOnce.Do(func() { close(newerRelease) })
					check(<-newer, "invalidated_by_failure")
					checkSamples(0)
					before := dispatched.Load()
					call(stream)
					if dispatched.Load() != before {
						t.Fatal("pre-failure in-flight success bypassed cooldown")
					}
					clock.Add(int64(value.server.Config.Cooldown) + 1)
					check(call(stream), "accepted")
					checkSamples(1)
					oldOnce.Do(func() { close(oldRelease) })
					check(<-older, "invalidated_by_failure")
					checkSamples(1)
					check(call(!stream), "accepted")
					checkSamples(2)
				})
			}
		}
	}
}
