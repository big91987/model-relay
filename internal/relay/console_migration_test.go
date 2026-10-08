package relay

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"path/filepath"
	"testing"
)

func TestConsoleMigrationConservation(test *testing.T) {
	for _, version := range []int{1, 2} {
		test.Run(map[int]string{1: "schema1", 2: "schema2"}[version], func(test *testing.T) {
			var directory, keyFile, password, secret string
			if version == 1 {
				directory, keyFile = legacyInstallation(test, true)
				secret = "old-key1"
			} else {
				store, initialPassword, initialDirectory, initialKey := installation(test)
				directory, keyFile, password = initialDirectory, initialKey, initialPassword
				server := New(store, Config{Authority: "127.0.0.1:5544", Policy: Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}})
				client := consoleClient{test: test, server: server}
				if status, _ := client.call("POST", "login", map[string]any{"password": password}, ""); status != 200 {
					test.Fatal("schema2 login", status)
				}
				if status, _ := client.call("POST", "upstreams", map[string]any{"name": "schema2 source", "url": "https://public.example/v1", "token": "controlled-old-token", "enabled": true}, ""); status != 201 {
					test.Fatal("legacy source", status)
				}
				if status, _ := client.call("POST", "models", map[string]any{"alias": "alpha", "enabled": true, "candidates": []map[string]any{{"upstream_id": 1, "remote_model": "old", "enabled": true, "priority": 0, "weight": 1}}}, ""); status != 201 {
					test.Fatal("legacy model", status)
				}
				status, issued := client.call("POST", "keys", map[string]any{"name": "old app"}, "")
				if status != 201 {
					test.Fatal("legacy key", status)
				}
				secret = issued["key"].(string)
				server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) { return upstreamJSON(), nil })
				request := httptest.NewRequest("POST", "http://127.0.0.1:5544/v1/chat/completions", bytes.NewBufferString(`{"model":"alpha","messages":[{"role":"user","content":"controlled legacy call"}]}`))
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Authorization", "Bearer "+secret)
				writer := httptest.NewRecorder()
				server.ServeHTTP(writer, request)
				if writer.Code != 200 {
					test.Fatal("legacy generation", writer.Code)
				}
				if store.Close() != nil {
					test.Fatal("legacy close")
				}
			}
			before := filepath.Join(test.TempDir(), "before.tar")
			if err := Backup(directory, before); err != nil {
				test.Fatal("pre migration backup", err)
			}
			if _, err := UpgradeConsole(directory, keyFile, before); err == nil {
				test.Fatal("migration accepted existing backup destination")
			}
			inspect, err := openInstallation(directory, keyFile, true)
			if err != nil {
				test.Fatal("source after failed backup", err)
			}
			if inspect.Schema != version {
				test.Fatal("failed upgrade changed version")
			}
			inspect.Close()
			result, err := UpgradeConsole(directory, keyFile, filepath.Join(test.TempDir(), "upgrade.tar"))
			if err != nil || result.Schema != 3 || result.FromSchema != version {
				test.Fatal("console upgrade", result, err)
			}
			store, err := Open(directory, keyFile)
			if err != nil {
				test.Fatal("migrated open", err)
			}
			server := New(store, Config{Authority: "127.0.0.1:5544"})
			request := httptest.NewRequest("GET", "http://127.0.0.1:5544/v1/models", nil)
			request.Header.Set("Authorization", "Bearer "+secret)
			valid := false
			if version == 1 {
				_, valid = generationIdentity(store.DB, secret, store.now())
			} else {
				_, valid = server.appID(request)
			}
			if !valid {
				test.Fatal("legacy digest identity lost")
			}
			var originalCount, pending int
			if store.DB.QueryRow("SELECT count(*) FROM requests").Scan(&originalCount) != nil || store.DB.QueryRow("SELECT count(*) FROM legacy_pending WHERE resolved=0").Scan(&pending) != nil || originalCount != 1 || pending != 1 {
				test.Fatal("historical responsibility lost")
			}
			if version == 2 {
				client := consoleClient{test: test, server: server}
				if status, _ := client.call("POST", "login", map[string]any{"password": password}, ""); status != 200 {
					test.Fatal("migrated password", status)
				}
				var requestID string
				store.DB.QueryRow("SELECT id FROM requests").Scan(&requestID)
				if status, overview := client.call("GET", "workspace/overview", nil, ""); status != 200 || overview["legacy_pending_requests"] != float64(1) {
					test.Fatal("overview omitted unknown historical liability", status)
				}
				evidence := map[string]any{"expected_revision": 0, "reason": "controlled historical receipt", "evidence": "test-not-supplier", "attempts": []map[string]any{{"ordinal": 1, "upstream_id": 1, "historical_price_reference": "controlled-original-contract", "input_tokens": 1, "output_tokens": 2, "cost_input_rate": 1000000, "cost_output_rate": 2000000, "tenant_input_rate": 3000000, "tenant_output_rate": 4000000}}}
				if status, _ := client.call("POST", "workspace/requests/"+requestID+"/legacy-reconciliations", evidence, "historical"); status != 200 {
					test.Fatal("formal historical reconcile", status)
				}
				if status, _ := client.call("POST", "workspace/requests/"+requestID+"/legacy-reconciliations", evidence, "historical"); status != 200 {
					test.Fatal("historical replay", status)
				}
				status, detail := client.call("GET", "workspace/requests/"+requestID, nil, "")
				if status != 200 || detail["legacy_pending"] != float64(0) || detail["state"] != "success" || detail["charge_entries"].([]any)[0].(map[string]any)["fee"] != "11" {
					test.Fatal("historical facts and financial evidence", status)
				}
				if status, overview := client.call("GET", "workspace/overview", nil, ""); status != 200 || overview["legacy_pending_requests"] != float64(0) {
					test.Fatal("overview retained resolved historical liability", status)
				}
				server.Config.Policy = Policy{Resolver: fixedResolver{[]netip.Addr{netip.MustParseAddr("8.8.8.8")}}}
				server.client.Transport = roundTrip(func(request *http.Request) (*http.Response, error) { return upstreamJSON(), nil })
				request = httptest.NewRequest("POST", "http://127.0.0.1:5544/v1/chat/completions", bytes.NewBufferString(`{"model":"alpha","messages":[{"role":"user","content":"controlled new legacy call"}]}`))
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Authorization", "Bearer "+secret)
				writer := httptest.NewRecorder()
				server.ServeHTTP(writer, request)
				if writer.Code != 200 {
					test.Fatal("legacy soft unpriced compatibility", writer.Code)
				}
				newID := writer.Header().Get("X-Request-ID")
				evidence["expected_revision"] = 1
				if status, _ := client.call("POST", "workspace/requests/"+newID+"/legacy-reconciliations", evidence, "new-unpriced"); status != 200 {
					test.Fatal("unpriced admission evidence", status)
				}
				status, detail = client.call("GET", "workspace/requests/"+newID, nil, "")
				if status != 200 || detail["pending"] != float64(0) || detail["price_snapshots"].([]any)[0].(map[string]any)["price_id"] != "legacy_unpriced" || len(detail["historical_price_evidence"].([]any)) != 1 {
					test.Fatal("append historical price without overwriting snapshot", status)
				}
			}
			var snapshot []byte
			rows, err := store.DB.Query("SELECT id,key_id,state FROM requests ORDER BY id")
			if err != nil {
				test.Fatal(err)
			}
			values := []string{}
			for rows.Next() {
				var id, key, state string
				rows.Scan(&id, &key, &state)
				values = append(values, id+key+state)
			}
			rows.Close()
			snapshot, _ = json.Marshal(values)
			store.Close()
			archive := filepath.Join(test.TempDir(), "schema3.tar")
			if err := Backup(directory, archive); err != nil {
				test.Fatal("schema3 backup", err)
			}
			restored := filepath.Join(test.TempDir(), "restored")
			if err := Restore(archive, restored, keyFile); err != nil {
				test.Fatal("schema3 restore", err)
			}
			store, err = Open(restored, keyFile)
			if err != nil {
				test.Fatal("restored schema3", err)
			}
			defer store.Close()
			rows, err = store.DB.Query("SELECT id,key_id,state FROM requests ORDER BY id")
			if err != nil {
				test.Fatal(err)
			}
			values = []string{}
			for rows.Next() {
				var id, key, state string
				rows.Scan(&id, &key, &state)
				values = append(values, id+key+state)
			}
			rows.Close()
			restoredSnapshot, _ := json.Marshal(values)
			if !bytes.Equal(snapshot, restoredSnapshot) {
				test.Fatal("restore changed history")
			}
		})
	}
}
