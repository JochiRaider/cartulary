---
title: Cartulary Reporting Subsystem NLSpec
status: proposed-amendment
document_class: nlspec
profile: snapshot_reporting
schema_id: cartulary.reporting_subsystem_nlspec.v1
document_version: 1.4.0
adopted_base_version: 1.3.0
---

# 1. Status, scope, and authority

This 1.4.0 cutover amendment is proposed over the adopted 1.3.0 base. Reference Pack binding and admitted-model identity changes require coordinated companion adoption; the implementation ledger records evidence and does not promote this revision.

This NLSpec defines the Cartulary Reporting Subsystem for the Snapshot and Reporting Extension Profile. It becomes implementation-conformance authority only after promotion to `status: adopted/current` and after every Core companion amendment in §5 is adopted or the affected requirement is explicitly reclassified as future-only.

**REQ-RPT-001**
This NLSpec governs only reporting-subsystem behavior inside the Snapshot and Reporting Extension Profile:

- canonical reporting export-model derivation;
- report, deck, and diagram intermediate models;
- party-scoped disclosure partitioning after snapshot materialization;
- redaction, token substitution, token manifests, and reveal-map production;
- Mermaid auto-layout source generation, manual-layout SVG serialization, validation, and local rendering;
- Slidev source generation, validation, rendering, and export;
- reveal-only presentation click steps;
- digest-bound report-composition consumption as a render input;
- render validation;
- deterministic render-bundle packaging and hashing;
- reporting-subsystem behavioral acceptance outcomes.

**REQ-RPT-002**
This NLSpec MUST NOT redefine Core 00 through Core 04 behavior outside the reporting subsystem. It MUST NOT redefine live workbook authorization, workbook-surface identity, source-record mutation semantics, graph-projection semantics, public snapshot or release route admission, incident membership, evidence access authorization, release approvals, deployment administration, WebSocket behavior, or Core 05 claim-publication behavior.

**REQ-RPT-003**
When this NLSpec conflicts with Core 00 through Core 04 outside the reporting subsystem, the conflict is a defect in this NLSpec. When this NLSpec conflicts with non-normative appendices, research reports, design guides, operating guidance, implementation guides, repository handoff material, or examples, this NLSpec governs only the reporting subsystem after adoption.

**REQ-RPT-004**
This NLSpec MUST specialize the existing Snapshot and Reporting route family rather than create a second reporting route family. Public snapshot creation, release creation, release-resource state, release approval, release invalidation, public route admission, and public route-family error envelopes remain owned by Core 01 and Core 04. The Reporting Subsystem owns render-attempt behavior after Core-owned route admission accepts a release render operation.

**REQ-RPT-005**
The installed rendering runtime is selected by the `snapshot_reporting.configuration.v2` configuration contract in §20 and Core 04 REQ-04-046c. Template deadlines and resource limits are manifest-owned and MUST NOT become deployment settings. Any additional deployment-configuration key MUST define its key name, type, default, allowed values, fail-closed validation behavior, secret classification, backup or portability behavior, and owner interaction with Core 04 deployment configuration.

**REQ-RPT-006**
A repository MUST NOT promote this NLSpec to `adopted/current` until the promotion conditions in Table 1-A are satisfied.

**Table 1-A. Promotion conditions**

| Promotion condition | Required state before adoption |
| --- | --- |
| Core 00 adopted-subsystem registry | This NLSpec is listed as adopted/current for the Reporting Subsystem only. |
| Core 01 output vocabulary | Core 01 recognizes only `slidev` and `mermaid` as current Reporting v1 conformant output kinds, or explicitly marks any other route selector as non-conformant future territory for this NLSpec. |
| Core 01 release tuple | `render_admitted_at`, public `output_options` if exposed, and `output_sha256=bundle_manifest_sha256` are adopted. |
| Core 02 party assignments | `record_party_assignment.v1` is adopted as source state for Host and Identity party assignment. |
| Core 03 party-assignment editing | Workbook or inspector editing behavior for Host and Identity party assignment is adopted without creating a hot-path recipient-visibility workflow. |
| Core 04 reveal-map boundary | Reveal maps are classified as internal sensitive release artifacts with authorization and retention rules. |
| Core 04 render sandbox | The render sandbox security boundary is adopted. |
| Core 01 composition tuple | Nullable composition tuple fields and composition byte-form recognition are adopted. |
| Core 04 authored presentation text gate | `allow_authored_presentation_text` is adopted with default `false`. |
| Report Composition NLSpec (`docs/report-composition-nlspec.md`) | `cartulary.report_composition.v1` is adopted as the authoring and schema owner for compositions consumed by this NLSpec. |
| Reporting derivation profile | `cartulary.reporting_derivation_profile.v2` is adopted and every `derivation_version` resolves to it under REQ-RPT-027a. |
| Reporting behavioral validation | Every current behavior in this NLSpec has an implementation disposition and production-facing validation where implemented. |

# 2. Normative language and document discipline

**REQ-RPT-007**
The key words **MUST**, **MUST NOT**, and **MAY** are normative in this NLSpec. **MUST** and **MUST NOT** define conformance requirements. **MAY** defines optional behavior whose omission behavior is explicit in the same requirement, table row, or immediately following paragraph.

**REQ-RPT-008**
The word `default` defines the value or behavior used when an optional member is omitted and omission is valid. A default is a conformance requirement, not advice.

**REQ-RPT-009**
A conforming implementation MUST treat object member names, closed-vocabulary tokens, schema identifiers, failure codes, reason codes, file roles, algorithm identifiers, and profile identifiers as exact Unicode code point sequences. It MUST NOT apply case folding, locale comparison, trimming, or Unicode normalization unless this NLSpec names the exact operation.

**REQ-RPT-010**
Accepted normative sections of this NLSpec MUST NOT contain unresolved placeholder text. Any behavior not fully determined in this draft MUST be marked as an external dependency, a Core amendment dependency, a future-only area, or an explicit owner-blocker entry with a conformance consequence.

**REQ-RPT-011**
Accepted normative prose MUST NOT use open delegation phrases. Invalid open-delegation patterns include appropriateness-based delegation, need-based delegation, template-defined behavior without a named template schema, renderer-local behavior without a named renderer contract, implementation-local behavior without a closed observable boundary, tool-local behavior without a pinned tool contract, equivalence language without an equivalence table, and availability language without an exact data-presence precondition. Such language is valid only when the same sentence or immediately adjacent table names the closed schema, grammar, algorithm, extension container, owner dependency, or future-only boundary that makes the phrase decidable.

**REQ-RPT-012**
A paragraph, table, or code fence that describes examples, rationale, implementation notes, or non-normative guidance MUST NOT introduce behavior that is absent from the normative requirements. If an example contradicts a requirement, the requirement governs and the example is defective.

# 3. Purpose

**REQ-RPT-013**
The Reporting Subsystem MUST transform immutable incident snapshots into deterministic, self-contained reporting artifacts without weakening the workbook hot path or the release approval boundary.

**REQ-RPT-014**
The subsystem MUST make the reporting questions in Table 3-A answerable without querying live workbook state after export-model materialization.

**Table 3-A. Reporting questions and required support**

| Question | Required subsystem support |
| --- | --- |
| What timeline facts are eligible for a recipient-specific report or deck? | Canonical export-model sections, timeline-selection rules, content classes, support references, disclosure partitions, and redaction outcomes. |
| How are identities, hosts, indicators, evidence, parties, and related entities represented in diagrams? | Completed Graph Projection consumption, diagram objects, auto-layout Mermaid source generation, manual-layout SVG serialization, source references, overflow reporting, and validation. |
| Which values are visible, tokenized, stubbed, masked, dropped, or blocked for a recipient? | Disclosure partitions, redaction profiles, token manifests, redaction manifests, and reveal-map handling. |
| How are Slidev decks generated reproducibly? | `cartulary.reporting_slide_deck.v2`, Slidev subset validation, reveal-only click-step profile, toolchain snapshot, and deterministic render bundle. |
| What bytes are approved and published? | Per-file SHA-256 entries, canonical render-bundle manifest, root `output_sha256`, release binding, and rerender determinism. |

**REQ-RPT-015**
The subsystem MUST preserve the evidence-versus-presentation boundary. Generated reports, decks, diagrams, and summaries are permitted to rearrange, select, redact, visualize, and summarize snapshot facts only according to this NLSpec. They MUST NOT invent facts, actors, timestamps, commands, causal chains, source evidence, party relationships, or conclusions not present in the reporting export model.

# 4. Non-goals

**REQ-RPT-016**
The Reporting Subsystem MUST NOT introduce behavior in Table 4-A.

**Table 4-A. Non-goals**

| Non-goal | Required omission behavior |
| --- | --- |
| Live workbook visibility changes | Export redaction MUST NOT hide live workbook rows, fields, evidence, search results, saved views, or filters from incident participants. |
| Record-specific ACLs | Recipient-specific withholding MUST occur at snapshot, render, and release time, not by adding record-level, field-level, or party-specific live ACLs. |
| New first-class `post` record family | External or analyst “post” language MUST map through §6.3 unless a later Core owner defines a source record family. |
| Graph-projection redefinition | Graph derivation, projected identity, lifecycle, query behavior, traversal, and graph validation remain owned by the Graph Projection NLSpec. |
| New public reporting routes | Snapshot and release route families remain Core-owned. This NLSpec owns render semantics after route admission. |
| Composition authoring ownership | Composition authoring routes, resource lifecycle, version history, operation wire schemas, and builder-facing validation codes remain owned by Core 01 and `docs/report-composition-nlspec.md`. |
| Raw generated-source editing | Report composition MUST NOT edit generated `slides.md`, `.mmd`, rendered files, bundle manifests, or post-redaction output bytes. |
| Composition workbook mutation | Report composition MUST NOT create, update, or delete workbook records, links, graph-projection output, snapshots, or template packs. |
| Cross-template composition portability | A composition binds to one exact template version; cross-template migration is future-only. |
| Collaborative composition editing | Concurrent builder co-editing and WebSocket collaboration are future-only and do not affect Reporting conformance. |
| Composition-authored case facts | Free-text findings, analysis, causal chains, commands, timestamps, actors, or conclusions authored in a composition are invalid for current Reporting; source case narrative must enter through snapshot artifact records and template narrative slots. |
| Template-executed arbitrary code | Templates MUST NOT execute arbitrary user-supplied code or fetch undeclared resources. |
| Remote runtime assets | External release outputs MUST be self-contained and MUST NOT require remote JavaScript, CSS, fonts, images, media, themes, or package resolution. |
| Interactive external deck runtime | External release v1 outputs MUST NOT require viewer-side JavaScript interaction beyond ordinary PDF, PPTX, SVG, or image viewing. |
| Claim-publication benchmarks | Timed or fixture-sensitive publication claims remain Core 05 material and are not created by this NLSpec. |
| Canonical archive bytes | Physical ZIP, TAR, or other archive bytes are not approval bytes in this revision. Approval binds to `bundle_manifest_sha256`. |

# 5. Core interactions, companion edits, and promotion blockers

## 5.1 Core interaction map

**REQ-RPT-017**
Reporting-to-Core interactions MUST use the classifications in Table 5-A. A row classified as `blocked-until-core-adoption` disables the affected Reporting requirements for external-release conformance until the Core owner is adopted.

**Table 5-A. Core interaction classifications**

| Classification | Meaning |
| --- | --- |
| `core-owned-current` | Core has adopted the needed behavior and Reporting imports it. |
| `reporting-owned-current` | Reporting owns behavior after Core route admission. |
| `blocked-until-core-adoption` | Requirement is inactive for external-release conformance until Core adopts the dependency. |
| `future-only` | Not current conformance behavior. |
| `adopted-subsystem-current` | An adopted subsystem NLSpec owns adjacent behavior and Reporting imports only its declared consumer boundary. |

**REQ-RPT-018**
A `blocked-until-core-adoption` row MUST name the affected Reporting requirements. Those requirements MUST NOT be counted as satisfied in an external-release conformance claim until the Core owner is adopted. A Reporting implementation MAY retain internal experimental behavior, but validation summaries MUST report `blocked_core_dependency=<dependency_id>` and external release MUST fail closed.

**Table 5-B. Core interaction map**

| Owner | Interaction | Classification | Affected Reporting requirements |
| --- | --- | --- | --- |
| Core 00 | Adopted-subsystem registry entry and precedence. | `core-owned-current` | REQ-RPT-001..REQ-RPT-006 |
| Core 01 | Snapshot and release route admission; public release resource state; public route error envelope; release tuple fields. | `core-owned-current` | REQ-RPT-022..REQ-RPT-035, REQ-RPT-109..REQ-RPT-115 |
| Core 01 | `render_admitted_at` binding, `output_options` if public, and `output_sha256=bundle_manifest_sha256`. | `core-owned-current` | REQ-RPT-027..REQ-RPT-035, REQ-RPT-047..REQ-RPT-052, REQ-RPT-097..REQ-RPT-099, REQ-RPT-105..REQ-RPT-108 |
| Core 01 | Nullable composition tuple fields, composition digest byte form, composition authoring route family, and version freeze at release binding. | `core-owned-current`; Report Composition adoption still required for authoring owner conformance | REQ-RPT-027f, REQ-RPT-053..REQ-RPT-054, REQ-RPT-087d..REQ-RPT-087h |
| Core 01 | Report-composition preview job or attempt identity for `cartulary.report_composition_preview_view.v1`. | `core-owned-current`; Reporting imports identity only | REQ-RPT-087h |
| Core 01 | Release tuple `graph_projection_refs[]` admission and exact completed-projection binding. | `core-owned-current`; Graph Projection remains adopted owner of lifecycle/output | REQ-RPT-027g, REQ-RPT-076..REQ-RPT-080 |
| Core 01 | Redaction-profile actions `mask` and `stub`; Reporting-owned token substitution consumes those actions through the closed Reporting-facing profile view. | `core-owned-current`; token derivation remains `reporting-owned-current` | REQ-RPT-091..REQ-RPT-105 |
| Core 02 | `record_party_assignment.v1` source-state relation for Host and Identity records. | `core-owned-current` | REQ-RPT-075..REQ-RPT-090 |
| Core 02 | Party partition segment exposed or derivable for every Party used in Reporting disclosure partition refs. | `core-owned-current` | REQ-RPT-057..REQ-RPT-059 |
| Core 02 | Immutable snapshot `record_created_at` materialized for every selected timeline event. | `core-owned-current` | REQ-RPT-070..REQ-RPT-073 |
| Core 03 | Workbook or inspector interaction boundary for Host and Identity party assignments. | `core-owned-current` | REQ-RPT-075..REQ-RPT-081 |
| Core 04 | Live authorization remains incident-role based; export partitions do not affect live workspace access. | `core-owned-current` | REQ-RPT-016, REQ-RPT-075..REQ-RPT-090 |
| Core 04 | Reveal-map sensitive-release-artifact authorization and retention. | `core-owned-current` | REQ-RPT-100..REQ-RPT-105 |
| Core 04 | Render sandbox trust boundary. | `core-owned-current` | REQ-RPT-119..REQ-RPT-122 |
| Core 04 | Redaction-profile `allow_authored_presentation_text` control with default `false`. | `core-owned-current` | REQ-RPT-062, REQ-RPT-087f |
| Core 01 | Release-create recipient partition validation against the selected redaction profile. | `core-owned-current` | REQ-RPT-027b, REQ-RPT-059a |
| Core 01 | Redaction-manifest digest byte form recognition. | `core-owned-current` | REQ-RPT-062a |
| Core 02 / Core 04 | Party public-directory eligibility and redaction-profile permission for public Party labels. | `core-owned-current` | Public-label branch of REQ-RPT-058 |
| Graph Projection NLSpec | Projection input, output, lifecycle, validation, identity, and consumer behavior. | `adopted-subsystem-current` | REQ-RPT-076..REQ-RPT-080 |
| Report Composition NLSpec (`docs/report-composition-nlspec.md`) | Composition schema, operation vocabulary, semantic anchors, lifecycle, authorization surface, versioning, and builder-facing validation codes. | `adopted-subsystem-current` | REQ-RPT-027f, REQ-RPT-053..REQ-RPT-054, REQ-RPT-079e, REQ-RPT-087d..REQ-RPT-087h |
| Reporting derivation profile | Adoption of `cartulary.reporting_derivation_profile.v2` as the versioned owner of snapshot-to-export-model content derivation referenced by `derivation_version`. | `reporting-owned-current` | REQ-RPT-027a |

## 5.2 Companion edit status

**REQ-RPT-019**
The companion edits in Table 5-C record the owner status required for this NLSpec. Rows marked `closed` are already adopted by the named owner document and MUST be treated as current owner input. A later row marked `open` MUST name the blocked requirements and MUST use the fail-closed behavior in REQ-RPT-018 until its owner document is adopted.

**Table 5-C. Companion edit status**

| Owner | Status | Dependency closed |
| --- | --- | --- |
| Core 00 | `closed` | Adopted-subsystem precedence for this NLSpec. |
| Core 01 | `closed` | Reporting v1 route output vocabulary, `render_admitted_at`, `output_options`, `output_sha256=bundle_manifest_sha256`, composition tuple fields, composition byte form, composition authoring route reservation, preview identity, redaction-manifest byte form, and `graph_projection_refs[]`. |
| Core 02 | `closed` | `record_party_assignment.v1`, stable Party partition segments, Party public-directory eligibility, immutable timeline `record_created_at`, and source/export metadata needed for current snapshot derivation. |
| Core 03 | `closed` | Workbook and inspector interaction boundary for Core-owned party assignment source state without creating a hot-path recipient-visibility workflow. |
| Core 04 | `closed` | Reveal-map sensitive-artifact authorization and retention, render sandbox boundary, authored-presentation-text default, superseded-record external-release default, `neutral_token_family` default, recipient partition validation boundary, and public Party-label permission default. |
| Report Composition NLSpec | `closed` | `cartulary.report_composition.v1` authoring and schema ownership for digest-bound compositions consumed by Reporting. |
| Graph Projection NLSpec | `closed` | Completed digest-bound graph projection output lifecycle and consumer validation. |
| Reporting derivation profile | `closed` | `cartulary.reporting_derivation_profile.v2` is the versioned owner for snapshot-to-export-model derivation decisions referenced by `derivation_version`. |

**REQ-RPT-019a**
Reporting imports the adopted Extensions Subsystem typed participant boundary
without transferring report, snapshot, render, release, composition, or
redaction ownership. Its shared-owner participant ID is
`snapshot_reporting.render_export_v1`, and its implementation algorithm has
that same exact identity. The participant specialization uses
`participant_kind='snapshot_reporting'`,
`shared_context_schema_id='cartulary.extension_snapshot_reporting_participant_context.v1'`,
and exactly one `emit` operation. That operation returns only
`cartulary.extension_snapshot_reporting_participant_result.v1`, emits only
`cartulary.reporting_export_model.v7`, orders under
`materialize_reporting_export_model_v2`, uses this requirement as its
authorization contract, uses REQ-RPT-059a as its redaction contract, and uses
REQ-RPT-114 as its error contract. It admits at most 67,108,864 input bytes,
67,108,864 output bytes, and 1,048,576 selected or emitted items.

The participant accepts only the
digest-bound `cartulary.extension_snapshot_reporting_participant_context.v1`
specialization and returns only its matching closed result. Reporting invokes a
profile only when its canonical descriptor selects
`snapshot_reporting_mode='participant'`, its exact specialization and packaged
binding pass admission, and the resolved claim/state matrix permits the
operation. `no_participation` means no call and no implicit data discovery.

Participant input is a scoped immutable snapshot/export-model view, never live
workbook state or unrestricted storage. Participant validation, ordering, the
64 MiB input/output ceilings, cancellation and deadlines use the shared
Extensions/Core 01 protocol. Reporting remains the sole owner of how a valid
participant result is admitted to a report model and rendered output. An absent,
inactive, malformed, incompatible, over-limit, timed-out, or unregistered
participant follows the adopted shared invocation matrix and never falls back to
package scanning, profile callbacks, a compatibility reader, or silent omission.

The exact durable job facts owned by this profile are
`snapshot_reporting.snapshot_create_v4` ->
`snapshot_reporting.snapshot_create` with progress unit
`snapshot_reporting.snapshot_create.materialization.v1`,
`snapshot_reporting.release_create_v4` ->
`snapshot_reporting.release_create` with progress unit
`snapshot_reporting.release_create.render_attempt.v1`, and
`snapshot_reporting.composition_preview_v4` ->
`snapshot_reporting.composition_preview` with progress unit
`snapshot_reporting.composition_preview.render_attempt.v1`. Their only worker kind is
`snapshot_reporting.job_worker_v2`. All three jobs require route-scoped
idempotency, a proof on terminal success, and precommit-observable
cancellation under Core 01 REQ-01-634. Snapshot create permits one `snapshot`
resource ref, release create permits one `release` resource ref, and
composition preview permits no resource ref. Preview still commits a canonical
terminal success and proof.

Release and preview v4 Jobs MUST retain a closed
`cartulary.reporting_render_admission.v1` recipe before enqueue commits. The
recipe freezes `preparation_algorithm=materialize_retained_reporting_render_v1`,
the exact request, actor, snapshot metadata and source-capture digest, Reference
Pack binding, nullable release/preview identity pair, fixed admission timestamp,
complete template manifest and files, exact authored redaction policy, output
options, recipients, Graph results and Composition bytes. Authored policy,
template, Graph and whole-recipe schema-domain digests MUST also be retained as
independent typed admission metadata. Readers reject unknown members, missing
members, changed bytes, retired algorithms or a mismatch with Jobs, snapshot or
Composition admission. Private byte arrays use JSON base64 before canonicalization.
The recipe MUST NOT contain an expanded snapshot or export model.

Source expansion and content-dependent policy materialization MUST run as
cancelable worker preparation, without retaining a Jobs transition lock. An
unsupported selector or malformed public request still rejects before enqueue.
Recipient membership in retained source and a preview's expected materialized
profile digest are checked during preparation; failure produces a failed Job
with owner diagnostics and no publishable output. This distinction MUST NOT
weaken authentication, snapshot visibility, role or CSRF checks at admission.
Worker preparation uses the retained authored policy bytes, never a fresh
selector resolution. The public request/idempotency digest remains the exact
normalized public request and is not replaced by a recipe or model digest.

If the snapshot's exact pinned Reference Pack set cannot be reconstructed during
worker preparation, the Job MUST fail with `error.code=required_reference_pack_unavailable`.
The retained render diagnostic uses `failure_code=export_model_invalid`,
`reason_code=required_reference_pack_unavailable`, and stage
`export_model_materialization`. Infrastructure errors and cancellation MUST NOT
be classified as missing content. No active or built-in set may be substituted.

A completed preparation MUST retain only the closed
`cartulary.reporting_render_preparation.v1` seal: `schema_id`,
`admission_sha256`, `source_capture_sha256`, `snapshot_model_id`, `export_model_id`,
`export_model_sha256`, `source_facts_sha256`, and `redaction_profile_sha256`.
Its independent schema-domain digest binds the exact seal. Restart derives the
same inputs from the recipe and capture and compares this seal; it MUST NOT
rewrite a different preparation. Preparation, participant invocation, rendering,
validation and publication share one manifest-owned total-attempt deadline.
Cancellation or execution loss drains active work before terminal mutation.

Candidate validation MUST finish before the final Jobs lock. The owner passes
an opaque, detached candidate into publication; the transaction rechecks its
recipe and preparation bindings and the live execution fence before committing
output. No incident-sized model construction or redaction runs under that lock.
A failure before preparation has no valid model or materialized-profile digest:
it retains Job diagnostics and MUST NOT invent a release row or sentinel digest.
After preparation, failed release rendering retains the truthful `render_failed`
release identity and prepared digests. Preview failure never creates a release.
Recovery accepts these distinct states only with their exact retained evidence.

Release/preview v2 and v3 Jobs and their full-model input formats are retired.
The pre-production cutover rejects incompatible history before mutation; it
provides no legacy reader, inferred provenance or in-place downgrade.

The participant may read only the immutable snapshot/export-model view supplied
through the scoped accessor. Reporting validates and admits the returned
canonical model before redaction and rendering. The current emit operation MUST return the exact canonical bytes and object digest of its frozen admitted model; an internally consistent but changed model is rejected. A future transforming participant requires a separately adopted contract that explicitly retains its input and output identities. Release creation and
authoritative composition preview invoke only the admitted participant when
the `snapshot_reporting` profile is claimed and the required immutable
snapshot exists. Profile-local extension state metadata, migrations, codecs,
and backup bindings are forbidden for this Core-managed profile.

**REQ-RPT-019b**
Every Reporting schema-validation rule used by the Extensions participant
surface MUST carry its complete condition annotation, and every procedural
participant validator MUST publish a closed decision table. Those declarations
are the sole Reporting input to the generated extension validation-condition
registry. A reachable undeclared condition, an extra or stale declaration, or an
emitted condition absent from that registry is a conformance failure. Reporting
validation results use the shared precedence: invocation failure, structural
invalidity, overflow at `4097+` findings, remaining schema defects, valid
findings, then valid empty result; `257..4096` is the ordinary 256-item bound
violation.

# 6. Concepts, identifiers, and source-family mapping

## 6.1 Core concepts

**REQ-RPT-020**
The terms in Table 6-A have the meanings defined here inside this NLSpec.

**Table 6-A. Concepts**

| Term | Definition |
| --- | --- |
| `reporting export model` | Canonical JSON-compatible model materialized from an immutable snapshot and used as the only case-content input for report, deck, diagram, redaction, and render operations. |
| `deck model` | Deterministic intermediate object between the reporting export model and generated Slidev Markdown. |
| `render bundle` | Directory-shaped logical artifact containing source, rendered files, local assets, manifests, and validation artifacts for one release candidate. |
| `release tuple` | Core-owned release identity plus snapshot, derivation, template, redaction, recipient, output, approval, toolchain, and environment values that affect released bytes. |
| `disclosure partition` | Export-time content partition used to determine recipient eligibility. It is not a live ACL. |
| `recipient partition` | Canonical allow-list of disclosure partitions selected for an external release. |
| `token` | Deterministic replacement string emitted by token-backed `mask` or `stub` redaction for an eligible subject. |
| `reveal map` | Sensitive internal adjunct that maps emitted tokens back to original subjects and display values. |
| `template pack` | Versioned local bundle that declares output kinds, bindings, sections, layouts, local assets, and renderer profiles. |
| `Mermaid source` | Validated `.mmd` text generated by this subsystem for auto-layout diagrams from redacted export-model and graph-projection data. |
| `manual diagram SVG` | Deterministic `image/svg+xml` generated by this subsystem for a manual-layout diagram from closed composition layout data and the post-redaction resolved diagram. |
| `Slidev source` | Generated `slides.md` conforming to `slidev_markdown_serialize_v2`. |
| `click step` | One deterministic reveal step under `cartulary.slidev_reveal_only.v2`. |
| `report composition` | Digest-bound companion-spec document that carries presentation operations, authored presentation text, and composition diagram declarations consumed before deck chunking and rendering. |
| `composition anchor` | Companion-spec semantic target reference resolved by Reporting against template declarations, source records, block context, or diagram declarations. Structural ordinal-path IDs are not composition anchors. |
| `render attempt` | One admitted Reporting-owned execution after Core route admission. |

## 6.2 Closed schema and algorithm identifiers

**REQ-RPT-021**
This revision defines exactly the identifiers in Table 6-B. Every identifier in
the table MUST have a normative owner section and observable behavioral
validation where implemented.

**Table 6-B. Closed identifiers**

| Identifier | Kind | Owner section |
| --- | --- | --- |
| `cartulary.reporting_render_request_options.v1` | schema | §7 |
| `cartulary.reporting_derivation_profile.v2` | schema | §7 |
| `cartulary.reporting_export_model.v7` | schema | §9 |
| `cartulary.reporting_section.v2` | schema | §9 |
| `cartulary.reporting_block.v3` | schema | §9 |
| `cartulary.reporting_field.v1` | schema | §9 |
| `cartulary.reporting_record_summary.v1` | schema | §9 |
| `cartulary.reporting_relationship_summary.v1` | schema | §9 |
| `source_record_ref.v1` | schema | §9 |
| `relationship_endpoint_ref.v1` | schema | §9 |
| `cartulary.reporting_asset_declaration.v1` | schema | §9 |
| `cartulary.reporting_support_ref.v1` | schema | §9 |
| `cartulary.reporting_timeline_event.v1` | schema | §14 |
| `cartulary.tokenizable_subject.v1` | schema | §13 |
| `stable_subject_ref_v1` | identifier grammar | §13 |
| `cartulary.reporting_token_manifest.v3` | schema | §13 |
| `cartulary.reporting_token_reveal_map.v3` | schema | §13 |
| `cartulary.reporting_redaction_profile_view.v2` | schema | §13 |
| `cartulary.reporting_redaction_rule_view.v1` | schema | §13 |
| `selected_rule_trace.v1` | schema | §13 |
| `cartulary.reporting_diagram.v1` | schema | §15 |
| `source_projection_ref.v2` | schema | §15 |
| `diagram_selection_rule.v1` | schema | §15 |
| `diagram_overflow_summary.v1` | schema | §15 |
| `derive_diagram_label_v1` | derivation algorithm | §15 |
| `mermaid_source_serialize_v3` | serializer | §16 |
| `cartulary.reporting_slide_deck.v2` | schema | §17 |
| `slidev_markdown_serialize_v2` | serializer | §18 |
| `field_value_to_text_v1` | serialization algorithm | §18 |
| `click_step.v1` | schema | §19 |
| `cartulary.reporting_toolchain_snapshot.v2` | schema | §20 |
| `cartulary.reporting_template_pack_manifest.v3` | schema | §21 |
| `template_output_option_support.v1` | schema | §21 |
| `template_asset_item.v1` | schema | §21 |
| `cartulary.reporting_timeline_adjacency.v1` | identity preimage | §15.2; exact directional Timeline edge identity, not a persisted source record. |
| `template_diagram_decl.v1` | schema | §21 |
| `cartulary.render_bundle_manifest.v3` | schema | §22 |
| `cartulary.reporting_render_validation_summary.v1` | schema | §23 |
| `cartulary.reporting_export_model_validation.v1` | schema | §9 |
| `cartulary.redaction_manifest.v5` | schema | §13 |
| `safe_details.v1` | schema | §23 |
| `first_failure.v1` | schema | §23 |
| `cartulary.reporting_export_model_id.v1` | schema | §10 |
| `cartulary.reporting_deck_id.v1` | schema | §10 |
| `cartulary.reporting_token_id.v1` | schema | §10 |
| `cartulary.click_state.v1` | schema | §19 |
| `reporting_canonical_json_v1` | canonicalization algorithm | §10 |
| `content_manifest_digest_v1` | digest algorithm | §10 |
| `materialize_reporting_export_model_v2` | derivation algorithm | §11 |
| `assign_party_disclosure_partitions_v1` | derivation algorithm | §12 |
| `filter_disclosure_partitions_v1` | derivation algorithm | §12 |
| `retain_declared_structure_v1` | derivation algorithm | §12 |
| `apply_token_backed_redaction_v1` | redaction algorithm | §13 |
| `select_timeline_rows_v1` | derivation algorithm | §14 |
| `derive_section_ordering_key_v1` | derivation algorithm | §9 |
| `derive_display_token_v1` | derivation algorithm | §13 |
| `derive_deck_v1` | derivation algorithm | §17 |
| `derive_deck_v2` | derivation algorithm | §17 |
| `template_and_composition_reachable_records_v1` | derivation-profile algorithm token | §7 |
| `truncate_label_v1` | serialization algorithm | §16 |
| `diagram_layout_svg_serialize_v1` | serialization algorithm | §16 |
| `serialize_block_markdown_v1` | serialization algorithm | §18 |
| `slidev_markdown_escape_v2` | serialization algorithm | §18 |
| `render_slidev_bundle_v1` | render algorithm | §24 |
| `render_mermaid_bundle_v1` | render algorithm | §24 |
| `render_sandbox_policy_v1` | security policy | §24 |
| `render_sandbox_observation.v1` | schema | §24 |
| `render_attempt_lifecycle_v1` | lifecycle | §24 |
| `validate_release_render_v1` | validation algorithm | §23 |

**REQ-RPT-021a**
Reporting imports the companion-owned identifiers in Table 6-B1 only for consumer-boundary validation and deterministic render behavior. This NLSpec MUST NOT define their authoring route shape, persistence model, draft mutation behavior, builder UI behavior, or complete wire schema. If the companion owner is not adopted/current, every non-null composition tuple MUST fail closed with `error.code='release_render_failed'`, `failure_code='composition_invalid'`, and `reason_code='blocked_core_dependency'`.

**Table 6-B1. Imported Report Composition identifiers**

| Identifier | Kind | Reporting consumer use |
| --- | --- | --- |
| `cartulary.report_composition.v1` | companion schema | Canonical digest-bound composition document named by the release tuple. |
| `composition_op.v1` | companion schema | Closed operation object whose Reporting effect is Table 17-D. |
| `section_anchor` | companion anchor grammar | Resolves to one emitted section instance using template declaration identity and expansion-dimension key. |
| `record_anchor` | companion anchor grammar | Resolves to one Reporting-generated record summary by source `record_id`. |
| `block_anchor` | companion anchor grammar | Resolves to one derived block by section anchor, block kind, and optional record anchor. |
| `diagram_anchor` | companion anchor grammar | Resolves to one template-owned or composition-owned diagram declaration by `decl_id`. |
| `authored_text.v1` | companion schema | Presentation-tier text node admitted only through REQ-RPT-087f. |
| `composition_diagram_decl.v1` | companion schema | Composition-owned diagram declaration mapped into `cartulary.reporting_diagram.v1`. |
| `cartulary.report_composition_preview_source.v1` | companion schema | Internal-draft preview descriptor accepted only by REQ-RPT-087h and never by external release. |

## 6.3 `post` source-family mapping

**REQ-RPT-022**
`post` is not a current Cartulary source-record family in this NLSpec. A conforming reporting export model MUST NOT emit `post_ref`, `post_id`, `record_type='post'`, or `source_record_family='post'` unless a later Core owner defines that source record family and this NLSpec is revised.

**REQ-RPT-023**
The subsystem MUST map incoming “post” terminology according to Table 6-C before export-model materialization.

**Table 6-C. `post` source-family anti-corruption mapping**

| Input term | Canonical Cartulary target |
| --- | --- |
| analyst note, draft update, status prose, communication excerpt, handoff note, lesson text | `artifact` with `artifact_type` equal to `note`, `status_update`, `handoff_note`, or `lesson`, or the existing coordination surface that owns the source record. |
| stakeholder update or meeting, chat, or email summary | `comm_log` artifact or Communications Log surface record. |
| finding-like report claim | Standardized `finding` artifact subtype if implemented; otherwise a curated export-model block with support references. |
| investigative query text | Standardized `investigative_query` artifact subtype if implemented; otherwise `artifact`. |
| external social-media or third-party “post” | Evidence excerpt or artifact excerpt linked to source evidence; future first-class source record only when Core defines it. |

**REQ-RPT-024**
When a fixture, import, template binding, or report section names an unmapped `post` source family, the implementation MUST fail with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='unsupported_source_family'` before render persistence.

# 7. Reporting source boundary, release tuple, and output options

## 7.1 Source boundary

**REQ-RPT-025**
The reporting subsystem MUST accept case content only through an immutable snapshot boundary admitted by the Core Snapshot and Reporting route family. It MUST NOT read mutable workbook tables, mutable projection tables, live search indexes, live graph views, or live evidence metadata after `materialize_reporting_export_model_v2` completes.

**REQ-RPT-026**
The implementation MUST enforce the source boundary with an observable guard. Any attempt by a renderer, template, Mermaid generator, Slidev generator, asset resolver, or validation stage to query live workbook tables or mutable projections after export-model materialization MUST fail the render with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='live_query_after_export_model'`.

### 7.1.1 Evidence provider-output contract

**REQ-RPT-026a**
The Evidence source provider MUST construct the value carried by
`cartulary.reporting_owner_field_fact.v1` from the closed allowlist in Table
7-A0. The provider MUST omit every Evidence member not listed in that table.
The allowlist applies before redaction and does not grant a field release
eligibility; every retained value continues through the ordinary Reporting
field classification, subject mapping, redaction, partition, and release-scope
rules.

**Table 7-A0. Allowed Evidence provider value members**

| Member | Required provider posture |
| --- | --- |
| `record_id` | Include when present in the eligible immutable source record. |
| `title` | Include when present, preserving the immutable source value and null posture. |
| `lifecycle_state` | Include when present. Eligibility filtering remains separately mandatory. |
| `requested_at` | Include when present, preserving null versus timestamp. |
| `received_at` | Include when present, preserving null versus timestamp. |
| `collector_party_text` | Include when present, preserving null versus string. |
| `collector_party_id` | Include when present, preserving null versus identifier. |
| `source_party_text` | Include when present, preserving null versus string. |
| `source_party_id` | Include when present, preserving null versus identifier. |
| `upload_state` | Include when present. |
| `created_at` | Include when present. |
| `updated_at` | Include when present. |

The provider MUST NOT emit any member or value class in Table 7-A0a through
this field-fact value, a sibling field fact, support metadata, diagnostics, or
an extension map.

**Table 7-A0a. Forbidden Evidence provider output**

| Forbidden member or value class | Required posture |
| --- | --- |
| `incident_id`, `blob_hash`, `storage_ref`, `object_blob_id` | Omit. These members MUST NOT occur in the provider value. |
| Physical object keys, bucket names, storage paths, endpoint identities, or other object-store identities | Omit. Logical source identity MUST NOT be derived from physical storage identity. |
| Upload targets, upload tokens, access handles, access capabilities, presigned parameters, or capability-binding material | Omit. |
| Object bytes, preview bytes, byte excerpts, or byte-derived content not admitted by another adopted Reporting source contract | Omit. |
| Credentials, credential material, secret references, authentication state, or session identifiers | Omit. |
| Reveal maps, reveal-map entries, token reversal material, or unredacted replacement material | Omit. |
| Internal errors, stack or package identities, implementation type names, SQL detail, or backend diagnostics | Omit. |

Unknown future Evidence fields MUST default to omission. They MUST remain
omitted until the Reporting owner revises Table 7-A0 through an adopted change;
adding a field to Evidence storage, a projection, a generated contract, or a
provider struct MUST NOT add it to Reporting output. Implementations MUST NOT
use reflection, unrestricted row-to-map conversion, whole-row JSON
serialization, or equivalent automatic field propagation to construct this
provider value.

Deleted Evidence and Evidence otherwise ineligible under the immutable source
selection rules MUST remain excluded. Eligible Evidence support references
MUST continue to use the logical identity `evidence:{evidence_id}` under the
`cartulary.reporting_support_ref.v1` grammar. They MUST NOT use a blob identity,
storage identity, capability identity, or physical object locator.

### 7.1.2 Links provider-output and support-reference contract

**REQ-RPT-026b**
Reporting MUST receive Links source state through Reporting-owned field-provider
and support-reference-provider ports. Application composition MUST implement
both ports by adapting the Links-owned active fact reader and the logical target
paths produced by eligible source owners. Reporting MUST NOT query
`record_links`, `record_tags`, or a Links-owned view directly, and Links MUST NOT
construct Reporting DTOs, paths, content classes, support identifiers,
selection policy, redaction policy, or canonical output.

The Links reader MUST accept the caller's context, caller-owned transaction, and
`incident_id`. It MUST return non-nil ordered `record_links[]` and
`record_tags[]` collections, including empty collections, or one typed Links
read error with no partial fact set. It MUST select only active link and tag
facts under Core 02 §10.4.3, including active endpoint eligibility for links.
It MUST NOT authorize, redact, select report content, expose SQL detail, or
begin, commit, roll back, replace, or detach the caller's transaction.

**Table 7-A0b. Links-owned active link fact members**

| Member | Required source posture |
| --- | --- |
| `record_link_id` | Exact stable link identity. |
| `src_record_id` | Exact canonical source endpoint identity. |
| `dst_record_id` | Exact canonical destination endpoint identity. |
| `link_type` | Exact persisted Core 02 token. |
| `field_key` | Exact nullable canonical field key; null means no owning field. |
| `provenance` | Exact persisted Core 02 token. |
| `confidence` | Exact nullable integer in `0..100`. |
| `owner_user_id` | Exact current owner attribution. |
| `created_by_user_id` | Exact creation attribution. |
| `decided_at` | Exact source timestamp. |
| `created_at` | Exact source timestamp. |

**Table 7-A0c. Links-owned active tag fact members**

| Member | Required source posture |
| --- | --- |
| `record_tag_id` | Exact stable tag-assignment identity. |
| `record_id` | Exact tagged-record identity. |
| `tag_name` | Exact stored display form. |
| `normalized_tag_name` | Exact stored normalized identity. |
| `created_by_user_id` | Exact creation attribution. |
| `created_at` | Exact source timestamp. |
| `updated_at` | Exact source timestamp. |

The reader input already carries `incident_id`; the active fact variants MUST
NOT duplicate it. Active facts MUST NOT contain tombstone members or
link-local `note`, `description`, `comment`, extension maps, storage rows, or
consumer DTOs. Adding a source column MUST NOT add a Reporting value member
unless this NLSpec adopts it.

**REQ-RPT-026c**
The application adapter MUST construct provider output with provider key
`links` and the exact mapping in Table 7-A0d. The adapter MUST issue no SQL,
perform no authorization or redaction, control no transaction, and duplicate
no materializer orchestration. Reporting owns its provider ports and DTOs;
Links owns source fact meaning and errors; only application composition may
import both boundaries. No shared peer DTO package may become a third owner.

**Table 7-A0d. Links-to-Reporting field mapping**

| Source fact | Path | Source family | Content class | Reporting value |
| --- | --- | --- | --- | --- |
| Active link | `/relationships/{record_link_id}` | `record_link` | `derived_analytic` | Every Table 7-A0b member; explicit `deleted_at=null` and `deleted_by_user_id=null`; no `incident_id` or narrative member. |
| Active tag | `/tags/{record_tag_id}` | `record_tag` | `derived_analytic` | Every Table 7-A0c member; explicit `deleted_at=null` and `deleted_by_user_id=null`; no `incident_id`. |

Final field facts MUST sort by exact path ascending. Zero facts MUST produce an
empty provider output rather than an omitted or unknown family. A Links read or
translation failure MUST return one Reporting provider failure with no partial
field or support output; Reporting owns mapping that provider failure into its
existing materialization and render error vocabulary.

**REQ-RPT-026d**
The Links support-reference adapter MUST consider exactly active links whose
`link_type` is `supported_by`, `references_record`, or `attached_evidence`.
Eligible facts MUST be processed by `src_record_id` ascending and then
`dst_record_id` ascending. For each fact, its target path MUST be the non-empty
logical target path contributed for `dst_record_id`; a missing, nil, or empty
logical target MUST fall back to `/record_envelopes/{dst_record_id}`. Each
target MUST be appended to the source record's ordered target list. Repeated
targets MUST remain repeated in this source output; any later support-index
deduplication remains Reporting-owned. Zero eligible links MUST return a
non-nil empty map.

The field-provider and support-reference-provider calls MUST occur while the
Core-owned source-boundary transaction is open. Once their values are admitted
to the immutable snapshot, later Reporting stages MUST obey REQ-RPT-025 and
REQ-RPT-026 and MUST NOT re-read live Links state.

### 7.1.3 Provider identity and fact consistency

**REQ-RPT-026e**
Snapshot materialization MUST carry each source owner's declared record identity
and record type through its field bindings. It MUST NOT reconstruct either from
selector spelling, a display label, or the last segment of a field path. Each
field fact MUST bind to exactly one record fact in that provider output; a
record fact MUST name at least one field, and its field references MUST resolve
to that provider's declared fields with the same source family. Repeated record
identities, overlapping field ownership, missing bindings, and dangling or
foreign field references MUST reject the complete contribution before snapshot
publication. The incident-metadata contribution is not a Core record and MUST
NOT manufacture record identities from incident metadata member names.

Declared subject and Timeline-event field bindings MUST resolve to their exact
source record. Every contributed Host, Identity, Party and entity-mention source
record MUST have exactly one subject declaration with the Table 13-C identity;
every Timeline source record MUST have exactly one Timeline-event declaration.
Those declarations MUST preserve the complete field membership and partition
union of their source record; Timeline-event declarations additionally preserve
its support union. An entity mention MUST declare its host or identity family
from typed source facts, not selector spelling. Frozen snapshot subjects MUST
be derived once from the complete source record and preserve that record's
source binding, display-name selection and partition union. A later field MUST
NOT overwrite the subject with a narrower partition set or another source
reference. Conflicting subject families for one entity-mention source record
MUST reject materialization with `failure_code='token_manifest_invalid'` and
`reason_code='token_subject_not_stable'`. Declared partition and support bindings MUST agree with the
source field metadata, and every field partition/support reference MUST have
its corresponding declaration. Record-level partition and support sets MUST
preserve the union of their fields. This comparison uses reference sets; it
MUST NOT remove repeated targets from the Links source output required by
REQ-RPT-026d. Support targets in another source family are resolved at the
aggregate snapshot boundary, not rejected merely because another provider owns
them. A rejected contribution MUST publish no partial field or record output.
Source classifications MUST be explicit closed content-class tokens. Missing,
unknown or padded tokens MUST NOT be repaired from selector spelling. Snapshot
case-data fields MUST NOT carry raw-blob, opaque-binary or generated-presentation
contributions; these must fail before source metadata could be lost in a
structured projection. Declared assets and generated presentation continue to
use their separate owner contracts. Redaction MUST consume the structured model
only; a legacy flat-field cache or a synthesized source reference is not a valid
substitute for required structured content.

These checks do not replace the complete canonical model, disclosure, source
reference and support reference validation required by §§9–13.

### 7.1.4 Source containers and canonical scalar fields

**REQ-RPT-026f**
An owner-declared record-value object in provider output is a source container,
not a `cartulary.reporting_field.v1` value. Before immutable snapshot publication,
materialization MUST project its immediate named attributes in ascending UTF-8
key order into distinct fields. Each key MUST satisfy `member_name`; empty
objects, nested objects, arrays containing containers and noncanonical scalar
values fail with `export_model_invalid / invalid_scalar_value`. Materialization
MUST NOT recursively flatten containers or serialize them as JSON display text.
The source owner's allowed-member contract continues to limit which attributes
may enter this projection; this rule does not authorize whole-row discovery.

Each projected field uses the declared member name as `field_key`, contiguous
`field_ordinal`, and an exact selector formed by the source container selector,
`/`, and that member name. A directly contributed scalar or scalar array keeps
its declared selector and uses `field_key='value'`. Generated selectors MUST be
unique across the complete snapshot, including collisions with independently
contributed selectors. Every projected value retains its source record binding,
content classification, support references and disclosure labels. Canonical
record and Timeline-event fields retain the same projected values; subjects
retain the complete source record identity and partition union. Record-valued
input MUST NOT survive in a canonical field or be reconstructed by a legacy
reader. Display labels remain null unless an adopted view-schema or derivation
label explicitly supplies one; storage names and selector spelling are not
presentation labels.

Reporting materialization MUST coalesce repeated source support/partition
references into ordered unique arrays without mutating the owner contribution.
An empty canonical reference array is `[]`, never null. Blank or control-bearing
references fail admission; coalescing MUST NOT trim or otherwise reinterpret an
identifier. This does not change REQ-RPT-026d's repeated Links source output.

Before snapshot/render admission, redaction, portability import or recovery,
every repeated field MUST match its section field in every member except
`field_ordinal`, which belongs to its containing list. Comparison uses canonical
JSON value semantics, independent of the decoder's in-memory numeric type.
Record source, support and disclosure reference arrays MUST equal the respective
ordered unions of their field references;
duplicate source ownership and orphaned copies are invalid. Each selected
Timeline source record has exactly one event copy with identical fields, full
source record reference, support references and disclosure labels. Missing,
duplicate or divergent event copies are invalid. A matching object digest does
not waive these consistency checks. No consumer may prefer one conflicting
copy or reconstruct a missing copy during admission.

### 7.1.5 Closed source-field projections

**REQ-RPT-026g**
Source adapters MUST select the closed member sets below before materialization.
Storage schema expansion MUST NOT expand Reporting input. Whole-row conversion,
catalog-driven column discovery and wildcard projection are prohibited for these
source containers. Explicitly selected nullable members remain present as null;
they are not omitted based on their value. Integer tokens MUST reach canonical
admission without conversion through binary floating point. Source semantics,
authorization, incident scoping, deletion filtering and classification remain
with their existing owners; this table defines only their Reporting projection.

| Source container | Complete allowed members |
| --- | --- |
| Record envelope | `record_id`, `record_type`, `created_by_user_id`, `created_at`, `updated_by_user_id`, `updated_at`, `row_version` |
| Timeline event | `record_id`, `capture_state`, `row_version`, `recorded_at`, `edited_at`, `created_by_user_id`, `updated_by_user_id`, `reviewed_by_user_id`, `reviewed_at`, `superseded_by_user_id`, `superseded_at`, `date_entered_text`, `analyst_text`, `mitre_stage_text`, `device_object_text`, `ip_address_text`, `activity_utc_text`, `activity_local_text`, `raw_activity_text`, `activity_synopsis_text`, `data_source_text`, `activity_utc_generated`, `activity_local_generated`, `activity_time_pair_state` |
| Host | `record_id`, `row_version`, `display_name`, `hostname`, `host_state`, `linked_event_count`, `evidence_count`, `location`, `os_platform`, `business_owner`, `criticality`, `containment_status`, `edited_at` |
| Identity | `record_id`, `row_version`, `display_name`, `upn`, `email`, `sam_account_name`, `identity_state`, `linked_event_count`, `evidence_count`, `privilege_level`, `mfa_state`, `reset_status`, `edited_at` |
| Party | `record_id`, `display_name`, `party_kind`, `organization_name`, `role_title`, `primary_email`, `timezone_name`, `external_ref`, `notes`, `created_at`, `updated_at` |
| Task request | `record_id`, `row_version`, `title`, `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `requester_party_id`, `blocked_reason`, `completed_at`, `external_ticket_ref`, `closure_summary`, `decision_record_id`, `linked_record_count`, `updated_at`, `no_owner` |
| Decision | `record_id`, `row_version`, `summary`, `status`, `owner_user_id`, `decision_type`, `decided_at`, `rationale`, `affected_record_count`, `supersedes_record_id`, `updated_at`, `is_superseded` |
| Entity mention | `entity_mention_id`, `source_record_id`, `entity_type`, `source_field_key`, `origin_kind`, `origin_locator`, `raw_text`, `normalized_text`, `resolution_status`, `row_version`, `ordinal`, `created_by_user_id`, `created_at`, `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, `resolution_method` |

Artifact containers contain the common members `record_id`, `row_version`,
`artifact_type`, `title`, `body`, `timestamp_utc`, `updated_at`, `created_at`,
`created_by_user_id`, and `linked_record_count`, plus only the members for their
declared artifact family below. No other artifact families enter this projection.

| Artifact family | Additional allowed members |
| --- | --- |
| Note | None |
| Communication log | `comm_id`, `comm_type`, `audience`, `channel_or_meeting`, `summary`, `next_report_at`, `privilege_tag` |
| Handoff | `handoff_id`, `outgoing_owner_user_id`, `incoming_owner_user_id`, `current_state_summary`, `next_checks`, `acknowledged_at`, `ack_state` |
| Status review | `status_review_id`, `review_owner_user_id`, `current_state_summary`, `active_risks_summary`, `next_report_at` |
| Lesson | `lesson_id`, `owner_user_id`, `closure_state` |
| Finding/hypothesis | `finding_statement`, `finding_kind`, `finding_state`, `finding_owner_user_id`, `finding_confidence_score`, `finding_closed_at`, `finding_updated_at`, `finding_confidence_band` |

Evidence and Links retain their separately defined closed projections. Incident
identity remains in the enclosing immutable model, not an automatically copied
storage column. Active record envelopes do not carry redundant tombstone nulls.
Projection-only search/day-index columns and fields belonging to another
artifact family are excluded even when null. Admission of another member
requires an explicit source-owner and Reporting contract amendment, including
classification, disclosure, type and derivation-identity review. A database
migration alone is insufficient.

## 7.2 Release tuple

**REQ-RPT-027**
A render operation MUST bind to one immutable source tuple with at least the members in Table 7-A. Unknown members in the normalized Reporting-owned tuple are invalid unless a later revision defines an extension container.

**Table 7-A. Release tuple members**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `release_id` | `identifier` | Yes | No | None | Stable release candidate identity from Core release creation. |
| `incident_id` | `identifier` | Yes | No | None | Incident identity from Core route binding. |
| `snapshot_id` | `identifier` | Yes | No | None | Immutable snapshot identity. |
| `snapshot_at` | `timestamp` | Yes | No | None | Source snapshot timestamp. |
| `source_change_set_high_watermark` | `identifier` | Yes | Yes | None | Source-state high-water mark when the snapshot exposes it; otherwise null with `snapshot_boundary_kind` non-null. |
| `snapshot_boundary_kind` | string | Yes | Yes | None | Null only when `source_change_set_high_watermark` is non-null; otherwise names the Core-owned immutable boundary. |
| `render_admitted_at` | `timestamp` | Yes | No | None | Deterministic timestamp fixed by Core render admission. All Reporting-owned generated timestamps that participate in canonical bytes MUST equal this value unless diagnostic-only. |
| `derivation_version` | `identifier` | Yes | No | None | Reporting derivation version; MUST resolve to an adopted `cartulary.reporting_derivation_profile.v2` under REQ-RPT-027a. |
| `template_id` | `identifier` | Yes | No | None | Local template identity. |
| `template_version` | `identifier` | Yes | No | None | Exact template version; `latest` is invalid. |
| `template_manifest_sha256` | `sha256_hex` | Yes | No | None | Digest of canonical template-pack manifest bytes. |
| `composition_id` | `identifier` | Yes | Yes | None | Null only when no composition is bound. Must be non-null exactly when `composition_version` and `composition_sha256` are non-null. |
| `composition_version` | `identifier` | Yes | Yes | None | Exact companion-owned composition version; `latest` is invalid. Null only when no composition is bound. |
| `composition_sha256` | `sha256_hex` | Yes | Yes | None | Digest of `cartulary.report_composition.v1` canonical bytes. Null only when no composition is bound. |
| `redaction_profile_id` | `identifier` | Yes | No | None | Exact redaction profile identity. |
| `redaction_profile_version` | `identifier` | Yes | No | None | Exact redaction profile version. |
| `redaction_profile_sha256` | `sha256_hex` | Yes | No | None | Digest of redaction profile bytes. |
| `release_scope` | string | Yes | No | None | Closed token from §7.3. |
| `recipient_partition_refs[]` | array | Yes | No | None | May be empty only when `release_scope!='external_release'`; omission is invalid; external validation follows REQ-RPT-027b. |
| `graph_projection_refs[]` | array of `source_projection_ref.v2` | Yes | No | `[]` | Exact immutable Graph Projection references available to diagram selection; sorted by `graph_view_id`; duplicate `graph_view_id` values invalid. |
| `output_kind` | string | Yes | No | None | Closed token from §7.4. |
| `output_options` | `cartulary.reporting_render_request_options.v1` | Yes | No | §7.5 defaults before Reporting receives tuple | Normalized object conforming to §7.5. If omitted on a public route, Core 01 MUST materialize defaults before Reporting receives the tuple. |
| `render_environment_profile_id` | `identifier` | Yes | No | None | Exact profile identifier from the template pack and toolchain snapshot. |

**REQ-RPT-027f**
The composition tuple fields `composition_id`, `composition_version`, and `composition_sha256` MUST be all `null` or all non-null. All-null means the render has no composition input, `derive_deck_v1` remains the only deck derivation algorithm, and existing no-composition canonical bytes MUST remain unchanged. A partial-null tuple MUST fail before render output bytes with `error.code='invalid_release_request'`, `failure_code=null`, and `reason_code='composition_tuple_incomplete'`.

When the composition tuple is non-null, Reporting MUST resolve it through the companion-owned Report Composition consumer interface and load exactly the companion-owned `cartulary.report_composition.v1` canonical bytes named by the tuple. An absent composition, including a composition that exists only in a different incident, MUST fail before render output bytes with `reason_code='composition_not_found'`. An absent immutable version under an otherwise valid in-incident composition MUST fail with `reason_code='composition_version_not_found'`. A digest mismatch MUST fail with `error.code='release_render_failed'`, `failure_code='composition_invalid'`, and `reason_code='composition_digest_mismatch'`. A composition whose `template_id` or `template_version` differs from the release tuple MUST fail with `failure_code='composition_invalid'` and `reason_code='composition_template_mismatch'`. A non-null composition tuple is valid only after the Core 01 and `docs/report-composition-nlspec.md` dependencies in §5 are adopted.

**REQ-RPT-027a**
`derivation_version` MUST resolve to exactly one adopted `cartulary.reporting_derivation_profile.v2`. That profile is the versioned owner of every snapshot-to-export-model content-derivation decision that an export-model schema in §9 does not itself fix, and it MUST close each obligation in Table 7-A1 deterministically. The Reporting-owned derivations `derive_section_ordering_key_v1` (REQ-RPT-040a), `derive_display_token_v1` (REQ-RPT-063a), `filter_disclosure_partitions_v1` (REQ-RPT-059a), and `select_timeline_rows_v1` (REQ-RPT-073) are fixed by this NLSpec and MUST NOT be redefined by a derivation profile. An unresolved or unsupported derivation profile MUST fail before render output bytes with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='unsupported_derivation_algorithm'`.

**Table 7-A1. Derivation-profile closure obligations**

| Derivation decision | Required deterministic closure |
| --- | --- |
| Record and timeline-event selection | A predicate over immutable snapshot state fixing which source records and timeline events enter the export model. |
| `field_key` assignment | A mapping from source field identity to stable `field_key`, independent of visible labels, emitting only `member_name` values. |
| `display_label` derivation | Label text or explicit null per §9.3, redaction-safe after the redaction phase. |
| Section expansion | Expansion of template `sections[]` declarations into emitted sections, supplying the `decl_index` and expansion-dimension sort consumed by `derive_section_ordering_key_v1` (REQ-RPT-040a). |
| Support-reference selection | Selection of `support_index[]` entries and per-object `support_refs[]`. |
| Ordinal assignment | Contiguous `field_ordinal` and `block_ordinal` assignment consistent with §9.3. |
| Tokenizable-subject selection | A predicate fixing which Host, Identity, Party, and unresolved mention subjects enter `subjects[]`. At minimum, every subject referenced by a token substitution or by a diagram vertex of family `host`, `identity`, or `party` MUST enter `subjects[]`. |
| Field-to-subject resolution | A deterministic mapping from an export-model field to zero or one `stable_subject_ref`, consumed by `apply_token_backed_redaction_v1`. |
| Missing, null, deleted, and superseded source handling | Exact source-value state, omission, counting, eligibility, and failure behavior for every selected family, consistent with §9.3 and §9.4. |
| Failure behavior | The specific stage, `failure_code`, `reason_code`, and safe-details keys for every unresolved profile input. |

**REQ-RPT-027b**
For `external_release`, every member of `recipient_partition_refs[]` MUST match the `party:{party_partition_segment}` disclosure-partition grammar from REQ-RPT-057a, resolve to a Party record present in the immutable snapshot, and appear in the selected redaction profile's declared allowed `disclosure_partition_refs[]` under Core 01. The set of `recipient_partition_refs[]` MUST exactly equal the `party:*` subset of that profile's declared allowed partitions. These content-bound checks run during cancelable worker preparation against the retained authored policy, without resolving its selector again. Violations MUST fail the admitted Job before render output bytes with `error.code='release_render_failed'`, stage `redaction`, `failure_code='redaction_manifest_invalid'`, and respectively `reason_code='invalid_recipient_partition_ref'`, `reason_code='unknown_recipient_partition'`, or `reason_code='recipient_partition_profile_mismatch'`. They MUST NOT create a release row before a valid preparation exists.

**REQ-RPT-027g**
`graph_projection_refs[]` is the release tuple's complete graph-projection binding surface. It MUST materialize to `[]` when the route request omits graph projections and omission is valid. Every item MUST satisfy Table 15-A and bind an exact immutable projection result. `source_snapshot_id` may equal the Reporting snapshot or the Network Flow owner-declared alternate immutable source boundary. The array MUST sort bytewise ascending by `graph_view_id`, and two items with the same `graph_view_id` are invalid before render output bytes with `error.code='invalid_release_request'`, `failure_code=null`, and `reason_code='graph_projection_ambiguous'`.

A graph-derived template or composition diagram MUST resolve its `source_graph_view_id` against `graph_projection_refs[]` to exactly one item. No match MUST fail with `failure_code='graph_projection_unavailable'` and `reason_code='graph_projection_not_bound'`. More than one match MUST fail with `failure_code='graph_projection_unavailable'` and `reason_code='graph_projection_ambiguous'`. Reporting MUST NOT select a latest result, request projection during render, fall back to a mutable declaration, or substitute a result whose exact tuple differs.

**REQ-RPT-027c**
`cartulary.reporting_derivation_profile.v2` MUST use Table 7-A2. Unknown members
are invalid. The v1 allowed algorithm tokens in Table 7-A2 are exhaustive; a
later revision that adds an algorithm token MUST define the token's inputs,
ordering, output shape, failure behavior, and behavioral validation in the same
revision.

**Table 7-A2. Reporting derivation profile schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_derivation_profile.v2`. |
| `derivation_version` | `identifier` | Yes | No | None | Exact value referenced by the release tuple. |
| `profile_status` | string | Yes | No | None | `adopted_current`, `adopted_deprecated`, or `future_only`; only `adopted_current` may satisfy external-release conformance. |
| `record_selection_algorithm` | string | Yes | No | None | Exact `template_binding_reachable_records_v1` when the composition tuple is all-null; exact `template_and_composition_reachable_records_v1` when the composition tuple is non-null. |
| `timeline_selection_algorithm` | string | Yes | No | None | Exact `template_timeline_sections_v1`; it selects timeline records reachable from timeline section declarations and orders them only through `select_timeline_rows_v1`. |
| `field_key_algorithm` | string | Yes | No | None | Exact `view_schema_member_name_v1`; standardized view-schema field IDs emit their `member_name`, and derived fields emit the `member_name` declared by this profile. |
| `display_label_algorithm` | string | Yes | No | None | Exact `view_schema_label_or_null_v1`; standardized labels are copied when redaction-safe, and absent labels emit explicit `null`. |
| `section_expansion_algorithm` | string | Yes | No | None | Exact `template_sections_expansion_v1`; expansion follows Table 21-C and supplies expansion-dimension sort keys to `derive_section_ordering_key_v1`. |
| `support_reference_algorithm` | string | Yes | No | None | Exact `source_refs_to_support_refs_v1`; every support-bearing object links to one or more Table 9-H1 support refs selected from reachable source refs and evidence/artifact references. |
| `subject_selection_algorithm` | string | Yes | No | None | Exact `host_identity_party_mentions_v1`; it emits every Host, Identity, Party, and unresolved mention subject referenced by token substitution, field-subject resolution, or diagram vertices. |
| `field_subject_resolution_algorithm` | string | Yes | No | None | Exact `single_subject_ref_from_source_path_v1`; it maps a field to one subject only when the field's source path resolves to exactly one selected subject, otherwise zero subjects. |
| `derived_field_keys[]` | array of `member_name` | Yes | No | `[]` | Ordered bytewise ascending; contains every derived field key emitted by this profile. |
| `profile_sha256` | `sha256_hex` | Yes | No | None | SHA-256 of the canonical profile bytes under §10. |

**REQ-RPT-027d**
The v1 derivation algorithms named by Table 7-A2 MUST use only immutable snapshot state, the release tuple, the selected template manifest, and typed owner contracts projected from the adopted owners. Adopted documents govern human fidelity review; algorithms MUST NOT read, stat, hash, or otherwise consume those documents. They MUST NOT read live workbook state, mutable graph projections, renderer output, system time, filesystem order, map iteration order, or network resources. A profile that names any algorithm token outside Table 7-A2 MUST fail before render output bytes with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='unsupported_derivation_algorithm'`.

The `record_selection_algorithm` token MUST match the composition tuple state in Table 7-A2. A profile that names `template_binding_reachable_records_v1` for a non-null composition tuple, or `template_and_composition_reachable_records_v1` for an all-null composition tuple, MUST fail with `failure_code='export_model_invalid'` and `reason_code='unsupported_derivation_algorithm'`.

**REQ-RPT-027e**
The algorithm tokens in `cartulary.reporting_derivation_profile.v2` MUST have the semantics in Table 7-A3. The profile MUST choose source families, fields, section declarations, narrative slots, and diagram declarations only through its declared template manifest and `derived_field_keys[]`; it MUST NOT add an implementation-local selection rule.

**Table 7-A3. V1 derivation algorithm semantics**

| Algorithm token | Required generative behavior |
| --- | --- |
| `template_binding_reachable_records_v1` | Parse every valid template `bindings[]`, section `bindings[]`, narrative-slot binding, timeline declaration, and diagram declaration into source-family and source-record predicates. Exact stable-ID predicates select only the named record. Wildcard predicates select every snapshot record of the addressed source family that has at least one addressed field or relationship, sorted by `(record_type, record_id)`. Relationship predicates also select their endpoint records. Evidence and artifact references reachable only as support material produce support refs but do not become record summaries unless a binding addresses their record family directly. Deleted records are selected only for validation counts and never for rendered bytes. This token is valid only when the release tuple has no composition. |
| `template_and_composition_reachable_records_v1` | Execute `template_binding_reachable_records_v1`, then add every record, timeline event, subject, relationship, support target, diagram vertex, and diagram edge reachable from the digest-bound composition's valid diagram declarations, operation anchors, authored subject placeholders, and authored block insertion targets. Added records use the same source-family predicates, relationship endpoint inclusion, support-ref emission, deleted-record exclusion, and deterministic `(record_type, record_id)` ordering as `template_binding_reachable_records_v1`. A composition reference to a source record, subject, vertex, edge, or support target outside the immutable snapshot or completed projection fails at composition validation or graph adapter validation; it MUST NOT be silently omitted from reachability. This token is valid only when the release tuple has a non-null composition. |
| `template_timeline_sections_v1` | Select only source records addressed by template sections whose `section_kind='timeline'` or by diagram declarations whose `diagram_source_kind='timeline'`. The selected rows are converted to `cartulary.reporting_timeline_event.v1` and ordered only by `select_timeline_rows_v1`; source insertion order is never an input. |
| `view_schema_member_name_v1` | For standardized source fields, `field_key` equals the adopted view-schema `member_name`. For derived fields, `field_key` must appear in `derived_field_keys[]`. If two selected source fields would emit the same `field_key` in one field list, materialization fails with `failure_code='export_model_invalid'` and `reason_code='duplicate_stable_id'` before returning model bytes. It MUST NOT silently discard one source value, assign an ordinal-derived replacement key, or prefer traversal order. Every retained field list must have unique keys at admission, replay and recovery. |
| `view_schema_label_or_null_v1` | `display_label` equals the adopted view-schema display label only when the label is not source content and passes `safe_string`; otherwise it is explicit `null`. Derived fields use the profile-declared label when present and safe, otherwise explicit `null`. Visible workbook labels, localized UI text, and storage-column names are not inputs. |
| `template_sections_expansion_v1` | Iterate template `sections[]` by 1-based declaration index. `expansion='none'` emits one section. `per_party` emits one section per recipient-partition Party sorted by `party_id`; internal scopes use every snapshot Party sorted by `party_id`. `per_subject` emits one section per selected subject sorted by `stable_subject_ref`. `{expansion_label}` is replaced only by the redaction-safe Party label, subject token, or ordinal fallback selected by §15 label rules. |
| `source_refs_to_support_refs_v1` | Emit one source ref for each selected source record or source field that is referenced by a retained object. Emit support refs for source records, evidence metadata, artifact records, timeline events, relationships, and diagram selections using Table 9-H2. Multiple support targets produce multiple support refs. Support refs sort by `support_ref_id`. |
| `host_identity_party_mentions_v1` | Emit one subject for every selected Host, Identity, and Party record and for every selected unresolved host-like or identity-like entity mention that can provide a stable `entity_mention_id`. Subjects referenced by token substitution, field-subject resolution, relationship endpoints, or diagram vertices are mandatory. |
| `single_subject_ref_from_source_path_v1` | A field resolves to one subject when its source path addresses exactly one selected Host, Identity, Party, or unresolved mention. A path addressing no selected subject resolves to zero subjects. A path addressing more than one selected subject resolves to multiple subjects and therefore fails token-backed redaction with `token_subject_not_unique`. |

Reachability MUST precede Core action selection and token allocation. The
template's `allowed_export_model_bindings[]` is an admission allowlist, not a
selection request. Actual section and narrative-slot bindings select addressed
fields, subject to their expansion membership. A field-level binding MUST NOT
implicitly select sibling fields, a parallel envelope or auxiliary sources with
the same record ID. Unselected fields MUST NOT enter retained collections,
redaction entries, token manifests or reveal maps. Nonselection is not a Core
`drop` and MUST NOT generate structural-removal evidence. Exact profile paths
continue to validate against the full admitted candidate inventory, including
currently unselected fields; reachability cannot hide a stale rule.

Required derivation dependencies are explicit: a Party expansion using
`{expansion_label}` selects its participating Party's `display_name`; an authored
subject placeholder selects that subject's `display_name` or unresolved mention
`raw_text`; a Timeline diagram selects the fields of its exact declared event
IDs. A selected relationship selects its structural source fields and endpoint
Core records. A composition block anchor with an exact Core record selects that
record's fields, except an `exclude_block` operation does not expand selection.
Support-only references retain citation metadata without selecting record
values. Each dependency undergoes the same disclosure, lifecycle and Core
action checks as ordinary bindings. Dependency selection never restores a
dropped display value; required unresolved uses fail at their owning validation
stage. A composition dependency alone grants no external superseded opt-in;
an explicitly bound section/diagram placement must also permit it under §14.

Field ordinals inside each emitted field list are assigned after redaction filtering from the deterministic source order: template binding order first, then adopted view-schema field order, then `field_key` bytewise ascending. Block ordinals inside a section are assigned after redaction filtering from section declaration order, then selected record order `(record_type, record_id)`, then block kind order as declared by the template. No ordinal may depend on map iteration order or database row-return order without an explicit owner sort key.

## 7.3 Release-scope vocabulary

**REQ-RPT-028**
The reporting subsystem MUST use the release-scope vocabulary in Table 7-B. It MUST NOT accept local synonyms.

**Table 7-B. Release-scope vocabulary**

| `release_scope` | Approval and redaction consequence |
| --- | --- |
| `internal_draft` | No Reporting-owned approval requirement. The redaction pipeline MUST execute with the selected profile; when no explicit internal profile is selected, Core MUST materialize `internal_passthrough`. `recipient_partition_refs[]` MUST be empty. |
| `internal_review` | Core reviewer approval applies when Core requires it. The redaction pipeline MUST execute with the selected profile; when no explicit internal profile is selected, Core MUST materialize `internal_passthrough`. `recipient_partition_refs[]` MUST be empty. |
| `external_release` | Core external-release approvals apply. `recipient_partition_refs[]` MUST be non-empty. Redaction, token, disclosure-partition, support-reference, local-asset, sandbox, and bundle validation gates are mandatory. |

**REQ-RPT-028a**
Redaction is not optional in any release scope. `internal_passthrough` is the default internal-scope profile and MUST behave as a profile whose `profile_default.action='allow'`, whose `allowed_disclosure_partition_refs[]` is `{public, internal_only} plus every party:{party_partition_segment} present in the snapshot`, whose authored rule arrays are empty, and whose token-backed parameters are absent. Materialization adds the deterministic typed-container allow rules under REQ-RPT-060b without changing field disclosure or default allow behavior. Omission behavior: if an internal-scope request omits `redaction_profile_id`, Core normalizes the exact `internal_passthrough` identity and version before admission, and Reporting derives and seals the materialized profile digest during worker preparation; if Reporting cannot materialize that retained profile, Reporting fails before render output bytes with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='blocked_core_dependency'`. External release has no default profile; omission of any redaction-profile tuple member is a schema failure.

For this coordinated cutover, `cartulary.redaction.tokenized_review` version `5` is an internal-scope profile that permits review of source-evidence subjects through tokenized display values. It MUST preserve the same release-resource shape and sensitive-artifact non-exposure rules as every other profile: token manifests may be bound by digest and persisted in the release bundle, while reveal maps remain internal sensitive artifacts and are not public release resources.

Core's coordinated projection identifies `internal_passthrough` as
`cartulary.redaction.internal@5`; no additional selector alias is admitted.
Both public request members may be omitted together only for internal scopes.
Partial, null and empty tuples remain invalid. The defaulted pair participates
in request normalization. Admission retains the exact authored policy and its
independent digest; cancelable worker preparation seals the complete
snapshot-bound profile and materialized digest under §5. Retained tuples are
never repaired by this default.

## 7.4 Output-kind vocabulary

**REQ-RPT-029**
The Reporting Subsystem MUST accept only the output kinds in Table 7-C for current v1 conformance. It MUST NOT accept case variants, aliases, or future-only names.

**Table 7-C. Current v1 output kinds**

| `output_kind` | Required behavior |
| --- | --- |
| `mermaid` | Generate canonical `.mmd` source using `mermaid_source_serialize_v3` for auto-layout diagrams, reject retained manual-layout diagrams, render SVG when required by §7.5, reject unsupported rendered PNG requests, validate security and bundle requirements. |
| `slidev` | Generate canonical `slides.md` using `slidev_markdown_serialize_v2`, route auto-layout diagrams through Mermaid and manual-layout diagrams through `diagram_layout_svg_serialize_v1`, render PDF for external release, reject unsupported PPTX/PNG requests, validate click steps, security, toolchain, and bundle requirements. |

**REQ-RPT-030**
`markdown`, `html`, and `reenactment` are future-only in this revision. A release request for one of those values, an unknown string, an alias, or a case variant MUST fail before render output bytes with `error.code='invalid_release_request'` and `reason_code='unsupported_output_kind'` unless Core 01 retains the value only as a non-conformant future selector outside this NLSpec's adopted Reporting v1 profile.

## 7.5 Output options

**REQ-RPT-031**
`cartulary.reporting_render_request_options.v1` MUST use the schema in Table 7-D after default materialization.

**Table 7-D. Output options schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_render_request_options.v1`. |
| `source_only` | boolean | No | No | `false` | Valid only for `internal_draft` and `internal_review`; invalid for `external_release`. When `true`, default materialization forces every rendered-output member to `false` per REQ-RPT-032a. |
| `pdf` | boolean | No | No | `true` for `slidev`, `false` for `mermaid` | External `slidev` requires `true`. |
| `svg` | boolean | No | No | `true` for `mermaid`, `false` for `slidev` | External `mermaid` requires `true`. |
| `png` | boolean | No | No | `false` | Current renderer profiles do not support rendered PNG output; `true` fails with `unsupported_output_option`. |
| `pptx` | boolean | No | No | `false` | Current renderer profiles do not support PPTX output; `true` fails with `unsupported_output_option`. |
| `rendered_diagrams` | boolean | No | No | `true` | External releases with diagrams require `true`. |

**REQ-RPT-032**
The output-option conflicts in Table 7-E MUST fail before render output bytes are persisted.

**Table 7-E. Output option conflict mapping**

| Conflict | `error.code` | `failure_code` | `reason_code` |
| --- | --- | --- | --- |
| External release with `source_only=true` | `invalid_release_request` | null | `source_only_external_release_invalid` |
| Requested option unsupported by template | `invalid_release_request` | null | `unsupported_output_option` |
| `pptx=true` for `mermaid` | `invalid_release_request` | null | `unsupported_output_option` |
| `pdf=false` for external `slidev` | `invalid_release_request` | null | `required_output_omitted` |
| `svg=false` for external `mermaid` | `invalid_release_request` | null | `required_output_omitted` |
| `rendered_diagrams=false` for external release containing diagrams | `invalid_release_request` | null | `required_output_omitted` |
| Internal `source_only=true` with any explicit `pdf`, `svg`, `png`, `pptx`, or `rendered_diagrams` set to `true` | `invalid_release_request` | null | `source_only_conflict` |
| Internal `source_only=true` with any retained manual-layout diagram | `invalid_release_request` | null | `source_only_conflict` |

**REQ-RPT-032a**
When `source_only=true` (valid only for `internal_draft` and `internal_review`), default materialization MUST force `pdf`, `svg`, `png`, `pptx`, and `rendered_diagrams` to `false`, and the render bundle MUST contain only source-role files plus the mandatory manifest and validation-role files required by §22. A request that sets `source_only=true` together with any of those members explicitly `true` MUST fail before render output bytes with `error.code='invalid_release_request'` and `reason_code='source_only_conflict'`. Because manual-layout diagrams require a generated SVG dependency rather than a source Mermaid file, a source-only request with any retained manual-layout diagram MUST fail with the same error and reason before bundle bytes persist. `source_only` retains its default `false` when omitted.

# 8. Common scalar contracts and schema notation

**REQ-RPT-033**
A Reporting input received as bytes MUST be decoded as UTF-8 JSON before schema validation. Invalid UTF-8, invalid JSON syntax, duplicate JSON object member names at any depth, a decoded top-level value that is not a JSON object, or an explicit `null` where the schema forbids `null` MUST fail at the earliest attributable validation stage.

**REQ-RPT-034**
Duplicate JSON object member handling MUST NOT depend on parser first-wins, last-wins, insertion-order, or map-overwrite behavior. A duplicate member MUST fail with `reason_code='duplicate_object_member'` and a safe `details.field` when one field is attributable.

**REQ-RPT-035**
Schema tables use `[]` as type notation, not as a literal JSON member-name suffix. A table member written as `sections[]` denotes the JSON member named `sections` whose value is an array.

**REQ-RPT-036**
Unless a narrower scalar contract is defined, Reporting schemas MUST use the scalar contracts in Table 8-A.

**Table 8-A. Common scalar contracts**

| Contract | Definition |
| --- | --- |
| `identifier` | JSON string containing 1 to 128 Unicode scalar values. It MUST NOT contain U+0000, Unicode surrogate code points, C0 controls, C1 controls, leading or trailing Unicode whitespace, `/`, `\`, or `#`. It is compared by exact code point sequence. |
| `member_name` | JSON string matching `[A-Za-z_][A-Za-z0-9_]{0,63}`. |
| `generated_id` | `identifier` with a prefix from Table 10-B, including its trailing `_`, followed by exactly 64 lowercase hexadecimal characters. |
| `timestamp` | JSON string in UTC form `YYYY-MM-DDTHH:MM:SSZ` or `YYYY-MM-DDTHH:MM:SS.ffffffZ`. It MUST be a valid proleptic Gregorian calendar date-time: month `01`–`12`, day `01` through the last day valid for that month and year, hour `00`–`23`, minute `00`–`59`, and second `00`–`59`. Leap seconds and any out-of-range or non-existent calendar value (for example `2026-02-30T00:00:00Z` or `2026-13-01T00:00:00Z`) are invalid and MUST fail at the earliest attributable validation stage with `reason_code='invalid_timestamp_value'`. Generated Reporting timestamps MUST use exactly six fractional digits when the value participates in canonical bytes. |
| `sha256_hex` | JSON string containing exactly 64 lowercase hexadecimal characters. |
| `media_type` | JSON string matching an allowed media type row in §22. Template-declared media types are valid only for `local_asset` rows and only when declared in the template manifest. |
| `export_model_path_v1` | JSON string path using `$` root; `.name` member access where `name` is `member_name`; `[n]` zero-based array index for diagnostic paths only; and `['id']` keyed access into ID-keyed arrays, where `id` is the stable ID with `'` and `\` escaped as `\'` and `\\`. Paths are diagnostic and template-binding identifiers, not JSONPath. |
| `bundle_path` | POSIX-relative UTF-8 path in Unicode NFC with no leading `/`, no empty segment, no `.`, no `..`, no backslash, no NUL, and no duplicate exact byte sequence after NFC validation. |
| `safe_string` | JSON string that has passed redaction for the target scope and contains no C0/C1 controls. LF is valid only in the fields explicitly allowed by REQ-RPT-036a. |
| `finite_integer` | JSON number token matching `0` or `-?[1-9][0-9]*` with mathematical value in `[-9007199254740991, 9007199254740991]`. Decimal-point notation, exponent notation, leading plus sign, leading zeroes, and `-0` are invalid. |
| `positive_integer` | `finite_integer` with mathematical value in `[1, 9007199254740991]`. |

**REQ-RPT-036a**
`safe_string` members MUST be single-line unless this NLSpec explicitly allows LF for that member. LF is allowed only in `speaker_notes` on slide objects and in `fields[].value` strings inside blocks whose `block_kind='paragraph'`. A LF in any other `safe_string` member MUST fail with `failure_code='export_model_invalid'` and `reason_code='invalid_multiline_value'`.

**REQ-RPT-037**
Every schema object defined by this NLSpec is closed. Unknown members are invalid unless the object contains an explicitly named extension container defined by this revision. Omitted optional members MUST materialize to their declared default before canonical serialization when the member participates in hash input. Explicit JSON `null` is valid only for members whose `Nullable` column says `Yes`.

# 9. Canonical reporting export model and child schemas

## 9.1 Top-level export model

**REQ-RPT-038**
`cartulary.reporting_export_model.v7` is the only canonical case-content render input for this subsystem. Templates, Mermaid generation, Slidev generation, render validation, redaction, tokenization, and bundle generation MUST read case content only from this model after materialization.

**REQ-RPT-039**
A reporting export model MUST be a JSON-compatible object with exactly the top-level members in Table 9-A. Unknown top-level members are invalid.

The admitted and post-redaction models MUST use this same closed shape.
The retained model MUST preserve admission identity and timestamps while
reconstructing content collections from selected retained values. A flat
decision list, private source facts or an additional top-level input digest
MUST NOT substitute for or extend this canonical case-content object. Private
decision indexes may retain pre-removal addresses for evidence; case-content
serialization and runtime-asset inspection MUST consume the retained model.
The post-redaction digest covers its complete canonical bytes under the export
model schema domain. The independently bound admitted-input digest remains
distinct. Validation summaries MUST describe the reconstructed collections.

**Table 9-A. `cartulary.reporting_export_model.v7` top-level schema**

| Member | Type | Required | Nullable | Default | Ordering and notes |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_export_model.v7`. |
| `export_model_id` | `generated_id` | Yes | No | None | Generated under Table 10-B. |
| `snapshot_model_id` | `generated_id` | Yes | No | None | Exact immutable materialization identity from REQ-RPT-039a. |
| `release_id` | `identifier` | Yes | Yes | None | Non-null for release renders; null for previews. |
| `preview_attempt_id` | `identifier` | Yes | Yes | None | Non-null only for an admitted internal-draft preview; null for release renders. |
| `reference_packs` | `reference_pack_set_binding.v1` | Yes | No | None | Exact set and immutable provenance captured and pinned in snapshot admission. |
| `incident_id` | `identifier` | Yes | No | None | Copied from release tuple. |
| `snapshot_id` | `identifier` | Yes | No | None | Copied from release tuple. |
| `snapshot_at` | `timestamp` | Yes | No | None | Copied from source snapshot. |
| `render_admitted_at` | `timestamp` | Yes | No | None | Deterministic timestamp from Core release admission. |
| `source_change_set_high_watermark` | `identifier` | Yes | Yes | None | Null only when `snapshot_boundary_kind` is non-null. |
| `snapshot_boundary_kind` | string | Yes | Yes | None | Names Core immutable boundary when high watermark is null. |
| `derivation_version` | `identifier` | Yes | No | None | Reporting derivation version. |
| `export_model_created_at` | `timestamp` | Yes | No | None | MUST equal `render_admitted_at`; hash-participating. |
| `export_model_generator_id` | `identifier` | Yes | No | None | Stable generator ID. |
| `export_model_generator_version` | `identifier` | Yes | No | None | Exact generator version. |
| `release_scope` | string | Yes | No | None | Token from §7.3. |
| `recipient_partition_refs[]` | array | Yes | No | None | Ordered by §12.2. |
| `sections[]` | array of `cartulary.reporting_section.v2` | Yes | No | `[]` invalid for external release | Ordered by `ordering_key`, then `section_id`; empty external-release sections fail with `reason_code='empty_export_model_sections'`. |
| `records[]` | array of `cartulary.reporting_record_summary.v1` | Yes | No | `[]` | Ordered by `record_type`, then `record_id`. |
| `relationships[]` | array of `cartulary.reporting_relationship_summary.v1` | Yes | No | `[]` | Ordered by `relationship_id`. |
| `timeline_events[]` | array of `cartulary.reporting_timeline_event.v1` | Yes | No | `[]` | Ordered by §14.2. |
| `subjects[]` | array of `cartulary.tokenizable_subject.v1` | Yes | No | `[]` | Ordered by `stable_subject_ref`. |
| `diagrams[]` | array of `cartulary.reporting_diagram.v1` | Yes | No | `[]` | Ordered by `diagram_id`. |
| `assets[]` | array of `cartulary.reporting_asset_declaration.v1` | Yes | No | `[]` | Ordered by `bundle_path`, then `asset_id`. |
| `presentation_texts[]` | array of `cartulary.reporting_presentation_text.v1` | Yes | No | `[]` | Ordered by exact `authored_text_id`; empty at render admission, derived only from the independently bound composition by REQ-RPT-087f. |
| `source_index[]` | array of Table 9-H objects | Yes | No | `[]` | Snapshot-bound source identities, strictly ordered by `source_ref_id`. |
| `support_index[]` | array of `cartulary.reporting_support_ref.v1` | Yes | No | `[]` | Ordered by `support_ref_id`. |
| `validation_summary` | `cartulary.reporting_export_model_validation.v1` | Yes | No | None | Export-model-local validation summary per Table 9-I (REQ-RPT-046a); distinct from the render validation summary. |

## 9.2 Section object

**REQ-RPT-039a**
The release/preview identity union changes adopted artifact schemas. The slide-deck schema advances to v2; the token and reveal artifacts advance to v3 with the subject/use ownership correction in §13, and the redaction manifest advances to v4 with distinct structural-removal evidence. Their retired variants are not admitted by this cutover. The new snapshot materialization and render-identity projections use previously unpublished v1 identifiers.

The snapshot worker MUST retain `cartulary.reporting_snapshot_capture.v1` from the single consistent capture defined by Core 01 REQ-01-477, using the exact snapshot identity allocated by its admitted Job and the Reference Data binding pinned at request admission. Job admission and immutable source capture are separate durable steps. The closed `cartulary.reporting_snapshot_admission.v1` retains the exact Job, incident, actor, transaction, admission timestamp, nullable requested boundary and Reference Pack binding under a schema-domain digest. The complete capture payload retains its own timestamp, the admission timestamp and requested boundary, exact source-boundary bytes, source capture and schema-domain digest. A retained capture MUST match the independently retained admission and Jobs request identity. Unknown, omitted, altered or retired admission input fails closed. Capture is cancelable work; completion MUST never reread current source or replace an already retained capture. Validation of retained case-content bytes MUST complete before acquiring the Jobs transition lock. Snapshot publication MUST use a short-lived, owner-created validated candidate containing only bounded metadata and MUST recheck that the exact retained payload version validated by that candidate is still current. A changed payload invalidates the candidate even if its embedded digest is unchanged or its former bytes are restored. Candidate validation remains cancelable and MUST drain before a terminal cancellation is published. The final transaction rechecks the independent admission and publishes only snapshot metadata plus the Jobs outcome; it MUST NOT decode, hash or materialize the full capture under the Jobs lock. Recovery independently validates retained capture bytes and their snapshot digest.

The capture has exactly `schema_id`, `snapshot_id`, `incident_id`, `snapshot_at`, `source_change_set_high_watermark`, `derivation_version`, `reference_packs`, `incident`, `source_facts`, `fields`, and `timeline_sort`. `incident` contains exactly `id`, `title`, nullable `description`, `status`, nullable `severity`, nullable `tlp`, nullable `current_phase`, and positive integer `version`. Each field retains exactly `source_record_id`, `source_record_type`, `path`, `content_class`, `source_family`, required `value` (including explicit null), `disclosure_partition_refs`, and `support_refs`. Fields sort by unique source path; Timeline sort facts sort by unique record ID and cover the captured Timeline roster exactly. Source selectors are inert owner-port identifiers and never filesystem paths. Support selectors must resolve to captured source identities. Source facts, classification, content admission, disclosure labels, and source-owner closed-member selection retain their existing requirements. No raw binary, generated presentation, or undeclared private source value is admitted.

The capture retains source values once; sections, record summaries and Timeline presentation are derived projections. The snapshot record independently binds `source_capture_sha256` to its create Job's retained capture. The API exposes that digest, not a digest falsely described as an already materialized render model. The retired full-model snapshot storage and Job formats have no compatibility reader. Render preparation derives `cartulary.reporting_snapshot_model.v6` from the retained capture and then binds the release or preview identity. This projection remains deterministic and subject to all model validation below. Its top-level members are Table 9-A with `schema_id` replaced by that snapshot schema, and with `export_model_id`, `release_id`, `preview_attempt_id`, `render_admitted_at`, `export_model_created_at`, `release_scope`, `recipient_partition_refs`, and `presentation_texts` removed. `snapshot_model_id` and `source_facts` are required. `source_facts` is a closed `cartulary.reporting_source_facts.v2` object with required non-null arrays `parties`, `assignments`, and `deleted_records`. Each Party fact has exactly `party_id`, `party_partition_segment`, and boolean `public_directory_eligible`; Core 02 REQ-02-143a.2 owns their source meaning. Parties are ordered by exact Party ID, and assignments contain the complete Table 12-A objects ordered by exact assignment ID. Duplicate identities, duplicate active assignment tuples, missing Party facts, unknown members and facts whose endpoints do not match the snapshot fail materialization. Each deletion fact has exactly `record_id` and the closed Core `record_type`; the Records owner collects the complete incident-scoped deleted-identity roster in the same snapshot transaction. It contains no deleted values, deletion actors or deletion timestamps. Deletion facts are strictly ordered by record ID, reject duplicate or invalid identities and types, and MUST NOT overlap Core source content. Only the Records provider may contribute them. Empty collections are explicit arrays. These private source facts are frozen at the consistent snapshot capture and retained in release and preview Job input with their schema-domain digest; replay additionally compares them to the independently retained snapshot. They MUST NOT be included in a distributable post-redaction ExportModel, native renderer input or ordinary diagnostics. Historical v1 snapshots and jobs without these admitted facts are unsupported; preflight rejects before mutation rather than inferring facts. All remaining nested schemas, required nullable members and ordering rules are unchanged. Unknown members are invalid. Every nested `source_snapshot_id`, including support references and relationship endpoints, MUST equal the allocated snapshot ID. Completion MUST NOT repair an empty reference or recalculate materialization identity.

Snapshot Job failures MAY include closed, value-free `error_summary.details`: `snapshot_stage` is one of `request_validation`, `source_capture`, `model_materialization`, `model_serialization`, or `capture_persistence`; `failure_kind` is one of `internal`, `canceled`, `deadline`, `storage_limit`, `storage_constraint`, `transaction_conflict`, `storage`, or `model_validation`. When serialization completed, `canonical_input_bytes` MAY report its exact nonnegative byte count. Only `model_validation` additionally includes the already owned `failure_code` and `reason_code` pair. Unknown stages or diagnoses MUST NOT expose raw error text, SQL, source values, paths or database details. These details do not change terminal status, retry policy or the public error-code vocabulary.


Frozen snapshot section trees MUST already have explicit canonical section and
block identities, contiguous ordinals, correct parent links, ordered reference
arrays and complete child arrays. Container source, support and disclosure
references are the exact union of their admitted descendants. A source snapshot
MUST NOT assert generated aggregate proof or excluded-field
claims. Those belong to later Reporting derivation. Admission, replay and
recovery reject contradictory trees; post-redaction reconstruction MUST NOT
repair retained source history. After selection, reconstruction uses only the
selected retained field values, drops empty containers, recomputes ordinals and
structural IDs, and recomputes each container's references and class counts.
Selected-rule evidence continues to address the pre-removal model.

Release worker preparation MUST derive a distinct Table 9-A model from the exact retained source capture, preserving source fields, `snapshot_model_id`, and the complete Reference Pack binding. It MUST allocate the release identity before rendering, freeze one `render_admitted_at`, set `export_model_created_at` to that instant, and derive `export_model_id` from Table 10-B. Its immutable recipe and preparation seal bind the exact derived model and its schema-domain digest without retaining another expanded model. Successful and prepared failed render finalization MUST use that same release identity. Restart, participant invocation, redaction, rendering, rerendering and completion MUST NOT substitute execution wall time or the snapshot's capture time. A new render admission gets its own identity; replay of one admitted Job preserves its identity and bytes.

An authoritative composition preview uses the companion's exact `preview_attempt_id` and admission instant. Exactly one of `release_id` and `preview_attempt_id` MUST be non-null; both members MUST appear. Preview-generated IDs use Table 10-B's separate preview domains. Preview identity is admitted only for `internal_draft` and MUST NOT authorize external release, approval or release publication. Generated artifacts carry the same required nullable identity pair; tables naming a required release identity apply to release artifacts, while preview artifacts substitute the required preview identity and explicit null release identity. The render model, redacted model, redaction/token/reveal manifests, deck and bundle retain the exact snapshot ID and identity pair. Their generated timestamps follow §10, including the required `created_at` on retained toolchain and validation objects and `bundle_created_at` on the bundle.

Local admission captures time at microsecond precision once. Input with more than six fractional digits is rejected at subsequent canonical boundaries; consumers MUST NOT round, truncate, or replace received or retained time values. Snapshot capture is source time and is distinct from render admission. A snapshot has no generated render timestamps.

For native retained output, internally consistent hashes do not establish
fidelity to admission. Candidate validation, replay and recovery MUST recompute canonical
presentation models and generated source from the independently retained recipe and capture,
profile, template, composition and output options, and compare the exact bytes
and required member identities. This pure comparison MUST NOT invoke a renderer,
read live workbook state or substitute the current machine's toolchain for the
recorded one. Missing, renamed, extra or changed presentation members, and
rendered artifacts not authorized by the admitted options, fail validation even
if every outer digest has been recomputed. Native output must identify its
supported renderer contract; a retired producer cannot be silently interpreted
as current. This is distinct from portable historical artifact integrity, which
retains its own provenance and admission requirements.


**REQ-RPT-040**
`cartulary.reporting_section.v2` MUST use Table 9-B.

**Table 9-B. Section object schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_section.v2`. |
| `section_id` | `identifier` | Yes | No | None | Structural ID from REQ-RPT-040b; unique in export model. |
| `section_kind` | string | Yes | No | None | `timeline`, `entity_summary`, `evidence_summary`, `relationship_summary`, `diagram`, `narrative`, `appendix`, or `validation`. |
| `title` | `safe_string` | Yes | No | None | Redaction-safe after redaction phase. |
| `ordering_key` | string | Yes | No | None | Canonical lexical sort key generated by `derive_section_ordering_key_v1` (REQ-RPT-040a). |
| `blocks[]` | array of block objects | Yes | No | `[]` | Empty valid only for `validation` sections with no findings. |
| `source_refs[]` | array of `source_ref_id` | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array of `support_ref_id` | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |
| `content_class_summary` | object per Table 9-J | Yes | No | None | Closed content-class count object (REQ-RPT-046a). |
| `section_validation` | object per Table 9-K | Yes | No | None | Closed section-local validation summary (REQ-RPT-046a). |

**REQ-RPT-040a**
`ordering_key` MUST be produced by `derive_section_ordering_key_v1` so that section order is fully determined by the template manifest and export-model content rather than by generator iteration order. The algorithm assigns, for each emitted section:

- singleton section from the template `sections[]` declaration at 1-based index `i`: `ordering_key = zeropad4(i)`;
- one instance of a section expanded from the declaration at index `i`: `ordering_key = zeropad4(i) + "." + zeropad4(j)`, where `j` is the 1-based position of the instance in the ascending canonical sort of that declaration's expansion dimension (for example ascending `party_id` or `stable_subject_ref`);
- a Reporting-appended section not derived from a template declaration (for example `section_kind='validation'`): `ordering_key = "9999." + zeropad4(k)`, where `k` is the 1-based appended order.

`zeropad4` left-pads a decimal integer to four digits; the §25 `sections.count` hard limit guarantees four digits suffice. Sections are ordered by `ordering_key` bytewise ascending, then by `section_id`, per Table 9-A. Two conforming implementations given the same template manifest and export-model content MUST produce identical `ordering_key` values.

**REQ-RPT-040b**
Reporting structural IDs MUST be derived from post-redaction ordinal paths, not hashes:

- `section_id = 'sec-' + ordering_key`.
- Top-level `block_id = section_id + '.b' + zeropad4(block_ordinal)`.
- Nested `block_id = parent_block_id + '.b' + zeropad4(block_ordinal)`.
- Base slide IDs use `slide_id = 'sld-' + zeropad4(slide_ordinal)`; chunk slides use the `__chunk_` suffix in REQ-RPT-075.
- `diagram_id` equals the template manifest diagram declaration `decl_id` verbatim.

Structural IDs are assigned after disclosure filtering and ordinal recomputation. Two conforming implementations given the same post-redaction model and template manifest MUST produce identical structural IDs.

## 9.3 Block and field objects

**REQ-RPT-041**
`cartulary.reporting_block.v3` MUST use Table 9-C.

**Table 9-C. Block object schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_block.v3`. |
| `block_id` | `identifier` | Yes | No | None | Structural ID from REQ-RPT-040b; unique in export model. |
| `block_kind` | string | Yes | No | None | `paragraph`, `bullet_list`, `table`, `metric`, `timeline_rows`, `diagram_ref`, `asset_ref`, `speaker_note`, or `overflow_summary`. |
| `block_ordinal` | `finite_integer` | Yes | No | None | Starts at `1` within parent container, no gaps. |
| `parent_block_id` | `identifier` | Yes | Yes | None | Null for top-level blocks. |
| `content_class` | string | Yes | No | None | `source_evidence`, `derived_analytic`, `curated_narrative`, `working_material`, `presentation_text`, `support_reference`, `validation`, or `template_boilerplate`; admission by origin follows REQ-RPT-041a. |
| `fields[]` | array of field objects | Yes | No | `[]` | Ordered by `field_ordinal`. |
| `children[]` | array of block objects | Yes | No | `[]` | Ordered by `block_ordinal`. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-041a — Classification and origin**
Source materialization MUST preserve the four Core 01 REQ-01-377 classifications
verbatim: `source_evidence`, `derived_analytic`, `curated_narrative`, and
`working_material`. Snapshot case-data blocks admit only those four tokens.
Release scope, template choice, rendering, or summary generation MUST NOT
reclassify them. The former `case_fact` and `derived_summary` tokens are retired,
not aliases. Source owners cannot contribute the four generated categories.

Reporting-generated `presentation_text`, `support_reference`, `validation`, and
`template_boilerplate` remain distinct closed categories. Their eligibility is
defined by the respective derivation and security owners: author-declared text
uses REQ-RPT-087f, support references use REQ-RPT-046, diagnostics use §23, and
static template text uses §21. Copying or transforming case content into a
generated block MUST first satisfy the source classification, support,
partition and redaction rules; changing its category cannot authorize a value.
Authored presentation prose MUST NOT be relabeled `curated_narrative` or given
inferred supporting evidence. Profile content-class selectors admit exactly
the eight Table 9-C tokens; missing, padded, retired and unknown tokens are
invalid before rendering.

The changed block vocabulary and section summary use independent v2 schemas.
Retained v1 sections/blocks are rejected by the pre-production cutover. There is
no reader that guesses a classification or converts an old summary. Table 9-J
counts every top-level and nested block once under its actual classification;
zero counts are explicit. Snapshot admission checks these counts against frozen
blocks. After filtering or derivation, Reporting recomputes the counts for the
resulting model under REQ-RPT-059b, never by counting fields or assuming a class.

**REQ-RPT-042**
`cartulary.reporting_field.v1` MUST use Table 9-D.

**Table 9-D. Field object schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_field.v1`. |
| `field_key` | `member_name` | Yes | No | None | Stable field key from view schema or Reporting derivation. |
| `display_label` | `safe_string` | Yes | Yes | None | Null when no label is emitted. |
| `field_ordinal` | `finite_integer` | Yes | No | None | Starts at `1` within parent field list, no gaps. |
| `source_value_state` | string | Yes | No | None | `present`, `missing`, `null`, `unavailable`, `withheld`, `superseded`, or `derived`. `deleted` is a source input state but is invalid in a retained field. |
| `redacted_value_state` | string | Yes | No | None | `unchanged`, `allowed`, `masked`, `truncated`, `stubbed`, or `tokenized`. `dropped` and `blocked` are redaction-manifest outcomes, not retained field states. |
| `value` | JSON scalar or array | Yes | Yes | None | Post-redaction value or null when state allows null. |
| `raw_value_sha256` | `sha256_hex` | Yes | Yes | None | MUST be `null` in current v1 canonical export models. Raw-value digests are forbidden in external release bundles. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-042a**
Retained field values MUST satisfy Table 9-D1. A retained canonical export model MUST NOT contain a field whose `redacted_value_state` is `dropped` or `blocked`; those outcomes appear only in `cartulary.redaction_manifest.v5`. If a redaction action removes a field, the field is absent after ordinal recomputation under REQ-RPT-059b and the redaction manifest records the removal. In current v1, `raw_value_sha256` MUST be `null` for every retained field in every release scope; a later revision that emits raw-value digests MUST define a separate internal-only artifact and MUST NOT place those digests in an external bundle.

**Table 9-D1. Retained field value-state matrix**

| `source_value_state` | Allowed retained `redacted_value_state` | `value` rule | `raw_value_sha256` rule |
| --- | --- | --- | --- |
| `present` | `unchanged`, `allowed` | Non-null JSON scalar or array after `safe_string` validation for strings. | Always `null` in v1. |
| `present` | `masked`, `truncated`, `stubbed`, `tokenized` | Non-null redaction literal or display token produced by the selected rule; arrays retain original element order after redacted elements are removed. | Always `null` in v1. |
| `null` | `unchanged`, `allowed` | Explicit JSON `null`. | Always `null` in v1. |
| `missing` | `unchanged` | Explicit JSON `null`; `display_label` remains non-null only when the template requires the field label. | Always `null` in v1. |
| `unavailable` | `unchanged`, `stubbed` | `null` for `unchanged`; non-null safe placeholder for `stubbed`. | Always `null` in v1. |
| `withheld` | `masked`, `stubbed`, `tokenized` | Non-null safe placeholder or display token; if no safe placeholder is selected, the field MUST be dropped and recorded only in the redaction manifest. | Always `null` in v1. |
| `superseded` | `unchanged`, `allowed`, `masked`, `truncated`, `stubbed`, `tokenized` | Same value rule as `present`, but only after the source record is admitted by REQ-RPT-043a. If superseded-record admission is not satisfied, the field MUST be absent after filtering and recorded only in the redaction manifest. | Always `null` in v1. |
| `derived` | `unchanged`, `allowed`, `masked`, `truncated`, `stubbed`, `tokenized` | Deterministic derived scalar or array after the selected redaction action. | Always `null` in v1. |

**REQ-RPT-042b**
Source input state MUST be mapped before redaction using Table 9-D2. This mapping is part of `cartulary.reporting_derivation_profile.v2`; an implementation MUST NOT infer an alternate state from visible labels, renderer output, database nullability, or UI omission. If the derivation profile cannot classify a selected value into one Table 9-D2 row, materialization fails with `failure_code='export_model_invalid'` and `reason_code='invalid_scalar_value'`.

**Table 9-D2. Source input state matrix**

| Source input state | Retained field state | Omission and validation behavior |
| --- | --- | --- |
| Source field is present with a non-null value | `present` | Retain when the field survives disclosure filtering and redaction. |
| Source field is present with JSON null | `null` | Retain explicit JSON `null` only when the template binding permits a null display value; otherwise omit the field and recompute ordinals. |
| Source field is absent from the selected source record or projection | `missing` | Retain explicit JSON `null` only when the template declares the field label or slot as required; otherwise omit the field and recompute ordinals. |
| Source owner cannot expose the value at the immutable snapshot boundary | `unavailable` | Retain only for internal scopes or when a selected redaction rule emits a safe stub; otherwise fail with `redaction_profile_invalid` if required by the template. |
| Source value exists but is not releasable to the selected recipient partition | `withheld` | Apply Core redaction rule selection. If the selected action cannot emit a recipient-safe value, drop the field and record a redaction-manifest entry. |
| Source record has `deleted_state='deleted'` | none | A deleted source record MUST NOT produce retained fields. Forced inclusion fails with `deleted_record_not_releasable`. |
| Source record has `deleted_state='superseded'` | `superseded` | Internal scopes may retain it. External release excludes it unless both the template and redaction profile explicitly admit superseded records under REQ-RPT-043a. |
| Value is computed solely by the Reporting derivation profile | `derived` | Retain only when every input to the derivation is selected, deterministic, and eligible after disclosure filtering. |

## 9.4 Record, relationship, asset, and source-reference schemas

**REQ-RPT-043**
`cartulary.reporting_record_summary.v1` MUST use Table 9-E.

**Table 9-E. Record summary schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_record_summary.v1`. |
| `record_id` | `identifier` | Yes | No | None | Source record envelope identity. |
| `record_type` | string | Yes | No | None | Core record type token; `post` invalid in this revision. |
| `source_record_ref` | `source_record_ref.v1` | Yes | No | None | Closed source-record reference from Table 9-F1. |
| `display_name` | `safe_string` | Yes | Yes | None | Null when no release-safe label is emitted. |
| `deleted_state` | string | Yes | No | None | `active`, `deleted`, or `superseded`. Release eligibility per REQ-RPT-043a. |
| `fields[]` | array | Yes | No | `[]` | Field objects ordered by `field_ordinal`. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-043b**
The `records[]` collection MUST contain exactly one summary per selected Core
record envelope identity. The closed Core types are `timeline_event`, `host`,
`identity`, `party`, `indicator`, `artifact`, `task_request`, `decision`,
`evidence`, and `assessment`. Artifact source families (`note`,
`finding_hypothesis`, `comm_log`, `handoff`, `status_review`, `lesson`) retain
their source-family provenance and MUST declare Core type `artifact`.
`record_envelope` identifies the generic source projection, not a Core type;
its declared type MUST equal its source-owned `record_type` value. Domain and
envelope projections of the same ID MUST agree on Core type. Distinct domain
families claiming that ID MUST reject materialization with
`export_model_invalid` / `duplicate_stable_id`.

When both projections are present, the domain projection supplies the record
summary's fields and display name; otherwise the envelope projection supplies
its metadata. Envelope and domain fields MUST remain separately bound in
sections and `source_index`, without merging similarly named fields or changing
their classification. Display selection MUST operate on the complete assembled
source group, not the first provider field encountered. Entity mentions, record
links and tags MUST retain source-index anchors and source fields but MUST NOT
be emitted as Core record summaries. Subject, support and disclosure processing
MUST use their explicit source bindings, including auxiliary groups. It MUST NOT
depend on auxiliary objects appearing in `records[]` or conflate their IDs with
Core IDs. Frozen admission MUST reject missing, duplicate, substituted or extra
Core summaries against the complete source-bound projection.

After redaction, a Core summary MUST retain only its selected fields, with
contiguous field ordinals and recomputed reference and disclosure aggregates.
Its display name MUST be selected from those retained values using the same
source-family display rule. A removed display field MUST NOT be restored from
the admitted summary. A summary with no retained fields MUST be removed;
surviving generic-envelope or auxiliary fields MUST NOT resurrect it.

**REQ-RPT-043a**
A record summary's `deleted_state` MUST govern release eligibility as in Table 9-E1. `active` records are eligible in every scope. A `superseded` record is eligible for internal scopes; for `external_release` it MUST be excluded by default and is included only when the template section explicitly opts in and the redaction profile permits it under the §5.2 Core 04 companion edit. A `deleted` record MUST NOT appear in any rendered output bytes or in any bundle file that carries case content; it can be counted only in export-model-local validation summaries. If a template binding or a required section forces a `deleted` record into external-release output, materialization MUST fail with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='deleted_record_not_releasable'`.

Lifecycle is materialized from explicit source-owner state. Timeline uses its
closed `capture_state`: `rough`, `enriched` and `reviewed` are active;
`superseded` is superseded. Decision uses boolean `is_superseded`, including
executed decisions whose status remains executed. Missing or invalid state
MUST fail admission rather than default to active. A Timeline or Decision
summary requires its domain source projection; a generic envelope alone cannot
establish its lifecycle. Other live source-owner projections are active; deleted
Core identities are retained only in the private Records roster of REQ-RPT-039a.
Every non-null source field belonging to a superseded Core record, including
its parallel envelope projection, MUST have `source_value_state='superseded'`.
Null and missing states retain their meaning. Frozen admission MUST reject
inconsistent summary or field states without repairing them.

Eligibility MUST precede Core rule selection. Exact path rules bind against
the full admitted candidate inventory; a valid rule bound to a lifecycle-excluded
candidate remains inactive, including an admitted nullable-field rule. Unknown
or unsupported bindings still fail. Excluded candidates cannot be
restored by a profile rule and MUST NOT be represented as fabricated Core drop
decisions or redaction structural removals. Reference-bearing fields inherit
the lifecycle restriction of their explicit support target; link endpoints,
mention source/resolved records, and tag record bindings inherit the strongest
restriction of those Core dependencies. Restrictions propagate transitively
through frozen reference chains and cycles using the same dependency graph as
disclosure labels, without converting lifecycle states to disclosure tokens.
Auxiliary IDs MUST NOT be conflated
with Core IDs. An external superseded candidate requires both profile permission
and a matching opted-in template declaration. Selection in one declaration
MUST NOT grant it to another declaration. Exact bindings to a deleted record
or Timeline identity, and composition operations forcing content at a deleted
record anchor or Timeline diagram selection, MUST fail with
`deleted_record_not_releasable`. An exclusion
operation supplies no inclusion authority. No lifecycle decision may query
current mutable state during replay. A template diagram inherits its owning
section's opt-in; a composition diagram also requires every explicit insertion
section to opt in. An unplaced composition diagram has no such permission.
Derived diagram selection and the final post-composition source-bound slide
content MUST enforce these restrictions before any renderer receives bytes.

**Table 9-E1. Record release eligibility by `deleted_state`**

| `deleted_state` | `internal_draft` / `internal_review` | `external_release` |
| --- | --- | --- |
| `active` | Eligible. | Eligible. |
| `superseded` | Eligible. | Excluded by default; included only on explicit template opt-in permitted by the redaction profile. |
| `deleted` | Validation counts only; not in rendered bytes. | Not in rendered bytes; forced inclusion fails with `deleted_record_not_releasable`. |

**REQ-RPT-044**
`cartulary.reporting_relationship_summary.v1` MUST use Table 9-F.

**Table 9-F. Relationship summary schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_relationship_summary.v1`. |
| `relationship_id` | `identifier` | Yes | No | None | Unique in export model. |
| `relationship_kind` | string | Yes | No | None | Core link kind or Reporting derivation kind. |
| `src_record_ref` | `relationship_endpoint_ref.v1` | Yes | No | None | Closed source endpoint reference from Table 9-F2. |
| `dst_record_ref` | `relationship_endpoint_ref.v1` | Yes | No | None | Closed destination endpoint reference from Table 9-F2. |
| `direction` | string | Yes | No | None | `directed`, `undirected`, or `bidirectional`. |
| `confidence` | string | Yes | Yes | None | Null when source does not expose confidence. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-044a**
`source_record_ref.v1` and `relationship_endpoint_ref.v1` MUST use Tables 9-F1 and 9-F2. Unknown members are invalid. Every identifier must resolve inside the same export model and immutable snapshot. A missing, duplicate, cross-snapshot, or source-family-invalid reference MUST fail export-model validation with `failure_code='export_model_invalid'` and respectively `reason_code='dangling_source_ref'`, `reason_code='duplicate_stable_id'`, `reason_code='snapshot_ref_mismatch'`, or `reason_code='unsupported_source_family'`.

**Table 9-F1. `source_record_ref.v1` schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `source_record_ref.v1`. |
| `source_family` | string | Yes | No | None | Core source family token; `post` invalid in this revision. |
| `source_record_id` | `identifier` | Yes | No | None | Source record envelope identity in the immutable snapshot. |
| `source_snapshot_id` | `identifier` | Yes | No | None | Must equal export-model `snapshot_id`. |
| `source_ref_id` | `identifier` | Yes | Yes | None | Null only when no Table 9-H source ref was emitted for the record; non-null values MUST resolve to one source ref. |

**Table 9-F2. `relationship_endpoint_ref.v1` schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `relationship_endpoint_ref.v1`. |
| `endpoint_role` | string | Yes | No | None | `src` or `dst`; must match the containing relationship member. |
| `source_record_ref` | `source_record_ref.v1` | Yes | No | None | Endpoint record reference. |
| `display_ref` | `safe_string` | Yes | Yes | None | Redaction-safe endpoint label, or null when the endpoint has no releasable label. |
| `stable_subject_ref` | string | Yes | Yes | None | Non-null only when the endpoint resolves to one Table 13-B tokenizable subject. |

**REQ-RPT-045**
Asset declaration objects MUST use Table 9-G.

**Table 9-G. Asset declaration schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_asset_declaration.v1`. |
| `asset_id` | `identifier` | Yes | No | None | Unique in export model. |
| `asset_kind` | string | Yes | No | None | `local_asset`, `local_theme`, `rendered_diagram`, `rendered_slide`, or `source_asset`. |
| `bundle_path` | `bundle_path` | Yes | No | None | Must match role/path matrix in §22 when bundle-emitted. |
| `media_type` | `media_type` | Yes | No | None | Must be allowed for role. |
| `sha256` | `sha256_hex` | Yes | No | None | Raw file bytes digest. |
| `byte_size` | `finite_integer` | Yes | No | None | Exact raw byte count. |
| `declared_by` | string | Yes | No | None | `template_pack`, `renderer`, `mermaid`, `slidev`, or `reporting_validation`. |
| `required_for_release` | boolean | Yes | No | None | True when omission invalidates release. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty unless asset is validation-only and contains no case content. |

**REQ-RPT-045a**
Asset declarations are closed objects. Unknown members, duplicate `asset_id`, duplicate normalized `bundle_path`, negative `byte_size`, and `sha256` mismatches MUST fail with `failure_code='export_model_invalid'` and `reason_code='invalid_asset_declaration'`. `byte_size` is the exact count of file content octets. A declaration with `asset_kind='source_asset'` MUST NOT carry raw evidence bytes, object-store keys, or raw storage paths; a violation fails with `failure_code='export_model_invalid'` and `reason_code='invalid_asset_declaration'`.

**REQ-RPT-046**
Source reference objects MUST use Table 9-H.

**Table 9-H. Source reference schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `source_ref_id` | `identifier` | Yes | No | None | Unique in export model. |
| `source_family` | string | Yes | No | None | Core source family token. `post` invalid. |
| `source_record_id` | `identifier` | Yes | No | None | Source record identity. |
| `source_path` | `export_model_path_v1` | Yes | Yes | None | Null when the reference is record-level. |
| `source_snapshot_id` | `identifier` | Yes | No | None | Must equal export model `snapshot_id`. |
| `source_summary` | `safe_string` | Yes | Yes | None | Safe summary or null. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-046d**
The snapshot and admitted render model MUST contain a closed `source_index` of
Table 9-H objects. Owner selectors are transient materialization inputs, never
public reference identities. Each projected field MUST bind exactly one indexed
source reference whose `source_path` is its exact keyed address in the immutable
snapshot's section tree. Record-level references have `source_path=null`.
Table 10-B defines both identities. Source paths identify the admitted snapshot
address; subsequent removal or reindexing of rendered containers does not change
that historical address or identity. Readers MUST NOT infer a source family,
record identity or provenance from a selector, display value or path prefix.

Each frozen source-record binding and the incident itself MUST have exactly one
record-level reference. Incident metadata binds the supplied incident identity;
its record-level partitions are the union of its fields, or `internal_only` when
no incident fields were collected. Other record-level partitions equal the
source record's partition union. Field-level partitions equal the bound field's
partitions. This revision emits `source_summary=null`; presentation labels remain
redactable fields. Source-backed support references MUST bind their exact target's
record-level reference and partition set. Relationship endpoints bind the exact
frozen source record, never a fabricated selector.

After redaction, the indexes MUST retain the exact identity closure required by
selected content, including record-level anchors of retained field references
and cited support targets. A citation may retain an identity-only anchor for
a removed target; it MUST NOT restore that target's fields, Core summary or
display label. Source paths and IDs remain the admitted snapshot addresses.
Reference labels MUST preserve the restrictions of selected content and its
dependencies. Source and support summaries remain null unless an explicitly
selected field supplies a release-safe summary. Subject display digests retain
their admitted identity meaning; they MUST NOT become raw display copies.

A retained relationship view requires its exact source and destination Core
summaries and its selected, unchanged structural endpoint and link-kind inputs.
A dropped or transformed structural input MUST NOT invent a different edge.
Endpoint labels MUST come from retained Core display names, with no raw-label
fallback. Removing a derived view for missing inputs does not alter the actual
Core rule decision on any surviving source field.

Admission, replay and recovery MUST reject absent indexes, unknown members,
retired selector identities, duplicate or orphaned identities, invalid paths,
cross-snapshot references, mismatched source owners, partition disagreements and
missing record-level bindings. Indexes MUST NOT be repaired from retained values
or live data. The clean cutover retires snapshot v2 and render-model v3 inputs;
there is no compatibility reader or inferred historical provenance.

**REQ-RPT-046b**
`cartulary.reporting_support_ref.v1` MUST use Table 9-H1. A support ref identifies exactly one support target and MUST NOT contain raw evidence bytes, raw evidence previews, raw storage paths, object-store keys, blob hashes, credentials, secret references, reveal-map values, or raw source values. A support ref that requires those values to be meaningful is invalid for Reporting v1 and MUST fail with `failure_code='export_model_invalid'` and `reason_code='invalid_support_ref'`.

**Table 9-H1. Support-reference schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_support_ref.v1`. |
| `support_ref_id` | `identifier` | Yes | No | None | Unique in export model. |
| `support_kind` | string | Yes | No | None | `source_record`, `evidence_item`, `artifact_record`, `timeline_event`, `relationship`, or `diagram_selection`. |
| `support_target_ref` | string | Yes | No | None | One target ref matching Table 9-H2. Multiple targets require multiple support refs. |
| `source_ref_id` | `identifier` | Yes | Yes | None | Non-null when the support target is source-backed; it MUST resolve to exactly one Table 9-H source ref. |
| `source_snapshot_id` | `identifier` | Yes | No | None | Must equal export model `snapshot_id`. |
| `support_role` | string | Yes | No | None | `primary`, `corroborating`, `context`, or `derived_from`. |
| `custody_state` | string | Yes | No | None | `not_applicable`, `referenced`, `available_in_evidence_store`, or `withheld_by_policy`; no state grants access to raw bytes. |
| `source_summary` | `safe_string` | Yes | Yes | None | Redaction-safe summary or null. It MUST NOT be an evidence excerpt unless the excerpt is represented as a redacted field elsewhere in the export model. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment and ordered by §12.2. |

**Table 9-H2. Support target ref grammar**

| `support_kind` | `support_target_ref` grammar |
| --- | --- |
| `source_record` | `record:{source_family}:{source_record_id}` where `source_family!='post'`. |
| `evidence_item` | `evidence:{evidence_id}` for a Core evidence metadata record, never a blob/object-store key. |
| `artifact_record` | `artifact:{artifact_id}`. |
| `timeline_event` | `timeline:{record_id}`. |
| `relationship` | `relationship:{relationship_id}`. |
| `diagram_selection` | `diagram:{diagram_id}`. |

**REQ-RPT-046c**
Source-owner support selectors are admission inputs only. Materialization MUST
resolve them to Table 9-H2 targets and emit Table 10-B support identities in every
`support_refs` array. Selectors resolving to the same target in one snapshot
coalesce to one identity. Every emitted support identity MUST resolve to exactly
one entry in that snapshot's `support_index`; every index entry MUST be referenced.
The index and reference arrays are strictly ordered by identity without duplicates.
Retained models containing source selectors or retired truncated hash identities
MUST be rejected, not converted.

Every target MUST exist in the frozen inventory: `source_record` resolves by
source family and record ID (or the model's own incident for `incident_metadata`),
`evidence_item` by an Evidence metadata record, `artifact_record` by an Artifact
record, `timeline_event` by `timeline_events`, `relationship` by `relationships`,
and `diagram_selection` by `diagrams`. Target kind and identity must agree; a
matching ID in another family does not suffice. Snapshot admission, render
admission, replay and imported-model validation MUST reject missing, mismatched,
duplicate or orphaned references with `export_model_invalid / invalid_support_ref`.
They MUST NOT query live state, synthesize absent records, infer provenance or
drop invalid support to repair retained history. Table 9-H1's source-reference
and disclosure obligations apply independently of this target-closure check.

## 9.5 Closed nested objects

**REQ-RPT-046a**
The nested objects `validation_summary` (a member of Table 9-A), `content_class_summary` and `section_validation` (members of Table 9-B), and `display_times` (a member of Table 14-B) are closed objects whose exact members are defined in Tables 9-I, 9-J, 9-K, and 14-D respectively. Each MUST always serialize its full member set with explicit values so that canonical bytes are fully determined; absent data uses the stated null or `0` value, never member omission. Unknown members are invalid. A later revision can add a named extension container to any of these objects under REQ-RPT-037.

**Table 9-I. `cartulary.reporting_export_model_validation.v1` (export-model `validation_summary`)**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_export_model_validation.v1`. |
| `result` | string | Yes | No | None | `passed` or `failed`. |
| `issue_count` | `finite_integer` | Yes | No | `0` | Count of retained export-model validation issues. |
| `issues[]` | array | Yes | No | `[]` | Each item is `{stage, severity, export_model_path, failure_code, reason_code}` restricted to §23.3 stage tokens and §23.4 safe details; ordered by §23.4. |

**Table 9-J. `content_class_summary` object**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `source_evidence` | `finite_integer` | Yes | No | `0` | Count of source-evidence blocks. |
| `derived_analytic` | `finite_integer` | Yes | No | `0` | Count of deterministic analytic blocks. |
| `curated_narrative` | `finite_integer` | Yes | No | `0` | Count of curated narrative blocks. |
| `working_material` | `finite_integer` | Yes | No | `0` | Count of working-material blocks. |
| `presentation_text` | `finite_integer` | Yes | No | `0` | Count for `presentation_text`. |
| `support_reference` | `finite_integer` | Yes | No | `0` | Count for `support_reference`. |
| `validation` | `finite_integer` | Yes | No | `0` | Count for `validation`. |
| `template_boilerplate` | `finite_integer` | Yes | No | `0` | Count for `template_boilerplate`. |

**Table 9-K. `section_validation` object**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `result` | string | Yes | No | None | `passed` or `failed`. |
| `issue_count` | `finite_integer` | Yes | No | `0` | Count of section-local validation issues. |

## 9.8 Exact Reference Pack binding

**REQ-RPT-046rp**
Snapshot admission MUST capture and pin one current Reference Pack set in the same transaction as its Job and immutable input. `reference_pack_set_binding.v1` is closed and contains required, non-null `pack_set_id`, `pack_set_sha256`, and `provenance`. The identifier and digest MUST identify the same canonical Reference Pack set. `provenance` MUST contain exactly the set's first-success provenance anchor for every member, ordered by `pack_key`; current administrative verification MUST NOT replace that history. The snapshot Job owns the admission pin, retained while its snapshot, releases, pending work or history require it. Replay MUST reuse the admitted binding.

Incident exports obtain Reporting's exact retained snapshot bindings through a narrow owner interface in the export's consistent transaction. Reporting verifies snapshot identity, derivation, admitted source-capture shape and its independent stored digest before returning a binding. Repeated uses of one set coalesce only when the whole immutable binding agrees. More than 1024 distinct sets fails under the Reference Data portable catalog bound. This discovery is independent of whether optional snapshot payloads are embedded and remains available for retained snapshots when current Reporting routes are unclaimed. It never resolves the current active set.

Snapshot export models, release inputs, composition previews, render outputs and rerenders MUST preserve this binding. Render admission and execution MUST validate the retained binding and bytes through Reference Data's interface. Missing, altered or unavailable required content MUST fail explicitly without resolving the current deployment set or substituting Base. A participant MUST NOT change the binding. Pack-sensitive derived fields MUST additionally retain the exact entry identity and the member's immutable provenance. The new export model and derivation contracts are v2; retained v1 inputs are rejected by the coordinated cutover preflight, never assigned a guessed set.

## 9.9 Portable historical artifacts

**REQ-RPT-046rv — Native snapshot readiness**
Before application readiness, Reporting MUST validate every retained native
source capture, its stored digest, identity and independent snapshot
row bindings. The retained capture metadata MUST equal the originating
retained snapshot capture and independent request admission; recomputing a modified capture's digest does not
authorize replacement content. Required native admission history remains retained
while its snapshot exists. Missing or inconsistent evidence blocks readiness;
the reader MUST NOT repair it, rerender it, or replace it with current source.

Snapshot Job input is a closed typed owner payload under the current Job kind.
Replay, worker completion and readiness MUST validate its entire shape, original
row bindings, source-boundary object/token through the Revisions owner, frozen
capture identity and digest, and independent Jobs admission. Unknown or omitted
members MUST NOT disappear through permissive decoding. Worker completion of
an already materialized snapshot MUST validate the original binding before
returning success. Public Jobs response expiry does not expire required native
snapshot history: an owner reader may validate a successful private tombstone
using its retained immutable identity and the owner payload, without recreating
erased response or idempotency fields. A non-expired Job additionally requires
its exact admitted normalized-request digest and idempotency identity.

The native inventory MUST be read through one consistent read-only transaction,
with bounded per-row traversal. It is deployment history, not an incident export:
portable-catalog cardinality limits MUST NOT truncate or reject otherwise valid
native history. Reference Data MUST validate each exact immutable set binding
within that same read transaction through its retained-binding interface.
Validation does not create a pin, alter availability, select today's active set,
or require a historical attestation to remain fresh today. Imported artifact
validation remains the separate REQ-RPT-046pa boundary. This check does not
substitute for release-bundle, preview-output, pending-job or restored object
inventory validation.

**REQ-RPT-046rr — Native release completion and readiness**
Before application readiness, Reporting MUST validate every native release
against its independently retained release Job admission and originating
snapshot admission. Immutable bindings include the derived release identity,
incident, actor, transaction identity, snapshot/capture boundary, derivation,
source-model digest, template, profile, output kind/options, graph selection,
composition tuple, recipient partitions, scope and fixed render-admission time.
The complete closed Job payload and its independent Jobs admission MUST be
validated before either first completion or replayed completion can succeed.
Changing a self-reported digest does not authorize replacing those bindings.

A completed output MUST match its release row, root manifest, exact member
inventory and bytes, redaction manifest/profile view, and any separate sensitive
adjunct. The manifest must retain the admitted model identity, source digest and
exact Reference Pack binding. Missing, additional or inconsistent members block
readiness and repeated completion. Replayed completion MUST NOT replace bytes,
change the success/failure outcome, or change a retained failure reason. A failed
release has no render bundle or successful-output fields. Lifecycle transitions
such as approval, publication and invalidation do not alter admission or output.

Readiness MUST additionally verify the matching terminal Jobs outcome. An
unexpired success identifies the exact release resource; an unexpired render
failure identifies the exact release and retained reason. Legitimate expired
successful or failed private tombstones remain usable with the owner admission
and matching outcome, without reconstructing erased public summaries or
idempotency state. Queued, running or canceled Jobs cannot establish a completed
release. The native reader uses a consistent read-only transaction, bounded
per-release traversal, and Reference Data's retained-binding interface. It MUST
NOT rerender, repair, select current source or apply portable-catalog cardinality
limits. This boundary does not establish renderer/sandbox conformance and does
not replace preview-output or pending-job recovery validation.

**REQ-RPT-046pa**
When an incident export selects `optional_sections=['snapshots']` (possibly alongside other admitted tokens), Reporting supplies a canonical `cartulary.reporting_portable_artifacts.v3` catalog at `ext/snapshots/catalog.json`. The catalog is a transport object using Reference Data's strict JSON admission and RFC 8785 canonical serialization; it does not change the Reporting canonical bytes or digest of a transported artifact. Its required, non-null members are `schema_id`, `incident_id`, `snapshots`, and `releases`. Unknown members are invalid at every catalog object boundary. Empty inventories serialize as arrays, never null or omission. An export with no artifacts emits the empty catalog. A wholly absent optional section on import means no artifact payloads to retain; it never selects the current pack set.

Each snapshot row contains exactly `snapshot_id`, `source_capture_path`, and `source_capture_sha256`. The path is `ext/snapshots/{snapshot_id}/source-capture.json`. Each release row contains exactly `release_id`, `snapshot_id`, `manifest_path`, and `manifest_sha256`; its snapshot must occur in the same catalog, and its path is `ext/snapshots/releases/{release_id}/manifest.json`. Identifiers use lowercase canonical non-nil UUID syntax. Digests use lowercase SHA-256 hexadecimal. `source_capture_sha256` and `manifest_sha256` are the schema-domain object digests from REQ-RPT-049, not checksums of the archive member bytes alone. The enclosing portability archive inventory and each rendered file inventory use SHA-256 of exact file bytes. Implementations must retain this distinction for token manifests and all other schema-bearing files as well. Snapshot and release arrays sort by their respective identities and prohibit duplicates. Limits are 1024 snapshots and 4096 releases; the enclosing archive member and extracted-byte limits apply additionally. Catalog bytes are bounded at 16 MiB before decoding. Equality is admitted by each guard; a complete fixture must also satisfy the enclosing archive's limits. The byte ceiling is a defensive parser guard: the other closed catalog bounds and path lengths make a valid 16 MiB catalog unreachable, so validation uses a constraint argument plus an isolated guard test rather than increasing any limit.

The transported source capture must bind the same incident and snapshot and use the adopted derivation and `cartulary.reporting_snapshot_capture.v1` schema. Its closed source inventory and all source identities must validate. A release manifest must bind the catalog release and snapshot, carry null preview identity, and preserve its generated admission time. Its complete immutable Reference Pack binding must agree with the incident's exact reference catalog. A release's complete binding must equal its snapshot's binding. Reporting must validate stored object shape and digest before recovering bytes from JSONB. Import must additionally receive those exact producer bytes. Unknown, omitted, malformed or digest-inconsistent admitted content fails the import before publication; the consumer must not repair the source bytes or replace their provenance with destination verification.

Every rendered member is carried under `ext/snapshots/releases/{release_id}/files/{original_path}` and checked against its original manifest's byte count, digest, role and media type. Original paths are at most 512 UTF-8 bytes, relative, and contain no empty, dot or parent segment, reverse solidus, colon, ASCII control or DEL. The original manifest is retained unchanged. Every declared member, including the primary output, must be present; undeclared members and duplicate paths are invalid. Members sort by original path. Sensitive reveal maps are private adjuncts outside the bundle inventory under REQ-RPT-069. Their role, internal namespace, media type and bytes are invalid in this section; import must not admit a sensitive member by marking it omitted. The retired v1/v2 catalogs, full-model snapshot members, and `omitted_files` member are rejected. Retained incompatible history requires the coordinated clean cutover; readers do not convert it.

Imported artifacts are immutable historical source evidence owned by Reporting. They publish with the incident, exact reference catalog and pins in the parent's transaction, with attribution to the import operation. They do not create native snapshot/render Jobs, local release approvals or active pack selections. Exact operation replay preserves byte identity; conflicting identity or bytes fail. Re-export includes this retained source evidence and any subsequent native artifacts, preserving each source binding. Native reads and rerenders retain their existing exact-set requirements; importing source evidence does not create a native rerender route for an unexecuted source release. Recovery includes all imported files and validates catalog, inventory, digests and source bindings before readiness, without demanding source freshness at restore time or treating source attestations as destination trust.

# 10. Canonical JSON, identifiers, hashes, and deterministic timestamps

**REQ-RPT-047**
`reporting_canonical_json_v1` MUST be used for every Reporting-owned canonical object that participates in a hash. It MUST reject non-JSON values, duplicate object members, numbers outside `finite_integer`, NaN, infinities, binary blobs, and host-language object values.

**REQ-RPT-048**
Canonical object members MUST be serialized in bytewise ascending UTF-8 order of member names. Arrays MUST preserve the owner-defined order from the relevant schema or algorithm. Strings MUST serialize using JSON string escaping with shortest valid escapes for quotation mark, reverse solidus, and required control escapes. No insignificant whitespace is emitted. In particular, `&`, `<`, `>`, U+2028 and U+2029 serialize as their literal UTF-8 bytes, without HTML or JavaScript escaping. Report Composition imports this serialization profile; its composition and preview-source digests retain their companion-owned self-member exclusion and do not acquire the schema-domain prefix in REQ-RPT-049.

**REQ-RPT-049**
The hash input for a canonical object MUST be:

```text
<schema_id> LF
<canonical_json_bytes>
```

The output digest MUST be lowercase 64-character SHA-256 hexadecimal.

**REQ-RPT-050**
Reporting-generated identifiers MUST be derived from canonical tuple bytes using lowercase SHA-256 and a stable prefix. Implementations MUST NOT use random numbers, database sequence order, map iteration order, object insertion order, or render-time wall-clock values in generated IDs.

**REQ-RPT-050b**
Reporting-generated hash IDs MUST use Table 10-B. No other Reporting-generated hash ID is valid in this revision. A later revision that adds a Reporting-generated hash ID MUST add one row to Table 10-B in the same revision.

**Table 10-B. Generated-ID registry**

| ID | Prefix | Hash-input tuple |
| --- | --- | --- |
| `export_model_id` | `expm_` | Canonical object `{schema_id:'cartulary.reporting_export_model_id.v1', release_id, snapshot_id, derivation_version, render_admitted_at}` hashed under REQ-RPT-049. |
| `deck_id` | `deck_` | Canonical object `{schema_id:'cartulary.reporting_deck_id.v1', release_id, template_id, template_version, export_model_id}` hashed under REQ-RPT-049. |
| `token_id` | `tok_` | Canonical object `{schema_id:'cartulary.reporting_token_id.v1', release_id, stable_subject_ref}` hashed under REQ-RPT-049. |
| `snapshot_model_id` | `snapm_` | Canonical object `{schema_id:'cartulary.reporting_snapshot_model_id.v1', snapshot_id, derivation_version, snapshot_at}`; timestamp uses six fractional digits. |
| `source_ref_id` | `src_` | Canonical object `{schema_id:'cartulary.reporting_source_ref_id.v1', snapshot_id, source_family, source_record_id, source_path}` hashed under REQ-RPT-049. `source_path` is the immutable keyed snapshot field address or explicit null for a record-level reference. |
| `support_ref_id` | `sup_` | Canonical object `{schema_id:'cartulary.reporting_support_ref_id.v1', snapshot_id, support_kind, support_target_ref}` hashed under REQ-RPT-049. The identity belongs to the snapshot and is preserved across release and preview bindings. |
| Preview `export_model_id` | `expm_` | The release export tuple with `schema_id='cartulary.reporting_preview_export_model_id.v1'` and `preview_attempt_id` replacing `release_id`. |
| Preview `deck_id` | `deck_` | The release deck tuple with `schema_id='cartulary.reporting_preview_deck_id.v1'` and `preview_attempt_id` replacing `release_id`. |
| Preview `token_id` | `tok_` | The release token tuple with `schema_id='cartulary.reporting_preview_token_id.v1'` and `preview_attempt_id` replacing `release_id`. |


**REQ-RPT-050a**
`content_manifest_digest_v1` MUST be used to reduce a set of files to one digest wherever this NLSpec requires a digest "of every" file in a declared set (for example `font_manifest_sha256` and `package_store_digest` in §20). The algorithm MUST:

1. enumerate every file in the declared set as a `{path, sha256}` pair, where `path` is the file's POSIX-relative UTF-8 path in Unicode NFC and `sha256` is the lowercase SHA-256 of the file's exact bytes; file modification time, ownership, and permission bits MUST NOT be inputs;
2. reject symbolic links, hard links, and filesystem-special entries in the set;
3. sort the pairs bytewise ascending by `path`, rejecting any duplicate normalized path;
4. serialize the sorted array under `reporting_canonical_json_v1`; and
5. emit the lowercase 64-character SHA-256 of that canonical byte string.

Because the digest depends only on relative paths and content bytes, two conforming implementations enumerating the same file set MUST produce the same digest.

**REQ-RPT-051**
A Reporting-owned generated timestamp participates in canonical hash input only when it equals `render_admitted_at`. Wall-clock timestamps are valid only in diagnostic-only artifacts excluded from `output_sha256`; such fields MUST be named `observed_at` or `diagnostic_observed_at`.

**REQ-RPT-051a**
When copied into a hash-participating field, `render_admitted_at` MUST serialize with exactly six fractional digits. A Core-supplied value without fractional digits MUST be normalized by appending `.000000Z`. A Core-supplied value with one through five fractional digits MUST be right-padded with zeroes to six digits. A Core-supplied value with more than six fractional digits is invalid and MUST fail with `reason_code='invalid_timestamp_value'`.

**REQ-RPT-052**
The fields in Table 10-A MUST equal `render_admitted_at` whenever they appear in hash-participating objects.

**Table 10-A. Deterministic generated timestamp fields**

| Member | Required rule |
| --- | --- |
| `export_model_created_at` | Equals `render_admitted_at`; participates in export-model hash. |
| `toolchain_snapshot.created_at` | Equals `render_admitted_at`; participates in toolchain snapshot hash. |
| `bundle_created_at` | Equals `render_admitted_at`; participates in bundle-manifest hash. |
| `validation_summary.created_at` | Equals `render_admitted_at` when retained in bundle. |
| Redaction manifest generated timestamps | Equal `render_admitted_at` when emitted. |
| Token manifest generated timestamps | Equal `render_admitted_at` when emitted. |

# 11. Export-model materialization algorithm

**REQ-RPT-053**
`materialize_reporting_export_model_v2` MUST execute the deterministic stages in Table 11-A in order. A stage MUST NOT read live source state after the source boundary is frozen for that stage.

**Table 11-A. Materialization stages**

| Order | Stage | Required behavior |
| ---: | --- | --- |
| 1 | Decode and validate release tuple | Validate tuple fields, defaults, output options, and Core dependency status. |
| 2 | Bind deterministic timestamps | Copy `render_admitted_at` into every hash-participating generated timestamp field. |
| 3 | Validate template manifest | Validate `cartulary.reporting_template_pack_manifest.v3`, supported output kind, supported release scope, bindings, assets, and render profiles. |
| 3a | Validate composition input | When the composition tuple is non-null, validate `cartulary.report_composition.v1` canonical bytes, tuple digest, template binding, schema closure, layout and click-profile references, authored-text role limits, diagram declarations, and pre-snapshot anchor syntax. |
| 4 | Read immutable snapshot | Read only the Core-owned immutable snapshot boundary. |
| 5 | Materialize source refs | Create source-reference objects for selected source material. |
| 6 | Materialize records and relationships | Emit record summaries and relationship summaries using §9 schemas. |
| 7 | Materialize timeline events | Emit timeline event objects using §14 and Core-provided timeline sort fields. |
| 8 | Materialize subjects | Emit tokenizable subjects using §13. |
| 9 | Bind section declarations | Bind eligible source candidates to §9 section declarations and their exact semantic expansion dimensions. Preserve admitted source addresses for Core rule selection; do not substitute display labels for dimension identities. |
| 10 | Assign disclosure partitions | Execute `assign_party_disclosure_partitions_v1` and companion §12 algorithms. |
| 11 | Filter disclosure partitions and apply redaction | Execute `filter_disclosure_partitions_v1`, Core redaction rule selection, `apply_token_backed_redaction_v1`, and `derive_display_token_v1` token substitution. |
| 12 | Assemble retained sections and diagrams | Resolve frozen Graph Projection inputs and diagram selection rules under §15. Emit the retained template instances and narrative blocks, assign `ordering_key` via `derive_section_ordering_key_v1`, and include template-owned diagram references in declaration order. Derive expansion labels only from retained values and selected tokens. |
| 13 | Validate resources | Apply §25 resource limits to the post-redaction model before render source generation. |
| 14 | Canonicalize export model | Serialize under §10 and compute export-model hash. |

The canonical export model MUST contain these final expanded sections and full
retained diagram objects before deck or renderer-source generation. The same
retained source field MAY appear in distinct selected sections; each occurrence
counts toward the section/block/field limits. Snapshot source-field identities
remain unique at source admission. Section source/support/partition unions MUST
include case-derived title labels as well as retained child blocks. Diagram
reference blocks also count toward the limits; an empty non-validation section
is removed, or fails when required, after resolving its diagram declarations.
The canonical byte limit applies after this assembly.

Deck derivation and diagram serialization MUST consume this assembled model.
They MUST NOT reselect fields from the snapshot or re-resolve diagram content
from graph inputs. Composition operations and chunking work on independent
copies and MUST NOT mutate the canonical model. Private semantic origin indexes
MAY bind section IDs to declaration/dimension identities; they MUST NOT hold a
parallel case-content tree. An altered model or different template must be
rejected before source generation. Replay performs the same deterministic
assembly from the frozen admission, without live source reads.

**REQ-RPT-053a**
Composition validation in Table 11-A stage 3a MUST be fail-closed and MUST complete before snapshot-dependent anchor resolution. The stage MUST validate only the normalized release tuple, the selected template manifest, the digest-bound composition bytes, adopted companion-schema closure, and declared template vocabularies. It MUST NOT read live workbook state, mutable projections, renderer output, system time, filesystem order, map iteration order, or network resources.

A valid non-null composition input MAY affect export-model admission only through the following Reporting-owned surfaces:

- replacement section or slide titles;
- inserted `paragraph` blocks with `content_class='presentation_text'`;
- inserted or replaced `speaker_note` blocks and slide `speaker_notes`;
- composition-owned diagram declarations admitted into `diagrams[]` with `diagram_id` equal to the composition declaration `decl_id`;
- deck-overlay operations whose effects are defined in Table 17-D.

No other composition-authored case content is valid in Reporting v1. Raw Markdown, raw Mermaid, raw HTML, arbitrary diagram vertices, arbitrary diagram edges, post-redaction bytes, and workbook mutations MUST fail with `failure_code='composition_invalid'` and the most specific reason code in §23.6 when attributable.

**REQ-RPT-054**
Materialization failures MUST use Table 11-B. The first failure MUST be selected by validation issue ordering in §23.4.

**Table 11-B. Materialization failures**

| Condition | `error.code` | `failure_code` | `reason_code` |
| --- | --- | --- | --- |
| Duplicate stable ID | `release_render_failed` | `export_model_invalid` | `duplicate_stable_id` |
| Duplicate export-model path | `release_render_failed` | `export_model_invalid` | `duplicate_export_model_path` |
| Dangling source reference | `release_render_failed` | `export_model_invalid` | `dangling_source_ref` |
| Dangling support reference | `release_render_failed` | `export_model_invalid` | `dangling_support_ref` |
| Cross-snapshot source reference | `release_render_failed` | `export_model_invalid` | `snapshot_ref_mismatch` |
| Invalid asset declaration | `release_render_failed` | `export_model_invalid` | `invalid_asset_declaration` |
| Missing content class | `release_render_failed` | `content_class_missing` | `content_class_missing` |
| Invalid generated identifier | `release_render_failed` | `export_model_invalid` | `invalid_generated_identifier` |
| Export model exceeds resource limit | `release_render_failed` | `export_model_resource_limit_exceeded` | Limit-specific reason from §25 |
| Missing blocked Core dependency | `release_render_failed` | `export_model_invalid` | `blocked_core_dependency` |
| Required pinned Reference Pack unavailable during preparation | `required_reference_pack_unavailable` | `export_model_invalid` | `required_reference_pack_unavailable` |
| Composition tuple digest, template binding, schema, or vocabulary invalid | `release_render_failed` | `composition_invalid` | Composition-specific reason from §23 |
| Required section becomes empty after filtering | `release_render_failed` | `export_model_invalid` | `required_section_empty` |
| External release has no emitted sections | `release_render_failed` | `export_model_invalid` | `empty_export_model_sections` |

# 12. Party assignments, disclosure partitions, and retained structure

## 12.1 Reporting-facing assignment projection

**REQ-RPT-055**
Reporting MUST consume party assignment source state through the Core-owned projection in Table 12-A.

**Table 12-A. `record_party_assignment.v1` Reporting-facing projection**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `assignment_id` | `identifier` | Yes | No | None | Stable assignment identity. |
| `incident_id` | `identifier` | Yes | No | None | Must match the release tuple incident. |
| `subject_record_id` | `identifier` | Yes | No | None | Host or Identity record ID. |
| `subject_record_type` | string | Yes | No | None | `host` or `identity`. |
| `party_id` | `identifier` | Yes | No | None | Party record ID present in the snapshot. |
| `assignment_kind` | string | Yes | No | None | `subject_party`, `responsible_party`, or `observer_party`. |
| `assignment_state` | string | Yes | No | None | `active`, `superseded`, or `deleted`; Reporting consumes only `active`. |
| `source_ref` | object | Yes | Yes | None | Core source reference object or null when Core source state has none. |
| `provenance` | string | Yes | No | None | `manual`, `import`, `system`, or `rollback`. |
| `created_by_user_id` | `identifier` | Yes | No | None | Core user identity. |
| `created_at` | `timestamp` | Yes | No | None | Core assignment creation timestamp. |
| `superseded_by_assignment_id` | `identifier` | Yes | Yes | None | Null unless `assignment_state='superseded'`. |
| `deleted_at` | `timestamp` | Yes | Yes | None | Null unless `assignment_state='deleted'`. |
| `row_version` | `identifier` | Yes | No | None | Core row-version token. |

**REQ-RPT-056**
A snapshot containing more than one active assignment for the same `(subject_record_id, party_id, assignment_kind)` MUST fail export-model materialization with `failure_code='export_model_invalid'` and `reason_code='duplicate_active_party_assignment'`.

## 12.2 Disclosure partition vocabulary and ordering

**REQ-RPT-057**
Disclosure partition references MUST use the vocabulary in Table 12-B. Arrays of disclosure partition references MUST be sorted by `partition_order`, then exact partition token.

**Table 12-B. Disclosure partition references**

| Partition ref | `partition_order` | Meaning |
| --- | ---: | --- |
| `public` | 0 | Eligible for every recipient only after §12.4 or no subject-bearing content is present. |
| `party:{party_partition_segment}` | 10 | Eligible only when the recipient partition includes the Party. |
| `internal_only` | 90 | Eligible only for internal draft or internal review. |
| `blocked` | 99 | Not eligible for release output; must be dropped, redacted, or fail. |

**REQ-RPT-057a**
The segment after `party:` MUST be a `party_partition_segment`: a JSON string segment containing 1 to 128 Unicode scalar values and no colon, slash, backslash, hash, Unicode whitespace, C0 controls, C1 controls, NUL, or surrogate code points. The full partition ref is the exact string `party:` followed by that segment. Reporting MUST NOT parse or accept nested delimiters, percent-decoding, case folding, trimming, or alternate forms. If Core exposes only a Party identifier that does not satisfy `party_partition_segment`, Reporting MUST fail preparation under REQ-RPT-027b with `reason_code='invalid_recipient_partition_ref'` rather than inventing an encoding.

**REQ-RPT-058**
Direct Party display labels default to non-public. A Party display value MAY be assigned `public` only when the Party source state marks the label as public-directory eligible and the redaction profile explicitly permits public Party labels for the release scope. Omission of either condition means the value remains party-partitioned or is redacted.

## 12.3 Partition assignment algorithm

**REQ-RPT-059**
`assign_party_disclosure_partitions_v1` MUST apply Table 12-C in order. Later rows do not override an earlier `blocked` partition unless the redaction stage removes the blocked content entirely.

**Table 12-C. Disclosure partition assignment**

| Content condition | Partition result |
| --- | --- |
| Template-owned boilerplate with no case facts | `public`. |
| Case content that references no Host, Identity, Party, tokenizable subject, source display value, or free-text case detail | `public`. |
| Subject has one or more active `subject_party` assignments | One `party:{party_partition_segment}` per active assignment, ordered by §12.2. |
| Subject has no active assignment and source content is not redacted away | `internal_only` for internal scopes; `blocked` for external release. |
| Derived summary or aggregate | Union of all contributor partitions; Reporting does not declassify it. |
| A block contains fields or children with different partition sets | Evaluate each declared canonical candidate under its selected Core rule; reconstruct the existing structure with `retain_declared_structure_v1`. No partition is removed merely because another recipient can receive another field. |
| Any content class marked security-sensitive by redaction profile | `blocked` unless redaction removes the sensitive value. |
| Content from a record with `deleted_state='deleted'`, or a `superseded` record not opted in for external release under REQ-RPT-043a | `blocked`; the content MUST be dropped before render and MUST NOT be emitted. |

Source-backed references preserve the target's restrictions. A field carrying a
`support_ref` MUST inherit the union of that support object's labels and the
labels of its exact frozen target. A relationship's fields MUST inherit both
endpoint records' restrictions. Record-level targets use the union of their
fields, including the Records envelope and domain projections of the same
source identity; a generic envelope is not a separate public identity. A
rough-capture Host or Identity mention without a canonical subject assignment
inherits `internal_only` internally and `blocked` externally. These rules add
restrictions; they never erase previously assigned labels or treat observation
or responsibility links as recipient permission.

Reference propagation MUST compute the least fixed point before selecting
redaction actions. Cycles are legal and MUST terminate without changing the
result under input reordering. Every target resolves through exact frozen
family, identity and snapshot bindings; a missing or foreign target fails
materialization. A selected drop of a target does not retroactively make an
unchanged citation to that target public. These reference rules do not establish
that unlinked free text is non-identifying and do not authorize declassification. Content without an adequate disclosure derivation remains unresolved.

Source-record assembly precedes subject-family resolution. Disjoint source
contributions for one record may supply different required metadata fields;
they need not repeat the same identity fields in each contribution. The complete
record must have unique field keys and satisfy the same frozen mention binding
at admission, replay and recovery. Conflicting repeated identity data is rejected,
not resolved by contribution order.

An entity mention MUST bind its explicit `source_record_id`,
`source_field_key`, `entity_type`, `resolution_status` and nullable
`resolved_record_id` from frozen source data. Missing members, unknown status,
a resolved mention without an exact selected canonical Host/Identity of the
same family, or an unresolved mention with a non-null resolved ID fail
materialization. The origin must exist in the frozen source inventory. The
origin record and mention inherit their combined restrictions; separate field
names alone do not establish finer semantic disclosure. A semantic mention-list key
need not name a scalar field in the canonical projection. Resolution never
erases prior labels. A resolved mention uses the canonical subject for token
resolution, including its canonical display digest and private reveal value;
it MUST NOT emit a second unresolved-mention subject or reveal the mention's
raw alias as that canonical subject's display. An unresolved mention retains
its own stable mention identity. The mandatory subject census is validated on
admission and replay; a reader MUST NOT repair a missing subject.

## 12.3A Disclosure partition filtering

**REQ-RPT-059a**
`filter_disclosure_partitions_v1` MUST run inside Table 11-A stage 11 after `assign_party_disclosure_partitions_v1` and before token substitution. It computes the effective allowed-partition set `A`:

- `release_scope='external_release'`: `A = {public} ∪ recipient_partition_refs[]`.
- `release_scope='internal_review'` or `release_scope='internal_draft'`: `A = {public, internal_only} ∪ {party:{party_partition_segment} for every Party in the snapshot}`.
- `blocked` is never a member of `A`.

For every section, block, field, record summary, relationship summary, timeline event, subject, asset, and source reference, the element is releasable as-is iff every member of its `disclosure_partition_refs[]` is in `A`. An element carrying any ref outside `A` MUST be resolved only through the
canonical content decisions below. Fields, blocks and sections are canonical
selection candidates: exact path rule, then declared `content_class` rule (if
any), then profile default. Section selection has no class fallback. Records,
relationships, timeline events, subjects and source/support indexes are derived
views of retained canonical content, not independent copies eligible for a
second action. They MUST be rebuilt only from retained inputs, omit unsupported
or empty references, and qualify their recomputed partition unions. They MUST
NOT restore excluded content, retain a raw display fallback, or accept exact
redaction selectors that pretend those projections are independent candidates.
Assets and graph values retain their own explicitly admitted typed candidates.
A new independently authored value cannot enter a derived view without its own
canonical candidate and disclosure derivation. The implementation MUST NOT choose among `drop`, `mask`, `truncate`, `stub`, `allow`, and failure independently of the selected Core rule. If the selected action's outcome still emits content carrying a ref outside `A`, materialization MUST fail with `error.code='release_render_failed'`, `failure_code='redaction_manifest_invalid'`, and `reason_code='disclosure_partition_unresolved'`.

**REQ-RPT-059b**
After `filter_disclosure_partitions_v1` and redaction complete, the canonical export model MUST contain only retained elements. `block_ordinal` and `field_ordinal` MUST be recomputed contiguously from `1` over retained elements in their original relative order. `content_class_summary`, `section_validation`, and export-model `validation_summary` MUST be computed over the post-redaction model. A retained non-`validation` section with zero blocks after filtering MUST be removed from `sections[]`; if the removed section was declared `required=true` by the template, materialization MUST fail with `failure_code='export_model_invalid'` and `reason_code='required_section_empty'`. Every selected Core `drop` MUST retain its actual rule entry and `dropped` outcome under §13.5. Consequential removal of an empty block or non-validation section MUST instead appear in `structural_removals`, with exact immediate causes. Structural pruning MUST NOT invent a Core drop rule or attribute a child's selected rule to its parent. This distinction does not exempt any element from disclosure evaluation or authorize silently ignoring a container rule.

## 12.4 No aggregate declassification

**REQ-RPT-060**
Reporting MUST preserve the union of all contributor disclosure partitions in
any retained derived summary. Counts, rounded dates, categorical labels, a
minimum contributor count, or template prose MUST NOT convert restricted
case-derived content to `public`. Every retained summary remains subject to the
same selected Core rule and effective-partition subset predicate as its inputs.
There is no partly-public or public-with-warnings result.

**REQ-RPT-060a**
The current profile does not derive public aggregate statistics or bucket source
timestamps for declassification. Source values remain source-preserving except
for their explicitly selected redaction actions. `time_bucket_granularity` is a
retired profile/view member, rejected even when null; silently accepting an
ineffective privacy setting is prohibited.

**REQ-RPT-060b**
The `aggregate_public_v1` algorithm and `aggregate_category_allowlist.v1` schema
are retired. Templates MUST reject `aggregate_category_allowlists`, including an
empty array. Blocks MUST reject `aggregate_only_non_identifying`,
`aggregate_policy_id`, `contributor_count` and `excluded_field_keys`, including
false, null or empty values. Neither an author-supplied flag nor a matching
artifact digest is declassification evidence.

**REQ-RPT-060c**
Any future declassification mechanism requires a separately adopted source,
derivation, proof-admission and disclosure contract. The current implementation
MUST NOT accept a compatibility alias, infer a historical proof, or use ordinary
template extension omission to enable such behavior. Existing typed graph
aggregation remains with Graph Projection and Network Flow; these computations
do not grant Reporting permission to remove their disclosure restrictions.

## 12.5 Retained structure

**REQ-RPT-061**
`retain_declared_structure_v1` MUST use Table 12-E after canonical rule selection
and field redaction. Its unit of selection is an admitted field or container,
never a substring, inferred sentence, array element or renderer-created fragment.
The historical `split_mixed_block_v1` algorithm and `split_from_block_id` member
are retired. Snapshot model v6, export model v7 and block v3 reject earlier
shapes and unknown members; no compatibility reader may reconstruct split
provenance or silently discard the retired member.

**Table 12-E. `retain_declared_structure_v1`**

| Step | Required behavior |
| ---: | --- |
| 1 | Select section, block and field rules against their immutable admitted addresses. A selected container drop suppresses its subtree before descendant values or tokens are produced. |
| 2 | Apply the selected field action to the complete field value. A scalar array remains one field. A free-form paragraph remains text; there is no NLP-based splitting or implicit declassification. Every retained value must still satisfy the effective partition subset predicate. |
| 3 | Reconstruct the existing block hierarchy from retained fields and children, preserving their relative order, block kind and content class. Do not create sibling blocks or a second family of fragment identities. A source field cannot survive in a parallel record, event or source index after its canonical removal. |
| 4 | Recompute each container's source, support and disclosure unions from retained descendants. Do not retain a removed child's labels or broaden a surviving child's eligibility. Remove empty containers under REQ-RPT-059b and record their actual causal removals separately from selected Core drops. |
| 5 | Renumber retained block and field ordinals contiguously and derive ordinary structural IDs under REQ-RPT-040b. Preserve immutable source-reference identities and original decision addresses; a new ordinal must not retarget a selected rule. |
| 6 | Serializers validate the retained block kind's complete structural contract. If removing a field leaves an invalid table or other structure, fail through that serializer's existing reason instead of inventing a replacement cell, row, label or value. |

This rule applies to declared list items, table rows, metric fields, Timeline
fields and child fields through their existing typed containers. Diagram labels
continue to use the typed graph/timeline redaction boundary in §15. A string
carrying several disclosure partitions remains indivisible: it is retained only
when the selected action's complete result qualifies, dropped only by a selected
drop, or rejected with `disclosure_partition_unresolved`. Structure retention
never chooses a more permissive redaction action.

# 13. Redaction, tokenization, token manifests, and reveal maps

## 13.1 Reporting-facing redaction interface

**REQ-RPT-062**
Reporting MUST consume a Core-owned redaction profile through the closed Reporting-facing view in Tables 13-A through 13-A3. Token substitution is Reporting-owned behavior triggered only by token-backed `mask` or `stub` rules in that view. Unknown members in any redaction-profile view object are invalid.

**Table 13-A. `cartulary.reporting_redaction_profile_view.v2`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_redaction_profile_view.v2`. |
| `redaction_profile_id` | `identifier` | Yes | No | None | Selected profile identity. |
| `redaction_profile_version` | `identifier` | Yes | No | None | Exact selected profile version. |
| `redaction_profile_sha256` | `sha256_hex` | Yes | No | None | Digest of selected profile bytes. |
| `profile_status` | string | Yes | No | None | `adopted_current`, `adopted_deprecated`, or `future_only`; only `adopted_current` may satisfy external-release conformance. |
| `allowed_release_scopes[]` | array | Yes | No | None | Non-empty subset of §7.3 values, ordered as Table 7-B. |
| `allowed_disclosure_partition_refs[]` | array | Yes | No | None | Declared allowed partitions for this profile, ordered by §12.2. |
| `include_superseded` | boolean | Yes | No | `false` | If `false`, superseded records are excluded from external release under REQ-RPT-043a. |
| `neutral_token_family` | boolean | Yes | No | `false` | Consumed by `derive_display_token_v1`. |
| `allow_authored_presentation_text` | boolean | Yes | No | `false` | If `false`, external release MUST reject composition-authored presentation text under REQ-RPT-087f. Internal scopes still apply partition filtering and redaction. |
| `path_rules[]` | array of `cartulary.reporting_redaction_rule_view.v1` | Yes | No | `[]` | Each item has `selector_kind='path'`; ordered by `rule_order`, then `rule_id`. |
| `content_class_rules[]` | array of `cartulary.reporting_redaction_rule_view.v1` | Yes | No | `[]` | Each item has `selector_kind='content_class'`; ordered by `rule_order`, then `rule_id`. |
| `profile_default` | `cartulary.reporting_redaction_rule_view.v1` | Yes | No | None | Exactly one rule with `selector_kind='profile_default'`. |

**Table 13-A2. `cartulary.reporting_redaction_rule_view.v1`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_redaction_rule_view.v1`. |
| `rule_id` | `identifier` | Yes | No | None | Unique within the profile. |
| `rule_order` | `finite_integer` | Yes | No | None | Positive integer. Duplicates are allowed only when `rule_id` breaks ties. |
| `selector_kind` | string | Yes | No | None | `path`, `content_class`, or `profile_default`. |
| `path_selector` | `export_model_path_v1` | Yes | Yes | None | Required when `selector_kind='path'`; otherwise null. Matches exact export-model path only. |
| `content_class_selector` | string | Yes | Yes | None | Required when `selector_kind='content_class'`; otherwise null. Must be a Table 9-C `content_class` token. |
| `action` | string | Yes | No | None | `allow`, `drop`, `mask`, `truncate`, or `stub`; token substitution uses token-backed `mask` or `stub`. |
| `token_backed` | boolean | Yes | No | `false` | Valid only for `action='mask'` or `action='stub'`. |
| `literal_output` | string | Yes | Yes | None | Required for non-token `mask` and non-token `stub` after Core default materialization; a safe single-line Unicode literal, including an explicitly empty string. Null for `allow`, `drop`, `truncate`, and token-backed rules. It is never a source-derived value or executable template. |
| `truncate_max_scalars` | `finite_integer` | Yes | Yes | None | Required only when `action='truncate'`; value MUST be in `[1, 240]`. Core REQ-01-387 applies this bound to the selected source-text prefix and appends its fixed truncation marker only when truncation occurs. The marker is additional to the prefix bound. The source-dependent result MUST NOT be copied into profile metadata. |
| `selected_rule_trace` | `selected_rule_trace.v1` | Yes | No | None | Safe trace object retained in redaction-manifest entries. |

**Table 13-A3. `selected_rule_trace.v1`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `selected_rule_trace.v1`. |
| `rule_id` | `identifier` | Yes | No | None | Selected rule identity. |
| `rule_precedence` | string | Yes | No | None | `path`, `content_class`, or `profile_default`. |
| `selector_kind` | string | Yes | No | None | Same value as the selected rule's `selector_kind`. |
| `selector_value` | string | Yes | Yes | None | Exact path selector or content-class selector when present; null for profile default. |

**REQ-RPT-062c**
Core redaction rule selection consumed by Reporting MUST be deterministic. For each redaction candidate, Reporting evaluates path rules in Table 13-A order and selects the first rule whose `path_selector` exactly equals the candidate `export_model_path`. If no path rule matches, it evaluates content-class rules in Table 13-A order and selects the first rule whose `content_class_selector` equals the candidate content class. If no content-class rule matches, it selects `profile_default`. Core REQ-01-386 continues to reject duplicate selectors at the same precedence; ordered rule arrays do not relax that prohibition. A profile with duplicate `rule_id`, duplicate same-precedence selectors, invalid selector nullability, a token-backed non-`mask`/`stub` action, missing required `literal_output`, a `truncate_max_scalars` outside `[1, 240]`, or a `literal_output` that cannot satisfy Table 9-D1 MUST fail before render output bytes with `failure_code='redaction_manifest_invalid'` and `reason_code='redaction_profile_invalid'`.

A field rule addresses the field object in the admitted model before removal,
for example `$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['value']`.
The field's `source_refs[]` remain separate provenance identities. Source
pointers such as `/hosts/...`, diagnostic numeric indexes, wildcards and query
expressions are not exact redaction selectors and MUST be rejected, not
translated. Keyed ancestry uses `children` for a nested block and escapes its
stable ID under Table 8-A. Snapshot-bound generated profiles MUST resolve and
freeze these model addresses before computing their profile digest; they MUST
NOT resolve rules from current live source state at render time.

Exact-selector shape validation and binding validation are separate. Every
snapshot-bound path rule MUST resolve to exactly one executable redaction
candidate in its admitted model before output. A stale, missing, ambiguous, or
unsupported target MUST fail with `redaction_manifest_invalid` /
`redaction_action_unresolved`; it MUST NOT silently fall through to class or
default behavior. Frozen-admission replay validates the same binding. Rejection
of an unimplemented target kind is fail-closed behavior, not conformance evidence
for required container or other element selection.

The full typed candidate census is bound before selection, including candidates
inside subsequently dropped subtrees. Selection is preorder over the eligible
canonical tree. Container `drop` records every eligible descendant path in
`container_decisions[].suppressed_paths` and stops traversal; no descendant
allocates tokens or receives a fabricated drop entry. An allow container is
qualified from retained descendants after field actions. Empty allowed
containers may then undergo separate structural pruning. Non-scalar actions
that cannot preserve the typed container schema fail; JSON stringification,
coercion to a paragraph and invented content classes are prohibited.

**REQ-RPT-062e**
`allow_authored_presentation_text` controls only composition-authored presentation text. It MUST NOT permit composition-authored case facts, raw source values, post-redaction edits, raw Markdown, raw Mermaid, raw HTML, arbitrary diagram nodes, arbitrary diagram edges, or workbook mutations. When the member is absent from a Core-owned profile view, Reporting MUST materialize the default `false`; if Core cannot expose the member, external release MUST fail with `failure_code='redaction_manifest_invalid'` and `reason_code='redaction_profile_invalid'`.

**REQ-RPT-062d**
Every redaction-manifest entry MUST include the selected rule's `selected_rule_trace.v1`. Trace objects MUST NOT contain raw source values, literal replacement values, token display values, credentials, raw evidence, object-store keys, or external URLs. Omission behavior: if Core omits a trace object, Reporting MUST synthesize the Table 13-A3 object from the selected rule metadata; if any required metadata is unavailable, materialization fails with `failure_code='redaction_manifest_invalid'` and `reason_code='redaction_profile_invalid'`.

**REQ-RPT-062a**
The redaction manifest byte form MUST be `cartulary.redaction_manifest.v5` from §13.5, serialized under `reporting_canonical_json_v1` and hashed under REQ-RPT-049. It MUST NOT contain source values, post-redaction literal values, replacement text, stub text, or mask parameters. The internal-scope and external-scope byte forms are identical for the same redaction outcomes.

**REQ-RPT-062b**
Core redaction actions consumed by Reporting MUST map to retained field states and manifest outcomes through Table 13-A1. Reporting MUST NOT reinterpret a selected Core action, invent a replacement value, or choose a different action to make a section renderable. If the selected action cannot produce a retained value that satisfies Table 9-D1 and the action is not `drop`, materialization MUST fail with `error.code='release_render_failed'`, `failure_code='redaction_manifest_invalid'`, and `reason_code='redaction_action_unresolved'`.

**Table 13-A1. Redaction action and outcome mapping**

| Selected Core action | Reporting precondition | Retained export-model result | Redaction-manifest outcome |
| --- | --- | --- | --- |
| `allow` | All disclosure partitions are releasable under REQ-RPT-059a. | `redacted_value_state='allowed'` when the Core rule explicitly allowed release; `unchanged` when no redaction was required. `value` is the redaction-safe source or derived value. | `unchanged`. |
| `allow` | Any disclosure partition remains outside the effective allowed set. | No retained value. | Invalid; fail with `disclosure_partition_unresolved`. |
| `drop` | Always valid when selected by Core rule. | Element absent after filtering and ordinal recomputation. | `dropped`. |
| `mask` | Rule emits a non-token literal output. | `redacted_value_state='masked'`; `value` is the literal output. | `masked`. |
| `mask` | Rule is token-backed and the field resolves to exactly one stable subject. | `redacted_value_state='tokenized'`; `value` is the `display_token`. | `tokenized`. |
| `mask` | Rule is token-backed and the field resolves to zero or multiple subjects. | No retained value. | Invalid; fail with `token_subject_not_unique`. |
| `truncate` | Rule emits a deterministic non-token literal output. | `redacted_value_state='truncated'`; `value` is the literal output. | `truncated`. |
| `stub` | Rule emits a non-token literal output. | `redacted_value_state='stubbed'`; `value` is the literal output. | `stubbed`. |
| `stub` | Rule is token-backed and the field resolves to exactly one stable subject. | `redacted_value_state='tokenized'`; `value` is the `display_token`. | `tokenized`. |

The coordinated cutover admits built-in profile revision `3`. Its internal
passthrough policy remains `allow` with no rule array entries. The external and
tokenized-review profiles materialize an exact-path `allow` rule for every
selected source field whose frozen `source_value_state` is `null`, except
external `working_material`, which retains its class-level `drop`. Such a rule
preserves JSON null and remains subject to ordinary disclosure qualification;
it does not authorize an otherwise restricted field. A null field receives no
token-backed rule. Rules are derived in ascending frozen source-binding order,
carry deterministic IDs and are part of the admitted profile bytes and digest.
Selection and rendering MUST NOT insert an alternate action after admission.
Retired built-in profile revisions are unsupported. Custom selected actions
that violate Table 9-D1 fail with `redaction_action_unresolved`, with no usable
partial model or render artifacts.

**REQ-RPT-063**
`apply_token_backed_redaction_v1` MUST apply Core redaction rule selection first, then apply Reporting token substitution only when the selected Core rule is token-backed and the field resolves to exactly one stable tokenizable subject.

**REQ-RPT-063a**
When `apply_token_backed_redaction_v1` substitutes a token, the emitted `display_token` MUST be produced by the Reporting-owned, fully deterministic `derive_display_token_v1`:

- **Grammar.** A `display_token` matches `^(HOST|IDENTITY|PARTY|SUBJECT)-[0-9]{4,}$`. Its character set is limited to `A`–`Z`, `0`–`9`, and `-`, which is a no-op under both `mermaid_source_serialize_v3` escaping (Table 16-C) and `slidev_markdown_escape_v2` (Table 18-B); a token therefore appears byte-identically in `.mmd` and `slides.md`.
- **Family tag.** The tag is `HOST`, `IDENTITY`, or `PARTY` from the subject's `subject_family`, unless the redaction profile sets `neutral_token_family=true` (default `false`), in which case every tag is `SUBJECT`.
- **Ordinal assignment.** Within one release, ordinals are assigned per emitted family over the set of subjects that receive at least one token substitution in the release. Within each emitted family, subjects sort by ascending `stable_subject_ref`; ordinals start at `1`, increment by `1`, and assign one token per distinct emitted `stable_subject_ref`. The ordinal is left-zero-padded to four digits; the §25 `subjects.count` hard limit guarantees four digits suffice.
- **Stability and non-reversibility.** The mapping from `stable_subject_ref` to `display_token` is one-to-one within a release and is recorded in the token manifest (Table 13-D) and reveal map (Table 13-E). The visible token MUST NOT encode any part of the subject's display value or its `canonical_display_value_sha256`.

Two conforming implementations given the same subjects and redaction profile MUST emit identical `display_token` values.

## 13.2 Tokenizable subject schema

**REQ-RPT-064**
`cartulary.tokenizable_subject.v1` MUST use Table 13-B.

**Table 13-B. Tokenizable subject schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.tokenizable_subject.v1`. |
| `stable_subject_ref` | string | Yes | No | None | Must match `stable_subject_ref_v1`. |
| `subject_family` | string | Yes | No | None | `host`, `identity`, or `party`. |
| `source_record_id` | `identifier` | Yes | Yes | None | Required for canonical Host, Identity, or Party; null only for unresolved mention subjects. |
| `source_record_type` | string | Yes | Yes | None | `host`, `identity`, `party`, or null for unresolved mention. |
| `entity_mention_id` | `identifier` | Yes | Yes | None | Required when `source_record_id=null`; otherwise null. |
| `canonical_display_value_sha256` | `sha256_hex` | Yes | No | None | SHA-256 over canonical object `{schema_id, stable_subject_ref, value}`; the raw value itself MUST NOT be emitted in external bundles. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty and sorted by §12.2. |
| `source_refs[]` | array | Yes | No | `[]` | Non-empty for incident-derived subjects; ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | May be empty only for unresolved rough-capture subjects; ordered by exact identifier. |

The display digest preimage uses `schema_id='cartulary.tokenizable_subject.v1'`,
the exact `stable_subject_ref`, and the frozen source record's selected display
value. Host, Identity and Party use their frozen `display_name` member;
an unresolved entity mention uses its frozen `raw_text` member. Neither case
falls back to normalized text, an alias, title, identifier or another field.
An absent source display value is explicit JSON null; it MUST NOT be
replaced by an identifier, selector, current label or fabricated placeholder.
The digest is SHA-256 of those canonical JSON bytes, with no second schema-domain
prefix. The subject object never contains `display_name` or the raw display
value. Snapshot admission and replay validate the digest against the independent
frozen record summary. A later token/reveal stage that requires an available
display value must reject its absence at that stage rather than infer one.

The three nullable identity members form a closed union. A canonical subject
has `source_record_id` and matching `source_record_type`, with null
`entity_mention_id`; an unresolved Host or Identity mention has both source
record members null and a non-null `entity_mention_id`. Unknown members,
omitted nullable members, mixed arms, altered display digests and unresolved
source bindings fail before rendering. Source/support reference arrays preserve
the exact frozen record binding and do not infer evidentiary support from Party
assignment links.

**REQ-RPT-065**
`stable_subject_ref_v1` MUST use Table 13-C. No other subject-reference grammar is valid in this revision.

**Table 13-C. `stable_subject_ref_v1` grammar**

| Subject case | Stable subject ref |
| --- | --- |
| Canonical host | `subject:host:record:{record_id}` |
| Canonical identity | `subject:identity:record:{record_id}` |
| Canonical party | `subject:party:record:{record_id}` |
| Unresolved host-like mention | `subject:host:mention:{entity_mention_id}` |
| Unresolved identity-like mention | `subject:identity:mention:{entity_mention_id}` |

**REQ-RPT-066**
An unresolved subject without `entity_mention_id` MUST fail with `failure_code='token_manifest_invalid'` and `reason_code='token_subject_not_stable'`. A token path that resolves to zero or multiple subjects MUST fail with `failure_code='token_manifest_invalid'` and `reason_code='token_subject_not_unique'`.

## 13.3 Token manifest

**REQ-RPT-067**
`cartulary.reporting_token_manifest.v3` MUST contain exactly the top-level fields in Table 13-D. External token manifests MUST NOT include raw subject display values or reversible display material.

**Table 13-D. Token manifest schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_token_manifest.v3`. |
| `release_id` | `identifier` | Yes | Yes | None | Exact admitted release identity, or null for a preview. |
| `preview_attempt_id` | `identifier` | Yes | Yes | None | Exact admitted preview identity, or null for a release. Exactly one identity arm is non-null. |
| `render_admitted_at` | `timestamp` | Yes | No | None | Exact normalized admission time; never recomputed during rendering or recovery. |
| `snapshot_id` | `identifier` | Yes | No | None | From release tuple. |
| `redaction_profile_sha256` | `sha256_hex` | Yes | No | None | Exact selected profile digest. |
| `created_at` | `timestamp` | Yes | No | None | Equals normalized `render_admitted_at`. |
| `redaction_manifest_sha256` | `sha256_hex` | Yes | No | None | Exact schema-domain digest of the complete selected-rule inventory in Table 13-F. |
| `entries[]` | array | Yes | No | `[]` | Ordered by `token_id`; empty when no tokens are used. |
| `entries[].token_id` | `generated_id` | Yes | No | None | Stable generated ID from Table 10-B. |
| `entries[].display_token` | string | Yes | No | None | Deterministic token string produced by `derive_display_token_v1` (REQ-RPT-063a). |
| `entries[].stable_subject_ref` | string | Yes | No | None | Must match one subject in the export model. |
| `entries[].subject_family` | string | Yes | No | None | `host`, `identity`, or `party`. |
| `entries[].source_record_id` | `identifier` | Yes | Yes | None | Null only for unresolved mention subjects. |
| `entries[].entity_mention_id` | `identifier` | Yes | Yes | None | Non-null only for unresolved mention subjects. |
| `entries[].canonical_display_value_sha256` | `sha256_hex` | Yes | No | None | Digest only; raw display value forbidden. |
| `entries[].recipient_partition_refs[]` | array | Yes | No | None | Required and ordered by §12.2. |


Each subject receives exactly one token entry, regardless of how many fields
selected token-backed actions. Multiple selected field rules may legitimately
use the same subject token. The token manifest MUST NOT contain a second path
inventory, `action`, `rule_id`, `actions`, `source_families`, `content_classes`,
or any raw source value. Recipient partitions are the exact admitted recipient
set, an explicit empty array for internal scopes; they are not the union of
source disclosure labels. Display-token ordinals must satisfy REQ-RPT-063a over
the complete emitted subject set.

## 13.4 Reveal map

**REQ-RPT-068**
`cartulary.reporting_token_reveal_map.v3` MUST use Table 13-E and MUST be retained only as a Core-authorized internal sensitive release artifact.

**Table 13-E. Reveal-map schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_token_reveal_map.v3`. |
| `release_id` | `identifier` | Yes | Yes | None | Exact admitted release identity, or null for a preview. |
| `preview_attempt_id` | `identifier` | Yes | Yes | None | Exact admitted preview identity, or null for a release. Exactly one identity arm is non-null. |
| `render_admitted_at` | `timestamp` | Yes | No | None | Exact normalized admission time; never recomputed during rendering or recovery. |
| `snapshot_id` | `identifier` | Yes | No | None | From release tuple. |
| `token_manifest_sha256` | `sha256_hex` | Yes | No | None | Digest of Table 13-D canonical bytes. |
| `redaction_profile_sha256` | `sha256_hex` | Yes | No | None | Exact selected profile digest. |
| `created_at` | `timestamp` | Yes | No | None | Equals normalized `render_admitted_at`. |
| `entries[]` | array | Yes | No | `[]` | Ordered by `token_id`; complete for every token manifest entry. |
| `entries[].token_id` | `generated_id` | Yes | No | None | Must match one token manifest entry. |
| `entries[].display_token` | string | Yes | No | None | Must equal token manifest value. |
| `entries[].stable_subject_ref` | string | Yes | No | None | Must equal token manifest value. |
| `entries[].subject_family` | string | Yes | No | None | `host`, `identity`, or `party`. |
| `entries[].source_record_id` | `identifier` | Yes | Yes | None | Null only for unresolved mention subjects. |
| `entries[].entity_mention_id` | `identifier` | Yes | Yes | None | Non-null only for unresolved mention subjects. |
| `entries[].canonical_display_value` | `safe_string` | Yes | No | None | Required internal value; forbidden in external bundle and token manifest. |
| `entries[].canonical_display_value_sha256` | `sha256_hex` | Yes | No | None | Must equal token manifest digest. |
| `entries[].recipient_partition_refs[]` | array | Yes | No | None | Required and ordered by §12.2. |

There MUST be exactly one reveal entry per token-manifest entry in the same
order. Its nullable source identity, subject family, recipient set and display
digest MUST exactly match that token entry. `canonical_display_value` is the
frozen selected record display value whose digest was admitted under Table
13-B. Missing display values fail with `token_manifest_invalid` /
`token_subject_not_stable`; Reporting MUST NOT substitute an arbitrary original
field value, identifier or placeholder. Empty but explicitly available display
values remain subject to the ordinary safe-string contract. Per-field original
values, rule IDs, actions and paths MUST NOT be retained in the reveal map.

These three artifact formats are version 3. Version 2 artifacts and retained
jobs using their semantics are unsupported by this clean pre-production cutover.
The dependency order is redaction manifest, then token manifest, then private
reveal map. The bundle binds both public manifests independently. No digest cycle,
legacy reader, dual writer or historical conversion is permitted. A fresh
initialization is required; preflight rejects retained incompatible history
before schema mutation. Nullable members and empty arrays remain explicit;
unknown or omitted members and duplicate JSON keys are invalid.

**REQ-RPT-069**
Reveal maps MUST remain private Reporting adjuncts, separately persisted from all shareable release and preview bundle members. They MUST NOT appear in any ordinary bundle manifest, token manifest, download, or incident-portability inventory. The renderer's artifact input MUST NOT include reveal-map bytes. Release approval hashes continue to bind shareable bundle content; a private reveal-map digest MUST NOT be added as an external child-file descriptor.

The sensitive adjunct and its parent output MUST commit atomically. A private map MUST bind the exact release or preview identity, snapshot, admission time, token-manifest digest, canonical subject-display digests and the complete token inventory. A preview map MUST NOT acquire a synthetic release identity. Missing, mismatched, truncated, duplicate or altered adjuncts MUST prevent successful output persistence. Core 04 authorization and retention apply independently of ordinary bundle access; this separation introduces no public reveal-map read route.

Every retained token use MUST agree with its exact path's selected-rule entry in
the bound redaction manifest, including the rule identity, action and disclosure
partitions. Each tokenized Table 13-F entry references exactly one Table 13-D
entry through `token_id`; non-tokenized entries have explicit null `token_id`.
Every token-manifest entry MUST be referenced by at least one tokenized entry.
Table 13-D binds the digest of the complete Table 13-F inventory, preventing a
truncated or changed use inventory from being accepted against an unchanged
token manifest. Per-path action and rule identity are owned only by Table 13-F;
they MUST NOT be collapsed into one subject-level action, rule, or union.
Completion and recovery MUST validate these relationships from the retained
canonical artifacts, including their identity, profile, model and digest bindings;
recomputing a modified private artifact's own digest does not satisfy them.
Native completion and recovery MUST additionally verify redaction, token and
private-map bytes against the exact retained admission model and profile. This
verification uses deterministic redaction only, without rerendering, live-source
queries, current-profile substitution or artifact repair. Matching display-value digests
inside an altered map do not establish agreement with the admitted source.

Full deployment recovery MUST preserve the canonical sensitive bytes and their schema-domain digests in Reporting's authoritative state. Before readiness, Reporting MUST validate each retained map against its own parent and token manifest in a consistent read-only transaction, using bounded per-record traversal without portable-catalog limits. Missing or altered required maps block readiness and MUST NOT be reconstructed. This adjunct check does not substitute for complete release/preview, renderer, or authorization qualification. Bundle v2 already excludes the sensitive role; enforcing this boundary corrects implementation without changing that schema identity.

## 13.5 Redaction manifest

**REQ-RPT-069a**
`cartulary.redaction_manifest.v5` MUST use Table 13-F. The object is canonicalized under §10 and its digest is the `redaction_manifest_sha256` bound by the release record and bundle manifest.

**Table 13-F. Redaction manifest schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.redaction_manifest.v5`. |
| `release_id` | `identifier` | Yes | Yes | None | Exact admitted release identity, or null for a preview. |
| `preview_attempt_id` | `identifier` | Yes | Yes | None | Exact admitted preview identity, or null for a release. Exactly one identity arm is non-null. |
| `render_admitted_at` | `timestamp` | Yes | No | None | Exact normalized admission time; never recomputed during rendering or recovery. |
| `snapshot_id` | `identifier` | Yes | No | None | From release tuple. |
| `redaction_profile_id` | `identifier` | Yes | No | None | Exact selected profile identity. |
| `redaction_profile_version` | `identifier` | Yes | No | None | Exact selected profile version. |
| `redaction_profile_sha256` | `sha256_hex` | Yes | No | None | Exact selected profile digest. |
| `created_at` | `timestamp` | Yes | No | None | Equals normalized `render_admitted_at`. |
| `entries[]` | array | Yes | No | None | One entry per canonical field reached by redaction, including `allow` outcomes and selected removals; ordered by `export_model_path` bytewise ascending, then `rule_id`. |
| `entries[].export_model_path` | `export_model_path_v1` | Yes | No | None | Pre-removal canonical path. |
| `entries[].content_class` | string | Yes | No | None | Table 9-C vocabulary. |
| `entries[].action` | string | Yes | No | None | `allow`, `drop`, `mask`, `truncate`, or `stub`; `hash` is invalid. |
| `entries[].rule_id` | `identifier` | Yes | No | None | Selected Core rule. |
| `entries[].rule_precedence` | string | Yes | No | None | `path`, `content_class`, or `profile_default`. |
| `entries[].selected_rule_trace` | `selected_rule_trace.v1` | Yes | No | None | Safe trace object from Table 13-A3. |
| `entries[].disclosure_partition_refs[]` | array | Yes | No | `[]` | The element's pre-filter refs, ordered by §12.2. |
| `entries[].partition_handling` | string | Yes | No | None | `releasable`, `redacted`, `dropped`, or `blocked_removed`. |
| `entries[].outcome` | string | Yes | No | None | `unchanged`, `masked`, `truncated`, `stubbed`, `tokenized`, or `dropped`. |
| `entries[].token_id` | `identifier` | Yes | Yes | None | Non-null only when `outcome='tokenized'`. |
| `container_decisions[]` | array | Yes | No | None | One decision per reached canonical section/block, ordered by pre-removal path. |
| `container_decisions[].export_model_path` | `export_model_path_v1` | Yes | No | None | Unique pre-removal section/block path. |
| `container_decisions[].element_kind` | string | Yes | No | None | `section` or `block`, agreeing with path ancestry. |
| `container_decisions[].content_class` | string | Yes | Yes | None | Null for sections; exact Table 9-C class for blocks. |
| `container_decisions[].action` | string | Yes | No | None | Selected `allow` or `drop`; other actions cannot preserve these typed schemas and fail. |
| `container_decisions[].selected_rule_trace` | `selected_rule_trace.v1` | Yes | No | None | Exact selected Core rule; no invented inherited rule. |
| `container_decisions[].disclosure_partition_refs[]` | array | Yes | No | None | For drop: original container labels. For allow: exact surviving descendant-field union, ordered by §12.2. |
| `container_decisions[].outcome` | string | Yes | No | None | `unchanged` for allow, `dropped` for drop. |
| `container_decisions[].suppressed_paths[]` | array | Yes | No | None | Strictly ordered complete eligible canonical subtree below a dropped container; empty for allow. No path may also have a field/container decision or another suppression owner. |
| `structural_removals[]` | array | Yes | No | `[]` | Consequential empty-container removals, ordered by pre-removal path; distinct from actual Core rule decisions. |
| `structural_removals[].export_model_path` | `export_model_path_v1` | Yes | No | None | Unique pre-removal section or block path; it cannot also have a selected Core `drop`. A retaining rule entry for the same container may coexist with its later structural removal. |
| `structural_removals[].removal_reason` | string | Yes | No | None | `empty_block` or `empty_section`, agreeing with the addressed container kind. |
| `structural_removals[].causal_paths[]` | array | Yes | No | `[]` | Strictly ordered immediate child paths whose removal empties this container. Field causes reference an actual `dropped` decision; block causes reference an actual `dropped` decision or another structural removal. Empty only for a container that was already empty before filtering. |

Structural removal is derived bottom-up from the complete admitted container
inventory. A block is empty only when no field or child block survives; a
non-validation section is empty only when no block survives. A retained
validation section may be empty. Every causal edge is to an immediate child,
so the evidence graph is acyclic. Missing known causes, duplicated, unrelated,
retained or non-child causes are invalid. Every published immediate child of a
removed container must appear in its causal list and be removed. An actual
retaining rule and later empty-container removal can describe the same path;
they are distinct stage facts, not contradictory final outcomes. A selected
Core drop must not be duplicated as consequential empty-container removal. Native completion/recovery recomputes the complete
removal inventory from frozen admission and compares canonical bytes; a portable
reader validates the closed shape and causal relationships without pretending
that those alone prove the unseen original inventory. The manifest's digest
binds field/container decisions, suppressed subtrees and structural removals.
Every reached canonical child has a reached allowed parent. Every suppressed
child has complete ancestry to its exact dropped owner. Native replay compares
the suppression census with frozen admission; portable validation cannot infer
unseen original content from a structurally valid list. Retired v3/v4 manifests are rejected
without repair or conversion; unchanged token/reveal shapes keep their v3 IDs.


# 14. Timeline section derivation

## 14.1 Core-owned timeline sort materialization

**REQ-RPT-070**
Reporting MUST consume Core-provided timeline sort fields in Table 14-A for every selected timeline event. Reporting MUST NOT independently parse, reinterpret, or normalize source timestamp text for ordering. Core 02 REQ-02-143a.1 owns the source interpretation. The typed provider contribution and Timeline fact contracts are version 2; retired version 1 contributions are rejected, with no fallback or inferred sort facts. Each Timeline source record has exactly one ordering fact bound by its explicit identity. Generic record-envelope copies do not create additional Timeline events.

**Table 14-A. Timeline sort materialization fields**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `activity_sort_ts` | `timestamp` | Yes | Yes | None | UTC timestamp or null when unresolved. |
| `activity_parse_state` | string | Yes | No | None | `parsed`, `missing`, `incomplete`, `ambiguous`, or `unparseable`. |
| `activity_precision_rank` | `finite_integer` | Yes | No | None | `6=second`, `5=minute`, `4=hour`, `3=day`, `2=month`, `1=year`, `0=unresolved`. |
| `date_entered_sort_key` | string | Yes | Yes | None | `YYYY-MM-DD` or null. |
| `record_created_at` | `timestamp` | Yes | No | None | Immutable source record creation timestamp from Core snapshot materialization. |
| `time_bucket` | string | Yes | No | None | `parsed_time` or `unresolved_time`. |

**REQ-RPT-071**
If the immutable snapshot lacks any required timeline sort materialization field for a selected timeline event, external release MUST fail with `failure_code='export_model_invalid'` and `reason_code='timeline_sort_key_missing'`.

## 14.2 Timeline event object and sort key

**REQ-RPT-072**
`cartulary.reporting_timeline_event.v1` MUST include the fields in Table 14-B.

**Table 14-B. Timeline event schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_timeline_event.v1`. |
| `record_id` | `identifier` | Yes | No | None | Source timeline record ID. |
| `activity_sort_key` | string | Yes | No | None | Serialized sort key from REQ-RPT-073. |
| `activity_sort_ts` | `timestamp` | Yes | Yes | None | From Table 14-A. |
| `activity_parse_state` | string | Yes | No | None | From Table 14-A. |
| `activity_precision_rank` | `finite_integer` | Yes | No | None | From Table 14-A. |
| `date_entered_sort_key` | string | Yes | Yes | None | Date key or null. |
| `record_created_at` | `timestamp` | Yes | No | None | From Table 14-A. |
| `display_times` | object | Yes | No | None | Closed redaction-safe object per Table 14-D (REQ-RPT-072a). |
| `fields[]` | array | Yes | No | `[]` | Field objects ordered by `field_ordinal`. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered source references. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered support references. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-072a**
`display_times` is a closed object with the exact members in Table 14-D. All members are always present; a member with no emitted value is explicit `null`. Every emitted string is redaction-safe after the redaction phase. Original activity text remains in the source fields and is subject to ordinary field redaction. It is not a Core ordering fact or a pre-redacted copy in the sort contribution. Template display selection must explicitly select the source field to display; Reporting MUST NOT reparse its text to derive ordering.

**Table 14-D. `display_times` object**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `primary_display` | `safe_string` | Yes | Yes | None | Emitted primary time string, or null when no time is displayed. |
| `precision_label` | string | Yes | Yes | None | One of `second`, `minute`, `hour`, `day`, `month`, `year`, or `unresolved`, matching `activity_precision_rank`; null when no time is displayed. |
| `source_text_display` | `safe_string` | Yes | Yes | None | Redaction-eligible original source time text, or null when withheld or absent. |

**REQ-RPT-073**
`select_timeline_rows_v1` MUST sort selected timeline rows by the tuple below. For descending variants, only the `activity_sort_ts` comparison in tuple element 2 is inverted. `time_bucket_order`, the null sentinel, precision ranking, date-entered sort, created-at ordering, and `record_id` tie-break order remain unchanged.

```text
time_bucket_order
activity_sort_ts or "9999-12-31T23:59:59.999999Z"
negative activity_precision_rank
date_entered_sort_key or "9999-12-31"
record_created_at
record_id
```

The serialized `activity_sort_key` is the canonical JSON array of the six tuple members above under the default ascending, unresolved-last order. All source ordering timestamps have exactly six fractional digits. Consumers compare typed tuple members, not the lexical JSON encoding. Variant ordering does not rewrite the stored key. Snapshot materialization retains the default order, including all selected source rows before template-specific filtering and limits; snapshot admission rejects malformed or inconsistent keys and out-of-order events. Snapshot `display_times` contains all three members with null values because no template or redaction profile is selected. Render derivation must populate display values only from redacted source, under Table 14-D; null snapshot values do not establish completion of display derivation.

`time_bucket_order` is `0` for `parsed_time` and `1` for `unresolved_time`. For a template-declared `unresolved_first` option, invert only `time_bucket_order`.

## 14.3 Timeline slide chunking and bounds

**REQ-RPT-074**
Timeline slide chunking MUST use Table 14-C.

**Table 14-C. Timeline bounds and chunking**

| Limit | Default | Hard limit | Behavior |
| --- | ---: | ---: | --- |
| Selected timeline rows | 240 | 240 | Beyond hard limit, apply the overflow selection rule per REQ-RPT-074a. |
| Rows per timeline slide | 8 | 12 | Template MAY set `1..12`; omitted means `8`. |
| Timeline overflow block | Required when rows omitted | Not configurable | Contains `omitted_row_count`, `selection_rule_id`, `first_omitted_sort_key`, and `filter_summary`. |

**REQ-RPT-074a**
When more rows are eligible than the `Selected timeline rows` hard limit, the timeline `selection_rule_id` determines behavior. When the template declares no `selection_rule_id`, the default is `timeline_overflow_summary_v1`, which keeps the first `240` rows in `select_timeline_rows_v1` order (REQ-RPT-073) and emits the overflow block. Behavior is scope-graded:

- `internal_draft` and `internal_review`: the default emits the overflow block and MUST NOT fail on volume.
- `external_release`: overflow MUST fail with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='timeline_overflow_unresolved'` unless the template explicitly declares a summarizing `selection_rule_id` (for example `timeline_overflow_summary_v1`), so that an external report never silently drops case content without a declared rule.

The overflow block required by Table 14-C records `omitted_row_count`, `selection_rule_id`, `first_omitted_sort_key`, and `filter_summary` in every scope in which rows are omitted.

**REQ-RPT-075**
Chunked timeline slide IDs MUST use `{base_slide_id}__chunk_{0001..N}`. Slide titles MUST append ` ({chunk_index}/{chunk_count})` when `chunk_count>1`.

The same chunk identity and title rules apply to block-count chunking in §17.
The base ID uses the first emitted chunk's final slide ordinal. Subsequent
slides continue the final ordinal sequence. Notes attached to the unchunked
slide by `set_speaker_notes` appear on its first chunk only; speaker-note blocks
stay with the chunk containing those blocks. Chunking MUST preserve every block
exactly once in its prior order and MUST NOT modify block identities or content.

# 15. Graph Projection consumption and diagram model

## 15.1 Reporting-to-Graph Projection adapter

**REQ-RPT-076**
Reporting consumes immutable Graph Projection v2 output through
`source_projection_ref.v2`. Reporting MUST NOT redefine projection derivation,
projected identity, or Graph validation.

Release admission MUST dispatch every Graph binding through the typed
source-owner registry. The current registry contains exactly
`network_flow_activity`. Its adapter validates incident ownership, active
declaration visibility, selected exact result, alternate immutable source
boundary, and all digests using the same database transaction and visibility
snapshot that admits the release tuple. In that transaction Reporting acquires
a durable result lease before the render job can observe the payload. An
implementation MUST NOT validate through an independent pool read or re-read a
mutable latest declaration after tuple admission.

For Network Flow bindings, lease admission MUST lock the immutable projection
result before locking or validating the saved declaration. Cleanup and Network
Flow publication replay MUST use the same result-before-declaration lock order.
Reporting MUST NOT reverse that order, admit a lease from an unlocked latest-
result lookup, or treat declaration reachability as a substitute for the
durable result lease.

**REQ-RPT-077**
`source_projection_ref.v2` MUST use Table 15-A.

**Table 15-A. `source_projection_ref.v2`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `source_owner_id` | `identifier` | Yes | No | None | Exactly one registered Reporting graph source owner. |
| `graph_view_id` | `identifier` | Yes | No | None | Graph Projection view identity. |
| `projection_result_id` | `identifier` | Yes | No | None | Exact immutable Graph Projection result. |
| `source_snapshot_id` | `identifier` | Yes | No | None | Must equal the Reporting `snapshot_id` or a Core-declared alternate immutable source-boundary token. |
| `projection_schema_id` | `identifier` | Yes | No | None | Exactly `graph_projection.v2`. |
| `projection_version` | `identifier` | Yes | No | None | Exact projection version. |
| `normalized_configuration_sha256` | `sha256_hex` | Yes | No | None | Exact normalized configuration digest. |
| `normalized_source_sha256` | `sha256_hex` | Yes | No | None | Exact normalized source digest. |
| `canonical_output_sha256` | `sha256_hex` | Yes | No | None | Exact canonical output digest. |

**REQ-RPT-078**
Graph adapter failures MUST use Table 15-B.

**Table 15-B. Graph adapter failures**

| Condition | `failure_code` | `reason_code` |
| --- | --- | --- |
| Projection snapshot mismatch | `graph_projection_unavailable` | `graph_projection_stale` |
| Selected result absent | `graph_projection_unavailable` | `graph_projection_not_completed` |
| Graph view not bound in release tuple | `graph_projection_unavailable` | `graph_projection_not_bound` |
| Graph view bound by more than one tuple item | `graph_projection_unavailable` | `graph_projection_ambiguous` |
| Required vertex or edge missing | `graph_projection_unavailable` | `graph_projection_selection_unresolved` |
| Projection output digest mismatch | `graph_projection_unavailable` | `graph_projection_digest_mismatch` |
| Duplicate diagram selection input ref | `graph_projection_unavailable` | `diagram_selection_duplicate_ref` |
| Missing diagram selection input ref | `graph_projection_unavailable` | `diagram_selection_missing_ref` |

## 15.2 Diagram selection rules

**REQ-RPT-079**
Diagram selection rules MUST use Table 15-C. A graph-derived diagram MUST name `source_graph_view_id` and MUST resolve that value through release tuple `graph_projection_refs[]` under REQ-RPT-027g before diagram model validation. If no graph view is named, the diagram source kind MUST be `timeline`. `template_static` is future-only in Reporting v1 and MUST fail template or composition validation with `reason_code='template_static_future_only'` before render output bytes. Reporting MUST NOT construct ad hoc Graph Projection input inside export-model materialization or render execution. It consumes only the exact leased tuple-bound result or fails with Table 15-B.

**Table 15-C. Diagram selection rules**

| Rule | Required inputs | Defaults and bounds | Ordering |
| --- | --- | --- | --- |
| `explicit_refs` | `vertex_refs[]`, `edge_refs[]` | Empty arrays invalid. | Graph Projection output order. |
| `neighborhood` | `seed_vertex_refs[]` | `depth` default `1`, max `2`; `edge_kind_filter[]` default `[]`. | Graph Projection traversal order. |
| `timeline_sequence` | `timeline_record_ids[]` | Non-empty; all IDs must exist in selected timeline. | §14.2 order. |
| `all_with_bounds` | None | Must fit §16.1 Mermaid bounds. | Graph Projection output order. |

**REQ-RPT-079a**
`diagram_selection_rule.v1` MUST use Table 15-C1. Unknown members are invalid. Members not applicable to the selected `rule_id` MUST be present with the stated empty or null value; they MUST NOT be omitted from canonical bytes.

**Table 15-C1. `diagram_selection_rule.v1` schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `diagram_selection_rule.v1`. |
| `rule_id` | string | Yes | No | None | `explicit_refs`, `neighborhood`, `timeline_sequence`, or `all_with_bounds`. |
| `vertex_refs[]` | array | Yes | No | `[]` | Used only by `explicit_refs`; duplicate values invalid. |
| `edge_refs[]` | array | Yes | No | `[]` | Used only by `explicit_refs`; duplicate values invalid. |
| `seed_vertex_refs[]` | array | Yes | No | `[]` | Used only by `neighborhood`; non-empty and duplicate-free. |
| `depth` | `finite_integer` | Yes | Yes | `1` | Used only by `neighborhood`; must be `1` or `2`; null for other rules. |
| `edge_kind_filter[]` | array | Yes | No | `[]` | Used only by `neighborhood`; empty means all edge kinds; sorted bytewise ascending when present. |
| `timeline_record_ids[]` | array | Yes | No | `[]` | Used only by `timeline_sequence`; non-empty and duplicate-free. |
| `overflow_policy` | string | Yes | No | `fail` | `fail` or `summarize`. `summarize` is invalid for external release unless the template declaration is explicit. |

**REQ-RPT-079b**
Diagram selection MUST use the deterministic algorithms in Table 15-C2. A referenced vertex, edge, or timeline record that is absent from the applicable source MUST fail with `failure_code='graph_projection_unavailable'` and `reason_code='diagram_selection_missing_ref'`. A duplicate input ref MUST fail with `failure_code='graph_projection_unavailable'` and `reason_code='diagram_selection_duplicate_ref'`. If selected content exceeds §16.1 bounds and `overflow_policy='fail'`, selection fails with `failure_code='mermaid_invalid'` and `reason_code='diagram_hard_limit_exceeded'`. If selected content exceeds bounds and `overflow_policy='summarize'`, the implementation MUST keep the first bounded set in selection order, omit the remainder, and emit `diagram_overflow_summary.v1`; external release may use this behavior only when the template declaration explicitly sets `overflow_policy='summarize'`.

**Table 15-C2. Diagram selection algorithms**

| Rule | Algorithm |
| --- | --- |
| `explicit_refs` | Validate every input ref against the completed projection, then emit `included_vertex_refs[]` as the subset of projection vertices whose refs appear in `vertex_refs[]`, ordered by projection output order. Emit `included_edge_refs[]` as the subset of projection edges whose refs appear in `edge_refs[]`, ordered by projection output order. Every selected edge endpoint MUST be included in `included_vertex_refs[]`. |
| `neighborhood` | Initialize the included vertex set with `seed_vertex_refs[]` sorted by projection output order. For each depth level from `1` through `depth`, scan projection edges in projection output order and include an edge when exactly one endpoint is already included and `edge_kind_filter[]` is empty or contains the edge kind. The opposite endpoint becomes newly included for the next level. After each level, append newly included vertices in projection output order. A vertex or edge already included is not emitted again. |
| `timeline_sequence` | Sort `timeline_record_ids[]` by the corresponding selected timeline events' §14.2 order. Emit one vertex per timeline record in that order and one directed edge between each adjacent pair. This rule does not require Graph Projection output and MUST set `source_projection_ref=null`. |
| `all_with_bounds` | Emit all projection vertices and edges in projection output order, subject to §16.1 bounds. |

Projection output order means the canonical vertex and edge array order inside the digest-bound Graph Projection output. If the adopted Graph Projection owner exposes only unordered vertex or edge maps, Reporting MUST derive projection output order by sorting refs bytewise ascending and MUST record that ordering mode in `safe_details.ordering_mode='bytewise_ref_sort'` for diagram-model validation.

Graph labels, properties, and mapped metadata are source-derived values, not
inherently safe presentation text. Reporting MUST first map every selected value
to a typed derived-analytic field with source owner, source object ref,
classification, disclosure partitions, and redaction behavior. It then applies
disclosure filtering and redaction before deriving labels, Mermaid, SVG, Slidev,
or validation samples. A removed vertex removes every dependent edge. A
transformed identifier is used consistently for its vertex and remaining edge
endpoints. External release fails closed when any selected source value lacks a
classification or supported redaction action. Raw graph values and replacement
values MUST NOT appear in safe details or the redaction manifest.

The Network Flow source-owner adapter MUST provide graph label inputs only as
typed redaction candidates. A vertex candidate contains the source endpoint
component and its source classification. A default-edge candidate contains
protocol and destination-port presence/value components. A temporal-edge
candidate additionally contains canonical bucket start and end components.
Each component MUST carry a stable field path, source object reference,
classification, disclosure partitions, and applicable redaction behavior.
Graph properties, Graph Projection labels, canonical selectors, and Network
Flow source values MUST NOT be marked release-safe merely because they were
admitted to a graph result.

The pre-redaction graph selection has a closed decision address for each typed
label component: `$.diagrams['<diagram_id>'].nodes['<vertex_ref>'].label_components['endpoint_value']`
or `$.diagrams['<diagram_id>'].edges['<edge_ref>'].label_components['<component_kind>']`.
The edge component vocabulary is `protocol`, `destination_port`,
`bucket_start_utc`, and `bucket_end_utc`; the latter two occur only for temporal
edges. Keys use the exact path escaping contract. These are addresses in the
selected pre-redaction diagram input, not source-owner paths or renderer IDs.
Components are derived scalar candidates, with null for a typed absent, removed,
or missing input. They use the same exact-path, content-class, and profile-default
selection as other fields, including complete value-free manifest entries and
post-action partition checks. Dropping a label component leaves the selected
topology intact and permits only the specified safe label fallback; it is not a
vertex-drop rule. An unselected component cannot satisfy an exact path rule.
The final diagram retains labels and topology, not the pre-redaction components.

Network Flow graph-source v2 supplies no Core subject mapping. Reporting MUST
NOT infer one from IP text, a Graph label, or a source object reference. A
token-backed rule therefore fails the ordinary unique-subject binding check.
Under this exact contract a structured label override targets a non-tokenized
vertex or edge by its opaque selected ref and remains subject to the existing
composition text and label safety rules. A future source contract introducing
subject mappings must also implement the token-priority and override-denial
rules; it cannot inherit this v2 absence of mappings.

For internal diagrams, `derive_diagram_label_v1` MUST construct endpoint labels
from deterministic post-redaction endpoint components. A temporal edge label
MUST contain canonical `[start_utc,end_utc)`, post-redaction protocol, and
post-redaction destination port only; an absent port remains a typed absence
and is not rendered as a raw sentinel. A component removed by disclosure or
redaction is omitted, and an item with no remaining label component uses the
deterministic ordinal fallback from Table 15-C4. It MUST NOT fall back to an
unredacted source value, Graph property, selector field, digest, or raw Graph
label.

The current contract adds no external Network Flow graph allow rule. An external release that
selects any Network Flow graph field MUST fail closed through the existing
unresolved-redaction failure unless a separately adopted Reporting policy
classifies and resolves every selected field. Internal-policy success MUST NOT
be reused as external admission.

Reporting renews the lease while the render attempt can still read the result
and releases it only after a durable terminal outcome no longer depends on Graph
rows. Retry reuses the exact reference and lease purpose. Cancellation, crash,
and expired-worker reconciliation release orphaned leases only after the durable
job state proves no active reader. Cleanup MUST treat uncertain lease state as
reachable.

Handler input reads and Graph-result lease renewal MUST validate the same live
Jobs execution under the Jobs transition lock, in the caller-owned transaction
that copies the immutable bytes. A failed input or result read, cancellation,
expired lease or replaced attempt MUST roll back renewal. The handler MUST
revalidate lease liveness before returning the copied input. Rendering MUST run
after that transaction closes and MUST NOT retain its database locks across
renderer execution. Recovery readers remain separate read-only operations and do
not authorize handler execution or renew leases.

**REQ-RPT-079c**
`diagram_overflow_summary.v1` MUST use Table 15-C3. Unknown members are invalid. It is null when no diagram content is omitted.

**Table 15-C3. `diagram_overflow_summary.v1` schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `diagram_overflow_summary.v1`. |
| `omitted_vertex_count` | `finite_integer` | Yes | No | `0` | Count of selected vertices omitted by bounds. |
| `omitted_edge_count` | `finite_integer` | Yes | No | `0` | Count of selected edges omitted by bounds. |
| `selection_rule_id` | string | Yes | No | None | Rule that omitted content. |
| `first_omitted_ref` | string | Yes | Yes | None | First omitted vertex or edge ref in selection order; null when count is zero. |
| `filter_summary` | `safe_string` | Yes | Yes | None | Redaction-safe summary or null. |

**REQ-RPT-079d**
`derive_diagram_label_v1` MUST choose labels using Table 15-C4 before Table 16-B normalization and Table 16-C escaping. If the selected source is null or unavailable at every priority level, the label is the family token plus the 1-based selection ordinal, for example `Vertex 0001` or `Message 0001`. A label source that is not releasable after redaction MUST be skipped rather than used and redacted later.

**Table 15-C4. Diagram label-source priority**

| Diagram item | Priority order |
| --- | --- |
| Graph vertex mapped to a tokenized subject | `display_token`; then post-redaction record `display_name`; then Graph Projection safe label; then ordinal fallback. |
| Graph vertex not mapped to a tokenized subject | Post-redaction record `display_name`; then Graph Projection safe label; then ordinal fallback. |
| Graph edge | Post-redaction relationship kind label; then Graph Projection safe edge label; then ordinal fallback. |
| Timeline-sequence vertex | `display_times.primary_display`; then `record_id`; then ordinal fallback. |
| Timeline-sequence edge | Literal `next` unless the template declares an edge label literal; the literal MUST be `safe_string`. |

Timeline diagrams are never synthesized from arbitrary report fields. With no
retained template or composition diagram declaration, Mermaid output fails with
`mermaid_invalid` / `no_diagrams_selected` before renderer execution. A Timeline
record removed by disclosure or redaction cannot satisfy a selection reference.
Until a template display binding explicitly selects a retained time field, its
`display_times` members remain null and the vertex uses the record-ID fallback.
Core ordering facts are not display bindings.

Timeline `included_vertex_refs[]` contains exact selected record IDs in selection
order. Each adjacent pair has an opaque, directional edge reference:
`timeline-edge-` followed by the Reporting object digest of the closed object
`{"schema_id":"cartulary.reporting_timeline_adjacency.v1","from_record_id":<id>,"to_record_id":<id>}`.
The preimage uses canonical Reporting JSON and the schema-domain digest rule;
no delimiter parsing or renderer ordinal participates in identity. This object
is an identity preimage, not another persisted graph or source record. Bounds
retain a prefix, so retained adjacency identities stay unchanged when later
vertices are omitted. Flowcharts admit at most 80 vertices; sequences admit at
most 24 participants; both admit at most 160 edges. A composition declaration
alone cannot authorize the external template-owned summarization exception.

**REQ-RPT-079e**
When the release tuple carries a non-null composition, Reporting MUST admit composition-owned diagram declarations and composition label overrides only through the closed consumer rules in this requirement. A composition-owned diagram declaration MUST map to `cartulary.reporting_diagram.v1`, MUST set `diagram_id` to the composition declaration `decl_id`, MUST use one Table 15-C selection rule, MUST carry the companion-owned `layout_mode` and `composition_diagram_layout.v1` when present, and MUST validate against the same completed digest-bound projection or selected timeline input as a template-owned diagram declaration.

Diagram replacement MUST execute before duplicate `diagram_id` validation. If a composition-owned diagram declaration has `replaces_decl_id=null`, its `decl_id` MUST NOT equal any template-owned or composition-owned retained diagram ID. If `replaces_decl_id` is non-null, `replaces_decl_id` MUST equal `decl_id`, MUST resolve to exactly one template-owned diagram declaration, and Reporting MUST remove that template-owned declaration before inserting the composition-owned declaration. A missing replacement target fails with `failure_code='composition_invalid'` and `reason_code='composition_replacement_target_missing'`. Replacing a composition-owned declaration, replacing more than one declaration, or replacing a template declaration under a different ID fails with `failure_code='composition_invalid'` and `reason_code='diagram_layout_invalid'` unless the companion validation fails earlier with a narrower code. After replacement, any duplicate `diagram_id` values across retained template-owned and composition-owned diagrams fail with `failure_code='composition_invalid'` and `reason_code='duplicate_stable_id'`.

Composition label overrides are presentation hints applied before Table 15-C4 label fallback and before Table 16-B normalization. Each companion-owned structured override target MUST resolve to exactly one selected vertex or edge in the resolved diagram by exact opaque ref equality; Reporting MUST NOT parse colon-delimited target strings or accept generated Mermaid IDs, rendered SVG IDs, or DOM selectors as override targets. Override values MUST be `safe_string`, MUST satisfy Table 16-A label bounds, and MUST NOT contain LF. A label override targeting a vertex mapped to a tokenized subject is invalid and MUST fail with `failure_code='composition_invalid'` and `reason_code='diagram_label_override_invalid'`; the visible label for that vertex remains governed by `display_token` priority in Table 15-C4. Raw `.mmd`, arbitrary node declarations, arbitrary edge declarations, renderer-local Mermaid syntax, and graph-projection mutation are invalid composition inputs and MUST fail with `failure_code='composition_invalid'` and `reason_code='invalid_mermaid_construct'` or `reason_code='diagram_selection_missing_ref'` when that narrower reason is attributable.

Composition manual layout is presentation data applied after diagram selection and before diagram rendering. Reporting imports `composition_diagram_layout.v1` from the Report Composition NLSpec and MUST NOT redefine its schema. Manual layout targets MUST resolve by exact opaque ref equality against `included_vertex_refs[]` and `included_edge_refs[]`. Reporting MUST reject generated Mermaid IDs, rendered SVG IDs, DOM IDs, labels, array indexes, React Flow IDs, React Flow handles, or endpoint-changing route data with `failure_code='composition_invalid'` and the narrowest applicable layout reason code in Table 23-F.

## 15.3 Diagram object

**REQ-RPT-080**
`cartulary.reporting_diagram.v1` MUST include Table 15-D fields.

**Table 15-D. Diagram schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_diagram.v1`. |
| `diagram_id` | `identifier` | Yes | No | None | Template declaration `decl_id`; unique in the export model. |
| `diagram_kind` | string | Yes | No | None | `flowchart` or `sequence`. |
| `diagram_source_kind` | string | Yes | No | None | `graph` or `timeline`. `template_static` is future-only in v1. |
| `source_projection_ref` | `source_projection_ref.v2` | Yes | Yes | None | Required when `diagram_source_kind='graph'`; otherwise null. |
| `selection_rule` | `diagram_selection_rule.v1` | Yes | No | None | One closed rule from Table 15-C1 with its required inputs. |
| `included_vertex_refs[]` | array | Yes | No | `[]` | Ordered by the selection rule. |
| `included_edge_refs[]` | array | Yes | No | `[]` | Ordered by the selection rule. |
| `nodes[]` | array of closed node objects | Yes | No | `[]` | Each object contains exactly string `node_id` and string `label`. Node IDs equal `included_vertex_refs[]` in the same order. Labels contain only retained values selected under REQ-RPT-079d. |
| `edges[]` | array of closed edge objects | Yes | No | `[]` | Each object contains string `edge_id`, `from`, and `to`, plus optional string `label`. Edge IDs equal `included_edge_refs[]` in the same order; endpoints name retained nodes. Omitted label means empty and an empty label is omitted canonically. |
| `layout_mode` | string | Yes | No | `auto` | `auto` or `manual`; template-owned diagrams materialize `auto` in this revision. |
| `composition_layout` | Report Composition `composition_diagram_layout.v1` | Yes | Yes | None | Non-null only when `layout_mode='manual'`; schema ownership remains with the Report Composition NLSpec. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `overflow_summary` | `diagram_overflow_summary.v1` | Yes | Yes | None | Null unless content was omitted by a declared bound. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-080a**
Rendered diagram outputs MUST be located only by the §22 bundle path convention: `diagrams/{diagram_id}.svg` for SVG. `diagrams/{diagram_id}.png` is reserved future-only vocabulary and is not admitted by the current renderer profile. The diagram object MUST NOT include post-render file references in the pre-render export model.

**REQ-RPT-080b**
Diagram rendering mode MUST follow Table 15-E. Reporting MUST NOT silently convert a manual-layout diagram to Mermaid auto-layout. A retained diagram whose `layout_mode` is unsupported by the selected output kind MUST fail before approvable bytes with `failure_code='composition_invalid'` and `reason_code='manual_layout_not_supported_for_output_kind'`.

**Table 15-E. Diagram rendering modes**

| `layout_mode` | Diagram source artifact | Render path | Valid output kinds |
| --- | --- | --- | --- |
| `auto` | `diagrams/{diagram_id}.mmd` | Existing `mermaid_source_serialize_v3` then Mermaid render. | `mermaid`, `slidev`. |
| `manual` | No raw Mermaid source; composition bytes are the source authority. | `diagram_layout_svg_serialize_v1`, then local SVG embedding or bundling. | `slidev`; standalone manual-diagram export is future-only. |

# 16. Diagram source and SVG generation

## 16.1 Mermaid bounds

**REQ-RPT-081**
Reporting v1 supports only Mermaid `flowchart` and `sequenceDiagram` sources. It MUST reject all other Mermaid diagram types.

**REQ-RPT-082**
Mermaid diagram bounds MUST use Table 16-A. Reporting v1 does not define graph chunking. Flowchart nodes or edges beyond hard limits MUST fail.

**Table 16-A. Mermaid bounds**

| Limit | Default | Hard limit | Failure |
| --- | ---: | ---: | --- |
| Flowchart nodes | 80 | 80 | `failure_code='mermaid_invalid'`, `reason_code='diagram_hard_limit_exceeded'` |
| Flowchart edges | 160 | 160 | same |
| Sequence participants | 24 | 24 | same |
| Sequence messages | 160 | 160 | same |
| Label Unicode scalar values | 120 | 240 | Truncation and failure per `truncate_label_v1` (REQ-RPT-082a); over the hard limit fails with `reason_code='label_length_exceeded'`. |

**REQ-RPT-082a**
Label truncation MUST use `truncate_label_v1`, applied after the Table 16-B normalization (NFC and whitespace collapse) and before Table 16-C escaping. Let `L` be the label length in Unicode scalar values, `D=120` the default limit, and `H=240` the hard limit; the algorithm follows Table 16-A1. Truncation cuts strictly at a Unicode scalar-value boundary and appends U+2026 (HORIZONTAL ELLIPSIS) as the final scalar, so the result is exactly `D` scalar values. Grapheme-cluster-aware and word-aware truncation are forbidden because they vary across Unicode library versions and would break byte determinism.

**Table 16-A1. Label length handling**

| Condition | Behavior |
| --- | --- |
| `L ≤ D` | Emit the label unchanged. |
| `D < L ≤ H` and the template declares truncation for the diagram | Keep the first `D − 1` scalar values and append U+2026, yielding exactly `D` scalar values. |
| `D < L ≤ H` and the template declares no truncation | Fail with `failure_code='mermaid_invalid'`, `reason_code='label_length_exceeded'`. |
| `L > H` | Fail with `failure_code='mermaid_invalid'`, `reason_code='label_length_exceeded'`. |

## 16.2 Canonical Mermaid serializer

**REQ-RPT-083**
`mermaid_source_serialize_v3` MUST use the shared serialization rules in Table 16-B.

**Table 16-B. Shared Mermaid serialization**

| Property | Required value |
| --- | --- |
| Encoding | UTF-8. |
| Line endings | LF only. |
| Final newline | Required. |
| Indentation | Two ASCII spaces for declarations after header. |
| Label normalization | Unicode NFC, no C0/C1 controls, collapse all Unicode whitespace runs to U+0020. |
| Comments | Forbidden. |
| Raw HTML | Forbidden. |
| Mermaid directives | Forbidden. |
| Mermaid init blocks | Forbidden. |
| Security profile | Renderer MUST use strict Mermaid security configuration that disables HTML labels, raw HTML, and click functionality; a wrapper configuration is valid only when it rejects HTML and click functionality before render and emits the same failure codes as this section. |

**REQ-RPT-084**
Flowchart source MUST use this grammar exactly after variable substitution:

```text
flowchart TD
  {node_id}["{escaped_label}"]
  {from_node_id} -->|"{escaped_label}"| {to_node_id}
```

For an edge whose normalized label is empty, v3 MUST instead emit
`  {from_node_id} --> {to_node_id}`. Quoted empty edge labels are invalid in v3.
This replaces `mermaid_source_serialize_v2`; v2 source evidence does not qualify
v3 serialization. Nonempty labels retain the quoted form above.

`TD` is the default direction. A template MAY declare `LR`; no other direction is valid. Node declarations sort by selected vertex order. Edge declarations sort by selected edge order.

**REQ-RPT-084a**
Mermaid source identifiers MUST be ordinal and derived only from diagram selection order. For flowcharts, `node_id = 'n' + zeropad4(i)` where `i` is the 1-based position of the vertex in `included_vertex_refs[]`. Edge declarations reference the node IDs of their endpoints. An edge whose endpoint is absent from `included_vertex_refs[]` MUST fail with `failure_code='mermaid_invalid'` and `reason_code='invalid_mermaid_construct'`. For sequence diagrams, `participant_id = 'p' + zeropad4(i)` over `included_vertex_refs[]`; participant declarations emit in that order; message lines emit in `included_edge_refs[]` order, with `escaped_message` derived from the edge label after Table 16-B normalization and Table 16-C escaping. The §16 Mermaid hard limits guarantee four digits suffice. The vertex-to-Mermaid-ID mapping is fully determined by order and MUST NOT be persisted as a separate artifact.

**REQ-RPT-085**
Sequence source MUST use this grammar exactly after variable substitution:

```text
sequenceDiagram
  participant {participant_id} as {escaped_label}
  {from_participant_id}->>{to_participant_id}: {escaped_message}
```

Only `participant` declarations and `->>` messages are valid in v1. Mermaid notes, loops, boxes, links, actor menus, autonumber, styling, activation, comments, and other arrows are future-only.

**REQ-RPT-086**
Mermaid escaping MUST use Table 16-C.

**Table 16-C. Mermaid escaping (`mermaid_source_serialize_v3`)**

| Input | Serialized output or validation behavior |
| --- | --- |
| Quotation mark U+0022 | `#34;` |
| Number sign U+0023 | `#35;` |
| Percent sign U+0025 | `#37;` |
| Ampersand U+0026 | `#38;` |
| Apostrophe U+0027 | `#39;` |
| Semicolon U+003B | `#59;` |
| Backslash U+005C | `#92;` |
| Backtick U+0060 | `#96;`, except a code fence remains invalid before escaping. |
| Vertical bar U+007C | `#124;` |
| A boundary U+0020 after normalization | `#32;` at the beginning or end of a label; interior spaces remain literal. |
| LF, CR, TAB | U+0020 after whitespace collapse. |
| C0/C1 controls | Validation failure. |
| `<`, `>` | A normalized label containing U+003C or U+003E fails with `mermaid_invalid` / `invalid_mermaid_construct`; no raw-HTML heuristic is used. |

Escaping MUST visit normalized Unicode scalars once; newly emitted entity text
MUST NOT be re-escaped. Literal input that resembles an entity remains literal:
`#34;` serializes as `#35;34#59;`. Current rendering MUST reject the retired v1
backslash-quote grammar. The v2 sequence participant alias is unquoted because
Mermaid treats those quotation marks as visible alias text. Flowchart labels
retain the quotation delimiters shown in Table 16-B and REQ-RPT-084.

The pinned renderer MUST decode the closed entity set into inert SVG text before
text measurement and layout, exactly once and after source tokenization. It MUST
NOT interpret the resulting text as HTML, Markdown, Mermaid, or a line-break
escape. Thus a literal backslash followed by `n` remains two visible characters.
The emitted labels MUST preserve the normalized source text. A vendor correction
used to meet this rule is part of the exact package-store and lockfile bindings
in §20; a version string without those bindings is insufficient. The serializer
identity change does not authorize reading or translating retired source formats.

**REQ-RPT-086a**
`render_mermaid_bundle_v1` MUST emit one source file for every retained auto-layout diagram after disclosure filtering, redaction, diagram selection, and render-mode validation. Retained auto-layout diagrams sort bytewise ascending by `diagram_id`; each retained auto-layout diagram emits exactly one `diagrams/{diagram_id}.mmd` file whose bytes are produced by `mermaid_source_serialize_v3`. A conforming implementation MUST NOT concatenate multiple diagrams into one `.mmd` file, drop a retained auto-layout diagram because another diagram rendered successfully, or synthesize a placeholder diagram for an empty selection. A retained manual-layout diagram for `output_kind='mermaid'` MUST fail before source serialization with `failure_code='composition_invalid'` and `reason_code='manual_layout_not_supported_for_output_kind'`.

For `release_scope='external_release'` and `output_kind='mermaid'`, every retained auto-layout diagram MUST also emit exactly one `diagrams/{diagram_id}.svg` rendered under the §16 security profile. For internal scopes with `source_only=true`, rendered SVG and PNG auto-layout diagram artifacts are omitted and only the `.mmd` files plus required validation and manifest artifacts are emitted. Mermaid PNG output is future-only; the current profile rejects its request before rendering. If no diagrams remain after filtering, selection, and render-mode validation for `output_kind='mermaid'`, render MUST fail before publishable bundle bytes with `failure_code='mermaid_invalid'` and `reason_code='no_diagrams_selected'`.

**REQ-RPT-086b**
`diagram_layout_svg_serialize_v1` MUST serialize a retained manual-layout flowchart from the post-redaction resolved `cartulary.reporting_diagram.v1` plus the companion-owned `composition_diagram_layout.v1`. Labels are derived by `derive_diagram_label_v1`, then normalized and bounded under Table 16-A, Table 16-A1, and the Table 16-B label rules before SVG serialization. Node order follows `included_vertex_refs[]`. Edge order follows `included_edge_refs[]`. Node boxes render at exact integer `x`, `y`, `width`, and `height` from `node_positions[]`. Edge paths render only selected existing edges; omitted edge routes use deterministic straight routing.

**REQ-RPT-086c**
`diagram_layout_svg_serialize_v1` MUST use a deterministic endpoint algorithm. For each edge, build the route point list as source node center, declared interior waypoints in authored order, and target node center. A node center is `(x + floor(width / 2), y + floor(height / 2))`. For the source endpoint, compare the source center to the next route point. For the target endpoint, compare the target center to the previous route point. If `dx=0` and `dy=0`, the route is invalid and fails with `failure_code='composition_invalid'` and `reason_code='diagram_layout_invalid'`. If `abs(dx) >= abs(dy)`, choose the right side when `dx > 0` and the left side when `dx < 0`; the endpoint y coordinate is `y + floor(height / 2)`. If `abs(dx) < abs(dy)`, choose the bottom side when `dy > 0` and the top side when `dy < 0`; the endpoint x coordinate is `x + floor(width / 2)`. SVG path data MUST emit an `M` command for the source endpoint, `L` commands for every waypoint, and one final `L` command for the target endpoint. Numeric output MUST be base-10 integer text with no leading plus sign, decimal point, exponent, or negative zero.

The SVG bytes MUST be UTF-8 with LF line endings, deterministic element ordering, deterministic attribute ordering, no comments, no script, no event-handler attributes, no `foreignObject`, no external references, no remote assets, no `data:` URLs, and no renderer-local syntax. The root SVG MUST use `viewBox="0 0 {coordinate_space.width} {coordinate_space.height}"`. Manual-layout SVG output MUST pass the rendered-output security checks and external-release determinism validation in §24 before bundle persistence.

The v1 manual serializer uses the following fixed presentation profile. The root
attributes are `xmlns`, then `viewBox`. Each child is indented by two ASCII
spaces and followed by LF; the closing root is also followed by LF. In selected
edge order, emit a `path` with attributes `d`, `fill="none"`,
`stroke="#334155"`, `stroke-width="2"`. Commands are `M{x} {y}` followed by
` L{x} {y}` for every remaining route point. Immediately after each path,
emit its marker-free arrow `polygon`, with attributes `points`, then
`fill="#334155"`. Let `E` be the clipped target endpoint and `U` the outward
axis unit vector selected by REQ-RPT-086c for that endpoint (left `(-1,0)`,
right `(1,0)`, top `(0,-1)`, bottom `(0,1)`). Its three points are `E`,
`E+6U+3(-Uy,Ux)`, and `E+6U-3(-Uy,Ux)`, serialized as six space-separated
integers in that order. This preserves edge direction without SVG marker IDs.
For a nonempty normalized edge
label, emit `text` at the integer midpoint of the two clipped endpoints, rounding
each coordinate down. Its attributes are `x`, `y`, `font-family="sans-serif"`,
`font-size="12"`, `fill="#0f172a"`. Then, in selected vertex order, emit a
`rect` with `x`, `y`, `width`, `height`, `fill="#ffffff"`, `stroke="#334155"`,
`stroke-width="1"`, followed by its centered label `text` with `x`, `y`,
`font-family="sans-serif"`, `font-size="14"`, `text-anchor="middle"`,
`fill="#0f172a"`. The rect/path/polygon tags are self-closing. Text uses literal UTF-8
except XML escaping: `&`, `<`, `>`, double quote and apostrophe become
`&amp;`, `&lt;`, `&gt;`, `&#34;` and `&#39;`. No theme, host defaults, current
time or generated renderer IDs enter these bytes. This profile defines first
implementation bytes; it does not reinterpret an already published manual
serializer.

A manual SVG is deterministic source material for a Slidev diagram reference,
including a source-only deck. It remains a required `rendered_svg` bundle member
even when browser rendering is disabled. Its presence alone does not assert
browser execution. Auto-layout source-only decks retain Mermaid source without
SVG. A native Slidev render receives only the exact retained manual SVG bytes
at `diagrams/{diagram_id}.svg`; it MUST reject substituted bytes, duplicate or
escaping paths and active SVG constructs before launch. The private stage
request/result v4 bind local assets independently of source text. Retired stage
v1/v2/v3 requests and results have no reader fallback. Completion and recovery
recompute manual SVG and source/model members from frozen inputs; matching
hashes among substituted bundle files alone do not establish fidelity.

# 17. Slide-deck model

**REQ-RPT-087**
`cartulary.reporting_slide_deck.v2` MUST use Table 17-A.

**Table 17-A. Slide-deck schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_slide_deck.v2`. |
| `deck_id` | `generated_id` | Yes | No | None | Generated ID from Table 10-B. |
| `release_id` | `identifier` | Yes | No | None | From release tuple. |
| `snapshot_id` | `identifier` | Yes | No | None | From release tuple. |
| `template_id` | `identifier` | Yes | No | None | From release tuple. |
| `template_version` | `identifier` | Yes | No | None | Exact template version. |
| `title` | `safe_string` | Yes | No | None | Deck title derived by REQ-RPT-087c. |
| `output_options` | `cartulary.reporting_render_request_options.v1` | Yes | No | None | Materialized options from §7.5. |
| `slides[]` | array | Yes | No | `[]` | Ordered by `slide_ordinal`. |
| `slide_count` | `finite_integer` | Yes | No | None | Equals `slides[].length`. |
| `click_step_count` | `finite_integer` | Yes | No | `0` | Total count across slides. |
| `expected_export_page_count` | `finite_integer` | Yes | No | None | Sum over slides of `1 + click_steps.length`. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |

**REQ-RPT-087a**
For `output_kind='slidev'`, `derive_deck_v1` MUST derive the deck model from the post-redaction export model as follows:

1. Take emitted sections in export-model order: `ordering_key`, then `section_id`.
2. Skip sections whose template declaration sets `include_in_deck=false`; omission defaults to `true`. `validation` sections are included only for internal scopes.
3. For each remaining section instance, emit one base slide whose `title` is the section `title`, whose `layout_id` is the declaration's `slide_layout_id`, and whose `blocks[]` are the section's top-level blocks in `block_ordinal` order. If `slide_layout_id` is omitted, the default is `cover` for the first emitted slide of the deck and `default` for every later emitted slide.
4. Split any base slide whose block count exceeds the operative `blocks_per_slide` limit or whose `timeline_rows` block exceeds the operative rows-per-timeline-slide limit. Blocks are distributed in ordinal order, filling each chunk to the operative limit before starting the next. Chunk IDs and titles follow REQ-RPT-075.
5. Generate click steps per slide from the declaration's `click_profile` under REQ-RPT-087b.
6. Assign `slide_ordinal` contiguously from `1` over the final slide list. Base `slide_id` values follow REQ-RPT-040b; chunk slides retain the `__chunk_` suffix. `slide_count`, `click_step_count`, and `expected_export_page_count` are then computed from Table 17-A.

Two conforming implementations given the same post-redaction export model and template manifest MUST produce byte-identical canonical deck models.

**REQ-RPT-087c**
The deck model `title` MUST equal the template manifest `deck_title`. If the template omits `deck_title`, the default is the literal `Cartulary Report`. `deck_title` is template-owned presentation text, not case content; v1 permits no title placeholders, incident-field interpolation, date interpolation, or renderer-local title derivation. A title containing C0/C1 controls, LF, or more than `120` Unicode scalar values fails with `failure_code='deck_model_invalid'` and `reason_code='deck_title_invalid'`.

**REQ-RPT-087b**
`click_profile` is a closed per-section-declaration token with the values `none`, `reveal_blocks`, and `reveal_list_items`.

- `none` is the default and emits `click_steps[]=[]`.
- `reveal_blocks` emits one `reveal` click step per top-level body block after the first, in `block_ordinal` order, with `component='v-click'`, one target, and ordinals starting at `1`. `speaker_note` blocks are excluded: they produce no body element and cannot consume a page state.
- `reveal_list_items` emits one `reveal` click step for each direct child item of each top-level `bullet_list`, in block order then child order, with `component='v-clicks'`. Each target is the child block ID. Nested lists reveal with their containing item; other top-level blocks are not click targets.

`hide` actions and `v-after` components are schema-valid for future profiles only. No v1 `click_profile` generates them.

**REQ-RPT-087d**
`derive_deck_v2` is selected only when the release tuple carries a non-null composition. When the composition tuple is all-null, Reporting MUST select `derive_deck_v1`, MUST NOT execute composition validation beyond the all-null tuple check, and MUST produce byte-identical canonical deck models and downstream source bytes for every existing no-composition fixture.

For a non-null composition, `derive_deck_v2` MUST execute these steps in order:

1. Execute `derive_deck_v1` steps 1 through 3 to produce the base, unchunked slide list from the post-redaction export model and template manifest.
2. Resolve composition anchors under REQ-RPT-087e against the post-redaction export model, selected template manifest, and admitted composition-owned diagram declarations.
3. Apply composition operations in array order using Table 17-D.
4. Execute `derive_deck_v1` steps 4 through 6: chunking, click-step generation, slide ordinal assignment, structural-ID assignment, and `slide_count`, `click_step_count`, and `expected_export_page_count` computation.

Two conforming implementations given the same post-redaction export model, template manifest, and digest-bound composition MUST produce byte-identical canonical deck models.

**REQ-RPT-087e**
Composition operations MUST target semantic anchors, never Reporting structural IDs derived from ordinal paths. A composition anchor vocabulary imported from Table 6-B1 resolves as follows:

| Anchor kind | Resolution input | Required unique target |
| --- | --- | --- |
| `section_anchor` | Template `decl_id` plus the expansion-dimension key supplied by `template_sections_expansion_v1`. | One emitted section instance before deck chunking. |
| `record_anchor` | Source `record_id`. | One `cartulary.reporting_record_summary.v1` in the post-redaction export model. |
| `block_anchor` | `section_anchor`, block kind, and optional `record_anchor`. | One top-level or nested block before deck chunking; bare `block_ordinal` is invalid. |
| `diagram_anchor` | Diagram declaration `decl_id`. | One template-owned or composition-owned diagram declaration. |

For a record-qualified block anchor, the record must first resolve uniquely in
the retained Core summary collection. A candidate derived block matches that
record only through a current field source reference in its own or descendant
derived blocks. The source index must bind that reference to the exact Core ID
through a Core source family or its generic `record_envelope` projection.
Auxiliary source IDs, display values, selector spelling and cached container
reference aggregates MUST NOT establish record membership. Removed children and
composition-authored insertions MUST NOT manufacture a derived target. Matching
multiple blocks remains ambiguous even when they reference the same record.

An anchor that resolves to zero targets MUST fail with `failure_code='composition_invalid'` and `reason_code='composition_anchor_unresolved'`. An anchor that resolves to more than one target MUST fail with `failure_code='composition_invalid'` and `reason_code='composition_anchor_ambiguous'`. Every operation's `on_unresolved` policy defaults to `fail`; `on_unresolved='drop'` is invalid for `external_release` and MUST fail with `reason_code='composition_drop_invalid_for_external_release'`. Structural identifiers such as `sec-0001`, `sec-0001.b0001`, `sld-0001`, chunk IDs, and split IDs MUST NOT be accepted as composition targets.

**REQ-RPT-087f**
Composition-authored text is presentation-tier text only. The only valid authored text roles are `title_override`, `speaker_notes`, and `authored_text`. Every authored text node MUST carry exactly one author-declared `disclosure_partition_ref` matching Table 12-B except `blocked`, and Reporting MUST admit it into the post-redaction model as a singleton `disclosure_partition_refs[]` so that `filter_disclosure_partitions_v1`, Core redaction rule selection, and token substitution apply exactly as for derived presentation text.

The v3 export model adds the required `presentation_texts` collection. It is
empty in an admitted render model; immutable snapshot materialization has no
such collection. A participant cannot prepopulate it. Reporting alone derives
its retained nodes from the independently digest-bound composition. This is a
clean cutover from export-model v2; no reader upgrades retained models or
reinterprets earlier completed output as evidence of this contract.

Each `cartulary.reporting_presentation_text.v1` object has exactly `schema_id`
(the schema token), `authored_text_id` (the companion identity), `text_role`
(the companion role), `content_class` (exact `presentation_text`), and `body`
(a Table 9-D field with `field_key='body'`, ordinal 1, source state `derived`,
null display label and raw-value digest). Its disclosure array is the singleton
author-declared partition. Its source/support arrays are the sorted unions of
explicit resolved subject placeholders; literal prose receives no inferred
source record or supporting evidence. The same references survive insertion,
title override and speaker-note derivation, including chunk-specific note
placement. The pool is ordered by exact authored-text identity.

The selected-rule address is
`$.presentation_texts['<authored_text_id>'].body`, using the ordinary exact-path
escaping grammar. Select and retain a decision for every declared body, including
unused or dropped nodes. Resolve placeholders only through eligible snapshot
subject uses, then choose an assigned display token before the selected
post-redaction `display_name` (or unresolved mention `raw_text`). An absent,
filtered or ambiguous display has no fallback. A whole-body token-backed rule
requires exactly one distinct placeholder subject; zero or multiple subjects
fail the ordinary unique-subject rule. Such uses participate in the same token
ordinal allocation as snapshot field uses.

Apply the selected action to the substituted body without selecting another
rule. Validate role, NFC, controls and operative scalar limits before and after
substitution and after any retained literal replacement. Substitution MUST retain
at most the body's operative scalar limit; reject before appending
a fragment that would exceed that limit. A short placeholder must not trigger
unbounded allocation from a large selected display value. Each body contributes
one selected-rule manifest entry with its exact outcome. A dropped body is
absent from the retained pool; its operation resolves its semantic anchor using
the ordinary policy, then has no presentation effect. Dropping a body cannot
hide an invalid anchor or restore an earlier raw body. Deck derivation consumes
only the retained body and its redaction state. Completion, replay and recovery
recompute both the presentation pool and its decisions from frozen admission.

For `external_release`, any composition-authored text MUST fail with `failure_code='composition_invalid'` and `reason_code='authored_text_not_permitted'` unless the selected redaction profile has `allow_authored_presentation_text=true`. When permitted, the authored text still MUST satisfy partition filtering and redaction; the profile permission does not imply `allow`.

`title_override` MUST be single-line, MUST satisfy `composition.authored_title_chars`, and is admitted only as a section or slide title. `speaker_notes` MAY contain LF, MUST satisfy `speaker_notes_chars_per_slide` after placeholder substitution, and is admitted only as slide `speaker_notes` or `speaker_note` blocks. `authored_text` MAY contain LF, MUST satisfy `composition.authored_text_chars`, and is admitted only as `paragraph` blocks with `content_class='presentation_text'`. Reporting v1 does not semantically classify authored free text as fact, conclusion, or presentation prose; deterministic admission is governed only by role, partition, profile permission, placeholder, limit, redaction, and schema rules. Case facts, actors, timestamps, commands, causal chains, source evidence, party relationships, or conclusions belong in snapshot artifact records and template narrative slots as an authoring policy outside release-byte conformance.

Inline subject placeholders MUST use the exact form `{{subject:<stable_subject_ref>}}`. Each placeholder MUST resolve to exactly one selected `cartulary.tokenizable_subject.v1` and MUST be substituted through the existing post-redaction display-token or display-name priority used for subjects. An unresolved, filtered, ambiguous, malformed, or unreleasable placeholder MUST fail with `failure_code='composition_invalid'` and `reason_code='authored_subject_ref_unresolved'`. Reporting v1 does not scan authored free text for literal sensitive values; builder linting and review guidance for such values are non-normative and cannot satisfy or fail this requirement.

**REQ-RPT-087g**
Composition operations MUST use the Reporting effects in Table 17-D. The companion `docs/report-composition-nlspec.md` owns each operation's complete wire schema, but Reporting owns the observable deck, diagram, validation, and determinism effects listed here.

**Table 17-D. Composition operation effects**

| `op_kind` | Reporting effect | Resolution rule |
| --- | --- | --- |
| `exclude_section` | Remove an emitted section before chunking. If the resolved section was declared `required=true`, the operation fails with `required_section_empty`. | `section_anchor` |
| `reorder_sections` | Move listed sections before unlisted retained sections in listed order; unlisted retained sections keep their derived relative order. Duplicate section anchors are invalid. | `section_anchor[]` |
| `override_slide_layout` | Replace `layout_id` with a template-declared layout token before block-kind validation. | `section_anchor` |
| `override_title` | Replace section or slide title with an admitted `title_override` authored text node. | `section_anchor`, `authored_text_ref` |
| `set_speaker_notes` | Attach or replace speaker notes on the resolved base slide using an admitted `speaker_notes` authored text node. | `section_anchor`, `authored_text_ref` |
| `insert_authored_block` | Insert one admitted `authored_text` paragraph block before or after the resolved block. | `block_anchor`, `position`, `authored_text_ref` |
| `exclude_block` | Remove one derived block before chunking; if the section becomes empty, REQ-RPT-059b and `required_section_empty` apply. | `block_anchor` |
| `override_click_profile` | Replace the section declaration click profile with `none`, `reveal_blocks`, or `reveal_list_items`. | `section_anchor` |
| `insert_diagram_slide` | Emit a base slide for a composition-owned diagram declaration before chunking. | `diagram_anchor`, `section_anchor`, `position` |
| `exclude_diagram` | Remove a template-declared diagram before diagram serialization and deck serialization. | `diagram_anchor` |
| `override_diagram_labels` | Apply valid label overrides under REQ-RPT-079e before diagram serialization. | `diagram_anchor` |

A `diagram_ref` block carries exactly one derived field named `diagram_id`,
whose string value identifies the retained diagram. It has no children. Its
content class is `derived_analytic`; its field and container reference unions
come from that diagram, never from an inferred source path or a default public
partition. A composition-inserted diagram slide initially uses the declaration
ID as its title, template layout token `default`, and click profile `none`.
These values are presentation defaults; generated slide/block identities are
assigned after operations and chunking. Template-owned diagram references
appear in their declaring section, in declaration-array order.

If more than one operation writes the same scalar property, the later operation in array order wins unless a narrower row says the operation is invalid. If an operation removes a target, a later operation targeting that removed object MUST resolve according to `on_unresolved`; Reporting MUST NOT retain hidden tombstones to make later ordinal-path targeting work. Repeated `reorder_sections` operations each apply to the current retained section list. Multiple `insert_authored_block` or `insert_diagram_slide` operations with the same resolved anchor and `position` emit in `deck_ops[]` order relative to that anchor. A second `insert_diagram_slide` for the same composition-owned diagram declaration MUST fail with `failure_code='composition_invalid'` and `reason_code='composition_duplicate_diagram_insert'`.

**REQ-RPT-087h**
Authoritative report-composition previews are ordinary `internal_draft` render attempts through `render_slidev_bundle_v1` or `render_mermaid_bundle_v1` using a companion-owned `cartulary.report_composition_preview_source.v1` for draft previews or the immutable composition tuple for version previews. A preview source descriptor is valid only for `internal_draft`; it MUST NOT be accepted for `external_release`, approval, release publication, or bundle evidence. Draft preview descriptors MUST bind by `preview_source_sha256`, not by release tuple `composition_sha256`; immutable-version previews MAY also report the immutable `composition_sha256` supplied by the companion owner. Previews use the same tuple validation, composition validation, partition assignment, redaction, sandbox, limits, source serialization, and bundle validation as any other internal draft. A builder UI MAY provide a client-side live approximation, but that approximation is non-normative, is not reviewable or approvable bytes, and MUST NOT be used as evidence that external-release bytes will pass.

Preview admission MUST atomically retain the exact normalized public request,
Composition-owned immutable source bytes and attempt identity, fixed microsecond
admission time, authored policy, Graph results and template pack as the §5 recipe.
The expected materialized-profile digest remains an exact request expectation
checked during worker preparation. Preview v2/v3 Jobs and full-model inputs are
retired. Completion and recovery validate retained bindings and independent Jobs
outcomes; unsuccessful or unfinished attempts have no output. Public-response
expiry preserves this private admission. Validation MUST NOT re-render, recreate
missing inputs, reacquire Graph leases, select a later draft, or replace the exact
Reference Pack binding. Composition owns its source read and enumeration inside
the caller's consistent transaction. Shared Jobs and Auth idempotency retain the
same normalized public-request digest.

**REQ-RPT-088**
A slide object MUST use Table 17-B.

**Table 17-B. Slide object schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `slide_id` | `identifier` | Yes | No | None | Structural ID from REQ-RPT-040b. |
| `slide_ordinal` | `finite_integer` | Yes | No | None | Starts at `1`, no gaps. |
| `title` | `safe_string` | Yes | No | None | Redaction-safe slide title. |
| `layout_id` | `identifier` | Yes | No | None | Closed layout token declared by template manifest. |
| `blocks[]` | array | Yes | No | `[]` | Allowed slide block kinds in `block_ordinal` order. |
| `click_steps[]` | array of `click_step.v1` | Yes | No | `[]` | Empty when no click steps are generated. |
| `speaker_notes` | `safe_string` | Yes | Yes | None | Null when omitted; LF allowed by REQ-RPT-036a. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Required non-empty array. |

**REQ-RPT-089**
Slide model bounds MUST use Table 17-C. The numeric default and hard limits are defined once in §25 Table 25-A; Table 17-C maps each limit to its Slidev-source failure and reason code.

**Table 17-C. Slide model bounds**

| Limit | Default and hard limit | Failure |
| --- | --- | --- |
| `slides.count` | Defined once in §25 Table 25-A | `failure_code='slidev_source_invalid'`, `reason_code='slide_count_exceeded'` |
| `blocks_per_slide` | Defined once in §25 Table 25-A | `failure_code='slidev_source_invalid'`, `reason_code='slide_block_limit_exceeded'` |
| `speaker_notes_chars_per_slide` | Defined once in §25 Table 25-A | `failure_code='slidev_source_invalid'`, `reason_code='speaker_notes_limit_exceeded'` |

# 18. Slidev source generation

## 18.1 Slidev subset

**REQ-RPT-090**
Reporting v1 MUST generate only the Slidev subset defined by this section. Arbitrary Vue components, parser extensions, pre-parser extensions, Monaco, Twoslash, remote themes, remote assets, arbitrary scripts, iframes, external embeds, and external fetch behavior are invalid for external release.

## 18.2 Canonical Slidev Markdown serializer

**REQ-RPT-091**
`slidev_markdown_serialize_v2` MUST use Table 18-A.

**Table 18-A. Slidev Markdown serialization**

| Area | Required rule |
| --- | --- |
| Encoding | UTF-8. |
| Line endings | LF only. |
| Final newline | Required. |
| Headmatter | First block only, delimited by `---`. |
| Headmatter keys | Emit exactly the keys in Table 18-C, in that order, always present with the explicit values in that table (REQ-RPT-091a). |
| YAML strings | Double-quoted scalars only. |
| YAML booleans | `true` or `false`. |
| YAML forbidden forms | Anchors, aliases, tags, folded blocks, and multiline scalars. |
| Slide boundary | The headmatter is the first slide's frontmatter and includes its resolved `layout` as the last key. First-slide content follows immediately after its closing delimiter and one blank line. Every later slide starts with a blank line, `---`, LF, exactly one `layout` key as a double-quoted scalar, `---`, and one blank line. There is no synthetic cover or empty slide between headmatter and the first model slide. |
| Slide title | First content line of every slide is `# ` followed by the escaped slide title. |
| Block serialization | Blocks serialize in `block_ordinal` order under `serialize_block_markdown_v1` (Table 18-D), separated by exactly one blank line. |
| Markdown escaping | Escape generated text runs per `slidev_markdown_escape_v2` (Table 18-B, REQ-RPT-091a). |
| Raw HTML | Forbidden except allowed reveal-only components and speaker-note comments. |
| Diagram refs | Auto-layout refs emit exact opening ```` ```mermaid```` and canonical `.mmd` source bytes inside; manual-layout refs emit one local SVG image reference. |
| Speaker notes | Final slide block serialized as `<!--\n{escaped note lines}\n-->`. |
| Tables | Pipe table, deterministic column order, escaped cells, no alignment syntax. |
| Code fences | Only `text` and `mermaid`. A later revision that adds a code-fence language MUST define the language token, allowed source positions, escaping rules, parser safety checks, render behavior, and behavioral validation in the same revision. |
| Click components | `v-click` and `v-clicks` wrap the target block Markdown as an HTML block: opening line `<v-click at="{at}">` or `<v-clicks at="{at}">`, blank line, block content, blank line, and closing line `</v-click>` or `</v-clicks>`. |

**REQ-RPT-091a**
`slidev_markdown_escape_v2` MUST escape generated text runs so that `slides.md` is byte-deterministic and no generated content can alter Markdown or Slidev structure. It applies only to generated text placed in a Markdown text position (paragraph text, heading text, list-item text, table-cell text); it MUST NOT alter serializer-emitted structure such as headmatter, slide separators, fence markers, or the speaker-note comment wrapper. The escape set is closed in Table 18-B and is defined against the pinned Slidev markdown-it dialect (CommonMark plus GFM) recorded in the toolchain snapshot. After escaping, a generated run MUST parse to exactly one text run yielding the intended code points.

**Table 18-B. `slidev_markdown_escape_v2` escape set**

| Position | Characters backslash-escaped |
| --- | --- |
| Any position in a generated text run | REVERSE SOLIDUS (U+005C), GRAVE ACCENT (U+0060), VERTICAL LINE (U+007C), and each of `` `* _ [ ] < ~ $ &` `` |
| Start of a generated line (first non-space character) | additionally `#`, `>`, `-`, `+`, `=`; a leading run of GRAVE ACCENT or `~`; and a leading `[0-9]+` immediately followed by `.` or `)` |

Escaping inserts a single REVERSE SOLIDUS (U+005C) before the character. Characters outside this set are emitted literally.
For an ordered-list prefix, escape the separating `.` or `)`, not a digit.
AMPERSAND is escaped so source text resembling an HTML character reference
retains its original code points. Version 2 retires v1 source-escaping evidence;
there is no dual serializer or fallback reader.

After Markdown tokenization and before Vue compilation, every text token MUST
be emitted as inert HTML text. Escape ampersand and HTML delimiters, and encode
both braces as numeric character references at this final text-rendering boundary.
A Markdown backslash or source-level character reference alone does not satisfy
this boundary because Markdown can decode it before Vue compilation. Text such
as `{{ expression }}` MUST render literally and MUST NOT evaluate. Automatic
linkification and typographic substitution are disabled; plain URLs remain text.
Serializer-owned components such as `v-click` retain their declared behavior.
Both standalone SVG and Slidev-embedded Mermaid MUST use the same installed
render profile: strict security, automatic startup disabled, HTML labels
disabled, deterministic IDs with seed `cartulary`, sans-serif fonts, default
theme and classic look. Flowcharts use linear curves and disable maximum-width
scaling; sequence actor, message and note fonts are sans-serif with maximum-width
scaling disabled. Each isolated render uses `cartulary-diagram` as its renderer
root identity; embedded diagrams are isolated in separate shadow roots. Slidev
or theme defaults MUST NOT override this profile. Setup code is renderer-owned
and MUST NOT interpolate case content or accept author-supplied configuration.
Together with the declared-asset resolution and failure rules below, this changes
the renderer contract to version 15; earlier rendered-byte evidence does not
qualify the new profile.

Canonical local asset references remain relative bundle paths in `slides.md`.
The renderer MUST resolve only declared paths through the same attempt's
allocated endpoint, without treating them as package imports, filesystem-relative
module imports or inline `data:` URLs. Missing or failed local resources and
failed slide imports MUST reject the export before PDF capture. A renderer error
layout is never a successful slide. Only closed diagnostic classes may leave the
native stage; resource URLs, vendor messages and input content remain excluded.

PDF capture MUST wait for every mounted Mermaid component in the pinned print
surface to contain its completed SVG with a non-empty view box and graphical
content. Slide/component loading completion alone is insufficient: Mermaid fills
its shadow root asynchronously. Font readiness, local-image decoding and the
two-frame paint barrier follow diagram completion. A missing or failed diagram
MUST fail within the stage deadline; it MUST NOT produce a partially painted PDF
or trigger a render-action retry.
Tests MUST exercise these properties through the pinned renderer, including
entity-like text, expression-like braces, escaped HTML, URLs and click expansion.


**Table 18-C. Headmatter keys**

| Key | Always emitted | Value |
| --- | --- | --- |
| `title` | Yes | Deck title from the deck model. |
| `author` | Yes | Template-declared author, or `""` when none. |
| `theme` | Yes | Template-declared theme. |
| `download` | Yes | `false`. |
| `export.format` | Yes | Literal `"pdf"` in canonical `slides.md` for every `slidev` bundle. Supported PDF and internal source-only variants preserve these source bytes; unsupported PPTX/PNG requests reject before rendering. |
| `export.withClicks` | Yes | `true`. |
| `clickAnimation` | Yes | `false`. |
| `lineNumbers` | Yes | `false`. |
| `monaco` | Yes | `false`. |
| `twoslash` | Yes | `false`. |
| `layout` | Yes | First model slide's resolved layout token. Later slides declare their own layout in their slide frontmatter. |

The first-slide rule follows the pinned Slidev parser's document model. A separate
frontmatter block before first-slide content would create an extra slide and is
invalid. Version 2 source evidence supersedes v1; the serializer emits exactly
the declared slide list. Template manifest v3 retains explicit `author` and `theme`
ownership. The current built-in template is selected under REQ-RPT-102a;
retired versions have no alias. Its admitted metadata and manifest digest must match
the current template exactly. Existing history is subject to the clean
pre-production cutover; no reader fills missing metadata or rewrites old source.
Before applying migration 72, preflight rejects retained Reporting snapshots,
Jobs, imported artifacts or previews without mutation. The migration repeats
that inventory under locks and admits only explicit, request-bound template
metadata. The active template registry owns supported versions and themes; the
database does not become a second template registry.

**Table 18-D. `serialize_block_markdown_v1` block-kind mapping**

| `block_kind` | Canonical Markdown form |
| --- | --- |
| `paragraph` | Each `fields[].value` converted by `field_value_to_text_v1` in `field_ordinal` order as one paragraph. Values are joined by LF within the paragraph and escaped per Table 18-B. |
| `bullet_list` | One `- ` item per child block in `block_ordinal` order; child `paragraph` content escaped per Table 18-B; nesting indents by two spaces per depth. Maximum nesting depth is `3`; deeper structures fail with `failure_code='slidev_source_invalid'` and `reason_code='list_nesting_exceeded'`. |
| `table` | GFM pipe table. Header row uses `display_label`, or `field_key` when label is null, from the first child block's fields in `field_ordinal` order. Body rows follow child `block_ordinal`; cells follow field `field_ordinal`; cell values convert through `field_value_to_text_v1` and escape per Table 18-B. Delimiter row uses `---` per column with no alignment syntax. |
| `metric` | One line per field in `field_ordinal` order: `**{escaped label}:** {escaped value}`. Null label uses `field_key`; value converts through `field_value_to_text_v1`. |
| `timeline_rows` | Pipe table. Column 1 header is `Time`; value is `display_times.primary_display` or the literal `unresolved`. Remaining columns use the event fields as in `table`, including `field_value_to_text_v1`. Rows follow §14.2 order. |
| `diagram_ref` | Auto-layout diagrams serialize as a Mermaid fence per Table 18-A containing the referenced diagram's canonical `.mmd` bytes verbatim. Manual-layout diagrams serialize as `![{escaped diagram_id}](diagrams/{diagram_id}.svg)`. |
| `asset_ref` | Resolve `asset_id` to exactly one template `assets[]` item and serialize `![{escaped display_label or asset_id}]({bundle_path})` using that asset item's declared `bundle_path`. The serializer MUST NOT derive a filename, directory, extension, or media type from `asset_id`, source path, or renderer output. |
| `speaker_note` | Not serialized in the body. Speaker-note blocks are collected in `block_ordinal` order, joined by LF, and emitted as the slide-final speaker-note comment. |
| `overflow_summary` | One paragraph with exact form `Omitted {omitted_row_count} rows under {selection_rule_id}. {escaped filter_summary}`. The trailing space and `{escaped filter_summary}` segment are omitted when `filter_summary` is null. |

The deck carries the retained Table 9-C blocks and Table 9-D fields, without an
implementation-specific flattened text representation. Chunking preserves
block identities and field values while recomputing block ordinals inside each
chunk and reference unions for that chunk. Source arrays and values remain
immutable. A paragraph outside a list has no child blocks. Each list child is a
paragraph item whose optional children are nested `bullet_list` blocks; those
nested lists follow their item's text. Continuation lines in an item are indented
by two additional spaces. A table has no own fields, at least one paragraph row,
and at least one column. Every row has the same ordered field keys as the first
row and no child blocks. Invalid shapes fail with `slidev_source_invalid` /
`invalid_slidev_construct`; a missing cell is never repaired by shifting later
columns. Metric blocks have no child blocks. Slide-level speaker notes, when
present, precede speaker-note block values in the single final note comment.
The combined note text is subject to the per-slide limit and delimiter check.

**REQ-RPT-091b**
A speaker-note line set containing the byte sequence `-->` before or after escaping MUST fail before persistence with `failure_code='slidev_source_invalid'` and `reason_code='invalid_speaker_note_sequence'`. The serializer MUST NOT mutate, split, or further escape that sequence inside the HTML comment wrapper.

**REQ-RPT-091c**
`field_value_to_text_v1` MUST convert field values to generated text before Markdown escaping using Table 18-E. Objects are invalid field values in Reporting v1. Arrays MUST NOT contain arrays or objects. A conversion failure MUST fail with `failure_code='slidev_source_invalid'` and `reason_code='invalid_field_value_for_text'`.

**Table 18-E. `field_value_to_text_v1`**

| Input value | Text output |
| --- | --- |
| JSON string | The string unchanged after `safe_string` validation. |
| JSON integer | The decimal integer token with no leading plus sign and no leading zeroes except `0`. |
| JSON boolean | `true` or `false`. |
| JSON null | Empty string. |
| Array of scalars | Scalar text outputs joined by comma followed by space, preserving array order. Empty array emits empty string. |

**REQ-RPT-092**
Generated `slides.md` MUST be byte-identical for the same deck model, template manifest, and render environment profile. Admitted materialized output-option values MUST NOT alter `slides.md` bytes; supported PDF and internal source-only choices alter only render execution and bundle membership. Unsupported PPTX/PNG requests fail under §7.5 rather than selecting another source form. Valid Markdown variants that differ in bytes are non-conformant.

# 19. Reveal-only click-step profile

**REQ-RPT-093**
`cartulary.slidev_reveal_only.v2` MUST use only deterministic reveal behavior described by `click_step.v1`. It MUST NOT use timing animations, CSS transitions, randomization, arbitrary Vue state, route navigation, remote scripts, or viewer-side runtime dependencies for external release. Version 2 corrects the former whole-list step count to match the pinned `v-clicks` item's reveal behavior. Version 1 click evidence is retired without reinterpretation; no-click source bytes retain their existing meaning.

Click steps are generated after chunking and final slide identity assignment.
Each step ID is `clk-` followed by its four-digit ordinal within the slide.
The serializer MUST independently validate complete target membership, order,
component, visibility, ordinal, and state hashes before emitting any component.
For a list, one `v-clicks` wrapper encloses the complete list, with `at` equal to
its first item's step ordinal; subsequent items consume consecutive ordinals.
For block reveals, each target has its own `v-click` wrapper. Unknown, missing,
duplicate, stale, cross-slide, hide, and `v-after` steps are not admitted by v2.
The exported page-state count MUST match the resulting item-level count.

**REQ-RPT-094**
`click_step.v1` MUST use Table 19-A.

**Table 19-A. `click_step.v1` schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `click_step_id` | `identifier` | Yes | No | None | Stable in slide. |
| `ordinal` | `positive_integer` | Yes | No | None | Starts at `1`, increments by `1` with no gaps, and MUST be in `[1, click_steps_per_slide]`. |
| `action` | string | Yes | No | None | `reveal` or `hide`. |
| `targets[]` | array | Yes | No | None | Non-empty array of block or item IDs already present on same slide. |
| `initial_visibility` | string | No | No | Action-derived | Default `hidden` for reveal targets, `visible` for hide targets. |
| `component` | string | Yes | No | None | `v-click`, `v-clicks`, or `v-after`. |
| `at` | `positive_integer` | Yes | No | None | Equal to `ordinal` and in `[1, click_steps_per_slide]`. |
| `resulting_state_hash` | `sha256_hex` | Yes | No | None | Hash input defined by REQ-RPT-094a. |

`click_steps_per_slide` in Table 25-A is the operative upper bound for both `ordinal` and `at`. JSON numbers using zero, negative values, decimal notation, exponent notation, `-0`, or values above the operative bound MUST fail with `failure_code='slidev_source_invalid'` and `reason_code='click_step_limit_exceeded'`.

**REQ-RPT-094a**
`resulting_state_hash` MUST hash the §10 canonical serialization of `{"schema_id":"cartulary.click_state.v1","slide_id":<slide_id>,"states":[{"target_id":...,"visible":...} ...]}` under REQ-RPT-049. `states[]` MUST cover every click-targeted element on the slide and sort bytewise ascending by `target_id`.

**REQ-RPT-095**
Click-step components MUST follow Table 19-B.

**Table 19-B. Click-step component constraints**

| Component | Valid use |
| --- | --- |
| `v-click` | One target. |
| `v-clicks` | Homogeneous ordered list or table children only. |
| `v-after` | Same ordinal as immediately prior reveal group only. |

**REQ-RPT-096**
A slide with `N` click steps MUST export `N+1` page states: base state plus one state after each ordinal. A rendered PDF whose page-state count differs from this rule MUST fail with `failure_code='slidev_export_failed'` and `reason_code='click_export_page_count_mismatch'`.

# 20. Toolchain snapshot and render environment

**REQ-RPT-097**
`cartulary.reporting_toolchain_snapshot.v2` MUST bind every toolchain and environment value that affects generated or rendered bytes. Floating versions, `latest`, network package resolution during render, and undeclared renderer binaries are invalid.

**REQ-RPT-098**
The toolchain snapshot MUST use Table 20-A.

**Table 20-A. Toolchain snapshot fields**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_toolchain_snapshot.v2`. |
| `created_at` | `timestamp` | Yes | No | None | Equals normalized `render_admitted_at`. |
| `node_version` | string | Yes | No | None | Exact version string. |
| `package_manager` | string | Yes | No | None | Exact name and version. |
| `lockfile_sha256` | `sha256_hex` | Yes | No | None | Digest of package lockfile bytes. |
| `package_store_digest` | `sha256_hex` | Yes | No | None | `content_manifest_digest_v1` (REQ-RPT-050a) over every file in the pinned package store. |
| `slidev_version` | string | Yes | Yes | None | Required when `output_kind='slidev'`; otherwise null. |
| `mermaid_version` | string | Yes | Yes | None | Required when Mermaid generation or rendering occurs; otherwise null. |
| `chromium_version` | string | Yes | Yes | None | Required when browser rendering occurs; otherwise null. |
| `chromium_executable_path` | `bundle_path` | Yes | Yes | None | Required render-sandbox path when browser rendering occurs; otherwise null. |
| `chromium_executable_sha256` | `sha256_hex` | Yes | Yes | None | Required when browser rendering occurs; otherwise null. |
| `os_image_id` | string | Yes | No | None | Exact repo-control or container image ID. |
| `os_image_sha256` | `sha256_hex` | Yes | Yes | None | Required when image is content-addressed; otherwise null. |
| `timezone` | string | Yes | No | `UTC` | Exact `UTC`. |
| `locale` | string | Yes | No | `C.UTF-8` | Exact `C.UTF-8`. Template-declared locale variation is future-only in v1. |
| `font_manifest_sha256` | `sha256_hex` | Yes | No | None | `content_manifest_digest_v1` (REQ-RPT-050a) over every usable font file and font configuration file. |
| `font_files[]` | array of closed `{path, sha256}` objects | Yes | No | None | Nonempty, at most 1,024 entries, sorted by bytewise path; no duplicates. Installed font/configuration paths are `installed/{relative_path}`. Every retained theme stylesheet and WOFF2 file also appears as `template/{bundle_path}` with its exact byte digest. No other namespace is admitted. This exact array is the preimage for `font_manifest_sha256`. |
| `viewport_css_px` | string | Yes | No | `1280x720` | Exact `1280x720` in v1. Other viewport sizes are future-only. |
| `device_scale_factor` | `finite_integer` | Yes | No | `1` | Exact integer `1` in v1. |
| `color_scheme` | string | Yes | No | `light` | `light` or `dark`. |
| `browser_launch_args[]` | array | Yes | No | `[]` | Exact ordered array emitted by the Core 04 render sandbox profile; templates and release requests MUST NOT set or extend it. |
| `env_allowlist[]` | array | Yes | No | `[]` | Exact names and value digests emitted by the Core 04 render sandbox profile; templates and release requests MUST NOT set or extend it. |
| `render_command[]` | array | Yes | No | None | Exact argv array; shell string invalid. |
| `network_policy_id` | `identifier` | Yes | No | None | Exact render sandbox policy ID. |

For font inventory v2, the installed relative path is relative to the admitted
installation root. Include every file under `runtime/fonts/`, plus every `.css`,
`.woff`, `.woff2`, `.ttf`, `.otf`, `.ttc`, `.otc` and `.eot` file under the pinned
`runtime/renderer/` package tree, using case-insensitive extension comparison.
This conservative set includes package webfonts and potential font stylesheets,
even if the current deck does not reference them. The installation's narrower
font-directory checksum is not a substitute for this complete toolchain inventory.

**REQ-RPT-098a**
Render environment variability MUST use Table 20-B. A template, release request, renderer wrapper, or deployment-local setting MUST NOT vary a byte-affecting field unless Table 20-B assigns that owner authority for v1. A variation outside the table fails before rendered bytes with `failure_code='nondeterministic_render'` and `reason_code='toolchain_snapshot_mismatch'` when detected during determinism validation, or with `error.code='invalid_release_request'` and `reason_code='template_manifest_invalid'` when detected during template admission.

**REQ-RPT-098b**
Native and portable bundle admission MUST require exactly one required
`toolchain_snapshot` member at `validation/toolchain.json`, with its declared
media type, canonical closed Table 20-A bytes and schema-domain digest matching
the root's `toolchain_snapshot_sha256`. The recorded `created_at` equals the
render admission instant. The validation-artifact roster contains exactly
`validation/summary.json` and `validation/toolchain.json` in that order. Renderer
and browser presence must agree with the bundle's source and rendered roles;
unused nullable members remain null. Digest shapes, required non-null arrays,
the fixed environment values and ordered unique environment-name bindings are
validated independently of the outer file hash. These checks read retained
bytes only. They do not re-attest execution on the current host or establish
rendered-byte determinism for a different installation.

**Table 20-B. Render environment byte-affecting defaults**

| Field | V1 default | V1 bounds | Owner allowed to vary in v1 | Omission behavior |
| --- | --- | --- | --- | --- |
| `timezone` | `UTC` | Exact `UTC` | None | Materialize `UTC`. |
| `locale` | `C.UTF-8` | Exact `C.UTF-8` | None | Materialize `C.UTF-8`. |
| `viewport_css_px` | `1280x720` | Exact `1280x720` | None | Materialize `1280x720`. |
| `device_scale_factor` | `1` | Exact integer `1` | None | Materialize `1`. |
| `color_scheme` | `light` | `light` or `dark` | Template render profile only | Materialize `light`. |
| `browser_launch_args[]` | `[]` | Exact Core 04 sandbox profile array | Core 04 render sandbox profile only | Materialize `[]`. |
| `env_allowlist[]` | `[]` | Exact Core 04 sandbox profile array of names and value digests | Core 04 render sandbox profile only | Materialize `[]`. |

**REQ-RPT-099**
Rendered-byte determinism is scoped to the exact toolchain snapshot and render environment. An implementation MUST NOT claim byte-identical rendered outputs across undeclared host machines, font sets, locales, timezones, Chromium executables, or viewport settings.

**REQ-RPT-099a**
The claimed profile requires `snapshot_reporting.runtime_receipt_path`, a canonical
absolute path to the installed `installation.json` receipt. The value has no
default, is forbidden when the profile is inactive, resolves as `regular_file_ref`,
and uses name-only diagnostics. `snapshot_reporting.configuration.v2` and
`cartulary.snapshot_reporting.configuration_namespace.v2` own this configuration;
`cartulary.reporting_runtime_receipt_path.v1` admits only a nonempty absolute path
of at most 4096 bytes, with no NUL, traversal, or noncanonical segments.
Application startup MUST admit the complete receipt-bound installation before
profile state mutation or worker registration. A missing or incompatible runtime
prevents admission; it MUST NOT select placeholder output or another host binary.
Startup rejection uses `invalid_deployment_config` with the key name and
`reason_code='reporting_runtime_unavailable'`, without receipt or host path values.
The receipt's parent directory is the installation boundary. The deployment's
temporary-work root supplies private attempt directories. These host paths MUST
NOT enter release content or safe diagnostic values.

**REQ-RPT-099g**
Admission of a release or preview that names completed Graph Projection results
MUST retain the exact source-owner contributions inside the same admission
transaction: the complete projection binding, ordered vertices and edges,
canonical result object, and typed label candidates. Reporting MUST retain a
separate admission digest for these private inputs. The private codec is
`cartulary.reporting_graph_input.v1`; an empty selection retains an explicit
empty result array. Publication timestamps, lease timestamps and lease IDs are
operational state and MUST NOT enter the render-input digest. These private
source contributions MUST NOT become shareable bundle members.

The worker MUST compare renewed leased contributions against the admitted
inputs before rendering; it MUST NOT replace them with a newly selected result.
Completion, replay and recovery MUST recompute canonical diagram models and
Mermaid source from the admitted inputs, independent of current source state,
lease availability and renderer installation. Exact source/model membership and
requested rendered-artifact membership MUST match. Internally consistent
replacement hashes do not establish fidelity to admission. Recovery of pending
jobs retains the existing Graph owner lease reconciliation requirement; recovery
of completed output does not require an unexpired lease.

Admissions without this retained input and independent digest require the clean
pre-production cutover. A reader or migration MUST NOT reconstruct missing
historical graph data or source-owner labels.

# 21. Template packs and local assets

**REQ-RPT-100**
`cartulary.reporting_template_pack_manifest.v3` MUST be a closed schema. Unknown top-level members are invalid.

**REQ-RPT-101**
Template manifest top-level fields MUST use Table 21-A.

**Table 21-A. Template pack manifest schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_template_pack_manifest.v3`. |
| `template_id` | `identifier` | Yes | No | None | Stable template identity. |
| `template_version` | `identifier` | Yes | No | None | Exact version; `latest` invalid. |
| `manifest_version` | `identifier` | Yes | No | None | Manifest schema version. |
| `deck_title` | `safe_string` | No | No | `Cartulary Report` | Literal deck title consumed by REQ-RPT-087c. No placeholders are valid in v1. |
| `author` | `safe_string` | No | No | Empty string | Template-owned, single-line literal of at most 120 Unicode scalar values; no C0/C1 controls or placeholders. Materialize the default before admission. |
| `theme` | `identifier` | Yes | No | None | Exact locally installed theme declared by the template and bound by the toolchain snapshot. The built-in template selects `default`; it cannot trigger package resolution or remote loading. |
| `supported_output_kinds[]` | array | Yes | No | None | Non-empty subset of `mermaid`, `slidev`. |
| `supported_release_scopes[]` | array | Yes | No | None | Non-empty subset of §7.3. |
| `allowed_export_model_bindings[]` | array | Yes | No | `[]` | Array of `binding_path_pattern_v1`, ordered bytewise ascending. |
| `sections[]` | array | Yes | No | None | Deterministic section expansion declarations from Table 21-C. |
| `layouts[]` | array | Yes | No | None | Closed layout tokens and allowed block kinds from Table 21-D. |
| `narrative_slots[]` | array | Yes | No | `[]` | Slot declarations from Table 21-E. |
| `assets[]` | array of `template_asset_item.v1` | Yes | No | `[]` | Declared local assets from Table 21-G, ordered by `bundle_path`, then `asset_id`. |
| `render_profiles[]` | array | Yes | No | None | Allowed Slidev, Mermaid, and render-environment profiles from Table 21-F. |
| `supported_output_options[]` | array of `template_output_option_support.v1` | Yes | No | None | One item per supported output kind from Table 21-H, ordered by `output_kind`. |
| `declared_limits` | object | No | No | `{}` | Keys MUST be Table 24-B `declared_limits` timeout keys or Table 25-A limit keys; unknown keys are invalid. Omission means every §24 and §25 default applies. |
| `extension_features[]` | array | Yes | No | `[]` | Extension feature declarations from Table 21-B. |

**REQ-RPT-102**
`binding_path_pattern_v1` is either an exact `export_model_path_v1` or an `export_model_path_v1` where one quoted stable-ID segment is replaced by `[*]`. `**`, regex, diagnostic array indexes, visible labels, storage paths, arbitrary JSONPath syntax, and grid vendor coordinates are invalid.

**REQ-RPT-102a**
The built-in `cartulary.report.default` template uses version `4` and manifest
version `3`. Its manifest is the Table 21-A object itself; private parallel
binding lists and undeclared assets are invalid. It declares one required
`snapshot` appendix section with the literal title `Snapshot export`, singleton
expansion, binding `$.sections[*]`, inclusion in the deck, no clicks, and the
default first-slide layout. The allowed binding set contains that same pattern.
The deck title is `Incident Report`, the author is empty, and the local theme is
`default`. Its `cover`, `default`, and `section` layout tokens map to the same
Slidev names and admit the Table 9-C block vocabulary. It declares no local
assets, narrative slots, diagrams, or extension
features. Its local render profile uses Table 21-F default geometry and light
color scheme. It supports the current release scopes and output kinds; PDF is
supported for Slidev, SVG for Mermaid, and rendered diagrams for both. PNG and
PPTX are unsupported. Version `1`–`3` requests and retained render admissions
are retired at the clean pre-production cutover; they MUST NOT be resolved to
version `4` or rewritten using reconstructed manifest provenance. This built-in
selection does not remove the general template obligations elsewhere in §21.

**REQ-RPT-103**
Template extension features MUST use Table 21-B. Unsupported required features fail before render output bytes.

**Table 21-B. Template extension features**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `feature_id` | `identifier` | Yes | No | None | Namespaced identifier. |
| `required` | boolean | Yes | No | `false` | `true` means unsupported feature fails before render output bytes. |
| `payload_sha256` | `sha256_hex` | Yes | Yes | None | Digest of local payload bytes; null when the feature has no payload. |
| `payload_path` | `bundle_path` | Yes | Yes | None | Local bundle path; null when the feature has no payload. |

If `required=true` and the renderer does not understand the feature, render admission MUST fail with `error.code='invalid_release_request'` and `reason_code='unsupported_template_feature'`. If `required=false`, unsupported feature handling follows REQ-RPT-103a.

**REQ-RPT-103a**
An unsupported extension feature with `required=false` MUST be ignored by omitting its payload from all source bytes, rendered bytes, manifests, validation decisions, and bundle files. If ignoring the feature would change any required binding, local asset, layout, output-option support value, render profile, security control, or validation result declared elsewhere in the manifest, the feature is not safely ignorable and the request MUST fail with `error.code='invalid_release_request'` and `reason_code='unsupported_template_feature'`. Optional feature omission never creates an extension container in the export model or bundle manifest.

**Table 21-C. Section declaration schema (`sections[]` items)**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `decl_id` | `member_name` | Yes | No | None | Unique in manifest. |
| `section_kind` | string | Yes | No | None | Table 9-B vocabulary. |
| `title` | string | Yes | No | None | Literal, except the closed placeholder `{expansion_label}` for expanded instances. |
| `expansion` | string | No | No | `none` | `none`, `per_party`, or `per_subject`. `per_party` expands one instance per recipient-partition Party sorted by `party_id`; `per_subject` expands one instance per subject sorted by `stable_subject_ref`. |
| `bindings[]` | array of `binding_path_pattern_v1` | Yes | No | None | Every binding MUST be a subset of `allowed_export_model_bindings[]`. |
| `slide_layout_id` | string | No | No | `default` | Must name a Table 21-D layout. The first emitted deck slide defaults to `cover` when this member is omitted. |
| `click_profile` | string | No | No | `none` | Vocabulary from REQ-RPT-087b. |
| `include_in_deck` | boolean | No | No | `true` | Governs `derive_deck_v1` inclusion. |
| `required` | boolean | No | No | `false` | If filtering empties the section, REQ-RPT-059b applies. |
| `include_superseded` | boolean | No | No | `false` | External superseded-record opt-in under REQ-RPT-043a. |
| `truncate_labels` | boolean | No | No | `false` | Enables Table 16-A1 truncation. |
| `selection_rule_id` | `identifier` | Yes | Yes | None | Required non-null for timeline sections that declare overflow summarization; null elsewhere. |
| `diagram_decls[]` | array of `template_diagram_decl.v1` | No | No | `[]` | Items from Table 21-I; `decl_id` becomes `diagram_id`. |

Expansion is a projection of retained content, after lifecycle eligibility and
Core redaction. Its dimension key is the exact Party ID or stable subject ref;
it is independent of labels, structural IDs, chunking, and the current order of
source arrays. Internal Party dimensions use the frozen snapshot Party identity
roster, including Parties whose display fields were removed. External Party
dimensions intersect that roster with the exact recipient partitions. Subject
dimensions use only retained `subjects[]`. No dimensions for a required
declaration fail with `required_section_empty`; optional declarations emit none.

Within a Party instance, a selected field must have a nonempty partition set
contained in `public` and that Party's partition; internal scopes additionally
permit `internal_only`. A field requiring multiple Parties is not implicitly
assigned to any one Party instance. Within a subject instance, a field must
share an exact source ref with the retained subject. These filters apply to
ordinary bindings and narrative slots and never enlarge the release's effective
permissions. Bindings are still evaluated against admitted addresses before
structural renumbering. Optional empty instances preserve gaps in dimension
ordinals rather than renumbering the surviving dimensions.

Field-valued section bindings resolve through typed `sections`, `records`, and
`timeline_events` collections. A collection or keyed container selects its
addressed descendant fields; a keyed field selects that field object. Record
and Timeline addresses refer to the same canonical field as its source-section
address. Selecting several such views in one section emits that value once,
preserves its selected content class and provenance, and does not create another
redaction identity. A Core record binding never selects an auxiliary source or
the separate envelope projection merely because it shares the same ID.
Lifecycle opt-in uses this same binding relation. Metadata scalars, traversal
inside a field's `value`, numeric indexes, unsupported collection kinds and
query expressions are not field bindings; they must reject even on an optional
section, rather than silently produce an empty successful selection. Dedicated
diagram, asset and narrative declarations retain their own typed contracts.

Asset-valued section bindings are exactly `$.assets['{asset_id}']` or
`$.assets[*]`. They select declared template assets, not source fields, storage
objects or a field inside an asset declaration. Selected assets append typed
`asset_ref` blocks after the section's selected case fields and narrative slots,
in asset inventory order. They participate in required-section emptiness and
the ordinary section expansion limits. A missing exact asset binding is an
invalid declaration, even for an optional section.

An expansion title uses an eligible selected display token first. A Party may
then use its retained, instance-eligible `display_name`; a subject does not fall
back to an unmasked display field. Remaining fallbacks are `Party NNNN` and
`Subject NNNN`, using the 1-based dimension ordinal. Case-derived title labels
retain their exact selected source/support refs and partitions in deck evidence.
The fixed placeholder introduces no other template expression language. Repeated
sections, blocks, and fields count on each emission against the existing §25
limits. Section anchors bind the declaration and exact dimension key.

**Table 21-D. Layout declaration schema (`layouts[]` items)**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `layout_id` | `identifier` | Yes | No | None | Unique closed token in the manifest. |
| `slidev_layout` | string | Yes | No | None | Exactly one of `default`, `cover`, or `section`; emitted verbatim as the per-slide frontmatter `layout` value. |
| `allowed_block_kinds[]` | array | Yes | No | None | Subset of Table 9-C `block_kind` vocabulary. A slide block outside the set fails with `failure_code='deck_model_invalid'` and `reason_code='block_kind_not_allowed'`. |

**Table 21-E. Narrative slot schema (`narrative_slots[]` items)**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `slot_id` | `identifier` | Yes | No | None | Unique in manifest. |
| `section_decl_id` | `member_name` | Yes | No | None | Must name one Table 21-C section declaration. |
| `binding` | `binding_path_pattern_v1` | Yes | No | None | MUST address snapshot artifact fields using `$.records[record-key].fields[field-key]`, where each key is a quoted identity or the one wildcard allowed by the path grammar. It must be a subset of `allowed_export_model_bindings[]`. |
| `required` | boolean | No | No | `false` | Unbound required slots fail with `failure_code='export_model_invalid'` and `reason_code='narrative_slot_unbound'`; unbound optional slots emit nothing. |
| `max_chars` | `finite_integer` | Yes | No | None | Positive maximum scalar count for slot content. |

Slot content MUST be sourced only from snapshot artifact records through the declared binding. It MUST NOT be sourced from the release request.

Slots are evaluated in manifest order after ordinary section bindings, for each
emitted section instance. Matching records sort by `record_id`, then matching
fields by `field_key`, bytewise ascending. Any matched non-Artifact record is an
invalid binding (`export_model_invalid` / `invalid_scalar_value`), even if Core
redaction would later remove it. Each retained nonempty string emits one
`paragraph` block preserving the field's content class, selected value, source
state, refs, and disclosure partitions. Null, dropped, empty, or
instance-ineligible fields emit nothing; other retained types fail with
`export_model_invalid` / `invalid_type`. A required slot with no emitted content
fails with `narrative_slot_unbound`. No fallback value is inferred.

`max_chars` bounds the sum of Unicode scalar counts of emitted values per slot
instance, before source escaping. It neither counts Markdown escaping bytes nor
permits silent truncation. Exceeding it fails with `export_model_invalid` /
`invalid_scalar_value`. A slot cannot concatenate values into another source
field, reclassify content as presentation text, or reopen live source storage.

**Table 21-F. Render profile schema (`render_profiles[]` items)**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `render_environment_profile_id` | `identifier` | Yes | No | None | The release tuple member MUST name one declared profile. |
| `viewport_css_px` | string | No | No | `1280x720` | Exact `1280x720` in v1; any other value is invalid. |
| `device_scale_factor` | `finite_integer` | No | No | `1` | Exact integer `1` in v1; any other value is invalid. |
| `color_scheme` | string | No | No | `light` | `light` or `dark`. |
| `locale` | string | No | No | `C.UTF-8` | Exact `C.UTF-8`; any other value is invalid in v1. |

**Table 21-G. `template_asset_item.v1` (`assets[]` items)**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `template_asset_item.v1`. |
| `asset_id` | `identifier` | Yes | No | None | Unique in the template manifest. |
| `asset_role` | string | Yes | No | None | `local_asset` or `local_theme`. |
| `bundle_path` | `bundle_path` | Yes | No | None | Must match the Table 22-B path pattern for the role. |
| `source_path` | `bundle_path` | Yes | No | None | POSIX-relative path inside the template pack. |
| `media_type` | `media_type` | Yes | No | None | Must be allowed for the asset role. |
| `sha256` | `sha256_hex` | Yes | No | None | SHA-256 of exact local asset bytes. |
| `byte_size` | `finite_integer` | Yes | No | None | Exact local asset byte count; must satisfy §25. |
| `required_for_release` | boolean | Yes | No | `true` | V1 assets referenced by source are always release-required. |

**REQ-RPT-103d**
Reporting owns the immutable template pack registry independently of Reference
Packs. Admission selects an exact `(template_id, template_version)` from that
registry; duplicate identities fail assembly. The registered manifest and all
declared regular-file bytes are copied into owner-controlled immutable values.
No release or preview request supplies file paths, bytes, URLs, executable code
or a resolver capability. Source paths are unique within a pack and files are
ordered by source path. Missing, extra, repeated or hash/size-mismatched file
contents fail admission. A pack installer must reject links and special files
before constructing these regular-file values; the render path does not open
source paths at all.

Release and preview admission retain the complete manifest and exact file bytes
in the frozen job input. An independently retained job-row digest binds the
manifest, including every declared asset's path, media type, byte size and hash.
Replay, completion and recovery validate that binding and the exact bytes
without reselecting a current registry version. A self-consistent replacement
manifest and replacement file hashes do not satisfy the original admission.
Pre-production history without this independent binding requires clean
initialization; provenance must not be inferred or backfilled.

Template image assets are static public presentation, classified
`template_boilerplate` with the `public` partition. They cannot carry incident
content, raw Evidence bytes or source-storage capabilities. Selection records
one rule decision at `$.assets['{asset_id}']`, with path, class and default
precedence. Only allow and drop preserve this typed binary boundary; scalar
mask, stub, truncate and token-backed actions fail unresolved. Drop removes the
declaration, reference blocks and copied bytes before serialization. Other asset
origins require their own explicit typed admission and disclosure contract.
The retained model materializes the complete Table 9-G declaration for each
retained template asset. Its declaration is bound before rendered bytes exist;
generated diagram and PDF files remain owned by the bundle's output inventory.

**REQ-RPT-103e**
Local theme assets use `reporting_theme_css_v1`, a closed presentation profile
over the installed `default` theme. They do not install a Slidev package, expose
internal component selectors, or supply executable setup. A template containing
local theme assets MUST support only `slidev`. It MUST declare exactly one
`text/css; charset=utf-8` entry at `theme/style.css`, at most two `font/woff2`
files at `theme/fonts/{asset_id}.woff2`, and at most one PNG or passive SVG logo
at `theme/images/{asset_id}.png` or `.svg`. Theme font and image IDs match
`[a-z][a-z0-9_-]{0,63}`. All theme files are required, immutable public
`template_boilerplate`; all participate in asset redaction accounting regardless
of section bindings. Only an allow decision admits a complete theme. Dropping
or transforming any required theme member fails closed rather than silently
substituting default presentation. Theme files never produce image-content
`asset_ref` blocks; an exact section binding to a theme member is invalid.

The stylesheet is at most 8,192 ASCII bytes, with LF/tab/space whitespace and no
comments, escapes, CR, imports, nested rules, expressions, custom selectors or
unrecognized properties. Each property occurs once in its block and ends in a
semicolon. There is exactly one `:root` block with at least one property from
the following table, preceded by zero to two `@font-face` blocks. Rules and
properties may vary in order within those constraints; their original bytes
remain immutable. Rejection never rewrites CSS into a different accepted file.

| Root property | Accepted value and effect |
| --- | --- |
| `--cartulary-report-background` | Six-digit hexadecimal color; slide background. |
| `--cartulary-report-text` | Six-digit hexadecimal color; body text. |
| `--cartulary-report-accent` | Six-digit hexadecimal color; headings and table heading borders. |
| `--cartulary-report-font-family` | `sans-serif`, `serif`, `monospace`, or quoted `rpt-{asset_id}` for a declared font; body family. |
| `--cartulary-report-heading-font-family` | Same family vocabulary; headings. |
| `--cartulary-report-font-size` | Integer `18px` through `32px`; body size. |
| `--cartulary-report-logo` | Exact `url("images/{asset_id}.png")` or `.svg` for the declared logo; renderer-owned upper-right logo box, at most 100 by 48 CSS pixels, without changing document flow. |

Each declared font has exactly one `@font-face` with exactly these properties:
`font-family: "rpt-{asset_id}"`,
`src: url("fonts/{asset_id}.woff2") format("woff2")`,
`font-style: normal`, `font-weight: 400`, and `font-display: block`.
Every declared font is selected by a root family property; the declared logo is
selected by the logo property. Missing or unreferenced dependencies fail. Only
these exact relative URLs are accepted. Remote, absolute, encoded, query,
fragment, traversal and data URLs fail before renderer execution.

Font files are at most 16 MiB, identify a WOFF2 TrueType or OpenType font rather
than a collection, declare their exact file length, 1–256 tables, a nonempty
bounded compressed block, and at most 64 MiB of reconstructed font data.
Reserved fields and optional metadata/private blocks are zero in this profile.
Header admission does not assert that compressed font data is valid: the
confined browser MUST successfully decode and load every declared font before
PDF production. Missing styles, failed fonts and fallback caused by a declared
font's load failure fail the attempt. The complete usable font and font
configuration inventory, including this stylesheet and its font files, is
bound by the toolchain snapshot; the installed font digest alone is insufficient.

**Table 21-H. `template_output_option_support.v1` (`supported_output_options[]` items)**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `template_output_option_support.v1`. |
| `output_kind` | string | Yes | No | None | `mermaid` or `slidev`; must appear in `supported_output_kinds[]`. |
| `pdf` | boolean | Yes | No | `false` | Valid support only for `slidev`. Must be `false` for `mermaid`. |
| `svg` | boolean | Yes | No | `false` | Valid support only for `mermaid` or rendered diagrams. |
| `png` | boolean | Yes | No | `false` | Exact `false` for every current renderer profile. A declaration of `true` fails template admission. |
| `pptx` | boolean | Yes | No | `false` | Exact `false` for every current renderer profile. A declaration of `true` fails template admission. |
| `rendered_diagrams` | boolean | Yes | No | `true` | Must be `true` for external releases containing diagrams. |

**Table 21-I. `template_diagram_decl.v1` (`diagram_decls[]` items)**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `template_diagram_decl.v1`. |
| `decl_id` | `identifier` | Yes | No | None | Unique in the manifest; becomes `diagram_id`. |
| `diagram_kind` | string | Yes | No | None | `flowchart` or `sequence`. |
| `diagram_source_kind` | string | Yes | No | None | `graph` or `timeline`. `template_static` is future-only in v1. |
| `source_graph_view_id` | `identifier` | Yes | Yes | None | Required when `diagram_source_kind='graph'`; otherwise null. |
| `selection_rule` | `diagram_selection_rule.v1` | Yes | No | None | Closed selection rule from Table 15-C1. |
| `required` | boolean | Yes | No | `false` | If true and the diagram cannot be produced, render fails with the mapped diagram failure. |

**REQ-RPT-103b**
Template nested objects in Tables 21-C through 21-I are closed. Unknown members, duplicate declaration IDs, output-option rows for unsupported output kinds, missing per-kind output-option rows, asset path/media mismatches, undeclared asset references, retired aggregate allowlists, and invalid diagram declarations MUST fail before render output bytes with `error.code='invalid_release_request'` and `reason_code='template_manifest_invalid'` unless a more specific reason code in §23.6 applies. Template `assets[]` MUST NOT contain remote URLs, absolute paths, path traversal, `data:` URIs, `javascript:` URIs, symlinks, hard links, filesystem-special entries, or media types outside Table 22-B.

**REQ-RPT-103c**
Template manifest v3 rejects the retired aggregate allowlist member under
REQ-RPT-060b. Its `manifest_version` is exactly `3`. The required built-in
candidate is `cartulary.report.default@4`; earlier built-in versions are retired
without aliases. Redaction profile schema v3 and normalized view v2 omit the
retired bucketing setting. The current built-in redaction profile version is
`5`; prior built-in revisions cannot satisfy current admission or replay.

**REQ-RPT-104**
Local assets MUST be declared in the template manifest before use. A generated source file that references an undeclared asset, remote asset, absolute file path, path traversal, `data:` URI, `javascript:` URI, or non-relative URL MUST fail with `failure_code='undeclared_asset'` or `failure_code='remote_asset_reference'` as applicable.

An `asset_ref` block MUST carry an `asset_id` that resolves to exactly one Table 21-G asset item. No-match, duplicate-match, or mismatch between a serialized path and the declared `bundle_path` fails with `failure_code='undeclared_asset'` and `reason_code='undeclared_asset_ref'`. A generated or template-provided source reference whose target begins with `http:`, `https:`, `//`, absolute `file:`, `data:`, or `javascript:` fails with `failure_code='remote_asset_reference'` and `reason_code='remote_asset_reference_invalid'` before renderer execution.

# 22. Render bundle manifest, file roles, and archive boundary

## 22.1 Bundle path and archive boundary

**REQ-RPT-105**
Bundle paths MUST be POSIX-relative UTF-8 strings in Unicode NFC with no leading `/`, no empty segment, no `.`, no `..`, no backslash, no NUL, and no duplicate exact byte sequence after NFC validation. Sorting is bytewise ascending UTF-8 by `path`.

**REQ-RPT-106**
Physical archive bytes are a delivery wrapper, not the approval hash. `output_sha256` binds to `bundle_manifest_sha256`. Archive extraction for verification MUST reject symlinks, hardlinks, device files, absolute paths, path traversal, duplicate paths, and filesystem-special entries.

**REQ-RPT-106a**
`manifest.json` is the physical root bundle manifest object. It MUST NOT appear as a `files[]` child item and MUST NOT be assigned a self-referential file-item digest. For bundle-backed outputs, `output_sha256` MUST equal `bundle_manifest_sha256`.

## 22.2 Bundle manifest schema

**REQ-RPT-107**
`cartulary.render_bundle_manifest.v3` MUST include exactly the fields in Table 22-A. Unknown members are invalid.

**Table 22-A. Bundle manifest top-level schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.render_bundle_manifest.v3`. |
| `manifest_version` | string | Yes | No | None | Exact `3`. |
| `release_id` | `identifier` | Yes | Yes | None | Exact admitted release identity; null for preview. |
| `preview_attempt_id` | `identifier` | Yes | Yes | None | Exact admitted preview identity; null for release. Exactly one identity is non-null. |
| `render_admitted_at` | `timestamp` | Yes | No | None | Frozen admission time with six fractional digits. |
| `snapshot_id` | `identifier` | Yes | No | None | Exact frozen snapshot identity. |
| `export_model_id` | `identifier` | Yes | No | None | Deterministic render-bound model identity from §10. |
| `admitted_export_model_sha256` | `sha256_hex` | Yes | No | None | Schema-domain digest of the frozen admitted model before redaction and final assembly; equals the release/preview admission's `export_model_sha256`. |
| `export_model_sha256` | `sha256_hex` | Yes | No | None | Schema-domain digest of the final retained, expanded canonical export model consumed by renderer-source generation. |
| `renderer_id` | `identifier` | Yes | No | None | Exact producer identity, independently checked by native completion/recovery. |
| `renderer_version` | string | Yes | No | None | Exact producer version; no inferred historical producer. |
| `template_id` | `identifier` | Yes | No | None | Exact admitted template. |
| `template_version` | string | Yes | No | None | Exact admitted version. |
| `template_manifest_sha256` | `sha256_hex` | Yes | No | None | Schema-domain digest of the complete admitted template manifest. |
| `output_kind` | string | Yes | No | None | `slidev` or `mermaid`. |
| `release_scope` | string | Yes | No | None | From §7.3. |
| `bundle_created_at` | `timestamp` | Yes | No | None | Equals normalized `render_admitted_at`. |
| `reference_packs` | `reference_pack_set_binding.v1` | Yes | No | None | Exact snapshot binding from REQ-RPT-046rp. |
| `primary_path` | string | Yes | No | None | Exact path of one required `files[]` member. |
| `primary_media_type` | string | Yes | No | None | Matches that primary member. |
| `toolchain_snapshot_sha256` | `sha256_hex` | Yes | No | None | Schema-domain digest of the required canonical toolchain snapshot for successful output. |
| `validation_summary_sha256` | `sha256_hex` | Yes | No | None | Schema-domain digest of the canonical passing validation summary. |
| `redaction_manifest_sha256` | `sha256_hex` | Yes | No | None | Schema-domain digest of the canonical redaction manifest. |
| `token_manifest_sha256` | `sha256_hex` | Yes | Yes | None | Required when tokens are used; otherwise explicit null. |
| `external_determinism_sha256` | `sha256_hex` | Yes | Yes | None | Non-null deterministic renderer-output comparison digest for external output; null for internal output. |
| `generated_presentation` | boolean | Yes | No | None | Exact true; this is generated presentation, not source evidence. |
| `self_contained` | boolean | Yes | No | None | Exact true after local-file and sandbox validation. |
| `files[]` | array | Yes | No | None | Nonempty file inventory ordered by `path` bytewise ascending UTF-8. |
| `validation_artifacts[]` | array of string | Yes | No | None | Exact ordered paths `validation/summary.json`, `validation/toolchain.json`; both occur in `files[]`. |

The v3 clean cutover retires v2 manifests without a reader or identity-conversion
fallback. `bundle_manifest_sha256` is the schema-domain digest of the exact
canonical manifest bytes, retained in the owning resource/descriptor; it is not
a member of its own hashed object. Native completion and recovery MUST compare
both model digests against their distinct independently reconstructed inputs.
Rehashing a mutually consistent bundle cannot replace either identity. Portable
readers validate the retained manifest and file inventory without fabricating
missing admission facts or treating an input digest as proof of redacted bytes.
This binding does not add private canonical-model content to shareable bundles.


**REQ-RPT-107a**
`bundle_manifest_sha256` MUST be computed over `reporting_canonical_json_v1` serialization of the manifest object after deleting only the top-level `bundle_manifest_sha256` member. Every listed child file digest MUST be validated before the manifest is publishable. A verifier MUST first recompute `bundle_manifest_sha256` from the root `manifest.json`, then validate every listed non-manifest file item. A missing root `manifest.json`, any listed child file with role `manifest`, or any listed child file with path `manifest.json` MUST fail bundle validation with `failure_code='bundle_manifest_invalid'` and `reason_code='bundle_manifest_self_reference'`.

**REQ-RPT-108**
Every file item MUST contain `path`, `role`, `media_type`, `byte_size`, `sha256`, and `required_for_release`. File item roles, path patterns, media types, required conditions, and `required_for_release` values MUST follow Table 22-B.

**Table 22-B. File role, path, and media-type matrix**

| Role | Required path pattern | Allowed media type | Required when | `required_for_release` |
| --- | --- | --- | --- | --- |
| `validation_summary` | `validation/summary.json` | `application/vnd.cartulary.reporting-validation+json` | Always | `true` |
| `toolchain_snapshot` | `validation/toolchain.json` | `application/vnd.cartulary.reporting-toolchain+json` | Always | `true` |
| `redaction_manifest` | `validation/redaction-manifest.json` | `application/vnd.cartulary.redaction-manifest+json` | Always | `true` |
| `redaction_profile_view` | `validation/redaction-profile-view.json` | `application/vnd.cartulary.redaction-profile-view+json` | Always; frozen selected-rule verification input | `true` |
| `token_manifest` | `validation/token-manifest.json` | `application/vnd.cartulary.reporting-token-manifest+json` | When tokens used | `true` |
| `source_slidev` | `slides.md` | `text/markdown; charset=utf-8` | `slidev` | `true` |
| `deck_model` | `intermediate/deck-model.json` | `application/vnd.cartulary.reporting-deck-model+json` | `slidev`; exact final deck used by the serializer | `true` |
| `diagram_model` | `intermediate/diagrams/{diagram_id}.json` | `application/vnd.cartulary.reporting-diagram+json` | Every retained diagram; exact typed model used by its serializer | `true` |
| `source_mermaid` | `diagrams/{diagram_id}.mmd` | `text/vnd.cartulary.mermaid; charset=utf-8` | Every retained auto-layout diagram under REQ-RPT-086a, including auto-layout diagrams embedded in Slidev decks | `true` |
| `rendered_pdf` | `deck.pdf` | `application/pdf` | External `slidev` | `true` |
| `rendered_pptx` | `deck.pptx` | `application/vnd.openxmlformats-officedocument.presentationml.presentation` | Future-only; rejected by current renderer profiles | `true` |
| `rendered_png` | `slides/page-{0001..N}.png` | `image/png` | Future-only; rejected by current renderer profiles | `true` |
| `rendered_svg` | `diagrams/{diagram_id}.svg` | `image/svg+xml` | External `mermaid` for every retained auto-layout diagram, every retained manual-layout diagram in Slidev, or requested auto-layout diagram render | `true` |
| `rendered_png` | `diagrams/{diagram_id}.png` | `image/png` | Future-only; rejected by current renderer profiles | `true` |
| `local_asset` | `assets/{asset_id}/{filename}` | `image/png`, `image/jpeg`, sanitized `image/svg+xml` | Referenced by source | `true` |
| `local_theme` | `theme/{path}` | `text/css; charset=utf-8`, `font/woff2`, sanitized `image/svg+xml`, `image/png` | Referenced by template | `true` |

**REQ-RPT-108a**
Sanitized `image/svg+xml` local assets and local themes MUST satisfy the rendered SVG safety subset in REQ-RPT-122b before they are copied into a bundle. HTML, JavaScript, executable content, remote URLs, `data:` URLs, unsanitized SVG, `font/ttf`, and media types outside Table 22-B are invalid for v1 local assets and themes. A role/path/media mismatch MUST fail with `failure_code='bundle_manifest_invalid'` and `reason_code='invalid_asset_media_type'` before publishable bytes exist.

The same role/path/media/requiredness admission applies at native persistence,
portable import and export, and recovery validation. An allowed role alone does
not authorize an arbitrary path or media type. NFC validation rejects a
non-normalized path; it MUST NOT rename or normalize an admitted member. Optional
output kinds become valid only through an implemented and qualified producer;
their appearance in this matrix does not authorize accepting opaque files under
an otherwise unsupported role. Intermediate models and the frozen public
redaction profile view are required replay inputs, not sensitive reveal maps.

PNG and JPEG assets must have headers consistent with their declared media type,
positive dimensions and at most 40,000,000 decoded pixels. SVG assets pass the
passive SVG validator without repair. These checks precede sandbox launch;
browser decoding and successful image load remain required before PDF capture.
Local-image URLs are segment-wise UTF-8 percent-encoded when serialized into
Markdown. Decoding that relative URL must yield exactly the declared bundle path;
encoding cannot change the manifest path or authorize another resource.
Image filenames use `.png`, `.jpg`/`.jpeg`, or `.svg` respectively, with exact
lowercase extensions. Local-image bundle paths exclude literal `%`, `?` and `#`
to keep URL routing unambiguous across the installed renderer's HTTP stack.

The private stage v5 transports exact file octets as canonical padded base64,
alongside path, media type and raw-byte SHA-256. Its asset digest uses the domain
`cartulary.reporting_render_assets.v2` plus NUL, followed in bytewise path order
by decimal-byte-length/colon framed path, media type and raw file octets. It has
no executable, source-directory or network-selector fields. Stage admission
allows at most 500 declared local assets and theme members (50 MiB each, 250 MiB total) and 100
manual diagram SVGs (8 MiB each, 32 MiB total); operative template limits remain
additional. Source text is bounded at 64 MiB and the entire private JSON request
at 768 MiB. Unknown fields, noncanonical base64, duplicate/escaping paths,
hash disagreement and other media reject before copying. The runtime writes
only into the fresh attempt's private `public/assets`, `public/diagrams` and `public/theme`
trees with exclusive file creation. The bounded theme profile and complete
font inventory and the closed structural-redaction model define renderer version 21; earlier renderer evidence cannot qualify
these added paths.

# 23. Validation, error-envelope model, and reason codes

## 23.1 Error-envelope model

**REQ-RPT-109**
Reporting errors MUST use the three-level model in Table 23-A. Public payloads and normative requirements MUST NOT combine the Core public error family and Reporting failure code into one colon-delimited machine-readable identifier.

**Table 23-A. Error layers**

| Layer | Field | Rule |
| --- | --- | --- |
| Core public error | `error.code` | `invalid_release_request` before an admitted durable render job; `release_render_failed` after admission. |
| Reporting failure | `failure_code` | Closed registry in §23.5; null only on pass or before route admission when `error.code='invalid_release_request'`. |
| Specific reason | `reason_code` | Closed registry in §23.6; null only when no finer reason exists. |

## 23.2 Validation summary schema

**REQ-RPT-110**
`cartulary.reporting_render_validation_summary.v1` MUST use Table 23-B.

**Table 23-B. Validation summary schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_render_validation_summary.v1`. |
| `release_id` | `identifier` | Yes | Yes | None | Required when allocated; otherwise null with `safe_details.pre_release=true`. |
| `snapshot_id` | `identifier` | Yes | Yes | None | Required when known; otherwise null only before snapshot binding. |
| `result` | string | Yes | No | None | `passed` or `failed`. |
| `terminal_stage` | string | Yes | No | None | Closed stage token from §23.3. |
| `failure_code` | string | Yes | Yes | None | Null only on pass or before route admission when no Reporting failure applies. |
| `reason_code` | string | Yes | Yes | None | Null only when no finer reason exists. |
| `safe_details` | `safe_details.v1` | Yes | No | `{}` | Closed safe-details object; always present. |
| `issue_count` | `finite_integer` | Yes | No | `0` | Total evaluated issue count under REQ-RPT-113a, not only retained `issues[]` length. |
| `issues[]` | array | Yes | No | `[]` | Table 23-B1 items; empty only on pass. |
| `first_failure` | `first_failure.v1` | Yes | Yes | None | Null on pass; otherwise copy of first issue identity and codes. |
| `created_at` | `timestamp` | Yes | No | None | Equals normalized `render_admitted_at` for bundle-retained summary. |

**Table 23-B1. Validation issue item schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `stage` | string | Yes | No | None | Token from §23.3. |
| `severity` | string | Yes | No | None | `error` or `warning`. |
| `export_model_path` | `export_model_path_v1` | Yes | Yes | None | Null when not attributable to export-model content. |
| `bundle_path` | `bundle_path` | Yes | Yes | None | Null when not attributable to a bundle file. |
| `failure_code` | string | Yes | Yes | None | Registry value or null. |
| `reason_code` | string | Yes | Yes | None | Registry value or null. |
| `safe_details` | `safe_details.v1` | Yes | No | `{}` | Restricted by Table 23-C plus reason-code-specific tables. |

## 23.3 Stage vocabulary

**REQ-RPT-111**
Validation stage tokens MUST use this exact ordered vocabulary:

```text
route_admission
composition_validation
export_model_materialization
redaction
token_manifest
diagram_model
mermaid_source
mermaid_render
deck_model
slidev_source
slidev_export
bundle_manifest
determinism
security_sandbox
release_state
```

## 23.4 Issue ordering and safe details

**REQ-RPT-112**
Validation issues MUST sort by the Table 23-B2 key. A conforming implementation MUST use this ordering for retained `issues[]`, `first_failure`, terminal failure selection when multiple issues are found in the same stage, and deterministic validation-summary bytes. It MUST NOT use source discovery order, map iteration order, database row order, filesystem order, or renderer diagnostic order except where Table 23-B2 explicitly names a source index.

**Table 23-B2. Validation issue sort key**

| Key position | Sort key | Rule |
| ---: | --- | --- |
| 1 | Stage | Ordered by §23.3 stage vocabulary. |
| 2 | Severity | `error` before `warning`. |
| 3 | Source kind | `export_model_path`, then `bundle_path`, then `composition_source`, then `safe_detail_only`, then `unattributed`. |
| 4 | Primary path | Canonical `export_model_path` or `bundle_path`; null sorts after all non-null paths. |
| 5 | Source array name | `deck_ops`, `diagram_decls`, `authored_texts`, `layout.node_positions`, `layout.edge_routes`, `graph_projection_refs`, `assets`, `sections`, `blocks`, `fields`, then exact code point order for future named arrays. |
| 6 | Source index | Zero-based array index after default materialization and canonical array ordering; absent index sorts after all numeric indexes. |
| 7 | `failure_code` | Null sorts after non-null; otherwise exact code point order. |
| 8 | `reason_code` | Null sorts after non-null; otherwise exact code point order. |
| 9 | Safe details | `reporting_canonical_json_v1` serialization of `safe_details` after forbidden-value filtering. |

**REQ-RPT-113**
`safe_details.v1` is the embedded bounded diagnostic map defined by Table 23-C; it intentionally has no `schema_id` member so that `{}` remains the canonical empty safe-details value. It MAY contain only the keys in Table 23-C unless a reason-code-specific table in this NLSpec adds a key. Omission of an allowed key means the detail is unavailable or not attributable; omission MUST NOT change failure classification. Unknown keys are invalid. Forbidden values in Table 23-D MUST NOT appear in safe details. Issue severity MUST be exactly `error` or `warning`; no other severity token is valid.

**Table 23-C. Generic safe-details allow-list**

| Allowed key | Type | Rule |
| --- | --- | --- |
| `field` | string | Schema member name or null-free diagnostic field token. |
| `path` | string | `export_model_path_v1` or `bundle_path` string already safe to expose. |
| `stage` | string | §23.3 stage token. |
| `limit_key` | string | Table 24-B `declared_limits` timeout key or Table 25-A resource-limit key. |
| `limit` | `finite_integer` or string | Numeric limit or byte-size string from §25. |
| `actual` | `finite_integer` or string | Observed count, byte size, or closed token; no raw content. |
| `template_id` | `identifier` | Template identity. |
| `template_version` | `identifier` | Template version. |
| `output_kind` | string | §7.4 token. |
| `release_scope` | string | §7.3 token. |
| `graph_view_id` | `identifier` | Graph view identity. |
| `projection_result_id` | `identifier` | Immutable projection result identity. |
| `file_role` | string | Table 22-B role token. |
| `media_type` | `media_type` | Allowed media type. |
| `bundle_path` | `bundle_path` | Bundle path. |
| `pre_release` | boolean | True only before release ID allocation. |
| `blocked_core_dependency` | string | Dependency ID from §5. |
| `composition_id` | `identifier` | Composition identity from the release tuple. |
| `composition_version` | `identifier` | Composition version from the release tuple. |
| `composition_op_id` | `identifier` | Companion-owned operation identity when safe and attributable. |
| `composition_anchor_kind` | string | `section_anchor`, `record_anchor`, `block_anchor`, or `diagram_anchor`. |
| `diagram_id` | `identifier` | Diagram identity when safe and attributable. |
| `ordering_mode` | string | `projection_output_order` or `bytewise_ref_sort`. |
| `network_policy_id` | `identifier` | Render sandbox policy ID. |
| `non_loopback_outbound_attempt_count` | `finite_integer` | Count only; no URL or hostname. |
| `first_blocked_destination_class` | string | `dns`, `tcp`, `udp`, `http`, `https`, `file`, `data`, `javascript`, or `other`. |
| `timeout_kind` | string | `stage` or `total`. |
| `process_group_signal` | string | `terminate` or `kill`. |

**Table 23-C1. `first_failure.v1` schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `first_failure.v1`. |
| `stage` | string | Yes | No | None | Stage token copied from the first issue after §23.4 ordering. |
| `severity` | string | Yes | No | None | `error` or `warning`. |
| `export_model_path` | `export_model_path_v1` | Yes | Yes | None | Copied from the first issue. |
| `bundle_path` | `bundle_path` | Yes | Yes | None | Copied from the first issue. |
| `failure_code` | string | Yes | Yes | None | Copied from the first issue. |
| `reason_code` | string | Yes | Yes | None | Copied from the first issue. |
| `safe_details` | `safe_details.v1` | Yes | No | `{}` | Copied from the first issue after forbidden-value filtering. |

**REQ-RPT-113b**
Generic schema-validation failures MUST use Table 23-C2 unless a narrower table in this NLSpec names a more specific reason code. The earliest attributable stage is the stage whose input object failed validation. The public `error.code` is `invalid_release_request` before render admission and `release_render_failed` after admission.

**Table 23-C2. Generic schema-validation failure mapping**

| Condition | `failure_code` after admission | `reason_code` |
| --- | --- | --- |
| Top-level `schema_id` missing or not exact | `export_model_invalid` or stage-specific failure | `invalid_schema_id` |
| Required member omitted | `export_model_invalid` or stage-specific failure | `required_member_missing` |
| Unknown object member | `export_model_invalid` or stage-specific failure | `unknown_object_member` |
| Explicit null where `Nullable=No` | `export_model_invalid` or stage-specific failure | `invalid_null` |
| Value type does not match schema | `export_model_invalid` or stage-specific failure | `invalid_type` |
| Scalar violates its contract | `export_model_invalid` or stage-specific failure | `invalid_scalar_value` |
| Array not in required order | `export_model_invalid` or stage-specific failure | `array_order_invalid` |
| Duplicate array item where unique values are required | `export_model_invalid` or stage-specific failure | `duplicate_array_item` |

Stage-specific failure means the failure code that owns the object being validated, for example `deck_model_invalid` for a slide-deck object, `slidev_source_invalid` for Slidev source inputs, or `bundle_manifest_invalid` for bundle-manifest objects.

**REQ-RPT-113a**
Validation issue evaluation MUST be exhaustive within each stage up to and including the terminal stage. `issues[]` retains the first `500` issues in §23.4 order. `issue_count` is the total evaluated issue count, not only the retained length. When the total evaluated issue count exceeds `2000`, evaluation MUST stop at the end of that stage. The summary MUST retain a `validation_summary_truncated` warning with `reason_code='validation_issue_limit_reached'` at that terminal stage. Reserve one of the 500 retained positions for this warning: retain the first 499 evaluated issues plus the warning, ordered by §23.4. The synthetic retention warning does not increase `issue_count`; its safe details are `limit_key='validation_issues.count'`, `limit=2000`, and `actual=issue_count`. This explicit exception to prefix retention ensures that truncation is observable even when more than 500 earlier errors sort ahead of the warning. Evaluation MUST NOT stop midway through a stage; retained diagnostic memory MAY remain bounded while the evaluated count increases.

**Table 23-D. Forbidden safe-detail values**

| Forbidden value family |
| --- |
| raw incident content |
| raw evidence bytes |
| object-store keys |
| raw storage paths |
| credentials or secret references |
| redacted display values |
| reveal-map values |
| raw source values for disallowed partitions |
| external URLs from rejected templates |

## 23.5 Failure-code registry

**REQ-RPT-114**
Reporting failure codes MUST use Table 23-E.

**Table 23-E. Failure-code registry**

| `failure_code` | Stage family |
| --- | --- |
| `export_model_invalid` | export model materialization |
| `export_model_resource_limit_exceeded` | export model materialization |
| `content_class_missing` | export model materialization |
| `composition_invalid` | composition validation and deck overlay |
| `redaction_manifest_invalid` | redaction |
| `token_manifest_invalid` | token manifest |
| `graph_projection_unavailable` | diagram model |
| `mermaid_invalid` | Mermaid source |
| `mermaid_render_failed` | Mermaid render |
| `deck_model_invalid` | deck model |
| `slidev_source_invalid` | Slidev source |
| `slidev_export_failed` | Slidev export |
| `undeclared_asset` | template or bundle validation |
| `remote_asset_reference` | security sandbox |
| `bundle_manifest_invalid` | bundle manifest |
| `bundle_resource_limit_exceeded` | bundle manifest |
| `asset_limit_exceeded` | template or bundle validation |
| `support_ref_limit_exceeded` | export model materialization |
| `validation_summary_truncated` | validation summary warning |
| `nondeterministic_render` | determinism |
| `security_sandbox_violation` | security sandbox |
| `render_canceled` | lifecycle |
| `render_timeout` | lifecycle |

## 23.6 Reason-code registry

**REQ-RPT-115**
Reason codes MUST use the closed registry in Table 23-F. A later revision MAY add reason codes only when the failure-code mapping and safe-details schema are defined in the same revision.

**Table 23-F. Reason-code registry**

| `reason_code` | Typical `failure_code` |
| --- | --- |
| `unsupported_output_kind` | null before admission |
| `unsupported_output_option` | null before admission |
| `required_output_omitted` | null before admission |
| `source_only_external_release_invalid` | null before admission |
| `source_only_conflict` | null before admission |
| `unsupported_template_feature` | null before admission |
| `unsupported_derivation_algorithm` | `export_model_invalid` |
| `unsupported_source_family` | `export_model_invalid` |
| `composition_tuple_incomplete` | null before admission |
| `composition_not_found` | null before admission |
| `composition_version_not_found` | null before admission |
| `composition_digest_mismatch` | `composition_invalid` |
| `composition_template_mismatch` | `composition_invalid` |
| `composition_anchor_unresolved` | `composition_invalid` |
| `composition_anchor_ambiguous` | `composition_invalid` |
| `composition_drop_invalid_for_external_release` | `composition_invalid` |
| `composition_duplicate_diagram_insert` | `composition_invalid` |
| `composition_replacement_target_missing` | `composition_invalid` |
| `authored_text_not_permitted` | `composition_invalid` |
| `authored_subject_ref_unresolved` | `composition_invalid` |
| `authored_title_limit_exceeded` | `composition_invalid` |
| `authored_text_limit_exceeded` | `composition_invalid` |
| `diagram_label_override_invalid` | `composition_invalid` |
| `diagram_layout_invalid` | `composition_invalid` |
| `diagram_layout_missing_node_position` | `composition_invalid` |
| `diagram_layout_duplicate_target` | `composition_invalid` |
| `diagram_layout_unknown_target` | `composition_invalid` |
| `manual_layout_not_supported_for_output_kind` | `composition_invalid` |
| `invalid_schema_id` | `export_model_invalid` after admission; stage-specific failure when validating another object |
| `required_member_missing` | `export_model_invalid` after admission; stage-specific failure when validating another object |
| `unknown_object_member` | `export_model_invalid` after admission; stage-specific failure when validating another object |
| `invalid_null` | `export_model_invalid` after admission; stage-specific failure when validating another object |
| `invalid_type` | `export_model_invalid` after admission; stage-specific failure when validating another object |
| `invalid_scalar_value` | `export_model_invalid` after admission; stage-specific failure when validating another object |
| `array_order_invalid` | `export_model_invalid` after admission; stage-specific failure when validating another object |
| `duplicate_array_item` | `export_model_invalid` after admission; stage-specific failure when validating another object |
| `live_query_after_export_model` | `export_model_invalid` |
| `duplicate_object_member` | `export_model_invalid` after admission; null before admission |
| `duplicate_stable_id` | `export_model_invalid` |
| `duplicate_export_model_path` | `export_model_invalid` |
| `dangling_source_ref` | `export_model_invalid` |
| `dangling_support_ref` | `export_model_invalid` |
| `invalid_support_ref` | `export_model_invalid` |
| `snapshot_ref_mismatch` | `export_model_invalid` |
| `invalid_asset_declaration` | `export_model_invalid` |
| `content_class_missing` | `content_class_missing` |
| `invalid_generated_identifier` | `export_model_invalid` |
| `blocked_core_dependency` | `export_model_invalid` |
| `duplicate_active_party_assignment` | `export_model_invalid` |
| `token_subject_not_stable` | `token_manifest_invalid` |
| `token_subject_not_unique` | `token_manifest_invalid` |
| `redaction_profile_invalid` | `redaction_manifest_invalid` |
| `timeline_sort_key_missing` | `export_model_invalid` |
| `invalid_timestamp_value` | `export_model_invalid` |
| `deleted_record_not_releasable` | `export_model_invalid` |
| `timeline_overflow_unresolved` | `export_model_invalid` |
| `diagrams_count_exceeded` | `export_model_resource_limit_exceeded` |
| `graph_projection_stale` | `graph_projection_unavailable` |
| `graph_projection_not_completed` | `graph_projection_unavailable` |
| `graph_projection_not_bound` | `graph_projection_unavailable` |
| `graph_projection_ambiguous` | `graph_projection_unavailable` |
| `graph_projection_selection_unresolved` | `graph_projection_unavailable` |
| `graph_projection_digest_mismatch` | `graph_projection_unavailable` |
| `diagram_selection_missing_ref` | `graph_projection_unavailable` |
| `diagram_selection_duplicate_ref` | `graph_projection_unavailable` |
| `diagram_hard_limit_exceeded` | `mermaid_invalid` |
| `template_static_future_only` | null before admission; `composition_invalid` after admission |
| `invalid_mermaid_construct` | `mermaid_invalid` |
| `label_length_exceeded` | `mermaid_invalid` |
| `no_diagrams_selected` | `mermaid_invalid` |
| `slide_count_exceeded` | `slidev_source_invalid` |
| `slide_block_limit_exceeded` | `slidev_source_invalid` |
| `speaker_notes_limit_exceeded` | `slidev_source_invalid` |
| `click_step_limit_exceeded` | `slidev_source_invalid` |
| `click_export_page_count_mismatch` | `slidev_export_failed` |
| `invalid_template_binding_pattern` | `bundle_manifest_invalid` |
| `asset_limit_exceeded` | `asset_limit_exceeded` before bundle persistence; `bundle_resource_limit_exceeded` after bundle file materialization |
| `undeclared_asset_ref` | `undeclared_asset` |
| `remote_asset_reference_invalid` | `remote_asset_reference` |
| `outbound_request_observed` | `remote_asset_reference` |
| `render_canceled` | `render_canceled` |
| `render_timeout` | `render_timeout` |
| `export_model_canonical_bytes_exceeded` | `export_model_resource_limit_exceeded` |
| `sections_count_exceeded` | `export_model_resource_limit_exceeded` |
| `blocks_count_exceeded` | `export_model_resource_limit_exceeded` |
| `fields_count_exceeded` | `export_model_resource_limit_exceeded` |
| `records_count_exceeded` | `export_model_resource_limit_exceeded` |
| `relationships_count_exceeded` | `export_model_resource_limit_exceeded` |
| `subject_limit_exceeded` | `token_manifest_invalid` |
| `support_ref_limit_exceeded` | `support_ref_limit_exceeded` |
| `local_asset_count_exceeded` | `asset_limit_exceeded` |
| `local_asset_bytes_exceeded` | `asset_limit_exceeded` |
| `local_assets_total_bytes_exceeded` | `asset_limit_exceeded` |
| `rendered_pdf_bytes_exceeded` | `bundle_resource_limit_exceeded` |
| `validation_issue_limit_reached` | `validation_summary_truncated` |
| `export_model_mismatch` | `nondeterministic_render` |
| `redaction_manifest_mismatch` | `nondeterministic_render` |
| `token_manifest_mismatch` | `nondeterministic_render` |
| `toolchain_snapshot_mismatch` | `nondeterministic_render` |
| `validation_summary_mismatch` | `nondeterministic_render` |
| `bundle_manifest_mismatch` | `nondeterministic_render` |
| `bundle_file_mismatch` | `nondeterministic_render` |
| `disclosure_partition_unresolved` | `redaction_manifest_invalid` |
| `redaction_action_unresolved` | `redaction_manifest_invalid` |
| `required_section_empty` | `export_model_invalid` |
| `invalid_recipient_partition_ref` | `redaction_manifest_invalid` during preparation |
| `unknown_recipient_partition` | `redaction_manifest_invalid` during preparation |
| `recipient_partition_profile_mismatch` | `redaction_manifest_invalid` during preparation |
| `empty_export_model_sections` | `export_model_invalid` |
| `list_nesting_exceeded` | `slidev_source_invalid` |
| `block_kind_not_allowed` | `deck_model_invalid` |
| `invalid_speaker_note_sequence` | `slidev_source_invalid` |
| `invalid_field_value_for_text` | `slidev_source_invalid` |
| `invalid_slidev_construct` | `slidev_source_invalid` |
| `narrative_slot_unbound` | `export_model_invalid` |
| `deck_title_invalid` | `deck_model_invalid` |
| `invalid_template_timeout_key` | null before admission |
| `template_limit_exceeds_hard_limit` | null before admission |
| `template_manifest_invalid` | null before admission |
| `invalid_asset_media_type` | `bundle_manifest_invalid` |
| `external_reference_in_rendered_output` | `security_sandbox_violation` |
| `partial_output_cleanup_failed` | `security_sandbox_violation` |
| `atomic_publish_failed` | `bundle_manifest_invalid` |
| `bundle_manifest_self_reference` | `bundle_manifest_invalid` |
| `loopback_not_declared` | `security_sandbox_violation` |
| `sandbox_observation_missing` | `security_sandbox_violation` |
| `render_runtime_unavailable` | `security_sandbox_violation` |
| `renderer_process_failed` | `slidev_export_failed` or `mermaid_render_failed`, according to the executing stage |
| `invalid_multiline_value` | `export_model_invalid` |

# 24. Render algorithms, lifecycle, sandbox, and determinism

## 24.1 Render-attempt lifecycle

**REQ-RPT-116**
Reporting render attempts MUST use `render_attempt_lifecycle_v1` with the states in this exact order. Worker states MUST NOT be exposed as Core release states.

```text
admitted
materializing_export_model
redacting
generating_diagrams
generating_deck
rendering_outputs
validating_bundle
validating_determinism
persisting_bundle
succeeded
failed
canceled
timed_out
```

**REQ-RPT-117**
Render attempt outcomes MUST use Table 24-A.

**Table 24-A. Render attempt outcome behavior**

| Case | Required behavior |
| --- | --- |
| Success | Persist bundle, compute `output_sha256`, and update Core release state according to Core lifecycle. |
| Failure before output bytes | Retain validation summary, set durable release `render_failed`, and expose no approvable bytes. |
| Failure after partial files | Delete partial release-output files. Diagnostic retention is valid only outside the release bundle and outside approval bytes. |
| Cancellation after release row allocation | Release becomes `render_failed` with `reason_code='render_canceled'`. |
| Timeout | Release becomes `render_failed` with `reason_code='render_timeout'`; partial output is deleted or diagnostic-only. |
| Exact retry | Core idempotency returns original job or release when Core classifies it as exact replay. |
| New retry | Requires a new `client_txn_id` and new release candidate when Core permits. |

**REQ-RPT-117a**
`render_attempt_lifecycle_v1` MUST implement cancellation, timeout, and partial-output cleanup with this observable algorithm:

```text
create isolated working directory W
create unpublished bundle staging directory S
start total attempt timer
for each stage in Table 24-B order:
  start stage timer
  execute stage in a process group when the stage invokes an external renderer
  if cancellation requested:
    terminate process group, wait 2 seconds, then kill remaining processes
    delete S and every release-output file produced from W
    retain validation summary with reason_code='render_canceled'
    stop
  if stage timer or total timer fires:
    terminate process group, wait 2 seconds, then kill remaining processes
    delete S and every release-output file produced from W
    retain validation summary with reason_code='render_timeout'
    stop
  stop stage timer
validate bundle from S
publish S atomically as the release bundle only after every validation gate passes
```

If a stage timer and the total timer fire in the same scheduler tick, total timeout wins and `safe_details.timeout_kind='total'`. Otherwise the timer observed first wins. Partial release-output files MUST NOT be retained in a quarantine that is addressable as release output in v1. If cleanup cannot delete a partial release-output file, the release remains `render_failed`, the validation summary uses `failure_code='security_sandbox_violation'` and `reason_code='partial_output_cleanup_failed'`, and the release MUST NOT expose any approvable bytes.

**REQ-RPT-117b**
Bundle publish is atomic at the logical release-resource boundary. Files may be written in a staging directory or object-store prefix only while unpublished. The bundle becomes publishable only when the canonical bundle manifest, every listed file digest, validation summary, redaction manifest, token manifest when present, and `bundle_manifest_sha256` have been validated. A reader MUST observe either no publishable bundle or the complete manifest-bound bundle; it MUST NOT observe a partially published release bundle. A publish race, missing staged file, digest mismatch, or non-atomic visibility failure MUST fail with `failure_code='bundle_manifest_invalid'` and `reason_code='atomic_publish_failed'`.

**REQ-RPT-117c**
Failed and canceled admitted render attempts, including Composition previews,
MUST retain their canonical Table 23-B validation summary with the immutable
admission. The summary and terminal Jobs outcome MUST commit atomically with
the failed release state when a release identity was allocated. Diagnostic
retention MUST survive public Jobs response expiry and full recovery; it MUST
NOT turn failed output into a release bundle or add diagnostics to the portable
shareable-artifact catalog. Preview diagnostics bind the admitted preview through
its retained input, use `release_id=null` and `safe_details.pre_release=true`, and
MUST NOT allocate a synthetic release identity. Recovery MUST reject missing,
altered or admission-mismatched diagnostics, including after Jobs response expiry.
An exact terminal replay MUST preserve the original bytes; it MUST NOT reconstruct
missing diagnostics from a reason code or rerun a participant. The pre-production
cutover rejects incompatible retained history before mutation.

The total attempt deadline includes materialization and publication. A proven
commit remains authoritative if its acknowledgement arrives after the deadline.
An indeterminate commit remains subject to the shared Jobs reconciliation and
integrity policy; it MUST NOT be reclassified as a proven timeout or retried by
Reporting. A cancellation cleanup defect retains the security failure identity
from REQ-RPT-117a even when the Jobs terminal outcome is canceled.

## 24.2 Timeouts

**REQ-RPT-118**
Render timeouts MUST use Table 24-B. The operative timeout for each stage is the template-declared value when the template manifest declares the corresponding `declared_limits` key, otherwise the Default column. A template-declared timeout MUST be an integer number of seconds greater than `0` and less than or equal to the Hard max. Suffixes such as `s`, fractional values, exponent notation, byte-size units, and unknown timeout keys are invalid and fail template validation before render output bytes with `error.code='invalid_release_request'` and `reason_code='invalid_template_timeout_key'` unless the value is above the hard max, which fails with `reason_code='template_limit_exceeds_hard_limit'`.

**Table 24-B. Render attempt timeouts**

| Stage | `declared_limits` key | Default seconds | Hard max seconds |
| --- | --- | ---: | ---: |
| Export-model materialization | `timeout.export_model_materialization` | 120 | 300 |
| Redaction and tokenization | `timeout.redaction_tokenization` | 120 | 300 |
| Mermaid source/render per diagram | `timeout.mermaid_per_diagram` | 60 | 180 |
| Slidev export | `timeout.slidev_export` | 600 | 900 |
| Determinism rerender | `timeout.determinism_rerender` | Operative Slidev-export timeout for Slidev; operative Mermaid-per-diagram timeout for Mermaid | 900 |
| Total attempt | `timeout.total_attempt` | 900 | 1800 |

**REQ-RPT-118a**
`declared_limits` timeout keys are exactly the six Table 24-B keys. `timeout.determinism_rerender` inherits the operative first-render timeout when omitted; when declared, its integer value MUST be in `[1, 900]`. Timeout values are counted in elapsed seconds measured by the render supervisor for the named stage. Resource-limit byte sizes and counts continue to use §25 units and MUST NOT be accepted for timeout keys.

## 24.3 Render sandbox policy

**REQ-RPT-119**
`render_sandbox_policy_v1` MUST enforce Table 24-C. The implementation mechanism is unconstrained only when the observable evidence and failure behavior remain identical.

**Table 24-C. Render sandbox policy**

| Boundary | Rule |
| --- | --- |
| Network | Only loopback connections required to drive the local renderer are allowed; non-loopback DNS, TCP, UDP, and HTTP(S) attempts fail. |
| Package resolution | No package install or remote resolution during release rendering. |
| Filesystem read | Only render working directory, declared template pack directory, pinned package store, declared local assets, and the complete installed runtime admitted under Core 04 REQ-04-046c (Node, headless shell, launcher, shared libraries, local fonts and their configuration). Private namespace kernel interfaces required by these executables grant no host-file or host-process visibility. |
| Filesystem write | Only isolated render working directory and diagnostic directory. |
| Environment | Only `env_allowlist[]`. |
| URL schemes | Relative bundle paths only. `http:`, `https:`, protocol-relative URLs, external `file:`, `data:`, and `javascript:` are invalid. |
| Browser content | No iframes, remote embeds, scripts, arbitrary Vue, Monaco, Twoslash, code runners, or external fetch behavior. |
| Evidence | Raw evidence bytes are not mounted unless a redacted, declared local asset has already been produced. |
| Observation | Validation summary records `network_policy_id`, `non_loopback_outbound_attempt_count`, and first blocked destination class, not raw URLs. |

**REQ-RPT-119a**
The render sandbox loopback allow-list is closed. Allowed network activity is limited to TCP connections from renderer child processes to a loopback address (`127.0.0.1`, `::1`, or `localhost` resolving only to those addresses) on a port allocated by the render supervisor for the same render attempt. DNS resolution is allowed only for `localhost`; any other DNS query is a non-loopback outbound attempt. UDP is forbidden except operating-system loopback name-resolution mechanics required to resolve `localhost` to loopback. A connection to a loopback port not allocated for the same render attempt MUST fail with `failure_code='security_sandbox_violation'` and `reason_code='loopback_not_declared'`. A non-loopback DNS, TCP, UDP, HTTP, or HTTPS attempt MUST fail with `failure_code='remote_asset_reference'` and `reason_code='outbound_request_observed'`.

**REQ-RPT-119b**
`render_sandbox_observation.v1` MUST use Table 24-D. Sandbox observations are validation inputs, not release content. Unknown members are invalid. Observations retained in external bundles MUST NOT contain raw URLs, hostnames other than `localhost`, request paths, query strings, IP addresses other than loopback literals, headers, cookies, credentials, or response bodies.

**Table 24-D. `render_sandbox_observation.v1` schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `render_sandbox_observation.v1`. |
| `network_policy_id` | `identifier` | Yes | No | None | Must equal toolchain snapshot `network_policy_id`. |
| `stage` | string | Yes | No | None | §23.3 stage token. |
| `observation_kind` | string | Yes | No | None | `loopback_allowed`, `loopback_blocked`, `non_loopback_blocked`, `filesystem_blocked`, `env_blocked`, or `url_scheme_blocked`. |
| `destination_class` | string | Yes | Yes | None | `dns`, `tcp`, `udp`, `http`, `https`, `file`, `data`, `javascript`, `loopback`, or null when not network-related. |
| `allowed` | boolean | Yes | No | None | True only for `loopback_allowed`. |
| `attempt_count` | `finite_integer` | Yes | No | None | Count of equivalent observations after coalescing by stage, kind, and destination class. |
| `loopback_port_ref` | `identifier` | Yes | Yes | None | Non-null only for supervisor-allocated loopback ports; it is an opaque per-attempt ID, not the numeric port. |

If a render attempt performs browser rendering and emits no sandbox observation evidence, validation MUST fail with `failure_code='security_sandbox_violation'` and `reason_code='sandbox_observation_missing'`.

## 24.4 Render algorithms

**REQ-RPT-120**
`render_mermaid_bundle_v1` MUST execute these steps in order: validate release tuple and output options; validate composition input when the composition tuple is non-null; materialize export model; apply redaction and token manifests; resolve template-owned and composition-owned diagrams; apply valid composition label overrides and layout-mode validation; fail any retained manual-layout diagram with `failure_code='composition_invalid'` and `reason_code='manual_layout_not_supported_for_output_kind'`; serialize `.mmd` bytes for retained auto-layout diagrams under §16; render the admitted SVG output according to §7.5 and §22; validate rendered SVG security when SVG is emitted; validate sandbox evidence; build bundle manifest; run determinism validation for external release under REQ-RPT-122a; persist bundle only after validation passes.

**REQ-RPT-121**
`render_slidev_bundle_v1` MUST execute these steps in order: validate release tuple and output options; validate composition input when the composition tuple is non-null; materialize export model; apply redaction and token manifests; resolve template-owned and composition-owned diagrams; apply valid composition label overrides and layout-mode validation; for each retained diagram in `diagram_id` order, produce Mermaid source and rendered SVG through `mermaid_source_serialize_v3` when `layout_mode='auto'`, or produce rendered SVG through `diagram_layout_svg_serialize_v1` when `layout_mode='manual'`; generate the deck model using `derive_deck_v1` when the composition tuple is all-null and `derive_deck_v2` when it is non-null; embed diagram outputs into the deck model using local bundle paths only; serialize `slides.md` under §18; render the admitted PDF output according to §7.5 and §22; validate rendered SVG security for emitted diagram SVG files; validate click page count; validate sandbox evidence; build bundle manifest; run determinism validation for external release under REQ-RPT-122a; persist bundle only after validation passes.

**REQ-RPT-122**
For external release, the same tuple rendered twice in clean working directories under the same toolchain snapshot MUST produce byte-identical canonical export model, redaction manifest, token manifest, toolchain snapshot, validation summary, render-bundle manifest, and `output_sha256`. A mismatch MUST fail with `failure_code='nondeterministic_render'` and a reason code identifying the first mismatching artifact class.

External qualification MUST repeat pure materialization from the frozen
admission before repeating renderer execution. Comparing renderer outputs from
one reused materialization does not establish deterministic selection,
redaction, token allocation or section/diagram assembly. Compare canonical
model, redaction-manifest and token-manifest bytes and digests in that order,
using `export_model_mismatch`, `redaction_manifest_mismatch` or
`token_manifest_mismatch` respectively. This repeat remains bounded by the
redaction-stage and total-attempt deadlines and does not authorize live reads.

**REQ-RPT-122a**
Determinism validation MUST execute on every `external_release` render attempt by performing a second render in a clean working directory under the same toolchain snapshot and comparing artifacts under REQ-RPT-122. For `internal_draft` and `internal_review`, determinism validation MUST NOT run as a release-state-visible stage; an implementation MAY run it diagnostically only when diagnostic output is excluded from release-state and approval bytes.

**REQ-RPT-122b**
A rendered SVG file MUST pass `validate_rendered_svg_security_v1` before it is copied into a release bundle. The validator MUST parse the file as UTF-8 XML 1.0 with DTDs, external entities, parameter entities, and network entity resolution disabled. A parse error, decoding error, processing instruction, `DOCTYPE`, entity declaration, or root element other than `<svg>` in the SVG namespace fails with `failure_code='security_sandbox_violation'` and `reason_code='external_reference_in_rendered_output'`.

The rendered SVG safety subset is closed by Table 24-E. PDF and PPTX rendered-content scanning is future-only in this revision.

**Table 24-E. Rendered SVG safety subset**

| Category | V1 rule |
| --- | --- |
| Allowed elements | `svg`, `g`, `defs`, `linearGradient`, `radialGradient`, `stop`, `rect`, `circle`, `ellipse`, `line`, `polyline`, `polygon`, `path`, `text`, `tspan`, `title`, and `desc`. |
| Forbidden elements | `script`, `foreignObject`, `iframe`, `object`, `embed`, `image`, `video`, `audio`, `canvas`, `style`, `use`, `animate`, `animateMotion`, `animateTransform`, `set`, and any element outside the allowed set. |
| Attribute names | Attribute local names beginning with `on` after ASCII lowercase comparison are forbidden. Unknown namespace-qualified attributes are forbidden except `xml:space` on text-bearing elements. |
| URL-bearing attributes | `href`, `xlink:href`, `src`, `poster`, `formaction`, and equivalent URL-bearing attributes are forbidden. Fragment-only `href` values are future-only, not allowed in v1. |
| CSS and style | `style` elements, `style` attributes, `@import`, CSS custom properties, and CSS `url(...)` are forbidden. Presentation MUST use allowed presentation attributes only. |
| Presentation attributes | Numeric geometry, `viewBox`, `d`, `points`, `transform`, `fill`, `stroke`, `stroke-width`, `font-family`, `font-size`, `font-weight`, `text-anchor`, `dominant-baseline`, `opacity`, `fill-opacity`, and `stroke-opacity` are allowed only when their values contain no `url(`, control characters, external scheme, or parser-invalid numeric token. |
| Comments and metadata | XML comments are forbidden. `title` and `desc` text are allowed after `safe_string` validation and MUST NOT contain raw incident secrets beyond the already-redacted label text. |
| Failure mapping | Any violation in this table fails with `failure_code='security_sandbox_violation'` and `reason_code='external_reference_in_rendered_output'`. |

For this closed rendered-output profile, numeric geometry attributes are `x`,
`y`, `dx`, `dy`, `x1`, `y1`, `x2`, `y2`, `cx`, `cy`, `r`, `rx`, `ry`, `width`,
`height`, and `offset`. Scalar lengths are finite SVG numbers, optionally followed
by `px` or `%`; compound units are invalid. Radii, dimensions, stroke width and
font size MUST NOT be negative. Opacity values MUST be in `[0,1]`. `viewBox` has
four finite numbers with nonnegative dimensions; `points` has complete coordinate
pairs. Path data uses the SVG moveto, line, curve, arc and closepath grammar,
including single-character arc flags. Transform lists contain only `matrix`,
`translate`, `scale`, `rotate`, `skewX`, and `skewY` with their SVG argument counts.
Paint values use named colors, `none`, `currentColor`, or hexadecimal color
notation; the renderer MUST materialize computed colors into this subset before
validation. Font families are literal family names, not CSS expressions.
Namespace declarations may bind only the SVG namespace or the reserved XML
namespace. Duplicate attributes, multiple document roots, and undeclared
attributes are invalid. Inert label text is not interpreted as CSS, a URL or XML;
escaped metacharacters remain displayable text. Validation MUST reject invalid
bytes rather than repair them while constructing an approvable bundle.

# 25. Resource limit registry

**REQ-RPT-123**
Reporting resource limits MUST use Table 25-A. Except for the fixed diagnostic protocol `validation_issues.count`, the operative value for each Table 25-A limit key is the template-declared value when the template manifest declares one in `declared_limits`, otherwise the Default column. A template-declared value MUST be an integer or byte size in `[1, Hard limit]`; a value above the hard limit fails template validation before render output bytes with `error.code='invalid_release_request'` and `reason_code='template_limit_exceeds_hard_limit'`. `validation_issues.count` describes two distinct fixed protocol quantities under REQ-RPT-113a, not one configurable count. Templates MUST NOT declare that key; its presence fails with `reason_code='template_manifest_invalid'`. Deployments MUST NOT configure Reporting resource limits in v1.

**REQ-RPT-123a**
Resource-limit units and count points are closed. `MiB` means `1,048,576` octets. Byte-size limits count exact serialized bytes after UTF-8 encoding for text artifacts and exact file octets for binary artifacts; filesystem allocation size, compression ratio, and archive wrapper bytes are not inputs. Count limits are evaluated over post-redaction retained objects unless the row names bundle files or validation issues. Template-declared byte sizes MUST be expressed as either an integer octet count or an integer followed by one space and `MiB`; fractional units are invalid and fail with `reason_code='invalid_scalar_value'`.

**Table 25-A. Resource limits**

| Limit key | Default | Hard limit | Failure code | Reason code |
| --- | ---: | ---: | --- | --- |
| `export_model.canonical_bytes` | 64 MiB | 128 MiB | `export_model_resource_limit_exceeded` | `export_model_canonical_bytes_exceeded` |
| `sections.count` | 100 | 500 | `export_model_resource_limit_exceeded` | `sections_count_exceeded` |
| `blocks.count` | 10,000 | 50,000 | `export_model_resource_limit_exceeded` | `blocks_count_exceeded` |
| `fields.count` | 100,000 | 250,000 | `export_model_resource_limit_exceeded` | `fields_count_exceeded` |
| `records.count` | 50,000 | 100,000 | `export_model_resource_limit_exceeded` | `records_count_exceeded` |
| `relationships.count` | 100,000 | 250,000 | `export_model_resource_limit_exceeded` | `relationships_count_exceeded` |
| `diagrams.count` | 50 | 100 | `export_model_resource_limit_exceeded` | `diagrams_count_exceeded` |
| `subjects.count` | 25,000 | 50,000 | `token_manifest_invalid` | `subject_limit_exceeded` |
| `support_refs.count` | 100,000 | 250,000 | `support_ref_limit_exceeded` | `support_ref_limit_exceeded` |
| `composition.authored_title_chars` | 120 | 120 | `composition_invalid` | `authored_title_limit_exceeded` |
| `composition.authored_text_chars` | 2,000 | 5,000 | `composition_invalid` | `authored_text_limit_exceeded` |
| `slides.count` | 120 | 240 | `slidev_source_invalid` | `slide_count_exceeded` |
| `blocks_per_slide` | 30 | 60 | `slidev_source_invalid` | `slide_block_limit_exceeded` |
| `speaker_notes_chars_per_slide` | 10,000 | 25,000 | `slidev_source_invalid` | `speaker_notes_limit_exceeded` |
| `click_steps_per_slide` | 20 | 40 | `slidev_source_invalid` | `click_step_limit_exceeded` |
| `local_assets.count` | 250 | 500 | `asset_limit_exceeded` | `local_asset_count_exceeded` |
| `local_asset.byte_size` | 25 MiB | 50 MiB | `asset_limit_exceeded` | `local_asset_bytes_exceeded` |
| `local_assets.total_bytes` | 100 MiB | 250 MiB | `asset_limit_exceeded` | `local_assets_total_bytes_exceeded` |
| `rendered_pdf.byte_size` | 250 MiB | 500 MiB | `bundle_resource_limit_exceeded` | `rendered_pdf_bytes_exceeded` |
| `validation_issues.count` | 500 retained | 2,000 evaluated | `validation_summary_truncated` warning | `validation_issue_limit_reached`; retention and stop behavior per REQ-RPT-113a |

**REQ-RPT-124**
The implementation MUST evaluate every applicable resource limit before persisting external release output. When multiple limits fail, issue ordering in §23.4 determines `first_failure`.

# 26. Behavioral verification scenarios

**REQ-RPT-125**
The scenarios in Table 26-A describe observable behavior that a conforming
implementation MUST satisfy. They do not require a checked-in fixture corpus,
one test per row, stable test identities, or a scenario-to-test completeness
map. Tests MAY combine scenarios or decompose a scenario across layers when the
result exercises production code and proves the complete observable outcome.

**Table 26-A. Reporting behavior scenarios**

| Scenario label | Scenario | Required result |
| --- | --- | --- |
| `RPT-FIX-001` | Minimal external Mermaid release with no tokens. | Passes and emits `.mmd`, SVG, validation, manifests, and deterministic `output_sha256`. |
| `RPT-FIX-002` | Minimal external Slidev release with one auto-layout diagram. | Passes and emits `slides.md`, PDF, `.mmd`, SVG, validation, manifests, and deterministic `output_sha256`. |
| `RPT-FIX-003` | Request `markdown`, `html`, or `reenactment`. | Fails with `invalid_release_request`, `unsupported_output_kind`. |
| `RPT-FIX-004` | Same tuple rendered with different system clocks. | Canonical bytes and `output_sha256` identical. |
| `RPT-FIX-005` | Duplicate active party assignments. | Fails with `duplicate_active_party_assignment`. |
| `RPT-FIX-081` | A block asserts aggregate proof, including false/null/empty retired members. | Admission fails before output; no matching digest can authorize declassification. |
| `RPT-FIX-007` | Mixed paragraph requiring NLP split. | Does not split; redacts, drops, or fails. |
| `RPT-FIX-008` | Unstable unresolved subject without mention ID. | Fails with `token_subject_not_stable`. |
| `RPT-FIX-009` | Reveal map present in external bundle. | Bundle validation fails. |
| `RPT-FIX-010` | Stale Graph Projection snapshot. | Fails with `graph_projection_stale`. |
| `RPT-FIX-011` | Auto-layout Mermaid golden source. | `.mmd` bytes match exactly. |
| `RPT-FIX-012` | Slidev golden source. | `slides.md` bytes match exactly. |
| `RPT-FIX-013` | Invalid template binding pattern with `**` or array index. | Fails with `invalid_template_binding_pattern`. |
| `RPT-FIX-014` | Missing font manifest in toolchain snapshot. | Fails before external release. |
| `RPT-FIX-015` | Archive wrapper metadata changed, child files unchanged. | `output_sha256` unchanged. |
| `RPT-FIX-016` | Child file bytes changed. | `output_sha256` changes. |
| `RPT-FIX-017` | External `slidev` with `source_only=true`. | Fails with `source_only_external_release_invalid`. |
| `RPT-FIX-018` | Render timeout. | Release `render_failed`, reason `render_timeout`, no publishable bytes. |
| `RPT-FIX-019` | Non-loopback network attempt. | Fails with `remote_asset_reference`, `outbound_request_observed`, safe details only. |
| `RPT-FIX-020` | Slide with three click steps. | Exports four page states. |
| `RPT-FIX-021` | Timeline rows with parsed, partial, ambiguous, duplicate, and unparseable times. | Sorts exactly by §14.2 and preserves source text subject to redaction. |
| `RPT-FIX-022` | Bundle file with role/path/media mismatch. | Fails bundle validation. |
| `RPT-FIX-023` | Same tuple tokenized twice. | Identical `display_token` values assigned per family in `stable_subject_ref` order under `derive_display_token_v1`. |
| `RPT-FIX-024` | Section set mixing singleton and expanded template declarations. | `ordering_key` values and final section order match the `derive_section_ordering_key_v1` golden. |
| `RPT-FIX-025` | Generated text containing Markdown metacharacters inline and at line start. | `slides.md` bytes match the `slidev_markdown_escape_v2` golden exactly. |
| `RPT-FIX-026` | Mermaid label containing `<` or `>`. | Fails with `mermaid_invalid`, `invalid_mermaid_construct`. |
| `RPT-FIX-027` | Label of 121–240 scalars with and without template truncation, and a label over 240. | With truncation, exactly 120 scalars ending in U+2026; without, fails with `label_length_exceeded`; over 240 fails. |
| `RPT-FIX-028` | Export model with empty and populated nested validation objects. | `validation_summary`, `content_class_summary`, `section_validation`, and `display_times` serialize full fixed key sets; canonical bytes stable. |
| `RPT-FIX-029` | Timeline exceeding 240 rows with no template `selection_rule_id`. | Internal scope emits the overflow block; `external_release` fails with `timeline_overflow_unresolved`. |
| `RPT-FIX-030` | Internal `source_only=true` deck, including explicit rendered-output true and retained manual-layout diagram cases. | Explicit `pdf=true` and retained manual-layout diagrams fail with `source_only_conflict`; valid defaults suppress all rendered-output roles and emit a source-only bundle. |
| `RPT-FIX-031` | `snapshot_at='2026-02-30T00:00:00Z'`. | Fails with `invalid_timestamp_value`. |
| `RPT-FIX-032` | Font and package sets enumerated in shuffled filesystem order. | `content_manifest_digest_v1` digests are identical across orderings. |
| `RPT-FIX-033` | Export model with more than 100 diagrams. | Fails with `diagrams_count_exceeded`. |
| `RPT-FIX-034` | Snapshot with `deleted` and `superseded` records. | `deleted` never appears in rendered bytes; `superseded` is excluded from external release unless the template opts in. |
| `RPT-FIX-082` | A profile/view includes retired time bucketing or a template includes a retired aggregate allowlist. | Current schema admission rejects the member, including null or an empty array. |
| `RPT-FIX-036` | Two decks rendered from the same deck model. | Headmatter emits the full Table 18-C key set with fixed values; `slides.md` bytes match the golden. |
| `RPT-FIX-037` | Subject with refs `[party:A, party:B]`, recipient `[party:A]`, and profile default `drop`. | Element is absent, ordinals are recomputed, and redaction manifest records outcome `dropped`. |
| `RPT-FIX-038` | Same partition shape as `RPT-FIX-037`, but profile rule selects `allow`. | Fails with `redaction_manifest_invalid` and `disclosure_partition_unresolved`. |
| `RPT-FIX-039` | External recipient ref does not match the selected profile's allowed `party:*` set. | Fails before render output bytes with `recipient_partition_profile_mismatch`. |
| `RPT-FIX-040` | Deck derived from mixed singleton and expanded sections with chunking. | `derive_deck_v1` golden matches, including `slide_id`, per-slide `layout`, block distribution, and `expected_export_page_count`. |
| `RPT-FIX-041` | Section declaration with `click_profile='reveal_blocks'`. | Emits `N-1` click steps for `N` top-level blocks, fixed `resulting_state_hash` values, and `N` page states. |
| `RPT-FIX-042` | Redaction manifest with `allow`, tokenized, and partition-filter-dropped entries. | `cartulary.redaction_manifest.v5` bytes match the golden and contain no literal source or replacement values. |
| `RPT-FIX-043` | Export model with nested, pruned, and chunked structures. | Section, block, slide, diagram, and generated hash IDs match the ID goldens. |
| `RPT-FIX-044` | Valid and invalid support refs for source records, evidence items, artifacts, timeline events, relationships, and diagrams. | Valid support refs serialize as `cartulary.reporting_support_ref.v1`; refs containing raw evidence bytes, object-store keys, blob hashes, or multiple targets fail with `invalid_support_ref`. |
| `RPT-FIX-045` | Retained fields covering every Table 9-D1 source/redaction state combination and every Table 9-D2 source input state, plus invalid retained `dropped`, `blocked`, and deleted-source states. | Valid fields canonicalize with `raw_value_sha256=null`; deleted sources do not retain fields; retained `dropped` or `blocked` fields fail export-model validation. |
| `RPT-FIX-046` | Derivation profile with a valid v1 algorithm set and a profile naming an unknown algorithm. | Valid profile resolves deterministically; unknown algorithm fails with `unsupported_derivation_algorithm`. |
| `RPT-FIX-047` | Graph-derived diagram without a completed `source_projection_ref`, and timeline diagrams with no graph view. | Graph-derived diagram fails with `graph_projection_not_completed` or `graph_projection_selection_unresolved`; timeline diagrams do not require graph projection. |
| `RPT-FIX-048` | Selected Core redaction action cannot produce a retained value satisfying Table 9-D1. | Fails with `redaction_manifest_invalid`, `redaction_action_unresolved`, and no release bytes. |
| `RPT-FIX-049` | Internal draft omits an explicit redaction profile. | Core materializes `internal_passthrough`; redaction manifest contains `allow` outcomes and canonical bytes match explicit `internal_passthrough`. |
| `RPT-FIX-050` | Export model containing `source_record_ref.v1`, `relationship_endpoint_ref.v1`, `diagram_selection_rule.v1`, `diagram_overflow_summary.v1`, `safe_details.v1`, and `first_failure.v1` objects with extra members and invalid nulls. | Valid objects canonicalize with full required member sets; invalid objects fail with generic schema reason codes. |
| `RPT-FIX-051` | Diagram selections covering `explicit_refs`, `neighborhood`, `timeline_sequence`, and `all_with_bounds`, with duplicate and missing refs. | Valid selections match ordered goldens; duplicates fail with `diagram_selection_duplicate_ref`; missing refs fail with `diagram_selection_missing_ref`. |
| `RPT-FIX-084` | Same Slidev deck serialized for supported PDF and internal source-only variants; requests and template declarations also attempt PNG/PPTX support. | Supported variants preserve identical `slides.md` and `export.format: "pdf"`; unsupported requests and declarations fail before render execution. |
| `RPT-FIX-053` | Field values containing string, integer, boolean, null, empty array, scalar array, nested array, and object values. | Scalar and scalar-array values stringify by Table 18-E; nested arrays and objects fail with `invalid_field_value_for_text`. |
| `RPT-FIX-054` | Template manifest with per-kind output support, local assets, diagram declarations, and unsupported optional extension features. | Valid manifest passes; unsupported required feature fails; unsupported optional feature is omitted; invalid nested declarations fail with `template_manifest_invalid`. |
| `RPT-FIX-055` | Generic schema failures before and after render admission. | Missing members, unknown members, invalid nulls, invalid types, scalar violations, order drift, and duplicate array items map to Table 23-C2 reason codes. |
| `RPT-FIX-056` | Render timeout after partial renderer files are created. | Renderer process group is terminated then killed after 2 seconds if needed, partial release-output files are deleted, and no approvable bytes are exposed. |
| `RPT-FIX-057` | Atomic publish race with a missing staged file. | Release remains unpublished and fails with `bundle_manifest_invalid`, `atomic_publish_failed`. |
| `RPT-FIX-058` | Browser render with loopback access, undeclared loopback access, non-loopback DNS, and missing sandbox evidence. | Declared loopback is allowed; undeclared loopback fails with `loopback_not_declared`; non-loopback fails with `outbound_request_observed`; missing evidence fails with `sandbox_observation_missing`. |
| `RPT-FIX-059` | Tuple with all composition fields null for an existing Slidev golden. | Selects `derive_deck_v1`; deck model, `slides.md`, bundle manifest, and `output_sha256` match the pre-composition golden bytes exactly. |
| `RPT-FIX-060` | Composition tuple partial-null, absent/cross-incident resource, missing version, digest mismatch, and template mismatch cases. | Partial-null fails before admission with `composition_tuple_incomplete`; absent and cross-incident resources fail with `composition_not_found`; missing versions fail with `composition_version_not_found`; digest mismatch fails with `composition_digest_mismatch`; template mismatch fails with `composition_template_mismatch`. |
| `RPT-FIX-061` | Composition anchors covering resolved, unresolved, ambiguous, and `on_unresolved='drop'` external-release cases. | Resolved anchors apply the expected operation; unresolved and ambiguous anchors fail with exact composition reasons; external drop fails with `composition_drop_invalid_for_external_release`. |
| `RPT-FIX-062` | External authored text with profile default `allow_authored_presentation_text=false`, then with explicit `true`. | Default fails with `authored_text_not_permitted`; explicit permission admits the text only after partition filtering, redaction, and limit validation. |
| `RPT-FIX-063` | Authored subject placeholders for tokenized, display-name, unresolved, filtered, and malformed subjects. | Valid placeholders resolve through the existing token/display pipeline; invalid placeholders fail with `authored_subject_ref_unresolved`; no raw subject value appears in external bundle diagnostics. |
| `RPT-FIX-064` | Composition-owned diagrams selecting records outside template reachability. | `template_and_composition_reachable_records_v1` includes required records, subjects, support refs, and diagram refs; missing projection refs fail with existing graph adapter reasons. |
| `RPT-FIX-065` | Composition diagram label overrides for tokenized-subject and non-tokenized vertices. | Non-tokenized overrides serialize deterministically; tokenized-subject overrides fail with `diagram_label_override_invalid`. |
| `RPT-FIX-066` | Composition operation matrix covering every Table 17-D `op_kind`. | `derive_deck_v2` deck-model goldens match exactly, including operation order, chunking after overlay, click steps, structural IDs, and page counts. |
| `RPT-FIX-067` | Bundle containing a `manifest` file item, and the same bundle with `manifest.json` present only as the root object. | Listed manifest fails with `bundle_manifest_self_reference`; unlisted root manifest verifies and determines `output_sha256`. |
| `RPT-FIX-068` | Timeline rows tied on activity, precision, and date-entered fields but differing in `record_created_at` and `record_id`. | Rows sort by `record_created_at`, then `record_id`; missing `record_created_at` fails with `timeline_sort_key_missing`. |
| `RPT-FIX-069` | Template `declared_limits` timeout keys and local asset references, including invalid units, unknown timeout key, undeclared asset refs, remote asset refs, `font/ttf`, unsafe SVG, and HTML. | Valid integer-second timeouts and allowed media types pass; invalid timeout keys fail with `invalid_template_timeout_key`; undeclared assets fail with `undeclared_asset_ref`; remote refs fail with `remote_asset_reference_invalid`; media mismatches fail with `invalid_asset_media_type`. |
| `RPT-FIX-083` | Derived content contains restricted contributors and the selected rule would retain them outside the allowed recipient set. | The partition subset check fails; contributor count and category spelling never grant public eligibility. |
| `RPT-FIX-071` | Template and composition diagram declarations using `diagram_source_kind='template_static'`. | Both fail before render output bytes with `template_static_future_only`. |
| `RPT-FIX-072` | Authoritative preview using `cartulary.report_composition_preview_source.v1` for `internal_draft`, then attempting the same descriptor for `external_release`. | Internal preview passes through Reporting validation by `preview_source_sha256`; external release rejects the descriptor before approvable bytes exist. |
| `RPT-FIX-073` | Graph-derived diagrams with one matching tuple ref, no tuple ref, duplicate tuple refs for one `graph_view_id`, absent selected result, and digest mismatch. | Exact-one immutable tuple binding renders; no binding fails with `graph_projection_not_bound`; duplicate binding fails with `graph_projection_ambiguous`; absent and mismatched refs use their exact graph adapter reasons. |
| `RPT-FIX-074` | Mermaid output with three retained auto-layout diagrams and with zero retained diagrams after filtering. | Multi-diagram output emits one `.mmd` and required render per auto-layout diagram in `diagram_id` order; zero diagrams fail with `no_diagrams_selected`. |
| `RPT-FIX-075` | Draft preview source bytes and immutable-version preview bytes for the same composition resource. | Draft preview binds by `preview_source_sha256` over `cartulary.report_composition_preview_source.v1`; immutable preview may also expose `composition_sha256`; neither digest is accepted as external-release evidence unless the immutable tuple is bound. |
| `RPT-FIX-076` | `click_step.v1` with valid bounds and invalid zero, negative, exponent notation, decimal notation, `-0`, and above-limit ordinals or `at` values. | Valid values pass; every invalid scalar fails with `click_step_limit_exceeded` before rendered output persists. |
| `RPT-FIX-077` | Composition operation sequence with repeated excludes, repeated reorders, same-anchor inserts, later scalar overrides, and duplicate diagram-slide insertion. | Array-order effects match the golden; later scalar overrides win; duplicate insertion of the same composition-owned diagram fails with `composition_duplicate_diagram_insert`. |
| `RPT-FIX-078` | Slidev bundle containing one manual-layout composition flowchart and one auto-layout diagram. | Manual diagram emits deterministic `diagrams/{diagram_id}.svg` and no `.mmd`; auto diagram emits `.mmd`; `slides.md` references only local bundle paths. |
| `RPT-FIX-079` | Mermaid output request containing a retained manual-layout diagram. | Render fails before approvable bytes with `failure_code='composition_invalid'` and `reason_code='manual_layout_not_supported_for_output_kind'`. |
| `RPT-FIX-080` | Same manual-layout Slidev tuple and toolchain rendered twice in clean working directories. | Manual SVG bytes, bundle manifest, and `output_sha256` are byte-identical across both attempts. |

# 27. Behavioral acceptance criteria

## 27.1 Validation structure

**REQ-RPT-126**
Acceptance criteria summarize required outcomes for implementation and review.
They do not create a one-to-one requirement-to-test mapping and do not require
machine coverage counts. A current behavior without observable evidence MUST
be implemented and tested, or this NLSpec MUST be revised before that behavior
is removed or moved to a future profile.

## 27.2 Acceptance criteria

**REQ-RPT-127**
A conforming implementation MUST satisfy Table 27-A.

**Table 27-A. Acceptance criteria**

| Acceptance ID | Pass condition |
| --- | --- |
| `RPT-AC-AUTH-001` | A conformance claim fails when this NLSpec is marked adopted/current while any Table 1-A promotion condition remains unmet or future-only. |
| `RPT-AC-LINT-001` | Adopted text contains no open delegation phrase outside rationale, examples, source notes, future-only material, or an adjacent closed owner table. |
| `RPT-AC-CORE-001` | Every Core dependency row has owner, classification, affected Reporting requirements, and conformance consequence. |
| `RPT-AC-ID-001` | Every schema or algorithm identifier listed in §6.2 has a body section defining required fields, nullability, defaults, ordering, failure behavior, and behavioral validation. |
| `RPT-AC-KIND-001` | `markdown`, `html`, `reenactment`, unknown strings, aliases, and case variants are rejected with `unsupported_output_kind`. |
| `RPT-AC-OPT-001` | Omitted `output_options` materializes to exact §7.5 defaults and canonical bytes match explicit defaults. |
| `RPT-AC-OPT-002` | Every invalid option, scope, or template combination fails with the mapped structured error before output bytes persist. |
| `RPT-AC-SCHEMA-001` | Every top-level export-model array item validates against its named schema and rejects unknown members, invalid nulls, omitted required fields, duplicate IDs, and ordering drift. |
| `RPT-AC-SCHEMA-002` | Explicit defaults and omitted defaults canonicalize identically where omission is valid. |
| `RPT-AC-SOURCE-001` | Evidence provider fixtures prove exact set equality with the twelve-member Table 7-A0 allowlist; preserve null and timestamp values deterministically; exclude deleted and otherwise ineligible Evidence; omit every Table 7-A0a value class; reject malformed provider contributions; use only logical `evidence:{evidence_id}` support identity; and prove that an unknown Evidence field added to a source fixture does not appear in field facts, support metadata, diagnostics, or rendered output until this NLSpec adds it to the allowlist. |
| `RPT-AC-TIME-001` | Rerendering the same tuple with different system clocks produces byte-identical canonical export model, validation summary, redaction manifest, token manifest, toolchain snapshot, bundle manifest, and `output_sha256`. |
| `RPT-AC-MAT-001` | Export-model validation emits the expected structured failure for duplicate IDs, dangling refs, missing content class, invalid generated IDs, and resource-limit failures. |
| `RPT-AC-PART-001` | Duplicate active party assignments fail deterministically. |
| `RPT-AC-PART-006` | Derived summaries preserve contributor restrictions; aggregate proof flags, retired allowlists and bucketing settings reject, and no metadata-only declassification path exists. |
| `RPT-AC-PART-003` | Mixed-partition fields and child blocks retain their declared hierarchy after exact rule selection; removed content disappears from every projection, unions and ordinals are recomputed, and scalar/array values are never split or implicitly declassified. Diagram labels retain their typed owner boundary. |
| `RPT-AC-TOKEN-001` | Same tuple yields identical tokens and token manifests. |
| `RPT-AC-TOKEN-002` | Unstable unresolved subjects fail before bundle persistence. |
| `RPT-AC-REVEAL-001` | External bundles exclude reveal maps; internal sensitive artifacts contain complete reveal entries and require Core 04 authorization. |
| `RPT-AC-TIMEORDER-001` | Null, partial, ambiguous, duplicate, tied, and unparseable timeline values sort exactly by §14.2, including `record_created_at` tie-breaks, and preserve original display text subject to redaction. |
| `RPT-AC-GRAPH-001` | Stale projection fails with `graph_projection_stale`; matching completed projection renders. |
| `RPT-AC-GRAPH-002` | Each diagram selection rule produces golden ordered vertex and edge refs. |
| `RPT-AC-MMD-001` | Canonical `.mmd` bytes match golden fixtures exactly, including line endings and final newline. |
| `RPT-AC-MMD-002` | Unsupported Mermaid constructs fail before persistence with exact reason codes. |
| `RPT-AC-DECK-001` | Slide ordinals are contiguous, expected page count is deterministic, and slide resource-limit failures map to exact reason codes. |
| `RPT-AC-SLIDEV-001` | Canonical `slides.md` bytes match golden fixtures exactly. |
| `RPT-AC-CLICK-001` | Deck with `N` click steps exports exactly `N+1` page states per slide and fails with `click_export_page_count_mismatch` otherwise. |
| `RPT-AC-TOOLCHAIN-001` | Missing or changed font manifest, locale, timezone, Chromium digest, viewport, or render command changes validation result or bundle hash as specified. |
| `RPT-AC-TEMPLATE-001` | Template manifests reject unknown members, unsupported required features, invalid binding patterns, undeclared assets, and unsupported output options. |
| `RPT-AC-BUNDLE-001` | Every file role/path/media combination outside Table 22-B fails. |
| `RPT-AC-BUNDLE-002` | `manifest.json` is not listed in `files[]`; `bundle_manifest_sha256` is computed from the root manifest excluding only its own digest field and equals `output_sha256`. |
| `RPT-AC-ARCHIVE-001` | Mutating archive metadata alone does not alter `output_sha256`; mutating a child file does. |
| `RPT-AC-ERR-001` | Every validation failure emits exactly one public `error.code`, one `failure_code` when admitted, one `reason_code` when attributable, and safe details restricted to the allow-list. |
| `RPT-AC-LIFE-001` | Cancellation, timeout, partial output, exact replay, and new retry have deterministic job and release outcomes. |
| `RPT-AC-SANDBOX-001` | Non-loopback outbound attempt fails before persistence and safe details do not include raw external URLs. |
| `RPT-AC-LIMIT-001` | Every limit key enforces the default when the template is silent, the template-declared value when present, rejection above the hard limit, and exact failure mapping. |
| `RPT-AC-DERIVE-001` | `derivation_version` resolves to an adopted `cartulary.reporting_derivation_profile.v2`; an unresolved or unsupported profile fails closed with `unsupported_derivation_algorithm`, and every Table 7-A1 obligation is closed. |
| `RPT-AC-DERIVE-002` | `cartulary.reporting_derivation_profile.v2` accepts only Table 7-A2 algorithm tokens and fails unknown tokens with `unsupported_derivation_algorithm`. |
| `RPT-AC-DERIVE-003` | V1 derivation algorithms in Table 7-A3 select records, timelines, field keys, labels, support refs, subjects, field-subject mappings, and ordinals identically across implementations. |
| `RPT-AC-COMP-001` | Composition tuple all-null selects `derive_deck_v1` and preserves no-composition golden bytes; partial-null, digest mismatch, and template mismatch fail with exact composition reason codes. |
| `RPT-AC-COMP-002` | `template_and_composition_reachable_records_v1` includes composition-reachable records, relationships, subjects, diagram refs, authored placeholders, and support refs without reading live state. |
| `RPT-AC-COMP-003` | `section_anchor`, `record_anchor`, `block_anchor`, and `diagram_anchor` resolve semantically; zero, multiple, ordinal-path, and invalid external-drop targets fail with exact reasons. |
| `RPT-AC-COMP-004` | Authored text roles, LF rules, mandatory disclosure partition refs, profile permission, redaction participation, subject placeholders, and authored-text limits behave exactly as REQ-RPT-087f. |
| `RPT-AC-COMP-005` | Composition-owned diagrams use Table 15-C selection rules, completed projections, v2 reachability, deterministic label overrides, closed manual layout objects, and reject raw Mermaid, arbitrary nodes or edges, endpoint mutation, and tokenized-subject label overrides. |
| `RPT-AC-COMP-006` | `derive_deck_v2` applies Table 17-D operations in array order before chunking and click generation, while `derive_deck_v1` remains byte-identical for no-composition renders. |
| `RPT-AC-OPT-003` | Internal `source_only=true` forces every rendered-output member to `false`; an explicit rendered-output `true` or any retained manual-layout diagram fails with `source_only_conflict`. |
| `RPT-AC-ORDER-001` | `derive_section_ordering_key_v1` yields identical `ordering_key` values and section order for the same template manifest and content. |
| `RPT-AC-PART-004` | `deleted` records never appear in rendered bytes; `superseded` records are excluded from external release unless the template opts in; forced inclusion fails with `deleted_record_not_releasable`. |
| `RPT-AC-SCHEMA-003` | `validation_summary`, `content_class_summary`, `section_validation`, and `display_times` always serialize their full closed member sets and canonicalize deterministically. |
| `RPT-AC-TOKEN-003` | `derive_display_token_v1` produces grammar-valid, family-scoped tokens assigned in `stable_subject_ref` order, byte-identical across reruns, with no reversible material in the visible token. |
| `RPT-AC-TIMEORDER-002` | Timeline overflow emits the overflow block for internal scopes and fails external release with `timeline_overflow_unresolved` when no summarizing `selection_rule_id` is declared. |
| `RPT-AC-MMD-003` | Labels over the default truncate deterministically to 120 scalars ending in U+2026 only when declared, else fail with `label_length_exceeded`; labels containing `<` or `>` fail with `invalid_mermaid_construct`. |
| `RPT-AC-SLIDEV-002` | `slidev_markdown_escape_v2` escapes generated text to the closed Table 18-B set, code fences are limited to `text` and `mermaid`, asset refs serialize declared `bundle_path`, and headmatter emits the full Table 18-C key set with fixed values; `slides.md` bytes match golden. |
| `RPT-AC-TOOLCHAIN-002` | `content_manifest_digest_v1` produces identical `font_manifest_sha256` and `package_store_digest` regardless of file enumeration order and excludes mtime, ownership, and permission bits. |
| `RPT-AC-TIME-003` | Invalid calendar timestamps fail with `invalid_timestamp_value`; source timestamps are not rebucketed by Reporting to broaden disclosure. |
| `RPT-AC-PART-005` | Disclosure filtering uses the subset predicate, profile-rule-only resolution, post-filter ordinal recomputation, required-empty-section failure, and redaction-manifest removal entries. |
| `RPT-AC-REDACT-001` | `cartulary.redaction_manifest.v5` canonical bytes match the digest-bound artifact, include allow and removal entries, and exclude source and replacement values. |
| `RPT-AC-REDACT-002` | Core redaction actions map through Table 13-A1; unresolved actions fail with `redaction_action_unresolved` and never invent replacement values. |
| `RPT-AC-REDACT-003` | Redaction executes in every scope; omitted internal profiles materialize `internal_passthrough`; rule precedence, trace objects, token-backed parameters, and truncation bounds follow §13. |
| `RPT-AC-SUPPORT-001` | `cartulary.reporting_support_ref.v1` identifies exactly one support target, rejects raw evidence/storage/blob material, orders by `support_ref_id`, and fails invalid refs with `invalid_support_ref`. |
| `RPT-AC-FIELD-001` | Every retained field satisfies Table 9-D1 and every selected source value maps through Table 9-D2; external bundles contain no raw-value digests, deleted sources retain no fields, and retained `dropped` or `blocked` field states are invalid. |
| `RPT-AC-ID-002` | Generated IDs follow Table 10-B and structural IDs follow REQ-RPT-040b across nested, pruned, and chunked models. |
| `RPT-AC-DECK-002` | `derive_deck_v1` and `serialize_block_markdown_v1` produce golden deck models and `slides.md` bytes for every Table 18-D block kind, including list-depth failure. |
| `RPT-AC-DECK-003` | Deck title defaulting and validation follow REQ-RPT-087c, and `field_value_to_text_v1` stringifies every valid scalar and scalar array while rejecting objects and nested arrays. |
| `RPT-AC-MMD-004` | Mermaid node and participant IDs are ordinal, independent of source vertex-ref characters, and dangling edge endpoints fail. |
| `RPT-AC-GRAPH-003` | Graph-derived diagrams consume only exact leased `source_projection_ref.v2`; Reporting does not construct ad hoc projection input during materialization or render execution. |
| `RPT-AC-GRAPH-006` | Network Flow label components enter typed disclosure and redaction; internal endpoint and temporal labels use only deterministic post-redaction components or ordinals, external release remains fail-closed without a separately adopted allow rule, and lease admission uses result-before-declaration lock order. |
| `RPT-AC-GRAPH-004` | `diagram_selection_rule.v1`, overflow summaries, duplicate handling, missing-ref handling, traversal order, and label-source priority match §15 goldens. |
| `RPT-AC-GRAPH-005` | Graph-derived diagrams resolve `source_graph_view_id` through release tuple `graph_projection_refs[]` to exactly one completed digest-bound projection; no-match, duplicate-match, incomplete, stale, and digest-mismatch cases use exact reason codes. |
| `RPT-AC-ERR-002` | Every normative `failure_code` and `reason_code` literal appears in the Table 23-E or Table 23-F registry with the correct mapping. |
| `RPT-AC-TEMPLATE-002` | Section, layout, narrative-slot, render-profile, supported-output-option, and declared-limit schemas reject unknown members and invalid defaults. |
| `RPT-AC-TEMPLATE-005` | Template per-output-kind support, asset items, diagram declarations and optional extension-feature omission follow Tables 21-G through 21-I and REQ-RPT-103a; retired aggregate declarations reject under REQ-RPT-103c. |
| `RPT-AC-TEMPLATE-004` | Template `declared_limits` accepts only Table 24-B timeout keys as integer seconds and Table 25-A resource-limit keys with their declared units. |
| `RPT-AC-VALID-001` | Validation summaries serialize `safe_details`, Table 23-B1 issue items, retained issue limits, evaluated issue counts, and truncation warnings deterministically. |
| `RPT-AC-VALID-002` | Generic schema failures map to Table 23-C2 reason codes and `first_failure.v1` copies the first ordered issue exactly. |
| `RPT-AC-SANDBOX-002` | Rendered SVG files outside the Table 24-E safety subset, including scripts, event attributes, foreign objects, disallowed elements, style blocks or attributes, external hrefs, `data:` URLs, and external `url()` references, fail with `external_reference_in_rendered_output`. |
| `RPT-AC-SANDBOX-003` | Sandbox observations use `render_sandbox_observation.v1`; declared loopback is allowed, undeclared loopback and non-loopback attempts fail with exact reason codes, and raw URLs are not retained. |
| `RPT-AC-LIFE-002` | Timeout and cancellation terminate renderer process groups, wait 2 seconds before kill, delete partial release-output files, retain only safe validation summaries, and publish bundles atomically. |
| `RPT-AC-SLIDEV-004` | `slides.md` headmatter always emits `export.format: "pdf"`; source bytes remain stable across supported PDF and internal source-only variants. Unsupported rendered PNG/PPTX requests and template capabilities reject before render execution. |
| `RPT-AC-MMD-005` | Mermaid bundles emit one `.mmd` and required render per retained auto-layout diagram in `diagram_id` order; zero retained diagrams fail with `no_diagrams_selected`. |
| `RPT-AC-DIAGRAM-LAYOUT-001` | Manual layout succeeds only through render paths that honor exact positions, emits deterministic safe SVG, emits no `.mmd`, and fails for `output_kind='mermaid'` with `manual_layout_not_supported_for_output_kind`. |
| `RPT-AC-PREVIEW-001` | `cartulary.report_composition_preview_source.v1` is accepted only for `internal_draft`, binds by `preview_source_sha256`, and never becomes approval or external-release evidence. |
| `RPT-AC-CLICK-002` | `click_step.v1.ordinal` and `click_step.v1.at` use `positive_integer`, reject non-integer JSON number forms and out-of-bound values, and remain contiguous `1..N`. |
| `RPT-AC-EXTENSIONS-001` | Reporting invokes only the admitted `snapshot_reporting.render_export_v1` specialization for profiles selecting participant mode, never invokes `no_participation`, and retains sole ownership of result admission and rendering. |
| `RPT-AC-EXTENSIONS-002` | Reporting participant schemas and procedural validators declare every reachable condition, emit no unregistered condition, and apply the shared validation-result precedence and exact finding-count boundaries. |
| `RPT-AC-BEHAVIOR-001` | Current behavior has production-facing evidence or a prior specification disposition; no status row, identifier count, or static corpus is accepted as behavioral evidence. |

# 28. Future-only areas

**REQ-RPT-128**
The areas in Table 28-A are future-only. A conforming v1 implementation MUST use the current omission behavior in the table.

**Table 28-A. Future-only areas**

| Future-only area | Current omission behavior |
| --- | --- |
| Markdown report output | `output_kind='markdown'` fails with `unsupported_output_kind`. |
| Static HTML report output | `output_kind='html'` fails with `unsupported_output_kind`. |
| Rendered PNG and PPTX output | Request options `png=true` or `pptx=true` fail with `unsupported_output_option`; template support declarations must be false. The explicit false capability members and reserved role/path vocabulary do not authorize execution. A future renderer profile must adopt its execution, validation and determinism contract before enabling either capability. PNG/JPEG/SVG template image assets remain current and are independent of rendered-output capability. |
| Reenactment output kind | `output_kind='reenactment'` fails with `unsupported_output_kind`. |
| Graph chunking | Diagrams beyond hard Mermaid bounds fail; templates must select smaller diagrams. |
| Template-static semantic diagrams | `diagram_source_kind='template_static'` fails with `template_static_future_only`; static art may be included only as a sanitized local asset. |
| Source-only external release | Fails with `source_only_external_release_invalid`. |
| Arbitrary Slidev parser or pre-parser extensions | Invalid for external release. |
| Custom Slidev components beyond reveal-only subset | Invalid for external release. |
| Canonical archive bytes | Approval and publication bind to manifest hash, not ZIP/TAR bytes. |
| External toolchain auto-update or remote package resolution | Invalid during render. |
| Deployment-configurable Reporting resource limits | Templates MAY declare values within hard limits; deployments MUST NOT configure Reporting limits in v1. |
| External reveal-map sharing | Reveal maps remain internal sensitive release artifacts. |
| New `post` record family | Invalid unless a later Core owner defines it. |
| PDF/PPTX rendered-content scanning | No current v1 scanner claim; SVG content checks remain mandatory under REQ-RPT-122b. |
| Standalone manual-diagram export | Manual layout is current only through `output_kind='slidev'`; standalone manual-diagram export requires a later output-kind amendment. |
| Raw Markdown, Mermaid, HTML, or rendered-byte composition editing | Invalid composition input; Reporting consumes only composition-as-data through the companion schema. |
| Post-redaction composition editing | Invalid; composition is applied before chunking, redaction-sensitive serialization, and render. |
| Composition-driven workbook, snapshot, or projection mutation | Invalid; source changes must occur through workbook owners and new immutable snapshots. |
| Cross-template composition portability | Not current; compositions bind to exact template version and migration requires a later owner. |
| Collaborative composition editing | Not current Reporting behavior; any builder collaboration belongs to a later authoring-surface owner. |
| Free-text sensitive-value scanning of authored presentation text | Not current; v1 relies on profile permission, disclosure partitions, subject placeholders, redaction gates, and non-normative builder linting. |

# 29. Revision completion checklist

**REQ-RPT-129**
A document revision that claims to close this draft MUST satisfy Table 29-A.

**Table 29-A. Binary completion checklist**

| Check | Pass condition |
| --- | --- |
| Current output scope | Only `slidev` and `mermaid` appear in the current accepted `output_kind` table. |
| Future-only migration | `markdown`, `html`, and `reenactment` appear only in future-only, migration, or rationale material. |
| Core alignment | Core route vocabulary, release resource shape, release hash, party-assignment source state, reveal-map security, and sandbox requirements are adopted or explicitly marked as blockers. |
| Schema closure | Every schema identifier in §6.2 has a complete field table or imported owner table with requiredness, nullability, defaults, ordering, unknown-member behavior, and failure mapping. |
| Timestamp determinism | No hash-participating Reporting-owned generated timestamp is wall-clock-derived. |
| Error consistency | No public payload or accepted normative requirement combines the Core public error family and Reporting failure code into one colon-delimited machine-readable identifier. |
| Redaction interface | Token substitution, token manifest, redaction manifest, and reveal map have closed schemas and safe omission rules. |
| Internal redaction default | Internal scopes run the redaction pipeline, and omitted internal profiles materialize `internal_passthrough`. |
| Redaction rule closure | Rule precedence, selectors, token-backed parameters, truncation bounds, and selected-rule trace objects are closed. |
| Partition closure | Assignment lifecycle, duplicate active assignments, contributor restriction retention, and retained-structure redaction are deterministic. |
| Timeline closure | Sort materialization, precision ranks, `record_created_at` tie-breaks, unresolved-time ordering, bounds, and overflow summary are specified. |
| Graph closure | Reporting consumes only exact immutable Graph Projection output through release tuple `graph_projection_refs[]`, `source_projection_ref.v2`, leases, typed source-owner dispatch, and closed selection rules. |
| Diagram selection closure | Selection-rule schemas, traversal order, duplicate handling, missing-ref handling, overflow summaries, and label-source priority are closed. |
| Mermaid closure | `.mmd` bytes are deterministically serializable from auto-layout diagram objects; mermaid bundles emit one source and required render per retained auto-layout diagram, manual layout fails closed for Mermaid output, and zero retained diagrams fail closed. |
| Manual layout closure | Manual-layout diagrams consume only closed composition layout objects, emit deterministic safe SVG through `diagram_layout_svg_serialize_v1`, never emit `.mmd`, and cannot mutate selected graph or timeline semantics. |
| Slidev closure | `slides.md` bytes are deterministically serializable from deck objects. |
| Slidev option stability | `slides.md` headmatter uses `export.format: "pdf"`; supported PDF and internal source-only variants preserve identical source bytes, while unsupported PPTX/PNG requests and capabilities reject before execution. |
| Field stringification closure | Field-value scalar and array stringification is closed, and invalid value shapes fail with mapped reasons. |
| Click closure | Click-step schema and page-count rule are closed. |
| Toolchain closure | Render environment fields are sufficient to reproduce rendered outputs inside the declared environment. |
| Bundle closure | Role/path/media matrix is exhaustive; `manifest.json` is not self-listed; archive wrapper semantics are explicit. |
| Sandbox closure | Network, filesystem, environment, URL scheme, and evidence-mount boundaries are explicit and testable. |
| Lifecycle closure | Success, failure, cancellation, timeout, partial output, retry, and Core release-state effects are explicit. |
| Limit closure | Every resource limit and timeout has a declared key, unit, default, hard limit, stage, failure code, and overflow or fail behavior. |
| Token-string closure | `display_token` is deterministically and non-reversibly derived by `derive_display_token_v1`. |
| Ordering-key closure | Section `ordering_key` is deterministically derived by `derive_section_ordering_key_v1`. |
| Nested-object closure | Export-model `validation_summary`, `content_class_summary`, `section_validation`, and `display_times` are closed with fixed member sets. |
| Escaping closure | Mermaid and Slidev generated text have closed escape sets, deterministic label truncation, and no unnamed raw-HTML detector. |
| Digest closure | Multi-file digests use `content_manifest_digest_v1` and are enumeration-order-independent and filesystem-metadata-free. |
| Derivation-profile closure | `derivation_version` resolves to an adopted derivation profile that deterministically closes snapshot-to-export-model content derivation. |
| Composition tuple closure | Nullable composition tuple fields are all-null or all non-null; all-null preserves no-composition golden bytes, and non-null tuple validation has exact failure codes. |
| Composition consumer-boundary closure | Reporting imports the companion composition identifiers without owning authoring routes, lifecycle, builder UI, or full wire schema; draft previews bind by `cartulary.report_composition_preview_source.v1` and `preview_source_sha256`, not release `composition_sha256`. |
| Composition reachability closure | `template_and_composition_reachable_records_v1` includes composition-reachable records, subjects, diagrams, placeholders, and support refs without live-state reads. |
| Composition anchor closure | Semantic anchors are the only valid operation targets; zero, multiple, ordinal-path, and external-drop cases fail deterministically. |
| Authored-presentation-text closure | Text roles, profile permission, partition labels, LF rules, subject placeholders, limits, and residual free-text-scanning boundary are explicit. |
| Composition diagram closure | Composition diagrams use existing selection rules, completed projections, and optional closed layout data; raw Mermaid, arbitrary nodes or edges, endpoint mutation, and tokenized-subject label overrides are invalid. |
| Deck v2 closure | `derive_deck_v2` applies composition operations before chunking and click generation while preserving `derive_deck_v1` for all-null composition; operation conflicts and duplicate diagram insertion fail deterministically. |
| Behavioral disposition | Every current behavior is retained and routed, implemented and tested, or revised before being pruned or moved to a future profile. |
| Filtering algorithm | `filter_disclosure_partitions_v1` exists with effective-set construction, subset predicate, profile-rule-only resolution, and fail-closed unresolved-disclosure behavior. |
| Recipient validation | `external_release` recipient partitions are validated against snapshot Parties and the selected redaction profile's allowed `party:*` set. |
| Redaction-manifest bytes | `cartulary.redaction_manifest.v5` is defined, digest-bound, safe, and acceptance-covered. |
| Generated-ID registry | Every Reporting-generated hash ID appears in Table 10-B, and every structural ID has an ordinal derivation rule. |
| Deck derivation | `derive_deck_v1`, `serialize_block_markdown_v1`, click-profile generation, and per-slide layout frontmatter are closed. |
| Template nested schemas | Section, layout, narrative-slot, render-profile, output-support, and limit declarations have requiredness, defaults, and failure behavior. |
| Template declarations | Asset items, diagram declarations, optional extension features and per-output-kind option support have closed schemas and fail/omit behavior; retired aggregate declarations reject. |
| Registry closure | Every normative `failure_code` and `reason_code` literal appears in Table 23-E or Table 23-F. |
| Rendered-output security | Rendered SVG self-containment checks are mandatory, and PDF/PPTX scanning remains future-only. |
| Validation-summary closure | `safe_details.v1`, `first_failure.v1`, issue ordering, and generic schema failure mapping are closed. |
| Sandbox observation closure | Loopback allow-list behavior and `render_sandbox_observation.v1` are closed and safe for external bundles. |
| Atomic publish closure | Timeout/cancel cleanup, process termination, partial-output deletion, and atomic bundle visibility are explicit. |

# 30. Non-normative source notes

This section is non-normative. It records supporting inputs used to shape this revision. It does not create requirements, defaults, exceptions, or conformance evidence beyond the normative sections above.

**Table 30-A. Supporting research notes**

| Source | Material use in this NLSpec revision | Boundary |
| --- | --- | --- |
| `docs/research/R01-aurora_incident_response_report.md` | Supported the decision to avoid live report views, renderer-local HTML handling, and browser-date parsing as release-byte authorities. | Normative closure remains in §§10, 16, 18, 20, 22, and 24. |
| `docs/research/R02-cartulary_crm_tem_dfir_research_report.md` | Supported keeping briefings, training cadence, and playbook guidance out of the Reporting artifact contract unless represented as snapshot artifacts. | Product-versus-playbook separation remains a Core/domain boundary, not a Reporting workflow addition. |
| `docs/research/R03-Kanvas_technical_research_report.md` | Supported self-contained export bundles, no remote/CDN asset dependency, closed schema names, and avoiding separate report implementations that silently omit fields or drift from canonical case state. | Normative closure remains in the export model, redaction, template asset, bundle, and sandbox sections. |
| `docs/research/R06-spreadsheet_of_doom_dfir_research_report.md` | Supported first-class timeline, evidence-reference, support-reference, custody-state, and raw-evidence-outside-report boundaries. | Reporting support refs point to metadata and redacted summaries; raw evidence bytes remain outside release bundles. |
| `docs/research/R07-spreadsheet-of-doom-sod-report.cr.md` | Supported keeping source-of-truth findings and case facts in structured source records or narrative slots rather than treating presentation composition as an evidence source. | Authored presentation text remains governed by role, partition, placeholder, and redaction rules; review guidance is non-normative. |
| `docs/research/R04-responsive_browser_spreadsheet_ui_research_memo.md`, `docs/research/R05-responsive-interface-design-report.cr.md`, `docs/research/R08-handsontable-react-research-report.md`, and `docs/research/R09-react-data-grid-research-report.md` | Reviewed as adjacent workbook/front-end implementation research. They did not materially change the Reporting artifact contract. | They remain implementation-support inputs, not Reporting conformance sources. |

**Table 30-B. External source freshness notes**

| Source checked as of 2026-07-05 | Material use in this NLSpec revision | Boundary |
| --- | --- | --- |
| RFC 8785 JSON Canonicalization Scheme, `https://datatracker.ietf.org/doc/html/rfc8785` | Informed deterministic JSON serialization review and the decision to keep Cartulary's existing canonical JSON owner explicit. | The normative algorithm is `reporting_canonical_json_v1`; this NLSpec does not claim RFC 8785 compliance unless a later owner adopts its exact rules. |
| Mermaid configuration documentation, `https://mermaid.ai/open-source/config/schema-docs/config-properties-securitylevel.html` | Supported pinning strict Mermaid rendering behavior and rejecting HTML/click constructs rather than relying on renderer defaults. | The normative rule is Table 16-B plus the pinned toolchain snapshot, not upstream defaults. |
| Slidev export and CLI documentation, `https://sli.dev/guide/exporting` and `https://sli.dev/builtin/cli` | Supported explicit export formats, Playwright/Chromium binding, and click-page export checks. | The normative rule is the Slidev subset, command snapshot, and page-count validation in §§18-20. |
| W3C SVG Integration / SVG processing-mode material, `https://www.w3.org/TR/svg-integration/` | Supported requiring rendered SVG self-containment and script/external-reference rejection. | The normative rule is REQ-RPT-122b; no broader SVG viewer conformance claim is made. |
| OWASP File Upload Cheat Sheet and Web Security Testing Guide archive guidance, `https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html` and `https://github.com/OWASP/www-project-web-security-testing-guide/blob/master/latest/4-Web_Application_Security_Testing/10-Business_Logic_Testing/09-Test_Upload_of_Malicious_Files.md` | Supported rejecting archive traversal, symlinks, special files, and unsafe archive metadata as release-byte authorities. | The normative rule is REQ-RPT-106 and the bundle path constraints in §22. |
