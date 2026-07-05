---
title: Network Flow Activity NLSpec
status: draft/proposed
profile_id: network_flow_activity
document_class: nlspec
---

# Network Flow Activity NLSpec

## 1. Status, scope, authority, and extension profile

Status: `draft/proposed`.

This NLSpec defines the proposed implementation-conformance contract for the `network_flow_activity` extension profile. It is not adopted/current until Core 00, Core 01, Core 03, and Core 04 recognize the extension profile, its route family, its import-target class, its extension tab, and its authorization/conformance boundary.

**NF-REQ-001**
The `network_flow_activity` extension profile MUST own only the following behavior families:

- incident-scoped Network Analysis workspace behavior;
- headered CSV import target behavior that creates Network Analysis table tabs;
- normalized network-flow table metadata, row storage, row validation, rejected-row diagnostics, and row query behavior;
- source-column mapping behavior for network-flow import targets;
- deterministic source-profile, parser-profile, timestamp-profile, mapping, canonicalization, and digest behavior;
- cross-table graph-composition behavior over selected active network-flow tables;
- flow-specific source-entity and source-relationship adapter input into the adopted Graph Projection subsystem;
- explicit indicator-linking initiation from selected flow values, graph vertices, graph edges, and flow rows;
- network-flow route-family request, response, default, omission, error, resource-limit, audit, and acceptance-criteria behavior.

**NF-REQ-002**
This NLSpec MUST NOT redefine Base Profile workbook routes, generic upload framing, Core record-envelope semantics, Core indicator identity, Core incident authorization, Core saved-view semantics, evidence access, deployment administration, OpenTelemetry, Testing Harness behavior, Core 05 claim publication, or generic Graph Projection behavior.

Observation creation from flow sources requires a future Core 02 amendment that admits extension-sourced `indicator_observation` rows through a closed `origin_kind` extension token and a non-record extension source reference. That amendment is a future adoption gate for a later Network Flow Activity revision, not a v1 requirement.

**NF-REQ-003**
When this NLSpec conflicts with Core 00 through Core 04 outside the Network Flow Activity extension boundary, the conflict is a defect in this NLSpec. When this NLSpec conflicts with the adopted Graph Projection NLSpec outside the flow-specific adapter and route behavior defined here, the Graph Projection NLSpec governs graph-projection behavior.

**NF-REQ-004**
When the `network_flow_activity` extension profile is claimed, the incident workbook shell MUST expose one top-level extension tab labeled `Network Analysis` for each incident visible to the caller. `Network Analysis` is an extension-contributed incident workspace. It is not a Base Profile built-in tab, not a Core system view, not a saved view, and not a Core record family.

**NF-REQ-005**
A successful CSV import targeting this extension MUST create exactly one new `network_flow_table` and one visible inner table tab within the top-level `Network Analysis` tab for each applied CSV import unit whose network-flow row validation admits at least one accepted row.

**NF-REQ-006**
Research reports, UI guides, implementation guides, appendices, and external vendor documents MAY justify design choices or source-profile fixtures, but they MUST NOT become implementation-conformance authority unless this NLSpec or Core 00 through Core 04 restates the behavior as a requirement.

Omission behavior: an implementation that ignores research reports, UI guides, implementation guides, appendices, or external vendor documents for conformance purposes remains conformant when it satisfies the normative requirements in this NLSpec and Core 00 through Core 04.

## 2. Normative language

**NF-REQ-007**
The key words **MUST**, **MUST NOT**, and **MAY** are normative inside this NLSpec.

| Term | Meaning |
| --- | --- |
| **MUST** | Conformance requirement. |
| **MUST NOT** | Conformance prohibition. |
| **MAY** | Optional behavior whose omission semantics are explicit in this document. |
| `default` | Required value or behavior when a caller, user, deployment, or owner-controlled configuration omits a more specific value. |

**NF-REQ-008**
A `MAY` statement in this NLSpec is valid only when the same paragraph, table row, or immediately following paragraph states the omission behavior. A `MAY` statement that affects public routes, persisted state, identifiers, digests, authorization, security, audit, or fixture output without explicit omission behavior is invalid.

**NF-REQ-009**
Object member names, field keys, route path segments, error codes, lifecycle states, profile IDs, schema IDs, digest labels, and closed-vocabulary tokens in this NLSpec MUST be compared by exact Unicode code point sequence after JSON decoding unless a requirement explicitly defines another normalization algorithm.

**NF-REQ-010**
A requirement in this NLSpec that refers to Core-owned behavior is a dependency statement, not a redefinition. The implementation MUST use the owning Core route, record, indicator, authorization, import, security, or graph-projection contract where this NLSpec names that owner.

**NF-REQ-011**
If this NLSpec defines a table with a declared scope, the table MUST cover every case in that scope. An omitted row in a closed mapping table means the case is invalid, not implementation-defined.

## 3. Adoption gates and Core amendment prerequisites

**NF-REQ-012**
This NLSpec MUST remain `status: draft` until every adoption gate in Table 3-A is satisfied.

**Table 3-A. Adoption gates**

| Gate ID | Owner artifact | Required adoption change | Required evidence before adoption |
| --- | --- | --- | --- |
| `NF-GATE-001` | Core 00 | Add `network_flow_activity` to the extension-profile model and adopted-subsystem map. | Core 00 lists this NLSpec as adopted for the extension boundary only. |
| `NF-GATE-002` | Core 01 | Add `network_flow_activity` to extension discovery returned by `GET /api/v1/extensions`, with route family `/api/v1/incidents/{incident_id}/network-flow`. | Extension discovery fixture returns the route family only when the profile is claimed. |
| `NF-GATE-003` | Core 01 | Extend import apply to permit extension-owned analytical import targets that produce durable extension resources rather than Core `record_id` rows. | Import contract names `target_kind='network_flow_table'` as an extension result target. |
| `NF-GATE-004` | Core 01 | Permit terminal import results to reference `network_flow_table` resources when `target_kind='network_flow_table'`. | Import result schema accepts extension resource references without treating them as saved views or record-envelope rows. |
| `NF-GATE-005` | Core 03 | Admit extension-contributed top-level incident tabs without adding `Network Analysis` to the Base Profile built-in tab list. | Base built-in tabs remain Timeline, Hosts, Identities, Evidence, and Notes; `Network Analysis` appears only when the extension is claimed. |
| `NF-GATE-006` | Core 04 | Add route-family authorization and conformance references for `network_flow_activity`, preserving incident membership, incident roles, and the no-`deployment_admin` incident-data bypass invariant. | Authorization fixtures prove route-time reauthorization and deny `deployment_admin` without incident membership. |

**NF-REQ-013**
The adoption process MUST NOT satisfy any gate by silently treating a flow table as a saved view, a Core system view, a Core record-envelope row family, a generic projection table, a visual graph artifact, or a base-profile workbook surface.

**NF-REQ-014**
Until all adoption gates are satisfied, every fixture row in §22 whose file bytes or expected output are not authored MUST retain an explicit `TODO:` value rather than pretending the conformance artifact exists.

## 4. Scope and non-goals

**NF-REQ-015**
The implementation MUST enforce the scope and omission behavior in Table 4-A.

**Table 4-A. Scope and non-goals**

| Area | Scope in this revision | Omission behavior |
| --- | --- | --- |
| Headered CSV import | Current v1 scope is headered CSV imported through the Import Extension Profile. | Raw binary flow messages and packet captures are unsupported and fail with `network_flow_unsupported_source_profile` or Core import unsupported-media behavior before Network Flow row validation. |
| Cisco Secure Network Analytics profile | `cisco_sna_netflow_csv_v1` is the only required v1 semantic source profile. | Generic IPFIX and NetFlow v9 CSV profiles are reserved future profiles and are not claimable in v1. |
| Raw IPFIX parser | Out of scope. | A raw IPFIX message, template stream, or exporter packet upload is unavailable. |
| Live collector ingest | Out of scope. | UDP/TCP listener, exporter registration, collector lifecycle, and live telemetry collection are unavailable. |
| PCAP and packet payloads | Out of scope. | Packet payload storage, stream reconstruction, packet parsing, and PCAP graphing are unavailable. |
| SIEM replacement | Out of scope. | Network Flow Activity remains incident-scoped analysis state, not a historical telemetry lake. |
| Third-party enrichment | Forbidden in v1. | No third-party lookup, geolocation, reputation, DNS, WHOIS, VirusTotal, MISP, or vendor API egress is available. |
| Automatic maliciousness inference | Forbidden in v1. | Graph shape, exact indicator suggestions, or high traffic volume may inform explicit user-visible suggestions, but MUST NOT change indicator lifecycle state automatically. |
| Per-row Core revisions | Out of scope. | `network_flow_row` values are immutable import source rows and do not receive Core record-envelope revision history. |
| Cross-incident graphing | Out of scope. | Graph queries are incident-scoped only. |
| Flow-row editing | Out of scope. | A correction requires importing a corrected source as a new table or soft-deleting the incorrect table and importing again. |
| Export | Out of scope in v1 public routes. | Any later export route requires a versioned export contract and CSV formula-injection mitigation before implementation. |
| Saved graph views | Out of scope in v1. | Graph responses are ephemeral route responses. The route does not create saved views, Graph Projection retained views, or workbook system views. |
| Time-bucketed graph output | Out of scope in v1. | A later revision MUST define UTC-epoch bucket alignment, bucket-count formula, bucketed response schema, and a versioned successor to `network_flow_graph_query_digest_v1` before bucket members return. |
| Binding list/read routes | Out of scope in v1. | Bindings are observable through the link-route response and the `network_flow_indicator_binding_created` audit event only. A later revision MAY add read routes. Omission behavior: v1 exposes no binding list, get, search, or indicator-centric binding route. |

## 5. Common public boundary, JSON admission, envelopes, and idempotency

### 5.1 Public route root and extension availability

**NF-REQ-016**
Network Flow public HTTP routes MUST be rooted at:

```text
/api/v1/incidents/{incident_id}/network-flow
```

The implementation MUST NOT expose an alternate public v1 route root for the same contract.

**NF-REQ-017**
When the `network_flow_activity` extension profile is not claimed, every Network Flow public route MUST fail through Core extension-unavailable behavior and the workbook shell MUST NOT render `Network Analysis` as an available incident tab.

### 5.2 Common success and error envelope use

**NF-REQ-018**
Network Flow public HTTP routes MUST use the Core public HTTP success and error envelope. Route-specific success payloads in this NLSpec define the value under `data`, not a replacement top-level envelope.

**NF-REQ-019**
A Network Flow response MUST NOT expose storage table names, projection-table names, graph-engine implementation names, temporary paths, import-worker internals, object-store keys, raw CSV rows, or raw CSV cells outside the route-local data members explicitly defined in this NLSpec.

### 5.3 Common request admission

**NF-REQ-020**
Every Network Flow public route with a request body MUST apply `network_flow_json_admission_v1` before route-specific validation.

```text
network_flow_json_admission_v1(request):
  1. If the route declares no body, any non-empty body is invalid.
  2. If the route declares a body, request bytes MUST decode as UTF-8 JSON without replacement characters.
  3. The decoded top-level value MUST be a JSON object.
  4. Duplicate object member names at any depth are invalid before schema validation.
  5. Trailing non-whitespace tokens after the top-level JSON value are invalid.
  6. Explicit JSON null is invalid unless the member table declares nullable=true.
  7. Omitted optional members materialize the declared default.
  8. Unknown object members at any depth are invalid unless a member table declares an extension container. This revision declares no public request extension container.
  9. Admission failures MUST NOT allocate durable state, idempotency records, graph state, tables, rows, diagnostics, or indicator bindings.
```

**NF-REQ-021**
Every Network Flow route that admits query parameters MUST define the complete query parameter set in its route table. Unknown query parameters MUST fail with `network_flow_invalid_request`.

### 5.4 Authorization matrix

**NF-REQ-022**
Network Flow routes MUST rederive authorization at route time from the current session, incident membership, incident lifecycle, incident role, and extension claim state.

**Table 5-A. Route-family authorization matrix**

| Operation family | Minimum incident role | Additional rule |
| --- | --- | --- |
| Source-profile discovery | `viewer` | Extension must be claimed. |
| Effective-limit discovery | `viewer` | Extension must be claimed. |
| Table list/read | `viewer` | Table must be active unless the route explicitly names another state. This revision defines no deleted-table inspection route. |
| Accepted row query | `viewer` | Every table in scope must be active and in the same incident. |
| Rejected-row diagnostic query | `viewer` | Diagnostics are incident content and require current incident visibility. |
| Graph query | `viewer` | Every table in scope must be active and in the same incident. |
| Import target apply through Core import | `editor` | Core import admission still governs upload/session/apply. |
| Table rename | `editor` | Table must be active and version must match. |
| Table soft delete | `reviewer` | Table must be active and version must match. |
| Indicator link to existing indicator | `editor` | Core indicator target must be visible in the same incident. |
| Indicator create from flow value | `editor` | Core indicator creation owner governs canonical indicator creation and dedupe. |

**NF-REQ-023**
A caller with `deployment_admin` and no current incident membership MUST NOT list, query, import, graph, soft-delete, rename, or link Network Flow incident data.

### 5.5 Idempotency and mutation replay

**NF-REQ-024**
Every Network Flow mutating route MUST require `client_txn_id` and MUST be idempotent within this tuple:

```text
(actor_user_id, incident_id, route_id, path_identity, client_txn_id)
```

`path_identity` and normalized request comparison MUST use Table 5-A.

**Table 5-A. Mutation idempotency comparison registry**

| Route ID | `path_identity` | Normalized comparison body | Replay point |
| --- | --- | --- | --- |
| `nf.tables.patch` | `network_flow_table_id:{network_flow_table_id}` | Object containing exactly `base_table_version` and normalized `display_name`; excludes `client_txn_id` and path members. | After admission, authorization, extension availability, and path syntax validation; before current table-version comparison. |
| `nf.tables.delete` | `network_flow_table_id:{network_flow_table_id}` | Object containing exactly `base_table_version`; excludes `client_txn_id` and path members. | After admission, authorization, extension availability, and path syntax validation; before current table-version comparison. |
| `nf.indicator_links.create` | `indicator-links` | Object containing exactly normalized `selector`, normalized `target`, `observation_mode`, and normalized `confirm_exact_value`; excludes `client_txn_id`. | After admission, authorization, and extension availability; before selector resolution, target visibility checks, duplicate-binding lookup, or Core indicator creation. |
| Core import apply for `target_kind='network_flow_table'` | Core-owned import apply idempotency scope. | Core import apply normalized request. Network Flow MUST NOT add a second extension-local idempotency key for the same apply action. | Core import owner replays the apply result; Network Flow MUST return the same created table references for the replayed terminal result. |

**NF-REQ-025**
For each Network Flow-owned mutating route, the implementation MUST compute `network_flow_mutation_request_digest_v1` over the Table 5-A normalized comparison body after defaults are materialized and before mutation commit. The digest input MUST exclude `client_txn_id`, path parameters, actor identity, incident ID, route ID, and any server-computed response members.

```text
network_flow_mutation_request_digest_v1(input):
  UTF8("cartulary.network_flow.mutation_request_digest.v1") NUL
  UTF8(route_id) NUL
  UTF8(path_identity) NUL
  network_flow_canonical_json_v1(normalized_body) NUL
  return lowercase_hex(SHA256(bytes))
```

**NF-REQ-026**
If a mutation with the same idempotency tuple and same request digest has already committed, the route MUST return the original success response and MUST NOT perform the mutation again. Exact committed replay MUST occur at the replay point in Table 5-A even when the current resource version or lifecycle state has since changed. If the tuple matches but the request digest differs, the route MUST fail with Core error code `client_txn_conflict` and MUST NOT perform the mutation.

**NF-REQ-027**
Admission failures, authorization failures, path-parameter failures, stale-version failures, semantic validation failures, and limit failures MUST NOT create committed Network Flow idempotency records unless the owning Core idempotency contract requires retaining failed mutation attempts. If Core requires retaining failed attempts, the retained failure MUST NOT later replay as success.

## 6. Concepts, identifiers, scalar contracts, canonicalization, and digests

### 6.1 Concept registry

**NF-REQ-028**
The concepts in Table 6-A have the exact meanings shown.

**Table 6-A. Concept registry**

| Concept | Meaning |
| --- | --- |
| `network_analysis_workspace` | Incident-scoped extension workspace rendered as the `Network Analysis` top-level extension tab. |
| `network_flow_table` | One imported, incident-scoped, normalized flow table displayed as one inner table tab inside Network Analysis. |
| `network_flow_table_id` | Stable server-generated identifier for a `network_flow_table`. It remains stable across rename, filter, graph, link, query, and table-tab switches. |
| `network_flow_row` | One accepted normalized flow row within exactly one `network_flow_table`. It is immutable after import commit. |
| `network_flow_row_id` | Stable row identifier for one accepted row. It is not a Core `record_id`. |
| `network_flow_rejected_row` | Source row that failed validation and did not enter the accepted flow-row set. |
| `network_flow_endpoint` | Canonical endpoint value derived from accepted flow-row fields. In this revision, the only endpoint kind is IP address. |
| `network_flow_graph_query` | Request state that selects one or more flow tables plus field filters, time filters, and graph aggregation options. |
| `network_flow_flow_edge` | Directed aggregated relationship from one source endpoint to one destination endpoint in a Network Flow graph. |
| `network_flow_indicator_binding` | Explicit user-created binding connecting a flow value or selected flow rows to a Core indicator. |
| `source_profile` | Named external-format semantic profile that supplies required fields, aliases, transforms, empty-value defaults, and validation rules for import mapping. |
| `parser_profile` | Named CSV syntax profile that controls decoding, header handling, quoting, record splitting, and row-locator behavior. |
| `mapping_fingerprint` | Deterministic digest of the approved source-column-to-field mapping and its import context. |
| `graph_query_digest` | Deterministic digest of a default-materialized graph query request. |

### 6.2 Identifier contracts

**NF-REQ-029**
Network Flow Activity identifiers MUST satisfy Table 6-B.

**Table 6-B. Identifier contracts**

| Identifier | Contract | Generation rule |
| --- | --- | --- |
| `network_flow_table_id` | `nft_` followed by 32 lowercase hexadecimal characters. | Server-generated at first table allocation from CSPRNG output. Replay of the same committed import apply action MUST return the original ID. |
| `network_flow_row_id` | `nfr_` followed by 64 lowercase hexadecimal characters. | Deterministic `network_flow_row_id_v1` in §6.5. |
| `network_flow_endpoint_id` | `nfe_` followed by 64 lowercase hexadecimal characters. | Deterministic `network_flow_endpoint_id_v1` in §14.3. |
| `network_flow_flow_edge_id` | `nff_` followed by 64 lowercase hexadecimal characters. | Deterministic `network_flow_flow_edge_id_v1` in §14.3. |
| `network_flow_indicator_binding_id` | `nfb_` followed by 32 lowercase hexadecimal characters. | Server-generated at first committed binding action from CSPRNG output. Replay of the same committed binding action MUST return the original ID. |
| `network_flow_graph_query_digest` | 64 lowercase hexadecimal characters. | Deterministic `network_flow_graph_query_digest_v1` in §6.8. |
| `network_flow_source_snapshot_id` | `nfsnap_` followed by 64 lowercase hexadecimal characters. | Deterministic `network_flow_source_snapshot_digest_v1` in §6.9. |

**NF-REQ-030**
`network_flow_table_id`, `network_flow_row_id`, `network_flow_endpoint_id`, `network_flow_flow_edge_id`, `network_flow_indicator_binding_id`, and `network_flow_source_snapshot_id` MUST NOT be accepted as Core `record_id` values. Core record mutation, view-row, revision, rollback, merge, and saved-view routes MUST NOT treat these identifiers as Core record identifiers.

### 6.3 Canonical JSON

**NF-REQ-031**
`network_flow_canonical_json_v1` MUST serialize JSON-compatible values using Table 6-C.

**Table 6-C. Canonical JSON serialization**

| Aspect | Required behavior |
| --- | --- |
| Input precondition | Input is an already-decoded JSON value with no duplicate object members. |
| Encoding | UTF-8 without BOM. |
| Whitespace | No insignificant whitespace. |
| Closed object member order | Schema-table order declared by this NLSpec. |
| Dynamic object member order | Exact Unicode code point order of member names. |
| Arrays | Preserve the order defined by the owning schema. |
| Strings | JSON string syntax using the shortest required escapes for quotation mark, reverse solidus, and control characters; non-control Unicode scalar values are emitted as UTF-8 rather than escaped. |
| Integers | Base-10 ASCII, no leading zeroes except `0`. |
| Decimal counters | JSON strings under `uint64_decimal_string_v1`; never JSON numbers. |
| Floating point | Invalid in all Network Flow canonical objects. |
| Booleans and null | Exact JSON literals `true`, `false`, and `null`. |

**NF-REQ-032**
Every canonical JSON object used for a digest MUST be normalized before serialization. Omitted optional members that have defaults MUST be materialized before digesting. Members whose omission behavior is true omission MUST remain omitted and MUST NOT serialize as JSON `null`.

