// This feature owns its DOM, app-key request, stream and disposal lifecycle.
export function createPlayground({document,notify,openRecord,fetch=(...args)=>globalThis.fetch(...args),randomUUID=()=>crypto.randomUUID()}) {
  let active=true,controller=null;
  const node=(tag,text='')=>{const element=document.createElement(tag);element.textContent=text;return element;};
  const button=(label,action)=>{const element=node('button',label);element.type='button';element.onclick=async event=>{try{await action(event);}catch(error){if(active&&error.name!=='AbortError')notify('操作失败：'+error.message);}};return element;};
  const form=node('form');form.className='card';form.id='call-form';
  function field(id,label,type){
    const input=node('input');input.id=id;input.type=type;input.required=true;input.autocomplete='off';
    if(type==='password')input.className='secret';
    const heading=node('label',label);heading.htmlFor=id;form.append(heading,input);return input;
  }
  const key=field('call-key','应用API Key（不会保存）','password'),model=field('call-model','公共模型名','text');
  const message=node('textarea');message.id='call-message';message.required=true;message.dataset.sensitive='true';
  const label=node('label','消息');label.htmlFor=message.id;form.append(label,message);
  const streaming=node('input');streaming.type='checkbox';streaming.id='call-stream';
  const streamLabel=node('label','SSE流式');streamLabel.htmlFor=streaming.id;form.append(streamLabel,streaming);
  const send=node('button','发送');send.type='submit';send.className='primary';
  const cancel=button('停止本地接收（费用可能待确认）',()=>controller?.abort());cancel.disabled=true;
  const output=node('pre','尚未调用');output.id='call-output';
  const record=button('查看同Request ID日志',()=>openRecord(record.dataset.id));record.disabled=true;
  const actions=node('div');actions.className='actions';form.append(actions,send,cancel,output,record);
  const controls=[send,key,model,message,streaming];
  form.onsubmit=async event=>{
    event.preventDefault();if(!active||controller)return;
    const current=new AbortController();controller=current;
    for(const control of controls)control.disabled=true;
    cancel.disabled=false;record.disabled=true;output.textContent='正在调用…';let id='';
    try {
      const response=await fetch('/v1/chat/completions',{
        method:'POST',signal:current.signal,
        headers:{'Content-Type':'application/json','Authorization':'Bearer '+key.value,'Idempotency-Key':randomUUID()},
        body:JSON.stringify({model:model.value,messages:[{role:'user',content:message.value}],stream:streaming.checked,...(streaming.checked?{stream_options:{include_usage:true}}:{})})
      });
      id=response.headers.get('X-Request-ID')??'';if(!active)return;
      if(!response.ok){const value=await response.json();throw new Error(value.error?.code??'调用失败');}
      output.textContent='';
      if(!streaming.checked){const value=await response.json();if(active)output.textContent=value.choices?.[0]?.message?.content??'响应中没有文本';}
      else await receiveStream(response.body,output,()=>active);
      if(active)notify('本地接收结束；真实计量与多尝试请查看同ID记录。');
    } catch(error){
      if(active)output.textContent+=(error.name==='AbortError'?'\n已停止本地接收；上游用量可能待确认。':'\n调用失败：'+error.message);
    } finally {
      if(active){controller=null;for(const control of controls)control.disabled=false;cancel.disabled=true;record.dataset.id=id;record.disabled=!id;}
    }
  };
  return {
    element:form,
    dispose(){active=false;controller?.abort();controller=null;key.value='';message.value='';output.textContent='';record.dataset.id='';record.disabled=true;}
  };
}

async function receiveStream(body,output,isActive) {
  const reader=body.getReader(),decoder=new TextDecoder();let buffer='',done=false;
  try {
    while(true){
      const result=await reader.read();if(!isActive())return;if(result.done)break;
      buffer+=decoder.decode(result.value,{stream:true});buffer=buffer.replace(/\r\n/g,'\n');
      if(buffer.length>1048576)throw new Error('流帧超过安全上限');
      let boundary;
      while((boundary=buffer.indexOf('\n\n'))>=0){
        const frame=buffer.slice(0,boundary);buffer=buffer.slice(boundary+2);
        for(const line of frame.split('\n')){
          if(!line.startsWith('data:'))continue;
          const data=line.slice(5).trim();if(data==='[DONE]'){done=true;continue;}
          const value=JSON.parse(data);if(value.error)throw new Error(value.error.code??'流中断');
          const text=value.choices?.[0]?.delta?.content;if(text)output.textContent+=text;
          if(output.textContent.length>1048576)throw new Error('输出超过本地展示上限');
        }
      }
    }
    if(!done)throw new Error('流未正常结束，费用可能待确认');
  } finally {reader.releaseLock();}
}
