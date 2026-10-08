# Native schema3 A workspace

## Scenario: controlled route cleanup has one continuation owner

The interrupted-editor driver releases its barrier, awaits the held route's continuation, then removes only that handler with `page.unroute(pattern,handler)` in a nested finally. Never run continuation and unroute in `Promise.all`: locked Playwright 1.59.1 server `Page.removeRequestInterceptor` notifies in-flight routes, and `Route.removeHandler` automatically continues the current handler. Removal racing an explicit continuation can produce `Route is already handled!`. Do not catch/ignore this error or use `unrouteAll({behavior:'ignoreErrors'})` to turn a broken gate green.

Success receipts are written only after cleanup succeeds. On editor assertion or continuation failure, cleanup still removes the registered handler and the original gate stays red. `workspace.test.mjs` models asynchronous continuation settlement and requires unroute to follow settlement on both success and failed editor-enable assertions. `route-cleanup.test.mjs` executes locked Playwright's actual server Route methods with an isolated transport delegate: removal-first rejects late continue, ordered cleanup delegates exactly once. This is source-level mechanism evidence, not native Chromium/network cancellation validation; all three host `make verify` editor journeys and original assertions remain mandatory.

## Scenario: controlled tenant-readback interception includes collection queries

`collection('tenants')` sends `/admin/api/tenants?page=1&page_size=100&limit=100` and may add a cursor. A Playwright glob ending in `/tenants` does not match query-bearing URLs. The interrupted-editor driver uses `/\/admin\/api\/tenants(?:\?[^#]*)?$/`, holds only the first GET and continues POST and tenant member/detail requests. Never remove pagination or broaden the hold to all management requests to satisfy a test.

`workspace.test.mjs` executes the actual collection function, requires its emitted path to match, and checks bare/query/cursor, POST bypass, detail/unrelated rejection, cleanup, save enablement and cancel/no-extra-write assertions. Seq51's POST201 plus missing controlled checkpoint was a driver mismatch, not proof that the editor-lock repair failed. The locked Playwright matcher proves URL matching locally, not native cancellation or product delivery; only host `make verify` closes the Go/browser gate. Safe failure diagnostics expose `readback_checkpoint_reached` as a boolean, never URLs, bodies or credentials.

## Scenario: editor save-lock ownership after interrupted readback

`edit(title, fields, submit)` initializes both `editor-save.hidden=false` and `editor-save.disabled=false`. The new dialog owns its initial lock; it must not inherit the previous dialog's pending or superseded state. Submit still synchronously sets `state.busy` and disables Save. The old `finally` may unlock only when its captured epoch is current. Never remove that epoch guard to repair a disabled new dialog: a stale completion would then unlock a newer in-flight write.

Acknowledged write closes the dialog before `syncTenants()` and `renderPage()` complete. Navigation can invalidate that readback, preventing the old finally from running its unlock. Good: opening the next editor explicitly initializes its lock. Bad: depend on old readback settlement, add sleeps to the test or unlock every dialog from stale finally. Unknown writes retain GET-only recovery and no replay; no backend method or persistence contract changes.

`workspace.test.mjs` deterministically holds first acknowledged readback, advances the epoch, opens/submits the second editor, and proves old completion cannot unlock it. The driver `save` checks native enablement before arming one response waiter, awaits click and response together, and returns that same response. Native `console.mjs` holds only the temporary service-create tenant GET, navigates, requires the new editor to be enabled, cancels it and requires exactly one real service POST. All three original widths remain mandatory. `test-results/editor-interruption-{width}.json` is a controlled timing receipt, not proof of OS/network cancellation or supplier integration. Native execution remains host-only `make verify`; VM fixtures cannot close that gate.

## Scenario: summary ownership and same-scope context actions

### 1. Scope / Trigger
QA47's model/service/tenant/workbench summaries live in the production `workspace.js` renderer, backed by Go collection projections. Do not copy prototype fixtures or promote QA47 findings to Closed before independent QA.

### 2. Signatures
Model rows consume catalog `summary.sources` and `average_duration_ms`; platform catalog expansion remains scoped for latency. Tenant rows consume authorized `summary` or display a permission boundary. Service rows join actual global models' candidates with actual service/check revisions. `modelCall(alias)` transfers only a public model name after an epoch-owned navigation; `tenantDestination(tenant,page)` selects that tenant before navigating.

### 3. Contracts
Sources/types, per-candidate current USD input/output rates and measured/unknown latency are visible in multi-object rows. Connection-check timestamp/revision/applicability never substitutes for generation latency or a uniform service priority; priority belongs to each associated model candidate. Model enable toggle uses the existing expected-revision PATCH and retains all candidate identities/flags/priorities/weights. Overview uses scoped catalog state and direct model/budget entrances. Tenant owner/Key/budget summaries are not inferred from browser role labels. Request rows retain `Token总量` with explicit original-upstream explanation and separate confirmed input/output Token cells; detail/analysis/CSV use the corresponding immutable and ledger-derived fields.

### 4. Validation & Error Matrix
No sample/current price -> explicit unmeasured/unknown, not0. Restricted tenant summaries -> administrator-permission text, not blank false facts. Epoch superseded during model-call navigation -> do not populate the new page. Existing GET failures keep the renderer's failed-state/retry contract. No application credential is auto-filled by a model-row call action.

### 5. Good / Base / Bad Cases
Good: two models, two tenants and multiple real Keys show actual values and context actions at all three widths. Base: disabled model has real metadata, explicit status and explicit toggle; unsupported authority is retained. Bad: copying demo latencies, selecting only a table header, accepting either of two old/new tenant scopes or changing stored amounts for screenshots.

### 6. Tests Required
`console-summary.mjs` checks exact current-row cells by actual column ownership against authorized APIs and the objects created by original native J01–J07. Preserve multiple-object cardinality, fixed input/output rates, unmeasured states, context call/name/no-secret, disable/enable, association details and same-scope model/budget/Key entrances. `console-visual.mjs` retains all previous gates and adds the new required fields/actions; original63 images remain. J06 checks unknown/pending before reconciliation, original-null/confirmed3+4 after it, request detail, analysis and original+confirmed CSV. `console-summary.test.mjs` tests strict assertion helpers and money precision, not Chromium mechanisms. Existing VM service-check fixtures must supply the new models collection without changing their execution-revision/failed-check assertions.

