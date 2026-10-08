import assert from 'node:assert/strict';
import test from 'node:test';
import { routingProgress } from './routing-progress.mjs';

test('routing progress preserves values, errors and whitelist-only diagnostics', async () => {
  const output = [];
  const progress = routingProgress(390, 7, line => output.push(JSON.parse(line)), 5);
  assert.equal(await progress.step('panel', async () => 'secret-value'), 'secret-value');
  const failure = new Error('secret-error');
  await assert.rejects(progress.step('send', async () => {
    await new Promise(resolve => setTimeout(resolve, 20));
    throw failure;
  }), error => error === failure);
  progress.finish(false);
  const count = output.length;
  await new Promise(resolve => setTimeout(resolve, 20));
  assert.equal(output.length, count);
  assert.deepEqual(output.map(entry => entry.routing_call), ['start', 'pending', 'end']);
  assert.equal(output[1].operation, 'send');
  assert.equal(output[2].ok, false);
  assert.deepEqual(Object.keys(output[2].durations), ['panel', 'send']);
  for (const entry of output) assert.equal(new Date(entry.utc).toISOString(), entry.utc);
  assert.equal(JSON.stringify(output).includes('secret'), false);
  await assert.rejects(progress.step('secret-label', async () => {}), /Unknown routing diagnostic operation/);
});
