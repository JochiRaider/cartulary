# Cartulary Normative Core 00: Document Set Status, Precedence, and Conformance

## 1. Status

This document set is the authoritative normative core for the current Cartulary profile.

Core 00 through Core 04 are the implementation-conformance corpus for the current profile. Core 05 is a separate normative companion for claim publication and benchmark reproducibility only. It governs claim-bearing publication requirements for timed or fixture-sensitive criteria and is not part of Base Profile or extension-profile implementation conformance. Appendices remain non-normative.

It is derived from the exploratory design artifact preserved in Appendix G. For the current profile, the base profile and the currently defined extension-profile boundaries are closed here. The normative core is authoritative for current-profile requirements, closed design decisions, and current conformance claims. Roadmap items, rationale, illustrative UI mockups, explanatory diagrams, source extracts, and historical source-question material are non-normative unless restated in the normative core as explicit requirements.

Contract-owner coverage for repeated families is complete in Core 00 §5.1. Appendix E is roadmap, historical source-question material, and future-only editorial backlog; it is not a live source of unresolved current-profile contract decisions. Any later typo, formatting, link, or similar non-substantive correction is corpus maintenance only and does not reopen current-profile design status, conformance scope, or profile boundaries.

The remaining hardening work for this corpus is editorial normalization only. It does not reopen current-profile runtime behavior, profile boundaries, or implementation-conformance scope.

Current-profile workbook-surface identity and registry closure, including pack-dependent surface constraints, are owned by Core 01 §7.4 and Core 03 §2. Appendices may describe future candidates but do not define current-profile workbook surfaces.

**REQ-00-051**
Non-normative supporting-guidance artifacts MAY describe recommended operator practice for tracker hygiene, companion findings-document discipline, handoff quality, status-review cadence, workload redistribution, debrief discipline, and challenge/escalation practice. Such guidance MUST NOT define implementation-conformance requirements, required row-creation cadence, required per-edit ritual, required authorization behavior, or required runtime workflow unless the behavior is restated normatively in Core 00 through Core 04.
Profiles: base
Verified by: AC-231

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

This reservation also includes local-account WebAuthn or passkey support, including registration, assertion, credential enumeration, credential revocation or reset, and recovery semantics. Such support is non-conformant in the current profile unless later NLSpecs define it.

## 5. Document map

- **Core 01** defines authoritative system topology, storage boundaries, view contracts, public success/error envelope registries, projections, portability, failure handling, and background-job behavior.
- **Core 02** defines authoritative record types, identifiers, mention/entity semantics, canonical closed vocabularies, provenance, deduplication, merge rules, schema invariants, and history mechanics.
- **Core 03** defines authoritative workbook interactions, collaboration semantics, workflow contracts, auto-resolution policy, grouping behavior, and write-back rules.
- **Core 04** defines authoritative security, deployment, trust boundaries, and acceptance criteria.
- **Core 05** defines authoritative claim publication and benchmark reproducibility for public timed or fixture-sensitive claims. It is normative companion material and is not part of Base Profile or extension-profile implementation conformance.

### 5.1 Contract-owner matrix

**REQ-00-005**
A contract family that appears in more than one normative core document MUST have one primary owner section. Non-owner sections MAY restate the owner only to declare local consequences, UI affordances, storage realization, or conformance checks. Owner-owned create-time behavior includes, at minimum, minimum semantic create signal, zero-field-create eligibility, create-time writeability, omitted-member defaults, create-time omitted-versus-`null` behavior, and server-managed initial values. Core 04 acceptance criteria MAY repeat owner-held create-time behavior only to verify it. Outside the owner section, and outside such Core 04 verification text, non-owner sections MUST NOT define or restate those owner-owned create-time behaviors as independent normative behavior. When a non-owner restatement and the owner differ, the owner governs and the restatement is editorial drift that MUST be repaired.
Profiles: base
Verified by: AC-231