### 6.4 Scalar contracts and canonicalization algorithms

**NF-REQ-033**
The scalar contracts in Table 6-D apply wherever referenced.

**Table 6-D. Scalar contracts**

| Contract | Definition |
| --- | --- |
| `timestamp_utc_v1` | JSON string in UTC RFC3339 form with `Z` suffix, year `0001..9999`, no leap seconds, and at most 6 fractional decimal digits. The canonical output omits a fractional part for whole-second values and otherwise trims trailing fractional zeroes. |
| `ip_literal_v1` | Valid IPv4 dotted-decimal or IPv6 literal. IPv4 octal, hexadecimal, plus signs, leading-zero octets, and IPv6 zone identifiers are invalid. IPv6 storage uses lowercase RFC 5952-style compressed canonical text. IPv4-mapped IPv6 addresses remain IPv6. |
| `cidr_literal_v1` | IP literal plus prefix length. Host bits are canonicalized to the network address. IPv4 prefix range is `0..32`; IPv6 prefix range is `0..128`. |
| `port_number_v1` | JSON integer or decoded decimal source string representing `0..65535`; no sign, exponent, decimal point, leading plus sign, or leading zeroes except `0`. |
| `ip_protocol_number_v1` | JSON integer or decoded decimal source string representing `0..255`; protocol tokens map only through Table 9-E. |
| `uint64_decimal_string_v1` | Decimal string for an unsigned integer in `0..18446744073709551615` with no sign, decimal point, exponent, leading plus sign, or leading zeroes except `0`. Comparisons use integer arithmetic. |
| `bounded_text_256_v1` | JSON string with `0..256` Unicode scalar values, no NUL, and no C0 or C1 controls except tab only where a source profile explicitly permits it. |
| `bounded_text_1024_v1` | JSON string with `0..1024` Unicode scalar values, no NUL, and no C0 or C1 controls except tab and line break only where a source profile explicitly permits them. |
| `sha256_hex_v1` | Exactly 64 lowercase hexadecimal characters. |
| `client_txn_id` | Core route-scoped idempotency key. This NLSpec does not redefine the Core `client_txn_id` contract. |

**NF-REQ-034**
`timestamp_utc_v1` parsing MUST reject leap seconds, local timestamps without an applicable timestamp profile timezone, timezone names embedded in source values outside the selected timestamp-profile schema, DST folds, DST gaps, and values outside year `0001..9999`.

**NF-REQ-035**
`ip_literal_v1` comparison MUST compare IP family first with IPv4 before IPv6, then compare unsigned address bytes within the same family. Textual IP comparison MUST NOT use locale collation, database collation, or lexical string ordering.

**NF-REQ-036**
`bounded_text_256_v1` and `bounded_text_1024_v1` preserve decoded source field text for row provenance. Display names, mapping suggestions, and user-entered non-source labels MUST use Unicode NFC normalization and trim leading/trailing Unicode whitespace where their owning algorithm states so.

### 6.5 Row digest and row ID algorithms

**NF-REQ-037**
`source_row_digest_v1` MUST compute the lowercase SHA-256 digest of this byte stream.

```text
source_row_digest_v1(input):
  UTF8("cartulary.network_flow.source_row_digest.v1") NUL
  UTF8(input.parser_profile_id) NUL
  UTF8(decimal(input.source_row_number)) NUL
  UTF8(decimal(input.field_count)) NUL
  for each decoded field in source order:
    UTF8(decimal(source_column_ordinal)) NUL
    UTF8(decimal(byte_length(UTF8(decoded_field)))) NUL
    UTF8(decoded_field) NUL
  return lowercase_hex(SHA256(bytes))
```

**NF-REQ-038**
`normalized_row_digest_v1` MUST compute the lowercase SHA-256 digest of this byte stream.

```text
normalized_row_digest_v1(input):
  UTF8("cartulary.network_flow.normalized_row_digest.v1") NUL
  UTF8(input.mapping_fingerprint) NUL
  for each registered field_key in Table 9-C order:
    UTF8(field_key) NUL
    UTF8(value_presence_token) NUL
    UTF8(canonical_scalar_or_null) NUL
  UTF8("unmapped_raw") NUL
  network_flow_canonical_json_v1(unmapped_raw) NUL
  return lowercase_hex(SHA256(bytes))
```

**NF-REQ-039**
Diagnostics MUST NOT participate in accepted row identity. Accepted row identity depends on incident ID, table ID, source row locator, decoded source row, mapping fingerprint, and accepted normalized values.

**NF-REQ-040**
`network_flow_row_id_v1` MUST compute the row identifier using this byte input and output rule.

```text
network_flow_row_id_v1(input):
  UTF8("cartulary.network_flow_row_id.v1") NUL
  UTF8(input.incident_id) NUL
  UTF8(input.network_flow_table_id) NUL
  UTF8(decimal(input.source_row_number)) NUL
  UTF8(input.source_row_digest_sha256) NUL
  UTF8(input.normalized_row_digest_sha256) NUL
  return "nfr_" + lowercase_hex(SHA256(bytes))
```

**NF-REQ-041**
Duplicate normalized rows MUST remain distinct when their source row locators differ. The implementation MUST NOT collapse duplicate flow rows unless a later revision defines a row-deduplication operation and its provenance consequences.

### 6.6 Mapping fingerprint algorithm

**NF-REQ-042**
`mapping_fingerprint_v1` MUST compute the lowercase SHA-256 digest of this canonical byte stream.

```text
mapping_fingerprint_v1(input):
  UTF8("cartulary.network_flow_mapping_fingerprint.v1") NUL
  UTF8(target_kind) NUL
  UTF8(target_table_schema_id) NUL
  UTF8(source_profile_id) NUL
  UTF8(parser_profile_id) NUL
  UTF8(source_content_sha256) NUL
  UTF8(unknown_column_policy) NUL
  network_flow_canonical_json_v1(timestamp_profile) NUL
  for each source column in source_column_ordinal ASC:
    UTF8(decimal(source_column_ordinal)) NUL
    UTF8(raw_header_text) NUL
    UTF8(normalized_header_for_suggestion) NUL
    UTF8(raw_header_sha256) NUL
  for each field mapping sorted by mapping_sort_key_v1:
    network_flow_canonical_json_v1(field_mapping) NUL
  return lowercase_hex(SHA256(bytes))
```

`mapping_sort_key_v1` MUST be:

| Mapping kind | Sort tuple |
| --- | --- |
| `source_column` | `(field_key, "source_column", source_column_ordinal, "", "")` |
| `system_derivation` | `(field_key, "system_derivation", 0, derivation_id, "")` |
| `ignored_source_column` | `("", "ignored_source_column", source_column_ordinal, "", ignore_reason)` |

**NF-REQ-043**
The mapping fingerprint MUST change when any source column ordinal, raw header, source profile, parser profile, source content hash, target field key, mapping kind, transform, empty-value policy, timestamp profile, unknown-column policy, derivation, ignore reason, combinability, or source-column descriptor changes. Header text alone MUST NOT identify a column when duplicate headers exist.

### 6.7 Safe digest algorithm

**NF-REQ-044**
`network_flow_safe_digest_v1` MUST compute a keyed safe digest for logs, telemetry, and administrative audit summaries.

```text
network_flow_safe_digest_v1(value_class, canonical_value):
  HMAC-SHA256(
    deployment_audit_secret,
    UTF8("cartulary.network_flow.safe_digest.v1") NUL
    UTF8(value_class) NUL
    UTF8(canonical_value)
  )
  return lowercase_hex(hmac)
```

**NF-REQ-045**
`deployment_audit_secret` MUST be deployment-local secret material with at least 256 bits of CSPRNG entropy. It MUST NOT be exported through public routes, logs, telemetry, fixtures, import diagnostics, or Graph Projection metadata.

### 6.8 Graph query digest algorithm

**NF-REQ-046**
`network_flow_graph_query_digest_v1` MUST compute the lowercase SHA-256 digest of this byte stream after defaults are materialized and request objects are normalized.

```text
network_flow_graph_query_digest_v1(input):
  UTF8("cartulary.network_flow.graph_query_digest.v1") NUL
  UTF8(incident_id) NUL
  network_flow_canonical_json_v1(normalized_table_scope) NUL
  network_flow_canonical_json_v1(normalized_filters) NUL
  network_flow_canonical_json_v1(normalized_time_range_or_null) NUL
  network_flow_canonical_json_v1(normalized_aggregation) NUL
  return lowercase_hex(SHA256(bytes))
```

**NF-REQ-047**
Omitted `filters[]` and supplied empty `filters[]` MUST produce the same graph query digest. Omitted `aggregation` and explicit default aggregation from Table 14-C MUST produce the same graph query digest. Omitted `time_range` MUST serialize as literal JSON `null`. `limit_overrides` never participates in `network_flow_graph_query_digest_v1`; changing deployment limit configuration or supplying a lower limit override MUST NOT change the graph query digest.

For digest purposes, `normalized_table_scope` MUST be the object `{ "table_ids": [...] }`, where `table_ids` is the resolved active table set sorted by `network_flow_table_id ASC`. `active_table`, `selected_tables`, and `all_active_tables` requests that resolve to the same active table set MUST produce the same `normalized_table_scope` and the same graph query digest. Adding a new active table or soft-deleting an active table MUST change the digest for a later `all_active_tables` request when the resolved table set changes.

### 6.9 Source snapshot digest algorithm

**NF-REQ-047a**
`network_flow_source_snapshot_digest_v1` MUST compute the Network Flow source snapshot identifier submitted to the Graph Projection boundary.

```text
network_flow_source_snapshot_digest_v1(input):
  UTF8("cartulary.network_flow.source_snapshot_digest.v1") NUL
  UTF8(incident_id) NUL
  UTF8(decimal(count(selected_tables))) NUL
  for each selected table ordered by network_flow_table_id ASC:
    UTF8(network_flow_table_id) NUL
    UTF8(mapping_fingerprint) NUL
  UTF8(graph_query_digest) NUL
  return "nfsnap_" + lowercase_hex(SHA256(bytes))
```

**NF-REQ-047b**
`network_flow_source_snapshot_digest_v1` MUST exclude table display names, `table_version`, `limit_overrides`, selected row digest sets, Graph Projection run identifiers, Graph Projection retention settings, and graph layout state. A table rename MUST NOT change `source_snapshot_id`.

## 7. Network Analysis workspace and UI semantic state

### 7.1 Workspace regions

**NF-REQ-048**
When the `network_flow_activity` extension profile is claimed and the caller has current incident visibility, the workbook shell MUST expose the `Network Analysis` top-level extension tab.

**NF-REQ-049**
The `Network Analysis` workspace MUST contain the regions in Table 7-A.

**Table 7-A. Network Analysis workspace regions**

| Region | Required behavior |
| --- | --- |
| Workspace header | Displays `Network Analysis`, table count, active table name when one table is selected, and import action when the caller has the required role. |
| Inner table tab strip | Displays one tab per active `network_flow_table`. Ordering is `created_at ASC`, then `network_flow_table_id ASC`. |
| Flow table grid | Displays accepted rows for the active table using `field_key`-bound columns. |
| Diagnostics summary | Displays accepted-row count, rejected-row count, mapping profile, parser profile, diagnostics truncation state, and source filename display value. |
| Filter selector | Displays table scope, time filters, endpoint filters, protocol/counter filters, and advanced field filters. In graph mode, table-selection controls are mandatory. |
| Graph panel | Displays the derived graph for the current `network_flow_graph_query`. It is unavailable until at least one active table exists. |
| Contributing rows drawer | Displays graph vertex or edge contributors grouped by table when the user pivots from graph to rows. |
| Indicator-link affordance | Starts explicit linking from selected endpoint, row value, graph vertex, or graph edge. |
| Status strip contribution | Emits exactly one effective Network Flow semantic state chosen by Table 7-B precedence when any Network Flow state applies. |

**NF-REQ-050**
If no active flow table exists, Network Analysis MUST render an empty state with one primary action labeled `Import NetFlow CSV` when the caller may import and a non-mutating explanation when the caller may not import. The empty state MUST NOT render an empty graph.

**NF-REQ-051**
Workspace behavior MUST preserve the workbook-first interaction model. Routine table inspection, filtering, graphing, graph-to-table pivot, table rename, table soft delete, and indicator-link initiation MUST occur inside the Network Analysis workspace rather than through a detached administration context.

**NF-REQ-052**
Network Analysis MUST NOT use visible tab labels, screen coordinates, graph layout positions, row indexes, grid vendor coordinates, or rendered order as authoritative identity. It MUST use stable identifiers and `field_key` values for every public action.

### 7.2 UI semantic state registry

**NF-REQ-053**
When multiple Network Flow workspace states apply, the status strip MUST render the lowest numeric precedence in Table 7-B.

**Table 7-B. Network Flow status-strip precedence**

| Precedence | State family | Tokens | Trigger | Omission behavior |
| ---: | --- | --- | --- | --- |
| 1 | Blocking import/mapping | `mapping_required`, `validation_failed` | Required mapping, timestamp profile, transform, or validation failure blocks apply. | If no blocking condition exists, no token from this family is emitted. |
| 2 | Active work | `validating`, `graph_pending`, `link_pending` | Preview validation, graph request, or link request is in progress. | If no active work exists, no token from this family is emitted. |
| 3 | Stale derived state | `graph_stale` | Current graph output no longer matches current table scope, filters, time range, aggregation, or table lifecycle. | If no graph output exists, `graph_stale` is not emitted. |
| 4 | Loaded data | `loaded_with_rejections`, `loaded` | Active table data is loaded. Use `loaded_with_rejections` when the active table has `row_count_rejected > 0`. | If no active table exists, no token from this family is emitted. |
| 5 | Transient success | `graph_available`, `link_committed` | Graph result returned or link mutation committed. | Transient token is suppressed when any higher-precedence state applies. |

**NF-REQ-054**
`link_committed` MUST persist for 5 seconds or until the next user action, whichever occurs first. `graph_available` MUST persist until query state changes or graph becomes stale. `graph_stale` MUST persist until graph revalidation, graph panel close, or graph query reset.

## 8. Table-tab lifecycle and table registry

### 8.1 Lifecycle states

**NF-REQ-055**
Each `network_flow_table` MUST use exactly one lifecycle state from Table 8-A.

**Table 8-A. Table lifecycle state machine**

| State | Meaning | Queryable | Graphable | Listed by default | Terminal | Allowed transitions |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `creating` | Import apply has allocated table metadata, but validation and commit are not terminal. | No | No | No | No | `active`, `failed` |
| `active` | Accepted rows are queryable and graphable. | Yes | Yes | Yes | No | `soft_deleted` |
| `soft_deleted` | Table is hidden from default workspace use but retained for audit, provenance, and binding traceability. | No | No | No | Yes | none |
| `failed` | Table allocation failed before accepted rows became queryable. | No | No | No | Yes | none |

**NF-REQ-056**
`renamed` MUST NOT be a lifecycle state. A table rename is a metadata mutation and audit event only.

**NF-REQ-057**
A table in `creating`, `failed`, or `soft_deleted` state MUST NOT be returned by ordinary table query, row query, graph query, or default table list routes. This revision defines no deleted-table inspection route and no non-active table listing mode.

**NF-REQ-058**
A table in `soft_deleted` state MUST NOT appear in the default inner tab strip, MUST NOT be included by `table_scope.mode='all_active_tables'`, and MUST fail table, row, rejected-row, graph, and indicator-link route admission with `network_flow_table_not_active` when the caller references it directly.

Direct references to non-active table IDs MUST use Table 8-B1 after authorization and hidden-resource checks.

**Table 8-B1. Direct table-reference state failures**

| Referenced table state | Route-local error |
| --- | --- |
| `soft_deleted` | `network_flow_table_not_active` |
| `creating` | `network_flow_table_not_found` |
| `failed` | `network_flow_table_not_found` |

### 8.2 Table creation from import

**NF-REQ-059**
A successful import apply for one selected CSV import unit with `target_kind='network_flow_table'` MUST create exactly one `network_flow_table` when row validation admits at least one accepted row.

**NF-REQ-060**
A partially valid CSV import MUST create one table when at least one row is accepted. Rejected rows MUST be retained as diagnostics attached to that table subject to diagnostic truncation rules.

**NF-REQ-061**
An all-invalid CSV import MUST create no table and MUST fail that import unit with `network_flow_all_rows_rejected`. The error details MUST include `row_count_rejected`, `diagnostics_truncated`, and `diagnostics_sample[]` under Table 21-B. `diagnostics_sample[]` MUST contain the first `min(50, diagnostic_count)` diagnostics under §12.4 ordering using the Table 12-E shape.

**NF-REQ-062**
A CSV import with no data rows after its header MUST create no table and MUST fail that import unit with `network_flow_no_data_rows`.

**NF-REQ-063**
After a successful import creates one or more tables, the client MUST select the first newly created table in import-apply result order as the active inner table tab. If the current caller loses access before the table is displayed, the workspace MUST revalidate visibility and render the authorized state instead of showing stale table contents.

### 8.3 Table registry resource

**NF-REQ-064**
The `network_flow_table` resource MUST expose the fields in Table 8-B.

**Table 8-B. `network_flow_table` resource**

| Field | Type | Required | Mutable | Rule |
| --- | --- | ---: | ---: | --- |
| `network_flow_table_id` | Identifier | Yes | No | Stable table identity. |
| `incident_id` | Core incident ID | Yes | No | Owning incident. |
| `display_name` | `bounded_text_256_v1` | Yes | Yes | Mutated only by table rename route. |
| `table_version` | Positive integer | Yes | Yes | Starts at `1` when table becomes active; increments by `1` on rename and soft delete. |
| `table_status` | Lifecycle token | Yes | Yes | Closed vocabulary from Table 8-A. |
| `source_import_session_id` | Import session ID | Yes | No | Source import session. |
| `source_import_unit_id` | Import unit ID | Yes | No | Source import unit. |
| `source_content_sha256` | `sha256_hex_v1` | Yes | No | Exact uploaded source bytes hash from import route. |
| `source_filename_display` | `bounded_text_256_v1` | Yes | No | Safe display name derived from source filename; no path segments. |
| `source_filename_digest` | safe digest | Yes | No | `network_flow_safe_digest_v1("source_filename", source_filename_display)`. |
| `mapping_fingerprint` | `sha256_hex_v1` | Yes | No | Approved mapping fingerprint. |
| `source_profile_id` | Source profile token | Yes | No | Declared source profile. |
| `parser_profile_id` | Parser profile token | Yes | No | Declared parser profile. |
| `row_count_accepted` | Non-negative integer | Yes | No | Count of accepted rows. |
| `row_count_rejected` | Non-negative integer | Yes | No | Count of rejected rows. |
| `diagnostics_truncated` | Boolean | Yes | No | `true` only when rejected-row diagnostics exceeded retention limit. |
| `created_by_user_id` | User ID | Yes | No | Import actor. |
| `created_at` | `timestamp_utc_v1` | Yes | No | Creation time. |
| `updated_at` | `timestamp_utc_v1` | Yes | Yes | Updated by table metadata mutation. |
| `deleted_at` | `timestamp_utc_v1` or null | Yes | Yes | Null unless `table_status='soft_deleted'`. |

### 8.4 Display-name algorithm

**NF-REQ-065**
The `source_filename_display` value on a table MUST be derived by `sanitize_source_filename_display_v1` from the Core import filename hint. If Core supplies no filename hint, the input to the algorithm is the empty string.

```text
sanitize_source_filename_display_v1(filename_hint):
  value = Unicode_NFC(filename_hint)
  value = replace every "\" with "/"
  segments = split value on "/"
  candidate = last segment whose scalar length is greater than 0, or "uploaded.csv"
  candidate = remove_NUL_C0_C1_controls(candidate)
  candidate = trim_unicode_whitespace(candidate)
  if candidate == "":
    candidate = "uploaded.csv"
  return first_256_unicode_scalars(candidate)
```

`source_filename_digest` MUST be `network_flow_safe_digest_v1("source_filename", source_filename_display)`. The implementation MUST NOT preserve directory segments, drive prefixes, UNC prefixes, object-store keys, temporary paths, or upload-worker paths in `source_filename_display`.

**NF-REQ-065a**
The default table display name MUST be derived by `derive_table_display_name_v1`.

```text
derive_table_display_name_v1(original_filename, existing_active_display_names):
  source_display = sanitize_source_filename_display_v1(original_filename)
  stem = filename_stem_after_path_stripping_v1(source_display)
  candidate = normalize_table_display_name_input_v1(stem)
  if candidate == "":
    candidate = "Imported NetFlow"
  candidate = first_64_unicode_scalars(candidate)
  if candidate not in existing_active_display_names:
    return candidate
  for n in 2..9999:
    suffix = " (" + decimal(n) + ")"
    base = first_(64 - scalar_length(suffix))_unicode_scalars(candidate)
    suffixed = base + suffix
    if suffixed not in existing_active_display_names:
      return suffixed
  fail network_flow_table_name_exhausted
```

