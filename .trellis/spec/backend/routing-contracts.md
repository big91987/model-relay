# Multi-upstream executable contracts

## QA remediation: connection checks (2026-10-06)

### Scope and signatures
`checkUpstream(writer,request,id)` is a single GET-list observation, not generation or routing-health recovery. POST `/admin/api/upstreams/{id}/check` keeps `expected_execution_revision` and the existing persisted `CheckResult`; no cancellation endpoint or schema migration.

### Contracts and error matrix
The outbound context is `context.WithTimeout(request.Context(),10*time.Second)` and covers headers plus all body reading. A 2xx observation succeeds only for media type `application/json` (parameters accepted), a bounded complete JSON object with `object:"list"`, non-null array `data`, and each element having a non-whitespace string `id`. Empty arrays are legitimate. Invalid type/object/id/JSON/oversize/body read, HTTP errors, network errors and deadline are `failed`; 401/403 use `upstream_rejected`, invalid list uses `upstream_protocol_error`, deadline uses `upstream_timeout`. Parent cancellation is `canceled/client_canceled`, never success. Interrupted historical checks may still be `unknown`; do not rewrite history to a known failure.

### Good, base, bad
Good: `{"object":"list","data":[{"id":"remote"}]}` with JSON content type. Base: same object and `data:[]`. Bad: `{"data":[]}`, numeric/null/empty IDs or HTML that happens to contain JSON. Save and check success never constitute generation success or erase cooling.

### Required tests and wrong versus correct
Preserve independent `qa_connection_contract_test.go` expectations. `connection_lifecycle_test.go` verifies deduplication, cancel propagation, cleanup/reentry and real HTTP header/body 10-second budget separately. `configuration_test.go` retains late-execution scoping. `routing_contract_regression_test.go` covers atomic numerical/reference boundaries, all-cooling zero dispatch, toggle-preserved cooldown, old sequence/execution/candidate isolation and restart unknown health. Tool EOF/idle/cancel stop assertions belong in `routing_http_test.go`, including zero secondary visits after committed tool delta. Wrong: check only non-null data or label a known 503 unknown. Correct: strict list validation followed by the persisted safe terminal state. Synthetic transports are layer evidence only; actual HTTP/UI runs require host `make verify`.

## 1. Scope / Trigger
Route-2 changes storage, operational CLI, authoritative management configuration, dispatch and diagnostics. Read with runtime-contracts.md and console-contracts.md; this supersedes baseline single-upstream/no-switch examples. Source ownership: upgrade.go/store.go/backup.go, configuration.go, routing.go/routed_proxy.go and console.go. This source-backed contract is not acceptance evidence.

## 2. Signatures
- `Upgrade(directory, keyFile, output string) (UpgradeResult, error)` holds the stopped-installation lock. `upgrade --data-dir --key-file --output` prints one JSON with `binary_version,result,from_schema,schema`. Read-only `deployment-contract`: contract_version=1, init_schema=2, serve_schemas=[2], upgrade_from=[1], explicit_upgrade=true, backup_schemas/restore_schemas=[1,2]. `/healthz` adds actual `storage_schema` to status/version.
- Management: GET/POST `/admin/api/upstreams`, GET/PATCH `/admin/api/upstreams/{id}`, POST `/{id}/check`; GET/POST `/admin/api/models`, GET/PATCH `/admin/api/models/{alias}`. List envelopes: `data,next_cursor`; default limit 50, range 1–100. Singular `/upstream` is 410. Session, Origin and CSRF controls remain authoritative.
- Upstream writes: `name,url,token,enabled,expected_revision` (revision required for PATCH). Reads expose id/name/url/enabled/revision/execution_revision/credential_configured/check, never token. Check input is `expected_execution_revision`; result carries state/execution_revision/applicable/started_at/finished_at/duration_ms/error_code/upstream_status.
- Model create: `alias,enabled,candidates`; edit: `expected_revision,enabled,candidates`. Candidate: `upstream_id,remote_model,enabled,priority,weight`; each upstream occurs once. Writes replace the collection atomically.
- `snapshot(alias)`, `beginAttempt(ctx,request,record,snapshot,tried,round)`, `finishAttempt(id,attempt)` and `observe(snapshot,candidate,attempt,temporary)` split snapshot, dispatch registration, durable completion and runtime observation.

