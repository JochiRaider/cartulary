# One-time authored fixture revision. Inputs are the immutable predecessor
# expectations and CSV bytes. No application output or Go implementation is run.
import copy, csv, hashlib, io, ipaddress, json, pathlib, subprocess, functools
ROOT = pathlib.Path(__file__).resolve().parents[2]
BASE = 'e6ca992e634147fd257f8a7462026a80053cd1d8'
P = 'network_flow.'
RENAMES = {P+k:P+v for k,v in [('src_ip','input_first_ip'),('dst_ip','input_second_ip'),('src_port','input_first_port'),('dst_port','input_second_port')]}
VERSIONS = {f'cartulary.network_flow.{k}.v1':f'cartulary.network_flow.{k}.v2' for k in ['import_source_inspection_result','import_preview_result','mapping_candidate','mapping_approved','table_query_request','table_query_continuation_request','rejected_row_diagnostic','import_apply_result']}
VERSIONS.update({'cartulary.network_flow_table.v1':'cartulary.network_flow_table.v2','cisco_sna_netflow_csv_v1':'cisco_sna_netflow_csv_v2','cisco_sna_flow_search_csv_v1':'cisco_sna_flow_search_csv_v2'})
ORDER = ['flow_start_utc','flow_end_utc','endpoint_a_ip','endpoint_b_ip','endpoint_a_port','endpoint_b_port','ip_protocol','bytes_count','packets_count','exporter_id','input_interface','output_interface','tcp_flags','application_label','direction','direction_evidence','observation_source_ref']
def prior(path): return subprocess.check_output(['git','show',BASE+':'+str(path)],cwd=ROOT)
def oldjson(path): return json.loads(prior(path))
def sha(b): return hashlib.sha256(b).hexdigest()
def cjson(v,sort=False): return json.dumps(v,ensure_ascii=False,separators=(',',':'),sort_keys=sort).encode()
def parts(*vs): return b''.join(str(v).encode()+b'\0' for v in vs)
def edit(v):
 if isinstance(v,str):
  for a,b in VERSIONS.items():v=v.replace(a,b)
  return RENAMES.get(v,v)
 if isinstance(v,list):return [edit(x) for x in v]
 if isinstance(v,dict):
  x={RENAMES.get(k,k):edit(vv) for k,vv in v.items()}
  if 'target_table_schema_id' in x and 'source_profile_id' in x:
   x['endpoint_derivation_id']='cisco_flow_search_orientation_v1' if x['source_profile_id']=='cisco_sna_flow_search_csv_v2' else 'declared_source_destination_v1'
  if x.get('source_profile_id')=='cisco_sna_flow_search_csv_v2' and 'field_mappings' in x:
   for key,ordinal in [('input_first_orientation',20),('input_second_orientation',61)]:
    if not any(f.get('field_key')==P+key for f in x['field_mappings']):
     binding={'mapping_kind':'source_column','field_key':P+key,'source_column_ordinal':ordinal,'transform_id':'orientation_role_v1','empty_value_policy':'empty_string_is_null','combinability':'single_source_only'}
     position=next((i for i,f in enumerate(x['field_mappings']) if f['mapping_kind']=='system_derivation'),len(x['field_mappings']));x['field_mappings'].insert(position,binding)
  return x
 return v

def fingerprint(m,source_sha):
 data=parts('cartulary.network_flow_mapping_fingerprint.v2',m['target_kind'],m['target_table_schema_id'],m['source_profile_id'],m['endpoint_derivation_id'],m['parser_profile_id'],source_sha,m['unknown_column_policy'])
 t=m['timestamp_profile']; keys=['schema_id','mode','precision']
 if t['mode']=='rfc3339':keys+=['timezone','timezone_ruleset_id','ambiguous_local_time_policy','local_time_gap_policy']
 if t['mode']=='netflow_sys_uptime_milliseconds':keys+=['netflow_export_time_column_ordinal','netflow_export_time_mode','netflow_exporter_uptime_at_export_column_ordinal']
 data+=cjson({k:t.get(k) for k in keys})+b'\0'
 for col in sorted(m['source_columns'],key=lambda c:c['source_column_ordinal']):
  data+=parts(*(col[k] for k in ['source_column_ordinal','raw_header_text','normalized_header_for_suggestion','raw_header_sha256']))
 for f in sorted(m['field_mappings'],key=lambda f:(f.get('field_key',''),f['mapping_kind'],f.get('source_column_ordinal',0),f.get('derivation_id',''),f.get('ignore_reason',''))):
  keys={'source_column':['mapping_kind','field_key','source_column_ordinal','transform_id','empty_value_policy','combinability'],'ignored_source_column':['mapping_kind','source_column_ordinal','ignore_reason'],'system_derivation':['mapping_kind','field_key','derivation_id','combinability']}[f['mapping_kind']]
  data+=cjson({k:f[k] for k in keys})+b'\0'
 return sha(data)

