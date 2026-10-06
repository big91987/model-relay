# First-call console ownership

## QA remediation: cancel and completed-task review (2026-10-06)

### Scope and signatures
`upstreamAction(token,callback)` still owns one AbortController and synchronous busy lock. `cancel-check.onclick` aborts this controller without changing upstream config or route health. `resetUpstream()` invalidates it on navigation, pagehide, signout and uncertain session. The check callback rejects completion after abort even when a mocked transport resolves late; catches only its own aborted work, not genuine network failures.

### Contracts and errors
Only `check-upstream` work exposes native button `cancel-check`; its cancellation text explains that reread is needed for final persisted state and no automatic retry occurs. Controls stay locked until settlement, then restored for the current operation only; a disabled/unselected upstream remains uncheckable. Hide cancellation before readback once the terminal POST succeeds. Do not let late readback or old finally unlock a new operation.

`renderSetup(completed)` moves the original five numbered rows (and original buttons/IDs) between `setup-steps` and `completed-steps` inside native `details#completed-setup`. It never clones the evidence button or changes its Request ID handler. Incomplete rows remain current tasks; completed rows are collapsed by default and keyboard-reviewable. `setup-complete` is visible only after all five have facts. Refresh clears completion and disables the old evidence action before GET; failed reads stay unknown. Restart keeps saved first three steps but needs a new current-process generation observation for the last two.

### Good, base, bad and required tests
Good: cancel yields feedback, releases current ownership and allows explicit reread/recheck. Base: no upstream means check disabled; empty installation has all five tasks. Bad: leaving page while late success updates the next editor; all five completed rows still occupy current tasks; cloning nodes loses identity. VM regressions enforce cancel/dedup/late success/navigation/expiry and row identity/ordering/unknown/restart behavior. Real two-screen `routing.mjs` must prove outgoing GET cancellation, failure/edit recovery, total deadline, keyboard-only creation/edit/cancel/check/key/call/stop, folded evidence access and restart review. Tests may inspect inputValue/read APIs but must not set business DOM values or seed via API. Host evidence is required; VM is not a native-browser pass.

## Scenario: restore, submitted parameters and row focus

### Scope and signatures
`upstreamAction(token, callback)` owns one save-or-restore operation and disables `upstream-url/token/enabled`, `save-upstream` and `refresh-upstream` together. `resetUpstream()` invalidates that operation on navigation, sign-out, pagehide and uncertain-session ownership change. `callInputs(disabled)` controls the four submitted generation inputs. `restoreRowFocus(list, identity, previous, token)` restores the same object's state button only for the current owner.

### Contracts and error matrix
- GET-only restore immediately clears the unsaved credential before awaiting `load('upstream')`. Success restores only server URL/enabled and credential-configured metadata. Failure keeps the public draft, marks restore unconfirmed and reports the actual error; no PUT, no automatic retry and no invented success.
- Save and restore share one synchronous busy identity through PUT and its GET readback. Neither can start another action while busy. Unlock only if both operation identity and session/navigation still match; old finally cannot unlock a newer operation. PUT-success/GET-failure continues to disclose acknowledged save plus unconfirmed readback.
- Generation snapshots model/key/message/SSE and then disables those same visible controls. Current terminal success/failure unlocks; cancellation/navigation/session invalidation first severs call identity, clears secrets and unlocks. Stale finally must not unlock a new request; no automatic replay or protocol change.
- Key rows render `known(value.created)` from the same GET record as name/ID/hint/state. Missing/null/empty is “未知”, with no client clock or migration.
- State actions capture the focused button before disabling it. After reread, locate model alias or key ID (not key name or row position). Restore the new state action only if the old action had focus and focus remains there or fell to body on removal; do not steal focus from another input, navigate to another row or cross owner changes. Missing object means no fallback.

### Good, base, bad and tests
Good: restore clears secrets while GET hangs, then restores the confirmed fields; a reordered list with duplicate key names retains focus by ID. Base: unchanged server state still permits restore, and successful calls keep their input values visible until explicit next actions. Bad: only disable send, clear every public draft after a failed restore, unlock in an unowned finally, show Date.now as created or use a replacement row's array position.

