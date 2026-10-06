import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';

const source = readFileSync('internal/relay/web/app.js', 'utf8');
const harness = readFileSync('tests/browser/ui.test.mjs', 'utf8').split("test('password toggle")[0]
  .replace(/^import .*;\n/gm, '')
  .replace(/^const source = .*;\n/m, '');
const fixture = new Function('source', 'vm', 'assert', 'setImmediate', harness + '\nreturn fixture;')(source, vm, assert, setImmediate);
const tick = () => new Promise(resolve => setImmediate(resolve));
const snapshot = {upstream: {configured: true, enabled: true, credential_configured: true}, models: {total: 1, enabled: 1}, keys: {total: 1, enabled: 1}, call_evidence: null};

async function authenticated(evidence = null) {
  const app = fixture();
  app.respond('GET', 'session', {csrf: 'qa-fixture'});
  await tick();
  app.respond('GET', 'overview', {...snapshot, call_evidence: evidence});
  await tick();
  return app;
}

async function upstream(app) {
  const navigation = app.navigate('upstream');
  app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true, credential_configured: true});
  await navigation;
}

test('QA-01 rejected upstream save clears credential and preserves public draft', async () => {
  const app = await authenticated();
  await upstream(app);
  app.element('upstream-url').value = 'http://127.0.0.1:1/v1';
  app.element('upstream-token').value = 'qa-only-placeholder';
  const completed = app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save')});
  app.respond('PUT', 'upstream', {error: {code: 'invalid_destination'}}, 400);
  await completed;
  assert.match(app.element('error').textContent, /上游地址不符合/);
  assert.equal(app.element('upstream-url').value, 'http://127.0.0.1:1/v1');
  assert.equal(app.element('upstream-token').value === '', true, 'failed save must clear the secret per interaction section 4');
});

test('QA-02 successful save followed by failed refresh distinguishes committed write', async () => {
  const app = await authenticated();
  await upstream(app);
  app.element('upstream-token').value = 'qa-only-placeholder';
  const completed = app.element('upstream-form').onsubmit({preventDefault() {}, submitter: app.element('save')});
  app.respond('PUT', 'upstream', {});
  await tick();
  app.respond('GET', 'upstream', {error: {code: 'storage_unavailable'}}, 503);
  await completed;
  assert.equal(app.element('upstream-token').value === '', true);
  const feedback = app.element('notice').textContent + app.element('error').textContent;
  assert.equal(/保存.*(已受理|已保存)/.test(feedback), true, 'the committed write must be disclosed separately from refresh failure');
});

test('QA-03 overview evidence can open its exact record after session restoration', async () => {
  const evidenceID = 'QAOnlyRecordIdentity01';
  assert.equal(evidenceID.length, 22);
  const app = await authenticated({request_id: evidenceID, alias: 'alpha', scope: '本次服务启动、当前配置的一次已记录成功'});
  assert.equal(app.element('summary').textContent.includes(evidenceID), true);
  const navigation = app.navigate('records');
  app.respond('GET', 'requests', {data: [], next_cursor: null});
  await tick();
  if (app.pending.some(request => request.path === '/admin/api/requests/' + evidenceID)) {
    app.respond('GET', 'requests/' + evidenceID, {id: evidenceID, time: '2026-10-06T02:00:00Z', alias: 'alpha', model: 'provider', key_id: 'qa-public-key-id', state: 'success', duration_ms: 1, prompt_tokens: 0, completion_tokens: 2, total_tokens: 2});
  }
  await navigation;
  assert.equal(app.element('error').textContent, '');
  assert.equal(app.element('records').hidden, false);
  assert.equal(app.element('find-request-id').value === evidenceID, true, 'the overview evidence action must carry the displayed identity, not an absent previous local call');
  assert.equal(app.element('record-status').textContent.includes(evidenceID), true, 'the exact record must be visibly identified');
});

test('QA-04 application rejection clears sensitive call input but retains public model choice', async () => {
  const app = await authenticated();
  const navigation = app.navigate('test', false);
  app.respond('GET', 'models', [{alias: 'alpha', model: 'provider', enabled: true}]);
  app.respond('GET', 'overview', snapshot);
  await navigation;
  app.element('test-model').value = 'alpha';
  app.element('test-key').value = 'qa-only-placeholder';
  app.element('test-message').value = 'qa-only-placeholder';
  app.element('test-form').onsubmit({preventDefault() {}});
  const index = app.pending.findIndex(request => request.path === '/v1/chat/completions');
  assert.notEqual(index, -1);
  const [request] = app.pending.splice(index, 1);
  request.resolve({ok: false, status: 401, headers: {get: () => null}, json: async () => ({error: {code: 'unauthorized'}})});
  await tick();
  assert.equal(app.element('test-status').textContent.includes('失败'), true);
  assert.equal(app.element('console').hidden, false);
  assert.equal(app.element('test-model').value, 'alpha');
  assert.equal(app.element('test-key').value === '' && app.element('test-message').value === '', true, 'failure must clear key and input message per interaction section 7');
});
