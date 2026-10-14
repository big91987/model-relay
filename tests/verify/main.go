package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type receipt struct {
	StartedUTC         string            `json:"started_utc"`
	FinishedUTC        string            `json:"finished_utc"`
	Command            string            `json:"command"`
	Isolation          string            `json:"isolation"`
	NodeModulesBefore  bool              `json:"node_modules_before"`
	CleanStartMeasured bool              `json:"clean_start_measured"`
	NPMCacheBefore     bool              `json:"npm_cache_before"`
	NodeModulesAfter   bool              `json:"node_modules_after"`
	NPMCacheAfter      bool              `json:"npm_cache_after"`
	SourceUnchanged    bool              `json:"source_unchanged"`
	CandidateUnchanged bool              `json:"candidate_unchanged"`
	NodePath           string            `json:"NODE_PATH"`
	ManualPreinstall   string            `json:"manual_preinstall"`
	Before             json.RawMessage   `json:"source_before,omitempty"`
	After              json.RawMessage   `json:"source_after,omitempty"`
	Versions           map[string]string `json:"versions"`
	DependencyReceipt  string            `json:"dependency_receipt,omitempty"`
	GateExit           int               `json:"gate_exit"`
	GateBudgetMS       int64             `json:"gate_configured_budget_ms"`
	Gate               *commandReceipt   `json:"gate,omitempty"`
	ResourcesBefore    resourceSnapshot  `json:"resources_before"`
	ResourcesAfter     resourceSnapshot  `json:"resources_after"`
	ResourceDelta      *resourceDelta    `json:"resource_delta,omitempty"`
	Exit               int               `json:"exit_code"`
	LogSHA256          string            `json:"log_sha256"`
	LogBytes           int64             `json:"log_bytes"`
	Cleanup            string            `json:"cleanup"`
	OwnedGroupAbsent   *bool             `json:"owned_group_absent,omitempty"`
	AuxiliaryCommands  []commandReceipt  `json:"auxiliary_commands,omitempty"`
}

