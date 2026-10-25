import assert from 'node:assert/strict';

export async function captureStreamBody(page,action) {
  await page.evaluate(()=>{
    if(window.__relayStreamCapture)throw new Error('stream capture already active');
    const state={originalFetch:window.fetch,requests:0,body:null};
    state.fetch=async(...args)=>{
      const response=await state.originalFetch.apply(window,args);
      const input=args[0];
      const url=new URL(input instanceof Request?input.url:String(input),location.href);
      if(url.origin===location.origin&&url.pathname==='/v1/chat/completions') {
        state.requests++;
        state.body=response.clone().text().then(text=>({ok:true,text}),()=>({ok:false}));
      }
      return response;
    };
    window.__relayStreamCapture=state;window.fetch=state.fetch;
  });
  try {
    const value=await action();
    const capture=await page.evaluate(async()=>{
      const state=window.__relayStreamCapture;
      if(state.requests!==1||!state.body)return {ok:false};
      return state.body;
    });
    assert.equal(capture.ok,true,'same-request SSE response capture unavailable');
    return {value,body:capture.text};
  } finally {
    await page.evaluate(()=>{
      const state=window.__relayStreamCapture;
      if(window.fetch===state.fetch)window.fetch=state.originalFetch;
      delete window.__relayStreamCapture;
    });
  }
}
