import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';
import { createHash } from 'node:crypto';

const source = readFileSync(new URL('../../internal/relay/web/app.js', import.meta.url), 'utf8');
const tick = () => new Promise(resolve => setImmediate(resolve));
const snapshot = {upstream: {configured: false, enabled: false, credential_configured: false}, models: {total: 0, enabled: 0}, keys: {total: 0, enabled: 0}, call_evidence: null};

function fixture() {
  const clock = {now: Date.now()};
  class FixtureDate extends Date { static now() { return clock.now; } }
  const elements = new Map();
  const element = id => {
    if (!elements.has(id)) elements.set(id, {id, value: '', type: id === 'password' ? 'password' : '', textContent: '', hidden: true, disabled: false, checked: false, children: [], focus() { document.activeElement = this; }, setAttribute() {}, removeAttribute() {}, querySelector() { return null; }, querySelectorAll() { return []; }, replaceChildren(...children) { if (this.children.some(child => child === document.activeElement || child.children?.includes(document.activeElement))) document.activeElement = document.body; this.children = children; }, append(...children) { this.children.push(...children); }});
    return elements.get(id);
  };
  const panels = ['overview', 'upstream', 'models', 'keys', 'test', 'records'].map(element);
  const buttons = panels.map(panel => ({dataset: {view: panel.id}, setAttribute() {}, removeAttribute() {}}));
  const pending = [];
  const timers = new Map();
  const sessionTimers = new Map();
  const timerDelays = new Map();
  const timerDue = new Map();
  const events = new Map();
  const location = {origin: 'http://controlled.test', hash: ''};
  const clipboard = {writeText: async () => {}};
  let timerID = 0;
  const document = {
    body: {tag: 'body'},
    getElementById: element,
    querySelectorAll: selector => selector === '.panel' ? panels : selector === '[data-view]' ? buttons : [...elements.values()],
    createElement: tag => ({tag, dataset: {}, disabled: false, children: [], focus() { document.activeElement = this; }, querySelector(selector) { return selector === 'button' ? this.children.find(child => child.tag === 'button') : null; }, append(...children) { this.children.push(...children); }}),
    addEventListener: (name, callback) => events.set(name, callback),
  };
  document.activeElement = document.body;
  vm.runInNewContext(source, {document, AbortController, Date: FixtureDate, TextDecoder, Uint8Array, location, history: {pushState(_state, _title, hash) { location.hash = hash; }, replaceState(_state, _title, hash) { location.hash = hash; }}, window: {addEventListener: (name, callback) => events.set(name, callback)}, navigator: {clipboard}, confirm: () => true,
    setTimeout: (callback, delay) => {
      const id = ++timerID; timerDelays.set(id, delay); timerDue.set(id, clock.now + delay);
      const once = () => {
        if (!timerDelays.has(id)) return;
        timers.delete(id); sessionTimers.delete(id); timerDelays.delete(id); timerDue.delete(id);
        return callback();
      };
      (delay === 5000 ? timers : sessionTimers).set(id, once); return id;
    },
    clearTimeout: id => { timers.delete(id); sessionTimers.delete(id); timerDelays.delete(id); timerDue.delete(id); },
    fetch: (path, options) => new Promise((resolve, reject) => pending.push({path, options, resolve, reject}))});
  function respond(method, path, body, status = 200) {
    const index = pending.findIndex(request => request.path === '/admin/api/' + path && request.options.method === method);
    assert.notEqual(index, -1, `expected ${method} ${path}`);
    const [request] = pending.splice(index, 1);
    request.resolve({ok: status < 400, status, json: async () => body});
    return request;
  }
  const fireNextSessionTimer = () => {
    const [id] = [...sessionTimers.keys()].sort((first, second) => timerDue.get(first) - timerDue.get(second));
    assert.notEqual(id, undefined, 'expected an independently armed session timer');
    clock.now = Math.max(clock.now, timerDue.get(id));
    return sessionTimers.get(id)();
  };
  return {element, panels, pending, timers, sessionTimers, timerDelays, timerDue, fireNextSessionTimer, clock, document, events, location, clipboard, respond, navigate: (name, auto = true) => {
    const result = buttons.find(button => button.dataset.view === name).onclick();
    if (auto && name === 'test' && pending.some(request => request.path === '/admin/api/models')) { respond('GET', 'models', []); respond('GET', 'overview', snapshot); }
    return result;
  }};
}

async function ready() {
  const app = fixture();
  app.respond('GET', 'session', {error: {code: 'unauthorized'}}, 401);
  await tick();
  return app;
}

function hiddenPassword(app) {
  assert.equal(app.element('password').type, 'password');
  assert.equal(app.element('password').value === '', true, 'secret must be cleared');
  assert.equal(app.element('password-toggle').textContent, '显示密码');
  assert.equal(app.element('password-visibility-status').textContent, '密码已隐藏');
}

test('password toggle preserves the input node and value without requesting or submitting', async () => {
  const app = await ready();
  hiddenPassword(app);
  const input = app.element('password');
  const baseline = app.pending.length;
  for (const value of ['', '  controlled <&> 密码 !  ']) {
    input.value = value;
    for (let attempt = 0; attempt < 3; attempt++) {
      app.element('password-toggle').onclick();
      assert.equal(input.type, 'text');
      assert.equal(app.element('password-toggle').textContent, '隐藏密码');
      assert.equal(app.element('password-visibility-status').textContent, '密码已显示');
      assert.equal(input.value === value, true);
      app.element('password-toggle').onclick();
      assert.equal(input.type, 'password');
      assert.equal(app.element('password-toggle').textContent, '显示密码');
      assert.equal(app.element('password-visibility-status').textContent, '密码已隐藏');
      assert.equal(input.value === value, true);
      assert.equal(app.element('password'), input);
      assert.equal(app.pending.length, baseline);
    }
  }
});

test('login success, logout and rejected login reset password visibility', async () => {
  const app = await ready();
  app.element('password').value = 'controlled login input';
  app.element('password-toggle').onclick();
  app.element('login-form').onsubmit({preventDefault() {}, submitter: app.element('login-button')});
  const request = app.respond('POST', 'login', {csrf: 'test-only-csrf'});
  assert.equal(JSON.parse(request.options.body).password === 'controlled login input', true);
  await tick();
  hiddenPassword(app);
  assert.equal(app.element('console').hidden, false);
  app.respond('GET', 'overview', snapshot);
  await tick();
  app.element('password-toggle').onclick();
  app.element('logout').onclick();
  app.respond('POST', 'logout', {});
  await tick();
  hiddenPassword(app);
  assert.equal(app.element('login').hidden, false);
  app.element('password').value = 'wrong controlled input';
  app.element('password-toggle').onclick();
  app.element('login-form').onsubmit({preventDefault() {}, submitter: app.element('login-button')});
  app.respond('POST', 'login', {error: {code: 'unauthorized'}}, 401);
  await tick();
  hiddenPassword(app);
  assert.match(app.element('error').textContent, /管理员密码不正确/);
});

test('pending rejected login keeps the prior session error until its own 401 clears the password', async () => {
  const app = await ready();
  const priorError = app.element('error').textContent;
  assert.equal(priorError, '管理会话已过期，请重新登录；不会自动重发调用。');
  const submitter = app.element('login-button');
  app.element('password').value = 'wrong controlled input';
  app.element('password-toggle').onclick();
  const submitted = app.element('login-form').onsubmit({preventDefault() {}, submitter});
  await tick();
  assert.equal(app.element('error').textContent, priorError);
  assert.equal(app.element('password').type, 'text');
  assert.equal(app.element('password').value === 'wrong controlled input', true, 'pending rejection retains input');
  assert.equal(submitter.disabled, true);
  assert.equal(app.pending.filter(request => request.path === '/admin/api/login').length, 1);
  app.respond('POST', 'login', {error: {code: 'unauthorized'}}, 401);
  await submitted;
  hiddenPassword(app);
  assert.equal(app.element('error').textContent, '管理员密码不正确，请重新输入。');
  assert.equal(submitter.disabled, false);
});

test('session restoration resets an entered visible password', async () => {
  const app = fixture();
  app.element('password').value = 'controlled pending-session input';
  app.element('password-toggle').onclick();
  app.respond('GET', 'session', {csrf: 'test-only-csrf'});
  await tick();
  hiddenPassword(app);
  assert.equal(app.element('login').hidden, true);
  app.respond('GET', 'overview', snapshot);
  await tick();
});

test('upstream form is unavailable until GET completes, then saves entered values', async () => {
  const app = await authenticated();
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
  app.respond('GET', 'overview', snapshot);
  await navigation;
}

async function authenticated(ttl = 0) {
  const app = fixture();
  assert.equal(app.pending.some(request => request.path === '/healthz'), false);
  app.respond('GET', 'session', {csrf: 'test-only-csrf', ...(ttl ? {expires_at: new Date(app.clock.now + ttl).toISOString()} : {})});
  await tick();
  app.respond('GET', 'overview', snapshot);
  await tick();
  return app;
}

function health(app) {
  const index = app.pending.findIndex(request => request.path === '/healthz');
  assert.notEqual(index, -1);
  return app.pending.splice(index, 1)[0];
}

test('overview distinguishes empty, saved and successful persisted configuration on reentry', async () => {
  const app = await authenticated();
  assert.equal(app.element('summary').textContent, '待配置：从接入上游开始');
  const configured = {
    upstream: {configured: true, enabled: true, credential_configured: true},
    models: {total: 2, enabled: 2}, keys: {total: 2, enabled: 1}, call_evidence: null,
  };
  const saved = app.navigate('overview');
  app.respond('GET', 'overview', configured);
  await saved;
  assert.equal(app.element('summary').textContent, '上游配置已保存；本次启动尚无当前配置成功证据，请显式调用核验。');
  await app.navigate('test');
  const restored = app.navigate('overview');
  app.respond('GET', 'overview', {...configured, call_evidence: {request_id: 'controlled-success-id', alias: 'alpha-1280', time: '2001-02-03T04:05:06Z', upstream_revision: 17, scope: '本次服务启动、当前配置'}});
  await restored;
  assert.equal(app.element('summary').textContent, '当前配置曾成功调用 · 请求 controlled-success-id · 模型 alpha-1280 · 调用时间 2001-02-03T04:05:06Z · 配置版本 17 · 本次服务启动、当前配置');
  assert.equal(app.element('overview-upstream').textContent, '已启用 · 凭据已配置');
  assert.equal(app.element('overview-models').textContent, '2 个 / 启用 2');
  assert.equal(app.element('overview-keys').textContent, '2 个 / 启用 1');
});

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
  app.respond('GET', 'overview', snapshot);
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
  app.element('password').value = 'controlled stale input';
  app.element('password-toggle').onclick();
  const old = health(app);
  const navigation = app.navigate('upstream');
  app.respond('GET', 'upstream', {error: {code: 'unauthorized'}}, 401);
  await navigation;
  old.reject(new Error('late health error'));
  await tick();
  assert.equal(app.element('console').hidden, true);
  assert.equal(app.element('runtime-version').textContent, '');
  assert.equal(app.timers.size, 0);
  hiddenPassword(app);
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
  const app = await authenticated();
  const navigation = app.navigate('upstream');
  app.respond('GET', 'upstream', {error: {code: 'storage_failed'}}, 503);
  await navigation;
  assert.match(app.element('error').textContent, /请求未能完成/);
  const retry = app.navigate('upstream');
  assert.equal(app.element('upstream').hidden, true);
  app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true});
  await retry;
  assert.equal(app.element('upstream').hidden, false);
  assert.equal(app.element('upstream-url').value, 'https://controlled.example/v1');
  assert.equal(app.element('error').textContent, '');
});

