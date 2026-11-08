// Canonical artifacts and independently retained JSON expectations only.
// No upstream prose, network access, producer implementation or signing keys.
import fs from 'node:fs';
import { createHash } from 'node:crypto';

const canonical = v => Array.isArray(v) ? `[${v.map(canonical).join(',')}]` : v !== null && typeof v === 'object' ? `{${Object.keys(v).sort().map(k => `${JSON.stringify(k)}:${canonical(v[k])}`).join(',')}}` : JSON.stringify(v);
const sha = b => createHash('sha256').update(b).digest('hex');
const check = (ok, reason) => { if (!ok) throw new Error(`VERIS witness: ${reason}`); };
const equal = (a, b) => canonical(a) === canonical(b);
const escape = s => s.replaceAll('~', '~0').replaceAll('/', '~1');
const unescape = s => s.replaceAll('~1', '/').replaceAll('~0', '~');
const identity = (kind, key, ...value) => `veris-${kind}-${sha(canonical(['veris-community', kind, key, ...value]))}`;
const read = name => JSON.parse(fs.readFileSync(`contracts/reference-pack-producers/veris/${name}.v1.json`));
const pointer = (v, p) => p === '' ? v : p.slice(1).split('/').reduce((o, k) => o?.[unescape(k)], v);

export function witnessVERIS(objectBytes, relationshipBytes, manifest = null) {
  const baseline = read('baseline'), rules = read('rules').rules, inventory = read('inventory');
  const rows = bytes => {
    check(bytes.at(-1) === 10, 'terminal LF');
    return bytes.toString('utf8').slice(0, -1).split('\n').map(line => {
      const row = JSON.parse(line);
      check(line === canonical(row) && Buffer.byteLength(line) + 1 <= 1048576, 'canonical bounded row');
      return row;
    });
  };
  const objects = rows(objectBytes), edges = rows(relationshipBytes);
  const byID = new Map(objects.map(o => [o.object_id, o]));
  check(byID.size === objects.length, 'unique object identities');
  const nodes = objects.filter(o => ['schema-root', 'schema-node'].includes(o.object_type));
  const enums = objects.filter(o => o.object_type === 'enum-value');
  const actualRules = objects.filter(o => o.object_type === 'semantic-rule');
  check(objects.length === nodes.length + enums.length + actualRules.length, 'closed kinds');
  check(nodes.length === 261 && enums.length === 2168 && actualRules.length === rules.length, 'complete corpus counts');
  if (manifest) {
    check(manifest.source_version === '1.4.2' && manifest.content_profile_id === 'cartulary.reference_pack.framework.veris.v2', 'revision and profile');
    for (const s of inventory.sources) {
      const a = manifest.source_artifacts.filter(a => a.source_ref === s.url);
      check(a.length === 1 && a[0].sha256 === s.sha256 && a[0].source_version === `${inventory.source_version}@${inventory.commit}`, 'exact source pins');
    }
  }
  const expectedEdges = [];
  const edge = (kind, from, to, refs) => expectedEdges.push({ kind, from, to, refs });
  const seen = new Set();
  const reconstruct = (id, sourcePointer = '', depth = 0) => {
    check(depth <= 32 && !seen.has(id), 'acyclic bounded tree');
    seen.add(id);
    const o = byID.get(id), s = o?.source;
    check(s && s.pointer === sourcePointer && s.kind === (sourcePointer === '' ? 'schema-root' : 'schema-node'), 'node role/pointer');
    check(id === identity(s.kind, s.pointer) && s.schema_version === '1.4.2', 'stable node identity');
    const out = structuredClone(s.keywords);
    if (Object.hasOwn(out, 'properties')) {
      for (const [name, child] of Object.entries(out.properties)) {
        const childPointer = `${sourcePointer}/properties/${escape(name)}`;
        edge('contains', id, child, byID.get(child)?.source_refs);
        out.properties[name] = reconstruct(child, childPointer, depth + 1);
      }
    }
    if (Object.hasOwn(out, 'items')) {
      const child = out.items;
      edge('has-item-schema', id, child, byID.get(child)?.source_refs);
      out.items = reconstruct(child, `${sourcePointer}/items`, depth + 1);
    }
    if (Object.hasOwn(out, 'enum')) {
      for (const value of out.enum) {
        const enumID = identity('enum-value', sourcePointer, value);
        check(byID.has(enumID), 'all contextual enums');
        edge('allows-value', id, enumID, byID.get(enumID).source_refs);
      }
    }
    return out;
  };
  const roots = nodes.filter(o => o.object_type === 'schema-root');
  check(roots.length === 1, 'one root');
  const restored = reconstruct(roots[0].object_id);
  check(seen.size === nodes.length && equal(restored, baseline.schema), 'exact parsed schema reconstruction');
  for (const o of enums) {
    const s = o.source;
    check(s.kind === 'enum-value' && s.schema_version === '1.4.2' && o.object_id === identity(s.kind, s.pointer, s.value), 'contextual enum identity');
    const node = pointer(baseline.schema, s.pointer);
    check(node?.enum?.includes(s.value), 'no extra enum');
    const parts = s.pointer.slice(1).split('/'), labelParts = [];
    for (let i = 0; i < parts.length; i++) if (parts[i] === 'properties') labelParts.push(parts[++i]);
    check(s.definition === pointer(baseline.labels, `/${labelParts.join('/')}/${escape(s.value)}`), 'exact contextual definition');
  }
  for (const expected of rules) {
    const id = identity('semantic-rule', expected.rule_id), o = byID.get(id);
    check(o?.object_type === 'semantic-rule' && equal(o.source, { kind: 'semantic-rule', schema_version: '1.4.2', rule_id: expected.rule_id, category: expected.category, statement: expected.statement, applies_to: expected.applies_to, qualification: expected.qualification }), 'exact reviewed rule');
    for (const p of expected.applies_to) {
      const from = identity(p === '' ? 'schema-root' : 'schema-node', p);
      check(byID.has(from), 'existing rule target');
      edge('governed-by-rule', from, id, o.source_refs);
    }
    if (manifest) {
      check(o.source_refs.length === 1 && manifest.source_artifacts[o.source_refs[0].artifact_index].source_ref === inventory.sources.find(s => s.path === expected.source_path)?.url && o.source_refs[0].locator === expected.locator, 'rule provenance');
    }
  }
  const tuples = xs => xs.map(e => canonical(e)).sort();
  check(equal(tuples(edges.map(e => ({kind: e.relationship_type, from: e.source_object_id, to: e.target_object_id, refs: e.source_refs}))), tuples(expectedEdges)), 'exact source-backed edges');
  // Explicit source distinctions survive even when their display tokens coincide.
  const unknown = enums.filter(e => e.source.value === 'Unknown');
  check(unknown.length > 1 && new Set(unknown.map(e => e.object_id)).size === unknown.length, 'contextual Unknown');
  check(restored.properties.schema_version.default === '1.4.0', 'source version-default discrepancy');
  check(Object.keys(restored.properties.plus).length > 0, 'open plus preserved');
  return { object_count: objects.length, relationship_count: edges.length, schema_nodes: nodes.length, enum_values: enums.length, semantic_rules: actualRules.length, objects_sha256: sha(objectBytes), relationships_sha256: sha(relationshipBytes), reconstructed_schema_sha256: sha(canonical(restored)), exact_schema_reconstruction: true, exact_contextual_definitions: true, exact_reviewed_rules: true };
}
