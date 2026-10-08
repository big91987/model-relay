import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';

const workspace=readFileSync('internal/relay/web/workspace.js','utf8');
const journey=readFileSync('tests/browser/journeys.mjs','utf8');
const html=readFileSync('internal/relay/web/workspace.html','utf8');
const backend=readFileSync('internal/relay/console_keys.go','utf8');

test('journey failure boundary includes summary assertion source but never raw stack paths or errors',()=>{
  const helper=journey.slice(journey.indexOf('function failureBoundary('),journey.indexOf('function conservationFacts('));
  const context=vm.createContext({});
  const boundary=vm.runInContext(helper+'failureBoundary;',context);
  for(const source of ['journeys','console-summary']) {
    const value=boundary({stack:`AssertionError: private-error\n    at summaryCell (file:///private/local/tests/browser/${source}.mjs:37:12)\n`});
    assert.deepEqual(JSON.parse(JSON.stringify(value)),{source:`tests/browser/${source}.mjs`,line:37,column:12});
  }
  for(const error of [{},{stack:'private-error'},{stack:'Error\n    at file:///private/local/tests/browser/private-module.mjs:1:2\n'}])assert.deepEqual(JSON.parse(JSON.stringify(boundary(error))),{available:false});
  assert.ok(journey.includes('summary_cell:summaryFailureFacts(error)'));
  assert.ok(journey.includes('content_loading:'));
});

async function requestDetailContract(entry='call',clickFailure=false) {
  const events=[];
  const response={url:()=> 'http://controlled.test/admin/api/workspace/requests/controlled-request?tenant=controlled',status:()=>200,json:async()=>({id:'controlled-request'})};
  const page={
    waitForResponse:predicate=>{events.push('response-armed');assert.equal(predicate(response),true);return Promise.resolve(response);},
    locator:selector=>{assert.equal(selector,'#detail-title');return {filter:options=>{assert.equal(options.hasText,'controlled-request');return {waitFor:async()=>events.push('detail-visible')};}};}
  };
  const click=kind=>({click:async()=>{events.push(kind);if(clickFailure)throw new Error('controlled-click-failure');}});
  const context=vm.createContext({assert,URL,page,entry,action:'',
    button:(target,name)=>{assert.equal(target,page);assert.equal(name,'查看同Request ID日志');assert.equal(entry,'call');return click('call-click');},
    row:(target,id)=>{assert.equal(target,page);assert.equal(id,'controlled-request');assert.equal(entry,'records');return {getByRole:(role,options)=>{assert.equal(role,'button');assert.equal(options.name,'详情 / 核对');assert.equal(options.exact,true);return click('records-click');}};}
  });
  const helper=journey.slice(journey.indexOf('async function requestDetail('),journey.indexOf('async function editKey('));
  const run=vm.runInContext(helper+"requestDetail(page,'controlled-request',entry);",context);
  if(clickFailure)await assert.rejects(run,/controlled-click-failure/);
  else {const value=await run;assert.equal(value.id,'controlled-request');assert.deepEqual(events,['response-armed',entry+'-click','detail-visible']);}
}

test('request detail uses the call action on the call page',async()=>requestDetailContract());
test('request detail uses the exact request row on the records page',async()=>requestDetailContract('records'));
test('request detail retains a failed native activation',async()=>requestDetailContract('records',true));
test('J06 reconciled detail activates the records entry',()=>{
  assert.ok(journey.includes("const confirmedDetail=await requestDetail(admin,unknown.id,'records');"));
});

