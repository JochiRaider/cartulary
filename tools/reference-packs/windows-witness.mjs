// Independent source-identity accounting; never imports the Go producer.
import { createHash } from 'node:crypto';
import assert from 'node:assert/strict';
import { witnessCombinedCSV } from './windows-office-witness.mjs';

const sha = (b) => createHash('sha256').update(b).digest('hex');
const utf8 = (b) => new TextDecoder('utf-8', { fatal: true }).decode(b);
function presentation(value) {
  if(typeof value==='string') return value.replace(/[\u0009-\u000d\u0020\u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+$/u,'');
  if(Array.isArray(value)) return value.map(presentation);
  if(value && typeof value==='object') return Object.fromEntries(Object.entries(value).map(([k,v])=>[k,presentation(v)]));
  return value;
}

export function readEvidenceTar(bytes) {
  const members = new Map();
  let offset = 0;
  let previous = '';
  while (offset + 512 <= bytes.length) {
    const header = bytes.subarray(offset, offset + 512);
    if (header.every(v => v === 0)) break;
    const field = (start, end) => utf8(header.subarray(start, end)).replace(/\0.*$/s, '');
    const name = field(0, 100);
    const size = Number.parseInt(field(124, 136).trim(), 8);
    let checksum = 0;
    for (let i = 0; i < 512; i++) checksum += i >= 148 && i < 156 ? 32 : header[i];
    assert.equal(checksum, Number.parseInt(field(148, 156).trim(), 8));
    assert.equal(field(257, 263), 'ustar');
    assert(name > previous && !members.has(name));
    assert.equal(Number.parseInt(field(136, 148).trim(), 8), 0);
    const end = offset + 512 + size;
    const next = offset + 512 + Math.ceil(size / 512) * 512;
    assert(Number.isSafeInteger(size) && size >= 0 && next <= bytes.length);
    assert(bytes.subarray(end, next).every(v => v === 0));
    members.set(name, bytes.subarray(offset + 512, end));
    previous = name;
    offset = next;
  }
  assert(bytes.length - offset >= 1024 && bytes.subarray(offset).every(v => v === 0));
  return members;
}

const decode = s => s.replace(/&#x([a-f0-9]+);|&#([0-9]+);|&(lt|gt|amp|quot|apos|nbsp);/gi, (_, hex, decimal, named) => {
  if (hex || decimal) return String.fromCodePoint(Number.parseInt(hex || decimal, hex ? 16 : 10));
  return { lt: '<', gt: '>', amp: '&', quot: '"', apos: "'", nbsp: ' ' }[named.toLowerCase()];
});
const attributes = text => Object.fromEntries([...text.matchAll(/([\w:]+)\s*=\s*["']([^"']*)["']/g)].map(m => [m[1], decode(m[2])]));
const identity = (provider, id, version) => `${provider.toLowerCase()}:${Number(id)}:${version === null ? '*' : Number(version)}`;
const providerNames = ['Microsoft-Windows-Eventlog', 'Microsoft-Windows-PowerShell', 'Microsoft-Windows-Security-Auditing', 'Microsoft-Windows-TaskScheduler', 'Microsoft-Windows-WMI-Activity', 'PowerShellCore'];

