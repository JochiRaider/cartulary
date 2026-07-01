---
title: Graph Projection NLSpec
status: draft
document_class: nlspec
created_at: 2026-05-30
---

## 1. Purpose and authority

This draft NLSpec defines the proposed contract for **Graph Projection**: the deterministic derivation of graph-oriented representations from authoritative source data.

Status note: this document is `status: draft`. It is not adopted implementation-conformance authority unless a later authority update promotes it. It governs only graph-oriented projection concepts in this document and MUST NOT be used to reinterpret workbook-grid projection tables, `view_row_v1`, workbook query routes, saved views, import owner facades, or projection-table rebuild behavior owned by Core 00 through Core 04.

This NLSpec controls:

- projection input shape, normalization, defaults, resource bounds, and validation;
- source-to-graph mapping rule schemas and mapping algorithms;
- projected graph view, vertex, edge, property, metadata, validation, and schema-registry output shape;
- generated graph-view, projection-run, vertex, edge, sort-key, digest, and validation-issue identity;
- aggregation grouping, source-reference preservation, property merge behavior, endpoint derivation, and deterministic ordering;
- projection lifecycle, failed-run inspection, replacement, invalidation, and exact retention behavior;
- consumer query contracts, traversal behavior, pagination, result shape, and query errors;
- the intentional implementation latitude that remains below the observable boundary.

Authoritative source data remains the source of truth. A projected graph view is a derived representation. A projected vertex or projected edge MUST NOT be treated as an authoritative source record unless the consumer follows its source reference back to the authoritative source data.

A completed projection is authoritative only for the derived graph representation it emits. Graph consumers MAY rely on a completed projection for deterministic vertex IDs, edge IDs, traversal behavior, output ordering, metadata, validation summary, schema registry, and source-reference traceability within that projection result. Graph consumers MUST NOT infer that projected graph state supersedes, mutates, or replaces authoritative source data.

A conforming implementation MUST satisfy this NLSpec at every observable boundary, regardless of storage engine, graph engine, programming language, database, transport protocol, scheduling mechanism, or visualization mechanism.

## 2. Scope and non-scope

This NLSpec owns the following graph projection concerns.

| Concern | Scope statement |
| --- | --- |
| Projection inputs | Defines complete input objects, member requirements, nullability, defaults, canonicalization, validation, and resource limits. |
| Projection outputs | Defines graph-view, vertex, edge, schema-registry, metadata, validation-summary, consumer-capability, and canonical-output shape. |
| Projection configuration | Defines source-kind declarations, mapping rules, aggregation rules, filters, property definitions, metadata mappings, retention policy, and custom-configuration boundaries. |
| Source entity mapping | Defines how eligible source entities become direct projected vertices. |
| Source relationship mapping | Defines how eligible source relationships become direct projected edges. |
| Aggregation | Defines how eligible contributors become aggregated vertices or aggregated edges. |
| Projected identity | Defines graph-view, projection-run, vertex, edge, sort-key, digest, and validation-issue identity. |
| Property derivation | Defines source compatibility, defaults, omissions, explicit null handling, and aggregation merge behavior. |
| Metadata derivation | Defines metadata mapping, metadata output schemas, and metadata privacy boundaries. |
| Inclusion and omission | Defines filtering, source-item exclusion, omitted fields, unavailable source fields, empty-input behavior, and omitted-output behavior. |
| Validation | Defines validation phases, severity, issue shape, issue identity, ordering, closed code registry, and code-specific details. |
| Lifecycle | Defines graph-view state, projection-run/result state, transitions, failure behavior, replacement, invalidation, and exact retention. |
| Consumer contract | Defines query shapes, traversal semantics, result shapes, pagination, defaults, bounds, and error behavior. |

Adjacent concerns outside this NLSpec are boundary concerns only.

| Concern | Boundary |
| --- | --- |
| Workbook-grid projections | Workbook projection tables, workbook row refreshes, `view_row_v1`, workbook query/sort/filter/group semantics, saved views, import owner facades, and restore projection rebuilds remain governed by Core 01/Core 03 and their generated view-schema contracts. |
| Authoritative source-data schema | This NLSpec consumes source entities and source relationships through the input contract. It does not define the source system's internal schema. |
| Source-data mutation | Projection is read-derived. This NLSpec does not define create, update, delete, merge, or authorization behavior for source data. |
| Storage layout | Implementations MAY store projection inputs, outputs, indexes, and caches in any layout that preserves the observable contract. |
| Transport protocol | Query shapes are contract-level interfaces. This NLSpec does not require HTTP, RPC, SQL, message queues, files, or any other transport. |
| Graph query language | This NLSpec defines consumer query shapes and traversal behavior. It does not adopt an external graph-query language. |
| Visualization | This NLSpec does not define graph rendering, layout, colors, node shapes, edge styling, or UI interaction. |
| Scheduling | Projection MAY execute synchronously, asynchronously, in batches, incrementally, or on demand, provided lifecycle and consumer guarantees are satisfied. |
| Authorization | Access control is outside this NLSpec. A caller that is allowed to consume a graph view receives the behavior defined here. |
| External graph engines | Implementations MAY use a graph database, relational database, document store, in-memory index, or no graph engine, provided outputs and queries satisfy this NLSpec. |

## 3. Normative language and core concepts

The key words **MUST**, **MUST NOT**, and **MAY** are normative. **MUST** and **MUST NOT** define conformance requirements. **MAY** defines optional behavior whose omission semantics are explicit. The word **default** defines the value or behavior applied when a member is omitted and the member permits omission.

A conforming implementation MUST treat object member names, enum tokens, generated prefixes, validation codes, query error codes, and lifecycle state names as exact code point sequences.

The following terms have exactly the meanings in this table.

| Concept | Definition |
| --- | --- |
| `projection` | A deterministic derivation process that reads authoritative source data and emits a graph-oriented representation. |
| `source entity` | An authoritative object eligible to become a projected vertex. |
| `source relationship` | An authoritative relationship fact eligible to become a projected edge. |
| `direct projected vertex` | A projected vertex emitted from one eligible source entity through one entity mapping rule. |
| `direct projected edge` | A projected edge emitted from one eligible source relationship through one relationship mapping rule. |
| `aggregated projected vertex` | A projected vertex emitted from a declared aggregation group. |
| `aggregated projected edge` | A projected edge emitted from a declared aggregation group and endpoint grouping rule. |
| `projected property` | A derived key-value member attached to a projected vertex, projected edge, or graph view. |
| `mapped metadata` | Metadata emitted by `metadata_mappings[]` under `metadata.mapped_metadata`. |
| `system metadata` | Metadata members defined by this NLSpec, such as mapping rule IDs, aggregation source references, digests, and reverse-edge links. |
| `graph view` | A named projected graph representation produced under one graph-view key and one projection schema. |
| `projection identity` | The canonical tuple family used to identify a graph view and its emitted vertices and edges. |
| `projection refresh` | A lifecycle operation that computes a new projection run for an existing graph view. |
| `source snapshot` | A declared immutable read boundary over authoritative source data used as the input basis for one projection run. |
| `projection run` | One accepted execution of projection for one graph view, one source snapshot, and one normalized projection configuration. |
| `projection result` | The retained output, validation summary, lifecycle state, and metadata produced by one projection run. |
| `projection version` | A non-empty identifier carried by projection configuration and copied to projection output. |
| `mapping rule` | A declared deterministic rule that maps one source entity kind or source relationship kind to one projected vertex kind or projected edge kind. |
| `aggregation rule` | A declared deterministic rule that groups contributors and emits aggregated vertices or edges. |
| `source relationship direction` | One of `forward`, `reverse`, `undirected`, or `bidirectional`. |
| `projected edge direction` | One of `directed`, `undirected`, or `bidirectional`. |
| `traversal` | A consumer operation that walks emitted vertices and emitted edges in a completed graph view according to direction, depth, and kind filters. |
| `filter` | A deterministic inclusion predicate applied during projection. |
| `consumer` | A caller or subsystem that reads completed graph views, projected vertices, projected edges, validation summaries, metadata, schema registries, or traversal results. |

## 4. Projection input contract

A projection input object is a JSON-compatible object with the members defined in this section. A conforming implementation MAY receive the object through any transport, but after decoding it MUST validate and normalize it according to this NLSpec before projection begins.

### 4.0 JSON decoding, schema-table notation, and canonical input paths

A projection input received as bytes MUST be decoded as UTF-8 JSON before schema validation. Invalid UTF-8, invalid JSON syntax, duplicate JSON object members, a decoded top-level value that is not a JSON object, or any other condition classified as pre-admission by §4.0.1 MUST fail through the lifecycle operation error contract in §10.0. A pre-admission failure MUST NOT allocate `projection_run_id`, MUST NOT emit a validation issue object, MUST NOT emit `issue_id`, and MUST NOT create retained run state.

A conforming implementation MUST reject duplicate JSON object member names at every object depth before schema validation. Duplicate member handling MUST NOT depend on parser first-wins, last-wins, insertion-order, or map-overwrite behavior. A duplicate member MUST return lifecycle operation error `invalid_projection_request` with `reason_code=duplicate_object_member`; `details.field` MUST be the canonical input path to the duplicate member when the path is attributable.

Schema tables use `[]` as type notation, not as a literal JSON member-name suffix. A table member written as `source_entities[]` denotes the JSON member named `source_entities` whose value is an array. The canonical input path for that member is `$.source_entities`.

Canonical input paths MUST use this grammar.

| Path part | Required syntax |
| --- | --- |
| Root | `$` |
| Object member with an ASCII path identifier | `.` followed by the exact member name |
| Object member without an ASCII path identifier | `[` + canonical JSON string for the exact member name + `]` |
| Array element | `[` + zero-based base-10 index with no leading zeroes except `0` + `]` |

An ASCII path identifier is a non-empty string whose first character is `A-Z`, `a-z`, or `_`, and whose remaining characters are `A-Z`, `a-z`, `0-9`, or `_`. Canonical input paths are diagnostic identifiers only. They do not change the field-path grammar in §4.2.

The canonical JSON string used inside canonical input paths for non-ASCII-path object member names MUST use the canonical string serialization grammar in §5.8.

Validation detail members named `field` MUST use canonical input path syntax unless the owning detail schema explicitly says the member carries a §4.2 `field_path`.

### 4.0.1 Projection request admission boundary

Run admission is the boundary between lifecycle operation errors and projection-run validation. A conforming implementation MUST classify every projection request condition according to this table before constructing any validation issue.

| Phase family | Required outcome | Run created? | Validation issues emitted? | Public error family |
| --- | --- | ---: | ---: | --- |
| Operation request envelope malformed | Reject before projection input processing | No | No | Lifecycle operation error |
| Invalid UTF-8 or invalid JSON syntax in `projection_input` bytes | Reject before run admission | No | No | `invalid_projection_request` |
| Duplicate JSON object member at any depth | Reject before run admission | No | No | `invalid_projection_request` |
| Top-level decoded value is not an object | Reject before run admission | No | No | `invalid_projection_request` |
| Missing, unknown, or explicit-null top-level member required to derive config or source identity | Reject before run admission | No | No | `invalid_projection_request` |
| Invalid `projection_schema_id` | Reject before run admission | No | No | `invalid_projection_request` |
| `projection_config` cannot be default-materialized or normalized deterministically | Reject before run admission | No | No | `invalid_projection_request` |
| Supplied `graph_view_id` is malformed or does not equal the derived graph-view ID | Reject before run admission | No | No | `invalid_projection_request` |
| Whole-input resource limit prevents deterministic decoding, default materialization, or normalization | Reject before run admission | No | No | `invalid_projection_request` |
| Mapping, filter, property, metadata, aggregation, source-item, derivation, or output validation after admission | Admit run, then complete as `available` or `failed` | Yes | Yes | Run inspection through `get_projection_run()` |
| Computation failure after admission | Admit run, then fail it | Yes | Yes, with `projection_computation_failed` | Run inspection through `get_projection_run()` |

A table in §4 that names a validation code for a pre-admission condition defines the corresponding `invalid_projection_request.details.validation_code`; it does not require a validation issue object. Validation issue objects exist only for admitted runs.

A pre-admission rejection MUST NOT create graph-view state, retained run state, projected output, validation summary, or an idempotency record unless §10 explicitly requires a record for that operation family. In this revision, §10 creates idempotency records only after run admission or successful invalidation target validation.

If admission succeeds, the implementation MUST fix `graph_view_id`, `projection_run_id`, `source_snapshot_id`, `projection_version`, `accepted_at`, `projection_config_digest`, and `projection_source_digest` before constructing any validation issue for that run.

### 4.1 Common scalar contracts

The following scalar contracts apply wherever referenced.

| Contract | Definition |
| --- | --- |
| `identifier` | A JSON string containing 1 to 128 Unicode scalar values. It MUST NOT contain U+0000, Unicode surrogate code points, C0 controls U+0000 through U+001F, C1 controls U+0080 through U+009F, leading `unicode_whitespace`, trailing `unicode_whitespace`, `/`, `\`, or `#`. It MUST be compared by exact Unicode code point sequence after validation. No case folding, trimming, locale comparison, or Unicode normalization is applied. |
| `generated_id` | An `identifier` with a required prefix from §7.1 followed by exactly 64 lowercase hexadecimal characters. |
| `timestamp` | A JSON string in proleptic Gregorian UTC form `YYYY-MM-DDTHH:MM:SSZ` or `YYYY-MM-DDTHH:MM:SS.ffffffZ`. Year range is `0001` through `9999`. Month is `01` through `12`. Day MUST be valid for the month and Gregorian leap-year rules. Hour is `00` through `23`. Minute is `00` through `59`. Second is `00` through `59`; leap seconds are invalid. Fractional seconds, when present, contain 1 to 6 ASCII decimal digits and preserve supplied precision after validation. Timestamps generated by §10.8 MUST use exactly 6 fractional digits. |
| `property_key` | An `identifier` used as an object member key for projected properties or mapped metadata. It MUST NOT contain `.`. It MUST NOT equal `kind`, `properties`, `metadata`, `source_metadata`, or `projected` when used as a field-path terminal. |
| `kind` | An `identifier` that names a source entity kind, source relationship kind, projected vertex kind, or projected edge kind. |
| `finite_integer` | A JSON number token matching exactly `0` or `-?[1-9][0-9]*`, with mathematical value in the inclusive range `-9007199254740991` through `9007199254740991`. Decimal-point notation, exponent notation, leading plus sign, leading zeroes, and `-0` are invalid even when the mathematical value would be integral. |
| `property_value` | One of JSON string, JSON boolean, `finite_integer`, JSON null, or an array of those scalar values. A JSON string value MUST contain no more than `max_string_property_value_length` Unicode scalar values. Arrays MUST contain at most 1024 values. Nested arrays and objects are invalid as property values. |
| `sha256_hex` | A JSON string containing exactly 64 lowercase hexadecimal characters. |
| `field_path` | A JSON string matching exactly one row in §4.2. It is parsed by splitting on literal `.` into exactly the declared path segments. No escaping is supported in this NLSpec revision. |
| `cursor_token` | A JSON string containing 1 to 4096 Unicode scalar values. The token is opaque to the caller and valid only under the query-specific cursor rules that emitted it. |
| `idempotency_key` | A JSON string containing 1 to 128 Unicode scalar values. It MUST NOT contain U+0000, Unicode surrogate code points, C0 controls, C1 controls, leading `unicode_whitespace`, or trailing `unicode_whitespace`. It is compared by exact Unicode code point sequence after JSON decoding. No case folding, trimming, locale comparison, or Unicode normalization is applied. |

For `identifier` and `idempotency_key` leading and trailing checks, `unicode_whitespace` means exactly these code points: U+0009 through U+000D, U+0020, U+0085, U+00A0, U+1680, U+2000 through U+200A, U+2028, U+2029, U+202F, U+205F, and U+3000.

Object member names in projection inputs MUST be exact code point matches for the member names in this NLSpec after applying the schema-table notation rule in §4.0. Unknown members are invalid unless a schema table explicitly allows them.

### 4.2 Field-path grammar

A field path MUST match one and only one row in the following grammar. A `property_key` containing `.` is invalid so that dotted field paths are unambiguous.

| Path family | Allowed form | Allowed target scopes | Terminal key contract | Value when present |
| --- | --- | --- | --- | --- |
| Source entity ID | `source_entity_id` | source entity filter, source entity aggregation | none | Source entity ID string. |
| Source relationship ID | `source_relationship_id` | source relationship filter, source relationship aggregation | none | Source relationship ID string. |
| Source relationship source endpoint | `src_source_entity_id` | source relationship filter, source relationship aggregation | none | Source entity ID string. |
| Source relationship destination endpoint | `dst_source_entity_id` | source relationship filter, source relationship aggregation | none | Source entity ID string. |
| Kind | `kind` | source entity, source relationship, projected vertex, projected edge | none | Source or projected kind. |
| Source relationship direction | `direction` | source relationship filter, source relationship aggregation | none | Source relationship direction token. |
| Source property | `properties.<property_key>` | source entity, source relationship, direct vertex property derivation, direct edge property derivation | `property_key` | Source property value. |
| Source metadata | `metadata.<property_key>` | source entity, source relationship, direct vertex metadata derivation, direct edge metadata derivation | `property_key` | Source metadata value. |
| Source metadata input | `source_metadata.<property_key>` | graph-view property derivation, graph-view metadata derivation | `property_key` | Top-level source metadata value. |
| Projected vertex ID | `projected.vertex_id` | projected vertex aggregation | none | Projected vertex ID. |
| Projected edge ID | `projected.edge_id` | projected edge aggregation | none | Projected edge ID. |
| Projected source endpoint | `projected.src_vertex_id` | projected edge aggregation | none | Projected edge source vertex ID. |
| Projected destination endpoint | `projected.dst_vertex_id` | projected edge aggregation | none | Projected edge destination vertex ID. |
| Projected direction | `projected.direction` | projected edge aggregation | none | Projected edge direction token. |
| Projected property | `projected.properties.<property_key>` | projected vertex aggregation, projected edge aggregation | `property_key` | Projected property value. |
| Projected metadata | `projected.metadata.<property_key>` | projected vertex aggregation, projected edge aggregation | `property_key` | System or mapped metadata value. |

A field path valid in one target scope is invalid in every other scope unless another row explicitly lists that scope. A malformed field path MUST produce `invalid_field_path` when the path is part of configuration and `invalid_filter` when the path is part of a filter predicate.

### 4.3 Top-level projection input object

The top-level projection input object MUST be a JSON object with exactly the following members. Unknown top-level members are invalid.

| Member | Type | Required | Nullable | Default when omitted | Validation rule |
| --- | --- | ---: | ---: | --- | --- |
| `projection_schema_id` | `identifier` | Yes | No | None | MUST equal exactly `graph_projection.v1`. Omission produces `missing_required_input`. Explicit `null` produces `explicit_null_not_allowed`. Any other value produces `invalid_projection_schema`. |
| `graph_view_id` | `generated_id` with prefix `gv_` | Yes | No | None | MUST equal the generated graph-view ID derived by §7.4. A syntactically invalid value or a syntactically valid but non-derived value produces `invalid_graph_view_id`. The implementation MUST NOT silently replace it. |
| `source_snapshot_id` | `identifier` | Yes | No | None | MUST identify one source snapshot. It MUST remain stable for the source snapshot it names. |
| `projection_config` | Object | Yes | No | None | MUST satisfy §4.5. |
| `source_entities[]` | Array of `source_entity` | Yes | No | None | MUST contain zero or more source entities. Duplicate valid `source_entity_id` values are fatal. |
| `source_relationships[]` | Array of `source_relationship` | Yes | No | None | MUST contain zero or more source relationships. Duplicate valid `source_relationship_id` values are fatal. |
| `source_metadata` | Object | No | No | `{}` | Keys MUST satisfy `property_key`. Values MUST satisfy `property_value`. |
| `filters` | Object | No | No | Include all source entities and relationships eligible under declared mappings. | MUST satisfy §4.11. Explicit `null` is invalid. |
| `relationship_definitions[]` | Array of `relationship_mapping_rule` | No | No | `[]` | Provides top-level relationship mapping rules. It MUST NOT be non-empty when `projection_config.relationship_mappings[]` is non-empty. |
| `property_definitions[]` | Array of `property_definition` | No | No | `[]` | Defines projected graph-view, vertex, and edge properties. |
| `requested_at` | `timestamp` | Yes | No | None | Records projection request time. It does not define `generated_at`. |
| `requested_by` | `identifier` | Yes | No | None | Identifies the requesting actor or system process. |

No top-level projection input member is nullable. Explicit JSON `null` for any top-level member is invalid. When the member is evaluated before admission, the operation MUST fail with lifecycle `invalid_projection_request` and `reason_code=explicit_null_not_allowed`; when the member is evaluated after admission, the run MUST emit `explicit_null_not_allowed`.

### 4.4 Identity participation matrix

The identity participation matrix is the only owner of input-family identity participation. Later sections define the exact tuples and digest algorithms.

| Input family | Affects `graph_view_id` | Affects `projection_run_id` | Affects `vertex_id` / `edge_id` | Affects canonical output bytes | Notes |
| --- | --- | --- | --- | --- | --- |
| `projection_schema_id` | Yes | Yes | Yes | Yes | It is included in graph-view, run, vertex, edge, and digest tuples. |
| `projection_config.graph_view_key` | Yes | Yes | Yes | Yes | It derives `graph_view_id`; graph-view ID then participates in object IDs. |
| `graph_view_id` | No independently | Yes | Yes | Yes | The caller-supplied value is accepted only after derivation validation. |
| `source_snapshot_id` | No | Yes | No | Yes | It identifies the source snapshot for a run. |
| `projection_config.projection_version` | No | Yes | No unless mapping identity fields change | Yes | It is copied to output and included in run metadata. |
| `projection_config.declared_source_entity_kinds[]` | No | Yes through config digest | Indirectly through validation and mapping eligibility | Yes | It controls mapping completeness and source eligibility. |
| `projection_config.declared_source_relationship_kinds[]` | No | Yes through config digest | Indirectly through validation and mapping eligibility | Yes | It controls relationship mapping completeness and source eligibility. |
| `entity_mappings[]` | No | Yes through config digest | Identity-participating fields affect direct vertex IDs | Yes | Identity-participating fields are defined in §7.5. |
| `relationship_mappings[]` and `relationship_definitions[]` | No | Yes through config digest | Identity-participating fields affect direct and reverse edge IDs | Yes | Only one relationship mapping source may be active. |
| `aggregation_rules[]` | No | Yes through config digest | Identity-participating fields affect aggregated IDs | Yes | Grouping, target kind, endpoint grouping, and direction fields affect aggregated IDs. |
| `filters` | No | Yes through config digest | No reserved IDs for excluded objects | Yes | Filters affect which objects are emitted. |
| `property_definitions[]` | No | Yes through config digest | No unless used by aggregation grouping | Yes | Property definitions affect output properties and validation. |
| `metadata_mappings[]` | No | Yes through config digest | No | Yes | Metadata mappings affect metadata output and validation. |
| `retention_policy` | No | Yes through config digest | No | No for graph object bytes; yes for lifecycle query behavior | Retention changes query availability but not graph object derivation. |
| `custom_config` | No | No | No | No | It MUST NOT affect observable projection behavior in this revision. |
| `source_entities[]` | No | Yes through source digest | Source IDs and source kinds affect direct vertex IDs; property values affect IDs only when used by aggregation grouping | Yes | Source order is normalized before derivation. |
| `source_relationships[]` | No | Yes through source digest | Source IDs, source kinds, endpoints, projected direction, and mapping identity affect direct edge IDs; property values affect IDs only when used by aggregation grouping | Yes | Source order is normalized before derivation. |
| `source_metadata` | No | Yes through source digest | No | Yes when mapped or copied to graph-view output | It never changes `graph_view_id`. |
| `requested_at` | No | No | No | No | It is request audit input and is not included in canonical graph output. |
| `requested_by` | No | No | No | No | It is request audit input and is not included in canonical graph output. |

The graph-view output metadata MUST include:

- `projection_config_digest`: lowercase SHA-256 computed by the exact `GPCONFIG1` digest envelope in §7.3.1.
- `projection_source_digest`: lowercase SHA-256 computed by the exact `GPSOURCE1` digest envelope in §7.3.1.

These digests are output metadata and verification aids. They do not replace `projection_run_id`.

### 4.5 Projection configuration object

`projection_config` MUST be a JSON object with exactly the following members. Unknown members are invalid.