VM tests cover save/restore exclusion, duplicate submit, failed readback, late restore versus new save, terminal/cancel/new-call locking, created/unknown, reordered identity and moved/departed focus. Real Go two-screen journeys must independently check pending save/restore barriers with zero extra writes, current submitted values and locks for ordinary/SSE, terminal unlocks, same-ID created metadata and immediate same-row keyboard focus without another Tab. VM is not native focus evidence.

## 1. Scope / Trigger
Native `web/app.js` owns hash navigation, admin state, issued secrets, generation and same-request diagnosis. There is no client storage or frontend runtime dependency. Extend existing passwordVisibility, clearSecrets and health isolation rather than adding another lifecycle.

## 2. Signatures
`owner()` snapshots session/navigation; `current(token)` additionally supports load/find sequence identity; `api(path,method,body,token)` checks ownership after headers and JSON before returning or handling 401. `navigate(view,updateHash)` invalidates before fetching. `stopCall()` invalidates generation before aborting/canceling its reader. Form submit checks `call` synchronously; it does not rely on a disabled button alone.

`scheduleExpiry()` arms a separate session-epoch/deadline-owned timeout; `expireSession()` clears locally at a known deadline regardless of pending session fetch or page navigation. `scheduleSession()` owns only polling; check completion must never be the sole expiry trigger.

## 3. Contracts
Only safe page names enter hash; route-2 design additionally permits encoded public upstream ID/model alias for details, never request IDs, credentials or conversation in URL/localStorage/sessionStorage/IndexedDB. Login restores the safe hash without replay. Visible test generation checks session every 30s or earlier known expiry; other confirmed sessions use 60s plus visibility/pageshow. Management 401 clears the session, application 401 does not. Session network/invalid-response failures invalidate async ownership, abort generation and clear secrets/output without calling them authentication failures. Disable generation and stop automatic checking until explicit recheck succeeds or login restores the session. Keep safe page/model/request ID. Health uses no management headers and the existing independent 5s timeout.

Route-2 target interaction (implementation pending): `docs/workflow/runs/82227c5986e5f41875895d8016c686a6/routing-interaction.md`. Scope busy identities by object/operation; whole model candidate edits are atomic, cancel is zero writes, conflict/unconfirmed writes require GET reconciliation before another explicit write. Explicit connection checks have independent state, 10s bound and cost notice, cannot clear generation cooldown, and are aborted on departure. Overview emphasizes unfinished tasks, scoped expandable evidence and separate account recovery. Preserve all baseline secret-clearing, stale-owner, expiry, exact-ID, keyboard-focus and no-replay tests while replacing singleton forms. Real continuous keyboard journeys must exercise both 1280px and 390px, not only static layout or API setup.

Known expiry stays independently armed while checks hang, across safe page changes and during uncertain-state/manual recovery. Its ownership is session-wide, not navigation-bound. Stop both timers on logout/pagehide, re-arm the known deadline on pageshow. Same-session response metadata may shorten or first establish expiry, never extend a known deadline; check local expiry before accepting a response or error. Late responses cannot revive an expired/new session. Normal successful polling preserves current status, secrets, output and notices; only recovery from uncertainty tells the user to re-enter and manually send.

Page change/logout/expiry/pagehide clears password, upstream credential, issued key, test key/message, readable/raw output. Issued keys start masked and are explicitly revealed/copied; close/reissue increments issuedEpoch to isolate clipboard completion. Logout immediately clears locally and uses a keepalive request to revoke server session even across reload; failed revocation is not called confirmed.

Models are selected from enabled rows only; an invalidated previous selection becomes an empty required option, not another silently chosen model. Cancel create/edit folds the existing draft without writing. Successful creates fold; failed writes preserve public draft. Lists are bounded at 1000, overview counts are not.

Generate renders content text. `appendReadable(value,streaming,operation)` accumulates tool names by choice/tool index without executing or displaying arguments; summarize only on completion, including tool count and first five names (128 characters each). At most 1024 identities are kept; overflow explicitly says “至少”, not an exact count. Empty text without tools is separately explained. SSE uses streaming TextDecoder, CRLF normalization, bounded frame buffer and DONE confirmation; missing DONE/error is a failure, never replayed. Readable and raw displays are independently capped at 262144 characters with explicit truncation. Same request ID is preserved in session memory for precise authenticated lookup, independently of paginated list.