test('a stale navigation cannot reveal a panel after the user leaves', async () => {
  const app = await authenticated();
  const first = app.navigate('upstream');
  const second = app.navigate('test');
  await second;
  app.respond('GET', 'upstream', {url: '', enabled: false});
  await first;
  assert.equal(app.element('upstream').hidden, true);
  assert.equal(app.element('test').hidden, false);
});

test('rejected save recovers persisted configuration after navigation finishes', async () => {
  const app = await authenticated();
  const savedURL = 'https://controlled.example/v1';
  const first = app.navigate('upstream');
  app.respond('GET', 'upstream', {url: savedURL, enabled: true, credential_configured: true});
  await first;
  app.element('upstream-url').value = 'http://127.0.0.1:1/v1';
  app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save')});
  app.respond('PUT', 'upstream', {error: {code: 'invalid_destination'}}, 400);
  await tick();
  assert.match(app.element('error').textContent, /上游地址不符合/);
  assert.equal(app.element('upstream-url').value, 'http://127.0.0.1:1/v1');
  const overview = app.navigate('overview');
  app.respond('GET', 'overview', {...snapshot, upstream: {configured: true, enabled: true, credential_configured: true}});
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

test('explicit upstream reread retains stale credential status until its own GET settles', async () => {
  const app = await authenticated();
  const savedURL = 'https://controlled.example/v1';
  const navigation = app.navigate('upstream');
  app.respond('GET', 'upstream', {url: savedURL, enabled: true, credential_configured: true});
  await navigation;
  app.element('upstream-url').value = 'https://changed.example/v1';
  app.element('upstream-token').value = '';
  const completed = app.element('refresh-upstream').onclick();
  await tick();
  assert.match(app.element('credential-state').textContent, /凭据已配置/);
  assert.equal(app.element('refresh-upstream').disabled, true);
  assert.equal(app.element('upstream-url').value, 'https://changed.example/v1');
  assert.equal(app.pending.filter(request => request.options.method !== 'GET').length, 0);
  app.respond('GET', 'upstream', {url: savedURL, enabled: true, credential_configured: true});
  await completed;
  assert.equal(app.element('refresh-upstream').disabled, false);
  assert.match(app.element('notice').textContent, /已重新读取上游状态/);
  assert.equal(app.element('upstream-url').value, savedURL);
  assert.equal(app.element('upstream-token').value, '');
  app.element('upstream-url').value = 'http://127.0.0.1:1/v1';
  const rejected = app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save')});
  const write = app.pending.find(request => request.options.method === 'PUT');
  assert.equal(JSON.parse(write.options.body).url, 'http://127.0.0.1:1/v1');
  app.respond('PUT', 'upstream', {error: {code: 'invalid_destination'}}, 400);
  await rejected;
  assert.match(app.element('error').textContent, /上游地址不符合/);
});

async function callable(app) {
  const navigation = app.navigate('test', false);
  app.respond('GET', 'models', [{alias: 'alpha', model: 'provider', enabled: true}, {alias: 'disabled', model: 'other', enabled: false}]);
  app.respond('GET', 'overview', {...snapshot, upstream: {enabled: true, credential_configured: true}, keys: {total: 1, enabled: 1}});
  await navigation;
  app.element('test-model').value = 'alpha';
  app.element('test-key').value = 'mr_isolated';
  app.element('test-message').value = 'isolated message';
}
function send(app) { return app.element('test-form').onsubmit({preventDefault() {}}); }
function generation(app) {
  const index = app.pending.findIndex(request => request.path === '/v1/chat/completions');
  assert.notEqual(index, -1);
  return app.pending.splice(index, 1)[0];
}
function response(id, value) { return {ok: true, headers: {get: () => id}, json: async () => value}; }

async function upstreamReady() {
  const app = await authenticated();
  const loaded = app.navigate('upstream');
  app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true, credential_configured: true}); await loaded;
  return app;
}

for (const failure of [false, true]) test('upstream restore immediately clears secret, restores only confirmed public fields and performs zero writes: ' + failure, async () => {
  const app = await upstreamReady();
  app.element('upstream-url').value = 'https://draft.example/v1'; app.element('upstream-enabled').checked = false; app.element('upstream-token').value = 'controlled-unsaved';
  const restored = app.element('refresh-upstream').onclick();
  assert.equal(app.element('upstream-token').value, '');
  for (const id of ['refresh-upstream', 'save-upstream', 'upstream-url', 'upstream-token', 'upstream-enabled']) assert.equal(app.element(id).disabled, true);
  await app.element('refresh-upstream').onclick();
  await app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save-upstream')});
  assert.equal(app.pending.filter(request => request.path === '/admin/api/upstream').length, 1);
  assert.equal(app.pending.some(request => request.options.method === 'PUT'), false);
  app.respond('GET', 'upstream', failure ? {error: {code: 'storage_unavailable'}} : {url: 'https://controlled.example/v1', enabled: true, credential_configured: true}, failure ? 503 : 200); await restored;
  assert.equal(app.element('upstream-url').value, failure ? 'https://draft.example/v1' : 'https://controlled.example/v1');
  assert.equal(app.element('upstream-enabled').checked, !failure);
  assert.equal(app.element('upstream-token').value, '');
  if (failure) assert.match(app.element('error').textContent, /存储不可用/);
  else assert.match(app.element('notice').textContent, /撤销本页未保存修改/);
  assert.equal(app.element('save-upstream').disabled, false);
});

test('upstream save owns both controls through its readback; restore and repeated save cannot overlap', async () => {
  const app = await upstreamReady(); app.element('upstream-token').value = 'controlled-save';
  const saved = app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save-upstream')});
  await app.element('refresh-upstream').onclick();
  await app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save-upstream')});
  assert.equal(app.pending.length, 2);
  assert.equal(app.pending.filter(request => request.options.method === 'PUT').length, 1);
  app.respond('PUT', 'upstream', {}); await tick();
  assert.equal(app.element('refresh-upstream').disabled, true);
  await app.element('refresh-upstream').onclick();
  assert.equal(app.pending.filter(request => request.path === '/admin/api/upstream').length, 1);
  app.respond('GET', 'upstream', {error: {code: 'storage_unavailable'}}, 503); await saved;
  assert.match(app.element('error').textContent, /已保存，但最新读取未确认/);
  assert.equal(app.element('refresh-upstream').disabled, false);
});

test('late upstream restore cannot replace a new draft or unlock a newer operation', async () => {
  const app = await upstreamReady(); const old = app.element('refresh-upstream').onclick();
  const oldRead = app.pending.splice(app.pending.findIndex(request => request.path === '/admin/api/upstream'), 1)[0];
  const reopened = app.navigate('upstream'); app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true}); await reopened;
  app.element('upstream-url').value = 'https://new.example/v1'; app.element('upstream-token').value = 'controlled-new';
  const saved = app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save-upstream')});
  oldRead.resolve({ok: true, status: 200, json: async () => ({url: 'https://old.example/v1', enabled: false})}); await old;
  assert.equal(app.element('upstream-url').value, 'https://new.example/v1');
  assert.equal(app.element('upstream-token').value === 'controlled-new', true);
  assert.equal(app.element('refresh-upstream').disabled, true);
  app.respond('PUT', 'upstream', {}); await tick(); app.respond('GET', 'upstream', {url: 'https://new.example/v1', enabled: true}); await saved;
  assert.equal(app.element('refresh-upstream').disabled, false);
});

for (const boundary of ['pagehide', 'uncertain session']) test('upstream busy controls release on owner invalidation without applying a late restore: ' + boundary, async () => {
  const app = await upstreamReady(); const restored = app.element('refresh-upstream').onclick();
  if (boundary === 'pagehide') app.events.get('pagehide')();
  else {
    const checking = app.fireNextSessionTimer();
    const index = app.pending.findIndex(request => request.path === '/admin/api/session');
    app.pending.splice(index, 1)[0].reject(new Error('offline')); await checking;
  }
  app.element('upstream-url').value = 'https://new-owner.example/v1';
  app.respond('GET', 'upstream', {url: 'https://old-owner.example/v1', enabled: false}); await restored;
  assert.equal(app.element('upstream-url').value, 'https://new-owner.example/v1');
  for (const id of ['upstream-url', 'upstream-token', 'upstream-enabled', 'save-upstream', 'refresh-upstream']) assert.equal(app.element(id).disabled, false);
});