async function visualActivityContract(change={}) {
  const events=[];
  const before={today:'2026-11-01',period:'2026-11',requests_today:0,trend:[]};
  const after={...before,requests_today:1,success_rate_percent:100,confirmed_fee_micro_usd:'21',trend:[{day:before.today,requests:1,successes:1}],...change.overview};
  const detail={pending:0,period:before.period,charge_entries:[{fee:'21',cost:'7'}],...change.detail};
  const bill={original_fee:'126',corrections:[{amount:'5',period:before.period}]};
  let overviewReads=0,billReads=0;
  const helper=journey.slice(journey.indexOf('async function prepareVisualActivity('),journey.indexOf('async function safeScreenshot('));
  assert.ok(helper.startsWith('async function prepareVisualActivity('),'real current-period preparation exists');
  const context=vm.createContext({assert,now:'2026-11-01T00:00:00Z',
    read:async(page,path)=>{events.push(path);if(path==='workspace/overview?tenant=scope')return overviewReads++?after:before;assert.equal(path,'bills/closed');return billReads++?{...bill,...change.bill}:bill;},
    call:async(page,key,model)=>{events.push('native-call');assert.equal(key,'ephemeral-key');assert.equal(model,'public-model');return {id:'actual-request'};},
    requestDetail:async(page,id)=>{assert.equal(id,'actual-request');return detail;},
    closeDetail:async()=>events.push('close-detail'),nav:async(page,title)=>{assert.equal(title,'工作台');events.push('native-cleanup-navigation');}
  });
  const facts=await vm.runInContext(helper+"prepareVisualActivity({},'ephemeral-key','public-model','scope','closed')",context);
  return {facts,events};
}
test('current-period visual preparation uses the existing native call and conserves the closed bill',async()=>{
  const {facts,events}=await visualActivityContract();
  assert.equal(facts.requests_before,0);assert.equal(facts.requests_after,1);assert.equal(facts.closed_bill_unchanged,true);
  assert.deepEqual(events,['workspace/overview?tenant=scope','bills/closed','native-call','close-detail','workspace/overview?tenant=scope','bills/closed','native-cleanup-navigation']);
  const entry=journey.slice(journey.indexOf("complete('J06');"),journey.indexOf('  await consoleVisualReview('));
  assert.ok(entry.includes('await prepareVisualActivity(admin,current.key,models[0],firstTenant,bill.id)'));
});
test('visual preparation rejects stale or missing current-period activity',async()=>{
  for(const overview of [{requests_today:0},{trend:[]},{period:'2026-10'},{confirmed_fee_micro_usd:'0'}])await assert.rejects(visualActivityContract({overview}));
  for(const detail of [{pending:1},{period:'2026-10'},{charge_entries:[{fee:'0',cost:'7'}]}])await assert.rejects(visualActivityContract({detail}));
});
test('visual preparation rejects changes to closed bill amount or corrections',async()=>{
  for(const bill of [{original_fee:'147'},{corrections:[]}])await assert.rejects(visualActivityContract({bill}));
});