## 4. Validation & Error Matrix
Every fetch/JSON/read/clipboard error/finally checks identity before DOM mutation. Session checks are single-flight for the current owner; a stale check/finally cannot reschedule or overwrite a new session. Old management 401 cannot sign out a new page/session. Invalid address/credential/alias/name gets Chinese recovery and field focus. Failed reads remain unknown/stale, not zero. Exact404 states not yet found/possibly unrecorded, not free/no billing. Clipboard failure gives manual selection, not success. Unknown signing outcome explicitly warns no automatic reissue.

## 5. Good / Base / Bad Cases
Good: canceled old JSON/stream resolves while a new request runs; no output, notice or finally crosses ownership. Base: known enabled dependencies allow manual sending; no automatic probing. Bad: treating abort as identity, JSON displayed as primary prose, capturing key screenshots, or treating application 401 as admin logout.

## 6. Tests Required
`ui.test.mjs` retains password and health regressions and adds deferred generation/issuance/401, repeated submit, UTF8/CRLF increments, cancellation, tool-only results, unknown versus zero and session clearing. `journey.mjs` uses the actual Go service with isolated controlled upstream: all first config steps are UI, same-ID ordinary/SSE/error/disabled metadata, deep links/reload/back/forward, native keyboard and both 1280/390px. No screenshots while secrets are visible; supplier integration is untested without real credentials.

## 7. Wrong vs Correct
Wrong: `await reader.read(); output.textContent += chunk` after cancel; `catch { signedOut(); }` for proxy 401. Correct: `ensure()` after each await checks operation plus session/navigation, while only current management 401 calls signedOut.

## Route-2 native task ownership

### 1. Scope / Trigger
Multi-upstream/candidate editing and explanatory records in actual embedded app.js/index.html/style.css. Keep original session/secret/generation regression contracts; no framework or duplicated browser configuration source.

### 2. Signatures
`upstreamAction(token,callback)`, `resetUpstream()`, `modelCreateControls(disabled)`, `resetModelCreate()`, `renderCandidates(mode,values)`, `readCandidates(mode)`, `load(view,token,pageCursor)` and `findRecord(token)`. Save/check use plural management APIs and expected revision fields from the observed object. Create/edit candidate modes own separate fieldsets and drafts.

### 3. Fields / State
Native labeled fieldsets carry upstream_id/remote_model/enabled/priority/weight; server replaces whole set. Upstream and model creation lock sibling controls synchronously before await. Acknowledged write and readback are separate facts. Operation identity plus session/navigation epoch guards feedback/finally; signedOut, uncertain session, navigation and pagehide invalidate busy ownership. Token inputs clear on save/error/navigation; no automatic write replay. Candidate drafts stay editable after rejection, cancel explicitly discards draft. Account/recovery actions are grouped separately from business navigation. Overview short status links to exact evidence identity; folded details preserve time/revision/scope. Exact records show attempts/decisions/unknown partial totals, historical evidence remains unknown.

### 4. Validation / Error Matrix
Write rejection keeps public draft and shows safe error; upstream secrets clear. Write acknowledged then GET fails says saved but readback unconfirmed, never invites automatic recreate. Concurrent create ignored; stale operation cannot unlock or clear a newer owner's inputs. Management401 resets account; proxy401 remains business failure. Connection check can be unknown/failed/success/not applicable and explicitly warns list-query costs and non-generation semantics. Header/timeout/network/cancel never imply zero cost or automatic retry.

### 5. Good / Base / Bad Cases
Good: UI creates A/B, edits two candidates, cancels draft, compares actual target visits and same-ID attempt detail; keyboard reaches actions at both widths. Base: no configured upstream offers creation before candidate selection; unobserved generation health says unknown. Bad: SQL/API preseed of critical UI steps, only last-attempt token sum, stale readback overwriting fresh draft, technical stack text used as primary overview.

