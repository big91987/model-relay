import assert from 'node:assert/strict';
import test from 'node:test';
import {verifyMigratedKeys} from './migration-contract.mjs';

const items = [{name: 'migration-enabled', enabled: true}, {name: 'migration-disabled', enabled: false}];

test('migration driver consumes schema3 paginated Key contract', () => {
  verifyMigratedKeys({items, total: 2, page: 1, page_size: 25});
});

for (const [name, body] of Object.entries({
  obsolete_data: {data: items, total: 2},
  missing_key: {items: items.slice(0, 1), total: 2},
  additional_key: {items: [...items, {name: 'unexpected'}], total: 3},
  wrong_total: {items, total: 3},
  disabled_changed: {items: [items[0], {...items[1], enabled: true}], total: 2},
  enabled_changed: {items: [{...items[0], enabled: false}, items[1]], total: 2},
  missing_identity: {items: [items[0], {name: 'unexpected', enabled: false}], total: 2}
})) {
  test(`migration driver rejects ${name} without exposing collection values`, () => {
    assert.throws(() => verifyMigratedKeys(body), error => error.name === 'AssertionError' && !error.message.includes('unexpected'));
  });
}
