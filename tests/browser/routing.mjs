import assert from 'node:assert/strict';
import { chromium } from 'playwright';
import { mkdir, appendFile } from 'node:fs/promises';
import { observeSelection } from './selection-trace.mjs';
import { typeAheadCandidate } from './candidate-keyboard.mjs';

const browser = await chromium.launch({headless: true});
const origin = process.env.RELAY_BROWSER_URL;
const controlled = new URL(process.env.RELAY_BROWSER_UPSTREAM).origin;
const control = process.env.RELAY_BROWSER_CONTROL;
let stage = 'launch';
let selectionProbe = null;
let reviewProbe = null;
const modelResponses = [];
const selectionDiagnosticFile = `test-results/routing-selection-${process.pid}.jsonl`;
console.log('routing_diagnostic_execution=' + process.pid);
async function focus(page, control) { for (let index = 0; index < 160; index++) { if (await control.evaluate(node => node === document.activeElement)) { assert.equal(await control.evaluate(node => { const style = getComputedStyle(node); return node.matches(':focus-visible') && style.outlineStyle !== 'none' && parseFloat(style.outlineWidth) > 0; }), true); return; } await page.keyboard.press('Tab'); } throw new Error('Keyboard could not reach control'); }
async function press(page, name) { await focus(page, page.getByRole('button', {name, exact: true})); await page.keyboard.press('Enter'); }
async function input(page, control, value) { await focus(page, control); await page.keyboard.press('ControlOrMeta+A'); await page.keyboard.type(String(value)); }
async function checked(page, control, value) { if (await control.isChecked() === value) return; await focus(page, control); await page.keyboard.press('Space'); assert.equal(await control.isChecked(), value); }
async function select(page, control, label) {
  selectionProbe = {boundary: 'locator', matches: await control.count()};
  assert.equal(selectionProbe.matches, 1);
  const index = await control.evaluate((node, text) => [...node.options].findIndex(option => option.label === text), label);
  selectionProbe.targetIndex = index; assert.ok(index >= 0);
  const trace = await control.evaluateHandle(observeSelection);
  async function sample(boundary) {
    const observed = await trace.evaluate((state, boundary) => state.sample(boundary), boundary);
    selectionProbe = {...selectionProbe, ...observed, events: observed.events.slice(-6)};
  }
  try {
    await sample('before focus'); await focus(page, control);
    await typeAheadCandidate(control, page.keyboard, label, sample);
    assert.equal(await control.evaluate(node => node.selectedOptions[0]?.label), label);
  } finally {
    let observed;
    try { observed = await trace.evaluate(state => state.stop()); } finally { await trace.dispose(); }
    await mkdir('test-results', {recursive: true});
    await appendFile(selectionDiagnosticFile, JSON.stringify({execution: process.pid, stage, matches: selectionProbe.matches, targetIndex: index, ...observed}) + '\n');
  }
}
async function panel(page, name, id) { await press(page, name); await page.locator('#' + id).waitFor({state: 'visible'}); }
async function counts(context) { const result = await context.request.post(control + '/route/counts'); assert.equal(result.status(), 204); return {primary: Number(result.headers()['x-primary']), secondary: Number(result.headers()['x-secondary']), checks: Number(result.headers()['x-checks'])}; }
try {
  for (const width of [1280, 390]) {
    const context = await browser.newContext({viewport: {width, height: 900}});
    const page = await context.newPage(); page.on('dialog', dialog => dialog.accept());
    page.on('response', response => {
      const path = new URL(response.url()).pathname; const method = response.request().method();
      if ((path === '/admin/api/models' || path.startsWith('/admin/api/models/')) && ['GET', 'POST', 'PATCH'].includes(method)) {
        modelResponses.push({method, status: response.status()}); if (modelResponses.length > 12) modelResponses.shift();
      }
    });
    stage = `login ${width}`;
    await page.goto(origin + '/admin/');
    if (await page.locator('#login').isVisible()) { await input(page, page.getByLabel('管理员密码'), process.env.RELAY_BROWSER_PASSWORD); await press(page, '登录'); }
    await page.locator('#overview').waitFor({state: 'visible'});
    await context.request.post(control + '/route/reset');
    const names = ['A-' + width + ' 受控', 'B-' + width + ' 受控'];
    for (let index = 0; index < 2; index++) {
      stage = `create/check upstream ${index} ${width}`;
      await panel(page, '上游', 'upstream'); await press(page, '新建上游');
      await input(page, page.getByLabel('上游名称', {exact: true}), names[index]);
      await input(page, page.getByLabel('上游地址'), controlled + (index ? '/routing-b/v1' : '/routing-a/v1'));
      await input(page, page.getByLabel('上游凭据'), 'controlled-browser-routing'); await checked(page, page.getByLabel('启用上游'), true);
      const saved = page.waitForResponse(response => new URL(response.url()).pathname === '/admin/api/upstreams' && response.request().method() === 'POST'); await press(page, '保存上游'); assert.equal((await saved).status(), 201);
      await page.locator('#notice').filter({hasText: '上游已保存'}).waitFor(); assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
      await press(page, '核验连接（不生成）'); await page.locator('#check-state').filter({hasText: '核验成功'}).waitFor();
    }
    stage = `connection failure and cancel recovery ${width}`;
    for (const mode of ['invalid', 'auth', 'rate']) {
      await context.request.post(control + '/route/check-' + mode);
      await press(page, '核验连接（不生成）'); await page.locator('#check-state').filter({hasText: '核验失败'}).waitFor();
      assert.equal((await counts(context)).primary + (await counts(context)).secondary, 0);
    }
    await context.request.post(control + '/route/check-delay');
    const beforeCancel = await context.request.get(control + '/route/counts');
    const canceledBefore = Number(beforeCancel.headers()['x-check-canceled']);
    const checksBefore = (await counts(context)).checks;
    await press(page, '核验连接（不生成）');
    await page.locator('#cancel-check').waitFor({state: 'visible'});
    assert.equal(await page.locator('#check-upstream').isDisabled(), true);
    for (let poll = 0; (await counts(context)).checks === checksBefore && poll < 100; poll++) await page.waitForTimeout(20);
    assert.equal((await counts(context)).checks, checksBefore + 1);
    await press(page, '取消核验'); await page.locator('#check-state').filter({hasText: '核验已取消'}).waitFor();
    for (let poll = 0; poll < 100; poll++) {
      const receipt = await context.request.get(control + '/route/counts');
      if (Number(receipt.headers()['x-check-canceled']) === canceledBefore + 1) break;
      await page.waitForTimeout(20);
    }
    assert.equal(Number((await context.request.get(control + '/route/counts')).headers()['x-check-canceled']), canceledBefore + 1);
    assert.equal((await counts(context)).checks, checksBefore + 1);
    await page.locator('#check-upstream:enabled').waitFor();
    let persistedCanceled = false;
    for (let poll = 0; poll < 100; poll++) {
      const receipt = await context.request.get(origin + '/admin/api/upstreams');
      const listing = await receipt.json();
      persistedCanceled = listing.data.some(value => value.name === names[1] && value.check?.state === 'canceled');
      if (persistedCanceled) break;
      await page.waitForTimeout(20);
    }
    assert.equal(persistedCanceled, true);
    await press(page, '重新读取上游'); await page.locator('#check-state').filter({hasText: '核验已取消'}).waitFor();
    stage = `connection total timeout ${width}`;
    const timedCheck = page.waitForResponse(response => response.request().method() === 'POST' && new URL(response.url()).pathname.endsWith('/check'));
    await press(page, '核验连接（不生成）');
    const timedResult = await timedCheck; const timeoutState = await timedResult.json();
    assert.equal(timeoutState.state, 'failed'); assert.equal(timeoutState.error_code, 'upstream_timeout');
    assert.ok(timeoutState.duration_ms >= 9000 && timeoutState.duration_ms <= 12000);
    await page.locator('#check-state').filter({hasText: '核验失败'}).waitFor();
    for (let poll = 0; poll < 100; poll++) {
      if (Number((await context.request.get(control + '/route/counts')).headers()['x-check-canceled']) === canceledBefore + 2) break;
      await page.waitForTimeout(20);
    }
    assert.equal(Number((await context.request.get(control + '/route/counts')).headers()['x-check-canceled']), canceledBefore + 2);
    stage = `connection navigation abort ${width}`;
    const beforeLeave = Number((await context.request.get(control + '/route/counts')).headers()['x-check-canceled']);
    const checksBeforeLeave = (await counts(context)).checks;
    await press(page, '核验连接（不生成）');
    for (let poll = 0; (await counts(context)).checks === checksBeforeLeave && poll < 100; poll++) await page.waitForTimeout(20);
    assert.equal((await counts(context)).checks, checksBeforeLeave + 1);
    await panel(page, '概览', 'overview');
    for (let poll = 0; poll < 100; poll++) {
      if (Number((await context.request.get(control + '/route/counts')).headers()['x-check-canceled']) === beforeLeave + 1) break;
      await page.waitForTimeout(20);
    }
    assert.equal(Number((await context.request.get(control + '/route/counts')).headers()['x-check-canceled']), beforeLeave + 1);
    await panel(page, '上游', 'upstream');
    stage = `upstream edit rejection recovery ${width}`;
    await input(page, page.getByLabel('上游名称', {exact: true}), names[1] + ' draft'); await press(page, '取消修改');
    assert.equal(await page.getByLabel('上游名称', {exact: true}).inputValue(), names[1]);
    await input(page, page.getByLabel('上游名称', {exact: true}), names[0]);
    let rejected = page.waitForResponse(response => response.request().method() === 'PATCH' && new URL(response.url()).pathname.startsWith('/admin/api/upstreams/'));
    await press(page, '保存上游'); assert.equal((await rejected).status(), 409);
    await page.locator('#error').filter({hasText: '名称已存在'}).waitFor();
    await input(page, page.getByLabel('上游名称', {exact: true}), names[1]);
    await input(page, page.getByLabel('上游地址'), controlled + '/changed/v1');
    rejected = page.waitForResponse(response => response.request().method() === 'PATCH' && new URL(response.url()).pathname.startsWith('/admin/api/upstreams/'));
    await press(page, '保存上游'); assert.equal((await rejected).status(), 400);
    await page.locator('#error').filter({hasText: '凭据'}).waitFor();
    await input(page, page.getByLabel('上游地址'), controlled + '/routing-b/v1');
    await input(page, page.getByLabel('上游凭据'), 'controlled-browser-routing');
    const corrected = page.waitForResponse(response => response.request().method() === 'PATCH' && new URL(response.url()).pathname.startsWith('/admin/api/upstreams/'));
    const reread = page.waitForResponse(response => response.request().method() === 'GET' && new URL(response.url()).pathname === '/admin/api/upstreams');
    await press(page, '保存上游'); assert.equal((await corrected).status(), 200); await reread;
    await page.locator('#notice').filter({hasText: '上游已保存'}).waitFor();
    await context.request.post(control + '/route/check-valid'); await press(page, '核验连接（不生成）');
    await page.locator('#check-state').filter({hasText: '核验成功'}).waitFor();
    const alias = 'routing-' + width;
    stage = `create candidate set ${width}`;
    await panel(page, '模型', 'models'); if (!await page.locator('#model-create').evaluate(node => node.open)) { await focus(page, page.locator('#model-create summary')); await page.keyboard.press('Enter'); }
    stage = `candidate alias ${width}`; await input(page, page.getByLabel('模型别名', {exact: true}), alias);
    let rows = page.locator('#create-candidates fieldset');
    stage = `candidate primary selection ${width}`; await select(page, rows.nth(0).getByLabel('上游', {exact: true}), names[0]);
    stage = `candidate primary fields ${width}`; await input(page, rows.nth(0).getByLabel('远端模型名称'), 'routing-a-model'); await input(page, rows.nth(0).getByLabel('同级权重'), '3');
    stage = `candidate add row ${width}`; await press(page, '添加路由候选');
    stage = `candidate secondary selection ${width}`; await select(page, rows.nth(1).getByLabel('上游', {exact: true}), names[1]);
    stage = `candidate secondary fields ${width}`; await input(page, rows.nth(1).getByLabel('远端模型名称'), 'routing-b-model');
    stage = `candidate create acknowledgement ${width}`; await press(page, '创建模型'); await page.locator('#model-list').filter({hasText: alias}).waitFor();
    stage = `candidate invalid edit ${width}`; await press(page, '编辑映射 ' + alias); await input(page, page.locator('#edit-candidates fieldset').nth(1).getByLabel('同级权重'), '0'); await press(page, '保存映射');
    stage = `candidate cancel draft ${width}`;
    assert.equal(await page.locator('#model-edit').isVisible(), true); await input(page, page.locator('#edit-candidates fieldset').nth(1).getByLabel('同级权重'), '1'); await press(page, '取消编辑');
    await panel(page, '应用密钥', 'keys'); if (!await page.locator('#key-create').evaluate(node => node.open)) { await focus(page, page.locator('#key-create summary')); await page.keyboard.press('Enter'); };
    await input(page, page.getByLabel('密钥名称'), 'routing-key-' + width); await press(page, '签发密钥'); await page.locator('#issued').waitFor({state: 'visible'}); const key = await page.locator('#issued-key').inputValue(); await press(page, '关闭展示');
    async function call(stream = false, message = 'route ordinary') {
      await panel(page, '调用测试', 'test'); await select(page, page.getByLabel('调用别名'), alias); await input(page, page.getByLabel('应用密钥', {exact: true}), key); await input(page, page.getByLabel('文本消息'), message); await checked(page, page.getByLabel('SSE 流式'), stream);
      const pending = page.waitForResponse(response => new URL(response.url()).pathname === '/v1/chat/completions'); await press(page, '发送请求'); const result = await pending; if (!stream) { assert.equal(result.status(), 200); await page.locator('#test-status').filter({hasText: '完成'}).waitFor(); } return result;
    }
    async function edit(callback) { await panel(page, '模型', 'models'); await press(page, '编辑映射 ' + alias); await callback(page.locator('#edit-candidates fieldset')); const saved = page.waitForResponse(response => response.request().method() === 'PATCH' && new URL(response.url()).pathname === '/admin/api/models/' + alias); const reread = page.waitForResponse(response => response.request().method() === 'GET' && new URL(response.url()).pathname === '/admin/api/models'); await press(page, '保存映射'); assert.equal((await saved).status(), 200); await reread; await page.locator('#model-edit').waitFor({state: 'hidden'}); await page.locator('#notice').filter({hasText: '映射已保存'}).waitFor(); }
    await context.request.post(control + '/route/reset'); stage = `actual weighted selection ${width}`;
    for (let index = 0; index < 40; index++) await call(); assert.deepEqual(await counts(context), {primary: 30, secondary: 10, checks: 0});
    stage = `completed task fold and keyboard review ${width}`;
    const overviewRead = page.waitForResponse(response => response.request().method() === 'GET' && new URL(response.url()).pathname === '/admin/api/overview');
    await panel(page, '概览', 'overview');
    const overviewResponse = await overviewRead; assert.equal(overviewResponse.status(), 200);
    const evidenceID = (await overviewResponse.json()).call_evidence?.request_id;
    assert.equal(typeof evidenceID, 'string'); assert.ok(evidenceID.length > 0);
    assert.equal(await page.locator('#setup-steps li').count(), 0);
    assert.equal(await page.locator('#completed-steps li').count(), 5);
    assert.equal(await page.locator('#completed-setup').evaluate(node => node.open), false);
    reviewProbe = {pendingCount: 0, completedCount: 5, open: false};
    stage = `completed review keyboard expansion ${width}`;
    await focus(page, page.locator('#completed-setup summary')); await page.keyboard.press('Enter');
    await page.locator('#completed-steps').waitFor({state: 'visible'});
    reviewProbe.open = await page.locator('#completed-setup').evaluate(node => node.open);
    assert.equal(reviewProbe.open, true);
    stage = `completed review exact evidence action ${width}`;
    const detailRead = page.waitForResponse(response => response.request().method() === 'GET' && new URL(response.url()).pathname === '/admin/api/requests/' + encodeURIComponent(evidenceID));
    await press(page, '查看成功请求详情');
    const detailResponse = await detailRead; reviewProbe.detailStatus = detailResponse.status();
    assert.equal(reviewProbe.detailStatus, 200);
    reviewProbe.sameIdentity = (await detailResponse.json()).id === evidenceID;
    assert.equal(reviewProbe.sameIdentity, true);
    stage = `completed review exact evidence rendering ${width}`;
    await page.locator('#record-detail').filter({hasText: evidenceID}).waitFor();
    assert.equal(await page.locator('#find-request-id').inputValue() === evidenceID, true);
    assert.equal((await page.locator('#record-detail').textContent()).split('\n')[0] === evidenceID, true);
    stage = `completed review keyboard collapse ${width}`;
    await panel(page, '概览', 'overview'); await focus(page, page.locator('#completed-setup summary')); await page.keyboard.press('Enter');
    assert.equal(await page.locator('#completed-setup').evaluate(node => node.open), false);
    await edit(async rows => input(page, rows.nth(0).getByLabel('同级权重'), '1')); await context.request.post(control + '/route/reset'); for (let index = 0; index < 20; index++) await call(); assert.deepEqual(await counts(context), {primary: 10, secondary: 10, checks: 0});
    await edit(async rows => input(page, rows.nth(1).getByLabel('优先级（大数先用）'), '10')); await context.request.post(control + '/route/reset'); for (let index = 0; index < 10; index++) await call(); assert.deepEqual(await counts(context), {primary: 0, secondary: 10, checks: 0});
    await edit(async rows => checked(page, rows.nth(1).getByLabel('参与路由'), false)); await context.request.post(control + '/route/reset'); await call(); assert.deepEqual(await counts(context), {primary: 1, secondary: 0, checks: 0});
    await edit(async rows => { await checked(page, rows.nth(1).getByLabel('参与路由'), true); await input(page, rows.nth(1).getByLabel('优先级（大数先用）'), '0'); await input(page, rows.nth(0).getByLabel('同级权重'), '1'); });
    stage = `failover and unknown consumption ${width}`; await context.request.post(control + '/route/reset'); await context.request.post(control + '/route/fail'); const recovered = await call(); const id = recovered.headers()['x-request-id'];
    await press(page, '查看本次记录'); await page.locator('#record-detail').filter({hasText: id}).waitFor(); const text = await page.locator('#record-detail').textContent(); assert.ok(text.includes('尝试 1') && text.includes('尝试 2') && text.includes('完整合计 未知') && text.includes('未知尝试 1')); assert.deepEqual(await counts(context), {primary: 1, secondary: 1, checks: 0});
    await call(); assert.deepEqual(await counts(context), {primary: 1, secondary: 2, checks: 0}); await panel(page, '模型', 'models'); await page.locator('#model-list').filter({hasText: '暂时冷却'}).waitFor();
    await context.request.post(control + '/route/recover'); await page.waitForTimeout(31000); stage = `passive recovery ${width}`; await call(); assert.equal((await counts(context)).primary, 2);
    await edit(async rows => input(page, rows.nth(0).getByLabel('优先级（大数先用）'), '100')); await context.request.post(control + '/route/reset'); stage = `committed stream no replay ${width}`;
    await call(true, 'route stream cut'); await page.locator('#test-status').filter({hasText: '失败'}).waitFor(); assert.equal((await counts(context)).secondary, 0);
    await call(true, 'route cancel'); await page.locator('#test-output').filter({hasText: 'controlled routed first'}).waitFor(); await press(page, '取消请求'); await page.locator('#test-status').filter({hasText: '已取消'}).waitFor(); assert.equal((await counts(context)).secondary, 0);
    await panel(page, '调用记录', 'records'); assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true); assert.equal(await page.evaluate(() => ['password', 'upstream-token', 'issued-key', 'test-key', 'test-message'].every(id => document.getElementById(id).value === '')), true);
    stage = `restart task review ${width}`;
    assert.equal((await context.request.post(control + '/restart')).status(), 204);
    await page.reload(); await page.locator('#login').waitFor({state: 'visible'});
    await input(page, page.getByLabel('管理员密码'), process.env.RELAY_BROWSER_PASSWORD); await press(page, '登录');
    await page.locator('#console').waitFor({state: 'visible'}); await panel(page, '概览', 'overview');
    assert.equal(await page.locator('#setup-steps li').count(), 2);
    assert.equal(await page.locator('#completed-steps li').count(), 3);
    await focus(page, page.locator('#completed-setup summary')); await page.keyboard.press('Enter');
    await page.locator('#completed-steps').waitFor({state: 'visible'});
    await mkdir('test-results', {recursive: true}); await page.screenshot({path: `test-results/routing-${width}.png`, fullPage: true}); await context.close();
    console.log(`PASS real multi-upstream UI ${width}px: connection GET, priority/3:1 visits, disable, failover/unknown usage, passive recovery, SSE/cancel no replay; controlled upstreams only`);
  }
} catch (error) {
  const kind = ['AssertionError', 'TimeoutError', 'Error'].includes(error.name) ? error.name : 'OtherError';
  console.error('routing_failure=' + JSON.stringify({stage, kind, selectionProbe, reviewProbe, modelResponses, selectionDiagnosticFile}));
  throw new Error('Real routing browser journey failed at ' + stage + '; no credential or response-body diagnostic');
}
finally { await browser.close(); }