| Member | Type | Required | Nullable | Default when omitted | Validation rule |
| --- | --- | ---: | ---: | --- | --- |
| `graph_view_key` | `identifier` | Yes | No | None | Stable caller-declared key used to derive `graph_view_id`. |
| `projection_version` | `identifier` | No | No | `1` | Copied to output `projection_version`. |
| `declared_source_entity_kinds[]` | Array of `kind` | Yes | No | None | MUST contain at least one kind unless `allow_empty_kind_registry=true`. Duplicate values are invalid. |
| `declared_source_relationship_kinds[]` | Array of `kind` | No | No | `[]` | Duplicate values are invalid. |
| `entity_mappings[]` | Array of `entity_mapping_rule` | Yes | No | None | MUST contain exactly one mapping rule for every declared source entity kind unless `allow_empty_kind_registry=true` and no entity kinds are declared. |
| `relationship_mappings[]` | Array of `relationship_mapping_rule` | No | No | `[]` | Active only when top-level `relationship_definitions[]` is empty. It MUST NOT be non-empty when top-level `relationship_definitions[]` is non-empty. |
| `metadata_mappings[]` | Array of `metadata_mapping_rule` | No | No | `[]` | Duplicate `(target_scope,target_kind,projected_metadata_key)` tuples are invalid. |
| `aggregation_rules[]` | Array of `aggregation_rule` | No | No | `[]` | Duplicate `aggregation_rule_id` values are invalid. |
| `default_vertex_labels[]` | Array of strings | No | No | `[]` | Labels MUST be non-empty strings with at most 256 Unicode scalar values. Duplicate labels are removed during normalization. |
| `default_edge_labels[]` | Array of strings | No | No | `[]` | Labels MUST be non-empty strings with at most 256 Unicode scalar values. Duplicate labels are removed during normalization. |
| `allow_empty_kind_registry` | Boolean | No | No | `false` | When `true`, `declared_source_entity_kinds[]` MAY be empty and projection emits an empty graph unless aggregation rules emit graph objects. |
| `retention_policy` | Object | No | No | Defaults in §10.6 | MUST satisfy §10.6. |
| `custom_config` | Object | No | No | `{}` | Keys MUST satisfy `property_key`; values MUST satisfy `property_value`. It MUST NOT affect projection output, validation, identity, ordering, lifecycle, traversal, retention, or query behavior in this NLSpec revision. |

A mapping rule, property definition, metadata mapping, filter, or aggregation rule that references `custom_config` is invalid with `invalid_projection_config`.

### 4.6 Mapping rule schemas

All mapping-rule objects are closed. Unknown members are invalid. Mapping rule arrays are normalized by sorting on the rule identifier after validation.

#### 4.6.1 Entity mapping rule

An `entity_mapping_rule` MUST be an object with exactly the following members.

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `mapping_rule_id` | `identifier` | Yes | No | None | Unique across all entity and relationship mapping rules. |
| `source_entity_kind` | `kind` | Yes | No | None | MUST appear in `declared_source_entity_kinds[]`. |
| `projected_vertex_kind` | `kind` | Yes | No | None | Emitted vertex kind. |
| `inclusion_predicate` | Filter predicate or string `always` | No | No | `always` | Applies after global entity filters. |
| `label_policy` | String | No | No | `mapping_only` | MUST be `mapping_only`, `preserve_source`, or `mapping_then_source`. |
| `mapping_labels[]` | Array of strings | No | No | `[]` | Labels MUST be non-empty strings with at most 256 Unicode scalar values. Duplicate labels are removed and values sort by code point order. |
| `required_property_keys[]` | Array of `property_key` | No | No | `[]` | Keys MUST resolve to applicable `property_definitions[]` after wildcard expansion. Duplicate keys are invalid. |
| `optional_property_keys[]` | Array of `property_key` | No | No | `[]` | Keys MUST resolve to applicable `property_definitions[]` after wildcard expansion. Duplicate keys are invalid. Values MUST NOT overlap `required_property_keys[]`. |

#### 4.6.2 Relationship mapping rule

A `relationship_mapping_rule` MUST be an object with exactly the following members. Top-level `relationship_definitions[]` and `projection_config.relationship_mappings[]` use this same schema.

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `mapping_rule_id` | `identifier` | Yes | No | None | Unique across all entity and relationship mapping rules. |
| `source_relationship_kind` | `kind` | Yes | No | None | MUST appear in `declared_source_relationship_kinds[]`. |
| `projected_edge_kind` | `kind` | Yes | No | None | Emitted edge kind. |
| `inclusion_predicate` | Filter predicate or string `always` | No | No | `always` | Applies after global relationship filters. |
| `direction_policy` | String | No | No | `preserve` | MUST be `preserve`, `normalize_forward`, `normalize_reverse`, `undirected`, or `bidirectional`. Closed by §6.4. |
| `emit_reverse_edge` | Boolean | No | No | `false` | Closed by §6.5. |
| `reverse_edge_kind` | `kind` | No | No | `projected_edge_kind` | Valid only when `emit_reverse_edge=true`. |
| `label_policy` | String | No | No | `mapping_only` | MUST be `mapping_only`, `preserve_source`, or `mapping_then_source`. |
| `mapping_labels[]` | Array of strings | No | No | `[]` | Labels MUST be non-empty strings with at most 256 Unicode scalar values. Duplicate labels are removed and values sort by code point order. |
| `required_property_keys[]` | Array of `property_key` | No | No | `[]` | Keys MUST resolve to applicable `property_definitions[]` after wildcard expansion. Duplicate keys are invalid. |
| `optional_property_keys[]` | Array of `property_key` | No | No | `[]` | Keys MUST resolve to applicable `property_definitions[]` after wildcard expansion. Duplicate keys are invalid. Values MUST NOT overlap `required_property_keys[]`. |

#### 4.6.3 Mapping property-key semantics

`required_property_keys[]` and `optional_property_keys[]` are references to projected property definitions. They MUST NOT read source properties directly. They MUST NOT synthesize passthrough properties. They MUST NOT create an additional property derivation mechanism.

For an entity mapping rule, every key in either array MUST match exactly one property definition whose expanded applicability is `(target_scope=vertex, concrete_target_kind=<projected_vertex_kind>, projected_key=<key>)`. For a relationship mapping rule, every key in either array MUST match exactly one property definition whose expanded applicability is `(target_scope=edge, concrete_target_kind=<projected_edge_kind>, projected_key=<key>)`.

| Mapping key condition | Required validation result |
| --- | --- |
| Key appears in both required and optional arrays | `invalid_mapping_rule` with `reason_code=required_optional_overlap`. |
| Key resolves to no applicable property definition | `invalid_mapping_rule` with `reason_code=property_key_not_defined`. |
| Key resolves to more than one definition after wildcard expansion | `invalid_property_definition` with `reason_code=duplicate_after_wildcard_expansion`. |
| Key in `required_property_keys[]` references a definition with `required=false` | `invalid_mapping_rule` with `reason_code=property_requiredness_mismatch`. |
| Key in `optional_property_keys[]` references a definition with `required=true` | `invalid_mapping_rule` with `reason_code=property_requiredness_mismatch`. |

#### 4.6.4 Metadata mapping rule

A `metadata_mapping_rule` MUST be an object with exactly the following members.

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `metadata_mapping_id` | `identifier` | Yes | No | None | Unique within `metadata_mappings[]`. |
| `target_scope` | String | Yes | No | None | MUST be `graph_view`, `vertex`, or `edge`. |
| `target_kind` | `kind` or `*` | Yes | No | None | `*` applies to all concrete kinds in scope after §4.10.2 expansion. For `graph_view`, `target_kind` MUST equal `*`. |
| `source_field_path` | `field_path` | Yes | No | None | MUST be valid for the target scope and derivation context. |
| `projected_metadata_key` | `property_key` | Yes | No | None | Unique after §4.10.2 wildcard expansion for `(target_scope, concrete_target_kind)`. |
| `projected_type` | String | Yes | No | None | MUST use the same closed type vocabulary as `property_definition.projected_type`. |
| `required` | Boolean | No | No | `false` | Missing or invalid source behavior follows `missing_behavior` and `source_null_behavior`. |
| `default_value` | `property_value` | No | Yes | No default | Required when `missing_behavior=default` or `source_null_behavior=default`. Must be compatible with `projected_type`. |
| `missing_behavior` | String | No | No | `error` when `required=true`, otherwise `omit` | MUST be `omit`, `default`, or `error`. |
| `source_null_behavior` | String | No | No | `error` when `required=true`, otherwise `omit` | MUST be `omit`, `default`, `emit_null`, or `error`. |
| `null_output_policy` | String | No | No | `omit` | MUST be `omit` or `emit_null`. |
| `merge_behavior` | String | No | No | `single_value` | Used only for aggregation. MUST be one of the merge behaviors in §6.8.6. |

A metadata mapping that would emit a metadata key reserved by this NLSpec is invalid. Reserved metadata keys are the required system metadata members in §5.5.

#### 4.6.5 Aggregation rule

An `aggregation_rule` MUST be an object with exactly the following members.

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `aggregation_rule_id` | `identifier` | Yes | No | None | Unique within `aggregation_rules[]`. |
| `target_scope` | String | Yes | No | None | MUST be `vertex` or `edge`. |
| `input_scope` | String | Yes | No | None | MUST be `source_entity`, `source_relationship`, `projected_vertex`, or `projected_edge`. |
| `input_kind` | `kind` or `*` | Yes | No | None | Selects candidates after global filtering and direct mapping. |
| `projected_kind` | `kind` | Yes | No | None | Emitted vertex or edge kind. |
| `grouping_keys[]` | Array of `field_path` | Yes | No | None | MUST contain 1 to 32 distinct paths valid for `input_scope`. |
| `missing_grouping_key_behavior` | String | No | No | `error` | MUST be `error`, `exclude`, or `use_null`. |
| `source_reference_policy` | String | No | No | `preserve_all` | Only `preserve_all` is defined in this revision. |
| `property_merge_behavior` | Object | No | No | `{}` | Keys are `property_key`; values are merge behavior tokens from §6.8.6. Unknown behavior tokens are invalid. |
| `edge_direction` | String | Conditional | No | None | Required when `target_scope=edge`; MUST be `directed`, `undirected`, or `bidirectional`. Forbidden when `target_scope=vertex`. |
| `endpoint_grouping` | Object | Conditional | No | None | Required when `target_scope=edge`; forbidden when `target_scope=vertex`. Must satisfy §4.6.6. |

#### 4.6.6 Aggregated edge endpoint grouping object

An `endpoint_grouping` object MUST have exactly the following members.

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `src_vertex_aggregation_rule_id` | `identifier` | Yes | No | None | MUST reference a vertex aggregation rule that executes before the edge aggregation rule. |
| `src_grouping_keys[]` | Array of `field_path` | Yes | No | None | MUST contain paths valid for the edge aggregation rule's `input_scope`. Its length MUST equal the referenced source vertex aggregation rule's `grouping_keys[]` length. |
| `dst_vertex_aggregation_rule_id` | `identifier` | Yes | No | None | MUST reference a vertex aggregation rule that executes before the edge aggregation rule. |
| `dst_grouping_keys[]` | Array of `field_path` | Yes | No | None | MUST contain paths valid for the edge aggregation rule's `input_scope`. Its length MUST equal the referenced destination vertex aggregation rule's `grouping_keys[]` length. |
| `missing_endpoint_behavior` | String | No | No | `error` | MUST be `error` or `exclude`. |

A referenced aggregation rule that is absent, is not a vertex aggregation rule, executes after the edge aggregation rule, or has a mismatched grouping-key count MUST produce `invalid_aggregation_rule`. Endpoint digest derivation is defined by §6.8.7.

### 4.7 Source entity object

Each item in `source_entities[]` MUST be a JSON object with exactly the following members. Unknown members are invalid.

| Member | Type | Required | Nullable | Default when omitted | Validation rule |
| --- | --- | ---: | ---: | --- | --- |
| `source_entity_id` | `identifier` | Yes | No | None | Required for source-item identity. Duplicate valid IDs are fatal. |
| `source_entity_kind` | `kind` | Yes | No | None | MUST be declared in `projection_config.declared_source_entity_kinds[]` to be eligible for projection. Undeclared kinds produce `undeclared_source_kind` and exclude the item. |
| `properties` | Object | No | No | `{}` | Keys MUST satisfy `property_key`. Values MUST satisfy `property_value`. |
| `metadata` | Object | No | No | `{}` | Keys MUST satisfy `property_key`. Values MUST satisfy `property_value`. |
| `labels[]` | Array of strings | No | No | `[]` | Labels are source hints only. They are emitted only when a mapping rule declares label preservation. |

### 4.8 Source relationship object

Each item in `source_relationships[]` MUST be a JSON object with exactly the following members. Unknown members are invalid.

| Member | Type | Required | Nullable | Default when omitted | Validation rule |
| --- | --- | ---: | ---: | --- | --- |
| `source_relationship_id` | `identifier` | Yes | No | None | Required for source-item identity. Duplicate valid IDs are fatal. |
| `source_relationship_kind` | `kind` | Yes | No | None | MUST be declared in `projection_config.declared_source_relationship_kinds[]` to be eligible for projection. Undeclared kinds produce `undeclared_source_kind` and exclude the item. |
| `src_source_entity_id` | `identifier` | No | No | No default | Required for relationship eligibility. Omission produces `missing_relationship_endpoint` and excludes this source relationship. |
| `dst_source_entity_id` | `identifier` | No | No | No default | Required for relationship eligibility. Omission produces `missing_relationship_endpoint` and excludes this source relationship. |
| `direction` | String | No | No | `forward` | Omitted value defaults to `forward`. Invalid explicit value produces `invalid_relationship_direction` and excludes this source relationship. |
| `properties` | Object | No | No | `{}` | Keys MUST satisfy `property_key`. Values MUST satisfy `property_value`. |
| `metadata` | Object | No | No | `{}` | Keys MUST satisfy `property_key`. Values MUST satisfy `property_value`. |
| `labels[]` | Array of strings | No | No | `[]` | Labels are source hints only. They are emitted only when a mapping rule declares label preservation. |

A relationship endpoint that references a missing, excluded, or unmapped source entity produces `relationship_endpoint_not_projected` and excludes that source relationship. This exclusion does not reject unrelated source entities, source relationships, or the whole projection.

### 4.9 Source item semantic validation layers

Validation has two layers for source items.

| Layer | Examples | Required result |
| --- | --- | --- |
| Input-shape validation | Top-level object is not an object, unknown top-level members, missing top-level required members, duplicate source IDs, malformed scalar type for source item ID | Fatal validation issue; no consumable graph is emitted. |
| Source-item semantic validation | Missing relationship endpoint, endpoint references missing or filtered entity, invalid relationship direction, undeclared source kind | Nonfatal issue unless the code registry says otherwise; affected source item is excluded. |

A conforming implementation MUST NOT reject the whole projection merely because one source relationship has an item-level endpoint or direction issue. A conforming implementation MUST reject the whole projection when duplicate source IDs prevent deterministic identity.

### 4.10 Property definition object

A `property_definition` declares how one source or projected field becomes one projected property. Each item in `property_definitions[]` MUST be a JSON object with exactly the following members. Unknown members are invalid.

| Member | Type | Required | Nullable | Default when omitted | Validation rule |
| --- | --- | ---: | ---: | --- | --- |
| `property_definition_id` | `identifier` | Yes | No | None | Unique within `property_definitions[]`. |
| `target_scope` | String | Yes | No | None | MUST be `vertex`, `edge`, or `graph_view`. |
| `target_kind` | `kind` or `*` | Yes | No | None | `*` applies to all projected kinds in the target scope. For `graph_view`, `target_kind` MUST equal `*`. |
| `source_field_path` | `field_path` | Yes | No | None | MUST be valid for `target_scope` and derivation context. |
| `projected_key` | `property_key` | Yes | No | None | Unique after §4.10.2 wildcard expansion for `(target_scope, concrete_target_kind)`. |
| `projected_type` | String | Yes | No | None | MUST be one of `string`, `integer`, `boolean`, `timestamp`, `identifier`, `string_array`, or `identifier_array`. |
| `required` | Boolean | No | No | `false` | When `true`, missing, null, and invalid values are errors unless default or null behavior explicitly handles them. |
| `default_value` | `property_value` | No | Yes | No default | If supplied, MUST be compatible with `projected_type`. Explicit `null` is valid only when `null_output_policy=emit_null`. |
| `missing_behavior` | String | No | No | `error` when `required=true`, otherwise `omit` | MUST be `omit`, `default`, or `error`. `default` requires `default_value`. |
| `source_null_behavior` | String | No | No | `error` when `required=true`, otherwise `omit` | MUST be `omit`, `default`, `emit_null`, or `error`. `default` requires `default_value`. `emit_null` requires `null_output_policy=emit_null`. |
| `null_output_policy` | String | No | No | `omit` | MUST be `omit` or `emit_null`. |
| `merge_behavior` | String | No | No | `single_value` | Applies to aggregation. MUST be one of `single_value`, `first_by_sort`, `last_by_sort`, `distinct_sorted_array`, `count`, or `omit`. |

#### 4.10.1 Projected property compatibility

No JSON type coercion is allowed. Explicit source JSON null is handled only by `source_null_behavior`.

| `projected_type` | Accepted non-null source value | Accepted default value | Normalization | Invalid behavior |
| --- | --- | --- | --- | --- |
| `string` | JSON string | JSON string | Emit exact string. | `invalid_property_type` |
| `integer` | `finite_integer` | `finite_integer` | Emit exact integer. | `invalid_property_type` |
| `boolean` | JSON boolean | JSON boolean | Emit exact boolean. | `invalid_property_type` |
| `timestamp` | JSON string satisfying `timestamp` | Same | Emit validated timestamp string exactly as supplied. | `invalid_property_type` |
| `identifier` | JSON string satisfying `identifier` | Same | Emit exact identifier. | `invalid_property_type` |
| `string_array` | Array of JSON strings, max 1024 | Same | Preserve declared source order unless aggregation merge changes it. | `invalid_property_type` |
| `identifier_array` | Array of strings satisfying `identifier`, max 1024 | Same | Preserve declared source order unless aggregation merge changes it. | `invalid_property_type` |

For `target_scope=graph_view`, `source_field_path` MAY use only `source_metadata.<property_key>` unless an aggregation rule explicitly derives a graph-view property in a later NLSpec revision. This revision does not define graph-view property aggregation over source entities or relationships. For direct vertex properties, `source_field_path` MAY read only the mapped source entity. For direct edge properties, `source_field_path` MAY read only the mapped source relationship. For aggregated properties, `source_field_path` MAY read only fields available on the aggregation rule's declared input scope.

#### 4.10.2 Wildcard expansion and collision rules

Wildcard expansion is part of configuration validation and MUST run after mapping and aggregation rules are validated enough to determine concrete projected kinds.

The implementation MUST compute the concrete projected vertex kind set from all valid entity mappings and vertex aggregation rules. It MUST compute the concrete projected edge kind set from all valid relationship mappings and edge aggregation rules. For property definitions and metadata mappings, `target_kind="*"` expands to every concrete kind in the same `target_scope`. For `target_scope=graph_view`, `target_kind` remains exactly `*` and does not expand.

After expansion, property definitions MUST NOT contain duplicate `(target_scope, concrete_target_kind, projected_key)` tuples. After expansion, metadata mappings MUST NOT contain duplicate `(target_scope, concrete_target_kind, projected_metadata_key)` tuples. Concrete definitions MUST NOT override wildcard definitions. A wildcard definition and a concrete definition that emit the same key for the same concrete target are invalid.

| Collision family | Validation result |
| --- | --- |
| Property duplicate after wildcard expansion | `invalid_property_definition` with `reason_code=duplicate_after_wildcard_expansion`. |
| Metadata duplicate after wildcard expansion | `invalid_metadata_mapping` with `reason_code=duplicate_after_wildcard_expansion`. |

Schema-registry output MUST expose the concrete expanded applicability for property and metadata keys. It MUST NOT leave consumers to repeat wildcard expansion privately.

### 4.11 Filter object

`filters` MUST be an object with exactly the following members. Unknown members are invalid.

| Member | Type | Required | Nullable | Default when omitted | Validation rule |
| --- | --- | ---: | ---: | --- | --- |
| `entity_filters[]` | Array of filter predicates | No | No | `[]` | Predicates apply to source entities before vertex mapping. |
| `relationship_filters[]` | Array of filter predicates | No | No | `[]` | Predicates apply to source relationships before edge mapping. |
| `logic` | String | No | No | `and` | MUST be `and`. No other logical operator is defined in this NLSpec. |

A filter predicate MUST be an object with exactly the following members. Unknown members are invalid.

| Member | Type | Required | Nullable | Default when omitted | Validation rule |
| --- | --- | ---: | ---: | --- | --- |
| `field_path` | `field_path` | Yes | No | None | MUST be valid for the filtered item scope. |
| `op` | String | Yes | No | None | MUST be `exists`, `equals`, `not_equals`, or `in`. |
| `value` | `property_value` or array of scalar values | Conditional | Yes | None | Required for `equals`, `not_equals`, and `in`. Forbidden for `exists`. For `in`, MUST be a non-empty array with at most 1024 scalar values or JSON null values. Nested arrays are invalid for `in`. |
| `include_if_missing` | Boolean | No | No | `false` | Applies only when the field path is missing. It does not apply to explicit JSON null. |

Explicit JSON `null` in `value` means the predicate compares against explicit source JSON null. Omitted `value` means no value was supplied. These states are distinct.

#### 4.11.1 Filter evaluation semantics

Filter evaluation MUST be type-strict. String comparison uses exact Unicode code point comparison. Integer comparison uses exact integer value. Boolean and null comparison use JSON type identity. No coercion is allowed.

| Field state | `exists` | `equals(value)` | `not_equals(value)` | `in(values[])` |
| --- | --- | --- | --- | --- |
| Missing | `include_if_missing` | `include_if_missing` | `include_if_missing` | `include_if_missing` |
| Present JSON null | `true` | `true` iff `value` is JSON null | `true` iff `value` is not JSON null | `true` iff `values[]` contains JSON null |
| Present scalar | `true` | `true` iff same JSON type and same value | Inverse of `equals` | `true` iff any candidate has same JSON type and same value |
| Present array | `true` | `true` iff `value` is an array with same length, same order, and same element values | Inverse of `equals` | `true` iff any array element equals any candidate scalar or JSON null |

The `exists` operator forbids `value`. The `equals` and `not_equals` operators require `value`. The `in` operator requires `value` to be a non-empty array of scalar JSON values or JSON null. A valid predicate that evaluates false excludes the item without a validation issue. An invalid predicate is fatal `invalid_filter`.

### 4.12 Resource-limit registry

A conforming implementation MUST enforce the following closed resource limits before projection emits output. Implementations MUST NOT lower or raise these public limits in a way that changes conformance behavior.

| Limit key | Applies to | Required maximum | Overflow behavior |
| --- | --- | ---: | --- |
| `max_input_bytes` | UTF-8 JSON projection input before decoding | `268435456` bytes | `resource_limit_exceeded` fatal |
| `max_source_entities` | `source_entities[]` | `100000` | `resource_limit_exceeded` fatal |
| `max_source_relationships` | `source_relationships[]` | `250000` | `resource_limit_exceeded` fatal |
| `max_entity_mappings` | `entity_mappings[]` | `10000` | `resource_limit_exceeded` fatal |
| `max_relationship_mappings` | active relationship mappings from top-level or config | `10000` | `resource_limit_exceeded` fatal |
| `max_property_definitions` | `property_definitions[]` | `10000` | `resource_limit_exceeded` fatal |
| `max_metadata_mappings` | `metadata_mappings[]` | `10000` | `resource_limit_exceeded` fatal |
| `max_aggregation_rules` | `aggregation_rules[]` | `1000` | `resource_limit_exceeded` fatal |
| `max_labels_per_source_item` | `labels[]` on one source entity or source relationship | `256` | `source_item_resource_limit_exceeded`; affected item excluded |
| `max_label_length` | each label string | `256` Unicode scalar values | `invalid_input_shape` for malformed label |
| `max_string_property_value_length` | each JSON string inside `property_value` | `16384` Unicode scalar values | `invalid_input_shape` for configuration or top-level values; `source_item_resource_limit_exceeded` for source item values |
| `max_metadata_keys_per_object` | `metadata`, `source_metadata`, `mapped_metadata` objects | `1024` | `resource_limit_exceeded` fatal for top-level or output metadata; `source_item_resource_limit_exceeded` for source item metadata |
| `max_properties_per_object` | `properties` objects | `1024` | `source_item_resource_limit_exceeded` for source item properties; `projected_output_limit_exceeded` for emitted object properties |
| `max_custom_config_keys` | `custom_config` | `256` | `resource_limit_exceeded` fatal |
| `max_validation_issues` | emitted validation issues | `100000` | Closed by §9.3: select the first `N-1` non-cap issues by discovery order, sort selected non-cap issues by ordinary issue ordering, then append final `validation_issue_limit_exceeded`. |
| `max_validation_message_length` | validation issue `message` | `1024` Unicode scalar values | Truncate deterministically by preserving the first 1024 Unicode scalar values; issue identity is unaffected. |
| `max_failure_reason_length` | `failure_reason` | `4096` Unicode scalar values | Truncate deterministically by preserving the first 4096 Unicode scalar values; emit `projection_computation_failed` when the failure reason came from computation failure. |
| `max_query_error_message_length` | query error message if a transport exposes one | `1024` Unicode scalar values | Truncate deterministically by preserving the first 1024 Unicode scalar values; query error identity is unaffected. |
| `max_cursor_token_length` | supplied `cursor_token` | `4096` Unicode scalar values | `cursor_invalid` with `reason_code=cursor_token_too_long` |
| `max_projected_vertices` | total emitted vertices in one result | `500000` | `projected_output_limit_exceeded` fatal |
| `max_projected_edges` | total emitted edges in one result | `1000000` | `projected_output_limit_exceeded` fatal |
| `max_traversal_seed_vertices` | `traverse.seed_vertex_ids[]` | `1024` | `invalid_argument` |
| `max_traversal_kind_filters` | `edge_kinds[]` and `vertex_kinds[]` | `1024` each | `invalid_argument` |
| `max_traversal_depth` | `traverse.max_depth` | `16` | `invalid_argument` |
| `max_list_graph_views_limit` | `list_graph_views.limit` | `1000` | `invalid_argument` |

