# Model Relay engineering rules

Build one original product: a team-operated model API relay with a management UI and backend. Do not read, copy, translate or adapt New API implementation code, schema or assets. Official protocol and runtime documentation are valid references; cite them in architecture documents.

This project is developed through Agent Platform orchestration. Keep requirements, design, implementation and QA evidence in docs/workflow. Use the user's delegated automatic recommendations: document reasonable choices and proceed. Do not repeatedly request approval to move stages. Ask only for a genuinely ambiguous product boundary, expanded privileges, paid external use or destructive change.

Build a working backend and UI, not a static demonstration. Never implement a fake provider as the product's default. Isolated, explicitly labeled test upstreams are permitted for deterministic HTTP/SSE/error tests, but are not evidence of a real supplier integration. Do not copy any existing unrelated todo application.

Use standard public operational entry points. Persist data outside source files. Never commit tokens, generated admin passwords, encryption keys, SQLite files, local host paths, sessions, node_modules or runtime evidence containing secrets. Avoid logging prompts, completions and credentials by default. Do not alter database records to make tests pass.

The backend, management API, proxy, persistence and operational commands must be implemented in Go. HTML/CSS/JavaScript are allowed for the frontend served by Go. Do not use Node or Python as the product backend. Follow docs/roadmap.md and implement only docs/releases/v0.1.0.md for the first pipeline task; later milestones are not current acceptance gates.

Provide a project-owned Go verification entry such as make verify or scripts/verify.sh: formatting, go vet, go test (including actual HTTP integration), and build. Use locked Go modules. Test processes own isolated temporary directories and ephemeral ports; never touch a user's configured upstream or data. A browser development dependency is allowed only for testing, never as the backend.

The platform's existing browser.check tool serves static roots only: use it for design prototypes, not as evidence that backend APIs work. Backend UI journeys must start the actual Go service and use the project verification entry; the fixed project test Connector runs host-only verification. Independent QA checks real receipts, code/version consistency and failure coverage; do not require broad Agent sandbox elevation to repeat a host-only listening test. Preserve sandbox limitations separately.

Git branch, publishing and draft PR are handled by platform Connectors. Do not merge the source branch, trigger production deployment, change unrelated repositories or widen permissions. No stage should claim real provider validation before credentials and provider responses are actually available.

## Architecture checks

[Package responsibilities](docs/engineering/architecture.md) and [architecture.json](architecture.json) declare production package boundaries and dependency directions. Run `make architecture-check`; read [the integration and scope](docs/engineering/architecture-checks.md). The installed tool in `.harness/archcheck` is versioned from Agent Platform, not maintained locally. Keep its SOURCE.json and use the maintenance installer for upgrades. Rule success does not prove domain-model quality or product acceptance.

新增代码按能力进入对应包：管理 HTTP 只调用能力服务；复杂业务操作拥有完整事务；供应商协议归 `upstream`，网关生命周期归 `relay`。业务服务不依赖 HTTP，HTTP 适配不写 SQL，纯规则和状态包不依赖 HTTP/SQL。入口导航与职责标准见上述包地图。前端每个功能拥有业务请求、交互状态和退出清理，工作台只做组装；不要把新功能重新堆回 shell。
