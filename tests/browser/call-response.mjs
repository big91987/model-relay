import assert from 'node:assert/strict';

export async function withStreamCapture(page,send) {
  await page.evaluate(()=>{
    if(window.__relayHarnessStreamCapture)throw new Error('stream capture already installed');
    const previous=window.fetch;
    const state={previous,count:0,body:null};
    const wrapped=async function(...args) {
      const response=await previous.apply(this,args);
      const method=args[1]?.method??args[0]?.method??'GET';
      if(new URL(response.url,location.href).pathname==='/v1/chat/completions'&&method.toUpperCase()==='POST') {
        state.count++;
        const id=response.headers.get('x-request-id');
        state.body=response.clone().text().then(value=>({ok:true,id,value}),()=>({ok:false,id}));
      }
      return response;
    };
    state.wrapped=wrapped;
    window.__relayHarnessStreamCapture=state;
    window.fetch=wrapped;
  });
  try {
    const response=await send();
    const captured=await page.evaluate(async()=>{
      const state=window.__relayHarnessStreamCapture;
      return {count:state.count,body:await state.body};
    });
    assert.equal(captured.count,1,'stream capture must observe exactly one dispatch');
    assert.equal(captured.body?.ok,true,'stream capture body read must succeed');
    assert.equal(captured.body.id,response.id,'stream capture must match the same request ID');
    return {...response,value:captured.body.value};
  } finally {
    await page.evaluate(()=>{
      const state=window.__relayHarnessStreamCapture;
      if(state){if(window.fetch===state.wrapped)window.fetch=state.previous;delete window.__relayHarnessStreamCapture;}
    });
  }
}