def normalized_digest(fp,values,raw):
 data=parts('cartulary.network_flow.normalized_row_digest.v2',fp)
 for name in ORDER:
  key=P+name;data+=parts(key)
  if values.get(key) is None:data+=parts('n','null')
  else:data+=parts('p')+cjson(values[key],sort=True)+b'\0'
 # Retained raw is a typed ordinal map, not a lexical JSON map.
 ordered={k:{f:raw[k][f] for f in ['source_column_ordinal','raw_header_text','raw_header_sha256','decoded_value','decoded_value_sha256']} for k in sorted(raw,key=int)}
 return sha(data+parts('unmapped_raw')+cjson(ordered)+b'\0')

@functools.lru_cache(maxsize=32)
def csv_rows(source):return list(csv.reader(io.StringIO(source.decode("utf-8-sig"),newline="")))

def normalize_record(old,m,fp,source):
 result=edit(old);v=result['values']; first=(v.pop(P+'input_first_ip'),v.pop(P+'input_first_port'));second=(v.pop(P+'input_second_ip'),v.pop(P+'input_second_port'))
 def key(e):
  ip=ipaddress.ip_address(e[0]);return(ip.version,ip.packed,e[1] is not None,e[1] or 0)
 a,b=sorted([first,second],key=key);first_slot='a' if key(first)<=key(second) else 'b'
 v.update({P+'endpoint_a_ip':a[0],P+'endpoint_b_ip':b[0],P+'endpoint_a_port':a[1],P+'endpoint_b_port':b[1],P+'direction':{'kind':'source_to_destination','from_endpoint':first_slot}})
 rows=csv_rows(source);cells=rows[old['source_row_number']-1]
 bindings={f['field_key']:f['source_column_ordinal'] for f in m['field_mappings'] if f['mapping_kind']=='source_column'}
 inputs=[]
 for name in ['input_first_ip','input_first_port','input_first_orientation','input_second_ip','input_second_port','input_second_orientation']:
  if P+name not in bindings:continue
  ordinal=bindings[P+name];inputs.append({'input_key':P+name,'source_column_ordinal':ordinal,'raw_header_sha256':sha(rows[0][ordinal-1].encode()),'raw_value_sha256':sha(cells[ordinal-1].encode())})
 evidence={'derivation_id':m['endpoint_derivation_id'],'first_endpoint':first_slot,'first_role':'','second_role':'','inputs':inputs}
 if m['source_profile_id']=='cisco_sna_flow_search_csv_v2':
  roles=[cells[bindings[P+key]-1].strip(' ').lower() for key in ['input_first_orientation','input_second_orientation']]
  evidence['first_role'],evidence['second_role']=roles
  if all(roles):
   assert sorted(roles)==['client','server']
   origin=first_slot if roles[0]=='client' else ('b' if first_slot=='a' else 'a')
   if a==b:origin='a'
   v[P+'direction']={'kind':'client_to_server','from_endpoint':origin}
  else:v[P+'direction']={'kind':'unknown','reason':'incomplete_orientation' if any(roles) else 'missing_orientation'}
  for key in ['input_first_orientation','input_second_orientation']:result['unmapped_raw'].pop(str(bindings[P+key]),None)
 result['direction_evidence']=evidence
 digest_values=dict(v);digest_values[P+'direction_evidence']=evidence
 result['normalized_row_digest_sha256']=normalized_digest(fp,digest_values,result['unmapped_raw'])
 return result

def update_diagnostics(obj,source_sha,fp):
 if isinstance(obj,list):
  for x in obj:update_diagnostics(x,source_sha,fp)
 elif isinstance(obj,dict):
  if 'diagnostic_id' in obj:
   ordinal=obj.get('source_column_ordinal');field=obj.get('field_key')
   obj['diagnostic_id']='nfd_'+sha(parts('cartulary.network_flow.diagnostic_id.v2',source_sha,fp,obj['source_row_number'],'n' if ordinal is None else 'p:'+str(ordinal),'n' if field is None else 'p:'+field,obj['error_code'],obj['reason_code']))
  for x in obj.values():update_diagnostics(x,source_sha,fp)

def replace_scalar(obj,old,new):
 if isinstance(obj,str):return new if obj==old else obj
 if isinstance(obj,list):return [replace_scalar(x,old,new) for x in obj]
 if isinstance(obj,dict):return {k:replace_scalar(v,old,new) for k,v in obj.items()}
 return obj

def write(path,x):
 (ROOT/path).write_text(json.dumps(x,ensure_ascii=False,indent=2)+'\n')
