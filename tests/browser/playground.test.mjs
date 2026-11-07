import assert from 'node:assert/strict';
import test from 'node:test';
import {createPlayground} from '../../internal/relay/web/features/playground/index.mjs';
const deferred=()=>{let resolve;const promise=new Promise(done=>{resolve=done;});return {promise,resolve};};
const document={createElement:tag=>({tag,children:[],dataset:{},value:'',textContent:'',disabled:false,checked:false,append(...children){this.children.push(...children);}})};
function find(element,predicate){if(predicate(element))return element;for(const child of element.children){const found=find(child,predicate);if(found)return found;}}
function fixture(fetch){const notices=[],records=[];const page=createPlayground({document,fetch,randomUUID:()=> 'operation',notify:text=>notices.push(text),openRecord:id=>records.push(id)});return {page,notices,records,control:id=>find(page.element,element=>element.id===id),submit:()=>page.element.onsubmit({preventDefault(){}})};}
test('disposing a pending call aborts and clears secrets without late output or notices',async()=>{
  const pending=deferred();let signal;
  const value=fixture(async(_,options)=>{signal=options.signal;return pending.promise;});
  value.control('call-key').value='private-key';value.control('call-message').value='private-input';
  const call=value.submit();value.page.dispose();assert.equal(signal.aborted,true);
  pending.resolve({ok:true,headers:new Headers({'X-Request-ID':'old'}),json:async()=>({choices:[{message:{content:'private-output'}}]})});
  await call;assert.equal(value.control('call-key').value,'');assert.equal(value.control('call-message').value,'');assert.equal(value.control('call-output').textContent,'');assert.deepEqual(value.notices,[]);
});
test('playground owns one active call and keeps the returned record ID',async()=>{
  const pending=deferred();let calls=0,request;
  const value=fixture(async(path,options)=>{calls++;request={path,options};return pending.promise;});
  value.control('call-key').value='key';value.control('call-model').value='model';value.control('call-message').value='hello';
  const call=value.submit();await value.submit();assert.equal(calls,1);
  pending.resolve({ok:true,headers:new Headers({'X-Request-ID':'record'}),json:async()=>({choices:[{message:{content:'reply'}}]})});await call;
  assert.equal(value.control('call-output').textContent,'reply');assert.equal(request.options.headers.Authorization,'Bearer key');
  const record=find(value.page.element,element=>element.textContent==='查看同Request ID日志');assert.equal(record.disabled,false);record.onclick();assert.deepEqual(value.records,['record']);
});
test('a stream without terminal DONE stays a visible failure',async()=>{
  const data='data: {"choices":[{"delta":{"content":"partial"}}]}\n\n';
  const value=fixture(async()=>({ok:true,headers:new Headers(),body:new Response(data).body}));value.control('call-stream').checked=true;
  await value.submit();assert.match(value.control('call-output').textContent,/partial[\s\S]*流未正常结束/);assert.deepEqual(value.notices,[]);
});

test('failed record navigation remains a visible action error',async()=>{
  const notices=[];
  const page=createPlayground({document,notify:text=>notices.push(text),openRecord:async()=>{throw new Error('denied');}});
  const record=find(page.element,element=>element.textContent==='查看同Request ID日志');
  await record.onclick();assert.deepEqual(notices,['操作失败：denied']);
});

test('default network request uses the current browser fetch at call time',async context=>{
  context.mock.method(globalThis,'fetch',async()=>assert.fail('stale browser fetch'));
  const page=createPlayground({document,notify(){},openRecord(){}});
  let calls=0;context.mock.method(globalThis,'fetch',async()=>{calls++;return {ok:true,headers:new Headers(),json:async()=>({choices:[{message:{content:'reply'}}]})};});
  await page.element.onsubmit({preventDefault(){}});assert.equal(calls,1);
});
