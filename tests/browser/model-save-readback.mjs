import assert from 'node:assert/strict';
import {summaryCell,summaryFailureFacts} from './console-summary.mjs';

export function readbackPhase(request,tenant,acknowledged) {
  if(!acknowledged||request.method()!=='GET')return null;
  const address=new URL(request.url());
  if(address.pathname==='/admin/api/tenants')return 'tenants';
  if(address.pathname==='/admin/api/catalog'&&address.searchParams.get('tenant')===tenant&&address.searchParams.get('include_ungranted')==='true')return 'catalog';
  return null;
}

const deferred=()=>{
  let resolve,reject;
  const promise=new Promise((success,failure)=>{resolve=success;reject=failure;});
  promise.catch(()=>{});
  return {promise,resolve,reject};
};

export async function waitForSaveReadback(page,expected) {
  await page.waitForFunction(expected=>{
    const form=document.getElementById('editor-form');
    return form.dataset.operation!==expected.operation||['complete','failed','interrupted'].includes(form.dataset.readback)||document.getElementById('tenant').value!==expected.tenant||document.getElementById('page-title').textContent!==expected.page;
  },expected);
  const result=await page.evaluate(expected=>({
    operation_matches:document.getElementById('editor-form').dataset.operation===expected.operation,
    scope_matches:document.getElementById('tenant').value===expected.tenant&&document.getElementById('page-title').textContent===expected.page,
    committed:document.getElementById('editor-form').dataset.readback==='complete'
  }),expected);
  assert.equal(result.operation_matches,true,'save belongs to the current editor operation');
  assert.equal(result.scope_matches,true,'save belongs to the written scope and page');
  assert.equal(result.committed,true,'acknowledged write must finish its page readback and DOM commit');
}

async function snapshot(page,scopeMatches,saveCompleted) {
  const facts=await page.evaluate(()=>({
    dialog_open:document.getElementById('editor').open,
    content_loading:document.getElementById('content').textContent.includes('正在读取当前范围…'),
    write_notice:document.getElementById('notice').textContent==='写入已确认；正在读取当前事实。',
    save_disabled:document.getElementById('editor-save').disabled,
    readback_pending:document.getElementById('editor-form').dataset.readback==='pending',
    page_models:document.getElementById('page-title').textContent==='模型托管'
  }));
  return {...facts,scope_matches:scopeMatches,save_completed:saveCompleted,utc:new Date().toISOString()};
}

