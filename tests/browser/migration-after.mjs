import { chromium } from 'playwright';

const browser = await chromium.launch({headless: true});
const context = await browser.newContext({viewport: {width: 1280, height: 900}});
const page = await context.newPage();
try {
  await page.goto(process.env.RELAY_LEGACY_URL + '/admin/');
  await page.locator('#login').waitFor({state: 'visible'});
  await page.getByLabel('管理员密码').fill(process.env.RELAY_LEGACY_PASSWORD);
  await page.getByRole('button', {name: '登录', exact: true}).click();
  await page.locator('#overview').waitFor({state: 'visible'});
  await page.getByRole('button', {name: '应用密钥', exact: true}).click();
  await page.locator('#keys').waitFor({state: 'visible'});
  await page.getByRole('button', {name: '停用密钥 migration-enabled', exact: true}).waitFor();
  await page.getByRole('button', {name: '启用密钥 migration-disabled', exact: true}).waitFor();
  if (!await page.locator('#key-create').evaluate(node => node.open)) await page.locator('#key-create summary').click();
  await page.getByLabel('密钥名称').fill('upgrade-new-key');
  await page.getByRole('button', {name: '签发密钥', exact: true}).click();
  await page.locator('#issued').waitFor({state: 'visible'});
  await page.getByRole('button', {name: '关闭展示', exact: true}).click();
  await page.getByRole('button', {name: '停用密钥 upgrade-new-key', exact: true}).waitFor();
} catch {
  throw new Error('Upgraded UI identity and new-object journey failed; no secret diagnostic emitted');
} finally {
  await context.close(); await browser.close();
}