### 7. Wrong vs Correct
Wrong: green VM tests or old seq46 browser results validate the new renderer. Correct: retain local failure and green receipts, fingerprint the candidate and require the configured host's unchanged full `make verify`, followed by independent frozen-material UI review.

## Scenario: visual editors must use the action's owning surface

### Scope / Signatures
Host seq44 settled limits facts pass title/scope/role/ready/fields/actions/secret/layout gates; `调整租户限额` is enabled, but capture times out at `open-editor` with no open dialog. Production `mainAction` binds `#page-action` in `.page-heading`, outside `#content`. `visualEditorControl(page,definition,row)` in `tests/browser/console-visual.mjs` uses `editor.location:'heading'` only for limits; other editors retain the exact current-object row or content owner.

### Contracts / Error Matrix
Resolve the exact button name within the declared owner, assert count1, then native click and require `#editor[open]`. Missing/duplicate action fails; never search the whole page or switch to another row/content action. Keep all existing editor field, page capture, nonempty, scope, role, layout and secret assertions. This is a driver scope fix, not permission or API behavior.

### Good / Base / Bad / Tests
Good: limits uses the heading action; Key/model/service/billing use the exact live row and invitation uses content. Base: absent action fails before capture. Bad: `#content` lookup for a heading action, global `.first()` fallback, moving production buttons solely to satisfy a driver or dropping the editor check. `console-visual.test.mjs` checks the actual helper's scoped locator contract and production HTML boundary, including no fallback; fixtures are not native browser evidence. `make verify` must still run the original real service, J01–J07 and ten-page review at1280/1440/390 before independent QA can close QA38-01.

## Scenario: nonempty visual review after month-end billing

### 1. Scope / Trigger
Host seq42 proves the first visual page has a settled scoped overview GET200, all required fields/actions, and exactly the legitimate `当前范围近7日暂无调用` empty state. J06 advances to the next month to test closing/correcting the old bill. The later nonempty capture cannot use historical activity as current-period data. Fix the owning journey's preparation, not production rendering or `assertCaptureFacts.ready`.

### 2. Signatures
`prepareVisualActivity(page,key,model,tenantID,billID)` in `tests/browser/journeys.mjs` runs after J06 and before `consoleVisualReview`. It uses existing native `call`, `requestDetail`, `closeDetail` and `nav` helpers. Only overview and closed-bill GETs are direct evidence reads; no API or database business-object seeding.

### 3. Contracts
Confirm current UTC day/month with zero current-day calls and no seven-day trend; perform one real native normal call through the controlled upstream. Require one dispatch (owned by `call`), current-period nonpending detail, one charge with cost7/fee21 at J06's published rates, and authoritative overview with one successful call and one current-day trend row. The full closed-bill response including original amounts and corrections must be unchanged. Navigate away from the call surface before screenshots so secret/output lifecycle clearing still applies. Write only numeric activity facts and a fixed conservation boolean to `test-results/visual-activity-<width>.json`.

### 4. Validation / Error Matrix
Wrong period/day, missing activity, unknown usage, incorrect fee/cost or changed closed bill → fail before visual capture. Loading/errors/any `.empty`, missing fields/rows, wrong role/scope, overflow and visible secrets → retain original capture failures. Actual UI/API/dispatch failure → original stage/receipt; never substitute a fabricated series or reset the clock backward.

### 5. Good / Base / Bad Cases
Good: a native request creates real new-period facts without altering an ended bill. Base: the pre-call legitimate empty trend is asserted, not presented as a populated screenshot. Bad: accepting any empty state as nonempty, backdating records, restoring the old month or weakening production permissions/charging to suit a capture.

### 6. Tests Required
`journeys.test.mjs` executes the actual preparation function with explicit contract fixtures, rejecting missing/stale activity and changed bills; these are helper-contract tests, not proof of native mechanisms. The unchanged `make verify` must execute the actual service and all J01–J07/nonempty ten-page/keyboard/screenshot checks at1280/1440/390. Independent QA reviews those same-version receipts; local helper green does not close QA38-01.

### 7. Wrong / Correct
Wrong: remove `.empty` from `ready` because GET200 returned. Correct: retain `ready`, prove the empty-state prerequisite, generate one real current-period call, verify facts and conservation, then capture.

## Scenario: failed visual capture must retain the assertion boundary

Before `assertCaptureFacts(facts)`, persist `pending_capture` in the existing visual receipt. Keep exact title/scope/role, expected-field/action booleans, row/card counts, root/content dimensions, loading/error-node counts and classified visible empty-state counts. Only classify known labels (`当前范围近7日暂无调用`, `当前范围暂无数据`); unknown DOM text is a count, never a text dump. Record `failure.check` and the fixed capture assertion message, not arbitrary browser errors. Filter actual response events with `overviewReadFact(response,tenant)` to exact GET `/admin/api/workspace/overview`; retain only resource/status/scope-match, at most16 reads, never URL/query/body/account/credentials. Remove the listener even on failure. A prior J06 expected404 is not overview HTTP evidence.

Good: host failure has `pending_capture.facts.ready=false`, classified empties, actual overview GET status and the named assertion. Base: navigation timeout retains a navigation boundary and null facts, not invented DOM facts. Bad: broad DOM/response dump or silently permitting `.empty` based on an unproved hypothesis. `console-visual.test.mjs` proves receipt ordering, strict gate preservation, whitelist and redaction only; its mock facts do not establish the host cause. Keep the original assertion until real host evidence distinguishes legitimate period emptiness from error/loading/role/scope/fields/layout defects. Run the unchanged `make verify` via host, then make only the evidenced repair; QA38-01 stays open until current-fingerprint host captures and independent comparison.

## Scenario: ten-page real-console design evidence