export async function modelSaveReadbackJourney(page,{tenant,otherTenant,alias,width,save,ready,changeScope,record,mark}) {
  const cases=[];
  const exactRow=()=>page.getByRole('row').filter({has:page.getByRole('cell',{name:alias,exact:true})});
  const pattern=/\/admin\/api\/(?:tenants|catalog)(?:\?[^#]*)?$/;
  try {
    for(const viewport of [width,390]) {
      await page.setViewportSize({width:viewport,height:900});
      for(const enabled of [false,true]) {
        mark('model-save-readback-'+(enabled?'enable':'disable'));
        await summaryCell(exactRow(),'状态',enabled?'停用':'启用');
        await exactRow().getByRole('button',{name:'启停',exact:true}).click();
        await page.getByLabel('启用',{exact:true}).setChecked(enabled);
        const path='/admin/api/models/'+encodeURIComponent(alias);
        const barriers=Object.fromEntries(['tenants','catalog'].map(phase=>[phase,{reached:deferred(),release:deferred(),finished:deferred(),claimed:false}]));
        const facts={viewport,enabled,role:'platform_administrator',write_status:null,stages:[],final_dom_matches:false,cleanup:false};
        cases.push(facts);
        let acknowledged=false,saveCompleted=false,closing=false,saving;
        const onResponse=response=>{
          if(new URL(response.url()).pathname===path&&response.request().method()==='PATCH') {
            facts.write_status=response.status();
            acknowledged=response.status()===200;
            facts.write_utc=new Date().toISOString();
          }
        };
        const handler=async route=>{
          const phase=readbackPhase(route.request(),tenant,acknowledged),barrier=barriers[phase];
          if(closing||!barrier||barrier.claimed)return route.continue();
          barrier.claimed=true;
          try {
            const response=await route.fetch();
            barrier.reached.resolve({status:response.status(),utc:new Date().toISOString()});
            await barrier.release.promise;
            await route.fulfill({response});
          } catch(error) {
            facts.read_error_kind=error.name==='TimeoutError'?'timeout':'other';
            barrier.failure=error;barrier.reached.reject(error);
          } finally {barrier.finished.resolve();}
        };
        page.on('response',onResponse);
        await page.route(pattern,handler);
        try {
          const requested=Object.fromEntries(['tenants','catalog'].map(phase=>[phase,page.waitForRequest(request=>readbackPhase(request,tenant,acknowledged)===phase)]));
          for(const pending of Object.values(requested))pending.catch(()=>{});
          saving=save('models/'+encodeURIComponent(alias)).then(value=>{saveCompleted=true;facts.save_utc=new Date().toISOString();return value;});
          saving.catch(()=>{});
          for(const phase of ['tenants','catalog']) {
            await requested[phase];
            const reached=await Promise.race([barriers[phase].reached.promise,saving.then(()=>barriers[phase].reached.promise)]);
            assert.equal(reached.status,200,'real post-write readback status');
            await page.locator('#editor').waitFor({state:'hidden'});
            const measured=await snapshot(page,await page.getByLabel('当前租户').inputValue()===tenant,saveCompleted);
            const stage={phase,read_status:reached.status,read_utc:reached.utc,...measured};
            facts.stages.push(stage);
            await record(cases);
            assert.equal(measured.scope_matches,true,'readback still belongs to the written scope');
            assert.equal(measured.page_models,true,'readback still belongs to the model page');
            assert.equal(measured.save_completed,false,'save cannot finish while its post-write response is held');
            assert.equal(measured.readback_pending,true,'current operation remains pending');
            assert.equal(measured.content_loading,true,'old rows are not current facts');
            assert.equal(measured.save_disabled,true,'readback retains its operation lock');
            await record(cases);
            stage.release_utc=new Date().toISOString();
            barriers[phase].release.resolve();
            await barriers[phase].finished.promise;
          }
          await saving;
          await ready();
          facts.final_assertion_utc=new Date().toISOString();
          try {await summaryCell(exactRow(),'状态',enabled?'启用':'停用');}
          catch(error){facts.final_summary_cell=summaryFailureFacts(error);throw error;}
          facts.final_dom_matches=true;facts.final_utc=new Date().toISOString();
        } finally {
          closing=true;
          for(const barrier of Object.values(barriers))barrier.release.resolve();
          for(const barrier of Object.values(barriers))if(barrier.claimed)await barrier.finished.promise;
          try {await page.unroute(pattern,handler);}finally{page.off('response',onResponse);}
          facts.cleanup=true;
          await record(cases);
          if(saving)await saving;
          for(const barrier of Object.values(barriers))if(barrier.failure)throw barrier.failure;
        }
      }
    }
  } finally {await page.setViewportSize({width,height:900});}
  assert.ok(cases.every(value=>value.stages.every(stage=>!stage.save_completed)),'save must not complete while its same-scope post-write readback is held');
  await modelSaveBoundaryJourney(page,{tenant,otherTenant,alias,width,save,ready,changeScope,record:async boundaries=>record([...cases,...boundaries]),mark});
  return cases;
}

async function modelSaveBoundaryJourney(page,{tenant,otherTenant,alias,width,save,ready,changeScope,record,mark}) {
  const cases=[],exactRow=()=>page.getByRole('row').filter({has:page.getByRole('cell',{name:alias,exact:true})});
  const path='/admin/api/models/'+encodeURIComponent(alias),pattern=/\/admin\/api\/(?:tenants|catalog)(?:\?[^#]*)?$/;
  for(const phase of ['tenants','catalog','scope-late-opposite']) {
    mark('model-save-readback-'+phase);
    const late=phase==='scope-late-opposite',enabled=phase==='catalog';
    await summaryCell(exactRow(),'状态',enabled?'停用':'启用');
    await exactRow().getByRole('button',{name:'启停',exact:true}).click();
    await page.getByLabel('启用',{exact:true}).setChecked(enabled);
    let acknowledged=false,claimed=false,closing=false,saving,routeFailure;
    const reached=deferred(),release=deferred(),finished=deferred();
    const facts={case:phase,viewport:width,enabled,role:'platform_administrator',injected_read_failure:!late,cleanup:false};cases.push(facts);
    const onResponse=response=>{if(facts.write_status==null&&new URL(response.url()).pathname===path&&response.request().method()==='PATCH'){acknowledged=response.status()===200;facts.write_status=response.status();facts.write_utc=new Date().toISOString();}};
    const handler=async route=>{
      if(closing||claimed||readbackPhase(route.request(),tenant,acknowledged)!==(late?'catalog':phase))return route.continue();
      claimed=true;
      try {
        const response=await route.fetch();facts.real_read_status=response.status();facts.read_utc=new Date().toISOString();
        reached.resolve();await release.promise;
        facts.forwarded_read_status=late?response.status():503;
        if(late)await route.fulfill({response});
        else await route.fulfill({status:503,contentType:'application/json',body:JSON.stringify({error:{code:'controlled_readback_failure'}})});
      } catch(error){routeFailure=error;facts.read_error_kind=error.name==='TimeoutError'?'timeout':'other';reached.reject(error);}
      finally{finished.resolve();}
    };
    page.on('response',onResponse);await page.route(pattern,handler);
    try {
      const requested=page.waitForRequest(request=>readbackPhase(request,tenant,acknowledged)===(late?'catalog':phase));requested.catch(()=>{});
      saving=save('models/'+encodeURIComponent(alias)).then(()=>({completed:true}),error=>({completed:false,error}));
      await requested;await reached.promise;assert.equal(facts.real_read_status,200);
      if(late) {
        await changeScope(otherTenant);
        assert.equal(await page.getByLabel('当前租户').inputValue(),otherTenant);
        const first=await saving;assert.equal(first.completed,false);assert.equal(first.error.name,'AssertionError');facts.first_save_rejected=true;
        await changeScope(tenant);await summaryCell(exactRow(),'状态','停用');
        await exactRow().getByRole('button',{name:'启停',exact:true}).click();
        await page.getByLabel('启用',{exact:true}).setChecked(true);
        await save('models/'+encodeURIComponent(alias));facts.second_save_utc=new Date().toISOString();
        await summaryCell(exactRow(),'状态','启用');
        release.resolve();await finished.promise;facts.release_utc=new Date().toISOString();
        await summaryCell(exactRow(),'状态','启用');facts.final_dom_matches=true;
        assert.equal(await page.getByLabel('当前租户').inputValue(),tenant);
      } else {
        release.resolve();await finished.promise;facts.release_utc=new Date().toISOString();
        const failed=await saving;assert.equal(failed.completed,false);assert.equal(failed.error.name,'AssertionError');facts.save_rejected=true;
        await ready();assert.equal(await exactRow().count(),0,'failed refresh cannot present old model rows');
        assert.equal(await page.locator('#content .error').count(),1);
        assert.equal(await page.locator('#notice').textContent(),'写入已确认，但读取失败；请重试读取，不要重复写入。');
        const reads=page.waitForResponse(response=>readbackPhase(response.request(),tenant,true)==='catalog');
        await page.getByRole('button',{name:'重试读取',exact:true}).click();const recovered=await reads;assert.equal(recovered.status(),200);facts.recovery_read_status=recovered.status();await ready();
        await summaryCell(exactRow(),'状态',enabled?'启用':'停用');facts.recovered_dom_matches=true;
      }
      facts.final_utc=new Date().toISOString();
    } finally {
      closing=true;release.resolve();if(claimed)await finished.promise;
      try{await page.unroute(pattern,handler);}finally{page.off('response',onResponse);}
      facts.cleanup=true;await record(cases);if(saving)await saving;if(routeFailure)throw routeFailure;
    }
  }
}
