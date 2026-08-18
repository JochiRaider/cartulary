# Cartulary Normative Core 00: Document Set Status, Precedence, and Conformance

## 1. Status

This document set is the authoritative normative core for the current Cartulary
profile.

Core 00 through Core 04 are the implementation-conformance corpus for the
current profile. Core 05 is a separate normative companion for claim
publication and benchmark reproducibility only. It governs claim-bearing
publication requirements for timed or fixture-sensitive criteria and is not
part of Base Profile or extension-profile implementation conformance.
Appendices remain non-normative.

`docs/design.md` is the sole normative design-direction owner for the current
profile. The UI/UX design guide is non-normative design support: it may restate
owner decisions and explain their rationale or implementation consequences,
but it does not own product behavior, defaults, interfaces, acceptance claims,
or release evidence. Runtime code, generators, test routing, conformance
checks, and release evidence MUST NOT consume either guide text or other
Markdown as a behavioral input; executable design values are projected through
the machine contracts identified by `docs/design.md`.

Versioned schemas, enums, limits, mappings, algorithms, fixtures, generated
artifacts, implementation, tests, and retained evidence are downstream of the
applicable normative owner. A difference between an adopted specification and
a derived machine artifact is a defect in that projection or implementation;
executable tools do not resolve the defect by parsing the specification text.

It is derived from the exploratory design artifact preserved in Appendix G. For the current profile, the base profile and the currently defined extension-profile boundaries are closed here. The normative core is authoritative for current-profile requirements, closed design decisions, and current conformance claims. Roadmap items, rationale, illustrative UI mockups, explanatory diagrams, source extracts, and historical source-question material are non-normative unless restated in the normative core as explicit requirements.

This revision records behavior-affecting closure for deployment administration entry, imported-incident initial access, Reference Pack list search and filters, prohibited aggregate administration concepts, the Timeline operational-field refactor, view-schema-owned workbook inspector configuration, exhaustive inspector feature-group workflow routing, single-click committed-cell editing with bounded fill-down behavior, the deployment-local Collaboration stream-quarantine requeue contract, Recovery's complete authored public-base-table catalog, and typed-only Object Store initialization failure classification. With those contracts adopted in Core 00 through Core 04, current-profile behavior, profile boundaries, and implementation-conformance scope are closed again here.

Contract-owner coverage for repeated families is complete in Core 00 §5.1. Appendix E is roadmap, historical source-question material, and future-only editorial backlog; it is not a live source of unresolved current-profile contract decisions. Any later typo, formatting, link, or similar non-substantive correction is corpus maintenance only and does not reopen current-profile design status, conformance scope, or profile boundaries.

The remaining hardening work for this corpus is editorial normalization only after this behavior-affecting closure. It does not reopen current-profile runtime behavior, profile boundaries, or implementation-conformance scope.

Current-profile workbook-surface identity and registry closure, including pack-dependent surface constraints, are owned by Core 01 §7.4 and Core 03 §2. Appendices may describe future candidates but do not define current-profile workbook surfaces.

**REQ-00-051**
Non-normative supporting-guidance artifacts MAY describe recommended operator practice for tracker hygiene, companion findings-document discipline, handoff quality, status-review cadence, workload redistribution, debrief discipline, challenge/escalation practice, and inspector-supported row-context cleanup. Such guidance MUST NOT define implementation-conformance requirements, required row-creation cadence, required per-edit ritual, mandatory row-level approval, required inspector workflow cadence, required authorization behavior, or required runtime workflow unless the behavior is restated normatively in Core 00 through Core 04.
Profiles: base
Verified by: AC-231

**REQ-00-062**
Projection behavior is governed by Core 00 through Core 04 and by adopted ADR, SPEC, or NLSpec artifacts only when those artifacts are explicitly adopted for their named scope. A projections-specific subsystem NLSpec is authoritative only when it is explicitly marked `Adopted` and listed as adopted by the project document taxonomy. `docs/graph_projection_nlspec.md` is `status: adopted/current` and authoritative for the graph-projection subsystem only. Draft, research, exploratory, or evidence documents, including research reports R01 through R09, are informative unless promoted through the adopted-document process.

When a projections-specific NLSpec is adopted or substantively revised, projection-related Core sections, implementation trackers, provider descriptors, rebuild behavior, query behavior, and boundary guard tests MUST be re-audited before accepting new projection changes. Graph Projection NLSpec 2.1.0 has completed that owner-level re-audit against Core 01's workbook projection provider and restore adapter contracts, Core 04's generation, serving-lease, journal, and reinitialization rules, the Graph implementation tracker, Network Flow producer, Reporting consumer, Recovery contribution, and boundary guards. GP3 machine projections and behavioral evidence remain downstream work and cannot be inferred from this owner adoption.

Graph Projection NLSpec 2.1.0 owns the distinct catalog-resolved Graph restore participant and pure deterministic graph derivation. Network Flow owns saved graph declarations and public routes; Reporting owns exact-result release consumption. Graph Projection does not redefine workbook-grid projection tables, `view_row_v1`, workbook query routes, Core saved views, import owner facades, the workbook `RestoreProjectionRebuilder`, workbook provider descriptors, public HTTP routes, WebSocket messages, or deployment-configuration keys. The phrase "workbook restore rebuild" names the Core 01 provider path; the phrase "Graph restore participant" names current `graphprojection.restore_rebuild.v3` plus the exact read-only v2 dispatcher only for supported retained pre-GP3 backup catalogs. Neither is an alias for the other.
Profiles: base
Verified by: AC-469

## 2. Precedence

The order of authority is:

1. future Cartulary NLSpecs derived from this core, once adopted,
2. this normative core,
3. non-normative appendices,
4. the exploratory source artifact.

If two normative core documents appear to conflict, the apparent contradiction is a corpus defect to be repaired. Overlapping contract families resolve through the primary owner identified in §5.1 rather than through later document numbering.

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

Implementation conformance and claim publication are separate. An implementation MAY conform to the Base Profile and any claimed extension profiles without making any public timed or fixture-sensitive claim. Claim-bearing benchmark publication is governed by Core 05.

### 4.2 Extension profiles

The source artifact mixes current-state requirements with roadmap language in several areas. To preserve all source information without forcing contradictory scope into a single profile, this core defines optional extension profiles.

An implementation MAY additionally claim any of the following extension profiles:

- **Import Extension Profile** for file-based structured import beyond clipboard paste, including bounded CSV and XLSX onboarding.
- **Snapshot and Reporting Extension Profile** for immutable incident snapshots, report-composition authoring inputs, and self-contained report or presentation outputs.
- **Incident Portability Extension Profile** for full-fidelity administrative whole-incident export/import between trusted Cartulary deployments.
- **Reference Pack Extension Profile** for reference-pack activation, refresh, and overlay behavior.
- **Enterprise Authentication Extension Profile** for OIDC and SAML provider integration.
- **Network Flow Activity Extension Profile** for incident-scoped Network Analysis tables, flow-row query, graph composition, and explicit indicator-link initiation.

