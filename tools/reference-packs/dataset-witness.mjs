// Independent qualification of delivered bytes. This uses Node crypto and a
// separate canonical encoder, never the producer, upstream prose, or private keys.
import fs from 'node:fs';
import path from 'node:path';
import { createHash, createPublicKey, verify } from 'node:crypto';

const [action, inputPath, reportPath] = process.argv.slice(2);
const input = JSON.parse(fs.readFileSync(inputPath));
const report = JSON.parse(fs.readFileSync(reportPath));
const check = (value, reason) => { if (!value) throw new Error(`dataset witness: ${reason}`); };
const canonical = value => Array.isArray(value) ? `[${value.map(canonical).join(',')}]` : value !== null && typeof value === 'object' ? `{${Object.keys(value).sort().map(k => `${JSON.stringify(k)}:${canonical(value[k])}`).join(',')}}` : JSON.stringify(value);
const sha = value => createHash('sha256').update(value).digest('hex');
const equal = (a, b) => canonical(a) === canonical(b);
const compare = (a, b) => Buffer.compare(Buffer.from(a), Buffer.from(b));
const ascii = value => value.normalize('NFC').replace(/[A-Z]/gu, value => value.toLowerCase());
function document(bytes) { const value = JSON.parse(bytes); check(Buffer.from(canonical(value)).equals(bytes), 'canonical bytes'); return value; }
function archive(bytes) {
  const files = new Map(); let offset = 0, previous = '';
  for (; offset + 512 <= bytes.length;) {
    const header = bytes.subarray(offset, offset + 512);
    if (header.every(value => value === 0)) break;
    const field = (start, end) => header.subarray(start, end).toString('utf8').replace(/\0.*$/su, '');
    const name = field(0, 100), size = Number.parseInt(field(124, 136).trim(), 8);
    let sum = 0; for (let i = 0; i < 512; i++) sum += i >= 148 && i < 156 ? 32 : header[i];
    check(sum === Number.parseInt(field(148, 156).trim(), 8) && field(257, 263) === 'ustar' && compare(name, previous) > 0 && !files.has(name), 'archive header');
    check(Number.isSafeInteger(size) && size >= 0 && Number.parseInt(field(136, 148).trim(), 8) === 0, 'archive size/time');
    const end = offset + 512 + size, next = offset + 512 + Math.ceil(size / 512) * 512;
    check(next <= bytes.length && bytes.subarray(end, next).every(value => value === 0), 'archive padding');
    files.set(name, bytes.subarray(offset + 512, end)); previous = name; offset = next;
  }
  check(bytes.length - offset >= 1024 && bytes.subarray(offset).every(value => value === 0), 'archive terminator');
  return files;
}
function logical(read, expectedHash) {
  const raw = read('manifest.json'), manifest = document(raw);
  check(sha(raw) === expectedHash, 'exact manifest identity');
  const framed = createHash('sha256').update('cartulary.reference_pack.payload.v1\0');
  for (const f of manifest.files) {
    check(/^(payload|notices)\/[a-zA-Z0-9_.-]+$/u.test(f.path), 'logical member path');
    const bytes = read(f.path), digest = sha(bytes), name = Buffer.from(f.path);
    check(digest === f.sha256 && bytes.length === f.size_bytes, 'logical inventory');
    const length = Buffer.alloc(4); length.writeUInt32BE(name.length);
    const size = Buffer.alloc(8); size.writeBigUInt64BE(BigInt(bytes.length));
    framed.update(length).update(name).update(size).update(Buffer.from(digest, 'hex'));
  }
  return { manifest, payload_sha256: framed.digest('hex') };
}
function rows(bytes) {
  if (bytes.length === 0) return [];
  check(bytes.at(-1) === 10, 'NDJSON terminal LF');
  return bytes.subarray(0, -1).toString('utf8').split('\n').map(line => { check(Buffer.byteLength(line) + 1 <= 4194304, 'row limit'); return document(Buffer.from(line)); });
}
function key(lookup, input) {
  return canonical(lookup.components.map(component => {
    let value = lookup.input_type === 'scalar' ? input : input[component.name];
    if (component.type === 'integer') { check(Number.isSafeInteger(value), 'safe integer lookup'); return value; }
    check(typeof value === 'string' && value.length > 0 && !/[\u0000-\u001f\u007f-\u009f\u2028\u2029]/u.test(value), 'lookup string');
    if (component.normalization === 'ascii_case_insensitive') return ascii(value);
    if (component.normalization === 'uuid') { check(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/iu.test(value), 'UUID lookup'); return value.toLowerCase(); }
    check(component.normalization === 'exact_nfc', 'finite normalization vocabulary'); return value.normalize('NFC');
  }));
}
function expectedLookups(pack, data) {
  if (pack === 'enrichment.windows_event_ids') return { event_id: [data.event_id], provider_event_id: [{ event_id: data.event_id, provider_name: data.provider_name }], ...(data.event_version === null ? {} : { provider_event_version: [{ event_id: data.event_id, event_version: data.event_version, provider_name: data.provider_name }] }) };
  if (pack === 'enrichment.entra_app_ids') return { app_id: [data.app_id], ...(data.aliases.length ? { alias: data.aliases } : {}), ...(data.service_principal_names.length ? { service_principal_name: data.service_principal_names } : {}) };
  const external = [...new Set(data.external_refs.map(ref => ref.external_id))];
  return { object_id: [data.object_id], ...(data.aliases.length ? { alias: data.aliases } : {}), ...(external.length ? { external_id: external } : {}) };
}
function corpus(read, manifest) {
  check(manifest.pack_contract_version === 'cartulary.reference_pack_contract.v3' && manifest.content_profile_id === 'cartulary.reference_pack.dataset.v1' && manifest.content_profile_version === '1', 'current generic contract');
  const expectations = JSON.parse(fs.readFileSync('fixtures/reference-packs/dataset-v1/predecessors.v1.json'));
  const baseline = expectations.source_versions?.[manifest.pack_key]?.[manifest.source_version] ?? expectations.packs[manifest.pack_key];
  check(baseline, 'explicit predecessor expectation');
  let oldRead;
  if (baseline.container_sha256) { const bytes = fs.readFileSync(path.join(baseline.logical_root, 'container.pack')); check(sha(bytes) === baseline.container_sha256, 'predecessor archive'); const files = archive(bytes); oldRead = name => files.get(name); }
  else oldRead = name => fs.readFileSync(path.join(baseline.logical_root, name));
  const prior = logical(oldRead, baseline.manifest_sha256).manifest;
  const descriptorBytes = read('payload/descriptor.json'), descriptor = document(descriptorBytes);
  check(descriptorBytes.length <= 65536 && descriptor.lookups.length <= 32, 'descriptor limits');
  const objects = rows(read('payload/objects.ndjson')), edges = rows(read('payload/relationships.ndjson'));
  const beforeObjects = rows(oldRead(prior.content_summary.kind === 'entries' ? 'payload/entries.ndjson' : 'payload/objects.ndjson'));
  const beforeEdges = prior.content_summary.kind === 'entries' ? [] : rows(oldRead('payload/relationships.ndjson'));
  check(objects.length === baseline.objects && edges.length === baseline.relationships && beforeObjects.length === objects.length && beforeEdges.length === edges.length, 'complete corpus counts');
  const coverage = new Map(descriptor.lookups.map(lookup => [lookup.name, { objects: 0, postings: 0, keys: new Map() }]));
  let postings = 0, keyBytes = 0, previous = '';
  for (let i = 0; i < objects.length; i++) {
    const object = objects[i], data = beforeObjects[i];
    check(equal(object.data, data), `complete object preservation at index ${i}`);
    check(object.object_id === (data.object_id ?? data.entry_id) && compare(object.object_id, previous) > 0, 'source-stable ordered identity'); previous = object.object_id;
    for (const name of ['display_label', 'description', 'deprecated', 'source_refs']) check(equal(object[name], data[name]), 'common factual field');
    check(equal(object.replacement_object_id, data.replacement_object_id ?? data.replacement_entry_id ?? null), 'replacement preservation');
    const expected = expectedLookups(manifest.pack_key, data);
    check(equal(Object.keys(object.lookup_values).sort(), Object.keys(expected).sort()), 'lookup names');
    let objectPostings = 0;
    for (const lookup of descriptor.lookups) {
      const actual = object.lookup_values[lookup.name] ?? [], keys = actual.map(value => key(lookup, value));
      check(equal(keys, (expected[lookup.name] ?? []).map(value => key(lookup, value)).sort(compare)), 'factual lookup mapping and ordering');
      check(new Set(keys).size === keys.length, 'no repeated normalized value');
      const totals = coverage.get(lookup.name); if (keys.length) totals.objects++;
      for (const key of keys) {
        const size = Buffer.byteLength(key); check(size <= 65536, 'encoded key limit');
        if (lookup.unique) check(!totals.keys.has(key), 'declared lookup uniqueness');
        totals.keys.set(key, true); totals.postings++; postings++; objectPostings++; keyBytes += size;
      }
    }
    check(objectPostings <= 128, 'per-object posting limit');
  }
  for (let i = 0; i < edges.length; i++) {
    check(equal(edges[i].data, beforeEdges[i]), `complete relationship preservation at index ${i}`);
    for (const name of ['relationship_id', 'relationship_type', 'source_object_id', 'target_object_id', 'source_refs']) check(equal(edges[i][name], beforeEdges[i][name]), 'relationship fact preservation');
  }
  check(postings <= 8000000 && keyBytes <= 268435456, 'aggregate lookup limits');
  for (const lookup of descriptor.lookups) { const c = coverage.get(lookup.name); check(equal(lookup.coverage, { objects: c.objects, postings: c.postings, distinct_keys: c.keys.size }), 'actual descriptor coverage'); }
  for (const notice of prior.license.notice_paths) check(oldRead(notice).equals(read(notice)), 'exact retained notices');
  return { objects: objects.length, relationships: edges.length, postings, encoded_key_bytes: keyBytes, predecessor_manifest_sha256: baseline.manifest_sha256, complete_native_preservation: true, coverage: descriptor.lookups.map(({ name, coverage }) => ({ name, ...coverage })) };
}
let witness, output;
if (action === 'assemble' || action === 'assemble-fixture') {
  const root = path.join(input.output_root, report.logical_path), read = name => fs.readFileSync(path.join(root, name));
  const result = logical(read, report.manifest_sha256); check(result.payload_sha256 === report.payload_sha256, 'payload identity');
  witness = { schema_id: 'cartulary.reference_pack_dataset_witness.v1', manifest_sha256: report.manifest_sha256, payload_sha256: report.payload_sha256, ...corpus(read, result.manifest) };
  output = path.join(input.output_root, 'witnesses', `logical-${report.manifest_sha256}.json`);
} else {
  check(['envelope', 'renew', 'envelope-fixture', 'renew-fixture'].includes(action), 'explicit action');
  const root = path.join(input.envelope.output_root, report.envelope_path);
  const bytes = fs.readFileSync(path.join(root, 'container.pack')), bootstrapBytes = fs.readFileSync(path.join(root, 'bootstrap.json'));
  check(sha(bytes) === report.container_sha256 && bytes.length === report.container_bytes && sha(bootstrapBytes) === report.bootstrap_sha256, 'exact envelope identity');
  const files = archive(bytes), bootstrap = document(bootstrapBytes), trusted = bootstrap.repositories[0].trusted_root.signed, signers = {};
  const now = Date.now();
  for (const [role, signed] of [['root', bootstrap.repositories[0].trusted_root], ...['targets', 'snapshot', 'timestamp'].map(role => [role, document(files.get(`metadata/${role}.json`))])]) {
    check(Date.parse(signed.signed.expires) > now, 'trusted current-clock expiry');
    const ids = new Set();
    for (const signature of signed.signatures) {
      check(trusted.roles[role].keyids.includes(signature.keyid) && !ids.has(signature.keyid), 'authorized distinct signer');
      const key = trusted.keys[signature.keyid]; check(sha(canonical(key)) === signature.keyid, 'public key identity');
      const publicKey = createPublicKey({ key: Buffer.concat([Buffer.from('302a300506032b6570032100', 'hex'), Buffer.from(key.keyval.public, 'hex')]), format: 'der', type: 'spki' });
      check(verify(null, Buffer.from(canonical(signed.signed)), publicKey, Buffer.from(signature.sig, 'hex')), 'signature'); ids.add(signature.keyid);
    }
    check(ids.size >= trusted.roles[role].threshold, 'signature threshold'); signers[role] = [...ids].sort();
  }
  const targets = document(files.get('metadata/targets.json')).signed;
  for (const [name, d] of Object.entries(targets.targets)) { const raw = files.get(name); check(raw && raw.length === d.length && sha(raw) === d.hashes.sha256, 'signed inventory'); }
  for (const [parent, child] of [['snapshot', 'targets'], ['timestamp', 'snapshot']]) { const p = document(files.get(`metadata/${parent}.json`)).signed, raw = files.get(`metadata/${child}.json`), d = p.meta[`${child}.json`]; check(raw.length === d.length && sha(raw) === d.hashes.sha256 && document(raw).signed.version === d.version, 'metadata linkage'); }
  const result = logical(name => files.get(name), report.manifest_sha256);
  check(result.payload_sha256 === report.payload_sha256 && targets.cartulary.payload_sha256 === result.payload_sha256 && targets.cartulary.manifest_sha256 === report.manifest_sha256, 'signed logical binding');
  witness = { schema_id: 'cartulary.reference_pack_dataset_envelope_witness.v1', manifest_sha256: report.manifest_sha256, payload_sha256: report.payload_sha256, container_sha256: report.container_sha256, bootstrap_sha256: report.bootstrap_sha256, signers, freshness_verified: true, verified_at: new Date(now).toISOString() };
  output = path.join(input.envelope.output_root, 'witnesses', `envelope-${report.container_sha256}-${now}.json`);
}
const raw = Buffer.from(canonical(witness)); fs.mkdirSync(path.dirname(output), { recursive: true, mode: 0o700 });
try { fs.writeFileSync(output, raw, { flag: 'wx', mode: 0o600 }); } catch (error) { if (error.code !== 'EEXIST' || !fs.readFileSync(output).equals(raw)) throw error; }
process.stdout.write(`${canonical(report)}\n${canonical({ witness_path: output, witness_sha256: sha(raw) })}\n`);
