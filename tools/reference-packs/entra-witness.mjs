// Independent canonical corpus witness. Inputs are JSON expectations and pack
// bytes; upstream Markdown and production signing secrets are never read.
import fs from 'node:fs';
import { createHash } from 'node:crypto';

export function witnessEntraEntries(bytes, manifest = null) {
  const expected = JSON.parse(fs.readFileSync('contracts/reference-pack-producers/entra-app-ids/baseline.v2.json')).entries;
  const editorial = JSON.parse(fs.readFileSync('contracts/reference-pack-producers/entra-app-ids/editorial.v2.json'));
  const inventory = JSON.parse(fs.readFileSync('contracts/reference-pack-producers/entra-app-ids/inventory.v2.json'));
  const sourceIDs = new Map(inventory.sources.map(s => [s.url, s.source_id]));
  if (manifest) {
    for (const source of inventory.sources) {
      const matches = manifest.source_artifacts.filter(a => a.source_ref === source.url);
      if (matches.length !== 1 || matches[0].source_version !== source.commit || matches[0].sha256 !== source.sha256) throw new Error('Entra witness: exact source revision/digest');
    }
  }

  const summaries = new Map(editorial.summaries.map(s => [s.app_id, s.text]));
  const check = (condition, message) => { if (!condition) throw new Error(`Entra witness: ${message}`); };
  const canonical = value => Array.isArray(value) ? `[${value.map(canonical).join(',')}]` : value !== null && typeof value === 'object' ? `{${Object.keys(value).sort().map(k => `${JSON.stringify(k)}:${canonical(value[k])}`).join(',')}}` : JSON.stringify(value);
  check(bytes.at(-1) === 10 && !bytes.includes(13), 'NDJSON framing');
  const lines = bytes.toString('utf8').slice(0, -1).split('\n');
  check(lines.length === 139 && expected.length === 139, 'complete baseline');
  let aliases = 0, historical = 0;
  for (let i = 0; i < lines.length; i++) {
    const entry = JSON.parse(lines[i]), want = expected[i];
    check(lines[i] === canonical(entry), 'canonical entry');
    check(entry.schema_id === 'cartulary.reference_pack.enrichment.entra_app_ids.entry.v2' && entry.entry_id === want.app_id && entry.app_id === want.app_id && entry.display_label === want.display_label, 'exact identity and name');
    check(canonical(entry.aliases) === canonical(want.aliases), 'aliases');
    const old = want.source_status === 'historical_documented';
    check(entry.description === (old ? editorial.historical_note : summaries.get(want.app_id) ?? null), 'original description and historical qualifier');
    check(entry.publisher === 'Microsoft' && entry.deprecated === false && entry.replacement_entry_id === null && entry.service_principal_names.length === 0 && canonical(entry.extensions) === '{}', 'explicit materialization');
    if (manifest) {
      const locators = entry.source_refs.map(ref => `${sourceIDs.get(manifest.source_artifacts[ref.artifact_index].source_ref)}:${ref.locator}`).sort();
      check(canonical(locators) === canonical(want.source_locators), 'exact source locators');
    }
    aliases += entry.aliases.length;
    historical += Number(old);
  }
  check(aliases === 139 && historical === 119, 'alias and historical totals');
  return { entry_count: lines.length, alias_count: aliases, historical_only_entry_count: historical, entries_sha256: createHash('sha256').update(bytes).digest('hex'), exact_baseline: true, upstream_source_verification: false };
}
