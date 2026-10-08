package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func main() {
	paths := []string{".gitignore", "Makefile", "README.md", "go.mod", "go.sum", "package.json", "package-lock.json", "docs/workflow/task.md", "docs/workflow/requirements.md", "docs/workflow/design.md", "docs/product-brief.md", "docs/roadmap.md", "docs/releases/v0.1.0.md", "docs/local-deployment.md"}
	for _, name := range []string{"product-definition.md", "console-architecture.md", "console-interaction.md", "routing-architecture.md", "routing-interaction.md", "routing-validation.md", "adr-001-storage-upgrade.md", "deployment-upgrade-contract.md", "architecture.md", "implementation-contract.md", "migration-and-operations.md", "verification-design.md", "acceptance-matrix.md"} {
		matches, err := filepath.Glob(filepath.Join("docs", "workflow", "runs", "*", name))
		if err != nil {
			fmt.Fprintln(os.Stderr, "contract inventory failed")
			os.Exit(1)
		}
		paths = append(paths, matches...)
	}
	for _, root := range []string{"cmd", "internal", "tests", ".trellis/spec"} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "source inventory failed")
			os.Exit(1)
		}
	}
	sort.Strings(paths)
	hash := sha256.New()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "source digest failed")
			os.Exit(1)
		}
		fmt.Fprintf(hash, "%s\x00%d\x00", filepath.ToSlash(path), len(data))
		hash.Write(data)
	}
	head, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "HEAD discovery failed")
		os.Exit(1)
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"go": runtime.Version(), "head": strings.TrimSpace(string(head)), "working_tree_sha256": hex.EncodeToString(hash.Sum(nil)), "files": len(paths), "scope": "product + tests + build locks + project contracts + historical/current acceptance/design + Trellis specs; excludes delivery receipts to avoid self-reference"})
}
