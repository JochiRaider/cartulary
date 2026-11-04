---
title: Graph Projection NLSpec
status: adopted/current
document_class: nlspec
created_at: 2026-05-30
document_version: 2.2.1
---

## 1. Status, scope, and authority

Version `2.2.1` is adopted/current under the 2026-10-31 decision in Core 00
§1.7. Extension-source reconstruction follows successful claim; Graph v2
computation/publication and current-only rebuild v5 remain current.

The frozen-input lifetime amendment adopted on 2026-11-02 removes live Graph
leases; consumer execution uses its retained admission. Earlier adoption records
do not make retired behavior current.

This NLSpec defines Graph Projection 2.x: deterministic, side-effect-free
derivation of a graph representation from an explicit immutable source.
Version 2.2.0 adopts one native-v2 engine entrypoint, the complete closed v2
semantic matrix, a pure root package, a Graph-owned restore subpackage, and a
current-only Recovery v5 contribution. It replaces Graph Projection 1.2.0 and
the operational compatibility posture of 2.1.0 in full. Version 1 lifecycle,
query, retention, cursor, idempotency, invalidation, run-history, historical
Recovery dispatch, and five-table Recovery behavior are not current behavior
and MUST NOT be implemented through a compatibility alias, translator, dual
reader, dual writer, fallback dispatcher, or hidden feature flag.

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
| Admission lock | Transaction-scoped lock protecting an exact result against cleanup while a consumer admits its immutable contribution. |
| Retained consumer input | Complete required immutable contribution retained under the consumer's admission contract, independently of later live Graph availability. |

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
| `finite_number` | JSON number representable by the adopted runtime numeric representation whose decoded and computed value is finite; NaN, infinity, and overflow are invalid. |
| `property_value` | Null, boolean, finite integer, finite number, string, or a flat array of those scalar values; nested objects/arrays are invalid. |

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
Each emitted edge derives its properties and mapped metadata from the original
source relationship using its own output kind and ID. Wildcards apply to both
kinds. Required, missing, and null policies apply independently to each edge.

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

Definition paths are admitted against every configured producer input context
for their target scope and kind, including wildcard expansion and reverse-edge
outputs. Direct mappings read their source entity or relationship; aggregations
read their declared input scope. Validity does not depend on source cardinality,
filters, or which producers emit objects in a particular invocation. A path
invalid in any applicable context rejects before missing/default handling.
Definitions without an applicable producer remain inert but must have valid
source/projected path syntax. Graph-view definitions read only source metadata.
Metadata mappings follow the same applicability rules. Different producer
contexts needing different definitions must use distinct output kinds.

The valid projected-type and merge-behavior pairs are closed:

| Projected type | Valid merge behaviors |
| --- | --- |
| `boolean` | `single_value`, `first`, `last` |
| `integer` | `single_value`, `first`, `last`, `min`, `max`, `sum`, `count` |
| `number` | `single_value`, `first`, `last`, `min`, `max`, `sum` |
| `string`, `timestamp`, `identifier` | `single_value`, `first`, `last`, `min`, `max` |
| `boolean_array`, `integer_array`, `number_array`, `string_array`, `timestamp_array`, `identifier_array` | `single_value`, `first`, `last`, `set`, `ordered_list` |

Every other pair is invalid and MUST fail admission before derivation.
`single_value` requires canonical equality across all admitted contributors.
`first` and `last` select in the canonical contributor order from §5. `min`
and `max` compare integers and numbers numerically and compare strings,
timestamps, and identifiers by their adopted canonical byte order. `sum`
detects signed-int64 overflow for integer output and rejects a non-finite
number result. `count` emits the integer count of admitted integer values.
`set` flattens exactly one array level, deduplicates elements by canonical
bytes, and emits canonical byte order. `ordered_list` concatenates arrays in
canonical contributor order and preserves duplicates. Empty, singleton, null,
missing, and defaulted contributor sets follow the declared
missing/null/default policies before merge selection; a failure emits no
partial object or result.

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

