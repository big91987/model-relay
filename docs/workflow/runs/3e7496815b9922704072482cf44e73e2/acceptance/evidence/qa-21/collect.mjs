import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFile,writeFile,mkdir,readdir,copyFile} from 'node:fs/promises';
import path from 'node:path';

const destination='docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/evidence/qa-21';
const source='test-results/verification/20261008T192336Z-2658854976';
const digest='f8befd158a543afeb1ba77778e4caf72e876f6e4b0b988dd4e6ab73de5e5eaab';
const hash=data=>createHash('sha256').update(data).digest('hex');
const state=JSON.parse(await readFile(source+'/state.json'));
const log=await readFile(source+'/verify.log');
assert.equal(log.length,state.log_bytes);
assert.equal(hash(log),state.log_sha256);
assert.equal(state.exit_code,0);
assert.equal(state.gate_exit,0);
assert.equal(state.clean_start_measured,true);
assert.equal(state.node_modules_before,false);
assert.equal(state.npm_cache_before,false);
assert.equal(state.NODE_PATH,'unset');
assert.equal(state.manual_preinstall,'none');
assert.equal(state.source_unchanged,true);
assert.equal(state.candidate_unchanged,true);
assert.deepEqual(state.source_before,state.source_after);
assert.equal(state.source_before.working_tree_sha256,digest);
assert.equal(state.source_before.files,147);
assert.equal(state.owned_group_absent,true);
assert.equal(state.cleanup,'owned_process_group_cleanup_attempted; candidate_removed');
assert.ok(state.auxiliary_commands.every(entry=>entry.exit===0&&!entry.cancelled&&entry.owned_group_absent));
const unsafe=/(?:\/Users\/|\/private\/var\/folders\/|Bearer\s+[A-Za-z0-9_-]{20,}|sk-[A-Za-z0-9_-]{20,})/;
assert.equal(unsafe.test(log.toString()),false,'unsafe log');
await mkdir(destination+'/host-seq20',{recursive:true});
const artifacts=[];
async function archive(relative) {
  const data=await readFile(source+'/'+relative);
  if(/\.(?:json|jsonl|log)$/.test(relative))assert.equal(unsafe.test(data.toString()),false,relative);
  const target=destination+'/host-seq20/'+relative;
  await mkdir(path.dirname(target),{recursive:true});
  await copyFile(source+'/'+relative,target);
  assert.equal(hash(await readFile(target)),hash(data));
  artifacts.push({path:'host-seq20/'+relative,bytes:data.length,sha256:hash(data)});
}
async function walk(relative) {
  for(const entry of await readdir(source+'/'+relative,{withFileTypes:true})) {
    const child=relative+'/'+entry.name;
    if(entry.isDirectory())await walk(child);
    else {assert.ok(/\.(?:png|json|jsonl)$/.test(child),child);await archive(child);}
  }
}
await archive('verify.log');await archive('state.json');await walk('browser');
const widths=[1280,1440,390];
const browser=[];
for(const width of widths) {
  const candidates=artifacts.filter(entry=>new RegExp('browser/continuous-'+width+'-\\d+\\.json$').test(entry.path));
  assert.equal(candidates.length,1);
  const continuous=JSON.parse(await readFile(destination+'/'+candidates[0].path));
  assert.equal(continuous.result,'passed');
  assert.equal(continuous.provider,'controlled_http_not_supplier');
  assert.deepEqual([...continuous.checkpoints].sort(),['J01','J02','J03','J04','J05','J06','J07']);
  for(const relative of ['management-summary-'+width+'.json','visual-'+width+'/review.json','fault-'+width+'.json']) {
    const data=JSON.parse(await readFile(source+'/browser/'+relative));
    if(data.suite==='console-faults') {
      assert.equal(data.cases.length,5);
      for(const entry of data.cases)assert.equal(entry.result,entry.case==='QA35-01-historical-red'?'expected_failure_reproduced':'passed');
    } else assert.equal(data.result,'passed',relative);
    if(data.suite==='console-visual')assert.equal(data.pages.length,10);
  }
  browser.push({width,continuous:candidates[0].path,summary:'host-seq20/browser/management-summary-'+width+'.json',visual:'host-seq20/browser/visual-'+width+'/review.json',faults:'host-seq20/browser/fault-'+width+'.json'});
}
const materialRoot='.workflow-input/3e7496815b9922704072482cf44e73e2/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7';
const manifest=JSON.parse(await readFile(materialRoot+'/files/manifest.json'));
assert.equal(hash(await readFile(materialRoot+'/archive.zip')),'5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7');
const material=[];
for(const entry of manifest.files){const actual=await readFile(materialRoot+'/files/'+entry.path);assert.equal(hash(actual),entry.sha256);assert.equal(actual.length,entry.bytes);material.push({path:materialRoot+'/files/'+entry.path,sha256:entry.sha256,bytes:entry.bytes});}
const text=log.toString();
const projection={candidate:digest,receipt:source,runner_bytes:log.length,runner_sha256:hash(log),connector:{seq:20,bytes:310815,exit_code:0,truncated:false,offsets:[0,32768,65536,98304,131072,163840,196608,229376,262144,294912],next_offset:310815,eof:true,full_runner_is_not_connector_full_output:true},browser,material:{version:manifest.version,status:manifest.status,archive_sha256:'5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7',files:material},gate_lines:text.split('\n').filter(line=>/^(?:PASS |ok\s|go test |go build |go vet |all modules |ℹ |BROWSER_PHASE|CLEAN_END)/.test(line)),skips:text.split('\n').filter(line=>/--- SKIP:/.test(line)),artifacts};
await writeFile(destination+'/archive-index.json',JSON.stringify(projection,null,2)+'\n');
console.log(JSON.stringify({runner_bytes:log.length,runner_sha256:hash(log),archived:artifacts.length,widths,material_files:material.length,skips:projection.skips,gate_lines:projection.gate_lines},null,2));
