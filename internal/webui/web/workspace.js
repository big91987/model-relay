import {createTransport} from './shared/transport.mjs';
import {createPlayground} from './features/playground/index.mjs';
import {collection as readCollection,list} from './shared/data.mjs';
import {createFactSummary} from './shared/facts.mjs';
import {createModels} from './features/models/index.mjs';
import {createServices} from './features/services/index.mjs';
import {createKeys} from './features/keys/index.mjs';
import {createLimits} from './features/limits/index.mjs';
import {createBilling} from './features/billing/index.mjs';
import {createTenants} from './features/tenants/index.mjs';
import {createReports} from './features/reports/index.mjs';
import {createOverview} from './features/overview/index.mjs';
const byId = id => document.getElementById(id);
const pages = {
  overview: ['工作台', '真实调用、已确认费用与待确认风险。'],
  models: ['模型托管', '公共模型、路由候选、租户授权与能力边界。'],
  services: ['供应商与服务', '接入已有端点，核验当前配置；不部署GPU算力。'],
  playground: ['调用测试', '仅使用应用Key，普通与SSE调用均保留同一Request ID。'],
  keys: ['API Keys', '稳定应用身份、一次性秘密、诊断成员与共享额度。'],
  tenants: ['租户与成员', '真实邀请、人员角色、成员退出与租户暂停。'],
  limits: ['配额与预算', '次数、RPM、TPM、并发与费用独立生效；空值表示无限，0表示拒绝。'],
  usage: ['用量分析', '确认金额与未知用量分开，筛选和导出遵循相同权限。'],
  billing: ['定价与账单', 'USD测试计价；成本和租户费用独立，历史快照不回算。'],
  records: ['调用日志', '同Request ID追踪多尝试、价格快照、拒绝和费用核对。']
};
const state = {user:null,csrf:'',tenant:'',tenants:[],page:'overview',epoch:0,busy:false,invitation:''};
let playground=null,features=null,featureScope='';
const node = (tag,text='',className='') => { const element=document.createElement(tag); element.textContent=text; if(className) element.className=className; return element; };
const button = (label,action) => { const element=node('button',label); element.type='button'; element.onclick=event=>runAction(action,event); return element; };
async function runAction(action,event) { const epoch=state.epoch;try {await action(event);}catch(error){if(epoch===state.epoch&&error.name!=='AbortError')notice('操作失败：'+error.message);} }
const collection=path=>readCollection(api,path);
function notice(text) { byId('notice').textContent=text; }
let modalScroll=null;
function releaseModalScroll() {
  if(!modalScroll||document.querySelector('dialog:modal'))return;
  const position=modalScroll;modalScroll=null;
  document.body.classList.remove('modal-scroll-locked');
  document.body.style.removeProperty('--modal-scroll-top');
  document.body.style.removeProperty('--modal-scroll-left');
  window.scrollTo(position.left,position.top);
}
function closeModal(dialog) {
  dialog.close();
  releaseModalScroll();
}
function openModal(dialog) {
  if(!modalScroll){
    modalScroll={left:window.scrollX,top:window.scrollY};
    document.body.style.setProperty('--modal-scroll-top',`${-modalScroll.top}px`);
    document.body.style.setProperty('--modal-scroll-left',`${-modalScroll.left}px`);
    document.body.classList.add('modal-scroll-locked');
  }
  try{dialog.showModal();}catch(error){releaseModalScroll();throw error;}
}
function clearSensitive() { for(const element of document.querySelectorAll('input[type=password],.secret,textarea[data-sensitive]')) element.value=''; byId('detail-content').replaceChildren(); if(byId('call-output')) byId('call-output').textContent=''; }
function invalidate() { state.epoch++; if(byId('editor-form').dataset.readback==='pending')byId('editor-form').dataset.readback='interrupted'; transport.cancel(); if(features)for(const feature of new Set(Object.values(features)))feature.dispose();playground?.dispose(); playground=null; state.busy=false; if(byId('editor').open) byId('editor').close(); if(byId('detail').open) byId('detail').close(); releaseModalScroll(); clearSensitive(); }
function signedOut() { invalidate(); state.user=null; state.csrf=''; state.tenant=''; state.tenants=[]; features=null;featureScope=''; byId('content').replaceChildren(); byId('tenant').replaceChildren(); byId('identity').textContent=''; byId('workspace').hidden=true; byId('login-panel').hidden=false; }
async function syncTenants() {
  const epoch=state.epoch,tenants=await collection('tenants');if(epoch!==state.epoch)return;state.tenants=tenants;const select=byId('tenant');select.replaceChildren();
  if(state.user.platform){const option=node('option','全部租户');option.value='';select.append(option);}
  for(const tenant of tenants){const option=node('option',tenant.name+(tenant.active?'':' · 已暂停'));option.value=tenant.id;select.append(option);}
  if(state.tenant&&!tenants.some(tenant=>tenant.id===state.tenant))state.tenant='';
  if(!state.user.platform&&!state.tenant)state.tenant=tenants[0]?.id??'';select.value=state.tenant;
}
const transport=createTransport({
  session:()=>({epoch:state.epoch,user:state.user?.id,csrf:state.csrf}),
  onError(status,path){if(status===401&&path!=='login'&&path!=='invitations/accept')signedOut();if(status===403)clearSensitive();}
});
const api=transport.request;
function table(columns,values,filtered=false) {
  if(!filtered&&values.length){const wrap=node('div'),search=node('input'),body=node('div');search.type='search';search.placeholder='搜索当前列表';search.setAttribute('aria-label','搜索当前列表');search.oninput=()=>{const query=search.value.trim().toLocaleLowerCase();body.replaceChildren(table(columns,values.filter(value=>JSON.stringify(value).toLocaleLowerCase().includes(query)),true));};wrap.append(search,body);body.append(table(columns,values,true));return wrap;}
  if(values.length>25){const wrap=node('div'),body=node('div'),controls=node('div','','toolbar');let page=1;const previous=button('上一组',()=>{page--;render();}),next=button('下一组',()=>{page++;render();}),count=node('span');const render=()=>{body.replaceChildren(table(columns,values.slice((page-1)*25,page*25),true));previous.disabled=page===1;next.disabled=page*25>=values.length;count.textContent=`第${page}组 · 共${values.length}项`;};controls.append(previous,count,next);wrap.append(body,controls);render();return wrap;}
  if(!values.length) return node('div','当前范围暂无数据','empty');
  const wrap=node('div','','table-wrap'), tableElement=node('table'), head=node('thead'), row=node('tr');
  wrap.tabIndex=0;wrap.setAttribute('role','region');wrap.setAttribute('aria-label','数据表，可横向滚动查看完整列');
  for(const column of columns) row.append(node('th',column.label)); head.append(row); tableElement.append(head);
  const body=node('tbody');
  for(const value of values) { const rowElement=node('tr'); for(const column of columns) { const cell=node('td'), output=column.render(value); if(output instanceof Node) cell.append(output); else cell.textContent=output==null?'未知':String(output); rowElement.append(cell); } body.append(rowElement); }
  tableElement.append(body); wrap.append(tableElement); return wrap;
}
const {column,factSummary}=createFactSummary({node});
const actions = handlers => { const group=node('div','','actions'); for(const [label,handler] of handlers) group.append(button(label,handler)); return group; };
function detail(title,value,extra=[]) { const dialog=byId('detail'); byId('detail-title').textContent=title; const content=byId('detail-content'); content.replaceChildren(value instanceof Node?value:factSummary(value),...extra); openModal(dialog); byId('detail-close').focus(); }
byId('detail-close').onclick=()=>closeModal(byId('detail'));
byId('detail').addEventListener('close',()=>{releaseModalScroll();byId('detail-content').replaceChildren();});
function closeEditor() { if(state.busy) return; closeModal(byId('editor')); }
byId('editor-close').onclick=closeEditor; byId('editor-cancel').onclick=closeEditor;
byId('editor').addEventListener('cancel',event=>{if(state.busy) event.preventDefault();});
byId('editor').addEventListener('close',()=>{ if(byId('editor').open)return;releaseModalScroll();for(const element of byId('editor-fields').querySelectorAll('input[type=password],.secret,textarea[data-sensitive]')) element.value=''; byId('editor-fields').replaceChildren(); byId('editor-error').replaceChildren(); byId('editor-form').onsubmit=null; });
function edit(title,fields,submit) {
  if(state.busy)return;
  const dialog=byId('editor'), container=byId('editor-fields'); container.replaceChildren(); byId('editor-title').textContent=title; byId('editor-error').textContent=''; byId('editor-save').hidden=false; byId('editor-save').disabled=false;
  for(const field of fields) {
    const group=node('div','','field'), id='field-'+field.name, label=node('label',field.label); label.htmlFor=id;
    let input;
    if(field.options) { input=node('select'); input.multiple=!!field.multiple; for(const optionValue of field.options) { const option=node('option',optionValue.label); option.value=optionValue.value; option.selected=field.multiple?(field.value??[]).includes(optionValue.value):field.value===optionValue.value; input.append(option); } }
    else if(field.type==='textarea') { input=node('textarea'); input.value=field.value??''; }
    else { input=node('input'); input.type=field.type??'text'; if(input.type==='checkbox') input.checked=field.value??false; else input.value=field.value??''; }
    input.id=id; input.name=field.name; input.required=!!field.required; if(field.min!=null) input.min=field.min; if(field.max!=null) input.max=field.max;
    if(field.sensitive) { input.classList.add('secret'); input.autocomplete='off'; }
    group.append(label,input); if(field.help) group.append(node('small',field.help)); container.append(group);
  }
  const operation=crypto.randomUUID(), epoch=state.epoch;
  const form=byId('editor-form');form.dataset.operation=operation;form.dataset.readback='idle';
  let unknown=false;
  byId('editor-form').onsubmit=async event=>{
    event.preventDefault(); if(state.busy||unknown) return; state.busy=true; byId('editor-save').disabled=true; byId('editor-error').textContent='';
    const values={}; for(const field of fields) { const input=byId('field-'+field.name); values[field.name]=field.multiple?Array.from(input.selectedOptions,option=>option.value):field.type==='checkbox'?input.checked:input.value; }
    try {
      const result=await submit(values,operation); if(epoch!==state.epoch) return;
      closeModal(dialog);
      if(result?.key){state.busy=false;secretDialog(result);}
      else if(result?.acceptance_token){state.busy=false;invitationDialog(result);}
      else {
        form.dataset.readback='pending';notice('写入已确认；正在读取当前事实。');
        byId('content').replaceChildren(node('p','正在读取当前范围…'));byId('page-action').hidden=true;
        try {
          await syncTenants();if(epoch!==state.epoch)return;
          const committed=await renderPage();if(epoch!==state.epoch)return;
          form.dataset.readback=committed?'complete':'failed';
          notice(committed?'写入已确认；当前范围已更新。':'写入已确认，但读取失败；请重试读取，不要重复写入。');
        } catch(error) {
          if(epoch!==state.epoch)return;
          form.dataset.readback='failed';notice('写入已确认，但读取失败；请重试读取，不要重复写入。');
          byId('content').replaceChildren(node('div','读取失败：'+error.message+'。旧数据未作为当前事实展示。','error'),button('重试读取',async()=>{await syncTenants();if(epoch===state.epoch)await renderPage();}));
        }
      }
    }
    catch(error) {
      if(epoch!==state.epoch) return;
      state.busy=false; clearSensitive();
      const output=byId('editor-error');
      output.textContent=error.status?`操作失败：${error.message}。非敏感草稿仍保留。`:'写入结果未知，请先安全读回；不要重复签发或计费。';
      if(!error.status) {
        unknown=true; byId('editor-save').hidden=true;
        const reference=node('p','操作标识：'+operation), status=node('p','');
        const retry=button('安全读回',async()=>{
          if(retry.disabled||epoch!==state.epoch||!dialog.open) return;
          retry.disabled=true;
          try {
            const result=await api('operations/'+encodeURIComponent(operation));
            if(epoch===state.epoch&&dialog.open) detail('操作已提交，秘密不可恢复',result);
          } catch {
            if(epoch===state.epoch&&dialog.open) status.textContent='读回失败，结果仍未知。请保留操作标识并稍后重试读回。';
          } finally { retry.disabled=false; }
        });
        output.append(reference,status,retry);
      }
    }
    finally { if(epoch===state.epoch){state.busy=false;byId('editor-save').disabled=false;} }
  };
  openModal(dialog); container.querySelector('input,select,textarea')?.focus();
}
function secretDialog(result) {
  const dialog=byId('editor'); byId('editor-title').textContent='一次性 API Key'; byId('editor-fields').replaceChildren(node('p','只显示本次。关闭、切租户、退出或刷新后不可恢复；轮换保留应用额度与历史。'));
  const input=node('input'); input.className='secret'; input.type='password'; input.readOnly=true; input.value=result.key; input.setAttribute('aria-label','一次性API Key'); byId('editor-fields').append(input);
  byId('editor-fields').append(button('显示 / 隐藏',()=>{input.type=input.type==='password'?'text':'password';}),button('复制',async()=>{try{await navigator.clipboard.writeText(input.value); byId('editor-error').textContent='已复制';}catch{input.focus(); input.select(); byId('editor-error').textContent='无法自动复制，请手动复制。';}}),button('显式用于调用测试',async()=>{const secret=input.value; closeModal(dialog); const navigation=navigate('playground'),epoch=state.epoch; await navigation; if(epoch===state.epoch&&state.page==='playground'&&byId('call-key')) byId('call-key').value=secret;}));
  byId('editor-save').hidden=true; byId('editor-form').onsubmit=event=>event.preventDefault(); openModal(dialog); input.focus();
}
function invitationDialog(result) { const link=new URL('/admin/accept',location.origin); link.hash=result.acceptance_token; secretDialog({key:link.toString()}); byId('editor-title').textContent='邀请接受链接（手工交付）'; byId('editor-fields').querySelector('p').textContent='未发送邮件；请通过受信渠道交给受邀者。链接只显示一次，默认24小时有效。'; const buttons=byId('editor-fields').querySelectorAll('button'); buttons[buttons.length-1].remove(); }
function mainAction(label,handler) { const element=byId('page-action'); element.hidden=false; element.textContent=label; element.onclick=event=>runAction(handler,event); }

