# Current Graph execution binding cleanup

Status: the final eight-member binding is reviewed and adopted on 2026-11-02
under standing authorization, superseding the E-334 ten-member candidate; see
controlling tracker E-336.

## Finding and decision

The current binding contains per-stage algorithm-name/digest arrays and constant
packaged-subject/build-provenance placeholders. Inspection found no production
dispatcher, validator or engine consumer for these fields. They contribute only
to the binding's serialized shape and hash. Two digests are hashes of identifier
text, not definitions. Preserving or expanding this metadata would create a
second description of an execution pipeline that already has one typed owner.

The standalone database lineage/head fields also have no runtime consumer and
are not facts exposed by the frozen table catalog. Core owns actual database
generation admission; repeating unvalidated values in a Graph binding provides
no additional assurance.

Remove those fields. Retain the actual rebuild algorithm and explicitly bind
the projection schema consumed and returned by the real engine. Preserve real
build provenance and release qualification outside this logical compatibility
contract. Do not substitute source-file hashes, callback registries or new
facades for the removed placeholders.

## Concrete Graph owner amendment

In Graph Projection §9, declare the sole current implementation binding to be
`cartulary.graph_projection_restore_implementation_binding.v7`. Rebuild algorithm
v5, rebuild result v5, source registry v5 and pure projection input/result v2
remain current. Earlier versions of each individual contract are unsupported.

Add the following closed binding definition:

| Required member | Contract |
| --- | --- |
| `schema_id` | `cartulary.graph_projection_restore_implementation_binding.v7` |
| `algorithm_id` | `graphprojection.restore_rebuild.v5`, resolved to the actual typed Graph reconstruction implementation |
| `binding_id` | `graphprojection.restore_rebuild.network_flow_graph_views.v5` |
| `projection_schema_id` | `graph_projection.v2`, resolved to the actual typed projection input/result implementation |
| `semantic_query_schema_ids` | Exactly the sorted semantic query schema set supported by the current typed source registrations |
| `recovery_state_catalog_sha256` | Exact selected frozen Recovery catalog digest |
| `source_registry_sha256` | Exact selected frozen Graph source-registry digest |
| `graph_table_ids` | Exact derived Graph table set, in ASCII order |

No other members are admitted. Remove `graph_projection_contract_id`,
`graph_engine_algorithm_ids`, `graph_engine_algorithm_digests`,
`packaged_subject_sha256`, `build_provenance_sha256`,
`database_schema_lineage` and `database_schema_head` entirely. Neither a
document-version label nor a name/digest declaration proves an executable
implementation. Admission MUST verify the supported binding schema, actual
rebuild algorithm and actual projection schema independently of declaration
agreement, before any Graph mutation or publication. Decode definition bytes
strictly: duplicate members, invalid Unicode, unknown members and trailing JSON
are invalid; the admitted complete object must match its selected canonical
definition digest.

The current-only Recovery generation MUST select binding v7 and its exact
digest. Binding v6 and earlier, removed members, unsupported algorithm/schema
identities and stale generation pairs reject before mutation. No converter,
alias, historical reader, deletion or automatic control-record rewrite is added.
Pure Graph result identities, canonical bytes and database layout do not change.

Core 00’s frozen-input status summary and Core 01’s structural Recovery rule
will import the current Graph contracts from Graph Projection §9 instead of
repeating binding v6. Core 04 has no binding-v6 restatement to change. This
removes stale version restatements without changing the ownership/precedence
relationship.

## Implementation, validation and handoff

Update authored schema, binding and generation inputs, then regenerate through
Make. Resolve the projection schema against the existing `ProjectV2` boundary
and the rebuild algorithm against the existing closed restore service; preserve
the root/subpackage API boundaries. Ensure generator and runtime admission both
reject unsupported or coherently rebound identities. Keep static retired-shape
fixtures only where they prove rejection.

Validate current schema/generation parity, deterministic engine vectors, typed
binding admission, no mutation on rejection, current reconstruction and current
Recovery selection/capture/restore as selected through the owner guides. Record
the new current generation and digests, exact native exits and retained roots.
Old initialized targets remain non-serving when their existing completion tuple
does not prove this current generation; their data is retained.

This amendment changes the current binding format and therefore requires a
coherent unpublished cutover. Existing valid logical Graph data needs no
migration. It introduces no feature or source-owner reassignment. Final module
readiness still depends on the controlling tracker's remaining workstreams.
