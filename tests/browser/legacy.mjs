import assert from 'node:assert/strict';
import { chromium } from 'playwright';
import { mkdir } from 'node:fs/promises';

const browser = await chromium.launch({headless: true});
const context = await browser.newContext({viewport: {width: 1280, height: 900}});
const page = await context.newPage();
const origin = process.env.RELAY_LEGACY_URL;
try {
  await page.goto(origin + '/admin/');
  await page.locator('#login').waitFor({state: 'visible'});
  await page.getByLabel('管理员密码').fill(process.env.RELAY_LEGACY_PASSWORD);
  await page.getByRole('button', {name: '登录', exact: true}).click();
  await page.locator('#overview').waitFor({state: 'visible'});
  await page.getByRole('button', {name: '上游', exact: true}).click();
  await page.locator('#upstream').waitFor({state: 'visible'});
  await page.getByLabel('上游地址').fill(process.env.RELAY_LEGACY_UPSTREAM);
  await page.getByLabel('上游凭据').fill('isolated-legacy-upstream');
  await page.getByLabel('启用上游').check();
  await page.getByRole('button', {name: '保存上游', exact: true}).click();
  await page.locator('#notice').filter({hasText: '上游已保存'}).waitFor();
  await page.getByRole('button', {name: '模型', exact: true}).click();
  await page.locator('#models').waitFor({state: 'visible'});
  for (const alias of ['migration-a', 'migration-b']) {
    if (!await page.locator('#model-create').evaluate(node => node.open)) await page.locator('#model-create summary').click();
    await page.getByLabel('模型别名', {exact: true}).fill(alias);
    await page.getByLabel('真实模型', {exact: true}).fill('legacy-' + alias);
    await page.getByRole('button', {name: '创建模型', exact: true}).click();
    await page.locator('#model-list').filter({hasText: alias}).waitFor();
  }
  page.on('dialog', dialog => dialog.accept());
  await page.getByRole('button', {name: '停用模型 migration-b', exact: true}).click();
  await page.getByRole('button', {name: '启用模型 migration-b', exact: true}).waitFor();
  await page.getByRole('button', {name: '应用密钥', exact: true}).click();
  await page.locator('#keys').waitFor({state: 'visible'});
  let applicationKey;
  for (const name of ['migration-enabled', 'migration-disabled']) {
    if (!await page.locator('#key-create').evaluate(node => node.open)) await page.locator('#key-create summary').click();
    await page.getByLabel('密钥名称').fill(name);
    await page.getByRole('button', {name: '签发密钥', exact: true}).click();
    await page.locator('#issued').waitFor({state: 'visible'});
    if (name === 'migration-enabled') applicationKey = await page.locator('#issued-key').inputValue();
    await page.getByRole('button', {name: '关闭展示', exact: true}).click();
  }
  await page.getByRole('button', {name: '停用密钥 migration-disabled', exact: true}).click();
  await page.getByRole('button', {name: '启用密钥 migration-disabled', exact: true}).waitFor();
  await page.getByRole('button', {name: '调用测试', exact: true}).click();
  await page.locator('#test').waitFor({state: 'visible'});
  await page.getByLabel('应用密钥', {exact: true}).fill(applicationKey);
  await page.getByLabel('调用别名').selectOption('migration-a');
  await page.getByLabel('文本消息').fill('isolated legacy UI generation');
  const response = page.waitForResponse(value => new URL(value.url()).pathname === '/v1/chat/completions');
  await page.getByRole('button', {name: '发送请求', exact: true}).click();
  const generated = await response; assert.equal(generated.status(), 200);
  const requestID = generated.headers()['x-request-id']; assert.ok(requestID);
  await page.getByRole('button', {name: '查看本次记录', exact: true}).click();
  await page.locator('#record-detail').filter({hasText: requestID}).waitFor();
  await page.getByRole('button', {name: '调用测试', exact: true}).click();
  await page.locator('#test').waitFor({state: 'visible'});
  await page.getByLabel('应用密钥', {exact: true}).fill(applicationKey);
  await page.getByLabel('调用别名').selectOption('migration-a');
  await page.getByLabel('文本消息').fill('isolated legacy failure');
  const rejected = page.waitForResponse(value => new URL(value.url()).pathname === '/v1/chat/completions');
  await page.getByRole('button', {name: '发送请求', exact: true}).click();
  const failure = await rejected; assert.equal(failure.status(), 429);
  const failureID = failure.headers()['x-request-id']; assert.ok(failureID && failureID !== requestID);
  await page.getByRole('button', {name: '查看本次记录', exact: true}).click();
  await page.locator('#record-detail').filter({hasText: failureID}).waitFor();
  assert.equal(await page.evaluate(() => ['upstream-token', 'issued-key', 'test-key', 'test-message'].every(id => document.getElementById(id).value === '')), true);
  await mkdir('test-results', {recursive: true}); await page.screenshot({path: 'test-results/legacy-ui-before-upgrade.png', fullPage: true});
  process.stdout.write(JSON.stringify({application_key: applicationKey, request_id: requestID, failure_id: failureID, aliases: ['migration-a', 'migration-b']}));
} catch {
  throw new Error('Exact legacy UI migration seed journey failed; no secret diagnostic emitted');
} finally {
  await context.close(); await browser.close();
}