**REQ-00-064**
The **Network Flow Activity Extension Profile**, identified by the stable
`profile_id='network_flow_activity'`, is a claimable current extension profile
only through `docs/network-flow-activity-nlspec.md` while that document is
marked `status: adopted/current` by the repository document-status process. A
deployment that does not claim the profile MUST continue to enumerate the
reserved identity as unclaimed and MUST NOT expose Network Flow routes or a
Network Analysis workspace. A deployment that claims the profile MUST satisfy
the Network Flow Activity NLSpec, the owner interfaces imported from Core 01
through Core 04, the adopted Graph Projection NLSpec, and the adopted Testing
Harness evidence boundary for that extension. This profile does not add Base
Profile behavior, Base Profile workbook surfaces, Core record families, Core
saved views, or any whole-incident purge claim.
Profiles: base
Verified by: AC-231

**REQ-00-065**
For the adopted Extensions companion manifest, this owner document has `owner_document_schema_id='cartulary.core00.current.v1'` and `owner_document_version='extensions-adoption-1'`.

With `docs/extension-subsystem-nlspec.md` and every companion named by its coordinated-adoption gates adopted/current together, Core 00 recognizes exactly the six current profile identities listed below and is the sole authority for their claimability, current contract major, primary owner, and runtime dependencies. Generated descriptors, packaged code, route presence, deployment configuration, tests, and prose MUST NOT add, remove, recognize, retire, or reclassify a profile.

| `profile_id` | `claimable` | Current contract major | Primary owner | Exact runtime dependencies |
| --- | --- | ---: | --- | --- |
| `enterprise_authentication` | `true` | `1` | Enterprise Authentication owner sections in Core 01/Core 04 | `[]` |
| `import` | `true` | `1` | Import owner sections in Core 01 | `[]` |
| `incident_portability` | `true` | `1` | Incident Portability owner sections in Core 01 | `[]` |
| `network_flow_activity` | `true` | `4` | `docs/network-flow-activity-nlspec.md` version `4.0.0` | exactly `import@1` |
| `reference_pack` | `true` | `1` | Reference Pack owner sections in Core 01 | `[]` |
| `snapshot_reporting` | `true` | `1` | adopted Reporting and Report Composition owner documents | `[]` |

Each `recognized_profile` owner fact MUST carry a non-null `primary_owner_contract_ref` that resolves to the exact adopted primary owner above. Every dependency declaration MUST bind the exact owner document version, document digest, owner-manifest identity, and owner-manifest digest. Capability facts are prohibited in extension contract major `1`; every required capability array is present and empty, and attempted activation fails with `extension_capability_not_supported`. A structurally valid nonempty capability array is an activation attempt even when it contains an unknown future string; callers MUST NOT receive a token-specific classification or have any supplied token echoed. Structural array or member-type failures remain request or manifest validation failures. This requirement is current through the atomic Extensions companion adoption; no pre-adoption recognition or capability contract remains current.
Profiles: base
Verified by: AC-231, EXT-AC-145, EXT-AC-146, EXT-AC-157

**REQ-00-066**
Core 01 is the primary owner for generic Import Extension Profile orchestration: upload and source
integrity, sessions and units, mapping approval storage, target selection, the generated target
registry, common authorization, idempotency, cancellation, durable unit outcomes, session/job
finalization, and the internal analytical-facade binding shape. Each adopted analytical target
NLSpec is the exclusive owner of the exact mapping, preview, apply, result, diagnostic, error, and
resource-mutation schemas referenced by its binding. Core MUST NOT duplicate those exact target
member lists, and a target owner MUST NOT redefine Core import lifecycle or job behavior.
View-schema import targets remain owned by their source-owner create facades.
Profiles: import, network_flow_activity
Verified by: AC-264A, AC-464, AC-465, NF-AC-106

**REQ-00-004**
If an implementation claims an extension profile, it MUST satisfy the matching profile-specific requirements and acceptance criteria in Core 01 through Core 04.
Profiles: import, snapshot_reporting, incident_portability, reference_pack, enterprise_authentication, network_flow_activity
Verified by: AC-232, AC-233, AC-234, AC-235, AC-236, NF-AC-106

For the Snapshot and Reporting Extension Profile, Core 00 adopts `docs/reporting-subsystem-nlspec.md` as the Reporting render/export authority and `docs/report-composition-nlspec.md` as the report-composition authoring authority when those documents are marked adopted/current by the repository document-status process. If either required NLSpec is absent, not adopted, or marked blocked for a claimed behavior, the affected external-release or report-composition conformance claim MUST fail closed; implementations MAY expose internal experimental behavior only when public validation reports the missing owner dependency explicitly.

### 4.3 Unsupported future areas

The source artifact mentions several future areas without defining enough detail for current conformance, including restricted evidence visibility beyond the incident-scoped workspace model, promotion of `hypothesis` to a first-class record type, generalized workflow engines beyond the bounded analyst-work coordination and lifecycle model defined in this core, duplicate-resolution suggestions, cross-incident analytics, and presentation automation beyond the bounded snapshot and reporting controls defined in this core. These areas are reserved for future specification work and are non-conformant claims unless later NLSpecs define them.

This reservation also includes local-account WebAuthn or passkey support, including registration, assertion, credential enumeration, credential revocation or reset, and recovery semantics. Such support is non-conformant in the current profile unless later NLSpecs define it.

Incident archive, incident hard deletion, incident soft deletion, incident purge, and any equivalent incident-removal lifecycle are also future-only areas. The current profile defines only `active` and `closed` incident lifecycle states plus close and reopen actions; removal, retention, tombstone, or purge semantics for whole incidents MUST NOT be claimed unless a later NLSpec defines them.

The Network Flow Activity boundary does not narrow this reservation. Its
claimable v1 revision defines terminal soft delete for its own analytical
tables, but it MUST NOT claim whole-incident removal or install a private
incident-purge lifecycle. A future generic incident-removal profile may define a
cascade participant for Network Flow after that owner boundary exists.

## 5. Document map

- **Core 01** defines authoritative system topology, storage boundaries, view contracts, public success/error envelope registries, projections, portability, failure handling, and background-job behavior.
- **Core 02** defines authoritative record types, identifiers, mention/entity semantics, canonical closed vocabularies, provenance, deduplication, merge rules, schema invariants, and history mechanics.
- **Core 03** defines authoritative workbook interactions, collaboration semantics, workflow contracts, auto-resolution policy, grouping behavior, and write-back rules.
- **Core 04** defines authoritative security, deployment, trust boundaries, and acceptance criteria.
- **Core 05** defines authoritative claim publication and benchmark reproducibility for public timed or fixture-sensitive claims. It is normative companion material and is not part of Base Profile or extension-profile implementation conformance.

Adopted subsystem NLSpecs derived from this core MAY define bounded implementation-conformance requirements for their named subsystem only. Each adopted subsystem NLSpec MUST state its scope, non-goals, owner interactions, and whether it adds deployment-configuration keys. An adopted telemetry NLSpec owns telemetry generation, telemetry configuration, telemetry export, resource identity, privacy, and telemetry verification only. An adopted graph-projection NLSpec owns graph-oriented projection input, output, lifecycle, validation, identity, and consumer-query behavior only. Adopted subsystem NLSpecs do not redefine Core 00 through Core 04 product behavior or Core 05 claim-publication behavior unless a later Core revision explicitly says so.

`docs/network-flow-activity-nlspec.md` is `status: adopted/current` and owns
only its analytical table, immutable flow row, graph-adapter, indicator-binding,
and Network Analysis workspace semantics inside the owner interfaces provided
by Core 01 through Core 04. It must not promote flow tables or rows into Core
records, `view_schema` resources, saved views, or Base Profile workbook
surfaces.

