import assert from 'node:assert/strict';

const failures=new WeakMap();
export const readabilityFailureFacts=error=>failures.get(error)??{available:false};

export async function startKeyboardScrollCompletion(region) {
  return region.evaluateHandle(element=>{
    let moved=false,resolveCompletion;
    const completed=new Promise(resolve=>{resolveCompletion=resolve;});
    const cleanup=()=>{
      clearTimeout(deadline);
      element.removeEventListener('scroll',onScroll);
      element.removeEventListener('scrollend',onEnd);
    };
    const onScroll=event=>{if(event.target===element&&element.scrollLeft>0)moved=true;};
    const onEnd=event=>{
      if(event.target!==element||!moved)return;
      cleanup();resolveCompletion(true);
    };
    const deadline=setTimeout(()=>{cleanup();resolveCompletion(false);},2000);
    element.addEventListener('scroll',onScroll,{passive:true});
    element.addEventListener('scrollend',onEnd,{passive:true});
    return {completed,cancel:()=>{cleanup();resolveCompletion(false);}};
  });
}

export async function startWheelDispatchCompletion(region) {
  return region.evaluateHandle(element=>{
    let resolveCompletion;
    const completed=new Promise(resolve=>{resolveCompletion=resolve;});
    const eventDocument=element.ownerDocument;
    const cleanup=()=>{
      clearTimeout(deadline);
      eventDocument.removeEventListener('wheel',onWheel,{capture:true});
    };
    const onWheel=event=>{
      if(!event.isTrusted||!element.contains(event.target)||event.deltaX!==100000||event.deltaY!==0||event.deltaMode!==0)return;
      queueMicrotask(()=>{cleanup();resolveCompletion(!event.defaultPrevented);});
    };
    const deadline=setTimeout(()=>{cleanup();resolveCompletion(false);},2000);
    eventDocument.addEventListener('wheel',onWheel,{passive:true,capture:true});
    return {completed,cancel:()=>{cleanup();resolveCompletion(false);}};
  });
}

export async function startScrollEvents(region) {
  try {
    return await region.evaluateHandle(element=>{
      const started=performance.now();
      const events=[];
      const landmarks=[];
      let phase='keyboard',omitted=0,stopped=false;
      const landmark=stage=>{
        if(landmarks.length<3)landmarks.push({stage,elapsed_ms:performance.now()-started,scroll_left:element.scrollLeft,window_x:globalThis.scrollX,window_y:globalThis.scrollY});
      };
      const record=event=>{
        if(stopped)return;
        if(events.length===32){omitted++;return;}
        events.push({kind:event.type,phase,elapsed_ms:performance.now()-started,
          scroll_left:element.scrollLeft,trusted:event.isTrusted,
          window_x:globalThis.scrollX,window_y:globalThis.scrollY,
          default_prevented:event.defaultPrevented,
          delta_x:event.type==='wheel'?event.deltaX:null,
          delta_y:event.type==='wheel'?event.deltaY:null,
          delta_mode:event.type==='wheel'?event.deltaMode:null,
          target_inside_region:event.type==='wheel'?element.contains(event.target):null,
          client_x:event.type==='wheel'?event.clientX:null,
          client_y:event.type==='wheel'?event.clientY:null});
      };
      const listener=event=>{
        if(['scroll','scrollend'].includes(event.type)&&event.target!==element)return;
        if(['keydown','keyup'].includes(event.type)&&event.key!=='ArrowRight')return;
        if(event.type==='wheel')queueMicrotask(()=>record(event));
        else record(event);
      };
      const kinds=['keydown','keyup','scroll','scrollend'];
      for(const kind of kinds)element.addEventListener(kind,listener,{passive:true});
      const eventDocument=element.ownerDocument;
      eventDocument.addEventListener('wheel',listener,{passive:true,capture:true});
      return {
        wheel(){phase='wheel';landmark('wheel-dispatch-start');},
        capture(started){phase=started?'capture':'after-capture';landmark(started?'capture-start':'capture-end');},
        finish(){
          stopped=true;
          for(const kind of kinds)element.removeEventListener(kind,listener);
          eventDocument.removeEventListener('wheel',listener,{capture:true});
          return {events,omitted,landmarks};
        }
      };
    });
  } catch {return null;}
}

