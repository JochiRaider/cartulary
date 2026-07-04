---
title: Cartulary Reporting Subsystem NLSpec
status: draft/proposed
document_class: nlspec
profile: snapshot_reporting
schema_id: cartulary.reporting_subsystem_nlspec.v1
---

# 1. Status, scope, and authority

Status: `draft/proposed`.

This NLSpec defines the Cartulary Reporting Subsystem for the Snapshot and Reporting Extension Profile. It becomes implementation-conformance authority only after promotion to `status: adopted/current` and after every Core companion amendment in §5 is adopted or the affected requirement is explicitly reclassified as future-only.

**REQ-RPT-001**
This NLSpec governs only reporting-subsystem behavior inside the Snapshot and Reporting Extension Profile:

- canonical reporting export-model derivation;
- report, deck, and diagram intermediate models;
- party-scoped disclosure partitioning after snapshot materialization;
- redaction, token substitution, token manifests, and reveal-map production;
- Mermaid source generation, validation, and local rendering;
- Slidev source generation, validation, rendering, and export;
- reveal-only presentation click steps;
- render validation;
- deterministic render-bundle packaging and hashing;
- reporting-subsystem conformance fixtures and acceptance criteria.

**REQ-RPT-002**
This NLSpec MUST NOT redefine Core 00 through Core 04 behavior outside the reporting subsystem. It MUST NOT redefine live workbook authorization, workbook-surface identity, source-record mutation semantics, graph-projection semantics, public snapshot or release route admission, incident membership, evidence access authorization, release approvals, deployment administration, WebSocket behavior, or Core 05 claim-publication behavior.

**REQ-RPT-003**
When this NLSpec conflicts with Core 00 through Core 04 outside the reporting subsystem, the conflict is a defect in this NLSpec. When this NLSpec conflicts with non-normative appendices, research reports, design guides, operating guidance, implementation guides, repository handoff material, or examples, this NLSpec governs only the reporting subsystem after adoption.

**REQ-RPT-004**
This NLSpec MUST specialize the existing Snapshot and Reporting route family rather than create a second reporting route family. Public snapshot creation, release creation, release-resource state, release approval, release invalidation, public route admission, and public route-family error envelopes remain owned by Core 01 and Core 04. The Reporting Subsystem owns render-attempt behavior after Core-owned route admission accepts a release render operation.

**REQ-RPT-005**
This NLSpec adds no deployment-configuration key in this revision. A later revision that adds a deployment-configuration key MUST define its key name, type, default, allowed values, fail-closed validation behavior, secret classification, backup or portability behavior, and owner interaction with Core 04 deployment configuration.

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
| Reporting derivation profile | `cartulary.reporting_derivation_profile.v1` is adopted and every `derivation_version` resolves to it under REQ-RPT-027a. |
| Reporting acceptance matrix | Every `REQ-RPT-*` in this NLSpec maps to at least one `RPT-AC-*` or fixture. |

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
| How are identities, hosts, indicators, evidence, parties, and related entities represented in diagrams? | Completed Graph Projection consumption, diagram objects, Mermaid source generation, source references, overflow reporting, and validation. |
| Which values are visible, tokenized, stubbed, masked, dropped, or blocked for a recipient? | Disclosure partitions, redaction profiles, token manifests, redaction manifests, and reveal-map handling. |
| How are Slidev decks generated reproducibly? | `cartulary.reporting_slide_deck.v1`, Slidev subset validation, reveal-only click-step profile, toolchain snapshot, and deterministic render bundle. |
| What bytes are approved and published? | Per-file SHA-256 entries, canonical render-bundle manifest, root `output_sha256`, release binding, and rerender determinism. |

**REQ-RPT-015**
The subsystem MUST preserve the evidence-versus-presentation boundary. Generated reports, decks, diagrams, and summaries MAY rearrange, select, redact, visualize, and summarize snapshot facts according to this NLSpec. They MUST NOT invent facts, actors, timestamps, commands, causal chains, source evidence, party relationships, or conclusions not present in the reporting export model.

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
| Core 00 | Adopted-subsystem registry entry and precedence. | `blocked-until-core-adoption` until adopted | REQ-RPT-001..REQ-RPT-006 |
| Core 01 | Snapshot and release route admission; public release resource state; public route error envelope. | `core-owned-current` for existing route family; companion edit required for new tuple fields | REQ-RPT-022..REQ-RPT-035, REQ-RPT-109..REQ-RPT-115 |
| Core 01 | `render_admitted_at` binding, `output_options` if public, and `output_sha256=bundle_manifest_sha256`. | `blocked-until-core-adoption` | REQ-RPT-027..REQ-RPT-035, REQ-RPT-047..REQ-RPT-052, REQ-RPT-097..REQ-RPT-099, REQ-RPT-105..REQ-RPT-108 |
| Core 01 | Token-backed parameters for `mask` and `stub` redaction rules. | `blocked-until-core-adoption` | REQ-RPT-091..REQ-RPT-105 |
| Core 02 | `record_party_assignment.v1` source-state relation for Host and Identity records. | `blocked-until-core-adoption` | REQ-RPT-075..REQ-RPT-090 |
| Core 03 | Workbook or inspector editing for Host and Identity party assignments. | `blocked-until-core-adoption` | REQ-RPT-075..REQ-RPT-081 |
| Core 04 | Live authorization remains incident-role based; export partitions do not affect live workspace access. | `core-owned-current` | REQ-RPT-016, REQ-RPT-075..REQ-RPT-090 |
| Core 04 | Reveal-map sensitive-release-artifact authorization and retention. | `blocked-until-core-adoption` | REQ-RPT-100..REQ-RPT-105 |
| Core 04 | Render sandbox trust boundary. | `blocked-until-core-adoption` | REQ-RPT-119..REQ-RPT-122 |
| Graph Projection NLSpec | Projection input, output, lifecycle, validation, identity, and consumer behavior. | `adopted-subsystem-current` | REQ-RPT-076..REQ-RPT-080 |
| Reporting derivation profile | Adoption of `cartulary.reporting_derivation_profile.v1` as the versioned owner of snapshot-to-export-model content derivation referenced by `derivation_version`. | `blocked-until-core-adoption` | REQ-RPT-027a |

## 5.2 Required Core companion edits

**REQ-RPT-019**
The companion edits in Table 5-C MUST be made outside this NLSpec before adoption. This NLSpec MUST NOT restate those edits as already adopted behavior unless the corresponding Core document is revised.

**Table 5-C. Required Core companion edits**

| Core document | Required edit | Reporting dependency closed |
| --- | --- | --- |
| Core 00 | Add this NLSpec to the adopted subsystem registry after promotion. | Establishes adopted precedence. |
| Core 01 | Align Reporting v1 route output vocabulary to `slidev` and `mermaid`, or classify other selectors as non-conformant future-only selectors. | Closes output-kind ambiguity. |
| Core 01 | Add `render_admitted_at` to the render tuple or define it as a deterministic alias of release admission timestamp. | Closes hash-participating timestamp determinism. |
| Core 01 | State that Reporting `output_sha256` equals canonical `render_bundle_manifest.v1` SHA-256 for multi-file bundles. | Closes bundle approval hash semantics. |
| Core 01 | Add public `output_options` when the route exposes option selection. | Closes PDF, SVG, PNG, PPTX, and source-only omission semantics. |
| Core 02 | Add `record_party_assignment.v1` as source state for Host and Identity subjects. | Closes recipient partition derivation. |
| Core 03 | Add source-state editing through ordinary workbook fields or inspector feature groups. | Preserves the hot-path boundary. |
| Core 04 | Add reveal-map sensitive-artifact authorization. Default: incident `admin` unless a later Core capability narrows it. | Protects reversible token material. |
| Core 04 | Add render sandbox trust boundary and non-egress rule. | Closes render security. |
| Reporting derivation profile owner | Adopt `cartulary.reporting_derivation_profile.v1` as the versioned owner of snapshot-to-export-model content derivation, resolvable from `derivation_version`, with deterministic record and timeline selection, `field_key` assignment, `display_label` derivation, section expansion, and support-reference selection. | Closes content-derivation determinism above the export-model schema boundary. |
| Core 04 | Add a redaction-profile control that MAY include `superseded` records in an external release. Default: excluded. | Closes the superseded-record external disclosure default in REQ-RPT-043a. |
| Core 04 | Add a redaction-profile `neutral_token_family` control. Default: `false`. | Closes optional subject-class suppression for display tokens in REQ-RPT-063a. |

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
| `Mermaid source` | Validated `.mmd` text generated by this subsystem from redacted export-model and graph-projection data. |
| `Slidev source` | Generated `slides.md` conforming to `slidev_markdown_serialize_v1`. |
| `click step` | One deterministic reveal or hide step under `cartulary.slidev_reveal_only.v1`. |
| `render attempt` | One admitted Reporting-owned execution after Core route admission. |

## 6.2 Closed schema and algorithm identifiers

**REQ-RPT-021**
This revision defines exactly the identifiers in Table 6-B. Every identifier in the table MUST have a normative owner section and acceptance coverage.

**Table 6-B. Closed identifiers**

| Identifier | Kind | Owner section |
| --- | --- | --- |
| `cartulary.reporting_render_request_options.v1` | schema | §7 |
| `cartulary.reporting_export_model.v1` | schema | §9 |
| `cartulary.reporting_section.v1` | schema | §9 |
| `cartulary.reporting_block.v1` | schema | §9 |
| `cartulary.reporting_field.v1` | schema | §9 |
| `cartulary.reporting_record_summary.v1` | schema | §9 |
| `cartulary.reporting_relationship_summary.v1` | schema | §9 |
| `cartulary.reporting_timeline_event.v1` | schema | §14 |
| `cartulary.tokenizable_subject.v1` | schema | §13 |
| `stable_subject_ref_v1` | identifier grammar | §13 |
| `cartulary.reporting_token_manifest.v1` | schema | §13 |
| `cartulary.reporting_token_reveal_map.v1` | schema | §13 |
| `cartulary.reporting_diagram.v1` | schema | §15 |
| `source_projection_ref.v1` | schema | §15 |
| `mermaid_source_serialize_v1` | serializer | §16 |
| `cartulary.reporting_slide_deck.v1` | schema | §17 |
| `slidev_markdown_serialize_v1` | serializer | §18 |
| `click_step.v1` | schema | §19 |
| `cartulary.reporting_toolchain_snapshot.v1` | schema | §20 |
| `cartulary.reporting_template_pack_manifest.v1` | schema | §21 |
| `cartulary.render_bundle_manifest.v1` | schema | §22 |
| `cartulary.reporting_render_validation_summary.v1` | schema | §23 |
| `cartulary.reporting_export_model_validation.v1` | schema | §9 |
| `reporting_canonical_json_v1` | canonicalization algorithm | §10 |
| `content_manifest_digest_v1` | digest algorithm | §10 |
| `materialize_reporting_export_model_v1` | derivation algorithm | §11 |
| `assign_party_disclosure_partitions_v1` | derivation algorithm | §12 |
| `aggregate_public_v1` | derivation algorithm | §12 |
| `split_mixed_block_v1` | derivation algorithm | §12 |
| `apply_token_backed_redaction_v1` | redaction algorithm | §13 |
| `select_timeline_rows_v1` | derivation algorithm | §14 |
| `derive_section_ordering_key_v1` | derivation algorithm | §9 |
| `derive_display_token_v1` | derivation algorithm | §13 |
| `truncate_label_v1` | serialization algorithm | §16 |
| `slidev_markdown_escape_v1` | serialization algorithm | §18 |
| `render_slidev_bundle_v1` | render algorithm | §24 |
| `render_mermaid_bundle_v1` | render algorithm | §24 |
| `render_sandbox_policy_v1` | security policy | §24 |
| `render_attempt_lifecycle_v1` | lifecycle | §24 |
| `validate_release_render_v1` | validation algorithm | §23 |

## 6.3 `post` source-family mapping

**REQ-RPT-022**
`post` is not a current Cartulary source-record family in this NLSpec. A conforming reporting export model MUST NOT emit `post_ref`, `post_id`, `record_type='post'`, or `source_record_family='post'` unless a later Core owner defines that source record family and this NLSpec is revised.

