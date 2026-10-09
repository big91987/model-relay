import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';
import {previousEditorAsset} from './console-faults.mjs';

const source=readFileSync('internal/relay/web/workspace.js','utf8');
const handler=source.slice(source.indexOf('function edit('),source.indexOf('function secretDialog('));
function element() {
  return {
    children:[],open:false,hidden:false,disabled:false,dataset:{},
    get textContent(){return (this.text??'')+this.children.map(child=>child.textContent).join('');},
    set textContent(value){this.text=String(value);this.children=[];},
    append(...values){this.children.push(...values);},replaceChildren(...values){this.children=[...values];},
    querySelector(){return null;},showModal(){this.open=true;},close(){this.open=false;}
  };
}
function fixture(read,editorHandler=handler) {
  const elements=new Map(['editor','editor-fields','editor-title','editor-error','editor-save','editor-form'].map(id=>[id,element()]));
  const counts={mutations:0,reads:0,details:0},state={epoch:1,busy:false};
  const context=vm.createContext({state,crypto:{randomUUID:()=> 'controlled-operation'},byId:id=>elements.get(id),clearSensitive(){},openModal:dialog=>dialog.showModal(),closeModal:dialog=>dialog.close(),
    node(tag,text=''){const value=element();value.textContent=text;return value;},
    button(label,onclick){const value=element();value.textContent=label;value.onclick=onclick;return value;},
    api:async path=>{counts.reads++;assert.equal(path,'operations/controlled-operation');return read(counts.reads);},
    detail(){counts.details++;},submit:async()=>{counts.mutations++;throw new Error('controlled lost response');}});
  vm.runInContext(editorHandler+"\nedit('controlled rotation',[],submit);",context);
  return {elements,counts,state,submit:()=>elements.get('editor-form').onsubmit({preventDefault(){}}),retry:()=>elements.get('editor-error').children.find(child=>child.textContent==='安全读回')};
}
test('unknown mutation retains operation reference and retry through read failure and recovery',async()=>{
  const value=fixture(async count=>{if(count===1)throw new Error('transient');return {action:'key.rotations'};});
  await value.submit();await value.retry().onclick();
  assert.match(value.elements.get('editor-error').textContent,/读回失败，结果仍未知/);
  assert.match(value.elements.get('editor-error').textContent,/controlled-operation/);
  assert.equal(value.elements.get('editor-save').hidden,true);
  assert.ok(value.retry());await value.submit();await value.retry().onclick();
  assert.deepEqual(value.counts,{mutations:1,reads:2,details:1});
});
test('safe reads deduplicate concurrent clicks and ignore scope-late results',async()=>{
  let finish;
  const value=fixture(()=>new Promise(resolve=>{finish=resolve;}));await value.submit();
  const first=value.retry().onclick();await value.retry().onclick();
  assert.equal(value.counts.reads,1);value.state.epoch++;finish({action:'key.rotations'});await first;
  assert.equal(value.counts.details,0);
});
test('closing unknown editor isolates its late recovery detail',async()=>{
  let finish;
  const value=fixture(()=>new Promise(resolve=>{finish=resolve;}));await value.submit();
  const pending=value.retry().onclick();value.elements.get('editor').close();finish({action:'key.create'});await pending;
  assert.equal(value.counts.details,0);
});
test('isolated previous-editor reconstruction reproduces the QA35 DOM retry deletion',async()=>{
  const previous=previousEditorAsset(source);
  const value=fixture(()=>{throw new Error('transient');},previous.slice(previous.indexOf('function edit('),previous.indexOf('function secretDialog(')));
  await value.submit();await value.retry().onclick();
  assert.match(value.elements.get('editor-error').textContent,/读回失败，结果仍未知/);
  assert.equal(value.retry(),undefined);
  assert.equal(value.elements.get('editor-error').textContent.includes('controlled-operation'),false);
  assert.equal(value.counts.mutations,1);
});
