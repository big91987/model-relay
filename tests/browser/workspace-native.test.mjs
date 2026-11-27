import {readFileSync} from 'node:fs';
import assert from 'node:assert/strict';
import test from 'node:test';
class Element {
  constructor(tag='div',text=''){this.tag=tag;this.textContent=text;this.children=[];this.dataset={};this.value='';this.open=false;this.classes=new Set();this.classList={add:name=>this.classes.add(name),remove:name=>this.classes.delete(name),toggle:name=>{if(this.classes.has(name)){this.classes.delete(name);return false;}this.classes.add(name);return true;}};this.style={setProperty(){},removeProperty(){}};}
  append(...values){this.children.push(...values);}
  replaceChildren(...values){this.children=values;}
  setAttribute(name,value){this[name]=value;}
  removeAttribute(name){delete this[name];}
  querySelectorAll(){return [];}
  querySelector(){return null;}
  addEventListener(){}
  focus(){}
  showModal(){this.open=true;}
  close(){this.open=false;}
}
test('native workspace modules compose every management page and isolate a tenant change during a report read',async context=>{
  const elements=new Map(),byId=id=>{if(!elements.has(id))elements.set(id,new Element());return elements.get(id);};
  const document={getElementById:byId,createElement:tag=>new Element(tag),querySelectorAll:()=>[],querySelector:()=>null,body:new Element('body')};
  const paths=[],requests=[],events=new Map();let delayed=false,release;
  const values={
    session:{csrf:'controlled',user:{id:'admin',platform:true,account:'admin@example.test'}},
    tenants:[{id:'team',name:'Team',active:true},{id:'other',name:'Other',active:true}],
    'workspace/overview':{requests_today:0,confirmed_fee_micro_usd:'0',reserved_micro_usd:'0',pending_attempts:0,legacy_pending_requests:0},
    'workspace/requests':{items:[],total:0,page_size:25},
    'tenants/team/keys':[{id:'key',name:'App',application:'App',revision:1,models:[]}],
    'tenants/team/keys/key/limits':{period:'2026-10',limits:{calls:10,hard:false},usage:{calls:2,rpm:0,tpm:0,concurrency:0,fee_micro_usd:'7',reserved_micro_usd:'3',pending:1}},
    'tenants/team/limits':{limits:{hard:false},usage:{calls:0,rpm:0,tpm:0,concurrency:0,fee_micro_usd:'0',reserved_micro_usd:'0',pending:0}},
    'tenants/other/limits':{limits:{hard:false},usage:{calls:0,rpm:0,tpm:0,concurrency:0,fee_micro_usd:'0',reserved_micro_usd:'0',pending:0}}
  };
  const originals=new Map();
  for(const [name,value] of Object.entries({document,window:{scrollX:0,scrollY:0,scrollTo(){},addEventListener:(name,callback)=>events.set(name,callback)},location:{pathname:'/admin',hash:'',origin:'http://controlled.test'},Node:Element,history:{replaceState(){}},fetch:async(path,options)=>{
    paths.push(path);requests.push({path,options});const url=new URL(path,'http://controlled.test'),resource=url.pathname.replace('/admin/api/','');
    const response={ok:true,status:200,json:async()=>values[resource]??[]};
    if(delayed&&resource==='workspace/requests'&&url.searchParams.get('tenant')==='team'){delayed=false;return new Promise(resolve=>{release=()=>resolve(response);});}
    return response;
  }})){originals.set(name,Object.getOwnPropertyDescriptor(globalThis,name));Object.defineProperty(globalThis,name,{configurable:true,writable:true,value});}
  context.after(()=>{for(const [name,descriptor] of originals)if(descriptor)Object.defineProperty(globalThis,name,descriptor);else delete globalThis[name];});
  const workspaceURL=new URL('../../internal/webui/web/workspace.js',import.meta.url);
  const source=readFileSync(workspaceURL,'utf8').replace(/from '(\.[^']+)'/g,(_match,path)=>'from '+JSON.stringify(new URL(path,workspaceURL).href));
  await import('data:text/javascript;base64,'+Buffer.from(source).toString('base64'));
  const settle=()=>new Promise(resolve=>setImmediate(resolve));await settle();
  const navigate=async page=>{const button=byId('navigation').children.find(value=>value.dataset.page===page);assert.ok(button,page);await button.onclick({});};
  for(const page of ['services','models','tenants','keys','limits','usage','billing','records','playground','overview']){
    await navigate(page);assert.equal(byId('content').children.length,1,page);assert.notEqual(byId('content').children[0].className,'error',page);
  }
  byId('tenant').value='team';await byId('tenant').onchange();
  await navigate('keys');
  const descendants=element=>[element,...element.children.flatMap(descendants)];
  const quota=descendants(byId('content')).find(element=>element.tag==='button'&&element.textContent==='额度 / 用量');
  assert.ok(quota);await quota.onclick({});
  assert.equal(byId('detail').open,true);assert.equal(byId('detail-content').children[0].tag,'div');
  const labels=descendants(byId('detail-content')).map(element=>element.textContent);
  for(const label of ['独立维度','本周期次数','每分钟请求 RPM','每分钟收费Token TPM（含预留）','本地活跃并发','费用预算（USD）','0.000010 USD'])assert.ok(labels.includes(label),label);
  delayed=true;const old=navigate('records');await settle();
  const previous=requests.findLast(request=>request.path.includes('workspace/requests')&&request.path.includes('tenant=team'));
  byId('tenant').value='other';await byId('tenant').onchange();
  assert.equal(previous.options.signal.aborted,true);release();await old;
  assert.ok(paths.some(path=>path.includes('workspace/requests')&&path.includes('tenant=other')));
  assert.equal(byId('tenant').value,'other');assert.equal(byId('content').children[0].tag,'section');
  events.get('pagehide')();
});
