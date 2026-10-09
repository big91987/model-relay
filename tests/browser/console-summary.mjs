import assert from 'node:assert/strict';

const summaryFailures=new WeakMap();

export function summaryFailureFacts(error) {
  return summaryFailures.get(error)??{available:false};
}

export function summaryMoney(value) {
  if(value==null)return '未知';
  const amount=BigInt(value),absolute=amount<0n?-amount:amount;
  return `${amount<0n?'-':''}${absolute/1000000n}.${String(absolute%1000000n).padStart(6,'0')} USD`;
}

export async function summaryCell(row,label,expected) {
  const facts={available:true,check:'row-count',column:label==='状态'?'model-status':'other'};
  try {
    facts.row_count=await row.count();
    assert.equal(facts.row_count,1,'one exact live object row');
    facts.check='column-present';
    const headings=await row.locator('xpath=ancestor::table').locator('thead th').allTextContents();
    const index=headings.indexOf(label);
    facts.column_count=headings.length;facts.column_present=index>=0;
    assert.ok(index>=0,'required summary column: '+label);
    facts.check='cell-value';
    const actual=(await row.getByRole('cell').nth(index).textContent()).trim();
    facts.value_matches=actual===String(expected);
    if(label==='状态') {
      facts.expected_status=['启用','停用'].includes(String(expected))?String(expected):'other';
      facts.observed_status=['启用','停用'].includes(actual)?actual:'other';
    }
    assert.equal(actual,String(expected),'live summary value: '+label);
  } catch(error) {
    summaryFailures.set(error,facts);
    throw error;
  }
}

