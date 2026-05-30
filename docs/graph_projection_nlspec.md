---
title: Graph Projection NLSpec
status: draft
document_class: nlspec
created_at: 2026-05-30
---

## 1. Purpose and authority

This NLSpec defines the normative contract for **Graph Projection**: the deterministic derivation of graph-oriented representations from authoritative source data.

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

A projection input received as bytes MUST be decoded as UTF-8 JSON before schema validation. Invalid UTF-8, invalid JSON syntax, or a decoded top-level value that is not a JSON object MUST produce `invalid_input_shape`.

A conforming implementation MUST reject duplicate JSON object member names at every object depth before schema validation. Duplicate member handling MUST NOT depend on parser first-wins, last-wins, insertion-order, or map-overwrite behavior. A duplicate member MUST produce one `invalid_input_shape` issue with:

| Field | Required value |
| --- | --- |
| `target_kind` | `projection_input` |
| `target_id` | `projection_input` |
| `field_path` | canonical input path to the duplicate member |
| `details.field` | same canonical input path |
| `details.reason_code` | `duplicate_object_member` |

Schema tables use `[]` as type notation, not as a literal JSON member-name suffix. A table member written as `source_entities[]` denotes the JSON member named `source_entities` whose value is an array. The canonical input path for that member is `$.source_entities`.

Canonical input paths MUST use this grammar.

| Path part | Required syntax |
| --- | --- |
| Root | `$` |
| Object member with an ASCII path identifier | `.` followed by the exact member name |
| Object member without an ASCII path identifier | `[` + canonical JSON string for the exact member name + `]` |
| Array element | `[` + zero-based base-10 index with no leading zeroes except `0` + `]` |

An ASCII path identifier is a non-empty string whose first character is `A-Z`, `a-z`, or `_`, and whose remaining characters are `A-Z`, `a-z`, `0-9`, or `_`. Canonical input paths are diagnostic identifiers only. They do not change the field-path grammar in §4.2.

Validation detail members named `field` MUST use canonical input path syntax unless the owning detail schema explicitly says the member carries a §4.2 `field_path`.

### 4.1 Common scalar contracts

The following scalar contracts apply wherever referenced.

| Contract | Definition |
| --- | --- |
| `identifier` | A JSON string containing 1 to 128 Unicode scalar values. It MUST NOT contain U+0000, Unicode surrogate code points, C0 controls, C1 controls, leading whitespace, trailing whitespace, `/`, `\\`, or `#`. It MUST be compared by exact Unicode code point sequence after validation. No case folding, trimming, locale comparison, or Unicode normalization is applied. |
| `generated_id` | An `identifier` with a required prefix from §7.1 followed by exactly 64 lowercase hexadecimal characters. |
| `timestamp` | A JSON string in UTC form `YYYY-MM-DDTHH:MM:SSZ` or `YYYY-MM-DDTHH:MM:SS.ffffffZ`, where fractional seconds, when present, contain 1 to 6 decimal digits. Leap seconds are invalid. A timestamp with fractional seconds MUST preserve the supplied fractional precision after validation. |
| `property_key` | An `identifier` used as an object member key for projected properties or mapped metadata. It MUST NOT contain `.`. It MUST NOT equal `kind`, `properties`, `metadata`, `source_metadata`, or `projected` when used as a field-path terminal. |
| `kind` | An `identifier` that names a source entity kind, source relationship kind, projected vertex kind, or projected edge kind. |
| `finite_integer` | A JSON number with no fractional part in the inclusive range `-9007199254740991` through `9007199254740991`. |
| `property_value` | One of JSON string, JSON boolean, `finite_integer`, JSON null, or an array of those scalar values. A JSON string value MUST contain no more than `max_string_property_value_length` Unicode scalar values. Arrays MUST contain at most 1024 values. Nested arrays and objects are invalid as property values. |
| `sha256_hex` | A JSON string containing exactly 64 lowercase hexadecimal characters. |
| `field_path` | A JSON string matching exactly one row in §4.2. It is parsed by splitting on literal `.` into exactly the declared path segments. No escaping is supported in this NLSpec revision. |
| `cursor_token` | A JSON string containing 1 to 4096 Unicode scalar values. The token is opaque to the caller and valid only under the query-specific cursor rules that emitted it. |

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

