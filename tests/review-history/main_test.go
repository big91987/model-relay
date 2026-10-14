package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoricalCommandStreams(test *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		test.Fatal(err)
	}
	tools := test.TempDir()
	for _, tool := range []string{"git", "tar", "go"} {
		script := "#!/bin/sh\nexec \"$REVIEW_HELPER_EXECUTABLE\" -test.run '^TestHistoricalCommandHelper$' -- " + tool + " \"$@\"\n"
		if err := os.WriteFile(filepath.Join(tools, tool), []byte(script), 0700); err != nil {
			test.Fatal(err)
		}
	}
	cases := []struct {
		name      string
		mode      string
		wantError string
	}{
		{name: "expected historical red", mode: "valid"},
		{name: "download stderr preserves expected red", mode: "download"},
		{name: "invalid stdout rejected with diagnostics", mode: "invalid", wantError: "invalid-stdout-marker"},
		{name: "execution failure is not expected red", mode: "execution", wantError: "historical execution failed"},
		{name: "successful baseline rejected", mode: "success", wantError: "baseline did not fail"},
		{name: "missing historical assertion rejected", mode: "missing", wantError: "missing historical assertion"},
		{name: "build failure on stderr rejected", mode: "build", wantError: "historical build failed"},
		{name: "race on stderr rejected", mode: "race", wantError: "historical race"},
		{name: "listener failure on stderr rejected", mode: "listener", wantError: "historical listener unavailable"},
	}
	for _, scenario := range cases {
		test.Run(scenario.name, func(test *testing.T) {
			command := exec.Command(executable, "-test.run", "^TestHistoricalCommandHelper$", "--", "review")
			command.Dir = root
			command.Env = append(os.Environ(), "PATH="+tools+string(os.PathListSeparator)+os.Getenv("PATH"), "REVIEW_HELPER_EXECUTABLE="+executable, "REVIEW_HELPER_MODE="+scenario.mode)
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			runErr := command.Run()
			if scenario.wantError == "" {
				if runErr != nil || !strings.Contains(stdout.String(), "Exact historical assertions observed") {
					test.Fatalf("expected historical red accepted: err=%v stdout=%s stderr=%s", runErr, &stdout, &stderr)
				}
			} else if runErr == nil || !strings.Contains(stderr.String(), scenario.wantError) {
				test.Fatalf("expected rejection %q: err=%v stdout=%s stderr=%s", scenario.wantError, runErr, &stdout, &stderr)
			}
			if !strings.Contains(stderr.String(), "diagnostic-first-marker") || !strings.Contains(stderr.String(), "diagnostic-last-marker") {
				test.Fatalf("full stderr diagnostics not retained: %s", &stderr)
			}
		})
	}
}

func TestHistoricalCommandHelper(test *testing.T) {
	if os.Getenv("REVIEW_HELPER_EXECUTABLE") == "" {
		return
	}
	separator := 0
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index + 1
			break
		}
	}
	if separator == 0 || separator >= len(os.Args) {
		os.Exit(2)
	}
	switch os.Args[separator] {
	case "review":
		main()
	case "git":
		fmt.Print("isolated command fixture archive")
	case "tar":
		for index := separator + 1; index+1 < len(os.Args); index++ {
			if os.Args[index] == "-C" {
				if err := os.MkdirAll(filepath.Join(os.Args[index+1], "internal/relay"), 0700); err != nil {
					os.Exit(2)
				}
			}
		}
	case "go":
		fmt.Fprintln(os.Stderr, "diagnostic-first-marker")
		mode := os.Getenv("REVIEW_HELPER_MODE")
		switch mode {
		case "download":
			fmt.Fprintln(os.Stderr, "go: downloading isolated.example/test-fixture v0.0.0")
		case "invalid":
			fmt.Println("invalid-stdout-marker")
		case "build":
			fmt.Fprintln(os.Stderr, "build failed")
		case "race":
			fmt.Fprintln(os.Stderr, "DATA RACE")
		case "listener":
			fmt.Fprintln(os.Stderr, "real HTTP listener unavailable")
		}
		encoder := json.NewEncoder(os.Stdout)
		for _, name := range []string{"TestConsoleReviewZeroSoftBudgetActualHTTP/tenant", "TestConsoleReviewZeroSoftBudgetActualHTTP/key", "TestConsoleReviewZeroSoftBudgetActualHTTP", "TestConsoleReviewClockRollbackActualHTTP"} {
			if mode == "missing" && name == "TestConsoleReviewClockRollbackActualHTTP" {
				continue
			}
			output := ""
			if strings.Contains(name, "/") {
				output = "zero soft budget: status=200 dispatch_delta=1 error=<nil>\n"
			} else if name == "TestConsoleReviewClockRollbackActualHTTP" {
				output = "expired unretired generation revived after rollback {200 controlled_request_id <nil>} 1\n"
			}
			if err := encoder.Encode(map[string]string{"Action": "output", "Test": name, "Output": output}); err != nil {
				os.Exit(2)
			}
			if err := encoder.Encode(map[string]string{"Action": "fail", "Test": name}); err != nil {
				os.Exit(2)
			}
		}
		fmt.Fprintln(os.Stderr, "diagnostic-last-marker")
		if mode == "success" {
			os.Exit(0)
		}
		if mode == "execution" {
			os.Exit(2)
		}
		os.Exit(1)
	default:
		os.Exit(2)
	}
	os.Exit(0)
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
