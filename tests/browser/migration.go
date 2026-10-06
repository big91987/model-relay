package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

const migrationBaseline = "85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc"

func migrationCommand(directory string, args ...string) ([]byte, error) {
	command := exec.Command(args[0], args[1:]...)
	command.Dir = directory
	output, err := command.Output()
	if err != nil {
		return nil, errors.New("isolated migration command failed: " + filepath.Base(args[0]) + " " + args[1])
	}
	return output, nil
}

func snapshotRows(directory, query string) ([][]any, error) {
	db, err := sql.Open("sqlite3", "file:"+filepath.ToSlash(filepath.Join(directory, "relay.db"))+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := [][]any{}
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for index := range values {
			pointers[index] = &values[index]
		}
		if err = rows.Scan(pointers...); err != nil {
			return nil, err
		}
		result = append(result, values)
	}
	return result, rows.Err()
}

func isolatedService(binary, directory, key, authority string) (string, func() error, error) {
	command := exec.Command(binary, "--data-dir", directory, "--key-file", key, "--test-authority", authority)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return "", nil, err
	}
	command.Stderr = io.Discard
	if err = command.Start(); err != nil {
		return "", nil, errors.New("isolated service start failed")
	}
	ready := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "ISOLATED ") {
				ready <- strings.TrimPrefix(scanner.Text(), "ISOLATED ")
				return
			}
		}
		ready <- ""
	}()
	stop := func() error {
		command.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- command.Wait() }()
		select {
		case err := <-done:
			return err
		case <-time.After(10 * time.Second):
			command.Process.Kill()
			<-done
			return errors.New("isolated service did not stop")
		}
	}
	select {
	case address := <-ready:
		if address != "" {
			return address, stop, nil
		}
		stop()
		return "", nil, errors.New("isolated service not ready")
	case <-time.After(15 * time.Second):
		stop()
		return "", nil, errors.New("isolated service readiness timeout")
	}
}

