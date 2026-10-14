export function installStreamingCapture(target=window) {
  const originalFetch=target.fetch;
  const capture={body:null,restore:()=>{target.fetch=originalFetch;delete target.__relayStreamCapture;}};
  target.__relayStreamCapture=capture;
  target.fetch=async (...args)=>{
    const matches=typeof args[0]==='string'&&new URL(args[0],target.location.href).pathname==='/v1/chat/completions';
    if(matches)target.fetch=originalFetch;
    const response=await originalFetch.apply(target,args);
    if(matches)capture.body=response.clone().text().then(value=>({value}),()=>({failed:true}));
    return response;
  };
}

export async function readStreamingCapture(target=window) {
  const capture=target.__relayStreamCapture;
  if(!capture?.body)throw new Error('stream capture unavailable');
  const result=await capture.body;
  if(result.failed)throw new Error('stream capture failed');
  return result.value;
}