`tests/browser/console.mjs` retains its original heading/root-width smoke loop and native keyboard gate. `journeys.mjs` calls `consoleVisualReview(page,{width,tenant,objects,secrets,nav,step})` from `console-visual.mjs` after the original J01–J07 assertions and before writing the final continuous-success receipt. Objects come from the existing UI-created two tenants/two public models/service/application Keys and reconciled request, not HTTP-seeded completion states. At1280/1440/390 assert exact page title, authenticated role, selected tenant, settled nonempty content, substantive table headers/summary/control labels and actions. Click actual detail/editor controls, wait for the open native dialog and required fields before screenshots; close without writing. Record page, role/scope, assertion booleans, safe fields, image hashes and repository-relative paths in `test-results/visual-{width}/review.json`. Partial/failure receipts retain the last safe phase and previous captures.

Screenshots require cleared passwords/Key/invitation/message controls and absent call output; reject known test secrets in rendered text. A normal untouched call form is applicable empty state, not a failed data list. Wrong role/scope, read failure, missing field/action or root overflow fails the original host entry. Contract/VM tests only validate assertions and production rendering; they do not close AC01. QA independently compares actual host PNGs with the frozen prototype and records deviations. Never replace this gate with static `browser.check`, title-only navigation or synthesized screenshots.

Production Key budget cards reuse `limitSummary` (tenant and Key share the five independent dimensions, exact micro-USD and unknown/reserved explanation). `factSummary` labels the actual `attempts` API field and four price-rate fields; no payload, amount, auth or ledger semantics change. Required regression: `workspace.test.mjs` executes the actual render branch and validates both resource-summary calls and exact price strings. The browser visual suite asserts the corresponding visible detail fields. Good: open price versions with actual cost/tenant rates; base: editor inspection without POST; bad: raw Key JSON as the primary budget presentation or claiming all-page QA from supplemental VM green.

## Scenario: unknown writes retain a safe, repeatable recovery control

### Scope and signatures
`edit(title, fields, submit)` owns the operation UUID, synchronous busy guard and unknown-result latch. `submit(values, operation)` performs at most one unresolved mutation. Recovery calls only `api('operations/'+encodeURIComponent(operation))`; it never retries POST to recover a secret.

### Contracts and error matrix
On a lost acknowledgement (including an abort without a scope change), hide Save and latch the form against additional submission. Display the safe operation reference via `textContent`. Append a dedicated recovery-status node and a separate native `安全读回` button. A transient GET503/404 keeps that button and the reference; update only the status node, never the parent's `textContent`. While GET is pending, disable the recovery button. Success opens only safe operation metadata, not a replacement credential. The captured epoch **and** the original dialog's open state must match before painting either success or failure. Management401 still uses the existing `api` signed-out invalidation, rather than keeping a supposedly authenticated old editor alive.

### Good / base / bad
Good: one committed rotation, lost POST response, GET503 then GET404, same-operation GET200; original application, usage and history remain unchanged. Base: ordinary acknowledged writes retain the previous path. Bad: clearing the recovery parent on failure, replaying POST, showing a late detail after editor close, or calling Contract tests Chromium evidence.

### Required checks
`tests/browser/readback.test.mjs` extracts the production editor for retry/reference, unknown-submit latch, concurrent read deduplication and late-scope/closed-dialog checks. `console-faults.mjs`, imported by the original `console.mjs`, runs actual Go + Chromium recovery for issuance and an already-used application's rotation at1280/1440/390, including native double-click/Enter, stacked-transition409, preserved price/ledger/usage, and ordinary/SSE double-send with real attempt/admission counts. Its historical red is explicitly a test-only reconstruction of the previous editor served by a page route, records its SHA256, and is **not** current production acceptance or a claim about original Chromium execution. Current-asset recovery must separately pass. No static browser tool or VM fixture closes this gate. Partial safe JSON and secret-free screenshots are retained in ignored `test-results/fault-*`; host/QA archive actual files under this Run.

### Wrong vs correct
Wrong: `output.textContent='读回失败…'` deletes both reference and retry. Correct: `status.textContent='读回失败…'` changes only the dedicated sibling; `unknown=true` blocks further form mutation while recovery remains available. These rules avoid accidentally issuing another application credential after a committed-but-unacknowledged write.

## Scenario: reconciliation reads an append-only ledger

### Scope and signatures
`assertReconciledLedger(before, after)` in `tests/browser/journeys.mjs` verifies J06 after the native reconciliation dialog POST and authoritative detail GET. `workspace/requests/{id}` returns `charge_entries` ordered by ordinal/revision, not a replaced total row. `settleAdmission` initially records revision0 with zero confirmed amounts for wholly unknown usage; reconciliation appends the difference at the new admission revision (`before.revision+1`, not necessarily1).

### Contracts and validation
Keep all prior ledger items and price snapshots identical. For the one-attempt controlled 3-input/4-output fixture at cost1000000/tenant3000000 microUSD per million, require cumulative cost7/fee21, confirmed Tokens3/4, one new entry, pending0, reserved0 and matching after/delta revision, ordinal, reason and evidence. Invalid money/Token representations, altered historical rows, duplicate delta, wrong sums, retained pending/reserve or changed snapshots must fail. `ledger_facts` exposes only bounded numeric totals/counts, never request identity, reason, evidence or raw assertions.

### Good / base / bad and tests
Good: sum all append-only entries using BigInt and preserve the first zero-confirmed row. Base: existing three-attempt Go governance scenario confirms unknown ordinal2 revision0 remains unchanged and reconciliation revision2 adds fee5/cost16 before bill totals19/60. Bad: `charge_entries[0].fee === '21'`, modifying the product to replace prior entries, or weakening to nonempty ledger. Actual driver helper regression includes wrong fee/cost/Token/revision, duplicate delta, mutated history, pending/reserve and snapshot negatives; both GovernanceInProcess/ActualHTTP share production handler/SQLite assertions. All join original `make verify`; local helper/in-process passes are not Chromium or actual-network evidence. Preserve seq32 failure and require fresh complete host verification at1280/1440/390.

## Scenario: member removal refresh settlement and failure ownership