type commandReceipt struct {
	Stage                  string `json:"stage"`
	Exit                   int    `json:"exit"`
	Cancelled              bool   `json:"cancelled"`
	GroupAbsent            *bool  `json:"owned_group_absent,omitempty"`
	StartedUTC             string `json:"started_utc"`
	BudgetMilliseconds     int64  `json:"budget_ms"`
	ExecuteMilliseconds    int64  `json:"execute_ms"`
	GroupCheckMilliseconds int64  `json:"group_check_ms"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		root, _ := os.Getwd()
		fmt.Fprintln(os.Stderr, "isolated verification failed:", pathRedactor(root, "", os.TempDir(), os.Getenv("HOME")).Replace(err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context) (result error) {
	return runWithSampler(ctx, sampleResources)
}

func runWithSampler(ctx context.Context, sampler func() resourceSnapshot) (result error) {
	source, err := os.Getwd()
	if err != nil {
		return err
	}
	outputRoot := filepath.Join(source, "test-results", "verification")
	if err = os.MkdirAll(outputRoot, 0700); err != nil {
		return err
	}
	output, err := os.MkdirTemp(outputRoot, time.Now().UTC().Format("20060102T150405Z")+"-")
	if err != nil {
		return err
	}
	fmt.Println("verification_receipt=" + filepath.ToSlash(strings.TrimPrefix(output, source+string(os.PathSeparator))))
	state := receipt{StartedUTC: time.Now().UTC().Format(time.RFC3339Nano), Command: "make verify -> isolated make verify-gates", Isolation: "new private candidate; git history read-only", NodePath: "unset", ManualPreinstall: "none", Versions: map[string]string{}, GateExit: -1, Exit: 1, Cleanup: "not_started"}
	state.GateBudgetMS = verificationGateBudget.Milliseconds()
	logFile, err := os.OpenFile(filepath.Join(output, "verify.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	redaction := newLog(io.MultiWriter(logFile, os.Stdout), source, "", os.Getenv("HOME"))
	defer func() {
		state.ResourcesAfter = sampler()
		var deltaErr error
		state.ResourceDelta, deltaErr = resourceDifference(state.ResourcesBefore, state.ResourcesAfter)
		result = errors.Join(result, resourceSampleError("after", state.ResourcesAfter), deltaErr)
		resourceData, _ := json.Marshal(state.ResourcesAfter)
		_, resourceWriteErr := fmt.Fprintln(redaction, "RESOURCE_AFTER", string(resourceData))
		result = errors.Join(result, resourceWriteErr)
		if state.ResourceDelta != nil {
			deltaData, _ := json.Marshal(state.ResourceDelta)
			_, deltaWriteErr := fmt.Fprintln(redaction, "RESOURCE_DELTA", string(deltaData))
			result = errors.Join(result, deltaWriteErr)
		}
		if result != nil {
			fmt.Fprintln(redaction, "FAILED:", result)
		}
		if flushErr := redaction.flush(); flushErr != nil {
			result = errors.Join(result, flushErr)
		}
		if closeErr := logFile.Close(); closeErr != nil {
			result = errors.Join(result, closeErr)
		}
		data, readErr := os.ReadFile(filepath.Join(output, "verify.log"))
		result = errors.Join(result, readErr)
		state.LogSHA256 = fmt.Sprintf("%x", sha256.Sum256(data))
		state.LogBytes = int64(len(data))
		state.FinishedUTC = time.Now().UTC().Format(time.RFC3339Nano)
		if result == nil {
			state.Exit = 0
		}
		result = errors.Join(result, writeReceipt(output, state))
	}()
	if err = writeReceipt(output, state); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp("", "model-relay-verify-")
	if err != nil {
		return err
	}
	defer func() {
		if removeErr := os.RemoveAll(temporary); removeErr != nil {
			state.Cleanup = "candidate_remove_failed"
			result = errors.Join(result, removeErr)
		} else {
			state.Cleanup += "; candidate_removed"
		}
	}()
	redaction.replacer = pathRedactor(source, temporary, os.TempDir(), os.Getenv("HOME"))
	state.ResourcesBefore = sampler()
	resourceData, _ := json.Marshal(state.ResourcesBefore)
	if _, err = fmt.Fprintln(redaction, "RESOURCE_BEFORE", string(resourceData)); err != nil {
		return err
	}
	if err = resourceSampleError("before", state.ResourcesBefore); err != nil {
		return err
	}
	files, err := captureCommand(ctx, "source-inventory", exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"), redaction, &state)
	if err != nil {
		return fmt.Errorf("source inventory: %w", err)
	}
	if err = copyFiles(source, temporary, bytes.Split(files, []byte{0})); err != nil {
		return err
	}
	gitDir, err := captureCommand(ctx, "git-directory", exec.Command("git", "rev-parse", "--absolute-git-dir"), redaction, &state)
	if err != nil {
		return err
	}
	environment := candidateEnv(os.Environ(), temporary, strings.TrimSpace(string(gitDir)))
	if err = prepareNPMConfig(temporary); err != nil {
		return err
	}
	state.NodeModulesBefore = exists(filepath.Join(temporary, "node_modules"))
	state.NPMCacheBefore = exists(filepath.Join(temporary, ".cache", "npm"))
	state.CleanStartMeasured = true
	if state.NodeModulesBefore || state.NPMCacheBefore {
		return errors.New("candidate is not dependency-clean")
	}
	state.Before, err = sourceEvidence(ctx, temporary, environment, redaction, &state)
	if err != nil {
		return err
	}
	original, err := sourceEvidence(ctx, source, os.Environ(), redaction, &state)
	if err != nil {
		return err
	}
	if !bytes.Equal(original, state.Before) {
		return errors.New("isolated candidate source differs")
	}
	for _, tool := range []string{"go", "node", "npm"} {
		argument := "--version"
		if tool == "go" {
			argument = "version"
		}
		command := exec.Command(tool, argument)
		command.Dir, command.Env = temporary, environment
		version, versionErr := captureCommand(ctx, "prerequisite-"+tool, command, redaction, &state)
		if _, err = fmt.Fprintf(redaction, "PREREQUISITE tool=%s output=%s\n", tool, strings.TrimSpace(string(version))); err != nil {
			return fmt.Errorf("capture %s prerequisite: %w", tool, err)
		}
		if versionErr != nil {
			return fmt.Errorf("%s prerequisite: %w", tool, versionErr)
		}
		state.Versions[tool] = strings.TrimSpace(string(version))
	}
	if err = writeReceipt(output, state); err != nil {
		return err
	}
	encoded, _ := json.Marshal(state)
	fmt.Fprintln(redaction, "CLEAN_START", string(encoded))
	err = executeVerificationGate(ctx, exec.Command("make", "verify-gates"), temporary, environment, redaction, &state, verificationGateBudget)
	state.After, result = sourceEvidence(ctx, temporary, environment, redaction, &state)
	result = errors.Join(err, result)
	state.NodeModulesAfter = exists(filepath.Join(temporary, "node_modules"))
	state.NPMCacheAfter = exists(filepath.Join(temporary, ".cache", "npm"))
	state.CandidateUnchanged = bytes.Equal(state.Before, state.After)
	if !state.CandidateUnchanged {
		result = errors.Join(result, errors.New("candidate source changed during gates"))
	}
	originalAfter, originalErr := sourceEvidence(ctx, source, os.Environ(), redaction, &state)
	result = errors.Join(result, originalErr)
	state.SourceUnchanged = bytes.Equal(original, originalAfter)
	if !state.SourceUnchanged {
		result = errors.Join(result, errors.New("source changed during verification"))
	}
	if state.GateExit == 0 {
		probe := exec.Command("node", "-e", dependencyProbe)
		probe.Dir, probe.Env = temporary, environment
		data, probeErr := captureCommand(ctx, "dependency-probe", probe, redaction, &state)
		state.DependencyReceipt = strings.TrimSpace(string(data))
		result = errors.Join(result, probeErr)
	}
	if exists(filepath.Join(temporary, "test-results")) {
		result = errors.Join(result, copyTree(filepath.Join(temporary, "test-results"), filepath.Join(output, "browser")))
	}
	fmt.Fprintf(redaction, "CLEAN_END gate_exit=%d source_unchanged=%t NODE_PATH=unset\n", state.GateExit, bytes.Equal(state.Before, state.After))
	return result
}

const dependencyProbe = `const path=require('node:path'); const fs=require('node:fs'); const lock=JSON.parse(fs.readFileSync('package-lock.json')); for(const name of ['playwright','playwright-core']) { const resolved=require.resolve(name+'/package.json'); if(!resolved.startsWith(path.join(process.cwd(),'node_modules')+path.sep)) throw Error('nonlocal dependency'); const version=require(resolved).version; if(version!==lock.packages['node_modules/'+name].version) throw Error('lock mismatch'); console.log(name+'='+version+'; resolution=candidate/node_modules'); } const browser=require('node:child_process').spawnSync(require('playwright').chromium.executablePath(),['--version'],{encoding:'utf8'}); if(browser.status!==0) { process.stdout.write(browser.stdout||''); process.stderr.write(browser.stderr||''); console.error('CHROMIUM_PROBE exit='+browser.status+' signal='+browser.signal+' error='+(browser.error?.code||'none')); throw Error('Chromium version unavailable'); } console.log('chromium='+browser.stdout.trim());`

func exists(path string) bool { _, err := os.Lstat(path); return err == nil }

func writeReceipt(directory string, state receipt) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(directory, "state.json")
	if err = os.WriteFile(path+".tmp", append(data, '\n'), 0600); err != nil {
		return err
	}
	return os.Rename(path+".tmp", path)
}

func sourceEvidence(ctx context.Context, directory string, environment []string, output io.Writer, state *receipt) ([]byte, error) {
	command := exec.Command("go", "run", "./tests/evidence")
	command.Dir, command.Env = directory, environment
	data, err := captureCommand(ctx, "source-evidence", command, output, state)
	if err != nil {
		return nil, fmt.Errorf("source digest: %w", err)
	}
	if !json.Valid(data) {
		return nil, errors.New("invalid source digest")
	}
	return bytes.TrimSpace(data), nil
}

func candidateEnv(inherited []string, directory, gitDir string) []string {
	var environment []string
	for _, entry := range inherited {
		name, _, _ := strings.Cut(entry, "=")
		upper := strings.ToUpper(name)
		if strings.HasPrefix(upper, "NODE_") || strings.HasPrefix(upper, "NPM_CONFIG_") || strings.HasPrefix(upper, "GIT_") || name == "GOCACHE" || name == "GOTOOLCHAIN" || name == "CGO_ENABLED" || name == "MAKEFLAGS" || name == "MFLAGS" || name == "MAKELEVEL" {
			continue
		}
		environment = append(environment, entry)
	}
	return append(environment, "GIT_DIR="+gitDir, "GIT_WORK_TREE="+directory, "GOCACHE="+filepath.Join(directory, ".cache", "go-build"), "GOTOOLCHAIN=local", "CGO_ENABLED=1", "NPM_CONFIG_USERCONFIG="+filepath.Join(directory, ".cache", "verification-config", "user.npmrc"), "NPM_CONFIG_GLOBALCONFIG="+filepath.Join(directory, ".cache", "verification-config", "global.npmrc"))
}

func prepareNPMConfig(directory string) error {
	configRoot := filepath.Join(directory, ".cache", "verification-config")
	if err := os.MkdirAll(configRoot, 0700); err != nil {
		return err
	}
	for _, name := range []string{"user.npmrc", "global.npmrc"} {
		if err := os.WriteFile(filepath.Join(configRoot, name), nil, 0600); err != nil {
			return err
		}
	}
	return nil
}

func copyFiles(source, target string, files [][]byte) error {
	sourceRoot, err := openCopyRoot(source)
	if err != nil {
		return err
	}
	defer sourceRoot.Close()
	targetRoot, err := openCopyRoot(target)
	if err != nil {
		return err
	}
	defer targetRoot.Close()
	for _, raw := range files {
		name := string(raw)
		if name == "" {
			continue
		}
		if filepath.IsAbs(name) || filepath.Clean(name) != name || strings.HasPrefix(name, ".."+string(os.PathSeparator)) || name == ".." || name == "." {
			return errors.New("unsafe inventory path")
		}
		if err := copyFile(sourceRoot, targetRoot, name); err != nil {
			return err
		}
	}
	return nil
}

func openCopyRoot(path string) (*os.Root, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("copy root must be a real directory")
	}
	return os.OpenRoot(path)
}

func copyFile(source, target *os.Root, name string) error {
	parent := "."
	components := strings.Split(name, string(os.PathSeparator))
	for _, component := range components[:len(components)-1] {
		parent = filepath.Join(parent, component)
		info, err := source.Lstat(parent)
		if err != nil || !info.IsDir() {
			return errors.Join(errors.New("source parent must be a real directory"), err)
		}
		if err := target.Mkdir(parent, 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		info, err = target.Lstat(parent)
		if err != nil || !info.IsDir() {
			return errors.Join(errors.New("target parent must be a real directory"), err)
		}
	}
	info, err := source.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("inventory must contain regular files only")
	}
	data, err := source.ReadFile(name)
	if err != nil {
		return err
	}
	file, err := target.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return errors.Join(err, file.Close())
}

func copyTree(source, target string) error {
	if err := os.MkdirAll(target, 0700); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		return copyFiles(source, target, [][]byte{[]byte(relative)})
	})
}

func captureCommand(ctx context.Context, stage string, requested *exec.Cmd, output io.Writer, state *receipt) ([]byte, error) {
	bounded, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(bounded, requested.Path, requested.Args[1:]...)
	command.Dir, command.Env = requested.Dir, requested.Env
	var stdout bytes.Buffer
	command.Stdout = io.MultiWriter(output, &stdout)
	command.Stderr = output
	deadline, _ := bounded.Deadline()
	entry := commandReceipt{Stage: stage, StartedUTC: time.Now().UTC().Format(time.RFC3339Nano), BudgetMilliseconds: max(0, time.Until(deadline).Milliseconds())}
	if bounded.Err() != nil {
		entry.BudgetMilliseconds = 0
	}
	if _, err := fmt.Fprintf(output, "COMMAND_START stage=%s started_utc=%s budget_ms=%d\n", stage, entry.StartedUTC, entry.BudgetMilliseconds); err != nil {
		return nil, err
	}
	executeStarted := time.Now()
	code, err := execute(bounded, command, output)
	entry.ExecuteMilliseconds = time.Since(executeStarted).Milliseconds()
	entry.Exit, entry.Cancelled = code, bounded.Err() != nil
	if command.Process != nil {
		groupCheckStarted := time.Now()
		absent := ownedGroupAbsent(command.Process.Pid)
		entry.GroupCheckMilliseconds = time.Since(groupCheckStarted).Milliseconds()
		entry.GroupAbsent = &absent
		if !absent {
			err = errors.Join(err, errors.New("auxiliary owned process group remains after bounded cleanup"))
		}
	}
	state.AuxiliaryCommands = append(state.AuxiliaryCommands, entry)
	_, writeErr := fmt.Fprintf(output, "\nCOMMAND_END stage=%s exit=%d cancelled=%t execute_ms=%d group_check_ms=%d\n", stage, code, entry.Cancelled, entry.ExecuteMilliseconds, entry.GroupCheckMilliseconds)
	err = errors.Join(err, writeErr)
	if err != nil {
		err = fmt.Errorf("%s: %w", stage, err)
	}
	return stdout.Bytes(), err
}

func execute(ctx context.Context, command *exec.Cmd, output io.Writer) (int, error) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if command.Stdout == nil {
		command.Stdout = output
	}
	if command.Stderr == nil {
		command.Stderr = output
	}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGTERM) }
	command.WaitDelay = 5 * time.Second
	if err := command.Start(); err != nil {
		return -1, err
	}
	err := command.Wait()
	cleanupErr := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	if cleanupErr != nil && cleanupErr != syscall.ESRCH {
		err = errors.Join(err, cleanupErr)
	}
	if ctx.Err() != nil {
		err = errors.Join(err, ctx.Err())
	}
	return command.ProcessState.ExitCode(), err
}

func ownedGroupAbsent(processID int) bool {
	deadline := time.Now().Add(time.Second)
	for {
		if syscall.Kill(-processID, 0) == syscall.ESRCH {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(20 * time.Millisecond)
	}
}

type cleanLog struct {
	mu       sync.Mutex
	target   io.Writer
	pending  []byte
	replacer *strings.Replacer
}

func pathRedactor(source, temporary, tempRoot, home string) *strings.Replacer {
	var pairs []string
	for _, pair := range [][2]string{{source, "<source>"}, {temporary, "<isolated>"}, {home, "<home>"}, {"/private" + tempRoot, "<temp>"}, {tempRoot, "<temp>"}} {
		if pair[0] != "" {
			pairs = append(pairs, pair[0], pair[1])
		}
	}
	pairs = append(pairs, "\x00", "\\0")
	return strings.NewReplacer(pairs...)
}

func newLog(target io.Writer, source, temporary, home string) *cleanLog {
	return &cleanLog{target: target, replacer: pathRedactor(source, temporary, os.TempDir(), home)}
}

func (log *cleanLog) Write(data []byte) (int, error) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.pending = append(log.pending, data...)
	for {
		index := bytes.IndexByte(log.pending, '\n')
		if index < 0 {
			break
		}
		if _, err := io.WriteString(log.target, log.replacer.Replace(string(log.pending[:index+1]))); err != nil {
			return 0, err
		}
		log.pending = log.pending[index+1:]
	}
	return len(data), nil
}

func (log *cleanLog) flush() error {
	log.mu.Lock()
	defer log.mu.Unlock()
	_, err := io.WriteString(log.target, log.replacer.Replace(string(log.pending)))
	log.pending = nil
	return err
}
