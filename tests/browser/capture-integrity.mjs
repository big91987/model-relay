import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {writeFile} from 'node:fs/promises';
import {inflateSync} from 'node:zlib';

export function decodeCapturePNG(bytes) {
  assert.ok(bytes.subarray(0,8).equals(Buffer.from([137,80,78,71,13,10,26,10])),'PNG signature');
  let width,height,channels;
  const chunks=[];
  for(let offset=8;offset<bytes.length;) {
    const length=bytes.readUInt32BE(offset),type=bytes.toString('ascii',offset+4,offset+8);
    assert.ok(offset+12+length<=bytes.length,'complete PNG chunk');
    const data=bytes.subarray(offset+8,offset+8+length);
    if(type==='IHDR') {
      width=data.readUInt32BE(0);height=data.readUInt32BE(4);
      assert.equal(data[8],8);assert.ok([2,6].includes(data[9]));assert.equal(data[12],0);
      channels=data[9]===2?3:4;
    }
    if(type==='IDAT')chunks.push(data);
    offset+=length+12;
  }
  assert.ok(width>0&&height>0&&width*height<=40000000,'bounded screenshot dimensions');
  const stride=width*channels,raw=inflateSync(Buffer.concat(chunks),{maxOutputLength:(stride+1)*height});
  assert.equal(raw.length,(stride+1)*height);
  const pixels=Buffer.alloc(stride*height);
  for(let row=0;row<height;row++) {
    const filter=raw[row*(stride+1)];assert.ok(filter<=4);
    for(let column=0;column<stride;column++) {
      const left=column>=channels?pixels[row*stride+column-channels]:0;
      const above=row>0?pixels[(row-1)*stride+column]:0;
      const corner=row>0&&column>=channels?pixels[(row-1)*stride+column-channels]:0;
      const prediction=left+above-corner;
      const distances=[Math.abs(prediction-left),Math.abs(prediction-above),Math.abs(prediction-corner)];
      const paeth=distances[0]<=distances[1]&&distances[0]<=distances[2]?left:distances[1]<=distances[2]?above:corner;
      const predictors=[0,left,above,Math.floor((left+above)/2),paeth];
      pixels[row*stride+column]=(raw[row*(stride+1)+column+1]+predictors[filter])&255;
    }
  }
  return {width,height,channels,pixels};
}

export function capturePixelFacts(bytes,anchors) {
  const image=decodeCapturePNG(bytes);
  const firstRGB=[...image.pixels.subarray(0,3)];
  let uniformRGB=true;
  for(let offset=0;offset<image.pixels.length;offset+=image.channels) {
    if(image.pixels[offset]!==firstRGB[0]||image.pixels[offset+1]!==firstRGB[1]||image.pixels[offset+2]!==firstRGB[2]) {uniformRGB=false;break;}
  }
  return {width:image.width,height:image.height,surface:{uniform_rgb:uniformRGB,first_pixel_rgb:firstRGB},anchors:anchors.map(anchor=>{
    const inside=anchor.width>0&&anchor.height>0&&anchor.left>=-1&&anchor.top>=-1&&anchor.left+anchor.width<=image.width+1&&anchor.top+anchor.height<=image.height+1;
    let minimum=255,maximum=0,dark=0,samples=0;
    if(inside)for(let row=Math.max(0,Math.floor(anchor.top));row<Math.min(image.height,Math.ceil(anchor.top+anchor.height));row++) {
      for(let column=Math.max(0,Math.floor(anchor.left));column<Math.min(image.width,Math.ceil(anchor.left+anchor.width));column++) {
        const offset=(row*image.width+column)*image.channels;
        const light=Math.round((image.pixels[offset]+image.pixels[offset+1]+image.pixels[offset+2])/3);
        minimum=Math.min(minimum,light);maximum=Math.max(maximum,light);dark+=Number(light<180);samples++;
      }
    }
    return {inside,samples,contrast:maximum-minimum,dark_pixels:dark,ink_present:inside&&maximum-minimum>=32&&dark>=8};
  })};
}

