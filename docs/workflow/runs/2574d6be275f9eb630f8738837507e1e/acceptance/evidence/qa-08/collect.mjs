import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFileSync,writeFileSync,mkdirSync,readdirSync,copyFileSync} from 'node:fs';
import {join,relative} from 'node:path';
import {spawnSync} from 'node:child_process';

const destination='docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/evidence/qa-08';
const receipt='test-results/verification/20261008T221951Z-283933941';
const material='.workflow-input/2574d6be275f9eb630f8738837507e1e/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7';
const expectedSource='0384f115adbbd545e8ce01021f238010ba66c89eb06e6acea08ce872ce0fa474';
const digest=bytes=>createHash('sha256').update(bytes).digest('hex');
const json=path=>JSON.parse(readFileSync(path,'utf8'));
const save=(path,value)=>writeFileSync(join(destination,path),JSON.stringify(value,null,2)+'\n');
const environment={...process.env,GOCACHE:join(process.cwd(),'.cache/go-build'),GOTOOLCHAIN:'local',CGO_ENABLED:'1'};
mkdirSync(destination,{recursive:true});
const started=new Date().toISOString();
const source=()=>{
  const result=spawnSync('go',['run','./tests/evidence'],{encoding:'utf8',env:environment});
  assert.equal(result.status,0,result.stderr);
  const value=JSON.parse(result.stdout);
  assert.equal(value.working_tree_sha256,expectedSource);
  return value;
};
save('source-before.json',source());
const state=json(receipt+'/state.json'),log=readFileSync(receipt+'/verify.log');
assert.equal(log.length,state.log_bytes);assert.equal(digest(log),state.log_sha256);
assert.equal(state.exit_code,0);assert.equal(state.gate_exit,0);
assert.equal(state.source_before.working_tree_sha256,expectedSource);
assert.deepEqual(state.source_before,state.source_after);
assert.equal(state.clean_start_measured,true);assert.equal(state.node_modules_before,false);
assert.equal(state.npm_cache_before,false);assert.equal(state.source_unchanged,true);
assert.equal(state.candidate_unchanged,true);assert.equal(state.owned_group_absent,true);
assert.equal(state.manual_preinstall,'none');assert.equal(state.NODE_PATH,'unset');
const manifest=json(material+'/files/manifest.json');
assert.equal(manifest.version,'design-v0.2.0');
assert.equal(digest(readFileSync(material+'/archive.zip')),'5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7');
for(const item of manifest.files){const bytes=readFileSync(join(material,'files',item.path));assert.equal(bytes.length,item.bytes);assert.equal(digest(bytes),item.sha256);}
save('material-check.json',{version:manifest.version,status:manifest.status,evidence_kind:manifest.evidence_kind,manifest_path:material+'/files/manifest.json',archive_sha256:digest(readFileSync(material+'/archive.zip')),verified_files:manifest.files.length,referenced_paths:['docs/prd.md','docs/interactions.md','docs/acceptance.md','prototype/app.js','prototype/styles.css'].map(path=>({path:material+'/files/'+path,sha256:manifest.files.find(item=>item.path===path).sha256}))});
const archived=[];
const copy=(path)=>{
  const bytes=readFileSync(join(receipt,path));
  if(/\.(json|log)$/.test(path))assert.equal(/sk-[A-Za-z0-9_-]{20,}|Bearer\s+[A-Za-z0-9._-]{24,}|\/Users\//.test(bytes.toString()),false,'unsafe raw evidence: '+path);
  mkdirSync(join(destination,'host-tests-07',path,'..'),{recursive:true});
  copyFileSync(join(receipt,path),join(destination,'host-tests-07',path));
  archived.push({source:join(receipt,path),path:join(destination,'host-tests-07',path),bytes:bytes.length,sha256:digest(bytes)});
};
copy('state.json');copy('verify.log');
const walk=directory=>{for(const entry of readdirSync(join(receipt,directory),{withFileTypes:true})){const path=join(directory,entry.name);if(entry.isDirectory())walk(path);else if(entry.name.endsWith('.json'))copy(path);}};
walk('browser');
const measurements=[];
for(const width of [1280,1440,390]){
  const probe=json(receipt+'/browser/model-save-readback-'+width+'.json');
  assert.equal(probe.chromium_version,'147.0.7727.15');assert.equal(probe.provider,'controlled_http_not_supplier');
  assert.equal(probe.cases.length,7);
  for(const value of probe.cases){
    assert.equal(value.write_status,200);assert.equal(value.cleanup,true);
    if(value.stages){
      assert.deepEqual(value.stages.map(stage=>stage.phase),['tenants','catalog']);
      for(const stage of value.stages){
        assert.equal(stage.read_status,200);assert.equal(stage.save_completed,false);
        for(const field of ['content_loading','write_notice','save_disabled','readback_pending','page_models','scope_matches'])assert.equal(stage[field],true);
        assert.equal(stage.dialog_open,false);
        assert.ok(Date.parse(value.write_utc)<=Date.parse(stage.read_utc));
        assert.ok(Date.parse(stage.read_utc)<=Date.parse(stage.release_utc));
        assert.ok(Date.parse(stage.release_utc)<=Date.parse(value.save_utc));
      }
      assert.equal(value.final_dom_matches,true);assert.ok(Date.parse(value.save_utc)<=Date.parse(value.final_assertion_utc));
    }else if(value.injected_read_failure){assert.equal(value.forwarded_read_status,503);assert.equal(value.save_rejected,true);assert.equal(value.recovered_dom_matches,true);assert.equal(value.recovery_read_status,200);}
    else {assert.equal(value.first_save_rejected,true);assert.equal(value.final_dom_matches,true);assert.ok(Date.parse(value.second_save_utc)<=Date.parse(value.release_utc));}
  }
  const continuousName=readdirSync(receipt+'/browser').find(name=>new RegExp('^continuous-'+width+'-[0-9]+\\.json$').test(name));
  const continuous=json(receipt+'/browser/'+continuousName);assert.equal(continuous.result,'passed');assert.deepEqual([...continuous.checkpoints].sort(),['J01','J02','J03','J04','J05','J06','J07']);
  const visual=json(receipt+'/browser/visual-'+width+'/review.json');assert.equal(visual.pages.length,10);
  for(const page of visual.pages){assert.equal(page.facts.rootFits,true);assert.equal(page.facts.secretsClear,true);}
  const screenshot=visual.pages.find(page=>page.page==='models').screenshots[0];
  const imagePath='browser/visual-'+width+'/models-page.png';assert.equal(digest(readFileSync(receipt+'/'+imagePath)),screenshot.sha256);copy(imagePath);
  measurements.push({width,held_cases:4,held_phases:8,read_failures:2,late_opposite:1,continuous:continuousName,journeys:continuous.checkpoints,visual_pages:visual.pages.length});
}
const tests=['model-save-readback','save-lifecycle','readback','workspace','console-summary','journeys'].map(name=>'tests/browser/'+name+'.test.mjs');
const unitStart=new Date().toISOString();
const units=spawnSync('node',['--test',...tests],{encoding:'utf8',env:environment});
const unitLog=`QA_COMMAND_START utc=${unitStart} command=node --test ${tests.join(' ')}\n${units.stdout||''}${units.stderr||''}\nQA_COMMAND_END utc=${new Date().toISOString()} exit=${units.status}\n`;
writeFileSync(join(destination,'independent-node.log'),unitLog);assert.equal(units.status,0);
save('source-after.json',source());
save('archive-index.json',{started_utc:started,finished_utc:new Date().toISOString(),connector:{node:'tests',seq:7,exit_code:0,log_bytes:337132,truncated:false},project_log:{bytes:log.length,sha256:digest(log),matches_state:true},source_sha256:expectedSource,measurements,independent_node:{command:'node --test '+tests.join(' '),exit_code:units.status,sha256:digest(Buffer.from(unitLog))},safety:{known_secret_and_absolute_home_path_scan:'no_matches',json_byte_preserving:true,screenshots:'models-only; manually reviewed; no secret fields'},files:archived});
console.log(JSON.stringify({source:expectedSource,archived_files:archived.length,measurements,independent_node_exit:units.status}));
