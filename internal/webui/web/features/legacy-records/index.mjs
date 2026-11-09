import {friendly,known} from '../../shared/legacy-presentation.mjs';
export function createLegacyRecords({document,element,api:request,current:scopeCurrent,assertCurrent:assertScope,stale,guarded,form,load,message}) {
let nextCursor=null,findSequence=0;
const current=token=>scopeCurrent(token)&&(token.find===undefined||token.find===findSequence);
const assertCurrent=token=>{assertScope(token);if(!current(token))throw stale();};
const api=async(path,method,body,token,...rest)=>{const value=await request(path,method,body,token,...rest);assertCurrent(token);return value;};
const decisionReasons = {selected: '选中', disabled_model: '模型停用', disabled_upstream: '上游停用', disabled_candidate: '候选停用或移出', credential_missing: '凭据缺失', cooling: '暂时冷却', already_attempted: '本请求已尝试', lower_priority: '优先级较低', not_selected_same_priority: '同级轮转未选中'};
const states = {started: '网关已受理，尚未落盘终态', success: '成功', rejected: '网关拒绝', upstream_rejected: '供应商拒绝', upstream_failed: '上游失败', timeout: '超时', stream_interrupted: '流不完整', canceled: '取消', process_interrupted: '服务重启中断'};
function recordValues(value) { return [value.id + '\n' + value.time, known(value.alias) + ' → ' + known(value.model) + '\n密钥标识：' + (value.key_id || '未识别应用密钥'), (states[value.state] || '未知') + ' (' + value.state + ')\n' + (value.error_code ? friendly(value.error_code) + ' (' + value.error_code + ')' : '') + '\n上游状态 ' + known(value.upstream_status), known(value.duration_ms), [value.prompt_tokens, value.completion_tokens, value.total_tokens].map(known).join(' / ')]; }
function recordRow(value) { const row = document.createElement('tr'); for (const text of recordValues(value)) { const cell = document.createElement('td'); cell.textContent = text; row.append(cell); } return row; }
async function findRecord(token) {
  const find=++findSequence;
  token = {...token, find,valid:()=>find===findSequence};
  const id = element('find-request-id').value.trim(); if (!id) { message('请输入请求 ID；未取得网关 ID 时无法定位记录。', true); return; }
  element('record-detail').textContent = ''; element('record-status').textContent = '正在定位同一请求…';
  try {
    const value = await api('requests/' + encodeURIComponent(id), 'GET', undefined, token);
    let evidence = '\n历史记录：没有逐尝试证据，不反推实际上游。';
    if (value.evidence_version === 2) {
      evidence = '\n尝试证据：' + (value.attempts_complete ? '终态已记录' : '尚不完整，登记不证明供应商已执行') + '\n响应提交：' + known(value.committed_at) + '\n配置适用版本：' + known(value.config_revision);
      evidence += '\n路由策略与选择依据（未知不为零）：' + JSON.stringify(value.routing_decisions || []) + '\n本次观测：' + JSON.stringify(value.latency_observations || []);
      evidence += '\n' + value.attempts.map(attempt => '尝试 ' + attempt.ordinal + ' · ' + attempt.upstream_name + ' → ' + attempt.remote_model + ' · ' + (states[attempt.state] || attempt.state) + ' · ' + known(attempt.started_at) + ' · ' + known(attempt.duration_ms) + 'ms · ' + (attempt.error_code ? friendly(attempt.error_code) : '无错误') + ' · 用量 ' + [attempt.prompt_tokens, attempt.completion_tokens, attempt.total_tokens].map(known).join(' / ')).join('\n');
      evidence += '\n选择说明（不计作生成尝试）：\n' + value.selection_decisions.map(decision => '第 ' + decision.selection_round + ' 轮 · 上游 ' + decision.upstream_id + ' · ' + (decisionReasons[decision.reason] || decision.reason) + (decision.cooldown_until ? ' · 冷却至 ' + decision.cooldown_until : '')).join('\n');
      evidence += '\n用量合计：\n' + Object.entries(value.usage_summary).map(([field, part]) => field + '：完整合计 ' + known(part.total) + '；已知部分 ' + part.known_sum + '；未知尝试 ' + part.unknown_attempts + '；' + part.completeness).join('\n');
    }
    element('record-detail').textContent = recordValues(value).join('\n') + evidence; element('record-status').textContent = '已定位同一请求 ' + value.id; element('record-detail').focus();
  } catch (error) {
    if (error.stale || !current(token)) return;
    element('record-status').textContent = error.code === 'not_found' ? '暂未找到该 ID 记录：可能尚未落盘或未能记录。可手动再次定位；不代表未计费。' : error.message;
  }
}
async function render(token,pageCursor='') {

    const value = await api('requests' + (nextCursor ? '?cursor=' + encodeURIComponent(nextCursor) : ''), 'GET', undefined, token);
    element('record-list').replaceChildren(...value.data.map(recordRow)); nextCursor = value.next_cursor; element('more-records').hidden = !nextCursor;
    element('record-status').textContent = value.data.length ? '已读取真实记录；可用请求 ID 精确定位，不受分页影响。' : '尚无调用记录，请先发起调用；不会显示演示记录。';
  
}
element('refresh-records').onclick = () => { nextCursor = null; return guarded(element('refresh-records'), token => load('records', token)); };
element('more-records').onclick = () => guarded(element('more-records'), token => load('records', token));
form('find-record-form', findRecord);
return {render,find:findRecord,dispose(){nextCursor=null;findSequence++;},reset(){nextCursor=null;element('find-request-id').value='';element('record-detail').textContent='';element('record-list').replaceChildren();}};
}
