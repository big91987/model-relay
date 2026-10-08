import assert from 'node:assert/strict';
import {chromium} from 'playwright';
import {verifyMigratedKeys} from './migration-contract.mjs';

const origin = process.env.RELAY_LEGACY_URL;
const browser = await chromium.launch({headless: true});
let stage = 'health';
const diagnostic = {health_status: null, schema: null, keys_status: null, key_count: null, overview_status: null};
try {
  for (const width of [1280, 1440, 390]) {
    stage = `health-${width}`;
    for (const name of Object.keys(diagnostic)) diagnostic[name] = null;
    const context = await browser.newContext({viewport: {width, height: 900}});
    try {
      const health = await context.request.get(origin + '/healthz');
      diagnostic.health_status = health.status();
      assert.equal(health.status(), 200);
      const schema = (await health.json()).storage_schema;
      diagnostic.schema = Number.isSafeInteger(schema) ? schema : null;
      assert.equal(diagnostic.schema, 3);
      const page = await context.newPage();
      stage = `login-${width}`;
      await page.goto(origin + '/admin/');
      await page.getByLabel('账号', {exact: true}).fill('admin');
      await page.getByLabel('密码', {exact: true}).fill(process.env.RELAY_LEGACY_PASSWORD);
      await page.getByRole('button', {name: '登录', exact: true}).click();
      await page.locator('#workspace').waitFor({state: 'visible'});
      await page.getByLabel('当前租户').selectOption('legacy');
      if (width === 390) await page.getByRole('button', {name: '菜单', exact: true}).click();
      stage = `keys-navigation-${width}`;
      await page.locator('#navigation').getByRole('button', {name: 'API Keys', exact: true}).click();
      await page.locator('#content').getByText('正在读取当前范围…').waitFor({state: 'hidden'});
      await page.getByRole('row').filter({hasText: 'migration-enabled'}).waitFor();
      await page.getByRole('row').filter({hasText: 'migration-disabled'}).waitFor();
      stage = `keys-http-${width}`;
      const keys = await context.request.get(origin + '/admin/api/tenants/legacy/keys');
      diagnostic.keys_status = keys.status();
      assert.equal(keys.status(), 200);
      const body = await keys.json();
      stage = `keys-contract-${width}`;
      diagnostic.key_count = Array.isArray(body.items) ? body.items.length : null;
      verifyMigratedKeys(body);
      stage = `history-${width}`;
      const overview = await context.request.get(origin + '/admin/api/workspace/overview');
      diagnostic.overview_status = overview.status();
      assert.equal(overview.status(), 200);
      assert.ok((await overview.json()).legacy_pending_requests > 0);
      assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1));
      console.log(`PASS migrated schema3 native identity/keys/pending-history viewport=${width}`);
    } finally {
      await context.close();
    }
  }
} catch {
  console.error(JSON.stringify({suite: 'production-migration-console', stage, ...diagnostic, result: 'failed'}));
  throw new Error(`Production migration console failed at ${stage}; secret diagnostics withheld`);
} finally {
  await browser.close();
}
