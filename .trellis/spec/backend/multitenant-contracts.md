# Schema 3 console contracts

## Scenario: authoritative management summaries and reconciled Token meaning

### 1. Scope / Trigger
`identity.go` tenant reads, `console_summaries.go` and `console_api.go` catalog reads, and `console_reports.go` request projections own QA47 summary remediation. No new tables, migrations or writes are required. The frozen interactions table remains the source of the model/service/tenant/workbench information contract.

### 2. Signatures
`GET /admin/api/tenants` adds optional `summary` only for platform/real tenant administrators. `GET /admin/api/catalog?tenant=<id>` adds `summary.sources`, nullable `average_duration_ms` and explicit `latency_scope`. Platform-only `include_ungranted=true` includes global model management metadata while still using the selected tenant for request latency. Requests list/detail/export add decimal-string `confirmed_input_tokens` and `confirmed_output_tokens` alongside immutable original response fields.

### 3. Contracts
Tenant owners are active users with actual administrator memberships, not the caller's platform role or pending invite. Member count includes active memberships; Key count includes historical/revoked stable applications. Model aliases are actual tenant grants. Budget/usage/platform cap use `limitsFor`, `usageFor` and the same admission clock as the supported limits API. Developer/finance tenant reads omit the privileged personnel/Key aggregate entirely. Catalog sources use real route candidates and service profiles; current prices must match candidate/upstream execution revisions and effective UTC time, otherwise null. Latency is the selected personnel/tenant scope's finished dispatched-request end-to-end mean, not provider health or a connection-check duration. Confirmed Token sums use append-only charge-entry deltas across all attempts; incomplete original upstream usage remains null. Pending counters explain incomplete totals; zero confirmed entries do not mean zero total usage.

### 4. Validation & Error Matrix
Unauthenticated -> existing401; cross-tenant catalog ->404; nonplatform `include_ungranted=true` ->403; unavailable aggregate storage ->503, never a fabricated zero summary. No sample ->null latency. Changed service execution ->null affected current price; earlier request snapshots remain unchanged. Scoped request IDs are selected before Token aggregation/CSV. Original `total_tokens`/attempts/entries are never rewritten by the summary projection.

### 5. Good / Base / Bad Cases
Good: known first/third dispatch plus unknown second yields6/8 confirmed with one pending, then authoritative3/4 reconciliation yields9/12 once, with all original attempts/snapshots/entries conserved. Base: unmeasured model, no real tenant owner or no current price is explicitly unknown/unassigned. Bad: inferring owner from platform role, leaking unassigned-Key latency to a developer, selecting stale execution prices, or replacing null original usage with confirmed ledger totals.

### 6. Tests Required
`TestConsoleSummariesInProcess` and `ActualHTTP` use production handlers/SQLite, public invitation acceptance and controlled dispatch. Assert real owner/membership/grants/Key/budget summaries, current source/type/priority/price, scope/expansion negatives, unmeasured and scoped latency, invalidated prices, filtered list/detail/CSV confirmed totals, reconciliation replay and immutable history. The in-process variant is not native HTTP evidence. Native `managementSummaryJourney` supplements unchanged J01–J07 at1280/1440/390 with exact live values and context actions; full `make verify` on the host and independent QA are required.

### 7. Wrong vs Correct
Wrong: a column caption or a screenshot proves its financial/identity values; original response null implies no trusted cumulative usage even after reconciliation. Correct: compare current UI cells against same-scope authorized APIs and controlled created objects; display original upstream usage separately from confirmed multi-attempt ledger usage.

## Scenario: independent quota boundaries and shared resource competition

### Scope / signatures
`ResourceLimits` uses nullable Calls/RPM/TPM/Concurrency/Budget and an expected revision; `resourceUsage` test helper reads production `GET tenants/{id}[/keys/{id}]/limits`. `TestConsoleResourceMatrix{InProcess,ActualHTTP}` exercises both scopes and every dimension at0, exact positive capacity and null recovery. `TestConsoleSharedResources{InProcess,ActualHTTP}` uses a start barrier and an explicitly blocked controlled upstream to compete eight calls for two slots, across separate applications or old/new generations of one application.

### Contracts and error matrix
Configured0 ->429, zero upstream dispatch and no other-scope debit. Exact positive capacity ->one admitted reservation, then429 with unchanged tenant **and** Key usage. Restoring null while running must not reset the reservation/history. With single controlled candidate cap100+100 at tenant rate1000000/million, active reserved fee/TPM are200; completed known3+4 settles fee7 with no pending/reserve. Shared concurrency2 admits exactly2 of8, preserves the prior warm-up call, and allows another tenant while the first scope is occupied. New/old generations must have the same application ID, usage and control pool, not two independent accounts.