for (const terminal of ['complete', 'failed', 'cancel']) test('in-flight submitted controls stay visible and disabled until current call terminates: ' + terminal, async () => {
  const app = await authenticated(); await callable(app);
  const completed = send(app); const request = generation(app);
  for (const id of ['test-model', 'test-key', 'test-message', 'test-stream']) assert.equal(app.element(id).disabled, true);
  assert.equal(app.element('test-key').value === 'mr_isolated', true); assert.equal(app.element('test-message').value === 'isolated message', true);
  const payload = JSON.parse(request.options.body); assert.equal(payload.model, app.element('test-model').value); assert.equal(payload.stream, app.element('test-stream').checked);
  if (terminal === 'cancel') app.element('cancel').onclick();
  request.resolve(terminal === 'failed' ? {ok: false, headers: {get: () => 'rejected'}, json: async () => ({error: {code: 'unauthorized'}})} : response('done', {choices: [{message: {content: 'controlled reply'}}]})); await completed;
  for (const id of ['test-model', 'test-key', 'test-message', 'test-stream']) assert.equal(app.element(id).disabled, false);
  if (terminal !== 'complete') assert.equal(app.element('test-key').value + app.element('test-message').value, '');
});

test('late canceled-call finally cannot unlock new submitted controls', async () => {
  const app = await authenticated(); await callable(app);
  const old = send(app); const previous = generation(app); app.element('cancel').onclick();
  app.element('test-key').value = 'controlled-new'; app.element('test-message').value = 'controlled-new-message';
  const newer = send(app); const request = generation(app);
  previous.resolve(response('old', {choices: [{message: {content: 'late'}}]})); await old;
  for (const id of ['test-model', 'test-key', 'test-message', 'test-stream']) assert.equal(app.element(id).disabled, true);
  assert.equal(app.element('test-key').value === 'controlled-new', true);
  request.resolve(response('new', {choices: [{message: {content: 'current'}}]})); await newer;
  assert.equal(app.element('test-key').disabled, false);
});

test('key list displays server created time and explicitly unknown, never a client timestamp', async () => {
  const app = await authenticated(); const loaded = app.navigate('keys');
  app.respond('GET', 'keys', [{id: 'one', name: 'same', hint: 'safe', enabled: true, created: '2001-02-03T04:05:06Z'}, {id: 'two', name: 'same', hint: 'safe', enabled: false, created: null}]); await loaded;
  assert.match(app.element('key-list').children[0].children[0].textContent, /创建时间 2001-02-03T04:05:06Z/);
  assert.match(app.element('key-list').children[1].children[0].textContent, /创建时间 未知/);
});

for (const entity of ['models', 'keys']) for (const moved of [false, true]) test('state-change replacement restores only same identity unless focus moved: ' + entity + '/' + moved, async () => {
  const app = await authenticated();
  const rows = entity === 'models' ? [{alias: 'alpha', model: 'provider', enabled: true}, {alias: 'beta', model: 'provider', enabled: true}] : [{id: 'one', name: 'duplicate', enabled: true}, {id: 'two', name: 'duplicate', enabled: true}];
  const loaded = app.navigate(entity); app.respond('GET', entity, rows); await tick();
  if (entity === 'models') app.respond('GET', 'upstream', {enabled: true}); await loaded;
  const list = app.element(entity === 'models' ? 'model-list' : 'key-list'); const original = list.children[0].children[1]; original.focus();
  const changed = original.onclick();
  if (moved) app.element('key-name').focus();
  app.respond('PATCH', entity + '/' + (entity === 'models' ? 'alpha' : 'one'), {}); await tick();
  app.respond('GET', entity, [rows[1], {...rows[0], enabled: false}]); await tick();
  if (entity === 'models') app.respond('GET', 'upstream', {enabled: true}); await changed;
  assert.equal(app.document.activeElement, moved ? app.element('key-name') : list.children[1].children[1]);
  assert.notEqual(list.children[1].children[1], original);
});

test('late state change on departed page cannot restore old row focus', async () => {
  const app = await authenticated(); const loaded = app.navigate('keys');
  app.respond('GET', 'keys', [{id: 'one', name: 'key', enabled: true}]); await loaded;
  const button = app.element('key-list').children[0].children[1]; button.focus(); const changed = button.onclick();
  const navigation = app.navigate('upstream'); app.respond('GET', 'upstream', {url: '', enabled: false}); await navigation; app.element('upstream-url').focus();
  app.respond('PATCH', 'keys/one', {}); await changed;
  assert.equal(app.document.activeElement, app.element('upstream-url'));
  assert.equal(app.pending.some(request => request.path === '/admin/api/keys'), false);
});

for (const failure of ['rejected', 'network', 'refresh']) {
  test(`upstream ${failure} clears credential, preserves public draft and distinguishes acknowledged write`, async () => {
    const app = await authenticated();
    const navigation = app.navigate('upstream');
    app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true, credential_configured: true});
    await navigation;
    app.element('upstream-url').value = 'https://changed.example/v1';
    app.element('upstream-token').value = 'fixture secret';
    const completed = app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save')});
    if (failure === 'network') app.pending.splice(app.pending.findIndex(request => request.options.method === 'PUT'), 1)[0].reject(new Error('network'));
    else app.respond('PUT', 'upstream', failure === 'rejected' ? {error: {code: 'new_credential_required'}} : {}, failure === 'rejected' ? 400 : 200);
    if (failure === 'refresh') { await tick(); app.respond('GET', 'upstream', {error: {code: 'storage_unavailable'}}, 503); }
    await completed;
    assert.equal(app.element('upstream-token').value, '');
    assert.equal(app.element('upstream-url').value, 'https://changed.example/v1');
    assert.equal(app.element('upstream-enabled').checked, true);
    assert.equal(app.element('error').textContent.includes('已保存'), failure === 'refresh');
    assert.match(app.element('error').textContent, failure === 'refresh' ? /最新读取未确认.*存储不可用/ : /凭据已清除.*请重新输入/);
    assert.equal(app.pending.filter(request => request.options.method === 'PUT').length, 0, 'no automatic rewrite');
  });
}

for (const failure of ['application', 'provider', 'network', 'protocol', 'stream']) {
  test(`generation ${failure} failure clears only sensitive inputs and keeps diagnosis and identity`, async () => {
    const app = await authenticated(); await callable(app);
    app.element('test-stream').checked = failure === 'stream';
    const completed = send(app), request = generation(app);
    if (failure === 'network') request.reject(new Error('network'));
    else if (failure === 'protocol') request.resolve({ok: true, headers: {get: () => 'safe-id'}, json: async () => { throw new SyntaxError('fixture'); }});
    else if (failure === 'stream') request.resolve({ok: true, headers: {get: () => 'safe-id'}, body: {getReader: () => ({read: async () => ({done: true}), cancel: async () => {}})}});
    else request.resolve({ok: false, status: 401, headers: {get: () => 'safe-id'}, json: async () => ({error: {code: failure === 'application' ? 'unauthorized' : 'upstream_rejected'}})});
    await completed;
    assert.equal(app.element('test-key').value, ''); assert.equal(app.element('test-message').value, '');
    assert.equal(app.element('test-model').value, 'alpha'); assert.equal(app.element('console').hidden, false);
    assert.match(app.element('test-status').textContent, /失败/); assert.match(app.element('error').textContent, /已清除.*重新输入/);
    assert.equal(app.element('view-record').hidden, failure === 'network');
    if (failure === 'application' || failure === 'provider') assert.match(app.element('test-raw').textContent, /error/);
  });
}

test('overview evidence action uses displayed ID, while generic records remains unfiltered', async () => {
  const app = await authenticated(); await callable(app);
  const completed = send(app); generation(app).resolve(response('different-local-id', {choices: []})); await completed;
  const navigation = app.navigate('overview');
  const evidenceID = 'QAOnlyRecordIdentity01';
  app.respond('GET', 'overview', {...snapshot, upstream: {configured: true, enabled: true, credential_configured: true}, call_evidence: {request_id: evidenceID, alias: 'alpha', scope: '本次服务启动、当前配置'}});
  await navigation;
  assert.match(app.element('summary').textContent, /模型 alpha.*本次服务启动/);
  const exact = app.element('overview-record').onclick();
  app.respond('GET', 'requests', {data: [], next_cursor: 'older'}); await tick();
  assert.equal(app.element('find-request-id').value, evidenceID);
  app.respond('GET', 'requests/' + evidenceID, {id: evidenceID, alias: 'alpha', state: 'success', duration_ms: 0, prompt_tokens: 0}); await exact;
  assert.match(app.element('record-status').textContent, new RegExp(evidenceID));
  const generic = app.navigate('records'); app.respond('GET', 'requests', {data: [], next_cursor: null}); await generic;
  assert.equal(app.element('find-request-id').value, '');
  assert.equal(app.element('record-detail').textContent, '');
  assert.equal(app.pending.some(request => request.path.includes('/requests/')), false);
});

for (const lifecycle of ['session restore', 'new login']) {
  test(`overview precise evidence action survives ${lifecycle} with no local generation identity`, async () => {
    const app = lifecycle === 'session restore' ? fixture() : await ready();
    if (lifecycle === 'new login') app.element('login-form').onsubmit({preventDefault() {}, submitter: app.element('login-button')});
    app.respond(lifecycle === 'session restore' ? 'GET' : 'POST', lifecycle === 'session restore' ? 'session' : 'login', {csrf: 'fixture'}); await tick();
    const id = 'QAOnlyRecordIdentity01';
    app.respond('GET', 'overview', {...snapshot, upstream: {configured: true}, call_evidence: {request_id: id, alias: 'alpha', scope: '本次服务启动、当前配置'}}); await tick();
    const navigation = app.element('overview-record').onclick();
    app.respond('GET', 'requests', {data: [], next_cursor: 'older'}); await tick();
    app.respond('GET', 'requests/' + id, {id, alias: 'alpha', model: 'provider', state: 'success', duration_ms: 1, prompt_tokens: 0, completion_tokens: null, total_tokens: null}); await navigation;
    assert.equal(app.element('find-request-id').value, id); assert.match(app.element('record-status').textContent, new RegExp(id));
    assert.match(app.element('record-detail').textContent, /0 \/ 未知 \/ 未知/);
  });
}

for (const failure of ['rejected', 'refresh']) {
  test(`late upstream ${failure} failure cannot clear new-page credential or notice`, async () => {
    const app = await authenticated();
    const first = app.navigate('upstream'); app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true, credential_configured: true}); await first;
    app.element('upstream-token').value = 'old-fixture';
    const completed = app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save')});
    let oldRead;
    if (failure === 'refresh') { app.respond('PUT', 'upstream', {}); await tick(); oldRead = app.pending.splice(app.pending.findIndex(request => request.path === '/admin/api/upstream'), 1)[0]; }
    const newer = app.navigate('upstream'); app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true, credential_configured: true}); await newer;
    app.element('upstream-token').value = 'new-fixture'; app.element('notice').textContent = 'new owner notice';
    if (failure === 'rejected') app.respond('PUT', 'upstream', {error: {code: 'invalid_destination'}}, 400);
    else oldRead.reject(new Error('network'));
    await completed;
    assert.equal(app.element('upstream-token').value === 'new-fixture', true); assert.equal(app.element('notice').textContent, 'new owner notice');
    assert.equal(app.element('error').textContent, '');
  });
}

