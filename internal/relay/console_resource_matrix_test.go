package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func singleResourceInstallation(test *testing.T, actual bool) *governedFixture {
	test.Helper()
	fixture := governedInstallation(test, actual)
	fixture.allowUnbounded = true
	if status, _ := fixture.admin.call("PATCH", "models/governed", map[string]any{"expected_revision": 1, "enabled": true, "candidates": []map[string]any{{"upstream_id": 3, "remote_model": "controlled", "enabled": true, "priority": 1, "weight": 1}}}, ""); status != 200 {
		test.Fatal("single controlled route", status)
	}
	if status, _ := fixture.admin.call("PUT", "tenants/"+fixture.tenant+"/limits", ResourceLimits{Revision: 1, Hard: true}, ""); status != 200 {
		test.Fatal("unlimited initial tenant", status)
	}
	return fixture
}

type resourceReceipt struct {
	status int
	id     string
	err    error
}

func resourceGenerate(fixture *governedFixture, secret string) resourceReceipt {
	body := []byte(`{"model":"governed","messages":[{"role":"user","content":"controlled resource competition"}]}`)
	request, _ := http.NewRequest("POST", fixture.origin+"/v1/chat/completions", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+secret)
	var response *http.Response
	if fixture.server.Config.Authority == "127.0.0.1:5544" {
		writer := httptest.NewRecorder()
		fixture.server.ServeHTTP(writer, request)
		response = writer.Result()
	} else {
		var err error
		response, err = http.DefaultClient.Do(request)
		if err != nil {
			return resourceReceipt{err: err}
		}
	}
	defer response.Body.Close()
	_, err := io.Copy(io.Discard, response.Body)
	return resourceReceipt{status: response.StatusCode, id: response.Header.Get("X-Request-ID"), err: err}
}

func resourceUsage(test *testing.T, fixture *governedFixture, path string) ResourceUsage {
	test.Helper()
	status, report := fixture.admin.call("GET", path+"/limits", nil, "")
	var usage ResourceUsage
	data, err := json.Marshal(report["usage"])
	if status != 200 || err != nil || json.Unmarshal(data, &usage) != nil {
		test.Fatal("authoritative resource usage", status)
	}
	return usage
}

func resourceWait(test *testing.T, entered <-chan struct{}) {
	test.Helper()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		test.Fatal("controlled upstream barrier not reached")
	}
}

func resourceFinish(test *testing.T, finished <-chan resourceReceipt, status int) resourceReceipt {
	test.Helper()
	select {
	case receipt := <-finished:
		if receipt.err != nil || receipt.status != status {
			test.Fatal("resource receipt mismatch", receipt.status, status, receipt.err)
		}
		return receipt
	case <-time.After(5 * time.Second):
		test.Fatal("controlled resource request did not finish")
		return resourceReceipt{}
	}
}

func releaseResources(fixture *governedFixture) {
	select {
	case <-fixture.block:
	default:
		close(fixture.block)
	}
}

func TestConsoleResourceMatrixInProcess(test *testing.T)  { resourceMatrixScenario(test, false) }
func TestConsoleResourceMatrixActualHTTP(test *testing.T) { resourceMatrixScenario(test, true) }

func TestLatencyResourceMatrixActualHTTP(test *testing.T) {
	resourceMatrixScenario(test, true, "low_latency")
}