export async function captureRenderFacts(page) {
  try {
    return await page.evaluate(()=>{
      const known=(value,values)=>values.includes(value)?value:'other';
      const dialog=document.querySelector('dialog[open]'),surface=dialog??document.getElementById('content');
      const title=dialog?.querySelector('h2')??document.getElementById('page-title');
      const targets=[title,surface,document.querySelector('main'),document.body,document.documentElement];
      const nodes=targets.map(element=>{
        if(!element)return {available:false};
        const box=element.getBoundingClientRect();
        const point={x:Math.max(0,box.left)+Math.min(8,box.width/2),y:Math.max(0,box.top)+Math.min(8,box.height/2)};
        const pointInBox=point.x>=box.left&&point.x<box.right&&point.y>=box.top&&point.y<box.bottom&&point.x<innerWidth&&point.y<innerHeight;
        const hit=pointInBox?document.elementFromPoint(point.x,point.y):null;
        const ancestors=[];
        let current=element;
        for(;current&&ancestors.length<16;current=current.parentElement) {
          const style=getComputedStyle(current),opacity=Number(style.opacity);
          ancestors.push({hidden:current.hidden===true,inert:current.inert===true,display:known(style.display,['none','block','inline','inline-block','flex','grid','table','table-row','table-cell','contents']),visibility:known(style.visibility,['visible','hidden','collapse']),opacity:Number.isFinite(opacity)&&opacity>=0&&opacity<=1?opacity:null,content_visibility:known(style.contentVisibility,['visible','auto','hidden']),transformed:style.transform!=='none',filtered:style.filter!=='none',clipped:style.clipPath!=='none'});
        }
        return {available:true,connected:element.isConnected,point_in_box:pointInBox,hit_own_tree:!!hit&&(hit===element||element.contains(hit)),ancestors,ancestors_omitted:!!current};
      });
      return {available:true,ready_state:known(document.readyState,['loading','interactive','complete']),visibility_state:known(document.visibilityState,['visible','hidden']),focused:document.hasFocus(),fonts:known(document.fonts.status,['loading','loaded']),device_pixel_ratio:devicePixelRatio,animations:document.getAnimations().length,nodes};
    });
  } catch {
    return {available:false};
  }
}

export async function captureBoundaryFacts(page) {
  return page.evaluate(()=>{
    const rectangle=element=>{
      if(!element)return null;
      const box=element.getBoundingClientRect();
      return {left:box.left+scrollX,top:box.top+scrollY,width:box.width,height:box.height,scroll_left:element.scrollLeft,scroll_top:element.scrollTop,client_width:element.clientWidth,client_height:element.clientHeight,scroll_width:element.scrollWidth,scroll_height:element.scrollHeight};
    };
    const dialog=document.querySelector('dialog[open]'),surface=dialog??document.getElementById('content');
    const title=dialog?.querySelector('h2')??document.getElementById('page-title');
    const candidates=[...surface.querySelectorAll('h2,th,label,dt,p')].filter(element=>{
      const box=element.getBoundingClientRect();
      const table=element.closest('.table-wrap');
      const tableBox=table?.getBoundingClientRect();
      const clipLeft=tableBox?tableBox.left+table.clientLeft:0;
      const clipRight=tableBox?clipLeft+table.clientWidth:innerWidth;
      return element!==title&&element.textContent.trim()&&box.width>0&&box.height>0&&box.left+scrollX>=0&&box.right+scrollX<=innerWidth&&box.left>=clipLeft&&box.right<=clipRight;
    });
    const tables=[...surface.querySelectorAll('.table-wrap')];
    const active=document.activeElement;
    const focus={inside_content:!!active&&surface.contains(active),technical_value:!!active?.closest('.technical-value'),table_region:!!active?.closest('.table-wrap'),box:rectangle(active)};
    const anchors=[rectangle(title),rectangle(candidates[0])];
    for(const table of tables.slice(0,8)) {
      const box=table.getBoundingClientRect();
      for(const selector of ['thead th','tbody tr:first-child td']) {
        const visibleCells=[...table.querySelectorAll(selector)].filter(element=>{
          const cellBox=element.getBoundingClientRect();
          return element.textContent.trim()&&Math.min(cellBox.right,box.right,innerWidth)-Math.max(cellBox.left,box.left,0)>=24;
        });
        const cell=visibleCells.find(element=>{
          const cellBox=element.getBoundingClientRect();
          return cellBox.left>=Math.max(box.left,0)-1&&cellBox.right<=Math.min(box.right,innerWidth)+1;
        })??visibleCells.sort((first,second)=>{
          const visibleWidth=element=>{const cellBox=element.getBoundingClientRect();return Math.min(cellBox.right,box.right,innerWidth)-Math.max(cellBox.left,box.left,0);};
          return visibleWidth(second)-visibleWidth(first);
        })[0];
        if(cell) {
          const cellBox=cell.getBoundingClientRect(),left=Math.max(cellBox.left,box.left,0),right=Math.min(cellBox.right,box.right,innerWidth);
          anchors.push({left:left+scrollX,top:cellBox.top+scrollY,width:right-left,height:cellBox.height});
        }
      }
    }
    const position=dialog?getComputedStyle(dialog).position:null;
    const dialogPosition=['fixed','absolute','relative','static','sticky'].includes(position)?position:null;
    return {utc:new Date().toISOString(),window:{x:scrollX,y:scrollY,width:innerWidth,height:innerHeight},visual_viewport:visualViewport?{left:visualViewport.offsetLeft,top:visualViewport.offsetTop,page_left:visualViewport.pageLeft,page_top:visualViewport.pageTop,width:visualViewport.width,height:visualViewport.height,scale:visualViewport.scale}:null,root:rectangle(document.documentElement),body:rectangle(document.body),main:rectangle(document.querySelector('main')),content:rectangle(surface),dialog:!!dialog,dialog_position:dialogPosition,title:rectangle(title),focus,anchors,tables:tables.slice(0,8).map(rectangle),tables_omitted:Math.max(0,tables.length-8)};
  });
}