test('overview success metadata comes from the request evidence, never the clock or global revision', async () => {
  const app = await authenticated();
  const evidence = {request_id: 'EvidenceMetadataOnly01', alias: 'alpha', time: '2001-02-03T04:05:06Z', upstream_revision: 17, scope: '本次服务启动、当前配置'};
  const loaded = app.navigate('overview');
  app.respond('GET', 'overview', {...snapshot, upstream: {configured: true, revision: 900}, call_evidence: evidence});
  await loaded;
  assert.equal(app.element('summary').textContent, '当前配置曾成功调用 · 请求 ' + evidence.request_id + ' · 模型 alpha · 调用时间 ' + evidence.time + ' · 配置版本 17 · ' + evidence.scope);
  const refreshed = app.element('refresh-overview').onclick();
  assert.equal(app.element('summary').textContent, '加载中');
  assert.equal(app.element('overview-record').disabled, true);
  app.respond('GET', 'overview', {...snapshot, upstream: {configured: true, revision: 901}, call_evidence: {...evidence, time: null, upstream_revision: null}});
  await refreshed;
  assert.match(app.element('summary').textContent, /调用时间 未知 · 配置版本 未知/);
  assert.equal(app.element('summary').textContent.includes(evidence.time), false);
  assert.equal(app.element('summary').textContent.includes('901'), false);
  const changed = app.element('refresh-overview').onclick();
  app.respond('GET', 'overview', {...snapshot, upstream: {configured: true, revision: 902}, call_evidence: null});
  await changed;
  assert.equal(app.element('overview-record').disabled, true);
  assert.match(app.element('summary').textContent, /尚无当前配置成功证据/);
  assert.equal(/调用时间|配置版本|EvidenceMetadataOnly01/.test(app.element('summary').textContent), false);
  const degraded = app.element('refresh-overview').onclick();
  app.respond('GET', 'overview', {...snapshot, upstream: {configured: true}, degraded: true, call_evidence: evidence});
  await degraded;
  assert.equal(app.element('overview-record').disabled, true);
  assert.equal(app.element('summary').textContent.includes(evidence.time), false);
  assert.equal(app.element('summary').textContent.includes('配置版本'), false);
});

test('late overview metadata cannot replace a newer response', async () => {
  const app = await authenticated();
  const older = app.navigate('overview');
  const oldRead = app.pending.splice(app.pending.findIndex(request => request.path === '/admin/api/overview'), 1)[0];
  const newer = app.navigate('overview');
  app.respond('GET', 'overview', {...snapshot, upstream: {configured: true}, call_evidence: {request_id: 'NewEvidenceMetadata01', alias: 'alpha', time: '2002-03-04T05:06:07Z', upstream_revision: 18, scope: 'current'}});
  await newer;
  const expected = app.element('summary').textContent;
  assert.match(expected, /调用时间 2002-03-04T05:06:07Z · 配置版本 18/);
  oldRead.resolve({ok: true, status: 200, json: async () => ({...snapshot, upstream: {configured: true}, call_evidence: {request_id: 'OldEvidenceMetadata01', alias: 'old', time: '2001-02-03T04:05:06Z', upstream_revision: 17, scope: 'old'}})});
  await older;
  assert.equal(app.element('summary').textContent, expected);
});

test('failed overview refresh disables and discards the old precise evidence action', async () => {
  const app = await authenticated();
  const loaded = app.navigate('overview'); app.respond('GET', 'overview', {...snapshot, upstream: {configured: true}, call_evidence: {request_id: 'QAOnlyRecordIdentity01', alias: 'alpha', time: '2001-02-03T04:05:06Z', upstream_revision: 17, scope: '本次启动'}}); await loaded;
  assert.equal(app.element('overview-record').disabled, false);
  const failed = app.element('refresh-overview').onclick(); app.respond('GET', 'overview', {error: {code: 'storage_unavailable'}}, 503); await failed;
  assert.equal(app.element('overview-record').disabled, true); assert.match(app.element('summary').textContent, /状态未知/);
  assert.equal(/调用时间|配置版本|2001-02-03/.test(app.element('summary').textContent), false);
  const requestsBefore = app.pending.length; app.element('overview-record').onclick();
  assert.equal(app.pending.length, requestsBefore, 'no old evidence fallback');
});

test('model selection, placeholder example, readable text and request ID are independent of raw JSON', async () => {
  const app = await authenticated(); await callable(app);
  assert.equal(app.element('test-model').children.length, 1);
  assert.match(app.element('client-example').textContent, /<APPLICATION_KEY>/);
  assert.equal(app.element('client-example').textContent.includes('mr_isolated'), false);
  const pending = send(app); const request = generation(app);
  send(app);
  assert.equal(app.pending.filter(value => value.path === '/v1/chat/completions').length, 0);
  assert.equal(app.element('send').disabled, true);
  request.resolve(response('same-request', {choices: [{message: {content: '<b>plain</b>'}}], usage: {prompt_tokens: 0}}));
  await pending;
  assert.equal(app.element('test-output').textContent, '<b>plain</b>');
  assert.match(app.element('test-raw').textContent, /prompt_tokens/);
  assert.match(app.element('test-request-id').textContent, /same-request/);
  assert.equal(app.element('view-record').hidden, false);
});

for (const mode of ['headers', 'json', 'failure']) {
  test(`cancel isolates late ${mode}, old finally cannot enable a new request`, async () => {
    const app = await authenticated(); await callable(app);
    const oldDone = send(app), old = generation(app);
    let release;
    if (mode === 'json') { old.resolve({ok: true, headers: {get: () => 'old-id'}, json: () => new Promise(resolve => { release = resolve; })}); await tick(); }
    app.element('cancel').onclick();
    assert.equal(old.options.signal.aborted, true);
    assert.equal(app.element('test-key').value, ''); assert.equal(app.element('test-message').value, '');
    assert.match(app.element('test-status').textContent, /已取消/);
    app.element('test-key').value = 'mr_new'; app.element('test-message').value = 'new';
    const freshDone = send(app), fresh = generation(app);
    if (mode === 'json') release({choices: [{message: {content: 'old'}}]});
    else if (mode === 'failure') old.reject(new Error('old error'));
    else old.resolve(response('old-id', {choices: [{message: {content: 'old'}}]}));
    await oldDone;
    assert.equal(app.element('test-output').textContent, ''); assert.equal(app.element('send').disabled, true); assert.equal(app.element('cancel').disabled, false); assert.equal(app.element('error').textContent, '');
    fresh.resolve(response('new-id', {choices: [{message: {content: 'new'}}]})); await freshDone;
    assert.equal(app.element('test-output').textContent, 'new'); assert.match(app.element('test-request-id').textContent, /new-id/);
  });
}

test('SSE decodes split UTF8 and CRLF, presents increments before DONE and isolates canceled reads', async () => {
  const app = await authenticated(); await callable(app); app.element('test-stream').checked = true;
  const completed = send(app), request = generation(app), reads = [];
  const reader = {read: () => new Promise(resolve => reads.push(resolve)), cancel: async () => {}};
  request.resolve({...response('stream-id', {}), body: {getReader: () => reader}}); await tick();
  const bytes = new TextEncoder().encode('data: {"choices":[{"delta":{"content":"你好"}}]}\r\n\r\n');
  const split = bytes.indexOf(0xe4) + 1;
  reads.shift()({value: bytes.slice(0, split), done: false}); await tick();
  assert.equal(app.element('test-output').textContent, '');
  reads.shift()({value: bytes.slice(split), done: false}); await tick();
  assert.equal(app.element('test-output').textContent, '你好');
  app.element('cancel').onclick();
  reads.shift()({value: new TextEncoder().encode('data: {"choices":[{"delta":{"content":"late"}}]}\n\ndata: [DONE]\n\n'), done: false});
  await completed;
  assert.equal(app.element('test-output').textContent, ''); assert.match(app.element('test-status').textContent, /已取消/);
});

test('tool-only output is not treated as network failure; unknown usage stays unknown and zero stays zero', async () => {
  const app = await authenticated(); await callable(app);
  const done = send(app); generation(app).resolve(response('tool-id', {choices: [{index: 0, message: {role: 'assistant', content: null, tool_calls: [{id: 'call_tool', type: 'function', function: {name: 'lookup_weather', arguments: '{"city":"Paris"}'}}]}, finish_reason: 'tool_calls'}]})); await done;
  assert.match(app.element('test-output').textContent, /返回1个工具调用：lookup_weather/); assert.match(app.element('test-output').textContent, /未执行工具/); assert.doesNotMatch(app.element('test-output').textContent, /Paris/); assert.equal(app.pending.every(request => request.path === '/healthz'), true); assert.match(app.element('test-raw').textContent, /tool_calls/);
  const navigate = app.element('view-record').onclick();
  app.respond('GET', 'requests', {data: [{id: 'tool-id', state: 'success', prompt_tokens: 0, completion_tokens: null}], next_cursor: null}); await tick();
  app.respond('GET', 'requests/tool-id', {id: 'tool-id', state: 'success', prompt_tokens: 0, completion_tokens: null}); await navigate;
  assert.match(app.element('record-detail').textContent, /0 \/ 未知 \/ 未知/);
});

test('late key issuance and management 401 cannot affect a new page or login', async () => {
  const app = await authenticated(); const keys = app.navigate('keys'); app.respond('GET', 'keys', []); await keys;
  app.element('key-form').onsubmit({preventDefault() {}, submitter: app.element('issue-button')});
  await app.navigate('test'); app.respond('POST', 'keys', {key: 'mr_late_secret'}); await tick();
  assert.equal(app.element('issued-key').value, ''); assert.equal(app.element('issued').hidden, true);
  const upstream = app.navigate('upstream'); await app.navigate('test'); app.respond('GET', 'upstream', {error: {code: 'unauthorized'}}, 401); await upstream;
  assert.equal(app.element('console').hidden, false); assert.equal(app.element('error').textContent, '');
});