### 1. Scope / Trigger
Host seq30 retained a `J05-resume` label after subsequent member operations. All observed pause/resume fee, reservation and count pairs matched; `action:authoritative-read` cannot locate the failing assertion. Do not report this as a proven fee-conservation failure. Source inspection additionally shows `refresh.onclick` awaits `syncTenants()` before `navigate()` creates the loading placeholder, so old content already satisfies `ready()` while tenant synchronization is pending.

### 2. Signatures
Test driver `refresh(page, expectedTenantCount)` waits for the declared selector outcome using native `page.waitForFunction(predicate, count, {timeout:5000})`, then the existing ready check. J05 member removal supplies zero and retains its exact zero-option assertion. No product endpoint, permission, fee rule or UI state is changed. `failureBoundary(error)` emits only the version-bound `tests/browser/journeys.mjs` line/column, never stack text or host directories.

### 3. Contracts
J05 separately attributes member lookup, role write/readback, diagnostic recovery, removal, removed Key scope, membership list, UI scope, application survival and other-tenant Request ID denial. `read_facts` contains only expected/actual status. `membership_facts` contains an allowlisted role and counts. Existing usage facts remain exact decimal strings/counters; they describe the last observed conservation boundary, not every subsequent action.

### 4. Validation / Error Matrix
Pending tenant synchronization with old content -> refresh with expected scope must remain pending. Zero options observed -> existing strict zero assertion runs. Failure to settle within five seconds -> timeout, not acceptance. Wrong role, retained diagnostic membership, non-404 removed scope or cross-tenant read -> retain original failure. Missing own-source stack frame -> `available:false`; arbitrary exception text is never printed.

### 5. Good / Base / Bad Cases
Good: bounded predicate waits for the exact expected removal fact, then asserts it. Base: unchanged ordinary refresh has no extra scope expectation. Bad: sleeping, retrying mutations, accepting either selector count, or changing production billing from a stale label.

### 6. Tests Required
`journeys.test.mjs` executes the actual refresh helper with delayed tenant settlement; pre-fix returns early (red), post-fix waits (green). This proves helper ordering only, not Chromium's seq30 failure location. Position-projection tests reject unrelated stacks and secrets. `TestConsolePeopleScopeInProcess`/`ActualHTTP` additionally check live-session finance role, diagnostic cleanup, reassign-before-delete cascade and an empty post-removal membership list. Full unchanged `make verify` with native J01–J07 at all three widths is still required; preserve original host failures and independent QA No-Go.

### 7. Wrong / Correct
Wrong: `click refresh; ready(); count()` assumes an existing non-loading page means the new scope settled. Correct: declare zero expected options, wait boundedly for that new fact and keep the original count/status/dispatch assertions. Do not claim this source-backed candidate is the native root cause until the host receipt confirms the boundary.

## Scenario: J05 pause/resume conservation diagnosis

The continuous native driver must preserve the original resumed `usage.fee_micro_usd` equality and all J01–J07 actions. `assertUsageConserved(before,after)` supplements it after primary ordinary rejection, primary stream rejection, standby rejection, another tenant's successful call and resume. Compare period, exact string fee/reservation and safe nonnegative calls/pending/concurrency. Never coerce micro-USD to a JS Number or normalize missing fields to zero.

Failure `usage_facts` contains only `period_same` and the before/after amounts and counters for these five fields; invalid values become null and fail, not pass. No tenant/Key IDs, accounts, arbitrary object bodies or request data enter this projection. Good: same nonzero fee/calls survive pause while the other tenant is charged independently. Base: unknown usage reservations remain held. Bad: reduced/reset fee, extra admission, missing metadata or period drift fails the original layer. `journeys.test.mjs` executes the actual projection/assertion and proves exact large amounts, negative comparisons and redaction, not browser/network mechanisms.

`TestConsolePauseUsageInProcess` uses production management/proxy/SQLite with an explicitly labeled transport fixture; `TestConsolePauseUsageActualHTTP` repeats the same contract using temporary HTTP listeners. They require nonzero starting fees, two application Keys, zero dispatch for paused ordinary/SSE calls, independent tenant charge, identical usage before resume and appended usage after resumed calls. Both remain additive in `make verify`. Passing the in-process variant cannot resolve a conflicting native receipt; retain the native failure, run the bounded diagnostic via the configured host and use its evidence before choosing a product fix.

## Scenario: new installation is not a legacy migration

`InitConsole(directory,keyFile)` targets schema3 and removes the transitional `legacy` tenant after creating the schema. `GET /admin/api/upstreams` and `/models` return `data:[]`; `/tenants` returns `items:[]`. The platform administrator's `syncTenants()` still shows one `全部租户` option with empty value; that is a scope control, not an existing tenant. Migration journeys separately retain historical tenant/Key conservation assertions.

`freshInstallation(page)` in `tests/browser/journeys.mjs` must read all three actual collections, require non-null arrays of length zero and the single empty-valued platform scope option before creating the first tenant through UI. Bad: expecting `legacy` on an isolated new installation or adding database records to satisfy that expectation. Good: empty new console; base: migrated console tested in its separate suite; bad input: unexpected existing tenant, null collection or missing scope option must fail.

`TestConsoleFreshInstallationEmptyCollections` verifies real InitConsole/Open, login and actual handlers/SQLite with no listening port, including legacy Key-list 404. `journeys.test.mjs` executes the actual driver segment against empty/nonempty/malformed inputs; neither proves Chromium or network behavior. Native host `make verify` must still pass all J01–J07 at 1280/1440/390. Failed reads record numeric HTTP status; initial stages identify the collection and report numeric counts only, never response bodies or account/secret values. Preserve historical failed receipts rather than closing QA on these supplemental tests.

## Scenario: continuous production journeys and accurate Key states

### 1. Scope / Trigger
QA22-01 found that ten-page navigation and platform-issued calls did not prove frozen J01–J07. QA22-02 found the Key formatter ignored enablement and expiry. Changes belong to the production workspace and the additive browser runner, not schema2 compatibility, authorization fixtures or the frozen material.

