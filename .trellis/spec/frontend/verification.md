# UI verification

## Fast feedback
node --test tests/browser/ui.test.mjs executes app.js in a VM with deferred fetch responses and timers. Extend fixture DOM shape when introducing native properties; do not pretend the VM implements browser focus, constraint validation or accessibility.

## Real journeys
make verify includes tests/browser/main.go and journey.mjs using locked Playwright from package-lock.json. The Go harness owns temporary state and ephemeral listeners. The journey covers 1280px and 390px, real login/logout/session, management, HTTP/SSE/error/cancel and storage/log exclusions. Keep these tests; add scenarios to this journey rather than a fake static substitute. Use native Tab/Enter/Space for activation and ariaSnapshot for accessible semantics. No screen-reader audio claim.

## Safe evidence
Record only method/path/status counts and named checkpoints; never print passwords, request bodies, storage secrets, traces or displayed-password screenshots. Compare secrets as booleans so assertion errors cannot echo values. Verify localStorage/sessionStorage/IndexedDB/URL invariance for local-only controls. Keep failure receipts and host-only limitations distinct from passes; make verify is the authoritative entry.