def revise_mapping(identifier):
 directory=pathlib.Path('fixtures/network-flow')/identifier;slug=identifier.split('-',3)[3]
 baseline=oldjson(directory/'expected'/f'{slug}.expected.json'); result=edit(baseline);revision=baseline['freeze_revision']+1;result['freeze_revision']=revision
 for old,out in zip(baseline['outcomes'],result['outcomes']):
  if not isinstance(old['result'],dict) or old['result'].get('status')!='accepted' or 'prepared' not in old['result']:continue
  prepared=out['result']['prepared'];m=prepared['OwnerMapping'];source=prior(directory/old['request']['source']);fp=fingerprint(m,sha(source))
  out.update(replace_scalar(out,prepared['MappingFingerprint'],fp))
  prepared=out['result']['prepared']
  records=[]
  for record in old['result']['normalized_records']:
   records.append(normalize_record(record,m,fp,source) if record['status']=='accepted' else edit(record))
  out['result']['normalized_records']=records
  preview=prepared['OwnerResult'];preview['preview_resolved_direction_count']=preview['preview_accepted_count'];preview['preview_unknown_direction_count']=0
  update_diagnostics(out,sha(source),fp)
 write(directory/'source/cases.json',edit(oldjson(directory/'source/cases.json')))
 write(directory/'expected'/f'{slug}.expected.json',result);write(directory/'transcripts'/f'{slug}.transcript.json',result)
 if identifier.startswith('NF-FIX-011'):
  write(directory/'source/client-cases.json',edit(oldjson(directory/'source/client-cases.json')))
  client=edit(oldjson(directory/'expected'/f'{slug}.client.expected.json'));client['freeze_revision']=revision
  write(directory/'expected'/f'{slug}.client.expected.json',client);write(directory/'transcripts'/f'{slug}.client.transcript.json',client)
 update_manifest(directory,revision)

def update_manifest(directory,revision):
 m=oldjson(directory/'manifest.json');m['freeze']['revision']=revision
 for bucket in ['source_files','expected_artifacts','transcript_files']:
  for entry in m[bucket]:
   data=(ROOT/directory/entry['logical_path']).read_bytes();entry['size_bytes']=len(data);entry['sha256']=sha(data)
 def bundle(entries):return sha(b''.join(f"{e['logical_path']}\0{e['sha256']}\0{e['size_bytes']}\n".encode() for e in entries))
 m['source_bundle_sha256']=bundle(m['source_files']);m['expected_bundle_sha256']=bundle(m['expected_artifacts']+m['transcript_files']);write(directory/'manifest.json',m)

for identifier in ['NF-FIX-003-duplicate-headers','NF-FIX-011-alias-collision','NF-FIX-012-sys-uptime-timestamps','NF-FIX-022-timestamp-rulesets']:
 revise_mapping(identifier)
identifier='NF-FIX-010-json-admission';directory=pathlib.Path('fixtures/network-flow')/identifier
for path in (ROOT/directory).rglob('*.json'):
 if path.name=='manifest.json':continue
 relative=path.relative_to(ROOT);x=edit(oldjson(relative))
 if isinstance(x,dict) and 'freeze_revision' in x:x['freeze_revision']+=1
 write(relative,x)
update_manifest(directory,5)
# Strict JSON admission keeps malformed/duplicate/invalid UTF-8 bytes intact;
# only the version token in otherwise identical request bytes changes.
for path in (ROOT/directory/'source').glob('*.jsonl'):
 data=prior(path.relative_to(ROOT))
 for a,b in VERSIONS.items():data=data.replace(a.encode(),b.encode())
 path.write_bytes(data)
update_manifest(directory,5)

# Canonical transcript vectors are recomputed from the declared byte grammar.
import struct
identifier='NF-FIX-019-canonical-json-unicode';directory=pathlib.Path('fixtures/network-flow')/identifier
baseline=oldjson(directory/'expected/canonical-json-unicode.expected.json');result=edit(baseline);result['freeze_revision']=6
mappings=edit(oldjson(directory/'source/mapping-vectors.json'))
write(directory/'source/mapping-vectors.json',mappings)
for vector in mappings:
 out=next(o for o in result['outcomes'] if o['case']==vector['name']);out['request']=vector
 if 'approved_mapping' in vector:out['result']={'fingerprint':fingerprint(vector['approved_mapping'],vector['source_sha256'])}
 else:
  f=vector['field_mapping'];keys={'source_column':['mapping_kind','field_key','source_column_ordinal','transform_id','empty_value_policy','combinability'],'ignored_source_column':['mapping_kind','source_column_ordinal','ignore_reason'],'system_derivation':['mapping_kind','field_key','derivation_id','combinability']}[f['mapping_kind']]
  out['result']={'canonical':cjson({k:f[k] for k in keys}).decode()}
for out in result['outcomes']:
 if out['case'].startswith('digest-'):
  r=out['request'];out['result']=normalized_digest(r['fingerprint'],r['values'],r['unmapped'])
graphs=oldjson(directory/'source/graph-vectors.json')
for vector in graphs:
 s=vector['semantic'];s['schema_id']='cartulary.network_flow.graph_semantic_query.v3'
 for f in s['filters']:f['field_key']=f['field_key'].replace('network_flow.src_ip','network_flow.from_ip').replace('network_flow.dst_ip','network_flow.to_ip')
 encoded=cjson(s);out=next(o for o in result['outcomes'] if o['case']==vector['name']);out['request']=vector
 tr=s['time_range'] if any(s['time_range'].values()) else None
 fields=[b'cartulary.network_flow.graph_query_digest.v3',vector['incident_id'].encode(),cjson(sorted(s['selected_table_ids'])),cjson(s['filters']),cjson(tr),s['schema_id'].encode(),cjson(s['aggregation'])]
 graph_digest=sha(b''.join(struct.pack('>Q',len(f))+f for f in fields))
 mutation=parts('cartulary.network_flow.mutation_request_digest.v1','nf.graph_views.create','collection')+b'{"display_name":'+cjson(vector['display_name'])+b',"semantic_query":'+encoded+b'}\0'
 out['result']={'canonical':encoded.decode(),'sha256':sha(encoded),'graph_query_digest':graph_digest,'mutation_sha256':sha(mutation)}
