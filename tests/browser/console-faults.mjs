import assert from 'node:assert/strict';
import {writeFile} from 'node:fs/promises';
import {createHash} from 'node:crypto';

export function previousEditorAsset(source) {
  const start=source.indexOf('function edit('),end=source.indexOf('function secretDialog(');
  assert.ok(start>=0&&end>start);
  const editor=source.slice(start,end),catchStart=editor.indexOf('    catch(error) {'),catchEnd=editor.indexOf('    finally {');
  assert.ok(catchStart>=0&&catchEnd>catchStart);
  const historicalCatch="    catch(error) { if(epoch!==state.epoch) return; state.busy=false; clearSensitive(); if(error.name==='AbortError') return; const output=byId('editor-error'); output.textContent=error.status?`操作失败：${error.message}。非敏感草稿仍保留。`:'写入结果未知，请先安全读回；不要重复签发或计费。'; if(!error.status) { byId('editor-save').hidden=true; output.append(button('安全读回',async()=>{try{const result=await api('operations/'+encodeURIComponent(operation)); detail('操作已提交，秘密不可恢复',result); }catch{output.textContent='读回失败，结果仍未知。请保留操作标识并稍后重试读回。';}})); } }\n";
  const previous=(editor.slice(0,catchStart)+historicalCatch+editor.slice(catchEnd)).replace('  let unknown=false;\n','').replace('if(state.busy||unknown)','if(state.busy)');
  return source.slice(0,start)+previous+source.slice(end);
}

