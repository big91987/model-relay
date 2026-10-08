import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {copyFileSync,existsSync,mkdirSync,readFileSync,readdirSync,statSync,writeFileSync} from 'node:fs';
import path from 'node:path';

const run='4cd62ced6a4f98f50e3cca6f7596eb06';
const destination=`docs/workflow/runs/${run}/acceptance/evidence/attempt-58`;
const digest=bytes=>createHash('sha256').update(bytes).digest('hex');
const walk=root=>readdirSync(root,{withFileTypes:true}).flatMap(entry=>entry.isDirectory()?walk(path.join(root,entry.name)):[path.join(root,entry.name)]);
const productPaths=['.gitignore','Makefile','README.md','go.mod','go.sum','package.json','package-lock.json','docs/workflow/task.md','docs/workflow/requirements.md','docs/workflow/design.md','docs/product-brief.md','docs/roadmap.md','docs/releases/v0.1.0.md','docs/local-deployment.md'];
const contracts=['product-definition.md','console-architecture.md','console-interaction.md','routing-architecture.md','routing-interaction.md','routing-validation.md','adr-001-storage-upgrade.md','deployment-upgrade-contract.md','architecture.md','implementation-contract.md','migration-and-operations.md','verification-design.md','acceptance-matrix.md'];
for(const entry of readdirSync('docs/workflow/runs',{withFileTypes:true}).filter(entry=>entry.isDirectory()))for(const contract of contracts){const filename=`docs/workflow/runs/${entry.name}/${contract}`;if(existsSync(filename))productPaths.push(filename);}
for(const root of ['cmd','internal','tests','.trellis/spec'])productPaths.push(...walk(root));
productPaths.sort();
const sourceHash=createHash('sha256');
const sourceFiles=productPaths.map(filename=>{const bytes=readFileSync(filename);sourceHash.update(`${filename}\0${bytes.length}\0`);sourceHash.update(bytes);return {path:filename,bytes:bytes.length,sha256:digest(bytes)};});
const fingerprint=sourceHash.digest('hex');
assert.equal(fingerprint,'fb4066fa37d1e7865d3a18f72dd78fef98ae79bde17e20875af373b24771c3a6');
assert.equal(sourceFiles.length,134);
const head=execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim();
assert.equal(head,'98511771871cf0951ecef716bbb55deb13e18da5');
const archiveSHA='5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7';
const materialBase=`.workflow-input/${run}/${archiveSHA}`;
const manifestPath=`${materialBase}/files/manifest.json`;
const manifest=JSON.parse(readFileSync(manifestPath));
assert.equal(manifest.version,'design-v0.2.0');
assert.equal(digest(readFileSync(`${materialBase}/archive.zip`)),archiveSHA);
const materials=manifest.files.map(entry=>{const filename=`${materialBase}/files/${entry.path}`,bytes=readFileSync(filename);assert.equal(bytes.length,entry.bytes);assert.equal(digest(bytes),entry.sha256);return {...entry,path:filename,match:true};});
const receiptChecks=[];
for(const width of [1280,1440,390]){
  const review=JSON.parse(readFileSync(`test-results/visual-${width}/review.json`));
  assert.equal(review.result,'passed');assert.equal(review.phase,'completed');assert.equal(review.pages.length,10);assert.equal(review.role,'platform_administrator');assert.equal(review.provider,'controlled_http_not_supplier');
  const screenshots=new Map();
  const inspect=value=>{if(value&&typeof value==='object'){if(typeof value.path==='string'&&typeof value.sha256==='string'){assert.equal(digest(readFileSync(value.path)),value.sha256);screenshots.set(value.path,value.sha256);}for(const child of Object.values(value))inspect(child);}};
  inspect(review);assert.equal(screenshots.size,21);
  for(const page of review.pages){assert.equal(page.facts.role,true);assert.equal(page.facts.scope,true);assert.equal(page.facts.rootFits,true);assert.equal(page.facts.secretsClear,true);assert.equal(page.facts.settlement.loading,false);assert.equal(page.facts.settlement.error_nodes,0);}
  const summary=JSON.parse(readFileSync(`test-results/management-summary-${width}.json`));
  assert.equal(summary.result,'passed');assert.deepEqual(summary.facts,{models:2,tenants:2,services:1,keys:5,api_values_match:true,context_actions:true});
  const execution={1280:68740,1440:68790,390:68850}[width];
  const journey=JSON.parse(readFileSync(`test-results/continuous-${width}-${execution}.json`));
  assert.equal(journey.result,'passed');assert.deepEqual([...journey.checkpoints].sort(),['J01','J02','J03','J04','J05','J06','J07']);
  const fault=JSON.parse(readFileSync(`test-results/fault-${width}.json`));
  for(const entry of fault.cases)assert.equal(entry.result,entry.case==='QA35-01-historical-red'?'expected_failure_reproduced':'passed');
  const editor=JSON.parse(readFileSync(`test-results/editor-interruption-${width}.json`));
  assert.equal(editor.result,'passed');assert.equal(editor.service_posts,1);assert.equal(editor.canceled_editor_posts,0);
  receiptChecks.push({width,pages:10,receipt_png_hashes_matched:21,role_scope_layout_secrets_settlement:true,journeys:journey.checkpoints,management_summary:summary.facts,editor_and_faults:'passed'});
}
const evidencePaths=[];
for(const width of [1280,1440,390]){
  evidencePaths.push(...walk(`test-results/visual-${width}`));
  for(const prefix of ['management-summary','editor-interruption','fault'])evidencePaths.push(`test-results/${prefix}-${width}.json`);
  const execution={1280:68740,1440:68790,390:68850}[width];
  evidencePaths.push(`test-results/continuous-${width}-${execution}.json`);
  for(const suffix of ['before-correction','after-correction'])evidencePaths.push(`test-results/continuous-${width}-${suffix}-${execution}.png`);
}
const evidence=evidencePaths.map(filename=>{const bytes=readFileSync(filename),retained=`${destination}/host/${filename.slice('test-results/'.length)}`;mkdirSync(path.dirname(retained),{recursive:true});copyFileSync(filename,retained);return {source:filename,path:retained,bytes:bytes.length,sha256:digest(bytes),source_modified_at:statSync(filename).mtime.toISOString()};});
const result={run,node:'qa',execution:58,candidate:'records-detail-entry-r1',time:new Date().toISOString(),head,branch:execFileSync('git',['branch','--show-current'],{encoding:'utf8'}).trim(),working_tree_sha256:fingerprint,files:134,sourceFiles,material:{version:manifest.version,status:manifest.status,manifest_path:manifestPath,manifest_sha256:digest(readFileSync(manifestPath)),archive_path:`${materialBase}/archive.zip`,archive_sha256:archiveSHA,files:materials},evidence,receiptChecks,host:{seq:57,command:'make verify',exit_code:0,bytes:106669,truncated:false,read_offsets:[0,32768,65536,98304],additional_review_offsets:[42000,51000],next_offset:106669,eof:true},local:{node:process.version,go:execFileSync('go',['version'],{encoding:'utf8'}).trim(),playwright:JSON.parse(readFileSync('package-lock.json')).packages['node_modules/playwright'].version}};
writeFileSync(`${destination}/integrity.json`,JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({fingerprint,files:134,material_files:materials.length,evidence_files:evidence.length,head,local:result.local}));