write(directory/'source/graph-vectors.json',graphs)
write(directory/'expected/canonical-json-unicode.expected.json',result);write(directory/'transcripts/canonical-json-unicode.transcript.json',result);update_manifest(directory,6)
# Publication expectations retain the independently reviewed lifecycle and
# timestamp results; only the specified semantic fields and their identities
# are transformed. This deliberately does not use any captured actual.json.
VERSIONS.update({f'cartulary.network_flow.{k}.v1':f'cartulary.network_flow.{k}.v2' for k in ['approved_mapping','import_unit_result','table_query_result','rejected_rows_query_result']})
VERSIONS.update({f'cartulary.network_flow_{k}.v1':f'cartulary.network_flow_{k}.v2' for k in ['table_list','table_get_result','table_mutation_result']})
def walk(v):
 if isinstance(v,dict):
  yield v
  for x in v.values():yield from walk(x)
 elif isinstance(v,list):
  for x in v:yield from walk(x)

def raw_cells(source,row_number,m):
 rows=csv_rows(source);used={f['source_column_ordinal'] for f in m['field_mappings'] if f['mapping_kind']=='source_column'}
 return {str(i):{'source_column_ordinal':i,'raw_header_text':header,'raw_header_sha256':sha(header.encode()),'decoded_value':rows[row_number-1][i-1],'decoded_value_sha256':sha(rows[row_number-1][i-1].encode())} for i,header in enumerate(rows[0],1) if i not in used}

def revise_publication(identifier):
 directory=pathlib.Path('fixtures/network-flow')/identifier;slug=identifier.split('-',3)[3]
 baseline=oldjson(directory/'expected'/f'{slug}.expected.json');revision=baseline['freeze_revision']+1
 contexts={};changes={};normalized={}
 for o in walk(baseline):
  if not isinstance(o.get('preview'),dict) or 'owner_result' not in o['preview'] or 'source' not in o.get('request',{}):continue
  preview=o['preview']['owner_result'];m=edit(preview['materialized_mapping']);source=(ROOT/directory/o['request']['source']).read_bytes();oldfp=preview['mapping_fingerprint'];fp=fingerprint(m,sha(source));contexts[oldfp]=(m,fp,source);changes[oldfp]=fp
 for row in walk(baseline):
  if 'network_flow_row_id' not in row or 'src_ip' not in row or 'normalized_row_digest_sha256' not in row:continue
  m,fp,source=contexts[row['mapping_fingerprint']]
  values={P+k:row[k] for k in ['flow_start_utc','flow_end_utc','src_ip','dst_ip','src_port','dst_port','ip_protocol','bytes_count','packets_count','exporter_id','input_interface','output_interface','tcp_flags','application_label']}
  for k in ('flow_start_utc','flow_end_utc'):values[P+k]=values[P+k].replace('+00:00','Z')
  old={'status':'accepted','source_row_number':row['source_row_number'],'source_row_digest_sha256':row['source_row_digest_sha256'],'normalized_row_digest_sha256':row['normalized_row_digest_sha256'],'unmapped_raw':row.get('unmapped_raw',raw_cells(source,row['source_row_number'],m)),'values':values}
  n=normalize_record(old,m,fp,source);normalized[row['network_flow_row_id']]=n
  changes[row['normalized_row_digest_sha256']]=n['normalized_row_digest_sha256']
  changes[row['network_flow_row_id']]='nfr_'+sha(parts('cartulary.network_flow_row_id.v2',row['incident_id'],row['network_flow_table_id'],row['source_row_number'],row['source_row_digest_sha256'],n['normalized_row_digest_sha256']))
 result=edit(baseline);result['freeze_revision']=revision
 for obj in walk(result):
  if 'network_flow_row_id' in obj and obj['network_flow_row_id'] in normalized and 'normalized_row_digest_sha256' in obj:
   n=normalized[obj['network_flow_row_id']];v=n['values'];public=P+'input_first_ip' in obj
   if public:
    for key in ['input_first_ip','input_second_ip','input_first_port','input_second_port']:del obj[P+key]
    for slot in ['a','b']:obj[P+'endpoint_'+slot]={'ip':v[P+'endpoint_'+slot+'_ip'],'port':v[P+'endpoint_'+slot+'_port']}
    obj[P+'direction']=v[P+'direction'];obj[P+'direction_evidence']=n['direction_evidence']
   else:
    for key in ['src_ip','dst_ip','src_port','dst_port']:del obj[key]
    for key in ['endpoint_a_ip','endpoint_b_ip','endpoint_a_port','endpoint_b_port']:obj[key]=v[P+key]
    obj['direction_kind']=v[P+'direction']['kind'];obj['from_endpoint']=v[P+'direction'].get('from_endpoint','');obj['direction_reason']=v[P+'direction'].get('reason','');obj['direction_evidence']=n['direction_evidence']
   if 'unmapped_raw' in obj:obj['unmapped_raw']=n['unmapped_raw']
   if 'unmapped_raw_sha256' in obj:obj['unmapped_raw_sha256']=sha(cjson(n['unmapped_raw'],sort=True))
  if obj.get('schema_id')=='cartulary.network_flow.import_preview_result.v2':
   obj['preview_resolved_direction_count']=obj['preview_accepted_count'];obj['preview_unknown_direction_count']=0
  if 'owner_mapping' in obj and obj.get('extension_profile_id')=='network_flow_activity':obj['owner_mapping_schema_id']='cartulary.network_flow.approved_mapping.v2'
  if 'source_profile_id' in obj and 'created_at' in obj and 'table_status' in obj and obj['created_at'].endswith('Z'):
   obj['counter_semantics']='conversation_total' if obj['source_profile_id']=='cisco_sna_flow_search_csv_v2' else 'source_reported_flow_total'
  if 'diagnostic_id' in obj:
   # Diagnostic resources do not repeat source/fingerprint; table storage does.
   match=None
   if obj.get('mapping_fingerprint') in contexts:match=contexts[obj['mapping_fingerprint']]
   elif len(contexts)==1:match=next(iter(contexts.values()))
   else:
    for oldfp,(m,fp,source) in contexts.items():
     ordinal=obj.get('source_column_ordinal');field=obj.get('field_key')
     oldfield=next((k for k,v in RENAMES.items() if v==field),field)
     oldid='nfd_'+sha(parts('cartulary.network_flow.diagnostic_id.v1',sha(source),oldfp,obj['source_row_number'],'n' if ordinal is None else 'p:'+str(ordinal),'n' if oldfield is None else 'p:'+oldfield,obj['error_code'],obj['reason_code']))
     if obj['diagnostic_id']==oldid:match=(m,fp,source);break
   if match:
    before=obj['diagnostic_id'];update_diagnostics(obj,sha(match[2]),match[1]);changes[before]=obj['diagnostic_id']
 def replace(v):
  if isinstance(v,str):return changes.get(v,v)
  if isinstance(v,list):return [replace(x) for x in v]
  if isinstance(v,dict):return {k:replace(x) for k,x in v.items()}
  return v
 result=replace(result)
 for path in (ROOT/directory/'source').glob('*.json'):
  relative=path.relative_to(ROOT);write(relative,replace(edit(oldjson(relative))))
 write(directory/'expected'/f'{slug}.expected.json',result);write(directory/'transcripts'/f'{slug}.transcript.json',result);update_manifest(directory,revision)
