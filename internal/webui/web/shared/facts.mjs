import {enumLabel} from './catalog.mjs';

export function createFactSummary({node}) {
  function technicalValue(value) {
    if(value==='')return node('span','—');
    const element=node('span',value==null?'未知':String(value),'technical-value');
    element.tabIndex=0;element.setAttribute('role','region');element.setAttribute('aria-label',element.textContent);element.setAttribute('aria-description','完整值，可横向滚动查看');return element;
  }

  function factSummary(value,fieldName='',record={}) {
    if(value==null)return node('span','未知');
    if(typeof value!=='object'){const text=typeof value==='boolean'?(value?'是':'否'):enumLabel(fieldName,value,record);return ['id','url','alias','model','remote_model','tenant_id','key_id','hint','account','error_code','price_id','request_id'].includes(fieldName)?technicalValue(text):node('span',String(text));}
    if(Array.isArray(value)){const list=node('ol');for(const item of value){const entry=node('li');entry.append(factSummary(item));list.append(entry);}if(!value.length)list.append(node('li','暂无记录'));return list;}
    const labels={id:'安全标识',name:'名称',application:'应用',alias:'公共模型',url:'基础地址',enabled:'启用',state:'状态',revision:'版本',expires_at:'期限',created_at:'创建时间',tenant_id:'租户标识',key_id:'应用Key标识',model:'模型',prices:'价格版本',bounds:'上界依据',candidates:'路由候选',ordinal:'尝试序号',error_code:'原因代码',input_tokens:'输入Token',output_tokens:'输出Token',prompt_tokens:'原始上游输入Token',completion_tokens:'原始上游输出Token',total_tokens:'原始上游总Token',confirmed_input_tokens:'已确认计费输入Token',confirmed_output_tokens:'已确认计费输出Token',cost:'供应商成本（微美元）',fee:'租户费用（微美元）',charge_entries:'追加计费明细',price_snapshots:'受理价格快照',historical_price_evidence:'历史依据',request_attempts:'调用尝试',attempts:'调用尝试',cost_input_rate:'供应商输入成本（微美元 / 百万Token）',cost_output_rate:'供应商输出成本（微美元 / 百万Token）',tenant_input_rate:'租户输入价格（微美元 / 百万Token）',tenant_output_rate:'租户输出价格（微美元 / 百万Token）',pending:'待确认尝试',period:'UTC周期',reason:'原因',evidence:'依据引用'};
    Object.assign(labels,{routing_strategy:'路由策略',routing_decisions:'当时选择依据',latency_observations:'本次延迟观测',strategy:'路由策略',metric_name:'测量口径',sample_state:'观测状态',sample_count:'近期成功样本数',ewma_ms:'EWMA毫秒',measured_at:'实际测量时间',age_ms:'选择时样本年龄ms',ttl_seconds:'测量有效期秒',selection_reason:'选择原因',sample_acceptance:'采纳状态',latency_ms:'本次起始延迟ms',evaluated_at:'评估时间（不续期）',generation:'配置代次',exploration_sequence:'探索选择序号',exploration_bound:'稳定候选探索上界（选择次数）',algorithm:'算法版本',source_kind:'来源类型',role:'角色',hint:'安全标识',check:'当前配置核验',execution_revision:'执行配置版本',applicable:'适用于当前配置',started_at:'开始时间',finished_at:'结束时间',duration_ms:'耗时ms',upstream_status:'上游HTTP状态',remote_model:'远端模型名',upstream_id:'服务ID',priority:'候选优先级',weight:'候选权重',protocol:'协议',stream:'SSE流式',retired:'已退役'});
    const facts=node('dl');for(const [name,item] of Object.entries(value)){const field=node('dd');field.append(factSummary(item,name,value));facts.append(node('dt',labels[name]??name),field);}return facts;
  }

  const column = (label,field,format) => ({label,render:row=>format?format(row[field]):['url','alias','id','hint','account'].includes(field)?technicalValue(row[field]):enumLabel(field,row[field],row)});
  return {factSummary,technicalValue,column};
}
