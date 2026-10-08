import assert from 'node:assert/strict';
import {test} from 'node:test';
import {summaryCell,summaryMoney,summaryFailureFacts} from './console-summary.mjs';

test('summary money preserves exact decimal precision and distinguishes unknown from zero',()=>{
  assert.equal(summaryMoney(null),'未知');assert.equal(summaryMoney('0'),'0.000000 USD');
  assert.equal(summaryMoney('900719925474099312'),'900719925474.099312 USD');assert.equal(summaryMoney('-21'),'-0.000021 USD');
});
test('live summary assertion rejects absent columns, wrong values and ambiguous rows',async()=>{
  const row=(count,headers,text)=>({count:async()=>count,locator:()=>({locator:()=>({allTextContents:async()=>headers})}),getByRole:()=>({nth:()=>({textContent:async()=>text})})});
  await summaryCell(row(1,['price'],' 0.000021 USD '),'price','0.000021 USD');
  await assert.rejects(summaryCell(row(2,['price'],'21'),'price','21'));
  await assert.rejects(summaryCell(row(1,['other'],'21'),'price','21'));
  await assert.rejects(summaryCell(row(1,['price'],'未知'),'price','21'));
});

test('summary diagnostics keep original assertion and distinguish its exact boundary without raw values',async()=>{
  const row=(count,headers,text)=>({count:async()=>count,locator:()=>({locator:()=>({allTextContents:async()=>headers})}),getByRole:()=>({nth:()=>({textContent:async()=>text})})});
  for(const [target,label,expected,facts] of [
    [row(0,['状态'],'启用'),'状态','停用',{available:true,check:'row-count',column:'model-status',row_count:0}],
    [row(2,['状态'],'启用'),'状态','停用',{available:true,check:'row-count',column:'model-status',row_count:2}],
    [row(1,['private-heading'],'private-value'),'状态','停用',{available:true,check:'column-present',column:'model-status',row_count:1,column_count:1,column_present:false}],
    [row(1,['状态'],'启用'),'状态','停用',{available:true,check:'cell-value',column:'model-status',row_count:1,column_count:1,column_present:true,value_matches:false,expected_status:'停用',observed_status:'启用'}],
    [row(1,['状态'],'private-value'),'状态','private-expected',{available:true,check:'cell-value',column:'model-status',row_count:1,column_count:1,column_present:true,value_matches:false,expected_status:'other',observed_status:'other'}],
    [row(1,['private-label'],'private-value'),'private-label','private-expected',{available:true,check:'cell-value',column:'other',row_count:1,column_count:1,column_present:true,value_matches:false}]
  ]) {
    await assert.rejects(summaryCell(target,label,expected),error=>{
      assert.equal(error.name,'AssertionError');assert.deepEqual(summaryFailureFacts(error),facts);
      assert.ok(!JSON.stringify(summaryFailureFacts(error)).includes('private-'));return true;
    });
  }
  assert.deepEqual(summaryFailureFacts(new Error('private-error')),{available:false});
});

test('summary locator failures retain original error identity and bounded diagnostic',async()=>{
  const failure=new Error('private-transport');
  await assert.rejects(summaryCell({count:async()=>{throw failure;}},'状态','停用'),error=>{
    assert.equal(error,failure);assert.deepEqual(summaryFailureFacts(error),{available:true,check:'row-count',column:'model-status'});return true;
  });
});