Aggregation producer lookup is keyed by `(target_scope, projected_kind)`;
vertex and edge kind namespaces are distinct. A projected input depends on
every other aggregation producer of its input kind in the matching scope.
Endpoint references add dependencies on the named vertex aggregation rules;
missing references and references to edge rules are invalid. Repeated references
to one producer represent one dependency. A rule never depends on its own
newly emitted output; matching direct input remains available. Ready rules are
processed in complete batches sorted by `aggregation_rule_id`, then the next
ready batch is selected. Duplicate rule IDs and cycles are invalid.

## 4. Admission, normalization, and canonical bytes

Admission performs decoding, closed-shape validation, default materialization,
scalar validation, duplicate detection, resource admission, mapping validation,
and source-reference validation in that order. Failure returns a closed
`ProjectionError` and no partial result.

One shared semantic validator owns signed-int64 integers, finite numbers,
identifiers, timestamps, flat arrays, field paths, labels, strings, defaults,
filters, and merge operands. Admission, default validation, filters,
canonicalization, merge evaluation, and execution MUST use the same rules.
Every owner-bounded textual ceiling is measured in UTF-8 bytes after decoding,
never Unicode scalar count, code-unit count, or locale-dependent width. Exact
field paths use `source_entity_kind`, `source_relationship_kind`, the other
Table 3-C and Table 3-D names, and declared `properties.` or `metadata.` paths;
the aliases `kind`, `op`, and `graph_view_key` are invalid.

The source-kind paths resolve only against their corresponding source object,
including in filters, inclusion predicates, definitions, and aggregation or
endpoint grouping where that source context is admitted. Bare `kind` is invalid
in every field-path position. Existing projected object ID, endpoint ID,
direction, property and mapped-metadata paths remain supported in their
respective contexts. There is no projected-kind replacement path:
`projected.vertex_kind` and `projected.edge_kind` are invalid. Invalid paths
fail admission; missing/null/default policies cannot make them valid.

Objects are serialized as UTF-8 canonical JSON: keys in Unicode code-point
order, no insignificant whitespace, shortest valid JSON escapes, exact JSON
literals, and normalized finite numbers. Arrays preserve semantic order except
arrays explicitly declared sorted. Object-map iteration and filesystem order
MUST NOT affect bytes.

The v2 configuration, source, result envelope and child rows retain the same
number spelling. Integers use decimal digits. Finite binary64 values use the
shortest round-tripping significand with general-format exponent selection
(equivalent to Go `FormatFloat(f, 'g', -1, 64)` with the exponent plus sign
removed). Decoded negative zero normalizes to zero. For example, the float
`1000000.0` serializes as `1e06` and the integer `1000000` as `1000000`.
The integer-equivalence rule in §5 applies only to grouping and endpoint tuples.

`normalized_configuration_sha256` is SHA-256 of the canonical materialized
`projection_config`, `filters`, and `property_definitions`. It excludes source
objects, source metadata, trusted context, and operational data.

`normalized_source_sha256` is SHA-256 of canonical `source_snapshot_id`, sorted
source entities, sorted source relationships, and `source_metadata`. It excludes
configuration, trusted context, and operational data.

## 5. Projection semantics

The engine evaluates cancellation before admission, after every 1,024 source or
output items, between direct and aggregated phases, and before returning.
Once observed, cancellation is terminal for that invocation. A checkpoint
callback stop, parent cancellation or deadline yields `projection_cancelled`,
`context_cancelled`, `do_not_retry` and no result.

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

Every contributor to an edge group MUST supply the same ordered pair of
canonical endpoint tuples before its fields are merged or provenance emitted.
Non-scalar supplied members fail validation, even alongside missing members.
Distinct complete pairs fail with `projection_validation_failed`,
`aggregation_endpoint_conflict` and `do_not_retry` under either missing policy.
After checking supplied values and agreement, any missing member or unmatched
vertex applies the missing-endpoint policy to the entire group. Null is a
present scalar. Neither validation nor exclusion depends on contributor order.

Grouping and endpoint grouping share one fallible scalar-tuple encoder and the
shared signed-int64/finite-number normalization rules. Exactly integral finite
numbers within signed-int64 range have the same canonical decimal encoding as
the corresponding integer; int64 values never round through binary64. Numeric
lexemes are not string keys. Null is a scalar; missing values follow the declared policy.
Arrays and objects are invalid grouping members. Encoding failure MUST NOT
produce a grouping digest or successful endpoint lookup. Invalid grouping
values fail closed with no result; diagnostics never expose their values.

