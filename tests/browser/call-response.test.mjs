import assert from 'node:assert/strict';
import test from 'node:test';
import vm from 'node:vm';
import {withStreamCapture} from './call-response.mjs';

function fixture({id='same-request',body='data: controlled\n\ndata: [DONE]\n\n',bodyFailure=false}={}) {
  let dispatches=0;
  const previous=async function(url){
    assert.equal(this,window);
    dispatches++;
    const response=new Response(body,{headers:{'x-request-id':id}});
    Object.defineProperty(response,'url',{value:new URL(url,'http://controlled.test').href});
    if(bodyFailure)response.clone=()=>({text:()=>Promise.reject(new Error('private-body-error'))});
    return response;
  };
  const window={fetch:previous};
  const context=vm.createContext({window,URL,location:{href:'http://controlled.test/admin/'}});
  const page={evaluate:callback=>vm.runInContext('('+callback.toString()+')()',context)};
  return {page,window,previous,dispatches:()=>dispatches,send:async()=>{
    const response=await window.fetch('/v1/chat/completions',{method:'POST'});
    assert.equal(await response.text(),body);
    return {id:'same-request'};
  }};
}

test('SSE observer reads the same response clone without CDP or a second dispatch',async()=>{
  const current=fixture();
  const result=await withStreamCapture(current.page,current.send);
  assert.equal(result.id,'same-request');assert.ok(result.value.includes('data:'));assert.ok(result.value.includes('[DONE]'));
  assert.equal(current.dispatches(),1);assert.equal(current.window.fetch,current.previous);
  assert.equal('__relayHarnessStreamCapture' in current.window,false);
});
for(const [name,options,message] of [
  ['wrong request ID',{id:'different'},/same request ID/],
  ['failed clone read',{bodyFailure:true},/body read must succeed/]
])test('SSE observer fails closed on '+name,async()=>{
  const current=fixture(options);
  await assert.rejects(withStreamCapture(current.page,current.send),message);
  assert.equal(current.dispatches(),1);assert.equal(current.window.fetch,current.previous);
  assert.equal('__relayHarnessStreamCapture' in current.window,false);
});
test('SSE observer rejects missing and duplicate calls rather than weakening one-dispatch assertions',async()=>{
  for(const count of [0,2]){
    const current=fixture();
    await assert.rejects(withStreamCapture(current.page,async()=>{for(let index=0;index<count;index++)await current.send();return {id:'same-request'};}),/exactly one dispatch/);
    assert.equal(current.dispatches(),count);assert.equal(current.window.fetch,current.previous);
  }
});
test('SSE observer restores fetch on native send failure and retains the original error',async()=>{
  const current=fixture();const error=new Error('private-send-error');
  await assert.rejects(withStreamCapture(current.page,async()=>{throw error;}),value=>value===error);
  assert.equal(current.window.fetch,current.previous);assert.equal('__relayHarnessStreamCapture' in current.window,false);
});
test('SSE observer ignores other paths and GET requests',async()=>{
  const current=fixture();
  const result=await withStreamCapture(current.page,async()=>{
    await current.window.fetch('/healthz',{method:'POST'});
    await current.window.fetch('/v1/chat/completions',{method:'GET'});
    return current.send();
  });
  assert.ok(result.value.includes('[DONE]'));assert.equal(current.dispatches(),3);
});
test('SSE observer does not overwrite a newer fetch wrapper during cleanup',async()=>{
  const current=fixture();const newer=()=>{};
  await withStreamCapture(current.page,async()=>{const response=await current.send();current.window.fetch=newer;return response;});
  assert.equal(current.window.fetch,newer);assert.equal('__relayHarnessStreamCapture' in current.window,false);
});
