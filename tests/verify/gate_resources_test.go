package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestVerificationGateBudgetAndCancellation(test *testing.T) {
	if verificationGateBudget != 24*time.Minute {
		test.Fatal("production cumulative budget must remain fixed at 24 minutes")
	}
	for _, scenario := range []string{"success", "nonzero", "short-budget", "parent-deadline", "parent-cancelled", "zero", "negative", "over-limit"} {
		test.Run(scenario, func(test *testing.T) {
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			budget := verificationGateBudget
			script := "printf 'gate-stdout'; printf 'gate-stderr' >&2"
			ready := filepath.Join(test.TempDir(), "child")
			switch scenario {
			case "nonzero":
				script += "; exit 23"
			case "short-budget", "parent-deadline":
				script = "sleep 30 & child=$!; printf '%s' \"$child\" > \"$READY\"; wait"
				if scenario == "short-budget" {
					budget = 100 * time.Millisecond
				} else {
					var deadlineCancel context.CancelFunc
					parent, deadlineCancel = context.WithTimeout(parent, 100*time.Millisecond)
					defer deadlineCancel()
				}
			case "parent-cancelled":
				cancel()
			case "zero":
				budget = 0
			case "negative":
				budget = -time.Second
			case "over-limit":
				budget++
			}
			var output bytes.Buffer
			state := receipt{GateExit: -1}
			started := time.Now()
			err := executeVerificationGate(parent, exec.Command("/bin/sh", "-c", script), "", append(os.Environ(), "READY="+ready), &output, &state, budget)
			if scenario == "zero" || scenario == "negative" || scenario == "over-limit" {
				if err == nil || state.Gate != nil || output.Len() != 0 || state.GateExit != -1 {
					test.Fatal("invalid budget started a command")
				}
				return
			}
			if state.Gate == nil || state.GateBudgetMS != budget.Milliseconds() || !strings.Contains(output.String(), "GATE_START") || !strings.Contains(output.String(), "GATE_END") {
				test.Fatal("missing auditable gate budget/outcome")
			}
			if scenario == "success" || scenario == "nonzero" {
				wanted := 0
				if scenario == "nonzero" {
					wanted = 23
				}
				if (err != nil) != (wanted != 0) || state.GateExit != wanted || state.Gate.Cancelled || state.OwnedGroupAbsent == nil || !*state.OwnedGroupAbsent || !strings.Contains(output.String(), "gate-stdout") || !strings.Contains(output.String(), "gate-stderr") {
					test.Fatalf("gate semantics lost: state=%+v err=%v", state.Gate, err)
				}
				return
			}
			wanted := context.DeadlineExceeded
			if scenario == "parent-cancelled" {
				wanted = context.Canceled
			}
			if !errors.Is(err, wanted) || state.GateExit == 0 || !state.Gate.Cancelled || time.Since(started) > 8*time.Second || state.Gate.BudgetMilliseconds > 100 {
				test.Fatalf("gate not bounded/cancelled: %+v err=%v", state.Gate, err)
			}
			if scenario != "parent-cancelled" {
				if state.OwnedGroupAbsent == nil || !*state.OwnedGroupAbsent {
					test.Fatal("gate left its process group")
				}
				data, readErr := os.ReadFile(ready)
				var child int
				if readErr != nil {
					test.Fatal("real child did not start before cancellation")
				}
				if _, scanErr := fmt.Sscanf(string(data), "%d", &child); scanErr != nil || syscall.Kill(child, 0) != syscall.ESRCH {
					test.Fatal("cancelled child remains")
				}
			}
		})
	}
}

func resourceFixtureRead(test *testing.T, overrides map[string]string, failed string) func(string) ([]byte, error) {
	test.Helper()
	values := map[string]string{
		"cgroup": "0::/\n", "status": "Cpus_allowed_list:\t0-111\n",
		"cpu.max": "200000 100000", "memory.max": "8589934592", "cpuset.cpus.effective": "0-111",
		"cpu.stat":      "usage_usec 10\nuser_usec 6\nsystem_usec 4\nnr_periods 3\nnr_throttled 2\nthrottled_usec 7\n",
		"memory.events": "low 0\nhigh 0\nmax 0\noom 0\noom_kill 0\noom_group_kill 0\n",
	}
	for key, value := range overrides {
		values[key] = value
	}
	return func(path string) ([]byte, error) {
		name := filepath.Base(path)
		if name == failed {
			return nil, errors.New("controlled read failure; must not expose this path")
		}
		value, exists := values[name]
		if !exists {
			test.Fatalf("unexpected resource read: %s", name)
		}
		return []byte(value), nil
	}
}

