import {createLimitSummary} from './presentation.mjs';
import {createLifetime} from '../../shared/lifetime.mjs';

export function createLimits({request,scope,ui}) {
  const lifetime=createLifetime(request),{api,collection}=lifetime;
  const {node,button,edit,mainAction,detail}=ui;
  const limitSummary=createLimitSummary(ui);
  const tenantPath=resource=>`tenants/${encodeURIComponent(scope.tenant)}/${resource}`;
function limitsEditor(path,current) { edit('调整独立限额',[...['calls','rpm','tpm','concurrency','budget_micro_usd'].map((name,index)=>({name,label:['本周期次数','每分钟请求 RPM','每分钟收费Token TPM','本地并发','费用预算（微美元）'][index],type:'number',min:0,value:current[name]??'',help:'留空无限；0为禁止；调整不会清零已用量。'})),{name:'hard',label:'硬费用预算（需要供应商可信收费Token上界）',type:'checkbox',value:current.hard}],values=>api(path,'PUT',{...Object.fromEntries(['calls','rpm','tpm','concurrency','budget_micro_usd'].map(name=>[name,values[name]===''?null:Number(values[name])])),hard:values.hard,revision:current.revision})); }
  async function render() {
    lifetime.activate();
    const section=node('section');
    if(!scope.tenant){section.append(node('div','请先选择一个租户；跨租户全局数据仅平台管理员可读取。','empty'));return lifetime.mount(section);}
if(scope.role!=='developer'){const tenantLimits=await api(tenantPath('limits'));section.append(node('h2','租户共享资源'),limitSummary(tenantLimits));if(scope.role==='administrator')mainAction('调整租户限额',()=>limitsEditor(tenantPath('limits'),tenantLimits.limits));}else section.append(node('p','仅显示分配应用的Key限额与使用量；租户账单和整租户用量不可见。'));const keys=await collection(tenantPath('keys'));for(const key of keys){const value=await api(tenantPath('keys')+'/'+key.id+'/limits'),card=node('div','','card');card.append(node('h3',key.name+' · 独立Key限额'),limitSummary(value));if(scope.role==='administrator')card.append(button('调整Key限额',()=>limitsEditor(tenantPath('keys')+'/'+key.id+'/limits',value.limits)));section.append(card);}
    return lifetime.mount(section);
  }
  return {render,dispose:()=>lifetime.dispose(),showResources(title,report){detail(title,limitSummary(report));}};
}
