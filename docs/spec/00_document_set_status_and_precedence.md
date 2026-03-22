# Cartulary Normative Core 00: Document Set Status, Precedence, and Conformance

## 1. Status

This document set is the **provisional normative core** for Cartulary.

It is derived from the exploratory design artifact preserved in Appendix G. The normative core is authoritative for all closed decisions expressed here. Open design questions, roadmap items, rationale, illustrative UI mockups, explanatory diagrams, and source extracts are non-normative unless restated in the normative core as explicit requirements.

This provisional status exists because the source artifact still carried unresolved product questions. Those unresolved items are preserved in Appendix E and do not weaken requirements that are already closed in this core.

## 2. Precedence

The order of authority is:

1. future Cartulary NLSpecs derived from this core, once adopted,
2. this normative core,
3. non-normative appendices,
4. the exploratory source artifact.

If two normative core documents appear to conflict, the more specific contract governs. If specificity is equal, the later-numbered normative core document governs until the contradiction is repaired.

When the normative core and an appendix differ, the normative core governs.

## 3. Normative language

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are normative.

- **MUST / MUST NOT** indicates a conformance requirement.
- **SHOULD / SHOULD NOT** indicates a strong default whose exceptions must remain compatible with the rest of this core.
- **MAY** indicates an optional behavior whose omission semantics are explicit.

## 4. Conformance model

### 4.1 Base profile

An implementation that claims **Cartulary Base Profile** conformance MUST satisfy all requirements in:

- Core 00,
- Core 01,
- Core 02,
- Core 03,
- the Base Profile criteria in Core 04.

The Base Profile covers the workbook-first incident workspace, record model, mention resolution, evidence attachment, collaboration, revision history, rollback, local authentication, deployment baseline, and the built-in/system views defined by this core.

### 4.2 Extension profiles

The source artifact mixes current-state requirements with roadmap language in several areas. To preserve all source information without forcing contradictory scope into a single profile, this core defines optional extension profiles.

An implementation MAY additionally claim any of the following extension profiles:

- **Import Extension Profile** for file-based structured import beyond clipboard paste, including bounded CSV and XLSX onboarding.
- **Snapshot and Reporting Extension Profile** for immutable incident snapshots and self-contained report or presentation outputs.
- **Reference Pack Extension Profile** for reference-pack activation, refresh, and overlay behavior.
- **Enterprise Authentication Extension Profile** for OIDC and SAML provider integration.

If an implementation claims an extension profile, it MUST satisfy the matching profile-specific requirements and acceptance criteria in Core 01 through Core 04.

### 4.3 Unsupported future areas

The source artifact mentions several future areas without defining enough detail for current conformance, including restricted evidence visibility beyond the incident-scoped workspace model, promotion of `hypothesis` to a first-class record type, generalized workflow engines beyond the bounded analyst-work coordination model defined in this core, duplicate-resolution suggestions, cross-incident analytics, and presentation automation beyond the bounded snapshot and reporting controls defined in this core. These areas are reserved for future specification work and are non-conformant claims unless later NLSpecs define them.

## 5. Document map

- **Core 01** defines authoritative system topology, storage boundaries, view contracts, projections, portability, failure handling, and background-job behavior.
- **Core 02** defines authoritative record types, identifiers, mention/entity semantics, provenance, deduplication, merge rules, schema invariants, and history mechanics.
- **Core 03** defines authoritative workbook interactions, collaboration semantics, workflow contracts, auto-resolution policy, grouping behavior, and write-back rules.
- **Core 04** defines authoritative security, deployment, trust boundaries, and acceptance criteria.

## 6. System boundary

Cartulary is an incident-scoped, workbook-first investigation system for multi-user incident response work.

Cartulary is within scope for:

- incident workspaces,
- timeline capture,
- host and identity normalization,
- notes and artifacts,
- task requests, decisions, and structured coordination artifacts,
- evidence envelopes and blob-backed evidence,
- typed record relationships,
- tags,
- revision history and rollback,
- workbook views and saved views,
- optional reference-pack overlays,
- immutable snapshots and derived outputs when the Snapshot and Reporting Extension Profile is implemented.

The following are out of scope for current conformance unless a later normative document adds them:

- fully offline browser sync,
- multi-master replication,
- formulas, macros, merged cells, and spreadsheet computation engines,
- manual row-range grouping,
- unrestricted record-level ACL design,
- automatic entity merge based on fuzzy similarity alone,
- free-form user-defined grouping expressions,
- arbitrary field-picker rollback from historical snapshots.