No top-level projection input member is nullable. Explicit JSON `null` for any top-level member is invalid and MUST produce `explicit_null_not_allowed`.

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
| `max_validation_issues` | emitted validation issues | `100000` | Closed by §9.3: emit first `N-1` issues plus final `validation_issue_limit_exceeded`. |
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

Whole-input limits MUST use `resource_limit_exceeded` and prevent a consumable graph. Item-scoped limits MUST use `source_item_resource_limit_exceeded` and exclude only the affected item. Derived-output limits MUST use `projected_output_limit_exceeded` and prevent a consumable graph.

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

The graph view output MUST include `schema_registry` so consumers can rely on declared vertex and edge schemas, labels, property keys, and metadata keys without private configuration access.

`schema_registry` MUST be a JSON object with exactly the following members.

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `vertex_kinds[]` | Array | Yes | No | Sorted by `vertex_kind`. |
| `edge_kinds[]` | Array | Yes | No | Sorted by `edge_kind`. |
| `property_keys[]` | Array | Yes | No | All declared projected property keys sorted by `target_scope`, `target_kind`, `projected_key`. |
| `metadata_keys[]` | Array | Yes | No | All declared mapped metadata keys sorted by `target_scope`, `target_kind`, `projected_metadata_key`. |

Each `vertex_kinds[]` item MUST be a JSON object with exactly these members.

| Member | Rule |
| --- | --- |
| `vertex_kind` | Projected vertex kind. |
| `source_entity_kinds[]` | Source entity kinds that can emit this vertex kind, sorted. |
| `aggregation_rule_ids[]` | Aggregation rules that can emit this vertex kind, sorted. |
| `labels[]` | Mapping-declared labels only, sorted. |
| `properties[]` | Property schema references for this kind, sorted by `projected_key`. |

Each `edge_kinds[]` item MUST be a JSON object with exactly these members.

| Member | Rule |
| --- | --- |
| `edge_kind` | Projected edge kind. |
| `source_relationship_kinds[]` | Source relationship kinds that can emit this edge kind, sorted. |
| `aggregation_rule_ids[]` | Aggregation rules that can emit this edge kind, sorted. |
| `directions[]` | Projected edge directions that may be emitted for this kind, sorted in order `directed`, `undirected`, `bidirectional`. |
| `labels[]` | Mapping-declared labels only, sorted. |
| `properties[]` | Property schema references for this kind, sorted by `projected_key`. |

Each property schema item MUST be a JSON object with exactly these members.

| Member | Rule |
| --- | --- |
| `target_scope` | `graph_view`, `vertex`, or `edge`. |
| `target_kind` | Kind or `*`. |
| `projected_key` | Property key. |
| `projected_type` | Closed projected type. |
| `required` | Boolean from property definition. |
| `nullable_output` | Boolean derived from `null_output_policy=emit_null`. |
| `missing_behavior` | Effective missing behavior after defaults. |
| `source_null_behavior` | Effective null behavior after defaults. |

Each metadata schema item MUST use the same shape as a property schema item except the key member is `projected_metadata_key`.

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
| Strings | JSON escaping is used only for `"`, `\\`, and required control escapes. Other Unicode scalar values emit as UTF-8. |
| Integers | Base-10 ASCII without leading zeroes, except zero emits `0`. |
| Booleans | `true` or `false`. |
| Null | `null`. |
| Timestamps | The validated timestamp string form from §4.1. |
| Floating point | Forbidden by scalar contracts. |

Canonical graph-output equivalence is byte identity of canonical JSON after excluding the run-specific fields listed in §12.2.

## 6. Mapping and derivation algorithms

Projection derivation MUST follow the algorithms and mappings in this section after input validation and normalization.

### 6.1 Projection algorithm overview

A conforming implementation MUST derive a projection result in this order.

```text
1. Decode input as a JSON-compatible object.
2. Validate top-level shape, scalar contracts, closed schemas, unknown members, duplicates, and resource limits.
3. Normalize inputs according to §4.13.
4. Derive and validate graph_view_id according to §7.4.
5. Select the active relationship mapping source:
   a. if relationship_definitions[] is non-empty, use it;
   b. otherwise use projection_config.relationship_mappings[].
6. Validate mapping completeness and mapping cross-references.
7. Evaluate global source entity and source relationship filters.
8. Exclude semantically invalid source items and emit nonfatal validation issues as required.
9. Emit direct vertices from eligible source entities.
10. Emit direct and reverse edges from eligible source relationships whose endpoint vertices exist.
11. Emit aggregated vertices and aggregated edges in normalized aggregation rule order.
12. Derive graph-view properties and metadata.
13. Build schema_registry and consumer_capabilities.
14. Run post-projection output validation.
15. If no fatal issue exists, publish a consumable projection result atomically.
16. If any fatal issue exists, persist a failed projection run summary and do not publish a consumable graph result.
```

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