**REQ-RPT-023**
The subsystem MUST map incoming “post” terminology according to Table 6-C before export-model materialization.

**Table 6-C. `post` source-family anti-corruption mapping**

| Input term | Canonical Cartulary target |
| --- | --- |
| analyst note, draft update, status prose, communication excerpt, handoff note, lesson text | `artifact` with an appropriate `artifact_type`, or the existing coordination surface that owns the source record. |
| stakeholder update or meeting, chat, or email summary | `comm_log` artifact or Communications Log surface record. |
| finding-like report claim | Standardized `finding` artifact subtype if implemented; otherwise a curated export-model block with support references. |
| investigative query text | Standardized `investigative_query` artifact subtype if implemented; otherwise `artifact`. |
| external social-media or third-party “post” | Evidence excerpt or artifact excerpt linked to source evidence; future first-class source record only when Core defines it. |

**REQ-RPT-024**
When a fixture, import, template binding, or report section names an unmapped `post` source family, the implementation MUST fail with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='unsupported_source_family'` before render persistence.

# 7. Reporting source boundary, release tuple, and output options

## 7.1 Source boundary

**REQ-RPT-025**
The reporting subsystem MUST accept case content only through an immutable snapshot boundary admitted by the Core Snapshot and Reporting route family. It MUST NOT read mutable workbook tables, mutable projection tables, live search indexes, live graph views, or live evidence metadata after `materialize_reporting_export_model_v1` completes.

**REQ-RPT-026**
The implementation MUST enforce the source boundary with an observable guard. Any attempt by a renderer, template, Mermaid generator, Slidev generator, asset resolver, or validation stage to query live workbook tables or mutable projections after export-model materialization MUST fail the render with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='live_query_after_export_model'`.

## 7.2 Release tuple

**REQ-RPT-027**
A render operation MUST bind to one immutable source tuple with at least the members in Table 7-A. Unknown members in the normalized Reporting-owned tuple are invalid unless a later revision defines an extension container.

**Table 7-A. Release tuple members**

| Field | Required rule |
| --- | --- |
| `release_id` | Required stable release candidate identity from Core release creation. |
| `incident_id` | Required incident identity. |
| `snapshot_id` | Required immutable snapshot identity. |
| `snapshot_at` | Required source snapshot timestamp. |
| `source_change_set_high_watermark` | Required source-state high-water mark when the snapshot exposes it; otherwise `null` with `snapshot_boundary_kind` naming the Core-owned immutable boundary. |
| `snapshot_boundary_kind` | Required string or `null`; non-null when high watermark is null. |
| `render_admitted_at` | Required deterministic timestamp fixed by Core render admission. All Reporting-owned generated timestamps that participate in canonical bytes MUST equal this value unless the field is explicitly diagnostic-only. |
| `derivation_version` | Required reporting derivation version. It MUST resolve to an adopted `cartulary.reporting_derivation_profile.v1` under REQ-RPT-027a. |
| `template_id` | Required local template identity. |
| `template_version` | Required exact template version; `latest` is invalid. |
| `template_manifest_sha256` | Required digest of canonical template-pack manifest bytes. |
| `redaction_profile_id` | Required exact redaction profile identity. |
| `redaction_profile_version` | Required exact redaction profile version. |
| `redaction_profile_sha256` | Required digest of redaction profile bytes. |
| `release_scope` | Required closed token from §7.3. |
| `recipient_partition_refs[]` | Required array. It MAY be empty only when `release_scope!='external_release'`; omission behavior is invalid. |
| `output_kind` | Required closed token from §7.4. |
| `output_options` | Required normalized object conforming to §7.5. If omitted on a public route, Core 01 MUST materialize defaults before Reporting receives the tuple. |
| `render_environment_profile_id` | Required exact profile identifier from the template pack and toolchain snapshot. |

**REQ-RPT-027a**
`derivation_version` MUST resolve to exactly one adopted `cartulary.reporting_derivation_profile.v1`. That profile is the versioned owner of every snapshot-to-export-model content-derivation decision that an export-model schema in §9 does not itself fix, and it MUST close each obligation in Table 7-A1 deterministically. The Reporting-owned derivations `derive_section_ordering_key_v1` (REQ-RPT-040a) and `derive_display_token_v1` (REQ-RPT-063a) are fixed by this NLSpec and MUST NOT be redefined by a derivation profile. Until the referenced profile is adopted, external-release conformance is blocked under §5: an external-release attempt MUST fail closed with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='blocked_core_dependency'`, and the validation summary MUST record `blocked_core_dependency='reporting_derivation_profile'`.

**Table 7-A1. Derivation-profile closure obligations**

| Derivation decision | Required deterministic closure |
| --- | --- |
| Record and timeline-event selection | A predicate over immutable snapshot state fixing which source records and timeline events enter the export model. |
| `field_key` assignment | A mapping from source field identity to stable `field_key`, independent of visible labels. |
| `display_label` derivation | Label text or explicit null per §9.3, redaction-safe after the redaction phase. |
| Section expansion | Expansion of template `sections[]` declarations into emitted sections, supplying the `decl_index` and expansion-dimension sort consumed by `derive_section_ordering_key_v1` (REQ-RPT-040a). |
| Support-reference selection | Selection of `support_index[]` entries and per-object `support_refs[]`. |
| Ordinal assignment | Contiguous `field_ordinal` and `block_ordinal` assignment consistent with §9.3. |

## 7.3 Release-scope vocabulary

**REQ-RPT-028**
The reporting subsystem MUST use the release-scope vocabulary in Table 7-B. It MUST NOT accept local synonyms.

**Table 7-B. Release-scope vocabulary**

| `release_scope` | Approval and redaction consequence |
| --- | --- |
| `internal_draft` | No Reporting-owned approval requirement. Redaction MAY be applied. `recipient_partition_refs[]` MUST be empty. |
| `internal_review` | Core reviewer approval applies when Core requires it. Redaction MAY be applied. `recipient_partition_refs[]` MUST be empty. |
| `external_release` | Core external-release approvals apply. `recipient_partition_refs[]` MUST be non-empty. Redaction, token, disclosure-partition, support-reference, local-asset, sandbox, and bundle validation gates are mandatory. |

## 7.4 Output-kind vocabulary

**REQ-RPT-029**
The Reporting Subsystem MUST accept only the output kinds in Table 7-C for current v1 conformance. It MUST NOT accept case variants, aliases, or future-only names.

**Table 7-C. Current v1 output kinds**

| `output_kind` | Required behavior |
| --- | --- |
| `mermaid` | Generate canonical `.mmd` source using `mermaid_source_serialize_v1`, render SVG when required by §7.5, optionally render PNG when requested and supported, validate security and bundle requirements. |
| `slidev` | Generate canonical `slides.md` using `slidev_markdown_serialize_v1`, render PDF for external release, optionally render PPTX or PNG when requested and supported, validate click steps, security, toolchain, and bundle requirements. |

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
| `png` | boolean | No | No | `false` | Valid only when the template pack declares PNG support for the output kind. |
| `pptx` | boolean | No | No | `false` | Valid only for `slidev` and only when the template pack declares PPTX support. |
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

**REQ-RPT-032a**
When `source_only=true` (valid only for `internal_draft` and `internal_review`), default materialization MUST force `pdf`, `svg`, `png`, `pptx`, and `rendered_diagrams` to `false`, and the render bundle MUST contain only source-role files plus the mandatory manifest and validation-role files required by §22. A request that sets `source_only=true` together with any of those members explicitly `true` MUST fail before render output bytes with `error.code='invalid_release_request'` and `reason_code='source_only_conflict'`. `source_only` retains its default `false` when omitted.

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
| `generated_id` | `identifier` with a required prefix named by the owner algorithm followed by exactly 64 lowercase hexadecimal characters. |
| `timestamp` | JSON string in UTC form `YYYY-MM-DDTHH:MM:SSZ` or `YYYY-MM-DDTHH:MM:SS.ffffffZ`. It MUST be a valid proleptic Gregorian calendar date-time: month `01`–`12`, day `01` through the last day valid for that month and year, hour `00`–`23`, minute `00`–`59`, and second `00`–`59`. Leap seconds and any out-of-range or non-existent calendar value (for example `2026-02-30T00:00:00Z` or `2026-13-01T00:00:00Z`) are invalid and MUST fail at the earliest attributable validation stage with `reason_code='invalid_timestamp_value'`. Generated Reporting timestamps MUST use exactly six fractional digits when the value participates in canonical bytes. |
| `sha256_hex` | JSON string containing exactly 64 lowercase hexadecimal characters. |
| `media_type` | JSON string matching an allowed media type row in §22. Template-declared media types are valid only for `local_asset` rows and only when declared in the template manifest. |
| `export_model_path_v1` | JSON string path using `$` root, `.` object member notation for ASCII identifiers, and `[<zero-based-index>]` array notation. Paths are diagnostic and template-binding identifiers, not JSONPath. |
| `bundle_path` | POSIX-relative UTF-8 path with no leading `/`, no empty segment, no `.`, no `..`, no backslash, no NUL, and no duplicate normalized path. |
| `safe_string` | JSON string that has passed redaction for the target scope and contains no C0/C1 controls except LF in fields that explicitly allow multiline text. |
| `finite_integer` | JSON number token matching `0` or `-?[1-9][0-9]*` with mathematical value in `[-9007199254740991, 9007199254740991]`. Decimal-point notation, exponent notation, leading plus sign, leading zeroes, and `-0` are invalid. |

**REQ-RPT-037**
Every schema object defined by this NLSpec is closed. Unknown members are invalid unless the object contains an explicitly named extension container defined by this revision. Omitted optional members MUST materialize to their declared default before canonical serialization when the member participates in hash input. Explicit JSON `null` is valid only for members whose `Nullable` column says `Yes`.

# 9. Canonical reporting export model and child schemas

## 9.1 Top-level export model

**REQ-RPT-038**
`cartulary.reporting_export_model.v1` is the only canonical case-content render input for this subsystem. Templates, Mermaid generation, Slidev generation, render validation, redaction, tokenization, and bundle generation MUST read case content only from this model after materialization.

**REQ-RPT-039**
A reporting export model MUST be a JSON-compatible object with exactly the top-level members in Table 9-A. Unknown top-level members are invalid.

**Table 9-A. `cartulary.reporting_export_model.v1` top-level schema**

| Member | Type | Required | Nullable | Default | Ordering and notes |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_export_model.v1`. |
| `export_model_id` | `generated_id` | Yes | No | None | Generated under §10. |
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
| `sections[]` | array of `cartulary.reporting_section.v1` | Yes | No | `[]` invalid for external release | Ordered by `ordering_key`, then `section_id`. |
| `records[]` | array of `cartulary.reporting_record_summary.v1` | Yes | No | `[]` | Ordered by `record_type`, then `record_id`. |
| `relationships[]` | array of `cartulary.reporting_relationship_summary.v1` | Yes | No | `[]` | Ordered by `relationship_id`. |
| `timeline_events[]` | array of `cartulary.reporting_timeline_event.v1` | Yes | No | `[]` | Ordered by §14.2. |
| `subjects[]` | array of `cartulary.tokenizable_subject.v1` | Yes | No | `[]` | Ordered by `stable_subject_ref`. |
| `diagrams[]` | array of `cartulary.reporting_diagram.v1` | Yes | No | `[]` | Ordered by `diagram_id`. |
| `assets[]` | array of asset declarations | Yes | No | `[]` | Ordered by `bundle_path`. |
| `support_index[]` | array of source refs | Yes | No | `[]` | Ordered by `source_ref_id`. |
| `validation_summary` | `cartulary.reporting_export_model_validation.v1` | Yes | No | None | Export-model-local validation summary per Table 9-I (REQ-RPT-046a); distinct from the render validation summary. |

## 9.2 Section object

**REQ-RPT-040**
`cartulary.reporting_section.v1` MUST use Table 9-B.

