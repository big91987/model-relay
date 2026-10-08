import assert from 'node:assert/strict';

export function verifyMigratedKeys(body) {
  assert.ok(Array.isArray(body?.items), 'schema3 Key collection must contain items');
  assert.equal(body.items.length, 2, 'migrated Key count');
  assert.equal(body.total, 2, 'migrated Key total');
  assert.equal(body.items.find(key => key.name === 'migration-enabled')?.enabled, true, 'enabled Key preserved');
  assert.equal(body.items.find(key => key.name === 'migration-disabled')?.enabled, false, 'disabled Key preserved');
}