For each non-empty vertex aggregation group, emit exactly one aggregated vertex with:

- `vertex_family=aggregated`;
- `vertex_kind=aggregation_rule.projected_kind`;
- `source_entity_ref=null`;
- `metadata.aggregation_rule_id=aggregation_rule_id`;
- `metadata.aggregation_source_refs[]` from all group contributors, sorted by §5.5.4;
- identity derived by §7.6.

#### 6.8.6 Property merge behavior

When aggregation emits properties or mapped metadata, the effective merge behavior for a key is the per-rule `property_merge_behavior` override when present, otherwise the property or metadata definition's `merge_behavior`.

| `merge_behavior` | Input values considered | Output | Conflict behavior |
| --- | --- | --- | --- |
| `single_value` | Present non-null values | Emit the one distinct value. | More than one distinct value emits `aggregation_merge_conflict` and omits the key. |
| `first_by_sort` | Present non-null values | Value from first contributor by contributor sort key. | No conflict. |
| `last_by_sort` | Present non-null values | Value from last contributor by contributor sort key. | No conflict. |
| `distinct_sorted_array` | Present scalar values and array element values | Distinct scalar values sorted by canonical value order. | Non-scalar object or nested array emits `invalid_property_type`; invalid values are skipped. |
| `count` | All contributors in group | Integer count of contributing items. | No conflict. |
| `omit` | None | Omit key. | No conflict. |

Canonical value order sorts first by JSON type rank `null`, `boolean`, `integer`, `string`, then by canonical JSON bytes.

#### 6.8.6.1 Merge behavior compatibility

Merge behavior MUST be compatible with the declared `projected_type` for the property definition or metadata mapping that owns the emitted key.

| Merge behavior | Allowed projected types | Null handling | Array handling | Invalid combination |
| --- | --- | --- | --- | --- |
| `single_value` | all projected types | Ignore JSON null inputs. | Arrays compare as whole normalized values. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `first_by_sort` | all projected types | Ignore JSON null inputs. | Emits first whole normalized value. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `last_by_sort` | all projected types | Ignore JSON null inputs. | Emits last whole normalized value. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `distinct_sorted_array` | `string_array`, `identifier_array` only | Ignore JSON null inputs. | Flattens scalar strings and array elements of the declared element type. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `count` | `integer` only | Counts contributors, not values. | Arrays are irrelevant. | `invalid_property_definition` or `invalid_metadata_mapping`. |
| `omit` | all projected types | No values considered. | No values considered. | Never invalid by type. |

`distinct_sorted_array` for `identifier_array` MUST validate every emitted element as `identifier`. `distinct_sorted_array` for `string_array` MUST accept only JSON strings. `single_value` conflict detection MUST compare canonical JSON bytes after projected-type normalization.

If no non-null values remain after missing/null behavior, the key is omitted unless the property definition or metadata mapping is required. A required property emits `required_property_missing`; a required metadata mapping emits `invalid_metadata_mapping` with `reason_code=required_metadata_missing`.

#### 6.8.7 Aggregated edge emission

For each non-empty edge aggregation group, the implementation MUST derive source and destination endpoint grouping digests through the referenced vertex aggregation rule context.

Endpoint digest derivation MUST use this algorithm for each endpoint side.

```text
1. Let R be the referenced vertex aggregation rule for the endpoint side.
2. Evaluate the endpoint side's grouping key paths in the edge aggregation rule's contributor context.
3. Construct canonical_endpoint_grouping_key as the canonical JSON array of those values in declared endpoint key order.
4. Compute endpoint_digest = sha256_hex(serialize_tuple("GPGROUP1\n", [
     R.aggregation_rule_id,
     "vertex",
     R.projected_kind,
     canonical_endpoint_grouping_key
   ])).
5. Match only an emitted aggregated vertex whose canonical_grouping_key_digest equals endpoint_digest and whose aggregation rule ID equals R.aggregation_rule_id.
```

