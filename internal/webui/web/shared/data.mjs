export const list = value => Array.isArray(value) ? value : value?.items ?? value?.data ?? [];
export async function collection(api,path) {
  const values=[];let page=1,cursor='';
  while(true){const separator=path.includes('?')?'&':'?',response=await api(path+separator+new URLSearchParams({page:String(page),page_size:'100',limit:'100',...(cursor?{cursor}:{})}));const items=list(response);values.push(...items);if(values.length>10000)throw new Error('列表超过10000项，请缩小范围后读取');if(response.next_cursor){cursor=response.next_cursor;continue;}if(response.total!=null&&values.length<response.total){page++;continue;}return values;}
}
