---
title: Graph Projection NLSpec
status: adopted/current
document_class: nlspec
created_at: 2026-05-30
document_version: 2.0.0
---

## 1. Status, scope, and authority

Status: `adopted/current` (`Adopted`).

This NLSpec defines Graph Projection 2.0: deterministic, side-effect-free
derivation of a graph representation from an explicit immutable source. It
replaces Graph Projection 1.2.0 in full. Version 1 lifecycle, query, retention,
cursor, idempotency, invalidation, run-history, and five-table Recovery behavior
is not current behavior and MUST NOT be implemented through a compatibility
alias, translator, dual writer, or hidden feature flag.

Graph Projection owns semantic input admission, normalization, mapping,
aggregation, validation, canonical output, stable object identity, immutable
result identity, bounded traversal semantics, publication invariants, and
deterministic Recovery reconstruction. It does not own public routes,
authorization, authenticated actor identity, operation time, saved-graph
declarations, display names, job attempts, retry, cancellation requests,
retention policy, consumer release policy, visualization, or deployment
configuration.

Workbook projections, saved views, `view_row_v1`, workbook restore providers,
and workbook query behavior remain Core-owned. The Graph restore participant is
distinct from the workbook restore-provider registry.

## 2. Normative language and concepts

`MUST`, `MUST NOT`, and `MAY` are normative. Objects are closed unless a table
explicitly says otherwise. Unknown members, duplicate JSON members, invalid
UTF-8, invalid JSON, non-object roots, wrong JSON types, and explicit null for a
non-nullable member fail before semantic projection.

| Concept | Definition |
| --- | --- |
| Projection invocation | One caller-owned attempt to compute a result. Attempt identity and operational state are outside this owner. |
| Semantic input | The complete closed `graph_projection.v2` value whose normalized configuration and source participate in result identity. |
| Invocation context | Trusted non-serialized values `graph_view_id` and `source_owner_id`, plus cancellation/deadline capability. |
| Projection result | One immutable, deterministic graph output identified by `projection_result_id`. It has no lifecycle state or attempt timestamp. |
| Graph view | Caller-owned semantic declaration identity under which a result is computed. Graph Projection does not create or mutate the declaration. |
| Source snapshot | Caller-declared immutable boundary from which all submitted source objects were read. |
| Direct object | Vertex or edge emitted from one mapped source object. |
| Aggregated object | Vertex or edge emitted from a deterministic contributor group. |
| Publication | Caller-coordinated atomic insertion of an immutable result envelope and all of its objects. |
| Lease | Operational reachability claim preventing cleanup of an exact result while a consumer depends on it. |

Authoritative source data remains the source of truth. Graph results are derived
and MUST NOT mutate, replace, or become authority for their source objects.

## 3. Invocation and input contract

### 3.1 Trusted invocation context

The caller MUST supply `graph_view_id` and `source_owner_id` through a trusted
typed invocation boundary, not inside untrusted semantic JSON. Both use the
identifier scalar below. The engine MUST reject an empty or malformed value
before decoding semantic input. The same context also carries cancellation and
deadline signals; those signals never enter canonical bytes or identity.

### 3.2 Common scalars

| Scalar | Contract |
| --- | --- |
| `identifier` | UTF-8 string, 1..255 bytes, no leading/trailing Unicode whitespace, no C0/C1 control, slash, backslash, or NUL. |
| `property_key` | Identifier with 1..255 bytes; `*` is reserved for explicit wildcard positions. |
| `sha256_hex` | Exactly 64 lowercase hexadecimal characters. |
| `finite_integer` | JSON integer with no fraction or exponent and within signed 64-bit range. |
| `property_value` | Null, boolean, finite integer, finite JSON number, string, or a flat array of those scalar values; nested objects/arrays are invalid. |

### 3.3 `graph_projection.v2`

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `projection_schema_id` | string | Yes | No | None | Exactly `graph_projection.v2`. |
| `source_snapshot_id` | identifier | Yes | No | None | Immutable caller-owned source boundary. |
| `projection_config` | object | Yes | No | None | Table 3-B. |
| `source_entities` | array | Yes | No | `[]` | Table 3-C, unique IDs. |
| `source_relationships` | array | Yes | No | `[]` | Table 3-D, unique IDs. |
| `source_metadata` | object | Yes | No | `{}` | Closed only by caller/adapter schema; values must be canonical JSON. |
| `filters` | object | Yes | No | Empty filter object | Table 3-E. |
| `property_definitions` | array | Yes | No | `[]` | Table 3-H; unique definition IDs. |

