package relay

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

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
func TestLatencySamplesExpireFromMeasurement(t *testing.T) {
	origin := time.Now()
	pool := &latencyPool{}
	for index := 0; index < 3; index++ {
		pool.Samples = append(pool.Samples, latencySample{Measured: origin.Add(time.Duration(index) * time.Second), Milliseconds: float64(10 + index), Sequence: int64(index + 1)})
	}
	state, count, score, _, _ := pool.describe(origin.Add(60 * time.Second))
	if state != "ready" || count != 3 || score == nil {
		t.Fatal("60s boundary", state, count)
	}
	state, count, score, _, _ = pool.describe(origin.Add(60*time.Second + time.Nanosecond))
	if state != "insufficient" || count != 2 || score != nil {
		t.Fatal("individual expiry", state, count)
	}
	pool.Samples = append(pool.Samples, latencySample{Measured: origin.Add(61 * time.Second), Milliseconds: 100, Sequence: 4})
	_, count, _, _, _ = pool.describe(origin.Add(63 * time.Second))
	if count != 1 {
		t.Fatal("new sample renewed old observations", count)
	}
	for index := 0; index < 2; index++ {
		state, count, score, _, _ = pool.describe(origin.Add(122 * time.Second))
		if state != "expired" || count != 0 || score != nil {
			t.Fatal("expired explanation lost", state, count)
		}
	}
}
func TestLatencyExplorationBoundAndModeIsolation(t *testing.T) {
	for _, count := range []int{1, 2, 3, 100} {
		server := &Server{latency: map[string]*latencyPool{}}
		snapshot := routeSnapshot{Model: RouteModel{Alias: "alpha", Strategy: "low_latency"}, Upstreams: map[int64]Upstream{}, Runtime: &routeRuntime{Current: map[int64]int{}}}
		group := []Candidate{}
		reasons := []RouteDecision{}
		for index := 1; index <= count; index++ {
			group = append(group, Candidate{UpstreamID: int64(index)})
			reasons = append(reasons, RouteDecision{UpstreamID: int64(index)})
		}
		visits := map[int64]bool{}
		for index := 0; index < 5*count; index++ {
			candidate, evidence := server.chooseLatency(snapshot, group, false, 1, &group[0], reasons)
			for _, fact := range evidence {
				if fact.UpstreamID == candidate.UpstreamID && fact.Reason == "explore" {
					visits[candidate.UpstreamID] = true
				}
			}
		}
		if len(visits) != count {
			t.Fatal("exploration starvation", count, len(visits))
		}
		_, facts := server.chooseLatency(snapshot, group, true, 1, &group[0], reasons)
		if facts[0].Exploration != 1 {
			t.Fatal("mode counter shared")
		}
	}
}
func setLatencyStrategy(t *testing.T, value *fixture, strategy string) {
	t.Helper()
	model, err := value.store.routeModel("alpha")
	if err != nil {
		t.Fatal(err)
	}
	response := value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": cleanTestCandidates(model.Candidates), "routing_strategy": strategy})
	expect(t, response, 200)
}
func TestLatencyRealHTTPMeasurementAgeAndLateConfiguration(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, age := range []int{59, 60, 61} {
			t.Run(fmt.Sprintf("stream=%t/age=%d", stream, age), func(t *testing.T) {
				var clock atomic.Int64
				var notified sync.Once
				measured := make(chan struct{})
				release := make(chan struct{})
				defer close(release)
				value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
					if stream {
						writer.Header().Set("Content-Type", "text/event-stream")
					} else {
						writer.Header().Set("Content-Type", "application/json")
					}
					writer.WriteHeader(200)
					writer.(http.Flusher).Flush()
					clock.Store(int64(10 * time.Millisecond))
					if stream {
						fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\n")
					} else {
						fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]`)
					}
					writer.(http.Flusher).Flush()
					select {
					case <-release:
					case <-request.Context().Done():
						return
					}
					if stream {
						fmt.Fprint(writer, "data: [DONE]\n\n")
					} else {
						fmt.Fprint(writer, `}`)
					}
				})
				value.server.latencyClock = func() time.Time {
					stamp := clock.Load()
					if stamp == int64(10*time.Millisecond) {
						notified.Do(func() { close(measured) })
					}
					return time.Unix(0, stamp)
				}
				setLatencyStrategy(t, value, "low_latency")
				result := make(chan string, 1)
				go func() {
					body := ordinary
					if stream {
						body = streaming
					}
					response := value.call(t, body)
					io.Copy(io.Discard, response.Body)
					response.Body.Close()
					result <- response.Header.Get("X-Request-ID")
				}()
				select {
				case <-measured:
				case <-time.After(time.Second):
					t.Fatal("measurement not reached")
				}
				clock.Store(int64(10*time.Millisecond) + int64(time.Duration(age)*time.Second))
				release <- struct{}{}
				id := <-result
				_, facts, err := readRoutingEvidence(value.store.DB, id)
				if err != nil || len(facts) != 1 {
					t.Fatal("evidence", err, len(facts))
				}
				expected := "accepted"
				if age > 60 {
					expected = "expired"
				}
				if facts[0].Acceptance != expected || facts[0].Value == nil || *facts[0].Value != 10 {
					t.Fatalf("measurement/age result: %+v", facts[0])
				}
			})
		}
	}
	t.Run("late_configuration", func(t *testing.T) {
		arrived := make(chan struct{})
		release := make(chan struct{})
		defer close(release)
		value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
			close(arrived)
			<-release
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`)
		})
		setLatencyStrategy(t, value, "low_latency")
		done := make(chan string, 1)
		go func() {
			response := value.call(t, ordinary)
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
			done <- response.Header.Get("X-Request-ID")
		}()
		<-arrived
		setLatencyStrategy(t, value, "weighted")
		release <- struct{}{}
		id := <-done
		_, facts, err := readRoutingEvidence(value.store.DB, id)
		if err != nil || len(facts) != 1 || facts[0].Acceptance != "configuration_changed" {
			t.Fatal("late result applied", err, facts)
		}
	})
}
func TestLatencyRealHTTPHeartbeatCancellation(t *testing.T) {
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\"}}]}\n\n: heartbeat\n\n")
		writer.(http.Flusher).Flush()
		<-request.Context().Done()
	})
	configureHTTPPair(t, value, 1)
	setLatencyStrategy(t, value, "low_latency")
	ctx, cancel := context.WithCancel(context.Background())
	request, _ := http.NewRequestWithContext(ctx, "POST", value.http.URL+"/v1/chat/completions", strings.NewReader(streaming))
	request.Header.Set("Authorization", "Bearer "+value.key)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	bufio.NewReader(response.Body).ReadString('\n')
	cancel()
	response.Body.Close()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		_, facts, err := readRoutingEvidence(value.store.DB, response.Header.Get("X-Request-ID"))
		if err == nil && len(facts) > 0 {
			if facts[0].Acceptance != "not_success" || facts[0].Value != nil {
				t.Fatal("canceled role-only scored", facts)
			}
			var attempts int
			value.store.DB.QueryRow("SELECT count(*) FROM request_attempts WHERE request_id=?", response.Header.Get("X-Request-ID")).Scan(&attempts)
			if attempts != 1 {
				t.Fatal("cancellation replayed")
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("cancel evidence missing")
}
func TestLatencyPolicyMigrationAndCompatibility(t *testing.T) {
	value := newConsoleFixture(t)
	multiConfiguration(t, value, 2, nil)
	model, _ := value.server.Store.routeModel("alpha")
	if model.Strategy != "weighted" {
		t.Fatal("default")
	}
	body := map[string]any{"enabled": true, "expected_revision": model.Revision, "candidates": cleanTestCandidates(model.Candidates), "routing_strategy": "low_latency"}
	if value.admin("PATCH", "models/alpha", body).Code != 200 {
		t.Fatal("strategy save")
	}
	model, _ = value.server.Store.routeModel("alpha")
	delete(body, "routing_strategy")
	body["expected_revision"] = model.Revision
	if value.admin("PATCH", "models/alpha", body).Code != 200 {
		t.Fatal("legacy patch")
	}
	model, _ = value.server.Store.routeModel("alpha")
	if model.Strategy != "low_latency" {
		t.Fatal("legacy patch reset strategy")
	}
	body["expected_revision"] = model.Revision
	body["routing_strategy"] = "invalid"
	if value.admin("PATCH", "models/alpha", body).Code != 400 {
		t.Fatal("invalid policy accepted")
	}
	transaction, err := value.server.Store.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err = routingExtension(transaction, false); err != nil {
		t.Fatal(err)
	}
	transaction.Rollback()
}

func TestLatencyRealHTTPConcurrentGenerationAndSequence(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, change := range []string{"strategy", "disable", "remote", "credential", "candidate_recreate"} {
			t.Run(fmt.Sprintf("sixteen_inflight_configuration/stream=%t/%s", stream, change), func(t *testing.T) {
				entered := make(chan struct{}, 16)
				release := make(chan struct{})
				var once sync.Once
				defer once.Do(func() { close(release) })
				value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
					entered <- struct{}{}
					select {
					case <-release:
					case <-request.Context().Done():
						return
					}
					if stream {
						writer.Header().Set("Content-Type", "text/event-stream")
						fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\ndata: [DONE]\n\n")
					} else {
						writer.Header().Set("Content-Type", "application/json")
						fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`)
					}
				})
				value.server.Config.TotalTimeout = 10 * time.Second
				setLatencyStrategy(t, value, "low_latency")
				done := make(chan string, 16)
				for index := 0; index < 16; index++ {
					go func() {
						body := ordinary
						if stream {
							body = streaming
						}
						response := value.call(t, body)
						io.Copy(io.Discard, response.Body)
						response.Body.Close()
						done <- response.Header.Get("X-Request-ID")
					}()
				}
				for index := 0; index < 16; index++ {
					select {
					case <-entered:
					case <-time.After(5 * time.Second):
						t.Fatal("inflight barrier")
					}
				}
				model, _ := value.store.routeModel("alpha")
				switch change {
				case "strategy":
					setLatencyStrategy(t, value, "weighted")
				case "credential":
					upstream, _ := value.store.upstreamByID(1)
					expect(t, value.admin(t, "PATCH", "upstreams/1", map[string]any{"expected_revision": upstream.Revision, "name": upstream.Name, "url": upstream.URL, "enabled": true, "token": "controlled-rotated-token"}), 200)
				case "candidate_recreate":
					expect(t, value.admin(t, "POST", "upstreams", map[string]any{"name": "replacement", "url": value.upstream.URL + "/replacement/v1", "enabled": true, "token": "controlled-replacement-token"}), 201)
					replacement := cleanTestCandidates(model.Candidates)
					replacement[0]["upstream_id"] = int64(2)
					expect(t, value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": replacement}), 200)
					current, _ := value.store.routeModel("alpha")
					expect(t, value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": current.Revision, "enabled": true, "candidates": cleanTestCandidates(model.Candidates)}), 200)
				default:
					enabled := change != "disable"
					if change == "remote" {
						model.Candidates[0].RemoteModel = "controlled-changed-remote"
					}
					expect(t, value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": enabled, "candidates": cleanTestCandidates(model.Candidates)}), 200)
				}
				once.Do(func() { close(release) })
				for index := 0; index < 16; index++ {
					id := <-done
					_, facts, err := readRoutingEvidence(value.store.DB, id)
					if err != nil || len(facts) != 1 || facts[0].Acceptance != "configuration_changed" {
						t.Fatal("old generation adopted", err, facts)
					}
					t.Logf("inflight request_id=%s sample_acceptance=configuration_changed attempts=1", id)
				}
			})
		}
		t.Run(fmt.Sprintf("older_dispatch_completes_last/stream=%t", stream), func(t *testing.T) {
			entered := make(chan chan struct{}, 2)
			value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
				release := make(chan struct{})
				entered <- release
				select {
				case <-release:
				case <-request.Context().Done():
					return
				}
				if stream {
					writer.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\ndata: [DONE]\n\n")
				} else {
					writer.Header().Set("Content-Type", "application/json")
					fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`)
				}
			})
			setLatencyStrategy(t, value, "low_latency")
			first := make(chan string, 1)
			second := make(chan string, 1)
			call := func(out chan<- string) {
				body := ordinary
				if stream {
					body = streaming
				}
				response := value.call(t, body)
				io.Copy(io.Discard, response.Body)
				response.Body.Close()
				out <- response.Header.Get("X-Request-ID")
			}
			go call(first)
			firstRelease := <-entered
			go call(second)
			secondRelease := <-entered
			close(secondRelease)
			newID := <-second
			close(firstRelease)
			oldID := <-first
			for _, check := range []struct{ id, acceptance string }{{newID, "accepted"}, {oldID, "out_of_order"}} {
				_, facts, err := readRoutingEvidence(value.store.DB, check.id)
				if err != nil || len(facts) != 1 || facts[0].Acceptance != check.acceptance {
					t.Fatal("completion ordering", err, facts)
				}
				t.Logf("sequence request_id=%s sample_acceptance=%s", check.id, check.acceptance)
			}
		})
	}
}
func TestLatencyTerminalWriteNeverScoresFailure(t *testing.T) {
	for _, mode := range []string{"success", "write_failure", "cancel_after_write"} {
		t.Run(mode, func(t *testing.T) {
			value := newConsoleFixture(t)
			multiConfiguration(t, value, 2, nil)
			model, _ := value.server.Store.routeModel("alpha")
			if value.admin("PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": cleanTestCandidates(model.Candidates), "routing_strategy": "low_latency"}).Code != 200 {
				t.Fatal("policy")
			}
			value.server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(terminalChunk + "data: [DONE]\n\n"))}, nil
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			request := httptest.NewRequest("POST", "http://console.test/v1/chat/completions", strings.NewReader(terminalRequest)).WithContext(ctx)
			request.Header.Set("Authorization", "Bearer "+value.key)
			request.Header.Set("Content-Type", "application/json")
			writer := &terminalWriter{ResponseRecorder: httptest.NewRecorder(), onDone: func() error {
				if mode == "cancel_after_write" {
					cancel()
				}
				if mode == "write_failure" {
					return errors.New("controlled terminal write failure")
				}
				return nil
			}}
			value.server.ServeHTTP(writer, request)
			_, facts, err := readRoutingEvidence(value.server.Store.DB, writer.Header().Get("X-Request-ID"))
			expected := "accepted"
			if mode != "success" {
				expected = "not_success"
			}
			if err != nil || len(facts) != 1 || facts[0].Acceptance != expected {
				t.Fatal("terminal sample", err, facts)
			}
		})
	}
}
func TestLatencyMigrationOldSchemaAndCorruption(t *testing.T) {
	store, _, directory, keyFile := installation(t)
	if _, err := store.DB.Exec("DROP TABLE route_latency_evidence;DROP TABLE route_selection_evidence;DROP TABLE model_route_policies;DROP TABLE routing_extension_meta;"); err != nil {
		t.Fatal("prepare old fixture", err)
	}
	store.Close()
	reopened, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal("old format migration", err)
	}
	defer reopened.Close()
	if err = reopened.ensureRouting(); err != nil {
		t.Fatal("idempotent migration", err)
	}
	if _, err = reopened.DB.Exec("DROP TABLE route_latency_evidence"); err != nil {
		t.Fatal(err)
	}
	if err = reopened.ensureRouting(); err == nil {
		t.Fatal("partial extension silently repaired")
	}
}

