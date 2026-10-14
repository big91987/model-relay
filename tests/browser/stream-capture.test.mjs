import assert from 'node:assert/strict';
import test from 'node:test';
import {installStreamingCapture,readStreamingCapture} from './stream-capture.mjs';

test('captures the same native response without another dispatch and restores fetch',async()=>{
  const body='data: {"choices":[{"delta":{"content":"controlled stream"}}]}\n\ndata: [DONE]\n\n';
  const response=new Response(body,{headers:{'Content-Type':'text/event-stream'}});
  const calls=[];
  const target={location:{href:'http://127.0.0.1/admin/'},fetch:async(...args)=>{calls.push(args);return response;}};
  const originalFetch=target.fetch;
  installStreamingCapture(target);
  const input={method:'POST',body:JSON.stringify({stream:true})};
  const observed=await target.fetch('/v1/chat/completions',input);
  assert.equal(observed,response);
  assert.equal(target.fetch,originalFetch);
  assert.deepEqual(calls,[['/v1/chat/completions',input]]);
  assert.equal(await observed.text(),body);
  assert.equal(await readStreamingCapture(target),body);
  target.__relayStreamCapture.restore();
  assert.equal(target.fetch,originalFetch);
  assert.equal(target.__relayStreamCapture,undefined);
});

test('missing capture fails rather than inventing SSE frames',async()=>{
  await assert.rejects(readStreamingCapture({}),/stream capture unavailable/);
});

test('failed clone read fails rather than hiding the error',async()=>{
  const target={location:{href:'http://127.0.0.1/admin/'},fetch:async()=>({clone:()=>({text:async()=>{throw new Error('controlled failure');}})})};
  installStreamingCapture(target);
  await target.fetch('/v1/chat/completions');
  await assert.rejects(readStreamingCapture(target),/stream capture failed/);
  target.__relayStreamCapture.restore();
});

test('unrelated reads are not captured and cancellation restores fetch',async()=>{
  const target={location:{href:'http://127.0.0.1/admin/'},fetch:async()=>new Response('{}')};
  const originalFetch=target.fetch;
  installStreamingCapture(target);
  await target.fetch('/admin/api/workspace/requests');
  await assert.rejects(readStreamingCapture(target),/stream capture unavailable/);
  target.__relayStreamCapture.restore();
  assert.equal(target.fetch,originalFetch);
});

test('transport rejection remains a failure',async()=>{
  const target={location:{href:'http://127.0.0.1/admin/'},fetch:async()=>{throw new Error('transport rejected');}};
  installStreamingCapture(target);
  await assert.rejects(target.fetch('/v1/chat/completions'),/transport rejected/);
  target.__relayStreamCapture.restore();
});
