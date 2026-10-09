# Native schema3 A workspace

## Scenario: synchronous release after explicit dialog close (Issue32 execution55)

Execution54 on candidate `00d8519e5b54d97eccce2cae8652735bec46319108044d0d075c96da9b33ce8c` isolates `body-unlock`: the native modal is gone, root overflow is visible and the exact trigger is focused, but body is still fixed/locked and the passive observer has received zero close events. The original models-detail PNG and sidecar passed. This is not a screenshot failure. The HTML Standard's dialog closing algorithm removes modal ownership and restores focus synchronously, then queues the close event on the user interaction task source (official reference: https://html.spec.whatwg.org/multipage/interactive-elements.html#close-the-dialog). Do not require a queued close listener to establish synchronous product scroll ownership.

`closeModal(dialog)` calls native `dialog.close()` then the existing `releaseModalScroll()`. Explicit detail/editor buttons, successful writes, one-time-key transfer and detail-to-editor transitions use this shared entry. `invalidate()` retains its existing close-both-then-release path. Native close listeners remain for browser cancellation and sensitive cleanup; `releaseModalScroll()` remains idempotent and refuses to unlock while any native modal is open. The helper neither restores focus manually nor waits for events. Native cancellation ordering is not inferred from the VM.

Good/base: after explicit close with no other modal, remove body lock/CSS offsets and restore the saved background position before returning, even without a delivered close event. A later close event does not unlock a newly opened modal; closing one of two open modals keeps the lock. Bad: add sleeps/assert retries, weaken the body/focus assertions, or alter API/security/fees to make UI evidence pass. Extracted-editor VM fixtures explicitly inject `closeModal: dialog => dialog.close()` alongside `openModal`; they prove editor business lifecycle only. Actual shared-helper regressions cover immediate release, reopen/event idempotence, remaining modal ownership and native-close-before-release ordering. Full wildcard Node plus original three-viewport HTTP/browser host `make verify` remain required; local green is not native proof.

## Scenario: distinguish successful capture from modal-close failure (Issue32 execution53)

Execution52's `models-detail.png.integrity.json` is passed, with zero background scroll and stable document geometry; the checkpoint already contains that screenshot. Its enclosing visual receipt still labels the next failure `screenshot-content-integrity`, without closure facts. This does not prove another screenshot defect or identify which close assertion failed. Do not change product scroll/focus behavior on that evidence alone.

`closeVisualModal(page,control,selector,name)` arms a passive native `close` listener immediately before the original close-button click, retains the original hidden wait, then calls `assertModalClosed(page,control,observer)`. The assertions remain: no native modal, root no longer clipped, body no longer fixed, and exact original trigger focused. There are no sleeps, extra settlement waits, retries or alternative triggers. The listener samples after the already-registered product close listener; it records only count and booleans for native ownership/root clip/body lock. Zero close events is an observation, not a permitted pass or proof of timing causation.

Set the visual check to `close-detail` / `close-editor` before closing. `modalCloseFailureFacts(error)` projects one fixed boundary (`native-ownership`, `root-unlock`, `body-unlock`, `trigger-focus`), normalized CSS enums, finite scroll coordinates, trigger/focus booleans and fixed event facts. Never emit DOM text, control values, object IDs, URLs or arbitrary error messages. Cleanup always removes the passive listener and disposes its handle; cleanup failure must not replace an existing assertion failure. Success stores closure facts beside the original screenshot; failure keeps that screenshot and its sidecar unchanged.

Required tests: reject each original close failure; distinguish undelivered vs delivered close events; retain exact trigger checks; prove one click/one assertion path and passive observer cleanup; do not serialize arbitrary injected fields or replace the original failure on cleanup errors. VM facts test diagnostic contracts only. Native closure/event ordering is still verified by the original host `make verify` at all three viewports.

## Extracted editor VM dependency contract (Issue32 execution51)

When a VM fixture extracts `edit` or `detail` without the shared modal implementation, explicitly inject `openModal: dialog => dialog.showModal()` alongside its other dependencies. Execution50's full wildcard Node gate found ten `ReferenceError: openModal is not defined` failures in `readback.test.mjs` and `save-lifecycle.test.mjs`; the earlier selected 116 tests did not include these files. Keep their mutation counts, readback identity, historical failure reconstruction, duplicate-submit, late-scope and one-time-secret assertions unchanged. The adapter delegates to the fixture's existing dialog; it proves editor lifecycle only, never native focus or body scroll locking. `workspace.test.mjs` separately exercises the actual shared modal helper; the complete wildcard Node gate and host `make verify` remain required. Do not suppress missing dependencies in product code or treat selected-test success as full-gate success.

## Scenario: fixed background ownership for native modals (Issue32 execution49)

Execution48 on candidate `dcdab7fc4bbdbbcc944377450a30d36cb6e20a67c65443cfb950de22a779f193` disproves the execution47 CSS-only lock: at1280px keys-detail has native modal, root overflow clip, dialog auto, and focus inside, but the original one-shot `scrollBy(0,100)` moves the background100px. Four earlier modal probes returned zero, which is not evidence that a scrollable background is locked. Full gate exits2; no keys-detail screenshot or new self_hosted branch is reached. Preserve the failure and prior document-coordinate/PNG assertions; do not infer a Chromium-internal cause.

`openModal(dialog)` is the shared entry for details, editors and one-time-secret/invitation dialogs. Before native `showModal`, capture `{left:window.scrollX,top:window.scrollY}` once and set only body-owned `--modal-scroll-top` / `--modal-scroll-left` and `.modal-scroll-locked`. CSS fixes that body at the negative captured offset and width100%; retain the original native root clip and dialog auto. Never transform or reposition the top-layer dialog, change its keyboard ownership, or change the screenshot API. These offsets are local layout state, not storage or a second secret holder.

`releaseModalScroll()` does nothing while any native modal remains open. After the last close it removes just the owned class/properties, restores the captured window position once, and clears ownership. Editor/detail native close events cover Escape and button closes; `invalidate()` closes both dialogs and releases synchronously before sensitive cleanup for navigation/session/logout/pagehide. Queued close from an editor immediately reopened must not release a new modal's background. A throwing `showModal` releases when no modal exists and rethrows unchanged. No mutation observer, polling, timeouts, framework or general style snapshot is needed.

Good: a long page stays visually at its previous position while modal contents can scroll; final close restores root scrolling and the exact native trigger focus. Base: no modal leaves existing page/table scrolling unchanged. Bad: relying on root computed clip alone, fixing body permanently, overwriting an unrelated inline style, restoring while a replacement modal is open, or bypassing existing secret cleanup. Unit tests cover single capture/reopen/queued close/idempotent release, native opening failure and synchronous invalidation order. They prove lifecycle contracts only, not browser mechanics.

The original `assertModalScrollFacts` now additionally requires `body_position:'fixed'`; it still requires native_modal, root clip, dialog auto, background delta0 and focus_inside. `modal_scroll_checks` also records only root scroll/client heights for bounded attribution. `assertModalClosed` additionally rejects a fixed body after close, retaining root unlock and exact trigger-focus assertions. Original one-shot fullPage PNG, pixel/document geometry, secret safety, all business/HTTP/Node/race checks and three-width self_hosted journeys remain mandatory in cold host `make verify`; local green is not native acceptance. Execution47's proposal below is historical and superseded, not a verified current contract.

## Scenario: native modal background scroll boundary (Issue32 execution47)

Execution46 on candidate `88794346a0c678b418671a625365a555a28bd0626c2533492c692428cfc9a0a8` retained a nonblank keys-editor PNG at390px. The allowlisted computed position is `fixed` before/after; root scroll182→327 and dialog document top254.75→399.75, but viewport top72.75 is unchanged. This establishes background scroll displacement during the capture, not its browser-internal cause or a historical blank-image fix.

`workspace.css` owns the boundary: `html:has(dialog:modal){overflow:clip}` applies only while a native modal is open. Preserve native `showModal`, close/Escape/focus ownership and the dialog's `overflow:auto`; do not lock the dialog itself, set a permanent root lock, add a layout framework, move its coordinates in screenshot code or change secrets cleanup. Closing the last modal removes this CSS rule through native modal state, without a separate JavaScript lifecycle flag. Background scroll position restoration is native focus behavior, not a promise of exact offset restoration.

`assertModalScrollFacts(facts)` accepts only `{native_modal:true,root_overflow:'clip',dialog_overflow:'auto',background_scroll_delta:0,focus_inside:true}`. The actual Go UI visual journey performs one bounded `window.scrollBy(0,100)` with each open modal and records these allowlisted facts before its original fullPage capture in `modal_scroll_checks` (maximum32). It also requires root scroll unlock and focus restored to the exact trigger after each visual editor/detail close. This is a background-scroll boundary probe, not a UI business write or a screenshot retry. No secrets, DOM text or arbitrary styles are serialized.

Good: modal blocks both user/background scroll while its long form retains internal scrolling and focus. Base: no modal means this rule/probe is inactive; ordinary page/table scrolling remains covered by existing journeys. Bad: `overflow:hidden` still permits programmatic background scrolling; clipping the dialog makes its controls unreachable; changing document-coordinate assertions to viewport coordinates hides the original failure. Missing ownership, wrong overflow, nonzero movement or lost focus fails before the unchanged original pixel/document-geometry gate. Node tests check CSS and assertion contracts only; cold host `make verify` must still establish actual Chromium behavior, complete all three widths and self_hosted journeys, and feed independent QA.

## Scenario: local horizontal boundary containment and loss-aware trace (Issue32 execution41)

### Scope / evidence
Execution40 associates seven retained compositor snapshots with the main page's same-process tree. After the call marker, the substantive content quads have right edges from -2399 to -320, entirely left of the viewport, while DOM scrollX remains0 and the sole original1280×1149 PNG is still uniformRGB243/246/250. This establishes actual main-page draw displacement, not its full causal chain: trace loss=true and the return marker is missing. Preserve that receipt and QA14 No-Go; never classify this as the native listening restriction or declare historical failures explained.

### Signatures / contracts
`workspace.css` applies `.table-wrap,.technical-value{overscroll-behavior-x:none}`. Local scroll containers keep `overflow:auto`/`overflow-x:auto`, native focus and complete text. At a horizontal boundary they must not chain to ancestor scroll containers or produce a local horizontal overscroll affordance. Vertical page scrolling remains auto; no root overflow hiding, JS preventDefault, recentering, replacement input, removed columns or secret exposure. This is the smallest product boundary correction for keeping the page readable during local horizontal gestures; its effectiveness on Chromium still requires actual host evidence.

`readabilityEvidence(page,captureRight)` adds `horizontal_boundaries_isolated:boolean`, computed from every table and identifier's `getComputedStyle(...).overscrollBehaviorX`; `assertReadability(facts)` rejects anything other than true. `readabilityFailureFacts(error)` retains this single safe boolean alongside all original failure fields. All original keyboard-scrollend, trusted wheel100000, exact last-column bounds/2000ms, single fullPage PNG, pixel/geometry/secret and business gates remain.

`observeCaptureFrameTrace(session)` retains ReportEvents, recordUntilFull,512KiB, eight snapshots/sixteen layers, four markers and unchanged identity projection/cleanup. Enable only `disabled-by-default-devtools.timeline.layers` and `disabled-by-default-devtools.timeline`, excluding wildcard and screenshot categories as before; stop additionally enabling broad `disabled-by-default-cc.debug`. FixedChromium147 `frame_viewer_instrumentation.h:CategoryLayerTree()` declares cc.debug/viz.quads/devtools.timeline.layers as a grouped snapshot category, so the layers category still enables the same snapshots without all cc.debug events. Host receipts, not this source fact or mock, decide whether loss and missing markers actually disappear.

### Validation / examples / required tests
Good: original native scroll and original pixel gates pass with actual safe readable PNGs, and host trace has page attribution plus both markers and explicit loss state. Base: local Node proves policy projection, CSS contract and unchanged bounded trace configuration only. Bad: treating DOM geometry alone, missing return, data loss, CSS none or a synthetic trace as proof of native rendering or supplier integration. Regression tests must reject auto/contain/missing policy on either type of region and preserve ordinary horizontal overflow/vertical scrolling; original trace unavailable/lost/unattributed, blank-PNG, single-capture and cleanup tests remain. Full fixed cold `make verify` and independent QA are mandatory, even after local green.