func resourceMatrixScenario(test *testing.T, actual bool, strategies ...string) {
	for _, scope := range []string{"tenant", "key"} {
		for _, dimension := range []string{"calls", "rpm", "tpm", "concurrency", "budget"} {
			test.Run(scope+"/"+dimension, func(test *testing.T) {
				test.Parallel()
				fixture := singleResourceInstallation(test, actual)
				if len(strategies) > 0 {
					model, err := fixture.store.routeModel("governed")
					if err != nil {
						test.Fatal(err)
					}
					if status, _ := fixture.admin.call("PATCH", "models/governed", map[string]any{"expected_revision": model.Revision, "enabled": true, "candidates": cleanTestCandidates(model.Candidates), "routing_strategy": strategies[0]}, ""); status != 200 {
						test.Fatal("low latency resource configuration", status)
					}
				}
				tenantPath := "tenants/" + fixture.tenant
				keyPath := tenantPath + "/keys/" + fixture.keyID
				path, revision := tenantPath, int64(2)
				if scope == "key" {
					path, revision = keyPath, 1
				}
				limit := func(value *int64) ResourceLimits {
					limits := ResourceLimits{Revision: revision, Hard: true}
					switch dimension {
					case "calls":
						limits.Calls = value
					case "rpm":
						limits.RPM = value
					case "tpm":
						limits.TPM = value
					case "concurrency":
						limits.Concurrency = value
					case "budget":
						limits.Budget = value
					}
					return limits
				}
				update := func(value *int64) {
					if status, _ := fixture.admin.call("PUT", path+"/limits", limit(value), ""); status != 200 {
						test.Fatal("independent dimension update", status)
					}
					revision++
				}
				update(ptrInt64(0))
				if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 429 || receipt.err != nil || fixture.hits.Load() != 0 {
					test.Fatal("zero boundary must reject before dispatch", receipt.status)
				}
				if resourceUsage(test, fixture, tenantPath).Calls != 0 || resourceUsage(test, fixture, keyPath).Calls != 0 {
					test.Fatal("zero rejection partially debited scope")
				}
				capacity := int64(1)
				if dimension == "tpm" || dimension == "budget" {
					capacity = 200
				}
				update(&capacity)
				fixture.block, fixture.entered = make(chan struct{}), make(chan struct{}, 1)
				defer releaseResources(fixture)
				finished := make(chan resourceReceipt, 1)
				go func() { finished <- resourceGenerate(fixture, fixture.key) }()
				resourceWait(test, fixture.entered)
				beforeTenant, beforeKey := resourceUsage(test, fixture, tenantPath), resourceUsage(test, fixture, keyPath)
				if beforeKey.Calls != 1 || beforeKey.Concurrency != 1 || beforeKey.Reserved != 200 || beforeKey.TPM != 200 {
					test.Fatal("exact finite boundary reservation", beforeKey)
				}
				if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 429 || receipt.err != nil {
					test.Fatal("finite boundary competitor", receipt.status)
				}
				if resourceUsage(test, fixture, tenantPath) != beforeTenant || resourceUsage(test, fixture, keyPath) != beforeKey || fixture.hits.Load() != 1 {
					test.Fatal("rejection mutated either scope or dispatched")
				}
				update(nil)
				if resourceUsage(test, fixture, tenantPath) != beforeTenant || resourceUsage(test, fixture, keyPath) != beforeKey {
					test.Fatal("null update cleared accrued usage")
				}
				releaseResources(fixture)
				resourceFinish(test, finished, 200)
				if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 200 || receipt.err != nil {
					test.Fatal("null boundary must restore admission", receipt.status)
				}
				usage := resourceUsage(test, fixture, keyPath)
				if usage.Calls != 2 || usage.Fee != 14 || usage.Reserved != 0 || usage.Concurrency != 0 || usage.Pending != 0 {
					test.Fatal("finite/null transition usage conservation", usage)
				}
			})
		}
	}
}

func TestConsoleSharedResourcesInProcess(test *testing.T) {
	test.Parallel()
	sharedResourcesScenario(test, false)
}
func TestConsoleSharedResourcesActualHTTP(test *testing.T) {
	test.Parallel()
	sharedResourcesScenario(test, true)
}

