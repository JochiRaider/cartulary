---
title: Network Flow Activity NLSpec
status: adopted/current
document_version: 2.0.0
contract_major: 2
profile_id: network_flow_activity
document_class: nlspec
---

## 1. Status, scope, authority, and extension profile

Status: `adopted/current`.

This NLSpec defines the implementation-conformance contract for the `network_flow_activity` extension profile. Its adoption dependencies and gates in Tables 1-B, 3-A, and 24-A are closed for version `2.0.0`, including the required Core, Extensions Subsystem, Graph Projection, Testing Harness, timezone-ruleset, fixture, import-preview, and presentation contracts.

Document version: `2.0.0`. Contract major: `2`. The state schema remains version `1`; the public-contract major change removes the competing profile-local discovery shape and does not reinterpret durable Network Flow state.

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
Research reports, UI guides, implementation guides, appendices, and external vendor documents MAY justify design choices or source-profile fixtures, but they MUST NOT become implementation-conformance authority unless this NLSpec or Core 00 through Core 04 restates the behavior as a requirement. Omission behavior: an implementation requires no supporting citation to conform.

Omission behavior: an implementation that ignores research reports, UI guides, implementation guides, appendices, or external vendor documents for conformance purposes remains conformant when it satisfies the normative requirements in this NLSpec and Core 00 through Core 04.

### 1.1 Version and compatibility

**NF-REQ-006a**
Network Flow MUST NOT define or emit profile-local contract-discovery metadata. Core 01 `GET /api/v1/extensions` is the sole discovery owner and emits this profile through the generic seven-member item with `profile_id='network_flow_activity'`, `claimable=true`, `contract_major=2`, reserved route family `/api/v1/incidents/{incident_id}/network-flow`, `workspace_keys=['network_analysis']`, and `capabilities=[]`; only `claimed` varies with the published resolved claim set. `document_version`, singular `route_root`, a profile-local item, a second decoder, and a compatibility alias are forbidden. A client that does not support major `2` omits the workspace without affecting Base behavior.

**Table 1-A. Contract version-change registry**

| Change class | Required version action |
| --- | --- |
| Editorial correction that changes no observable behavior | Increment patch version only. |
| New immutable source profile advertised through discovery | Increment minor version. |
| New additive capability | Requires a later adopted Extensions capability contract and the version action that contract assigns; contract major `2` advertises none. |
| New public route, request member, response member, error code, or closed token not explicitly reserved as additive | Increment contract major and affected schema IDs. |
| Changed default, limit, ordering, normalization, identity, digest, lifecycle, authorization, disclosure, or audit behavior | Increment contract major and every affected schema or algorithm ID. |
| Changed source aliases, requiredness, transform, empty-value policy, or timestamp interpretation | Introduce a new immutable `source_profile_id`; do not mutate the prior profile. |

Patch-version changes MUST NOT alter canonical bytes, identifiers, persisted resources, route status codes, error selection, audit output, fixture output, or caller-visible UI semantic state. Capability facts and nonempty capability arrays are invalid in contract major `2`; attempted activation fails with `extension_capability_not_supported`.

### 1.2 Normative dependency registry

**NF-REQ-006b**
Table 1-B MUST contain an adopted document version and exact imported section or schema for every row. A `TODO:` value is an adoption blocker for this NLSpec or any later revision. A dependency supplies only its named interface; it does not expand Network Flow scope implicitly.

**Table 1-B. Normative dependency registry**

| Dependency | Imported contract | Required adopted version and locator |
| --- | --- | --- |
| Core 00 | Extension ownership, recognition, primary-owner provenance, precedence, and adoption. | `cartulary.core00.current.v1`, version `extensions-adoption-1`, SHA-256 `3358cc76a3f5a7db45d5373929c43a35e6a1ef0149eea977ca5938edb068f90f`; `REQ-00-065`. |
| Core 01 | Generic seven-member discovery, import target/result union, bounded transaction boundary, staged objects, backup/restore, and public envelopes. | `cartulary.core01.current.v1`, version `extensions-adoption-1`, SHA-256 `9ad58e1339ab2bc2fa7f32e43cb2bf2f1739877406b80acbe6d8e80440c44627`; `REQ-01-151.1`, `REQ-01-542`, `REQ-01-629..633`. |
| Core 02 | Authoritative-state presence, canonical IP-literal indicator type, indicator transaction participation, and explicit no-private-purge boundary. | `cartulary.core02.current.v1`, version `extensions-adoption-1`, SHA-256 `79803b2d0b65cfec99e8fd053d2e3be0cab5540a442d5d19b1e47ca49e825454`; `REQ-02-074A..074C`, `REQ-02-210`, `REQ-02-261`. |
| Core 03 | Extension-contributed incident workspace, availability generation, stable Base identity, and resource invalidation. | `cartulary.core03.current.v1`, version `extensions-adoption-1`, SHA-256 `79afa553168703b6acfd3977562566962b8c1cf107caa55962636d4b0b3e00c5`; `REQ-03-011A`, `REQ-03-303`. |
| Core 04 | Closed inactive configuration, validation precedence, lease/publication lifecycle, authorization, cursor protection, audit, secrets, and retention. | `cartulary.core04.current.v1`, version `extensions-adoption-1`, SHA-256 `67583e759aea19b52ca1718b53537d6c1ca1328f4a83031267891c916542658b`; `REQ-04-123..146`. |
| Extensions Subsystem NLSpec | Owner manifests and fragments, generated registry, state coordination, bindings/codecs, participants, and conformance accounting. | Adopted/current `docs/extension-subsystem-nlspec.md` version `0.6.0`, SHA-256 `1d10578b2d10df2bfa17e3ab48d0cd3ad0cf36a1f0c285b8875f5df4109e2440`; `EXT-REQ-001..236`, `EXT-AC-001..158`. This dependency closed in the same atomic companion revision. |
| Graph Projection NLSpec | Ephemeral projection request, property and metadata mapping, result, and error interface. | Adopted/current Graph Projection NLSpec `docs/graph_projection_nlspec.md`; owner artifacts `4e446354`, `f177fb6b`, `81941bba`; locator: front matter `status: adopted/current`, §§4, 5.1.1, 10.0, 10.9, 12, 13, 14; `GP-AC-033`, `GP-AC-053`, `GP-AC-069`. |
| Testing Harness NLSpec | Contract artifact generation, fixture execution, and drift checks. | Adopted/current Testing Harness NLSpec `docs/testing-harness-nlspec.md`; locator: front matter `status: adopted/current`, §§8, 11, 12, 16, 17; `TH-HARNESS-REQ-657..663`, `TH-HARNESS-AC-049..055`, schemas `cartulary.network_flow_fixture_manifest.v1`, `cartulary.network_flow_activity_accounting.v2`, `cartulary.network_flow_timezone_ruleset_provenance.v1`. |

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

**NF-REQ-011a**
In a public request table, `Required=Yes` means the member MUST be present in the decoded request. A default is valid only for `Required=No`. In a canonical persisted-object table, a member may be required after default materialization only when the table labels itself as a canonical-object schema and separately defines request omission behavior.

**NF-REQ-011b**
A public Network Flow request or response object MUST be exact or a closed variant. Phrases such as `at least these members`, `optional mapped fields`, or `implementation-defined members` MUST NOT define a public object. Closed variants MUST use an explicit discriminator and MUST reject members belonging only to another variant.

**NF-REQ-011c**
JSON integers admitted by Network Flow request schemas MUST use a base-10 integer lexeme with no fraction, exponent, or leading plus sign and no leading zero except `0`. A negative lexeme is valid only when the owning scalar contract permits negative values. JSON numbers that are mathematically integral but use fraction or exponent syntax are not integers under this NLSpec.

## 3. Adoption gates and Core amendment prerequisites

**NF-REQ-012**
This NLSpec may be marked `status: adopted/current` only while every adoption gate in Table 3-A remains satisfied.

**Table 3-A. Adoption gates**

| Gate ID | Owner artifact | Required adoption change | Required evidence before adoption |
| --- | --- | --- | --- |
| `NF-GATE-001` | Core 00 | Add `network_flow_activity` to the extension-profile model and adopted-subsystem map. | Core 00 lists this NLSpec as adopted for the extension boundary only. |
| `NF-GATE-002` | Core 01 | Add `network_flow_activity` to the sole generic extension discovery producer with major `2`, its reserved route family, `workspace_keys=['network_analysis']`, and `capabilities=[]`. | Both claimed and unclaimed fixtures contain the same reserved route/workspace facts; only `claimed` changes, and no profile-local producer or decoder exists. |
| `NF-GATE-003` | Core 01 | Extend import apply to permit extension-owned analytical import targets that produce durable extension resources rather than Core `record_id` rows. | Import contract names `target_kind='network_flow_table'` as an extension result target. |
| `NF-GATE-004` | Core 01 | Permit terminal import results to reference `network_flow_table` resources when `target_kind='network_flow_table'`. | Import result schema accepts extension resource references without treating them as saved views or record-envelope rows. |
| `NF-GATE-005` | Core 03 | Admit extension-contributed top-level incident tabs without adding `Network Analysis` to the Base Profile built-in tab list. | Base built-in tabs remain Timeline, Hosts, Identities, Evidence, and Notes; `Network Analysis` appears only when the extension is claimed. |
| `NF-GATE-006` | Core 04 | Add route-family authorization and conformance references for `network_flow_activity`, preserving incident membership, incident roles, and the no-`deployment_admin` incident-data bypass invariant. | Authorization fixtures prove route-time reauthorization and deny `deployment_admin` without incident membership. |
| `NF-GATE-007` | Core 01 | Provide one common unit-of-work boundary spanning Core indicator find-or-create, extension binding commit, idempotency success, and audit outbox writes. | Injected-failure fixture proves all-or-nothing commit. |
| `NF-GATE-008` | Core 02 | Designate the exact canonical IP-literal indicator type, indicator find/create participant, and no-private-purge boundary. | Indicator and lifecycle fixtures use the adopted tokens, transaction participant, and no-current-purge behavior. |
| `NF-GATE-009` | Core 03 | Provide a generic current-authorization resource-invalidation event for extension workspaces. | Rename, delete, and authorization-change UI fixtures consume the event without polling-specific assumptions. |
| `NF-GATE-010` | Core 04 | Provide confidential and integrity-protected cursor behavior, transactional audit delivery, deployment-secret lifecycle, and retention hooks. | Cursor, audit, rotation, and retention fixtures pass. |
| `NF-GATE-011` | Graph Projection NLSpec | Adopt the ephemeral invocation, metadata mapping, result, and dependency-error boundary used by §14. | Exact adapter-input and dependency-failure fixtures pass. |
| `NF-GATE-012` | Testing Harness NLSpec | Admit Network Flow generated contracts, canonical fixtures, structural lint, and drift checks. | All §22 fixtures and §23 criteria execute under the adopted harness. |

**NF-REQ-013**
The adoption process MUST NOT satisfy any gate by silently treating a flow table as a saved view, a Core system view, a Core record-envelope row family, a generic projection table, a visual graph artifact, or a base-profile workbook surface.

**NF-REQ-014**
Every fixture row in §22 whose file bytes or expected output are not authored MUST retain an explicit `TODO:` value rather than pretending the conformance artifact exists. A `TODO:` fixture row is an adoption blocker for this NLSpec or any later revision.

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
| Binding list/read routes | Out of scope in v1. | Bindings are observable through the link-route response and created/reused audit events only. A later revision MAY add read routes. Omission behavior: v1 exposes no binding list, get, search, or indicator-centric binding route. |

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
| Graph contributor query | `viewer` | Recompute the graph selector; every table in scope must remain visible and active. |
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

`path_identity` and normalized request comparison MUST use Table 5-B.

**Table 5-B. Mutation idempotency comparison registry**

| Route ID | `path_identity` | Normalized comparison body | Replay point |
| --- | --- | --- | --- |
| `nf.tables.patch` | `network_flow_table_id:{network_flow_table_id}` | Object containing exactly `base_table_version` and normalized `display_name`; excludes `client_txn_id` and path members. | After admission, authorization, extension availability, and path syntax validation; before current table-version comparison. |
| `nf.tables.delete` | `network_flow_table_id:{network_flow_table_id}` | Object containing exactly `base_table_version`; excludes `client_txn_id` and path members. | After admission, authorization, extension availability, and path syntax validation; before current table-version comparison. |
| `nf.indicator_links.create` | `indicator-links` | Object containing exactly normalized `selector`, normalized `target`, `observation_mode`, and normalized `confirm_exact_value`; excludes `client_txn_id`. | After admission, current incident authorization, and extension availability; before selector resolution and Core mutation. Exact replay MUST revalidate current target-indicator visibility before returning the stored success. |
| Core import apply for `target_kind='network_flow_table'` | Core-owned import apply idempotency scope. | Core import apply normalized request. Network Flow MUST NOT add a second extension-local idempotency key for the same apply action. | Core import owner replays the apply result; Network Flow MUST return the same created table references for the replayed terminal result. |

**NF-REQ-025**
For each Network Flow-owned mutating route, the implementation MUST compute `network_flow_mutation_request_digest_v1` over the Table 5-B normalized comparison body after defaults are materialized and before mutation commit. The digest input MUST exclude `client_txn_id`, path parameters, actor identity, incident ID, route ID, and any server-computed response members.

```text
network_flow_mutation_request_digest_v1(input):
  UTF8("cartulary.network_flow.mutation_request_digest.v1") NUL
  UTF8(route_id) NUL
  UTF8(path_identity) NUL
  network_flow_canonical_json_v1(normalized_body) NUL
  return lowercase_hex(SHA256(bytes))
```

**NF-REQ-026**
If a mutation with the same idempotency tuple and same request digest has already committed, the route MUST return the original success response and MUST NOT perform the mutation again. Exact committed replay MUST occur at the replay point in Table 5-B even when the current resource version or lifecycle state has since changed, subject to the indicator-target visibility rule below. If the tuple matches but the request digest differs, the route MUST fail with Core error code `client_txn_conflict` and MUST NOT perform the mutation.

For `nf.indicator_links.create`, the replay rule above does not bypass current visibility. After finding an exact committed replay, the implementation MUST revalidate that the caller remains incident-authorized and that the stored target indicator remains visible. If either check fails, the route MUST fail through current Core authorization or hidden-resource behavior and MUST NOT return the stored target reference. This visibility failure does not delete or alter the committed idempotency record.

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
| `network_flow_diagnostic_id` | `nfd_` followed by 64 lowercase hexadecimal characters. | Deterministic `network_flow_diagnostic_id_v1` in §12.4. |
| `network_flow_graph_query_digest` | 64 lowercase hexadecimal characters. | Deterministic `network_flow_graph_query_digest_v1` in §6.8. |
| `network_flow_source_snapshot_id` | `nfsnap_` followed by 64 lowercase hexadecimal characters. | Deterministic `network_flow_source_snapshot_digest_v1` in §6.9. |

**NF-REQ-030**
`network_flow_table_id`, `network_flow_row_id`, `network_flow_endpoint_id`, `network_flow_flow_edge_id`, `network_flow_indicator_binding_id`, `network_flow_diagnostic_id`, and `network_flow_source_snapshot_id` MUST NOT be accepted as Core `record_id` values. Core record mutation, view-row, revision, rollback, merge, and saved-view routes MUST NOT treat these identifiers as Core record identifiers.

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

**NF-REQ-032a**
`network_flow_canonical_json_v1` MUST reject any decoded string containing an unpaired UTF-16 surrogate. It MUST emit `\"` for quotation mark, `\\` for reverse solidus, and the short escapes `\b`, `\t`, `\n`, `\f`, and `\r` for their corresponding control scalars. Every other scalar in U+0000 through U+001F MUST be emitted as `\u00xx` using lowercase hexadecimal digits. Solidus MUST NOT be escaped. U+2028, U+2029, and every other non-control Unicode scalar MUST be emitted as UTF-8.

**NF-REQ-032b**
Dynamic object member names MUST be ordered by Unicode scalar numeric value, comparing the first unequal scalar and then shorter length when one name is a prefix of another. Runtime locale, UTF-16 code-unit order, database collation, and insertion order MUST NOT affect dynamic member order. Closed objects MUST use their normative schema-table order.

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
| `aggregate_decimal_string_v1` | Canonical non-negative base-10 string with no sign, fraction, exponent, leading plus sign, or leading zero except `0`. Arithmetic is arbitrary precision. The maximum digit count is the applied `network_flow.max_aggregate_counter_digits`; `0` has one digit. |
| `bounded_text_256_v1` | JSON string with `0..256` Unicode scalar values, no NUL, and no C0 or C1 controls except tab only where a source profile explicitly permits it. |
| `bounded_text_1024_v1` | JSON string with `0..1024` Unicode scalar values, no NUL, and no C0 or C1 controls except tab and line break only where a source profile explicitly permits them. |
| `sha256_hex_v1` | Exactly 64 lowercase hexadecimal characters. |
| `safe_key_id_v1` | Non-secret ASCII identifier matching `^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`. It identifies one safe-digest key epoch and is not key material. |
| `client_txn_id` | Core route-scoped idempotency key. This NLSpec does not redefine the Core `client_txn_id` contract. |

**NF-REQ-034**
`timestamp_utc_v1` parsing MUST reject leap seconds, local timestamps without an applicable timestamp profile timezone, timezone names embedded in source values outside the selected timestamp-profile schema, DST folds, DST gaps, and values outside year `0001..9999`.

**NF-REQ-035**
`ip_literal_v1` comparison MUST compare IP family first with IPv4 before IPv6, then compare unsigned address bytes within the same family. Textual IP comparison MUST NOT use locale collation, database collation, or lexical string ordering.

**NF-REQ-036**
`bounded_text_256_v1` and `bounded_text_1024_v1` preserve decoded source field text for row provenance. Display names, mapping suggestions, and user-entered non-source labels MUST use Unicode NFC normalization and trim leading/trailing Unicode whitespace where their owning algorithm states so.

**NF-REQ-036a**
`trim_unicode_whitespace_v1` MUST remove only leading and trailing scalars in this closed Unicode 17.0 set: U+0009..U+000D, U+0020, U+0085, U+00A0, U+1680, U+2000..U+200A, U+2028, U+2029, U+202F, U+205F, and U+3000. It MUST NOT remove any other scalar. Unicode NFC processing MUST conform to Unicode Standard Annex #15. Conformance fixtures MUST use Unicode 17.0.0 data.

**NF-REQ-036b**
When `trim_unicode_whitespace_v1` is required, the owning algorithm MUST call it by name. The unqualified phrases `trim whitespace`, `Unicode trim`, and `runtime trim` MUST NOT define normative behavior.

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
  for each registered field_key in Table 9-D order:
    UTF8(field_key) NUL
    if value is null:
      UTF8("n") NUL
      UTF8("null") NUL
    else:
      UTF8("p") NUL
      network_flow_canonical_json_v1(value) NUL
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

For this requirement, the identity-bearing source-column descriptor members are exactly `source_column_ordinal`, `raw_header_text`, `normalized_header_for_suggestion`, and `raw_header_sha256`. Preview-only `sample_values[]`, `detected_empty_count`, preview accepted/rejected counts, preview messages, UI suggestion state, and UI ordering state MUST NOT participate in `mapping_fingerprint_v1`. Changing only a preview-only member MUST NOT change the fingerprint.

### 6.7 Safe digest algorithm

**NF-REQ-044**
`network_flow_safe_digest_v1` MUST compute a keyed safe digest for logs, telemetry, and administrative audit summaries.

```text
network_flow_safe_digest_v1(value_class, canonical_value):
  HMAC-SHA256(
    network_flow_safe_digest_key_material,
    UTF8("cartulary.network_flow.safe_digest.v1") NUL
    UTF8(value_class) NUL
    UTF8(canonical_value)
  )
  return lowercase_hex(hmac)
```

