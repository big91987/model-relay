# Strict backup manifest validation

## Scope and behavior

Restore now parses `manifest.json` as a single, exact JSON object before opening the restored database. The object must contain `version` and `sha256` exactly once and no other fields. Field types remain an integer and a string respectively. After the closing object, only JSON whitespace is accepted.

The checksum string must be the canonical SHA-256 representation produced by `Backup`: exactly 64 lowercase hexadecimal characters. Empty, short, non-hexadecimal, and uppercase representations are rejected. Supported archive versions, tar layout, database validation, schema reconciliation, and the final atomic directory rename are unchanged.

## Regression input categories

The table-driven restore regression creates a real installation in a test temporary directory, uses the project `Backup` function for the baseline archive, and rewrites only the test archive's manifest for negative cases. It covers:

- a valid project-generated backup and restore round trip;
- duplicate `version` and duplicate `sha256` members;
- an unknown member;
- a second JSON value after the manifest object;
- string `version` and numeric `sha256` field-type errors;
- empty, short, uppercase, and non-hexadecimal digests;
- non-object roots and either required member being absent.

Every archive-level negative case asserts that the requested restore directory does not exist afterward, demonstrating that no partial installation reached the target.

## Execution evidence

On 2026-11-01, the focused command was attempted:

```text
go test ./internal/relay -run 'TestRestoreStrictManifest|TestDecodeManifestRequiresExactObject' -count=1
```

It did not reach package compilation because the managed Go 1.25.0 toolchain installation lacked standard-library packages (`archive/tar` and `net/mail`). The default Go build cache was also read-only. Retrying with an isolated cache and `GOTOOLCHAIN=local` found Go 1.21.6, which cannot build this module because `go.mod` requires Go 1.25.0. `GOCACHE=/tmp/model-relay-go-cache make verify` was also attempted and exited 2 at `go run ./tests/verify` for that same Go 1.21.6 versus 1.25.0 requirement. Consequently, related command tests could not produce meaningful additional execution evidence in this environment.

`gofmt` completed successfully for both changed Go files.

### Third-round verification

The repaired execution environment exposed Go 1.25.6, Node 22.14.0, and npm 10.9.2 on 2026-11-01. The project entry point was run without changing its gates:

```text
make verify
```

It exited 2. The generated verification receipt was `test-results/verification/20261101T033438Z-476732183`. Prerequisite and source-evidence stages succeeded, but the isolated `verify-gates` run stopped in the existing `tests/review-history` command with `review historical HTTP rejected: baseline unavailable: exit status 128`; therefore the full gate did not reach evidence capable of accepting or rejecting this restore change.

The default Go build cache remained read-only for a direct test invocation. Re-running with a task-owned cache in `/tmp` succeeded:

```text
GOCACHE=/tmp/strict-backup-manifest-go-cache go test ./internal/relay -run 'TestRestoreStrictManifest|TestDecodeManifestRequiresExactObject|TestInstallationRecovery|TestHTTPRestartAndBackup' -count=1
ok model-relay/internal/relay 1.781s

GOCACHE=/tmp/strict-backup-manifest-go-cache go test ./cmd/model-relay -count=1
ok model-relay/cmd/model-relay 4.764s
```

### Fourth-round verification

After the maintenance source reported the historical tree restored, the unchanged project entry point was run again on 2026-11-01 with Go 1.25.6, Node 22.14.0, and npm 10.9.2:

```text
make verify
```

It again exited 2. The receipt was `test-results/verification/20261101T034452Z-3281180308`. Source inventory, source evidence, and all tool prerequisites passed, but the isolated gate again stopped at `go run ./tests/review-history` with `review historical HTTP rejected: baseline unavailable: exit status 128`. The receipt reports `source_unchanged=true`; no later required gate ran.

Using the controller-requested worktree-local Go cache, both task-relevant suites passed:

```text
GOCACHE="$PWD/.cache/go-build" go test ./internal/relay -run 'TestRestoreStrictManifest|TestDecodeManifestRequiresExactObject|TestInstallationRecovery|TestHTTPRestartAndBackup' -count=1
ok model-relay/internal/relay 1.747s

GOCACHE="$PWD/.cache/go-build" go test ./cmd/model-relay -count=1
ok model-relay/cmd/model-relay 4.420s
```

### Fifth-round verification

After a same-sandbox Git archive probe succeeded, the unchanged project entry point was run again on 2026-11-01:

```text
make verify
```

It exited 2 with receipt `test-results/verification/20261101T035009Z-1676340719`. This time `tests/review-history` successfully reconstructed baseline `66f1328ada75ea90e5638fbf2bd58663ac86347e` and observed the expected historical failures. The next required stage, `make health-history`, failed in `go run ./tests/health-history` with `historical HTTP reproduction rejected: baseline archive unavailable: exit status 128`. The receipt reports `source_unchanged=true`; stages after health-history did not run. Per the controller instruction, already-green task-specific Go tests were not repeated in this round.

### Sixth-round verification

After all three test-referenced historical baselines and their partial-clone objects were prepared, the unchanged project entry point was run again on 2026-11-01:

```text
make verify
```

It completed successfully with exit 0. The verification receipt was `test-results/verification/20261101T035409Z-1805577619`, and its final clean-start record reported `gate_exit=0`, `source_unchanged=true`, and a successful dependency probe. The required stages included successful historical review and health reconstructions, formatting, locked-module verification, `go vet`, all 286 frontend tests, the full race-enabled Go suite (including `TestRestoreStrictManifest` and `TestDecodeManifestRequiresExactObject`), build, a 2,480-sample latency experiment with zero failures, and all actual-Go-service browser phases. Browser providers remained explicitly controlled test upstreams rather than supplier validation.

## Unverified boundaries

The focused restore regression, existing backup/restore integration selections, and the command package passed in the third and fourth rounds, and the complete project verification passed in the sixth round. Tests intentionally do not read real backups, real key files, or user data. Resource exhaustion beyond the existing 4096-byte manifest and 1 GiB archive-entry limits is not expanded by this change.
