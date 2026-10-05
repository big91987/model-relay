package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const baseline = "9b6994ed80de2bdba79ff267419150dccaac1360"
const rootTest = "TestHealthMethodsHTTP"

type testEvent struct {
	Action string
	Test   string
	Output string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "historical HTTP reproduction rejected:", err)
		os.Exit(1)
	}
}

func run() error {
	workspace, err := os.Getwd()
	if err != nil {
		return err
	}
	directory, err := os.MkdirTemp("", "relay-health-history-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	archive, err := exec.Command("git", "archive", baseline).Output()
	if err != nil {
		return fmt.Errorf("baseline archive unavailable: %w", err)
	}
	extract := exec.Command("tar", "-x", "-C", directory)
	extract.Stdin = bytes.NewReader(archive)
	if err := extract.Run(); err != nil {
		return fmt.Errorf("baseline extraction: %w", err)
	}
	testSource, err := os.ReadFile("internal/relay/health_test.go")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "internal/relay/health_test.go"), testSource, 0600); err != nil {
		return err
	}
	serverSource, err := os.ReadFile(filepath.Join(directory, "internal/relay/server.go"))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-json", "-race", "-count=1", "-run", "^"+rootTest+"$", "./internal/relay")
	command.Dir = directory
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runErr := command.Run()
	exitCode := 0
	if runErr != nil {
		exitCode = -1
		if failure, ok := runErr.(*exec.ExitError); ok {
			exitCode = failure.ExitCode()
		}
	}
	assessment := assess(stdout.Bytes(), exitCode)
	if ctx.Err() != nil {
		assessment = fmt.Errorf("historical test timeout: %w", ctx.Err())
	}
	var receipt bytes.Buffer
	fmt.Fprintf(&receipt, "修复后隔离重建旧版网络复现（非原开发时成功记录）\nbaseline=%s\nbaseline_archive_sha256=%x\nserver_sha256=%x\nfinal_test_sha256=%x\ncommand=go test -json -race -count=1 -run ^TestHealthMethodsHTTP$ ./internal/relay\nexit_code=%d\n", baseline, sha256.Sum256(archive), sha256.Sum256(serverSource), sha256.Sum256(testSource), exitCode)
	receipt.Write(stdout.Bytes())
	receipt.Write(stderr.Bytes())
	if assessment != nil {
		fmt.Fprintln(&receipt, "REJECTED:", assessment)
	} else {
		fmt.Fprintln(&receipt, "PASS expected historical HTTP red: 21 method failures with exact status/Allow diagnostics; 12 query/recovery passes")
	}
	redact := strings.NewReplacer(directory, "<isolated-history>", workspace, "<workspace>", "/private"+os.TempDir(), "<temp>/", os.TempDir(), "<temp>/")
	cleanReceipt := redact.Replace(receipt.String())
	if _, err := fmt.Print(cleanReceipt); err != nil {
		return fmt.Errorf("write historical HTTP receipt to stdout: %w", err)
	}
	return assessment
}

func assess(data []byte, exitCode int) error {
	if exitCode != 1 {
		return fmt.Errorf("expected assertion exit 1, got %d", exitCode)
	}
	expected := map[string]string{rootTest: "fail"}
	diagnostics := map[string]map[string]bool{}
	for _, state := range []string{"healthy", "degraded", "database_closed"} {
		parent := rootTest + "/" + state
		expected[parent] = "fail"
		for _, method := range []string{"GET", "HEAD", "recovered_GET", "recovered_HEAD"} {
			expected[parent+"/"+method] = "pass"
		}
		status := 503
		if state == "healthy" {
			status = 200
		}
		for _, method := range []string{"POST", "PUT", "PATCH", "DELETE", "OPTIONS", "TRACE", "CUSTOM"} {
			name := parent + "/" + method
			expected[name] = "fail"
			diagnostics[name] = map[string]bool{
				fmt.Sprintf("%s status = %d, want 405", method, status): false,
				method + ` Allow = "", want GET, HEAD`:                  false,
			}
		}
	}
	observed := map[string]string{}
	diagnosticLine := regexp.MustCompile(`^\s*\S+_test\.go:\d+: (.*)$`)
	linkerWarning := regexp.MustCompile(`^# model-relay/internal/relay\.test\nld: warning: '[^'\n]+' has malformed LC_DYSYMTAB, expected \d+ undefined symbols to start at index \d+, found \d+ undefined symbols starting at index \d+\n$`)
	decoder := json.NewDecoder(bytes.NewReader(data))
	packageFailed := false
	for {
		var event testEvent
		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("invalid test JSON: %w", err)
		}
		if event.Action == "build-output" && linkerWarning.MatchString(event.Output) {
			continue
		}
		if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			if event.Test == "" {
				if event.Action != "fail" || packageFailed {
					return fmt.Errorf("unexpected package result %q", event.Action)
				}
				packageFailed = true
			} else {
				if expected[event.Test] != event.Action || observed[event.Test] != "" {
					return fmt.Errorf("unexpected test result %s: %s", event.Test, event.Action)
				}
				observed[event.Test] = event.Action
			}
		}
		if match := diagnosticLine.FindStringSubmatch(strings.TrimSpace(event.Output)); match != nil {
			allowed := diagnostics[event.Test]
			seen, exists := allowed[match[1]]
			if !exists || seen {
				return fmt.Errorf("unexpected or duplicate assertion in %s: %s", event.Test, match[1])
			}
			allowed[match[1]] = true
		} else if event.Output != "" {
			line := strings.TrimSpace(event.Output)
			if !strings.HasPrefix(line, "=== RUN ") && !strings.HasPrefix(line, "--- PASS: ") && !strings.HasPrefix(line, "--- FAIL: ") && line != "FAIL" && !strings.HasPrefix(line, "FAIL\tmodel-relay/internal/relay\t") {
				return fmt.Errorf("unexpected test output in %s", event.Test)
			}
		}
	}
	if !packageFailed || len(observed) != len(expected) {
		return fmt.Errorf("incomplete network journey: got %d/%d results", len(observed), len(expected))
	}
	for name, messages := range diagnostics {
		for message, seen := range messages {
			if !seen {
				return fmt.Errorf("missing historical assertion %s: %s", name, message)
			}
		}
	}
	return nil
}
