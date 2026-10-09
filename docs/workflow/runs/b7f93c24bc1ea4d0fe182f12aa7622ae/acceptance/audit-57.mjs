import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFileSync, writeFileSync, mkdirSync, readdirSync, copyFileSync, statSync} from 'node:fs';
import path from 'node:path';

const run='docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae';
const original='test-results/verification/20261009T110042Z-3580683419';
const destination=run+'/acceptance/evidence/attempt-57';
const material='.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7';
const digest=data=>createHash('sha256').update(data).digest('hex');
const json=file=>JSON.parse(readFileSync(file,'utf8'));
const save=(file,data)=>writeFileSync(file,JSON.stringify(data,null,2)+'\n');
const filesUnder=root=>readdirSync(root,{withFileTypes:true}).flatMap(entry=>entry.isDirectory()?filesUnder(path.join(root,entry.name)):[path.join(root,entry.name)]).sort();
mkdirSync(destination,{recursive:true});
const source=json(destination+'/source-current.json');
const state=json(original+'/state.json');
assert.deepEqual(source,state.source_before);
assert.deepEqual(source,state.source_after);
assert.deepEqual(source,json(run+'/delivery/verification/development-55/source-final.json'));
assert.equal(state.exit_code,0);
assert.equal(state.gate_exit,0);
assert.equal(state.owned_group_absent,true);
assert.equal(state.clean_start_measured,true);
assert.equal(state.node_modules_before,false);
assert.equal(state.npm_cache_before,false);
assert.equal(state.manual_preinstall,'none');
assert.equal(state.source_unchanged,true);
assert.equal(state.candidate_unchanged,true);
const log=readFileSync(original+'/verify.log');
assert.equal(digest(log),state.log_sha256);
assert.equal(log.length,state.log_bytes);
assert.ok(log.toString().includes('ℹ tests 276'));
assert.ok(log.toString().includes('ℹ pass 276'));
assert.ok(log.toString().includes('ℹ fail 0'));
assert.ok(log.toString().includes('ℹ skipped 0'));
const manifest=json(material+'/files/manifest.json');
assert.equal(manifest.version,'design-v0.2.0');
const references=manifest.files.map(entry=>{
  const actualPath=material+'/files/'+entry.path;
  const data=readFileSync(actualPath);
  assert.equal(data.length,entry.bytes);
  assert.equal(digest(data),entry.sha256);
  return {...entry,actual_path:actualPath,match:true};
});
assert.equal(digest(readFileSync(material+'/archive.zip')),'5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7');
const integrity=filesUnder(original).map(file=>{
  const data=readFileSync(file);
  if(!file.endsWith('.png'))assert.ok(!/\/Users\/|\/private\/|\/var\/folders\//.test(data.toString()),'no absolute local paths in archived text');
  const relative=path.relative(original,file);
  const target=destination+'/host/'+relative;
  mkdirSync(path.dirname(target),{recursive:true});
  copyFileSync(file,target);
  assert.equal(digest(readFileSync(target)),digest(data));
  return {original:file,path:target,bytes:data.length,sha256:digest(data),original_mtime_utc:statSync(file).mtime.toISOString(),byte_identical:true};
});
save(destination+'/integrity.json',integrity);
const browser=destination+'/host/browser';
const sidecars=filesUnder(browser).filter(file=>file.endsWith('.png.integrity.json')).map(file=>{
  const data=json(file);
  assert.equal(data.result,'passed');
  assert.equal(digest(readFileSync(file.replace(/\.integrity\.json$/,''))),data.sha256);
  assert.equal(data.pixels.surface.uniform_rgb,false);
  assert.ok(data.pixels.anchors.length>=2);
  assert.ok(data.pixels.anchors.every(anchor=>anchor.inside&&anchor.ink_present));
  assert.equal(data.geometry.scroll_delta.x,0);
  assert.equal(data.geometry.scroll_delta.y,0);
  for(const field of data.geometry.fields.filter(field=>['title','content'].includes(field.field)))assert.equal(field.document_stable,true);
  return {path:file,result:data.result,sha256_match:true,anchors:data.pixels.anchors.length,scroll_delta:data.geometry.scroll_delta};
});
const receipts=[1280,1440,390].map(width=>{
  const visual=json(browser+`/visual-${width}/review.json`);
  assert.equal(visual.result,'passed');
  assert.equal(visual.pages.length,10);
  for(const page of visual.pages){
    assert.equal(page.facts.rootFits,true);
    assert.equal(page.facts.secretsClear,true);
    for(const [field,value] of Object.entries(page.readability))if(typeof value==='boolean')assert.equal(value,true);
    for(const screenshot of page.screenshots){
      const actual=browser+'/'+screenshot.path.replace(/^test-results\//,'');
      assert.equal(digest(readFileSync(actual)),screenshot.sha256);
    }
  }
  const matching=filesUnder(browser).filter(file=>new RegExp(`/self-hosted-${width}-\\d+\\.json$`).test(file));
  assert.equal(matching.length,1);
  const selfHosted=json(matching[0]);
  assert.equal(selfHosted.result,'passed');
  assert.deepEqual(selfHosted.initial_counts,{upstreams:0,models:0,tenants:0});
  assert.equal(selfHosted.business_writes,'native_user_actions_only');
  assert.equal(selfHosted.steps.length,12);
  for(const step of selfHosted.steps){
    if(step.readability)for(const value of Object.values(step.readability))if(typeof value==='boolean')assert.equal(value,true);
    if(step.same_request){assert.equal(step.same_request,true);assert.equal(step.fee_micro_usd,'21');assert.equal(step.cost_micro_usd,'7');}
  }
  const continuous=json(filesUnder(browser).find(file=>new RegExp(`/continuous-${width}-\\d+\\.json$`).test(file)));
  assert.equal(continuous.result,'passed');
  assert.deepEqual([...continuous.checkpoints].sort(),['J01','J02','J03','J04','J05','J06','J07']);
  return {width,visual_result:visual.result,pages:visual.pages.map(page=>({page:page.page,readability:page.readability,screenshots:page.screenshots.length})),modal_scroll_checks:visual.modal_scroll_checks,continuous,self_hosted_receipt:matching[0],self_hosted_steps:selfHosted.steps.map(step=>step.step),self_hosted_result:selfHosted.result};
});
const summary={qa_execution:57,checked_utc:new Date().toISOString(),mode:'full acceptance evidence review; no QA host browser execution',candidate:source,host_source_matches:true,connector_execution:56,log:{bytes:log.length,sha256:digest(log),state_match:true,connector_wrapper_bytes:729568,wrapper_is_not_raw_verify_log:true},archive:{files:integrity.length,png:integrity.filter(entry=>entry.path.endsWith('.png')).length,bytes:integrity.reduce((total,entry)=>total+entry.bytes,0),all_bytes_match:true},gate:{started_utc:state.started_utc,finished_utc:state.finished_utc,exit_code:state.exit_code,elapsed_ms:Date.parse(state.finished_utc)-Date.parse(state.started_utc),cold_start:true,owned_group_absent:true,dependency_receipt:state.dependency_receipt,versions:state.versions},material:{version:manifest.version,status:manifest.status,archive_sha256:'5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7',references},screenshots:{sidecars:sidecars.length,all_sidecars_passed:true,all_hashes_match:true,all_anchors_present:true,geometry_stable:true},receipts};
save(destination+'/audit-summary.json',summary);
save(destination+'/sidecar-audit.json',sidecars);
console.log(JSON.stringify({archive:summary.archive,candidate:source.working_tree_sha256,log:summary.log,sidecars:sidecars.length,viewports:receipts.map(receipt=>({width:receipt.width,visual:receipt.visual_result,self_hosted:receipt.self_hosted_result}))}));
