# Console facts and request identity

## 1. Scope / Trigger
The first-call slice crosses Store → admin Server → native UI. `console.go`, `server.go` and `proxy.go` own its executable baseline contracts. The first-call change required no migration. Issue #18's route-2 design supersedes the singleton portions below; it requires schema 2, not a new external service. See the target-design section; do not describe unimplemented design as runtime behavior.

## 2. Signatures
`overview(http.ResponseWriter, *http.Request)` serves GET `/admin/api/overview` under the existing admin lock and session guard. `record(...)` serves GET `/admin/api/requests/{id}`. `identifyKey(*http.Request) (string, bool)` recognizes a digest internally without granting access. `configurationChanged()` must run only under `server.mu` after a successful management write.

## 3. Contracts
Overview: upstream `{configured,enabled,credential_configured,revision}`, models/keys `{total,enabled}`, `degraded`, `call_evidence:null|{request_id,time,alias,model,key_id,upstream_revision,scope}`. Counts use full SQL COUNT/SUM in one read transaction, not the existing 1000-row lists. The scope is one persisted success under the current process/configuration, not continuous provider health.

Login/session add RFC3339 `expires_at` from the unchanged 8h server expiry. Exact query returns existing `Record`; missing status/duration/usage stay null. Requests use random 16-byte, unpadded base64url IDs (22 characters), parameterized SQL and shared `recordColumns`/`scanRecord`. No token, prompt, completion or upstream URL in records or overview.

Each successful upstream/model/key write increments process-local `configEpoch` and clears `successID`. Generate captures the epoch with its configuration under RLock. Only a successfully persisted `Finish(success)` at the same epoch registers an ID; later canceled correction clears only its own candidate. Overview rechecks persisted state, enabled identities, mapping and revision before reporting evidence. Restart starts with no candidate; old records remain.

## 4. Validation & Error Matrix
- Missing admin session → unchanged 401; exact invalid/not found ID → 404; exact non-GET → 405 with Allow GET; SQL failure → 503, not empty counts.
- Unknown/missing/disabled application key → identical external 401 unauthorized, zero upstream POST. Start/Finish rejected metadata first; unknown keys have empty identity/alias/model and no body read. Known disabled keys parse only the bounded validated request and may resolve the public mapping.
- Second authorization failure also records the same ID. Storage Start/Finish failure → existing degraded/fail-closed 503, never an in-memory fake record.
- Invalid alias → do not persist arbitrary body text as a public alias. Body caps, read deadlines and proxy error envelope remain unchanged; route-2 adds bounded distinct-upstream switching before commitment, never replay after output/cancellation (routing-contracts.md).

## 5. Good / Base / Bad Cases
Good: known count=1006 even if list has 1000; usage=0 remains actual zero. Base: empty DB known zero, no success evidence. Bad: health derived from enabled flag, old/late success after mapping change, unknown credential attributed from its hint, or a query failure rendered zero.

## 6. Tests Required
`console_test.go` checks full counts, exact/admin/method boundaries, real/unknown usage, terminal correction, all management invalidations, late config race, restart, safe denied metadata and storage failure. `TestConsoleRealHTTP` repeats success and rejection over real Go HTTP. Existing proxy/HTTP/SSE/race/health-history regressions remain in `make verify`; restricted listeners are not passes.

## 7. Wrong vs Correct
Wrong: `len(list)` is the total; `successID=id` before Finish; identify disabled key and authorize it. Correct: `SELECT count(*),coalesce(sum(enabled),0)`, persist before short locked candidate update, and return 401 for both unknown and disabled keys while retaining only safe internal identity.

## Route-2 source contracts (2026-10-06, host acceptance pending)
Current executable signatures and state fields are in [routing-contracts.md](routing-contracts.md), superseding baseline single-upstream examples above. Implementation is present; complete HTTP/browser acceptance remains pending host. Below are retained design constraints, not verification receipts.
Authoritative design: `docs/workflow/runs/82227c5986e5f41875895d8016c686a6/routing-architecture.md` sections 4–8; UI: `routing-interaction.md` in that Run. Replace the singular upstream aggregate with full upstream/model/candidate counts and scoped call evidence. Use plural `/admin/api/upstreams` and revision-guarded whole-candidate model writes; old singular management endpoint returns 410, never a fixed-id routing fallback. Public generation/model-list contracts remain compatible.

New exact request details add ordered attempts, selection decisions, per-item known/unknown usage totals and evidence completeness. Preserve original Record fields and historical rowid/nullable values; legacy attempts are null, never invented. Only persisted terminal success under current configuration can arm overview evidence; name/address/token checks do not establish generation health. Store/network dispatch atomicity is not promised; interrupted registered dispatch remains unknown. Implementation must replace the baseline field/signature examples above with verified ones and add real HTTP/race coverage through unchanged `make verify`.
