import assert from 'node:assert/strict';
import {test} from 'node:test';
import {deflateSync} from 'node:zlib';
import {mkdtemp,readFile,rm} from 'node:fs/promises';
import {join} from 'node:path';
import {tmpdir} from 'node:os';
import vm from 'node:vm';
import {EventEmitter} from 'node:events';
import {assertCaptureIntegrity,captureBoundaryFacts,captureGeometryFacts,captureIntegrityScreenshot,capturePixelFacts,captureRenderFacts,decodeCapturePNG,finishCaptureLifecycle,observeCaptureCompositor,observeCaptureFrameTrace,observeCaptureLifecycle} from './capture-integrity.mjs';

const anchors=[{left:20,top:2,width:12,height:8},{left:20,top:14,width:12,height:8}];
const geometry=()=>({window:{width:40},title:anchors[0],content:anchors[1],anchors});
function png(kind='complete') {
  const width=40,height=24,stride=width*3,raw=Buffer.alloc((stride+1)*height,240);
  for(let row=0;row<height;row++) {
    raw[row*(stride+1)]=0;
    for(let column=0;column<width;column++) {
      const ink=kind==='sidebar'?column<6:kind==='complete'&&column>=22&&column<26&&((row>=4&&row<8)||(row>=16&&row<20));
      if(ink)raw.fill(20,row*(stride+1)+1+column*3,row*(stride+1)+1+column*3+3);
    }
  }
  const header=Buffer.alloc(13);header.writeUInt32BE(width);header.writeUInt32BE(height,4);header[8]=8;header[9]=2;
  const chunk=(name,data)=>{const bytes=Buffer.alloc(data.length+12);bytes.writeUInt32BE(data.length);bytes.write(name,4);data.copy(bytes,8);return bytes;};
  return Buffer.concat([Buffer.from([137,80,78,71,13,10,26,10]),chunk('IHDR',header),chunk('IDAT',deflateSync(raw)),chunk('IEND',Buffer.alloc(0))]);
}

