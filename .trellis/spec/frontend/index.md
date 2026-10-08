# Native management UI specs

The frontend is internal/relay/web/{index.html,app.js,style.css}; no framework, build transform or TypeScript layer.

Schema3 production uses [workspace contracts](workspace-contracts.md) and `workspace.{html,css,js}`. Read it before changing the production console; the earlier surface remains for compatibility regressions.

## Pre-Development Checklist
Read [DOM and lifecycle](dom-lifecycle.md), [console ownership](console-lifecycle.md) and [verification](verification.md), then ../guides/index.md and applicable thinking guides. Trace signedIn, signedOut and clearSecrets before editing secret controls.

## Quality Check
Re-read these three guides; inspect changed DOM IDs/semantics and state transitions. Run make verify. VM tests are not evidence of native keyboard behavior or real authentication. Keep health/version regression tests intact.
