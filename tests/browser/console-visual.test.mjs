import assert from 'node:assert/strict';
import {test} from 'node:test';
import {mkdtemp,readFile,rm} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {assertCaptureFacts,consoleVisualReview,overviewReadFact,visualEditorControl,visualPages} from './console-visual.mjs';

const passing=()=>({title:true,scope:true,role:true,ready:true,rootFits:true,secretsClear:true,columns:{field:true},text:{summary:true},actions:{edit:true},labels:{input:true},dataRows:1,form:false,cards:0});

test('visual contract names exactly all ten approved pages with substantive checks',()=>{
  assert.deepEqual(visualPages.map(page=>page.id),['overview','models','services','keys','tenants','limits','usage','billing','records','playground']);
  for(const page of visualPages) assert.ok((page.columns?.length??0)+(page.text?.length??0)+(page.labels?.length??0)>=4,page.id);
  assert.equal(visualPages.filter(page=>page.detail).length,5);
  assert.equal(visualPages.filter(page=>page.editor).length,6);
});
test('capture contract rejects title-only, failure, wrong role/scope, overflow and secrets',()=>{
  assertCaptureFacts(passing());
  for(const flag of ['title','scope','role','ready','rootFits','secretsClear']) {
    assert.throws(()=>assertCaptureFacts({...passing(),[flag]:false}));
  }
  for(const group of ['columns','text','actions','labels']) assert.throws(()=>assertCaptureFacts({...passing(),[group]:{missing:false}}));
  assert.throws(()=>assertCaptureFacts({...passing(),dataRows:0}));
  assertCaptureFacts({...passing(),dataRows:0,form:true});
  assertCaptureFacts({...passing(),dataRows:0,cards:1});
  assertCaptureFacts({...passing(),minimumRows:2,primaryRows:2});
  assert.throws(()=>assertCaptureFacts({...passing(),minimumRows:2,primaryRows:1}));
});

test('limits visual editor targets its heading action outside content',async()=>{
  const definition=visualPages.find(page=>page.id==='limits');
  const html=await readFile(new URL('../../internal/relay/web/workspace.html',import.meta.url),'utf8');
  assert.match(html,/<div class="page-heading">[\s\S]*id="page-action"[\s\S]*<div id="content"/);
  const expected={};
  const heading={getByRole:(role,options)=>{assert.equal(role,'button');assert.deepEqual(options,{name:'调整租户限额',exact:true});return expected;}};
  const page={locator:selector=>{assert.equal(selector,'.page-heading');return heading;}};
  const content={getByRole:()=>assert.fail('heading action must not be searched inside content')};
  assert.equal(visualEditorControl(page,definition,content),expected);
});

test('all other visual editors retain exact current row or content ownership',()=>{
  const page={locator:()=>assert.fail('row editors must not broaden to heading or page')};
  for(const definition of visualPages.filter(value=>value.editor&&value.id!=='limits')) {
    const expected={};
    const row={getByRole:(role,options)=>{assert.equal(role,'button');assert.deepEqual(options,{name:definition.editor.action,exact:true});return expected;}};
    assert.equal(visualEditorControl(page,definition,row),expected);
  }
});

test('missing heading editor has no content or page-wide fallback',()=>{
  const definition=visualPages.find(page=>page.id==='limits');
  const absent={count:async()=>0};
  const page={locator:selector=>{assert.equal(selector,'.page-heading');return {getByRole:()=>absent};}};
  const content={getByRole:()=>assert.fail('no fallback to a different owner')};
  assert.equal(visualEditorControl(page,definition,content),absent);
});

