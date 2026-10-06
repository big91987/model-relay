import assert from 'node:assert/strict';
import { chromium } from 'playwright';
import { mkdir, appendFile } from 'node:fs/promises';
import { createHash } from 'node:crypto';

let browser;
let checkpoint = 'launch';
let viewport = 0;
let passwordProbe = null;
let keyboardProbe = null;
let clipboardProbe = null;
let modelProbe = null;
let modelDiagnosticFile = null;
const modelReadStates = new WeakMap();
function trackModelReads(page) {
  const state = {models: {started: 0, finished: 0, failed: 0, pending: 0}, overview: {started: 0, finished: 0, failed: 0, pending: 0}};
  const pending = new Map();
  modelReadStates.set(page, state);
  page.on('request', request => {
    const path = new URL(request.url()).pathname;
    const name = path === '/admin/api/models' ? 'models' : path === '/admin/api/overview' ? 'overview' : null;
    if (!name || request.method() !== 'GET') return;
    pending.set(request, name); state[name].started++; state[name].pending++;
  });
  const finish = (request, failed) => {
    const name = pending.get(request);
    if (!name) return;
    pending.delete(request); state[name].pending--; state[name][failed ? 'failed' : 'finished']++;
  };
  page.on('requestfinished', request => finish(request, false));
  page.on('requestfailed', request => finish(request, true));
}
async function beginModelDiagnostics(page) {
  await page.evaluate(() => {
    const original = document.getElementById('test-model');
    const originalOptions = Array.from(original.options);
    const trace = [];
    let samples = 0, optionRenders = 0, replacements = 0, lastSelect = original;
    const sample = stage => {
      const select = document.getElementById('test-model');
      if (select !== lastSelect) { replacements++; lastSelect = select; }
      const focused = document.activeElement;
      const focusName = focused === select ? 'test-model' : focused === document.querySelector('#test h1') ? 'test-heading' : focused === document.querySelector('nav [data-view="test"]') ? 'nav-test' : !focused ? 'none' : 'other';
      const value = {stage, focusName, focused: focused === select, documentFocused: document.hasFocus(), sameSelect: select === original, connected: !!select?.isConnected, disabled: !select || select.disabled, optionCount: select?.options.length || 0, selectedIndex: select?.selectedIndex ?? -1, sameOptions: !!select && select.options.length === originalOptions.length && originalOptions.every((option, index) => select.options[index] === option), optionRenders, replacements, loading: !document.getElementById('page-loading').hidden, panelVisible: !document.getElementById('test').hidden, navigation: typeof navigationVersion === 'number' ? navigationVersion : -1, load: typeof loadSequence === 'number' ? loadSequence : -1, viewIsTest: typeof active === 'string' && active === 'test', hashIsTest: location.hash === '#test'};
      samples++; trace.push(value);
      if (trace.length > 40) trace.splice(8, 1);
      return value;
    };
    const key = event => {
      const selectTarget = event.target === document.getElementById('test-model');
      const keyName = ['ArrowUp', 'ArrowDown', 'Tab'].includes(event.key) ? event.key : selectTarget && event.key.length === 1 ? 'printable' : null;
      if (!keyName) return;
      sample(event.type + '-' + keyName + (selectTarget ? '-select' : '-other'));
    };
    const focus = event => sample(event.type);
    const navigation = event => sample(event.type);
    const mutation = new MutationObserver(records => {
      let relevant = false;
      for (const record of records) {
        const select = document.getElementById('test-model');
        if (record.type === 'childList' && (record.target === original || record.target === select)) { optionRenders++; relevant = true; }
        if (select !== lastSelect || record.target === select || ['page-loading', 'test'].includes(record.target.id)) relevant = true;
      }
      if (relevant) sample('render-or-loading');
    });
    mutation.observe(document.documentElement, {subtree: true, childList: true, attributes: true, attributeFilter: ['hidden', 'disabled']});
    document.addEventListener('keydown', key, true); document.addEventListener('keyup', key, true);
    document.addEventListener('focusin', focus, true); document.addEventListener('focusout', focus, true);
    window.addEventListener('hashchange', navigation); window.addEventListener('popstate', navigation);
    window.__relayModelDiagnostics = {
      sample: stage => ({...sample(stage), samples, truncated: samples > 40, trace: trace.slice()}),
      stop: () => { mutation.disconnect(); document.removeEventListener('keydown', key, true); document.removeEventListener('keyup', key, true); document.removeEventListener('focusin', focus, true); document.removeEventListener('focusout', focus, true); window.removeEventListener('hashchange', navigation); window.removeEventListener('popstate', navigation); delete window.__relayModelDiagnostics; }
    };
  });
}
async function sampleModelDiagnostics(page, stage) {
  const boundary = await page.evaluate(value => window.__relayModelDiagnostics.sample(value), stage);
  const reads = modelReadStates.get(page);
  return {...boundary, reads: reads ? {models: {...reads.models}, overview: {...reads.overview}} : null, readsSettled: !!reads && reads.models.started > 0 && reads.overview.started > 0 && reads.models.pending === 0 && reads.overview.pending === 0 && !boundary.loading && boundary.panelVisible && boundary.viewIsTest};
}
async function endModelDiagnostics(page) { await page.evaluate(() => window.__relayModelDiagnostics?.stop()); }
function modelFailureDiagnostic(probe) {
  if (!probe) return null;
  const boundary = value => ({stage: value.stage, focusName: value.focusName, focused: value.focused, documentFocused: value.documentFocused, sameSelect: value.sameSelect, sameOptions: value.sameOptions, disabled: value.disabled, optionCount: value.optionCount, selectedIndex: value.selectedIndex, navigation: value.navigation, load: value.load, loading: value.loading, optionRenders: value.optionRenders, replacements: value.replacements, readsSettled: value.readsSettled, modelsPending: value.reads?.models.pending, overviewPending: value.reads?.overview.pending});
  return {...boundary(probe), targetIndex: probe.targetIndex, enabledTargetIndex: probe.enabledTargetIndex, targetEnabled: probe.targetEnabled, matchesTarget: probe.matchesTarget, reads: probe.reads, samples: probe.samples, truncated: probe.truncated, boundaryCount: probe.boundaryCount, boundaries: probe.boundaries?.map(boundary), recentEvents: probe.trace?.slice(-8).map(boundary)};
}
const failureLineLimit = 8192;
const failurePrefix = 'FAIL actual Go service browser journey: ';
function serializeFailure(summary) {
  const model = summary.modelProbe;
  const boundaries = model?.boundaries || [];
  const chosen = new Set();
  for (const stage of ['before keyboard reach', 'after keyboard reach', 'before ArrowUp', 'after ArrowUp', 'before ArrowDown', 'after ArrowDown', 'before type-ahead', 'after type-ahead', 'before Tab', 'after Tab', 'after commit']) {
    const first = boundaries.findIndex(value => value.stage === stage);
    if (first < 0) continue;
    chosen.add(first);
    chosen.add(boundaries.findLastIndex(value => value.stage === stage));
  }
  const retained = boundaries.filter((value, index) => chosen.has(index));
  const value = {...summary, checkpoint: summary.checkpoint.slice(0, 512), location: summary.location.slice(0, 64), error: summary.error.slice(0, 64), outputLimitBytes: failureLineLimit, compacted: false,
    modelProbe: model ? {...model, boundaries: retained, omittedBoundaries: Math.max(0, (model.boundaryCount || boundaries.length) - retained.length), omittedEvents: Math.max(0, (model.samples || 0) - (model.recentEvents?.length || 0))} : null};
  const line = () => failurePrefix + JSON.stringify(value);
  if (Buffer.byteLength(line() + '\n', 'utf8') > failureLineLimit) {
    value.compacted = true;
    if (value.modelProbe) {
      value.modelProbe.omittedEvents = value.modelProbe.samples || 0;
      value.modelProbe.recentEvents = [];
      value.modelProbe.boundaries = retained.map(boundary => ({stage: boundary.stage, focusName: boundary.focusName, focused: boundary.focused, sameSelect: boundary.sameSelect, sameOptions: boundary.sameOptions, disabled: boundary.disabled, optionCount: boundary.optionCount, selectedIndex: boundary.selectedIndex, navigation: boundary.navigation, load: boundary.load, loading: boundary.loading, modelsPending: boundary.modelsPending, overviewPending: boundary.overviewPending}));
    }
    value.omittedResponses = value.responses.length;
    value.responses = [];
    value.passwordProbe = value.passwordProbe ? {omitted: true} : null;
    value.keyboardProbe = value.keyboardProbe ? {omitted: true} : null;
    value.clipboardProbe = value.clipboardProbe ? {omitted: true} : null;
  }
  if (Buffer.byteLength(line() + '\n', 'utf8') > failureLineLimit) {
    value.checkpoint = 'failure-summary-compacted';
    if (value.modelProbe) {
      value.modelProbe.boundaries = value.modelProbe.boundaries.filter((boundary, index, values) => values.findIndex(candidate => candidate.stage === boundary.stage) === index);
      value.modelProbe.omittedBoundaries = Math.max(0, (model.boundaryCount || boundaries.length) - value.modelProbe.boundaries.length);
    }
  }
  assert.equal(Buffer.byteLength(line() + '\n', 'utf8') <= failureLineLimit, true, 'safe failure summary must fit output budget');
  return line();
}
const responses = [];
const origin = process.env.RELAY_BROWSER_URL;
async function visibleText(page, selector, text) {
  if (selector === '#overview-evidence') { const detail = page.locator('#overview-evidence').locator('..'); if (!await detail.evaluate(node => node.open)) await detail.locator('summary').click(); }
  checkpoint = `visible ${selector} contains ${text}`;
  await page.locator(selector).filter({hasText: text}).waitFor({state: 'visible', timeout: 10000});
}
function nextOverview(page) {
  return page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/overview' && response.request().method() === 'GET');
}
async function evidenceSummary(page, value, expectedID) {
  checkpoint = 'visible success time and configuration version match request evidence';
  const evidence = value.degraded ? null : value.call_evidence;
  if (!evidence) {
    assert.equal(expectedID == null, true, 'expected absence of current success evidence');
    await visibleText(page, '#summary', value.degraded ? '存储不可用' : value.upstream.configured ? '尚无当前配置的生成成功记录' : '待配置：从接入上游开始');
    assert.equal(await page.locator('#overview-record').isDisabled(), true);
    assert.equal(await page.locator('#summary').evaluate(node => /调用时间|配置版本/.test(node.textContent)), false, 'no old metadata without evidence');
    return null;
  }
  if (expectedID) assert.equal(evidence.request_id, expectedID);
  assert.equal(typeof evidence.time === 'string' && evidence.time.length > 0, true);
  assert.equal(Number.isSafeInteger(evidence.upstream_revision) && evidence.upstream_revision > 0, true);
  await visibleText(page, '#summary', '当前配置曾成功调用 · 模型 ' + evidence.alias);
  await visibleText(page, '#overview-evidence', '请求 ' + evidence.request_id + '\n调用时间 ' + evidence.time + '\n上游配置版本 ' + evidence.upstream_revision + '\n' + evidence.scope);
  assert.equal(await page.locator('#overview-record').isEnabled(), true);
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true, 'evidence metadata fits viewport');
  return evidence;
}
async function evidenceOverview(page, expectedID) {
  const completed = nextOverview(page);
  await navigate(page, '概览');
  const response = await completed;
  assert.equal(response.status(), 200);
  return evidenceSummary(page, await response.json(), expectedID);
}
async function loginOverview(page, completed, width) {
  checkpoint = 'login overview matches persisted configuration and success evidence';
  const response = await completed;
  assert.equal(response.status(), 200);
  const value = await response.json();
  assert.equal(value.degraded, false);
  assert.equal(value.upstream.configured, width === 390);
  assert.equal(value.upstream.enabled, width === 390);
  assert.equal(value.upstream.credential_configured, width === 390);
  assert.deepEqual(value.models, width === 1280 ? {total: 0, enabled: 0} : {total: 2, enabled: 2});
  assert.deepEqual(value.keys, width === 1280 ? {total: 0, enabled: 0} : {total: 2, enabled: 1});
  if (width === 1280) {
    assert.equal(value.call_evidence, null);
    await visibleText(page, '#summary', '待配置：从接入上游开始');
    await visibleText(page, '#overview-upstream', '未配置');
  } else {
    assert.ok(value.call_evidence?.request_id);
    assert.equal(value.call_evidence.alias, 'alpha-1280');
    await visibleText(page, '#overview-evidence', value.call_evidence.request_id);
    await visibleText(page, '#overview-upstream', '1 个 / 启用 1');
  }
  await visibleText(page, '#overview-models', value.models.total + ' 个 / 启用 ' + value.models.enabled);
  await visibleText(page, '#overview-keys', value.keys.total + ' 个 / 启用 ' + value.keys.enabled);
  await evidenceSummary(page, value, width === 390 ? value.call_evidence.request_id : null);
}
async function navigate(page, name) {
  await revealAccount(page, name);
  checkpoint = `click ${name}`;
  const completed = name === '退出' ? page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/logout') : null;
  await page.getByRole('button', {name, exact: true}).click();
  if (completed) await completed;
}
async function expand(page, id) { if (!await page.locator(id).evaluate(node => node.open)) await page.locator(id + ' summary').click(); }
async function copyIssuedKey(page, context) {
  clipboardProbe = {stage: 'injected rejection'};
  await page.evaluate(() => {
    window.restoreClipboardWrite = navigator.clipboard.writeText;
    navigator.clipboard.writeText = async () => { throw new DOMException('Controlled clipboard rejection', 'NotAllowedError'); };
  });
  try {
    await keyboardButton(page, '复制密钥');
    await visibleText(page, '#error', '自动复制失败，请在当前展示中手动选择复制。');
    assert.equal(await page.locator('#issued').isVisible(), true, 'rejection retains one-time display for manual copy');
    assert.equal(await page.locator('#issued-key').evaluate(node => /^mr_/.test(node.value) && node.type === 'password'), true, 'manual path retains controlled key without exposing it');
    assert.equal(await page.locator('#copy-key').isEnabled(), true, 'rejected copy can be retried');
  } finally {
    await page.evaluate(() => { navigator.clipboard.writeText = window.restoreClipboardWrite; delete window.restoreClipboardWrite; });
  }
  clipboardProbe = {stage: 'origin-scoped native write'};
  await context.grantPermissions(['clipboard-read', 'clipboard-write'], {origin});
  try {
    await keyboardButton(page, '复制密钥');
    await page.waitForFunction(() => !document.getElementById('copy-key').disabled && (document.getElementById('notice').textContent.includes('已复制，请保管') || document.getElementById('error').textContent.includes('手动选择复制')));
    clipboardProbe = {stage: 'origin-scoped native write', ...await page.evaluate(() => ({
      secureContext: window.isSecureContext,
      documentFocused: document.hasFocus(),
      manualFallback: document.getElementById('error').textContent.includes('手动选择复制'),
      successNotice: document.getElementById('notice').textContent.includes('已复制，请保管')
    }))};
    await visibleText(page, '#notice', '已复制，请保管');
    checkpoint = 'clipboard native value matches controlled issued key';
    assert.equal(await page.evaluate(async () => (await navigator.clipboard.readText()) === document.getElementById('issued-key').value), true, 'native clipboard matches without returning key bytes');
  } finally {
    await page.evaluate(() => navigator.clipboard.writeText(''));
    await context.clearPermissions();
  }
  clipboardProbe = null;
}
async function keyboardAction(page, locator, key = 'Enter') {
  await keyboardFocus(page, locator, key);
  await page.keyboard.press(key);
  keyboardProbe = null;
}
async function keyboardFocus(page, locator, key = 'input') {
  checkpoint = 'native keyboard action with visible focus';
  await locator.waitFor({state: 'visible', timeout: 10000});
  const describe = node => {
    const style = getComputedStyle(node);
    const safeIDs = ['cancel-edit', 'copy-example', 'diagnostics'];
    const view = node.getAttribute('data-view');
    return {
      target: safeIDs.includes(node.id) ? node.id : ['overview', 'upstream', 'models', 'keys', 'test', 'records'].includes(view) ? 'nav-' + view : node.tagName === 'SUMMARY' ? 'summary' : 'button',
      focused: node === document.activeElement,
      focusVisible: node.matches(':focus-visible'),
      outlineStyle: ['none', 'auto', 'solid'].includes(style.outlineStyle) ? style.outlineStyle : 'other',
      outlineWidth: parseFloat(style.outlineWidth) || 0
    };
  };
  let tabCount = 1;
  await page.keyboard.press('Tab');
  keyboardProbe = {key, tabCount, ...await locator.evaluate(describe)};
  while (!keyboardProbe.focused && tabCount < 80) {
    await page.keyboard.press('Tab');
    tabCount++;
    keyboardProbe = {key, tabCount, ...await locator.evaluate(describe)};
  }
  assert.equal(keyboardProbe.focused, true, 'target must be reachable through native Tab');
  assert.equal(keyboardProbe.focusVisible && keyboardProbe.outlineStyle !== 'none' && keyboardProbe.outlineWidth > 0, true, 'native keyboard focus must have a visible outline');
  keyboardProbe = null;
}
async function keyboardInput(page, locator, value) {
  await keyboardFocus(page, locator);
  await page.keyboard.press('ControlOrMeta+A');
  if (value) await page.keyboard.insertText(value);
  else await page.keyboard.press('Backspace');
  assert.equal((await locator.inputValue()) === value, true, 'native keyboard input matches without disclosing value');
}
async function keyboardCheck(page, locator, checked) {
  await keyboardFocus(page, locator);
  if (await locator.isChecked() !== checked) await page.keyboard.press('Space');
  assert.equal(await locator.isChecked(), checked);
}
async function keyboardModel(page, alias) {
  modelDiagnosticFile = null;
  const locator = page.getByLabel('调用别名');
  await beginModelDiagnostics(page);
  const boundaries = [];
  let boundaryCount = 0;
  const capture = async stage => {
    const value = await sampleModelDiagnostics(page, stage);
    const {trace, ...boundary} = value;
    boundaryCount++; boundaries.push(boundary);
    if (boundaries.length > 24) boundaries.splice(8, 1);
    return {...value, boundaries: boundaries.slice(), boundaryCount};
  };
  try {
    modelProbe = await capture('before keyboard reach');
    await keyboardFocus(page, locator);
    checkpoint = 'native model selection commits enabled target';
    modelProbe = {...await capture('after keyboard reach'), ...await locator.evaluate((node, target) => ({targetIndex: Array.from(node.options).findIndex(option => option.value === target), enabledTargetIndex: Array.from(node.options).filter(option => !option.disabled).findIndex(option => option.value === target), targetEnabled: !node.disabled && Array.from(node.options).some(option => option.value === target && !option.disabled)}), alias)};
    assert.equal(modelProbe.targetIndex >= 0 && modelProbe.targetEnabled, true, 'requested enabled model must exist');
    const {targetIndex} = modelProbe;
    modelProbe = {...modelProbe, ...await capture('before type-ahead')};
    await page.keyboard.type(alias);
    modelProbe = {...modelProbe, ...await capture('after type-ahead')};
    modelProbe = {...modelProbe, ...await capture('before Tab')};
    await page.keyboard.press('Tab');
    modelProbe = {...modelProbe, ...await capture('after Tab')};
    await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => resolve())));
    modelProbe = {...modelProbe, ...await capture('after commit'), ...await locator.evaluate((node, target) => ({matchesTarget: node.options[node.selectedIndex]?.value === target}), alias)};
    assert.equal(modelProbe.selectedIndex === targetIndex && modelProbe.matchesTarget, true, 'native selection must commit the requested model');
    assert.equal((await locator.inputValue()) === alias, true);
  } finally {
    const {trace, ...cleanupBoundary} = await sampleModelDiagnostics(page, 'before cleanup');
    modelProbe = {...modelProbe, trace, cleanupBoundary};
    await endModelDiagnostics(page);
    await mkdir('test-results', {recursive: true});
    const path = 'test-results/model-selection-' + viewport + '.jsonl';
    const row = JSON.stringify({execution: process.pid, viewport, ...modelProbe}) + '\n';
    await appendFile(path, row);
    modelDiagnosticFile = {path, sha256: createHash('sha256').update(row).digest('hex'), bytes: Buffer.byteLength(row), execution: process.pid};
  }
  modelProbe = null;
}
async function revealAccount(page, name) { if (['退出', '重新登录', '重新检查会话'].includes(name) && !await page.locator('#account-actions').evaluate(node => node.open)) await page.locator('#account-actions summary').click(); }
async function keyboardButton(page, name) { await revealAccount(page, name); await keyboardAction(page, page.getByRole('button', {name, exact: true})); }
async function keyboardExpand(page, id) { if (!await page.locator(id).evaluate(node => node.open)) await keyboardAction(page, page.locator(id + ' summary')); }
async function keyboardStateChange(page, name, entity) {
  const path = entity === '模型' ? 'models' : 'keys';
  const written = page.waitForResponse(response => new URL(response.url()).pathname.startsWith('/admin/api/' + path + '/') && response.request().method() === 'PATCH');
  const reread = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/' + path && response.request().method() === 'GET');
  await keyboardButton(page, name);
  assert.equal((await written).status(), 200); assert.equal((await reread).status(), 200);
  const updated = page.getByRole('button', {name: name.replace(/^(停用|启用)/, name.startsWith('停用') ? '启用' : '停用'), exact: true});
  await updated.waitFor({state: 'visible'});
  await visibleText(page, '#notice', entity + '状态已更新');
  checkpoint = 'same-object action retains native focus after state change';
  assert.equal(await updated.evaluate(node => document.activeElement === node && node.matches(':focus-visible')), true, 'replacement action must retain same-row keyboard focus without another Tab');
}
async function keyboardSaveUpstream(page) {
  let releaseWrite;
  const gate = new Promise(resolve => { releaseWrite = resolve; });
  const heldWrite = async route => { if (['POST', 'PATCH'].includes(route.request().method())) await gate; await route.continue(); };
  let writeCount = 0;
  const countWrites = request => { if (new URL(request.url()).pathname.startsWith('/admin/api/upstreams') && ['POST', 'PATCH'].includes(request.method())) writeCount++; };
  page.on('request', countWrites);
  await page.route('**/admin/api/upstreams**', heldWrite);
  const started = page.waitForRequest(request => new URL(request.url()).pathname.startsWith('/admin/api/upstreams') && ['POST', 'PATCH'].includes(request.method()));
  const written = page.waitForResponse(response => new URL(response.url()).pathname.startsWith('/admin/api/upstreams') && ['POST', 'PATCH'].includes(response.request().method()));
  const reread = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/upstreams' && response.request().method() === 'GET');
  try {
    await keyboardButton(page, '保存上游'); await started;
    for (const id of ['upstream-url', 'upstream-token', 'upstream-enabled', 'save-upstream', 'refresh-upstream']) assert.equal(await page.locator('#' + id).isDisabled(), true);
    await page.locator('#upstream-form').evaluate(form => { form.requestSubmit(); form.requestSubmit(); });
    assert.equal(writeCount, 1, 'shared save/restore busy ownership prevents repeated writes');
    releaseWrite();
    const receipt = await written; assert.equal(receipt.status(), receipt.request().method() === 'POST' ? 201 : 200); assert.equal((await reread).status(), 200);
    await visibleText(page, '#notice', '上游已保存');
    assert.equal(writeCount, 1);
    assert.equal(await page.locator('#refresh-upstream').isDisabled(), false);
  } finally { releaseWrite(); page.off('request', countWrites); await page.unroute('**/admin/api/upstreams**', heldWrite); }
}
async function callControlState(page, disabled) {
  checkpoint = 'submitted parameter controls preserve current-call ownership';
  for (const id of ['test-model', 'test-key', 'test-message', 'test-stream']) assert.equal(await page.locator('#' + id).isDisabled(), disabled);
}
async function submittedValuesVisible(page, key, text, stream, alias) {
  await callControlState(page, true);
  assert.equal(await page.evaluate(expected => document.getElementById('test-key').value === expected.key && document.getElementById('test-message').value === expected.text && document.getElementById('test-model').value === expected.alias && document.getElementById('test-stream').checked === expected.stream, {key, text, stream, alias}), true, 'submitted values remain visible without reporting secrets');
}
function callInputsCleared(page) {
  return page.evaluate(() => document.getElementById('test-key').value === '' && document.getElementById('test-message').value === '');
}
async function keyboardReloadUpstream(page) {
  await keyboardInput(page, page.getByLabel('上游地址'), 'https://controlled-restore-draft.example/v1');
  await keyboardInput(page, page.getByLabel('上游凭据'), 'controlled-restore-secret');
  await keyboardCheck(page, page.getByLabel('启用上游'), false);
  checkpoint = 'upstream explicit reread completion barrier';
  let writes = 0;
  const countWrites = request => { if (new URL(request.url()).pathname.startsWith('/admin/api/upstreams') && ['POST', 'PATCH'].includes(request.method())) writes++; };
  page.on('request', countWrites);
  let releaseRead;
  const gate = new Promise(resolve => { releaseRead = resolve; });
  const heldRead = async route => {
    if (route.request().method() === 'GET') await gate;
    await route.continue();
  };
  const started = page.waitForRequest(request => new URL(request.url()).pathname === '/admin/api/upstreams' && request.method() === 'GET');
  const completed = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/upstreams' && response.request().method() === 'GET');
  await page.route('**/admin/api/upstreams**', heldRead);
  try {
    await keyboardButton(page, '重新读取上游');
    await started;
    assert.equal(await page.locator('#refresh-upstream').isDisabled(), true, 'pending reread is not settled despite retained credential status');
    assert.equal(await page.getByLabel('上游凭据').inputValue(), '', 'restore clears draft credential before GET completes');
    for (const id of ['upstream-url', 'upstream-token', 'upstream-enabled', 'save-upstream']) assert.equal(await page.locator('#' + id).isDisabled(), true);
    await page.keyboard.press('Enter'); await page.keyboard.press('Enter');
    assert.equal(writes, 0, 'pending restore cannot submit or overlap a save');
    releaseRead();
    assert.equal((await completed).status(), 200);
    await visibleText(page, '#notice', '已重新读取上游状态；保存不代表调用成功。');
    await page.waitForFunction(() => !document.getElementById('refresh-upstream').disabled);
    await visibleText(page, '#credential-state', '凭据已配置');
    assert.equal(await page.getByLabel('上游地址').inputValue(), process.env.RELAY_BROWSER_UPSTREAM);
    assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
    assert.equal(await page.getByLabel('启用上游').isChecked(), true);
    assert.equal(writes, 0, 'restoring public fields is GET-only');
  } finally {
    page.off('request', countWrites);
    releaseRead();
    await page.unroute('**/admin/api/upstreams**', heldRead);
  }
}
async function upstreamFailureJourney(page) {
  const save = async () => {
    const completed = page.waitForResponse(response => new URL(response.url()).pathname.startsWith('/admin/api/upstreams') && ['POST', 'PATCH'].includes(response.request().method()));
    await keyboardButton(page, '保存上游');
    return completed;
  };
  await keyboardInput(page, page.getByLabel('上游地址'), 'http://127.0.0.1:1/v1');
  await keyboardInput(page, page.getByLabel('上游凭据'), 'controlled-rejected-secret');
  assert.equal((await save()).status(), 400);
  await visibleText(page, '#error', '上游地址不符合');
  assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
  assert.equal(await page.getByLabel('上游地址').inputValue(), 'http://127.0.0.1:1/v1');
  const changedAddress = process.env.RELAY_BROWSER_UPSTREAM + '/changed';
  await keyboardInput(page, page.getByLabel('上游地址'), changedAddress);
  assert.equal((await save()).status(), 400);
  await visibleText(page, '#error', '地址变化时必须提供');
  assert.equal(await page.getByLabel('上游地址').inputValue(), changedAddress);
  await keyboardReloadUpstream(page);
  assert.equal((await save()).status(), 200, 'same address retains existing encrypted credential');
  await visibleText(page, '#notice', '上游已保存');
  let reads = 0, writes = 0;
  const failedRead = route => {
    if (route.request().method() === 'GET' && ++reads === 1) return route.fulfill({status: 503, contentType: 'application/json', body: JSON.stringify({error: {code: 'storage_unavailable'}})});
    if (['POST', 'PATCH'].includes(route.request().method())) writes++;
    return route.continue();
  };
  await page.route('**/admin/api/upstreams**', failedRead);
  try {
    await keyboardInput(page, page.getByLabel('上游凭据'), 'controlled-browser-upstream');
    assert.equal((await save()).status(), 200);
    await visibleText(page, '#error', '最新读取未确认');
    await visibleText(page, '#error', '存储不可用');
    assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
    assert.equal(await page.getByLabel('上游地址').inputValue(), process.env.RELAY_BROWSER_UPSTREAM);
    assert.equal(writes, 1, 'read failure never rewrites');
  } finally { await page.unroute('**/admin/api/upstreams**', failedRead); }
  await keyboardReloadUpstream(page);
  const failedWrite = route => ['POST', 'PATCH'].includes(route.request().method()) ? route.abort('failed') : route.continue();
  await page.route('**/admin/api/upstreams**', failedWrite);
  try {
    await keyboardInput(page, page.getByLabel('上游凭据'), 'controlled-network-secret');
    await keyboardButton(page, '保存上游');
    await visibleText(page, '#error', '保存未确认');
    assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
    assert.equal(await page.getByLabel('上游地址').inputValue(), process.env.RELAY_BROWSER_UPSTREAM);
  } finally { await page.unroute('**/admin/api/upstreams**', failedWrite); }
  await keyboardReloadUpstream(page);
  const failedRestore = route => route.request().method() === 'GET' ? route.fulfill({status: 503, contentType: 'application/json', body: JSON.stringify({error: {code: 'storage_unavailable'}})}) : route.continue();
  await page.route('**/admin/api/upstreams**', failedRestore);
  let restoreWrites = 0;
  const countRestoreWrites = request => { if (new URL(request.url()).pathname.startsWith('/admin/api/upstreams') && ['POST', 'PATCH'].includes(request.method())) restoreWrites++; };
  page.on('request', countRestoreWrites);
  try {
    await keyboardInput(page, page.getByLabel('上游地址'), 'https://controlled-failed-restore.example/v1');
    await keyboardInput(page, page.getByLabel('上游凭据'), 'controlled-restore-secret');
    await keyboardCheck(page, page.getByLabel('启用上游'), false);
    await keyboardButton(page, '重新读取上游'); await visibleText(page, '#error', '存储不可用');
    assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
    assert.equal(await page.getByLabel('上游地址').inputValue(), 'https://controlled-failed-restore.example/v1');
    assert.equal(await page.getByLabel('启用上游').isChecked(), false);
    assert.equal(restoreWrites, 0);
    assert.equal(await page.locator('#save-upstream').isDisabled(), false);
  } finally { page.off('request', countRestoreWrites); await page.unroute('**/admin/api/upstreams**', failedRestore); }
  await keyboardReloadUpstream(page);
}
async function dependencyJourney(page, width, keys, keyIDs, fillCall) {
  const switchView = async view => { await keyboardAction(page, page.locator('nav [data-view="' + view + '"]')); await page.locator('#' + view).waitFor({state: 'visible'}); };
  await switchView('upstream');
  await keyboardCheck(page, page.getByLabel('启用上游'), false);
  await keyboardSaveUpstream(page);
  await evidenceOverview(page, null);
  await switchView('test'); await visibleText(page, '#test-dependency', '配置并启用上游');
  assert.equal(await page.locator('#send').isDisabled(), true);
  await keyboardButton(page, '配置上游'); await page.locator('#upstream').waitFor({state: 'visible'});
  await keyboardCheck(page, page.getByLabel('启用上游'), true);
  await keyboardSaveUpstream(page);
  await switchView('test'); await keyboardModel(page, 'alpha-' + width);
  await switchView('models'); await keyboardStateChange(page, '停用模型 alpha-' + width, '模型');
  await switchView('test');
  assert.equal(await page.getByLabel('调用别名').inputValue(), '', 'disabled selection must not silently select beta');
  await keyboardButton(page, '配置模型'); await page.locator('#models').waitFor({state: 'visible'});
  await keyboardStateChange(page, '启用模型 alpha-' + width, '模型');
  await switchView('keys'); await keyboardStateChange(page, '停用密钥 first-' + width, '密钥');
  await switchView('test'); await fillCall(keys[0], 'dependency-rejected-fixture', false);
  await keyboardButton(page, '发送请求'); await visibleText(page, '#test-status', '失败');
  assert.equal(await callInputsCleared(page), true); assert.equal(await page.locator('#console').isVisible(), true);
  const rejectedID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
  await sameRecord(page, rejectedID, 'alpha-' + width, keyIDs[0], 'rejected', '未知 / 未知 / 未知');
  await switchView('keys'); await keyboardStateChange(page, '启用密钥 first-' + width, '密钥');
  await switchView('test'); await fillCall(keys[0], 'dependency-restored-fixture', false);
  await keyboardButton(page, '发送请求'); await visibleText(page, '#test-status', '完成');
  const restoredID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
  await sameRecord(page, restoredID, 'alpha-' + width, keyIDs[0], 'success', '0 / 2 / 2');
  for (const [view, entity, guidance] of [['models', '模型', '创建或启用模型'], ['keys', '密钥', '签发或启用密钥']]) {
    await switchView(view);
    const enabledNames = await page.locator('#' + (view === 'models' ? 'model-list' : 'key-list')).getByRole('button', {name: new RegExp('^停用' + entity + ' ')}).allTextContents();
    for (const name of enabledNames) await keyboardStateChange(page, name, entity);
    await switchView('test'); await visibleText(page, '#test-dependency', guidance);
    assert.equal(await page.locator('#send').isDisabled(), true);
    await keyboardButton(page, view === 'models' ? '配置模型' : '管理应用密钥'); await page.locator('#' + view).waitFor({state: 'visible'});
    for (const name of enabledNames) await keyboardStateChange(page, name.replace(/^停用/, '启用'), entity);
  }
  await switchView('test');
}
async function recordsBoundaryJourney(page, context, width, key, keyID, fillCall) {
  const upstreamCount = async () => {
    const response = await context.request.post(process.env.RELAY_BROWSER_CONTROL + '/upstream-count');
    assert.equal(response.status(), 204);
    return Number(response.headers()['x-controlled-upstream-requests']);
  };
  await navigate(page, '调用测试');
  const admin = await context.newPage();
  let releaseCall;
  const released = new Promise(resolve => { releaseCall = resolve; });
  const delayedCall = async route => { await released; await route.continue(); };
  await page.route('**/v1/chat/completions', delayedCall);
  try {
    await admin.goto(origin + '/admin/#models'); await admin.locator('#models').waitFor({state: 'visible'});
    admin.on('dialog', dialog => dialog.accept());
    const beforeModelRejection = await upstreamCount();
    await fillCall(key, 'model rejection fixture', false);
    const submitted = page.waitForRequest('**/v1/chat/completions'); await keyboardButton(page, '发送请求'); await submitted;
    await keyboardStateChange(admin, '停用模型 alpha-' + width, '模型');
    releaseCall(); await visibleText(page, '#test-status', '失败'); await visibleText(page, '#error', '模型不存在或已停用');
    assert.equal(await callInputsCleared(page), true);
    assert.equal(await upstreamCount(), beforeModelRejection, 'actual model rejection never reaches upstream');
    const id = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
    await sameRecord(page, id, 'alpha-' + width, keyID, 'rejected', '未知 / 未知 / 未知');
    await keyboardStateChange(admin, '启用模型 alpha-' + width, '模型');
  } finally { releaseCall(); await page.unroute('**/v1/chat/completions', delayedCall); await admin.close(); }
  await navigate(page, '调用测试');
  for (const [text, stream, state] of [['protocol fixture', false, 'upstream_failed'], ['interrupt fixture', true, 'stream_interrupted']]) {
    await fillCall(key, text, stream); await keyboardButton(page, '发送请求');
    await visibleText(page, '#test-status', '失败'); assert.equal(await callInputsCleared(page), true);
    const id = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
    await sameRecord(page, id, 'alpha-' + width, keyID, state, '未知 / 未知 / 未知');
    await navigate(page, '调用测试');
  }
  const networkFailure = route => route.abort('failed');
  await page.route('**/v1/chat/completions', networkFailure);
  try {
    await fillCall(key, 'network failure fixture', false); await keyboardButton(page, '发送请求');
    await visibleText(page, '#test-status', '失败'); assert.equal(await callInputsCleared(page), true);
    assert.equal(await page.locator('#view-record').isVisible(), false, 'network failure without ID does not invent a record');
  } finally { await page.unroute('**/v1/chat/completions', networkFailure); }
  await fillCall(key, 'unknown usage fixture', false); await keyboardButton(page, '发送请求'); await visibleText(page, '#test-status', '完成');
  const evidenceID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
  await sameRecord(page, evidenceID, 'alpha-' + width, keyID, 'success', '未知 / 未知 / 未知');
  await navigate(page, '调用测试');
  await fillCall(key, 'cancel fixture', true); await keyboardButton(page, '发送请求'); await visibleText(page, '#test-output', 'controlled first');
  const pendingID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
  const observer = await context.newPage();
  try {
    await observer.goto(origin + '/admin/#records'); await observer.locator('#records').waitFor({state: 'visible'});
    await keyboardInput(observer, observer.getByLabel('精确请求 ID'), pendingID); await keyboardButton(observer, '定位记录');
    await visibleText(observer, '#record-detail', 'started');
    await visibleText(observer, '#record-detail', '未知 / 未知 / 未知');
  } finally { await observer.close(); }
  await keyboardButton(page, '取消请求'); await visibleText(page, '#test-status', '已取消');
  assert.equal(await callInputsCleared(page), true);
  const beforeUnknownKeys = await upstreamCount();
  for (let index = 0; index < 52; index++) {
    await fillCall('mr_unrecognized_pagination_fixture', 'pagination rejection fixture', false);
    const completed = page.waitForResponse('**/v1/chat/completions');
    await keyboardButton(page, '发送请求'); assert.equal((await completed).status(), 401);
    await visibleText(page, '#test-status', '失败'); assert.equal(await callInputsCleared(page), true);
  }
  assert.equal(await upstreamCount(), beforeUnknownKeys, '52 actual unknown-key requests have zero upstream touches');
  const newestID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
  await keyboardButton(page, '查看本次记录'); await visibleText(page, '#record-detail', newestID);
  await visibleText(page, '#record-detail', '未识别应用密钥');
  assert.equal((await page.locator('#record-detail').textContent()).includes('alpha-' + width), false, 'unknown key must not acquire submitted alias');
  let exactQueries = 0;
  const countExact = request => { if (/\/admin\/api\/requests\/[A-Za-z0-9_-]+$/.test(new URL(request.url()).pathname)) exactQueries++; };
  page.on('request', countExact);
  try {
    await navigate(page, '调用记录'); await page.locator('#records').waitFor({state: 'visible'});
    assert.equal(exactQueries, 0, 'generic records navigation is not implicitly filtered');
    assert.equal(await page.getByLabel('精确请求 ID').inputValue(), '');
    assert.equal((await page.locator('#record-list').textContent()).includes(evidenceID), false, 'older success lies beyond first real 50-row page');
    const nextPage = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/requests' && new URL(response.url()).searchParams.has('cursor'));
    await keyboardButton(page, '下一页'); assert.equal((await nextPage).status(), 200);
    await visibleText(page, '#record-list', evidenceID);
    await keyboardInput(page, page.getByLabel('精确请求 ID'), evidenceID); await keyboardButton(page, '定位记录');
    await visibleText(page, '#record-detail', evidenceID);
    await keyboardInput(page, page.getByLabel('精确请求 ID'), 'Z'.repeat(22)); await keyboardButton(page, '定位记录');
    await visibleText(page, '#record-status', '暂未找到'); assert.equal(await page.locator('#record-detail').textContent(), '');
    const unavailable = route => route.fulfill({status: 503, contentType: 'application/json', body: JSON.stringify({error: {code: 'storage_unavailable'}})});
    await page.route('**/admin/api/requests/' + evidenceID, unavailable);
    try {
      await keyboardInput(page, page.getByLabel('精确请求 ID'), evidenceID); await keyboardButton(page, '定位记录');
      await visibleText(page, '#record-status', '存储不可用');
      assert.equal(await page.locator('#record-detail').textContent(), '');
    } finally { await page.unroute('**/admin/api/requests/' + evidenceID, unavailable); }
    await keyboardButton(page, '定位记录'); await visibleText(page, '#record-detail', evidenceID);
    await evidenceOverview(page, evidenceID);
    const failedOverview = route => route.fulfill({status: 503, contentType: 'application/json', body: JSON.stringify({error: {code: 'storage_unavailable'}})});
    await page.route('**/admin/api/overview', failedOverview);
    try {
      await keyboardButton(page, '重新读取状态'); await visibleText(page, '#summary', '状态未知');
      assert.equal(await page.locator('#overview-record').isDisabled(), true, 'failed refresh must not retain an actionable old evidence ID');
      await visibleText(page, '#overview-models', '未知');
      assert.equal(await page.locator('#summary').evaluate(node => /调用时间|配置版本/.test(node.textContent)), false, 'failed read discards old time and revision');
    } finally { await page.unroute('**/admin/api/overview', failedOverview); }
    const recoveredOverview = nextOverview(page);
    await keyboardButton(page, '重新读取状态');
    const recoveredResponse = await recoveredOverview;
    assert.equal(recoveredResponse.status(), 200);
    let currentEvidence = await evidenceSummary(page, await recoveredResponse.json(), evidenceID);
    for (const lifecycle of ['current local ID differs', 'reload', 'relogin']) {
      const lifecycleOverview = lifecycle === 'current local ID differs' ? null : nextOverview(page);
      if (lifecycle === 'current local ID differs') {
        const concurrent = await context.newPage();
        try {
          await concurrent.goto(origin + '/admin/#test'); await concurrent.locator('#test').waitFor({state: 'visible'});
          await keyboardInput(concurrent, concurrent.getByLabel('应用密钥', {exact: true}), 'mr_unrecognized_pagination_fixture');
          await keyboardModel(concurrent, 'alpha-' + width);
          await keyboardInput(concurrent, concurrent.getByLabel('文本消息'), 'pagination rejection fixture');
          await keyboardButton(concurrent, '发送请求'); await visibleText(concurrent, '#test-status', '失败');
          assert.equal(await callInputsCleared(concurrent), true);
          assert.equal(await upstreamCount(), beforeUnknownKeys, 'concurrent new rejected row does not reach upstream');
        } finally { await concurrent.close(); }
      }
      if (lifecycle === 'reload') await page.reload();
      if (lifecycle === 'relogin') {
        await navigate(page, '退出'); await page.locator('#login').waitFor({state: 'visible'});
        await keyboardInput(page, page.getByLabel('管理员密码'), process.env.RELAY_BROWSER_PASSWORD);
        await keyboardButton(page, '登录');
      }
      if (lifecycleOverview) {
        const response = await lifecycleOverview;
        assert.equal(response.status(), 200);
        currentEvidence = await evidenceSummary(page, await response.json(), evidenceID);
      }
      await visibleText(page, '#overview-evidence', evidenceID); await visibleText(page, '#summary', '模型 alpha-' + width);
      const exact = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/requests/' + evidenceID);
      if (!await page.locator('#completed-setup').evaluate(node => node.open)) {
        await keyboardFocus(page, page.locator('#completed-setup summary')); await page.keyboard.press('Enter');
      }
      await keyboardButton(page, '查看成功请求详情');
      const exactResponse = await exact;
      assert.equal(exactResponse.status(), 200);
      const record = await exactResponse.json();
      assert.equal(record.id, currentEvidence.request_id);
      assert.equal(currentEvidence.time, record.time, 'displayed call time belongs to the exact successful request');
      assert.equal(currentEvidence.upstream_revision, record.revision, 'displayed configuration version belongs to the exact successful request');
      await visibleText(page, '#record-detail', evidenceID);
      assert.equal(await page.getByLabel('精确请求 ID').inputValue(), evidenceID);
      currentEvidence = await evidenceOverview(page, evidenceID);
    }
  } finally { page.off('request', countExact); }
  console.log(`PASS QA recovery, precise older evidence, started/unknown/404/injected503, native pagination ${width}px (controlled upstream; injected failures explicitly scoped)`);
}
async function sameRecord(page, id, alias, keyID, state, usage) {
  const exact = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/requests/' + id);
  await keyboardButton(page, '查看本次记录');
  const response = await exact; assert.equal(response.status(), 200);
  const record = await response.json();
  assert.equal(record.id, id); assert.equal(record.alias, alias); assert.equal(record.key_id, keyID); assert.equal(record.state, state);
  assert.equal(typeof record.duration_ms === 'number' && record.duration_ms >= 0, true, 'completed request has actual duration');
  assert.equal([record.prompt_tokens, record.completion_tokens, record.total_tokens].map(value => value == null ? '未知' : String(value)).join(' / '), usage);
  await visibleText(page, '#record-detail', id);
  const text = await page.locator('#record-detail').textContent();
  for (const expected of [alias, keyID, state, usage]) assert.equal(text.includes(expected), true, 'same request safe metadata');
  if (record.model) assert.equal(text.includes(alias + ' → ' + record.model), true, 'visible mapping equals exact response');
  assert.equal(text.split('\n').includes(String(record.duration_ms)), true, 'visible duration equals exact response');
  assert.equal(await page.getByLabel('精确请求 ID').inputValue(), id);
  assert.equal(await page.getByLabel('应用密钥', {exact: true}).inputValue(), '');
  assert.equal(await page.locator('#test-output').textContent(), '');
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
}
async function passwordState(page, visible, cleared = false, stage = null) {
  if (stage) passwordProbe = {stage};
  checkpoint = visible ? 'password shown with accessible action and state' : 'password hidden with accessible action and state';
  passwordProbe = {...passwordProbe, expectedVisible: visible, ...await page.evaluate(() => {
    const input = document.getElementById('password');
    return {
      type: ['password', 'text'].includes(input.type) ? input.type : 'other',
      inputFocused: input === document.activeElement,
      toggleFocused: document.getElementById('password-toggle') === document.activeElement,
      events: window.passwordEvents ? {...window.passwordEvents} : null,
    };
  })};
  const toggle = page.locator('#password-toggle');
  const name = visible ? '隐藏密码' : '显示密码';
  const status = visible ? '密码已显示' : '密码已隐藏';
  assert.equal(await page.locator('#password').getAttribute('type'), visible ? 'text' : 'password');
  assert.equal(await toggle.textContent(), name);
  assert.equal(await page.getByRole('button', {name, exact: true}).count(), 1);
  assert.equal((await toggle.ariaSnapshot()).includes('button "' + name + '"'), true);
  assert.equal(await toggle.getAttribute('type'), 'button');
  assert.equal(await toggle.getAttribute('aria-controls'), 'password');
  assert.equal(await toggle.getAttribute('aria-describedby'), 'password-visibility-status');
  assert.equal(await toggle.getAttribute('aria-pressed'), null);
  assert.equal(await page.locator('#password-visibility-status').getAttribute('role'), 'status');
  assert.equal(await page.locator('#password-visibility-status').getAttribute('aria-live'), 'polite');
  assert.equal(await page.locator('#password-visibility-status').textContent(), status);
  assert.equal((await page.locator('#password-visibility-status').ariaSnapshot()).includes(status), true);
  if (cleared) assert.equal((await page.locator('#password').inputValue()) === '', true, 'password cleared');
}
async function passwordVisibilityJourney(page, context, requests) {
  passwordProbe = {stage: 'initial login'};
  await page.locator('#login').waitFor({state: 'visible'});
  await passwordState(page, false, true);
  await page.reload();
  passwordProbe = {stage: 'reloaded login'};
  await page.locator('#login').waitFor({state: 'visible'});
  await passwordState(page, false, true);
  const before = requests.count;
  const url = page.url();
  const storage = JSON.stringify(await context.storageState());
  await page.evaluate(() => {
    window.passwordEvents = {submit: 0, invalid: 0, toggle: 0};
    document.getElementById('login-form').addEventListener('submit', () => window.passwordEvents.submit++);
    document.getElementById('login-form').addEventListener('invalid', () => window.passwordEvents.invalid++, true);
    document.getElementById('password-toggle').addEventListener('click', () => window.passwordEvents.toggle++);
  });
  const input = page.getByLabel('管理员密码');
  const toggle = page.locator('#password-toggle');
  let activationCount = 0;
  for (const value of ['', '  controlled <&> 密码 !  ']) {
    await input.fill(value);
    await input.focus();
    await page.keyboard.press('Tab');
    assert.equal(await toggle.evaluate(node => node === document.activeElement), true);
    assert.equal(await toggle.evaluate(node => getComputedStyle(node).outlineStyle !== 'none' && parseFloat(getComputedStyle(node).outlineWidth) >= 3), true);
    for (const activation of ['Enter', 'Space', 'click', 'click']) {
      const visible = (await input.getAttribute('type')) === 'password';
      passwordProbe = {stage: value === '' ? 'empty input' : 'controlled input', activation, activationCount: ++activationCount, expectedVisible: visible};
      if (activation === 'click') await toggle.click();
      else await page.keyboard.press(activation);
      try {
        await page.waitForFunction(expected => window.passwordEvents.toggle >= expected, activationCount, {timeout: 2000});
      } finally {
        await passwordState(page, visible);
      }
      assert.equal(await page.evaluate(() => window.passwordEvents.toggle), activationCount, 'one native click per activation');
      assert.equal((await input.inputValue()) === value, true, 'toggle preserves original value');
      assert.equal(await toggle.evaluate(node => node === document.activeElement), true);
    }
    await page.keyboard.press('Tab');
    assert.equal(await page.getByRole('button', {name: '登录', exact: true}).evaluate(node => node === document.activeElement), true);
  }
  await page.waitForTimeout(50);
  checkpoint = 'password toggles have zero requests, submit and invalid events or storage writes';
  assert.equal(requests.count, before);
  assert.deepEqual(await page.evaluate(() => window.passwordEvents), {submit: 0, invalid: 0, toggle: activationCount});
  assert.equal(page.url() === url, true);
  assert.equal(JSON.stringify(await context.storageState()) === storage, true);
  assert.equal(await page.evaluate(() => localStorage.length + sessionStorage.length), 0);
  assert.deepEqual(await page.evaluate(async () => indexedDB.databases()), []);
  assert.equal(await page.evaluate(async () => (await navigator.serviceWorker.getRegistrations()).length), 0);
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
  assert.equal(await toggle.evaluate(node => node.getBoundingClientRect().height >= 44), true);
  await input.fill('');
  passwordProbe = null;
}
async function rejectedLoginJourney(page) {
  let releaseLogin;
  const released = new Promise(resolve => { releaseLogin = resolve; });
  const started = page.waitForRequest(request => new URL(request.url()).pathname === '/admin/api/login' && request.method() === 'POST');
  const completed = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/login' && response.request().method() === 'POST');
  const delayed = async route => { await released; await route.continue(); };
  await page.route('**/admin/api/login', delayed);
  try {
    passwordProbe = {stage: 'rejected login pending'};
    await page.getByLabel('管理员密码').fill('wrong fixture password');
    await navigate(page, '显示密码');
    await navigate(page, '登录');
    await started;
    await passwordState(page, true);
    assert.equal(await page.getByRole('button', {name: '登录', exact: true}).isDisabled(), true);
    assert.equal((await page.getByLabel('管理员密码').inputValue()) === 'wrong fixture password', true, 'pending rejection retains input');
    await visibleText(page, '#error', '管理会话已过期，请重新登录；不会自动重发调用。');
    passwordProbe = {stage: 'rejected login response'};
    releaseLogin();
    assert.equal((await completed).status(), 401);
    await visibleText(page, '#error', '管理员密码不正确，请重新输入。');
    await passwordState(page, false, true);
  } finally {
    releaseLogin();
    await page.unroute('**/admin/api/login', delayed);
  }
}
async function loadUpstreamPanel(page) {
  let releaseLoad;
  const loadReleased = new Promise(resolve => { releaseLoad = resolve; });
  const loadStarted = page.waitForRequest('**/admin/api/upstreams**', {timeout: 10000});
  const delayedLoad = async route => { await loadReleased; await route.continue(); };
  await page.route('**/admin/api/upstreams**', delayedLoad);
  try {
    await navigate(page, '上游');
    checkpoint = 'upstream load barrier';
    await loadStarted;
    assert.equal(await page.locator('#upstream').isVisible(), false);
    releaseLoad();
    checkpoint = 'upstream panel loaded';
    await page.getByLabel('上游地址').waitFor({state: 'visible', timeout: 10000});
  } finally {
    releaseLoad();
    await page.unroute('**/admin/api/upstreams**', delayedLoad);
  }
}
async function runtimeVersionJourney(page, context, width, healthRequests) {
  checkpoint = 'runtime version actual Go health consistency';
  const actual = await context.request.get(origin + '/healthz');
  assert.equal(actual.status(), 200);
  const {version} = await actual.json();
  assert.equal(version, 'dev');
  await visibleText(page, '#runtime-version', version);
  assert.equal(await page.locator('#runtime-version').textContent(), version);
  assert.equal(await page.locator('#runtime-version').getAttribute('role'), 'status');
  assert.equal(await page.locator('#runtime-version').getAttribute('aria-live'), 'polite');
  const beforeReload = healthRequests.count;
  await page.reload();
  await visibleText(page, '#runtime-version', version);
  assert.equal(healthRequests.count, beforeReload + 1, 'refresh performs exactly one health request');

  checkpoint = 'runtime version failure and navigation recovery';
  const unavailable = route => route.fulfill({status: 401, contentType: 'application/json', body: '{}'});
  await page.route('**/healthz', unavailable);
  await navigate(page, '模型');
  assert.equal(await page.locator('#runtime-version').isVisible(), false);
  await navigate(page, '概览');
  await visibleText(page, '#runtime-version', '版本未知');
  await page.locator('#summary').waitFor({state: 'visible'});
  assert.equal(await page.locator('#console').isVisible(), true);
  assert.equal(await page.locator('#error').textContent(), '');
  const beforeFailureReload = healthRequests.count;
  await page.reload();
  await visibleText(page, '#runtime-version', '版本未知');
  assert.equal(healthRequests.count, beforeFailureReload + 1);
  await page.unroute('**/healthz', unavailable);
  await navigate(page, '模型');
  await navigate(page, '概览');
  await visibleText(page, '#runtime-version', version);
  await page.reload();
  await visibleText(page, '#runtime-version', version);

  checkpoint = 'runtime version pending health does not block management';
  let releaseHealth;
  const released = new Promise(resolve => { releaseHealth = resolve; });
  const started = page.waitForRequest('**/healthz');
  const delayed = async route => { await released; await route.fulfill({status: 200, contentType: 'application/json', body: JSON.stringify({version: 'late-version'})}).catch(() => {}); };
  await page.route('**/healthz', delayed);
  try {
    await navigate(page, '概览');
    await started;
    await visibleText(page, '#runtime-version', '加载中');
    await page.locator('#summary').waitFor({state: 'visible'});
    await navigate(page, '模型');
    await page.locator('#models').waitFor({state: 'visible'});
    assert.equal(await page.locator('#runtime-version').isVisible(), false);
  } finally {
    releaseHealth();
    await page.unroute('**/healthz', delayed);
  }

  checkpoint = 'runtime version safe text and long value layout';
  const literal = '  <img src=x onerror="window.versionInjected=true"> & ' + 'long-version-'.repeat(50) + '  ';
  const textResponse = route => route.fulfill({status: 200, contentType: 'application/json', body: JSON.stringify({version: literal})});
  await page.route('**/healthz', textResponse);
  await navigate(page, '概览');
  await page.waitForFunction(expected => document.getElementById('runtime-version').textContent === expected, literal);
  assert.equal(await page.locator('#runtime-version > *').count(), 0);
  assert.equal(await page.evaluate(() => window.versionInjected), undefined);
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
  await page.unroute('**/healthz', textResponse);
  await navigate(page, '模型');
  await navigate(page, '概览');
  await visibleText(page, '#runtime-version', version);
  console.log(`PASS runtime version actual health, refresh, failure/reentry, nonblocking and text/layout ${width}px`);
}
try {
  console.log('model_diagnostic_execution=' + process.pid);
  browser = await chromium.launch({headless: true});
  for (const width of [1280, 390]) {
    viewport = width;
    responses.length = 0;
    const context = await browser.newContext({viewport: {width, height: 900}});
    context.on('page', trackModelReads);
    const page = await context.newPage();
    const healthRequests = {count: 0};
    const allRequests = {count: 0};
    const generations = {count: 0};
    page.on('request', () => allRequests.count++);
    page.on('request', request => { if (request.method() === 'POST' && new URL(request.url()).pathname === '/v1/chat/completions') generations.count++; });
    page.on('request', request => { if (new URL(request.url()).pathname === '/healthz') healthRequests.count++; });
    page.on('dialog', dialog => dialog.accept());
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('response', response => {
      const pathname = new URL(response.url()).pathname;
      if (!['/admin/api/overview', '/admin/api/session', '/admin/api/login', '/admin/api/upstreams', '/admin/api/models', '/admin/api/keys', '/admin/api/requests', '/admin/api/logout', '/v1/chat/completions'].includes(pathname)) return;
      responses.push({method: response.request().method(), path: pathname, status: response.status()});
      if (responses.length > 12) responses.shift();
    });
    checkpoint = 'open login';
    await page.goto(origin + '/admin/');
    assert.equal(await page.locator('#runtime-version').isVisible(), false);
    assert.equal(healthRequests.count, 0, 'no health request before login');
    await passwordVisibilityJourney(page, context, allRequests);
    await rejectedLoginJourney(page);
    await page.getByLabel('管理员密码').fill(process.env.RELAY_BROWSER_PASSWORD);
    checkpoint = 'hidden password Enter submits normal login';
    const hiddenLoginOverview = nextOverview(page);
    await page.getByLabel('管理员密码').press('Enter');
    await loginOverview(page, hiddenLoginOverview, width);
    if (width === 1280) {
      await visibleText(page, '#overview-models', '0 个 / 启用 0');
      await visibleText(page, '#overview-keys', '0 个 / 启用 0');
      await navigate(page, '调用测试'); await visibleText(page, '#test-dependency', '调用依赖尚缺');
      assert.equal(await page.locator('#send').isDisabled(), true);
      await navigate(page, '概览');
    }
    assert.equal((await page.locator('#password').inputValue()) === '', true);
    assert.equal(await page.locator('#password').getAttribute('type'), 'password');
    await navigate(page, '退出');
    await page.locator('#login').waitFor({state: 'visible'});
    await passwordState(page, false, true, 'logged out');
    await page.reload();
    await page.locator('#login').waitFor({state: 'visible'});
    await passwordState(page, false, true, 'reloaded after logout');
    await page.getByLabel('管理员密码').fill(process.env.RELAY_BROWSER_PASSWORD);
    await navigate(page, '显示密码');
    await passwordState(page, true, false, 'visible password login');
    const visibleLoginOverview = nextOverview(page);
    await navigate(page, '登录');
    await loginOverview(page, visibleLoginOverview, width);
    assert.equal((await page.locator('#password').inputValue()) === '', true);
    assert.equal(await page.locator('#password').getAttribute('type'), 'password');
    await runtimeVersionJourney(page, context, width, healthRequests);
    assert.equal((await page.locator('#password').inputValue()) === '', true, 'session restoration clears password');
    assert.equal(await page.locator('#password').getAttribute('type'), 'password');
    await loadUpstreamPanel(page);
    await keyboardInput(page, page.getByLabel('上游名称', {exact: true}), '受控主上游');
    await keyboardInput(page, page.getByLabel('上游地址'), process.env.RELAY_BROWSER_UPSTREAM);
    await keyboardInput(page, page.getByLabel('上游凭据'), 'controlled-browser-upstream');
    await keyboardCheck(page, page.getByLabel('启用上游'), true);
    await keyboardButton(page, '保存上游');
    await visibleText(page, '#notice', '上游已保存');
    await upstreamFailureJourney(page);
    await page.getByLabel('上游地址').fill('http://127.0.0.1:1/v1');
    const invalidSave = page.waitForResponse(response => new URL(response.url()).pathname.startsWith('/admin/api/upstreams') && ['POST', 'PATCH'].includes(response.request().method()));
    await navigate(page, '保存上游');
    assert.equal((await invalidSave).status(), 400, 'invalid address remains rejected after settled reread');
    await visibleText(page, '#error', '上游地址不符合');
    assert.equal(await page.locator('#notice').textContent(), '');
    await navigate(page, '概览');
    await loadUpstreamPanel(page);
    checkpoint = 'rejected save restores persisted upstream and hides credential';
    assert.equal(await page.getByLabel('上游地址').inputValue(), process.env.RELAY_BROWSER_UPSTREAM);
    assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
    await navigate(page, '模型');
    await keyboardExpand(page, '#model-create');
    const canceledModelWrites = {count: 0};
    const countModelWrites = request => { if (new URL(request.url()).pathname === '/admin/api/models' && request.method() === 'POST') canceledModelWrites.count++; };
    page.on('request', countModelWrites);
    await keyboardInput(page, page.getByLabel('模型别名', {exact: true}), 'canceled-draft-' + width);
    await keyboardButton(page, '取消创建模型');
    assert.equal(canceledModelWrites.count, 0, 'keyboard create cancel makes no object');
    page.off('request', countModelWrites);
    for (const alias of ['alpha-' + width, 'beta-' + width]) {
      await keyboardExpand(page, '#model-create');
      await keyboardInput(page, page.getByLabel('模型别名', {exact: true}), alias);
      await keyboardInput(page, page.locator('#model-create').getByLabel('远端模型名称', {exact: true}), alias.startsWith('alpha-') ? 'controlled-model' : 'controlled-beta-model');
      await keyboardButton(page, '创建模型');
      await visibleText(page, '#model-list', alias);
    }
    await expand(page, '#model-create');
    await navigate(page, '创建模型');
    await visibleText(page, '#error', '该模型别名已存在');
    page.removeAllListeners('dialog');
    page.on('dialog', dialog => dialog.type() === 'prompt' ? dialog.accept('edited-controlled-model') : dialog.accept());
    await keyboardButton(page, '编辑映射 alpha-' + width);
    const writesBeforeCancel = responses.filter(value => value.method === 'PATCH').length;
    await keyboardInput(page, page.locator('#model-edit').getByLabel('远端模型名称', {exact: true}), 'discarded-edit');
    await keyboardAction(page, page.getByRole('button', {name: '取消编辑', exact: true}));
    assert.equal(responses.filter(value => value.method === 'PATCH').length, writesBeforeCancel);
    await visibleText(page, '#model-list', 'alpha-' + width + ' → controlled-model');
    await keyboardButton(page, '编辑映射 alpha-' + width);
    await keyboardInput(page, page.locator('#model-edit').getByLabel('远端模型名称', {exact: true}), 'edited-controlled-model');
    await keyboardButton(page, '保存映射');
    await visibleText(page, '#notice', '映射已保存');
    await visibleText(page, '#model-list', 'alpha-' + width + ' → edited-controlled-model');
    await navigate(page, '应用密钥');
    await keyboardExpand(page, '#key-create');
    const canceledKeyWrites = {count: 0};
    const countKeyWrites = request => { if (new URL(request.url()).pathname === '/admin/api/keys' && request.method() === 'POST') canceledKeyWrites.count++; };
    page.on('request', countKeyWrites);
    await keyboardInput(page, page.getByLabel('密钥名称'), 'canceled-key-draft-' + width);
    await keyboardButton(page, '取消签发');
    assert.equal(canceledKeyWrites.count, 0, 'keyboard signing cancel makes no object');
    page.off('request', countKeyWrites);
    const keys = [];
    const keyIDs = [];
    for (const name of ['first-' + width, 'second-' + width]) {
      await keyboardExpand(page, '#key-create');
      await keyboardInput(page, page.getByLabel('密钥名称'), name);
      const keysRead = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/keys' && response.request().method() === 'GET');
      await keyboardButton(page, '签发密钥');
      checkpoint = 'visible #issued';
      await page.locator('#issued').waitFor({state: 'visible'});
      const key = await page.getByLabel('新密钥（仅本次）').inputValue();
      assert.equal(/^mr_/.test(key), true, 'controlled key shape');
      keys.push(key);
      await visibleText(page, '#key-list', name);
      const row = page.locator('#key-list .item').filter({hasText: name});
      keyIDs.push((await row.textContent()).match(/标识 ([A-Za-z0-9_-]+)/)[1]);
      const keyResponse = await keysRead; assert.equal(keyResponse.status(), 200);
      const persistedKey = (await keyResponse.json()).find(value => value.id === keyIDs.at(-1));
      assert.equal(typeof persistedKey.created === 'string' && persistedKey.created.length > 0, true);
      assert.equal((await row.textContent()).includes('创建时间 ' + persistedKey.created), true, 'key creation time comes from same-ID Go record');
      assert.equal(await page.locator('#issued-key').getAttribute('type'), 'password');
      await copyIssuedKey(page, context);
      await keyboardButton(page, '关闭展示');
      assert.equal(await page.getByLabel('新密钥（仅本次）').inputValue(), '');
    }
    checkpoint = 'reload';
    await page.reload();
    await page.locator('#keys').waitFor({state: 'visible'});
    assert.equal(new URL(page.url()).hash, '#keys');
    await navigate(page, '概览');
    await visibleText(page, '#summary', '配置已保存');
    await navigate(page, '模型');
    await visibleText(page, '#model-list', 'alpha-' + width + ' → edited-controlled-model');
    await visibleText(page, '#model-list', 'beta-' + width + ' → controlled-beta-model');
    checkpoint = 'graceful Go service shutdown and store reopen';
    const restarted = await context.request.post(process.env.RELAY_BROWSER_CONTROL + '/restart');
    assert.equal(restarted.status(), 204);
    const expiredSession = await context.request.get(origin + '/admin/api/models');
    assert.equal(expiredSession.status(), 401);
    await page.reload();
    checkpoint = 'login after actual service restart';
    await page.getByLabel('管理员密码').waitFor({state: 'visible', timeout: 10000});
    await page.getByLabel('管理员密码').fill(process.env.RELAY_BROWSER_PASSWORD);
    await navigate(page, '登录');
    await page.locator('#models').waitFor({state: 'visible'});
    assert.equal(new URL(page.url()).hash, '#models');
    await navigate(page, '概览');
    await visibleText(page, '#summary', '配置已保存');
    const restoredModels = await context.request.get(origin + '/admin/api/models');
    assert.equal(restoredModels.status(), 200);
    const mappings = (await restoredModels.json()).data;
    checkpoint = 'two persisted mappings from real management API';
    assert.equal(mappings.some(value => value.alias === 'alpha-' + width && value.candidates.some(candidate => candidate.remote_model === 'edited-controlled-model') && value.enabled), true);
    assert.equal(mappings.some(value => value.alias === 'beta-' + width && value.candidates.some(candidate => candidate.remote_model === 'controlled-beta-model') && value.enabled), true);
    await navigate(page, '模型');
    await visibleText(page, '#model-list', 'alpha-' + width + ' → edited-controlled-model');
    await visibleText(page, '#model-list', 'beta-' + width + ' → controlled-beta-model');
    await loadUpstreamPanel(page);
    checkpoint = 'upstream configuration and credential state survive restart';
    assert.equal(await page.getByLabel('上游地址').inputValue(), process.env.RELAY_BROWSER_UPSTREAM);
    assert.equal(await page.getByLabel('启用上游').isChecked(), true);
    assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
    await visibleText(page, '#credential-state', '凭据已配置');
    await navigate(page, '应用密钥');
    assert.equal(await page.locator('#issued').isVisible(), false);
    await navigate(page, '调用测试');
    async function fillCall(key, text, stream) {
      await keyboardInput(page, page.getByLabel('应用密钥', {exact: true}), key);
      await keyboardModel(page, 'alpha-' + width);
      await keyboardInput(page, page.getByLabel('文本消息'), text);
      await keyboardCheck(page, page.getByLabel('SSE 流式'), stream);
    }
    await dependencyJourney(page, width, keys, keyIDs, fillCall);
    await keyboardAction(page, page.getByText('客户端调用示例（仅占位符）', {exact: true}));
    const beforeCopy = generations.count;
    await keyboardAction(page, page.locator('#copy-example'));
    await page.waitForFunction(() => document.getElementById('notice').textContent.includes('示例已复制') || document.getElementById('error').textContent.includes('手动选择调用示例'));
    assert.equal(generations.count, beforeCopy, 'example copy never generates');
    const example = await page.locator('#client-example').textContent();
    assert.equal(example.includes('<APPLICATION_KEY>') && example.includes(origin + '/v1'), true);
    assert.equal(keys.some(key => example.includes(key)), false);
    await fillCall(keys[0], 'browser-prompt-sentinel', false);
    let releaseOrdinary;
    const ordinaryGate = new Promise(resolve => { releaseOrdinary = resolve; });
    const heldOrdinary = async route => { await ordinaryGate; await route.continue(); };
    await page.route('**/v1/chat/completions', heldOrdinary);
    const ordinaryStarted = page.waitForRequest(request => new URL(request.url()).pathname === '/v1/chat/completions' && request.method() === 'POST');
    try {
      await navigate(page, '发送请求'); await ordinaryStarted;
      await submittedValuesVisible(page, keys[0], 'browser-prompt-sentinel', false, 'alpha-' + width);
      releaseOrdinary();
      await visibleText(page, '#test-status', '完成');
      await callControlState(page, false);
    } finally { releaseOrdinary(); await page.unroute('**/v1/chat/completions', heldOrdinary); }
    await visibleText(page, '#test-status', '完成');
    await visibleText(page, '#test-output', 'controlled ordinary');
    assert.equal(await page.locator('#test-output').textContent(), 'controlled ordinary');
    const ordinaryID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
    await keyboardAction(page, page.locator('#diagnostics summary'));
    await visibleText(page, '#test-raw', 'prompt_tokens');
    await sameRecord(page, ordinaryID, 'alpha-' + width, keyIDs[0], 'success', '0 / 2 / 2');
    await evidenceOverview(page, ordinaryID);
    await navigate(page, '调用测试');
    await fillCall(keys[0], 'stream fixture', true);
    const beforeStream = generations.count;
    await keyboardAction(page, page.getByRole('button', {name: '发送请求', exact: true}));
    await page.locator('#test-form').evaluate(form => { form.requestSubmit(); form.requestSubmit(); });
    await visibleText(page, '#test-output', 'controlled first');
    await visibleText(page, '#test-status', '接收中');
    await submittedValuesVisible(page, keys[0], 'stream fixture', true, 'alpha-' + width);
    assert.equal(generations.count, beforeStream + 1, 'synchronous repeat-submit isolation');
    const released = await context.request.post(process.env.RELAY_BROWSER_CONTROL + '/release-stream');
    assert.equal(released.status(), 204);
    await visibleText(page, '#test-status', '完成');
    await callControlState(page, false);
    await visibleText(page, '#test-output', 'controlled first');
    const streamID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
    await sameRecord(page, streamID, 'alpha-' + width, keyIDs[0], 'success', '0 / 2 / 2');
    await navigate(page, '调用测试');
    await fillCall(keys[0], 'cancel fixture', true);
    await navigate(page, '发送请求');
    await visibleText(page, '#test-output', 'controlled first');
    await submittedValuesVisible(page, keys[0], 'cancel fixture', true, 'alpha-' + width);
    await keyboardAction(page, page.getByRole('button', {name: '取消请求', exact: true}), 'Space');
    await visibleText(page, '#test-status', '已取消');
    await callControlState(page, false);
    assert.equal(await page.getByLabel('应用密钥', {exact: true}).inputValue(), '');
    for (const streaming of [false, true]) {
      await fillCall(keys[0], 'tool fixture', streaming);
      const beforeTools = generations.count;
      await navigate(page, '发送请求');
      if (streaming) {
        await visibleText(page, '#test-status', '接收中');
        assert.equal(await page.locator('#test-output').textContent(), '');
        const released = await page.request.post(process.env.RELAY_BROWSER_CONTROL + '/release-stream');
        assert.equal(released.status(), 204);
      }
      await visibleText(page, '#test-status', '完成');
      await visibleText(page, '#test-output', '返回1个工具调用：lookup_weather');
      await visibleText(page, '#test-output', '未执行工具');
      assert.equal(generations.count, beforeTools + 1);
      const toolID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
      await sameRecord(page, toolID, 'alpha-' + width, keyIDs[0], 'success', '未知 / 未知 / 未知');
      await navigate(page, '调用测试');
    }
    await fillCall(keys[0], 'cancel fixture', true);
    await navigate(page, '发送请求');
    await visibleText(page, '#test-output', 'controlled first');
    const beforeCheckFailure = generations.count;
    await page.route('**/admin/api/session', route => route.abort('failed'));
    await visibleText(page, '#test-status', '接收中');
    await page.locator('#error').filter({hasText: '管理会话暂无法确认'}).waitFor({state: 'visible', timeout: 35000});
    for (const selector of ['#test-key', '#test-message']) assert.equal(await page.locator(selector).inputValue(), '');
    for (const selector of ['#test-output', '#test-raw']) assert.equal(await page.locator(selector).textContent(), '');
    assert.equal(await page.locator('#send').isDisabled(), true);
    assert.equal(await page.locator('#console').isVisible(), true);
    assert.equal(await page.evaluate(() => location.hash), '#test');
    assert.equal(await page.locator('#test-model').inputValue(), 'alpha-' + width);
    await page.unroute('**/admin/api/session');
    await navigate(page, '重新检查会话');
    await visibleText(page, '#notice', '管理会话已确认');
    assert.equal(await page.locator('#send').isDisabled(), false);
    assert.equal(generations.count, beforeCheckFailure);
    await fillCall(keys[0], 'browser-error-prompt-sentinel', false);
    await navigate(page, '发送请求');
    await visibleText(page, '#error', '供应商拒绝调用');
    await visibleText(page, '#test-status', '失败');
    await callControlState(page, false);
    assert.equal(await callInputsCleared(page), true, 'provider failure clears inputs without discarding safe diagnostics');
    const rejectedID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
    await sameRecord(page, rejectedID, 'alpha-' + width, keyIDs[0], 'upstream_rejected', '未知 / 未知 / 未知');
    await navigate(page, '调用记录');
    await visibleText(page, '#record-list', 'success');
    for (let attempt = 0; attempt < 30 && !(await page.locator('#record-list').textContent()).includes('canceled'); attempt++) {
      await page.waitForTimeout(100);
      await navigate(page, '刷新记录');
    }
    await visibleText(page, '#record-list', 'canceled');
    await visibleText(page, '#record-list', 'upstream_rejected');
    checkpoint = 'sensitive failure classification belongs to current viewport alias';
    const rejectedRow = page.locator('#record-list tr').filter({hasText: 'alpha-' + width}).filter({hasText: 'upstream_rejected'});
    await rejectedRow.waitFor({state: 'visible', timeout: 10000});
    assert.equal(await rejectedRow.count(), 1);
    const secrets = [process.env.RELAY_BROWSER_PASSWORD, 'wrong fixture password', '  controlled <&> 密码 !  ', 'controlled after logout input', 'controlled-browser-upstream', ...keys, 'browser-prompt-sentinel', 'browser-error-prompt-sentinel', 'controlled ordinary', 'controlled first', 'stream fixture', 'cancel fixture', 'browser-sensitive-diagnostic-sentinel'];
    checkpoint = 'record-list excludes credentials, conversations and sensitive upstream diagnostics';
    const recordText = await page.locator('#record-list').textContent();
    for (const secret of secrets) assert.equal(recordText.includes(secret), false, 'record-list exclusion failed');
    checkpoint = 'captured actual Go service logs exclude credentials, conversations and upstream diagnostics';
    const logAudit = await context.request.post(process.env.RELAY_BROWSER_CONTROL + '/audit-logs', {data: secrets});
    assert.equal(logAudit.status(), 204, 'service log exclusion failed');
    assert.equal(Number(logAudit.headers()['x-controlled-sensitive-failures']), width === 1280 ? 1 : 2, 'sensitive upstream failure was not exercised exactly once per viewport');
    await navigate(page, '应用密钥');
    await navigate(page, '停用密钥 first-' + width);
    await visibleText(page, '#key-list', '停用');
    await navigate(page, '调用测试');
    await fillCall(keys[0], 'disabled fixture', false);
    await navigate(page, '发送请求');
    await visibleText(page, '#error', '请');
    const disabledID = (await page.locator('#test-request-id').textContent()).replace('请求 ID：', '');
    assert.equal(await page.locator('#console').isVisible(), true, 'application 401 must not sign out administrator');
    await sameRecord(page, disabledID, 'alpha-' + width, keyIDs[0], 'rejected', '未知 / 未知 / 未知');
    await navigate(page, '调用测试');
    await fillCall(keys[1], 'second fixture', false);
    await navigate(page, '发送请求');
    await visibleText(page, '#test-status', '完成');
    await navigate(page, '调用记录');
    await visibleText(page, '#record-list', 'upstream_rejected');
    checkpoint = 'final record-list and service logs exclude all journey secrets and conversations';
    const finalSecrets = [...secrets, 'disabled fixture', 'second fixture'];
    const finalRecords = await page.locator('#record-list').textContent();
    for (const secret of finalSecrets) assert.equal(finalRecords.includes(secret), false, 'record-list exclusion failed');
    await recordsBoundaryJourney(page, context, width, keys[1], keyIDs[1], fillCall);
    finalSecrets.push('controlled-rejected-secret', 'controlled-network-secret', 'controlled-restore-secret', 'mr_unrecognized_pagination_fixture', 'pagination rejection fixture', 'dependency-rejected-fixture', 'dependency-restored-fixture', 'model rejection fixture', 'protocol fixture', 'interrupt fixture', 'network failure fixture', 'unknown usage fixture', 'controlled partial', 'controlled unknown usage');
    await navigate(page, '调用记录'); await page.locator('#records').waitFor({state: 'visible'});
    for (const secret of finalSecrets) assert.equal((await page.locator('#record-list').textContent()).includes(secret), false, 'extended record-list exclusion failed');
    const finalLogAudit = await context.request.post(process.env.RELAY_BROWSER_CONTROL + '/audit-logs', {data: finalSecrets});
    assert.equal(finalLogAudit.status(), 204, 'service log exclusion failed');
    assert.equal(Number(finalLogAudit.headers()['x-controlled-sensitive-failures']), width === 1280 ? 1 : 2, 'sensitive failure count changed unexpectedly');
    await navigate(page, '概览');
    for (const view of ['upstream', 'models', 'keys', 'test', 'records', 'overview']) {
      await page.goto(origin + '/admin/#' + view);
      await page.locator('#' + view).waitFor({state: 'visible'});
      assert.equal(await page.locator('nav button[aria-current="page"]').getAttribute('data-view'), view);
      await page.reload(); await page.locator('#' + view).waitFor({state: 'visible'});
      assert.equal(new URL(page.url()).hash, '#' + view);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
    }
    await keyboardAction(page, page.locator('nav [data-view="models"]'));
    await page.locator('#models').waitFor({state: 'visible'});
    await page.goBack(); await page.locator('#overview').waitFor({state: 'visible'});
    await page.goForward(); await page.locator('#models').waitFor({state: 'visible'});
    await navigate(page, '概览');
    const storage = await context.storageState();
    assert.deepEqual(storage.origins, []);
    assert.ok(storage.cookies.every(cookie => cookie.httpOnly && cookie.sameSite === 'Strict' && cookie.path === '/admin'));
    assert.equal(await page.evaluate(() => localStorage.length + sessionStorage.length), 0);
    assert.deepEqual(await page.evaluate(async () => indexedDB.databases()), []);
    assert.equal(await page.evaluate(async () => (await navigator.serviceWorker.getRegistrations()).length), 0);
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
    assert.equal(await page.evaluate(() => ['password', 'upstream-token', 'issued-key', 'test-key', 'test-message'].every(id => document.getElementById(id).value === '') && document.getElementById('test-output').textContent === '' && document.getElementById('test-raw').textContent === ''), true, 'screenshots only after secret clearing');
    await mkdir('test-results', {recursive: true});
    await page.screenshot({path: `test-results/console-first-call-${width}.png`, fullPage: true});
    assert.deepEqual(errors, []);
    await navigate(page, '退出');
    checkpoint = 'visible #login after logout';
    await page.locator('#login').waitFor({state: 'visible'});
    await passwordState(page, false, true, 'final logout');
    await page.getByLabel('管理员密码').fill('controlled after logout input');
    await passwordState(page, false, false, 'input after final logout');
    await page.reload();
    await page.locator('#login').waitFor({state: 'visible'});
    await passwordState(page, false, true, 'final login reload');
    assert.equal(await page.locator('#runtime-version').isVisible(), false);
    assert.equal(await page.locator('#runtime-version').textContent(), '');
    const response = await context.request.get(origin + '/admin/api/keys');
    assert.equal(response.status(), 401);
    await context.close();
    console.log(`PASS actual Go service browser journey ${width}px (controlled upstream, not supplier validation)`);
  }
} catch (error) {
  const location = String(error.stack).match(/journey\.mjs:\d+:\d+/)?.[0] || 'unavailable';
  console.error(serializeFailure({viewport, checkpoint, location, error: error.name, passwordProbe: checkpoint.startsWith('password ') ? passwordProbe : null, keyboardProbe: checkpoint.startsWith('native keyboard ') ? keyboardProbe : null, clipboardProbe, modelProbe: modelFailureDiagnostic(modelProbe), modelDiagnosticExecution: process.pid, modelDiagnosticFile: modelProbe ? modelDiagnosticFile : null, responses}));
  process.exitCode = 1;
} finally {
  if (browser) await browser.close();
}
