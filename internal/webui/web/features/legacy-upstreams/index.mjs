import {friendly,known} from '../../shared/legacy-presentation.mjs';
export function createLegacyUpstreams({element,api,current,guarded,item,showEmpty,form,load,message}) {
let upstreamOperation=null,selectedUpstream=null,upstreamItems=[],upstreamCursor=null;
function focusError(error) {
  if(error.code==='invalid_destination')element('upstream-url').focus();
  else if(error.code==='new_credential_required')element('upstream-token').focus();
}
function upstreamControls(disabled) {
  for (const id of ['upstream-name', 'upstream-url', 'upstream-token', 'upstream-enabled', 'save-upstream', 'refresh-upstream', 'new-upstream', 'cancel-upstream', 'check-upstream']) element(id).disabled = disabled;
  for (const button of element('upstream-list').querySelectorAll('button')) button.disabled = disabled;
  if (!disabled) element('check-upstream').disabled = !selectedUpstream?.enabled;
}
function resetUpstream() { upstreamOperation?.controller.abort(); upstreamOperation = null; upstreamControls(false); element('cancel-check').hidden = true; }
async function upstreamAction(token, callback) {
  if (upstreamOperation && current(upstreamOperation.token)) return;
  const operation = {token, controller: new AbortController()}; upstreamOperation = operation; upstreamControls(true);
  try { await callback(); }
  finally { if (upstreamOperation === operation && current(token)) { upstreamOperation = null; upstreamControls(false); element('cancel-check').hidden = true; } }
}
function setUpstream(value) {
  selectedUpstream = value;
  element('upstream-name').value = value?.name || '';
  element('upstream-url').value = value?.url || ''; element('upstream-enabled').checked = value ? value.enabled : true;
  element('upstream-token').value = '';
  element('upstream-editor-title').textContent = value ? '编辑上游 · ' + value.name : '新建上游';
  element('credential-state').textContent = value?.credential_configured ? '凭据已配置（不返回原值）；保存不代表调用成功' : '尚无凭据，首次保存需输入';
  element('check-state').textContent = value?.check ? (value.check.applicable ? '' : '历史配置 · ') + ({success: '连接核验成功（不是生成成功）', failed: '连接核验失败', unknown: '核验未知', canceled: '核验已取消', checking: '核验进行中'}[value.check.state] || '核验未知') + ' · ' + known(value.check.finished_at) + (value.check.error_code ? ' · ' + friendly(value.check.error_code) : '') : '未核验';
  element('check-upstream').disabled = !value || !value.enabled;
}
function renderUpstreams(token) {
  element('upstream-list').replaceChildren(...upstreamItems.map(value => {
    const row = item(value.name + ' · ' + (value.enabled ? '启用' : '停用') + ' · 凭据' + (value.credential_configured ? '已配置' : '缺失'), '编辑上游 ' + value.name, own => upstreamAction(own, async () => { setUpstream(await api('upstreams/' + value.id, 'GET', undefined, own)); element('upstream-name').focus(); }), String(value.id),focusError);
    return row;
  }));
  if (!upstreamItems.length) showEmpty('upstream-list', '尚无上游。新建真实接入配置；不会创建演示供应商。');
  element('more-upstreams').hidden = !upstreamCursor;
}
async function render(token,pageCursor='') {

    const page = await api('upstreams' + (pageCursor ? '?cursor=' + encodeURIComponent(pageCursor) : ''), 'GET', undefined, token);
    upstreamItems = page.data; upstreamCursor = page.next_cursor; renderUpstreams(token);
    const selected = upstreamItems.find(value => value.id === selectedUpstream?.id);
    if (selected) setUpstream(selected);
    else if (selectedUpstream && pageCursor) setUpstream(await api('upstreams/' + selectedUpstream.id, 'GET', undefined, token));
    else setUpstream(upstreamItems[0] || null);
  
}
element('new-upstream').onclick = () => { if (upstreamOperation) return; setUpstream(null); element('upstream-name').focus(); };
element('cancel-upstream').onclick = () => { if (upstreamOperation) return; element('upstream-token').value = ''; const identity = selectedUpstream?.id; if (selectedUpstream) setUpstream(selectedUpstream); else setUpstream(null); const row = [...element('upstream-list').children].find(value => value.dataset?.identity === String(identity)); (row?.querySelector('button') || element('new-upstream')).focus(); };
element('check-upstream').onclick = () => guarded(element('check-upstream'), token => upstreamAction(token, async () => {
  if (!selectedUpstream) return; const value = selectedUpstream; element('check-state').textContent = '核验进行中（不发起生成）';
  const operation = upstreamOperation; element('cancel-check').hidden = false;
  try {
    await api('upstreams/' + value.id + '/check', 'POST', {expected_execution_revision: value.execution_revision}, token, operation.controller.signal);
    if (operation.controller.signal.aborted || upstreamOperation !== operation) return;
    element('cancel-check').hidden = true;
    await load('upstream', token); message('连接核验已完成；实际生成及费用需另行核对。');
  } catch (error) {
    if (!operation.controller.signal.aborted) throw error;
  }
}),focusError);
element('cancel-check').onclick = () => { if (!upstreamOperation) return; upstreamOperation.controller.abort(); element('cancel-check').hidden = true; element('check-state').textContent = '核验已取消；最终状态可重新读取，不会自动重试。'; };
element('more-upstreams').onclick = () => guarded(element('more-upstreams'), token => load('upstream', token, upstreamCursor),focusError);
form('upstream-form', async token => upstreamAction(token, async () => {
  let saved = false;
  try {
    const input = {name: element('upstream-name').value, url: element('upstream-url').value, token: element('upstream-token').value, enabled: element('upstream-enabled').checked};
    if (selectedUpstream) input.expected_revision = selectedUpstream.revision;
    selectedUpstream = await api(selectedUpstream ? 'upstreams/' + selectedUpstream.id : 'upstreams', selectedUpstream ? 'PATCH' : 'POST', input, token);
    saved = true; element('upstream-token').value = '';
    await load('upstream', token); message('上游已保存；不代表联调成功。');
  } catch (error) {
    if (!error.stale && current(token)) {
      element('upstream-token').value = '';
      if (saved) element('credential-state').textContent = '最新状态未确认，请重新读取上游；凭据不返回原值。';
      error.message = (saved ? '上游保存已受理、已保存，但最新读取未确认；请重新读取，不会自动重写。' : '保存未确认；上游凭据已清除，请重新输入。') + ' ' + error.message;
    }
    throw error;
  }
}),focusError);
element('refresh-upstream').onclick = () => guarded(element('refresh-upstream'), token => upstreamAction(token, async () => {
  element('upstream-token').value = '';
  try { await load('upstream', token); message('已重新读取上游状态；保存不代表调用成功。已撤销本页未保存修改。'); }
  catch (error) { if (!error.stale && current(token)) { element('credential-state').textContent = '恢复未确认，保留公开草稿；请重新读取服务器状态。'; error.message = '恢复未确认；输入凭据已清除，公开草稿保留。' + error.message; } throw error; }
}),focusError);
return {render,dispose(){resetUpstream();element('upstream-token').value='';},clearSecret(){element('upstream-token').value='';}};
}
