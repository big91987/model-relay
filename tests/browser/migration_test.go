package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"model-relay/internal/relay"
)

func TestMigrationUpgradeTargetContracts(test *testing.T) {
	root := test.TempDir()
	for _, scenario := range []struct {
		name   string
		target string
		schema int
		args   []string
	}{
		{"schema2_compatibility", "./tests/browser/schema2", 2, nil},
		{"production_cli", "./cmd/model-relay", 3, []string{"upgrade"}},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			binary := filepath.Join(root, scenario.name)
			build := exec.Command("go", "build", "-o", binary, scenario.target)
			build.Dir = "../.."
			if err := build.Run(); err != nil {
				test.Fatal("isolated migration binary build failed")
			}
			directory := filepath.Join(test.TempDir(), "data")
			key := filepath.Join(test.TempDir(), "key")
			if _, err := relay.Init(directory, key); err != nil {
				test.Fatal("isolated schema2 initialization failed")
			}
			archive := filepath.Join(test.TempDir(), "before.tar")
			args := append(append([]string{}, scenario.args...), "--data-dir", directory, "--key-file", key, "--output", archive)
			output, err := exec.Command(binary, args...).Output()
			if err != nil {
				test.Fatal("isolated migration command failed")
			}
			var result relay.UpgradeResult
			if json.Unmarshal(output, &result) != nil || result.Schema != scenario.schema || result.FromSchema != 2 {
				test.Fatal("migration target contract mismatch")
			}
			if scenario.schema == 3 && result.Result != "migrated" || scenario.schema == 2 && result.Result != "already_current" {
				test.Fatal("migration disposition mismatch")
			}
			store, err := relay.Open(directory, key)
			if err != nil {
				test.Fatal("migration result unavailable")
			}
			defer store.Close()
			if store.Schema != scenario.schema {
				test.Fatal("response schema differs from opened installation")
			}
		})
	}
}
