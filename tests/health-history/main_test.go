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

func TestHistoryCommandHelper(t *testing.T) {
	if os.Getenv("HISTORY_TEST_HELPER") != "1" {
		return
	}
	if os.Getenv("HISTORY_TEST_CLOSE_STDOUT") == "1" {
		if err := os.Stdout.Close(); err != nil {
			t.Fatal(err)
		}
	}
	main()
}

func TestHistoryCommandOutput(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	archiveCommand := exec.Command("git", "archive", baseline)
	archiveCommand.Dir = root
	archive, err := archiveCommand.Output()
	if err != nil {
		t.Fatal(err)
	}
	testSource, err := os.ReadFile(filepath.Join(root, "internal/relay/health_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"expected_red", "rejected", "closed_stdout"} {
		t.Run(scenario, func(t *testing.T) {
			directory := t.TempDir()
			files := map[string][]byte{
				"baseline.tar":                             archive,
				"internal/relay/health_test.go":            testSource,
				"docs/workflow/health-methods-history.log": []byte("preserved historical receipt\n"),
				"docs/workflow/runs/prior/tasks/task.md":   []byte("preserved historical task\n"),
				"tools/git":                                []byte("#!/bin/sh\n[ \"$*\" = \"archive " + baseline + "\" ] || exit 99\ncat \"$HISTORY_TEST_ARCHIVE\"\n"),
				"tools/go":                                 []byte("#!/bin/sh\n[ \"$*\" = \"test -json -race -count=1 -run ^TestHealthMethodsHTTP$ ./internal/relay\" ] || exit 99\ncat \"$HISTORY_TEST_EVENTS\"\nexit 1\n"),
			}
			events := historicalEvents()
			if scenario == "rejected" {
				events = []testEvent{{Action: "output", Output: "listen tcp 127.0.0.1:0: bind: operation not permitted\n"}, {Action: "fail"}}
			}
			files["events.json"] = encodeEvents(t, events)
			for path, data := range files {
				target := filepath.Join(directory, path)
				if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, data, 0700); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.Command(executable, "-test.run=^TestHistoryCommandHelper$")
			command.Dir = directory
			command.Env = append(os.Environ(),
				"HISTORY_TEST_HELPER=1",
				"PATH="+filepath.Join(directory, "tools")+string(os.PathListSeparator)+os.Getenv("PATH"),
				"HISTORY_TEST_ARCHIVE="+filepath.Join(directory, "baseline.tar"),
				"HISTORY_TEST_EVENTS="+filepath.Join(directory, "events.json"),
				"HISTORY_TEST_CLOSE_STDOUT=0")
			if scenario == "closed_stdout" {
				command.Env = append(command.Env, "HISTORY_TEST_CLOSE_STDOUT=1")
			}
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			runErr := command.Run()
			if scenario == "expected_red" {
				if runErr != nil || !strings.Contains(stdout.String(), "PASS expected historical HTTP red") || !strings.Contains(stdout.String(), "baseline="+baseline) || stderr.Len() != 0 {
					t.Fatalf("expected classified red receipt: error=%v stdout=%s stderr=%s", runErr, &stdout, &stderr)
				}
			} else {
				failure, ok := runErr.(*exec.ExitError)
				if !ok || failure.ExitCode() != 1 || !strings.Contains(stderr.String(), "historical HTTP reproduction rejected:") {
					t.Fatalf("failure not propagated: error=%v stdout=%s stderr=%s", runErr, &stdout, &stderr)
				}
				if scenario == "rejected" && (!strings.Contains(stdout.String(), "REJECTED:") || strings.Contains(stdout.String(), "PASS expected historical HTTP red")) {
					t.Fatalf("missing rejected receipt: %s", &stdout)
				}
				if scenario == "closed_stdout" && !strings.Contains(stderr.String(), "write historical HTTP receipt to stdout:") {
					t.Fatalf("missing output diagnostic: %s", &stderr)
				}
			}
			for path, expected := range files {
				if !strings.HasPrefix(path, "docs/") {
					continue
				}
				actual, err := os.ReadFile(filepath.Join(directory, path))
				if err != nil || !bytes.Equal(actual, expected) {
					t.Errorf("historical file changed: %s (error=%v)", path, err)
				}
			}
		})
	}
}

func historicalEvents() []testEvent {
	var events []testEvent
	for _, state := range []string{"healthy", "degraded", "database_closed"} {
		parent := rootTest + "/" + state
		for _, method := range []string{"GET", "HEAD", "recovered_GET", "recovered_HEAD"} {
			events = append(events, testEvent{Action: "pass", Test: parent + "/" + method})
		}
		status := 503
		if state == "healthy" {
			status = 200
		}
		for _, method := range []string{"POST", "PUT", "PATCH", "DELETE", "OPTIONS", "TRACE", "CUSTOM"} {
			name := parent + "/" + method
			events = append(events,
				testEvent{Action: "output", Test: name, Output: fmt.Sprintf("    health_test.go:117: %s status = %d, want 405\n", method, status)},
				testEvent{Action: "output", Test: name, Output: fmt.Sprintf("    health_test.go:117: %s Allow = \"\", want GET, HEAD\n", method)},
				testEvent{Action: "fail", Test: name})
		}
		events = append(events, testEvent{Action: "fail", Test: parent})
	}
	return append(events, testEvent{Action: "fail", Test: rootTest}, testEvent{Action: "fail"})
}

func encodeEvents(t *testing.T, events []testEvent) []byte {
	t.Helper()
	var data bytes.Buffer
	for _, event := range events {
		if err := json.NewEncoder(&data).Encode(event); err != nil {
			t.Fatal(err)
		}
	}
	return data.Bytes()
}

func TestHistoricalAssessment(t *testing.T) {
	if err := assess(encodeEvents(t, historicalEvents()), 1); err != nil {
		t.Fatal(err)
	}
	warning := testEvent{Action: "build-output", Output: "# model-relay/internal/relay.test\nld: warning: '<temp>/000023.o' has malformed LC_DYSYMTAB, expected 98 undefined symbols to start at index 1626, found 95 undefined symbols starting at index 1626\n"}
	if err := assess(encodeEvents(t, append([]testEvent{warning}, historicalEvents()...)), 1); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"bind", "build", "warning_with_error", "timeout", "green", "missing_assertion", "extra_assertion", "race", "wrong_status", "query_failure", "skip", "duplicate", "missing_package", "invalid_json"} {
		t.Run(scenario, func(t *testing.T) {
			events := historicalEvents()
			exitCode := 1
			switch scenario {
			case "bind":
				events = []testEvent{{Action: "output", Test: rootTest + "/healthy", Output: "relay_test.go:39: listen tcp 127.0.0.1:0: bind: operation not permitted\n"}, {Action: "fail", Test: rootTest + "/healthy"}, {Action: "fail"}}
			case "build":
				events = []testEvent{{Action: "fail"}}
			case "warning_with_error":
				warning.Output += "build failed\n"
				events = append(events, warning)
			case "timeout":
				exitCode = -1
			case "green":
				exitCode = 0
			case "missing_assertion":
				events = append(events[:4], events[5:]...)
			case "extra_assertion":
				events = append(events, testEvent{Action: "output", Test: rootTest + "/healthy/POST", Output: "health_test.go:99: unexpected failure\n"})
			case "race":
				events = append(events, testEvent{Action: "output", Test: rootTest + "/healthy/POST", Output: "WARNING: DATA RACE\n"})
			case "wrong_status":
				events[4].Output = "health_test.go:117: POST status = 500, want 405\n"
			case "query_failure":
				events[0].Action = "fail"
			case "skip":
				events[0].Action = "skip"
			case "duplicate":
				events = append(events, events[0])
			case "missing_package":
				events = events[:len(events)-1]
			}
			data := encodeEvents(t, events)
			if scenario == "invalid_json" {
				data = append(data, []byte("invalid")...)
			}
			if err := assess(data, exitCode); err == nil {
				t.Fatal("unexpected failure accepted as historical network red")
			}
		})
	}
}
