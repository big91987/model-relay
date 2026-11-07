import assert from 'node:assert/strict';
import test from 'node:test';
import {createTransport} from '../../internal/relay/web/shared/transport.mjs';

const deferred=()=>{let resolve;const promise=new Promise(done=>{resolve=done;});return {promise,resolve};};
test('scope cancellation aborts pending requests and discards a late response',async()=>{
  let session={epoch:1,user:'first',csrf:'token'},signal;
  const pending=deferred(),errors=[];
  const transport=createTransport({session:()=>session,onError:(status,path)=>errors.push([status,path]),fetch:async(_,options)=>{signal=options.signal;return pending.promise;}});
  const reading=transport.request('catalog');
  session={epoch:2,user:'second',csrf:'next'};transport.cancel();
  assert.equal(signal.aborted,true);
  pending.resolve({ok:true,json:async()=>({secret:'old'})});
  await assert.rejects(reading,{name:'AbortError'});assert.deepEqual(errors,[]);
});
test('transport preserves CSRF and idempotency and delegates current authorization failures',async()=>{
  const requests=[],errors=[];
  const transport=createTransport({session:()=>({epoch:1,user:'member',csrf:'csrf'}),onError:(status,path)=>errors.push([status,path]),fetch:async(path,options)=>{requests.push({path,options});return {ok:false,status:403,json:async()=>({error:{code:'denied'}})};}});
  await assert.rejects(transport.request('keys','POST',{name:'demo'},'one-operation'),error=>error.status===403&&error.message==='denied');
  assert.equal(requests[0].path,'/admin/api/keys');assert.equal(requests[0].options.headers['X-CSRF-Token'],'csrf');
  assert.equal(requests[0].options.headers['Idempotency-Key'],'one-operation');assert.equal(requests[0].options.body,'{"name":"demo"}');
  assert.deepEqual(errors,[[403,'keys']]);
});
test('downloads also reject data from a changed scope',async()=>{
  const pending=deferred();let epoch=1;
  const transport=createTransport({session:()=>({epoch}),onError(){},fetch:async()=>({ok:true,blob:()=>pending.promise})});
  const reading=transport.download('usage/export');await Promise.resolve();epoch++;
  pending.resolve(new Blob(['old']));await assert.rejects(reading,{name:'AbortError'});
});

test('default transport uses the current browser fetch at request time',async context=>{
  context.mock.method(globalThis,'fetch',async()=>assert.fail('stale browser fetch'));
  const transport=createTransport({session:()=>({epoch:1}),onError(){}});
  let calls=0;context.mock.method(globalThis,'fetch',async()=>{calls++;return {ok:true,json:async()=>({items:[]})};});
  await transport.request('catalog');assert.equal(calls,1);
});