**NF-REQ-045**
`network_flow_safe_digest_key_material` MUST be deployment-local secret material with at least 256 bits of CSPRNG entropy. It MUST NOT be exported through public routes, logs, telemetry, fixtures, import diagnostics, or Graph Projection metadata.

**NF-REQ-045a**
Every production safe digest MUST be paired with a non-secret `safe_digest_key_id`. All nodes in one deployment MUST use the same active key and key ID. Network Flow route service MUST fail startup when either value is absent or invalid. Core 04 owns secret generation, distribution, storage, access control, and rotation.

**NF-REQ-045b**
Safe-digest rotation establishes a new correlation epoch. Previously persisted digests MUST NOT be rewritten. Equality comparison across different `safe_digest_key_id` values is undefined and MUST NOT be used for deduplication or authorization. Every table field, error detail, or audit field carrying a safe digest MUST carry the corresponding key ID in the same enclosing object.

**NF-REQ-045c**
For `network_flow_table_id` and `network_flow_indicator_binding_id`, generation MUST attempt an atomic unique insert. A collision MUST cause generation of fresh CSPRNG bytes and retry. After eight collisions in one allocation action, the implementation MUST return `network_flow_id_generation_failed` through the Core error envelope and MUST commit no resource, idempotency success, or domain audit event.

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
When multiple Network Flow workspace states apply, the status strip MUST render exactly the first applicable token in Table 7-B. Table 7-B is a total order; an implementation MUST NOT select a different token from the same conceptual family.

**Table 7-B. Network Flow status-strip total precedence**

| Precedence | Token | Exact trigger | Omission behavior |
| ---: | --- | --- | --- |
| 1 | `validation_failed` | Current mapping preview or import preparation failed and the failure blocks apply. | Omit when no current blocking validation failure exists. |
| 2 | `mapping_required` | Current mapping lacks a required mapping or has a blocking mapping conflict. | Omit when the current mapping is complete and valid. |
| 3 | `validating` | Preview validation for the current mapping fingerprint is pending. | Omit after the current preview becomes terminal. |
| 4 | `link_pending` | An indicator-link request for the current selection is pending. | Omit after that request becomes terminal or is superseded. |
| 5 | `graph_pending` | A graph request for the current semantic query is pending. | Omit after that request becomes terminal or is superseded. |
| 6 | `graph_stale` | Displayed graph output no longer matches current semantic query, authorization, or table lifecycle. | Omit when no graph output exists or after revalidation, close, or reset. |
| 7 | `link_committed` | A non-superseded link success was applied less than 5 seconds ago and no clearing semantic action occurred. | Omit after expiry or a clearing semantic action. |
| 8 | `graph_available` | A non-stale graph result exists for the current semantic query. | Omit after query change, authorization loss, close, reset, or staleness. |
| 9 | `loaded_with_rejections` | The active table exists and `row_count_rejected > 0`. | Omit when no active table exists or rejected count is zero. |
| 10 | `loaded` | The active table exists and `row_count_rejected = 0`. | Omit when no active table exists. |

**NF-REQ-054**
The `link_committed` interval starts when the successful response is applied to the current UI state. It ends after 5 seconds or the first clearing semantic action, whichever occurs first. Clearing semantic actions are changing active table, mapping, semantic query, graph selector, row or graph selection, or starting a mutation. Pointer movement, scrolling, focus movement, opening help, and copying displayed text MUST NOT clear it. `graph_available` MUST persist until query state changes, authorization is lost, graph panel closes, graph query resets, or graph becomes stale. `graph_stale` MUST persist until graph revalidation, graph panel close, or graph query reset.

**NF-REQ-054a**
A preview, row-query, graph-query, contributor-query, or link response MUST NOT replace visible result state when its bound mapping fingerprint, semantic query, selection, actor authorization, or request generation no longer equals the current state. The client MUST discard a superseded response without changing the current success or failure token.

**Table 7-C. Active inner-table selection**

| Event | Required selected table |
| --- | --- |
| Fresh workspace load | First visible active table by workspace order; empty state when none exists. |
| Successful import | First newly created table in import-result order. |
| Active table soft delete | Next visible table in the prior workspace order; otherwise previous; otherwise empty state. |
| Authorization loss | First remaining visible active table by workspace order; otherwise empty state. |
| Table rename | Preserve selection by `network_flow_table_id`. |
| Browser reload | First visible active table by workspace order. Browser-local selection restoration is unavailable in v1. |

### 7.3 Grid presentation, identity, and session layout

**NF-REQ-054b**
Network Analysis MUST render accepted rows, rejected-row diagnostics, and graph
contributors through the shared semantic grid adapter using respectively
`network_flow.accepted_rows.v1`, `network_flow.rejected_rows.v1`, and
`network_flow.graph_contributors.v1`. These identifiers are extension grid
schemas, not `view_schema_id` values. A flow table, accepted row, or diagnostic
MUST NOT be represented as a Core record or workbook projection.

The closed repo-local presentation registry at
`contracts/network-flow/presentation.v1.json` owns column field keys, labels,
value/renderer kinds, filter and sort eligibility, copy and indicator-link
eligibility, default visibility/order/width, minimum width, and inspector-only
posture. UI code MUST consume generated registry metadata and MUST NOT infer
behavior from visible labels or vendor grid coordinates.

**NF-REQ-054c**
Accepted-row default presentation order is `source_row_number` as a structural
gutter, then `network_flow.flow_start_utc`, `network_flow.flow_end_utc`,
`network_flow.src_ip`, `network_flow.src_port`, `network_flow.dst_ip`,
`network_flow.dst_port`, `network_flow.ip_protocol`,
`network_flow.bytes_count`, `network_flow.packets_count`,
`network_flow.input_interface`, and `network_flow.output_interface`.
`network_flow.exporter_id`, `network_flow.tcp_flags`, and
`network_flow.application_label` are hidden by default. Resource identifiers,
digests, mapping fingerprint, observation provenance, and `unmapped_raw` are
inspector-only and MUST NOT become linkable, filterable, sortable, or default
clipboard fields.

Rejected-row diagnostics default to source row, source column, field key, error
code, reason code, and a safe localized message derived from `message_key` and
`message_args`, falling back to the safe server `message`. Samples, hashes,
limits, and raw provenance are inspector-only. Contributor rows reuse accepted
row rendering, group only by `network_flow_table_id` with display name as its
label, and preserve §14.6 server order without client aggregation or sorting.

**NF-REQ-054d**
Network Flow grid rows are immutable and read-only. Authorized users MAY select
ranges and copy canonical semantic scalar values. Omission behavior: an
authorized user who does not make a selection performs no copy action and no
selection state is retained. The workspace MUST NOT expose
editor, paste, fill, draft/create, row-reorder, bulk-selection, local
authoritative filtering/sorting, or client aggregation behavior. Graph
visualization remains Network Flow-owned and MUST NOT be compiled through the
grid adapter; only contributor rows use grid grouping.

**NF-REQ-054e**
Column visibility, order, and width are session-memory state keyed by extension
profile, workspace key, and grid schema ID. State survives active-table changes
for the same grid schema and resets on browser reload or explicit reset. It MUST
NOT be stored in table metadata, Core saved views, browser local storage, or a
server resource.

**NF-REQ-054f**
Every Network Flow grid MUST use fixed-height rows and always-on row and column
virtualization. The supported client page envelope is 1,000 rows with every
declared column. Validation MUST prove a bounded rendered row and cell set,
semantic reachability of the first and last rows and off-screen columns, and
stable object identity for unchanged immutable resources. This envelope is
implementation/measurement evidence and is not a timing, benchmark, or Core 05
publication claim.

## 8. Table-tab lifecycle and table registry

### 8.1 Lifecycle states

**NF-REQ-055**
Each `network_flow_table` MUST use exactly one lifecycle state from Table 8-A.

**Table 8-A. Table lifecycle state machine**

| State | Meaning | Queryable | Graphable | Listed by default | Terminal | Allowed transitions |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `active` | Accepted rows are queryable and graphable. | Yes | Yes | Yes | No | `soft_deleted` |
| `soft_deleted` | Table is hidden from default workspace use but retained for audit, provenance, and binding traceability. | No | No | No | Yes | none |

**NF-REQ-056**
`renamed` MUST NOT be a lifecycle state. A table rename is a metadata mutation and audit event only.

**NF-REQ-057**
A table in `soft_deleted` state MUST NOT be returned by ordinary table query, row query, graph query, or default table list routes. This revision defines no deleted-table inspection route and no non-active table listing mode. Import parsing, mapping validation, and row validation MUST occur in import-job staging that is not addressable as a `network_flow_table`; an implementation MUST NOT expose `creating` or `failed` as table lifecycle states.

**NF-REQ-058**
A table in `soft_deleted` state MUST NOT appear in the default inner tab strip, MUST NOT be included by `table_scope.mode='all_active_tables'`, and MUST fail table, row, rejected-row, graph, contributor-query, and indicator-link route admission with `network_flow_table_not_active` when the caller references it directly. This failure is returned only after Core authorization and hidden-resource disclosure rules have admitted disclosure of the resource state; otherwise the route MUST return the Core hidden-resource result.

### 8.2 Table creation from import

**NF-REQ-059**
A successful import apply for one selected CSV import unit with `target_kind='network_flow_table'` MUST create exactly one `network_flow_table` when row validation admits at least one accepted row.

**NF-REQ-060**
A partially valid CSV import MUST create one table when at least one row is accepted. Rejected rows MUST be retained as diagnostics attached to that table subject to diagnostic truncation rules.

**NF-REQ-061**
An all-invalid CSV import MUST create no table and MUST fail that import unit with `network_flow_all_rows_rejected`. The error details MUST include `row_count_rejected`, `diagnostics_truncated`, and `diagnostics_sample[]` under Table 21-B. `diagnostics_sample[]` MUST contain the first `min(50, diagnostic_count)` diagnostics under §12.4 ordering using the Table 12-E shape.

**NF-REQ-062**
A CSV import with no data rows after its header MUST create no table and MUST fail that import unit with `network_flow_no_data_rows`.

**NF-REQ-062a**
The owner facade MUST perform table creation through `commit_network_flow_import_unit_v1` as one Core unit of work:

```text
commit_network_flow_import_unit_v1(validated_staging, approved_mapping, import_context):
  require validated_staging.row_count_accepted > 0
  acquire the Core incident-scoped Network Flow commit lock
  reauthorize import_context.actor for import apply
  revalidate source-content identity and approved mapping fingerprint
  reject when active_table_count + 1 > network_flow.max_active_tables_per_incident
  reject when retained_table_count + 1 > network_flow.max_retained_tables_per_incident
  determine the final display name under §8.4
  allocate a new network_flow_table_id
  insert exactly one active network_flow_table with table_version = 1
  insert all accepted immutable rows in source_row_number order
  insert retained diagnostics under §12.4
  emit the required audit occurrences under §16.2
  commit all preceding effects atomically
```

If any step fails, the unit of work MUST roll back the table, rows, diagnostics, audit occurrences, and any staged result publication. Retriable staging artifacts MAY remain under Core import-job retention but MUST NOT be queryable through Network Flow routes. Omission behavior: when Core does not retain failed job staging, it deletes those artifacts under its normal job cleanup. The import unit MUST expose the mapped route error and MUST NOT expose a table ID.

**NF-REQ-062b**
Cancellation observed before final transaction commit MUST roll back and publish a cancelled Core import-unit result with no Network Flow table reference. Once the final transaction commits, later cancellation delivery MUST NOT convert the unit to cancelled or delete the committed table; Core MUST publish or replay the committed success result. A worker crash after commit but before result delivery MUST recover the committed result through Core idempotency rather than rerun table creation.

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
| `source_filename_digest` | `sha256_hex_v1` | Yes | No | `network_flow_safe_digest_v1("source_filename", source_filename_display)`. |
| `source_filename_digest_key_id` | `safe_key_id_v1` | Yes | No | Active safe-digest key ID used to derive `source_filename_digest`. |
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
  candidate = trim_unicode_whitespace_v1(candidate)
  if candidate == "":
    candidate = "uploaded.csv"
  return first_256_unicode_scalars(candidate)
```

`remove_NUL_C0_C1_controls` removes exactly U+0000..U+001F and U+007F..U+009F. `source_filename_digest` MUST be `network_flow_safe_digest_v1("source_filename", source_filename_display)`. The implementation MUST NOT preserve directory segments, drive prefixes, UNC prefixes, object-store keys, temporary paths, or upload-worker paths in `source_filename_display`.

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
  if value contains U+0000..U+001F or U+007F..U+009F:
    fail network_flow_invalid_display_name with reason_code "forbidden_control"
  value = trim_unicode_whitespace_v1(value)
  return value
```

Examples: `C:\tmp\flows.csv` and `/tmp/flows.csv` both produce source filename display `flows.csv` and default display name `flows`; `.csv` produces default display name `.csv`; `file.` produces default display name `file`.

**NF-REQ-065b**
When import apply omits `display_name_override`, the final commit MUST use `derive_table_display_name_v1` and its deterministic suffix allocation. When import apply supplies `display_name_override`, the implementation MUST normalize and validate it using the rename rules in NF-REQ-066 and MUST NOT suffix, truncate, or otherwise repair a duplicate or overlong explicit value. A duplicate explicit value MUST fail the import unit with `network_flow_invalid_display_name` and `reason_code='duplicate_display_name'`.

**NF-REQ-066**
A table rename MUST be a metadata-only mutation on an `active` table. When the normalized requested name differs from the current name, it MUST increment `table_version`, update `display_name` and `updated_at`, and emit `network_flow_table_renamed`. When the normalized name equals the current name, it MUST return the unchanged table, MUST NOT increment the version or timestamp, and MUST NOT emit the rename event. Neither case may change table ID, row IDs, source provenance, mapping fingerprint, graph identity, live query cursor validity, diagnostics, or indicator bindings.

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
| `active` | Yes | Yes |
| `soft_deleted` | No | Yes |

The active count is the number of committed `active` tables in the incident. The retained count is the number of committed `active` plus `soft_deleted` tables in the incident. Staging objects and rolled-back commits count against neither limit.

### 8.6 Incident retention, soft delete, and future removal boundary

**NF-REQ-069a**
Network Flow resources MUST inherit the owning incident lifecycle, retention, and future incident-removal boundary from Core Documents 02 and 04. Soft delete is terminal in v1 and defines no restore operation. The current Network Flow Activity revision makes no whole-incident purge claim and MUST NOT define a private Network Flow purge cascade. Incident closure or authorization loss makes Network Flow data non-queryable through the ordinary route-admission and hidden-resource rules while retaining owner state for provenance, exact replay where route admission permits it, binding traceability, and Core-governed audit correlation. Whole-incident removal or retention-expiry deletion of Network Flow owner state requires a later adopted generic Core incident-removal profile with an explicit Network Flow cascade participant. Core audit records MUST remain or be removed only according to the Core audit-retention policy; this extension MUST NOT independently shorten that policy.

**Table 8-D. Lifecycle and retention consequences**

| Event | Table/row data | Diagnostics | Bindings | Idempotency state | Live cursors | Audit |
| --- | --- | --- | --- | --- | --- | --- |
| Table soft delete | Retained but non-queryable | Retained but non-queryable | Retained for traceability; non-actionable | Retained for exact replay under route-owned expiry | Invalidated | Retained |
| Incident closure or authorization loss | Retained but non-queryable while route admission fails | Retained but non-queryable while route admission fails | Retained; no link action admitted while route admission fails | Retained for exact replay only when route admission permits it | Invalidated | Governed by Core 04 |
| Future generic incident-removal profile | Not specified by v1; no purge evidence is claimable without a later Core owner | Not specified by v1 | Not specified by v1 | Not specified by v1 | Not specified by v1 | Governed by Core 04 and the later Core owner |
| Failed or cancelled import before final commit | No table or rows | No table diagnostics | None | No committed success | None | Import-job audit governed by Core; no table-created occurrence |

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
| Encoding | UTF-8 only. A UTF-8 BOM is accepted only at byte offset `0` and is stripped before header parsing. The UTF-8 BOM byte sequence at any later byte offset fails with `network_flow_invalid_utf8` and `reason_code='bom_not_at_offset_zero'`. Any other invalid UTF-8 fails with `reason_code='invalid_utf8_sequence'`. |
| Empty file | Fail import unit with `network_flow_csv_empty_file`; allocate no table. |
| Header | The first logical record is always the header and is source record number `1`. There is no header-discovery heuristic and no no-header mode. A decoded header containing NUL, C0 controls other than horizontal tab, or any C1 control fails the import unit with `network_flow_invalid_header`. |
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
| Preview size | Exactly the first `min(50, logical_data_record_count)` complete logical data records after the header, including blank and field-count-mismatched records. The preview parser MUST stop after the 50th complete data record and MUST NOT report errors that occur only in later records. |
| Limit timing | Column/header limits after header decode; raw cell scalar limit after field decode; row limit while streaming; accepted-row limit during validation. |

The `network_flow.max_rows_per_csv` count is the number of complete logical data records after the header, including blank records and field-count-mismatched records and excluding the header. The parser MUST fail before reading record `limit + 1` into validation. An unterminated quoted record encountered while seeking the next record remains `network_flow_csv_malformed_quote`, even when its ordinal would exceed the row limit.

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
| `network_flow.exporter_id` | `not_supported` | `bounded_text_256_v1` or null | Null for the claimable v1 profile. | A mapping to this target is invalid under Cisco SNA v1. |
| `network_flow.input_interface` | `optional_map_when_present` | `bounded_text_256_v1` or null | For Cisco SNA, apply `trim_ascii_space_v1`, then preserve decoded text. | Contains NUL/control or exceeds bound. |
| `network_flow.output_interface` | `optional_map_when_present` | `bounded_text_256_v1` or null | For Cisco SNA, apply `trim_ascii_space_v1`, then preserve decoded text. | Contains NUL/control or exceeds bound. |
| `network_flow.tcp_flags` | `not_supported` | Integer bitmask `0..255` or null | Null for the claimable v1 profile. | A mapping to this target is invalid under Cisco SNA v1. |
| `network_flow.application_label` | `not_supported` | `bounded_text_256_v1` or null | Null for the claimable v1 profile. | A mapping to this target is invalid under Cisco SNA v1. |
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
| `Start Time`, `Flow Start Time`, `First Seen`, `FIRST_SWITCHED`, `flowStartMilliseconds`, `flowStartSeconds` | `network_flow.flow_start_utc` | `required` | Mapping must declare timestamp profile. |
| `End Time`, `Flow End Time`, `Last Seen`, `LAST_SWITCHED`, `flowEndMilliseconds`, `flowEndSeconds` | `network_flow.flow_end_utc` | `required` | End must be greater than or equal to start. |
| `Interface input`, `Input Interface`, `ingressInterface` | `network_flow.input_interface` | `optional_map_when_present` | Mapped when present; absence is not a blocking warning. |
| `Interface output`, `Output Interface`, `egressInterface` | `network_flow.output_interface` | `optional_map_when_present` | Mapped when present; absence is not a blocking warning. |

**NF-REQ-080**
The Cisco SNA profile MUST treat the nine required Cisco fields as required profile fields and only the input-interface and output-interface fields as optional mapping targets. It MAY accept additional columns. Omission behavior: additional columns default to `unknown_column_policy='preserve_unmapped_raw'` unless the user explicitly ignores them. A mapping to a Table 9-D `not_supported` target MUST fail approval with `network_flow_mapping_conflict` and `reason_code='field_not_supported_by_profile'`.

