package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"model-relay/internal/relay"
)

func configureCLIInstallation(t *testing.T, admin func(string, string, any) *http.Response) {
	t.Helper()
	response := admin("GET", "upstreams", nil)
	var initial struct {
		Data []struct{ ID int64 }
		Next json.RawMessage `json:"next_cursor"`
	}
	err := json.NewDecoder(response.Body).Decode(&initial)
	response.Body.Close()
	if response.StatusCode != 200 || err != nil || initial.Data == nil || len(initial.Data) != 0 || string(initial.Next) != "null" {
		t.Fatal("CLI empty installation contains invented configuration or invalid envelope")
	}
	response = admin("POST", "upstreams", map[string]any{"name": "isolated CLI metadata only", "url": "https://8.8.8.8/v1", "token": "isolated-cli-sentinel", "enabled": true})
	var upstream struct {
		ID                   int64
		CredentialConfigured bool `json:"credential_configured"`
	}
	err = json.NewDecoder(response.Body).Decode(&upstream)
	response.Body.Close()
	if response.StatusCode != 201 || err != nil || upstream.ID < 1 || !upstream.CredentialConfigured {
		t.Fatal("CLI upstream creation contract failed")
	}
	response = admin("POST", "models", map[string]any{"alias": "alpha", "enabled": true, "candidates": []map[string]any{{"upstream_id": upstream.ID, "remote_model": "real", "enabled": true, "priority": 0, "weight": 1}}})
	response.Body.Close()
	if response.StatusCode != 201 {
		t.Fatalf("CLI model creation status %d, wanted 201", response.StatusCode)
	}
}

func TestCLIManagementConfigurationContract(t *testing.T) {
	root := t.TempDir()
	directory, keyFile := filepath.Join(root, "data"), filepath.Join(root, "material")
	password, err := relay.Init(directory, keyFile)
	if err != nil {
		t.Fatal("init failed")
	}
	store, err := relay.Open(directory, keyFile)
	if err != nil {
		t.Fatal("open failed")
	}
	defer store.Close()
	server := relay.New(store, relay.Config{Authority: "cli.test"})
	var cookie *http.Cookie
	csrf := ""
	admin := func(method, path string, body any) *http.Response {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal("invalid test input")
		}
		request := httptest.NewRequest(method, "http://cli.test/admin/api/"+path, bytes.NewReader(data))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "http://cli.test")
		request.Header.Set("X-CSRF-Token", csrf)
		request.Header.Set("Idempotency-Key", fmt.Sprintf("cli-%d", time.Now().UnixNano()))
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response.Result()
	}
	response := admin("POST", "login", map[string]string{"password": password})
	var login struct{ CSRF string }
	err = json.NewDecoder(response.Body).Decode(&login)
	response.Body.Close()
	if response.StatusCode != 200 || err != nil || len(response.Cookies()) != 1 {
		t.Fatal("login failed")
	}
	cookie, csrf = response.Cookies()[0], login.CSRF
	response = admin("GET", "upstream", nil)
	response.Body.Close()
	if response.StatusCode != 410 {
		t.Fatal("singleton endpoint did not report contract change")
	}
	response = admin("POST", "models", map[string]any{"alias": "alpha", "model": "real", "enabled": true})
	response.Body.Close()
	if response.StatusCode != 400 {
		t.Fatal("legacy model payload accepted")
	}
	configureCLIInstallation(t, admin)
}

