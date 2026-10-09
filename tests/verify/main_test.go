package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCandidateCopyAndEnvironment(test *testing.T) {
	source, target := test.TempDir(), test.TempDir()
	if err := os.WriteFile(filepath.Join(source, "package-lock.json"), []byte("locked"), 0600); err != nil {
		test.Fatal(err)
	}
	if err := copyFiles(source, target, [][]byte{[]byte("package-lock.json")}); err != nil {
		test.Fatal(err)
	}
	if exists(filepath.Join(target, "node_modules")) || exists(filepath.Join(target, ".cache/npm")) {
		test.Fatal("not clean")
	}
	for _, bad := range []string{"../escape", "/absolute", "sub/../escape"} {
		if copyFiles(source, target, [][]byte{[]byte(bad)}) == nil {
			test.Fatalf("accepted %s", bad)
		}
	}
	if err := os.Symlink(filepath.Join(source, "package-lock.json"), filepath.Join(source, "link")); err != nil {
		test.Fatal(err)
	}
	if copyFiles(source, target, [][]byte{[]byte("link")}) == nil {
		test.Fatal("accepted symlink")
	}
	environment := strings.Join(candidateEnv([]string{"PATH=/tools", "NODE_PATH=external", "NODE_OPTIONS=inject", "npm_config_cache=external", "NPM_CONFIG_USERCONFIG=external", "GIT_WORK_TREE=old", "MAKEFLAGS=-n", "GOCACHE=old"}, target, "/history"), "\n")
	for _, forbidden := range []string{"external", "inject", "=old", "MAKEFLAGS="} {
		if strings.Contains(environment, forbidden) {
			test.Fatal(environment)
		}
	}
	for _, required := range []string{"PATH=/tools", "GIT_DIR=/history", "GIT_WORK_TREE=" + target, "GOCACHE=" + filepath.Join(target, ".cache/go-build"), "NPM_CONFIG_USERCONFIG=" + filepath.Join(target, ".cache/verification-config/user.npmrc"), "NPM_CONFIG_GLOBALCONFIG=" + filepath.Join(target, ".cache/verification-config/global.npmrc")} {
		if !strings.Contains(environment, required) {
			test.Fatal(environment)
		}
	}
}

func TestCopyRejectsParentEscape(test *testing.T) {
	source, target, outside := test.TempDir(), test.TempDir(), test.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "input.txt"), []byte("outside-private-input"), 0600); err != nil {
		test.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(source, "internal")); err != nil {
		test.Fatal(err)
	}
	if err := copyFiles(source, target, [][]byte{[]byte("internal/input.txt")}); err == nil {
		test.Fatal("copied outside input through tracked parent symlink")
	}
	if exists(filepath.Join(target, "internal/input.txt")) {
		test.Fatal("outside input reached candidate")
	}
}

func TestCopyRootAndTargetBoundaries(test *testing.T) {
	for _, scenario := range []string{"normal_nested", "source_root", "target_root", "source_parent_internal", "target_parent", "target_file"} {
		test.Run(scenario, func(test *testing.T) {
			source, target, outside := test.TempDir(), test.TempDir(), test.TempDir()
			if err := os.Mkdir(filepath.Join(source, "nested"), 0700); err != nil {
				test.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(source, "nested/input.txt"), []byte("safe"), 0600); err != nil {
				test.Fatal(err)
			}
			link := func(destination, path string) {
				test.Helper()
				if err := os.Symlink(destination, path); err != nil {
					test.Fatal(err)
				}
			}
			switch scenario {
			case "source_root":
				alias := filepath.Join(test.TempDir(), "alias")
				link(source, alias)
				source = alias
			case "target_root":
				alias := filepath.Join(test.TempDir(), "alias")
				link(target, alias)
				target = alias
			case "source_parent_internal":
				if err := os.Rename(filepath.Join(source, "nested"), filepath.Join(source, "real")); err != nil {
					test.Fatal(err)
				}
				link("real", filepath.Join(source, "nested"))
			case "target_parent":
				link(outside, filepath.Join(target, "nested"))
			case "target_file":
				if err := os.Mkdir(filepath.Join(target, "nested"), 0700); err != nil {
					test.Fatal(err)
				}
				link(filepath.Join(outside, "output.txt"), filepath.Join(target, "nested/input.txt"))
			}
			err := copyFiles(source, target, [][]byte{[]byte("nested/input.txt")})
			if (err == nil) != (scenario == "normal_nested") {
				test.Fatalf("copy boundary: %v", err)
			}
			if exists(filepath.Join(outside, "output.txt")) || exists(filepath.Join(outside, "input.txt")) {
				test.Fatal("write escaped target")
			}
			if scenario == "normal_nested" {
				data, err := os.ReadFile(filepath.Join(target, "nested/input.txt"))
				if err != nil || string(data) != "safe" {
					test.Fatalf("normal copy: %s %v", data, err)
				}
			}
		})
	}
}

