# Runtime and persistence contracts

## Ownership and signatures
cmd/model-relay/main.go owns init, serve and reset-password operational entry points. internal/relay/server.go owns New(store *Store, config Config) *Server, ServeHTTP, reply and failure. internal/relay/store.go owns Init(directory, keyFile string) (string, error) and Store.DB; SQLite access uses database/sql and bound parameters, not an ORM. proxy.go owns outbound protocol handling; web assets are served by Go.

## Persistence and secrets
Follow store.go: separate key file, 0700 directory, 0600 key, exclusive installation lock, SQLite WAL, FULL synchronous and one connection. Admin passwords are salted Argon2 hashes; upstream credentials are encrypted; application keys use digests. Tests own temporary installations. Never edit user databases to pass tests; use README operational commands. No schema change belongs in a UI visibility feature.

## Input and failure boundary
server.go jsonInput requires application/json, a JSON object, bounded input, unique fields and DisallowUnknownFields. Example: failure(writer, http.StatusUnauthorized, "unauthorized", "") emits an error envelope with code/type/message and request_id; never include raw credentials or provider diagnostics. storageFailure logs only storage_write_failed and marks degraded state. UI api treats management 401 as signedOut; preserve that behavior.

## Health and version contract
/healthz supports GET and HEAD; HEAD has no body. Other methods are 405 with Allow: GET, HEAD. Healthy GET carries status and the raw Config.Version; degraded/closed database is unavailable. internal/relay/health_test.go covers handler and real HTTP method/state matrix; version_test.go covers raw version and real HTTP consistency. Do not bypass health-history checks or turn earlier receipts into current passes.

## Required verification
Makefile verify runs evidence, health-history, formatting, go mod verify, vet, Node UI tests, Go race/HTTP tests, build and real-service browser journeys. tests/browser/main.go starts isolated Go services and controlled upstreams; these are not real supplier evidence. Preserve failed attempts. Listening restrictions require host verification, not a weaker replacement gate. Avoid logging request/response bodies; tests/browser/journey.mjs audits logs and records for credential/conversation exclusion.

## Scenario: historical HTTP receipt output

### 1. Scope / Trigger
tests/health-history/main.go owns reconstruction and classification, not archival document updates. Default verification must not rewrite tracked historical evidence on either accepted or rejected results.

### 2. Signatures
`run() error` builds and classifies the isolated baseline; `assess(data []byte, exitCode int) error` accepts only the exact expected historical red. `main()` reports errors on stderr and exits 1. Public commands remain `make health-history` and `make verify`.

### 3. Contracts
The sanitized receipt (baseline/source hashes, command, child exit, events, PASS or REJECTED classification) goes only to stdout. No new env key or output-file flag. Caller-selected redirection stores new evidence; do not choose historical paths. `fmt.Print(cleanReceipt)` must have its write error checked. Keep all existing verify stages and fail-fast behavior.

### 4. Validation & Error Matrix
Exact baseline assertion exit 1 plus all required diagnostics/results → receipt and command success. Unexpected green, build/bind/race/timeout/invalid/incomplete events → rejected receipt and nonzero command. Closed stdout → `write historical HTTP receipt to stdout` stderr diagnostic and exit 1; never silently accept output loss. Pre-reconstruction errors remain nonzero. No outcome writes docs/workflow.

### 5. Good / Base / Bad Cases
Good: redirect `make verify` into a caller-selected new writable evidence path, retaining the exit code. Base: run with terminal output and no file side effect. Bad: persist to docs/workflow/health-methods-history.log, restore it by hand, or pipe through tee without exit propagation.

### 6. Tests Required
TestHistoryCommandOutput runs the actual command main in an isolated fixture with deterministic subprocess events; assert success receipt, rejected receipt and exit 1, closed-stdout diagnosis/exit 1, and unchanged historical-log/task sentinels in every case. The fixture stubs git/go only for output regression, not as evidence of real network reproduction. TestHistoricalAssessment retains exact classifier negative cases. Both enter make verify through existing Go race tests; host make verify must still run real baseline/current HTTP and Go browser journeys. Compare actual tracked historical files before/after host validation independently.

### 7. Wrong vs Correct
Wrong: `os.WriteFile("docs/workflow/health-methods-history.log", receipt, 0600)` followed by unchecked print. Correct: `if _, err := fmt.Print(cleanReceipt); err != nil { return fmt.Errorf("write historical HTTP receipt to stdout: %w", err) }`, then return assessment.
