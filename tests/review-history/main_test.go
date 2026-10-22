package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewCommandHelper(test *testing.T) {
	if os.Getenv("REVIEW_HISTORY_HELPER") != "1" {
		return
	}
	main()
}

func TestReviewCommandSeparatesDiagnostics(test *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	for _, scenario := range []string{"expected_red_with_stderr", "invalid_stdout", "missing_assertions", "stderr_race", "stderr_build_failure"} {
		test.Run(scenario, func(test *testing.T) {
			root := test.TempDir()
			var events bytes.Buffer
			encoder := json.NewEncoder(&events)
			for _, name := range []string{"TestConsoleReviewZeroSoftBudgetActualHTTP/tenant", "TestConsoleReviewZeroSoftBudgetActualHTTP/key", "TestConsoleReviewZeroSoftBudgetActualHTTP", "TestConsoleReviewClockRollbackActualHTTP"} {
				if err := encoder.Encode(map[string]string{"Action": "fail", "Test": name}); err != nil {
					test.Fatal(err)
				}
			}
			if scenario != "missing_assertions" {
				if err := encoder.Encode(map[string]string{"Action": "output", "Output": "zero soft budget: status=200 dispatch_delta=1 error=<nil>\nzero soft budget: status=200 dispatch_delta=1 error=<nil>\nexpired unretired generation revived after rollback {200 controlled_request_id <nil>} 1\n"}); err != nil {
					test.Fatal(err)
				}
			}
			if scenario == "invalid_stdout" {
				events.WriteString("invalid JSON\n")
			}
			diagnostic := "controlled compiler diagnostic"
			if scenario == "stderr_race" {
				diagnostic += " DATA RACE"
			}
			if scenario == "stderr_build_failure" {
				diagnostic += " build failed"
			}
			files := map[string]string{
				"internal/relay/console_review_regression_test.go": "package relay\n",
				"events.json": events.String(),
				"tools/git":   "#!/bin/sh\n[ \"$*\" = 'archive " + baseline + "' ] || exit 99\nprintf 'archive-fixture'\n",
				"tools/tar":   "#!/bin/sh\n[ \"$1\" = '-x' ] && [ \"$2\" = '-C' ] || exit 99\nmkdir -p \"$3/internal/relay\"\n",
				"tools/go":    "#!/bin/sh\n[ \"$*\" = 'test -json -race -count=1 -run ^TestConsoleReview.*ActualHTTP$ ./internal/relay' ] || exit 99\nprintf '%s\\n' \"$REVIEW_HISTORY_DIAGNOSTIC\" >&2\ncat \"$REVIEW_HISTORY_EVENTS\"\nexit 1\n",
			}
			for name, content := range files {
				path := filepath.Join(root, name)
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					test.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0700); err != nil {
					test.Fatal(err)
				}
			}
			command := exec.Command(executable, "-test.run=^TestReviewCommandHelper$")
			command.Dir = root
			command.Env = append(os.Environ(), "REVIEW_HISTORY_HELPER=1", "REVIEW_HISTORY_EVENTS="+filepath.Join(root, "events.json"), "REVIEW_HISTORY_DIAGNOSTIC="+diagnostic, "PATH="+filepath.Join(root, "tools")+string(os.PathListSeparator)+os.Getenv("PATH"))
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			runErr := command.Run()
			if !strings.Contains(stderr.String(), "controlled compiler diagnostic") {
				test.Fatalf("child stderr was lost: %s", &stderr)
			}
			if scenario == "expected_red_with_stderr" {
				if runErr != nil || !strings.Contains(stdout.String(), "Exact historical assertions observed") {
					test.Fatalf("valid JSON rejected after stderr diagnostic: %v stdout=%s stderr=%s", runErr, &stdout, &stderr)
				}
			} else if runErr == nil || !strings.Contains(stderr.String(), "review historical HTTP rejected:") || strings.Contains(stdout.String(), "Exact historical assertions observed") {
				test.Fatalf("invalid evidence accepted: %v stdout=%s stderr=%s", runErr, &stdout, &stderr)
			}
		})
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
