import assert from 'node:assert/strict';
import test from 'node:test';
import {createModels} from '../../internal/webui/web/features/models/index.mjs';
import {createReports} from '../../internal/webui/web/features/reports/index.mjs';
const deferred=()=>{let resolve,reject;const promise=new Promise((done,fail)=>{resolve=done;reject=fail;});return {promise,resolve,reject};};
const node=(tag,text='')=>({tag,textContent:text,children:[],append(...values){this.children.push(...values);},replaceChildren(...values){this.children=values;},setAttribute(){}});
const ui={node,button:(text,onclick)=>({...node('button',text),onclick}),table:(columns,values)=>({...node('table'),columns,values}),column:(label,field)=>({label,field}),actions:values=>values,mainAction(){},notice(){},detail(){},edit(){}};
const scope={tenant:'team',platform:true,role:'administrator',tenants:[]};
test('models feature discards an asynchronous editor dependency after disposal',async()=>{
  let resolve,edits=0;
  const feature=createModels({request:()=>new Promise(done=>{resolve=done;}),scope,ui:{...ui,edit(){edits++;}},openPlayground(){}});
  const editing=feature.editModel();feature.dispose();resolve([]);
  await assert.rejects(editing,{name:'AbortError'});assert.equal(edits,0);
});
test('reports owns filters and pagination separately from another tenant instance',async()=>{
  const paths=[];
  const request=async path=>{paths.push(path);return {items:[],total:0,page_size:25};};
  const first=createReports({request,scope,ui,navigate(){},download(){}});
  first.setFilter({key:'app'});await first.render('records');first.dispose();
  const second=createReports({request,scope:{...scope,tenant:'other'},ui,navigate(){},download(){}});
  await second.render('records');
  assert.match(paths[0],/key=app/);assert.match(paths[0],/tenant=team/);
  assert.doesNotMatch(paths[1],/key=app/);assert.match(paths[1],/tenant=other/);
});

import {createServices} from '../../internal/webui/web/features/services/index.mjs';
import {createKeys} from '../../internal/webui/web/features/keys/index.mjs';
import {createLimits} from '../../internal/webui/web/features/limits/index.mjs';
import {createBilling} from '../../internal/webui/web/features/billing/index.mjs';
import {createTenants} from '../../internal/webui/web/features/tenants/index.mjs';
import {createOverview} from '../../internal/webui/web/features/overview/index.mjs';
function controls() {
  const buttons=new Map(),actions=new Map(),edits=[],details=[];
  return {buttons,actions,edits,details,ui:{...ui,
    button(label,onclick){buttons.set(label,onclick);return {...node('button',label),onclick};},
    table(columns,values){for(const value of values)for(const column of columns)if(column.label==='操作')column.render(value);return node('table');},
    actions(values){for(const [label,handler] of values)actions.set(label,handler);return node('actions');},
    mainAction(label,handler){buttons.set(label,handler);},
    edit(title,fields,submit){edits.push({title,fields,submit});},
    detail(title,value,extra){details.push({title,value,extra});},
    factSummary:()=>node('facts'),closeDetail(){}
  }};
}
const resource={limits:{revision:4},usage:{fee_micro_usd:'0',reserved_micro_usd:'0'}};
test('limit editor keeps null, zero and each independent resource dimension in its write',async()=>{
  const control=controls(),writes=[];
  const feature=createLimits({scope,ui:control.ui,request:async(path,method,body)=>{if(method){writes.push({path,method,body});return {};}return path.includes('/limits')?resource:[];}});
  await feature.render();control.buttons.get('调整租户限额')();
  await control.edits[0].submit({calls:'0',rpm:'3',tpm:'7',concurrency:'2',budget_micro_usd:'',hard:true});
  assert.deepEqual(writes,[{path:'tenants/team/limits',method:'PUT',body:{calls:0,rpm:3,tpm:7,concurrency:2,budget_micro_usd:null,hard:true,revision:4}}]);
});
test('key rotation preserves application revision, transition hours and the editor operation',async()=>{
  const control=controls(),writes=[];
  const feature=createKeys({scope,ui:control.ui,request:async(path,method,body,operation)=>{if(method){writes.push({path,method,body,operation});return {key:'one-time'};}return path.includes('/limits')?resource:[{id:'key',name:'App',revision:9}];}});
  const element=await feature.render();control.actions.get('轮换')();
  assert.equal(control.edits[0].title,'轮换（共享额度与历史）');
  const issued=await control.edits[0].submit({transition_hours:'24'},'operation');
  assert.equal(issued.key,'one-time');assert.deepEqual(writes,[{path:'tenants/team/keys/key/rotations',method:'POST',body:{transition_hours:24,expected_revision:9},operation:'operation'}]);
  feature.dispose();assert.deepEqual(element.children,[]);
});
test('billing publishes independent candidate costs with the same acknowledged operation',async()=>{
  const control=controls(),writes=[];
  const model={alias:'m/a',candidates:[{upstream_id:1},{upstream_id:2}]};
  const feature=createBilling({scope:{...scope,tenant:''},ui:control.ui,request:async(path,method,body,operation)=>{if(method){writes.push({path,method,body,operation});return {};}return [model];}});
  await feature.render();await control.actions.get('发布新价')();
  await control.edits[0].submit({costs:'1,10,20\n2,30,40',tenant_input_rate:'50',tenant_output_rate:'60',effective:'utc'},'price-operation');
  assert.deepEqual(writes,[{path:'models/m%2Fa/prices',method:'POST',operation:'price-operation',body:{effective_at:'utc',prices:[{upstream_id:1,cost_input_rate:10,cost_output_rate:20,tenant_input_rate:50,tenant_output_rate:60},{upstream_id:2,cost_input_rate:30,cost_output_rate:40,tenant_input_rate:50,tenant_output_rate:60}]}}]);
});
test('tenant invitations preserve the selected role and explicit lifetime',async()=>{
  const control=controls(),writes=[];
  const feature=createTenants({scope,ui:control.ui,request:async(path,method,body)=>{if(method){writes.push({path,method,body});return {acceptance_token:'once'};}return [];}});
  await feature.render();control.buttons.get('邀请成员')();
  const invited=await control.edits[0].submit({account:'a@example.test',role:'finance',hours:'1'});
  assert.equal(invited.acceptance_token,'once');assert.deepEqual(writes,[{path:'tenants/team/invitations',method:'POST',body:{account:'a@example.test',role:'finance',hours:1}}]);
});
test('disposing a supplier check suppresses both late details and action errors',async()=>{
  const control=controls(),pending=deferred(),notices=[];
  const feature=createServices({scope,ui:{...control.ui,notice:text=>notices.push(text)},request:async(path,method)=>method?pending.promise:path.startsWith('upstreams')?[{id:1,execution_revision:3}]:[]});
  await feature.render();const check=control.actions.get('核验')();feature.dispose();pending.resolve({state:'success'});await check;
  assert.deepEqual(control.details,[]);assert.deepEqual(notices,[]);
});
test('overview cannot commit an old catalog after its feature is disposed',async()=>{
  const control=controls(),pending=deferred();
  const feature=createOverview({scope,ui:control.ui,navigate(){},request:async path=>path.startsWith('catalog')?pending.promise:{requests_today:1,pending_attempts:0,legacy_pending_requests:0,reserved_micro_usd:'0',confirmed_fee_micro_usd:'0'}});
  const reading=feature.render();await Promise.resolve();await Promise.resolve();feature.dispose();pending.resolve([]);
  await assert.rejects(reading,{name:'AbortError'});
});