func sharedResourcesScenario(test *testing.T, actual bool) {
	for _, mode := range []string{"multi_key", "old_new_generation"} {
		test.Run(mode, func(test *testing.T) {
			fixture := singleResourceInstallation(test, actual)
			tenantPath := "tenants/" + fixture.tenant
			keyPath := tenantPath + "/keys/" + fixture.keyID
			if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 200 || receipt.err != nil {
				test.Fatal("historical call", receipt.status)
			}
			before := resourceUsage(test, fixture, keyPath)
			secrets := []string{fixture.key}
			var otherKey string
			if mode == "multi_key" {
				status, value := fixture.admin.call("POST", tenantPath+"/keys", KeyInput{Name: "competing application", Application: "controlled tests", Enabled: true, Models: []string{"governed"}}, "competing-key")
				if status != 201 {
					test.Fatal("second key", status)
				}
				secrets = append(secrets, value["key"].(string))
				otherKey = value["id"].(string)
			} else {
				status, value := fixture.admin.call("POST", keyPath+"/rotations", map[string]any{"expected_revision": 1, "transition_hours": 1}, "resource-rotation")
				if status != 200 {
					test.Fatal("rotation", status)
				}
				secrets = append(secrets, value["key"].(string))
				if resourceUsage(test, fixture, keyPath) != before {
					test.Fatal("rotation cleared historical usage")
				}
			}
			path, revision := tenantPath, int64(2)
			if mode == "old_new_generation" {
				path, revision = keyPath, 1
			}
			if status, _ := fixture.admin.call("PUT", path+"/limits", ResourceLimits{Revision: revision, Concurrency: ptrInt64(2), Hard: true}, ""); status != 200 {
				test.Fatal("shared capacity", status)
			}
			fixture.block, fixture.entered = make(chan struct{}), make(chan struct{}, 2)
			defer releaseResources(fixture)
			gate, finished := make(chan struct{}), make(chan resourceReceipt, 8)
			for index := 0; index < 8; index++ {
				secret := secrets[index%2]
				go func() { <-gate; finished <- resourceGenerate(fixture, secret) }()
			}
			close(gate)
			resourceWait(test, fixture.entered)
			resourceWait(test, fixture.entered)
			for index := 0; index < 6; index++ {
				resourceFinish(test, finished, 429)
			}
			usage := resourceUsage(test, fixture, tenantPath)
			if usage.Calls != 3 || usage.Concurrency != 2 || usage.Reserved != 400 || fixture.hits.Load() != 3 {
				test.Fatal("shared atomic N capacity", usage)
			}
			keyUsage := resourceUsage(test, fixture, keyPath)
			if mode == "multi_key" {
				second := resourceUsage(test, fixture, tenantPath+"/keys/"+otherKey)
				if keyUsage.Calls+second.Calls != usage.Calls || keyUsage.Concurrency+second.Concurrency != 2 || keyUsage.Reserved+second.Reserved != 400 {
					test.Fatal("multi-key scope accounting")
				}
			} else if keyUsage != usage {
				test.Fatal("old/new generations did not share application usage")
			}
			status, otherTenant := fixture.admin.call("POST", "tenants", map[string]any{"name": "independent tenant"}, "resource-other-tenant")
			if status != 201 {
				test.Fatal("other tenant", status)
			}
			otherID := otherTenant["id"].(string)
			if status, _ := fixture.admin.call("PUT", "models/governed/tenant-grants", map[string]any{"tenants": []string{fixture.tenant, otherID}}, ""); status != 200 {
				test.Fatal("other grant", status)
			}
			status, other := fixture.admin.call("POST", "tenants/"+otherID+"/keys", KeyInput{Name: "independent", Application: "controlled tests", Enabled: true, Models: []string{"governed"}}, "resource-independent-key")
			if status != 201 {
				test.Fatal("independent key", status)
			}
			otherFinished := make(chan resourceReceipt, 1)
			go func() { otherFinished <- resourceGenerate(fixture, other["key"].(string)) }()
			resourceWait(test, fixture.entered)
			if resourceUsage(test, fixture, tenantPath) != usage {
				test.Fatal("other tenant shared occupied scope")
			}
			releaseResources(fixture)
			first, second := resourceFinish(test, finished, 200), resourceFinish(test, finished, 200)
			if first.id == "" || second.id == "" || first.id == second.id {
				test.Fatal("admitted Request IDs must be distinct")
			}
			resourceFinish(test, otherFinished, 200)
			after := resourceUsage(test, fixture, tenantPath)
			if after.Calls != 3 || after.Fee != 21 || after.Concurrency != 0 || after.Reserved != 0 || after.Pending != 0 {
				test.Fatal("shared completed usage", after)
			}
			if value := resourceUsage(test, fixture, "tenants/"+otherID); value.Calls != 1 || value.Fee != 7 {
				test.Fatal("independent tenant completion", value)
			}
			if mode == "old_new_generation" {
				status, metadata := fixture.admin.call("GET", tenantPath+"/keys", nil, "")
				if status != 200 || len(metadata["items"].([]any)) != 1 {
					test.Fatal("rotation did not preserve single application")
				}
				key := metadata["items"].([]any)[0].(map[string]any)
				if key["revision"] != float64(2) || len(key["generations"].([]any)) != 2 {
					test.Fatal("rotation generation conservation")
				}
			}
			test.Log(fmt.Sprintf("scope=%s contenders=8 accepted=2 denied=6 other_tenant=1 real_http=%t", mode, actual))
			if reflect.DeepEqual(after, before) {
				test.Fatal("accepted calls not accounted")
			}
		})
	}
}
