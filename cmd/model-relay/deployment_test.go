package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDeploymentContractAndIdempotentUpgradeCLI(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "relay")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s", output)
	}
	workspace := filepath.Join(root, "work")
	os.Mkdir(workspace, 0700)
	command := exec.Command(binary, "deployment-contract")
	command.Dir = workspace
	output, err := command.Output()
	if err != nil {
		t.Fatal("contract failed")
	}
	var contract struct {
		Version int    `json:"contract_version"`
		Binary  string `json:"binary_version"`
		Storage struct {
			Init     int   `json:"init_schema"`
			Serve    []int `json:"serve_schemas"`
			Upgrade  []int `json:"upgrade_from"`
			Explicit bool  `json:"explicit_upgrade"`
			Backup   []int `json:"backup_schemas"`
			Restore  []int `json:"restore_schemas"`
		}
	}
	if json.Unmarshal(output, &contract) != nil || contract.Version != 1 || contract.Binary != "dev" || contract.Storage.Init != 3 || len(contract.Storage.Serve) != 1 || contract.Storage.Serve[0] != 3 || !contract.Storage.Explicit || len(contract.Storage.Upgrade) != 2 || contract.Storage.Upgrade[0] != 1 || contract.Storage.Upgrade[1] != 2 || len(contract.Storage.Backup) != 3 || len(contract.Storage.Restore) != 3 {
		t.Fatal("contract mismatch")
	}
	entries, _ := os.ReadDir(workspace)
	if len(entries) != 0 {
		t.Fatal("read-only contract created material")
	}
	directory := filepath.Join(root, "data")
	key := filepath.Join(root, "material.key")
	if _, err = exec.Command(binary, "init", "--data-dir", directory, "--key-file", key).Output(); err != nil {
		t.Fatal("init")
	}
	archive := filepath.Join(root, "should-not-exist.tar")
	for index := 0; index < 2; index++ {
		output, err = exec.Command(binary, "upgrade", "--data-dir", directory, "--key-file", key, "--output", archive).Output()
		var result struct {
			Result string
			Schema int
			From   int `json:"from_schema"`
		}
		if err != nil || json.Unmarshal(output, &result) != nil || result.Result != "already_current" || result.Schema != 3 || result.From != 3 {
			t.Fatal("idempotent current upgrade")
		}
	}
	if _, err = os.Stat(archive); !os.IsNotExist(err) {
		t.Fatal("already-current created backup")
	}
	if output, err = exec.Command(binary, "upgrade", "--data-dir", directory, "--key-file", key).Output(); err == nil || len(output) != 0 {
		t.Fatal("missing output accepted")
	}
}
