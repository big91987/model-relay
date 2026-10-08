import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {mkdir, readFile, writeFile} from 'node:fs/promises';

export const visualPages = [
  {id:'overview',title:'工作台',text:['今日调用','今日成功率','本月已确认费用','已结束调用平均延迟','近7日调用趋势','租户费用分布','待确认与保守预留','当前范围模型状态'],actions:['发起调用','查看请求与核对','查看用量分析','管理模型','查看预算']},
  {id:'models',minimumRows:2,title:'模型托管',columns:['公共模型','来源 / 类型','协议能力','状态','当前租户价格（USD）','调用延迟','版本'],actions:['接入模型'],row:'model',rowActions:['编辑','租户授权','上界','详情','启停','调用测试'],detail:{action:'详情',text:['路由候选','价格版本','上界依据']},editor:{action:'编辑',labels:['公共模型名','服务','远端模型名','首候选优先级（越大越优先）','首候选权重','首候选启用','启用']}},
  {id:'services',title:'供应商与服务',columns:['服务','来源','基础地址','状态','关联模型','核验事实','核验时间 / 配置版本'],actions:['接入服务'],row:'service',rowActions:['编辑','核验','详情','关联模型'],detail:{action:'详情',text:['名称','基础地址','启用']},editor:{action:'编辑',labels:['服务名称','来源类型','兼容API基础地址','上游凭据（编辑留空保留同地址原值）']}},
  {id:'keys',minimumRows:2,title:'API Keys',columns:['名称','应用','安全标识','模型授权','状态','期限','费用已用 / 预留 / 上限'],actions:['创建Key'],row:'key',rowActions:['详情','额度 / 用量','历史','编辑','轮换','撤销'],detail:{action:'额度 / 用量',text:['本周期次数','每分钟请求 RPM','每分钟收费Token TPM（含预留）','本地活跃并发','费用预算（USD）']},editor:{action:'编辑',labels:['名称','应用','模型授权（可多选）','诊断开发者（可多选）','UTC到期时间（留空为无限）','启用']}},
  {id:'tenants',minimumRows:2,title:'租户与成员',columns:['租户','状态','人员资格','负责人','成员数','授权模型','应用Key数（含撤销历史）','预算 / 已用','账号','角色','到期'],text:['当前租户成员','邀请记录（未发送邮件）'],actions:['创建租户','邀请成员'],editor:{action:'邀请成员',labels:['邮箱账号','角色','有效小时（1–168）']}},
  {id:'limits',title:'配额与预算',columns:['独立维度','有效上限','已用 / 含预留','剩余','平台上界'],text:['租户共享资源','本周期次数','每分钟请求 RPM','每分钟收费Token TPM（含预留）','本地活跃并发','费用预算（USD）'],actions:['调整租户限额','调整Key限额'],editor:{location:'heading',action:'调整租户限额',labels:['本周期次数','每分钟请求 RPM','每分钟收费Token TPM','本地并发','费用预算（微美元）','硬费用预算（需要供应商可信收费Token上界）']}},
  {id:'usage',title:'用量分析',columns:['模型','请求','已确认输入Tokens','已确认输出Tokens','已确认租户费','已确认供应商成本','待确认','保守预留'],actions:['筛选','导出当前筛选'],filters:true},
  {id:'billing',title:'定价与账单',columns:['模型','租户输入 USD/百万Token','租户输出 USD/百万Token','周期','原 / 已确认费用','成本','待确认'],text:['模型价格（USD测试价）','租户周期账单 · 外部支付未接通'],row:'model',rowActions:['版本','发布新价'],detail:{action:'版本',text:['供应商输入成本（微美元 / 百万Token）','供应商输出成本（微美元 / 百万Token）','租户输入价格（微美元 / 百万Token）','租户输出价格（微美元 / 百万Token）']},editor:{action:'发布新价',labels:['服务ID','供应商输入成本（微美元 / 百万Token）','供应商输出成本（微美元 / 百万Token）','租户输入价格（微美元 / 百万Token）','租户输出价格（微美元 / 百万Token）']}},
  {id:'records',title:'调用日志',columns:['时间','Request ID','公共模型','状态','Token总量','已确认租户费','已确认供应商成本','耗时ms','待确认'],actions:['筛选','管理审计','导出筛选','上一页','下一页'],filters:true,row:'request',detail:{action:'详情 / 核对',text:['调用尝试','受理价格快照','追加计费明细']}},
  {id:'playground',title:'调用测试',actions:['发送','停止本地接收（费用可能待确认）','查看同Request ID日志'],labels:['应用API Key（不会保存）','公共模型名','消息','SSE流式']}
];

