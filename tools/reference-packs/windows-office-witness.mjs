// Independent CSV/OOXML source witness. No producer extraction code is imported.
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { inflateRawSync } from 'node:zlib';

const sha = bytes => createHash('sha256').update(bytes).digest('hex');
const utf8 = bytes => new TextDecoder('utf-8', { fatal: true }).decode(bytes);
const entities = text => text.replace(/&#x([0-9a-f]+);|&#([0-9]+);|&(amp|lt|gt|quot|apos);/gi, (_, hex, dec, name) => hex || dec ? String.fromCodePoint(Number.parseInt(hex || dec, hex ? 16 : 10)) : ({amp:'&',lt:'<',gt:'>',quot:'"',apos:"'"})[name]);
const local = name => name.split(':').at(-1);

export function officeMembers(bytes) {
  let end = bytes.length - 22;
  while (end >= Math.max(0, bytes.length - 65557) && bytes.readUInt32LE(end) !== 0x06054b50) end--;
  assert(end >= 0, 'ZIP directory missing');
  assert.equal(bytes.readUInt16LE(end + 4), 0); assert.equal(bytes.readUInt16LE(end + 6), 0);
  const count = bytes.readUInt16LE(end + 10); assert(count <= 2048);
  let pos = bytes.readUInt32LE(end + 16); let total = 0;
  const out = new Map();
  for (let i=0; i<count; i++) {
    assert.equal(bytes.readUInt32LE(pos), 0x02014b50);
    const flags=bytes.readUInt16LE(pos+8), method=bytes.readUInt16LE(pos+10);
    const packed=bytes.readUInt32LE(pos+20), size=bytes.readUInt32LE(pos+24);
    const nameLen=bytes.readUInt16LE(pos+28), extra=bytes.readUInt16LE(pos+30), comment=bytes.readUInt16LE(pos+32);
    const name=utf8(bytes.subarray(pos+46,pos+46+nameLen));
    assert(!out.has(name) && !name.startsWith('/') && !name.includes('\\') && name.split('/').every(p => p && p!=='.' && p!=='..'));
    total+=size; assert(!(flags&1) && size <= 67108864 && total <= 268435456);
    const at=bytes.readUInt32LE(pos+42); assert.equal(bytes.readUInt32LE(at),0x04034b50);
    const start=at+30+bytes.readUInt16LE(at+26)+bytes.readUInt16LE(at+28); assert(start+packed <= pos);
    const compressed=bytes.subarray(start,start+packed);
    assert(method===0 || method===8);
    const raw=method===0 ? compressed : inflateRawSync(compressed,{maxOutputLength:67108864}); assert.equal(raw.length,size);
    out.set(name,raw); pos+=46+nameLen+extra+comment;
  }
  return out;
}
export function xmlTree(raw) {
  const text=typeof raw==='string'?raw:utf8(raw); assert(!/<!DOCTYPE|<!ENTITY/i.test(text));
  const stack=[]; let root; let last=0;
  const tokens=/<!--[\s\S]*?-->|<\?[\s\S]*?\?>|<!\[CDATA\[[\s\S]*?\]\]>|<\/[\w:.-]+\s*>|<[\w:.-]+(?:\s+[^<>]*?)?\/?\s*>|[^<]+/g;
  for (const m of text.matchAll(tokens)) {
    assert.equal(m.index,last,'unparsed XML'); last=m.index+m[0].length;
    const token=m[0];
    if (token.startsWith('<?') || token.startsWith('<!--')) continue;
    if (token.startsWith('</')) {
      const node=stack.pop(); assert(node && node.full===token.slice(2,-1).trim()); node.raw=text.slice(node.start,last); continue;
    }
    if (token.startsWith('<') && !token.startsWith('<![CDATA[')) {
      const full=token.match(/^<([\w:.-]+)/)[1]; const attrs={};
      for (const a of token.matchAll(/([\w:.-]+)\s*=\s*(?:"([^"]*)"|'([^']*)')/g)) {const key=local(a[1]); assert(!(key in attrs)); attrs[key]=entities(a[2]??a[3]);}
      const node={name:local(full),full,attrs,children:[],text:'',start:m.index,raw:''};
      if (stack.length) stack.at(-1).children.push(node); else {assert(!root);root=node;}
      if (/\/\s*>$/.test(token)) node.raw=token; else stack.push(node);
      assert(stack.length<=128); continue;
    }
    if (stack.length) stack.at(-1).text+=token.startsWith('<![CDATA[')?token.slice(9,-3).replace(/\r\n?/g,'\n'):entities(token.replace(/\r\n?/g,'\n')); else assert(!token.trim());
  }
  assert.equal(last,text.length); assert.equal(stack.length,0); assert(root); return root;
}
const children=(node,name)=>node.children.filter(c=>c.name===name);
const one=(node,name)=>children(node,name)[0];
function descendants(node,name) {return [...(node.name===name?[node]:[]),...node.children.flatMap(c=>descendants(c,name))];}
export function wordText(node) {
  if(node.name==='t') return node.text;
  if(node.name==='tab') return '\t';
  if(['br','cr'].includes(node.name)) return '\n';
  return node.children.map(wordText).join('')+(node.name==='p'?'\n':'');
}
export function csvRecords(bytes) {
  const text=utf8(bytes); assert(!text.startsWith('\ufeff'));
  const rows=[]; let fields=[],field='',quoted=false,closed=false,start=0;
  for(let i=0;i<text.length;i++) {
    const c=text[i];
    if(quoted) {if(c==='"'){if(text[i+1]==='"'){field+='"';i++;}else{quoted=false;closed=true;}}else field+=c;continue;}
    if(c==='"'){assert(!field && !closed);quoted=true;continue;}
    if(c===',' || c==='\n' || c==='\r' && text[i+1]==='\n') {
      fields.push(field.replace(/\r\n/g,'\n'));field='';closed=false;
      if(c!==','){if(c==='\r')i++; rows.push({fields,raw:text.slice(start,i+1)});fields=[];start=i+1;}continue;
    }
    assert(!closed);field+=c;
  }
  assert(!quoted);
  if(start<text.length){fields.push(field);rows.push({fields,raw:text.slice(start)});}
  const header=rows.shift().fields;assert.equal(new Set(header).size,header.length);
  return rows.map((r,i)=>{assert.equal(r.fields.length,header.length);return {number:i+1,raw:r.raw,fields:Object.fromEntries(header.map((h,j)=>[h,r.fields[j]]))};});
}
function workbook(members) {
  const book=xmlTree(members.get('xl/workbook.xml')), rels=xmlTree(members.get('xl/_rels/workbook.xml.rels'));
  const paths=new Map(children(rels,'Relationship').filter(r=>r.attrs.TargetMode!=='External').map(r=>[r.attrs.Id,r.attrs.Target.startsWith('/')?r.attrs.Target.slice(1):`xl/${r.attrs.Target}`]));
  const shared=members.has('xl/sharedStrings.xml')?children(xmlTree(members.get('xl/sharedStrings.xml')),'si').map(wordText):[];
  const sheets=new Map();
  for(const s of descendants(book,'sheet')) {
    const part=paths.get(s.attrs.id), sheet=xmlTree(members.get(part)), cells=new Map(), rows=new Map();
    for(const r of descendants(sheet,'row')) {rows.set(Number(r.attrs.r),r.raw);for(const c of children(r,'c')) {assert(!one(c,'f'));let value=one(c,'v')?.text??'';if(c.attrs.t==='s')value=shared[Number(value)];if(c.attrs.t==='inlineStr')value=wordText(one(c,'is'));cells.set(c.attrs.r,value);}}
    sheets.set(s.attrs.name,{part,cells,rows});
  }
  return sheets;
}
const values=(row,key)=>JSON.parse(row[key]||'[]');
const comparable=s=>s.split(/\s+/u).filter(Boolean).join(' ');
function referenceText(nodes) {
  return nodes.flatMap(n=>n.name!=='tbl'?[wordText(n).trim()]:children(n,'tr').map(row=>children(row,'tc').map(cell=>children(cell,'p').map(wordText).map(s=>s.trim()).join(' / ')).join(' | '))).join('\n');
}
export function witnessCombinedCSV(source, bytes, lock, members, evidence, report) {
  const originals=new Map(lock.sources.filter(s=>['original_docx','original_xlsx'].includes(s.source.kind)).map(s=>[s.source.kind,s]));
  assert(originals.size>0,'retained original evidence required');
  const wordSource=originals.get('original_docx'), bookSource=originals.get('original_xlsx');
  const word=wordSource?officeMembers(members.get(wordSource.evidence.path)):new Map();
  const body=wordSource?one(xmlTree(word.get('word/document.xml')),'body').children:[];
  const headingStyle=n=>one(one(n,'pPr')??{children:[]},'pStyle')?.attrs.val;
  const eventSections=body.flatMap((n,i)=>{
    const match=wordText(n).trim().match(/^(\d{1,5})(?:\([^)]*\))?:/);
    if(headingStyle(n)!=='Heading3' || !match)return [];
    let end=i+1;while(end<body.length && !['Heading1','Heading2','Heading3'].includes(headingStyle(body[end])))end++;
    return [{event:Number(match[1]),nodes:body.slice(i,end)}];
  });
  const sheets=bookSource?workbook(officeMembers(members.get(bookSource.evidence.path))):new Map();
  const rows=csvRecords(bytes), expected=new Set(), claimed=new Set(), messages=new Set(), catalogs=new Set(), sections=new Set();
  const editorial=JSON.parse(utf8(members.get('editorial.json')));
  const bindings=editorial.provider_bindings.filter(b=>b.csv_source_id===source.id);
  assert.equal(new Set(bindings.map(b=>b.csv_locator)).size,bindings.length);
  const usedBindings=new Set();
  const fragment=(start,end)=>{assert(start>=1 && end>=start && end<=body.length);return body.slice(start-1,end);};
  for(const row of rows) {
    const r=row.fields, loc=`csv/record/${row.number}`, event=Number(r.event_id), version=Number(r.event_version);
    const dispositions=report.dispositions.filter(d=>d.source_id===source.id && d.locator===loc);assert.equal(dispositions.length,1);
    for(const meta of values(r,'source_artifact_metadata_json')) assert.equal(meta.sha256,meta.filename.endsWith('.xlsx')?bookSource?.evidence.sha256:wordSource?.evidence.sha256);
    for(const [name,idColumn] of [['Complete Event Messages','A'],['Security Audit Events','C']]) {
      const sheet=sheets.get(name);
      if(!sheet)continue;
      for(const n of sheet.rows.keys()) {
        if(sheet.cells.get(`${idColumn}${n}`)!==r.event_id)continue;
        if(name==='Complete Event Messages') {if(sheet.cells.get(`B${n}`)!==r.event_version)continue;messages.add(n);assert.equal(sheet.cells.get(`C${n}`)??'',r.xlsx_message_template);assert.equal(sheet.cells.get(`D${n}`)??'',r.xlsx_minimum_os_for_message_version);} else catalogs.add(n);
      }
    }
    for(const section of values(r,'docx_section_contexts_json')) {sections.add(section.source.body_child_start);assert(wordText(fragment(section.source.body_child_start,section.source.body_child_start)[0]).trim()===section.heading.trim());}
    for(const f of values(r,'docx_field_definitions_json')) {
      const nodes=fragment(f.source.body_child_start,f.source.body_child_end), plain=comparable(nodes.map(wordText).join('')), formatted=comparable(referenceText(nodes));
      for(const key of ['label','type_as_documented','definition'])assert(plain.includes(comparable(f[key])),`original field ${key}`);
      for(const text of f.continuation_text)assert(formatted.includes(comparable(text)),'original field continuation');
    }
    for(const g of values(r,'docx_subcategory_guidance_json'))assert(comparable(referenceText(fragment(g.source.body_child_start,g.source.body_child_end))).includes(comparable(g.text)),'original subcategory guidance');
    for(const table of values(r,'docx_reference_tables_json')) {
      const tables=fragment(table.source.body_child_start,table.source.body_child_end).flatMap(n=>descendants(n,'tbl'));assert.equal(tables.length,1);
      const original=children(tables[0],'tr').map(row=>children(row,'tc').map(cell=>{const props=one(cell,'tcPr')??{children:[]}, span=one(props,'gridSpan'), merge=one(props,'vMerge');return {text:comparable(wordText(cell)),grid_span:span?Number(span.attrs.val):1,vertical_merge:merge?(merge.attrs.val||'continue'):null};}));
      assert.deepEqual(original,table.rows.map(row=>row.map(c=>({...c,text:comparable(c.text)}))),'original table structure');
    }
    const providers=new Set();
    // Original Word evidence stands on its own even when the derivative omitted
    // its sample citation or there is no corresponding worksheet row.
    for(const section of eventSections.filter(s=>s.event===event)) {
      const systems=section.nodes.map(wordText).join('').match(/<System>[\s\S]*?<\/System>/g);
      if(systems?.length!==1)continue;
      let system;try{system=xmlTree(systems[0]);}catch{continue;}
      if(Number(one(system,'EventID')?.text)!==event)continue;
      const provider=one(system,'Provider')?.attrs.Name;if(provider)providers.add(provider);
    }
    for(const a of [...values(r,'docx_sample_metadata_for_row_version_json'),...values(r,'docx_other_sample_metadata_json')]) {
      const text=fragment(a.source.body_child_start,a.source.body_child_end).map(wordText).join('');
      const system=text.match(/<System>[\s\S]*?<\/System>/g);if(system?.length!==1)continue;
      let x;try{x=xmlTree(system[0]);}catch{continue;}
      for(const [key,name] of Object.entries({event_id:'EventID',event_version:'Version',level:'Level',task:'Task',opcode:'Opcode',channel:'Channel',keywords:'Keywords'})) if(a[key]!==null && a[key]!==undefined) assert.equal(one(x,name)?.text.trim(),String(a[key]),`original sample ${key}`);
      for(const [key,name] of Object.entries({provider_name:'Name',provider_guid:'Guid'})) if(a[key]!==null && a[key]!==undefined) assert.equal(one(x,'Provider')?.attrs[name]?.toLowerCase(),a[key].toLowerCase(),`original sample ${key}`);
      if(Number(one(x,'EventID')?.text)!==event)continue;
      const provider=one(x,'Provider')?.attrs.Name;if(provider)providers.add(provider);
    }
    if(r.record_kind==='source_boundary_marker') {assert.equal(event,8191);assert.equal(r.xlsx_message_template.trim(),'Highest System-Defined Audit Message Value.');assert.equal(dispositions[0].disposition,'excluded');continue;}
    for(const binding of bindings.filter(b=>b.csv_locator===loc)) {
      assert.equal(binding.csv_fragment_sha256,sha(Buffer.from(row.raw)));
      const admitted=lock.sources.find(s=>s.source.id===binding.source_id);assert.equal(admitted?.source.kind,'event_document');
      const raw=members.get(admitted.evidence.path);assert.equal(sha(raw),binding.source_sha256);
      const html=utf8(raw), decodeHTML=s=>entities(s.replace(/<[^>]*>/g,''));
      const index=binding.locator.match(/^html\/pre\/(\d+)\/text\(\)$/);assert(index);
      const xml=decodeHTML([...html.matchAll(/<pre\b[^>]*>([\s\S]*?)<\/pre>/gi)][Number(index[1])][1])+'\n';
      assert.equal(sha(Buffer.from(xml)),binding.fragment_sha256);
      const systems=xml.match(/<System>[\s\S]*?<\/System>/g);assert.equal(systems?.length,1);const system=xmlTree(systems[0]);
      assert.equal(one(system,'EventID')?.text.trim(),r.event_id);assert.equal(one(system,'Version')?.text.trim(),r.event_version);
      const provider=one(system,'Provider')?.attrs.Name;assert(provider);providers.add(provider);
      const recovered=evidence.records.find(x=>x.source_id===source.id && x.locator===loc);assert(recovered);
      const guid=one(system,'Provider')?.attrs.Guid;
      if(guid)assert.equal(recovered.provider_guid,guid.replace(/[{}]/g,'').toLowerCase());
      for(const [locator,digest,role] of [[binding.locator,binding.fragment_sha256,'reviewed_provider_system'],['document',binding.source_sha256,'reviewed_provider_document']]) {
        assert(evidence.assertions.some(a=>a.source_id===binding.source_id && a.locator===locator && a.fragment_sha256===digest && a.role===role && a.csv_records.includes(row.number)));
      }
      usedBindings.add(loc);
    }
    if(providers.size!==1) {
      const exclusions=(editorial.provider_exclusions??[]).filter(x=>x.source_id===source.id && x.locator===loc);
      if(exclusions.length) {
        assert.equal(exclusions.length,1);assert.equal(exclusions[0].fragment_sha256,sha(Buffer.from(row.raw)));
        assert.equal(dispositions[0].disposition,'excluded');assert.equal(dispositions[0].reason,'user_excluded_provider_unresolved');
      } else {assert.equal(dispositions[0].disposition,'blocking');assert.equal(dispositions[0].reason,'provider_evidence_missing');}
      continue;
    }
    const provider=[...providers][0];const id=`${provider.toLowerCase()}:${event}:${version}`;expected.add(id);
    const record=evidence.records.find(x=>x.source_id===source.id && x.locator===loc);assert(record);assert.equal(record.fragment_sha256,sha(Buffer.from(row.raw)));assert.equal(record.provider_name,provider);assert.equal(record.event_id,event);assert.equal(record.event_version,version);claimed.add(loc);
    for(const block of record.documentation) {
      const match=block.locator.match(/^csv\/record\/(\d+)\/field\/([a-z_]+)(?:\/(\d+))?$/);assert(match && Number(match[1])===row.number);
      const column=match[2], index=match[3]===undefined?null:Number(match[3]);assert(column in r);
      if(index===null && ['title','message_template','explanation','field_reference'].includes(block.kind))assert.equal(block.content.text,r[column]);
      if(index!==null) {
        const original=values(r,column)[index];assert(original);
        if(['field_definition','reference_table','source_link','untranscribed_image','source_metadata'].includes(block.kind))for(const [key,value] of Object.entries(block.content))assert.deepEqual(value,original[key]??null,`${column}/${key}`);
      }
    }
    const arrayKinds={'docx_field_definitions_json':'field_definition','docx_reference_tables_json':'reference_table','docx_links_json':'source_link','docx_image_references_json':'untranscribed_image','source_artifact_metadata_json':'source_metadata'};
    for(const [column,kind] of Object.entries(arrayKinds)) assert.equal(record.documentation.filter(b=>b.kind===kind).length,values(r,column).length,`complete ${column}`);
    for(const [column,kind] of [['xlsx_message_template','message_template'],['docx_message_template','message_template'],['docx_event_field_reference','field_reference'],['docx_event_description','explanation']])if(r[column])assert(record.documentation.some(b=>b.kind===kind && b.content.text===r[column]));
  }
  for(const [name,set] of [['Complete Event Messages',messages],['Security Audit Events',catalogs]]) assert.deepEqual([...set].sort((a,b)=>a-b),[...(sheets.get(name)?.rows.keys()??[])].filter(n=>n!==1).sort((a,b)=>a-b));
  const originalSections=body.map((n,i)=>({n,i})).filter(({n})=>one(one(n,'pPr')??{children:[]},'pStyle')?.attrs.val==='Heading3' && /^\s*\d{1,5}(?:\([^\r\n)]{0,32}\))?:/.test(wordText(n))).map(({i})=>i+1);
  assert.deepEqual([...sections].sort((a,b)=>a-b),originalSections);
  for(const binding of evidence.assertions??[]) {
    let raw;
    if(binding.source_id===wordSource?.source.id) {const match=binding.locator.match(/^ooxml\/word\/document\.xml\/body\/(\d+)\/(\d+)$/);assert(match);raw=fragment(Number(match[1]),Number(match[2])).map(n=>n.raw).join('');}
    else if(binding.source_id===bookSource?.source.id){const match=binding.locator.match(/^ooxml\/(.+)\/row\/(\d+)$/);assert(match);raw=[...sheets.values()].find(s=>s.part===match[1]).rows.get(Number(match[2]));}
    else continue;
    assert.equal(sha(Buffer.from(raw)),binding.fragment_sha256);assert(binding.csv_records.every(n=>n>=1 && n<=rows.length));
  }
  assert.equal(evidence.records.filter(r=>r.source_id===source.id).length,claimed.size);
  assert.equal(usedBindings.size,bindings.length);
  const exclusions=(editorial.provider_exclusions??[]).filter(x=>x.source_id===source.id);
  const excluded=report.dispositions.filter(d=>d.source_id===source.id && d.reason==='user_excluded_provider_unresolved');
  assert.equal(excluded.length,exclusions.length);assert.deepEqual(excluded.map(d=>d.locator).sort(),exclusions.map(d=>d.locator).sort());
  return {identities:expected,counts:{csv_records:rows.length,message_rows:messages.size,catalog_rows:catalogs.size,word_sections:sections.size,admitted_csv_records:claimed.size,excluded_provider_records:excluded.length,reviewed_provider_recoveries:usedBindings.size}};
}