function sourceIdentities(source, bytes) {
  const result = new Set();
  if (!['powershell_manifest','event_document','provider_article','sysmon_manifest','sysmon_document','event_catalog'].includes(source.kind)) return result;
  const raw = utf8(bytes);
  if (source.kind === 'sysmon_manifest') {
    for (const event of raw.matchAll(/<event\s+([^>]+)>/g)) {
      const a = attributes(event[1]);
      if (a.target === 'linux') continue;
      assert(['all','windows'].includes(a.target));
      assert(/^(?:\d+|0x[\da-f]+)$/i.test(a.value) && /^\d+$/.test(a.version));
      result.add(identity('Microsoft-Windows-Sysmon',a.value,a.version));
    }
    return result;
  }
  if (source.kind === 'sysmon_document') {
    for (const h of raw.matchAll(/<h3\b[^>]*>\s*Event ID\s+(\d+):/g)) result.add(identity('Microsoft-Windows-Sysmon',h[1],null));
    return result;
  }
  if (source.kind === 'event_catalog') {
    for (const m of raw.matchAll(/<(pre|blockquote|ul)\b[^>]*>([\s\S]*?)<\/\1>/gi)) {
      const text=decode(m[2].replace(/<br\s*\/?\s*>|<\/(?:p|li)>/gi,'\n').replace(/<[^>]*>/g,''));
      if (text.includes('<QueryList>')) {
        for (const selector of text.matchAll(/<Select\s+Path="(Security|System|Application|Setup)">([\s\S]*?)<\/Select>/g)) {
          const expr=decode(selector[2]).trim().match(/^\*\[System\[Provider\[@Name=['"]([^'"]+)['"]\] and \(([^()]*)\)\]\]$/);
          if(!expr) continue;
          const range=expr[2].trim().match(/^EventID\s*>=\s*(\d+)\s+and\s+EventID\s*<=\s*(\d+)$/);
          if(range) { for(let id=Number(range[1]);id<=Number(range[2]);id++) result.add(identity(expr[1],id,null)); }
          else for(const equality of expr[2].split(' or ')) {
            const id=equality.trim().match(/^EventID\s*=\s*(\d+)$/);assert(id);result.add(identity(expr[1],id[1],null));
          }
        }
        continue;
      }
      for(const line of text.split('\n')) {
        const item=line.match(/^(?:Information|Error|Warning)\s+\S+\s+(\d+)\s+(Microsoft-Windows-[A-Za-z0-9-]+)\s+N\/A\s+/);
        if(item) result.add(identity(item[2],item[1],null));
      }
      for(const block of text.split(/(?=^\s*Log name:)/im)) {
        const names=[...block.matchAll(/^\s*Source:\s*([^\r\n]+)/gim)];
        const ids=[...block.matchAll(/^\s*(?:Event ID|ID):\s*(\d+)\s*$/gim)];
        if(names.length!==1 || ids.length!==1) continue;
        const version=block.match(/^\s*Event Version:\s*(\d+)\s*$/im)?.[1]??null;
        result.add(identity(names[0][1].trim(),ids[0][1],version));
      }
    }
    return result;
  }
  if (source.kind === 'powershell_manifest') {
    for (const p of raw.matchAll(/<provider\b([^>]*)>([\s\S]*?)<\/provider>/g)) {
      if (!/<event\b/.test(p[2])) continue;
      const provider = attributes(p[1]).name;
      assert(providerNames.includes(provider));
      for (const e of p[2].matchAll(/<event\b([^>]*)\/>/g)) {
        const a = attributes(e[1]);
        result.add(identity(provider, a.value, a.version ?? null));
      }
    }
    assert(result.size > 0);
    return result;
  }
  if (!['event_document', 'provider_article'].includes(source.kind)) return result;
  const body = raw.match(/<main\b[^>]*>([\s\S]*?)<\/main>/i)?.[1] ?? raw;
  const text = decode(body.replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '').replace(/<[^>]*>/g, '\n'));
  const pageTuples = new Map();
  for (const pre of body.matchAll(/<pre\b[^>]*>([\s\S]*?)<\/pre>/gi)) {
    const xml = decode(pre[1].replace(/<[^>]*>/g, ''));
    const p = xml.match(/<Provider\b([^>]*)\/?\s*>/);
    const event = xml.match(/<EventID\b[^>]*>([0-9]+)<\/EventID>/);
    if (!p || !event) continue;
    const provider = attributes(p[1]).Name;
    if (!providerNames.includes(provider)) continue;
    const version = xml.match(/<Version>([0-9]+)<\/Version>/)?.[1] ?? null;
    result.add(identity(provider, event[1], version));
    pageTuples.set(Number(event[1]), provider);
  }
  if (source.kind === 'event_document' && pageTuples.size === 1) {
    const [id, provider] = [...pageTuples][0];
    const section = text.split('Event Versions:')[1]?.split('Field Descriptions:')[0] ?? '';
    for (const match of section.matchAll(/(?:^|\n)\s*([0-9]{1,3})(?:\s*[-–—:.]|\s*(?=\n|$))/g)) {
      result.add(identity(provider, id, match[1]));
    }
  }
  if (source.kind === 'provider_article') {
    if ([...body.matchAll(/<th\b[^>]*>([\s\S]*?)<\/th>/gi)].some(h => /^(?:current|legacy)windowseventid$/.test(decode(h[1].replace(/<[^>]*>/g, '')).replace(/\s/g, '').toLowerCase()))) return result;
    let providers = providerNames.filter(p => text.includes(p));
    if (source.url.includes('/about_logging?') && /a0c1853b-5c40-4b15-8766-3cf1c58f985a/i.test(text)) providers = ['Microsoft-Windows-PowerShell'];
    if (source.url.includes('/about_logging_windows?') && /f90714a8-5509-434a-bf6d-b1624c8a19a2/i.test(text)) providers = ['PowerShellCore'];
    if (providers.length === 1) {
      for (const match of text.matchAll(/Event\s*ID\s*:?\s*([0-9]{1,5})/gi)) {
        if (!pageTuples.has(Number(match[1]))) result.add(identity(providers[0], match[1], null));
      }
      for (const table of body.matchAll(/<table\b[^>]*>([\s\S]*?)<\/table>/gi)) {
        let column = -1;
        for (const row of table[1].matchAll(/<tr\b[^>]*>([\s\S]*?)<\/tr>/gi)) {
          const cells = [...row[1].matchAll(/<(td|th)\b[^>]*>([\s\S]*?)<\/\1>/gi)].map(c => ({ header: c[1] === 'th', value: decode(c[2].replace(/<[^>]*>/g, '')).trim() }));
          let value;
          for (let i=0; i<cells.length; i++) {
            if (/^event\s*id$/i.test(cells[i].value)) {
              if (cells[i].header) column=i; else value=cells[i+1]?.value;
            }
          }
          if (column>=0 && !cells.some(c=>c.header)) value=cells[column]?.value;
          const id=value?.match(/^([0-9]+)(?:\s|\/|$)/)?.[1];
          if (id && !pageTuples.has(Number(id))) result.add(identity(providers[0],id,null));
        }
      }
    }
  }
  return result;
}

