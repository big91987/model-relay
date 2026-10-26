import assert from 'node:assert/strict';
import {mkdir, readFile, writeFile} from 'node:fs/promises';
import {chromium} from 'playwright';
import {consoleVisualReview} from './console-visual.mjs';
import {managementSummaryJourney,summaryCell,summaryFailureFacts} from './console-summary.mjs';
import {modelSaveReadbackJourney,waitForSaveReadback} from './model-save-readback.mjs';
import {readabilityEvidence,readabilityFailureFacts} from './readability.mjs';
import {captureIntegrityScreenshot} from './capture-integrity.mjs';
import {captureStreamBody} from './stream-capture.mjs';
import {latencyJourney} from './latency-journey.mjs';

const origin=process.env.RELAY_JOURNEY_ORIGIN;
const width=Number(process.env.RELAY_JOURNEY_WIDTH);
const mode=process.env.RELAY_JOURNEY_MODE??'provider';
const browser=await chromium.launch({headless:true});
const contexts=[];
const pages=[];
const checkpoints=[];
let stage='start',action='start',lastStatus=null;
let callFacts={};
let usageFacts={};
let readFacts={};
let membershipFacts={};
let ledgerFacts={};
let selfHostedFacts={};
const selfHostedProgress={steps:[],screenshots:[]};
const initialCounts={};
let activePage;
let now;
const mark=value=>{stage=value;};
const complete=journey=>{checkpoints.push(journey);console.log(JSON.stringify({suite:'continuous-console',width,journey,result:'passed',provider:'controlled_http_not_supplier'}));};
const button=(page,name)=>page.getByRole('button',{name,exact:true});
const row=(page,name)=>page.getByRole('row').filter({hasText:name});
const items=value=>value.items??value.data??value;
async function control(path='facts',seconds) {
  const response=await fetch(process.env.RELAY_JOURNEY_CONTROL+'/'+path,{method:seconds==null?'GET':'POST',headers:{'Content-Type':'application/json','X-Test-Capability':process.env.RELAY_JOURNEY_CAPABILITY},...(seconds==null?{}:{body:JSON.stringify({seconds})})});
  assert.equal(response.status,200);
  return response.json();
}
async function newPage() {
  const context=await browser.newContext({viewport:{width,height:900},acceptDownloads:true});
  const page=await context.newPage();
  contexts.push(context);pages.push(page);activePage=page;
  await page.clock.setFixedTime(new Date(now));
  return page;
}
async function advance(seconds) {
  now=(await control('advance',seconds)).now;
  for(const page of pages)await page.clock.setFixedTime(new Date(now));
}
async function ready(page) {
  await page.locator('#content').getByText('正在读取当前范围…',{exact:true}).waitFor({state:'hidden'});
}
async function nav(page,title) {
  action='navigation';
  activePage=page;
  if(width===390)await button(page,'菜单').click();
  await page.locator('#navigation').getByRole('button',{name:title,exact:true}).click();
  await page.locator('#page-title').filter({hasText:title}).waitFor();
  await ready(page);
}
async function login(page,account,password) {
  action='personnel-login';
  activePage=page;
  if(await page.locator('#workspace').isVisible()){
    await button(page,'退出').click();await page.locator('#login-panel').waitFor({state:'visible'});
  }
  await page.goto(origin+'/admin/');
  await page.locator('#login-panel').waitFor({state:'visible'});
  await page.getByLabel('账号',{exact:true}).fill(account);
  await page.getByLabel('密码',{exact:true}).fill(password);
  await button(page,'登录').click();
  await page.locator('#workspace').waitFor({state:'visible'});
  await ready(page);
}
async function scope(page,tenant) {
  action='tenant-scope';
  await page.getByLabel('当前租户').selectOption(tenant);
  await ready(page);
  assert.equal(await page.getByLabel('当前租户').inputValue(),tenant);
}
async function save(page,path,method='POST',status=201,keyboard=false) {
  action='form-submit';
  const expected=await page.evaluate(()=>({operation:document.getElementById('editor-form').dataset.operation,tenant:document.getElementById('tenant').value,page:document.getElementById('page-title').textContent}));
  const receipt=page.waitForResponse(response=>new URL(response.url()).pathname===path&&response.request().method()===method);
  if(keyboard){await tabTo(page,'editor-save');await page.keyboard.press('Enter');}else await button(page,'保存').click();
  const response=await receipt;lastStatus=response.status();assert.equal(lastStatus,status);
  const value=await response.json();
  if(status>=400){await page.locator('#editor-error').filter({hasText:'操作失败：'}).waitFor();}
  else if(value.key||value.acceptance_token)await page.getByLabel('一次性API Key').waitFor();
  else {await waitForSaveReadback(page,expected);await page.locator('#editor').waitFor({state:'hidden'});await ready(page);}
  return value;
}
async function read(page,path,status=200) {
  action='authoritative-read';
  activePage=page;
  const response=await page.request.get(origin+'/admin/api/'+path);
  lastStatus=response.status();
  readFacts={expected_status:status,actual_status:lastStatus};
  assert.equal(response.status(),status);
  return response.json();
}
function failureBoundary(error) {
  const match=typeof error.stack==='string'?error.stack.match(/\n\s+at (?:[^\n]* \()?file:\/\/[^\n]*\/tests\/browser\/(journeys|console-summary|readability)\.mjs:(\d{1,6}):(\d{1,6})\)?(?:\n|$)/):null;
  return match?{source:'tests/browser/'+match[1]+'.mjs',line:Number(match[2]),column:Number(match[3])}:{available:false};
}
function conservationFacts(before,after) {
  const amount=value=>typeof value==='string'&&/^-?\d{1,19}$/.test(value)?value:null;
  const count=value=>Number.isSafeInteger(value)&&value>=0?value:null;
  const result={period_same:before.period===after.period};
  for(const field of ['fee_micro_usd','reserved_micro_usd']){
    result[field]={before:amount(before.usage?.[field]),after:amount(after.usage?.[field])};
  }
  for(const field of ['calls','pending','concurrency']){
    result[field]={before:count(before.usage?.[field]),after:count(after.usage?.[field])};
  }
  return result;
}
function assertUsageConserved(before,after) {
  action='usage-conservation';
  usageFacts=conservationFacts(before,after);
  assert.equal(usageFacts.period_same,true);
  for(const field of ['fee_micro_usd','reserved_micro_usd','calls','pending','concurrency']){
    assert.notEqual(usageFacts[field].before,null);
    assert.notEqual(usageFacts[field].after,null);
    assert.equal(usageFacts[field].after,usageFacts[field].before);
  }
}
function assertReconciledLedger(before,after) {
  action='reconciliation-ledger';
  const entries=after.charge_entries;
  ledgerFacts={entries:Array.isArray(entries)?entries.length:null};
  assert.ok(Array.isArray(entries));
  assert.equal(entries.length,before.charge_entries.length+1);
  assert.deepEqual(entries.slice(0,-1),before.charge_entries);
  assert.deepEqual(after.price_snapshots,before.price_snapshots);
  assert.equal(before.pending,1);assert.equal(after.pending,0);assert.equal(after.fee_reserved,'0');
  for(const field of ['fee','cost','input_tokens','output_tokens']){
    let total=0n;
    for(const entry of entries){
      const amount=entry[field];
      assert.ok(field.endsWith('_tokens')?Number.isSafeInteger(amount)&&amount>=0:typeof amount==='string'&&/^\d{1,19}$/.test(amount));
      total+=BigInt(amount);
    }
    ledgerFacts[field]=total.toString();
  }
  assert.equal(ledgerFacts.fee,'21');assert.equal(ledgerFacts.cost,'7');
  assert.equal(ledgerFacts.input_tokens,'3');assert.equal(ledgerFacts.output_tokens,'4');
  const delta=entries.at(-1);
  assert.equal(delta.ordinal,1);assert.ok(Number.isSafeInteger(before.revision));
  assert.equal(delta.revision,before.revision+1);assert.equal(after.revision,delta.revision);
  assert.equal(delta.reason,'controlled authoritative test usage');
  assert.equal(delta.evidence,'controlled fixture HTTP usage contract; not supplier validation');
}
async function freshInstallation(page) {
  for(const resource of ['upstreams','models','tenants']){
    mark('J01-empty-'+resource);
    const values=items(await read(page,resource));
    assert.ok(Array.isArray(values));
    initialCounts[resource]=values.length;
    assert.equal(values.length,0);
  }
  action='empty-tenant-selector';
  const selector=page.getByLabel('当前租户');
  assert.equal(await selector.locator('option').count(),1);
  assert.equal(await selector.inputValue(),'');
  assert.equal(await selector.locator('option').textContent(),'全部租户');
}
async function close(page) {await button(page,'关闭编辑').click();await page.locator('#editor').waitFor({state:'hidden'});await ready(page);}
async function closeDetail(page) {await button(page,'关闭详情').click();await page.locator('#detail').waitFor({state:'hidden'});}
async function refresh(page,expectedTenantCount) {action='scope-refresh';await button(page,'刷新当前事实').click();if(expectedTenantCount!==undefined)await page.waitForFunction(count=>document.getElementById('tenant').options.length===count,expectedTenantCount,{timeout:5000});await ready(page);}
async function tabTo(page,id) {
  for(let index=0;index<30;index++){
    if(await page.evaluate(target=>document.activeElement?.id===target,id))return;
    await page.keyboard.press('Tab');
  }
  assert.fail('keyboard target not reachable');
}
async function tenant(page,name,keyboard=false) {
  await nav(page,'租户与成员');await button(page,'创建租户').click();
  if(keyboard){
    await page.keyboard.press('Enter');
    assert.equal(await page.locator('#field-name').evaluate(element=>!element.validity.valid),true);
    assert.equal(await page.locator('#editor').isVisible(),true);
    await page.keyboard.type(name);
  }else await page.getByLabel('租户名称').fill(name);
  const value=await save(page,'/admin/api/tenants','POST',201,keyboard);
  await page.locator('#editor').waitFor({state:'hidden'});
  await row(page,name).waitFor();await scope(page,value.id);
  return value.id;
}
async function budget(page,tenantID,amount,hard=true) {
  await nav(page,'配额与预算');await button(page,'调整租户限额').click();
  await page.getByLabel('费用预算（微美元）',{exact:true}).fill(String(amount));
  await page.getByLabel('硬费用预算（需要供应商可信收费Token上界）').setChecked(hard);
  await save(page,`/admin/api/tenants/${tenantID}/limits`,'PUT',200);
  await page.locator('#editor').waitFor({state:'hidden'});await ready(page);
}
async function invite(page,tenantID,account,role) {
  await nav(page,'租户与成员');await button(page,'邀请成员').click();
  await page.getByLabel('邮箱账号').fill(account);await page.getByLabel('角色',{exact:true}).selectOption(role);
  const value=await save(page,`/admin/api/tenants/${tenantID}/invitations`);
  await page.locator('#editor-title').filter({hasText:'邀请接受链接（手工交付）'}).waitFor();await close(page);
  const member=await newPage();
  action='invitation-accept';
  await member.goto(origin+'/admin/accept#'+value.acceptance_token);
  await member.getByLabel('新账户密码（至少12字符）').fill('controlled-journey-password');
  await button(member,'接受邀请').click();await member.locator('#login-panel').waitFor({state:'visible'});
  await login(member,account,'controlled-journey-password');
  assert.equal(await member.getByLabel('当前租户').inputValue(),tenantID);
  return member;
}
async function createKey(page,tenantID,name,models,diagnostic,expires) {
  await nav(page,'API Keys');await button(page,'创建Key').click();
  await page.getByLabel('名称',{exact:true}).fill(name);await page.getByLabel('应用',{exact:true}).fill(name+' application');
  await page.getByLabel('模型授权（可多选）').selectOption(models);
  if(diagnostic)await page.getByLabel('诊断开发者（可多选）').selectOption({label:diagnostic});
  if(expires)await page.getByLabel('UTC到期时间（留空为无限）').fill(expires);
  const value=await save(page,`/admin/api/tenants/${tenantID}/keys`);
  await page.getByLabel('一次性API Key').waitFor();
  assert.equal(await page.getByLabel('一次性API Key').getAttribute('type'),'password');
  assert.ok(await page.getByLabel('一次性API Key').evaluate((element,secret)=>element.value===secret,value.key));
  await close(page);await refresh(page);await row(page,name).waitFor();
  assert.equal(await page.getByLabel('一次性API Key').count(),0);
  return value;
}
async function call(page,key,model,status=200,message='controlled journey input',stream=false) {
  await nav(page,'调用测试');
  action='proxy-call';
  await page.getByLabel('应用API Key（不会保存）').fill(key);
  await page.getByLabel('公共模型名',{exact:true}).fill(model);await page.getByLabel('消息',{exact:true}).fill(message);
  await page.getByLabel('SSE流式',{exact:true}).setChecked(stream);
  const send=async()=>{
    const before=(await control()).dispatches;
    const receipt=page.waitForResponse(response=>new URL(response.url()).pathname==='/v1/chat/completions');
    await button(page,'发送').click();const response=await receipt;lastStatus=response.status();
    callFacts={expected_status:status,dispatch_delta:(await control()).dispatches-before};
    if(lastStatus!==status){
      const id=response.headers()['x-request-id'];
      if(id){
        try {
          const detail=await page.request.get(origin+'/admin/api/workspace/requests/'+encodeURIComponent(id),{timeout:5000});
          callFacts.detail_status=detail.status();
          if(detail.status()===200)Object.assign(callFacts,completionFacts(await detail.json()));
        } catch {callFacts.detail_available=false;}
      }
    }
    assert.equal(lastStatus,status);
    assert.equal(response.request().postDataJSON().stream,stream);
    await button(page,'发送').waitFor({state:'visible'});await page.waitForFunction(()=>!document.querySelector('#call-form button[type=submit]')?.disabled);
    const after=(await control()).dispatches;
    assert.equal(after-before,status===200?1:0);
    const id=response.headers()['x-request-id'];assert.ok(id);
    if(status===200)await page.locator('#call-output').filter({hasText:'controlled'}).waitFor();
    else await page.locator('#call-output').filter({hasText:'调用失败：'}).waitFor();
    return {id,value:stream&&status===200?null:await response.json()};
  };
  if(stream&&status===200) {
    const captured=await captureStreamBody(page,send);
    return {...captured.value,value:captured.body};
  }
  return send();
}
function completionFacts(value) {
  const states=['started','success','rejected','upstream_failed','upstream_rejected','stream_interrupted','timeout','canceled'];
  const codes=['','upstream_protocol_error','upstream_connection_failed','upstream_http_error','upstream_rejected','upstream_rate_limited','upstream_attempts_exhausted','upstream_timeout','request_timeout','storage_unavailable','client_canceled'];
  const numeric=value=>Number.isInteger(value)&&value>=100&&value<=599?value:null;
  const attempts=Array.isArray(value.attempts)?value.attempts:[];
  return {terminal:states.includes(value.state)?value.state:'other',terminal_code:codes.includes(value.error_code)?value.error_code:'other',upstream_status:numeric(value.upstream_status),pending:value.pending===1,attempt_count:attempts.length,attempts:attempts.slice(0,3).map(attempt=>({state:states.includes(attempt.state)?attempt.state:'other',code:codes.includes(attempt.error_code)?attempt.error_code:'other',upstream_status:numeric(attempt.upstream_status),dispatch_possible:attempt.dispatch_possible===1,input_known:Number.isInteger(attempt.prompt_tokens),output_known:Number.isInteger(attempt.completion_tokens)}))};
}
async function requestDetail(page,id,entry='call') {
  action='same-request-detail';
  const activation=entry==='records'?row(page,id).getByRole('button',{name:'详情 / 核对',exact:true}):button(page,'查看同Request ID日志');
  const [response]=await Promise.all([
    page.waitForResponse(response=>new URL(response.url()).pathname===`/admin/api/workspace/requests/${id}`),
    activation.click()
  ]);assert.equal(response.status(),200);
  await page.locator('#detail-title').filter({hasText:id}).waitFor();
  const value=await response.json();assert.equal(value.id,id);return value;
}
async function billDetail(page,period,id) {
  action='bill-detail';
  const [response]=await Promise.all([
    page.waitForResponse(response=>new URL(response.url()).pathname===`/admin/api/bills/${id}`&&response.request().method()==='GET'),
    row(page,period).getByRole('button',{name:'详情 / 更正',exact:true}).click()
  ]);
  assert.equal(response.status(),200);
  const value=await response.json();assert.equal(value.id,id);
  await page.locator('#detail').waitFor({state:'visible'});
  assert.equal(await page.locator('#detail-title').textContent(),'已结账，原金额不可覆写（调整仅计入各调整周期一次）');
  return value;
}
async function editKey(page,tenantID,key,name,models,enabled=true) {
  await nav(page,'API Keys');await row(page,name).getByRole('button',{name:'编辑',exact:true}).click();
  await page.getByLabel('模型授权（可多选）').selectOption(models);await page.getByLabel('启用',{exact:true}).setChecked(enabled);
  await save(page,`/admin/api/tenants/${tenantID}/keys/${key}`,'PATCH',200);
  await page.locator('#editor').waitFor({state:'hidden'});await row(page,name).waitFor();
}
async function rotation(page,tenantID,key,name,hours,status=200) {
  await nav(page,'API Keys');await row(page,name).getByRole('button',{name:'轮换',exact:true}).click();
  await page.getByLabel('旧代际过渡').selectOption(String(hours));
  const value=await save(page,`/admin/api/tenants/${tenantID}/keys/${key}/rotations`,'POST',status);
  if(status===200){await page.getByLabel('一次性API Key').waitFor();await close(page);await refresh(page);}
  else await close(page);
  return value;
}
async function keyAction(page,tenantID,key,name,label,suffix,status=200) {
  await nav(page,'API Keys');await row(page,name).getByRole('button',{name:label,exact:true}).click();
  const value=await save(page,`/admin/api/tenants/${tenantID}/keys/${key}/${suffix}`,'POST',status);
  await page.locator('#editor').waitFor({state:'hidden'});await ready(page);return value;
}
async function download(page,label) {
  const event=page.waitForEvent('download');await button(page,label).click();
  const file=await event;assert.equal(await file.failure(),null);
  const content=await readFile(await file.path(),'utf8');assert.ok(content.length>0);return content;
}
async function prepareVisualActivity(page,key,model,tenantID,billID) {
  const before=await read(page,`workspace/overview?tenant=${tenantID}`);
  assert.equal(before.today,now.slice(0,10));assert.equal(before.period,now.slice(0,7));
  assert.equal(before.requests_today,0);assert.equal(before.trend.length,0);
  const closedBill=await read(page,`bills/${billID}`);
  const actual=await call(page,key,model);
  const detail=await requestDetail(page,actual.id);
  assert.equal(detail.pending,0);assert.equal(detail.period,now.slice(0,7));
  assert.equal(detail.charge_entries.length,1);
  assert.equal(detail.charge_entries[0].fee,'21');assert.equal(detail.charge_entries[0].cost,'7');
  await closeDetail(page);
  const after=await read(page,`workspace/overview?tenant=${tenantID}`);
  assert.equal(after.today,before.today);assert.equal(after.period,before.period);
  assert.equal(after.requests_today,1);assert.equal(after.success_rate_percent,100);
  assert.equal(after.confirmed_fee_micro_usd,'21');assert.equal(after.trend.length,1);
  assert.equal(after.trend[0].day,before.today);assert.equal(after.trend[0].requests,1);assert.equal(after.trend[0].successes,1);
  assert.deepEqual(await read(page,`bills/${billID}`),closedBill);
  await nav(page,'工作台');
  return {requests_before:before.requests_today,requests_after:after.requests_today,trend_days_before:before.trend.length,trend_days_after:after.trend.length,fee_micro_usd:after.confirmed_fee_micro_usd,closed_bill_unchanged:true};
}
async function safeScreenshot(page,name) {
  await nav(page,'API Keys');
  await safeCurrentScreenshot(page,name);
}
async function safeCurrentScreenshot(page,name) {
  assert.ok(await page.evaluate(()=>[...document.querySelectorAll('input[type=password],.secret,textarea[data-sensitive]')].every(element=>element.value==='')));
  assert.ok(await page.evaluate(()=>localStorage.length===0&&sessionStorage.length===0));
  assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  await captureIntegrityScreenshot(page,`test-results/continuous-${width}-${name}-${process.pid}.png`);
}

