import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFile, writeFile, copyFile, mkdir, stat} from 'node:fs/promises';

const run='4cd62ced6a4f98f50e3cca6f7596eb06';
const archiveDigest='5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7';
const material=`.workflow-input/${run}/${archiveDigest}`;
const destination=`docs/workflow/runs/${run}/acceptance/evidence/attempt-35`;
const digest=bytes=>createHash('sha256').update(bytes).digest('hex');
const manifest=JSON.parse(await readFile(`${material}/files/manifest.json`,'utf8'));
assert.equal(manifest.version,'design-v0.2.0');
assert.equal(digest(await readFile(`${material}/archive.zip`)),archiveDigest);
const materials=[];
for(const entry of manifest.files){
  const path=`${material}/files/${entry.path}`;
  const bytes=await readFile(path);
  assert.equal(digest(bytes),entry.sha256);
  assert.equal(bytes.length,entry.bytes);
  materials.push({path,sha256:entry.sha256,bytes:bytes.length});
}
await writeFile(`${destination}/material-integrity.json`,JSON.stringify({version:manifest.version,status:manifest.status,manifest_path:`${material}/files/manifest.json`,archive_path:`${material}/archive.zip`,archive_sha256:archiveDigest,files:materials},null,2)+'\n');
const artifacts=[];
await mkdir(`${destination}/host-seq34`,{recursive:true});
for(const [width,pid] of [[1280,96921],[1440,96937],[390,96963]]){
  const receiptSource=`test-results/continuous-${width}-${pid}.json`;
  const receipt=JSON.parse(await readFile(receiptSource,'utf8'));
  assert.equal(receipt.width,width);
  assert.equal(receipt.result,'passed');
  assert.equal(receipt.provider,'controlled_http_not_supplier');
  assert.deepEqual([...receipt.checkpoints].sort(),['J01','J02','J03','J04','J05','J06','J07']);
  for(const suffix of ['.json',...['before-rotation','key-states','before-correction','after-correction','final'].map(name=>`-${name}.png`)]){
    const filename=suffix==='.json'?`continuous-${width}-${pid}.json`:`continuous-${width}${suffix.slice(0,-4)}-${pid}.png`;
    const source=`test-results/${filename}`;
    const path=`${destination}/host-seq34/${filename}`;
    const bytes=await readFile(source);
    const metadata=await stat(source);
    await copyFile(source,path);
    assert.equal(digest(await readFile(path)),digest(bytes));
    artifacts.push({source,path,sha256:digest(bytes),bytes:bytes.length,source_modified_at:metadata.mtime.toISOString(),provenance:'host artifact copied by QA; seq34 checkpoints independently read; not QA browser execution'});
  }
}
const sources=[];
for(const path of ['Makefile','tests/evidence/main.go','tests/browser/journeys.go','tests/browser/journeys.mjs','tests/browser/console.mjs','tests/browser/workspace.test.mjs','internal/relay/web/workspace.js','internal/relay/console_contract_test.go','internal/relay/console_http_test.go']){
  sources.push({path,sha256:digest(await readFile(path))});
}
await writeFile(`${destination}/artifact-index.json`,JSON.stringify({recorded_at:new Date().toISOString(),seq:34,connector_exit_code:0,connector_log_bytes:94335,connector_log_truncated:false,receipt_limitation:'success JSON has no source digest; association relies on connector source stamp, this checkout stamp, artifact timestamp window and unchanged reviewed runner, not self-contained signed artifact attestation',artifacts,reviewed_sources:sources},null,2)+'\n');
console.log(JSON.stringify({material_files:materials.length,archive_matches:true,artifact_files:artifacts.length,viewports:3,checkpoints:21,source_review_files:sources.length}));