func TestResourceSamplingAndDelta(test *testing.T) {
	before := readResourcesForPlatform("linux", resourceFixtureRead(test, nil, ""))
	after := readResourcesForPlatform("linux", resourceFixtureRead(test, map[string]string{"cpu.stat": "usage_usec 15\nuser_usec 9\nsystem_usec 6\nnr_periods 4\nnr_throttled 3\nthrottled_usec 11\n"}, ""))
	delta, err := resourceDifference(before, after)
	if resourceSampleError("before", before) != nil || err != nil || delta.Counters["cpu.stat.usage_usec"] != 5 || delta.Counters["cpu.stat.throttled_usec"] != 4 || delta.FromUTC != before.FinishedUTC || delta.ToUTC != after.FinishedUTC || !strings.Contains(delta.Scope, "not per-process") || before.RuntimeCPUs < 1 || before.GOMAXPROCS < 1 {
		test.Fatalf("invalid resource/delta receipt: %+v err=%v", delta, err)
	}
	for _, scenario := range []string{"read-failure", "bad-quota", "bad-counter", "missing-counter", "decrease", "changed-cgroup", "changed-limit"} {
		test.Run(scenario, func(test *testing.T) {
			overrides := map[string]string{}
			failed := ""
			switch scenario {
			case "read-failure":
				failed = "cpu.stat"
			case "bad-quota":
				overrides["cpu.max"] = "0 100000"
			case "bad-counter":
				overrides["memory.events"] = "oom_kill invalid"
			case "missing-counter":
				overrides["cpu.stat"] = "usage_usec 1"
			case "decrease":
				overrides["cpu.stat"] = strings.Replace(before.Values["cpu.stat"], "usage_usec 10", "usage_usec 9", 1)
			case "changed-cgroup":
				overrides["cgroup"] = "0::/other"
			case "changed-limit":
				overrides["cpu.max"] = "400000 100000"
			}
			changed := readResourcesForPlatform("linux", resourceFixtureRead(test, overrides, failed))
			if result, deltaErr := resourceDifference(before, changed); result != nil || deltaErr == nil {
				test.Fatal("invalid samples produced a misleading delta")
			}
		})
	}
}

func TestResourceUnsupportedPlatform(test *testing.T) {
	before := readResourcesForPlatform("darwin", func(string) ([]byte, error) {
		test.Fatal("Darwin must not read Linux proc/cgroup files")
		return nil, nil
	})
	after := before
	if before.GOOS != "darwin" || before.MetricsStatus != "unsupported" || before.MetricsReason == "" || before.RuntimeCPUs < 1 || before.GOMAXPROCS < 1 || len(before.Values) != 0 || len(before.Counters) != 0 || before.CgroupIdentity != "" || resourceSampleError("before", before) != nil {
		test.Fatal("unsupported receipt fabricated metrics or suppressed runtime CPU")
	}
	if delta, err := resourceDifference(before, after); err != nil || delta != nil {
		test.Fatal("unsupported metrics must not fail gates or fabricate a zero delta")
	}
	after.GOOS = "linux"
	if resourceSampleError("after", after) == nil {
		test.Fatal("Linux cannot bypass required sampling as unsupported")
	}
	if delta, err := resourceDifference(before, after); err == nil || delta != nil {
		test.Fatal("platform change was hidden")
	}
}

func TestLinuxResourceReadFailureRemainsFatal(test *testing.T) {
	validReader := resourceFixtureRead(test, nil, "")
	missing := filepath.Join(test.TempDir(), "missing-cpu-stat")
	readFailed := false
	failed := readResourcesForPlatform("linux", func(path string) ([]byte, error) {
		if filepath.Base(path) == "cpu.stat" {
			data, err := os.ReadFile(missing)
			readFailed = os.IsNotExist(err)
			return data, err
		}
		return validReader(path)
	})
	if !readFailed {
		test.Fatal("isolated real filesystem read did not produce ENOENT")
	}
	if failed.GOOS != "linux" || failed.MetricsStatus != "error" || resourceSampleError("before", failed) == nil {
		test.Fatal("actual Linux read failure was mislabeled unsupported")
	}
	valid := readResourcesForPlatform("linux", resourceFixtureRead(test, nil, ""))
	if delta, err := resourceDifference(valid, failed); delta != nil || err == nil {
		test.Fatal("Linux read failure produced a false valid delta")
	}
	test.Log("LINUX_READ_ERROR_FIXTURE errno=ENOENT metrics_status=error fatal=true delta=absent")
}