Whole-input limits that prevent request admission MUST use lifecycle `invalid_projection_request` according to §4.0.1 and MUST NOT emit validation issues. Whole-input limits discovered after admission MUST use `resource_limit_exceeded` and prevent a consumable graph. Item-scoped limits MUST use `source_item_resource_limit_exceeded` and exclude only the affected item. Derived-output limits MUST use `projected_output_limit_exceeded` and prevent a consumable graph.

### 4.13 Input normalization

After validation and before derivation, the implementation MUST normalize inputs as follows.

| Input family | Normalization rule |
| --- | --- |
| Object members | Preserve closed schema member order for canonical serialization. Dynamic object keys sort by Unicode code point order. |
| Declared kinds | Sort by exact code point order after duplicate detection. |
| Mapping rules | Sort by `mapping_rule_id`, `metadata_mapping_id`, or `aggregation_rule_id` as applicable after duplicate detection. |
| Property definitions | Sort by `property_definition_id` after duplicate detection. |
| Source entities | Sort by `source_entity_id` after duplicate detection. |
| Source relationships | Sort by `source_relationship_id` after duplicate detection. |
| Labels | Remove exact duplicates and sort by exact code point order unless source order is explicitly part of source reference output. |
| Filters | Preserve array order; filter order is observable only through first validation issue ordering, not through inclusion result because `logic` is `and`. |
| `custom_config` | Validate shape, sort keys for canonical input representation, and exclude from config digest and projection behavior. |
| Omitted optional objects | Materialize their default value before digest computation unless this NLSpec says the member remains omitted. |
| Omitted optional arrays | Materialize `[]` before digest computation. |

## 5. Projection output contract

A projection result MUST expose the output object family defined in this section when the selected projection run is consumable.

### 5.1 Graph view output object

A graph view output object MUST be a JSON object with exactly the following members and member order for canonical serialization.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `projection_schema_id` | String | Yes | No | Always `graph_projection.v1`. |
| `graph_view_id` | `generated_id` with prefix `gv_` | Yes | No | The validated graph view ID. |
| `graph_view_key` | `identifier` | Yes | No | Copied from `projection_config.graph_view_key`. |
| `projection_run_id` | `generated_id` with prefix `gpr_` | Yes | No | The run that produced this result. |
| `source_snapshot_id` | `identifier` | Yes | No | Copied from input. |
| `projection_version` | `identifier` | Yes | No | Copied from normalized projection config. |
| `generated_at` | `timestamp` | Yes | No | Timestamp when the result became available. It is run-specific. |
| `state` | String | Yes | No | Projection-run/result state for this result. Consumable reads return only `available`, `replaced`, or `invalidated` when a query explicitly allows invalidated inspection. |
| `properties` | Object | Yes | No | Projected graph-view properties sorted by key. Empty object when none. |
| `metadata` | Graph-view metadata object | Yes | No | Must satisfy §5.5.1. |
| `schema_registry` | Object | Yes | No | Must satisfy §5.2. |
| `vertices[]` | Array of projected vertex objects | Yes | No | Sorted by §5.6. |
| `edges[]` | Array of projected edge objects | Yes | No | Sorted by §5.6. |
| `validation_summary` | Validation summary object | Yes | No | Must satisfy §9.4. |
| `consumer_capabilities` | Consumer capabilities object | Yes | No | Must satisfy §5.7. |

### 5.2 Schema registry object

The graph view output MUST include `schema_registry` so consumers can rely on declared vertex and edge schemas, labels, property keys, and metadata keys without private configuration access. The registry exposes concrete expanded output applicability. Consumers MUST NOT repeat wildcard expansion privately.

`schema_registry` MUST be a JSON object with exactly the following members in this order.

| Member | Type | Required | Nullable | Empty allowed | Sorting rule |
| --- | --- | ---: | ---: | ---: | --- |
| `vertex_kinds[]` | Array of `vertex_kind_schema_item` | Yes | No | Yes | Sorted by `vertex_kind`. |
| `edge_kinds[]` | Array of `edge_kind_schema_item` | Yes | No | Yes | Sorted by `edge_kind`. |
| `property_keys[]` | Array of `property_schema_item` | Yes | No | Yes | Sorted by `target_scope`, `target_kind`, `projected_key`. |
| `metadata_keys[]` | Array of `metadata_schema_item` | Yes | No | Yes | Sorted by `target_scope`, `target_kind`, `projected_metadata_key`. |

A `vertex_kind_schema_item` MUST be a JSON object with exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `vertex_kind` | `kind` | Yes | No | Projected vertex kind. |
| `source_entity_kinds[]` | Array of `kind` | Yes | No | Source entity kinds that can emit this vertex kind, sorted by exact code point order. Empty when only aggregation emits the kind. |
| `aggregation_rule_ids[]` | Array of `identifier` | Yes | No | Vertex aggregation rules that can emit this vertex kind, sorted by exact code point order. Empty when only direct mappings emit the kind. |
| `labels[]` | Array of strings | Yes | No | Concrete static labels that can be emitted for this vertex kind from default labels and mapping labels, sorted by exact code point order. Dynamic source labels are not enumerated. |
| `source_labels_preserved` | Boolean | Yes | No | `true` when any entity mapping for this kind can preserve source item `labels[]`; otherwise `false`. |
| `properties[]` | Array of `property_schema_reference` | Yes | No | Property schema references for this kind, sorted by `projected_key`. Empty when no property applies. |

An `edge_kind_schema_item` MUST be a JSON object with exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `edge_kind` | `kind` | Yes | No | Projected edge kind. |
| `source_relationship_kinds[]` | Array of `kind` | Yes | No | Source relationship kinds that can emit this edge kind, sorted by exact code point order. Empty when only aggregation emits the kind. |
| `aggregation_rule_ids[]` | Array of `identifier` | Yes | No | Edge aggregation rules that can emit this edge kind, sorted by exact code point order. Empty when only direct mappings emit the kind. |
| `directions[]` | Array of strings | Yes | No | Projected edge directions that may be emitted for this kind, sorted in the fixed order `directed`, `undirected`, `bidirectional`. |
| `labels[]` | Array of strings | Yes | No | Concrete static labels that can be emitted for this edge kind from default labels and mapping labels, sorted by exact code point order. Dynamic source labels are not enumerated. |
| `source_labels_preserved` | Boolean | Yes | No | `true` when any relationship mapping for this kind can preserve source item `labels[]`; otherwise `false`. |
| `properties[]` | Array of `property_schema_reference` | Yes | No | Property schema references for this kind, sorted by `projected_key`. Empty when no property applies. |

A `property_schema_reference` MUST be a JSON object with exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `projected_key` | `property_key` | Yes | No | Property key. |
| `projected_type` | String | Yes | No | Closed projected type from §4.10. |
| `required` | Boolean | Yes | No | Effective requiredness for this concrete kind. |
| `nullable_output` | Boolean | Yes | No | `true` iff `null_output_policy=emit_null`. |

A `property_schema_item` MUST be a JSON object with exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `target_scope` | String | Yes | No | `graph_view`, `vertex`, or `edge`. |
| `target_kind` | `kind` or `*` | Yes | No | Concrete kind after wildcard expansion for `vertex` and `edge`; exactly `*` only when `target_scope=graph_view`. |
| `projected_key` | `property_key` | Yes | No | Property key. |
| `projected_type` | String | Yes | No | Closed projected type from §4.10. |
| `required` | Boolean | Yes | No | Effective requiredness after defaults. |
| `nullable_output` | Boolean | Yes | No | `true` iff `null_output_policy=emit_null`. |
| `missing_behavior` | String | Yes | No | Effective missing behavior after defaults. |
| `source_null_behavior` | String | Yes | No | Effective null behavior after defaults. |

A `metadata_schema_item` MUST be a JSON object with exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `target_scope` | String | Yes | No | `graph_view`, `vertex`, or `edge`. |
| `target_kind` | `kind` or `*` | Yes | No | Concrete kind after wildcard expansion for `vertex` and `edge`; exactly `*` only when `target_scope=graph_view`. |
| `projected_metadata_key` | `property_key` | Yes | No | Mapped metadata key. |
| `projected_type` | String | Yes | No | Closed projected type from §4.10. |
| `required` | Boolean | Yes | No | Effective requiredness after defaults. |
| `nullable_output` | Boolean | Yes | No | `true` iff `null_output_policy=emit_null`. |
| `missing_behavior` | String | Yes | No | Effective missing behavior after defaults. |
| `source_null_behavior` | String | Yes | No | Effective null behavior after defaults. |

Registry `property_keys[]` and `metadata_keys[]` MUST emit concrete expanded applicability. `target_kind="*"` MUST NOT appear in registry items except for `target_scope=graph_view`, where `target_kind` remains exactly `*`.

Static label reporting MUST follow this matrix.

| Label source | Direct output object | Aggregated output object | Schema registry `labels[]` | Schema registry dynamic indicator |
| --- | --- | --- | --- | --- |
| `default_vertex_labels[]` | Included on every vertex | Included on aggregated vertices | Included for every vertex kind | No dynamic indicator. |
| `default_edge_labels[]` | Included on every edge | Included on aggregated edges | Included for every edge kind | No dynamic indicator. |
| Entity mapping `mapping_labels[]` | Included according to entity `label_policy` | Not applicable | Included for mapped vertex kind when any mapping can emit it | No dynamic indicator. |
| Relationship mapping `mapping_labels[]` | Included according to relationship `label_policy` | Not applicable | Included for mapped edge kind when any mapping can emit it | No dynamic indicator. |
| Source entity `labels[]` | Included only when the entity label policy preserves source labels | Not applicable | Not enumerated | `source_labels_preserved=true`. |
| Source relationship `labels[]` | Included only when the relationship label policy preserves source labels | Not applicable | Not enumerated | `source_labels_preserved=true`. |
| Aggregation rule labels | Not applicable | No labels in this revision | Not applicable | No dynamic indicator. |

### 5.3 Projected vertex object

A projected vertex object MUST be a JSON object with exactly the following members and member order for canonical serialization.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `vertex_id` | `generated_id` with prefix `vx_` | Yes | No | Derived by §7.6. |
| `vertex_kind` | `kind` | Yes | No | Projected vertex kind. |
| `vertex_family` | String | Yes | No | `direct` or `aggregated`. |
| `labels[]` | Array of strings | Yes | No | Normalized labels sorted by exact code point order. |
| `properties` | Object | Yes | No | Projected properties sorted by key. Empty object when none. |
| `metadata` | Vertex metadata object | Yes | No | Must satisfy §5.5.2. |
| `source_entity_ref` | Object | Yes | Yes | Required and non-null for direct vertices; `null` for aggregated vertices. |
| `sort_key` | String | Yes | No | Derived by §5.6.1. |

A direct vertex `source_entity_ref` MUST be an object with exactly these members: `source_entity_id`, `source_entity_kind`, and `mapping_rule_id`. An aggregated vertex MUST set `source_entity_ref=null` and MUST preserve source traceability through `metadata.aggregation_source_refs[]`.

### 5.4 Projected edge object

A projected edge object MUST be a JSON object with exactly the following members and member order for canonical serialization.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `edge_id` | `generated_id` with prefix `ed_` | Yes | No | Derived by §7.6. |
| `edge_kind` | `kind` | Yes | No | Projected edge kind. |
| `edge_family` | String | Yes | No | `direct`, `reverse`, or `aggregated`. |
| `src_vertex_id` | `generated_id` with prefix `vx_` | Yes | No | Source endpoint in projected graph. |
| `dst_vertex_id` | `generated_id` with prefix `vx_` | Yes | No | Destination endpoint in projected graph. |
| `direction` | String | Yes | No | `directed`, `undirected`, or `bidirectional`. |
| `labels[]` | Array of strings | Yes | No | Normalized labels sorted by exact code point order. |
| `properties` | Object | Yes | No | Projected properties sorted by key. Empty object when none. |
| `metadata` | Edge metadata object | Yes | No | Must satisfy §5.5.3. |
| `source_relationship_ref` | Object | Yes | Yes | Required and non-null for direct and reverse edges; `null` for aggregated edges. |
| `sort_key` | String | Yes | No | Derived by §5.6.1. |

A direct or reverse edge `source_relationship_ref` MUST be an object with exactly these members: `source_relationship_id`, `source_relationship_kind`, and `mapping_rule_id`. An aggregated edge MUST set `source_relationship_ref=null` and MUST preserve source traceability through `metadata.aggregation_source_refs[]`.

### 5.5 Metadata object schemas

Implementation-private metadata is forbidden in canonical output. All output metadata MUST be either one of the schema-owned system metadata members below or a key under `mapped_metadata` emitted by `metadata_mappings[]`.

#### 5.5.1 Graph-view metadata

Graph-view metadata MUST be a JSON object with exactly the following members.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `previous_projection_run_id` | `generated_id` with prefix `gpr_` | Yes | Yes | `null` on initial creation; prior available run ID on replacement. |
| `projection_config_digest` | `sha256_hex` | Yes | No | Digest from §7.3.1. |
| `projection_source_digest` | `sha256_hex` | Yes | No | Digest from §7.3.1. |
| `mapped_metadata` | Object | Yes | No | Contains only metadata emitted by `metadata_mappings[]`; keys sorted by code point order. |
| `invalidation` | Object | Yes | Yes | `null` unless the selected run is invalidated. When non-null, it MUST satisfy the table below. |

A non-null `invalidation` object MUST have exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `invalidated_at` | `timestamp` | Yes | No | Timestamp when invalidation was accepted. |
| `reason_code` | String | Yes | No | One code from §10.7. |
| `requested_by` | `identifier` | Yes | No | Actor or system process that requested invalidation. |
| `target_scope` | String | Yes | No | `graph_view` or `projection_run`. |
| `target_projection_run_id` | `generated_id` with prefix `gpr_` | Yes | Yes | Non-null for run-specific invalidation; `null` for graph-view cascade. |

#### 5.5.2 Vertex metadata

Vertex metadata MUST be a JSON object with exactly the following members.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `mapping_rule_id` | `identifier` | Yes | Yes | Present for direct vertices; `null` for aggregation-only vertices. |
| `aggregation_rule_id` | `identifier` | Yes | Yes | Present for aggregated vertices; `null` for direct vertices. |
| `aggregation_source_refs[]` | Array | Yes | No | Empty for direct vertices; source refs for aggregated vertices. |
| `mapped_metadata` | Object | Yes | No | Contains only metadata emitted by `metadata_mappings[]`; keys sorted by code point order. |

#### 5.5.3 Edge metadata

Edge metadata MUST be a JSON object with exactly the following members.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `mapping_rule_id` | `identifier` | Yes | Yes | Present for direct and reverse edges; `null` for aggregation-only edges. |
| `aggregation_rule_id` | `identifier` | Yes | Yes | Present for aggregated edges; `null` for direct and reverse edges. |
| `aggregation_source_refs[]` | Array | Yes | No | Empty for direct and reverse edges; source refs for aggregated edges. |
| `is_reverse_edge` | Boolean | Yes | No | `true` only for emitted reverse edges. |
| `reverse_of_edge_id` | `generated_id` with prefix `ed_` | Yes | Yes | Primary edge ID for reverse edges; `null` otherwise. |
| `mapped_metadata` | Object | Yes | No | Contains only metadata emitted by `metadata_mappings[]`; keys sorted by code point order. |

#### 5.5.4 Aggregation source reference shape

Each item in `aggregation_source_refs[]` MUST be a JSON object with exactly the following members.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `ref_kind` | String | Yes | No | `source_entity`, `source_relationship`, `projected_vertex`, or `projected_edge`. |
| `ref_id` | `identifier` | Yes | No | Source or projected object ID. |
| `ref_kind_name` | `kind` | Yes | No | Source kind or projected kind. |
| `contributor_sort_key` | String | Yes | No | Contributor sort key from §6.8.3. |

Source references MUST be sorted by `contributor_sort_key`, then `ref_kind`, then `ref_id`.

### 5.6 Output ordering and sort keys

Output ordering MUST NOT depend on source input order, database row order, map iteration order, locale, timezone, insertion order, or graph engine behavior.

`vertices[]` MUST be sorted by `sort_key`, then `vertex_id`. `edges[]` MUST be sorted by `sort_key`, then `edge_id`. Validation issues MUST be sorted by §9.3. Dynamic object keys MUST be sorted by Unicode code point order.

#### 5.6.1 Sort key derivation

`sort_key` MUST be a string equal to `sk_` followed by the lowercase SHA-256 hex digest of canonical sort tuple bytes. Canonical sort tuple bytes use the length-prefixed tuple serialization rules in §7.3 with prefix `GPSORT1\n`.

| Object family | Sort tuple fields |
| --- | --- |
| Direct vertex | `vertex`, `direct`, `vertex_kind`, `source_entity_kind`, `source_entity_id`, `mapping_identity_digest`, `vertex_id` |
| Aggregated vertex | `vertex`, `aggregated`, `vertex_kind`, `aggregation_rule_id`, `canonical_grouping_key_digest`, `vertex_id` |
| Direct edge | `edge`, `direct`, `edge_kind`, `source_relationship_kind`, `source_relationship_id`, `src_vertex_id`, `dst_vertex_id`, `direction`, `mapping_identity_digest`, `false`, `edge_id` |
| Reverse edge | `edge`, `reverse`, `edge_kind`, `source_relationship_kind`, `source_relationship_id`, `src_vertex_id`, `dst_vertex_id`, `direction`, `mapping_identity_digest`, `true`, `edge_id` |
| Aggregated edge | `edge`, `aggregated`, `edge_kind`, `aggregation_rule_id`, `src_vertex_id`, `dst_vertex_id`, `direction`, `canonical_grouping_key_digest`, `edge_id` |

### 5.7 Consumer capabilities object

`consumer_capabilities` MUST be a JSON object with exactly the following members.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `query_shapes[]` | Array of strings | Yes | No | MUST include `get_graph_view`, `get_vertex`, `get_edge`, `traverse`, `list_graph_views`, and `get_projection_run`. Sorted by code point order. |
| `supports_direct_vertex_lookup` | Boolean | Yes | No | Always `true` for this revision. |
| `supports_direct_edge_lookup` | Boolean | Yes | No | Always `true` for this revision. |
| `supports_breadth_first_traversal` | Boolean | Yes | No | Always `true` for this revision. |
| `supports_alternate_traversal_order[]` | Array | Yes | No | Always `[]` in this revision. |
| `max_traversal_depth` | Integer | Yes | No | `16`. |
| `max_traversal_seed_vertices` | Integer | Yes | No | `1024`. |
| `max_kind_filters` | Integer | Yes | No | `1024`. |

No alternate traversal order is defined in this revision. A request member or transport parameter that attempts to select an alternate traversal order is invalid with `invalid_argument`.

### 5.8 Canonical output serialization

Canonical output serialization MUST use canonical JSON with these rules.

| Aspect | Required behavior |
| --- | --- |
| Encoding | UTF-8 without BOM. |
| Top-level value | One JSON object. |
| Whitespace | No insignificant whitespace is emitted. |
| Closed object member order | Members appear in the exact order declared by their schema table. |
| Dynamic object member order | Dynamic property, metadata, and details map members sort by exact Unicode code point order. |
| Arrays | Preserve the order defined by the relevant ordering rule. |
| Strings | Use the canonical string grammar below. |
| Integers | Base-10 ASCII without leading zeroes, except zero emits `0`. |
| Booleans | `true` or `false`. |
| Null | `null`. |
| Timestamps | The validated timestamp string form from §4.1. Generated lifecycle timestamps use §10.8 precision. |
| Floating point | Forbidden by scalar contracts. |

Canonical string serialization MUST use this grammar for every JSON string emitted by canonical output serialization, canonical tuple JSON fields, canonical input paths, and digest fixture transcripts.

| Character class | Required serialization |
| --- | --- |
| U+0022 quotation mark | `\"` |
| U+005C reverse solidus | `\\` |
| U+0008 backspace | `\b` |
| U+0009 tab | `\t` |
| U+000A line feed | `\n` |
| U+000C form feed | `\f` |
| U+000D carriage return | `\r` |
| Other U+0000 through U+001F controls | `\u00xx` with lowercase hexadecimal digits. |
| U+002F solidus | `/`, never escaped. |
| U+2028 and U+2029 | UTF-8 bytes, never escaped. |
| All other Unicode scalar values | UTF-8 bytes, never escaped. |

Escaped UTF-16 surrogate pairs in input JSON MUST decode to one Unicode scalar value before validation. A lone escaped surrogate is invalid JSON for this NLSpec and MUST fail before admission. Canonical output strings MUST NOT contain surrogate code points.

Canonical graph-output equivalence is byte identity of canonical JSON after excluding the run-specific fields listed in §12.2.

## 6. Mapping and derivation algorithms

Projection derivation MUST follow the algorithms and mappings in this section after input validation and normalization.

### 6.1 Projection algorithm overview

Projection derivation begins only after request admission under §4.0.1. By the time this algorithm starts, `graph_view_id`, `projection_run_id`, `source_snapshot_id`, `projection_version`, `accepted_at`, `projection_config_digest`, and `projection_source_digest` are fixed for the admitted run.

A conforming implementation MUST derive an admitted projection run in this order.

```text
1. Start from admitted decoded input after default materialization and normalization.
2. Select the active relationship mapping source:
   a. if relationship_definitions[] is non-empty, use it;
   b. otherwise use projection_config.relationship_mappings[].
3. Validate mapping completeness and mapping cross-references.
4. Evaluate global source entity and source relationship filters.
5. Exclude semantically invalid source items and emit nonfatal validation issues as required.
6. Emit direct vertices from eligible source entities.
7. Emit direct and reverse edges from eligible source relationships whose endpoint vertices exist.
8. Emit aggregated vertices and aggregated edges in normalized aggregation rule order.
9. Derive graph-view properties and metadata.
10. Build schema_registry and consumer_capabilities.
11. Run post-projection output validation.
12. Construct validation summary under §9.
13. If no fatal issue exists, publish a consumable projection result atomically.
14. If any fatal issue exists, persist the admitted run as failed with its validation summary and do not publish a consumable graph result.
```

Validation errors after admission MUST NOT be returned as lifecycle operation errors. They are observable through retained run inspection and, when a consumable graph exists, the graph-view validation summary.

### 6.2 Mapping completeness

Entity mapping completeness is mandatory. For every value in `declared_source_entity_kinds[]`, exactly one `entity_mapping_rule.source_entity_kind` MUST exist. Absence produces `missing_entity_mapping_rule` with severity `fatal`. Duplicate mappings for the same source entity kind produce `invalid_mapping_rule` with severity `fatal`.

Relationship mapping completeness is partial. For every value in `declared_source_relationship_kinds[]`, at most one active relationship mapping rule may exist. Absence produces `missing_relationship_mapping_rule` with severity `error`; source relationships of that kind are excluded. If an aggregation rule depends on a missing relationship mapping, that aggregation rule is invalid and produces `invalid_aggregation_rule` with severity `fatal`.

