package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type clockRecoveryState struct {
	Directory  string
	KeyFile    string
	Tenant     string
	Old        string
	Current    string
	Expiring   string
	Invitation string
	Deadline   time.Time
	Actual     bool
}

func TestConsoleClockRecoveryInProcess(test *testing.T)  { clockRecoveryScenario(test, false) }
func TestConsoleClockRecoveryActualHTTP(test *testing.T) { clockRecoveryScenario(test, true) }

func clockRecoveryScenario(test *testing.T, actual bool) {
	fixture := singleResourceInstallation(test, actual)
	start := time.Unix(0, fixture.clock.Load())
	deadline := start.Add(time.Hour)
	tenantPath := "tenants/" + fixture.tenant
	if status, _ := fixture.admin.call("PUT", tenantPath+"/limits", ResourceLimits{Revision: 2, Calls: ptrInt64(2), RPM: ptrInt64(2), Hard: true}, ""); status != 200 {
		test.Fatal("finite window caps", status)
	}
	status, rotation := fixture.admin.call("POST", tenantPath+"/keys/"+fixture.keyID+"/rotations", map[string]any{"expected_revision": 1, "transition_hours": 1}, "clock-rotation")
	if status != 200 {
		test.Fatal("rotation", status)
	}
	expiry := deadline.Format(time.RFC3339Nano)
	status, expiring := fixture.admin.call("POST", tenantPath+"/keys", KeyInput{Name: "expiry test", Application: "controlled", Expires: &expiry, Enabled: true, Models: []string{"governed"}}, "clock-expiring")
	if status != 201 {
		test.Fatal("expiring key", status)
	}
	status, invitation := fixture.admin.call("POST", tenantPath+"/invitations", map[string]any{"account": "clock-test@example.com", "role": "developer", "hours": 1}, "")
	if status != 201 {
		test.Fatal("invitation", status)
	}
	if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 200 || receipt.err != nil {
		test.Fatal("initial accepted usage", receipt)
	}
	fixture.clock.Store(deadline.Add(-time.Nanosecond).UnixNano())
	if receipt := resourceGenerate(fixture, fixture.key); receipt.status != 200 || receipt.err != nil {
		test.Fatal("one nanosecond before deadline", receipt)
	}
	if receipt := resourceGenerate(fixture, expiring["key"].(string)); receipt.status != 200 || receipt.err != nil {
		test.Fatal("current key before deadline", receipt)
	}
	fixture.clock.Store(deadline.UnixNano())
	beforeHits := fixture.hits.Load()
	for _, secret := range []string{fixture.key, expiring["key"].(string)} {
		if receipt := resourceGenerate(fixture, secret); receipt.status != 401 || receipt.err != nil {
			test.Fatal("exact deadline auth", receipt.status)
		}
	}
	if status, _ := fixture.admin.call("POST", "invitations/accept", map[string]any{"token": invitation["acceptance_token"], "password": "controlled-clock-password"}, ""); status != 400 {
		test.Fatal("exact invitation deadline", status)
	}
	fixture.clock.Store(start.UnixNano())
	for _, secret := range []string{fixture.key, expiring["key"].(string)} {
		if receipt := resourceGenerate(fixture, secret); receipt.status != 401 || receipt.err != nil {
			test.Fatal("rollback auth", receipt.status)
		}
	}
	if fixture.hits.Load() != beforeHits {
		test.Fatal("expired key dispatched")
	}
	if status, _ := fixture.admin.call("POST", "invitations/accept", map[string]any{"token": invitation["acceptance_token"], "password": "controlled-clock-password"}, ""); status != 400 {
		test.Fatal("invitation revived", status)
	}
	status, report := fixture.admin.call("GET", tenantPath+"/limits", nil, "")
	if status != 200 || report["period"] != "2026-10" {
		test.Fatal("month regressed", status)
	}
	usage := resourceUsage(test, fixture, tenantPath)
	if usage.Calls != 2 || usage.RPM != 0 || usage.Fee != 14 || usage.Reserved != 0 || usage.Pending != 0 {
		test.Fatal("month/minute or ledger conservation", usage)
	}
	if receipt := resourceGenerate(fixture, rotation["key"].(string)); receipt.status != 429 || receipt.err != nil || fixture.hits.Load() != beforeHits {
		test.Fatal("rollback restored monthly quota", receipt.status)
	}
	var sequence int
	var name, filename string
	if err := fixture.store.DB.QueryRow("PRAGMA database_list").Scan(&sequence, &name, &filename); err != nil {
		test.Fatal(err)
	}
	directory := filepath.Dir(filename)
	state := clockRecoveryState{Directory: directory, KeyFile: filepath.Join(filepath.Dir(directory), "key"), Tenant: fixture.tenant, Old: fixture.key, Current: rotation["key"].(string), Expiring: expiring["key"].(string), Invitation: invitation["acceptance_token"].(string), Deadline: deadline, Actual: actual}
	stateFile := filepath.Join(test.TempDir(), "clock-state.json")
	encoded, err := json.Marshal(state)
	if err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(stateFile, encoded, 0600); err != nil {
		test.Fatal(err)
	}
	if err := fixture.store.Close(); err != nil {
		test.Fatal("close before restart", err)
	}
	binary, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	command := exec.Command(binary, "-test.run=^TestConsoleClockRecoveryWorker$", "-test.v")
	command.Env = append(os.Environ(), "MODEL_RELAY_CLOCK_STATE="+stateFile)
	output, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "CLOCK_RECOVERY_PASS") {
		test.Fatalf("clock recovery child: error=%v output=%s", err, output)
	}
	archive := filepath.Join(test.TempDir(), "backup.tar")
	if err := Backup(directory, archive); err != nil {
		test.Fatal("clock backup", err)
	}
	restored := filepath.Join(test.TempDir(), "restored")
	if err := Restore(archive, restored, state.KeyFile); err != nil {
		test.Fatal("clock restore", err)
	}
	state.Directory = restored
	encoded, err = json.Marshal(state)
	if err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(stateFile, encoded, 0600); err != nil {
		test.Fatal(err)
	}
	command = exec.Command(binary, "-test.run=^TestConsoleClockRecoveryWorker$", "-test.v")
	command.Env = append(os.Environ(), "MODEL_RELAY_CLOCK_STATE="+stateFile)
	output, err = command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "CLOCK_RECOVERY_PASS") {
		test.Fatalf("restored clock child: error=%v output=%s", err, output)
	}
}

