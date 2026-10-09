import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';
import {assertReadability,readabilityEvidence,readabilityFailureFacts,scrollBoundaryFacts,startScrollEvents,finishScrollEvents,startKeyboardScrollCompletion,startWheelDispatchCompletion} from './readability.mjs';

const source=readFileSync(new URL('../../internal/relay/web/workspace.js',import.meta.url),'utf8');
const context=vm.createContext({node:(tag,text,className)=>({tag,textContent:text,className,attributes:{},setAttribute(name,value){this.attributes[name]=value;}})});
vm.runInContext(source.slice(source.indexOf('function enumLabel('),source.indexOf('function modelSources(')),context);

test('known call, check, invitation, source and role labels express the actual business result',()=>{
  for(const [field,value,record,expected] of [
    ['state','success',{},'成功'],['state','success',{execution_revision:2},'核验通过'],
    ['state','failed',{execution_revision:2},'核验失败'],['state','checking',{execution_revision:2},'核验中'],
    ['state','pending',{account:'controlled'},'待接受'],['state','accepted',{account:'controlled'},'已接受'],
    ['state','revoked',{account:'controlled'},'已撤销'],['state','expired',{account:'controlled'},'已过期'],
    ['state','upstream_failed',{},'上游失败'],['state','process_interrupted',{},'进程中断'],
    ['role','administrator',{},'租户管理员'],['role','developer',{},'开发者'],['role','finance',{},'财务观察者'],
    ['source_kind','provider',{},'供应商API'],['source_kind','self_hosted',{},'已有自托管']
  ])assert.equal(context.enumLabel(field,value,record),expected);
});

test('unknown enum values remain distinguishable and prototype keys cannot pretend to be states',()=>{
  for(const field of ['state','role','source_kind'])for(const value of ['future_value','__proto__','constructor','<not-markup>'])assert.equal(context.enumLabel(field,value),'未知（'+value+'）');
  assert.equal(context.enumLabel('state',null),'未知');
  for(const field of ['fee_micro_usd','total_tokens','error_code','protocol','url'])assert.equal(context.enumLabel(field,'precise_raw_value'),'precise_raw_value');
});

test('full technical identifiers use native text, retain exact content and expose keyboard scroll access',()=>{
  const value='model_identifier_'.repeat(20)+'<literal>';
  const element=context.technicalValue(value);
  assert.equal(element.textContent,value);assert.equal(element.tabIndex,0);assert.equal(element.className,'technical-value');
  assert.equal(element.attributes.role,'region');assert.equal(element.attributes['aria-label'],value);
  assert.equal(element.attributes['aria-description'],'完整值，可横向滚动查看');
});

test('technical region names preserve each public identity instead of replacing it with scroll instructions',()=>{
  for(const value of ['journey-model-1','journey-model-10','https://controlled.test/long/path','rk_public_hint','<literal>',null]) {
    const element=context.technicalValue(value);
    assert.equal(element.attributes['aria-label'],value==null?'未知':value);
    assert.equal(element.attributes['aria-label'],element.textContent);
    assert.notEqual(element.attributes['aria-label'],element.attributes['aria-description']);
  }
});

test('empty error codes are absent facts, not empty scrollable identifiers',()=>{
  const element=context.technicalValue('');
  assert.equal(element.textContent,'—');assert.equal(element.className,undefined);assert.equal(element.tabIndex,undefined);
});

test('readability gate rejects overflow, hidden or clipped identifiers and unreachable table columns',()=>{
  const passing={root_fits:true,tables_accessible:true,headers_unbroken:true,identifiers_complete:true,actions_unbroken:true,horizontal_boundaries_isolated:true};
  assertReadability(passing);
  for(const field of Object.keys(passing))assert.throws(()=>assertReadability({...passing,[field]:false}));
});

