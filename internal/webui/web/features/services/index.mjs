import {createLifetime} from '../../shared/lifetime.mjs';

import {enumLabel,sourceKind} from '../../shared/catalog.mjs';
export function createServices({request,scope,ui,openModel}) {
  const lifetime=createLifetime(request),{api,collection}=lifetime;
  const {node,button,table,column,actions,detail,edit,mainAction,notice,closeDetail}=ui;
function serviceEditor(service) { edit(service?'编辑服务':'接入供应商 / 已有自托管服务',[{name:'name',label:'服务名称',required:true,value:service?.name},{name:'source_kind',label:'来源类型',options:[{value:'provider',label:'供应商API'},{value:'self_hosted',label:'已有自托管推理端点'},{value:'unspecified',label:'历史来源待确认'}],value:service?.source_kind??'provider'},{name:'url',label:'兼容API基础地址',required:true,value:service?.url,help:'例如https://provider.example/v1；私网必须由运营者显式允许。'},{name:'token',label:'上游凭据（编辑留空保留同地址原值）',type:'password',sensitive:true},{name:'enabled',label:'启用',type:'checkbox',value:service?.enabled??true}],values=>api('upstreams'+(service?'/'+service.id:''),service?'PATCH':'POST',{...values,...(service?{expected_revision:service.revision}:{})})); }
  async function render() {
    lifetime.activate();
    const section=node('section');
if(!scope.platform) section.append(node('div','仅平台管理员可管理供应商端点。','empty')); else {const [values,models]=await Promise.all([collection('upstreams'),collection('models')]);for(const service of values)service.models=models.filter(model=>model.candidates.some(candidate=>candidate.upstream_id===service.id));section.append(table([column('服务','name'),column('来源','source_kind',sourceKind),column('基础地址','url'),column('状态','enabled',value=>value?'启用':'停用'),{label:'关联模型',render:row=>row.models.length+' 个 · '+row.models.map(model=>model.alias+'（候选优先级 '+model.candidates.find(candidate=>candidate.upstream_id===row.id).priority+'）').join('；')},{label:'核验事实',render:row=>row.check?.applicable?enumLabel('state',row.check.state,row.check):'未知 / 配置已变更'},{label:'核验时间 / 配置版本',render:row=>(row.check?.finished_at??'未核验')+' · 当前执行版本 '+row.execution_revision+' · '+(row.check?'核验版本 '+row.check.execution_revision+(row.check.applicable?'（适用）':'（不适用）'):'无核验记录')},{label:'操作',render:row=>actions([['编辑',()=>serviceEditor(row)],['核验',async()=>{try{const value=await api('upstreams/'+row.id+'/check','POST',{expected_execution_revision:row.execution_revision});detail('当前配置核验',value);}catch(error){if(error.name!=='AbortError')notice(error.message);} }],['详情',()=>detail('服务详情',row)],['关联模型',()=>detail('服务关联模型（优先级属于各模型候选）',row.models,row.models.map(model=>button('编辑 · '+model.alias,()=>{closeDetail();return openModel(model);})))]] )}],values));mainAction('接入服务',()=>serviceEditor());}
    return lifetime.mount(section);
  }
  return {render,dispose:()=>lifetime.dispose()};
}