### 2. Signatures
`keyStatus(key, now=Date.now())` displays revoked, disabled, expired, active grace, enabled or unknown using safe metadata; it does not authorize a call. `needTenant()` replaces the loading placeholder with the actual no-scope state. Request details offer `在当前范围查看日志`, preserving tenant and existing safe filters while adding an explicit request filter.

`make verify` → existing `make browser` → existing compatibility/native journeys → `fullConsoleJourneys()` → three independent `fullConsoleJourney(width)` installations → `tests/browser/journeys.mjs`. Original gates remain. New test-only environment names are `RELAY_JOURNEY_WIDTH`, `ORIGIN`, `PASSWORD`, `UPSTREAM`, `CONTROL`, `CAPABILITY` with the common `RELAY_JOURNEY_` prefix. Never print secret values.

### 3. Contracts
Key table consumes `enabled`, `revoked`, `expires_at`, `generations[{expires,retired}]`, explicit models and the real `/keys/{id}/limits` fee/reservation/budget. Revocation precedes disablement; enabled expiry is inclusive at the deadline; only unretired, future grace displays transition. Missing state metadata or malformed global expiry is unknown, never guessed active. Authentication remains Go-owned.

Each continuous runner calls production `InitConsole`/`Open`, creates ephemeral loopback servers and deletes its isolated directory afterward. Empty Start State contains only the init-owned, asset-free legacy tenant; no user tenant, endpoint, model, Key, bill or request is seeded. All business mutations use UI; authenticated API reads only check final facts. A separately capability-protected test server exposes `GET /facts` and `POST /advance {seconds:int64}` for monotonic UTC time and dispatch counts. No product control endpoint or SQL mutation is added. Initial test time is the first day of the current UTC month at noon, so the 60-second expiry step cannot accidentally reset monthly usage on a real month-end run. Clock advancement affects the existing Store clock and browser Date; personnel sessions retain their actual wall-clock contract.

### 4. Validation / Error Matrix
- Revoked/disabled/expired/transition/enabled/unknown metadata → distinct informational list labels; never client permission enforcement.
- No tenant → empty guidance replaces loading; a removed member cannot retain old diagnostic DOM.
- Missing test capability → 403; negative or >35-day advance → 400; other test-control routes → 404.
- Expected rejection → exact status and zero upstream POST delta; unexpected status, unavailable scope or failed browser assertion fails the original entry.
- Failed UI read → visible failure/retry, not empty success. Held old-scope real response is released only after scope switch and cannot replace the new tenant.

### 5. Good / Base / Bad Cases
Good: invite and authenticate a real tenant administrator before issuance, reject then recover authorization and hard money, preserve shared rotation usage, pause two Keys while another tenant works, reconcile and close an ended bill, append correction once in its adjustment period. Base: test-only controlled HTTP proves relay semantics, not supplier billing facts. Bad: API-create intermediate business objects, change DB amounts/dates, replace old gates, or mark newly written scripts as browser Pass.

### 6. Tests Required
`journeys.test.mjs` executes the actual formatter at exact expiry/grace boundaries and the actual no-scope renderer; checks driver/product label and rotation-status contracts plus additive gate wiring. These are supplemental Node contracts, not native evidence. `journeys.mjs` must execute J01–J07 at 1280/1440/390 against real Go/SQLite, including native keyboard invalid-field recovery and submission, UI-issued >25 records, delayed reads, same-ID fees, all Key states and before/after secret-free screenshots. Preserve original Go/Node/HTTP/SSE/cancel/unknown-issuance/keyboard/migration/history gates.

### 7. Wrong / Correct
Wrong: reuse seq21's old fingerprint green result to close QA22 or infer that VM fixtures prove Chromium/network behavior. Correct: generate the new source fingerprint, obtain full host `make verify` and independent QA; keep the original No-Go and failed attempts. Emit only fixed stage/action/error categories, numeric status, checkpoint IDs and DOM booleans/counts. Files under ignored `test-results/continuous-{width}-{pid}.json` and associated screenshots carry actual execution evidence; do not invent them before host execution.

## 1. Scope / Trigger
Go embeds `web/workspace.html`, `workspace.css` and `workspace.js` for schema3. Preserve schema2 `index.html/app.js/style.css` and their trusted regressions; they are not the current production screen. No client framework or Node backend.

## 2. Signatures
`api(path,method,body,operation)` checks personnel/epoch ownership and handles 401. `collection(path)` follows server page/cursor metadata without silently cutting off the last page. `invalidate()` aborts reads/calls and clears secrets on navigation/scope/logout/pagehide. `edit(title,fields,submit)` owns a synchronous busy state and one operation ID. `runAction` catches async action errors for the current epoch. `limitSummary` displays independent units, used/held/remaining and platform upper bounds.

## 3. Contracts
`detail` consumes multiple shapes, not just a resource-limit report. `limitSummary` must dispatch to structured `factSummary` when either limits or usage is absent; services, Keys, requests, operation readback and bills must not throw for missing quota fields. Use textContent for scalar values, never HTML from an API. Preserve the same native dialog ownership, focus and secret cleanup across all summaries.

Table search is local to the fetched safe collection (or current log page); label it as current-list search, not a whole-database query. Search before slicing groups and reset to the first filtered group. The server-side log/usage filters remain authoritative for tenant/model/Key/time scope. Token analytics display confirmed ledger input/output separately from pending usage; a successful response with unknown usage is not a zero-token fact.
Ten fixed A pages use the real server identity and tenant selector, not a demo role switch. Lists, prices, risk and trends derive from Go APIs. Source kind is explicit supplier/self-hosted/legacy-unspecified; capabilities describe the relay protocol, not an invented supplier certification. Model details show cost versions and current/stale operator caps. Money uses BigInt and exact decimal strings.

Secret dialogs only hold the received issuance value in memory. Closing, navigation, tenant switch and logout erase it and conversation fields. A committed write with a lost response is unknown: hide the save action and offer operation-only safe readback, not blind retry. Browser route fault injection must commit the real backend write before aborting the response. The proof cannot be a fake UI fixture.