Primary references: [CSS Overscroll specification §4](https://drafts.csswg.org/css-overscroll-1/#overscroll-behavior-properties) for none/contain and axis semantics; [fixed Chromium147 category definition](https://github.com/chromium/chromium/blob/147.0.7727.15/cc/tiles/frame_viewer_instrumentation.h), [snapshot tracing](https://github.com/chromium/chromium/blob/147.0.7727.15/cc/tiles/frame_viewer_instrumentation.cc). These are references, not copied product implementation or evidence of a real supplier.

## Scenario: observe the single screenshot's transient lifecycle (Issue32 execution33)

Execution32 establishes trusted wheel delivery190ms before capture-start195ms, with keyboard scrollend162ms and unchanged right-edge29+988=1017. The original1280×1149 PNG still has uniform RGB243/246/250 and no ink in all four anchors. Dispatch acknowledgment is therefore insufficient to resolve this failure; do not add another input wait or call it fixed. Before/after900px viewport metrics and visible CSS cannot rule out transient changes inside the screenshot operation.

Signatures: `observeCaptureLifecycle(page)` arms passive window/visualViewport resize and scroll plus document visibility observers; `finishCaptureLifecycle(handle)` reads the bounded result and disposes the handle. `captureIntegrityScreenshot(page,path)` marks the original screenshot call/return, retains at most16 monotonic samples plus an omission count, and cleans up in finally. Samples whitelist numerical viewport/scroll geometry, document visibility and body/content connected/checkVisibility booleans. No text, identifiers, credentials, raw CSS or error payloads are recorded. This diagnoses transient viewport conversion or hiding versus stable endpoints; event observations do not prove paint/compositor readiness.

Good/base/bad: one unchanged fullPage PNG followed by all original pixel/secret/geometry gates; observer unavailable remains explicit and cannot turn a failed PNG green. Blank PNG and original screenshot exceptions remain failures, with original error identity and listener/handle cleanup. No added timer, sleep, RAF, retry, recapture, input substitution, expanded timeout or dependency change. Required Node tests cover transient resize/hiding,16-sample bound, omitted count, observer unavailable, cleanup and single capture on success/blank/exception. These are driver contracts, not real Chromium evidence; fixed host `make verify` must provide the actual lifecycle receipt.

## Scenario: wheel dispatch must complete before the one capture (Issue32 execution31)

Scope: execution30's original keys trace has keyboard scrollend161ms, wheel-dispatch-start169ms, capture-start183ms, trusted page wheel189ms (inside the table, deltaX100000/deltaY0, not prevented), capture-end272ms. The table was already at33+988=1021 before dispatch. The old last-column predicate was therefore already true and did not establish wheel delivery. The uniform original PNG remains a failure; this ordering fact is not proof of its compositor cause.

Signature: `startWheelDispatchCompletion(region)` returns a handle with `completed:Promise<boolean>` and `cancel()`. Before the existing single `page.mouse.wheel(100000,0)`, arm a passive ownerDocument capture listener. Resolve in a microtask only after observing a trusted wheel inside this region with exactly that pixel delta; prevention observed at acknowledgment resolves false. This is not a claim that a microtask proves all bubbling/default actions have finished. Missing matching input fails at the existing2000ms bound. Timeout, cancel and observation remove the listener/timer; the driver disposes the handle on success and every failure.

Contract/error matrix: `readabilityEvidence(page,captureRight)` requires completion at `native-wheel-dispatch` before the original last-column wait/bounds assertions and one capture. Failed delivery is nonzero and retains the original bounded scroll trace; dispatch errors preserve their identity. No input is retried, no screenshot is retaken, no DOM scroll is substituted, no original assertion or deadline is relaxed. This acknowledges DOM delivery, not paint/compositor completion or a supplier response.

Good/base/bad: await native delivery even when keyboard has already reached the edge; then apply unchanged pixel/geometry/secret gates. Timeout or prevented/wrong-target input fails. Wrong: infer delivery from an already-true scrollLeft predicate or from awaiting the mouse command alone. Correct: arm the event-owned promise, dispatch once, require its result, then use the original exact geometric assertion and PNG.

Required tests: trusted versus synthetic/outside/wrong-delta events, prevention sampled at acknowledgment, timeout/cancel listener cleanup, already-at-edge deferred delivery preventing capture, failed delivery and original dispatch errors with no capture and one cleanup. These fixtures establish driver contracts only, not native event-loop/default-action behavior. The unchanged wildcard entry includes them in `make verify`; cold host receipts plus actual original PNG review are still required. If blank output persists after actual observation-before-capture, retain it and diagnose that narrower boundary rather than extend timeouts or declare the old cause fixed.

## Scenario: missing region wheel is not missing page input (Issue32 execution27)

Execution26 retained the complete execution25 trace: keyboard scrollend at170ms, wheel-dispatch-start178ms, capture-start193ms, capture-end282ms, no region wheel event, original PNG still uniform with the same historical digest. The observer was attached only to the table; its negative observation cannot distinguish input outside that table from no DOM wheel observation. Neither four layer-tree events nor ten paint events prove usable raster output.

Signature: `startScrollEvents(region)` obtains `element.ownerDocument` and attaches exactly one passive capture-phase `wheel` listener there; keyboard/own-scroll/scrollend remain on the region. `finishScrollEvents(trace)` retains at most32 events and3 landmarks and removes both listener scopes before disposing. The existing wheel record adds only `target_inside_region:boolean|null`, `client_x:number|null`, `client_y:number|null`; target nodes/IDs/text/paths are never serialized. No extra input, scroll, waits, screenshot, retry or acceptance assertion is added.

Good: observed page wheel outside the table is an explicit false ownership fact with finite coordinates. Base: zero observed wheel remains unknown, not proof of absence or a raster mechanism. Bad: weaken the mandatory pixel/geometry gate, infer browser behavior from the fixture, install dependencies manually, or force repaint. Required Node tests exercise both inside/outside targets, single capture listener, bounded projection, unavailable observations and both-scope cleanup; they prove the observation contract only. Current-source cold host `make verify` and independent QA remain mandatory; historical blank-image/R28-04 causality stays unresolved.

## Scenario: retain native input evidence when the original PNG fails (Issue32 execution25)

Host execution24 on candidate `6ce878eb3a41ec7fa0a18488c6625c09f9ef6c3d243ee66812036550ccf6a65b` observed three LayerTree changes and thirteen paint events, yet the original keys PNG remained uniformly RGB243/246/250. Paint events do not prove that the captured pixels contain content. The table was already at its horizontal maximum before wheel dispatch (35+988=1023). This does not establish an input/capture race: the old readability catch overwrote the already-finished input trace with `finishScrollEvents(null)`, losing evidence.

Signatures: `startScrollEvents(region)` returns a passive page-owned handle with `wheel()`, `capture(started)` and `finish()`. `finishScrollEvents(handle)` projects at most 32 events and three landmarks (`wheel-dispatch-start`, `capture-start`, `capture-end`) onto one `performance.now()` timeline, including finite local/root scroll offsets. `readabilityEvidence(page,captureRight)` keeps this observer active through the one original capture callback and finalizes/disposes once on success or failure. The catch retains that finalized receipt instead of replacing it with unavailable. Capture landmarks bracket the entire callback, not the internal CDP raster command.

Good: `if(scrollEvents) {scrollProbe.events=await finishScrollEvents(scrollEvents);scrollEvents=null;}` preserves the original capture exception and its passive trace. Base: unsupported/failed observers remain explicitly unavailable and cannot mask or pass the original pixel assertion. Bad: infer a browser mechanism from simulated events, wait for arbitrary frames, retry a screenshot, remove a wheel assertion or substitute another PNG. Required wildcard tests cover observation through the failing callback, one callback/finish/disposal, unavailable observer failure preserving the original exception, bounded numeric projection and secret exclusion. Real input delivery order still requires cold host `make verify`; absence of an observed wheel is not proof that Chromium never processed input.

## Scenario: uniform PNG needs render-boundary diagnosis (Issue32 execution21)

Host execution20 on candidate `566a5eb651efd268740b9d2f876b8e860015da062acc7e3fcd7bee63403f2e15` reached the keys table's last column (scrollLeft34/clientWidth988/scrollWidth1022). Its original fullPage PNG is uniformly RGB243/246/250; all four anchors have zero ink, while before/after geometry is stable. This differs from execution18's single clipped content anchor. Geometry and successful business calls cannot prove rendered pixels, CSS visibility or a browser-internal cause.

Signature: `captureRenderFacts(page)` observes document lifecycle/font/focus and five fixed node scopes (title, content/dialog, main, body, root), bounded to 16 ancestors each. Only whitelisted style enums, numeric opacity, hidden/inert/transformed/filtered/clipped and hit-ownership booleans are returned; no text, arbitrary selector, CSS value, URL or raw exception. Unavailable diagnostics are explicitly `available:false`, never a passing rendering claim. `capturePixelFacts(bytes,anchors)` also returns `surface.uniform_rgb` and three numeric first-pixel RGB channels from the original image.

Good: collect the white-list before/after the one original screenshot, then fail on the unchanged mandatory pixel/geometry assertions. Base: observer failure retains the original screenshot or assertion exception and its sidecar. Bad: waiting for frames, retrying screenshots, lowering ink thresholds, changing viewport/scroll/DOM or claiming these mock observations identify Chromium's mechanism. Required wildcard Node tests verify bounded output, secrets exclusion, ancestor visibility/hit flags, uniform/sidebar/complete PNG separation and unchanged one-shot failure. Real diagnosis still requires the original cold host `make verify` with original image/sidecar and independent QA; if another blank image occurs, compare the new observed boundary before choosing a repair.

## Scenario: capture anchors must be in the table's painted clip (Issue32 execution19)

Host execution18 on `ec88f5c79c35689c7587e83e50b0c3c68ba99728dc382a2fe196088e74957619` reached the services table's last column: scrollLeft67, clientWidth988, scrollWidth1055, last column right1246.71875 within region right1248. The original image shows the table and its actions. Geometry is stable and title/table header/data anchors have ink; the separate generic content anchor at left192,width110 has zero dark pixels because the table starts at258 and clips the header's text. This is an anchor-selection defect, not a failed wheel assertion or evidence of QA14's historical blank-image mechanism.

Signature: `captureBoundaryFacts(page)` keeps title + mandatory substantive content + each sampled table's header/data anchors. For generic candidates inside `.table-wrap`, the entire horizontal box must lie within the actual client clip (`tableBox.left + table.clientLeft` through `clipLeft + table.clientWidth`), as well as the existing document width bounds. No candidate leaves a null anchor and must fail; never silently drop it. Whitelisted receipts still contain only geometry, not text or secrets.

Good: select the next fully visible table header before the one original PNG capture. Base: normal non-table content is unchanged; native scroll still has its original keyboard/scrollend/wheel and 2000ms last-column assertions. Bad: sample clipped text under a sidebar, remove the failed anchor, lower pixel thresholds, move/reset the table before right-column capture, or retry the PNG. Required Node wildcard tests cover left/right/fully clipped headers, container borders, document Y coordinates, missing content, and continued rejection of blank/sidebar-only pixels. Fixture and offline original-PNG analysis do not prove live browser completion; full cold `make verify` and independent image review remain mandatory.

## Scenario: bill detail must finish opening before capture (Issue32 execution17)

Host execution16 failed at J06 after-correction on source `4a5046dee6b66429a75652f7ad8c5344a60dff1448a10c7293f9423bad1cf10c`. The original PNG shows the billing list, all six pixel anchors have ink, and the sidecar changes from `dialog:false` to `dialog:true` across capture. Production's async bill-row action awaits GET before calling `detail(...).showModal()`. A completed native click is not completion of this async handler. This explains this particular capture boundary, not QA14's historical uniform images or Chromium internals.

Signature: `billDetail(page,period,id)` in `tests/browser/journeys.mjs` arms the exact bill GET before one original row-button click, requires HTTP200 and matching bill ID, then waits for the actual `#detail` to be visible and asserts the complete immutable-bill title. Both J06 openings use it. It does not issue an extra management request or change any UI/API/DB data. The production header is synchronous before `showModal`; visible dialog therefore identifies rendered completion. Good: await `billDetail` then the original safety/pixel capture. Base: failed HTTP, wrong bill or wrong title fails. Bad: click then screenshot while GET is outstanding; retrying the image or relaxing geometry instead of synchronizing the native action.

Required tests in the unchanged Node wildcard gate: defer actual helper completion until visibility, match GET/path and one activation, preserve click/visibility errors, reject status/ID/title mismatch, and require both J06 call sites before the original capture. These fixtures prove driver contracts only. The original full `make verify` host run must still prove real timing, three viewports, visual/summary and self_hosted; no sleeps, retries, timeout increases or disabled integrity assertions.

## Scenario: screenshot integrity at the real capture boundary (Issue32 execution15)

`captureIntegrityScreenshot(page,path)` owns the existing fullPage PNG capture in console visual and continuous/self-hosted journeys. Existing password/secret/conversation/storage gates run first. It does not change focus, scroll, CSS, screenshot mode, input phases, deadlines or business data. Before and after the single screenshot, `captureBoundaryFacts(page)` collects only UTC, numeric window/visualViewport/root/body/main/content/dialog/title/table geometry and scroll positions, plus focus-kind booleans. Up to eight tables / eighteen pixel anchors; no DOM text, attribute values, credentials or raw errors. The receipt is `${path}.integrity.json`, beside the original unmodified PNG; the existing Go runner archives the complete test-results tree even on failure.

`capturePixelFacts(bytes,anchors)` decodes 8-bit non-interlaced Chromium RGB/RGBA PNGs, validates dimensions (at most40million pixels) and PNG scanline filters, and measures title + substantive content + visible first header/data-cell ink in each bounded table. `assertCaptureIntegrity(before,after,pixels)` requires viewport-width correspondence, existing in-image anchors, contrast>=32 and at least8 dark pixels per anchor, and stable document title/content geometry within1px across capture. Failure is fail-closed with original exception, digest/geometry/pixel counts/boundary retained; no image editing, retry or sleep. The pixel gate is necessary evidence integrity, **not** a substitute for independent human review of every state/action/column.

Good: stable geometry and actual main-content ink → inspect the real PNG in QA. Base: dialog captures use its own title/content, excluding the title from the second substantive anchor. Bad: nonuniform sidebar alone, uniform image, missing/displaced anchor or document movement → fail and retain sidecar; do not declare usable from DOM-green alone. `capture-integrity.test.mjs` covers normal/blank/sidebar/displacement/missing-body/geometry/whitelist/error-identity/one-capture contracts in the unchanged Node wildcard gate. Fixtures prove the gate contract only; execution15 separately decoded all181 original QA14 PNGs and matched the independent Go decoder's dimensions and four uniform images. QA14 has no live geometry, so that historical audit cannot establish root displacement, capture/composition cause or product severity. Next exact-source host `make verify` supplies the live boundary; preserve original keyboard-scrollend → one wheel →2000ms exact last-column assertions unchanged.

## Scenario: independent native scroll inputs (Issue32 execution12)

`startKeyboardScrollCompletion(region)` installs passive own-region `scroll`/`scrollend` listeners before ArrowRight and returns a handle with a `completed` promise and `cancel()` cleanup. Require movement followed by own-region scrollend within2000ms, then issue the original single wheel. Movement alone is not completion. Missing completion fails at `native-keyboard-scrollend`; no wheel fallback, sleep, input replacement, assertion retry or timeout extension. Success, timeout and cancellation remove listeners/timer; cleanup errors must not replace the original failure.

Host execution11 observed ArrowRight at1ms, wheel47ms at scrollLeft7, continued scrolling through146ms to40 and final scrollend163ms. Input reached the correct region and was not observed prevented; these facts prove overlapping input phases, not the Chromium internal cause or a product defect. Synchronize the driver at the native event boundary before retesting. Keep the exact rightmost-column bounds, screenshot secret gates and root-overflow checks unchanged. Good: `scroll -> scrollend -> wheel -> exact last-column assertion`; bad: `scrollLeft>0 -> wheel` while the prior input remains in flight, or script-setting scrollLeft to satisfy reachability.

`readability.test.mjs` verifies own-target filtering, movement prerequisite,2000ms failure, cancellation/disposal, wheel ordering and retained original rightmost error. Before the wheel probe exists, failure retains the original passive whitelist trace in `readabilityFailureFacts(error).keyboard_events`; missing scrollend must not discard the chronology. Fixtures prove driver contracts only; the locked-browser host `make verify` and independent QA remain required. API/DB/authorization/product scroll handling are untouched.

## Scenario: business terminology and full technical values (Issue32)

### Scope / signatures
`enumLabel(field,value,record={})` maps only role/source_kind/state display, with check state distinguished by execution_revision and invitation state by account. `technicalValue(value)` uses textContent on a focusable native region retaining the full public value. `column(label,field,format)` preserves explicit formatters; otherwise it applies the display mapping or complete technical text for address/model/ID/hint/account. `factSummary(value,fieldName,record)` applies the same context to read-only details. No API/DB/price/Token/auth changes.

### Contracts and cases
Good: check success → 核验通过; request success → 成功; pending invitation → 待接受; source self_hosted → 已有自托管. Unknown, including `__proto__`, stays `未知（original_value）` via own-property lookup, never inferred enabled/success. Exact URL/model/safe hint/ID stays complete, no ellipsis or secret. Error codes and monetary values stay raw. Base: existing Chinese boolean/key states and explicit money formatters unchanged. Bad: relabel API/export values, map every success to a check, truncate identifiers irretrievably or reveal a one-time secret.

### Layout and validation
Existing table-wrap is a named focusable native region; headers and action names stay unbroken, cells avoid arbitrary mid-word breaks. Complete technical-value regions scroll independently with visible focus. Details keep full values within their drawer, not at the root. `readabilityEvidence(page,captureRight)` checks actual computed styles/root dimensions and native ArrowRight/wheel reach of the last column, captures extra safe right-column images, then restores scroll; it waits for actual scroll completion, not sleep or assertion retries. Errors propagate, never skip; bounded failure facts name computed-layout/native-keyboard-scroll/rightmost-column and contain only booleans/counts.

When a native dialog is open, measurement and scroll activation belong to that dialog, not its inert background tables. Overflowing technical values additionally require actual keyboard scrolling; failure boundary is native-identifier-scroll. The driver-only unit regression verifies surface selection, not Chromium inert/focus/scroll mechanisms. Preserve host assertions and use their measured facts for acceptance.

### Required tests
`readability.test.mjs` covers context/unknown/prototype/raw-value preservation and negative layout gate predicates, not real browser behavior. Existing strict service summary expects translated check status derived independently from real API state. All ten `consoleVisualReview` pages collect readability facts at1280/1440/390; new self-hosted journey uses long URLs/models and actual safe Key metadata. Only full host make verify plus independent QA can close QA58-01; source/VM cannot.

### Accessible identity contract (Issue32 host execution5)
`technicalValue(value)` names its focusable region with exactly its complete public `textContent`; scroll guidance belongs in `aria-description`, not a replacement name. This keeps models, addresses and safe identifiers distinguishable to assistive technology and preserves the existing exact `getByRole('cell',{name:alias,exact:true})` row selectors. Wrong: every model region named only “完整值，可横向滚动查看”; correct: `aria-label=element.textContent`, `aria-description=guidance`. Empty values remain a non-focusable dash; null remains 未知; no secrets enter this helper. Unit tests enforce name/content equality for prefix-related model names, URL, safe hint and literal markup. The original model-save-readback row-count/status assertions and summary negatives stay intact; only real host execution can prove Chromium naming and the full save boundary. Execution5 measured zero exact rows before its first write, not a failed acknowledged-write readback; do not call this the root cause of R28-04.

### Native rightmost-column attribution (Issue32 host execution7)
Execution7 completed 1280px J01–J07 and management-summary, then timed out at overview `rightmost-column`. Its layout booleans passed and the keyboard scroll predicate returned, but it recorded no pointer-hit or scroll geometry. That receipt does not establish whether the table, input targeting, nested scroller or native dispatch failed.

`scrollBoundaryFacts(region,pointer)` performs read-only DOM measurement: region/viewport/last-column geometry, scroll/client widths, focus/connectivity and fixed booleans for point containment, hit containment and nearest horizontal scroll-owner identity. Return only explicitly allowlisted finite numbers/booleans (missing or invalid -> null); measurement failure -> `{available:false}`. No text, URLs, attributes, IDs, HTML, raw errors or host paths. `readabilityFailureFacts(error).scroll_probe` retains region index and before/after snapshots at the original pointer position, preserving the original error identity even if the second probe fails.

Good: diagnose an off-viewport or nested-scroll hit using host measurements before selecting a repair. Base: native ArrowRight, one wheel input, original 2000ms completion waits, exact last-column bounds and screenshot secret gates unchanged. Bad: infer native behavior from the diagnostic fixture, replace wheel with programmatic scrolling, add sleeps/retries, hide columns or widen the timeout. `readability.test.mjs` checks whitelist/finite values, unavailable diagnostics, original error and exactly one native input sequence; it proves diagnostic plumbing, not actual Chromium scroll behavior. The complete fixed host `make verify` remains mandatory; unchanged local listener restrictions do not explain this host failure.

### Native input timing attribution (Issue32 host execution9)
Execution9's original pointer was inside the viewport and the table; the hit's nearest horizontal owner was that table, not a technical-value or nested scroller. Its scrollLeft changed from1 before wheel to40 at timeout, with client399/scroll650. This rules out the previously proposed wrong-hit and nested-owner explanations for that run, but does not identify wheel delivery, cancellation or keyboard animation ordering. Do not call it a product defect or a listener restriction.

`startScrollEvents(region)` installs passive listeners on the original region before ArrowRight. `finishScrollEvents(handle)` removes listeners and disposes the handle; unavailable diagnostics remain explicit and cannot replace the original exception. The original ArrowRight, single wheel, waits and exact bounds assertion stay unchanged. Record at most32 fixed keydown/keyup(ArrowRight only), wheel, own-region scroll/scrollend observations with relative elapsed milliseconds, phase, scrollLeft, trusted/default-prevented booleans and finite wheel deltas/mode; overflow has an omitted count. No targets, keys, text, IDs, URLs, raw events/errors or host paths. The wheel's default-prevented flag is observed in a microtask, not proof of all later listeners' behavior. The observer never prevents events, sets scroll, waits for scrollend, sleeps or retries an assertion.

Good: use host event chronology together with geometry to select a source-backed repair. Base: diagnostic failure preserves the original gate failure; finish cleanup on success/error. Bad: infer browser animation from VM events, assume scrollLeft40 proves a universal Chromium rule, or relax the rightmost-column gate. Unit regressions prove only bounded/redacted passive observer plumbing, lifecycle and unchanged input/error sequence. Host `make verify` must supply the actual event trace before any further mechanism claim.

## Scenario: independent first self-hosted webpage journey

### Scope / signatures
`fullConsoleJourneyMode(width,mode)` creates a new private SQLite installation/server/controlled HTTP service per mode and viewport. Existing provider J01–J07 runs unchanged; self_hosted invokes `selfHostedJourney(admin)` in the same driver with separate receipts. Test-only `selfHostedUpstream` supports a long base path and closes `/unreachable/models` at the real TCP boundary; wrong credentials return401. It is not a real supplier or an injected product default.

### Contracts / error matrix
All source/model/price/grant/invitation/Key writes are native page actions; reads are authorized GETs for comparison. Empty starting counts must be zero. Connection failure → check failed/upstream_connection_failed/null HTTP status; wrong credential → failed/upstream_rejected/401; UI edit invalidates old applicability; correction → check success/applicable. Owner and assigned diagnostic developer use ordinary and actual SSE respectively, then page sameID logs/charge/price snapshots and aggregate usage must agree (3/4 tokens, fee21/cost7 each; total42/14). Keep original save ownership barrier and all denial/rotation/budget/cancel/error/UTC/migration/race gates. No SQL/API write seeding, delays or assertion retries.

### Evidence and required tests
Native branch receipts carry UTC, actual Chromium version, zero counts, fixed step names, scalar facts and safe screenshot relative names. They omit IDs/addresses/key values/bodies; screenshot capture first clears secret controls and asserts no credential text, no browser storage and no root overflow. Failures retain bounded state/code/status facts; parent make verify supplies source stamp, locked dependency/cold start/log/exit and actual cleanup facts. `TestContinuousUpstreamRelayHTTP` adds known SSE without removing known/unknown cases; `TestSelfHostedControlledHTTPBoundary` distinguishes TCP close from401. Tests require host listeners; unavailable sandbox is neither product failure nor pass. VM/fixture behavior never proves Chromium or real-provider integration.

## Scenario: operation-owned post-write page completion

Issue30 tests5 actually measured four UI toggles (false/true at1280/390) on source `ba8816e8f4cb38bb3b6c782195da2324bd16964f98cc2ff00448d91e78940d6f`: baseline save returned while both real post-PATCH reads were held; released final strict DOM assertions passed. This proves this execution's premature observation, not the root cause of old tests18 or a persistent final-state defect. Preserve that Run's failed receipt.

### Signatures and ownership
`edit(title,fields,submit)` reuses its existing idempotency operation UUID and scope epoch. The editor form carries `dataset.operation` and `dataset.readback` (`idle`, `pending`, `complete`, `failed`, `interrupted`); these are internal DOM lifecycle facts, not an HTTP/public refresh protocol. `renderPage()` returns true only after current-epoch `content.replaceChildren(section)`, false on stale/failed reads. `syncTenants()` checks its captured epoch after the collection settles before mutating selectors/state. No API payload, DB, CSRF, module or dependency changes.

### Contract and error matrix
- Non-secret acknowledged write: close editor, mark pending and replace old content with the existing loading view before tenants GET. Keep busy/save disabled through tenants → current-page reads → DOM commit; ignore replacement edit attempts while this lock is owned. Mark complete only on successful current-epoch commit; then release the lock in its epoch-owned finally.
- Tenants GET failure: replace old content with error and GET-only retry; mark failed, explain confirmed write versus failed reread. Catalog/render failure: existing error view, failed marker. Neither failure is an unknown mutation, a success, an automatic write retry or a later completion of the old save.
- Navigation/logout/scope change: invalidate pending marker and abort existing requests. Old sync/renderer/catch/finally cannot commit into, render for or unlock a new editor/scope. A new editor replaces its operation identity.
- Secret/invitation result: preserve the separate one-time-secret lifecycle, masking, cleanup and no-replay contracts; no ordinary page-read completion is claimed. Unknown writes retain their original operation readback/GET recovery logic.
- `waitForSaveReadback(page,expected)` waits only for the captured operation's terminal DOM lifecycle or scope/editor replacement, then asserts operation identity, scope/page identity and complete exactly once. Dialog hidden, HTTP200, old loading hidden or an identical notice alone are not save completion. Failed/interrupted/replaced operations reject, not silently succeed.

### Required tests and examples
`save-lifecycle.test.mjs` executes the actual editor, sync and renderer functions with explicit dependency checkpoints; `model-save-readback.test.mjs` covers terminal predicates. These prove contracts only. Original summary negatives still reject wrong status, missing columns and duplicate rows with original error identity.

Host `make verify` additionally holds each real same-role/scope tenants and catalog response, checks loading/pending/save-lock/noncompletion at both widths and directions, forwards unchanged responses and requires final strict unique row/column/status. Explicit injected GET503 cases require save rejection/no old rows and user GET-only recovery. A real held catalog snapshot spans UI switch to all-tenant scope, back to original scope and a second opposite UI save; release must not overwrite the second DOM/lock. Keep finite original request deadlines and ordered route cleanup, not sleeps, global network-idle, assertion retries or API/SQL writes. Receipts whitelist UTC/status/boolean facts and actual `browser.version()`, never identifiers or secrets.

Wrong: `await PATCH; await dialogHidden; await loadingHidden; assertRow()`. Correct: await the same operation's successful current-scope read/render commit, then assert the row once. A later user retry may restore visible facts but does not turn the already failed original save into a success.

## Scenario: pre-repair post-write model readback attribution

`tests/browser/model-save-readback.mjs` runs from the existing `journeys.mjs` immediately after real UI model/price/grant preparation. `readbackPhase(request,tenant,acknowledged)` excludes reads before the current PATCH200, non-GETs, tenant detail routes and catalogs without the written tenant and `include_ungranted=true`. One temporary authenticated page owns the write and both read barriers. No API/SQL mutation seeds the model status.

The diagnostic holds each real `route.fetch()` response, forwards it unchanged with `route.fulfill({response})`, first for the tenants collection, then for the catalog. Record PATCH/read/measurement/release/save/final UTC, direction, viewport, current-scope/page booleans, editor/loading/notice/lock booleans and `summaryFailureFacts`; never serialize IDs, URLs, bodies, headers, secrets or error text. Release and await each claimed handler before removing only its registered route. Request waits and actual forwarding retain the existing Playwright deadlines; no sleep, deadline expansion, retry or global network-idle wait.

This is a **pre-repair diagnostic**, not the final save driver: while tenants are held it deliberately awaits the baseline `save`, which closes on dialog-hidden/loading-hidden. An early return records the original strict stale/absent row assertion; all four UI toggles (false/true at the current width and 390px) finish real forwarding and strict final DOM checks before rethrowing that same first error. If baseline save does not return, retain the existing timeout and partial held-stage receipt; do not assume the candidate mechanism reproduced. On a later repair, replace this deliberate legacy-await with an operation-associated completion barrier based on the host evidence, rather than disabling the diagnostic or making the strict assertion retry until green.

Good: new host receipt distinguishes early observation from a final DOM error. Base: local matcher/summary tests validate only whitelist and assertion contracts. Bad: treat controlled locators as Chromium evidence, close historical R28-04 based on this different execution, or classify final wrong data as a harmless race. The fixed host `make verify` must run this slice and return the actual receipt before selecting a driver/product repair; full journeys, negative summaries, race, UTC, migration, visual and independent QA remain subsequent exit gates.

## Scenario: controlled route cleanup has one continuation owner

The interrupted-editor driver releases its barrier, awaits the held route's continuation, then removes only that handler with `page.unroute(pattern,handler)` in a nested finally. Never run continuation and unroute in `Promise.all`: locked Playwright 1.59.1 server `Page.removeRequestInterceptor` notifies in-flight routes, and `Route.removeHandler` automatically continues the current handler. Removal racing an explicit continuation can produce `Route is already handled!`. Do not catch/ignore this error or use `unrouteAll({behavior:'ignoreErrors'})` to turn a broken gate green.

Success receipts are written only after cleanup succeeds. On editor assertion or continuation failure, cleanup still removes the registered handler and the original gate stays red. `workspace.test.mjs` models asynchronous continuation settlement and requires unroute to follow settlement on both success and failed editor-enable assertions. `route-cleanup.test.mjs` executes locked Playwright's actual server Route methods with an isolated transport delegate: removal-first rejects late continue, ordered cleanup delegates exactly once. This is source-level mechanism evidence, not native Chromium/network cancellation validation; all three host `make verify` editor journeys and original assertions remain mandatory.

## Scenario: controlled tenant-readback interception includes collection queries

`collection('tenants')` sends `/admin/api/tenants?page=1&page_size=100&limit=100` and may add a cursor. A Playwright glob ending in `/tenants` does not match query-bearing URLs. The interrupted-editor driver uses `/\/admin\/api\/tenants(?:\?[^#]*)?$/`, holds only the first GET and continues POST and tenant member/detail requests. Never remove pagination or broaden the hold to all management requests to satisfy a test.

`workspace.test.mjs` executes the actual collection function, requires its emitted path to match, and checks bare/query/cursor, POST bypass, detail/unrelated rejection, cleanup, save enablement and cancel/no-extra-write assertions. Seq51's POST201 plus missing controlled checkpoint was a driver mismatch, not proof that the editor-lock repair failed. The locked Playwright matcher proves URL matching locally, not native cancellation or product delivery; only host `make verify` closes the Go/browser gate. Safe failure diagnostics expose `readback_checkpoint_reached` as a boolean, never URLs, bodies or credentials.

## Scenario: editor save-lock ownership after interrupted readback

`edit(title, fields, submit)` initializes both `editor-save.hidden=false` and `editor-save.disabled=false`. The new dialog owns its initial lock; it must not inherit the previous dialog's pending or superseded state. Submit still synchronously sets `state.busy` and disables Save. The old `finally` may unlock only when its captured epoch is current. Never remove that epoch guard to repair a disabled new dialog: a stale completion would then unlock a newer in-flight write.

Acknowledged write closes the dialog before `syncTenants()` and `renderPage()` complete. Navigation can invalidate that readback, preventing the old finally from running its unlock. Good: opening the next editor explicitly initializes its lock. Bad: depend on old readback settlement, add sleeps to the test or unlock every dialog from stale finally. Unknown writes retain GET-only recovery and no replay; no backend method or persistence contract changes.

`workspace.test.mjs` deterministically holds first acknowledged readback, advances the epoch, opens/submits the second editor, and proves old completion cannot unlock it. The driver `save` checks native enablement before arming one response waiter, awaits click and response together, and returns that same response. Native `console.mjs` holds only the temporary service-create tenant GET, navigates, requires the new editor to be enabled, cancels it and requires exactly one real service POST. All three original widths remain mandatory. `test-results/editor-interruption-{width}.json` is a controlled timing receipt, not proof of OS/network cancellation or supplier integration. Native execution remains host-only `make verify`; VM fixtures cannot close that gate.

## Scenario: summary ownership and same-scope context actions

### 1. Scope / Trigger
QA47's model/service/tenant/workbench summaries live in the production `workspace.js` renderer, backed by Go collection projections. Do not copy prototype fixtures or promote QA47 findings to Closed before independent QA.

### 2. Signatures
Model rows consume catalog `summary.sources` and `average_duration_ms`; platform catalog expansion remains scoped for latency. Tenant rows consume authorized `summary` or display a permission boundary. Service rows join actual global models' candidates with actual service/check revisions. `modelCall(alias)` transfers only a public model name after an epoch-owned navigation; `tenantDestination(tenant,page)` selects that tenant before navigating.

### 3. Contracts
Sources/types, per-candidate current USD input/output rates and measured/unknown latency are visible in multi-object rows. Connection-check timestamp/revision/applicability never substitutes for generation latency or a uniform service priority; priority belongs to each associated model candidate. Model enable toggle uses the existing expected-revision PATCH and retains all candidate identities/flags/priorities/weights. Overview uses scoped catalog state and direct model/budget entrances. Tenant owner/Key/budget summaries are not inferred from browser role labels. Request rows retain `Token总量` with explicit original-upstream explanation and separate confirmed input/output Token cells; detail/analysis/CSV use the corresponding immutable and ledger-derived fields.

### 4. Validation & Error Matrix
No sample/current price -> explicit unmeasured/unknown, not0. Restricted tenant summaries -> administrator-permission text, not blank false facts. Epoch superseded during model-call navigation -> do not populate the new page. Existing GET failures keep the renderer's failed-state/retry contract. No application credential is auto-filled by a model-row call action.

### 5. Good / Base / Bad Cases
Good: two models, two tenants and multiple real Keys show actual values and context actions at all three widths. Base: disabled model has real metadata, explicit status and explicit toggle; unsupported authority is retained. Bad: copying demo latencies, selecting only a table header, accepting either of two old/new tenant scopes or changing stored amounts for screenshots.

### 6. Tests Required
`console-summary.mjs` checks exact current-row cells by actual column ownership against authorized APIs and the objects created by original native J01–J07. Preserve multiple-object cardinality, fixed input/output rates, unmeasured states, context call/name/no-secret, disable/enable, association details and same-scope model/budget/Key entrances. `console-visual.mjs` retains all previous gates and adds the new required fields/actions; original63 images remain. J06 checks unknown/pending before reconciliation, original-null/confirmed3+4 after it, request detail, analysis and original+confirmed CSV. `console-summary.test.mjs` tests strict assertion helpers and money precision, not Chromium mechanisms. Existing VM service-check fixtures must supply the new models collection without changing their execution-revision/failed-check assertions.

### 7. Wrong vs Correct
Wrong: green VM tests or old seq46 browser results validate the new renderer. Correct: retain local failure and green receipts, fingerprint the candidate and require the configured host's unchanged full `make verify`, followed by independent frozen-material UI review.

## Scenario: visual editors must use the action's owning surface

### Scope / Signatures
Host seq44 settled limits facts pass title/scope/role/ready/fields/actions/secret/layout gates; `调整租户限额` is enabled, but capture times out at `open-editor` with no open dialog. Production `mainAction` binds `#page-action` in `.page-heading`, outside `#content`. `visualEditorControl(page,definition,row)` in `tests/browser/console-visual.mjs` uses `editor.location:'heading'` only for limits; other editors retain the exact current-object row or content owner.

### Contracts / Error Matrix
Resolve the exact button name within the declared owner, assert count1, then native click and require `#editor[open]`. Missing/duplicate action fails; never search the whole page or switch to another row/content action. Keep all existing editor field, page capture, nonempty, scope, role, layout and secret assertions. This is a driver scope fix, not permission or API behavior.

### Good / Base / Bad / Tests
Good: limits uses the heading action; Key/model/service/billing use the exact live row and invitation uses content. Base: absent action fails before capture. Bad: `#content` lookup for a heading action, global `.first()` fallback, moving production buttons solely to satisfy a driver or dropping the editor check. `console-visual.test.mjs` checks the actual helper's scoped locator contract and production HTML boundary, including no fallback; fixtures are not native browser evidence. `make verify` must still run the original real service, J01–J07 and ten-page review at1280/1440/390 before independent QA can close QA38-01.

## Scenario: nonempty visual review after month-end billing

### 1. Scope / Trigger
Host seq42 proves the first visual page has a settled scoped overview GET200, all required fields/actions, and exactly the legitimate `当前范围近7日暂无调用` empty state. J06 advances to the next month to test closing/correcting the old bill. The later nonempty capture cannot use historical activity as current-period data. Fix the owning journey's preparation, not production rendering or `assertCaptureFacts.ready`.

### 2. Signatures
`prepareVisualActivity(page,key,model,tenantID,billID)` in `tests/browser/journeys.mjs` runs after J06 and before `consoleVisualReview`. It uses existing native `call`, `requestDetail`, `closeDetail` and `nav` helpers. Only overview and closed-bill GETs are direct evidence reads; no API or database business-object seeding.

### 3. Contracts
Confirm current UTC day/month with zero current-day calls and no seven-day trend; perform one real native normal call through the controlled upstream. Require one dispatch (owned by `call`), current-period nonpending detail, one charge with cost7/fee21 at J06's published rates, and authoritative overview with one successful call and one current-day trend row. The full closed-bill response including original amounts and corrections must be unchanged. Navigate away from the call surface before screenshots so secret/output lifecycle clearing still applies. Write only numeric activity facts and a fixed conservation boolean to `test-results/visual-activity-<width>.json`.

### 4. Validation / Error Matrix
Wrong period/day, missing activity, unknown usage, incorrect fee/cost or changed closed bill → fail before visual capture. Loading/errors/any `.empty`, missing fields/rows, wrong role/scope, overflow and visible secrets → retain original capture failures. Actual UI/API/dispatch failure → original stage/receipt; never substitute a fabricated series or reset the clock backward.

### 5. Good / Base / Bad Cases
Good: a native request creates real new-period facts without altering an ended bill. Base: the pre-call legitimate empty trend is asserted, not presented as a populated screenshot. Bad: accepting any empty state as nonempty, backdating records, restoring the old month or weakening production permissions/charging to suit a capture.

### 6. Tests Required
`journeys.test.mjs` executes the actual preparation function with explicit contract fixtures, rejecting missing/stale activity and changed bills; these are helper-contract tests, not proof of native mechanisms. The unchanged `make verify` must execute the actual service and all J01–J07/nonempty ten-page/keyboard/screenshot checks at1280/1440/390. Independent QA reviews those same-version receipts; local helper green does not close QA38-01.

### 7. Wrong / Correct
Wrong: remove `.empty` from `ready` because GET200 returned. Correct: retain `ready`, prove the empty-state prerequisite, generate one real current-period call, verify facts and conservation, then capture.

## Scenario: failed visual capture must retain the assertion boundary

Before `assertCaptureFacts(facts)`, persist `pending_capture` in the existing visual receipt. Keep exact title/scope/role, expected-field/action booleans, row/card counts, root/content dimensions, loading/error-node counts and classified visible empty-state counts. Only classify known labels (`当前范围近7日暂无调用`, `当前范围暂无数据`); unknown DOM text is a count, never a text dump. Record `failure.check` and the fixed capture assertion message, not arbitrary browser errors. Filter actual response events with `overviewReadFact(response,tenant)` to exact GET `/admin/api/workspace/overview`; retain only resource/status/scope-match, at most16 reads, never URL/query/body/account/credentials. Remove the listener even on failure. A prior J06 expected404 is not overview HTTP evidence.

Good: host failure has `pending_capture.facts.ready=false`, classified empties, actual overview GET status and the named assertion. Base: navigation timeout retains a navigation boundary and null facts, not invented DOM facts. Bad: broad DOM/response dump or silently permitting `.empty` based on an unproved hypothesis. `console-visual.test.mjs` proves receipt ordering, strict gate preservation, whitelist and redaction only; its mock facts do not establish the host cause. Keep the original assertion until real host evidence distinguishes legitimate period emptiness from error/loading/role/scope/fields/layout defects. Run the unchanged `make verify` via host, then make only the evidenced repair; QA38-01 stays open until current-fingerprint host captures and independent comparison.

## Scenario: ten-page real-console design evidence

`tests/browser/console.mjs` retains its original heading/root-width smoke loop and native keyboard gate. `journeys.mjs` calls `consoleVisualReview(page,{width,tenant,objects,secrets,nav,step})` from `console-visual.mjs` after the original J01–J07 assertions and before writing the final continuous-success receipt. Objects come from the existing UI-created two tenants/two public models/service/application Keys and reconciled request, not HTTP-seeded completion states. At1280/1440/390 assert exact page title, authenticated role, selected tenant, settled nonempty content, substantive table headers/summary/control labels and actions. Click actual detail/editor controls, wait for the open native dialog and required fields before screenshots; close without writing. Record page, role/scope, assertion booleans, safe fields, image hashes and repository-relative paths in `test-results/visual-{width}/review.json`. Partial/failure receipts retain the last safe phase and previous captures.

Screenshots require cleared passwords/Key/invitation/message controls and absent call output; reject known test secrets in rendered text. A normal untouched call form is applicable empty state, not a failed data list. Wrong role/scope, read failure, missing field/action or root overflow fails the original host entry. Contract/VM tests only validate assertions and production rendering; they do not close AC01. QA independently compares actual host PNGs with the frozen prototype and records deviations. Never replace this gate with static `browser.check`, title-only navigation or synthesized screenshots.

Production Key budget cards reuse `limitSummary` (tenant and Key share the five independent dimensions, exact micro-USD and unknown/reserved explanation). `factSummary` labels the actual `attempts` API field and four price-rate fields; no payload, amount, auth or ledger semantics change. Required regression: `workspace.test.mjs` executes the actual render branch and validates both resource-summary calls and exact price strings. The browser visual suite asserts the corresponding visible detail fields. Good: open price versions with actual cost/tenant rates; base: editor inspection without POST; bad: raw Key JSON as the primary budget presentation or claiming all-page QA from supplemental VM green.

## Scenario: unknown writes retain a safe, repeatable recovery control

### Scope and signatures
`edit(title, fields, submit)` owns the operation UUID, synchronous busy guard and unknown-result latch. `submit(values, operation)` performs at most one unresolved mutation. Recovery calls only `api('operations/'+encodeURIComponent(operation))`; it never retries POST to recover a secret.

### Contracts and error matrix
On a lost acknowledgement (including an abort without a scope change), hide Save and latch the form against additional submission. Display the safe operation reference via `textContent`. Append a dedicated recovery-status node and a separate native `安全读回` button. A transient GET503/404 keeps that button and the reference; update only the status node, never the parent's `textContent`. While GET is pending, disable the recovery button. Success opens only safe operation metadata, not a replacement credential. The captured epoch **and** the original dialog's open state must match before painting either success or failure. Management401 still uses the existing `api` signed-out invalidation, rather than keeping a supposedly authenticated old editor alive.

### Good / base / bad
Good: one committed rotation, lost POST response, GET503 then GET404, same-operation GET200; original application, usage and history remain unchanged. Base: ordinary acknowledged writes retain the previous path. Bad: clearing the recovery parent on failure, replaying POST, showing a late detail after editor close, or calling Contract tests Chromium evidence.

### Required checks
`tests/browser/readback.test.mjs` extracts the production editor for retry/reference, unknown-submit latch, concurrent read deduplication and late-scope/closed-dialog checks. `console-faults.mjs`, imported by the original `console.mjs`, runs actual Go + Chromium recovery for issuance and an already-used application's rotation at1280/1440/390, including native double-click/Enter, stacked-transition409, preserved price/ledger/usage, and ordinary/SSE double-send with real attempt/admission counts. Its historical red is explicitly a test-only reconstruction of the previous editor served by a page route, records its SHA256, and is **not** current production acceptance or a claim about original Chromium execution. Current-asset recovery must separately pass. No static browser tool or VM fixture closes this gate. Partial safe JSON and secret-free screenshots are retained in ignored `test-results/fault-*`; host/QA archive actual files under this Run.

### Wrong vs correct
Wrong: `output.textContent='读回失败…'` deletes both reference and retry. Correct: `status.textContent='读回失败…'` changes only the dedicated sibling; `unknown=true` blocks further form mutation while recovery remains available. These rules avoid accidentally issuing another application credential after a committed-but-unacknowledged write.

## Scenario: reconciliation reads an append-only ledger

### Scope and signatures
`assertReconciledLedger(before, after)` in `tests/browser/journeys.mjs` verifies J06 after the native reconciliation dialog POST and authoritative detail GET. `workspace/requests/{id}` returns `charge_entries` ordered by ordinal/revision, not a replaced total row. `settleAdmission` initially records revision0 with zero confirmed amounts for wholly unknown usage; reconciliation appends the difference at the new admission revision (`before.revision+1`, not necessarily1).

### Contracts and validation
Keep all prior ledger items and price snapshots identical. For the one-attempt controlled 3-input/4-output fixture at cost1000000/tenant3000000 microUSD per million, require cumulative cost7/fee21, confirmed Tokens3/4, one new entry, pending0, reserved0 and matching after/delta revision, ordinal, reason and evidence. Invalid money/Token representations, altered historical rows, duplicate delta, wrong sums, retained pending/reserve or changed snapshots must fail. `ledger_facts` exposes only bounded numeric totals/counts, never request identity, reason, evidence or raw assertions.

### Good / base / bad and tests
Good: sum all append-only entries using BigInt and preserve the first zero-confirmed row. Base: existing three-attempt Go governance scenario confirms unknown ordinal2 revision0 remains unchanged and reconciliation revision2 adds fee5/cost16 before bill totals19/60. Bad: `charge_entries[0].fee === '21'`, modifying the product to replace prior entries, or weakening to nonempty ledger. Actual driver helper regression includes wrong fee/cost/Token/revision, duplicate delta, mutated history, pending/reserve and snapshot negatives; both GovernanceInProcess/ActualHTTP share production handler/SQLite assertions. All join original `make verify`; local helper/in-process passes are not Chromium or actual-network evidence. Preserve seq32 failure and require fresh complete host verification at1280/1440/390.

## Scenario: member removal refresh settlement and failure ownership

### 1. Scope / Trigger
Host seq30 retained a `J05-resume` label after subsequent member operations. All observed pause/resume fee, reservation and count pairs matched; `action:authoritative-read` cannot locate the failing assertion. Do not report this as a proven fee-conservation failure. Source inspection additionally shows `refresh.onclick` awaits `syncTenants()` before `navigate()` creates the loading placeholder, so old content already satisfies `ready()` while tenant synchronization is pending.

### 2. Signatures
Test driver `refresh(page, expectedTenantCount)` waits for the declared selector outcome using native `page.waitForFunction(predicate, count, {timeout:5000})`, then the existing ready check. J05 member removal supplies zero and retains its exact zero-option assertion. No product endpoint, permission, fee rule or UI state is changed. `failureBoundary(error)` emits only the version-bound `tests/browser/journeys.mjs` line/column, never stack text or host directories.

### 3. Contracts
J05 separately attributes member lookup, role write/readback, diagnostic recovery, removal, removed Key scope, membership list, UI scope, application survival and other-tenant Request ID denial. `read_facts` contains only expected/actual status. `membership_facts` contains an allowlisted role and counts. Existing usage facts remain exact decimal strings/counters; they describe the last observed conservation boundary, not every subsequent action.

### 4. Validation / Error Matrix
Pending tenant synchronization with old content -> refresh with expected scope must remain pending. Zero options observed -> existing strict zero assertion runs. Failure to settle within five seconds -> timeout, not acceptance. Wrong role, retained diagnostic membership, non-404 removed scope or cross-tenant read -> retain original failure. Missing own-source stack frame -> `available:false`; arbitrary exception text is never printed.

### 5. Good / Base / Bad Cases
Good: bounded predicate waits for the exact expected removal fact, then asserts it. Base: unchanged ordinary refresh has no extra scope expectation. Bad: sleeping, retrying mutations, accepting either selector count, or changing production billing from a stale label.

### 6. Tests Required
`journeys.test.mjs` executes the actual refresh helper with delayed tenant settlement; pre-fix returns early (red), post-fix waits (green). This proves helper ordering only, not Chromium's seq30 failure location. Position-projection tests reject unrelated stacks and secrets. `TestConsolePeopleScopeInProcess`/`ActualHTTP` additionally check live-session finance role, diagnostic cleanup, reassign-before-delete cascade and an empty post-removal membership list. Full unchanged `make verify` with native J01–J07 at all three widths is still required; preserve original host failures and independent QA No-Go.

### 7. Wrong / Correct
Wrong: `click refresh; ready(); count()` assumes an existing non-loading page means the new scope settled. Correct: declare zero expected options, wait boundedly for that new fact and keep the original count/status/dispatch assertions. Do not claim this source-backed candidate is the native root cause until the host receipt confirms the boundary.

## Scenario: J05 pause/resume conservation diagnosis

The continuous native driver must preserve the original resumed `usage.fee_micro_usd` equality and all J01–J07 actions. `assertUsageConserved(before,after)` supplements it after primary ordinary rejection, primary stream rejection, standby rejection, another tenant's successful call and resume. Compare period, exact string fee/reservation and safe nonnegative calls/pending/concurrency. Never coerce micro-USD to a JS Number or normalize missing fields to zero.

Failure `usage_facts` contains only `period_same` and the before/after amounts and counters for these five fields; invalid values become null and fail, not pass. No tenant/Key IDs, accounts, arbitrary object bodies or request data enter this projection. Good: same nonzero fee/calls survive pause while the other tenant is charged independently. Base: unknown usage reservations remain held. Bad: reduced/reset fee, extra admission, missing metadata or period drift fails the original layer. `journeys.test.mjs` executes the actual projection/assertion and proves exact large amounts, negative comparisons and redaction, not browser/network mechanisms.

`TestConsolePauseUsageInProcess` uses production management/proxy/SQLite with an explicitly labeled transport fixture; `TestConsolePauseUsageActualHTTP` repeats the same contract using temporary HTTP listeners. They require nonzero starting fees, two application Keys, zero dispatch for paused ordinary/SSE calls, independent tenant charge, identical usage before resume and appended usage after resumed calls. Both remain additive in `make verify`. Passing the in-process variant cannot resolve a conflicting native receipt; retain the native failure, run the bounded diagnostic via the configured host and use its evidence before choosing a product fix.

## Scenario: new installation is not a legacy migration

`InitConsole(directory,keyFile)` targets schema3 and removes the transitional `legacy` tenant after creating the schema. `GET /admin/api/upstreams` and `/models` return `data:[]`; `/tenants` returns `items:[]`. The platform administrator's `syncTenants()` still shows one `全部租户` option with empty value; that is a scope control, not an existing tenant. Migration journeys separately retain historical tenant/Key conservation assertions.

`freshInstallation(page)` in `tests/browser/journeys.mjs` must read all three actual collections, require non-null arrays of length zero and the single empty-valued platform scope option before creating the first tenant through UI. Bad: expecting `legacy` on an isolated new installation or adding database records to satisfy that expectation. Good: empty new console; base: migrated console tested in its separate suite; bad input: unexpected existing tenant, null collection or missing scope option must fail.

`TestConsoleFreshInstallationEmptyCollections` verifies real InitConsole/Open, login and actual handlers/SQLite with no listening port, including legacy Key-list 404. `journeys.test.mjs` executes the actual driver segment against empty/nonempty/malformed inputs; neither proves Chromium or network behavior. Native host `make verify` must still pass all J01–J07 at 1280/1440/390. Failed reads record numeric HTTP status; initial stages identify the collection and report numeric counts only, never response bodies or account/secret values. Preserve historical failed receipts rather than closing QA on these supplemental tests.

## Scenario: continuous production journeys and accurate Key states

### 1. Scope / Trigger
QA22-01 found that ten-page navigation and platform-issued calls did not prove frozen J01–J07. QA22-02 found the Key formatter ignored enablement and expiry. Changes belong to the production workspace and the additive browser runner, not schema2 compatibility, authorization fixtures or the frozen material.

### 2. Signatures
`keyStatus(key, now=Date.now())` displays revoked, disabled, expired, active grace, enabled or unknown using safe metadata; it does not authorize a call. `needTenant()` replaces the loading placeholder with the actual no-scope state. Request details offer `在当前范围查看日志`, preserving tenant and existing safe filters while adding an explicit request filter.

`make verify` → existing `make browser` → existing compatibility/native journeys → `fullConsoleJourneys()` → three independent `fullConsoleJourney(width)` installations → `tests/browser/journeys.mjs`. Original gates remain. New test-only environment names are `RELAY_JOURNEY_WIDTH`, `ORIGIN`, `PASSWORD`, `UPSTREAM`, `CONTROL`, `CAPABILITY` with the common `RELAY_JOURNEY_` prefix. Never print secret values.

### 3. Contracts
Key table consumes `enabled`, `revoked`, `expires_at`, `generations[{expires,retired}]`, explicit models and the real `/keys/{id}/limits` fee/reservation/budget. Revocation precedes disablement; enabled expiry is inclusive at the deadline; only unretired, future grace displays transition. Missing state metadata or malformed global expiry is unknown, never guessed active. Authentication remains Go-owned.

Each continuous runner calls production `InitConsole`/`Open`, creates ephemeral loopback servers and deletes its isolated directory afterward. Empty Start State contains only the init-owned, asset-free legacy tenant; no user tenant, endpoint, model, Key, bill or request is seeded. All business mutations use UI; authenticated API reads only check final facts. A separately capability-protected test server exposes `GET /facts` and `POST /advance {seconds:int64}` for monotonic UTC time and dispatch counts. No product control endpoint or SQL mutation is added. Initial test time is the first day of the current UTC month at noon, so the 60-second expiry step cannot accidentally reset monthly usage on a real month-end run. Clock advancement affects the existing Store clock and browser Date; personnel sessions retain their actual wall-clock contract.

### 4. Validation / Error Matrix
- Revoked/disabled/expired/transition/enabled/unknown metadata → distinct informational list labels; never client permission enforcement.
- No tenant → empty guidance replaces loading; a removed member cannot retain old diagnostic DOM.
- Missing test capability → 403; negative or >35-day advance → 400; other test-control routes → 404.
- Expected rejection → exact status and zero upstream POST delta; unexpected status, unavailable scope or failed browser assertion fails the original entry.
- Failed UI read → visible failure/retry, not empty success. Held old-scope real response is released only after scope switch and cannot replace the new tenant.

### 5. Good / Base / Bad Cases
Good: invite and authenticate a real tenant administrator before issuance, reject then recover authorization and hard money, preserve shared rotation usage, pause two Keys while another tenant works, reconcile and close an ended bill, append correction once in its adjustment period. Base: test-only controlled HTTP proves relay semantics, not supplier billing facts. Bad: API-create intermediate business objects, change DB amounts/dates, replace old gates, or mark newly written scripts as browser Pass.

### 6. Tests Required
`journeys.test.mjs` executes the actual formatter at exact expiry/grace boundaries and the actual no-scope renderer; checks driver/product label and rotation-status contracts plus additive gate wiring. These are supplemental Node contracts, not native evidence. `journeys.mjs` must execute J01–J07 at 1280/1440/390 against real Go/SQLite, including native keyboard invalid-field recovery and submission, UI-issued >25 records, delayed reads, same-ID fees, all Key states and before/after secret-free screenshots. Preserve original Go/Node/HTTP/SSE/cancel/unknown-issuance/keyboard/migration/history gates.

### 7. Wrong / Correct
Wrong: reuse seq21's old fingerprint green result to close QA22 or infer that VM fixtures prove Chromium/network behavior. Correct: generate the new source fingerprint, obtain full host `make verify` and independent QA; keep the original No-Go and failed attempts. Emit only fixed stage/action/error categories, numeric status, checkpoint IDs and DOM booleans/counts. Files under ignored `test-results/continuous-{width}-{pid}.json` and associated screenshots carry actual execution evidence; do not invent them before host execution.

## 1. Scope / Trigger
Go embeds `web/workspace.html`, `workspace.css` and `workspace.js` for schema3. Preserve schema2 `index.html/app.js/style.css` and their trusted regressions; they are not the current production screen. No client framework or Node backend.

## 2. Signatures
`api(path,method,body,operation)` checks personnel/epoch ownership and handles 401. `collection(path)` follows server page/cursor metadata without silently cutting off the last page. `invalidate()` aborts reads/calls and clears secrets on navigation/scope/logout/pagehide. `edit(title,fields,submit)` owns a synchronous busy state and one operation ID. `runAction` catches async action errors for the current epoch. `limitSummary` displays independent units, used/held/remaining and platform upper bounds.

## 3. Contracts
`detail` consumes multiple shapes, not just a resource-limit report. `limitSummary` must dispatch to structured `factSummary` when either limits or usage is absent; services, Keys, requests, operation readback and bills must not throw for missing quota fields. Use textContent for scalar values, never HTML from an API. Preserve the same native dialog ownership, focus and secret cleanup across all summaries.

Table search is local to the fetched safe collection (or current log page); label it as current-list search, not a whole-database query. Search before slicing groups and reset to the first filtered group. The server-side log/usage filters remain authoritative for tenant/model/Key/time scope. Token analytics display confirmed ledger input/output separately from pending usage; a successful response with unknown usage is not a zero-token fact.
Ten fixed A pages use the real server identity and tenant selector, not a demo role switch. Lists, prices, risk and trends derive from Go APIs. Source kind is explicit supplier/self-hosted/legacy-unspecified; capabilities describe the relay protocol, not an invented supplier certification. Model details show cost versions and current/stale operator caps. Money uses BigInt and exact decimal strings.

Secret dialogs only hold the received issuance value in memory. Closing, navigation, tenant switch and logout erase it and conversation fields. A committed write with a lost response is unknown: hide the save action and offer operation-only safe readback, not blind retry. Browser route fault injection must commit the real backend write before aborting the response. The proof cannot be a fake UI fixture.

`dialog.close()` queues a native close event. When the same dialog is immediately reopened for a one-time value, the old close event must not erase the newly opened dialog. Only a currently closed dialog's handler cleans its contents. Busy Escape is prevented; native Tab/focus trap/Escape return are tested in a real browser.

Editor close cleanup is scoped to `editor-fields`: erase its password/secret/sensitive controls and detach the fields, error and submit handler. Do not invoke global `clearSensitive()` from that queued event: the explicitly transferred `call-key` belongs to the destination page, not the closed editor. Global navigation/logout/scope/pagehide invalidation still clears all sensitive controls. The explicit transfer captures the post-navigation epoch and writes only if that epoch and playground destination remain current; an intervening navigation, even to another playground, must not restore the captured secret.

`workspace.test.mjs` exercises actual close and transfer callbacks with close-before/after-settlement schedules, reopened-dialog protection and superseded-epoch rejection. These are deterministic ownership tests, not proof of Chromium scheduling. Native `console.mjs` requires a nonempty transferred control and boolean equality to the UI-issued value at `key-secret-transfer-value` on each viewport, then retains HTTP200 and same-Request-ID detail assertions. Never fill the call Key from a test API response, add arbitrary sleeps or remove the original gate. Failure snapshots remain boolean-only. Host seq17 proves issuance201 and empty call input for the old candidate; the revised candidate still needs full host verification.

Refreshing tenants updates both the array and select; creation must not leave a stale selector. Version reads capture the post-navigation epoch. Failed identity/permission reads clear sensitive fields; failures are not empty datasets. SSE accepts LF or CRLF, bounds local frames/output, displays chunks incrementally and treats missing DONE/usage as interruption/pending rather than fabricated success. A local cancel says upstream fees may remain unknown.

## 4. Validation / Error Matrix

### Keyboard failure attribution
`console.mjs` delegates only the existing create-dialog/13 Tab/Escape/return assertions to `keyboardJourney(page,width)`. `keyboardSnapshot(page)` is a read-only DOM projection: dialog-open, focus containment, body/root focus and document-focus booleans. A bounded 16-sample trace carries only fixed phase enums, Tab counts and these booleans; never emit active element IDs, labels, values, HTML or browser errors. Keep the original first-Tab and final-cycle containment assertions, immediate Escape visibility assertion and exact trigger-return assertion unchanged. Do not insert focus assignment, sleeps, broad retry or an easier assertion into the driver.

Host seq19 passed beyond issuance/explicit transfer/call/ledger/quota and failed with an open editor at coarse `keyboard-1280`. That alone does not distinguish first Tab, cycle boundary or Escape. Source inspection shows `keyEditor` launches `Promise.all(...).then(edit)` and returns before the dialog opens; `page.keyboard.press` does not target or await that dialog. The driver must await `#editor` visibility at `keyboard-dialog-ready` before sending the original first Tab, not force focus or sleep. Use `keyboard-dialog`, `keyboard-dialog-ready`, `keyboard-first-tab`, `keyboard-tab-cycle-N`, `keyboard-escape`, `keyboard-return-focus` plus the safe trace to choose any further product/driver/timing fix. `workspace.test.mjs` executes the actual driver with success, delayed opening and four rejecting boundary fixtures. The delayed fixture is red without readiness and green with it, proving a driver readiness contract, not Chromium scheduling or the unique cause of seq19. No product focus repair is claimed; fresh host `make verify` remains required.

| Event | Result |
|---|---|
| Late response from old tenant/navigation | Abort/suppress; never paint current scope |
| 401 | Clear personnel/scope/content/secrets; show login |
| Permission failure | Clear secrets, show actual failure; refresh obtains current role |
| Duplicate submit | One operation/call only |
| Committed write, lost acknowledgement | Unknown + safe readback; no new secret replay |
| Readback fails | Still unknown; not success |
| Secret dialog closes or page leaves | Erased, no browser storage |
| Canceled SSE | Local stop, same Request ID, potentially pending fee |

## 5. Good / Base / Bad Cases
Good: real invitation acceptance/login, tenant-scoped issuance, ordinary/SSE/cancel → same-ID diagnostics. Base: no supplier/price results means explicit unknown/empty, with real configuration actions. Bad: prototype amounts, client role switching, truncating a 101-item collection to 100, or unhandled async rejection after an action.

## 6. Tests Required
`workspace.test.mjs` extracts the actual fact/limit summary and table functions for supplemental VM regressions: non-quota detail objects, null/unknown, quota dimensions, filtering beyond the first group and clearing the filter. Source boundaries must include the full function but not leave the following async prefix; fixture observations must read the same textContent the product writes. `frontend-test` includes original ui.test.mjs and the new *.test.mjs suite without replacing native browser gates.
`tests/browser/console.go` starts isolated real Go and explicitly controlled HTTP/SSE upstreams. `console.mjs` covers 1280/1440/390, ten pages, invitation/login/logout, ordinary/SSE/cancel, immediate rotation, quota rejection, commit-before-abort issuance recovery, native dialog keyboard/secret cleanup and secret-free screenshots. Existing journeys and VM tests remain in `make verify`. The static-only platform browser tool cannot validate this backend UI.

## 7. Wrong / Correct
Migration console drivers must consume the production `replyCollection` envelope `{items,total,page,page_size}`, not a schema2 or protocol `data` envelope. `migration-contract.mjs` requires both migrated Key identities and their enabled flags, plus exact count/total; `migration-contract.test.mjs` rejects the obsolete envelope, missing/extra identities and changed flags. `TestConsoleKeyCollectionContract` exercises the real login/create/list Go handlers and SQLite with a recorder, not a listening browser. These supplemental tests cannot certify native journeys.

`migration-console.mjs` splits failures into keys-navigation/keys-http/keys-contract/history per viewport. Its failure receipt permits only suite/stage/result, numeric HTTP statuses, a validated integer schema and numeric item count. Reset these facts at each viewport; never print raw errors, response bodies, arbitrary fields or secrets. Keep native row visibility, history and overflow assertions intact and obtain a new host `make verify` receipt.

Wrong: unconditional queued close cleanup clears the newly issued value; checking a checkbox with an invented accessible name; treating static rendering as backend evidence. Correct: guard closed-state cleanup, derive exact labels from actual HTML/JS, register bounded real response waits and inspect the host receipt before claiming browser acceptance.

## Scenario: schema3 service connection check

### 1. Scope / Trigger
Host seq11 stopped at service-1280. The actual service row action sent `{}`; the existing Go handler rejects a missing execution revision before dispatch. Fix the caller contract, not the server guard. The old coarse assertion is not proof of the sole native failure cause.

### 2. Signatures
`renderPage()` service-row action calls `api('upstreams/'+row.id+'/check','POST',{expected_execution_revision:row.execution_revision})`. `Server.checkUpstream` in configuration.go owns revision validation and connection-only GET `/models`.

### 3. Contracts
Use the safe listed `execution_revision`, not the management `revision`. Name-only metadata revisions and execution revisions can differ. A 200 check response is a completed observation, not necessarily success: require state/applicability/current revision for a successful controlled-native assertion. Connection success does not certify generation or a real supplier. Controlled model-list fixtures must include `object:"list"` and a non-null `data` array; never loosen protocol validation to fit a fixture.

### 4. Validation / Error Matrix
Missing expected execution revision → 400/no dispatch. Mismatched revision → 409/no dispatch, surfaced notice/no automatic retry. Valid current revision and protocol list → 200/success/applicable. HTTP200 with malformed list → 200/failed/upstream_protocol_error, not certified success.

### 5. Good / Base / Bad Cases
Good: row revision19/execution4 sends `{expected_execution_revision:4}`. Base: rejected stale row exposes failure and requires operator refresh. Bad: `{}`, a management revision, or auto-refresh-and-recheck against an unreviewed new configuration.

### 6. Tests Required
`workspace.test.mjs` executes the actual render function and service action for exact payload/success and conflict/no retry. `TestConsoleServiceCheckRevisionContract` runs schema3 login/create/check handlers and SQLite, with explicitly isolated transport; it proves missing/mismatched rejection, current success, GET-only and zero generation attempts, not listening behavior. `console.mjs` retains native three-viewpoint creation/check and requires status200, a positive integer sent revision, success/applicability and matching revision. All are in unchanged `make verify`; full native checks still require the host.

### 7. Wrong / Correct
Wrong: logging a check body or raw assertion to diagnose failures. Correct: split service-create/check-http/check-contract/check-dialog and emit only numeric write/check statuses and revision-present/success/applicable/revision-match booleans, reset per viewport. Preserve every original gate; no credentials, IDs, URLs, body or DOM dump in the failure receipt.

## Scenario: native model editor locator ownership

### 1. Scope / Trigger
Host seq13 completed the service check successfully but stopped at model-1280 with Error. Source inspection identifies a deterministic driver ambiguity; this does not establish the only host failure. Do not change product labels or backend validation to fit a test selector.

### 2. Signatures
`modelEditor(model)` constructs both the `remote` field labeled `远端模型名` and the `routes` textarea whose label includes that same phrase. `console.mjs` must use `page.getByLabel('远端模型名',{exact:true})` for the remote input.

### 3. Contracts
Accessible label substring matching is not unique here. Select the actual exact label; preserve the additional route field and all create/grant assertions. Do not add a first/nth fallback, programmatic product value assignment or API preseed. The native driver requires one exact control before filling it.

### 4. Validation / Error Matrix
The actual editor produces two substring matches and one exact match. The supplemental VM regression rejects an ambiguous match, and the corrected driver fills alias/upstream/remote only. Native model-dialog/fields/create-http/grants-dialog/grants-http stages distinguish locator, write and readback failures. Failure receipts expose only fixed stage, numeric statuses/counts, booleans and an allowlisted error kind; never the raw exception message or page/API contents.

### 5. Good / Base / Bad Cases
Good: exact remote label selects the intended input without editing routes. Base: a missing or duplicated exact label fails rather than selecting another field. Bad: substring lookup, `.first()`, changing visible product copy solely to satisfy a driver or claiming the VM's label simulation proves Chromium behavior.

### 6. Tests Required
`workspace.test.mjs` executes actual `modelEditor` field construction and the actual model-field segment of `console.mjs`, requiring the exact three intended writes. This is a supplemental driver contract test, not real DOM/auth/network evidence. The original three-viewport real Go console journey remains in unchanged `make verify`, with all model creation/tenant grants, later Key/billing/identity and keyboard assertions intact. Obtain a fresh host receipt before closing the native failure.

### 7. Wrong / Correct
Wrong: treating a coarse Error as proof of a supplier/Go failure or repeatedly changing async behavior without diagnostics. Correct: preserve seq13, demonstrate the exact source mismatch with a red/green regression, add bounded stages and defer actual browser mechanism and the remaining journey to the host.

## Scenario: coarse Key journey timeout

### 1. Scope / Trigger
The host seq15 failed at `key-1280`, which covered navigation, asynchronous catalog/member dependencies, issuance, one-time dialog, explicit transfer, ordinary call and same-ID detail. That receipt does not locate the failing action. Do not change product cleanup, waits or provider logic from that coarse timeout alone.

### 2. Signatures
`consoleSnapshot(page)` executes a read-only DOM projection through the running browser. `console.mjs` retains the original actions/HTTP status and same-ID assertions and emits `key-navigation`, `key-dialog`, `key-fields`, `key-model-selection`, `key-create-http`, `key-secret-dialog`, `key-secret-transfer`, `key-call-fields`, `key-call-http`, `key-request-detail-action` and `key-request-detail-visible` with viewport suffixes.

### 3. Contracts
Snapshot fields are fixed page enums, boolean dialog/control/failure states and numeric invalid-control/option/selection counts. `call_key_nonempty` is only a boolean, never the value or length. Response observations retain only numeric catalog/member/issuance/call/detail statuses, never paths, headers or bodies. Page errors and failed requests are counted without printing their messages. Failure projection allowlists exception names/kinds and excludes secrets, prompts, completions, Request IDs, account names and raw API/notice/editor text.

### 4. Validation / Error Matrix
Missing DOM controls produce false/zero and unknown page, not fabricated success. An unavailable page produces `snapshot_available:false` and retains the failing stage. Dependency failure and invalid form are separate from HTTP rejection; an issuance success does not establish secret transfer or call success. A timeout with no call response and an empty call field is evidence to investigate transfer/validation, not proof of native event ordering.

### 5. Good / Base / Bad Cases
Good: obtain the new native stage and safe status/DOM facts before selecting the next fix. Base: no extra observations on success beyond the existing journey. Bad: dumping an error stack, DOM, network body or screenshot while a one-time secret is visible; substituting an API-issued Key; delaying arbitrarily or weakening assertions to force green.

### 6. Tests Required
`workspace.test.mjs` executes the actual snapshot function with secret-bearing and missing-control fixtures and verifies the exact projection and non-disclosure. These tests prove projection only, not Chromium lifecycle or backend behavior. The full unchanged `make verify` remains the host gate with all three console viewports, migration/Go/SSE/identity/billing assertions. Diagnosis-only candidates must not be reported as repaired or accepted.

### 7. Wrong / Correct
Wrong: infer that seq15 `write_status:201` is Key issuance; it may still be the preceding price publication. Correct: use the new `key-create-http` stage and `key_create_status` to bind the observation to the Key request, preserving seq15 as failed historical evidence.

## Scenario: continuous controlled completion protocol

### 1. Scope / Trigger
Host seq26 reached J01's first proxy call but returned 502. The continuous fixture omitted `message.role` in both known and unknown usage responses. Production `validResponse` requires `assistant`; do not relax product parsing to fit a test supplier.

### 2. Signatures
`continuousUpstream(dispatched *atomic.Int64) http.Handler` in tests/browser/journeys.go is the actual browser upstream handler and the regression target. `completionFacts(value)` in journeys.mjs projects authoritative request detail only after a status mismatch.

### 3. Contracts
The controlled handler emits `{"choices":[{"message":{"role":"assistant","content":"controlled"}}]}`. Known usage remains 3 input / 4 output / 7 total; unknown usage omits usage entirely, retaining pending liability. Do not seed business objects or change records to satisfy the journey. Failure evidence records expected/actual status and dispatch delta, plus terminal/code from explicit allowlists, numeric upstream status, pending flag, attempt count and at most three attempt summaries. Never output IDs, raw bodies, credentials, prompts, completions, upstream names or arbitrary error strings. Detail lookup is GET-only, bounded to five seconds; diagnostic failure must not replace the original status assertion.

### 4. Validation / Error Matrix
| Condition | Required outcome |
|---|---|
| Normal completion missing/wrong assistant role | Keep production protocol rejection |
| Controlled assistant completion, known usage | 200; cost 7 / tenant fee 14 at journey rates |
| Controlled assistant completion, usage omitted | 200; pending remains 1, not fabricated zero |
| Failure detail unavailable | Retain original failed status; report unavailable safely |
| Arbitrary fields/error values in detail | Exclude or replace with `other`; no raw data |

### 5. Good / Base / Bad Cases
Good: repair the owning test upstream and retain the negative parser regression. Base: an omitted supplier usage field is a valid completion but an unresolved financial liability. Bad: accept missing roles, replace unknown usage with zero, or claim helper tests passed all browser journeys.

### 6. Tests Required
`TestContinuousUpstreamResponseContract` invokes the actual fixture with known and unknown requests (non-listening). `TestCompletionAssistantRoleContract` invokes the production validator with missing/wrong/assistant roles. `TestContinuousUpstreamRelayHTTP` uses actual temporary upstream HTTP through schema3 relay handlers, management APIs and persisted request/charge detail; it requires host listening and is not a complete browser or relay-network journey. The Node projection regression excludes secret-bearing/unrecognized fields and bounds attempts. All enter the unchanged make verify; full continuous J01–J07 at 1280/1440/390 and independent QA still decide acceptance.

### 7. Wrong / Correct
Wrong: `message:{content:"controlled"}` or weakening `validResponse`. Correct: `message:{role:"assistant",content:"controlled"}`, with missing-role rejection intact and separately attributed host evidence.

## Scenario: request detail entry belongs to the current page

### 1. Scope / Trigger
Host seq55 completed the three editor-interruption viewports and J01–J05/J07 at1280, then J06 timed out waiting for request detail. The reconciled request is on the records page; the call-page action does not exist there. This is a test-driver entry mismatch, not evidence of missing backend reconciliation or a listening failure.

### 2. Signatures
`requestDetail(page,id,entry='call')` in `tests/browser/journeys.mjs`. The explicit `records` entry uses `row(page,id).getByRole('button',{name:'详情 / 核对',exact:true})`; the default retains the call-page `查看同Request ID日志` action.

### 3. Contracts
Arm the same exact request-detail HTTP response before native activation. Await response and click with `Promise.all` so either rejection has an attached handler. Require200, matching response ID and visible matching detail title. Do not replace the UI step with `page.request.get`, relax the timeout or change product controls.

### 4. Validation / Error Matrix
Call-page entry -> original local Request ID action. Records-page entry -> exact request row action. Click/HTTP rejection -> propagate failure, not a partial success. Missing controls -> keep failure; do not fall back to a different page or request.

### 5. Good / Base / Bad Cases
Good: J06 explicitly selects `records` after reconciling and preserves confirmed-token/detail/export assertions. Base: all call-page uses keep the default entry. Bad: activate the absent call action on records, leave an orphan response promise, seed an API detail as UI evidence, or claim a VM fixture proves Chromium behavior.

### 6. Tests Required
`journeys.test.mjs` executes the actual helper with call/records locator contracts, response-before-click ordering and click failure. It binds J06 to the records entry. These prove the driver contract only. Current-fingerprint host `make verify` must still complete every J01–J07, management summary,63 visual images/roles, migration and backend gate; independent QA owns closure.

### 7. Wrong / Correct
Wrong: `requestDetail(admin,unknown.id)` after navigating to records. Correct: `requestDetail(admin,unknown.id,'records')`, with unchanged HTTP/detail/token/export assertions.

## Scenario: blank original PNG with visible DOM ancestry

### Scope and signatures
Run b7f93c24bc1ea4d0fe182f12aa7622ae execution22 reproduced execution20's exact uniform PNG digest. The original sidecar shows stable geometry, loaded fonts, zero animations and connected visible ancestry; this does not identify a compositor or raster cause. `observeCaptureCompositor(page)` in `tests/browser/capture-integrity.mjs` supplies `{facts,finish}` to the existing single `captureIntegrityScreenshot(page,path)`.

### Contracts and errors
Use a page-owned CDP session, `LayerTree.enable` and `Page.getLayoutMetrics` only. Observe tree/paint events without waiting for them. Retain at most8 events of each kind,16 layers per tree and16 numeric transform entries per layer; retain total event counts and omitted layer counts. Project numeric geometry, content/paint flags and UTC only, never protocol IDs, node IDs, text, credentials, URLs, raw errors or arbitrary payloads. Read CSS viewport/content geometry before and after the original screenshot and detach in finally. Unsupported CDP/layout -> available:false; no event -> unknown, not proof that nothing painted. Observer availability or cleanup cannot replace the original PNG integrity result.

### Good / base / bad cases and required tests
Good: one original PNG, original pixel/geometry/secret/scroll assertions, bounded compositor sidecar and session cleanup even after failure. Base: no CDP available still preserves the original error and existing DOM/pixel facts. Bad: enable rendering-changing domains, force redraw, take a second image, add waits/retries, infer Chromium behavior from a fixture, or call a DOM-visible blank PNG usable. Node regressions exercise bounded projection, unsupported domain/layout, detach and unchanged single-capture failure. Only current-source host `make verify` can establish actual CDP observations; even a full green needs independent QA and cannot prove historical empty-image root cause.

## Scenario: stable capture lifecycle but blank original pixels

### Scope / trigger and signatures
Execution34 reproduced the same blank PNG after wheel delivery was observed before capture. Six passive lifecycle samples remained visible at 1280x900/scrollY148; resize event names alone did not demonstrate a size change. Do not rotate speculative product, input or viewport fixes. `observeCaptureFrameTrace(session)` returns `{facts,mark,finish}` on the existing page-owned CDP session, enabled by `observeCaptureCompositor(page,{frameTrace:true})` only for the original capture helper.

### Contracts
Read `Page.getFrameTree` for the main-frame ID in memory only. Use `Tracing.start` with ReportEvents, a 512KiB recordUntilFull buffer, sampling/system tracing disabled, and only `disabled-by-default-{cc.debug,devtools.timeline.layers,devtools.timeline}` categories; explicitly exclude `*` and the screenshot category. Disabled-by-default inclusions alone do not exclude ordinary default categories: Chromium147's `base/trace_event/trace_config_category_filter.cc` keeps them in `disabled_categories_`, and the empty ordinary include list enables unexcluded default categories. Execution36's trace lost data, omitted the return marker and had 6854 ignored events. The exclusion narrows collection, not a larger buffer or longer wait. Do not collect a raw trace file, images, stack samples, request payloads or DOM. Two fixed `Tracing.recordClockSyncMarker` names bracket the unchanged sole `page.screenshot({path,fullPage:true})`. There is no timer, RAF, screenshot/assertion retry, alternative capture, synthetic delay, viewport/input override, force redraw or deadline extension.

When snapshots cannot be attributed, report only fixed `identity_shape` booleans: finite `args.layerTreeId`, finite `snapshot.layer_tree_id`, and hexadecimal presence at `id`, `id2.local`, `id2.global`. Do not serialize any of those identifiers, arbitrary keys or values. Shape observations alone do not prove equivalent identity. Preserve the original explicit numeric argument matching; the Chromium147 contract below additionally permits its narrow legacy object-snapshot identity. Execution36's four frames all remained unattributed despite a main-frame SetLayerTreeId association; offscreen quads in a lost/unattributed trace are not proof of root overscroll. Unit fixtures prove serialization/filter configuration, not Chromium category delivery or compositor behavior; host must validate coverage.

Project only the two known snapshot event names (`cc::LayerTreeHostImpl`, `LayerTreeHostImpl:snapshot`), fixed clock markers and numeric draw geometry: at most8 snapshots,16 layers each,16 row-major screen transform entries and8 quad coordinates per layer,4 markers, total snapshot/ignored counts. Shape fields are fixed; nonfinite or wrong-type numbers become null. Layer IDs, trace IDs, frame IDs, URLs, names, paint records, tile records and arbitrary args never enter the sidecar. `SetLayerTreeId` main-frame associations are resolved after all chunks arrive; at most8 tree IDs stay in memory and are then discarded. Only an explicit numeric `args.layerTreeId` may be matched; never guess from legacy hexadecimal trace object IDs. Report `matched`, `other` or `unattributed`, never treat browser-global draws as page evidence without attribution. Snapshot ordering/timestamps are the trace clock, not page fixed Date; compare only with same-trace clock markers. LayerTree protocol transforms and trace screen-space matrices have different layouts and are not interchangeable.

### Validation / error matrix
- Start unsupported/already occupied -> available:false, no Tracing.end (do not stop somebody else's trace).
- End rejected or session close -> complete:false; always remove owned listeners and detach the original CDP session.
- Flush completes -> complete:true; retain the actual dataLossOccurred boolean, null if unavailable. This is collection completion, not screenshot success.
- No matching main-frame association/snapshot or buffer loss -> unknown/incomplete diagnostic coverage, never a render verdict.
- Screenshot/geometry/pixel assertion fails -> preserve the original error identity, PNG and strict failure result; trace cannot turn it green. Tracing itself may perturb scheduling, so new green does not prove historical root cause.

### Good / base / bad and tests
Good: compare matched draw-time layer quads/opacity/transforms and device viewport inside the original capture markers with the original failed PNG. Base: unsupported trace still preserves existing lifecycle, LayerTree and pixel facts. Bad: infer root overscroll merely from a large wheel delta, equate paint notifications with output pixels, export raw snapshots, silently switch capture methods or infer Chromium behavior from the Node fixture. Required wildcard Node tests cover bounds/privacy, late attribution, unsupported/end-failure/closed/lost/malformed cases and one unchanged capture for success/blank/exception, including listener cleanup and original failure identity. Host-only `make verify` plus independent QA remain mandatory.

### Chromium147 legacy object-snapshot identity
Execution38 observed eight unscoped hexadecimal object IDs and no numeric argument/snapshot identity; the PNG remained blank and trace loss remained true. Chromium147.0.7727.15 `inspector_trace_events.cc:inspector_set_layer_tree_id::Data` obtains the main frame's `ChromeClientImpl::GetLayerTreeId`, which delegates to `WebFrameWidgetImpl::GetLayerTreeId` returning the widget's `LayerTreeHost::GetId`. `LayerTreeHost::CreateLayerTreeHostImpl` passes that same int ID to `LayerTreeHostImpl`, whose `GenerateCompositorFrame` emits `cc::LayerTreeHostImpl` object snapshots with `id_`. These are integer IDs, not object pointers. Source references and the boundary audit are in the current Run's `delivery/verification/development-resume.md`.

Only this exact event name, legacy phase `O`, unscoped plain `id` without `id2`, a positive int32 hexadecimal value, and a positive safe-integer PID may match the same PID plus ID from a main-frame `SetLayerTreeId`. Store at most eight process keys transiently, resolve after collection (association events can arrive later), then clear them. Never export PID, frame/tree/object ID, raw trace or scope. Aliases, other phases, scoped IDs, malformed/overflow/zero IDs or missing PID remain unattributed; another known process/tree is other. Keep the existing numeric path unchanged. Matching is only attribution: loss, missing markers and non-bracketed frames still mean incomplete coverage, not a rendering verdict. Wildcard tests cover same-process match, same-ID different process/subframe rejection, late mapping, int32 boundaries, privacy and loss remaining true after a match. No screenshot, input, timeout, buffer or product behavior changes accompany this identity correction.