### 6.3 Direct vertex mapping

A source entity is eligible for direct vertex mapping only when all of the following are true:

1. Its source item shape is valid enough to provide `source_entity_id` and `source_entity_kind`.
2. Its `source_entity_kind` is declared.
3. Global entity filters evaluate true.
4. The applicable entity mapping rule's `inclusion_predicate` is `always` or evaluates true.
5. The applicable entity mapping rule is valid.

For each eligible source entity, the implementation MUST emit exactly one direct projected vertex for the applicable entity mapping rule.

Labels are derived by this table.

| `label_policy` | Output labels |
| --- | --- |
| `mapping_only` | `default_vertex_labels[] + mapping_labels[]` |
| `preserve_source` | `default_vertex_labels[] + source_entity.labels[]` |
| `mapping_then_source` | `default_vertex_labels[] + mapping_labels[] + source_entity.labels[]` |

After concatenation, exact duplicate labels are removed and remaining labels sort by exact code point order.

### 6.4 Direct edge direction matrix

Direction projection occurs before edge ID derivation. The projected endpoint IDs and projected direction from this table are the endpoint and direction fields used in the direct edge identity tuple.

| `direction_policy` | Source `direction` | Projected `src_vertex_id` | Projected `dst_vertex_id` | Projected `direction` |
| --- | --- | --- | --- | --- |
| `preserve` | `forward` | source `src_source_entity_id` mapped vertex | source `dst_source_entity_id` mapped vertex | `directed` |
| `preserve` | `reverse` | source `dst_source_entity_id` mapped vertex | source `src_source_entity_id` mapped vertex | `directed` |
| `preserve` | `undirected` | source `src_source_entity_id` mapped vertex | source `dst_source_entity_id` mapped vertex | `undirected` |
| `preserve` | `bidirectional` | source `src_source_entity_id` mapped vertex | source `dst_source_entity_id` mapped vertex | `bidirectional` |
| `normalize_forward` | any allowed source direction | source `src_source_entity_id` mapped vertex | source `dst_source_entity_id` mapped vertex | `directed` |
| `normalize_reverse` | any allowed source direction | source `dst_source_entity_id` mapped vertex | source `src_source_entity_id` mapped vertex | `directed` |
| `undirected` | any allowed source direction | source `src_source_entity_id` mapped vertex | source `dst_source_entity_id` mapped vertex | `undirected` |
| `bidirectional` | any allowed source direction | source `src_source_entity_id` mapped vertex | source `dst_source_entity_id` mapped vertex | `bidirectional` |

A source relationship is eligible for direct edge mapping only when all of the following are true:

1. Its source item shape is valid enough to provide `source_relationship_id` and `source_relationship_kind`.
2. Its `source_relationship_kind` is declared.
3. It has valid endpoints and a valid source direction after defaulting.
4. Global relationship filters evaluate true.
5. The applicable relationship mapping rule's `inclusion_predicate` is `always` or evaluates true.
6. The endpoint source entities produced direct projected vertices.
7. The applicable relationship mapping rule is valid.

For each eligible source relationship, the implementation MUST emit one primary direct projected edge.

#### 6.4.1 Direct edge label derivation

For each emitted primary direct edge, output edge labels MUST be derived as follows.

| `label_policy` | Output edge labels |
| --- | --- |
| `mapping_only` | `default_edge_labels[] + mapping_labels[]` |
| `preserve_source` | `default_edge_labels[] + source_relationship.labels[]` |
| `mapping_then_source` | `default_edge_labels[] + mapping_labels[] + source_relationship.labels[]` |

Label concatenation happens after source relationship direction projection and before direct edge metadata emission. Exact duplicate labels MUST be removed. Remaining labels MUST sort by exact code point order. Reverse edges MUST use the same derived label set as their primary edge. `reverse_edge_kind` MUST NOT alter labels.

### 6.5 Reverse-edge behavior

Reverse-edge validity is a static mapping-rule property. A conforming implementation MUST validate `emit_reverse_edge` and `direction_policy` before per-source relationship projection.

| `emit_reverse_edge` | `direction_policy` | Mapping-rule validity | Required behavior |
| --- | --- | --- | --- |
| `false` | any allowed policy | Valid | Emit only the primary edge. |
| `true` | `normalize_forward` | Valid | Emit the primary directed edge and one reverse directed edge. |
| `true` | `normalize_reverse` | Valid | Emit the primary directed edge and one reverse directed edge. |
| `true` | `preserve` | Invalid | Emit `invalid_reverse_edge_policy`; no graph is consumable. |
| `true` | `undirected` | Invalid | Emit `invalid_reverse_edge_policy`; no graph is consumable. |
| `true` | `bidirectional` | Invalid | Emit `invalid_reverse_edge_policy`; no graph is consumable. |

A valid relationship mapping with `emit_reverse_edge=true` MUST emit a second edge for each emitted primary edge.

| Aspect | Required behavior |
| --- | --- |
| Edge family | Reverse edges set `edge_family=reverse`. |
| Edge kind | `reverse_edge_kind` if supplied; otherwise `projected_edge_kind`. |
| Source reference | Same `source_relationship_ref` shape as the primary direct edge, with the same `source_relationship_id`, `source_relationship_kind`, and `mapping_rule_id`. |
| Endpoints | Primary `dst_vertex_id` becomes reverse `src_vertex_id`; primary `src_vertex_id` becomes reverse `dst_vertex_id`. |
| Direction | Always `directed`. |
| Labels | Same derived label set as the primary edge under §6.4.1. |
| Properties | Same projected property derivation as the primary edge. |
| Metadata | Reverse edge metadata sets `is_reverse_edge=true` and `reverse_of_edge_id=<primary edge_id>`. Primary edge metadata sets `is_reverse_edge=false` and `reverse_of_edge_id=null`. |
| Identity tuple | Uses the `reverse_edge` identity tuple family in §7.6. |

For static invalidity before a source-specific projected direction exists, `invalid_reverse_edge_policy.details.projected_direction` MUST be JSON null. Reverse edge identity MUST NOT collide with the primary edge identity.

### 6.6 Projected property derivation

For each graph-view, vertex, or edge output object, applicable property definitions are those whose `target_scope` matches the output scope and whose expanded concrete target kind equals the output kind. `target_kind="*"` expansion is closed by §4.10.2.

`required_property_keys[]` and `optional_property_keys[]` on mapping rules do not alter property derivation. They are validation references only. The emitted properties are determined solely by applicable `property_definitions[]` and the rules in this section.

For each applicable property definition, the implementation MUST evaluate `source_field_path` in the context allowed by §4.10.1.

| Source field state | Required behavior |
| --- | --- |
| Field missing and `missing_behavior=omit` | Omit the projected property. |
| Field missing and `missing_behavior=default` | Emit `default_value` after compatibility validation. |
| Field missing and `missing_behavior=error` | Emit `required_property_missing`; omit the projected property. |
| Field present as JSON null and `source_null_behavior=omit` | Omit the projected property. |
| Field present as JSON null and `source_null_behavior=default` | Emit `default_value` after compatibility validation. |
| Field present as JSON null and `source_null_behavior=emit_null` | Emit JSON null if `null_output_policy=emit_null`; otherwise the definition is invalid. |
| Field present as JSON null and `source_null_behavior=error` | Emit `source_null_for_required_property`; omit the projected property. |
| Field present as non-null compatible value | Emit normalized compatible value. |
| Field present as non-null incompatible value | Emit `invalid_property_type`; omit the projected property. |

A property definition whose cross-field defaults are internally inconsistent is invalid with `invalid_property_definition` and prevents a consumable graph only when the issue severity is fatal in §9.5.

### 6.7 Metadata mapping derivation

Metadata mappings follow the same field-state and compatibility rules as projected property derivation, but emit into `metadata.mapped_metadata`. They MUST NOT emit system metadata keys reserved by §5.5.

Metadata mapping source contexts are closed by this table.

| Output object | Allowed `source_field_path` context |
| --- | --- |
| Graph view | `source_metadata.<property_key>` only. |
| Direct vertex | The mapped source entity only. |
| Direct edge | The mapped source relationship only. |
| Reverse edge | The same mapped source relationship as the primary edge. |
| Aggregated vertex | Contributors from the emitting vertex aggregation rule's `input_scope` and `input_kind`. |
| Aggregated edge | Contributors from the emitting edge aggregation rule's `input_scope` and `input_kind`. |

Metadata mappings apply to aggregated outputs when `target_scope` and expanded concrete `target_kind` match. For aggregated outputs, `merge_behavior` MUST be interpreted through §6.8.6 and §6.8.6.1. Multiple aggregation rules that emit the same concrete kind evaluate metadata mappings independently for each emitted object.

This revision does not define graph-view metadata aggregation over source entities or source relationships. A graph-view metadata mapping whose `source_field_path` is not `source_metadata.<property_key>` is invalid with `invalid_metadata_mapping`. When no metadata mappings apply, `mapped_metadata` MUST be `{}`.

### 6.8 Aggregation rules

Aggregation MUST be deterministic and MUST use the algorithms in this section.

#### 6.8.1 Aggregation input selection

For each aggregation rule in `aggregation_rule_id` order, the implementation MUST select contributors from the rule's `input_scope` and `input_kind`.

| `input_scope` | Contributor set |
| --- | --- |
| `source_entity` | Source entities that remain after source-item semantic validation and global entity filtering. |
| `source_relationship` | Source relationships that remain after source-item semantic validation and global relationship filtering. |
| `projected_vertex` | Already emitted direct and aggregated vertices available before this rule executes. |
| `projected_edge` | Already emitted direct, reverse, and aggregated edges available before this rule executes. |

An aggregation rule MUST NOT depend on objects emitted by a later aggregation rule. If such a dependency exists, the rule is invalid with `invalid_aggregation_rule`.

#### 6.8.2 Grouping-key evaluation

For each contributor, evaluate every field path in `grouping_keys[]` in declared order.

| Field-path result | `missing_grouping_key_behavior=error` | `missing_grouping_key_behavior=exclude` | `missing_grouping_key_behavior=use_null` |
| --- | --- | --- | --- |
| Present value | Use the value. | Use the value. | Use the value. |
| Missing value | Emit `aggregation_grouping_key_missing`; exclude contributor. | Exclude contributor without issue. | Use JSON null as the value. |

A grouping key value MUST be a `property_value` scalar or array. Object values are invalid with `invalid_aggregation_rule`. Empty groups emit no object.

#### 6.8.3 Contributor sort key

Each contributor has a contributor sort key used for source-reference ordering and `first_by_sort` / `last_by_sort` merge behavior.

| Contributor family | Contributor sort tuple fields |
| --- | --- |
| Source entity | `source_entity`, `source_entity_kind`, `source_entity_id` |
| Source relationship | `source_relationship`, `source_relationship_kind`, `source_relationship_id` |
| Projected vertex | `projected_vertex`, `vertex_kind`, `sort_key`, `vertex_id` |
| Projected edge | `projected_edge`, `edge_kind`, `sort_key`, `edge_id` |

The contributor sort key is `csk_` plus the lowercase SHA-256 hex digest of the tuple serialized with prefix `GPCONTRIB1\n`.

#### 6.8.4 Canonical grouping-key serialization

A canonical grouping key is the canonical JSON array of evaluated grouping key values in `grouping_keys[]` order. `canonical_grouping_key_digest` is the lowercase SHA-256 hex digest of canonical tuple serialization with prefix `GPGROUP1\n` and fields:

1. `aggregation_rule_id`,
2. `target_scope`,
3. `projected_kind`,
4. canonical grouping key JSON bytes.

Two groups are the same group if and only if their `canonical_grouping_key_digest` values are equal.

#### 6.8.5 Aggregated vertex emission

For each aggregation group from a vertex aggregation rule, emit one aggregated projected vertex with:

- `vertex_kind=aggregation_rule.projected_kind`;
- `vertex_family=aggregated`;
- `labels[]` equal to normalized `projection_config.default_vertex_labels[]` sorted by exact code point order;
- `source_entity_ref=null`;
- properties and mapped metadata derived by §6.8.6.2;
- `metadata.aggregation_rule_id=aggregation_rule_id`;
- `metadata.aggregation_source_refs[]` from all group contributors, sorted by §5.5.4;
- identity derived by §7.6.

Aggregation rules in this revision do not declare additional labels. Source item labels are not available to aggregated vertices except as source-reference metadata.

#### 6.8.6 Property merge behavior

When aggregation emits properties or mapped metadata, the effective merge behavior for a key is the per-rule `property_merge_behavior` override when present, otherwise the property or metadata definition's `merge_behavior`.

| `merge_behavior` | Input values considered | Output | Conflict behavior |
| --- | --- | --- | --- |
| `single_value` | Candidate values remaining after missing/null/default/type evaluation | Emit the one distinct value. | More than one distinct value emits `aggregation_merge_conflict` and omits the key. |
| `first_by_sort` | Candidate values remaining after missing/null/default/type evaluation | Value from first candidate by contributor sort key. | No conflict. |
| `last_by_sort` | Candidate values remaining after missing/null/default/type evaluation | Value from last candidate by contributor sort key. | No conflict. |
| `distinct_sorted_array` | Present scalar values and array element values from candidates | Distinct scalar values sorted by canonical value order. | Non-scalar object or nested array emits `invalid_property_type`; invalid values are skipped. |
| `count` | All contributors in group | Integer count of contributing items. | No conflict. |
| `omit` | None | Omit key. | No conflict. |

Canonical value order sorts first by JSON type rank `null`, `boolean`, `integer`, `string`, then by canonical JSON bytes.

#### 6.8.6.1 Merge behavior compatibility

Merge behavior MUST be compatible with the declared `projected_type` for the property definition or metadata mapping that owns the emitted key.

| Merge behavior | Allowed projected types | Null handling | Array handling | Invalid combination |
| --- | --- | --- | --- | --- |
| `single_value` | all projected types | Uses candidate values produced by §6.8.6.2. JSON null candidates compare as ordinary JSON null values only when emitted by `source_null_behavior=emit_null`. | Arrays compare as whole normalized values. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `first_by_sort` | all projected types | Uses candidate values produced by §6.8.6.2. | Emits first whole normalized value. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `last_by_sort` | all projected types | Uses candidate values produced by §6.8.6.2. | Emits last whole normalized value. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `distinct_sorted_array` | `string_array`, `identifier_array` only | JSON null candidates are ignored. | Flattens scalar strings and array elements of the declared element type. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `count` | `integer` only | Counts contributors, not values. | Arrays are irrelevant. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `omit` | all projected types | No values considered. | No values considered. | Never invalid by type. |

`distinct_sorted_array` for `identifier_array` MUST validate every emitted element as `identifier`. `distinct_sorted_array` for `string_array` MUST accept only JSON strings. `single_value` conflict detection MUST compare canonical JSON bytes after projected-type normalization.

#### 6.8.6.2 Aggregated property and metadata candidate algorithm

Aggregated property and metadata derivation MUST use this deterministic algorithm.

```text
for each aggregation group in canonical_grouping_key_digest order:
  for each applicable property definition or metadata mapping in expanded concrete-key order:
    if effective merge_behavior is count:
      emit the contributor count after projected_type validation.
      skip source_field_path candidate evaluation.
      continue.

    initialize candidate_values as an empty ordered list.
    initialize contributor_issue_emitted_for_key as false.

    for each contributor in contributor_sort_key order:
      evaluate source_field_path in the contributor context.
      apply missing_behavior, source_null_behavior, default_value, null_output_policy, and projected_type compatibility.
      produce zero or one normalized candidate value.
      emit zero or one contributor-scoped validation issue when required.

    apply merge_behavior to candidate_values.
    emit the merged key when the merge result is present.
    omit the key when the merge result is absent and required=false.
    emit one group-scoped required-missing issue when the merge result is absent, required=true, and no contributor-scoped required/null issue was emitted for the same group/key.
```

Contributor candidate evaluation MUST use this table.

| Condition | Required behavior |
| --- | --- |
| Missing source field and `missing_behavior=omit` | No candidate value, no issue. |
| Missing source field and `missing_behavior=default` | One candidate value for that contributor using `default_value`. |
| Missing source field and `missing_behavior=error` | One contributor-scoped issue, no candidate value. |
| Explicit JSON null and `source_null_behavior=omit` | No candidate value, no issue. |
| Explicit JSON null and `source_null_behavior=default` | One candidate value for that contributor using `default_value`. |
| Explicit JSON null and `source_null_behavior=emit_null` | One JSON null candidate only when `null_output_policy=emit_null`; otherwise configuration validation is fatal before derivation. |
| Explicit JSON null and `source_null_behavior=error` | One contributor-scoped issue, no candidate value. |
| Non-null incompatible value | One contributor-scoped `invalid_property_type` or `invalid_metadata_mapping` issue, no candidate value. |
| `merge_behavior=count` | Emit count of contributors in the group; do not evaluate `source_field_path` for values and do not emit source-field missing issues. |
| `merge_behavior=omit` | Omit key; do not evaluate `source_field_path` for values and do not emit source-field missing issues. |
| No candidates and `required=false` | Omit key. |
| No candidates and `required=true` | Emit one group-scoped missing issue unless contributor-scoped required/null issues already represent the absence for that same group/key. |

For contributor-scoped aggregation issues, `details` MUST identify `aggregation_rule_id`, `canonical_grouping_key_digest`, `contributor_id`, `projected_key` or `projected_metadata_key`, and `source_field_path` when the issue code's detail schema includes those fields. For group-scoped aggregation issues, `details` MUST identify `aggregation_rule_id`, `canonical_grouping_key_digest`, and the emitted key. Any field required to distinguish two emitted issues MUST be a required detail field for that code and MUST participate in issue identity.

For aggregated metadata, `required=true` with no candidates emits `invalid_metadata_mapping` with `reason_code=required_metadata_missing`. For aggregated properties, `required=true` with no candidates emits `required_property_missing`.

Defaults in aggregation apply per contributor, not once per group. Requiredness is evaluated after merge, not as a requirement that every contributor provide a value, except when `missing_behavior=error` or `source_null_behavior=error` emits contributor-scoped issues.

#### 6.8.7 Aggregated edge emission

An edge aggregation rule MUST resolve source and destination aggregated vertices by evaluating `endpoint_grouping.src_grouping_keys[]` and `endpoint_grouping.dst_grouping_keys[]` in the edge aggregation contributor context.

For each edge aggregation group:

1. Evaluate source endpoint grouping keys in the contributor context for each contributor.
2. Evaluate destination endpoint grouping keys in the contributor context for each contributor.
3. Convert each endpoint grouping-key array to canonical grouping-key serialization using §6.8.4.
4. Compute endpoint digests with the same digest function as vertex aggregation grouping.
5. Match each digest to an already emitted aggregated vertex from the referenced vertex aggregation rule.
6. If either endpoint is missing or excluded under §6.8.8, do not emit the aggregated edge.
7. If both endpoints resolve, emit exactly one aggregated edge for the edge aggregation group.

An emitted aggregated edge MUST set:

- `edge_kind=aggregation_rule.projected_kind`;
- `edge_family=aggregated`;
- `labels[]` equal to normalized `projection_config.default_edge_labels[]` sorted by exact code point order;
- `src_vertex_id` equal to the matched source aggregated vertex;
- `dst_vertex_id` equal to the matched destination aggregated vertex;
- `direction=aggregation_rule.edge_direction`;
- `source_relationship_ref=null`;
- properties and mapped metadata derived by §6.8.6.2;
- `metadata.aggregation_rule_id=aggregation_rule_id`;
- `metadata.aggregation_source_refs[]` from all group contributors, sorted by §5.5.4;
- identity derived by §7.6.

Aggregation rules in this revision do not declare additional labels. Source item labels are not available to aggregated edges except as source-reference metadata.

#### 6.8.8 Aggregated endpoint key evaluation and missing endpoints

Endpoint grouping key values MUST be valid `property_value` values. JSON null is valid and participates in endpoint digest computation when it is the present value of the field path. A missing field is not the same as a present JSON null.

Endpoint grouping key evaluation MUST follow this table.

| Endpoint key result | `missing_endpoint_behavior=error` | `missing_endpoint_behavior=exclude` |
| --- | --- | --- |
| Present scalar, array, or JSON null `property_value` | Include value in canonical endpoint key array. | Include value in canonical endpoint key array. |
| Missing field | Emit `aggregation_endpoint_missing` with `reason_code=endpoint_key_missing`; do not emit edge. | Do not emit edge; no issue. |
| Object value | Emit `invalid_aggregation_rule` with `reason_code=endpoint_grouping_key_invalid`; do not emit edge. | Emit `invalid_aggregation_rule` with `reason_code=endpoint_grouping_key_invalid`; do not emit edge. |
| Nested array or non-`property_value` | Emit `invalid_aggregation_rule` with `reason_code=endpoint_grouping_key_invalid`; do not emit edge. | Emit `invalid_aggregation_rule` with `reason_code=endpoint_grouping_key_invalid`; do not emit edge. |
| Digest computed but no matching aggregated vertex exists | Emit `aggregation_endpoint_missing` with `reason_code=endpoint_vertex_not_found`; do not emit edge. | Do not emit edge; no issue. |

For `aggregation_endpoint_missing`, `details.endpoint_digest` MUST be non-null when the digest was computed. It MUST be JSON null when the endpoint key was missing before digest computation. `details.field_path` MUST be the missing or invalid endpoint grouping key path when attributable and JSON null otherwise.

## 7. Identity and stability

### 7.1 Generated ID syntax

Generated IDs MUST use the following prefixes.

| ID family | Prefix | Full syntax |
| --- | --- | --- |
| Graph view | `gv_` | `gv_` + 64 lowercase hex characters |
| Projection run | `gpr_` | `gpr_` + 64 lowercase hex characters |
| Vertex | `vx_` | `vx_` + 64 lowercase hex characters |
| Edge | `ed_` | `ed_` + 64 lowercase hex characters |
| Validation issue | `gpi_` | `gpi_` + 64 lowercase hex characters |

### 7.2 Digest function

Every generated ID and digest in this NLSpec MUST use SHA-256 over the exact canonical bytes specified for that ID or digest family. The lowercase hex encoding of the 32-byte digest is used without truncation.

### 7.3 Canonical tuple serialization

Canonical tuple serialization MUST be byte-identical across conforming implementations.

```text
serialize_tuple(prefix, fields):
  bytes = UTF8(prefix)
  for field in fields:
    value_bytes = canonical_field_bytes(field)
    bytes += ASCII(decimal_length(value_bytes))
    bytes += 0x3A  # ':'
    bytes += value_bytes
    bytes += 0x0A  # LF
  return bytes
```

`canonical_field_bytes` MUST produce:

| Field kind | Bytes |
| --- | --- |
| String token or identifier | UTF-8 bytes of the exact string. |
| Boolean | ASCII `true` or `false`. |
| Integer | Canonical integer bytes from §5.8. |
| JSON null | ASCII `null`. |
| Object or array | Canonical JSON bytes from §5.8. |
| SHA-256 digest | Lowercase hex ASCII. |

### 7.3.1 Projection digest envelopes

Projection digest envelopes are byte-level contracts. A conforming implementation MUST compute these digest bytes exactly after JSON decoding, duplicate-member rejection, validation sufficient to materialize defaults, and input normalization.

`projection_config_digest` MUST be the lowercase SHA-256 over `serialize_tuple("GPCONFIG1\n", fields)` with fields in this exact order.

| Field position | Field value |
| ---: | --- |
| 1 | `projection_schema_id` |
| 2 | `projection_config_core` |
| 3 | `active_relationship_mapping_source` |
| 4 | `active_relationship_mappings[]` |
| 5 | normalized `filters` |
| 6 | normalized `property_definitions[]` |
| 7 | normalized `metadata_mappings[]` |
| 8 | normalized `aggregation_rules[]` |

`projection_config_core` MUST be a normalized object with exactly these members in this order: `graph_view_key`, `projection_version`, `declared_source_entity_kinds[]`, `declared_source_relationship_kinds[]`, `entity_mappings[]`, `default_vertex_labels[]`, `default_edge_labels[]`, `allow_empty_kind_registry`, and `retention_policy`. It MUST exclude `custom_config`, `relationship_mappings[]`, `metadata_mappings[]`, and `aggregation_rules[]` because those families are represented by separate digest fields.

`active_relationship_mapping_source` MUST be exactly one of these string tokens.