`dialog.close()` queues a native close event. When the same dialog is immediately reopened for a one-time value, the old close event must not erase the newly opened dialog. Only a currently closed dialog's handler cleans its contents. Busy Escape is prevented; native Tab/focus trap/Escape return are tested in a real browser.

Editor close cleanup is scoped to `editor-fields`: erase its password/secret/sensitive controls and detach the fields, error and submit handler. Do not invoke global `clearSensitive()` from that queued event: the explicitly transferred `call-key` belongs to the destination page, not the closed editor. Global navigation/logout/scope/pagehide invalidation still clears all sensitive controls. The explicit transfer captures the post-navigation epoch and writes only if that epoch and playground destination remain current; an intervening navigation, even to another playground, must not restore the captured secret.

`workspace.test.mjs` exercises actual close and transfer callbacks with close-before/after-settlement schedules, reopened-dialog protection and superseded-epoch rejection. These are deterministic ownership tests, not proof of Chromium scheduling. Native `console.mjs` requires a nonempty transferred control and boolean equality to the UI-issued value at `key-secret-transfer-value` on each viewport, then retains HTTP200 and same-Request-ID detail assertions. Never fill the call Key from a test API response, add arbitrary sleeps or remove the original gate. Failure snapshots remain boolean-only. Host seq17 proves issuance201 and empty call input for the old candidate; the revised candidate still needs full host verification.

Refreshing tenants updates both the array and select; creation must not leave a stale selector. Version reads capture the post-navigation epoch. Failed identity/permission reads clear sensitive fields; failures are not empty datasets. SSE accepts LF or CRLF, bounds local frames/output, displays chunks incrementally and treats missing DONE/usage as interruption/pending rather than fabricated success. A local cancel says upstream fees may remain unknown.

## 4. Validation / Error Matrix

### Keyboard failure attribution
`console.mjs` delegates only the existing create-dialog/13 Tab/Escape/return assertions to `keyboardJourney(page,width)`. `keyboardSnapshot(page)` is a read-only DOM projection: dialog-open, focus containment, body/root focus and document-focus booleans. A bounded 16-sample trace carries only fixed phase enums, Tab counts and these booleans; never emit active element IDs, labels, values, HTML or browser errors. Keep the original first-Tab and final-cycle containment assertions, immediate Escape visibility assertion and exact trigger-return assertion unchanged. Do not insert focus assignment, sleeps, broad retry or an easier assertion into the driver.

Host seq19 passed beyond issuance/explicit transfer/call/ledger/quota and failed with an open editor at coarse `keyboard-1280`. That alone does not distinguish first Tab, cycle boundary or Escape. Source inspection shows `keyEditor` launches `Promise.all(...).then(edit)` and returns before the dialog opens; `page.keyboard.press` does not target or await that dialog. The driver must await `#editor` visibility at `keyboard-dialog-ready` before sending the original first Tab, not force focus or sleep. Use `keyboard-dialog`, `keyboard-dialog-ready`, `keyboard-first-tab`, `keyboard-tab-cycle-N`, `keyboard-escape`, `keyboard-return-focus` plus the safe trace to choose any further product/driver/timing fix. `workspace.test.mjs` executes the actual driver with success, delayed opening and four rejecting boundary fixtures. The delayed fixture is red without readiness and green with it, proving a driver readiness contract, not Chromium scheduling or the unique cause of seq19. No product focus repair is claimed; fresh host `make verify` remains required.

| Event | Result |
|---|---|
| Late response from old tenant/navigation | Abort/suppress; never paint current scope |
| 401 | Clear personnel/scope/content/secrets; show login |
| Permission failure | Clear secrets, show actual failure; refresh obtains current role |
| Duplicate submit | One operation/call only |
| Committed write, lost acknowledgement | Unknown + safe readback; no new secret replay |
| Readback fails | Still unknown; not success |
| Secret dialog closes or page leaves | Erased, no browser storage |
| Canceled SSE | Local stop, same Request ID, potentially pending fee |

## 5. Good / Base / Bad Cases
Good: real invitation acceptance/login, tenant-scoped issuance, ordinary/SSE/cancel → same-ID diagnostics. Base: no supplier/price results means explicit unknown/empty, with real configuration actions. Bad: prototype amounts, client role switching, truncating a 101-item collection to 100, or unhandled async rejection after an action.

## 6. Tests Required
`workspace.test.mjs` extracts the actual fact/limit summary and table functions for supplemental VM regressions: non-quota detail objects, null/unknown, quota dimensions, filtering beyond the first group and clearing the filter. Source boundaries must include the full function but not leave the following async prefix; fixture observations must read the same textContent the product writes. `frontend-test` includes original ui.test.mjs and the new *.test.mjs suite without replacing native browser gates.
`tests/browser/console.go` starts isolated real Go and explicitly controlled HTTP/SSE upstreams. `console.mjs` covers 1280/1440/390, ten pages, invitation/login/logout, ordinary/SSE/cancel, immediate rotation, quota rejection, commit-before-abort issuance recovery, native dialog keyboard/secret cleanup and secret-free screenshots. Existing journeys and VM tests remain in `make verify`. The static-only platform browser tool cannot validate this backend UI.

## 7. Wrong / Correct
Migration console drivers must consume the production `replyCollection` envelope `{items,total,page,page_size}`, not a schema2 or protocol `data` envelope. `migration-contract.mjs` requires both migrated Key identities and their enabled flags, plus exact count/total; `migration-contract.test.mjs` rejects the obsolete envelope, missing/extra identities and changed flags. `TestConsoleKeyCollectionContract` exercises the real login/create/list Go handlers and SQLite with a recorder, not a listening browser. These supplemental tests cannot certify native journeys.

`migration-console.mjs` splits failures into keys-navigation/keys-http/keys-contract/history per viewport. Its failure receipt permits only suite/stage/result, numeric HTTP statuses, a validated integer schema and numeric item count. Reset these facts at each viewport; never print raw errors, response bodies, arbitrary fields or secrets. Keep native row visibility, history and overflow assertions intact and obtain a new host `make verify` receipt.

