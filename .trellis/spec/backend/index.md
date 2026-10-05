# Go backend specs

Single Go module; no package map is needed. Runtime ownership is cmd/model-relay (CLI) and internal/relay (HTTP, persistence, proxy). Node is test-only.

## Pre-Development Checklist
Read [runtime contracts](runtime-contracts.md), then the relevant source and HTTP tests. Read ../guides/index.md. Do not change authentication or storage for DOM-only features.

## Quality Check
Run make verify unchanged. Review HTTP methods, authorization, CSRF, persistence and secret redaction if those paths change. See [runtime contracts](runtime-contracts.md) for trusted examples.
