# Go backend specs

Single Go module; no package map is needed. Runtime ownership is cmd/model-relay (CLI) and internal/relay (HTTP, persistence, proxy). Node is test-only.

## Pre-Development Checklist
Read [runtime contracts](runtime-contracts.md), [console contracts](console-contracts.md) and [multi-upstream contracts](routing-contracts.md), then the relevant source and HTTP tests. Read ../guides/index.md. Do not change authentication or storage for DOM-only features.

For schema3 personnel, tenancy, application governance or finance, also read [multi-tenant contracts](multitenant-contracts.md). This is the production console; schema2 remains a compatibility test surface.

## Quality Check
Run make verify unchanged. Review HTTP methods, authorization, CSRF, persistence and secret redaction if those paths change. Review [console contracts](console-contracts.md) for counts, configuration epochs and safe rejection. See [runtime contracts](runtime-contracts.md) for trusted examples.