function tenantDestination(tenant,page) {state.tenant=tenant.id;byId('tenant').value=tenant.id;features?.records.setFilter({});return navigate(page);}
async function modelCall(alias) {const navigation=navigate('playground'),epoch=state.epoch;await navigation;if(epoch===state.epoch&&byId('call-model'))byId('call-model').value=alias;}

function role() { return state.user?.platform?'administrator':state.tenants.find(tenant=>tenant.id===state.tenant)?.role; }

async function exportData(path) {
  const epoch=state.epoch;
  try {const blob=await transport.download(path);if(epoch!==state.epoch)return;const address=URL.createObjectURL(blob),anchor=node('a','下载CSV');anchor.href=address;anchor.download='relay-export.csv';anchor.click();setTimeout(()=>URL.revokeObjectURL(address),1000);}
  catch(error){if(epoch===state.epoch&&error.name!=='AbortError')notice(error.message);}
}

function featurePages() {
  const identity=state.user.id+':'+state.tenant;
  if(features&&featureScope===identity)return features;
  if(features)for(const feature of new Set(Object.values(features)))feature.dispose();
  featureScope=identity;
  const scope={platform:state.user.platform,tenant:state.tenant,get tenants(){return state.tenants;},get role(){return role();}};
  const ui={node,button,table,column,actions,detail,edit,mainAction,notice,factSummary,closeDetail:()=>closeModal(byId('detail'))};
  const options={request:api,scope,ui,navigate,openPlayground:modelCall,selectTenant:tenantDestination,download:exportData};
  const reports=createReports(options),models=createModels(options),limits=createLimits(options);
  features={overview:createOverview(options),models,services:createServices({...options,openModel:models.editModel}),tenants:createTenants(options),keys:createKeys({...options,showResources:limits.showResources,showRecords:filters=>{reports.setFilter(filters);return navigate('records');}}),limits,billing:createBilling(options),usage:reports,records:reports};
  return features;
}
async function renderPage() {
  const epoch=state.epoch, page=state.page, content=byId('content'); content.replaceChildren(node('p','正在读取当前范围…')); byId('page-action').hidden=true; byId('page-title').textContent=pages[page][0];byId('page-description').textContent=pages[page][1];byId('breadcrumb').textContent=pages[page][0];
  try {
    const modules=featurePages();
    let section;
    if(page==='playground'){playground?.dispose();playground=createPlayground({document,notify:notice,openRecord:modules.records.openRecord});section=playground.element;}
    else section=await modules[page].render(page);
    if(epoch!==state.epoch)return false;content.replaceChildren(section);return true;
  } catch(error) { if(epoch!==state.epoch||error.name==='AbortError')return false;content.replaceChildren(node('div','读取失败：'+error.message+'。旧数据未作为当前事实展示。','error'),button('重试读取',()=>renderPage()));return false; }
}
async function navigate(page,preserve=false) {invalidate();state.page=pages[page]?page:'overview';if(!preserve&&page!=='records')features?.records.resetPage();notice('');for(const element of byId('navigation').children)element.setAttribute('aria-current',element.dataset.page===state.page?'page':'false');byId('sidebar').classList.remove('open');byId('nav-toggle').setAttribute('aria-expanded','false');await renderPage();}
for(const [page,[title]] of Object.entries(pages)){const element=button(title,()=>navigate(page));element.dataset.page=page;byId('navigation').append(element);}
async function signedIn() { const session=await api('session');state.csrf=session.csrf;state.user=session.user;await syncTenants();byId('identity').textContent=state.user.account+' · '+(state.user.platform?'平台管理员':'服务端成员身份');byId('login-panel').hidden=true;byId('invite-panel').hidden=true;byId('workspace').hidden=false;clearSensitive();await navigate('overview');const epoch=state.epoch;fetch('/healthz').then(response=>response.json()).then(value=>{if(epoch===state.epoch&&state.user)byId('version').textContent='运行版本 '+value.version+' · schema '+value.storage_schema;}).catch(()=>{}); }
byId('login-form').onsubmit=async event=>{event.preventDefault();const submit=byId('login-form').querySelector('button');if(submit.disabled)return;submit.disabled=true;try{const value=await api('login','POST',{account:byId('account').value,password:byId('password').value});state.csrf=value.csrf;await signedIn();if(state.invitation){byId('workspace').hidden=true;byId('invite-panel').hidden=false;byId('invitation-password').required=false;byId('invitation-password').value='';byId('invitation-password').hidden=true;byId('invite-panel').querySelector('label').hidden=true;}}catch(error){byId('login-error').textContent='登录失败：'+error.message;}finally{byId('password').value='';submit.disabled=false;}};
byId('tenant').onchange=async()=>{state.tenant=byId('tenant').value;await navigate(state.page);};
byId('refresh').onclick=async()=>{try{await syncTenants();await navigate(state.page,true);}catch(error){notice('刷新失败：'+error.message);}};
byId('logout').onclick=async()=>{const token=state.csrf;signedOut();try{await fetch('/admin/api/logout',{method:'POST',credentials:'same-origin',headers:{'Content-Type':'application/json','X-CSRF-Token':token},body:'{}'});}catch{byId('login-error').textContent='本地已退出；服务端注销未确认，请关闭浏览器并等待会话到期。';}};
byId('nav-toggle').onclick=()=>{const open=byId('sidebar').classList.toggle('open');byId('nav-toggle').setAttribute('aria-expanded',String(open));};
byId('invite-login').onclick=()=>{byId('invite-panel').hidden=true;byId('login-panel').hidden=false;};
byId('accept-form').onsubmit=async event=>{event.preventDefault();const submit=byId('accept-form').querySelector('button');if(submit.disabled)return;submit.disabled=true;try{const result=await api('invitations/accept','POST',{token:state.invitation,password:byId('invitation-password').value});state.invitation='';byId('invite-error').textContent='已加入租户；请用 '+result.account+' 登录。';byId('invite-panel').hidden=true;if(state.user)await signedIn();else byId('login-panel').hidden=false;}catch(error){byId('invite-error').textContent='接受失败：'+error.message;}finally{byId('invitation-password').value='';submit.disabled=false;}};
const fragment=location.hash.slice(1);if(location.pathname==='/admin/accept'&&fragment){state.invitation=fragment;history.replaceState(null,'',location.pathname);byId('login-panel').hidden=true;byId('invite-panel').hidden=false;}else{signedIn().catch(()=>signedOut());}
window.addEventListener('pagehide',()=>{invalidate();state.invitation='';});