Contributor ordering uses immediate provenance, comparing these four UTF-8
byte-string components lexicographically:

| Contributor | Kind | ID | Producer rule | Prior projected ID |
| --- | --- | --- | --- | --- |
| Source entity/relationship | Source kind | Source ID | Empty | Empty |
| Direct/reverse projected object | Referenced source kind | Referenced source ID | Mapping rule ID | Projected object ID |
| Aggregated projected object | Own projected kind | Own projected ID | Aggregation rule ID | Own projected ID |

Aggregated ancestry is not recursively traversed. Property merges, metadata
merges and source-reference arrays consume the same order. The string
`contributor_sort_key` is `csk_` followed by the four lowercase UTF-8 hex
components joined with `.`; its byte order matches the tuple order. Hashes are
not contributor comparison keys. Object/result identity framing and final
vertex/edge sorting remain unchanged.

The admitted dependency order governs one execution pass across both target
scopes. Each rule gathers its contributors before emitting any of its own
objects and executes exactly once. Completed output becomes available to later
rules, including vertex-to-edge and edge-to-vertex dependencies. An implementation
MUST NOT execute all vertex rules before edge rules independently of this order.

Projected values are decoded once into their native v2 type before validation,
comparison, merge, or canonicalization. Integer operations use the complete
signed-int64 domain. Number operations reject overflow and any non-finite
intermediate or result. Array operations accept only flat arrays of the exact
declared element type. Canonical equality and ordering are applied identically
in ordinary execution and Recovery.

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

Vertex families are `direct` and `aggregated`. Edge families are `direct`,
`reverse` and `aggregated`. A reverse edge is directed, sets `is_reverse_edge`
to true and names its directed direct edge in `reverse_of_edge_id`. Both edges
share the typed source relationship and mapping rule and exchange endpoints.
Other families carry a false reverse flag and null reverse reference.

## 7. Errors, validation, and limits

`ProjectionError` contains `code`, `reason_code`, `retry_action`, and closed
safe `details`. Codes are `invalid_projection_request`,
`projection_validation_failed`, `projection_resource_limit_exceeded`,
`projection_cancelled`, and `projection_computation_failed`. Retry action is
`do_not_retry` except an infrastructure computation failure may be
`retry_with_backoff`. Source property values, metadata values, stack text, SQL,
paths, and secrets MUST NOT appear in errors, logs, metrics, or traces.

The owner-authored semantic registry contains exactly these non-configurable
limits and the cancellation cadence in §5. Machine contracts project the
registry as one object; the engine keeps one private runtime representation
whose exact equality with that projection is tested. Cursor, list, history,
job, retry, and deployment-capacity limits are not Graph semantic limits.

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
| Flat array elements per property or mapped metadata value | 100,000 |
| Traversal seed vertices / kind filters / depth | 1,024 / 1,024 / 16 |

Limit checks use bounded counting and report at most `limit + 1` for count
failures. Projection returns no partial graph on any validation or limit error.

One issue budget covers admission and every derivation phase. Exactly 100,000
issues retain ordinary deterministic validation reporting. Attempting to record
another stops processing before allocating that issue. Array mergers check
capacity before retaining another output element. Ordered lists count every
element and sets count distinct canonical elements. The first excess returns
`projection_resource_limit_exceeded` and `do_not_retry`, with the respective
`maximum_validation_issues` or `maximum_array_items` limit key, limit 100,000
and observed count 100,001. Failed builders publish no object to later rules.

Projected vertex and edge counts are cumulative across direct, reverse and
aggregated emission. Before constructing or retaining the next emitted object,
the engine rejects an attempted count above its limit, reports `limit + 1`, and
stops subsequent rule execution. Filtered or excluded objects consume no output
capacity. Exhaustion returns `projection_resource_limit_exceeded`, with the
`maximum_projected_vertices` or `maximum_projected_edges` reason/limit key and
`do_not_retry`; it is not a retryable computation failure. Final output checks
are defensive and do not replace emission-time enforcement.

## 8. Narrow persistence and consumer ports

Graph Projection defines behavior, not a repository facade. An adapter MAY
implement only these typed capabilities:

- `PublishResultTx`: insert one immutable result envelope, all vertices and edges
  through a caller-owned transaction;
- `ReadExactResult`: read one result by full binding tuple;
- `ReadVertices` and `ReadEdges`: return canonical output order;
- `Traverse`: bounded breadth-first traversal, breaking ties by output order;
- `LockResultEnvelope`: lock an exact immutable result against cleanup through a
  borrowed transaction;
- `LockCleanupCandidateTx`: lock at most one oldest result for an exact source
  owner by publication time then result ID, using skip-locked semantics;
- `DeleteLockedResultTx`: delete at most the locked envelope and cascading
  objects after the caller's same-transaction selected-binding recheck.

Publication validates all digests, counts, IDs, endpoints and order before the
caller commits. Partial publication is forbidden. Exact admission validates family and reverse
linkage independently of row order, plus the array ceiling in properties and
mapped metadata. Exact reads retain the original v2 envelope and row bytes. Adapters borrow their handles,
start no hidden workers and close no borrowed resources. The current contract
has no Graph lease model, persistence, acquisition, renewal, expiry or release.

### 8.0 Exact admission and retained consumer input

A consumer needing a result beyond its admission transaction MUST retain the
complete required immutable contribution under its own current admission
contract. The adapter MUST support an exact binding read through a borrowed
transaction that locks the result envelope against cleanup. All ordered objects
and canonical bytes MUST belong to the same admitted result. Result locking
precedes source declaration locking. An earlier read or unretained result ID
does not establish continued availability.

Cleanup locks at most one candidate using the existing owner, age, ordering and
skip-locked rules, then rechecks the source owner's selected binding in that
transaction. It MUST NOT delete a selected result or bypass an in-flight
admission's result lock. Unselected results remain subject to the result
retention policy. Private retained consumer inputs are not reachable Graph
results and MUST NOT be deleted by Graph cleanup.

### 8.1 Implementation and construction boundary

The production root package `internal/modules/graphprojection` owns only the
pure deterministic engine and completed-result model. It imports only the Go
standard library and its own private implementation. Its sole engine
entrypoint is:

```go
func ProjectV2(
    ctx context.Context,
    invocation InvocationContextV2,
    rawInput []byte,
) (ProjectionResultV2, error)
```

The root exposes no engine type, engine constructor, mutable option, graph-view
identity helper, restore contract, database capability, job capability, or
cross-owner alias. Parsing, normalization, validation, derivation, limits,
canonicalization, and projection-result identity helpers remain private unless
an independently adopted cross-package production consumer requires a narrower
capability.

`internal/modules/graphprojection/restore` owns Graph restore contracts,
current source-registry construction, rebuild orchestration, and the pure
Graph recovery-state contribution. PostgreSQL realization remains in
`internal/modules/graphprojection/postgresrestore` and depends on the narrow
restore-owned ports. Recovery may expose or alias only the minimal cross-owner
port required to invoke the participant. Constructors are closed and fallible:
the PostgreSQL restore writer requires both its database and its reconciler,
and no valid composition can omit reconciliation. Positive import, export,
adapter-capability, constructor, and application-composition allowlists define
the supported topology; permanent tests MUST NOT enumerate every historical
deleted symbol or package.

Cleanup capabilities MUST NOT accept a deployment-wide reachable-ID list or
query a source owner's declaration table. The caller supplies one exact source
owner and performs its selected-binding check through that owner's adapter in
the same borrowed transaction. Candidate selection uses
`FOR UPDATE SKIP LOCKED`; one result transaction deletes at most one envelope.
Result, publication-replay and consumer-admission paths that can race cleanup
acquire the result row before any source-owner declaration row. This lock order
and the same-transaction selected-binding recheck are mandatory. Lease-only
configuration, counters and telemetry are removed; controls still used by result
cleanup remain current.

## 9. Recovery

Graph Projection publishes only current algorithm
`graphprojection.restore_rebuild.v5` and result schema
`cartulary.graph_projection_restore_rebuild_result.v5`. Its matching v5 source
registry and v7 implementation binding admit only Network Flow semantic-query v2
declarations and rebuild Graph Projection v2 results. Earlier versions of each
individual algorithm, result, binding and registry are unsupported and absent
from the active build, including historical codecs, aliases and dispatchers. Recovery rejects
an old Graph binding during selection or binding admission before opening the
Graph mutation transaction or invoking Graph clear, rebuild, reconciliation,
or publication. No inventory gate, translator, digest heuristic, fallback, or
historical dispatcher is permitted. The current Recovery catalog contains
exactly these derived tables in ASCII order:

```text
graph_projection_result_edges
graph_projection_result_vertices
graph_projection_results
```

Recovery structurally restores source-owner authoritative declarations and leaves
these derived tables empty. It MUST NOT invoke an inactive profile source
adapter, enumerate or interpret Network Flow semantic queries, or reconstruct
Graph results during structural restore. Catalog and current codec identity
admission still precedes any target mutation. This explicit extension-derived
exception is governed by Core 01 REQ-01-424 and EXT-REQ-235; it does not remove
independent Base projection rebuild requirements.

After successful claim compatibility, migration where supported, final semantic
validation and publication admission, the application MUST perform the current
Graph reconstruction for that claimed source profile. The admitted restored
generation and retained source declarations identify the work; no historical
reader or automatic source-state conversion is introduced. Reconstruction is
idempotent and restartable under the same exact generation and source binding.
The runtime attempts reconstruction once per successful process claim activation;
failure does not trigger an unbounded in-process retry loop. A later activation
may resume from the same authoritative facts under current identity checks.
While it is pending or failed, graph-dependent operations MUST fail closed as
unavailable, and dependent durable jobs MUST remain gated without fabricated
success or loss of their immutable admitted input. These gates cover Network
Flow materialization and new Reporting admission requiring live Graph results,
not Reporting work whose complete input is already retained. Unrelated Base and
admitted non-graph operations remain available. A failed source claim leaves its derived
state absent and never invokes its source adapter.

Reconstruction of a current restored installation uses the original operation
and target-generation identities from the admitted database control record,
together with the current catalog, source-registry and implementation-binding
definitions and the successful source claim. It MUST NOT require a Recovery
decryption key, an unexpired marker, source backup storage, or a fabricated
consistency timestamp. Source enumeration and reconstruction use the same
caller-controlled transaction over the admitted authoritative state. Result
publication time is operational state and does not change logical result
identity or canonical output digests. A new initialized installation follows
ordinary Graph admission; it has no restored generation to reconstruct.

Unavailable operations use their existing owner-defined unavailable/error
contracts; this scheduling change introduces no new public reason or job status.

During reconstruction, affected graph workers and dependent job consumers remain
quiescent. The participant validates the frozen catalog, typed v5 source
registry, v7 implementation binding, admitted claim and restored generation
before persistent changes; then, in one caller-controlled
transaction, clears the three tables, enumerates active v2 declarations with
selected results, reconstructs each result from the restored immutable source
boundary, publishes immutable results, reconciles Network Flow nonterminal
materialization jobs and verifies exact result, digest, vertex, edge, job and
readiness postconditions. Reconciliation is mandatory before dependent gates
open and cannot be omitted by construction. Reporting jobs and immutable inputs
are outside this derived reconstruction transaction; ordinary Core Jobs recovery
and Reporting retained-input validation remain mandatory. Neither may invoke an
unclaimed source adapter. Structural completion and Graph availability are
separate outcomes.

The current result retains its existing closed fields except that
`reconciled_lease_count` is removed entirely. Pure projection input/result v2 and
their canonical identities remain unchanged.

The result status is `succeeded/ready` or `failed/incomplete`. Failure claims no
committed cleared/rebuilt facts and returns only closed safe errors. An
indeterminate commit invokes Core 04 fatal-integrity shutdown and keeps the
affected admission gates closed. It never triggers automatic deletion or reset.

The current implementation binding is the closed object
`cartulary.graph_projection_restore_implementation_binding.v7` defined below.

