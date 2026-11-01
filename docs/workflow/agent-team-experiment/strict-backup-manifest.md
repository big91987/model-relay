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

## Unverified boundaries

Runtime execution of the focused regression, the broader relay/command suites, and `make verify` remains pending a complete Go 1.25 toolchain. Tests intentionally do not read real backups, real key files, or user data. Resource exhaustion beyond the existing 4096-byte manifest and 1 GiB archive-entry limits is not expanded by this change.
