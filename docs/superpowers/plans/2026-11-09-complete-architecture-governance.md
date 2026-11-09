# Complete Model Relay architecture governance

**Goal:** Close every substantive finding in the installed checker's final 2026-11-07 report on the dedicated remediation branch, with observable package ownership, preserved behavior and a truthful before/after result.

**Architecture:** A modular monolith, organized by existing capabilities. Gateway lifecycle, upstream protocol, routing policy, personnel/tenant access, application keys, resource/finance use cases, records and installation each own their state and operations. HTTP adapters parse/map calls; operation services own rules/atomic persistence. A capability uses extra layers only when it has distinct responsibilities. Frontend shell owns session/tenant/navigation; features own business API/state/cleanup.

**Tech Stack:** Existing Go 1.25, SQLite, vanilla browser modules, installed archcheck0.4.1; no new product runtime or framework.

**Spec:** User's complete-governance request plus docs/engineering/evidence/architecture-2026-11-07/after/model-result.json and current product contracts.

## Global constraints

- Preserve external endpoints, JSON, schema, currency rounding, strict protocol parsing, deadlines/cancellation, idempotency, clock rollback protection, secrets and configuration revision semantics.
- Preserve original operation transaction scopes and audit writes; do not split atomic admission or settlement into independent commits.
- Do not retain aliases, compatibility shims, pass-through application layers, mutable shared Server bags or rename the monolith into another directory.
- Do not change installed checker copies or lower/delete checks to obtain a passing result. Rules and model judgments remain separate; all unresolved findings get explicit dispositions.
- Existing regression assertions remain. Package ownership adaptations may change test imports and fixtures; historical red gates must keep exact assertions and original baselines.
- Work only in Model Relay's dedicated branch. Existing Agent Platform worktree modifications remain untouched. No merge or deployment.

## Review focus

1. Configuration change during request, check and latency adoption: stale attempts/evidence cannot enter the active generation; durable evidence remains before sample adoption.
2. Personnel/tenant changes during UI reads, writes or native dialogs: no stale state, leaked secrets, duplicate writes or released new-editor lock.
3. Quota/price/key changes during admission/settlement/reconciliation: preserve the original single-transaction decision, reservation and audit.
4. Ordinary/SSE upstream errors, terminal events, malformed/duplicate JSON and late cancellation: preserve output commitment and replay restrictions.
5. Installation/upgrade/restart and persisted effective time: preserve identities, original ledger/history and rollback-resistant expiry.

## Tasks and ownership

- [x] Task1: Restore exact branch064eb0b9; establish clean baseline and capability/dependency map. Existing branch archive and Git metadata must match tree87483d47.
- [x] Task2: Extract upstream client, SSRF policy and protocol into internal/upstream (new package tests first). Gateway consumes explicit adapter; retain original protocol/HTTP regression assertions.
- [x] Task3: Extract sessions and configuration/check state owners. Keep explicit observed time and original limit/window/order semantics. Integrate through operations rather than raw maps.
- [x] Task4: Move management and finance business use cases into owned capabilities; separate HTTP request/response handling and SQL operation scopes. Split installation/records from gateway. Module tests and original integration regressions protect behavior.
- [x] Task5: Split all production management UI features and legacy supported UI responsibilities from shared shell, retaining business API/state/cleanup per feature. Existing pinned frontend tests and actual browser journeys remain gates.
- [x] Task6: Declare actual module direction/public entries in architecture.json; update AGENTS.md/package map and add isolated reverse-import probes. Run full verification, installed-tool model rescan and fresh whole-branch review; resolve material findings and publish exact-tree updates to PR79 with before/after evidence.

## Execution rulings

Native core implementation continues in this session. The frontend, new upstream module and new state owners have disjoint write ownership and are processed in parallel under dispatching-parallel-agents; root owns wiring, existing Go files, capability services, configuration and final integration. The previous temporary checkout is gone; the canonical GitHub branch archive has been restored into the project's ignored persistent harness-output/model-relay-source directory, with canonical tree/commit SHA verified. Managed-worktree creation did not return; no worktree appeared in Git's inventory, so this already isolated dedicated checkout is used.

## Completion conditions

Every original model finding is resolved in implementation or explicitly rejected with source/test evidence; no substantive finding is merely relabeled follow-up. All production files have coherent capability ownership and explicit allowed dependencies. The supported regression entry, real browser suites and new module checks pass, with infrastructure failures reported separately and fixed/retried through supported paths. Final comparison includes actual directory/symbol responsibilities and same-policy tool reports, not only file counts or number of findings.

## Integration follow-through

The original118 Go test entries remain under their actual owners. Provider preparation/dispatch ordering, SSE idle-close classification and stale execution guards have focused regression/mutation evidence. Follow-up model candidates were resolved by moving report scope into reporting, sharing transactional model-access facts, translating missing-route errors at catalog, self-coordinating rejected-request observations, aligning current Trellis specs and moving quota presentation into the limits feature. Read-only management reports no longer take the configuration write lock; original write linearization stays. The legacy overview now invalidates a runtime success ID only after a known-stale assessment, preserving it on unassessed storage errors.

Final source is frozen for the supported isolated make verify and same-policy DeepSeek rescan. No new source edits or tests are to be added during those runs; evidence and delivery summary are recorded afterward.

Final same-source make verify completed with exit0 in 936.550 seconds: cold locked Node dependencies, unchanged source/candidate, full race,314 frontend tests,2480 controlled requests with zero failures and all five real-browser phases passed. Final rule scan conforms; final accepted model report retains four candidates with independent dispositions. Two rejected citations were corrected by the model on the identical bundle; judgments and all findings stayed unchanged. Delivery receipts and the before/after explanation are archived under docs/engineering, outside the tested product digest. This commit updates the original remediation branch/PR; no merge or deployment is included.