func TestLatencyRealHTTPCooldownRecoverySamples(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, strategy := range []string{"weighted", "low_latency"} {
			t.Run(fmt.Sprintf("stream=%t/%s", stream, strategy), func(t *testing.T) {
				var primary atomic.Int64
				var clock atomic.Int64
				clock.Store(time.Now().UnixNano())
				value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
					isPrimary := !strings.HasPrefix(request.URL.Path, "/secondary/")
					if isPrimary && primary.Add(1) <= 6 {
						writer.WriteHeader(503)
						return
					}
					clock.Add(int64(time.Millisecond))
					if stream {
						writer.Header().Set("Content-Type", "text/event-stream")
						fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\ndata: [DONE]\n\n")
					} else {
						writer.Header().Set("Content-Type", "application/json")
						fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`)
					}
				})
				configureHTTPPair(t, value, 1)
				value.server.latencyClock = func() time.Time { return time.Unix(0, clock.Load()) }
				model, _ := value.store.routeModel("alpha")
				candidates := cleanTestCandidates(model.Candidates)
				candidates[0]["priority"] = 1
				expect(t, value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": candidates, "routing_strategy": strategy}), 200)
				recovered := false
				for index := 0; index < 30; index++ {
					before := primary.Load()
					body := ordinary
					if stream {
						body = streaming
					}
					response := value.call(t, body)
					io.Copy(io.Discard, response.Body)
					response.Body.Close()
					id := response.Header.Get("X-Request-ID")
					var detail Record
					decode(t, value.admin(t, "GET", "requests/"+id, nil), &detail)
					if response.StatusCode != 200 || detail.State != "success" {
						t.Fatal("controlled recovery request", response.StatusCode, detail.State)
					}
					if index < 12 && index%2 == 1 {
						if primary.Load() != before {
							t.Fatal("cooldown was bypassed")
						}
						clock.Add(int64(value.server.Config.Cooldown) + 1)
					}
					for _, fact := range detail.LatencyEvidence {
						if fact.Ordinal == 1 && len(detail.Attempts) > 1 && fact.Acceptance == "accepted" {
							t.Fatal("503 adopted as successful score")
						}
					}
					if primary.Load() > 6 {
						recovered = true
					}
					t.Logf("recovery request_id=%s sample=%d attempts=%d selected=%d state=%s", id, index, len(detail.Attempts), detail.Attempts[len(detail.Attempts)-1].UpstreamID, detail.State)
				}
				if !recovered || primary.Load() <= 6 {
					t.Fatal("normal traffic never recovered")
				}
			})
		}
	}
}

func TestLatencyStorageFailureNeverScoresSuccess(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
				if stream {
					writer.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled\"}}]}\n\ndata: [DONE]\n\n")
					return
				}
				normal(writer, request)
			})
			setLatencyStrategy(t, value, "low_latency")
			if _, err := value.store.DB.Exec("CREATE TRIGGER latency_fail_finish BEFORE UPDATE ON requests BEGIN SELECT RAISE(FAIL,'controlled storage failure'); END"); err != nil {
				t.Fatal(err)
			}
			body := ordinary
			if stream {
				body = streaming
			}
			response := value.call(t, body)
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
			value.server.mu.Lock()
			defer value.server.mu.Unlock()
			for _, pool := range value.server.latency {
				if len(pool.Samples) > 0 {
					t.Fatal("failed durable completion accepted")
				}
			}
			if !value.server.degraded.Load() {
				t.Fatal("fixture did not fail storage")
			}
		})
	}
}

func TestLatencyRealHTTPNoOutputAndNewCandidate(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			var calls atomic.Int32
			value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
				current := calls.Add(1)
				if stream {
					writer.Header().Set("Content-Type", "text/event-stream")
					delta := `{"role":"assistant","content":""}`
					if current > 6 {
						delta = `{"content":"controlled"}`
					}
					fmt.Fprintf(writer, "data: {\"choices\":[{\"delta\":%s}]}\n\ndata: [DONE]\n\n", delta)
				} else {
					normal(writer, request)
				}
			})
			setLatencyStrategy(t, value, "low_latency")
			seen := map[int64]int{}
			for index := 0; index < 12; index++ {
				if index == 6 {
					expect(t, value.admin(t, "POST", "upstreams", map[string]any{"name": "new candidate", "url": value.upstream.URL + "/new/v1", "token": "controlled-new-candidate-token", "enabled": true}), 201)
					model, _ := value.store.routeModel("alpha")
					candidates := cleanTestCandidates(model.Candidates)
					candidates = append(candidates, map[string]any{"upstream_id": 2, "remote_model": "controlled-new", "priority": 0, "weight": 1, "enabled": true})
					expect(t, value.admin(t, "PATCH", "models/alpha", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": candidates}), 200)
				}
				body := ordinary
				if stream {
					body = streaming
				}
				response := value.call(t, body)
				io.Copy(io.Discard, response.Body)
				response.Body.Close()
				id := response.Header.Get("X-Request-ID")
				decisions, facts, err := readRoutingEvidence(value.store.DB, id)
				if err != nil || len(facts) != 1 {
					t.Fatal("sample evidence", err)
				}
				expected := "accepted"
				if stream && index < 6 {
					expected = "no_output_delta"
				}
				if facts[0].Acceptance != expected {
					t.Fatal("sample acceptance", facts[0].Acceptance)
				}
				if expected == "no_output_delta" && facts[0].Value != nil {
					t.Fatal("metadata became fast success")
				}
				if index == 6 {
					for _, decision := range decisions {
						if decision.State != "unknown" || decision.EWMA != nil {
							t.Fatal("new generation inherited score")
						}
					}
				}
				if index >= 6 {
					seen[facts[0].UpstreamID]++
				}
				t.Logf("new_candidate request_id=%s index=%d stream=%t acceptance=%s", id, index, stream, expected)
			}
			if seen[1] == 0 || seen[2] == 0 {
				t.Fatal("new candidate starved", seen)
			}
		})
	}
}