test('J06 reconciliation reads append-only cumulative charges, not the first revision',async()=>{
  const prior={ordinal:1,revision:0,input_tokens:0,output_tokens:0,cost:'0',fee:'0',reason:'',evidence:''};
  const delta={ordinal:1,revision:2,input_tokens:3,output_tokens:4,cost:'7',fee:'21',reason:'controlled authoritative test usage',evidence:'controlled fixture HTTP usage contract; not supplier validation'};
  const pending={pending:1,revision:1,fee_reserved:'600',price_snapshots:[{id:'original-price'}],charge_entries:[prior]};
  const reconciled={pending:0,revision:2,fee_reserved:'0',price_snapshots:pending.price_snapshots,charge_entries:[prior,delta]};
  const segment=journey.slice(journey.indexOf('  const reconciled=await read('),journey.indexOf("  mark('J06-reconciled-token-meaning')"));
  const helper=journey.includes('function assertReconciledLedger(')?journey.slice(journey.indexOf('function assertReconciledLedger('),journey.indexOf('async function freshInstallation(')):'';
  const run=async value=>{
    const context=vm.createContext({assert,pending,read:async()=>value,admin:{},unknown:{id:'controlled-id'},firstTenant:'controlled-tenant',ledgerFacts:{},action:''});
    vm.runInContext(helper,context);
    await vm.runInContext('(async()=>{'+segment+'})()',context);
    return context;
  };
  const context=await run(reconciled);
  assert.equal(context.ledgerFacts.fee,'21');assert.equal(context.ledgerFacts.cost,'7');assert.equal(context.ledgerFacts.entries,2);
  for(const value of [
    {...reconciled,charge_entries:[prior,{...delta,fee:'20'}]},
    {...reconciled,charge_entries:[prior,{...delta,cost:'8'}]},
    {...reconciled,charge_entries:[{...prior,fee:'1'},delta]},
    {...reconciled,charge_entries:[prior,delta,delta]},
    {...reconciled,pending:1},
    {...reconciled,fee_reserved:'1'},
    {...reconciled,price_snapshots:[{id:'changed-price'}]},
    {...reconciled,charge_entries:[prior,{...delta,input_tokens:2}]},
    {...reconciled,charge_entries:[prior,{...delta,revision:1}]},
    {...reconciled,revision:3},
    {...reconciled,charge_entries:[prior,{...delta,fee:'forbidden-secret'}]},
  ])await assert.rejects(run(value),{name:'AssertionError'});
  assert.ok(!JSON.stringify(context.ledgerFacts).includes('controlled-id'));
});

test('scope refresh cannot complete on old content before asynchronous tenant removal settles',async()=>{
  const segment=journey.slice(journey.indexOf('async function refresh('),journey.indexOf('async function tabTo('));
  let release,finished=false,readyCalls=0,waits=0;
  const removed=new Promise(resolve=>{release=resolve;});
  const selector={options:[{value:'removed-tenant'}]};
  const context=vm.createContext({
    button:()=>({click:async()=>{removed.then(()=>{selector.options=[];});}}),
    ready:async()=>{readyCalls++;},
    page:{waitForFunction:async(predicate,expected,options)=>{
      assert.equal(options.timeout,5000);
      waits++;context.expected=expected;
      assert.equal(vm.runInContext('('+predicate.toString()+')(expected)',context),false);
      await removed;
      assert.equal(vm.runInContext('('+predicate.toString()+')(expected)',context),true);
    }},
    document:{getElementById:()=>selector},
  });
  const run=vm.runInContext(segment+'\nrefresh(page,0);',context).then(()=>{finished=true;});
  await new Promise(resolve=>setImmediate(resolve));
  try {assert.equal(finished,false);assert.equal(readyCalls,0);assert.equal(waits,1);}
  finally {release();await run;}
  assert.equal(finished,true);assert.equal(readyCalls,1);
  assert.ok(journey.includes('await refresh(developer,0)'));
});

test('failure boundary projects only the owning source position, never assertion values or host paths',()=>{
  const segment=journey.slice(journey.indexOf('function failureBoundary('),journey.indexOf('function conservationFacts('));
  const context=vm.createContext({});
  vm.runInContext(segment,context);
  for(const location of ['at file:///private/secret-checkout/tests/browser/journeys.mjs:363:44','at read (file:///private/secret-checkout/tests/browser/journeys.mjs:90:10)']){
    context.error={stack:'AssertionError: forbidden-secret\n    '+location+'\n    at arbitrary-file:1:1'};
    const result=vm.runInContext('failureBoundary(error);',context);
    assert.equal(result.source,'tests/browser/journeys.mjs');assert.ok(result.line>0);assert.ok(result.column>0);
    assert.ok(!JSON.stringify(result).includes('secret'));
  }
  for(const error of [{},{stack:'forbidden-secret'},{stack:'\n    at file:///private/secret-checkout/unrelated.mjs:1:1'}]){
    context.error=error;assert.equal(vm.runInContext('failureBoundary(error).available;',context),false);
  }
});

