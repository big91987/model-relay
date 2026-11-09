import {errors,friendly} from '../../shared/legacy-presentation.mjs';
import {allPages as readAllPages} from '../../shared/legacy-data.mjs';
export function createLegacyPlayground({document,element,api,owner,current,assertCurrent,stale,guarded,navigate,message,scheduleSession,expireSession,isSessionUncertain}) {
let call=null,selectedModel='',lastRequestID='',testReady=false;
const allPages=(path,token)=>readAllPages(api,path,token);
function callInputs(disabled) {
  for (const id of ['test-model', 'test-key', 'test-message', 'test-stream']) element(id).disabled = disabled;
}
function bounded(text, limit) { return text.length > limit ? text.slice(0, limit) + '\n[展示已截断；原始响应可能更长]' : text; }
function appendRaw(text) { element('test-raw').textContent = bounded(element('test-raw').textContent + text, 262144); }
function appendReadable(value, streaming, operation) {
  let text = '';
  for (const [choicePosition, choice] of (value.choices || []).entries()) {
    const result = streaming ? choice.delta : choice.message;
    const content = result?.content;
    if (typeof content === 'string') text += content;
    for (const [toolPosition, tool] of (result?.tool_calls || []).entries()) {
      const identity = (choice.index ?? choicePosition) + ':' + (streaming ? tool.index ?? toolPosition : toolPosition);
      if (!operation.tools.has(identity)) {
        if (operation.tools.size >= 1024) { operation.toolsTruncated = true; continue; }
        operation.tools.set(identity, '');
      }
      const name = tool.function?.name;
      if (typeof name === 'string') operation.tools.set(identity, (operation.tools.get(identity) + name).slice(0, 128));
    }
  }
  element('test-output').textContent = bounded(element('test-output').textContent + text, 262144);
}
function clearSecrets() {
  for (const id of ['test-key', 'test-message']) element(id).value = '';
  element('test-output').textContent = ''; element('test-raw').textContent = ''; element('diagnostics').open = false;
}
function stopCall(canceled = false) {
  const previous = call; call = null;
  if (previous) { previous.controller.abort(); if (previous.reader) previous.reader.cancel().catch(() => {}); }
  clearSecrets(); callInputs(false); element('cancel').disabled = true; element('send').disabled = !testReady || isSessionUncertain();
  if (canceled) element('test-status').textContent = '已取消；本地取消不代表服务端已记为取消，请核对记录。';
}
async function render(token,pageCursor='') {

    testReady = false; element('send').disabled = true;
    const [models, overview] = await Promise.all([allPages('models', token), api('overview', 'GET', undefined, token)]);
    assertCurrent(token);
    const enabled = models.filter(value => value.enabled && value.candidates.some(candidate => candidate.enabled));
    element('test-model').replaceChildren(...enabled.map(value => { const option = document.createElement('option'); option.value = value.alias; option.textContent = value.alias; return option; }));
    if (enabled.some(value => value.alias === selectedModel)) element('test-model').value = selectedModel;
    else if (selectedModel) { const placeholder = document.createElement('option'); placeholder.value = ''; placeholder.textContent = '原选模型已不可用，请重新选择'; element('test-model').append(placeholder); element('test-model').value = ''; }
    testReady = !!(overview.callable_models && enabled.length && overview.upstream.enabled && overview.upstream.credential_configured && overview.keys.enabled && !overview.degraded);
    element('send').disabled = !testReady || isSessionUncertain();
    const missing = [];
    if (!overview.upstream.enabled || !overview.upstream.credential_configured) missing.push('配置并启用上游');
    if (!enabled.length) missing.push('创建或启用模型');
    if (!overview.keys.enabled) missing.push('签发或启用密钥');
    element('test-dependency').textContent = overview.degraded ? errors.storage_unavailable : missing.length ? '调用依赖尚缺：' + missing.join('；') : '依赖已具备，请输入应用密钥和消息；由实际请求核验。';
    element('test-status').textContent = isSessionUncertain() ? '管理会话暂无法确认，请重新检查会话或重新登录。' : '准备就绪';
    element('gateway-base').textContent = location.origin + '/v1';
    element('client-example').textContent = 'GATEWAY_BASE_URL=\'' + location.origin + '/v1\'\ncurl "$GATEWAY_BASE_URL/chat/completions" \\\n  -H "Authorization: Bearer <APPLICATION_KEY>" \\\n  -H "Content-Type: application/json" \\\n  -d \'{"model":"<MODEL_ALIAS>","messages":[{"role":"user","content":"你好"}]}\'';
    element('test-request-id').textContent = ''; element('view-record').hidden = true;
  
}
element('copy-example').onclick = () => guarded(element('copy-example'), async token => { try { await navigator.clipboard.writeText(element('client-example').textContent); assertCurrent(token); message('占位示例已复制，请在客户端替换占位符。'); } catch { assertCurrent(token); message('自动复制失败，请手动选择调用示例。', true); } });
element('view-record').onclick = () => navigate('records', true, lastRequestID);
element('cancel').onclick = () => stopCall(true);
element('test-form').onsubmit = event => {
  event.preventDefault(); if (call || !testReady || isSessionUncertain()) return;
  if (expireSession()) return;
  const token = owner(); const operation = {token, controller: new AbortController(), tools: new Map(), toolsTruncated: false}; call = operation; scheduleSession();
  const streaming = element('test-stream').checked;
  const payload = {model: element('test-model').value, messages: [{role: 'user', content: element('test-message').value}], stream: streaming};
  const key = element('test-key').value;
  callInputs(true);
  element('send').disabled = true; element('cancel').disabled = false; element('test-output').textContent = ''; element('test-raw').textContent = ''; element('test-request-id').textContent = '尚未取得网关请求 ID'; element('view-record').hidden = true; lastRequestID = '';
  element('test-status').textContent = '等待首包'; message('');
  const owns = () => call === operation && current(token);
  const ensure = () => { if (!owns()) throw stale(); };
  return (async () => {
    try {
      const response = await fetch('/v1/chat/completions', {method: 'POST', signal: operation.controller.signal, headers: {'Content-Type': 'application/json', Authorization: 'Bearer ' + key}, body: JSON.stringify(payload)});
      ensure(); lastRequestID = response.headers.get('X-Request-ID') || ''; element('test-request-id').textContent = lastRequestID ? '请求 ID：' + lastRequestID : '未取得网关请求 ID，无法定位同请求记录'; element('view-record').hidden = !lastRequestID;
      if (!response.ok) { const value = await response.json(); ensure(); appendRaw(JSON.stringify(value, null, 2)); throw new Error(friendly(value.error?.code)); }
      if (!streaming) {
        const value = await response.json(); ensure(); appendRaw(JSON.stringify(value, null, 2)); appendReadable(value, false, operation);
      } else {
        element('test-status').textContent = '接收中';
        const reader = response.body.getReader(); operation.reader = reader; const decoder = new TextDecoder(); let buffer = '', done = false;
        const consume = () => {
          buffer = buffer.replace(/\r\n/g, '\n');
          let boundary;
          while ((boundary = buffer.indexOf('\n\n')) >= 0) {
            const frame = buffer.slice(0, boundary); buffer = buffer.slice(boundary + 2);
            const data = frame.split('\n').filter(line => line.startsWith('data:')).map(line => line.slice(5).trimStart()).join('\n');
            if (!data) continue;
            appendRaw(frame + '\n\n');
            if (data === '[DONE]') { done = true; continue; }
            const value = JSON.parse(data);
            if (value.error) throw new Error(friendly(value.error.code || 'stream_interrupted'));
            appendReadable(value, true, operation);
          }
          if (buffer.length > 262144) throw new Error(friendly('upstream_protocol_error'));
        };
        while (!done) { const chunk = await reader.read(); ensure(); buffer += decoder.decode(chunk.value || new Uint8Array(), {stream: !chunk.done}); consume(); if (chunk.done) break; }
        if (!done) throw new Error(friendly('stream_interrupted'));
        await reader.cancel(); ensure();
      }
      if (operation.tools.size) {
        const names = [...operation.tools.values()].slice(0, 5).map(name => name || '名称未提供');
        const summary = '返回' + (operation.toolsTruncated ? '至少' : '') + operation.tools.size + '个工具调用：' + names.join('、') + (operation.tools.size > 5 ? '（名称仅展示前5个）' : '') + '。未执行工具；详情见原始响应诊断。';
        element('test-output').textContent = bounded(element('test-output').textContent + '\n' + summary, 262144);
      } else if (!element('test-output').textContent) element('test-output').textContent = '调用已完成，但没有文本输出或工具调用。请展开原始响应诊断。';
      element('test-status').textContent = '完成；请核对本次记录中的实际或未知用量。';
    } catch (error) {
      if (owns()) { operation.controller.abort(); if (operation.reader) operation.reader.cancel().catch(() => {}); element('test-key').value = ''; element('test-message').value = ''; element('test-status').textContent = '失败（已有输出不自动重放）'; message((error instanceof SyntaxError ? friendly('upstream_protocol_error') : Object.values(errors).includes(error.message) ? error.message : '无法读取网关响应，请检查服务。未取得 ID 时无法定位记录。') + ' 应用密钥与消息已清除，请重新输入后手动发送。', true); }
    } finally {
      if (owns()) { call = null; callInputs(false); element('send').disabled = !testReady || isSessionUncertain(); element('cancel').disabled = true; scheduleSession(); }
    }
  })();
};
return {render,stop:stopCall,dispose(){selectedModel=element('test-model').value||selectedModel;stopCall();},get running(){return !!call;},get ready(){return testReady;},reset(){selectedModel='';lastRequestID='';element('test-model').value='';element('test-request-id').textContent='';element('view-record').hidden=true;},clearSecret:clearSecrets};
}