test('clipboard completion is isolated after closing, reissuing or leaving; failures have a manual path', async () => {
  const app = await authenticated(); const keys = app.navigate('keys'); app.respond('GET', 'keys', []); await keys;
  app.element('issued-key').value = 'mr_isolated'; app.element('issued').hidden = false;
  let complete;
  app.clipboard.writeText = () => new Promise(resolve => { complete = resolve; });
  app.element('copy-key').onclick(); app.element('close-key').onclick();
  app.element('notice').textContent = 'new notification'; complete(); await tick();
  assert.equal(app.element('notice').textContent, 'new notification');
  app.element('issued-key').value = 'mr_new'; app.element('issued').hidden = false;
  app.clipboard.writeText = async () => { throw new Error('blocked'); };
  app.element('copy-key').onclick(); await tick(); assert.match(app.element('error').textContent, /手动选择复制/);
  assert.equal(app.element('issued-key').value === 'mr_new', true);
  assert.equal(app.element('issued').hidden, false);
  assert.equal(app.element('copy-key').disabled, false);
  let copied = false;
  app.clipboard.writeText = async value => { copied = value === 'mr_new'; };
  app.element('copy-key').onclick(); await tick();
  assert.equal(copied, true);
  assert.equal(app.element('notice').textContent, '已复制，请保管');
  assert.equal(app.element('error').textContent, '');
  let reject;
  app.clipboard.writeText = () => new Promise((_resolve, failure) => { reject = failure; });
  app.element('copy-key').onclick(); await app.navigate('test'); reject(new Error('late')); await tick();
  assert.equal(app.element('error').textContent, ''); assert.equal(app.element('issued-key').value, '');
});

test('SSE DONE succeeds, missing DONE and event error fail without a second request', async () => {
  for (const mode of ['done', 'missing', 'error']) {
    const app = await authenticated(); await callable(app); app.element('test-stream').checked = true;
    const done = send(app), request = generation(app);
    const suffix = mode === 'done' ? 'data: [DONE]\n\n' : mode === 'error' ? 'event: error\ndata: {"error":{"code":"stream_interrupted"}}\n\n' : '';
    let sent = false;
    const reader = {read: async () => sent ? {done: true} : (sent = true, {value: new TextEncoder().encode('data: {"choices":[{"delta":{"content":"partial"}}]}\n\n' + suffix), done: false}), cancel: async () => {}};
    request.resolve({...response('sse-id', {}), body: {getReader: () => reader}}); await done;
    assert.equal(app.element('test-output').textContent, 'partial');
    assert.match(app.element('test-status').textContent, mode === 'done' ? /完成/ : /失败/);
    assert.equal(app.pending.filter(value => value.path === '/v1/chat/completions').length, 0);
  }
});

test('overview query failure remains unknown rather than zero and retry recovers', async () => {
  const app = await authenticated();
  const navigation = app.navigate('overview'); app.respond('GET', 'overview', {error: {code: 'storage_unavailable'}}, 503); await navigation;
  assert.equal(app.element('overview-models').textContent, '未知'); assert.match(app.element('summary').textContent, /状态未知/);
  await overview(app); assert.equal(app.element('overview-models').textContent, '0 个 / 启用 0');
});

test('creation cancellation preserves public drafts with zero writes', async () => {
  const app = await authenticated(); app.element('model-alias').value = 'draft'; app.element('key-name').value = 'draft purpose';
  const before = app.pending.length;
  app.element('cancel-model-create').onclick(); app.element('cancel-key-create').onclick();
  assert.equal(app.element('model-alias').value, 'draft'); assert.equal(app.element('key-name').value, 'draft purpose');
  assert.equal(app.pending.length, before);
});

test('model edit keeps canceled draft and a pending save cannot be replaced by another editor', async () => {
  const app = await authenticated();
  const models = [{alias: 'alpha', model: 'first', enabled: true}, {alias: 'beta', model: 'second', enabled: true}];
  const navigation = app.navigate('models'); app.respond('GET', 'models', models); await tick(); app.respond('GET', 'upstream', {enabled: true, credential_configured: true}); await navigation;
  const edit = index => app.element('model-list').children[index].children[2].onclick();
  await edit(0); app.element('edit-real').value = 'draft'; app.element('cancel-edit').onclick(); await edit(0);
  assert.equal(app.element('edit-real').value, 'draft');
  const saved = app.element('model-edit-form').onsubmit({preventDefault() {}, submitter: app.element('save-edit')});
  assert.equal(app.element('cancel-edit').disabled, true);
  await edit(1); assert.equal(app.element('edit-alias').value, 'alpha');
  app.respond('PATCH', 'models/alpha', {ok: true}); await tick();
  app.respond('GET', 'models', models); await tick(); app.respond('GET', 'upstream', {enabled: true, credential_configured: true}); await saved;
  assert.equal(app.element('model-edit').hidden, true); assert.equal(app.element('cancel-edit').disabled, false);
});

test('unknown navigation hashes normalize to a safe overview target', async () => {
  const app = await authenticated(); app.location.hash = '#unknown-target';
  const navigation = app.events.get('hashchange')(); app.respond('GET', 'overview', snapshot); await navigation;
  assert.equal(app.location.hash, '#overview'); assert.equal(app.element('overview').hidden, false);
});

test('expired session clears current generation and restores safe hash target without replay', async () => {
  const app = await authenticated(); await callable(app);
  const pending = send(app), request = generation(app);
  const check = [...app.sessionTimers.values()][0]; check(); app.respond('GET', 'session', {error: {code: 'unauthorized'}}, 401); await tick();
  assert.equal(app.location.hash, '#test'); assert.equal(app.element('test-key').value, ''); assert.equal(app.element('test-message').value, ''); assert.equal(app.element('console').hidden, true);
  request.resolve(response('expired-id', {choices: [{message: {content: 'late'}}]})); await pending;
  assert.equal(app.element('test-output').textContent, ''); assert.match(app.element('error').textContent, /管理会话已过期/);
});

test('stream tool fragments summarize only after DONE without executing arguments', async () => {
  const app = await authenticated(); await callable(app); app.element('test-stream').checked = true;
  const reads = []; const completed = send(app);
  generation(app).resolve({ok: true, headers: {get: () => 'stream-tools'}, body: {getReader: () => ({read: () => new Promise(resolve => reads.push(resolve)), cancel: async () => {}})}});
  await tick();
  const frame = value => 'data: ' + JSON.stringify(value) + '\n\n';
  reads.shift()({value: new TextEncoder().encode(frame({choices: [{index: 0, delta: {role: 'assistant', content: null, tool_calls: [{index: 0, id: 'call_weather', type: 'function', function: {name: 'lookup_', arguments: ''}}]}, finish_reason: null}]})), done: false}); await tick();
  assert.equal(app.element('test-output').textContent, '');
  reads.shift()({value: new TextEncoder().encode(frame({choices: [{index: 0, delta: {tool_calls: [{index: 0, function: {name: 'weather', arguments: '{"city":"Paris"}'}}, {index: 1, id: 'call_clock', type: 'function', function: {name: 'clock', arguments: '{}'}}]}, finish_reason: null}]})), done: false}); await tick();
  assert.equal(app.element('test-output').textContent, '');
  reads.shift()({value: new TextEncoder().encode(frame({choices: [{index: 0, delta: {}, finish_reason: 'tool_calls'}]}) + 'data: [DONE]\n\n'), done: false}); await completed;
  assert.match(app.element('test-output').textContent, /返回2个工具调用：lookup_weather、clock/);
  assert.match(app.element('test-output').textContent, /未执行工具/); assert.doesNotMatch(app.element('test-output').textContent, /Paris/);
  assert.equal(app.pending.every(request => request.path === '/healthz'), true);
});

test('empty content remains distinct and tool names/count storage are bounded', async () => {
  const app = await authenticated(); await callable(app);
  let completed = send(app); generation(app).resolve(response('empty', {choices: [{message: {role: 'assistant', content: ''}}]})); await completed;
  assert.match(app.element('test-output').textContent, /没有文本输出或工具调用/);
  completed = send(app);
  const tools = Array.from({length: 1100}, (_, index) => ({id: 'call_' + index, type: 'function', function: {name: 'name_' + index, arguments: '{}'}}));
  generation(app).resolve(response('bounded', {choices: [{index: 0, message: {role: 'assistant', content: null, tool_calls: tools}, finish_reason: 'tool_calls'}]})); await completed;
  assert.match(app.element('test-output').textContent, /至少1024个工具调用/); assert.match(app.element('test-output').textContent, /名称仅展示前5个/);
  assert.doesNotMatch(app.element('test-output').textContent, /name_5/); assert.ok(app.element('test-output').textContent.length < 800);
});

test('visible long stream checks at 30s; network uncertainty clears secrets, isolates late output and requires manual recovery', async () => {
  const app = await authenticated(); await callable(app); app.element('test-stream').checked = true;
  const reads = []; const completed = send(app); const request = generation(app);
  request.resolve({ok: true, headers: {get: () => 'uncertain-id'}, body: {getReader: () => ({read: () => new Promise(resolve => reads.push(resolve)), cancel: async () => {}})}}); await tick();
  reads.shift()({value: new TextEncoder().encode('data: {"choices":[{"index":0,"delta":{"content":"private output"}}]}\n\n'), done: false}); await tick();
  assert.equal([...app.sessionTimers.keys()].some(id => app.timerDelays.get(id) === 30000), true);
  const check = [...app.sessionTimers.values()][0]; const checking = check();
  const index = app.pending.findIndex(value => value.path === '/admin/api/session'); app.pending.splice(index, 1)[0].reject(new Error('offline')); await checking;
  for (const id of ['test-key', 'test-message']) assert.equal(app.element(id).value, '');
  for (const id of ['test-output', 'test-raw']) assert.equal(app.element(id).textContent, '');
  assert.equal(request.options.signal.aborted, true); assert.equal(app.element('console').hidden, false);
  assert.equal(app.element('test-model').value, 'alpha'); assert.equal(app.location.hash, '#test'); assert.equal(app.element('send').disabled, true);
  assert.match(app.element('error').textContent, /暂无法确认/); assert.doesNotMatch(app.element('error').textContent, /已过期/);
  assert.equal(app.sessionTimers.size, 0); assert.equal(send(app), undefined);
  app.events.get('visibilitychange')(); app.events.get('pageshow')();
  assert.equal(app.pending.some(value => value.path === '/admin/api/session'), false);
  reads.shift()({value: new TextEncoder().encode('data: {"choices":[{"delta":{"content":"late"}}]}\n\ndata: [DONE]\n\n'), done: false}); await completed;
  assert.equal(app.element('test-output').textContent, '');
  const recovery = app.element('check-session').onclick(); app.respond('GET', 'session', {csrf: 'controlled', expires_at: new Date(Date.now() + 60000).toISOString()}); await recovery;
  assert.equal(app.element('send').disabled, false); assert.equal(app.element('test-message').value, '');
  assert.equal(app.pending.some(value => value.path === '/v1/chat/completions'), false);
});