**Table 9-B. Section object schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_section.v1`. |
| `section_id` | `identifier` | Yes | No | None | Unique in export model. |
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

## 9.3 Block and field objects

**REQ-RPT-041**
`cartulary.reporting_block.v1` MUST use Table 9-C.

**Table 9-C. Block object schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_block.v1`. |
| `block_id` | `identifier` | Yes | No | None | Unique in export model. |
| `block_kind` | string | Yes | No | None | `paragraph`, `bullet_list`, `table`, `metric`, `timeline_rows`, `diagram_ref`, `asset_ref`, `speaker_note`, or `overflow_summary`. |
| `block_ordinal` | `finite_integer` | Yes | No | None | Starts at `1` within parent container, no gaps. |
| `parent_block_id` | `identifier` | Yes | Yes | None | Null for top-level blocks. |
| `split_from_block_id` | `identifier` | Yes | Yes | None | Non-null only for `split_mixed_block_v1` output. |
| `content_class` | string | Yes | No | None | `case_fact`, `derived_summary`, `presentation_text`, `support_reference`, `validation`, or `template_boilerplate`. |
| `aggregate_only_non_identifying` | boolean | Yes | No | `false` | True only when §12.4 passes. |
| `aggregate_policy_id` | string | Yes | Yes | None | `aggregate_public_v1` when aggregate public proof is used; otherwise null. |
| `contributor_count` | `finite_integer` | Yes | Yes | None | Required when aggregate policy is used. |
| `excluded_field_keys[]` | array | Yes | No | `[]` | Ordered by exact field key. |
| `fields[]` | array of field objects | Yes | No | `[]` | Ordered by `field_ordinal`. |
| `children[]` | array of block objects | Yes | No | `[]` | Ordered by `block_ordinal`. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-042**
`cartulary.reporting_field.v1` MUST use Table 9-D.