### Good / base / bad and required checks
Good: all counter comparisons are authoritative HTTP/handler reads backed by SQLite; admitted and rejected contenders are distinguished before releasing the upstream barrier. Base: in-process variants verify transaction logic only, ActualHTTP variants require the host `make verify`. Bad: sequential calls labelled simultaneous, final aggregate only, or altering DB rows/amounts to pass. The existing governed fixture still requires max_tokens100 by default; only the new null-limit matrix explicitly allows omitted max_tokens when unbounded, because the product legitimately stops injecting a cap when no bounded quota applies. This is a labeled test-fixture mode, never the product's supplier behavior. Original Governance ledger19/60 and all prior assertions remain untouched.

## 1. Scope / Trigger
`identity.go`, `console_*`, `accounting.go`, `billing.go` and `legacy_accounting.go` own the design-v0.2.0 console. Existing schema2 protocol/routing regressions remain supported in isolated tests; production CLI init/serve/upgrade targets schema3. Read the current Run's implementation contract before changing these layers.

## 2. Signatures
- `InitConsole(directory,keyFile)` installs schema3; `UpgradeConsole(directory,keyFile,output)` explicitly migrates schema1/2 with a new pre-migration archive. `Open` validates all schemas/material before recovery; operational upgrade opens read-only with respect to request state until its migration transaction.
- Personnel: `/admin/api/login`, `/session`, `/identity`, `/tenants/{id}/members`, `/invitations/accept`. Sessions hold `UserID`; every request obtains current roles from the database. Logout belongs to every authenticated personnel role, not just platform administrators.
- Application operations: `/tenants/{id}/keys`, `/{key}/rotations`, `/retire-previous`, `/revoke`, `/limits`. Stable app-key IDs own generations, limits, authorization and history.
- Finance: `/models/{alias}/prices`, `/bounds`, `/tenant-grants`; `/requests/{id}/reconciliations/{ordinal}` and `/legacy-reconciliations`; `/tenants/{id}/bills`, `/bills/{id}/corrections`, `/export`.
- `admit(record,snapshot,payload,request) (code,status,priorRequestID)` atomically checks both scopes and returns the original safe request ID for a logical duplicate, without generating a second paid request.

## 3. Contracts
Usage aggregate confirmed Token totals are sums of charge_entries input/output deltas with the same server-owned tenant/Key/model/time/request filter as fees and costs. Do not use only the winning request's response usage when multiple attempts may have billed, and do not count cumulative reconciliations twice. Unknown attempts remain explicitly pending and are not filled with zero as a complete usage fact. ConsoleAccountingInProcess and its ActualHTTP variant verify full and exact-request Token scope; only the host can establish the network variant's result.
Personnel cookies cannot call a model; application credentials cannot obtain management authority. Developers read only assigned Key diagnostics and limits, not whole-tenant usage/bills. Membership removal/demotion removes diagnostic assignments without revoking the tenant-owned application. Cross-tenant resources are 404; insufficient roles are 403. All authenticated roles can explicitly log out.

High-risk writes require `Idempotency-Key`; persisted actor/action/path/payload digests reject changed replays. Operations store tenant scope; safe readback rechecks current scope and never returns an issuance secret. Check replay before temporal validation/default effective dates so an old committed write remains readable. New Key/model grants are explicit intersections; empty is not wildcard.

Money is int64 micro-USD, rates are micro-USD per million tokens, read JSON/CSV amounts are decimal strings. Each attempt computes independent cost/fee with half-up rounding; reservation rounds upward. Reconciliation charges cumulative amount minus prior cumulative amount, not separately rounded token deltas. `request_prices` is an immutable admission snapshot; historical evidence is appended in `historical_price_evidence`, never replaced with today's price.

Admission counts/RPM/C/TPM/budget are independent, tenant and stable-Key scopes checked in one transaction. Finite TPM or hard money needs execution-version-bound operator evidence for input/output billing-token caps. Only the bounded text subset is accepted; unsupported tools/multimodal/n>1/oversized output is rejected before dispatch. Send an explicit output cap; provider contract evidence is not proof of a live supplier's behavior.

`attempt_dispatches.possible` is durable before network dispatch. Known failed attempts charge too. Unknown usage retains conservative reservation across windows/restarts; local concurrency releases once. Late attempts cannot replace a terminal execution state. A restarted interrupted request is never automatically replayed. Only an authorized evidence-bearing reconciliation can resolve unknown usage.

Rotation deadline checks use Go RFC3339Nano comparisons, not SQLite Julian-day millisecond rounding. Expired previous generations are retired before rotating the current generation. Never reset all nonretired generation expiries: doing so resurrects an expired credential. A partial unique index permits only one nonretired current generation.

Bills include request periods, legacy-pending periods, closed periods and adjustment-only periods. Closing checks ended/no running/no pending; a duplicate close returns the same immutable bill. Corrections have signed amount/reason/actor/open adjustment period and cannot make the original bill's current total negative. Financial period totals include an adjustment once, while the original bill's current-total display is a trace view. Exports preserve signed decimal numeric amounts and escape formula-capable text.