Wrong: unconditional queued close cleanup clears the newly issued value; checking a checkbox with an invented accessible name; treating static rendering as backend evidence. Correct: guard closed-state cleanup, derive exact labels from actual HTML/JS, register bounded real response waits and inspect the host receipt before claiming browser acceptance.

## Scenario: schema3 service connection check

### 1. Scope / Trigger
Host seq11 stopped at service-1280. The actual service row action sent `{}`; the existing Go handler rejects a missing execution revision before dispatch. Fix the caller contract, not the server guard. The old coarse assertion is not proof of the sole native failure cause.

### 2. Signatures
`renderPage()` service-row action calls `api('upstreams/'+row.id+'/check','POST',{expected_execution_revision:row.execution_revision})`. `Server.checkUpstream` in configuration.go owns revision validation and connection-only GET `/models`.

### 3. Contracts
Use the safe listed `execution_revision`, not the management `revision`. Name-only metadata revisions and execution revisions can differ. A 200 check response is a completed observation, not necessarily success: require state/applicability/current revision for a successful controlled-native assertion. Connection success does not certify generation or a real supplier. Controlled model-list fixtures must include `object:"list"` and a non-null `data` array; never loosen protocol validation to fit a fixture.

### 4. Validation / Error Matrix
Missing expected execution revision → 400/no dispatch. Mismatched revision → 409/no dispatch, surfaced notice/no automatic retry. Valid current revision and protocol list → 200/success/applicable. HTTP200 with malformed list → 200/failed/upstream_protocol_error, not certified success.

### 5. Good / Base / Bad Cases
Good: row revision19/execution4 sends `{expected_execution_revision:4}`. Base: rejected stale row exposes failure and requires operator refresh. Bad: `{}`, a management revision, or auto-refresh-and-recheck against an unreviewed new configuration.

### 6. Tests Required
`workspace.test.mjs` executes the actual render function and service action for exact payload/success and conflict/no retry. `TestConsoleServiceCheckRevisionContract` runs schema3 login/create/check handlers and SQLite, with explicitly isolated transport; it proves missing/mismatched rejection, current success, GET-only and zero generation attempts, not listening behavior. `console.mjs` retains native three-viewpoint creation/check and requires status200, a positive integer sent revision, success/applicability and matching revision. All are in unchanged `make verify`; full native checks still require the host.

### 7. Wrong / Correct
Wrong: logging a check body or raw assertion to diagnose failures. Correct: split service-create/check-http/check-contract/check-dialog and emit only numeric write/check statuses and revision-present/success/applicable/revision-match booleans, reset per viewport. Preserve every original gate; no credentials, IDs, URLs, body or DOM dump in the failure receipt.

## Scenario: native model editor locator ownership

### 1. Scope / Trigger
Host seq13 completed the service check successfully but stopped at model-1280 with Error. Source inspection identifies a deterministic driver ambiguity; this does not establish the only host failure. Do not change product labels or backend validation to fit a test selector.

### 2. Signatures
`modelEditor(model)` constructs both the `remote` field labeled `远端模型名` and the `routes` textarea whose label includes that same phrase. `console.mjs` must use `page.getByLabel('远端模型名',{exact:true})` for the remote input.

### 3. Contracts
Accessible label substring matching is not unique here. Select the actual exact label; preserve the additional route field and all create/grant assertions. Do not add a first/nth fallback, programmatic product value assignment or API preseed. The native driver requires one exact control before filling it.

### 4. Validation / Error Matrix
The actual editor produces two substring matches and one exact match. The supplemental VM regression rejects an ambiguous match, and the corrected driver fills alias/upstream/remote only. Native model-dialog/fields/create-http/grants-dialog/grants-http stages distinguish locator, write and readback failures. Failure receipts expose only fixed stage, numeric statuses/counts, booleans and an allowlisted error kind; never the raw exception message or page/API contents.

### 5. Good / Base / Bad Cases
Good: exact remote label selects the intended input without editing routes. Base: a missing or duplicated exact label fails rather than selecting another field. Bad: substring lookup, `.first()`, changing visible product copy solely to satisfy a driver or claiming the VM's label simulation proves Chromium behavior.

### 6. Tests Required
`workspace.test.mjs` executes actual `modelEditor` field construction and the actual model-field segment of `console.mjs`, requiring the exact three intended writes. This is a supplemental driver contract test, not real DOM/auth/network evidence. The original three-viewport real Go console journey remains in unchanged `make verify`, with all model creation/tenant grants, later Key/billing/identity and keyboard assertions intact. Obtain a fresh host receipt before closing the native failure.

### 7. Wrong / Correct
Wrong: treating a coarse Error as proof of a supplier/Go failure or repeatedly changing async behavior without diagnostics. Correct: preserve seq13, demonstrate the exact source mismatch with a red/green regression, add bounded stages and defer actual browser mechanism and the remaining journey to the host.

## Scenario: coarse Key journey timeout

### 1. Scope / Trigger
The host seq15 failed at `key-1280`, which covered navigation, asynchronous catalog/member dependencies, issuance, one-time dialog, explicit transfer, ordinary call and same-ID detail. That receipt does not locate the failing action. Do not change product cleanup, waits or provider logic from that coarse timeout alone.

### 2. Signatures
`consoleSnapshot(page)` executes a read-only DOM projection through the running browser. `console.mjs` retains the original actions/HTTP status and same-ID assertions and emits `key-navigation`, `key-dialog`, `key-fields`, `key-model-selection`, `key-create-http`, `key-secret-dialog`, `key-secret-transfer`, `key-call-fields`, `key-call-http`, `key-request-detail-action` and `key-request-detail-visible` with viewport suffixes.

### 3. Contracts
Snapshot fields are fixed page enums, boolean dialog/control/failure states and numeric invalid-control/option/selection counts. `call_key_nonempty` is only a boolean, never the value or length. Response observations retain only numeric catalog/member/issuance/call/detail statuses, never paths, headers or bodies. Page errors and failed requests are counted without printing their messages. Failure projection allowlists exception names/kinds and excludes secrets, prompts, completions, Request IDs, account names and raw API/notice/editor text.