for identifier in ['NF-FIX-013-filename-display','NF-FIX-020-atomic-import-commit','NF-FIX-023-import-facade-source-change','NF-FIX-027-retention-soft-delete']:
 revise_publication(identifier)

# The adopted cross-owner transaction contract closes proven-absent commits.
# Retain the independent pre-write failure oracle for this equivalent absence;
# committed lost replies still recover the durable success proof.
directory=pathlib.Path('fixtures/network-flow/NF-FIX-020-atomic-import-commit')
for relative in ['expected/atomic-import-commit.expected.json','transcripts/atomic-import-commit.transcript.json']:
 result=json.loads((ROOT/directory/relative).read_text())
 absent=copy.deepcopy(next(o for o in result['outcomes'] if o['case']=='before_owner_apply'))
 for i,o in enumerate(result['outcomes']):
  if o['case']=='before_transaction_commit':
   absent['case']=o['case']; absent['request']=o['request']; absent['request']['fault']['expect_retry']=False
   result['outcomes'][i]=absent
 write(directory/relative,result)
cases=json.loads((ROOT/directory/'source/cases.json').read_text())
for c in cases:
 if c['name']=='before_transaction_commit':c['fault']['expect_retry']=False
write(directory/'source/cases.json',cases);update_manifest(directory,4)
# Indicator selectors bind canonical endpoint slots, not derivation inputs.
directory=pathlib.Path('fixtures/network-flow/NF-FIX-027-retention-soft-delete')
for relative in ['expected/retention-soft-delete.expected.json','transcripts/retention-soft-delete.transcript.json']:
 result=json.loads((ROOT/directory/relative).read_text())
 for obj in walk(result):
  if obj.get('kind')=='row_field_value' and obj.get('field_key')=='network_flow.input_first_ip':obj['field_key']='network_flow.endpoint_a_ip'
 write(directory/relative,result)
update_manifest(directory,5)

# Storage evidence is ordered by the newly derived diagnostic identities.
for identifier in ['NF-FIX-013-filename-display','NF-FIX-020-atomic-import-commit','NF-FIX-023-import-facade-source-change','NF-FIX-027-retention-soft-delete']:
 directory=pathlib.Path('fixtures/network-flow')/identifier;slug=identifier.split('-',3)[3]
 for relative in ['expected/'+slug+'.expected.json','transcripts/'+slug+'.transcript.json']:
  result=json.loads((ROOT/directory/relative).read_text())
  for obj in walk(result):
   values=obj.get('diagnostics')
   if isinstance(values,list) and values and 'created_at' in values[0]:values.sort(key=lambda d:(d['network_flow_table_id'],d['source_row_number'],d['diagnostic_id']))
  result=replace_scalar(result,'cartulary.network_flow.cursor.v3','cartulary.network_flow.cursor.v4')
  result=replace_scalar(result,'cartulary.network_flow_indicator_link_result.v1','cartulary.network_flow_indicator_link_result.v2')
  write(directory/relative,result)
 update_manifest(directory,result['freeze_revision'])

