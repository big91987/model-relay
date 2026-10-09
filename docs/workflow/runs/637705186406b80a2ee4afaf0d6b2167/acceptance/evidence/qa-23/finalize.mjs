import assert from 'node:assert/strict';
import {readFileSync, writeFileSync, readdirSync, mkdirSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';

const root = 'docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167';
const directory = `${root}/acceptance/evidence/qa-23`;
const digest = bytes => createHash('sha256').update(bytes).digest('hex');
const save = (name,value) => writeFileSync(`${directory}/${name}`,JSON.stringify(value,null,2)+'\n');
const source = JSON.parse(execFileSync('go',['run','./tests/evidence'],{encoding:'utf8',env:{...process.env,GOCACHE:`${process.cwd()}/.cache/go-build`,GOTOOLCHAIN:'local'}}));
assert.deepEqual(source,JSON.parse(readFileSync(`${directory}/source-before.json`)));
save('source-after.json',source);
let goLog = readFileSync(`${directory}/independent-go.log`,'utf8');
if (/\/(?:private\/)?var\/folders\//.test(goLog)) {
  const original = Buffer.from(goLog);
  mkdirSync('.cache/qa23', {recursive:true});
  writeFileSync('.cache/qa23/independent-go-raw.log', original);
  goLog = goLog.replace(/\/(?:private\/)?var\/folders\/[^\s"']+/g,'<TMP_LINKER_PATH>');
  writeFileSync(`${directory}/independent-go.log`,goLog);
  save('local-log-transforms.json',{scope:'QA independent Go log only; host log unmodified',original_sha256:digest(original),original_bytes:original.length,public_sha256:digest(Buffer.from(goLog)),public_bytes:Buffer.byteLength(goLog),replacement:'one linker warning temporary absolute path -> <TMP_LINKER_PATH>',raw_cache_ignored:true,tests_and_numeric_lines_unchanged:true});
}
const nodeLog = readFileSync(`${directory}/independent-node.log`,'utf8');
const topPass = goLog.split('\n').filter(line=>/^--- PASS: TestMetering/.test(line));
assert.equal(topPass.length,8);
assert.match(goLog,/\nPASS\nok\s+model-relay\/internal\/relay\s+21\.991s/);
assert.match(nodeLog,/ℹ tests 53/);
assert.match(nodeLog,/ℹ fail 0/);
assert.match(nodeLog,/ℹ skipped 0/);
const hostLog = readFileSync(`${directory}/host-tests22/verify.log`,'utf8');
const snapshots = hostLog.split('\n').filter(line=>line.includes('METERING_SNAPSHOT ')).map(line=>JSON.parse(line.split('METERING_SNAPSHOT ')[1]));
const expected = {
  bounded_complete:{calls:6,rpm:6,tpm:42,concurrency:0,fee_micro_usd:'42',reserved_micro_usd:'0',pending:0},
  rpm_recovered:{calls:3,rpm:1,tpm:7,concurrency:0,fee_micro_usd:'21',reserved_micro_usd:'0',pending:0},
  unknown_carry:{calls:2,rpm:2,tpm:207,concurrency:0,fee_micro_usd:'7',reserved_micro_usd:'200',pending:1},
  known_with_pending:{calls:3,rpm:3,tpm:214,concurrency:0,fee_micro_usd:'14',reserved_micro_usd:'200',pending:1},
  pending_across_minute:{calls:3,rpm:0,tpm:200,concurrency:0,fee_micro_usd:'14',reserved_micro_usd:'200',pending:1},
  reconciled_old_minute:{calls:4,rpm:1,tpm:7,concurrency:0,fee_micro_usd:'28',reserved_micro_usd:'0',pending:0},
  undispatched_settled:{calls:1,rpm:1,tpm:0,concurrency:0,fee_micro_usd:'0',reserved_micro_usd:'0',pending:0}
};
const counts = {};
for(const [phase,usage] of Object.entries(expected)) {
  const selected = snapshots.filter(value=>value.phase===phase);
  assert.ok(selected.length>=2,phase);
  for(const value of selected) assert.deepEqual(value.usage,usage,phase);
  counts[phase]=selected.length;
}
const pngChecks=[];
for(const name of readdirSync(`${directory}/host-tests22/browser`)) {
  if(!name.endsWith('.png.integrity.json')) continue;
  const value=JSON.parse(readFileSync(`${directory}/host-tests22/browser/${name}`));
  assert.equal(value.result,'passed',name);
  assert.equal(value.sha256,digest(readFileSync(`${directory}/host-tests22/browser/${name.replace('.integrity.json','')}`)),name);
  pngChecks.push(name);
}
save('independent-checks.json',{qa_execution:23,qa_date:'2026-10-10',source_unchanged:true,go:{command:"go test -race -count=1 -v ./internal/relay -parallel=4 -run '^TestMetering(SafetyInProcess|TerminalInProcess|WindowsInProcess|CompetitionInProcess|BoundViolationInProcess|UndispatchedInProcess|CancellationInProcess|TimeoutInProcess)$'",exit:0,top_level_pass:8,elapsed_seconds:21.991,real_http:false},node:{command:'node --test tests/browser/{capture-integrity,readability,console-summary,cooldown-observation}.test.mjs',exit:0,tests:53,skips:0,browser:false},host_numeric_reconciliation:{phases:counts,all_snapshots:snapshots.length},png_integrity_hash_matches:pngChecks.length,qa_browser_reexecution:'Not Run; current same-source tests22 host UI evidence reviewed',supplier_requests:0});
const records=[];
const walk=path=>{
  for(const entry of readdirSync(path,{withFileTypes:true})) {
    const file=`${path}/${entry.name}`;
    if(entry.isDirectory()) walk(file);
    else if(file!==`${directory}/archive-index.json`) {
      const bytes=readFileSync(file);
      if(!file.endsWith('.png')) assert.equal(/\/Users\/|\/private\/var\/|\/var\/folders\/|Bearer\s+\S+|\bsk-[A-Za-z0-9_-]{12,}/.test(bytes.toString()),false,file);
      records.push({path:file,bytes:bytes.length,sha256:digest(bytes)});
    }
  }
};
walk(directory);
for(const name of ['release-assessment','acceptance-matrix','test-plan-review','core-journeys','findings-and-next-actions']) {
  const path=`${root}/acceptance/${name}-qa-23.md`,bytes=readFileSync(path);
  records.push({path,bytes:bytes.length,sha256:digest(bytes)});
}
save('archive-index.json',{scope:'QA23 evidence and report assets; excludes self hash',records});
console.log(JSON.stringify({source,independent_go_pass:8,node_pass:53,snapshot_phase_counts:counts,png_integrity_matches:pngChecks.length,archive_files:records.length}));
