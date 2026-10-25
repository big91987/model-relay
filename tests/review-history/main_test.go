package main

import (
	"encoding/json"
	"os/exec"
	"testing"
)

func TestHistoricalJSONExcludesDownloadDiagnostics(test *testing.T) {
	for _, exit := range []string{"0", "1"} {
		command := exec.Command("/bin/sh", "-c", `printf 'go: downloading controlled-module v1.0.0\n' >&2; printf '{"Action":"fail","Test":"controlled"}\n'; exit `+exit)
		output, diagnostics, err := historicalOutput(command)
		var event struct{ Action, Test string }
		if json.Unmarshal(output, &event) != nil || event.Action != "fail" || event.Test != "controlled" {
			test.Fatalf("stdout JSON polluted: %q", output)
		}
		if string(diagnostics) != "go: downloading controlled-module v1.0.0\n" || (err != nil) != (exit == "1") {
			test.Fatalf("lost diagnostics or exit: %q %v", diagnostics, err)
		}
	}
}

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