func migrationJourney() error {
	root, err := os.MkdirTemp("", "model-relay-migration-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	source := filepath.Join(root, "source")
	if err = os.Mkdir(source, 0700); err != nil {
		return err
	}
	archive := filepath.Join(root, "source.tar")
	if _, err = migrationCommand(".", "git", "archive", "--format=tar", "--output="+archive, migrationBaseline); err != nil {
		return err
	}
	if _, err = migrationCommand(".", "tar", "-xf", archive, "-C", source); err != nil {
		return err
	}
	fixtureSource, err := os.ReadFile("tests/browser/legacy/main.go")
	if err != nil {
		return err
	}
	fixtureDirectory := filepath.Join(source, "cmd/migration-fixture")
	if err = os.Mkdir(fixtureDirectory, 0700); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(fixtureDirectory, "main.go"), fixtureSource, 0600); err != nil {
		return err
	}
	oldBinary := filepath.Join(root, "old")
	oldFixture := filepath.Join(root, "old-fixture")
	candidate := filepath.Join(root, "candidate")
	newFixture := filepath.Join(root, "candidate-fixture")
	for _, build := range []struct{ directory, output, target string }{{source, oldBinary, "./cmd/model-relay"}, {source, oldFixture, "./cmd/migration-fixture"}, {".", candidate, "./cmd/model-relay"}, {".", newFixture, "./tests/browser/legacy"}} {
		if _, err = migrationCommand(build.directory, "go", "build", "-o", build.output, build.target); err != nil {
			return err
		}
	}
	directory := filepath.Join(root, "installation")
	key := filepath.Join(root, "material.key")
	initialized, err := migrationCommand(".", oldBinary, "init", "--data-dir", directory, "--key-file", key)
	if err != nil {
		return err
	}
	password := strings.TrimSpace(strings.TrimPrefix(string(initialized), "Administrator password (shown once): "))
	if password == "" || strings.Contains(password, "\n") {
		return errors.New("old initialization response invalid")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return errors.New("migration controlled listener unavailable")
	}
	defer listener.Close()
	var calls atomic.Int64
	provider := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer isolated-legacy-upstream" {
			writer.WriteHeader(401)
			return
		}
		var input struct {
			Model    string
			Messages []struct{ Content string }
		}
		if json.NewDecoder(request.Body).Decode(&input) != nil || input.Model != "legacy-migration-a" {
			writer.WriteHeader(400)
			return
		}
		calls.Add(1)
		if len(input.Messages) > 0 && input.Messages[0].Content == "isolated legacy failure" {
			writer.WriteHeader(429)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"isolated migration"}}],"usage":{"prompt_tokens":0,"completion_tokens":2,"total_tokens":2}}`)
	})}
	go provider.Serve(listener)
	defer provider.Close()
	address, stop, err := isolatedService(oldFixture, directory, key, listener.Addr().String())
	if err != nil {
		return err
	}
	command := exec.Command("node", "tests/browser/legacy.mjs")
	command.Env = append(os.Environ(), "RELAY_LEGACY_URL="+address, "RELAY_LEGACY_PASSWORD="+password, "RELAY_LEGACY_UPSTREAM=http://"+listener.Addr().String()+"/v1")
	var private bytes.Buffer
	command.Stdout = &private
	command.Stderr = io.Discard
	err = command.Run()
	stopErr := stop()
	if err != nil || stopErr != nil {
		return errors.New("exact old UI seed failed")
	}
	var seed struct {
		ApplicationKey string `json:"application_key"`
		RequestID      string `json:"request_id"`
		FailureID      string `json:"failure_id"`
	}
	if json.Unmarshal(private.Bytes(), &seed) != nil || seed.ApplicationKey == "" || seed.RequestID == "" || seed.FailureID == "" {
		return errors.New("old UI seed identity missing")
	}
	private.Reset()
	queries := []string{"SELECT salt,hash FROM admin", "SELECT id,name,digest,hint,enabled,created FROM app_keys ORDER BY id", "SELECT rowid,id,time,key_id,alias,model,stream,revision,state,error_code,upstream_status,duration_ms,prompt_tokens,completion_tokens,total_tokens FROM requests ORDER BY rowid"}
	before := make([][][]any, len(queries))
	for index, query := range queries {
		before[index], err = snapshotRows(directory, query)
		if err != nil {
			return err
		}
	}
	models, err := snapshotRows(directory, "SELECT alias,model,enabled FROM models ORDER BY alias")
	if err != nil || len(models) != 2 {
		return errors.New("old UI did not create nonempty mappings")
	}
	if len(before[1]) != 2 || len(before[2]) == 0 {
		return errors.New("old UI did not create keys and real history")
	}
	rollback := filepath.Join(root, "old-rollback.tar")
	if _, err = migrationCommand(".", oldBinary, "backup", "--data-dir", directory, "--output", rollback); err != nil {
		return err
	}
	migrationBackup := filepath.Join(root, "migration.tar")
	upgraded, err := migrationCommand(".", candidate, "upgrade", "--data-dir", directory, "--key-file", key, "--output", migrationBackup)
	if err != nil {
		return err
	}
	var result struct {
		Result string
		Schema int
	}
	if json.Unmarshal(upgraded, &result) != nil || result.Result != "migrated" || result.Schema != 2 {
		return errors.New("candidate upgrade response invalid")
	}
	for index, query := range queries {
		after, err := snapshotRows(directory, query)
		if err != nil || !reflect.DeepEqual(before[index], after) {
			return errors.New("nonempty migration changed identity or historical rows")
		}
	}
	afterModels, err := snapshotRows(directory, "SELECT m.alias,c.remote_model,m.enabled FROM models m JOIN route_candidates c ON c.alias=m.alias ORDER BY m.alias")
	if err != nil || !reflect.DeepEqual(models, afterModels) {
		return errors.New("migration changed public mappings")
	}
	address, stop, err = isolatedService(newFixture, directory, key, listener.Addr().String())
	if err != nil {
		return err
	}
	response, err := http.Get(address + "/healthz")
	if err != nil {
		stop()
		return err
	}
	var health struct {
		Schema int `json:"storage_schema"`
	}
	json.NewDecoder(response.Body).Decode(&health)
	response.Body.Close()
	if health.Schema != 2 {
		stop()
		return errors.New("health did not report opened schema")
	}
	request, _ := http.NewRequest("POST", address+"/v1/chat/completions", strings.NewReader(`{"model":"migration-a","messages":[{"role":"user","content":"isolated after upgrade"}]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+seed.ApplicationKey)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		stop()
		return err
	}
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	status := response.StatusCode
	command = exec.Command("node", "tests/browser/migration-after.mjs")
	command.Env = append(os.Environ(), "RELAY_LEGACY_URL="+address, "RELAY_LEGACY_PASSWORD="+password)
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err = command.Run(); err != nil {
		stop()
		return errors.New("upgraded UI new-object verification failed")
	}
	if err = stop(); err != nil {
		return err
	}
	if status != 200 || calls.Load() != 3 {
		return errors.New("old credential did not survive real generation")
	}
	restored := filepath.Join(root, "restored")
	if _, err = migrationCommand(".", oldBinary, "restore", "--input", rollback, "--data-dir", restored, "--key-file", key); err != nil {
		return err
	}
	for index, query := range queries {
		after, err := snapshotRows(restored, query)
		if err != nil || !reflect.DeepEqual(before[index], after) {
			return errors.New("old binary restore changed pre-upgrade snapshot")
		}
	}
	keys, err := snapshotRows(restored, "SELECT name FROM app_keys ORDER BY name")
	if err != nil || len(keys) != 2 {
		return errors.New("rollback retained post-upgrade object")
	}
	address, stop, err = isolatedService(oldFixture, restored, key, listener.Addr().String())
	if err != nil {
		return err
	}
	request, _ = http.NewRequest("POST", address+"/v1/chat/completions", strings.NewReader(`{"model":"migration-a","messages":[{"role":"user","content":"isolated after rollback"}]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+seed.ApplicationKey)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		stop()
		return err
	}
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	status = response.StatusCode
	if err = stop(); err != nil {
		return err
	}
	if status != 200 || calls.Load() != 4 {
		return errors.New("old binary rollback was not runnable")
	}
	fmt.Printf("PASS exact legacy UI migration baseline=%s schema=1->2 aliases=2 keys=2 historical_requests=%d controlled_posts=%d; old-binary rollback retains pre-upgrade identities; not formal Actions or supplier validation\n", migrationBaseline, len(before[2]), calls.Load())
	return nil
}