export async function managementSummaryJourney(page,{tenants,models,nav,read,scope,save,closeDetail,ready,mark}) {
  const itemList=value=>value.items??value.data;
  const exactRow=name=>page.getByRole('row').filter({has:page.getByRole('cell',{name,exact:true})});
  const kind=value=>({provider:'供应商API',self_hosted:'已有自托管',unspecified:'待确认'})[value]??'待确认';
  mark('summary-models-values');
  await scope(tenants[0]);await nav('模型托管');
  const catalog=itemList(await read(`catalog?tenant=${tenants[0]}&include_ungranted=true`));
  assert.equal(catalog.length,2,'two published model summaries');
  for(const alias of models) {
    const model=catalog.find(value=>value.alias===alias),row=exactRow(alias),sources=model.summary.sources;
    assert.equal(sources.length,1,'one actually attached controlled service');
    assert.equal(sources[0].name,'Journey controlled upstream');assert.equal(sources[0].source_kind,'provider');
    assert.equal(sources[0].tenant_input_rate,alias===models[0]?3000000:2000000);assert.equal(sources[0].tenant_output_rate,alias===models[0]?3000000:2000000);
    await summaryCell(row,'来源 / 类型',sources.map(source=>source.name+' · '+kind(source.source_kind)).join('；'));
    await summaryCell(row,'当前租户价格（USD）',sources.map(source=>source.name+'：输入 '+summaryMoney(source.tenant_input_rate)+' / 输出 '+summaryMoney(source.tenant_output_rate)+' / 百万Token').join('；'));
    await summaryCell(row,'调用延迟',model.summary.average_duration_ms==null?'未测（当前范围无已结束调用）':model.summary.average_duration_ms.toFixed(0)+' ms（当前范围已结束调用均值）');
    await summaryCell(row,'状态',model.enabled?'启用':'停用');
  }
  mark('summary-model-context-call');
  await exactRow(models[1]).getByRole('button',{name:'调用测试',exact:true}).click();
  await page.getByLabel('公共模型名',{exact:true}).waitFor({state:'visible'});
  assert.equal(await page.getByLabel('公共模型名',{exact:true}).inputValue(),models[1]);
  assert.equal(await page.getByLabel('应用API Key（不会保存）',{exact:true}).inputValue(),'');
  await nav('模型托管');
  for(const enabled of [false,true]) {
    mark('summary-model-context-toggle');
    await exactRow(models[1]).getByRole('button',{name:'启停',exact:true}).click();
    await page.getByLabel('启用',{exact:true}).setChecked(enabled);
    await save(`models/${models[1]}`);
    await summaryCell(exactRow(models[1]),'状态',enabled?'启用':'停用');
  }
  mark('summary-service-values');await nav('供应商与服务');
  const services=itemList(await read('upstreams')),routes=itemList(await read('models'));
  assert.equal(services.length,1,'one service created in the native journey');
  for(const service of services) {
    const related=routes.filter(model=>model.candidates.some(candidate=>candidate.upstream_id===service.id)),row=exactRow(service.name);
    assert.equal(related.length,2,'both actual model routes reference this service');
    await summaryCell(row,'关联模型',related.length+' 个 · '+related.map(model=>model.alias+'（候选优先级 '+model.candidates.find(candidate=>candidate.upstream_id===service.id).priority+'）').join('；'));
    await summaryCell(row,'核验事实',service.check?.applicable?({checking:'核验中',success:'核验通过',failed:'核验失败',canceled:'已取消'}[service.check.state]??'未知（'+service.check.state+'）'):'未知 / 配置已变更');
    await summaryCell(row,'核验时间 / 配置版本',(service.check?.finished_at??'未核验')+' · 当前执行版本 '+service.execution_revision+' · '+(service.check?'核验版本 '+service.check.execution_revision+(service.check.applicable?'（适用）':'（不适用）'):'无核验记录'));
    await row.getByRole('button',{name:'关联模型',exact:true}).click();
    await page.getByRole('heading',{name:'服务关联模型（优先级属于各模型候选）',exact:true}).waitFor();
    for(const model of related)assert.equal(await page.locator('#detail-content').getByRole('button',{name:'编辑 · '+model.alias,exact:true}).count(),1);
    await closeDetail();
  }
  mark('summary-tenant-values');await nav('租户与成员');
  const visible=itemList(await read('tenants'));
  assert.equal(visible.length,2,'two real tenants');
  let keyCount=0;
  for(const tenantID of tenants) {
    const tenant=visible.find(value=>value.id===tenantID),row=exactRow(tenant.name);
    const members=itemList(await read(`tenants/${tenantID}/members`)),keys=itemList(await read(`tenants/${tenantID}/keys`)),limits=await read(`tenants/${tenantID}/limits`),grants=itemList(await read(`catalog?tenant=${tenantID}`));
    const owners=members.filter(member=>member.active&&member.role==='administrator').map(member=>member.account).sort();
    assert.deepEqual(tenant.summary.owners,owners,'owner is a real membership, not the caller role');
    assert.deepEqual(tenant.summary.usage,limits.usage);assert.deepEqual(tenant.summary.limits,limits.limits);
    await summaryCell(row,'负责人',owners.join('、')||'尚无有效租户管理员');
    await summaryCell(row,'成员数',members.filter(member=>member.active).length);
    await summaryCell(row,'授权模型',grants.map(model=>model.alias).join('、')||'无授权');
    await summaryCell(row,'应用Key数（含撤销历史）',keys.length);keyCount+=keys.length;
    await summaryCell(row,'预算 / 已用',`${limits.period} · 已确认 ${summaryMoney(limits.usage.fee_micro_usd)} / 预留 ${summaryMoney(limits.usage.reserved_micro_usd)} / 上限 ${limits.limits.budget_micro_usd==null?'无限':summaryMoney(limits.limits.budget_micro_usd)} · ${limits.limits.hard?'硬预算':'软预算'} · 待确认 ${limits.usage.pending} · 平台上界 ${limits.platform_limits?.budget_micro_usd==null?'未设置':summaryMoney(limits.platform_limits.budget_micro_usd)}`);
  }
  assert.ok(keyCount>=3,'multiple real application Keys');
  mark('summary-tenant-context');
  const second=visible.find(tenant=>tenant.id===tenants[1]);
  await exactRow(second.name).getByRole('button',{name:'应用Key',exact:true}).click();
  await page.locator('#page-title').getByText('API Keys',{exact:true}).waitFor();
  await ready();
  assert.equal(await page.getByLabel('当前租户').inputValue(),tenants[1]);
  await scope(tenants[0]);await nav('工作台');
  const scopedModels=itemList(await read(`catalog?tenant=${tenants[0]}`));
  for(const model of scopedModels)await summaryCell(exactRow(model.alias),'状态',model.enabled?'启用':'停用');
  await page.getByRole('button',{name:'管理模型',exact:true}).click();
  await page.locator('#page-title').getByText('模型托管',{exact:true}).waitFor();
  await ready();
  assert.equal(await page.getByLabel('当前租户').inputValue(),tenants[0]);
  await nav('工作台');await page.getByRole('button',{name:'查看预算',exact:true}).click();
  await page.locator('#page-title').getByText('配额与预算',{exact:true}).waitFor();
  await ready();
  assert.equal(await page.getByLabel('当前租户').inputValue(),tenants[0]);
  return {models:catalog.length,tenants:visible.length,services:services.length,keys:keyCount,api_values_match:true,context_actions:true};
}
