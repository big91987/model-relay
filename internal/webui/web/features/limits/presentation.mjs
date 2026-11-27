import {money} from '../../shared/catalog.mjs';

export function createLimitSummary({node,table,column,factSummary}) {
  return function limitSummary(report) {
  if(!report?.limits||!report?.usage)return factSummary(report);
  const content=node('div'),limits=report.limits,usage=report.usage;
  const rows=[['本周期次数','calls',usage.calls],['每分钟请求 RPM','rpm',usage.rpm],['每分钟收费Token TPM（含预留）','tpm',usage.tpm],['本地活跃并发','concurrency',usage.concurrency],['费用预算（USD）','budget_micro_usd',BigInt(usage.fee_micro_usd)+BigInt(usage.reserved_micro_usd)]].map(([name,field,used])=>{const monetary=field==='budget_micro_usd',limit=limits[field],remaining=limit==null?null:BigInt(limit)-BigInt(used);return {name,limit:limit==null?'无限':monetary?money(limit):limit,used:monetary?money(used):used,remaining:remaining==null?'无限':monetary?money(remaining):String(remaining),platform:report.platform_limits?.[field]==null?'未设置':monetary?money(report.platform_limits[field]):report.platform_limits[field]};});
  content.append(node('p',report.period+' UTC · '+(limits.hard?'硬费用准入（可信上界必需）':'软预算（超额告警，仍允许调用）')),table([column('独立维度','name'),column('有效上限','limit'),column('已用 / 含预留','used'),column('剩余','remaining'),column('平台上界','platform')],rows),node('p','已确认费用 '+money(usage.fee_micro_usd)+' · 未决费用预留 '+money(usage.reserved_micro_usd)+' · '+usage.pending+' 个待确认尝试；未知金额不是零。'));
  if(report.soft_budget_exceeded)content.append(node('p','软预算已超额，请核对费用或调整预算。','error'));return content;
};
}
