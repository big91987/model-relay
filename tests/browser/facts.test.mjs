import assert from 'node:assert/strict';
import test from 'node:test';
import {createFactSummary} from '../../internal/webui/web/shared/facts.mjs';

const node=(tag,text='',className)=>({tag,textContent:text,className,children:[],attributes:{},append(...values){this.children.push(...values);},setAttribute(name,value){this.attributes[name]=value;}});
const {factSummary,column}=createFactSummary({node});

test('shared fact presentation preserves nested labels, unknown values and unformatted amounts',()=>{
  const rendered=factSummary({id:'',enabled:false,attempts:[{ordinal:1,state:'success',fee:'000007'}],check:{state:'success',execution_revision:2},unknown_field:null});
  assert.equal(rendered.tag,'dl');
  assert.deepEqual(rendered.children.filter(value=>value.tag==='dt').map(value=>value.textContent),['安全标识','启用','调用尝试','当前配置核验','unknown_field']);
  const values=rendered.children.filter(value=>value.tag==='dd').map(value=>value.children[0]);
  assert.equal(values[0].textContent,'—');assert.equal(values[0].tabIndex,undefined);assert.equal(values[1].textContent,'否');
  assert.equal(values[2].tag,'ol');assert.equal(values[2].children[0].tag,'li');
  const attempt=values[2].children[0].children[0].children.filter(value=>value.tag==='dd').map(value=>value.children[0].textContent);
  assert.deepEqual(attempt,['1','成功','000007']);
  const check=values[3].children.filter(value=>value.tag==='dd').map(value=>value.children[0].textContent);
  assert.deepEqual(check,['核验通过','2']);assert.equal(values[4].textContent,'未知');
  const empty=factSummary([]);assert.equal(empty.tag,'ol');assert.equal(empty.children[0].textContent,'暂无记录');
});

test('default fact columns preserve identifier accessibility and explicit format overrides',()=>{
  const identifier='request_identifier_<literal>';
  const rendered=column('标识','id').render({id:identifier});
  assert.equal(rendered.className,'technical-value');assert.equal(rendered.textContent,identifier);assert.equal(rendered.tabIndex,0);assert.equal(rendered.attributes['aria-label'],identifier);
  assert.equal(column('角色','role').render({role:'developer'}),'开发者');
  assert.equal(column('费率','tenant_input_rate').render({tenant_input_rate:'000007'}),'000007');
  const custom=node('strong','explicit');assert.equal(column('标识','id',()=>custom).render({id:identifier}),custom);
});
