package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const baseline = "66f1328ada75ea90e5638fbf2bd58663ac86347e"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "review historical HTTP rejected:", err)
		os.Exit(1)
	}
}

func run() error {
	directory, err := os.MkdirTemp("", "relay-review-history-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	archive, err := exec.Command("git", "archive", baseline).Output()
	if err != nil {
		return fmt.Errorf("baseline unavailable: %w", err)
	}
	extract := exec.Command("tar", "-x", "-C", directory)
	extract.Stdin = bytes.NewReader(archive)
	if err := extract.Run(); err != nil {
		return err
	}
	source, err := os.ReadFile("internal/relay/console_review_regression_test.go")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "internal/relay/console_review_regression_test.go"), source, 0600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-json", "-race", "-count=1", "-run", "^TestConsoleReview.*ActualHTTP$", "./internal/relay")
	command.Dir = directory
	var output, diagnostics bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostics
	runErr := command.Run()
	fmt.Fprint(os.Stderr, diagnostics.String())
	reject := func(err error) error {
		fmt.Fprintf(os.Stderr, "historical stdout:\n%s", output.Bytes())
		return err
	}
	var failed []string
	observations := []map[string]string{}
	var evidence strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(output.Bytes()))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		var event struct{ Action, Test, Output string }
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return reject(fmt.Errorf("unparsed baseline output: %w", err))
		}
		if event.Action == "fail" && event.Test != "" {
			failed = append(failed, event.Test)
		}
		evidence.WriteString(event.Output)
		if event.Action == "output" && (strings.Contains(event.Output, "zero soft budget:") || strings.Contains(event.Output, "expired unretired generation revived after rollback")) {
			observations = append(observations, map[string]string{"test": event.Test, "actual_assertion_output": strings.TrimSpace(event.Output)})
		}
	}
	if err := scanner.Err(); err != nil {
		return reject(err)
	}
	if ctx.Err() != nil {
		return reject(fmt.Errorf("historical test timed out"))
	}
	if runErr == nil {
		return reject(fmt.Errorf("baseline did not fail"))
	}
	var exitError *exec.ExitError
	if !errors.As(runErr, &exitError) || exitError.ExitCode() != 1 {
		return reject(fmt.Errorf("historical execution failed; not business red: %w", runErr))
	}
	if err := assess(failed, evidence.String()); err != nil {
		return reject(err)
	}
	if err := assess(failed, evidence.String()+diagnostics.String()); err != nil {
		return reject(err)
	}
	fmt.Printf("historical controlled actual HTTP reconstruction (not original development-time receipt) baseline=%s archive_sha256=%x regression_sha256=%x exit_code=1\n", baseline, sha256.Sum256(archive), sha256.Sum256(source))
	if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"historical_actual_http_assertions": observations}); err != nil {
		return err
	}
	fmt.Println("P1-1 tenant/key budget=0 hard=false: status=200 dispatch_delta=1; P1-2 expired unretired old generation after deadline401 and clock rollback: status=200 dispatch=1. Exact historical assertions observed; current HTTP green still required.")
	return nil
}

func assess(failed []string, evidence string) error {
	if strings.Contains(evidence, "real HTTP listener unavailable") {
		return fmt.Errorf("historical listener unavailable; not business red")
	}
	if strings.Contains(evidence, "DATA RACE") {
		return fmt.Errorf("historical race; not business red")
	}
	if strings.Contains(evidence, "build failed") {
		return fmt.Errorf("historical build failed; not business red")
	}
	expected := map[string]bool{"TestConsoleReviewZeroSoftBudgetActualHTTP": false, "TestConsoleReviewZeroSoftBudgetActualHTTP/tenant": false, "TestConsoleReviewZeroSoftBudgetActualHTTP/key": false, "TestConsoleReviewClockRollbackActualHTTP": false}
	for _, name := range failed {
		seen, exists := expected[name]
		if !exists || seen {
			return fmt.Errorf("unexpected or duplicate historical failure")
		}
		expected[name] = true
	}
	for _, present := range expected {
		if !present {
			return fmt.Errorf("missing historical assertion")
		}
	}
	if strings.Count(evidence, "zero soft budget: status=200 dispatch_delta=1 error=<nil>") != 2 || !strings.Contains(evidence, "expired unretired generation revived after rollback {200 ") || strings.Contains(evidence, "DATA RACE") || strings.Contains(evidence, "real HTTP listener unavailable") || strings.Contains(evidence, "build failed") {
		return fmt.Errorf("wrong cause of historical failures")
	}
	return nil
}