test('local horizontal boundary policy is required without disabling normal overflow or vertical page scroll',()=>{
  const styles=readFileSync(new URL('../../internal/relay/web/workspace.css',import.meta.url),'utf8');
  assert.match(styles,/\.table-wrap,\.technical-value\{overscroll-behavior-x:none\}/);
  assert.doesNotMatch(styles,/overscroll-behavior(?:-y)?\s*:/);
  assert.match(styles,/\.table-wrap\{overflow:auto/);
  assert.match(styles,/\.technical-value\{[^}]*overflow-x:auto/);
  const passing={root_fits:true,tables_accessible:true,headers_unbroken:true,identifiers_complete:true,actions_unbroken:true,horizontal_boundaries_isolated:true};
  for(const value of [false,undefined,null,'none'])assert.throws(()=>assertReadability({...passing,horizontal_boundaries_isolated:value}));
});

test('native evidence reads the horizontal policy for both tables and identifiers rather than assuming isolation',async()=>{
  for(const [tablePolicy,identifierPolicy] of [['none','none'],['auto','none'],['none','auto'],['contain','none'],['none',undefined]]) {
    const element=policy=>({tabIndex:0,textContent:'public-identifier',clientWidth:100,scrollWidth:100,getAttribute:name=>name==='role'?'region':'public-region',style:{overflowX:'auto',whiteSpace:'nowrap',textOverflow:'clip',overscrollBehaviorX:policy}});
    const table=element(tablePolicy),identifier=element(identifierPolicy);
    const root={querySelectorAll:selector=>selector==='.table-wrap'?[table]:selector==='.technical-value'?[identifier]:[]};
    const sandbox=vm.createContext({document:{querySelector:()=>null,getElementById:()=>root,documentElement:{scrollWidth:390}},innerWidth:390,getComputedStyle:element=>element.style});
    const page={evaluate:async callback=>vm.runInContext('('+callback.toString()+')()',sandbox),locator:selector=>selector==='dialog[open]'?{count:async()=>0}:{locator:()=>({count:async()=>0})}};
    if(tablePolicy==='none'&&identifierPolicy==='none')assert.equal((await readabilityEvidence(page)).horizontal_boundaries_isolated,true);
    else {
      let failure;
      try {await readabilityEvidence(page);}catch(error){failure=error;}
      assert.ok(failure);
      assert.equal(readabilityFailureFacts(failure).horizontal_boundaries_isolated,false);
      assert.equal(readabilityFailureFacts(failure).boundary,'computed-layout');
    }
  }
});

test('failed native layout evidence retains bounded facts, never arbitrary page text or paths',async()=>{
  const secret='never-retain-page-values';
  const facts={root_fits:false,tables_accessible:true,headers_unbroken:true,identifiers_complete:true,actions_unbroken:true,horizontal_boundaries_isolated:true,tables:1,identifiers:2,overflowing_tables:1,overflowing_identifiers:1,secret,url:secret};
  let error;
  try {await readabilityEvidence({evaluate:async()=>facts});}catch(value){error=value;}
  assert.ok(error);const receipt=readabilityFailureFacts(error);
  assert.equal(receipt.root_fits,false);assert.equal(receipt.boundary,'computed-layout');assert.equal(receipt.tables,1);
  assert.equal(JSON.stringify(receipt).includes(secret),false);assert.deepEqual(readabilityFailureFacts(new Error(secret)),{available:false});
});

test('scroll driver measures the open dialog instead of trying to focus inert background tables',async()=>{
  for(const open of [false,true]) {
    const selected=[];
    const region={count:async()=>0};
    const surface=name=>({count:async()=>name==='dialog'?Number(open):1,locator:selector=>{selected.push([name,selector]);return region;}});
    const facts={root_fits:true,tables_accessible:true,headers_unbroken:true,identifiers_complete:true,actions_unbroken:true,horizontal_boundaries_isolated:true,tables:0,identifiers:0,overflowing_tables:0,overflowing_identifiers:0};
    const page={evaluate:async()=>facts,locator:selector=>{assert.ok(['dialog[open]','#content'].includes(selector));return surface(selector==='dialog[open]'?'dialog':'page');}};
    const result=await readabilityEvidence(page);
    assert.equal(result.keyboard_scrolled_tables,0);assert.equal(result.keyboard_scrolled_identifiers,0);
    assert.deepEqual(selected,[[open?'dialog':'page','.table-wrap'],[open?'dialog':'page','.technical-value']]);
  }
});

test('scroll boundary diagnostic retains only finite geometry and boolean hit facts',async()=>{
  const secret='private-page-text-or-path';
  const measured={connected:true,focused:true,point_in_viewport:false,point_inside_region:true,hit_inside_region:false,hit_technical_value:false,hit_scroll_owner_region:false,hit_nested_scroll_owner:false,pointer_x:100,pointer_y:-85,region_top:-100,scroll_left:12,client_width:200,scroll_width:500,last_column_left:Infinity,last_column_right:secret,hit_owner_scroll_left:NaN,viewport_width:1280,viewport_height:720,text:secret,url:secret,html:secret};
  const facts=await scrollBoundaryFacts({evaluate:async()=>measured},{x:100,y:-85});
  assert.equal(facts.available,true);assert.equal(facts.point_in_viewport,false);
  assert.equal(facts.region_top,-100);assert.equal(facts.scroll_left,12);
  assert.equal(facts.last_column_left,null);assert.equal(facts.last_column_right,null);assert.equal(facts.hit_owner_scroll_left,null);
  assert.equal(JSON.stringify(facts).includes(secret),false);
  assert.deepEqual(await scrollBoundaryFacts({evaluate:async()=>{throw new Error(secret);}},{x:100,y:-85}),{available:false});
});

test('passive event diagnostics are bounded, phase-labelled and removed without driving the page',async()=>{
  const listeners=new Map(),removed=[];
  const documentListeners=new Map(),documentRemoved=[];
  const ownerDocument={
    addEventListener(kind,listener,options){assert.deepEqual({...options},{passive:true,capture:true});documentListeners.set(kind,listener);},
    removeEventListener(kind,listener,options){assert.equal(options.capture,true);assert.equal(documentListeners.get(kind),listener);documentRemoved.push(kind);documentListeners.delete(kind);}
  };
  let clock=0;
  const element={scrollLeft:1,ownerDocument,contains:target=>target===element,
    addEventListener(kind,listener,options){assert.equal(options.passive,true);listeners.set(kind,listener);},
    removeEventListener(kind,listener){assert.equal(listeners.get(kind),listener);removed.push(kind);listeners.delete(kind);}
  };
  const sandbox=vm.createContext({element,performance:{now:()=>clock++},queueMicrotask,scrollX:0,scrollY:148});
  const handle={evaluate:async callback=>callback(handle.value),dispose:async()=>{handle.disposed=true;}};
  const region={evaluateHandle:async callback=>{handle.value=vm.runInContext('('+callback.toString()+')(element)',sandbox);return handle;}};
  const trace=await startScrollEvents(region);
  listeners.get('keydown')({type:'keydown',key:'other',isTrusted:true,target:element});
  listeners.get('scroll')({type:'scroll',isTrusted:true,target:{}});
  listeners.get('keydown')({type:'keydown',key:'ArrowRight',isTrusted:true,defaultPrevented:false,target:element});
  handle.value.wheel();
  const wheel={type:'wheel',isTrusted:true,defaultPrevented:true,deltaX:100000,deltaY:0,deltaMode:0,clientX:753,clientY:112,target:element,secret:'never-retain'};
  handle.value.capture(true);
  assert.equal(listeners.has('wheel'),false);
  documentListeners.get('wheel')(wheel);await Promise.resolve();
  element.scrollLeft=40;
  for(let index=0;index<40;index++)listeners.get('scroll')({type:'scroll',isTrusted:true,defaultPrevented:false,target:element});
  handle.value.capture(false);
  const receipt=await finishScrollEvents(trace);
  assert.equal(receipt.available,true);assert.equal(receipt.events.length,32);assert.equal(receipt.omitted,10);
  assert.equal(receipt.events[0].phase,'keyboard');assert.equal(receipt.events[1].phase,'capture');
  assert.equal(receipt.events[1].kind,'wheel');assert.equal(receipt.events[1].delta_x,100000);
  assert.equal(receipt.events[1].default_prevented,true);assert.equal(receipt.events[2].scroll_left,40);
  assert.equal(receipt.events[1].target_inside_region,true);
  assert.equal(receipt.events[1].client_x,753);assert.equal(receipt.events[1].client_y,112);
  assert.deepEqual(receipt.landmarks.map(value=>value.stage),['wheel-dispatch-start','capture-start','capture-end']);
  assert.equal(receipt.events[1].window_y,148);assert.equal(receipt.landmarks[1].window_y,148);
  assert.ok(receipt.events[1].elapsed_ms>receipt.landmarks[1].elapsed_ms);
  assert.ok(receipt.events[1].elapsed_ms<receipt.landmarks[2].elapsed_ms);
  assert.equal(JSON.stringify(receipt).includes('never-retain'),false);
  assert.equal(listeners.size,0);assert.equal(removed.length,4);assert.equal(handle.disposed,true);
  assert.equal(documentListeners.size,0);assert.deepEqual(documentRemoved,['wheel']);
});

test('document capture observes wheel outside the region without recording target identity',async()=>{
  const listeners=new Map();
  const outside={textContent:'private-outside-target',id:'private-outside-id'};
  const ownerDocument={
    addEventListener(kind,listener,options){assert.equal(options.passive,true);assert.equal(options.capture,true);listeners.set(kind,listener);},
    removeEventListener(kind,listener,options){assert.equal(options.capture,true);assert.equal(listeners.get(kind),listener);listeners.delete(kind);}
  };
  const element={scrollLeft:32,ownerDocument,contains:target=>target===element,addEventListener(){},removeEventListener(){}};
  const sandbox=vm.createContext({element,performance:{now:()=>10},queueMicrotask,scrollX:0,scrollY:148});
  const handle={evaluate:async callback=>callback(handle.value),dispose:async()=>{}};
  const trace=await startScrollEvents({evaluateHandle:async callback=>{handle.value=vm.runInContext('('+callback.toString()+')(element)',sandbox);return handle;}});
  handle.value.wheel();handle.value.capture(true);
  listeners.get('wheel')({type:'wheel',target:outside,clientX:500,clientY:120,deltaX:100000,deltaY:0,deltaMode:0,isTrusted:true,defaultPrevented:false});
  await Promise.resolve();
  const receipt=await finishScrollEvents(trace);
  assert.equal(receipt.events.length,1);assert.equal(receipt.events[0].target_inside_region,false);
  assert.equal(receipt.events[0].phase,'capture');assert.equal(receipt.events[0].client_x,500);
  assert.equal(JSON.stringify(receipt).includes('private-outside'),false);assert.equal(listeners.size,0);
});

test('event receipts whitelist fields and dispose unavailable diagnostics',async()=>{
  const secret='private-diagnostic-payload';
  let disposed=0;
  const trace={evaluate:async()=>({events:[{kind:secret,phase:secret,elapsed_ms:Infinity,scroll_left:40,window_y:secret,delta_x:secret,client_x:secret,client_y:Infinity,target_inside_region:secret,trusted:true,default_prevented:false,url:secret,text:secret}],landmarks:Array(10).fill({stage:secret,elapsed_ms:Infinity,scroll_left:40,window_x:0,window_y:secret,secret}),omitted:-1,secret}),dispose:async()=>{disposed++;}};
  const result=await finishScrollEvents(trace);
  assert.equal(result.events[0].client_x,null);assert.equal(result.events[0].client_y,null);assert.equal(result.events[0].target_inside_region,null);
  assert.equal(JSON.stringify(result).includes(secret),false);assert.equal(result.events[0].kind,null);
  assert.equal(result.events[0].elapsed_ms,null);assert.equal(result.events[0].scroll_left,40);assert.equal(result.omitted,null);
  assert.equal(result.landmarks.length,3);assert.equal(result.landmarks[0].stage,null);assert.equal(result.landmarks[0].window_y,null);
  assert.deepEqual(await finishScrollEvents({...trace,evaluate:async()=>{throw new Error(secret);}}),{available:false});
  assert.deepEqual(await finishScrollEvents(null),{available:false});assert.equal(disposed,2);
  assert.equal(await startScrollEvents({evaluateHandle:async()=>{throw new Error(secret);}}),null);
});

test('keyboard completion requires own-region movement then scrollend, with bounded timeout and cleanup',async()=>{
  for(const ending of ['scrollend','timeout','cancel']) {
    const listeners=new Map();let deadline,cleared=false;
    const element={scrollLeft:0,addEventListener(kind,listener,options){assert.equal(options.passive,true);listeners.set(kind,listener);},removeEventListener(kind){listeners.delete(kind);}};
    const sandbox=vm.createContext({element,setTimeout(callback,timeout){assert.equal(timeout,2000);deadline=callback;return 1;},clearTimeout(){cleared=true;}});
    const handle=await startKeyboardScrollCompletion({evaluateHandle:async callback=>vm.runInContext('('+callback.toString()+')(element)',sandbox)});
    let resolved=false;handle.completed.then(()=>{resolved=true;});
    listeners.get('scrollend')({target:element});await Promise.resolve();assert.equal(resolved,false);
    element.scrollLeft=40;
    listeners.get('scroll')({target:{}});listeners.get('scrollend')({target:element});await Promise.resolve();assert.equal(resolved,false);
    listeners.get('scroll')({target:element});listeners.get('scrollend')({target:{}});await Promise.resolve();assert.equal(resolved,false);
    if(ending==='scrollend')listeners.get('scrollend')({target:element});
    else if(ending==='timeout')deadline();
    else handle.cancel();
    assert.equal(await handle.completed,ending==='scrollend');assert.equal(listeners.size,0);assert.equal(cleared,true);
  }
});

test('wheel completion requires trusted own-region delivery, not preexisting edge geometry',async()=>{
  for(const ending of ['delivered','prevented','timeout','cancel']) {
    const listeners=new Map();let deadline,cleared=false;
    const eventDocument={addEventListener(kind,listener,options){assert.equal(options.passive,true);assert.equal(options.capture,true);listeners.set(kind,listener);},removeEventListener(kind,listener,options){assert.equal(options.capture,true);assert.equal(listeners.get(kind),listener);listeners.delete(kind);}};
    const element={ownerDocument:eventDocument,contains(target){return target===element;},scrollLeft:33,clientWidth:988,scrollWidth:1021};
    const sandbox=vm.createContext({element,queueMicrotask,setTimeout(callback,timeout){assert.equal(timeout,2000);deadline=callback;return 1;},clearTimeout(){cleared=true;}});
    const completion=await startWheelDispatchCompletion({evaluateHandle:async callback=>vm.runInContext('('+callback.toString()+')(element)',sandbox)});
    let resolved=false;completion.completed.then(()=>{resolved=true;});
    const wheel={isTrusted:true,target:element,deltaX:100000,deltaY:0,deltaMode:0,defaultPrevented:false};
    for(const mismatch of [{isTrusted:false},{target:{}},{deltaX:1},{deltaY:1},{deltaMode:1}]) {
      listeners.get('wheel')({...wheel,...mismatch});await Promise.resolve();assert.equal(resolved,false);
    }
    if(['delivered','prevented'].includes(ending)) {
      const event={...wheel};listeners.get('wheel')(event);
      assert.equal(resolved,false);
      event.defaultPrevented=ending==='prevented';
    } else if(ending==='timeout')deadline();
    else completion.cancel();
    assert.equal(await completion.completed,ending==='delivered');assert.equal(listeners.size,0);assert.equal(cleared,true);
  }
});

test('one wheel must complete before capture even when keyboard already reached the edge',async()=>{
  for(const ending of ['delivered','timeout','dispatch-error']) {
    const operations=[];let resolveWheel,resolveDispatch;
    const completion=new Promise(resolve=>{resolveWheel=resolve;});
    const dispatched=new Promise(resolve=>{resolveDispatch=resolve;});
    const original=new Error('private-dispatch-error');
    const handles=[
      {evaluate:async callback=>callback({wheel(){},capture(){},finish(){return {events:[],landmarks:[],omitted:0};}}),dispose:async()=>{}},
      {evaluate:async()=>true,dispose:async()=>{}},
      {evaluate:async callback=>callback({completed:completion,cancel(){operations.push('cancel');resolveWheel(false);}}),dispose:async()=>operations.push('dispose')}
    ];
    const region={evaluateHandle:async()=>handles.shift(),evaluate:async(callback,point)=>point?{scroll_left:33,client_width:988,scroll_width:1021}:true,focus:async()=>{},boundingBox:async()=>({x:258,y:82,width:990,height:818}),elementHandle:async()=>region};
    let selectors=0;
    const surface={locator:()=>({count:async()=>selectors++===0?1:0,nth:()=>region})};
    const page={evaluate:async()=>({root_fits:true,tables_accessible:true,headers_unbroken:true,identifiers_complete:true,actions_unbroken:true,horizontal_boundaries_isolated:true,tables:1,identifiers:0,overflowing_tables:1,overflowing_identifiers:0}),locator:selector=>selector==='dialog[open]'?{count:async()=>0}:surface,keyboard:{press:async()=>operations.push('ArrowRight')},mouse:{move:async()=>{},wheel:async()=>{operations.push('wheel');resolveDispatch();if(ending==='dispatch-error')throw original;}},waitForFunction:async()=>{}};
    const outcome=readabilityEvidence(page,async()=>operations.push('capture')).then(value=>({value}),error=>({error}));
    await dispatched;await Promise.resolve();assert.equal(operations.includes('capture'),false);
    if(ending!=='dispatch-error')resolveWheel(ending==='delivered');
    const result=await outcome;
    if(ending==='delivered') {
      assert.equal(result.value.keyboard_scrolled_tables,1);assert.deepEqual(operations,['ArrowRight','wheel','dispose','capture']);
    } else {
      assert.ok(result.error);if(ending==='dispatch-error')assert.equal(result.error,original);
      assert.equal(readabilityFailureFacts(result.error).boundary,'native-wheel-dispatch');
      assert.deepEqual(operations,['ArrowRight','wheel','cancel','dispose']);
      assert.equal(JSON.stringify(readabilityFailureFacts(result.error)).includes('private-'),false);
    }
  }
});

test('rightmost timeout preserves original assertion and one native input sequence with bounded before/after evidence',async()=>{
  for(const unavailableAfter of [false,true]) {
    const original=new Error('private-error-message');
    const operations=[];
    let waits=0,measurements=0;
    let finished=0,disposed=0;
    const eventHandle={evaluate:async callback=>callback({wheel(){},finish(){finished++;return {events:[],omitted:0};}}),dispose:async()=>{disposed++;}};
    const handles=[eventHandle,{evaluate:async()=>{operations.push('keyboard-scrollend');return true;},dispose:async()=>{}},{evaluate:async()=>true,dispose:async()=>{}}];
    const region={
      evaluateHandle:async()=>handles.shift(),
      evaluate:async(callback,point)=>{
        if(!point)return true;
        measurements++;
        if(unavailableAfter&&measurements===2)throw new Error('private-diagnostic-error');
        return {focused:true,point_in_viewport:false,hit_inside_region:false,scroll_left:12,client_width:200,scroll_width:500,private:'private-page-data'};
      },
      focus:async()=>operations.push('focus'),
      boundingBox:async()=>({x:0,y:-100,width:200,height:30}),
      elementHandle:async()=>region
    };
    const tables={count:async()=>1,nth:()=>region};
    const facts={root_fits:true,tables_accessible:true,headers_unbroken:true,identifiers_complete:true,actions_unbroken:true,horizontal_boundaries_isolated:true,tables:1,identifiers:0,overflowing_tables:1,overflowing_identifiers:0};
    const page={
      evaluate:async()=>facts,
      locator:selector=>selector==='dialog[open]'?{count:async()=>0}:{locator:()=>tables},
      keyboard:{press:async key=>operations.push(key)},
      mouse:{move:async(x,y)=>operations.push(['move',x,y]),wheel:async(x,y)=>operations.push(['wheel',x,y])},
      waitForFunction:async()=>{waits++;if(waits===2)throw original;}
    };
    await assert.rejects(readabilityEvidence(page),error=>error===original);
    assert.deepEqual(operations,['focus','ArrowRight','keyboard-scrollend',['move',100,-85],['wheel',100000,0]]);
    assert.equal(waits,2);assert.equal(measurements,2);
    const receipt=readabilityFailureFacts(original);
    assert.equal(receipt.boundary,'rightmost-column');assert.equal(receipt.keyboard_scrolled_tables,0);
    assert.equal(receipt.scroll_probe.region_index,0);assert.equal(receipt.scroll_probe.before.scroll_left,12);
    assert.equal(receipt.scroll_probe.after.available,!unavailableAfter);
    assert.equal(receipt.scroll_probe.events.available,true);assert.equal(finished,1);assert.equal(disposed,1);
    assert.equal(JSON.stringify(receipt).includes('private-'),false);
  }
});

test('wheel waits for keyboard completion and does not run after a missing completion',async()=>{
  for(const completed of [true,false]) {
    const operations=[];let resolveCompletion,resolveKeyboard;
    const completion=new Promise(resolve=>{resolveCompletion=resolve;});
    const keyboardStarted=new Promise(resolve=>{resolveKeyboard=resolve;});
    const handles=[
      {evaluate:async callback=>callback({wheel(){operations.push('wheel-phase');},finish(){return {events:[],omitted:0};}}),dispose:async()=>{}},
      {evaluate:async callback=>callback({completed:completion,cancel(){operations.push('cancel');}}),dispose:async()=>operations.push('dispose')},
      {evaluate:async()=>true,dispose:async()=>{}}
    ];
    const region={evaluateHandle:async()=>handles.shift(),evaluate:async()=>true,focus:async()=>{},boundingBox:async()=>({x:0,y:0,width:200,height:30}),elementHandle:async()=>region};
    let selectorCount=0;
    const surface={locator:()=>({count:async()=>selectorCount++===0?1:0,nth:()=>region})};
    const page={evaluate:async()=>({root_fits:true,tables_accessible:true,headers_unbroken:true,identifiers_complete:true,actions_unbroken:true,horizontal_boundaries_isolated:true,tables:1,identifiers:0,overflowing_tables:1,overflowing_identifiers:0}),locator:selector=>selector==='dialog[open]'?{count:async()=>0}:surface,keyboard:{press:async()=>{operations.push('ArrowRight');resolveKeyboard();}},mouse:{move:async()=>operations.push('move'),wheel:async()=>operations.push('wheel')},waitForFunction:async()=>{}};
    const outcome=readabilityEvidence(page).then(value=>({value}),error=>({error}));
    await keyboardStarted;
    await Promise.resolve();assert.equal(operations.includes('wheel'),false);
    resolveCompletion(completed);
    const result=await outcome;
    if(completed) {
      assert.equal(result.value.keyboard_scrolled_tables,1);
      assert.deepEqual(operations,['ArrowRight','dispose','move','wheel-phase','wheel']);
    } else {
      assert.ok(result.error);assert.equal(readabilityFailureFacts(result.error).boundary,'native-keyboard-scrollend');
      assert.deepEqual(readabilityFailureFacts(result.error).keyboard_events,{available:true,events:[],landmarks:[],omitted:0});
      assert.deepEqual(operations,['ArrowRight','cancel','dispose']);
    }
  }
});

test('original right-column capture failure retains events observed through capture and disposes once',async()=>{
  for(const observerUnavailable of [false,true]) {
    const original=new Error('private-original-capture-failure');
    const operations=[],events=[],landmarks=[];
    let finished=0,disposed=0,captures=0;
    const eventHandle={evaluate:async callback=>{
      if(observerUnavailable&&finished)throw new Error('private-observer-failure');
      return callback({wheel(){operations.push('wheel-phase');},capture(started){operations.push(started?'capture-start':'capture-end');landmarks.push({stage:started?'capture-start':'capture-end',elapsed_ms:landmarks.length,scroll_left:35,window_x:0,window_y:148});},finish(){finished++;if(observerUnavailable)throw new Error('private-observer-failure');return {events,landmarks,omitted:0};}});
    },dispose:async()=>{disposed++;}};
    const handles=[eventHandle,{evaluate:async()=>true,dispose:async()=>{}},{evaluate:async()=>true,dispose:async()=>{}}];
    const region={evaluateHandle:async()=>handles.shift(),evaluate:async(callback,point)=>point?{scroll_left:35,client_width:988,scroll_width:1023}:true,focus:async()=>{},boundingBox:async()=>({x:258,y:82,width:990,height:818}),elementHandle:async()=>region};
    const page={evaluate:async()=>({root_fits:true,tables_accessible:true,headers_unbroken:true,identifiers_complete:true,actions_unbroken:true,horizontal_boundaries_isolated:true,tables:1,identifiers:0,overflowing_tables:1,overflowing_identifiers:0}),locator:selector=>selector==='dialog[open]'?{count:async()=>0}:{locator:()=>({count:async()=>1,nth:()=>region})},keyboard:{press:async()=>operations.push('ArrowRight')},mouse:{move:async()=>{},wheel:async()=>operations.push('wheel-dispatch')},waitForFunction:async()=>{}};
    await assert.rejects(readabilityEvidence(page,async()=>{
      captures++;operations.push('capture');
      events.push({kind:'wheel',phase:'capture',elapsed_ms:1,scroll_left:35,window_x:0,window_y:148,trusted:true,default_prevented:false,delta_x:100000,delta_y:0,delta_mode:0});
      throw original;
    }),error=>error===original);
    const receipt=readabilityFailureFacts(original);
    assert.equal(receipt.boundary,'rightmost-column');
    assert.equal(receipt.scroll_probe.events.available,!observerUnavailable);
    if(!observerUnavailable) {
      assert.equal(receipt.scroll_probe.events.events[0].phase,'capture');
      assert.deepEqual(receipt.scroll_probe.events.landmarks.map(value=>value.stage),['capture-start','capture-end']);
    }
    assert.deepEqual(operations,['ArrowRight','wheel-phase','wheel-dispatch','capture-start','capture','capture-end']);
    assert.equal(finished,1);assert.equal(disposed,1);assert.equal(captures,1);
    assert.equal(JSON.stringify(receipt).includes('private-'),false);
  }
});
