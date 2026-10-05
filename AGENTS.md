# Model Relay engineering rules

Build one original product: a team-operated model API relay with a management UI and backend. Do not read, copy, translate or adapt New API implementation code, schema or assets. Official protocol and runtime documentation are valid references; cite them in architecture documents.

This project is developed through Agent Platform orchestration. Keep requirements, design, implementation and QA evidence in docs/workflow. Use the user's delegated automatic recommendations: document reasonable choices and proceed. Do not repeatedly request approval to move stages. Ask only for a genuinely ambiguous product boundary, expanded privileges, paid external use or destructive change.

Build a working backend and UI, not a static demonstration. Never implement a fake provider as the product's default. Isolated, explicitly labeled test upstreams are permitted for deterministic HTTP/SSE/error tests, but are not evidence of a real supplier integration. Do not copy any existing unrelated todo application.

Use standard public operational entry points. Persist data outside source files. Never commit tokens, generated admin passwords, encryption keys, SQLite files, local host paths, sessions, node_modules or runtime evidence containing secrets. Avoid logging prompts, completions and credentials by default. Do not alter database records to make tests pass.

Use npm test as the complete automated acceptance entry. Include real HTTP integration tests and project-owned real Chromium journeys against the running backend. Node.js 24.14+ or 26 is available; prefer built-in HTTP, crypto and SQLite for the small first slice, and a locked Playwright development dependency for browser tests. Tests own isolated temporary directories and ephemeral ports; never touch a user's configured upstream or data. Browser binaries are prepared by the operator, not downloaded by tests.

The platform's existing browser.check tool serves static roots only: use it for design prototypes, not as evidence that backend APIs work. Backend UI journeys belong in npm test and must start the actual service; the fixed project test Connector runs them on the host. Independent QA checks real receipts, code/version consistency and failure coverage; do not require broad Agent sandbox elevation to repeat a host-only listening test. Preserve sandbox limitations separately.

Git branch, publishing and draft PR are handled by platform Connectors. Do not merge the source branch, trigger production deployment, change unrelated repositories or widen permissions. No stage should claim real provider validation before credentials and provider responses are actually available.
