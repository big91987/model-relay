import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFileSync, readdirSync, mkdirSync, writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {dirname} from 'node:path';

const destination = 'docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/qa-23';
const receipt = 'test-results/verification/20261009T162225Z-2545185297';
const material = '.workflow-input/637705186406b80a2ee4afaf0d6b2167/a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49/files';
const digest = bytes => createHash('sha256').update(bytes).digest('hex');
const records = [];
const save = (path, bytes) => {
  if (!path.endsWith('.png')) assert.equal(/\/Users\/|\/private\/var\/|\/var\/folders\/|Bearer\s+\S+|\bsk-[A-Za-z0-9_-]{12,}/.test(bytes.toString()), false, path);
  mkdirSync(dirname(`${destination}/${path}`), {recursive:true});
  writeFileSync(`${destination}/${path}`, bytes);
  records.push({path:`${destination}/${path}`, bytes:bytes.length, sha256:digest(bytes)});
};
const json = (path, value) => save(path, Buffer.from(JSON.stringify(value, null, 2)+'\n'));
const manifest = JSON.parse(readFileSync(`${material}/manifest.json`));
assert.equal(manifest.version, 'product-test33-handoff-r1');
assert.equal(manifest.files.length, 61);
for (const entry of manifest.files) {
  const bytes = readFileSync(`${material}/${entry.path}`);
  assert.equal(bytes.length, entry.bytes, entry.path);
  assert.equal(digest(bytes), entry.sha256, entry.path);
}
const archiveHash = digest(readFileSync(`${material}/../archive.zip`));
assert.equal(archiveHash, 'a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49');
json('material-integrity.json', {version:manifest.version, manifest_path:`${material}/manifest.json`, material_root:material, archive_sha256:archiveHash, checked:61, mismatches:[], files:manifest.files});
const state = JSON.parse(readFileSync(`${receipt}/state.json`));
const log = readFileSync(`${receipt}/verify.log`);
assert.equal(log.length, state.log_bytes);
assert.equal(digest(log), state.log_sha256);
for (const field of ['source_unchanged','candidate_unchanged','owned_group_absent','clean_start_measured']) assert.equal(state[field], true, field);
for (const field of ['node_modules_before','npm_cache_before']) assert.equal(state[field], false, field);
assert.equal(state.exit_code, 0);
assert.equal(state.gate_exit, 0);
assert.deepEqual(state.source_before, state.source_after);
assert.equal(state.source_before.working_tree_sha256, '6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc');
const git = (...args) => execFileSync('git',args,{encoding:'utf8'}).trim();
const identity = {head:git('rev-parse','HEAD'), branch:git('branch','--show-current'), main:git('rev-parse','origin/main'), merge_head:git('rev-parse','MERGE_HEAD'), unmerged:git('diff','--name-only','--diff-filter=U')};
assert.equal(identity.head,'8917f942cb38af2f56aa93914ab62e763bf9cc6e');
assert.equal(identity.main,'6929376089535038112da48aed715fd729615d76');
assert.equal(identity.main,identity.merge_head);
assert.equal(identity.unmerged,'');
const inventory = JSON.parse(execFileSync('node',['docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/verification/collect-integration-exec21.mjs'],{encoding:'utf8',maxBuffer:32*1024*1024}));
json('integration-recomputed.json',inventory);
const source = JSON.parse(execFileSync('go',['run','./tests/evidence'],{encoding:'utf8',env:{...process.env,GOCACHE:`${process.cwd()}/.cache/go-build`,GOTOOLCHAIN:'local'}}));
assert.deepEqual(source,state.source_after);
json('source-before.json',source);
json('host-tests22/state.json',state);
save('host-tests22/verify.log',log);
const prefix = Buffer.from(`go run ./tests/verify\nverification_receipt=${receipt}\n`);
assert.equal(prefix.length+log.length,959549);
const copy = (directory,target) => {
  for (const entry of readdirSync(directory,{withFileTypes:true})) {
    if (entry.isDirectory()) copy(`${directory}/${entry.name}`,`${target}/${entry.name}`);
    else {
      assert.match(entry.name,/\.(png|json|jsonl)$/);
      save(`${target}/${entry.name}`,readFileSync(`${directory}/${entry.name}`));
    }
  }
};
copy(`${receipt}/browser`,'host-tests22/browser');
const browserResults = records.filter(entry => entry.path.endsWith('.json') && entry.path.includes('/browser/')).map(entry => {
  const value = JSON.parse(readFileSync(entry.path));
  return {path:entry.path,suite:value.suite,width:value.width,result:value.result,provider:value.provider,journeys:value.journeys,steps:Array.isArray(value.steps)?value.steps.length:undefined};
});
json('receipt-review.json',{connector_seq:22,connector_exit:0,connector_truncated:false,connector_bytes:959549,local_log_bytes:log.length,local_log_sha256:digest(log),reconstructed_bytes:prefix.length+log.length,comparison:'receipt length and hash plus Connector framing; partial read_command_output not claimed as full independent byte comparison',identity,elapsed_ms:Date.parse(state.finished_utc)-Date.parse(state.started_utc),browser_results:browserResults});
json('archive-index.json',{records});
console.log(JSON.stringify({source,material:61,archive_files:records.length,log_bytes:log.length,identity}));
