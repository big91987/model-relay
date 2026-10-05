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
  const document = {
    getElementById: element,
    querySelectorAll: selector => selector === '.panel' ? panels : buttons,
  };
  vm.runInNewContext(source, {document, AbortController, fetch: (path, options) => new Promise(resolve => pending.push({path, options, resolve}))});
  function respond(method, path, body, status = 200) {
    const index = pending.findIndex(request => request.path === '/admin/api/' + path && request.options.method === method);
    assert.notEqual(index, -1, `expected ${method} ${path}`);
    const [request] = pending.splice(index, 1);
    request.resolve({ok: status < 400, status, json: async () => body});
    return request;
  }
  return {element, panels, pending, respond, navigate: name => buttons.find(button => button.dataset.view === name).onclick()};
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
