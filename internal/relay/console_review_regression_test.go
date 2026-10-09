package relay

import (
	"testing"
	"time"
)

func TestConsoleReviewZeroSoftBudgetInProcess(test *testing.T) {
	test.Parallel()
	reviewZeroSoftBudget(test, false)
}
func TestConsoleReviewZeroSoftBudgetActualHTTP(test *testing.T) {
	test.Parallel()
	reviewZeroSoftBudget(test, true)
}

func reviewZeroSoftBudget(test *testing.T, actual bool) {
	for _, scope := range []string{"tenant", "key"} {
		test.Run(scope, func(test *testing.T) {
			fixture := singleResourceInstallation(test, actual)
			tenantPath := "tenants/" + fixture.tenant
			keyPath := tenantPath + "/keys/" + fixture.keyID
			path, revision := tenantPath, int64(2)
			if scope == "key" {
				path, revision = keyPath, 1
			}
			update := func(budget *int64) {
				if status, _ := fixture.admin.call("PUT", path+"/limits", ResourceLimits{Revision: revision, Budget: budget, Hard: false}, ""); status != 200 {
					test.Fatal("limits", status)
				}
				revision++
			}
			update(ptrInt64(0))
			beforeTenant, beforeKey := resourceUsage(test, fixture, tenantPath), resourceUsage(test, fixture, keyPath)
			assertRejected := func() {
				beforeHits := fixture.hits.Load()
				receipt := resourceGenerate(fixture, fixture.key)
				if receipt.err != nil || receipt.status != 429 || fixture.hits.Load() != beforeHits {
					test.Fatalf("zero soft budget: status=%d dispatch_delta=%d error=%v", receipt.status, fixture.hits.Load()-beforeHits, receipt.err)
				}
				if resourceUsage(test, fixture, tenantPath) != beforeTenant || resourceUsage(test, fixture, keyPath) != beforeKey {
					test.Fatal("rejection changed fees/reservations/usage")
				}
			}
			assertRejected()
			fixture.clock.Add(int64(time.Second))
			if status, _ := fixture.admin.call("POST", "models/governed/prices", map[string]any{"prices": []map[string]any{{"upstream_id": 3, "cost_input_rate": 0, "cost_output_rate": 0, "tenant_input_rate": 0, "tenant_output_rate": 0}}}, "zero-price"); status != 201 {
				test.Fatal("zero price", status)
			}
			assertRejected()
			update(nil)
			if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 200 || receipt.err != nil {
				test.Fatal("null allows free call", receipt)
			}
			if usage := resourceUsage(test, fixture, keyPath); usage.Fee != 0 {
				test.Fatal("zero-priced accepted call charged", usage)
			}
			fixture.clock.Add(int64(time.Second))
			if status, _ := fixture.admin.call("POST", "models/governed/prices", map[string]any{"prices": []map[string]any{{"upstream_id": 3, "cost_input_rate": 1000000, "cost_output_rate": 1000000, "tenant_input_rate": 1000000, "tenant_output_rate": 1000000}}}, "paid-price"); status != 201 {
				test.Fatal("paid price", status)
			}
			update(ptrInt64(1))
			for range 2 {
				if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 200 || receipt.err != nil {
					test.Fatal("positive soft budget must not become hard", receipt)
				}
			}
			usage := resourceUsage(test, fixture, keyPath)
			if usage.Calls != 3 || usage.Fee != 14 || usage.Reserved != 0 || usage.Pending != 0 || fixture.hits.Load() != 3 {
				test.Fatal("soft/null fee conservation", usage, fixture.hits.Load())
			}
			if status, report := fixture.admin.call("GET", path+"/limits", nil, ""); status != 200 || report["soft_budget_exceeded"] != true {
				test.Fatal("positive soft budget warning", status)
			}
		})
	}
}

func TestConsoleReviewClockRollbackInProcess(test *testing.T) {
	test.Parallel()
	reviewClockRollback(test, false)
}
func TestConsoleReviewClockRollbackActualHTTP(test *testing.T) {
	test.Parallel()
	reviewClockRollback(test, true)
}

func reviewClockRollback(test *testing.T, actual bool) {
	fixture := singleResourceInstallation(test, actual)
	start := time.Unix(0, fixture.clock.Load())
	path := "tenants/" + fixture.tenant + "/keys/" + fixture.keyID
	status, rotation := fixture.admin.call("POST", path+"/rotations", map[string]any{"expected_revision": 1, "transition_hours": 1}, "review-rotation")
	if status != 200 {
		test.Fatal("rotation", status)
	}
	fixture.clock.Store(start.Add(time.Hour).UnixNano())
	if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 401 || receipt.err != nil {
		test.Fatal("exact deadline", receipt)
	}
	fixture.clock.Store(start.Add(30 * time.Minute).UnixNano())
	if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 401 || receipt.err != nil || fixture.hits.Load() != 0 {
		test.Fatal("expired unretired generation revived after rollback", receipt, fixture.hits.Load())
	}
	if receipt := resourceGenerate(fixture, rotation["key"].(string)); receipt.status != 200 || receipt.err != nil {
		test.Fatal("current generation remains usable", receipt)
	}
}
