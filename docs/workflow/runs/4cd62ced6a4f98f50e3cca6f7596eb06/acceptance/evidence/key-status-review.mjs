import fs from 'node:fs';
import vm from 'node:vm';

const source = fs.readFileSync('internal/relay/web/workspace.js', 'utf8');
const matched = source.match(/column\('状态','revoked',(value=>value\?'已撤销':'[^']+')\)/);
if (!matched) throw new Error('Reviewed status renderer changed; review the new implementation');
const formatter = vm.runInNewContext(`(${matched[1]})`);
const scenarios = [
  {name: 'enabled', revoked: false, enabled: true, expires_at: null},
  {name: 'disabled', revoked: false, enabled: false, expires_at: null},
  {name: 'expired', revoked: false, enabled: true, expires_at: '2000-01-01T00:00:00Z'},
  {name: 'revoked', revoked: true, enabled: false, expires_at: null}
];
console.log(JSON.stringify({
  evidence_type: 'Static Review; extracted actual formatter, not UI-E2E',
  source: 'internal/relay/web/workspace.js',
  field_passed_by_column: 'revoked',
  outcomes: scenarios.map(scenario => ({scenario: scenario.name, rendered: formatter(scenario.revoked)})),
  distinguishes_enabled_disabled_expired: new Set(scenarios.slice(0, 3).map(scenario => formatter(scenario.revoked))).size === 3
}, null, 2));