| Required member | Contract |
| --- | --- |
| `schema_id` | `cartulary.graph_projection_restore_implementation_binding.v7` |
| `algorithm_id` | `graphprojection.restore_rebuild.v5`, resolved to the actual typed reconstruction implementation |
| `binding_id` | `graphprojection.restore_rebuild.network_flow_graph_views.v5` |
| `projection_schema_id` | `graph_projection.v2`, resolved to the actual typed projection input/result implementation |
| `semantic_query_schema_ids` | Exactly the sorted semantic query schema set supported by the current typed source registrations |
| `recovery_state_catalog_sha256` | Exact selected frozen Recovery catalog digest |
| `source_registry_sha256` | Exact selected frozen Graph source-registry digest |
| `graph_table_ids` | Exact derived Graph table set, in ASCII order |

No other members are admitted. Informational document-version labels,
independent engine-stage algorithm lists/digests, packaged-subject/build
provenance and standalone database lineage/head fields are absent from this
logical compatibility binding. Core owns database generation admission. Actual
build attestation remains a separate release obligation. Admission MUST verify
the supported binding schema, actual rebuild algorithm and actual projection
schema independently of declaration agreement, before any Graph mutation or
publication. Definition bytes MUST be decoded strictly: duplicate members,
invalid Unicode, unknown members and trailing JSON are invalid. The admitted
complete object MUST match its selected canonical definition digest.

The current-only Recovery generation MUST select binding v7 and its exact
digest. Binding v6 and earlier, removed members, unsupported algorithm/schema
identities and stale generation pairs reject before mutation. No converter,
alias, historical reader, automatic deletion or control-record rewrite is
permitted. Pure projection identities and canonical bytes are unchanged.

The current source registry is v5 and implementation binding is v7. No earlier
Graph generation is selectable or executable. Failure or cancellation in any
clear, rebuild, publication, reconciliation, or postcondition phase rolls back
the complete Graph mutation. An indeterminate commit keeps the target
unavailable under Core 04; any operator recovery is separate from this protocol.

The current catalog removes the retired lease table. Existing migrations MUST
remain unchanged. A new Core-owned migration may remove that table only when it
is empty, checked under a table lock in the same transaction. Nonempty state
MUST reject before mutation; migration MUST NOT delete or expire rows to pass.
Old backup generations reject before restore mutation; old completion cannot
authorize current startup and its control tuple MUST NOT be rewritten. No reader,
converter, alias, implicit claim or automatic reset is introduced.

## 10. Security and operational invariants

- The engine is authorization-, transport-, storage-, and job-system agnostic.
- Caller-owned source adapters must validate authorization before constructing
  semantic input; Graph Projection must not query authoritative owner tables.
- Ordinary computation occurs outside publication transactions, whose work is
  validation, immutable inserts and caller-owned compare-and-swap. The §9
  post-claim reconstruction is the explicit exception: source enumeration,
  computation, publication and reconciliation share one transaction.
- Cleanup is admission-lock- and reachability-aware and fails closed on uncertain owner
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
6. The pure root imports only the standard library and its own private engine
   implementation; restore ownership and every PostgreSQL, HTTP, auth, Common
   Jobs, Recovery, sibling-module, and application dependency remain outside it.
7. Atomic publication, exact reads, traversal, admission locking, cleanup, same-ID byte
   conflict, rollback, and cancellation have service-backed evidence.
8. Current runtime contains no v1 parser, run state, retention, cursor,
   idempotency, invalidation, broad repository, public Graph route, or hidden
   worker.
9. Recovery v5 reconstructs exact v2 selected results, requires reconciliation
   and postcondition verification before readiness, and rejects every Graph
   earlier rebuild artifact before mutation.
10. Every projected-type/merge pair, signed-int64 and finite-number boundary,
    UTF-8 byte ceiling, null/default/wildcard branch, contributor permutation,
    cancellation checkpoint, and failure-atomicity branch has deterministic
    owner-routed evidence sourced from machine contracts rather than Markdown.
11. Reverse field selection, reverse linkage and publication, historical numeric
    bytes, endpoint unanimity, issue and merge exhaustion, and one-shot callback
    cancellation have public-entrypoint evidence. Reverse reconstruction and
    exact stored reads have PostgreSQL evidence.

## Appendix A. Design rationale

Attempts and immutable content have different identities. Common Jobs therefore
owns attempts, while content-derived `projection_result_id` names the reusable
result. Saved graph declarations belong to their source owner because only that
owner can define authorization, source validity, refresh intent, and product
lifecycle. Narrow result ports keep Graph Projection reusable without turning it
into a second application platform.
