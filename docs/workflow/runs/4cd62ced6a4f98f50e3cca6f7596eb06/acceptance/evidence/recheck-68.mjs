import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { execFileSync } from 'node:child_process';

const root = 'docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06';
const prior = `${root}/acceptance/evidence/attempt-65`;
const digest = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const readJSON = filename => JSON.parse(fs.readFileSync(filename, 'utf8'));
const baseline = readJSON(`${prior}/integrity.json`);
const sourceHash = crypto.createHash('sha256');
const checkFile = entry => {
  const bytes = fs.readFileSync(entry.path);
  if (!fs.statSync(entry.path).isFile() || bytes.length === 0) throw new Error(`Not a nonempty file: ${entry.path}`);
  const sha256 = digest(bytes);
  if (sha256 !== entry.sha256 || (entry.bytes !== undefined && bytes.length !== entry.bytes)) throw new Error(`Changed file: ${entry.path}`);
  return { path: entry.path, bytes: bytes.length, sha256 };
};
const sourceFiles = baseline.source_files.map(entry => {
  const checked = checkFile(entry);
  sourceHash.update(`${entry.path}\0${checked.bytes}\0`);
  sourceHash.update(fs.readFileSync(entry.path));
  return checked;
});
const fingerprint = sourceHash.digest('hex');
if (fingerprint !== baseline.working_tree_sha256 || sourceFiles.length !== 140) throw new Error('Source inventory mismatch');
const materialRoot = '.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7';
const manifestPath = `${materialRoot}/files/manifest.json`;
const manifest = readJSON(manifestPath);
const materialFiles = manifest.files.map(entry => checkFile({ ...entry, path: `${materialRoot}/files/${entry.path}` }));
const archiveHash = digest(fs.readFileSync(`${materialRoot}/archive.zip`));
if (materialFiles.length !== 19 || manifest.version !== 'design-v0.2.0' || manifest.status !== 'frozen-for-test-implementation' || archiveHash !== baseline.archive_sha256) throw new Error('Material baseline mismatch');
const uiIndex = readJSON(`${prior}/ui-artifact-index.json`);
const uiFiles = uiIndex.retained.map(entry => checkFile({ path: entry.retained, sha256: entry.sha256 }));
if (uiIndex.seq !== 64 || uiFiles.length !== 51) throw new Error('UI inventory mismatch');
const head = execFileSync('git', ['rev-parse', 'HEAD'], { encoding: 'utf8' }).trim();
const staged = execFileSync('git', ['diff', '--cached', '--name-only'], { encoding: 'utf8' }).trim();
if (head !== baseline.head || staged !== '') throw new Error('HEAD or staging changed');
const result = {
  checked_at: new Date().toISOString(), mode: 'read-only evidence revalidation; no product test rerun',
  head, branch: execFileSync('git', ['branch', '--show-current'], { encoding: 'utf8' }).trim(), staged_files: 0,
  source_scope: 'exact QA65 source_files inventory, not every repository file',
  files: sourceFiles.length, working_tree_sha256: fingerprint, source_files: sourceFiles,
  material_version: manifest.version, material_status: manifest.status, material_manifest: manifestPath,
  manifest_sha256: digest(fs.readFileSync(manifestPath)), archive_sha256: archiveHash, material_files: materialFiles,
  ui_seq: uiIndex.seq, ui_files: uiFiles,
  tests_receipt: { seq: 64, exit_code: 0, archive_bytes: 110897, archive_truncated: false, eof: true },
  failed_publish_receipt: { seq: 67, exit_code: 1, archive_bytes: 978, archive_truncated: false, eof: true },
  limitation: 'Connector logs and retained QA65 review remain the execution evidence; this script only checks bytes and paths'
};
fs.writeFileSync(`${root}/acceptance/evidence/recheck-68.json`, JSON.stringify(result, null, 2) + '\n');
const walk = directory => fs.readdirSync(directory, { withFileTypes: true }).flatMap(entry => entry.isDirectory() ? walk(path.join(directory, entry.name)) : [path.join(directory, entry.name)]);
const paths = [
  ...walk(prior),
  `${root}/acceptance/acceptance-report-65.md`, `${root}/acceptance/acceptance-matrix-65.md`,
  `${root}/acceptance/report-handoff-65.md`, `${root}/acceptance/handoff-correction-68.md`,
  `${root}/acceptance/README.md`, `${root}/acceptance/index.md`,
  `${root}/acceptance/evidence/recheck-68.mjs`, `${root}/acceptance/evidence/recheck-68.json`,
  `${root}/delivery/README.md`, `${root}/delivery/tasks/T001-console.md`, `${root}/validation/report.md`,
  `${root}/pr.md`, `${root}/delivery/verification/report-66.md`, `${root}/delivery/verification/history-before-report-66.md`
].sort();
const files = paths.map(filename => {
  if (!filename.startsWith(`${root}/`) || !fs.statSync(filename).isFile()) throw new Error(`Invalid artifact: ${filename}`);
  const bytes = fs.readFileSync(filename);
  if (bytes.length === 0) throw new Error(`Empty artifact: ${filename}`);
  return { path: filename, bytes: bytes.length, sha256: digest(bytes) };
});
fs.writeFileSync(`${root}/acceptance/artifact-files-68.json`, JSON.stringify({
  checked_at: result.checked_at, document_root: root,
  purpose: 'explicit nonempty file inventory; downstream artifacts must use file paths, never directories',
  self_excluded: `${root}/acceptance/artifact-files-68.json`, files
}, null, 2) + '\n');
console.log(JSON.stringify({ head, files: sourceFiles.length, working_tree_sha256: fingerprint, material_files: materialFiles.length, ui_files: uiFiles.length, artifact_files: files.length, staged_files: 0 }));