# Direct default-graph successor expectations, independently re-bound after
# normalization. Generic Graph Projection itself retains its v2 contract.
import importlib.util
spec=importlib.util.spec_from_file_location('graph_oracle',ROOT/'fixtures/network-flow-orientation/graph-revision-oracle.py')
graph_oracle=importlib.util.module_from_spec(spec);spec.loader.exec_module(graph_oracle)
projection_template=oldjson(pathlib.Path('fixtures/network-flow/NF-FIX-015-graph-adapter-input/expected/graph-adapter-input.expected.json'))['outcomes'][0]['graph_queries'][0]['composition']['projection_invocations'][0]['projection_input']
for identifier in ['NF-FIX-001-cisco-sna-minimal','NF-FIX-002-cisco-sna-interface-fields','NF-FIX-004-rejected-rows']:
 revise_publication(identifier)
 directory=pathlib.Path('fixtures/network-flow')/identifier;slug=identifier.split('-',3)[3]
 result=json.loads((ROOT/directory/'expected'/f'{slug}.expected.json').read_text());changes={}
 for out in result['outcomes']:
  if out.get('graph'):graph_oracle.upgrade_graph(out['graph'],out['request']['incident_id'],projection_template,changes)
  for obj in walk(out):
   values=obj.get('diagnostics')
   if isinstance(values,list) and values and 'created_at' in values[0]:values.sort(key=lambda d:(d['network_flow_table_id'],d['source_row_number'],d['diagnostic_id']))
 def replace_graph(v):
  if isinstance(v,str):return changes.get(v,v)
  if isinstance(v,list):return [replace_graph(x) for x in v]
  if isinstance(v,dict):return {k:replace_graph(x) for k,x in v.items()}
  return v
 result=replace_graph(result)
 write(directory/'expected'/f'{slug}.expected.json',result);write(directory/'transcripts'/f'{slug}.transcript.json',result);update_manifest(directory,result['freeze_revision'])

VERSIONS.update({f'cartulary.network_flow.{k}.v2':f'cartulary.network_flow.{k}.v3' for k in ['graph_query_request','graph_query_result','graph_semantic_query','graph_contributor_query_request','graph_contributor_query_result']})
VERSIONS.update({'cartulary.network_flow.graph_contributor_query_continuation.v1':'cartulary.network_flow.graph_contributor_query_continuation.v2'})
def current_query_fields(result):
 for obj in walk(result):
  if 'field_key' in obj and ('op' in obj or 'direction' in obj):obj['field_key']=obj['field_key'].replace('network_flow.input_first_','network_flow.from_').replace('network_flow.input_second_','network_flow.to_')
  if obj.get('kind') in ('row_field_value','row_refs','graph_edge') and 'field_key' in obj:obj['field_key']=obj['field_key'].replace('network_flow.input_first_','network_flow.endpoint_a_').replace('network_flow.input_second_','network_flow.endpoint_b_')

def revise_graph_family(identifier):
 revise_publication(identifier)
 directory=pathlib.Path('fixtures/network-flow')/identifier;slug=identifier.split('-',3)[3]
 result=json.loads((ROOT/directory/'expected'/f'{slug}.expected.json').read_text());changes={};projections={}
 current_query_fields(result)
 for out in result['outcomes']:
  for obj in list(walk(out)):
   if 'graph_projection_result' not in obj or 'semantic_query' not in obj:continue
   previous=obj['graph_projection_result']['source_snapshot_id']
   context=out.get('result',out) if out.get('kind') in ('import','graph') else out
   projection=graph_oracle.upgrade_graph(obj,context['request']['incident_id'],projection_template,changes)
   projections[previous]=(projection,obj['graph_projection_result']['graph_view_id'])
  for obj in walk(out):
   if 'projection_input' in obj:
    old_snapshot=obj['projection_input']['source_snapshot_id']
    if old_snapshot not in projections:raise ValueError('missing independent graph result '+old_snapshot)
    obj['projection_input'],obj['graph_view_id']=copy.deepcopy(projections[old_snapshot])
   if obj.get('kind') in ('default_edge','time_bucket_edge'):
    obj.update(direction_kind='source_to_destination',counter_semantics='source_reported_flow_total')
   values=obj.get('diagnostics')
   if isinstance(values,list) and values and 'created_at' in values[0]:values.sort(key=lambda d:(d['network_flow_table_id'],d['source_row_number'],d['diagnostic_id']))
 def replace(v):
  if isinstance(v,str):return changes.get(v,v)
  if isinstance(v,list):return [replace(x) for x in v]
  if isinstance(v,dict):return {k:replace(x) for k,x in v.items()}
  return v
 result=replace(result)
 for relative in ['expected/'+slug+'.expected.json','transcripts/'+slug+'.transcript.json']:write(directory/relative,result)
 for path in (ROOT/directory/'source').glob('*.json'):
  value=json.loads(path.read_text());current_query_fields(value);value=replace(value)
  for obj in walk(value):
   if obj.get('kind') in ('default_edge','time_bucket_edge'):obj.update(direction_kind='source_to_destination',counter_semantics='source_reported_flow_total')
  write(path.relative_to(ROOT),value)
 update_manifest(directory,result['freeze_revision'])