test('known expiry cancels without a session fetch; relogin restores target without replay', async () => {
  const app = await authenticated(); await callable(app);
  const checking = app.element('check-session').onclick(); app.respond('GET', 'session', {csrf: 'controlled', expires_at: new Date(app.clock.now + 1).toISOString()}); await checking;
  const completed = send(app); const request = generation(app); app.clock.now += 2; await [...app.sessionTimers.values()][0]();
  assert.equal(app.element('login').hidden, false); assert.equal(request.options.signal.aborted, true); assert.equal(app.location.hash, '#test');
  assert.equal(app.pending.some(value => value.path === '/admin/api/session'), false);
  request.resolve(response('old', {choices: [{message: {role: 'assistant', content: 'late'}}]})); await completed;
  app.element('password').value = 'controlled'; const login = app.element('login-form').onsubmit({preventDefault() {}, submitter: {disabled: false}});
  app.respond('POST', 'login', {csrf: 'new', expires_at: new Date(Date.now() + 60000).toISOString()}); await tick();
  app.respond('GET', 'models', [{alias: 'alpha', enabled: true}]); app.respond('GET', 'overview', {...snapshot, upstream: {enabled: true, credential_configured: true}, keys: {enabled: 1}}); await login;
  assert.equal(app.element('console').hidden, false); assert.equal(app.location.hash, '#test'); assert.equal(app.element('test-output').textContent, '');
  assert.equal(app.pending.some(value => value.path === '/v1/chat/completions'), false);
});

test('uncertain session invalidates late issuance and explicit relogin clears and preserves safe target', async () => {
  const app = await authenticated(); const navigation = app.navigate('keys'); app.respond('GET', 'keys', []); await navigation;
  app.element('key-name').value = 'controlled'; const issued = app.element('key-form').onsubmit({preventDefault() {}, submitter: {disabled: false}});
  const checking = app.element('check-session').onclick(); const index = app.pending.findIndex(value => value.path === '/admin/api/session'); app.pending.splice(index, 1)[0].reject(new Error('offline')); await checking;
  app.respond('POST', 'keys', {key: 'mr_late_controlled'}); await issued;
  assert.equal(app.element('issued-key').value, ''); assert.equal(app.element('issued').hidden, true);
  app.element('relogin').onclick(); assert.equal(app.element('login').hidden, false); assert.equal(app.location.hash, '#keys');
  assert.equal(app.pending.some(value => value.path === '/v1/chat/completions'), false);
});

test('known expiry also blocks submit before timer runs', async () => {
  const app = await authenticated(); await callable(app);
  const checking = app.element('check-session').onclick(); app.respond('GET', 'session', {csrf: 'controlled', expires_at: new Date(app.clock.now + 1).toISOString()}); await checking;
  app.clock.now += 2; assert.equal(send(app), undefined);
  assert.equal(app.element('login').hidden, false); assert.equal(app.element('test-key').value, '');
  assert.equal(app.pending.some(value => value.path === '/v1/chat/completions'), false);
});

test('hidden generation skips polling; visibility check is single flight and stale failure cannot clear a new page', async () => {
  const app = await authenticated(); await callable(app);
  const completed = send(app); const request = generation(app); app.document.hidden = true;
  await [...app.sessionTimers.values()][0]();
  assert.equal(app.pending.some(value => value.path === '/admin/api/session'), false);
  app.document.hidden = false; app.events.get('visibilitychange')(); app.events.get('pageshow')();
  const checks = app.pending.filter(value => value.path === '/admin/api/session'); assert.equal(checks.length, 1);
  const navigation = app.navigate('upstream'); app.respond('GET', 'upstream', {enabled: true, credential_configured: true}); await navigation;
  app.element('upstream-token').value = 'controlled new draft';
  app.respond('GET', 'session', {error: {code: 'unauthorized'}}, 401); await tick();
  assert.equal(app.element('console').hidden, false); assert.equal(app.element('upstream-token').value === 'controlled new draft', true);
  request.resolve(response('old', {choices: [{message: {content: 'late'}}]})); await completed;
  assert.equal(app.element('test-output').textContent, '');
});

for (const outcome of ['success', 'network failure', '401']) {
  test('known expiry survives a hanging one-shot session check and ignores late ' + outcome, async () => {
    const app = await authenticated(45000); await callable(app);
    const start = app.clock.now; app.element('test-stream').checked = true;
    const reads = []; let canceled = 0; const completed = send(app); const request = generation(app);
    request.resolve({ok: true, headers: {get: () => 'hanging-session'}, body: {getReader: () => ({read: () => new Promise(resolve => reads.push(resolve)), cancel: async () => { canceled++; }})}}); await tick();
    reads.shift()({value: new TextEncoder().encode('data: {"choices":[{"index":0,"delta":{"content":"controlled partial output"}}]}\n\n'), done: false}); await tick();
    assert.equal(app.element('test-output').textContent.length > 0, true); assert.equal(app.element('test-raw').textContent.length > 0, true);
    app.clock.now = start + 30000; const checking = app.fireNextSessionTimer();
    assert.equal(app.pending.filter(value => value.path === '/admin/api/session').length, 1);
    assert.equal(app.element('login').hidden, true); assert.equal(request.options.signal.aborted, false);
    assert.equal(app.sessionTimers.size > 0, true, 'known expiry stays armed while GET session is pending');
    app.clock.now = start + 46000; app.fireNextSessionTimer();
    assert.equal(app.element('login').hidden, false); assert.equal(request.options.signal.aborted, true);
    assert.equal(app.element('test-key').value, ''); assert.equal(app.element('test-message').value, '');
    assert.equal(app.element('test-output').textContent, ''); assert.equal(app.element('test-raw').textContent, ''); assert.equal(canceled, 1);
    assert.equal(app.location.hash, '#test'); assert.equal(app.sessionTimers.size, 0);
    if (outcome === 'network failure') {
      const index = app.pending.findIndex(value => value.path === '/admin/api/session'); app.pending.splice(index, 1)[0].reject(new Error('late offline'));
    } else app.respond('GET', 'session', outcome === 'success' ? {csrf: 'old', expires_at: new Date(start + 600000).toISOString()} : {error: {code: 'unauthorized'}}, outcome === '401' ? 401 : 200);
    await checking;
    reads.shift()({value: new TextEncoder().encode('data: {"choices":[{"delta":{"content":"late"}}]}\n\ndata: [DONE]\n\n'), done: false}); await completed;
    assert.equal(app.element('login').hidden, false); assert.equal(app.element('test-output').textContent, ''); assert.equal(app.sessionTimers.size, 0);
    assert.match(app.element('error').textContent, /管理会话已过期/);
    assert.equal(app.pending.some(value => value.path === '/v1/chat/completions'), false);
  });
}

for (const outcome of ['success', 'network failure', '401']) {
  test('hanging session result across navigation cannot extend expiry or affect a new login: ' + outcome, async () => {
    const app = await authenticated(45000); await callable(app); const start = app.clock.now;
    const oldCompleted = send(app); const oldRequest = generation(app);
    app.clock.now = start + 30000; const oldChecking = app.fireNextSessionTimer();
    const navigation = app.navigate('upstream'); app.respond('GET', 'upstream', {enabled: true, credential_configured: true}); await navigation;
    app.element('upstream-token').value = 'controlled draft';
    app.clock.now = start + 46000; app.fireNextSessionTimer();
    assert.equal(app.element('upstream-token').value, ''); assert.equal(app.element('login').hidden, false); assert.equal(app.location.hash, '#upstream');
    oldRequest.resolve(response('old', {choices: [{message: {content: 'old'}}]})); await oldCompleted;
    app.element('password').value = 'controlled'; const login = app.element('login-form').onsubmit({preventDefault() {}, submitter: {disabled: false}});
    app.respond('POST', 'login', {csrf: 'new', expires_at: new Date(app.clock.now + 300000).toISOString()}); await tick();
    app.respond('GET', 'upstream', {enabled: true, credential_configured: true}); await login;
    await callable(app); const newCompleted = send(app); const newRequest = generation(app);
    const deadlines = [...app.sessionTimers.keys()].map(id => app.timerDue.get(id));
    if (outcome === 'network failure') {
      const index = app.pending.findIndex(value => value.path === '/admin/api/session'); app.pending.splice(index, 1)[0].reject(new Error('late offline'));
    } else app.respond('GET', 'session', outcome === 'success' ? {csrf: 'old', expires_at: new Date(app.clock.now + 900000).toISOString()} : {error: {code: 'unauthorized'}}, outcome === '401' ? 401 : 200);
    await oldChecking;
    assert.equal(app.element('console').hidden, false); assert.equal(newRequest.options.signal.aborted, false);
    assert.equal(app.element('test-key').value === 'mr_isolated', true); assert.equal(app.element('test-message').value === 'isolated message', true);
    assert.deepEqual([...app.sessionTimers.keys()].map(id => app.timerDue.get(id)), deadlines);
    assert.doesNotMatch(app.element('error').textContent, /已过期|暂无法确认/);
    newRequest.resolve(response('new', {choices: [{message: {role: 'assistant', content: 'new output'}}]})); await newCompleted;
    assert.equal(app.element('test-output').textContent, 'new output');
    assert.equal(app.pending.some(value => value.path === '/v1/chat/completions'), false);
  });
}

