'use strict';
const element = id => document.getElementById(id);
const views = ['overview', 'upstream', 'models', 'keys', 'test', 'records'];
let csrf = '', active = 'overview', nextCursor = null, navigationVersion = 0, sessionEpoch = 0;
let versionRequest = null, sessionTimer = null, expiry = 0, call = null, selectedModel = '', lastRequestID = '', editModel = null, testReady = false;
let loadSequence = 0, findSequence = 0, issuedEpoch = 0;
let editDraft = null;
let editSaving = false;
let sessionUncertain = false, sessionCheck = null;
let sessionExpiryTimer = null, overviewEvidence = null;
let upstreamOperation = null;
let modelCreateOperation = null;
let selectedUpstream = null, upstreamItems = [], upstreamCursor = null, modelCursor = null, modelItems = [];
const errors = {
  management_api_changed: '管理界面已升级，请刷新页面后继续。',
  configuration_conflict: '配置已被其他操作修改，请重新读取后检查草稿；不会自动覆盖。',
  name_conflict: '该上游名称已存在，请读取列表确认，不会自动重复创建。',
  invalid_route_candidate: '请检查路由候选的上游、模型名称、优先级和权重。',
  check_in_progress: '该上游正在核验，请等待结果后重新读取。',
  upstream_attempts_exhausted: '可用上游的有限尝试已耗尽，请查看逐尝试记录；不会无限重发。',
  request_timeout: '整个请求的时间预算已用尽，请查看记录；不会开始新的尝试。',
  unauthorized: '应用密钥缺失、无效或已停用，请前往应用密钥确认并重新输入。',
  invalid_destination: '上游地址不符合安全规则，请检查 HTTPS 地址和主机。',
  new_credential_required: '首次配置或地址变化时必须提供新的上游凭据。',
  invalid_model: '请使用有效的公共别名，并填写真实模型名称。',
  alias_conflict: '该模型别名已存在，请编辑现有映射或使用不同别名。',
  invalid_name: '请输入有效名称，长度不超过128字符。',
  upstream_unavailable: '上游未配置或已停用，请前往上游配置。',
  model_unavailable: '模型不存在或已停用，请前往模型页面恢复。',
  upstream_rejected: '供应商拒绝调用，请核对供应商凭据、授权或额度，再手动发送。',
  upstream_rate_limited: '供应商暂时限流，请稍后手动重试。',
  upstream_connection_failed: '无法连接上游，请检查地址与网络。',
  upstream_redirect: '上游发生重定向，网关不跟随重定向，请修正地址。',
  upstream_timeout: '上游响应超时，请查看本次记录，不会自动重发。',
  stream_interrupted: '流式响应不完整，请查看本次记录，已有输出不会自动重放。',
  upstream_protocol_error: '上游响应协议不合法，请查看本次记录并检查供应商兼容性。',
  upstream_http_error: '上游服务失败，请查看本次记录并检查供应商状态。',
  storage_unavailable: '存储不可用，请停止重试并按项目 README 排查。请求记录可能不可用。',
  process_interrupted: '服务重启中断了调用，请查看记录并检查服务。',
  invalid_csrf: '管理安全校验失败，请退出后从本服务正式入口重新登录。',
  untrusted_origin: '请求来源不受信任，请从本服务正式入口重新登录。',
  invalid_request: '请求格式不合法，请修正输入后手动发送。',
  request_too_large: '消息过大，请缩短输入后手动发送。',
  unsupported_field: '请求包含不支持的字段，请修正后手动发送。',
  login_rate_limited: '登录尝试过多，请稍后再试。',
  session_limit: '管理会话数量已达上限，请稍后重新登录。',
  not_found: '目标不存在，请重新读取列表。',
  invalid_cursor: '分页信息无效，请刷新记录。'
};
function friendly(code) { return errors[code] || '请求未能完成，请检查服务后手动重试。'; }
function message(text, failed = false) { element(failed ? 'error' : 'notice').textContent = text; element(failed ? 'notice' : 'error').textContent = ''; }
function owner() { return {session: sessionEpoch, navigation: navigationVersion}; }
function current(token) { return token.session === sessionEpoch && token.navigation === navigationVersion && (token.load === undefined || token.load === loadSequence) && (token.find === undefined || token.find === findSequence); }
function stale() { const error = new Error('stale'); error.stale = true; return error; }
function assertCurrent(token) { if (!current(token)) throw stale(); }
async function api(path, method = 'GET', body, token = owner(), signal) {
  let response, value;
  try { response = await fetch('/admin/api/' + path, {method, signal, headers: {'Content-Type': 'application/json', 'X-CSRF-Token': csrf}, body: body === undefined ? undefined : JSON.stringify(body)}); }
  catch { assertCurrent(token); throw new Error('无法连接网关，请检查服务后手动重试。'); }
  assertCurrent(token);
  if (response.status === 401) {
    const login = path === 'login'; signedOut();
    message(login ? '管理员密码不正确，请重新输入。' : '管理会话已过期，请重新登录；不会自动重发调用。', true);
    throw stale();
  }
  try { value = await response.json(); }
  catch { assertCurrent(token); throw new Error('网关响应无法读取，请检查服务后重新读取。'); }
  assertCurrent(token);
  if (!response.ok) {
    const error = new Error(friendly(value.error?.code)); error.code = value.error?.code; throw error;
  }
  return value;
}
function passwordVisibility(visible) { element('password').type = visible ? 'text' : 'password'; element('password-toggle').textContent = visible ? '隐藏密码' : '显示密码'; element('password-visibility-status').textContent = visible ? '密码已显示' : '密码已隐藏'; }
element('password-toggle').onclick = () => passwordVisibility(element('password').type === 'password');
function clearVersion() {
  const request = versionRequest; versionRequest = null;
  if (request) { clearTimeout(request.timer); request.controller.abort(); }
  element('runtime-version').textContent = '';
}
async function loadVersion() {
  if (element('console').hidden) return;
  const request = {controller: new AbortController(), navigation: navigationVersion}; versionRequest = request;
  const isCurrent = () => versionRequest === request && request.navigation === navigationVersion && active === 'overview' && !element('console').hidden;
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
function clearSecrets() {
  issuedEpoch++;
  for (const id of ['password', 'upstream-token', 'issued-key', 'test-key', 'test-message']) element(id).value = '';
  passwordVisibility(false); element('issued').hidden = true;
  element('issued-key').type = 'password'; element('show-key').textContent = '显示新密钥';
  element('test-output').textContent = ''; element('test-raw').textContent = ''; element('diagnostics').open = false;
}
function stopCall(canceled = false) {
  const previous = call; call = null;
  if (previous) { previous.controller.abort(); if (previous.reader) previous.reader.cancel().catch(() => {}); }
  clearSecrets(); callInputs(false); element('cancel').disabled = true; element('send').disabled = !testReady || sessionUncertain;
  if (canceled) element('test-status').textContent = '已取消；本地取消不代表服务端已记为取消，请核对记录。';
}
function signedOut() {
  resetUpstream(); resetModelCreate();
  sessionEpoch++; navigationVersion++; clearTimeout(sessionTimer); clearTimeout(sessionExpiryTimer); clearVersion(); stopCall(); csrf = ''; expiry = 0; lastRequestID = ''; selectedModel = ''; overviewEvidence = null;
  element('find-request-id').value = ''; element('record-detail').textContent = ''; element('record-list').replaceChildren();
  element('test-request-id').textContent = ''; element('view-record').hidden = true;
  element('test-model').value = '';
  element('console').hidden = true; element('login').hidden = false; element('page-loading').hidden = true; element('password').focus();
  for (const button of document.querySelectorAll('button')) button.disabled = false;
}
function signedIn(value) {
  sessionEpoch++; sessionUncertain = false; sessionCheck = null; csrf = value.csrf; expiry = Date.parse(value.expires_at) || 0;
  element('password').value = ''; passwordVisibility(false); element('console').hidden = false; element('login').hidden = true; scheduleSession();
}
function expireSession() {
  if (!csrf || !expiry || Date.now() < expiry) return false;
  signedOut(); message('管理会话已过期，请重新登录；不会自动重发调用。', true); return true;
}
function scheduleExpiry() {
  clearTimeout(sessionExpiryTimer);
  if (!csrf || !expiry || expireSession()) return;
  const epoch = sessionEpoch, deadline = expiry;
  sessionExpiryTimer = setTimeout(() => {
    if (csrf && sessionEpoch === epoch && expiry === deadline && !expireSession()) scheduleExpiry();
  }, Math.max(0, deadline - Date.now()));
}
function scheduleSession() {
  scheduleExpiry();
  clearTimeout(sessionTimer); if (!csrf || sessionUncertain) return;
  const interval = active === 'test' && call && !document.hidden ? 30000 : 60000;
  sessionTimer = setTimeout(() => {
    if (document.hidden && active === 'test' && call && !(expiry && Date.now() >= expiry)) { scheduleSession(); return; }
    return checkSession();
  }, expiry ? Math.max(0, Math.min(interval, expiry - Date.now())) : interval);
}
async function checkSession() {
  if (!csrf || expireSession()) return;
  scheduleExpiry();
  const token = owner();
  if (sessionCheck && current(sessionCheck)) return;
  sessionCheck = token;
  try {
    const value = await api('session', 'GET', undefined, token);
    if (expireSession()) return;
    const refreshedExpiry = Date.parse(value.expires_at) || 0;
    if (refreshedExpiry) expiry = expiry ? Math.min(expiry, refreshedExpiry) : refreshedExpiry;
    if (expireSession()) return;
    const recovered = sessionUncertain;
    sessionUncertain = false; element('send').disabled = !!call || !testReady;
    if (recovered) message('管理会话已确认；需要重新输入调用内容并手动发送，不会自动重发。');
  }
  catch (error) {
    if (!error.stale && current(token)) {
      if (expireSession()) return;
      sessionUncertain = true; sessionEpoch++; resetUpstream(); resetModelCreate(); stopCall(); clearTimeout(sessionTimer);
      scheduleExpiry();
      element('test-status').textContent = '管理会话暂无法确认；调用已停止，敏感内容已清除。';
      message('管理会话暂无法确认，请检查服务连接后点击“重新检查会话”，或重新登录；网络异常不等于认证失效，不会自动重发。', true);
    }
  }
  finally { if (sessionCheck === token) { sessionCheck = null; if (csrf && current(token)) scheduleSession(); } }
}
element('check-session').onclick = () => checkSession();
element('relogin').onclick = () => { signedOut(); message('本地敏感内容已清除，请重新登录。'); };
async function guarded(button, callback) {
  if (button.disabled) return;
  const token = owner(); button.disabled = true;
  try { await callback(token); }
  catch (error) { if (!error.stale && current(token)) { message(error.message === 'Failed to fetch' ? '无法连接网关，请检查服务后手动重试。' : error.message, true); const fields = {invalid_destination: 'upstream-url', new_credential_required: 'upstream-token', invalid_model: 'model-alias', alias_conflict: 'model-alias', invalid_name: 'key-name'}; if (fields[error.code]) element(fields[error.code]).focus(); } }
  finally { if (current(token)) button.disabled = false; }
}
function action(text, callback) {
  const button = document.createElement('button'); button.type = 'button'; button.textContent = text; button.onclick = () => { const focused = document.activeElement === button ? button : null; return guarded(button, token => callback(token, focused)); }; return button;
}
function item(text, name, callback, identity) {
  const row = document.createElement('div'); row.className = 'item'; row.dataset.identity = identity; const span = document.createElement('span'); span.textContent = text; row.append(span, action(name, callback)); return row;
}
function restoreRowFocus(list, identity, previous, token) {
  if (!previous || !current(token) || (document.activeElement !== previous && document.activeElement !== document.body)) return;
  const row = [...element(list).children].find(value => value.dataset.identity === identity);
  if (row) row.querySelector('button')?.focus();
}
function callInputs(disabled) {
  for (const id of ['test-model', 'test-key', 'test-message', 'test-stream']) element(id).disabled = disabled;
}
function upstreamControls(disabled) {
  for (const id of ['upstream-name', 'upstream-url', 'upstream-token', 'upstream-enabled', 'save-upstream', 'refresh-upstream', 'new-upstream', 'cancel-upstream', 'check-upstream']) element(id).disabled = disabled;
  for (const button of element('upstream-list').querySelectorAll('button')) button.disabled = disabled;
  if (!disabled) element('check-upstream').disabled = !selectedUpstream?.enabled;
}
function resetUpstream() { upstreamOperation?.controller.abort(); upstreamOperation = null; upstreamControls(false); element('cancel-check').hidden = true; }
function modelCreateControls(disabled) { for (const id of ['model-alias', 'cancel-model-create', 'add-create-candidate']) element(id).disabled = disabled; for (const input of element('model-form').querySelectorAll('input,select,button')) input.disabled = disabled; }
function resetModelCreate() { modelCreateOperation = null; modelCreateControls(false); }
async function upstreamAction(token, callback) {
  if (upstreamOperation && current(upstreamOperation.token)) return;
  const operation = {token, controller: new AbortController()}; upstreamOperation = operation; upstreamControls(true);
  try { await callback(); }
  finally { if (upstreamOperation === operation && current(token)) { upstreamOperation = null; upstreamControls(false); element('cancel-check').hidden = true; } }
}
function showEmpty(id, text) { const paragraph = document.createElement('p'); paragraph.textContent = text; element(id).replaceChildren(paragraph); }
const decisionReasons = {selected: '选中', disabled_model: '模型停用', disabled_upstream: '上游停用', disabled_candidate: '候选停用或移出', credential_missing: '凭据缺失', cooling: '暂时冷却', already_attempted: '本请求已尝试', lower_priority: '优先级较低', not_selected_same_priority: '同级轮转未选中'};
async function allPages(path, token) {
  const values = []; let cursor = null;
  do { const page = await api(path + (cursor ? '?cursor=' + encodeURIComponent(cursor) : ''), 'GET', undefined, token); values.push(...page.data); cursor = page.next_cursor; } while (cursor);
  return values;
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
    const row = item(value.name + ' · ' + (value.enabled ? '启用' : '停用') + ' · 凭据' + (value.credential_configured ? '已配置' : '缺失'), '编辑上游 ' + value.name, own => upstreamAction(own, async () => { setUpstream(await api('upstreams/' + value.id, 'GET', undefined, own)); element('upstream-name').focus(); }), String(value.id));
    return row;
  }));
  if (!upstreamItems.length) showEmpty('upstream-list', '尚无上游。新建真实接入配置；不会创建演示供应商。');
  element('more-upstreams').hidden = !upstreamCursor;
}
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
}));
element('cancel-check').onclick = () => { if (!upstreamOperation) return; upstreamOperation.controller.abort(); element('cancel-check').hidden = true; element('check-state').textContent = '核验已取消；最终状态可重新读取，不会自动重试。'; };
element('more-upstreams').onclick = () => guarded(element('more-upstreams'), token => load('upstream', token, upstreamCursor));
element('more-models').onclick = () => guarded(element('more-models'), token => load('models', token, modelCursor));
for (const mode of ['create', 'edit']) element('add-' + mode + '-candidate').onclick = () => { if (editSaving) return; if (!upstreamItems.length) { message('请先接入一个真实上游。', true); return; } const values = readCandidates(mode); values.push(defaultCandidate()); renderCandidates(mode, values); };
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
async function load(view, token = owner(), pageCursor = '') {
  token = {...token, load: ++loadSequence};
  if (view === 'overview') {
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
  if (view === 'upstream') {
    const page = await api('upstreams' + (pageCursor ? '?cursor=' + encodeURIComponent(pageCursor) : ''), 'GET', undefined, token);
    upstreamItems = page.data; upstreamCursor = page.next_cursor; renderUpstreams(token);
    const selected = upstreamItems.find(value => value.id === selectedUpstream?.id);
    if (selected) setUpstream(selected);
    else if (selectedUpstream && pageCursor) setUpstream(await api('upstreams/' + selectedUpstream.id, 'GET', undefined, token));
    else setUpstream(upstreamItems[0] || null);
  }
  if (view === 'models') {
    const [page, sources] = await Promise.all([api('models' + (pageCursor ? '?cursor=' + encodeURIComponent(pageCursor) : ''), 'GET', undefined, token), allPages('upstreams', token)]);
    upstreamItems = sources; modelItems = page.data; modelCursor = page.next_cursor;
    const list = modelItems;
    element('model-list').replaceChildren(...list.map(value => {
      const row = item(value.alias + ' → ' + (value.candidates.length ? value.candidates.map(candidate => candidateName(candidate)).join('；') : '尚未关联上游' + (value.legacy_remote_model ? '（已保留原模型名 ' + value.legacy_remote_model + '）' : '')) + (value.enabled ? ' · 启用' : ' · 停用'), (value.enabled ? '停用模型 ' : '启用模型 ') + value.alias, async (own, focused) => {
        if (!confirm('确认变更模型状态？')) return;
        await api('models/' + encodeURIComponent(value.alias), 'PATCH', {expected_revision: value.revision, enabled: !value.enabled, candidates: cleanCandidates(value.candidates)}, own); await load('models', own); restoreRowFocus('model-list', value.alias, focused, own); message('模型状态已更新，需重新核验当前配置');
      }, value.alias);
      row.append(action('编辑映射 ' + value.alias, () => { if (editSaving) { message('映射正在保存，请等待结果。'); return; } editModel = value; element('edit-alias').value = value.alias; element('edit-routing-strategy').value = value.routing_strategy || 'weighted'; renderCandidates('edit', editDraft?.alias === value.alias ? editDraft.candidates : value.candidates); element('model-edit').hidden = false; element('edit-candidates').querySelector('input,select')?.focus(); })); return row;
    }));
    if (!list.length) showEmpty('model-list', '尚无模型。填写别名与供应商模型名称，创建首个映射；不会创建演示对象。');
    if (!element('create-candidates').children.length) renderCandidates('create', upstreamItems.length ? [defaultCandidate()] : []);
    else if (!modelCreateOperation) renderCandidates('create', readCandidates('create'));
    element('more-models').hidden = !modelCursor;
    element('models-dependency').textContent = sources.some(value => value.credential_configured && value.enabled) ? '上游已保存并启用，模型保存后可进入调用测试。' : '依赖尚缺：调用前请先配置并启用上游。可以保留模型草稿。';
  }
  if (view === 'keys') {
    const list = await api('keys', 'GET', undefined, token);
    element('key-list').replaceChildren(...list.map(value => item(value.name + ' · ' + value.hint + ' · 标识 ' + value.id + ' · 创建时间 ' + known(value.created) + (value.enabled ? ' · 启用' : ' · 停用'), (value.enabled ? '停用密钥 ' : '启用密钥 ') + value.name, async (own, focused) => {
      if (!confirm('确认变更密钥状态？')) return;
      await api('keys/' + encodeURIComponent(value.id), 'PATCH', {enabled: !value.enabled}, own); await load('keys', own); restoreRowFocus('key-list', value.id, focused, own); message('密钥状态已更新，需重新核验当前配置');
    }, value.id)));
    if (!list.length) showEmpty('key-list', '尚无应用密钥。填写用途名称后签发，完整秘密只展示一次。调用前还需上游与模型。');
  }
  if (view === 'test') {
    testReady = false; element('send').disabled = true;
    const [models, overview] = await Promise.all([allPages('models', token), api('overview', 'GET', undefined, token)]);
    assertCurrent(token);
    const enabled = models.filter(value => value.enabled && value.candidates.some(candidate => candidate.enabled));
    element('test-model').replaceChildren(...enabled.map(value => { const option = document.createElement('option'); option.value = value.alias; option.textContent = value.alias; return option; }));
    if (enabled.some(value => value.alias === selectedModel)) element('test-model').value = selectedModel;
    else if (selectedModel) { const placeholder = document.createElement('option'); placeholder.value = ''; placeholder.textContent = '原选模型已不可用，请重新选择'; element('test-model').append(placeholder); element('test-model').value = ''; }
    testReady = !!(overview.callable_models && enabled.length && overview.upstream.enabled && overview.upstream.credential_configured && overview.keys.enabled && !overview.degraded);
    element('send').disabled = !testReady || sessionUncertain;
    const missing = [];
    if (!overview.upstream.enabled || !overview.upstream.credential_configured) missing.push('配置并启用上游');
    if (!enabled.length) missing.push('创建或启用模型');
    if (!overview.keys.enabled) missing.push('签发或启用密钥');
    element('test-dependency').textContent = overview.degraded ? errors.storage_unavailable : missing.length ? '调用依赖尚缺：' + missing.join('；') : '依赖已具备，请输入应用密钥和消息；由实际请求核验。';
    element('test-status').textContent = sessionUncertain ? '管理会话暂无法确认，请重新检查会话或重新登录。' : '准备就绪';
    element('gateway-base').textContent = location.origin + '/v1';
    element('client-example').textContent = 'GATEWAY_BASE_URL=\'' + location.origin + '/v1\'\ncurl "$GATEWAY_BASE_URL/chat/completions" \\\n  -H "Authorization: Bearer <APPLICATION_KEY>" \\\n  -H "Content-Type: application/json" \\\n  -d \'{"model":"<MODEL_ALIAS>","messages":[{"role":"user","content":"你好"}]}\'';
    element('test-request-id').textContent = ''; element('view-record').hidden = true;
  }
  if (view === 'records') {
    const value = await api('requests' + (nextCursor ? '?cursor=' + encodeURIComponent(nextCursor) : ''), 'GET', undefined, token);
    element('record-list').replaceChildren(...value.data.map(recordRow)); nextCursor = value.next_cursor; element('more-records').hidden = !nextCursor;
    element('record-status').textContent = value.data.length ? '已读取真实记录；可用请求 ID 精确定位，不受分页影响。' : '尚无调用记录，请先发起调用；不会显示演示记录。';
  }
}
const states = {started: '网关已受理，尚未落盘终态', success: '成功', rejected: '网关拒绝', upstream_rejected: '供应商拒绝', upstream_failed: '上游失败', timeout: '超时', stream_interrupted: '流不完整', canceled: '取消', process_interrupted: '服务重启中断'};
const known = value => value === null || value === undefined || value === '' ? '未知' : String(value);
function recordValues(value) { return [value.id + '\n' + value.time, known(value.alias) + ' → ' + known(value.model) + '\n密钥标识：' + (value.key_id || '未识别应用密钥'), (states[value.state] || '未知') + ' (' + value.state + ')\n' + (value.error_code ? friendly(value.error_code) + ' (' + value.error_code + ')' : '') + '\n上游状态 ' + known(value.upstream_status), known(value.duration_ms), [value.prompt_tokens, value.completion_tokens, value.total_tokens].map(known).join(' / ')]; }
function recordRow(value) { const row = document.createElement('tr'); for (const text of recordValues(value)) { const cell = document.createElement('td'); cell.textContent = text; row.append(cell); } return row; }
async function findRecord(token) {
  token = {...token, find: ++findSequence};
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
function hashView() { const value = location.hash.slice(1); return views.includes(value) ? value : 'overview'; }
async function navigate(view, updateHash = true, requestID = '') {
  if (!views.includes(view)) view = 'overview';
  selectedModel = element('test-model').value || selectedModel;
  navigationVersion++; resetUpstream(); resetModelCreate(); clearVersion(); stopCall(); active = view; nextCursor = null; editModel = null; editDraft = null; editSaving = false; element('model-edit').hidden = true;
  if (!views.includes(location.hash.slice(1))) history.replaceState(null, '', '#' + view);
  else if (updateHash && location.hash !== '#' + view) history.pushState(null, '', '#' + view);
  for (const panel of document.querySelectorAll('.panel')) panel.hidden = true;
  for (const button of document.querySelectorAll('[data-view]')) { if (button.dataset.view === view) button.setAttribute('aria-current', 'page'); else button.removeAttribute('aria-current'); }
  for (const button of document.querySelectorAll('button')) button.disabled = false;
  element('cancel').disabled = true; element('send').disabled = true;
  message(''); if (!csrf) return;
  const token = owner(); element('page-loading').hidden = false; scheduleSession();
  if (view === 'overview') loadVersion();
  if (view === 'records') { element('find-request-id').value = requestID; element('record-detail').textContent = ''; }
  try { await load(view, token); if (view === 'records' && requestID) await findRecord(token); }
  catch (error) { if (!error.stale && current(token)) { message(error.message, true); if (view === 'overview') element('summary').textContent = '状态未知，读取失败，请重新读取。'; if (view === 'records') element('record-status').textContent = '读取失败；已展示的记录并非最新，请手动刷新。'; } }
  finally { if (current(token) && csrf) { element('page-loading').hidden = true; element(view).hidden = false; const heading = element(view).querySelector('h1'); if (heading) { heading.tabIndex = -1; heading.focus(); } } }
}
for (const button of document.querySelectorAll('[data-view]')) button.onclick = () => navigate(button.dataset.view);
function form(id, callback) { element(id).onsubmit = event => { event.preventDefault(); return guarded(event.submitter || element(id).querySelector('button'), callback); }; }
form('login-form', async token => { const value = await api('login', 'POST', {password: element('password').value}, token); signedIn(value); await navigate(hashView()); });
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
}));
element('refresh-upstream').onclick = () => guarded(element('refresh-upstream'), token => upstreamAction(token, async () => {
  element('upstream-token').value = '';
  try { await load('upstream', token); message('已重新读取上游状态；保存不代表调用成功。已撤销本页未保存修改。'); }
  catch (error) { if (!error.stale && current(token)) { element('credential-state').textContent = '恢复未确认，保留公开草稿；请重新读取服务器状态。'; error.message = '恢复未确认；输入凭据已清除，公开草稿保留。' + error.message; } throw error; }
}));
form('model-form', async token => { if (modelCreateOperation && current(modelCreateOperation.token)) return; const operation = {token}; modelCreateOperation = operation; const input = {alias: element('model-alias').value, routing_strategy: element('create-routing-strategy').value, candidates: readCandidates('create'), enabled: true}; modelCreateControls(true); let saved = false; try { await api('models', 'POST', input, token); saved = true; element('model-create').open = false; await load('models', token); message('模型已保存；保存不代表生成成功'); } catch (error) { if (!error.stale && current(token) && saved) error.message = '模型已保存，但最新列表读取未确认；请重新读取，不会重复创建。' + error.message; throw error; } finally { if (modelCreateOperation === operation && current(token)) resetModelCreate(); } });
element('cancel-model-create').onclick = () => { if (modelCreateOperation) return; element('model-create').open = false; element('model-create').querySelector('summary')?.focus(); };
element('cancel-key-create').onclick = () => { element('key-create').open = false; element('key-create').querySelector('summary')?.focus(); };
form('model-edit-form', async token => {
  if (!editModel || editSaving) return;
  editSaving = true; element('cancel-edit').disabled = true;
  const editing = editModel;
  for (const input of element('model-edit-form').querySelectorAll('input,select,button')) input.disabled = true;
  try { await api('models/' + encodeURIComponent(editing.alias), 'PATCH', {expected_revision: editing.revision, enabled: editing.enabled, routing_strategy: element('edit-routing-strategy').value, candidates: readCandidates('edit')}, token); element('model-edit').hidden = true; editModel = null; editDraft = null; await load('models', token); message('映射已保存，需重新核验'); }
  finally { if (current(token)) { editSaving = false; element('cancel-edit').disabled = false; for (const input of element('model-edit-form').querySelectorAll('input,select,button')) input.disabled = false; } }
});
element('cancel-edit').onclick = () => { if (editSaving) return; const alias = editModel?.alias; if (alias) editDraft = {alias, candidates: readCandidates('edit')}; editModel = null; element('model-edit').hidden = true; const button = [...element('model-list').querySelectorAll('button')].find(value => value.textContent === '编辑映射 ' + alias); if (button) button.focus(); };
form('key-form', async token => {
  closeKey(); let value; element('cancel-key-create').disabled = true;
  try { value = await api('keys', 'POST', {name: element('key-name').value}, token); }
  catch (error) { if (error.stale || error.code) throw error; throw new Error('签发结果未确认，请刷新密钥列表后再决定是否重新签发，并停用遗失值。不会自动重签。'); }
  finally { if (current(token)) element('cancel-key-create').disabled = false; }
  element('key-create').open = false; element('issued-key').value = value.key; element('issued').hidden = false; element('issued-key').focus(); message('密钥已签发，请立即保管；关闭后不再返回。'); await load('keys', token);
});
function closeKey() { issuedEpoch++; element('issued-key').value = ''; element('issued-key').type = 'password'; element('show-key').textContent = '显示新密钥'; element('issued').hidden = true; }
element('show-key').onclick = () => { const visible = element('issued-key').type === 'password'; element('issued-key').type = visible ? 'text' : 'password'; element('show-key').textContent = visible ? '隐藏新密钥' : '显示新密钥'; };
element('close-key').onclick = () => { closeKey(); element('key-create').querySelector('summary')?.focus(); };
element('copy-key').onclick = () => guarded(element('copy-key'), async token => { const epoch = issuedEpoch; try { await navigator.clipboard.writeText(element('issued-key').value); assertCurrent(token); if (epoch === issuedEpoch && !element('issued').hidden) message('已复制，请保管'); } catch (error) { assertCurrent(token); if (epoch === issuedEpoch && !element('issued').hidden) message('自动复制失败，请在当前展示中手动选择复制。', true); } });
element('copy-example').onclick = () => guarded(element('copy-example'), async token => { try { await navigator.clipboard.writeText(element('client-example').textContent); assertCurrent(token); message('占位示例已复制，请在客户端替换占位符。'); } catch { assertCurrent(token); message('自动复制失败，请手动选择调用示例。', true); } });
element('logout').onclick = async () => {
  const savedCSRF = csrf; signedOut(); const epoch = sessionEpoch; message('已退出；正在撤销服务端会话。');
  try { const response = await fetch('/admin/api/logout', {method: 'POST', keepalive: true, headers: {'Content-Type': 'application/json', 'X-CSRF-Token': savedCSRF}, body: '{}'}); if (!response.ok && response.status !== 401) throw new Error('logout'); if (sessionEpoch === epoch) message('已退出，服务端会话已撤销。'); }
  catch { if (sessionEpoch === epoch) message('本地已清除；无法确认服务端撤销，请检查服务。', true); }
};
element('refresh-overview').onclick = () => navigate('overview');
element('refresh-records').onclick = () => { nextCursor = null; return guarded(element('refresh-records'), token => load('records', token)); };
element('more-records').onclick = () => guarded(element('more-records'), token => load('records', token));
form('find-record-form', findRecord);
element('view-record').onclick = () => navigate('records', true, lastRequestID);
element('overview-record').onclick = () => { if (overviewEvidence) return navigate('records', true, overviewEvidence.request_id); };
element('cancel').onclick = () => stopCall(true);
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
element('test-form').onsubmit = event => {
  event.preventDefault(); if (call || !testReady || sessionUncertain) return;
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
      if (owns()) { call = null; callInputs(false); element('send').disabled = !testReady || sessionUncertain; element('cancel').disabled = true; scheduleSession(); }
    }
  })();
};
window.addEventListener('hashchange', () => navigate(hashView(), false));
window.addEventListener('pagehide', () => { navigationVersion++; resetUpstream(); resetModelCreate(); clearVersion(); stopCall(); clearTimeout(sessionTimer); clearTimeout(sessionExpiryTimer); });
window.addEventListener('pageshow', () => { scheduleExpiry(); if (csrf && !sessionUncertain) checkSession(); });
document.addEventListener('visibilitychange', () => {
  if (document.hidden || !csrf || expireSession()) return;
  scheduleExpiry(); if (!sessionUncertain) checkSession();
});
(async () => { try { const value = await api('session'); signedIn(value); await navigate(hashView()); } catch (error) { if (!error.stale) { signedOut(); message('无法确认管理会话，请检查服务或重新登录。', true); } } })();