## 7. Supported operating envelope

The implementation MUST support the following operating assumptions:

- a single incident workspace typically has 2 to 8 active users and MAY reach 25 active users,
- a serious incident MAY accumulate 1,000 to 20,000 timeline rows,
- a serious incident MAY accumulate hundreds to low thousands of host and identity records,
- a serious incident MAY accumulate tens of GB of evidence,
- initial deployments are single-tenant,
- “disconnected” means isolated deployment operation and MUST NOT be interpreted as a requirement for offline browser sync or multi-master replication,
- the default client is a browser UI,
- rough capture and progressive normalization MUST preserve original user-entered text,
- large binary evidence MUST NOT be stored inline in Postgres,
- optional reference packs MAY be present or absent by deployment and the core workbook MUST remain usable without them,
- report and presentation artifacts, when generated, MUST be renderable without remote runtime assets.

## 8. Canonical terms and identifiers

### 8.1 Core identifiers

- **`incident_id`**: stable identifier of the incident workspace boundary.
- **`record_id`**: stable identifier of a user-visible record envelope or other record-bearing object defined by this core.
- **`row_version`**: monotonically increasing version for a mutable record or mention row used by optimistic concurrency control.
- **`change_set`**: immutable attribution and transaction grouping unit for one committed user or system action.
- **mutation entry**: one reversible change target recorded within a parent `change_set`.
- **`view_schema_id`**: stable identifier of a built-in sheet or contract-backed system view.
- **`field_key`**: stable identifier of a write-back-capable field in a view contract, import mapping, or API contract.
- **`client_txn_id`**: client-generated identifier used to correlate a batch of mutations with the user action that produced them.

### 8.2 Domain terms

- **record envelope**: the common record identity, version, attribution, and delete-state wrapper shared across first-class records.
- **primary record**: a user-visible record that owns a `record_id` and participates directly in linking, revisions, and projection materialization.
- **projection row**: a denormalized row materialized for a workbook sheet or system view. Projection rows are derived state.
- **saved view**: a user- or system-scoped workbook tab configuration over a projection or `view_schema_id`.
- **reference pack**: a separately versioned vocabulary, framework, registry, or enrichment dataset that is not incident data.
- **entity mention**: an observed textual reference captured inside another record before or without canonical normalization.
- **stub entity**: a host or identity record with stable identity but incomplete or unverified canonical detail.
- **canonical entity**: a host or identity record that has been normalized to the point required by the implementation’s entity workflow.
- **active entity**: a host or identity record that is neither soft-deleted nor merged away.
- **system view**: a contract-backed workbook surface whose semantics come from `view_schema_id`, not from visible labels.

### 8.3 Binding-mode terms

- **`mention_origin`**: a field contract in which typed or imported text creates `entity_mentions` rather than host or identity records.
- **`entity_origin`**: a field contract in which typed or imported text creates or updates a host or identity record directly.

## 9. Global invariants

Cartulary implementations MUST satisfy all of the following invariants:

1. The workbook metaphor lives at the view layer. Source data MUST remain disciplined relational state rather than independent sheet silos.
2. Behavior MUST follow explicit contracts and stable identifiers rather than visible tab names, column headers, or UI labels.
3. Rough capture and later normalization MUST preserve original analyst-entered text and provenance.
4. Every mutation MUST be attributable to an authenticated actor or an explicitly identified system process.
5. Projection rows, exports, overlays, and reports MUST derive from canonical source state or an explicitly versioned snapshot of canonical derivation state.
6. The client MUST address mutable rows by `record_id` and `row_version`, never by visible row position or displayed values.
7. Incident data and optional reference packs MUST version independently.
8. History MUST be authoritative at `change_set` plus mutation-entry granularity rather than row-snapshot granularity alone.
9. Optional overlays and enrichment MUST NOT block the primary capture path.
10. If the implementation cannot stay within one interaction of spreadsheet-style row creation and editing for the primary capture flow, it fails the design objective preserved from the source artifact.

## 10. Source-preservation note

The exploratory artifact contained several mixed-status areas, especially around:

- the boundary between clipboard paste and file-based structured import,
- immutable snapshot and report generation,
- reference-pack refresh and distribution.

This core resolves those areas as follows:

- current closed behavior is expressed normatively here,
- optional extension profiles capture source behaviors that were described as important but not uniformly day-one,
- unresolved evolution questions remain in Appendix E,
- the original artifact remains preserved in Appendix G.
