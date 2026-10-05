import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';

const source = readFileSync(new URL('../../internal/relay/web/app.js', import.meta.url), 'utf8');
const tick = () => new Promise(resolve => setImmediate(resolve));

function fixture() {
  const elements = new Map();
  const element = id => {
    if (!elements.has(id)) elements.set(id, {id, value: '', textContent: '', hidden: true, disabled: false, checked: false});
    return elements.get(id);
  };
  const panels = ['overview', 'upstream', 'models', 'keys', 'test', 'records'].map(element);
  const buttons = panels.map(panel => ({dataset: {view: panel.id}}));
  const pending = [];
  const timers = new Map();
  let timerID = 0;
  const document = {
    getElementById: element,
    querySelectorAll: selector => selector === '.panel' ? panels : buttons,
  };
  vm.runInNewContext(source, {document, AbortController,
    setTimeout: (callback, delay) => { assert.equal(delay, 5000); timers.set(++timerID, callback); return timerID; },
    clearTimeout: id => timers.delete(id),
    fetch: (path, options) => new Promise((resolve, reject) => pending.push({path, options, resolve, reject}))});
  function respond(method, path, body, status = 200) {
    const index = pending.findIndex(request => request.path === '/admin/api/' + path && request.options.method === method);
    assert.notEqual(index, -1, `expected ${method} ${path}`);
    const [request] = pending.splice(index, 1);
    request.resolve({ok: status < 400, status, json: async () => body});
    return request;
  }
  return {element, panels, pending, timers, respond, navigate: name => buttons.find(button => button.dataset.view === name).onclick()};
}

async function ready() {
  const app = fixture();
  app.respond('GET', 'session', {error: {code: 'unauthorized'}}, 401);
  await tick();
  return app;
}

test('upstream form is unavailable until GET completes, then saves entered values', async () => {
  const app = await ready();
  const navigation = app.navigate('upstream');
  assert.equal(app.element('upstream').hidden, true, 'pending configuration must not expose an editable form');
  app.respond('GET', 'upstream', {url: '', enabled: false, credential_configured: false});
  await navigation;
  assert.equal(app.element('upstream').hidden, false);
  app.element('upstream-url').value = 'https://controlled.example/v1';
  app.element('upstream-token').value = 'test-only-token';
  app.element('upstream-enabled').checked = true;
  app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save')});
  const request = app.respond('PUT', 'upstream', {});
  assert.deepEqual(JSON.parse(request.options.body), {url: 'https://controlled.example/v1', token: 'test-only-token', enabled: true});
  await tick();
  app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true, credential_configured: true});
  await tick();
  assert.match(app.element('notice').textContent, /上游已保存/);
  assert.equal(app.element('upstream-token').value, '');
  assert.equal(app.element('save').disabled, false);
});

async function overview(app) {
  const navigation = app.navigate('overview');
  app.respond('GET', 'upstream', {credential_configured: false});
  await tick();
  app.respond('GET', 'session', {degraded: false});
  await navigation;
}

async function authenticated() {
  const app = fixture();
  assert.equal(app.pending.some(request => request.path === '/healthz'), false);
  app.respond('GET', 'session', {csrf: 'test-only-csrf'});
  await tick();
  app.respond('GET', 'upstream', {credential_configured: false});
  await tick();
  app.respond('GET', 'session', {degraded: false});
  await tick();
  return app;
}

function health(app) {
  const index = app.pending.findIndex(request => request.path === '/healthz');
  assert.notEqual(index, -1);
  return app.pending.splice(index, 1)[0];
}

test('runtime version preserves raw strings including dev without blocking overview', async () => {
  const app = await authenticated();
  assert.equal(app.element('overview').hidden, false);
  assert.match(app.element('summary').textContent, /待配置/);
  for (const version of ['dev', 'v1.2.3-build', '  <img src=x onerror=alert(1)> & version  ']) {
    const request = health(app);
    assert.equal(request.options.cache, 'no-store');
    assert.equal(request.options.method, 'GET');
    assert.equal(request.options.headers, undefined);
    request.resolve({ok: true, json: async () => ({version})});
    await tick();
    assert.equal(app.element('runtime-version').textContent, version);
    assert.equal(app.timers.size, 0);
    await app.navigate('test');
    assert.equal(app.element('runtime-version').textContent, '');
    await overview(app);
  }
});