export function assertCaptureFacts(facts) {
  assert.equal(facts.title,true,'exact current page title');
  assert.equal(facts.scope,true,'current tenant scope');
  assert.equal(facts.role,true,'server-authenticated platform administrator');
  assert.equal(facts.ready,true,'no loading/failure/empty masquerading as data');
  assert.equal(facts.rootFits,true,'no root horizontal overflow');
  assert.equal(facts.secretsClear,true,'sensitive controls and output cleared before capture');
  for(const group of ['columns','text','actions','labels']) {
    for(const [label,present] of Object.entries(facts[group])) assert.equal(present,true,`${group}: ${label}`);
  }
  assert.ok(facts.dataRows>0||facts.form||facts.cards,'actual rows or applicable summary/form');
  if(facts.minimumRows) assert.ok(facts.primaryRows>=facts.minimumRows,'multi-object primary list required');
}

export function visualEditorControl(page,definition,row) {
  const owner=definition.editor.location==='heading'?page.locator('.page-heading'):row;
  return owner.getByRole('button',{name:definition.editor.action,exact:true});
}

export function overviewReadFact(response,tenant) {
  const address=new URL(response.url());
  if(response.request().method()!=='GET'||address.pathname!=='/admin/api/workspace/overview') return null;
  return {resource:'workspace/overview',status:response.status(),scope_matches:address.searchParams.get('tenant')===tenant};
}

