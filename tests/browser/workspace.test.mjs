import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';

const source = readFileSync(new URL('../../internal/relay/web/workspace.js', import.meta.url), 'utf8');

test('workspace clips background scrolling only while a native modal is open',()=>{
  const style=readFileSync(new URL('../../internal/relay/web/workspace.css',import.meta.url),'utf8');
  assert.match(style,/html:has\(dialog:modal\)\{overflow:clip\}/);
  assert.match(style,/body\.modal-scroll-locked\{position:fixed;top:var\(--modal-scroll-top\);left:var\(--modal-scroll-left\);width:100%\}/);
  assert.match(style,/dialog\{[^}]*overflow:auto/);
  assert.doesNotMatch(style,/html\s*\{[^}]*overflow:(?:hidden|clip)/);
});

test('modal scroll lock preserves one background position across queued close and reopening',()=>{
  let modal=false;
  const properties=new Map(),classes=new Set(),restored=[];
  const document={body:{style:{setProperty:(name,value)=>properties.set(name,value),removeProperty:name=>properties.delete(name)},classList:{add:name=>classes.add(name),remove:name=>classes.delete(name)}},querySelector:()=>modal?{}:null};
  const window={scrollX:12,scrollY:182,scrollTo:(left,top)=>restored.push({left,top})};
  const context=vm.createContext({document,window});
  vm.runInContext(source.slice(source.indexOf('let modalScroll='),source.indexOf('function clearSensitive(')),context);
  const dialog={showModal(){modal=true;}};
  context.openModal(dialog);
  assert.equal(classes.has('modal-scroll-locked'),true);
  assert.equal(properties.get('--modal-scroll-top'),'-182px');
  assert.equal(properties.get('--modal-scroll-left'),'-12px');
  modal=false;window.scrollY=0;
  context.openModal(dialog);
  context.releaseModalScroll();
  assert.equal(restored.length,0);
  assert.equal(properties.get('--modal-scroll-top'),'-182px');
  modal=false;context.releaseModalScroll();
  assert.deepEqual(restored,[{left:12,top:182}]);
  assert.equal(classes.size,0);assert.equal(properties.size,0);
  context.releaseModalScroll();assert.equal(restored.length,1);
});

test('explicit modal close restores background before a queued close event is delivered',()=>{
  let modal=false;
  const properties=new Map(),classes=new Set(),restored=[];
  const document={body:{style:{setProperty:(name,value)=>properties.set(name,value),removeProperty:name=>properties.delete(name)},classList:{add:name=>classes.add(name),remove:name=>classes.delete(name)}},querySelector:()=>modal?{}:null};
  const window={scrollX:12,scrollY:182,scrollTo:(left,top)=>restored.push({left,top})};
  const context=vm.createContext({document,window});
  vm.runInContext(source.slice(source.indexOf('let modalScroll='),source.indexOf('function clearSensitive(')),context);
  const dialog={showModal(){modal=true;},close(){modal=false;}};
  context.openModal(dialog);
  context.closeModal(dialog);
  assert.equal(modal,false);
  assert.equal(classes.size,0);
  assert.equal(properties.size,0);
  assert.deepEqual(restored,[{left:12,top:182}]);
  context.openModal(dialog);
  context.releaseModalScroll();
  assert.equal(classes.has('modal-scroll-locked'),true);
  assert.equal(restored.length,1);
  context.closeModal(dialog);
  assert.equal(classes.size,0);
  assert.equal(restored.length,2);
  context.releaseModalScroll();
  assert.equal(restored.length,2);
});

test('closing one modal preserves the lock while another native modal remains',()=>{
  let modalCount=0,locked=false,restored=0;
  const context=vm.createContext({document:{body:{style:{setProperty(){},removeProperty(){}},classList:{add(){locked=true;},remove(){locked=false;}}},querySelector:()=>modalCount?{}:null},window:{scrollX:0,scrollY:182,scrollTo(){restored++;}}});
  vm.runInContext(source.slice(source.indexOf('let modalScroll='),source.indexOf('function clearSensitive(')),context);
  const first={showModal(){modalCount++;},close(){modalCount--;}};
  const second={showModal(){modalCount++;},close(){modalCount--;}};
  context.openModal(first);context.openModal(second);
  context.closeModal(first);
  assert.equal(locked,true);assert.equal(restored,0);
  context.closeModal(second);
  assert.equal(locked,false);assert.equal(restored,1);
});

