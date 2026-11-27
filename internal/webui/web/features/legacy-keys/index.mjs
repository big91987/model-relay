import {known} from '../../shared/legacy-presentation.mjs';
export function createLegacyKeys({element,api,current,assertCurrent,guarded,item,restoreRowFocus,showEmpty,form,load,message}) {
let issuedEpoch=0;
function focusError(error) {
  if(error.code==='invalid_name')element('key-name').focus();
}
function closeKey() { issuedEpoch++; element('issued-key').value = ''; element('issued-key').type = 'password'; element('show-key').textContent = '显示新密钥'; element('issued').hidden = true; }
async function render(token,pageCursor='') {

    const list = await api('keys', 'GET', undefined, token);
    element('key-list').replaceChildren(...list.map(value => item(value.name + ' · ' + value.hint + ' · 标识 ' + value.id + ' · 创建时间 ' + known(value.created) + (value.enabled ? ' · 启用' : ' · 停用'), (value.enabled ? '停用密钥 ' : '启用密钥 ') + value.name, async (own, focused) => {
      if (!confirm('确认变更密钥状态？')) return;
      await api('keys/' + encodeURIComponent(value.id), 'PATCH', {enabled: !value.enabled}, own); await load('keys', own); restoreRowFocus('key-list', value.id, focused, own); message('密钥状态已更新，需重新核验当前配置');
    }, value.id,focusError)));
    if (!list.length) showEmpty('key-list', '尚无应用密钥。填写用途名称后签发，完整秘密只展示一次。调用前还需上游与模型。');
  
}
element('cancel-key-create').onclick = () => { element('key-create').open = false; element('key-create').querySelector('summary')?.focus(); };
form('key-form', async token => {
  closeKey(); let value; element('cancel-key-create').disabled = true;
  try { value = await api('keys', 'POST', {name: element('key-name').value}, token); }
  catch (error) { if (error.stale || error.code) throw error; throw new Error('签发结果未确认，请刷新密钥列表后再决定是否重新签发，并停用遗失值。不会自动重签。'); }
  finally { if (current(token)) element('cancel-key-create').disabled = false; }
  element('key-create').open = false; element('issued-key').value = value.key; element('issued').hidden = false; element('issued-key').focus(); message('密钥已签发，请立即保管；关闭后不再返回。'); await load('keys', token);
},focusError);
element('show-key').onclick = () => { const visible = element('issued-key').type === 'password'; element('issued-key').type = visible ? 'text' : 'password'; element('show-key').textContent = visible ? '隐藏新密钥' : '显示新密钥'; };
element('close-key').onclick = () => { closeKey(); element('key-create').querySelector('summary')?.focus(); };
element('copy-key').onclick = () => guarded(element('copy-key'), async token => { const epoch = issuedEpoch; try { await navigator.clipboard.writeText(element('issued-key').value); assertCurrent(token); if (epoch === issuedEpoch && !element('issued').hidden) message('已复制，请保管'); } catch (error) { assertCurrent(token); if (epoch === issuedEpoch && !element('issued').hidden) message('自动复制失败，请在当前展示中手动选择复制。', true); } });
return {render,dispose:closeKey,clearSecret:closeKey};
}