test('visual read diagnostic keeps only current overview GET status and scope match',()=>{
  const response=(url,method='GET',status=200)=>({url:()=>url,request:()=>({method:()=>method}),status:()=>status});
  assert.deepEqual(overviewReadFact(response('http://localhost/admin/api/workspace/overview?tenant=scope&token=never-log'),'scope'),{resource:'workspace/overview',status:200,scope_matches:true});
  assert.deepEqual(overviewReadFact(response('http://localhost/admin/api/workspace/overview?tenant=other','GET',503),'scope'),{resource:'workspace/overview',status:503,scope_matches:false});
  assert.equal(overviewReadFact(response('http://localhost/admin/api/workspace/bills/private-id','GET',404),'scope'),null);
  assert.equal(overviewReadFact(response('http://localhost/admin/api/workspace/overview','POST'),'scope'),null);
});

test('failed visual capture persists facts before assertion without relaxing empty gate',async context=>{
  const diagnostic=[];
  context.mock.method(console,'log',message=>diagnostic.push(JSON.parse(message)));
  const previous=process.cwd(),directory=await mkdtemp(join(tmpdir(),'relay-visual-contract-'));
  let listener=null,removed=false;
  const facts={...passing(),ready:false,empty_states:{total:1,visible:1,seven_day_no_calls:1,no_data:0,other:0},settlement:{loading:false,error_nodes:0,notice_present:false},dimensions:{root_scroll_width:1280,viewport_width:1280,content_width:1016}};
  const page={on:(event,callback)=>{assert.equal(event,'response');listener=callback;},off:(event,callback)=>{assert.equal(event,'response');assert.equal(callback,listener);removed=true;},evaluate:async()=>facts};
  try {
    process.chdir(directory);
    await assert.rejects(consoleVisualReview(page,{width:1280,tenant:'scope',objects:{},step:()=>{},nav:async()=>listener({url:()=> 'http://localhost/admin/api/workspace/overview?tenant=scope&token=never-log',request:()=>({method:()=> 'GET'}),status:()=>200})}),/no loading\/failure\/empty masquerading as data/);
    const raw=await readFile('test-results/visual-1280/review.json','utf8'),receipt=JSON.parse(raw);
    assert.deepEqual(receipt.pages,[]);
    assert.equal(receipt.result,'failed');
    assert.equal(receipt.failure.capture_assertion,'no loading/failure/empty masquerading as data');
    assert.equal(receipt.failure.check,'capture-facts');
    assert.deepEqual(receipt.pending_capture.facts,facts);
    assert.deepEqual(receipt.pending_capture.http_reads,[{resource:'workspace/overview',status:200,scope_matches:true}]);
    assert.equal(raw.includes('never-log'),false);
    assert.equal(raw.includes('localhost'),false);
    assert.deepEqual(diagnostic[0].capture,receipt.pending_capture);
    assert.equal(removed,true);
  } finally {
    process.chdir(previous);
    await rm(directory,{recursive:true,force:true});
  }
});

test('navigation failure retains null facts and omits arbitrary error text',async context=>{
  context.mock.method(console,'log',()=>{});
  const previous=process.cwd(),directory=await mkdtemp(join(tmpdir(),'relay-visual-navigation-'));
  let removed=false;
  const page={on:()=>{},off:()=>{removed=true;},evaluate:()=>assert.fail('DOM facts must not be invented after failed navigation')};
  try {
    process.chdir(directory);
    await assert.rejects(consoleVisualReview(page,{width:1280,tenant:'scope',objects:{},step:()=>{},nav:async()=>{const error=new Error('never-log-arbitrary-browser-secret');error.name='TimeoutError';throw error;}}),/never-log-arbitrary-browser-secret/);
    const raw=await readFile('test-results/visual-1280/review.json','utf8'),receipt=JSON.parse(raw);
    assert.equal(receipt.failure.kind,'timeout');
    assert.equal(receipt.failure.check,'navigation');
    assert.equal(receipt.failure.capture_assertion,null);
    assert.equal(receipt.pending_capture.facts,null);
    assert.deepEqual(receipt.pending_capture.http_reads,[]);
    assert.equal(raw.includes('never-log'),false);
    assert.equal(removed,true);
  } finally {
    process.chdir(previous);
    await rm(directory,{recursive:true,force:true});
  }
});
