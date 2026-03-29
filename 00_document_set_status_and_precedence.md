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

**REQ-00-001**
The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are normative.
Profiles: base
Verified by: AC-231

**REQ-00-002**
- **MUST / MUST NOT** indicates a conformance requirement.
- **SHOULD / SHOULD NOT** indicates a strong default whose exceptions must remain compatible with the rest of this core.
- **MAY** indicates an optional behavior whose omission semantics are explicit.
Profiles: base
Verified by: AC-231

## 4. Conformance model

### 4.1 Base profile

**REQ-00-003**
An implementation that claims **Cartulary Base Profile** conformance MUST satisfy all requirements in:

- Core 00,
- Core 01,
- Core 02,
- Core 03,
- the Base Profile criteria in Core 04.
Profiles: base
Verified by: AC-231

The Base Profile covers the workbook-first incident workspace, record model, mention resolution, evidence attachment, collaboration, revision history, rollback, local authentication, deployment baseline, and the built-in/system views defined by this core.

### 4.2 Extension profiles

The source artifact mixes current-state requirements with roadmap language in several areas. To preserve all source information without forcing contradictory scope into a single profile, this core defines optional extension profiles.

An implementation MAY additionally claim any of the following extension profiles:

- **Import Extension Profile** for file-based structured import beyond clipboard paste, including bounded CSV and XLSX onboarding.
- **Snapshot and Reporting Extension Profile** for immutable incident snapshots and self-contained report or presentation outputs.
- **Incident Portability Extension Profile** for full-fidelity administrative whole-incident export/import between trusted Cartulary deployments.
- **Reference Pack Extension Profile** for reference-pack activation, refresh, and overlay behavior.
- **Enterprise Authentication Extension Profile** for OIDC and SAML provider integration.

**REQ-00-004**
If an implementation claims an extension profile, it MUST satisfy the matching profile-specific requirements and acceptance criteria in Core 01 through Core 04.
Profiles: import, snapshot_reporting, incident_portability, reference_pack, enterprise_authentication
Verified by: AC-232, AC-233, AC-234, AC-235, AC-236

### 4.3 Unsupported future areas

The source artifact mentions several future areas without defining enough detail for current conformance, including restricted evidence visibility beyond the incident-scoped workspace model, promotion of `hypothesis` to a first-class record type, generalized workflow engines beyond the bounded analyst-work coordination and lifecycle model defined in this core, duplicate-resolution suggestions, cross-incident analytics, and presentation automation beyond the bounded snapshot and reporting controls defined in this core. These areas are reserved for future specification work and are non-conformant claims unless later NLSpecs define them.

## 5. Document map

- **Core 01** defines authoritative system topology, storage boundaries, view contracts, public success/error envelope registries, projections, portability, failure handling, and background-job behavior.
- **Core 02** defines authoritative record types, identifiers, mention/entity semantics, canonical closed vocabularies, provenance, deduplication, merge rules, schema invariants, and history mechanics.
- **Core 03** defines authoritative workbook interactions, collaboration semantics, workflow contracts, auto-resolution policy, grouping behavior, and write-back rules.
- **Core 04** defines authoritative security, deployment, trust boundaries, and acceptance criteria.

### 5.1 Contract-owner matrix

**REQ-00-005**
A contract family that appears in more than one normative core document MUST have one primary owner section. Non-owner sections MAY restate the owner only to declare local consequences, UI affordances, storage realization, or conformance checks. When a non-owner restatement and the owner differ, the owner governs and the restatement is editorial drift that MUST be repaired.
Profiles: base
Verified by: AC-231