for (const [name, body, status] of [
  ['missing', {}, 200], ['empty', {version: ''}, 200], ['whitespace', {version: ' \n '}, 200],
  ['number', {version: 123}, 200], ['null version', {version: null}, 200],
  ['null body', null, 200], ['array', [], 200], ['string body', 'dev', 200],
  ['401', {version: 'wrong'}, 401], ['503', {version: 'wrong'}, 503],
  ['network', undefined, 0], ['invalid json', undefined, 200],
]) {
  test(`runtime version ${name} degrades locally and reentry recovers`, async () => {
    const app = await authenticated();
    health(app).resolve({ok: true, json: async () => ({version: 'previous'})});
    await tick();
    await overview(app);
    assert.equal(app.element('runtime-version').textContent, '加载中');
    app.element('notice').textContent = 'management notice';
    const request = health(app);
    if (status === 0) request.reject(new Error('network'));
    else request.resolve({ok: status < 400, status, json: async () => {
      if (name === 'invalid json') throw new SyntaxError('invalid json');
      return body;
    }});
    await tick();
    assert.equal(app.element('runtime-version').textContent, '版本未知');
    assert.equal(app.element('console').hidden, false);
    assert.equal(app.element('notice').textContent, 'management notice');
    assert.equal(app.element('error').textContent, '');
    assert.equal(app.pending.length, 0, 'no polling');
    await app.navigate('test');
    await overview(app);
    health(app).resolve({ok: true, json: async () => ({version: 'recovered'})});
    await tick();
    assert.equal(app.element('runtime-version').textContent, 'recovered');
  });
}

for (const lateFailure of [false, true]) {
  test(`late ${lateFailure ? 'failure' : 'success'} cannot clear a newer timer or version`, async () => {
    const app = await authenticated();
    const old = health(app);
    await app.navigate('test');
    assert.equal(old.options.signal.aborted, true);
    await overview(app);
    const fresh = health(app);
    if (lateFailure) old.reject(new Error('late failure'));
    else old.resolve({ok: true, json: async () => ({version: 'stale'})});
    await tick();
    assert.equal(app.element('runtime-version').textContent, '加载中');
    assert.equal(app.timers.size, 1);
    fresh.resolve({ok: true, json: async () => ({version: 'current'})});
    await tick();
    assert.equal(app.element('runtime-version').textContent, 'current');
    assert.equal(app.timers.size, 0);
  });
}

for (const bodyPending of [false, true]) {
  test(`5 second timeout covers ${bodyPending ? 'body' : 'headers'} and ignores late resolution`, async () => {
    const app = await authenticated();
    const request = health(app);
    let releaseBody;
    if (bodyPending) {
      request.resolve({ok: true, json: () => new Promise(resolve => { releaseBody = resolve; })});
      await tick();
    }
    assert.equal(app.timers.size, 1);
    const [timerID, expire] = [...app.timers][0];
    app.timers.delete(timerID);
    expire();
    assert.equal(app.element('runtime-version').textContent, '版本未知');
    assert.equal(request.options.signal.aborted, true);
    if (bodyPending) releaseBody({version: 'late'});
    else request.resolve({ok: true, json: async () => ({version: 'late'})});
    await tick();
    assert.equal(app.element('runtime-version').textContent, '版本未知');
    await overview(app);
    health(app).resolve({ok: true, json: async () => ({version: 'retry'})});
    await tick();
    assert.equal(app.element('runtime-version').textContent, 'retry');
  });
}