export async function consoleFaults(page,{origin,tenant,width,issued,requestID,nav,step=()=>{}}) {
  const facts={suite:'console-faults',width,provider:'controlled_http_not_supplier',cases:[],observations:[]};
  const phase=value=>{facts.phase=value;step(value);};
  const read=async path=>{
    const response=await page.request.get(origin+'/admin/api/'+path);assert.equal(response.status(),200);return response.json();
  };
  const screenshot=async name=>{
    assert.ok(await page.evaluate(()=>[...document.querySelectorAll('.secret,input[type=password]')].every(input=>input.value==='')));
    await page.screenshot({path:`test-results/fault-${width}-${name}.png`,fullPage:true});
  };
  const snapshot=await read('workspace/requests/'+requestID);
  const keysPath=`tenants/${tenant}/keys`;
  const upstreamFacts=async()=>{
    const response=await page.request.get(new URL(process.env.RELAY_CONSOLE_UPSTREAM).origin+'/test/facts',{headers:{Authorization:'Bearer controlled-console-upstream'}});
    assert.equal(response.status(),200);return (await response.json()).dispatches;
  };
  async function historicalRed() {
    phase('historical-readback-red');
    let assetHash,mutations=0;
    const assetRoute='**/admin/workspace.js',writeRoute='**/admin/api/'+keysPath,readRoute='**/admin/api/operations/*';
    await page.route(assetRoute,async route=>{
      const response=await route.fetch(),body=previousEditorAsset(await response.text());
      assetHash=createHash('sha256').update(body).digest('hex');await route.fulfill({response,body});
    });
    await page.route(writeRoute,async route=>{
      if(route.request().method()!=='POST')return route.continue();
      mutations++;const response=await route.fetch();assert.equal(response.status(),201);await route.abort('failed');
    });
    await page.route(readRoute,route=>route.fulfill({status:503,contentType:'application/json',body:'{"error":{"code":"controlled_read_failure"}}'}));
    try {
      await page.reload();await page.locator('#workspace').waitFor({state:'visible'});await page.getByLabel('当前租户').selectOption(tenant);
      await nav('API Keys');await page.getByRole('button',{name:'创建Key',exact:true}).click();
      await page.getByLabel('名称',{exact:true}).fill('Historical readback '+width);await page.getByLabel('应用',{exact:true}).fill('isolated red regression');
      await page.getByLabel('模型授权（可多选）').selectOption(['model-'+width]);
      await page.getByRole('button',{name:'保存',exact:true}).click();await page.locator('#editor-error').filter({hasText:'写入结果未知'}).waitFor();
      await page.getByRole('button',{name:'安全读回',exact:true}).click();await page.locator('#editor-error').filter({hasText:'读回失败，结果仍未知'}).waitFor();
      assert.equal(await page.getByRole('button',{name:'安全读回',exact:true}).count(),0);assert.equal(mutations,1);
      await screenshot('historical-readback-red');
      facts.cases.push({case:'QA35-01-historical-red',evidence_kind:'isolated_reconstruction_of_previous_editor_not_current_asset',asset_sha256:assetHash,mutations,committed_status:201,read_status:503,retry_present:false,result:'expected_failure_reproduced'});
    } finally {
      await page.unroute(assetRoute);await page.unroute(writeRoute);await page.unroute(readRoute);
      await page.reload();await page.locator('#workspace').waitFor({state:'visible'});await page.getByLabel('当前租户').selectOption(tenant);
    }
  }
  async function unknownWrite(kind) {
    phase(kind+'-open');
    await nav('API Keys');
    const before=await read(keysPath),usage=await read(keysPath+'/'+issued.id+'/limits');
    const beforeKey=before.items.find(key=>key.id===issued.id);
    let mutations=0,reads=0,operation,committedStatus;
    let release;
    const barrier=new Promise(resolve=>{release=resolve;});
    const path='/admin/api/'+keysPath+(kind==='rotation'?'/'+issued.id+'/rotations':'');
    const routePath='**'+path,operations='**/admin/api/operations/*';
    await page.route(routePath,async route=>{
      if(route.request().method()!=='POST')return route.continue();
      mutations++;operation=route.request().headers()['idempotency-key'];
      const response=await route.fetch();committedStatus=response.status();assert.equal(committedStatus,kind==='rotation'?200:201);
      facts.observations.push({phase:kind,method:'POST',status:committedStatus,mutations});
      await barrier;await route.abort('failed');
    });
    await page.route(operations,async route=>{
      reads++;assert.equal(decodeURIComponent(new URL(route.request().url()).pathname.split('/').at(-1)),operation);
      if(reads<=2)return route.fulfill({status:reads===1?503:404,contentType:'application/json',body:'{"error":{"code":"controlled_read_failure"}}'});
      return route.continue();
    });
    try {
      if(kind==='rotation') {
        await page.getByRole('row').filter({hasText:'Application '+width}).getByRole('button',{name:'轮换',exact:true}).click();
        await page.getByLabel('旧代际过渡').selectOption('1');
      } else {
        await page.getByRole('button',{name:'创建Key',exact:true}).click();
        await page.getByLabel('名称',{exact:true}).fill('Unknown '+width);
        await page.getByLabel('应用',{exact:true}).fill('controlled unknown recovery');
        await page.getByLabel('模型授权（可多选）').selectOption(['model-'+width]);
      }
      await screenshot(kind+'-before');
      phase(kind+'-double-submit');
      await page.getByRole('button',{name:'保存',exact:true}).dblclick();await page.keyboard.press('Enter');release();
      await page.locator('#editor-error').filter({hasText:'写入结果未知'}).waitFor();
      assert.equal(await page.getByRole('button',{name:'保存',exact:true}).isVisible(),false);
      assert.ok((await page.locator('#editor-error').textContent()).includes(operation));
      if(kind==='issuance')await page.getByLabel('名称',{exact:true}).press('Enter');
      await screenshot(kind+'-unknown');
      for(const expected of [503,404]) {
        phase(kind+'-read-'+expected);
        const receipt=page.waitForResponse(response=>new URL(response.url()).pathname==='/admin/api/operations/'+operation);
        await page.getByRole('button',{name:'安全读回',exact:true}).click();assert.equal((await receipt).status(),expected);
        facts.observations.push({phase:kind,method:'GET',status:expected,reads});
        await page.locator('#editor-error').filter({hasText:'读回失败，结果仍未知'}).waitFor();
        await page.waitForFunction(()=>[...document.querySelectorAll('#editor-error button')].some(button=>button.textContent==='安全读回'&&!button.disabled));
        assert.equal(await page.getByRole('button',{name:'安全读回',exact:true}).isEnabled(),true);
        assert.ok((await page.locator('#editor-error').textContent()).includes(operation));
        await screenshot(kind+'-read-'+expected);
      }
      phase(kind+'-recover');
      const recoveredReceipt=page.waitForResponse(response=>new URL(response.url()).pathname==='/admin/api/operations/'+operation);
      await page.getByRole('button',{name:'安全读回',exact:true}).click();
      const recoveredResponse=await recoveredReceipt;assert.equal(recoveredResponse.status(),200);
      const recovered=await recoveredResponse.json();assert.equal(recovered.secret_recoverable,false);assert.equal(Object.hasOwn(recovered,'key'),false);
      assert.equal(recovered.action,kind==='rotation'?'key.rotations':'key.create');
      if(kind==='rotation')assert.equal(recovered.object_id,issued.id);
      await page.locator('#detail-title').filter({hasText:'操作已提交'}).waitFor();
      assert.equal(await page.getByLabel('一次性API Key').count(),0);await screenshot(kind+'-recovered');
      assert.equal(mutations,1);assert.ok(reads>=3);
      await page.getByRole('button',{name:'关闭详情',exact:true}).click();await page.getByRole('button',{name:'关闭编辑',exact:true}).click();
      const after=await read(keysPath),afterKey=after.items.find(key=>key.id===issued.id);
      assert.equal(after.items.length,before.items.length+(kind==='rotation'?0:1));
      if(kind==='rotation') {
        assert.equal(afterKey.revision,beforeKey.revision+1);assert.equal(afterKey.generations.length,beforeKey.generations.length+1);
        assert.equal(afterKey.generations.filter(value=>!value.retired&&value.expires).length,1);
        assert.deepEqual(await read(keysPath+'/'+issued.id+'/limits'),usage);
        assert.deepEqual(await read('workspace/requests/'+requestID),snapshot);
        await page.unroute(routePath);
        phase('rotation-stacked-transition');
        await nav('API Keys');
        await page.getByRole('row').filter({hasText:'Application '+width}).getByRole('button',{name:'轮换',exact:true}).click();
        await page.getByLabel('旧代际过渡').selectOption('1');
        const denied=page.waitForResponse(response=>new URL(response.url()).pathname===path);
        await page.getByRole('button',{name:'保存',exact:true}).click();assert.equal((await denied).status(),409);
        await page.locator('#editor-error').filter({hasText:'操作失败：'}).waitFor();
        await page.getByRole('button',{name:'关闭编辑',exact:true}).click();
        assert.deepEqual((await read(keysPath)).items.find(key=>key.id===issued.id),afterKey);
      } else {
        assert.equal(after.items.filter(key=>key.name==='Unknown '+width).length,1);
        assert.equal(after.items.find(key=>key.name==='Unknown '+width).id,recovered.object_id);
      }
      facts.cases.push({case:kind+'-lost-response',mutations,reads,committed_status:committedStatus,read_statuses:[503,404,200],secret_replayed:false,...(kind==='rotation'?{stacked_transition_status:409}:{}),result:'passed'});
    } finally {release();await page.unroute(routePath);await page.unroute(operations);}
  }
  try {
  await historicalRed();await unknownWrite('issuance');await unknownWrite('rotation');
  for(const stream of [false,true]) {
    phase(stream?'double-sse':'double-normal');
    await nav('调用测试');
    await page.getByLabel('应用API Key（不会保存）').fill(issued.key);await page.getByLabel('公共模型名',{exact:true}).fill('model-'+width);
    await page.getByLabel('消息',{exact:true}).fill(stream?'controlled double SSE':'controlled double normal');
    await page.getByLabel('SSE流式',{exact:true}).setChecked(stream);
    const before=await read(keysPath+'/'+issued.id+'/limits'),dispatchBefore=await upstreamFacts();let posts=0;
    const routePath='**/v1/chat/completions';
    await page.route(routePath,async route=>{posts++;await route.continue();});
    try {
      const receipt=page.waitForResponse(response=>new URL(response.url()).pathname==='/v1/chat/completions');
      await page.getByRole('button',{name:'发送',exact:true}).dblclick();await page.keyboard.press('Enter');
      const response=await receipt;assert.equal(response.status(),200);
      await page.waitForFunction(()=>!document.querySelector('#call-form button[type=submit]').disabled);
      const id=response.headers()['x-request-id'];assert.ok(id);
      const detail=await read('workspace/requests/'+id),after=await read(keysPath+'/'+issued.id+'/limits');
      assert.equal(posts,1);assert.equal(after.usage.calls-before.usage.calls,1);assert.equal(detail.attempts.length,1);
      assert.equal((await upstreamFacts())-dispatchBefore,1);
      assert.equal(detail.charge_entries.reduce((sum,entry)=>sum+entry.input_tokens,0),3);
      assert.equal(detail.charge_entries.reduce((sum,entry)=>sum+entry.output_tokens,0),4);assert.equal(detail.pending,0);
      await page.getByRole('button',{name:'查看同Request ID日志'}).click();await page.locator('#detail-title').filter({hasText:id}).waitFor();
      await page.getByRole('button',{name:'关闭详情',exact:true}).click();
      await nav('API Keys');await screenshot(stream?'double-sse':'double-normal');
      facts.cases.push({case:stream?'double-sse':'double-normal',proxy_posts:posts,admissions:1,attempts:1,actual_upstream_dispatches:1,input_tokens:3,output_tokens:4,result:'passed'});
    } finally {await page.unroute(routePath);}
  }
  phase('completed');console.log(JSON.stringify({suite:facts.suite,width,result:'passed',cases:facts.cases}));
  } finally {await writeFile(`test-results/fault-${width}.json`,JSON.stringify(facts,null,2)+'\n');}
}