export async function finishScrollEvents(trace) {
  if(!trace)return {available:false};
  try {
    const measured=await trace.evaluate(value=>value.finish());
    const events=measured.events.slice(0,32).map(event=>{
      const safe={};
      safe.kind=['keydown','keyup','wheel','scroll','scrollend'].includes(event.kind)?event.kind:null;
      safe.phase=['keyboard','wheel','capture','after-capture'].includes(event.phase)?event.phase:null;
      for(const field of ['elapsed_ms','scroll_left','window_x','window_y','delta_x','delta_y','delta_mode','client_x','client_y'])safe[field]=typeof event[field]==='number'&&Number.isFinite(event[field])?event[field]:null;
      for(const field of ['trusted','default_prevented','target_inside_region'])safe[field]=typeof event[field]==='boolean'?event[field]:null;
      return safe;
    });
    const landmarks=Array.isArray(measured.landmarks)?Array.from(measured.landmarks.slice(0,3),value=>({
      stage:['wheel-dispatch-start','capture-start','capture-end'].includes(value.stage)?value.stage:null,
      ...Object.fromEntries(['elapsed_ms','scroll_left','window_x','window_y'].map(field=>[field,typeof value[field]==='number'&&Number.isFinite(value[field])?value[field]:null]))
    })):[];
    return {available:true,events,landmarks,omitted:Number.isSafeInteger(measured.omitted)&&measured.omitted>=0?measured.omitted:null};
  } catch {return {available:false};}
  finally {try {await trace.dispose();}catch {}}
}

export async function scrollBoundaryFacts(region,pointer) {
  try {
    const measured=await region.evaluate((element,point)=>{
      const box=element.getBoundingClientRect();
      const hit=document.elementFromPoint(point.x,point.y);
      const last=element.querySelector('thead th:last-child')?.getBoundingClientRect();
      let scrollOwner=hit;
      while(scrollOwner) {
        if(scrollOwner.scrollWidth>scrollOwner.clientWidth&&['auto','scroll'].includes(getComputedStyle(scrollOwner).overflowX))break;
        scrollOwner=scrollOwner.parentElement;
      }
      return {
        connected:element.isConnected,focused:document.activeElement===element,
        point_in_viewport:point.x>=0&&point.x<innerWidth&&point.y>=0&&point.y<innerHeight,
        point_inside_region:point.x>=box.left&&point.x<box.right&&point.y>=box.top&&point.y<box.bottom,
        hit_inside_region:!!hit&&element.contains(hit),hit_technical_value:!!hit?.closest('.technical-value'),
        hit_scroll_owner_region:scrollOwner===element,hit_nested_scroll_owner:!!scrollOwner&&scrollOwner!==element&&element.contains(scrollOwner),
        pointer_x:point.x,pointer_y:point.y,viewport_width:innerWidth,viewport_height:innerHeight,
        region_left:box.left,region_top:box.top,region_right:box.right,region_bottom:box.bottom,
        scroll_left:element.scrollLeft,client_width:element.clientWidth,scroll_width:element.scrollWidth,
        last_column_left:last?.left,last_column_right:last?.right,
        hit_owner_scroll_left:scrollOwner?.scrollLeft,hit_owner_client_width:scrollOwner?.clientWidth,hit_owner_scroll_width:scrollOwner?.scrollWidth
      };
    },pointer);
    const safe={available:true};
    for(const field of ['connected','focused','point_in_viewport','point_inside_region','hit_inside_region','hit_technical_value','hit_scroll_owner_region','hit_nested_scroll_owner'])safe[field]=typeof measured[field]==='boolean'?measured[field]:null;
    for(const field of ['pointer_x','pointer_y','viewport_width','viewport_height','region_left','region_top','region_right','region_bottom','scroll_left','client_width','scroll_width','last_column_left','last_column_right','hit_owner_scroll_left','hit_owner_client_width','hit_owner_scroll_width'])safe[field]=typeof measured[field]==='number'&&Number.isFinite(measured[field])?measured[field]:null;
    return safe;
  } catch {
    return {available:false};
  }
}