## 3. Field / State Contract
Schema2 removes upstream fixed-ID restriction, adds name/execution revision, model revision/legacy_remote_model, route_candidates and meta.config_revision. requests preserves original rowid/fields and adds evidence_version/config_revision/model_revision/committed_at/successful_upstream_id/attempts_complete. request_attempts, request_route_decisions and upstream_checks persist safe facts. Ciphertext AAD: `upstream:v2:{id}:{url}`. No prompt/completion/raw supplier error/full credential in diagnostics.

Snapshot captures candidate/config/credentials. A short configuration lock serializes dispatch registration with commits and rechecks app authorization and current enablement/removal. Later edits do not silently change the admitted target snapshot. Lock/transaction end before network IO. `adminResponse` buffers management JSON until handlers release locks; slow response writes cannot hold configuration lock.

Priority 0–100 (larger first), weight 1–100; eligible highest-priority group uses smooth weighted round robin with upstream-ID tie break and shared locked counters. Fingerprint resets on eligibility/config change. Exclude tried IDs, maximum 3 attempts. Header30s, whole request300s without reset, precommit attempt60s, SSE idle60s. First stream commit stops precommit timer; total/idle/write deadlines continue. Passive cooldown30s is runtime-only and keyed by alias/upstream execution revision/remote mapping; stale execution observations are ignored and concurrent updates obey dispatch sequence. Manual disablement is not cooldown. Connection checks are independent scoped observations, not generation health.

Historical attempts=null, never invented. Token summaries separately expose total (null if any unknown), known_sum, unknown_attempts and completeness. Dispatch registration is not proof provider executed. Restart marks unfinished evidence incomplete. Only persisted current-config terminal success arms overview; commits invalidate it.

## 4. Validation / Error Matrix
| Boundary | Result |
|---|---|
| schema1 serve / bad schema/material / occupied install | safe nonzero CLI, no implicit migration |
| schema1 valid upgrade / schema2 re-upgrade | backup + atomic migrated / already_current with no new archive |
| bad/duplicate candidates / stale revision | 400 no partial commit / 409 configuration_conflict |
| old singular route / simultaneous same upstream check | 410 management_api_changed / 409 check_in_progress |
| no candidate / exhausted temporary attempts | 503 upstream_unavailable / 502 upstream_attempts_exhausted |
| HTTP408/429/500/502/503/504 or transport/TLS before commit | next distinct eligible target, total budget retained |
| protocol/redirect/auth/local policy/storage error | safe terminal category, no replay |
| cancellation / total deadline / committed output | no next dispatch; canceled or request_timeout |
| check valid2xx JSON list / invalid protocol, HTTP error, network or deadline / caller cancellation | success / failed / canceled; never generation evidence; unresolved historical checks remain unknown |
| check late after credential/URL execution edit | retained observation, applicable=false |

## 5. Good / Base / Bad Cases
Good: two mapped providers receive distinct actual remote names in 3:1 ratio; first temporary failure plus second success has unknown aggregate and visible known partial. Base: re-upgrade no-op, legacy orphan mapping remains editable, unobserved health stays unknown. Bad: id=1 fallback, committed SSE replay, unknown=zero, enabled/check flag used as generation health, DB alteration to fake conservation, network wait under lock.

## 6. Required Tests
`upgrade_test.go` covers atomicity/conservation/material/structure/idempotency with synthetic schema1 boundary fixtures. `cmd/model-relay/deployment_test.go` covers actual read-only CLI and structured schema2 behavior. `configuration_test.go` covers no-generation query, late scope and atomic duplicate rejection. `routing_test.go`/`routing_boundary_test.go` cover ratio/priority/allowlist/unknown/max attempts/cooldown/precommit timeout/total deadline/cancellation/unlocked management response. In-process tests do not prove actual network behavior.

`routing_http_test.go` uses genuine HTTP/concurrency/disable/SSE. `tests/browser/routing.mjs` performs native1280/390px UI with controlled provider counts; `migration.go`/`legacy.mjs` run exact old baseline UI into nonempty data before official backup/upgrade/restore. SQL/API only supplement reconciliation. All enter unchanged `make verify`; listener restrictions leave complete gate pending host. Formal Actions and real suppliers are distinct Not Run boundaries.