### 4. Validation / Error Matrix
Missing DOM controls produce false/zero and unknown page, not fabricated success. An unavailable page produces `snapshot_available:false` and retains the failing stage. Dependency failure and invalid form are separate from HTTP rejection; an issuance success does not establish secret transfer or call success. A timeout with no call response and an empty call field is evidence to investigate transfer/validation, not proof of native event ordering.

### 5. Good / Base / Bad Cases
Good: obtain the new native stage and safe status/DOM facts before selecting the next fix. Base: no extra observations on success beyond the existing journey. Bad: dumping an error stack, DOM, network body or screenshot while a one-time secret is visible; substituting an API-issued Key; delaying arbitrarily or weakening assertions to force green.

### 6. Tests Required
`workspace.test.mjs` executes the actual snapshot function with secret-bearing and missing-control fixtures and verifies the exact projection and non-disclosure. These tests prove projection only, not Chromium lifecycle or backend behavior. The full unchanged `make verify` remains the host gate with all three console viewports, migration/Go/SSE/identity/billing assertions. Diagnosis-only candidates must not be reported as repaired or accepted.

### 7. Wrong / Correct
Wrong: infer that seq15 `write_status:201` is Key issuance; it may still be the preceding price publication. Correct: use the new `key-create-http` stage and `key_create_status` to bind the observation to the Key request, preserving seq15 as failed historical evidence.

## Scenario: continuous controlled completion protocol

### 1. Scope / Trigger
Host seq26 reached J01's first proxy call but returned 502. The continuous fixture omitted `message.role` in both known and unknown usage responses. Production `validResponse` requires `assistant`; do not relax product parsing to fit a test supplier.

### 2. Signatures
`continuousUpstream(dispatched *atomic.Int64) http.Handler` in tests/browser/journeys.go is the actual browser upstream handler and the regression target. `completionFacts(value)` in journeys.mjs projects authoritative request detail only after a status mismatch.

### 3. Contracts
The controlled handler emits `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`. Known usage remains 3 input / 4 output / 7 total; unknown usage omits usage entirely, retaining pending liability. Do not seed business objects or change records to satisfy the journey. Failure evidence records expected/actual status and dispatch delta, plus terminal/code from explicit allowlists, numeric upstream status, pending flag, attempt count and at most three attempt summaries. Never output IDs, raw bodies, credentials, prompts, completions, upstream names or arbitrary error strings. Detail lookup is GET-only, bounded to five seconds; diagnostic failure must not replace the original status assertion.

### 4. Validation / Error Matrix
| Condition | Required outcome |
|---|---|
| Normal completion missing/wrong assistant role | Keep production protocol rejection |
| Controlled assistant completion, known usage | 200; cost 7 / tenant fee 14 at journey rates |
| Controlled assistant completion, usage omitted | 200; pending remains 1, not fabricated zero |
| Failure detail unavailable | Retain original failed status; report unavailable safely |
| Arbitrary fields/error values in detail | Exclude or replace with `other`; no raw data |

### 5. Good / Base / Bad Cases
Good: repair the owning test upstream and retain the negative parser regression. Base: an omitted supplier usage field is a valid completion but an unresolved financial liability. Bad: accept missing roles, replace unknown usage with zero, or claim helper tests passed all browser journeys.

### 6. Tests Required
`TestContinuousUpstreamResponseContract` invokes the actual fixture with known and unknown requests (non-listening). `TestCompletionAssistantRoleContract` invokes the production validator with missing/wrong/assistant roles. `TestContinuousUpstreamRelayHTTP` uses actual temporary upstream HTTP through schema3 relay handlers, management APIs and persisted request/charge detail; it requires host listening and is not a complete browser or relay-network journey. The Node projection regression excludes secret-bearing/unrecognized fields and bounds attempts. All enter the unchanged make verify; full continuous J01–J07 at 1280/1440/390 and independent QA still decide acceptance.

### 7. Wrong / Correct
Wrong: `message:{content:"controlled"}` or weakening `validResponse`. Correct: `message:{role:"assistant",content:"controlled"}`, with missing-role rejection intact and separately attributed host evidence.

## Scenario: request detail entry belongs to the current page

### 1. Scope / Trigger
Host seq55 completed the three editor-interruption viewports and J01–J05/J07 at1280, then J06 timed out waiting for request detail. The reconciled request is on the records page; the call-page action does not exist there. This is a test-driver entry mismatch, not evidence of missing backend reconciliation or a listening failure.

### 2. Signatures
`requestDetail(page,id,entry='call')` in `tests/browser/journeys.mjs`. The explicit `records` entry uses `row(page,id).getByRole('button',{name:'详情 / 核对',exact:true})`; the default retains the call-page `查看同Request ID日志` action.

### 3. Contracts
Arm the same exact request-detail HTTP response before native activation. Await response and click with `Promise.all` so either rejection has an attached handler. Require200, matching response ID and visible matching detail title. Do not replace the UI step with `page.request.get`, relax the timeout or change product controls.

### 4. Validation / Error Matrix
Call-page entry -> original local Request ID action. Records-page entry -> exact request row action. Click/HTTP rejection -> propagate failure, not a partial success. Missing controls -> keep failure; do not fall back to a different page or request.

### 5. Good / Base / Bad Cases
Good: J06 explicitly selects `records` after reconciling and preserves confirmed-token/detail/export assertions. Base: all call-page uses keep the default entry. Bad: activate the absent call action on records, leave an orphan response promise, seed an API detail as UI evidence, or claim a VM fixture proves Chromium behavior.

### 6. Tests Required
`journeys.test.mjs` executes the actual helper with call/records locator contracts, response-before-click ordering and click failure. It binds J06 to the records entry. These prove the driver contract only. Current-fingerprint host `make verify` must still complete every J01–J07, management summary,63 visual images/roles, migration and backend gate; independent QA owns closure.

### 7. Wrong / Correct
Wrong: `requestDetail(admin,unknown.id)` after navigating to records. Correct: `requestDetail(admin,unknown.id,'records')`, with unchanged HTTP/detail/token/export assertions.