for identifier in ['NF-FIX-006-cross-table-graph','NF-FIX-015-graph-adapter-input','NF-FIX-028-graph-aggregate-bounds']:
 revise_graph_family(identifier)
VERSIONS.update({f'cartulary.network_flow.{k}.v1':f'cartulary.network_flow.{k}.v2' for k in ['indicator_link_request','indicator_binding','table_query_continuation','rejected_rows_query_request','rejected_rows_query_continuation']})
VERSIONS.update({'cartulary.network_flow_table_get.v1':'cartulary.network_flow_table_get.v2','cartulary.network_flow_indicator_link_result.v1':'cartulary.network_flow_indicator_link_result.v2','cartulary.network_flow.cursor.v3':'cartulary.network_flow.cursor.v4'})
for identifier in ['NF-FIX-007-indicator-linking','NF-FIX-008-large-limits','NF-FIX-009-soft-delete-stale-graph','NF-FIX-014-cursor-pagination','NF-FIX-016-redaction','NF-FIX-017-indicator-link-mismatch','NF-FIX-024-query-normalization-cursors','NF-FIX-025-graph-contributors','NF-FIX-026-audit-and-replay']:
 revise_graph_family(identifier)

# Immutable link receipts bind the current selectors and new row identities.
# Recalculate their byte transcripts from the authored request, not the reply.
def link_digest(request):
 value={k:request[k] for k in ['selector','target','observation_mode','confirm_exact_value']}
 def order(v):
  if isinstance(v,list):return [order(x) for x in v]
  if isinstance(v,dict):
   keys=['network_flow_table_id','network_flow_row_id','source_row_number','mapping_fingerprint'] if set(v)=={'network_flow_table_id','network_flow_row_id','source_row_number','mapping_fingerprint'} else sorted(v)
   return {k:order(v[k]) for k in keys}
  return v
 return sha(parts('cartulary.network_flow.mutation_request_digest.v1','nf.indicator_links.create','indicator-links')+cjson(order(value))+b'\0')
for identifier in ['NF-FIX-007-indicator-linking','NF-FIX-009-soft-delete-stale-graph','NF-FIX-017-indicator-link-mismatch','NF-FIX-025-graph-contributors','NF-FIX-026-audit-and-replay']:
 directory=pathlib.Path('fixtures/network-flow')/identifier;slug=identifier.split('-',3)[3]
 result=json.loads((ROOT/directory/'expected'/f'{slug}.expected.json').read_text())
 for out in result['outcomes']:
  requests={}
  for step in out.get('link_steps',[]):
   if step.get('response',{}).get('data') is not None and step['response'].get('error') is None:
    request=step['request'];requests.setdefault(request['client_txn_id'],request)
  for obj in walk(out):
   if obj.get('route_key')=='nf.indicator_links.create' and 'request_hash' in obj:obj['request_hash']=link_digest(requests[obj['client_txn_id']])
   if obj.get('version')=='cartulary.network_flow.cursor.v4' and 'graph_query' in obj.get('query_echo',{}):obj['query_hash']=sha(cjson(obj['query_echo'],sort=True))
  for step in out.get('link_steps',[]):
   error=step.get('response',{}).get('error')
   if isinstance(error,dict) and 'field_key' in error.get('details',{}):
    error['details']['field_key']=error['details']['field_key'].replace('network_flow.input_first_','network_flow.endpoint_a_').replace('network_flow.input_second_','network_flow.endpoint_b_')
 for relative in ['expected/'+slug+'.expected.json','transcripts/'+slug+'.transcript.json']:write(directory/relative,result)
 update_manifest(directory,result['freeze_revision'])

# Rows and saved graph request variants also cross their public contract boundary.
directory=pathlib.Path('fixtures/network-flow/NF-FIX-024-query-normalization-cursors')
for path in (ROOT/directory).rglob('*.json'):
 if path.name=='manifest.json':continue
 result=json.loads(path.read_text())
 for before,after in [('cartulary.network_flow.rows_query_request.v1','cartulary.network_flow.rows_query_request.v2'),('cartulary.network_flow.rows_query_continuation.v1','cartulary.network_flow.rows_query_continuation.v2'),('cartulary.network_flow.rows_query_result.v1','cartulary.network_flow.rows_query_result.v2'),('cartulary.network_flow.graph_view_create_request.v3','cartulary.network_flow.graph_view_create_request.v4')]:result=replace_scalar(result,before,after)
 write(path.relative_to(ROOT),result)
update_manifest(directory,6)

revise_graph_family("NF-FIX-018-resource-limits")