| Contract family | Primary owner | Allowed secondary sections | Ownership rule | Requirement ID | Profiles | Verified by |
| --- | --- | --- | --- | --- | --- | --- |
| Public success/error envelope and public error-code and reason-code registries | Core 01 §3.3.6, §3.3.6.1, and §3.3.6.2 | Core 03 §3.3.4; Core 04 §9.6, §9.9, and §9.10 | Secondary sections MAY require a specific code or payload member but MUST NOT assign a conflicting meaning, transport status, or retry hint. | REQ-00-006 | base | AC-231 |
| Background-job resource shell, cancel semantics, retention semantics, and reusable `job_progress` payload members | Core 01 §3.3.9 and §3.3.9.1 | Core 01 §3.3.10.1; Core 03 §4.3.1 and §4.4; Core 04 §2 and §9.10; Appendix E and Appendix F | Core 01 owns the canonical HTTP job resource, cancel route semantics, post-terminal retention contract, the shared `scope`, `status`, `progress`, `cancelable`, `result_summary`, `error_summary`, and `retained_until` members reused by `job_progress`, the common `result_summary.code` rules, the shared `result_summary.resource_refs.kind`, `id`, and `route` semantics, and the fact that `job_progress` inherits those exact result-summary semantics from the canonical job resource. Core 03 owns only local client behavior under replay, resync, auth churn, and same-surface terminal-result rendering. Core 04 owns only authorization and conformance criteria. | REQ-00-024 | base | AC-231 |
| Canonical `view_row_v1` and `view_row_patch_v1` objects, full-row field inclusion, row-refresh payloads, and sparse collaboration row patches | Core 01 §3.3.4, §3.3.5, §3.3.10.1, and §7.4 | Core 03 §4.3.1 and §16.2; Core 04 §9.0, §9.1, and §9.10; Appendix C; Appendix D; Appendix F | Core 01 owns the canonical row-envelope family, full-row `cells` membership derived from the active `view_schema` field registry, `data.row` row-refresh responses for row-returning create, patch, and attach-blob success paths, `patch_cells` omission-versus-`null` semantics, and the `invalidate` fallback when a safe sparse patch cannot be expressed. Core 03 owns only client-application behavior and inspector consequences. Core 04 owns only conformance. Appendices MAY describe realization, worked examples, or traceability only. | REQ-00-025 | base | AC-231 |
| Workbook row-query search/filter semantics and any public discovery contract for workbook surfaces | Core 01 §3.3.4 and §3.3.4.1 | Core 02 §6.4 and §16; Core 03 §12, §13.2, §14, and §16.2; Core 04 §9.1 and §9.10; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns the public route inventory, operator vocabulary, operand normalization, tokenization, thresholds, result ordering, limits, cursor binding, and any future public discovery surface. Core 02 owns only suggestion-boundary consequences and realization notes. Core 03 owns only workbook UI consequences such as quick link/resolve entry points and same-surface suggestion presentation. Core 04 owns only conformance. Appendices MAY describe realization, rationale, examples, or backlog only. | REQ-00-026 | base | AC-231 |
| Collaboration payload-array canonicalization for `presence_snapshot.payload.presences[]`, `record_changed.payload.changed_field_keys[]`, and `record_changed.payload.affected_views[]` | Core 01 §3.3.10.1 | Core 03 §4.3.1; Core 04 §9.10; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns required presence, explicit empty-array rules, duplicate prohibition, canonical exact-identifier ordering, keyed-list or set semantics, and canonicalization across live, replay, and reset pathways for these public arrays. Core 03 owns only client interpretation and UI consequences. Core 04 owns only conformance. Appendices MAY describe realization, worked examples, roadmap notes, or traceability only. | REQ-00-027 | base | AC-231 |
| Benchmark-profile registry, benchmark manifests, measurement-predicate registry, and claim-bearing publication rules for timed or fixture-sensitive implementation criteria | Core 05 | Core 04 §9; Appendix B; Appendix D; Appendix E; Appendix F | Core 04 owns behavioral thresholds, user-visible states, semantic timed-state terms, and implementation claim manifests. Core 05 owns benchmark-profile identifiers, exact benchmark-environment fields, the benchmark-manifest contract, the measurement-predicate registry, the claim-bearing versus informative distinction, and audit-bundle retention for public timed or fixture-sensitive claims. Appendices MAY describe topology diagrams, example manifests, timing illustrations, historical notes, or traceability only. | REQ-00-028 | claim_publication | PC-006 |
| Session resource shape and expiry fields | Core 01 §3.3.2.1 | Core 03 §4.4; Core 04 §1.1.1 and §9.10 | Core 01 owns the authenticated-session resource fields returned to clients. | REQ-00-029 | base | AC-231 |
| Session issuance, expiry, revocation, and concurrent-session behavior | Core 04 §1.1.1 | Core 01 §3.3.2.1 and §3.3.10.1; Core 03 §4.4 | Non-owner sections MAY reference `session_expires_at`, `session_revoked`, and retry behavior but MUST NOT widen allowed lifetime or revocation semantics. | REQ-00-007 | base | AC-231 |
| Credential-lifecycle public contract, bootstrap-token transport, and deployment-admin credential reset or revoke-all actions | Core 01 §3.3.2.2 and §3.3.5.1 | Core 02 §3 and §14.1; Core 03 §4.4; Core 04 §1.1, §1.1.1, §2, §3, and §9.10; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns public route inventory, request and response bodies, bootstrap-token transport, route-scoped idempotency, and error use. Core 02 owns deployment-local credential-state persistence minima only. Core 03 owns only client auth-loss behavior after session revocation. Core 04 owns only security invariants, authorization boundaries, audit constraints, and conformance. | REQ-00-030 | base | AC-231 |
| Persistence realization status and deployment-local persistence invariants | Core 02 §14.1 | Core 04 §3 and §9.10; Appendix C; Appendix F | Core 02 owns the statement that exact physical persistence realization is non-normative in the current profile and owns the exact persistence invariants for deployment-local credential, bootstrap-completion, auth-binding, reference-pack, and portability-boundary state. Core 04 owns only audit, secrecy, authorization, and conformance consequences. Appendix C MAY describe realization only. | REQ-00-031 | base | AC-231 |
| First-deployment-admin bootstrap artifact, validation, consumption algorithm, one-time semantics, and handoff into ordinary MFA bootstrap | Core 01 §3.3.5.1 | Core 02 §3 and §14.1; Core 04 §2, §3, §9.12, and §12; Appendix B; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns the bootstrap-admin manifest schema, allowed and forbidden fields, normalization, startup-time consumption algorithm, one-time semantics, created-user defaults, and reuse of the ordinary TOTP bootstrap flow. Core 02 owns deployment-local persistence minima only. Core 04 owns trust boundaries, config-gated startup preflight, audit constraints, fail-closed behavior, and conformance. Appendices MAY describe examples, DDL sketches, diagrams, roadmap notes, or traceability only. | REQ-00-032 | base | AC-231 |
| First-admin bootstrap deployment-configuration binding, path rules, and startup-preflight diagnostics | Core 04 §12.3.2 and §12.6 | Core 01 §3.3.5.1; Core 04 §9.12; Appendix B; Appendix E; Appendix F | Core 04 owns the stable config key, absolute-path rules, when bootstrap validation is required, fail-closed startup behavior, and the bootstrap-specific `invalid_deployment_config` reason-code registry. Core 01 owns only the cross-referenced artifact and consumption semantics. Core 04 §9.12 owns conformance only. | REQ-00-033 | base | AC-231 |
| Saved-view route, mutability, and authorization contract | Core 01 §3.3.5.2 | Core 02 §11.1; Core 03 §2.3-§2.4; Core 04 §9.10 | Core 02 owns persistence fields only. Core 03 owns workbook discoverability and startup interaction only. | REQ-00-034 | base | AC-231 |
| Workbook startup-preference objects | Core 02 §11.2 | Core 01 §3.3.5.2; Core 03 §2.4; Core 04 §9.10 | `home_sheet_ref` and `default_sheet_ref` semantics MUST remain separate and MUST NOT be collapsed into saved-view flags. | REQ-00-008 | base | AC-231 |
| Incident resource shape, incident-create contract, and incident-metadata mutability | Core 01 §3.3.5.3 and §3.3.5.3.1 | Core 02 §4.5 and §14.1; Core 04 §9.10; Appendix C and Appendix E | Core 01 owns the public incident resource fields, `POST /api/v1/incidents` request and response contract, server-managed initial values, and the create-only versus patchable field boundary. Core 02 owns persistence minima only. | REQ-00-035 | base | AC-231 |
| Same-field conflict transport and `collection_review` resolver payloads | Core 03 §3.3.4 | Core 01 §3.3.5 and §3.3.6; Core 04 §9.6 and §9.10 | Core 01 owns the common envelope. Core 03 owns the conflict object, resolver semantics, and `collection_value_v1` conflict payload rules. | REQ-00-036 | base | AC-231 |
| Domain closed-vocabulary registry | Core 02 §18 | Core 01 view contracts; Core 03 workflow surfaces; Core 04 conformance criteria | Non-owner sections MAY require a subset only when they reference the exact tokens owned by Core 02. | REQ-00-037 | base | AC-231 |
| Lifecycle-machine states and legal transitions for `task_request` and `decision` | Core 02 §10.4.1.1 and §10.4.2.1 | Core 01 §3.3.6; Core 03 §6 and §16.4; Core 04 §9.9 | Core 02 owns state sets, legal transitions, and post-commit guard semantics. Core 01 owns the common illegal-transition transport shape. Core 02 does not own omitted-on-create defaults, owner-assignment defaults, priority defaults, decided-at defaults, or any other create-time initial-value policy. Core 01 owns those create-time defaults where the relevant workbook surface is defined. Core 04 owns pass/fail verification. | REQ-00-038 | base | AC-231 |
| Lifecycle-machine states, legal transitions, bridge-derived outcomes, and quarantine or recovery semantics for `object_blobs.upload_state` and `evidence_records.lifecycle_state` | Core 02 §13.1 and §13.2 | Core 01 §3.3.6; Core 03 §8.3-§8.4; Core 04 §9.9 | Core 02 owns the closed state sets, legal transitions, bridge-derived outcomes, contradiction handling, recovery paths from `quarantined`, and post-commit observable signals for the blob-upload and evidence-custody machines. Core 01 owns only the common illegal-transition transport shape. Core 03 owns only workbook-surface consequences and blocked-state presentation. Core 04 owns only pass/fail fixtures and contradiction checks. | REQ-00-039 | base | AC-231 |
| Timeline supersede replacement relation | Core 01 §3.3.5 and §7.4.1 | Core 02 §12.1-§12.3; Core 03 §6; Core 04 §9.1 and §9.11; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns the supersede-route request and response contract, idempotency comparison, and the Timeline read-side field. Core 02 owns the authoritative `record_links` realization, legal endpoint pairs, canonical direction, active-link cardinality, and source-state export consequences. Core 03 owns reviewer-surface invocation, nearby visibility, row-history presentation, and correction-path semantics. Core 04 owns conformance only. Appendices MAY describe realization, illustrations, roadmap notes, or traceability only. | REQ-00-040 | base | AC-231 |
| Deployment topology, application-unit boundary, and authoritative service separation | Core 01 §1 | Core 04 §5.2-§8; Appendix B; Appendix F | Core 01 owns the modular-monolith choice, the one web application deployable boundary, and the authoritative Postgres service plus S3-compatible object-storage service. Secondary sections MAY describe deployment-profile consequences, explanatory rationale, traceability, or conformance only. Secondary sections MUST NOT restate, weaken, or relax those topology rules as independent normative behavior. | REQ-00-041 | base | AC-231 |
| Operational backup, coherent restore, retention floor, equivalence, and restore verification | Core 01 §12.1 and §12.2 | Core 04 §2, §4.4, §4.5, §6, §9.0.1, §9.14, and §12; Appendix B; Appendix E; Appendix F | Core 01 owns `backup_set`, `backup_attestation`, `consistency_point_at`, minimum artifact-set requirements, coherent restore semantics, restore order, retention floor, restore verification, equivalence criteria, and the separation from incident portability. Core 04 owns only deployment-local operator boundaries, backup-storage binding, encrypted-root consequences, and pass/fail conformance criteria. Appendices MAY describe realization examples, operator checklists, roadmap notes, or traceability only. | REQ-00-023 | base | AC-231 |
| Destructive-operation concurrency and public contention failure for restore, rollback, and merge | Core 01 §3.3.5.0 | Core 03 §5; Core 04 §9.1; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 owns destructive-operation family membership, protected-set rules, canonical lock-acquisition order, fail-fast contention behavior, evaluation precedence, and required use of `record_locked`. Secondary sections MAY describe only local workbook consequences, non-normative realization notes, or conformance checks. | REQ-00-042 | base | AC-231 |
| Reference-pack durable version conditions and verification/activation lifecycle | Core 01 §11.3, §11.3.1, §11.4, and §11.4.1 | Core 01 §17; Core 02 §14.1 and §17; Core 04 §9.4; Appendix C; Appendix F | Core 01 owns the public durable-condition vocabulary, the storage-to-public derivation boundary, legal activation preconditions, and verification/activation lifecycle semantics. Core 02 owns persistence minima only. Core 04 owns conformance only. Appendix C MAY describe storage realization only. Appendix F MAY record traceability only. | REQ-00-043 | reference_pack | AC-234 |
| Base-profile `view_schema` registry and per-schema field registries | Core 01 §7.4 and §19 | Core 03 §14, §16.1, and §20; Core 04 §9.1; Appendix E and Appendix F | Core 01 owns the exact base-profile `view_schema_id` set and the exhaustive per-field contracts for each schema, including stable `field_key`, default sort, filter whitelist, minimum create signal, zero-field-create eligibility, create-time writeability, omitted-member defaults, create-time omitted-versus-`null` behavior, server-managed initial values, write target or action, `conflict_resolution_class`, and `entity_binding_mode` where applicable. Secondary sections MAY describe local interaction, conformance, or roadmap consequences only. | REQ-00-044 | base | AC-231 |
| Writable direct-temporal-scalar contract registry | Core 01 §18A | Core 01 §7.4 and §19; Core 03 §13.1; Core 04 §9.0 and §9.1; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 §18A owns accepted timestamp lexical form, canonical UTC normalization and equality, and authoritative clear semantics for writable temporal scalars bound through `direct_scalar_contract_id`. Core 01 §7.4 and §19 own only per-field bindings and `clearable` declarations. Core 03 §13.1 owns only workbook-surface invalid-draft and failed-clear consequences for fields bound to `direct_scalar_contract_id=timestamp_instant_v1`. Core 04 owns only claim manifests and conformance checks. | REQ-00-045 | base | AC-231 |
| Writable direct-reference-scalar contract registry | Core 01 §18B | Core 01 §7.4 and §19; Core 04 §9.0 and §9.1; Appendix C; Appendix D; Appendix E; Appendix F | Core 01 §18B owns accepted lexical shape, normalization and equality, and authoritative clear semantics for writable direct-reference scalars bound through `direct_reference_contract_id`. Core 01 §7.4 and §19 own only per-field bindings and `clearable` declarations. Core 04 owns only claim manifests and conformance checks. | REQ-00-020 | base | AC-231 |
| Evidence blob attachment and evidence-access handle issuance and redemption | Core 01 §3.3.8 for `POST /api/v1/evidence-records/{record_id}/attach-blob`, and Core 01 §16 for handle issuance and redemption | Core 02 §13 and §18; Core 03 §8.1 and §8.4; Core 04 §2, §4.3, §4.5, and §9.10; Appendix F | Core 01 §3.3.8 owns `POST /api/v1/evidence-records/{record_id}/attach-blob`, including request shape, record-scoped idempotency, normalized replay comparison, optimistic-concurrency behavior, and route-specific error use. Core 01 §16 owns preview and download issuance plus redeem semantics, handle lifetime, revocation, filename and disposition rules, and route-specific error use. Core 02 owns only the exact tokens, metadata fields, and bridge preconditions referenced by those route contracts; the lifecycle-machine states and legal transitions themselves are owned by the dedicated Core 02 §13.1 and §13.2 lifecycle row above. Core 03 owns workbook-surface invocation and blocked-state UI only. Core 04 owns authorization re-derivation, active-content blocking, and fail-closed behavior only. | REQ-00-046 | base | AC-231 |
| Extension route-family public contracts for imports, reference packs, snapshots and releases, and incident bundles | Core 01 §17 | Core 03 §11.2; Core 04 §9.0, §9.2, §9.3, §9.4, and §9.11; Appendix E and Appendix F | Core 01 owns exact route inventory, request and response defaults, omitted-versus-`null` behavior, route-scoped idempotency, reuse of the common job resource, family-specific error registries, durable terminal-state representation, family-specific terminal success-code mappings, and the required `result_summary.resource_refs[]` kind/count/route mapping for each route family. Core 03 owns only underlying workflow and same-surface UI semantics. Core 04 owns only claim manifests and conformance checks. | REQ-00-047 | import, snapshot_reporting, incident_portability, reference_pack | AC-232, AC-233, AC-234, AC-236 |
| Runtime extension discovery and reserved-unclaimed extension-family semantics | Core 01 §3.3.3 and §3.3.6.1 | Core 04 §2 and §9.10 | Core 01 owns `GET /api/v1/extensions`, the exact current-profile `profile_id` registry exposed by that route, the exact `profile_id -> route_families[]` mapping, reserved-family path matching, dispatch precedence, and required use of `extension_profile_not_claimed` for reserved but unclaimed extension-family paths. Core 04 owns only route authorization and conformance checks. | REQ-00-022 | base | AC-231 |
| Enterprise-auth route family, callback/correlation-state semantics, provider-claim mapping, and provider-to-session convergence public contract | Core 01 §20 | Core 04 §1.2 and §9.5; Appendix C; Appendix E; Appendix F | Core 01 owns exact route inventory, discovery and initiation request and response defaults, callback transport, server-side correlation-state semantics, omitted-versus-`null` behavior, explicit non-idempotency exceptions, family-specific error registries, and provider-claim mapping. Core 04 owns only security invariants and conformance. Appendix C owns only persistence minima and deployment-local auth-state notes. | REQ-00-048 | enterprise_authentication | AC-235 |
| Enterprise-auth identity-binding lifecycle and deployment-admin binding mutation contract | Core 01 §20 | Core 01 §3.3.5.1; Core 02 §14.1; Core 04 §1.2, §2, §3, and §9.5; Appendix C; Appendix E; Appendix F | Core 01 §20 owns the deployment-admin binding route inventory, request and response defaults, enterprise-binding summary shape, route-scoped idempotency, callback interaction, and family-specific error use. Core 01 §3.3.5.1 owns only the base safe-user resource shell and local-binding summary. Core 02 owns deployment-local persistence minima and active-binding invariants only. Core 04 owns authorization, audit, session-revocation, and conformance only. Appendices MAY describe realization, roadmap notes, or traceability only. | REQ-00-021 | enterprise_authentication | AC-235 |
| Deployment configuration artifact, discovery and override precedence, root-binding key registry, binding-kind model, filesystem-root path validation, disconnected-layout defaults, and startup validation contract | Core 04 §12 | Core 01 §14; Core 04 §6 and §9.12; Appendix B; Appendix E; Appendix F | Core 04 §12 owns the operator-facing config artifact, key names, binding semantics, default disconnected-layout locations, validation error family, and fail-closed startup behavior. Core 01 §14 may state only local packaging and runtime consequences. Core 04 §6 may summarize runtime-root consequences and Core 04 §9.12 may verify them, but neither may redefine the config surface. | REQ-00-049 | base | AC-231 |
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