| Token | Required condition |
| --- | --- |
| `top_level_relationship_definitions` | Top-level `relationship_definitions[]` is non-empty and `projection_config.relationship_mappings[]` is empty. |
| `projection_config_relationship_mappings` | `projection_config.relationship_mappings[]` is non-empty and top-level `relationship_definitions[]` is empty. |
| `none` | Both relationship mapping arrays are empty. |

`active_relationship_mappings[]` MUST be the normalized relationship mapping array selected by `active_relationship_mapping_source`, or `[]` when the source is `none`.

`projection_source_digest` MUST be the lowercase SHA-256 over `serialize_tuple("GPSOURCE1\n", fields)` with fields in this exact order.

| Field position | Field value |
| ---: | --- |
| 1 | `projection_schema_id` |
| 2 | `source_snapshot_id` |
| 3 | normalized `source_entities[]` |
| 4 | normalized `source_relationships[]` |
| 5 | normalized `source_metadata` |

The following fields MUST be excluded from both digest envelopes: `graph_view_id`, `requested_at`, `requested_by`, `projection_run_nonce`, and every lifecycle timestamp. `custom_config` MUST be excluded from `projection_config_digest` and MUST NOT appear in `projection_source_digest`.

Any pre-admission condition that prevents default materialization or deterministic normalization prevents digest emission and MUST fail through §4.0.1. Any admitted condition that prevents digest use for a consumable graph result MUST fail the admitted run.

#### 7.3.1.1 Golden fixture: minimal empty graph

Fixture source JSON:

```json
{"projection_schema_id":"graph_projection.v1","graph_view_id":"gv_0bfa120793d470c3cf37aa2c6ac0f69c067fa2e598da3f5116512b92f3bc3752","source_snapshot_id":"snap_empty_1","projection_config":{"graph_view_key":"empty","declared_source_entity_kinds":[],"entity_mappings":[],"allow_empty_kind_registry":true},"source_entities":[],"source_relationships":[],"requested_at":"2026-05-30T00:00:00Z","requested_by":"fixture"}
```

Normalized input after default materialization:

```json
{"projection_schema_id":"graph_projection.v1","graph_view_id":"gv_0bfa120793d470c3cf37aa2c6ac0f69c067fa2e598da3f5116512b92f3bc3752","source_snapshot_id":"snap_empty_1","projection_config":{"graph_view_key":"empty","projection_version":"1","declared_source_entity_kinds":[],"declared_source_relationship_kinds":[],"entity_mappings":[],"relationship_mappings":[],"metadata_mappings":[],"aggregation_rules":[],"default_vertex_labels":[],"default_edge_labels":[],"allow_empty_kind_registry":true,"retention_policy":{"retain_replaced_results":true,"retention_count":5,"retention_duration_seconds":2592000,"retain_failed_results":true,"failed_retention_count":20,"failed_retention_duration_seconds":2592000},"custom_config":{}},"source_entities":[],"source_relationships":[],"source_metadata":{},"filters":{"entity_filters":[],"relationship_filters":[],"logic":"and"},"relationship_definitions":[],"property_definitions":[],"requested_at":"2026-05-30T00:00:00Z","requested_by":"fixture"}
```

| Digest field | Canonical bytes |
| --- | --- |
| `projection_config_core` | `{"graph_view_key":"empty","projection_version":"1","declared_source_entity_kinds":[],"declared_source_relationship_kinds":[],"entity_mappings":[],"default_vertex_labels":[],"default_edge_labels":[],"allow_empty_kind_registry":true,"retention_policy":{"retain_replaced_results":true,"retention_count":5,"retention_duration_seconds":2592000,"retain_failed_results":true,"failed_retention_count":20,"failed_retention_duration_seconds":2592000}}` |
| `active_relationship_mapping_source` | `none` |
| `active_relationship_mappings[]` | `[]` |
| `filters` | `{"entity_filters":[],"relationship_filters":[],"logic":"and"}` |
| `property_definitions[]` | `[]` |
| `metadata_mappings[]` | `[]` |
| `aggregation_rules[]` | `[]` |
| `source_entities[]` | `[]` |
| `source_relationships[]` | `[]` |
| `source_metadata` | `{}` |

| Tuple | Lowercase hex bytes |
| --- | --- |
| `projection_config_digest` input | `4750434f4e464947310a31393a67726170685f70726f6a656374696f6e2e76310a3434313a7b2267726170685f766965775f6b6579223a22656d707479222c2270726f6a656374696f6e5f76657273696f6e223a2231222c226465636c617265645f736f757263655f656e746974795f6b696e6473223a5b5d2c226465636c617265645f736f757263655f72656c6174696f6e736869705f6b696e6473223a5b5d2c22656e746974795f6d617070696e6773223a5b5d2c2264656661756c745f7665727465785f6c6162656c73223a5b5d2c2264656661756c745f656467655f6c6162656c73223a5b5d2c22616c6c6f775f656d7074795f6b696e645f7265676973747279223a747275652c22726574656e74696f6e5f706f6c696379223a7b2272657461696e5f7265706c616365645f726573756c7473223a747275652c22726574656e74696f6e5f636f756e74223a352c22726574656e74696f6e5f6475726174696f6e5f7365636f6e6473223a323539323030302c2272657461696e5f6661696c65645f726573756c7473223a747275652c226661696c65645f726574656e74696f6e5f636f756e74223a32302c226661696c65645f726574656e74696f6e5f6475726174696f6e5f7365636f6e6473223a323539323030307d7d0a343a6e6f6e650a323a5b5d0a36313a7b22656e746974795f66696c74657273223a5b5d2c2272656c6174696f6e736869705f66696c74657273223a5b5d2c226c6f676963223a22616e64227d0a323a5b5d0a323a5b5d0a323a5b5d0a` |
| `projection_source_digest` input | `4750534f55524345310a31393a67726170685f70726f6a656374696f6e2e76310a31323a736e61705f656d7074795f310a323a5b5d0a323a5b5d0a323a7b7d0a` |

| Digest | Expected lowercase SHA-256 |
| --- | --- |
| `graph_view_id` | `gv_0bfa120793d470c3cf37aa2c6ac0f69c067fa2e598da3f5116512b92f3bc3752` |
| `projection_config_digest` | `c0d919dc8a5093e2e6f81eab3a4f2b0a9e03381e5fc834144f33a3d737fb2b06` |
| `projection_source_digest` | `e7bb613e7b8b295359e0ebba2f7ab6fe845b00f5de94f82f599327a409dcc56c` |

#### 7.3.1.2 Golden fixture: one host property graph

Fixture source JSON:

```json
{"projection_schema_id":"graph_projection.v1","graph_view_id":"gv_7b34489c234cb6caa432f92afc6fb122788e525f8bb057d214c96be9289d8893","source_snapshot_id":"snap_incident_1","projection_config":{"graph_view_key":"incident_graph","declared_source_entity_kinds":["host"],"declared_source_relationship_kinds":["logon"],"entity_mappings":[{"mapping_rule_id":"map_host","source_entity_kind":"host","projected_vertex_kind":"host_vertex","required_property_keys":["hostname"]}],"relationship_mappings":[{"mapping_rule_id":"map_logon","source_relationship_kind":"logon","projected_edge_kind":"logon_edge"}]},"source_entities":[{"source_entity_id":"host1","source_entity_kind":"host","properties":{"hostname":"WS-023"}}],"source_relationships":[],"source_metadata":{"case":"alpha"},"property_definitions":[{"property_definition_id":"pd_hostname","target_scope":"vertex","target_kind":"host_vertex","source_field_path":"properties.hostname","projected_key":"hostname","projected_type":"string","required":true}],"requested_at":"2026-05-30T00:00:00Z","requested_by":"fixture"}
```

Normalized input after default materialization:

```json
{"projection_schema_id":"graph_projection.v1","graph_view_id":"gv_7b34489c234cb6caa432f92afc6fb122788e525f8bb057d214c96be9289d8893","source_snapshot_id":"snap_incident_1","projection_config":{"graph_view_key":"incident_graph","projection_version":"1","declared_source_entity_kinds":["host"],"declared_source_relationship_kinds":["logon"],"entity_mappings":[{"mapping_rule_id":"map_host","source_entity_kind":"host","projected_vertex_kind":"host_vertex","inclusion_predicate":"always","label_policy":"mapping_only","mapping_labels":[],"required_property_keys":["hostname"],"optional_property_keys":[]}],"relationship_mappings":[{"mapping_rule_id":"map_logon","source_relationship_kind":"logon","projected_edge_kind":"logon_edge","inclusion_predicate":"always","direction_policy":"preserve","emit_reverse_edge":false,"reverse_edge_kind":"logon_edge","label_policy":"mapping_only","mapping_labels":[],"required_property_keys":[],"optional_property_keys":[]}],"metadata_mappings":[],"aggregation_rules":[],"default_vertex_labels":[],"default_edge_labels":[],"allow_empty_kind_registry":false,"retention_policy":{"retain_replaced_results":true,"retention_count":5,"retention_duration_seconds":2592000,"retain_failed_results":true,"failed_retention_count":20,"failed_retention_duration_seconds":2592000},"custom_config":{}},"source_entities":[{"source_entity_id":"host1","source_entity_kind":"host","properties":{"hostname":"WS-023"},"metadata":{},"labels":[]}],"source_relationships":[],"source_metadata":{"case":"alpha"},"filters":{"entity_filters":[],"relationship_filters":[],"logic":"and"},"relationship_definitions":[],"property_definitions":[{"property_definition_id":"pd_hostname","target_scope":"vertex","target_kind":"host_vertex","source_field_path":"properties.hostname","projected_key":"hostname","projected_type":"string","required":true,"missing_behavior":"error","source_null_behavior":"error","null_output_policy":"omit","merge_behavior":"single_value"}],"requested_at":"2026-05-30T00:00:00Z","requested_by":"fixture"}
```

| Digest field | Canonical bytes |
| --- | --- |
| `projection_config_core` | `{"graph_view_key":"incident_graph","projection_version":"1","declared_source_entity_kinds":["host"],"declared_source_relationship_kinds":["logon"],"entity_mappings":[{"mapping_rule_id":"map_host","source_entity_kind":"host","projected_vertex_kind":"host_vertex","inclusion_predicate":"always","label_policy":"mapping_only","mapping_labels":[],"required_property_keys":["hostname"],"optional_property_keys":[]}],"default_vertex_labels":[],"default_edge_labels":[],"allow_empty_kind_registry":false,"retention_policy":{"retain_replaced_results":true,"retention_count":5,"retention_duration_seconds":2592000,"retain_failed_results":true,"failed_retention_count":20,"failed_retention_duration_seconds":2592000}}` |
| `active_relationship_mapping_source` | `projection_config_relationship_mappings` |
| `active_relationship_mappings[]` | `[{"mapping_rule_id":"map_logon","source_relationship_kind":"logon","projected_edge_kind":"logon_edge","inclusion_predicate":"always","direction_policy":"preserve","emit_reverse_edge":false,"reverse_edge_kind":"logon_edge","label_policy":"mapping_only","mapping_labels":[],"required_property_keys":[],"optional_property_keys":[]}]` |
| `filters` | `{"entity_filters":[],"relationship_filters":[],"logic":"and"}` |
| `property_definitions[]` | `[{"property_definition_id":"pd_hostname","target_scope":"vertex","target_kind":"host_vertex","source_field_path":"properties.hostname","projected_key":"hostname","projected_type":"string","required":true,"missing_behavior":"error","source_null_behavior":"error","null_output_policy":"omit","merge_behavior":"single_value"}]` |
| `metadata_mappings[]` | `[]` |
| `aggregation_rules[]` | `[]` |
| `source_entities[]` | `[{"source_entity_id":"host1","source_entity_kind":"host","properties":{"hostname":"WS-023"},"metadata":{},"labels":[]}]` |
| `source_relationships[]` | `[]` |
| `source_metadata` | `{"case":"alpha"}` |

| Tuple | Lowercase hex bytes |
| --- | --- |
| `projection_config_digest` input | `4750434f4e464947310a31393a67726170685f70726f6a656374696f6e2e76310a3730373a7b2267726170685f766965775f6b6579223a22696e636964656e745f6772617068222c2270726f6a656374696f6e5f76657273696f6e223a2231222c226465636c617265645f736f757263655f656e746974795f6b696e6473223a5b22686f7374225d2c226465636c617265645f736f757263655f72656c6174696f6e736869705f6b696e6473223a5b226c6f676f6e225d2c22656e746974795f6d617070696e6773223a5b7b226d617070696e675f72756c655f6964223a226d61705f686f7374222c22736f757263655f656e746974795f6b696e64223a22686f7374222c2270726f6a65637465645f7665727465785f6b696e64223a22686f73745f766572746578222c22696e636c7573696f6e5f707265646963617465223a22616c77617973222c226c6162656c5f706f6c696379223a226d617070696e675f6f6e6c79222c226d617070696e675f6c6162656c73223a5b5d2c2272657175697265645f70726f70657274795f6b657973223a5b22686f73746e616d65225d2c226f7074696f6e616c5f70726f70657274795f6b657973223a5b5d7d5d2c2264656661756c745f7665727465785f6c6162656c73223a5b5d2c2264656661756c745f656467655f6c6162656c73223a5b5d2c22616c6c6f775f656d7074795f6b696e645f7265676973747279223a66616c73652c22726574656e74696f6e5f706f6c696379223a7b2272657461696e5f7265706c616365645f726573756c7473223a747275652c22726574656e74696f6e5f636f756e74223a352c22726574656e74696f6e5f6475726174696f6e5f7365636f6e6473223a323539323030302c2272657461696e5f6661696c65645f726573756c7473223a747275652c226661696c65645f726574656e74696f6e5f636f756e74223a32302c226661696c65645f726574656e74696f6e5f6475726174696f6e5f7365636f6e6473223a323539323030307d7d0a33393a70726f6a656374696f6e5f636f6e6669675f72656c6174696f6e736869705f6d617070696e67730a3332393a5b7b226d617070696e675f72756c655f6964223a226d61705f6c6f676f6e222c22736f757263655f72656c6174696f6e736869705f6b696e64223a226c6f676f6e222c2270726f6a65637465645f656467655f6b696e64223a226c6f676f6e5f65646765222c22696e636c7573696f6e5f707265646963617465223a22616c77617973222c22646972656374696f6e5f706f6c696379223a227072657365727665222c22656d69745f726576657273655f65646765223a66616c73652c22726576657273655f656467655f6b696e64223a226c6f676f6e5f65646765222c226c6162656c5f706f6c696379223a226d617070696e675f6f6e6c79222c226d617070696e675f6c6162656c73223a5b5d2c2272657175697265645f70726f70657274795f6b657973223a5b5d2c226f7074696f6e616c5f70726f70657274795f6b657973223a5b5d7d5d0a36313a7b22656e746974795f66696c74657273223a5b5d2c2272656c6174696f6e736869705f66696c74657273223a5b5d2c226c6f676963223a22616e64227d0a3332333a5b7b2270726f70657274795f646566696e6974696f6e5f6964223a2270645f686f73746e616d65222c227461726765745f73636f7065223a22766572746578222c227461726765745f6b696e64223a22686f73745f766572746578222c22736f757263655f6669656c645f70617468223a2270726f706572746965732e686f73746e616d65222c2270726f6a65637465645f6b6579223a22686f73746e616d65222c2270726f6a65637465645f74797065223a22737472696e67222c227265717569726564223a747275652c226d697373696e675f6265686176696f72223a226572726f72222c22736f757263655f6e756c6c5f6265686176696f72223a226572726f72222c226e756c6c5f6f75747075745f706f6c696379223a226f6d6974222c226d657267655f6265686176696f72223a2273696e676c655f76616c7565227d5d0a323a5b5d0a323a5b5d0a` |
| `projection_source_digest` input | `4750534f55524345310a31393a67726170685f70726f6a656374696f6e2e76310a31353a736e61705f696e636964656e745f310a3131393a5b7b22736f757263655f656e746974795f6964223a22686f737431222c22736f757263655f656e746974795f6b696e64223a22686f7374222c2270726f70657274696573223a7b22686f73746e616d65223a2257532d303233227d2c226d65746164617461223a7b7d2c226c6162656c73223a5b5d7d5d0a323a5b5d0a31363a7b2263617365223a22616c706861227d0a` |

| Digest | Expected lowercase SHA-256 |
| --- | --- |
| `graph_view_id` | `gv_7b34489c234cb6caa432f92afc6fb122788e525f8bb057d214c96be9289d8893` |
| `projection_config_digest` | `30afd15b557695969b76d262b8942a70cd0da3bb84176ccaf64dc233eb57731f` |
| `projection_source_digest` | `2a63da49700e5cfe7bfe0fd5e8f9d0e8ccbd03d8297abf3cef6774fd5c578e31` |

A conforming fixture harness MUST compare every transcript canonical value and tuple byte transcript before comparing the final SHA-256 digest. A digest-only comparison is not sufficient for `GP-AC-032` or `GP-AC-068`.

### 7.4 Graph-view identity

`graph_view_id` MUST equal `gv_` plus SHA-256 over tuple prefix `GPID1\n` and fields:

1. `graph_view`,
2. `projection_schema_id`,
3. `projection_config.graph_view_key`.

The caller supplies `graph_view_id`, but the supplied value is accepted only if it equals this derived value.

### 7.5 Mapping and aggregation identity digests

Mapping and aggregation identity digests are lowercase SHA-256 hex strings. They are not emitted as top-level output members unless needed inside sort-key or diagnostic details, but they participate in object identity.

| Digest | Tuple prefix | Fields |
| --- | --- | --- |
| Entity mapping identity digest | `GPMAPENTITY1\n` | `mapping_rule_id`, `source_entity_kind`, `projected_vertex_kind` |
| Relationship mapping identity digest | `GPMAPREL1\n` | `mapping_rule_id`, `source_relationship_kind`, `projected_edge_kind`, `direction_policy`, `emit_reverse_edge`, `reverse_edge_kind` |
| Aggregation identity digest | `GPAGG1\n` | `aggregation_rule_id`, `target_scope`, `input_scope`, `input_kind`, `projected_kind`, `grouping_keys[]`, `missing_grouping_key_behavior`, `edge_direction` or JSON null, `endpoint_grouping` or JSON null |

Changing labels, label policy, display-only metadata mappings, property definitions not used in identity-participating aggregation grouping, or other non-identity output fields MUST NOT change vertex or edge IDs.

### 7.6 Canonical identity tuples

Generated IDs use the following tuple families.

| Object | ID prefix | Tuple prefix | Tuple fields |
| --- | --- | --- | --- |
| Projection run | `gpr_` | `GPRUN1\n` | `projection_run`, `projection_schema_id`, `graph_view_id`, `source_snapshot_id`, `projection_config_digest`, `projection_source_digest`, `projection_run_nonce` |
| Direct projected vertex | `vx_` | `GPVERTEX1\n` | `direct_vertex`, `projection_schema_id`, `graph_view_id`, `source_entity_kind`, `source_entity_id`, `mapping_identity_digest` |
| Aggregated projected vertex | `vx_` | `GPVERTEX1\n` | `aggregated_vertex`, `projection_schema_id`, `graph_view_id`, `aggregation_identity_digest`, `canonical_grouping_key_digest` |
| Direct projected edge | `ed_` | `GPEDGE1\n` | `direct_edge`, `projection_schema_id`, `graph_view_id`, `source_relationship_kind`, `source_relationship_id`, `edge_kind`, `src_vertex_id`, `dst_vertex_id`, `direction`, `mapping_identity_digest` |
| Reverse projected edge | `ed_` | `GPEDGE1\n` | `reverse_edge`, `projection_schema_id`, `graph_view_id`, `source_relationship_kind`, `source_relationship_id`, `edge_kind`, `src_vertex_id`, `dst_vertex_id`, `direction`, `mapping_identity_digest` |
| Aggregated projected edge | `ed_` | `GPEDGE1\n` | `aggregated_edge`, `projection_schema_id`, `graph_view_id`, `aggregation_identity_digest`, `src_vertex_id`, `dst_vertex_id`, `direction`, `canonical_grouping_key_digest` |

`projection_run_nonce` is a run-specific opaque identifier generated at run acceptance. It MUST contain at least 128 bits of unpredictability or be an implementation-owned monotonic unique value whose canonical string form is collision-free within the implementation. Differences in `projection_run_nonce` and derived `projection_run_id` are intentional run-specific variance.

### 7.7 Validation issue identity

A validation issue ID is defined only for an admitted projection run. Pre-admission operation errors MUST NOT emit `issue_id`, MUST NOT emit a validation issue object, and MUST NOT use sentinel `graph_view_id`, sentinel `projection_run_id`, raw input bytes, or implementation-local attempt IDs to construct issue identity.

For every admitted run, `graph_view_id` and `projection_run_id` MUST be fixed before any validation issue is constructed. If the implementation cannot derive either value, the request is pre-admission and MUST fail through the lifecycle operation error contract.

For an admitted run, a validation issue ID MUST equal `gpi_` plus SHA-256 over tuple prefix `GPISSUE1\n` and fields:

1. `projection_schema_id`,
2. `graph_view_id`,
3. `projection_run_id`,
4. `severity`,
5. `code`,
6. `target_kind`,
7. `target_id`,
8. canonical JSON of required code-specific `details` members only.

The human-readable `message`, optional details, query context, transport-level context, and lifecycle operation context MUST NOT affect `issue_id`.

## 8. Defaults, omissions, and explicit nulls

### 8.1 Default and omission table

| Boundary | Omitted behavior | Explicit `null` behavior |
| --- | --- | --- |
| Top-level required input member | Pre-admission lifecycle `invalid_projection_request` with `reason_code=missing_required_member` when required for admission; admitted-run `missing_required_input` only for post-admission nested validation. | Pre-admission lifecycle `invalid_projection_request` with `reason_code=explicit_null_not_allowed` when required for admission; admitted-run `explicit_null_not_allowed` only for post-admission nested validation. |
| Top-level optional object | Materialize documented default before admission when needed for identity or digest. | Pre-admission lifecycle `invalid_projection_request` with `reason_code=explicit_null_not_allowed` when it prevents admission; otherwise admitted-run `explicit_null_not_allowed`. |
| Top-level optional array | Materialize `[]` unless another default is stated before admission when needed for identity or digest. | Pre-admission lifecycle `invalid_projection_request` with `reason_code=explicit_null_not_allowed` when it prevents admission; otherwise admitted-run `explicit_null_not_allowed`. |
| `projection_schema_id` | Pre-admission lifecycle `invalid_projection_request` with `reason_code=missing_required_member`; no default exists. | Pre-admission lifecycle `invalid_projection_request` with `reason_code=explicit_null_not_allowed`. |
| `projection_config.projection_version` | `1` | `explicit_null_not_allowed` |
| `projection_config.custom_config` | `{}` and no observable effect | `explicit_null_not_allowed` |
| `source_entity.properties` / `metadata` | `{}` | `explicit_null_not_allowed` |
| `source_relationship.direction` | `forward` | `explicit_null_not_allowed` |
| Missing source relationship endpoint | `missing_relationship_endpoint`; source relationship excluded | `explicit_null_not_allowed` |
| Property `default_value` | No default value exists | Valid only when compatible and `null_output_policy=emit_null` |
| Property source field missing | Governed by `missing_behavior` | Not applicable |
| Property source field present as null | Not applicable | Governed by `source_null_behavior` |
| Metadata source field missing | Governed by `missing_behavior` | Not applicable |
| Metadata source field present as null | Not applicable | Governed by `source_null_behavior` |
| Relationship mapping `direction_policy` | `preserve` | `explicit_null_not_allowed` |
| Relationship mapping `emit_reverse_edge` | `false` | `explicit_null_not_allowed` |
| Retention policy | Defaults in §10.6 | `explicit_null_not_allowed` |
| Query optional `projection_run_id` | Latest available run | `invalid_argument` |
| Query pagination `limit` | `100` | `invalid_argument` |
| Query pagination `cursor_token` | First page | `invalid_argument` |

### 8.2 Empty-input behavior

When `source_entities[]` is empty and the kind registry is valid, projection MUST emit a consumable graph with `vertices=[]`, `edges=[]`, and a validation summary with no fatal issues unless aggregation rules or configuration errors make the input invalid.

When `source_relationships[]` is empty, no direct or reverse edges are emitted. Aggregated edges MAY still be emitted only from projected-edge or other aggregation inputs that exist.

When both source arrays are empty and `allow_empty_kind_registry=true`, the graph MAY have an empty schema registry except for declared property and metadata schemas. The result remains consumable if no fatal issue exists.

## 9. Validation behavior

### 9.1 Validation phases and discovery order