**NF-REQ-080a**
Header alias matching MUST use `source_alias_match_key_v1` for both source headers and profile aliases.

```text
source_alias_match_key_v1(input):
  value = Unicode_NFC(input)
  value = trim_unicode_whitespace_v1(value)
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

**NF-REQ-080c**
The closed repo-local registry at
`contracts/network-flow/mapping-registry.v1.json` is the derived machine-readable
owner input for the claimable source-profile identifiers, parser and unknown
column defaults, timestamp modes, field requirements, aliases, transforms,
empty-value policies, and suggestion order in §§9.4 through 9.7. Backend and
frontend generated registries MUST derive from this artifact. The route-exposed
source-profile list remains authoritative for which compiled profiles are
currently claimable; a client MUST intersect that response with its generated
registry and fail closed on a missing or version-mismatched profile.

### 9.7 Timestamp profile

**NF-REQ-081**
Every approved Network Flow mapping MUST contain a `timestamp_profile_v1` object satisfying Table 9-G.

`timestamp_profile_v1` is a closed discriminated union selected by `mode`. A request variant MUST contain `schema_id='cartulary.network_flow.timestamp_profile.v1'` and the selected `mode`; `precision` is optional in the request and materializes the Table 9-H default. The approved canonical mapping MUST contain `schema_id`, `mode`, and materialized `precision`. A variant MUST contain exactly the additional members in Table 9-G; members belonging to another variant are invalid.

**Table 9-G. `timestamp_profile_v1` closed variants**

| `mode` | Additional required members | Member rules |
| --- | --- | --- |
| `rfc3339` | `timezone`, `timezone_ruleset_id`, `ambiguous_local_time_policy`, `local_time_gap_policy` | `timezone` is null, `UTC`, or an IANA zone name. `timezone_ruleset_id` is required and non-null only for a non-UTC IANA zone; otherwise it MUST be null. Both policies are exactly `reject`. |
| `epoch_seconds` | none | No timezone member is permitted; values are UTC by definition. |
| `epoch_milliseconds` | none | No timezone member is permitted; values are UTC by definition. |
| `netflow_sys_uptime_milliseconds` | `netflow_export_time_column_ordinal`, `netflow_export_time_mode`, `netflow_exporter_uptime_at_export_column_ordinal` | Ordinals are `1..network_flow.max_columns_per_csv`; export mode is `rfc3339`, `epoch_seconds`, or `epoch_milliseconds` and MUST produce a UTC instant without deployment-local inference. |

`schema_id`, `mode`, `precision`, `timezone`, `timezone_ruleset_id`, both policy members, and `netflow_export_time_mode` are JSON strings; `timezone` and `timezone_ruleset_id` are nullable only as Table 9-G permits. Both ordinal members are non-null JSON integers. Explicit null is invalid for every other member.

**Table 9-H. Timestamp mode precision**

| `mode` | Default `precision` | Permitted `precision` values |
| --- | --- | --- |
| `rfc3339` | `microseconds` | `seconds`, `milliseconds`, `microseconds` |
| `epoch_seconds` | `seconds` | `seconds` |
| `epoch_milliseconds` | `milliseconds` | `milliseconds` |
| `netflow_sys_uptime_milliseconds` | `milliseconds` | `milliseconds` |

**NF-REQ-082**
Under `mode='rfc3339'`, when `timezone` is null, every source value MUST carry `Z` or an explicit numeric offset. When `timezone` is a non-null IANA name, source values carrying an offset MUST use their own offset; source values without an offset MUST be interpreted in that zone with `reject` behavior for DST folds and gaps. The v1 Cisco SNA profile default timestamp behavior MUST NOT infer a deployment-local timezone.

A source value with finer resolution than `precision` MUST fail row validation with `network_flow_invalid_timestamp`. RFC3339 input MUST use the exact case-sensitive grammar `YYYY-MM-DD'T'HH:MM:SS[.fraction](Z|(+|-)HH:MM)` when an offset is present, or the same date-time without the final offset only when `timezone` is non-null. Calendar fields MUST denote a real proleptic-Gregorian date in year `0001..9999`; clock `HH` MUST be `00..23`, clock `MM` and `SS` MUST be `00..59`; the fraction MUST contain `1..6` digits; offset hour MUST be `00..23`; offset minute MUST be `00..59`; and `-00:00` MUST be rejected. Lowercase `t` or `z`, spaces, comma fractions, named-zone suffixes, more than six fractional digits, `24:00:00`, and leap-second `60` are invalid. Epoch-mode values MUST be unsigned canonical decimal integers in the mode's unit; fractional, signed, exponent, or leading-zero forms other than `0` are invalid. A `precision` outside the mode's permitted set MUST fail mapping approval with `network_flow_mapping_conflict`.

**NF-REQ-083**
Epoch seconds and epoch milliseconds are interpreted as UTC instants. They MUST NOT be interpreted using the caller's browser timezone, deployment timezone, database timezone, or locale settings.

**NF-REQ-084**
When `mode='netflow_sys_uptime_milliseconds'`, the implementation MUST derive an absolute timestamp using this formula:

```text
derived_event_time = export_time_utc - (exporter_uptime_at_export_ms - event_uptime_ms)
```

`export_time_utc` is parsed from `netflow_export_time_column_ordinal` using `netflow_export_time_mode`. When that mode is `rfc3339`, the export value MUST carry `Z` or an explicit numeric offset under the exact grammar above; offsetless export time is invalid. `exporter_uptime_at_export_ms` is parsed as an unsigned decimal integer from `netflow_exporter_uptime_at_export_column_ordinal`. `event_uptime_ms` is parsed as an unsigned decimal integer from the source column mapped to the timestamp target field being transformed. A row using this mode MUST fail with `network_flow_invalid_timestamp` when any participating value is missing, negative, non-integer, finer than the declared precision, outside its scalar bounds, or would produce a timestamp outside `timestamp_utc_v1`. A row MUST also fail with `network_flow_invalid_timestamp` when `event_uptime_ms > exporter_uptime_at_export_ms`, because that case is reboot- or wrap-ambiguous in v1.

All uptime values MUST be unsigned 32-bit integers in `0..4294967295`. The export-time ordinal and exporter-uptime-at-export ordinal MUST be distinct. For each timestamp field mapping, both ordinals MUST also be distinct from that field's event-uptime source ordinal. A violation MUST fail mapping approval with `network_flow_mapping_conflict` and `reason_code='timestamp_column_reused'`.

**NF-REQ-084a**
A non-UTC IANA-zone timestamp profile MUST declare `timezone_ruleset_id='tzdb-2026c'` for v1 conformance. This value is part of the mapping fingerprint. An implementation MAY internally use a later ruleset only when every timestamp fixture and every offset transition exercised by `tzdb-2026c` yields byte-identical normalized UTC output; otherwise it MUST reject the mapping as unsupported. Omission behavior: the implementation uses the `tzdb-2026c` ruleset. A future normative ruleset change requires a new document or profile version.

The immutable v1 provenance artifact for this ruleset is
`contracts/network-flow/timezone/tzdb-2026c.provenance.json` with
`schema_id='cartulary.network_flow_timezone_ruleset_provenance.v1'`. That
artifact MUST pin the IANA data-only archive `tzdata2026c.tar.gz`, its exact
release version, release timestamp, byte size, SHA-256 digest, detached
signature URL and digest, signing key identifier, license file digest, owner
references, and conformance policy. Timestamp fixtures or implementation
derivation that use IANA-zone local time MUST verify the archive bytes against
that artifact before deriving transition expectations. Mutable `latest` URLs,
the host operating system timezone database, browser `Intl` timezone data,
database server timezone tables, locale settings, or package-manager timezone
packages MUST NOT be treated as the normative source for v1 conformance.

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

**NF-REQ-088a**
`network_flow_import_facade_v1` MUST expose exactly the two operations in Table 10-A0. The facade boundary is logical and MAY be in-process or remote, but its semantic request and result objects are closed public objects under NF-REQ-011b. Omission behavior: an implementation that does not provide a remote boundary uses an in-process call with identical semantics.

**Table 10-A0. Import owner-facade operations**

| Operation | Request | Success result | Side effects |
| --- | --- | --- | --- |
| `preview` | `network_flow_import_preview_request_v1` | `network_flow_import_preview_result_v1` | None outside Core preview cache. MUST NOT allocate a table, persist rows or diagnostics, or emit domain audit occurrences. |
| `apply` | `network_flow_import_apply_request_v1` | `network_flow_import_unit_result_v1` from §18 | Atomic effects from NF-REQ-062a. |

**Table 10-A1. Common facade request members**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | Yes | Operation-specific schema ID: `cartulary.network_flow.import_preview_request.v1` or `cartulary.network_flow.import_apply_request.v1`. |
| `operation` | token | Yes | `preview` or `apply`, matching `schema_id`. |
| `import_session_id` | Core import-session ID | Yes | Authorized session reference. |
| `import_unit_id` | Core import-unit ID | Yes | Unit within the session. |
| `actor_context_ref` | opaque Core capability | Yes | Authorizes this operation; MUST NOT contain reusable bearer credentials in logs or results. |
| `source_stream_ref` | opaque Core byte-stream capability | Yes | Read-only stream for the selected unit. Filesystem paths, object-store paths, and client-supplied URLs are forbidden. |
| `source_content_sha256` | `sha256_hex_v1` | Yes | Server-derived from exact unit bytes and revalidated while consuming the stream. |

Each operation MUST contain exactly the common members plus its Table 10-A1a members.

**Table 10-A1a. Operation-specific facade request members**

| Operation | Member | Type | Required rule |
| --- | --- | --- | --- |
| `preview` | `mapping_candidate` | Table 10-B0 object | Client choices to validate and materialize into Table 10-B; server-derived descriptors and fingerprints are forbidden as client inputs. |
| `apply` | `approved_mapping` | Table 10-B object | Immutable server-materialized approved mapping. |
| `apply` | `mapping_fingerprint` | `sha256_hex_v1` | Must equal the approved and recomputed fingerprint. |
| `apply` | `mapping_approval_ref` | opaque Core capability | References the stored approval; forbidden in logs and results. |
| `apply` | `idempotency_context_ref` | opaque Core capability | References the admitted apply idempotency record and request digest; forbidden in logs and results. |

A preview request MUST NOT contain apply-only members, and an apply request MUST NOT contain `mapping_candidate`. Core owns authorization, job cancellation, request-size admission, mapping-approval storage, apply idempotency, and job result publication. The owner facade owns target parsing, target validation, staging, mapping fingerprint materialization, and final commit.

**Table 10-A2. `network_flow_import_preview_result_v1`**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | Yes | Exactly `cartulary.network_flow.import_preview_result.v1`. |
| `source_content_sha256` | `sha256_hex_v1` | Yes | Recomputed server value. |
| `source_columns[]` | array | Yes | Exact descriptors from Table 10-A. |
| `materialized_mapping` | object | Yes | Table 10-B object with defaults expanded. |
| `mapping_fingerprint` | `sha256_hex_v1` | Yes | Computed under §6.4. |
| `preview_record_count` | integer | Yes | `0..50`; exact number of complete logical data records examined. |
| `preview_accepted_count` | integer | Yes | Exact accepted count within preview scope. |
| `preview_rejected_count` | integer | Yes | Exact rejected count within preview scope; sum with accepted count equals `preview_record_count`. |
| `diagnostics[]` | array | Yes | Deterministic diagnostics under §12.4 for rejected preview records. |
| `diagnostics_truncated` | Boolean | Yes | Always `false` because preview scope is at most 50 records. |

**NF-REQ-088b**
Core Imports MUST expose the owner preview operation through the additive route
`POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview`.
The request is a closed object containing exactly `target_kind`,
`extension_profile_id`, `owner_mapping_schema_id`, and `owner_mapping`.
For this profile those values are respectively `network_flow_table`,
`network_flow_activity`, `cartulary.network_flow.mapping_candidate.v1`, and a
Table 10-B0 object.

The imports service MUST derive actor context, source capability, source hash,
discovered columns, header row, and data-start row from the authorized session
and unit. These values and any reusable capability are forbidden as client
members. The route requires the same editor-or-higher role set as mapping
approval and MUST return a closed Core wrapper containing exactly `schema_id`,
`import_session_id`, `import_unit_id`, `target_kind`, `extension_profile_id`,
`owner_result_schema_id`, and `owner_result`. `schema_id` is
`cartulary.imports.extension_mapping_preview_result.v1`;
`owner_result_schema_id` is
`cartulary.network_flow.import_preview_result.v1`; and `owner_result` is the
validated Table 10-A2 object.

The preview route is repeatable and side-effect-free except for an optional
bounded Core preview cache. It MUST NOT require `client_txn_id`, persist mapping
approval or selection, allocate a table, persist rows or diagnostics, start
apply, or emit a domain audit occurrence. Preview failures use the existing
Core import error family with safe `field` and `reason_code` details.

**NF-REQ-088c**
`PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping`
remains the only durable approval route and retains its Core response shape.
Before selecting or applying a Network Flow unit, the client MUST compare the
durable unit's returned `mapping_fingerprint` byte-for-byte with the latest
non-stale preview result. A mismatch MUST return the UI to
`validation_preview_pending` and MUST NOT select or apply the unit. The Core
transport `source_columns[]` supplied for approval MUST be constructed from the
discovered unit response and MUST NOT be hard-coded or inferred from a source
profile.

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

`source_columns[]` MUST contain exactly one descriptor for every header field, its array length MUST equal the decoded header field count, and ordinals MUST be the contiguous sequence `1..length` in array order. The server MUST derive every descriptor from the source stream and MUST ignore or reject client-supplied descriptor values. `sample_values[]` MUST contain one sample for each structurally valid preview record for that column, in `source_row_number ASC`, up to 50; field-count-mismatched records contribute no column sample. `detected_empty_count` is the number of those sampled decoded values equal to `""` before transforms. `raw_header_sha256` is computed from the exact UTF-8 encoding of `raw_header_text` after CSV decode and before normalization.

### 10.3 Approved mapping metadata

**NF-REQ-090**
An approved mapping for this extension MUST include the top-level members in Table 10-B.

**Table 10-B0. `mapping_candidate_v1` request object**

| Member | Type | Required | Nullable | Default or omission behavior |
| --- | --- | ---: | ---: | --- |
| `target_kind` | string | Yes | No | Exactly `network_flow_table`. |
| `target_table_schema_id` | string | Yes | No | Exactly `cartulary.network_flow_table.v1`. |
| `source_profile_id` | string | Yes | No | One claimable Table 9-A profile; UI preselection does not make this member optional. |
| `parser_profile_id` | string | No | No | `rfc4180_headered_csv_v1`. |
| `unknown_column_policy` | string | No | No | `preserve_unmapped_raw`. |
| `display_name_override` | string | No | No | True omission invokes §8.4 derivation. Explicit null invalid. |
| `timestamp_profile` | object | Yes | No | Closed request variant from §9.7. |
| `field_mappings[]` | array | Yes | No | Candidate Table 10-D variants; length `0..(effective network_flow.max_columns_per_csv + 1)`. |

Unknown members, `source_columns`, mapping fingerprints, preview samples, source hashes, and server capability references are forbidden in `mapping_candidate_v1`.

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
| `source_columns[]` | array | Yes | No | none | Exhaustive server-derived descriptors for every source column in source order. Clients MUST NOT supply or alter this member. |
| `field_mappings[]` | array | Yes | No | none | Closed mapping-row variants from §10.4. |

Table 10-B is the canonical approved-object schema after request defaults and server-derived members are materialized; its `Required=Yes` values are not a claim that those members were required in Table 10-B0. The server MUST materialize omitted defaults and `source_columns[]`, then compute the separate `mapping_fingerprint` before approval. The approved mapping is immutable and is the only mapping object passed to apply. Apply MUST fail with `network_flow_source_changed` when source bytes, header descriptors, or the recomputed fingerprint differ from the approved values.

**Table 10-C. Unknown-column policy**

| Token | Required behavior |
| --- | --- |
| `preserve_unmapped_raw` | Default. Retain every source column without a source mapping or explicit ignored mapping as inert row provenance. They are not filterable, sortable, graphable, or linkable until mapped to a registered field in a later import. |
| `reject_unmapped_columns` | Reject mapping approval when any source column is unmapped and not explicitly ignored. |
| `ignore_unmapped_columns` | Do not retain unmapped values. This token requires explicit user approval in the mapping modal and one `ignored_source_column` mapping for every source column not used by a target mapping. |

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

**NF-REQ-093a**
An approved mapping MUST satisfy all of the following cardinality invariants: every `required` target has exactly one `source_column` mapping; every `system_derived` target has exactly one `system_derivation` mapping; every `optional_map_when_present` target has zero or one `source_column` mapping; every `not_supported` target has zero mappings; each source ordinal has exactly one disposition as a source mapping, ignored mapping, or preservation under `unknown_column_policy`; no source ordinal appears in two mapping rows; and no target field appears twice. Violations MUST fail with `network_flow_mapping_conflict` and the most specific Table 21-B reason code.

**NF-REQ-093b**
The implementation MUST materialize `network_flow.observation_source_ref` exactly once per accepted row after source-row identity and the approved mapping fingerprint are final. Clients, CSV columns, and transforms MUST NOT supply or override any member of that object.

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

`trim_ascii_space_v1` removes zero or more U+0020 SPACE scalars from each end and no other scalar. `identity_text_v1` preserves the decoded scalar sequence exactly. Neither transform performs Unicode normalization. Text output containing NUL, C0 controls other than horizontal tab, or C1 controls is invalid; the target scalar bound is measured after the transform.

**Table 10-E1. Cisco SNA v1 transform matrix**

| Target field family | Required transform |
| --- | --- |
| `flow_start_utc`, `flow_end_utc` | `timestamp_profile_v1` |
| `src_ip`, `dst_ip` | `ip_literal_v1` |
| `src_port`, `dst_port` | `port_number_v1` |
| `ip_protocol` | `protocol_number_or_token_v1` |
| `bytes_count`, `packets_count` | `uint64_decimal_string_v1` |
| `input_interface`, `output_interface` | `trim_ascii_space_v1` |

Any other field/transform pair MUST fail mapping approval with `network_flow_mapping_conflict` and `reason_code='transform_target_mismatch'`.

**NF-REQ-095**
For each mapped value, validation MUST apply this exact pipeline: decode CSV; apply the transform; compare the transformed value with the empty string; apply the materialized empty-value policy; then validate the target scalar contract. Transform failure on a mapped required or optional field MUST reject the row. A null produced by `empty_string_is_null` is valid only for a nullable target; otherwise it rejects the row as the field-specific missing-value diagnostic. This post-transform empty test means a Cisco interface value containing only U+0020 becomes null under `empty_string_is_null`, invalid under `empty_string_is_invalid`, or `""` under `empty_string_preserved`.

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
For Cisco SNA required fields, default empty-value policy MUST materialize as `empty_string_is_invalid` before mapping approval. For Cisco SNA input and output interface mappings, it MUST materialize as `empty_string_is_null`. A persisted mapping containing `profile_default` is invalid. `empty_string_preserved` is unavailable for Cisco SNA interface mappings because the public nullable field contract uses null for a transformed empty value.

## 11. Mapping modal behavior

**NF-REQ-098**
The mapping modal MUST be explicit. The implementation MAY provide suggestions. Omission behavior: suggestions do not approve a mapping, do not create a table, and do not start import apply without explicit user approval.

**NF-REQ-099**
The mapping modal MUST display, at minimum:

- source filename display value;
- source profile selector showing only claimable profiles;
- parser profile identifier;
- source columns by `source_column_ordinal`, raw header, and safe sample values;
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
| `network_flow.exporter_id` | `bounded_text_256_v1` or null | Yes | Null when unsupported, unmapped, or empty under a nullable mapping policy. |
| `network_flow.input_interface` | `bounded_text_256_v1` or null | Yes | Null when unmapped or empty under a nullable mapping policy. |
| `network_flow.output_interface` | `bounded_text_256_v1` or null | Yes | Null when unmapped or empty under a nullable mapping policy. |
| `network_flow.tcp_flags` | integer `0..255` or null | Yes | Null when unsupported or unmapped. |
| `network_flow.application_label` | `bounded_text_256_v1` or null | Yes | Null when unsupported or unmapped. |
| `unmapped_raw` | object | Yes | Contains retained unmapped source values only when unknown-column policy preserves them. Empty object when none. |
| `network_flow.observation_source_ref` | object | Yes | Source provenance object from §12.3. |
| `created_at` | `timestamp_utc_v1` | Yes | Import commit time. |
| `created_by_user_id` | User ID | Yes | Importing actor. |

**NF-REQ-105**
`unmapped_raw` values MUST be inert provenance. They MUST NOT be filterable, sortable, graphable, exported, linked to indicators, or used for canonical row identity except through `source_row_digest_sha256` unless a later revision defines a promotion operation.

**NF-REQ-106**
`unmapped_raw` MUST be a canonical JSON object keyed by source column ordinal as a canonical decimal string without leading zeroes. Keys MUST appear in numeric ordinal order in canonical serialization. Each value MUST be an exact `unmapped_raw_value_v1` object from Table 12-A1. The object MUST be `{}` when the policy is not `preserve_unmapped_raw` or no unmapped values exist.

**Table 12-A1. `unmapped_raw_value_v1`**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `source_column_ordinal` | positive integer | Yes | Equals the enclosing object key interpreted as decimal. |
| `raw_header_text` | string | Yes | Exact decoded header text. |
| `raw_header_sha256` | `sha256_hex_v1` | Yes | Digest of exact UTF-8 header text. |
| `decoded_value` | string | Yes | Exact decoded cell value before transforms. |
| `decoded_value_sha256` | `sha256_hex_v1` | Yes | Digest of exact UTF-8 `decoded_value`. |

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
The `network_flow.observation_source_ref` object MUST include exactly the members in Table 12-C.

**Table 12-C. `network_flow.observation_source_ref`**

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

Every member is server-derived under NF-REQ-093b. The object MUST NOT contain filenames, storage paths, upload URLs, raw source values, or actor credentials.

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
| `diagnostic_id` | `network_flow_diagnostic_id` | Yes | Generated by `network_flow_diagnostic_id_v1`. |
| `source_row_number` | positive integer | Yes | CSV record number. |
| `source_column_ordinal` | integer or null | Yes | Null only when error is row-level. |
| `raw_header_sha256` | `sha256_hex_v1` or null | Yes | Digest for an attributable column; null for a row-level diagnostic. |
| `field_key` | string or null | Yes | Null only when no target field is attributable. |
| `error_code` | error token | Yes | From §21. |
| `reason_code` | reason token | Yes | Exhaustive route-local reason from Table 21-B. |
| `safe_sample` | string or null | Yes | Bounded safe sample after redaction under Table 12-F. |
| `raw_value_sha256` | `sha256_hex_v1` or null | Yes | Required whenever a source scalar exists; null only for row-level diagnostics with no attributable source scalar. |
| `message_key` | string | Yes | Exactly `network_flow.diagnostic.{error_code}.{reason_code}` using the emitted tokens. |
| `message_args` | object | Yes | Exactly `{}` in v1; future arguments require a versioned diagnostic schema. |
| `message` | `bounded_text_1024_v1` | Yes | Exactly equal to `message_key` in v1. UI localization is presentation-only and does not alter the wire resource. |
| `limit_name` | string or null | Yes | Configuration key only for a limit diagnostic; otherwise null. |
| `limit_value` | non-negative integer or null | Yes | Effective limit only for a limit diagnostic; otherwise null. |
| `actual_value` | non-negative integer or null | Yes | Observed value only for a limit diagnostic; otherwise null. |

**NF-REQ-113**
Diagnostics MUST be ordered by `source_row_number ASC`, then `source_column_ordinal ASC` with nulls last, then `field_key ASC` with nulls last, then `error_code ASC`.

Within one row, diagnostics MUST be produced deterministically as follows: a field-count mismatch emits exactly one row-level `network_flow_csv_field_count_mismatch` diagnostic and skips all target-field validation; otherwise validate mapped targets in Table 9-D order and emit at most one diagnostic per target using the failure precedence `missing_or_empty`, `transform_syntax`, `scalar_range_or_bound`, then `cross_field_semantics`; after target validation, evaluate `flow_end_utc < flow_start_utc` and, when true, emit exactly one diagnostic attributed to `network_flow.flow_end_utc`. The final retained order is NF-REQ-113 and is independent of worker scheduling.

**NF-REQ-113a**
`network_flow_diagnostic_id_v1` MUST hash this byte stream so the same preview and apply diagnostic has the same identity without depending on a table allocation:

```text
UTF8("cartulary.network_flow.diagnostic_id.v1") NUL
UTF8(source_content_sha256) NUL
UTF8(mapping_fingerprint) NUL
UTF8(decimal(source_row_number)) NUL
UTF8(source_column_ordinal is null ? "n" : "p:" + decimal(source_column_ordinal)) NUL
UTF8(field_key is null ? "n" : "p:" + field_key) NUL
UTF8(error_code) NUL
UTF8(reason_code) NUL
return "nfd_" + lowercase_hex(SHA256(bytes))
```

Message localization, safe samples, raw-value digests, worker identity, and table ID MUST NOT affect diagnostic identity.

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
| `schema_id` | string | Yes | No | none | Route-specific initial-query schema ID. |
| `filters[]` | array of filter objects | No | No | `[]` | Max `network_flow.max_filters_per_query`. |
| `sort[]` | array of sort objects | No | No | `[]`, then default sort tail | Max `network_flow.max_sorts_per_query`. |
| `limit` | integer | No | No | `min(200, effective network_flow.max_query_limit)` | `1..effective network_flow.max_query_limit`. |

A supplied `limit` MUST be an integer in `1..effective network_flow.max_query_limit` and MUST fail with `network_flow_invalid_limit` when outside that range. Clamping applies only to the omitted-member default and MUST NOT apply to explicit caller input.

**NF-REQ-115a**
Every paginated query body MUST be exactly one closed variant. An initial request contains its route-specific `schema_id`, route scope where required, and the semantic members in Table 13-A; it MUST NOT contain `cursor_token`. A continuation request contains exactly `schema_id` and `cursor_token`, where `schema_id` is the route-specific continuation schema ID. It MUST NOT repeat or override scope, filters, sort, limit, time range, aggregation, selector, or any other semantic input. A mixed request MUST fail with `network_flow_cursor_invalid`.

**Table 13-A1. Paginated route schema IDs**

| Route | Initial `schema_id` | Continuation `schema_id` |
| --- | --- | --- |
| `nf.tables.query` | `cartulary.network_flow.table_query_request.v1` | `cartulary.network_flow.table_query_continuation.v1` |
| `nf.rows.query` | `cartulary.network_flow.rows_query_request.v1` | `cartulary.network_flow.rows_query_continuation.v1` |
| `nf.rejected_rows.query` | `cartulary.network_flow.rejected_rows_query_request.v1` | `cartulary.network_flow.rejected_rows_query_continuation.v1` |
| `nf.graphs.contributors.query` | `cartulary.network_flow.graph_contributor_query_request.v1` | `cartulary.network_flow.graph_contributor_query_continuation.v1` |

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

`table_scope_v1` is a closed discriminated union. It MUST contain exactly the members shown for one Table 13-B variant; members from other variants are invalid.

**Table 13-B. `table_scope_v1` closed variants**

| `mode` | Exact additional members | Rule |
| --- | --- | --- |
| `active_table` | `active_table_id` | One active table visible to the actor. |
| `selected_tables` | `selected_table_ids[]` | Array length `1..network_flow.max_selected_tables_per_query`. |
| `all_active_tables` | none | Resolve all visible active tables in workspace order. |

**NF-REQ-118**
`selected_table_ids[]` with duplicate table IDs MUST fail with `network_flow_invalid_table_scope`. The implementation MUST NOT silently deduplicate the list.

**NF-REQ-119**
`all_active_tables` MUST select all visible active tables in the incident ordered by `created_at ASC`, then `network_flow_table_id ASC`. It MUST NOT include soft-deleted tables. Every ID explicitly supplied through `active_table` or `selected_tables` MUST independently pass current authorization and active-state admission; hidden IDs fail through Core hidden-resource behavior and disclosed soft-deleted IDs fail with `network_flow_table_not_active`. Explicit IDs MUST NOT be silently removed. Resolved IDs use workspace order regardless of request array order. Every resolved cross-table scope, including `all_active_tables`, MUST contain no more than `network_flow.max_selected_tables_per_query`; an excess fails with `reason_code='selected_table_limit_exceeded'`. If a variant resolves to zero active tables, the request MUST fail with `network_flow_invalid_table_scope` and `reason_code='empty_resolved_scope'`; it MUST NOT return an empty success that reveals hidden-table counts.

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

**NF-REQ-121a**
`normalized_filters_v1` MUST canonicalize every scalar with its field contract, canonicalize CIDRs to network address plus prefix length, sort each `in` array by the field's Table 13-F ascending comparator, reject duplicates after canonicalization, and sort the resulting filter objects by canonical JSON bytes. Two filters with identical canonical JSON after normalization MUST fail with `network_flow_invalid_filter` and `reason_code='duplicate_filter'`. A syntactically different value that canonicalizes to an existing `in` member, including equivalent IP or timestamp text, is a duplicate and MUST NOT be silently removed. The normalized array, not caller order or spelling, MUST bind cursors and query digests.

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
| Token encoding or length invalid | Fail with `network_flow_cursor_invalid`; token MUST be ASCII and at most 4096 bytes. |

**NF-REQ-126**
A cursor token expires exactly 15 minutes after `issued_at`. It is valid only while `issued_at <= now < expires_at`; equality with `expires_at` is expired. Core Document 04 MUST provide integrity protection, confidentiality, key rotation, and constant-disclosure failure behavior for cursor tokens. Plaintext client-readable cursor payloads are non-conformant.

**NF-REQ-126a**
A cursor token MUST bind actor, route, incident, table scope, filters, sort, limit, issued-at time, expiry time, table IDs, and continuation position. It MUST NOT bind table versions.

**NF-REQ-126b**
The cursor continuation position MUST be the last emitted item's full `effective_sort` tuple plus `network_flow_table_id` and `network_flow_row_id`. Continuation MUST return only rows that compare after that tuple under the same `effective_sort` comparator. The implementation MUST NOT use page offsets, visual row positions, browser grid indices, or table display names as continuation identity.

Rejected-row diagnostic pagination MUST use its own full comparator tuple: `(source_row_number ASC, source_column_ordinal ASC NULLS LAST, field_key ASC NULLS LAST, error_code ASC, reason_code ASC, diagnostic_id ASC)`. A diagnostic cursor MUST bind that comparator and MUST NOT reuse the accepted-row tuple. Graph-contributor pagination MUST use the contributor comparator defined in §14.6.

**NF-REQ-126c**
A query response MUST set `meta.paging.next_cursor_token` to an opaque string only when at least one additional row, diagnostic, or graph contributor exists after the last returned item under the route's bound comparator and current authorization. Otherwise it MUST set `next_cursor_token` to JSON `null`. A zero-result response MUST set `next_cursor_token` to JSON `null`.

## 14. Cross-table graph-composition contract

### 14.1 Graph query request

**NF-REQ-127**
The graph query route MUST accept a request body with the members in Table 14-A.

**Table 14-A. `network_flow_graph_query_request_v1`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | none | Exactly `cartulary.network_flow.graph_query_request.v1`. |
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
| `bytes_sum` | `aggregate_decimal_string_v1` | Arbitrary-precision sum of contributing `bytes_count`; fail before output if canonical output exceeds `network_flow.max_aggregate_counter_digits`. |
| `packets_sum` | `aggregate_decimal_string_v1` | Arbitrary-precision sum of contributing `packets_count`; fail before output if canonical output exceeds `network_flow.max_aggregate_counter_digits`. |
| `first_flow_start_utc` | `timestamp_utc_v1` | Earliest contributing start. |
| `last_flow_end_utc` | `timestamp_utc_v1` | Latest contributing end. |
| `contributing_table_ids[]` | array | Unique active table IDs ordered by workspace table order. |
| `example_row_refs[]` | array | First refs under §14.5 ordering. |
| `example_refs_truncated` | boolean | True when omitted contributors exist. |
| `example_refs_total_count` | integer | Total contributing row ref count before truncation. |

### 14.4 Graph Projection adapter

**NF-REQ-136**
The graph query route MUST construct `network_flow_graph_projection_adapter_v1` and invoke Graph Projection `project_ephemeral` with exactly one `projection_input` request member. The implementation MUST NOT submit v1 Network Flow graph queries through public retained Graph Projection lifecycle operations that allocate addressable graph views, retained runs, or retained graph output.

The exact Graph Projection input submitted as `projection_input` MUST satisfy Table 14-G after Graph Projection-owned default materialization. Network Flow owns the values in Table 14-G; Graph Projection owns validation, projected graph IDs, sorting, and graph-object derivation after the input is admitted.

**Table 14-G. `network_flow_graph_projection_input_v1`**

| Top-level member | Required value or derivation |
| --- | --- |
| `projection_schema_id` | Exactly `graph_projection.v1`. |
| `graph_view_id` | Graph Projection §7.4 ID derived from `projection_config.graph_view_key`; Network Flow MUST NOT invent another ID algorithm. |
| `source_snapshot_id` | `network_flow_source_snapshot_digest_v1` from §6.9. |
| `projection_config` | Exact object from Table 14-G1. |
| `source_entities[]` | One Table 14-G2 object per endpoint vertex after filters and time selection, ordered by `source_entity_id ASC` before submission. |
| `source_relationships[]` | One Table 14-G3 object per aggregated default flow edge, ordered by `source_relationship_id ASC` before submission. |
| `source_metadata` | Object containing exactly `incident_id`, `graph_query_digest`, `source_snapshot_id`, `selected_table_ids[]`, and `mapping_fingerprints[]` under NF-REQ-141a. |
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
| `entity_mappings[]` | One rule with `mapping_rule_id='nf.map.ip_endpoint.v1'`, `source_entity_kind='network_flow.ip_endpoint.v1'`, `projected_vertex_kind='network_flow.ip_endpoint.v1'`, `inclusion_predicate='always'`, `label_policy='mapping_only'`, `mapping_labels=[]`, `required_property_keys=["contributing_table_ids","endpoint_kind","endpoint_value","flow_row_count","indicator_candidate_value"]`, and `optional_property_keys=[]`. |
| `relationship_mappings[]` | One rule with `mapping_rule_id='nf.map.flow_edge.v1'`, `source_relationship_kind='network_flow.flow_edge.v1'`, `projected_edge_kind='network_flow.flow_edge.v1'`, `inclusion_predicate='always'`, `direction_policy='preserve'`, `emit_reverse_edge=false`, `reverse_edge_kind='network_flow.flow_edge.v1'`, `label_policy='mapping_only'`, `mapping_labels=[]`, `required_property_keys=["bytes_sum","contributing_table_ids","dst_endpoint_id","dst_port","edge_id","example_refs_total_count","example_refs_truncated","first_flow_start_utc","flow_row_count","ip_protocol","last_flow_end_utc","packets_sum","src_endpoint_id"]`, and `optional_property_keys=[]`. |
| `metadata_mappings[]` | Exact array from Table 14-G1a in `metadata_mapping_id ASC` order. |
| `aggregation_rules[]` | `[]`; Network Flow performs aggregation before adapter submission. |
| `default_vertex_labels[]` | `[]`. |
| `default_edge_labels[]` | `[]`. |
| `allow_empty_kind_registry` | `false`. |
| `retention_policy` | `{ "retain_replaced_results": false, "retention_count": 0, "retention_duration_seconds": 0, "retain_failed_results": false, "failed_retention_count": 0, "failed_retention_duration_seconds": 0 }`. |
| `custom_config` | `{}`. |

**Table 14-G1a. Fixed Graph Projection metadata mappings**

| `metadata_mapping_id` | `target_scope` | `target_kind` | `source_field_path` | `projected_metadata_key` | `projected_type` |
| --- | --- | --- | --- | --- | --- |
| `nf.mm.edge.contributing_table_ids.v1` | `edge` | `network_flow.flow_edge.v1` | `metadata.contributing_table_ids` | `contributing_table_ids` | `identifier_array` |
| `nf.mm.edge.example_refs_total_count.v1` | `edge` | `network_flow.flow_edge.v1` | `metadata.example_refs_total_count` | `example_refs_total_count` | `integer` |
| `nf.mm.edge.mapping_fingerprints.v1` | `edge` | `network_flow.flow_edge.v1` | `metadata.mapping_fingerprints` | `mapping_fingerprints` | `identifier_array` |
| `nf.mm.vertex.contributing_table_ids.v1` | `vertex` | `network_flow.ip_endpoint.v1` | `metadata.contributing_table_ids` | `contributing_table_ids` | `identifier_array` |
| `nf.mm.vertex.flow_row_count.v1` | `vertex` | `network_flow.ip_endpoint.v1` | `metadata.flow_row_count` | `flow_row_count` | `integer` |
| `nf.mm.vertex.mapping_fingerprints.v1` | `vertex` | `network_flow.ip_endpoint.v1` | `metadata.mapping_fingerprints` | `mapping_fingerprints` | `identifier_array` |

Every Table 14-G1a item MUST use Graph Projection's `metadata_mapping_rule`
schema with `required=true`, no `default_value`, `missing_behavior='error'`,
`source_null_behavior='error'`, `null_output_policy='omit'`, and
`merge_behavior='single_value'`. The adapter MUST NOT submit Graph
Projection-unknown members such as `mode`, `transform`, `direct_copy`,
`source_key`, or `target_key`.

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
| `vertex` | `network_flow.ip_endpoint.v1` | `contributing_table_ids`, `endpoint_kind`, `endpoint_value`, `flow_row_count`, `indicator_candidate_value`. |
| `edge` | `network_flow.flow_edge.v1` | `bytes_sum`, `contributing_table_ids`, `dst_endpoint_id`, `dst_port`, `edge_id`, `example_refs_total_count`, `example_refs_truncated`, `first_flow_start_utc`, `flow_row_count`, `ip_protocol`, `last_flow_end_utc`, `packets_sum`, `src_endpoint_id`. |

The `property_definitions[]` array MUST be ordered by Table 14-G4 row order, then projected-key order within the row.

Each property definition item MUST use Graph Projection's `property_definition` schema with:

- `property_definition_id='nf.pd.{target_scope}.{projected_key}.v1'`;
- `source_field_path='properties.{projected_key}'`;
- `required=true`;
- `missing_behavior='error'`;
- `source_null_behavior='emit_null'` only for `dst_port` and `source_null_behavior='error'` for every other projected key;
- `null_output_policy='emit_null'` only for `dst_port` and `null_output_policy='omit'` for every other projected key;
- `merge_behavior='single_value'`;
- `projected_type` from Table 14-G5.

Each property definition item MUST omit `default_value`. The adapter MUST NOT
submit Graph Projection-unknown members such as `mode`, `transform`,
`direct_copy`, `source_key`, or `target_key`.

**Table 14-G5. Graph Projection property type mapping**

| Projected key | `projected_type` |
| --- | --- |
| `endpoint_kind`, `edge_id`, `src_endpoint_id`, `dst_endpoint_id` | `identifier` |
| `endpoint_value`, `indicator_candidate_value`, `bytes_sum`, `packets_sum` | `string` |
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

The implementation MUST evaluate limits in this exact order after filter/time selection: distinct endpoint vertex count, distinct aggregate edge count, `bytes_sum` digit count by edge ID order, then `packets_sum` digit count by edge ID order. It MUST stop at the first failure and return respectively `reason_code='vertex_limit_exceeded'`, `edge_limit_exceeded`, `bytes_sum_digit_limit_exceeded`, or `packets_sum_digit_limit_exceeded`. For count limits, `actual_value` MUST be `limit_value + 1`, established by bounded streaming; for digit limits it MUST be the exact canonical decimal digit count of the first failing aggregate.

**NF-REQ-138a**
The adapter MUST map only public Graph Projection `project_ephemeral` outcomes according to Table 14-G6 and MUST NOT leak provider stack traces, internal kind registries, storage identifiers, Graph Projection validation issue details, private limit keys, or retained lifecycle selectors. A Graph Projection success that contains any validation issue is not a Network Flow partial success because Network Flow owns the adapter input construction.

**Table 14-G6. Graph Projection outcome mapping**

| Graph Projection outcome | Network Flow result |
| --- | --- |
| `project_ephemeral` success whose `data` is exactly one Graph Projection `ephemeral_projection_result`, `state='ephemeral_available'`, and `validation_summary` has `fatal_count=0`, `error_count=0`, `warning_count=0`, `info_count=0`, and `issues[]=[]` | Continue with Table 14-H response construction. |
| `invalid_projection_request`, `ephemeral_projection_failed` with `reason_code='fatal_validation'`, a success envelope with any validation issue, a result that contains retained lifecycle members, or a result that violates Table 14-G schema expectations | `network_flow_graph_projection_failed`, `reason_code='adapter_contract_rejected'`, `retry_action='do_not_retry'`, no partial graph output. |
| Caller cancellation before a Graph Projection result is available | `network_flow_graph_projection_failed`, `reason_code='projection_cancelled'`, `retry_action='do_not_retry'`, no partial graph output. |
| Deadline exceeded before a Graph Projection result is available | `network_flow_graph_projection_failed`, `reason_code='projection_timeout'`; `retry_action='retry_with_backoff'` only when Core classifies the cause transient, otherwise `retry_action='do_not_retry'`; no partial graph output. |
| `ephemeral_projection_failed` with `reason_code='projection_computation_failed'`, Graph Projection unavailable, or an implementation failure before a public Graph Projection outcome is available | `network_flow_graph_projection_failed`, `reason_code='projection_unavailable'`; `retry_action='retry_with_backoff'` only when Core classifies the cause transient, otherwise `retry_action='do_not_retry'`; no partial graph output. |

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

### 14.6 Graph response and contributor-query shape

**NF-REQ-141**
A successful graph query response `data` MUST contain Table 14-H members.

**Table 14-H. Graph query response data**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | Yes | `cartulary.network_flow_graph_query_result.v1`. |
| `graph_query_digest` | `sha256_hex_v1` | Yes | `network_flow_graph_query_digest_v1`. |
| `semantic_query` | `network_flow_graph_semantic_query_v1` | Yes | Exact default-materialized semantic query from Table 14-H2 used for the digest. |
| `graph_projection_result` | Graph Projection `ephemeral_projection_result` | Yes | Exact `project_ephemeral` success data used for response construction; retained lifecycle selectors and retained-run state are forbidden. |
| `edge_annotations[]` | array | Yes | Network Flow edge annotations from Table 14-H1 ordered by `edge_id ASC`. |
| `source_table_refs[]` | array of `network_flow_graph_source_table_ref_v1` | Yes | Exact objects from Table 14-H3 in workspace order. |
| `result_limits` | `network_flow_graph_result_limits_v1` | Yes | Exact object from Table 14-H4. |

**Table 14-H1. Network Flow edge annotation**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `edge_id` | edge ID | Yes | `network_flow_flow_edge_id_v1` output. |
| `example_row_refs[]` | array of `network_flow_row_ref_v1` | Yes | First retained refs under §14.5 ordering. Empty when `include_example_row_refs=false` or the effective example-ref limit is `0`. |
| `example_refs_truncated` | boolean | Yes | `true` only when omitted contributors exist. |
| `example_refs_total_count` | non-negative integer | Yes | Total contributing row ref count before truncation. |

**Table 14-H2. `network_flow_graph_semantic_query_v1`**

| Member | Required value |
| --- | --- |
| `schema_id` | Exactly `cartulary.network_flow.graph_semantic_query.v1`. |
| `selected_table_ids[]` | Exact resolved active table IDs in workspace order. |
| `filters[]` | `normalized_filters_v1`. |
| `time_range` | Exact object containing `start_utc` and `end_utc`; both null when request omitted the member. |
| `aggregation` | Exact object containing `mode` and `include_example_row_refs`, with defaults materialized. |
| `result_limits` | Exact Table 14-H4 object. |

**Table 14-H3. `network_flow_graph_source_table_ref_v1`**

| Member | Type | Required |
| --- | --- | ---: |
| `network_flow_table_id` | table ID | Yes |
| `table_version` | positive integer | Yes |
| `mapping_fingerprint` | `sha256_hex_v1` | Yes |
| `row_count_accepted` | non-negative integer | Yes |
| `row_count_rejected` | non-negative integer | Yes |

**Table 14-H4. `network_flow_graph_result_limits_v1`**

| Member | Type | Required |
| --- | --- | ---: |
| `max_vertices` | positive integer | Yes |
| `max_edges` | positive integer | Yes |
| `max_example_row_refs_per_edge` | non-negative integer | Yes |
| `max_aggregate_counter_digits` | positive integer | Yes |

**NF-REQ-141a**
`mapping_fingerprints[]` in adapter metadata MUST be the unique fingerprints of contributing tables in workspace table order. Entity and relationship metadata MUST include exactly the fingerprints of tables contributing to that entity or relationship. `source_metadata` MUST contain exactly `incident_id`, `graph_query_digest`, `source_snapshot_id`, `selected_table_ids[]`, and `mapping_fingerprints[]`; both arrays use workspace order.

**NF-REQ-141b**
`source_table_refs[]` MUST contain exactly one item for every `semantic_query.selected_table_ids[]` item in the same order and no item for an unselected table. `edge_annotations[]` MUST contain exactly one annotation for every returned projected Network Flow edge and no extra annotation; `edge_id` is the join key. A mismatch is `network_flow_graph_projection_failed` with `reason_code='adapter_contract_rejected'`, not a partial success.

**NF-REQ-141c**
The contributor-query route MUST accept an initial `network_flow_graph_contributor_query_request_v1` containing exactly `schema_id='cartulary.network_flow.graph_contributor_query_request.v1'`, `graph_query`, `graph_query_digest`, `selector`, and optional `limit`. `graph_query` MUST be the exact `network_flow_graph_semantic_query_v1` returned by Table 14-H. `selector` is exactly one closed variant: `{ "kind":"vertex", "vertex_id":... }` or `{ "kind":"edge", "edge_id":... }`. `limit` defaults and validates under Table 13-A. A continuation uses only the §13.1 continuation variant.

The server MUST recompute the semantic graph query and digest under current authorization, require the selected vertex or edge to exist, and return `data` with exactly `schema_id='cartulary.network_flow.graph_contributor_query_result.v1'`, `graph_query_digest`, `selector`, `contributors[]`, and `meta`. `meta` MUST satisfy Table 17-E4C. Each contributor contains exactly `row_ref` and `row`, where `row_ref` satisfies Table 12-B and `row` satisfies Table 12-A. Contributors are ordered by workspace table order, then `effective_sort_v1([])`, then `network_flow_row_id ASC`; that full tuple plus table and row ID is the keyset continuation comparator. A stale digest or missing selected object MUST fail with `network_flow_graph_query_stale`; contributor output MUST never include rejected rows.

## 15. Indicator linking and observation behavior

### 15.1 Link request schema

**NF-REQ-142**
The indicator-link route MUST accept `indicator_link_request_v1` with Table 15-A members.

**Table 15-A. `indicator_link_request_v1`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow.indicator_link_request.v1`. |
| `client_txn_id` | `client_txn_id` | Yes | No | Mutating route idempotency key. |
| `selector` | object | Yes | No | One selector from Table 15-B. |
| `target` | object | Yes | No | One target from Table 15-C. |
| `observation_mode` | token | Yes | No | Exactly `binding_only` in v1. |
| `confirm_exact_value` | `ip_literal_v1` string | Yes | No | Exact canonical candidate IP echoed by UI to prevent ambiguous links. |

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
| `graph_vertex` | `graph_query`, `graph_query_digest`, `vertex_id` | `graph_query` is `network_flow_graph_semantic_query_v1`; `vertex_id` is a `network_flow_endpoint_id`. Server recomputes graph query digest and validates that the selected endpoint vertex exists in the recomputed composition. A `field_key` member is invalid. |
| `graph_edge` | `graph_query`, `graph_query_digest`, `edge_id`, `field_key` | `graph_query` is `network_flow_graph_semantic_query_v1`; `edge_id` is a `network_flow_flow_edge_id`. `field_key` must be `network_flow.src_ip` or `network_flow.dst_ip`. Server recomputes graph query digest and validates that the selected edge exists in the recomputed composition plus exact candidate value. |