The endpoint digest MUST NOT match by projected kind alone. The edge aggregation rule is invalid with `invalid_aggregation_rule` when an endpoint grouping key count does not equal the referenced vertex aggregation rule's `grouping_keys[]` count.

For each group whose endpoints are resolved, emit exactly one aggregated edge with:

- `edge_family=aggregated`;
- `edge_kind=aggregation_rule.projected_kind`;
- `src_vertex_id` equal to the matched source aggregated vertex;
- `dst_vertex_id` equal to the matched destination aggregated vertex;
- `direction=aggregation_rule.edge_direction`;
- `source_relationship_ref=null`;
- `metadata.aggregation_rule_id=aggregation_rule_id`;
- `metadata.aggregation_source_refs[]` from all group contributors, sorted by §5.5.4;
- identity derived by §7.6.

#### 6.8.8 Missing aggregated endpoints

| `endpoint_grouping.missing_endpoint_behavior` | Required behavior |
| --- | --- |
| `error` | Emit `aggregation_endpoint_missing`; do not emit the aggregated edge. |
| `exclude` | Do not emit the aggregated edge and do not emit a validation issue. |

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

Any validation issue that prevents default materialization or prevents deterministic normalization prevents digest emission for a consumable graph result.

Golden fixture: minimal empty graph digest inputs.

| Digest | Expected lowercase SHA-256 |
| --- | --- |
| `projection_config_digest` | `c0d919dc8a5093e2e6f81eab3a4f2b0a9e03381e5fc834144f33a3d737fb2b06` |
| `projection_source_digest` | `e7bb613e7b8b295359e0ebba2f7ab6fe845b00f5de94f82f599327a409dcc56c` |

This fixture uses `projection_schema_id="graph_projection.v1"`, `graph_view_key="empty"`, `source_snapshot_id="snap_empty_1"`, `allow_empty_kind_registry=true`, default filters, empty source arrays, empty property definitions, empty metadata mappings, empty aggregation rules, no relationship mappings, and the default retention policy from §10.6.

Golden fixture: one host property digest inputs.

| Digest | Expected lowercase SHA-256 |
| --- | --- |
| `projection_config_digest` | `b61205128991059ae08ae5c1cd9594b4c8f21b3e0b497e6d8a92cdd27a6dbee5` |
| `projection_source_digest` | `2a63da49700e5cfe7bfe0fd5e8f9d0e8ccbd03d8297abf3cef6774fd5c578e31` |

This fixture uses one entity mapping `map_host`, one relationship mapping `map_logon` from `projection_config.relationship_mappings[]`, one required string property definition `pd_hostname`, one source entity `host1` with `properties.hostname="WS-023"`, `source_snapshot_id="snap_incident_1"`, `source_metadata.case="alpha"`, default filters, no source relationships, no metadata mappings, no aggregation rules, and the default retention policy from §10.6.

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

A validation issue ID MUST equal `gpi_` plus SHA-256 over tuple prefix `GPISSUE1\n` and fields:

1. `projection_schema_id`,
2. `graph_view_id`,
3. `projection_run_id`,
4. `severity`,
5. `code`,
6. `target_kind`,
7. `target_id`,
8. canonical JSON of required code-specific `details` members only.

The human-readable `message` and any transport-level context MUST NOT affect `issue_id`.

## 8. Defaults, omissions, and explicit nulls

### 8.1 Default and omission table

| Boundary | Omitted behavior | Explicit `null` behavior |
| --- | --- | --- |
| Top-level required input member | `missing_required_input` | `explicit_null_not_allowed` |
| Top-level optional object | Materialize documented default | `explicit_null_not_allowed` |
| Top-level optional array | Materialize `[]` unless another default is stated | `explicit_null_not_allowed` |
| `projection_schema_id` | `missing_required_input`; no default exists | `explicit_null_not_allowed` |
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

Validation MUST execute in the phases in this table. A fatal issue in a phase whose `Stops later phases when fatal?` value is `Yes` prevents all later phases from running.

| Phase | Scope | Stops later phases when fatal? |
| ---: | --- | ---: |
| 1 | JSON decoding and duplicate member detection | Yes |
| 2 | Top-level object shape, unknown members, requiredness, and nullability | Yes |
| 3 | Scalar contracts, enum contracts, and whole-input resource limits | Yes |
| 4 | Projection configuration schemas and default materialization | Yes |
| 5 | Field-path, filter, mapping, property, metadata, aggregation, retention, wildcard, and merge-compatibility validation | Yes |
| 6 | Source entity and source relationship item validation and item exclusion | No |
| 7 | Direct mapping derivation validation | No, unless a fatal issue is produced |
| 8 | Aggregation derivation validation | No, unless a fatal issue is produced |
| 9 | Output schema and projected-output resource-limit validation | Yes |
| 10 | Validation summary construction | N/A |

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