| Contract family | Primary owner | Allowed secondary sections | Ownership rule | Requirement ID | Profiles | Verified by |
| --- | --- | --- | --- | --- | --- | --- |
| Public success/error envelope and public error-code and reason-code registries | Core 01 §3.3.6, §3.3.6.1, and §3.3.6.2 | Core 03 §3.3.4; Core 04 §9.6, §9.9, and §9.10 | Secondary sections MAY require a specific code or payload member but MUST NOT assign a conflicting meaning, transport status, or retry hint. | REQ-00-006 | base | AC-231 |
| Background-job resource shell, cancel semantics, retention semantics, and reusable `job_progress` payload members | Core 01 §3.3.9 and §3.3.9.1 | Core 01 §3.3.10.1; Core 03 §4.3.1 and §4.4; Core 04 §2 and §9.10; Appendix E and Appendix F | Core 01 owns the canonical HTTP job resource, cancel route semantics, post-terminal retention contract, and the shared `scope`, `status`, `progress`, `cancelable`, `result_summary`, `error_summary`, and `retained_until` members reused by `job_progress`. Core 03 owns only local client behavior under replay, resync, and auth churn. Core 04 owns only authorization and conformance criteria. |  |  |  |
| Session resource shape and expiry fields | Core 01 §3.3.2.1 | Core 03 §4.4; Core 04 §1.1.1 and §9.10 | Core 01 owns the authenticated-session resource fields returned to clients. |  |  |  |
| Session issuance, expiry, revocation, and concurrent-session behavior | Core 04 §1.1.1 | Core 01 §3.3.2.1 and §3.3.10.1; Core 03 §4.4 | Non-owner sections MAY reference `session_expires_at`, `session_revoked`, and retry behavior but MUST NOT widen allowed lifetime or revocation semantics. | REQ-00-007 | base | AC-231 |
| Saved-view route, mutability, and authorization contract | Core 01 §3.3.5.2 | Core 02 §11.1; Core 03 §2.3-§2.4; Core 04 §9.10 | Core 02 owns persistence fields only. Core 03 owns workbook discoverability and startup interaction only. |  |  |  |
| Workbook startup-preference objects | Core 02 §11.2 | Core 01 §3.3.5.2; Core 03 §2.4; Core 04 §9.10 | `home_sheet_ref` and `default_sheet_ref` semantics MUST remain separate and MUST NOT be collapsed into saved-view flags. | REQ-00-008 | base | AC-231 |
| Incident resource shape, incident-create contract, and incident-metadata mutability | Core 01 §3.3.5.3 and §3.3.5.3.1 | Core 02 §4.5 and §14.1; Core 04 §9.10; Appendix C and Appendix E | Core 01 owns the public incident resource fields, `POST /api/v1/incidents` request and response contract, server-managed initial values, and the create-only versus patchable field boundary. Core 02 owns persistence minima only. |  |  |  |
| Same-field conflict transport and `collection_review` resolver payloads | Core 03 §3.3.4 | Core 01 §3.3.5 and §3.3.6; Core 04 §9.6 and §9.10 | Core 01 owns the common envelope. Core 03 owns the conflict object, resolver semantics, and `collection_value_v1` conflict payload rules. |  |  |  |
| Domain closed-vocabulary registry | Core 02 §18 | Core 01 view contracts; Core 03 workflow surfaces; Core 04 conformance criteria | Non-owner sections MAY require a subset only when they reference the exact tokens owned by Core 02. |  |  |  |
| Lifecycle-machine states and legal transitions for `task_request` and `decision` | Core 02 §10.4.1.1 and §10.4.2.1 | Core 01 §3.3.6; Core 03 §6 and §16.4; Core 04 §9.9 | Core 02 owns state sets and legal transitions. Core 01 owns the common illegal-transition transport shape. Core 04 owns pass/fail verification. |  |  |  |
| Base-profile `view_schema` registry and per-schema field registries | Core 01 §7.4 | Core 03 §14 and §16.1; Core 04 §9.1; Appendix E and Appendix F | Core 01 owns the exact base-profile `view_schema_id` set and the exhaustive per-field contracts for each schema, including stable `field_key`, default sort, filter whitelist, write target or action, `conflict_resolution_class`, and `entity_binding_mode` where applicable. Secondary sections MAY describe local interaction, conformance, or roadmap consequences only. |  |  |  |
| Evidence blob attachment and evidence-access handle issuance and redemption | Core 01 §3.3.8 for `POST /api/v1/evidence-records/{record_id}/attach-blob`, and Core 01 §16 for handle issuance and redemption | Core 02 §13 and §18; Core 03 §8.1 and §8.4; Core 04 §2, §4.3, §4.5, and §9.10; Appendix F | Core 01 §3.3.8 owns `POST /api/v1/evidence-records/{record_id}/attach-blob`, including request shape, record-scoped idempotency, normalized replay comparison, optimistic-concurrency behavior, and route-specific error use. Core 01 §16 owns preview and download issuance plus redeem semantics, handle lifetime, revocation, filename and disposition rules, and route-specific error use. Core 02 owns lifecycle state, object metadata, and exact token sets only. Core 03 owns workbook-surface invocation and blocked-state UI only. Core 04 owns authorization re-derivation, active-content blocking, and fail-closed behavior only. |  |  |  |
| Extension route-family public contracts for imports, reference packs, snapshots and releases, and incident bundles | Core 01 §17 | Core 03 §11.2; Core 04 §9.0, §9.2, §9.3, §9.4, and §9.11; Appendix E and Appendix F | Core 01 owns exact route inventory, request and response defaults, omitted-versus-`null` behavior, route-scoped idempotency, reuse of the common job resource, family-specific error registries, and durable terminal-state representation. Core 03 owns only underlying workflow and resource semantics. Core 04 owns only claim manifests and conformance checks. |  |  |  |