export async function consoleVisualReview(page,{width,tenant,objects,secrets=[],nav,step}) {
  const directory=`test-results/visual-${width}`;
  await mkdir(directory,{recursive:true});
  const receipt={suite:'console-visual',width,role:'platform_administrator',scope:{kind:'tenant',id:tenant},provider:'controlled_http_not_supplier',scenario:'continuous_J01-J07',clock:'explicit_test_clock_no_database_edits',pages:[],phase:'start',result:'in_progress'};
  const mark=phase=>{receipt.phase=phase;step(phase);};
  const saveReceipt=()=>writeFile(`${directory}/review.json`,JSON.stringify(receipt,null,2)+'\n');
  let captureAssertion=null;
  const observe=response=>{
    const fact=overviewReadFact(response,tenant);
    if(fact&&receipt.pending_capture) {
      if(receipt.pending_capture.http_reads.length<16) receipt.pending_capture.http_reads.push(fact);
      else receipt.pending_capture.omitted_reads++;
    }
  };
  await saveReceipt();
  const screenshot=async name=>{
    receipt.pending_capture.check='screenshot-secret-clear';
    const clear=await page.evaluate(secrets=>[...document.querySelectorAll('input[type=password],.secret,textarea[data-sensitive]')].every(control=>control.value==='')&&(!document.getElementById('call-output')||['','尚未调用'].includes(document.getElementById('call-output').textContent))&&secrets.filter(Boolean).every(secret=>!document.body.textContent.includes(secret)),secrets);
    assert.equal(clear,true,'never capture passwords, keys, invitation links or conversations');
    await page.evaluate(()=>document.fonts.ready);
    const path=`${directory}/${name}.png`;
    await page.screenshot({path,fullPage:true});
    return {path,sha256:createHash('sha256').update(await readFile(path)).digest('hex')};
  };
  page.on('response',observe);
  try {
  for(const definition of visualPages) {
    mark(definition.id);
    captureAssertion=null;
    receipt.pending_capture={page:definition.id,expected_title:definition.title,check:'navigation',http_reads:[],omitted_reads:0,facts:null};
    await saveReceipt();
    await nav(definition.title);
    const facts=await page.evaluate(({definition,tenant})=>{
      const content=document.getElementById('content');
      const visible=element=>!!element&&element.getClientRects().length>0;
      const names=selector=>[...content.querySelectorAll(selector)].filter(visible).map(element=>element.textContent.trim());
      const columns=names('th'),text=content.textContent;
      const buttons=[...document.querySelectorAll('#content button,#page-action')].filter(visible);
      const labels=[...content.querySelectorAll('label,input[aria-label]')].filter(visible).map(element=>element.getAttribute('aria-label')??element.textContent.trim());
      const empty=[...content.querySelectorAll('.empty')].filter(visible);
      const knownEmpty=['当前范围近7日暂无调用','当前范围暂无数据'];
      return {
        title:document.getElementById('page-title').textContent===definition.title,
        scope:document.getElementById('tenant').value===tenant,
        role:document.getElementById('identity').textContent.endsWith(' · 平台管理员'),
        ready:!text.includes('正在读取当前范围…')&&!content.querySelector('.error')&&!content.querySelector('.empty'),
        settlement:{loading:text.includes('正在读取当前范围…'),error_nodes:content.querySelectorAll('.error').length,notice_present:!!document.getElementById('notice')?.textContent.trim()},
        empty_states:{total:content.querySelectorAll('.empty').length,visible:empty.length,seven_day_no_calls:empty.filter(element=>element.textContent.trim()===knownEmpty[0]).length,no_data:empty.filter(element=>element.textContent.trim()===knownEmpty[1]).length,other:empty.filter(element=>!knownEmpty.includes(element.textContent.trim())).length},
        dimensions:{root_scroll_width:document.documentElement.scrollWidth,viewport_width:innerWidth,content_width:content.getBoundingClientRect().width},
        identity:{platform_administrator:document.getElementById('identity').textContent.endsWith(' · 平台管理员'),tenant_selected:!!document.getElementById('tenant').value,tenant_option_count:document.getElementById('tenant').options.length},
        rootFits:document.documentElement.scrollWidth<=innerWidth,
        secretsClear:[...document.querySelectorAll('input[type=password],.secret,textarea[data-sensitive]')].every(control=>control.value===''),
        columns:Object.fromEntries((definition.columns??[]).map(label=>[label,columns.includes(label)])),
        text:Object.fromEntries((definition.text??[]).map(label=>[label,text.includes(label)])),
        actions:Object.fromEntries((definition.actions??[]).map(label=>[label,buttons.some(element=>element.textContent.trim()===label)])),
        action_states:Object.fromEntries((definition.actions??[]).map(label=>[label,buttons.some(element=>element.textContent.trim()===label&&!element.disabled)?'enabled':'disabled_or_absent'])),
        labels:Object.fromEntries((definition.labels??[]).map(label=>[label,labels.includes(label)])),
        minimumRows:definition.minimumRows??0,primaryRows:content.querySelector('tbody')?.querySelectorAll('tr').length??0,
        dataRows:content.querySelectorAll('tbody tr').length,form:!!content.querySelector('#call-form'),cards:content.querySelectorAll('.card').length
      };
    },{definition,tenant});
    receipt.pending_capture.facts=facts;
    receipt.pending_capture.check='capture-facts';
    await saveReceipt();
    try { assertCaptureFacts(facts); }
    catch(error) { captureAssertion=error.message.split('\n')[0]; throw error; }
    if(definition.filters) for(const label of ['公共模型','Key ID','终态','Request ID','起始UTC时间','截止UTC时间']) {
      receipt.pending_capture.check='filter: '+label;
      await page.getByRole('textbox',{name:label,exact:true}).waitFor({state:'visible'});
    }
    receipt.pending_capture.check='exact-current-object-row';
    const rowText=objects[definition.row];
    if(definition.row) assert.ok(rowText,'exact live object required');
    const row=rowText?page.getByRole('row').filter({has:page.getByRole('cell',{name:rowText,exact:true})}):page.locator('#content');
    if(rowText)assert.equal(await row.count(),1,'exact current fixture row');
    for(const action of definition.rowActions??[]) {
      receipt.pending_capture.check='row-action: '+action;
      const control=row.getByRole('button',{name:action,exact:true});
      await control.waitFor({state:'visible'});
      assert.equal(await control.isDisabled(),false,'current row operation is enabled');
    }
    if(definition.id==='playground') {
      receipt.pending_capture.check='unused-call-form';
      assert.equal(await page.getByRole('button',{name:'发送',exact:true}).isDisabled(),false);
      for(const name of ['停止本地接收（费用可能待确认）','查看同Request ID日志']) assert.equal(await page.getByRole('button',{name,exact:true}).isDisabled(),true);
      assert.equal(await page.locator('#call-output').textContent(),'尚未调用');
    }
    const checkpoint={page:definition.id,title:definition.title,state:'nonempty_or_unused_call_form',facts,row_actions:definition.rowActions??[],screenshots:[await screenshot(`${definition.id}-page`)]};
    receipt.pages.push(checkpoint);
    await saveReceipt();
    if(definition.detail) {
      mark(`${definition.id}-detail`);
      receipt.pending_capture.check='open-detail';
      await row.getByRole('button',{name:definition.detail.action,exact:true}).first().click();
      await page.locator('#detail[open]').waitFor({state:'visible'});
      for(const label of definition.detail.text) {
        receipt.pending_capture.check='detail-field: '+label;
        await page.locator('#detail-content').getByText(label,{exact:true}).first().waitFor({state:'visible'});
      }
      checkpoint.detail={action:definition.detail.action,asserted_fields:definition.detail.text,screenshot:await screenshot(`${definition.id}-detail`)};
      await page.getByRole('button',{name:'关闭详情',exact:true}).click();
      await page.locator('#detail').waitFor({state:'hidden'});
      await saveReceipt();
    }
    if(definition.editor) {
      mark(`${definition.id}-editor`);
      receipt.pending_capture.check='open-editor';
      const control=visualEditorControl(page,definition,row);
      assert.equal(await control.count(),1,'exact editor action in its owning surface');
      await control.click();
      await page.locator('#editor[open]').waitFor({state:'visible'});
      for(const label of definition.editor.labels) {
        receipt.pending_capture.check='editor-field: '+label;
        await page.getByLabel(label,{exact:true}).waitFor({state:'visible'});
      }
      checkpoint.editor={action:definition.editor.action,asserted_labels:definition.editor.labels,screenshot:await screenshot(`${definition.id}-editor`),mutation:false};
      await page.getByRole('button',{name:'关闭编辑',exact:true}).click();
      await page.locator('#editor').waitFor({state:'hidden'});
      await saveReceipt();
    }
  }
  } catch(error) {
    receipt.result='failed';
    receipt.failure={phase:receipt.phase,check:receipt.pending_capture?.check,kind:error.name==='AssertionError'?'assertion':error.name==='TimeoutError'?'timeout':'other',capture_assertion:captureAssertion};
    await saveReceipt();
    console.log(JSON.stringify({suite:receipt.suite,width,result:'failed',phase:receipt.phase,failure:receipt.failure,evidence:`${directory}/review.json`,capture:receipt.pending_capture}));
    throw error;
  } finally {
    page.off('response',observe);
  }
  receipt.result='passed';
  receipt.phase='completed';
  delete receipt.pending_capture;
  await saveReceipt();
  console.log(JSON.stringify({suite:receipt.suite,width,result:receipt.result,pages:receipt.pages.length,evidence:`${directory}/review.json`,provider:receipt.provider,scope:'current_authenticated_tenant'}));
}