`docs/domain.md` is a first-class domain vocabulary and concept-reference document for repository terminology and owner-section navigation. It does not replace Core 00 through Core 05, appendices, or owner sections, and it does not add implementation-conformance behavior. If `docs/domain.md` and an owner section differ, the owner section governs and the difference is documentation drift.

### 5.1 Contract-owner matrix

**REQ-00-005**
A contract family that appears in more than one normative core document MUST have one primary owner section. Non-owner sections MAY restate the owner only to declare local consequences, UI affordances, storage realization, or conformance checks. Owner-owned create-time behavior includes, at minimum, minimum semantic create signal, zero-field-create eligibility, create-time writeability, omitted-member defaults, create-time omitted-versus-`null` behavior, and server-managed initial values. Core 04 acceptance criteria MAY repeat owner-held create-time behavior only to verify it. Outside the owner section, and outside such Core 04 verification text, non-owner sections MUST NOT define or restate those owner-owned create-time behaviors as independent normative behavior. When a non-owner restatement and the owner differ, the owner governs and the restatement is editorial drift that MUST be repaired.
Profiles: base
Verified by: AC-231

**REQ-00-067**
The authoritative current record envelope and retained record history are
distinct ownership concerns. Core 02 §3 owns record-envelope meaning and the
closed current-profile record-type registry. Core 01 owns the internal
current-envelope boundary, physical-relation ownership, transaction
direction, and Incident Portability row contract. Core 02 owns history
meaning; Core 01 owns public history, delete, restore, and rollback routes; and
the internal Revisions concern owns change-set, mutation-entry,
record-revision, rollback, and destructive-operation coordination.

An implementation MUST NOT infer that retained-history ownership transfers
authority over current envelope state. A current-envelope mutation coordinated
by Revisions or another source owner MUST use the current-envelope owner's
transaction port, while history append, projection refresh, collaboration
publication, authorization, and transport mapping remain with their existing
owners.
Profiles: base, incident_portability
Verified by: AC-509, AC-510, AC-511, AC-512, AC-514

**REQ-00-068**
The deployment-local Collaboration stream-quarantine requeue operation is one
cross-document contract family with three distinct primary owners. Core 03
owns the semantic transition and its atomic state effects. Core 01 owns the
logical CLI grammar, typed result envelope, error/reason registry, and exit
mapping. Core 04 owns local invocation authority, configuration-path trust,
redaction, cancellation and timeout safety, private journal constraints, and
negative public-surface requirements. None of those owners transfers semantic
authority to the Operator application facade, the Collaboration contract
projection, the Testing Harness, or an implementation guide.
Profiles: base
Verified by: AC-535

**REQ-00-069**
Database migration lifecycle ownership is distinct from PostgreSQL connection
adapter ownership and from the source owners whose schemas are evolved. Core 01
§2.1A owns migration source identity, typed execution, lineage readiness,
remediation, evidence, and recovery metadata. Core 04 owns the secret boundary
and conformance proof. Source-owner specifications retain schema meaning;
OpenTelemetry retains PostgreSQL signal and privacy requirements; and the
Testing Harness retains execution/evidence accounting. Package location, SQL
location, test location, and verification routing MUST NOT transfer those
responsibilities.
Profiles: base
Verified by: AC-537

**REQ-00-070**
`docs/decisions/projections-module-boundary.md` is an adopted implementation
architecture decision for the workbook-grid Projections module only. It owns
the exact Go package topology, constructor boundary, internal transition order,
and removal of repository-internal compatibility surfaces named in that
decision. It MUST NOT define or change public workbook query behavior,
`view_row_v1`, saved-view behavior, cursor semantics, authorization,
Collaboration publication, restore result vocabulary, projection descriptor
wire shape, database schema, or source-owner semantics.

Core 01 remains the behavioral owner for workbook-grid projections, provider
descriptors, query behavior, physical projection storage, and restore rebuilds.
Core 04 remains the conformance owner. The adopted Graph Projection NLSpec
remains authoritative only for the graph-projection subsystem. The adopted
implementation decision MUST be revised or withdrawn when it conflicts with a
later adopted behavioral owner; implementation or tracker text MUST NOT be used
to settle such a conflict.
Profiles: base
Verified by: AC-539

**REQ-00-071**
`docs/decisions/revisions-module-boundary.md` is an adopted implementation
architecture decision for Revisions only. It owns the exact internal
composition boundary, provider/catalog topology, transition order, and removal
of repository-internal compatibility paths named in that decision. It MUST NOT
redefine public history, delete, restore, rollback, conflict, WebSocket, or
Incident Bundle behavior; source-owned current state; history and snapshot
meaning; or security and conformance.

Core 01 remains authoritative for application, route, portability, and storage
boundaries. Core 02 remains authoritative for canonical snapshots, mutation
targets, history association, selector meaning, and rollback semantics. Core
03 remains authoritative for Collaboration consequences. Core 04 remains
authoritative for security and conformance. The decision MUST be revised or
withdrawn when it conflicts with a later adopted behavioral owner; an
implementation, contract projection, test, tracker, or generated artifact MUST
NOT settle such a conflict.
Profiles: base, incident_portability
Verified by: AC-529

