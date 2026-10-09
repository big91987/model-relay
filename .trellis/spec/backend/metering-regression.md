# Scenario: bounded generation and numeric metering evidence

## 1. Scope / trigger

Issue #34 addresses missing evidence, not an established RPM/TPM production failure. Source owners: accounting.go (`admit`, `boundedPayload`, `usageFor`, `settleAdmission`), protocol.go (`validateRequest`), routed_proxy.go and console_reports.go. Never add a guessed UI control just because a historical test checked for it.

## 2. Signatures

`Store.admit(Record, routeSnapshot, map[string]json.RawMessage, *http.Request) (code string, status int, priorID string)` checks tenant and stable-Key resources in one transaction. `boundedPayload(payload, maximumOutput int64) bool` injects or validates an explicit cap. `governedInstallationWithHandler(test, actual, controlled)` is test-only; a nil handler retains all original three-attempt fixture assertions. `TestMetering*ActualHTTP` uses a real local Go upstream and relay; `TestMetering*InProcess` is supplemental.

## 3. Contracts

Finite TPM or a finite hard budget requires execution-version-bound input/output evidence. Missing output parameters get max_tokens equal to the minimum configured candidate output bound. Explicit max_tokens or max_completion_tokens must be positive and within the cap; both together are forbidden. An unlimited/soft-only request is not automatically a bounded request. Operator evidence and parameter enforcement do not prove provider billing or hidden-reasoning safety.

RPM counts admitted logical requests in fixed effective UTC minutes. Calls count the UTC month; C counts running admissions. TPM is confirmed charge-entry Token deltas for the admission minute plus all active/unknown reservations, including older minutes. Known 3+4 settles to7; unknown does not become known zero. Actual cost and tenant fee use different immutable price rates. In the metering fixture, candidate3 cost rates3/6 and tenant rates1/1 micro-USD per Token produce cost33/fee7. Test prices are not supplier prices.

Diagnostics use `error_code`, not `code`. Limits GET exposes `period` and `budget_remaining_micro_usd`, but no minute field. METERING_SNAPSHOT therefore explicitly labels its fixed-minute observation as coming from committed `Store.now()` through the existing test-clock seam; never claim it was returned by the API. Whitelist counters/limits/safe IDs only; no prompts, outputs, credentials or raw provider bodies in evidence.

## 4. Validation / error matrix

| Case | Expected |
|---|---|
| max_tokens=0 | protocol invalid_request400, no admission/dispatch |
| thinking field | protocol unsupported_field400; not a claim of supported vendor thinking controls |
| positive cap above bound, dual caps, n=2, tools in bounded mode | unsupported_bounded_request400, no partial debit |
| Current U7 + P200 + B200, limit406 | scope_tpm_exceeded429, unchanged tenant/Key/dispatch |
| Same resources, limit407 | accepts, then known7 converts the new reservation |
| Output reported101 against bound100 | records actual known usage, removes violated bound; next bounded request trusted_bound_required503 |
| Dispatched timeout/cancel without usage | retains RPM/Calls and P200; releases local C once |
| Admitted but demonstrably undispatched cancel | retains RPM/Calls, releases C/TPM/fee reservation, no attempt/charge |

## 5. Good / base / bad cases

Good: synchronize8 contenders behind one remaining slot, hold the upstream open, observe7 zero-attempt rejections before releasing1 accepted call, and separately read both Key and tenant pools. Base: default output64 is received by the controlled ordinary/SSE handler. Bad: equating a JSON rejection with successful SSE, logging only field names, treating mocked buffering as a real downstream flush, or resetting budget history.

## 6. Tests required

metering_regression_test.go covers safety triggers, protocol/admission errors, exact and one-less TPM with U/P/B, RPM N/N+1 and exact-minute seam, JSON/SSE known/unknown/inconsistent usage, shared Key/generation competition, timeout/cancel and bound invalidation. Undispatched settlement directly exercises production admission/dispatch contracts, not an HTTP journey. The in-process cancellation transport buffers the upstream body: it proves cancellation after upstream prefix writing, NOT after client output; only ActualHTTP reads the actual first SSE line before canceling.

Existing console_resource_matrix_test.go retains all independent zero/exact/null limits, console_contract_test.go retains multi-attempt reconciliation, and console_process_test.go retains real-process crash conservation. All are part of unchanged make verify. A sandbox listener failure leaves host verification pending; handler green never substitutes for it. Preserve driver failures separately from product reds.

## 7. Wrong vs correct

### Bounded fixture scheduling after host tests5

The same-source host receipt passed every metering ActualHTTP/InProcess case with race detection but the relay package took423.472s; the complete twelve-minute gate later terminated during the390px routing cooldown display. This is failed complete verification, not a metering product failure or proof that the display is correct.

Each `TestMetering*` top-level case now calls `test.Parallel()` before creating its fixture. The unchanged Makefile `-parallel=4` bounds concurrency. `consoleInstallation(test)` owns its TempDir, SQLite/key, cleanup and password; each governed fixture owns its Server, atomic clock, transport, observation channels and HTTP listeners. Subcases sharing that fixture remain ordered; contention barriers remain unchanged. Never parallelize operations within a shared fixture, mutate DefaultTransport or process environment, reduce password cost, alter limits/time, or drop ActualHTTP cases for speed. Local eight-family in-process race evidence verifies isolation only; full cold host duration and browser behavior must still be measured separately.

Wrong: `detail["code"] == "key_rpm_exceeded"`, treating a missing minute JSON field as an observed server window, or labelling test-only fee7 as provider charges. Correct: `detail["error_code"]`, safe numeric snapshots with an explicit clock source, and separate live-provider Not Run/Blocked rows.

### Fixed-gate cumulative budget after host tests7

The unchanged tests7 candidate passed both routing widths, including actual cooling response/target DOM and natural recovery. The relay package took363.407s, and the twelve-minute gate then terminated in full-console1440px J03. This does not reproduce a cooling product defect or a J03 failure independent of cancellation. Record the complete state/log, not nested runner tests' expected failures or a truncated Connector projection.

`TestConsoleGovernance*`, `ResourceDimensions*`, `PauseUsage*`, `SharedResources*`, `ReviewZeroSoftBudget*`, `ReviewClockRollback*`, `ClockQueuedDeadline*`, `ClockRecovery*`, `AccountingInProcess` and `ActualHTTPQuotaBillingAndSSE` now opt into the same top-level `Parallel()` bound. These enforce this task's independent limits, sharing, accounting, queued deadline and recovery contracts. Every invocation creates its own `consoleInstallation`/governed fixture; clocks, server locks, DB and channels are not shared across tests. Recovery children receive their private state-file path through `exec.Cmd.Env`, not process-wide `Setenv`; backup/restore directories and ephemeral listeners are owned by that fixture. Subcase sequences, competition workers, child processes and all assertions remain unchanged. No production code, hash cost, deadline, test recipe, dependency lock or browser journey changes are allowed for this scheduling repair.

Required verification: same selected in-process race command before/after; a combined metering/resource/recovery race run; unchanged format/vet/build and all-package compilation; new exact-source complete host `make verify`. Good: retain the old failed gate and measure package wall time. Base: local speed improvement proves only that subset, not cold-gate capacity. Bad: remove recovery subprocesses or HTTP cases, interpret a parallel parent's duration as package wall time, or claim supplier validation from controlled traces.
