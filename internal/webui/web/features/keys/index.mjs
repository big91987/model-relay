import {createLifetime} from '../../shared/lifetime.mjs';
import {list} from '../../shared/data.mjs';
import {money} from '../../shared/catalog.mjs';
export function createKeys({request,scope,ui,showRecords,showResources}) {
  const lifetime=createLifetime(request),{api,collection}=lifetime;
  const {node,table,column,actions,detail,edit,mainAction,notice}=ui;
  const scoped=path=>path+(path.includes('?')?'&':'?')+(scope.tenant?'tenant='+encodeURIComponent(scope.tenant):'');
  const tenantPath=resource=>`tenants/${encodeURIComponent(scope.tenant)}/${resource}`;

function keyEditor(key) {
  Promise.all([collection(scoped('catalog')),scope.role==='administrator'?collection(tenantPath('members')):Promise.resolve([])]).then(([models,members])=>edit(key?'编辑应用Key':'创建应用Key',[
    {name:'name',label:'名称',value:key?.name,required:true},{name:'application',label:'应用',value:key?.application,required:true},
    {name:'models',label:'模型授权（可多选）',options:list(models).map(model=>({value:model.alias,label:model.alias})),multiple:true,value:key?.models??[]},
    {name:'diagnostics',label:'诊断开发者（可多选）',options:list(members).filter(member=>member.role==='developer'&&member.active).map(member=>({value:member.id,label:member.account})),multiple:true,value:key?.diagnostic_members??[]},
    {name:'expires',label:'UTC到期时间（留空为无限）',value:key?.expires_at??'',help:'ISO 8601，例如2027-01-01T00:00:00Z'},
    {name:'enabled',label:'启用',type:'checkbox',value:key?.enabled??true}
  ],(values,operation)=>api(tenantPath('keys')+(key?'/'+key.id:''),key?'PATCH':'POST',{name:values.name,application:values.application,models:values.models,diagnostic_members:values.diagnostics,expires_at:values.expires||null,enabled:values.enabled,expected_revision:key?.revision??0},operation))).catch(error=>{if(error.name!=='AbortError')notice('读取编辑依赖失败：'+error.message);});
}
  async function render() {
    lifetime.activate();
    const section=node('section');
    if(!scope.tenant){section.append(node('div','请先选择一个租户；跨租户全局数据仅平台管理员可读取。','empty'));return lifetime.mount(section);}
const values=await collection(tenantPath('keys'));await Promise.all(values.map(async key=>{key.resources=await api(tenantPath('keys')+'/'+key.id+'/limits');}));section.append(table([column('名称','name'),column('应用','application'),column('安全标识','hint'),column('模型授权','models',value=>Array.isArray(value)?value.join('、')||'无授权':'未知'),{label:'状态',render:keyStatus},column('期限','expires_at',value=>value??'无限'),{label:'费用已用 / 预留 / 上限',render:row=>`${money(row.resources.usage.fee_micro_usd)} / ${money(row.resources.usage.reserved_micro_usd)} / ${row.resources.limits.budget_micro_usd==null?'无限':money(row.resources.limits.budget_micro_usd)}`},{label:'操作',render:row=>actions([['详情',()=>detail('Key元数据（无秘密）',row)],['额度 / 用量',async()=>showResources('应用独立资源与使用量',await api(tenantPath('keys')+'/'+row.id+'/limits'))],['历史',()=>{showRecords({key:row.id});}],...(scope.role==='administrator'?[['编辑',()=>keyEditor(row)],['轮换',()=>edit('轮换（共享额度与历史）',[{name:'transition_hours',label:'旧代际过渡',options:[{value:'0',label:'立即失效'},{value:'1',label:'1小时'},{value:'24',label:'24小时'}],value:'0'}],(values,operation)=>api(tenantPath('keys')+'/'+row.id+'/rotations','POST',{transition_hours:Number(values.transition_hours),expected_revision:row.revision},operation))],['结束过渡',()=>edit('提前撤销旧代际',[],(values,operation)=>api(tenantPath('keys')+'/'+row.id+'/retire-previous','POST',{expected_revision:row.revision},operation))],['撤销',()=>edit('永久撤销应用Key',[],(values,operation)=>api(tenantPath('keys')+'/'+row.id+'/revoke','POST',{expected_revision:row.revision},operation))]]:[])])}],values));if(scope.role==='administrator')mainAction('创建Key',()=>keyEditor());
    return lifetime.mount(section);
  }
  return {render,dispose:()=>lifetime.dispose()};
}
export function keyStatus(key, now=Date.now()) {
  if(key.revoked===true)return '已撤销';
  if(key.enabled===false)return '已停用';
  if(key.enabled!==true||key.revoked!==false)return '未知';
  if(key.expires_at!=null){const expiry=Date.parse(key.expires_at);if(!Number.isFinite(expiry))return '未知';if(expiry<=now)return '已过期';}
  if(!Array.isArray(key.generations))return '未知';
  const grace=key.generations.some(generation=>!generation.retired&&generation.expires!=null&&Date.parse(generation.expires)>now);
  return grace?'过渡中':'已启用';
}
