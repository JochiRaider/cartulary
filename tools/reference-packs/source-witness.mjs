// Independent source-audit witnesses. Publication uses dataset-witness.mjs.
// No Go producer implementation supplies these factual checks.
import { createHash } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { readAuditReport } from './audit-report.mjs';
import { witnessWindowsSources } from './windows-witness.mjs';
import { witnessVERIS } from './veris-witness.mjs';
import { witnessEntraEntries } from './entra-witness.mjs';

const [action, inputPath, reportPath, sourceRoot, contextPath] = process.argv.slice(2);
if (action !== 'source-audit') throw new Error('unsupported witness action');
const input = JSON.parse(fs.readFileSync(inputPath));
const producer = { 'cartulary.veris_inventory.v1': 'veris', 'cartulary.windows_event_inventory.v2': 'windows', 'cartulary.entra_inventory.v2': 'entra' }[input.schema_id];
if (!producer) throw new Error('unsupported witness input schema');
const report = readAuditReport(producer, reportPath, JSON.parse(fs.readFileSync(contextPath)));
const sha = b => createHash('sha256').update(b).digest('hex');
const canonical = value => Array.isArray(value) ? `[${value.map(canonical).join(',')}]` : value !== null && typeof value === 'object' ? `{${Object.keys(value).sort().map(k => `${JSON.stringify(k)}:${canonical(value[k])}`).join(',')}}` : JSON.stringify(value);
const digest = sha(fs.readFileSync(reportPath));
const checked = (file, expected) => {
  const bytes = fs.readFileSync(file);
  if (sha(bytes) !== expected) throw new Error('source-audit witness input digest mismatch');
  return bytes;
};
function retain(file, value) {
  const bytes = Buffer.from(canonical(value));
  fs.mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  try { fs.writeFileSync(file, bytes, { flag: 'wx', mode: 0o600 }); }
  catch (error) { if (error.code !== 'EEXIST' || !fs.readFileSync(file).equals(bytes)) throw error; }
  return sha(bytes);
}

let witness;
if (producer === 'veris') {
  const root = path.join(sourceRoot, 'candidates', digest, 'payload');
  witness = { schema_id: 'cartulary.veris_independent_corpus_witness.v1', ...witnessVERIS(checked(path.join(root, 'objects.ndjson'), report.objects_sha256), checked(path.join(root, 'relationships.ndjson'), report.relationships_sha256)) };
} else if (producer === 'entra') {
  witness = { schema_id: 'cartulary.entra_independent_corpus_witness.v2', ...witnessEntraEntries(checked(path.join(sourceRoot, 'candidates', `${report.entries_sha256}.ndjson`), report.entries_sha256)) };
} else {
  const root = path.join(sourceRoot, 'audits', digest);
  witness = witnessWindowsSources(checked(path.join(root, 'evidence.tar'), report.evidence_sha256), checked(path.join(root, 'entries.candidate.ndjson'), report.payload_file_sha256), report);
}
const output = path.join(sourceRoot, 'witnesses', `audit-${digest}.json`);
process.stdout.write(`Independent witness retained: ${output}; sha256=${retain(output, witness)}\n`);
