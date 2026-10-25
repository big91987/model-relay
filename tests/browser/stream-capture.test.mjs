import assert from 'node:assert/strict';
import test from 'node:test';
import vm from 'node:vm';
import {captureStreamBody} from './stream-capture.mjs';

function fixture(response) {
  const requests=[];
  const window={fetch:async(...args)=>{requests.push(args);return response();}};
  const originalFetch=window.fetch;
  const context=vm.createContext({window,location:new URL('http://controlled.test/admin/'),URL,Request,Error});
  const page={evaluate:callback=>vm.runInContext(`(${callback})()`,context)};
  const clean=()=>{
    assert.equal(window.fetch,originalFetch);
    assert.equal(Object.hasOwn(window,'__relayStreamCapture'),false);
  };
  return {page,window,requests,clean};
}

test('SSE capture reads the same native fetch response without replay or CDP body access',async()=>{
  const frames=['data: {"choices":[{"delta":{"content":"controlled"}}]}\n\n','data: [DONE]\n\n'];
  const state=fixture(()=>new Response(new ReadableStream({start(controller){
    for(const frame of frames)controller.enqueue(new TextEncoder().encode(frame));
    controller.close();
  }}),{headers:{'Content-Type':'text/event-stream'}}));
  const captured=await captureStreamBody(state.page,async()=>{
    const response=await state.window.fetch('/v1/chat/completions',{method:'POST',body:'controlled'});
    assert.equal(await response.text(),frames.join(''));
    return {id:'controlled-request-id'};
  });
  assert.deepEqual(captured,{value:{id:'controlled-request-id'},body:frames.join('')});
  assert.equal(state.requests.length,1);state.clean();
});

test('SSE capture preserves the original action failure and removes the retained body',async()=>{
  const state=fixture(()=>new Response('data: [DONE]\n\n'));
  const failure=new Error('controlled assertion failure');
  await assert.rejects(captureStreamBody(state.page,async()=>{
    await state.window.fetch(new Request('http://controlled.test/v1/chat/completions',{method:'POST'}));
    throw failure;
  }),error=>error===failure);
  assert.equal(state.requests.length,1);state.clean();
});

test('SSE capture never accepts absent, foreign or multiple dispatches as one response',async()=>{
  for(const urls of [[],['http://foreign.test/v1/chat/completions'],['/admin/api/overview'],['/v1/chat/completions','/v1/chat/completions']]) {
    const state=fixture(()=>new Response('data: [DONE]\n\n'));
    await assert.rejects(captureStreamBody(state.page,async()=>{
      for(const url of urls)await (await state.window.fetch(url)).text();
    }),/same-request SSE response capture unavailable/);
    assert.equal(state.requests.length,urls.length);state.clean();
  }
});

test('SSE capture rejects broken bodies without retrying or retaining raw errors',async()=>{
  const state=fixture(()=>new Response(new ReadableStream({start(controller){controller.error(new Error('private-body-error'));}})));
  await assert.rejects(captureStreamBody(state.page,async()=>{
    await state.window.fetch('/v1/chat/completions');
  }),error=>error.code==='ERR_ASSERTION'&&error.message.includes('same-request SSE response capture unavailable')&&!error.message.includes('private-body-error'));
  assert.equal(state.requests.length,1);state.clean();
});

test('SSE capture preserves fetch rejection, restores fetch and never retries',async()=>{
  const failure=new Error('private-network-error');
  const state=fixture(()=>{throw failure;});
  await assert.rejects(captureStreamBody(state.page,()=>state.window.fetch('/v1/chat/completions')),error=>error===failure);
  assert.equal(state.requests.length,1);state.clean();
});