func TestInventoryCancellationReceipt(test *testing.T) {
	root := test.TempDir()
	tools := filepath.Join(root, "tools")
	if err := os.Mkdir(tools, 0700); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tools, "git"), []byte("#!/bin/sh\nsleep 0.6\nexit 27\n"), 0700); err != nil {
		test.Fatal(err)
	}
	test.Chdir(root)
	test.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := run(ctx)
	if err == nil || time.Since(started) > 500*time.Millisecond {
		test.Fatalf("inventory did not observe cancellation: err=%v elapsed=%s", err, time.Since(started))
	}
	matches, _ := filepath.Glob("test-results/verification/*/state.json")
	if len(matches) != 1 {
		test.Fatal("missing cancelled receipt")
	}
	data, readErr := os.ReadFile(matches[0])
	var state receipt
	if readErr != nil || json.Unmarshal(data, &state) != nil || state.FinishedUTC == "" || state.LogBytes == 0 || !strings.Contains(state.Cleanup, "candidate_removed") {
		test.Fatalf("incomplete cancellation receipt: %s %v", data, readErr)
	}
}

func TestRealNPMIsolatedConfig(test *testing.T) {
	directory := test.TempDir()
	if err := prepareNPMConfig(directory); err != nil {
		test.Fatal(err)
	}
	environment := candidateEnv(os.Environ(), directory, "/unused-history")
	for _, name := range []string{"user", "global"} {
		path := filepath.Join(directory, ".cache", "verification-config", name+".npmrc")
		data, err := os.ReadFile(path)
		if err != nil || len(data) != 0 {
			test.Fatalf("config is not empty: %v", err)
		}
		command := exec.Command("npm", "config", "get", name+"config")
		command.Dir, command.Env = directory, environment
		output, err := command.CombinedOutput()
		if err != nil || strings.TrimSpace(string(output)) != path {
			test.Fatalf("npm %sconfig: %v\n%s", name, err, output)
		}
	}
	command := exec.Command("npm", "--version")
	command.Dir, command.Env = directory, environment
	output, err := command.CombinedOutput()
	if err != nil || len(bytes.TrimSpace(output)) == 0 {
		test.Fatalf("real npm prerequisite: %v\n%s", err, output)
	}
	if exists(filepath.Join(directory, "node_modules")) || exists(filepath.Join(directory, ".cache", "npm")) {
		test.Fatal("prerequisite probe installed dependencies")
	}
}

