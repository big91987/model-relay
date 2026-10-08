package relay

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConsoleRecoveryWorker(test *testing.T) {
	if os.Getenv("MODEL_RELAY_ISOLATED_RECOVERY_WORKER") != "1" {
		return
	}
	store, err := Open(os.Getenv("MODEL_RELAY_TEST_DIRECTORY"), os.Getenv("MODEL_RELAY_TEST_KEY"))
	if err != nil {
		test.Fatal("worker open failed")
	}
	defer store.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		test.Fatal("worker listener unavailable")
	}
	server := New(store, Config{Authority: listener.Addr().String(), Policy: Policy{TestAuthority: os.Getenv("MODEL_RELAY_TEST_AUTHORITY")}, Version: "isolated-recovery-test"})
	fmt.Println("RECOVERY_READY http://" + listener.Addr().String())
	if err := http.Serve(listener, server); err != nil {
		test.Fatal("worker serve failed")
	}
}

func TestConsoleRealProcessCrashRecovery(test *testing.T) {
	root := test.TempDir()
	directory, keyFile := filepath.Join(root, "data"), filepath.Join(root, "key")
	password, err := InitConsole(directory, keyFile)
	if err != nil {
		test.Fatal("process installation", err)
	}
	store, err := Open(directory, keyFile)
	if err != nil {
		test.Fatal(err)
	}
	entered, release := make(chan struct{}, 1), make(chan struct{})
	upstream := localServer(test, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		select {
		case entered <- struct{}{}:
		default:
		}
		select {
		case <-release:
		case <-request.Context().Done():
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled recovery"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`)
	}))
	defer close(release)
	server := New(store, Config{Authority: "127.0.0.1:5544", Policy: Policy{TestAuthority: strings.TrimPrefix(upstream.URL, "http://")}})
	admin := consoleClient{test: test, server: server}
	if status, _ := admin.call("POST", "login", map[string]any{"password": password}, ""); status != 200 {
		test.Fatal("parent configuration login", status)
	}
	status, tenant := admin.call("POST", "tenants", map[string]any{"name": "Recovery team"}, "")
	if status != 201 {
		test.Fatal("recovery tenant", status)
	}
	tenantID := tenant["id"].(string)
	steps := []struct {
		method, path string
		body         any
		operation    string
		status       int
	}{
		{"POST", "upstreams", map[string]any{"name": "Controlled recovery only", "url": upstream.URL + "/v1", "token": "controlled", "enabled": true}, "", 201},
		{"POST", "models", map[string]any{"alias": "recover", "enabled": true, "candidates": []map[string]any{{"upstream_id": 1, "remote_model": "controlled", "priority": 0, "weight": 1, "enabled": true}}}, "", 201},
		{"PUT", "models/recover/tenant-grants", map[string]any{"tenants": []string{tenantID}}, "", 200},
		{"PUT", "models/recover/bounds", map[string]any{"upstream_id": 1, "max_input": 100, "max_output": 100, "evidence": "test upstream bounded contract"}, "", 200},
		{"POST", "models/recover/prices", map[string]any{"prices": []map[string]any{{"upstream_id": 1, "cost_input_rate": 1000000, "cost_output_rate": 1000000, "tenant_input_rate": 1000000, "tenant_output_rate": 1000000}}}, "process-prices", 201},
		{"PUT", "tenants/" + tenantID + "/limits", ResourceLimits{Concurrency: ptrInt64(1), TPM: ptrInt64(1000), Budget: ptrInt64(1000), Hard: true}, "", 200},
	}
	for _, step := range steps {
		if status, _ := admin.call(step.method, step.path, step.body, step.operation); status != step.status {
			test.Fatal("process setup", step.path, status)
		}
	}
	status, key := admin.call("POST", "tenants/"+tenantID+"/keys", KeyInput{Name: "crash app", Application: "test-only", Models: []string{"recover"}, Enabled: true}, "process-key")
	if status != 201 {
		test.Fatal("process key", status)
	}
	secret := key["key"].(string)
	store.Close()
	binary, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	start := func() (*exec.Cmd, string) {
		command := exec.Command(binary, "-test.run=^TestConsoleRecoveryWorker$")
		command.Env = append(os.Environ(), "MODEL_RELAY_ISOLATED_RECOVERY_WORKER=1", "MODEL_RELAY_TEST_DIRECTORY="+directory, "MODEL_RELAY_TEST_KEY="+keyFile, "MODEL_RELAY_TEST_AUTHORITY="+strings.TrimPrefix(upstream.URL, "http://"))
		stdout, err := command.StdoutPipe()
		if err != nil {
			test.Fatal(err)
		}
		command.Stderr = io.Discard
		if command.Start() != nil {
			test.Fatal("worker process start")
		}
		test.Cleanup(func() { command.Process.Kill(); command.Wait() })
		ready := make(chan string, 1)
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				if strings.HasPrefix(scanner.Text(), "RECOVERY_READY ") {
					ready <- strings.TrimPrefix(scanner.Text(), "RECOVERY_READY ")
					return
				}
			}
			ready <- ""
		}()
		select {
		case address := <-ready:
			if address == "" {
				test.Fatal("recovery worker not ready")
			}
			return command, address
		case <-time.After(10 * time.Second):
			test.Fatal("recovery worker startup timeout")
		}
		return nil, ""
	}
	worker, address := start()
	childServer := &Server{Config: Config{Authority: strings.TrimPrefix(address, "http://")}}
	client := consoleClient{test: test, server: childServer}
	if status, _ := client.call("POST", "login", map[string]any{"password": password}, ""); status != 200 {
		test.Fatal("child actual login", status)
	}
	callEnded := make(chan struct{})
	go func() {
		defer close(callEnded)
		data, _ := json.Marshal(map[string]any{"model": "recover", "messages": []map[string]string{{"role": "user", "content": "controlled interrupted request"}}})
		request, _ := http.NewRequest("POST", address+"/v1/chat/completions", bytes.NewReader(data))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+secret)
		response, err := http.DefaultClient.Do(request)
		if err == nil {
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		test.Fatal("dispatch did not reach actual upstream")
	}
	status, records := client.call("GET", "workspace/requests?tenant="+tenantID, nil, "")
	if status != 200 || len(records["items"].([]any)) != 1 {
		test.Fatal("active persisted request", status)
	}
	id := records["items"].([]any)[0].(map[string]any)["id"].(string)
	if worker.Process.Kill() != nil {
		test.Fatal("crash failed")
	}
	worker.Wait()
	select {
	case <-callEnded:
	case <-time.After(5 * time.Second):
		test.Fatal("killed HTTP call remained open")
	}
	_, address = start()
	childServer.Config.Authority = strings.TrimPrefix(address, "http://")
	client.cookie = nil
	client.csrf = ""
	if status, _ := client.call("POST", "login", map[string]any{"password": password}, ""); status != 200 {
		test.Fatal("restart login", status)
	}
	status, detail := client.call("GET", "workspace/requests/"+id+"?tenant="+tenantID, nil, "")
	if status != 200 || detail["state"] != "process_interrupted" || detail["pending"] != float64(1) || detail["fee_reserved"] != "200" {
		test.Fatal("crash conserves unknown", status)
	}
	status, limits := client.call("GET", "tenants/"+tenantID+"/limits", nil, "")
	usage := limits["usage"].(map[string]any)
	if status != 200 || usage["calls"] != float64(1) || usage["concurrency"] != float64(0) || usage["reserved_micro_usd"] != "200" {
		test.Fatal("restart releases only local concurrency", status)
	}
	receipt := map[string]any{"expected_revision": 1, "input_tokens": 3, "output_tokens": 4, "reason": "controlled post-restart receipt", "evidence": "isolated-test-upstream"}
	path := "workspace/requests/" + id + "/reconciliations/1"
	if status, _ := client.call("POST", path, receipt, "post-crash-reconcile"); status != 200 {
		test.Fatal("restart reconcile", status)
	}
	if status, _ := client.call("POST", path, receipt, "post-crash-reconcile"); status != 200 {
		test.Fatal("restart reconcile replay", status)
	}
	status, detail = client.call("GET", "workspace/requests/"+id+"?tenant="+tenantID, nil, "")
	if status != 200 || detail["state"] != "process_interrupted" || detail["pending"] != float64(0) || detail["fee_reserved"] != "0" || detail["charge_entries"].([]any)[1].(map[string]any)["fee"] != "7" {
		test.Fatal("restart reconciliation preserves execution and money", status)
	}
}
