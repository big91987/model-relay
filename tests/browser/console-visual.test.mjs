import assert from 'node:assert/strict';
import {test} from 'node:test';
import {mkdtemp,readFile,rm} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {assertCaptureFacts,assertModalClosed,assertModalScrollFacts,closeVisualModal,consoleVisualReview,modalCloseFailureFacts,overviewReadFact,visualEditorControl,visualPages} from './console-visual.mjs';

test('modal scroll contract requires native ownership, clipped background, internal scroll and focus',()=>{
  const facts={native_modal:true,root_overflow:'clip',body_position:'fixed',dialog_overflow:'auto',background_scroll_delta:0,focus_inside:true};
  assertModalScrollFacts(facts);
  for(const [field,value] of Object.entries({native_modal:false,root_overflow:'hidden',body_position:'static',dialog_overflow:'clip',background_scroll_delta:145,focus_inside:false})) {
    assert.throws(()=>assertModalScrollFacts({...facts,[field]:value}));
  }
  assert.throws(()=>assertModalScrollFacts({...facts,background_scroll_delta:null}));
});

test('modal close contract rejects a remaining modal, retained root lock or lost trigger focus',async()=>{
  const check=(facts,focused)=>assertModalClosed({evaluate:async()=>facts},{evaluate:async()=>({focused})});
  await check({modal_open:false,root_overflow:'visible'},true);
  await assert.rejects(check({modal_open:true,root_overflow:'visible'},true));
  await assert.rejects(check({modal_open:false,root_overflow:'clip'},true));
  await assert.rejects(check({modal_open:false,root_overflow:'visible',body_position:'fixed'},true));
  await assert.rejects(check({modal_open:false,root_overflow:'visible'},false));
});

test('modal close failures identify the exact unchanged assertion with safe facts',async()=>{
  const base={modal_open:false,root_overflow:'visible',body_position:'static',body_lock_class:false,scroll_x:0,scroll_y:37};
  const trigger={connected:true,focused:true,focus_in_dialog:false,focus_is_body:false,focus_is_root:false};
  const event={available:true,close_events:1,modal_open_at_event:false,body_locked_at_event:false,body_fixed_at_event:false,root_clipped_at_event:false};
  for(const [boundary,override,focus] of [['native-ownership',{modal_open:true},true],['root-unlock',{root_overflow:'clip'},true],['body-unlock',{body_position:'fixed',body_lock_class:true},true],['trigger-focus',{},false]]) {
    await assert.rejects(assertModalClosed({evaluate:async()=>({...base,...override,private:'never-log'})},{evaluate:async()=>({...trigger,focused:focus,private:'never-log'})},{evaluate:async()=>event}),error=>{
      const facts=modalCloseFailureFacts(error);
      assert.equal(facts.boundary,boundary);assert.equal(facts.trigger_focused,focus);assert.deepEqual(facts.event,event);
      assert.equal(JSON.stringify(facts).includes('never-log'),false);
      assert.equal(facts.scroll_y,37);
      return true;
    });
  }
  assert.deepEqual(modalCloseFailureFacts(new Error('never-log')),{available:false});
});

test('close observer is passive, records event delivery, and cleans up without retrying assertions',async()=>{
  const sequence=[],listeners=new Map();
  const body={classList:{contains:()=>false}},root={};
  const dialog={addEventListener:(event,listener)=>listeners.set(event,listener),removeEventListener:(event,listener)=>{assert.equal(listeners.get(event),listener);listeners.delete(event);}};
  const previous={document:globalThis.document,getComputedStyle:globalThis.getComputedStyle};
  globalThis.document={body,documentElement:root,querySelector:selector=>selector==='#detail'?dialog:null};
  globalThis.getComputedStyle=element=>element===body?{position:'static'}:{overflowY:'visible'};
  let observed;
  const page={evaluateHandle:async(callback,selector)=>{sequence.push('arm');observed=callback(selector);return {evaluate:async read=>read(observed),dispose:async()=>sequence.push('dispose')};},evaluate:async()=>({modal_open:false,root_overflow:'visible',body_position:'static'}),getByRole:()=>({click:async()=>{sequence.push('click');listeners.get('close')();}}),locator:()=>({waitFor:async()=>sequence.push('hidden')})};
  try {
    const result=await closeVisualModal(page,{evaluate:async()=>({focused:true})},'#detail','关闭详情');
    assert.equal(result.event.close_events,1);assert.equal(result.event.body_fixed_at_event,false);
    assert.deepEqual(sequence,['arm','click','hidden','dispose']);assert.equal(listeners.size,0);
    page.getByRole=()=>({click:async()=>sequence.push('click')});
    await assert.rejects(closeVisualModal(page,{evaluate:async()=>({focused:false})},'#detail','关闭详情'),error=>{
      assert.equal(modalCloseFailureFacts(error).boundary,'trigger-focus');assert.equal(modalCloseFailureFacts(error).event.close_events,0);return true;
    });
    assert.equal(listeners.size,0);assert.equal(sequence.filter(item=>item==='click').length,2);
  } finally {
    for(const [name,value] of Object.entries(previous))if(value===undefined)delete globalThis[name];else globalThis[name]=value;
  }
});

test('close observer cleanup cannot replace the original strict assertion failure',async()=>{
  const page={evaluateHandle:async()=>({evaluate:async callback=>{
    if(callback.toString().includes('finish'))throw new Error('private-cleanup');
    return {available:true,close_events:0};
  },dispose:async()=>{throw new Error('private-dispose');}}),evaluate:async()=>({modal_open:false,root_overflow:'visible',body_position:'fixed'}),getByRole:()=>({click:async()=>{}}),locator:()=>({waitFor:async()=>{}})};
  await assert.rejects(closeVisualModal(page,{evaluate:async()=>({focused:true})},'#detail','关闭详情'),error=>{
    assert.equal(error.name,'AssertionError');assert.equal(modalCloseFailureFacts(error).boundary,'body-unlock');
    assert.equal(JSON.stringify(modalCloseFailureFacts(error)).includes('private-'),false);return true;
  });
});

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