export function witnessWindowsSources(archive, payload, report) {
  assert.equal(sha(archive), report.evidence_sha256);
  assert.equal(sha(payload), report.payload_file_sha256);
  const members = readEvidenceTar(archive);
  const lockBytes = members.get('source-lock.json');
  assert.equal(sha(lockBytes), report.source_lock_sha256);
  const lock = JSON.parse(utf8(lockBytes));
  assert.equal(sha(members.get('inventory.json')), lock.inventory_sha256);
  const evidence = JSON.parse(utf8(members.get('evidence.json')));
  const expected = new Set();
  const corpus = [];
  for (const item of lock.sources) {
    const bytes = members.get(item.evidence.path);
    assert.equal(bytes.length, item.evidence.size_bytes);
    assert.equal(sha(bytes), item.evidence.sha256);
    const csv = item.source.kind === 'combined_csv' ? witnessCombinedCSV(item.source, bytes, lock, members, evidence, report) : null;
    if(csv) corpus.push(csv.counts);
    const identities = csv ? csv.identities : sourceIdentities(item.source, bytes);
    for (const id of identities) expected.add(id);
    const recorded = new Set(evidence.records.filter(r => r.source_id === item.source.id).map(r => identity(r.provider_name, r.event_id, r.event_version)));
    assert.deepEqual([...recorded].sort(), [...identities].sort(), `source accounting: ${item.source.id}`);
    if (item.source.kind === 'provider_article') {
      const raw = utf8(bytes);
      const body = raw.match(/<main\b[^>]*>([\s\S]*?)<\/main>/i)?.[1] ?? raw;
      const text = decode(body.replace(/<[^>]*>/g, ' '));
      const named = providerNames.filter(p => text.includes(p));
      const guidSelector = /\/about_logging(?:_windows)?\?/.test(item.source.url)
        && /a0c1853b-5c40-4b15-8766-3cf1c58f985a|f90714a8-5509-434a-bf6d-b1624c8a19a2/i.test(text);
      const monitoringTable = [...body.matchAll(/<th\b[^>]*>([\s\S]*?)<\/th>/gi)].some(h => /^(?:current|legacy)windowseventid$/.test(decode(h[1].replace(/<[^>]*>/g, '')).replace(/\s/g, '').toLowerCase()));
      if ((named.length !== 1 && !guidSelector) || monitoringTable) {
        const excludedRows = [];
        let ti = 0;
        for (const table of body.matchAll(/<table\b[^>]*>([\s\S]*?)<\/table>/gi)) {
          let eventTable = false;
          let ri = 0;
          for (const row of table[1].matchAll(/<tr\b[^>]*>([\s\S]*?)<\/tr>/gi)) {
            const headers = [...row[1].matchAll(/<th\b[^>]*>([\s\S]*?)<\/th>/gi)];
            if (headers.length) {
              eventTable ||= headers.some(h => ['eventid', 'currentwindowseventid', 'legacywindowseventid'].includes(decode(h[1].replace(/<[^>]*>/g, '')).replace(/\s/g, '').toLowerCase()));
            } else if (eventTable && /<td\b/i.test(row[1])) {
              excludedRows.push(`html/table/${ti}/row/${ri}`);
            }
            ri++;
          }
          ti++;
        }
        const dispositions = report.dispositions.filter(d => d.source_id === item.source.id && d.reason === 'provider_not_stated_or_ambiguous_in_event_table');
        assert(dispositions.every(d => d.disposition === 'excluded' && d.entry_id === null));
        assert.deepEqual(dispositions.map(d => d.locator).sort(), excludedRows.sort(), `unqualified table accounting: ${item.source.id}`);
      }
    }
  }
  assert.equal(payload.at(-1), 10);
  const rows = utf8(payload).slice(0, -1).split('\n').map(line => JSON.parse(line));
  assert.deepEqual(rows.map(r => r.entry_id), [...expected].sort());
  for (const row of rows) {
    assert.equal(row.entry_id, identity(row.provider_name, row.event_id, row.event_version));
    for (const ref of row.source_refs) {
      assert.equal(ref.artifact_index, 0);
      const index = ref.locator.match(/^evidence\.json#\/(records|notes)\/([0-9]+)$/);
      assert(index);
      const record = evidence[index[1]][Number(index[2])];
      assert(record);
      assert.equal(index[1] === 'records' ? identity(record.provider_name, record.event_id, record.event_version) : record.entry_id, row.entry_id);
    }
    for(const block of row.documentation) {
      for(const ref of block.source_refs) {
        assert.equal(ref.artifact_index,0);
        const match=ref.locator.match(/^evidence\.json#\/(records|notes)\/(\d+)(?:\/documentation\/(\d+))?$/);assert(match);
        const record=evidence[match[1]][Number(match[2])];assert(record);
        if(match[3]!==undefined){const doc=record.documentation[Number(match[3])];assert(doc);assert.equal(block.kind,doc.kind);assert.equal(block.scope,doc.scope);assert.deepEqual(block.content,presentation(doc.content));}
      }
      assert.deepEqual(block.content,presentation(block.content),'canonical documentation has trailing empty space');
    }
  }
  assert.equal(rows.length, report.entry_count);
  return { schema_id: 'cartulary.windows_event_independent_source_witness.v2', implementation: 'node_crypto_independent_source_identity_scan', source_lock_sha256: report.source_lock_sha256, evidence_sha256: report.evidence_sha256, payload_file_sha256: report.payload_file_sha256, source_count: lock.sources.length, entry_count: rows.length, complete_source_identity_accounting: true, corpus, candidate_status: report.status??'accepted' };
}