test('feature disposal also isolates a pending supplier network rejection',async()=>{
  const control=controls(),pending=deferred(),notices=[];
  const feature=createServices({scope,ui:{...control.ui,notice:text=>notices.push(text)},request:async(path,method)=>method?pending.promise:path.startsWith('upstreams')?[{id:1,execution_revision:3}]:[]});
  await feature.render();const check=control.actions.get('核验')();feature.dispose();pending.reject(new Error('old network failure'));await check;
  assert.deepEqual(control.details,[]);assert.deepEqual(notices,[]);
});

test('limits owns resource tables without a shell-provided quota renderer',async()=>{
  const control=controls(),tables=[];
  const report={period:'2026-10',limits:{calls:10,rpm:0,tpm:20,concurrency:null,budget_micro_usd:'30',hard:false},usage:{calls:2,rpm:0,tpm:7,concurrency:1,fee_micro_usd:'7',reserved_micro_usd:'3',pending:1},soft_budget_exceeded:true};
  const genericUI=control.ui;
  const feature=createLimits({scope,request:async path=>path.includes('/limits')?report:[{id:'key',name:'App'}],ui:{...genericUI,table(columns,values){tables.push({columns,values});return node('resource-table');},factSummary:()=>node('facts')}});
  const rendered=await feature.render();
  assert.equal(tables.length,2);
  for(const {values} of tables){assert.equal(values.length,5);assert.deepEqual(values.map(value=>value.name),['本周期次数','每分钟请求 RPM','每分钟收费Token TPM（含预留）','本地活跃并发','费用预算（USD）']);assert.equal(values[0].remaining,'8');assert.equal(values[1].limit,0);assert.equal(values[3].limit,'无限');assert.equal(values[4].used,'0.000010 USD');assert.equal(values[4].remaining,'0.000020 USD');}
  const texts=element=>[element.textContent,...element.children.flatMap(texts)];
  assert.equal(texts(rendered).filter(value=>value==='软预算已超额，请核对费用或调整预算。').length,2);
});

test('key quota details use the limits-owned presentation through composition',async()=>{
  const control=controls(),tables=[];
  const report={period:'2026-10',limits:{calls:10,hard:false},usage:{calls:2,rpm:0,tpm:0,concurrency:0,fee_micro_usd:'7',reserved_micro_usd:'3',pending:1}};
  const genericUI={...control.ui,table(columns,values){tables.push({columns,values});return node('resource-table');},factSummary:()=>node('facts')};
  const limits=createLimits({scope,request:async()=>[],ui:genericUI});
  const feature=createKeys({scope,ui:control.ui,request:async path=>path.includes('/limits')?report:[{id:'key',name:'App',revision:1}],showResources:(title,value)=>limits.showResources(title,value)});
  await feature.render();await control.actions.get('额度 / 用量')();
  assert.equal(control.details.length,1);assert.equal(control.details[0].title,'应用独立资源与使用量');
  assert.equal(control.details[0].value.tag,'div');assert.equal(tables.length,1);assert.equal(tables[0].values.length,5);assert.equal(tables[0].values[4].used,'0.000010 USD');
});
