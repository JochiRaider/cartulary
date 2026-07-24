---
title: Cartulary Extensions Subsystem NLSpec
status: adopted/current
document_class: nlspec
profile: base
schema_id: cartulary.extensions_subsystem_nlspec.v1
document_version: 0.6.1
contract_major: 1
---

# 1. Status, scope, and authority

This NLSpec defines the Cartulary Extensions Subsystem. The subsystem is part of the Base Profile because profile recognition, extension discovery, reserved-route dispatch, claim resolution, inactive-profile behavior, registry integrity validation, verification routing, and extension contract accounting exist even when every optional extension profile is unclaimed.

This document is `status: adopted/current`. Its coordinated adoption gates in §29 were closed together with the exact companion owner revisions, generated artifacts, implementation evidence, and Harness accounting recorded by the controlling implementation tracker.

**EXT-REQ-001**
The Extensions Subsystem MUST own only these behavior families:

- authored extension owner-fragment admission;
- deterministic owner-input registry construction;
- exact dependency-snapshot validation;
- generated extension-profile descriptor source and canonical descriptor contracts;
- deterministic extension-registry construction and canonical serialization;
- descriptor, registry, integrity-object, and implementation-binding validation;
- extension-registry collision detection and deterministic finding enumeration;
- deployment claim-state derivation;
- required runtime extension dependency resolution;
- extension implementation admission;
- the closed extension contribution-point vocabulary;
- public contract-major and persisted-state compatibility classification;
- shared extension state-presence, migration-ledger, and migration-lock behavior;
- non-destructive unclaim, reclaim, upgrade, and retirement mechanics;
- extension-owned job proof and reconciliation when a profile is inactive;
- the closed extension fatal-condition registry and extension-local consequences imported into the Core 04 fatal lifecycle;
- shared extension security, egress, backup, portability, observability, and conformance boundaries where those behaviors are not already owned by Core 00 through Core 04.

Profiles: base
Verified by: EXT-AC-001, EXT-AC-003, EXT-AC-004, EXT-AC-007, EXT-AC-011, EXT-AC-021, EXT-AC-071, EXT-AC-076, EXT-AC-079, EXT-AC-085, EXT-AC-090

**EXT-REQ-002**
This NLSpec MUST NOT own:

- profile recognition, claimability, or adopted-document status;
- the public HTTP success or error envelope;
- the public `GET /api/v1/extensions` route;
- Base route matching or common route authorization;
- Base record families, `view_schema` resources, saved views, workbook query semantics, or Base workbook tabs;
- deployment-configuration artifact discovery, overlay parsing, unknown-key rejection, or the top-level deployment-configuration error envelope;
- extension-specific resources, route bodies, domain algorithms, resource limits, state transitions, or authorization matrices except for the shared boundaries explicitly allocated by this NLSpec;
- Reporting derivation, report composition, Graph Projection, OpenTelemetry signal semantics, or Testing Harness v2 owner catalogs, verification-contract validation, runner selection, execution profiles, scheduling, fixture lifecycle, retained evidence, or evidence-audit mechanics;
- process lifecycle, readiness-envelope shape, health-envelope shape, or process-exit semantics owned by Core 04;
- physical database-table, object-bucket, source-package, or generated-file placement;
- runtime installation, update, revocation, or execution of independently distributed third-party packages.

The owner allocations in Table 1-A MUST govern these exclusions.

Profiles: base
Verified by: EXT-AC-001, EXT-AC-059, EXT-AC-070, EXT-AC-074

**Table 1-A. Contract owner boundary**

| Contract family | Primary normative owner | Extensions Subsystem interaction |
| --- | --- | --- |
| Recognized `profile_id` values, current claimability, current contract major, adopted owner identity | Core 00 | Consumes digest-bound owner facts; MUST NOT infer or override them. |
| Public extension discovery and workbook-startup routes, response envelope, Base route-reservation registry, reserved route families, dispatch precedence, authorized workspace availability, common public errors, common job shell, cross-owner final-commit protocol, staged-object publication/cleanup, and physical backup orchestration | Core 01 | Supplies descriptor-derived inputs and imports the public contracts. |
| Core record and relationship model | Core 02 | Enforces the no-implicit-promotion and no-cross-owner-write boundaries. |
| Extension workspace rendering, startup fallback, lazy loading, unknown-value behavior, and authorization-loss consequences | Core 03 | Supplies declared workspace identities and imports client behavior. |
| Deployment configuration, claim keys, secret references, authorization, startup validation, application-process lease, extension deadline keys, publication admission gate, readiness, fatal shutdown lifecycle, and process-exit behavior | Core 04 | Defines namespace-local claim semantics and requires Core 04 companion adoption. |
| Extension-specific resources, routes, algorithms, limits, lifecycle, migration semantics, and role matrix | Named extension owner contract | Publishes owner fragments and imports shared descriptor, claim, dependency, compatibility, state, and lifecycle mechanics. |
| Graph-oriented projection | Graph Projection NLSpec | No redefinition. Extensions consume only declared interfaces admitted by their named owners. |
| Reporting and report composition | Reporting Subsystem and Report Composition NLSpecs | No redefinition. Extension participation is typed and explicit. |
| Telemetry signal shape | OpenTelemetry Instrumentation NLSpec | Consumes the canonical claimed-profile set. |
| Harness execution and retained evidence | Testing Harness NLSpec v2 | Owns verification-contract and owner-catalog validation, exact runner selectors, runtime/resource/fixture profiles, owner slices, scheduling, retained row evidence, owner summaries, evidence auditing, and harness gates. The Extensions Subsystem supplies product requirements and static contract inputs only. |
| Timed or fixture-sensitive public claims | Core 05 | This NLSpec creates no claim-publication behavior. |
| Vocabulary and owner navigation | `docs/domain.md` | Adds terms only after owner behavior is adopted. |
| Rationale, examples, operator guidance, diagrams, physical paths, and implementation mechanisms | Appendices and implementation-support guides | Non-normative only. |

**EXT-REQ-174**
The specification owner MUST author one closed `cartulary.extension_dependency_declaration_set.v1` object containing exactly `schema_id`, `extensions_document_version`, and `dependencies[]`. `schema_id` MUST equal `cartulary.extension_dependency_declaration_set.v1`; the document version MUST equal this NLSpec's exact `document_version`; and `dependencies[]` MUST be present, MUST NOT be `null`, and MUST contain exactly one declaration for every dependency in Table 1-B and no other declaration. An empty array is valid only when Table 1-B is empty. Every declaration MUST supply the exact owner-document identity, version and digest, owner-contract-manifest identity and digest, imported anchors, schemas, algorithms, artifacts, and required status used by the snapshot row below. Omission, explicit `null`, a duplicate, an extra declaration, a stale version or digest, and a manifest version or digest mismatch are invalid. The generator MUST NOT derive a dependency from imports, package references, source code, prose, or a prior generated snapshot.

Coordinated adoption MUST generate one canonical `cartulary.extension_dependency_snapshot.v1` object solely from the validated declaration set and the exact digest-bound owner manifests. The object MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_dependency_snapshot.v1`;
- `extensions_document_version`, exactly this NLSpec's adopted `document_version`;
- `dependencies[]`, containing exactly one row for every dependency in Table 1-B and no other row.

Each `dependencies[]` row MUST contain exactly:

- `dependency_id`;
- `owner_document_ref`;
- `owner_document_schema_id`;
- `owner_document_version`;
- `owner_document_sha256`;
- `owner_contract_manifest_ref`;
- `owner_contract_manifest_id`;
- `owner_contract_manifest_sha256`;
- `imported_anchor_refs[]`;
- `imported_schema_ids[]`;
- `imported_algorithm_ids[]`;
- `imported_artifacts[]`;
- `required_status`.

`dependencies[]` MUST be ordered by ascending UTF-8 bytes of `dependency_id`. `dependency_id` MUST equal one Table 1-B token and satisfy `^[a-z][a-z0-9_]{0,63}$`. `owner_document_ref` MUST be a whole-document `owner_locator_v1`. `owner_document_schema_id` MUST satisfy the public schema-ID scalar contract. `owner_document_version` MUST contain `1..64` ASCII bytes and match `[A-Za-z0-9][A-Za-z0-9_.+-]{0,63}`. `required_status` MUST equal `adopted/current`. `owner_document_sha256` MUST be the lowercase 64-character result of `owner_document_sha256_v1`.

`owner_contract_manifest_ref` MUST be a normalized repository-relative POSIX path under EXT-REQ-203 and MUST resolve to one regular file without following a symlink. `owner_contract_manifest_id` MUST contain `1..160` ASCII bytes and match `[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}`. `owner_contract_manifest_sha256` MUST be the lowercase digest returned by `extension_owner_contract_manifest_sha256_v1`. The manifest's owner-document identity, version, and digest MUST exactly equal the dependency row. Across `dependencies[]`, owner-document refs, owner-contract-manifest refs, and owner-contract-manifest IDs MUST each be unique.

`imported_anchor_refs[]` MUST contain `1..256` unique `owner_locator_v1` values sorted by ascending UTF-8 bytes. Every imported anchor MUST exist in the exact owner contract manifest named by the dependency row. `imported_schema_ids[]` and `imported_algorithm_ids[]` MUST each contain `0..256` unique public schema or algorithm IDs sorted by ascending UTF-8 bytes. `[]` for one of those two arrays means the dependency imports no schema or no algorithm directly; it does not waive the required anchor set.

Each `imported_artifacts[]` row MUST contain exactly:

- `artifact_id`;
- `schema_id`;
- `artifact_sha256`;
- `safe_ref`.

`imported_artifacts[]` MUST contain `0..64` rows, reject duplicate `artifact_id` values, and sort by ascending UTF-8 bytes of `artifact_id`. An `artifact_id` MUST occur in at most one dependency row. `artifact_id` and `schema_id` MUST each satisfy the public schema-ID scalar contract. `artifact_sha256` MUST be a SHA-256 digest string. `safe_ref` MUST satisfy EXT-REQ-233. The Core 01 dependency row MUST include exactly one artifact whose `artifact_id='cartulary.base_route_reservation_registry.current'` and `schema_id='cartulary.base_route_reservation_registry.v1'`. Every other dependency row MUST NOT declare that artifact ID.

Every array in the authored declaration set and generated snapshot MUST be present even when empty and MUST reject explicit `null`. The generated snapshot MUST exactly equal the normalized declaration set after owner-manifest verification. Omission behavior: no dependency row or member has a default. A missing, duplicate, extra, unresolved, non-adopted, stale, version-mismatched, digest-mismatched, schema-mismatched, algorithm-mismatched, manifest-mismatched, or imported-artifact-mismatched dependency MUST block registry generation and coordinated adoption before any downstream artifact is emitted. No `cartulary.extensions.phase2.v1` alias, reader, translation, or fallback is conformant.

Profiles: base
Verified by: EXT-AC-076, EXT-AC-077, EXT-AC-098, EXT-AC-145

`owner_document_sha256_v1` MUST execute these steps:

1. validate the `owner_document_ref` grammar, require `anchor-kind='document'`, require the anchor ID to equal `owner_document_schema_id`, and extract only the repository-relative path component;
2. resolve that path beneath the repository root without consulting or following an owner anchor, manifest anchor range, symlink, search result, or Markdown parser;
3. reject a symlink, non-regular file, absolute path, path traversal, backslash, NUL, unreadable file, UTF-8 BOM, invalid UTF-8, or any CR byte;
4. require every line ending to be LF and permit the file either to end in LF or to contain no final line terminator;
5. hash the exact accepted file bytes with SHA-256;
6. return lowercase 64-character hexadecimal.

Dependency validation MUST then compare that digest with both the dependency row and the owner contract manifest before any non-document anchor is resolved. This two-stage path-and-digest bootstrap avoids a manifest-resolution cycle.

The digest algorithm MUST NOT rewrite whitespace, normalize Unicode, alter the final LF state, or exclude front matter, comments, examples, or appendices. The dependency snapshot itself MUST serialize under `extension_registry_canonical_json_v1`; `extension_dependency_snapshot_sha256_v1` is the digest of that canonical byte form.

**Table 1-B. Required dependency declarations**

| `dependency_id` | Owner document | Imported interface | Required owner-fragment fact families |
| --- | --- | --- | --- |
| `core00` | `docs/spec/00_document_set_status_and_precedence.md` | Recognition, claimability, contract major, owner identity, adopted-document precedence, and runtime dependencies | `recognized_profile`, `runtime_dependency` |
| `core01` | `docs/spec/01_architecture_storage_and_view_contracts.md` | Discovery, Base route-reservation registry, route-family reservation and dispatch, authorized workspace availability, common jobs, `sheet_ref`, invalidation, import owner façades, deterministic cross-owner transaction publication, staged-object cleanup, backup, and portability orchestration | `route_family`, `workspace`, `public_schema` |
| `core02` | `docs/spec/02_domain_model_schema_and_history.md` | Core-record boundary, cross-owner source-state boundary, and incident-removal reservation | No descriptor fact; exact imported anchors remain required |
| `core03` | `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Extension workspace shell, startup fallback, lazy interaction, unknown-value handling, and client cleanup | No descriptor fact; exact imported anchors remain required |
| `core04` | `docs/spec/04_security_deployment_and_conformance.md` | Claim configuration, normalized profile configuration, secret references, authorization, application-process lease, extension limits and deadlines, publication admission, startup findings, egress trust boundary, readiness, and exit behavior | `claim_configuration` |
| `network_flow_activity` | `docs/network-flow-activity-nlspec.md` | Current adopted extension-resource, workspace, import-target, state, job, lifecycle, and portability specialization | Every named-profile fact allocated to Network Flow by Table 5-A |
| `testing_harness` | `docs/testing-harness-nlspec.md` | `cartulary.testing_harness.v2`; owner and verification registries, exact catalog selectors, runner and execution profiles, owner slices, retained row accounting and owner summaries, evidence-root auditing, generated-artifact policy, and drift | No descriptor fact; exact imported anchors remain required |
| `opentelemetry` | `docs/opentelemetry-instrumentation-nlspec.md` | Canonical claimed-profile telemetry representation | No descriptor fact; exact imported anchors remain required |
| `reporting` | `docs/reporting-subsystem-nlspec.md` | Snapshot and Reporting mode registry and participant boundary | No descriptor fact; exact imported anchors remain required |
| `report_composition` | `docs/report-composition-nlspec.md` | Report-composition ownership boundary | No descriptor fact; exact imported anchors remain required |

Table 1-B declares the required dependency identities and interface classes. The dependency snapshot carries the exact adopted version, content digest, stable anchor locators, schema IDs, and algorithm IDs. Human-readable section numbers, headings, line numbers, branch names, and repository timestamps MUST NOT substitute for those exact fields.

The `testing_harness` dependency MUST bind `doc_id='cartulary.testing_harness.v2'`, `conformance_profile_id='cartulary.testing_harness.current.v2'`, and `status='adopted/current'`. Its imported schema set MUST include the current v2 verification registry and contract, test-owner registry, test-family manifest, runner registry, test-slice plan and summary, test-evidence accounting, test-owner summary, evidence-root manifest, and evidence-audit summary schemas used by §27. Pre-v2 commands, ordinal execution identities, retired schemas, custom extension fixture-result interfaces, compatibility readers, and historical retained runs MUST NOT satisfy dependency validation, registry accounting, adoption, or release closure.

**EXT-REQ-003**
The Extensions Subsystem MUST NOT appear in extension discovery, MUST NOT have a `<profile_id>.claimed` key, and MUST NOT be claimable. Its Base Profile behavior MUST operate with every optional extension profile unclaimed.

Profiles: base
Verified by: EXT-AC-039, EXT-AC-070, EXT-AC-074

**EXT-REQ-004**
When this NLSpec conflicts with Core 00 through Core 04 outside the behavior allocated to it by Table 1-A, the conflict is a defect in this NLSpec. When two primary owners appear to conflict, the contradiction is a corpus defect; an implementation MUST NOT choose one side by local convention.

Profiles: base
Verified by: EXT-AC-001, EXT-AC-072

**EXT-REQ-005**
The artifact classes in Table 1-C MUST remain distinct. An artifact in one class MUST NOT acquire the authority of another class merely because it is generated, packaged, persisted, or consumed at runtime.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-005, EXT-AC-072, EXT-AC-076

**Table 1-C. Artifact authority classes**

| Artifact class | Examples | Authority rule |
| --- | --- | --- |
| Owner-authored normative input | Adopted owner document and separately stored `cartulary.extension_owner_fragment.v1` explicitly adopted by that document | Normative only for the exact facts allocated to that owner. |
| Generated normalized input | Dependency snapshot and owner-input registry | Derived from owner-authored input; cannot add or override behavior. |
| Generated runtime input | Descriptor, registry, registry-integrity object, conformance-manifest index | Runtime and conformance input; cannot override owner facts. |
| Build-time implementation declaration | `cartulary.extension_implementation_binding.v1` | Declares packaged support; cannot widen the descriptor or owner contract. |
| Durable runtime state | State metadata, migration ledger, job commit proof, cancellation observation | Authoritative only for the runtime fact explicitly represented by its schema. |
| Harness v2 evidence | Verification contracts, catalog rows, exact selectors, paired row-accounting and owner-summary shards, evidence-audit summaries, and retained target results | Derived evidence only; cannot define product behavior, make an incomplete owner contract complete, or activate Core 05 publication. |
| Non-normative support | Examples, paths, diagrams, operator guidance, implementation recipes | No implementation-conformance authority. |

# 2. Normative language and document discipline

**EXT-REQ-006**
The key words **MUST**, **MUST NOT**, and **MAY** are normative. **MUST** and **MUST NOT** define conformance requirements. **MAY** defines optional behavior only when the same requirement, table row, or immediately following paragraph defines omission behavior.

Profiles: base
Verified by: EXT-AC-001

**EXT-REQ-007**
The word `default` defines the required value or behavior when an optional member or configuration value is omitted and omission is valid. A default is observable behavior and is not implementation advice.

Profiles: base
Verified by: EXT-AC-007, EXT-AC-011, EXT-AC-029, EXT-AC-078

**EXT-REQ-008**
Object member names, schema identifiers, algorithm identifiers, profile IDs, capability IDs, contribution kinds, route-family strings, workspace keys, error codes, reason codes, lifecycle states, and other closed-vocabulary tokens MUST be compared as exact Unicode scalar-value sequences after decoding. No trimming, case folding, Unicode normalization, locale comparison, aliasing, or visible-label matching is allowed unless this NLSpec or the primary owner defines the exact operation.

Equality and ordering are separate contracts. A statement that values are equal MUST NOT imply an ordering comparator. Every array or output that is required to be sorted MUST name its comparator. A requirement that says only `sorted` without naming a comparator is incomplete and MUST fail specification conformance review.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-021, EXT-AC-040, EXT-AC-082

**EXT-REQ-009**
Every current producer schema and every canonical artifact schema defined by this NLSpec is closed. Unknown members, duplicate members, missing required members, wrong JSON type categories, and explicit JSON `null` where nullability is not declared MUST be rejected before the object is used.

A compatible consumer MAY accept unknown additive members only through a separately named decoder algorithm that explicitly defines that behavior. Omission behavior: without such a decoder, the consumer rejects the unknown member. Consumer tolerance MUST NOT be interpreted as producer permission to emit members absent from the current producer schema.

Profiles: base
Verified by: EXT-AC-008, EXT-AC-009, EXT-AC-012, EXT-AC-091

**EXT-REQ-010**
An optional source or request member whose omission has a default MUST be materialized before cross-field validation, canonical serialization, digest calculation, persistence, replay comparison, or canonical-output construction. The owning schema table MUST identify the source object in which omission is valid and the canonical object in which the materialized member is required.

Explicit JSON `null` MUST NOT be treated as omission unless the owning field contract explicitly permits null and defines its meaning.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-007, EXT-AC-012, EXT-AC-078

**EXT-REQ-011**
Every set-like array defined by this NLSpec MUST reject duplicate normalized values and MUST serialize in the ordering declared for that array. An implementation MUST NOT silently coalesce duplicates unless a requirement explicitly defines duplicate coalescing.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-021, EXT-AC-024, EXT-AC-082

**EXT-REQ-012**
Accepted normative text MUST NOT contain open delegation phrases such as `as appropriate`, `where needed`, `implementation-defined`, `equivalent`, `supported when available`, `owner-specific`, or `ordinary fallback` unless the same sentence or adjacent table names a closed schema, algorithm, owner contract, state matrix, or future-only boundary that makes the behavior decidable.

Profiles: base
Verified by: EXT-AC-001

**EXT-REQ-013**
Examples, rationale, research synthesis, diagrams, and implementation notes are non-normative. They MUST NOT introduce a value, default, error, interface, lifecycle rule, timeout, limit, exit code, or cleanup consequence absent from the normative requirements. A conflicting example is defective and MUST NOT override a requirement.

Profiles: base
Verified by: EXT-AC-001

**EXT-REQ-014**
`EXT-REQ-*`, `EXT-AC-*`, and `EXT-GATE-*` identifiers MUST remain unique and MUST NOT be reused after publication. Harness v2 verification, owner, family, and row identifiers follow their imported immutability and non-reuse contracts.

Profiles: base
Verified by: EXT-AC-002

JSON members declared as integers by this NLSpec MUST use a base-10 integer lexeme with no fraction, exponent, leading plus sign, or leading zero except `0`. A negative lexeme is valid only when the owning scalar contract permits negative values. A JSON number that is mathematically integral but uses fraction or exponent syntax is not an integer under this NLSpec.

Byte limits in this NLSpec are measured over UTF-8 bytes unless the owning field explicitly states another unit. Count limits include every decoded element before semantic deduplication; an implementation MUST NOT evade a count limit by silently coalescing invalid duplicates.

Implementation freedom is intentional only when a requirement states that the mechanism is not prescribed and every permitted mechanism produces the same observable bytes, identifiers, ordering, errors, timing boundary, persistence result, authorization result, security result, recovery result, and interoperability result. Silence at any such boundary MUST NOT be interpreted as implementation freedom.

# 3. Purpose and non-goals

**EXT-REQ-015**
The Extensions Subsystem MUST provide one governed framework through which statically packaged Cartulary capabilities can be recognized, claimed, discovered, admitted, versioned, integrated, disabled non-destructively, and verified without changing Base Profile semantics or creating arbitrary execution hooks.

Profiles: base
Verified by: EXT-AC-003, EXT-AC-039, EXT-AC-070

**EXT-REQ-016**
Extension executable code and browser assets MUST be packaged with the single Cartulary application deployable. A profile claim selects from packaged capabilities; it does not install code.

Profiles: base
Verified by: EXT-AC-010, EXT-AC-070

**EXT-REQ-017**
The Extensions Subsystem MUST preserve these properties:

- Base Profile conformance with all optional profiles unclaimed;
- deterministic profile and contribution identity;
- fail-closed startup admission for requested claims;
- non-destructive unclaim;
- typed owner interfaces instead of direct cross-owner storage writes;
- explicit public and persisted-state compatibility;
- default-deny external egress;
- tolerance for unknown future profiles and capabilities at compatible client boundaries;
- no mandatory extension work on the ordinary workbook capture hot path.

Profiles: base
Verified by: EXT-AC-013, EXT-AC-022, EXT-AC-023, EXT-AC-024, EXT-AC-036, EXT-AC-039, EXT-AC-067

**EXT-REQ-018**
The current revision MUST NOT provide any behavior in Table 3-A.

Profiles: base
Verified by: EXT-AC-038, EXT-AC-070, EXT-AC-074

**Table 3-A. Current non-goals**

| Non-goal | Required current behavior |
| --- | --- |
| Extension marketplace | No marketplace route, catalog, publisher account, rating, or package-distribution surface exists. |
| Runtime extension installation | No public or operator surface installs executable extension code or UI assets. |
| Runtime extension uninstall | No public or operator surface removes packaged extension code or extension-owned state. |
| Live claim mutation | No public route, WebSocket message, file watcher, or browser action enables, disables, or reloads a profile while the process is running. |
| Independent extension deployables | Extensions remain internal modules of the one application deployable. |
| Arbitrary callback bus | No generic before/after hook, middleware chain, reflection callback, database trigger registry, or lifecycle callback registry exists. |
| Arbitrary UI injection | Extensions MUST NOT inject arbitrary React components into Base surfaces or replace Base shell components. |
| Automatic Core promotion | Extension resources do not become Core records, views, or saved views automatically. |
| Implicit report/export inclusion | Claiming an extension does not include its data in snapshots, reports, releases, or portability bundles automatically. |
| Extension-specific auth bypass | A claim does not create a new session family, incident-role model, or `deployment_admin` incident-data bypass. |
| Extension microservices | No separate extension host, sidecar, remote execution service, or plugin process is required or admitted. |

**EXT-REQ-019**
A later independently distributed extension-package model requires a separate adopted NLSpec. §30 defines the minimum topics that future specification MUST close. Nothing in this revision reserves a compatible package format or trust model.

Profiles: base
Verified by: EXT-AC-070, EXT-AC-074

# 4. Concepts and identifiers

**EXT-REQ-020**
The terms in Table 4-A have the exact meanings defined in this NLSpec.

Profiles: base
Verified by: EXT-AC-001, EXT-AC-040

**Table 4-A. Concepts**

| Term | Definition |
| --- | --- |
| `extension profile` | A recognized, bounded optional Cartulary capability identified by one stable `profile_id`. |
| `recognized profile` | A profile identity present in the current Core 00 registry, whether or not it is claimable. |
| `claimable profile` | A recognized profile that Core 00 permits a deployment to claim under its current owner contract. |
| `claim request` | Effective startup configuration value `<profile_id>.claimed=true`. It is not proof of successful admission. |
| `claimed profile` | A claimable profile whose claim request passed every admission stage and is present in the canonical resolved claim set. |
| `unclaimed profile` | A claimable profile whose claim key is omitted or false. |
| `recognized unclaimable profile` | A recognized profile for which Core 00 reports `claimable=false`. |
| `owner fragment` | Separately stored, closed `cartulary.extension_owner_fragment.v1` object explicitly adopted by one owner document and containing only facts allocated to that owner. |
| `dependency snapshot` | Canonical digest-bound list of exact adopted owner documents, anchors, schemas, and algorithms imported by this NLSpec revision. |
| `owner-input registry` | Canonical joined collection of the dependency snapshot and every adopted owner fragment used to derive descriptors. |
| `descriptor source` | Generated intermediate object that permits omission only for explicitly defaultable members. |
| `extension-profile descriptor` | Generated closed canonical object containing every materialized descriptor member for one recognized profile. |
| `canonical extension registry` | Deterministic ordered collection containing exactly one valid canonical descriptor per recognized profile. |
| `registry-integrity object` | Generated closed object that binds the dependency snapshot, owner inputs, descriptors, registry, implementation bindings, schemas, and generator provenance. |
| `implementation binding` | Build-time closed object that declares the packaged behavior implemented for exactly one profile and descriptor digest. It is not a runtime-loaded callback or package. |
| `runtime dependency` | Exact required relation from one explicitly requested extension profile to another explicitly requested profile and contract major. |
| `capability` | Additive advertised behavior not required by the profile's baseline contract major. |
| `contribution` | Typed declaration that lets a claimed profile participate in one shared owner-controlled integration point. |
| `extension resource` | Durable resource owned by a named extension profile and not automatically represented by a Core record envelope. |
| `extension workspace` | Top-level incident workbook workspace identified by `extension_profile_id` and `workspace_key`, distinct from a Base tab, system view, saved view, or `view_schema`. |
| `contract major` | Positive integer governing public caller compatibility for one extension profile. |
| `document version` | Repository document revision using the exact SemVer subset in §14; not a public runtime compatibility field. |
| `state schema version` | Positive integer governing interpretation of extension-owned durable state. |
| `migration lineage` | Stable profile-scoped identity that prevents state from one incompatible migration family from being interpreted by another family. |
| `state-presence declaration` | Declarative shared-owner contract that identifies authoritative state families whose presence proves that profile state exists. |
| `migration ledger entry` | Durable proof that one immutable consecutive migration definition committed. |
| `job commit proof` | Durable proof, committed with the authoritative effect, that binds one job to its exact original terminal success. |
| `startup finding` | Closed secret-safe diagnostic item with one canonical path, reason code, exact generic message, and reason-specific details object. |
| `strict producer schema` | Current wire or artifact schema whose producer emits exactly its declared members. |
| `compatible consumer decoder` | Named decoder that validates known required members and safely ignores declared classes of unknown additive data. |
| `resolved claim set` | Canonical sorted set of profiles admitted as claimed for the running process. |
| `retirement` | Core 00 transition that makes a recognized profile unclaimable while preserving its reserved identity and retained state. |

**EXT-REQ-021**
Identifiers defined by this NLSpec MUST satisfy Table 4-B. Owner-defined identifiers MAY use a narrower grammar but MUST NOT exceed the common bound. Omission behavior: when an owner defines no narrower grammar, Table 4-B is the complete grammar.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-021, EXT-AC-040, EXT-AC-077

**Table 4-B. Identifier grammar and bounds**

| Identifier | Required grammar and bound |
| --- | --- |
| `profile_id` | ASCII; `^[a-z][a-z0-9_]{0,63}$`. |
| `owner_id` | ASCII; `^[a-z][a-z0-9_]{0,63}$`. |
| `workspace_key` | ASCII; `^[a-z][a-z0-9_]{0,63}$`. |
| Profile-local key | ASCII; `^[a-z][a-z0-9_]{0,63}$`. |
| `capability_id` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `contribution_id` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `owner_fragment_id` | `<owner_id>.<local_key>`; maximum 129 ASCII bytes. |
| `migration_lineage_id` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `migration_id` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `worker_kind` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `job_kind` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `panel_key` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `entry_key` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `target_kind` | `<profile_id>.<local_key>` unless an already adopted owner fragment declares one exact legacy token; maximum 129 ASCII bytes. |
| `resource_ref_kind` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| `participant_id` | `<profile_id>.<local_key>`; maximum 129 ASCII bytes. |
| Extension resource kind | `<profile_id>.<local_key>` unless an already adopted owner fragment declares one exact legacy token; maximum 129 ASCII bytes. |
| `final_commit_id` | Exact EXT-REQ-233 safe identifier; deployment-unique. |
| `contract_major` | JSON integer in `1..2147483647`. |
| State schema version | JSON integer in `1..2147483647`. |
| Claim key | Exactly `<profile_id>.claimed`. |
| Configuration namespace | Exactly `<profile_id>`. |
| Fully qualified profile configuration key | `<profile_id>` followed by one or more `.`-separated segments; each segment matches `[a-z][a-z0-9_]{0,63}`; maximum 256 ASCII bytes. |
| `conformance_manifest_id` | Exactly `<profile_id>.conformance.v<contract_major>`. |
| Public schema ID | ASCII `1..160` bytes matching `[A-Za-z0-9][A-Za-z0-9_.-]{0,159}`. |
| Owner-defined extension resource ID | Non-empty UTF-8 `1..256` bytes; the named owner MUST define a narrower exact scalar contract. |
| SHA-256 digest string | Exactly 64 lowercase ASCII hexadecimal characters. |

**EXT-REQ-175**
`owner_contract_ref`, `owner_document_ref`, and every owner locator used by this NLSpec MUST conform to `owner_locator_v1`:

```text
<repo-relative-posix-path>#<anchor-kind>:<anchor-id>
```

The closed `anchor-kind` values are defined by Table 4-C.

**Table 4-C. Owner locator anchor kinds**

| `anchor-kind` | Required target |
| --- | --- |
| `document` | The complete owner document. |
| `req` | One stable normative requirement ID. |
| `table` | One explicitly assigned stable table ID. |
| `schema` | One explicitly assigned schema ID. |
| `algorithm` | One explicitly assigned algorithm ID. |

The complete locator MUST contain `1..512` UTF-8 bytes. The path MUST be repository-relative POSIX syntax and MUST contain no leading slash, backslash, NUL, empty segment, `.` segment, `..` segment, percent escape, URI scheme, query, or fragment other than the single `#` delimiter in the grammar. `anchor-id` MUST contain `1..160` ASCII bytes and MUST match `[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}`. A heading text, section number, line number, source-control branch, mutable tag, display label, or first textual occurrence MUST NOT be an anchor identity.

`resolve_owner_locator_v1` MUST:

1. validate the locator grammar;
2. locate the exact dependency-snapshot row for the path;
3. validate that row's owner contract manifest under EXT-REQ-203;
4. require the path and owner-document digest to equal the manifest's `owner_document` object;
5. find exactly one `anchors[]` row whose `anchor_kind` and `anchor_id` equal the locator;
6. validate that row's byte range and `anchor_sha256` against the exact accepted owner-document bytes;
7. return the exact owner-document identity, anchor identity, byte interval, and anchor digest;
8. fail on zero matches, multiple matches, stale digest, wrong anchor kind, invalid range, path mismatch, or unresolved path without scanning Markdown headings, prose, code fences, search results, or nearby similarly named identifiers.

A whole-document `owner_document_ref` MUST use `#document:<owner_document_schema_id>`, and the matching manifest document anchor MUST cover the exact byte interval `[0, byte_length)`. Within this NLSpec, every normative caption of the exact form `Table <section>-<suffix>` explicitly assigns the stable table anchor ID `EXT-TABLE-<section>-<suffix>`, preserving the caption characters after replacing the space with `-`; for example, `Table 7-F` assigns `EXT-TABLE-7-F` and `Table 27-A1` assigns `EXT-TABLE-27-A1`. A locator MUST NOT target a companion owner-document anchor until that owner contract manifest assigns the anchor a stable ID. The visible caption does not independently declare an anchor.

Profiles: base
Verified by: EXT-AC-077, EXT-AC-098

**EXT-REQ-203**
Every dependency row MUST resolve to one closed canonical `cartulary.extension_owner_contract_manifest.v1` object containing exactly:

- `schema_id`, exactly `cartulary.extension_owner_contract_manifest.v1`;
- `owner_contract_manifest_id`;
- `owner_id`;
- `owner_document`;
- `anchors[]`;
- `owner_fragments[]`.

`owner_document` MUST contain exactly:

- `owner_document_ref`;
- `owner_document_schema_id`;
- `owner_document_version`;
- `owner_document_sha256`;
- `byte_length`.

`owner_document_ref` MUST be the whole-document locator used by the dependency row. `byte_length` MUST equal the exact accepted owner-document byte length and be a JSON integer in `1..67108864`.

Each `anchors[]` row MUST contain exactly:

- `anchor_kind`;
- `anchor_id`;
- `start_byte`;
- `end_byte`;
- `anchor_sha256`;
- `closure_categories[]`.

For `anchor_kind='req'`, `closure_categories[]` MUST contain `1..19` unique category tokens from Table 19-A, sorted by ascending UTF-8 bytes. For every other anchor kind, `closure_categories[]` MUST equal `[]`. Category assignment is owner-authored normative metadata used only to derive the closure catalog; it does not replace the requirement text or permit a profile owner to omit the requirement from closure.

`anchors[]` MUST contain `1..4096` rows. `(anchor_kind, anchor_id)` MUST be unique. `start_byte` and `end_byte` MUST be zero-based JSON integers satisfying `0 <= start_byte < end_byte <= owner_document.byte_length`. The interval is half-open `[start_byte, end_byte)`, MUST begin and end on UTF-8 scalar boundaries, and `anchor_sha256` MUST hash exactly those bytes. Exactly one `document` anchor MUST exist, its ID MUST equal `owner_document_schema_id`, and its interval MUST equal `[0, byte_length)`. Overlapping non-document anchor intervals are permitted and have no identity effect.

Anchor rows MUST sort first by this closed kind order and then by ascending UTF-8 bytes of `anchor_id`:

```text
document
req
table
schema
algorithm
```

Each `owner_fragments[]` row MUST contain exactly:

- `owner_fragment_id`;
- `owner_fragment_ref`;
- `owner_fragment_sha256`.

`owner_fragments[]` MUST contain `0..512` rows, reject duplicate fragment IDs and normalized paths, and sort by ascending UTF-8 bytes of `owner_fragment_id`. `owner_fragment_ref` MUST be a normalized repository-relative POSIX path to one regular file. Absolute paths, backslashes, NUL, empty segments, `.`, `..`, symlinks, non-regular files, and repository-root escape are invalid.

The manifest MUST serialize under `extension_registry_canonical_json_v1`, contain `1..1048576` bytes including the final LF, and have digest `extension_owner_contract_manifest_sha256_v1`. A manifest MUST NOT be embedded in its owner document. The owner document MAY name the stable manifest ID but MUST NOT embed the manifest digest. Omission behavior: when the owner document does not name the ID, the dependency-snapshot association remains complete and unchanged. The dependency snapshot binds the manifest digest, and the manifest binds the owner-document and fragment digests; this is the only current adoption association and introduces no digest cycle.

Profiles: base
Verified by: EXT-AC-098

**EXT-REQ-022**
Visible labels, icons, path strings other than declared route identities, package names, module names, database object names, React component names, row positions, array positions, and display order MUST NOT be used as profile, capability, contribution, workspace, migration, manifest, worker, job, or extension-resource identity.

Profiles: base
Verified by: EXT-AC-040, EXT-AC-059

**EXT-REQ-023**
A capability MUST be owned by exactly one profile, MUST use that profile's prefix, and MUST represent additive behavior whose availability can be determined from discovery. Behavior required by the profile's current contract major MUST NOT be redundantly advertised as a capability.

Profiles: base
Verified by: EXT-AC-017, EXT-AC-024, EXT-AC-029

# 5. Owner inputs and derived-artifact boundary

**EXT-REQ-024**
The canonical descriptor for one profile MUST be derived only from validated `cartulary.extension_owner_fragment.v1` objects joined through `cartulary.extension_owner_input_registry.v1`. A missing required owner fact is a registry error. The generator MUST NOT infer it from prose, code, routes, configuration, database schema, package structure, generated output, or runtime behavior.

Profiles: base
Verified by: EXT-AC-003, EXT-AC-005, EXT-AC-006, EXT-AC-072, EXT-AC-076

**Table 5-A. Sole machine-fact owner allocation**

| Fact kind or family | Sole owner-fragment producer | Supporting owner interaction |
| --- | --- | --- |
| `recognized_profile`, `runtime_dependency` | Core 00 | Other owners consume the exact identity and dependency result. |
| `route_family`, `workspace`, `public_schema` | Core 01 | Core 03 and named profile owners define behavior behind the registered identities but emit no duplicate fact. |
| `claim_configuration` | Core 04 | Named profile owners define their namespace-local value schemas but emit no duplicate claim fact. |
| `capability`, `state_ownership`, `migration_definition`, `worker_kind`, `job_kind`, `admission_validation`, `egress`, `portability`, `snapshot_reporting`, `contribution`, `conformance_manifest` | Named profile owner contract, including an explicit Core section when Core is that profile's owner | Shared Core, Incident Portability, and Reporting owners admit the typed interface through digest-bound anchors and MUST NOT emit a second fact. Harness v2 validates only derived verification routing and exact catalog coverage; it emits no descriptor fact and derives no behavior from owner prose. |

**EXT-REQ-025**
A profile owner contract MUST be either an adopted standalone NLSpec or an explicit adopted Core section. The Extensions Subsystem MUST NOT require a standalone document when Core already owns the profile-specific behavior completely.

Each owner document that contributes descriptor facts MUST have one validated `cartulary.extension_owner_contract_manifest.v1` whose `owner_fragments[]` is the complete adopted fragment set. An owner fragment is adopted if and only if its exact ID, normalized repository reference, and canonical digest appear in that set. Directory scanning, filename convention, document prose, matching `owner_document_sha256`, implementation code, or repository search MUST NOT add a fragment. An owner fragment MUST NOT be embedded inside the owner document because the fragment carries the independent digest of that document.

Profiles: base
Verified by: EXT-AC-006, EXT-AC-073, EXT-AC-076

**EXT-REQ-176**
`cartulary.extension_owner_fragment.v1` MUST be a closed canonical object containing exactly the members in Table 5-B.

Profiles: base
Verified by: EXT-AC-076

**Table 5-B. `cartulary.extension_owner_fragment.v1`**

| Member | Required rule |
| --- | --- |
| `schema_id` | Exactly `cartulary.extension_owner_fragment.v1`. |
| `owner_fragment_id` | Table 4-B; globally unique across the owner-input registry. |
| `owner_id` | Table 4-B; exact normative owner identity. |
| `owner_document_ref` | Whole-document `owner_locator_v1`. |
| `owner_document_schema_id` | Exact non-empty owner-document identifier, `1..160` ASCII bytes. |
| `owner_document_version` | Exact adopted owner version, `1..64` ASCII bytes; no default. |
| `owner_document_sha256` | Exact `owner_document_sha256_v1` digest. |
| `facts[]` | `1..4096` closed fact objects from Table 5-C. |

A canonical owner fragment MUST serialize under `extension_registry_canonical_json_v1` and contain `1..1048576` bytes including its required final LF. `extension_owner_fragment_sha256_v1` is the digest of that canonical byte form. The canonical owner-input registry MUST contain `0..512` owner fragments and `0..65536` normalized facts. An individual fragment MUST still contain `1..4096` facts; an empty fragment is invalid and MUST be omitted. Exceeding any bound MUST fail with `extension_registry_limit_exceeded`; truncation is forbidden.

Every fact object MUST contain exactly:

- `fact_kind`;
- `profile_id`;
- `owner_contract_ref`;
- the variant-specific members in Table 5-C.

`owner_contract_ref` MUST resolve inside the same owner document named by the fragment. The fragment MUST NOT claim a fact allocated to another owner by Table 5-A.

**Table 5-C. Closed owner-fact variants**

| `fact_kind` | Variant-specific exact members and contracts | Fact identity key |
| --- | --- | --- |
| `recognized_profile` | `claimable` Boolean; `contract_major` positive integer or `null`; `primary_owner_id` under Table 4-B; `primary_owner_contract_ref` under `owner_locator_v1` | `profile_id` |
| `runtime_dependency` | `dependency_profile_id` under `profile_id`; positive integer `required_contract_major` | `(profile_id, dependency_profile_id)` |
| `route_family` | `route_family` under `route_family_template_v1` | `(profile_id, route_family)` |
| `workspace` | `workspace_key` under Table 4-B | `(profile_id, workspace_key)` |
| `claim_configuration` | exact `claim_config_key='<profile_id>.claimed'`; `configuration_contract_ref` under `owner_locator_v1`; `configuration_contract_sha256` as a SHA-256 digest string | `profile_id` |
| `capability` | `capability_id` under Table 4-B; nullable `prerequisite_contract_ref` under `owner_locator_v1` | `capability_id` |
| `public_schema` | `public_schema_id` under Table 4-B | `public_schema_id` |
| `state_ownership` | `state_ownership` under Table 7-C | `profile_id` |
| `migration_definition` | one `cartulary.extension_migration_definition.v1` object | `migration_definition.migration_id` |
| `worker_kind` | `worker_kind` under Table 4-B | `worker_kind` |
| `job_kind` | one `cartulary.extension_job_kind_contract.v1` object | `job_kind_contract.job_kind` |
| `admission_validation` | one `cartulary.extension_admission_validation.v1` object | `profile_id` |
| `egress` | `egress_mode` exactly `none` or `owner_declared`; nullable `egress_contract_ref` under `owner_locator_v1` | `profile_id` |
| `portability` | `incident_portability_mode` under Table 7-D; nullable Table 4-B `participant_id`; nullable `blocking_predicate` under `cartulary.extension_state_blocking_predicate.v1` | `profile_id` |
| `snapshot_reporting` | `snapshot_reporting_mode` under Table 7-E; nullable Table 4-B `participant_id` | `profile_id` |
| `contribution` | one closed contribution object under §16 | `contribution.contribution_id` |
| `conformance_manifest` | nullable `conformance_manifest_id` under Table 4-B | `profile_id` |

The following variant nullability rules are exhaustive:

- `recognized_profile.contract_major` MUST be non-null when `claimable=true` and MUST be `null` when `claimable=false`;
- `recognized_profile.primary_owner_contract_ref` MUST resolve to the exact adopted primary owner document and MUST agree with `primary_owner_id`; it is required even when `claimable=false`;
- `capability.prerequisite_contract_ref` MAY be `null`; omission is invalid;
- `egress_contract_ref` MUST be non-null only when `egress_mode='owner_declared'` and MUST be `null` when `egress_mode='none'`;
- `portability.participant_id` MUST be non-null only for `participant` and MUST otherwise be `null`;
- `portability.blocking_predicate` MUST be non-null only for `blocked_when_present` and MUST otherwise be `null`;
- `snapshot_reporting.participant_id` MUST be non-null only for `participant` and MUST be `null` for `no_participation`;
- `admission_validation.preflight_algorithm_ref` and `post_migration_algorithm_ref` MAY be `null`; omission is invalid and `null` means no additional profile-owned algorithm for that phase;
- `admission_validation.dependency_probes[]` MAY be empty; `[]` means no external startup dependency probe;
- `conformance_manifest_id` MUST be non-null for a claimable profile and MAY be `null` for a recognized unclaimable profile.

`cartulary.extension_admission_validation.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_admission_validation.v1`;
- `preflight_algorithm_ref`, an algorithm `owner_locator_v1` or `null`;
- `post_migration_algorithm_ref`, an algorithm `owner_locator_v1` or `null`;
- `dependency_probes[]`, containing `0..16` closed probe objects.

Each `dependency_probes[]` item MUST contain exactly `probe_id` and `algorithm_ref`. `probe_id` MUST use the profile prefix under Table 4-B. `algorithm_ref` MUST be an algorithm `owner_locator_v1` in the same profile owner document. Probe objects MUST reject duplicate `probe_id` values and sort by ascending UTF-8 bytes of `probe_id`. A non-empty probe array requires `egress_mode='owner_declared'` and an exact Table 24-A startup-probe contract. `egress_mode='none'` requires `dependency_probes=[]`.

`cartulary.extension_migration_definition.v1`, carried by a `migration_definition` fact, MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_migration_definition.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `migration_id`;
- `from_state_version`;
- `to_state_version`;
- `apply_algorithm_id`;
- `apply_algorithm_ref`;
- `validation_algorithm_id`;
- `validation_algorithm_ref`.

`to_state_version` MUST equal `from_state_version + 1`. Every algorithm ID MUST satisfy the public schema-ID scalar contract. Each algorithm reference MUST use `anchor-kind='algorithm'` with an `anchor-id` exactly equal to its paired algorithm ID and MUST resolve in the same owner contract as the enclosing fact. The canonical migration-definition bytes are its canonical JSON bytes followed by one LF. `migration_definition_sha256_v1` is the lowercase SHA-256 digest of those bytes.

**Table 5-D. Fact multiplicity per recognized profile**

| Fact kind | Required count and condition |
| --- | --- |
| `recognized_profile` | Exactly `1`. |
| `claim_configuration` | Exactly `1`. |
| `state_ownership` | Exactly `1`. |
| `admission_validation` | Exactly `1`. |
| `egress` | Exactly `1`. |
| `portability` | Exactly `1`. |
| `snapshot_reporting` | Exactly `1`. |
| `conformance_manifest` | Exactly `1`; its ID follows claimability nullability. |
| `runtime_dependency` | `0..16`, unique by dependency profile. |
| `route_family` | `0..16`. |
| `workspace` | `0..16`. |
| `capability` | `0..64`. |
| `public_schema` | `0..128`. |
| `migration_definition` | `0..256`; MUST be `0` unless state ownership is `extension_versioned`. |
| `worker_kind` | `0..64`. |
| `job_kind` | `0..64`. |
| `contribution` | `0..64`. |

A count outside Table 5-D, a missing exactly-one fact, or two facts with the same identity key MUST fail owner-input derivation. Shared supporting owners MUST be represented by dependency-snapshot anchors or contract-closure locators, not by a duplicate machine fact.

**EXT-REQ-177**
`cartulary.extension_owner_input_registry.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_owner_input_registry.v1`;
- `dependency_snapshot`, one complete `cartulary.extension_dependency_snapshot.v1` object;
- `owner_contract_manifests[]`, containing exactly one validated manifest per dependency row;
- `owner_fragments[]`, containing every fragment adopted by those manifests and no other fragment.

`derive_extension_owner_input_registry_v1` MUST execute these steps in order:

1. validate the dependency snapshot and every imported owner-document digest;
2. decode and validate every owner contract manifest as strict canonical UTF-8 JSON;
3. require exact one-to-one parity between dependency rows and manifests;
4. derive the complete adopted fragment set only from manifest `owner_fragments[]` rows;
5. reject a missing fragment, extra fragment, duplicate fragment path, duplicate fragment ID, or fragment digest mismatch;
6. decode every owner fragment as strict UTF-8 JSON, reject BOM, duplicate members, trailing bytes, unknown members, and noncanonical bytes;
7. validate every fragment scalar, collection bound, owner allocation, owner-document digest, and same-document owner locator;
8. resolve every `owner_document_ref` and `owner_contract_ref` exactly once under `resolve_owner_locator_v1`;
9. derive `cartulary.extension_owner_fact_identity.v1` for every fact under EXT-REQ-205;
10. reject duplicate fact identities, including byte-identical duplicates;
11. reject a fact naming an unrecognized profile except while validating `recognized_profile` facts;
12. enforce every Table 5-D per-profile multiplicity and no profile identity derived from another fact kind;
13. order manifests by ascending UTF-8 bytes of `owner_contract_manifest_id`;
14. order fragments by ascending UTF-8 bytes of `owner_fragment_id`;
15. order facts within each fragment by `profile_id`, then `fact_kind`, then canonical fact-identity bytes, all ascending;
16. serialize the owner-input registry under `extension_registry_canonical_json_v1`;
17. return canonical bytes and the normalized in-memory owner-input registry.

`owner_contract_manifests[]` MUST remain the exact non-empty one-per-dependency set even when the recognized-profile set is empty. `owner_fragments[]` MAY be empty only when no validated manifest adopts a fragment; in the zero-profile case, every manifest MUST adopt zero descriptor-fact fragments. The canonical owner-input registry MUST contain `1..67108864` bytes. The same dependency snapshot, manifest bytes, and owner-fragment bytes MUST produce byte-identical output.

Profiles: base
Verified by: EXT-AC-076, EXT-AC-077, EXT-AC-098, EXT-AC-099

**EXT-REQ-204**
The complete adopted owner-fragment set MUST equal the set union of every validated manifest `owner_fragments[]` row. A fragment whose owner-document identity or digest does not match the manifest owner document is invalid. A matching digest without an adoption row does not adopt the fragment. A fragment adoption row without a matching regular file is invalid. The generator MUST reject both missing and extra fragments before decoding facts.

Profiles: base
Verified by: EXT-AC-098

**EXT-REQ-205**
`cartulary.extension_owner_fact_identity.v1` MUST be a closed discriminated object containing `fact_kind` and exactly the identity members in Table 5-E. An identity object MUST NOT contain `owner_contract_ref`, a display value, a nested source object, or a member absent from Table 5-E.

**Table 5-E. Canonical owner-fact identity objects**

| `fact_kind` | Exact identity members after `fact_kind` |
| --- | --- |
| `recognized_profile` | `profile_id` |
| `runtime_dependency` | `profile_id`, `dependency_profile_id` |
| `route_family` | `profile_id`, `route_family` |
| `workspace` | `profile_id`, `workspace_key` |
| `claim_configuration` | `profile_id` |
| `capability` | `capability_id` |
| `public_schema` | `public_schema_id` |
| `state_ownership` | `profile_id` |
| `migration_definition` | `migration_id` |
| `worker_kind` | `worker_kind` |
| `job_kind` | `job_kind` |
| `admission_validation` | `profile_id` |
| `egress` | `profile_id` |
| `portability` | `profile_id` |
| `snapshot_reporting` | `profile_id` |
| `contribution` | `contribution_id` |
| `conformance_manifest` | `profile_id` |

`derive_extension_owner_fact_identity_v1` MUST lift nested source identities such as `migration_definition.migration_id`, `job_kind_contract.job_kind`, and `contribution.contribution_id` into the exact top-level scalar members in Table 5-E. It MUST serialize the identity under `extension_registry_canonical_json_v1` without the final LF for in-memory ordering. Those exact bytes MUST be used for duplicate rejection, fact ordering, collision identity, and `identity_sha256`. No second identity representation is conformant.

Profiles: base
Verified by: EXT-AC-099

**EXT-REQ-206**
Contract major `1` permits zero recognized profiles. The empty case MUST satisfy all of these invariants:

- `recognized profiles=[]` requires zero normalized facts;
- `owner_contract_manifests[]` still contains exactly one validated manifest per dependency row, while `owner_fragments[]=[]`;
- a fragment with `facts=[]` is invalid and MUST be omitted;
- a synthetic profile, owner fragment, or placeholder fact MUST NOT be created to avoid an empty collection;
- the dependency snapshot remains non-empty because this NLSpec still imports its companion owners;
- the canonical profile registry bytes equal the canonical JSON object `{"profiles":[],"schema_id":"cartulary.extension_profile_registry.v1"}` followed by one LF under the existing member-ordering algorithm;
- the owner-input registry, integrity object, static accounting object, and Harness v2 owner evidence MUST each define and test their own empty-profile representation under their respective owners.

Profiles: base
Verified by: EXT-AC-100

**EXT-REQ-026**
The descriptor and registry generator MUST consume only explicit normalized owner facts from `cartulary.extension_owner_input_registry.v1`. Repository search results, Markdown tables, headings, prose, implementation package discovery, route enumeration from a running server, database introspection, generated contract scanning, and source-code reflection MAY be used as drift evidence but MUST NOT create, alter, default, or select a missing owner fact. Omission behavior: when these secondary checks are not used, owner-derived generation behavior is unchanged.

Profiles: base
Verified by: EXT-AC-005, EXT-AC-072, EXT-AC-076

**EXT-REQ-027**
Any owner fact required by a descriptor, dependency, state-version declaration, portability declaration, egress declaration, contribution declaration, conformance manifest, contract-closure row, or adoption gate that is absent, unresolved, or internally inconsistent MUST block the profile from becoming claimable or this NLSpec from becoming `adopted/current`, according to the owning gate.

Profiles: base
Verified by: EXT-AC-073, EXT-AC-075, EXT-AC-095

# 6. Recognition and adoption

**EXT-REQ-028**
Core 00 MUST be the only owner that can add, remove, recognize, make claimable, make unclaimable, or assign the current contract major of an extension profile. A packaged implementation, owner fragment, descriptor, route, configuration key, database table, or generated artifact MUST NOT create recognition or claimability.

Profiles: base
Verified by: EXT-AC-003, EXT-AC-005, EXT-AC-014

**EXT-REQ-029**
A recognized profile MUST have exactly one Core 00 `recognized_profile` owner fact. A claimable profile MUST additionally have:

- one non-null current `contract_major`;
- one adopted primary owner contract;
- a complete adopted owner-fragment set;
- an exact dependency-snapshot match;
- one generated valid canonical descriptor;
- one matching registry-integrity entry;
- one packaged implementation binding in every build that permits the profile to be claimed;
- one uniquely indexed complete conformance manifest;
- one complete generated contract-closure catalog and matching `contract_closure[]` resolution set;
- one complete profile configuration contract, even when it declares zero profile-local keys;
- one current job-kind contract per declared job kind and one current participant specialization contract per typed contribution;
- closed state-presence, physical-state binding, migration, final-state validation, worker, job, portability, reporting, egress, contribution, client-support, active Harness v2 verification contracts, and exact-selector catalog coverage as applicable.

Profiles: base
Verified by: EXT-AC-006, EXT-AC-010, EXT-AC-073, EXT-AC-076, EXT-AC-080, EXT-AC-081, EXT-AC-095

**EXT-REQ-030**
A recognized unclaimable profile MUST remain present in the canonical registry and public discovery. Its claim key MUST be omitted or false. A true claim request MUST fail before readiness with `extension_profile_not_claimable`.

Profiles: base
Verified by: EXT-AC-009, EXT-AC-014, EXT-AC-024

**EXT-REQ-031**
Removing a recognized profile or changing its stable `profile_id` is a breaking Base Profile change because it changes discovery and reserved dispatch. Such a change requires a new Core 00 revision that changes the current Base Profile discovery contract and MUST NOT be performed through a profile document patch.

Profiles: base
Verified by: EXT-AC-003, EXT-AC-072

**EXT-REQ-032**
At coordinated adoption, normalized owner inputs MUST resolve to the current-profile baselines in Tables 6-A and 6-B. These tables are generated or parity-checked human-readable renderings of owner-fragment facts. They MUST NOT be edited as a second profile registry and MUST NOT override a validated owner fragment.

A mismatch between either table and the canonical owner-input registry MUST fail generation and conformance accounting.

Profiles: base
Verified by: EXT-AC-003, EXT-AC-004, EXT-AC-011, EXT-AC-016, EXT-AC-024, EXT-AC-035, EXT-AC-064, EXT-AC-065, EXT-AC-067, EXT-AC-068, EXT-AC-072, EXT-AC-076

**Table 6-A. Current-profile adoption baseline**

| `profile_id` | Claimable | Major | Claim key | Reserved route families | Workspace keys | Required runtime dependency |
| --- | ---: | ---: | --- | --- | --- | --- |
| `enterprise_authentication` | `true` | `1` | `enterprise_authentication.claimed` | `/api/v1/auth/oidc`; `/api/v1/auth/providers`; `/api/v1/auth/saml`; `/api/v1/users/{user_id}/auth-bindings` | `[]` | none |
| `import` | `true` | `1` | `import.claimed` | `/api/v1/import-sessions` | `[]` | none |
| `incident_portability` | `true` | `1` | `incident_portability.claimed` | `/api/v1/incident-bundles` | `[]` | none |
| `network_flow_activity` | `true` | `2` | `network_flow_activity.claimed` | `/api/v1/incidents/{incident_id}/network-flow` | `["network_analysis"]` | `import`, major `1` |
| `reference_pack` | `true` | `1` | `reference_pack.claimed` | `/api/v1/reference-packs` | `[]` | none |
| `snapshot_reporting` | `true` | `1` | `snapshot_reporting.claimed` | `/api/v1/incidents/{incident_id}/report-compositions`; `/api/v1/releases`; `/api/v1/snapshots` | `[]` | none |

**Table 6-B. Current-profile descriptor classifications**

| `profile_id` | `state_ownership` | `egress_mode` | `incident_portability_mode` | `snapshot_reporting_mode` |
| --- | --- | --- | --- | --- |
| `enterprise_authentication` | `core_managed` | `owner_declared` | `no_authoritative_incident_state` | `no_participation` |
| `import` | `core_managed` | `none` | `no_authoritative_incident_state` | `no_participation` |
| `incident_portability` | `core_managed` | `none` | `no_authoritative_incident_state` | `no_participation` |
| `network_flow_activity` | `extension_versioned`, current `1`, minimum `1`, lineage `network_flow_activity.state_v1`, `empty_state_policy=allowed`, initialization `empty` | `none` | `blocked_when_present` | `no_participation` |
| `reference_pack` | `core_managed` | `owner_declared` | `no_authoritative_incident_state` | `no_participation` |
| `snapshot_reporting` | `core_managed` | `none` | `participant` | `participant` |

**EXT-REQ-033**
At coordinated adoption, every current profile's `capability_ids[]` and `prestage_config_keys[]` MUST be empty. A later additive capability or pre-stageable configuration key requires an owner contract, owner-fragment revision, descriptor revision, active v2 verification coverage with exact selectors, and public discovery or configuration support before emission is permitted.

Profiles: base
Verified by: EXT-AC-017, EXT-AC-024, EXT-AC-029

# 7. Extension-profile descriptor

**EXT-REQ-034**
For each recognized profile, the generator MUST construct exactly one in-memory source object conforming to `cartulary.extension_profile_descriptor_source.v1` and exactly one canonical descriptor conforming to `cartulary.extension_profile_descriptor.v1`. The source object exists only to make omission materialization explicit. It is not a repo-control artifact instance, runtime input, persisted object, traceability object, or digest input. The canonical descriptor is the only descriptor admitted into the canonical registry.

Profiles: base
Verified by: EXT-AC-005, EXT-AC-008, EXT-AC-009, EXT-AC-078, EXT-AC-101

**Table 7-A. `cartulary.extension_profile_descriptor_source.v1`**

| Member | Type | Source requiredness | Omission or constraint |
| --- | --- | --- | --- |
| `schema_id` | String | Required | Exactly `cartulary.extension_profile_descriptor_source.v1`. |
| `profile_id` | String | Required | Table 4-B; exact Core 00 identity. |
| `claimable` | Boolean | Required | Exact Core 00 value. |
| `contract_major` | Integer or `null` | Required | Non-null when `claimable=true`; exact Core 00 value. |
| `owner_contract_ref` | String or `null` | Required | Non-null when `claimable=true`. |
| `claim_config_key` | String | Required | Exactly `<profile_id>.claimed`; generated, not independently authored. |
| `route_families[]` | Array of strings | Optional | Omission materializes `[]`. |
| `workspace_keys[]` | Array of strings | Optional | Omission materializes `[]`. |
| `capability_ids[]` | Array of strings | Optional | Omission materializes `[]`. |
| `runtime_dependencies[]` | Array of dependency objects | Optional | Omission materializes `[]`. |
| `contributions[]` | Array of contribution declarations | Optional | Omission materializes `[]`. |
| `public_schema_ids[]` | Array of strings | Optional | Omission materializes `[]`. |
| `prestage_config_keys[]` | Array of strings | Optional | Omission materializes `[]`. |
| `state_ownership` | Closed object | Required | One variant from Table 7-C. |
| `admission_validation` | Closed object | Required | One canonical `cartulary.extension_admission_validation.v1` object. |
| `egress_mode` | String | Optional | Omission materializes `none`. |
| `incident_portability_mode` | String | Required | One token from Table 7-D. |
| `snapshot_reporting_mode` | String | Required | One token from Table 7-E. |
| `conformance_manifest_id` | String or `null` | Required | Non-null when `claimable=true`. |

Explicit JSON `null` is invalid for every optional source member. No required source member has an omission default.

**EXT-REQ-178**
`materialize_extension_descriptor_v1` MUST execute:

1. validate that the decoded source value is an object and contains no duplicate or unknown member;
2. validate every present member's JSON type category;
3. materialize every Table 7-A omission default;
4. replace `schema_id` with `cartulary.extension_profile_descriptor.v1`;
5. validate every scalar, collection, variant, limit, and cross-field invariant;
6. sort every set-like array under Table 7-F;
7. emit one closed canonical descriptor containing every Table 7-B member exactly once.

Default materialization MUST occur before canonical serialization, descriptor digest calculation, collision detection, or registry construction. Explicit `null` MUST never invoke a default.

Each descriptor scalar MUST have exactly one source fact allocated by Table 5-A, and each descriptor set member MUST have exactly one normalized source fact. `recognized_profile.primary_owner_contract_ref` is the sole source of the descriptor's primary owner contract. Generation MUST reject a missing scalar source, multiple scalar sources, duplicate set members even when their values agree, a stale owner locator, and any attempt to infer a descriptor fact from source code, package imports, routes, configuration behavior, filenames, tests, or prose. The materializer may apply only the explicit Table 7-A defaults; it MUST NOT use a default to hide an absent required owner fact.

Profiles: base
Verified by: EXT-AC-078, EXT-AC-146

**EXT-REQ-209**
A `cartulary.extension_profile_descriptor_source.v1` instance is an ephemeral derivation intermediate with no canonical bytes and no digest. It MUST NOT be written to the repository, packaged output, runtime storage, logs, telemetry, conformance bundles, traceability manifests, or drift-accounting inputs. Runtime code MUST NOT read it. Default-valued source members MAY be omitted only until `materialize_extension_descriptor_v1` executes; every canonical descriptor member is required afterward. A later revision that retains descriptor-source instances MUST introduce a new schema ID, byte form, digest algorithm, artifact class, compatibility action, verification-contract revision, and exact catalog coverage.

Profiles: base
Verified by: EXT-AC-101

**Table 7-B. `cartulary.extension_profile_descriptor.v1`**

| Member | Type | Canonical rule |
| --- | --- | --- |
| `schema_id` | String | Required; exactly `cartulary.extension_profile_descriptor.v1`. |
| `profile_id` | String | Required; Table 4-B. |
| `claimable` | Boolean | Required. |
| `contract_major` | Integer or `null` | Required; conditional nullability in EXT-REQ-039 and EXT-REQ-040. |
| `owner_contract_ref` | String or `null` | Required; conditional nullability in EXT-REQ-039 and EXT-REQ-040. |
| `claim_config_key` | String | Required; exactly `<profile_id>.claimed`. |
| `route_families[]` | Array | Required; Table 7-F. |
| `workspace_keys[]` | Array | Required; Table 7-F. |
| `capability_ids[]` | Array | Required; Table 7-F. |
| `runtime_dependencies[]` | Array | Required; Table 7-F. |
| `contributions[]` | Array | Required; Table 7-F. |
| `public_schema_ids[]` | Array | Required; Table 7-F. |
| `prestage_config_keys[]` | Array | Required; Table 7-F. |
| `state_ownership` | Closed object | Required; one Table 7-C variant. |
| `admission_validation` | Closed object | Required; one canonical `cartulary.extension_admission_validation.v1` object. |
| `egress_mode` | String | Required; exactly `none` or `owner_declared`. |
| `incident_portability_mode` | String | Required; one Table 7-D token. |
| `snapshot_reporting_mode` | String | Required; one Table 7-E token. |
| `conformance_manifest_id` | String or `null` | Required; conditional nullability in EXT-REQ-039 and EXT-REQ-040. |

**EXT-REQ-035**
A descriptor source or canonical descriptor MUST NOT contain secrets, resolved secret values, deployment-specific claim state, incident data, live resource counts, database table or column names, object-store keys, absolute filesystem paths, Go package names, React component names, executable callbacks, middleware, SQL, display-state caches, visible labels used as identity, remotely loadable code or asset locations, runtime timestamps, hostnames, process IDs, or build-directory names.

Profiles: base
Verified by: EXT-AC-008, EXT-AC-030, EXT-AC-069

**Table 7-C. `state_ownership` variants**

| `kind` | Exact members | Required meaning |
| --- | --- | --- |
| `none` | `kind` | The profile owns no durable state. |
| `core_managed` | `kind` | Durable state is owned and semantically versioned entirely by existing Core owners. |
| `extension_versioned` | `kind`, `current_state_version`, `minimum_migratable_state_version`, `migration_lineage_id`, `empty_state_policy`, `state_presence_contract_ref`, `initialization_definition_ref`, `initialization_definition_sha256`, `final_state_validation_algorithm_id`, `final_state_validation_algorithm_ref` | The profile owns durable state under §21. |

For `extension_versioned`:

- `1 <= minimum_migratable_state_version <= current_state_version <= 2147483647`;
- `migration_lineage_id` MUST satisfy Table 4-B and use the profile prefix;
- `empty_state_policy` MUST equal `allowed` or `forbidden`; omission and explicit `null` are invalid;
- `state_presence_contract_ref` MUST be an `owner_locator_v1` with `anchor-kind='schema'` that resolves to the profile's `cartulary.extension_state_presence_manifest.v1` declaration;
- `initialization_definition_ref` MUST be an `owner_locator_v1` with `anchor-kind='schema'` that resolves to the profile's `cartulary.extension_state_initialization_definition.v1` object;
- `initialization_definition_sha256` MUST equal the object's `extension_state_initialization_definition_sha256_v1` digest;
- `final_state_validation_algorithm_id` MUST satisfy the public schema-ID scalar contract;
- `final_state_validation_algorithm_ref` MUST use `anchor-kind='algorithm'`, have an anchor ID exactly equal to `final_state_validation_algorithm_id`, and resolve in the profile owner contract.

**Table 7-D. Incident portability modes**

| Token | Required meaning |
| --- | --- |
| `no_authoritative_incident_state` | The profile owns no authoritative incident-scoped state that a bundle must carry. Any detected authoritative state is an owner-integrity failure. |
| `participant` | The profile owner defines versioned logical export and import contracts consumed by Incident Portability. |
| `blocked_when_present` | Incident export fails when the exact declarative blocking-state predicate in the named profile owner evaluates true. |

**Table 7-E. Snapshot and Reporting modes**

| Token | Required meaning |
| --- | --- |
| `no_participation` | Snapshot and Reporting MUST NOT query or include the profile implicitly. The omission is intentional. |
| `participant` | The profile owner defines one typed participant and exact input, empty-state, validation, ordering, limit, error, and output behavior. |

**Table 7-F. Canonical descriptor collection contracts**

| Array | Uniqueness key | Comparator | Count bound | Empty-array meaning |
| --- | --- | --- | ---: | --- |
| `route_families[]` | Exact route-family string | Ascending UTF-8 bytes of the route-family string | `0..16` | No reserved route family. |
| `workspace_keys[]` | Exact workspace key | Ascending UTF-8 bytes of the key | `0..16` | No extension workspace. |
| `capability_ids[]` | Exact capability ID | Ascending UTF-8 bytes of the ID | `0..64` | No advertised additive capability. |
| `runtime_dependencies[]` | Dependency `profile_id` | Dependency profile ID, then `required_contract_major`, ascending | `0..16` | No runtime profile dependency. |
| `contributions[]` | `contribution_id` | Contribution ID, then contribution kind, ascending UTF-8 bytes | `0..64` | No contribution. |
| `public_schema_ids[]` | Exact schema ID | Ascending UTF-8 bytes of the schema ID | `0..128` | No profile-owned public schema. |
| `prestage_config_keys[]` | Exact fully qualified key | Ascending UTF-8 bytes of the key | `0..32` | No profile-local key other than the claim key is valid while inactive. |
| `websocket_invalidation.resource_kinds[]` | Exact resource-kind token | Ascending UTF-8 bytes of the token | `1..64` | Invalid; the contribution MUST be omitted instead. |

One canonical descriptor MUST contain `1..262144` bytes including its final LF. A route-family root MUST contain `1..512` ASCII bytes and at most 32 path segments. Exceeding a descriptor, array, route-byte, or route-segment limit MUST fail with `extension_registry_limit_exceeded`; truncation is forbidden.

**EXT-REQ-036**
A `runtime_dependencies[]` item MUST be a closed object containing exactly:

- `profile_id`;
- `required_contract_major`.

`profile_id` MUST identify another recognized profile. `required_contract_major` MUST be an exact positive integer, not a range.

Profiles: base
Verified by: EXT-AC-016, EXT-AC-018, EXT-AC-019, EXT-AC-083

**EXT-REQ-037**
`prestage_config_keys[]` MUST be derived as the exact sorted set of configuration-contract keys whose `inactive_policy='syntax_only'`. The array MUST NOT be independently authored in an owner fragment, descriptor source, implementation binding, or deployment configuration. Every item MUST be an exact fully qualified key inside the profile namespace, MUST NOT equal the claim key, and MUST resolve to one key row in the digest-bound `cartulary.extension_profile_configuration_contract.v1`. `[]` means no profile-local key other than the claim key is valid while the profile is unclaimed.

Profiles: base
Verified by: EXT-AC-012, EXT-AC-013, EXT-AC-087, EXT-AC-102

**EXT-REQ-038**
A `contributions[]` item MUST conform to one closed variant in §16. `contribution_id` MUST be globally unique and MUST use the owning profile prefix. A contribution declaration MUST describe only typed integration metadata; it MUST NOT carry executable behavior.

Descriptor cross-field validation MUST enforce:

- `route_families[]` equals the set of `http_route_family.route_family` values;
- `workspace_keys[]` equals the set of `incident_workspace.workspace_key` values;
- every `websocket_invalidation.resource_kinds[]` value has exactly one `extension_resource_kind` declaration;
- `incident_portability_mode='participant'` has exactly one `incident_portability_participant`; the other two portability modes have none;
- `snapshot_reporting_mode='participant'` has exactly one `snapshot_reporting_participant`; `no_participation` has none;
- `state_ownership.kind='none'` and `state_ownership.kind='core_managed'` have no `backup_restore_participant`; `state_ownership.kind='extension_versioned'` has exactly one;
- every `import_target`, `job_resource_ref_kind`, transaction participant, administration panel, and authentication entry target resolves exactly once in its shared owner registry;
- every contribution whose Table 16-A `binding_requirement` is not `none` appears in the implementation binding under EXT-REQ-182;
- every non-null admission algorithm reference resolves to one statically packaged algorithm whose identifier is declared by the implementation binding;
- `admission_validation.dependency_probes[]` is empty when `egress_mode='none'` and every declared probe is admitted by the exact owner egress contract when `egress_mode='owner_declared'`;
- no implementation binding supplies a contribution, validation algorithm, or dependency probe absent from the descriptor.

Profiles: base
Verified by: EXT-AC-021, EXT-AC-031, EXT-AC-080

**EXT-REQ-039**
A descriptor for `claimable=false` MUST have `contract_major=null`. `owner_contract_ref` and `conformance_manifest_id` MAY be `null`. Every contribution or capability not required solely to reserve identity MUST be absent from its arrays. Reserved route families and workspace identities MAY remain present when Core owners reserve them. Omission behavior: owner identities that are not reserved are emitted as empty arrays.

Profiles: base
Verified by: EXT-AC-009, EXT-AC-024

**EXT-REQ-040**
A descriptor for `claimable=true` MUST have non-null `contract_major`, `owner_contract_ref`, and `conformance_manifest_id`. `conformance_manifest_id` MUST equal `<profile_id>.conformance.v<contract_major>`. A valid implementation binding is not encoded in the descriptor and MUST be verified independently at startup.

Profiles: base
Verified by: EXT-AC-006, EXT-AC-010, EXT-AC-081

# 8. Canonical extension registry

**EXT-REQ-041**
The canonical extension registry MUST conform to `cartulary.extension_profile_registry.v1` and contain exactly:

```json
{
  "schema_id": "cartulary.extension_profile_registry.v1",
  "profiles": []
}
```

`profiles[]` MUST contain exactly one valid canonical descriptor for every Core 00 recognized profile and no other descriptor. It MUST contain `0..256` items and MUST be ordered by ascending UTF-8 bytes of `profile_id`.

Profiles: base
Verified by: EXT-AC-003, EXT-AC-004, EXT-AC-005, EXT-AC-082

**EXT-REQ-042**
`extension_registry_canonical_json_v1` MUST serialize every object explicitly assigned to it using these rules:

1. input values are limited to JSON object, array, string, integer, Boolean, and `null`;
2. duplicate object members are invalid;
3. object members sort by ascending UTF-8 bytes of the exact member name;
4. arrays preserve their contract-defined order;
5. integers serialize in base 10 with no leading plus sign, no leading zero except `0`, no fraction, and no exponent;
6. strings serialize in double quotes; quotation mark and reverse solidus use `\"` and `\\`; U+0008, U+0009, U+000A, U+000C, and U+000D use `\b`, `\t`, `\n`, `\f`, and `\r`; other U+0000 through U+001F scalars use lowercase `\u00xx`; solidus is not escaped; all other Unicode scalar values are emitted as UTF-8;
7. unpaired surrogate values are invalid;
8. no insignificant whitespace is emitted;
9. one canonical file or object byte form is the canonical JSON bytes followed by exactly one LF byte;
10. no BOM or trailing byte after the required LF is emitted or accepted.

`extension_dependency_snapshot_sha256_v1`, `extension_owner_contract_manifest_sha256_v1`, `extension_owner_fragment_sha256_v1`, `extension_descriptor_sha256_v1`, `extension_contribution_sha256_v1`, `extension_owner_input_registry_sha256_v1`, `extension_registry_sha256_v1`, `extension_implementation_binding_sha256_v1`, `extension_profile_configuration_contract_sha256_v1`, `extension_base_route_reservation_registry_sha256_v1`, `extension_client_support_registry_sha256_v1`, `client_asset_set_sha256_v1`, `extension_physical_state_binding_sha256_v1`, `extension_state_presence_manifest_sha256_v1`, `extension_state_initialization_definition_sha256_v1`, `extension_backup_binding_codec_sha256_v1`, `extension_job_kind_contract_sha256_v1`, `extension_participant_contract_sha256_v1`, `extension_validation_condition_registry_sha256_v1`, `extension_contract_closure_catalog_sha256_v1`, and `extension_registry_integrity_sha256_v1` are the lowercase SHA-256 hexadecimal digests of their respective canonical byte forms. Harness v2 semantic digests retain their imported `sha256:<64-lowercase-hex>` form and MUST NOT be converted into an extension canonical-artifact digest. No other digest input or newline convention is conformant for those identifiers.

This algorithm is a Cartulary-specific canonicalization algorithm. It MUST NOT be labeled or treated as another JSON canonicalization standard.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-007, EXT-AC-079

**EXT-REQ-043**
`derive_extension_registry_v1` MUST execute these steps in order:

1. load and validate the canonical `cartulary.extension_owner_input_registry.v1` bytes and digest;
2. require exactly one normalized `recognized_profile` fact per recognized profile;
3. join every other owner fact by exact `profile_id` and fact identity key;
4. reject a recognized profile missing a required fact;
5. reject any fact naming an unrecognized profile;
6. derive `claim_config_key` as `<profile_id>.claimed` and reject owner drift;
7. construct one `cartulary.extension_profile_descriptor_source.v1` object per recognized profile;
8. execute `materialize_extension_descriptor_v1`;
9. validate every canonical descriptor scalar, collection, variant, bound, locator, and cross-field invariant;
10. calculate every `extension_descriptor_sha256_v1` digest;
11. sort descriptors by ascending UTF-8 bytes of `profile_id`;
12. execute `validate_extension_registry_collisions_v1` from §13;
13. serialize `cartulary.extension_profile_registry.v1` under `extension_registry_canonical_json_v1`;
14. enforce the registry byte limit;
15. return the canonical bytes, normalized in-memory registry, owner-input digest, and descriptor digests.

The canonical registry MUST contain `1..67108864` bytes including its final LF.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-005, EXT-AC-021, EXT-AC-072, EXT-AC-076, EXT-AC-079

**EXT-REQ-044**
The same dependency snapshot, owner-fragment bytes, schemas, algorithms, and generator source bytes MUST produce byte-identical owner-input, descriptor, registry, and integrity bytes. A generated artifact governed by this section MUST NOT contain a generated timestamp, hostname, absolute checkout path, branch name, build directory, process ID, random value, file modification time, environment-dependent separator, or unordered map iteration output.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-079

**EXT-REQ-179**
Coordinated generation MUST emit one closed `cartulary.extension_registry_integrity.v1` object containing exactly the members in Table 8-A.

Profiles: base
Verified by: EXT-AC-079, EXT-AC-136

**Table 8-A. `cartulary.extension_registry_integrity.v1`**

| Member | Required rule |
| --- | --- |
| `schema_id` | Exactly `cartulary.extension_registry_integrity.v1`. |
| `canonicalization_algorithm_id` | Exactly `extension_registry_canonical_json_v1`. |
| `dependency_snapshot_sha256` | Canonical dependency-snapshot digest. |
| `owner_input_registry_sha256` | `extension_owner_input_registry_sha256_v1`. |
| `registry_schema_id` | Exactly `cartulary.extension_profile_registry.v1`. |
| `registry_sha256` | Exact `extension_registry_sha256_v1` digest. |
| `descriptor_digests[]` | One closed `{profile_id, descriptor_sha256}` object per recognized profile. |
| `owner_contract_manifest_digests[]` | One closed `{owner_contract_manifest_id, owner_contract_manifest_sha256}` object per dependency owner manifest. |
| `owner_fragment_digests[]` | One closed `{owner_fragment_id, owner_fragment_sha256}` object per owner fragment. |
| `implementation_binding_digests[]` | One closed `{profile_id, binding_sha256}` object per packaged binding. |
| `supporting_contract_artifact_digests[]` | One closed `{artifact_id, schema_id, artifact_sha256}` object per packaged static extension contract artifact not represented by another digest member. |
| `generator_id` | Stable ASCII `1..160` byte identifier. |
| `generator_sources[]` | `1..1024` closed `{source_ref, source_sha256}` objects covering every byte-affecting repo-control generator input not already represented by another digest in this object. |
| `generated_schema_digests[]` | One closed `{schema_id, schema_sha256}` object for every generated runtime schema used to validate or emit the dependency snapshot, owner inputs, descriptors, registry, integrity object, bindings, admission-validation contracts, state contracts, proof contracts, or packaged conformance artifacts. Documentation traceability and Harness v2 schemas are excluded. |

Each `generator_sources[]` item MUST contain exactly `source_ref` and `source_sha256`. `source_ref` MUST be a normalized repository-relative POSIX path of `1..512` UTF-8 bytes; it MUST reject an absolute path, backslash, NUL, empty segment, `.` segment, `..` segment, symlink, non-regular file, or path outside the repository root. `source_sha256` MUST be the lowercase SHA-256 digest of the exact regular-file bytes. The array MUST include every generator source file, template, schema attachment, static mapping, and lock or manifest file whose bytes the generator reads and whose effect is not already bound by `dependency_snapshot_sha256`, `owner_input_registry_sha256`, an owner-fragment digest, a descriptor digest, an implementation-binding digest, or `generated_schema_digests[]`. A generator MUST NOT read an undeclared byte-affecting repo-control file, network resource, clock, random source, host setting, or environment value.

Each `supporting_contract_artifact_digests[]` item MUST contain exactly `artifact_id`, `schema_id`, and `artifact_sha256`. It MUST include every profile configuration contract, Base route-reservation registry, client support registry, client asset-set manifest, physical-state binding, state-presence manifest, state-initialization definition, backup-binding codec, job-kind contract, participant specialization contract, validation-condition registry, contract-closure catalog, conformance manifest, and conformance-manifest index required by the recognized profile set. It MUST NOT contain a deployment-generated profile configuration view, runtime state, Harness verification or catalog input, clause-traceability object, registry-accounting object, owner result, retained result, or other run evidence. Duplicate artifact IDs are invalid.

Each `generated_schema_digests[]` item MUST contain exactly `schema_id` and `schema_sha256`. The schema artifact MUST be a closed canonical JSON object serialized under `extension_registry_canonical_json_v1`; `schema_sha256` MUST be the lowercase SHA-256 digest of that canonical byte form. Duplicate schema IDs are invalid. Each array MUST reject duplicates and sort by its first identity member, then digest, using ascending UTF-8 bytes. `owner_contract_manifest_digests[]` MUST sort by manifest ID; `supporting_contract_artifact_digests[]` MUST sort by `artifact_id`, then `schema_id`, then digest; `generator_sources[]` MUST sort by `source_ref`, then `source_sha256`; `generated_schema_digests[]` MUST sort by `schema_id`, then `schema_sha256`. The integrity object MUST contain `1..8388608` bytes including its final LF. It MUST contain no timestamp or environment-dependent path.

Generation and drift verification MUST resolve every `generator_sources[]` path inside the repository root and verify its exact regular-file bytes against `source_sha256`. The build MUST package exactly one canonical dependency snapshot, owner-input registry, canonical extension registry, registry-integrity object, set of implementation bindings, and every canonical runtime supporting artifact named by the integrity object. The build MUST bind the expected `extension_registry_integrity_sha256_v1` digest into the application executable or another packaged application artifact whose bytes cannot be changed independently of that executable. The embedding mechanism is not prescribed; the expected digest's immutability and runtime comparison are required.

**EXT-REQ-180**
Before claim resolution, runtime admission MUST:

1. enforce the byte ceiling for every packaged extension contract artifact before unbounded allocation;
2. reject an absent artifact, invalid UTF-8, BOM, invalid JSON, duplicate member, unknown member, wrong type, missing member, explicit invalid `null`, noncanonical object-member order, noncanonical integer, missing final LF, extra LF, or any trailing byte;
3. decode and validate the dependency snapshot, owner-input registry, canonical registry, registry-integrity object, and every implementation binding;
4. reserialize each object and require byte identity with its packaged bytes;
5. verify the build-bound integrity-object digest;
6. verify the dependency-snapshot, owner-contract-manifest, owner-input, owner-fragment, descriptor, registry, binding, supporting-contract-artifact, and schema digest and exact packaged identity set;
7. require no extra or missing recognized profile, descriptor, owner contract manifest, owner fragment, binding digest, supporting contract artifact, or generated-schema digest entry;
8. execute all registry cross-field and collision validation;
9. fail before claim-key processing, listener startup, worker startup, state mutation, or external egress on any discrepancy.

Semantically equivalent but noncanonical JSON MUST fail. Runtime MUST NOT silently recanonicalize packaged bytes and continue. Runtime MUST validate each `generator_sources[]` row structurally, including its path grammar and digest grammar, but MUST NOT resolve the path, require the source bytes, or require a repository checkout. Generator sources are generation and drift provenance inputs, not runtime dependencies.

A generated or built artifact set is stale if any owner-document version or digest, owner-contract-manifest identity or digest, imported anchor, imported schema, algorithm, or artifact ID, owner-fragment digest, dependency-snapshot digest, owner-input digest, descriptor digest, registry digest, binding digest, supporting-contract-artifact identity or digest, generator source set or digest, generated schema ID or digest, or build-bound integrity digest differs from the generation inputs. Runtime staleness is determined only from packaged artifact bytes, their exact packaged identity sets and digests, and the embedded root integrity-object digest.

Every failure in this requirement MUST use `extension_registry_invalid`, except a declared byte or count ceiling uses `extension_registry_limit_exceeded`.

Profiles: base
Verified by: EXT-AC-005, EXT-AC-079, EXT-AC-136

**EXT-REQ-045**
The runtime MUST complete EXT-REQ-180 before evaluating claim keys. A missing, malformed, stale, extra, noncanonical, unmatched, or internally inconsistent registry artifact MUST prevent readiness.

Profiles: base
Verified by: EXT-AC-005, EXT-AC-013, EXT-AC-079

**EXT-REQ-046**
A generated dependency snapshot, owner-input registry, descriptor, registry, integrity object, schema, manifest index, or accounting object MUST NOT be hand-edited. A source-control change that modifies generated bytes without the corresponding owner-input or generator change and successful regeneration MUST fail drift validation.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-072, EXT-AC-076

The required canonicalization golden set MUST cover, at minimum, non-ASCII member names, escape forms, empty objects, empty arrays, `null`, minimum and maximum admitted integers, invalid surrogate input, BOM, missing LF, extra LF, duplicate members, and noncanonical member ordering. Every supported implementation language MUST produce the same expected bytes for every valid vector and the same accept/reject classification for every invalid vector.

# 9. Claim configuration and runtime state

**EXT-REQ-047**
Core 04 MUST adopt one generic claim key for every recognized profile. The key MUST be exactly `<profile_id>.claimed`, have Boolean type, default to `false` when omitted, and reject explicit `null` and every non-Boolean value.

Profiles: base
Verified by: EXT-AC-005, EXT-AC-011, EXT-AC-012

**EXT-REQ-048**
A claim-key value of `true` is a claim request. It MUST NOT be reported as `claimed=true` until every admission stage in §10 succeeds. A claim-key value of `false` or omission means inactive.

Profiles: base
Verified by: EXT-AC-011, EXT-AC-013, EXT-AC-024

**EXT-REQ-049**
Claim keys MUST be evaluated only during process startup. The current revision defines no file watching, runtime reload, public route, operator command, WebSocket message, or browser action that changes the resolved claim set of a running process.

Profiles: base
Verified by: EXT-AC-013, EXT-AC-038, EXT-AC-070

**EXT-REQ-050**
A profile-local configuration key other than the claim key is invalid while a claimable profile is unclaimed unless its digest-bound configuration-contract row has `inactive_policy='syntax_only'`. A profile-local key for a recognized unclaimable profile is governed by the retired-key rules in EXT-REQ-208 and Table 21-B. An invalid unclaimed key MUST fail with `extension_config_without_claim`; the implementation MUST NOT ignore or use the value to start profile work.

Profiles: base
Verified by: EXT-AC-012, EXT-AC-013, EXT-AC-087, EXT-AC-102

**EXT-REQ-207**
Every `claim_configuration` fact MUST resolve to one closed canonical `cartulary.extension_profile_configuration_contract.v1` object containing exactly:

- `schema_id`, exactly `cartulary.extension_profile_configuration_contract.v1`;
- `profile_id`;
- `configuration_contract_id`;
- `configuration_contract_major`;
- `namespace_schema_id`;
- `keys[]`.

`configuration_contract_id` MUST equal `<profile_id>.configuration.v<configuration_contract_major>`. `configuration_contract_major` MUST be a JSON integer in `1..2147483647`. `namespace_schema_id` MUST satisfy the public schema-ID scalar contract. The contract's canonical digest MUST equal the `configuration_contract_sha256` in the owner fact, and `configuration_contract_ref` MUST resolve to the exact schema declaration under `owner_locator_v1`.

Each `keys[]` row MUST contain exactly:

- `key`;
- `value_schema_ref`;
- `inactive_value_schema_ref`;
- `omission_policy`;
- `inactive_policy`;
- `resolution_kind`;
- `diagnostic_policy`.

`key` MUST be one fully qualified key inside the profile namespace and MUST NOT equal the claim key. `value_schema_ref` MUST use `anchor-kind='schema'` and resolve in the profile owner contract. `inactive_policy` MUST equal `forbidden` or `syntax_only`. `inactive_value_schema_ref` MUST be a non-null `owner_locator_v1` with `anchor-kind='schema'` exactly when `inactive_policy='syntax_only'` and MUST be `null` when `inactive_policy='forbidden'`. The inactive schema vocabulary is closed to JSON type, object-member grammar, required presence of explicitly structural members, scalar grammar, collection length, string byte length, and nesting depth; it MUST NOT declare a default, semantic format, external reference existence, value substitution, conditional execution, profile algorithm, view, secret, file, trust material, DNS, endpoint, or egress behavior. `resolution_kind` MUST equal `plain`, `secret_ref`, `regular_file_ref`, or `trust_material_ref`. `diagnostic_policy` MUST equal `name_only` or `safe_value`.

`omission_policy` MUST be exactly one closed variant:

```json
{"kind":"required"}
```

```json
{"kind":"absent"}
```

```json
{"kind":"default","value":<schema-valid JSON value>}
```

The meanings are exact:

- `required`: omission is `missing_required_key`;
- `absent`: omission produces no normalized key/value row and no implied value;
- `default`: omission materializes the declared value before cross-field validation.

Explicit JSON `null` is invalid for every profile-local configuration key in contract major `1`; an optional value is omitted rather than represented by `null`. A default value MUST validate under `value_schema_ref`, contain at most `16384` canonical bytes, have nesting depth at most `16`, and contain no secret, resolved reference, absolute path, environment-specific endpoint, credential, or trust-store bytes.

`keys[]` MUST contain `0..256` rows, reject duplicate keys, and sort by ascending UTF-8 bytes of `key`. The count of rows with `inactive_policy='syntax_only'` MUST be `0..32`. The complete contract MUST contain `1..1048576` canonical bytes. A configuration-contract major change, key removal, key rename, changed default, changed omission, changed explicit-null behavior, changed value schema, changed inactive policy, changed resolution kind, or changed diagnostic policy is a public configuration compatibility change and MUST follow Table 14-B.

Profiles: base
Verified by: EXT-AC-102

**EXT-REQ-208**
Core 04 MUST normalize a claimed profile namespace into one closed `cartulary.extension_profile_configuration_view.v1` object containing exactly:

- `schema_id`, exactly `cartulary.extension_profile_configuration_view.v1`;
- `profile_id`;
- `configuration_contract_sha256`;
- `values[]`.

Each `values[]` row MUST contain exactly `key`, `source`, and `value`. `source` MUST equal `explicit` or `default`. Rows MUST reject duplicate keys and sort by ascending UTF-8 bytes of `key`. A key with `omission_policy.kind='absent'` and no explicit value MUST have no row. References remain normalized reference objects at this boundary; raw secret values, file bytes, trust-store bytes, and external connection results are forbidden.

Core 04 processing MUST execute in this order:

1. parse the deployment configuration and overlays;
2. reject unknown top-level and profile-local keys;
3. materialize the claim key default;
4. load and digest-validate the exact configuration contract;
5. for a claimed profile, apply omission policies, validate every value schema, classify references, and construct the complete normalized view;
6. for an unclaimed profile, reject every `forbidden` key and validate each explicitly present `syntax_only` value only against its inert `inactive_value_schema_ref`, without applying defaults or omission requirements, constructing a configuration view, retaining the value, resolving a reference or secret, reading a file, loading trust material, opening a connection, invoking profile code, or performing egress;
7. for a recognized unclaimable profile, recognize keys only from the retained last-adopted configuration-contract digest, apply each row's `forbidden` or `syntax_only` policy exactly, and keep every syntax-only value inert without semantic value validation, reference resolution, logging, diagnostics containing the value, profile code, or egress.

Required failure classes are `unknown_key`, `inactive_key_forbidden`, `explicit_null_not_allowed`, `missing_required_key`, `value_schema_mismatch`, `configuration_contract_digest_mismatch`, and `configuration_contract_major_unsupported`. Their exact diagnostic mapping MUST come from EXT-REQ-224.

The only inactive policies are `forbidden` and `syntax_only`. `syntax_only` is applied only to an explicitly present value, validates only the inert vocabulary above, and discards the accepted value before the next startup phase. Omission is accepted without materializing a required-field error or default. It MUST NOT construct any normalized profile configuration view; retain, log, or expose the value; perform secret resolution, file access, trust loading, reference-existence checks, connection validation, DNS, egress, or profile-code invocation. Local availability checks begin only for a claimed profile in Stage 2.

Profiles: base
Verified by: EXT-AC-102, EXT-AC-103, EXT-AC-129, EXT-AC-150

**EXT-REQ-051**
The runtime extension state model is closed to the three states in Table 9-A.

Profiles: base
Verified by: EXT-AC-009, EXT-AC-011, EXT-AC-024, EXT-AC-054

**Table 9-A. Runtime profile states**

| State | Core claimability | Effective claim key | Required behavior |
| --- | ---: | --- | --- |
| `recognized_unclaimable` | `false` | omitted or `false` | Reserved identity remains discoverable; routes remain unavailable; `true` fails startup; inactive behavior follows Table 21-B. |
| `unclaimed` | `true` | omitted or `false` | Reserved routes use Core unclaimed dispatch; workspaces, actions, capabilities, owner workers, and owner semantic behavior are unavailable; authoritative state is retained. |
| `claimed` | `true` | `true` | Complete startup admission succeeded; owner routes proceed to ordinary authorization. |

**EXT-REQ-052**
The current revision MUST NOT expose or internally treat as a served state any `degraded_claim`, `claimed_but_broken`, `partially_claimed`, `loading_claim`, `disabling`, or `reloading` state. A requested claim that cannot become `claimed` MUST block readiness.

Profiles: base
Verified by: EXT-AC-013, EXT-AC-054

**EXT-REQ-053**
The resolved claim set MUST contain exactly the `profile_id` values in state `claimed`, sorted by ascending UTF-8 bytes. It MUST be immutable for the lifetime of the running process and MUST have one `extension_registry_sha256_v1` and one resolved-claim-set digest bound at publication time.

The resolved-claim-set digest input MUST be the canonical JSON object `{ "profile_ids": [...] }` under `extension_registry_canonical_json_v1`, with `profile_ids[]` ordered by ascending UTF-8 bytes. `extension_resolved_claim_set_sha256_v1` is the SHA-256 digest of that canonical byte form.

Profiles: base
Verified by: EXT-AC-024, EXT-AC-054, EXT-AC-071, EXT-AC-090

**EXT-REQ-184**
The static and configured limits in Tables 9-B and 9-C are closed for contract major `1`. An implementation MUST NOT widen a static ceiling or use a configuration value outside its declared range. Exceeding a limit MUST fail before partial admission or mutation; truncation, sampling, silent dropping, and partial registry construction are forbidden.

Profiles: base
Verified by: EXT-AC-082

**Table 9-B. Static extension limits**

| Quantity | Bound | Required overflow reason |
| --- | ---: | --- |
| Recognized profiles | `0..256` | `extension_registry_limit_exceeded` |
| Owner fragments | `0..512` | `extension_registry_limit_exceeded` |
| Normalized owner facts | `0..65536` | `extension_registry_limit_exceeded` |
| Canonical owner contract manifest bytes | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical owner fragment bytes | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical profile configuration contract bytes | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical client support registry bytes | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical client asset-set manifest bytes | `1..16777216` | `extension_registry_limit_exceeded` |
| Ephemeral publication plan or component bytes | `1..16777216` | `extension_registry_limit_exceeded` |
| Canonical Base route-reservation registry bytes | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical physical-state binding bytes per profile | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical state-presence manifest bytes per profile | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical state-initialization definition bytes per profile | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical backup-binding codec bytes per binding | `1..1048576` | `extension_registry_limit_exceeded` |
| Canonical contract-closure catalog bytes per profile | `1..33554432` | `extension_registry_limit_exceeded` |
| Static supporting contract artifacts | `0..65536` | `extension_registry_limit_exceeded` |
| Canonical validation-condition registry bytes | `1..16777216` | `extension_registry_limit_exceeded` |
| Canonical descriptor bytes | `1..262144` | `extension_registry_limit_exceeded` |
| Canonical owner-input registry bytes | `1..67108864` | `extension_registry_limit_exceeded` |
| Canonical extension registry bytes | `1..67108864` | `extension_registry_limit_exceeded` |
| Canonical registry-integrity bytes | `1..8388608` | `extension_registry_limit_exceeded` |
| Canonical implementation-binding bytes | `1..1048576` | `extension_registry_limit_exceeded` |
| Migration definitions per profile | `0..256` | `extension_migration_path_too_long` |
| Startup findings per validation phase | `0..4096` | `extension_diagnostic_overflow` |
| Conformance accounting findings | `0..4096` | `extension_accounting_overflow` |
| Canonical clause-traceability rows | `0..65536` | `extension_registry_limit_exceeded` |
| Canonical clause-traceability bytes | `1..67108864` | `extension_registry_limit_exceeded` |

**Table 9-C. Core 04 extension configuration keys**

| Key | Type | Default | Valid range | Required expiration or overflow result |
| --- | --- | ---: | ---: | --- |
| `timeouts.extensions.migration_lock_seconds` | Integer | `30` | `1..300` | `extension_migration_lock_timeout` before mutation. |
| `timeouts.extensions.migration_step_seconds` | Integer | `900` | `1..3600` | Roll back the active uncommitted step when the outcome is determinate; `extension_migration_timeout`. |
| `timeouts.extensions.profile_migration_seconds` | Integer | `3600` | `60..14400` | Start no later step after expiry; `extension_migration_timeout`. |
| `timeouts.extensions.validation_seconds` | Integer | `60` | `1..600` | Stop the active profile validation algorithm and fail with `extension_admission_validation_failed`. |
| `timeouts.extensions.dependency_probe_seconds` | Integer | `30` | `1..300` | Stop the active read-only dependency probe and fail with `extension_dependency_probe_failed`. |
| `timeouts.extensions.reconciliation_seconds` | Integer | `300` | `1..3600` | Preserve nonterminal state and fail with `extension_reconciliation_timeout`. |
| `timeouts.extensions.shutdown_drain_seconds` | Integer | `30` | `1..300` | Force remaining connection and worker termination at expiry. |
| `timeouts.extensions.process_lease_acquire_seconds` | Integer | `30` | `1..300` | Fail before Stage 1 with `extension_application_process_active`; perform no profile mutation. |
| `timeouts.extensions.process_lease_loss_detection_seconds` | Integer | `5` | `1..30` | Enter fatal integrity shutdown after confirmed lease loss. |
| `timeouts.extensions.publication_seconds` | Integer | `30` | `1..300` | Keep the admission gate closed and fail startup with `extension_publication_failed`. |
| `timeouts.extensions.transaction_participant_seconds` | Integer | `30` | `1..300` | Roll back when absence of commit is proven; otherwise apply the final-commit outcome rules. |
| `timeouts.extensions.cancellation_grace_seconds` | Integer | `2` | `0..30` | Terminate the process when a canceled operation remains active after grace. |
| `intervals.extensions.staged_object_sweep_seconds` | Integer | `300` | `30..3600` | Begin the next cleanup sweep; logical inaccessibility does not depend on the sweep. |
| `limits.extensions.staged_object_cleanup_batch` | Integer | `1000` | `1..10000` | Process at most the configured ordered candidates per batch. |
| `timeouts.extensions.staged_object_cleanup_seconds` | Integer | `300` | `30..3600` | Startup remains not-ready until its captured cutoff set is drained or cleanup fails. |
| `timeouts.extensions.portability_participant_seconds` | Integer | `300` | `1..3600` | Cancel and fail the portability operation without publishing a bundle or target state. |
| `timeouts.extensions.snapshot_reporting_participant_seconds` | Integer | `300` | `1..3600` | Cancel and fail the snapshot/reporting operation without publishing participant output. |
| `timeouts.extensions.backup_restore_participant_seconds` | Integer | `900` | `1..7200` | Cancel and fail the active participant operation under the backup/restore owner; an indeterminate mutation is fatal. |
| `limits.extensions.max_nonterminal_jobs_per_profile` | Integer | `100000` | `1..1000000` | Fail before reconciliation mutation with `extension_reconciliation_limit_exceeded`. |

Every omitted Table 9-C key MUST materialize its default before startup validation. Explicit `null`, Boolean, string, fraction, exponent-form number, array, object, negative value, zero where excluded, or out-of-range integer MUST fail the Core 04 deployment-configuration envelope. The configured nonterminal-job limit MUST NOT exceed `1000000` and MUST be enforced by fetching at most `limit + 1` ordered rows. When the extra row exists, diagnostic `actual` MUST equal `limit + 1`; the implementation MUST NOT scan the remaining collection merely to calculate a larger count.

**EXT-REQ-213**
Contract major `1` permits exactly one active Cartulary application process to serve one deployment. Multiple threads, goroutines, listeners, and workers inside that process are permitted. Concurrent serving application processes, mixed-build rolling upgrades, and active-active replicas are future-only.

Before Stage 1, Core 04 MUST acquire one deployment-global, crash-released `application_process_lease`. The lease MUST be exclusive, remain held for the full serving-process lifetime, and be released by process exit or loss of the underlying lease session. Failure to acquire it within `timeouts.extensions.process_lease_acquire_seconds` MUST perform no migration or profile mutation, start no listener or job dequeuer, emit startup reason `extension_application_process_active`, and exit with code `2`.

The lease state machine is exactly:

```text
unacquired -> acquiring -> held -> uncertain -> held
                                      \-> lost
held -> released
```

Acquisition begins only from `unacquired`; explicit release is permitted only from `held` during orderly process shutdown. Any loss of proof by the underlying lease session transitions immediately from `held` to `uncertain`, closes readiness and every admission gate, and starts the loss-detection deadline. Only the original underlying session may prove continuous ownership and return `uncertain -> held`; a different or recreated session cannot prove recovery. Confirmed session loss or expiry of the detection deadline without continuous-owner proof transitions to `lost`. `lost` is irreversible in that process: reacquisition and return to serving are forbidden.

The runtime MUST detect confirmed lease loss within `timeouts.extensions.process_lease_loss_detection_seconds`. Lease loss while starting or serving is fatal condition `application_process_lease_lost`; it MUST close readiness, stop admitting HTTP, WebSocket, and job work, and execute EXT-REQ-193 with exit code `70`. Release or process crash MUST release the underlying lease. Browser sessions, incident roles, `deployment_admin`, and profile claim state MUST NOT authorize or replace the lease. A PostgreSQL advisory lease MAY implement this contract but is not normative and MUST preserve the exact session identity and transitions above.

Profiles: base
Verified by: EXT-AC-104, EXT-AC-151

**EXT-REQ-214**
Stage 6 MUST use this closed publication lifecycle:

```text
unpublished -> prepared -> committed -> serving
```

Any failure before `serving` transitions to terminal `failed`.

In `unpublished`, no claim set is visible to dispatch, the runtime MUST NOT dequeue an extension job, no extension workspace is available, and public readiness is false.

In `prepared`, the process MUST construct one immutable `cartulary.extension_publication_plan.v1` containing exactly:

- `schema_id`, exactly `cartulary.extension_publication_plan.v1`;
- `registry_sha256`;
- `resolved_claim_set_sha256`;
- `contribution_registry_sha256`;
- `route_dispatch_plan_sha256`;
- `workspace_registry_sha256`;
- `worker_plan_sha256`;
- `listener_plan_sha256`;
- `client_support_registry_sha256`;
- `implementation_binding_set_sha256`.

The component digests MUST be derived from the closed canonical objects in Table 9-D under `extension_registry_canonical_json_v1`.

**Table 9-D. Publication plan component digests**

| Digest member | Canonical component object |
| --- | --- |
| `contribution_registry_sha256` | `cartulary.extension_contribution_publication_set.v1`, containing exactly `schema_id` and `items[]`; each claimed-profile item contains exactly `profile_id`, `contribution_id`, `kind`, `contribution_sha256`, and `implementation_binding_sha256`. |
| `route_dispatch_plan_sha256` | `cartulary.extension_route_dispatch_plan.v1`, containing exactly `schema_id` and `routes[]`; each recognized route row contains exactly `profile_id`, `route_family`, `dispatch_state`, and `contribution_id`. `dispatch_state` is `claimed` or `inactive`; `contribution_id` is non-null only for `claimed`. |
| `workspace_registry_sha256` | `cartulary.extension_workspace_registry.v1`, containing exactly `schema_id` and `workspaces[]`; each claimed workspace row contains exactly `profile_id`, `workspace_key`, and `contribution_id`. |
| `worker_plan_sha256` | `cartulary.extension_worker_plan.v1`, containing exactly `schema_id` and `workers[]`; each row contains exactly `profile_id` and `worker_kind` for one claimed profile worker. |
| `listener_plan_sha256` | `cartulary.extension_listener_activation_plan.v1`, containing exactly `schema_id`, `http=true`, `websocket=true`, and `job_dequeue=true`. These members identify activation gates, not physical listeners. |
| `implementation_binding_set_sha256` | `cartulary.extension_implementation_binding_set.v1`, containing exactly `schema_id` and `bindings[]`; each claimed-profile row contains exactly `profile_id` and `binding_sha256`. |

`contribution_sha256` MUST equal `extension_contribution_sha256_v1` over the exact canonical contribution declaration in the descriptor. Every component collection MUST reject duplicate identities. `items[]` MUST contain `0..16384` rows, `routes[]` `0..4096`, `workspaces[]` `0..4096`, `workers[]` `0..16384`, and `bindings[]` `0..256`. Contribution items MUST sort by `profile_id`, then `contribution_id`; route rows by `route_family`, then `profile_id`; workspace rows by `profile_id`, then `workspace_key`; worker rows by `profile_id`, then `worker_kind`; binding rows by `profile_id`. All string comparisons use ascending UTF-8 bytes. Every component object and the publication plan MUST contain `1..16777216` canonical bytes. The digest member MUST equal the lowercase SHA-256 digest of the corresponding canonical component bytes. The component objects and publication plan are process-local ephemeral values; they MUST NOT be persisted, packaged, logged, or used as behavior owners.

The contribution set, workspace registry, worker plan, and binding set MUST include only profiles in the resolved claim set. The route plan MUST include every recognized reserved route family so inactive dispatch remains fixed. Each component MUST derive only from the canonical registry, resolved claim set, verified implementation bindings, and packaged client support registry; implementation package discovery, runtime reflection, or unordered registration MUST NOT add an item.

The plan is process-local and ephemeral. Every mandatory component MUST be initialized but quiescent. Listeners MAY be bound but MUST NOT accept public requests; workers MUST NOT dequeue jobs; WebSocket handlers MUST NOT accept subscriptions; workspaces and capabilities MUST remain hidden. Omission behavior: a listener not bound during `prepared` MUST bind and reach activation readiness before it acknowledges activation, and it remains incapable of accepting public work until the admission gate opens.

The `prepared -> committed` transition MUST atomically install the immutable plan as the active process-local extension epoch while the external admission gate remains closed. The `committed -> serving` transition MUST occur only after every mandatory listener and worker acknowledges activation. A listener acknowledges activation only after its event loop is ready to accept immediately when the gate opens while still accepting nothing before that point. A worker acknowledges activation only after its dependencies and claimed-profile state are validated, its run loop is ready, and it is blocked from dequeue by the same gate. A duplicate acknowledgment, acknowledgment for an identity absent from the listener or worker plan, acknowledgment before `committed`, or failure acknowledgment is `extension_publication_failed`. Acknowledgments are process-local synchronization events and MUST NOT be persisted or exposed as public state.

One atomic admission-gate transition is the public publication instant for HTTP acceptance, WebSocket upgrades, job dequeue, readiness success, discovery claim state, and workspace availability.

If any component fails between `committed` and `serving`, the runtime MUST keep the admission gate closed, stop every prepared component, discard the process-local plan, emit no successful extension response, fail with `extension_publication_failed`, and exit with code `2`. The complete publication operation MUST finish within `timeouts.extensions.publication_seconds`.

Profiles: base
Verified by: EXT-AC-105

**EXT-REQ-215**
`extension_deadline_v1` MUST govern every timeout in Table 9-C. Elapsed-time enforcement MUST use an authoritative monotonic clock. Persisted timestamps, audit timestamps, and operator-visible times continue to use the Core canonical UTC wall-clock contract.

For a local timeout `timeout_seconds`, the candidate monotonic deadline MUST be calculated with checked integer arithmetic as `now_monotonic_ns + timeout_seconds * 1000000000`. If the multiplication or addition would exceed `9223372036854775807`, the candidate MUST saturate to `9223372036854775807`; wraparound is forbidden. When a parent deadline exists, the effective deadline is the earlier of the parent and candidate. Equal deadlines select the inherited parent deadline so cancellation ownership and diagnostics remain stable. A deadline is expired exactly when `now_monotonic_ns >= effective_deadline_monotonic_ns`; no positive or scheduler-dependent grace exists at the semantic boundary.

The deadline boundaries in Table 9-E are exhaustive.

**Table 9-E. Extension deadline boundaries**

| Operation | Deadline starts | Deadline ends |
| --- | --- | --- |
| Application-process lease acquisition | Immediately before the first acquisition attempt | Lease acquired or acquisition failed |
| Migration lock acquisition | Immediately before the first lock attempt | Lock acquired or acquisition failed |
| Fresh state initialization | Immediately before the initialization transaction is created | Commit or rollback outcome is known |
| One migration step | Immediately before the step transaction is created | Commit or rollback outcome is known |
| Whole-profile migration | Immediately after the migration lock is acquired | Final profile-state validation completes |
| Preflight or post-migration algorithm | Immediately before invocation | Result passes schema validation |
| Dependency probe | Immediately before invocation | Result passes schema validation |
| Incident Portability participant | Immediately before the claimed-profile participant invocation | Result and referenced payload pass schema, digest, and limit validation |
| Snapshot/Reporting participant | Immediately before the claimed-profile participant invocation | Result and referenced output pass schema, digest, redaction, and limit validation |
| Backup/restore participant | Immediately before the claimed-profile participant invocation | Result passes schema validation and any admitted mutation outcome is known |
| Cross-owner transaction protocol | Immediately before step 3 of EXT-REQ-219 constructs participant inputs | Commit or rollback outcome is known |
| Inactive-job reconciliation | Immediately before the first selected-job read | Reconciliation transaction outcome is known |
| Stage 6 publication | Immediately before component preparation begins | `serving` or `failed` |
| Staged-object startup cleanup | Immediately before the cutoff is captured | No eligible unprocessed row in the cutoff set remains, or cleanup fails |
| Fatal shutdown | On the imported lifecycle's first transition out of `starting` or `running` | Process exit |

On expiry, the runtime MUST issue cancellation, reject and discard any later normal result, roll back an active transaction where one exists, retain required locks until rollback or process termination is proven, and wait no longer than `timeouts.extensions.cancellation_grace_seconds`. A non-cooperative operation remaining active after grace MUST cause process termination; it MUST NOT continue in the background after a timeout result is reported.

Cancellation, timeout, and commit classification MUST use this precedence: a proven successful commit returns committed success or replay regardless of an earlier or simultaneous cancellation signal; a proven absent commit returns cancellation when cancellation was sampled before expiry, otherwise timeout when expiry was sampled first or at the same monotonic instant; an unknown commit or mutation outcome is fatal. An equal cancellation/deadline observation is timeout. `timeouts.extensions.cancellation_grace_seconds=0` means the process-isolation or termination consequence begins immediately after cancellation is issued; it does not move the semantic deadline or permit a late normal result.

Outcome classification MUST use Table 9-F.

**Table 9-F. Deadline outcome classification**

| Condition | Required result |
| --- | --- |
| No durable mutation and rollback proven | Return the operation-specific timeout or startup failure. |
| Durable commit proven successful | Return or replay the committed success. |
| Durable commit proven absent | Return timeout, cancellation, or conflict as the owning operation specifies. |
| Commit or mutation outcome indeterminate | Enter fatal integrity shutdown with `indeterminate_database_commit`. |

No cancellation mechanism is required by name. Any selected cooperative-cancellation, subprocess-isolation, transaction-cancellation, or process-supervision mechanism MUST satisfy every observable result above. For the cross-owner transaction protocol, queueing, participant input construction, prepare, lock acquisition, validation, write, and commit time are included. For other operations, queueing time and lock-wait time are included only where the start boundary table says they are included.

Profiles: base
Verified by: EXT-AC-106, EXT-AC-152

# 10. Claim-resolution algorithm

**EXT-REQ-054**
After acquiring the application-process lease under EXT-REQ-213, startup MUST execute `resolve_extension_claims_v1` exactly once before any HTTP listener, WebSocket listener, background-job runner, extension worker, or extension browser-asset advertisement becomes available.

Profiles: base
Verified by: EXT-AC-013, EXT-AC-020, EXT-AC-085

**EXT-REQ-055**
`resolve_extension_claims_v1` MUST execute the six stages in EXT-REQ-187 in their declared order. A later stage MUST NOT begin when an earlier stage has any finding or terminal failure. The algorithm MUST NOT serve a subset of explicitly requested profiles.

Profiles: base
Verified by: EXT-AC-013, EXT-AC-014, EXT-AC-015, EXT-AC-016, EXT-AC-018, EXT-AC-019, EXT-AC-020, EXT-AC-051, EXT-AC-085

**EXT-REQ-056**
Startup failure precedence is:

1. deployment-configuration envelope and packaged-artifact byte admission;
2. dependency snapshot, owner-input, registry integrity, and collision validation;
3. claim-key and inactive-configuration validation;
4. implementation-binding structural validation;
5. dependency validation and ordering;
6. all-profile side-effect-free preflight;
7. ordered migration;
8. post-migration validation and read-only dependency probes;
9. inactive-profile job reconciliation;
10. publication.

Within stages 1 through 6, the runtime MUST collect every deterministically detectable safe finding in the stage, up to the exact phase limit, and sort findings under §26. If the phase limit is exceeded, the phase result is the single overflow finding defined by EXT-REQ-186. During migration and post-migration validation, the runtime MUST stop at the first failing profile in dependency order and the first failing migration step in state-version order. During inactive-profile reconciliation, complete classification of the selected ordered job set MUST precede any terminal job mutation.

Profiles: base
Verified by: EXT-AC-005, EXT-AC-012, EXT-AC-016, EXT-AC-019, EXT-AC-020, EXT-AC-084, EXT-AC-085

**EXT-REQ-057**
A claim request for a profile whose packaged implementation binding is absent, duplicated, malformed, stale, descriptor-digest mismatched, or contract-major mismatched MUST fail before profile-local migration or contribution publication. The runtime MUST NOT substitute another implementation, compatibility alias, remote service, fallback profile, or older binding.

Profiles: base
Verified by: EXT-AC-010, EXT-AC-018, EXT-AC-080

**EXT-REQ-058**
Profile-local admission MUST be side-effect free until EXT-REQ-187 Stage 3 begins. A profile MUST NOT publish routes, start workers, create public resources, perform third-party egress, expose workspaces, dequeue jobs, or publish capabilities while its admission remains incomplete.

Committed migration steps are the only permitted durable side effect before publication. Their persistence and later resumption MUST follow EXT-REQ-187 and §21.

Profiles: base
Verified by: EXT-AC-013, EXT-AC-020, EXT-AC-042, EXT-AC-067, EXT-AC-085

**EXT-REQ-187**
`resolve_extension_claims_v1` MUST execute these stages.

A profile admission algorithm declared by `admission_validation` MUST be statically packaged and invoked through the shared logical contracts below. Internal language-level function signatures are not prescribed.

`cartulary.extension_admission_validation_context.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_admission_validation_context.v1`;
- `phase`, exactly `preflight`, `post_migration`, or `dependency_probe`;
- `profile_id`;
- `descriptor_sha256`;
- `binding_sha256`;
- `profile_configuration_view`, one complete `cartulary.extension_profile_configuration_view.v1` object containing normalized references but no resolved secret, file, trust-material, or connection value;
- `state_present`, Boolean;
- `state_metadata`, one canonical `cartulary.extension_state_metadata.v1` object or `null`;
- `migration_definition_digests[]`, containing closed `{migration_id, migration_definition_sha256}` objects sorted by `migration_id`;
- `probe_id`, a profile-prefixed probe ID only for `phase='dependency_probe'`, otherwise `null`;
- `timeout_seconds`, the effective Table 9-C timeout for the selected phase.

`migration_definition_digests[]` MUST contain `0..256` items, contain every packaged migration definition for the profile and no other item, and equal `[]` when `state_ownership.kind` is not `extension_versioned`. `state_metadata` MUST be `null` if and only if no state metadata exists at the read boundary. The context MUST contain no raw secret value, access token, incident-authored value, absolute path, database object name, object key, resolved file or trust material, external connection result, or executable value.

`cartulary.extension_admission_validation_result.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_admission_validation_result.v1`;
- `phase`, equal to the input phase;
- `algorithm_id`, equal to the anchor ID of the invoked algorithm locator;
- `findings[]`, containing `0..4096` `cartulary.extension_startup_finding.v1` objects sorted under EXT-REQ-162.

A preflight or post-migration algorithm MUST return exactly one result object before `timeouts.extensions.validation_seconds` expires. A dependency-probe algorithm MUST return exactly one result before `timeouts.extensions.dependency_probe_seconds` expires. A returned non-empty `findings[]` array fails that profile and phase. An invalid result, wrong phase or algorithm ID, thrown or returned implementation error, process panic, or timeout MUST be replaced by exactly one generic `extension_admission_validation_failed` finding for preflight/post-migration or `extension_dependency_probe_failed` finding for a dependency probe. The shared runtime MUST discard any partial profile-produced findings in that condition.

Preflight and post-migration algorithms MUST be deterministic, local, read-only, and free of third-party egress. A dependency probe MUST be read-only, MUST be listed in the descriptor and binding, MUST use only the owner-declared egress destination and secret-resolution interface, and MUST NOT create, update, or delete external state. When a dependency probe requires a raw secret, Core 04 MUST make that value available only through its in-process secret boundary; otherwise no raw secret value is supplied. A raw secret value MUST NOT appear in the context, result, diagnostic, log, telemetry, or retained artifact.

## Stage 1: global structural admission

The runtime MUST:

1. validate the effective Core 04 deployment-configuration container;
2. execute EXT-REQ-180 for every packaged extension contract artifact;
3. validate every registry collision;
4. materialize omitted claim keys as `false`;
5. validate every claim-key value, configuration-contract digest, omission policy, explicit-null rule, and inactive-profile configuration rule;
6. collect the explicit claim-request set;
7. require every requested profile to be claimable and have a non-null contract major;
8. validate every implementation binding structurally;
9. execute dependency validation and produce the canonical dependency order.

Stage 1 MUST perform no profile-state mutation and no third-party egress.

## Stage 2: all-profile side-effect-free preflight

For every requested profile in dependency order, the runtime MUST validate:

- the complete `cartulary.extension_profile_configuration_view.v1`;
- required `secret_ref_v1` references through the Core 04 secret-availability boundary without emitting or transmitting raw secret values;
- required local regular files and trust material only for the claimed profile;
- authoritative state-presence and state-metadata consistency;
- stored state-version compatibility;
- the complete consecutive migration path and every immutable migration-definition digest;
- required contribution, capability, worker, job, and admission-validation declarations;
- the owner preflight algorithm exactly once when `admission_validation.preflight_algorithm_ref` is non-null.

A null preflight algorithm reference means no additional profile-owned preflight algorithm. The shared structural checks in this stage remain mandatory. The preflight algorithm MUST use the context and result contracts in EXT-REQ-187 and MUST NOT acquire a public resource identity, create a job, modify authoritative or derived state, register a route, start a worker, publish a contribution, or perform third-party egress.

A migration MUST NOT begin unless every requested profile passes Stage 2.

## Stage 3: ordered migration

For each requested profile in dependency order, the runtime MUST:

1. acquire the profile migration lock under §21;
2. re-read authoritative state presence, state metadata, migration ledger, and packaged migration definitions after acquiring the lock;
3. initialize state metadata or apply required consecutive forward migrations under §21;
4. validate each committed step and the final local state;
5. release the profile migration lock only after the profile's final local state validation completes.

Each migration step commits independently and atomically. A committed step remains committed if a later step or later profile fails. Cross-profile rollback is forbidden. No route, workspace, capability, profile worker, or listener is published after such a failure. The next startup MUST resume from the committed state metadata and ledger.

## Stage 4: post-migration validation

For every requested profile in dependency order, the runtime MUST:

1. invoke the post-migration algorithm exactly once when `admission_validation.post_migration_algorithm_ref` is non-null;
2. treat a null post-migration algorithm reference as no additional profile-owned post-migration validation;
3. invoke every declared dependency probe exactly once in ascending UTF-8 order of `probe_id`;
4. use the context, result, timeout, finding, egress, and no-external-mutation contracts in EXT-REQ-187.

A failed validation or probe blocks readiness but does not reverse a committed migration. The runtime MUST stop Stage 4 at the first failing profile and first failing algorithm or probe in the declared order.

## Stage 5: inactive-profile reconciliation

For every profile in state `unclaimed` or `recognized_unclaimable`, ordered by ascending UTF-8 bytes of `profile_id`, the runtime MUST execute `reconcile_inactive_extension_jobs_v1` under §22. The runtime MUST perform no profile-local semantic migration, start no profile worker, and make no third-party request for an inactive profile.

## Stage 6: publication

The runtime MUST NOT execute the `unpublished -> prepared -> committed -> serving` lifecycle in EXT-REQ-214 until Stages 1 through 5 all succeed. After they succeed, the runtime MUST execute that exact lifecycle. The resolved claim set, contribution registry, route dispatch plan, workspace registry, worker plan, listener plan, client support registry, and implementation-binding set MUST be bound into one immutable publication plan. Discovery, dispatch, workbook startup, telemetry, job admission, and conformance MUST consume that same `serving` epoch. A publication failure MUST expose no partial epoch, accept no public work, and terminate startup with exit code `2`.

Profiles: base
Verified by: EXT-AC-085

# 11. Required runtime dependencies

**EXT-REQ-059**
The current dependency model supports only required profile dependencies. Optional dependencies, capability-gated dependencies, semantic-version ranges, alternative providers, and automatic dependency claims are unavailable in contract major `1`.

Profiles: base
Verified by: EXT-AC-016, EXT-AC-017, EXT-AC-018, EXT-AC-019

**EXT-REQ-060**
A profile dependency is satisfied only when:

- the dependency profile is recognized;
- the dependency is explicitly requested as claimed;
- the dependency is claimable;
- the dependency's current contract major equals `required_contract_major`;
- the dependency has a valid implementation binding;
- the dependency completes claim admission before the dependent profile.

Profiles: base
Verified by: EXT-AC-016, EXT-AC-018, EXT-AC-020

**EXT-REQ-061**
The implementation MUST NOT auto-claim, auto-enable, auto-install, or silently substitute a required dependency. A dependent profile whose dependency claim key is false or omitted MUST fail with `extension_dependency_not_claimed`.

Profiles: base
Verified by: EXT-AC-016, EXT-AC-017

**EXT-REQ-062**
`validate_extension_dependencies_v1` MUST execute before topological ordering:

1. validate every normalized dependency object;
2. emit one self-dependency finding per self-edge;
3. emit one duplicate-dependency finding per duplicated `(dependent_profile_id, dependency_profile_id)` pair;
4. emit one `extension_dependency_not_claimed` finding for every dependency edge whose target is absent from the explicit request set;
5. validate recognition, claimability, current contract major, required contract major, and implementation binding for every target;
6. construct a graph containing only explicitly requested vertices and only dependency edges whose target is explicitly requested;
7. compute the graph's maximal strongly connected components;
8. emit exactly one cycle finding for each component containing two or more profiles;
9. produce no topological order when any dependency finding exists.

A missing dependency MUST be classified before graph construction and MUST NOT be added as an implicit vertex, misclassified as a cycle, or auto-claimed.

Each cycle finding MUST contain every profile ID in the component and every dependency edge whose two endpoints are in the component. Profile IDs MUST be unique and sorted by ascending UTF-8 bytes. Internal edges MUST be unique and sorted by `from_profile_id`, then `to_profile_id`, ascending UTF-8 bytes. The implementation MUST NOT emit an arbitrary representative cycle or every possible cycle path.

For a valid graph, `resolve_extension_dependency_order_v1` MUST use this exact ordering algorithm:

1. For each vertex, create `unresolved_dependencies` containing the profile IDs named by its outgoing dependency edges.
2. Create `eligible` from every un-emitted vertex whose `unresolved_dependencies` is empty.
3. If `eligible` is non-empty, select its smallest `profile_id` by ascending UTF-8 bytes, append it to the output, mark it emitted, and remove its `profile_id` from every remaining vertex's `unresolved_dependencies`; then repeat step 2.
4. If `eligible` is empty and un-emitted vertices remain, fail because dependency validation did not remove a cycle.
5. If no un-emitted vertices remain, return the output.

Profiles: base
Verified by: EXT-AC-019, EXT-AC-083

**EXT-REQ-063**
The current runtime dependency registry MUST contain exactly one edge:

`network_flow_activity -> import@1`

No other current profile has a required runtime profile dependency.

Profiles: base, network_flow_activity
Verified by: EXT-AC-016, EXT-AC-017, EXT-AC-072

**EXT-REQ-064**
Build-time or conformance dependencies on Graph Projection, Testing Harness, Reporting, or other adopted subsystem NLSpecs are not runtime profile dependencies unless Core 00 explicitly adds them to the runtime dependency registry.

Profiles: base
Verified by: EXT-AC-016, EXT-AC-073

**EXT-REQ-185**
Validation finding multiplicity and grouping MUST follow Table 11-A. A validator MUST NOT substitute a summary, first-only result, every-pair expansion, or every-cycle expansion for the declared representation.

Profiles: base
Verified by: EXT-AC-083, EXT-AC-084

**Table 11-A. Canonical dependency and collision finding enumeration**

| Condition | Required finding representation |
| --- | --- |
| Missing dependency | One finding per normalized dependency edge. |
| Self-dependency | One finding per self-edge. |
| Duplicate dependency | One finding per duplicated dependent/dependency identity, containing the duplicate count. |
| Directed cycle | One finding per maximal strongly connected component containing at least two profiles. |
| Duplicate scalar identity | One grouped finding per duplicated identity, containing every conflicting profile ID. |
| Route-family overlap | One finding per unordered conflicting route-family pair. |
| Missing required owner fact | One finding per missing `(profile_id, fact_kind)` identity. |
| Unrecognized owner fact | One finding per unrecognized fact identity. |

Every grouped `conflicting_profile_ids[]` array MUST be unique and ordered by ascending UTF-8 bytes. Every unordered pair MUST be emitted with the lexically smaller canonical token or profile tuple first. Findings MUST then sort under EXT-REQ-162.

A missing required owner fact MUST use path `$.owner_fragments`, reason code `extension_registry_invalid`, and details containing `artifact_kind='owner_input_registry'`, `safe_ref` equal to that profile's `recognized_profile.owner_contract_ref`, `expected='fact_kind=<fact_kind>;required_count=1'`, and `actual='fact_kind=<fact_kind>;actual_count=0'`. An unrecognized owner fact MUST use the exact normalized path `$.owner_fragments[<fragment_index>].facts[<fact_index>].profile_id`, reason code `extension_registry_invalid`, and details containing `artifact_kind='owner_fragment'`, `safe_ref` equal to that fact's `owner_contract_ref`, `expected='recognized_profile_id'`, and `actual` equal to the exact unrecognized `profile_id`. `<fragment_index>` and `<fact_index>` are the zero-based positions after the canonical sorting in EXT-REQ-177. These two invalid-input conditions MUST NOT use `collision_class`, `conflicting_profile_ids[]`, `conflicting_tokens[]`, or `duplicate_count`.

# 12. Implementation bindings

**EXT-REQ-065**
An implementation binding MUST be registered at build time inside the application deployable and MUST associate exactly one `profile_id` with exactly one canonical descriptor digest. The binding MUST NOT be discovered from a runtime directory, database row, uploaded file, environment variable containing code, remote URL, object-store object, reference pack, or browser input.

Profiles: base
Verified by: EXT-AC-010, EXT-AC-070, EXT-AC-080

**EXT-REQ-066**
A binding MUST provide the packaged behavior required by its declared contributions, capabilities, admission algorithms, dependency probes, state ownership, migrations, worker kinds, and job kinds. The implementation mechanism and internal function signatures are not normative. Their observable effects MUST satisfy this NLSpec and the named profile owner contract.

Profiles: base
Verified by: EXT-AC-010, EXT-AC-013, EXT-AC-029, EXT-AC-051, EXT-AC-080

**EXT-REQ-067**
The runtime MUST reject duplicate bindings for one `profile_id` and any binding for an unrecognized profile. A build MAY omit a binding for a recognized unclaimable profile or for a claimable profile the build cannot claim. Omission behavior: a true claim request for a profile without one valid binding MUST fail with `extension_implementation_unavailable`.

Profiles: base
Verified by: EXT-AC-005, EXT-AC-010, EXT-AC-014

**EXT-REQ-181**
`cartulary.extension_implementation_binding.v1` MUST be a closed canonical object containing exactly the members in Table 12-A.

Profiles: base
Verified by: EXT-AC-080

**Table 12-A. `cartulary.extension_implementation_binding.v1`**

| Member | Required rule |
| --- | --- |
| `schema_id` | Exactly `cartulary.extension_implementation_binding.v1`. |
| `profile_id` | Exact recognized profile ID. |
| `contract_major` | Exact canonical descriptor contract major; non-null. |
| `descriptor_sha256` | Exact `extension_descriptor_sha256_v1` digest. |
| `implemented_contribution_ids[]` | Unique declared executable contribution IDs. |
| `supported_capability_ids[]` | Unique capability IDs actually supported by the packaged implementation. |
| `state_ownership_kind` | Exactly `none`, `core_managed`, or `extension_versioned`. |
| `preflight_algorithm_id` | Algorithm anchor ID from the descriptor admission-validation object, or `null`. |
| `post_migration_algorithm_id` | Algorithm anchor ID from the descriptor admission-validation object, or `null`. |
| `initialization_definition_sha256` | EXT-REQ-234 definition digest for `extension_versioned`, otherwise `null`. |
| `initialization_algorithm_id` | Algorithm variant anchor ID, or `null` for `empty` and non-versioned profiles. |
| `final_state_validation_algorithm_id` | Final state-validator anchor ID for `extension_versioned`, otherwise `null`. |
| `dependency_probe_ids[]` | Unique probe IDs implemented by the packaged binding. |
| `migration_definitions[]` | Closed migration binding objects. |
| `physical_state_binding_sha256` | EXT-REQ-223 binding digest for `extension_versioned`, otherwise `null`. |
| `backup_codec_bindings[]` | Closed binding-to-codec identity and digest rows. |
| `rebuild_algorithm_ids[]` | Exact derived-state rebuild algorithm set. |
| `transaction_participant_limits` | Closed EXT-REQ-219 limit object. |
| `supporting_schema_ids[]` | Exact packaged shared/profile supporting schema set required by this binding. |
| `worker_kinds[]` | Unique implemented worker kinds. |
| `job_kind_contracts[]` | Closed implemented job-kind contract bindings. |
| `participant_contracts[]` | Closed implemented participant-contract bindings. |

A `migration_definitions[]` item MUST contain exactly:

- `migration_id`;
- `from_state_version`;
- `to_state_version`;
- `migration_definition_sha256`;
- `apply_algorithm_id`;
- `validation_algorithm_id`.

The versions, digest, and algorithm IDs MUST exactly match one normalized owner `migration_definition` fact. The array MUST contain `0..256` items and sort by `from_state_version`, then `to_state_version`, then `migration_id`.

A `job_kind_contracts[]` item MUST contain exactly `job_kind` and `job_kind_contract_sha256`. The digest MUST equal the canonical digest of the normalized owner `cartulary.extension_job_kind_contract.v1`. The array MUST contain `0..64` items and sort by ascending UTF-8 bytes of `job_kind`.

A `participant_contracts[]` item MUST contain exactly `participant_id`, `participant_contract_sha256`, and `algorithm_ids[]`. The digest MUST equal the canonical digest of the participant contract selected by the contribution. `algorithm_ids[]` MUST equal every packaged algorithm ID required by that contract, contain `1..16` unique values, and sort by ascending UTF-8 bytes. The array MUST contain `0..64` items, reject duplicate participant IDs, and sort by ascending UTF-8 bytes of `participant_id`.

A `backup_codec_bindings[]` item MUST contain exactly `binding_id`, `backup_codec_id`, and `backup_codec_sha256`. It MUST equal one physical-state binding row and one packaged EXT-REQ-235 codec. The array MUST contain `0..4096` rows, reject duplicate binding and codec identities, and sort by `binding_id`.

`transaction_participant_limits` MUST contain exactly `participant_input_bytes=16777216`, `aggregate_input_bytes=67108864`, `serialization_keys_per_participant=1024`, `aggregate_serialization_keys=4096`, `result_bytes=1048576`, and `validation_findings=256`. These values are parity declarations, not configurable limits. `supporting_schema_ids[]` MUST contain `0..4096` unique current schema IDs sorted by UTF-8 bytes. `rebuild_algorithm_ids[]` MUST contain `0..4096` unique current algorithm IDs sorted by UTF-8 bytes.

All other binding arrays, including `dependency_probe_ids[]`, MUST sort by ascending UTF-8 bytes of their exact token. A null algorithm ID is valid only when the matching descriptor algorithm reference is null. A canonical binding MUST contain `1..1048576` bytes including its final LF.

**EXT-REQ-182**
Implementation-binding admission MUST enforce all rows in Table 12-B. A mismatch MUST fail with `extension_implementation_unavailable` before migration or contribution publication, except a migration-definition digest mismatch with a committed ledger entry uses `extension_migration_definition_changed`.

Profiles: base
Verified by: EXT-AC-080

**Table 12-B. Binding-to-descriptor parity**

| Binding surface | Required parity |
| --- | --- |
| `profile_id` | Exact equality with the descriptor. |
| `contract_major` | Exact equality with the non-null descriptor major. |
| `descriptor_sha256` | Exact equality with the packaged canonical descriptor digest. |
| `state_ownership_kind` | Exact equality with `descriptor.state_ownership.kind`. |
| `preflight_algorithm_id` | Equal to the preflight algorithm locator anchor ID, or `null` exactly when the descriptor reference is `null`. |
| `post_migration_algorithm_id` | Equal to the post-migration algorithm locator anchor ID, or `null` exactly when the descriptor reference is `null`. |
| `initialization_definition_sha256` | Equal to `descriptor.state_ownership.initialization_definition_sha256` for `extension_versioned`; `null` otherwise. |
| `initialization_algorithm_id` | Equal to the EXT-REQ-234 algorithm ID for `kind='algorithm'`; `null` for `kind='empty'` and non-versioned profiles. |
| `final_state_validation_algorithm_id` | Equal to `descriptor.state_ownership.final_state_validation_algorithm_id` for `extension_versioned`; `null` otherwise. |
| `dependency_probe_ids[]` | Exact equality with descriptor `admission_validation.dependency_probes[].probe_id`. |
| `implemented_contribution_ids[]` | Exact equality with descriptor contribution IDs whose Table 16-A `binding_requirement` is `handler`, `participant`, or `worker`. |
| `supported_capability_ids[]` | Subset of descriptor `capability_ids[]`; every emitted discovery capability must also be present here and pass capability-specific admission. |
| `migration_definitions[]` | Exact equality with normalized owner migration definitions for `extension_versioned`; `[]` for the other state kinds. |
| `physical_state_binding_sha256` | Exact digest of the required EXT-REQ-223 object for `extension_versioned`; `null` otherwise. |
| `backup_codec_bindings[]` | Exact binding ID, codec ID, and codec digest set from the physical-state binding; `[]` when no binding exists. |
| `rebuild_algorithm_ids[]` | Exact non-null rebuild algorithm set from derived binding rows. |
| `transaction_participant_limits` | Exact equality with the closed EXT-REQ-219 limits. |
| `supporting_schema_ids[]` | Exact current runtime schema set required by the descriptor, initialization, migration, state, transaction, backup, job, and participant contracts supplied by the binding; Harness-only schemas remain integrity-accounted and are not binding dependencies. |
| `worker_kinds[]` | Exact equality with normalized owner `worker_kind` facts. |
| `job_kind_contracts[]` | Exact equality with normalized owner job-kind identities and canonical contract digests. |
| `participant_contracts[]` | Exact equality with every contribution that requires a participant, its canonical contract digest, and every algorithm ID declared by that contract. |

An extra contribution, capability, admission algorithm, dependency probe, initialization, migration, physical binding, codec, rebuild algorithm, schema, worker, job, participant contract, or participant algorithm declaration is invalid. A missing required declaration in any of those families is invalid. Capability omission is valid because capabilities are additive; an omitted capability MUST NOT be advertised or invoked.

# 13. Registry collision detection

**EXT-REQ-068**
`validate_extension_registry_collisions_v1` MUST reject every collision class in Table 13-A before claim-state resolution.

Profiles: base
Verified by: EXT-AC-021, EXT-AC-022, EXT-AC-023, EXT-AC-083

**Table 13-A. Closed collision registry**

| `collision_class` | Collision key or predicate | Required `conflicting_tokens[]` derivation | Required reason family |
| --- | --- | --- | --- |
| `profile_id` | Duplicate `profile_id` | One exact duplicated profile ID | `extension_registry_conflict` |
| `owner_fragment_id` | Duplicate `owner_fragment_id` | One exact duplicated owner-fragment ID | `extension_registry_conflict` |
| `claim_key` | Duplicate `claim_config_key` | One exact duplicated claim key | `extension_registry_conflict` |
| `configuration_namespace` | Duplicate exact `profile_id` namespace | One exact duplicated namespace | `extension_registry_conflict` |
| `capability_id` | Duplicate `capability_id` | One exact duplicated capability ID | `extension_registry_conflict` |
| `contribution_id` | Duplicate `contribution_id` | One exact duplicated contribution ID | `extension_registry_conflict` |
| `workspace_identity` | Duplicate tuple `(profile_id, workspace_key)` | One `<profile_id>/<workspace_key>` token | `extension_registry_conflict` |
| `extension_resource_kind` | Duplicate exact resource-kind token | One exact duplicated resource-kind token | `extension_registry_conflict` |
| `import_target` | Duplicate exact `target_kind` | One exact duplicated target kind | `extension_registry_conflict` |
| `job_resource_ref_kind` | Duplicate exact `resource_ref_kind` | One exact duplicated resource-reference kind | `extension_registry_conflict` |
| `public_schema_id` | Duplicate exact schema ID | One exact duplicated schema ID | `extension_registry_conflict` |
| `transaction_participant` | Duplicate exact `participant_id` | One exact duplicated participant ID | `extension_registry_conflict` |
| `deployment_admin_panel` | Duplicate exact `panel_key` | One exact duplicated panel key | `extension_registry_conflict` |
| `authentication_entry` | Duplicate exact `entry_key` | One exact duplicated entry key | `extension_registry_conflict` |
| `conformance_manifest_id` | Duplicate exact manifest ID | One exact duplicated manifest ID | `extension_registry_conflict` |
| `migration_lineage_id` | Duplicate lineage across different profiles or incompatible declarations | One exact duplicated lineage ID | `extension_registry_conflict` |
| `migration_id` | Duplicate exact migration ID | One exact duplicated migration ID | `extension_registry_conflict` |
| `migration_edge` | Duplicate `(profile_id, from_state_version, to_state_version)` | One `<profile_id>:<from_state_version>-><to_state_version>` token | `extension_registry_conflict` |
| `worker_kind` | Duplicate exact worker kind | One exact duplicated worker kind | `extension_registry_conflict` |
| `job_kind` | Duplicate exact job kind | One exact duplicated job kind | `extension_registry_conflict` |
| `implementation_binding_duplicate` | More than one binding declares the same recognized `profile_id` | One exact duplicated binding profile ID | `extension_registry_conflict` |
| `owner_fact_identity` | Duplicate `cartulary.extension_owner_fact_identity.v1` bytes not assigned a narrower collision class in this table | One token equal to `<fact_kind>:sha256:<identity_sha256>`, where `identity_sha256` is derived exactly by EXT-REQ-205 | `extension_registry_conflict` |
| `route_family_overlap` | Equal or overlapping extension route-family templates | The two exact route roots, sorted by UTF-8 bytes | `extension_registry_conflict` |
| `base_route_capture` | Extension route overlaps a Base route-family root | Extension root and Base root, sorted by UTF-8 bytes | `extension_registry_conflict` |
| `dependency_self` | Self-dependency | One `<profile_id>-><profile_id>` token | `extension_registry_conflict` |
| `dependency_duplicate` | Duplicate dependency edge | One `<dependent_profile_id>-><dependency_profile_id>` token | `extension_registry_conflict` |
| `dependency_cycle` | Maximal strongly connected component containing at least two profiles | Every internal `<from_profile_id>-><to_profile_id>` edge token, sorted | `extension_registry_conflict` |

When one duplicate declaration satisfies both `owner_fact_identity` and a narrower Table 13-A class, the validator MUST emit only the narrower class. A binding for an unrecognized profile is not a collision; it MUST fail implementation-binding validation under EXT-REQ-067 with `extension_implementation_unavailable`.

**EXT-REQ-069**
Core 01 MUST adopt `route_family_template_v1` with these lexical rules:

- the string begins with `/api/v1/`;
- the string contains `1..512` ASCII bytes and no more than 32 segments;
- the string contains no query, fragment, percent escape, backslash, NUL, empty segment, `.` segment, `..` segment, or trailing slash;
- a literal segment contains one or more ASCII characters from `[A-Za-z0-9._~-]`;
- a parameter segment is exactly `{` + a parameter name + `}`;
- a parameter name matches `[a-z][a-z0-9_]{0,63}`;
- parameter names are identity-neutral for overlap comparison;
- the root identifies itself and every descendant route beneath it.

Profiles: base
Verified by: EXT-AC-022, EXT-AC-023, EXT-AC-082

**EXT-REQ-070**
Core 01 MUST adopt `route_family_overlap_v1` with this algorithm:

1. parse both roots under `route_family_template_v1`;
2. compare segments from the beginning through the shorter segment count;
3. two compared segments are compatible when they are equal literals or either segment is a parameter;
4. if any compared segments are incompatible, the roots do not overlap;
5. if all compared segments are compatible, the roots overlap because an equal-length path or a descendant of the shorter root can match both.

Two roots that overlap under this algorithm MUST NOT coexist in the combined Base and extension route-family registry.

Profiles: base
Verified by: EXT-AC-022, EXT-AC-023

**EXT-REQ-210**
Core 01 MUST produce one canonical `cartulary.base_route_reservation_registry.v1` object containing exactly:

- `schema_id`, exactly `cartulary.base_route_reservation_registry.v1`;
- `reservations[]`.

Each `reservations[]` row MUST contain exactly:

- `reservation_id`;
- `path_template`;
- `match_scope`;
- `owner_contract_ref`.

`reservation_id` MUST satisfy the public schema-ID scalar contract. `path_template` MUST satisfy `route_family_template_v1`. `match_scope` MUST equal `exact` or `descendants`. `owner_contract_ref` MUST resolve to the exact Core 01 requirement that owns the reservation. Rows MUST reject duplicate reservation IDs and duplicate `(path_template, match_scope)` tuples and sort by `path_template`, then `match_scope`, then `reservation_id`, all by ascending UTF-8 bytes.

The registry reserves path namespaces independently of HTTP method. An `exact` row reserves the named path for every method. A `descendants` row reserves the named path and every descendant path. For overlap comparison, a literal segment is compatible only with the same literal; a parameter segment is compatible with every legal literal or parameter segment in contract major `1`. The closed overlap matrix is Table 13-B.

**Table 13-B. Base route-reservation overlap matrix**

| Left scope | Right scope | Collision condition |
| --- | --- | --- |
| `exact` | `exact` | Same segment count and every segment compatible. |
| `exact` | `descendants` | The exact path has at least the root segment count and its prefix is compatible with the descendant root. |
| `descendants` | `exact` | Symmetric to the prior row. |
| `descendants` | `descendants` | Either root is a compatible prefix of the other. |

Every extension `route_family` is treated as `descendants`. `validate_extension_registry_collisions_v1` MUST consume only the digest-bound Base registry imported by EXT-REQ-174 when evaluating `base_route_capture`. Parsing Core prose, reflecting a running router, enumerating handlers, or scanning OpenAPI output MUST NOT create the normative reservation set. Runtime route enumeration MAY provide parity evidence only. Omission behavior: when enumeration is not performed, route-collision admission remains unchanged and uses only the canonical Base registry.

The Base registry MUST cover every current Base public path namespace exactly once and MUST contain no extension-owned namespace. A broad descendant reservation MUST NOT capture an intentionally extensible incident root. Its canonical digest MUST appear in the dependency snapshot, registry-integrity object, and drift accounting.

Profiles: base
Verified by: EXT-AC-107

**EXT-REQ-071**
Collision detection MUST compare normalized owner declarations, not visible labels, implementation package names, database tables, generated type names, or array positions. A collision MUST fail before any requested profile performs profile-local preflight or migration.

Profiles: base
Verified by: EXT-AC-021, EXT-AC-022

**EXT-REQ-072**
Registry collision output MUST use `cartulary.extension_startup_finding.v1` and the grouping rules in EXT-REQ-185. After grouping, findings MUST sort by:

1. `path` ascending UTF-8 bytes;
2. `reason_code` ascending UTF-8 bytes;
3. canonical `details` bytes under `extension_registry_canonical_json_v1`.

A route-overlap pair MUST be emitted only once with the lexically smaller route-family string first. A duplicate scalar identity MUST produce one grouped finding containing every conflicting profile ID. A collision output MUST NOT depend on owner-fragment order, hash-map order, search ranking, or the first conflict encountered.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-021, EXT-AC-083, EXT-AC-084

# 14. Contract, document, capability, and state compatibility

**EXT-REQ-073**
The subsystem MUST keep the version families in Table 14-A distinct. An implementation MUST NOT infer one version from another. `document_version` MUST use exactly three dot-separated base-10 integers `MAJOR.MINOR.PATCH`, each in `0..2147483647`, with no leading zero except `0`, no `v` prefix, no prerelease suffix, and no build metadata.

Profiles: base
Verified by: EXT-AC-015, EXT-AC-018, EXT-AC-028

**Table 14-A. Version families**

| Version | Purpose | Public discovery | Primary owner |
| --- | --- | ---: | --- |
| `contract_major` | Public caller compatibility and dependency compatibility | Yes | Core 00 plus named profile owner |
| `document_version` | Repository revision and adoption traceability | No | Named profile owner |
| `state_schema_version` | Interpretation of extension-owned durable state | No | Named profile owner plus this lifecycle contract |
| Schema or algorithm version suffix | Canonical member, byte, identity, or algorithm interpretation | Only when the public owner exposes the identifier | Owning schema or algorithm document |

**EXT-REQ-074**
A profile owner MUST classify every behavior-affecting change according to Table 14-B. A change not represented by one row is a specification defect and MUST NOT be released until this table or a narrower owner table classifies it.

Profiles: base
Verified by: EXT-AC-015, EXT-AC-018, EXT-AC-028, EXT-AC-050, EXT-AC-093

**Table 14-B. Required version action**

| Change class | Required action |
| --- | --- |
| Editorial correction with no observable effect | Increment document patch only. |
| Added test, fixture, or conformance evidence with no behavior change | Increment document patch only. |
| Implementation fix restoring already specified behavior | Increment document patch; no contract-major change. |
| New additive advertised capability | Increment document minor; retain contract major; advertise only when available. |
| New additive workspace or resource kind that compatible clients safely ignore and discovery explicitly advertises | Increment document minor; retain contract major. |
| New route beneath an already reserved route family with no changed existing route semantics | Increment document minor; retain contract major. |
| New optional response member accepted by the named tolerant consumer decoder | Increment document minor; retain contract major. |
| New optional request member | Retain major only when capability-advertised and clients send it only while advertised; otherwise increment contract major. |
| New error code for a new condition | Retain major only when an unknown-error fallback is normative and HTTP status, retryability, and existing condition mapping do not change. |
| Changed error code, HTTP status, retryability, precedence, or required details for an existing condition | Increment contract major. |
| New unreserved public route family | Increment contract major. |
| Removed, renamed, or semantically changed route, required field, workspace, resource kind, contribution, capability definition, public schema, or required job result | Increment contract major. |
| Changed default, limit, ordering, normalization, identity, digest, lifecycle, authorization, disclosure, audit behavior, persistence interpretation, or recovery behavior visible to callers | Increment contract major and every affected schema or algorithm ID. |
| Changed extension-owned persisted interpretation without a public behavior change | Increment state schema version; define a consecutive forward migration. |
| Changed persisted interpretation visible to callers | Increment state schema version and contract major. |
| Changed canonicalization or digest input | Introduce a new algorithm ID and every dependent digest or schema ID; do not reinterpret old bytes. |
| Changed published migration definition | Introduce a new migration ID and, when required, a later state version; reusing the old migration ID is forbidden. |
| Removed support for an advertised capability in a later deployment release | Increment contract major unless the capability owner explicitly defined revocability and the current request consistency rule. |

**EXT-REQ-075**
Clients and dependent profiles MUST determine compatibility from `contract_major` and advertised capabilities. They MUST NOT infer compatibility from `/api/v1`, route presence, document filename, document version, package version, implementation version, registry digest, or state schema version.

Profiles: base
Verified by: EXT-AC-015, EXT-AC-018, EXT-AC-026, EXT-AC-093

**EXT-REQ-076**
A runtime dependency requires exact contract-major equality. A range, minimum, maximum, wildcard, document-version comparison, or capability-based substitution is invalid in contract major `1`.

Profiles: base
Verified by: EXT-AC-018, EXT-AC-019

**EXT-REQ-077**
Capabilities are disabled in contract major `1`. A capability owner fact, descriptor member, implementation binding, client-support row, discovery item, resolved claim value, or other generated or runtime capability array MUST be present where its schema requires the member and MUST equal `[]`. Any nonempty capability fact or array MUST fail generation or startup as applicable. Any public, browser, operator, or internal attempt to activate a capability MUST fail without side effects using exact reason code `extension_capability_not_supported`; it MUST NOT be dispatched, ignored as successful, or translated into profile behavior. A later capability model requires an adopted owner contract defining execution, authorization, lifecycle, compatibility, and conformance before any nonempty value is admitted.

Profiles: base
Verified by: EXT-AC-017, EXT-AC-029, EXT-AC-080, EXT-AC-157

**EXT-REQ-078**
Removing support for an advertised capability from a running deployment requires restart and a new discovery result. A request admitted while the capability was advertised MUST use the owning route's ordinary consistency and job semantics; the subsystem MUST NOT silently change the claim set or reinterpret the request as another capability.

Profiles: base
Verified by: EXT-AC-054

**EXT-REQ-079**
A stored state version greater than the packaged `current_state_version`, lower than `minimum_migratable_state_version`, separated by a missing consecutive migration, or associated with another `migration_lineage_id` MUST be incompatible and MUST block the profile claim before profile mutation.

Profiles: base
Verified by: EXT-AC-028, EXT-AC-044, EXT-AC-045, EXT-AC-046, EXT-AC-086

**EXT-REQ-196**
Every profile owner, Core companion, generated schema, implementation binding, client compatibility registry, and conformance manifest MUST apply the same Table 14-B classification. A conformance check MUST fail when two artifacts classify the same change differently or when a behavior-affecting change lacks the required document, contract, state, schema, or algorithm version action.

Profiles: base
Verified by: EXT-AC-093

# 15. Public extension discovery integration

**EXT-REQ-080**
Core 01 MUST expose one generic producer item schema named `cartulary.extension_discovery_item.v1` containing exactly the members in Table 15-A. Profile-specific discovery item shapes are invalid.

Profiles: base
Verified by: EXT-AC-024, EXT-AC-025, EXT-AC-026, EXT-AC-092

**Table 15-A. `cartulary.extension_discovery_item.v1` strict producer schema**

| Member | Type | Required behavior |
| --- | --- | --- |
| `profile_id` | String | Exact recognized identity. |
| `claimable` | Boolean | Exact Core 00 value. |
| `claimed` | Boolean | `true` only when present in the resolved claim set. |
| `contract_major` | Integer or `null` | Exact Core 00 value; `null` for a recognized profile without a current claimable contract. |
| `route_families[]` | Array of strings | All reserved roots, including while inactive. |
| `workspace_keys[]` | Array of strings | All declared workspace identities, including while inactive. |
| `capabilities[]` | Array of strings | Currently admitted capabilities; empty while inactive. |

**EXT-REQ-081**
`GET /api/v1/extensions` MUST return every recognized profile ordered by ascending UTF-8 bytes of `profile_id`. The item count MUST be `0..256`. Every nested set-like array MUST be unique and ordered by ascending UTF-8 bytes. Pagination remains unsupported under the Core 01 singleton-discovery contract. The route requires a current authenticated session, requires no incident membership, and requires no `deployment_admin` capability. Exceeding the item bound MUST fail; truncation is forbidden.

Profiles: base
Verified by: EXT-AC-024, EXT-AC-025, EXT-AC-082

**EXT-REQ-082**
Reserved route families and workspace keys MUST remain present while a profile is inactive. Their presence reserves stable identity and dispatch only; it MUST NOT expose owner resources, authorize the caller, start workers, or render an extension workspace.

Profiles: base
Verified by: EXT-AC-016, EXT-AC-027, EXT-AC-028, EXT-AC-034

**EXT-REQ-083**
The public discovery item MUST NOT contain `document_version`, singular `route_root`, implementation package version, build revision, stored state version, registry digest, secret-reference names, provider metadata, live resource counts, incident data, table names, or configuration values.

Profiles: base
Verified by: EXT-AC-026, EXT-AC-030, EXT-AC-069, EXT-AC-092

**EXT-REQ-084**
`claimed=true` means only that startup admission succeeded for the profile. It MUST NOT imply that any resource exists, that any capability is available unless advertised, or that the current caller is authorized for every route or incident workspace.

Profiles: base
Verified by: EXT-AC-024, EXT-AC-032

**EXT-REQ-085**
Compatible clients MUST process discovery only through `decode_extension_discovery_item_v1` from EXT-REQ-194. They MUST NOT treat an unknown profile, member, capability, or workspace as a Base workbook surface, Core record, route authorization grant, or executable instruction.

Profiles: base
Verified by: EXT-AC-023, EXT-AC-039, EXT-AC-091

**EXT-REQ-086**
For the coordinated adoption baseline, the `network_flow_activity` discovery item MUST report `claimable=true`, `contract_major=2`, the reserved Network Flow route family, `workspace_keys=["network_analysis"]`, and `capabilities=[]` whether claimed or unclaimed; only `claimed` changes with the resolved claim set.

Profiles: base, network_flow_activity
Verified by: EXT-AC-003, EXT-AC-016, EXT-AC-027, EXT-AC-028, EXT-AC-029

**EXT-REQ-194**
The current server producer and compatible client decoder MUST be intentionally asymmetric:

- a current v1 producer MUST emit exactly the seven Table 15-A members and no unknown member;
- `decode_extension_discovery_item_v1` MUST accept unknown additive members while enforcing every known v1 member.

`decode_extension_discovery_item_v1` MUST:

1. reject invalid UTF-8, invalid JSON, duplicate members, or a non-object item;
2. require every known v1 member;
3. validate the exact type, nullability, grammar, bound, uniqueness, and ordering of every known member;
4. ignore unknown additional members after the enclosing Core response byte limit is enforced;
5. never treat an unknown member as a route, component, callback, schema, authorization fact, or executable instruction;
6. pass known values to the unknown-value matrix in §18;
7. treat a missing or malformed known member as an invalid complete discovery response;
8. on an invalid complete discovery response, expose no extension workspace or action from that response and continue the Base workbook through the Core 03 Base fallback surface.

A tolerant decoder MUST NOT repair a malformed known member, infer a missing field, or accept duplicate members using first-wins or last-wins parser behavior.

Profiles: base
Verified by: EXT-AC-091

**EXT-REQ-195**
The coordinated discovery transition MUST retain the existing Core 01 members `profile_id`, `claimed`, and `route_families[]` unchanged and add exactly:

- `claimable`;
- `contract_major`;
- `workspace_keys[]`;
- `capabilities[]`.

The transition is additive on `/api/v1/extensions`; no public v2 route is created by this revision. Core 01 remains the discovery owner. The Network Flow Activity owner MUST remove its competing `document_version` and singular `route_root` discovery requirements and MUST import `cartulary.extension_discovery_item.v1`. A server MUST NOT emit both the generic item and a profile-local discovery variant.

This requirement resolves the prior owner contradiction in favor of the Core 01 generic discovery owner. A repository MUST NOT adopt this NLSpec while the competing Network Flow item remains normative.

Profiles: base, network_flow_activity
Verified by: EXT-AC-092

**EXT-REQ-231**
The generic seven-member discovery item is the sole discovery contract after coordinated adoption. Because the currently adopted Network Flow Activity `contract_major=1` explicitly requires public `document_version` and singular `route_root`, removal of those members is a breaking correction unless the adopted-document authority records that the profile-local item was never effective in any conforming released implementation.

The default coordinated action is therefore:

- revise the Network Flow Activity owner to `document_version=2.0.0` and `contract_major=2`;
- replace its local discovery schema with `cartulary.extension_discovery_item.v1`;
- update Core 00 recognition, owner fragments, dependency snapshot, descriptors, implementation bindings, client support registry, discovery verification contracts and exact selectors, conformance manifests, and accounting;
- leave persisted Network Flow state version unchanged unless another owner change independently requires a state migration.

A contract-major-1 correction is valid only when one adopted authority finding proves all of these facts: no released server emitted the local shape; no released compatible client depended on it; no compatible retained Harness v2 evidence or claim treated it as valid; and the correction changes no previously conforming observable implementation. Without that exact finding, contract major `2` is mandatory. A patch or minor document-version increment is insufficient.

Profiles: base, network_flow_activity
Verified by: EXT-AC-110

# 16. Closed contribution-point registry

**EXT-REQ-087**
An extension MUST integrate with shared application owners only through one of the contribution kinds in Table 16-A. An unknown contribution kind is invalid and MUST block registry validation.

Profiles: base
Verified by: EXT-AC-021, EXT-AC-031, EXT-AC-070

**Table 16-A. Current contribution kinds**

| `kind` | Required target | Target grammar | Primary shared owner registry | `binding_requirement` |
| --- | --- | --- | --- | --- |
| `http_route_family` | `route_family` | `route_family_template_v1` | Core 01 route-family registry | `handler` |
| `incident_workspace` | `workspace_key` | Table 4-B | Core 01/Core 03 workspace registry | `handler` |
| `deployment_admin_panel` | `panel_key` | Table 4-B | Core 01/Core 04 administration-panel registry | `handler` |
| `authentication_entry` | `entry_key` | Table 4-B | Core 01/Core 04 authentication-entry registry | `handler` |
| `import_target` | `target_kind` | Table 4-B | Core 01 Imports target registry | `participant` |
| `extension_resource_kind` | `resource_kind` | Table 4-B | Named profile resource-kind registry | `none` |
| `websocket_invalidation` | `resource_kinds[]` | Table 4-B extension resource kind | Core 01/Core 03 invalidation registry | `participant` |
| `job_resource_ref_kind` | `resource_ref_kind` | Table 4-B | Core 01 common-job reference registry | `none` |
| `cross_owner_transaction_participant` | `participant_id` | Table 4-B | Invoking owner transaction-participant registry | `participant` |
| `snapshot_reporting_participant` | `participant_id` | Table 4-B | Snapshot and Reporting participant registry | `participant` |
| `incident_portability_participant` | `participant_id` | Table 4-B | Incident Portability participant registry | `participant` |
| `backup_restore_participant` | `participant_id` | Table 4-B | Core backup/restore participant registry | `participant` |

The closed `binding_requirement` values are `none`, `handler`, `participant`, and `worker`. A later contribution kind MAY use `worker` only after this table defines its owner registry and observable invocation contract. Omission behavior: no current contribution kind uses `binding_requirement='worker'`.

**EXT-REQ-088**
Every contribution declaration MUST contain exactly `kind`, `contribution_id`, and the kind-specific members in Table 16-B. Members belonging only to another variant are invalid.

Profiles: base
Verified by: EXT-AC-008, EXT-AC-021

**Table 16-B. Contribution declaration variants**

| `kind` | Additional exact members |
| --- | --- |
| `http_route_family` | `route_family` |
| `incident_workspace` | `workspace_key` |
| `deployment_admin_panel` | `panel_key` |
| `authentication_entry` | `entry_key` |
| `import_target` | `target_kind` |
| `extension_resource_kind` | `resource_kind` |
| `websocket_invalidation` | `resource_kinds[]` |
| `job_resource_ref_kind` | `resource_ref_kind` |
| `cross_owner_transaction_participant` | `participant_id`, `participant_contract_ref`, `participant_contract_sha256` |
| `snapshot_reporting_participant` | `participant_id`, `participant_contract_ref`, `participant_contract_sha256` |
| `incident_portability_participant` | `participant_id`, `participant_contract_ref`, `participant_contract_sha256` |
| `backup_restore_participant` | `participant_id`, `participant_contract_ref`, `participant_contract_sha256` |

The byte size of one contribution object MUST be measured by serializing that object in isolation under `extension_registry_canonical_json_v1`, including the required final LF. That byte size MUST be `1..16384`. `contributions[]` MUST contain `0..64` items. `websocket_invalidation.resource_kinds[]` MUST contain `1..64` items. A contribution that has no target items MUST be omitted rather than emitted with an empty required collection.

**EXT-REQ-089**
For every contribution declaration:

- `contribution_id` MUST use the owning profile prefix;
- the target token MUST satisfy the exact grammar named by Table 16-A;
- the target MUST resolve exactly once in the shared owner registry named by Table 16-A;
- zero target matches and multiple target matches are registry errors;
- the named profile contract MUST define its behavior;
- the shared owner MUST admit the contribution before this NLSpec is adopted or the profile becomes claimable;
- the implementation binding MUST satisfy the Table 16-A binding requirement.

Profiles: base
Verified by: EXT-AC-021, EXT-AC-031, EXT-AC-073, EXT-AC-080

**EXT-REQ-090**
The current revision MUST NOT expose a generic callback bus, arbitrary `before` or `after` hook, mutation-cancellation hook, HTTP middleware hook, database trigger registry, reflection-discovered lifecycle hook, arbitrary event subscriber, arbitrary browser component injection point, or runtime contribution-registration API.

Profiles: base
Verified by: EXT-AC-038, EXT-AC-070

**EXT-REQ-091**
A contribution declaration is metadata only. The implementation binding supplies packaged behavior, and the shared owner controls invocation. Deployment configuration, incident data, reference-pack content, owner-fragment content, and browser input MUST NOT supply contribution callbacks or executable functions.

Profiles: base
Verified by: EXT-AC-031, EXT-AC-070

**EXT-REQ-092**
An `import_target` contribution MUST use the Core 01 owner preview/apply façade and common final-commit boundary. The Import owner MUST NOT write extension-owned tables directly. A target unavailable because its profile is inactive MUST fail rather than fall back to another target.

Profiles: import
Verified by: EXT-AC-031, EXT-AC-055, EXT-AC-056, EXT-AC-057

**EXT-REQ-093**
A `websocket_invalidation` contribution MUST use the Core 01 `extension_resource_changed` envelope. It MUST NOT create a second read API, transmit raw resource values, or identify an extension resource by a Core `record_id`, visible label, route, or storage name.

Profiles: base
Verified by: EXT-AC-039, EXT-AC-059

**EXT-REQ-094**
A `snapshot_reporting_participant`, `incident_portability_participant`, `backup_restore_participant`, or `cross_owner_transaction_participant` MUST be typed and owner-admitted. `participant_contract_ref` MUST resolve to the exact shared-interface specialization, and `participant_contract_sha256` MUST match its canonical bytes. The profile claim alone MUST NOT authorize implicit data discovery, serialization, rendering, transport, or cross-owner mutation.

Profiles: base
Verified by: EXT-AC-033, EXT-AC-034, EXT-AC-035, EXT-AC-061, EXT-AC-062, EXT-AC-063

# 17. Route-family and dispatch integration

**EXT-REQ-095**
An extension profile MUST use separate route families under `/api/v1`. It MUST NOT overload Base workbook routes, reinterpret an existing Base request body, or intercept a Base route through middleware or fallback dispatch.

Profiles: base
Verified by: EXT-AC-022, EXT-AC-023, EXT-AC-031

**EXT-REQ-096**
Core 01 reserved-extension dispatch MUST use this order:

1. Base route match;
2. claimed extension-family route match;
3. reserved inactive extension-family match;
4. ordinary unknown route.

A reserved inactive match covers both `unclaimed` and `recognized_unclaimable`. It MUST return HTTP `404` with `error.code='extension_profile_not_claimed'` under the Core common error envelope. Public route dispatch MUST NOT expose a separate invented read-only or retired route mode. Claimability remains visible through discovery and startup configuration diagnostics.

Profiles: base
Verified by: EXT-AC-031, EXT-AC-032, EXT-AC-033, EXT-AC-087

**EXT-REQ-097**
A reserved inactive match MUST occur before family-specific authentication beyond the common route envelope, authorization, incident lookup, path-parameter validation, request-body validation, idempotency lookup, cursor validation, resource lookup, job admission, secret resolution, or external dependency access.

Profiles: base
Verified by: EXT-AC-031

**EXT-REQ-098**
A claimed-family request MUST proceed to the named profile's ordinary authentication, authorization, validation, resource visibility, capability, and error-precedence contract. It MUST NOT return `extension_profile_not_claimed` because the family has no resources, the caller lacks authorization, one capability is unavailable, or an external dependency failed.

Profiles: base
Verified by: EXT-AC-032, EXT-AC-033

**EXT-REQ-099**
A profile-local error-precedence table applies only after claimed-family dispatch. The current Network Flow Activity owner MUST remove its local inactive-profile precedence row and import Core 01 dispatch.

Profiles: base, network_flow_activity
Verified by: EXT-AC-033, EXT-AC-072

# 18. Extension workspace and client behavior

**EXT-REQ-100**
An extension workspace identity MUST use the Core 01 `sheet_ref.kind='extension_workspace'` shape with exact `extension_profile_id` and `workspace_key`. It MUST NOT use `id`, `view_schema_id`, `saved_view_id`, visible label, icon, route, component name, or inner-tab label as identity.

Profiles: base
Verified by: EXT-AC-034, EXT-AC-040

**EXT-REQ-101**
The workbook shell MUST render an extension workspace entry only when the exact identity is present in the deterministic intersection defined by EXT-REQ-212. Local component presence, route success observed earlier, cached membership, visible label, or discovery alone MUST NOT satisfy that intersection.

If the identity is absent from the intersection, the entry MUST be omitted. Explicit launch and persisted-pointer handling MUST follow EXT-REQ-197 and the Core 01/Core 03 fallback algorithm.

Profiles: base
Verified by: EXT-AC-034, EXT-AC-038, EXT-AC-093

**EXT-REQ-211**
Every packaged browser build MUST contain one canonical `cartulary.client_extension_support_registry.v1` object containing exactly:

- `schema_id`, exactly `cartulary.client_extension_support_registry.v1`;
- `client_build_id`;
- `client_build_class`, exactly `standard`;
- `asset_set_sha256`;
- `profiles[]`.

`client_build_id` MUST contain `1..160` ASCII bytes and match `[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}`. It has no default and MUST identify one immutable packaged browser build. `asset_set_sha256` MUST be a SHA-256 digest string and MUST equal `client_asset_set_sha256_v1` from EXT-REQ-220.

Each `profiles[]` row MUST contain exactly:

- `profile_id`;
- `supported_contract_majors[]`;
- `workspace_keys[]`;
- `capability_ids[]`;
- `public_schema_ids[]`.

`client_build_class` has no omission default; `standard` is the only admitted build class in this revision. `profiles[]` MUST contain `0..256` rows, reject duplicate profile IDs, and sort by ascending UTF-8 bytes of `profile_id`. A row is required for every canonical descriptor with `claimable=true` and a nonempty `workspace_keys[]`; omission of any such row is invalid. Omission of a row for a claimable profile without a workspace means that browser build supports no contract major, capability, or public schema for that profile. A profile row is valid only when the canonical descriptor has `claimable=true` and a non-null `contract_major`; a recognized-unclaimable profile MUST be omitted. In contract major `1`, each `supported_contract_majors[]` array MUST contain exactly one positive integer equal to the canonical descriptor's current `contract_major` for that profile. Supporting more than one profile contract major in one browser build is future-only until a later revision defines digest-bound historical owner inputs and decoder-selection behavior. `workspace_keys[]` MUST exactly equal the descriptor workspaces for every required row and contain `0..64` values; `capability_ids[]` MUST equal `[]`; `public_schema_ids[]` MUST contain `0..256` values. Each array must reject duplicates and sort by ascending UTF-8 bytes. The required Network Flow row MUST select contract major `2` and workspace `network_analysis` exactly.

The registry declares semantic support only. It MUST NOT contain React component names, source-package paths, chunk names, URLs, DOM selectors, display labels, icons, or callback identifiers. It MUST serialize under `extension_registry_canonical_json_v1`, contain `1..1048576` canonical bytes, and have digest `extension_client_support_registry_sha256_v1`. Build validation MUST fail when the registry advertises a profile absent from the canonical extension registry; when its one supported major differs from the canonical descriptor; when it advertises a workspace, capability, or public schema absent from that descriptor; when packaged assets do not match `asset_set_sha256`; or when an implementation binding claims client participation absent from this registry.

Profiles: base
Verified by: EXT-AC-108, EXT-AC-156

**EXT-REQ-220**
Every packaged browser build MUST produce one canonical `cartulary.client_asset_set_manifest.v1` object containing exactly:

- `schema_id`, exactly `cartulary.client_asset_set_manifest.v1`;
- `assets[]`.

Each `assets[]` row MUST contain exactly:

- `logical_path`;
- `byte_length`;
- `sha256`.

The manifest MUST enumerate every immutable regular file that the packaged application can serve through its static browser-asset mapping, including static HTML, JavaScript, CSS, fonts, images, and other browser-consumable bytes. A file absent from the serving mapping MUST NOT appear. A served source map or debug asset MUST appear; omitting it from the production package is the only way to omit it from the manifest.

`logical_path` MUST be a repository-independent POSIX path relative to the browser asset root, contain `1..512` UTF-8 bytes, and reject a leading slash, backslash, NUL, empty segment, `.`, `..`, duplicate normalized path, symlink, and non-regular file. `byte_length` MUST equal the exact file length and be a JSON integer in `0..1073741824`. `sha256` MUST hash the exact raw file bytes. Rows MUST sort by ascending UTF-8 bytes of `logical_path`.

The client support registry and the asset-set manifest themselves MUST be packaged as application contract artifacts outside the static browser-asset mapping and MUST NOT appear in `assets[]`; this prevents a digest cycle. A dynamically generated browser-bootstrap response that carries their values is not a static asset and MUST NOT be included. Every static byte referenced by such a bootstrap response remains included through its own row.

`client_asset_set_sha256_v1` MUST equal the lowercase SHA-256 digest of the manifest's canonical byte form under `extension_registry_canonical_json_v1`. Build and startup validation MUST verify every listed raw file against its row, reject an extra served static file, reject a missing listed file, and require the client support registry's `asset_set_sha256` to equal this digest. The manifest MUST contain `1..16777216` canonical bytes and `1..65536` asset rows. A packaged browser build with no served static asset is invalid under the current Core 01 packaged-browser boundary.

Profiles: base
Verified by: EXT-AC-108, EXT-AC-127

**EXT-REQ-212**
Core 01 MUST add one required `data.extension_workspace_availability` member to the successful common response envelope for:

```text
GET /api/v1/incidents/{incident_id}/workbook-startup
```

The member MUST conform to `cartulary.extension_workspace_availability.v1` and contain exactly:

- `schema_id`, exactly `cartulary.extension_workspace_availability.v1`;
- `incident_id`, exactly the addressed incident identity;
- `workspaces[]`.

`workspaces[]` MUST always be present and contain `0..64` closed rows. Each row MUST contain exactly:

- `extension_profile_id`;
- `workspace_key`.

Rows MUST reject duplicate identities and sort by `extension_profile_id`, then `workspace_key`, ascending UTF-8 bytes. The response MUST include a row only when the profile is in the resolved claim set and the current caller is authorized at response construction time to open that workspace shell for the addressed incident. The object MUST disclose no denial reason, required role, hidden resource count, or profile-local authorization detail.

`workspaces[]` MUST be present even when empty. Every successful response from this route MUST use `Cache-Control: no-store`; an ETag validator, `If-None-Match` reuse, and HTTP `304` response are forbidden. An error response MUST use the Core error envelope and MUST NOT carry `extension_workspace_availability`. An older compatible client that does not know the additive member MUST ignore it under the Core tolerant-consumer rule.

The client MUST compute:

```text
renderable_extension_workspaces =
    discovered claimed workspace identities
    INTERSECT client-supported profile-major workspace identities
    INTERSECT extension_workspace_availability.workspaces
```

The intersection is by exact `(extension_profile_id, workspace_key)`. A client version that requires the member MUST treat a missing member, malformed object, incident mismatch, stale response, or invalid client-support registry as an extension protocol defect: it MUST clear incident-local extension state, hide every extension workspace, retain Base behavior, and MUST NOT enter an automatic retry loop. A later explicit user action or a Core 03 current-authorization invalidation event MAY initiate one new workbook-startup request; when neither occurs, no request is initiated. Discovery does not authorize a workspace, and hiding a workspace does not replace route-time authorization.

For each `(client_instance_id, incident_id)`, the client MUST maintain:

- `extension_availability_epoch_id`, exactly 64 lowercase hexadecimal characters generated from 32 cryptographically secure random bytes; and
- `extension_availability_generation`, an unsigned 64-bit integer initialized to `0`.

Generation reservation is one linearizable client-local operation. Before a workbook-startup request, before every HTTP request initiated from an extension route contribution for extension workspace or resource state, and immediately upon any current-authorization invalidation affecting that incident, the client MUST atomically reserve exactly one next generation and tag all resulting local request work with the reserved `(extension_availability_epoch_id, extension_availability_generation)` tuple. Concurrent reservations MUST produce distinct, strictly increasing generations in reservation order. A response is eligible to update client state only when both tagged values equal the current tuple; an eligible response MUST then be processed under its owner response contract. A response tagged with another epoch or a lower generation MUST be ignored without rendering, caching, queue replay, draft mutation, or automatic retry.

On a reservation when the current generation equals `18446744073709551615`, the same linearizable operation MUST generate a new epoch ID, reset the generation to `1`, clear incident-local extension state, and return that new tuple; no request may observe an intermediate old-epoch generation or duplicate generation. It MUST NOT change `client_instance_id`. The epoch and generation are local-only, non-authorizing, non-persisted across a client-instance restart, and never transmitted. If secure randomness is unavailable when an epoch is required, the client MUST disable extension rendering for that client instance and retain Base behavior.

An explicit launch, home, default, presence, or deep-link target absent from the intersection MUST use the Core 03 unavailable-target fallback. When authorization is lost while a workspace is open, the client MUST advance the tuple, discard loaded profile data, invalidate every pending extension-route result tagged with an older tuple, ignore every later response from that tuple, remove the workspace from navigation, execute Base fallback, and retain no cached content that can be reopened without a new authorized response. The client MAY additionally issue transport-level cancellation; omission means tuple invalidation and response discard remain mandatory.

Epoch creation, rollover, invalidation, and extension fallback MUST preserve the Core WebSocket resume identity, `client_instance_id`, Base drafts, and the Core pending queue. Extension state MUST NOT be stored in or recovered from those Base or transport identities.

Profiles: base
Verified by: EXT-AC-109, EXT-AC-139, EXT-AC-156

**EXT-REQ-102**
An extension workspace is not a Base built-in tab, `view_schema`, system view, saved view, or member of the Base surface registry. Claiming any profile MUST NOT change the five Base built-in tabs or the canonical Base `view_schema_id` ordering.

Profiles: base
Verified by: EXT-AC-035, EXT-AC-059

**EXT-REQ-103**
Packaged extension UI assets MAY be code-split. Omission behavior: when they are not code-split, they remain packaged local assets and MUST still perform no profile work merely because the Base workbook opens. Opening an incident workbook or switching among Base surfaces MUST NOT require loading extension resource data, starting profile-specific browser polling, calling extension external services, or waiting for extension route responses.

Profiles: base
Verified by: EXT-AC-036, EXT-AC-037, EXT-AC-042

**EXT-REQ-104**
Extension resource loading MUST begin only after the user selects the extension workspace or invokes an explicit extension action that requires the resource. Discovery metadata MAY be loaded during ordinary application bootstrap. Omission behavior: when discovery is not loaded at bootstrap, the client MUST load and validate it before presenting extension navigation.

Profiles: base
Verified by: EXT-AC-036, EXT-AC-037

**EXT-REQ-105**
When extension work cannot complete immediately, the UI MUST display the exact user-facing semantic state defined by the named profile owner or common job contract. A claimable profile that exposes long-running work MUST define that state vocabulary and every transition that can reach it. The UI MUST NOT invent an undeclared state or expose reducers, RPC method names, queue internals, storage keys, SQL, object-store paths, or synchronization implementation details as user-facing status.

Profiles: base
Verified by: EXT-AC-036, EXT-AC-039

**EXT-REQ-106**
Client cleanup MUST distinguish authoritative server cache, derived cache, cursor or continuation state, pending request, queued mutation, optimistic mutation, and unsent user edit. The exact consequences are defined by EXT-REQ-201. A client MUST NOT use the phrase `pending owner data` as an untyped cleanup category.

Profiles: base
Verified by: EXT-AC-038, EXT-AC-039, EXT-AC-096

**EXT-REQ-107**
Unknown future profiles, capabilities, workspace keys, discovery members, error codes, extension resource kinds, and invalidation reasons MUST follow EXT-REQ-197. They MUST NOT fail the Base workbook or WebSocket solely because they are unknown, and they MUST NOT be executed or treated as authorization.

Profiles: base
Verified by: EXT-AC-023, EXT-AC-039, EXT-AC-093

**EXT-REQ-108**
Current extension workspaces MUST NOT use `field_key` in presence or startup identity. A later writable extension-workspace field contract requires an explicit Core 01 amendment and a new profile contract major when it changes public interaction semantics.

Profiles: base
Verified by: EXT-AC-040

**EXT-REQ-109**
An extension MUST NOT introduce mandatory checklists, challenge-and-response prompts, approval dialogs, detached forms, external lookups, or profile configuration steps into ordinary Base row creation, inline edit, paste, correction, or rough capture.

Profiles: base
Verified by: EXT-AC-036, EXT-AC-039

**EXT-REQ-197**
A client that recognizes `profile_id` but does not support the discovered `contract_major` MUST enter state `unsupported_contract_major` for that profile. In that state it MUST:

- omit every profile workspace and action;
- make no profile route call;
- decode no extension resource body;
- discard profile-scoped cursors, query results, graph results, cached resources, selections, optimistic mutations, and queued mutations;
- preserve server-side workbook home and default pointers unchanged;
- execute the Core 03 Base fallback surface algorithm for an explicit, home, default, presence, or deep-link target;
- present a safe `extension_version_unsupported` state only when a user explicitly attempts to open the profile.

Unknown values MUST follow Table 18-A.

Profiles: base
Verified by: EXT-AC-093

**Table 18-A. Unknown-value handling**

| Unknown value | Required client behavior |
| --- | --- |
| Unknown `profile_id` in valid discovery | Ignore behaviorally after validating the generic item. |
| Unknown discovery member | Ignore under `decode_extension_discovery_item_v1`. |
| Unknown capability ID | Do not expose or invoke the capability. |
| Unknown workspace key | Do not render or navigate to the workspace. |
| Unknown public error code | Render generic failure from HTTP status and `retryable`; do not branch on message text. |
| Unknown invalidation reason for a known resource kind | Treat as `invalidate`, purge that resource's derived state, and refetch after current authorization succeeds. |
| Unknown resource kind for a known profile | Purge all cached resource, cursor, query, graph, and selection state for that profile; refetch discovery; infer no Core resource effect. |
| Unknown profile in a WebSocket event | Validate the generic envelope, advance the replay high-water mark, and ignore the event behaviorally. |
| Known profile with unsupported contract major | Apply `unsupported_contract_major`; do not process the event semantically. |
| Unknown contribution metadata in runtime data | Reject; contribution metadata is not a compatible public runtime extension container. |

**EXT-REQ-201**
The client state categories and cleanup behavior in Tables 18-B and 18-C are closed.

Profiles: base
Verified by: EXT-AC-096

**Table 18-B. Client state categories**

| Category | Definition |
| --- | --- |
| `authoritative server cache` | Last successfully authorized server resource representation retained by the client. |
| `derived cache` | Client-computed query, graph, grouping, display, or lookup state reproducible from authorized server data. |
| `cursor` | Opaque continuation or replay position bound to an authorized query or event stream. |
| `pending request` | Request transmitted or ready for transmission whose authoritative result is not known. |
| `queued mutation` | Mutation retained for automatic replay under the Core transient-failure queue contract. |
| `optimistic mutation` | Client-visible provisional effect awaiting authoritative acceptance. |
| `unsent user edit` | User-authored text or structured input not yet admitted as a queued mutation or server request. |

**Table 18-C. Client cleanup matrix**

| Condition | Authoritative and derived cache | Pending request, queued mutation, and optimistic mutation | Unsent user edit |
| --- | --- | --- | --- |
| Transient transport loss | Preserve. | Preserve and replay under the Core FIFO retry contract. | Preserve; replay only after it becomes an admitted mutation. |
| Session expired or revoked while current incident authorization has not been rederived | Hide protected state; preserve it only in volatile memory and never reveal it to a newly authenticated user before current authorization is rederived. | Pause; do not transmit or replay. | Preserve in volatile memory; no replay until reauthentication and reauthorization both succeed. |
| Write permission lost while read permission remains | Refresh authorized read state. | Cancel writes and remove optimistic effects. | Preserve as visibly non-authoritative, copyable draft; a fresh user action is required after write permission returns. |
| Complete incident read authorization lost | Purge all profile resources, labels, cursors, handles, query/graph state, selection, queues, optimistic effects, and drafts for that incident. | Cancel and do not replay. | Discard. |
| Resource-specific authorization lost or resource removed | Purge all state for the resource and dependent derived state. | Cancel resource-scoped work and do not replay. | Discard resource-scoped drafts. |
| Profile becomes inactive, retired, or unsupported on a new discovery/session boundary | Purge all profile-scoped resources, cursors, queues, optimistic effects, and drafts; execute Base fallback. | Cancel and do not replay. | Discard profile-scoped drafts. |
| Incident becomes closed | Preserve currently authorized read state. | Cancel writes and remove optimistic effects. | Preserve rejected drafts as locally visible and copyable; do not replay while closed or automatically after reopen. |
| Hard refresh, tab close, or browser restart | No durability guarantee in contract major `1`. | Terminate volatile work. | MAY be lost; omission behavior is no browser-local durable draft store. |

No matrix row authorizes client state to bypass route-time authorization. A client MUST NOT automatically replay a copy-only or closed-incident draft after permission or lifecycle state changes; a fresh user action is required.

# 19. Extension data ownership

**EXT-REQ-110**
Every profile that owns a durable resource or public interface MUST define, in its named owner contract:

- every public and internal owner interface that affects callers or shared owners;
- request and response schemas;
- omission, explicit-`null`, and default behavior;
- scalar, collection, byte, time, and work limits;
- stable resource identifier grammar and generation;
- normalization, canonicalization, ordering, pagination, and cursor behavior;
- incident-scoped or deployment-scoped ownership;
- authoritative fields and source of truth;
- mutable and immutable fields;
- optimistic concurrency and idempotency, or explicit unavailability of mutation;
- complete authorization and disclosure matrices;
- lifecycle states and complete transition matrix;
- delete, restore, supersession, retention, and purge behavior;
- error codes, reason codes, detail schemas, precedence, HTTP or internal status, and retryability;
- audit occurrences and safe fields;
- revision behavior or explicit absence of per-resource revisions;
- state schema version, migration lineage, state-presence declaration, and migration definitions;
- durable job kinds, proof, cancellation, reconciliation, and terminal result behavior;
- derived cache and projection status;
- object-storage use and publication boundary;
- client UI states and cleanup consequences;
- security, secret, trust, and egress behavior;
- backup and restore classification;
- portability and Snapshot/Reporting participation;
- binary acceptance criteria, active Harness v2 verification IDs, and exact-selector coverage obligations.

Profiles: base
Verified by: EXT-AC-059, EXT-AC-060, EXT-AC-061, EXT-AC-064, EXT-AC-073, EXT-AC-095

**EXT-REQ-111**
An extension-owned resource is not a Core record envelope, `view_schema`, saved view, system view, entity mention, indicator observation, evidence record, projection source record, or Base workbook surface unless the primary owner named in Table 1-A is amended explicitly.

Profiles: base
Verified by: EXT-AC-035, EXT-AC-059, EXT-AC-060

**EXT-REQ-112**
An extension MUST NOT directly write another owner's authoritative tables, Core projection tables, Core record envelopes, saved views, common job tables except through the common job interface, audit tables except through an admitted audit participant, or indicator state except through the Core indicator participant.

Profiles: base
Verified by: EXT-AC-031, EXT-AC-055, EXT-AC-058, EXT-AC-060

**EXT-REQ-113**
A cross-owner effect MUST use one typed owner façade or transaction participant declared by the shared owner. Parser-shaped rows, vendor-grid coordinates, filesystem paths, object-store keys, visible labels, or implementation-private storage DTOs MUST NOT cross that owner boundary as authoritative domain state.

Profiles: base
Verified by: EXT-AC-031, EXT-AC-055, EXT-AC-059

**EXT-REQ-114**
A profile MAY retain derived caches or projections only when its owner declares them rebuildable and non-authoritative. Omission behavior: when no derived state is declared, the profile retains authoritative state only. Backup, restore, portability, state-presence, and state-version admission MUST distinguish authoritative state from rebuildable state.

Profiles: base
Verified by: EXT-AC-061, EXT-AC-062, EXT-AC-063

Every `extension_versioned` profile owner MUST publish one `cartulary.extension_state_presence_manifest.v1` under §21. The manifest MUST use logical authoritative database-family and object-reference-family IDs, not physical table, bucket, key, or package names. Physical binding of those logical families remains implementation-support manifest data owned by the storage and backup implementation.

**EXT-REQ-200**
Every claimable profile MUST have one generated `cartulary.extension_contract_closure_catalog.v1` and one matching `contract_closure[]` resolution in its `cartulary.extension_conformance_manifest.v1`. Closure MUST operate at closure-item level, not by asserting that one broad Table 19-A category is complete.

A closure item is closed only by resolved exact owner locators or by one item-specific adopted `not_applicable` reason permitted by that catalog item. Broad prose, a directory path, a heading-only reference, generated code, an implementation test without an owner requirement, a category-level assertion, or a `TODO` placeholder does not close an item.

Profiles: base
Verified by: EXT-AC-095, EXT-AC-121

**Table 19-A. Required profile contract-closure categories**

| Category | Required owner subject |
| --- | --- |
| `public_interfaces` | Complete route, operation, contribution, and shared-owner interface inventory. |
| `request_response_schemas` | Closed request, response, event, job-result, durable proof, participant-context, and participant-result schemas. |
| `defaults_omission_null` | Every default, omission, empty, and explicit-`null` case. |
| `scalar_collection_bounds` | Every scalar, byte, item, nesting, time, and work bound. |
| `identity_canonicalization_ordering` | Identifiers, normalization, canonical bytes, digests, comparators, and output ordering. |
| `pagination` | Cursor, ordering, continuation, complete-collection, and no-pagination behavior. |
| `authorization` | Complete role, scope, current-authorization, and disclosure behavior. |
| `idempotency_concurrency` | Idempotency identity, replay, optimistic concurrency, serialization, commit, cancellation, and race behavior. |
| `errors_precedence_retry` | Error schemas, precedence, retryability, timeouts, and partial-result prohibition. |
| `resource_lifecycle_retention` | States, transitions, delete, restore, retirement, retention, and purge. |
| `jobs_reconciliation` | Job kinds, terminal results, commit proof, cancellation, crash recovery, and inactive reconciliation. |
| `state_migration` | State presence, metadata, lineage, definitions, locks, forward migration, final validation, and downgrade rejection. |
| `client_ui_states` | Workspace availability, loading, progress, empty, error, unsupported, cleanup, and fallback states. |
| `security_secrets_egress` | Trust, inert input, secret/file/trust-material handling, upload, and external dependency boundaries. |
| `audit_observability` | Audit occurrences, safe fields, telemetry, and secret-safe diagnostics. |
| `backup_restore` | Authoritative/derived classification, physical backup inclusion, restore validation, and rebuild behavior. |
| `portability` | Export/import mode, state-presence matrix, target compatibility, inactive behavior, and errors. |
| `snapshot_reporting` | Participation mode, inputs, outputs, empty/omit behavior, redaction, retained inactive state, and errors. |
| `conformance_evidence` | Required observable-branch inventory, active Harness v2 verification-contract identity, exact-selector catalog coverage, and retained-evidence boundary. |

**EXT-REQ-226**
`cartulary.extension_contract_closure_catalog.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_contract_closure_catalog.v1`;
- `profile_id`;
- `contract_major`;
- `owner_document_sha256`;
- `items[]`.

Each `items[]` row MUST contain exactly:

- `closure_item_id`;
- `category`;
- `subject_kind`;
- `subject_id`;
- `allowed_not_applicable_reason_codes[]`.

`subject_kind` MUST equal `baseline`, `owner_requirement`, `configuration_key`, `public_schema`, `contribution`, `job_kind`, `migration`, `initialization`, `state_family`, `backup_codec`, or `verification_contract`. `category` MUST use Table 19-A. `subject_id` MUST be the exact stable owner identity for the subject and contain `1..512` UTF-8 bytes. `allowed_not_applicable_reason_codes[]` MUST contain `0..8` unique Table 27-A2 tokens sorted by ascending UTF-8 bytes. `[]` means the item can be closed only as `specified`.

`closure_item_id` MUST equal `extclosure:` followed by the first 32 lowercase hexadecimal characters of SHA-256 over the canonical JSON object containing exactly `profile_id`, `contract_major`, `category`, `subject_kind`, and `subject_id`, serialized under `extension_registry_canonical_json_v1` without the final LF.

`derive_extension_contract_closure_catalog_v1` MUST construct the required item set as the union of:

1. every fixed baseline item in Table 19-B;
2. one `owner_requirement` item for every profile-owner `req` anchor, repeated once for each category in that anchor's `closure_categories[]`;
3. one `configuration_key` item for every row in the profile configuration contract, under exactly `defaults_omission_null`, `scalar_collection_bounds`, and `security_secrets_egress`;
4. one `public_schema` item for every declared public schema, under exactly `request_response_schemas`, `defaults_omission_null`, `scalar_collection_bounds`, and `identity_canonicalization_ordering`;
5. one `contribution` item for every contribution, under exactly the Table 19-C categories for its contribution kind;
6. one `job_kind` item for every job-kind contract, covering `jobs_reconciliation`, `idempotency_concurrency`, `errors_precedence_retry`, and `resource_lifecycle_retention`;
7. one `migration` item for every migration definition, covering `state_migration`, `errors_precedence_retry`, and `audit_observability`;
8. one `initialization` item for the state-initialization definition, covering `state_migration`, `defaults_omission_null`, `scalar_collection_bounds`, and `security_secrets_egress`;
9. one `state_family` item for every logical state family, covering `resource_lifecycle_retention`, `backup_restore`, and the selected portability and Snapshot/Reporting categories;
10. one `backup_codec` item for every physical binding codec, covering `backup_restore`, `identity_canonicalization_ordering`, `scalar_collection_bounds`, and `errors_precedence_retry`;
11. one `verification_contract` item for every required applicable Harness v2 verification ID, covering `conformance_evidence`.

The profile owner MUST NOT remove, rename, merge, suppress, or mark optional a derived item. Duplicate item identities are invalid. Items MUST sort by `category`, `subject_kind`, `subject_id`, then `closure_item_id`, using ascending UTF-8 bytes. The catalog MUST contain `1..65536` items and `1..33554432` canonical bytes. `extension_contract_closure_catalog_sha256_v1` is the lowercase SHA-256 digest of the canonical byte form.

A Core-owned behavior is not `not_applicable`. The manifest MUST close that item as `specified` with the exact Core owner locator. Generated artifacts and implementation bindings are permitted as evidence inputs but MUST NOT be the sole locator proving a behavior. Every row generated from items 2 through 11 MUST have `allowed_not_applicable_reason_codes=[]`; an owner-authored override, omission, or `not_applicable` reason for such a row is invalid. Only the fixed baseline rows in Table 19-B may carry nonempty reason arrays, and each may contain only the exact enumerated reason for that row. An unknown reason token is invalid.

Profiles: base
Verified by: EXT-AC-121, EXT-AC-148

**Table 19-B. Fixed baseline closure items**

| Baseline `subject_id` | Category | Required closure question | Permitted not-applicable reason codes |
| --- | --- | --- | --- |
| `interface_inventory` | `public_interfaces` | Are every public and shared-owner callable interface and contribution enumerated? | `no_public_interface` |
| `wire_schema_family` | `request_response_schemas` | Are every request, response, event, job result, proof, participant context, and participant result closed? | `no_public_interface` |
| `omission_matrix` | `defaults_omission_null` | Are omission, default, empty, explicit-null, and no-value states exhaustive? | None |
| `resource_limits` | `scalar_collection_bounds` | Are scalar, collection, byte, nesting, time, and work limits exhaustive? | None |
| `identity_and_bytes` | `identity_canonicalization_ordering` | Are identity, normalization, serialization, digest, and ordering rules deterministic? | None |
| `collection_continuation` | `pagination` | Is pagination defined, or is every collection explicitly non-pageable? | `no_pagination` |
| `current_authorization` | `authorization` | Are current authorization, disclosure, and revalidation behavior complete? | `no_public_interface` |
| `commit_and_replay` | `idempotency_concurrency` | Are replay, serialization, final commit, and cancel/commit races complete? | `read_only_profile`, `no_public_interface` |
| `error_contract` | `errors_precedence_retry` | Are error shape, selection, retryability, timeout, and partial-result behavior complete? | None |
| `lifecycle_and_retention` | `resource_lifecycle_retention` | Are lifecycle, deletion, restoration, retirement, retention, and purge complete? | `no_durable_state` |
| `job_contract` | `jobs_reconciliation` | Are job kind, proof, cancellation, crash, and inactive reconciliation complete? | `no_jobs` |
| `state_contract` | `state_migration` | Are state presence, metadata, migration, final validation, and downgrade complete? | `no_durable_state` |
| `client_availability` | `client_ui_states` | Are availability, loading, empty, error, unsupported, cleanup, and fallback complete? | `no_client_surface` |
| `trust_boundary` | `security_secrets_egress` | Are inert input, secrets, file references, uploads, and egress fail-closed? | None |
| `safe_evidence` | `audit_observability` | Are audit, telemetry, diagnostic, and redaction fields complete? | None |
| `recovery_contract` | `backup_restore` | Are physical bindings, backup, restore validation, and rebuild complete? | `no_durable_state` |
| `portability_contract` | `portability` | Is the selected portability mode and its state matrix complete? | None |
| `snapshot_reporting_contract` | `snapshot_reporting` | Is the selected Snapshot/Reporting mode and its state matrix complete? | None |
| `scenario_coverage` | `conformance_evidence` | Are all required boundary and semantic branches mapped to active Harness v2 verification IDs and exact-selector catalog coverage, with execution closure remaining in retained v2 row evidence? | None |

`None` in the final column means `allowed_not_applicable_reason_codes=[]`.

**Table 19-C. Exhaustive contribution closure-category mapping**

| Contribution kind | Exact categories |
| --- | --- |
| `http_route_family` | `public_interfaces`, `request_response_schemas`, `authorization`, `errors_precedence_retry`, `security_secrets_egress`, `audit_observability`, `conformance_evidence` |
| `incident_workspace` | `public_interfaces`, `client_ui_states`, `authorization`, `errors_precedence_retry`, `resource_lifecycle_retention`, `conformance_evidence` |
| `deployment_admin_panel` | `public_interfaces`, `client_ui_states`, `authorization`, `security_secrets_egress`, `audit_observability`, `conformance_evidence` |
| `authentication_entry` | `public_interfaces`, `client_ui_states`, `authorization`, `security_secrets_egress`, `audit_observability`, `conformance_evidence` |
| `import_target` | `public_interfaces`, `request_response_schemas`, `idempotency_concurrency`, `errors_precedence_retry`, `resource_lifecycle_retention`, `security_secrets_egress`, `audit_observability`, `conformance_evidence` |
| `extension_resource_kind` | `public_interfaces`, `request_response_schemas`, `resource_lifecycle_retention`, `backup_restore`, `portability`, `snapshot_reporting`, `conformance_evidence` |
| `websocket_invalidation` | `public_interfaces`, `request_response_schemas`, `authorization`, `identity_canonicalization_ordering`, `errors_precedence_retry`, `conformance_evidence` |
| `job_resource_ref_kind` | `public_interfaces`, `request_response_schemas`, `jobs_reconciliation`, `resource_lifecycle_retention`, `conformance_evidence` |
| `cross_owner_transaction_participant` | `request_response_schemas`, `scalar_collection_bounds`, `identity_canonicalization_ordering`, `idempotency_concurrency`, `errors_precedence_retry`, `security_secrets_egress`, `audit_observability`, `conformance_evidence` |
| `snapshot_reporting_participant` | `request_response_schemas`, `scalar_collection_bounds`, `identity_canonicalization_ordering`, `authorization`, `errors_precedence_retry`, `security_secrets_egress`, `snapshot_reporting`, `conformance_evidence` |
| `incident_portability_participant` | `request_response_schemas`, `scalar_collection_bounds`, `identity_canonicalization_ordering`, `authorization`, `idempotency_concurrency`, `errors_precedence_retry`, `security_secrets_egress`, `portability`, `conformance_evidence` |
| `backup_restore_participant` | `request_response_schemas`, `scalar_collection_bounds`, `identity_canonicalization_ordering`, `errors_precedence_retry`, `state_migration`, `security_secrets_egress`, `backup_restore`, `conformance_evidence` |

The mapping is closed. Adding a contribution kind or changing its exact category set requires an amendment to Tables 16-A and 19-C in the same coordinated owner revision; a generator or owner fragment MUST NOT add, remove, or infer categories.

# 20. Cross-owner transaction semantics

**EXT-REQ-115**
An operation that commits extension state together with Core, import, job, audit, idempotency, indicator, release, portability, object-reference, or other owner state MUST define one final atomic database commit boundary.

Profiles: base
Verified by: EXT-AC-055, EXT-AC-056, EXT-AC-057, EXT-AC-058, EXT-AC-089

**EXT-REQ-116**
The final atomic commit MUST include every Table 20-A item whose inclusion condition evaluates true. Omitting such a participant is a contract defect; the implementation MUST fail before public resource publication rather than commit a partial result.

Profiles: base
Verified by: EXT-AC-055, EXT-AC-056, EXT-AC-058

**Table 20-A. Common final-commit participants**

| Participant | Inclusion condition |
| --- | --- |
| Extension resource mutation | The operation creates, updates, deletes, or binds an extension resource. |
| Core owner mutation | The operation changes Core source state. |
| Import session/unit state | The operation is admitted by the Import owner. |
| Idempotency success record | The public or owner route defines idempotency. |
| Common job terminal publication | The operation runs as a common job. |
| Extension job commit proof | The operation runs as a durable extension-owned or extension-producing job. |
| Domain audit outbox or occurrence | The owner requires a domain audit occurrence. |
| Change set or revision | A participating Core owner requires history. |
| Indicator participant | The operation finds, creates, or binds Core indicator state. |
| Collaboration invalidation outbox | The committed effect requires an `extension_resource_changed` event. |
| Authoritative object reference and digest | The operation publishes authoritative non-database bytes. |
| Portability/reporting/release binding | The operation creates a durable cross-owner reference. |

**EXT-REQ-117**
All participant validation MUST complete before final commit. A third-party network call, browser callback, or non-transactional external API MUST NOT be part of the atomic commit. When an operation requires post-commit external work, it MUST use the durable outbox or job contract named by the operation's primary owner and MUST NOT report the external effect as complete before that contract's terminal success.

Profiles: base
Verified by: EXT-AC-056, EXT-AC-057, EXT-AC-067

**EXT-REQ-118**
A failure or cancellation observed before final commit MUST leave no queryable extension resource and no terminal success. A crash, cancellation, or worker restart after final commit MUST recover to exactly one committed resource and exactly one terminal success without duplicating audit, idempotency, object-reference, or participant effects.

Profiles: base
Verified by: EXT-AC-052, EXT-AC-053, EXT-AC-056, EXT-AC-057, EXT-AC-088, EXT-AC-089

**EXT-REQ-119**
Participant ordering before commit MUST be deterministic for validation and diagnostics: participants sort by exact `participant_id` ascending UTF-8 bytes. Database statement ordering inside one atomic transaction is implementation freedom only when it cannot change observable results, errors, lock acquisition, deadlock behavior, publication timing, or recovery.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-058

**EXT-REQ-219**
Every cross-owner final commit MUST use the shared typed transaction-participant protocol. A `cross_owner_transaction_participant` contribution MUST resolve to one canonical `cartulary.extension_transaction_participant_contract.v1` containing exactly:

- `schema_id`, exactly `cartulary.extension_transaction_participant_contract.v1`;
- `participant_id`;
- `owner_profile_id`, a profile ID or `null` for a Core-owned participant;
- `participant_input_schema_id`;
- `prepare_algorithm_id`;
- `validation_algorithm_id`;
- `write_algorithm_id`;
- `serialization_key_kinds[]`;
- `owned_state_family_ids[]`;
- `error_contract_ref`.

The participant ID and owner profile MUST equal the contribution. Schema and algorithm IDs MUST resolve through the owner manifest and implementation binding. `serialization_key_kinds[]` MUST contain `1..32` unique profile-prefixed or Core-owner-prefixed key-kind tokens sorted by UTF-8 bytes. `owned_state_family_ids[]` MUST contain `1..64` unique logical state-family IDs sorted by UTF-8 bytes. `error_contract_ref` MUST resolve to the complete participant error mapping. The canonical contract MUST contain `1..1048576` bytes and have digest `extension_participant_contract_sha256_v1`.

The resolved participant set for one transaction MUST contain `1..16384` participants. Zero participants and participant `16385` MUST fail before input construction or transaction creation. There is no unbounded or empty legacy mode.

The shared logical interfaces are:

- `cartulary.extension_transaction_participant_context.v1`;
- `cartulary.extension_transaction_participant_prepare_result.v1`;
- `cartulary.extension_transaction_participant_validation_result.v1`;
- `cartulary.extension_transaction_participant_write_result.v1`;
- `cartulary.extension_transaction_participant_finding.v1`.

The context MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_transaction_participant_context.v1`;
- `phase`;
- `operation_id`;
- `participant_id`;
- `owner_profile_id`, a profile ID or `null`;
- `normalized_request_sha256`;
- `participant_input`;
- `cancellation_requested`, Boolean;
- `deadline_monotonic_ns`;
- `transaction_access`.

`phase` MUST equal `prepare`, `validate`, or `write`. `participant_input` MUST conform to the contract's `participant_input_schema_id`, contain no value outside the participant's declared operation scope, and contain `1..67108864` canonical bytes per participant. The aggregate participant input constructed at step 3 MUST contain `1..67108864` canonical bytes. `deadline_monotonic_ns` MUST be a process-local JSON integer in `1..9223372036854775807` and MUST represent the Table 9-E deadline. `operation_id` MUST use the exact Core operation-ID scalar imported by the invoking owner. `transaction_access` MUST be `null` for `prepare`, a shared read-only transaction capability for `validate`, and a shared read-write transaction capability for `write`.

The transaction capability is process-local and MUST NOT be serialized, persisted, logged, or returned. The read-only capability MUST read only declared mutable preconditions and declared logical state families. The read-write capability MUST mutate only `owned_state_family_ids[]`. Each capability MUST NOT acquire an undeclared lock, commit, roll back, open another transaction, perform egress, read a secret not present in `participant_input`, or access another state family.

The prepare result MUST contain exactly `schema_id`, `participant_id`, and `serialization_keys[]`. Each serialization key MUST contain exactly `participant_id`, `key_kind`, and `key`. `key_kind` MUST occur in the participant contract. `key` MUST be an internal non-secret UTF-8 scalar of `1..512` bytes. Keys MUST reject duplicates and sort by `participant_id`, then `key_kind`, then ascending UTF-8 bytes of `key`. `serialization_keys[]` MUST contain `0..1024` rows per participant, and their aggregate union MUST contain `0..4096` rows. Each canonical prepare result MUST contain `1..67108864` bytes, and aggregate canonical prepare-result bytes across all participants MUST contain `1..67108864` bytes.

The validation result MUST be exactly one closed variant:

```json
{"schema_id":"cartulary.extension_transaction_participant_validation_result.v1","participant_id":"...","status":"valid","findings":[]}
```

or:

```json
{"schema_id":"cartulary.extension_transaction_participant_validation_result.v1","participant_id":"...","status":"invalid","findings":[...]}
```

`findings[]` MUST contain `0..256` `cartulary.extension_transaction_participant_finding.v1` objects. Each finding MUST contain exactly `path`, `reason_code`, `message`, and `details`. `path` MUST use `extension_diagnostic_path_v1` against the normalized participant input. `reason_code`, `message`, and the exact closed `details` shape MUST resolve through `error_contract_ref`; unknown members and owner-unmapped tokens are invalid. Findings MUST contain no secret or physical identity, reject duplicates, and sort by `path`, then `reason_code`, then canonical `details` bytes. `valid` requires `[]`; `invalid` requires `1..256` findings. Each canonical validation result MUST contain `1..1048576` bytes.

The write result MUST contain exactly:

```json
{"schema_id":"cartulary.extension_transaction_participant_write_result.v1","participant_id":"...","status":"written"}
```

The canonical write result MUST contain `1..1048576` bytes.

The final-commit protocol MUST execute exactly:

1. resolve the exact participant set from admitted contributions and Table 20-A;
2. sort participants by `participant_id` ascending UTF-8 bytes;
3. construct and schema-validate every participant input;
4. invoke every prepare algorithm side-effect free with `phase='prepare'`;
5. reject malformed, duplicate, undeclared-kind, or contradictory serialization keys;
6. sort the union of serialization keys by the comparator above;
7. begin one final database transaction;
8. acquire serialization locks in that exact order;
9. re-read every mutable precondition inside the transaction;
10. invoke every validation algorithm in participant order with `phase='validate'` and read-only access;
11. if every validator returns `valid`, invoke every write algorithm in the same order with `phase='write'` and read-write access;
12. require every write result to equal the exact `written` result;
13. write shared job proof, idempotency outcome, audit outbox, history, collaboration invalidation, staged-object publication references, and terminal result as applicable;
14. commit exactly once.

At every numbered boundary, immediately before and immediately after each participant invocation, and immediately before transaction begin, every lock acquisition, rollback request, and commit invocation, the coordinator MUST sample cancellation and the effective deadline under EXT-REQ-215. Participant inputs and results MUST be validated in participant order, and validation MUST stop at the first invalid participant or first aggregate-bound violation. A later participant MUST NOT be invoked to replace, aggregate, or reorder that first failure.

Prepare and validation algorithms MUST be side-effect free. Write algorithms MUST NOT begin until every validation passes. Contract major `1` performs no automatic transaction retry. A deadlock or serialization abort with a proven absent commit MUST return HTTP `409` where public, `error.code='transaction_conflict'`, `reason_code` equal to exactly one of `deadlock_detected` or `serialization_failure`, and `retryable=true`. A timeout with proven absent commit MUST return HTTP `503` where public, `error.code='service_unavailable'`, `reason_code='extension_transaction_timeout'`, and `retryable=true`. Any caller retry MUST use the owning idempotency contract; without a retry, the conflict or timeout is terminal for that invocation.

Cancellation or timeout before step 7 MUST cancel without opening a database transaction. During steps 7 through 13, it MUST request rollback and MUST report cancellation or timeout only after rollback or other absence of commit is proven. The final-commit boundary begins when step 14 invokes database commit. From that point, the operation MUST return only committed success, committed replay, proven absent-commit cancellation or conflict, or fatal indeterminate outcome. It MUST NOT report ordinary cancellation, timeout, or conflict while commit outcome is unknown.

Profiles: base
Verified by: EXT-AC-111, EXT-AC-130, EXT-AC-153

**EXT-REQ-192**
An operation that publishes authoritative non-database bytes MUST use one shared non-public `cartulary.extension_staged_object.v1` record containing exactly:

- `schema_id`, exactly `cartulary.extension_staged_object.v1`;
- `staging_id`, an ASCII identifier of `1..160` bytes matching `[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}` and unique within the deployment;
- `owner_profile_id`;
- `storage_identity`, an opaque non-secret UTF-8 string of `1..512` bytes under the shared object-storage owner;
- `expected_byte_size`, a JSON integer in `0..9223372036854775807` that also satisfies the owning operation's lower ceiling;
- `expected_sha256`, a SHA-256 digest string;
- `staged_at`, a Core canonical UTC timestamp;
- `staging_expires_at`, a Core canonical UTC timestamp;
- `ready_at`, a canonical timestamp or `null`;
- `published_at`, a canonical timestamp or `null`;
- `abandoned_at`, a canonical timestamp or `null`;
- `state`, exactly `allocated`, `ready`, `published`, or `abandoned`;
- `delete_state`, exactly `not_applicable`, `pending`, or `deleted`;
- `delete_attempt_count`, a JSON integer in `0..2147483647`;
- `next_delete_attempt_at`, a canonical timestamp or `null`;
- `last_delete_error_code`, a safe ASCII token of `1..128` bytes or `null`.

At allocation, the exact defaults are `ready_at=null`, `published_at=null`, `abandoned_at=null`, `state='allocated'`, `delete_state='not_applicable'`, `delete_attempt_count=0`, `next_delete_attempt_at=null`, and `last_delete_error_code=null`; each value MUST be written explicitly and MUST NOT be inferred from a missing column, zero value, or historical representation.

`staging_expires_at` MUST equal `staged_at + 24 hours`. Every transition timestamp MUST use the Core canonical UTC timestamp contract. At or after `staging_expires_at`, redemption of every unpublished object MUST fail before object-storage access, independently of janitor progress. Published objects follow their authoritative resource contract and MUST NOT be invalidated by staging expiry. Physical deletion has no completion service-level guarantee during deployment downtime or storage failure.

The allowed state transitions are `allocated -> ready -> published`, `allocated -> abandoned`, and `ready -> abandoned`. No reverse transition is valid. `ready_at` MUST be non-null only in `ready` or `published`. `published_at` MUST be non-null only in `published`. `abandoned_at` MUST be non-null only in `abandoned`. `delete_state='not_applicable'` is required in `allocated`, `ready`, and `published`; `delete_state` is `pending` or `deleted` only in `abandoned`. `next_delete_attempt_at` and `last_delete_error_code` MUST be null unless `delete_state='pending'`.

`stage_extension_bytes_v1` MUST:

1. allocate one immutable storage identity and `allocated` record before upload;
2. write the complete bytes to that identity;
3. validate exact byte size and every required digest;
4. prohibit overwrite with different bytes;
5. keep the bytes inaccessible through Cartulary because no authoritative resource reference exists;
6. transition to `ready` only after validation;
7. complete every final-commit participant validation;
8. transition to `published` in the same database final commit that publishes the resource reference, idempotency result, audit, history, job result, job proof, and invalidation;
9. treat the database commit instant as the exact public publication instant.

An upload error, cancellation, digest mismatch, byte-length mismatch, or indeterminate upload outcome before `ready` MUST immediately transition the staging record to `abandoned` with `delete_state='pending'` and `next_delete_attempt_at` equal to the current Core wall time before reporting the operation failure. If that transition cannot be made durable, the operation is an integrity failure and MUST NOT be retried as a normal upload.

The staged storage identity MUST be the final referenced identity. Public correctness MUST NOT depend on a post-commit move or rename. A final publication commit at or after `staging_expires_at` is forbidden. An existing storage identity MUST satisfy idempotent replay only when the owning idempotency contract proves the same normalized request and committed result and the stored bytes match the committed size and digest exactly. In every other identity-reuse case, the operation MUST fail integrity validation; different bytes MUST NOT overwrite.

The shared janitor, not an inactive profile worker, MUST execute `cleanup_extension_staged_objects_v1` at startup before Stage 6 and every `intervals.extensions.staged_object_sweep_seconds` while serving. At most one sweep may execute per deployment. A timer firing while a sweep is active MUST coalesce into exactly one follow-up sweep; multiple missed intervals MUST NOT queue multiple sweeps or execute concurrently. One sweep cycle MUST:

1. capture one Core canonical UTC `cycle_cutoff_at` before its first candidate read;
2. repeatedly select eligible rows using the exact predicate `(state IN ('allocated','ready') AND staging_expires_at <= cycle_cutoff_at) OR (state='abandoned' AND delete_state='pending' AND next_delete_attempt_at <= cycle_cutoff_at)`, ordered by the applicable eligibility timestamp (`staging_expires_at` for the first branch and `next_delete_attempt_at` for the second), then `staging_id`;
3. process at most `limits.extensions.staged_object_cleanup_batch` rows in one database transaction;
4. continue bounded batches until no eligible row at or before the cutoff remains; and
5. leave every later expiration for the next cycle.

For each candidate, the janitor MUST:

1. re-read authoritative references in a database transaction;
2. enter fatal integrity shutdown if a committed authoritative reference exists while the staging record is not `published`;
3. otherwise transition `allocated` or `ready` to `abandoned`, set `abandoned_at`, `delete_state='pending'`, and `next_delete_attempt_at` to the current Core wall time;
4. commit the inaccessible state before attempting physical deletion;
5. after the state transaction has committed and no database transaction or lock is held, attempt physical deletion and classify the outcome through Table 20-B;
6. durably record the exact Table 20-B result;
7. for every outcome that preserves `delete_state='pending'`, increment `delete_attempt_count` with saturation at `2147483647`, record one safe error code, and calculate `next_delete_attempt_at` only after the increment using:

```text
retry_delay_seconds = min(60 * 2^(delete_attempt_count - 1), 86400)
```

No random jitter is permitted in contract major `1`. After saturation, retries continue every `86400` seconds.

The retry eligibility check, reference check, durable abandonment, physical deletion, result classification, attempt increment, delay calculation, and durable retry write MUST occur in exactly the order above. Physical deletion while a database transaction or row lock is held is a contract violation. Failure of the cleanup dependency at startup blocks readiness; while serving it sets readiness to `degraded_dependency` without reopening access to expired bytes.

**Table 20-B. Staged-object physical deletion outcomes**

| Outcome | Required behavior |
| --- | --- |
| Success or object already absent | Set `delete_state='deleted'`, clear `next_delete_attempt_at` and `last_delete_error_code`, and durably record the outcome. |
| Retryable or unknown failure | Preserve `delete_state='pending'` and use the retry schedule above. |
| Authorization or storage-configuration failure | Preserve `delete_state='pending'`, use the same retry schedule, and expose Core readiness `degraded_dependency` with reason `extension_staged_object_cleanup_dependency_failed` until the dependency succeeds. |
| Digest, storage identity, publication-state, or authoritative-reference contradiction | Invoke fatal integrity shutdown under EXT-REQ-134. |

Logical inaccessibility is enforced at redemption and has no sweep lag. Startup cleanup MUST capture its cutoff set, continue bounded batches until no row in that set remains unprocessed, and complete within `timeouts.extensions.staged_object_cleanup_seconds`; otherwise startup fails and remains not-ready. A physical deletion that has not succeeded MUST remain pending only under the durable retry state above. Authorization or storage-configuration failure may leave physical bytes pending indefinitely, but it MUST NOT make them accessible through Cartulary.

A preview, download, owner route, object-store handle, portability payload, snapshot, or report MUST NOT expose `allocated`, `ready`, or `abandoned` bytes. The shared owner MAY retain a non-secret deleted staging record under its diagnostic-retention contract. Omission behavior: without such retention, the record is removed only after the deletion outcome is durable.

If the storage-write outcome is indeterminate, final database commit MUST NOT begin. If final database-commit outcome is indeterminate, the runtime MUST invoke fatal integrity shutdown. Post-commit finalization MAY be retried only when it is idempotent and cannot change public identity or committed bytes; omission behavior is no required post-commit finalization.

Profiles: base
Verified by: EXT-AC-089, EXT-AC-112, EXT-AC-131, EXT-AC-154

# 21. Unclaim, reclaim, retirement, and state migration

**EXT-REQ-120**
On restart with a previously claimed profile now inactive, the runtime MUST:

- retain the profile descriptor and reserved identities;
- omit profile workspaces, controls, and active capabilities;
- reject new family work through Core inactive dispatch;
- start no profile worker;
- perform no profile-local external dependency probe;
- perform no profile-local semantic migration;
- preserve authoritative extension state unchanged except for generic job reconciliation under §22;
- perform no down migration, automatic deletion, downgrade, detach, or Core reclassification;
- preserve Core records previously created through valid Core owner operations as ordinary Core records.

Profiles: base
Verified by: EXT-AC-024, EXT-AC-041, EXT-AC-042, EXT-AC-087

**EXT-REQ-121**
Unclaiming or retirement is an availability and conformance decision, not an uninstall, purge, deletion, or retention event. The current revision defines no public or operator extension-data purge.

Profiles: base
Verified by: EXT-AC-041, EXT-AC-050, EXT-AC-070

**EXT-REQ-122**
Reclaiming retained extension state MUST validate state presence, metadata, lineage, compatibility, immutable migration definitions, and apply only required forward migrations before route, workspace, action, capability, or worker availability. Reclaim MUST preserve stable resource IDs, owner provenance, committed idempotency semantics, and existing Core references. It MUST NOT duplicate imported resources, bindings, jobs, audit occurrences, migration ledger entries, or object references.

Profiles: base
Verified by: EXT-AC-043, EXT-AC-047, EXT-AC-048, EXT-AC-086

**EXT-REQ-123**
Retirement in contract major `1` is represented only by Core 00 changing `claimable` to `false`. A retired profile remains recognized, discoverable, and route-reserved; its state remains preserved. This revision defines no automatic read-only retirement route state.

Profiles: base
Verified by: EXT-AC-050, EXT-AC-087

**EXT-REQ-124**
A future read-only retirement mode requires a later contract major that defines readable routes, write prohibitions, authorization, export behavior, final supported state version, retention, migration path, client cleanup, and conformance. An implementation MUST NOT invent such a mode locally.

Profiles: base
Verified by: EXT-AC-050, EXT-AC-074

**EXT-REQ-188**
Every `extension_versioned` profile MUST publish one closed `cartulary.extension_state_presence_manifest.v1` object containing exactly:

- `schema_id`, exactly `cartulary.extension_state_presence_manifest.v1`;
- `state_presence_manifest_id`, exactly `<profile_id>.state_presence.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `empty_state_policy`, exactly the descriptor's `allowed` or `forbidden` value;
- `database_family_ids[]`;
- `object_reference_family_ids[]`;
- `presence_mode`, exactly `any_authoritative_member`.

The two arrays contain authoritative families only. Each array MUST contain `0..4096` family IDs. The combined unique count MUST be `1..4096`. Each family ID MUST use `<profile_id>.<local_key>`, be unique across both arrays, and sort by ascending UTF-8 bytes. A shared storage owner MUST bind logical family IDs to physical storage through implementation-support manifests. The profile MUST NOT provide an arbitrary executable state-presence callback.

State presence is true exactly when at least one authoritative member exists in any declared database or object-reference family. Generic extension metadata, state metadata, migration ledgers, jobs, caches, projections, temporary files, staged or orphan objects, audit rows, conformance artifacts, and every other non-family record MUST NOT appear in the manifest and MUST NOT make state presence true. A shared owner or profile MUST NOT reclassify metadata as authoritative state.

When no authoritative member exists, `state_present=false`. Pending or committed generic metadata with `state_present=false` is valid only when `empty_state_policy='allowed'`; with `forbidden`, initialization and final validation MUST produce at least one authoritative member before metadata may commit. `allowed` does not create a synthetic member and does not make state present. Network Flow Activity MUST declare `empty_state_policy='allowed'`.

The manifest MUST serialize under `extension_registry_canonical_json_v1`, contain `1..1048576` canonical bytes, and have digest `extension_state_presence_manifest_sha256_v1`. Its schema, canonical bytes, digest, and identity MUST appear in generated-schema accounting, registry integrity, implementation-binding parity, the contract-closure catalog, and conformance accounting.

Profiles: base
Verified by: EXT-AC-086, EXT-AC-132, EXT-AC-142

**EXT-REQ-234**
Every `extension_versioned` profile MUST publish exactly one closed `cartulary.extension_state_initialization_definition.v1` object containing exactly:

- `schema_id`, exactly `cartulary.extension_state_initialization_definition.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `target_state_version`;
- `initialization`.

`target_state_version` MUST equal the descriptor's `current_state_version`. `initialization` MUST be exactly one of these closed variants:

```json
{"kind":"empty"}
```

```json
{"kind":"algorithm","algorithm_id":"...","algorithm_ref":"...","algorithm_definition_sha256":"..."}
```

For `kind='algorithm'`, `algorithm_id` MUST satisfy the public schema-ID scalar, `algorithm_ref` MUST be an `owner_locator_v1` with `anchor-kind='algorithm'` and anchor ID equal to `algorithm_id`, and `algorithm_definition_sha256` MUST bind the exact canonical algorithm definition resolved by that locator. For `kind='empty'`, all three algorithm members are absent. The initialization definition MUST serialize under `extension_registry_canonical_json_v1`, contain `1..1048576` canonical bytes, and have digest `extension_state_initialization_definition_sha256_v1`. The descriptor, owner facts, implementation binding, registry integrity object, and conformance artifacts MUST agree on the exact definition identity and digest.

An algorithm initialization MUST receive one closed `cartulary.extension_state_initialization_context.v1` object containing exactly:

- `schema_id`, exactly `cartulary.extension_state_initialization_context.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `target_state_version`;
- `authoritative_state_family_ids[]`;
- `initialization_definition_sha256`;
- `deadline_monotonic_ns`;
- `scoped_write_capability_id`.

`authoritative_state_family_ids[]` MUST equal the state-presence manifest's combined family set and sort by ascending UTF-8 bytes. The digest MUST equal the verified definition digest. The deadline MUST use `timeouts.extensions.migration_step_seconds` and be a process-local integer in `1..9223372036854775807`. The scoped capability is transaction-local, permits writes only to the declared authoritative families and pending state metadata, and cannot commit or roll back. The context and capability MUST expose no secret, wall clock, incident-authored value, egress, filesystem, unrestricted SQL, Core-owned family, or other-profile family.

The algorithm MUST return exactly:

```json
{"schema_id":"cartulary.extension_state_initialization_result.v1","status":"ready_to_validate"}
```

Initialization runs inside the initial metadata transaction. `kind='empty'` is permitted only when `empty_state_policy='allowed'`; it performs no profile-code invocation and constructs empty authoritative state plus pending metadata. `kind='empty'` with `forbidden` is invalid at generation. `kind='algorithm'` invokes only the digest-bound packaged algorithm and, when the policy is `forbidden`, MUST leave at least one authoritative member before final validation. Both variants then invoke the final state validator exactly once against pending state and metadata and commit exactly once only when it returns `valid` and the policy/presence relation is satisfied. A malformed result, forbidden access, timeout, policy violation, or final validation failure rolls back the complete initialization transaction and blocks claim publication.

Profiles: base
Verified by: EXT-AC-132, EXT-AC-133, EXT-AC-142

**EXT-REQ-216**
Every `extension_versioned` profile MUST implement the shared migration interfaces:

- `cartulary.extension_migration_context.v1`;
- `cartulary.extension_migration_apply_result.v1`;
- `cartulary.extension_migration_validation_context.v1`;
- `cartulary.extension_migration_validation_result.v1`;
- `cartulary.extension_final_state_validation_context.v1`;
- `cartulary.extension_final_state_validation_result.v1`.

The migration context MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_migration_context.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `migration_id`;
- `from_state_version`;
- `to_state_version`;
- `migration_definition_sha256`;
- `state_family_ids[]`;
- `metadata_before_sha256`;
- `deadline_monotonic_ns`;
- `state_access_capability_id`.

`state_access_capability_id` is an internal transaction-scoped opaque capability. It MUST expose only the profile's declared authoritative state families. It MUST NOT expose unrestricted SQL, Core-owned state, another profile's state, deployment configuration outside `cartulary.extension_profile_configuration_view.v1`, unrelated secrets, arbitrary filesystem access, or network egress.

The apply result MUST equal exactly:

```json
{"schema_id":"cartulary.extension_migration_apply_result.v1","status":"ready_to_validate"}
```

`state_family_ids[]` in every migration or final-state context MUST equal the profile's authoritative state-family set, reject duplicates, and sort by ascending UTF-8 bytes. Every digest member MUST be a SHA-256 digest string over the named canonical object. Every capability ID MUST identify one process-local capability created for the current invocation and MUST NOT be reused after transaction end or lock release.

The migration-validation context MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_migration_validation_context.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `migration_id`;
- `from_state_version`;
- `to_state_version`;
- `migration_definition_sha256`;
- `state_family_ids[]`;
- `metadata_before_sha256`;
- `deadline_monotonic_ns`;
- `state_access_capability_id`, exactly the capability supplied to the apply algorithm for the same transaction.

The final-state context MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_final_state_validation_context.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `state_version`;
- `state_family_ids[]`;
- `state_metadata_sha256`;
- `state_presence_manifest_sha256`;
- `read_only_state_access_capability_id`;
- `deadline_monotonic_ns`.

The migration-validation result MUST be exactly one closed variant:

```json
{"schema_id":"cartulary.extension_migration_validation_result.v1","status":"valid","findings":[]}
```

or:

```json
{"schema_id":"cartulary.extension_migration_validation_result.v1","status":"invalid","findings":[...]}
```

The final-state validation result MUST use the same two variants with `schema_id='cartulary.extension_final_state_validation_result.v1'`. In either result, `findings[]` MUST contain `0..256` `cartulary.extension_startup_finding.v1` objects whose paths are relative to the validated profile state. `status='valid'` requires `findings=[]`; `status='invalid'` requires `1..256` findings. Findings MUST be secret-safe and ordered under EXT-REQ-162.

A wrong schema, wrong profile, wrong migration identity, wrong algorithm identity, thrown error, panic, timeout, malformed or unsorted finding, or result outside the closed variants is validation failure.

Profiles: base
Verified by: EXT-AC-113

**EXT-REQ-217**
One migration step MUST execute exactly:

1. open one scoped database transaction;
2. invoke the definition's apply algorithm through `cartulary.extension_migration_context.v1`;
3. validate the apply result;
4. invoke the definition's validation algorithm against pending transaction state;
5. if apply or validation fails, times out, or returns `invalid`, roll back, write no ledger row, and do not advance metadata;
6. if valid, write the migration ledger row and updated metadata inside the same transaction;
7. commit exactly once;
8. re-read committed ledger and metadata through shared runtime code and verify their exact digests and version transition.

Owner semantic validation MUST NOT first occur after commit. Any migration mutation MUST be limited to transactionally protected structured state. When immutable object bytes are required, it MUST use the staged-object contract and publish only references in a final database commit; it MUST NOT perform an irreversible external mutation that cannot be rolled back, abandoned, or reconciled.

After the complete migration sequence, including an already-current or zero-step path, the runtime MUST invoke the final state validator exactly once with read-only state access while holding the profile migration lock. It MUST NOT invoke the final validator after each migration step; each step uses only its migration-definition pending-state validator. Fresh initialization invokes the final validator exactly once against pending state and pending metadata inside the initialization transaction. Fresh metadata becomes authoritative only when that result is `valid`. A final validation failure after earlier committed migration steps leaves those steps committed, leaves the profile unclaimed, preserves the resumable state version, and emits `extension_state_validation_failed`.

Profiles: base
Verified by: EXT-AC-114

**Table 21-A. State metadata and authoritative-state consistency**

| State metadata | Authoritative owner state | Required result |
| --- | --- | --- |
| Absent | Absent | Execute the exact EXT-REQ-234 initialization variant. `kind='empty'` is legal only for `empty_state_policy='allowed'`; an algorithm initializer may satisfy `forbidden` only by creating authoritative state. Create pending state and metadata, invoke final validation once, and commit once only when valid; no migration ledger row is created. |
| Absent | Present | Fail `extension_state_metadata_missing`; do not infer or initialize a version. |
| Present | Absent | With `empty_state_policy='allowed'`, continue to lineage and stored-version validation, migration when required, and final validation. With `forbidden`, fail `extension_state_incomplete`; do not treat the profile as fresh. |
| Present | Present | Validate lineage and stored version; migrate when required. |

`cartulary.extension_state_metadata.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_state_metadata.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `state_version`;
- `last_migration_id`, string or `null`;
- `metadata_version`.

There MUST be exactly one current metadata object per deployment and profile. `metadata_version` MUST be a JSON integer in `1..2147483647`, MUST equal `1` on first metadata creation, and MUST increase by exactly one on every committed metadata change. `state_version` MUST be a JSON integer in `1..2147483647`. `last_migration_id` MUST be `null` on fresh metadata initialization and MUST equal the most recently committed ledger `migration_id` after any migration.

**EXT-REQ-189**
Every committed migration step MUST create one `cartulary.extension_migration_ledger_entry.v1` object containing exactly:

- `schema_id`, exactly `cartulary.extension_migration_ledger_entry.v1`;
- `profile_id`;
- `migration_lineage_id`;
- `migration_id`;
- `from_state_version`;
- `to_state_version`;
- `migration_definition_sha256`;
- `committed_at` under the Core canonical UTC timestamp contract;
- `resulting_state_version`.

`from_state_version` and `to_state_version` MUST be JSON integers in `1..2147483647`; `to_state_version` MUST equal `from_state_version + 1`; `resulting_state_version` MUST equal `to_state_version`; and the tuple `(profile_id, migration_lineage_id, from_state_version, to_state_version)` MUST be unique. `committed_at` MUST satisfy the Core canonical RFC3339 UTC timestamp contract. A committed `migration_id` MUST always have the same `migration_definition_sha256`. A changed digest MUST fail with `extension_migration_definition_changed`; the implementation MUST NOT reapply, replace, or reinterpret the committed step.

A deployment-global exclusive migration lock for the profile MUST be acquired before evaluating Table 21-A, initializing metadata, or applying a migration. Profiles MUST be processed in dependency order. No migration-lock mechanism is required by name. The selected mechanism MUST provide the same deployment-global exclusion and crash-release behavior. It MUST wait no longer than `timeouts.extensions.migration_lock_seconds`; timeout fails before mutation with `extension_migration_lock_timeout`. After acquiring the lock, the runtime MUST re-read state presence, metadata, ledger, and packaged migration definitions.

The lock MUST remain held through the Stage 3 final local state validation and MUST be released before Stage 4 dependency probes. The owner post-migration algorithm in Stage 4 runs after lock release. Each migration step MUST atomically commit only after EXT-REQ-217 pending-state validation passes:

- the step's authoritative state changes;
- one migration ledger entry;
- the resulting state metadata version and `last_migration_id`.

A step is limited by `timeouts.extensions.migration_step_seconds`; the complete profile sequence is limited by `timeouts.extensions.profile_migration_seconds`. A later step MUST NOT begin after the profile deadline. A timed-out active step with a determinate uncommitted outcome MUST roll back and fail with `extension_migration_timeout`. An indeterminate commit outcome MUST invoke the fatal integrity-shutdown contract. Migrations MUST make no third-party request.

Profiles: base
Verified by: EXT-AC-086

**EXT-REQ-125**
For `state_ownership.kind='extension_versioned'`, `migrate_extension_state_v1` MUST execute:

1. acquire the profile lock and re-read state under EXT-REQ-189;
2. evaluate Table 21-A;
3. if fresh initialization applies, execute the exact EXT-REQ-234 definition in one initialization transaction, construct pending state and metadata at `current_state_version`, run the final state validator exactly once against that pending state, and commit only when the result is `valid`;
4. require metadata `migration_lineage_id` to equal the descriptor lineage;
5. if the stored version equals `current_state_version`, run the final validator exactly once without migration;
6. if the stored version is greater than `current_state_version`, fail with `extension_state_version_unsupported` before mutation;
7. if the stored version is lower than `minimum_migratable_state_version`, fail with `extension_state_version_unsupported` before mutation;
8. for each integer version from the stored version through `current_state_version - 1`, require exactly one owner migration definition from N to N+1 and one matching implementation-binding declaration;
9. reject a missing, duplicate, changed-digest, nonconsecutive, or wrong-lineage step before applying the first migration;
10. apply each migration in ascending version order and validate pending state before atomically committing under EXT-REQ-189 and EXT-REQ-217;
11. verify each committed ledger and metadata transition after commit;
12. after every required step has committed, validate final state exactly once at `current_state_version` before profile availability.

Profiles: base
Verified by: EXT-AC-044, EXT-AC-045, EXT-AC-046, EXT-AC-047, EXT-AC-048, EXT-AC-086

**EXT-REQ-126**
A migration step MUST be deterministic and idempotent with respect to a crash before its atomic commit. A committed ledger entry MUST prevent reapplication. A missing consecutive step, validation failure, migration failure, timeout, or lock failure MUST block the profile claim with no route, listener, workspace, capability, or profile worker publication.

Profiles: base
Verified by: EXT-AC-046, EXT-AC-047, EXT-AC-048, EXT-AC-086

**EXT-REQ-127**
Production down migration is unavailable. An older implementation MUST NOT reinterpret newer extension state. Unsupported downgrade MUST fail claim admission before extension mutation.

Profiles: base
Verified by: EXT-AC-028, EXT-AC-049

**EXT-REQ-128**
A profile with `state_ownership.kind='core_managed'` MUST use the Core migration and restore contracts named by its owner declaration and MUST NOT run a second profile-local semantic migration line. A profile with `kind='none'` MUST create no authoritative extension state, state metadata, migration ledger, or state-presence manifest.

Profiles: base
Verified by: EXT-AC-061, EXT-AC-063

**EXT-REQ-129**
For the current adoption baseline, `network_flow_activity` MUST declare:

- `state_ownership.kind='extension_versioned'`;
- `current_state_version=1`;
- `minimum_migratable_state_version=1`;
- `migration_lineage_id='network_flow_activity.state_v1'`;
- one adopted `cartulary.extension_state_presence_manifest.v1` owner declaration;
- one `cartulary.extension_state_initialization_definition.v1` with `initialization={"kind":"empty"}`;
- one final state-validation algorithm and implementation binding.

Profiles: network_flow_activity
Verified by: EXT-AC-004, EXT-AC-044, EXT-AC-072, EXT-AC-086

**EXT-REQ-190**
Inactive profiles MUST follow Table 21-B. The table is exhaustive for shared inactive behavior.

Profiles: base
Verified by: EXT-AC-087

**Table 21-B. Inactive-profile behavior**

| Behavior | `unclaimed` | `recognized_unclaimable` |
| --- | --- | --- |
| `claimed=true` | Ordinary requested-profile admission. | Fail `extension_profile_not_claimable`. |
| Claim key omitted or false | Accepted. | Accepted. |
| `inactive_policy='forbidden'` key present | Fail `extension_config_without_claim`. | Fail `extension_config_without_claim` using the last-adopted contract. |
| `inactive_policy='syntax_only'` key present | Validate only JSON shape, scalar/reference grammar, bytes, and depth; keep inert. | Apply the same checks using the last-adopted contract; keep inert. |
| Unknown profile-local key | Fail. | Fail. |
| Syntax-only value handling | Do not semantically validate, resolve, log, emit, or transmit the value. | Same. |
| Secret resolution | None. Syntax-only checks validate reference grammar, bytes, and depth only. | None. |
| File or trust-material resolution | None. Syntax-only checks validate reference grammar, bytes, and depth only. | None. |
| Third-party egress or dependency probe | Forbidden. | Forbidden. |
| Profile-local semantic migration | Forbidden. | Forbidden. |
| Profile worker | Zero. | Zero. |
| New profile resource work | Rejected through reserved-unclaimed dispatch. | Rejected through reserved inactive dispatch. |
| Declarative shared state-presence check | Permitted. | Permitted. |
| Generic nonterminal-job reconciliation | Required. | Required. |
| Authoritative state | Preserve unchanged. | Preserve unchanged. |
| Reserved route and workspace identity | Preserve in registry and discovery. | Preserve in registry and discovery. |
| Later activation | Ordinary reclaim admission. | Ordinary reclaim admission only after Core 00 makes the profile claimable again. |

Local availability, secret-existence, file-existence, and trust-material checks begin only for a claimed profile in Stage 2. No inactive-policy exception exists in contract major `1`.

The set of last-adopted known retired keys MUST come from the retained digest-bound owner fragment, not from implementation code or local configuration history.

# 22. Jobs and runtime failure isolation

**EXT-REQ-218**
Every durable extension job kind MUST be declared by one closed canonical `cartulary.extension_job_kind_contract.v1` containing exactly:

- `schema_id`, exactly `cartulary.extension_job_kind_contract.v1`;
- `profile_id`;
- `job_kind`;
- `operation_kind`;
- `proof_policy`;
- `idempotency_policy`;
- `idempotency_identity_schema_id`;
- `terminal_result_schema_id`;
- `resource_ref_contracts[]`;
- `cancellation_policy`;
- `max_proof_bytes`.

`operation_kind` MUST equal `<profile_id>.<local_key>` under the Table 4-B local-key grammar. `proof_policy` MUST equal `required_on_terminal_success` or `forbidden`. `idempotency_policy` MUST equal `required` or `none`. `cancellation_policy` MUST equal `precommit_observable` or `not_cancelable`. `terminal_result_schema_id` MUST be a non-null public schema ID that resolves through the current owner manifest and implementation binding. `idempotency_policy='required'` requires a non-null public identity schema ID that resolves through the same boundaries; `none` requires `idempotency_identity_schema_id=null`. A job that can publish any extension-owned or cross-owner resource MUST use `required_on_terminal_success`. A proof-required terminal success MUST commit one proof even when `resource_refs=[]`. A proof-forbidden job MUST never create a proof.

`resource_ref_contracts[]` MUST contain `0..64` rows. Each row MUST contain exactly `resource_ref_kind`, `resource_id_schema_id`, and `max_refs`. `resource_ref_kind` MUST satisfy Table 4-B. `resource_id_schema_id` MUST be a public schema ID that resolves through the current owner manifest and implementation binding. Rows MUST reject duplicate kinds and sort by ascending UTF-8 bytes of `resource_ref_kind`. `max_refs` MUST be a JSON integer in `1..1024`; a kind that permits no references MUST be omitted rather than declared with zero. The sum across rows MUST NOT exceed `1024`. `max_proof_bytes` MUST be a JSON integer in `1..1048576`.

A proof-required job MUST have at most one immutable proof, and its committed terminal success MUST have exactly one immutable proof. Proof replacement and proof deletion are forbidden while the job or its idempotency outcome is retained. The proof's total nesting depth MUST NOT exceed `32`; resource references MUST sort by `resource_ref_kind`, then canonical resource-ID bytes. The job-kind contract MUST serialize under `extension_registry_canonical_json_v1`, contain `1..1048576` canonical bytes, and have digest `extension_job_kind_contract_sha256_v1`. Reconciliation MUST determine proof requiredness only from this digest-bound contract and MUST reject a proof present under `forbidden` or absent under a committed required success.

Profiles: base
Verified by: EXT-AC-115

**EXT-REQ-130**
Every durable job owned by or producing an extension resource MUST carry internal `owner_profile_id` and `job_kind`. `owner_profile_id` MUST equal the exact profile ID, and `job_kind` MUST resolve to one `cartulary.extension_job_kind_contract.v1` whose digest is present in the implementation binding. These members are internal job ownership metadata and need not be public unless Core 01 adopts them in a public schema.

Profiles: base
Verified by: EXT-AC-051, EXT-AC-080

**EXT-REQ-191**
A durable job whose final commit can publish an extension or cross-owner resource MUST commit one `cartulary.extension_job_commit_proof.v1` in the same final database transaction as the authoritative effect and canonical terminal success.

The proof MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_job_commit_proof.v1`;
- `job_id`;
- `owner_profile_id`;
- `operation_kind`;
- `final_commit_id`;
- `idempotency_identity`;
- `normalized_request_sha256`;
- `terminal_result`;
- `terminal_result_sha256`;
- `resource_refs[]`;
- `audit_correlation_id`;
- `committed_at` under the Core canonical UTC timestamp contract.

`idempotency_identity` MUST conform to the exact schema selected by the job-kind contract or be `null` exactly when `idempotency_policy='none'`. `job_id`, `audit_correlation_id`, and `final_commit_id` MUST satisfy EXT-REQ-233. `audit_correlation_id` is `null` exactly when the owner declares no audit occurrence. `terminal_result` MUST conform to the job-kind contract `terminal_result_schema_id` and be the exact canonical Core common-job terminal success object originally committed. `terminal_result_sha256` MUST be the digest of its canonical bytes. `resource_refs[]` MUST use only declared resource-reference contracts, respect every per-kind and aggregate bound, reject duplicates, and sort by kind then canonical resource ID bytes. The complete proof MUST NOT exceed `max_proof_bytes` or nesting depth `32`. Committed replay MUST preserve both original correlation identifiers.

A durable precommit cancellation observation MUST conform to `cartulary.extension_job_cancellation_observation.v1` and contain exactly:

- `schema_id`, exactly `cartulary.extension_job_cancellation_observation.v1`;
- `cancellation_request_id`;
- `job_id`;
- `observed_at`;
- `observed_before_final_commit`, exactly `true`.

`cancellation_request_id` and `job_id` MUST satisfy the exact Core owner scalars imported by EXT-REQ-233. `observed_at` MUST satisfy the Core canonical UTC timestamp contract.

Profiles: base
Verified by: EXT-AC-088

**EXT-REQ-131**
`reconcile_inactive_extension_jobs_v1` MUST process each inactive profile as follows:

1. select nonterminal jobs ordered by `submitted_at`, then `job_id`, ascending;
2. fetch at most configured limit plus one and fail before mutation with `extension_reconciliation_limit_exceeded` if the extra row exists;
3. classify every selected job without changing job state;
4. load the digest-bound job-kind contract and validate proof policy, proof schema, proof ownership, request digest, terminal-result schema and digest, resource-reference contracts and bounds, audit correlation, and cancellation observation;
5. fail the complete profile reconciliation before any terminal mutation when any proof is missing where required, malformed, contradictory, or digest-inconsistent;
6. after complete classification succeeds, apply every terminal update for the selected profile job set in one atomic database transaction and in the same deterministic order;
7. commit no terminal update when classification fails, the reconciliation timeout expires before commit, or any selected update fails;
8. treat an indeterminate reconciliation transaction commit outcome as the fatal `indeterminate_database_commit` condition under EXT-REQ-134.

Outcome precedence MUST use Table 22-A.

**Table 22-A. Inactive-job reconciliation outcome precedence**

| Evidence | Required terminal result |
| --- | --- |
| Valid final-commit proof | Publish the exact stored original terminal success. |
| No final-commit proof and valid precommit cancellation observation | Set `status='canceled'` under the Core common-job contract. |
| Neither proof | Set `status='failed'`, `error_summary.code='extension_profile_unclaimed'`, and `retryable=false`. |
| Invalid, missing-required, or contradictory proof | Fail startup with `extension_unclaim_reconciliation_failed`; do not guess a terminal state. |

A valid final-commit proof takes precedence over cancellation because the authoritative effect committed. The original success MUST be replayed from the stored canonical terminal result; it MUST NOT be reconstructed from current resource state.

The algorithm MUST NOT start an extension worker, create, delete, migrate, or reinterpret extension domain resources, or perform third-party egress.

Profiles: base
Verified by: EXT-AC-052, EXT-AC-053, EXT-AC-088

**EXT-REQ-132**
If the runtime cannot determine a safe reconciliation outcome, startup MUST fail with `extension_unclaim_reconciliation_failed`. It MUST NOT leave the job indefinitely nonterminal after reporting readiness, guess success, select cancellation over a valid commit proof, or publish a newly reconstructed result.

Profiles: base
Verified by: EXT-AC-052, EXT-AC-053, EXT-AC-088

**EXT-REQ-133**
A runtime request, job, renderer, parser, or external dependency failure inside one claimed extension MUST fail the addressed operation under the named owner contract. It MUST NOT automatically remove the profile from the resolved claim set, change discovery to inactive, silently invoke another profile, mutate Base authorization, or initiate process shutdown unless EXT-REQ-134 identifies a fatal integrity condition.

Profiles: base
Verified by: EXT-AC-054

**EXT-REQ-134**
The extension framework MUST initiate fatal integrity shutdown only after detecting one of these conditions:

- an indeterminate final atomic database transaction outcome;
- an in-memory canonical-registry, registry-integrity, or resolved-claim-set digest mismatch;
- a committed extension migration-ledger version or digest that does not match committed extension state metadata;
- confirmed loss of the deployment-global application-process lease while the process is starting or serving;
- unexpected termination of a mandatory publication listener, job-dequeue gate, or worker after it was prepared or published, classified as `published_component_lost`;
- a staged-object digest, storage-identity, publication-state, or authoritative-reference contradiction, including a committed authoritative object reference whose staging record is not `published`, or a `published` staging record whose committed object reference is absent or digest-inconsistent.

An individual request, job, poll, or listener-handled operation failure is not component termination and MUST remain isolated under EXT-REQ-133. Component loss before listener bind, between bind and serving, or while serving is fatal; the process MUST NOT restart or replace the component in-process. Every other runtime extension failure MUST remain isolated under EXT-REQ-133. A profile owner MUST NOT add a fatal condition through a profile-local error mapping. Adding another fatal condition requires a revision of this requirement and Table 26-A1.

Profiles: base
Verified by: EXT-AC-054, EXT-AC-090, EXT-AC-104, EXT-AC-112

**EXT-REQ-193**
Core 04 is the sole lifecycle and process-exit owner. Before coordinated adoption, Core 04 MUST publish `fatal_integrity_shutdown_v1`; this subsystem MUST import that exact contract and MUST NOT define an independent lifecycle. The imported contract MUST admit exactly these transitions:

```text
starting -> terminating -> exited
running  -> quiescing -> terminating -> exited
```

For fatal detection while `starting`, the imported contract MUST close the admission gate, cancel or roll back every precommit operation, close any prepared or bound listener without admitting a public socket, preserve committed state and durable queued jobs, emit exactly one safe fatal diagnostic, and exit with code `70`. It MUST emit no WebSocket terminal event when no public WebSocket was admitted.

For fatal detection while `running`, the imported contract MUST:

1. transition atomically from `running` to `quiescing`; a repeated fatal signal MUST NOT restart the drain deadline;
2. set `/readyz` to HTTP `503` with safe status token `fatal_integrity_failure`;
3. keep `/healthz` at HTTP `200` only while the process remains alive;
4. reject newly admitted HTTP work with HTTP `503`, `error.code='service_unavailable'`, `reason_code='extension_integrity_failure'`, and `retryable=true`;
5. reject new WebSocket upgrades and stop common-job dequeue;
6. send every existing WebSocket the Core terminal error event with `code='server_shutdown'` and `reason_code='extension_integrity_failure'`, then close it with WebSocket code `1011`;
7. cancel work that has not crossed its final-commit boundary;
8. allow work whose final commit is already in progress or committed to complete only the publication required to make that committed outcome recoverable;
9. preserve queued durable jobs and committed state;
10. wait no longer than `timeouts.extensions.shutdown_drain_seconds` from entry to `quiescing`;
11. transition to `terminating`, force-close remaining connections and workers, and emit one safe fatal diagnostic;
12. exit with process exit code `70`.

A repeated signal for the same or another fatal condition MUST be idempotent: it MUST NOT reopen admission, duplicate the safe diagnostic, duplicate a WebSocket terminal event, restart a timeout, or change the selected exit code. A startup configuration or ordinary admission failure MUST start no public listener and exit with code `2`; it MUST NOT use fatal shutdown. Core 04 MUST reserve these exact exit codes and readiness/health behaviors before coordinated adoption. A fatal diagnostic MUST NOT contain a secret, incident value, storage identifier, SQL, or raw cryptographic detail.

For `published_component_lost`, the same lifecycle MUST close readiness and admission immediately, drain under the existing deadline, preserve every committed state and queued durable job, and exit `70`. The process MUST NOT return to `running`, rebuild the publication plan, or restart the listener, dequeue gate, or worker in-process; recovery belongs to the external supervisor starting a new process.

Profiles: base
Verified by: EXT-AC-090, EXT-AC-137, EXT-AC-158

**EXT-REQ-135**
Profile workers MUST start only for claimed profiles and only after complete Stage 6 publication. Inactive profiles MUST have zero active profile workers.

Profiles: base
Verified by: EXT-AC-042, EXT-AC-087

# 23. Backup, restore, portability, and reporting participation

**EXT-REQ-136**
Physical backup MUST preserve all authoritative extension state, state-presence logical-family bindings, extension state metadata, migration ledger entries, required idempotency references, job commit proofs, and authoritative object-store content regardless of whether the profile is claimed at backup time.

Profiles: base
Verified by: EXT-AC-061

**EXT-REQ-137**
Every profile with durable state MUST register authoritative database object families and authoritative object-storage reference families in the repository schema-ownership and backup manifests. It MUST separately declare derived caches, projections, staged objects, and temporary files as discardable or rebuildable. Physical table, bucket, and object-key layout remains implementation-support data.

Profiles: base
Verified by: EXT-AC-061, EXT-AC-063, EXT-AC-073

**EXT-REQ-223**
Every profile with `state_ownership.kind='extension_versioned'` MUST have one canonical `cartulary.extension_physical_state_binding.v1` build-time object containing exactly:

- `schema_id`, exactly `cartulary.extension_physical_state_binding.v1`;
- `profile_id`;
- `state_presence_manifest_sha256`;
- `bindings[]`.

Each `bindings[]` row MUST contain exactly:

- `binding_id`;
- `logical_family_id`;
- `storage_kind`;
- `physical_ref`;
- `state_class`;
- `backup_inclusion`;
- `restore_order_group`;
- `backup_codec_id`;
- `backup_codec_sha256`;
- `post_restore_validation_algorithm_id`;
- `rebuild_algorithm_id`.

`state_presence_manifest_sha256` MUST equal the current canonical profile manifest's `extension_state_presence_manifest_sha256_v1` digest.

`binding_id` MUST equal `<profile_id>.<local_key>` under Table 4-B. `storage_kind` MUST equal `postgres`, `object_store`, or `filesystem`. `state_class` MUST equal `authoritative` or `derived`. An authoritative row's `logical_family_id` MUST resolve exactly once in the digest-bound state-presence manifest. A derived row's `logical_family_id` MUST NOT resolve in that manifest; the binding row is its sole family declaration. Logical family IDs MUST be unique across authoritative and derived rows. `backup_inclusion` MUST equal `required` or `excluded_rebuildable`. An authoritative binding MUST use `backup_inclusion='required'`. A derived binding MUST use `backup_inclusion='excluded_rebuildable'` and declare a rebuild algorithm. In contract major `1`, every `filesystem` binding MUST be derived and `excluded_rebuildable`; authoritative filesystem binding is invalid. `restore_order_group` MUST be a JSON integer in `0..1024`.

`physical_ref` MUST be an opaque, non-secret implementation-local storage selector containing `1..512` UTF-8 bytes. It MUST contain no credential, endpoint host, bucket name, object key, database name, absolute filesystem path, query, fragment, C0 or C1 control, or incident value. It MUST NOT appear in public responses, ordinary diagnostics, telemetry, portability payloads, profile descriptors, or conformance findings. Its resolution to actual implementation storage is build-time implementation data and MUST preserve the exact one-to-one binding checks in this requirement.

`backup_codec_id` MUST satisfy the public schema-ID scalar and identify the exact EXT-REQ-235 codec for the row. `backup_codec_sha256` MUST equal that codec's `extension_backup_binding_codec_sha256_v1` digest. `post_restore_validation_algorithm_id` MUST equal the exact shared storage-owner algorithm in Table 23-E for the row's `storage_kind`. It MUST resolve through the Core backup/restore owner manifest and packaged Core implementation binding, not through profile executable code. `rebuild_algorithm_id` MUST be non-null exactly when `state_class='derived'` and `backup_inclusion='excluded_rebuildable'`; otherwise it MUST be `null`. A rebuild algorithm MUST resolve through the profile owner manifest and implementation binding. Any rebuild execution MUST occur only after the profile is claimed and its compatibility and final-state admission have succeeded.

`bindings[]` MUST contain `1..4096` rows, reject duplicate `binding_id`, `logical_family_id`, and `physical_ref` values, and sort by `restore_order_group`, then `logical_family_id`, then `binding_id`, using ascending numeric or UTF-8 order as applicable. Every authoritative logical family in the state-presence manifest MUST have exactly one required backup binding. Every physical authoritative store owned by the implementation MUST appear exactly once. Derived state never affects state presence and MUST NOT be a backup source for authoritative state.

The object MUST serialize under `extension_registry_canonical_json_v1`, contain `1..1048576` canonical bytes, and have digest `extension_physical_state_binding_sha256_v1`.

Backup, restore, and conformance MUST use this digest-bound mapping. The implementation binding MUST repeat and exactly match every initialization algorithm and definition digest, backup codec ID and digest, rebuild algorithm, transaction-participant limit, and supporting schema identity required by the profile. Physical names remain implementation-specific and do not become public interoperability contracts. During inactive-profile restore, the shared storage-owner validator MUST verify restored count, byte length, codec, and digest parity against the backup manifest for the binding without parsing profile semantics or executing profile code. Excluded rebuildable derived state remains absent until a later successful claim invokes its rebuild algorithm. Missing authoritative binding, extra undeclared authoritative storage, inactive-profile exclusion, unvalidated restore, or unrebuildable excluded derived state is a conformance failure.

Profiles: base
Verified by: EXT-AC-116, EXT-AC-132, EXT-AC-135

**EXT-REQ-235**
Every physical-state binding MUST resolve to one canonical `cartulary.extension_backup_binding_codec.v1` object containing exactly:

- `schema_id`, exactly `cartulary.extension_backup_binding_codec.v1`;
- `backup_codec_id`;
- `binding_id`;
- `storage_kind`;
- `codec_contract_ref`;
- `logical_identity_algorithm_id`;
- `content_encoding_algorithm_id`;
- `max_items`;
- `max_entry_bytes`;
- `max_binding_bytes`;
- `historical_restore_codecs[]`.

The codec ID and binding ID MUST equal the physical-binding row. `storage_kind` MUST equal that row's storage kind. The three algorithm or contract identifiers MUST resolve to exact packaged implementation-binding definitions. `max_items` MUST be a JSON integer in `1..1048576`; `max_entry_bytes` and `max_binding_bytes` MUST be JSON integers in `1..9223372036854775807`, with `max_entry_bytes <= max_binding_bytes`. An empty binding remains admitted because its actual item and content-byte counts are zero. `historical_restore_codecs[]` MUST contain `0..16` closed `{backup_codec_id, backup_codec_sha256}` rows, reject duplicate IDs and digests, exclude the current codec, and sort by `backup_codec_id`, then digest. Each listed historical codec MUST also be packaged, resolve under the same binding, and be digest-bound by registry integrity. Omission materializes `historical_restore_codecs=[]`; explicit `null` is invalid. The codec object MUST serialize under `extension_registry_canonical_json_v1`, contain `1..1048576` bytes, and have digest `extension_backup_binding_codec_sha256_v1`.

For any byte string `x`, define:

```text
frame(x) = uint64_big_endian_length(x) || x

digest_input =
  "cartulary.extension.binding.v1\0"
  + frame(binding_id_utf8)
  + frame(storage_kind_utf8)
  + frame(decimal_item_count_ascii)
  + each ordered entry

entry =
  frame(logical_identity_bytes)
  + frame(raw_32_byte_content_sha256)
  + frame(decimal_byte_length_ascii)

binding_digest = SHA256(digest_input)
```

The length prefix is the unsigned 64-bit length of `x` in exactly eight big-endian bytes. Decimal counts and byte lengths use base-10 ASCII with no sign and no leading zero except `0`. Entry identities MUST contain `1..4096` bytes. The item count MUST be `0..max_items`; each entry byte length MUST be `0..max_entry_bytes`; and the sum of entry content byte lengths MUST be `0..max_binding_bytes`. The first excess MUST fail before partial backup publication; truncation is forbidden. Entries sort by unsigned lexicographic `logical_identity_bytes`; duplicate identities are invalid. An empty binding has item count `0`, no entries, and the digest of the remaining framed input. `binding_digest` is stored as exactly 64 lowercase hexadecimal characters encoding the raw SHA-256 result. The backup manifest MUST record the exact binding ID, storage kind, item count, binding digest, codec ID, and codec digest.

A PostgreSQL codec MUST map each row to canonical JSON under `extension_registry_canonical_json_v1` using stable logical field IDs declared by `codec_contract_ref`, use those canonical bytes as the content bytes, and hash those exact bytes; physical table or column names and database-native row serialization are forbidden digest inputs. An object-store codec MUST hash the exact immutable object bytes and use the codec's declared logical object identity. A filesystem codec MUST use a normalized relative POSIX path as logical identity and hash exact regular-file bytes; a leading slash, backslash, NUL, empty segment, `.`, `..`, symlink, or non-regular file is invalid.

Restore MUST select the codec by the backup metadata's exact ID and digest before any restore mutation. An unrecognized or unpackaged pair MUST fail with `backup_binding_codec_unsupported`. A recognized pair whose packaged bytes do not match its digest is `extension_registry_invalid`. Historical codec support exists only through an explicit packaged and digest-bound `historical_restore_codecs[]` row; inference, best-effort decoding, and use of the current codec for historical bytes are forbidden.

Restore contract v1 admits only a stopped deployment whose target authoritative state, extension metadata, migration ledger, authoritative object references, and nonterminal extension jobs are all empty. A running, serving, quiescing, or nonempty target MUST be rejected before any restore mutation; in-place or merge restore is unsupported. Restore MUST sort binding groups by ascending numeric `restore_order_group`, then process the bindings within one group sequentially in the physical-binding order. It MUST decode, enforce bounds, write, and run the shared post-restore validator for one binding before advancing to the next. Parallel binding mutation is forbidden.

A failed or interrupted target MUST remain non-serving and MUST NOT be admitted by startup until the restore owner proves a complete valid restore or the target is discarded through an owner-defined recovery procedure. Inactive profile code MUST never be invoked during restore. Historical decoding is admitted only by the exact packaged digest-bound codec row above. Derived state remains absent during restore and may be rebuilt only after a later successful claim has completed compatibility, migration, final validation, and publication admission.

Physical backup is implementation-binding-specific. Incident Portability remains the logical cross-implementation interchange boundary and MUST NOT expose physical codec frames or storage identities.

Profiles: base
Verified by: EXT-AC-135, EXT-AC-155

**Table 23-E. Shared post-restore binding validators**

| `storage_kind` | Required `post_restore_validation_algorithm_id` | Exact validation scope |
| --- | --- | --- |
| `postgres` | `extension_validate_postgres_binding_restore_v1` | Apply the exact EXT-REQ-235 codec, then compare binding identity, codec ID/digest, row count, entry identities, byte lengths, content hashes, and binding digest with the backup manifest. |
| `object_store` | `extension_validate_object_store_binding_restore_v1` | Apply the exact EXT-REQ-235 codec, then compare binding identity, codec ID/digest, object count, entry identities, exact immutable-byte hashes, byte lengths, and binding digest. |
| `filesystem` | `extension_validate_filesystem_binding_restore_v1` | Apply the exact EXT-REQ-235 codec, reject forbidden path or file forms, then compare binding identity, codec ID/digest, ordered relative paths, file hashes, byte lengths, and binding digest. |

Every validator result MUST be `valid` or `invalid` under `cartulary.extension_backup_restore_participant_result.v1`. A storage read failure, missing item, extra item, digest mismatch, byte-length mismatch, symlink, non-regular file, or manifest mismatch is `invalid`; it MUST NOT be repaired, ignored, or converted into a profile-semantic validation result.

**EXT-REQ-138**
Restore MUST preserve inactive extension state without invoking profile semantic code, semantically migrating it, or exposing it. A profile claimed after restore MUST complete physical binding and codec validation, state-presence and metadata validation, compatibility and any migration, and then final state validation exactly once before availability. A migrated restored profile uses each migration step's pending-state validator but no intermediate final validator. Derived state, when rebuilt after successful claim admission, MUST use only declared authoritative inputs.

Profiles: base
Verified by: EXT-AC-062, EXT-AC-063

**EXT-REQ-139**
Every incident-scoped profile MUST declare exactly one `incident_portability_mode`. Silent omission of authoritative extension state from an incident bundle is forbidden unless the selected mode explicitly declares that no authoritative incident state exists.

Profiles: base
Verified by: EXT-AC-034, EXT-AC-066, EXT-AC-073, EXT-AC-094

**EXT-REQ-232**
Every `incident_portability_participant`, `snapshot_reporting_participant`, and `backup_restore_participant` contribution MUST resolve to one canonical `cartulary.extension_participant_specialization.v1` containing exactly:

- `schema_id`, exactly `cartulary.extension_participant_specialization.v1`;
- `profile_id`;
- `participant_id`;
- `participant_kind`;
- `shared_context_schema_id`;
- `operations[]`.

`participant_kind` MUST equal `incident_portability`, `snapshot_reporting`, or `backup_restore`. `shared_context_schema_id` MUST equal the exact context in Table 23-D. Each `operations[]` row MUST contain exactly:

- `operation_kind`;
- `result_schema_id`;
- `algorithm_id`;
- `output_schema_id`;
- `ordering_algorithm_id`;
- `authorization_contract_ref`;
- `redaction_contract_ref`;
- `error_contract_ref`;
- `state_family_ids[]`;
- `max_input_bytes`;
- `max_output_bytes`;
- `max_items`.

`operations[]` MUST reject duplicate `operation_kind` values and sort by ascending UTF-8 bytes of `operation_kind`; array position has no other semantic meaning, and the specialization digest uses that order. The operation set and each `result_schema_id` MUST equal Table 23-D for the participant kind and operation. `shared_context_schema_id` is the complete serialized input schema for every operation of that specialization; an operation MUST NOT extend the shared context with profile-local members. Algorithm, result, and output schema IDs MUST resolve through current owner manifests and implementation bindings. `authorization_contract_ref`, `error_contract_ref`, and every non-null `redaction_contract_ref` MUST resolve to exact adopted owner requirements. `redaction_contract_ref` MUST be non-null only for `snapshot_reporting.emit`; it MUST be `null` for every other current operation.

`state_family_ids[]` MUST contain `0..64` unique logical state-family IDs sorted by UTF-8 bytes. Portability, Snapshot/Reporting, `backup_enumerate`, and `restore_validate` IDs MUST resolve in the authoritative state-presence manifest. `restore_rebuild` IDs MUST resolve only to derived `excluded_rebuildable` physical-binding rows and MUST NOT resolve in the state-presence manifest. `max_input_bytes` and `max_output_bytes` MUST be JSON integers in `1..67108864` and `0..67108864`, respectively; `max_items` MUST be a JSON integer in `0..1048576`. `max_input_bytes` limits the aggregate canonical bytes of the shared context plus every state, snapshot, portability-payload, or physical-binding item made available through the scoped input accessor. `max_items` independently limits both the selected input item count and emitted output item count. `max_output_bytes=0` means the operation permits only the generic empty or omit result and no referenced payload or output bytes. `max_items=0` means the scoped input accessor exposes no selected item and the result emits no item. `ordering_algorithm_id` is required even when the output can contain at most one item; in that case it MUST resolve to an exact identity-order algorithm.

The specialization MUST define all profile-specific input selection, payload/output schema, ordering, authorization, redaction, error mapping, state-family access, and limits. It MUST NOT contain physical table, bucket, object key, package, component, callback, secret, incident value, or deployment endpoint data. It MUST serialize under `extension_registry_canonical_json_v1`, contain `1..1048576` canonical bytes, and have digest `extension_participant_contract_sha256_v1` equal to the contribution's `participant_contract_sha256`.

Profiles: base
Verified by: EXT-AC-117, EXT-AC-134, EXT-AC-144

**Table 23-D. Participant specialization operation sets**

| `participant_kind` | `shared_context_schema_id` | Exact operation-to-result mapping |
| --- | --- | --- |
| `incident_portability` | `cartulary.extension_portability_participant_context.v1` | `export` -> `cartulary.extension_portability_export_result.v1`; `import` -> `cartulary.extension_portability_import_preparation_result.v1` |
| `snapshot_reporting` | `cartulary.extension_snapshot_reporting_participant_context.v1` | `emit` -> `cartulary.extension_snapshot_reporting_participant_result.v1` |
| `backup_restore` | `cartulary.extension_backup_restore_participant_context.v1` | `backup_enumerate`, `restore_validate`, and `restore_rebuild` -> `cartulary.extension_backup_restore_participant_result.v1` |

**Table 23-F. Shared-owner and profile-code invocation matrix**

| Operation | Authoritative-state and claim condition | Required invocation behavior |
| --- | --- | --- |
| Portability export | No authoritative state; any claim state | Shared owner returns `omit`; invoke no profile code. |
| Portability export | Authoritative state present; claimed and participant mode | Invoke the typed participant. |
| Snapshot/Reporting emit | No authoritative state; any claim state | Shared owner returns `empty`; invoke no profile code. |
| Snapshot/Reporting emit | Authoritative state present; claimed participant | Invoke the typed participant. |
| Backup enumerate | Any state or claim condition | Shared owner enumerates the physical-state binding; invoke no profile code. |
| Restore validate | Any state or claim condition | Shared storage validator executes EXT-REQ-235; invoke no profile code. |
| Restore rebuild | Claimed profile with an excluded derived binding requiring rebuild | Invoke only that binding's declared rebuild algorithm. |
| Restore rebuild | Every other condition | Perform no rebuild and invoke no profile code. |
| Portability import with payload | Claimed and compatible | Invoke the typed participant irrespective of prior target state. |
| Portability import with payload | Any other condition | Fail under Table 23-B before profile-code invocation. |

**EXT-REQ-221**
The shared portability, Snapshot/Reporting, and backup/restore participant boundaries are closed logical interfaces. Each participant contribution MUST bind one specialization through `participant_contract_ref` and `participant_contract_sha256`. Profile-code invocation MUST follow Table 23-F exactly and may occur only for `claim_state='claimed'`. Shared-owner enumeration, omission, empty-result, and storage validation are not participant-algorithm invocations and MUST run in the Table 23-F rows that require them.

The shared `claim_state` vocabulary is exactly `claimed`, `unclaimed`, and `recognized_unclaimable`. `state_present` is Boolean. For an `extension_versioned` profile with present authoritative state, `state_version` MUST be the exact stored positive state version; otherwise it MUST be `null`. Every `authorization_view_sha256`, `redaction_profile_sha256`, payload digest, output digest, and physical-binding digest MUST be a SHA-256 digest string. Every `timeout_seconds` MUST equal the effective Table 9-C value for the participant kind.

`cartulary.extension_portability_participant_context.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_portability_participant_context.v1`;
- `operation`, exactly `export` or `import`;
- `profile_id`;
- `contract_major`;
- `claim_state`;
- `state_present`;
- `state_version`, an integer or `null`;
- `incident_ref`;
- `authorization_view_sha256`;
- `payload_ref`, an `extension_safe_logical_ref_v1` value or `null`;
- `payload_sha256`, a digest or `null`;
- `timeout_seconds`.

`incident_ref` MUST satisfy the Core incident-reference scalar contract. For export, `payload_ref` and `payload_sha256` MUST both be `null`. For import, both MUST be non-null, `payload_ref` MUST satisfy `extension_safe_logical_ref_v1`, and both MUST bind exact inert payload bytes admitted by the bundle owner. The shared context is the complete serialized invocation input and MUST conform to the specialization's `shared_context_schema_id`. The runtime MUST expose no profile-local context extension.

Export MUST return `cartulary.extension_portability_export_result.v1` as exactly one closed variant:

```json
{"schema_id":"cartulary.extension_portability_export_result.v1","kind":"omit"}
```

or:

```json
{"schema_id":"cartulary.extension_portability_export_result.v1","kind":"payload","payload_schema_id":"...","payload_contract_major":1,"state_version":1,"canonical_payload_sha256":"...","payload_byte_size":0,"payload_ref":"..."}
```

`omit` means successful participation with no portable state and is distinct from participant absence or failure. In a `payload` result, `payload_schema_id` MUST equal the specialization output schema, `payload_contract_major` and `state_version` MUST be positive integers, `payload_byte_size` MUST equal the exact bytes and remain within `max_output_bytes`, and `payload_ref` MUST satisfy `extension_safe_logical_ref_v1`. The payload contract MUST define canonical bytes, exact ordering, empty-payload semantics, import compatibility, state-version behavior, and permitted Core references. An output that exceeds its declared byte, item, or nesting limit is invalid and MUST NOT be published.

Import MUST return one distinct `cartulary.extension_portability_import_preparation_result.v1` object containing exactly `schema_id`, `status='prepared'`, `participant_input`, `participant_input_sha256`, and `staged_output_refs[]`. `participant_input` MUST conform to the exact import transaction-participant input schema and is inert until supplied to EXT-REQ-219. Each `staged_output_refs[]` item MUST be an `extension_safe_logical_ref_v1` produced only through the operation's process-local scoped staged-output capability; rows MUST be unique and sorted by ascending UTF-8 bytes. The participant MUST NOT return an export payload result from import or an import preparation result from export.

Import invocation is side-effect-free preparation. The accessor may read only the admitted inert bundle payload, and the staged-output capability may allocate and upload bytes only under EXT-REQ-192; it cannot publish, redeem, mutate authoritative state, acquire a database transaction, resolve a physical storage identity, commit, roll back, or access another operation. Every authoritative import mutation and staged-output publication MUST occur later through the shared transaction protocol in one final commit. Preparation failure or rollback MUST abandon every scoped staged output. The canonical payload bytes exposed to one participant and the aggregate bytes exposed to all portability participants MUST each be at most `67108864` bytes; byte `67108865` MUST fail before participant invocation. No compatibility adapter for `cartulary.extension_portability_participant_result.v1` is permitted.

`cartulary.extension_snapshot_reporting_participant_context.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_snapshot_reporting_participant_context.v1`;
- `operation`, exactly `emit`;
- `profile_id`;
- `contract_major`;
- `claim_state`;
- `state_present`;
- `state_version`, an integer or `null`;
- `snapshot_ref`;
- `authorization_view_sha256`;
- `redaction_profile_sha256`;
- `timeout_seconds`.

`snapshot_ref` MUST satisfy `extension_safe_logical_ref_v1` and bind one immutable snapshot through the Snapshot/Reporting owner's scoped accessor. Its result MUST be exactly:

```json
{"schema_id":"cartulary.extension_snapshot_reporting_participant_result.v1","kind":"empty"}
```

or:

```json
{"schema_id":"cartulary.extension_snapshot_reporting_participant_result.v1","kind":"output","output_schema_id":"...","output_sha256":"...","output_byte_size":0,"output_ref":"...","item_count":0}
```

`empty` is successful participation with no emitted items. In an `output` result, `output_schema_id` MUST equal the specialization output schema, `output_byte_size` and `item_count` MUST equal the exact output and remain within the specialization limits, and `output_ref` MUST satisfy `extension_safe_logical_ref_v1`. The operation MUST apply the exact authorization, redaction, selection, and ordering contracts named by the specialization before returning a digest.

`cartulary.extension_backup_restore_participant_context.v1` MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_backup_restore_participant_context.v1`;
- `operation`;
- `profile_id`;
- `claim_state`;
- `state_present`;
- `state_version`, an integer or `null`;
- `physical_state_binding_sha256`;
- `logical_family_ids[]`;
- `timeout_seconds`.

`operation` MUST equal `backup_enumerate`, `restore_validate`, or `restore_rebuild`. `logical_family_ids[]` MUST equal the specialization operation's admitted state-family set, reject duplicates, and sort by UTF-8 bytes. `cartulary.extension_backup_restore_participant_result.v1` MUST be one operation-matching closed variant:

- `backup_enumerate`: exactly `schema_id`, `operation='backup_enumerate'`, `binding_ids[]`, and `object_reference_family_ids[]`;
- `restore_validate`: exactly `schema_id`, `operation='restore_validate'`, `status`, and `findings[]`, where `findings[]` contains `0..256` `cartulary.extension_transaction_participant_finding.v1` objects, `status='valid'` requires `findings=[]`, and `status='invalid'` requires `1..256` findings;
- `restore_rebuild`: exactly `schema_id`, `operation='restore_rebuild'`, `status='completed'`, and `rebuilt_family_ids[]`.

The `schema_id` in every backup/restore result MUST equal `cartulary.extension_backup_restore_participant_result.v1`. Result arrays MUST reject duplicates and sort by ascending UTF-8 bytes. Restore findings MUST follow the EXT-REQ-219 path, owner-mapping, secrecy, uniqueness, and sorting rules. The participant MUST NOT return a binding or family absent from the physical binding and specialization.

All participant inputs and results MUST reject unknown members, duplicate members, wrong types, and explicit invalid `null`. For a claimed-profile invocation, the shared owner MUST provide one process-local read-only input accessor constrained to the exact context references and `state_family_ids[]` in the specialization. The accessor MUST count every selected item's canonical bytes and item count against `max_input_bytes` and `max_items`, MUST apply the specialization ordering before exposure, and MUST reject any attempt to enumerate or read another incident, snapshot, payload, binding, or logical family. The accessor MUST NOT be serialized, persisted, logged, returned, or retained after invocation.

A result returned after its deadline is discarded under EXT-REQ-215. Backup and restore for inactive profiles MUST use the physical state binding and shared validators without profile code. A participant MUST NOT discover physical storage, incident data, another owner's state, secrets, or egress capability outside its exact context, scoped accessor, and specialization.

Profiles: base
Verified by: EXT-AC-117, EXT-AC-134

**EXT-REQ-140**
For `incident_portability_mode='participant'`, the named profile owner MUST specialize the context and result schemas in EXT-REQ-221. The specialization MUST define logical export and import schema IDs and versions, canonical payload bytes and digest, authorization inputs, state and claim behavior, validation and errors, byte and item bounds, ordering, target publication participation, and every compatibility case in EXT-REQ-198. Unknown payloads remain inert and MUST NOT be executed.

Profiles: base
Verified by: EXT-AC-066, EXT-AC-073, EXT-AC-094

**EXT-REQ-141**
For `incident_portability_mode='blocked_when_present'`, Incident Portability MUST evaluate the exact `cartulary.extension_state_blocking_predicate.v1` object under EXT-REQ-222. When it evaluates true, export MUST fail before bundle publication. Generic error `details` MUST contain exactly `profile_id`; a family ID, count, physical identity, or profile-authored detail is forbidden. The implementation MUST NOT silently omit state or publish a partial-success bundle.

Profiles: base
Verified by: EXT-AC-034, EXT-AC-064

**EXT-REQ-222**
`cartulary.extension_state_blocking_predicate.v1` is closed in contract major `1` to exactly:

```json
{
  "schema_id":"cartulary.extension_state_blocking_predicate.v1",
  "kind":"any_authoritative_state_present",
  "family_ids":["..."]
}
```

`family_ids[]` MUST contain `1..256` unique logical family IDs declared by the profile's state-presence manifest and sort by ascending UTF-8 bytes. The shared evaluator returns true when any named family contains at least one retained authoritative member.

Present state includes active state, soft-deleted retained state, failed-but-retained state, retained authoritative job or resource state assigned to the family, and migration metadata only when the profile owner explicitly classifies that metadata as authoritative state. Disposable projections, caches, temporary files, staged orphan objects, ordinary audit rows, and generic job rows do not count unless explicitly assigned to an authoritative family.

The evaluator MUST use shared state-presence bindings and MUST NOT invoke inactive or retired profile code, migrations, workers, dependency probes, external egress, or arbitrary callbacks. A new predicate kind requires a new schema or contract major.

Profiles: base
Verified by: EXT-AC-118

**EXT-REQ-142**
For the current adoption baseline, Network Flow Activity MUST use `incident_portability_mode='blocked_when_present'`. Blocking state is any retained authoritative `network_flow_table`, including a soft-deleted retained table, unless a later Network Flow contract major defines a portability participant.

Profiles: network_flow_activity
Verified by: EXT-AC-064, EXT-AC-072

**EXT-REQ-143**
Every profile MUST declare `snapshot_reporting_mode`. A profile with `no_participation` MUST NOT be queried or included by Snapshot and Reporting because it is claimed, visible in the UI, referenced by browser-local state, or present in physical backup. That omission is intentional and complete for contract major `1`.

Profiles: base
Verified by: EXT-AC-035, EXT-AC-065, EXT-AC-073, EXT-AC-094

**EXT-REQ-144**
A `snapshot_reporting_mode='participant'` owner MUST specialize the context and result schemas in EXT-REQ-221 and define exact snapshot inputs, immutable source references, state-presence and claim-state behavior, output schema, ordering, redaction interaction, empty-state and omission behavior, errors, byte and item limits, and deterministic derivation. Snapshot and Reporting owners retain orchestration and output ownership.

Profiles: base
Verified by: EXT-AC-035, EXT-AC-065, EXT-AC-094

**EXT-REQ-198**
Incident Portability MUST apply Tables 23-A and 23-B. The shared declarative state-presence evaluator MUST run whenever the selected matrix row depends on whether authoritative state is present. Omission behavior: it is not invoked for a row whose result is independent of state presence. Arbitrary profile code, profile migration, profile workers, external probes, and third-party egress MUST NOT run for an inactive profile during portability admission.

Profiles: base
Verified by: EXT-AC-094

**Table 23-A. Incident portability export matrix**

| Portability mode | Authoritative state present | Source claim state | Required export result |
| --- | ---: | --- | --- |
| `no_authoritative_incident_state` | No | Any | Emit no profile payload. |
| `no_authoritative_incident_state` | Yes | Any | Fail owner-integrity validation before bundle publication. |
| `participant` | No | Any | Shared owner returns `omit` and emits no profile payload without invoking profile code. |
| `participant` | Yes | Claimed | Invoke the typed participant; include only a valid `payload` result or omit on a valid `omit` result. |
| `participant` | Yes | Unclaimed or recognized unclaimable | Fail before bundle publication with `extension_state_unavailable_for_portability`. |
| `blocked_when_present` | No | Any | Continue without profile payload. |
| `blocked_when_present` | Yes | Any | Fail before bundle publication with the owner blocking reason. |

**Table 23-B. Incident portability import matrix**

| Profile payload | Target condition | Required import result |
| --- | --- | --- |
| Absent | Any | Perform no profile action. |
| Present | Profile unrecognized | Fail before target publication; retain no active payload interpretation. |
| Present | Profile recognized unclaimable | Fail before target publication. |
| Present | Profile claimable but unclaimed | Fail before target publication. |
| Present | Profile claimed with unsupported contract major, state version, lineage, schema, or digest | Fail before target publication. |
| Present | Profile claimed and compatible | Invoke the typed participant irrespective of prior target state and join final target publication. |

Contract major `1` defines no store-now-activate-later portability payload. Unknown or inactive payloads remain inert and cause import failure before publication.

**EXT-REQ-199**
Snapshot and Reporting MUST apply Table 23-C.

Profiles: base
Verified by: EXT-AC-094

**Table 23-C. Snapshot and Reporting claim-state matrix**

| Mode | Authoritative state present | Claim state | Required result |
| --- | ---: | --- | --- |
| `no_participation` | Any | Any | Do not query the profile and emit no profile output. |
| `participant` | No | Any | Shared owner returns `empty` and invokes no profile code. |
| `participant` | Yes | Claimed | Invoke the typed participant and accept only `empty` or `output`. |
| `participant` | Yes | Unclaimed or recognized unclaimable | Fail snapshot or render admission with `extension_state_unavailable_for_snapshot`. |

A future read-only inactive-state portability or reporting adapter requires a later contract major. An implementation MUST NOT create one locally in v1.

**EXT-REQ-145**
This NLSpec creates no generic whole-incident extension purge. A future Core incident-removal owner MUST explicitly include each extension owner, idempotency references, job proofs, migration metadata, retained diagnostics, bindings, object references, and audit-retention consequences in one owner-approved removal contract.

Profiles: base
Verified by: EXT-AC-041, EXT-AC-050

# 24. Security, secrets, and external egress

**EXT-REQ-146**
All executable extension code and browser assets MUST be packaged with the application and covered by the ordinary build, dependency, provenance, vulnerability, and release controls. Uploaded documents, reference packs, portability bundles, import files, database values, owner fragments, descriptors, bindings, and object-store objects MUST be treated as data, not executable extension packages.

Profiles: base
Verified by: EXT-AC-066, EXT-AC-070

**EXT-REQ-147**
Every claimed extension request, job read, job cancellation, workspace open, preview, download, and mutation MUST rederive current authentication and authorization through the Core 04 contract named by the route, job, or workspace owner. Profile claim, discovery visibility, workspace visibility, cached browser state, implementation binding, and `deployment_admin` alone MUST NOT authorize incident data.

Profiles: base
Verified by: EXT-AC-032, EXT-AC-034, EXT-AC-038

**EXT-REQ-148**
An extension MUST use Core 04 `secret_ref_v1` for deployment-local secret material. Raw secrets MUST NOT appear in dependency snapshots, owner fragments, descriptors, registries, integrity objects, bindings, discovery, browser state, incident data, workbook rows, public errors, startup findings, jobs, job proofs, audit output, readiness output, logs, telemetry, generated fixtures outside a harness-owned secret boundary, or WebSocket payloads.

Stage 2 secret validation MAY establish that a declared reference can be resolved through the Core 04 local secret boundary. Omission behavior: it MUST NOT reveal, persist in an extension artifact, transmit, compare in diagnostics, or consume the secret for third-party authentication during preflight.

Profiles: base
Verified by: EXT-AC-069

**EXT-REQ-149**
`egress_mode` defaults to `none`. A profile with `egress_mode='none'` MUST make no third-party request from its HTTP routes, workers, browser code, migrations, startup preflight, validation, background reconciliation, portability admission, reporting admission, or fatal shutdown. Calls to Cartulary's own Postgres, object storage, same-origin application routes, and deployment-local secret provider are not third-party egress.

Profiles: base
Verified by: EXT-AC-067

**EXT-REQ-150**
A profile with `egress_mode='owner_declared'` MUST define every item in Table 24-A. A missing item, unresolved locator, or open delegation blocks claimability.

Profiles: base
Verified by: EXT-AC-068, EXT-AC-073

**Table 24-A. Required egress declaration**

| Egress concern | Required owner contract |
| --- | --- |
| Destination | Exact configuration keys and destination validation. |
| Protocol | Exact allowed protocol and method family. |
| TLS | Certificate and hostname validation; no insecure bypass. |
| Allowed data | Exhaustive data categories permitted to leave the deployment. |
| Prohibited data | Exhaustive sensitive categories that MUST NOT leave. |
| Initiator | Exact role, route, action, and whether explicit user action is required. |
| Disclosure | User-visible disclosure before transmission when incident data leaves the deployment. |
| Audit | Exact admitted, success, failure, and cancellation occurrences. |
| Timeout | Exact default, maximum, and timeout result. |
| Retry | Exact retry count, retryable conditions, and deterministic backoff. |
| Rate limit | Exact deployment and actor limit. |
| Size limits | Exact request and response byte ceilings. |
| Cancellation | Observable cancellation and post-send behavior. |
| Disconnected behavior | Exact failure without fallback or hidden queueing. |
| Redaction | Exact transformation before transmission. |
| Secret handling | Exact `secret_ref_v1` keys and no-secret-output rules. |
| Partial result | Exact retain, discard, or publish behavior. |
| Startup probe | Exact read-only probe, timeout, safe result, and prohibition on external mutation. |

**EXT-REQ-151**
Browser code MUST NOT send incident-authored content, evidence bytes, filenames, indicators, hostnames, identities, investigative queries, or extension resource values directly to third-party endpoints. An owner-declared egress flow MUST be server-mediated unless a later adopted security owner defines a browser boundary explicitly.

Profiles: base
Verified by: EXT-AC-067, EXT-AC-068

**EXT-REQ-152**
A profile that accepts uploaded or imported data MUST define byte ceilings, record ceilings, nesting limits, archive limits when archive input is accepted, explicit archive rejection when it is not accepted, parsing timeouts, cancellation, inert-content handling, and safe diagnostics. HTML, scripts, macros, formulas, remote assets, executable metadata, and package manifests MUST remain inert unless an adopted owner defines one safe non-executable transformation.

Registry, binding, migration-definition, and job-proof digests are integrity evidence. They MUST NOT be accepted as passwords, bearer credentials, signing keys, authorization grants, or substitutes for current authentication and authorization.

Profiles: base
Verified by: EXT-AC-066, EXT-AC-069

**EXT-REQ-153**
The current revision MUST expose no route or operator command for extension install, uninstall, enable, disable, reload, executable upload, package download, package update, publisher registration, or package permission grant.

Profiles: base
Verified by: EXT-AC-038, EXT-AC-070

# 25. Observability and audit

**EXT-REQ-154**
The canonical resolved claim set MUST be the only input to any telemetry resource attribute that represents claimed profiles. A separate hand-maintained claim list is invalid.

Profiles: base
Verified by: EXT-AC-071, EXT-AC-072

**EXT-REQ-155**
Telemetry MAY identify a profile by low-cardinality `profile_id`, contract major, operation family, and safe terminal classification. Omission behavior: when profile-specific attributes are not emitted, the canonical claimed-profile resource identity remains the only profile telemetry requirement. Extension telemetry MUST NOT include incident-authored values, extension resource values, raw route inputs, filenames, source cells, provider assertions, secrets, object keys, or stable incident identifiers.

Profiles: base
Verified by: EXT-AC-069, EXT-AC-071

**EXT-REQ-156**
An extension MUST NOT configure or ship a separate telemetry exporter, browser-to-third-party telemetry path, or privacy policy. It MUST use the adopted OpenTelemetry subsystem and Core 04 deployment configuration.

Profiles: base
Verified by: EXT-AC-067, EXT-AC-071

**EXT-REQ-157**
Every extension mutation or high-value read that the named owner audit matrix marks as auditable MUST use the stable audit code and safe field set from that matrix. When the operation participates in an atomic commit, the audit occurrence or transactional outbox record MUST join that same commit.

Profiles: base
Verified by: EXT-AC-056, EXT-AC-058, EXT-AC-069

**EXT-REQ-158**
Exact committed idempotency replay MUST return the original success and original audit correlation without creating a duplicate extension domain audit occurrence.

Profiles: base
Verified by: EXT-AC-053, EXT-AC-057, EXT-AC-058

# 26. Errors and startup diagnostics

**EXT-REQ-233**
`extension_safe_logical_ref_v1` MUST have exactly this form:

```text
<namespace>:<kind>:<opaque-id>
```

`namespace` and `kind` MUST each match `[a-z][a-z0-9_.-]{0,63}`. `opaque-id` MUST match `[A-Za-z0-9][A-Za-z0-9._~-]{0,381}`. The complete value MUST contain `5..512` UTF-8 bytes and exactly two colons. A slash, backslash, whitespace, control character, credential, endpoint, physical storage identity, absolute path, or incident-authored text is forbidden. A safe logical reference is non-authorizing and MUST resolve only through the owning scoped accessor; parsing its opaque ID into a physical identity is forbidden.

The shared identifier closures are:

- `audit_correlation_id` and `final_commit_id` MUST match `[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}` and contain `1..160` ASCII bytes;
- `final_commit_id` MUST be unique within one deployment;
- committed replay MUST return the original, identical `audit_correlation_id` and `final_commit_id`;
- `operation_id`, `job_id`, and `cancellation_request_id` MUST import the exact scalar and bound from their adopted Core owner and MUST NOT define a local extension substitute; and
- `config_path` MUST use `extension_diagnostic_path_v1` against the normalized configuration view, including the exact normalized key path.

For malformed packaged bytes, `extension_registry_invalid.details.safe_ref` MUST derive only from the package slot's declared artifact kind and declared artifact identity, never from the malformed content. The value MUST be `cartulary:<artifact_kind>:<declared-artifact-id>` when that value satisfies this grammar. Otherwise it MUST be `cartulary:<artifact_kind>:sha256-<digest>`, where `<digest>` is the lowercase SHA-256 of the UTF-8 declared slot identity. It MUST NOT hash, quote, truncate, or encode malformed artifact bytes.

Profiles: base
Verified by: EXT-AC-138

**EXT-REQ-159**
Core 04 MUST add the startup-admission reason codes in Table 26-A under top-level `error.code='invalid_deployment_config'`. The three rows explicitly owned by transaction, backup/restore, and readiness use their stated non-startup envelopes. The fatal runtime reason `extension_integrity_failure` uses the Core runtime `service_unavailable` and fatal-readiness envelopes defined in §22. Profile-local startup codes MAY replace a generic code only under EXT-REQ-186.

Profiles: base
Verified by: EXT-AC-005, EXT-AC-009, EXT-AC-010, EXT-AC-012, EXT-AC-013, EXT-AC-016, EXT-AC-018, EXT-AC-028, EXT-AC-084

**Table 26-A. Generic extension reason-code registry**

| `reason_code` | Required use |
| --- | --- |
| `extension_registry_invalid` | Missing, malformed, stale, extra, unmatched, noncanonical, or owner-inconsistent contract artifact. |
| `extension_registry_limit_exceeded` | Static registry, fragment, fact, descriptor, binding, or artifact byte/count ceiling exceeded. |
| `extension_registry_conflict` | Collision defined by Table 13-A. |
| `extension_profile_not_claimable` | Configuration requests a recognized unclaimable profile. |
| `extension_implementation_unavailable` | Requested profile has no valid packaged binding or binding parity fails. |
| `extension_admission_validation_failed` | A preflight or post-migration validation algorithm returned an invalid result, failed, or timed out. |
| `extension_dependency_probe_failed` | A declared read-only dependency probe returned an invalid result, failed, or timed out. |
| `extension_dependency_not_claimed` | Required dependency is not explicitly claimed. |
| `extension_dependency_incompatible` | Required dependency has a different current contract major or incompatible binding. |
| `extension_config_without_claim` | Profile-local configuration is present while inactive and is not permitted by Table 21-B. |
| `extension_state_metadata_missing` | Authoritative state exists but state metadata does not. |
| `extension_state_incomplete` | State metadata exists without declared authoritative state while `empty_state_policy='forbidden'`. |
| `extension_state_version_unsupported` | Stored state is newer, too old, wrong-lineage, or has no complete forward path. |
| `extension_migration_path_too_long` | Migration-definition count exceeds the static maximum. |
| `extension_migration_definition_changed` | A committed migration identity has a different canonical definition digest. |
| `extension_migration_lock_timeout` | Profile migration lock was not acquired within the configured timeout. |
| `extension_migration_timeout` | One migration step or complete profile migration exceeded its configured timeout. |
| `extension_migration_failed` | A required forward migration or post-step validation failed. |
| `extension_reconciliation_limit_exceeded` | Inactive-profile nonterminal job count exceeds the configured maximum. |
| `extension_reconciliation_timeout` | Inactive-profile reconciliation exceeded its configured timeout. |
| `extension_unclaim_reconciliation_failed` | Nonterminal extension work cannot be reconciled safely while inactive. |
| `extension_application_process_active` | Another application process owns the deployment-global serving lease. |
| `extension_publication_failed` | Stage 6 could not reach the externally visible `serving` state. |
| `extension_transaction_timeout` | Cross-owner transaction deadline expired and absence of commit was proven; HTTP `503`, `error.code='service_unavailable'`, and `retryable=true`. |
| `backup_binding_codec_unsupported` | Backup metadata names a codec ID and digest that are not explicitly packaged for the binding; fail before restore mutation. |
| `extension_staged_object_cleanup_dependency_failed` | Authorization or storage configuration prevents physical staged-object deletion; Core readiness is `degraded_dependency`. |
| `extension_state_validation_failed` | Fresh, current-version, migrated, or restored profile state failed the final state validator. |
| `extension_validation_result_invalid` | An owner algorithm returned malformed validation-result bytes or a structurally invalid result. |
| `extension_diagnostic_overflow` | A validation phase or valid findings array would produce more than 4096 findings. |
| `extension_accounting_overflow` | Conformance accounting would produce more than 4096 findings. |
| `extension_integrity_failure` | Fatal runtime integrity condition in EXT-REQ-134. |

**Table 26-A1. Closed diagnostic detail token registries**

| Detail member | Closed values or derivation |
| --- | --- |
| `artifact_kind` | Exactly `dependency_snapshot`, `owner_contract_manifest`, `owner_fragment`, `owner_input_registry`, `profile_configuration_contract`, `profile_configuration_view`, `descriptor_source`, `descriptor`, `registry`, `registry_integrity`, `implementation_binding`, `client_support_registry`, `base_route_reservation_registry`, `physical_state_binding`, `state_presence_manifest`, `state_initialization_definition`, `backup_binding_codec`, `job_kind_contract`, `participant_contract`, `validation_condition_registry`, `contract_closure_catalog`, `admission_validation`, `admission_validation_context`, `admission_validation_result`, `generated_schema`, `generator_source`, `conformance_manifest`, `manifest_index`, `accounting`, or `clause_traceability`. |
| `phase` | Exactly `dependency_snapshot`, `owner_contract_manifest`, `owner_fragment`, `owner_input_registry`, `configuration_contract`, `descriptor_materialization`, `registry_generation`, `registry_collision`, `claim_configuration`, `process_lease`, `implementation_binding`, `dependency_validation`, `profile_preflight`, `migration`, `state_validation`, `post_migration_validation`, `dependency_probe`, `transaction`, `backup_restore`, `inactive_reconciliation`, `staged_object_cleanup`, `publication`, `normative_source_lint`, or `conformance_accounting`. |
| `collision_class` | Exactly one `collision_class` token from Table 13-A; owner-input completeness and unrecognized-profile failures are not collision classes. |
| Migration-timeout `operation_kind` | Exactly `migration_step` or `profile_migration`. |
| `fatal_condition` | Exactly `indeterminate_database_commit`, `in_memory_contract_digest_mismatch`, `migration_ledger_state_mismatch`, `application_process_lease_lost`, or `staged_object_publication_mismatch`. |

`staged_object_publication_mismatch` is the single safe classification for every digest, storage-identity, publication-state, or authoritative-reference contradiction in Table 20-B; it MUST NOT disclose the physical identity or select another fatal token.

**EXT-REQ-160**
Every generic finding MUST conform to `cartulary.extension_startup_finding.v1` and the reason-specific detail contract in Table 26-B. No unlisted detail member is valid.

Profiles: base
Verified by: EXT-AC-069, EXT-AC-084

**Table 26-B. Generic messages and exact detail members**

| Reason code | Exact `message` | Exact `details` members |
| --- | --- | --- |
| `extension_registry_invalid` | `The extension contract artifact set is invalid.` | `artifact_kind`, `safe_ref`, `expected`, `actual`; `expected` and `actual` are nullable. |
| `extension_registry_limit_exceeded` | `An extension registry limit was exceeded.` | `phase`, `limit`, `actual`. |
| `extension_registry_conflict` | `The extension registry contains conflicting declarations.` | `collision_class`, `conflicting_profile_ids[]`, `conflicting_tokens[]`, `duplicate_count`. |
| `extension_profile_not_claimable` | `The requested extension profile is not claimable.` | `profile_id`. |
| `extension_implementation_unavailable` | `The requested extension implementation is unavailable or incompatible.` | `profile_id`, `expected_contract_major`, `actual_contract_major`; `actual_contract_major` is nullable. |
| `extension_admission_validation_failed` | `Extension admission validation failed.` | `profile_id`, `phase`, `algorithm_id`, `timed_out`, `timeout_seconds`. |
| `extension_dependency_probe_failed` | `An extension dependency probe failed.` | `profile_id`, `probe_id`, `algorithm_id`, `timed_out`, `timeout_seconds`. |
| `extension_dependency_not_claimed` | `A required extension dependency was not explicitly claimed.` | `profile_id`, `dependency_profile_id`. |
| `extension_dependency_incompatible` | `A required extension dependency is incompatible.` | `profile_id`, `dependency_profile_id`, `expected_contract_major`, `actual_contract_major`. |
| `extension_config_without_claim` | `Extension configuration is present while the profile is inactive.` | `profile_id`, `config_path`. |
| `extension_state_metadata_missing` | `Authoritative extension state is missing required state metadata.` | `profile_id`, `migration_lineage_id`. |
| `extension_state_incomplete` | `Extension state metadata requires authoritative state under the declared empty-state policy.` | `profile_id`, `migration_lineage_id`. |
| `extension_state_version_unsupported` | `Stored extension state is not supported by the packaged implementation.` | `profile_id`, `migration_lineage_id`, `minimum_migratable_state_version`, `current_state_version`, `stored_state_version`. |
| `extension_migration_path_too_long` | `The extension migration path exceeds the supported limit.` | `profile_id`, `limit`, `actual`. |
| `extension_migration_definition_changed` | `A committed extension migration definition changed.` | `profile_id`, `migration_id`, `expected_digest`, `actual_digest`. |
| `extension_migration_lock_timeout` | `The extension migration lock could not be acquired before timeout.` | `profile_id`, `timeout_seconds`. |
| `extension_migration_timeout` | `Extension state migration exceeded its timeout.` | `profile_id`, `operation_kind`, `migration_id`, `from_state_version`, `to_state_version`, `timeout_seconds`; `migration_id`, `from_state_version`, and `to_state_version` are nullable for a profile-wide timeout. |
| `extension_migration_failed` | `Extension state migration failed.` | `profile_id`, `migration_id`, `from_state_version`, `to_state_version`. |
| `extension_reconciliation_limit_exceeded` | `Inactive extension job reconciliation exceeds the supported limit.` | `profile_id`, `limit`, `actual`. |
| `extension_reconciliation_timeout` | `Inactive extension job reconciliation exceeded its timeout.` | `profile_id`, `timeout_seconds`. |
| `extension_unclaim_reconciliation_failed` | `Extension-owned nonterminal work could not be reconciled safely.` | `profile_id`, `job_id`; `job_id` is nullable. |
| `extension_application_process_active` | `Another Cartulary application process is active for this deployment.` | `timeout_seconds`. |
| `extension_publication_failed` | `Extension publication did not reach the serving state.` | `phase`, `timeout_seconds`. |
| `extension_transaction_timeout` | `The extension transaction did not complete before its deadline.` | `operation_id`, `timeout_seconds`. |
| `backup_binding_codec_unsupported` | `The backup binding codec is not supported by this build.` | `profile_id`, `binding_id`, `backup_codec_id`, `backup_codec_sha256`. |
| `extension_staged_object_cleanup_dependency_failed` | `Staged-object cleanup cannot access its storage dependency.` | None; `details` is exactly `{}`. |
| `extension_state_validation_failed` | `Extension state failed final validation.` | `profile_id`, `phase`, `algorithm_id`. |
| `extension_validation_result_invalid` | `An extension validation algorithm returned an invalid result.` | `profile_id`, `phase`, `algorithm_id`, `actual`; `profile_id` and `algorithm_id` are nullable; `actual` MUST use a safe formatter token from Table 26-D. |
| `extension_diagnostic_overflow` | `Extension validation produced too many findings.` | `phase`, `limit`, `actual`. |
| `extension_accounting_overflow` | `Extension conformance accounting produced too many findings.` | `phase`, `limit`, `actual`. |
| `extension_integrity_failure` | `The process detected an extension integrity failure and is shutting down.` | `fatal_condition`. |

All identifiers and digests in details MUST satisfy their owning scalar contracts. `artifact_kind`, `phase`, `collision_class`, migration-timeout `operation_kind`, and `fatal_condition` MUST use Table 26-A1. `safe_ref` MUST satisfy EXT-REQ-233. Nullable `expected` and `actual` under `extension_registry_invalid` MUST each be either `null` or a non-secret UTF-8 string of `0..512` bytes. `conflicting_profile_ids[]` MUST contain every recognized profile ID whose normalized declaration participates in the collision, no other profile ID, and `0..256` unique values sorted by ascending UTF-8 bytes; it MUST be `[]` only when the collision is not attributable to a recognized profile. `conflicting_tokens[]` MUST contain `1..256` unique strings derived exactly by Table 13-A and sorted by ascending UTF-8 bytes. `duplicate_count` MUST equal the actual normalized declaration or edge count as a JSON integer in `2..2147483647` when the Table 13-A predicate requires two or more declarations or duplicate edges. It MUST be `null` exactly for `route_family_overlap`, `base_route_capture`, `dependency_self`, and `dependency_cycle`. `timed_out` is Boolean. `timeout_seconds` MUST equal the effective Table 9-C value. For `extension_admission_validation_failed`, diagnostic `phase` MUST be `profile_preflight` when context phase is `preflight` and `post_migration_validation` when context phase is `post_migration`. For `extension_dependency_probe_failed`, diagnostic `phase` is not a details member and the startup phase is `dependency_probe`. `algorithm_id` and `probe_id` MUST satisfy their owning identifier contracts. `limit` and numeric `actual` values are non-negative JSON integers. `timeout_seconds` for `extension_application_process_active`, `extension_publication_failed`, and `extension_transaction_timeout` MUST equal the effective Table 9-C value for the applicable operation. For diagnostic or accounting overflow, `limit` MUST equal `4096` and `actual` MUST equal `4097`, representing the first finding beyond the permitted set. A string-valued `actual` MUST be one exact safe formatter output admitted by Table 26-D.

**EXT-REQ-161**
Diagnostics MUST NOT include raw secret values, transformed secret values, incident content, provider assertions, access tokens, table names, column names, SQL, object keys, bucket names, database names, raw endpoint credentials, source cells, file bytes, raw cryptographic failure details, or absolute filesystem paths. A `safe_ref` MUST be a non-secret `owner_locator_v1` or `extension_safe_logical_ref_v1`. Malformed packaged bytes use the package-slot derivation in EXT-REQ-233.

Profiles: base
Verified by: EXT-AC-069

**EXT-REQ-162**
`extension_diagnostic_path_v1` MUST use Table 26-E.

**Table 26-E. Diagnostic path grammar**

| Path part | Required syntax |
| --- | --- |
| Root | `$` |
| Object member matching ASCII path identifier `[A-Za-z_][A-Za-z0-9_]*` | `.` followed by the exact member name |
| Other object member | `[` + canonical JSON string for the exact member name + `]` |
| Array element | `[` + zero-based decimal index with no leading zero except `0` + `]` |

Paths identify normalized input or contract artifact locations, not source-code, database, or UI locations.

Within one phase, findings MUST sort by:

1. `path` ascending UTF-8 bytes;
2. `reason_code` ascending UTF-8 bytes;
3. canonical `details` bytes under `extension_registry_canonical_json_v1`.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-021, EXT-AC-084

**EXT-REQ-186**
`cartulary.extension_startup_finding.v1` MUST be a closed object containing exactly:

- `path`;
- `reason_code`;
- `message`;
- `details`.

`path` MUST satisfy `extension_diagnostic_path_v1`. A generic reason code MUST use the exact Table 26-B message and exact closed details object. A profile-local reason code MAY replace one generic reason only when the adopted profile owner contains an exact mapping from that generic code to the local code, local exact message, local closed details schema, startup phase, retryability, and public status when the mapped error is public; a non-public mapped error MUST declare that no public status applies. Without that mapping, the generic code is mandatory.

A validation phase MUST stop constructing ordinary findings when it would construct finding 4097. It MUST discard the ordinary phase finding list and emit exactly one `extension_diagnostic_overflow` finding at path `$`, with `phase`, `limit=4096`, and `actual=4097`. It MUST fail readiness. It MUST NOT emit a truncated list as complete or continue scanning merely to calculate a larger actual count.

An owner-produced result with a decoded `findings` array of 4097 or more elements MUST use EXT-REQ-225. The implementation MUST NOT classify that result through a profile-local replacement before the generic structural and overflow precedence has been applied.

Profiles: base
Verified by: EXT-AC-084, EXT-AC-119, EXT-AC-120

**EXT-REQ-224**
Every validation owner MUST author one closed `cartulary.extension_validation_surface_declaration.v1` containing exactly `schema_id`, `owner_contract_ref`, `schema_surfaces[]`, and `procedural_surfaces[]`. Every reachable invalid condition in a schema surface MUST carry one explicit condition annotation binding its schema location and constraint to a condition ID. Every procedural surface MUST provide a closed, exhaustive decision table whose mutually exclusive rows cover every outcome and bind every invalid row to one condition ID. Empty arrays are present; omission and explicit `null` are invalid. A declaration with a missing, duplicate, stale, extra, overlapping, or unreachable row is invalid.

Coordinated generation MUST produce one canonical `cartulary.extension_validation_condition_registry.v1` object solely from those declarations and the digest-bound owner inputs. It MUST contain exactly:

- `schema_id`, exactly `cartulary.extension_validation_condition_registry.v1`;
- `conditions[]`, containing exactly one row for every invalid condition reachable through dependency and owner-input derivation, descriptor materialization, registry generation, startup admission, migration, participant validation, runtime contract validation, or conformance accounting.

Each `conditions[]` row MUST contain exactly:

- `condition_id`;
- `phase`;
- `condition_class`;
- `path_algorithm_id`;
- `reason_code`;
- `expected_formatter_id`;
- `actual_formatter_id`;
- `multiplicity`;
- `secret_policy`;
- `owner_contract_ref`.

`condition_id` MUST satisfy `^[a-z][a-z0-9_]{0,127}$` and be globally unique. `phase` MUST use Table 26-A1. `condition_class` MUST use Table 26-C. `path_algorithm_id`, both formatter IDs, and `owner_contract_ref` MUST resolve through an adopted owner contract manifest. `multiplicity` MUST equal `single`, `one_per_occurrence`, `one_per_identity`, or `one_per_profile`. `secret_policy` MUST equal `safe_value` or `redacted`. A row with `secret_policy='redacted'` MUST use `actual_formatter_id='diagnostic_redacted_v1'`.

Rows MUST sort by `phase`, then `condition_id`, using ascending UTF-8 bytes. The array MUST contain `1..16384` rows. The canonical object MUST contain `1..16777216` bytes and serialize under `extension_registry_canonical_json_v1`. `extension_validation_condition_registry_sha256_v1` is the lowercase SHA-256 digest of that canonical byte form.

The registry MUST include exact rows for transaction deadline expiry with proven absent commit using `extension_transaction_timeout`, unsupported backup codec selection using `backup_binding_codec_unsupported`, and staged-object cleanup authorization or storage-configuration failure using `extension_staged_object_cleanup_dependency_failed`. Their paths, formatters, multiplicity, secrecy, and owner locators MUST match EXT-REQ-219, EXT-REQ-235, and Table 20-B; omission of any row is invalid.

An implementation MUST NOT infer a diagnostic path, reason code, expected value, actual value, multiplicity, or secret policy from local validation-library behavior. Every schema constraint and procedural decision row MUST resolve to exactly one registry condition, and every registry condition MUST be reachable from exactly one declared surface. A condition absent from the canonical registry is a specification or generator defect and MUST fail conformance rather than produce an implementation-selected diagnostic. A validator that attempts to emit an unregistered condition MUST terminate that validation surface with `extension_validation_result_invalid`; it MUST NOT expose the invented condition or library error.

Every public error token named by an extension profile but owned elsewhere MUST resolve through `owner_contract_ref` to the exact owner contract for HTTP status or process outcome, `error.code`, reason-code vocabulary, retryability, and closed safe details. A locator that proves only the token spelling is insufficient.

Profiles: base
Verified by: EXT-AC-119, EXT-AC-126, EXT-AC-147

**Table 26-C. Closed validation condition classes**

| `condition_class` | Required scope |
| --- | --- |
| `invalid_utf8` | Input bytes are not valid UTF-8. |
| `invalid_json` | Input bytes are not syntactically valid JSON. |
| `duplicate_member` | A closed object contains the same decoded member name more than once. |
| `unknown_member` | A closed object contains an undeclared member. |
| `missing_required_member` | A non-defaulted required member is absent. |
| `wrong_type` | A value has the wrong JSON type category. |
| `explicit_null` | Explicit JSON `null` is supplied where null is forbidden. |
| `invalid_scalar` | A scalar fails its exact grammar or domain. |
| `noncanonical_bytes` | Valid semantic input is not in the required canonical byte form. |
| `digest_mismatch` | An expected and actual digest differ. |
| `locator_not_found` | No anchor of the required identity exists. |
| `locator_kind_mismatch` | An anchor exists under a different kind. |
| `locator_range_invalid` | A declared anchor byte range is invalid or stale. |
| `duplicate_identity` | A normalized semantic identity occurs more than once. |
| `multiplicity_mismatch` | A required exact, minimum, or maximum declaration count is violated. |
| `limit_exceeded` | A byte, count, depth, time, or work limit is exceeded. |
| `collision` | Two or more otherwise valid declarations collide under a closed collision predicate. |
| `algorithm_result_invalid` | An invoked algorithm returns bytes or a value outside its result schema. |
| `timeout` | A declared deadline expires. |
| `dependency_unavailable` | A required scoped storage or authorization dependency cannot be used. |
| `commit_indeterminate` | A durable mutation or commit outcome cannot be proven. |

**Table 26-D. Closed diagnostic formatter outputs**

| Formatter ID | Exact output rule |
| --- | --- |
| `diagnostic_missing_v1` | Exactly `missing`. |
| `diagnostic_json_type_v1` | Exactly `type:<token>`, where token is `null`, `boolean`, `integer`, `string`, `array`, or `object`. |
| `diagnostic_exact_token_v1` | Exactly `exact:<token>`; `<token>` MUST be a non-secret closed-vocabulary value of `1..128` ASCII bytes. |
| `diagnostic_integer_range_v1` | Exactly `integer:<minimum>..<maximum>` using base-10 integers. |
| `diagnostic_count_v1` | Exactly `count:<decimal>` using a non-negative base-10 integer. |
| `diagnostic_sha256_v1` | Exactly `sha256:<64-lowercase-hex>`. |
| `diagnostic_safe_logical_ref_v1` | Exactly one `extension_safe_logical_ref_v1` value; no quoting or transformation. |
| `diagnostic_redacted_v1` | Exactly `redacted`. |

A configuration value classified as `secret_ref`, `regular_file_ref`, `trust_material_ref`, or `diagnostic_policy='name_only'` MUST use `diagnostic_redacted_v1`. Raw value interpolation, library exception text, parser excerpts, SQL errors, storage names, or absolute paths are forbidden.

**EXT-REQ-225**
Owner-produced admission, migration-validation, state-validation, participant-validation, and conformance-check invocations and result bytes MUST use this precedence before owner-local semantic findings are interpreted:

1. invocation failure, panic, cancellation without a schema result, or inaccessible result bytes MUST fail with the invocation surface's registered generic failure;
2. invalid UTF-8, invalid JSON, a non-object result, or an absent or non-array `findings` member MUST fail with `extension_validation_result_invalid`;
3. a `findings` array containing `4097` or more elements MUST discard every ordinary element and produce exactly one `extension_diagnostic_overflow` finding with `limit=4096` and `actual=4097`;
4. after the count is below overflow, any remaining closed-schema defect in the result or an individual finding, including a result whose owning schema permits at most `256` findings but contains `257..4096`, MUST fail with `extension_validation_result_invalid`;
5. otherwise a schema-valid nonempty finding set MUST be normalized and sorted under its owning result contract and returned as that contract's valid semantic invalidity result;
6. otherwise a schema-valid empty finding set MUST return that contract's valid empty success result.

Invocation failure precedes all byte classification because no valid result exists. Structural invalidity precedes overflow unless a valid top-level findings array count is available. Overflow precedes defects inside individual findings. Counts `257..4096` are ordinary violations of an owning `0..256` result schema, never overflow. Counts `4097+` always report bounded `actual=4097`, not the complete larger count. No truncated ordinary list is conformant.

Profiles: base
Verified by: EXT-AC-120, EXT-AC-143

**EXT-REQ-163**
Invalid requested-profile admission MUST cause process exit code `2` and MUST start no HTTP listener, WebSocket listener, common job runner, or profile worker. The runtime MUST NOT serve a partial profile set when any explicitly requested claim fails.

Profiles: base
Verified by: EXT-AC-013, EXT-AC-020, EXT-AC-090

# 27. Conformance contracts and Harness v2 evidence

**EXT-REQ-164**
Coordinated adoption MUST provide every repo-control artifact and Harness v2 input in Table 27-A. Each extension artifact MUST be generated or authored by its named owner and included in drift accounting. Harness inputs and retained artifacts MUST validate under `cartulary.testing_harness.current.v2`. Physical file paths, Make targets, result roots, run IDs, scheduling, fixture lifecycle, cleanup, and retained-run metadata remain owned exclusively by the Testing Harness or its authored implementation-support manifests.

Profiles: base
Verified by: EXT-AC-004, EXT-AC-072, EXT-AC-073, EXT-AC-076, EXT-AC-081

**Table 27-A. Required extension artifacts and Harness v2 inputs**

| Artifact or input | Required owner or producer |
| --- | --- |
| `cartulary.extension_dependency_declaration_set.v1` | Extensions specification owner |
| `cartulary.extension_dependency_snapshot.v1` | Extensions generator from adopted dependencies |
| `cartulary.extension_owner_contract_manifest.v1` schema and one manifest per dependency owner document | Named owner plus coordinated adoption tooling |
| `cartulary.extension_owner_fragment.v1` schema and zero or more exactly adopted fragments per contributing owner | Named owner documents |
| `cartulary.extension_owner_input_registry.v1` | Extensions generator |
| `cartulary.extension_owner_fact_identity.v1` schema and derivation vectors | Extensions Subsystem |
| `cartulary.extension_profile_configuration_contract.v1` schema and one contract per claimable profile | Extensions Subsystem shape; named profile owner content |
| `cartulary.extension_profile_configuration_view.v1` schema | Core 04 plus Extensions Subsystem |
| `cartulary.extension_profile_descriptor_source.v1` schema | Extensions Subsystem; instances are ephemeral under EXT-REQ-209 |
| `cartulary.extension_profile_descriptor.v1` schema and one descriptor per recognized profile | Extensions generator |
| `cartulary.extension_profile_registry.v1` schema and canonical registry | Extensions generator |
| `cartulary.extension_registry_integrity.v1` | Extensions generator and build packaging |
| `cartulary.base_route_reservation_registry.v1` | Core 01 |
| `cartulary.client_extension_support_registry.v1` | Packaged client build |
| `cartulary.client_asset_set_manifest.v1` | Packaged client build |
| `cartulary.extension_workspace_availability.v1` | Core 01 plus Extensions Subsystem |
| `cartulary.extension_publication_plan.v1` and its six canonical component schemas | Extensions Subsystem plus Core 04 startup orchestration |
| `cartulary.extension_implementation_binding.v1` schema and packaged bindings | Build system and packaged implementation |
| `cartulary.extension_admission_validation.v1` schema | Extensions Subsystem |
| `cartulary.extension_admission_validation_context.v1` schema | Extensions Subsystem |
| `cartulary.extension_admission_validation_result.v1` schema | Extensions Subsystem |
| `cartulary.extension_migration_context.v1` and migration apply/validation/final-state result schemas | Extensions Subsystem |
| `cartulary.extension_state_presence_manifest.v1`, canonical digest artifacts, and golden vectors | Each `extension_versioned` profile owner plus Extensions generator |
| `cartulary.extension_state_initialization_definition.v1`, context, and result schemas | Extensions Subsystem shape; each `extension_versioned` profile owner supplies one definition |
| `cartulary.extension_physical_state_binding.v1` | Build system plus each durable profile owner |
| `cartulary.extension_backup_binding_codec.v1` and codec/digest vectors | Build system, backup owner, and each durable profile owner |
| `cartulary.extension_state_metadata.v1` logical schema | Extensions Subsystem |
| `cartulary.extension_migration_ledger_entry.v1` logical schema | Extensions Subsystem |
| `cartulary.extension_job_kind_contract.v1` and one contract per declared job kind | Extensions Subsystem shape; named profile owner content |
| `cartulary.extension_job_commit_proof.v1` and cancellation-observation schemas | Core 01 common jobs plus Extensions Subsystem |
| `cartulary.extension_transaction_participant_contract.v1`, shared context/result schemas, and `cartulary.extension_transaction_participant_finding.v1` | Core 01 plus Extensions Subsystem |
| `cartulary.extension_participant_specialization.v1` plus distinct portability export/import-preparation, Snapshot/Reporting, and backup/restore participant context/result schemas and specialization contracts | Applicable shared and profile owners |
| `cartulary.extension_state_blocking_predicate.v1` | Extensions Subsystem plus applicable profile owner |
| `cartulary.extension_staged_object.v1` logical schema | Core 01 object-storage owner plus Extensions Subsystem |
| `cartulary.extension_validation_surface_declaration.v1` per validation owner | Each validation owner |
| `cartulary.extension_validation_condition_registry.v1` | Extensions generator plus every validation owner |
| `cartulary.extension_startup_finding.v1` | Core 04 plus Extensions Subsystem |
| `cartulary.extension_contract_closure_catalog.v1` per claimable profile | Extensions generator |
| `cartulary.extension_conformance_manifest.v1` and `cartulary.extension_conformance_manifest_index.v1` | Named profile owners plus Extensions generator |
| `cartulary.extension_registry_accounting.v1` | Extensions generator from static contract inputs |
| `cartulary.extension_traceability_mapping_source.v1` | Extensions specification owner |
| `cartulary.extension_clause_traceability.v1` | Specification owner plus documentation-generation tooling |
| Canonicalization and normative-source-lint golden vectors | Extensions Subsystem plus documentation-generation tooling |
| `extension_safe_logical_ref_v1` schema and golden vectors | Extensions Subsystem |
| `cartulary.verification_registry.v1` and owner `cartulary.verification_contract.v1` inputs | Harness v2 verification owners |
| `cartulary.test_owner_registry.v1` and `cartulary.test_family_manifest.v1` inputs | Harness v2 test owners |
| `cartulary.test_runner_registry.v1` and authored runtime/resource/fixture profiles | Testing Harness v2 |
| `cartulary.test_slice_plan.v2` and `cartulary.test_slice_scheduler_summary.v1` | Harness v2 owner-slice planner and scheduler |
| `cartulary.test_evidence_accounting.v1` and `cartulary.test_owner_summary.v1` | Harness v2 owner-aware target finalizers |
| `cartulary.test_evidence_root_manifest.v1` and `cartulary.test_evidence_audit_summary.v1` | Harness v2 evidence-audit caller and auditor |

The four removed extension-fixture schema families have no successor alias. Product coverage is represented by owner requirements and acceptance criteria, routed through v2 verification IDs and exact catalog selectors. Harness v2 artifacts prove execution only; they MUST NOT enter the extension registry, descriptor, runtime admission set, or owner-fragment authority chain.

**EXT-REQ-183**
`cartulary.extension_conformance_manifest.v1`, `cartulary.extension_conformance_manifest_index.v1`, and `cartulary.extension_registry_accounting.v1` MUST satisfy Tables 27-A1 through 27-A6.

Profiles: base
Verified by: EXT-AC-081, EXT-AC-095, EXT-AC-121, EXT-AC-122

**Table 27-A1. `cartulary.extension_conformance_manifest.v1`**

| Member | Type and required rule |
| --- | --- |
| `schema_id` | Exactly `cartulary.extension_conformance_manifest.v1`. |
| `conformance_manifest_id` | Exactly `<profile_id>.conformance.v<contract_major>`. |
| `profile_id` | Exact claimable profile ID. |
| `contract_major` | Exact positive current contract major. |
| `descriptor_sha256` | Exact `extension_descriptor_sha256_v1` digest. |
| `contract_closure_catalog_sha256` | Exact `extension_contract_closure_catalog_sha256_v1` digest. |
| `owner_contract_refs[]` | `1..4096` unique owner locators, sorted by UTF-8 bytes. |
| `requirement_ids[]` | `1..4096` unique ASCII owner requirement IDs of `1..128` bytes, sorted by numeric suffix where the family has one, then UTF-8 bytes. |
| `acceptance_criterion_ids[]` | `1..4096` unique ASCII owner acceptance IDs of `1..128` bytes, sorted by numeric suffix where the family has one, then UTF-8 bytes. |
| `verification_ids[]` | `1..4096` unique Harness v2 verification IDs, sorted by ascending ASCII bytes. |
| `public_schema_ids[]` | Exact canonical descriptor array. |
| `contribution_ids[]` | Exact descriptor contribution IDs, sorted by UTF-8 bytes. |
| `contract_closure[]` | Exactly one resolution row per catalog item, ordered by matching closure-catalog item order. |

`verification_ids[]` MUST equal the complete set of active v2 verification contracts that route the manifest profile's extension-owned postconditions and imported shared-owner postconditions. Every ID MUST satisfy the Harness v2 grammar, resolve exactly once through the current verification registry, and use `profile='base'` or `profile='extension.<profile_id>'` as allocated by its behavior owner. A support or claim-publication verification MUST NOT appear in this product conformance set. Catalog row IDs, runner selectors, Make targets, runtime profiles, result roots, and retained artifacts MUST NOT appear in the manifest.

`owner_contract_refs[]` MUST equal the complete union of every owner locator that supplied a descriptor fact, every locator in a `contract_closure[]` row with `status='specified'`, and every locator required by the profile's configuration, migration, job, state, participant, public-schema, and verification obligations.

The manifest MUST contain no run timestamp, run ID, result root, manually asserted pass/fail claim, secret, incident value, implementation package path, physical storage name, absolute path, catalog row ID, or selector. It MUST serialize under `extension_registry_canonical_json_v1`. `extension_conformance_manifest_sha256_v1` is the lowercase SHA-256 digest of that canonical byte form.

Each `contract_closure[]` row MUST contain exactly `closure_item_id`, `category`, `status`, `owner_contract_refs[]`, and `not_applicable_reason_code`. `status` MUST equal `specified` or `not_applicable`. A specified row requires `1..32` resolved owner locators and a null reason. A not-applicable row requires an empty locator array and one permitted Table 27-A2 reason. A Core-owned behavior MUST use `specified` with the Core locator.

**Table 27-A2. Closed contract-closure not-applicable reasons**

| Reason code | Required meaning |
| --- | --- |
| `no_public_interface` | The profile exposes no public or shared-owner callable interface for the addressed item. |
| `read_only_profile` | The addressed item concerns mutation, idempotency, commit, or concurrency and the profile is strictly read-only. |
| `no_durable_state` | The profile owns no durable state for the addressed item. |
| `no_pagination` | The addressed collection is explicitly non-pageable or no addressed collection exists. |
| `no_jobs` | The profile owns and produces no durable job for the addressed item. |
| `no_client_surface` | The profile exposes no browser workspace or client action for the addressed item. |
| `no_external_egress` | `egress_mode='none'`, no external dependency probe exists, and the addressed item is egress-only. |
| `no_portability_participation` | The selected portability mode has no participant, and the addressed item is participant-specific rather than mode semantics. |
| `no_snapshot_reporting_participation` | `snapshot_reporting_mode='no_participation'`, and the addressed item is participant-specific rather than mode semantics. |
| `no_cross_owner_transaction` | The addressed item can never commit state with another owner. |
| `no_profile_configuration` | The profile configuration contract contains zero profile-local keys other than the claim key. |
| `no_migrations` | The profile owns no extension-versioned state or its current lineage requires zero migration definitions. |
| `no_backup_participation` | The profile owns no durable state and therefore no backup/restore participant. |
| `no_runtime_worker` | The profile defines no worker kind and no background execution. |

**Table 27-A3. `cartulary.extension_conformance_manifest_index.v1`**

| Member | Required rule |
| --- | --- |
| `schema_id` | Exactly `cartulary.extension_conformance_manifest_index.v1`. |
| `manifests[]` | One row per claimable descriptor, sorted by `conformance_manifest_id`. |

Each `manifests[]` row MUST contain exactly `conformance_manifest_id`, `profile_id`, `contract_major`, `manifest_sha256`, and `safe_ref`. `safe_ref` MUST be an `extension_safe_logical_ref_v1` whose namespace is exactly `extensions`. Each claimable descriptor MUST resolve to exactly one index row and matching manifest. An unclaimable descriptor with `conformance_manifest_id=null` MUST have no row and no manifest.

**Table 27-A4. `cartulary.extension_registry_accounting.v1`**

| Member | Type and required rule |
| --- | --- |
| `schema_id` | Exactly `cartulary.extension_registry_accounting.v1`. |
| `registry_sha256` | Exact `extension_registry_sha256_v1` digest. |
| `registry_integrity_sha256` | Exact `extension_registry_integrity_sha256_v1` digest. |
| `verification_semantic_digest` | Exact current Harness v2 verification-contract semantic digest. |
| `catalog_semantic_digest` | Exact current Harness v2 owner-catalog semantic digest. |
| `status` | Exactly `pass` or `fail`. |
| `checks[]` | Exactly one row for every required Table 27-A5 check instance. |
| `findings[]` | `0..4096` finding objects, sorted by nullable profile ID, check ID, reason code, and safe reference. |

Each `checks[]` row MUST contain exactly `check_id`, `profile_id`, `status`, `input_digests[]`, and `safe_refs[]`. `profile_id` MUST be null exactly for registry-global checks. Each input-digest row contains exactly `artifact_id` and `sha256`, rejects duplicate artifact IDs, and sorts by `artifact_id`. Extension artifact digests use their owning lowercase hexadecimal contract; Harness semantic digests retain the exact `sha256:<64-lowercase-hex>` v2 form. `safe_refs[]` contains `0..64` unique non-secret owner locators or extension-safe logical references sorted by UTF-8 bytes.

A finding contains exactly `profile_id`, `check_id`, `reason_code`, and `safe_ref`. `status='pass'` if and only if every required static check instance is present exactly once, every check passes from the declared current input digests, and `findings=[]`. Finding overflow MUST fail with `extension_accounting_overflow` without emitting a truncated accounting object.

**EXT-REQ-227**
Registry accounting MUST be computed only by the named deterministic static predicates in Table 27-A5. A human assertion, test-run exit code, owner-summary status, evidence-audit status, broad target success, retained prior run, or implementation-local heuristic MUST NOT populate a passing registry-accounting check.

Every registry-scoped row occurs exactly once with `profile_id=null`. Every profile-scoped row occurs exactly once for each recognized profile. Missing, duplicate, malformed, stale, or mismatched static input MUST produce a failing check and at least one Table 27-A6 finding; it MUST NOT omit the check. Registry accounting validates routing and catalog structure only. Harness v2 retained evidence independently proves execution under EXT-REQ-236.

Profiles: base
Verified by: EXT-AC-122

**Table 27-A5. Required static accounting checks and predicates**

| `check_id` | Scope | Exact predicate input and pass condition |
| --- | --- | --- |
| `registry_artifact_set_match` | registry | Dependency snapshot, owner manifests/input, descriptors, registry, integrity object, bindings, supporting contracts, schemas, and generator inputs equal the integrity-object identity and digest sets. |
| `validation_condition_registry_match` | registry | Every reachable invalid condition has exactly one current registry row, every row resolves, and the registry digest equals the integrity object. |
| `normative_source_lint_match` | registry | This NLSpec and every adopted profile owner pass EXT-REQ-229 under exact current document digests. |
| `verification_registry_match` | registry | The current v2 verification registry resolves each required owner contract exactly once with no inactive, support-only, claim-only, or unknown substitute. |
| `test_catalog_match` | registry | Every required verification ID is referenced by at least one active exact-selector catalog row owned under the Harness v2 owner algorithm, and no row derives behavior from documentation text. |
| `dependency_snapshot_match` | profile | The profile's complete imported dependency, anchor, schema, algorithm, and artifact set equals the current dependency snapshot. |
| `owner_input_match` | profile | Normalized facts used for the descriptor equal exact owner-fragment facts and canonical fact identities. |
| `core00_match` | profile | Recognition, claimability, contract major, owner identity, and runtime dependencies equal Core 00 facts. |
| `core01_discovery_match` | profile | Generic discovery, route families, capabilities, and current major equal the descriptor and Core 01 producer contract. |
| `core03_workspace_match` | profile | Workspace declarations and client eligibility inputs equal Core 03 and the descriptor. |
| `core04_claim_configuration_match` | profile | Claim key, default, explicit-null behavior, and configuration-contract digest equal Core 04 and the descriptor. |
| `owner_contract_match` | profile | Every owner locator, requirement, schema, algorithm, configuration, migration, job, state, and participant declaration resolves through current owner manifests. |
| `implementation_binding_match` | profile | The packaged binding satisfies every descriptor and owner-contract parity rule and supplies no undeclared behavior. |
| `client_support_match` | profile | The packaged client support registry advertises only current descriptor majors, workspaces, capabilities, and schemas and binds the verified asset-set digest. |
| `physical_state_binding_match` | profile | State-presence, implementation-owned stores, initialization definitions, backup codecs, backup inclusion, validators, and rebuild declarations have exact parity. |
| `job_kind_contract_match` | profile | Every declared job kind has exactly one digest-bound contract and every proof/cancellation declaration matches it. |
| `participant_contract_match` | profile | Every contribution requiring a participant resolves to one current shared interface and specialization digest, with no extra participant. |
| `telemetry_match` | profile | The OpenTelemetry claimed-profile representation derives from the same resolved claim-set identity and current profile major. |
| `conformance_manifest_match` | profile | Claimable descriptors resolve to exactly one matching manifest and index row; unclaimable descriptors resolve to neither. |
| `contract_closure_match` | profile | The closure catalog digest is current and every catalog item has exactly one valid manifest resolution. |
| `verification_contract_match` | profile | Manifest verification IDs equal the current active v2 contracts allocated to that profile's owner postconditions. |
| `catalog_coverage_match` | profile | Current catalog rows reference every required verification ID, use exact non-overlapping selectors and admitted profiles, and contain no historical or document-derived identity. |

**Table 27-A6. Closed accounting finding reasons**

| `reason_code` | Required meaning |
| --- | --- |
| `missing_check` | A required check instance is absent. |
| `duplicate_check` | A required check identity occurs more than once. |
| `missing_input` | A predicate input artifact or semantic digest is absent. |
| `duplicate_input` | An input artifact identity occurs more than once. |
| `input_digest_mismatch` | The supplied digest differs from the current exact artifact or semantic digest. |
| `semantic_mismatch` | Canonical inputs are valid but the named predicate comparison fails. |
| `unexpected_artifact` | An undeclared artifact, manifest, binding, participant, verification contract, or catalog row is present. |
| `invalid_check_result` | The check object or predicate result violates its closed schema. |
| `stale_contract` | A traceability, lint, verification, catalog, or owner input is bound to older source bytes. |

**EXT-REQ-165**
The v2 verification contracts introduced for the Extensions Subsystem are exactly:

- `module.extensions.verification.behavior_contract`, for shared Base Profile extension runtime postconditions;
- `module.extensions.verification.contract_accounting`, for owner-input, generated-contract, registry-integrity, closure, diagnostic, and static routing postconditions.

Both IDs are immutable Harness v2 verification IDs owned by `module.extensions`. Coordinated adoption MUST add one active closed `cartulary.verification_contract.v1` owner file and its verification-registry row together with one active `module.extensions` test-owner row and a nonempty active family manifest. An active zero-row owner, unresolved verification ID, or partially adopted owner input is invalid. Imported Core, web, platform, application, and named-profile postconditions MUST continue through their primary owner's active verification IDs under Table 1-A and the Harness v2 ownership algorithm; `module.extensions` MUST NOT duplicate an imported owner's verification contract.

Profiles: base
Verified by: EXT-AC-073, EXT-AC-097

**EXT-REQ-166**
Every normative requirement MUST map to at least one acceptance criterion. Every acceptance criterion MUST map to one or more active Harness v2 verification IDs, and every verification ID cited by this NLSpec MUST map back to at least one current requirement and acceptance criterion. The mappings MUST be carried by `cartulary.extension_clause_traceability.v1` and validated against the current verification registry. Unmapped requirements, criteria, or verification contracts block adoption.

Profiles: base
Verified by: EXT-AC-002, EXT-AC-073, EXT-AC-097

**EXT-REQ-202**
Traceability MUST operate at independently observable clause level. The specification owner MUST maintain one closed `cartulary.extension_clause_traceability.v1` object containing exactly `schema_id`, `extensions_document_sha256`, and `clauses[]`. `schema_id` equals `cartulary.extension_clause_traceability.v1`, the document digest equals `owner_document_sha256_v1` for this NLSpec, and `clauses[]` contains `1..65536` rows.

Each clause row contains exactly `clause_id`, `document_ordinal`, `parent_anchor_kind`, `parent_anchor_id`, `clause_kind`, `clause_ordinal`, `source_start_byte`, `source_end_byte`, `clause_text_sha256`, `requirement_ids[]`, `acceptance_criterion_ids[]`, and `verification_ids[]`. `parent_anchor_kind` MUST equal `document`, `h1`, or `requirement`. `clause_kind` MUST equal `frontmatter_member`, `normative_table_caption`, `normative_table_row`, `list_item`, `fenced_literal`, `prose_block`, or `acceptance_row`. All ordinals are zero-based JSON integers. `document_ordinal` MUST be contiguous from `0`; `clause_ordinal` MUST be contiguous from `0` within one exact `(parent_anchor_kind,parent_anchor_id)` scope. Source ranges are zero-based byte offsets into the exact digested source and use half-open `[source_start_byte,source_end_byte)` form with `0 <= start < end <= document_byte_length`. Ranges MUST equal the exact hashed clause bytes, MUST NOT overlap, and MUST remain within the extraction region. `verification_ids[]` contains unique current Harness v2 IDs sorted by ascending ASCII bytes; every ID MUST resolve exactly once and MUST share at least one requirement or acceptance mapping with the clause.

The specification owner MUST author one closed `cartulary.extension_traceability_mapping_source.v1` object containing exactly `schema_id`, `extensions_document_sha256`, and `mappings[]`. Each mapping contains exactly `source_start_byte`, `source_end_byte`, `parent_anchor_kind`, `parent_anchor_id`, `clause_kind`, `requirement_ids[]`, `acceptance_criterion_ids[]`, and `verification_ids[]`. It MUST bind this document's exact digest and map every extracted clause exactly once. A mapping to another document, stale digest, overlap, gap, out-of-bounds range, wrong parent scope, wrong kind, or association with clause bytes not at the declared range is invalid. Heading heuristics, text search, and manually assigning generated clause IDs are forbidden.

Before extraction, exact source bytes MUST pass EXT-REQ-229. `extract_extension_clauses_v1` MUST parse from the opening front-matter delimiter through the LF immediately preceding Appendix A and MUST:

1. bind the valid linter result to the exact document digest;
2. emit each non-delimiter front-matter member as one document-parent clause;
3. reset the parent at each H1 and select a requirement parent after an exact requirement marker;
4. ignore headings, blank lines, requirement markers, `Profiles:` lines, `Verified by:` lines, table headers, and table delimiters;
5. emit each fenced block as one `fenced_literal`;
6. emit each normative table caption and data row separately;
7. emit each one-line list item separately;
8. emit every remaining maximal contiguous nonblank block under the same parent as one `prose_block`;
9. classify Table 28-A data rows as `acceptance_row`;
10. hash exact raw clause bytes without prefix removal, whitespace normalization, Unicode normalization, line reflow, or Markdown rendering.

`clause_id` MUST equal `extcl:` plus the first 32 lowercase hexadecimal characters of SHA-256 over the canonical JSON object containing exactly `extensions_document_sha256`, `document_ordinal`, `parent_anchor_kind`, `parent_anchor_id`, `clause_kind`, `clause_ordinal`, `source_start_byte`, `source_end_byte`, and `clause_text_sha256`, serialized under `extension_registry_canonical_json_v1` without the final LF. Rows sort by contiguous `document_ordinal`. Requirement and acceptance arrays sort by numeric suffix; verification IDs sort by ASCII bytes.

Any unmapped clause, orphan criterion or verification, duplicate identity, noncontiguous ordinal, stale digest, unresolved parent, inactive verification, or linter/source mismatch MUST fail adoption accounting. This artifact is documentation-generation and drift input only. It MUST NOT be loaded by catalog selection, generate a verification contract or row, infer an owner, create a runner selector, or serve as retained execution evidence. Documentation tooling MAY read the exact source under the Testing Harness documentation-read exception; product and catalog executors MUST NOT.

Profiles: base
Verified by: EXT-AC-097, EXT-AC-123, EXT-AC-124, EXT-AC-149

**EXT-REQ-228**
Every active catalog row used for this NLSpec MUST validate as `cartulary.test_family_manifest.v1` and contain one owner, nonempty verification IDs, an allowlisted runner, an exact selector, one evidence class, immutable runtime/resource/fixture profile IDs, `default_check`, claim posture, and `status='active'`. Ownership MUST follow the Harness v2 normative-postcondition algorithm. Cross-owner participants are collaborators and MUST NOT create duplicate ownership.

Catalog rows MUST NOT embed commands, ports, service topology, environment overrides, fixture paths, document-derived behavior, requirement text, acceptance text, or clause bytes. Optional documentation references are inert and MUST NOT be opened, hashed, parsed, or used for selection. Go, Vitest, Playwright, and shell evidence MUST use their exact v2 selector forms; zero resolution, multiple resolution, selector overlap, globbing, regular expressions, or unmapped executed cases MUST fail before a passing owner result is emitted.

The unclaimed Base runtime and each claimed-profile runtime are immutable startup identities. Runtime reset MUST NOT toggle extension claims, mutate the key-ring identity, or translate between profiles. Current Network Flow browser evidence uses the adopted `network_flow_claimed` profile; ordinary unclaimed evidence uses `default` or `none` according to service needs. A new public command, runner, schema family, or execution-profile contract requires a Testing Harness NLSpec amendment before use. Adding rows or selectors within existing v2 contracts requires only the authored verification, owner, catalog, and topology updates plus their generated projections.

Profiles: base
Verified by: EXT-AC-123, EXT-AC-140

**EXT-REQ-236**
Every independently observable acceptance branch, boundary value, enum token, matrix row, failure-injection position, and semantic condition required by this NLSpec MUST resolve to exact executable selector inventory in at least one active catalog row. A row MAY select multiple exact cases only when the runner adapter accounts for every registered symbol, title, or scenario and fails on any missing, duplicate, skipped-without-authorization, or unexpected observation. A family name, filename, title prefix, direct package script, raw target result, or aggregate process exit MUST NOT close row evidence.

A current passing execution obligation requires:

1. the exact active verification and catalog semantic digests;
2. a compatible successful full-owner `test-slice` partition for every affected owner;
3. every evidence-class gate required by the verification contracts;
4. paired `cartulary.test_evidence_accounting.v1` and `cartulary.test_owner_summary.v1` shards for each selected owner/target partition;
5. exactly one successful terminal record for every required row;
6. one `cartulary.test_evidence_audit_summary.v1` that accepts the exact supplied roots and closes every required owner/target/row partition.

An explicit selected-subset slice is focused execution evidence only and MUST NOT claim full-owner closure. A broad passing `check` root does not prove a separately required browser, accessibility, visual, measurement, static, security, or release target ran. Pre-v2 artifacts, newest-run discovery, mixed semantic digests, stale roots, support-only rows, informative rows, compatibility translations, and absent explicit roots MUST NOT close extension adoption. Core 05 claim publication remains separately gated and is never implied by implementation or release-readiness evidence.

Profiles: base
Verified by: EXT-AC-123, EXT-AC-140

**EXT-REQ-229**
`lint_extension_normative_source_v1` MUST accept a normative source document if and only if all of these conditions hold:

1. bytes are valid UTF-8, contain no BOM, CR, NUL, or tab byte, and end with exactly one LF;
2. headings use ATX syntax only, with one ASCII space after the marker; Setext headings are forbidden;
3. outside code fences, heading depth may remain the same or decrease freely, but an increase is exactly one level;
4. fenced code blocks use exactly three backticks at column 1 and terminate;
5. indented code blocks and raw HTML blocks are forbidden;
6. pipe tables use leading and trailing pipes, one header and delimiter row, and one column count;
7. a literal pipe in a table cell is escaped exactly as `\|`;
8. list markers begin at column 1; nested markers and multiline continuations are forbidden;
9. requirement markers use exact `**EXT-REQ-NNN**` form, acceptance IDs use `EXT-AC-NNN`, and duplicate IDs are invalid;
10. Table 28-A is continuous with unique acceptance IDs from `EXT-AC-001` through `EXT-AC-158`;
11. requirement-looking and acceptance-looking text inside code fences is ignored structurally.

The result MUST bind the exact source digest and contain either `status='valid'` with no findings or `status='invalid'` with `1..4096` findings under Section 26 ordering, formatting, secrecy, and overflow rules. The linter MUST run before owner-anchor validation, clause extraction, closure derivation, or static adoption accounting.

Golden vectors MUST cover valid and invalid heading transitions, split tables, escaped pipes, fenced pipes, nested lists, final LF, duplicate and skipped acceptance IDs, unequal columns, tabs, CRLF, BOM, raw HTML, Setext headings, four-backtick fences, and unterminated fences. Linting and clause extraction are support/documentation mechanics. Their passing result MUST NOT become a product row result or authorize catalog derivation from prose.

Profiles: base
Verified by: EXT-AC-124, EXT-AC-141

**EXT-REQ-230**
Unless a narrower contract applies, every canonical extension JSON artifact MUST enforce Table 27-D before emitting canonical bytes or a digest.

**Table 27-D. Shared canonical artifact limits**

| Property | Maximum |
| --- | ---: |
| Nesting depth | `32` |
| Members in one object | `4096` |
| UTF-8 bytes in one string | `1048576` |
| Items in one array | `65536` |
| Canonical bytes in one artifact | `67108864` |

The clause-traceability artifact has a narrower maximum nesting depth of `16`, `65536` clause rows, and `67108864` canonical bytes. A participant payload contract MUST declare its own maximum canonical bytes and MUST NOT exceed `67108864`; omission is invalid.

For every shared or narrower limit, the first value outside the admitted domain MUST fail deterministically with the owning `limit_exceeded` condition. Truncation, partial canonical output, partial digest output, silent item dropping, saturating byte counts, and continued scanning solely to discover a larger overage are forbidden. The implementation MUST prove item count and aggregate byte ceiling before publishing ordered output.

Profiles: base
Verified by: EXT-AC-082, EXT-AC-125, EXT-AC-127

**EXT-REQ-167**
A required owner fragment, dependency snapshot, descriptor, binding, state-presence declaration, initialization definition, migration definition, backup codec, owner reference, conformance manifest, generated schema, verification contract, owner-catalog row, execution-profile reference, registry-accounting artifact, clause-traceability artifact, or adoption gate MUST NOT contain a `TODO` placeholder when this document is promoted to `adopted/current`.

Profiles: base
Verified by: EXT-AC-075, EXT-AC-128

# 28. Acceptance criteria

The implementation and coordinated document set are conformant only when every criterion below passes. When a criterion names an absent optional condition, it passes only by proving the criterion's specified omission behavior. A broad criterion does not satisfy an independently observable clause unless clause traceability maps that clause to an active Harness v2 verification ID and exact executed selector inventory.

**Table 28-A. Acceptance criteria**

| ID | Binary criterion |
| --- | --- |
| `EXT-AC-001` | The adopted document has one owner for every behavior family and contains no unresolved contradiction or open delegation in normative text. |
| `EXT-AC-002` | Every `EXT-REQ-*` maps to at least one `EXT-AC-*`, every criterion maps to one or more active Harness v2 verification IDs, every cited verification maps back, and no identifier is reused. |
| `EXT-AC-003` | Core 00, the canonical registry, and public discovery contain the same recognized profile IDs and current contract majors. |
| `EXT-AC-004` | The same owner inputs produce byte-identical descriptors and canonical registry bytes under `extension_registry_canonical_json_v1`. |
| `EXT-AC-005` | A missing, duplicate, extra, malformed, stale, or unrecognized descriptor fails before listeners or workers start. |
| `EXT-AC-006` | Every claimable profile has a non-null owner reference, contract major, descriptor, and conformance manifest. |
| `EXT-AC-007` | Omitted descriptor and configuration defaults materialize before validation and canonical serialization. |
| `EXT-AC-008` | Unknown descriptor, registry, dependency, state-variant, or contribution members are rejected. |
| `EXT-AC-009` | A recognized unclaimable profile remains discoverable and rejects a true claim request before startup. |
| `EXT-AC-010` | A missing, duplicate, unrecognized, or major-mismatched implementation binding fails before profile migration or contribution registration. |
| `EXT-AC-011` | Every omitted claim key materializes as `false`; explicit `false` produces the same resolved unclaimed state; true is only a request until admission succeeds. |
| `EXT-AC-012` | Explicit null, string, integer, array, or object claim values fail; unclaimed profile-local config is either forbidden or syntax-only under its exact configuration-contract row. |
| `EXT-AC-013` | Any requested-claim failure starts no HTTP listener, WebSocket listener, or background-job runner and yields non-zero exit. |
| `EXT-AC-014` | A requested unclaimable profile fails with `extension_profile_not_claimable`. |
| `EXT-AC-015` | Public compatibility is determined by contract major and capabilities, not route version, document version, or package version. |
| `EXT-AC-016` | `network_flow_activity.claimed=true` with `import.claimed=false` fails with `extension_dependency_not_claimed`. |
| `EXT-AC-017` | Import is never auto-claimed and current unclaimed profiles emit empty capability arrays. |
| `EXT-AC-018` | A dependency contract-major mismatch fails with `extension_dependency_incompatible`, and an implementation-binding mismatch fails with `extension_implementation_unavailable`, before profile migration. |
| `EXT-AC-019` | Self-dependencies, duplicate dependency rows, and cycles fail deterministically before profile-local validation. |
| `EXT-AC-020` | Claim resolution processes dependencies before dependents, breaks independent ties by profile ID, and starts listeners only after complete success. |
| `EXT-AC-021` | Every collision class in Table 13-A fails deterministically before claim-state resolution. |
| `EXT-AC-022` | Equal, parameter-compatible, ancestor, or descendant route roots collide; incompatible literal roots do not. |
| `EXT-AC-023` | An extension route that captures a Base route fails, and unknown future extension identities do not fail compatible Base clients. |
| `EXT-AC-024` | Discovery returns every recognized profile in profile-ID order with accurate claimable, claimed, major, reserved roots, workspace keys, and capabilities. |
| `EXT-AC-025` | Every current discovery item emits the exact generic v1 members and canonical nested-array ordering. |
| `EXT-AC-026` | Public discovery emits no `document_version`, singular `route_root`, implementation version, state version, or secret-bearing metadata. |
| `EXT-AC-027` | Reserved route families remain listed while unclaimed. |
| `EXT-AC-028` | Declared workspace keys remain listed while unclaimed. |
| `EXT-AC-029` | Unclaimed profiles emit `capabilities=[]`; claimed profiles emit only capabilities declared by the owner descriptor and admitted under EXT-REQ-077. |
| `EXT-AC-030` | Discovery contains no provider maps, live resource counts, incident data, configuration values, or storage details. |
| `EXT-AC-031` | Reserved-unclaimed dispatch occurs before family authorization, incident lookup, request validation, resource lookup, idempotency, cursor handling, or job admission. |
| `EXT-AC-032` | Claimed-family authorization denial or zero-resource state never returns `extension_profile_not_claimed`. |
| `EXT-AC-033` | Network Flow local error precedence begins only after Core claimed-family dispatch. |
| `EXT-AC-034` | An extension workspace renders only when claimed, declared, and currently authorized. |
| `EXT-AC-035` | Extension claims do not alter the five Base tabs or promote workspaces into the Base view registry. |
| `EXT-AC-036` | Base workbook startup and Base surface switching do not wait for extension resources or external services. |
| `EXT-AC-037` | Extension resource loading begins only after workspace selection or an explicit extension action. |
| `EXT-AC-038` | No current public or operator surface changes claim state or installs, removes, reloads, or uploads executable extensions. |
| `EXT-AC-039` | Authorization loss clears stale extension state, and unknown future extension values do not fail the Base workbook or WebSocket. |
| `EXT-AC-040` | Profile, workspace, capability, contribution, and resource identity never depends on visible label, route, row position, or component name. |
| `EXT-AC-041` | Unclaim removes routes, workspaces, actions, and workers without deleting, rewriting, downgrading, or reclassifying authoritative state. |
| `EXT-AC-042` | Unclaimed profiles start no workers, perform no profile-local external probe, and run no extension-local semantic migration. |
| `EXT-AC-043` | Reclaiming compatible state restores the same resource IDs and does not duplicate imports, bindings, jobs, or audit occurrences. |
| `EXT-AC-044` | Stored state newer than the packaged implementation fails claim admission before mutation. |
| `EXT-AC-045` | Stored state below the minimum migratable version fails claim admission before mutation. |
| `EXT-AC-046` | A missing consecutive migration fails claim admission. |
| `EXT-AC-047` | A committed migration step is not reapplied after restart. |
| `EXT-AC-048` | An interrupted uncommitted migration can be retried without duplicating committed effects. |
| `EXT-AC-049` | Production downgrade never uses a down migration to reinterpret newer state. |
| `EXT-AC-050` | Retirement is represented by `claimable=false`, preserves state and reserved identity, and exposes no invented read-only mode. |
| `EXT-AC-051` | Every extension-owned durable job has internal owner profile identity. |
| `EXT-AC-052` | A pre-commit nonterminal job becomes canceled or failed during unclaim reconciliation without publishing a resource. |
| `EXT-AC-053` | A post-commit job recovers to exactly one original success without duplicate resource, result, or audit effects. |
| `EXT-AC-054` | A runtime extension failure does not silently change the resolved claim set or substitute another extension. |
| `EXT-AC-055` | Import and other shared owners never write extension-owned authoritative tables directly. |
| `EXT-AC-056` | A failed or canceled pre-commit cross-owner operation leaves no queryable extension resource and no terminal success. |
| `EXT-AC-057` | A crash after final commit recovers one resource and one terminal success. |
| `EXT-AC-058` | Applicable owner, idempotency, job, audit, indicator, history, and invalidation participants commit atomically. |
| `EXT-AC-059` | An extension resource is not assigned Core record, view-schema, saved-view, or Base surface identity without a Core amendment. |
| `EXT-AC-060` | An extension cannot create or mutate Core indicator observations or other Core state except through an admitted typed participant. |
| `EXT-AC-061` | Physical backup preserves claimed and unclaimed authoritative extension state and state-version metadata. |
| `EXT-AC-062` | Restore preserves unclaimed state without interpreting or migrating it. |
| `EXT-AC-063` | A claimed profile becomes available after restore only after compatibility, migration, and post-restore validation. |
| `EXT-AC-064` | Incident-bundle export fails before publication when Network Flow blocking state exists. |
| `EXT-AC-065` | A profile with no Snapshot and Reporting participant contributes no snapshot, report, diagram, or release content implicitly. |
| `EXT-AC-066` | Unknown extension payloads and uploaded content remain inert and are not executed or interpreted as extension code. |
| `EXT-AC-067` | A profile with `egress_mode='none'` performs no third-party request from routes, workers, migrations, validation, or browser code. |
| `EXT-AC-068` | An owner-declared egress profile satisfies every Table 24-A contract and fails closed when one dimension is invalid. |
| `EXT-AC-069` | Secrets and prohibited sensitive values do not appear in descriptors, discovery, errors, audit, telemetry, jobs, readiness, browser state, or WebSocket payloads. |
| `EXT-AC-070` | No runtime package installation, marketplace, arbitrary callback bus, arbitrary UI injection, separate extension host, or extension microservice is present. |
| `EXT-AC-071` | OpenTelemetry claimed-profile identity derives from the canonical resolved claim set. |
| `EXT-AC-072` | Drift among Core 00, Core 01, Core 03, Core 04, profile owners, descriptors, discovery evidence, telemetry, v2 verification contracts, and owner-catalog routing fails the owning canonical gate. |
| `EXT-AC-073` | Every claimable profile has complete owner, dependency, contribution, state, portability, reporting, egress, conformance-manifest, verification-contract, and exact-selector coverage. |
| `EXT-AC-074` | Base Profile conformance passes with every optional profile unclaimed and with no dynamic-package behavior. |
| `EXT-AC-075` | No required owner reference, descriptor, dependency, migration, schema, conformance manifest, verification contract, catalog row, traceability object, or adoption gate contains a `TODO` placeholder at adoption. |
| `EXT-AC-076` | Every descriptor fact derives from a digest-bound adopted owner fragment; arbitrary prose, implementation, route, and database extraction cannot create a fact. |
| `EXT-AC-077` | Every owner locator resolves exactly once against the declared owner-document digest; absolute, traversal, backslash, symlink, stale, zero-match, and multi-match cases fail. |
| `EXT-AC-078` | Descriptor-source omissions materialize exactly as specified, explicit `null` never invokes a default, and every canonical descriptor member is present. |
| `EXT-AC-079` | Generation and build verify every declared generator-source path and byte digest, while runtime accepts only the exact packaged canonical artifact identity and digest sets plus the embedded root digest and does not require source bytes. |
| `EXT-AC-080` | Binding profile, major, descriptor digest, contributions, capabilities, admission, initialization, state, migration, codec, rebuild, transaction-limit, schema, worker, job, and participant declarations satisfy exact parity. |
| `EXT-AC-081` | Every claimable descriptor manifest ID resolves to exactly one matching manifest; static registry-accounting pass/fail is biconditional; an unclaimable null manifest has no entry; overflow emits no conformant truncated accounting object. |
| `EXT-AC-082` | Every static and configured limit accepts both valid boundaries and rejects the first value outside the domain with the exact reason and no partial admission. |
| `EXT-AC-083` | Missing dependency edges are reported before graph ordering; each maximal multi-profile strongly connected component produces one canonical cycle finding; duplicate and pair findings follow the closed multiplicity table. |
| `EXT-AC-084` | Startup finding paths, exact generic messages, closed detail objects, ordering, profile-local replacement rules, and overflow behavior match §26. |
| `EXT-AC-085` | Every requested profile completes side-effect-free preflight before the first migration; a later failure publishes no routes or workers, while earlier committed migration steps remain resumable. |
| `EXT-AC-086` | State presence, metadata consistency, lineage, migration lock, immutable definition digest, ledger, step timeout, profile timeout, and indeterminate-commit handling produce their exact outcomes. |
| `EXT-AC-087` | Unclaimed and recognized-unclaimable profiles satisfy every Table 21-B row; syntax-only values remain inert, local availability checks never run, and generic job reconciliation remains shared-owner behavior. |
| `EXT-AC-088` | Valid commit proof takes precedence over cancellation, the exact stored original success is replayed once, absent proof fails or cancels as specified, and contradictory proof blocks startup. |
| `EXT-AC-089` | Authoritative non-database bytes are durable before database publication, become queryable exactly at final database commit, reject different-byte replay, and orphan within the required interval. |
| `EXT-AC-090` | Every closed fatal condition imports Core 04 `fatal_integrity_shutdown_v1`, preserves committed state and durable jobs, is idempotent under repeated signals, emits no unadmitted WebSocket event, and exits `70`; ordinary admission failure exits `2`. |
| `EXT-AC-091` | Current discovery producers emit no extra member; compatible consumers ignore additive unknown members but reject malformed known members and never execute unknown data. |
| `EXT-AC-092` | Discovery retains the existing three Core fields, adds exactly the four generic fields, and no Network Flow profile-local discovery item remains normative or emitted. |
| `EXT-AC-093` | Every compatibility-matrix row causes the required document, contract, state, schema, or algorithm action; unsupported majors and unknown values follow their exact client outcomes. |
| `EXT-AC-094` | Every portability and Snapshot/Reporting state-presence and claim-state matrix combination produces the specified include, omit, empty, or fail result. |
| `EXT-AC-095` | Every claimable profile resolves every generated closure-catalog item exactly once with current owner locators or one item-permitted closed not-applicable reason. |
| `EXT-AC-096` | Every transport, session, authorization, closure, unclaim, retirement, unsupported-major, and resource-removal condition produces the exact cache, request, queue, optimistic-state, and draft consequence. |
| `EXT-AC-097` | Every independently observable normative clause maps to an acceptance criterion and active Harness v2 verification ID, and every cited verification maps back to current digest-bound clauses. |
| `EXT-AC-098` | Every owner locator resolves through one digest-bound owner contract manifest; anchor ranges and adopted fragment IDs, paths, and digests match exactly, and Markdown search cannot add an anchor or fragment. |
| `EXT-AC-099` | Every owner fact kind produces exactly one declared canonical identity object, and the same identity bytes drive ordering, duplicate rejection, collision reporting, and identity digests. |
| `EXT-AC-100` | Zero recognized profiles, with the required dependency manifests retained but zero adopted owner fragments and zero normalized facts, produce the exact canonical empty-profile owner input, registry, integrity, discovery, and accounting states. |
| `EXT-AC-101` | Descriptor-source instances remain ephemeral and are never persisted, hashed, packaged, logged, drift-checked, or consumed at runtime. |
| `EXT-AC-102` | Every profile-local configuration key has a value schema, omission policy, inactive policy, resolution kind, diagnostic policy, and bound; claimed normalization produces the exact configuration view. |
| `EXT-AC-103` | Inactive and retired configuration performs only JSON-shape, scalar/reference-grammar, byte, and depth checks and never performs resolution, existence checks, DNS, connection validation, egress, or profile code. |
| `EXT-AC-104` | Exactly one application process owns the deployment serving lease; a concurrent process performs no mutation or listener start, and confirmed lease loss enters fatal shutdown. |
| `EXT-AC-105` | Stage 6 exposes no route, workspace, job dequeue, WebSocket subscription, or readiness success before every mandatory component reaches the atomic `serving` transition. |
| `EXT-AC-106` | Every declared timeout uses the exact monotonic start/end points, included queue/lock work, cancellation checkpoint, grace, late-result discard, rollback/lock behavior, and final-commit outcome classification. |
| `EXT-AC-107` | Every Base public path namespace appears exactly once in the canonical Base route-reservation registry, and exact/descendant overlap with every extension route family is deterministic. |
| `EXT-AC-108` | The packaged client support registry advertises only digest-bound current profile majors, workspace keys, capabilities, and public schemas present in the canonical extension registry. |
| `EXT-AC-109` | The no-store workbook-startup member and local epoch/generation tuple gate the exact discovery/support/authorization intersection; stale or defective extension responses cannot render or alter Base behavior. |
| `EXT-AC-110` | Network Flow discovery uses only the generic seven-member discovery item under explicitly adopted contract major 2 unless the adopted never-effective-defect exception proves every required condition. |
| `EXT-AC-111` | Cross-owner participants satisfy exact input, result, key, finding, deadline, ordering, cancellation, conflict, timeout, replay, and final-commit outcomes with no automatic retry or partial result. |
| `EXT-AC-112` | At expiry, unpublished staged bytes fail before storage access independently of cleanup; cutoff batches and every physical-deletion outcome follow the closed table, and integrity contradictions are fatal. |
| `EXT-AC-113` | Every migration receives only its closed scoped context and returns only the closed apply and validation result variants; undeclared state and cross-owner access are impossible through the interface. |
| `EXT-AC-114` | Fresh initialization commits only after one final validation; current, fully migrated, and restored paths each invoke final validation exactly once, while migration steps use only pending-state validation and remain resumable after later failure. |
| `EXT-AC-115` | Proof requiredness, proof prohibition, idempotency identity, terminal result, resource-reference contract, proof bounds, and cancellation behavior derive from the job-kind contract alone. |
| `EXT-AC-116` | Every authoritative and derived family has exact presence/binding classification, codec, backup inclusion, restore ordering, validator, and rebuild parity; filesystem is derived-only and inactive restore invokes no profile code. |
| `EXT-AC-117` | Participant operations have canonical order and unique kinds; closed contexts/results, scoped access, limits, logical references, findings, and every shared-owner/profile-code invocation row are exact. |
| `EXT-AC-118` | Inactive-state portability blockage is evaluated only by the declarative shared predicate over declared logical state families; inactive and retired profile code never executes. |
| `EXT-AC-119` | Every reachable invalid condition, including transaction timeout, unsupported backup codec, and cleanup dependency failure, maps to exactly one phase, path algorithm, reason, formatter set, multiplicity, secret policy, and owner locator. |
| `EXT-AC-120` | Validation-result precedence distinguishes invalid bytes, invalid shape, 4097-item overflow, remaining schema defects, and valid ordinary findings exactly, without a truncated result. |
| `EXT-AC-121` | The closure catalog contains every baseline, owner requirement, configuration, schema, contribution, job, migration, initialization, state-family, backup-codec, and verification-obligation item, and no profile owner can reduce it. |
| `EXT-AC-122` | Every accounting status is produced by a named predicate over exact current input digests; all required check instances are present once, and no manual Boolean or broad target result can produce `pass`. |
| `EXT-AC-123` | Every independently observable acceptance branch resolves to exact active v2 selector inventory and one compatible successful terminal row record; family names, title prefixes, raw scripts, and aggregate exits cannot close accounting. |
| `EXT-AC-124` | The normative-source linter accepts only the declared Markdown subset and heading transitions; Table 28-A is continuous with unique contiguous IDs, and all golden vectors classify exactly. |
| `EXT-AC-125` | Every canonical extension artifact, including state-presence, initialization, codec, registry-accounting, and clause-traceability artifacts, enforces total bytes, nesting, members, strings, arrays, and first-overflow without partial output or digest. |
| `EXT-AC-126` | Every public error token owned outside the profile resolves to the exact owner contract for status, code, reason vocabulary, retryability, and closed safe details. |
| `EXT-AC-127` | Registry integrity and runtime admission cover every owner manifest and required static supporting artifact with exact identity/digest parity while generator source bytes remain build-only provenance. |
| `EXT-AC-128` | No current-scope owner input, closure item, verification contract, catalog row, generated artifact, accounting predicate, acceptance mapping, or adoption gate contains an unresolved `TODO` placeholder or open delegation phrase. |
| `EXT-AC-129` | Inactive syntax-only configuration never resolves, probes, validates existence of, or activates an external or local resource. |
| `EXT-AC-130` | Transaction input, result, finding, key, and deadline bounds; cancellation checkpoints; and commit-boundary outcomes are exact. |
| `EXT-AC-131` | Staged bytes become inaccessible independently of cleanup, and every storage outcome follows Table 20-B. |
| `EXT-AC-132` | Authoritative/derived families, filesystem restriction, state-presence bytes and digest, and implementation-binding parity are exact. |
| `EXT-AC-133` | Fresh, already-current, fully migrated, and restored paths invoke initialization and final validation exactly as specified. |
| `EXT-AC-134` | Participant operation ordering and every state/claim invocation-matrix row are exact. |
| `EXT-AC-135` | Backup codec framing, bounds, ordering, empty state, digest, historical-codec, and unsupported-codec behavior are exact. |
| `EXT-AC-136` | Build verifies generator-source bytes while runtime needs only packaged artifacts and the embedded root digest. |
| `EXT-AC-137` | Fatal conditions during startup and serving use the imported Core lifecycle and the correct idempotent exit-code behavior. |
| `EXT-AC-138` | Every new or reused scalar satisfies its exact grammar, bound, secrecy, derivation, owner-import, and replay rule. |
| `EXT-AC-139` | Workbook startup carries no-store availability; stale responses cannot render; `client_instance_id` and every Base/transport identity remain stable through epoch rollover. |
| `EXT-AC-140` | Every affected owner has compatible successful full-owner v2 evidence, all evidence-class gates pass, and one evidence audit closes every required owner/target/row partition without subset, broad-target, stale-root, or historical fallback. |
| `EXT-AC-141` | The exact document passes repository Markdown lint and the updated normative-source linter with acceptance IDs through `EXT-AC-158`. |
| `EXT-AC-142` | Empty-state policy is required and exact; metadata never establishes state presence; every metadata/state/policy combination and empty initialization obeys `allowed` or `forbidden`, with Network Flow selecting `allowed`. |
| `EXT-AC-143` | Validation results apply invocation, structural, overflow, remaining-schema, valid-findings, and valid-empty precedence exactly for counts `0`, `256`, `257`, `4096`, and `4097`. |
| `EXT-AC-144` | Portability export and import use distinct result schemas; import preparation is side-effect free, uses only scoped staged output and the shared final transaction, and rejects per-participant or aggregate byte `67108865`. |
| `EXT-AC-145` | The authored dependency declaration set requires present non-null arrays and exact manifest versions/digests; empty, omitted, null, duplicate, extra, stale, and mismatched declarations classify exactly, with no phase-shaped reader. |
| `EXT-AC-146` | Every descriptor scalar and set member has one exact owner source through `primary_owner_contract_ref`; zero or multiple scalar sources, duplicate sets, stale refs, and code/prose inference fail generation. |
| `EXT-AC-147` | Every schema constraint and procedural decision maps exactly once through a complete validation-surface declaration, and missing, duplicate, stale, extra, incomplete, or emitted-unregistered conditions fail closed. |
| `EXT-AC-148` | Every subject and contribution kind receives exactly its exhaustive closure categories; generated rows reject omission and owner-authored not-applicable, while fixed baseline rows admit only their enumerated reasons. |
| `EXT-AC-149` | Traceability enforces closed parent/clause kinds, zero-based scoped ordinals, exact half-open byte ranges, document digest/mapping, clause-ID inputs, and rejection of overlap, bounds, scope, digest, and document mismatches. |
| `EXT-AC-150` | `syntax_only` requires an inert non-null schema and never applies omission/default semantics, creates a view, retains a value, resolves a resource or secret, performs egress, or invokes profile code. |
| `EXT-AC-151` | The application lease follows every declared acquisition, held, uncertainty, recovery, loss, release, crash, session-identity, deadline, and no-reacquisition branch with startup exit `2` and confirmed-loss exit `70`. |
| `EXT-AC-152` | Deadline calculation checks and saturates arithmetic, selects the inherited minimum with deterministic equal ties, expires at `now >= deadline`, supports zero grace, and applies cancellation/timeout/proven-commit precedence exactly. |
| `EXT-AC-153` | Transaction participant counts `0`, `1`, `16384`, and `16385`, every 64 MiB boundary, first-invalid order, result validity, and cancellation checkpoint execute exactly with no partial effects. |
| `EXT-AC-154` | Every staged-object default, expiry/retry predicate, upload abandonment, deletion outcome/order, transaction-free delete, serialized sweep, missed-interval coalescing, startup dependency, and readiness branch executes exactly. |
| `EXT-AC-155` | Restore accepts only stopped empty targets, processes numeric groups and sequential bindings with validation before advance, never serves failure, invokes no inactive code, uses only digest-bound codecs, and rebuilds only after claim. |
| `EXT-AC-156` | The standard client registry covers every claimable workspace profile including Network Flow major `2`/`network_analysis`, and linearized reservation, stale/current generation, concurrency, rollover, authorization loss, and Base identity preservation are exact. |
| `EXT-AC-157` | Every capability surface emits `[]`; nonempty facts, arrays, and activation attempts are rejected with `extension_capability_not_supported` and execute no behavior. |
| `EXT-AC-158` | Loss of each mandatory listener, dequeue gate, or worker before bind, before serving, or while serving is distinguished from operation failure, drains and preserves state, exits `70`, and never restarts in-process. |

# 29. Coordinated adoption gates and required companion amendments

**EXT-REQ-168**
This NLSpec MUST remain `status: draft` until every gate in Table 29-A is closed in one coordinated specification change or an ordered change series that leaves no document claiming the new subsystem as adopted before its dependencies are adopted.

Profiles: base
Verified by: EXT-AC-001, EXT-AC-072, EXT-AC-075, EXT-AC-128

**Table 29-A. Adoption gates**

| Gate ID | Required closure |
| --- | --- |
| `EXT-GATE-001` | Core 00 adopts this NLSpec for shared extension mechanics, associates every adopted owner document with one digest-bound owner contract manifest, sets every current profile major, sets `network_flow_activity@2`, and records `network_flow_activity -> import@1`. |
| `EXT-GATE-002` | Core 01 adopts the strict seven-member discovery producer, tolerant decoder, Base route reservations, exact workbook-startup availability member and no-store behavior, reserved inactive routes, route overlap, and public dispatch precedence. |
| `EXT-GATE-003` | Core 01 adopts the bounded transaction protocol and errors, staged-object access/cleanup and errors, backup codec selection/error, job proof/reconciliation, typed portability/reporting/backup participation, final-commit boundary, and recovery. |
| `EXT-GATE-004` | Core 02 adopts or confirms the generic extension-resource boundary, authoritative/derived logical state-family ownership boundary, state-presence exclusion rules, and cross-owner authoritative-write prohibition. |
| `EXT-GATE-005` | Core 03 adopts availability epoch/generation, stable `client_instance_id` and WebSocket identity, exact discovery/support/authorization intersection, lazy loading, unsupported-major behavior, Base fallback, authorization-loss disposal, and Base cache/request/queue/draft preservation. |
| `EXT-GATE-006` | Core 04 adopts forbidden/syntax-only inactive processing, every timeout, process lease, single-active-process rule, Stage 6 publication, readiness including cleanup dependency degradation, `fatal_integrity_shutdown_v1`, and exit codes `2` and `70`. |
| `EXT-GATE-007` | Network Flow Activity publishes its owner manifest and fragments, removes its competing discovery shape, adopts major 2 unless the complete exception is recorded, and declares Import dependency, empty initialization, migration/final validation, state presence/bindings/codecs, job kinds, participants, rebuilds, and portability blocking. |
| `EXT-GATE-008` | Reporting and Report Composition import the generic descriptor, claim, compatibility, state-presence, participant-context, result, and lifecycle contracts without transferring reporting or composition ownership. |
| `EXT-GATE-009` | Testing Harness v2 remains `adopted/current`; the verification registry contains both `module.extensions` contracts; the owner registry and family manifests contain every required active exact-selector row; runner and execution profiles validate; and no pre-v2 identity, custom extension-fixture result, compatibility reader, or historical run can close adoption. |
| `EXT-GATE-010` | OpenTelemetry derives `cartulary.profile.claims` only from the canonical resolved claim set and its published digest and records no profile-local secret or incident content. |
| `EXT-GATE-011` | `docs/domain.md` and implementation-support guides remove stale discovery, unclaimable, multi-process, migration, client-support, and owner-boundary language and add the adopted extension vocabulary without becoming behavior owners. |
| `EXT-GATE-012` | Every required extension artifact, static supporting contract, closure catalog, generated schema, v2 verification contract, owner-catalog row, execution-profile reference, and traceability object exists; every acceptance criterion passes; and no required placeholder or open delegation remains. |
| `EXT-GATE-013` | Every contributing owner document has one current owner contract manifest; its anchor ranges validate against the exact owner-document digest; and its adopted fragment ID/path/digest set matches exactly. |
| `EXT-GATE-014` | Dependency snapshot, owner manifests/input, descriptors, registry/integrity, bindings, state-presence/init/codec artifacts, schemas, static accounting, clause traceability, and embedded root digest generate and validate without drift. |
| `EXT-GATE-015` | Core 04 and the Extensions generator adopt every static/configurable limit, timeout, validation-condition row, new runtime reason, diagnostic formatter, exact message/details object, readiness state, lifecycle import, and exit code. |
| `EXT-GATE-016` | Every state-owning profile adopts state presence, initialization definition, physical binding, backup codecs, migration definitions/interfaces, exact-once final validation, rebuild algorithms, metadata, ledger, locks, and job-kind contracts. |
| `EXT-GATE-017` | Core 01 and Network Flow complete the additive generic discovery transition under the adopted major-version action, and no competing profile-local discovery item or inactive precedence remains normative or emitted. |
| `EXT-GATE-018` | Incident Portability, Reporting, Report Composition, and backup owners adopt the closed participant contexts/results/findings, logical-reference scalar, invocation matrix, declarative inactive predicate, physical codec boundary, and complete state/claim matrices. |
| `EXT-GATE-019` | Every claimable profile has a current generated closure catalog and its conformance manifest resolves every catalog item through exact owner locators or one item-permitted not-applicable reason. |
| `EXT-GATE-020` | Clause traceability maps every requirement and criterion bidirectionally to active v2 verification IDs; every independently observable branch has exact executable selector coverage; verification and catalog semantic digests are current; and static accounting passes. |
| `EXT-GATE-021` | The packaged browser assets include one digest-bound client support registry whose profile majors, workspaces, capabilities, and public schemas pass registry parity. |
| `EXT-GATE-022` | The canonical Base route-reservation registry covers every Base public path namespace exactly once and passes parity against packaged Base handlers without capturing an extension-owned namespace. |
| `EXT-GATE-023` | Exact v2 catalog selectors prove second-process denial and crash release, plus lease loss before bind, after bind before serving, and while serving under the imported fatal lifecycle. |
| `EXT-GATE-024` | Exact v2 catalog selectors prove empty/algorithm initialization, scoped access, pending-state validation, rollback, exact-once final validation, resumability, and fresh/current/migrated/restored outcomes. |
| `EXT-GATE-025` | Exact v2 catalog selectors inject failure or cancellation at every ordered transaction position and prove input/result/key limits, deadline, lock order, no retry, exact conflict/timeout/commit outcomes, and no partial effects. |
| `EXT-GATE-026` | Exact v2 catalog selectors prove multi-batch staged-object cutoff sweeping, access denial at expiry independently of cleanup, every deletion outcome, deterministic retry, no exposure, readiness degradation, and fatal contradictions. |
| `EXT-GATE-027` | Every affected owner has compatible successful full-owner row accounting and owner-summary shards, every required evidence-class gate passes, and one v2 evidence audit closes every required owner/target/row partition without subset, broad-target, stale-root, newest-run, or historical fallback. |
| `EXT-GATE-028` | Support-only documentation tooling accepts this exact normative source, Table 28-A remains continuous from `EXT-AC-001` through `EXT-AC-158`, aggregate-limit exact selectors prove no partial output or digest, and no linter result is promoted into product row evidence. |

**EXT-REQ-169**
Core 00 MUST preserve ownership of recognition and claimability after adoption. This NLSpec's current-profile parity tables and generated registry MUST be regenerated when Core 00 changes; they MUST NOT prevent a valid later profile addition or retirement performed through a new coordinated contract revision.

Profiles: base
Verified by: EXT-AC-003, EXT-AC-072

**EXT-REQ-170**
Core 00, Core 01, Core 03, Core 04, and the Network Flow Activity NLSpec MUST be amended together so that discovery has one generic seven-member shape, inactive dispatch has one precedence, client support uses one major, and the current Network Flow contract-major action is explicit.

The default coordinated action is Network Flow `contract_major=2` and document version `2.0.0` because removing public `document_version` and singular `route_root` changes a previously required response shape. A contract-major-1 correction is conformant only when the adopted document-status authority records every condition in EXT-REQ-231 as proven. A patch or minor document-version change without that proof is invalid.

A repository MUST NOT adopt this NLSpec while an older competing discovery item, inactive-route precedence, client major, owner fragment, conformance manifest, pre-v2 harness identity, stale catalog row, compatibility reader, or generated artifact remains current.

Profiles: base, network_flow_activity
Verified by: EXT-AC-025, EXT-AC-026, EXT-AC-027, EXT-AC-031, EXT-AC-033, EXT-AC-072, EXT-AC-092, EXT-AC-110

**EXT-REQ-171**
Appendices and guides MAY preserve rationale, examples, diagrams, operator procedures, physical path suggestions, implementation mechanisms, and future package research only. Omission behavior: no supporting appendix or guide is required for runtime conformance. They MUST NOT be the sole owner of a claim key, descriptor member, default, bound, timeout, algorithm, error, exit code, state, contribution, compatibility rule, cleanup consequence, or acceptance criterion.

Profiles: base
Verified by: EXT-AC-001, EXT-AC-072

The coordinated adoption order is mandatory:

1. revise this NLSpec and retain `status: draft`;
2. revise Core 00, Core 01, Core 03, and Core 04;
3. revise affected profile owners and state-owning migrations;
4. revise Incident Portability, Reporting, and Report Composition when this revision changes an interface imported from that owner; when no imported interface changes for one of those owners, record an exact no-change parity result in conformance accounting;
5. register the Harness v2 verification contracts, owner and family rows, exact selectors, and topology profiles; amend the Testing Harness NLSpec only if a new public command, runner, schema family, or execution-profile contract is required; revise OpenTelemetry;
6. generate every §27 extension artifact, clause-traceability object, and v2 projection;
7. execute every affected full-owner slice and required evidence-class gate;
8. audit the exact compatible retained roots and execute static registry, clause-level traceability, and drift accounting;
9. promote this NLSpec and all required companion revisions together.

An intermediate artifact MUST NOT claim the generic Extensions Subsystem is current while importing an older competing discovery, state, diagnostic, or lifecycle contract.

# 30. Future-only independently distributed extension packages

**EXT-REQ-172**
A future NLSpec that permits independently distributed extension packages MUST define at least:

- publisher identity and authorization;
- package manifest and canonical bytes;
- package signatures and threshold trust;
- trusted root rotation and revocation;
- rollback, freeze, mix-and-match, and wrong-package protection;
- dependency pinning and resolution;
- provenance, SBOM, and vulnerability response;
- package permission declarations;
- process, filesystem, database, browser, and network isolation;
- secret access and incident-data access permissions;
- installation, update, rollback, disable, uninstall, and retained-state behavior;
- package-supplied migration restrictions;
- compatibility, recovery, audit, and conformance fixtures.

Profiles: future-only
Verified by: EXT-AC-070, EXT-AC-074

**EXT-REQ-173**
Until that future NLSpec is adopted, no current input surface accepts an independently distributed executable extension package. Bytes presented to an existing upload or import surface MUST be processed only under that surface's inert-data contract and MUST be rejected when they violate that contract. The runtime MUST NOT recognize, install, execute, or reserve a compatible format for an executable extension package.

Profiles: base
Verified by: EXT-AC-066, EXT-AC-070

# Appendix A. Non-normative research rationale

This appendix is non-normative. It explains why the normative contract uses static packaging, typed registries, stable identity, lazy workspace loading, and strict owner boundaries.

R01 and R03 document the maintenance problems created by global controller state, schema-by-convention, rudimentary migration behavior, and unclear integration ownership. Those findings support generated descriptors, explicit state-version admission, and typed owner façades rather than feature discovery through implementation structure.[^1][^2]

R02 supports keeping coordination and review mechanisms adjacent to the workbook while avoiding mandatory challenge, checklist, or approval ritual on ordinary row editing. That evidence supports EXT-REQ-109 rather than a generic extension workflow hook.[^3]

R04 and R05 treat responsiveness as preservation of the perception-action loop and semantic continuity, not merely throughput. They support local packaged assets, lazy extension-resource loading, meaningful progress states, and insulation from transport and storage internals.[^4][^5]

R06 and R07 treat the spreadsheet-like workbook as a coordination surface over structured and auditable state, not as a SIEM, binary evidence repository, or unrestricted integration container. That supports the extension-resource versus Core-record boundary and explicit reporting and portability participation.[^6][^7]

R08 demonstrates the value of declarative plugin registration, dependency checks, explicit enablement, and lifecycle cleanup. It also demonstrates the coupling risk of a generalized hook substrate. The current contract adopts the declarative registration pattern but rejects arbitrary hooks.[^8]

R09 demonstrates the importance of stable owner-provided row identity, controlled state, explicit renderer boundaries, and browser/node/visual verification. Those findings support stable extension workspace and resource identifiers, closed contribution points, and conformance fixtures that do not depend on visible position or labels.[^9]

Across R01 through R09, the recurring architectural lesson is that boundaries must remain visible at failure and recovery time. That rationale supports explicit migration and shutdown commit boundaries, trust-isolated initialization and validation capabilities, stale-result suppression that does not rotate Base identity, and separation of authoritative state from rebuildable projections. It also supports the selected static typed extension points: stable identity and owner-controlled interfaces allow extensibility without a generic executable hook bus.

The reports remain supporting evidence only. Product behavior, public compatibility, security outcomes, and conformance conditions belong in the Core Documents and named NLSpecs; worked examples, operational recipes, library observations, and design advice remain in appendices or implementation-support guides. This ownership split prevents research guidance from becoming an accidental second normative contract.

# Appendix B. Non-normative worked examples

These examples illustrate the normative contracts. They do not add behavior.

## B.1 Selected Network Flow descriptor fields

The following object is an illustrative field fragment, not a conforming complete descriptor. Required fields not relevant to this example are omitted.

```json
{
  "profile_id": "network_flow_activity",
  "claimable": true,
  "contract_major": 2,
  "owner_contract_ref": "docs/network-flow-activity-nlspec.md#req:NF-REQ-001",
  "claim_config_key": "network_flow_activity.claimed",
  "route_families": [
    "/api/v1/incidents/{incident_id}/network-flow"
  ],
  "workspace_keys": [
    "network_analysis"
  ],
  "capability_ids": [],
  "runtime_dependencies": [
    {
      "profile_id": "import",
      "required_contract_major": 1
    }
  ],
  "contributions": [
    {
      "kind": "http_route_family",
      "contribution_id": "network_flow_activity.route_family",
      "route_family": "/api/v1/incidents/{incident_id}/network-flow"
    },
    {
      "kind": "incident_workspace",
      "contribution_id": "network_flow_activity.network_analysis_workspace",
      "workspace_key": "network_analysis"
    },
    {
      "kind": "import_target",
      "contribution_id": "network_flow_activity.import_target",
      "target_kind": "network_flow_table"
    },
    {
      "kind": "extension_resource_kind",
      "contribution_id": "network_flow_activity.network_flow_table_resource",
      "resource_kind": "network_flow_table"
    }
  ],
  "state_ownership": {
    "kind": "extension_versioned",
    "current_state_version": 1,
    "minimum_migratable_state_version": 1,
    "migration_lineage_id": "network_flow_activity.state_v1",
    "state_presence_contract_ref": "docs/network-flow-activity-nlspec.md#schema:cartulary.extension_state_presence_manifest.v1",
    "initialization_definition_ref": "docs/network-flow-activity-nlspec.md#schema:cartulary.extension_state_initialization_definition.v1",
    "initialization_definition_sha256": "0000000000000000000000000000000000000000000000000000000000000000",
    "final_state_validation_algorithm_id": "network_flow_activity_validate_state_v1",
    "final_state_validation_algorithm_ref": "docs/network-flow-activity-nlspec.md#algorithm:network_flow_activity_validate_state_v1"
  },
  "egress_mode": "none",
  "incident_portability_mode": "blocked_when_present",
  "snapshot_reporting_mode": "no_participation"
}
```

A conforming complete descriptor would add every required member and the exhaustive owner-derived contribution and public-schema declarations.

## B.2 Unclaimed discovery item

```json
{
  "profile_id": "network_flow_activity",
  "claimable": true,
  "claimed": false,
  "contract_major": 2,
  "route_families": [
    "/api/v1/incidents/{incident_id}/network-flow"
  ],
  "workspace_keys": [
    "network_analysis"
  ],
  "capabilities": []
}
```

## B.3 Dependency failure

```text
network_flow_activity.claimed = true
import.claimed = false
```

The configuration fails before readiness with top-level `invalid_deployment_config` and reason code `extension_dependency_not_claimed`. The runtime does not change `import.claimed`.

## B.4 Non-destructive unclaim

```text
Previous run: network_flow_activity.claimed = true
Current run:  network_flow_activity.claimed = false
```

The current run retains Network Flow authoritative state, starts no Network Flow worker, omits the Network Analysis workspace, and returns Core's reserved-unclaimed route error for the Network Flow family. Reclaim on a later restart validates the retained state before restoring availability.

## B.5 Transaction cancellation positions

If cancellation is already present before protocol step 7, the shared coordinator opens no database transaction. If cancellation arrives while step 10 validates participants, the coordinator requests rollback and reports cancellation only after absence of commit is proven. If cancellation races with the step-14 commit invocation, the ordinary cancellation response is no longer eligible; the result follows the normative final-commit outcome classification.

## B.6 Staged-object sweep cutoff

Suppose a cycle captures cutoff `T` and its eligible set needs three configured-size batches. Objects expiring at or before `T` remain in that cycle until all three batches are processed. An object expiring immediately after `T` waits for the next cycle. Independently, redemption of any unpublished object fails at its own expiry before storage access, so cleanup backlog does not extend logical availability.

## B.7 Fresh state initialization

Network Flow selects the `empty` initialization variant. The shared coordinator creates pending empty authoritative state and metadata in one transaction, invokes no profile initialization code, runs the final validator once against that pending state, and commits once only after a valid result. A profile selecting the algorithm variant follows the same transaction and final-validation boundary but invokes only its digest-bound scoped algorithm.

## B.8 Backup binding framing

For an empty binding, the digest input contains the domain-separation prefix and the framed binding ID, storage kind, and ASCII item count `0`; it has no entry frames. For a populated binding, entries are first ordered by their logical identity bytes, then each identity, raw content digest, and ASCII byte length is framed. These examples merely illustrate EXT-REQ-235; the appendix does not define another codec.

## B.9 Availability epoch rollover

When a generation increment would overflow, the client clears incident-local extension state, generates a new epoch, and continues at generation `1`. The existing `client_instance_id`, WebSocket resume identity, Base drafts, and Core pending queue remain unchanged. A response tagged with the prior epoch is ignored even if its numeric generation happens to match.

## B.10 Owner evidence accounting

A bounded integer still requires exact executable selector inventory for both admitted endpoints and each representable first invalid neighbor. One catalog row may register several exact symbols, titles, or scenarios, but the runner adapter must observe and account for every registered case before that row passes. The resulting row record contributes only to its resolved owner/target partition; full adoption additionally requires the compatible full-owner slice, applicable evidence-class gates, paired row-accounting and owner-summary shards, and a successful evidence audit over the exact roots.

## Sources

[^1]: `docs/research/R01-aurora_incident_response_report.md`, §1 “Executive Summary,” lines 3–17. The report is research evidence, not implementation authority.
[^2]: `docs/research/R03-Kanvas_technical_research_report.md`, §1 “Executive summary,” lines 2–20. The report is research evidence, not implementation authority.
[^3]: `docs/research/R02-cartulary_crm_tem_dfir_research_report.md`, §§1–2, lines 4–22. The report explicitly treats product implications as selective translations rather than direct authority.
[^4]: `docs/research/R04-responsive_browser_spreadsheet_ui_research_memo.md`, §§1–2, lines 7–21. The research is used only for responsiveness rationale.
[^5]: `docs/research/R05-responsive-interface-design-report.cr.md`, “Thesis” and “Executive summary,” lines 3–19. The research is used only for interaction rationale.
[^6]: `docs/research/R06-spreadsheet_of_doom_dfir_research_report.md`, §1 “Executive Summary,” lines 8–26. The report is practitioner-architecture evidence only.
[^7]: `docs/research/R07-spreadsheet-of-doom-sod-report.cr.md`, §§1–2, lines 7–21. The report is practitioner-architecture evidence only.
[^8]: `docs/research/R08-handsontable-react-research-report.md`, “Plugin / Extension Flow,” lines 602–628. The cited pattern is informative and does not authorize a runtime plugin system.
[^9]: `docs/research/R09-react-data-grid-research-report.md`, “Row model,” “Controlled versus internal props,” and row-key discussion, lines 95–119 and 259–265. The repository report is implementation evidence only.
