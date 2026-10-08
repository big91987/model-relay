import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import {dirname,join} from 'node:path';
import test from 'node:test';

const require=createRequire(import.meta.url);
const {Route}=require(join(dirname(require.resolve('playwright-core/package.json')),'lib/server/network.js'));

function inFlightRoute() {
  const handler=()=>{};
  let delegated=0,removed=0;
  const route=Object.create(Route.prototype);
  Object.assign(route,{
    _handled:false,_futureHandlers:[],_currentHandler:handler,
    _request:{_applyOverrides:overrides=>overrides,_context:{emit(){},removeRouteInFlight(){removed++;}}},
    _delegate:{async continue(){delegated++;await new Promise(resolve=>setImmediate(resolve));}}
  });
  return {route,handler,counts:()=>({delegated,removed})};
}

test('locked Playwright removal continues its current in-flight handler before a late explicit continue',async()=>{
  const {route,handler,counts}=inFlightRoute();
  await route.removeHandler(handler);
  await assert.rejects(route.continue({isFallback:false}),/Route is already handled!/);
  assert.deepEqual(counts(),{delegated:1,removed:1});
});

test('locked Playwright explicit continue settles before removing its handler without a second dispatch',async()=>{
  const {route,handler,counts}=inFlightRoute();
  await route.continue({isFallback:false});
  await route.removeHandler(handler);
  assert.deepEqual(counts(),{delegated:1,removed:1});
});
