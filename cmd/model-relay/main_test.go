package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

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
		Status  string
		Version string
	}
	if response.StatusCode != 200 || json.NewDecoder(response.Body).Decode(&health) != nil || health.Status != "ok" || health.Version != buildVersion {
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
	response = admin("GET", "upstream", nil)
	var initial struct {
		CredentialConfigured bool `json:"credential_configured"`
	}
	if json.NewDecoder(response.Body).Decode(&initial) != nil || initial.CredentialConfigured {
		t.Fatal("CLI empty installation contains invented configuration")
	}
	response.Body.Close()
	check(admin("POST", "models", map[string]any{"alias": "alpha", "model": "real", "enabled": true}), 200)
	keys := []string{}
	for _, name := range []string{"first", "second"} {
		response = admin("POST", "keys", map[string]string{"name": name})
		var issued struct {
			Key      string
			Metadata struct{ ID string }
		}
		if response.StatusCode != 201 || json.NewDecoder(response.Body).Decode(&issued) != nil {
			t.Fatal("CLI key issuance failed")
		}
		response.Body.Close()
		keys = append(keys, issued.Key)
		if name == "second" {
			check(admin("PATCH", "keys/"+issued.Metadata.ID, map[string]bool{"enabled": false}), 200)
		}
	}
	stop(command)
	command, address = start()
	check(admin("GET", "keys", nil), 401)
	cookie = nil
	login()
	response = admin("GET", "models", nil)
	var models []struct{ Alias string }
	if json.NewDecoder(response.Body).Decode(&models) != nil || len(models) != 1 || models[0].Alias != "alpha" {
		t.Fatal("CLI restart lost mapping")
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
