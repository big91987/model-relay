# Code reuse checklist

- Search app.js for existing lifecycle handling before adding a helper.
- Is the shared concept one invariant, or merely similar text? Avoid speculative abstractions.
- Trace clearSecrets, signedIn and signedOut: do all exit paths restore the same state?
- Compare all DOM IDs against index.html and VM fixture.
- Keep native buttons native; avoid duplicating click with keydown handling.
- Follow ../frontend/dom-lifecycle.md and ../frontend/verification.md for executable rules.