This section applies only after run admission. Pre-admission failures are lifecycle operation errors under §4.0.1 and §10.0.

Admitted-run validation MUST execute in the phases in this table. A fatal issue in a phase whose `Stops later phases when fatal?` value is `Yes` prevents all later phases from running.

| Phase | Scope | Stops later phases when fatal? |
| ---: | --- | ---: |
| 1 | Scalar contracts, enum contracts, and admitted whole-input resource limits not resolved before admission | Yes |
| 2 | Projection configuration schemas and default materialization that depend on admitted identity | Yes |
| 3 | Field-path, filter, mapping, property, metadata, aggregation, retention, wildcard, and merge-compatibility validation | Yes |
| 4 | Source entity and source relationship item validation and item exclusion | No |
| 5 | Direct mapping derivation validation | No, unless a fatal issue is produced |
| 6 | Aggregation derivation validation | No, unless a fatal issue is produced |
| 7 | Output schema and projected-output resource-limit validation | Yes |
| 8 | Validation summary construction | N/A |

Validation discovery order inside each phase MUST be deterministic.

| Family | Required discovery order |
| --- | --- |
| Closed object members | The member order declared by the schema table. |
| Dynamic object keys | Exact Unicode code point order. |
| Arrays before contained IDs are valid | Input order. |
| Arrays after contained IDs are valid and the section requires normalization | Normalized order from §4.13. |
| Source entities | `source_entity_kind`, then `source_entity_id` after both fields are valid; otherwise input order. |
| Source relationships | `source_relationship_kind`, then `source_relationship_id` after both fields are valid; otherwise input order. |
| Entity and relationship mapping rules | `mapping_rule_id`. |
| Property definitions | `property_definition_id`. |
| Metadata mappings | `metadata_mapping_id`. |
| Aggregation rules | `aggregation_rule_id`. |
| Cross-reference checks | Referencing object identity, then referenced field path, then referenced identifier. |

The implementation MUST construct all reachable admitted-run issues under the phase rules, then apply the issue cap selection and emission algorithm in §9.3. It MUST NOT choose an arbitrary subset of issues.

### 9.2 Validation severities

| Severity | Meaning | Consumability |
| --- | --- | --- |
| `fatal` | Projection cannot emit a conforming consumable graph result. | No consumable graph is published for the run. |
| `error` | A source item, derived object, property, metadata key, or aggregation output is excluded or degraded under a specified rule. | Graph may be consumable when no fatal issue exists. |
| `warning` | Projection completed with a non-blocking concern that does not change required output shape. | Graph is consumable. |
| `info` | Diagnostic information only. | Graph is consumable. |

No `error` severity issue may by itself make the graph non-consumable. Any behavior that makes a graph non-consumable MUST use a `fatal` issue.

### 9.3 Validation issue shape, ordering, identity, and cap behavior

A validation issue object applies only to an admitted projection run. A validation issue MUST be a JSON object with exactly the following members.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `issue_id` | `generated_id` with prefix `gpi_` | Yes | No | Derived by §7.7. |
| `severity` | String | Yes | No | `fatal`, `error`, `warning`, or `info`. |
| `code` | String | Yes | No | One code from §9.5. |
| `target_kind` | String | Yes | No | One target kind from §9.6. |
| `target_id` | `identifier` or string | Yes | No | Derived by §9.6.1. |
| `field_path` | `field_path` or canonical input path string | Yes | Yes | The applicable §4.2 field path or §4.0 canonical input path; otherwise `null`. |
| `message` | String | Yes | No | Human-readable non-normative diagnostic. It MUST NOT affect issue identity. It MUST obey `max_validation_message_length`. |
| `details` | Object | Yes | No | Closed per validation code by §9.6. Unknown detail keys are forbidden. |

Ordinary validation issues MUST sort for output by:

1. severity order `fatal`, `error`, `warning`, `info`,
2. `code`,
3. `target_kind`,
4. `target_id`,
5. `issue_id`.

Issue-cap behavior is a two-stage selection and emission algorithm. Let `N=max_validation_issues`.

```text
1. Construct every reachable validation issue in validation discovery order.
2. If total discovered issue count <= N:
   a. Select every issue.
   b. Sort selected issues by ordinary validation issue ordering.
3. If total discovered issue count > N:
   a. Select the first N - 1 non-cap issues in validation discovery order.
   b. Construct one validation_issue_limit_exceeded issue.
   c. Sort the selected non-cap issues by ordinary validation issue ordering.
   d. Append validation_issue_limit_exceeded as the final issues[] element.
4. Severity counts MUST include every emitted issue, including the cap issue.
```

The cap issue is the only issue exempt from ordinary emission sorting. It MUST always be the final `issues[]` element when emitted. The cap issue MUST have `target_kind=projection_input`, `target_id=projection_input`, `field_path=null`, and `details.limit=N`.

### 9.4 Validation summary shape

A validation summary MUST be a JSON object with exactly these members.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `status` | String | Yes | No | `passed`, `passed_with_warnings`, `passed_with_errors`, or `failed`. |
| `fatal_count` | Integer | Yes | No | Count of fatal issues. |
| `error_count` | Integer | Yes | No | Count of error issues. |
| `warning_count` | Integer | Yes | No | Count of warning issues. |
| `info_count` | Integer | Yes | No | Count of info issues. |
| `issues[]` | Array of validation issues | Yes | No | Sorted by §9.3. |

`status` MUST be derived as follows.

| Condition | `status` |
| --- | --- |
| `fatal_count > 0` | `failed` |
| `fatal_count = 0` and `error_count > 0` | `passed_with_errors` |
| `fatal_count = 0`, `error_count = 0`, and `warning_count > 0` | `passed_with_warnings` |
| All counts zero or only `info_count > 0` | `passed` |

### 9.4.1 Validation condition construction matrix

Every admitted-run validation condition MUST map to exactly one validation construction row. The table below is closed for this revision. A condition not listed here MUST NOT produce a validation issue unless a later NLSpec revision adds a row.

| Phase | Triggering condition | Code | Severity | Target kind | Target ID derivation | Field path | Required details | Required reason code |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Configuration | Duplicate entity mapping for the same `source_entity_kind` | `invalid_mapping_rule` | `fatal` | `mapping_rule` | Later duplicate `mapping_rule_id` when valid; otherwise canonical input path | Mapping rule path | `mapping_rule_id`, `reason_code` | `duplicate_source_entity_kind_mapping` |
| Configuration | Duplicate relationship mapping for the same `source_relationship_kind` | `invalid_mapping_rule` | `fatal` | `mapping_rule` | Later duplicate `mapping_rule_id` when valid; otherwise canonical input path | Mapping rule path | `mapping_rule_id`, `reason_code` | `duplicate_source_relationship_kind_mapping` |
| Configuration | Duplicate mapping rule ID | `invalid_mapping_rule` | `fatal` | `mapping_rule` | Duplicate `mapping_rule_id` | Mapping rule path | `mapping_rule_id`, `reason_code` | `duplicate_mapping_rule_id` |
| Configuration | Missing entity mapping for a declared source entity kind | `missing_entity_mapping_rule` | `fatal` | `mapping_rule` | Missing source entity kind | `null` | `source_entity_kind` | none |
| Configuration | Missing relationship mapping for a declared source relationship kind | `missing_relationship_mapping_rule` | `error` | `mapping_rule` | Missing source relationship kind | `null` | `source_relationship_kind` | none |
| Configuration | Top-level and config relationship mappings are both non-empty | `invalid_projection_config` | `fatal` | `projection_config` | `projection_config` | Canonical input path | `field`, `reason_code` | `relationship_mapping_source_conflict` |
| Configuration | Endpoint grouping key count mismatch | `invalid_aggregation_rule` | `fatal` | `mapping_rule` | `aggregation_rule_id` | Canonical input path | `aggregation_rule_id`, `reason_code` | `endpoint_grouping_key_count_mismatch` |
| Configuration | Endpoint grouping key value is invalid for endpoint digest evaluation | `invalid_aggregation_rule` | `fatal` | `mapping_rule` | `aggregation_rule_id` | Endpoint `field_path` | `aggregation_rule_id`, `reason_code` | `endpoint_grouping_key_invalid` |
| Derivation | Aggregated endpoint key is missing and behavior is `error` | `aggregation_endpoint_missing` | `error` | `mapping_rule` | `aggregation_rule_id` | Endpoint `field_path` | `aggregation_rule_id`, `endpoint_side`, `reason_code`, `endpoint_digest`, `field_path` | `endpoint_key_missing` |
| Derivation | Aggregated endpoint digest has no matching vertex and behavior is `error` | `aggregation_endpoint_missing` | `error` | `mapping_rule` | `aggregation_rule_id` | `null` | `aggregation_rule_id`, `endpoint_side`, `reason_code`, `endpoint_digest`, `field_path` | `endpoint_vertex_not_found` |
| Derivation | Aggregation merge cannot emit one value | `aggregation_merge_conflict` | `error` | `mapping_rule` | `aggregation_rule_id` | `null` | `aggregation_rule_id`, `canonical_grouping_key_digest`, `projected_key` | none |
| Derivation | Required aggregated property has no candidate value | `required_property_missing` | `error` | `property` | `<aggregation_rule_id>#<canonical_grouping_key_digest>#<projected_key>` | Owning `source_field_path` | `projected_key`, `source_field_path`, `output_object_id`, `aggregation_rule_id`, `canonical_grouping_key_digest` | none |
| Derivation | Required aggregated metadata has no candidate value | `invalid_metadata_mapping` | `fatal` | `mapping_rule` | `metadata_mapping_id` | Owning `source_field_path` | `metadata_mapping_id`, `reason_code` | `required_metadata_missing` |
| Derivation | Computation fails after admission before ordinary validation summary | `projection_computation_failed` | `fatal` | `graph_view` | `graph_view_id` | `null` | `reason_code` | one of §9.6.2 |

Codes and reason codes listed in §9.5 and §9.6.2 that are not represented by a more specific row above use the detail and construction rules in §9.6 and §9.6.1. A conformance fixture that exercises a validation condition MUST assert the row selected by this matrix.

### 9.5 Closed validation code registry

The implementation MUST emit only the validation codes in this table.

| Code | Severity | Target kind | Required meaning |
| --- | --- | --- | --- |
| `invalid_input_shape` | `fatal` | `projection_input` | An admitted-run input shape, scalar type, or schema structure is invalid. Pre-admission JSON decoding and duplicate-member failures use lifecycle `invalid_projection_request`. |
| `missing_required_input` | `fatal` | `projection_input` | A required member is omitted. |
| `explicit_null_not_allowed` | `fatal` | `projection_input` | Explicit JSON null was supplied where null is forbidden. |
| `unknown_member` | `fatal` | `projection_input` | A closed object contains an undeclared member. |
| `duplicate_identifier` | `fatal` | `projection_input` | A uniqueness constraint is violated. |
| `invalid_projection_schema` | `fatal` | `projection_input` | `projection_schema_id` is not supported by this NLSpec. |
| `invalid_graph_view_id` | `fatal` | `graph_view` | Supplied graph view ID is syntactically invalid or does not equal the derived ID. |
| `invalid_projection_config` | `fatal` | `projection_config` | Projection configuration is internally inconsistent. |
| `invalid_field_path` | `fatal` | `projection_config` | Field path does not match the closed grammar or is invalid for its scope. |
| `invalid_filter` | `fatal` | `filter` | Filter predicate schema or operator/value combination is invalid. |
| `invalid_mapping_rule` | `fatal` | `mapping_rule` | Mapping rule schema, uniqueness, property-key reference, or cross-reference is invalid. |
| `missing_entity_mapping_rule` | `fatal` | `mapping_rule` | Required entity mapping is absent for a declared source entity kind. |
| `missing_relationship_mapping_rule` | `error` | `mapping_rule` | Relationship mapping is absent for a declared source relationship kind; affected relationships are excluded. |
| `invalid_metadata_mapping` | `fatal` | `mapping_rule` | Metadata mapping schema, compatibility, wildcard collision, or reserved-key behavior is invalid. |
| `invalid_property_definition` | `fatal` | `property_definition` | Property definition schema, wildcard collision, or cross-field behavior is invalid. |
| `invalid_property_type` | `error` | `property` | A source or default value is incompatible with the declared projected type. |
| `required_property_missing` | `error` | `property` | A required property source field is missing. |
| `source_null_for_required_property` | `error` | `property` | A required property source field is explicit JSON null and null is not allowed. |
| `undeclared_source_kind` | `error` | `source_item` | Source item kind is not declared and item is excluded. |
| `source_item_resource_limit_exceeded` | `error` | `source_item` | A source item exceeds an item-scoped resource limit and is excluded. |
| `missing_relationship_endpoint` | `error` | `source_relationship` | Source relationship endpoint is omitted and relationship is excluded. |
| `relationship_endpoint_not_projected` | `error` | `source_relationship` | Endpoint source entity is missing, excluded, or unmapped. |
| `invalid_relationship_direction` | `error` | `source_relationship` | Explicit source relationship direction is invalid and relationship is excluded. |
| `invalid_direction_policy` | `fatal` | `mapping_rule` | Relationship mapping uses an invalid direction policy. |
| `invalid_reverse_edge_policy` | `fatal` | `mapping_rule` | Reverse-edge emission is incompatible with the static direction policy. |
| `invalid_aggregation_rule` | `fatal` | `mapping_rule` | Aggregation rule schema, dependency, merge compatibility, or endpoint reference is invalid. |
| `aggregation_grouping_key_missing` | `error` | `source_or_projected_item` | Required grouping key is absent under `missing_grouping_key_behavior=error`. |
| `aggregation_endpoint_missing` | `error` | `mapping_rule` | Aggregated edge endpoint key is missing or endpoint vertex cannot be resolved under `missing_endpoint_behavior=error`. |
| `aggregation_merge_conflict` | `error` | `mapping_rule` | Aggregation merge behavior cannot produce one conforming value. |
| `resource_limit_exceeded` | `fatal` | `projection_input` | Admitted normalized projection input exceeds a closed whole-input resource limit. Pre-admission whole-input limit failures use lifecycle `invalid_projection_request`. |
| `projected_output_limit_exceeded` | `fatal` | `graph_view` | Derived output would exceed a closed projected-output limit. |
| `validation_issue_limit_exceeded` | `fatal` | `projection_input` | Validation issue cap reached. |
| `invalid_retention_policy` | `fatal` | `projection_config` | Retention policy schema or bounds are invalid. |
| `output_schema_violation` | `fatal` | `graph_view` | Derived output would violate this NLSpec. |
| `projection_computation_failed` | `fatal` | `graph_view` | Computation failed before conforming output could be emitted. |

Implementations MUST NOT emit `missing_mapping_rule`. That token is reserved as a non-emitted historical alias.

### 9.6 Validation detail schemas

This registry applies only to admitted-run validation issues. Pre-admission operation errors use lifecycle operation error details and MUST NOT use this registry.

For every validation issue, `details` MUST contain exactly the required keys in this table unless optional keys are listed. Optional keys, when present, participate in neither issue identity nor ordering. Unknown detail keys are forbidden. Any field needed to distinguish two emitted issues MUST be required, not optional.

| Code | Required details | Optional details |
| --- | --- | --- |
| `invalid_input_shape` | `field`, `reason_code` | none |
| `missing_required_input` | `field` | none |
| `explicit_null_not_allowed` | `field` | none |
| `unknown_member` | `field` | none |
| `duplicate_identifier` | `identifier_value`, `collection` | none |
| `invalid_projection_schema` | `field`, `supplied_value` | none |
| `invalid_graph_view_id` | `supplied_value`, `expected_value` | none |
| `invalid_projection_config` | `field`, `reason_code` | none |
| `invalid_field_path` | `field_path`, `scope` | none |
| `invalid_filter` | `field`, `reason_code` | none |
| `invalid_mapping_rule` | `mapping_rule_id`, `reason_code` | none |
| `missing_entity_mapping_rule` | `source_entity_kind` | none |
| `missing_relationship_mapping_rule` | `source_relationship_kind` | none |
| `invalid_metadata_mapping` | `metadata_mapping_id`, `reason_code` | none |
| `invalid_property_definition` | `property_definition_id`, `reason_code` | none |
| `invalid_property_type` | `projected_key`, `expected_type`, `actual_type`, `source_field_path`, `output_object_id` | `aggregation_rule_id`, `canonical_grouping_key_digest`, `contributor_id` |
| `required_property_missing` | `projected_key`, `source_field_path`, `output_object_id` | `aggregation_rule_id`, `canonical_grouping_key_digest` |
| `source_null_for_required_property` | `projected_key`, `source_field_path`, `output_object_id` | `aggregation_rule_id`, `canonical_grouping_key_digest`, `contributor_id` |
| `undeclared_source_kind` | `source_item_id`, `source_kind` | none |
| `source_item_resource_limit_exceeded` | `source_item_id`, `limit_key`, `limit`, `observed` | none |
| `missing_relationship_endpoint` | `source_relationship_id`, `endpoint_field` | none |
| `relationship_endpoint_not_projected` | `source_relationship_id`, `endpoint_field`, `endpoint_source_entity_id` | none |
| `invalid_relationship_direction` | `source_relationship_id`, `supplied_value` | none |
| `invalid_direction_policy` | `mapping_rule_id`, `supplied_value` | none |
| `invalid_reverse_edge_policy` | `mapping_rule_id`, `projected_direction` | none |
| `invalid_aggregation_rule` | `aggregation_rule_id`, `reason_code` | none |
| `aggregation_grouping_key_missing` | `aggregation_rule_id`, `field_path`, `contributor_id` | none |
| `aggregation_endpoint_missing` | `aggregation_rule_id`, `endpoint_side`, `reason_code`, `endpoint_digest`, `field_path` | none |
| `aggregation_merge_conflict` | `aggregation_rule_id`, `canonical_grouping_key_digest`, `projected_key` | none |
| `resource_limit_exceeded` | `limit_key`, `limit`, `observed` | none |
| `projected_output_limit_exceeded` | `limit_key`, `limit`, `observed` | none |
| `validation_issue_limit_exceeded` | `limit` | none |
| `invalid_retention_policy` | `field`, `reason_code` | none |
| `output_schema_violation` | `field`, `reason_code` | none |
| `projection_computation_failed` | `reason_code` | none |

Target kinds MUST use this closed vocabulary: `projection_input`, `projection_config`, `graph_view`, `filter`, `mapping_rule`, `property_definition`, `property`, `source_item`, `source_relationship`, `source_or_projected_item`.

#### 9.6.1 Validation issue construction registry

The implementation MUST construct `target_id` and `field_path` by code according to this table.

| Code family | Target ID derivation | `field_path` |
| --- | --- | --- |
| `invalid_input_shape`, `missing_required_input`, `explicit_null_not_allowed`, `unknown_member`, `duplicate_identifier`, `invalid_projection_schema`, `resource_limit_exceeded`, `validation_issue_limit_exceeded` | literal `projection_input` | Canonical input path when attributable; otherwise `null`. |
| `invalid_graph_view_id`, `projected_output_limit_exceeded`, `output_schema_violation`, `projection_computation_failed` | valid supplied or derived `graph_view_id`; otherwise literal `graph_view` | Canonical output or input path when attributable; otherwise `null`. |
| `invalid_projection_config`, `invalid_field_path`, `invalid_retention_policy` | literal `projection_config` unless a stable child identifier exists | Canonical input path or §4.2 field path, as applicable. |
| `invalid_filter` | Canonical input path to the filter predicate | Canonical input path to the invalid member. |
| `invalid_mapping_rule`, `missing_entity_mapping_rule`, `missing_relationship_mapping_rule`, `invalid_direction_policy`, `invalid_reverse_edge_policy` | `mapping_rule_id` when valid; otherwise canonical input path to the rule | Canonical input path when attributable; otherwise `null`. |
| `invalid_metadata_mapping` | `metadata_mapping_id` when valid; otherwise canonical input path to the mapping | Canonical input path when attributable; otherwise `null`. |
| `invalid_property_definition` | `property_definition_id` when valid; otherwise canonical input path to the definition | Canonical input path when attributable; otherwise `null`. |
| `invalid_property_type`, `required_property_missing`, `source_null_for_required_property` | `<output_object_id>#<projected_key>` when the output object ID is derivable; otherwise canonical input path | The owning `source_field_path`. |
| `undeclared_source_kind`, `source_item_resource_limit_exceeded` | source entity or relationship ID when valid; otherwise canonical input path to the item | Canonical input path when attributable; otherwise `null`. |
| `missing_relationship_endpoint`, `relationship_endpoint_not_projected`, `invalid_relationship_direction` | `source_relationship_id` when valid; otherwise canonical input path to the source relationship | Canonical input path when attributable; otherwise `null`. |
| `invalid_aggregation_rule`, `aggregation_endpoint_missing`, `aggregation_merge_conflict` | `aggregation_rule_id` when valid; otherwise canonical input path to the aggregation rule | Canonical input path or §4.2 field path when attributable; otherwise `null`. |
| `aggregation_grouping_key_missing` | contributor ID from the contributor sort-key family | The missing grouping `field_path`. |

#### 9.6.2 Validation reason-code registries

Reason-code fields are closed. A reason code outside these tables is non-conformant.

| Validation code | Allowed `reason_code` values |
| --- | --- |
| `invalid_input_shape` | `schema_type_mismatch`, `scalar_contract_violation`, `property_value_too_long`, `array_element_invalid`, `array_length_exceeded`, `invalid_label` |
| `invalid_projection_config` | `custom_config_referenced`, `relationship_mapping_source_conflict`, `empty_kind_registry_not_allowed`, `declared_kind_duplicate`, `mapping_rule_duplicate`, `metadata_mapping_duplicate`, `aggregation_rule_duplicate`, `unknown_configuration_member`, `invalid_default_materialization` |
| `invalid_filter` | `invalid_operator`, `value_required`, `value_forbidden`, `invalid_field_scope`, `invalid_value_shape`, `unsupported_logic` |
| `invalid_mapping_rule` | `duplicate_mapping_rule_id`, `duplicate_source_entity_kind_mapping`, `duplicate_source_relationship_kind_mapping`, `declared_source_kind_missing`, `property_key_not_defined`, `property_requiredness_mismatch`, `required_optional_overlap`, `reverse_edge_kind_without_reverse`, `label_invalid` |
| `invalid_metadata_mapping` | `reserved_metadata_key`, `duplicate_after_wildcard_expansion`, `invalid_source_scope`, `invalid_default_value`, `invalid_merge_behavior_type`, `invalid_projected_type`, `required_metadata_missing` |
| `invalid_property_definition` | `duplicate_after_wildcard_expansion`, `invalid_source_scope`, `invalid_default_value`, `invalid_null_policy`, `invalid_merge_behavior_type`, `invalid_projected_type` |
| `invalid_aggregation_rule` | `dependency_on_later_rule`, `aggregation_cycle`, `endpoint_rule_not_vertex_rule`, `endpoint_grouping_key_count_mismatch`, `endpoint_grouping_key_invalid`, `endpoint_field_scope_invalid`, `grouping_key_invalid`, `invalid_endpoint_behavior`, `invalid_edge_direction`, `input_scope_invalid`, `invalid_merge_behavior_type` |
| `aggregation_endpoint_missing` | `endpoint_key_missing`, `endpoint_vertex_not_found` |
| `invalid_retention_policy` | `unknown_member`, `out_of_bounds`, `invalid_type`, `explicit_null_not_allowed` |
| `output_schema_violation` | `id_mismatch`, `reference_missing`, `sort_order_invalid`, `schema_registry_mismatch`, `metadata_shape_invalid`, `closed_schema_violation`, `canonical_serialization_invalid` |
| `projection_computation_failed` | `internal_exception`, `dependency_unavailable`, `timeout`, `resource_exhausted`, `implementation_invariant_failed` |

## 10. Projection lifecycle and retention

Lifecycle state is split into graph-view state and projection-run/result state. A state name belongs to exactly one state machine unless the tables both declare the same token with machine-local meaning.

### 10.0 Lifecycle operation interfaces

Lifecycle operations are abstract interfaces. A conforming implementation MAY expose them through any transport, but the request members, defaults, idempotency, concurrency, state transitions, success envelopes, and errors in this section MUST remain observable.

Lifecycle operation responses MUST use one of these abstract JSON-compatible variants.

| Variant | Required shape |
| --- | --- |
| Success | `{ "status": "ok", "data": <operation-specific result object> }` |
| Error | `{ "status": "error", "error": { "code": <lifecycle_error_code>, "retryable": <boolean>, "details": <closed object> } }` |

A lifecycle error response MUST NOT contain partial success data.

