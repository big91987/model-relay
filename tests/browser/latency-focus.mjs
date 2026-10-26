import {chromium} from 'playwright';
import assert from 'node:assert/strict';
import {latencyJourney} from './latency-journey.mjs';
const browser=await chromium.launch({headless:true});
try {
 const width=Number(process.env.RELAY_LATENCY_WIDTH),origin=process.env.RELAY_LATENCY_ORIGIN;
 const page=await browser.newPage({viewport:{width,height:900}});await page.goto(origin+'/admin/');await page.getByLabel('账号',{exact:true}).fill('admin');await page.getByLabel('密码',{exact:true}).fill(process.env.RELAY_LATENCY_PASSWORD);await page.getByRole('button',{name:'登录',exact:true}).click();await page.locator('#workspace').waitFor({state:'visible'});
 await latencyJourney(page,{width,tenant:process.env.RELAY_LATENCY_TENANT,upstreams:[process.env.RELAY_LATENCY_UPSTREAM+'/1/v1',process.env.RELAY_LATENCY_UPSTREAM+'/2/v1'],credential:process.env.RELAY_LATENCY_CREDENTIAL,password:process.env.RELAY_LATENCY_PASSWORD,restart:async()=>{const response=await fetch(process.env.RELAY_LATENCY_CONTROL,{method:'POST',headers:{'X-Test-Capability':process.env.RELAY_LATENCY_CAPABILITY}});assert.equal(response.status,200);return (await response.json()).origin;}});
}finally{await browser.close();}
