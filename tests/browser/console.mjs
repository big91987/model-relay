import assert from 'node:assert/strict';
import {mkdir,writeFile} from 'node:fs/promises';
import {chromium} from 'playwright';
import {consoleFaults} from './console-faults.mjs';
const origin=process.env.RELAY_CONSOLE_ORIGIN;
const browser=await chromium.launch({headless:true});
let stage='start';
let diagnostics={};
let activePage;
async function keyboardSnapshot(page) {
  return page.evaluate(()=>{
    const editor=document.getElementById('editor'),active=document.activeElement;
    return {
      keyboard_editor_open:editor?.open===true,
      keyboard_focus_inside:!!active&&!!editor?.contains(active),
      keyboard_focus_body:!!active&&active===document.body,
      keyboard_focus_root:!!active&&active===document.documentElement,
      keyboard_document_focused:document.hasFocus()
    };
  });
}
async function keyboardJourney(page,width) {
  diagnostics.keyboard_focus_trace=[];
  const observe=async(phase,tabCount)=>diagnostics.keyboard_focus_trace.push({phase,tab_count:tabCount,...await keyboardSnapshot(page)});
  stage=`keyboard-dialog-${width}`;
  await page.getByRole('button',{name:'创建Key',exact:true}).click();
  stage=`keyboard-dialog-ready-${width}`;
  await page.locator('#editor').waitFor({state:'visible'});
  await observe('open',0);
  stage=`keyboard-first-tab-${width}`;
  await page.keyboard.press('Tab');
  await observe('tab',1);
  assert.ok(await page.evaluate(()=>document.getElementById('editor').contains(document.activeElement)));
  for(let index=0;index<12;index++) {
    stage=`keyboard-tab-cycle-${index+2}-${width}`;
    await page.keyboard.press('Tab');
    await observe('tab',index+2);
  }
  assert.ok(await page.evaluate(()=>document.getElementById('editor').contains(document.activeElement)));
  stage=`keyboard-escape-${width}`;
  await page.keyboard.press('Escape');
  await observe('escape',13);
  assert.equal(await page.locator('#editor').isVisible(),false);
  stage=`keyboard-return-focus-${width}`;
  await observe('return',13);
  assert.ok(await page.evaluate(()=>document.activeElement.textContent==='创建Key'));
}
async function consoleSnapshot(page) {
  return page.evaluate(()=>{
    const editor=document.getElementById('editor'),save=document.getElementById('editor-save'),callKey=document.getElementById('call-key');
    const selected=document.querySelector('#navigation [aria-current="page"]')?.dataset.page;
    const allowedPages=['overview','models','services','playground','keys','tenants','limits','usage','billing','records'];
    const notice=document.getElementById('notice')?.textContent??'',editorError=document.getElementById('editor-error')?.textContent??'';
    const modelSelect=document.getElementById('field-models');
    return {
      active_page:allowedPages.includes(selected)?selected:'unknown',editor_open:editor?.open===true,
      save_hidden:save?.hidden===true,save_disabled:save?.disabled===true,
      editor_invalid_controls:editor?.querySelectorAll('input:invalid,select:invalid,textarea:invalid').length??0,
      model_option_count:modelSelect?.options.length??0,model_selected_count:modelSelect?.selectedOptions.length??0,
      secret_control_count:document.querySelectorAll('[aria-label="一次性API Key"]').length,
      call_key_present:!!callKey,call_key_nonempty:!!callKey?.value,call_key_disabled:callKey?.disabled===true,
      call_invalid_controls:document.getElementById('call-form')?.querySelectorAll('input:invalid,textarea:invalid').length??0,
      detail_open:document.getElementById('detail')?.open===true,
      dependency_failure:notice.startsWith('读取编辑依赖失败：'),action_failure:notice.startsWith('操作失败：'),
      write_failure:editorError.startsWith('操作失败：'),write_unknown:editorError.startsWith('写入结果未知，')
    };
  });
}
async function save(page,path,method='POST',status=201) {
  assert.equal(await page.locator('#editor-save').isEnabled(),true);
  const [acknowledgement]=await Promise.all([
    page.waitForResponse(response=>new URL(response.url()).pathname===path && response.request().method()===method),
    page.getByRole('button',{name:'保存',exact:true}).click()
  ]);
  diagnostics.write_status=acknowledgement.status();
  assert.equal(diagnostics.write_status,status);
  return acknowledgement;
}
async function navigate(page,title) { await page.locator('#navigation').getByRole('button',{name:title,exact:true}).click(); await page.locator('#page-title').filter({hasText:title}).waitFor(); await page.locator('#content').getByText('正在读取当前范围…').waitFor({state:'hidden'}); }
async function interruptedEditorJourney(page,pressNav,write,width) {
  const pattern=/\/admin\/api\/tenants(?:\?[^#]*)?$/;
  let releaseReadback,readbackStarted,timeout;
  let held=false,writes=0,response;
  const onRequest=request=>{if(request.method()==='POST'&&new URL(request.url()).pathname==='/admin/api/upstreams')writes++;};
  const barrier=new Promise(resolve=>{releaseReadback=resolve;});
  const started=new Promise(resolve=>{readbackStarted=resolve;});
  let continuation=Promise.resolve();
  const handler=route=>{
    if(route.request().method()!=='GET'||held)return route.continue();
    held=true;diagnostics.readback_checkpoint_reached=true;readbackStarted();continuation=barrier.then(()=>route.continue());return continuation;
  };
  await page.route(pattern,handler);
  page.on('request',onRequest);
  try {
    response=await write();
    stage=`editor-held-readback-${width}`;
    diagnostics.readback_checkpoint_reached=held;
    await Promise.race([started,new Promise((resolve,reject)=>{timeout=setTimeout(()=>reject(new Error('controlled readback checkpoint not reached')),5000);})]);
    clearTimeout(timeout);
    stage=`editor-interrupted-navigation-${width}`;
    await pressNav('供应商与服务');
    await page.getByRole('button',{name:'接入服务',exact:true}).click();
    diagnostics.interrupted_editor_save_enabled=await page.locator('#editor-save').isEnabled();
    assert.equal(diagnostics.interrupted_editor_save_enabled,true);
    await page.getByRole('button',{name:'取消',exact:true}).click();
    assert.equal(writes,1);
  } finally {
    clearTimeout(timeout);releaseReadback();page.off('request',onRequest);
    try { await continuation; } finally { await page.unroute(pattern,handler); }
  }
  const evidence={suite:'console-editor-interruption',width,result:'passed',controlled_get_hold:true,new_editor_save_enabled:true,service_posts:writes,canceled_editor_posts:0};
  await writeFile(`test-results/editor-interruption-${width}.json`,JSON.stringify(evidence)+'\n');
  console.log(JSON.stringify(evidence));
  return response;
}
try {
  await mkdir('test-results',{recursive:true});
  for(const width of [1280,1440,390]) {
    diagnostics={};
    const context=await browser.newContext({viewport:{width,height:900}}),page=await context.newPage();
    activePage=page;
    const currentDiagnostics=diagnostics;
    currentDiagnostics.page_error_count=0;currentDiagnostics.request_failed_count=0;
    page.on('pageerror',()=>{currentDiagnostics.page_error_count++;});
    page.on('requestfailed',()=>{currentDiagnostics.request_failed_count++;});
    page.on('response',response=>{
      const path=new URL(response.url()).pathname,method=response.request().method();
      if(method==='GET'&&path==='/admin/api/catalog')currentDiagnostics.catalog_status=response.status();
      if(method==='GET'&&/^\/admin\/api\/tenants\/[^/]+\/members$/.test(path))currentDiagnostics.members_status=response.status();
      if(method==='POST'&&/^\/admin\/api\/tenants\/[^/]+\/keys$/.test(path))currentDiagnostics.key_create_status=response.status();
      if(method==='POST'&&path==='/v1/chat/completions')currentDiagnostics.call_status=response.status();
      if(method==='GET'&&/^\/admin\/api\/workspace\/requests\/[^/]+$/.test(path))currentDiagnostics.request_detail_status=response.status();
    });
    const pressNav=async title=>{if(width===390)await page.getByRole('button',{name:'菜单',exact:true}).click();await navigate(page,title);};
    stage=`login-${width}`;await page.goto(origin+'/admin/');await page.getByLabel('账号',{exact:true}).fill('admin');await page.getByLabel('密码',{exact:true}).fill(process.env.RELAY_CONSOLE_PASSWORD);await page.getByRole('button',{name:'登录',exact:true}).click();await page.locator('#workspace').waitFor({state:'visible'});
    stage=`tenant-${width}`;await pressNav('租户与成员');await page.getByRole('button',{name:'创建租户',exact:true}).click();await page.getByLabel('租户名称').fill('Team '+width);const created=await save(page,'/admin/api/tenants');const tenant=(await created.json()).id;await page.getByLabel('当前租户').selectOption(tenant);
    stage=`service-create-${width}`;await pressNav('供应商与服务');await page.getByRole('button',{name:'接入服务',exact:true}).click();await page.getByLabel('服务名称').fill('Controlled '+width);await page.getByLabel('兼容API基础地址').fill(process.env.RELAY_CONSOLE_UPSTREAM);await page.getByLabel('上游凭据（编辑留空保留同地址原值）').fill('controlled-console-upstream');const upstreamResponse=await interruptedEditorJourney(page,pressNav,()=>save(page,'/admin/api/upstreams'),width);const upstream=(await upstreamResponse.json()).id;
    stage=`service-check-http-${width}`;
    const row=page.getByRole('row').filter({hasText:'Controlled '+width});const checked=page.waitForResponse(response=>new URL(response.url()).pathname===`/admin/api/upstreams/${upstream}/check`);await row.getByRole('button',{name:'核验',exact:true}).click();
    const checkResponse=await checked,checkInput=checkResponse.request().postDataJSON();
    diagnostics.check_status=checkResponse.status();
    diagnostics.check_revision_present=Number.isSafeInteger(checkInput?.expected_execution_revision)&&checkInput.expected_execution_revision>0;
    assert.equal(diagnostics.check_status,200);
    stage=`service-check-contract-${width}`;
    const checkResult=await checkResponse.json();
    diagnostics.check_success=checkResult.state==='success';
    diagnostics.check_applicable=checkResult.applicable===true;
    diagnostics.check_revision_matches=checkResult.execution_revision===checkInput.expected_execution_revision;
    assert.equal(diagnostics.check_revision_present,true);assert.equal(diagnostics.check_success,true);assert.equal(diagnostics.check_applicable,true);assert.equal(diagnostics.check_revision_matches,true);
    stage=`service-check-dialog-${width}`;await page.getByRole('button',{name:'关闭详情',exact:true}).click();
    stage=`model-dialog-${width}`;await pressNav('模型托管');await page.getByRole('button',{name:'接入模型',exact:true}).click();
    stage=`model-fields-${width}`;await page.getByLabel('公共模型名',{exact:true}).fill('model-'+width);await page.getByLabel('服务',{exact:true}).selectOption(String(upstream));
    diagnostics.remote_label_matches=await page.getByLabel('远端模型名',{exact:true}).count();assert.equal(diagnostics.remote_label_matches,1);
    await page.getByLabel('远端模型名',{exact:true}).fill('controlled-model');
    stage=`model-create-http-${width}`;await save(page,'/admin/api/models');
    stage=`model-grants-dialog-${width}`;const modelRow=page.getByRole('row').filter({hasText:'model-'+width});await modelRow.getByRole('button',{name:'租户授权',exact:true}).click();await page.getByLabel('允许使用此模型的租户',{exact:true}).selectOption([tenant]);
    stage=`model-grants-http-${width}`;await save(page,`/admin/api/models/model-${width}/tenant-grants`,'PUT',200);
    stage=`price-${width}`;await pressNav('定价与账单');const priceRow=page.getByRole('row').filter({hasText:'model-'+width});await priceRow.getByRole('button',{name:'发布新价',exact:true}).click();await page.getByLabel('服务ID',{exact:true}).fill(String(upstream));for(const label of ['供应商输入成本','供应商输出成本','租户输入价格','租户输出价格'])await page.getByLabel(label+'（微美元 / 百万Token）').fill('1000000');await save(page,`/admin/api/models/model-${width}/prices`);
    stage=`key-navigation-${width}`;await pressNav('API Keys');
    stage=`key-dialog-${width}`;await page.getByRole('button',{name:'创建Key',exact:true}).click();
    stage=`key-fields-${width}`;await page.getByLabel('名称',{exact:true}).fill('Application '+width);await page.getByLabel('应用',{exact:true}).fill('Controlled browser');
    stage=`key-model-selection-${width}`;await page.getByLabel('模型授权（可多选）').selectOption(['model-'+width]);
    const issuance=page.waitForResponse(response=>new URL(response.url()).pathname===`/admin/api/tenants/${tenant}/keys`&&response.request().method()==='POST');
    stage=`key-create-http-${width}`;await save(page,`/admin/api/tenants/${tenant}/keys`);const issued=await (await issuance).json();
    diagnostics.issuance_has_secret=typeof issued.key==='string'&&issued.key.length>0;diagnostics.issuance_has_id=typeof issued.id==='string'&&issued.id.length>0;
    stage=`key-secret-dialog-${width}`;await page.getByLabel('一次性API Key').waitFor();assert.equal(await page.getByLabel('一次性API Key').getAttribute('type'),'password');
    stage=`key-secret-transfer-${width}`;await page.getByRole('button',{name:'显式用于调用测试'}).click();await page.locator('#call-key').waitFor();
    stage=`key-call-fields-${width}`;await page.getByLabel('公共模型名',{exact:true}).fill('model-'+width);await page.getByLabel('消息',{exact:true}).fill('controlled browser input');
    stage=`key-secret-transfer-value-${width}`;await page.waitForFunction(()=>!!document.getElementById('call-key')?.value);assert.equal(await page.evaluate(secret=>document.getElementById('call-key')?.value===secret,issued.key),true);
    const generation=page.waitForResponse(response=>new URL(response.url()).pathname==='/v1/chat/completions');
    stage=`key-call-http-${width}`;await page.getByRole('button',{name:'发送',exact:true}).click();const response=await generation;diagnostics.call_status=response.status();assert.equal(diagnostics.call_status,200);const requestID=response.headers()['x-request-id'];assert.ok(requestID);
    stage=`key-request-detail-action-${width}`;await page.getByRole('button',{name:'查看同Request ID日志'}).click();
    stage=`key-request-detail-visible-${width}`;await page.locator('#detail-title').filter({hasText:requestID}).waitFor();await page.getByRole('button',{name:'关闭详情',exact:true}).click();
    stage=`sse-${width}`;await pressNav('调用测试');await page.getByLabel('应用API Key（不会保存）').fill(issued.key);await page.getByLabel('公共模型名',{exact:true}).fill('model-'+width);await page.getByLabel('消息',{exact:true}).fill('controlled stream input');await page.getByLabel('SSE流式',{exact:true}).check();const streamed=page.waitForResponse(response=>new URL(response.url()).pathname==='/v1/chat/completions');await page.getByRole('button',{name:'发送',exact:true}).click();assert.equal((await streamed).status(),200);await page.locator('#call-output').filter({hasText:'controlled SSE response'}).waitFor();await page.getByRole('button',{name:'查看同Request ID日志'}).click();await page.locator('#detail-title').waitFor();await page.getByRole('button',{name:'关闭详情',exact:true}).click();
    stage=`cancel-${width}`;await pressNav('调用测试');await page.getByLabel('应用API Key（不会保存）').fill(issued.key);await page.getByLabel('公共模型名',{exact:true}).fill('model-'+width);await page.getByLabel('消息',{exact:true}).fill('controlled cancel input');await page.getByLabel('SSE流式',{exact:true}).check();await page.getByRole('button',{name:'发送',exact:true}).click();await page.locator('#call-output').filter({hasText:'controlled SSE response'}).waitFor();await page.getByRole('button',{name:'停止本地接收（费用可能待确认）',exact:true}).click();await page.locator('#call-output').filter({hasText:'已停止本地接收'}).waitFor();await page.getByRole('button',{name:'查看同Request ID日志'}).click();await page.locator('#detail-title').waitFor();await page.getByRole('button',{name:'关闭详情',exact:true}).click();
    stage=`rotate-${width}`;await pressNav('API Keys');const rotatingRow=page.getByRole('row').filter({hasText:'Application '+width});await rotatingRow.getByRole('button',{name:'轮换',exact:true}).click();const rotatedResponse=page.waitForResponse(response=>new URL(response.url()).pathname===`/admin/api/tenants/${tenant}/keys/${issued.id}/rotations`&&response.request().method()==='POST');await save(page,`/admin/api/tenants/${tenant}/keys/${issued.id}/rotations`,'POST',200);const rotated=await (await rotatedResponse).json();await page.getByLabel('一次性API Key').waitFor();await page.getByRole('button',{name:'关闭编辑'}).click();assert.equal((await page.request.get(origin+'/v1/models',{headers:{Authorization:'Bearer '+issued.key}})).status(),401);assert.equal((await page.request.get(origin+'/v1/models',{headers:{Authorization:'Bearer '+rotated.key}})).status(),200);issued.key=rotated.key;
    stage=`invitation-${width}`;await pressNav('租户与成员');await page.getByRole('button',{name:'邀请成员',exact:true}).click();const account=`owner${width}@controlled.example`;await page.getByLabel('邮箱账号').fill(account);await page.getByLabel('角色',{exact:true}).selectOption('administrator');const invitationResponse=page.waitForResponse(response=>new URL(response.url()).pathname===`/admin/api/tenants/${tenant}/invitations`&&response.request().method()==='POST');await save(page,`/admin/api/tenants/${tenant}/invitations`);const invitation=await (await invitationResponse).json();await page.getByRole('button',{name:'关闭编辑'}).click();const memberContext=await browser.newContext({viewport:{width,height:900}}),member=await memberContext.newPage();await member.goto(origin+'/admin/accept#'+invitation.acceptance_token);await member.getByLabel('新账户密码（至少12字符）').fill('controlled-member-password');await member.getByRole('button',{name:'接受邀请',exact:true}).click();await member.locator('#login-panel').waitFor({state:'visible'});await member.getByLabel('账号',{exact:true}).fill(account);await member.getByLabel('密码',{exact:true}).fill('controlled-member-password');await member.getByRole('button',{name:'登录',exact:true}).click();await member.locator('#workspace').waitFor({state:'visible'});assert.equal(await member.getByLabel('当前租户').inputValue(),tenant);if(width===390)await member.getByRole('button',{name:'菜单',exact:true}).click();await navigate(member,'供应商与服务');assert.equal(await member.getByRole('button',{name:'接入服务',exact:true}).count(),0);await member.getByRole('button',{name:'退出',exact:true}).click();await member.locator('#login-panel').waitFor({state:'visible'});assert.equal((await member.request.get(origin+'/admin/api/session')).status(),401);await memberContext.close();
    stage=`native-faults-${width}`;await consoleFaults(page,{origin,tenant,width,issued,requestID,nav:pressNav,step:phase=>{stage=`fault-${phase}-${width}`;}});
    stage=`quota-${width}`;await pressNav('配额与预算');await page.getByRole('button',{name:'调整租户限额'}).click();await page.getByLabel('本周期次数').fill('3');await save(page,`/admin/api/tenants/${tenant}/limits`,'PUT',200);await pressNav('调用测试');await page.getByLabel('应用API Key（不会保存）').fill(issued.key);await page.getByLabel('公共模型名',{exact:true}).fill('model-'+width);await page.getByLabel('消息',{exact:true}).fill('second controlled input');const denied=page.waitForResponse(response=>new URL(response.url()).pathname==='/v1/chat/completions');await page.getByRole('button',{name:'发送',exact:true}).click();assert.equal((await denied).status(),429);
    stage=`all-pages-${width}`;for(const title of ['工作台','模型托管','供应商与服务','API Keys','租户与成员','配额与预算','用量分析','定价与账单','调用日志','调用测试']){await pressNav(title);assert.equal(await page.locator('h1:visible').count(),1);assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));}
    stage=`keyboard-navigation-${width}`;await pressNav('API Keys');await keyboardJourney(page,width);
    stage=`secret-free-screenshot-${width}`;assert.ok(await page.evaluate(()=>[...document.querySelectorAll('.secret,input[type=password]')].every(input=>input.value==='')));await page.screenshot({path:`test-results/console-${width}.png`,fullPage:true});await page.getByRole('button',{name:'退出',exact:true}).click();await page.locator('#login-panel').waitFor({state:'visible'});assert.ok(await page.evaluate(()=>[...document.querySelectorAll('.secret,input[type=password]')].every(input=>input.value==='')));await context.close();console.log(JSON.stringify({suite:'console-real-go-ui',width,result:'passed',provider:'controlled_http_not_supplier',coverage:'ten-pages + scoped issuance/call/ledger/limit rejection + dialog keyboard + secret cleanup'}));
  }
} catch(error) {const errorKind=error.message?.includes('strict mode violation')?'ambiguous_locator':error.name==='TimeoutError'?'timeout':error.name==='AssertionError'?'assertion':'other';const errorName=['TimeoutError','AssertionError','Error'].includes(error.name)?error.name:'Other';let snapshot={snapshot_available:false};if(activePage)try{snapshot={snapshot_available:true,...await consoleSnapshot(activePage)};}catch{}console.error(JSON.stringify({suite:'console-real-go-ui',stage,result:'failed',error_name:errorName,error_kind:errorKind,...diagnostics,...snapshot}));process.exitCode=1;} finally {await browser.close();}