test('J05 member transitions have independent attribution without removing scope assertions',()=>{
  for(const stage of ['member-lookup','role-change','role-readback','diagnostic-recovery','member-remove','removed-key-scope','removed-membership','removed-ui-scope','application-survives','other-tenant-request'])assert.ok(journey.includes(`mark('J05-${stage}')`));
  assert.ok(journey.includes('assert.equal(freshRole.role,\'finance\')'));
  assert.ok(journey.includes('assert.ok(diagnosticMembers.length===0)'));
  assert.ok(journey.includes('assert.equal(remainingTenants.length,0)'));
  assert.ok(journey.includes('assert.equal(membershipFacts.tenant_option_count,0)'));
  assert.ok(journey.includes('read_facts:readFacts'));assert.ok(journey.includes('membership_facts:membershipFacts'));
});

test('pause conservation evidence only projects exact amounts and safe counters',()=>{
  const segment=journey.slice(journey.indexOf('function conservationFacts('),journey.indexOf('async function freshInstallation('));
  const forbidden='forbidden-account-key-or-prompt';
  const before={period:'2026-10',usage:{fee_micro_usd:'9007199254740993',reserved_micro_usd:'0',calls:2,pending:0,concurrency:0},secret:forbidden};
  const after=structuredClone(before);
  const context=vm.createContext({assert,before,after,action:'',usageFacts:{}});
  vm.runInContext(segment+'\nassertUsageConserved(before,after);',context);
  assert.equal(context.action,'usage-conservation');
  assert.equal(context.usageFacts.fee_micro_usd.before,'9007199254740993');
  assert.ok(!JSON.stringify(context.usageFacts).includes(forbidden));
  for(const field of ['fee_micro_usd','reserved_micro_usd','calls','pending','concurrency']){
    after.usage[field]=typeof before.usage[field]==='string'?String(BigInt(before.usage[field])+1n):before.usage[field]+1;
    assert.throws(()=>vm.runInContext('assertUsageConserved(before,after);',context),{name:'AssertionError'});
    after.usage[field]=before.usage[field];
  }
  after.period='2026-11';
  assert.throws(()=>vm.runInContext('assertUsageConserved(before,after);',context),{name:'AssertionError'});
  after.period=before.period;
  for(const invalid of [undefined,forbidden,12,'1'.repeat(20)]){
    after.usage.fee_micro_usd=invalid;
    assert.throws(()=>vm.runInContext('assertUsageConserved(before,after);',context),{name:'AssertionError'});
    assert.equal(context.usageFacts.fee_micro_usd.after,null);
    assert.ok(!JSON.stringify(context.usageFacts).includes(forbidden));
  }
});

test('J05 retains original resume fee assertion and checks each independent rejection boundary',()=>{
  for(const stage of ['J05-paused-ordinary-usage','J05-paused-stream-usage','J05-paused-standby-usage','J05-other-tenant-usage'])assert.ok(journey.includes(`mark('${stage}')`));
  assert.ok(journey.includes('assert.equal(resumedUsage.usage.fee_micro_usd,pauseUsage.usage.fee_micro_usd)'));
  assert.ok(journey.includes('usage_facts:usageFacts'));
});

test('completion failure evidence is a bounded whitelist, never raw provider or request data',()=>{
  const segment=journey.slice(journey.indexOf('function completionFacts('),journey.indexOf('async function requestDetail('));
  const secret='forbidden-secret-and-prompt';
  const input={state:'upstream_failed',error_code:'upstream_protocol_error',upstream_status:200,pending:1,key:secret,id:secret,message:secret,attempts:Array.from({length:8},()=>({state:'upstream_failed',error_code:'upstream_protocol_error',upstream_status:200,dispatch_possible:1,prompt_tokens:null,completion_tokens:null,remote_model:secret,upstream_name:secret}))};
  const context=vm.createContext({input});
  const facts=vm.runInContext(segment+'\ncompletionFacts(input);',context);
  assert.equal(facts.attempt_count,8);assert.equal(facts.attempts.length,3);
  assert.equal(facts.terminal_code,'upstream_protocol_error');assert.equal(facts.upstream_status,200);
  assert.equal(facts.attempts[0].dispatch_possible,true);assert.equal(facts.attempts[0].input_known,false);
  assert.ok(!JSON.stringify(facts).includes(secret));
  input.state=secret;input.error_code=secret;input.upstream_status=secret;input.attempts=[{state:secret,error_code:secret,upstream_status:secret}];
  const hostile=vm.runInContext('completionFacts(input);',context);
  assert.equal(hostile.terminal,'other');assert.equal(hostile.terminal_code,'other');assert.equal(hostile.upstream_status,null);
  assert.ok(!JSON.stringify(hostile).includes(secret));
  assert.ok(journey.includes('completionFacts(await detail.json())'));
});