The v2 input MUST NOT accept `graph_view_id`, `source_owner_id`, `requested_at`,
`requested_by`, retention fields, `custom_config`, relationship-definition
aliases, operation IDs, job IDs, or lifecycle members. Callers own those facts.

### 3.4 Projection configuration

**Table 3-B. `projection_config`**

| Member | Required | Default | Rule |
| --- | ---: | --- | --- |
| `projection_version` | Yes | None | Non-empty identifier naming caller mapping semantics. |
| `declared_source_entity_kinds` | Yes | `[]` | Sorted unique identifiers. |
| `declared_source_relationship_kinds` | Yes | `[]` | Sorted unique identifiers. |
| `entity_mappings` | Yes | `[]` | Table 3-F, sorted by `mapping_rule_id`. |
| `relationship_mappings` | Yes | `[]` | Table 3-G, sorted by `mapping_rule_id`. |
| `metadata_mappings` | Yes | `[]` | Table 3-I, sorted by `metadata_mapping_id`. |
| `aggregation_rules` | Yes | `[]` | Table 3-J, dependency order then rule ID. |
| `default_vertex_labels` | Yes | `[]` | Sorted unique strings. |
| `default_edge_labels` | Yes | `[]` | Sorted unique strings. |
| `allow_empty_kind_registry` | Yes | `false` | Empty declared-kind arrays are valid only when true. |

### 3.5 Source objects and filters

**Table 3-C. Source entity**

| Member | Required | Default | Rule |
| --- | ---: | --- | --- |
| `source_entity_id` | Yes | None | Identifier unique in the input. |
| `source_entity_kind` | Yes | None | Declared entity kind. |
| `properties` | Yes | `{}` | Property-value object. |
| `metadata` | Yes | `{}` | Canonical JSON object available only to declared metadata mappings. |
| `labels` | Yes | `[]` | Sorted unique strings. |

**Table 3-D. Source relationship**

| Member | Required | Default | Rule |
| --- | ---: | --- | --- |
| `source_relationship_id` | Yes | None | Identifier unique in the input. |
| `source_relationship_kind` | Yes | None | Declared relationship kind. |
| `src_source_entity_id` | Yes | None | Existing submitted source entity. |
| `dst_source_entity_id` | Yes | None | Existing submitted source entity. |
| `direction` | Yes | None | `forward`, `reverse`, `undirected`, or `bidirectional`. |
| `properties` | Yes | `{}` | Property-value object. |
| `metadata` | Yes | `{}` | Canonical JSON object available only to declared metadata mappings. |
| `labels` | Yes | `[]` | Sorted unique strings. |

**Table 3-E. Filters**

`filters` contains exactly `entity_filters`, `relationship_filters`, and
`logic='and'`. Each filter has `field_path`, `operator`, `value`, and
`include_if_missing`. Operators are `exists`, `equals`, `not_equals`, and `in`.
`exists` requires omitted `value`; the others require it. Entity paths begin
with `properties.`, `metadata.`, `source_entity_id`, or `source_entity_kind`;
relationship paths use the corresponding relationship names plus endpoint and
direction fields. Filters execute in declared order and only determine
eligibility; they do not mutate input.

### 3.6 Mapping definitions

**Table 3-F. Entity mapping**

| Member | Rule |
| --- | --- |
| `mapping_rule_id` | Unique identifier. |
| `source_entity_kind` | Exactly one declared source kind; one entity mapping per source kind. |
| `projected_vertex_kind` | Output kind identifier. |
| `inclusion_predicate` | `always` or one Table 3-E filter object. |
| `label_policy` | `mapping_only`, `preserve_source`, or `mapping_then_source`. |
| `mapping_labels` | Sorted unique strings. |
| `required_property_keys`, `optional_property_keys` | Sorted, unique, disjoint property keys. |

**Table 3-G. Relationship mapping**

