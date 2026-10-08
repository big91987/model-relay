import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFile, writeFile, mkdir, copyFile} from 'node:fs/promises';
import {dirname} from 'node:path';

const root = 'docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06';
const destination = `${root}/acceptance/evidence/attempt-47`;
const material = '.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7';
const digest = data => createHash('sha256').update(data).digest('hex');
const json = async path => JSON.parse(await readFile(path, 'utf8'));
const save = async (name, data) => writeFile(`${destination}/${name}`, JSON.stringify(data, null, 2) + '\n');
const manifest = await json(`${material}/files/manifest.json`);
assert.equal(manifest.version, 'design-v0.2.0');
const archive = await readFile(`${material}/archive.zip`);
assert.equal(digest(archive), '5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7');
const materialFiles = [];
for (const entry of manifest.files) {
  const path = `${material}/files/${entry.path}`;
  const data = await readFile(path);
  assert.equal(digest(data), entry.sha256, entry.path);
  assert.equal(data.length, entry.bytes, entry.path);
  materialFiles.push({path, bytes: data.length, sha256: digest(data)});
}
assert.equal(materialFiles.length, 19);
await save('material-integrity.json', {version: manifest.version, status: manifest.status, archive: `${material}/archive.zip`, archive_sha256: digest(archive), manifest: `${material}/files/manifest.json`, files: materialFiles});
const source = await json(`${destination}/source-before.json`);
assert.equal(source.head, '98511771871cf0951ecef716bbb55deb13e18da5');
assert.equal(source.files, 129);
assert.equal(source.working_tree_sha256, '7e0b20ddbdea4585959365d8ad3b6121abaa91cd3f450fa2c6c21f7afc05a9ab');
const artifacts = [];
async function retain(path, expected) {
  const data = await readFile(path);
  const sha256 = digest(data);
  if (expected) assert.equal(sha256, expected, path);
  const retained = `${destination}/host-seq46/${path.replace(/^test-results\//, '')}`;
  await mkdir(dirname(retained), {recursive: true});
  await copyFile(path, retained);
  artifacts.push({source: path, retained, bytes: data.length, sha256});
}
const widths = [1280, 1440, 390];
const executions = {1280: 34454, 1440: 34476, 390: 34501};
const reviews = [];
for (const width of widths) {
  const path = `test-results/visual-${width}/review.json`;
  const review = await json(path);
  assert.equal(review.result, 'passed');
  assert.equal(review.phase, 'completed');
  assert.equal(review.provider, 'controlled_http_not_supplier');
  assert.equal(review.pages.length, 10);
  await retain(path);
  for (const page of review.pages) {
    assert.equal(page.facts.ready, true);
    assert.equal(page.facts.rootFits, true);
    assert.equal(page.facts.secretsClear, true);
    for (const screenshot of [...(page.screenshots ?? []), page.detail?.screenshot, page.editor?.screenshot].filter(Boolean)) await retain(screenshot.path, screenshot.sha256);
  }
  const continuous = `test-results/continuous-${width}-${executions[width]}.json`;
  const journey = await json(continuous);
  assert.equal(journey.result, 'passed');
  assert.deepEqual([...journey.checkpoints].sort(), ['J01','J02','J03','J04','J05','J06','J07']);
  await retain(continuous);
  const activity = `test-results/visual-activity-${width}.json`;
  const facts = await json(activity);
  assert.equal(facts.result, 'passed');
  assert.equal(facts.facts.fee_micro_usd, '21');
  assert.equal(facts.facts.closed_bill_unchanged, true);
  await retain(activity);
  const fault = `test-results/fault-${width}.json`;
  const faultReceipt = await json(fault);
  assert.equal(faultReceipt.phase, 'completed');
  assert.equal(faultReceipt.cases.length, 5);
  for (const entry of faultReceipt.cases) assert.equal(entry.result, entry.case === 'QA35-01-historical-red' ? 'expected_failure_reproduced' : 'passed');
  await retain(fault);
  reviews.push({width, scope: review.scope, pages: review.pages.map(page => ({page: page.page, columns: page.facts.columns, row_actions: page.row_actions, detail: page.detail?.asserted_fields, editor: page.editor?.asserted_labels}))});
}
assert.equal(artifacts.filter(entry => entry.source.endsWith('.png')).length, 63);
await save('artifact-index.json', {host_seq: 46, host_exit_code: 0, log_bytes: 104299, log_truncated: false, source, artifacts});
await save('capture-contract-review.json', {kind: 'Static Review of current real-browser receipts, not a newly executed UI journey', reviews, observations: [{id: 'QA47-01', result: 'Fail', pages: ['models','services','tenants'], reason: 'Frozen major summary information absent from current tables; current capture contract does not assert it.'}, {id: 'QA47-02', result: 'Fail', page: 'records', reason: 'Unqualified Token total remains unknown for a reconciled request with confirmed input 3/output 4 in usage. Original upstream facts may stay immutable; UI must distinguish them from reconciled facts.'}]});
console.log(`PASS material 19/19, same 129-file source, 63 PNG digests, three current reviews/J01-J07/activity/fault receipts; retained ${artifacts.length} artifacts. This is evidence integrity, not product acceptance.`);