| Contract family | Primary owner | Allowed secondary sections | Ownership rule | Requirement ID | Profiles | Verified by |
| --- | --- | --- | --- | --- | --- | --- |
| Revisions implementation topology, source-provider composition, and repository-internal compatibility removal | `docs/decisions/revisions-module-boundary.md` for implementation structure; Core 01 and Core 02 for behavior | Core 03 Collaboration consequences; Core 04 security/conformance; `docs/domain.md` vocabulary; implementation trackers | The adopted decision owns internal package and constructor structure only. It cannot redefine source state, retained-history meaning, public contracts, portability, or conformance. | REQ-00-071 | base, incident_portability | AC-529 |
| Workbook-grid Projections implementation topology and repository-internal compatibility removal | `docs/decisions/projections-module-boundary.md` for implementation structure; Core 01 §8 and §12.2 for behavior | Core 04 §9.1A; Appendix I; implementation guides and trackers | The adopted decision owns exact package and constructor structure only. It cannot redefine projection behavior, storage meaning, source semantics, public contracts, or conformance. | REQ-00-070 | base | AC-539 |
| Current record-envelope authority versus retained record history | Core 02 §3 for envelope meaning and record-type membership; Core 01 §1 and §12.3 for implementation ownership and portability | Core 01 record mutation/history routes; Core 02 history substrate; Core 04 conformance; `docs/domain.md` for vocabulary only | The current-envelope owner controls current envelope persistence and transaction ports. Revisions controls history and destructive coordination but MUST NOT become current-envelope authority. | REQ-00-067 | base, incident_portability | AC-509..AC-512, AC-514 |
| Collaboration stream-quarantine requeue transition | Core 03 §4.3.3 | Core 01 §12.2.2 transport; Core 04 §2 security and conformance; `docs/domain.md` vocabulary; implementation guide | Core 03 alone owns admission, repaired-state proof, locking, preserved/reset fields, atomic journal participation, concurrency, typed semantic outcomes, and forbidden semantic effects. | REQ-00-068 | base | AC-535 |
| Collaboration requeue deployment-local CLI grammar, result envelope, and error/exit registry | Core 01 §12.2.2 | Core 03 §4.3.3 transition; Core 04 §2 security and conformance; typed projections under `contracts/collaboration` | Core 01 alone owns accepted tokens and values, config/timeout flags, schema/member/order rules, closed code/reason pairs, stream behavior, and exit mapping. The typed contract family is a projection, not an owner. | REQ-00-068 | base | AC-535 |
| Collaboration requeue local authority, redaction, private journal, timeout/cancellation safety, and negative surfaces | Core 04 §2 | Core 01 §12.2.2 transport; Core 03 §4.3.3 transition; Core 04 conformance | Core 04 alone owns the local trust boundary, forbidden values, journal visibility, resource closure, and the absence of browser, HTTP, WebSocket, job, session, CSRF, bearer, and public audit surfaces. | REQ-00-068 | base | AC-535 |
| Database migration lifecycle, readiness, remediation, evidence, and recovery metadata | Core 01 §2.1A | Core 04 security and conformance; source-owner specifications; OpenTelemetry NLSpec; Testing Harness NLSpec; implementation guides | Core 01 owns the typed lifecycle boundary. PostgreSQL connectivity and secrets remain in the platform adapter; source owners retain schema meaning; authored SQL and evidence routing do not transfer lifecycle ownership. | REQ-00-069 | base | AC-537 |
| Extension-profile recognition, claimability, current major, primary owner, dependencies, and adopted-document status | Core 00 §4.2 and §5 | Core 01 extension discovery; Core 04 claim authorization and conformance; adopted Extensions NLSpec | Core 00 alone owns whether an extension identity is recognized and claimable/current and assigns its current major, primary owner, and dependencies. Core 01 enumerates only Core-00-recognized identities and owns the public discovery shape. Core 04 owns authorization and lifecycle consequences. The Extensions NLSpec owns shared mechanics only after coordinated adoption and cannot create recognition. | REQ-00-064, REQ-00-065 | base | AC-231, EXT-AC-145, EXT-AC-146, EXT-AC-157 |
| Adopted subsystem NLSpec deployment-config namespace | Adopted subsystem NLSpec for namespace-local keys; Core 04 §12 for artifact, discovery, overlay, unknown-key, and startup validation mechanics | Core 04 §12, subsystem NLSpec | Core 04 owns the deployment-config container and fail-closed validation mechanics. The adopted subsystem NLSpec owns only its closed key namespace and namespace-local cross-key rules. | REQ-00-052 | base | AC-231 |
| Public success/error envelope and public error-code and reason-code registries | Core 01 §3.3.6, §3.3.6.1, and §3.3.6.2 | Core 03 §3.3.4; Core 04 §9.6, §9.9, and §9.10 | Secondary sections MAY require a specific code or payload member but MUST NOT assign a conflicting meaning, transport status, or retry hint. | REQ-00-006 | base | AC-231 |
| Background-job resource shell, cancel semantics, retention semantics, and reusable `job_progress` payload members | Core 01 §3.3.9 and §3.3.9.1 | Core 01 §3.3.10.1; Core 03 §4.3.1 and §4.4; Core 04 §2 and §9.10; Appendix E and Appendix F | Core 01 owns the canonical HTTP job resource, cancel route semantics, post-terminal retention contract, the shared `scope`, `status`, `progress`, `cancelable`, `result_summary`, `error_summary`, and `retained_until` members reused by `job_progress`, the common `result_summary.code` rules, the shared `result_summary.resource_refs.kind`, `id`, and `route` semantics, and the fact that `job_progress` inherits those exact result-summary semantics from the canonical job resource. Core 03 owns only local client behavior under replay, resync, auth churn, and same-surface terminal-result rendering. Core 04 owns only authorization and conformance criteria. | REQ-00-024 | base | AC-231 |
| Canonical `view_row_v1` and `view_row_patch_v1` objects, full-row field inclusion, row-refresh payloads, and sparse collaboration row patches | Core 01 §3.3.4, §3.3.5, §3.3.10.1, and §7.4 | Core 03 §4.3.1 and §16.2; Core 04 §9.0, §9.1, and §9.10; Appendix C; Appendix D; Appendix F | Core 01 owns the canonical row-envelope family, full-row `cells` membership derived from the active `view_schema` field registry, `data.row` row-refresh responses for row-returning create, patch, and attach-blob success paths, `patch_cells` omission-versus-`null` semantics, and the `invalidate` fallback when a safe sparse patch cannot be expressed. Core 03 owns only client-application behavior and inspector consequences. Core 04 owns only conformance. Appendices MAY describe realization, worked examples, or traceability only. | REQ-00-025 | base | AC-231 |
| Workbook row-query search/filter semantics and any public discovery contract for workbook surfaces | Core 01 §3.3.4 and §3.3.4.1 | Core 02 §6.4 and §16; Core 03 §12, §13.2, §14, and §16.2; Core 04 §9.1 and §9.10; Appendix C; Appendix D; Appendix E; Appendix F; Appendix I | Core 01 owns the public route inventory, route/viewquery contract, operator vocabulary, operand normalization, tokenization, thresholds, result ordering, limits, cursor binding, saved-view query validation, public error mapping, and any future public discovery surface. Core 02 owns only suggestion-boundary consequences and realization notes. Core 03 owns only workbook UI consequences such as quick link/resolve entry points and same-surface suggestion presentation. Core 04 owns only conformance. Appendices MAY describe realization, rationale, examples, characterization matrices, or backlog only. No implementation file name is normative for this contract family. | REQ-00-026 | base | AC-231, AC-471 |
| Projection provider descriptors, projection descriptor validation manifests, restore-rebuild adapter boundary, and production import-boundary guardrails | Core 01 §8 and §12.2 | Core 04 §9.1A; Appendix I; `docs/domain.md`; implementation guides | Core 01 owns projection-derived-state invariants, code-backed provider descriptor invariants, the validation-only descriptor manifest boundary, restore projection-rebuild adapter semantics, and production projection import-boundary policy. Core 04 owns only conformance criteria. Appendix I MAY record evidence, characterization matrices, current implementation allowlists, and guard-test guidance but MUST NOT become runtime authority. | REQ-00-063 | base | AC-469..AC-473 |
| Collaboration payload-array canonicalization for `presence_snapshot.payload.presences[]`, `record_changed.payload.changed_field_keys[]`, and `record_changed.payload.affected_views[]` | Core 01 §3.3.10.1 | Core 03 §4.3.1; Core 04 §9.10; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns required presence, explicit empty-array rules, duplicate prohibition, canonical exact-identifier ordering, keyed-list or set semantics, and canonicalization across live, replay, and reset pathways for these public arrays. Core 03 owns only client interpretation and UI consequences. Core 04 owns only conformance. Appendices MAY describe realization, worked examples, roadmap notes, or traceability only. | REQ-00-027 | base | AC-231 |
| Benchmark-profile registry, benchmark manifests, measurement-predicate registry, and claim-bearing publication rules for timed or fixture-sensitive implementation criteria | Core 05 | Core 04 §9; Appendix B; Appendix D; Appendix E; Appendix F | Core 04 owns behavioral thresholds, user-visible states, semantic timed-state terms, and implementation claim manifests. Core 05 owns benchmark-profile identifiers, exact benchmark-environment fields, the benchmark-manifest contract, the measurement-predicate registry, the claim-bearing versus informative distinction, and audit-bundle retention for public timed or fixture-sensitive claims. Appendices MAY describe topology diagrams, example manifests, timing illustrations, historical notes, or traceability only. | REQ-00-028 | claim_publication | PC-006 |
| Session resource shape and expiry fields | Core 01 §3.3.2.1 | Core 03 §4.4; Core 04 §1.1.1 and §9.10 | Core 01 owns the authenticated-session resource fields returned to clients. | REQ-00-029 | base | AC-231 |
| Session issuance, expiry, revocation, and concurrent-session behavior | Core 04 §1.1.1 | Core 01 §3.3.2.1 and §3.3.10.1; Core 03 §4.4 | Non-owner sections MAY reference `session_expires_at`, `session_revoked`, and retry behavior but MUST NOT widen allowed lifetime or revocation semantics. | REQ-00-007 | base | AC-231 |
| Authenticated root landing and visible-incident directory selection | Core 01 §3.3.2.1A | Core 01 §3.3.5.3.1; Core 03 §2.4; Core 04 §1, §2, and §9.10; Appendix D; Appendix E; Appendix F; UI/UX guide | Core 01 owns `/` as the default authenticated post-login destination, the visible-incident cardinality algorithm, the no-explicit-`sheet_ref` sole-incident open, the sole-visibility-loss fallback, and prohibited selection heuristics. Core 03 owns only workbook-internal startup after a workbook open begins. Core 04 owns only authorization and conformance checks. Appendices and guides MAY describe traceability, sequence illustrations, or composition only. | REQ-00-053 | base | AC-414 |
| Deployment administration browser entry, panels, and prohibited aggregate admin concepts | Core 01 §3.3.2.1B | Core 03 §2.4; Core 04 §2 and §9.10; Appendix D; Appendix E; Appendix F; `docs/design.md`; UI/UX guide | Core 01 owns the canonical browser identity, route boundary, globally reachable entry, allowed panel set, and prohibited all-incident catalog or generic deployment-settings concepts. Core 04 owns only authorization, capability-loss behavior, and conformance. Core 03 owns only workbook-startup non-interaction. Appendices, design material, and guides MAY describe presentation, sequence examples, and traceability only. | REQ-00-057 | base | AC-414, AC-427, AC-441 |
| Credential-lifecycle public contract, bootstrap-token transport, and deployment-admin credential reset or revoke-all actions | Core 01 §3.3.2.2 and §3.3.5.1 | Core 02 §3 and §14.1; Core 03 §4.4; Core 04 §1.1, §1.1.1, §2, §3, and §9.10; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns public route inventory, request and response bodies, bootstrap-token transport, route-scoped idempotency, and error use. Core 02 owns deployment-local credential-state persistence minima only. Core 03 owns only client auth-loss behavior after session revocation. Core 04 owns only security invariants, authorization boundaries, audit constraints, and conformance. | REQ-00-030 | base | AC-231 |
| Current-account profile and account-preference public contract | Core 01 §3.3.2.3 | Core 02 §14.1; Core 03 §2 and §4; Core 04 §1.1.1, §2, §3, and §9.10; Appendix C; Appendix D; Appendix E; Appendix F; `docs/domain.md`; `docs/design.md`; UI/UX guide | Core 01 owns `/api/v1/account/profile` and `/api/v1/account/preferences` route inventory, resource shapes, exact mutation members, omitted-versus-`null` behavior, density defaults, idempotency ordering, no-op behavior, and forbidden adjacent features. Core 02 owns deployment-local persistence minima only. Core 03 owns only workbook-density rendering consequences. Core 04 owns current-session authorization, audit, route classification, and conformance. Appendices, domain text, design material, and guides MAY illustrate or trace the contract only. | REQ-00-054 | base | AC-429..AC-432 |
| Administrative audit read projections | Core 01 §3.3.5.1A | Core 02 §14.1; Core 04 §2, §3, and §9.10; Core 01 §12; Appendix C; Appendix D; Appendix E; Appendix F; UI/UX guide | Core 01 owns `/api/v1/administrative-audit-events` and `/api/v1/incidents/{incident_id}/membership-audit-events`, resource shape, action-code and target-kind registries, list filters, ordering, pagination, cursor binding, and response semantics. Core 02 owns deployment-local persistence invariants only. Core 04 owns authorization, redaction, audit emission, immutability, retention, and conformance. Core 01 §12 owns backup inclusion and incident-portability exclusion consequences. Appendices and guides MAY illustrate realization, workflow, traceability, or UI presentation only. | REQ-00-056 | base | AC-437..AC-440 |
| Persistence realization status and deployment-local persistence invariants | Core 02 §14.1 | Core 04 §3 and §9.10; Appendix C; Appendix F | Core 02 owns the statement that exact physical persistence realization is non-normative in the current profile and owns the exact persistence invariants for deployment-local credential, bootstrap-completion, auth-binding, reference-pack, and portability-boundary state. Core 04 owns only audit, secrecy, authorization, and conformance consequences. Appendix C MAY describe realization only. | REQ-00-031 | base | AC-231 |
| First-deployment-admin bootstrap artifact, validation, consumption algorithm, one-time semantics, and handoff into ordinary MFA bootstrap | Core 01 §3.3.5.1 | Core 02 §3 and §14.1; Core 04 §2, §3, §9.12, and §12; Appendix B; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns the bootstrap-admin manifest schema, allowed and forbidden fields, normalization, startup-time consumption algorithm, one-time semantics, created-user defaults, and reuse of the ordinary TOTP bootstrap flow. Core 02 owns deployment-local persistence minima only. Core 04 owns trust boundaries, config-gated startup preflight, audit constraints, fail-closed behavior, and conformance. Appendices MAY describe examples, DDL sketches, diagrams, roadmap notes, or traceability only. | REQ-00-032 | base | AC-231 |
| First-admin bootstrap deployment-configuration binding, path rules, and startup-preflight diagnostics | Core 04 §12.3.2 and §12.6 | Core 01 §3.3.5.1; Core 04 §9.12; Appendix B; Appendix E; Appendix F | Core 04 owns the stable config key, absolute-path rules, when bootstrap validation is required, fail-closed startup behavior, and the bootstrap-specific `invalid_deployment_config` reason-code registry. Core 01 owns only the cross-referenced artifact and consumption semantics. Core 04 §9.12 owns conformance only. | REQ-00-033 | base | AC-231 |
| Saved-view route, mutability, and authorization contract | Core 01 §3.3.5.2 | Core 02 §11.1; Core 03 §2.3-§2.4; Core 04 §9.10 | Core 02 owns persistence fields only. Core 03 owns workbook discoverability and startup interaction only. | REQ-00-034 | base | AC-231 |
| Workbook startup-preference objects | Core 02 §11.2 | Core 01 §3.3.5.2; Core 03 §2.4; Core 04 §9.10 | `home_sheet_ref` and `default_sheet_ref` semantics MUST remain separate and MUST NOT be collapsed into saved-view flags. | REQ-00-008 | base | AC-231 |
| Incident resource shape, incident-create contract, incident-metadata mutability, and incident lifecycle closure | Core 01 §3.3.5.3, §3.3.5.3.1, §3.3.5.3.2, and §3.3.10 | Core 02 §4.5, §14.1, and §18; Core 03 §4.4; Core 04 §9.10; Appendix C; Appendix D; Appendix E; Appendix F; `docs/domain.md`; UI/UX guide | Core 01 owns the public incident resource fields, `POST /api/v1/incidents` request and response contract, server-managed initial values, the create-only versus patchable field boundary, lifecycle transition routes, closed-incident operation boundaries, lifecycle route idempotency, and closed-state source-mutation serialization. Core 02 owns persistence minima and closed-vocabulary token membership only. Core 03 owns only workbook and local-pending-queue consequences after a closed-state signal. Archive, hard deletion, soft deletion, and purge are future-only and out of current-profile scope. Appendices, `docs/domain.md`, and guides MAY describe migration realization, examples, traceability, vocabulary guidance, or presentation consequences only. | REQ-00-035 | base | AC-231, AC-418, AC-419, AC-420, AC-421, AC-422, AC-423, AC-424, AC-425, AC-426 |
| Same-field conflict transport and `collection_review` resolver payloads | Core 03 §3.3.4 | Core 01 §3.3.5 and §3.3.6; Core 04 §9.6 and §9.10 | Core 01 owns the common envelope. Core 03 owns the conflict object, resolver semantics, and `collection_value_v1` conflict payload rules. | REQ-00-036 | base | AC-231 |
| Domain closed-vocabulary registry | Core 02 §18 | Core 01 view contracts; Core 03 workflow surfaces; Core 04 conformance criteria | Non-owner sections MAY require a subset only when they reference the exact tokens owned by Core 02. | REQ-00-037 | base | AC-231 |
| Lifecycle-machine states and legal transitions for `task_request` and `decision` | Core 02 §10.4.1.1 and §10.4.2.1 | Core 01 §3.3.6; Core 03 §6 and §16.4; Core 04 §9.9 | Core 02 owns state sets, legal transitions, and post-commit guard semantics. Core 01 owns the common illegal-transition transport shape. Core 02 does not own omitted-on-create defaults, owner-assignment defaults, priority defaults, decided-at defaults, or any other create-time initial-value policy. Core 01 owns those create-time defaults where the relevant workbook surface is defined. Core 04 owns pass/fail verification. | REQ-00-038 | base | AC-231 |
| Lifecycle-machine states, legal transitions, bridge-derived outcomes, and quarantine or recovery semantics for `object_blobs.upload_state` and `evidence_records.lifecycle_state` | Core 02 §13.1 and §13.2 | Core 01 §3.3.6; Core 03 §8.3-§8.4; Core 04 §9.9 | Core 02 owns the closed state sets, legal transitions, bridge-derived outcomes, contradiction handling, recovery paths from `quarantined`, and post-commit observable signals for the blob-upload and evidence-custody machines. Core 01 owns only the common illegal-transition transport shape. Core 03 owns only workbook-surface consequences and blocked-state presentation. Core 04 owns only pass/fail fixtures and contradiction checks. | REQ-00-039 | base | AC-231 |
| Timeline supersede replacement relation | Core 01 §3.3.5 and §7.4.1 | Core 02 §12.1-§12.3; Core 03 §6; Core 04 §9.1 and §9.11; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns the supersede-route request and response contract, idempotency comparison, and the Timeline read-side field. Core 02 owns the authoritative `record_links` realization, legal endpoint pairs, canonical direction, active-link cardinality, and source-state export consequences. Core 03 owns reviewer-surface invocation, nearby visibility, row-history presentation, and correction-path semantics. Core 04 owns conformance only. Appendices MAY describe realization, illustrations, roadmap notes, or traceability only. | REQ-00-040 | base | AC-231 |
| Deployment topology, application-unit boundary, and authoritative service separation | Core 01 §1 | Core 04 §5.2-§8; Appendix B; Appendix F | Core 01 owns the modular-monolith choice, the one web application deployable boundary, and the authoritative Postgres service plus S3-compatible object-storage service. Secondary sections MAY describe deployment-profile consequences, explanatory rationale, traceability, or conformance only. Secondary sections MUST NOT restate, weaken, or relax those topology rules as independent normative behavior. | REQ-00-041 | base | AC-231 |
| Operational backup, coherent restore, retention floor, equivalence, restore verification, and operator recovery CLI | Core 01 §12.1 and §12.2 | Core 04 §2, §4.4, §4.5, §6, §9.0.1, §9.14, and §12; adopted source-owner subsystem specifications; Appendix B; Appendix E; Appendix F; Appendix I | Core 01 owns `backup_set`, `backup_attestation`, `consistency_point_at`, the recovery-state contribution and catalog protocols, backup creation admission and publication, failed-candidate non-publication, latest-success selection stability, minimum artifact-set requirements, coherent capture and restore semantics, versioned recovery artifact selection, restore order, restore projection-rebuild adapter contract, retention floor, restore verification, owner-registered workbook-probe coordination, per-backup due-verification execution, equivalence criteria, separation from incident portability, logical operator recovery command grammar, typed semantic failure mapping, operator recovery output schemas, timeout defaults and bounds, exit codes, operator error vocabulary, and the boundary that backup scheduling is deployment-owned external orchestration. Each source owner owns the classification, inventory, validation, rebuild, or invalidation behavior for its state contribution. Core 04 owns only deployment-local operator authorization, no-listener and no-browser trust boundaries, recovery-operation exclusion, bound restore-target markers, the serving-lease safety proof, restore-target preflight, backup-storage binding, encrypted-root consequences, operator output redaction, typed encrypted recovery journal requirements, atomic safe administrative-audit completion, and pass/fail conformance criteria. Appendices MAY describe realization examples, operator checklists, roadmap notes, characterization matrices, or traceability only. | REQ-00-023 | base | AC-231, AC-472 |
| Destructive-operation concurrency and public contention failure for restore, rollback, and merge | Core 01 §3.3.5.0 | Core 03 §5; Core 04 §9.1; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns destructive-operation family membership, protected-set rules, canonical lock-acquisition order, fail-fast contention behavior, evaluation precedence, and required use of `record_locked`. Secondary sections MAY describe only local workbook consequences, non-normative realization notes, or conformance checks. | REQ-00-042 | base | AC-231 |
| Reference-pack storage-independent semantics, durable version conditions, and verification/activation lifecycle | Core 01 §11.3, §11.3.1, §11.4, §11.4.1, and §17.4 | Core 02 §14.1 and §17; Core 04 §4.1, §6, §9.4, §12.3.1, §12.4, and §12.6; Appendix C; Appendix F; implementation guides | Core 01 owns Reference Pack meaning, logical storage references, the public durable-condition vocabulary, the storage-to-public derivation boundary, legal activation preconditions, and verification/activation lifecycle semantics. Core 02 owns persistence minima only. Core 04 owns admitted roots, resource limits, hostile-path and hostile-content safety, secret-safe diagnostics, and conformance. Physical Go package placement is non-normative implementation structure and MUST NOT transfer semantic ownership or security authority. Appendix C and implementation guides MAY describe storage realization only. Appendix F MAY record traceability only. | REQ-00-043 | reference_pack | AC-234 |
| Base-profile `view_schema` registry and per-schema field registries | Core 01 §7.4 and §19 | Core 03 §14, §16.1, and §20; Core 04 §9.1; Appendix E and Appendix F | Core 01 owns the exact base-profile `view_schema_id` set and the exhaustive per-field contracts for each schema, including stable `field_key`, default sort, filter whitelist, minimum create signal, zero-field-create eligibility, create-time writeability, omitted-member defaults, create-time omitted-versus-`null` behavior, server-managed initial values, write target or action, `conflict_resolution_class`, and `entity_binding_mode` where applicable. Secondary sections MAY describe local interaction, conformance, or roadmap consequences only. | REQ-00-044 | base | AC-231 |
| Timeline time-conversion profile resource | Core 01 §3.3.5 and §7.4.1 | Core 02 §14.1; Core 03 §15; Core 04 §2 and §9.1 | Core 01 owns the public route inventory, request and response shape, fixed-offset semantics, disabled defaults, concurrency token, and row-write consequences. Core 02 owns persistence minima only. Core 03 owns workbook rendering and editing consequences only. Core 04 owns authorization and conformance checks only. | REQ-00-060 | base | AC-444, AC-449, AC-451 |
| Writable direct-temporal-scalar contract registry | Core 01 §18A | Core 01 §7.4 and §19; Core 03 §13.1; Core 04 §9.0 and §9.1; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 §18A owns accepted timestamp lexical form, canonical UTC normalization and equality, and authoritative clear semantics for writable temporal scalars bound through `direct_scalar_contract_id`. Core 01 §7.4 and §19 own only per-field bindings and `clearable` declarations. Core 03 §13.1 owns only workbook-surface invalid-draft and failed-clear consequences for fields bound to `direct_scalar_contract_id=timestamp_instant_v1`. Core 04 owns only claim manifests and conformance checks. | REQ-00-045 | base | AC-231 |
| Writable direct-reference-scalar contract registry | Core 01 §18B | Core 01 §7.4 and §19; Core 04 §9.0 and §9.1; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 §18B owns accepted lexical shape, normalization and equality, and authoritative clear semantics for writable direct-reference scalars bound through `direct_reference_contract_id`. Core 01 §7.4 and §19 own only per-field bindings and `clearable` declarations. Core 04 owns only claim manifests and conformance checks. | REQ-00-020 | base | AC-231 |
| Evidence blob attachment and evidence-access handle issuance and redemption | Core 01 §3.3.8 for `POST /api/v1/evidence-records/{record_id}/attach-blob`, and Core 01 §16 for handle issuance and redemption | Core 02 §13 and §18; Core 03 §8.1 and §8.4; Core 04 §2, §4.3, §4.5, and §9.10; Appendix F | Core 01 §3.3.8 owns `POST /api/v1/evidence-records/{record_id}/attach-blob`, including request shape, record-scoped idempotency, normalized replay comparison, optimistic-concurrency behavior, and route-specific error use. Core 01 §16 owns preview and download issuance plus redeem semantics, handle lifetime, revocation, filename and disposition rules, and route-specific error use. Core 02 owns only the exact tokens, metadata fields, and bridge preconditions referenced by those route contracts; the lifecycle-machine states and legal transitions themselves are owned by the dedicated Core 02 §13.1 and §13.2 lifecycle row above. Core 03 owns workbook-surface invocation and blocked-state UI only. Core 04 owns authorization re-derivation, active-content blocking, and fail-closed behavior only. | REQ-00-046 | base | AC-231 |
| Extension route-family public contracts for imports, reference packs, snapshots and releases, and incident bundles | Core 01 §17 | Core 03 §11.2; Core 04 §9.0, §9.2, §9.3, §9.4, and §9.11; Appendix E and Appendix F | Core 01 owns exact route inventory, request and response defaults, omitted-versus-`null` behavior, route-scoped idempotency, reuse of the common job resource, family-specific error registries, durable terminal-state representation, family-specific terminal success-code mappings, and the required `result_summary.resource_refs[]` kind/count/route mapping for each route family. Core 03 owns only underlying workflow and same-surface UI semantics. Core 04 owns only claim manifests and conformance checks. | REQ-00-047 | import, snapshot_reporting, incident_portability, reference_pack | AC-232, AC-233, AC-234, AC-236 |
| Import orchestration, target registry, analytical binding shape, and exact analytical target payload ownership | Core 01 §17.2 for upload/session/unit/mapping/source integrity, dispatch, generated registry, common authorization, unit outcomes, finalization, and binding shape; each adopted analytical target NLSpec for its referenced exact mapping, preview, apply, result, diagnostic, error, and resource-mutation schemas | Core 03 §11.2 for workflow consequences; Core 04 §9.2 for security and conformance; Extensions NLSpec for contribution admission; target NLSpecs for exact analytical semantics | Core MUST NOT duplicate an analytical target's exact payload members. An analytical target MUST NOT redefine Core import lifecycle, authorization, idempotency, cancellation, unit/session outcomes, or job publication. The typed binding joins the two owners and is not a callback bus or public API. View-schema targets continue to use their source-owner create facades. | REQ-00-066 | import, network_flow_activity | AC-264A, AC-464, AC-465, NF-AC-106 |
| Incident-bundle import initial-admin bootstrap membership | Core 01 §12.3.6 and §17.5 | Core 02 §14.1; Core 03 §2.4; Core 04 §2, §3, and §9.11; Appendix D; Appendix E; Appendix F; Appendix H; UI/UX guide | Core 01 owns admission binding to the submitting internal user, final-publication validation, atomic target-local membership, workbook-preference and audit creation, replay behavior, and failure reason. Core 02 owns deployment-local persistence minima only. Core 03 owns only imported-incident open behavior. Core 04 owns authorization, audit constraints, and conformance. Appendices and guides MAY describe operator practice, UI action placement, and traceability only. | REQ-00-058 | incident_portability | AC-442 |
| Reference Pack list-query and browser request-generation contract | Core 01 §17.4 | Core 04 §2 and §9.4; Appendix D; Appendix E; Appendix F; `docs/design.md`; UI/UX guide | Core 01 owns the accepted query members, source fields, exact filters, ordering, complete-collection evaluation, cursor binding, route-specific list-query failures, and browser request-generation algorithm for `GET /api/v1/reference-packs`. Core 04 owns authorization and conformance only. Appendices, design material, and guides MAY illustrate sequencing and presentation but MUST NOT redefine admission, pending, stale-response, cursor, or authorization-loss behavior. | REQ-00-059 | reference_pack | AC-443 |
| Runtime extension discovery and reserved-unclaimed extension-family semantics | Core 01 §3.3.3 and §3.3.6.1 | Core 04 §2 and §9.10 | Core 01 owns `GET /api/v1/extensions`, the exact current-profile `profile_id` registry exposed by that route, the exact `profile_id -> route_families[]` mapping, reserved-family path matching, dispatch precedence, and required use of `extension_profile_not_claimed` for reserved but unclaimed extension-family paths. Core 04 owns only route authorization and conformance checks. | REQ-00-022 | base | AC-231 |
| Inspector configuration and row-context workflow entry | Core 01 §6 and §7.4 for discovery shape, `inspector_config_v1`, route-binding vocabulary, feature-key grammar, route-owner tokens, and the exhaustive per-surface feature-group registry; Core 03 §2.3 and §2.3A for interaction behavior and deterministic inspector workflow algorithms | Core 02 §2 for source-state boundary; Core 04 §2 for authorization and egress security; appendices and guides for non-normative usage and design only; Core 05 only for claim-bearing timing if later asserted | Core 01 owns view-schema metadata shape, emitted config validity, and feature-group routing registry. Core 02 owns the statement that inspector configuration is not incident source state and introduces no record family. Core 03 owns default-closed behavior, saved-view inheritance, no-row state, stale row-bound invalidation, scroll separation, continuity, and deterministic row-context workflow execution. Core 04 owns authorization re-derivation and base-profile non-egress. Appendices and guides MAY illustrate usage but MUST NOT create required workflow cadence, per-edit rituals, authorization behavior, or route semantics. | REQ-00-061 | base | AC-453 |
| Enterprise-auth route family, callback/correlation-state semantics, provider-claim mapping, and provider-to-session convergence public contract | Core 01 §20 | Core 04 §1.2 and §9.5; Appendix C; Appendix E; Appendix F | Core 01 owns exact route inventory, discovery and initiation request and response defaults, callback transport, server-side correlation-state semantics, omitted-versus-`null` behavior, explicit non-idempotency exceptions, family-specific error registries, and provider-claim mapping. Core 04 owns only security invariants and conformance. Appendix C owns only persistence minima and deployment-local auth-state notes. | REQ-00-048 | enterprise_authentication | AC-235 |
| Enterprise-auth identity-binding lifecycle and deployment-admin binding mutation contract | Core 01 §20 | Core 01 §3.3.5.1; Core 02 §14.1; Core 04 §1.2, §2, §3, and §9.5; Appendix C; Appendix E; Appendix F | Core 01 §20 owns the deployment-admin binding route inventory, request and response defaults, enterprise-binding summary shape, route-scoped idempotency, callback interaction, and family-specific error use. Core 01 §3.3.5.1 owns only the base safe-user resource shell and local-binding summary. Core 02 owns deployment-local persistence minima and active-binding invariants only. Core 04 owns authorization, audit, session-revocation, and conformance only. Appendices MAY describe realization, roadmap notes, or traceability only. | REQ-00-021 | enterprise_authentication | AC-235 |
| Enterprise-auth claim activation, provider manifest, startup validation, and provider-definition reconciliation | Core 04 §12.3.4 and §12.6 | Core 01 §20; Core 02 §14.1; Core 04 §1.2 and §9.5; Appendix B; Appendix C; Appendix D; Appendix E; Appendix F; implementation guides | Core 04 owns the `enterprise_authentication.claimed` and `enterprise_authentication.provider_manifest_path` deployment-configuration keys, manifest schema, defaults, omitted-versus-`null` behavior, bounds, referenced-file and `secret_ref_v1` validation, fail-closed startup diagnostics, and deterministic provider-definition reconciliation. Core 01 owns only the public route consequences, including safe provider discovery and the absence of runtime provider-definition mutation routes. Core 02 owns only deployment-local provider-definition persistence invariants. Appendices and guides MAY illustrate, trace, or support implementation but MUST NOT widen the runtime provider configuration surface. | REQ-00-055 | enterprise_authentication | AC-433..AC-436 |
| Deployment configuration artifact, discovery and override precedence, application public origin, root-binding key registry, binding-kind model, filesystem-root path validation, disconnected-layout defaults, and startup validation contract | Core 04 §12 | Core 01 §14; Core 04 §6 and §9.12; Appendix B; Appendix E; Appendix F | Core 04 §12 owns the operator-facing config artifact, key names, browser application origin binding, binding semantics, default disconnected-layout locations, validation error family, and fail-closed startup behavior. Core 01 §14 may state only local packaging and runtime consequences. Core 04 §6 may summarize runtime-root consequences and Core 04 §9.12 may verify them, but neither may redefine the config surface. | REQ-00-049 | base | AC-231 |
| Deployment resource-limit registry and numeric safety boundaries | Core 04 §12.3.1 Resource-limit registry | Core 01 §3.3.8; Core 01 §16; Core 01 §17.2; Core 01 §17.4; Core 01 §12.3.6; Core 04 §9.13; Appendix B; Appendix E; Appendix F | Core 04 §12.3.1 owns stable resource-limit key names, exact defaults, units, numeric domains, omitted-key semantics, per-family overrides, and computation rules for deployment-configurable resource ceilings only. This registry does not own or parameterize the fixed public-contract ceilings on `sort[]`, `filters[]`, `changes[]`, or `collection_actions_v1.actions[]`; those ceilings are owned by Core 01 owner sections and MUST remain deployment-invariant in the current profile. Core 01 MAY bind route behavior, transport status, and route-family error consequences to registry keys only where Core 04 declares such keys, and Core 04 §9.13 MAY verify both the registry keys and the exclusion of those fixed public ceilings from `limits.*`. | REQ-00-050 | base | AC-231 |