| Operation | Required request members | Optional request members | Success `data` object | Error codes |
| --- | --- | --- | --- | --- |
| `create_projection` | `projection_input` | `idempotency_key` | `accepted_run_summary` | `invalid_projection_request`, `invalid_operation`, `operation_conflict` |
| `refresh_projection` | `graph_view_id`, `projection_input` | `idempotency_key` | `accepted_run_summary` | `invalid_projection_request`, `graph_view_not_found`, `projection_not_available`, `invalid_operation`, `operation_conflict` |
| `invalidate_graph_view` | `graph_view_id`, `reason_code`, `requested_at`, `requested_by` | `idempotency_key` | `invalidation_summary` | `graph_view_not_found`, `invalid_operation`, `operation_conflict` |
| `invalidate_projection_run` | `graph_view_id`, `projection_run_id`, `reason_code`, `requested_at`, `requested_by` | `idempotency_key` | `invalidation_summary` | `projection_run_not_found`, `invalid_operation`, `operation_conflict` |

An `accepted_run_summary` MUST be a JSON object with exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `graph_view_id` | `generated_id` with prefix `gv_` | Yes | No | Derived graph view. |
| `projection_run_id` | `generated_id` with prefix `gpr_` | Yes | No | Admitted run. |
| `state` | String | Yes | No | Always `accepted` in this response, even if computation completes before response emission. |
| `source_snapshot_id` | `identifier` | Yes | No | From normalized input. |
| `projection_version` | `identifier` | Yes | No | From normalized configuration. |
| `accepted_at` | `timestamp` | Yes | No | Assigned by §10.8 timestamp rules. |
| `idempotency_expires_at` | `timestamp` | Yes | Yes | Non-null when an `idempotency_key` was supplied; otherwise `null`. |

If an implementation computes synchronously and the run reaches `available` or `failed` before the operation response is emitted, the operation response MUST still return `state="accepted"` and the accepted-run summary. The terminal state is observable through `get_projection_run()` and graph-reading queries.

Lifecycle operation errors MUST use this registry.

| Error code | Retryable | Required details | Applies to |
| --- | ---: | --- | --- |
| `invalid_projection_request` | No | `operation`, `reason_code`, `field`, `validation_code` | `create_projection`, `refresh_projection` pre-admission failures |
| `invalid_operation` | No | `operation`, `reason_code` | All lifecycle operations |
| `operation_conflict` | Yes only when `reason_code=run_already_active`; otherwise No | `operation`, `reason_code`, optional `active_projection_run_id` | All lifecycle operations |
| `graph_view_not_found` | No | `operation`, `graph_view_id` | Refresh and graph-view invalidation |
| `projection_not_available` | Yes only for active `creating`, `refreshing`, `accepted`, or `computing`; otherwise No | `operation`, `graph_view_id`, `state` | Refresh |
| `projection_run_not_found` | No | `operation`, `graph_view_id`, `projection_run_id` | Run invalidation |

For `invalid_projection_request`, `field` is a canonical input path when one request member is attributable; otherwise it is JSON null. `validation_code` is a §9.5 code when the rejected condition corresponds to an admitted-run validation family; otherwise it is JSON null.

Reason codes for lifecycle operation errors are closed.

| Reason code | Applies to | Meaning |
| --- | --- | --- |
| `invalid_utf8` | `invalid_projection_request` | Projection input bytes are not valid UTF-8. |
| `invalid_json_syntax` | `invalid_projection_request` | Projection input bytes are not valid JSON. |
| `duplicate_object_member` | `invalid_projection_request` | Any object contains a duplicate member name before schema validation. |
| `top_level_not_object` | `invalid_projection_request` | Decoded projection input is not a JSON object. |
| `missing_required_member` | `invalid_projection_request` | Required top-level member needed for admission is omitted. |
| `explicit_null_not_allowed` | `invalid_projection_request`, `invalid_operation` | A non-nullable admission or operation member is explicit JSON null. |
| `unknown_member` | `invalid_projection_request` | Unknown top-level admission member is present. |
| `invalid_projection_schema` | `invalid_projection_request` | `projection_schema_id` is not exactly `graph_projection.v1`. |
| `invalid_graph_view_id` | `invalid_projection_request` | Supplied `graph_view_id` is malformed or not the derived value. |
| `default_materialization_failed` | `invalid_projection_request` | Defaults cannot be materialized deterministically. |
| `normalization_failed` | `invalid_projection_request` | Input cannot be normalized deterministically. |
| `whole_input_limit_exceeded` | `invalid_projection_request` | A whole-input limit prevents admission. |
| `graph_view_already_exists` | `create_projection` | A graph view already has a selected available or refreshing run; caller must use refresh. |
| `graph_view_not_created` | `refresh_projection` | No graph view exists for refresh. |
| `no_consumable_prior_run` | `refresh_projection` | Refresh requires an available prior run unless the graph view is invalidated. |
| `run_already_active` | create or refresh | Graph view already has an `accepted` or `computing` run. |
| `idempotency_key_conflict` | any operation | Same operation idempotency key was reused with different normalized request bytes. |
| `invalid_idempotency_key` | any operation | Supplied `idempotency_key` violates §4.1. |
| `invalid_invalidation_target` | invalidation operations | Target state cannot be invalidated. |
| `invalid_reason_code` | invalidation operations | Reason code is outside §10.7. |

`create_projection` is allowed only when graph-view state is `not_created`, `failed`, or absent. It creates a new run in `accepted`. `refresh_projection` is allowed only when graph-view state is `available` or `invalidated`. It creates a new run in `accepted`. Retrying after an initial failed create uses `create_projection`, not `refresh_projection`.

At most one `accepted` or `computing` run may exist per graph view at a time. A concurrent create or refresh with a different idempotency key while a run is active MUST return `operation_conflict` with `reason_code=run_already_active` and `active_projection_run_id` set to the active run.

Normalized operation request bytes MUST be canonical JSON of an object with exact members `operation`, `graph_view_id` when applicable, `projection_input` after default materialization when applicable, `projection_run_id` when applicable, `reason_code` when applicable, `requested_at` when applicable, and `requested_by` when applicable. Unknown operation request members are invalid.

`idempotency_key` MUST satisfy §4.1. Idempotency behavior is closed by this table.

| Operation | Idempotency scope | Record creation point | Replay comparison bytes | Expiry |
| --- | --- | --- | --- | --- |
| `create_projection` | `(operation, derived graph_view_id, idempotency_key)` | After admission succeeds | Normalized operation request bytes | `accepted_at + 86400 seconds` |
| `refresh_projection` | `(operation, graph_view_id, idempotency_key)` | After admission succeeds | Normalized operation request bytes | `accepted_at + 86400 seconds` |
| `invalidate_graph_view` | `(operation, graph_view_id, idempotency_key)` | After target validation succeeds | Normalized operation request bytes | `invalidated_at + 86400 seconds` |
| `invalidate_projection_run` | `(operation, graph_view_id, projection_run_id, idempotency_key)` | After target validation succeeds | Normalized operation request bytes | `invalidated_at + 86400 seconds` |

| Input state | Required behavior |
| --- | --- |
| Omitted `idempotency_key` | Operation is not replay-protected and no idempotency record is created. |
| Explicit JSON null | Reject with lifecycle `invalid_operation`, `reason_code=explicit_null_not_allowed`. |
| Empty string | Reject with lifecycle `invalid_operation`, `reason_code=invalid_idempotency_key`. |
| Invalid scalar | Reject with lifecycle `invalid_operation`, `reason_code=invalid_idempotency_key`. |
| Same key, same normalized request, record unexpired | Return original success payload. |
| Same key, different normalized request, record unexpired | Return `operation_conflict`, `reason_code=idempotency_key_conflict`. |
| Same key after expiry | Treat as a new operation request. |

### 10.0.1 Admission, validation, and operation outcome matrix

Validation errors MUST NOT be returned as lifecycle operation errors after run admission. They are observable only through retained run inspection and graph-view validation summary where applicable.

| Condition after `create_projection` or `refresh_projection` is called | Operation response | Retained run? | `get_projection_run()` result |
| --- | --- | ---: | --- |
| Pre-admission request error | Lifecycle error | No | `projection_run_not_found` for any caller-guessed run ID |
| Admission succeeds and run remains queued | `accepted_run_summary` | Yes | `state=accepted`, `validation_summary=null` |
| Admission succeeds and computation starts | `accepted_run_summary` | Yes | `state=computing`, `validation_summary=null` |
| Admission succeeds and validation has no fatal issues | `accepted_run_summary` | Yes | `state=available`, validation summary non-null |
| Admission succeeds and validation has fatal issues | `accepted_run_summary` | Yes | `state=failed`, validation summary non-null |
| Admission succeeds and computation fails before ordinary validation summary | `accepted_run_summary` | Yes | `state=failed`, summary with exactly one `projection_computation_failed` issue |

### 10.1 Graph-view states

| State | Meaning |
| --- | --- |
| `not_created` | No run exists for the graph view. |
| `creating` | Initial run is accepted or computing and no available result exists. |
| `available` | Latest available result exists. |
| `refreshing` | Replacement run is accepted or computing while latest available result remains consumable. |
| `failed` | Latest initial create attempt failed and no available result exists. |
| `invalidated` | Latest available result has been invalidated and no newer available result exists. |

`list_graph_views().graph_views[].state` MUST return graph-view state only.

### 10.2 Projection-run/result states

| State | Meaning |
| --- | --- |
| `accepted` | Run ID assigned; computation not yet started or not yet observable as computing. |
| `computing` | Run is computing. |
| `available` | Run produced a consumable graph result. |
| `failed` | Run completed without a consumable graph result. |
| `replaced` | Run was available but superseded by a newer available run. |
| `invalidated` | Run exists but is no longer consumable. |
| `expired` | Run was retained previously but is no longer addressable. |

### 10.3 Allowed transitions

| Machine | From | Event | To | Required effect |
| --- | --- | --- | --- | --- |
| Graph view | `not_created` | create accepted | `creating` | Create initial run in `accepted`. |
| Graph view | `failed` | create accepted | `creating` | Create another initial run in `accepted`. |
| Graph view | `creating` | run available | `available` | Latest available run becomes selected. |
| Graph view | `creating` | run failed | `failed` | Failed run summary is retained under §10.6. |
| Graph view | `available` | refresh accepted | `refreshing` | Prior available result remains consumable. |
| Graph view | `refreshing` | replacement run available | `available` | Old run becomes `replaced`; new run becomes latest `available`. |
| Graph view | `refreshing` | replacement run failed | `available` | Failed run retained; prior available result remains selected. |
| Graph view | `available` | graph-view invalidation accepted | `invalidated` | Invalidation cascade in §10.7 applies. |
| Graph view | `invalidated` | refresh accepted | `refreshing` | New run may produce replacement. |
| Run/result | `accepted` | computation starts | `computing` | No graph result is visible yet. |
| Run/result | `computing` | no fatal issues | `available` | Result is published atomically. |
| Run/result | `computing` | fatal issues or computation failure | `failed` | Validation summary or failure summary is retained. |
| Run/result | `available` | newer run published | `replaced` | Exact retention policy applies. |
| Run/result | `available` or `replaced` | invalidation accepted | `invalidated` | Result is not consumable by graph read queries. |
| Run/result | `replaced`, `invalidated`, or retained `failed` | retention expires | `expired` | Run-specific reads return `projection_run_not_found`. |

### 10.4 Creation, refresh, and replacement rules

Publishing a projection result MUST be atomic from the perspective of consumers. A consumer MUST NOT observe a partially replaced graph view.

Initial creation publishes a graph view only when the initial run reaches run/result state `available`. A failed initial creation leaves graph-view state `failed` and MUST preserve the failed run summary for `get_projection_run()` under §10.6.

A refresh that succeeds MUST set the previous latest available run to `replaced`, set the new run to `available`, and set graph-view state to `available`. A refresh that fails MUST leave the previous latest available run queryable and MUST retain the failed run summary under §10.6.

Graph-view metadata for a successful replacement MUST set `previous_projection_run_id` to the previous latest available run ID. Initial creation MUST set it to `null`.

### 10.5 Failure behavior

Failed-run retention begins only after run admission. A pre-admission rejection MUST NOT create a failed projection run.

A failed projection run MUST retain:

- `graph_view_id`,
- `projection_run_id`,
- `source_snapshot_id`,
- `projection_version`,
- run/result `state=failed`,
- `validation_summary`,
- `failure_reason`,
- `started_at`,
- `completed_at` when known.

A failed run MUST NOT be returned by `get_graph_view()` as a consumable graph. Failed runs are inspectable only through `get_projection_run()` while retained.

If computation fails before ordinary validation can produce a summary, the run's `validation_summary` MUST have `status=failed`, `fatal_count=1`, and exactly one `projection_computation_failed` issue. That issue MUST be constructed under §7.7 because the run is admitted and has fixed identity.

### 10.6 Retention policy fields and exact retention

`retention_policy` MUST be an object with exactly the following members. Unknown members are invalid.

| Member | Type | Required | Nullable | Default | Bounds |
| --- | --- | ---: | ---: | --- | --- |
| `retain_replaced_results` | Boolean | No | No | `true` | Boolean only. |
| `retention_count` | Integer | No | No | `5` | `0` through `100`. |
| `retention_duration_seconds` | Integer | No | No | `2592000` | `0` through `31536000`. |
| `retain_failed_results` | Boolean | No | No | `true` | Boolean only. |
| `failed_retention_count` | Integer | No | No | `20` | `0` through `100`. |
| `failed_retention_duration_seconds` | Integer | No | No | `2592000` | `0` through `31536000`. |

The latest `available` run is always retained while graph-view state is `available` or `refreshing`.

If `retain_replaced_results=false`, a replaced run becomes `expired` immediately after replacement publication. If `retain_replaced_results=true`, a replaced run is retained only when both conditions are true:

1. it is within the most recent `retention_count` replaced runs sorted by `replaced_at DESC`, then `projection_run_id ASC`;
2. `query_received_at < replaced_at + retention_duration_seconds`.

The latest failed initial-create run is always retained while graph-view state is `failed`. Other failed runs are retained only when `retain_failed_results=true` and both conditions are true:

1. the run is within the most recent `failed_retention_count` failed runs sorted by `completed_at DESC`, then `projection_run_id ASC`;
2. `query_received_at < completed_at + failed_retention_duration_seconds`.

If `retain_failed_results=false`, a failed run becomes expired immediately unless it is the latest failed initial-create run for a graph view in state `failed`.

Runs outside the applicable bound are `expired` for all query behavior. Expired runs return `projection_run_not_found`. Retention MUST be evaluated before every run-specific read, immediately after successful replacement publication, and immediately after a failed terminal transition. Implementations MUST NOT retain additional query-addressable replaced or failed runs beyond this exact policy.

Count-bound expiry is event-driven. It MUST be evaluated immediately after successful replacement publication and immediately after a failed terminal transition. It MUST also be evaluated before every run-specific read.

A non-null `retention_expires_at` returned by `get_projection_run()` is the time-bound expiry timestamp under the applicable duration rule at response construction time. It is not a lease. A retained run can become expired before that timestamp only when a later run changes the applicable count-bound rank under this section.

### 10.7 Invalidation contract

Invalidation reason codes are closed.

| Reason code | Meaning |
| --- | --- |
| `operator_requested` | A permitted operator or system process explicitly withdrew the run or graph view. |
| `source_snapshot_withdrawn` | The source snapshot is no longer valid for consumer use. |
| `projection_config_retired` | The projection configuration was retired. |
| `security_withdrawal` | The result was withdrawn for safety or disclosure reasons. |
| `schema_version_retired` | The projection schema version is no longer accepted for consumption. |

`invalidate_graph_view` MUST invalidate every retained `available` and `replaced` run for that graph view. Failed run summaries remain inspectable after graph-view invalidation until failed-run retention expires. `invalidate_projection_run` MUST invalidate exactly one retained `available` or `replaced` run.

If `invalidate_projection_run` targets the latest selected available run, graph-view state becomes `invalidated`. If it targets a retained replaced run, graph-view state does not change. Invalidated runs are never consumable by graph-reading queries. Invalidated runs remain inspectable by `get_projection_run()` until retention expires.

The invalidation object stored on metadata and run inspection MUST use the shape in §5.5.1. `metadata.invalidation` is `null` for non-invalidated selected runs.

### 10.7.1 Invalidation summary

An `invalidation_summary` MUST be a JSON object with exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `graph_view_id` | `generated_id` with prefix `gv_` | Yes | No | Target graph view. |
| `target_scope` | String | Yes | No | `graph_view` or `projection_run`. |
| `target_projection_run_id` | `generated_id` with prefix `gpr_` | Yes | Yes | Non-null only for run-specific invalidation. |
| `invalidated_projection_run_ids[]` | Array of `gpr_` IDs | Yes | No | Runs newly invalidated by this operation, sorted by `projection_run_id` ascending. |
| `graph_view_state_after` | String | Yes | No | Graph-view state after invalidation transition. |
| `invalidated_at` | `timestamp` | Yes | No | Assigned by §10.8 timestamp rules. |
| `reason_code` | String | Yes | No | One code from §10.7. |
| `requested_by` | `identifier` | Yes | No | Copied from request after validation. |
| `idempotency_expires_at` | `timestamp` | Yes | Yes | Non-null when an `idempotency_key` was supplied; otherwise `null`. |

Exact idempotent replay within the idempotency retention window MUST return the original `invalidation_summary`. A new non-idempotent invalidation request against an already invalidated target MUST return `invalid_operation` with `reason_code=invalid_invalidation_target`. Invalidation MUST NOT change failed-run retention.

`invalidated_projection_run_ids[]` MUST contain only runs whose state changed because of this operation. It MUST NOT include already invalidated runs except on exact idempotent replay, where the original response is returned.

### 10.8 Lifecycle timestamp ownership

Lifecycle timestamps are implementation-owned server lifecycle values. Caller-supplied `requested_at` is audit input. It MUST NOT determine lifecycle ordering, retention, cursor ordering, or `updated_at`.

| Timestamp | Assigned by | Assignment instant | Precision | Monotonic rule | Tie-breaker |
| --- | --- | --- | --- | --- | --- |
| `accepted_at` | Implementation projection lifecycle clock | Durable run admission commit | Exactly 6 fractional digits | Non-decreasing per graph view | If clock value is not greater than prior lifecycle timestamp for the same graph view, increment by 1 microsecond. |
| `started_at` | Implementation projection lifecycle clock | Run enters `computing` | Exactly 6 fractional digits | `started_at >= accepted_at` | Same 1 microsecond increment rule. |
| `generated_at` | Implementation projection lifecycle clock | Consumable graph result is atomically published | Exactly 6 fractional digits | `generated_at >= started_at` | Same 1 microsecond increment rule. |
| `completed_at` | Implementation projection lifecycle clock | Run enters terminal `available` or `failed` | Exactly 6 fractional digits | `completed_at >= started_at` when `started_at` is non-null | Same 1 microsecond increment rule. |
| `replaced_at` | Implementation projection lifecycle clock | New available run replaces prior available run | Exactly 6 fractional digits | `replaced_at >= generated_at` of replacement run | Same 1 microsecond increment rule. |
| `invalidated_at` | Implementation projection lifecycle clock | Invalidation transition is committed | Exactly 6 fractional digits | Non-decreasing per graph view | Same 1 microsecond increment rule. |
| `updated_at` | Implementation projection lifecycle clock | Graph-view state or selected/latest run summary changes | Exactly 6 fractional digits | Non-decreasing per graph view | Same 1 microsecond increment rule. |

## 11. Consumer query contract

Query shapes are contract-level interfaces. A conforming implementation MAY expose them through any transport that preserves request shape, defaults, errors, result shape, retryability, and ordering.

### 11.1 Common query response and error behavior

Every query response MUST use one of these abstract variants.

| Variant | Shape |
| --- | --- |
| Success | `{ "status": "ok", "data": <query-specific result object> }` |
| Error | `{ "status": "error", "error": { "code": <query_error_code>, "retryable": <boolean>, "details": <closed object> } }` |

Query-specific sections define the `data` value for success responses. Error responses MUST NOT include partial query data. Unknown request members MUST return `invalid_argument`. Explicit JSON null for a non-nullable request member MUST return `invalid_argument`. A transport may wrap these objects only when the same fields, values, and semantics remain observable.

Query errors MUST use the following closed codes and detail shapes.

| Error | Retryable | Required details |
| --- | ---: | --- |
| `invalid_argument` | No | `field`, `reason_code` |
| `graph_view_not_found` | No | `graph_view_id` |
| `projection_not_available` | Yes only when `state` is `creating`, `refreshing`, `accepted`, or `computing`; otherwise No | `graph_view_id`, `state` |
| `projection_run_not_found` | No | `graph_view_id`, `projection_run_id` |
| `projection_run_failed` | No | `graph_view_id`, `projection_run_id` |
| `projection_run_invalidated` | No | `graph_view_id`, `projection_run_id`, `invalidation.reason_code`, `invalidation.invalidated_at` |
| `vertex_not_found` | No | `graph_view_id`, `projection_run_id`, `vertex_id` |
| `edge_not_found` | No | `graph_view_id`, `projection_run_id`, `edge_id` |
| `cursor_invalid` | No | `reason_code` |

`invalid_argument.reason_code` MUST be one of: `missing_required_parameter`, `explicit_null_not_allowed`, `invalid_type`, `out_of_bounds`, `duplicate_id`, `unknown_parameter`, `unsupported_parameter`, or `cursor_token_too_long`.

`cursor_invalid.reason_code` MUST be one of: `malformed`, `expired`, `wrong_query_shape`, or `cursor_token_too_long`.

Expired runs MUST use `projection_run_not_found`, not a separate public error. Alternate traversal order MUST use `invalid_argument` because no alternate order is defined in this revision.

### 11.2 Projection result selection

Every graph-reading query that accepts optional `projection_run_id` MUST select the projection result as follows.

| Request state | Selected result or error |
| --- | --- |
| `projection_run_id` omitted and latest graph-view state is `available` or `refreshing` | Latest available result for graph view. |
| `projection_run_id` omitted and graph-view state is `creating` or `refreshing` with no prior available result | `projection_not_available`, retryable `true`. |
| `projection_run_id` omitted and graph-view state is `not_created`, `failed`, or `invalidated` with no available selected run | `projection_not_available`, retryable `false`. |
| `projection_run_id` supplied and run state is `accepted` | `projection_not_available`, retryable `true`. |
| `projection_run_id` supplied and run state is `computing` | `projection_not_available`, retryable `true`. |
| `projection_run_id` supplied and run state is `available` or retained `replaced` | That run's retained graph result. |
| `projection_run_id` supplied and run state is `failed` | `projection_run_failed` for graph-reading queries. |
| `projection_run_id` supplied and run state is `invalidated` | `projection_run_invalidated` for graph-reading queries. |
| `projection_run_id` supplied and run state is `expired` or unknown | `projection_run_not_found`. |

For `projection_not_available`, `details.state` MUST be the graph-view state when no run is supplied and the run state when a specific nonterminal run is supplied.

### 11.3 `get_graph_view(graph_view_id, projection_run_id?)`

Request members:

| Parameter | Required | Nullable | Default | Validation |
| --- | ---: | ---: | --- | --- |
| `graph_view_id` | Yes | No | None | Valid `gv_` generated ID. |
| `projection_run_id` | No | No | Latest available result | Valid `gpr_` generated ID belonging to the graph view. |

Success `data` is one graph view output object from §5.1.

### 11.4 `get_vertex(graph_view_id, vertex_id, projection_run_id?)`

Request members:

| Parameter | Required | Nullable | Default | Validation |
| --- | ---: | ---: | --- | --- |
| `graph_view_id` | Yes | No | None | Valid `gv_` generated ID. |
| `vertex_id` | Yes | No | None | Valid `vx_` generated ID. |
| `projection_run_id` | No | No | Latest available result | Valid `gpr_` generated ID belonging to the graph view. |

Success `data` is one projected vertex object. If the selected projection result exists but `vertex_id` is absent, return `vertex_not_found`.

### 11.5 `get_edge(graph_view_id, edge_id, projection_run_id?)`

Request members:

| Parameter | Required | Nullable | Default | Validation |
| --- | ---: | ---: | --- | --- |
| `graph_view_id` | Yes | No | None | Valid `gv_` generated ID. |
| `edge_id` | Yes | No | None | Valid `ed_` generated ID. |
| `projection_run_id` | No | No | Latest available result | Valid `gpr_` generated ID belonging to the graph view. |

Success `data` is one projected edge object. If the selected projection result exists but `edge_id` is absent, return `edge_not_found`.