**Table 9-D. Field object schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_field.v1`. |
| `field_key` | string | Yes | No | None | Stable field key from view schema or Reporting derivation. |
| `display_label` | `safe_string` | Yes | Yes | None | Null when no label is emitted. |
| `field_ordinal` | `finite_integer` | Yes | No | None | Starts at `1` within parent field list, no gaps. |
| `source_value_state` | string | Yes | No | None | `present`, `missing`, `null`, `unavailable`, `withheld`, or `derived`. |
| `redacted_value_state` | string | Yes | No | None | `unchanged`, `allowed`, `masked`, `stubbed`, `tokenized`, `dropped`, or `blocked`. |
| `value` | JSON scalar or array | Yes | Yes | None | Post-redaction value or null when state allows null. |
| `raw_value_sha256` | `sha256_hex` | Yes | Yes | None | Required for external release when raw value existed and value was changed. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

## 9.4 Record, relationship, asset, and source-reference schemas

**REQ-RPT-043**
`cartulary.reporting_record_summary.v1` MUST use Table 9-E.

**Table 9-E. Record summary schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.reporting_record_summary.v1`. |
| `record_id` | `identifier` | Yes | No | None | Source record envelope identity. |
| `record_type` | string | Yes | No | None | Core record type token; `post` invalid in this revision. |
| `source_record_ref` | object | Yes | No | None | Contains source family, record ID, and snapshot ID. |
| `display_name` | `safe_string` | Yes | Yes | None | Null when no release-safe label is emitted. |
| `deleted_state` | string | Yes | No | None | `active`, `deleted`, or `superseded`. Release eligibility per REQ-RPT-043a. |
| `fields[]` | array | Yes | No | `[]` | Field objects ordered by `field_ordinal`. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-043a**
A record summary's `deleted_state` MUST govern release eligibility as in Table 9-E1. `active` records are eligible in every scope. A `superseded` record is eligible for internal scopes; for `external_release` it MUST be excluded by default and MAY be included only when the template section explicitly opts in and the redaction profile permits it under the §5.2 Core 04 companion edit. A `deleted` record MUST NOT appear in any rendered output bytes or in any bundle file that carries case content; it MAY be counted only in export-model-local validation summaries. If a template binding or a required section forces a `deleted` record into external-release output, materialization MUST fail with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='deleted_record_not_releasable'`.

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
| `src_record_ref` | object | Yes | No | None | Source endpoint reference. |
| `dst_record_ref` | object | Yes | No | None | Destination endpoint reference. |
| `direction` | string | Yes | No | None | `directed`, `undirected`, or `bidirectional`. |
| `confidence` | string | Yes | Yes | None | Null when source does not expose confidence. |
| `source_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `support_refs[]` | array | Yes | No | `[]` | Ordered by exact identifier. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty after partition assignment. |

**REQ-RPT-045**
Asset declaration objects MUST use Table 9-G.

**Table 9-G. Asset declaration schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `asset_id` | `identifier` | Yes | No | None | Unique in export model. |
| `asset_kind` | string | Yes | No | None | `local_asset`, `local_theme`, `rendered_diagram`, `rendered_slide`, or `source_asset`. |
| `bundle_path` | `bundle_path` | Yes | No | None | Must match role/path matrix in §22 when bundle-emitted. |
| `media_type` | `media_type` | Yes | No | None | Must be allowed for role. |
| `sha256` | `sha256_hex` | Yes | No | None | Raw file bytes digest. |
| `byte_size` | `finite_integer` | Yes | No | None | Exact raw byte count. |
| `declared_by` | string | Yes | No | None | `template_pack`, `renderer`, `mermaid`, `slidev`, or `reporting_validation`. |
| `required_for_release` | boolean | Yes | No | None | True when omission invalidates release. |
| `disclosure_partition_refs[]` | array | Yes | No | None | Non-empty unless asset is validation-only and contains no case content. |

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

## 9.5 Closed nested objects

**REQ-RPT-046a**
The nested objects `validation_summary` (a member of Table 9-A), `content_class_summary` and `section_validation` (members of Table 9-B), and `display_times` (a member of Table 14-B) are closed objects whose exact members are defined in Tables 9-I, 9-J, 9-K, and 14-D respectively. Each MUST always serialize its full member set with explicit values so that canonical bytes are fully determined; absent data uses the stated null or `0` value, never member omission. Unknown members are invalid. A later revision MAY add a named extension container to any of these objects under REQ-RPT-037.

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
| `case_fact` | `finite_integer` | Yes | No | `0` | Count of blocks with `content_class='case_fact'`. |
| `derived_summary` | `finite_integer` | Yes | No | `0` | Count for `derived_summary`. |
| `presentation_text` | `finite_integer` | Yes | No | `0` | Count for `presentation_text`. |
| `support_reference` | `finite_integer` | Yes | No | `0` | Count for `support_reference`. |
| `validation` | `finite_integer` | Yes | No | `0` | Count for `validation`. |
| `template_boilerplate` | `finite_integer` | Yes | No | `0` | Count for `template_boilerplate`. |

**Table 9-K. `section_validation` object**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `result` | string | Yes | No | None | `passed` or `failed`. |
| `issue_count` | `finite_integer` | Yes | No | `0` | Count of section-local validation issues. |

# 10. Canonical JSON, identifiers, hashes, and deterministic timestamps

**REQ-RPT-047**
`reporting_canonical_json_v1` MUST be used for every Reporting-owned canonical object that participates in a hash. It MUST reject non-JSON values, duplicate object members, numbers outside `finite_integer`, NaN, infinities, binary blobs, and host-language object values.

**REQ-RPT-048**
Canonical object members MUST be serialized in bytewise ascending UTF-8 order of member names. Arrays MUST preserve the owner-defined order from the relevant schema or algorithm. Strings MUST serialize using JSON string escaping with shortest valid escapes for quotation mark, reverse solidus, and required control escapes. No insignificant whitespace is emitted.

**REQ-RPT-049**
The hash input for a canonical object MUST be:

```text
<schema_id> LF
<canonical_json_bytes>
```

The output digest MUST be lowercase 64-character SHA-256 hexadecimal.

**REQ-RPT-050**
Reporting-generated identifiers MUST be derived from canonical tuple bytes using lowercase SHA-256 and a stable prefix. Implementations MUST NOT use random numbers, database sequence order, map iteration order, object insertion order, or render-time wall-clock values in generated IDs.

**REQ-RPT-050a**
`content_manifest_digest_v1` MUST be used to reduce a set of files to one digest wherever this NLSpec requires a digest "of every" file in a declared set (for example `font_manifest_sha256` and `package_store_digest` in §20). The algorithm MUST:

1. enumerate every file in the declared set as a `{path, sha256}` pair, where `path` is the file's POSIX-relative UTF-8 path in Unicode NFC and `sha256` is the lowercase SHA-256 of the file's exact bytes; file modification time, ownership, and permission bits MUST NOT be inputs;
2. reject symbolic links, hard links, and filesystem-special entries in the set;
3. sort the pairs bytewise ascending by `path`, rejecting any duplicate normalized path;
4. serialize the sorted array under `reporting_canonical_json_v1`; and
5. emit the lowercase 64-character SHA-256 of that canonical byte string.

Because the digest depends only on relative paths and content bytes, two conforming implementations enumerating the same file set MUST produce the same digest.

**REQ-RPT-051**
A Reporting-owned generated timestamp participates in canonical hash input only when it equals `render_admitted_at`. Wall-clock timestamps MAY appear only in diagnostic-only artifacts excluded from `output_sha256`; such fields MUST be named `observed_at` or `diagnostic_observed_at`.

**REQ-RPT-052**
The fields in Table 10-A MUST equal `render_admitted_at` whenever they appear in hash-participating objects.

**Table 10-A. Deterministic generated timestamp fields**

| Field | Required rule |
| --- | --- |
| `export_model_created_at` | Equals `render_admitted_at`; participates in export-model hash. |
| `toolchain_snapshot.created_at` | Equals `render_admitted_at`; participates in toolchain snapshot hash. |
| `bundle_created_at` | Equals `render_admitted_at`; participates in bundle-manifest hash. |
| `validation_summary.created_at` | Equals `render_admitted_at` when retained in bundle. |
| Redaction manifest generated timestamps | Equal `render_admitted_at` when emitted. |
| Token manifest generated timestamps | Equal `render_admitted_at` when emitted. |

# 11. Export-model materialization algorithm

**REQ-RPT-053**
`materialize_reporting_export_model_v1` MUST execute the deterministic stages in Table 11-A in order. A stage MUST NOT read live source state after the source boundary is frozen for that stage.

**Table 11-A. Materialization stages**

| Order | Stage | Required behavior |
| ---: | --- | --- |
| 1 | Decode and validate release tuple | Validate tuple fields, defaults, output options, and Core dependency status. |
| 2 | Bind deterministic timestamps | Copy `render_admitted_at` into every hash-participating generated timestamp field. |
| 3 | Validate template manifest | Validate `cartulary.reporting_template_pack_manifest.v1`, supported output kind, supported release scope, bindings, assets, and render profiles. |
| 4 | Read immutable snapshot | Read only the Core-owned immutable snapshot boundary. |
| 5 | Materialize source refs | Create source-reference objects for selected source material. |
| 6 | Materialize records and relationships | Emit record summaries and relationship summaries using §9 schemas. |
| 7 | Materialize timeline events | Emit timeline event objects using §14 and Core-provided timeline sort fields. |
| 8 | Materialize subjects | Emit tokenizable subjects using §13. |
| 9 | Assemble sections and assign ordering keys | Emit sections and blocks per §9 and assign each `ordering_key` via `derive_section_ordering_key_v1` (REQ-RPT-040a). |
| 10 | Assign disclosure partitions | Execute §12 algorithms. |
| 11 | Apply redaction | Execute §13 redaction and token algorithms, including `derive_display_token_v1` token substitution. |
| 12 | Prepare graph adapters | Resolve Graph Projection inputs and diagram selection rules under §15. |
| 13 | Validate resources | Apply §25 resource limits before render source generation. |
| 14 | Canonicalize export model | Serialize under §10 and compute export-model hash. |

**REQ-RPT-054**
Materialization failures MUST use Table 11-B. The first failure MUST be selected by validation issue ordering in §23.4.

**Table 11-B. Materialization failures**

| Condition | `error.code` | `failure_code` | `reason_code` |
| --- | --- | --- | --- |
| Duplicate stable ID | `release_render_failed` | `export_model_invalid` | `duplicate_stable_id` |
| Duplicate export-model path | `release_render_failed` | `export_model_invalid` | `duplicate_export_model_path` |
| Dangling source reference | `release_render_failed` | `export_model_invalid` | `dangling_source_ref` |
| Dangling support reference | `release_render_failed` | `export_model_invalid` | `dangling_support_ref` |
| Missing content class | `release_render_failed` | `content_class_missing` | `content_class_missing` |
| Invalid generated identifier | `release_render_failed` | `export_model_invalid` | `invalid_generated_identifier` |
| Export model exceeds resource limit | `release_render_failed` | `export_model_resource_limit_exceeded` | Limit-specific reason from §25 |
| Missing blocked Core dependency | `release_render_failed` | `export_model_invalid` | `blocked_core_dependency` |

# 12. Party assignments, disclosure partitions, aggregate-public proof, and mixed content

## 12.1 Reporting-facing assignment projection

**REQ-RPT-055**
Reporting MUST consume party assignment source state through the Core-owned projection in Table 12-A. If Core has not adopted this projection, external-release conformance is blocked under §5.

**Table 12-A. `record_party_assignment.v1` Reporting-facing projection**

| Field | Required rule |
| --- | --- |
| `assignment_id` | Required stable identifier. |
| `incident_id` | Required. |
| `subject_record_id` | Required Host or Identity record ID. |
| `subject_record_type` | `host` or `identity`. |
| `party_id` | Required Party record ID. |
| `assignment_kind` | `subject_party`, `responsible_party`, or `observer_party`. |
| `assignment_state` | `active`, `superseded`, or `deleted`. Reporting consumes only `active`. |
| `source_ref` | Source reference object or null. |
| `provenance` | `manual`, `import`, `system`, or `rollback`. |
| `created_by_user_id` | Required. |
| `created_at` | Required. |
| `superseded_by_assignment_id` | Nullable. |
| `deleted_at` | Nullable. |
| `row_version` | Required. |

**REQ-RPT-056**
A snapshot containing more than one active assignment for the same `(subject_record_id, party_id, assignment_kind)` MUST fail export-model materialization with `failure_code='export_model_invalid'` and `reason_code='duplicate_active_party_assignment'`.

## 12.2 Disclosure partition vocabulary and ordering

**REQ-RPT-057**
Disclosure partition references MUST use the vocabulary in Table 12-B. Arrays of disclosure partition references MUST be sorted by `partition_order`, then exact partition token.

**Table 12-B. Disclosure partition references**

| Partition ref | `partition_order` | Meaning |
| --- | ---: | --- |
| `public` | 0 | Eligible for every recipient only after §12.4 or no subject-bearing content is present. |
| `party:{party_id}` | 10 | Eligible only when the recipient partition includes the Party. |
| `internal_only` | 90 | Eligible only for internal draft or internal review. |
| `blocked` | 99 | Not eligible for release output; must be dropped, redacted, or fail. |

**REQ-RPT-058**
Direct Party display labels default to non-public. A Party display value MAY be assigned `public` only when the Party source state marks the label as public-directory eligible and the redaction profile explicitly permits public party labels for the release scope. Omission of either condition means the value remains party-partitioned or is redacted.

## 12.3 Partition assignment algorithm

**REQ-RPT-059**
`assign_party_disclosure_partitions_v1` MUST apply Table 12-C in order. Later rows do not override an earlier `blocked` partition unless the redaction stage removes the blocked content entirely.

**Table 12-C. Disclosure partition assignment**

| Content condition | Partition result |
| --- | --- |
| Template-owned boilerplate with no case facts | `public`. |
| Case content that references no Host, Identity, Party, tokenizable subject, source display value, or free-text case detail | `public`. |
| Subject has one or more active `subject_party` assignments | One `party:{party_id}` per active assignment, ordered by §12.2. |
| Subject has no active assignment and source content is not redacted away | `internal_only` for internal scopes; `blocked` for external release. |
| Derived aggregate passes `aggregate_public_v1` | `public`. |
| Mixed-content block contains multiple partition sets | Execute `split_mixed_block_v1`; if split invalid, redact, drop, or fail. |
| Any content class marked security-sensitive by redaction profile | `blocked` unless redaction removes the sensitive value. |
| Content from a record with `deleted_state='deleted'`, or a `superseded` record not opted in for external release under REQ-RPT-043a | `blocked`; the content MUST be dropped before render and MUST NOT be emitted. |

## 12.4 Aggregate-public proof

**REQ-RPT-060**
Derived summaries inherit the union of contributor partitions by default. A derived summary MAY become `public` only when `aggregate_public_v1` passes every condition in Table 12-D.

**Table 12-D. `aggregate_public_v1` conditions**

| Condition | Required rule |
| --- | --- |
| Contributor threshold | Each visible aggregate cell or scalar is derived from at least `3` distinct source subjects unless the block is template-owned non-case boilerplate. |
| Excluded fields | The derivation records `excluded_field_keys[]` containing all subject display values, identifiers, source snippets, object keys, raw paths, and free-text fields excluded from derivation. |
| Output value kind | Output contains only counts, bounded categorical labels, bucketed times per REQ-RPT-060a, or template-owned prose. |
| Rare contributor bucket | No visible bucket has contributor count `1` or `2`. |
| Support safety | External support summaries are themselves redacted and release-eligible. |
| Trace | The block records `aggregate_only_non_identifying=true`, `aggregate_policy_id='aggregate_public_v1'`, `contributor_count`, and `excluded_field_keys[]`. |

**REQ-RPT-060a**
When `aggregate_public_v1` output contains bucketed times, the bucketing MUST be deterministic. The redaction profile selects one `time_bucket_granularity` from the closed set `hour`, `day`, `week`, or `month`; when it selects none, the default is `day`. Each contributing `activity_sort_ts` is truncated toward the start of its bucket in UTC and labeled with the exact format in Table 12-F. Rows whose time is unresolved (`time_bucket='unresolved_time'` under §14) MUST all map to the single bucket labeled `unresolved`. Bucket labels are the only time values an aggregate-public block emits; raw or precise times MUST NOT appear.

**Table 12-F. Time-bucket label formats**

| `time_bucket_granularity` | UTC truncation | Label format | Example |
| --- | --- | --- | --- |
| `hour` | Start of hour | `YYYY-MM-DDTHH` | `2026-07-02T14` |
| `day` (default) | Start of day | `YYYY-MM-DD` | `2026-07-02` |
| `week` | Start of ISO-8601 week (Monday) | `YYYY-Www` | `2026-W27` |
| `month` | Start of month | `YYYY-MM` | `2026-07` |

## 12.5 Mixed-content splitting

**REQ-RPT-061**
`split_mixed_block_v1` MUST use the deterministic algorithm in Table 12-E.

**Table 12-E. `split_mixed_block_v1`**

| Step | Required behavior |
| ---: | --- |
| 1 | Split only at declared structural boundaries: bullet item, table row, metric item, timeline row, diagram label, or child field. |
| 2 | Child block IDs are `{parent_block_id}__split_{0001..N}`. |
| 3 | Child ordering preserves original structural order. |
| 4 | Child `source_refs[]`, `support_refs[]`, `content_class`, and `ordering_key` are inherited and narrowed to included content. |
| 5 | Child `disclosure_partition_refs[]` is recomputed from included content. |
| 6 | Free-form paragraphs MUST NOT be split by NLP. They must be tokenized, stubbed, dropped, or fail with `reason_code='disclosure_partition_mixed_content'`. |

# 13. Redaction, tokenization, token manifests, and reveal maps

## 13.1 Reporting-facing redaction interface

**REQ-RPT-062**
Reporting MUST consume a Core-owned redaction profile through the interface in Table 13-A. If Core has not adopted token-backed `mask` and `stub` parameters, token substitution is blocked under §5.

**Table 13-A. Core redaction interface consumed by Reporting**

| Field | Required rule |
| --- | --- |
| `redaction_profile_id` | Required. |
| `redaction_profile_version` | Required exact version. |
| `redaction_profile_sha256` | Required digest. |
| `rule_id` | Required for every applied action. |
| `action` | `allow`, `drop`, `mask`, `truncate`, or `stub`. Token substitution uses token-backed `mask` or `stub`. |
| `selected_rule_trace` | Required safe trace object in redaction manifest. |
| `literal_output` | Required post-redaction value or explicit drop record. |

**REQ-RPT-063**
`apply_token_backed_redaction_v1` MUST apply Core redaction rule selection first, then apply Reporting token substitution only when the selected Core rule is token-backed and the field resolves to exactly one stable tokenizable subject.

**REQ-RPT-063a**
When `apply_token_backed_redaction_v1` substitutes a token, the emitted `display_token` MUST be produced by the Reporting-owned, fully deterministic `derive_display_token_v1`:

- **Grammar.** A `display_token` matches `^(HOST|IDENTITY|PARTY|SUBJECT)-[0-9]{4,}$`. Its character set is limited to `A`–`Z`, `0`–`9`, and `-`, which is a no-op under both `mermaid_source_serialize_v1` escaping (Table 16-C) and `slidev_markdown_escape_v1` (Table 18-B); a token therefore appears byte-identically in `.mmd` and `slides.md`.
- **Family tag.** The tag is `HOST`, `IDENTITY`, or `PARTY` from the subject's `subject_family`, unless the redaction profile sets `neutral_token_family=true` (default `false`), in which case every tag is `SUBJECT`.
- **Ordinal assignment.** Within one release, ordinals are assigned per emitted family in ascending `stable_subject_ref` order — the order in which `subjects[]` is already sorted (Table 9-A) — starting at `1`, incrementing by `1`, one token per distinct `stable_subject_ref`. The ordinal is left-zero-padded to four digits; the §25 `subjects.count` hard limit guarantees four digits suffice.
- **Stability and non-reversibility.** The mapping from `stable_subject_ref` to `display_token` is one-to-one within a release and is recorded in the token manifest (Table 13-D) and reveal map (Table 13-E). The visible token MUST NOT encode any part of the subject's display value or its `canonical_display_value_sha256`.

Two conforming implementations given the same subjects and redaction profile MUST emit identical `display_token` values.

## 13.2 Tokenizable subject schema

**REQ-RPT-064**
`cartulary.tokenizable_subject.v1` MUST use Table 13-B.

**Table 13-B. Tokenizable subject schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.tokenizable_subject.v1`. |
| `stable_subject_ref` | Required `stable_subject_ref_v1`. |
| `subject_family` | `host`, `identity`, or `party`. |
| `source_record_id` | Required for canonical Host, Identity, or Party; null only for unresolved mention subjects. |
| `source_record_type` | `host`, `identity`, `party`, or null for unresolved mention. |
| `entity_mention_id` | Required when `source_record_id=null`; otherwise null. |
| `canonical_display_value_sha256` | SHA-256 over canonical object `{schema_id, stable_subject_ref, value}`. |
| `disclosure_partition_refs[]` | Non-empty, sorted by §12.2. |
| `source_refs[]` | Non-empty for incident-derived subjects. |
| `support_refs[]` | May be empty only for unresolved rough-capture subjects. |

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
`cartulary.reporting_token_manifest.v1` MUST contain exactly the top-level fields in Table 13-D. External token manifests MUST NOT include raw subject display values or reversible display material.

**Table 13-D. Token manifest schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.reporting_token_manifest.v1`. |
| `release_id` | Required. |
| `snapshot_id` | Required. |
| `redaction_profile_sha256` | Required. |
| `created_at` | Equals `render_admitted_at`. |
| `entries[]` | Required array ordered by `token_id`. |
| `entries[].token_id` | Stable generated ID. |
| `entries[].display_token` | Deterministic token string produced by `derive_display_token_v1` (REQ-RPT-063a). |
| `entries[].stable_subject_ref` | Required. |
| `entries[].subject_family` | Required. |
| `entries[].source_record_id` | Nullable. |
| `entries[].entity_mention_id` | Nullable. |
| `entries[].canonical_display_value_sha256` | Required. |
| `entries[].recipient_partition_refs[]` | Required, ordered by §12.2. |
| `entries[].action` | `mask` or `stub`. |
| `entries[].rule_id` | Redaction rule that selected token output. |

## 13.4 Reveal map

**REQ-RPT-068**
`cartulary.reporting_token_reveal_map.v1` MUST use Table 13-E and MUST be retained only as a Core-authorized internal sensitive release artifact.

**Table 13-E. Reveal-map schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.reporting_token_reveal_map.v1`. |
| `release_id` | Required. |
| `snapshot_id` | Required. |
| `token_manifest_sha256` | Required. |
| `redaction_profile_sha256` | Required. |
| `created_at` | Equals `render_admitted_at`. |
| `entries[]` | Required array ordered by `token_id`. |
| `entries[].token_id` | Required. |
| `entries[].display_token` | Required. |
| `entries[].stable_subject_ref` | Required. |
| `entries[].subject_family` | Required. |
| `entries[].source_record_id` | Nullable. |
| `entries[].entity_mention_id` | Nullable. |
| `entries[].canonical_display_value` | Required internal value; forbidden in external bundle. |
| `entries[].canonical_display_value_sha256` | Required. |
| `entries[].recipient_partition_refs[]` | Required. |