### 6. Required Tests
`ui.test.mjs` owns async fixtures, including creation busy lock/readback failure/account-reset late result. VM fixtures are not native focus/network evidence. Real `journey.mjs` retains old long-run/session/tool regressions; `routing.mjs` adds1280/390px native Tab/Enter creation/check/candidates/priority/3:1/disable/cooldown/SSE/cancel. `legacy.mjs` seeds exact old-version nonempty migration through real pages. Fixed `make verify` host execution and independent QA remain necessary; implementation alone is not an acceptance result.

### 7. Wrong vs Correct
Wrong: `await POST; message('connected')`, retaining old operation ownership on relogin or summing unknown as0. Correct: `await POST; saved=true; await load(...)`, safely report failed readback, reset ownership on account/page boundaries, display unknown and known partial, verify target counters using controlled external provider only.

Wrong: accept tool-only responses by merely asserting “no text”, or retain secrets after a failed session check. Correct: valid ordinary/SSE tool fixtures assert count/name summary and no extra requests; deferred long-stream tests assert 30000ms, clearing, abort, late-result isolation and explicit recovery. Real Go journeys repeat valid tool responses and session-network fault injection on both viewports, without pre-filling first configuration through APIs.

## Scenario: failed writes/generation and restored success evidence

### Scope and signatures
`form('upstream-form', callback)` owns a write acknowledgement flag separate from `load('upstream', token)`. `navigate(view, updateHash = true, requestID = '')` takes an explicit safe identity only for precise actions. `overview-record.onclick` snapshots `overviewEvidence.request_id`; `view-record.onclick` passes `lastRequestID`. Generic `[data-view=records]` must leave the exact input/detail blank, not silently use either identity.

### Contracts and error matrix
- Rejected/unconfirmed PUT: current owner clears only upstream credential, retains address/enabled draft, reports the original Chinese category plus re-entry guidance. A network error is “保存未确认”, not proof no write happened.
- PUT 200 then GET 503/network/unreadable: disclose “保存已受理、已保存，但最新读取未确认”, clear credential, mark latest credential state unconfirmed and offer GET-only `refresh-upstream`. Never repeat PUT automatically or turn the error into a success notice.
- Current application/provider/JSON/SSE/network failure: clear test key/message without `clearSecrets()` (which would destroy readable/raw diagnosis). Preserve public model/SSE, actual ID and record action. Do not sign out the administrator on application 401. Old catch/finally may not clear newer inputs.
- Overview read: discard previous action identity and disable `overview-record` before awaiting; only non-degraded current response enables the action. Display its request ID, alias, `time` as “调用时间”, `upstream_revision` as “配置版本”, and scope from that same `call_evidence` object using textContent. Use `known()` for missing fields; never use Date.now or `value.upstream.revision` as a fallback. Pending/absent/degraded/failed reads replace the complete summary, so old metadata cannot survive. Failed reads cannot leave a clickable old proof.

### Good, base, bad and required tests
Good: restored/relogged overview action precisely queries its displayed ID outside page one, even with a different local failed request. Base: no evidence leaves the action disabled while generic records remains available. Bad: replacing `lastRequestID` with overview evidence globally, clearing all diagnosis on generation failure, or calling “saved” when PUT itself was not acknowledged.

Use `await api(...PUT..., token); saved = true; await load(...)` with current-owner checks in catch, and `navigate('records', true, overviewEvidence.request_id)` only on the evidence action. VM tests cover write/read failures, scoped clearing, stale owner isolation, restored/new-login evidence, unknown/zero diagnosis and generic navigation separation. Real Go UI tests must repeat rejected save → re-read/correct → new generation → exact record at both viewports; VM-only green does not close QA defects.

Wrong: arm expiry only in a pending check's `finally`, or repeatedly invoke an already-fired fake timeout. Correct: the fixture deletes each one-shot timeout before calling it; with initial expiry at45s, fire polling at30s and leave GET session pending, then advance to46s and assert abort/clear/login. Matrix late success/failure/401, cross-page/relogin, uncertain recovery hangs and normal receiving-state refresh. These VM checks belong to make verify but do not constitute browser/network/supplier validation.