async function selfHostedJourney(admin) {
  const startedUTC=new Date().toISOString(),{steps,screenshots}=selfHostedProgress;
  const secrets=[process.env.RELAY_JOURNEY_PASSWORD,'controlled-journey-upstream','controlled-rejected-credential','controlled-journey-password'];
  const record=async(step,page)=>{
    mark('self-hosted-'+step);
    const layout=await readabilityEvidence(page);
    assert.equal(await page.evaluate(values=>values.every(value=>!document.body.innerText.includes(value)),secrets),true);
    await safeCurrentScreenshot(page,'self-hosted-'+step);
    screenshots.push(`continuous-${width}-self-hosted-${step}-${process.pid}.png`);
    steps.push({step,utc:new Date().toISOString(),readability:layout});
  };
  const name='Isolated controlled self-hosted service',model='self-hosted-long-public-model-name-readable-without-truncation-'+('segment-'.repeat(7));
  const base=process.env.RELAY_JOURNEY_UPSTREAM;
  mark('self-hosted-create-source');await nav(admin,'供应商与服务');await button(admin,'接入服务').click();
  await admin.getByLabel('服务名称').fill(name);await admin.getByLabel('来源类型').selectOption('self_hosted');
  await admin.getByLabel('兼容API基础地址').fill(new URL('/unreachable',base).toString());
  await admin.getByLabel('上游凭据（编辑留空保留同地址原值）').fill('controlled-journey-upstream');
  const service=await save(admin,'/admin/api/upstreams');assert.equal(service.source_kind,'self_hosted');
  await summaryCell(row(admin,name),'来源','已有自托管');
  const checkSource=async(expectedState,expectedCode)=>{
    mark('self-hosted-check-'+expectedCode);
    const [response]=await Promise.all([admin.waitForResponse(response=>new URL(response.url()).pathname===`/admin/api/upstreams/${service.id}/check`),row(admin,name).getByRole('button',{name:'核验',exact:true}).click()]);
    assert.equal(response.status(),200);const result=await response.json();
    selfHostedFacts={check_state:['success','failed','checking','canceled','unknown'].includes(result.state)?result.state:'other',check_code:['','upstream_connection_failed','upstream_rejected'].includes(result.error_code)?result.error_code:'other',applicable:result.applicable===true};
    assert.equal(result.state,expectedState);assert.equal(result.error_code,expectedCode);assert.equal(result.applicable,true);
    await admin.locator('#detail-content').getByText(expectedState==='success'?'核验通过':'核验失败',{exact:true}).waitFor();
    if(expectedCode)await admin.locator('#detail-content').getByText(expectedCode,{exact:true}).waitFor();
    await record(expectedCode||'verified',admin);await closeDetail(admin);
    await nav(admin,'供应商与服务');await summaryCell(row(admin,name),'核验事实',expectedState==='success'?'核验通过':'核验失败');
    return result;
  };
  const unreachable=await checkSource('failed','upstream_connection_failed');assert.equal(unreachable.upstream_status,null);
  const editSource=async(token)=>{
    await row(admin,name).getByRole('button',{name:'编辑',exact:true}).click();
    await admin.getByLabel('兼容API基础地址').fill(base);
    await admin.getByLabel('上游凭据（编辑留空保留同地址原值）').fill(token);
    await save(admin,`/admin/api/upstreams/${service.id}`,'PATCH',200);
    await summaryCell(row(admin,name),'核验事实','未知 / 配置已变更');
  };
  await editSource('controlled-rejected-credential');const rejected=await checkSource('failed','upstream_rejected');assert.equal(rejected.upstream_status,401);
  await editSource('controlled-journey-upstream');await checkSource('success','');
  await summaryCell(row(admin,name),'基础地址',base);await record('service-long-address',admin);
  mark('self-hosted-associate-model');await nav(admin,'模型托管');await button(admin,'接入模型').click();
  await admin.getByLabel('公共模型名',{exact:true}).fill(model);await admin.getByLabel('服务',{exact:true}).selectOption(String(service.id));
  await admin.getByLabel('远端模型名',{exact:true}).fill('controlled');await save(admin,'/admin/api/models');
  await summaryCell(row(admin,model),'公共模型',model);
  const tenantID=await tenant(admin,'Self-hosted isolated tenant');await nav(admin,'模型托管');
  await row(admin,model).getByRole('button',{name:'租户授权',exact:true}).click();await admin.getByLabel('允许使用此模型的租户').selectOption([tenantID]);
  await save(admin,`/admin/api/models/${model}/tenant-grants`,'PUT',200);
  await nav(admin,'定价与账单');await row(admin,model).getByRole('button',{name:'发布新价',exact:true}).click();
  for(const label of ['供应商输入成本','供应商输出成本','租户输入价格','租户输出价格'])await admin.getByLabel(label+'（微美元 / 百万Token）').fill(label.startsWith('供应商')?'1000000':'3000000');
  await save(admin,`/admin/api/models/${model}/prices`);
  const owner=await invite(admin,tenantID,'self-hosted-owner@example.test','administrator');
  const developer=await invite(admin,tenantID,'self-hosted-developer@example.test','developer');
  await nav(owner,'租户与成员');await summaryCell(row(owner,'self-hosted-developer@example.test').filter({has:owner.getByRole('button',{name:'改角色',exact:true})}),'角色','开发者');await record('roles',owner);
  const key=await createKey(owner,tenantID,'Self-hosted isolated application',[model],'self-hosted-developer@example.test');
  secrets.push(key.key);
  const keys=items(await read(owner,`tenants/${tenantID}/keys`));const metadata=keys.find(value=>value.id===key.id);assert.ok(metadata);
  await summaryCell(row(owner,'Self-hosted isolated application'),'安全标识',metadata.hint);await record('key-safe-identifier',owner);
  for(const [page,stream] of [[owner,false],[developer,true]]) {
    mark(stream?'self-hosted-sse':'self-hosted-normal');
    const before=await read(owner,`tenants/${tenantID}/limits`);
    const response=await call(page,key.key,model,200,'controlled self-hosted input',stream);
    if(stream){assert.ok(response.value.includes('data:'));assert.ok(response.value.includes('[DONE]'));}
    const detail=await requestDetail(page,response.id);assert.equal(detail.state,'success');assert.equal(detail.key_id,key.id);assert.equal(detail.tenant_id,tenantID);assert.equal(detail.alias,model);assert.equal(detail.pending,0);
    assert.equal(detail.confirmed_input_tokens,'3');assert.equal(detail.confirmed_output_tokens,'4');assert.equal(detail.charge_entries.length,1);
    assert.equal(detail.charge_entries[0].fee,'21');assert.equal(detail.charge_entries[0].cost,'7');
    await closeDetail(page);await nav(page,'调用日志');
    const logged=await requestDetail(page,response.id,'records');assert.deepEqual(logged.charge_entries,detail.charge_entries);assert.deepEqual(logged.price_snapshots,detail.price_snapshots);
    await record(stream?'sse-same-request':'normal-same-request',page);await closeDetail(page);
    const after=await read(owner,`tenants/${tenantID}/limits`);assert.equal(BigInt(after.usage.fee_micro_usd)-BigInt(before.usage.fee_micro_usd),21n);assert.equal(after.usage.pending,0);
    steps.push({step:stream?'sse-ledger':'normal-ledger',utc:new Date().toISOString(),input_tokens:3,output_tokens:4,fee_micro_usd:'21',cost_micro_usd:'7',same_request:true});
  }
  await nav(owner,'用量分析');const usage=items(await read(owner,`usage?tenant=${tenantID}`)).find(value=>value.alias===model);assert.ok(usage);
  assert.equal(usage.confirmed_input_tokens,6);assert.equal(usage.confirmed_output_tokens,8);assert.equal(usage.fee_micro_usd,'42');assert.equal(usage.cost_micro_usd,'14');assert.equal(usage.pending,0);
  await summaryCell(row(owner,model),'已确认租户费','0.000042 USD');await record('usage-confirmed',owner);
  await nav(owner,'模型托管');await summaryCell(row(owner,model),'公共模型',model);await record('model-long-name',owner);
  const receipt={suite:'first-self-hosted',width,result:'passed',started_utc:startedUTC,finished_utc:new Date().toISOString(),chromium_version:browser.version(),initial_counts:initialCounts,provider:'controlled_http_not_supplier',installation:'independent_temporary_sqlite',business_writes:'native_user_actions_only',steps,screenshots,cleanup_owner:'go_harness_and_browser_finally'};
  await writeFile(`test-results/self-hosted-${width}-${process.pid}.json`,JSON.stringify(receipt,null,2)+'\n');
  console.log(JSON.stringify({suite:receipt.suite,width,result:'passed',provider:receipt.provider,steps:steps.length}));
}

