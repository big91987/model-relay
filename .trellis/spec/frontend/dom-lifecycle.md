# DOM and lifecycle

## Native controls
index.html supplies stable IDs and labels; app.js uses element(id), native form submit handlers and onclick. Use type=button for non-submit actions, explicit label associations and textContent for untrusted text. Do not add framework hooks, innerHTML rendering or custom keyboard activation to native buttons. style.css supplies focus-visible outline and a 600px responsive breakpoint; keep changes local.

## State ownership
app.js api(path, method = 'GET', body) owns management requests and 401 signedOut handling. form prevents native navigation, guarded disables the submitting button and displays safe errors. signedIn clears the admin input; clearSecrets clears all secret inputs and output; signedOut aborts requests and clears secrets before showing login. Extend these boundaries rather than add parallel cleanup listeners. A secret must not be copied into URL, storage, telemetry or display messages.

## Async rendering
loadVersion uses request identity, navigationVersion and AbortController to reject stale completions. It reads /healthz without management headers and displays unknown on failure without logging out. Navigation and logout invalidate version work. Keep these independent from password visibility; tests/browser/ui.test.mjs already covers stale, aborted, unknown, retry and raw versions.

## Boundary checks
Changing only input type must retain the same node and value. Never rebuild the form or use the password to derive state text. Shared lifecycle cleanup must remain correct for login, session restoration, logout and management 401. Add native browser coverage for keyboard, semantics and unintended form submission; see verification.md.

## Admin password visibility contract
app.js passwordVisibility(visible) owns type, password-toggle text and password-visibility-status text. The click handler derives the next state from input.type, not a second flag. Good: passwordVisibility(element('password').type === 'password'). Wrong: copying input.value into a second visible input or browser storage; this duplicates secrets and bypasses cleanup.

Hidden/base: type=password, action 显示密码, status 密码已隐藏. Shown: type=text, action 隐藏密码, status 密码已显示. Empty input is valid for toggling; toggling neither submits nor runs constraint validation nor sends requests. signedIn and clearSecrets call passwordVisibility(false) after clearing values; API 401 still calls signedOut and therefore clears/hides even rejected login input. Other errors retain the existing error behavior.

index.html supplies type=button, aria-controls=password, aria-describedby=password-visibility-status, role=status and aria-live=polite. Dynamic action names are not combined with aria-pressed. Native Tab/Enter/Space and click use only onclick; keydown would risk double activation. ui.test.mjs covers original node/value, no fetch and lifecycle boundaries; journey.mjs proves native activation, focus outline, button/status AX snapshots, zero submit/invalid/network, URL/storage invariance and real two-state login/logout/restoration at both viewport sizes. AX snapshots must never include a revealed password input.