func TestDependencyProbeChromiumFailureOutput(test *testing.T) {
	root := test.TempDir()
	files := map[string]string{
		"package-lock.json":                         `{"packages":{"node_modules/playwright":{"version":"test-only"},"node_modules/playwright-core":{"version":"test-only"}}}`,
		"node_modules/playwright/package.json":      `{"version":"test-only","main":"index.js"}`,
		"node_modules/playwright-core/package.json": `{"version":"test-only"}`,
		"node_modules/playwright/index.js":          "exports.chromium={executablePath:()=>require('node:path').join(process.cwd(),'test-chromium')};",
		"test-chromium":                             "#!/bin/sh\nprintf 'chromium-stdout-diagnostic\\n'; printf 'chromium-stderr-diagnostic\\n' >&2; exit 37\n",
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
	var captured bytes.Buffer
	log := newLog(&captured, root, "", os.Getenv("HOME"))
	command := exec.Command("node", "-e", dependencyProbe)
	command.Dir, command.Env = root, candidateEnv(os.Environ(), root, "/unused-history")
	var state receipt
	stdout, err := captureCommand(context.Background(), "dependency-probe", command, log, &state)
	if err == nil || !bytes.Contains(stdout, []byte("chromium-stdout-diagnostic")) {
		test.Fatalf("probe failure lost stdout: %v %s", err, stdout)
	}
	if err := log.flush(); err != nil {
		test.Fatal(err)
	}
	for _, marker := range []string{"chromium-stdout-diagnostic", "chromium-stderr-diagnostic", "CHROMIUM_PROBE exit=37", "COMMAND_END stage=dependency-probe exit=1"} {
		if !strings.Contains(captured.String(), marker) {
			test.Fatal(marker)
		}
	}
	if strings.Contains(captured.String(), root) {
		test.Fatal("probe diagnostic leaked local path")
	}
}

func TestVerifySignalHelper(test *testing.T) {
	if os.Getenv("VERIFY_SIGNAL_HELPER") == "1" {
		main()
	}
}

func TestRunnerSIGTERMPersistsCleanup(test *testing.T) {
	root := test.TempDir()
	tools := filepath.Join(root, "tools")
	if err := os.Mkdir(tools, 0700); err != nil {
		test.Fatal(err)
	}
	ready := filepath.Join(root, "ready")
	script := "#!/bin/sh\nsleep 30 & child=$!; printf '%s %s\\n' \"$PPID\" \"$child\" > \"$CANCEL_READY\"; wait\n"
	if err := os.WriteFile(filepath.Join(tools, "git"), []byte(script), 0700); err != nil {
		test.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestVerifySignalHelper$")
	command.Dir = root
	command.Env = append(os.Environ(), "VERIFY_SIGNAL_HELPER=1", "CANCEL_READY="+ready, "PATH="+tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	var output bytes.Buffer
	done := make(chan error, 1)
	go func() {
		_, err := execute(ctx, command, &output)
		done <- err
	}()
	for !exists(ready) && ctx.Err() == nil {
		time.Sleep(time.Millisecond)
	}
	if ctx.Err() != nil {
		test.Fatal("signal runner did not start auxiliary child")
	}
	childData, err := os.ReadFile(ready)
	var runnerPID, child int
	if err != nil {
		test.Fatal(err)
	}
	if _, err := fmt.Sscanf(string(childData), "%d %d", &runnerPID, &child); err != nil {
		test.Fatal(err)
	}
	runner, err := os.FindProcess(runnerPID)
	if err != nil {
		test.Fatal(err)
	}
	defer runner.Release()
	if err := runner.Signal(syscall.SIGTERM); err != nil {
		test.Fatal(err)
	}
	if err := <-done; err == nil || ctx.Err() != nil {
		test.Fatalf("signal runner did not fail promptly: %v", err)
	}
	matches, _ := filepath.Glob(filepath.Join(root, "test-results/verification/*/state.json"))
	if len(matches) != 1 {
		test.Fatal("signal receipt missing")
	}
	data, err := os.ReadFile(matches[0])
	var state receipt
	if err != nil || json.Unmarshal(data, &state) != nil || state.FinishedUTC == "" || state.Exit == 0 || state.LogBytes == 0 || !strings.Contains(state.Cleanup, "candidate_removed") {
		test.Fatalf("signal state incomplete: %s %v", data, err)
	}
	if len(state.AuxiliaryCommands) != 1 || !state.AuxiliaryCommands[0].Cancelled || state.AuxiliaryCommands[0].GroupAbsent == nil || !*state.AuxiliaryCommands[0].GroupAbsent {
		test.Fatalf("signal auxiliary group not closed: %+v", state)
	}
	if syscall.Kill(child, 0) != syscall.ESRCH {
		test.Fatal("signal runner's auxiliary child remains")
	}
}

func TestCaptureChunkBoundariesAndFailure(test *testing.T) {
	var captured bytes.Buffer
	log := newLog(&captured, "/private/source", "/private/candidate", "/home/operator")
	for _, chunk := range []string{"/private/sou", "rce start\n/home/oper", "ator tail"} {
		if _, err := log.Write([]byte(chunk)); err != nil {
			test.Fatal(err)
		}
	}
	if err := log.flush(); err != nil {
		test.Fatal(err)
	}
	if captured.String() != "<source> start\n<home> tail" {
		test.Fatal(captured.String())
	}
	command := exec.CommandContext(context.Background(), "/bin/sh", "-c", "printf 'stdout\\n'; printf 'stderr\\n' >&2; exit 23")
	code, err := execute(context.Background(), command, log)
	if err == nil || code != 23 {
		test.Fatalf("code=%d err=%v", code, err)
	}
	if err := log.flush(); err != nil {
		test.Fatal(err)
	}
	for _, text := range []string{"stdout", "stderr"} {
		if !strings.Contains(captured.String(), text) {
			test.Fatal(captured.String())
		}
	}
}

func TestOwnedGroupCancellation(test *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	var output bytes.Buffer
	command := exec.CommandContext(ctx, "/bin/sh", "-c", "sleep 30 & wait")
	started := time.Now()
	_, err := execute(ctx, command, &output)
	if err == nil || time.Since(started) > 8*time.Second {
		test.Fatalf("cancellation err=%v elapsed=%s", err, time.Since(started))
	}
	if !ownedGroupAbsent(command.Process.Pid) {
		test.Fatal("owned cancellation group still exists")
	}
}

func TestCaptureCommandTiming(test *testing.T) {
	for _, scenario := range []string{"success", "nonzero", "cancelled"} {
		test.Run(scenario, func(test *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			script := "printf 'timing-stdout\\n'; printf 'timing-stderr\\n' >&2"
			if scenario == "nonzero" {
				script += "; exit 23"
			} else if scenario == "cancelled" {
				cancel()
			}
			var output bytes.Buffer
			log := newLog(&output, "", "", "")
			var state receipt
			started := time.Now()
			stdout, err := captureCommand(ctx, "timing-fixture", exec.Command("/bin/sh", "-c", script), log, &state)
			finished := time.Now()
			if flushErr := log.flush(); flushErr != nil {
				test.Fatal(flushErr)
			}
			if (err != nil) != (scenario != "success") || len(state.AuxiliaryCommands) != 1 {
				test.Fatalf("unexpected timing outcome: err=%v entries=%d", err, len(state.AuxiliaryCommands))
			}
			entry := state.AuxiliaryCommands[0]
			stamp, stampErr := time.Parse(time.RFC3339Nano, entry.StartedUTC)
			if stampErr != nil || stamp.Before(started) || stamp.After(finished) || entry.BudgetMilliseconds < 0 || entry.BudgetMilliseconds > (2*time.Minute).Milliseconds() || entry.ExecuteMilliseconds < 0 || entry.GroupCheckMilliseconds < 0 || entry.ExecuteMilliseconds+entry.GroupCheckMilliseconds > finished.Sub(started).Milliseconds() {
				test.Fatalf("invalid timing receipt: %+v", entry)
			}
			if entry.Stage != "timing-fixture" || entry.Cancelled != (scenario == "cancelled") {
				test.Fatalf("lost boundary: %+v", entry)
			}
			if scenario == "cancelled" {
				if entry.BudgetMilliseconds != 0 || entry.Exit != -1 || len(stdout) != 0 {
					test.Fatalf("cancelled command ran: %+v", entry)
				}
			} else if string(stdout) != "timing-stdout\n" || !strings.Contains(output.String(), "timing-stderr") || entry.GroupAbsent == nil || !*entry.GroupAbsent || (scenario == "success" && entry.Exit != 0) || (scenario == "nonzero" && entry.Exit != 23) {
				test.Fatalf("capture changed semantics: %+v", entry)
			}
			for _, marker := range []string{"COMMAND_START stage=timing-fixture started_utc=", "budget_ms=", "COMMAND_END stage=timing-fixture exit=", "execute_ms=", "group_check_ms="} {
				if !strings.Contains(output.String(), marker) {
					test.Fatal("missing timing marker: " + marker)
				}
			}
		})
	}
}

func TestFixedRunnerReceipts(test *testing.T) {
	for _, scenario := range []string{"green_fixture", "failed_fixture", "prerequisite_failed_fixture", "evidence_failed_fixture", "probe_failed_fixture", "copy_failed_fixture", "cancel_inventory", "cancel_git_directory", "cancel_evidence", "cancel_prerequisite", "cancel_probe"} {
		test.Run(scenario, func(test *testing.T) {
			failure := scenario != "green_fixture"
			root := test.TempDir()
			tools := filepath.Join(root, "tools")
			if err := os.MkdirAll(tools, 0700); err != nil {
				test.Fatal(err)
			}
			files := map[string]string{
				"input.txt":  "fixture-source",
				"tools/git":  "#!/bin/sh\ncase \"$1\" in ls-files) printf 'input.txt\\0';; rev-parse) printf '/fixture/history\\n';; *) exit 19;; esac\n",
				"tools/go":   "#!/bin/sh\ncase \"$1\" in version) echo go-fixture;; run) printf 'evidence-stderr-success\\n' >&2; echo '{\"head\":\"fixture\",\"working_tree_sha256\":\"fixture\"}';; *) exit 19;; esac\n",
				"tools/npm":  "#!/bin/sh\necho npm-fixture\n",
				"tools/node": "#!/bin/sh\nif [ \"$1\" = --version ]; then echo node-fixture; else test -d node_modules && echo locked-local-fixture; fi\n",
				"tools/make": "#!/bin/sh\n[ \"$1\" = verify-gates ] || exit 19\ntest ! -e node_modules || exit 20\ntest ! -e .cache/npm || exit 21\ntest -z \"${NODE_PATH+x}\" || exit 22\nmkdir node_modules\nprintf 'fixture gates stdout\\n'; printf 'fixture gates stderr\\n' >&2\n",
			}
			if scenario == "failed_fixture" {
				files["tools/make"] += "exit 23\n"
			}
			if scenario == "prerequisite_failed_fixture" {
				files["tools/npm"] = "#!/bin/sh\nprintf 'prerequisite-stderr-diagnostic\\n' >&2\nexit 24\n"
			}
			if scenario == "evidence_failed_fixture" {
				files["tools/go"] = "#!/bin/sh\nprintf 'evidence-stdout-diagnostic\\n'; printf 'evidence-stderr-diagnostic\\n' >&2; exit 25\n"
			}
			if scenario == "probe_failed_fixture" {
				files["tools/node"] = "#!/bin/sh\nif [ \"$1\" = --version ]; then echo node-fixture; else printf 'probe-stdout-diagnostic\\n'; printf 'probe-stderr-diagnostic\\n' >&2; exit 26; fi\n"
			}
			if scenario == "copy_failed_fixture" {
				outside := test.TempDir()
				if err := os.WriteFile(filepath.Join(outside, "input.txt"), []byte("private-outside-input"), 0600); err != nil {
					test.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Join(root, "internal")); err != nil {
					test.Fatal(err)
				}
				files["tools/git"] = "#!/bin/sh\nprintf 'input.txt\\0internal/input.txt\\0'\n"
			}
			cancelStage := map[string]string{"cancel_inventory": "source-inventory", "cancel_git_directory": "git-directory", "cancel_evidence": "source-evidence", "cancel_prerequisite": "prerequisite-npm", "cancel_probe": "dependency-probe"}[scenario]
			waitScript := "printf 'auxiliary-ready\\n'; sleep 30 & child=$!; printf '%s\\n' \"$child\" > \"$CANCEL_READY\"; wait\n"
			switch scenario {
			case "cancel_inventory":
				files["tools/git"] = "#!/bin/sh\n" + waitScript
			case "cancel_git_directory":
				files["tools/git"] = "#!/bin/sh\nif [ \"$1\" = ls-files ]; then printf 'input.txt\\0'; else\n" + waitScript + "fi\n"
			case "cancel_evidence":
				files["tools/go"] = "#!/bin/sh\n" + waitScript
			case "cancel_prerequisite":
				files["tools/npm"] = "#!/bin/sh\n" + waitScript
			case "cancel_probe":
				files["tools/node"] = "#!/bin/sh\nif [ \"$1\" = --version ]; then echo node-fixture; else\n" + waitScript + "fi\n"
			}
			for name, content := range files {
				if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0700); err != nil {
					test.Fatal(err)
				}
			}
			test.Chdir(root)
			test.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
			test.Setenv("NODE_PATH", "must-not-survive")
			if err := os.Mkdir(filepath.Join(root, "node_modules"), 0700); err != nil {
				test.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			ready := filepath.Join(root, "cancel-ready")
			test.Setenv("CANCEL_READY", ready)
			if cancelStage != "" {
				go func() {
					for !exists(ready) && ctx.Err() == nil {
						time.Sleep(time.Millisecond)
					}
					cancel()
				}()
			}
			err := run(ctx)
			runErr := err
			matches, err := filepath.Glob("test-results/verification/*/state.json")
			if err != nil || len(matches) != 1 {
				test.Fatalf("receipts=%v err=%v", matches, err)
			}
			data, err := os.ReadFile(matches[0])
			if err != nil {
				test.Fatal(err)
			}
			var state receipt
			if err := json.Unmarshal(data, &state); err != nil {
				test.Fatal(err)
			}
			if (runErr != nil) != failure {
				for _, entry := range state.AuxiliaryCommands {
					test.Logf("FIXTURE_TIMING stage=%s started_utc=%s budget_ms=%d execute_ms=%d group_check_ms=%d exit=%d cancelled=%t", entry.Stage, entry.StartedUTC, entry.BudgetMilliseconds, entry.ExecuteMilliseconds, entry.GroupCheckMilliseconds, entry.Exit, entry.Cancelled)
				}
				test.Fatalf("err=%v failure=%t", runErr, failure)
			}
			earlyFailure := scenario == "copy_failed_fixture" || scenario == "cancel_inventory" || scenario == "cancel_git_directory"
			if state.CleanStartMeasured == earlyFailure || state.NodeModulesBefore || state.NPMCacheBefore || state.NodePath != "unset" || !strings.Contains(state.Cleanup, "candidate_removed") || state.FinishedUTC == "" {
				test.Fatalf("bad state: %+v", state)
			}
			if scenario == "failed_fixture" && (state.GateExit != 23 || state.Exit == 0) {
				test.Fatalf("lost failure: %+v", state)
			}
			log, err := os.ReadFile(filepath.Join(filepath.Dir(matches[0]), "verify.log"))
			if err != nil {
				test.Fatal(err)
			}
			if state.LogSHA256 != fmtHash(log) || state.LogBytes != int64(len(log)) {
				test.Fatal("log digest mismatch")
			}
			markers := []string{"CLEAN_START", "fixture gates stdout", "fixture gates stderr", "CLEAN_END"}
			if scenario == "prerequisite_failed_fixture" {
				markers = []string{"prerequisite-stderr-diagnostic", "COMMAND_END stage=prerequisite-npm exit=24", "FAILED: npm prerequisite:"}
				if state.GateExit != -1 || state.Exit == 0 || bytes.Contains(log, []byte("fixture gates stdout")) {
					test.Fatalf("prerequisite failure did not stop gates: %+v", state)
				}
			} else if scenario == "evidence_failed_fixture" {
				markers = []string{"evidence-stdout-diagnostic", "evidence-stderr-diagnostic", "exit=25"}
			} else if scenario == "probe_failed_fixture" {
				markers = []string{"probe-stdout-diagnostic", "probe-stderr-diagnostic", "exit=26"}
			} else if scenario == "copy_failed_fixture" {
				markers = []string{"source parent must be a real directory", "FAILED:"}
				if bytes.Contains(log, []byte("private-outside-input")) || state.GateExit != -1 {
					test.Fatal("escape was not fail-fast")
				}
			} else if cancelStage != "" {
				markers = []string{"auxiliary-ready", "COMMAND_END stage=" + cancelStage, "cancelled=true"}
				found := false
				for _, entry := range state.AuxiliaryCommands {
					if entry.Stage == cancelStage && entry.Cancelled && entry.GroupAbsent != nil && *entry.GroupAbsent {
						found = true
					}
				}
				if !found || state.Exit == 0 {
					test.Fatalf("missing cancelled group evidence: %+v", state)
				}
				data, err := os.ReadFile(ready)
				var child int
				if err != nil {
					test.Fatal(err)
				}
				if _, err := fmt.Sscanf(string(data), "%d", &child); err != nil {
					test.Fatal("invalid auxiliary child receipt")
				}
				if probeErr := syscall.Kill(child, 0); probeErr != syscall.ESRCH {
					test.Fatalf("actual auxiliary child remains: signal_probe=%v", probeErr)
				}
			} else if !bytes.Equal(state.Before, state.After) {
				test.Fatal("candidate source changed")
			}
			for _, marker := range markers {
				if !bytes.Contains(log, []byte(marker)) {
					test.Fatal(marker)
				}
			}
			if bytes.Contains(log, []byte(root)) {
				test.Fatal("local path leaked")
			}
			if !exists(filepath.Join(root, "node_modules")) {
				test.Fatal("mutated original dependencies")
			}
		})
	}
}

func fmtHash(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
