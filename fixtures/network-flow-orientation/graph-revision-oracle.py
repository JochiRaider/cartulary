# Independent successor calculation for the direct Graph Projection grammar.
# Inputs are predecessor fixture facts and current normalized row expectations;
# no product code, captured run output, or documentation is evaluated.
import copy, hashlib, hmac, json, struct

def compact(v):return json.dumps(v,ensure_ascii=False,sort_keys=True,separators=(',',':')).encode()
def sha(v):return hashlib.sha256(v).hexdigest()
def framed(*xs):
 values=[x if isinstance(x,bytes) else str(x).encode() for x in xs]
 return b''.join(struct.pack('>Q',len(x))+x for x in values)
def nul(*xs):return b''.join(str(x).encode()+b'\0' for x in xs)
def tuple_hash(domain,*xs):
 values=[x.encode() if isinstance(x,str) else compact(x) for x in xs]
 return sha(domain.encode()+b''.join(str(len(x)).encode()+b':'+x+b'\n' for x in values))
def graph_digest(incident,s):
 # Network Flow's semantic JSON uses its explicit member order.
 encode=lambda v:json.dumps(v,ensure_ascii=False,separators=(',',':')).encode()
 tr=s['time_range'] if any(s['time_range'].values()) else None
 return sha(framed('cartulary.network_flow.graph_query_digest.v3',incident,encode(sorted(s['selected_table_ids'])),encode(s['filters']),encode(tr),'cartulary.network_flow.graph_semantic_query.v3',encode(s['aggregation'])))
def semantic_edge_id(incident,kind,counters,a,b,protocol,bucket=None):
 fields=['cartulary.network_flow_flow_edge_id.v2',incident,kind,counters]
 for ip,port in (a,b):
  fields.extend([ip,'n' if port is None else 'p'])
  if port is not None:fields.append(str(port))
 fields.append(str(protocol))
 if bucket:fields.extend(bucket)
 return ('nfbe_' if bucket else 'nff_')+sha(framed(*fields))

