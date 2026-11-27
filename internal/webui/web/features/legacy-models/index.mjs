import {allPages as readAllPages} from '../../shared/legacy-data.mjs';
export function createLegacyModels({document,element,api,current,guarded,action,item,restoreRowFocus,showEmpty,form,load,message}) {
let modelCreateOperation=null,editModel=null,editDraft=null,editSaving=false,upstreamItems=[],modelItems=[],modelCursor=null;
function focusError(error) {
  if(error.code==='invalid_model'||error.code==='alias_conflict')element('model-alias').focus();
}
const allPages=(path,token)=>readAllPages(api,path,token);
function modelCreateControls(disabled) { for (const id of ['model-alias', 'cancel-model-create', 'add-create-candidate']) element(id).disabled = disabled; for (const input of element('model-form').querySelectorAll('input,select,button')) input.disabled = disabled; }
function resetModelCreate() { modelCreateOperation = null; modelCreateControls(false); }
function cleanCandidates(values) { return values.map(({upstream_id, remote_model, enabled, priority, weight}) => ({upstream_id, remote_model, enabled, priority, weight})); }
function defaultCandidate() { return {upstream_id: upstreamItems[0]?.id || 0, remote_model: '', enabled: true, priority: 0, weight: 1}; }
function candidateName(value) { return value.remote_model + ' · ' + (upstreamItems.find(source => source.id === value.upstream_id)?.name || '上游 ' + value.upstream_id) + ' · ' + (value.enabled ? '参与' : '停用') + ' · 优先级 ' + value.priority + ' / 权重 ' + value.weight + (value.generation_health ? ' · ' + value.generation_health : ' · 本次运行生成健康未知'); }
function renderCandidates(mode, values) {
  const root = element(mode === 'create' ? 'create-candidates' : 'edit-candidates');
  root.replaceChildren(...values.map((value, index) => {
    const row = document.createElement('fieldset'); const legend = document.createElement('legend'); legend.textContent = '路由候选 ' + (index + 1); row.append(legend);
    const select = document.createElement('select'); select.dataset.field = 'upstream_id'; select.required = true;
    for (const upstream of upstreamItems) { const option = document.createElement('option'); option.value = upstream.id; option.textContent = upstream.name + (upstream.enabled ? '' : '（停用）'); select.append(option); }
    select.id = mode + '-candidate-upstream-' + index; select.value = value.upstream_id; const label = document.createElement('label'); label.textContent = '上游'; label.htmlFor = select.id; row.append(label, select);
    for (const [field, title, type] of [['remote_model', '远端模型名称', 'text'], ['priority', '优先级（大数先用）', 'number'], ['weight', '同级权重', 'number']]) {
      const input = document.createElement('input'); input.dataset.field = field; input.type = type; input.required = true; input.value = value[field];
      if (field === 'remote_model') { input.maxLength = 128; if (!index) input.id = mode === 'create' ? 'model-real' : 'edit-real'; }
      else { input.min = field === 'weight' ? '1' : '0'; input.max = '100'; input.step = '1'; }
      const label = document.createElement('label'); label.textContent = title; label.append(input); row.append(label);
    }
    const enabled = document.createElement('input'); enabled.type = 'checkbox'; enabled.dataset.field = 'enabled'; enabled.checked = value.enabled;
    const enabledLabel = document.createElement('label'); enabledLabel.className = 'check'; enabledLabel.textContent = '参与路由'; enabledLabel.append(enabled); row.append(enabledLabel);
    row.append(action('移除此候选', () => { const next = readCandidates(mode); next.splice(index, 1); renderCandidates(mode, next); root.querySelector('input,select')?.focus(); }));
    return row;
  }));
}
function readCandidates(mode) {
  return [...element(mode === 'create' ? 'create-candidates' : 'edit-candidates').children].map(row => {
    const value = {}; for (const input of row.querySelectorAll('[data-field]')) value[input.dataset.field] = input.type === 'checkbox' ? input.checked : input.type === 'number' || input.tagName === 'SELECT' ? Number(input.value) : input.value; return value;
  });
}
async function render(token,pageCursor='') {

    const [page, sources] = await Promise.all([api('models' + (pageCursor ? '?cursor=' + encodeURIComponent(pageCursor) : ''), 'GET', undefined, token), allPages('upstreams', token)]);
    upstreamItems = sources; modelItems = page.data; modelCursor = page.next_cursor;
    const list = modelItems;
    element('model-list').replaceChildren(...list.map(value => {
      const row = item(value.alias + ' → ' + (value.candidates.length ? value.candidates.map(candidate => candidateName(candidate)).join('；') : '尚未关联上游' + (value.legacy_remote_model ? '（已保留原模型名 ' + value.legacy_remote_model + '）' : '')) + (value.enabled ? ' · 启用' : ' · 停用'), (value.enabled ? '停用模型 ' : '启用模型 ') + value.alias, async (own, focused) => {
        if (!confirm('确认变更模型状态？')) return;
        await api('models/' + encodeURIComponent(value.alias), 'PATCH', {expected_revision: value.revision, enabled: !value.enabled, candidates: cleanCandidates(value.candidates)}, own); await load('models', own); restoreRowFocus('model-list', value.alias, focused, own); message('模型状态已更新，需重新核验当前配置');
      }, value.alias,focusError);
      row.append(action('编辑映射 ' + value.alias, () => { if (editSaving) { message('映射正在保存，请等待结果。'); return; } editModel = value; element('edit-alias').value = value.alias; element('edit-routing-strategy').value = value.routing_strategy || 'weighted'; renderCandidates('edit', editDraft?.alias === value.alias ? editDraft.candidates : value.candidates); element('model-edit').hidden = false; element('edit-candidates').querySelector('input,select')?.focus(); })); return row;
    }));
    if (!list.length) showEmpty('model-list', '尚无模型。填写别名与供应商模型名称，创建首个映射；不会创建演示对象。');
    if (!element('create-candidates').children.length) renderCandidates('create', upstreamItems.length ? [defaultCandidate()] : []);
    else if (!modelCreateOperation) renderCandidates('create', readCandidates('create'));
    element('more-models').hidden = !modelCursor;
    element('models-dependency').textContent = sources.some(value => value.credential_configured && value.enabled) ? '上游已保存并启用，模型保存后可进入调用测试。' : '依赖尚缺：调用前请先配置并启用上游。可以保留模型草稿。';
  
}
element('more-models').onclick = () => guarded(element('more-models'), token => load('models', token, modelCursor),focusError);
for (const mode of ['create', 'edit']) element('add-' + mode + '-candidate').onclick = () => { if (editSaving) return; if (!upstreamItems.length) { message('请先接入一个真实上游。', true); return; } const values = readCandidates(mode); values.push(defaultCandidate()); renderCandidates(mode, values); };
form('model-form', async token => { if (modelCreateOperation && current(modelCreateOperation.token)) return; const operation = {token}; modelCreateOperation = operation; const input = {alias: element('model-alias').value, routing_strategy: element('create-routing-strategy').value, candidates: readCandidates('create'), enabled: true}; modelCreateControls(true); let saved = false; try { await api('models', 'POST', input, token); saved = true; element('model-create').open = false; await load('models', token); message('模型已保存；保存不代表生成成功'); } catch (error) { if (!error.stale && current(token) && saved) error.message = '模型已保存，但最新列表读取未确认；请重新读取，不会重复创建。' + error.message; throw error; } finally { if (modelCreateOperation === operation && current(token)) resetModelCreate(); } },focusError);
element('cancel-model-create').onclick = () => { if (modelCreateOperation) return; element('model-create').open = false; element('model-create').querySelector('summary')?.focus(); };
form('model-edit-form', async token => {
  if (!editModel || editSaving) return;
  editSaving = true; element('cancel-edit').disabled = true;
  const editing = editModel;
  for (const input of element('model-edit-form').querySelectorAll('input,select,button')) input.disabled = true;
  try { await api('models/' + encodeURIComponent(editing.alias), 'PATCH', {expected_revision: editing.revision, enabled: editing.enabled, routing_strategy: element('edit-routing-strategy').value, candidates: readCandidates('edit')}, token); element('model-edit').hidden = true; editModel = null; editDraft = null; await load('models', token); message('映射已保存，需重新核验'); }
  finally { if (current(token)) { editSaving = false; element('cancel-edit').disabled = false; for (const input of element('model-edit-form').querySelectorAll('input,select,button')) input.disabled = false; } }
},focusError);
element('cancel-edit').onclick = () => { if (editSaving) return; const alias = editModel?.alias; if (alias) editDraft = {alias, candidates: readCandidates('edit')}; editModel = null; element('model-edit').hidden = true; const button = [...element('model-list').querySelectorAll('button')].find(value => value.textContent === '编辑映射 ' + alias); if (button) button.focus(); };
return {render,dispose(){resetModelCreate();editModel=null;editDraft=null;editSaving=false;element('model-edit').hidden=true;}};
}