Entity-mapping members apply with relationship/edge vocabulary. It additionally
contains `direction_policy` (`preserve`, `normalize_forward`,
`normalize_reverse`, `undirected`, or `bidirectional`), `emit_reverse_edge`,
and `reverse_edge_kind`. A reverse edge is valid only for normalized directed
output; when disabled, `reverse_edge_kind` MUST be absent.

**Table 3-H. Property definition**

Each item contains `property_definition_id`, `target_scope` (`vertex`, `edge`,
or `graph_view`), `target_kind` (kind or `*`), `source_field_path`,
`projected_key`, `projected_type`, `required`, optional `default_value`,
`missing_behavior`, `source_null_behavior`, `null_output_policy`, and
`merge_behavior`. Projected types are `boolean`, `integer`, `number`, `string`,
`timestamp`, `identifier`, and their `_array` forms. Missing behavior is
`error`, `omit`, or `default`; null behavior is `error`, `omit`, `emit_null`, or
`default`; null output is `omit` or `emit_null`; merge behavior is
`single_value`, `first`, `last`, `min`, `max`, `sum`, `count`, `set`, or
`ordered_list`. Default modes require a type-compatible default. Wildcard
expansion MUST NOT create duplicate `(scope, kind, projected_key)` tuples.

**Table 3-I. Metadata mapping**

Metadata mappings use the Table 3-H behavior members with
`metadata_mapping_id`, `projected_metadata_key`, and no `projected_key`.
System metadata keys are reserved and cannot be overwritten. Unmapped source
metadata MUST NOT appear in output.

**Table 3-J. Aggregation rule**

Each item contains `aggregation_rule_id`, `target_scope`, `input_scope`,
`input_kind`, `projected_kind`, non-empty `grouping_keys`,
`missing_grouping_key_behavior` (`error`, `exclude`, or `use_null`),
`property_merge_behavior`, and, for edges, `edge_direction` and
`endpoint_grouping`. Input scopes are source entity/relationship or projected
vertex/edge. References to earlier projected aggregation output form an acyclic
dependency graph. Endpoint grouping names the source and destination vertex
aggregation rules, their grouping keys, and `missing_endpoint_behavior`
(`error` or `exclude`).

## 4. Admission, normalization, and canonical bytes

Admission performs decoding, closed-shape validation, default materialization,
scalar validation, duplicate detection, resource admission, mapping validation,
and source-reference validation in that order. Failure returns a closed
`ProjectionError` and no partial result.

Objects are serialized as UTF-8 canonical JSON: keys in Unicode code-point
order, no insignificant whitespace, shortest valid JSON escapes, exact JSON
literals, and normalized finite numbers. Arrays preserve semantic order except
arrays explicitly declared sorted. Object-map iteration and filesystem order
MUST NOT affect bytes.

`normalized_configuration_sha256` is SHA-256 of the canonical materialized
`projection_config`, `filters`, and `property_definitions`. It excludes source
objects, source metadata, trusted context, and operational data.

`normalized_source_sha256` is SHA-256 of canonical `source_snapshot_id`, sorted
source entities, sorted source relationships, and `source_metadata`. It excludes
configuration, trusted context, and operational data.

## 5. Projection semantics

The engine evaluates cancellation before admission, after every 1,024 source or
output items, between direct and aggregated phases, and before returning.

Direct vertices are emitted for eligible source entities with a mapping. Direct
edges are emitted for eligible source relationships whose selected endpoints
exist. Direction policy is applied before reverse-edge behavior. Labels are the
sorted unique result of the default, mapping, and permitted source label sets.
Properties and mapped metadata follow their definitions; undeclared values are
not copied.

Aggregations execute in declared dependency order. Group keys use canonical
scalar tuple encoding. Contributors sort by source kind, source ID, mapping or
aggregation rule ID, then prior projected object ID. Merge operations consume
that order. Aggregated edges may be emitted only when both endpoints resolve;
an `exclude` rule omits the edge, while `error` fails projection.

Vertex and edge IDs retain prefixes `vx_` and `ed_` and are SHA-256-derived from
length-framed tuples of their semantic family, graph view, kind, mapping or
aggregation identity, source/contributor identity, endpoints, and direction.
The same semantic object MUST receive the same ID across retry, restart, and
Recovery.