Each selector is a closed variant and MUST contain exactly `kind` plus the members shown for that kind. Members from another selector kind, aliases, visible labels, display row numbers, and client-supplied candidate values are invalid.

**Table 15-C. Link target modes**

| `target.mode` | Required members | Required behavior |
| --- | --- | --- |
| `existing_indicator` | `indicator_id` | Validate same incident and current visibility; Core owns canonical indicator. The target indicator MUST have Core `value_kind='atomic'`, Core normalized value equal to the resolved candidate IP, and the Core indicator type designated for IP literals. |
| `create_indicator` | `indicator_type` | Delegate canonical creation and dedupe to Core using the server-resolved candidate as both requested display and normalization input. `indicator_type` MUST be the Core registry token designated for IP literals. Client-supplied `value_kind`, `display_value`, or `normalized_value` members are invalid. |

Each target is a closed variant and MUST contain exactly `mode` plus the members shown. Unknown target members MUST fail with `network_flow_invalid_indicator_target`.

If Core has no indicator type designated for canonical IP literals, `create_indicator` from Network Flow is an adoption blocker and MUST fail closed with Core indicator-create validation behavior until Core 02 or the adopted Core indicator registry closes that dependency.

**NF-REQ-143**
The link route MUST reject selectors that reference rejected rows, soft-deleted tables, missing rows, `unmapped_raw`, graph layout coordinates, visible graph labels, visible row numbers, or stale graph query digests.