def upgrade_graph(result,incident,template,changes):
 g=result['graph_projection_result'];s=result['semantic_query'];s['schema_id']='cartulary.network_flow.graph_semantic_query.v3'
 for f in s['filters']:
  f['field_key']=f['field_key'].replace('network_flow.input_first_','network_flow.from_').replace('network_flow.input_second_','network_flow.to_')
 digest=graph_digest(incident,s);
 safe=lambda v:hmac.new(bytes([2])*32,nul('cartulary.network_flow.safe_digest.v1','graph_query_digest')+v.encode(),hashlib.sha256).hexdigest()
 changes[safe(result['graph_query_digest'])]=safe(digest)
 changes[result['graph_query_digest']]=digest;result['graph_query_digest']=digest;result['schema_id']='cartulary.network_flow.graph_query_result.v3'
 refs=sorted(result['source_table_refs'],key=lambda x:x['network_flow_table_id'])
 snapshot_fields=['cartulary.network_flow.source_snapshot_digest.v1',incident,str(len(refs))]
 for r in refs:snapshot_fields.extend([r['network_flow_table_id'],r['mapping_fingerprint']])
 snapshot='nfsnap_'+sha(nul(*snapshot_fields,digest));changes[g['source_snapshot_id']]=snapshot
 key='network_flow_activity:'+incident+':'+snapshot.removeprefix('nfsnap_')
 view='gv_'+sha(framed('cartulary.graph_projection_graph_view_identity.v2','network_flow_activity',key));changes[g['graph_view_id']]=view
 temporal=s['aggregation']['mode']=='time_bucket_v1';edgekind='network_flow.bucketed_flow_edge.v2' if temporal else 'network_flow.flow_edge.v2';rule='nf.map.bucketed_flow_edge.v1' if temporal else 'nf.map.flow_edge.v1'
 version='network_flow_activity.time_bucket.v2' if temporal else 'network_flow_activity.v2'
 entities=[];relationships=[];vertex_ids={}
 for v in g['vertices']:
  ref=v['source_entity_ref'];mapping=tuple_hash('GPMAPENTITY1\n',ref['mapping_rule_id'],ref['source_entity_kind'],v['vertex_kind']);old=v['vertex_id']
  v['vertex_id']='vx_'+tuple_hash('GPVERTEX1\n','direct_vertex','graph_projection.v2',view,ref['source_entity_kind'],ref['source_entity_id'],mapping)
  v['sort_key']='sk_'+tuple_hash('GPSORT1\n','vertex','direct',v['vertex_kind'],ref['source_entity_kind'],ref['source_entity_id'],mapping,v['vertex_id']);changes[old]=v['vertex_id'];vertex_ids[old]=v['vertex_id']
  entities.append({'source_entity_id':ref['source_entity_id'],'source_entity_kind':ref['source_entity_kind'],'properties':v['properties'],'metadata':v['metadata']['mapped_metadata'],'labels':[]})
 endpoints={v['source_entity_ref']['source_entity_id']:v['properties']['endpoint_value'] for v in g['vertices']}
 for e in g['edges']:
  p=e['properties'];kind=p.get('direction_kind','source_to_destination');counters=p.get('counter_semantics','source_reported_flow_total');p.update(direction_kind=kind,counter_semantics=counters,src_port=p.get('src_port'))
  bucket=[p['bucket_start_utc'],p['bucket_end_utc']] if temporal else None
  identity=semantic_edge_id(incident,kind,counters,(endpoints[p['src_endpoint_id']],p['src_port']),(endpoints[p['dst_endpoint_id']],p['dst_port']),p['ip_protocol'],bucket)
  changes[p['edge_id']]=identity;p['edge_id']=identity;e['edge_kind']=edgekind
  e['src_vertex_id']=vertex_ids.get(e['src_vertex_id'],e['src_vertex_id']);e['dst_vertex_id']=vertex_ids.get(e['dst_vertex_id'],e['dst_vertex_id']);e['direction']='undirected' if kind=='unknown' else 'directed'
  mapping=tuple_hash('GPMAPREL1\n',rule,edgekind,edgekind,'preserve',False,edgekind);old=e['edge_id']
  e['edge_id']='ed_'+tuple_hash('GPEDGE1\n','direct_edge','graph_projection.v2',view,edgekind,identity,edgekind,e['src_vertex_id'],e['dst_vertex_id'],e['direction'],mapping)
  e['sort_key']='sk_'+tuple_hash('GPSORT1\n','edge','direct',edgekind,edgekind,identity,e['src_vertex_id'],e['dst_vertex_id'],e['direction'],mapping,False,e['edge_id']);changes[old]=e['edge_id']
  e['source_relationship_ref'].update(source_relationship_id=identity,source_relationship_kind=edgekind,mapping_rule_id=rule)
  relationships.append({'source_relationship_id':identity,'source_relationship_kind':edgekind,'src_source_entity_id':p['src_endpoint_id'],'dst_source_entity_id':p['dst_endpoint_id'],'direction':'undirected' if kind=='unknown' else 'forward','properties':p,'metadata':{**e['metadata']['mapped_metadata'],'example_refs_truncated':p['example_refs_truncated']},'labels':[]})
 # Fixed schema declarations are expanded from the independently reviewed
 # predecessor, with exactly the three newly required semantic properties.
 registry=g['schema_registry']
 def replace_kind(v):
  if isinstance(v,str):return v.replace('network_flow.flow_edge.v1',edgekind).replace('network_flow.bucketed_flow_edge.v1',edgekind)
  if isinstance(v,list):return [replace_kind(x) for x in v]
  if isinstance(v,dict):return {k:replace_kind(x) for k,x in v.items()}
  return v
 registry=replace_kind(registry);g['schema_registry']=registry
 for key,typ,nullable in [('src_port','integer',True),('direction_kind','string',False),('counter_semantics','string',False)]:
  for ek in registry['edge_kinds']:
   if not any(p['projected_key']==key for p in ek['properties']):ek['properties'].append({'projected_key':key,'projected_type':typ,'required':True,'nullable_output':nullable})
   ek['properties'].sort(key=lambda p:p['projected_key'])
  if not any(p['target_scope']=='edge' and p['projected_key']==key for p in registry['property_keys']):registry['property_keys'].append({'target_scope':'edge','target_kind':edgekind,'projected_key':key,'projected_type':typ,'required':True,'nullable_output':nullable,'missing_behavior':'error','source_null_behavior':'emit_null' if nullable else 'error'})
 registry['property_keys'].sort(key=lambda p:(p['target_scope'],p['target_kind'],p['projected_key']))
 config=replace_kind(copy.deepcopy(template['projection_config']));config['projection_version']=version
 rm=config['relationship_mappings'][0];rm['mapping_rule_id']=rule
 for key in ['src_port','direction_kind','counter_semantics']:
  if key not in rm['required_property_keys']:rm['required_property_keys'].append(key)
 if temporal:
  for key in ['bucket_start_utc','bucket_end_utc']:
   if key not in rm['required_property_keys']:rm['required_property_keys'].append(key)
 definitions=replace_kind(copy.deepcopy(template['property_definitions']))
 for key,typ,nullable in [('src_port','integer',True),('direction_kind','string',False),('counter_semantics','string',False)]+([('bucket_start_utc','timestamp',False),('bucket_end_utc','timestamp',False)] if temporal else []):
  if not any(p['target_scope']=='edge' and p['projected_key']==key for p in definitions):definitions.append({'property_definition_id':'nf.pd.edge.'+key+'.v1','target_scope':'edge','target_kind':edgekind,'source_field_path':'properties.'+key,'projected_key':key,'projected_type':typ,'required':True,'missing_behavior':'error','source_null_behavior':'emit_null' if nullable else 'error','null_output_policy':'emit_null' if nullable else 'omit','merge_behavior':'single_value'})
 # Submission order belongs to Network Flow; normalized hash order belongs to Graph.
 edgekeys=['bytes_sum','contributing_table_ids','dst_endpoint_id','dst_port','src_port','direction_kind','counter_semantics','edge_id','example_refs_total_count','example_refs_truncated','first_flow_start_utc','flow_row_count','ip_protocol','last_flow_end_utc','packets_sum','src_endpoint_id']+(['bucket_start_utc','bucket_end_utc'] if temporal else [])
 rm['required_property_keys']=edgekeys
 vertexkeys=['contributing_table_ids','endpoint_kind','endpoint_value','flow_row_count','indicator_candidate_value']
 definitions.sort(key=lambda p:(0 if p['target_scope']=='vertex' else 1,(vertexkeys if p['target_scope']=='vertex' else edgekeys).index(p['projected_key'])))
 source={'source_snapshot_id':snapshot,'source_entities':sorted(entities,key=lambda v:v['source_entity_id']),'source_relationships':sorted(relationships,key=lambda e:e['source_relationship_id']),'source_metadata':{'incident_id':incident,'graph_query_digest':digest,'source_snapshot_id':snapshot,'selected_table_ids':s['selected_table_ids'],'mapping_fingerprints':[r['mapping_fingerprint'] for r in result['source_table_refs']]}}
 projection={'projection_schema_id':'graph_projection.v2',**source,'projection_config':config,'filters':copy.deepcopy(template['filters']),'property_definitions':definitions}
 normalized_config=copy.deepcopy(config)
 for group in ['entity_mappings','relationship_mappings']:
  for m in normalized_config[group]:
   for key in ['required_property_keys','optional_property_keys','mapping_labels']:m[key]=sorted(set(m[key]))
 conf={'projection_config':normalized_config,'filters':projection['filters'],'property_definitions':sorted(definitions,key=lambda p:p['property_definition_id'])}
 g['vertices'].sort(key=lambda v:(v['sort_key'],v['vertex_id']));g['edges'].sort(key=lambda e:(e['sort_key'],e['edge_id']))
 output={k:g[k] for k in ['mapped_metadata','schema_registry','vertices','edges','validation_summary','consumer_capabilities']}
 g.update(graph_view_id=view,source_snapshot_id=snapshot,projection_version=version,normalized_configuration_sha256=sha(compact(conf)),normalized_source_sha256=sha(compact(source)),canonical_output_sha256=sha(compact(output)))
 old=g['projection_result_id'];g['projection_result_id']='gpres_'+sha(framed('cartulary.graph_projection_result_identity.v2',view,'network_flow_activity',snapshot,version,g['normalized_configuration_sha256'],g['normalized_source_sha256'],g['canonical_output_sha256']));changes[old]=g['projection_result_id']
 for annotation in result['edge_annotations']:
  annotation['selector'].setdefault('direction_kind','source_to_destination');annotation['selector'].setdefault('counter_semantics','source_reported_flow_total')
 def replace(v):
  if isinstance(v,str):return changes.get(v,v)
  if isinstance(v,list):return [replace(x) for x in v]
  if isinstance(v,dict):return {k:replace(x) for k,x in v.items()}
  return v
 result.update(replace(result))
 vertex_order={v['vertex_id']:i for i,v in enumerate(g['vertices'])};edge_order={e['edge_id']:i for i,e in enumerate(g['edges'])}
 result['vertex_selectors'].sort(key=lambda x:vertex_order[x['projected_vertex_id']]);result['edge_annotations'].sort(key=lambda x:edge_order[x['projected_edge_id']])
 return projection