The implementation MUST collect all issues reachable under the phase rules until the validation issue cap in §9.3 is reached. It MUST NOT choose an arbitrary subset of issues.

### 9.2 Validation severities

| Severity | Meaning | Consumability |
| --- | --- | --- |
| `fatal` | Projection cannot emit a conforming consumable graph result. | No consumable graph is published for the run. |
| `error` | A source item, derived object, property, metadata key, or aggregation output is excluded or degraded under a specified rule. | Graph may be consumable when no fatal issue exists. |
| `warning` | Projection completed with a non-blocking concern that does not change required output shape. | Graph is consumable. |
| `info` | Diagnostic information only. | Graph is consumable. |

No `error` severity issue may by itself make the graph non-consumable. Any behavior that makes a graph non-consumable MUST use a `fatal` issue.

### 9.3 Validation issue shape, ordering, identity, and cap behavior

A validation issue MUST be a JSON object with exactly the following members.

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

Validation issues MUST sort for output by:

1. severity order `fatal`, `error`, `warning`, `info`,
2. `code`,
3. `target_kind`,
4. `target_id`,
5. `issue_id`.

Issue-cap behavior is closed. Let `N=max_validation_issues`. If total discovered issues are `<= N`, the validation summary MUST emit all sorted issues. If total discovered issues are `> N`, the validation summary MUST emit the first `N-1` issues by deterministic discovery order after issue construction, then emit one `validation_issue_limit_exceeded` issue as the final issue. The cap issue participates in severity counts. The cap issue MUST have `target_kind=projection_input`, `target_id=projection_input`, `field_path=null`, and `details.limit=N`.

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

### 9.5 Closed validation code registry

The implementation MUST emit only the validation codes in this table.

| Code | Severity | Target kind | Required meaning |
| --- | --- | --- | --- |
| `invalid_input_shape` | `fatal` | `projection_input` | JSON decoding, JSON shape, scalar type, duplicate object member, or schema structure is invalid. |
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
| `aggregation_endpoint_missing` | `error` | `mapping_rule` | Aggregated edge endpoint vertex cannot be resolved. |
| `aggregation_merge_conflict` | `error` | `mapping_rule` | Aggregation merge behavior cannot produce one conforming value. |
| `resource_limit_exceeded` | `fatal` | `projection_input` | Whole projection input exceeds a closed whole-input resource limit. |
| `projected_output_limit_exceeded` | `fatal` | `graph_view` | Derived output would exceed a closed projected-output limit. |
| `validation_issue_limit_exceeded` | `fatal` | `projection_input` | Validation issue cap reached. |
| `invalid_retention_policy` | `fatal` | `projection_config` | Retention policy schema or bounds are invalid. |
| `output_schema_violation` | `fatal` | `graph_view` | Derived output would violate this NLSpec. |
| `projection_computation_failed` | `fatal` | `graph_view` | Computation failed before conforming output could be emitted. |

Implementations MUST NOT emit `missing_mapping_rule`. That token is reserved as a non-emitted historical alias.

### 9.6 Validation detail schemas