func TestConsoleClockRecoveryWorker(test *testing.T) {
	stateFile := os.Getenv("MODEL_RELAY_CLOCK_STATE")
	if stateFile == "" {
		test.Skip("isolated recovery subprocess only")
	}
	encoded, err := os.ReadFile(stateFile)
	var state clockRecoveryState
	if err != nil || json.Unmarshal(encoded, &state) != nil {
		test.Fatal("read private controlled state")
	}
	store, err := Open(state.Directory, state.KeyFile)
	if err != nil {
		test.Fatal("normal restart", err)
	}
	defer store.Close()
	if !store.now().Equal(state.Deadline) {
		test.Fatal("durable UTC high water lost")
	}
	server := New(store, Config{Authority: "127.0.0.1:5544", Clock: func() time.Time { return state.Deadline.Add(-time.Hour) }})
	fixture := &governedFixture{test: test, store: store, server: server, origin: "http://127.0.0.1:5544"}
	if state.Actual {
		fixture.origin = localServer(test, server).URL
	}
	for _, secret := range []string{state.Old, state.Expiring} {
		if receipt := resourceGenerate(fixture, secret); receipt.status != 401 || receipt.err != nil {
			test.Fatal("expired credential revived after process restart", receipt.status)
		}
	}
	client := consoleClient{test: test, server: server}
	if status, _ := client.call("POST", "invitations/accept", map[string]any{"token": state.Invitation, "password": "controlled-clock-password"}, ""); status != 400 {
		test.Fatal("expired invitation revived after restart", status)
	}
	request := httptest.NewRequest("GET", fixture.origin+"/v1/models", nil)
	request.Header.Set("Authorization", "Bearer "+state.Current)
	var response *http.Response
	if state.Actual {
		request.RequestURI = ""
		response, err = http.DefaultClient.Do(request)
		if err != nil {
			test.Fatal("live generation models HTTP", err)
		}
	} else {
		writer := httptest.NewRecorder()
		server.ServeHTTP(writer, request)
		response = writer.Result()
	}
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if response.StatusCode != 200 {
		test.Fatal("current generation lost after restart", response.StatusCode)
	}
	if receipt := resourceGenerate(fixture, state.Current); receipt.status != 429 || receipt.err != nil {
		test.Fatal("restart restored monthly quota", receipt.status)
	}
	transaction, err := store.DB.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer transaction.Rollback()
	usage, err := usageFor(transaction, "tenant", state.Tenant, store.now().Format("2006-01"), store.now().Format("2006-01-02T15:04"))
	if err != nil || usage.Calls != 2 || usage.RPM != 0 || usage.Fee != 14 || usage.Reserved != 0 || usage.Pending != 0 {
		test.Fatal("restart ledger/window changed", usage, err)
	}
	test.Log("CLOCK_RECOVERY_PASS expired generation/current key/invite remain rejected; live generation conserved; UTC month/minute and fees conserved")
}