test('explicit UI close paths use the shared synchronous release without replacing native focus',()=>{
  assert.match(source,/byId\('detail-close'\)\.onclick=\(\)=>closeModal\(byId\('detail'\)\)/);
  assert.match(source,/function closeEditor\(\) \{ if\(state\.busy\) return; closeModal\(byId\('editor'\)\); \}/);
  const closeSource=source.slice(source.indexOf('function closeModal('),source.indexOf('function openModal('));
  assert.match(closeSource,/dialog\.close\(\);\s+releaseModalScroll\(\);/);
  assert.doesNotMatch(closeSource,/focus\(|setTimeout|requestAnimationFrame/);
});

test('failed native modal opening releases the scroll lock without swallowing the error',()=>{
  let locked=false,restored=false;
  const context=vm.createContext({document:{body:{style:{setProperty(){},removeProperty(){}},classList:{add(){locked=true;},remove(){locked=false;}}},querySelector:()=>null},window:{scrollX:0,scrollY:327,scrollTo(left,top){restored=left===0&&top===327;}}});
  vm.runInContext(source.slice(source.indexOf('let modalScroll='),source.indexOf('function clearSensitive(')),context);
  const failure=new Error('controlled native opening failure');
  assert.throws(()=>context.openModal({showModal(){throw failure;}}),error=>error===failure);
  assert.equal(locked,false);assert.equal(restored,true);
});

test('workspace invalidation closes both native dialogs before synchronously releasing background scroll',()=>{
  const calls=[];
  const dialogs=new Map(['editor','detail'].map(name=>[name,{open:true,close(){this.open=false;calls.push(name);}}]));
  dialogs.set('editor-form',{dataset:{}});
  const context=vm.createContext({state:{epoch:0,controllers:new Set(),call:null,busy:true},byId:name=>dialogs.get(name),releaseModalScroll(){assert.equal(dialogs.get('editor').open,false);assert.equal(dialogs.get('detail').open,false);calls.push('release');},clearSensitive(){calls.push('clear');}});
  vm.runInContext(source.slice(source.indexOf('function invalidate('),source.indexOf('function signedOut(')),context);
  context.invalidate();assert.deepEqual(calls,['editor','detail','release','clear']);
});

test('a new editor owns its save lock after navigation interrupts acknowledged-write readback', async () => {
  const elements = new Map();
  const control = () => ({textContent:'',value:'',hidden:false,disabled:false,dataset:{},append(){},replaceChildren(){},querySelector(){return null;},focus(){}});
  for (const id of ['editor','editor-fields','editor-title','editor-error','editor-save','editor-form','content','page-action']) elements.set(id, control());
  const dialog = elements.get('editor');
  dialog.showModal = () => { dialog.open = true; };
  dialog.close = () => { dialog.open = false; };
  let releaseReadback;
  const state = {epoch:1,busy:false};
  const context = vm.createContext({
    state,crypto:{randomUUID:()=> 'controlled-operation'},byId:id=>elements.get(id),
    notice(){},clearSensitive(){},openModal:dialog=>dialog.showModal(),closeModal:dialog=>dialog.close(),node:()=>control(),syncTenants:()=>new Promise(resolve=>{releaseReadback=resolve;}),renderPage:async()=>true
  });
  vm.runInContext(source.slice(source.indexOf('function edit('),source.indexOf('function secretDialog(')),context);
  vm.runInContext("edit('first',[],async()=>({}));",context);
  const firstSubmit = elements.get('editor-form').onsubmit({preventDefault(){}});
  await Promise.resolve();
  assert.equal(elements.get('editor-save').disabled,true);
  assert.equal(dialog.open,false);
  state.epoch++;
  state.busy=false;
  vm.runInContext("edit('second',[],async()=>({}));",context);
  assert.equal(elements.get('editor-save').disabled,false);
  const secondSubmit = elements.get('editor-form').onsubmit({preventDefault(){}});
  assert.equal(elements.get('editor-save').disabled,true);
  releaseReadback();
  await firstSubmit;
  assert.equal(elements.get('editor-save').disabled,true);
  await Promise.resolve();
  releaseReadback();
  await secondSubmit;
  assert.equal(elements.get('editor-save').disabled,false);
});

test('save driver observes one acknowledged write and rejects a locked editor before arming a waiter', async () => {
  const driver = readFileSync(new URL('./console.mjs', import.meta.url), 'utf8');
  const handler = driver.slice(driver.indexOf('async function save('),driver.indexOf('async function navigate('));
  for (const enabled of [true,false]) {
    let waiters=0,clicks=0;
    const response={status:()=>201};
    const diagnostics={};
    const page={
      locator:()=>({isEnabled:async()=>enabled}),
      waitForResponse:async predicate=>{waiters++;assert.equal(predicate({url:()=> 'http://controlled.invalid/admin/api/upstreams',request:()=>({method:()=> 'POST'})}),true);return response;},
      getByRole:()=>({click:async()=>{clicks++;}})
    };
    const context=vm.createContext({assert,URL,diagnostics});
    vm.runInContext(handler,context);
    if(enabled) assert.equal(await context.save(page,'/admin/api/upstreams'),response);
    else await assert.rejects(context.save(page,'/admin/api/upstreams'));
    assert.equal(waiters,enabled?1:0);assert.equal(clicks,enabled?1:0);
  }
});

test('interrupted editor driver releases only its controlled GET and retains the unlocked/cancel/no-extra-write assertions', async () => {
  const driver = readFileSync(new URL('./console.mjs', import.meta.url), 'utf8');
  const start=driver.indexOf('async function interruptedEditorJourney(');
  const handler=driver.slice(start,driver.indexOf('\ntry {',start));
  for(const {enabled,cleanupFails} of [{enabled:true,cleanupFails:false},{enabled:false,cleanupFails:false},{enabled:true,cleanupFails:true}]) {
    let interceptor,routePattern,listener,pending,released=false,removed=false,cancels=0;
    const receipts=[];
    const page={
      async route(pattern,callback){routePattern=pattern;interceptor=callback;},
      on(event,callback){assert.equal(event,'request');listener=callback;},
      off(event,callback){assert.equal(event,'request');assert.equal(callback,listener);removed=true;},
      async unroute(pattern,callback){assert.equal(pattern,routePattern);assert.equal(callback,interceptor);assert.equal(released,true);await pending;if(cleanupFails)throw new Error('controlled cleanup failure');},
      locator:()=>({isEnabled:async()=>enabled}),
      getByRole:()=>({click:async()=>{cancels++;}})
    };
    const context=vm.createContext({assert,URL,setTimeout,clearTimeout,diagnostics:{},stage:'start',console:{log(){}},writeFile:async(path,value)=>receipts.push({path,value})});
    vm.runInContext(handler,context);
    const write=async()=>{
      listener({method:()=> 'POST',url:()=> 'http://controlled.invalid/admin/api/upstreams'});
      for(const path of ['/admin/api/tenants','/admin/api/tenants?page=1&page_size=100&limit=100','/admin/api/tenants?cursor=controlled%2Bnext'])assert.equal(routePattern.test('http://controlled.invalid'+path),true);
      for(const path of ['/admin/api/tenants/controlled/members','/admin/api/tenants-extra','/admin/api/catalog'])assert.equal(routePattern.test('http://controlled.invalid'+path),false);
      let nonReadContinued=false;
      await interceptor({request:()=>({method:()=> 'POST'}),continue:async()=>{nonReadContinued=true;}});
      assert.equal(nonReadContinued,true);
      const collectionContext=vm.createContext({URLSearchParams,list:value=>value,api:async path=>{assert.equal(routePattern.test('http://controlled.invalid/admin/api/'+path),true);return [];}});
      vm.runInContext(source.slice(source.indexOf('async function collection('),source.indexOf('const scope =')),collectionContext);
      await collectionContext.collection('tenants');
      pending=interceptor({request:()=>({method:()=> 'GET'}),continue:async()=>{await new Promise(resolve=>setImmediate(resolve));released=true;}});
      return 'controlled-acknowledgement';
    };
    const navigation=async title=>{assert.equal(title,'供应商与服务');assert.equal(released,false);assert.equal(context.diagnostics.readback_checkpoint_reached,true);};
    if(cleanupFails) await assert.rejects(context.interruptedEditorJourney(page,navigation,write,1440),/controlled cleanup failure/);
    else if(enabled) assert.equal(await context.interruptedEditorJourney(page,navigation,write,1440),'controlled-acknowledgement');
    else await assert.rejects(context.interruptedEditorJourney(page,navigation,write,1440));
    assert.equal(released,true);assert.equal(removed,true);
    assert.equal(cancels,enabled?2:1);assert.equal(receipts.length,enabled&&!cleanupFails?1:0);
    if(enabled&&!cleanupFails)assert.deepEqual(JSON.parse(receipts[0].value),{suite:'console-editor-interruption',width:1440,result:'passed',controlled_get_hold:true,new_editor_save_enabled:true,service_posts:1,canceled_editor_posts:0});
  }
});

function secretLifecycleFixture() {
  const controls = new Map(), elements = new Map(), secrets = [];
  const create = () => ({ value: '', textContent: '', append() {}, replaceChildren() {}, focus() {}, setAttribute() {} });
  for (const id of ['editor', 'editor-title', 'editor-fields', 'editor-error', 'editor-save', 'editor-form', 'call-key']) elements.set(id, create());
  const editor = elements.get('editor');
  editor.open = false;
  editor.addEventListener = (name, handler) => controls.set(name, handler);
  editor.close = () => { editor.open = false; };
  editor.showModal = () => { editor.open = true; };
  elements.get('editor-fields').querySelectorAll = () => secrets;
  let finishNavigation;
  const state = { epoch: 1, page: 'keys' };
  const context = vm.createContext({
    state, byId: id => elements.get(id),openModal:dialog=>dialog.showModal(),closeModal:dialog=>dialog.close(),releaseModalScroll(){},
    node(tag) { const element = create(); if (tag === 'input') secrets.push(element); return element; },
    button(label, handler) { controls.set(label, handler); return create(); },
    navigate(page) { state.epoch++; state.page = page; return new Promise(resolve => { finishNavigation = resolve; }); },
    clearSensitive() { for (const input of [...secrets, elements.get('call-key')]) input.value = ''; }
  });
  const closeHandler = source.split('\n').find(line => line.startsWith("byId('editor').addEventListener('close'"));
  assert.ok(closeHandler);
  vm.runInContext(closeHandler, context);
  const dialogSource = source.slice(source.indexOf('function secretDialog('), source.indexOf('function invitationDialog('));
  vm.runInContext(dialogSource + "\nsecretDialog({key:'isolated-fixture-value'});", context);
  return { controls, elements, secrets, state, finishNavigation: () => finishNavigation() };
}

test('editor queued close clears its own issued value without erasing a transferred call key', async () => {
  const fixture = secretLifecycleFixture();
  const transfer = fixture.controls.get('显式用于调用测试')();
  fixture.finishNavigation();
  await transfer;
  assert.ok(fixture.elements.get('call-key').value === 'isolated-fixture-value');
  fixture.controls.get('close')();
  assert.ok(fixture.secrets.every(input => input.value === ''));
  assert.ok(fixture.elements.get('call-key').value === 'isolated-fixture-value');
});

test('editor close before transfer settlement still permits the explicitly owned destination', async () => {
  const fixture = secretLifecycleFixture();
  const transfer = fixture.controls.get('显式用于调用测试')();
  fixture.controls.get('close')();
  fixture.finishNavigation();
  await transfer;
  assert.ok(fixture.elements.get('call-key').value === 'isolated-fixture-value');
  assert.ok(fixture.secrets.every(input => input.value === ''));
});

test('superseded navigation cannot restore an issued secret even on another playground epoch', async () => {
  const fixture = secretLifecycleFixture();
  const transfer = fixture.controls.get('显式用于调用测试')();
  fixture.state.epoch++;
  fixture.finishNavigation();
  await transfer;
  assert.equal(fixture.elements.get('call-key').value, '');
});

test('old editor close event cannot erase a reopened one-time dialog', () => {
  const fixture = secretLifecycleFixture();
  fixture.controls.get('close')();
  assert.ok(fixture.secrets.every(input => input.value !== ''));
});
const summaryStart = source.indexOf('function factSummary(');
const summaryEnd = source.indexOf('async function priceEditor(');
assert.ok(summaryStart >= 0 && summaryEnd > summaryStart, 'workspace summary source boundaries');
const summarySource = source.slice(summaryStart, summaryEnd);

function summary(value) {
  const context = vm.createContext({
    node(tag, text = '') { return { tag, text, children: [], append(...values) { this.children.push(...values); },setAttribute(){} }; },
    table(columns, values) { return { tag: 'table', columns, values }; },
    column(label, field) { return { label, field }; },
    money(value) { return String(value); },
    value
  });
  const displaySource=source.slice(source.indexOf('function enumLabel('),source.indexOf('function modelSources('));
  return vm.runInContext(displaySource + '\n' + summarySource + '\nlimitSummary(value);', context);
}

for (const [name, value] of Object.entries({
  service: { id: 1, name: 'controlled source', enabled: true, check: { state: 'unknown' } },
  key: { id: 'controlled-key', models: [], expires_at: null },
  request: { id: 'controlled-request', attempts: [{ ordinal: 1, state: 'success' }], fee: '11' },
  operation: { action: 'key.create', object_id: 'controlled-key' },
  bill: { period: '2026-09', original_fee: '11', corrections: [] }
})) {
  test(`workspace ${name} detail renders facts without requiring limit fields`, () => {
    const result = summary(value);
    assert.equal(result.tag, 'dl');
    assert.equal(result.children.length, Object.keys(value).length * 2);
    assert.equal(result.children[1].children[0].text, String(Object.values(value)[0]));
  });
}

test('workspace detail null remains explicitly unknown', () => {
  assert.equal(summary(null).text, '未知');
});

test('workspace limits retain their independent dimensions and warning', () => {
  const result = summary({ limits: { hard: false, calls: 10 }, usage: { calls: 2, rpm: 0, tpm: 0, concurrency: 0, fee_micro_usd: '1', reserved_micro_usd: '2', pending: 1 }, period: '2026-10', soft_budget_exceeded: true });
  assert.equal(result.children[1].tag, 'table');
  assert.equal(result.children[1].values.length, 5);
  assert.equal(result.children.at(-1).text, '软预算已超额，请核对费用或调整预算。');
});

test('workspace request and price details label actual API fields without changing amounts', () => {
  const value={attempts:[{ordinal:1}],cost_input_rate:'1000000',cost_output_rate:'2000000',tenant_input_rate:'3000000',tenant_output_rate:'4000000'};
  const result=summary(value);
  assert.deepEqual(result.children.filter(child=>child.tag==='dt').map(child=>child.text),['调用尝试','供应商输入成本（微美元 / 百万Token）','供应商输出成本（微美元 / 百万Token）','租户输入价格（微美元 / 百万Token）','租户输出价格（微美元 / 百万Token）']);
  assert.deepEqual(result.children.filter(child=>child.tag==='dd').slice(1).map(child=>child.children[0].text),['1000000','2000000','3000000','4000000']);
});

test('workspace Key budget cards render the shared resource contract instead of raw JSON', async () => {
  const resource={limits:{calls:10},usage:{calls:2,fee_micro_usd:'7',reserved_micro_usd:'3',pending:1},period:'2026-10'};
  const summarized=[];
  const node=(tag,text='')=>({tag,text,children:[],append(...children){this.children.push(...children);},replaceChildren(...children){this.children=children;}});
  const elements=new Map();
  const context=vm.createContext({
    state:{epoch:1,page:'limits',tenant:'controlled'},pages:{limits:['配额与预算','']},
    byId(id){if(!elements.has(id))elements.set(id,node('div'));return elements.get(id);},node,
    needTenant:()=>false,role:()=> 'administrator',tenantPath:path=>'tenants/controlled/'+path,
    api:async()=>resource,collection:async()=>[{id:'app',name:'App'}],
    limitSummary(value){summarized.push(value);return node('resource-summary');},
    mainAction(){},button:()=>node('button')
  });
  const render=source.slice(source.indexOf('async function renderPage('),source.indexOf('function callForm('));
  await vm.runInContext(render+'\nrenderPage();',context);
  assert.equal(summarized.length,2);
  assert.ok(summarized.every(value=>value===resource));
  const content=elements.get('content').children[0];
  const card=content.children.find(child=>child.tag==='div');
  assert.equal(card.children[1].tag,'resource-summary');
  assert.ok(!card.children.some(child=>child.tag==='pre'));
});

test('workspace local list search spans fetched groups and recovers after clearing', () => {
  class Element {
    constructor(tag, text = '') { this.tag = tag; this.text = text; this.children = []; }
    append(...values) { this.children.push(...values); }
    replaceChildren(...values) { this.children = values; }
    setAttribute() {}
  }
  const create = (tag, text) => new Element(tag, text);
  const context = vm.createContext({ Node: Element, node: create, button: create, values: Array.from({ length: 30 }, (_, index) => ({ name: `row-${index}` })) });
  const tableSource = source.slice(source.indexOf('function table('), source.indexOf('const column ='));
  const result = vm.runInContext(tableSource + '\ntable([{label:"name",render:value=>value.name}],values);', context);
  const search = result.children[0], body = result.children[1];
  const text = element => [element.textContent ?? element.text ?? '', ...element.children.map(text)].join(' ');
  search.value = 'row-29';
  search.oninput();
  assert.match(text(body), /row-29/);
  assert.doesNotMatch(text(body), /row-0/);
  search.value = '';
  search.oninput();
  assert.match(text(body), /row-0/);
  assert.match(text(body), /共30项/);
});

async function serviceCheckFixture(rejected = false) {
  const controls = new Map(), requests = [], details = [], notices = [], elements = new Map();
  const service = { id: 7, revision: 19, execution_revision: 4, enabled: true };
  const create = () => ({ append() {}, replaceChildren() {} });
  const context = vm.createContext({
    state: { epoch: 1, page: 'services', user: { platform: true } },
    pages: { services: ['供应商与服务', '服务事实'] },
    byId(id) { if (!elements.has(id)) elements.set(id, create()); return elements.get(id); },
    node: create,
    sourceKind: value => value,
    column(label, field) { return { label, field }; },
    async collection(path) { if (path === 'models') return []; assert.equal(path, 'upstreams'); return [service]; },
    table(columns, values) { return columns.at(-1).render(values[0]); },
    actions(values) { for (const [label, action] of values) controls.set(label, action); return create(); },
    mainAction() {},
    async api(path, method, body) {
      requests.push({ path, method, body: JSON.parse(JSON.stringify(body)) });
      if (rejected) throw new Error('configuration_conflict');
      return { state: 'success', applicable: true, execution_revision: 4 };
    },
    detail(title, value) { details.push({ title, value }); },
    notice(value) { notices.push(value); }
  });
  const renderSource = source.slice(source.indexOf('async function renderPage()'), source.indexOf('function callForm()'));
  await vm.runInContext(renderSource + '\nrenderPage();', context);
  assert.ok(controls.has('核验'), 'actual service row renders check action');
  await controls.get('核验')();
  return { requests, details, notices };
}

test('workspace service check sends execution revision, not management revision or empty input', async () => {
  const result = await serviceCheckFixture();
  assert.deepEqual(result.requests, [{ path: 'upstreams/7/check', method: 'POST', body: { expected_execution_revision: 4 } }]);
  assert.equal(result.details.length, 1);
  assert.equal(result.details[0].value.state, 'success');
  assert.equal(result.notices.length, 0);
});

test('workspace stale service check exposes rejection without detail or automatic retry', async () => {
  const result = await serviceCheckFixture(true);
  assert.deepEqual(result.requests, [{ path: 'upstreams/7/check', method: 'POST', body: { expected_execution_revision: 4 } }]);
  assert.equal(result.details.length, 0);
  assert.deepEqual(result.notices, ['configuration_conflict']);
});

test('console model driver uniquely targets actual editor labels without editing additional routes', async () => {
  let fields;
  const editorContext = vm.createContext({
    async collection(path) { assert.equal(path, 'upstreams'); return [{ id: 7, name: 'Controlled' }]; },
    edit(title, values) { fields = values; }
  });
  const editorSource = source.slice(source.indexOf('async function modelEditor('), source.indexOf('async function grantEditor('));
  await vm.runInContext(editorSource + '\nmodelEditor();', editorContext);
  assert.equal(fields.filter(field => field.label.includes('远端模型名')).length, 2);
  assert.equal(fields.filter(field => field.label === '远端模型名').length, 1);
  const writes = [];
  const driver = readFileSync(new URL('./console.mjs', import.meta.url), 'utf8');
  const start = driver.indexOf("await pressNav('模型托管')");
  const end = driver.indexOf("await save(page,'/admin/api/models')", start);
  assert.ok(start >= 0 && end > start);
  const driverContext = vm.createContext({
    width: 1280, upstream: 7, diagnostics: {}, stage: '', assert,
    async pressNav(title) { assert.equal(title, '模型托管'); },
    page: {
      getByRole() { return { async click() {} }; },
      getByLabel(label, options = {}) {
        const matches = fields.filter(field => options.exact ? field.label === label : field.label.includes(label));
        const write = async value => {
          assert.equal(matches.length, 1, 'model driver label must identify one actual editor control');
          writes.push({ name: matches[0].name, value });
        };
        return { fill: write, selectOption: write, async count() { return matches.length; } };
      }
    }
  });
  await vm.runInContext('(async()=>{' + driver.slice(start, end) + '})()', driverContext);
  assert.deepEqual(writes, [
    { name: 'alias', value: 'model-1280' },
    { name: 'upstream', value: '7' },
    { name: 'remote', value: 'controlled-model' }
  ]);
});

test('console failure snapshot projects only safe enums, counts and booleans', async () => {
  const driver = readFileSync(new URL('./console.mjs', import.meta.url), 'utf8');
  const start = driver.indexOf('async function consoleSnapshot(');
  const end = driver.indexOf('async function save(', start);
  assert.ok(start >= 0 && end > start);
  const secret = 'never-emit-test-secret';
  const elements = {
    editor: { open: true, querySelectorAll: () => [{ value: secret }] },
    'editor-save': { hidden: false, disabled: true },
    'field-models': { options: [{ label: secret }], selectedOptions: [] },
    'call-key': { value: secret, disabled: false },
    'call-form': { querySelectorAll: () => [{ value: secret }, { value: secret }] },
    detail: { open: false },
    notice: { textContent: '读取编辑依赖失败：' + secret },
    'editor-error': { textContent: '写入结果未知，' + secret }
  };
  const context = vm.createContext({
    document: {
      getElementById: id => elements[id],
      querySelector: () => ({ dataset: { page: secret } }),
      querySelectorAll: () => [{ value: secret }]
    },
    page: { evaluate: async callback => callback() }
  });
  const result = await vm.runInContext(driver.slice(start, end) + '\nconsoleSnapshot(page);', context);
  assert.deepEqual(JSON.parse(JSON.stringify(result)), {
    active_page: 'unknown', editor_open: true, save_hidden: false, save_disabled: true,
    editor_invalid_controls: 1, model_option_count: 1, model_selected_count: 0,
    secret_control_count: 1, call_key_present: true, call_key_nonempty: true, call_key_disabled: false,
    call_invalid_controls: 2, detail_open: false, dependency_failure: true, action_failure: false,
    write_failure: false, write_unknown: true
  });
  assert.ok(!JSON.stringify(result).includes(secret));
  elements.notice.textContent = '操作失败：' + secret;
  elements['editor-error'].textContent = '操作失败：' + secret;
  elements['call-key'].value = '';
  const failure = await vm.runInContext('consoleSnapshot(page);', context);
  assert.equal(failure.action_failure, true);
  assert.equal(failure.write_failure, true);
  assert.equal(failure.call_key_nonempty, false);
  assert.ok(!JSON.stringify(failure).includes(secret));
});

test('console failure snapshot tolerates missing controls without inventing facts', async () => {
  const driver = readFileSync(new URL('./console.mjs', import.meta.url), 'utf8');
  const start = driver.indexOf('async function consoleSnapshot(');
  const context = vm.createContext({
    document: { getElementById: () => null, querySelector: () => null, querySelectorAll: () => [] },
    page: { evaluate: async callback => callback() }
  });
  const result = await vm.runInContext(driver.slice(start, driver.indexOf('async function save(', start)) + '\nconsoleSnapshot(page);', context);
  assert.equal(result.active_page, 'unknown');
  for (const [name, value] of Object.entries(result)) if (name !== 'active_page') assert.equal(value, typeof value === 'boolean' ? false : 0);
});

test('console response diagnostics retain route statuses without URLs, bodies or errors', () => {
  const driver = readFileSync(new URL('./console.mjs', import.meta.url), 'utf8');
  const start = driver.indexOf('const currentDiagnostics=diagnostics;');
  const end = driver.indexOf('const pressNav=', start);
  assert.ok(start >= 0 && end > start);
  const events = new Map(), diagnostics = {};
  const context = vm.createContext({ diagnostics, URL, page: { on: (name, handler) => events.set(name, handler) } });
  vm.runInContext(driver.slice(start, end), context);
  const secret = 'never-emit-test-secret';
  events.get('pageerror')(new Error(secret));
  events.get('requestfailed')({ url: () => secret });
  for (const [path, method, status] of [
    ['/admin/api/catalog', 'GET', 403],
    ['/admin/api/tenants/private-tenant/members', 'GET', 200],
    ['/admin/api/tenants/private-tenant/keys', 'GET', 200],
    ['/admin/api/tenants/private-tenant/keys', 'POST', 201],
    ['/v1/chat/completions', 'POST', 401],
    ['/admin/api/workspace/requests/private-request', 'GET', 503],
    ['/admin/api/operations/private-operation', 'GET', 200]
  ]) events.get('response')({
    url: () => 'http://controlled.invalid' + path + '?sensitive=' + secret,
    request: () => ({ method: () => method }), status: () => status,
    json: () => { throw new Error('diagnostics must not read response bodies'); }
  });
  assert.deepEqual(diagnostics, {
    page_error_count: 1, request_failed_count: 1, catalog_status: 403, members_status: 200,
    key_create_status: 201, call_status: 401, request_detail_status: 503
  });
  assert.ok(Object.values(diagnostics).every(value => Number.isSafeInteger(value)));
  assert.ok(!JSON.stringify(diagnostics).includes(secret));
});

function keyboardDriverFixture(outcome = 'success') {
  const driver = readFileSync(new URL('./console.mjs', import.meta.url), 'utf8');
  const start = driver.indexOf('async function keyboardSnapshot(');
  const end = driver.indexOf('async function consoleSnapshot(', start);
  assert.ok(start >= 0 && end > start);
  const secret = 'never-emit-focus-secret';
  const field = { value: secret, id: secret, textContent: secret }, body = {}, root = {}, trigger = { textContent: '创建Key' };
  let tabs = 0;
  const actions = [], editor = { open: false, contains: active => active === field };
  const document = { activeElement: trigger, body, documentElement: root, getElementById: () => editor, hasFocus: () => true };
  const page = {
    getByRole(role, options) {
      assert.equal(role, 'button');
      assert.equal(options.name, '创建Key');
      assert.equal(options.exact, true);
      return { async click() { actions.push('create'); editor.open = outcome !== 'delayed'; document.activeElement = editor.open ? field : trigger; } };
    },
    keyboard: { async press(key) {
      actions.push(key);
      if (key === 'Tab') {
        tabs++;
        document.activeElement = !editor.open || outcome === 'first-tab' || (outcome === 'cycle' && tabs === 13) ? body : field;
      } else if (outcome !== 'escape') {
        editor.open = false;
        document.activeElement = outcome === 'return' ? body : trigger;
      }
    } },
    evaluate: async callback => callback(),
    locator(selector) {
      assert.equal(selector, '#editor');
      return {
        async waitFor(options) {
          assert.equal(options.state, 'visible');
          if (outcome === 'delayed') { editor.open = true; document.activeElement = field; }
          assert.equal(editor.open, true);
        },
        isVisible: async () => editor.open
      };
    }
  };
  const context = vm.createContext({ document, page, diagnostics: {}, stage: '', assert: {
    ok: assert.ok,
    equal: assert.equal
  } });
  vm.runInContext(driver.slice(start, end), context);
  return { context, actions, secret, document, field, body, root };
}

test('keyboard driver retains native actions and all original boundary assertions', async () => {
  const fixture = keyboardDriverFixture();
  await vm.runInContext('keyboardJourney(page,1280)', fixture.context);
  assert.deepEqual(fixture.actions, ['create', ...Array(13).fill('Tab'), 'Escape']);
  assert.equal(fixture.context.stage, 'keyboard-return-focus-1280');
  const trace = JSON.parse(JSON.stringify(fixture.context.diagnostics.keyboard_focus_trace));
  assert.equal(trace.length, 16);
  assert.deepEqual(trace.filter(sample => sample.phase === 'tab').map(sample => sample.tab_count), Array.from({ length: 13 }, (_, index) => index + 1));
  assert.ok(trace.filter(sample => sample.phase === 'tab').every(sample => sample.keyboard_focus_inside && sample.keyboard_editor_open));
  assert.ok(!JSON.stringify(trace).includes(fixture.secret));
});

test('keyboard driver waits for asynchronous editor visibility before its first native Tab', async () => {
  const fixture = keyboardDriverFixture('delayed');
  await vm.runInContext('keyboardJourney(page,1440)', fixture.context);
  assert.deepEqual(fixture.actions, ['create', ...Array(13).fill('Tab'), 'Escape']);
  assert.equal(fixture.context.stage, 'keyboard-return-focus-1440');
  assert.equal(fixture.context.diagnostics.keyboard_focus_trace[0].keyboard_editor_open, true);
});

for (const [outcome, stage, actionCount] of [
  ['first-tab', 'keyboard-first-tab-390', 2],
  ['cycle', 'keyboard-tab-cycle-13-390', 14],
  ['escape', 'keyboard-escape-390', 15],
  ['return', 'keyboard-return-focus-390', 15]
]) test(`keyboard driver preserves ${outcome} rejection and distinguishes its stage`, async () => {
  const fixture = keyboardDriverFixture(outcome);
  await assert.rejects(vm.runInContext('keyboardJourney(page,390)', fixture.context), { name: 'AssertionError' });
  assert.equal(fixture.context.stage, stage);
  assert.equal(fixture.actions.length, actionCount);
  assert.ok(!JSON.stringify(fixture.context.diagnostics).includes(fixture.secret));
});

test('keyboard projection emits only fixed booleans for editor, body, root and missing focus', async () => {
  const fixture = keyboardDriverFixture();
  for (const [active, expectedInside, expectedBody, expectedRoot] of [
    [fixture.field, true, false, false], [fixture.body, false, true, false],
    [fixture.root, false, false, true], [null, false, false, false]
  ]) {
    fixture.document.activeElement = active;
    const snapshot = JSON.parse(JSON.stringify(await vm.runInContext('keyboardSnapshot(page)', fixture.context)));
    assert.deepEqual(snapshot, { keyboard_editor_open: false, keyboard_focus_inside: expectedInside, keyboard_focus_body: expectedBody, keyboard_focus_root: expectedRoot, keyboard_document_focused: true });
    assert.ok(Object.values(snapshot).every(value => typeof value === 'boolean'));
    assert.ok(!JSON.stringify(snapshot).includes(fixture.secret));
  }
});
