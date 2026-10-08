import assert from 'node:assert/strict';
import {test} from 'node:test';
import {summaryCell,summaryMoney} from './console-summary.mjs';

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
