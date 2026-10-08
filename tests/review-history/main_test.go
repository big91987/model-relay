package main

import "testing"

func TestHistoricalReviewAssessment(test *testing.T) {
	failed := []string{"TestConsoleReviewZeroSoftBudgetActualHTTP/tenant", "TestConsoleReviewZeroSoftBudgetActualHTTP/key", "TestConsoleReviewZeroSoftBudgetActualHTTP", "TestConsoleReviewClockRollbackActualHTTP"}
	evidence := "zero soft budget: status=200 dispatch_delta=1 error=<nil>\nzero soft budget: status=200 dispatch_delta=1 error=<nil>\nexpired unretired generation revived after rollback {200 controlled_request_id <nil>} 1"
	if err := assess(failed, evidence); err != nil {
		test.Fatal(err)
	}
	for _, invalid := range []string{evidence + "DATA RACE", evidence + "real HTTP listener unavailable", "build failed", ""} {
		if assess(failed, invalid) == nil {
			test.Fatal("invalid evidence accepted")
		}
	}
	if assess(failed[:3], evidence) == nil || assess(append(failed, failed[0]), evidence) == nil {
		test.Fatal("missing/duplicate assertion accepted")
	}
}
