import {errors,friendly} from './shared/legacy-presentation.mjs';
import {createLegacyOverview} from './features/legacy-overview/index.mjs';
import {createLegacyUpstreams} from './features/legacy-upstreams/index.mjs';
import {createLegacyModels} from './features/legacy-models/index.mjs';
import {createLegacyKeys} from './features/legacy-keys/index.mjs';
import {createLegacyPlayground} from './features/legacy-playground/index.mjs';
import {createLegacyRecords} from './features/legacy-records/index.mjs';
'use strict';
const element = id => document.getElementById(id);
const views = ['overview', 'upstream', 'models', 'keys', 'test', 'records'];
let csrf='',active='overview',navigationVersion=0,sessionEpoch=0,loadSequence=0;
let sessionTimer=null,expiry=0,sessionUncertain=false,sessionCheck=null,sessionExpiryTimer=null;
function message(text, failed = false) { element(failed ? 'error' : 'notice').textContent = text; element(failed ? 'notice' : 'error').textContent = ''; }
function owner() { return {session: sessionEpoch, navigation: navigationVersion}; }
function current(token) { return token.session === sessionEpoch && token.navigation === navigationVersion && (token.load === undefined || token.load === loadSequence) && (!token.valid || token.valid()); }
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
const options={document,element,api,owner,current,assertCurrent,stale,guarded,action,item,restoreRowFocus,showEmpty,form,load,navigate,message,scheduleSession,expireSession,isSessionUncertain:()=>sessionUncertain,isOverviewActive:version=>version===navigationVersion&&active==='overview'};
const features={overview:createLegacyOverview(options),upstream:createLegacyUpstreams(options),models:createLegacyModels(options),keys:createLegacyKeys(options),test:createLegacyPlayground(options),records:createLegacyRecords(options)};
function disposeFeatures(){for(const feature of Object.values(features))feature.dispose();element('password').value='';passwordVisibility(false);}
function stopCall(canceled=false){features.test.stop(canceled);features.keys.clearSecret();features.upstream.clearSecret();element('password').value='';passwordVisibility(false);}
function signedOut() {
  disposeFeatures();
  sessionEpoch++; navigationVersion++; clearTimeout(sessionTimer); clearTimeout(sessionExpiryTimer); features.overview.clearVersion(); stopCall(); csrf = ''; expiry = 0;features.overview.clearEvidence();features.test.reset();features.records.reset();
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
  const interval = active === 'test' && features.test.running && !document.hidden ? 30000 : 60000;
  sessionTimer = setTimeout(() => {
    if (document.hidden && active === 'test' && features.test.running && !(expiry && Date.now() >= expiry)) { scheduleSession(); return; }
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
    sessionUncertain = false; element('send').disabled = features.test.running || !features.test.ready;
    if (recovered) message('管理会话已确认；需要重新输入调用内容并手动发送，不会自动重发。');
  }
  catch (error) {
    if (!error.stale && current(token)) {
      if (expireSession()) return;
      sessionUncertain = true; sessionEpoch++; disposeFeatures(); stopCall(); clearTimeout(sessionTimer);
      scheduleExpiry();
      element('test-status').textContent = '管理会话暂无法确认；调用已停止，敏感内容已清除。';
      message('管理会话暂无法确认，请检查服务连接后点击“重新检查会话”，或重新登录；网络异常不等于认证失效，不会自动重发。', true);
    }
  }
  finally { if (sessionCheck === token) { sessionCheck = null; if (csrf && current(token)) scheduleSession(); } }
}
element('check-session').onclick = () => checkSession();
element('relogin').onclick = () => { signedOut(); message('本地敏感内容已清除，请重新登录。'); };
async function guarded(button, callback, onError) {
  if (button.disabled) return;
  const token = owner(); button.disabled = true;
  try { await callback(token); }
  catch (error) { if (!error.stale && current(token)) { message(error.message === 'Failed to fetch' ? '无法连接网关，请检查服务后手动重试。' : error.message, true); onError?.(error); } }
  finally { if (current(token)) button.disabled = false; }
}
function action(text, callback, onError) {
  const button = document.createElement('button'); button.type = 'button'; button.textContent = text; button.onclick = () => { const focused = document.activeElement === button ? button : null; return guarded(button, token => callback(token, focused), onError); }; return button;
}
function item(text, name, callback, identity, onError) {
  const row = document.createElement('div'); row.className = 'item'; row.dataset.identity = identity; const span = document.createElement('span'); span.textContent = text; row.append(span, action(name, callback, onError)); return row;
}
function restoreRowFocus(list, identity, previous, token) {
  if (!previous || !current(token) || (document.activeElement !== previous && document.activeElement !== document.body)) return;
  const row = [...element(list).children].find(value => value.dataset.identity === identity);
  if (row) row.querySelector('button')?.focus();
}
function showEmpty(id, text) { const paragraph = document.createElement('p'); paragraph.textContent = text; element(id).replaceChildren(paragraph); }
async function load(view,token=owner(),pageCursor=''){token={...token,load:++loadSequence};return features[view].render(token,pageCursor);}
function hashView() { const value = location.hash.slice(1); return views.includes(value) ? value : 'overview'; }
async function navigate(view, updateHash = true, requestID = '') {
  if (!views.includes(view)) view = 'overview';
  navigationVersion++;disposeFeatures();active=view;
  if (!views.includes(location.hash.slice(1))) history.replaceState(null, '', '#' + view);
  else if (updateHash && location.hash !== '#' + view) history.pushState(null, '', '#' + view);
  for (const panel of document.querySelectorAll('.panel')) panel.hidden = true;
  for (const button of document.querySelectorAll('[data-view]')) { if (button.dataset.view === view) button.setAttribute('aria-current', 'page'); else button.removeAttribute('aria-current'); }
  for (const button of document.querySelectorAll('button')) button.disabled = false;
  element('cancel').disabled = true; element('send').disabled = true;
  message(''); if (!csrf) return;
  const token = owner(); element('page-loading').hidden = false; scheduleSession();
  if (view === 'overview') features.overview.loadVersion();
  if (view === 'records') { element('find-request-id').value = requestID; element('record-detail').textContent = ''; }
  try { await load(view, token); if (view === 'records' && requestID) await features.records.find(token); }
  catch (error) { if (!error.stale && current(token)) { message(error.message, true); if (view === 'overview') element('summary').textContent = '状态未知，读取失败，请重新读取。'; if (view === 'records') element('record-status').textContent = '读取失败；已展示的记录并非最新，请手动刷新。'; } }
  finally { if (current(token) && csrf) { element('page-loading').hidden = true; element(view).hidden = false; const heading = element(view).querySelector('h1'); if (heading) { heading.tabIndex = -1; heading.focus(); } } }
}
for (const button of document.querySelectorAll('[data-view]')) button.onclick = () => navigate(button.dataset.view);
function form(id, callback, onError) { element(id).onsubmit = event => { event.preventDefault(); return guarded(event.submitter || element(id).querySelector('button'), callback, onError); }; }
form('login-form', async token => { const value = await api('login', 'POST', {password: element('password').value}, token); signedIn(value); await navigate(hashView()); });
element('logout').onclick = async () => {
  const savedCSRF = csrf; signedOut(); const epoch = sessionEpoch; message('已退出；正在撤销服务端会话。');
  try { const response = await fetch('/admin/api/logout', {method: 'POST', keepalive: true, headers: {'Content-Type': 'application/json', 'X-CSRF-Token': savedCSRF}, body: '{}'}); if (!response.ok && response.status !== 401) throw new Error('logout'); if (sessionEpoch === epoch) message('已退出，服务端会话已撤销。'); }
  catch { if (sessionEpoch === epoch) message('本地已清除；无法确认服务端撤销，请检查服务。', true); }
};
window.addEventListener('hashchange', () => navigate(hashView(), false));
window.addEventListener('pagehide', () => { navigationVersion++; disposeFeatures(); clearTimeout(sessionTimer); clearTimeout(sessionExpiryTimer); });
window.addEventListener('pageshow', () => { scheduleExpiry(); if (csrf && !sessionUncertain) checkSession(); });
document.addEventListener('visibilitychange', () => {
  if (document.hidden || !csrf || expireSession()) return;
  scheduleExpiry(); if (!sessionUncertain) checkSession();
});
(async () => { try { const value = await api('session'); signedIn(value); await navigate(hashView()); } catch (error) { if (!error.stale) { signedOut(); message('无法确认管理会话，请检查服务或重新登录。', true); } } })();