export function captureGeometryFacts(before,after) {
  const delta=(first,second)=>Number.isFinite(first)&&Number.isFinite(second)&&Number.isFinite(second-first)?second-first:null;
  const scroll={x:delta(before?.window?.x,after?.window?.x),y:delta(before?.window?.y,after?.window?.y)};
  const fields=['root','body','main','title','content'].map(field=>{
    const documentDelta=Object.fromEntries(['left','top','width','height'].map(dimension=>[dimension,delta(before?.[field]?.[dimension],after?.[field]?.[dimension])]));
    const viewportDelta={left:delta(scroll.x,documentDelta.left),top:delta(scroll.y,documentDelta.top)};
    return {field,document_delta:documentDelta,viewport_delta:viewportDelta,document_stable:Object.values(documentDelta).every(value=>value!==null&&Math.abs(value)<=1),viewport_stable:[viewportDelta.left,viewportDelta.top,documentDelta.width,documentDelta.height].every(value=>value!==null&&Math.abs(value)<=1)};
  });
  return {scroll_delta:scroll,dialog_before:before?.dialog===true,dialog_after:after?.dialog===true,fields};
}

export function assertCaptureIntegrity(before,after,pixels) {
  assert.equal(pixels.width,before.window.width,'capture width matches layout viewport');
  assert.ok(before.anchors.every(Boolean),'title and substantive content anchors exist');
  assert.equal(pixels.anchors.length,before.anchors.length,'all title/content/table anchors checked');
  assert.ok(pixels.anchors.length>=2&&pixels.anchors.length<=18,'bounded substantive capture anchors');
  assert.ok(pixels.anchors.every(anchor=>anchor.inside&&anchor.ink_present),'screenshot contains title and substantive content, not blank or sidebar only');
  for(const field of ['title','content'])for(const dimension of ['left','top','width','height']) {
    assert.ok(Math.abs(before[field][dimension]-after[field][dimension])<=1,'document geometry stable across screenshot');
  }
}