**NF-REQ-143a**
For non-replay requests, after selector resolution and before any mutation or committed idempotency record, the implementation MUST require `confirm_exact_value` itself to be canonical `ip_literal_v1` text and compare it byte-for-byte with the resolved canonical candidate. The route MUST NOT normalize a noncanonical confirmation. A mismatch or noncanonical confirmation MUST fail with `network_flow_indicator_link_ambiguous`. Because the link route is incident-authorized, the error details MAY include the resolved candidate value. Omission behavior: if a future non-incident-authorized context reuses the error family, it MUST include only `network_flow_safe_digest_v1` of the resolved candidate value.

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
| `candidate_value` | `ip_literal_v1` | Yes | Canonical candidate IP linked. |
| `source_row_refs[]` | array | Yes | Length `1..network_flow.max_binding_source_row_refs`; a graph selector with no contributing accepted rows is invalid. |
| `source_row_refs_truncated` | boolean | Yes | True when retained `source_row_refs[]` omits contributing accepted rows. |
| `source_row_refs_total_count` | positive integer | Yes | Total source-row-ref count before truncation. |
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

All persisted and returned `source_row_refs[]` arrays MUST use §14.5 contributor order, including a caller-supplied `row_refs` selector after duplicate validation.

**NF-REQ-144b**
For a non-replay request, Core indicator resolution or canonical create/dedupe, binding identity comparison, optional binding insertion, idempotency-response commit, and the required created-or-reused audit occurrence MUST execute in one Core unit of work. Failure MUST roll back any newly created indicator and binding together. A new binding returns a result object containing exactly `schema_id`, `binding`, and `duplicate=false`; reuse returns the same shape with the existing binding and `duplicate=true`. `duplicate` is response metadata and MUST NOT be persisted in the binding resource.

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
| `network_flow_table_created` | `incident_id`, `actor_user_id`, `network_flow_table_id`, `source_filename_digest`, `source_filename_digest_key_id`, `source_content_sha256`, `source_profile_id`, `parser_profile_id`, `mapping_fingerprint`, `row_count_accepted`, `row_count_rejected`. |
| `network_flow_table_renamed` | `incident_id`, `actor_user_id`, `network_flow_table_id`, `old_display_name_digest`, `new_display_name_digest`, `display_name_digest_key_id`, `table_version`. |
| `network_flow_table_soft_deleted` | `incident_id`, `actor_user_id`, `network_flow_table_id`, `table_version`, `row_count_accepted`, `row_count_rejected`. |
| `network_flow_graph_query_executed` | `incident_id`, `actor_user_id`, `graph_query_digest_safe`, `graph_query_digest_safe_key_id`, `selected_table_count`, `result_vertex_count`, `result_edge_count`, `truncated_example_ref_count`. `graph_query_digest_safe` is `network_flow_safe_digest_v1("graph_query_digest", graph_query_digest)`. |
| `network_flow_indicator_binding_created` | `incident_id`, `actor_user_id`, `network_flow_indicator_binding_id`, `target_indicator_record_id`, `selector_kind`, `candidate_value_digest`, `candidate_value_digest_key_id`, `source_row_ref_count`. |
| `network_flow_indicator_binding_reused` | `incident_id`, `actor_user_id`, `network_flow_indicator_binding_id`, `target_indicator_record_id`, `selector_kind`, `candidate_value_digest`, `candidate_value_digest_key_id`, `source_row_ref_count`. |

**NF-REQ-151**
Audit event payloads MUST NOT include raw display names, raw filenames, raw CSV cells, raw graph query scalar values, or raw indicator candidates unless the same value is already a Core stable identifier and is safe under Core audit rules.

Every safe digest in an audit event MUST be accompanied by the `safe_key_id_v1` used to compute it. Audit consumers MUST compare safe digests only when their key IDs are equal.

**NF-REQ-151a**
Audit occurrence counts MUST follow Table 16-C. “One” means exactly one committed domain occurrence per non-replay operation, even when an internal worker retries.

**Table 16-C. Audit occurrence matrix**

| Operation outcome | Required domain occurrence |
| --- | --- |
| Import final commit succeeds | One `network_flow_table_created` per created table. |
| Rename commits a changed display name | One `network_flow_table_renamed`. A request whose normalized name equals the current name succeeds as an unchanged resource and emits no rename occurrence. |
| Soft delete commits | One `network_flow_table_soft_deleted`. |
| Graph query succeeds | One `network_flow_graph_query_executed`; failed, cancelled, or over-limit graph queries emit none of this domain family. |
| Indicator link inserts a binding | One `network_flow_indicator_binding_created`. |
| Indicator link reuses a binding under a new `client_txn_id` | One `network_flow_indicator_binding_reused`. |
| Exact committed idempotency replay | No new domain occurrence; return the originally committed response and its original audit correlation. |

For `network_flow_graph_query_executed`, `truncated_example_ref_count` MUST equal the sum over all returned edges of `example_refs_total_count - length(example_row_refs[])`; when examples are disabled, each returned edge contributes its full `example_refs_total_count`.

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
| `nf.graphs.contributors.query` | `POST /api/v1/incidents/{incident_id}/network-flow/graphs/contributors/query` | Page currently authorized accepted rows contributing to one recomputed graph vertex or edge. |
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
| `nf.graphs.contributors.query` | `viewer` | Contributor request from §14.6 | Default query limit from Table 13-A | Read route | `network_flow_graph_contributor_query_result.v1` | graph-stale, scope, cursor, and query errors | none |
| `nf.indicator_links.create` | `editor` | `indicator_link_request_v1` | No omitted members | Required | `network_flow_indicator_link_result.v1` | indicator-link errors | created-or-reused event from Table 16-C |

All successful GET, PATCH, DELETE, and read-query routes in Table 17-A return HTTP `200`. `nf.indicator_links.create` returns `201` when it inserts a binding and `200` when it reuses an existing binding under a new idempotency key. Exact replay returns the original committed status and body. Error HTTP status is the exact Table 21-B mapping; a route MUST NOT select status by implementation exception class.

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
`GET /tables/{network_flow_table_id}` MUST return active table metadata only. A soft-deleted, unknown, cross-incident, or hidden table MUST NOT leak cross-incident existence. The implementation MUST use Core hidden-resource behavior when hiding is required by Core authorization rules.

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
| `schema_id` | string | Yes | No | none | Exactly `cartulary.network_flow.rejected_rows_query_request.v1`. |
| `error_codes[]` | array | No | No | `[]` | Empty means no error-code filter. When non-empty, length `1..64`; values from §21; duplicates and unknown tokens invalid. |
| `field_keys[]` | array | No | No | `[]` | Empty means no field-key filter. When non-empty, length `1..64`; values from Table 9-D; duplicates and unknown tokens invalid. |
| `source_row_range` | object | No | No | omitted | Optional `{ "gte": positive_int\|null, "lte": positive_int\|null }`; at least one non-null when present; `gte > lte` invalid with `reason_code='empty_range'`. |
| `limit` | integer | No | No | `min(200, effective network_flow.max_query_limit)` | `1..effective network_flow.max_query_limit`. |

Invalid rejected-row query arrays, duplicate values, unknown tokens, and empty ranges MUST fail with `network_flow_invalid_filter`. Continuation uses the exact continuation-only variant from NF-REQ-115a.

### 17.8 Query response shapes

**NF-REQ-162**
A single-table accepted-row query response `data` MUST contain Table 17-E1 members.

**Table 17-E1. Single-table accepted-row query response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_table_query_result.v1`. |
| `network_flow_table_id` | table ID | Yes | No | Queried active table. |
| `rows[]` | array of `network_flow_row` | Yes | No | Accepted rows ordered by `effective_sort_v1`; each row includes `network_flow_table_id`. |
| `meta` | object | Yes | No | Exact accepted-row metadata from Table 17-E4A. |

**NF-REQ-163**
A cross-table accepted-row query response `data` MUST contain Table 17-E2 members.

**Table 17-E2. Cross-table accepted-row query response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_rows_query_result.v1`. |
| `table_scope` | normalized table scope object | Yes | No | Contains exactly `mode` and resolved `table_ids[]` in workspace order. |
| `rows[]` | array of `network_flow_row` | Yes | No | Accepted rows ordered by `effective_sort_v1`; every row includes `network_flow_table_id`. |
| `meta` | object | Yes | No | Exact accepted-row metadata from Table 17-E4A. |

**NF-REQ-164**
A rejected-row query response `data` MUST contain Table 17-E3 members.

**Table 17-E3. Rejected-row query response `data`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exactly `cartulary.network_flow_rejected_rows_query_result.v1`. |
| `network_flow_table_id` | table ID | Yes | No | Queried active table. |
| `diagnostics[]` | array of rejected-row diagnostics | Yes | No | Ordered by §12.4, after filters and cursor continuation. |
| `meta` | object | Yes | No | Exact diagnostic metadata from Table 17-E4B. |