```text
filename_stem_after_path_stripping_v1(source_filename_display):
  if source_filename_display has more than one scalar and the final scalar is ".":
    return all but the final scalar
  let last_dot be the final "." scalar position, if any
  if last_dot exists and last_dot is not the first scalar and not the final scalar:
    return all scalars before last_dot
  return source_filename_display

normalize_table_display_name_input_v1(value):
  value = Unicode_NFC(value)
  value = remove_NUL_C0_C1_controls(value)
  value = trim_unicode_whitespace(value)
  return value
```

Examples: `C:\tmp\flows.csv` and `/tmp/flows.csv` both produce source filename display `flows.csv` and default display name `flows`; `.csv` produces default display name `.csv`; `file.` produces default display name `file`.

**NF-REQ-066**
A table rename MUST be a metadata-only mutation on an `active` table. It MUST increment `table_version`, update `display_name` and `updated_at`, emit `network_flow_table_renamed`, and MUST NOT change table ID, row IDs, source provenance, mapping fingerprint, graph identity, live query cursor validity, diagnostics, or indicator bindings.

The supplied `display_name` MUST normalize through `normalize_table_display_name_input_v1`. The normalized value MUST be non-empty, MUST be at most 64 Unicode scalars, and MUST be unique among other `active` tables in the same incident by exact code-point comparison. The rename route MUST NOT truncate an explicit display name. A duplicate active-table display name MUST fail with `network_flow_invalid_display_name` and `reason_code='duplicate_display_name'`. An empty normalized display name MUST fail with `reason_code='empty_display_name'`. A value longer than 64 Unicode scalars after normalization MUST fail with `reason_code='display_name_too_long'`. Soft-deleted names do not reserve display names. Concurrent rename and import commits MUST be serialized so two committed active tables in the same incident cannot share the same normalized display name.

### 8.5 Row immutability and retained table consequences

**NF-REQ-067**
A `network_flow_row` MUST be immutable after import commit. A correction requires importing a corrected source as a new table or soft-deleting the incorrect table and importing again. This revision defines no individual row edit route.

**NF-REQ-068**
A `network_flow_table` is a source-data container for analytical rows. It is not a saved view and MUST NOT persist user query state, graph layout state, selected row state, scroll state, or open-drawer state as table metadata.

**NF-REQ-069**
Table lifecycle states MUST count against active and retained table limits according to Table 8-C. A new import of the same source bytes and mapping is allowed after soft delete unless blocked by route-scoped idempotency for the same committed import apply action.

**Table 8-C. Table limit accounting by lifecycle state**

| Table state | Counts against `network_flow.max_active_tables_per_incident` | Counts against `network_flow.max_retained_tables_per_incident` |
| --- | ---: | ---: |
| `creating` | Yes | Yes |
| `active` | Yes | Yes |
| `soft_deleted` | No | Yes |
| `failed` | No | No |

## 9. External format profiles, CSV parser, field registry, and timestamp profiles

### 9.1 Profile registry

**NF-REQ-070**
The extension MUST implement the profile registry in Table 9-A.

**Table 9-A. External format profile registry**

| Profile ID | Classification | Conformance status | Required behavior |
| --- | --- | --- | --- |
| `rfc4180_headered_csv_v1` | CSV parser profile | `required_v1` | Default parser profile for Network Flow CSV. It uses RFC 4180 as an informational interchange baseline but closes behavior through §9.2. |
| `cisco_sna_netflow_csv_v1` | Semantic source profile | `required_v1` | Required vendor interop fixture profile and standard example profile. |
| `ipfix_csv_projection_v1` | Semantic source profile | `reserved_future` | Not claimable in v1 unless this document adds exhaustive alias, unit, requiredness, and fixture tables. |
| `netflow_v9_csv_projection_v1` | Semantic source profile | `reserved_future` | Not claimable in v1 unless this document adds exhaustive alias, unit, requiredness, and fixture tables. |

**NF-REQ-071**
A route, mapping modal, import apply, fixture, or conformance manifest MUST NOT claim support for a `reserved_future` profile. A request naming a reserved profile MUST fail with `network_flow_unsupported_source_profile`.

**NF-REQ-072**
A format profile version is immutable. If a later Cisco, IPFIX, NetFlow, or CSV interpretation changes required aliases, transforms, required fields, or validation outcomes, the implementation MUST introduce a new profile ID rather than silently changing the behavior of an existing profile ID.

### 9.2 CSV parser profile

**NF-REQ-073**
`rfc4180_headered_csv_v1` MUST decode CSV using Table 9-B.

**Table 9-B. CSV decoding contract**

| Concern | Required behavior |
| --- | --- |
| Encoding | UTF-8 only. A UTF-8 BOM is accepted only at byte offset `0` and is stripped before header parsing. Invalid UTF-8 fails the import unit with `network_flow_invalid_utf8`. |
| Empty file | Fail import unit with `network_flow_csv_empty_file`; allocate no table. |
| Header | A header record is required. The header is source record number `1`. Missing header fails with `network_flow_no_header_row`. |
| Header-only file | Fail import unit with `network_flow_no_data_rows`; allocate no table. |
| Duplicate headers | Duplicate header text is allowed. Source columns are identified by `(source_column_ordinal, raw_header_text)`, not by text alone. |
| Empty header cell | Allowed. The mapping UI MUST display an ordinal-qualified placeholder for an empty header. |
| Record delimiter | CRLF, LF, and CR are accepted outside quoted fields. Quoted line breaks are part of the field value. |
| Terminal newline | Does not create an extra empty record. |
| Blank physical line outside quotes | Treat as a CSV record containing one empty field. If header column count is not one, reject that record as `network_flow_csv_field_count_mismatch`. |
| Field delimiter | Comma. No alternate delimiter is valid in this parser profile. |
| Spaces around delimiters | Preserved as field content. No trimming occurs unless a transform declares it. |
| Quote opening | A field is quoted only when the first byte of the field is `"`. Whitespace before a quote makes the quote literal. |
| Quote escaping | `""` inside a quoted field decodes to `"`. |
| Closing quote | After a closing quote, the next byte must be comma, record delimiter, or EOF. Any other byte fails the import unit with `network_flow_csv_malformed_quote`. |
| Unclosed quote | Fail the import unit with `network_flow_csv_malformed_quote`; allocate no table. |
| Field-count mismatch | A data record whose field count differs from the header field count is rejected as `network_flow_csv_field_count_mismatch`; valid records continue unless all rows reject. |
| Source row locator | `source_row_number` is the 1-based logical CSV record number. A multi-line quoted field still belongs to one source row number. |
| Empty field | Empty string remains empty until a target field transform defines null conversion. |
| Formula-looking value | Treated as inert text during import. The implementation MUST NOT evaluate formulas, hyperlinks, macros, external references, or spreadsheet expressions. |
| Preview row order | Source order by `source_row_number ASC`. |
| Preview size | At most 50 data rows. |
| Limit timing | Column/header limits after header decode; raw cell scalar limit after field decode; row limit while streaming; accepted-row limit during validation. |

**NF-REQ-074**
The parser MUST retain source column descriptors for every header column in source order. Each descriptor MUST satisfy `source_column_v1` in §10.2.

### 9.3 Field requirement tokens

**NF-REQ-075**
The field registry MUST NOT use `Recommended` as a normative status. Profile-specific field requirements MUST use exactly one token from Table 9-C.

**Table 9-C. Profile field requirement tokens**

| Token | Meaning |
| --- | --- |
| `required` | Mapping approval fails unless the field is mapped from a source column or system-derived under the source profile. |
| `optional_map_when_present` | If a recognized source alias exists, the mapping UI defaults to mapping it; absence is not a warning and does not block apply. |
| `optional_warn_when_missing` | Absence creates a non-blocking mapping warning. |
| `system_derived` | User cannot map directly; implementation derives the field. |
| `not_supported` | Field is ignored or retained only as unmapped raw. |

### 9.4 Network-flow field registry

**NF-REQ-076**
The v1 normalized target field registry MUST use Table 9-D. UI labels and source aliases MUST map into these stable `field_key` values.

Requirement-token columns are per-profile. A source profile becomes claimable only when Table 9-A registers it and Table 9-D carries its requirement-token column.

**Table 9-D. Network-flow field registry**

| `field_key` | Cisco SNA v1 status | Type and bounds | Normalization | Reject when |
| --- | --- | --- | --- | --- |
| `network_flow.flow_start_utc` | `required` | `timestamp_utc_v1` | Parse under selected timestamp profile and store canonical UTC. | Missing, invalid, or unsupported timestamp. |
| `network_flow.flow_end_utc` | `required` | `timestamp_utc_v1` | Parse under selected timestamp profile and store canonical UTC. | Missing, invalid, unsupported timestamp, or end earlier than start. |
| `network_flow.src_ip` | `required` | `ip_literal_v1` | Canonical IP text. | Missing or invalid IP. |
| `network_flow.dst_ip` | `required` | `ip_literal_v1` | Canonical IP text. | Missing or invalid IP. |
| `network_flow.src_port` | `required` | `port_number_v1` or null | Empty becomes null only when profile permits null. | Required field missing, non-empty invalid integer, or outside range. |
| `network_flow.dst_port` | `required` | `port_number_v1` or null | Empty becomes null only when profile permits null. | Required field missing, non-empty invalid integer, or outside range. |
| `network_flow.ip_protocol` | `required` | `ip_protocol_number_v1` or declared token | Store integer protocol number. | Missing, unknown token, invalid integer, or outside range. |
| `network_flow.bytes_count` | `required` | `uint64_decimal_string_v1` | Store canonical decimal string. | Missing, negative, decimal, exponent, non-numeric, leading zero except `0`, or above max. |
| `network_flow.packets_count` | `required` | `uint64_decimal_string_v1` | Store canonical decimal string. | Missing, negative, decimal, exponent, non-numeric, leading zero except `0`, or above max. |
| `network_flow.exporter_id` | `optional_map_when_present` | `bounded_text_256_v1` or null | Preserve decoded value except profile-declared transform. | Contains NUL/control or exceeds bound. |
| `network_flow.input_interface` | `optional_map_when_present` | `bounded_text_256_v1` or null | For Cisco SNA, apply `trim_ascii_space_v1`, then preserve decoded text. | Contains NUL/control or exceeds bound. |
| `network_flow.output_interface` | `optional_map_when_present` | `bounded_text_256_v1` or null | For Cisco SNA, apply `trim_ascii_space_v1`, then preserve decoded text. | Contains NUL/control or exceeds bound. |
| `network_flow.tcp_flags` | `optional_map_when_present` | Integer bitmask `0..255` or declared token string | Profile-specific canonical form. | Invalid bitmask or unsupported token. |
| `network_flow.application_label` | `optional_map_when_present` | `bounded_text_256_v1` or null | Preserve as source label. | Contains NUL/control or exceeds bound. |
| `network_flow.observation_source_ref` | `system_derived` | Object | Derived from import session, import unit, mapping, and row locator. | Missing after normalization. |

The only claimable v1 profile requires non-null `src_port` and `dst_port`. The null branches of the port contracts, the `is_null` and `not_null` port operators, and null-port edge aggregation are normative but unreachable in claimable v1; they bind future profiles without further amendment.

### 9.5 Protocol token registry

**NF-REQ-077**
Profile-declared protocol tokens MUST map according to Table 9-E. Tokens are compared after ASCII uppercase and trimming ASCII spaces only. Numeric protocol values bypass this table and are validated as `0..255`.

**Table 9-E. Protocol token mapping**

| Source token | Stored `ip_protocol` |
| --- | ---: |
| `TCP` | 6 |
| `UDP` | 17 |
| `ICMP` | 1 |
| `ICMPV6` | 58 |
| `ICMP_V6` | 58 |
| `GRE` | 47 |
| `ESP` | 50 |
| `AH` | 51 |

**NF-REQ-078**
A token not listed in Table 9-E MUST fail with `network_flow_invalid_protocol`. The implementation MUST NOT infer protocol numbers from locale-specific service names, port numbers, or application labels.

### 9.6 Cisco SNA NetFlow profile

**NF-REQ-079**
`cisco_sna_netflow_csv_v1` MUST enforce the field requirements in Table 9-F.

**Table 9-F. Cisco SNA alias mapping**

| Cisco SNA / NetFlow CSV alias | Target `field_key` | Requirement token | Notes |
| --- | --- | --- | --- |
| `Source IP Address`, `Source IP`, `IPV4_SRC_ADDR`, `sourceIPv4Address`, `IPV6_SRC_ADDR`, `sourceIPv6Address` | `network_flow.src_ip` | `required` | IPv4 and IPv6 literals normalize under `ip_literal_v1`. |
| `Destination IP Address`, `Destination IP`, `IPV4_DST_ADDR`, `destinationIPv4Address`, `IPV6_DST_ADDR`, `destinationIPv6Address` | `network_flow.dst_ip` | `required` | Same canonical IP normalization as `network_flow.src_ip`. |
| `Source Port`, `L4_SRC_PORT`, `sourceTransportPort` | `network_flow.src_port` | `required` | Null is invalid for this profile. |
| `Destination Port`, `L4_DST_PORT`, `destinationTransportPort` | `network_flow.dst_port` | `required` | Null is invalid for this profile. |
| `Layer 3 Protocol`, `Protocol`, `PROTOCOL`, `protocolIdentifier` | `network_flow.ip_protocol` | `required` | Accept numeric values and declared tokens in Table 9-E. |
| `Bytes Count`, `Bytes`, `IN_BYTES`, `octetDeltaCount` | `network_flow.bytes_count` | `required` | Decimal-string storage. |
| `Packet count`, `Packets`, `IN_PKTS`, `packetDeltaCount` | `network_flow.packets_count` | `required` | Decimal-string storage. |
| `Flow Start Time`, `First Seen`, `FIRST_SWITCHED`, `flowStartMilliseconds`, `flowStartSeconds` | `network_flow.flow_start_utc` | `required` | Mapping must declare timestamp profile. |
| `Flow End Time`, `Last Seen`, `LAST_SWITCHED`, `flowEndMilliseconds`, `flowEndSeconds` | `network_flow.flow_end_utc` | `required` | End must be greater than or equal to start. |
| `Interface input`, `Input Interface`, `ingressInterface` | `network_flow.input_interface` | `optional_map_when_present` | Mapped when present; absence is not a blocking warning. |
| `Interface output`, `Output Interface`, `egressInterface` | `network_flow.output_interface` | `optional_map_when_present` | Mapped when present; absence is not a blocking warning. |

**NF-REQ-080**
The Cisco SNA profile MUST treat the nine required Cisco fields as required profile fields. It MAY accept additional columns. Omission behavior: additional columns default to `unknown_column_policy='preserve_unmapped_raw'` unless the user maps them to a registered field or explicitly ignores them.

**NF-REQ-080a**
Header alias matching MUST use `source_alias_match_key_v1` for both source headers and profile aliases.

```text
source_alias_match_key_v1(input):
  value = Unicode_NFC(input)
  value = trim_unicode_whitespace(value)
  for each Unicode scalar in value:
    if scalar is ASCII "A".."Z", replace with corresponding "a".."z"
    otherwise preserve scalar exactly
  return value
```

The algorithm MUST NOT collapse internal whitespace, strip punctuation, strip underscores, strip hyphens, apply locale-specific case folding, apply Unicode compatibility normalization, or translate visible labels. Empty headers produce the empty match key.

**NF-REQ-080b**
Mapping suggestions under `cisco_sna_netflow_csv_v1` MUST be computed by Table 9-F1.

**Table 9-F1. Alias suggestion algorithm**

| Step | Required behavior |
| ---: | --- |
| 1 | Build an alias registry from Table 9-F by applying `source_alias_match_key_v1` to every alias. Multiple aliases that normalize to the same key and same target field collapse to one alias entry. The same normalized alias key targeting two different fields is an NLSpec defect and MUST fail profile conformance before use. |
| 2 | For every source column descriptor, set `normalized_header_for_suggestion` to `source_alias_match_key_v1(raw_header_text)`. |
| 3 | For each target field in Table 9-D order, collect source columns whose normalized header key maps to that field in the alias registry. |
| 4 | If exactly one source column matches a field, the UI MAY preselect a `source_column_mapping_v1` suggestion. Omission behavior: no suggestion is preselected, and the user must map the field manually. |
| 5 | If more than one source column matches the same field, the lowest `source_column_ordinal` is the deterministic default suggestion, but the mapping modal MUST emit a blocking duplicate-alias warning until the user explicitly approves one mapped source column and the unmapped handling for every other matched column. |
| 6 | If a source column is manually mapped to more than one non-combinable target field, mapping approval MUST fail with `network_flow_mapping_conflict` and `reason_code='source_column_reused'`. |
| 7 | Suggestions MUST be displayed and serialized for preview in Table 9-D field order, then `source_column_ordinal ASC`. Suggestions MUST NOT create a table, commit a mapping, or start import apply without explicit approval. |

### 9.7 Timestamp profile

**NF-REQ-081**
Every approved Network Flow mapping MUST contain a `timestamp_profile_v1` object satisfying Table 9-G.

**Table 9-G. `timestamp_profile_v1` object**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | none | Exactly `cartulary.network_flow.timestamp_profile.v1`. |
| `mode` | token | Yes | No | none | `rfc3339`, `epoch_seconds`, `epoch_milliseconds`, or `netflow_sys_uptime_milliseconds`. |
| `timezone` | string | No | Yes | `UTC` for epoch modes; null for `rfc3339` | `UTC` or an explicit IANA timezone name. Null is valid only for RFC3339 values that carry their own offset. |
| `precision` | token | No | No | mode-derived by Table 9-H | `seconds`, `milliseconds`, or `microseconds`, constrained by Table 9-H. |
| `netflow_export_time_column_ordinal` | integer | Yes | Yes | null | Required when `mode='netflow_sys_uptime_milliseconds'`; otherwise null. Value range is `1..network_flow.max_columns_per_csv`. |
| `netflow_export_time_mode` | token | Yes | Yes | null | Required when `mode='netflow_sys_uptime_milliseconds'`; otherwise null. One of `rfc3339`, `epoch_seconds`, or `epoch_milliseconds`, interpreted as UTC. |
| `netflow_exporter_uptime_at_export_column_ordinal` | integer | Yes | Yes | null | Required when `mode='netflow_sys_uptime_milliseconds'`; otherwise null. Value range is `1..network_flow.max_columns_per_csv`. |
| `ambiguous_local_time_policy` | token | Yes | No | `reject` | Only `reject` in v1. |
| `local_time_gap_policy` | token | Yes | No | `reject` | Only `reject` in v1. |

**Table 9-H. Timestamp mode precision**

| `mode` | Default `precision` | Permitted `precision` values |
| --- | --- | --- |
| `rfc3339` | `microseconds` | `seconds`, `milliseconds`, `microseconds` |
| `epoch_seconds` | `seconds` | `seconds` |
| `epoch_milliseconds` | `milliseconds` | `milliseconds` |
| `netflow_sys_uptime_milliseconds` | `milliseconds` | `milliseconds` |

**NF-REQ-082**
Under `mode='rfc3339'`, when `timezone` is null, every source value MUST carry `Z` or an explicit numeric offset. When `timezone` is a non-null IANA name, source values carrying an offset MUST use their own offset; source values without an offset MUST be interpreted in that zone with `reject` behavior for DST folds and gaps. The v1 Cisco SNA profile default timestamp behavior MUST NOT infer a deployment-local timezone.

A source value with finer resolution than `precision` MUST fail row validation with `network_flow_invalid_timestamp`. Epoch-mode values MUST be unsigned decimal integers in the mode's unit; fractional epoch values are invalid. A `precision` outside the mode's permitted set MUST fail mapping approval with `network_flow_mapping_conflict`.

**NF-REQ-083**
Epoch seconds and epoch milliseconds are interpreted as UTC instants. They MUST NOT be interpreted using the caller's browser timezone, deployment timezone, database timezone, or locale settings.

**NF-REQ-084**
When `mode='netflow_sys_uptime_milliseconds'`, the implementation MUST derive an absolute timestamp using this formula:

```text
derived_event_time = export_time_utc - (exporter_uptime_at_export_ms - event_uptime_ms)
```

`export_time_utc` is parsed from `netflow_export_time_column_ordinal` using `netflow_export_time_mode`. `exporter_uptime_at_export_ms` is parsed as an unsigned decimal integer from `netflow_exporter_uptime_at_export_column_ordinal`. `event_uptime_ms` is parsed as an unsigned decimal integer from the source column mapped to the timestamp target field being transformed. A row using this mode MUST fail with `network_flow_invalid_timestamp` when any participating value is missing, negative, non-integer, finer than the declared precision, outside its scalar bounds, or would produce a timestamp outside `timestamp_utc_v1`. A row MUST also fail with `network_flow_invalid_timestamp` when `event_uptime_ms > exporter_uptime_at_export_ms`, because that case is reboot- or wrap-ambiguous in v1.

