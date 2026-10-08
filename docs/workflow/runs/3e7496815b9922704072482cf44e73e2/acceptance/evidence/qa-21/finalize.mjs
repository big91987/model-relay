import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFile,writeFile,copyFile,mkdir} from 'node:fs/promises';
import {execFileSync} from 'node:child_process';

const root='docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/evidence/qa-21';
const hash=data=>createHash('sha256').update(data).digest('hex');
const evidence=JSON.parse(execFileSync('go',['run','./tests/evidence'],{env:{...process.env,GOCACHE:process.cwd()+'/.cache/go-build',GOTOOLCHAIN:'local'}}));
assert.equal(evidence.working_tree_sha256,'f8befd158a543afeb1ba77778e4caf72e876f6e4b0b988dd4e6ab73de5e5eaab');
await writeFile(root+'/source-qa-21.json',JSON.stringify(evidence,null,2)+'\n');
const output=await readFile('test-results/qa21-native-fixed-entry.log','utf8');
const nativeRoot=output.match(/^verification_receipt=(test-results\/verification\/[a-zA-Z0-9-]+)$/m)?.[1];
assert.ok(nativeRoot);
const state=JSON.parse(await readFile(nativeRoot+'/state.json'));
const log=await readFile(nativeRoot+'/verify.log');
assert.equal(log.length,state.log_bytes);assert.equal(hash(log),state.log_sha256);
assert.equal(state.gate_exit,2);assert.equal(state.exit_code,1);
assert.equal(state.source_before.working_tree_sha256,evidence.working_tree_sha256);
assert.equal(state.source_unchanged,true);assert.equal(state.candidate_unchanged,true);
assert.equal(state.owned_group_absent,true);
assert.ok(log.toString().includes('historical listener unavailable; not business red'));
await mkdir(root+'/qa-native',{recursive:true});
await copyFile(nativeRoot+'/verify.log',root+'/qa-native/verify.log');
await copyFile(nativeRoot+'/state.json',root+'/qa-native/state.json');
const checks=[];
for(const name of ['go','node','static']) {
  const raw=await readFile('test-results/qa21-independent-'+name+'.log','utf8');
  assert.ok(raw.includes('QA_'+name.toUpperCase()+'_EXIT=0'));
  const safe=raw.replaceAll(process.cwd(),'<source>').replaceAll(process.env.HOME,'<home>').replace(/(?:\/private)?\/var\/folders\/[^\s'"<>]+/g,'<temp>');
  await writeFile(root+'/independent-'+name+'.log',safe);
  checks.push({kind:name,result:'Pass',file:'independent-'+name+'.log',original_sha256:hash(raw),archived_sha256:hash(safe),bytes:Buffer.byteLength(safe),redaction:'source/home/temporary host paths only'});
}
const files=['Makefile','README.md','package.json','package-lock.json','go.mod','go.sum','tests/verify/main.go','tests/verify/main_test.go','tests/browser/console-summary.mjs','tests/browser/journeys.mjs'];
const locks=[];for(const file of files)locks.push({path:file,sha256:hash(await readFile(file))});
const final={source:evidence,branch:execFileSync('git',['branch','--show-current'],{encoding:'utf8'}).trim(),checks,locks,native:{result:'Blocked',make_exit:2,runner_exit:1,gate_exit:2,receipt:nativeRoot,bytes:log.length,sha256:hash(log),cleanup:state.cleanup,owned_group_absent:state.owned_group_absent,source_unchanged:state.source_unchanged,reason:'historical listener unavailable; not business red',browser:'Blocked; not reached'},collection_note:'First collector incorrectly expected a top-level result on console-faults JSON; corrected to validate five per-case outcomes without changing original receipts or product.'};
await writeFile(root+'/independent-review.json',JSON.stringify(final,null,2)+'\n');
console.log(JSON.stringify(final,null,2));
