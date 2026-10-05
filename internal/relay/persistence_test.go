package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHTTPRestartAndBackup(t *testing.T) {
	value := newFixture(t, normal)
	expect(t, value.call(t, ordinary), 200)
	var second struct {
		Key      string
		Metadata AppKey
	}
	decode(t, value.admin(t, "POST", "keys", map[string]any{"name": "second"}), &second)
	expect(t, value.admin(t, "PATCH", "keys/"+second.Metadata.ID, map[string]bool{"enabled": false}), 200)
	expect(t, value.call(t, strings.Replace(ordinary, "alpha", "missing", 1)), 404)
	oldCookie := value.cookie
	oldKey := value.key
	value.http.Close()
	if err := value.store.Close(); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "installation.backup")
	if err := Backup(value.directory, archive); err != nil {
		t.Fatal(err)
	}
	restored := filepath.Join(t.TempDir(), "restore")
	if err := Restore(archive, restored, value.keyFile); err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{value.directory, restored} {
		t.Run(filepath.Base(directory), func(t *testing.T) {
			store, err := Open(directory, value.keyFile)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			server := New(store, Config{Policy: value.server.Config.Policy})
			service := localServer(t, server)
			copy := *value
			copy.store = store
			copy.server = server
			copy.http = service
			copy.cookie = oldCookie
			copy.key = oldKey
			expect(t, copy.admin(t, "GET", "keys", nil), 401)
			copy.cookie = nil
			response := copy.admin(t, "POST", "login", map[string]any{"password": value.password})
			var login struct{ CSRF string }
			decode(t, response, &login)
			copy.csrf = login.CSRF
			copy.cookie = response.Cookies()[0]
			response = copy.admin(t, "GET", "requests", nil)
			var records struct{ Data []Record }
			decode(t, response, &records)
			if len(records.Data) < 2 || records.Data[len(records.Data)-1].State != "success" {
				t.Fatal("history not preserved")
			}
			expect(t, copy.call(t, ordinary), 200)
			copy.key = second.Key
			expect(t, copy.call(t, ordinary), 401)
			request, _ := http.NewRequest("GET", service.URL+"/v1/models", nil)
			request.Header.Set("Authorization", "Bearer "+oldKey)
			response, err = http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			var models struct{ Data []struct{ ID string } }
			decode(t, response, &models)
			if len(models.Data) != 1 || models.Data[0].ID != "alpha" {
				t.Fatal("restored mapping not visible")
			}
		})
	}
}

func TestInterruptedActualRequest(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		close(entered)
		select {
		case <-request.Context().Done():
		case <-release:
		}
		normal(writer, request)
	})
	completed := make(chan struct{})
	go func() {
		request, _ := http.NewRequest("POST", value.http.URL+"/v1/chat/completions", strings.NewReader(ordinary))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+value.key)
		response, _ := http.DefaultClient.Do(request)
		if response != nil {
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
		close(completed)
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("request not admitted")
	}
	if record := lastRecord(t, value); record.State != "started" {
		t.Fatal("durable admission absent")
	}
	value.store.Close()
	close(release)
	select {
	case <-completed:
	case <-time.After(5 * time.Second):
		t.Fatal("request did not end")
	}
	value.http.Close()
	reopened, err := Open(value.directory, value.keyFile)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	var state string
	var duration *int64
	if err = reopened.DB.QueryRow("SELECT state,duration_ms FROM requests ORDER BY rowid DESC LIMIT 1").Scan(&state, &duration); err != nil || state != "process_interrupted" || duration != nil {
		t.Fatal("interrupted admission falsely succeeded")
	}
}

func TestUsageSnapshots(t *testing.T) {
	for _, raw := range []string{`{}`, `{"usage":null}`, `{"usage":{"prompt_tokens":null,"completion_tokens":-1,"total_tokens":1.5}}`} {
		var value map[string]json.RawMessage
		json.Unmarshal([]byte(raw), &value)
		usage := usageFrom(value)
		if usage.Prompt != nil || usage.Completion != nil || usage.Total != nil {
			t.Fatal("invalid usage invented")
		}
	}
}
