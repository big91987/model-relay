import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {mkdirSync, readFileSync, readdirSync, writeFileSync} from 'node:fs';
import {join} from 'node:path';
import {spawnSync} from 'node:child_process';

const runRoot = 'docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167';
const destination = join(runRoot, 'acceptance/evidence/qa-10');
const receiptRoot = 'test-results/verification/20261009T131347Z-1173031951';
const materialRoot = '.workflow-input/637705186406b80a2ee4afaf0d6b2167/a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49/files';
const digest = bytes => createHash('sha256').update(bytes).digest('hex');
const records = [];
const unsafeText = /\/Users\/|\/private\/var\/|\/var\/folders\/|Bearer\s+\S+|\bsk-[A-Za-z0-9_-]{12,}/;
function save(path, bytes) {
  if (!path.endsWith('.png')) assert.equal(unsafeText.test(bytes.toString()), false, `unsafe text: ${path}`);
  mkdirSync(join(destination, path, '..'), {recursive: true});
  writeFileSync(join(destination, path), bytes);
  records.push({path: join(destination, path), bytes: bytes.length, sha256: digest(bytes)});
}
function saveJSON(path, value) {
  save(path, Buffer.from(JSON.stringify(value, null, 2) + '\n'));
}
const manifest = JSON.parse(readFileSync(join(materialRoot, 'manifest.json')));
assert.equal(manifest.version, 'product-test33-handoff-r1');
assert.equal(manifest.files.length, 61);
for (const entry of manifest.files) {
  const bytes = readFileSync(join(materialRoot, entry.path));
  assert.equal(bytes.length, entry.bytes, entry.path);
  assert.equal(digest(bytes), entry.sha256, entry.path);
}
const archive = readFileSync(join(materialRoot, '../archive.zip'));
assert.equal(digest(archive), 'a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49');
saveJSON('material-integrity.json', {version: manifest.version, material_root: materialRoot, manifest_path: join(materialRoot, 'manifest.json'), archive_sha256: digest(archive), manifest_sha256: digest(readFileSync(join(materialRoot, 'manifest.json'))), checked: 61, mismatches: [], files: manifest.files});
const state = JSON.parse(readFileSync(join(receiptRoot, 'state.json')));
const log = readFileSync(join(receiptRoot, 'verify.log'));
assert.equal(log.length, state.log_bytes);
assert.equal(digest(log), state.log_sha256);
assert.equal(state.exit_code, 0);
assert.equal(state.gate_exit, 0);
assert.equal(state.source_unchanged, true);
assert.equal(state.candidate_unchanged, true);
assert.equal(state.owned_group_absent, true);
assert.deepEqual(state.source_before, state.source_after);
const source = spawnSync('go', ['run', './tests/evidence'], {encoding: 'utf8', env: {...process.env, GOCACHE: join(process.cwd(), '.cache/go-build'), GOTOOLCHAIN: 'local'}});
assert.equal(source.status, 0, 'QA source evidence command failed');
const snapshot = JSON.parse(source.stdout);
assert.deepEqual(snapshot, state.source_after);
saveJSON('source-qa.json', snapshot);
saveJSON('host-tests9/state.json', state);
save('host-tests9/verify.log', log);
const prefix = Buffer.from(`go run ./tests/verify\nverification_receipt=${receiptRoot}\n`);
const reconstructed = Buffer.concat([prefix, log]);
assert.equal(reconstructed.length, 499596);
save('host-tests9/connector-output-reconstructed.log', reconstructed);
const lines = log.toString().split('\n');
save('host-tests9/selection.log', Buffer.from(lines.filter(line => /METERING_|^ok\s|^BROWSER_PHASE|^PASS |^CLEAN_END|routing_cooldown=|"suite":/.test(line)).join('\n') + '\n'));
function copyTree(sourceDirectory, targetDirectory) {
  for (const entry of readdirSync(sourceDirectory, {withFileTypes: true})) {
    const sourcePath = join(sourceDirectory, entry.name);
    const targetPath = join(targetDirectory, entry.name);
    if (entry.isDirectory()) copyTree(sourcePath, targetPath);
    else {
      assert.match(entry.name, /\.(json|jsonl|png)$/);
      save(targetPath, readFileSync(sourcePath));
    }
  }
}
copyTree(join(receiptRoot, 'browser'), 'host-tests9/browser');
const localLogTransforms = [];
for (const name of ['go.log', 'node.log', 'node-focused.log']) {
  const original = readFileSync(join('.cache/qa-10', name));
  const sanitized = Buffer.from(original.toString().replace(/\/private\/var\/folders\/[^'\s]+/g, '<temporary-object>'));
  save(`independent-${name}`, sanitized);
  localLogTransforms.push({name, raw_bytes: original.length, raw_sha256: digest(original), archived_sha256: digest(sanitized), redacted: !original.equals(sanitized), rule: 'only temporary Darwin linker object path; warning and all test results retained'});
}
saveJSON('local-log-transforms.json', localLogTransforms);
const diff = spawnSync('git', ['diff', '--check'], {encoding: 'utf8'});
assert.equal(diff.status, 0);
saveJSON('independent-checks.json', {source_command: 'GOCACHE=<workspace-cache> GOTOOLCHAIN=local go run ./tests/evidence', source_exit: source.status, diff_check_exit: diff.status, go_command: "GOCACHE=<workspace-cache> GOTOOLCHAIN=local CGO_ENABLED=1 go test -race -count=1 -v ./internal/relay -parallel=4 -run '^TestMetering.*InProcess$'", go_exit_observed: 0, go_pass_footer: /PASS\nok\s+model-relay\/internal\/relay/.test(readFileSync('.cache/qa-10/go.log', 'utf8')), node_focused_command: 'node --test tests/browser/cooldown-observation.test.mjs tests/browser/routing-progress.test.mjs', node_focused_exit_observed: 0, node_focused_tests: 3, initial_node_smoke: 'node --test additionally included selection-trace.mjs; 4 reported, not four assertion suites', sandbox_initial_source_attempt: {result: 'Blocked', cause: 'default user Go cache outside writable roots', corrected: 'workspace GOCACHE source command exit0', not_a_product_failure: true}, full_gate_origin: 'host Connector tests execution9; QA did not repeat listening make verify', supplier_generation_by_qa: 0});
saveJSON('receipt-review.json', {connector_execution: 9, connector_exit: 0, connector_reported_bytes: 499596, connector_truncated: false, receipt_path: receiptRoot, runner_log_bytes: log.length, runner_log_sha256: digest(log), prefix_bytes: prefix.length, reconstructed_connector_sha256: digest(reconstructed), reconstruction_limit: 'reconstructed from runner log verified against state plus observed Connector prefix; retained Connector beginning/end sampled, middle not bytewise compared', elapsed_seconds: (Date.parse(state.finished_utc) - Date.parse(state.started_utc)) / 1000, source: snapshot, top_level_gate: 'Pass', expected_runner_failures: lines.filter(line => line.startsWith('FAILED:')).length, browser_phases: lines.filter(line => line.startsWith('BROWSER_PHASE')), skipped: lines.filter(line => /--- SKIP:/.test(line)), text_scan: 'no personal absolute paths, auth-header payloads or sk-style keys in archived text; PNG helpers assert cleared secrets/storage; QA visually sampled desktop/narrow finals', supplier_result: 'Blocked; not supplier validation'});
saveJSON('archive-index.json', {run_id: '637705186406b80a2ee4afaf0d6b2167', node: 'qa', execution: 10, records});
const deliverables = ['acceptance/release-assessment-qa-10.md', 'acceptance/acceptance-matrix-qa-10.md', 'acceptance/test-plan-review-qa-10.md', 'acceptance/core-journeys-qa-10.md', 'acceptance/findings-and-next-actions-qa-10.md', 'acceptance/evidence/index.md', 'acceptance/evidence/qa-10/collect.mjs', 'acceptance/evidence/qa-10/collection-attempts.md', 'acceptance/evidence/qa-10/archive-index.json', 'verification/evidence-index.json'];
const deliveryRecords = deliverables.map(path => {
  const repositoryPath = join(runRoot, path);
  const bytes = readFileSync(repositoryPath);
  return {path: repositoryPath, bytes: bytes.length, sha256: digest(bytes)};
});
saveJSON('handoff-artifacts.json', {run_id: '637705186406b80a2ee4afaf0d6b2167', execution: 10, source: snapshot, reports_and_indexes: deliveryRecords, archived_evidence: records.slice(0, -1)});
console.log(JSON.stringify({destination, files: records.length, source: snapshot, material_checked: 61, host_log_integrity: true}));
