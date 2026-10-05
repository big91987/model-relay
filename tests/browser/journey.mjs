import assert from 'node:assert/strict';
import { chromium } from 'playwright';

let browser;
let checkpoint = 'launch';
let viewport = 0;
const responses = [];
const origin = process.env.RELAY_BROWSER_URL;
async function visibleText(page, selector, text) {
  checkpoint = `visible ${selector} contains ${text}`;
  await page.locator(selector).filter({hasText: text}).waitFor({state: 'visible', timeout: 10000});
}
async function navigate(page, name) {
  checkpoint = `click ${name}`;
  await page.getByRole('button', {name, exact: true}).click();
}
async function loadUpstreamPanel(page) {
  let releaseLoad;
  const loadReleased = new Promise(resolve => { releaseLoad = resolve; });
  const loadStarted = page.waitForRequest('**/admin/api/upstream', {timeout: 10000});
  const delayedLoad = async route => { await loadReleased; await route.continue(); };
  await page.route('**/admin/api/upstream', delayedLoad);
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
    await page.unroute('**/admin/api/upstream', delayedLoad);
  }
}
try {
  browser = await chromium.launch({headless: true});
  for (const width of [1280, 390]) {
    viewport = width;
    responses.length = 0;
    const context = await browser.newContext({viewport: {width, height: 900}});
    const page = await context.newPage();
    page.on('dialog', dialog => dialog.accept());
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('response', response => {
      const pathname = new URL(response.url()).pathname;
      if (!['/admin/api/session', '/admin/api/login', '/admin/api/upstream', '/admin/api/models', '/admin/api/keys', '/admin/api/requests', '/admin/api/logout', '/v1/chat/completions'].includes(pathname)) return;
      responses.push({method: response.request().method(), path: pathname, status: response.status()});
      if (responses.length > 12) responses.shift();
    });
    checkpoint = 'open login';
    await page.goto(origin + '/admin/');
    await page.getByLabel('管理员密码').fill('wrong fixture password');
    await navigate(page, '登录');
    await visibleText(page, '#error', 'unauthorized');
    await page.getByLabel('管理员密码').fill(process.env.RELAY_BROWSER_PASSWORD);
    await navigate(page, '登录');
    await visibleText(page, '#summary', width === 1280 ? '待配置' : '上游');
    await loadUpstreamPanel(page);
    await page.getByLabel('上游地址').fill(process.env.RELAY_BROWSER_UPSTREAM);
    await page.getByLabel('上游凭据').fill('controlled-browser-upstream');
    await page.getByLabel('启用上游').check();
    await navigate(page, '保存上游');
    await visibleText(page, '#notice', '上游已保存');
    await page.getByLabel('上游地址').fill('http://127.0.0.1:1/v1');
    await navigate(page, '保存上游');
    await visibleText(page, '#error', 'invalid_destination');
    assert.equal(await page.locator('#notice').textContent(), '');
    await navigate(page, '概览');
    await loadUpstreamPanel(page);
    checkpoint = 'rejected save restores persisted upstream and hides credential';
    assert.equal(await page.getByLabel('上游地址').inputValue(), process.env.RELAY_BROWSER_UPSTREAM);
    assert.equal(await page.getByLabel('上游凭据').inputValue(), '');
    await navigate(page, '模型');
    for (const alias of ['alpha-' + width, 'beta-' + width]) {
      await page.getByLabel('模型别名', {exact: true}).fill(alias);
      await page.getByLabel('真实模型', {exact: true}).fill(alias.startsWith('alpha-') ? 'controlled-model' : 'controlled-beta-model');
      await navigate(page, '创建模型');
      await visibleText(page, '#model-list', alias);
    }
    await navigate(page, '创建模型');
    await visibleText(page, '#error', 'alias_conflict');
    page.removeAllListeners('dialog');
    page.on('dialog', dialog => dialog.type() === 'prompt' ? dialog.accept('edited-controlled-model') : dialog.accept());
    await navigate(page, '编辑映射 alpha-' + width);
    await visibleText(page, '#notice', '映射已保存');
    await visibleText(page, '#model-list', 'alpha-' + width + ' → edited-controlled-model');
    await navigate(page, '应用密钥');
    const keys = [];
    for (const name of ['first-' + width, 'second-' + width]) {
      await page.getByLabel('密钥名称').fill(name);
      await navigate(page, '签发密钥');
      checkpoint = 'visible #issued';
      await page.locator('#issued').waitFor({state: 'visible'});
      const key = await page.getByLabel('新密钥（仅本次）').inputValue();
      assert.match(key, /^mr_/);
      keys.push(key);
      await navigate(page, '关闭展示');
      assert.equal(await page.getByLabel('新密钥（仅本次）').inputValue(), '');
    }
    checkpoint = 'reload';
    await page.reload();
    await visibleText(page, '#summary', '上游');
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
    await visibleText(page, '#summary', '上游');
    const restoredModels = await context.request.get(origin + '/admin/api/models');
    assert.equal(restoredModels.status(), 200);
    const mappings = await restoredModels.json();
    checkpoint = 'two persisted mappings from real management API';
    assert.equal(mappings.some(value => value.alias === 'alpha-' + width && value.model === 'edited-controlled-model' && value.enabled), true);
    assert.equal(mappings.some(value => value.alias === 'beta-' + width && value.model === 'controlled-beta-model' && value.enabled), true);
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
      await page.getByLabel('应用密钥', {exact: true}).fill(key);
      await page.getByLabel('调用别名').fill('alpha-' + width);
      await page.getByLabel('文本消息').fill(text);
      await page.getByLabel('SSE 流式').setChecked(stream);
    }
    await fillCall(keys[0], 'browser-prompt-sentinel', false);
    await navigate(page, '发送请求');
    await visibleText(page, '#test-status', '完成');
    await visibleText(page, '#test-output', 'controlled ordinary');
    await fillCall(keys[0], 'stream fixture', true);
    await navigate(page, '发送请求');
    await visibleText(page, '#test-status', '完成');
    await visibleText(page, '#test-output', 'controlled first');
    await fillCall(keys[0], 'cancel fixture', true);
    await navigate(page, '发送请求');
    await visibleText(page, '#test-output', 'controlled first');
    await navigate(page, '取消请求');
    await visibleText(page, '#test-status', '已取消');
    assert.equal(await page.getByLabel('应用密钥', {exact: true}).inputValue(), '');
    await fillCall(keys[0], 'browser-error-prompt-sentinel', false);
    await navigate(page, '发送请求');
    await visibleText(page, '#error', 'upstream_rejected');
    await visibleText(page, '#test-status', '不完整 / 失败');
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
    const secrets = [process.env.RELAY_BROWSER_PASSWORD, 'controlled-browser-upstream', ...keys, 'browser-prompt-sentinel', 'browser-error-prompt-sentinel', 'controlled ordinary', 'controlled first', 'stream fixture', 'cancel fixture', 'browser-sensitive-diagnostic-sentinel'];
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
    await visibleText(page, '#error', 'unauthorized');
    await fillCall(keys[1], 'second fixture', false);
    await navigate(page, '发送请求');
    await visibleText(page, '#test-status', '完成');
    await navigate(page, '调用记录');
    await visibleText(page, '#record-list', 'upstream_rejected');
    checkpoint = 'final record-list and service logs exclude all journey secrets and conversations';
    const finalSecrets = [...secrets, 'disabled fixture', 'second fixture'];
    const finalRecords = await page.locator('#record-list').textContent();
    for (const secret of finalSecrets) assert.equal(finalRecords.includes(secret), false, 'record-list exclusion failed');
    const finalLogAudit = await context.request.post(process.env.RELAY_BROWSER_CONTROL + '/audit-logs', {data: finalSecrets});
    assert.equal(finalLogAudit.status(), 204, 'service log exclusion failed');
    assert.equal(Number(finalLogAudit.headers()['x-controlled-sensitive-failures']), width === 1280 ? 1 : 2, 'sensitive failure count changed unexpectedly');
    await navigate(page, '概览');
    const storage = await context.storageState();
    assert.deepEqual(storage.origins, []);
    assert.ok(storage.cookies.every(cookie => cookie.httpOnly && cookie.sameSite === 'Strict' && cookie.path === '/admin'));
    assert.equal(await page.evaluate(() => localStorage.length + sessionStorage.length), 0);
    assert.deepEqual(await page.evaluate(async () => indexedDB.databases()), []);
    assert.equal(await page.evaluate(async () => (await navigator.serviceWorker.getRegistrations()).length), 0);
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
    assert.deepEqual(errors, []);
    await navigate(page, '退出');
    checkpoint = 'visible #login after logout';
    await page.locator('#login').waitFor({state: 'visible'});
    const response = await context.request.get(origin + '/admin/api/keys');
    assert.equal(response.status(), 401);
    await context.close();
    console.log(`PASS actual Go service browser journey ${width}px (controlled upstream, not supplier validation)`);
  }
} catch (error) {
  const location = String(error.stack).match(/journey\.mjs:\d+:\d+/)?.[0] || 'unavailable';
  console.error('FAIL actual Go service browser journey:', JSON.stringify({viewport, checkpoint, location, error: error.name, responses}));
  process.exitCode = 1;
} finally {
  if (browser) await browser.close();
}