CLI installation regressions must use the same route-2 management contract, without singleton fixture translation: `configureCLIInstallation(t, admin)` checks the empty `data,next_cursor` envelope, POSTs an upstream, and POSTs a model with candidates using the returned upstream ID (201). `TestCLIManagementConfigurationContract` exercises the real handler without listening and asserts old `/upstream` is 410 and the old `model` payload is 400. `TestCLIInstallationCommands` still owns actual process startup, occupied-port rejection, health/version/schema2, reset-password/session invalidation, restart, persisted upstream/candidate identity and enabled/disabled application keys; its model list is an envelope, not an array. The metadata-only fixture uses a public IP literal so URL validation needs no DNS; it performs no connection check or generation. Do not add legacy product compatibility to satisfy an outdated driver. A sandbox failure before startup does not validate the later management/restart steps; re-run the unchanged complete entry on the host.

## 7. Wrong vs Correct
Wrong: only final usage, resetting overall deadline on fallback, treating fixture green as real browser acceptance. Correct: atomically save candidates, persist each attempt with nullable usage, conservatively aggregate, stop on commitment/cancel, preserve fixed-entry receipts for QA. Example check input: `{"expected_execution_revision":1}`; example candidate: `{"upstream_id":1,"remote_model":"provider-a","enabled":true,"priority":10,"weight":3}`. Rollback is old binary + pre-upgrade archive + empty target, never merging post-upgrade changes.

## Scenario: SSE scanner size and corrected terminal attribution

### 1. Scope / Trigger
PR #19 pre-merge review found oversized first SSE lines incorrectly replayed and DONE write failures retaining success attribution. This is protocol → routing → SQLite → management evidence, not a new protocol or schema.

### 2. Signatures
`Server.stream(writer, request, ctx, response, record, finish)` owns Scanner error classification. The `streamFinish(resultState, resultCode) bool` closure in `routedGenerate` persists attempts; `generate`'s `finish(state, code) bool` persists the authoritative request terminal through `Store.Finish(record Record) error`.

### 3. Contracts
Scanner `bufio.ErrTooLong` is `upstream_protocol_error`, not a temporary transport interruption. Size limits remain 256 KiB/event and 32 MiB/stream. No backup dispatch or temporary cooldown, before or after commitment. HTTP after commitment remains 200 with an incomplete/error stream, not success. `successful_upstream_id` is nullable and must be null in every non-success request terminal, including a correction after provisional success. Caller cancellation overrides the requested state before clearing attribution and persisting. Attempt and request terminal must agree; overview cannot retain the corrected request as call evidence. No new env, payload, migration or error code.

### 4. Validation & Error Matrix
| Condition | Result |
|---|---|
| Oversized first line | 502 / stream_interrupted / upstream_protocol_error; one attempt, no cooldown |
| Oversized line after valid committed delta | HTTP200 incomplete / same protocol terminal; one attempt, no cooldown |
| Valid delta + DONE delivered | success; actual attempt upstream attribution and overview evidence |
| DONE write fails | canceled / downstream_write_failed; null attribution, no overview evidence |
| Cancellation during DONE write (error or successful write return) | canceled / client_canceled; matching attempt/request, null attribution |

### 5. Good / Base / Bad Cases
Good: size error terminates without generating on the backup. Base: normal DONE preserves success. Bad: classify every scanner error as `stream_interrupted`, or clear attribution only at the beginning of an attempt.

### 6. Tests Required
`sse_terminal_regression_test.go`: `TestSSEOverlongLineNeverSwitchesOrCools` uses management/proxy handlers, isolated Transport and real SQLite before/after commitment; asserts exact error, count, attempt, commit and cooldown. `TestSSEDoneTerminalAttribution` covers success, DONE write error and cancellation with/without write error; reads persisted management detail and overview. These are not network proofs. `TestSSERealHTTPOverlongLineNeverSwitchesOrCools` uses actual isolated HTTP servers for both size boundaries. All automatically enter original `go test ./...` / `make verify`; sandbox restrictions leave host execution pending.

### 7. Wrong vs Correct
Wrong: `classify(ctx, scanner.Err(), true)` for ErrTooLong → temporary fallback; `finish("canceled", ...)` while retaining a previously assigned successful upstream. Correct: match `errors.Is(err, bufio.ErrTooLong)` before transport classification; clear `record.SuccessfulUpstream` for non-success after cancellation override in the authoritative finish closure, and correct canceled DONE delivery after write returns.