**NF-REQ-085**
Flow end time MUST be greater than or equal to flow start time after UTC normalization. A row whose end time is earlier than start time MUST be rejected with `network_flow_end_before_start`.

## 10. Import-session target integration and mapping schemas

### 10.1 Import target kind and owner boundary

**NF-REQ-086**
The extension MUST register an import target with these exact values.

| Member | Required value |
| --- | --- |
| `target_kind` | `network_flow_table` |
| `target_table_schema_id` | `cartulary.network_flow_table.v1` |
| `default_source_profile_id` | `cisco_sna_netflow_csv_v1` when launched from `Import NetFlow CSV`; otherwise no default unless a source-profile selector supplies one. |
| `default_parser_profile_id` | `rfc4180_headered_csv_v1` |
| `default_unknown_column_policy` | `preserve_unmapped_raw` |

**NF-REQ-087**
The import module MUST continue to own upload, source byte validation, import session creation, import unit discovery, preview, mapping submission, selection, skip, background job admission, and apply orchestration. The Network Flow Activity extension MUST own only the target-specific mapping registry, row validation, table creation, row storage, graph adapter, and link behavior.

**NF-REQ-088**
The import dispatcher MUST call a Network Flow owner facade named `network_flow_import_facade_v1` for applied units whose approved mapping has `target_kind='network_flow_table'`. The import module MUST NOT write `network_flow_table` rows, `network_flow_row` rows, rejected-row diagnostics, or indicator bindings directly.

### 10.2 Source column descriptors

**NF-REQ-089**
Every source column descriptor MUST satisfy `source_column_v1` in Table 10-A.

**Table 10-A. `source_column_v1`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `source_column_ordinal` | integer | Yes | No | `1..network_flow.max_columns_per_csv`; source order. |
| `raw_header_text` | string | Yes | No | Exact decoded header field before suggestion normalization. |
| `normalized_header_for_suggestion` | string | Yes | No | `source_alias_match_key_v1(raw_header_text)` from §9.6; suggestion only, never identity. |
| `raw_header_sha256` | `sha256_hex_v1` | Yes | No | SHA-256 of UTF-8 raw header text. |
| `sample_values[]` | array | Yes | No | At most 50 `safe_sample_v1` objects in source order using Table 12-F. |
| `detected_empty_count` | non-negative integer | Yes | No | Count within preview scope. |

### 10.3 Approved mapping metadata

**NF-REQ-090**
An approved mapping for this extension MUST include the top-level members in Table 10-B.

**Table 10-B. Approved mapping metadata**

| Member | Type | Required | Nullable | Omission behavior | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `target_kind` | string | Yes | No | none | Exactly `network_flow_table`. |
| `target_table_schema_id` | string | Yes | No | none | Exactly `cartulary.network_flow_table.v1`. |
| `source_profile_id` | string | Yes | No | none | One conformant v1 profile ID from Table 9-A. Reserved profiles invalid. |
| `parser_profile_id` | string | Yes | No | none | Exactly `rfc4180_headered_csv_v1` in this revision. |
| `unknown_column_policy` | string | Yes | No | default `preserve_unmapped_raw` | Exactly one token from Table 10-C. |
| `display_name_override` | string | No | No | Omitted means derive by §8.4. | Explicit null invalid. When present, normalize through `normalize_table_display_name_input_v1`; empty or longer than 64 Unicode scalars invalid. |
| `timestamp_profile` | object | Yes | No | none | `timestamp_profile_v1` from §9.7. |
| `source_columns[]` | array | Yes | No | none | Exhaustive descriptors for every source column in source order. |
| `field_mappings[]` | array | Yes | No | none | Closed mapping-row variants from §10.4. |

**Table 10-C. Unknown-column policy**

| Token | Required behavior |
| --- | --- |
| `preserve_unmapped_raw` | Default. Retain unmapped source values as inert row provenance. They are not filterable, sortable, graphable, or linkable until mapped to a registered field in a later import. |
| `reject_unmapped_columns` | Reject mapping approval when any source column is unmapped and not explicitly ignored. |
| `ignore_unmapped_columns` | Do not retain unmapped values. This token requires explicit user approval in the mapping modal. |

**NF-REQ-091**
The default unknown-column policy MUST be `preserve_unmapped_raw`. The implementation MUST NOT default to rejecting unknown vendor columns.

### 10.4 Field mapping rows

**NF-REQ-092**
Every item in `field_mappings[]` MUST be exactly one variant from Table 10-D. The variant is selected only by `mapping_kind`. Members not listed for the selected variant are unknown members and MUST fail mapping approval with `network_flow_mapping_conflict`.

**Table 10-D. `field_mapping_v1` closed variants**

| `mapping_kind` | Variant schema | Creates normalized field value | Permitted target |
| --- | --- | ---: | --- |
| `source_column` | Table 10-D1 | Yes | Any non-`system_derived` field from Table 9-D. |
| `ignored_source_column` | Table 10-D2 | No | No `field_key`; accounting only. |
| `system_derivation` | Table 10-D3 | Yes | Only `network_flow.observation_source_ref`. |

**Table 10-D1. `source_column_mapping_v1`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `mapping_kind` | token | Yes | No | Exactly `source_column`. |
| `field_key` | string | Yes | No | One non-`system_derived` field key from Table 9-D. |
| `source_column_ordinal` | integer | Yes | No | `1..network_flow.max_columns_per_csv`; must identify one descriptor in `source_columns[]`. |
| `transform_id` | token | Yes | No | One token from Table 10-E valid for the target field contract. |
| `empty_value_policy` | token | Yes | No | One token from Table 10-F after `profile_default` expansion. |
| `combinability` | token | Yes | No | Exactly `single_source_only` in v1. |

**Table 10-D2. `ignored_source_column_mapping_v1`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `mapping_kind` | token | Yes | No | Exactly `ignored_source_column`. |
| `source_column_ordinal` | integer | Yes | No | `1..network_flow.max_columns_per_csv`; must identify one descriptor in `source_columns[]`. |
| `ignore_reason` | token | Yes | No | `user_ignored` or `policy_accounting`. |

**Table 10-D3. `system_derivation_mapping_v1`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `mapping_kind` | token | Yes | No | Exactly `system_derivation`. |
| `field_key` | string | Yes | No | Exactly `network_flow.observation_source_ref`. |
| `derivation_id` | token | Yes | No | Exactly `network_flow.observation_source_ref.v1`. |
| `combinability` | token | Yes | No | Exactly `single_source_only` in v1. |

**NF-REQ-093**
The tokens `derived` and `constant` are unavailable mapping kinds in v1. A persisted mapping or mapping approval request that uses either token MUST fail with `network_flow_mapping_conflict` and `reason_code='mapping_kind_unavailable'`. The implementation MUST NOT synthesize constants for required fields in v1.

### 10.5 Transform registry

**NF-REQ-094**
Transform behavior MUST use Table 10-E.

**Table 10-E. Transform registry**

| `transform_id` | Input | Output | Failure |
| --- | --- | --- | --- |
| `identity_text_v1` | decoded string | bounded text | Field diagnostic. |
| `trim_ascii_space_v1` | decoded string | string | Never fails. |
| `ip_literal_v1` | decoded string | canonical IP | `network_flow_invalid_ip`. |
| `port_number_v1` | decoded string | integer/null | `network_flow_invalid_port`. |
| `protocol_number_or_token_v1` | decoded string | integer | `network_flow_invalid_protocol`. |
| `uint64_decimal_string_v1` | decoded string | decimal string | `network_flow_invalid_counter`. |
| `timestamp_profile_v1` | decoded string | `timestamp_utc_v1` | `network_flow_invalid_timestamp`. |

**NF-REQ-095**
Transform failure on a mapped required field MUST reject the row. Transform failure on a mapped optional field MUST reject the row unless the mapping row declares an empty-value policy that converts only an empty string to null before the transform runs.

### 10.6 Empty-value policies

**NF-REQ-096**
Empty-value behavior MUST use Table 10-F.

**Table 10-F. Empty-value policy registry**

| `empty_value_policy` | Behavior |
| --- | --- |
| `empty_string_is_null` | Empty decoded source string becomes null before target validation. |
| `empty_string_is_invalid` | Empty decoded source string creates a diagnostic for mapped fields. |
| `empty_string_preserved` | Empty decoded source string remains `""` for text targets. |
| `profile_default` | Allowed only during UI suggestion; must expand before mapping approval and must not persist. |

**NF-REQ-097**
For Cisco SNA required fields, default empty-value policy MUST materialize as `empty_string_is_invalid` before mapping approval. A persisted mapping containing `profile_default` is invalid.

## 11. Mapping modal behavior

**NF-REQ-098**
The mapping modal MUST be explicit. The implementation MAY provide suggestions. Omission behavior: suggestions do not approve a mapping, do not create a table, and do not start import apply without explicit user approval.

**NF-REQ-099**
The mapping modal MUST display, at minimum:

- source filename display value;
- source profile selector showing only claimable profiles;
- parser profile identifier;
- source columns by `source_column_ordinal`, raw header, inferred sample type, and sample values;
- target field registry with requirement tokens under the selected source profile;
- selected timestamp profile;
- unknown-column policy;
- count of currently mapped required fields;
- row-validation preview summary with counts labeled as covering only the parser preview slice;
- blocking diagnostics that prevent apply.

**NF-REQ-100**
Mapping apply MUST be blocked with `network_flow_mapping_required` until every required field for the selected source profile is mapped from a source column or satisfied by an allowed `system_derivation` mapping.

**NF-REQ-101**
Mapping apply MUST fail with `network_flow_mapping_conflict` when one target field receives multiple non-combinable source columns, one source column is mapped to incompatible target field types, one transform cannot produce the target scalar contract, or `profile_default` remains in the stored mapping.

**NF-REQ-102**
Duplicate source headers MUST be displayed with source column ordinals. The mapping modal MUST NOT collapse duplicate header labels into one selectable source field.

**NF-REQ-103**
Switching `source_profile_id` after mapping approval MUST invalidate approval unless every mapped target field, transform, required-field rule, timestamp profile, and empty-value policy remains valid under the new profile. The UI MUST make the invalidation visible as `mapping_required`.

**Table 11-A. Mapping and validation UI state machine**

| State | Required trigger | User-visible consequence | Apply allowed |
| --- | --- | --- | ---: |
| `mapping_required` | Required field unmapped, profile changed, timestamp profile invalid, transform invalid, or persisted `profile_default`. | Blocking status and field-specific diagnostics. | No |
| `mapping_ready` | All required fields mapped from source columns or satisfied by allowed system derivations, with no blocking diagnostics. | Apply action enabled after current validation preview is not stale. | Yes |
| `validation_preview_pending` | Mapping changed and preview is recomputing. | Prior preview marked stale. | No |
| `validation_preview_ready` | Preview completed for current mapping fingerprint. | Accepted/rejected preview counts shown. | Yes if no blocking mapping diagnostics |
| `validation_preview_failed` | Preview could not complete due to parser, profile, admission, or resource failure. | Blocking diagnostic shown. | No |

`validation_preview_ready` counts are preview-scope only and cover the parser preview slice from Table 9-B. Apply-time full validation MAY reject rows the preview accepted. Omission behavior: preview acceptance does not guarantee import apply success; an apply whose full validation accepts zero rows fails with `network_flow_all_rows_rejected` regardless of preview state.

## 12. Normalized flow rows, row refs, validation, and diagnostics

### 12.1 Accepted row resource

**NF-REQ-104**
Each accepted `network_flow_row` MUST expose the fields in Table 12-A to extension-owned routes.

**Table 12-A. `network_flow_row` resource**

| Field | Type | Required | Rule |
| --- | --- | ---: | --- |
| `network_flow_row_id` | Identifier | Yes | Generated by §6.5. |
| `network_flow_table_id` | Identifier | Yes | Owning table. |
| `incident_id` | Core incident ID | Yes | Owning incident. |
| `source_row_number` | Positive integer | Yes | CSV record number. |
| `source_row_digest_sha256` | `sha256_hex_v1` | Yes | `source_row_digest_v1`. |
| `normalized_row_digest_sha256` | `sha256_hex_v1` | Yes | `normalized_row_digest_v1`. |
| `mapping_fingerprint` | `sha256_hex_v1` | Yes | Approved mapping fingerprint. |
| `network_flow.flow_start_utc` | `timestamp_utc_v1` | Yes | Normalized field. |
| `network_flow.flow_end_utc` | `timestamp_utc_v1` | Yes | Normalized field. |
| `network_flow.src_ip` | `ip_literal_v1` | Yes | Normalized field. |
| `network_flow.dst_ip` | `ip_literal_v1` | Yes | Normalized field. |
| `network_flow.src_port` | `port_number_v1` or null | Yes | Null permitted only when source profile permits null. |
| `network_flow.dst_port` | `port_number_v1` or null | Yes | Null permitted only when source profile permits null. |
| `network_flow.ip_protocol` | integer | Yes | `0..255`. |
| `network_flow.bytes_count` | `uint64_decimal_string_v1` | Yes | Stored as decimal string. |
| `network_flow.packets_count` | `uint64_decimal_string_v1` | Yes | Stored as decimal string. |
| Optional mapped fields | Field-specific | No | Present only when mapped and valid. |
| `unmapped_raw` | object | Yes | Contains retained unmapped source values only when unknown-column policy preserves them. Empty object when none. |
| `observation_source_ref` | object | Yes | Source provenance object from §12.3. |
| `created_at` | `timestamp_utc_v1` | Yes | Import commit time. |
| `created_by_user_id` | User ID | Yes | Importing actor. |

**NF-REQ-105**
`unmapped_raw` values MUST be inert provenance. They MUST NOT be filterable, sortable, graphable, exported, linked to indicators, or used for canonical row identity except through `source_row_digest_sha256` unless a later revision defines a promotion operation.

**NF-REQ-106**
`unmapped_raw` MUST be a canonical JSON object keyed by source column ordinal as a decimal string. Each member value MUST contain `raw_header_text`, `raw_header_sha256`, `decoded_value`, and `decoded_value_sha256`.

### 12.2 Row ref object

**NF-REQ-107**
`network_flow_row_ref_v1` MUST use Table 12-B.

**Table 12-B. `network_flow_row_ref_v1`**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `network_flow_table_id` | table ID | Yes | Owning table. |
| `network_flow_row_id` | row ID | Yes | Accepted row ID. |
| `source_row_number` | positive integer | Yes | Source CSV record number. |
| `mapping_fingerprint` | `sha256_hex_v1` | Yes | Mapping fingerprint for provenance. |

**NF-REQ-108**
`network_flow_row_ref_v1` MUST be used by graph example refs, selected-row indicator links, and contributing-row pivots.

### 12.3 Observation source ref

**NF-REQ-109**
The `observation_source_ref` object MUST include the members in Table 12-C.

**Table 12-C. `observation_source_ref`**

| Member | Type | Required |
| --- | --- | ---: |
| `import_session_id` | Core import session ID | Yes |
| `import_unit_id` | Core import unit ID | Yes |
| `source_content_sha256` | `sha256_hex_v1` | Yes |
| `source_profile_id` | source profile token | Yes |
| `parser_profile_id` | parser profile token | Yes |
| `mapping_fingerprint` | `sha256_hex_v1` | Yes |
| `source_row_number` | positive integer | Yes |
| `source_row_digest_sha256` | `sha256_hex_v1` | Yes |

### 12.4 Row validation and rejected-row diagnostics

**NF-REQ-110**
Row validation MUST classify each source data row as exactly one of the classifications in Table 12-D.

**Table 12-D. Row classification**

| Classification | Meaning |
| --- | --- |
| `accepted` | All required fields and mapped optional fields satisfy their contracts. The row is stored as a `network_flow_row`. |
| `rejected` | One or more blocking diagnostics exist. The row is not stored as an accepted row and MUST NOT contribute to graph output. |

**NF-REQ-111**
Row validation MUST run after mapping approval and before table activation. A row with one or more blocking diagnostics MUST NOT create a partial graph edge, partial endpoint, indicator observation, or accepted table row.

**NF-REQ-112**
Rejected-row diagnostics MUST expose the shape in Table 12-E.

**Table 12-E. Rejected-row diagnostic resource**

| Field | Type | Required | Rule |
| --- | --- | ---: | --- |
| `source_row_number` | positive integer | Yes | CSV record number. |
| `source_column_ordinal` | integer or null | Yes | Null only when error is row-level. |
| `raw_header` | string or null | Yes | Null only when error is row-level. |
| `field_key` | string or null | Yes | Null only when no target field is attributable. |
| `error_code` | error token | Yes | From §21. |
| `safe_sample` | string or null | Yes | Bounded safe sample after redaction under Table 12-F. |
| `raw_value_sha256` | `sha256_hex_v1` or null | Yes | Required whenever a source scalar exists; null only for row-level diagnostics with no attributable source scalar. |
| `message` | `bounded_text_1024_v1` | Yes | Safe, deterministic diagnostic text. |

**NF-REQ-113**
Diagnostics MUST be ordered by `source_row_number ASC`, then `source_column_ordinal ASC` with nulls last, then `field_key ASC` with nulls last, then `error_code ASC`.

**NF-REQ-114**
When rejected-row diagnostics exceed `network_flow.max_rejected_row_diagnostics`, the implementation MUST retain the first `N` diagnostics under the ordering in §12.4 and set `diagnostics_truncated=true` on the table. It MUST NOT pretend that omitted diagnostics do not exist.

**Table 12-F. Safe sample and digest behavior**

| Value class | `safe_sample` | `raw_value_sha256` |
| --- | --- | --- |
| Row-level diagnostic with no attributable source scalar | `null`. | `null`. |
| IP address, hostname, domain, indicator candidate, filename, raw CSV cell text, raw header text, interface text, application label, exporter ID, or any bounded text field | `null`. | SHA-256 of the exact decoded source scalar. |
| Formula-leading text whose first Unicode scalar is `=`, `+`, `-`, `@`, tab, CR, or LF | `null`. | SHA-256 of the exact decoded source scalar. |
| Oversized text or text rejected for a control character | `null`. | SHA-256 of the exact decoded source scalar. |
| Port, protocol number, counter, source-row number, column ordinal, row count, or other integer-like scalar whose exact decoded source value matches `^[0-9]{1,32}$` | Exact decoded source value. | SHA-256 of the exact decoded source scalar. |
| Any other value class | `null`. | SHA-256 of the exact decoded source scalar when a source scalar exists; otherwise `null`. |

`safe_sample_v1` objects used in source column descriptors MUST contain exactly `safe_sample` and `raw_value_sha256` members following Table 12-F. The implementation MUST NOT expose raw source scalar text through any other diagnostic, sample, metadata, audit, log, telemetry, or route member unless this NLSpec explicitly names that member.

## 13. Table query, row query, filter, sort, and cursor contracts

### 13.1 Query request members

**NF-REQ-115**
The single-table query and cross-table row query routes MUST use the query members in Table 13-A.

**Table 13-A. Query request members**

| Member | Type | Required | Nullable | Default | Bound |
| --- | --- | ---: | ---: | --- | --- |
| `filters[]` | array of filter objects | No | No | `[]` | Max `network_flow.max_filters_per_query`. |
| `sort[]` | array of sort objects | No | No | `[]`, then default sort tail | Max `network_flow.max_sorts_per_query`. |
| `limit` | integer | No | No | `min(200, effective network_flow.max_query_limit)` | `1..effective network_flow.max_query_limit`. |
| `cursor_token` | string | No | No | omitted | Opaque continuation token. |

A supplied `limit` MUST be an integer in `1..effective network_flow.max_query_limit` and MUST fail with `network_flow_invalid_limit` when outside that range. Clamping applies only to the omitted-member default and MUST NOT apply to explicit caller input.

**NF-REQ-116**
The default row sort MUST be:

```text
network_flow.flow_start_utc ASC,
network_flow.flow_end_utc ASC,
source_row_number ASC,
network_flow_row_id ASC
```

### 13.2 Table scope

**NF-REQ-117**
A cross-table row query or graph query MUST contain `table_scope` using Table 13-B.

**Table 13-B. `table_scope_v1`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `mode` | token | Yes | No | none | `active_table`, `selected_tables`, or `all_active_tables`. |
| `active_table_id` | table ID | Yes | Yes | null | Required when `mode='active_table'`; otherwise null. |
| `selected_table_ids[]` | array | Yes | Yes | null | Required when `mode='selected_tables'`; otherwise null. Length `1..network_flow.max_selected_tables_per_graph`. |

**NF-REQ-118**
`selected_table_ids[]` with duplicate table IDs MUST fail with `network_flow_invalid_table_scope`. The implementation MUST NOT silently deduplicate the list.

**NF-REQ-119**
`all_active_tables` MUST select all active tables in the incident ordered by `created_at ASC`, then `network_flow_table_id ASC`. It MUST NOT include creating, failed, or soft-deleted tables.

### 13.3 Filter grammar

**NF-REQ-120**
All filters MUST be conjunctive. The implementation MUST apply every filter in `filters[]` as logical AND. Disjunction is represented only by `op='in'`. This revision defines no general OR expression.