test('capture requires pixels in title and main content, not only nonuniform sidebar',()=>{
  const before=geometry();
  assertCaptureIntegrity(before,before,capturePixelFacts(png(),anchors));
  for(const kind of ['blank','sidebar']) {
    const pixels=capturePixelFacts(png(kind),anchors);
    assert.equal(pixels.anchors.every(anchor=>anchor.ink_present),false);
    assert.throws(()=>assertCaptureIntegrity(before,before,pixels),/substantive content/);
  }
});
test('original PNG uniformity distinguishes a whole blank image from sidebar-only ink without accepting either',()=>{
  for(const kind of ['blank','sidebar','complete']) {
    const pixels=capturePixelFacts(png(kind),anchors);
    assert.equal(pixels.surface.uniform_rgb,kind==='blank');
    assert.deepEqual(pixels.surface.first_pixel_rgb,kind==='sidebar'?[20,20,20]:[240,240,240]);
    if(kind!=='complete')assert.throws(()=>assertCaptureIntegrity(geometry(),geometry(),pixels),/substantive content/);
  }
});
test('render diagnostics whitelist ancestors, lifecycle and hit ownership without text or CSS payloads',async()=>{
  const style={display:'block',visibility:'visible',opacity:'1',contentVisibility:'visible',transform:'none',filter:'none',clipPath:'none',backgroundImage:'url(private-secret)'};
  const root={hidden:false,inert:false,isConnected:true,parentElement:null,textContent:'private-secret',getBoundingClientRect:()=>({left:0,top:0,right:40,bottom:24,width:40,height:24}),contains:()=>true};
  root.parentElement=root;
  const globals={innerWidth:40,innerHeight:24,devicePixelRatio:1,getComputedStyle:()=>style,document:{readyState:'complete',visibilityState:'visible',hasFocus:()=>true,fonts:{status:'loaded'},getAnimations:()=>[],querySelector:selector=>selector==='main'?root:null,getElementById:()=>root,body:root,documentElement:root,elementFromPoint:()=>root}};
  const page={evaluate:callback=>vm.runInNewContext(`(${callback})()`,globals)};
  const visible=await captureRenderFacts(page);
  assert.equal(visible.available,true);assert.equal(visible.nodes.length,5);
  assert.equal(visible.nodes[0].ancestors.length,16);assert.equal(visible.nodes[0].ancestors_omitted,true);
  assert.equal(visible.nodes[0].hit_own_tree,true);assert.equal(visible.fonts,'loaded');
  assert.equal(JSON.stringify(visible).includes('private'),false);
  root.hidden=true;root.inert=true;style.visibility='hidden';style.opacity='0';style.contentVisibility='hidden';style.transform='private-transform';
  globals.document.elementFromPoint=()=>null;
  const hidden=await captureRenderFacts(page);
  assert.equal(hidden.nodes[0].hit_own_tree,false);
  assert.equal(hidden.nodes[0].ancestors[0].hidden,true);assert.equal(hidden.nodes[0].ancestors[0].inert,true);
  assert.equal(hidden.nodes[0].ancestors[0].opacity,0);assert.equal(hidden.nodes[0].ancestors[0].visibility,'hidden');
  assert.equal(hidden.nodes[0].ancestors[0].content_visibility,'hidden');assert.equal(hidden.nodes[0].ancestors[0].transformed,true);
  assert.equal(JSON.stringify(hidden).includes('private'),false);
});
test('render observer failure remains unavailable, never serialized as a raw browser error',async()=>{
  assert.deepEqual(await captureRenderFacts({evaluate:async()=>{throw new Error('private-secret');}}),{available:false});
});
function lifecyclePage() {
  const listeners=new Map();
  const target=name=>({addEventListener:(event,listener,options)=>{assert.equal(options.passive,true);listeners.set(`${name}:${event}`,listener);},removeEventListener:event=>listeners.delete(`${name}:${event}`)});
  let visible=true,disposed=0,elapsed=0;
  const element={isConnected:true,textContent:'private-secret',checkVisibility:options=>{assert.equal(options.checkOpacity,true);assert.equal(options.checkVisibilityCSS,true);assert.equal(options.contentVisibilityAuto,true);return visible;}};
  const globals={...target('window'),performance:{now:()=>elapsed++},innerWidth:1280,innerHeight:900,scrollX:0,scrollY:148,visualViewport:{...target('visual'),width:1280,height:900,scale:1},document:{...target('document'),visibilityState:'visible',body:element,querySelector:()=>null,getElementById:()=>element}};
  const page={evaluateHandle:async callback=>{
    const value=vm.runInNewContext(`(${callback})()`,globals);
    return {evaluate:async(callback,argument)=>callback(value,argument),dispose:async()=>{disposed++;}};
  }};
  return {page,globals,listeners,hide:()=>{visible=false;},disposed:()=>disposed};
}
test('single-capture lifecycle distinguishes transient viewport and visibility from stable endpoints',async()=>{
  const fixture=lifecyclePage(),handle=await observeCaptureLifecycle(fixture.page);
  await handle.evaluate(value=>value.sample('screenshot-call'));
  fixture.globals.innerHeight=1149;fixture.globals.visualViewport.height=1149;fixture.hide();
  fixture.listeners.get('window:resize')();fixture.listeners.get('visual:resize')();
  fixture.globals.innerHeight=900;fixture.globals.visualViewport.height=900;
  fixture.listeners.get('window:resize')();
  await handle.evaluate(value=>value.sample('screenshot-return'));
  const facts=await finishCaptureLifecycle(handle);
  assert.equal(facts.available,true);assert.equal(facts.omitted,0);
  assert.equal(facts.samples[0].viewport.height,900);assert.equal(facts.samples[2].viewport.height,1149);
  assert.equal(facts.samples[2].content.visible,false);assert.equal(facts.samples.at(-1).viewport.height,900);
  assert.equal(facts.samples.every(sample=>sample.document_visible),true);
  assert.equal(JSON.stringify(facts).includes('private'),false);
  assert.equal(fixture.listeners.size,0);assert.equal(fixture.disposed(),1);
});
test('lifecycle samples are bounded and observers dispose on success or failure without raw errors',async()=>{
  const fixture=lifecyclePage(),handle=await observeCaptureLifecycle(fixture.page);
  for(let index=0;index<25;index++)fixture.listeners.get('window:scroll')();
  const facts=await finishCaptureLifecycle(handle);
  assert.equal(facts.samples.length,16);assert.equal(facts.omitted,11);
  assert.equal(fixture.listeners.size,0);assert.equal(fixture.disposed(),1);
  assert.equal(await observeCaptureLifecycle({evaluateHandle:async()=>{throw new Error('private-secret');}}),null);
  let disposed=0;
  assert.deepEqual(await finishCaptureLifecycle({evaluate:async()=>{throw new Error('private-secret');},dispose:async()=>{disposed++;}}),{available:false});
  assert.equal(disposed,1);
  const broken=lifecyclePage(),brokenHandle=await observeCaptureLifecycle(broken.page);
  broken.globals.document.body.checkVisibility=()=>{throw new Error('private-style-failure');};
  assert.deepEqual(await finishCaptureLifecycle(brokenHandle),{available:false});
  assert.equal(broken.listeners.size,0);assert.equal(broken.disposed(),1);
  const unavailable=lifecyclePage();unavailable.globals.document.body.checkVisibility=()=>{throw new Error('private-style-failure');};
  assert.equal(await observeCaptureLifecycle(unavailable.page),null);assert.equal(unavailable.listeners.size,0);
});
test('lifecycle observes the original single PNG and cleans up on blank or screenshot failure',async()=>{
  const directory=await mkdtemp(join(tmpdir(),'relay-capture-lifecycle-'));
  try {
    for(const ending of ['complete','blank','error']) {
      const fixture=lifecyclePage(),path=join(directory,`${ending}.png`),original=new Error('original-screenshot-failure');
      let captures=0;
      fixture.page.evaluate=async()=>geometry();
      fixture.page.screenshot=async options=>{
        captures++;assert.deepEqual(options,{path,fullPage:true});
        fixture.globals.innerHeight=1149;fixture.listeners.get('window:resize')();
        fixture.globals.innerHeight=900;fixture.listeners.get('window:resize')();
        if(ending==='error')throw original;
        return png(ending);
      };
      if(ending==='complete')await captureIntegrityScreenshot(fixture.page,path);
      else if(ending==='error')await assert.rejects(captureIntegrityScreenshot(fixture.page,path),error=>error===original);
      else await assert.rejects(captureIntegrityScreenshot(fixture.page,path),/substantive content/);
      const receipt=JSON.parse(await readFile(`${path}.integrity.json`,'utf8'));
      assert.equal(receipt.result,ending==='complete'?'passed':'failed');
      assert.equal(receipt.lifecycle.samples.some(sample=>sample.viewport.height===1149),true);
      assert.equal(receipt.lifecycle.samples.some(sample=>sample.kind==='screenshot-return'),ending!=='error');
      assert.equal(receipt.lifecycle.samples.at(-1).kind,'finished');
      assert.equal(captures,1);assert.equal(fixture.listeners.size,0);assert.equal(fixture.disposed(),1);
    }
  } finally {await rm(directory,{recursive:true,force:true});}
});
function compositorPage({failure=null}={}) {
  const handlers={},commands=[];
  let detachments=0;
  const session={on:(name,callback)=>{handlers[name]=callback;},send:async name=>{
    commands.push(name);
    if(name===failure)throw new Error('private-protocol-error');
    if(name==='Tracing.start')throw new Error('unsupported-test-domain');
    if(name==='Page.getLayoutMetrics')return {cssLayoutViewport:{pageX:0,pageY:148,clientWidth:1280,clientHeight:900,private:'private-secret'},cssVisualViewport:{pageY:148,scale:1},cssContentSize:{x:0,y:0,width:1280,height:1149},private:'private-secret'};
  },detach:async()=>{detachments++;}};
  return {page:{context:()=>({newCDPSession:async()=>session})},handlers,commands,detachments:()=>detachments};
}
function frameTraceSession({failure=null,ending='complete',entries=[]}={}) {
  const session=new EventEmitter(),commands=[];
  session.send=async(name,args)=>{
    commands.push({name,args});
    if(name===failure)throw new Error('private-protocol-error');
    if(name==='Page.getFrameTree')return {frameTree:{frame:{id:'private-frame',url:'private-url'}}};
    if(name==='Tracing.end') {
      session.emit('Tracing.dataCollected',{value:entries});
      if(ending==='closed')session.emit('close');
      else session.emit('Tracing.tracingComplete',{dataLossOccurred:ending==='lost',private:'private-secret'});
    }
  };
  return {session,commands};
}
test('frame trace projects bounded draw-time geometry and correlates page without retaining IDs or raw trace',async()=>{
  const layer={bounds:{width:1280,height:1149,private:'private-secret'},opacity:1,draws_content:1,screen_space_transform:[...Array(16).fill(0),999],layer_quad:[1,2,3,4,5,6,NaN,'private-secret',999],layer_id:123,paint_record:'private-secret'};
  const entries=[{name:'private-secret',args:{private:'private-secret'}},null,
    ...Array.from({length:10},(_,index)=>({name:index===0?'cc::LayerTreeHostImpl':'LayerTreeHostImpl:snapshot',ts:index,id:'0x7',args:{layerTreeId:7,snapshot:{device_viewport_size:{width:1280,height:1149},active_tree:{source_frame_number:9,layers:Array(20).fill(layer)},active_tiles:[{private:'private-secret'}]}}})),
    {name:'SetLayerTreeId',args:{data:{frame:'private-frame',layerTreeId:7}}},
    ...Array(6).fill({name:'clock_sync',ts:42,args:{sync_id:'relay-capture-call',private:'private-secret'}})];
  const fixture=frameTraceSession({entries}),observer=await observeCaptureFrameTrace(fixture.session);
  await observer.mark('private-secret');await observer.mark('relay-capture-call');await observer.mark('relay-capture-return');
  await observer.finish();await observer.finish();
  const facts=observer.facts;
  assert.equal(facts.available,true);assert.equal(facts.complete,true);assert.equal(facts.data_loss,false);assert.equal(facts.page_associated,true);
  assert.equal(facts.snapshot_events,10);assert.equal(facts.snapshots.length,8);assert.equal(facts.markers.length,4);assert.equal(facts.ignored_events,2);
  assert.equal(facts.snapshots[0].page_association,'matched');assert.equal(facts.snapshots[0].layer_count,20);assert.equal(facts.snapshots[0].layers.length,16);assert.equal(facts.snapshots[0].layers_omitted,4);
  assert.deepEqual(facts.snapshots[0].device_viewport,{width:1280,height:1149});
  assert.equal(facts.snapshots[0].layers[0].screen_space_transform.length,16);assert.deepEqual(facts.snapshots[0].layers[0].layer_quad,[1,2,3,4,5,6,null,null]);
  assert.equal(JSON.stringify(facts).includes('private'),false);assert.equal(JSON.stringify(facts).includes('layer_id'),false);
  assert.deepEqual(fixture.commands.map(command=>command.name),['Page.getFrameTree','Tracing.start','Tracing.recordClockSyncMarker','Tracing.recordClockSyncMarker','Tracing.end']);
  assert.deepEqual(fixture.commands[1].args,{transferMode:'ReportEvents',traceConfig:{recordMode:'recordUntilFull',traceBufferSizeInKb:512,enableSampling:false,enableSystrace:false,includedCategories:['disabled-by-default-devtools.timeline.layers','disabled-by-default-devtools.timeline'],excludedCategories:['*','disabled-by-default-devtools.screenshot']}});
  assert.equal(fixture.session.eventNames().length,0);
});
test('trace identity shape reports only fixed booleans and never guesses page association from object IDs',async()=>{
  const entries=[{name:'SetLayerTreeId',args:{data:{frame:'private-frame',layerTreeId:7}}},
    {name:'cc::LayerTreeHostImpl',id:'0x7',id2:{local:'0x7',global:'0x7'},args:{snapshot:{layer_tree_id:7}}},
    {name:'cc::LayerTreeHostImpl',id:'private-secret',id2:{local:7,global:'private-secret'},args:{layerTreeId:'7',snapshot:{layer_tree_id:Infinity}}},
    {name:'cc::LayerTreeHostImpl',args:{layerTreeId:7}}];
  const fixture=frameTraceSession({entries}),observer=await observeCaptureFrameTrace(fixture.session);
  await observer.finish();
  assert.deepEqual(observer.facts.snapshots.map(snapshot=>snapshot.identity_shape),[
    {argument_numeric:false,snapshot_numeric:true,object_hex:true,local_object_hex:true,global_object_hex:true},
    {argument_numeric:false,snapshot_numeric:false,object_hex:false,local_object_hex:false,global_object_hex:false},
    {argument_numeric:true,snapshot_numeric:false,object_hex:false,local_object_hex:false,global_object_hex:false}]);
  assert.deepEqual(observer.facts.snapshots.map(snapshot=>snapshot.page_association),['unattributed','unattributed','matched']);
  assert.equal(JSON.stringify(observer.facts).includes('private'),false);
  assert.equal(JSON.stringify(observer.facts).includes('0x7'),false);
});
test('Chromium object snapshot ID correlates only an unscoped int ID in the same main-frame process',async()=>{
  const snapshot={name:'cc::LayerTreeHostImpl',ph:'O',pid:12345,id:'0x7',args:{snapshot:{active_tree:{layers:[]}}}};
  const cases=[snapshot,{...snapshot,pid:12346},{...snapshot,id:'0x8'},
    {...snapshot,pid:undefined},{...snapshot,pid:'12345'},{...snapshot,pid:Infinity},
    {...snapshot,ph:'N'},{...snapshot,name:'LayerTreeHostImpl:snapshot'},
    {...snapshot,id2:{local:'0x7'}},{...snapshot,id:'0x80000000'},
    {...snapshot,id:'0x20000000000007'},{...snapshot,id:'0x7private'},
    {...snapshot,id:'0x0'},{...snapshot,id:'0X00000007'}];
  for(const [index,entry] of cases.entries()) {
    const fixture=frameTraceSession({ending:index===0?'lost':'complete',entries:[entry,
      {name:'SetLayerTreeId',pid:12345,args:{data:{frame:'private-frame',layerTreeId:7}}},
      {name:'SetLayerTreeId',pid:12346,args:{data:{frame:'private-subframe',layerTreeId:7}}}]});
    const observer=await observeCaptureFrameTrace(fixture.session);await observer.finish();
    assert.equal(observer.facts.snapshots[0].page_association,[0,13].includes(index)?'matched':[1,2].includes(index)?'other':'unattributed');
    assert.equal(observer.facts.data_loss,index===0);
    const serialized=JSON.stringify(observer.facts);
    for(const privateValue of ['12345','12346','0x7','private-frame','private-subframe','objectKey','treeId'])assert.equal(serialized.includes(privateValue),false);
    assert.equal(fixture.session.eventNames().length,0);
  }
});
test('frame trace unavailable, lost, closed or unattributed observations cannot become a render verdict',async()=>{
  for(const failure of ['Page.getFrameTree','Tracing.start','Tracing.end']) {
    const fixture=frameTraceSession({failure}),observer=await observeCaptureFrameTrace(fixture.session);
    await observer.mark('relay-capture-call');await observer.finish();
    assert.equal(observer.facts.available,failure==='Tracing.end');assert.equal(observer.facts.complete,false);
    assert.equal(fixture.commands.filter(command=>command.name==='Tracing.end').length,failure==='Tracing.end'?1:0);
    assert.equal(fixture.session.eventNames().length,0);assert.equal(JSON.stringify(observer.facts).includes('private'),false);
  }
  for(const ending of ['complete','lost','closed']) {
    const entries=[{name:'LayerTreeHostImpl:snapshot',ts:1,args:{snapshot:{active_tree:{layers:[null]}}}}];
    const fixture=frameTraceSession({ending,entries}),observer=await observeCaptureFrameTrace(fixture.session);
    await observer.finish();
    assert.equal(observer.facts.complete,ending!=='closed');assert.equal(observer.facts.data_loss,ending==='closed'?null:ending==='lost');
    assert.equal(observer.facts.page_associated,false);assert.equal(observer.facts.snapshots[0].page_association,'unattributed');
    assert.equal(fixture.session.eventNames().length,0);
  }
});
test('frame tracing preserves sole original screenshot options and failure identity and always cleans up',async()=>{
  const directory=await mkdtemp(join(tmpdir(),'relay-frame-trace-'));
  try {
    for(const ending of ['complete','blank','error']) {
      const fixture=frameTraceSession(),original=new Error('private-screenshot-error');
      let captures=0,detaches=0;
      fixture.session.detach=async()=>{detaches++;};
      const path=join(directory,`${ending}.png`);
      const page={context:()=>({newCDPSession:async()=>fixture.session}),evaluate:async()=>geometry(),screenshot:async options=>{
        captures++;assert.deepEqual(options,{path,fullPage:true});
        if(ending==='error')throw original;
        return png(ending==='blank'?'blank':'complete');
      }};
      if(ending==='complete')await captureIntegrityScreenshot(page,path);
      else if(ending==='error')await assert.rejects(captureIntegrityScreenshot(page,path),error=>error===original);
      else await assert.rejects(captureIntegrityScreenshot(page,path),/substantive content/);
      const raw=await readFile(`${path}.integrity.json`,'utf8'),receipt=JSON.parse(raw);
      assert.equal(receipt.result,ending==='complete'?'passed':'failed');assert.equal(receipt.compositor.frame_trace.complete,true);
      assert.equal(captures,1);assert.equal(detaches,1);assert.equal(raw.includes('private'),false);
      assert.equal(fixture.commands.filter(command=>command.name==='Tracing.end').length,1);
      assert.equal(fixture.commands.filter(command=>command.args?.syncId==='relay-capture-return').length,ending==='error'?0:1);
      assert.equal(fixture.session.listenerCount('Tracing.dataCollected'),0);assert.equal(fixture.session.listenerCount('Tracing.tracingComplete'),0);assert.equal(fixture.session.listenerCount('close'),0);
    }
  } finally {await rm(directory,{recursive:true,force:true});}
});
test('compositor observer bounds layer and paint events and never records protocol payloads',async()=>{
  const fixture=compositorPage(),observer=await observeCaptureCompositor(fixture.page);
  const layer={offsetX:0,offsetY:148,width:1280,height:1149,paintCount:3,drawsContent:true,transform:Array(20).fill(1),layerId:'private-secret',backendNodeId:99};
  for(let index=0;index<12;index++) {
    fixture.handlers['LayerTree.layerTreeDidChange']({layers:Array(20).fill(layer),private:'private-secret'});
    fixture.handlers['LayerTree.layerPainted']({clip:{x:0,y:148,width:1280,height:900,private:'private-secret'},layerId:'private-secret'});
  }
  await observer.finish();await observer.finish();
  assert.equal(observer.facts.available,true);assert.equal(observer.facts.detached,true);
  assert.equal(observer.facts.tree_events,12);assert.equal(observer.facts.paint_events,12);
  assert.equal(observer.facts.trees.length,8);assert.equal(observer.facts.paints.length,8);
  assert.equal(observer.facts.trees[0].layers.length,16);assert.equal(observer.facts.trees[0].layers_omitted,4);
  assert.equal(observer.facts.trees[0].layers[0].transform.length,16);
  assert.equal(observer.facts.before.layout_viewport.pageY,148);assert.equal(observer.facts.after.content_size.height,1149);
  assert.equal(JSON.stringify(observer.facts).includes('private'),false);
  assert.equal(JSON.stringify(observer.facts).includes('backendNodeId'),false);
  assert.deepEqual(fixture.commands,['LayerTree.enable','Page.getLayoutMetrics','Page.getLayoutMetrics']);
  assert.equal(fixture.detachments(),1);
});
test('missing compositor domain or layout is unavailable, not proof of rendering',async()=>{
  for(const failure of ['LayerTree.enable','Page.getLayoutMetrics']) {
    const fixture=compositorPage({failure}),observer=await observeCaptureCompositor(fixture.page);
    await observer.finish();
    assert.equal(observer.facts.available,failure!=='LayerTree.enable');
    if(failure==='Page.getLayoutMetrics')assert.deepEqual(observer.facts.before,{available:false});
    assert.equal(observer.facts.detached,true);assert.equal(fixture.detachments(),1);
    assert.equal(JSON.stringify(observer.facts).includes('private'),false);
  }
});
test('capture detaches compositor on original pixel failure without another screenshot',async()=>{
  const directory=await mkdtemp(join(tmpdir(),'relay-compositor-'));
  try {
    const fixture=compositorPage();let screenshots=0;
    fixture.page.evaluate=async()=>geometry();
    fixture.page.screenshot=async()=>{screenshots++;return png('blank');};
    const path=join(directory,'capture.png');
    await assert.rejects(captureIntegrityScreenshot(fixture.page,path),/substantive content/);
    const receipt=JSON.parse(await readFile(`${path}.integrity.json`,'utf8'));
    assert.equal(receipt.failure_boundary,'integrity');assert.equal(screenshots,1);
    assert.equal(receipt.compositor.detached,true);assert.equal(fixture.detachments(),1);
    assert.equal(receipt.compositor.tree_events,0);
    assert.equal(receipt.pixels.surface.uniform_rgb,true);
  } finally {await rm(directory,{recursive:true,force:true});}
});
test('capture rejects displaced content and geometry movement without retry',()=>{
  const before=geometry(),pixels=capturePixelFacts(png(),anchors);
  assert.throws(()=>assertCaptureIntegrity(before,{...before,content:{...before.content,left:-500}},pixels),/geometry stable/);
  const moved=[anchors[0],{...anchors[1],left:-100}];
  assert.throws(()=>assertCaptureIntegrity({...before,anchors:moved},before,capturePixelFacts(png(),moved)),/substantive content/);
  assert.throws(()=>assertCaptureIntegrity({...before,anchors:[anchors[0],null]},before,pixels),/anchors exist/);
  assert.throws(()=>assertCaptureIntegrity(before,before,{...pixels,width:39}),/width matches/);
});
test('scroll-correlated modal displacement stays a strict failure with bounded coordinate diagnosis',()=>{
  const before={...geometry(),window:{width:40,x:0,y:182},dialog:true,private:'never-export'};
  const after={...before,window:{...before.window,y:327},title:{...before.title,top:before.title.top+145},content:{...before.content,top:before.content.top+145}};
  const facts=captureGeometryFacts(before,after);
  assert.deepEqual(facts.scroll_delta,{x:0,y:145});
  assert.equal(facts.fields.find(field=>field.field==='title').document_stable,false);
  assert.equal(facts.fields.find(field=>field.field==='title').viewport_stable,true);
  assert.equal(facts.fields.find(field=>field.field==='root').document_stable,false);
  assert.equal(JSON.stringify(facts).includes('never-export'),false);
  assert.throws(()=>assertCaptureIntegrity(before,after,capturePixelFacts(png(),anchors)),/geometry stable/);
  const displaced={...after,content:{...after.content,left:-500}};
  assert.equal(captureGeometryFacts(before,displaced).fields.find(field=>field.field==='content').viewport_stable,false);
  assert.equal(captureGeometryFacts({window:{x:Infinity}},{}).scroll_delta.x,null);
});
test('live boundary whitelist captures viewport, root, scroll owners and table ink anchors without text',async()=>{
  const element=box=>({textContent:'private-value-never-in-receipt',scrollLeft:700,scrollTop:5,clientWidth:40,clientHeight:24,scrollWidth:1000,scrollHeight:24,closest:()=>null,getBoundingClientRect:()=>({...box,right:box.left+box.width,bottom:box.top+box.height})});
  const title=element(anchors[0]),heading=element({...anchors[1],left:4,width:32}),cell=element({...anchors[1],left:4,width:32});
  const table=element({left:0,top:14,width:40,height:8});
  table.querySelectorAll=selector=>selector==='thead th'?[heading]:[cell];
  const surface=element({left:0,top:14,width:40,height:10});
  surface.querySelectorAll=selector=>selector==='.table-wrap'?[table]:[title,heading];
  const root=element({left:0,top:0,width:40,height:24});
  const globals={scrollX:0,scrollY:0,innerWidth:40,innerHeight:24,visualViewport:{offsetLeft:0,offsetTop:0,pageLeft:0,pageTop:0,width:40,height:24,scale:1},document:{documentElement:root,body:root,querySelector:selector=>selector==='main'?root:null,getElementById:id=>id==='content'?surface:title}};
  const facts=await captureBoundaryFacts({evaluate:callback=>vm.runInNewContext(`(${callback})()`,globals)});
  assert.equal(facts.tables[0].scroll_left,700);
  assert.equal(facts.anchors.length,4);assert.equal(facts.anchors[1].top,14);
  assert.equal(facts.visual_viewport.width,40);
  assert.equal(JSON.stringify(facts).includes('private-value'),false);
  const pixels=capturePixelFacts(png(),facts.anchors);
  assertCaptureIntegrity(facts,facts,pixels);
  const missingBody={...pixels,anchors:pixels.anchors.map((anchor,index)=>index===3?{...anchor,ink_present:false}:anchor)};
  assert.throws(()=>assertCaptureIntegrity(facts,facts,missingBody),/substantive content/);
});
function clippedTableBoundary({hiddenLeft=4,visible=true,scrollY=0,clientLeft=0}={}) {
  const element=box=>({textContent:'private-clipped-label',scrollLeft:16,scrollTop:0,clientLeft:0,clientWidth:box.width,clientHeight:box.height,scrollWidth:100,scrollHeight:box.height,closest:()=>null,getBoundingClientRect:()=>({...box,right:box.left+box.width,bottom:box.top+box.height})});
  const title=element(anchors[0]),hidden=element({left:hiddenLeft,top:14,width:24,height:8}),shown=element({...anchors[1],left:12,width:24});
  const table=element({left:10,top:14,width:30,height:8});
  table.clientLeft=clientLeft;table.clientWidth=30-clientLeft*2;
  hidden.closest=()=>table;shown.closest=()=>table;
  table.querySelectorAll=()=>visible?[hidden,shown]:[hidden];
  const surface=element({left:10,top:14,width:30,height:10});
  surface.querySelectorAll=selector=>selector==='.table-wrap'?[table]:visible?[title,hidden,shown]:[title,hidden];
  const root=element({left:0,top:0,width:40,height:24});
  const globals={scrollX:0,scrollY,innerWidth:40,innerHeight:24,visualViewport:null,document:{documentElement:root,body:root,querySelector:selector=>selector==='main'?surface:null,getElementById:id=>id==='content'?surface:title,activeElement:null}};
  return {evaluate:callback=>vm.runInNewContext(`(${callback})()`,globals)};
}
test('content anchor respects table clipping instead of sampling an occluded header',async()=>{
  for(const hiddenLeft of [4,-8,39]) {
    const facts=await captureBoundaryFacts(clippedTableBoundary({hiddenLeft}));
    assert.equal(facts.anchors[1].left,12);
    assert.equal(facts.anchors[1].width,24);
    assert.equal(facts.anchors.length,4);
    assert.equal(JSON.stringify(facts).includes('private-'),false);
    assertCaptureIntegrity(facts,facts,capturePixelFacts(png(),facts.anchors));
    for(const kind of ['blank','sidebar'])assert.throws(()=>assertCaptureIntegrity(facts,facts,capturePixelFacts(png(kind),facts.anchors)),/substantive content/);
  }
});
test('clipping does not remove the mandatory substantive anchor or misplace document coordinates',async()=>{
  const missing=await captureBoundaryFacts(clippedTableBoundary({visible:false}));
  assert.equal(missing.anchors[1],null);
  assert.throws(()=>assertCaptureIntegrity(missing,missing,capturePixelFacts(png(),missing.anchors.filter(Boolean))),/anchors exist/);
  const scrolled=await captureBoundaryFacts(clippedTableBoundary({scrollY:30}));
  assert.equal(scrolled.anchors[1].left,12);assert.equal(scrolled.anchors[1].top,44);
  assert.equal(scrolled.anchors[2].top,44);assert.equal(scrolled.anchors[3].top,44);
});
test('content anchor excludes a header covered by the scroll container border',async()=>{
  const facts=await captureBoundaryFacts(clippedTableBoundary({hiddenLeft:10,clientLeft:1}));
  assert.equal(facts.anchors[1].left,12);assert.equal(facts.anchors[1].width,24);
  assertCaptureIntegrity(facts,facts,capturePixelFacts(png(),facts.anchors));
});
test('PNG audit rejects unsupported and incomplete bytes',()=>{
  assert.equal(decodeCapturePNG(png()).width,40);
  assert.throws(()=>decodeCapturePNG(Buffer.from('not png')));
  assert.throws(()=>decodeCapturePNG(png().subarray(0,40)));
});
test('failed capture preserves image receipt and original error identity with bounded facts',async()=>{
  const directory=await mkdtemp(join(tmpdir(),'relay-capture-'));
  try {
    let screenshots=0,reads=0;
    const page={evaluate:async()=>{reads++;return geometry();},screenshot:async()=>{screenshots++;return png('sidebar');}};
    const path=join(directory,'capture.png');
    await assert.rejects(captureIntegrityScreenshot(page,path),/substantive content/);
    const receipt=JSON.parse(await readFile(`${path}.integrity.json`,'utf8'));
    assert.equal(receipt.result,'failed');assert.equal(receipt.failure_boundary,'integrity');
    assert.equal(reads,4);assert.equal(screenshots,1);assert.equal(receipt.sha256.length,64);
    const error=new Error('private-message');
    page.screenshot=async()=>{throw error;};
    await assert.rejects(captureIntegrityScreenshot(page,path),value=>value===error);
    const raw=await readFile(`${path}.integrity.json`,'utf8');
    assert.equal(raw.includes('private-message'),false);assert.equal(JSON.parse(raw).failure_boundary,'screenshot');
  } finally {await rm(directory,{recursive:true,force:true});}
});
test('geometry failure adds post-capture sampling but does not accept scroll-correlated modal movement',async()=>{
  const directory=await mkdtemp(join(tmpdir(),'relay-capture-'));
  try {
    let boundaryReads=0,screenshots=0;
    const before={...geometry(),window:{width:40,x:0,y:0},dialog:true};
    const after={...before,window:{...before.window,y:2},title:{...before.title,top:before.title.top+2},content:{...before.content,top:before.content.top+2},anchors:anchors.map(anchor=>({...anchor,top:anchor.top+2}))};
    const page={evaluate:async callback=>callback.toString().includes('const known=')?{available:false}:++boundaryReads===1?before:after,screenshot:async options=>{
      screenshots++;assert.equal(options.fullPage,true);return png();
    }};
    const path=join(directory,'capture.png');
    await assert.rejects(captureIntegrityScreenshot(page,path),error=>{
      assert.match(error.message,/geometry stable/);
      assert.equal(error.captureIntegrity.geometry.fields.find(field=>field.field==='title').viewport_stable,true);
      assert.equal(error.captureIntegrity.geometry.fields.find(field=>field.field==='title').document_stable,false);
      assert.equal(error.captureIntegrity.after_pixels.anchors.length,2);
      return true;
    });
    const receipt=JSON.parse(await readFile(`${path}.integrity.json`,'utf8'));
    assert.equal(receipt.result,'failed');assert.equal(receipt.failure_boundary,'integrity');
    assert.equal(screenshots,1);assert.equal(boundaryReads,2);
    assert.equal(receipt.pixels.anchors.every(anchor=>anchor.ink_present),true);
  } finally {await rm(directory,{recursive:true,force:true});}
});
test('diagnostic observer does not replace the original failure or retry the single PNG',async()=>{
  const directory=await mkdtemp(join(tmpdir(),'relay-capture-'));
  try {
    const sequence=[];
    const page={evaluate:async callback=>{
      if(callback.toString().includes('const known=')) {sequence.push('render');throw new Error('private-observer');}
      sequence.push('geometry');return geometry();
    },screenshot:async()=>{sequence.push('screenshot');return png('blank');}};
    const path=join(directory,'capture.png');
    await assert.rejects(captureIntegrityScreenshot(page,path),/substantive content/);
    assert.deepEqual(sequence,['geometry','render','screenshot','geometry','render']);
    const raw=await readFile(`${path}.integrity.json`,'utf8'),receipt=JSON.parse(raw);
    assert.deepEqual(receipt.render,{before:{available:false},after:{available:false}});
    assert.equal(receipt.result,'failed');assert.equal(receipt.pixels.surface.uniform_rgb,true);
    assert.equal(raw.includes('private-observer'),false);
  } finally {await rm(directory,{recursive:true,force:true});}
});
test('successful capture names persisted sidecar and original PNG digest',async()=>{
  const directory=await mkdtemp(join(tmpdir(),'relay-capture-'));
  try {
    const path=join(directory,'capture.png');
    const result=await captureIntegrityScreenshot({evaluate:async()=>geometry(),screenshot:async()=>png()},path);
    assert.equal(result.integrity_path,`${path}.integrity.json`);
    assert.equal(JSON.parse(await readFile(result.integrity_path,'utf8')).result,'passed');
  } finally {await rm(directory,{recursive:true,force:true});}
});
test('all visual and continuous/self-hosted screenshots use integrity gate after existing secret checks',async()=>{
  const visual=await readFile(new URL('./console-visual.mjs',import.meta.url),'utf8');
  const journeys=await readFile(new URL('./journeys.mjs',import.meta.url),'utf8');
  for(const source of [visual,journeys]) {
    assert.ok(source.includes('captureIntegrityScreenshot(page,'));
    assert.equal(source.includes('await page.screenshot('),false);
  }
  assert.ok(visual.indexOf("assert.equal(clear,true")<visual.indexOf('return captureIntegrityScreenshot(page,path)'));
  assert.ok(journeys.includes('localStorage.length===0&&sessionStorage.length===0'));
});