### 5.2 Requirement and acceptance-criterion traceability

**REQ-00-009**
Every atomic conformance-critical requirement block in Core 00 through Core 04 MUST carry one stable requirement identifier of the form `REQ-<core>-<nnn>`, where `<core>` is `00` through `04` and `<nnn>` is a zero-padded document-local sequence number assigned in document order. Once assigned, a requirement identifier MUST NOT be reassigned. If a requirement is removed, its identifier is retired.
Profiles: base
Verified by: AC-231, AC-237

**REQ-00-010**
Each requirement block MUST carry:

- `Profiles:` with one or more canonical profile identifiers,
- `Verified by:` with one or more acceptance-criterion identifiers from Core 04.
Profiles: base
Verified by: AC-231, AC-237

The canonical profile identifiers are:

- `base`,
- `import`,
- `snapshot_reporting`,
- `incident_portability`,
- `reference_pack`,
- `enterprise_authentication`.

**REQ-00-011**
Each acceptance criterion in Core 04 MUST carry a `Verifies:` back-reference. Ordinary criteria MUST list explicit `REQ-*` identifiers. A profile-manifest criterion MAY use one canonical `profile:*` selector inline instead of enumerating every selected requirement identifier.
Profiles: base
Verified by: AC-231, AC-237

**REQ-00-012**
Appendix F MUST expand every `profile:*` selector into explicit `REQ -> owner section / profiles / ACs`, `AC -> REQs`, and `Profile -> required REQs + ACs` navigation tables. Those tables MUST sort by canonical identifier order and MAY compress contiguous identifier runs with `..` range notation.
Profiles: base
Verified by: AC-231, AC-237

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
- immutable snapshots and derived outputs when the Snapshot and Reporting Extension Profile is implemented,
- whole-incident export/import bundles when the Incident Portability Extension Profile is implemented.

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

**REQ-00-013**
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
Profiles: base, reference_pack
Verified by: AC-043, AC-044, AC-045, AC-046, AC-231, AC-234

## 8. Canonical terms and identifiers

### 8.1 Core identifiers

