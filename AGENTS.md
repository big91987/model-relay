# Model Relay engineering rules

Build one original product: a team-operated model API relay with a management UI and backend. Do not read, copy, translate or adapt New API implementation code, schema or assets. Official protocol and runtime documentation are valid references; cite them in architecture documents.

This project is developed through Agent Platform orchestration. Keep requirements, design, implementation and QA evidence in docs/workflow. Use the user's delegated automatic recommendations: document reasonable choices and proceed. Do not repeatedly request approval to move stages. Ask only for a genuinely ambiguous product boundary, expanded privileges, paid external use or destructive change.

Build a working backend and UI, not a static demonstration. Never implement a fake provider as the product's default. Isolated, explicitly labeled test upstreams are permitted for deterministic HTTP/SSE/error tests, but are not evidence of a real supplier integration. Do not copy any existing unrelated todo application.

Use standard public operational entry points. Persist data outside source files. Never commit tokens, generated admin passwords, encryption keys, SQLite files, local host paths, sessions, node_modules or runtime evidence containing secrets. Avoid logging prompts, completions and credentials by default. Do not alter database records to make tests pass.

The backend, management API, proxy, persistence and operational commands must be implemented in Go. HTML/CSS/JavaScript are allowed for the frontend served by Go. Do not use Node or Python as the product backend. Follow docs/roadmap.md and implement only docs/releases/v0.1.0.md for the first pipeline task; later milestones are not current acceptance gates.

Provide a project-owned Go verification entry such as make verify or scripts/verify.sh: formatting, go vet, go test (including actual HTTP integration), and build. Use locked Go modules. Test processes own isolated temporary directories and ephemeral ports; never touch a user's configured upstream or data. A browser development dependency is allowed only for testing, never as the backend.

The platform's existing browser.check tool serves static roots only: use it for design prototypes, not as evidence that backend APIs work. Backend UI journeys must start the actual Go service and use the project verification entry; the fixed project test Connector runs host-only verification. Independent QA checks real receipts, code/version consistency and failure coverage; do not require broad Agent sandbox elevation to repeat a host-only listening test. Preserve sandbox limitations separately.

Git branch, publishing and draft PR are handled by platform Connectors. Do not merge the source branch, trigger production deployment, change unrelated repositories or widen permissions. No stage should claim real provider validation before credentials and provider responses are actually available.