export async function observeCaptureFrameTrace(session) {
  const facts={available:false,complete:false,data_loss:null,page_associated:false,snapshot_events:0,snapshots:[],markers:[],ignored_events:0};
  const number=value=>Number.isFinite(value)?value:null;
  const vector=(value,length)=>Array.isArray(value)?value.slice(0,length).map(number):null;
  const size=value=>value?{width:number(value.width),height:number(value.height)}:null;
  let started=false,finished=false,resolveComplete,frameId;
  const pageTrees=new Set(),processTrees=new Set(),pending=[];
  const processKey=(pid,treeId)=>Number.isSafeInteger(pid)&&pid>0&&Number.isSafeInteger(treeId)&&treeId>0&&treeId<=0x7fffffff?`${pid}:${treeId}`:null;
  const completed=new Promise(resolve=>{resolveComplete=resolve;});
  const collect=event=>{
    for(const entry of Array.isArray(event.value)?event.value:[]) {
      if(entry?.name==='SetLayerTreeId') {
        const data=entry.args?.data;
        if(typeof frameId==='string'&&data?.frame===frameId&&Number.isFinite(data.layerTreeId)&&pageTrees.size<8) {
          pageTrees.add(data.layerTreeId);
          const key=processKey(entry.pid,data.layerTreeId);
          if(key!==null&&processTrees.size<8)processTrees.add(key);
        }
      } else if(['cc::LayerTreeHostImpl','LayerTreeHostImpl:snapshot'].includes(entry?.name)) {
        facts.snapshot_events++;
        if(pending.length===8)continue;
        const snapshot=entry.args?.snapshot,tree=snapshot?.active_tree;
        const layers=Array.isArray(tree?.layers)?tree.layers:[];
        const treeId=Number.isFinite(entry.args?.layerTreeId)?entry.args.layerTreeId:null;
        const hexadecimal=value=>typeof value==='string'&&/^0x[0-9a-f]+$/i.test(value);
        const objectId=entry.name==='cc::LayerTreeHostImpl'&&entry.ph==='O'&&entry.id2===undefined&&hexadecimal(entry.id)&&entry.id.length<=10?Number(entry.id):null;
        const objectKey=processKey(entry.pid,objectId);
        const identityShape={argument_numeric:Number.isFinite(entry.args?.layerTreeId),snapshot_numeric:Number.isFinite(snapshot?.layer_tree_id),object_hex:hexadecimal(entry.id),local_object_hex:hexadecimal(entry.id2?.local),global_object_hex:hexadecimal(entry.id2?.global)};
        pending.push({treeId,objectKey,facts:{timestamp_us:number(entry.ts),identity_shape:identityShape,tree_available:!!tree,device_viewport:size(snapshot?.device_viewport_size),source_frame_number:number(tree?.source_frame_number),layer_count:layers.length,layers:layers.slice(0,16).map(layer=>({bounds:size(layer?.bounds),opacity:number(layer?.opacity),draws_content:[0,1].includes(layer?.draws_content)?layer.draws_content:null,screen_space_transform:vector(layer?.screen_space_transform,16),layer_quad:vector(layer?.layer_quad,8)})),layers_omitted:Math.max(0,layers.length-16)}});
      } else if(entry?.name==='clock_sync'&&['relay-capture-call','relay-capture-return'].includes(entry.args?.sync_id)) {
        if(facts.markers.length<4)facts.markers.push({kind:entry.args.sync_id,timestamp_us:number(entry.ts)});
      } else facts.ignored_events++;
    }
  };
  const complete=event=>{facts.complete=true;facts.data_loss=typeof event.dataLossOccurred==='boolean'?event.dataLossOccurred:null;resolveComplete();};
  const close=()=>resolveComplete();
  const remove=()=>{for(const [name,listener] of [['Tracing.dataCollected',collect],['Tracing.tracingComplete',complete],['close',close]])try {session.off(name,listener);}catch {}};
  try {
    const frame=await session.send('Page.getFrameTree');frameId=frame?.frameTree?.frame?.id;
    session.on('Tracing.dataCollected',collect);session.on('Tracing.tracingComplete',complete);session.on('close',close);
    await session.send('Tracing.start',{transferMode:'ReportEvents',traceConfig:{recordMode:'recordUntilFull',traceBufferSizeInKb:512,enableSampling:false,enableSystrace:false,includedCategories:['disabled-by-default-devtools.timeline.layers','disabled-by-default-devtools.timeline'],excludedCategories:['*','disabled-by-default-devtools.screenshot']}});
    started=true;facts.available=true;
  } catch {remove();}
  return {facts,mark:async kind=>{
    if(!started||finished||!['relay-capture-call','relay-capture-return'].includes(kind))return;
    try {await session.send('Tracing.recordClockSyncMarker',{syncId:kind});}catch {}
  },finish:async()=>{
    if(!started||finished)return;
    finished=true;
    try {await session.send('Tracing.end');await completed;}catch {}
    finally {
      facts.page_associated=pageTrees.size>0;
      facts.snapshots=pending.map(entry=>({...entry.facts,page_association:pageTrees.has(entry.treeId)||processTrees.has(entry.objectKey)?'matched':(entry.treeId!==null&&pageTrees.size>0)||(entry.objectKey!==null&&processTrees.size>0)?'other':'unattributed'}));
      pending.length=0;pageTrees.clear();processTrees.clear();frameId=undefined;remove();
    }
  }};
}