Vertices sort by `(sort_key, vertex_id)` and edges by `(sort_key, edge_id)`, all
ascending bytewise. Every edge endpoint MUST name an emitted vertex. Schema
registry arrays and nested labels/property references sort by their semantic
identifier tuples.

## 6. Result and identity contract

### 6.1 `graph_projection_result.v2`

| Member | Rule |
| --- | --- |
| `projection_schema_id` | Exactly `graph_projection.v2`. |
| `projection_result_id` | `gpres_` plus lowercase SHA-256 hex from §6.2. |
| `graph_view_id` | Exact trusted invocation value. |
| `source_owner_id` | Exact trusted invocation value. |
| `source_snapshot_id` | Exact admitted input value. |
| `projection_version` | Exact normalized configuration value. |
| `normalized_configuration_sha256` | Exact §4 digest. |
| `normalized_source_sha256` | Exact §4 digest. |
| `canonical_output_sha256` | Exact §6.2 digest. |
| `properties` | Derived graph properties or `{}`. |
| `mapped_metadata` | Declared graph metadata or `{}`. |
| `schema_registry` | Exact output-kind/property/metadata schemas. |
| `vertices`, `edges` | Canonically ordered complete output. |
| `validation_summary` | Zero-issue success summary. |
| `consumer_capabilities` | Ordered-read and bounded-traversal capabilities only. |

The result has no run ID, attempt state, previous result, invalidation,
accepted/generated/completed time, actor, job, retention, cursor, or failure
history. Ephemeral and retained callers receive the same semantic result shape;
retention is a caller/storage decision and does not change identity.

### 6.2 Non-recursive identity transcript

`canonical_output_sha256` hashes canonical result content beginning at
`properties` through `consumer_capabilities`. The digest input excludes all
identity/digest fields and all operational storage metadata.

`projection_result_id` hashes this length-framed transcript:

```text
cartulary.graph_projection_result_identity.v2
graph_view_id
source_owner_id
source_snapshot_id
projection_version
normalized_configuration_sha256
normalized_source_sha256
canonical_output_sha256
```

Each field is UTF-8 encoded as unsigned 64-bit big-endian byte length followed
by exact bytes. No separator or terminal NUL is added. Existing ID with the same
tuple and bytes is an idempotent result; the same ID with any different tuple,
digest, or bytes is `projection_result_identity_conflict`.

### 6.3 Vertex and edge shapes

Vertices contain `vertex_id`, `vertex_kind`, `vertex_family`, `labels`,
`properties`, `metadata`, nullable `source_entity_ref`, and `sort_key`. Edges
contain corresponding edge fields plus source/destination vertex IDs,
direction, nullable `source_relationship_ref`, reverse-edge metadata, and
`sort_key`. Aggregated metadata contains ordered source refs; direct objects
contain exactly one typed source ref. Objects are closed.

## 7. Errors, validation, and limits

`ProjectionError` contains `code`, `reason_code`, `retry_action`, and closed
safe `details`. Codes are `invalid_projection_request`,
`projection_validation_failed`, `projection_resource_limit_exceeded`,
`projection_cancelled`, and `projection_computation_failed`. Retry action is
`do_not_retry` except an infrastructure computation failure may be
`retry_with_backoff`. Source property values, metadata values, stack text, SQL,
paths, and secrets MUST NOT appear in errors, logs, metrics, or traces.

The first producer uses these non-configurable semantic limits:

| Limit | Value |
| --- | ---: |
| Input bytes | 268,435,456 |
| Source entities / projected vertices | 100,000 each |
| Source relationships / projected edges | 250,000 each |
| Entity filters / relationship filters | 1,000 each |
| Entity mappings / relationship mappings / property definitions / metadata mappings | 10,000 each |
| Aggregation rules | 1,000 |
| Properties or metadata keys per object | 1,024 each |
| Labels per object / label bytes | 256 / 256 |
| String property bytes | 16,384 |
| Validation issues retained internally | 100,000 |
| Traversal seed vertices / kind filters / depth | 1,024 / 1,024 / 16 |

Limit checks use bounded counting and report at most `limit + 1` for count
failures. Projection returns no partial graph on any validation or limit error.

## 8. Narrow persistence and consumer ports

Graph Projection defines behavior, not a repository facade. An adapter MAY
implement only these typed capabilities:

- `PublishResultTx`: insert one immutable result envelope, all vertices, and
  all edges through a caller-owned transaction;
- `ReadExactResult`: read one result by full binding tuple;
- `ReadVertices` and `ReadEdges`: return canonical output order;
- `Traverse`: breadth-first traversal with depth, direction, kind, and count
  bounds, breaking ties by output order;
- `AcquireLeaseTx`, `RenewLease`, and `ReleaseLease`: protect an exact result for
  an owner/purpose until a server-owned expiry;
- `DeleteUnreachableResults`: delete only unleased results not selected by an
  active authoritative declaration.

Publication validates all digests, counts, IDs, endpoints, and order before the
caller transaction commits. Partial envelope/object publication is forbidden.
Leases are operational rows and do not enter result identity. Adapters accept
borrowed database/transaction handles, start no hidden worker, and close no
borrowed resource.

## 9. Recovery

Graph Projection publishes algorithm `graphprojection.restore_rebuild.v2` and
result schema `cartulary.graph_projection_restore_rebuild_result.v2`. The
current Recovery catalog contains exactly these derived tables in ASCII order:

```text
graph_projection_result_edges
graph_projection_result_leases
graph_projection_result_vertices
graph_projection_results
```

Recovery restores source-owner authoritative declarations before invoking the
Graph participant. Workers are quiescent. The participant validates the frozen
catalog, typed source registry, implementation binding, and target generation
before persistent changes; clears the four tables in one caller-controlled
publication transaction; enumerates active declarations with selected results;
reconstructs each result from the restored immutable source boundary; and
requires exact result, digest, vertex, and edge identity before readiness.
Leases are reconciled from durable Reporting outcomes and are never copied as
authoritative backup rows.

The result status is `succeeded/ready` or `failed/incomplete`. Failure claims no
committed cleared/rebuilt facts and returns only closed safe errors. An
indeterminate commit requires target reinitialization under Core 04.

The current binary packages and accepts only the v2 source registry and v2
implementation binding. The rollout-only historical empty-registry v1 binding
is retired after fresh-v2 backup and isolated-restore evidence; no backup
version, digest dispatcher, generated artifact, or runtime branch may admit it.

## 10. Security and operational invariants

- The engine is authorization-, transport-, storage-, and job-system agnostic.
- Caller-owned source adapters must validate authorization before constructing
  semantic input; Graph Projection must not query authoritative owner tables.
- Computation occurs outside publication transactions. Publication transactions
  contain only validation, immutable inserts, and caller-owned compare-and-swap.
- Cleanup is lease- and reachability-aware and fails closed on uncertain owner
  state.
- Telemetry may contain schema/version, counts, duration, result ID, safe error
  codes, and owner/job IDs already classified safe; it must not contain source
  values, labels, properties, mapped metadata, raw JSON, SQL, or secrets.

## 11. Definition of Done

Graph Projection v2 is conformant only when all criteria pass:

1. Strict decoding rejects removed, unknown, duplicate, malformed, and null
   members without partial output.
2. Golden canonicalization and length-framing fixtures reproduce every digest
   and identifier byte for byte.
3. Retry, process restart, and Recovery reproduce identical result, vertex, and
   edge identities.
4. Mapping, aggregation, direction, reverse-edge, filters, defaults, nulls,
   wildcard expansion, schema registry, ordering, and limits have owner-routed
   tests.
5. Cancellation is observed at every required checkpoint.
6. The pure engine imports no PostgreSQL, HTTP, auth, Common Jobs, Recovery, or
   application-composition package.
7. Atomic publication, exact reads, traversal, leases, cleanup, same-ID byte
   conflict, rollback, and cancellation have service-backed evidence.
8. Current runtime contains no v1 parser, run state, retention, cursor,
   idempotency, invalidation, broad repository, public Graph route, or hidden
   worker.
9. Recovery reconstructs exact selected results before readiness and the rollout
   bridge is absent from the final current binary.

## Appendix A. Design rationale

Attempts and immutable content have different identities. Common Jobs therefore
owns attempts, while content-derived `projection_result_id` names the reusable
result. Saved graph declarations belong to their source owner because only that
owner can define authorization, source validity, refresh intent, and product
lifecycle. Narrow result ports keep Graph Projection reusable without turning it
into a second application platform.
