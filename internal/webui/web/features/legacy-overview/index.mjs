import {errors,known} from '../../shared/legacy-presentation.mjs';
export function createLegacyOverview({element,api,owner,navigate,isOverviewActive}) {
let versionRequest=null,overviewEvidence=null;
function clearVersion() {
  const request = versionRequest; versionRequest = null;
  if (request) { clearTimeout(request.timer); request.controller.abort(); }
  element('runtime-version').textContent = '';
}
async function loadVersion() {
  if (element('console').hidden) return;
  const request = {controller: new AbortController(), navigation: owner().navigation}; versionRequest = request;
  const isCurrent = () => versionRequest === request && isOverviewActive(request.navigation) && !element('console').hidden;
  element('runtime-version').textContent = '加载中';
  request.timer = setTimeout(() => { if (isCurrent()) { element('runtime-version').textContent = '版本未知'; versionRequest = null; } request.controller.abort(); }, 5000);
  try {
    const response = await fetch('/healthz', {method: 'GET', cache: 'no-store', signal: request.controller.signal});
    if (!response.ok) throw new Error('health unavailable');
    const value = await response.json();
    const version = value && !Array.isArray(value) && typeof value.version === 'string' && value.version.trim() ? value.version : '版本未知';
    if (isCurrent()) element('runtime-version').textContent = version;
  } catch { if (isCurrent()) element('runtime-version').textContent = '版本未知'; }
  finally { clearTimeout(request.timer); if (versionRequest === request) versionRequest = null; }
}
function renderSetup(completed) {
  const steps = [...element('setup-steps').children, ...element('completed-steps').children];
  const order = ['step-upstream', 'step-models', 'step-keys', 'step-test', 'overview-record'];
  for (let index = 0; index < order.length; index++) {
    const row = steps.find(row => row.querySelector('#' + order[index]));
    if (!row) continue;
    row.value = index + 1; row.classList.toggle('completed', !!completed[index]);
    element(completed[index] ? 'completed-steps' : 'setup-steps').append(row);
  }
  element('completed-setup').hidden = !completed.some(Boolean);
  element('setup-complete').hidden = !completed.every(Boolean);
}
async function render(token,pageCursor='') {

    overviewEvidence = null; renderSetup([false, false, false, false, false]);
    element('overview-evidence').textContent = '';
    element('overview-record').disabled = true;
    element('summary').textContent = '加载中';
    for (const id of ['overview-upstream', 'overview-models', 'overview-keys']) element(id).textContent = '未知';
    for (const id of ['step-upstream', 'step-models', 'step-keys', 'step-test']) element(id).textContent = '状态待读取';
    const value = await api('overview', 'GET', undefined, token);
    const upstream = value.upstream;
    element('overview-upstream').textContent = !upstream.configured ? '未配置' : upstream.total + ' 个 / 启用 ' + upstream.enabled_count;
    element('overview-models').textContent = value.models.total + ' 个 / 启用 ' + value.models.enabled;
    element('overview-keys').textContent = value.keys.total + ' 个 / 启用 ' + value.keys.enabled;
    overviewEvidence = value.degraded ? null : value.call_evidence;
    element('overview-record').disabled = !overviewEvidence;
    element('summary').textContent = value.degraded ? errors.storage_unavailable : !upstream.configured ? '待配置：从接入上游开始' : overviewEvidence ? '当前配置曾成功调用 · 模型 ' + overviewEvidence.alias + '；请求身份、时间及适用范围可进入详情查看。' : '配置已保存；尚无当前配置的生成成功记录。';
    element('overview-evidence').textContent = overviewEvidence ? '请求 ' + overviewEvidence.request_id + '\n调用时间 ' + known(overviewEvidence.time) + '\n上游配置版本 ' + known(overviewEvidence.upstream_revision) + '\n' + overviewEvidence.scope : '尚无当前配置的成功证据。';
    element('step-upstream').textContent = upstream.configured && upstream.enabled ? '已保存并启用（尚非健康证明）' : '下一步：配置或启用';
    element('step-models').textContent = value.models.enabled ? '已有启用模型，可继续' : '下一步：创建或启用模型';
    element('step-keys').textContent = value.keys.enabled ? '已有启用密钥；遗失需重新签发' : '下一步：签发应用密钥';
    element('step-test').textContent = value.call_evidence ? '曾成功，可核对记录' : '尚待核验';
    const completed = [upstream.configured && upstream.enabled, value.callable_models > 0, value.keys.enabled > 0, !!overviewEvidence, !!overviewEvidence];
    renderSetup(completed);
  
}
element('refresh-overview').onclick = () => navigate('overview');
element('overview-record').onclick = () => { if (overviewEvidence) return navigate('records', true, overviewEvidence.request_id); };
return {render,dispose:clearVersion,clearVersion,loadVersion,clearEvidence(){overviewEvidence=null;}};
}