For every validation issue, `details` MUST contain exactly the required keys in this table unless optional keys are listed. Optional keys, when present, participate in neither issue identity nor ordering. Unknown detail keys are forbidden.

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
| `invalid_property_type` | `projected_key`, `expected_type`, `actual_type` | none |
| `required_property_missing` | `projected_key`, `source_field_path` | none |
| `source_null_for_required_property` | `projected_key`, `source_field_path` | none |
| `undeclared_source_kind` | `source_item_id`, `source_kind` | none |
| `source_item_resource_limit_exceeded` | `source_item_id`, `limit_key`, `limit`, `observed` | none |
| `missing_relationship_endpoint` | `source_relationship_id`, `endpoint_field` | none |
| `relationship_endpoint_not_projected` | `source_relationship_id`, `endpoint_field`, `endpoint_source_entity_id` | none |
| `invalid_relationship_direction` | `source_relationship_id`, `supplied_value` | none |
| `invalid_direction_policy` | `mapping_rule_id`, `supplied_value` | none |
| `invalid_reverse_edge_policy` | `mapping_rule_id`, `projected_direction` | none |
| `invalid_aggregation_rule` | `aggregation_rule_id`, `reason_code` | none |
| `aggregation_grouping_key_missing` | `aggregation_rule_id`, `field_path`, `contributor_id` | none |
| `aggregation_endpoint_missing` | `aggregation_rule_id`, `endpoint_side`, `endpoint_digest` | none |
| `aggregation_merge_conflict` | `aggregation_rule_id`, `projected_key` | none |
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
| `invalid_input_shape` | `invalid_utf8`, `invalid_json_syntax`, `top_level_not_object`, `duplicate_object_member`, `schema_type_mismatch`, `scalar_contract_violation`, `property_value_too_long`, `array_element_invalid`, `array_length_exceeded`, `invalid_label` |
| `invalid_projection_config` | `custom_config_referenced`, `relationship_mapping_source_conflict`, `empty_kind_registry_not_allowed`, `declared_kind_duplicate`, `mapping_rule_duplicate`, `metadata_mapping_duplicate`, `aggregation_rule_duplicate`, `unknown_configuration_member`, `invalid_default_materialization` |
| `invalid_filter` | `invalid_operator`, `value_required`, `value_forbidden`, `invalid_field_scope`, `invalid_value_shape`, `unsupported_logic` |
| `invalid_mapping_rule` | `duplicate_mapping_rule_id`, `declared_source_kind_missing`, `property_key_not_defined`, `property_requiredness_mismatch`, `required_optional_overlap`, `reverse_edge_kind_without_reverse`, `label_invalid` |
| `invalid_metadata_mapping` | `reserved_metadata_key`, `duplicate_after_wildcard_expansion`, `invalid_source_scope`, `invalid_default_value`, `invalid_merge_behavior_type`, `invalid_projected_type`, `required_metadata_missing` |
| `invalid_property_definition` | `duplicate_after_wildcard_expansion`, `invalid_source_scope`, `invalid_default_value`, `invalid_null_policy`, `invalid_merge_behavior_type`, `invalid_projected_type` |
| `invalid_aggregation_rule` | `dependency_on_later_rule`, `aggregation_cycle`, `endpoint_rule_not_vertex_rule`, `endpoint_grouping_key_count_mismatch`, `endpoint_field_scope_invalid`, `grouping_key_invalid`, `invalid_endpoint_behavior`, `invalid_edge_direction`, `input_scope_invalid`, `invalid_merge_behavior_type` |
| `invalid_retention_policy` | `unknown_member`, `out_of_bounds`, `invalid_type`, `explicit_null_not_allowed` |
| `output_schema_violation` | `id_mismatch`, `reference_missing`, `sort_order_invalid`, `schema_registry_mismatch`, `metadata_shape_invalid`, `closed_schema_violation`, `canonical_serialization_invalid` |
| `projection_computation_failed` | `internal_exception`, `dependency_unavailable`, `timeout`, `resource_exhausted`, `implementation_invariant_failed` |

## 10. Projection lifecycle and retention

Lifecycle state is split into graph-view state and projection-run/result state. A state name belongs to exactly one state machine unless the tables both declare the same token with machine-local meaning.

### 10.0 Lifecycle operation interfaces

Lifecycle operations are abstract interfaces. A conforming implementation MAY expose them through any transport, but the request members, defaults, idempotency, concurrency, state transitions, and errors in this section MUST remain observable.

| Operation | Required request members | Optional request members | Success result | Primary errors |
| --- | --- | --- | --- | --- |
| `create_projection` | `projection_input` | `idempotency_key` | Accepted run summary | `invalid_operation`, `operation_conflict`, validation errors |
| `refresh_projection` | `graph_view_id`, `projection_input` | `idempotency_key` | Accepted run summary | `projection_not_available`, `invalid_operation`, `operation_conflict`, validation errors |
| `invalidate_graph_view` | `graph_view_id`, `reason_code`, `requested_at`, `requested_by` | `idempotency_key` | Invalidation summary | `graph_view_not_found`, `invalid_operation`, `operation_conflict` |
| `invalidate_projection_run` | `graph_view_id`, `projection_run_id`, `reason_code`, `requested_at`, `requested_by` | `idempotency_key` | Invalidation summary | `projection_run_not_found`, `invalid_operation`, `operation_conflict` |

