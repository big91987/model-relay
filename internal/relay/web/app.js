'use strict';
const element = id => document.getElementById(id);
let csrf = '', active = 'overview', aborter = null, nextCursor = null, navigationVersion = 0;
function message(text, failed = false) { element(failed ? 'error' : 'notice').textContent = text; element(failed ? 'notice' : 'error').textContent = ''; }
async function api(path, method = 'GET', body) {
  const response = await fetch('/admin/api/' + path, {method, headers: {'Content-Type': 'application/json', 'X-CSRF-Token': csrf}, body: body === undefined ? undefined : JSON.stringify(body)});
  const value = await response.json();
  if (!response.ok) { if (response.status === 401) signedOut(); throw new Error(value.error?.code || '请求失败'); }
  return value;
}
function clearSecrets() { for (const id of ['password', 'upstream-token', 'issued-key', 'test-key', 'test-message']) element(id).value = ''; element('issued').hidden = true; element('test-output').textContent = ''; }
function signedOut() { if (aborter) aborter.abort(); csrf = ''; clearSecrets(); element('console').hidden = true; element('login').hidden = false; }
function signedIn() { element('password').value = ''; element('console').hidden = false; element('login').hidden = true; }
function item(text, action, callback) { const row = document.createElement('div'); row.className = 'item'; const span = document.createElement('span'); span.textContent = text; row.append(span); const button = document.createElement('button'); button.textContent = action; button.onclick = () => guarded(button, callback); row.append(button); return row; }
async function guarded(button, callback) { button.disabled = true; try { await callback(); } catch (error) { message(error.message, true); } finally { button.disabled = false; } }
async function load(view) {
  if (view === 'overview' || view === 'upstream') { const value = await api('upstream'); if (view === 'overview') { const status = await api('session'); element('summary').textContent = status.degraded ? '存储故障：已阻止调用，请停止并修复后重启。' : value.credential_configured ? (value.enabled ? '上游已启用 · 配置已保存，尚需调用核验' : '上游已停用') : '待配置：先接入上游，再创建模型与应用密钥'; } else { element('upstream-url').value = value.url; element('upstream-enabled').checked = value.enabled; element('credential-state').textContent = value.credential_configured ? '凭据已配置（不返回原值）' : '尚无凭据'; } }
  if (view === 'models') { const list = await api('models'); element('model-list').replaceChildren(...list.map(value => { const row = item(value.alias + ' → ' + value.model + (value.enabled ? ' · 启用' : ' · 停用'), value.enabled ? '停用模型 ' + value.alias : '启用模型 ' + value.alias, async () => { if (!confirm('确认变更模型状态？')) return; await api('models/' + encodeURIComponent(value.alias), 'PATCH', {...value, enabled: !value.enabled}); await load('models'); }); const edit = document.createElement('button'); edit.textContent = '编辑映射 ' + value.alias; edit.onclick = () => guarded(edit, async () => { const model = prompt('新的真实模型名', value.model); if (model === null) return; await api('models/' + encodeURIComponent(value.alias), 'PATCH', {...value, model}); await load('models'); message('映射已保存'); }); row.append(edit); return row; })); }
  if (view === 'keys') { const list = await api('keys'); element('key-list').replaceChildren(...list.map(value => item(value.name + ' · ' + value.hint + (value.enabled ? ' · 启用' : ' · 停用'), value.enabled ? '停用密钥 ' + value.name : '启用密钥 ' + value.name, async () => { if (!confirm('确认变更密钥状态？')) return; await api('keys/' + value.id, 'PATCH', {enabled: !value.enabled}); await load('keys'); }))); }
  if (view === 'records') { const value = await api('requests' + (nextCursor ? '?cursor=' + nextCursor : '')); nextCursor = value.next_cursor; element('more-records').hidden = !nextCursor; element('record-list').replaceChildren(...value.data.map(record => { const row = document.createElement('tr'); for (const text of [record.id + ' / ' + record.time, record.alias + ' / ' + record.key_id, record.state + (record.error_code ? ' · ' + record.error_code : ''), record.duration_ms ?? '未知', [record.prompt_tokens, record.completion_tokens, record.total_tokens].map(value => value ?? '未知').join(' / ')]) { const cell = document.createElement('td'); cell.textContent = text; row.append(cell); } return row; })); }
}
async function navigate(view) {
  const version = ++navigationVersion;
  if (aborter) aborter.abort();
  if (active === 'test' || active === 'keys' || active === 'upstream') clearSecrets();
  active = view;
  for (const panel of document.querySelectorAll('.panel')) panel.hidden = true;
  nextCursor = null;
  message('加载中');
  try { await load(view); if (version === navigationVersion) message(''); }
  catch (error) { if (version === navigationVersion) message(error.message, true); }
  finally { if (version === navigationVersion) element(view).hidden = false; }
}
for (const button of document.querySelectorAll('[data-view]')) button.onclick = () => navigate(button.dataset.view);
function form(id, callback) { element(id).onsubmit = event => { event.preventDefault(); guarded(event.submitter, callback); }; }
form('login-form', async () => { const value = await api('login', 'POST', {password: element('password').value}); csrf = value.csrf; signedIn(); await navigate('overview'); });
form('upstream-form', async () => { await api('upstream', 'PUT', {url: element('upstream-url').value, token: element('upstream-token').value, enabled: element('upstream-enabled').checked}); element('upstream-token').value = ''; await load('upstream'); message('上游已保存；不代表联调成功。'); });
form('model-form', async () => { await api('models', 'POST', {alias: element('model-alias').value, model: element('model-real').value, enabled: true}); await load('models'); message('模型已保存'); });
form('key-form', async () => { element('issued-key').value = ''; element('issued').hidden = true; const value = await api('keys', 'POST', {name: element('key-name').value}); element('issued-key').value = value.key; element('issued').hidden = false; await load('keys'); message('密钥已签发，请立即保管'); });
element('close-key').onclick = () => { element('issued-key').value = ''; element('issued').hidden = true; };
element('copy-key').onclick = () => guarded(element('copy-key'), async () => { await navigator.clipboard.writeText(element('issued-key').value); message('已复制，请保管'); });
element('logout').onclick = () => guarded(element('logout'), async () => { await api('logout', 'POST', {}); signedOut(); message('已退出'); });
element('refresh-records').onclick = () => { nextCursor = null; guarded(element('refresh-records'), () => load('records')); };
element('more-records').onclick = () => guarded(element('more-records'), () => load('records'));
element('cancel').onclick = () => { if (aborter) aborter.abort(); clearSecrets(); };
form('test-form', async () => {
  message('');
  aborter = new AbortController(); const controller = aborter; element('cancel').disabled = false; element('test-output').textContent = ''; element('test-status').textContent = '等待首包'; const streaming = element('test-stream').checked;
  try {
    const response = await fetch('/v1/chat/completions', {method: 'POST', signal: controller.signal, headers: {'Content-Type': 'application/json', Authorization: 'Bearer ' + element('test-key').value}, body: JSON.stringify({model: element('test-model').value, messages: [{role: 'user', content: element('test-message').value}], stream: streaming})});
    if (!response.ok) { const value = await response.json(); throw new Error(value.error?.code || '调用失败'); }
    if (!streaming) { const value = await response.json(); element('test-output').textContent = JSON.stringify(value, null, 2); element('test-status').textContent = '完成'; return; }
    element('test-status').textContent = '接收中'; const reader = response.body.getReader(); const decoder = new TextDecoder(); let buffer = '', complete = false;
    while (true) { const part = await reader.read(); buffer += decoder.decode(part.value || new Uint8Array(), {stream: !part.done}).replace(/\r\n/g, '\n'); let boundary; while ((boundary = buffer.indexOf('\n\n')) >= 0) { const event = buffer.slice(0, boundary); buffer = buffer.slice(boundary + 2); if (event.startsWith('event: error')) throw new Error('流中断'); const data = event.split('\n').filter(line => line.startsWith('data:')).map(line => line.slice(5).trimStart()).join('\n'); if (!data) continue; if (data === '[DONE]') { complete = true; continue; } const value = JSON.parse(data); if (value.error) throw new Error('流中断'); for (const choice of value.choices || []) element('test-output').textContent += choice.delta?.content || (choice.delta?.tool_calls ? JSON.stringify(choice.delta.tool_calls) : ''); } if (part.done) break; }
    if (!complete) throw new Error('流不完整：缺少结束标记'); element('test-status').textContent = '完成';
  } catch (error) { if (error.name === 'AbortError') { element('test-status').textContent = '已取消'; clearSecrets(); } else { element('test-status').textContent = '不完整 / 失败'; message(error.message, true); } } finally { element('cancel').disabled = true; if (aborter === controller) aborter = null; }
});
api('session').then(async value => { csrf = value.csrf; signedIn(); await navigate('overview'); }).catch(error => { signedOut(); if (error.message !== 'unauthorized') message(error.message, true); });