test('uncertain session recovery may hang but the original known expiry still clears it', async () => {
  const app = await authenticated(45000); await callable(app); const start = app.clock.now;
  const completed = send(app); const request = generation(app);
  const checking = app.fireNextSessionTimer();
  const index = app.pending.findIndex(value => value.path === '/admin/api/session'); app.pending.splice(index, 1)[0].reject(new Error('offline')); await checking;
  assert.match(app.element('error').textContent, /暂无法确认/);
  request.resolve(response('late', {choices: [{message: {content: 'late'}}]})); await completed;
  const recovery = app.element('check-session').onclick();
  app.clock.now = start + 46000; app.fireNextSessionTimer();
  assert.equal(app.element('login').hidden, false);
  app.respond('GET', 'session', {csrf: 'old', expires_at: new Date(app.clock.now + 300000).toISOString()}); await recovery;
  assert.equal(app.element('login').hidden, false); assert.equal(app.sessionTimers.size, 0); assert.match(app.element('error').textContent, /管理会话已过期/);
});

test('ordinary periodic session success does not tell an active call to re-enter or resend', async () => {
  const app = await authenticated(45000); await callable(app);
  app.element('test-stream').checked = true; const reads = [];
  const completed = send(app); const request = generation(app);
  request.resolve({ok: true, headers: {get: () => 'periodic'}, body: {getReader: () => ({read: () => new Promise(resolve => reads.push(resolve)), cancel: async () => {}})}}); await tick();
  reads.shift()({value: new TextEncoder().encode('data: {"choices":[{"delta":{"content":"controlled partial"}}]}\n\n'), done: false}); await tick();
  app.element('notice').textContent = '当前调用继续接收';
  const checking = app.fireNextSessionTimer(); app.respond('GET', 'session', {csrf: 'test-only-csrf', expires_at: new Date(app.clock.now + 600000).toISOString()}); await checking;
  assert.equal(app.element('test-status').textContent, '接收中'); assert.equal(request.options.signal.aborted, false);
  assert.equal(app.element('test-output').textContent, 'controlled partial'); assert.equal(app.element('notice').textContent, '当前调用继续接收');
  assert.doesNotMatch(app.element('notice').textContent, /重新输入|手动发送/);
  assert.equal(app.element('test-message').value === 'isolated message', true);
  const generationCount = app.pending.filter(value => value.path === '/v1/chat/completions').length;
  app.clock.now += 16000; app.fireNextSessionTimer();
  assert.equal(request.options.signal.aborted, true, 'a check must not extend the originally known session deadline');
  assert.equal(app.pending.filter(value => value.path === '/v1/chat/completions').length, generationCount);
  reads.shift()({value: new TextEncoder().encode('data: [DONE]\n\n'), done: false}); await completed;
});

test('pagehide stops timers and pageshow still enforces a known deadline before a hanging check can resume', async () => {
  const app = await authenticated(45000); await callable(app); const start = app.clock.now;
  const completed = send(app); const request = generation(app); const checking = app.fireNextSessionTimer();
  app.events.get('pagehide')(); assert.equal(app.sessionTimers.size, 0); assert.equal(request.options.signal.aborted, true);
  app.clock.now = start + 46000; app.events.get('pageshow')();
  assert.equal(app.element('login').hidden, false); assert.equal(app.sessionTimers.size, 0);
  app.respond('GET', 'session', {csrf: 'old', expires_at: new Date(app.clock.now + 300000).toISOString()}); await checking;
  request.resolve(response('old', {choices: [{message: {content: 'old'}}]})); await completed;
  assert.equal(app.element('login').hidden, false); assert.equal(app.element('test-output').textContent, '');
});

test('visibility restore enforces known expiry even when uncertain-state timer delivery was suspended', async () => {
  const app = await authenticated(45000); await callable(app); const start = app.clock.now;
  const completed = send(app); const request = generation(app); const checking = app.fireNextSessionTimer();
  const index = app.pending.findIndex(value => value.path === '/admin/api/session'); app.pending.splice(index, 1)[0].reject(new Error('offline')); await checking;
  request.resolve(response('old', {choices: [{message: {content: 'old'}}]})); await completed;
  app.clock.now = start + 46000; app.document.hidden = false; app.events.get('visibilitychange')();
  assert.equal(app.element('login').hidden, false); assert.equal(app.sessionTimers.size, 0);
  assert.equal(app.pending.some(value => value.path === '/admin/api/session'), false);
});

function modelKeyboardFixture(commitOnBlur, missing = false, disabled = false, initialIndex = 0, target = 'alpha-390', disabledFirst = false, ignoreTyping = false) {
  const journey = readFileSync(new URL('./journey.mjs', import.meta.url), 'utf8');
  const start = journey.indexOf('async function keyboardModel(');
  const end = journey.indexOf('async function keyboardButton(', start);
  assert.ok(start >= 0 && end > start);
  const options = ['alpha-1280', 'alpha-390', 'beta-1280', 'beta-390'].map((value, index) => ({value, disabled: (disabled && value === target) || (disabledFirst && index === 0)}));
  const node = {options, selectedIndex: initialIndex, disabled: false};
  let pendingIndex = initialIndex;
  const keys = [];
  const locator = {evaluate: async (callback, argument) => callback(node, argument), inputValue: async () => options[node.selectedIndex].value};
  const page = {evaluate: async () => {}, getByLabel: label => { assert.equal(label, '调用别名'); return locator; }, keyboard: {press: async key => {
    keys.push(key);
    assert.equal(key, 'Tab', 'type-ahead selection must not open an OS popup or submit the form');
    if (!commitOnBlur || key === 'Tab') node.selectedIndex = pendingIndex;
  }, type: async text => {
    keys.push('type-ahead');
    const candidate = options.findIndex(option => option.value === text && !option.disabled);
    if (!ignoreTyping && candidate >= 0) pendingIndex = candidate;
    if (!commitOnBlur) node.selectedIndex = pendingIndex;
  }}};
  const stages = [], receipts = [];
  let ended = false;
  const context = vm.createContext({assert, Buffer, createHash, modelDiagnosticFile: null, keyboardFocus: async () => {}, beginModelDiagnostics: async () => {}, sampleModelDiagnostics: async (page, stage) => { stages.push(stage); return {stage, selectedIndex: node.selectedIndex, optionCount: options.length, trace: []}; }, endModelDiagnostics: async () => { ended = true; }, mkdir: async () => {}, appendFile: async (path, text) => receipts.push(JSON.parse(text)), process: {pid: 1}, viewport: 390, checkpoint: '', modelProbe: null});
  vm.runInContext(journey.slice(start, end) + '\nthis.selectModel = keyboardModel;', context);
  return {run: () => context.selectModel(page, missing ? 'absent' : target), node, keys, context, stages, receipts, ended: () => ended};
}

for (const commitOnBlur of [false, true]) test('closed model selection helper commits later viewport option: ' + commitOnBlur, async () => {
  const selection = modelKeyboardFixture(commitOnBlur);
  await selection.run();
  assert.equal(selection.node.selectedIndex, 1);
  assert.deepEqual(selection.keys, ['type-ahead', 'Tab']);
  assert.equal(selection.context.modelProbe, null);
});

test('closed model selection helper fails rather than accepting an unchanged first option', async () => {
  const selection = modelKeyboardFixture(false, false, false, 0, 'alpha-390', false, true);
  await assert.rejects(selection.run());
  assert.equal(selection.context.modelProbe.stage, 'after commit');
  assert.equal(selection.context.modelProbe.selectedIndex, 0);
  assert.equal(selection.context.modelProbe.matchesTarget, false);
  assert.equal(selection.ended(), true);
  assert.equal(selection.receipts[0].matchesTarget, false);
  const row = JSON.stringify(selection.receipts[0]) + '\n';
  assert.equal(selection.context.modelDiagnosticFile.path, 'test-results/model-selection-390.jsonl');
  assert.equal(selection.context.modelDiagnosticFile.sha256, createHash('sha256').update(row).digest('hex'));
  assert.equal(selection.context.modelDiagnosticFile.bytes, Buffer.byteLength(row));
  assert.ok(selection.receipts[0].boundaries.some(boundary => boundary.stage === 'after type-ahead' && boundary.selectedIndex === 0));
});

test('model selection diagnostic records type-ahead and Tab boundaries without assigning DOM values', async () => {
  const selection = modelKeyboardFixture(false);
  await selection.run();
  assert.equal(selection.node.selectedIndex, 1);
  assert.deepEqual(selection.stages, ['before keyboard reach', 'after keyboard reach', 'before type-ahead', 'after type-ahead', 'before Tab', 'after Tab', 'after commit', 'before cleanup']);
  assert.equal(selection.ended(), true);
  assert.equal(selection.receipts.length, 1);
  assert.equal(selection.receipts[0].selectedIndex, 1);
});

for (const [target, expectedIndex, disabledFirst] of [['alpha-1280', 0, false], ['beta-390', 3, false], ['alpha-390', 1, true]]) test('closed model selection helper handles starting at end and disabled options: ' + expectedIndex + '/' + disabledFirst, async () => {
  const selection = modelKeyboardFixture(true, false, false, 3, target, disabledFirst);
  await selection.run();
  assert.equal(selection.node.selectedIndex, expectedIndex);
  assert.equal(selection.keys.at(-1), 'Tab');
  assert.equal(selection.context.modelProbe, null);
});

for (const [missing, disabled] of [[true, false], [false, true]]) test('native model selection helper rejects unavailable target before activation: ' + missing, async () => {
  const selection = modelKeyboardFixture(true, missing, disabled);
  await assert.rejects(selection.run());
  assert.deepEqual(selection.keys, []);
});