**Table 17-E4A. Accepted-row query response `meta`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `query` | object | Yes | No | Contains exactly `filters[]`, `sort[]`, `effective_sort[]`, and `table_ids[]`; all values are normalized and default-materialized, and table IDs use workspace order. |
| `paging` | object | Yes | No | Contains exactly `limit`, `returned_count`, and `next_cursor_token`. `next_cursor_token` follows NF-REQ-126c. |

**Table 17-E4B. Rejected-row query response `meta`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `query` | object | Yes | No | Contains exactly normalized `error_codes[]` ordered by token, `field_keys[]` in Table 9-D order, `source_row_range`, and `effective_sort`, where `source_row_range` is null when omitted and `effective_sort` is the diagnostic comparator from NF-REQ-126b. |
| `paging` | object | Yes | No | Contains exactly `limit`, `returned_count`, and `next_cursor_token`. |

**Table 17-E4C. Contributor query response `meta`**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `paging` | object | Yes | No | Contains exactly `limit`, `returned_count`, and `next_cursor_token`; cursor uses the contributor comparator from §14.6. |

Query response objects MUST NOT include raw cursor internals, SQL fragments, storage field names, table display names as identity, or visual row positions.

### 17.9 Indicator link response

**NF-REQ-165**
A successful indicator-link response `data` MUST contain Table 17-E members.

**Table 17-E. Indicator link response data**

| Member | Type | Required |
| --- | --- | ---: |
| `schema_id` | string `cartulary.network_flow_indicator_link_result.v1` | Yes |
| `binding` | `network_flow_indicator_binding` | Yes |
| `duplicate` | boolean | Yes |

## 18. Import apply result integration

**NF-REQ-166**
A terminal successful owner-facade apply MUST return exactly the `network_flow_import_unit_result_v1` members in Table 18-A0. A Core multi-unit import result MUST embed one such result for each applied Network Flow unit in Core import-unit order.

**Table 18-A0. `network_flow_import_unit_result_v1`**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | Yes | Exactly `cartulary.network_flow.import_unit_result.v1`. |
| `import_session_id` | Core import-session ID | Yes | Source session. |
| `import_unit_id` | Core import-unit ID | Yes | Applied unit. |
| `source_content_sha256` | `sha256_hex_v1` | Yes | Exact source bytes. |
| `source_profile_id` | token | Yes | Approved claimable profile. |
| `parser_profile_id` | token | Yes | Approved parser profile. |
| `mapping_fingerprint` | `sha256_hex_v1` | Yes | Approved materialized mapping. |
| `table_ref` | object | Yes | Exact Table 18-A object. |

**Table 18-A. Import result table reference**

| Member | Type | Required | Rule |
| --- | --- | ---: | --- |
| `kind` | string | Yes | Exactly `network_flow_table`. |
| `id` | table ID | Yes | Created table ID. |
| `route` | string | Yes | `/api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}` with concrete IDs in the response. |
| `display_name` | string | Yes | Table display name. |
| `row_count_accepted` | integer | Yes | Accepted count. |
| `row_count_rejected` | integer | Yes | Rejected count. |
| `diagnostics_truncated` | Boolean | Yes | Matches the created table. |
| `table_version` | positive integer | Yes | Exactly `1` at creation. |

**NF-REQ-167**
Replay of a committed import apply action with the same route-scoped idempotency key and normalized request MUST return the same created table references and MUST NOT create additional tables.

## 19. Graph UI interaction, pivots, and table synchronization

**NF-REQ-168**
Selecting a graph vertex MUST call `nf.graphs.contributors.query` with the stable `vertex_id`, `graph_query_digest`, and semantic graph query. The pivot MUST NOT use graph layout coordinates, visible labels, rendered order, browser-local graph node indices, or local filtering of example refs.

**NF-REQ-169**
Selecting a graph edge MUST call `nf.graphs.contributors.query` and open the returned contributing rows drawer grouped by table. Group ordering MUST follow workspace table order. Row ordering inside a group MUST follow `effective_sort_v1([])` from §13.4. A user-supplied pivot sort is a subsequent row-query request and MUST use §13 query sorting; it is not part of graph selector identity.

**NF-REQ-170**
When any active table in a displayed graph's table scope is renamed, the graph data remains semantically valid but display metadata may be stale. The UI MUST update table display labels on the next table-metadata refresh without changing graph query digest. When any active table in graph scope is soft-deleted, the displayed graph MUST become `graph_stale` and any selector action against it MUST fail until the graph is recomputed without the deleted table.

Core Document 03 invalidation delivery MUST invalidate table metadata on rename and graph/query/contributor state on soft delete or authorization loss. A later generic incident-removal profile must define any whole-incident removal invalidation before Network Flow can claim that behavior. Invalidation is advisory for freshness only: every subsequent route call MUST independently reauthorize and revalidate lifecycle state. A missed invalidation MUST therefore cause at most stale presentation, never unauthorized data disclosure or a successful stale selector mutation.

## 20. Resource limits

**NF-REQ-171**
Network Flow resource limits MUST use Table 20-A. Deployments MAY lower limits to the lowerable minimum. Omission behavior: deployments that omit a limit use the default. Deployments MUST NOT raise a limit above the default in this revision.

Limit configuration MUST be validated at process configuration load before serving Network Flow routes or import facade calls. A configured value that is absent uses the default. A configured value that is not an integer, is below the lowerable minimum, or is above the default-and-maximum value is invalid configuration. The effective `network_flow.max_active_tables_per_incident` MUST NOT exceed `network_flow.max_retained_tables_per_incident`. The implementation MUST fail configuration validation and MUST NOT silently clamp, ignore, round, or repair an invalid value or cross-limit relationship.

**Table 20-A. Resource limit registry**

| Limit key | Default and maximum | Lowerable minimum | Enforcement phase | Exceeded behavior |
| --- | ---: | ---: | --- | --- |
| `network_flow.max_active_tables_per_incident` | 128 | 1 | Final import commit. | Fail with `network_flow_table_limit_exceeded`; commit no table. |
| `network_flow.max_retained_tables_per_incident` | 512 | 1 | Final import commit. | Fail with `network_flow_table_limit_exceeded`; commit no table. |
| `network_flow.max_selected_tables_per_query` | 64 | 1 | Cross-table row, graph, and contributor query admission. | Fail with `network_flow_invalid_table_scope` and `reason_code='selected_table_limit_exceeded'`; emit no result data. |
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
| `network_flow.max_aggregate_counter_digits` | 128 | 1 | Aggregation. | Fail with `network_flow_counter_sum_limit_exceeded`; return no partial graph. |

**NF-REQ-172**
Every limit-exceeded error MUST include `limit_key`, `limit`, `actual`, and `phase` in `error.details`. Count-based streaming enforcement MUST stop after proving one excess item and set `actual=limit+1`; it MUST NOT scan further merely to disclose a total. Scalar-length and aggregate-digit errors MUST report the exact observed scalar or digit length. These values are available only on incident-authorized routes; logs and telemetry MUST omit `actual` unless Core policy permits it.

**Table 20-B. Fixed non-configurable protocol bounds**

| Bound | Value | Failure |
| --- | ---: | --- |
| CSV preview logical data records | 50 | Stop preview normally after the 50th complete record. |
| Mapping/diagnostic sample array | 50 | Retain the first 50 under specified order. |
| `in` filter array length | 256 | `network_flow_invalid_filter`. |
| Rejected-row `error_codes[]` or `field_keys[]` length | 64 | `network_flow_invalid_filter`. |
| Cursor token ASCII byte length | 4096 | `network_flow_cursor_invalid`. |
| RFC3339 fractional second digits | 6 | `network_flow_invalid_timestamp`. |
| Identifier random-collision attempts | 8 | `network_flow_id_generation_failed`. |

These are wire- or algorithm-version constants, not deployment configuration. Changing one requires a compatible protocol revision or a new versioned algorithm/profile as applicable.

### 20.1 Deployment key-ring configuration

**NF-REQ-172a**
The adopted top-level deployment-configuration namespace for this profile is
`network_flow_activity`. It contains exactly the following keys:

| Key | Type | Required | Default and omitted behavior |
| --- | --- | ---: | --- |
| `claimed` | boolean | No | `false`; Core 04 owns claim parsing and resolution. |
| `key_ring_manifest_path` | string | Required only when `claimed=true` | No default; `inactive_policy='forbidden'`, `inactive_value_schema_ref=null`, and it MUST be absent when `claimed=false`. |

`key_ring_manifest_path` MUST be an absolute normalized path with no NUL,
shell-variable form, `~`, or lexical `.` or `..` segment. Explicit `null` is
invalid. When the profile is unclaimed, a supplied path MUST fail with
`profile_incompatible_binding`. When the profile is claimed, an omitted,
empty, unreadable, non-regular, oversized, malformed, or schema-invalid
manifest MUST fail before any HTTP listener, WebSocket listener, or background
worker starts. Environment overlays use
`CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH` under the Core 04
overlay grammar.

**NF-REQ-172b**
The selected file MUST be UTF-8 JSON no larger than `65536` bytes and MUST
validate as `cartulary.network_flow_key_rings.v1`. Duplicate object members,
unknown members, and explicit `null` are invalid at every object boundary. The
top-level object contains exactly `schema_id`, `cursor_key_ring`, and
`safe_digest_key_ring`.

`cursor_key_ring` contains exactly `algorithm` and `keys`. `algorithm` is
`aes_256_gcm_v1`. `keys` contains `1..8` objects with exactly
`cursor_key_id`, `state`, `secret_ref`, and the state-dependent time members.
Exactly one key has `state='active'`; it MUST omit `deactivated_at` and
`retire_at`. A key with `state='decrypt_only'` MUST contain both timestamps in
`timestamp_utc_v1` form, MUST satisfy
`retire_at = deactivated_at + 15 minutes`, and MUST NOT validate a token issued
at or after `deactivated_at` or at or after `retire_at`.

`safe_digest_key_ring` contains exactly `algorithm` and `keys`. `algorithm` is
`hmac_sha256_v1`. `keys` contains `1..8` objects with exactly
`safe_digest_key_id`, `state`, `secret_ref`, and the state-dependent time
members. Exactly one key has `state='active'`; it MUST omit `deactivated_at`
and `retain_until`. A key with `state='inactive'` MUST contain both timestamps
in `timestamp_utc_v1` form and MUST satisfy `retain_until > deactivated_at`.
Inactive keys MUST NOT emit new digests.

Every key ID MUST satisfy `safe_key_id_v1`; IDs MUST be unique within their
ring. A `secret_ref` MUST satisfy Core 04 `secret_ref_v1`. Its resolved value
MUST be unpadded base64url text decoding to exactly 32 bytes. Raw or derived
key material is forbidden in the manifest, configuration, diagnostics,
readiness output, logs, telemetry, audit, public routes, browser state, and
ordinary fixtures. Expired retained entries are invalid at startup and MUST be
removed from a running provider no later than their retirement instant.

**NF-REQ-172c**
Cursor and safe-digest rings MUST be loaded and validated atomically. Reusing a
secret reference name or resolved key material across the two rings or across
another startup-resolved secret purpose is invalid. Ordinary development and
production assembly MUST NOT derive either ring from an authentication master
secret, use fallback material, or load fixture-only key IDs or values.
Harness-owned assembly MAY inject deterministic keys only through guarded
harness controls. Omission behavior: without an active guarded harness control,
assembly resolves and validates the deployment manifest exactly as production
does.

Rotation is restart-applied in this revision; live manifest reload is not
defined. All claimed nodes serving one deployment MUST use the same manifest
epoch. Previously persisted safe digests are retained with their original key
IDs and MUST NOT be rewritten. Cursor tokens produced by an implementation
before the `nfc2` envelope are unsupported and fail through
`network_flow_cursor_invalid`.

**NF-REQ-173**
Effective limits MUST be exposed through `GET /source-profiles` under `data.effective_limits`. The route MUST return all keys in Table 20-A exactly once.

## 21. Error registry and detail schemas

### 21.1 Error code registry

**NF-REQ-174**
Network Flow errors MUST use Table 21-A where route-local errors are required.

**Table 21-A. Error code registry**

| Error code | Exact HTTP status or non-route scope | Use |
| --- | ---: | --- |
| `network_flow_invalid_request` | 400 | JSON admission, unknown member, type mismatch, or route-local schema failure. |
| `network_flow_unsupported_source_profile` | 400 | Reserved or unsupported profile requested. |
| `network_flow_invalid_utf8` | 400 | CSV source is not valid UTF-8. |
| `network_flow_csv_empty_file` | 400 | CSV source has zero bytes after BOM removal. |
| `network_flow_invalid_header` | 400 | First logical record contains a forbidden header scalar or violates the header contract. |
| `network_flow_no_data_rows` | 400 | CSV contains header but no data rows. |
| `network_flow_csv_malformed_quote` | 400 | Quote grammar invalid. |
| `network_flow_source_changed` | 409 | Apply source bytes, descriptors, or mapping fingerprint differ from approval. |
| `network_flow_csv_field_count_mismatch` | row diagnostic | Row field count differs from header. |
| `network_flow_mapping_required` | 400 | Mapping incomplete or invalid. |
| `network_flow_mapping_conflict` | 400 | Mapping rows conflict or persisted defaults invalid. |
| `network_flow_invalid_timestamp` | row diagnostic | Timestamp parse/profile failure. |
| `network_flow_end_before_start` | row diagnostic | Flow end earlier than start. |
| `network_flow_invalid_ip` | row diagnostic | IP parse/canonicalization failure. |
| `network_flow_invalid_port` | row diagnostic | Port invalid. |
| `network_flow_invalid_protocol` | row diagnostic | Protocol invalid. |
| `network_flow_invalid_counter` | row diagnostic | Counter invalid. |
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
| `network_flow_invalid_limit_override` | 400 | Graph limit override invalid. |
| `network_flow_graph_limit_exceeded` | 413 | Graph exceeds vertex or edge limit. |
| `network_flow_counter_sum_limit_exceeded` | 413 | Aggregated counter sum exceeds digit limit. |
| `network_flow_graph_projection_failed` | 502 | Ephemeral Graph Projection adapter could not produce a conformant result. |
| `network_flow_graph_query_stale` | 409 | Graph digest or selected graph object no longer matches current authorized composition. |
| `network_flow_indicator_link_ambiguous` | 400 | Link action does not identify candidate value. |
| `network_flow_invalid_indicator_selector` | 400 | Link selector invalid or stale. |
| `network_flow_invalid_indicator_target` | 400 | Link target variant or target compatibility invalid. |
| `network_flow_indicator_link_forbidden` | 403 | Authorization or target visibility fails. |
| `network_flow_external_enrichment_forbidden` | 400 | Request attempts forbidden third-party enrichment. |
| `network_flow_id_generation_failed` | 500 | Eight CSPRNG identifier collision attempts failed. |

No code omitted from Table 21-A is a Network Flow route-local error in v1. Row-diagnostic entries MUST NOT be emitted as top-level route errors except when another Table 21-A import-unit error, including `network_flow_all_rows_rejected`, carries them as bounded diagnostic data.

**Table 21-A1. Exhaustive reason-code registry**

| Error code or family | Permitted `reason_code` values |
| --- | --- |
| `network_flow_invalid_request` | `duplicate_member`, `malformed_json`, `body_not_object`, `unknown_member`, `missing_member`, `explicit_null`, `type_mismatch`, `invalid_schema_id`, `variant_member_conflict` |
| `network_flow_unsupported_source_profile` | `reserved_profile`, `unknown_profile` |
| `network_flow_invalid_utf8` | `invalid_utf8_sequence`, `bom_not_at_offset_zero` |
| `network_flow_csv_empty_file` | `zero_bytes`, `bom_only` |
| `network_flow_invalid_header` | `forbidden_header_control` |
| `network_flow_no_data_rows` | `header_only` |
| `network_flow_csv_malformed_quote` | `invalid_after_closing_quote`, `unclosed_quote` |
| `network_flow_source_changed` | `content_digest_mismatch`, `header_descriptor_mismatch`, `mapping_fingerprint_mismatch` |
| `network_flow_csv_field_count_mismatch` | `field_count_mismatch` |
| `network_flow_mapping_required` | `required_field_unmapped`, `system_derivation_missing`, `preview_stale` |
| `network_flow_mapping_conflict` | `source_column_reused`, `target_field_duplicated`, `field_not_supported_by_profile`, `mapping_kind_unavailable`, `transform_target_mismatch`, `timestamp_column_reused`, `invalid_empty_value_policy`, `unaccounted_source_column`, `variant_member_conflict` |
| `network_flow_invalid_timestamp` | `missing_or_empty`, `invalid_syntax`, `precision_exceeded`, `out_of_range`, `ambiguous_local_time`, `nonexistent_local_time`, `sys_uptime_invalid`, `sys_uptime_wrap_ambiguous` |
| `network_flow_end_before_start` | `end_before_start` |
| `network_flow_invalid_ip` | `missing_or_empty`, `invalid_syntax`, `noncanonical_value` |
| `network_flow_invalid_port` | `missing_or_empty`, `invalid_syntax`, `out_of_range` |
| `network_flow_invalid_protocol` | `missing_or_empty`, `invalid_syntax`, `unknown_token`, `out_of_range` |
| `network_flow_invalid_counter` | `missing_or_empty`, `invalid_syntax`, `out_of_range` |
| `network_flow_all_rows_rejected` | `zero_accepted_rows` |
| `network_flow_table_name_exhausted` | `suffix_space_exhausted` |
| `network_flow_table_not_found` | `not_found` |
| `network_flow_table_not_active` | `soft_deleted` |
| `network_flow_table_version_conflict` | `stale_version` |
| `network_flow_invalid_display_name` | `empty_display_name`, `display_name_too_long`, `duplicate_display_name`, `forbidden_control` |
| `network_flow_invalid_table_scope` | `unknown_mode`, `variant_member_conflict`, `duplicate_table_id`, `selected_table_limit_exceeded`, `empty_resolved_scope`, `table_not_active` |
| `network_flow_invalid_filter` | `unknown_field`, `operator_not_allowed`, `invalid_value`, `duplicate_in_value`, `duplicate_filter`, `empty_range`, `too_many_filters` |
| `network_flow_invalid_sort` | `unknown_field`, `field_not_sortable`, `invalid_direction`, `duplicate_sort_field`, `too_many_sorts` |
| `network_flow_invalid_limit`, `network_flow_invalid_limit_override` | `not_integer`, `below_minimum`, `above_maximum`, `unknown_limit_key` |
| `network_flow_cursor_invalid` | `mixed_initial_and_continuation`, `malformed`, `too_long`, `expired`, `actor_mismatch`, `route_mismatch`, `semantic_query_mismatch`, `scope_stale`, `authorization_lost` |
| `network_flow_invalid_time_range` | `both_bounds_null`, `empty_range`, `invalid_bound` |
| `network_flow_graph_limit_exceeded` | `vertex_limit_exceeded`, `edge_limit_exceeded` |
| `network_flow_counter_sum_limit_exceeded` | `bytes_sum_digit_limit_exceeded`, `packets_sum_digit_limit_exceeded` |
| `network_flow_graph_projection_failed` | `adapter_contract_rejected`, `projection_cancelled`, `projection_timeout`, `projection_unavailable` |
| `network_flow_graph_query_stale` | `digest_mismatch`, `vertex_not_found`, `edge_not_found`, `scope_stale` |
| Indicator-link errors | `unknown_selector_kind`, `variant_member_conflict`, `field_not_linkable`, `row_not_accepted`, `candidate_mismatch`, `target_not_visible`, `target_value_mismatch`, `target_type_mismatch`, `core_ip_indicator_type_unavailable` |
| `network_flow_table_limit_exceeded`, `network_flow_resource_limit_exceeded` | `active_table_limit_exceeded`, `retained_table_limit_exceeded`, `column_limit_exceeded`, `header_scalar_limit_exceeded`, `cell_scalar_limit_exceeded`, `row_limit_exceeded`, `accepted_row_limit_exceeded` |
| `network_flow_external_enrichment_forbidden` | `capability_unavailable` |
| `network_flow_id_generation_failed` | `collision_attempts_exhausted` |

