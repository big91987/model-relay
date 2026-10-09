import assert from 'node:assert/strict';
import test from 'node:test';
import { cooldownObservation } from './cooldown-observation.mjs';

test('cooldown observation projects only target health counts and remaining time', () => {
  const now = Date.parse('2026-10-09T12:33:39Z');
  const value = {data: [
    {alias: 'private-target', candidates: [
      {generation_health: '暂时冷却', cooldown_until: '2026-10-09T12:34:09Z', credential: 'private-token'},
      {generation_health: '最近生成失败（仍参与）', remote_model: 'private-model'}
    ]},
    {alias: 'private-other', candidates: [{generation_health: '暂时冷却', cooldown_until: '2026-10-09T12:33:29Z'}]}
  ], next_cursor: 'private-cursor', prompt: 'private-prompt'};
  const result = cooldownObservation(value, 'private-target', now);
  assert.deepEqual(result, {model_count: 2, matching_models: 1, has_next_page: true, candidate_count: 2, cooling_count: 1, expired_count: 0, failed_count: 1, minimum_remaining_ms: 30000});
  assert.equal(JSON.stringify(result).includes('private'), false);
});

test('cooldown observation distinguishes missing, expired and invalid evidence', () => {
  const now = Date.parse('2026-10-09T12:33:39Z');
  assert.deepEqual(cooldownObservation(null, 'target', now), {model_count: 0, matching_models: 0, has_next_page: false, candidate_count: 0, cooling_count: 0, expired_count: 0, failed_count: 0, minimum_remaining_ms: null});
  const result = cooldownObservation({data: [{alias: 'target', candidates: [
    {generation_health: '待重新尝试', cooldown_until: '2026-10-09T12:33:29Z'},
    {generation_health: 'private-error', cooldown_until: 'private-date'}
  ]}]}, 'target', now);
  assert.equal(result.expired_count, 1);
  assert.equal(result.cooling_count, 0);
  assert.equal(result.minimum_remaining_ms, -10000);
  assert.equal(JSON.stringify(result).includes('private'), false);
  assert.equal(cooldownObservation({data: [{alias: 'other'}], next_cursor: null}, 'target', now).matching_models, 0);
});
