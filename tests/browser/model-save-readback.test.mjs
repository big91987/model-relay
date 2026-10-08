import assert from 'node:assert/strict';
import test from 'node:test';
import {readbackPhase,waitForSaveReadback} from './model-save-readback.mjs';

test('post-write probe excludes pre-write, other scopes, methods and detail routes',()=>{
  const request=(path,method='GET')=>({url:()=> 'http://isolated.test'+path,method:()=>method});
  for(const suffix of ['', '?page=1&page_size=100&limit=100', '?cursor=next'])assert.equal(readbackPhase(request('/admin/api/tenants'+suffix),'isolated',true),'tenants');
  assert.equal(readbackPhase(request('/admin/api/catalog?tenant=isolated&include_ungranted=true'),'isolated',true),'catalog');
  for(const [path,method,acknowledged] of [
    ['/admin/api/tenants','GET',false],
    ['/admin/api/tenants','POST',true],
    ['/admin/api/tenants/detail','GET',true],
    ['/admin/api/models','GET',true],
    ['/admin/api/catalog?tenant=other&include_ungranted=true','GET',true],
    ['/admin/api/catalog?tenant=isolated','GET',true]
  ])assert.equal(readbackPhase(request(path,method),'isolated',acknowledged),null);
});

for(const condition of ['complete','pending','failed','interrupted','other-operation','other-scope','other-page'])test(`save completion predicate distinguishes ${condition} without retrying a summary assertion`,async()=>{
  const expected={operation:'controlled-operation',tenant:'isolated',page:'模型托管'};
  const form={dataset:{operation:condition==='other-operation'?'new-operation':expected.operation,readback:condition}};
  const elements={'editor-form':form,tenant:{value:condition==='other-scope'?'other':expected.tenant},'page-title':{textContent:condition==='other-page'?'API Keys':expected.page}};
  const previous=globalThis.document;globalThis.document={getElementById:id=>elements[id]};
  let waits=0,evaluations=0;
  const pending=new Error('controlled pending wait');
  const page={waitForFunction:async(predicate,value)=>{waits++;if(!predicate(value))throw pending;},evaluate:async(callback,value)=>{evaluations++;return callback(value);}};
  try {
    if(condition==='complete')await waitForSaveReadback(page,expected);
    else if(condition==='pending')await assert.rejects(waitForSaveReadback(page,expected),error=>error===pending);
    else await assert.rejects(waitForSaveReadback(page,expected),{name:'AssertionError'});
    assert.equal(waits,1);assert.equal(evaluations,condition==='pending'?0:1);
  } finally {globalThis.document=previous;}
});
