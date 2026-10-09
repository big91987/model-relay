import {execFileSync} from 'node:child_process';
import {readFileSync, readdirSync} from 'node:fs';
import {createHash} from 'node:crypto';

const git = (...args) => execFileSync('git', args, {encoding:'utf8', maxBuffer:32*1024*1024}).trim();
const expectedHead = '8917f942cb38af2f56aa93914ab62e763bf9cc6e';
const expectedMain = '6929376089535038112da48aed715fd729615d76';
const head = git('rev-parse', 'HEAD');
const main = git('rev-parse', 'origin/main');
const mergeHead = git('rev-parse', 'MERGE_HEAD');
if (head !== expectedHead || main !== expectedMain || mergeHead !== expectedMain) throw new Error('Unexpected integration identity');
const base = git('merge-base', head, main);
const tree = ref => new Map(git('ls-tree', '-r', '-z', ref).split('\0').filter(Boolean).map(entry => {
  const [metadata, path] = entry.split('\t');
  return [path, metadata.split(' ')[2]];
}));
const mainTree = tree(main);
const headTree = tree(head);
const changes = ref => git('diff', '--name-only', base, ref).split('\n').filter(Boolean);
const inventory = (ref, paths, entries) => paths.map(path => {
  const data = readFileSync(path);
  const blob = createHash('sha1').update(`blob ${data.length}\0`).update(data).digest('hex');
  return {path, bytes:data.length, sha256:createHash('sha256').update(data).digest('hex'), matches_ref:blob === entries.get(path)};
});
const mainInventory = inventory(main, changes(main), mainTree);
const branchInventory = inventory(head, changes(head), headTree);
const unmerged = git('diff', '--name-only', '--diff-filter=U');
const unexpectedMain = mainInventory.filter(entry => !entry.matches_ref);
const allowedStatusDocs = new Set(['README.md', 'tasks/T001-metering.md', 'milestones/M01-metering.md', 'verification/test-plan.md'].map(path => `docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/${path}`));
const unexpectedBranch = branchInventory.filter(entry => !entry.matches_ref && !allowedStatusDocs.has(entry.path));
const materialRoot = '.workflow-input/637705186406b80a2ee4afaf0d6b2167/a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49';
const manifestPath = `${materialRoot}/files/manifest.json`;
const manifest = JSON.parse(readFileSync(manifestPath));
const materialFiles = manifest.files.map(entry => {
  const data = readFileSync(`${materialRoot}/files/${entry.path}`);
  const sha256 = createHash('sha256').update(data).digest('hex');
  return {path:entry.path, bytes:data.length, sha256, matches_manifest:data.length === entry.bytes && sha256 === entry.sha256};
});
const archiveSHA256 = createHash('sha256').update(readFileSync(`${materialRoot}/archive.zip`)).digest('hex');
const sourcePaths = ['.gitignore', 'Makefile', 'README.md', 'go.mod', 'go.sum', 'package.json', 'package-lock.json', 'docs/workflow/task.md', 'docs/workflow/requirements.md', 'docs/workflow/design.md', 'docs/product-brief.md', 'docs/roadmap.md', 'docs/releases/v0.1.0.md', 'docs/local-deployment.md'];
const contractNames = new Set(['product-definition.md', 'console-architecture.md', 'console-interaction.md', 'routing-architecture.md', 'routing-interaction.md', 'routing-validation.md', 'adr-001-storage-upgrade.md', 'deployment-upgrade-contract.md', 'architecture.md', 'implementation-contract.md', 'migration-and-operations.md', 'verification-design.md', 'acceptance-matrix.md']);
for (const run of readdirSync('docs/workflow/runs', {withFileTypes:true}).filter(entry => entry.isDirectory())) {
  for (const entry of readdirSync(`docs/workflow/runs/${run.name}`, {withFileTypes:true})) {
    if (entry.isFile() && contractNames.has(entry.name)) sourcePaths.push(`docs/workflow/runs/${run.name}/${entry.name}`);
  }
}
const walk = root => {
  for (const entry of readdirSync(root, {withFileTypes:true})) {
    const path = `${root}/${entry.name}`;
    if (entry.isDirectory()) walk(path); else sourcePaths.push(path);
  }
};
for (const root of ['cmd', 'internal', 'tests', '.trellis/spec']) walk(root);
const sourceHash = createHash('sha256');
const sourceInventory = sourcePaths.sort().map(path => {
  const data = readFileSync(path);
  sourceHash.update(`${path}\0${data.length}\0`).update(data);
  return {path, bytes:data.length, sha256:createHash('sha256').update(data).digest('hex')};
});
const workingTreeSHA256 = sourceHash.digest('hex');
const valid = !unmerged && !unexpectedMain.length && !unexpectedBranch.length && materialFiles.every(entry => entry.matches_manifest) && archiveSHA256 === 'a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49';
console.log(JSON.stringify({collected_utc:new Date().toISOString(), head, main, merge_head:mergeHead, merge_base:base, integration_pending:true, unmerged_paths:unmerged ? unmerged.split('\n') : [], valid, main_changed_files:mainInventory.length, branch_changed_files:branchInventory.length, unexpected_main:unexpectedMain, unexpected_branch:unexpectedBranch, status_document_changes:branchInventory.filter(entry => !entry.matches_ref).map(entry => entry.path), main_inventory:mainInventory, branch_inventory:branchInventory, source:{files:sourceInventory.length, working_tree_sha256:workingTreeSHA256, inventory:sourceInventory}, material:{version:manifest.version, archive_sha256:archiveSHA256, manifest_path:manifestPath, files:materialFiles}}, null, 2));
if (!valid) process.exitCode = 1;
