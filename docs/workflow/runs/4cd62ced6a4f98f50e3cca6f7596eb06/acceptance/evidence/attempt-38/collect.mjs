import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFile,writeFile,copyFile,mkdir,readdir,stat} from 'node:fs/promises';

const run='4cd62ced6a4f98f50e3cca6f7596eb06';
const archiveSHA='5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7';
const material=`.workflow-input/${run}/${archiveSHA}`;
const destination=`docs/workflow/runs/${run}/acceptance/evidence/attempt-38`;
const digest=bytes=>createHash('sha256').update(bytes).digest('hex');
const manifest=JSON.parse(await readFile(`${material}/files/manifest.json`,'utf8'));
assert.equal(manifest.version,'design-v0.2.0');
assert.equal(digest(await readFile(`${material}/archive.zip`)),archiveSHA);
const materials=[];
for(const entry of manifest.files){
  const path=`${material}/files/${entry.path}`,bytes=await readFile(path);
  assert.equal(digest(bytes),entry.sha256);assert.equal(bytes.length,entry.bytes);
  materials.push({path,sha256:entry.sha256,bytes:bytes.length});
}
await writeFile(`${destination}/material-integrity.json`,JSON.stringify({version:manifest.version,status:manifest.status,manifest_path:`${material}/files/manifest.json`,archive_sha256:archiveSHA,files:materials},null,2)+'\n');
const artifacts=[];
await mkdir(`${destination}/host-seq37`,{recursive:true});
for(const [width,pid] of [[1280,5283],[1440,5310],[390,5330]]){
  const receipt=JSON.parse(await readFile(`test-results/continuous-${width}-${pid}.json`,'utf8'));
  assert.equal(receipt.width,width);assert.equal(receipt.result,'passed');
  assert.equal(receipt.provider,'controlled_http_not_supplier');
  assert.deepEqual([...receipt.checkpoints].sort(),['J01','J02','J03','J04','J05','J06','J07']);
  const fault=JSON.parse(await readFile(`test-results/fault-${width}.json`,'utf8'));
  assert.equal(fault.width,width);assert.equal(fault.phase,'completed');
  assert.equal(fault.cases.length,5);
  for(const value of fault.cases)assert.equal(value.result,value.case==='QA35-01-historical-red'?'expected_failure_reproduced':'passed');
  const continuous=[`continuous-${width}-${pid}.json`,...['before-rotation','key-states','before-correction','after-correction','final'].map(name=>`continuous-${width}-${name}-${pid}.png`)];
  const faults=(await readdir('test-results')).filter(name=>name.startsWith(`fault-${width}-`)&&name.endsWith('.png'));
  const names=[...continuous,`fault-${width}.json`,...faults,`console-${width}.png`];
  for(const name of names){
    const source=`test-results/${name}`,path=`${destination}/host-seq37/${name}`,bytes=await readFile(source);
    const metadata=await stat(source);await copyFile(source,path);
    assert.equal(digest(await readFile(path)),digest(bytes));
    artifacts.push({source,path,sha256:digest(bytes),bytes:bytes.length,source_modified_at:metadata.mtime.toISOString(),provenance:'host-produced artifact copied unchanged by QA; not QA browser execution'});
  }
}
const source=JSON.parse(await readFile(`${destination}/source-before.json`,'utf8'));
assert.equal(source.head,'98511771871cf0951ecef716bbb55deb13e18da5');
assert.equal(source.files,127);
assert.equal(source.working_tree_sha256,'3ee2122eadac33a67342523a082f791d8a81154cc86b7321d53de856866a15ef');
await writeFile(`${destination}/artifact-index.json`,JSON.stringify({recorded_at:new Date().toISOString(),seq:37,connector_exit_code:0,connector_log_bytes:101748,connector_log_truncated:false,source,artifacts,limitations:['JSON receipts do not embed a signed source digest; linkage is connector source stamp, matching checkout, reviewed runner and artifact timestamps.','Raw full command archive retained by platform read_command_output(seq=37); this index is not a copy of that log.','Screenshot safety supported by secret-empty assertions and inspected samples; no claim of automated pixel OCR.']},null,2)+'\n');
console.log(JSON.stringify({material_files:materials.length,archive_matches:true,artifact_files:artifacts.length,viewports:3,continuous_checkpoints:21,fault_cases:15,source_matches:true}));