export async function observeCaptureCompositor(page,{frameTrace=false}={}) {
  const facts={available:false,before:null,after:null,tree_events:0,paint_events:0,trees:[],paints:[],detached:false};
  const number=value=>Number.isFinite(value)?value:null;
  const rectangle=value=>value?Object.fromEntries(['x','y','width','height'].map(field=>[field,number(value[field])])):null;
  let session,trace;
  const snapshot=async()=>{
    try {
      const layout=await session.send('Page.getLayoutMetrics');
      const viewport=value=>value?Object.fromEntries(['pageX','pageY','clientWidth','clientHeight','offsetX','offsetY','scale','zoom'].map(field=>[field,number(value[field])])):null;
      return {available:true,utc:new Date().toISOString(),layout_viewport:viewport(layout.cssLayoutViewport),visual_viewport:viewport(layout.cssVisualViewport),content_size:rectangle(layout.cssContentSize)};
    } catch {return {available:false};}
  };
  try {
    session=await page.context().newCDPSession(page);
    session.on('LayerTree.layerTreeDidChange',event=>{
      facts.tree_events++;
      if(facts.trees.length>=8)return;
      const layers=Array.isArray(event.layers)?event.layers:[];
      facts.trees.push({utc:new Date().toISOString(),layer_count:layers.length,draws_content:layers.filter(layer=>layer.drawsContent===true).length,layers:layers.slice(0,16).map(layer=>({offset_x:number(layer.offsetX),offset_y:number(layer.offsetY),width:number(layer.width),height:number(layer.height),paint_count:number(layer.paintCount),draws_content:layer.drawsContent===true,transform:Array.isArray(layer.transform)?layer.transform.slice(0,16).map(number):null})),layers_omitted:Math.max(0,layers.length-16)});
    });
    session.on('LayerTree.layerPainted',event=>{
      facts.paint_events++;
      if(facts.paints.length<8)facts.paints.push({utc:new Date().toISOString(),clip:rectangle(event.clip)});
    });
    await session.send('LayerTree.enable');
    facts.available=true;
    facts.before=await snapshot();
    if(frameTrace){trace=await observeCaptureFrameTrace(session);facts.frame_trace=trace.facts;}
  } catch {
    if(session)try {await session.detach();facts.detached=true;}catch {}
    session=null;
  }
  return {facts,mark:async kind=>{if(trace)await trace.mark(kind);},finish:async()=>{
    if(!session)return;
    facts.after=await snapshot();
    if(trace)await trace.finish();
    try {await session.detach();facts.detached=true;}catch {}
    session=null;
  }};
}