An accepted run summary MUST contain exactly `graph_view_id`, `projection_run_id`, `state`, `source_snapshot_id`, `projection_version`, and `accepted_at`. `state` MUST be `accepted`.

Lifecycle operation errors MUST use this registry.

| Error | Retryable | Required details |
| --- | ---: | --- |
| `invalid_operation` | No | `operation`, `reason_code` |
| `operation_conflict` | Yes only when `reason_code=run_already_active`; otherwise No | `operation`, `reason_code`, optional `active_projection_run_id` |

Reason codes for lifecycle operation errors are closed.

| Reason code | Applies to | Meaning |
| --- | --- | --- |
| `graph_view_already_exists` | `create_projection` | A graph view already has a selected available or refreshing run; caller must use refresh. |
| `graph_view_not_created` | `refresh_projection` | No graph view exists for refresh. |
| `no_consumable_prior_run` | `refresh_projection` | Refresh requires an available prior run unless the graph view is invalidated. |
| `run_already_active` | create or refresh | Graph view already has an `accepted` or `computing` run. |
| `idempotency_key_conflict` | any operation | Same operation idempotency key was reused with different normalized request bytes. |
| `invalid_invalidation_target` | invalidation operations | Target state cannot be invalidated. |
| `invalid_reason_code` | invalidation operations | Reason code is outside §10.7. |

`create_projection` is allowed only when graph-view state is `not_created`, `failed`, or absent. It creates a new run in `accepted`. `refresh_projection` is allowed only when graph-view state is `available` or `invalidated`. It creates a new run in `accepted`. Retrying after an initial failed create uses `create_projection`, not `refresh_projection`.

At most one `accepted` or `computing` run may exist per graph view at a time. A concurrent create or refresh with a different idempotency key while a run is active MUST return `operation_conflict` with `reason_code=run_already_active` and `active_projection_run_id` set to the active run.

Without an `idempotency_key`, a successful create or refresh creates a new run only when no run is active. With an `idempotency_key`, exact replay of the same normalized operation request MUST return the original accepted run summary while that run remains inspectable. Reusing the same operation and idempotency key with different normalized request bytes MUST return `operation_conflict` with `reason_code=idempotency_key_conflict`.

Normalized operation request bytes MUST be canonical JSON of an object with exact members `operation`, `graph_view_id` when applicable, `projection_input` after default materialization when applicable, `projection_run_id` when applicable, `reason_code` when applicable, `requested_at` when applicable, and `requested_by` when applicable. Unknown operation request members are invalid.

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

If computation fails before ordinary validation can produce a summary, the run's `validation_summary` MUST have `status=failed`, `fatal_count=1`, and exactly one `projection_computation_failed` issue.

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
| `graph_views[]` | Page of graph-view summary objects sorted by `graph_view_id`. |
| `next_cursor_token` | String for the next page, or JSON null when no next page exists. |

Each graph-view summary object MUST contain exactly `graph_view_id`, `graph_view_key`, `state`, `latest_projection_run_id`, `latest_source_snapshot_id`, `projection_version`, `updated_at`, and `validation_status`. The `state` member is graph-view state from §10.1.

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
| `retention_expires_at` | Yes | Yes | Non-null only when the run is retained under duration-bound retention; `null` for latest available and always-retained latest failed initial run. |

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
| `GP-AC-031` | JSON duplicate-member behavior is closed. | Duplicate object members at every depth are rejected before schema validation with deterministic `invalid_input_shape` details. |
| `GP-AC-032` | Digest bytes are canonical. | Golden fixtures produce byte-identical `projection_config_digest` and `projection_source_digest`. |
| `GP-AC-033` | Lifecycle operation contracts are closed. | Create, refresh, invalidation, idempotency, retry, and concurrent active-run behavior produce the specified states and errors. |
| `GP-AC-034` | Failed-run retention is exact. | The same failed-run history and retention policy produce the same retained failed-run set. |
| `GP-AC-035` | Invalidation scope is exact. | Graph-view invalidation cascades and run-specific invalidation affect exactly the specified retained runs. |
| `GP-AC-036` | Validation issue construction is deterministic. | Every validation code has deterministic target ID, field, detail, reason-code, and issue ID behavior. |
| `GP-AC-037` | Validation issue cap is deterministic. | Inputs exceeding the issue cap emit the first `N-1` issues by discovery order plus one cap issue. |
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