revise_publication("NF-FIX-029-cisco-sna-flow-search")
# The Cisco successor regroups independently calculated observations. Unlike
# directed-profile revisions, its predecessor's edge membership is discarded.
import datetime
directory=pathlib.Path('fixtures/network-flow/NF-FIX-029-cisco-sna-flow-search');slug='cisco-sna-flow-search'
result=json.loads((ROOT/directory/'expected'/f'{slug}.expected.json').read_text());out=result['outcomes'][0]
corpus=json.loads((ROOT/'fixtures/network-flow-orientation/v1/example1200.expected.json').read_text())
rows={r['source_row_number']:r for r in out['durable']['rows']}
preview=out['preview']['owner_result'];unknown=sum(r['direction_kind']=='unknown' for n,r in rows.items() if n<=preview['preview_record_count']+1)
preview['preview_unknown_direction_count']=unknown;preview['preview_resolved_direction_count']=preview['preview_accepted_count']-unknown
graph=out['graph'];g=graph['graph_projection_result'];edge_template=copy.deepcopy(g['edges'][0]);g['edges']=[];graph['edge_annotations']=[]
vertices={v['properties']['endpoint_value']:v for v in g['vertices']}
for i,group in enumerate(corpus['edges']):
 members=[rows[n] for n in group['source_rows']]
 members.sort(key=lambda r:(datetime.datetime.fromisoformat(r['flow_start_utc']),datetime.datetime.fromisoformat(r['flow_end_utc']),r['source_row_number'],r['network_flow_row_id']))
 first,second=group['first'],group['second'];av,bv=vertices[first['ip']],vertices[second['ip']]
 identity=f'oracle-edge-{i}';edge=copy.deepcopy(edge_template);edge['edge_id']=identity;edge['src_vertex_id']=av['vertex_id'];edge['dst_vertex_id']=bv['vertex_id']
 table=members[0]['network_flow_table_id'];fp=members[0]['mapping_fingerprint'];count=len(members);limit=graph['result_limits']['max_example_row_refs_per_edge'];truncated=count>limit
 props=edge['properties'];props.update(edge_id=identity,src_endpoint_id=av['source_entity_ref']['source_entity_id'],dst_endpoint_id=bv['source_entity_ref']['source_entity_id'],src_port=first['port'],dst_port=second['port'],ip_protocol=group['protocol'],direction_kind=group['direction_kind'],counter_semantics='conversation_total',flow_row_count=count,bytes_sum=group['bytes'],packets_sum=group['packets'],first_flow_start_utc=min(members,key=lambda r:datetime.datetime.fromisoformat(r['flow_start_utc']))['flow_start_utc'].replace('+00:00','Z'),last_flow_end_utc=max(members,key=lambda r:datetime.datetime.fromisoformat(r['flow_end_utc']))['flow_end_utc'].replace('+00:00','Z'),example_refs_total_count=count,example_refs_truncated=truncated)
 edge['metadata']['mapped_metadata'].update(example_refs_total_count=count,mapping_fingerprints=[fp]);g['edges'].append(edge)
 selector={'kind':'default_edge','source_edge_id':identity,'direction_kind':group['direction_kind'],'counter_semantics':'conversation_total','source_endpoint_value':first['ip'],'destination_endpoint_value':second['ip'],'protocol':group['protocol'],'destination_port_present':second['port'] is not None}
 if first['port'] is not None:selector['source_port']=first['port']
 if second['port'] is not None:selector['destination_port']=second['port']
 graph['edge_annotations'].append({'projected_edge_id':identity,'selector':selector,'example_row_refs':[{k:r[k] for k in ['network_flow_table_id','network_flow_row_id','source_row_number','mapping_fingerprint']} for r in members[:limit]],'example_refs_truncated':truncated,'example_refs_total_count':count})
changes={};projection=graph_oracle.upgrade_graph(graph,out['request']['incident_id'],projection_template,changes)
# Semantic source IDs and projected IDs occupy distinct reference domains.
for annotation,edge in zip(graph['edge_annotations'],graph['graph_projection_result']['edges']):annotation['selector']['source_edge_id']=edge['properties']['edge_id']
invocation=out['composition']['projection_invocations'][0];invocation['projection_input']=projection;invocation['graph_view_id']=graph['graph_projection_result']['graph_view_id']
def replace_cisco(v):
 if isinstance(v,str):return changes.get(v,v)
 if isinstance(v,list):return [replace_cisco(x) for x in v]
 if isinstance(v,dict):return {k:replace_cisco(x) for k,x in v.items()}
 return v
result=replace_cisco(result)
for relative in ['expected/'+slug+'.expected.json','transcripts/'+slug+'.transcript.json']:write(directory/relative,result)
update_manifest(directory,7)
# Audit counts are independently derived from the same complete membership,
# including examples omitted by the explicitly selected per-edge bound.
result=json.loads((ROOT/directory/'expected'/f'{slug}.expected.json').read_text())
for audit in result['outcomes'][0]['durable']['audit']:
 facts=audit.get('after_json',{})
 if 'result_edge_count' in facts:
  facts['result_edge_count']=len(corpus['edges'])
  facts['truncated_example_ref_count']=sum(max(0,len(group['source_rows'])-limit) for group in corpus['edges'])
for relative in ['expected/'+slug+'.expected.json','transcripts/'+slug+'.transcript.json']:write(directory/relative,result)
update_manifest(directory,7)