## 4. Validation / Error Matrix
| Boundary | Result |
|---|---|
| Invalid personnel/session or expired/revoked generation | 401; no unauthorized dispatch |
| Cross-tenant/unassigned object | 404; no other-tenant metadata |
| Developer whole-tenant limits/bills or finance mutation | 403 |
| Missing/changed operation identity or stale revision | 409; no repeated mutation |
| Active grace or attempted stacking at deadline minus 1ns | 409 |
| Independent quota exceeded, including configured zero | 429; no partial debit |
| Missing trusted cap/current price | 503; no claimed hard guarantee |
| Unsupported bounded payload | 400; zero dispatch |
| Current/pending bill close or closed-period usage mutation | 409 |
| Incomplete model-wide price publication | 400; preserve all prior versions |

## 5. Good / Base / Bad Cases
Good: failed-known → failed-unknown → success creates three attempts, two confirmed charges and held unknown liability; one reconciliation appends only the delta. Base: unlimited/soft migrated application remains compatible but unpriced responsibility stays visible. Bad: using a fake supplier as a default, filling unknown with zero, resetting usage on rotation, changing history to make a test green.

## 6. Tests Required
`console_v3_test.go` covers identity/secrets and precision; `console_http_test.go` covers actual HTTP/SSE/fees/shared quota; `console_contract_test.go` covers three attempts, independent dimensions, active reservations, people scopes and nanosecond rotation. `console_migration_test.go` covers schema1/2→3 and official backup/restore; `console_process_test.go` kills/restarts an isolated actual Go-service process and reconciles retained unknown liability. In-process variants are explicitly supplemental, not network evidence. All enter unchanged `make verify` via `go test -race ./...`; listening restrictions leave real HTTP/process/browser tests to the configured host Connector.

## 7. Wrong / Correct
Wrong: `UPDATE key_generations SET expires=? WHERE retired=0`, a single upstream price change that leaves conflicting tenant rates, or labelling mock transport green as full acceptance. Correct: retire ended generations, rotate only the current one, publish every candidate's cost with one consistent tenant rate, preserve first-failure receipts and require the host gate plus independent QA.

## Scenario: migration driver target ownership

### 1. Scope / Trigger
Schema3 production CLI and schema2 compatibility regressions have different targets. The host seq7 browser failed because the old driver built the production CLI but expected schema2; a passing Go suite does not prove the later native journey.

### 2. Signatures
Production `model-relay upgrade --data-dir --key-file --output` calls `UpgradeConsole` (target3). Test-only `tests/browser/schema2` calls `Upgrade` (target2); it is not an operational deployment entry. `migrationJourney()` runs both, rather than changing production to satisfy an obsolete driver.

### 3. Contracts
Upgrade JSON includes `result`, `from_schema`, `schema`; actual `Open().Schema` and `/healthz.storage_schema` must agree. Keep legacy schema1→2 UI, identity/history/mapping conservation, old-credential HTTP and old-binary rollback assertions. In addition, restore the same exact old UI snapshot into isolated schema1/2 sources, run the actual production CLI to schema3, preserve the same rows and mappings, exercise the original credential against the controlled HTTP upstream and log in to the native schema3 console at 1280/1440/390. `RELAY_LEGACY_URL` and `RELAY_LEGACY_PASSWORD` exist only in the child environment; never print password, key, response body or database contents.

### 4. Validation / Error Matrix
| Condition | Required outcome |
|---|---|
| Production target3 paired with expected2 | Reject contract mismatch; fix driver ownership, not product |
| Response schema differs from opened/health schema | Fail |
| Historical row/mapping differs or old key cannot dispatch | Fail |
| Native login, migrated Key enablement or pending history absent | Fail |
| Sandbox cannot listen | Retain restriction; full host entry still pending |

### 5. Good / Base / Bad Cases
Good: both target-specific paths execute and report evidence separately. Base: schema2 compatibility already_current retains schema2 while production upgrade returns migrated/schema3. Bad: changing only `want=3` while still driving schema2 DOM, or reporting the test-only schema2 entry as production CLI validation.

### 6. Tests Required
`TestMigrationUpgradeTargetContracts` builds the two actual executables, uses temporary schema2 installations and asserts disposition/from/schema plus actual reopen. Original migration browser assertions remain intact; `migration-console.mjs` derives current labels from workspace source and checks real authenticated Go APIs. All run through unchanged `make verify`; local in-process/command tests cannot establish browser results.

### 7. Wrong / Correct
Wrong: `candidate=./cmd/model-relay` with `result.Schema==2`. Correct: explicit test-only target2 executable for retained schema2 regression, plus actual production target3 and matching native schema3 journey. No package, CLI, storage or product interface changes are authorized by this repair.
