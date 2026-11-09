import {collection as readCollection} from './data.mjs';
// Every render and editor read belongs to one active feature lifetime.
export function createLifetime(request) {
  let revision=0,active=true,element=null;
  const assertCurrent=version=>{if(!active||version!==revision)throw new DOMException('feature disposed','AbortError');};
  async function api(...args){const version=revision;assertCurrent(version);try{const value=await request(...args);assertCurrent(version);return value;}catch(error){assertCurrent(version);throw error;}}
  return {api,collection:path=>readCollection(api,path),activate(){active=true;revision++;},mount(value){element=value;return value;},dispose(){active=false;revision++;element?.replaceChildren();element=null;}};
}