**REQ-RPT-069**
Reveal maps MUST NOT be listed as `required_for_release=true`, MUST NOT be included in external bundles, MUST NOT appear in token manifests, and MUST NOT be readable through ordinary external release download surfaces.

# 14. Timeline section derivation

## 14.1 Core-owned timeline sort materialization

**REQ-RPT-070**
Reporting MUST consume Core-provided timeline sort fields in Table 14-A for every selected timeline event. Reporting MUST NOT independently parse, reinterpret, or normalize source timestamp text for ordering when those fields exist.

**Table 14-A. Timeline sort materialization fields**

| Field | Rule |
| --- | --- |
| `activity_sort_ts` | UTC timestamp or null. |
| `activity_parse_state` | `parsed`, `missing`, `incomplete`, `ambiguous`, or `unparseable`. |
| `activity_precision_rank` | `6=second`, `5=minute`, `4=hour`, `3=day`, `2=month`, `1=year`, `0=unresolved`. |
| `date_entered_sort_key` | `YYYY-MM-DD` or null. |
| `source_time_text` | Original source text, redaction-eligible. |
| `time_bucket` | `parsed_time` or `unresolved_time`. |

**REQ-RPT-071**
If the immutable snapshot lacks any required timeline sort materialization field for a selected timeline event, external release MUST fail with `failure_code='export_model_invalid'` and `reason_code='timeline_sort_key_missing'`.

## 14.2 Timeline event object and sort key

**REQ-RPT-072**
`cartulary.reporting_timeline_event.v1` MUST include the fields in Table 14-B.

**Table 14-B. Timeline event schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.reporting_timeline_event.v1`. |
| `record_id` | Required source timeline record ID. |
| `activity_sort_key` | Required serialized sort key from REQ-RPT-073. |
| `activity_sort_ts` | UTC timestamp or null. |
| `activity_parse_state` | From Table 14-A. |
| `activity_precision_rank` | From Table 14-A. |
| `date_entered_sort_key` | Date key or null. |
| `display_times` | Closed redaction-safe object per Table 14-D (REQ-RPT-072a). |
| `fields[]` | Field objects ordered by `field_ordinal`. |
| `source_refs[]` | Ordered source references. |
| `support_refs[]` | Ordered support references. |
| `disclosure_partition_refs[]` | Non-empty after partition assignment. |

**REQ-RPT-072a**
`display_times` is a closed object with the exact members in Table 14-D. All members are always present; a member with no emitted value is explicit `null`. Every emitted string is redaction-safe after the redaction phase.

**Table 14-D. `display_times` object**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `primary_display` | `safe_string` | Yes | Yes | None | Emitted primary time string, or null when no time is displayed. |
| `precision_label` | string | Yes | Yes | None | One of `second`, `minute`, `hour`, `day`, `month`, `year`, or `unresolved`, matching `activity_precision_rank`; null when no time is displayed. |
| `source_text_display` | `safe_string` | Yes | Yes | None | Redaction-eligible original source time text, or null when withheld or absent. |

**REQ-RPT-073**
`select_timeline_rows_v1` MUST sort selected timeline rows by the tuple below. For descending variants, only the template-declared primary direction changes; tie-breaker order remains ascending.

```text
time_bucket_order
activity_sort_ts or "9999-12-31T23:59:59.999999Z"
negative activity_precision_rank
date_entered_sort_key or "9999-12-31"
record.created_at
record_id
```

`time_bucket_order` is `0` for `parsed_time` and `1` for `unresolved_time`. For a template-declared `unresolved_first` option, invert only `time_bucket_order`.

## 14.3 Timeline slide chunking and bounds

**REQ-RPT-074**
Timeline slide chunking MUST use Table 14-C.

**Table 14-C. Timeline bounds and chunking**

| Limit | Default | Hard limit | Behavior |
| --- | ---: | ---: | --- |
| Selected timeline rows | 240 | 240 | Beyond hard limit, apply the overflow selection rule per REQ-RPT-074a. |
| Rows per timeline slide | 8 | 12 | Template MAY set `1..12`; omitted means `8`. |
| Timeline overflow block | Required when rows omitted | N/A | Contains `omitted_row_count`, `selection_rule_id`, `first_omitted_sort_key`, and `filter_summary`. |

**REQ-RPT-074a**
When more rows are eligible than the `Selected timeline rows` hard limit, the timeline `selection_rule_id` determines behavior. When the template declares no `selection_rule_id`, the default is `timeline_overflow_summary_v1`, which keeps the first `240` rows in `select_timeline_rows_v1` order (REQ-RPT-073) and emits the overflow block. Behavior is scope-graded:

- `internal_draft` and `internal_review`: the default emits the overflow block and MUST NOT fail on volume.
- `external_release`: overflow MUST fail with `error.code='release_render_failed'`, `failure_code='export_model_invalid'`, and `reason_code='timeline_overflow_unresolved'` unless the template explicitly declares a summarizing `selection_rule_id` (for example `timeline_overflow_summary_v1`), so that an external report never silently drops case content without a declared rule.

The overflow block required by Table 14-C records `omitted_row_count`, `selection_rule_id`, `first_omitted_sort_key`, and `filter_summary` in every scope in which rows are omitted.

**REQ-RPT-075**
Chunked timeline slide IDs MUST use `{base_slide_id}__chunk_{0001..N}`. Slide titles MUST append ` ({chunk_index}/{chunk_count})` when `chunk_count>1`.

# 15. Graph Projection consumption and diagram model

## 15.1 Reporting-to-Graph Projection adapter

**REQ-RPT-076**
Reporting consumes completed Graph Projection output through `source_projection_ref.v1`. Reporting MUST NOT redefine projection derivation, projected identity, lifecycle, traversal, or Graph Projection validation.

**REQ-RPT-077**
`source_projection_ref.v1` MUST use Table 15-A.

**Table 15-A. `source_projection_ref.v1`**

| Field | Rule |
| --- | --- |
| `graph_view_id` | Required. |
| `projection_run_id` | Required completed run. |
| `source_snapshot_id` | Must equal the Reporting `snapshot_id` or a Core-declared alternate immutable source-boundary token. |
| `projection_schema_id` | Required. |
| `projection_version` | Required. |
| `projection_config_digest` | Required. |
| `projection_source_digest` | Required. |
| `projection_output_digest` | Required. |

**REQ-RPT-078**
Graph adapter failures MUST use Table 15-B.

**Table 15-B. Graph adapter failures**

| Condition | `failure_code` | `reason_code` |
| --- | --- | --- |
| Projection snapshot mismatch | `graph_projection_unavailable` | `graph_projection_stale` |
| Projection run not completed | `graph_projection_unavailable` | `graph_projection_not_completed` |
| Required vertex or edge missing | `graph_projection_unavailable` | `graph_projection_selection_unresolved` |
| Projection output digest mismatch | `graph_projection_unavailable` | `graph_projection_digest_mismatch` |

## 15.2 Diagram selection rules

**REQ-RPT-079**
Diagram selection rules MUST use Table 15-C. If no graph view is named, Reporting MAY construct Graph Projection input before export-model materialization, but diagram generation MUST still consume a completed projection satisfying Table 15-A.

**Table 15-C. Diagram selection rules**

| Rule | Required inputs | Defaults and bounds | Ordering |
| --- | --- | --- | --- |
| `explicit_refs` | `vertex_refs[]`, `edge_refs[]` | Empty arrays invalid. | Graph Projection output order. |
| `neighborhood` | `seed_vertex_refs[]` | `depth` default `1`, max `2`; `edge_kind_filter[]` default `[]`. | Graph Projection traversal order. |
| `timeline_sequence` | `timeline_record_ids[]` | Non-empty; all IDs must exist in selected timeline. | §14.2 order. |
| `all_with_bounds` | None | Must fit §16.1 Mermaid bounds. | Graph Projection output order. |

## 15.3 Diagram object

**REQ-RPT-080**
`cartulary.reporting_diagram.v1` MUST include Table 15-D fields.

**Table 15-D. Diagram schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.reporting_diagram.v1`. |
| `diagram_id` | Required stable ID. |
| `diagram_kind` | `flowchart` or `sequence`. |
| `source_projection_ref` | Required when graph-derived; otherwise null with `diagram_source_kind='timeline'` or `template_static`. |
| `selection_rule` | One rule from Table 15-C. |
| `included_vertex_refs[]` | Required, ordered by the selection rule. |
| `included_edge_refs[]` | Required, ordered by the selection rule. |
| `source_refs[]` | Required, ordered by exact identifier. |
| `support_refs[]` | Required, ordered by exact identifier. |
| `rendered_refs[]` | Required; MAY be empty before rendering. |
| `overflow_summary` | Null unless content was omitted by a declared bound. |
| `disclosure_partition_refs[]` | Non-empty after partition assignment. |

# 16. Mermaid source generation

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
`mermaid_source_serialize_v1` MUST use the shared serialization rules in Table 16-B.

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
| Security profile | Renderer MUST use strict Mermaid security configuration that disables HTML and click functionality; a wrapper configuration is valid only when it rejects HTML and click functionality before render and emits the same failure codes as this section. |

**REQ-RPT-084**
Flowchart source MUST use this grammar exactly after variable substitution:

```text
flowchart TD
  {node_id}["{escaped_label}"]
  {from_node_id} -->|"{escaped_label}"| {to_node_id}
```

`TD` is the default direction. A template MAY declare `LR`; no other direction is valid. Node declarations sort by selected vertex order. Edge declarations sort by selected edge order.

**REQ-RPT-085**
Sequence source MUST use this grammar exactly after variable substitution:

```text
sequenceDiagram
  participant {participant_id} as "{escaped_label}"
  {from_participant_id}->>{to_participant_id}: {escaped_message}
```

Only `participant` declarations and `->>` messages are valid in v1. Mermaid notes, loops, boxes, links, actor menus, autonumber, styling, activation, comments, and other arrows are future-only.

**REQ-RPT-086**
Mermaid escaping MUST use Table 16-C.

**Table 16-C. Mermaid escaping**

| Input | Serialized output or validation behavior |
| --- | --- |
| `"` | `\"` |
| `\` | `\\` |
| LF, CR, TAB | U+0020 after whitespace collapse. |
| C0/C1 controls | Validation failure. |
| `<`, `>` | A label whose normalized form contains U+003C or U+003E fails validation with `failure_code='mermaid_invalid'` and `reason_code='invalid_mermaid_construct'`. No raw-HTML detection is performed; angle brackets MUST be removed by redaction or derivation substitution before serialization. |
| Backtick | Literal backtick allowed in labels only if not part of a code fence; source comments remain forbidden. |

# 17. Slide-deck model

**REQ-RPT-087**
`cartulary.reporting_slide_deck.v1` MUST use Table 17-A.

**Table 17-A. Slide-deck schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.reporting_slide_deck.v1`. |
| `deck_id` | Required generated ID. |
| `release_id` | Required. |
| `snapshot_id` | Required. |
| `template_id` | Required. |
| `template_version` | Required exact version. |
| `output_options` | Required materialized options from §7.5. |
| `slides[]` | Required, ordered by `slide_ordinal`. |
| `slide_count` | Required integer equal to `slides[].length`. |
| `click_step_count` | Required total count across slides. |
| `expected_export_page_count` | Required sum over slides of `1 + click_steps.length`. |
| `source_refs[]` | Required array ordered by exact identifier. |
| `support_refs[]` | Required array ordered by exact identifier. |

**REQ-RPT-088**
A slide object MUST use Table 17-B.