### 5.2 Editorial traceability and corpus-maintenance notes

The corpus continues to use stable `REQ-*` and acceptance-criterion identifiers, `Profiles:` and `Verified by:` trailers, `Verifies:` back-references, selector expansion in Appendix F, and non-reuse of retired identifiers as editorial corpus-maintenance controls.

These controls are important for document quality, linting, navigation, and change review. They are not implementation-conformance requirements and they are not part of the Base Profile Definition of Done.

Corpus-maintenance failures MAY still be enforced in CI, spec-lint, or editorial review. Such failures are editorial corpus failures rather than implementation-conformance failures.

For corpus maintenance, every contract-owner matrix row must populate `Requirement ID`, `Profiles`, and `Verified by`. Blank cells are corpus defects. Non-owner sections and appendices must either describe local consequences, realization examples, traceability, or historical context, or be promoted into the owner section.

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
- when more than one persisted field, record, relation, or declared bridge rule determines the current machine condition, the deterministic derivation rule, the handling of contradictory persisted inputs, and the fail-closed behavior when no legal machine condition can be derived,
- the observable signals exposed to analysts, APIs, jobs, or logs,
- CI-verifiable conformance checks for happy path, terminal failure paths, illegal transitions, idempotent retry after simulated crash, and deterministic rerun from the same starting state.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-231, AC-313

**REQ-00-017**
This notation is a specification pattern. It MUST NOT be read as a requirement to adopt a runtime finite-state-machine library or framework.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-231

Representational lifecycle diagrams, sequences, and arrows that summarize a normative lifecycle machine are editorial aids. They remain non-normative, identify the owner section for the authoritative machine contract, and do not create additional implementation-conformance requirements.

**REQ-00-019**
If one normative lifecycle machine depends on another machine's result, the dependency MUST be expressed only through the other machine's contracted persisted state, declared bridge fields, or explicit terminal outcome. A lifecycle machine MUST NOT depend on another machine's in-memory state, UI-local state, background-worker-local state, or undeclared side effects.
Profiles: base
Verified by: AC-231, AC-313

## 10. Source-preservation note

The exploratory artifact contained several mixed-status areas, especially around:

- the boundary between clipboard paste and file-based structured import,
- immutable snapshot and report generation,
- reference-pack refresh and distribution.

This core resolves those areas as follows:

- current closed behavior is expressed normatively here,
- optional extension profiles capture source behaviors that were described as important but not uniformly day-one,
- future-evolution questions, historical source-question material, and editorial backlog remain in Appendix E,
- the original artifact remains preserved in Appendix G.