### 11.6 `traverse(...)`

Request members:

| Parameter | Required | Nullable | Default | Bounds and validation |
| --- | ---: | ---: | --- | --- |
| `graph_view_id` | Yes | No | None | Valid `gv_` generated ID. |
| `seed_vertex_ids[]` | Yes | No | None | Array of 0 to 1024 distinct `vx_` generated IDs. Duplicate seeds are `invalid_argument`. |
| `direction` | No | No | `outbound` | `outbound`, `inbound`, or `any`. |
| `max_depth` | No | No | `1` | Integer `0` through `16`. |
| `edge_kinds[]` | No | No | All edge kinds | Array of 0 to 1024 distinct `kind` values. Empty array means no edges are traversable. |
| `vertex_kinds[]` | No | No | All vertex kinds | Array of 0 to 1024 distinct `kind` values. Empty array means only existing seed vertices may be returned. |
| `projection_run_id` | No | No | Latest available result | Valid `gpr_` generated ID belonging to the graph view. |

Traversal success `data` MUST be a JSON object with exactly these members: `graph_view_id`, `projection_run_id`, `seed_vertex_ids[]`, `omitted_seed_vertex_ids[]`, `vertices[]`, `edges[]`, and `metadata`.

`vertices[]` and `edges[]` MUST use the same object shapes and ordering as the selected graph view. `omitted_seed_vertex_ids[]` MUST contain unknown seed IDs sorted by input order after duplicate validation. In this revision, traversal `metadata` MUST be exactly `{}`. Implementations MUST NOT add timing, counts, internal plan details, cache details, or storage-engine diagnostics to traversal metadata.

#### 11.6.1 Traversal direction table

For a current vertex `v`, an emitted edge `e`, and requested traversal `direction`, neighbor selection MUST follow this table.

| Edge direction | Requested `outbound` | Requested `inbound` | Requested `any` |
| --- | --- | --- | --- |
| `directed` | If `e.src_vertex_id=v`, neighbor is `e.dst_vertex_id`; otherwise none. | If `e.dst_vertex_id=v`, neighbor is `e.src_vertex_id`; otherwise none. | If either endpoint equals `v`, neighbor is the other endpoint; for a self-loop, neighbor is `v`. |
| `undirected` | If either endpoint equals `v`, neighbor is the other endpoint; for a self-loop, neighbor is `v`. | Same as `outbound`. | Same as `outbound`. |
| `bidirectional` | If either endpoint equals `v`, neighbor is the other endpoint; for a self-loop, neighbor is `v`. | Same as `outbound`. | Same as `outbound`. |

#### 11.6.2 Breadth-first traversal algorithm

Traversal MUST use this deterministic breadth-first algorithm.

```text
1. Select projection result by graph_view_id and optional projection_run_id.
2. Validate parameters and reject duplicate seed_vertex_ids[].
3. Resolve existing seed vertices.
4. Sort existing seed vertices by graph output vertex order.
5. Initialize visited_vertices with existing seed vertices.
6. Initialize frontier with existing seed vertices at depth 0.
7. Initialize included_edges as empty.
8. If max_depth = 0, return seed vertices and no edges.
9. For each depth d from 0 to max_depth - 1:
   a. Initialize next_frontier as empty.
   b. Iterate frontier vertices in graph output vertex order.
   c. Iterate all emitted edges in graph output edge order.
   d. Skip edge if edge_kind is not allowed by edge_kinds[].
   e. Apply the traversal direction table to determine neighbor.
   f. Skip if no neighbor exists.
   g. If neighbor is not already visited:
      i. Add neighbor only if it passes vertex_kinds[].
      ii. Add accepted neighbor to visited_vertices and next_frontier.
   h. Include the traversed edge only if both endpoints are in visited_vertices after this step.
10. Replace frontier with next_frontier and continue.
11. Emit each vertex and edge at most once.
12. Return vertices[] and edges[] sorted by graph output order.
```

| Case | Required behavior |
| --- | --- |
| Empty `seed_vertex_ids[]` | Return empty seeds, vertices, edges, and omitted seeds. |
| Unknown seed | Omit from traversal and list in `omitted_seed_vertex_ids[]`. |
| Duplicate seed | `invalid_argument`. |
| Self-loop | Traversable if direction table permits; edge included once if endpoint vertex is included. |
| Multi-edge | Each eligible emitted edge is considered independently and may be included once. |
| `vertex_kinds[]=[]` | Only existing seed vertices may be returned; no new neighbor is added. |
| `edge_kinds[]=[]` | No edges are traversable. |
| Seed not in `vertex_kinds[]` | Existing seed remains included. Vertex filter applies only to newly discovered neighbors. |

### 11.7 `list_graph_views(limit?, cursor_token?)`

Request members:

| Parameter | Required | Nullable | Default | Bounds and validation |
| --- | ---: | ---: | --- | --- |
| `limit` | No | No | `100` | Integer `1` through `1000`. |
| `cursor_token` | No | No | First page | `cursor_token` string with maximum length `4096`. |

Pagination is live keyset pagination over the currently visible graph-view set. No snapshot stability is promised. Sort order is `graph_view_id ASC`. The first page starts after no key.

A cursor token MUST encode:

- operation name `list_graph_views`,
- `after_graph_view_id`,
- `issued_at`,
- query-shape digest excluding `limit`,
- implementation authentication or visibility scope digest if the implementation has caller-specific visibility below this NLSpec boundary.

A cursor token expires 15 minutes after `issued_at`. `limit` MAY change between pages. If `after_graph_view_id` no longer exists, the next page starts at the first graph view with `graph_view_id > after_graph_view_id`. Graph views created after the first page may appear on later pages if their `graph_view_id` is greater than the cursor's `after_graph_view_id`. Graph views with `graph_view_id <= after_graph_view_id` MUST NOT appear on later pages from that cursor chain.

Success `data` members:

| Member | Rule |
| --- | --- |
| `graph_views[]` | Page of `graph_view_summary` objects sorted by `graph_view_id`. |
| `next_cursor_token` | String for the next page, or JSON null when no next page exists. |

A `graph_view_summary` MUST be a JSON object with exactly these members in this order.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `graph_view_id` | `generated_id` with prefix `gv_` | Yes | No | Graph view identity. |
| `graph_view_key` | `identifier` | Yes | No | From normalized config. |
| `state` | String | Yes | No | One graph-view state except `not_created`, which is not listed. |
| `latest_projection_run_id` | `generated_id` with prefix `gpr_` | Yes | No | Most recently admitted run for the graph view, whether active or terminal. |
| `latest_source_snapshot_id` | `identifier` | Yes | No | Source snapshot for `latest_projection_run_id`. |
| `projection_version` | `identifier` | Yes | No | Projection version for `latest_projection_run_id`. |
| `updated_at` | `timestamp` | Yes | No | Assigned by §10.8. |
| `validation_status` | String | Yes | No | One value from the state matrix below. |

State-specific summary behavior is closed by this matrix.

| Graph-view state | `latest_projection_run_id` | `latest_source_snapshot_id` | `projection_version` | `updated_at` | `validation_status` |
| --- | --- | --- | --- | --- | --- |
| `creating` | Active initial run | From active initial run | From active initial run | Last transition to `creating` or active-run state update | `pending` |
| `available` | Latest available run | From latest available run | From latest available run | Publication timestamp of latest available run or later metadata update | Selected run validation summary `status` |
| `refreshing` | Active refresh run | From active refresh run | From active refresh run | Last transition to `refreshing` or active-run state update | `pending` |
| `failed` | Latest failed initial-create run | From latest failed run | From latest failed run | Failed terminal transition timestamp | `failed` |
| `invalidated` | Latest invalidated selected run | From invalidated run | From invalidated run | Invalidation timestamp | Validation summary `status` from invalidated run |

`not_created` graph views MUST NOT appear in `list_graph_views()`.

`next_cursor_token` is `null` only when no graph view with `graph_view_id > last_returned_graph_view_id` exists at response construction time. Invalid, expired, oversized, malformed, or wrong-query-shape cursor tokens return `cursor_invalid`.

### 11.8 `get_projection_run(graph_view_id, projection_run_id)`

`get_projection_run()` is the required accepted-run, computing-run, failed-run, invalidated-run, and retained-run inspection query.

Request members:

| Parameter | Required | Nullable | Default | Validation |
| --- | ---: | ---: | --- | --- |
| `graph_view_id` | Yes | No | None | Valid `gv_` generated ID. |
| `projection_run_id` | Yes | No | None | Valid `gpr_` generated ID belonging to the graph view. |

Success `data` MUST be a JSON object with exactly these members in this order.

| Field | Required | Nullable | State-specific rule |
| --- | ---: | ---: | --- |
| `graph_view_id` | Yes | No | Selected graph view. |
| `projection_run_id` | Yes | No | Selected run. |
| `source_snapshot_id` | Yes | No | Source snapshot for the run. |
| `projection_version` | Yes | No | Projection version for the run. |
| `state` | Yes | No | One run/result state from §10.2 except `expired`, which is not returned. |
| `started_at` | Yes | Yes | `null` in `accepted`; timestamp in all later states. |
| `completed_at` | Yes | Yes | `null` in `accepted` and `computing`; timestamp in terminal states. |
| `validation_summary` | Yes | Yes | `null` in `accepted` and `computing`; non-null in `available`, `replaced`, `invalidated`, and `failed`. |
| `failure_reason` | Yes | Yes | Non-empty string only in `failed`; otherwise `null`. |
| `has_consumable_graph_view` | Yes | No | `true` only for `available` and retained `replaced`; `false` otherwise. |
| `invalidation` | Yes | Yes | Non-null only in `invalidated`; shape from §5.5.1. |
| `retention_expires_at` | Yes | Yes | Time-bound expiry timestamp under the applicable duration rule at response construction time, or `null` when no duration-bound expiry applies. |

`retention_expires_at` is `null` for the latest available run, the always-retained latest failed initial run, and any run that has no duration-bound expiry. A non-null `retention_expires_at` is not a lease. A retained run can become expired before that timestamp only when a later run changes the applicable count-bound rank under §10.6.

Expired or unknown runs return `projection_run_not_found`. Accepted and computing runs are inspectable through this query but are not graph-readable.

## 12. Intentional implementation latitude

### 12.1 Allowed internal variance

Implementations MAY vary these mechanisms without affecting conformance:

- programming language;
- storage engine and physical schema;
- indexing strategy;
- cache strategy;
- graph engine or absence of a graph engine;
- synchronous versus asynchronous execution;
- scheduling and worker topology;
- transport protocol;
- internal object representation;
- diagnostic log text;
- non-canonical private operator diagnostics that are not returned by consumer query shapes.

### 12.2 Canonical output comparison inputs and run-specific exclusions

A conformance fixture MUST declare whether it is `run_specific` or `run_independent`.

For `run_independent` canonical graph-result comparison, the comparison input MUST exclude only these run-specific fields:

| Field | Reason |
| --- | --- |
| `projection_run_id` | Deliberately unique per accepted run. |
| `projection_run_nonce` | Deliberately run-specific identity input. |
| `requested_at` | Request audit input, not graph derivation. |
| `requested_by` | Request audit input, not graph derivation. |
| `generated_at` | Deliberately records run publication time. |
| `metadata.previous_projection_run_id` | Depends on run history. |
| lifecycle timestamps | State-machine timing, not graph derivation. |
| retention timestamps | Retention availability, not graph derivation. |
| `validation_summary.issues[].issue_id` | Includes `projection_run_id`; may be excluded only by run-independent validation fixtures. |

Validation-summary comparison MUST include `issue_id` unless the fixture explicitly declares `run_independent_validation=true`. Graph object IDs, sort keys, property values, mapped metadata excluding lifecycle/run fields, schema registry, source refs, `projection_config_digest`, and `projection_source_digest` are always included.

For `run_specific` comparison, no fields are excluded unless the fixture itself scopes the comparison to a declared subobject. All other canonical graph output bytes MUST match for the same normalized non-run-specific inputs.

### 12.3 Forbidden implementation leakage

Implementations MUST NOT expose any of the following in canonical graph output or query behavior:

- database table names;
- temporary directory names;
- storage-engine row order;
- graph-engine internal IDs;
- map iteration order;
- locale-dependent sorting;
- private metadata keys outside `mapped_metadata`;
- unregistered validation codes;
- undocumented query errors;
- private `custom_config` semantics.

## 13. Definition of Done

A Graph Projection implementation is conformant only when every criterion in this table passes.

| ID | Criterion | Pass condition |
| --- | --- | --- |
| `GP-AC-001` | Projection schema version is closed. | `projection_schema_id` is required, accepts only `graph_projection.v1`, and omission is not defaulted. |
| `GP-AC-002` | Top-level input schema is closed. | All top-level members have requiredness, nullability, defaults, validation, and unknown-member behavior. |
| `GP-AC-003` | Identity participation is complete. | Every top-level input family appears in the identity participation matrix. |
| `GP-AC-004` | Graph-view ID derivation is enforced. | Supplied `graph_view_id` must equal the canonical derived value and is never silently replaced. |
| `GP-AC-005` | Field-path grammar is closed. | Every accepted field path matches the grammar, and property keys cannot make dotted paths ambiguous. |
| `GP-AC-006` | Resource limits are closed. | Every externally supplied unbounded collection or scalar family has a limit and overflow behavior. |
| `GP-AC-007` | Mapping schemas are closed. | Entity, relationship, metadata, and aggregation rules each have closed schemas with unknown-member behavior. |
| `GP-AC-008` | Custom config cannot affect interoperability. | `custom_config` cannot affect output, validation, identity, ordering, lifecycle, traversal, retention, or query behavior. |
| `GP-AC-009` | Filter truth table is complete. | Missing, null, scalar, and array field states have defined behavior for every operator. |
| `GP-AC-010` | Property type compatibility is complete. | Every projected type has source compatibility, default compatibility, normalization, and invalid behavior. |
| `GP-AC-011` | Omission and explicit-null behavior is closed. | Every optional and nullable boundary has distinct omitted and explicit-null semantics. |
| `GP-AC-012` | Direction policy is exhaustive. | Every `direction_policy × source direction` combination has one projected endpoint and direction result. |
| `GP-AC-013` | Reverse-edge behavior is closed. | Reverse edge kind, direction, endpoints, labels, properties, metadata, source ref, static validity, and identity are defined. |
| `GP-AC-014` | Aggregation is executable. | Aggregation schema, grouping, missing-key behavior, merge behavior, endpoint grouping, source refs, and empty groups are deterministic. |
| `GP-AC-015` | Metadata schemas are closed. | Graph, vertex, and edge metadata have required members and forbid private canonical-output keys. |
| `GP-AC-016` | Schema registry is exposed. | Graph-view output includes a parseable schema registry for consumer reliance. |
| `GP-AC-017` | Sort key derivation is deterministic. | Every emitted vertex and edge has a reproducible `sort_key`. |
| `GP-AC-018` | Canonical identity tuples are deterministic. | Graph-view, vertex, edge, aggregation, digest, and validation-issue IDs derive from specified tuples. |
| `GP-AC-019` | Validation issue identity is stable. | Details are closed per code, and issue ordering has a final tie-breaker. |
| `GP-AC-020` | Validation severity and consumability are aligned. | No nonfatal issue makes the graph non-consumable. |
| `GP-AC-021` | Source item exclusion is bounded. | Item-level relationship, source-kind, and item-resource errors exclude affected items without rejecting unrelated valid items. |
| `GP-AC-022` | Lifecycle state ownership is separated. | Graph-view state and projection-run/result state are independently defined. |
| `GP-AC-023` | Failed-run inspection is portable. | `get_projection_run()` exposes retained failed-run validation or failure details. |
| `GP-AC-024` | Retention is exact. | Same projection history and retention policy produce the same retained replaced and failed run set. |
| `GP-AC-025` | Query behavior is closed. | Every lifecycle state maps to a query result or query error. |
| `GP-AC-026` | Traversal is fully deterministic. | BFS pseudocode covers empty seeds, unknown seeds, self-loops, multi-edges, filters, metadata, and ordering. |
| `GP-AC-027` | Pagination is closed. | `list_graph_views()` has default limit, bounds, cursor semantics, mutation behavior, ordering, and errors. |
| `GP-AC-028` | Canonical output bytes are defined. | Canonical JSON serialization produces byte-identical comparison inputs after allowed run-specific exclusions. |
| `GP-AC-029` | Implementation latitude is bounded. | Allowed internal variance cannot change observable output, validation, lifecycle, retention, or query behavior. |
| `GP-AC-030` | The spec is self-contained. | A competent implementer can implement graph projection behavior from this NLSpec without project-specific assumptions or external graph standards. |
| `GP-AC-031` | JSON duplicate-member behavior is closed. | Duplicate object members at every depth are rejected before run admission with deterministic lifecycle `invalid_projection_request` details and no validation issue object. |
| `GP-AC-032` | Digest bytes are canonical. | Golden fixtures produce byte-identical `projection_config_digest` and `projection_source_digest`. |
| `GP-AC-033` | Lifecycle operation contracts are closed. | Create, refresh, invalidation, idempotency, retry, and concurrent active-run behavior produce the specified states and errors. |
| `GP-AC-034` | Failed-run retention is exact. | The same failed-run history and retention policy produce the same retained failed-run set. |
| `GP-AC-035` | Invalidation scope is exact. | Graph-view invalidation cascades and run-specific invalidation affect exactly the specified retained runs. |
| `GP-AC-036` | Validation issue construction is deterministic. | Every validation code has deterministic target ID, field, detail, reason-code, and issue ID behavior. |
| `GP-AC-037` | Validation issue cap is deterministic. | Inputs exceeding the issue cap select the first `N-1` non-cap issues by discovery order, sort those selected issues by ordinary issue ordering, and append one final cap issue. |
| `GP-AC-038` | Resource-limit severity is aligned. | Whole-input limits are fatal; source-item limits exclude only the affected item; output limits are fatal. |
| `GP-AC-039` | Wildcard collision behavior is closed. | Wildcard and concrete property or metadata definitions cannot collide after expansion. |
| `GP-AC-040` | Mapping property-key arrays are operationally closed. | Required and optional mapping property keys resolve only to applicable property definitions and never synthesize passthrough output. |
| `GP-AC-041` | Edge labels are deterministic. | Direct and reverse edge labels follow the edge label derivation table. |
| `GP-AC-042` | Reverse-edge validity is static. | `emit_reverse_edge=true` is valid only with `normalize_forward` or `normalize_reverse`. |
| `GP-AC-043` | Aggregated endpoint matching is deterministic. | Endpoint grouping digests use the referenced vertex aggregation rule's digest context. |
| `GP-AC-044` | Aggregation merge compatibility is closed. | Every merge behavior and projected type combination is accepted or rejected by table. |
| `GP-AC-045` | Metadata mapping contexts are closed. | Graph-view, direct, reverse, and aggregated metadata derivation use the specified source contexts. |
| `GP-AC-046` | Query errors are portable. | Every query error has closed details, retryability, and no partial-result behavior. |
| `GP-AC-047` | Nonterminal run queries are closed. | Accepted and computing runs are inspectable through `get_projection_run()` and not graph-readable. |
| `GP-AC-048` | Pagination mutation behavior is closed. | Live keyset pagination, cursor expiry, deletion, insertion, and limit-change cases produce deterministic pages or `cursor_invalid`. |
| `GP-AC-049` | Traversal metadata is closed. | `traverse(...).metadata` is exactly `{}` in this revision. |
| `GP-AC-050` | Canonical comparison fixtures state run specificity. | Conformance fixtures declare run-specific versus run-independent comparison inputs and exclude only the allowed fields. |
| `GP-AC-051` | Admission boundary is closed. | Pre-admission failures return lifecycle operation errors, allocate no run, emit no validation issue, and create no retained run. |
| `GP-AC-052` | Admitted failed-run boundary is closed. | Fatal validation after admission creates one retained failed run inspectable through `get_projection_run()`. |
| `GP-AC-053` | Lifecycle response envelopes are closed. | Every lifecycle operation returns exactly one success or error variant with the specified members and no partial data on error. |
| `GP-AC-054` | Invalidation success shape is closed. | Graph-view and run invalidation return `invalidation_summary` with exact member order, nullability, run-ID list ordering, and idempotent replay behavior. |
| `GP-AC-055` | `idempotency_key` behavior is closed. | Omitted, explicit null, invalid, exact replay, conflict reuse, and expiry cases produce the specified outcomes. |
| `GP-AC-056` | Early validation issue identity is impossible. | Invalid JSON, duplicate members, and invalid top-level identity fields never emit `issue_id`. |
| `GP-AC-057` | Aggregated property candidate derivation is closed. | Missing, null, default, invalid, required, count, and merge cases produce deterministic candidates, output keys, and issues. |
| `GP-AC-058` | Aggregated endpoint key evaluation is closed. | Missing, null, object, nested array, digest-not-found, `error`, and `exclude` cases produce deterministic edge and issue behavior. |
| `GP-AC-059` | Aggregated output labels are closed. | Aggregated vertices and edges always emit the defined default-label arrays. |
| `GP-AC-060` | Schema registry label and wildcard behavior is closed. | Registry label arrays, `source_labels_preserved`, and concrete wildcard-expanded property and metadata entries match fixtures. |
| `GP-AC-061` | Validation condition matrix is exhaustive. | Every validation condition maps to exactly one code, severity, target, field path, details object, and reason code where applicable. |
| `GP-AC-062` | Canonical string bytes are closed. | Canonical JSON string escaping matches the required grammar for all control, quote, reverse-solidus, solidus, U+2028/U+2029, and non-BMP cases. |
| `GP-AC-063` | Scalar lexical boundaries are closed. | Integer, timestamp, and identifier whitespace fixtures accept and reject exactly the specified cases. |
| `GP-AC-064` | Lifecycle timestamp ownership is closed. | Generated timestamps use fixed precision, monotonic per-graph-view ordering, and server-owned lifecycle assignment. |
| `GP-AC-065` | Graph-view summaries are state-complete. | Every graph-view state in the summary matrix emits exact field values and validation status. |
| `GP-AC-066` | Retention inspection is closed. | Duration-bound and count-bound retention fixtures produce deterministic `retention_expires_at` and addressability behavior. |
| `GP-AC-067` | Digest fixtures are byte-reproducible. | Implementations reproduce fixture canonical bytes, tuple fields, and digest outputs exactly. |

## 14. Conformance fixture registry

A conformance fixture MUST define its input JSON or operation request, normalized input after defaults when applicable, expected operation response or query response, expected retained state, expected validation summary when a run is admitted, and expected canonical bytes or digest bytes when bytes are under test.

| Fixture ID | Required coverage |
| --- | --- |
| `GP-FIX-001` | Malformed JSON pre-admission rejection. |
| `GP-FIX-002` | Duplicate object member pre-admission rejection. |
| `GP-FIX-003` | Invalid derived `graph_view_id` pre-admission rejection. |
| `GP-FIX-004` | Admitted invalid mapping creates retained failed run. |
| `GP-FIX-005` | Issue cap final ordering. |
| `GP-FIX-006` | Aggregated property missing/default/null candidate behavior. |
| `GP-FIX-007` | Aggregation `single_value` merge conflict after default materialization. |
| `GP-FIX-008` | Aggregation `count` ignores `source_field_path` candidate evaluation. |
| `GP-FIX-009` | Missing aggregated endpoint key with `error`. |
| `GP-FIX-010` | Missing aggregated endpoint key with `exclude`. |
| `GP-FIX-011` | Endpoint digest computed but no matching vertex. |
| `GP-FIX-012` | Schema registry default labels, mapping labels, and source label preservation indicator. |
| `GP-FIX-013` | Wildcard property and metadata registry expansion. |
| `GP-FIX-014` | Canonical JSON string escaping. |
| `GP-FIX-015` | Integer lexical forms. |
| `GP-FIX-016` | Timestamp calendar validity. |
| `GP-FIX-017` | Identifier Unicode whitespace boundary. |
| `GP-FIX-018` | Lifecycle idempotency exact replay and conflict. |
| `GP-FIX-019` | Invalidation summary graph-view cascade. |
| `GP-FIX-020` | Graph-view summary state matrix. |
| `GP-FIX-021` | Count-bound retention expiry before duration expiry. |
| `GP-FIX-022` | Full digest byte transcript for minimal empty graph. |
| `GP-FIX-023` | Full digest byte transcript for one host property graph. |

Every fixture MUST be deterministic. Fixture IDs are stable. A fixture may add explanatory examples, but the expected operation response, retained state, validation summary, and canonical bytes are the normative comparison artifacts for that fixture.
