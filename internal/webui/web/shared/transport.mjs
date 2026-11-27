// Owns request lifetime; the workspace supplies scope and authorization handling.
export function createTransport({session,onError,fetch=(...args)=>globalThis.fetch(...args)}) {
  const controllers=new Set();
  async function receive(path,method,body,operation,format) {
    const scope=session(),controller=new AbortController();controllers.add(controller);
    try {
      const response=await fetch('/admin/api/'+path,{
        method,credentials:'same-origin',signal:controller.signal,
        headers:{'Content-Type':'application/json','X-CSRF-Token':scope.csrf,...(operation?{'Idempotency-Key':operation}:{})},
        ...(body===undefined?{}:{body:JSON.stringify(body)})
      });
      const value=await response[format]();
      const current=session();
      if(controller.signal.aborted||scope.epoch!==current.epoch||scope.user!==current.user)throw new DOMException('scope changed','AbortError');
      if(!response.ok){
        const error=new Error(format==='json'?(value.error?.code??'读取失败'):'导出失败，请缩小筛选或检查权限');
        error.status=response.status;onError(response.status,path);throw error;
      }
      return value;
    } finally {controllers.delete(controller);}
  }
  return {
    request:(path,method='GET',body,operation)=>receive(path,method,body,operation,'json'),
    download:path=>receive(path,'GET',undefined,undefined,'blob'),
    cancel(){for(const controller of controllers)controller.abort();controllers.clear();}
  };
}
