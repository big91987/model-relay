import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { execFileSync } from 'node:child_process';

const run = '4cd62ced6a4f98f50e3cca6f7596eb06';
const destination = `docs/workflow/runs/${run}/acceptance/evidence/attempt-65`;
const material = `.workflow-input/${run}/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`;
const digest = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const walk = directory => fs.readdirSync(directory, { withFileTypes: true }).flatMap(entry => entry.isDirectory() ? walk(path.join(directory, entry.name)) : [path.join(directory, entry.name)]);
const sources = ['.gitignore', 'Makefile', 'README.md', 'go.mod', 'go.sum', 'package.json', 'package-lock.json', 'docs/workflow/task.md', 'docs/workflow/requirements.md', 'docs/workflow/design.md', 'docs/product-brief.md', 'docs/roadmap.md', 'docs/releases/v0.1.0.md', 'docs/local-deployment.md'];
const contracts = ['product-definition.md', 'console-architecture.md', 'console-interaction.md', 'routing-architecture.md', 'routing-interaction.md', 'routing-validation.md', 'adr-001-storage-upgrade.md', 'deployment-upgrade-contract.md', 'architecture.md', 'implementation-contract.md', 'migration-and-operations.md', 'verification-design.md', 'acceptance-matrix.md'];
for (const entry of fs.readdirSync('docs/workflow/runs')) {
  for (const contract of contracts) {
    const filename = `docs/workflow/runs/${entry}/${contract}`;
    if (fs.existsSync(filename)) sources.push(filename);
  }
}
for (const directory of ['cmd', 'internal', 'tests', '.trellis/spec']) sources.push(...walk(directory));
sources.sort();
const sourceHash = crypto.createHash('sha256');
const sourceFiles = sources.map(filename => {
  const bytes = fs.readFileSync(filename);
  sourceHash.update(`${filename}\0${bytes.length}\0`);
  sourceHash.update(bytes);
  return { path: filename, bytes: bytes.length, sha256: digest(bytes) };
});
const manifest = JSON.parse(fs.readFileSync(`${material}/files/manifest.json`));
const materialFiles = manifest.files.map(entry => {
  const filename = `${material}/files/${entry.path}`;
  const bytes = fs.readFileSync(filename);
  return { path: filename, sha256: digest(bytes), expected_sha256: entry.sha256, matches: digest(bytes) === entry.sha256 && bytes.length === entry.bytes };
});
const archiveHash = digest(fs.readFileSync(`${material}/archive.zip`));
const fingerprint = sourceHash.digest('hex');
const receipt = {
  checked_at: new Date().toISOString(), method: 'QA independent Node byte inventory; not Go evidence command',
  head: execFileSync('git', ['rev-parse', 'HEAD'], { encoding: 'utf8' }).trim(),
  branch: execFileSync('git', ['branch', '--show-current'], { encoding: 'utf8' }).trim(),
  files: sources.length, working_tree_sha256: fingerprint,
  matches_seq64: fingerprint === 'ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a',
  material_version: manifest.version, material_status: manifest.status, archive_sha256: archiveHash,
  material_all_match: archiveHash === '5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7' && materialFiles.every(entry => entry.matches),
  source_files: sourceFiles, material_files: materialFiles
};
fs.writeFileSync(`${destination}/integrity.json`, JSON.stringify(receipt, null, 2) + '\n');
console.log(JSON.stringify({ files: receipt.files, head: receipt.head, working_tree_sha256: fingerprint, matches_seq64: receipt.matches_seq64, material_all_match: receipt.material_all_match }));
if (!receipt.matches_seq64 || !receipt.material_all_match) process.exitCode = 1;
