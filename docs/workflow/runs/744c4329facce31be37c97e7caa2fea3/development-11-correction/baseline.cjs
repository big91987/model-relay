const fs = require('node:fs');
const crypto = require('node:crypto');
const { execFileSync } = require('node:child_process');
const root = 'docs/workflow/runs/744c4329facce31be37c97e7caa2fea3';
const manifestPath = `${root}/development-11-correction/SHA256SUMS`;
const digest = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const started = new Date().toISOString();
const manifest = fs.readFileSync(manifestPath);
const entries = manifest.toString().trim().split('\n').map(line => {
  const [, expected, path] = line.match(/^([a-f0-9]{64})  (.+)$/);
  const actual = digest(fs.readFileSync(path));
  return { path, expected_sha256: expected, actual_sha256: actual, matched: actual === expected };
});
const allowedChanges = new Set(['docs/harness/opensandbox-run-sharing.md', `${root}/development-11.md`]);
const previous = fs.readFileSync(`${root}/development-11/SHA256SUMS`, 'utf8').trim().split('\n').filter(line => !allowedChanges.has(line.slice(66))).map(line => {
  const [, expected, path] = line.match(/^([a-f0-9]{64})  (.+)$/);
  return { path, matched: digest(fs.readFileSync(path)) === expected };
});
const manifestUnchanged = digest(fs.readFileSync(manifestPath)) === digest(manifest);
const current = fs.readFileSync('docs/harness/opensandbox-run-sharing.md', 'utf8');
const snapshotMatches = current === fs.readFileSync(`${root}/development-11-correction/opensandbox-run-sharing.md`, 'utf8');
const boundariesPresent = current.includes('## 停止、恢复与环境释放') && current.includes('研发交接前基线') && current.includes('只执行冻结的 `make verify`');
const exitCode = entries.every(entry => entry.matched) && previous.every(entry => entry.matched) && manifestUnchanged && snapshotMatches && boundariesPresent ? 0 : 1;
console.log(JSON.stringify({ role: 'development', scope: 'pre-handoff delivery baseline; not fixed runner before/after sampling or QA', started_utc: started, finished_utc: new Date().toISOString(), head: execFileSync('git', ['rev-parse', 'HEAD'], { encoding: 'utf8' }).trim(), manifest_path: manifestPath, manifest_sha256: digest(manifest), manifest_unchanged_during_audit: manifestUnchanged, entries_count: entries.length, entries, preserved_previous_entries_count: previous.length, preserved_previous_entries_match: previous.every(entry => entry.matched), snapshot_matches: snapshotMatches, corrected_boundaries_present: boundariesPresent, exit_code: exitCode }, null, 2));
process.exitCode = exitCode;