export async function observeCaptureLifecycle(page) {
  let handle;
  try {
    handle=await page.evaluateHandle(()=>{
      const started=performance.now(),samples=[];
      let omitted=0,stopped=false;
      const sample=kind=>{
        if(stopped)return;
        if(samples.length===16){omitted++;return;}
        const content=document.querySelector('dialog[open]')??document.getElementById('content');
        const visibility=element=>element?{connected:element.isConnected,visible:element.checkVisibility({checkOpacity:true,checkVisibilityCSS:true,contentVisibilityAuto:true})}:null;
        samples.push({kind,elapsed_ms:performance.now()-started,
          viewport:{width:innerWidth,height:innerHeight,x:scrollX,y:scrollY},
          visual_viewport:visualViewport?{width:visualViewport.width,height:visualViewport.height,scale:visualViewport.scale}:null,
          document_visible:document.visibilityState==='visible',
          body:visibility(document.body),content:visibility(content)});
      };
      sample('armed');
      const bindings=[[globalThis,'resize','window-resize'],[globalThis,'scroll','window-scroll'],[document,'visibilitychange','visibility-change']];
      if(visualViewport)bindings.push([visualViewport,'resize','visual-resize'],[visualViewport,'scroll','visual-scroll']);
      const listeners=bindings.map(([target,event,kind])=>{
        const listener=()=>sample(kind);
        target.addEventListener(event,listener,{passive:true});
        return {target,event,listener};
      });
      return {sample,finish(){
        if(!stopped)try {sample('finished');}
        finally {stopped=true;for(const {target,event,listener} of listeners)target.removeEventListener(event,listener);}
        return {available:true,samples,omitted};
      }};
    });
  } catch {if(handle)try {await handle.dispose();}catch {}return null;}
  return handle;
}

export async function finishCaptureLifecycle(handle) {
  if(!handle)return {available:false};
  try {return await handle.evaluate(value=>value.finish());}
  catch {return {available:false};}
  finally {try {await handle.dispose();}catch {}}
}

export async function captureIntegrityScreenshot(page,path) {
  const compositor=await observeCaptureCompositor(page,{frameTrace:true});
  const lifecycle=await observeCaptureLifecycle(page);
  const markLifecycle=async kind=>{if(lifecycle)try {await lifecycle.evaluate((value,stage)=>value.sample(stage),kind);}catch {}};
  const receipt={mode:'fullPage',result:'in_progress',started_utc:new Date().toISOString(),before:null,after:null,render:{before:null,after:null},compositor:compositor.facts,lifecycle:null,pixels:null};
  let capturedBytes=null;
  let boundary='before';
  try {
    receipt.before=await captureBoundaryFacts(page);
    receipt.render.before=await captureRenderFacts(page);
    boundary='screenshot';
    await compositor.mark('relay-capture-call');
    await markLifecycle('screenshot-call');
    const bytes=await page.screenshot({path,fullPage:true});
    capturedBytes=bytes;
    await markLifecycle('screenshot-return');
    await compositor.mark('relay-capture-return');
    receipt.sha256=createHash('sha256').update(bytes).digest('hex');
    boundary='after';receipt.after=await captureBoundaryFacts(page);
    receipt.render.after=await captureRenderFacts(page);
    receipt.geometry=captureGeometryFacts(receipt.before,receipt.after);
    boundary='pixels';receipt.pixels=capturePixelFacts(bytes,receipt.before.anchors.filter(Boolean));
    boundary='integrity';assertCaptureIntegrity(receipt.before,receipt.after,receipt.pixels);
    receipt.result='passed';
    return {path,sha256:receipt.sha256,integrity_path:`${path}.integrity.json`};
  } catch(error) {
    receipt.result='failed';receipt.failure_boundary=boundary;
    if(boundary==='integrity') {
      try {receipt.after_pixels=capturePixelFacts(capturedBytes,receipt.after.anchors.filter(Boolean));}catch {receipt.after_pixels={available:false};}
      error.captureIntegrity={boundary,geometry:receipt.geometry,pixels:receipt.pixels,after_pixels:receipt.after_pixels};
    }
    if(receipt.before&&!receipt.after)try {receipt.after=await captureBoundaryFacts(page);}catch {}
    if(receipt.before&&!receipt.render.after)receipt.render.after=await captureRenderFacts(page);
    throw error;
  } finally {
    receipt.lifecycle=await finishCaptureLifecycle(lifecycle);
    await compositor.finish();
    receipt.finished_utc=new Date().toISOString();
    await writeFile(`${path}.integrity.json`,JSON.stringify(receipt,null,2)+'\n');
  }
}
