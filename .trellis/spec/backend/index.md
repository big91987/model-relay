# Go backend specs

One Go module with explicit package ownership. Read the [package map and dependency rules](../../../docs/engineering/architecture.md). cmd/model-relay assembles the application; internal/relay owns HTTP and persistence adapters; internal/routing owns selection policy and runtime state without HTTP or SQL. internal/accounting owns resource and money rules. Node is test-only.

## Pre-Development Checklist
Read [runtime contracts](runtime-contracts.md), [console contracts](console-contracts.md) and [multi-upstream contracts](routing-contracts.md), then the relevant source and HTTP tests. Read ../guides/index.md. Do not change authentication or storage for DOM-only features.

For schema3 personnel, tenancy, application governance or finance, also read [multi-tenant contracts](multitenant-contracts.md). This is the production console; schema2 remains a compatibility test surface.

For deadline/window/budget changes, read its durable shared UTC and zero-soft-budget scenario. Preserve actual HTTP historical red reconstruction and current regression tests in make verify.

For output safety or RPM/TPM verification, read [metering regression contracts](metering-regression.md). Keep in-process, actual HTTP, browser and live-provider evidence distinct.

## Quality Check
Run make verify unchanged. Review HTTP methods, authorization, CSRF, persistence and secret redaction if those paths change. Review [console contracts](console-contracts.md) for counts, configuration epochs and safe rejection. See [runtime contracts](runtime-contracts.md) for trusted examples.