function modelDiagnosticFixture() {
  const journey = readFileSync(new URL('./journey.mjs', import.meta.url), 'utf8');
  const start = journey.indexOf('const modelReadStates =');
  const end = journey.indexOf('const responses =', start);
  const listeners = new Map(), windowListeners = new Map(), network = new Map(), observers = [];
  const select = {options: [{value: 'private-option-sentinel'}, {value: 'private-other-sentinel'}], disabled: false, selectedIndex: 0, isConnected: true};
  const heading = {}, navigationButton = {};
  const nodes = {'test-model': select, 'page-loading': {hidden: true}, test: {hidden: false}};
  const document = {documentElement: {}, activeElement: select, hasFocus: () => true, getElementById: id => nodes[id], querySelector: selector => selector === '#test h1' ? heading : navigationButton, addEventListener: (name, callback) => listeners.set(name, callback), removeEventListener: name => listeners.delete(name)};
  const page = {evaluate: async (callback, argument) => callback(argument), on: (name, callback) => network.set(name, callback)};
  class Observer {
    constructor(callback) { this.callback = callback; observers.push(this); }
    observe() {}
    disconnect() { this.disconnected = true; }
  }
  const context = vm.createContext({assert, Buffer, URL, WeakMap, Map, document, window: {addEventListener: (name, callback) => windowListeners.set(name, callback), removeEventListener: name => windowListeners.delete(name)}, location: {hash: '#test'}, navigationVersion: 7, loadSequence: 9, active: 'test', MutationObserver: Observer});
  vm.runInContext(journey.slice(start, end) + '\nthis.begin = beginModelDiagnostics; this.sample = sampleModelDiagnostics; this.end = endModelDiagnostics; this.track = trackModelReads; this.projectFailure = modelFailureDiagnostic; this.serializeFailure = serializeFailure;', context);
  context.track(page);
  return {context, page, document, nodes, select, heading, listeners, windowListeners, network, observers, sample: stage => context.sample(page, stage)};
}

test('actual model diagnostic projection observes focus, option redraw and replacement without values', async () => {
  const diagnostic = modelDiagnosticFixture();
  await diagnostic.context.begin(diagnostic.page);
  let snapshot = await diagnostic.sample('after keyboard reach');
  assert.equal(snapshot.sameSelect && snapshot.sameOptions && snapshot.focused, true);
  diagnostic.document.activeElement = diagnostic.heading;
  diagnostic.listeners.get('focusin')({type: 'focusin'});
  snapshot = await diagnostic.sample('before ArrowDown');
  assert.equal(snapshot.focusName, 'test-heading'); assert.equal(snapshot.focused, false);
  diagnostic.select.options = diagnostic.select.options.map(option => ({...option}));
  diagnostic.observers[0].callback([{type: 'childList', target: diagnostic.select}]);
  snapshot = await diagnostic.sample('after ArrowDown');
  assert.equal(snapshot.sameSelect, true); assert.equal(snapshot.sameOptions, false); assert.equal(snapshot.optionRenders, 1);
  diagnostic.nodes['test-model'] = {...diagnostic.select};
  diagnostic.context.navigationVersion++; diagnostic.context.loadSequence++;
  diagnostic.windowListeners.get('hashchange')({type: 'hashchange'});
  diagnostic.observers[0].callback([{type: 'childList', target: diagnostic.document.documentElement}]);
  snapshot = await diagnostic.sample('after Tab');
  assert.equal(snapshot.sameSelect, false); assert.equal(snapshot.replacements, 1);
  assert.equal(snapshot.navigation, 8); assert.equal(snapshot.load, 10);
  assert.ok(snapshot.trace.some(boundary => boundary.stage === 'hashchange'));
  assert.equal(JSON.stringify(snapshot).includes('private-'), false);
  await diagnostic.context.end(diagnostic.page);
  assert.equal(diagnostic.listeners.size, 0); assert.equal(diagnostic.windowListeners.size, 0); assert.equal(diagnostic.observers[0].disconnected, true);
  assert.equal(diagnostic.context.window.__relayModelDiagnostics, undefined);
});

test('model diagnostics distinguish pending response bodies, failed reads and load-finally readiness', async () => {
  const diagnostic = modelDiagnosticFixture();
  await diagnostic.context.begin(diagnostic.page);
  const request = name => ({url: () => 'http://relay.invalid/admin/api/' + name, method: () => 'GET'});
  const models = request('models'), overview = request('overview');
  diagnostic.network.get('request')(models); diagnostic.network.get('request')(overview);
  diagnostic.nodes['page-loading'].hidden = false;
  let snapshot = await diagnostic.sample('before ArrowUp');
  assert.equal(snapshot.reads.models.pending, 1); assert.equal(snapshot.readsSettled, false);
  diagnostic.network.get('requestfinished')(models); diagnostic.network.get('requestfailed')(overview);
  snapshot = await diagnostic.sample('after ArrowUp');
  assert.equal(snapshot.reads.models.finished, 1); assert.equal(snapshot.reads.overview.failed, 1);
  assert.equal(snapshot.reads.overview.pending, 0); assert.equal(snapshot.readsSettled, false);
  diagnostic.nodes['page-loading'].hidden = true;
  snapshot = await diagnostic.sample('before Tab');
  assert.equal(snapshot.readsSettled, true);
  assert.equal(snapshot.reads.overview.finished, 0, 'settled does not mean successful');
  await diagnostic.context.end(diagnostic.page);
});

test('model diagnostic event trace is bounded and only records approved key and focus labels', async () => {
  const diagnostic = modelDiagnosticFixture();
  await diagnostic.context.begin(diagnostic.page);
  diagnostic.listeners.get('keydown')({key: 'private-key-sentinel', type: 'keydown', target: diagnostic.select});
  for (let index = 0; index < 50; index++) diagnostic.listeners.get('keydown')({key: 'ArrowDown', type: 'keydown', target: diagnostic.select});
  const snapshot = await diagnostic.sample('after ArrowDown');
  assert.equal(snapshot.trace.length, 40); assert.equal(snapshot.truncated, true);
  assert.equal(snapshot.samples, 51);
  assert.equal(snapshot.trace[0].stage, 'keydown-ArrowDown-select');
  assert.equal(JSON.stringify(snapshot).includes('private-'), false);
  await diagnostic.context.end(diagnostic.page);
});

test('model diagnostic printable keys are redacted and type-ahead boundaries survive compaction', async () => {
  const diagnostic = modelDiagnosticFixture();
  await diagnostic.context.begin(diagnostic.page);
  diagnostic.listeners.get('keydown')({key: 'x', type: 'keydown', target: diagnostic.select});
  diagnostic.listeners.get('keyup')({key: 'x', type: 'keyup', target: diagnostic.select});
  diagnostic.listeners.get('keydown')({key: 'y', type: 'keydown', target: diagnostic.heading});
  const snapshot = await diagnostic.sample('after type-ahead');
  assert.deepEqual(Array.from(snapshot.trace, value => value.stage), ['keydown-printable-select', 'keyup-printable-select', 'after type-ahead']);
  const boundaries = ['before type-ahead', 'after type-ahead'].map(stage => ({...snapshot, stage}));
  const modelProbe = diagnostic.context.projectFailure({...snapshot, boundaries, boundaryCount: 2});
  const line = diagnostic.context.serializeFailure({viewport: 390, checkpoint: 'native selection', location: 'journey.mjs:1:1', error: 'AssertionError', modelProbe, responses: [], clipboardProbe: {stage: '界'.repeat(6000)}});
  const summary = JSON.parse(line.slice(line.indexOf('{')));
  assert.equal(summary.compacted, true);
  assert.deepEqual(summary.modelProbe.boundaries.map(value => value.stage), ['before type-ahead', 'after type-ahead']);
  assert.equal(Buffer.byteLength(line + '\n') <= 8192, true);
  await diagnostic.context.end(diagnostic.page);
});

test('failure model projection omits arbitrary fields and keeps the bounded boundary evidence', async () => {
  const diagnostic = modelDiagnosticFixture();
  await diagnostic.context.begin(diagnostic.page);
  const snapshot = await diagnostic.sample('after commit');
  const unsafe = {...snapshot, value: 'private-value-sentinel', arbitraryDOM: 'private-dom-sentinel'};
  const projected = diagnostic.context.projectFailure({...unsafe, boundaries: [unsafe], trace: Array(20).fill(unsafe)});
  assert.equal(projected.recentEvents.length, 8);
  assert.equal(projected.boundaries[0].selectedIndex, 0);
  assert.equal(JSON.stringify(projected).includes('private-'), false);
  assert.equal(diagnostic.context.projectFailure(null), null);
  await diagnostic.context.end(diagnostic.page);
});

test('entire UTF-8 failure line fits host tail while retaining first and final key boundaries', async () => {
  const diagnostic = modelDiagnosticFixture();
  await diagnostic.context.begin(diagnostic.page);
  const snapshot = await diagnostic.sample('after commit');
  const stages = ['before keyboard reach', 'after keyboard reach', 'before ArrowUp', 'after ArrowUp', 'before ArrowDown', 'after ArrowDown', 'before Tab', 'after Tab', 'after commit'];
  const boundaries = stages.flatMap(stage => [{...snapshot, stage, load: 1}, ...(['before ArrowUp', 'after ArrowUp'].includes(stage) ? Array(3).fill({...snapshot, stage, load: 1}) : []), {...snapshot, stage, load: 2}]);
  const model = diagnostic.context.projectFailure({...snapshot, samples: 140, boundaryCount: 72, boundaries, trace: Array(40).fill(snapshot), value: 'private-sentinel'});
  const file = {path: 'test-results/model-selection-390.jsonl', sha256: 'a'.repeat(64), bytes: 18000, execution: 123};
  for (const inflated of [false, true]) {
    const line = diagnostic.context.serializeFailure({viewport: 390, checkpoint: '选择诊断'.repeat(200), location: 'journey.mjs:1:1', error: 'AssertionError', modelProbe: model, modelDiagnosticFile: file, modelDiagnosticExecution: 123, passwordProbe: null, keyboardProbe: null, clipboardProbe: inflated ? {stage: '界'.repeat(6000)} : null, responses: Array(12).fill({method: 'GET', path: '/admin/api/overview', status: 200})});
    assert.equal(Buffer.byteLength(line + '\n', 'utf8') <= 8192, true);
    const summary = JSON.parse(line.slice(line.indexOf('{')));
    assert.equal(summary.outputLimitBytes, 8192);
    assert.equal(summary.modelDiagnosticFile.sha256, file.sha256);
    for (const stage of stages) {
      const matching = summary.modelProbe.boundaries.filter(boundary => boundary.stage === stage);
      assert.equal(matching.length, 2);
      assert.equal(matching[0].load, 1);
      assert.equal(matching[1].load, 2);
    }
    assert.equal(summary.modelProbe.omittedBoundaries, 54);
    assert.equal(summary.modelProbe.omittedEvents >= 132, true);
    assert.equal(line.includes('private-sentinel'), false);
    if (inflated) assert.equal(summary.compacted, true);
  }
  await diagnostic.context.end(diagnostic.page);
});