test('actual fresh-install driver requires zero collections, not migrated legacy tenants',async()=>{
  const segment=journey.slice(journey.indexOf('async function freshInstallation('),journey.indexOf('async function close('));
  for(const [tenants,options,accepted] of [[[],1,true],[[{id:'legacy'}],2,false],[[],0,false],[null,1,false]]){
    const reads=[],stages=[],initialCounts={};
    const context=vm.createContext({assert,initialCounts,items:value=>value.items??value.data??value,mark:value=>stages.push(value),read:async(page,path)=>{reads.push(path);return {items:path==='tenants'?tenants:[]};},page:{getByLabel:()=>({inputValue:async()=>'',locator:()=>({count:async()=>options,textContent:async()=>'全部租户'})})}});
    const run=vm.runInContext(segment+'\nfreshInstallation(page);',context);
    if(accepted)await run;else await assert.rejects(run,assert.AssertionError);
    assert.deepEqual(reads,['upstreams','models','tenants']);
    assert.deepEqual(stages,['J01-empty-upstreams','J01-empty-models','J01-empty-tenants']);
    assert.equal(initialCounts.upstreams,0);assert.equal(initialCounts.models,0);
  }
  assert.ok(journey.includes('await freshInstallation(admin);'));
  assert.ok(!journey.includes("read(admin,'tenants/legacy/keys')"));
});

function keyStatus(value,now) {
  const segment=workspace.slice(workspace.indexOf('function keyStatus('),workspace.indexOf('async function collection('));
  const context=vm.createContext({value,now});
  return vm.runInContext(segment+'\nkeyStatus(value,now);',context);
}

test('production Key formatter distinguishes safe states and precise grace/expiry boundaries',()=>{
  const now=Date.parse('2026-10-08T00:00:00Z');
  const base={enabled:true,revoked:false,expires_at:null,generations:[{retired:0,expires:null}]};
  assert.equal(keyStatus(base,now),'已启用');
  assert.equal(keyStatus({...base,enabled:false},now),'已停用');
  assert.equal(keyStatus({...base,revoked:true,enabled:false},now),'已撤销');
  assert.equal(keyStatus({...base,expires_at:'2026-10-08T00:00:00Z'},now),'已过期');
  assert.equal(keyStatus({...base,expires_at:'2026-10-07T23:59:59Z'},now),'已过期');
  assert.equal(keyStatus({...base,expires_at:'2026-10-08T00:00:00.001Z'},now),'已启用');
  assert.equal(keyStatus({...base,generations:[...base.generations,{retired:0,expires:'2026-10-08T01:00:00Z'}]},now),'过渡中');
  assert.equal(keyStatus({...base,generations:[...base.generations,{retired:1,expires:'2026-10-08T01:00:00Z'}]},now),'已启用');
  assert.equal(keyStatus({...base,generations:[...base.generations,{retired:0,expires:'2026-10-08T00:00:00Z'}]},now),'已启用');
  for(const partial of [{},{...base,enabled:null},{...base,revoked:null},{...base,generations:null},{...base,expires_at:'invalid'}])assert.equal(keyStatus(partial,now),'未知');
});

