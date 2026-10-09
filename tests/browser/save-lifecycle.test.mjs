import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';

const source=readFileSync('internal/relay/web/workspace.js','utf8');
const editor=source.slice(source.indexOf('function edit('),source.indexOf('function secretDialog('));
const deferred=()=>{let resolve,reject;const promise=new Promise((success,failure)=>{resolve=success;reject=failure;});return {promise,resolve,reject};};
function fixture() {
  const element=()=>({dataset:{},children:[],textContent:'',disabled:false,hidden:false,open:false,replaceChildren(...values){this.children=values;},querySelector(){return null;},showModal(){this.open=true;},close(){this.open=false;}});
  const elements=new Map(['editor','editor-fields','editor-title','editor-error','editor-save','editor-form','content','page-action'].map(id=>[id,element()]));
  const state={epoch:1,busy:false},tenants=deferred(),catalog=deferred(),tenantStarted=deferred(),renderStarted=deferred(),counts={writes:0,renders:0,secrets:0};let sequence=0,notice='';
  const context=vm.createContext({state,crypto:{randomUUID:()=>`controlled-${++sequence}`},byId:id=>elements.get(id),clearSensitive(){},openModal:dialog=>dialog.showModal(),closeModal:dialog=>dialog.close(),notice:text=>{notice=text;},node:(tag,text)=>({...element(),textContent:text}),button:(text,onclick)=>({...element(),textContent:text,onclick}),syncTenants:()=>{tenantStarted.resolve();return tenants.promise;},renderPage:()=>{counts.renders++;renderStarted.resolve();return catalog.promise;},secretDialog:()=>{counts.secrets++;},invitationDialog:()=>{counts.secrets++;},submit:async()=>{counts.writes++;return {};}});
  vm.runInContext(editor+"\nedit('controlled',[],submit);",context);
  return {elements,state,tenants,catalog,tenantStarted,renderStarted,counts,context,notice:()=>notice,submit:()=>elements.get('editor-form').onsubmit({preventDefault(){}})};
}
test('current editor retains its lock through tenants and DOM commit, rejects duplicate submit',async()=>{
  const value=fixture(),form=value.elements.get('editor-form');let completed=false;
  const saving=value.submit().then(()=>{completed=true;});await value.tenantStarted.promise;
  assert.equal(form.dataset.readback,'pending');assert.equal(value.state.busy,true);assert.equal(completed,false);
  assert.equal(value.elements.get('content').children[0].textContent,'正在读取当前范围…');
  vm.runInContext("edit('blocked replacement',[],submit);",value.context);
  assert.equal(form.dataset.operation,'controlled-1');assert.equal(form.dataset.readback,'pending');
  await value.submit();assert.equal(value.counts.writes,1);
  value.tenants.resolve();await value.renderStarted.promise;assert.equal(value.counts.renders,1);assert.equal(completed,false);
  value.catalog.resolve(true);await saving;
  assert.equal(form.dataset.readback,'complete');assert.equal(value.state.busy,false);assert.equal(value.elements.get('editor-save').disabled,false);
});
for(const phase of ['tenants','catalog'])test(`acknowledged ${phase} read failure stays failed, never unknown write or old success`,async()=>{
  const value=fixture(),saving=value.submit();await value.tenantStarted.promise;
  if(phase==='tenants')value.tenants.reject(new Error('controlled read failed'));
  else{value.tenants.resolve();await value.renderStarted.promise;value.catalog.resolve(false);}
  await saving;
  assert.equal(value.elements.get('editor-form').dataset.readback,'failed');
  assert.match(value.notice(),/写入已确认，但读取失败/);assert.equal(value.counts.writes,1);
  assert.equal(value.state.busy,false);assert.equal(value.elements.get('editor-error').textContent,'');
});
test('old tenant read cannot start a new-scope renderer or release the new edit lock',async()=>{
  const value=fixture(),saving=value.submit();await value.tenantStarted.promise;value.state.epoch++;value.state.busy=false;
  vm.runInContext("edit('new',[],submit);",value.context);
  value.elements.get('editor-save').disabled=true;value.state.busy=true;
  value.tenants.resolve();await saving;
  assert.equal(value.counts.renders,0);assert.equal(value.elements.get('editor-form').dataset.operation,'controlled-2');
  assert.equal(value.elements.get('editor-form').dataset.readback,'idle');assert.equal(value.elements.get('editor-save').disabled,true);assert.equal(value.state.busy,true);
});
test('old catalog settlement cannot mark a new editor complete or overwrite its lock',async()=>{
  const value=fixture(),saving=value.submit();await value.tenantStarted.promise;value.tenants.resolve();await value.renderStarted.promise;
  value.state.epoch++;value.state.busy=false;vm.runInContext("edit('new',[],submit);",value.context);
  value.elements.get('editor-save').disabled=true;value.state.busy=true;
  value.catalog.resolve(true);await saving;
  assert.equal(value.elements.get('editor-form').dataset.readback,'idle');assert.equal(value.elements.get('editor-save').disabled,true);
});
test('one-time secrets retain their separate editor lifecycle without page-read completion',async()=>{
  const value=fixture();vm.runInContext("edit('secret',[],async()=>({key:'controlled-not-supplier'}));",value.context);
  await value.submit();assert.equal(value.counts.secrets,1);assert.equal(value.counts.renders,0);assert.equal(value.state.busy,false);
});

test('actual syncTenants discards a collection settled after scope invalidation',async()=>{
  const collection=deferred(),state={epoch:1,tenants:[{id:'current'}],tenant:'current'},select={replaceChildren(){throw new Error('stale DOM mutation');}};
  const context=vm.createContext({state,collection:()=>collection.promise,byId:()=>select});
  vm.runInContext(source.slice(source.indexOf('async function syncTenants('),source.indexOf('async function api(')),context);
  const reading=vm.runInContext('syncTenants()',context);state.epoch++;
  collection.resolve([{id:'old'}]);await reading;
  assert.deepEqual(state.tenants,[{id:'current'}]);assert.equal(state.tenant,'current');
});

test('actual renderPage returns failed without old rows after catalog read failure',async()=>{
  const node=(tag,text='',className='')=>({tag,textContent:text,className,children:[],append(...values){this.children.push(...values);},replaceChildren(...values){this.children=values;}});
  const elements=new Map(['content','page-action','page-title','page-description','breadcrumb'].map(id=>[id,node('div')]));
  elements.get('content').children=[node('table','old model')];
  const context=vm.createContext({state:{epoch:1,page:'models',user:{platform:true}},pages:{models:['模型托管','scope']},byId:id=>elements.get(id),node,scoped:path=>path,collection:async()=>{throw new Error('controlled read failed');},button:(text,onclick)=>({...node('button',text),onclick})});
  vm.runInContext(source.slice(source.indexOf('async function renderPage('),source.indexOf('function callForm(')),context);
  assert.equal(await vm.runInContext('renderPage()',context),false);
  assert.equal(elements.get('content').children[0].className,'error');
  assert.match(elements.get('content').children[0].textContent,/controlled read failed/);
  assert.equal(elements.get('content').children.some(child=>child.tag==='table'),false);
});
