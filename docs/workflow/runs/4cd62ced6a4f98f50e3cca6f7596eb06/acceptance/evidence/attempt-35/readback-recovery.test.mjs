import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';

const source=readFileSync('internal/relay/web/workspace.js','utf8');
const start=source.indexOf('function edit(');
const end=source.indexOf('function secretDialog(');
assert.ok(start>=0&&end>start);
const handler=source.slice(start,end);

function element() {
  return {
    children:[],value:'',open:false,hidden:false,disabled:false,
    get textContent(){return this.text??'';},
    set textContent(value){this.text=String(value);this.children=[];},
    append(...values){this.children.push(...values);},
    replaceChildren(...values){this.children=[...values];},
    querySelector(){return null;},
    showModal(){this.open=true;},
    close(){this.open=false;}
  };
}

test('AC08 preserves a visible safe-read retry after transient readback failure',async()=>{
  const elements=new Map(['editor','editor-fields','editor-title','editor-error','editor-save','editor-form'].map(id=>[id,element()]));
  const counts={mutations:0,reads:0};
  const context=vm.createContext({
    state:{epoch:1,busy:false},
    crypto:{randomUUID:()=> 'qa-controlled-operation'},
    byId:id=>elements.get(id),
    clearSensitive(){},
    button(label,onclick){const control=element();control.textContent=label;control.onclick=onclick;return control;},
    api:async()=>{counts.reads++;if(counts.reads===1)throw new Error('controlled transient read failure');return {action:'key.rotate'};},
    detail(){},
    submit:async()=>{counts.mutations++;throw new Error('controlled lost mutation response');}
  });
  vm.runInContext(handler+"\nedit('controlled key rotation',[],submit);",context);
  await elements.get('editor-form').onsubmit({preventDefault(){}});
  const output=elements.get('editor-error');
  assert.equal(elements.get('editor-save').hidden,true);
  const firstRetry=output.children.find(control=>control.textContent==='安全读回');
  assert.ok(firstRetry);
  await firstRetry.onclick();
  assert.match(output.textContent,/读回失败，结果仍未知/);
  const retry=output.children.find(control=>control.textContent==='安全读回');
  console.log(JSON.stringify({evidence_type:'Contract',browser_execution:false,real_handler:'internal/relay/web/workspace.js:94',dom_semantics:'textContent replaces descendant controls',mutation_count:counts.mutations,read_count:counts.reads,retry_control_present:!!retry,operation_identifier_visible:output.textContent.includes('qa-controlled-operation'),result:retry?'unexpected_recovery':'failure_reproduced'}));
  assert.ok(retry,'safe read retry must remain reachable after a transient failure');
  await retry.onclick();
  assert.equal(counts.mutations,1);
  assert.equal(counts.reads,2);
});