**Table 17-B. Slide object schema**

| Field | Rule |
| --- | --- |
| `slide_id` | Required stable ID. |
| `slide_ordinal` | Required integer starting at `1`, no gaps. |
| `title` | Required redaction-safe string. |
| `layout_id` | Closed layout token declared by template manifest. |
| `blocks[]` | Required array of allowed slide block kinds. |
| `click_steps[]` | Required array of `click_step.v1`; MAY be empty. |
| `speaker_notes` | Required string or null; null when omitted. |
| `source_refs[]` | Required array. |
| `support_refs[]` | Required array. |
| `disclosure_partition_refs[]` | Required non-empty array. |

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
`slidev_markdown_serialize_v1` MUST use Table 18-A.

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
| Slide separator | Blank line, `---`, blank line. |
| Markdown escaping | Escape generated text runs per `slidev_markdown_escape_v1` (Table 18-B, REQ-RPT-091a). |
| Raw HTML | Forbidden except allowed reveal-only components and speaker-note comments. |
| Mermaid fences | Exact opening ```` ```mermaid```` and canonical `.mmd` source bytes inside. |
| Speaker notes | Final slide block serialized as `<!--\n{escaped note lines}\n-->`. |
| Tables | Pipe table, deterministic column order, escaped cells, no alignment syntax. |
| Code fences | Only `text` and `mermaid` unless an internal-review template declares another closed language. |

**REQ-RPT-091a**
`slidev_markdown_escape_v1` MUST escape generated text runs so that `slides.md` is byte-deterministic and no generated content can alter Markdown or Slidev structure. It applies only to generated text placed in a Markdown text position (paragraph text, heading text, list-item text, table-cell text); it MUST NOT alter serializer-emitted structure such as headmatter, slide separators, fence markers, or the speaker-note comment wrapper. The escape set is closed in Table 18-B and is defined against the pinned Slidev markdown-it dialect (CommonMark plus GFM) recorded in the toolchain snapshot. After escaping, a generated run MUST parse to exactly one text run yielding the intended code points.

**Table 18-B. `slidev_markdown_escape_v1` escape set**

| Position | Characters backslash-escaped |
| --- | --- |
| Any position in a generated text run | REVERSE SOLIDUS (U+005C), GRAVE ACCENT (U+0060), VERTICAL LINE (U+007C), and each of `` `* _ [ ] < ~ $` `` |
| Start of a generated line (first non-space character) | additionally `#`, `>`, `-`, `+`, `=`; a leading run of GRAVE ACCENT or `~`; and a leading `[0-9]+` immediately followed by `.` or `)` |

Escaping inserts a single REVERSE SOLIDUS (U+005C) before the character. Characters outside this set are emitted literally.

**Table 18-C. Headmatter keys**

| Key | Always emitted | Value |
| --- | --- | --- |
| `title` | Yes | Deck title from the deck model. |
| `author` | Yes | Template-declared author, or `""` when none. |
| `theme` | Yes | Template-declared theme. |
| `download` | Yes | `false`. |
| `export.format` | Yes | From materialized `output_options`. |
| `export.withClicks` | Yes | `true`. |
| `clickAnimation` | Yes | `false`. |
| `lineNumbers` | Yes | `false`. |
| `monaco` | Yes | `false`. |
| `twoslash` | Yes | `false`. |

**REQ-RPT-092**
Generated `slides.md` MUST be byte-identical for the same deck model, template manifest, output options, and render environment profile. Valid Markdown variants that differ in bytes are non-conformant.

# 19. Reveal-only click-step profile

**REQ-RPT-093**
`cartulary.slidev_reveal_only.v1` MUST use only deterministic reveal and hide behavior described by `click_step.v1`. It MUST NOT use timing animations, CSS transitions, randomization, arbitrary Vue state, route navigation, remote scripts, or viewer-side runtime dependencies for external release.

**REQ-RPT-094**
`click_step.v1` MUST use Table 19-A.

**Table 19-A. `click_step.v1` schema**

| Field | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `click_step_id` | `identifier` | Yes | No | None | Stable in slide. |
| `ordinal` | integer | Yes | No | None | Starts at `1`, increments by `1` with no gaps. |
| `action` | string | Yes | No | None | `reveal` or `hide`. |
| `targets[]` | array | Yes | No | None | Non-empty array of block or item IDs already present on same slide. |
| `initial_visibility` | string | No | No | Action-derived | Default `hidden` for reveal targets, `visible` for hide targets. |
| `component` | string | Yes | No | None | `v-click`, `v-clicks`, or `v-after`. |
| `at` | integer | Yes | No | None | Equal to `ordinal`. |
| `resulting_state_hash` | `sha256_hex` | Yes | No | None | Hash over target visibility states after this step. |

**REQ-RPT-095**
Click-step components MUST follow Table 19-B.

**Table 19-B. Click-step component constraints**

| Component | Valid use |
| --- | --- |
| `v-click` | One target. |
| `v-clicks` | Homogeneous ordered list or table children only. |
| `v-after` | Same ordinal as immediately prior reveal group only. |

**REQ-RPT-096**
A slide with `N` click steps MUST export `N+1` page states: base state plus one state after each ordinal. A rendered PDF, PPTX, or PNG set whose page-state count differs from this rule MUST fail with `failure_code='slidev_export_failed'` and `reason_code='click_export_page_count_mismatch'`.

# 20. Toolchain snapshot and render environment

**REQ-RPT-097**
`cartulary.reporting_toolchain_snapshot.v1` MUST bind every toolchain and environment value that affects generated or rendered bytes. Floating versions, `latest`, network package resolution during render, and undeclared renderer binaries are invalid.

**REQ-RPT-098**
The toolchain snapshot MUST use Table 20-A.

**Table 20-A. Toolchain snapshot fields**

| Field | Required rule |
| --- | --- |
| `schema_id` | Exact `cartulary.reporting_toolchain_snapshot.v1`. |
| `created_at` | Equals `render_admitted_at`. |
| `node_version` | Exact version string. |
| `package_manager` | Exact name and version. |
| `lockfile_sha256` | Required. |
| `package_store_digest` | Required `content_manifest_digest_v1` (REQ-RPT-050a) over every file in the pinned package store. |
| `slidev_version` | Required when `output_kind='slidev'`. |
| `mermaid_version` | Required when Mermaid generation or rendering occurs. |
| `chromium_version` | Required when browser rendering occurs. |
| `chromium_executable_path` | Required render-sandbox path. |
| `chromium_executable_sha256` | Required. |
| `os_image_id` | Required exact repo-control or container image ID. |
| `os_image_sha256` | Required when image is content-addressed. |
| `timezone` | Exact `UTC`. |
| `locale` | Exact `C.UTF-8` unless template declares another exact locale. |
| `font_manifest_sha256` | Required `content_manifest_digest_v1` (REQ-RPT-050a) over every usable font file and font configuration file. |
| `viewport_css_px` | Default `1280x720`; template MAY declare exact override. |
| `device_scale_factor` | Default `1`. |
| `color_scheme` | `light` default unless template declares `dark`. |
| `browser_launch_args[]` | Exact ordered array. |
| `env_allowlist[]` | Exact names and values or digests. |
| `render_command[]` | Exact argv array; shell string invalid. |
| `network_policy_id` | Exact render sandbox policy ID. |

**REQ-RPT-099**
Rendered-byte determinism is scoped to the exact toolchain snapshot and render environment. An implementation MUST NOT claim byte-identical rendered outputs across undeclared host machines, font sets, locales, timezones, Chromium executables, or viewport settings.

# 21. Template packs and local assets

**REQ-RPT-100**
`cartulary.reporting_template_pack_manifest.v1` MUST be a closed schema. Unknown top-level members are invalid.

**REQ-RPT-101**
Template manifest top-level fields MUST use Table 21-A.

**Table 21-A. Template pack manifest schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.reporting_template_pack_manifest.v1`. |
| `template_id` | Required stable identifier. |
| `template_version` | Required exact version; `latest` invalid. |
| `manifest_version` | Required manifest schema version. |
| `supported_output_kinds[]` | Required subset of `mermaid`, `slidev`. |
| `supported_release_scopes[]` | Required subset of §7.3. |
| `allowed_export_model_bindings[]` | Required array of `binding_path_pattern_v1`. |
| `sections[]` | Required deterministic section expansion declarations. |
| `layouts[]` | Required closed layout tokens and allowed block kinds. |
| `narrative_slots[]` | Required slot declarations. |
| `assets[]` | Required declared local assets with path, media type, SHA-256, byte size, and role. |
| `render_profiles[]` | Required allowed Slidev, Mermaid, and render-environment profiles. |
| `output_options` | Required supported optional outputs. |
| `extension_features[]` | Required array; MAY be empty. |

**REQ-RPT-102**
`binding_path_pattern_v1` is either an exact `export_model_path_v1` or an `export_model_path_v1` where one stable-ID segment is replaced by `*`. `**`, regex, array indexes, visible labels, storage paths, arbitrary JSONPath syntax, and grid vendor coordinates are invalid.

**REQ-RPT-103**
Template extension features MUST use Table 21-B. Unsupported required features fail before render output bytes.

**Table 21-B. Template extension features**

| Field | Rule |
| --- | --- |
| `feature_id` | Namespaced identifier. |
| `required` | Boolean. |
| `payload_sha256` | Optional digest. |
| `payload_path` | Optional local bundle path. |

If `required=true` and the renderer does not understand the feature, render admission MUST fail with `error.code='invalid_release_request'` and `reason_code='unsupported_template_feature'`. If `required=false`, the renderer MAY ignore the feature only when ignoring it does not affect declared bindings, assets, layout, output bytes, security, or validation.

**REQ-RPT-104**
Local assets MUST be declared in the template manifest before use. A generated source file that references an undeclared asset, remote asset, absolute file path, path traversal, `data:` URI, `javascript:` URI, or non-relative URL MUST fail with `failure_code='undeclared_asset'` or `failure_code='remote_asset_reference'` as applicable.

# 22. Render bundle manifest, file roles, and archive boundary

## 22.1 Bundle path and archive boundary

**REQ-RPT-105**
Bundle paths MUST be POSIX-relative UTF-8 strings with no leading `/`, no empty segment, no `.`, no `..`, no backslash, no NUL, and no duplicate path. Sorting is bytewise ascending UTF-8 by `path`.

**REQ-RPT-106**
Physical archive bytes are a delivery wrapper, not the approval hash. `output_sha256` binds to `bundle_manifest_sha256`. Archive extraction for verification MUST reject symlinks, hardlinks, device files, absolute paths, path traversal, duplicate paths, and filesystem-special entries.

## 22.2 Bundle manifest schema

**REQ-RPT-107**
`cartulary.render_bundle_manifest.v1` MUST include at least the fields in Table 22-A. Unknown members are invalid.

**Table 22-A. Bundle manifest top-level schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.render_bundle_manifest.v1`. |
| `release_id` | Required. |
| `snapshot_id` | Required. |
| `output_kind` | `slidev` or `mermaid`. |
| `release_scope` | From §7.3. |
| `bundle_created_at` | Equals `render_admitted_at`. |
| `export_model_sha256` | Required. |
| `toolchain_snapshot_sha256` | Required. |
| `validation_summary_sha256` | Required. |
| `redaction_manifest_sha256` | Required. |
| `token_manifest_sha256` | Required when tokens used; otherwise null. |
| `files[]` | Required array ordered by `path` bytewise ascending UTF-8. |
| `bundle_manifest_sha256` | SHA-256 of canonical manifest object excluding this field. |

**REQ-RPT-108**
Every file item MUST contain `path`, `role`, `media_type`, `byte_size`, `sha256`, and `required_for_release`. File item roles, path patterns, media types, required conditions, and `required_for_release` values MUST follow Table 22-B.

**Table 22-B. File role, path, and media-type matrix**