func TestCLIInstallationCommands(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "model-relay")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("CLI build failed: %v %s", err, output)
	}
	directory := filepath.Join(root, "data")
	keyFile := filepath.Join(root, "material")
	execute := func(args ...string) ([]byte, error) {
		command := exec.Command(binary, args...)
		return command.CombinedOutput()
	}
	output, err := execute("init", "--data-dir", directory, "--key-file", keyFile)
	if err != nil || !bytes.HasPrefix(output, []byte("Administrator password (shown once): ")) {
		t.Fatal("CLI init failed")
	}
	password := strings.TrimSpace(strings.TrimPrefix(string(output), "Administrator password (shown once): "))
	if _, err = execute("init", "--data-dir", directory, "--key-file", keyFile); err == nil {
		t.Fatal("CLI init overwrote installation")
	}
	backup := filepath.Join(root, "backup")
	if _, err = execute("backup", "--data-dir", directory, "--output", backup); err != nil {
		t.Fatal("CLI backup failed")
	}
	restored := filepath.Join(root, "restored")
	if _, err = execute("restore", "--input", backup, "--data-dir", restored, "--key-file", keyFile); err != nil {
		t.Fatal("CLI restore failed")
	}
	if _, err = execute("serve", "--data-dir", restored, "--key-file", filepath.Join(root, "missing")); err == nil {
		t.Fatal("serve accepted missing material")
	}
	if _, err = execute("serve", "--data-dir", restored, "--key-file", keyFile, "--listen", "0.0.0.0:5544"); err == nil {
		t.Fatal("CLI exposed nonloopback service")
	}
	if entries, _ := os.ReadDir(restored); len(entries) == 0 {
		t.Fatal("failed operation lost data")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("CLI port occupancy verification unavailable: %v", err)
	}
	defer listener.Close()
	output, err = execute("serve", "--data-dir", restored, "--key-file", keyFile, "--listen", listener.Addr().String())
	if err == nil || !strings.Contains(string(output), "listen failed") {
		t.Fatal("occupied port not rejected")
	}
	start := func() (*exec.Cmd, string) {
		command := exec.Command(binary, "serve", "--data-dir", restored, "--key-file", keyFile, "--listen", "127.0.0.1:0")
		stdout, err := command.StdoutPipe()
		if err != nil {
			t.Fatal("CLI stdout unavailable")
		}
		command.Stderr = io.Discard
		if command.Start() != nil {
			t.Fatal("CLI serve process failed")
		}
		t.Cleanup(func() { command.Process.Kill(); command.Wait() })
		ready := make(chan string, 1)
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				if strings.HasPrefix(scanner.Text(), "Model Relay listening on ") {
					ready <- strings.TrimSuffix(strings.TrimPrefix(scanner.Text(), "Model Relay listening on "), "/admin/")
					return
				}
			}
			ready <- ""
		}()
		select {
		case address := <-ready:
			if address == "" {
				t.Fatal("CLI did not become ready")
			}
			return command, address
		case <-time.After(10 * time.Second):
			t.Fatal("CLI startup deadline exceeded")
		}
		return nil, ""
	}
	stop := func(command *exec.Cmd) {
		if command.Process.Signal(syscall.SIGTERM) != nil {
			t.Fatal("CLI SIGTERM failed")
		}
		finished := make(chan error, 1)
		go func() { finished <- command.Wait() }()
		select {
		case err := <-finished:
			if err != nil {
				t.Fatal("CLI graceful stop failed")
			}
		case <-time.After(12 * time.Second):
			t.Fatal("CLI stop deadline exceeded")
		}
	}
	command, address := start()
	var cookie *http.Cookie
	csrf := ""
	admin := func(method, path string, body any) *http.Response {
		data, _ := json.Marshal(body)
		request, _ := http.NewRequest(method, address+"/admin/api/"+path, bytes.NewReader(data))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", address)
		request.Header.Set("X-CSRF-Token", csrf)
		request.Header.Set("Idempotency-Key", fmt.Sprintf("cli-%d", time.Now().UnixNano()))
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal("CLI management HTTP failed")
		}
		return response
	}
	check := func(response *http.Response, status int) {
		defer response.Body.Close()
		io.Copy(io.Discard, response.Body)
		if response.StatusCode != status {
			t.Fatalf("CLI HTTP status %d, wanted %d", response.StatusCode, status)
		}
	}
	response, err := http.Get(address + "/healthz")
	if err != nil {
		t.Fatal("CLI health failed")
	}
	var health struct {
		Status        string
		Version       string
		StorageSchema int `json:"storage_schema"`
	}
	if response.StatusCode != 200 || json.NewDecoder(response.Body).Decode(&health) != nil || health.Status != "ok" || health.Version != buildVersion || health.StorageSchema != 3 {
		t.Fatal("CLI health did not report the running version")
	}
	response.Body.Close()
	login := func() {
		response := admin("POST", "login", map[string]string{"password": password})
		if response.StatusCode != 200 {
			t.Fatal("CLI administrator login failed")
		}
		var value struct{ CSRF string }
		if json.NewDecoder(response.Body).Decode(&value) != nil {
			t.Fatal("CLI login JSON failed")
		}
		response.Body.Close()
		csrf = value.CSRF
		cookie = response.Cookies()[0]
	}
	login()
	configureCLIInstallation(t, admin)
	response = admin("POST", "tenants", map[string]string{"name": "CLI persistence team"})
	var tenant struct{ ID string }
	if response.StatusCode != 201 || json.NewDecoder(response.Body).Decode(&tenant) != nil {
		t.Fatal("CLI tenant creation failed")
	}
	response.Body.Close()
	check(admin("PUT", "models/alpha/tenant-grants", map[string]any{"tenants": []string{tenant.ID}}), 200)
	keys := []string{}
	for _, name := range []string{"first", "second"} {
		response = admin("POST", "tenants/"+tenant.ID+"/keys", map[string]any{"name": name, "application": name, "enabled": true, "models": []string{"alpha"}})
		var issued struct {
			Key string
			ID  string
		}
		if response.StatusCode != 201 || json.NewDecoder(response.Body).Decode(&issued) != nil {
			t.Fatal("CLI key issuance failed")
		}
		response.Body.Close()
		keys = append(keys, issued.Key)
		if name == "second" {
			check(admin("PATCH", "tenants/"+tenant.ID+"/keys/"+issued.ID, map[string]any{"name": name, "application": name, "enabled": false, "models": []string{"alpha"}, "expected_revision": 1}), 200)
		}
	}
	reset := func(value string, extra ...string) ([]byte, error) {
		args := []string{"reset-password", "--data-dir", restored, "--key-file", keyFile}
		args = append(args, extra...)
		resetCommand := exec.Command(binary, args...)
		resetCommand.Stdin = strings.NewReader(value)
		return resetCommand.CombinedOutput()
	}
	if _, err := reset("admin\n", "--password-stdin"); err == nil {
		t.Fatal("password reset accepted a running installation")
	}
	stop(command)
	if _, err := reset("admin\n"); err == nil {
		t.Fatal("password reset accepted missing explicit stdin flag")
	}
	for _, invalid := range []string{"", "\n", "one\ntwo", strings.Repeat("x", 1025)} {
		if _, err := reset(invalid, "--password-stdin"); err == nil {
			t.Fatal("password reset accepted invalid input")
		}
	}
	if _, err := reset("admin\n", "--password-stdin", "--key-file", filepath.Join(root, "missing")); err == nil {
		t.Fatal("password reset accepted missing encryption key")
	}
	output, err = reset("admin\n", "--password-stdin")
	if err != nil {
		t.Fatalf("password reset failed: %v %s", err, output)
	}
	if bytes.Contains(output, []byte("admin")) {
		t.Fatal("password reset exposed replacement password")
	}
	command, address = start()
	check(admin("GET", "keys", nil), 401)
	cookie = nil
	check(admin("POST", "login", map[string]string{"password": password}), 401)
	password = "admin"
	login()
	response = admin("GET", "models", nil)
	var models struct {
		Data []struct {
			Alias      string
			Candidates []struct {
				UpstreamID  int64  `json:"upstream_id"`
				RemoteModel string `json:"remote_model"`
			}
		}
	}
	if response.StatusCode != 200 || json.NewDecoder(response.Body).Decode(&models) != nil || len(models.Data) != 1 || models.Data[0].Alias != "alpha" || len(models.Data[0].Candidates) != 1 || models.Data[0].Candidates[0].RemoteModel != "real" {
		t.Fatal("CLI restart lost mapping")
	}
	response.Body.Close()
	response = admin("GET", "upstreams", nil)
	var upstreams struct {
		Data []struct {
			ID                   int64
			Enabled              bool
			CredentialConfigured bool `json:"credential_configured"`
		}
	}
	if response.StatusCode != 200 || json.NewDecoder(response.Body).Decode(&upstreams) != nil || len(upstreams.Data) != 1 || upstreams.Data[0].ID != models.Data[0].Candidates[0].UpstreamID || !upstreams.Data[0].Enabled || !upstreams.Data[0].CredentialConfigured {
		t.Fatal("CLI restart lost upstream or candidate identity")
	}
	response.Body.Close()
	for index, key := range keys {
		request, _ := http.NewRequest("GET", address+"/v1/models", nil)
		request.Header.Set("Authorization", "Bearer "+key)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal("CLI restarted application HTTP failed")
		}
		status := 200
		if index == 1 {
			status = 401
		}
		check(response, status)
	}
	stop(command)
}