**REQ-00-014**
- **`incident_id`**: stable identifier of the incident workspace boundary.
- **`record_id`**: stable identifier of a user-visible record envelope or other record-bearing object defined by this core.
- **`row_version`**: monotonically increasing version for a mutable record or mention row used by optimistic concurrency control.
- **`change_set`**: immutable attribution and transaction grouping unit for one committed user or system action.
- **mutation entry**: one reversible change target recorded within a parent `change_set`.
- **`history_entry_ref`**: stable opaque identifier emitted by record history for one row-centric logical history item that maps to exactly one reversible mutation target. Clients MUST NOT parse or synthesize it.
- **`view_schema_id`**: stable identifier of a built-in sheet or contract-backed system view.
- **`saved_view_id`**: stable identifier of a saved-view configuration scoped to an incident or system-owned seed.
- **`field_key`**: stable identifier of a contract-declared field in a view contract, import mapping, or API contract, or of a synthetic filter predicate declared by a view contract.
- **`client_txn_id`**: client-generated identifier used to correlate a batch of mutations with the user action that produced them.
- **`entity_mention_id`**: stable identifier of an entity-mention row captured from source text before or after explicit resolution.
- **`object_blob_id`**: stable identifier of a stored binary blob slot or authoritative object-metadata row.
- **`conflict_token`**: opaque server-issued token that binds one same-field conflict payload to one current saved-field version and one explicit resolution attempt.
- **`job_id`**: stable identifier of a background job exposed through the public API or live-update stream.
- **`cursor_token`**: opaque pagination token bound to one versioned list or view-query contract.
Profiles: base
Verified by: AC-116, AC-118, AC-123, AC-124, AC-125, AC-127, AC-128, AC-129, AC-231

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

**REQ-00-015**
Cartulary implementations MUST satisfy all of the following invariants:

1. The workbook metaphor lives at the view layer. Source data MUST remain disciplined relational state rather than independent sheet silos.
2. Behavior MUST follow explicit contracts and stable identifiers rather than visible tab names, column headers, or UI labels.
3. Rough capture and later normalization MUST preserve original analyst-entered text and provenance.
4. Every mutation MUST be attributable to an authenticated actor or an explicitly identified system process.
5. Projection rows, exports, overlays, and reports MUST derive from canonical source state or an explicitly versioned snapshot of canonical derivation state.
6. The client MUST address mutable rows by `record_id` and `row_version`, never by visible row position or displayed values.
7. Incident data and optional reference packs MUST version independently.
8. History MUST be authoritative at `change_set` plus mutation-entry granularity rather than row-snapshot granularity alone.
9. The public client/server wire contract MUST be versioned and keyed by stable identifiers rather than visible labels, displayed row positions, or storage-specific table names.
10. Optional overlays and enrichment MUST NOT block the primary capture path.
11. If the implementation cannot stay within one interaction of spreadsheet-style row creation and editing for the primary capture flow, it fails the design objective preserved from the source artifact.
Profiles: base, snapshot_reporting, reference_pack
Verified by: AC-231, AC-233, AC-234

### 9.1 Lifecycle state-machine notation

Cartulary MAY define an explicit lifecycle state machine only for a conformance-critical flow whose legal transitions, failure handling, or recovery semantics are not already fully determined by simpler field constraints.

Explanatory workflow sequences, lifecycle arrows, and illustrative diagrams in appendices are non-normative unless the normative core restates them as an explicit machine contract.

**REQ-00-016**
Any normative lifecycle machine MUST declare, at minimum:

- the closed set of machine states or machine conditions,
- the allowed events or transition triggers,
- required guards or preconditions for each transition,
- the authoritative persisted representation, including which structured fields or records determine the current state,
- the observable signals exposed to analysts, APIs, jobs, or logs,
- CI-verifiable conformance checks for happy path, terminal failure paths, illegal transitions, idempotent retry after simulated crash, and deterministic rerun from the same starting state.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-231

**REQ-00-017**
This notation is a specification pattern. It MUST NOT be read as a requirement to adopt a runtime finite-state-machine library or framework.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-231

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