An emitted reason code not permitted by Table 21-A1 is non-conformant. Codes whose detail schema does not otherwise require `reason_code` MUST still include it.

### 21.2 Error detail schema registry

**NF-REQ-175**
Route-local errors MUST include details according to Table 21-B. A detail member listed as required MUST be present even when its value is JSON `null`, unless the row states that the member is conditionally present.

**Table 21-B. Error detail schemas**

| Error code or scope | Required `error.details` members |
| --- | --- |
| `network_flow_invalid_request` | `field`, `reason_code`, `expected_contract`, `actual_kind`. |
| `network_flow_unsupported_source_profile` | `source_profile_id`, `conformance_status`, `allowed_profile_ids[]`. |
| CSV import-unit failures: `network_flow_invalid_utf8`, `network_flow_csv_empty_file`, `network_flow_invalid_header`, `network_flow_no_data_rows`, `network_flow_csv_malformed_quote` | `phase`, `reason_code`, `source_row_number`, `source_column_ordinal`. Row and column members are `null` when not attributable. |
| `network_flow_source_changed` | `reason_code`, `approved_source_content_sha256`, `observed_source_content_sha256`, `approved_mapping_fingerprint`, `observed_mapping_fingerprint`; descriptor-sensitive values are omitted. |
| CSV row diagnostics: `network_flow_csv_field_count_mismatch`, `network_flow_invalid_timestamp`, `network_flow_end_before_start`, `network_flow_invalid_ip`, `network_flow_invalid_port`, `network_flow_invalid_protocol`, `network_flow_invalid_counter`, row-level `network_flow_resource_limit_exceeded` | Exact Table 12-E resource including `source_row_number`, `source_column_ordinal`, `field_key`, `reason_code`, safe message fields, limit fields, `safe_sample`, and `raw_value_sha256`. |
| Mapping failures: `network_flow_mapping_required`, `network_flow_mapping_conflict` | `field_key`, `source_column_ordinal`, `mapping_kind`, `reason_code`. |
| `network_flow_all_rows_rejected` | `row_count_rejected`, `diagnostics_truncated`, `diagnostics_sample[]`. |
| Limit failures: `network_flow_table_limit_exceeded`, `network_flow_resource_limit_exceeded`, `network_flow_graph_limit_exceeded`, `network_flow_counter_sum_limit_exceeded` | `limit_key`, `limit`, `actual`, `phase`. |
| `network_flow_table_name_exhausted` | `reason_code`, `base_display_name_digest`, `base_display_name_digest_key_id`, `attempted_suffix_min`, `attempted_suffix_max`. |
| Table lookup/state failures: `network_flow_table_not_found`, `network_flow_table_not_active` | `network_flow_table_id`, `table_status`, `allowed_states[]`. Hidden resources MUST use Core hidden-resource details instead when Core requires non-disclosure. |
| `network_flow_table_version_conflict` | `network_flow_table_id`, `base_table_version`, `current_table_version`. |
| `network_flow_invalid_display_name` | `field`, `reason_code`, `max_length`, `normalized_length`. |
| `network_flow_invalid_table_scope` | `reason_code`, `mode`, `table_ids[]`, `limit_key`. |
| `network_flow_invalid_filter` | `field_key`, `op`, `reason_code`, `filter_index`. |
| `network_flow_invalid_sort` | `field_key`, `direction`, `reason_code`, `sort_index`. |
| `network_flow_invalid_limit` | `limit_key`, `limit`, `minimum`, `maximum`, `reason_code`. |
| `network_flow_cursor_invalid` | `reason_code`, `cursor_scope`, `retry_action`. |
| `network_flow_invalid_time_range` | `field`, `reason_code`, `start_utc`, `end_utc`. |
| `network_flow_invalid_limit_override` | `limit_key`, `limit`, `minimum`, `maximum`, `reason_code`. |
| `network_flow_graph_projection_failed` | `reason_code`, `retry_action`, `projection_contract_version`; internal exception and provider details forbidden. |
| `network_flow_graph_query_stale` | `reason_code`, `graph_query_digest`, `retry_action`; selected object IDs MAY be included only after current authorization. Omission behavior: omit selected object IDs. |
| Indicator-link failures: `network_flow_indicator_link_ambiguous`, `network_flow_invalid_indicator_selector`, `network_flow_invalid_indicator_target`, `network_flow_indicator_link_forbidden` | `selector_kind`, `field_key`, `target_mode`, `reason_code`, and either `resolved_candidate_value` for incident-authorized route responses or `resolved_candidate_safe_digest` plus `resolved_candidate_safe_digest_key_id` otherwise. |
| `network_flow_external_enrichment_forbidden` | `requested_capability`, `reason_code`. |
| `network_flow_id_generation_failed` | `reason_code`, `attempt_count`; generated candidate IDs forbidden. |

Every route-local error MUST include `retry_action` at the top level of `error.details` with a value from Table 21-B1. When a Table 21-B row does not list the member explicitly, this sentence still requires it.

**Table 21-B1. Retry-action registry**

| `retry_action` | Required use |
| --- | --- |
| `correct_request` | Schema, mapping, parser, filter, sort, limit, selector, target, or semantic validation failures. |
| `refresh_resource` | Table-version conflict, non-active resource, source changed, or graph/query stale. |
| `restart_query` | Invalid, expired, or authorization-stale cursor. |
| `reduce_scope_or_limits` | Table, graph, counter, or request resource limit exceeded. |
| `retry_with_backoff` | Only `projection_timeout` or `projection_unavailable` when Core classifies the cause transient. |
| `do_not_retry` | Forbidden operation, hidden/authorization failure, ID collision exhaustion, or a non-transient projection contract failure. |

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

For Network Flow-owned mutating routes, exact committed idempotency replay lookup occurs at the Table 5-B replay point and returns the original success before Table 21-C resource/lifecycle, semantic, or limit failures for the current resource state. A same-tuple different-digest `client_txn_conflict` is reported at that same point.

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
Conformance fixtures MUST include Table 22-A for this NLSpec to remain adopted/current. Draft successor revisions MAY contain explicit `TODO:` fixture rows. Omission behavior: a fixture row with any `TODO:` value is a known adoption blocker and MUST NOT satisfy adoption or implementation conformance.

**Table 22-A. Fixture bundle registry**

| Fixture ID | Path | SHA-256 | Source profile | Parser profile | Required expected outputs |
| --- | --- | --- | --- | --- | --- |
| `NF-FIX-001-cisco-sna-minimal` | `fixtures/network-flow/NF-FIX-001-cisco-sna-minimal/manifest.json` | `b87f8eec4754ee852ef6d6ba6fcf26622b7eeb1f05c2a9e3771fe32191c4431c` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Table resource, mapping fingerprint, exact row IDs, zero diagnostics, graph result. |
| `NF-FIX-002-cisco-sna-interface-fields` | `fixtures/network-flow/NF-FIX-002-cisco-sna-interface-fields/manifest.json` | `01370193ba2cd0ba60ff5927733c8a6a21cde992baaf31cb679d519e900643c1` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Interface field mapping outputs, exact row IDs, graph result. |
| `NF-FIX-003-duplicate-headers` | `fixtures/network-flow/NF-FIX-003-duplicate-headers/manifest.json` | `d0ecd247b193ef6a889c89eb8de61d89aa30b16b208f1e8e8ce8ae42080206b5` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Source-column descriptors proving ordinal disambiguation. |
| `NF-FIX-004-rejected-rows` | `fixtures/network-flow/NF-FIX-004-rejected-rows/manifest.json` | `fb0ee5d1ac8b5ec72d35b28c0919a274f553b0cf7cdc7bca84d12ac250f1d859` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Invalid IP, port, protocol, timestamp, counter, field-count, and end-before-start diagnostics in exact order. |
| `NF-FIX-005-csv-parser-edges` | `fixtures/network-flow/NF-FIX-005-csv-parser-edges/manifest.json` | `f04665b06b128a986973b85a5817ab13b0516f0e7c36f2eb5a9e753b37d22dc5` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Terminal newline, blank line, quoted newline, quote escaping, and malformed quote outcomes. |
| `NF-FIX-006-cross-table-graph` | `fixtures/network-flow/NF-FIX-006-cross-table-graph/manifest.json` | `f75cfd9ec9d058fb0fa531da56c93c55ee289a3a200de764d031fa0fb36410c9` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Two table resources, shared endpoint vertex, aggregated edge, exact graph query digest, exact source snapshot ID, and exact edge IDs. |
| `NF-FIX-007-indicator-linking` | `fixtures/network-flow/NF-FIX-007-indicator-linking/manifest.json` | `3817ca7bb750e20bd36377c0c18dd63896f61ec95738f38c56f3f7d595b336e1` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Existing-indicator link, create-indicator link, duplicate binding result. |
| `NF-FIX-008-large-limits` | `fixtures/network-flow/NF-FIX-008-large-limits/manifest.json` | `796c5984554e79259acef7505699ee4f0588bac681ebacded5422e830eb53298` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Graph/table limit failures; engineering measurement only, not Core 05 publication evidence. |
| `NF-FIX-009-soft-delete-stale-graph` | `fixtures/network-flow/NF-FIX-009-soft-delete-stale-graph/manifest.json` | `a7e1c9db21c90b9e535d642d0b9ee1c52e882a873ca578b6c00a442c32877860` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Active graph query containing a table that is then soft-deleted; stale graph and cursor invalidation. |
| `NF-FIX-010-json-admission` | `fixtures/network-flow/NF-FIX-010-json-admission/manifest.json` | `8edc935cd2c5bc976aa8489a12ab9ee1f1573e9cb851c1b5cedc61e16dfc6c53` | n/a | n/a | Duplicate member, explicit null, unknown member, malformed JSON, non-object body failures. |
| `NF-FIX-011-alias-collision` | `fixtures/network-flow/NF-FIX-011-alias-collision/manifest.json` | `f9100a1047d6df34d78e04128c67cafc86f8f2d7252e043e1b24a307a2b008d2` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Alias match keys, duplicate-alias warning, explicit approved mapping, and mapping conflict for source reuse. |
| `NF-FIX-012-sys-uptime-timestamps` | `fixtures/network-flow/NF-FIX-012-sys-uptime-timestamps/manifest.json` | `fc69425962a5c04156fc65693ba5cbc72afe0e8aeef141bb9cf5a6750b843389` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Export time, exporter uptime-at-export, start/end event uptime derivations, and wrap-ambiguous rejection. |
| `NF-FIX-013-filename-display` | `fixtures/network-flow/NF-FIX-013-filename-display/manifest.json` | `3cb396782f502dc573cbaef01a8d20cc2dc47f08d0a8e5cbc6c07cc11df741ac` | n/a | n/a | Path stripping, hidden-file stem, trailing-dot stem, display-name override rejection, duplicate suffixing, and soft-delete name reuse. |
| `NF-FIX-014-cursor-pagination` | `fixtures/network-flow/NF-FIX-014-cursor-pagination/manifest.json` | `89aed2d8fd40ab4a944095386425d99b6676e17a693f767c4888df019857eb3c` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Mandatory default sort tail, keyset continuation tuple, terminal null cursor, actor mismatch, and table rename cursor survival. |
| `NF-FIX-015-graph-adapter-input` | `fixtures/network-flow/NF-FIX-015-graph-adapter-input/manifest.json` | `39c9c5c1a10b5163ef269d84afaa8dc80b173e592793e0f3ac6cd8abd87dd9e7` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Exact `network_flow_graph_projection_input_v1`, graph view key, source snapshot ID, property definitions, and safe metadata. |
| `NF-FIX-016-redaction` | `fixtures/network-flow/NF-FIX-016-redaction/manifest.json` | `04c84b78bb1a5822dbdc1fbabf55b4a619584478096d0c99425234de4255a064` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Deterministic `safe_sample` nulls, integer-like samples, raw value SHA-256s, audit safe digests, and no raw text leakage. |
| `NF-FIX-017-indicator-link-mismatch` | `fixtures/network-flow/NF-FIX-017-indicator-link-mismatch/manifest.json` | `b351f48f5fb23cdb886cf7e19d565a96c0e34a59e4f7f677d92fae578318f2fe` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Non-IP field rejection, edge field disambiguation, `confirm_exact_value` mismatch, existing-indicator normalized-value mismatch, and create-indicator Core dependency failure. |
| `NF-FIX-018-resource-limits` | `fixtures/network-flow/NF-FIX-018-resource-limits/manifest.json` | `1ed98e03e4a60afe97675cf68d59d75035ebdb06bef358d55827d5eb8e86dd63` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Parser/import resource limit failures, diagnostic truncation, graph limit failure, counter digit limit failure, and invalid deployment config cases. |
| `NF-FIX-019-canonical-json-unicode` | `fixtures/network-flow/NF-FIX-019-canonical-json-unicode/manifest.json` | `fd274df8eb5c6ec3530429c20a8aaf774b6b0d257a7c326aca7ce2dcc0373428` | n/a | n/a | Exact JSON escapes, Unicode scalar ordering, Unicode 17 whitespace, NFC, null/present digest framing, and unpaired-surrogate rejection. |
| `NF-FIX-020-atomic-import-commit` | `fixtures/network-flow/NF-FIX-020-atomic-import-commit/manifest.json` | `2d57eb1b14293c2a800af760f0918145ef7b9c919546c97b371081c49c795b59` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Failure injection at every final-commit step proves no ghost table, partial row, diagnostic, binding, or domain audit. |
| `NF-FIX-021-preview-boundaries` | `fixtures/network-flow/NF-FIX-021-preview-boundaries/manifest.json` | `65d0c177472782d735fb85338289439ed140d369f1869f6839e1db95fe343b43` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | First-record header semantics, invalid controls, exact 50-record preview stop, post-preview malformed data, blank/mismatch row counting, and row limit `limit+1`. |
| `NF-FIX-022-timestamp-rulesets` | `fixtures/network-flow/NF-FIX-022-timestamp-rulesets/manifest.json` | `6c05ce8005f057657e6c62ed0d9b4b0fbcf52c12fa3f986651e37b32f91b7c32` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Closed variants, exact RFC3339 grammar, `tzdb-2026c` fold/gap transitions, epoch canonical integers, uptime 32-bit bounds, and distinct ordinals. |
| `NF-FIX-023-import-facade-source-change` | `fixtures/network-flow/NF-FIX-023-import-facade-source-change/manifest.json` | `ce665c2d8574f8ec610c8b0869adbcba3de6f7051a8b98281ba97a83f513cdb1` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Exact preview/apply requests and results, server-derived descriptors, no path leakage, and all `network_flow_source_changed` reasons. |
| `NF-FIX-024-query-normalization-cursors` | `fixtures/network-flow/NF-FIX-024-query-normalization-cursors/manifest.json` | `d40ad09966abf40feb575142ca7672c77b5e3445e676f460a371d6481f27d851` | n/a | n/a | Closed table scopes, normalized duplicate filters, initial/continuation separation, 4096-byte bound, expiry boundary, and independent row/diagnostic cursor tuples. |
| `NF-FIX-025-graph-contributors` | `fixtures/network-flow/NF-FIX-025-graph-contributors/manifest.json` | `22bd2071d57ec54ae6e759e24764bce8ad329d876cd80962b971a7e82e6d0c6a` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Exact graph response objects, vertex/edge contributor pages, current-authorization recomputation, stale digest, and no rejected contributors. |
| `NF-FIX-026-audit-and-replay` | `fixtures/network-flow/NF-FIX-026-audit-and-replay/manifest.json` | `11e3a0066d1567564a9f2e316a5836f75d336571db49a7781cdbc537e356a252` | n/a | n/a | Created-versus-reused binding events, digest key IDs, exact replay with no second domain audit, graph-success-only audit, and exact truncated-ref count. |
| `NF-FIX-027-retention-soft-delete` | `fixtures/network-flow/NF-FIX-027-retention-soft-delete/manifest.json` | `3319cd131bcc817c9d0a9e295c29b4173c5130823786c1e5df9abcdd203e348b` | n/a | n/a | Soft-delete terminal behavior, incident-closure retention, non-queryability, retained counts, cursor invalidation, Core-governed audit retention, and no v1 whole-incident purge claim. |
| `NF-FIX-028-graph-aggregate-bounds` | `fixtures/network-flow/NF-FIX-028-graph-aggregate-bounds/manifest.json` | `5acf103d56b2e3e1db0711abdfb354b7e29f14f43dd220ccbebaa4a2070facc1` | `cisco_sna_netflow_csv_v1` | `rfc4180_headered_csv_v1` | Arbitrary-precision sums, exact digit limit, fixed vertex/edge/counter failure order, and no partial adapter output. |

