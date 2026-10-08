import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';

const destination = 'docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-65';
const digest = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const receipts = [];
const indexPath = `${destination}/ui-artifact-index.json`;
if (fs.existsSync(indexPath)) {
  const previous = JSON.parse(fs.readFileSync(indexPath));
  const rejected = previous.retained.filter(entry => /continuous-.*-(96921|96937|96963)\./.test(entry.retained));
  if (rejected.length) {
    fs.writeFileSync(`${destination}/rejected-artifact-selection.json`, JSON.stringify({ reason: 'Historical process IDs rejected by actual mtime check before QA decision; originals remain in test-results and historical QA evidence; not current seq64 evidence', rejected }, null, 2) + '\n');
    for (const entry of rejected) fs.unlinkSync(entry.retained);
  }
}
for (const [width, execution] of [[1280, 14624], [1440, 14640], [390, 14654]]) {
  const source = `test-results/visual-${width}/review.json`;
  const review = JSON.parse(fs.readFileSync(source));
  if (review.result !== 'passed' || review.phase !== 'completed' || review.pages.length !== 10) throw new Error('Incomplete current visual evidence');
  const entries = [source];
  for (const page of review.pages) {
    for (const image of page.screenshots) {
      if (digest(fs.readFileSync(image.path)) !== image.sha256) throw new Error('Visual byte mismatch');
      entries.push(image.path);
    }
  }
  const journey = `test-results/continuous-${width}-${execution}.json`;
  const facts = JSON.parse(fs.readFileSync(journey));
  if (facts.result !== 'passed' || facts.checkpoints.length !== 7) throw new Error('Incomplete current journey evidence');
  entries.push(journey);
  for (const name of fs.readdirSync('test-results')) {
    if (name.startsWith(`continuous-${width}-`) && name.endsWith(`-${execution}.png`)) entries.push(`test-results/${name}`);
  }
  for (const filename of entries) {
    const retained = `${destination}/host/${filename.replace(/^test-results\//, '')}`;
    fs.mkdirSync(path.dirname(retained), { recursive: true });
    fs.copyFileSync(filename, retained);
    receipts.push({ source: filename, retained, source_mtime_utc: fs.statSync(filename).mtime.toISOString(), sha256: digest(fs.readFileSync(filename)), provenance: 'seq64 current host artifacts (22:43-22:44 Asia/Shanghai); UI execution from host, not independent QA browser operation' });
  }
}
fs.writeFileSync(`${destination}/ui-artifact-index.json`, JSON.stringify({ checked_at: new Date().toISOString(), seq: 64, retained: receipts }, null, 2) + '\n');
console.log(JSON.stringify({ retained: receipts.length, all_visual_digests_match: true, current_journeys: 3 }));