test('continuous driver uses actual product labels and rotation receipt contract',()=>{
  for(const label of ['新账户密码（至少12字符）'])assert.ok(html.includes(label));
  for(const label of ['邀请接受链接（手工交付）',"['上界',",'供应商可证明最大收费输入Token','包括隐藏推理等最大收费输出Token','有符号更正金额（微美元）'])assert.ok(workspace.includes(label));
  assert.ok(journey.includes("getByLabel('新账户密码（至少12字符）')"));
  assert.ok(journey.includes("hasText:'邀请接受链接（手工交付）'"));
  assert.ok(journey.includes("name:'上界',exact:true"));
  assert.ok(journey.includes('hours,status=200'));
  assert.ok(backend.includes('reply(writer, 200, map[string]any{"id": keyID, "key": secret'));
  assert.ok(!journey.includes("hasText:'手工交付邀请'"));
});

test('scope-required empty state replaces loading instead of leaving a perpetual loading placeholder',()=>{
  const content={children:[{text:'正在读取当前范围…'}],replaceChildren(...children){this.children=children;}};
  const state={tenant:''};
  const segment=workspace.slice(workspace.indexOf('function needTenant('),workspace.indexOf('function role('));
  const context=vm.createContext({state,byId:()=>content,node:(tag,text,className)=>({tag,text,className})});
  assert.equal(vm.runInContext(segment+'\nneedTenant();',context),true);
  assert.equal(content.children.length,1);assert.equal(content.children[0].className,'empty');
  assert.ok(!content.children[0].text.includes('正在读取'));
  state.tenant='controlled-tenant';
  assert.equal(vm.runInContext('needTenant();',context),false);
  assert.equal(content.children.length,1);
});

test('request-detail log navigation keeps tenant and existing safe filters',async()=>{
  const state={tenant:'controlled-tenant',filters:{key:'controlled-key',model:'controlled-model',from:'2026-10-01T00:00:00Z'},pageNumber:2};
  let action,navigation;
  const dialog={showModal(){}};
  const context=vm.createContext({state,byId:id=>id==='detail'?dialog:{replaceChildren(){},append(){},focus(){}},limitSummary:()=>({}),button:(label,handler)=>{action=handler;return {};},navigate:async(page,preserve)=>{navigation={page,preserve};}});
  const segment=workspace.slice(workspace.indexOf('function detail('),workspace.indexOf("byId('detail-close').onclick"));
  vm.runInContext(segment+"\ndetail('请求详情 · controlled-request',{id:'controlled-request'});",context);
  await action();
  assert.equal(state.tenant,'controlled-tenant');assert.equal(state.filters.key,'controlled-key');assert.equal(state.filters.model,'controlled-model');assert.equal(state.filters.from,'2026-10-01T00:00:00Z');assert.equal(state.filters.request_id,'controlled-request');assert.equal(state.pageNumber,1);assert.deepEqual(navigation,{page:'records',preserve:true});
});

test('full suite stays additive to make verify with every frozen journey and all viewports',()=>{
  const makefile=readFileSync('Makefile','utf8');
  const runner=readFileSync('tests/browser/main.go','utf8');
  const controlled=readFileSync('tests/browser/journeys.go','utf8');
  for(let index=1;index<=7;index++)assert.ok(journey.includes(`complete('J0${index}')`));
  assert.ok(runner.includes('consoleJourney()'));assert.ok(runner.includes('fullConsoleJourneys()'));
  assert.ok(controlled.includes('[]int{1280, 1440, 390}'));
  for(const gate of ['health-history','frontend-test','test','build','browser'])assert.ok(makefile.includes('$(MAKE) '+gate));
  assert.ok(!controlled.includes('.DB.Exec('));assert.ok(!journey.includes('.request.post('));assert.ok(!journey.includes('.request.put('));assert.ok(!journey.includes('.request.patch('));
});