**NF-REQ-121**
A filter object MUST have exactly the members in Table 13-C.

**Table 13-C. Filter object**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `field_key` | string | Yes | No | Stable field key or pseudo-field from Table 13-D. Visible labels invalid. |
| `op` | string | Yes | No | Operator token allowed by Table 13-D. |
| `value` | JSON value | Depends on `op` | No unless table says so | Required unless `op` is `is_null` or `not_null`. Explicit null invalid except inside range bound objects. |

**Table 13-D. Filterable fields and operators**

| Field key | Operators | Value contract |
| --- | --- | --- |
| `network_flow.src_ip` | `eq`, `in`, `cidr_contains` | IP literal, array of IP literals, or CIDR literal. |
| `network_flow.dst_ip` | `eq`, `in`, `cidr_contains` | IP literal, array of IP literals, or CIDR literal. |
| `network_flow.endpoint_ip` | `eq`, `in`, `cidr_contains` | Pseudo-field matching either source or destination IP. |
| `network_flow.src_port` | `eq`, `in`, `range`, `is_null`, `not_null` | Port number, array of port numbers, or `{ "gte": n, "lte": n }`. |
| `network_flow.dst_port` | `eq`, `in`, `range`, `is_null`, `not_null` | Port number, array of port numbers, or range object. |
| `network_flow.ip_protocol` | `eq`, `in`, `range` | Integer protocol number or array/range of protocol numbers. |
| `network_flow.flow_start_utc` | `range` | `{ "gte": timestamp\|null, "lt": timestamp\|null }`; at least one bound non-null. |
| `network_flow.flow_end_utc` | `range` | `{ "gte": timestamp\|null, "lt": timestamp\|null }`; at least one bound non-null. |
| `network_flow.bytes_count` | `eq`, `range` | Decimal string or `{ "gte": decimal_string\|null, "lte": decimal_string\|null }`. |
| `network_flow.packets_count` | `eq`, `range` | Decimal string or range object. |
| `network_flow.exporter_id` | `eq`, `in`, `prefix`, `contains`, `is_null`, `not_null` | Bounded text value. |
| `network_flow.input_interface` | `eq`, `in`, `prefix`, `contains`, `is_null`, `not_null` | Bounded text value. |
| `network_flow.output_interface` | `eq`, `in`, `prefix`, `contains`, `is_null`, `not_null` | Bounded text value. |
| `source_row_number` | `eq`, `in`, `range` | Positive integer or range. |

**Table 13-E. Operator validation and evaluation**

| Operator | Validation | Evaluation |
| --- | --- | --- |
| `eq` | One scalar satisfying field contract; explicit null invalid. | Canonical scalar equality. Null fields do not match. |
| `in` | Array length `1..256`; all entries satisfy field contract; duplicates invalid. | Match any canonical scalar. |
| `range` | Object with allowed bounds only; at least one non-null bound. Numeric ranges with `gte > lte` and timestamp ranges with `gte >= lt` are invalid with `reason_code='empty_range'`. | Timestamp ranges use `{gte, lt}`. Numeric ranges use `{gte, lte}`. |
| `cidr_contains` | CIDR literal; IP family must match target field family except for `network_flow.endpoint_ip`. | True when canonical IP is inside canonical CIDR. IPv4-mapped IPv6 remains IPv6 and does not match IPv4 CIDR. |
| `prefix` | Non-empty string; text fields only. | Exact code-point prefix after stored normalization. |
| `contains` | Non-empty string; text fields only. | Exact code-point substring after stored normalization. |
| `is_null` | `value` member must be omitted. | True when stored field is null or optional field absent. |
| `not_null` | `value` member must be omitted. | True when stored field is present and non-null. |

For the pseudo-field `network_flow.endpoint_ip`, `eq`, `in`, and `cidr_contains` validation accepts IPv4 and IPv6 values. Evaluation compares each supplied value only against same-family stored source or destination IP values; IPv4-mapped IPv6 remains IPv6.

### 13.4 Sort grammar

**NF-REQ-122**
A sort object MUST contain exactly `field_key` and `direction`. `direction` MUST be `asc` or `desc`. Duplicate caller-supplied sort field keys MUST fail with `network_flow_invalid_sort`.

**NF-REQ-123**
Only the fields in Table 13-D except `network_flow.endpoint_ip`, plus `network_flow_row_id`, `network_flow_table_id`, and `source_row_number`, are sortable. The implementation MUST reject visible labels, unmapped raw fields, graph IDs, pseudo-field `network_flow.endpoint_ip`, and storage field names as sort keys.

**Table 13-F. Sort comparison rules**

| Field class | Ascending comparison | Null ordering |
| --- | --- | --- |
| Timestamp | Chronological instant, then canonical timestamp bytes. | Nulls last. |
| IP | IPv4 before IPv6, then unsigned address bytes. | Nulls last. |
| Port/protocol | Numeric integer. | Nulls last. |
| Decimal counters | Arbitrary-precision unsigned integer. | Nulls last. |
| Text | Exact Unicode code point order after stored normalization. | Nulls last. |
| IDs | ASCII byte order. | Never null. |

**NF-REQ-124**
Descending sort reverses value ordering but keeps nulls last. The implementation MUST compute `effective_sort_v1` for every row query.

```text
effective_sort_v1(caller_sort):
  result = caller_sort in caller-supplied order
  for each default sort item in NF-REQ-116 order:
    if result does not already contain that field_key:
      append default sort item with direction "asc"
  return result
```

Omission behavior: when caller `sort[]` is omitted or empty, `effective_sort_v1` is exactly the default row sort from NF-REQ-116. `network_flow.max_sorts_per_query` applies to caller-supplied sort entries before the default tail is appended. The response MUST echo both normalized caller `sort[]` and `effective_sort`.

### 13.5 Cursor lifecycle

**NF-REQ-125**
Network Flow query cursors MUST be opaque live-authorized keyset cursors.

**Table 13-G. Cursor lifecycle**

| Condition | Required behavior |
| --- | --- |
| TTL | Exactly 15 minutes after `issued_at`. |
| Authorization loss | Revalidate at continuation; fail closed with `network_flow_cursor_invalid`. |
| Actor/session mismatch | Fail with `network_flow_cursor_invalid`. |
| Route, incident, table scope, filters, sort, or limit mismatch | Fail with `network_flow_cursor_invalid`. |
| Table soft delete in cursor scope | Fail with `network_flow_cursor_invalid`. |
| Table rename in cursor scope | Continuation remains valid. |
| Newly imported table | Not included in an existing cross-table cursor. |
| Newly imported rows in same table | Impossible because rows are immutable after table activation. |
| Cursor malformed | Fail with `network_flow_cursor_invalid`; do not reveal cursor internals. |

**NF-REQ-126**
A cursor token expires exactly 15 minutes after `issued_at`.

**NF-REQ-126a**
A cursor token MUST bind actor, route, incident, table scope, filters, sort, limit, issued-at time, expiry time, table IDs, and continuation position. It MUST NOT bind table versions.

**NF-REQ-126b**
The cursor continuation position MUST be the last emitted item's full `effective_sort` tuple plus `network_flow_table_id` and `network_flow_row_id`. Continuation MUST return only rows that compare after that tuple under the same `effective_sort` comparator. The implementation MUST NOT use page offsets, visual row positions, browser grid indices, or table display names as continuation identity.

**NF-REQ-126c**
A query response MUST set `meta.paging.next_cursor_token` to an opaque string only when at least one additional row or diagnostic exists after the last returned item under the cursor's bound comparator and current authorization. Otherwise it MUST set `next_cursor_token` to JSON `null`. A zero-result response MUST set `next_cursor_token` to JSON `null`.

## 14. Cross-table graph-composition contract

### 14.1 Graph query request

**NF-REQ-127**
The graph query route MUST accept a request body with the members in Table 14-A.

**Table 14-A. `network_flow_graph_query_request_v1`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `table_scope` | `table_scope_v1` | Yes | No | none | Scope from §13.2. |
| `filters[]` | array | No | No | `[]` | Filter grammar from §13.3. |
| `time_range` | object | No | No | omitted means all time | `time_range_v1` from Table 14-B. |
| `aggregation` | object | No | No | default object from Table 14-C | `aggregation_v1`. |
| `limit_overrides` | object | No | No | omitted means effective configured limits | `limit_overrides_v1` from Table 14-D. |

**Table 14-B. `time_range_v1`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `start_utc` | `timestamp_utc_v1` | Yes | Yes | Null means unbounded lower bound. |
| `end_utc` | `timestamp_utc_v1` | Yes | Yes | Null means unbounded upper bound. |

**NF-REQ-128**
If `time_range` is present, at least one of `start_utc` or `end_utc` MUST be non-null. If both are non-null, `end_utc` MUST be greater than `start_utc`.

**Table 14-C. `aggregation_v1`**

| Member | Type | Required | Default | Rule |
| --- | --- | ---: | --- | --- |
| `mode` | token | No | `default_flow_edge_v1` | Only `default_flow_edge_v1` in this revision. |
| `include_example_row_refs` | boolean | No | `true` | Example refs are truncated by §14.5. |

**Table 14-D. `limit_overrides_v1`**

| Member | Type | Required | Default | Rule |
| --- | --- | ---: | --- | --- |
| `max_vertices` | integer | No | configured effective limit | Lower only. |
| `max_edges` | integer | No | configured effective limit | Lower only. |
| `max_example_row_refs_per_edge` | integer | No | configured effective limit | Lower only. |

**NF-REQ-129**
Unknown `limit_overrides` keys MUST fail with `network_flow_invalid_limit_override`. A supplied `limit_overrides` value MUST be an integer in `L..E`, where `L` is that key's Table 20-A lowerable minimum and `E` is its current effective limit. Values outside that range, including negative values, MUST fail with `network_flow_invalid_limit_override`.

### 14.2 Time filter rules

**NF-REQ-130**
A graph query time range MUST use interval overlap semantics.

```text
flow overlaps range when:
  (time_range.start_utc is null OR flow_end_utc >= time_range.start_utc)
  AND
  (time_range.end_utc is null OR flow_start_utc < time_range.end_utc)
```

**NF-REQ-131**
A zero-duration flow where `flow_start_utc == flow_end_utc` MUST be included when that instant is inside the half-open time range `[start_utc, end_utc)`. If `end_utc` equals the flow instant, the zero-duration flow is excluded.

### 14.3 Flow graph mapping

**NF-REQ-132**
The default graph composition MUST derive endpoint vertices from `network_flow.src_ip` and `network_flow.dst_ip` after filters and time-range selection.

**NF-REQ-133**
`network_flow_endpoint_id_v1` MUST compute endpoint IDs as follows.

```text
network_flow_endpoint_id_v1(incident_id, endpoint_kind, canonical_value):
  UTF8("cartulary.network_flow_endpoint_id.v1") NUL
  UTF8(incident_id) NUL
  UTF8(endpoint_kind) NUL
  UTF8(canonical_value) NUL
  return "nfe_" + lowercase_hex(SHA256(bytes))
```

**NF-REQ-134**
The only `endpoint_kind` in v1 is `ip`. The same canonical endpoint IP across multiple selected tables MUST emit one default graph vertex with combined contributing table provenance.

**NF-REQ-135**
Default graph edges MUST aggregate accepted rows by this key:

```text
(src_endpoint_id, dst_endpoint_id, ip_protocol, dst_port_presence, dst_port_value)
```

`dst_port_presence` is exactly `p` when `dst_port` is present and `n` when `dst_port` is null. `dst_port_value` is the canonical decimal destination port when present and empty when null. Rows with the same key across selected tables contribute to the same default edge. Direction is source IP to destination IP.

**NF-REQ-135a**
`network_flow_flow_edge_id_v1` MUST compute default graph edge IDs as follows.

```text
network_flow_flow_edge_id_v1(incident_id, src_endpoint_id, dst_endpoint_id,
                             ip_protocol, dst_port_or_null):
  UTF8("cartulary.network_flow_flow_edge_id.v1") NUL
  UTF8(incident_id) NUL
  UTF8(src_endpoint_id) NUL
  UTF8(dst_endpoint_id) NUL
  UTF8(decimal(ip_protocol)) NUL
  UTF8(dst_port_presence) NUL
  UTF8(decimal(dst_port) or empty) NUL
  return "nff_" + lowercase_hex(SHA256(bytes))
```

The presence token in this algorithm MUST use the same `p` and `n` values as the aggregation key.

**Table 14-E. Default vertex properties**

| Property | Type | Rule |
| --- | --- | --- |
| `endpoint_kind` | token | `ip`. |
| `endpoint_value` | `ip_literal_v1` | Canonical IP text. |
| `contributing_table_ids[]` | array | Unique active table IDs ordered by workspace table order. |
| `flow_row_count` | integer | Count of contributing accepted rows. |
| `indicator_candidate_value` | string | Same as endpoint value; user-visible candidate only. |

**Table 14-F. Default edge properties**

| Property | Type | Rule |
| --- | --- | --- |
| `edge_id` | edge ID | Derived by `network_flow_flow_edge_id_v1`. |
| `src_endpoint_id` | endpoint ID | Derived by §14.3. |
| `dst_endpoint_id` | endpoint ID | Derived by §14.3. |
| `ip_protocol` | integer | Aggregation key member. |
| `dst_port` | integer or null | Aggregation key member. |
| `flow_row_count` | integer | Count of contributing accepted rows. |
| `bytes_sum` | `uint64_decimal_string_v1` | Sum of contributing `bytes_count`; fail before output if sum exceeds digit limit. |
| `packets_sum` | `uint64_decimal_string_v1` | Sum of contributing `packets_count`; fail before output if sum exceeds digit limit. |
| `first_flow_start_utc` | `timestamp_utc_v1` | Earliest contributing start. |
| `last_flow_end_utc` | `timestamp_utc_v1` | Latest contributing end. |
| `contributing_table_ids[]` | array | Unique active table IDs ordered by workspace table order. |
| `example_row_refs[]` | array | First refs under §14.5 ordering. |
| `example_refs_truncated` | boolean | True when omitted contributors exist. |
| `example_refs_total_count` | integer | Total contributing row ref count before truncation. |

### 14.4 Graph Projection adapter

**NF-REQ-136**
The graph query route MUST construct `network_flow_graph_projection_adapter_v1` through an ephemeral Graph Projection-compatible derivation boundary. The implementation MUST NOT submit v1 Network Flow graph queries through public retained Graph Projection lifecycle operations that allocate addressable graph views, retained runs, or retained graph output.

The exact Graph Projection input submitted by the adapter MUST satisfy Table 14-G after Graph Projection-owned default materialization. Network Flow owns the values in Table 14-G; Graph Projection owns validation, projected graph IDs, sorting, and graph-object derivation after the input is admitted.

**Table 14-G. `network_flow_graph_projection_input_v1`**

| Top-level member | Required value or derivation |
| --- | --- |
| `projection_schema_id` | Exactly `graph_projection.v1`. |
| `graph_view_id` | Graph Projection §7.4 ID derived from `projection_config.graph_view_key`; Network Flow MUST NOT invent another ID algorithm. |
| `source_snapshot_id` | `network_flow_source_snapshot_digest_v1` from §6.9. |
| `projection_config` | Exact object from Table 14-G1. |
| `source_entities[]` | One Table 14-G2 object per endpoint vertex after filters and time selection, ordered by `source_entity_id ASC` before submission. |
| `source_relationships[]` | One Table 14-G3 object per aggregated default flow edge, ordered by `source_relationship_id ASC` before submission. |
| `source_metadata` | Object containing exactly `incident_id`, `graph_query_digest`, `source_snapshot_id`, and `selected_table_ids[]`. |
| `filters` | `{ "entity_filters": [], "relationship_filters": [], "logic": "and" }`. Network Flow filtering is already applied before adapter submission. |
| `relationship_definitions[]` | `[]`. Relationship mappings live only in `projection_config.relationship_mappings[]`. |
| `property_definitions[]` | Exact array from Table 14-G4. |
| `requested_at` | Server request timestamp for the graph query route. |
| `requested_by` | Authenticated actor user ID. |

**Table 14-G1. Fixed `projection_config`**

| Member | Required value |
| --- | --- |
| `graph_view_key` | `network_flow_activity:{incident_id}:{source_snapshot_id}`. |
| `projection_version` | `network_flow_activity.v1`. |
| `declared_source_entity_kinds[]` | `["network_flow.ip_endpoint.v1"]`. |
| `declared_source_relationship_kinds[]` | `["network_flow.flow_edge.v1"]`. |
| `entity_mappings[]` | One rule with `mapping_rule_id='nf.map.ip_endpoint.v1'`, `source_entity_kind='network_flow.ip_endpoint.v1'`, `projected_vertex_kind='network_flow.ip_endpoint.v1'`, `inclusion_predicate='always'`, `label_policy='mapping_only'`, `mapping_labels=[]`, `required_property_keys=["endpoint_kind","endpoint_value","contributing_table_ids","flow_row_count","indicator_candidate_value"]`, and `optional_property_keys=[]`. |
| `relationship_mappings[]` | One rule with `mapping_rule_id='nf.map.flow_edge.v1'`, `source_relationship_kind='network_flow.flow_edge.v1'`, `projected_edge_kind='network_flow.flow_edge.v1'`, `inclusion_predicate='always'`, `direction_policy='preserve'`, `emit_reverse_edge=false`, `reverse_edge_kind='network_flow.flow_edge.v1'`, `label_policy='mapping_only'`, `mapping_labels=[]`, `required_property_keys=["edge_id","src_endpoint_id","dst_endpoint_id","ip_protocol","dst_port","flow_row_count","bytes_sum","packets_sum","first_flow_start_utc","last_flow_end_utc","contributing_table_ids","example_refs_truncated","example_refs_total_count"]`, and `optional_property_keys=[]`. |
| `metadata_mappings[]` | Safe metadata mappings only for `contributing_table_ids`, `mapping_fingerprints`, `flow_row_count`, and `example_refs_total_count`. |
| `aggregation_rules[]` | `[]`; Network Flow performs aggregation before adapter submission. |
| `default_vertex_labels[]` | `[]`. |
| `default_edge_labels[]` | `[]`. |
| `allow_empty_kind_registry` | `false`. |
| `retention_policy` | `{ "retain_replaced_results": false, "retention_count": 0, "retention_duration_seconds": 0, "retain_failed_results": false, "failed_retention_count": 0, "failed_retention_duration_seconds": 0 }`. |
| `custom_config` | `{}`. |

**Table 14-G2. Source entity object**

| Member | Required value |
| --- | --- |
| `source_entity_id` | `network_flow_endpoint_id`. |
| `source_entity_kind` | `network_flow.ip_endpoint.v1`. |
| `properties` | Exact members from Table 14-E. |
| `metadata` | Object containing exactly `contributing_table_ids[]`, `mapping_fingerprints[]`, and `flow_row_count`. |
| `labels[]` | `[]`. |

**Table 14-G3. Source relationship object**

| Member | Required value |
| --- | --- |
| `source_relationship_id` | `network_flow_flow_edge_id_v1` output. |
| `source_relationship_kind` | `network_flow.flow_edge.v1`. |
| `src_source_entity_id` | Source endpoint ID from the edge aggregation key. |
| `dst_source_entity_id` | Destination endpoint ID from the edge aggregation key. |
| `direction` | `forward`. |
| `properties` | Exact Graph Projection-compatible members from Table 14-F: `edge_id`, `src_endpoint_id`, `dst_endpoint_id`, `ip_protocol`, `dst_port`, `flow_row_count`, `bytes_sum`, `packets_sum`, `first_flow_start_utc`, `last_flow_end_utc`, `contributing_table_ids[]`, `example_refs_truncated`, and `example_refs_total_count`. `example_row_refs[]` MUST NOT be submitted to Graph Projection. |
| `metadata` | Object containing exactly `contributing_table_ids[]`, `mapping_fingerprints[]`, `example_refs_total_count`, and `example_refs_truncated`. |
| `labels[]` | `[]`. |

**Table 14-G4. Property definitions**

| Target scope | Target kind | Projected keys |
| --- | --- | --- |
| `vertex` | `network_flow.ip_endpoint.v1` | `endpoint_kind`, `endpoint_value`, `contributing_table_ids`, `flow_row_count`, `indicator_candidate_value`. |
| `edge` | `network_flow.flow_edge.v1` | `edge_id`, `src_endpoint_id`, `dst_endpoint_id`, `ip_protocol`, `dst_port`, `flow_row_count`, `bytes_sum`, `packets_sum`, `first_flow_start_utc`, `last_flow_end_utc`, `contributing_table_ids`, `example_refs_truncated`, `example_refs_total_count`. |

The `property_definitions[]` array MUST be ordered by Table 14-G4 row order, then projected-key order within the row.

Each property definition item MUST use Graph Projection's `property_definition` schema with:

- `property_definition_id='nf.pd.{target_scope}.{projected_key}.v1'`;
- `source_field_path='properties.{projected_key}'`;
- `required=true`;
- `missing_behavior='error'`;
- `source_null_behavior='preserve'` only for `dst_port` and `source_null_behavior='error'` for every other projected key;
- `null_output_policy='emit_null'` only for `dst_port` and `null_output_policy='omit'` for every other projected key;
- `merge_behavior='single_value'`;
- `projected_type` from Table 14-G5.

**Table 14-G5. Graph Projection property type mapping**

| Projected key | `projected_type` |
| --- | --- |
| `endpoint_kind`, `endpoint_value`, `indicator_candidate_value`, `edge_id`, `src_endpoint_id`, `dst_endpoint_id`, `bytes_sum`, `packets_sum` | `string` |
| `ip_protocol`, `dst_port`, `flow_row_count`, `example_refs_total_count` | `integer` |
| `first_flow_start_utc`, `last_flow_end_utc` | `timestamp` |
| `contributing_table_ids` | `identifier_array` |
| `example_refs_truncated` | `boolean` |

Raw CSV cells, raw headers, raw filename strings, display names, graph layout coordinates, browser-local labels, and `example_row_refs[]` MUST NOT appear in `source_metadata`, Graph Projection `properties`, Graph Projection `metadata`, labels, property definitions, or metadata mappings.

**NF-REQ-136a**
A later revision that exposes retained Network Flow graph views MUST define retained view lifecycle, invalidation, inspection, authorization, cleanup behavior, and a Graph Projection-compatible structured retention policy before adoption.

**NF-REQ-137**
Network Flow graph responses MUST NOT create or retain a Graph Projection graph view beyond the response unless a later revision defines retained graph view lifecycle, invalidation, inspection, authorization, and cleanup behavior. Omission behavior: v1 graph responses are ephemeral.

**NF-REQ-138**
Graph over-limit cases MUST fail with deterministic errors before emitting partial graph output. The implementation MUST NOT render a partial unlabeled graph, drop vertices silently, or return partial edges without an explicit error.

### 14.5 Example row refs and truncation

**NF-REQ-139**
When `include_example_row_refs=true`, example row refs MUST be ordered by:

```text
workspace_table_order ASC,
effective_sort_v1([]),
network_flow_row_id ASC
```

**NF-REQ-140**
Edge example refs MUST use Table 14-I.

**Table 14-I. Edge example-ref members**

| Member | Type | Required behavior |
| --- | --- | --- |
| `example_row_refs[]` | array of `network_flow_row_ref_v1` | First retained refs under §14.5 ordering. |
| `example_refs_truncated` | boolean | `true` only when the total contributing row ref count exceeds the retained `example_row_refs[]` length. |
| `example_refs_total_count` | non-negative integer | Total contributing row ref count before truncation. |

For example, when `max_example_row_refs_per_edge=3` and an edge has 250 contributing rows, `example_row_refs[]` contains the first 3 refs under §14.5 ordering, `example_refs_truncated=true`, and `example_refs_total_count=250`.

### 14.6 Graph response shape

**NF-REQ-141**
A successful graph query response `data` MUST contain Table 14-H members.

**Table 14-H. Graph query response data**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | Yes | `cartulary.network_flow_graph_query_result.v1`. |
| `graph_query_digest` | `sha256_hex_v1` | Yes | `network_flow_graph_query_digest_v1`. |
| `normalized_query` | object | Yes | Default-materialized query used for digest. |
| `graph_projection_result` | object | Yes | Output satisfying adopted Graph Projection contract. |
| `edge_annotations[]` | array | Yes | Network Flow edge annotations from Table 14-H1 ordered by `edge_id ASC`. |
| `source_table_refs[]` | array | Yes | Table ID, table version, mapping fingerprint, and counts for each selected table. |
| `result_limits` | object | Yes | Effective graph limits applied. |

**Table 14-H1. Network Flow edge annotation**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `edge_id` | edge ID | Yes | `network_flow_flow_edge_id_v1` output. |
| `example_row_refs[]` | array of `network_flow_row_ref_v1` | Yes | First retained refs under §14.5 ordering. Empty when `include_example_row_refs=false` or the effective example-ref limit is `0`. |
| `example_refs_truncated` | boolean | Yes | `true` only when omitted contributors exist. |
| `example_refs_total_count` | non-negative integer | Yes | Total contributing row ref count before truncation. |

## 15. Indicator linking and observation behavior

### 15.1 Link request schema

**NF-REQ-142**
The indicator-link route MUST accept `indicator_link_request_v1` with Table 15-A members.

**Table 15-A. `indicator_link_request_v1`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `client_txn_id` | `client_txn_id` | Yes | No | Mutating route idempotency key. |
| `selector` | object | Yes | No | One selector from Table 15-B. |
| `target` | object | Yes | No | One target from Table 15-C. |
| `observation_mode` | token | Yes | No | Exactly `binding_only` in v1. |
| `confirm_exact_value` | scalar string/integer | Yes | No | Candidate value echoed by UI to prevent ambiguous graph-edge links. |

**Table 15-A1. V1 linkable candidate registry**

| Selector context | Linkable `field_key` | Resolved candidate |
| --- | --- | --- |
| `row_field_value` | `network_flow.src_ip` | Referenced row's canonical source IP. |
| `row_field_value` | `network_flow.dst_ip` | Referenced row's canonical destination IP. |
| `row_refs` | `network_flow.src_ip` | Canonical source IP shared by every referenced row. |
| `row_refs` | `network_flow.dst_ip` | Canonical destination IP shared by every referenced row. |
| `graph_vertex` | no `field_key` member | Selected endpoint vertex canonical IP. |
| `graph_edge` | `network_flow.src_ip` | Source endpoint canonical IP for the selected edge. |
| `graph_edge` | `network_flow.dst_ip` | Destination endpoint canonical IP for the selected edge. |

No other field is linkable in v1. Selectors for ports, protocol, counters, interface text, application labels, filenames, table display names, raw CSV values, `unmapped_raw`, graph labels, or Graph Projection projected property names MUST fail with `network_flow_invalid_indicator_selector` and `reason_code='field_not_linkable'`.

**Table 15-B. Indicator selector kinds**

| `selector.kind` | Required members | Rule |
| --- | --- | --- |
| `row_field_value` | `network_flow_table_id`, `network_flow_row_id`, `field_key` | `field_key` must be linkable for this context in Table 15-A1. |
| `row_refs` | `row_refs[]`, `field_key` | `row_refs[]` length `1..network_flow.max_binding_source_row_refs`; every ref validates as `network_flow_row_ref_v1`; duplicate row IDs are invalid; every referenced row must resolve to the same canonical candidate value for `field_key`. |
| `graph_vertex` | `graph_query`, `graph_query_digest`, `vertex_id` | `vertex_id` is a `network_flow_endpoint_id`. Server recomputes graph query digest and validates that the selected endpoint vertex exists in the recomputed composition. A `field_key` member is invalid. |
| `graph_edge` | `graph_query`, `graph_query_digest`, `edge_id`, `field_key` | `edge_id` is a `network_flow_flow_edge_id`. `field_key` must be `network_flow.src_ip` or `network_flow.dst_ip`. Server recomputes graph query digest and validates that the selected edge exists in the recomputed composition plus exact candidate value. |

**Table 15-C. Link target modes**

| `target.mode` | Required members | Required behavior |
| --- | --- | --- |
| `existing_indicator` | `indicator_id` | Validate same incident and current visibility; Core owns canonical indicator. The target indicator MUST have Core `value_kind='atomic'`, Core normalized value equal to the resolved candidate IP, and the Core indicator type designated for IP literals. |
| `create_indicator` | `indicator_type`, `value_kind`, `display_value`, `normalized_value` | Delegate canonical creation and dedupe to Core indicator owner. `value_kind` MUST be `atomic`; `display_value` and `normalized_value` MUST equal the resolved candidate IP after Core IP normalization; `indicator_type` MUST be the Core registry token designated for IP literals. |

If Core has no indicator type designated for canonical IP literals, `create_indicator` from Network Flow is an adoption blocker and MUST fail closed with Core indicator-create validation behavior until Core 02 or the adopted Core indicator registry closes that dependency.

**NF-REQ-143**
The link route MUST reject selectors that reference rejected rows, soft-deleted tables, failed tables, missing rows, `unmapped_raw`, graph layout coordinates, visible graph labels, visible row numbers, or stale graph query digests.

**NF-REQ-143a**
For non-replay requests, after selector resolution and before any mutation or committed idempotency record, the implementation MUST compare `confirm_exact_value` to the resolved candidate value by canonical IP text equality. A mismatch MUST fail with `network_flow_indicator_link_ambiguous`. Because the link route is incident-authorized, the error details MAY include the resolved candidate value. Omission behavior: if a future non-incident-authorized context reuses the error family, it MUST include only `network_flow_safe_digest_v1` of the resolved candidate value.

### 15.2 Duplicate and replay behavior

**Table 15-D. Indicator-link duplicate behavior**

| Case | Required behavior |
| --- | --- |
| Exact idempotency replay | Return original success response. |
| Request resolves to an existing binding identity tuple with new `client_txn_id` | Return existing binding with `duplicate=true`; create no second binding. |
| Request resolves to the same indicator and candidate value from a different source-row-ref set | Create a new binding. |
| Selector references rejected row or unmapped raw | Fail with `network_flow_invalid_indicator_selector`. |

**NF-REQ-144**
A `network_flow_indicator_binding` MUST NOT rewrite the flow row, modify `unmapped_raw`, alter graph identity, create a Core indicator observation, or change indicator lifecycle state automatically. Canonical indicator creation and canonical dedupe remain owned by Core indicator behavior.

**NF-REQ-144a**
Binding identity is the resolved tuple:

```text
(incident_id,
 resolved target canonical indicator record_id,
 canonical candidate_value,
 canonical source-row-ref set)
```

The canonical source-row-ref set is the post-population set from Table 15-F compared as sorted `network_flow_row_id` values. Graph selector ref sets are truncation-dependent and MUST be derived deterministically by Table 15-F and §14.5 ordering.

**Table 15-E. `network_flow_indicator_binding` resource**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `network_flow_indicator_binding_id` | binding ID | Yes | Stable binding ID. |
| `incident_id` | Core incident ID | Yes | Same incident as table and indicator. |
| `target_indicator_ref` | object | Yes | Core indicator reference. |
| `selector_kind` | token | Yes | Selector kind used. |
| `candidate_value` | scalar | Yes | Canonical candidate value linked. |
| `source_row_refs[]` | array | Yes | At least one source row ref required; a graph selector with no contributing accepted rows is invalid. |
| `source_row_refs_truncated` | boolean | Yes | True when retained `source_row_refs[]` omits contributing accepted rows. |
| `source_row_refs_total_count` | positive integer | Yes | Total source-row-ref count before truncation. |
| `duplicate` | boolean | Yes | True only for duplicate response under Table 15-D. |
| `created_observation_refs[]` | array | Yes | Always `[]` in v1. |
| `created_by_user_id` | user ID | Yes | Actor. |
| `created_at` | `timestamp_utc_v1` | Yes | Commit time. |

**Table 15-F. Binding source-row-ref population**

| Selector kind | `source_row_refs[]` population |
| --- | --- |
| `row_field_value` | Exactly the referenced accepted row. |
| `row_refs` | Exactly the supplied refs after duplicate validation. |
| `graph_vertex` | First `min(total, network_flow.max_binding_source_row_refs)` contributing accepted-row refs under §14.5 ordering. |
| `graph_edge` | First `min(total, network_flow.max_binding_source_row_refs)` contributing accepted-row refs under §14.5 ordering. |

For every selector kind, `source_row_refs_total_count` MUST be the total contributing accepted-row-ref count before truncation and `source_row_refs_truncated` MUST be true only when the retained array length is smaller than that total.

## 16. Security, authorization, audit, logging, and egress

### 16.1 Egress and enrichment boundary

**NF-REQ-145**
Network Flow import, mapping, validation, table query, rejected-row query, graph query, indicator-link initiation, and route error handling MUST NOT perform third-party network egress in v1.

**NF-REQ-146**
The implementation MUST NOT contact reputation, geolocation, passive DNS, WHOIS, VirusTotal, MISP, cloud enrichment, vendor telemetry, LLM, or external graph services as part of v1 Network Flow behavior. Omission behavior: enrichment controls are unavailable, and any request attempting enrichment fails with `network_flow_external_enrichment_forbidden`.

### 16.2 Raw value handling

**NF-REQ-147**
Formula-looking CSV values MUST be inert during import. The implementation MUST NOT evaluate formulas, macros, hyperlinks, external references, named ranges, spreadsheet functions, or URI fetches from imported CSV text.

**NF-REQ-148**
Network Flow logs, telemetry, audit summaries, and administrative diagnostics MUST use Table 16-A.

**Table 16-A. Raw-value handling matrix**

| Field class | Logs | Telemetry | Audit summary | Incident-authorized diagnostics |
| --- | --- | --- | --- | --- |
| Raw CSV row | Forbidden | Forbidden | Forbidden | Forbidden |
| Raw CSV cell | Forbidden | Forbidden | Forbidden | Redacted, digest only |
| Source filename | Safe digest only | Safe digest only | Safe digest only | Only `source_filename_display` returned by incident-authorized routes |
| Indicator candidate | Forbidden | Forbidden | Safe digest only | Allowed only through authorized indicator-link route response |
| Graph query raw values | Safe digest only | Safe digest only | Safe digest only | Allowed only in authorized route response |
| Import source bytes | Forbidden | Forbidden | Digest only | Digest only |

**NF-REQ-149**
Plain `raw_value_sha256` MAY appear only inside incident-authorized rejected-row diagnostics. Omission behavior: logs, telemetry, and administrative audit summaries MUST use `network_flow_safe_digest_v1` or omit the value entirely.

### 16.3 Audit event families

**NF-REQ-150**
The implementation MUST emit audit events for the families in Table 16-B using only the safe fields shown.

**Table 16-B. Audit event families**

| Event code | Required safe fields |
| --- | --- |
| `network_flow_table_created` | `incident_id`, `actor_user_id`, `network_flow_table_id`, `source_filename_digest`, `source_content_sha256`, `source_profile_id`, `parser_profile_id`, `mapping_fingerprint`, `row_count_accepted`, `row_count_rejected`. |
| `network_flow_table_renamed` | `incident_id`, `actor_user_id`, `network_flow_table_id`, `old_display_name_digest`, `new_display_name_digest`, `table_version`. |
| `network_flow_table_soft_deleted` | `incident_id`, `actor_user_id`, `network_flow_table_id`, `table_version`, `row_count_accepted`, `row_count_rejected`. |
| `network_flow_graph_query_executed` | `incident_id`, `actor_user_id`, `graph_query_digest`, `selected_table_count`, `result_vertex_count`, `result_edge_count`, `truncated_example_ref_count`. |
| `network_flow_indicator_binding_created` | `incident_id`, `actor_user_id`, `network_flow_indicator_binding_id`, `target_indicator_record_id`, `selector_kind`, `candidate_value_digest`, `source_row_ref_count`, `duplicate`. |

**NF-REQ-151**
Audit event payloads MUST NOT include raw display names, raw filenames, raw CSV cells, raw graph query scalar values, or raw indicator candidates unless the same value is already a Core stable identifier and is safe under Core audit rules.

## 17. Public route family contracts

### 17.1 Route inventory

**NF-REQ-152**
The Network Flow public route inventory is closed to Table 17-A in v1.

**Table 17-A. Route inventory**

| Route ID | Method and path | Operation |
| --- | --- | --- |
| `nf.source_profiles.list` | `GET /api/v1/incidents/{incident_id}/network-flow/source-profiles` | Discover conformant source profiles and effective limits. |
| `nf.tables.list` | `GET /api/v1/incidents/{incident_id}/network-flow/tables` | List active flow tables. |
| `nf.tables.get` | `GET /api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}` | Read active flow table metadata. |
| `nf.tables.patch` | `PATCH /api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}` | Rename active flow table. |
| `nf.tables.delete` | `DELETE /api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}` | Soft-delete active flow table. |
| `nf.tables.query` | `POST /api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}/query` | Query accepted rows for one table. |
| `nf.rejected_rows.query` | `POST /api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}/rejected-rows/query` | Query retained rejected-row diagnostics for one table. |
| `nf.rows.query` | `POST /api/v1/incidents/{incident_id}/network-flow/rows/query` | Query accepted rows across table scope. |
| `nf.graphs.query` | `POST /api/v1/incidents/{incident_id}/network-flow/graphs/query` | Compose graph across table scope. |
| `nf.indicator_links.create` | `POST /api/v1/incidents/{incident_id}/network-flow/indicator-links` | Create or return flow-to-indicator binding. |

**NF-REQ-153**
A route not listed in Table 17-A is unavailable in v1. The implementation MUST NOT expose aliases for the same behavior under another path.

### 17.2 Route contract index

**Table 17-B. Route contract index**

| Route ID | Auth context | Request body or query | Omission/default summary | Replay/idempotency | Success `data` shape | Primary errors | Audit |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `nf.source_profiles.list` | `viewer` | No body; no query params | Pagination unsupported | Read route | `source_profiles[]`, `effective_limits` | `network_flow_invalid_request` | none |
| `nf.tables.list` | `viewer` | No body; no query params | Lists active tables only | Read route | `tables[]` ordered by `created_at`, then ID | `network_flow_invalid_request` | none |
| `nf.tables.get` | `viewer` | No body | Active table only | Read route | `table` | `network_flow_table_not_found`, `network_flow_table_not_active` | none |
| `nf.tables.patch` | `editor` | `client_txn_id`, `base_table_version`, `display_name` | No omitted members | Required | `table` | `network_flow_table_version_conflict`, `network_flow_invalid_display_name` | `network_flow_table_renamed` |
| `nf.tables.delete` | `reviewer` | `client_txn_id`, `base_table_version` | No omitted members | Required | `table` | `network_flow_table_version_conflict`, `network_flow_table_not_active` | `network_flow_table_soft_deleted` |
| `nf.tables.query` | `viewer` | Query request from §13.1 | Defaults from Table 13-A | Read route | `network_flow_table_query_result.v1` | `network_flow_invalid_filter`, `network_flow_invalid_sort`, `network_flow_cursor_invalid` | none |
| `nf.rejected_rows.query` | `viewer` | Diagnostic query from §17.7 | Default `limit=min(200, effective network_flow.max_query_limit)`, no filters | Read route | `network_flow_rejected_rows_query_result.v1` | `network_flow_invalid_filter`, `network_flow_cursor_invalid` | none |
| `nf.rows.query` | `viewer` | `table_scope` plus §13.1 query | Defaults from Table 13-A | Read route | `network_flow_rows_query_result.v1` | `network_flow_invalid_table_scope`, query errors | none |
| `nf.graphs.query` | `viewer` | Graph request from §14.1 | Defaults from Table 14-A | Read route | `network_flow_graph_query_result.v1` | graph/time/limit errors | `network_flow_graph_query_executed` |
| `nf.indicator_links.create` | `editor` | `indicator_link_request_v1` | No omitted members | Required | `network_flow_indicator_link_result.v1` | indicator-link errors | `network_flow_indicator_binding_created` |

### 17.3 Source profile list

**NF-REQ-154**
`GET /source-profiles` MUST return only conformant v1 source profiles in `data.source_profiles[]`. Reserved profiles MAY appear in documentation but MUST NOT appear in this route's returned selectable list. Omission behavior: clients cannot select reserved profiles through discovery.

**NF-REQ-155**
`data.source_profiles[]` MUST be ordered by `source_profile_id ASC`. `data.effective_limits` MUST expose the effective resource-limit values from §20.

**Table 17-B1. Source profile list response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_source_profile_list.v1`. |
| `source_profiles[]` | array | Yes | No | Conformant v1 source profile objects from Table 17-B2 ordered by `source_profile_id ASC`. |
| `effective_limits` | object | Yes | No | Contains every Table 20-A limit key exactly once with the current effective integer value. |
| `meta` | object | Yes | No | Contains exactly `count`, equal to `source_profiles.length`. |

**Table 17-B2. Source profile list item**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `source_profile_id` | token | Yes | No | Claimable profile ID from Table 9-A. |
| `display_name` | string | Yes | No | Stable human-readable name for the profile. |
| `conformance_status` | token | Yes | No | Exactly `required_v1` for every returned item in v1. |
| `default_parser_profile_id` | token | Yes | No | `rfc4180_headered_csv_v1`. |
| `required_field_keys[]` | array | Yes | No | Required fields in Table 9-D order. |
| `optional_field_keys[]` | array | Yes | No | Optional source-mappable fields in Table 9-D order. |
| `system_derived_field_keys[]` | array | Yes | No | System-derived fields in Table 9-D order. |
| `supported_timestamp_modes[]` | array | Yes | No | `["rfc3339","epoch_seconds","epoch_milliseconds","netflow_sys_uptime_milliseconds"]` in that order. |

In v1, `source_profiles[]` MUST contain exactly one item: `source_profile_id='cisco_sna_netflow_csv_v1'`, `display_name='Cisco Secure Network Analytics NetFlow CSV'`, `conformance_status='required_v1'`, `default_parser_profile_id='rfc4180_headered_csv_v1'`, `required_field_keys[]` equal to the `required` Cisco SNA fields in Table 9-D order, `optional_field_keys[]` equal to the `optional_map_when_present` Cisco SNA fields in Table 9-D order, and `system_derived_field_keys[]=["network_flow.observation_source_ref"]`.

### 17.4 Table list and get

**NF-REQ-156**
`GET /tables` MUST return active tables only, ordered by `created_at ASC`, then `network_flow_table_id ASC`. It has no pagination and no non-active inclusion mode in v1.

**NF-REQ-157**
`GET /tables/{network_flow_table_id}` MUST return active table metadata only. A soft-deleted, failed, creating, unknown, cross-incident, or hidden table MUST NOT leak cross-incident existence. The implementation MUST use Core hidden-resource behavior when hiding is required by Core authorization rules.

**Table 17-B3. Table list response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_table_list.v1`. |
| `tables[]` | array of `network_flow_table` | Yes | No | Active tables ordered by `created_at ASC`, then `network_flow_table_id ASC`. |
| `meta` | object | Yes | No | Contains exactly `count`, equal to `tables.length`; pagination members are absent. |

