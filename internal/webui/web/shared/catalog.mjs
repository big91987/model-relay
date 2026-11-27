import {list} from './data.mjs';
export const money = value => { if(value==null)return '未知';const amount=BigInt(value),negative=amount<0n,absolute=negative?-amount:amount;return `${negative?'-':''}${absolute/1000000n}.${String(absolute%1000000n).padStart(6,'0')} USD`; };
export function enumLabel(field,value,record={}) {
  if(['routing_strategy','strategy','metric_name','sample_state','sample_acceptance','selection_reason'].includes(field)){const labels={weighted:'权重（默认）',low_latency:'低延迟优先',response_body_first_byte_ms:'响应体首字节（非首token）',first_output_delta_ms:'首个有效输出增量（非首token）',unknown:'无本进程样本 / 未知',expired:'测量已过期（完成不续期）',insufficient:'有效样本不足3次',ready:'近期有效样本',accepted:'有效成功观测',not_success:'非成功，不采样',no_output_delta:'无有效输出增量',configuration_changed:'配置已变化，不采纳',out_of_order:'旧序列 / 旧测量，不采纳',invalidated_by_failure:'故障前在途观测已失效，不采纳',clock_invalid:'时钟异常，不采纳',explore:'有界探索（正常请求）',lowest_latency:'近期最低延迟',weighted_fallback:'未知回退权重',not_selected:'未选中'};return labels[value]??value;}
  const roles={administrator:'租户管理员',developer:'开发者',finance:'财务观察者'};
  const sources={provider:'供应商API',self_hosted:'已有自托管',unspecified:'待确认'};
  const calls={started:'处理中',dispatching:'正在调用上游',success:'成功',rejected:'已拒绝',upstream_failed:'上游失败',upstream_rejected:'上游拒绝',stream_interrupted:'流式中断',timeout:'超时',canceled:'已取消',process_interrupted:'进程中断'};
  const checks={checking:'核验中',success:'核验通过',failed:'核验失败',canceled:'已取消',unknown:'未知'};
  const invitations={pending:'待接受',accepted:'已接受',revoked:'已撤销',expired:'已过期'};
  const labels=field==='role'?roles:field==='source_kind'?sources:field==='state'?(Object.hasOwn(record,'execution_revision')?checks:Object.hasOwn(record,'account')?invitations:calls):null;
  return labels?(Object.hasOwn(labels,value)?labels[value]:value==null?'未知':'未知（'+String(value)+'）'):value;
}
export const sourceKind=value=>enumLabel('source_kind',value);
export function modelSources(model) {return list(model.summary?.sources).map(source=>source.name+' · '+sourceKind(source.source_kind)).join('；')||'无候选服务';}
export function modelPrices(model) {return list(model.summary?.sources).map(source=>source.name+'：输入 '+money(source.tenant_input_rate)+' / 输出 '+money(source.tenant_output_rate)+' / 百万Token').join('；')||'未定价';}
export function modelLatency(model) {const value=model.summary?.average_duration_ms;return value==null?'未测（当前范围无已结束调用）':value.toFixed(0)+' ms（当前范围已结束调用均值）';}
