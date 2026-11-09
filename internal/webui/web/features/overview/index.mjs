import {createLifetime} from '../../shared/lifetime.mjs';
import {list} from '../../shared/data.mjs';
import {money,modelSources,modelLatency} from '../../shared/catalog.mjs';
export function createOverview({request,scope,ui,navigate,openPlayground}) {
  const lifetime=createLifetime(request),{api,collection}=lifetime;
  const {node,button,table,column,mainAction}=ui;
  const scoped=path=>path+(path.includes('?')?'&':'?')+(scope.tenant?'tenant='+encodeURIComponent(scope.tenant):'');
  async function render() {
    lifetime.activate();
    const section=node('section');
const report=await api(scoped('workspace/overview')),cards=node('div','','cards');
      for(const [label,value] of [['今日调用',report.requests_today],['今日成功率',report.success_rate_percent==null?'暂无样本':report.success_rate_percent.toFixed(2)+'%'],['本月已确认费用',money(report.confirmed_fee_micro_usd)],['已结束调用平均延迟',report.average_duration_ms==null?'暂无样本':report.average_duration_ms.toFixed(0)+' ms']]){const card=node('div','','card');card.append(node('p',label),node('div',String(value),'metric'));cards.append(card);}
      const charts=node('div','','overview-charts'),trend=node('div','','card'),distribution=node('div','','card');trend.append(node('h2','近7日调用趋势'),node('p','当前人员与租户范围 · UTC日期 · 缺少最终用量不视为零费用'));
      const points=list(report.trend),maximum=Math.max(1,...points.map(point=>Number(point.requests))),bars=node('div','','chart-bars');
      for(const point of points){const row=node('div','','bar'),progress=node('progress');progress.max=maximum;progress.value=Number(point.requests);progress.setAttribute('aria-label',point.day+'请求 '+point.requests+'，成功 '+point.successes);row.append(node('span',point.day),progress,node('span',point.requests+' 次'));bars.append(row);}trend.append(points.length?bars:node('div','当前范围近7日暂无调用','empty'));
      distribution.append(node('h2','租户费用分布'),node('p',report.period+' · 已确认租户费，不包含未知费用'),table([column('租户','name'),column('已确认','fee_micro_usd',money),column('待确认尝试','pending')],list(report.tenant_distribution)));charts.append(trend,distribution);
      const risk=node('div','','card');risk.append(node('h2','待确认与保守预留'),node('p',report.pending_attempts+' 个待确认尝试 · '+report.legacy_pending_requests+' 个历史未定价请求 · 费用预留 '+money(report.reserved_micro_usd)+'（不代表历史未知费用为零）'),button('查看请求与核对',()=>navigate('records')),button('查看用量分析',()=>navigate('usage')));
      const models=await collection(scoped('catalog')),status=node('div','','card');status.append(node('h2','当前范围模型状态'),table([column('公共模型','alias'),column('状态','enabled',value=>value?'启用':'停用'),{label:'来源 / 类型',render:modelSources},{label:'调用延迟',render:modelLatency},{label:'操作',render:row=>button('调用测试',()=>openPlayground(row.alias))}],models),button('管理模型',()=>navigate('models')),button('查看预算',()=>navigate('limits')));section.append(cards,charts,risk,status);mainAction('发起调用',()=>navigate('playground'));
    return lifetime.mount(section);
  }
  return {render,dispose:()=>lifetime.dispose()};
}
