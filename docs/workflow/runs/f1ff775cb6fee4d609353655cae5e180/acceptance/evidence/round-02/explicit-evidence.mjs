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
const evidenceID = 'QAOnlyRecordIdentity01';
const evidence = {request_id: evidenceID, alias: 'qa-public-alias', time: '2026-10-06T02:00:00Z', upstream_revision: 7, scope: '本次服务启动、当前配置的一次已记录成功'};
const snapshot = {upstream: {configured: true, enabled: true, credential_configured: true}, models: {total: 2, enabled: 1}, keys: {total: 3, enabled: 2}, call_evidence: evidence};
const record = {id: evidenceID, time: evidence.time, revision: 7, alias: evidence.alias, model: 'qa-public-model', key_id: 'qa-public-key-id', state: 'success', duration_ms: 0, prompt_tokens: 0, completion_tokens: null, total_tokens: null};

async function authenticated(value = snapshot) {
  const app = fixture();
  app.respond('GET', 'session', {csrf: 'qa-fixture'});
  await tick();
  app.respond('GET', 'overview', value);
  await tick();
  return app;
}

test('QA-D03 explicit overview action carries displayed identity and preserves metadata', async () => {
  const app = await authenticated();
  const summary = app.element('summary').textContent;
  for (const value of [evidenceID, evidence.alias, evidence.time, '配置版本 7', evidence.scope]) assert.equal(summary.includes(value), true);
  assert.equal(app.element('overview-record').disabled, false);
  const completed = app.element('overview-record').onclick();
  app.respond('GET', 'requests', {data: [], next_cursor: 'qa-public-cursor'});
  await tick();
  const queries = app.pending.filter(request => request.path.startsWith('/admin/api/requests/'));
  assert.equal(queries.length, 1);
  assert.equal(queries[0].path, '/admin/api/requests/' + evidenceID);
  assert.equal(queries[0].options.method, 'GET');
  app.respond('GET', 'requests/' + evidenceID, record);
  await completed;
  assert.equal(app.element('find-request-id').value, evidenceID);
  assert.equal(app.element('record-status').textContent.includes(evidenceID), true);
  for (const value of [evidenceID, evidence.alias, record.key_id, '0 / 未知 / 未知']) assert.equal(app.element('record-detail').textContent.includes(value), true);
  assert.equal(app.element('records').hidden, false);
});

test('QA-D03 generic records navigation stays unfiltered despite overview evidence', async () => {
  const app = await authenticated();
  const completed = app.navigate('records');
  app.respond('GET', 'requests', {data: [], next_cursor: null});
  await completed;
  assert.equal(app.element('find-request-id').value, '');
  assert.equal(app.pending.some(request => request.path.startsWith('/admin/api/requests/')), false);
});

test('QA-D03 refresh failure disables stale success action before read completes', async () => {
  const app = await authenticated();
  const completed = app.element('refresh-overview').onclick();
  assert.equal(app.element('overview-record').disabled, true);
  assert.equal(app.element('summary').textContent.includes(evidenceID), false);
  app.respond('GET', 'overview', {error: {code: 'storage_unavailable'}}, 503);
  await completed;
  assert.equal(app.element('overview-record').disabled, true);
  assert.equal(app.element('summary').textContent.includes(evidenceID), false);
  assert.equal(app.element('error').textContent.includes('存储不可用'), true);
});

test('QA-D03 late precise record cannot repaint another page', async () => {
  const app = await authenticated();
  const completed = app.element('overview-record').onclick();
  app.respond('GET', 'requests', {data: [], next_cursor: null});
  await tick();
  const navigation = app.navigate('models');
  app.respond('GET', 'models', []);
  await tick();
  app.respond('GET', 'upstream', {url: 'https://controlled.example/v1', enabled: true, credential_configured: true});
  await navigation;
  app.respond('GET', 'requests/' + evidenceID, record);
  await completed;
  assert.equal(app.element('models').hidden, false);
  assert.equal(app.element('records').hidden, true);
  assert.equal(app.element('record-detail').textContent, '');
});

test('QA-D03 absent metadata is unknown and absent evidence cannot open a stale record', async () => {
  const app = await authenticated({...snapshot, call_evidence: {...evidence, time: null, upstream_revision: null}});
  assert.equal(app.element('summary').textContent.includes('调用时间 未知 · 配置版本 未知'), true);
  const completed = app.element('refresh-overview').onclick();
  app.respond('GET', 'overview', {...snapshot, call_evidence: null});
  await completed;
  assert.equal(app.element('overview-record').disabled, true);
  assert.equal(app.element('summary').textContent.includes(evidenceID), false);
  assert.equal(app.element('overview-record').onclick(), undefined);
  assert.equal(app.pending.some(request => request.path.startsWith('/admin/api/requests')), false);
});