try {
  await mkdir('test-results',{recursive:true});now=(await control()).now;
  const admin=await newPage();await login(admin,'admin',process.env.RELAY_JOURNEY_PASSWORD);
  await freshInstallation(admin);
  if(mode==='self_hosted') {
    await selfHostedJourney(admin);
  } else {
  const firstTenant=await tenant(admin,'Journey Alpha',true);
  mark('J01-service-validation');await nav(admin,'供应商与服务');await button(admin,'接入服务').click();
  await admin.getByLabel('服务名称').fill('Journey controlled upstream');
  await admin.getByLabel('兼容API基础地址').fill(process.env.RELAY_JOURNEY_UPSTREAM+'?controlled_invalid_destination=1');
  await admin.getByLabel('上游凭据（编辑留空保留同地址原值）').fill('controlled-journey-upstream');
  const rejected=await save(admin,'/admin/api/upstreams','POST',400);
  assert.ok(rejected.error);assert.equal(await admin.getByLabel('服务名称').inputValue(),'Journey controlled upstream');
  assert.equal(await admin.getByLabel('兼容API基础地址').inputValue(),process.env.RELAY_JOURNEY_UPSTREAM+'?controlled_invalid_destination=1');assert.equal(items(await read(admin,'upstreams')).length,0);
  assert.equal(await admin.getByLabel('上游凭据（编辑留空保留同地址原值）').inputValue(),'');
  await admin.getByLabel('来源类型').selectOption('provider');
  await admin.getByLabel('兼容API基础地址').fill(process.env.RELAY_JOURNEY_UPSTREAM);
  await admin.getByLabel('上游凭据（编辑留空保留同地址原值）').fill('controlled-journey-upstream');
  const service=await save(admin,'/admin/api/upstreams');await admin.locator('#editor').waitFor({state:'hidden'});
  const check=admin.waitForResponse(response=>new URL(response.url()).pathname===`/admin/api/upstreams/${service.id}/check`);
  await row(admin,'Journey controlled upstream').getByRole('button',{name:'核验',exact:true}).click();
  const checked=await check;assert.equal(checked.status(),200);const checkedValue=await checked.json();assert.equal(checkedValue.state,'success');assert.equal(checkedValue.applicable,true);await closeDetail(admin);
  const models=['journey-one','journey-two'];
  for(const model of models){
    mark('J01-model-price-grant');await nav(admin,'模型托管');await button(admin,'接入模型').click();
    await admin.getByLabel('公共模型名',{exact:true}).fill(model);await admin.getByLabel('服务',{exact:true}).selectOption(String(service.id));await admin.getByLabel('远端模型名',{exact:true}).fill('controlled');await save(admin,'/admin/api/models');await admin.locator('#editor').waitFor({state:'hidden'});
    await row(admin,model).getByRole('button',{name:'租户授权',exact:true}).click();await admin.getByLabel('允许使用此模型的租户').selectOption([firstTenant]);await save(admin,`/admin/api/models/${model}/tenant-grants`,'PUT',200);await admin.locator('#editor').waitFor({state:'hidden'});
    await row(admin,model).getByRole('button',{name:'上界',exact:true}).click();await admin.getByLabel('服务ID',{exact:true}).fill(String(service.id));await admin.getByLabel('供应商可证明最大收费输入Token').fill('100');await admin.getByLabel('包括隐藏推理等最大收费输出Token').fill('100');await admin.getByLabel('上界依据引用').fill('controlled test contract; not supplier validation');await save(admin,`/admin/api/models/${model}/bounds`,'PUT',200);await admin.locator('#editor').waitFor({state:'hidden'});
    await nav(admin,'定价与账单');await row(admin,model).getByRole('button',{name:'发布新价',exact:true}).click();
    for(const label of ['供应商输入成本','供应商输出成本','租户输入价格','租户输出价格'])await admin.getByLabel(label+'（微美元 / 百万Token）').fill(label.startsWith('供应商')?'1000000':'2000000');
    await save(admin,`/admin/api/models/${model}/prices`);await admin.locator('#editor').waitFor({state:'hidden'});
  }
  await scope(admin,firstTenant);await nav(admin,'模型托管');
  await modelSaveReadbackJourney(admin,{tenant:firstTenant,otherTenant:'',alias:models[1],width,save:path=>save(admin,'/admin/api/'+path,'PATCH',200),ready:()=>ready(admin),changeScope:tenant=>scope(admin,tenant),mark,record:cases=>writeFile(`test-results/model-save-readback-${width}.json`,JSON.stringify({suite:'model-save-readback',mode:'operation_readback_regression',provider:'controlled_http_not_supplier',chromium_version:browser.version(),cases},null,2)+'\n')});
  mark('J01-responsibility-chain');await budget(admin,firstTenant,100000);
  const ownerAccount='journey-owner@example.test',developerAccount='journey-developer@example.test';
  const owner=await invite(admin,firstTenant,ownerAccount,'administrator');
  const developer=await invite(admin,firstTenant,developerAccount,'developer');
  const finance=await invite(admin,firstTenant,'journey-finance@example.test','finance');
  const primary=await createKey(owner,firstTenant,'Primary journey key',models,developerAccount);
  const first=await call(owner,primary.key,models[0]);const original=await requestDetail(owner,first.id);assert.equal(original.tenant_id,firstTenant);assert.equal(original.key_id,primary.id);assert.equal(original.charge_entries[0].fee,'14');assert.equal(original.charge_entries[0].cost,'7');await closeDetail(owner);
  await nav(developer,'API Keys');await row(developer,'Primary journey key').waitFor();assert.equal(await button(developer,'创建Key').count(),0);
  mark('J02-restricted-summary');await nav(developer,'租户与成员');
  await summaryCell(row(developer,'Journey Alpha'),'预算 / 已用','需租户管理员权限');
  const developerTenants=items(await read(developer,'tenants'));assert.ok(developerTenants.every(tenant=>tenant.summary==null));
  await read(developer,`catalog?tenant=${firstTenant}&include_ungranted=true`,403);
  await read(developer,'catalog?tenant=nonmember-summary-scope',404);
  await nav(finance,'定价与账单');assert.equal(await button(finance,'发布新价').count(),0);assert.equal(await button(finance,'结账').count(),0);
  await nav(owner,'租户与成员');await row(owner,ownerAccount).getByRole('button',{name:'移除',exact:true}).click();
  const ownerID=items(await read(owner,`tenants/${firstTenant}/members`)).find(value=>value.account===ownerAccount).id;
  const lastOwner=await save(owner,`/admin/api/tenants/${firstTenant}/members/${ownerID}`,'DELETE',409);assert.equal(lastOwner.error.code,'last_administrator');await close(owner);
  complete('J01');

  mark('J02-intersection');await nav(owner,'模型托管');await row(owner,models[0]).waitFor();await row(owner,models[1]).waitFor();
  await editKey(owner,firstTenant,primary.id,'Primary journey key',[models[1]]);
  const narrowed=await call(owner,primary.key,models[0],403);assert.equal(narrowed.value.error.code,'model_forbidden');
  await editKey(owner,firstTenant,primary.id,'Primary journey key',models);
  const restored=await call(owner,primary.key,models[0]);await requestDetail(owner,restored.id);await closeDetail(owner);
  const oldDetail=await read(owner,`workspace/requests/${first.id}?tenant=${firstTenant}`);assert.deepEqual(oldDetail.price_snapshots,original.price_snapshots);assert.deepEqual(oldDetail.charge_entries,original.charge_entries);
  await safeScreenshot(owner,'before-rotation');complete('J02');

  mark('J03-hour-grace');const beforeRotation=await read(owner,`tenants/${firstTenant}/keys/${primary.id}/limits`);
  let current=await rotation(owner,firstTenant,primary.id,'Primary journey key',1);
  await row(owner,'Primary journey key').getByText('过渡中',{exact:true}).waitFor();
  const metadata=items(await read(owner,`tenants/${firstTenant}/keys`)).find(key=>key.id===primary.id);assert.ok(metadata.generations.some(value=>value.expires&&Date.parse(value.expires)-Date.parse(now)===3600000));
  assert.ok(!JSON.stringify(metadata).includes(current.key));
  const shared=await read(owner,`tenants/${firstTenant}/keys/${primary.id}/limits`);assert.deepEqual(shared.usage,beforeRotation.usage);
  await call(owner,primary.key,models[0]);await call(owner,current.key,models[0]);
  await rotation(owner,firstTenant,primary.id,'Primary journey key',24,409);
  await keyAction(owner,firstTenant,primary.id,'Primary journey key','结束过渡','retire-previous');await call(owner,primary.key,models[0],401);await call(owner,current.key,models[0]);
  const previous=current.key;current=await rotation(owner,firstTenant,primary.id,'Primary journey key',24);
  const grace=items(await read(owner,`tenants/${firstTenant}/keys`)).find(key=>key.id===primary.id);assert.ok(grace.generations.some(value=>value.expires&&Date.parse(value.expires)-Date.parse(now)===86400000));
  await call(owner,previous,models[0]);await call(owner,current.key,models[0]);await rotation(owner,firstTenant,primary.id,'Primary journey key',1,409);
  await keyAction(owner,firstTenant,primary.id,'Primary journey key','结束过渡','retire-previous');await call(owner,previous,models[0],401);
  const revocable=await createKey(owner,firstTenant,'Revoked journey key',models);await keyAction(owner,firstTenant,revocable.id,'Revoked journey key','撤销','revoke');await row(owner,'Revoked journey key').getByText('已撤销',{exact:true}).waitFor();await call(owner,revocable.key,models[0],401);
  await editKey(owner,firstTenant,primary.id,'Primary journey key',models,false);await row(owner,'Primary journey key').getByText('已停用',{exact:true}).waitFor();await call(owner,current.key,models[0],401);await editKey(owner,firstTenant,primary.id,'Primary journey key',models,true);await row(owner,'Primary journey key').getByText('已启用',{exact:true}).waitFor();
  const expires=await createKey(owner,firstTenant,'Expiring journey key',models,null,new Date(Date.parse(now)+60000).toISOString());await advance(60);await refresh(owner);await row(owner,'Expiring journey key').getByText('已过期',{exact:true}).waitFor();await call(owner,expires.key,models[0],401);
  await nav(owner,'API Keys');await row(owner,'Primary journey key').getByRole('button',{name:'历史',exact:true}).click();await ready(owner);const history=await read(owner,`workspace/requests?tenant=${firstTenant}&key=${primary.id}`);assert.ok(history.total>=8);assert.ok(items(history).every(value=>value.key_id===primary.id));
  await safeScreenshot(owner,'key-states');complete('J03');

  mark('J04-money-budget');await scope(admin,firstTenant);const used=await read(admin,`tenants/${firstTenant}/limits`);assert.ok(BigInt(used.usage.fee_micro_usd)>0n);
  await budget(admin,firstTenant,used.usage.fee_micro_usd);
  const exhausted=await call(owner,current.key,models[0],429);assert.equal(exhausted.value.error.code,'tenant_budget_exceeded');
  const preserved=await read(admin,`tenants/${firstTenant}/limits`);assert.equal(preserved.usage.fee_micro_usd,used.usage.fee_micro_usd);
  await budget(admin,firstTenant,100000);const increased=await read(admin,`tenants/${firstTenant}/limits`);assert.equal(increased.usage.fee_micro_usd,used.usage.fee_micro_usd);
  const resent=await call(owner,current.key,models[0]);const resentDetail=await requestDetail(owner,resent.id);assert.equal(resentDetail.charge_entries[0].fee,'14');await closeDetail(owner);complete('J04');

  mark('J05-two-tenants');const secondTenant=await tenant(admin,'Journey Beta');await budget(admin,secondTenant,100000);
  for(const model of models){await nav(admin,'模型托管');await row(admin,model).getByRole('button',{name:'租户授权',exact:true}).click();await admin.getByLabel('允许使用此模型的租户').selectOption([firstTenant,secondTenant]);await save(admin,`/admin/api/models/${model}/tenant-grants`,'PUT',200);await admin.locator('#editor').waitFor({state:'hidden'});}
  const secondOwner=await invite(admin,secondTenant,'journey-beta-owner@example.test','administrator');
  const secondKey=await createKey(secondOwner,secondTenant,'Beta journey key',models);await call(secondOwner,secondKey.key,models[0]);
  const standby=await createKey(owner,firstTenant,'Standby journey key',models);await call(owner,standby.key,models[0]);
  await scope(admin,firstTenant);const pauseUsage=await read(admin,`tenants/${firstTenant}/limits`);
  for(const active of [false,true]){
    mark(active?'J05-resume':'J05-pause');await nav(admin,'租户与成员');await row(admin,'Journey Alpha').getByRole('button',{name:'暂停 / 恢复',exact:true}).click();await admin.getByLabel('启用新调用').setChecked(active);await save(admin,`/admin/api/tenants/${firstTenant}`,'PATCH',200);await admin.locator('#editor').waitFor({state:'hidden'});
    await row(admin,'Journey Alpha').getByText(active?'活跃':'暂停',{exact:true}).waitFor();
    if(!active){
      const paused=await call(owner,current.key,models[0],403);assert.equal(paused.value.error.code,'tenant_paused');
      mark('J05-paused-ordinary-usage');assertUsageConserved(pauseUsage,await read(admin,`tenants/${firstTenant}/limits`));
      await call(owner,current.key,models[0],403,'controlled paused stream',true);
      mark('J05-paused-stream-usage');assertUsageConserved(pauseUsage,await read(admin,`tenants/${firstTenant}/limits`));
      await call(owner,standby.key,models[0],403);
      mark('J05-paused-standby-usage');assertUsageConserved(pauseUsage,await read(admin,`tenants/${firstTenant}/limits`));
      await call(secondOwner,secondKey.key,models[0]);
      mark('J05-other-tenant-usage');assertUsageConserved(pauseUsage,await read(admin,`tenants/${firstTenant}/limits`));
    }
    else {const resumedUsage=await read(admin,`tenants/${firstTenant}/limits`);assertUsageConserved(pauseUsage,resumedUsage);assert.equal(resumedUsage.usage.fee_micro_usd,pauseUsage.usage.fee_micro_usd);await call(owner,current.key,models[0]);await call(owner,standby.key,models[0]);}
  }
  mark('J05-member-lookup');await nav(admin,'租户与成员');const memberID=items(await read(admin,`tenants/${firstTenant}/members`)).find(value=>value.account===developerAccount).id;
  mark('J05-role-change');
  await row(admin,developerAccount).getByRole('button',{name:'改角色',exact:true}).click();await admin.getByLabel('角色',{exact:true}).selectOption('finance');await save(admin,`/admin/api/tenants/${firstTenant}/members/${memberID}`,'PATCH',200);await admin.locator('#editor').waitFor({state:'hidden'});
  mark('J05-role-readback');const freshRole=items(await read(developer,'tenants')).find(value=>value.id===firstTenant);membershipFacts={role:['administrator','developer','finance'].includes(freshRole?.role)?freshRole.role:'other'};assert.equal(freshRole.role,'finance');
  mark('J05-diagnostic-recovery');const diagnosticMembers=items(await read(admin,`tenants/${firstTenant}/keys`)).find(value=>value.id===primary.id).diagnostic_members;membershipFacts.diagnostic_count=Array.isArray(diagnosticMembers)?diagnosticMembers.length:null;assert.ok(diagnosticMembers.length===0);
  mark('J05-member-remove');
  await row(admin,developerAccount).getByRole('button',{name:'移除',exact:true}).click();await save(admin,`/admin/api/tenants/${firstTenant}/members/${memberID}`,'DELETE',200);await admin.locator('#editor').waitFor({state:'hidden'});
  mark('J05-removed-key-scope');await read(developer,`tenants/${firstTenant}/keys/${primary.id}/limits`,404);
  mark('J05-removed-membership');const remainingTenants=items(await read(developer,'tenants'));membershipFacts.tenant_count=remainingTenants.length;assert.equal(remainingTenants.length,0);
  mark('J05-removed-ui-scope');await refresh(developer,0);membershipFacts.tenant_option_count=await developer.getByLabel('当前租户').locator('option').count();assert.equal(membershipFacts.tenant_option_count,0);
  mark('J05-application-survives');await call(owner,current.key,models[0]);
  mark('J05-other-tenant-request');await read(secondOwner,`workspace/requests/${first.id}?tenant=${secondTenant}`,404);complete('J05');

  mark('J07-native-pagination');await budget(admin,firstTenant,1000000);
  for(let index=0;index<27;index++)await call(owner,current.key,models[0]);
  await nav(owner,'API Keys');await row(owner,'Primary journey key').getByRole('button',{name:'历史',exact:true}).click();await ready(owner);
  assert.equal(await owner.getByLabel('Key ID',{exact:true}).inputValue(),primary.id);await button(owner,'下一页').click();await owner.getByText('第2页 ·',{exact:false}).waitFor();
  await owner.getByLabel('公共模型',{exact:true}).fill(models[0]);await button(owner,'筛选').click();await ready(owner);assert.equal(await owner.getByLabel('Key ID',{exact:true}).inputValue(),primary.id);
  const csv=await download(owner,'导出筛选');assert.ok(csv.includes(firstTenant));assert.ok(!csv.includes(secondTenant));
  const recent=await call(owner,current.key,models[0]);await requestDetail(owner,recent.id);await button(owner,'在当前范围查看日志').click();await ready(owner);assert.equal(await owner.getByLabel('Request ID',{exact:true}).inputValue(),recent.id);assert.equal(await owner.getByLabel('Key ID',{exact:true}).inputValue(),primary.id);assert.equal(await owner.getByLabel('公共模型',{exact:true}).inputValue(),models[0]);assert.equal(await owner.getByLabel('当前租户').inputValue(),firstTenant);await row(owner,recent.id).waitFor();
  await scope(admin,firstTenant);await nav(admin,'调用日志');
  let release,arrived,finished;const held=new Promise(resolve=>{release=resolve;}),observed=new Promise(resolve=>{arrived=resolve;}),settled=new Promise(resolve=>{finished=resolve;});let injected=false;
  const recordsRoute='**/admin/api/workspace/requests?*';
  await admin.route(recordsRoute,async route=>{const url=new URL(route.request().url());if(!injected&&url.searchParams.get('tenant')===firstTenant){injected=true;const response=await route.fetch();arrived();await held;try{await route.fulfill({response});}catch{}finally{finished();}}else await route.continue();});
  await button(admin,'刷新当前事实').click();await observed;
  const secondRecords=admin.waitForResponse(response=>{const url=new URL(response.url());return url.pathname==='/admin/api/workspace/requests'&&url.searchParams.get('tenant')===secondTenant;});
  await scope(admin,secondTenant);const secondRecordsResponse=await secondRecords;assert.equal(secondRecordsResponse.status(),200);const expectedSecondRecords=items(await secondRecordsResponse.json());release();await settled;await ready(admin);
  await admin.waitForFunction(()=>!document.getElementById('content').textContent.includes('正在读取当前范围…'));
  assert.equal(await admin.getByLabel('当前租户').inputValue(),secondTenant);assert.equal(await row(admin,first.id).count(),0);
  assert.equal(await admin.locator('#content tbody tr').count(),expectedSecondRecords.length);for(const value of expectedSecondRecords)await row(admin,value.id).waitFor();assert.ok(expectedSecondRecords.every(value=>value.tenant_id===secondTenant));await admin.unroute(recordsRoute);
  let failed=false;await admin.route(recordsRoute,async route=>{if(!failed){failed=true;await route.fulfill({status:503,contentType:'application/json',body:'{"error":{"code":"controlled_read_failure"}}'});}else await route.continue();});
  await button(admin,'刷新当前事实').click();await admin.getByText('读取失败：',{exact:false}).waitFor();assert.equal(await button(admin,'下一页').count(),0);await button(admin,'重试读取').click();await button(admin,'下一页').waitFor();await admin.unroute(recordsRoute);
  await call(owner,current.key,models[0]);await owner.reload();await owner.locator('#workspace').waitFor();await ready(owner);assert.equal(await owner.locator('#call-key').count(),0);assert.ok(await owner.evaluate(()=>[...document.querySelectorAll('input[type=password],.secret,textarea[data-sensitive]')].every(value=>value.value==='')));
  await button(owner,'退出').click();await owner.locator('#login-panel').waitFor();await read(owner,'session',401);await login(owner,ownerAccount,'controlled-journey-password');complete('J07');

  mark('J06-new-price-snapshot');await advance(1);await scope(admin,firstTenant);await nav(admin,'定价与账单');await row(admin,models[0]).getByRole('button',{name:'发布新价',exact:true}).click();
  for(const label of ['供应商输入成本','供应商输出成本','租户输入价格','租户输出价格'])await admin.getByLabel(label+'（微美元 / 百万Token）').fill(label.startsWith('供应商')?'1000000':'3000000');
  await save(admin,`/admin/api/models/${models[0]}/prices`);await admin.locator('#editor').waitFor({state:'hidden'});
  const newPrice=await call(owner,current.key,models[0]);const newDetail=await requestDetail(owner,newPrice.id);assert.equal(newDetail.charge_entries[0].fee,'21');await closeDetail(owner);assert.deepEqual((await read(owner,`workspace/requests/${first.id}?tenant=${firstTenant}`)).price_snapshots,original.price_snapshots);
  const unknown=await call(owner,current.key,models[0],200,'controlled unknown usage');const pending=await requestDetail(owner,unknown.id);assert.equal(pending.pending,1);await closeDetail(owner);
  const period=now.slice(0,7);await nav(admin,'定价与账单');await row(admin,period).getByRole('button',{name:'结账',exact:true}).click();await save(admin,`/admin/api/tenants/${firstTenant}/bills`,'POST',409);await close(admin);
  const nextMonth=new Date(now);nextMonth.setUTCMonth(nextMonth.getUTCMonth()+1,1);nextMonth.setUTCHours(0,0,0,0);await advance(Math.ceil((nextMonth.getTime()-Date.parse(now))/1000));
  await login(admin,'admin',process.env.RELAY_JOURNEY_PASSWORD);await scope(admin,firstTenant);await nav(admin,'定价与账单');await row(admin,period).getByRole('button',{name:'结账',exact:true}).click();const pendingClose=await save(admin,`/admin/api/tenants/${firstTenant}/bills`,'POST',409);assert.equal(pendingClose.error.code,'pending_usage');await close(admin);
  mark('J06-reconciliation');await nav(admin,'调用日志');await admin.getByLabel('Request ID',{exact:true}).fill(unknown.id);await button(admin,'筛选').click();await ready(admin);
  await summaryCell(row(admin,unknown.id),'Token总量','未知');await summaryCell(row(admin,unknown.id),'已确认计费Token（输入 / 输出）','0 / 0');await summaryCell(row(admin,unknown.id),'待确认','1');
  await row(admin,unknown.id).getByRole('button',{name:'详情 / 核对',exact:true}).click();await button(admin,'核对尝试用量').click();await admin.getByLabel('尝试序号').fill('1');await admin.getByLabel('权威累计输入Tokens').fill('3');await admin.getByLabel('权威累计输出Tokens').fill('4');await admin.getByLabel('原因',{exact:true}).fill('controlled authoritative test usage');await admin.getByLabel('可追溯证据引用').fill('controlled fixture HTTP usage contract; not supplier validation');await save(admin,`/admin/api/workspace/requests/${unknown.id}/reconciliations/1`,'POST',200);await admin.locator('#editor').waitFor({state:'hidden'});
  const reconciled=await read(admin,`workspace/requests/${unknown.id}?tenant=${firstTenant}`);assertReconciledLedger(pending,reconciled);
  mark('J06-reconciled-token-meaning');
  assert.equal(pending.total_tokens,null);assert.equal(pending.confirmed_input_tokens,'0');assert.equal(pending.confirmed_output_tokens,'0');
  assert.equal(reconciled.total_tokens,null);assert.equal(reconciled.confirmed_input_tokens,'3');assert.equal(reconciled.confirmed_output_tokens,'4');
  await nav(admin,'调用日志');
  await summaryCell(row(admin,unknown.id),'Token总量','未知');
  await summaryCell(row(admin,unknown.id),'已确认计费Token（输入 / 输出）','3 / 4');
  await summaryCell(row(admin,unknown.id),'待确认','0');
  const confirmedDetail=await requestDetail(admin,unknown.id,'records');assert.equal(confirmedDetail.confirmed_input_tokens,'3');assert.equal(confirmedDetail.confirmed_output_tokens,'4');
  for(const [label,value] of [['原始上游总Token','未知'],['已确认计费输入Token','3'],['已确认计费输出Token','4']])assert.equal(await admin.locator('#detail-content').getByText(label,{exact:true}).first().evaluate(element=>element.nextElementSibling.textContent),value);
  await closeDetail(admin);
  const confirmedCSV=await download(admin,'导出筛选'),csvLines=confirmedCSV.trim().split(/\r?\n/).map(line=>line.split(','));assert.equal(csvLines.length,2);
  for(const [name,value] of [['total_tokens',''],['confirmed_input_tokens','3'],['confirmed_output_tokens','4'],['pending','0']]){const index=csvLines[0].indexOf(name);assert.ok(index>=0);assert.equal(csvLines[1][index],value);}
  await nav(admin,'用量分析');await admin.getByLabel('Request ID',{exact:true}).fill(unknown.id);await button(admin,'筛选').click();await ready(admin);
  await summaryCell(row(admin,models[0]),'已确认输入Tokens','3');await summaryCell(row(admin,models[0]),'已确认输出Tokens','4');
  mark('J06-close-correction');await nav(admin,'定价与账单');await row(admin,period).getByRole('button',{name:'结账',exact:true}).click();const bill=await save(admin,`/admin/api/tenants/${firstTenant}/bills`);await admin.locator('#editor').waitFor({state:'hidden'});
  const before=await read(admin,`bills/${bill.id}`);await safeCurrentScreenshot(admin,'before-correction');await billDetail(admin,period,bill.id);await button(admin,'追加更正').click();await admin.getByLabel('有符号更正金额（微美元）').fill('5');await admin.getByLabel('开放调整周期（UTC月）').fill(now.slice(0,7));await admin.getByLabel('原因',{exact:true}).fill('controlled append correction');await save(admin,`/admin/api/bills/${bill.id}/corrections`);await admin.locator('#editor').waitFor({state:'hidden'});
  const after=await read(admin,`bills/${bill.id}`);assert.equal(after.original_fee,before.original_fee);assert.equal(after.corrections.length,1);assert.equal(after.corrections[0].amount,'5');assert.equal(after.corrections[0].period,now.slice(0,7));assert.equal(after.corrections[0].reason,'controlled append correction');
  const adjustedPeriod=items(await read(admin,`tenants/${firstTenant}/bills`)).find(value=>value.period===now.slice(0,7));assert.equal(adjustedPeriod.adjustment_micro_usd,'5');assert.equal(adjustedPeriod.fee_micro_usd,'5');await row(admin,now.slice(0,7)).getByText('0.000005 USD',{exact:true}).waitFor();
  await billDetail(admin,period,bill.id);await safeCurrentScreenshot(admin,'after-correction');const billCSV=await download(admin,'导出周期完整账单');assert.ok(billCSV.includes(period));assert.ok(!billCSV.includes(secondTenant));await closeDetail(admin);
  await login(finance,'journey-finance@example.test','controlled-journey-password');await nav(finance,'定价与账单');assert.equal(await button(finance,'追加更正').count(),0);assert.equal(await button(finance,'结账').count(),0);
  await login(secondOwner,'journey-beta-owner@example.test','controlled-journey-password');await read(secondOwner,`bills/${bill.id}`,404);complete('J06');
  mark('visual-current-period-call');
  const visualActivity=await prepareVisualActivity(admin,current.key,models[0],firstTenant,bill.id);
  await writeFile(`test-results/visual-activity-${width}.json`,JSON.stringify({suite:'visual-current-period',width,result:'passed',provider:'controlled_http_not_supplier',facts:visualActivity},null,2)+'\n');
  console.log(JSON.stringify({suite:'visual-current-period',width,result:'passed',facts:visualActivity,provider:'controlled_http_not_supplier'}));
  const summaryFacts=await managementSummaryJourney(admin,{tenants:[firstTenant,secondTenant],models,nav:title=>nav(admin,title),read:path=>read(admin,path),scope:tenant=>scope(admin,tenant),save:path=>save(admin,'/admin/api/'+path,'PATCH',200),closeDetail:()=>closeDetail(admin),ready:()=>ready(admin),mark});
  await writeFile(`test-results/management-summary-${width}.json`,JSON.stringify({suite:'management-summary',width,result:'passed',provider:'controlled_http_not_supplier',facts:summaryFacts},null,2)+'\n');
  console.log(JSON.stringify({suite:'management-summary',width,result:'passed',provider:'controlled_http_not_supplier',facts:summaryFacts}));
  await consoleVisualReview(admin,{width,tenant:firstTenant,objects:{model:models[0],service:'Journey controlled upstream',key:'Primary journey key',request:unknown.id},secrets:[current.key,primary.key,process.env.RELAY_JOURNEY_PASSWORD,'controlled-journey-upstream','controlled-journey-password'],nav:title=>nav(admin,title),step:phase=>{activePage=admin;mark('visual-'+phase);}});
  await safeScreenshot(admin,'final');assert.equal(checkpoints.length,7);
  if(width===1280||width===390)await latencyJourney(admin,{width,tenant:firstTenant,upstreams:[new URL(process.env.RELAY_JOURNEY_UPSTREAM).origin+'/latency-a/v1',new URL(process.env.RELAY_JOURNEY_UPSTREAM).origin+'/latency-b/v1'],credential:'controlled-journey-upstream',password:process.env.RELAY_JOURNEY_PASSWORD,restart:async()=>{const response=await fetch(process.env.RELAY_JOURNEY_CONTROL+'/restart',{method:'POST',headers:{'X-Test-Capability':process.env.RELAY_JOURNEY_CAPABILITY}});assert.equal(response.status,200);return (await response.json()).origin;}});
  await writeFile(`test-results/continuous-${width}-${process.pid}.json`,JSON.stringify({suite:'continuous-console',width,result:'passed',checkpoints,provider:'controlled_http_not_supplier',clock:'explicit_test_clock_no_database_edits',cleanup:'isolated_directory_servers_contexts'},null,2)+'\n');
  }
} catch(error) {
  const snapshot=activePage?await activePage.evaluate(()=>({editor_open:document.getElementById('editor')?.open===true,editor_invalid:document.getElementById('editor')?.querySelectorAll(':invalid').length??0,editor_failure:document.getElementById('editor-error')?.textContent.startsWith('操作失败：')===true,content_loading:document.getElementById('content')?.textContent.includes('正在读取当前范围…')===true,workspace_visible:document.getElementById('workspace')?.hidden===false,secret_present:[...document.querySelectorAll('input[type=password],.secret')].some(element=>element.value!=='')})).catch(()=>({snapshot_available:false})):{};
  const receipt={suite:mode==='self_hosted'?'first-self-hosted':'continuous-console',width,result:'failed',stage,action,last_status:lastStatus,initial_counts:initialCounts,call_facts:callFacts,usage_facts:usageFacts,read_facts:readFacts,membership_facts:membershipFacts,ledger_facts:ledgerFacts,self_hosted_facts:selfHostedFacts,self_hosted_progress:selfHostedProgress,failure_boundary:failureBoundary(error),summary_cell:summaryFailureFacts(error),readability_facts:readabilityFailureFacts(error),checkpoints,error_kind:error.name==='TimeoutError'?'timeout':error.name==='AssertionError'?'assertion':'other',...snapshot};
  console.error(JSON.stringify(receipt));await writeFile(`test-results/${mode==='self_hosted'?'self-hosted':'continuous'}-${width}-${process.pid}.json`,JSON.stringify(receipt,null,2)+'\n');process.exitCode=1;
} finally {for(const context of contexts)await context.close();await browser.close();}
