// Independent byte/framing/native/signature witness. Uses Node's crypto and an
// independent canonical encoder; it imports no Go producer implementation.
import { createHash, createPublicKey, verify } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { witnessWindowsSources } from './windows-witness.mjs';
import { witnessVERIS } from './veris-witness.mjs';
import { witnessEntraEntries } from './entra-witness.mjs';

const [action, inputPath, reportPath, sourceRoot] = process.argv.slice(2);
const input = JSON.parse(fs.readFileSync(inputPath));
const admittedInputs = {
  'source-audit': ['cartulary.veris_inventory.v1', 'cartulary.attack_source_lock.v1', 'cartulary.windows_event_inventory.v2', 'cartulary.entra_inventory.v2'],
  assemble: ['cartulary.veris_build_input.v1', 'cartulary.attack_package_input.v1', 'cartulary.windows_event_build_input.v2', 'cartulary.entra_build_input.v2'],
  envelope: ['cartulary.veris_envelope_input.v1', 'cartulary.veris_release_envelope_input.v1', 'cartulary.attack_envelope_input.v1', 'cartulary.windows_event_envelope_input.v2', 'cartulary.windows_event_release_envelope_input.v2', 'cartulary.entra_envelope_input.v2', 'cartulary.entra_release_envelope_input.v2'],
};
if (!admittedInputs[action === 'renew' ? 'envelope' : action]?.includes(input.schema_id)) throw new Error('unsupported witness input schema');
if (action === 'source-audit' && input.schema_id === 'cartulary.attack_source_lock.v1' && fs.statSync(reportPath).size === 0) process.exit(0);