**NF-REQ-178**
Each fixture bundle MUST include canonical expected-output transcripts for route success `data` objects, route error payloads, table resources, source-column descriptors, approved mapping JSON, mapping fingerprint, row IDs, row digests, diagnostics, graph result, Graph Projection adapter input, indicator-link result, resource-limit details, and redaction outputs where applicable. Fixture graph digests and source snapshot IDs MUST be independent of deployment limit configuration. Fixture-only safe digest expectations MUST declare deterministic fixture `network_flow_safe_digest_key_material`; production deployments MUST NOT use that fixture secret.

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
| `NF-AC-012` | Empty CSV, forbidden header controls, header-only CSV, terminal newline, blank line, malformed quote, quote escaping, and field-count mismatch match §9.2 outcomes; the first logical record is always the header. |
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
| `NF-AC-057` | Request members `time_range.bucket` and `aggregation.include_time_buckets` fail as unknown members in v1; no time-bucket error code is exposed. |
| `NF-AC-058` | `observation_mode` values other than `binding_only` fail, and `created_observation_refs[]` is always `[]` in binding resources and link responses. |
| `NF-AC-059` | Graph query digests and source snapshot IDs remain unchanged when deployment graph limits or caller `limit_overrides` change without changing query semantics. |
| `NF-AC-060` | Timestamp profile precision rejects finer source resolution, epoch modes reject fractional values, and sys-uptime parsing uses export time, exporter uptime-at-export, and per-field event uptime exactly as §9.7 specifies. |
| `NF-AC-061` | Duplicate table rename fails with `network_flow_invalid_display_name` and `reason_code='duplicate_display_name'`, while existing cursors over the renamed table continue. |
| `NF-AC-062` | Only `active` and `soft_deleted` are table lifecycle states; import staging is not a table, direct soft-deleted references follow NF-REQ-058, and Table 8-C limit accounting is exact. |
| `NF-AC-063` | `network_flow_all_rows_rejected` includes ordered `diagnostics_sample[]`, `row_count_rejected`, and `diagnostics_truncated` details without creating a table. |
| `NF-AC-064` | Omitted query limits materialize to `min(200, effective max_query_limit)`, explicit query limits outside range fail, and graph limit overrides outside `[lowerable_min, effective]` fail. |
| `NF-AC-065` | Validation preview counts are limited to the parser preview slice, and an apply may fail with `network_flow_all_rows_rejected` even when preview accepted rows. |
| `NF-AC-066` | Binding `source_row_refs[]`, `source_row_refs_truncated`, and `source_row_refs_total_count` are populated per Table 15-F, and duplicate `row_refs[]` selector entries fail. |
| `NF-AC-067` | Indicator-link dedupe uses the resolved binding identity tuple; structurally different selectors resolving to the same tuple dedupe, and different source-row-ref sets create distinct bindings. |
| `NF-AC-068` | `confirm_exact_value` mismatch fails before mutation with `network_flow_indicator_link_ambiguous` and the required indicator-link error details. |
| `NF-AC-069` | Graph Projection adapter input uses Network Flow endpoint IDs as source entity IDs, flow edge IDs as source relationship IDs, and no invalid `ephemeral_response_only` retention token. |
| `NF-AC-070` | Interface fields store bounded text or null only; numeric interface identifiers remain text and sort by code point. |
| `NF-AC-071` | Unknown aggregation modes fail request admission, no unavailable-mode error is exposed, and mapping combinability accepts only `single_source_only` in v1. |
| `NF-AC-072` | Table rename, table delete, indicator-link create, and Core import apply replay follow the idempotency comparison and replay points in Table 5-B. |
| `NF-AC-073` | Mapping approval accepts only the three §10.4 variants; `derived`, `constant`, fake ignored-field sentinels, and variant-extra members fail deterministically. |
| `NF-AC-074` | Alias suggestions use `source_alias_match_key_v1`; duplicate alias matches produce blocking visible warnings until explicit user approval resolves every matched column. |
| `NF-AC-075` | Every Network Flow success response `data` object contains the schema, required members, ordering, nullable behavior, and `meta` shape defined in §14.6 or §17. |
| `NF-AC-076` | Every Table 20-A resource limit uses the configured default/minimum behavior, enforcement phase, error/truncation behavior, and invalid-config rejection specified in §20. |
| `NF-AC-077` | Safe samples and source-column samples follow Table 12-F exactly, including null samples for raw text/IP-like values and numeric-only samples for bounded integer-like values. |
| `NF-AC-078` | Route-local errors include Table 21-B details and choose the first error under Table 21-C and Table 21-D ordering. |
| `NF-AC-079` | Indicator linking accepts only v1 IP endpoint candidates, rejects every non-linkable field in Table 15-A1, and validates existing/create targets against Core canonical IP indicator identity. |
| `NF-AC-080` | Source filename display, default table display names, explicit display-name overrides, duplicate suffixing, hidden-file stems, trailing-dot stems, and soft-delete name reuse match §8.4. |
| `NF-AC-081` | Final import commit is atomic under failure injection and never exposes `creating`, `failed`, a ghost table ID, partial rows, partial diagnostics, or a domain audit without its table. |
| `NF-AC-082` | Explicit import display names fail on duplicates without suffixing, while omitted display names use deterministic suffix allocation under the same incident-scoped commit lock. |
| `NF-AC-083` | CSV preview consumes exactly the first 50 complete logical data records, includes blank and mismatched records in preview counts, and does not report an error located only after that boundary. |
| `NF-AC-084` | `max_rows_per_csv` counts logical data records including blank and mismatched rows, excludes the header, and reports `actual=limit+1` without scanning to a total. |
| `NF-AC-085` | Cisco SNA v1 permits only nine required targets plus two optional interface targets; exporter, TCP flags, and application label mappings fail as unsupported and public rows contain nulls for them. |
| `NF-AC-086` | `trim_ascii_space_v1` trims only U+0020 and the empty-value policy is applied after transformation exactly as NF-REQ-095 specifies. |
| `NF-AC-087` | Timestamp objects reject cross-variant members, enforce exact RFC3339 and epoch grammar, bind non-UTC mappings to `tzdb-2026c`, and enforce unsigned 32-bit uptime plus distinct ordinals. |
| `NF-AC-088` | Import facade preview has no durable side effect; apply receives no filesystem path or URL; descriptors, source digest, defaults, and mapping fingerprint are server-derived; changed source fails closed. |
| `NF-AC-089` | Every source column has one contiguous descriptor and one disposition; every required/system target has exactly one valid mapping; observation provenance is materialized server-side once. |
| `NF-AC-090` | Public row objects always contain every nullable optional field, exact `unmapped_raw` value objects, and the namespaced `network_flow.observation_source_ref`. |
| `NF-AC-091` | Diagnostic discovery under parallel validation yields byte-identical ordered resources, stable reason/message keys, and exact conditional limit fields. |
| `NF-AC-092` | Table scope variants reject cross-variant members, duplicate selected IDs, over-limit selection, and zero-table resolved scopes without revealing hidden resources. |
| `NF-AC-093` | Filter normalization canonicalizes values and `in` order, rejects post-canonicalization duplicates, and binds the normalized array rather than caller spelling or order. |
| `NF-AC-094` | Initial and continuation request variants are mutually exclusive; cursor tokens reject non-ASCII and more than 4096 bytes; a token is invalid at `now == expires_at`. |
| `NF-AC-095` | Accepted-row, diagnostic, and graph-contributor cursors use their specified full independent keyset tuples and never skip or duplicate equal-prefix items. |
| `NF-AC-096` | Counter aggregates use arbitrary precision, accept values beyond uint64 when within the digit limit, and fail vertices, edges, bytes digits, then packet digits in the specified order without partial output. |
| `NF-AC-097` | Graph Projection metadata contains exact ordered table IDs and mapping fingerprints, fixed metadata mappings, and maps adapter outcomes to Table 14-G6 without internal leakage. |
| `NF-AC-098` | Graph success data uses exact `semantic_query`, source-table-ref, result-limit, and annotation schemas with no vague or implementation-defined members. |
| `NF-AC-099` | Contributor queries recompute current composition and authorization, paginate all contributing accepted rows deterministically, and fail stale or missing vertex/edge selectors without falling back to example refs. |
| `NF-AC-100` | Indicator selectors and targets are closed variants, confirmation requires byte-exact canonical IP text, and create-indicator clients cannot supply normalized or display values. |
| `NF-AC-101` | Core indicator create/dedupe and binding insert/reuse commit atomically; `duplicate` appears only in the link result; new and reused bindings return HTTP 201 and 200 respectively. |
| `NF-AC-102` | Every safe digest resource or audit field carries its key ID, and key rotation preserves comparison only within equal key IDs. |
| `NF-AC-103` | Audit occurrences match Table 16-C exactly, exact replay emits no new domain occurrence, and graph truncated-ref count equals the specified sum. |
| `NF-AC-104` | Soft delete and incident closure retain Network Flow state but make it non-queryable as specified in Table 8-D, invalidate affected cursors, preserve Core-governed audit retention, and expose no v1 whole-incident purge claim. |
| `NF-AC-105` | Every route returns its exact success status, exact closed data schema, Table 21-A status, exhaustive reason code, safe details, and retry action. |
| `NF-AC-106` | Every document dependency has an adopted version and immutable locator, every blocker in §24 is closed, and every Table 22-A fixture has concrete immutable bytes before adopted/current status is claimed. |
| `NF-AC-107` | Import cancellation before commit leaves no table, while cancellation or worker failure after commit recovers and publishes the one committed success without duplicate table creation. |
| `NF-AC-108` | The owner manifest and fragment resolve exactly to document version `2.0.0`, contract major `2`, Import major `1`, the reserved route/workspace, empty capabilities, and no competing discovery fact. |
| `NF-AC-109` | State presence uses only the four authoritative logical families, permits metadata-only empty state under `empty_state_policy='allowed'`, and never treats jobs, ledgers, caches, projections, or staged objects as state. |
| `NF-AC-110` | Fresh initialization is empty and invokes only the final validator; current state version `1` requires no profile migration definition, while an omitted, extra, or code-inferred migration fails contract generation. |
| `NF-AC-111` | Inactive Network Flow configuration rejects `key_ring_manifest_path` structurally without defaulting, retaining, resolving, reading, invoking profile code, or performing DNS, connection, or other egress. |
| `NF-AC-112` | Every authoritative family has one required PostgreSQL backup binding and digest-bound codec; restore is stopped-empty, group-ordered, sequential, validated before advance, and invokes no inactive profile code. |
| `NF-AC-113` | Import apply, indicator link, invalidation, and backup/restore use only their declared typed contributions; profile-owned job, worker, portability, reporting, and persisted rebuild declarations are exactly empty. |
| `NF-AC-114` | A standard client renders `network_analysis` only for major `2` at the current authorized availability generation; all capability facts and nonempty capability arrays fail with `extension_capability_not_supported`. |

## 24. Core amendments and adoption blocker checklist

**NF-REQ-180**
This NLSpec may remain `adopted/current` only while the adoption checklist in Table 24-A is closed.

**Table 24-A. Adoption blocker checklist**

| Blocker ID | Required closure |
| --- | --- |
| `NF-BLOCK-001` | Core 00 recognizes `network_flow_activity` as an adopted extension profile. |
| `NF-BLOCK-002` | Core 01 generic discovery always reserves the route family and `network_analysis` workspace for this recognized profile, emits major `2` and `capabilities=[]`, and varies only `claimed`; no profile-local discovery item or compatibility reader exists. |
| `NF-BLOCK-003` | Core 01 import terminal result references admit `kind='network_flow_table'`. |
| `NF-BLOCK-004` | Core 03 admits extension-contributed top-level incident tabs without expanding base built-in tabs. |
| `NF-BLOCK-005` | Core 04 adds Network Flow route-family authorization/conformance hooks. |
| `NF-BLOCK-006` | Every fixture row in §22 has concrete path, byte hash, and expected output transcript. |
| `NF-BLOCK-007` | The generated contract artifacts derived from this NLSpec exist and pass drift checks. |
| `NF-BLOCK-008` | Route, parser, digest, graph, indicator-link, security, and limit acceptance criteria have executable tests. |
| `NF-BLOCK-009` | Core 02 or the adopted Core indicator registry designates the exact IP-literal indicator type token required by §15. |
| `NF-BLOCK-010` | Every normative dependency in Table 1-B has an adopted version and immutable repository locator. |
| `NF-BLOCK-011` | Core 01 adopts the two-operation import owner-facade boundary, opaque stream capability, source-change check, and atomic final-commit/result publication contract. |
| `NF-BLOCK-012` | Core 02 adopts the exact IP canonicalization contract, explicit no-private-purge boundary, and indicator create/dedupe participation in the binding unit of work. |
| `NF-BLOCK-013` | Core 03 adopts extension invalidation topics and consequences for rename, soft delete, and authorization loss. |
| `NF-BLOCK-014` | Core 04 adopts cursor confidentiality/integrity/key rotation, safe-digest key-ID handling, audit occurrence semantics, and retention boundaries referenced here. |
| `NF-BLOCK-015` | The adopted Graph Projection contract accepts the exact ephemeral adapter input, property/metadata mappings, arbitrary-precision counter strings, and outcome mappings in §14.4. |
| `NF-BLOCK-016` | The adopted Testing Harness contract can execute immutable fixture manifests, failure injection, fake clock, authorization transitions, and audit-count assertions required by §23. |
| `NF-BLOCK-017` | `tzdb-2026c` fixture data and transition expectations are vendored or immutably identified for timestamp conformance, with the required provenance record validated by `make json-shape-check`. |

## 25. Extensions owner declarations

**NF-REQ-181**
The primary owner document identity is
`cartulary.network_flow_activity.current.v2`, version `2.0.0`. Its only runtime
dependency is `profile_id='import'`, `required_contract_major=1`, bound to the
exact Import owner manifest version and digest selected by the Extensions
dependency declaration set. The recognized profile is claimable at contract
major `2`; it declares exactly route family
`/api/v1/incidents/{incident_id}/network-flow`, workspace key
`network_analysis`, and no capability. The claim key is
`network_flow_activity.claimed`. Those facts come only from the Core 00 owner
fragment and this primary owner fragment; source code, route registration,
configuration, tests, and prose search are forbidden fact sources.

**NF-REQ-182**
The `cartulary.extension_state_presence_manifest.v1` owner declaration for this
profile uses `migration_lineage_id='network_flow_activity.state_v1'`, current
state version `1`, minimum migratable version `1`, and
`empty_state_policy='allowed'`. Its authoritative logical families are exactly,
in ascending UTF-8 order:

1. `network_flow_activity.indicator_bindings`;
2. `network_flow_activity.rejected_row_diagnostics`;
3. `network_flow_activity.rows`;
4. `network_flow_activity.tables`.

State is present if and only if at least one member exists in one of those four
families. Generic extension metadata, the migration ledger, Core Import jobs and
resource references, caches, ephemeral graph projections, staged objects,
temporary files, indexes, and configuration never make Network Flow state
present. Metadata with no authoritative member is valid empty state because the
policy is `allowed`; it is not synthetic state.

The `cartulary.extension_state_initialization_definition.v1` declaration has
`kind='empty'`. It invokes no Network Flow code, constructs no authoritative
member, and then invokes exactly
`network_flow_activity.validate_state_v1` through its digest-bound packaged
algorithm reference. There are no `cartulary.extension_migration_definition.v1`
objects while current and minimum versions are both `1`; an omitted migration
path is valid only for the `1 -> 1` case, and an extra migration is invalid.

**NF-REQ-183**
The `cartulary.extension_profile_configuration_contract.v1` declaration contains
exactly `network_flow_activity.claimed` and
`network_flow_activity.key_ring_manifest_path`. The claim row is Core 04 owned.
The path row has `inactive_policy='forbidden'` and
`inactive_value_schema_ref=null`; `syntax_only` is not selected and there is no
inactive-value schema for this profile. Its active `value_schema_ref` is the
owner schema `cartulary.network_flow_key_ring_manifest_path.v1`, the namespace
schema is `cartulary.network_flow_activity.configuration_namespace.v1`, omission
is `required`, resolution is `regular_file_ref`, and diagnostics are
`name_only`. While unclaimed, presence of the path
fails structural admission with `profile_incompatible_binding` before any
default, configuration view, value retention, reference or secret resolution,
file access, DNS lookup, connection, egress, or Network Flow invocation. While
claimed, §20.1 is the sole value and secret-handling contract.

**NF-REQ-184**
The `cartulary.extension_physical_state_binding.v1` declaration contains exactly
four authoritative PostgreSQL bindings, one for each NF-REQ-182 family. Every
binding uses `backup_inclusion='required'`, a digest-bound
`cartulary.network_flow_activity.postgres_rows.v1` codec, the shared PostgreSQL
post-restore structural validator, and `rebuild_algorithm_id=null`. Restore
order groups are `100` for `tables`, `200` for `rows`, `300` for
`rejected_row_diagnostics`, and `400` for `indicator_bindings`; bindings within
a group execute sequentially by binding ID. Historical codec declarations and
derived physical bindings are empty in version `2.0.0`. Backup/restore operates
on a stopped empty target, validates each binding before advancing, never invokes
Network Flow code while inactive, and never serves a failed target. There is no
persisted derived state to rebuild after claim; the graph remains an ephemeral
query result.

**NF-REQ-185**
The profile declares exactly these integration contributions:

- one `http_route_family` for the NF-REQ-181 route;
- one `incident_workspace` for `network_analysis`;
- one `import_target` for `network_flow_table`;
- resource kinds `network_flow_table` and `network_flow_indicator_binding`;
- one `websocket_invalidation` contribution covering those resource kinds;
- typed cross-owner transaction participants
  `network_flow_activity.import_apply_v1` and
  `network_flow_activity.indicator_link_v1`;
- one typed backup/restore participant
  `network_flow_activity.backup_restore_v1`.

Every participant reference and digest is owner-authored and must match the
shared-owner registry. The profile declares no portability participant because
`incident_portability_mode='blocked_when_present'` uses the NF-REQ-182 state
predicate. It declares `snapshot_reporting_mode='no_participation'`, no
profile-owned job kind, no profile-owned worker kind, no derived-state rebuild
algorithm, and no egress. Import scheduling remains Core Import behavior and is
not reclassified as a Network Flow job. Empty declarations are normative; the
generator must reject inferred entries from packages, SQL, or runtime behavior.

**NF-REQ-186**
The profile admission-validation declaration has no preflight algorithm, exactly
`network_flow_activity.validate_state_v1` as the post-migration algorithm, and
`dependency_probes=[]`. Schema-validation conditions are supplied only by
annotated owner schemas. Procedural conditions are supplied by the closed
Network Flow validation decision tables in this document. Every reachable
invalid condition must appear in the generated validation-condition registry;
an undeclared condition or an emitted unregistered condition blocks admission.
Validation results use the shared precedence: invocation failure, structural
invalidity, overflow at `4097+` findings, remaining schema defects, valid
findings, then valid empty result; `257..4096` is the ordinary 256-item bound
violation.

**NF-REQ-187**
The only supported browser build class is `standard`. Its generated support row
must name profile `network_flow_activity`, contract major `2`, workspace
`network_analysis`, and an empty capability set. Browser eligibility is the
intersection of generic discovery, that exact packaged support row, current
authorization/availability, and the current local epoch/generation. A stale,
unauthorized, unsupported-major, missing-support, or capability-bearing profile
must not render. All nonempty capability facts and arrays are invalid, and every
attempted activation returns `extension_capability_not_supported`. Base workbook
state and stable client/WebSocket identities survive every Network Flow
availability transition.

## Appendix E. Future-only decision backlog and rationale

This appendix is non-normative. It records deferred work and rationale for readers of this NLSpec. It does not add v1 implementation-conformance behavior.

**Table E-A. Future-only backlog**

| Backlog item | Future owner dependency | Current v1 handling |
| --- | --- | --- |
| Core 02 observation-source amendment | Core 02 would need to admit extension-sourced `indicator_observation` rows through a closed `origin_kind` extension token and a non-record extension source reference. | Network Flow v1 supports `observation_mode='binding_only'` only and always returns `created_observation_refs[]=[]`. |
| Binding list/read routes | A later Network Flow revision would need route inventory, authorization, pagination, filters, response schemas, and audit behavior for binding reads. | Bindings are observable only through link-route responses and created/reused audit events. |
| Time-bucketed graph output | A later Network Flow revision would need UTC-epoch bucket alignment, bucket-count formula, bucketed response schema, limits, errors, fixtures, and a new graph-query digest version. | Bucket request members are unknown members and fail admission in v1. |

**Table E-B. Rationale notes**

| Decision | Rationale |
| --- | --- |
| Cursor binding excludes table versions | Flow rows are immutable after activation. Table rename changes metadata only, so binding cursor validity to table versions would invalidate live query cursors without changing row population. |
| Graph query digest excludes `limit_overrides` | Limit overrides lower failure thresholds but do not change query semantics. Excluding them keeps graph identity and fixture digests deployment-independent. |
| Validation preview is preview-slice-scoped | The mapping modal must remain responsive for large CSVs. Full-file validation belongs to import apply, which already owns the background validation pass. |
| Import staging is not a table lifecycle state | Prevents addressable ghost tables and makes active/retained limit accounting depend only on committed domain resources. |
| Contributor retrieval is a dedicated route | Example refs are bounded evidence samples, not a complete pivot interface. Recomputing the graph selector under current authorization avoids stale local joins. |
| Aggregate counters are arbitrary precision | Each source counter is uint64, but a sum of many valid rows can exceed uint64. A digit bound controls resources without corrupting valid arithmetic. |

## Sources

This section records non-normative source evidence used to shape this NLSpec. It does not add requirements. Sources include Core 00 through Core 04, `graph_projection_nlspec.md`, `nlspec-spec.md`, research reports `R01` through `R09` under `docs/research`, RFC 4180, RFC 7011, RFC 5952, RFC 8785, RFC 9844, IANA IPFIX registry background, Cisco SNA NetFlow/IPFIX guidance, and CSV-injection security background. The timestamp and Unicode baselines were checked against the [IANA Time Zone Database](https://www.iana.org/time-zones), whose current release on 2026-07-09 was 2026c, and the [Unicode 17.0.0 specification](https://www.unicode.org/versions/Unicode17.0.0/). Internet and external-source material remains supporting evidence only unless this NLSpec restates the behavior as a Network Flow requirement.
