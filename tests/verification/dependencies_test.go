package verification

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const lockedInstall = "npm ci --ignore-scripts --omit=optional --no-audit --no-fund --cache .cache/npm"

func TestCleanVerificationDependencyOrder(t *testing.T) {
	makefile, err := os.ReadFile("../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"frontend-test", "browser", "verify"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "Makefile"), makefile, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("make", "-n", target)
			command.Dir = root
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("make dry run failed: %v\n%s", err, output)
			}
			plan := string(output)
			if target == "verify" {
				if !strings.Contains(plan, "go run ./tests/verify") {
					t.Fatal("fixed entry must invoke isolated capture runner")
				}
				command = exec.Command("make", "-n", "verify-gates")
				command.Dir = root
				output, err = command.CombinedOutput()
				if err != nil {
					t.Fatalf("gate dry run failed: %v\n%s", err, output)
				}
				plan = string(output)
			}
			install := strings.Index(plan, lockedInstall)
			consumer := "node --test tests/browser/*.test.mjs"
			if target == "browser" {
				consumer = "go run ./tests/browser"
			}
			consume := strings.Index(plan, consumer)
			if install < 0 || consume < 0 || install > consume {
				t.Fatalf("locked dependencies must precede %s in clean %s plan:\n%s", consumer, target, plan)
			}
			if target == "verify" {
				if !strings.Contains(plan, "go test -race -count=1 -v ./... -parallel=4") {
					t.Error("isolated fixture concurrency must be bounded without bypassing race or uncached tests")
				}
				for _, gate := range []string{"go run ./tests/evidence", "go run ./tests/review-history", "go run ./tests/health-history", "gofmt -l cmd internal tests", "go mod verify", "go vet ./...", "go test -race -count=1 -v ./...", "go build -trimpath -o bin/model-relay ./cmd/model-relay", "go run ./tests/browser"} {
					if !strings.Contains(plan, gate) {
						t.Errorf("missing original gate: %s", gate)
					}
				}
			}
		})
	}
}

func TestDependencyFailureStopsConsumers(t *testing.T) {
	makefile, err := os.ReadFile("../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	makePath, err := exec.LookPath("make")
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"frontend-test", "browser"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			files := map[string]string{
				"Makefile": string(makefile),
				"npm":      "#!/bin/sh\nprintf 'locked-install-refused\\n'\nexit 23\n",
				"node":     "#!/bin/sh\nprintf 'consumer-ran\\n'\nexit 0\n",
				"go":       "#!/bin/sh\nprintf 'consumer-ran\\n'\nexit 0\n",
			}
			for name, content := range files {
				if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0700); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.Command(makePath, target)
			command.Dir = root
			command.Env = []string{"PATH=" + root, "HOME=" + root}
			output, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "locked-install-refused") || strings.Contains(string(output), "consumer-ran") {
				t.Fatalf("installation failure must stop consumer with nonzero exit: %v\n%s", err, output)
			}
		})
	}
}