**Table 17-B4. Table get response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_table_get.v1`. |
| `table` | `network_flow_table` | Yes | No | Active table resource from Table 8-B. |

**Table 17-B5. Table mutation response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_table_mutation_result.v1`. |
| `table` | `network_flow_table` | Yes | No | Post-mutation table resource. Delete responses return the same resource with `table_status='soft_deleted'` and non-null `deleted_at`. |

### 17.5 Rename route

**NF-REQ-158**
`PATCH /tables/{network_flow_table_id}` request body MUST contain exactly Table 17-C members.

**Table 17-C. Rename request body**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `client_txn_id` | `client_txn_id` | Yes | No | Idempotency key. |
| `base_table_version` | positive integer | Yes | No | Must equal current table version. |
| `display_name` | string | Yes | No | Normalized by display-name rules; non-empty after trimming; max 64 Unicode scalars after normalization. |

**NF-REQ-159**
Rename with stale `base_table_version` MUST fail with `network_flow_table_version_conflict` and include current table version in `error.details`. Exact replay of a committed rename MUST return the original success response.

### 17.6 Soft-delete route

**NF-REQ-160**
`DELETE /tables/{network_flow_table_id}` request body MUST contain exactly `client_txn_id` and `base_table_version`. The route performs soft delete only. Exact replay returns the original success response. A second non-replay delete of the same table fails with `network_flow_table_not_active`.

### 17.7 Rejected-row diagnostic query

**NF-REQ-161**
The rejected-row diagnostic query request MUST use Table 17-D.

**Table 17-D. Rejected-row query request**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `error_codes[]` | array | No | No | `[]` | Empty means no error-code filter. When non-empty, length `1..64`; values from §21; duplicates and unknown tokens invalid. |
| `field_keys[]` | array | No | No | `[]` | Empty means no field-key filter. When non-empty, length `1..64`; values from Table 9-D; duplicates and unknown tokens invalid. |
| `source_row_range` | object | No | No | omitted | Optional `{ "gte": positive_int\|null, "lte": positive_int\|null }`; at least one non-null when present; `gte > lte` invalid with `reason_code='empty_range'`. |
| `limit` | integer | No | No | `min(200, effective network_flow.max_query_limit)` | `1..effective network_flow.max_query_limit`. |
| `cursor_token` | string | No | No | omitted | Opaque continuation token. |

Invalid rejected-row query arrays, duplicate values, unknown tokens, and empty ranges MUST fail with `network_flow_invalid_filter`.

### 17.8 Query response shapes

**NF-REQ-162**
A single-table accepted-row query response `data` MUST contain Table 17-E1 members.

**Table 17-E1. Single-table accepted-row query response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_table_query_result.v1`. |
| `network_flow_table_id` | table ID | Yes | No | Queried active table. |
| `rows[]` | array of `network_flow_row` | Yes | No | Accepted rows ordered by `effective_sort_v1`; each row includes `network_flow_table_id`. |
| `meta` | object | Yes | No | Exact object from Table 17-E4. |

**NF-REQ-163**
A cross-table accepted-row query response `data` MUST contain Table 17-E2 members.

**Table 17-E2. Cross-table accepted-row query response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_rows_query_result.v1`. |
| `table_scope` | normalized table scope object | Yes | No | Resolved scope with `table_ids[]` sorted by `network_flow_table_id ASC`. |
| `rows[]` | array of `network_flow_row` | Yes | No | Accepted rows ordered by `effective_sort_v1`; every row includes `network_flow_table_id`. |
| `meta` | object | Yes | No | Exact object from Table 17-E4. |

**NF-REQ-164**
A rejected-row query response `data` MUST contain Table 17-E3 members.

**Table 17-E3. Rejected-row query response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_rejected_rows_query_result.v1`. |
| `network_flow_table_id` | table ID | Yes | No | Queried active table. |
| `diagnostics[]` | array of rejected-row diagnostics | Yes | No | Ordered by §12.4, after filters and cursor continuation. |
| `meta` | object | Yes | No | Exact object from Table 17-E4; `query.effective_sort` uses §12.4 diagnostic ordering. |

**Table 17-E4. Query response `meta`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `query` | object | Yes | No | Contains exactly `normalized_request`, `filters`, `sort`, `effective_sort`, and `table_ids[]`. `normalized_request` excludes `cursor_token` and includes default-materialized non-cursor query members. `filters` and `sort` are default-materialized caller inputs. |
| `paging` | object | Yes | No | Contains exactly `limit`, `returned_count`, and `next_cursor_token`. `next_cursor_token` follows NF-REQ-126c. |

Query response objects MUST NOT include raw cursor internals, SQL fragments, storage field names, table display names as identity, or visual row positions.

### 17.9 Indicator link response

**NF-REQ-165**
A successful indicator-link response `data` MUST contain Table 17-E members.

**Table 17-E. Indicator link response data**

| Member | Type | Required |
| --- | --- | ---: |
| `schema_id` | string `cartulary.network_flow_indicator_link_result.v1` | Yes |
| `binding` | `network_flow_indicator_binding` | Yes |
| `target_indicator_ref` | Core indicator reference | Yes |
| `created_observation_refs[]` | array | Yes; always `[]` in v1 |
| `duplicate` | boolean | Yes |

## 18. Import apply result integration

**NF-REQ-166**
A terminal successful import apply that creates Network Flow tables MUST return or expose extension result references for every created table. Each table reference MUST contain at least the members in Table 18-A.

**Table 18-A. Import result table reference**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `kind` | string | Yes | Exactly `network_flow_table`. |
| `id` | table ID | Yes | Created table ID. |
| `route` | string | Yes | `/api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}` with concrete IDs in the response. |
| `display_name` | string | Yes | Table display name. |
| `row_count_accepted` | integer | Yes | Accepted count. |
| `row_count_rejected` | integer | Yes | Rejected count. |

**NF-REQ-167**
Replay of a committed import apply action with the same route-scoped idempotency key and normalized request MUST return the same created table references and MUST NOT create additional tables.

## 19. Graph UI interaction, pivots, and table synchronization

**NF-REQ-168**
Selecting a graph vertex MUST pivot to contributing accepted rows by stable `vertex_id`, `graph_query_digest`, and recomputed normalized graph query. The pivot MUST NOT use graph layout coordinates, visible labels, rendered order, or browser-local graph node indices.

**NF-REQ-169**
Selecting a graph edge MUST open the contributing rows drawer grouped by table. Group ordering MUST follow workspace table order. Row ordering inside a group MUST follow `effective_sort_v1([])` from §13.4. A user-supplied pivot sort is a subsequent row-query request and MUST use §13 query sorting; it is not part of graph selector identity.

**NF-REQ-170**
When any active table in a displayed graph's table scope is renamed, the graph data remains semantically valid but display metadata may be stale. The UI MUST update table display labels on the next table-metadata refresh without changing graph query digest. When any active table in graph scope is soft-deleted, the displayed graph MUST become `graph_stale` and any selector action against it MUST fail until the graph is recomputed without the deleted table.

## 20. Resource limits

**NF-REQ-171**
Network Flow resource limits MUST use Table 20-A. Deployments MAY lower limits to the lowerable minimum. Omission behavior: deployments that omit a limit use the default. Deployments MUST NOT raise a limit above the default in this revision.

Limit configuration MUST be validated at process configuration load before serving Network Flow routes or import facade calls. A configured value that is absent uses the default. A configured value that is not an integer, is below the lowerable minimum, or is above the default-and-maximum value is invalid configuration. The implementation MUST fail configuration validation and MUST NOT silently clamp, ignore, or round the value.

**Table 20-A. Resource limit registry**

| Limit key | Default and maximum | Lowerable minimum | Enforcement phase | Exceeded behavior |
| --- | ---: | ---: | --- | --- |
| `network_flow.max_active_tables_per_incident` | 128 | 1 | Table activation. | Fail with `network_flow_table_limit_exceeded`; commit no new active table. |
| `network_flow.max_retained_tables_per_incident` | 512 | 1 | Table allocation. | Fail with `network_flow_table_limit_exceeded`; allocate no retained table. |
| `network_flow.max_selected_tables_per_graph` | 64 | 1 | Graph query admission. | Fail with `network_flow_invalid_table_scope` and `reason_code='selected_table_limit_exceeded'`; emit no graph output. |
| `network_flow.max_columns_per_csv` | 512 | 1 | Header decode. | Fail import unit with `network_flow_resource_limit_exceeded`; allocate no table. |
| `network_flow.max_header_scalar_length` | 256 | 1 | Header decode. | Fail import unit with `network_flow_resource_limit_exceeded`; allocate no table. |
| `network_flow.max_raw_cell_scalar_length` | 16384 | 1 | Field decode. | Reject the affected row with `network_flow_resource_limit_exceeded`; do not retain the oversized raw value. |
| `network_flow.max_rows_per_csv` | 5000000 | 1 | CSV streaming. | Fail import unit with `network_flow_resource_limit_exceeded`; activate no table. |
| `network_flow.max_accepted_rows_per_table` | 5000000 | 1 | Row validation. | Fail import unit with `network_flow_resource_limit_exceeded` when the next accepted row would exceed the limit; activate no table. |
| `network_flow.max_rejected_row_diagnostics` | 100000 | 0 | Diagnostic retention. | Retain only the first `N` diagnostics under §12.4 ordering and set `diagnostics_truncated=true`; not an error by itself. |
| `network_flow.max_filters_per_query` | 16 | 0 | Query admission. | Fail with `network_flow_invalid_filter` and `reason_code='too_many_filters'`; return no rows. |
| `network_flow.max_sorts_per_query` | 8 | 0 | Query admission. | Fail with `network_flow_invalid_sort` and `reason_code='too_many_sorts'`; return no rows. |
| `network_flow.max_query_limit` | 1000 | 1 | Query admission. | Fail explicit out-of-range query limit with `network_flow_invalid_limit`; omitted limit materializes to §13.1 default. |
| `network_flow.max_graph_vertices` | 100000 | 1 | Graph construction before response. | Fail with `network_flow_graph_limit_exceeded`; return no partial graph. |
| `network_flow.max_graph_edges` | 250000 | 0 | Graph construction before response. | Fail with `network_flow_graph_limit_exceeded`; return no partial graph. |
| `network_flow.max_example_row_refs_per_edge` | 100 | 0 | Graph response construction. | Truncate `example_row_refs[]` and set `example_refs_truncated`; not an error by itself. |
| `network_flow.max_binding_source_row_refs` | 1000 | 1 | Indicator-link commit. | Truncate binding `source_row_refs[]` and set `source_row_refs_truncated`; not an error by itself. |
| `network_flow.max_counter_sum_digits` | 128 | 1 | Aggregation. | Fail with `network_flow_counter_sum_limit_exceeded`; return no partial graph. |

**NF-REQ-172**
Every limit-exceeded error MUST include `limit_key`, `limit`, `actual`, and `phase` in `error.details`. If the actual value cannot be safely disclosed without revealing incident content, `actual` MUST be the smallest safe lower bound known to exceed the limit.

**NF-REQ-173**
Effective limits MUST be exposed through `GET /source-profiles` under `data.effective_limits`. The route MUST return all keys in Table 20-A exactly once.

## 21. Error registry and detail schemas

### 21.1 Error code registry

**NF-REQ-174**
Network Flow errors MUST use Table 21-A where route-local errors are required.

**Table 21-A. Error code registry**

| Error code | Default transport | Use |
| --- | ---: | --- |
| `network_flow_invalid_request` | 400 | JSON admission, unknown member, type mismatch, or route-local schema failure. |
| `network_flow_unsupported_source_profile` | 400 | Reserved or unsupported profile requested. |
| `network_flow_invalid_utf8` | 400 | CSV source is not valid UTF-8. |
| `network_flow_csv_empty_file` | 400 | CSV source has zero bytes after BOM removal. |
| `network_flow_no_header_row` | 400 | Header row missing. |
| `network_flow_no_data_rows` | 400 | CSV contains header but no data rows. |
| `network_flow_csv_malformed_quote` | 400 | Quote grammar invalid. |
| `network_flow_csv_field_count_mismatch` | row diagnostic | Row field count differs from header. |
| `network_flow_mapping_required` | 400 | Mapping incomplete or invalid. |
| `network_flow_mapping_conflict` | 400 | Mapping rows conflict or persisted defaults invalid. |
| `network_flow_invalid_timestamp` | row diagnostic or 400 | Timestamp parse/profile failure. |
| `network_flow_end_before_start` | row diagnostic | Flow end earlier than start. |
| `network_flow_invalid_ip` | row diagnostic or 400 | IP parse/canonicalization failure. |
| `network_flow_invalid_port` | row diagnostic or 400 | Port invalid. |
| `network_flow_invalid_protocol` | row diagnostic or 400 | Protocol invalid. |
| `network_flow_invalid_counter` | row diagnostic or 400 | Counter invalid. |
| `network_flow_all_rows_rejected` | 400 | Import validation accepted zero rows. |
| `network_flow_table_limit_exceeded` | 409 | Table active or retained limit exceeded. |
| `network_flow_resource_limit_exceeded` | 413 or row diagnostic | Parser, import, row-validation, or raw-cell limit exceeded when no more specific table, graph, query, or counter limit code applies. |
| `network_flow_table_name_exhausted` | 409 | Display-name suffix space exhausted. |
| `network_flow_table_not_found` | 404 | Table not found or hidden under route policy. |
| `network_flow_table_not_active` | 409 | Table state not active for route. |
| `network_flow_table_version_conflict` | 409 | `base_table_version` stale. |
| `network_flow_invalid_display_name` | 400 | Rename display name invalid. |
| `network_flow_invalid_table_scope` | 400 | Table scope invalid. |
| `network_flow_invalid_filter` | 400 | Filter invalid. |
| `network_flow_invalid_sort` | 400 | Sort invalid. |
| `network_flow_invalid_limit` | 400 | Query limit invalid. |
| `network_flow_cursor_invalid` | 400 | Cursor invalid, stale, wrong scope, expired, or malformed. |
| `network_flow_invalid_time_range` | 400 | Time range invalid or empty. |
| `network_flow_time_bucket_limit_exceeded` | 400 | Reserved; no v1 trigger. |
| `network_flow_invalid_limit_override` | 400 | Graph limit override invalid. |
| `network_flow_graph_limit_exceeded` | 413 | Graph exceeds vertex or edge limit. |
| `network_flow_counter_sum_limit_exceeded` | 413 | Aggregated counter sum exceeds digit limit. |
| `network_flow_aggregation_mode_unavailable` | 400 | Reserved; no v1 trigger. |
| `network_flow_indicator_link_ambiguous` | 400 | Link action does not identify candidate value. |
| `network_flow_invalid_indicator_selector` | 400 | Link selector invalid or stale. |
| `network_flow_indicator_link_forbidden` | 403 | Authorization or target visibility fails. |
| `network_flow_external_enrichment_forbidden` | 400 | Request attempts forbidden third-party enrichment. |

### 21.2 Error detail schema registry

**NF-REQ-175**
Route-local errors MUST include details according to Table 21-B. A detail member listed as required MUST be present even when its value is JSON `null`, unless the row states that the member is conditionally present.

**Table 21-B. Error detail schemas**

| Error code or scope | Required `error.details` members |
| --- | --- |
| `network_flow_invalid_request` | `field`, `reason_code`, `expected_contract`, `actual_kind`. |
| `network_flow_unsupported_source_profile` | `source_profile_id`, `conformance_status`, `allowed_profile_ids[]`. |
| CSV import-unit failures: `network_flow_invalid_utf8`, `network_flow_csv_empty_file`, `network_flow_no_header_row`, `network_flow_no_data_rows`, `network_flow_csv_malformed_quote` | `phase`, `reason_code`, `source_row_number`, `source_column_ordinal`. Row and column members are `null` when not attributable. |
| CSV row diagnostics: `network_flow_csv_field_count_mismatch`, row-level `network_flow_invalid_timestamp`, `network_flow_end_before_start`, row-level `network_flow_invalid_ip`, row-level `network_flow_invalid_port`, row-level `network_flow_invalid_protocol`, row-level `network_flow_invalid_counter`, row-level `network_flow_resource_limit_exceeded` | `source_row_number`, `source_column_ordinal`, `field_key`, `reason_code`, `safe_sample`, `raw_value_sha256`. |
| Mapping failures: `network_flow_mapping_required`, `network_flow_mapping_conflict` | `field_key`, `source_column_ordinal`, `mapping_kind`, `reason_code`. |
| `network_flow_all_rows_rejected` | `row_count_rejected`, `diagnostics_truncated`, `diagnostics_sample[]`. |
| Limit failures: `network_flow_table_limit_exceeded`, `network_flow_resource_limit_exceeded`, `network_flow_graph_limit_exceeded`, `network_flow_counter_sum_limit_exceeded` | `limit_key`, `limit`, `actual`, `phase`. |
| `network_flow_table_name_exhausted` | `base_display_name_digest`, `attempted_suffix_min`, `attempted_suffix_max`. |
| Table lookup/state failures: `network_flow_table_not_found`, `network_flow_table_not_active` | `network_flow_table_id`, `table_status`, `allowed_states[]`. Hidden resources MUST use Core hidden-resource details instead when Core requires non-disclosure. |
| `network_flow_table_version_conflict` | `network_flow_table_id`, `base_table_version`, `current_table_version`. |
| `network_flow_invalid_display_name` | `field`, `reason_code`, `max_length`, `normalized_length`. |
| `network_flow_invalid_table_scope` | `reason_code`, `mode`, `table_ids[]`, `limit_key`. |
| `network_flow_invalid_filter` | `field_key`, `op`, `reason_code`, `filter_index`. |
| `network_flow_invalid_sort` | `field_key`, `direction`, `reason_code`, `sort_index`. |
| `network_flow_invalid_limit` | `limit_key`, `limit`, `minimum`, `maximum`, `reason_code`. |
| `network_flow_cursor_invalid` | `reason_code`, `cursor_scope`, `retry`. |
| `network_flow_invalid_time_range` | `field`, `reason_code`, `start_utc`, `end_utc`. |
| Reserved no-v1-trigger codes: `network_flow_time_bucket_limit_exceeded`, `network_flow_aggregation_mode_unavailable` | `reason_code`. These codes MUST NOT be emitted in v1. |
| `network_flow_invalid_limit_override` | `limit_key`, `limit`, `minimum`, `maximum`, `reason_code`. |
| Indicator-link failures: `network_flow_indicator_link_ambiguous`, `network_flow_invalid_indicator_selector`, `network_flow_indicator_link_forbidden` | `selector_kind`, `field_key`, `target_mode`, `reason_code`, and either `resolved_candidate_value` for incident-authorized route responses or `resolved_candidate_safe_digest` otherwise. |
| `network_flow_external_enrichment_forbidden` | `requested_capability`, `reason_code`. |

**NF-REQ-176**
When multiple errors apply, the implementation MUST report the first applicable error family under Table 21-C precedence.

**Table 21-C. Error precedence**

| Precedence | Error family |
| ---: | --- |
| 1 | Authentication/session failure. |
| 2 | Hidden incident or authorization failure. |
| 3 | Extension unclaimed. |
| 4 | Malformed JSON/admission failure. |
| 5 | Path parameter validation. |
| 6 | Request schema validation. |
| 7 | Resource/lifecycle validation. |
| 8 | Semantic validation. |
| 9 | Limit failure. |

For Network Flow-owned mutating routes, exact committed idempotency replay lookup occurs at the Table 5-A replay point and returns the original success before Table 21-C resource/lifecycle, semantic, or limit failures for the current resource state. A same-tuple different-digest `client_txn_conflict` is reported at that same point.

**Table 21-D. Route-local first-error ordering refinements**

| Scope | Required order inside the same Table 21-C precedence level |
| --- | --- |
| JSON objects | Validate members in canonical member-name order after detecting duplicate members. Report the first unknown, missing, explicit-null, or type-invalid member in that order. |
| Request arrays | Validate array length first, then items in input order. |
| `filters[]` | For each filter in input order, validate `field_key`, then `op`, then `value`. |
| `sort[]` | Validate duplicate field keys before field sortability; otherwise validate items in input order. |
| `field_mappings[]` | Validate `mapping_kind`, variant member closure, source column existence, target field compatibility, transform, empty-value policy, then cross-row conflicts. |
| Rejected-row diagnostics | Order diagnostics only by §12.4; do not use validation discovery order. |
| Indicator-link selectors | Validate selector kind, required selector members, table/row/graph freshness, linkable field, candidate resolution, `confirm_exact_value`, target compatibility, then duplicate binding. |

## 22. Fixtures

**NF-REQ-177**
Conformance fixtures MUST include Table 22-A before this NLSpec can be adopted. While this document remains draft, fixture rows MAY contain explicit `TODO:` values. Omission behavior: a fixture row with any `TODO:` value is a known adoption blocker and MUST NOT satisfy adoption or implementation conformance.

**Table 22-A. Fixture bundle registry**

| Fixture ID | Path | SHA-256 | Source profile | Parser profile | Required expected outputs |
| --- | --- | --- | --- | --- | --- |
| `NF-FIX-001-cisco-sna-minimal` | `TODO: fixtures/network-flow/cisco-sna-minimal.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Table resource, mapping fingerprint, exact row IDs, zero diagnostics, graph result. |
| `NF-FIX-002-cisco-sna-interface-fields` | `TODO: fixtures/network-flow/cisco-sna-interface-fields.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Interface field mapping outputs, exact row IDs, graph result. |
| `NF-FIX-003-duplicate-headers` | `TODO: fixtures/network-flow/duplicate-headers.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Source-column descriptors proving ordinal disambiguation. |
| `NF-FIX-004-rejected-rows` | `TODO: fixtures/network-flow/rejected-rows.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Invalid IP, port, protocol, timestamp, counter, field-count, and end-before-start diagnostics in exact order. |
| `NF-FIX-005-csv-parser-edges` | `TODO: fixtures/network-flow/csv-parser-edges.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Terminal newline, blank line, quoted newline, quote escaping, and malformed quote outcomes. |
| `NF-FIX-006-cross-table-graph` | `TODO: fixtures/network-flow/cross-table-graph/` | `TODO: sha256 manifest` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Two table resources, shared endpoint vertex, aggregated edge, exact graph query digest, exact source snapshot ID, and exact edge IDs. |
| `NF-FIX-007-indicator-linking` | `TODO: fixtures/network-flow/indicator-linking.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Existing-indicator link, create-indicator link, duplicate binding result. |
| `NF-FIX-008-large-limits` | `TODO: fixtures/network-flow/large-limits/` | `TODO: sha256 manifest` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Graph/table limit failures; engineering measurement only, not Core 05 publication evidence. |
| `NF-FIX-009-soft-delete-stale-graph` | `TODO: fixtures/network-flow/soft-delete-stale-graph.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Active graph query containing a table that is then soft-deleted; stale graph and cursor invalidation. |
| `NF-FIX-010-json-admission` | `TODO: fixtures/network-flow/json-admission.jsonl` | `TODO: sha256` | n/a | n/a | Duplicate member, explicit null, unknown member, malformed JSON, non-object body failures. |
| `NF-FIX-011-alias-collision` | `TODO: fixtures/network-flow/alias-collision.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Alias match keys, duplicate-alias warning, explicit approved mapping, and mapping conflict for source reuse. |
| `NF-FIX-012-sys-uptime-timestamps` | `TODO: fixtures/network-flow/sys-uptime-timestamps.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Export time, exporter uptime-at-export, start/end event uptime derivations, and wrap-ambiguous rejection. |
| `NF-FIX-013-filename-display` | `TODO: fixtures/network-flow/filename-display.jsonl` | `TODO: sha256` | n/a | n/a | Path stripping, hidden-file stem, trailing-dot stem, display-name override rejection, duplicate suffixing, and soft-delete name reuse. |
| `NF-FIX-014-cursor-pagination` | `TODO: fixtures/network-flow/cursor-pagination.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Mandatory default sort tail, keyset continuation tuple, terminal null cursor, actor mismatch, and table rename cursor survival. |
| `NF-FIX-015-graph-adapter-input` | `TODO: fixtures/network-flow/graph-adapter-input/` | `TODO: sha256 manifest` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Exact `network_flow_graph_projection_input_v1`, graph view key, source snapshot ID, property definitions, and safe metadata. |
| `NF-FIX-016-redaction` | `TODO: fixtures/network-flow/redaction.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Deterministic `safe_sample` nulls, integer-like samples, raw value SHA-256s, audit safe digests, and no raw text leakage. |
| `NF-FIX-017-indicator-link-mismatch` | `TODO: fixtures/network-flow/indicator-link-mismatch.csv` | `TODO: sha256` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Non-IP field rejection, edge field disambiguation, `confirm_exact_value` mismatch, existing-indicator normalized-value mismatch, and create-indicator Core dependency failure. |
| `NF-FIX-018-resource-limits` | `TODO: fixtures/network-flow/resource-limits/` | `TODO: sha256 manifest` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Parser/import resource limit failures, diagnostic truncation, graph limit failure, counter digit limit failure, and invalid deployment config cases. |

