import {createLifetime} from '../../shared/lifetime.mjs';

import {money} from '../../shared/catalog.mjs';
export function createBilling({request,scope,ui,download}) {
  const lifetime=createLifetime(request),{api,collection}=lifetime;
  const {node,button,table,column,actions,detail,edit,closeDetail}=ui;
  const scoped=path=>path+(path.includes('?')?'&':'?')+(scope.tenant?'tenant='+encodeURIComponent(scope.tenant):'');
  const tenantPath=resource=>`tenants/${encodeURIComponent(scope.tenant)}/${resource}`;
async function priceEditor(alias) {
  const models=await collection('models'),model=models.find(value=>value.alias===alias);if(!model)throw new Error('模型已不存在');
  const fields=model.candidates.length===1?[{name:'upstream',label:'服务ID',type:'number',min:1,required:true,value:model.candidates[0].upstream_id},...['cost_input_rate','cost_output_rate'].map((name,index)=>({name,label:['供应商输入成本','供应商输出成本'][index]+'（微美元 / 百万Token）',type:'number',min:0,required:true}))]:[{name:'costs',label:'各候选独立成本（每行 服务ID,输入费率,输出费率）',type:'textarea',required:true,value:model.candidates.map(value=>`${value.upstream_id},0,0`).join('\n'),help:'必须覆盖全部候选；不会套用其他候选的供应商成本。'}];
  fields.push(...['tenant_input_rate','tenant_output_rate'].map((name,index)=>({name,label:['租户输入价格','租户输出价格'][index]+'（微美元 / 百万Token）',type:'number',min:0,required:true})),{name:'effective',label:'UTC生效时间（空为当前时间）'});
  edit('发布测试价格 · '+alias,fields,(values,operation)=>{const costs=model.candidates.length===1?[[values.upstream,values.cost_input_rate,values.cost_output_rate]]:values.costs.split('\n').filter(Boolean).map(line=>line.split(','));const prices=costs.map(parts=>({upstream_id:Number(parts[0]),cost_input_rate:Number(parts[1]),cost_output_rate:Number(parts[2]),tenant_input_rate:Number(values.tenant_input_rate),tenant_output_rate:Number(values.tenant_output_rate)}));return api('models/'+encodeURIComponent(alias)+'/prices','POST',{effective_at:values.effective,prices},operation);});
}
  const exportData=path=>download(scoped(path));
  async function render() {
    lifetime.activate();
    const section=node('section');
const models=await collection(scoped('catalog'));section.append(node('h2','模型价格（USD测试价）'),table([column('模型','alias'),column('租户输入 USD/百万Token','tenant_input_rate',money),column('租户输出 USD/百万Token','tenant_output_rate',money),{label:'操作',render:row=>scope.platform?actions([['版本',async()=>detail('历史价格版本',await api('models/'+encodeURIComponent(row.alias)+'/prices'))],['发布新价',()=>priceEditor(row.alias)]]):'价格随授权目录展示'}],models));if(scope.tenant){const bills=await collection(tenantPath('bills'));section.append(node('h2','租户周期账单 · 外部支付未接通'),table([column('周期','period'),column('原 / 已确认费用','fee_micro_usd',money),column('成本','cost_micro_usd',money),column('待确认','pending'),{label:'操作',render:row=>actions(row.bill_id?[['详情 / 更正',async()=>{const value=await api('bills/'+row.bill_id),controls=[button('导出周期完整账单',()=>exportData('bills/'+row.bill_id+'/export'))];if(scope.platform)controls.push(button('追加更正',()=>{closeDetail();edit('追加账单更正',[{name:'amount_micro_usd',label:'有符号更正金额（微美元）',type:'number',required:true},{name:'adjustment_period',label:'开放调整周期（UTC月）',value:new Date().toISOString().slice(0,7)},{name:'reason',label:'原因',required:true}],(values,operation)=>api('bills/'+row.bill_id+'/corrections','POST',{...values,amount_micro_usd:Number(values.amount_micro_usd)},operation));}));detail('已结账，原金额不可覆写（调整仅计入各调整周期一次）',value,controls);}]]:scope.platform?[['结账',()=>edit('结束且无待确认才可结账',[],(values,operation)=>api(tenantPath('bills'),'POST',{period:row.period},operation))]]:[]) }],bills));}
    return lifetime.mount(section);
  }
  return {render,dispose:()=>lifetime.dispose()};
}
