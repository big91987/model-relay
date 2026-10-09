package relay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type meteringObservation struct {
	MaxTokens           *int64 `json:"max_tokens"`
	MaxCompletionTokens *int64 `json:"max_completion_tokens"`
	Model               string `json:"model"`
	Stream              bool   `json:"stream"`
}

type meteringFixture struct {
	*governedFixture
	mode     atomic.Int64
	observed chan meteringObservation
	canceled chan struct{}
	prefix   chan struct{}
}

func meteringInstallation(test *testing.T, actual bool) *meteringFixture {
	test.Helper()
	value := &meteringFixture{observed: make(chan meteringObservation, 16), canceled: make(chan struct{}, 1), prefix: make(chan struct{}, 1)}
	value.governedFixture = governedInstallationWithHandler(test, actual, func(fixture *governedFixture) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			fixture.hits.Add(1)
			var input meteringObservation
			if json.NewDecoder(request.Body).Decode(&input) != nil {
				test.Error("controlled metering payload decoding")
				writer.WriteHeader(400)
				return
			}
			value.observed <- input
			if value.mode.Load() == 7 {
				<-request.Context().Done()
				value.canceled <- struct{}{}
				return
			}
			if value.mode.Load() == 5 {
				writer.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(writer, terminalChunk)
				writer.(http.Flusher).Flush()
				value.prefix <- struct{}{}
				<-request.Context().Done()
				value.canceled <- struct{}{}
				return
			}
			if fixture.block != nil {
				select {
				case fixture.entered <- struct{}{}:
				default:
				}
				select {
				case <-fixture.block:
				case <-request.Context().Done():
					return
				}
			}
			mode := value.mode.Load()
			usage := `"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}`
			if mode == 1 || mode == 3 {
				usage = ""
			} else if mode == 4 {
				usage = `"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":8}`
			} else if mode == 6 {
				usage = `"usage":{"prompt_tokens":3,"completion_tokens":101,"total_tokens":104}`
			}
			if mode == 2 || mode == 3 {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(500)
				fmt.Fprint(writer, `{"error":{"code":"controlled_failure"}`)
				if usage != "" {
					fmt.Fprint(writer, ","+usage)
				}
				fmt.Fprint(writer, "}")
				return
			}
			if input.Stream {
				writer.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(writer, terminalChunk)
				writer.(http.Flusher).Flush()
				if usage != "" {
					fmt.Fprint(writer, `data: {"choices":[],`+usage+"}\n\n")
				}
				fmt.Fprint(writer, "data: [DONE]\n\n")
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]`)
			if usage != "" {
				fmt.Fprint(writer, ","+usage)
			}
			fmt.Fprint(writer, "}")
		})
	})
	if status, _ := value.admin.call("PATCH", "models/governed", map[string]any{"expected_revision": 1, "enabled": true, "candidates": []map[string]any{{"upstream_id": 3, "remote_model": "controlled", "enabled": true, "priority": 1, "weight": 1}}}, ""); status != 200 {
		test.Fatal("metering single candidate", status)
	}
	value.clock.Add(int64(2 * time.Minute))
	return value
}

func meteringBody(stream bool) map[string]any {
	return map[string]any{"model": "governed", "messages": []map[string]string{{"role": "user", "content": "controlled metering input"}}, "stream": stream}
}

func meteringPaths(fixture *governedFixture) (string, string) {
	tenant := "tenants/" + fixture.tenant
	return tenant, tenant + "/keys/" + fixture.keyID
}

func meteringSnapshot(test *testing.T, fixture *governedFixture, path, phase string) ResourceUsage {
	test.Helper()
	status, report := fixture.admin.call("GET", path+"/limits", nil, "")
	var usage ResourceUsage
	encoded, err := json.Marshal(report["usage"])
	if status != 200 || err != nil || json.Unmarshal(encoded, &usage) != nil {
		test.Fatal("metering snapshot", status)
	}
	whitelist, err := json.Marshal(map[string]any{"phase": phase, "scope": path, "limits": report["limits"], "usage": usage, "period": report["period"], "window_utc": fixture.store.now().Format("2006-01-02T15:04"), "window_source": "committed Store.now test seam; not an API minute field", "budget_remaining_micro_usd": report["budget_remaining_micro_usd"], "effective_utc": fixture.store.now().Format(time.RFC3339Nano), "dispatches": fixture.hits.Load()})
	if err != nil {
		test.Fatal("metering evidence encoding", err)
	}
	test.Log("METERING_SNAPSHOT " + string(whitelist))
	return usage
}

func meteringAssert(test *testing.T, fixture *governedFixture, phase string, want ResourceUsage) {
	test.Helper()
	tenant, key := meteringPaths(fixture)
	for _, path := range []string{tenant, key} {
		if got := meteringSnapshot(test, fixture, path, phase); got != want {
			test.Fatalf("%s usage got=%+v want=%+v", phase, got, want)
		}
	}
}

func meteringReject(test *testing.T, fixture *governedFixture, body map[string]any, wantStatus int, wantCode string) {
	test.Helper()
	tenant, key := meteringPaths(fixture)
	beforeTenant, beforeKey := resourceUsage(test, fixture, tenant), resourceUsage(test, fixture, key)
	beforeHits := fixture.hits.Load()
	status, id := fixture.generate(body, "")
	if status != wantStatus || id == "" || fixture.hits.Load() != beforeHits {
		test.Fatal("rejection dispatched or wrong status", status, wantStatus)
	}
	status, detail := fixture.admin.call("GET", "workspace/requests/"+id, nil, "")
	if status != 200 || detail["error_code"] != wantCode || len(detail["attempts"].([]any)) != 0 {
		test.Fatal("same-id rejection evidence", status, detail["error_code"])
	}
	if resourceUsage(test, fixture, tenant) != beforeTenant || resourceUsage(test, fixture, key) != beforeKey {
		test.Fatal("rejection partially debited another dimension or scope")
	}
	test.Logf("METERING_REJECT request_id=%s status=%d code=%s dispatch_delta=0", id, wantStatus, wantCode)
}

func TestMeteringSafetyInProcess(test *testing.T) {
	test.Parallel()
	meteringSafety(test, false)
}
func TestMeteringSafetyActualHTTP(test *testing.T) {
	test.Parallel()
	meteringSafety(test, true)
}

func meteringSafety(test *testing.T, actual bool) {
	for _, trigger := range []string{"tenant_tpm", "key_tpm", "hard_budget"} {
		test.Run(trigger, func(test *testing.T) {
			fixture := meteringInstallation(test, actual)
			tenant, key := meteringPaths(fixture.governedFixture)
			if status, _ := fixture.admin.call("PUT", tenant+"/limits", ResourceLimits{Revision: 1, Hard: true}, ""); status != 200 {
				test.Fatal("clear initial tenant limits", status)
			}
			path, revision := tenant, int64(2)
			limits := ResourceLimits{Revision: revision, Hard: true, TPM: ptrInt64(10000)}
			if trigger == "key_tpm" {
				path, limits.Revision = key, 1
			} else if trigger == "hard_budget" {
				limits.TPM, limits.Budget = nil, ptrInt64(10000)
			}
			if status, _ := fixture.admin.call("PUT", path+"/limits", limits, ""); status != 200 {
				test.Fatal("bounded trigger", status)
			}
			if status, _ := fixture.admin.call("PUT", "models/governed/bounds", map[string]any{"upstream_id": 3, "max_input": 100, "max_output": 64, "evidence": "controlled protocol cap only; not supplier billing proof"}, ""); status != 200 {
				test.Fatal("formal bounds configuration", status)
			}
			for _, stream := range []bool{false, true} {
				for _, field := range []string{"default", "max_tokens", "max_completion_tokens"} {
					body := meteringBody(stream)
					if field != "default" {
						body[field] = 32
					}
					status, id := fixture.generate(body, "")
					if status != 200 {
						test.Fatal("safe generation", field, status)
					}
					observed := <-fixture.observed
					cap := observed.MaxTokens
					if field == "max_completion_tokens" {
						cap = observed.MaxCompletionTokens
						if observed.MaxTokens != nil {
							test.Fatal("conflicting outgoing caps")
						}
					}
					want := int64(32)
					if field == "default" {
						want = 64
					}
					if cap == nil || *cap != want || observed.Stream != stream || observed.Model != "controlled" {
						test.Fatal("actual controlled upstream output limit/mapping", observed)
					}
					status, detail := fixture.admin.call("GET", "workspace/requests/"+id, nil, "")
					if status != 200 || len(detail["attempts"].([]any)) != 1 {
						test.Fatal("same-id bounded dispatch diagnostic", status)
					}
					test.Logf("METERING_CAP request_id=%s trigger=%s stream=%t field=%s effective_cap=%d real_http=%t", id, trigger, stream, field, *cap, actual)
				}
			}
			for _, invalid := range []struct {
				fields map[string]any
				code   string
			}{{map[string]any{"max_tokens": 65}, "unsupported_bounded_request"}, {map[string]any{"max_tokens": 0}, "invalid_request"}, {map[string]any{"max_tokens": 64, "max_completion_tokens": 64}, "unsupported_bounded_request"}, {map[string]any{"n": 2}, "unsupported_bounded_request"}, {map[string]any{"thinking": map[string]any{"type": "disabled"}}, "unsupported_field"}, {map[string]any{"tools": []any{}}, "unsupported_bounded_request"}} {
				body := meteringBody(false)
				for field, value := range invalid.fields {
					body[field] = value
				}
				meteringReject(test, fixture.governedFixture, body, 400, invalid.code)
			}
			meteringAssert(test, fixture.governedFixture, "bounded_complete", ResourceUsage{Calls: 6, RPM: 6, TPM: 42, Fee: 42})
		})
	}
}

func TestMeteringTerminalInProcess(test *testing.T) {
	test.Parallel()
	meteringTerminal(test, false)
}
func TestMeteringTerminalActualHTTP(test *testing.T) {
	test.Parallel()
	meteringTerminal(test, true)
}

func meteringTerminal(test *testing.T, actual bool) {
	for _, scenario := range []struct {
		name   string
		mode   int64
		stream bool
		known  bool
		status int
	}{{"ordinary_known", 0, false, true, 200}, {"sse_known", 0, true, true, 200}, {"ordinary_unknown", 1, false, false, 200}, {"sse_unknown", 1, true, false, 200}, {"failure_known", 2, false, true, 502}, {"failure_unknown", 3, false, false, 502}, {"ordinary_inconsistent", 4, false, false, 200}, {"sse_inconsistent", 4, true, false, 200}} {
		test.Run(scenario.name, func(test *testing.T) {
			fixture := meteringInstallation(test, actual)
			fixture.mode.Store(scenario.mode)
			status, id := fixture.generate(meteringBody(scenario.stream), "terminal-once")
			if status != scenario.status || fixture.hits.Load() != 1 {
				test.Fatal("controlled terminal", status, scenario.status, fixture.hits.Load())
			}
			want := ResourceUsage{Calls: 1, RPM: 1, TPM: 200, Reserved: 200, Pending: 1}
			if scenario.known {
				want.TPM, want.Fee, want.Reserved, want.Pending = 7, 7, 0, 0
			}
			meteringAssert(test, fixture.governedFixture, scenario.name, want)
			status, detail := fixture.admin.call("GET", "workspace/requests/"+id, nil, "")
			if status != 200 || len(detail["attempts"].([]any)) != 1 || len(detail["charge_entries"].([]any)) != 1 {
				test.Fatal("terminal same-id ledger", status)
			}
			entry := detail["charge_entries"].([]any)[0].(map[string]any)
			fee, cost := "0", "0"
			if scenario.known {
				fee, cost = "7", "33"
			}
			if entry["fee"] != fee || entry["cost"] != cost {
				test.Fatal("independent controlled cost/tenant fee", entry["fee"], entry["cost"])
			}
			if status, prior := fixture.generate(meteringBody(scenario.stream), "terminal-once"); status != 409 || prior != id || fixture.hits.Load() != 1 {
				test.Fatal("terminal replay dispatch", status)
			}
			meteringAssert(test, fixture.governedFixture, "duplicate_terminal", want)
		})
	}
}

func TestMeteringWindowsInProcess(test *testing.T) {
	test.Parallel()
	meteringWindows(test, false)
}
func TestMeteringWindowsActualHTTP(test *testing.T) {
	test.Parallel()
	meteringWindows(test, true)
}

func meteringWindows(test *testing.T, actual bool) {
	for _, scope := range []string{"tenant", "key"} {
		for _, dimension := range []string{"rpm", "tpm"} {
			test.Run(scope+"/"+dimension, func(test *testing.T) {
				fixture := meteringInstallation(test, actual)
				tenant, key := meteringPaths(fixture.governedFixture)
				path, revision := tenant, int64(1)
				if scope == "key" {
					path, revision = key, 1
				}
				update := func(value int64) {
					limits := ResourceLimits{Revision: revision, Hard: true, TPM: ptrInt64(10000)}
					if dimension == "rpm" {
						limits.RPM = &value
					} else {
						limits.TPM = &value
					}
					if status, _ := fixture.admin.call("PUT", path+"/limits", limits, ""); status != 200 {
						test.Fatal("window limits", status)
					}
					revision++
				}
				if status, _ := fixture.generate(meteringBody(false), ""); status != 200 {
					test.Fatal("known warmup", status)
				}
				meteringAssert(test, fixture.governedFixture, "warmup", ResourceUsage{Calls: 1, RPM: 1, TPM: 7, Fee: 7})
				if dimension == "rpm" {
					update(2)
					if status, _ := fixture.generate(meteringBody(false), ""); status != 200 {
						test.Fatal("RPM exact N", status)
					}
					meteringAssert(test, fixture.governedFixture, "rpm_exact", ResourceUsage{Calls: 2, RPM: 2, TPM: 14, Fee: 14})
					meteringReject(test, fixture.governedFixture, meteringBody(false), 429, scope+"_rpm_exceeded")
					start := time.Unix(0, fixture.clock.Load()).Truncate(time.Minute)
					fixture.clock.Store(start.Add(time.Minute - time.Nanosecond).UnixNano())
					meteringReject(test, fixture.governedFixture, meteringBody(false), 429, scope+"_rpm_exceeded")
					fixture.clock.Store(start.Add(time.Minute).UnixNano())
					meteringAssert(test, fixture.governedFixture, "new_fixed_minute", ResourceUsage{Calls: 2, Fee: 14})
					if status, _ := fixture.generate(meteringBody(true), ""); status != 200 {
						test.Fatal("explicit next-minute request", status)
					}
					meteringAssert(test, fixture.governedFixture, "rpm_recovered", ResourceUsage{Calls: 3, RPM: 1, TPM: 7, Fee: 21})
					return
				}
				update(206)
				meteringReject(test, fixture.governedFixture, meteringBody(false), 429, scope+"_tpm_exceeded")
				update(207)
				fixture.mode.Store(1)
				status, unknownID := fixture.generate(meteringBody(true), "unknown-once")
				if status != 200 {
					test.Fatal("TPM U+B equality", status)
				}
				meteringAssert(test, fixture.governedFixture, "unknown_carry", ResourceUsage{Calls: 2, RPM: 2, TPM: 207, Fee: 7, Reserved: 200, Pending: 1})
				update(406)
				meteringReject(test, fixture.governedFixture, meteringBody(false), 429, scope+"_tpm_exceeded")
				update(407)
				fixture.mode.Store(0)
				if status, _ := fixture.generate(meteringBody(false), ""); status != 200 {
					test.Fatal("TPM U+P+B equality", status)
				}
				meteringAssert(test, fixture.governedFixture, "known_with_pending", ResourceUsage{Calls: 3, RPM: 3, TPM: 214, Fee: 14, Reserved: 200, Pending: 1})
				fixture.clock.Add(int64(time.Minute))
				meteringAssert(test, fixture.governedFixture, "pending_across_minute", ResourceUsage{Calls: 3, TPM: 200, Fee: 14, Reserved: 200, Pending: 1})
				update(399)
				meteringReject(test, fixture.governedFixture, meteringBody(false), 429, scope+"_tpm_exceeded")
				update(400)
				if status, _ := fixture.generate(meteringBody(false), ""); status != 200 {
					test.Fatal("TPM P+B next-minute equality", status)
				}
				reconciliation := map[string]any{"input_tokens": 3, "output_tokens": 4, "reason": "controlled receipt", "evidence": "controlled test only, not supplier evidence", "expected_revision": 1}
				reconcilePath := "workspace/requests/" + unknownID + "/reconciliations/1"
				for range 2 {
					if status, _ := fixture.admin.call("POST", reconcilePath, reconciliation, "resolve-unknown"); status != 200 {
						test.Fatal("append-only reconciliation/replay", status)
					}
					meteringAssert(test, fixture.governedFixture, "reconciled_old_minute", ResourceUsage{Calls: 4, RPM: 1, TPM: 7, Fee: 28})
				}
			})
		}
	}
}

func TestMeteringCompetitionInProcess(test *testing.T) {
	test.Parallel()
	meteringCompetition(test, false)
}
func TestMeteringCompetitionActualHTTP(test *testing.T) {
	test.Parallel()
	meteringCompetition(test, true)
}

func meteringCompetition(test *testing.T, actual bool) {
	for _, pool := range []string{"multi_key", "old_new_generation"} {
		for _, dimension := range []string{"rpm", "tpm"} {
			test.Run(pool+"/"+dimension, func(test *testing.T) {
				fixture := meteringInstallation(test, actual)
				tenant, key := meteringPaths(fixture.governedFixture)
				if status, _ := fixture.generate(meteringBody(false), ""); status != 200 {
					test.Fatal("competition warmup", status)
				}
				secrets := []string{fixture.key}
				otherKeyID := ""
				if pool == "multi_key" {
					status, other := fixture.admin.call("POST", tenant+"/keys", KeyInput{Name: "second metering key", Application: "controlled tests", Enabled: true, Models: []string{"governed"}}, "metering-other-key")
					if status != 201 {
						test.Fatal("second metering key", status)
					}
					secrets = append(secrets, other["key"].(string))
					otherKeyID = other["id"].(string)
				} else {
					status, rotated := fixture.admin.call("POST", key+"/rotations", map[string]any{"expected_revision": 1, "transition_hours": 1}, "metering-rotate")
					if status != 200 {
						test.Fatal("metering rotation", status)
					}
					secrets = append(secrets, rotated["key"].(string))
				}
				path := tenant
				if pool == "old_new_generation" {
					path = key
				}
				limits := ResourceLimits{Revision: 1, Hard: true, TPM: ptrInt64(207)}
				if dimension == "rpm" {
					limits.RPM, limits.TPM = ptrInt64(2), ptrInt64(10000)
				}
				if status, _ := fixture.admin.call("PUT", path+"/limits", limits, ""); status != 200 {
					test.Fatal("shared remaining one admission", status)
				}
				meteringAssert(test, fixture.governedFixture, "rotation_limit_conservation", ResourceUsage{Calls: 1, RPM: 1, TPM: 7, Fee: 7})
				fixture.block, fixture.entered = make(chan struct{}), make(chan struct{}, 1)
				defer releaseResources(fixture.governedFixture)
				gate, finished := make(chan struct{}), make(chan resourceReceipt, 8)
				remaining := 8
				defer func() {
					releaseResources(fixture.governedFixture)
					deadline := time.NewTimer(5 * time.Second)
					defer deadline.Stop()
					for remaining > 0 {
						select {
						case <-finished:
							remaining--
						case <-deadline.C:
							test.Error("competition workers did not drain")
							return
						}
					}
				}()
				for index := 0; index < 8; index++ {
					secret := secrets[index%2]
					go func() { <-gate; finished <- resourceGenerate(fixture.governedFixture, secret) }()
				}
				close(gate)
				resourceWait(test, fixture.entered)
				seen := map[string]bool{}
				for range 7 {
					receipt := resourceFinish(test, finished, 429)
					remaining--
					if receipt.id == "" || seen[receipt.id] {
						test.Fatal("distinct concurrent rejection IDs")
					}
					seen[receipt.id] = true
					status, detail := fixture.admin.call("GET", "workspace/requests/"+receipt.id, nil, "")
					scope := "tenant"
					if pool == "old_new_generation" {
						scope = "key"
					}
					if status != 200 || detail["error_code"] != scope+"_"+dimension+"_exceeded" || len(detail["attempts"].([]any)) != 0 {
						test.Fatal("concurrent rejection zero dispatch", status, detail["error_code"])
					}
				}
				occupied := meteringSnapshot(test, fixture.governedFixture, tenant, "contenders_blocked")
				if occupied != (ResourceUsage{Calls: 2, RPM: 2, TPM: 207, Concurrency: 1, Fee: 7, Reserved: 200}) || fixture.hits.Load() != 2 {
					test.Fatal("atomic shared pool overadmitted", occupied, fixture.hits.Load())
				}
				if pool == "old_new_generation" && resourceUsage(test, fixture.governedFixture, key) != occupied {
					test.Fatal("rotated generations did not share stable key")
				}
				if pool == "multi_key" {
					primary := meteringSnapshot(test, fixture.governedFixture, key, "primary_key_blocked")
					secondary := meteringSnapshot(test, fixture.governedFixture, tenant+"/keys/"+otherKeyID, "secondary_key_blocked")
					if primary.Calls+secondary.Calls != 2 || primary.RPM+secondary.RPM != 2 || primary.TPM+secondary.TPM != 207 || primary.Concurrency+secondary.Concurrency != 1 || primary.Fee+secondary.Fee != 7 || primary.Reserved+secondary.Reserved != 200 || primary.Pending+secondary.Pending != 0 {
						test.Fatal("separate keys did not sum to shared tenant pool")
					}
				}
				status, otherTenant := fixture.admin.call("POST", "tenants", map[string]any{"name": "independent metering tenant"}, "metering-other-tenant")
				if status != 201 {
					test.Fatal("independent tenant", status)
				}
				otherID := otherTenant["id"].(string)
				if status, _ := fixture.admin.call("PUT", "models/governed/tenant-grants", map[string]any{"tenants": []string{fixture.tenant, otherID}}, ""); status != 200 {
					test.Fatal("independent tenant model grant", status)
				}
				status, other := fixture.admin.call("POST", "tenants/"+otherID+"/keys", KeyInput{Name: "independent metering", Application: "controlled tests", Enabled: true, Models: []string{"governed"}}, "metering-independent")
				if status != 201 {
					test.Fatal("independent tenant key", status)
				}
				otherFinished := make(chan resourceReceipt, 1)
				go func() { otherFinished <- resourceGenerate(fixture.governedFixture, other["key"].(string)) }()
				resourceWait(test, fixture.entered)
				if resourceUsage(test, fixture.governedFixture, tenant) != occupied {
					test.Fatal("independent tenant consumed original pool")
				}
				releaseResources(fixture.governedFixture)
				accepted := resourceFinish(test, finished, 200)
				remaining--
				if accepted.id == "" || seen[accepted.id] {
					test.Fatal("accepted contender identity")
				}
				resourceFinish(test, otherFinished, 200)
				final := resourceUsage(test, fixture.governedFixture, tenant)
				if final != (ResourceUsage{Calls: 2, RPM: 2, TPM: 14, Fee: 14}) {
					test.Fatal("competition settlement", final)
				}
				if got := resourceUsage(test, fixture.governedFixture, "tenants/"+otherID); got != (ResourceUsage{Calls: 1, RPM: 1, TPM: 7, Fee: 7}) {
					test.Fatal("independent metering settlement", got)
				}
				test.Logf("METERING_COMPETITION pool=%s dimension=%s contenders=8 accepted=1 denied=7 independent_tenant=1 real_http=%t", pool, dimension, actual)
			})
		}
	}
}

func TestMeteringBoundViolationInProcess(test *testing.T) {
	test.Parallel()
	meteringBoundViolation(test, false)
}
func TestMeteringBoundViolationActualHTTP(test *testing.T) {
	test.Parallel()
	meteringBoundViolation(test, true)
}

func meteringBoundViolation(test *testing.T, actual bool) {
	fixture := meteringInstallation(test, actual)
	fixture.mode.Store(6)
	if status, _ := fixture.generate(meteringBody(false), ""); status != 200 {
		test.Fatal("controlled provider violation receipt", status)
	}
	meteringAssert(test, fixture.governedFixture, "reported_bound_violation", ResourceUsage{Calls: 1, RPM: 1, TPM: 104, Fee: 104})
	meteringReject(test, fixture.governedFixture, meteringBody(false), 503, "trusted_bound_required")
	test.Logf("METERING_BOUND violated_output=101 configured_output=100 next_request_dispatch_delta=0 real_http=%t", actual)
}

func TestMeteringUndispatchedInProcess(test *testing.T) {
	test.Parallel()
	fixture := meteringInstallation(test, false)
	record := Record{ID: randomSecret(16), Time: fixture.store.now().Format(time.RFC3339Nano), KeyID: fixture.keyID, Alias: "governed"}
	encoded, err := json.Marshal(meteringBody(false))
	if err != nil {
		test.Fatal(err)
	}
	payload, _, _, code := validateRequest(encoded)
	if code != "" {
		test.Fatal("undispatched input", code)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request := httptest.NewRequest("POST", fixture.origin+"/v1/chat/completions", bytes.NewReader(encoded)).WithContext(ctx)
	request.Header.Set("Authorization", "Bearer "+fixture.key)
	fixture.server.mu.Lock()
	snapshot, err := fixture.server.snapshot("governed")
	fixture.server.mu.Unlock()
	if err != nil || fixture.store.Start(record) != nil {
		test.Fatal("undispatched snapshot/start", err)
	}
	if code, status, _ := fixture.store.admit(record, snapshot, payload, request); status != 200 || code != "" {
		test.Fatal("undispatched admission", status, code)
	}
	meteringAssert(test, fixture.governedFixture, "admitted_before_dispatch", ResourceUsage{Calls: 1, RPM: 1, TPM: 200, Concurrency: 1, Reserved: 200})
	cancel()
	fixture.server.routedGenerate(httptest.NewRecorder(), request, ctx, &record, payload, snapshot, func(state, code string) bool {
		record.State, record.Code = state, code
		return fixture.store.Finish(record) == nil
	})
	if record.State != "canceled" || record.Code != "client_canceled" || fixture.hits.Load() != 0 {
		test.Fatal("undispatched cancel must not call upstream", record.State, record.Code)
	}
	meteringAssert(test, fixture.governedFixture, "undispatched_settled", ResourceUsage{Calls: 1, RPM: 1})
	record.State, record.Code = "success", ""
	if fixture.store.Finish(record) != nil {
		test.Fatal("duplicate/late undispatched finish")
	}
	meteringAssert(test, fixture.governedFixture, "late_finish_no_second_release", ResourceUsage{Calls: 1, RPM: 1})
	status, detail := fixture.admin.call("GET", "workspace/requests/"+record.ID, nil, "")
	if status != 200 || detail["state"] != "canceled" || len(detail["attempts"].([]any)) != 0 {
		test.Fatal("late result replaced canceled terminal")
	}
	test.Logf("METERING_UNDISPATCHED request_id=%s calls=1 rpm=1 tpm=0 concurrency=0 fee=0 pending=0 attempts=0 layer=admission_dispatch_contract_not_http", record.ID)
}

func TestMeteringTimeoutInProcess(test *testing.T) {
	test.Parallel()
	meteringTimeout(test, false)
}
func TestMeteringTimeoutActualHTTP(test *testing.T) {
	test.Parallel()
	meteringTimeout(test, true)
}

func meteringTimeout(test *testing.T, actual bool) {
	fixture := meteringInstallation(test, actual)
	fixture.mode.Store(7)
	fixture.server.Config.TotalTimeout = 2 * time.Second
	status, id := fixture.generate(meteringBody(false), "")
	if status != 504 || fixture.hits.Load() != 1 {
		test.Fatal("controlled post-dispatch timeout", status, fixture.hits.Load())
	}
	resourceWait(test, fixture.canceled)
	meteringAssert(test, fixture.governedFixture, "post_dispatch_timeout", ResourceUsage{Calls: 1, RPM: 1, TPM: 200, Reserved: 200, Pending: 1})
	status, detail := fixture.admin.call("GET", "workspace/requests/"+id, nil, "")
	if status != 200 || detail["state"] != "timeout" || detail["error_code"] != "request_timeout" || len(detail["attempts"].([]any)) != 1 {
		test.Fatal("timeout diagnostic", status, detail["state"], detail["error_code"])
	}
	test.Logf("METERING_TIMEOUT request_id=%s status=504 attempts=1 rpm_retained=1 unknown_tokens=200 real_http=%t", id, actual)
}

type meteringCancelWriter struct {
	*httptest.ResponseRecorder
	first chan struct{}
}

func (writer *meteringCancelWriter) Write(data []byte) (int, error) {
	count, err := writer.ResponseRecorder.Write(data)
	if bytes.Contains(data, []byte("controlled")) {
		select {
		case writer.first <- struct{}{}:
		default:
		}
	}
	return count, err
}

func TestMeteringCancellationInProcess(test *testing.T) {
	test.Parallel()
	meteringCancellation(test, false)
}
func TestMeteringCancellationActualHTTP(test *testing.T) {
	test.Parallel()
	meteringCancellation(test, true)
}

func meteringCancellation(test *testing.T, actual bool) {
	fixture := meteringInstallation(test, actual)
	fixture.mode.Store(5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	body, err := json.Marshal(meteringBody(true))
	if err != nil {
		test.Fatal(err)
	}
	request, err := http.NewRequestWithContext(ctx, "POST", fixture.origin+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		test.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+fixture.key)
	first, finished := make(chan struct{}, 1), make(chan resourceReceipt, 1)
	go func() {
		if !actual {
			writer := &meteringCancelWriter{ResponseRecorder: httptest.NewRecorder(), first: first}
			fixture.server.ServeHTTP(writer, request)
			finished <- resourceReceipt{status: writer.Code, id: writer.Header().Get("X-Request-ID")}
			return
		}
		response, callErr := http.DefaultClient.Do(request)
		if callErr != nil {
			finished <- resourceReceipt{err: callErr}
			return
		}
		defer response.Body.Close()
		reader := bufio.NewReader(response.Body)
		line, readErr := reader.ReadString('\n')
		if readErr != nil || !strings.HasPrefix(line, "data: ") {
			finished <- resourceReceipt{err: fmt.Errorf("SSE first line not observed: %w", readErr)}
			return
		}
		first <- struct{}{}
		io.Copy(io.Discard, reader)
		finished <- resourceReceipt{status: response.StatusCode, id: response.Header.Get("X-Request-ID")}
	}()
	if actual {
		resourceWait(test, first)
	} else {
		resourceWait(test, fixture.prefix)
	}
	meteringAssert(test, fixture.governedFixture, "sse_before_cancel", ResourceUsage{Calls: 1, RPM: 1, TPM: 200, Concurrency: 1, Reserved: 200})
	cancel()
	var receipt resourceReceipt
	select {
	case receipt = <-finished:
		if receipt.err != nil || actual && receipt.status != 200 || receipt.id == "" {
			test.Fatal("cancel client receipt", receipt.status, receipt.err)
		}
	case <-time.After(5 * time.Second):
		test.Fatal("cancel client worker did not finish")
	}
	resourceWait(test, fixture.canceled)
	deadline := time.Now().Add(5 * time.Second)
	for {
		status, detail := fixture.admin.call("GET", "workspace/requests/"+receipt.id, nil, "")
		if status == 200 && detail["state"] == "canceled" && detail["pending"] == float64(1) {
			if detail["error_code"] != "client_canceled" || len(detail["attempts"].([]any)) != 1 {
				test.Fatal("cancel diagnostic", detail["error_code"])
			}
			break
		}
		if time.Now().After(deadline) {
			test.Fatal("canceled settlement not observable", status, detail["state"])
		}
		time.Sleep(10 * time.Millisecond)
	}
	meteringAssert(test, fixture.governedFixture, "sse_cancel_unknown", ResourceUsage{Calls: 1, RPM: 1, TPM: 200, Reserved: 200, Pending: 1})
	if fixture.hits.Load() != 1 {
		test.Fatal("cancel replayed upstream")
	}
	fixture.clock.Add(int64(time.Minute))
	meteringAssert(test, fixture.governedFixture, "cancel_next_minute", ResourceUsage{Calls: 1, TPM: 200, Reserved: 200, Pending: 1})
	test.Logf("METERING_CANCEL request_id=%s first_output_observed=%t upstream_prefix_written=true upstream_context_canceled=true attempts=1 real_http=%t", receipt.id, actual, actual)
}