const report = JSON.parse(fs.readFileSync(reportPath));
const sha = (b) => createHash('sha256').update(b).digest('hex');
const requireValue = (condition, reason) => { if (!condition) throw new Error(reason); };
function canonical(value) {
  if (Array.isArray(value)) return `[${value.map(canonical).join(',')}]`;
  if (value !== null && typeof value === 'object') return `{${Object.keys(value).sort().map(k => `${JSON.stringify(k)}:${canonical(value[k])}`).join(',')}}`;
  return JSON.stringify(value);
}
function retain(file, value) {
  const bytes = Buffer.from(canonical(value));
  fs.mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  try { fs.writeFileSync(file, bytes, { flag: 'wx', mode: 0o600 }); }
  catch (error) { if (error.code !== 'EEXIST' || !fs.readFileSync(file).equals(bytes)) throw error; }
  return sha(bytes);
}
function logical(root) {
  const raw = fs.readFileSync(path.join(root, 'manifest.json'));
  const manifest = JSON.parse(raw);
  requireValue(raw.equals(Buffer.from(canonical(manifest))), 'manifest JCS witness');
  const framed = createHash('sha256').update('cartulary.reference_pack.payload.v1\0');
  const files = [];
  for (const f of manifest.files) {
    const bytes = fs.readFileSync(path.join(root, f.path));
    const digest = sha(bytes);
    requireValue(digest === f.sha256 && bytes.length === f.size_bytes, 'member witness');
    const name = Buffer.from(f.path);
    const length = Buffer.alloc(4); length.writeUInt32BE(name.length);
    const size = Buffer.alloc(8); size.writeBigUInt64BE(BigInt(bytes.length));
    framed.update(length).update(name).update(size).update(Buffer.from(digest, 'hex'));
    files.push({ path: f.path, size_bytes: bytes.length, sha256: digest });
  }
  const payload = framed.digest('hex');
  requireValue(sha(raw) === report.manifest_sha256 && payload === report.payload_sha256, 'framed identity witness');
  return { manifest, manifest_sha256: sha(raw), payload_sha256: payload, files };
}
let witness;
let output;
if (action === 'source-audit' && input.schema_id === 'cartulary.veris_inventory.v1') {
  const root = path.join(sourceRoot, 'candidates', sha(fs.readFileSync(reportPath)), 'payload');
  witness = { schema_id: 'cartulary.veris_independent_corpus_witness.v1', ...witnessVERIS(fs.readFileSync(path.join(root, 'objects.ndjson')), fs.readFileSync(path.join(root, 'relationships.ndjson'))) };
  output = path.join(sourceRoot, 'witnesses', `audit-${sha(fs.readFileSync(reportPath))}.json`);
} else if (action === 'assemble' && input.schema_id === 'cartulary.veris_build_input.v1') {
  const root = path.join(input.output_root, report.logical_path);
  const result = logical(root);
  witness = { schema_id: 'cartulary.veris_independent_logical_witness.v1', ...result, ...witnessVERIS(fs.readFileSync(path.join(root, 'payload/objects.ndjson')), fs.readFileSync(path.join(root, 'payload/relationships.ndjson')), result.manifest) };
  delete witness.manifest;
  output = path.join(input.output_root, 'witnesses', `logical-${report.manifest_sha256}.json`);
} else if (action === 'source-audit' && input.schema_id === 'cartulary.entra_inventory.v2') {
  const result = witnessEntraEntries(fs.readFileSync(path.join(sourceRoot, 'candidates', `${report.entries_sha256}.ndjson`)));
  witness = { schema_id: 'cartulary.entra_independent_corpus_witness.v2', ...result };
  output = path.join(sourceRoot, 'witnesses', `audit-${sha(fs.readFileSync(reportPath))}.json`);
} else if (action === 'assemble' && input.schema_id === 'cartulary.entra_build_input.v2') {
  const root = path.join(input.output_root, report.logical_path);
  const result = logical(root);
  const corpus = witnessEntraEntries(fs.readFileSync(path.join(root, 'payload/entries.ndjson')), result.manifest);
  witness = { schema_id: 'cartulary.entra_independent_logical_witness.v2', manifest_sha256: result.manifest_sha256, payload_sha256: result.payload_sha256, files: result.files, ...corpus };
  output = path.join(input.output_root, 'witnesses', `logical-${report.manifest_sha256}.json`);
} else if (action === 'source-audit' && input.schema_id === 'cartulary.windows_event_inventory.v2') {
  const root = path.join(sourceRoot, 'audits', sha(fs.readFileSync(reportPath)));
  witness = witnessWindowsSources(fs.readFileSync(path.join(root, 'evidence.tar')), fs.readFileSync(path.join(root, 'entries.candidate.ndjson')), report);
  output = path.join(sourceRoot, 'witnesses', `audit-${sha(fs.readFileSync(reportPath))}.json`);
} else if (action === 'assemble' && input.schema_id === 'cartulary.windows_event_build_input.v2') {
  const root = path.join(input.output_root, report.logical_path);
  const result = logical(root);
  witness = { schema_id: 'cartulary.windows_event_independent_logical_witness.v2', manifest_sha256: result.manifest_sha256, payload_sha256: result.payload_sha256, files: result.files };
  output = path.join(input.output_root, 'witnesses', `logical-${report.manifest_sha256}.json`);
} else if (action === 'assemble' && input.schema_id === 'cartulary.attack_package_input.v1') {
  const root = path.join(input.output_root, report.logical_path);
  const result = logical(root);
  const lock = JSON.parse(fs.readFileSync(input.source_lock));
  const source = new Map();
  const seen = new Set();
  for (const artifact of lock.artifacts) {
    const bytes = fs.readFileSync(path.join(input.source_root, 'sources', artifact.sha256, artifact.name));
    requireValue(sha(bytes) === artifact.sha256 && bytes.length === artifact.size_bytes, 'source lock witness');
    if (artifact.domain === 'license') {
      requireValue(bytes.equals(fs.readFileSync(path.join(root, 'notices/MITRE-ATTACK.txt'))), 'exact notice witness');
      continue;
    }
    for (const object of JSON.parse(bytes).objects) {
      const digest = sha(canonical(object));
      requireValue(!source.has(object.id) || source.get(object.id) === digest, 'cross-domain native collision');
      source.set(object.id, digest);
    }
  }
  const rows = {};
  let embedded = 0;
  for (const kind of ['objects', 'relationships']) {
    const bytes = fs.readFileSync(path.join(root, `payload/${kind}.ndjson`));
    requireValue(bytes.at(-1) === 10, 'terminal LF witness');
    const lines = bytes.toString('utf8').slice(0, -1).split('\n');
    let maximum = 0;
    for (const line of lines) {
      const row = JSON.parse(line);
      requireValue(line === canonical(row), 'row JCS witness');
      maximum = Math.max(maximum, Buffer.byteLength(line) + 1);
      if (row.source !== null) {
        requireValue(source.get(row.source.id) === sha(canonical(row.source)) && !seen.has(row.source.id), 'exact native preservation witness');
        seen.add(row.source.id);
      } else embedded++;
    }
    requireValue(maximum <= 4194304, 'inclusive native line witness');
    rows[kind] = { count: lines.length, maximum_line_bytes_including_lf: maximum };
  }
  requireValue(seen.size === source.size, 'complete native category witness');
  requireValue(rows.objects.count === report.object_count && rows.relationships.count === report.relationship_count, 'canonical counts witness');
  witness = { schema_id: 'cartulary.attack_independent_logical_witness.v1', implementation: 'node_crypto_independent_jcs', manifest_sha256: result.manifest_sha256, payload_sha256: result.payload_sha256, files: result.files, rows, native_records: source.size, embedded_relationships: embedded, source_lock_sha256: sha(fs.readFileSync(input.source_lock)), exact_native_preservation: true, exact_mitre_notice: true };
  output = path.join(input.output_root, 'witnesses', `logical-${report.manifest_sha256}.json`);
} else if (action === 'envelope' || action === 'renew') {
  const schemas = {
    'cartulary.veris_envelope_input.v1': 'cartulary.veris_independent_envelope_witness.v1',
    'cartulary.veris_release_envelope_input.v1': 'cartulary.veris_independent_envelope_witness.v1',
    'cartulary.attack_envelope_input.v1': 'cartulary.attack_independent_envelope_witness.v1',
    'cartulary.windows_event_envelope_input.v2': 'cartulary.windows_event_independent_envelope_witness.v2',
    'cartulary.windows_event_release_envelope_input.v2': 'cartulary.windows_event_independent_envelope_witness.v2',
    'cartulary.entra_envelope_input.v2': 'cartulary.entra_independent_envelope_witness.v2',
    'cartulary.entra_release_envelope_input.v2': 'cartulary.entra_independent_envelope_witness.v2',
  };
  requireValue(Object.hasOwn(schemas, input.schema_id), 'unsupported envelope witness schema');
  const root = path.join(input.output_root, report.envelope_path);
  const bytes = fs.readFileSync(path.join(root, 'container.pack'));
  const bootstrapBytes = fs.readFileSync(path.join(root, 'bootstrap.json'));
  requireValue(sha(bytes) === report.container_sha256 && bytes.length === report.container_bytes && sha(bootstrapBytes) === report.bootstrap_sha256, 'transport witness');
  const members = new Map();
  let offset = 0;
  let previous = '';
  for (; offset + 512 <= bytes.length; ) {
    const header = bytes.subarray(offset, offset + 512);
    if (header.every(v => v === 0)) break;
    const field = (start, end) => header.subarray(start, end).toString('utf8').replace(/\0.*$/s, '');
    const name = field(0, 100);
    const size = Number.parseInt(field(124, 136).trim(), 8);
    const checksum = Number.parseInt(field(148, 156).trim(), 8);
    let sum = 0; for (let i = 0; i < 512; i++) sum += i >= 148 && i < 156 ? 32 : header[i];
    requireValue(sum === checksum && field(257, 263) === 'ustar' && name > previous && !members.has(name), 'ustar header witness');
    requireValue(Number.parseInt(field(136, 148).trim(), 8) === 0, 'deterministic archive time');
    const end = offset + 512 + size;
    const next = offset + 512 + Math.ceil(size / 512) * 512;
    requireValue(next <= bytes.length && bytes.subarray(end, next).every(v => v === 0), 'ustar padding witness');
    members.set(name, bytes.subarray(offset + 512, end)); previous = name; offset = next;
  }
  requireValue(bytes.length - offset >= 1024 && bytes.subarray(offset).every(v => v === 0), 'ustar terminator witness');
  const bootstrap = JSON.parse(bootstrapBytes);
  const rootEnvelope = bootstrap.repositories[0].trusted_root;
  const trusted = rootEnvelope.signed;
  const signers = {};
  for (const [role, envelope] of [['root', rootEnvelope], ...['targets', 'snapshot', 'timestamp'].map(r => [r, JSON.parse(members.get(`metadata/${r}.json`))])]) {
    const ids = [];
    for (const signature of envelope.signatures) {
      requireValue(trusted.roles[role].keyids.includes(signature.keyid), 'signature role witness');
      const key = trusted.keys[signature.keyid];
      requireValue(sha(canonical(key)) === signature.keyid, 'public key identity witness');
      const publicKey = createPublicKey({ key: Buffer.concat([Buffer.from('302a300506032b6570032100', 'hex'), Buffer.from(key.keyval.public, 'hex')]), format: 'der', type: 'spki' });
      requireValue(verify(null, Buffer.from(canonical(envelope.signed)), publicKey, Buffer.from(signature.sig, 'hex')), 'Ed25519 signature witness');
      requireValue(!ids.includes(signature.keyid), 'duplicate signer witness'); ids.push(signature.keyid);
    }
    requireValue(ids.length === (role === 'root' ? 2 : 1) && ids.length >= trusted.roles[role].threshold, 'threshold witness');
    signers[role] = ids.sort();
  }
  const targets = JSON.parse(members.get('metadata/targets.json')).signed;
  for (const [name, descriptor] of Object.entries(targets.targets)) {
    const b = members.get(name); requireValue(b && b.length === descriptor.length && sha(b) === descriptor.hashes.sha256, 'signed target witness');
  }
  for (const [parent, child] of [['snapshot', 'targets'], ['timestamp', 'snapshot']]) {
    const p = JSON.parse(members.get(`metadata/${parent}.json`)).signed;
    const b = members.get(`metadata/${child}.json`); const d = p.meta[`${child}.json`];
    requireValue(b.length === d.length && sha(b) === d.hashes.sha256 && JSON.parse(b).signed.version === d.version, 'metadata link witness');
  }
  requireValue(sha(members.get('manifest.json')) === report.manifest_sha256, 'envelope logical identity witness');
  witness = { schema_id: schemas[input.schema_id], implementation: 'node_crypto_ed25519_independent_ustar', manifest_sha256: report.manifest_sha256, payload_sha256: report.payload_sha256, container_sha256: sha(bytes), container_bytes: bytes.length, bootstrap_sha256: sha(bootstrapBytes), metadata_version: report.metadata_version, signers, byte_integrity: true, signature_integrity: true, freshness_verified: false };
  output = path.join(input.output_root, 'witnesses', `envelope-${report.container_sha256}.json`);
}
if (witness) process.stdout.write(`Independent witness retained: ${output}; sha256=${retain(output, witness)}\n`);
if (action !== 'source-audit') process.stdout.write(`${canonical(report)}\n`);