test('logout invalidates immediately and late success cannot contaminate a new login', async () => {
  const app = await authenticated();
  const old = health(app);
  app.element('logout').onclick();
  assert.equal(old.options.signal.aborted, true);
  assert.equal(app.element('runtime-version').textContent, '');
  app.respond('POST', 'logout', {});
  await tick();
  assert.equal(app.element('console').hidden, true);
  app.element('login-form').onsubmit({preventDefault() {}, submitter: app.element('login-button')});
  app.respond('POST', 'login', {csrf: 'new-csrf'});
  await tick();
  app.respond('GET', 'upstream', {});
  await tick();
  app.respond('GET', 'session', {});
  await tick();
  old.resolve({ok: true, json: async () => ({version: 'old login'})});
  await tick();
  assert.equal(app.element('runtime-version').textContent, '加载中');
  assert.equal(app.timers.size, 1);
  health(app).resolve({ok: true, json: async () => ({version: 'new login'})});
  await tick();
  assert.equal(app.element('runtime-version').textContent, 'new login');
});

test('real management 401 still logs out and invalidates runtime version', async () => {
  const app = await authenticated();
  const old = health(app);
  const navigation = app.navigate('upstream');
  app.respond('GET', 'upstream', {error: {code: 'unauthorized'}}, 401);
  await navigation;
  old.reject(new Error('late health error'));
  await tick();
  assert.equal(app.element('console').hidden, true);
  assert.equal(app.element('runtime-version').textContent, '');
  assert.equal(app.timers.size, 0);
});

test('older response after a newer completion cannot overwrite the current version', async () => {
  const app = await authenticated();
  const old = health(app);
  await app.navigate('test');
  await overview(app);
  health(app).resolve({ok: true, json: async () => ({version: 'new-completed'})});
  await tick();
  old.resolve({ok: true, json: async () => ({version: 'old-completed'})});
  await tick();
  assert.equal(app.element('runtime-version').textContent, 'new-completed');
  assert.equal(app.timers.size, 0);
});

test('failed configuration load leaves an error and navigation can recover', async () => {
  const app = await ready();
  const navigation = app.navigate('upstream');
  app.respond('GET', 'upstream', {error: {code: 'storage_failed'}}, 503);
  await navigation;
  assert.equal(app.element('error').textContent, 'storage_failed');
  const retry = app.navigate('upstream');
  assert.equal(app.element('upstream').hidden, true);
  app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true});
  await retry;
  assert.equal(app.element('upstream').hidden, false);
  assert.equal(app.element('upstream-url').value, 'https://controlled.example/v1');
  assert.equal(app.element('error').textContent, '');
});

test('a stale navigation cannot reveal a panel after the user leaves', async () => {
  const app = await ready();
  const first = app.navigate('upstream');
  const second = app.navigate('test');
  await second;
  app.respond('GET', 'upstream', {url: '', enabled: false});
  await first;
  assert.equal(app.element('upstream').hidden, true);
  assert.equal(app.element('test').hidden, false);
});

test('rejected save recovers persisted configuration after navigation finishes', async () => {
  const app = await ready();
  const savedURL = 'https://controlled.example/v1';
  const first = app.navigate('upstream');
  app.respond('GET', 'upstream', {url: savedURL, enabled: true, credential_configured: true});
  await first;
  app.element('upstream-url').value = 'http://127.0.0.1:1/v1';
  app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save')});
  app.respond('PUT', 'upstream', {error: {code: 'invalid_destination'}}, 400);
  await tick();
  assert.equal(app.element('error').textContent, 'invalid_destination');
  assert.equal(app.element('upstream-url').value, 'http://127.0.0.1:1/v1');
  const overview = app.navigate('overview');
  app.respond('GET', 'upstream', {url: savedURL, enabled: true, credential_configured: true});
  await tick();
  app.respond('GET', 'session', {degraded: false});
  await overview;
  const recovery = app.navigate('upstream');
  assert.equal(app.element('upstream').hidden, true);
  assert.equal(app.element('upstream-url').value, 'http://127.0.0.1:1/v1');
  app.respond('GET', 'upstream', {url: savedURL, enabled: true, credential_configured: true});
  await recovery;
  assert.equal(app.element('upstream').hidden, false);
  assert.equal(app.element('upstream-url').value, savedURL);
  assert.equal(app.element('upstream-token').value, '');
  assert.equal(app.element('error').textContent, '');
});