| Role | Required path pattern | Allowed media type | Required when | `required_for_release` |
| --- | --- | --- | --- | --- |
| `manifest` | `manifest.json` | `application/vnd.cartulary.render-bundle-manifest+json` | Always | `true` |
| `validation_summary` | `validation/summary.json` | `application/vnd.cartulary.reporting-validation+json` | Always | `true` |
| `toolchain_snapshot` | `validation/toolchain.json` | `application/vnd.cartulary.reporting-toolchain+json` | Always | `true` |
| `redaction_manifest` | `validation/redaction-manifest.json` | `application/vnd.cartulary.redaction-manifest+json` | Always | `true` |
| `token_manifest` | `validation/token-manifest.json` | `application/vnd.cartulary.reporting-token-manifest+json` | When tokens used | `true` |
| `source_slidev` | `slides.md` | `text/markdown; charset=utf-8` | `slidev` | `true` |
| `source_mermaid` | `diagrams/{diagram_id}.mmd` | `text/vnd.cartulary.mermaid; charset=utf-8` | `mermaid` or Slidev diagram | `true` |
| `rendered_pdf` | `deck.pdf` | `application/pdf` | External `slidev` | `true` |
| `rendered_pptx` | `deck.pptx` | `application/vnd.openxmlformats-officedocument.presentationml.presentation` | Requested and supported | `true` |
| `rendered_png` | `slides/slide-{0001..N}.png` | `image/png` | Requested and supported | `true` |
| `rendered_svg` | `diagrams/{diagram_id}.svg` | `image/svg+xml` | External `mermaid` or requested diagram render | `true` |
| `local_asset` | `assets/{asset_id}/{filename}` | Template allow-list | Referenced by source | Depends on reference scope |
| `local_theme` | `theme/{path}` | `text/css`, `font/woff2`, `font/ttf`, `image/svg+xml`, `image/png` | Referenced by template | `true` |

# 23. Validation, error-envelope model, and reason codes

## 23.1 Error-envelope model

**REQ-RPT-109**
Reporting errors MUST use the three-level model in Table 23-A. Public payloads and normative requirements MUST NOT combine the Core public error family and Reporting failure code into one colon-delimited machine-readable identifier.

**Table 23-A. Error layers**

| Layer | Field | Rule |
| --- | --- | --- |
| Core public error | `error.code` | `invalid_release_request` before an admitted durable render job; `release_render_failed` after admission. |
| Reporting failure | `failure_code` | Closed registry in §23.5; null only when not applicable or pass state. |
| Specific reason | `reason_code` | Closed registry in §23.6; null only when no finer reason exists. |

## 23.2 Validation summary schema

**REQ-RPT-110**
`cartulary.reporting_render_validation_summary.v1` MUST use Table 23-B.

**Table 23-B. Validation summary schema**

| Field | Rule |
| --- | --- |
| `schema_id` | Exact `cartulary.reporting_render_validation_summary.v1`. |
| `release_id` | Required when allocated; otherwise null with `safe_details.pre_release=true`. |
| `snapshot_id` | Required when known. |
| `result` | `passed` or `failed`. |
| `terminal_stage` | Closed stage token from §23.3. |
| `failure_code` | Null only on pass. |
| `reason_code` | Null only when no finer reason exists. |
| `issues[]` | Required array; empty only on pass. |
| `first_failure` | Null on pass; otherwise copy of first issue identity and codes. |
| `created_at` | Equals `render_admitted_at` for bundle-retained summary. |

## 23.3 Stage vocabulary

**REQ-RPT-111**
Validation stage tokens MUST use this exact ordered vocabulary:

```text
route_admission
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
Validation issues MUST sort by stage order from §23.3, then severity `error` before `warning`, then canonical `export_model_path` or bundle `path`, then `failure_code`, then `reason_code`.

**REQ-RPT-113**
Safe details MAY contain only the keys in Table 23-C unless a reason-code-specific table in this NLSpec adds a key. Forbidden values in Table 23-D MUST NOT appear in safe details.

**Table 23-C. Generic safe-details allow-list**

| Allowed key |
| --- |
| `field` |
| `path` |
| `stage` |
| `limit_key` |
| `limit` |
| `actual` |
| `template_id` |
| `template_version` |
| `output_kind` |
| `release_scope` |
| `graph_view_id` |
| `projection_run_id` |
| `file_role` |
| `media_type` |
| `bundle_path` |
| `pre_release` |
| `blocked_core_dependency` |

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
| `unsupported_source_family` | `export_model_invalid` |
| `live_query_after_export_model` | `export_model_invalid` |
| `duplicate_object_member` | stage-specific |
| `duplicate_stable_id` | `export_model_invalid` |
| `duplicate_export_model_path` | `export_model_invalid` |
| `dangling_source_ref` | `export_model_invalid` |
| `dangling_support_ref` | `export_model_invalid` |
| `blocked_core_dependency` | `export_model_invalid` |
| `duplicate_active_party_assignment` | `export_model_invalid` |
| `disclosure_partition_mixed_content` | `export_model_invalid` |
| `token_subject_not_stable` | `token_manifest_invalid` |
| `token_subject_not_unique` | `token_manifest_invalid` |
| `timeline_sort_key_missing` | `export_model_invalid` |
| `invalid_timestamp_value` | `export_model_invalid` |
| `deleted_record_not_releasable` | `export_model_invalid` |
| `timeline_overflow_unresolved` | `export_model_invalid` |
| `diagrams_count_exceeded` | `export_model_resource_limit_exceeded` |
| `graph_projection_stale` | `graph_projection_unavailable` |
| `graph_projection_not_completed` | `graph_projection_unavailable` |
| `graph_projection_selection_unresolved` | `graph_projection_unavailable` |
| `graph_projection_digest_mismatch` | `graph_projection_unavailable` |
| `diagram_hard_limit_exceeded` | `mermaid_invalid` |
| `invalid_mermaid_construct` | `mermaid_invalid` |
| `label_length_exceeded` | `mermaid_invalid` |
| `slide_count_exceeded` | `slidev_source_invalid` |
| `slide_block_limit_exceeded` | `slidev_source_invalid` |
| `speaker_notes_limit_exceeded` | `slidev_source_invalid` |
| `click_step_limit_exceeded` | `slidev_source_invalid` |
| `click_export_page_count_mismatch` | `slidev_export_failed` |
| `invalid_template_binding_pattern` | `bundle_manifest_invalid` |
| `asset_limit_exceeded` | `asset_limit_exceeded` |
| `outbound_request_observed` | `remote_asset_reference` |
| `render_canceled` | `render_canceled` |
| `render_timeout` | `render_timeout` |

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
| Failure after partial files | Delete partial bundle or quarantine it as diagnostic-only outside the release bundle; never expose it as release output. |
| Cancellation after release row allocation | Release becomes `render_failed` with `reason_code='render_canceled'`. |
| Timeout | Release becomes `render_failed` with `reason_code='render_timeout'`; partial output is deleted or diagnostic-only. |
| Exact retry | Core idempotency returns original job or release when Core classifies it as exact replay. |
| New retry | Requires a new `client_txn_id` and new release candidate when Core permits. |

## 24.2 Timeouts

**REQ-RPT-118**
Render timeouts MUST use Table 24-B unless a template declares stricter values. Templates MUST NOT widen hard max values.

**Table 24-B. Render attempt timeouts**

| Stage | Default | Hard max |
| --- | ---: | ---: |
| Export-model materialization | 120s | 300s |
| Redaction and tokenization | 120s | 300s |
| Mermaid source/render per diagram | 60s | 180s |
| Slidev export | 600s | 900s |
| Determinism rerender | Same as first render | Same as first render |
| Total attempt | 900s | 1800s |

## 24.3 Render sandbox policy

**REQ-RPT-119**
`render_sandbox_policy_v1` MUST enforce Table 24-C. The mechanism MAY vary only when the observable evidence and failure behavior remain identical.

**Table 24-C. Render sandbox policy**

| Boundary | Rule |
| --- | --- |
| Network | Only loopback connections required to drive the local renderer are allowed; non-loopback DNS, TCP, UDP, and HTTP(S) attempts fail. |
| Package resolution | No package install or remote resolution during release rendering. |
| Filesystem read | Only render working directory, declared template pack directory, pinned package store, and declared local assets. |
| Filesystem write | Only isolated render working directory and diagnostic directory. |
| Environment | Only `env_allowlist[]`. |
| URL schemes | Relative bundle paths only. `http:`, `https:`, protocol-relative URLs, external `file:`, `data:`, and `javascript:` are invalid. |
| Browser content | No iframes, remote embeds, scripts, arbitrary Vue, Monaco, Twoslash, code runners, or external fetch behavior. |
| Evidence | Raw evidence bytes are not mounted unless a redacted, declared local asset has already been produced. |
| Observation | Validation summary records `network_policy_id`, `non_loopback_outbound_attempt_count`, and first blocked destination class, not raw URLs. |

## 24.4 Render algorithms

**REQ-RPT-120**
`render_mermaid_bundle_v1` MUST execute these steps in order: validate release tuple and output options; materialize export model; apply redaction and token manifests; resolve diagrams; serialize `.mmd` bytes under §16; render SVG or PNG according to §7.5 and §22; validate sandbox evidence; build bundle manifest; run determinism validation when required; persist bundle only after validation passes.

**REQ-RPT-121**
`render_slidev_bundle_v1` MUST execute these steps in order: validate release tuple and output options; materialize export model; apply redaction and token manifests; resolve diagrams and Mermaid sources; generate deck model; serialize `slides.md` under §18; render PDF/PPTX/PNG according to §7.5 and §22; validate click page count; validate sandbox evidence; build bundle manifest; run determinism validation when required; persist bundle only after validation passes.

**REQ-RPT-122**
For external release, the same tuple rendered twice in clean working directories under the same toolchain snapshot MUST produce byte-identical canonical export model, redaction manifest, token manifest, toolchain snapshot, validation summary, render-bundle manifest, and `output_sha256`. A mismatch MUST fail with `failure_code='nondeterministic_render'` and a reason code identifying the first mismatching artifact class.

# 25. Resource limit registry

**REQ-RPT-123**
Reporting resource limits MUST use Table 25-A. A template pack MAY declare stricter values. A template pack MUST NOT widen a hard limit.

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
| `slides.count` | 120 | 240 | `slidev_source_invalid` | `slide_count_exceeded` |
| `blocks_per_slide` | 30 | 60 | `slidev_source_invalid` | `slide_block_limit_exceeded` |
| `speaker_notes_chars_per_slide` | 10,000 | 25,000 | `slidev_source_invalid` | `speaker_notes_limit_exceeded` |
| `click_steps_per_slide` | 20 | 40 | `slidev_source_invalid` | `click_step_limit_exceeded` |
| `local_assets.count` | 250 | 500 | `asset_limit_exceeded` | `local_asset_count_exceeded` |
| `local_asset.byte_size` | 25 MiB | 50 MiB | `asset_limit_exceeded` | `local_asset_bytes_exceeded` |
| `local_assets.total_bytes` | 100 MiB | 250 MiB | `asset_limit_exceeded` | `local_assets_total_bytes_exceeded` |
| `rendered_pdf.byte_size` | 250 MiB | 500 MiB | `bundle_resource_limit_exceeded` | `rendered_pdf_bytes_exceeded` |
| `validation_issues.count` | 500 retained | 2,000 evaluated | `validation_summary_truncated` warning | `validation_issue_limit_reached` |

**REQ-RPT-124**
The implementation MUST evaluate every applicable resource limit before persisting external release output. When multiple limits fail, issue ordering in §23.4 determines `first_failure`.

# 26. Conformance fixtures

**REQ-RPT-125**
A conforming implementation MUST provide fixtures in Table 26-A. Fixture IDs are stable and MUST NOT be reused for different behavior.

**Table 26-A. Required reporting fixtures**

| Fixture ID | Fixture | Required result |
| --- | --- | --- |
| `RPT-FIX-001` | Minimal external Mermaid release with no tokens. | Passes and emits `.mmd`, SVG, validation, manifests, and deterministic `output_sha256`. |
| `RPT-FIX-002` | Minimal external Slidev release with one diagram. | Passes and emits `slides.md`, PDF, `.mmd`, SVG, validation, manifests, and deterministic `output_sha256`. |
| `RPT-FIX-003` | Request `markdown`, `html`, or `reenactment`. | Fails with `invalid_release_request`, `unsupported_output_kind`. |
| `RPT-FIX-004` | Same tuple rendered with different system clocks. | Canonical bytes and `output_sha256` identical. |
| `RPT-FIX-005` | Duplicate active party assignments. | Fails with `duplicate_active_party_assignment`. |
| `RPT-FIX-006` | Aggregate with bucket contributor count `1` or `2`. | Not public; fails or remains partitioned according to §12.4. |
| `RPT-FIX-007` | Mixed paragraph requiring NLP split. | Does not split; redacts, drops, or fails. |
| `RPT-FIX-008` | Unstable unresolved subject without mention ID. | Fails with `token_subject_not_stable`. |
| `RPT-FIX-009` | Reveal map present in external bundle. | Bundle validation fails. |
| `RPT-FIX-010` | Stale Graph Projection snapshot. | Fails with `graph_projection_stale`. |
| `RPT-FIX-011` | Mermaid golden source. | `.mmd` bytes match exactly. |
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
| `RPT-FIX-025` | Generated text containing Markdown metacharacters inline and at line start. | `slides.md` bytes match the `slidev_markdown_escape_v1` golden exactly. |
| `RPT-FIX-026` | Mermaid label containing `<` or `>`. | Fails with `mermaid_invalid`, `invalid_mermaid_construct`. |
| `RPT-FIX-027` | Label of 121–240 scalars with and without template truncation, and a label over 240. | With truncation, exactly 120 scalars ending in U+2026; without, fails with `label_length_exceeded`; over 240 fails. |
| `RPT-FIX-028` | Export model with empty and populated nested validation objects. | `validation_summary`, `content_class_summary`, `section_validation`, and `display_times` serialize full fixed key sets; canonical bytes stable. |
| `RPT-FIX-029` | Timeline exceeding 240 rows with no template `selection_rule_id`. | Internal scope emits the overflow block; `external_release` fails with `timeline_overflow_unresolved`. |
| `RPT-FIX-030` | Internal `source_only=true` deck. | Explicit `pdf=true` fails with `source_only_conflict`; defaults suppress all rendered-output roles and emit a source-only bundle. |
| `RPT-FIX-031` | `snapshot_at='2026-02-30T00:00:00Z'`. | Fails with `invalid_timestamp_value`. |
| `RPT-FIX-032` | Font and package sets enumerated in shuffled filesystem order. | `content_manifest_digest_v1` digests are identical across orderings. |
| `RPT-FIX-033` | Export model with more than 100 diagrams. | Fails with `diagrams_count_exceeded`. |
| `RPT-FIX-034` | Snapshot with `deleted` and `superseded` records. | `deleted` never appears in rendered bytes; `superseded` is excluded from external release unless the template opts in. |
| `RPT-FIX-035` | Aggregate-public block with default day bucketing. | Times bucket to `YYYY-MM-DD` deterministically; unresolved rows map to the `unresolved` bucket. |
| `RPT-FIX-036` | Two decks rendered from the same deck model. | Headmatter emits the full Table 18-C key set with fixed values; `slides.md` bytes match the golden. |

# 27. Acceptance criteria and traceability

## 27.1 Requirement-to-acceptance rule

**REQ-RPT-126**
Every `REQ-RPT-*` requirement in this NLSpec MUST map to at least one acceptance criterion in §27.2 or to one fixture in §26. A requirement without acceptance coverage is incomplete and cannot support adoption.

## 27.2 Acceptance criteria

**REQ-RPT-127**
A conforming implementation MUST satisfy Table 27-A.

**Table 27-A. Acceptance criteria**

| Acceptance ID | Pass condition |
| --- | --- |
| `RPT-AC-AUTH-001` | A conformance claim fails when this NLSpec is marked adopted/current while any Table 1-A promotion condition remains unmet or future-only. |
| `RPT-AC-LINT-001` | Adopted text contains no open delegation phrase outside rationale, examples, source notes, future-only material, or an adjacent closed owner table. |
| `RPT-AC-CORE-001` | Every Core dependency row has owner, classification, affected Reporting requirements, and conformance consequence. |
| `RPT-AC-ID-001` | Every schema or algorithm identifier listed in §6.2 has a body section defining required fields, nullability, defaults, ordering, failure behavior, and acceptance coverage. |
| `RPT-AC-KIND-001` | `markdown`, `html`, `reenactment`, unknown strings, aliases, and case variants are rejected with `unsupported_output_kind`. |
| `RPT-AC-OPT-001` | Omitted `output_options` materializes to exact §7.5 defaults and canonical bytes match explicit defaults. |
| `RPT-AC-OPT-002` | Every invalid option, scope, or template combination fails with the mapped structured error before output bytes persist. |
| `RPT-AC-SCHEMA-001` | Every top-level export-model array item validates against its named schema and rejects unknown members, invalid nulls, omitted required fields, duplicate IDs, and ordering drift. |
| `RPT-AC-SCHEMA-002` | Explicit defaults and omitted defaults canonicalize identically where omission is valid. |
| `RPT-AC-TIME-001` | Rerendering the same tuple with different system clocks produces byte-identical canonical export model, validation summary, redaction manifest, token manifest, toolchain snapshot, bundle manifest, and `output_sha256`. |
| `RPT-AC-MAT-001` | Export-model validation emits the expected structured failure for duplicate IDs, dangling refs, missing content class, invalid generated IDs, and resource-limit failures. |
| `RPT-AC-PART-001` | Duplicate active party assignments fail deterministically. |
| `RPT-AC-PART-002` | Aggregate-derived public blocks require `aggregate_public_v1` proof; low-count buckets remain partitioned or fail. |
| `RPT-AC-PART-003` | Mixed bullet, table, metric, timeline, diagram-label, and child-field blocks split deterministically; free-form paragraph mixed content does not split by NLP. |
| `RPT-AC-TOKEN-001` | Same tuple yields identical tokens and token manifests. |
| `RPT-AC-TOKEN-002` | Unstable unresolved subjects fail before bundle persistence. |
| `RPT-AC-REVEAL-001` | External bundles exclude reveal maps; internal sensitive artifacts contain complete reveal entries and require Core 04 authorization. |
| `RPT-AC-TIMEORDER-001` | Null, partial, ambiguous, duplicate, and unparseable timeline values sort exactly by §14.2 and preserve original display text subject to redaction. |
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
| `RPT-AC-ARCHIVE-001` | Mutating archive metadata alone does not alter `output_sha256`; mutating a child file does. |
| `RPT-AC-ERR-001` | Every validation failure emits exactly one public `error.code`, one `failure_code` when admitted, one `reason_code` when attributable, and safe details restricted to the allow-list. |
| `RPT-AC-LIFE-001` | Cancellation, timeout, partial output, exact replay, and new retry have deterministic job and release outcomes. |
| `RPT-AC-SANDBOX-001` | Non-loopback outbound attempt fails before persistence and safe details do not include raw external URLs. |
| `RPT-AC-LIMIT-001` | Every limit key enforces default, stricter template value, hard max, and exact failure mapping. |
| `RPT-AC-DERIVE-001` | `derivation_version` resolves to an adopted `cartulary.reporting_derivation_profile.v1`; an unresolved profile blocks external release with `blocked_core_dependency`, and every Table 7-A1 obligation is closed. |
| `RPT-AC-OPT-003` | Internal `source_only=true` forces every rendered-output member to `false`; an explicit rendered-output `true` fails with `source_only_conflict`. |
| `RPT-AC-ORDER-001` | `derive_section_ordering_key_v1` yields identical `ordering_key` values and section order for the same template manifest and content. |
| `RPT-AC-PART-004` | `deleted` records never appear in rendered bytes; `superseded` records are excluded from external release unless the template opts in; forced inclusion fails with `deleted_record_not_releasable`. |
| `RPT-AC-SCHEMA-003` | `validation_summary`, `content_class_summary`, `section_validation`, and `display_times` always serialize their full closed member sets and canonicalize deterministically. |
| `RPT-AC-TOKEN-003` | `derive_display_token_v1` produces grammar-valid, family-scoped tokens assigned in `stable_subject_ref` order, byte-identical across reruns, with no reversible material in the visible token. |
| `RPT-AC-TIMEORDER-002` | Timeline overflow emits the overflow block for internal scopes and fails external release with `timeline_overflow_unresolved` when no summarizing `selection_rule_id` is declared. |
| `RPT-AC-MMD-003` | Labels over the default truncate deterministically to 120 scalars ending in U+2026 only when declared, else fail with `label_length_exceeded`; labels containing `<` or `>` fail with `invalid_mermaid_construct`. |
| `RPT-AC-SLIDEV-002` | `slidev_markdown_escape_v1` escapes generated text to the closed Table 18-B set and headmatter emits the full Table 18-C key set with fixed values; `slides.md` bytes match golden. |
| `RPT-AC-TOOLCHAIN-002` | `content_manifest_digest_v1` produces identical `font_manifest_sha256` and `package_store_digest` regardless of file enumeration order and excludes mtime, ownership, and permission bits. |
| `RPT-AC-TIME-002` | Invalid calendar timestamps fail with `invalid_timestamp_value`; aggregate-public bucketed times are deterministic per Table 12-F. |

# 28. Future-only areas

**REQ-RPT-128**
The areas in Table 28-A are future-only. A conforming v1 implementation MUST use the current omission behavior in the table.

**Table 28-A. Future-only areas**

| Future-only area | Current omission behavior |
| --- | --- |
| Markdown report output | `output_kind='markdown'` fails with `unsupported_output_kind`. |
| Static HTML report output | `output_kind='html'` fails with `unsupported_output_kind`. |
| Reenactment output kind | `output_kind='reenactment'` fails with `unsupported_output_kind`. |
| Graph chunking | Diagrams beyond hard Mermaid bounds fail; templates must select smaller diagrams. |
| Source-only external release | Fails with `source_only_external_release_invalid`. |
| Arbitrary Slidev parser or pre-parser extensions | Invalid for external release. |
| Custom Slidev components beyond reveal-only subset | Invalid for external release. |
| Canonical archive bytes | Approval and publication bind to manifest hash, not ZIP/TAR bytes. |
| External toolchain auto-update or remote package resolution | Invalid during render. |
| Deployment-configurable Reporting resource limits | Templates MAY narrow limits only; deployments MUST NOT widen them in v1. |
| External reveal-map sharing | Reveal maps remain internal sensitive release artifacts. |
| New `post` record family | Invalid unless a later Core owner defines it. |

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
| Partition closure | Assignment lifecycle, duplicate active assignments, public aggregate proof, and mixed-content splitting are deterministic. |
| Timeline closure | Sort materialization, precision ranks, unresolved-time ordering, bounds, and overflow summary are specified. |
| Graph closure | Reporting consumes only completed Graph Projection output through `source_projection_ref.v1` and closed selection rules. |
| Mermaid closure | `.mmd` bytes are deterministically serializable from diagram objects. |
| Slidev closure | `slides.md` bytes are deterministically serializable from deck objects. |
| Click closure | Click-step schema and page-count rule are closed. |
| Toolchain closure | Render environment fields are sufficient to reproduce rendered outputs inside the declared environment. |
| Bundle closure | Role/path/media matrix is exhaustive; archive wrapper semantics are explicit. |
| Sandbox closure | Network, filesystem, environment, URL scheme, and evidence-mount boundaries are explicit and testable. |
| Lifecycle closure | Success, failure, cancellation, timeout, partial output, retry, and Core release-state effects are explicit. |
| Limit closure | Every resource limit has default, hard limit, stage, failure code, and overflow or fail behavior. |
| Token-string closure | `display_token` is deterministically and non-reversibly derived by `derive_display_token_v1`. |
| Ordering-key closure | Section `ordering_key` is deterministically derived by `derive_section_ordering_key_v1`. |
| Nested-object closure | Export-model `validation_summary`, `content_class_summary`, `section_validation`, and `display_times` are closed with fixed member sets. |
| Escaping closure | Mermaid and Slidev generated text have closed escape sets, deterministic label truncation, and no unnamed raw-HTML detector. |
| Digest closure | Multi-file digests use `content_manifest_digest_v1` and are enumeration-order-independent and filesystem-metadata-free. |
| Derivation-profile closure | `derivation_version` resolves to an adopted derivation profile that deterministically closes snapshot-to-export-model content derivation. |
| Acceptance traceability | Every `REQ-RPT-*` maps to at least one `RPT-AC-*` or fixture. |