export function assertReadability(facts) {
  assert.equal(facts.root_fits,true,'no root overflow');
  assert.equal(facts.tables_accessible,true,'table regions are named and keyboard reachable');
  assert.equal(facts.headers_unbroken,true,'column headings are not split mid-word');
  assert.equal(facts.identifiers_complete,true,'full technical text is accessible, not clipped');
  assert.equal(facts.actions_unbroken,true,'row action names remain intact');
  assert.equal(facts.horizontal_boundaries_isolated,true,'local horizontal scroll boundaries do not chain or overscroll');
}

export async function readabilityEvidence(page,captureRight) {
  const facts=await page.evaluate(()=>{
    const root=document.querySelector('dialog[open]')??document.getElementById('content');
    const tables=[...root.querySelectorAll('.table-wrap')];
    const identifiers=[...root.querySelectorAll('.technical-value')];
    return {
      root_fits:document.documentElement.scrollWidth<=innerWidth,
      tables:tables.length,
      tables_accessible:tables.every(element=>element.tabIndex===0&&element.getAttribute('role')==='region'&&!!element.getAttribute('aria-label')&&getComputedStyle(element).overflowX==='auto'),
      headers_unbroken:[...root.querySelectorAll('th')].every(element=>getComputedStyle(element).whiteSpace==='nowrap'),
      identifiers:identifiers.length,
      identifiers_complete:identifiers.every(element=>element.tabIndex===0&&!!element.textContent&&getComputedStyle(element).whiteSpace==='nowrap'&&getComputedStyle(element).overflowX==='auto'&&getComputedStyle(element).textOverflow!=='ellipsis'),
      actions_unbroken:[...root.querySelectorAll('td .actions button')].every(element=>getComputedStyle(element).whiteSpace==='nowrap'),
      horizontal_boundaries_isolated:[...tables,...identifiers].every(element=>getComputedStyle(element).overscrollBehaviorX==='none'),
      overflowing_tables:tables.filter(element=>element.scrollWidth>element.clientWidth).length,
      overflowing_identifiers:identifiers.filter(element=>element.scrollWidth>element.clientWidth).length
    };
  });
  let keyboardScrolled=0;
  let identifiersScrolled=0;
  let boundary='computed-layout';
  let scrollProbe=null;
  let activeRegion=null;
  let pointer=null;
  let scrollEvents=null;
  let keyboardCompletion=null;
  let wheelCompletion=null;
  try {
  assertReadability(facts);
  const dialog=page.locator('dialog[open]');
  const surface=await dialog.count()?dialog:page.locator('#content');
  const regions=surface.locator('.table-wrap');
  for(let index=0;index<await regions.count();index++) {
    const region=regions.nth(index);
    if(!await region.evaluate(element=>element.scrollWidth>element.clientWidth))continue;
    boundary='native-keyboard-scroll';
    await region.focus();
    scrollEvents=await startScrollEvents(region);
    keyboardCompletion=await startKeyboardScrollCompletion(region);
    await page.keyboard.press('ArrowRight');
    await page.waitForFunction(element=>element.scrollLeft>0,await region.elementHandle(),{timeout:2000});
    assert.ok(await region.evaluate(element=>element.scrollLeft>0),'native keyboard scroll reached hidden columns');
    boundary='native-keyboard-scrollend';
    assert.equal(await keyboardCompletion.evaluate(value=>value.completed),true,'native keyboard scroll completed before wheel input');
    await keyboardCompletion.dispose();keyboardCompletion=null;
    const bounds=await region.boundingBox();assert.ok(bounds);
    boundary='rightmost-column';
    activeRegion=region;
    pointer={x:bounds.x+bounds.width/2,y:bounds.y+Math.min(bounds.height/2,30)};
    scrollProbe={region_index:index,before:await scrollBoundaryFacts(region,pointer)};
    await page.mouse.move(pointer.x,pointer.y);
    if(scrollEvents)try {await scrollEvents.evaluate(value=>value.wheel());}catch {}
    boundary='native-wheel-dispatch';
    wheelCompletion=await startWheelDispatchCompletion(region);
    await page.mouse.wheel(100000,0);
    assert.equal(await wheelCompletion.evaluate(value=>value.completed),true,'native wheel delivered to table before capture');
    await wheelCompletion.dispose();wheelCompletion=null;
    boundary='rightmost-column';
    await page.waitForFunction(element=>element.scrollLeft+element.clientWidth>=element.scrollWidth-1,await region.elementHandle(),{timeout:2000});
    assert.equal(await region.evaluate(element=>{
      const last=element.querySelector('thead th:last-child').getBoundingClientRect(),box=element.getBoundingClientRect();
      return last.right<=box.right+1&&last.left>=box.left;
    }),true,'rightmost column can be reached');
    scrollProbe.after=await scrollBoundaryFacts(region,pointer);
    try {
      if(captureRight) {
        if(scrollEvents)try {await scrollEvents.evaluate(value=>value.capture(true));}catch {}
        try {await captureRight(index);}
        finally {if(scrollEvents)try {await scrollEvents.evaluate(value=>value.capture(false));}catch {}}
      }
    } finally {
      scrollProbe.events=await finishScrollEvents(scrollEvents);scrollEvents=null;
    }
    await region.evaluate(element=>{element.scrollLeft=0;});
    keyboardScrolled++;
    activeRegion=null;pointer=null;scrollProbe=null;
  }
  assert.equal(keyboardScrolled,facts.overflowing_tables);
  boundary='native-identifier-scroll';
  const identifiers=surface.locator('.technical-value');
  for(let index=0;index<await identifiers.count();index++) {
    const identifier=identifiers.nth(index);
    if(!await identifier.evaluate(element=>element.scrollWidth>element.clientWidth))continue;
    await identifier.focus();await page.keyboard.press('ArrowRight');
    await page.waitForFunction(element=>element.scrollLeft>0,await identifier.elementHandle(),{timeout:2000});
    assert.ok(await identifier.evaluate(element=>element.scrollLeft>0),'complete identifier has native keyboard access');
    await identifier.evaluate(element=>{element.scrollLeft=0;});identifiersScrolled++;
  }
  assert.equal(identifiersScrolled,facts.overflowing_identifiers);
  return {...facts,keyboard_scrolled_tables:keyboardScrolled,keyboard_scrolled_identifiers:identifiersScrolled};
  } catch(error) {
    const safe={available:true,boundary,keyboard_scrolled_tables:keyboardScrolled,keyboard_scrolled_identifiers:identifiersScrolled};
    for(const field of ['root_fits','tables_accessible','headers_unbroken','identifiers_complete','actions_unbroken','horizontal_boundaries_isolated'])safe[field]=typeof facts[field]==='boolean'?facts[field]:null;
    for(const field of ['tables','identifiers','overflowing_tables','overflowing_identifiers'])safe[field]=Number.isSafeInteger(facts[field])&&facts[field]>=0?facts[field]:null;
    if(scrollProbe) {
      scrollProbe.after=await scrollBoundaryFacts(activeRegion,pointer);
      if(scrollEvents) {scrollProbe.events=await finishScrollEvents(scrollEvents);scrollEvents=null;}
      safe.scroll_probe=scrollProbe;
    } else if(scrollEvents) {
      safe.keyboard_events=await finishScrollEvents(scrollEvents);scrollEvents=null;
    }
    failures.set(error,safe);throw error;
  } finally {
    if(wheelCompletion) {
      try {await wheelCompletion.evaluate(value=>value.cancel());}catch {}
      try {await wheelCompletion.dispose();}catch {}
    }
    if(keyboardCompletion) {
      try {await keyboardCompletion.evaluate(value=>value.cancel());}catch {}
      try {await keyboardCompletion.dispose();}catch {}
    }
    if(scrollEvents)await finishScrollEvents(scrollEvents);
  }
}