**NF-REQ-178**
Each fixture bundle MUST include canonical expected-output transcripts for route success `data` objects, route error payloads, table resources, source-column descriptors, approved mapping JSON, mapping fingerprint, row IDs, row digests, diagnostics, graph result, Graph Projection adapter input, indicator-link result, resource-limit details, and redaction outputs where applicable. Fixture graph digests and source snapshot IDs MUST be independent of deployment limit configuration. Fixture-only safe digest expectations MUST declare a deterministic fixture `deployment_audit_secret`; production deployments MUST NOT use that fixture secret.

## 23. Acceptance criteria

**NF-REQ-179**
An implementation claiming `network_flow_activity` MUST satisfy every acceptance criterion in Table 23-A.

**Table 23-A. Binary acceptance criteria**

| ID | Criterion |
| --- | --- |
| `NF-AC-001` | When the extension is unclaimed, the Network Analysis tab and `/api/v1/incidents/{incident_id}/network-flow/*` routes are unavailable through claimed-extension behavior. |
| `NF-AC-002` | When the extension is claimed and an incident has no flow tables, the Network Analysis tab renders the empty state with `Import NetFlow CSV` for an authorized importer. |
| `NF-AC-003` | Every Network Flow route uses Core success/error envelopes and rejects unknown request members unless the route declares them. |
| `NF-AC-004` | Duplicate JSON object members at any depth are rejected before durable state or idempotency allocation. |
| `NF-AC-005` | Invalid UTF-8, malformed JSON, non-object bodies, explicit invalid nulls, and unknown members fail request admission deterministically. |
| `NF-AC-006` | Importing `NF-FIX-001-cisco-sna-minimal` through the Import Extension flow creates exactly one active `network_flow_table` and one inner table tab. |
| `NF-AC-007` | The created table tab display name is derived deterministically from the uploaded filename and collision suffix rules. |
| `NF-AC-008` | Import apply replay with the same committed `client_txn_id` returns the same table ID and creates no second table. |
| `NF-AC-009` | Duplicate CSV headers are displayed and mapped by source column ordinal, not by label alone. |
| `NF-AC-010` | Cisco SNA required fields are enforced exactly for `cisco_sna_netflow_csv_v1`. |
| `NF-AC-011` | Reserved source profiles are not claimable and fail with `network_flow_unsupported_source_profile`. |
| `NF-AC-012` | Empty CSV, missing header, header-only CSV, terminal newline, blank line, malformed quote, quote escaping, and field-count mismatch match §9.2 outcomes. |
| `NF-AC-013` | Formula-looking CSV values are inert during import and never evaluated. |
| `NF-AC-014` | Invalid rows produce deterministic diagnostics and do not appear in table query or graph results. |
| `NF-AC-015` | A partially valid CSV creates a table containing accepted rows and exposes rejected-row counts. |
| `NF-AC-016` | An all-invalid CSV creates no table tab and returns `network_flow_all_rows_rejected`. |
| `NF-AC-017` | `source_row_digest_v1`, `normalized_row_digest_v1`, `mapping_fingerprint_v1`, `network_flow_graph_query_digest_v1`, and `network_flow_source_snapshot_digest_v1` produce exact expected fixture digests or IDs. |
| `NF-AC-018` | Timestamp profile parsing rejects DST folds, DST gaps, leap seconds, invalid local-time inference, and end-before-start rows. |
| `NF-AC-019` | IPv6 outputs follow the canonical text contract, IPv4 leading-zero octets are rejected, and IPv6 zone identifiers are rejected. |
| `NF-AC-020` | `uint64_decimal_string_v1` rejects signs, exponents, decimals, leading zeroes except `0`, and values above max. |
| `NF-AC-021` | The table lifecycle state machine contains no `renamed` state. |
| `NF-AC-022` | Renaming a table changes only display metadata, increments `table_version`, and does not change row IDs, graph IDs, provenance, mapping fingerprints, diagnostics, live query cursor validity, or indicator bindings. |
| `NF-AC-023` | Soft-deleting a table removes it from the default tab strip, removes it from `all_active_tables`, invalidates active graph queries and cursors that include it, and leaves it terminal. |
| `NF-AC-024` | A table query uses stable `field_key` filters and rejects visible column labels. |
| `NF-AC-025` | Filter `in` arrays reject duplicates and empty arrays. |
| `NF-AC-026` | CIDR filtering rejects IPv4/IPv6 family mismatch and does not treat IPv4-mapped IPv6 as IPv4. |
| `NF-AC-027` | Sort null ordering, IP ordering, decimal counter ordering, duplicate sort-key rejection, `network_flow.endpoint_ip` sort rejection, and default sort tail match §13.4. |
| `NF-AC-028` | Cursor continuation fails after TTL expiry, actor mismatch, route mismatch, query mismatch, table soft delete, or authorization loss. |
| `NF-AC-029` | The filter selector always includes explicit table-selection controls in graph mode. |
| `NF-AC-030` | Default graph mode from a table tab uses `table_scope.mode='active_table'`. |
| `NF-AC-031` | `selected_tables` with duplicate table IDs is rejected rather than silently deduplicated. |
| `NF-AC-032` | Selecting multiple tables composes one graph using only rows from those tables after filters and time slicing. |
| `NF-AC-033` | The same canonical endpoint IP across two selected tables emits one default vertex with both contributing table IDs. |
| `NF-AC-034` | Default edges aggregate by `(src_ip, dst_ip, ip_protocol, dst_port)` and preserve contributing table provenance. |
| `NF-AC-035` | Time interval filtering uses overlap semantics and handles zero-duration flows exactly as specified. |
| `NF-AC-036` | Selecting a graph vertex pivots to contributing rows without using visible labels, graph coordinates, row order, or layout state as identity. |
| `NF-AC-037` | Selecting a graph edge opens the cross-table contributing rows drawer grouped by table and ordered deterministically. |
| `NF-AC-038` | Graph over-limit cases fail with deterministic errors and do not render partial unlabeled graphs. |
| `NF-AC-039` | `example_row_refs[]`, `example_refs_truncated`, and `example_refs_total_count` match §14.5. |
| `NF-AC-040` | Graph Projection adapter input contains the exact adapter fields and safe metadata in §14.4. |
| `NF-AC-041` | Linking an endpoint to an existing indicator creates or returns a source-bound binding and does not rewrite the flow row. |
| `NF-AC-042` | Creating an indicator from an endpoint uses the Core indicator owner behavior and fails if the caller lacks the required incident role. |
| `NF-AC-043` | Duplicate indicator-link requests return the existing binding with `duplicate=true` and create no second binding. |
| `NF-AC-044` | Selectors referencing rejected rows, soft-deleted tables, stale graph digests, or `unmapped_raw` fail. |
| `NF-AC-045` | A `deployment_admin` without incident membership cannot list, query, import, graph, soft-delete, rename, or link flow data in that incident. |
| `NF-AC-046` | No network-flow import, graph, table query, rejected-row query, or link action performs third-party egress in v1. |
| `NF-AC-047` | Logs, telemetry, audit summaries, and diagnostics obey the raw-value handling matrix in §16.2. |
| `NF-AC-048` | Effective limits are discoverable, lowerable only to declared minimums, and enforced at the phases in §20. |
| `NF-AC-049` | Every route-local error includes the detail schema required by §21.2. |
| `NF-AC-050` | Large-fixture timing results are classified as engineering measurements unless Core 05 claim-publication requirements are separately satisfied. |
| `NF-AC-051` | `unmapped_raw` values are retained as inert provenance under the default policy and are not filterable, sortable, graphable, or linkable. |
| `NF-AC-052` | Every fixture row has concrete path, byte SHA-256, approved mapping JSON, mapping fingerprint, expected IDs, expected diagnostics, and expected graph/link output before adoption. |
| `NF-AC-053` | The document contains no `MAY` statement whose omission behavior is absent. |
| `NF-AC-054` | No behavior is moved into appendices, research reports, UI guides, implementation guides, or vendor documentation as normative authority. |
| `NF-AC-055` | Internal section references, table references, error-code references, and requirement references resolve without dangling anchors before adoption. |
| `NF-AC-056` | `network_flow_flow_edge_id_v1` produces exact fixture edge IDs, and null destination ports form a distinct aggregation and edge-ID key. |
| `NF-AC-057` | Request members `time_range.bucket` and `aggregation.include_time_buckets` fail admission in v1, and `network_flow_time_bucket_limit_exceeded` has no v1 trigger. |
| `NF-AC-058` | `observation_mode` values other than `binding_only` fail, and `created_observation_refs[]` is always `[]` in binding resources and link responses. |
| `NF-AC-059` | Graph query digests and source snapshot IDs remain unchanged when deployment graph limits or caller `limit_overrides` change without changing query semantics. |
| `NF-AC-060` | Timestamp profile precision rejects finer source resolution, epoch modes reject fractional values, and sys-uptime parsing uses export time, exporter uptime-at-export, and per-field event uptime exactly as §9.7 specifies. |
| `NF-AC-061` | Duplicate table rename fails with `network_flow_invalid_display_name` and `reason_code='duplicate_display_name'`, while existing cursors over the renamed table continue. |
| `NF-AC-062` | Direct references to `soft_deleted`, `creating`, and `failed` tables map to the errors in Table 8-B1, and table lifecycle states count against limits according to Table 8-C. |
| `NF-AC-063` | `network_flow_all_rows_rejected` includes ordered `diagnostics_sample[]`, `row_count_rejected`, and `diagnostics_truncated` details without creating a table. |
| `NF-AC-064` | Omitted query limits materialize to `min(200, effective max_query_limit)`, explicit query limits outside range fail, and graph limit overrides outside `[lowerable_min, effective]` fail. |
| `NF-AC-065` | Validation preview counts are limited to the parser preview slice, and an apply may fail with `network_flow_all_rows_rejected` even when preview accepted rows. |
| `NF-AC-066` | Binding `source_row_refs[]`, `source_row_refs_truncated`, and `source_row_refs_total_count` are populated per Table 15-F, and duplicate `row_refs[]` selector entries fail. |
| `NF-AC-067` | Indicator-link dedupe uses the resolved binding identity tuple; structurally different selectors resolving to the same tuple dedupe, and different source-row-ref sets create distinct bindings. |
| `NF-AC-068` | `confirm_exact_value` mismatch fails before mutation with `network_flow_indicator_link_ambiguous` and the required indicator-link error details. |
| `NF-AC-069` | Graph Projection adapter input uses Network Flow endpoint IDs as source entity IDs, flow edge IDs as source relationship IDs, and no invalid `ephemeral_response_only` retention token. |
| `NF-AC-070` | Interface fields store bounded text or null only; numeric interface identifiers remain text and sort by code point. |
| `NF-AC-071` | `network_flow_aggregation_mode_unavailable` is reserved with no v1 trigger, and mapping combinability accepts only `single_source_only` in v1. |
| `NF-AC-072` | Table rename, table delete, indicator-link create, and Core import apply replay follow the idempotency comparison and replay points in Table 5-A. |
| `NF-AC-073` | Mapping approval accepts only the three §10.4 variants; `derived`, `constant`, fake ignored-field sentinels, and variant-extra members fail deterministically. |
| `NF-AC-074` | Alias suggestions use `source_alias_match_key_v1`; duplicate alias matches produce blocking visible warnings until explicit user approval resolves every matched column. |
| `NF-AC-075` | Every Network Flow success response `data` object contains the schema, required members, ordering, nullable behavior, and `meta` shape defined in §14.6 or §17. |
| `NF-AC-076` | Every Table 20-A resource limit uses the configured default/minimum behavior, enforcement phase, error/truncation behavior, and invalid-config rejection specified in §20. |
| `NF-AC-077` | Safe samples and source-column samples follow Table 12-F exactly, including null samples for raw text/IP-like values and numeric-only samples for bounded integer-like values. |
| `NF-AC-078` | Route-local errors include Table 21-B details and choose the first error under Table 21-C and Table 21-D ordering. |
| `NF-AC-079` | Indicator linking accepts only v1 IP endpoint candidates, rejects every non-linkable field in Table 15-A1, and validates existing/create targets against Core canonical IP indicator identity. |
| `NF-AC-080` | Source filename display, default table display names, explicit display-name overrides, duplicate suffixing, hidden-file stems, trailing-dot stems, and soft-delete name reuse match §8.4. |

## 24. Core amendments and adoption blocker checklist

**NF-REQ-180**
Before this NLSpec can move from `draft` to `adopted/current`, the adoption checklist in Table 24-A MUST be closed.

**Table 24-A. Adoption blocker checklist**

| Blocker ID | Required closure |
| --- | --- |
| `NF-BLOCK-001` | Core 00 recognizes `network_flow_activity` as an adopted extension profile. |
| `NF-BLOCK-002` | Core 01 extension discovery lists the route family only when the profile is claimed. |
| `NF-BLOCK-003` | Core 01 import terminal result references admit `kind='network_flow_table'`. |
| `NF-BLOCK-004` | Core 03 admits extension-contributed top-level incident tabs without expanding base built-in tabs. |
| `NF-BLOCK-005` | Core 04 adds Network Flow route-family authorization/conformance hooks. |
| `NF-BLOCK-006` | Every fixture row in §22 has concrete path, byte hash, and expected output transcript. |
| `NF-BLOCK-007` | The generated contract artifacts derived from this NLSpec exist and pass drift checks. |
| `NF-BLOCK-008` | Route, parser, digest, graph, indicator-link, security, and limit acceptance criteria have executable tests. |
| `NF-BLOCK-009` | Core 02 or the adopted Core indicator registry designates the exact IP-literal indicator type token required by §15. |

## Appendix E. Future-only decision backlog and rationale

This appendix is non-normative. It records deferred work and rationale for readers of this draft. It does not add v1 implementation-conformance behavior.

**Table E-A. Future-only backlog**

| Backlog item | Future owner dependency | Current v1 handling |
| --- | --- | --- |
| Core 02 observation-source amendment | Core 02 would need to admit extension-sourced `indicator_observation` rows through a closed `origin_kind` extension token and a non-record extension source reference. | Network Flow v1 supports `observation_mode='binding_only'` only and always returns `created_observation_refs[]=[]`. |
| Binding list/read routes | A later Network Flow revision would need route inventory, authorization, pagination, filters, response schemas, and audit behavior for binding reads. | Bindings are observable only through link-route responses and `network_flow_indicator_binding_created` audit events. |
| Time-bucketed graph output | A later Network Flow revision would need UTC-epoch bucket alignment, bucket-count formula, bucketed response schema, limits, errors, fixtures, and a new graph-query digest version. | Bucket request members are unknown members and fail admission in v1. |

**Table E-B. Rationale notes**

| Decision | Rationale |
| --- | --- |
| Cursor binding excludes table versions | Flow rows are immutable after activation. Table rename changes metadata only, so binding cursor validity to table versions would invalidate live query cursors without changing row population. |
| Graph query digest excludes `limit_overrides` | Limit overrides lower failure thresholds but do not change query semantics. Excluding them keeps graph identity and fixture digests deployment-independent. |
| Validation preview is preview-slice-scoped | The mapping modal must remain responsive for large CSVs. Full-file validation belongs to import apply, which already owns the background validation pass. |

## Sources

This section records non-normative source evidence used to shape this draft. It does not add requirements. Sources include Core 00 through Core 04, `graph_projection_nlspec.md`, `nlspec-spec.md`, research reports `R01` through `R09` under `docs/research`, RFC 4180, RFC 7011, RFC 5952, RFC 8785, RFC 9844, IANA IPFIX registry background, Cisco SNA NetFlow/IPFIX guidance, and CSV-injection security background. Internet and external-source material remains supporting evidence only unless this NLSpec restates the behavior as a Network Flow requirement.
