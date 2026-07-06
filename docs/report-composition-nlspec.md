---
title: Cartulary Report Composition NLSpec
status: draft
document_class: nlspec
profile: snapshot_reporting
schema_id: cartulary.report_composition_nlspec.v1
---

# 1. Status, Scope, And Authority

Status: `draft`.

This NLSpec defines the Cartulary Report Composition companion subsystem for the Snapshot and Reporting Extension Profile. It becomes implementation-conformance authority only after promotion to `status: adopted/current` and after the required Core and Reporting companion amendments named by this document are adopted.

**REQ-RC-001**
This NLSpec owns only the authoring-side composition boundary:

- `cartulary.report_composition.v1` canonical composition bytes;
- composition resource lifecycle, draft mutation, immutable versioning, and freeze behavior;
- composition authoring route family contracts;
- `composition_op.v1` operation schemas;
- semantic composition anchor grammars;
- `authored_text.v1` presentation-text objects;
- `composition_diagram_decl.v1` declaration objects;
- builder-facing validation summaries, issue codes, and safe detail keys;
- report builder UI conformance boundaries.

**REQ-RC-002**
This NLSpec MUST NOT own report materialization, release tuple admission, redaction, token substitution, Mermaid source generation, Slidev source generation, deck derivation effects, render sandboxing, render bundle hashing, release approval, release publication, or Reporting fixture bytes. Those behaviors remain owned by Core 01, Core 04, and `docs/reporting-subsystem-nlspec.md`.

**REQ-RC-003**
This NLSpec is a companion to `docs/reporting-subsystem-nlspec.md`. Reporting consumes immutable composition versions by `composition_id`, `composition_version`, and `composition_sha256`. Reporting owns the observable render effects of a valid composition after Core route admission accepts a render attempt.

**REQ-RC-004**
When this NLSpec conflicts with Core 00 through Core 04 outside the composition subsystem, the conflict is a defect in this NLSpec. When this NLSpec conflicts with `docs/reporting-subsystem-nlspec.md`, the owner boundary in Table 1-A MUST decide the defect location.

**Table 1-A. Owner Boundary**

| Area | Owner | Boundary rule |
| --- | --- | --- |
| Public route envelopes, common error envelope, idempotency mechanics, job resource conventions | Core 01 | This NLSpec names route-specific request and response members only. |
| Incident roles, authorization derivation, deployment-admin limits, release approval | Core 04 | This NLSpec imports role names and MUST NOT create new incident ACL machinery. |
| Composition draft, version, schema, operation vocabulary, anchors, authored text, diagram declaration authoring | This NLSpec | Reporting imports these identifiers but MUST NOT redefine their schema. |
| Composition render effect, redaction admission, `derive_deck_v2`, diagrams in render output, fixture bytes | Reporting Subsystem NLSpec | This NLSpec names the intended target and payload only. |
| Graph selection validation and graph projection lifecycle | Graph Projection NLSpec and Reporting Subsystem NLSpec | This NLSpec admits diagram declarations only through closed selection-rule objects. |
| Domain vocabulary | `docs/domain.md` | This NLSpec may introduce composition terms only after domain vocabulary rows are added or updated. |

**REQ-RC-005**
Raw Markdown, raw Mermaid, raw HTML, generated `slides.md`, generated `.mmd`, rendered PDF, rendered PPTX, rendered image output, bundle manifests, token manifests, redaction manifests, post-redaction bytes, workbook records, graph-projection output, snapshots, and template packs are not composition source. A conforming implementation MUST reject attempts to author report composition by editing any of those objects.

**REQ-RC-006**
Cross-template composition portability, collaborative composition editing, post-redaction editing, literal-sensitive-value scanning inside authored free text, arbitrary diagram node or edge creation, UI library selection, React component structure, editor implementation technology, and builder layout styling are future-only or non-normative in this revision.

# 2. Normative Language And Document Discipline

**REQ-RC-007**
The key words **MUST**, **MUST NOT**, and **MAY** are normative in this NLSpec. **MUST** and **MUST NOT** define conformance requirements. **MAY** defines optional behavior only when omission behavior is explicit in the same requirement, table row, or immediately following paragraph.

**REQ-RC-008**
The word `default` defines the required value or behavior used when an optional member is omitted and omission is valid.

**REQ-RC-009**
A conforming implementation MUST treat object member names, schema identifiers, route path segments, operation kinds, anchor kinds, validation codes, reason codes, role tokens, and closed-vocabulary values as exact Unicode code point sequences. It MUST NOT apply case folding, trimming, locale comparison, or Unicode normalization unless this NLSpec names the exact operation.

**REQ-RC-010**
Accepted normative prose MUST NOT use open delegation phrases. A delegation is valid only when it names the exact owner document, schema, algorithm, route convention, or future-only boundary that closes the behavior.

**REQ-RC-011**
A defaulted member omitted from a route request MUST be materialized into draft state before validation. Immutable version bytes MUST include the materialized default value.

# 3. Purpose And Non-Goals

**REQ-RC-012**
Report Composition MUST let authorized incident users customize the presentation shape of a report without changing incident source records, snapshots, report templates, generated report bytes, or release approval semantics.

**REQ-RC-013**
A composition MUST be data, not generated source text. The only conforming composition input to Reporting is an immutable `cartulary.report_composition.v1` document whose digest matches the Core release tuple.

**REQ-RC-014**
A composition MUST preserve the evidence-versus-presentation boundary. Authored composition text is presentation-tier text only and MUST enter Reporting only through the authored-text roles defined in §9.

**REQ-RC-015**
This NLSpec MUST NOT introduce the non-goals in Table 3-A.

**Table 3-A. Non-Goals**

| Non-goal | Required omission behavior |
| --- | --- |
| Raw generated-source editing | Generated `slides.md`, `.mmd`, HTML, PDF, PPTX, image output, and manifests are not composition inputs. |
| Workbook mutation from the builder | Composition authoring MUST NOT create, update, delete, reorder, or annotate workbook source records. |
| Template mutation from the builder | Composition authoring MUST NOT alter template packs, template manifests, layout declarations, or template section declarations. |
| Snapshot mutation from the builder | Composition authoring MUST NOT alter snapshot state or snapshot materialization. |
| Graph mutation from the builder | Composition authoring MUST NOT create graph-projection vertices, edges, or relationships. |
| Free-text fact narrative | Case facts, findings, causal claims, actors, timestamps, commands, evidence assertions, or conclusions belong in source records and template narrative slots, not composition-authored text. |
| Cross-template migration | A composition binds to exactly one `template_id` and `template_version`; migration to another template is future-only. |
| Collaborative editing | Multi-user real-time editing, merge cursors, and WebSocket collaboration are future-only. |
| New live ACL model | Composition access uses existing incident roles. It MUST NOT create record-level, field-level, recipient-level, or composition-specific live ACLs. |
| UI implementation prescription | Component libraries, editor internals, CSS, and visual design are outside this NLSpec unless they affect the emitted composition contract. |

# 4. Concepts And Identifiers

**REQ-RC-016**
The terms in Table 4-A have the meanings defined here inside this NLSpec.

**Table 4-A. Concepts**

| Term | Definition |
| --- | --- |
| `report composition` | Incident-scoped authoring input that carries presentation operations, authored presentation text, and composition diagram declarations for one exact report template version. |
| `composition resource` | Mutable incident-scoped resource that owns one current draft and zero or more immutable composition versions. |
| `composition draft` | Mutable server state for a composition resource before version freeze. |
| `composition version` | Immutable canonical `cartulary.report_composition.v1` document created from a draft. |
| `release-bound composition version` | Immutable composition version referenced by a release tuple. |
| `composition operation` | Closed `composition_op.v1` object that names one presentation edit by semantic anchors and payload fields. |
| `semantic composition anchor` | Stable authoring target reference that resolves by template declaration identity, source record identity, block context, or diagram declaration identity. |
| `authored presentation text` | Composition-authored text admitted only as a title override, speaker notes, or a paragraph block after partition and redaction checks. |
| `composition diagram declaration` | Composition-owned diagram declaration that uses Reporting diagram selection rules and no raw Mermaid. |
| `report builder` | UI surface that emits composition resources and requests validation or preview through this NLSpec's route family. |
| `authoritative preview` | Reporting-owned `internal_draft` render attempt using a composition draft or immutable version. |

**REQ-RC-017**
This revision defines the identifiers in Table 4-B. Each identifier MUST have the owner and use shown in the table.

**Table 4-B. Closed Identifiers**

| Identifier | Kind | Owner section | Required use |
| --- | --- | --- | --- |
| `cartulary.report_composition.v1` | canonical schema | §7 | Immutable digest-bound composition document. |
| `cartulary.report_composition_resource_view.v1` | route schema | §6 | Server view of a composition resource and current draft metadata. |
| `cartulary.report_composition_version_view.v1` | route schema | §6 | Server view of one immutable composition version. |
| `cartulary.report_composition_validate_request.v1` | route schema | §12 | Server validation request. |
| `cartulary.report_composition_validation_summary.v1` | route schema | §12 | Server validation result. |
| `composition_issue.v1` | route schema | §12 | One validation issue. |
| `composition_op.v1` | composition schema | §11 | Closed operation object. |
| `section_anchor` | anchor grammar | §8 | Section target. |
| `record_anchor` | anchor grammar | §8 | Source record target. |
| `block_anchor` | anchor grammar | §8 | Derived block target. |
| `diagram_anchor` | anchor grammar | §8 | Diagram declaration target. |
| `authored_text.v1` | composition schema | §9 | Authored presentation text object. |
| `composition_diagram_decl.v1` | composition schema | §10 | Composition-owned diagram declaration. |
| `create_composition_draft_v1` | route operation | §6 | Draft resource creation. |
| `update_composition_draft_v1` | route operation | §6 | Draft mutation. |
| `freeze_composition_version_v1` | route operation | §6 | Immutable version creation. |

**REQ-RC-018**
The identifier `composition_version` MUST match `v[1-9][0-9]*`. The first immutable version for a composition resource is `v1`. Later immutable versions for the same resource MUST increment the decimal suffix by one with no gap after a successful version creation. The literal `latest` is invalid in public requests and release tuples.

**REQ-RC-019**
The identifier `composition_id` is opaque, server-assigned, and immutable. Clients MUST NOT parse it or derive incident, template, or version information from it.

# 5. Lifecycle

**REQ-RC-020**
A composition resource MUST be scoped to exactly one `incident_id`, exactly one `template_id`, and exactly one `template_version`. These three values are fixed at resource creation and MUST NOT be changed by draft update, version creation, validation, preview, or release binding.

**REQ-RC-021**
Every composition resource MUST have at most one active draft. A draft is mutable until the resource is retired. A draft update MUST increment `draft_version` exactly once on success.

**REQ-RC-022**
Every mutable draft MUST carry a server-emitted `draft_version` positive integer. Mutating requests MUST carry `base_draft_version`. If `base_draft_version` does not equal the current server `draft_version`, the request MUST fail with `error.code='conflict'` and validation code `composition_draft_version_conflict`.

**REQ-RC-023**
Creating an immutable composition version MUST copy the current draft body, materialize all defaults, compute canonical bytes, compute `composition_sha256`, assign the next `composition_version`, and persist the resulting `cartulary.report_composition.v1` document atomically.

**REQ-RC-024**
Immutable composition versions MUST NOT be modified, reserialized with different canonical bytes, deleted while release-bound, or assigned a new digest. A request that attempts to mutate an immutable version MUST fail with validation code `composition_version_immutable`.

**REQ-RC-025**
When a release tuple binds `composition_id`, `composition_version`, and `composition_sha256`, the referenced immutable version is release-bound. Later draft edits and later immutable versions MUST NOT change the bound version or its digest.

**REQ-RC-026**
A release-bound composition version MUST NOT be deleted or rewritten. A request that attempts to delete or rewrite a release-bound version MUST fail with validation code `composition_version_bound`.

**REQ-RC-027**
Retiring a composition resource is permitted only when no immutable version under that resource is release-bound. Retirement MUST preserve immutable versions for audit and digest validation. Omission behavior: if retirement is not implemented, all delete requests MUST fail with validation code `composition_delete_not_supported`.

**REQ-RC-028**
The nullable `authored_against_snapshot_id` is advisory authoring context only. It MUST NOT bind the composition to a snapshot, MUST NOT replace release tuple snapshot binding, and MUST NOT make a composition valid for a different `template_id` or `template_version`.

# 6. Public Route Family

**REQ-RC-029**
The route family root MUST be `/api/v1/incidents/{incident_id}/report-compositions`. All routes MUST use Core 01 success and error envelopes. Mutating routes MUST use Core 01 idempotency behavior keyed by `client_txn_id`.

**REQ-RC-030**
All route authorization MUST be rederived at route time from current incident membership and Core 04 incident role rules. The `deployment_admin` capability alone MUST NOT grant access to any composition route.

**REQ-RC-031**
Routes MUST use the minimum incident roles in Table 6-A.

**Table 6-A. Route Inventory**

| Method and path | Operation | Minimum role | Idempotency | Request body | Success body |
| --- | --- | --- | --- | --- | --- |
| `GET /report-compositions` | List composition resources for the incident. | `viewer` | No | None | `composition_resources[]` |
| `POST /report-compositions` | Create draft resource. | `editor` | `client_txn_id` | Table 6-B create body | `cartulary.report_composition_resource_view.v1` |
| `GET /report-compositions/{composition_id}` | Read resource and current draft metadata. | `viewer` | No | None | `cartulary.report_composition_resource_view.v1` |
| `PATCH /report-compositions/{composition_id}` | Update active draft. | `editor` | `client_txn_id` | Table 6-C update body | `cartulary.report_composition_resource_view.v1` |
| `DELETE /report-compositions/{composition_id}` | Retire resource when allowed. | `editor` | `client_txn_id` | Table 6-D delete body | `cartulary.report_composition_resource_view.v1` |
| `POST /report-compositions/{composition_id}/versions` | Freeze immutable version. | `editor` | `client_txn_id` | Table 6-E freeze body | `cartulary.report_composition_version_view.v1` |
| `GET /report-compositions/{composition_id}/versions/{composition_version}` | Read immutable version. | `viewer` | No | None | `cartulary.report_composition_version_view.v1` |
| `POST /report-compositions/{composition_id}/validate` | Validate draft or version. | `viewer` | No | `cartulary.report_composition_validate_request.v1` | `cartulary.report_composition_validation_summary.v1` |
| `POST /report-compositions/{composition_id}/preview` | Request authoritative `internal_draft` preview through Reporting. | `viewer` | `client_txn_id` | Table 6-F preview body | Core-owned job or render-attempt reference |

**REQ-RC-032**
A route path `incident_id` MUST equal the resource `incident_id`. A body `incident_id`, when present, MUST equal the route path `incident_id`. A mismatch MUST fail with `error.code='invalid_request'` and validation code `composition_incident_mismatch`.

**REQ-RC-033**
A route path `composition_id` MUST identify a composition resource inside the route path incident. No-match requests MUST fail with `error.code='not_found'` and validation code `composition_not_found`.

**Table 6-B. Create Draft Body**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `client_txn_id` | string | Yes | No | None | Core 01 idempotency key. |
| `template_id` | identifier | Yes | No | None | Fixed for resource lifetime. |
| `template_version` | identifier | Yes | No | None | Fixed for resource lifetime. |
| `authored_against_snapshot_id` | identifier | No | Yes | `null` | Advisory only. |
| `deck_ops` | array of `composition_op.v1` | No | No | `[]` | Draft operation order. |
| `diagram_decls` | array of `composition_diagram_decl.v1` | No | No | `[]` | Draft diagram declarations. |
| `authored_texts` | array of `authored_text.v1` | No | No | `[]` | Draft authored text objects. |

**Table 6-C. Update Draft Body**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `client_txn_id` | string | Yes | No | None | Core 01 idempotency key. |
| `base_draft_version` | positive integer | Yes | No | None | Must equal current `draft_version`. |
| `authored_against_snapshot_id` | identifier | No | Yes | no change | Advisory only. |
| `deck_ops` | array of `composition_op.v1` | No | No | no change | Replaces the full draft array when present. |
| `diagram_decls` | array of `composition_diagram_decl.v1` | No | No | no change | Replaces the full draft array when present. |
| `authored_texts` | array of `authored_text.v1` | No | No | no change | Replaces the full draft array when present. |

**Table 6-D. Delete Body**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `client_txn_id` | string | Yes | No | None | Core 01 idempotency key. |
| `base_draft_version` | positive integer | Yes | No | None | Must equal current `draft_version` unless the resource is already retired by the same idempotent request. |

**Table 6-E. Freeze Version Body**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `client_txn_id` | string | Yes | No | None | Core 01 idempotency key. |
| `base_draft_version` | positive integer | Yes | No | None | Must equal current `draft_version`. |

**Table 6-F. Preview Body**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `client_txn_id` | string | Yes | No | None | Core 01 idempotency key. |
| `source_kind` | string | No | No | `draft` | Closed values: `draft`, `version`. |
| `composition_version` | identifier | Required when `source_kind='version'` | No | None | Must name an immutable version under the resource. |
| `snapshot_id` | identifier | Yes | No | None | Passed to Reporting as preview input. |
| `template_id` | identifier | Yes | No | None | Must equal resource `template_id`. |
| `template_version` | identifier | Yes | No | None | Must equal resource `template_version`. |
| `redaction_profile_id` | identifier | Yes | No | None | Passed to Reporting as preview input. |
| `recipient_partition_refs` | array of string | No | No | `[]` | Passed to Reporting as preview input. |

**REQ-RC-034**
The success schema `cartulary.report_composition_resource_view.v1` MUST contain the members in Table 6-G and MUST NOT contain unknown members.

**Table 6-G. Resource View**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `composition_id` | identifier | Yes | No | Resource identity. |
| `incident_id` | identifier | Yes | No | Resource incident. |
| `template_id` | identifier | Yes | No | Fixed template binding. |
| `template_version` | identifier | Yes | No | Fixed template version binding. |
| `draft_version` | positive integer | Yes | No | Current mutable draft version. |
| `authored_against_snapshot_id` | identifier | Yes | Yes | Advisory context. |
| `deck_ops` | array of `composition_op.v1` | Yes | No | Current draft operations. |
| `diagram_decls` | array of `composition_diagram_decl.v1` | Yes | No | Current draft declarations. |
| `authored_texts` | array of `authored_text.v1` | Yes | No | Current draft authored text. |
| `latest_composition_version` | identifier | Yes | Yes | Highest immutable version or `null`. |
| `release_bound_versions` | array of identifier | Yes | No | Immutable versions known to be release-bound. |
| `retired_at` | timestamp | Yes | Yes | `null` when active. |

**REQ-RC-035**
The success schema `cartulary.report_composition_version_view.v1` MUST contain the members in Table 6-H and MUST NOT contain unknown members.

**Table 6-H. Version View**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `composition_id` | identifier | Yes | No | Resource identity. |
| `composition_version` | identifier | Yes | No | Immutable version identifier. |
| `composition_sha256` | sha256_hex | Yes | No | Digest of canonical bytes excluding the digest member. |
| `canonical_composition` | `cartulary.report_composition.v1` | Yes | No | Immutable canonical document. |
| `created_at` | timestamp | Yes | No | Server version creation timestamp; excluded from canonical composition bytes. |
| `release_bound` | boolean | Yes | No | True when any release tuple references this exact version and digest. |

# 7. Canonical Composition Schema

**REQ-RC-036**
The canonical immutable composition document schema identifier MUST be `cartulary.report_composition.v1`.

**REQ-RC-037**
`cartulary.report_composition.v1` MUST contain exactly the members in Table 7-A. Unknown members are invalid. All required members MUST be present in immutable version bytes.

**Table 7-A. Canonical Composition Document**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.report_composition.v1`. |
| `composition_id` | identifier | Yes | No | None | Resource identity. |
| `composition_version` | identifier | Yes | No | None | Immutable version identifier from REQ-RC-018. |
| `incident_id` | identifier | Yes | No | None | Incident scope. |
| `template_id` | identifier | Yes | No | None | Fixed template binding. |
| `template_version` | identifier | Yes | No | None | Fixed template version binding. |
| `authored_against_snapshot_id` | identifier | Yes | Yes | `null` | Advisory only. |
| `deck_ops` | array of `composition_op.v1` | Yes | No | `[]` | Operation order is semantically significant. |
| `diagram_decls` | array of `composition_diagram_decl.v1` | Yes | No | `[]` | Ordered by input for canonical bytes; semantic identity is `decl_id`. |
| `authored_texts` | array of `authored_text.v1` | Yes | No | `[]` | Ordered by input for canonical bytes; semantic identity is `authored_text_id`. |
| `composition_sha256` | sha256_hex | Yes | No | None | Lowercase SHA-256 digest defined by REQ-RC-041. |

**REQ-RC-038**
The pairs in Table 7-B MUST be unique inside one canonical composition document.

**Table 7-B. Uniqueness Rules**

| Scope | Unique key | Failure code |
| --- | --- | --- |
| `deck_ops[]` | `op_id` | `composition_duplicate_id` |
| `diagram_decls[]` | `decl_id` | `composition_duplicate_id` |
| `authored_texts[]` | `authored_text_id` | `composition_duplicate_id` |

**REQ-RC-039**
A reference from a composition operation to `authored_text_ref` MUST resolve to exactly one `authored_texts[].authored_text_id` in the same composition document. A reference from a composition operation to a composition-owned `diagram_anchor` MUST resolve to exactly one `diagram_decls[].decl_id` unless Reporting resolution targets a template-owned diagram declaration. Unresolved local references MUST fail with `composition_schema_invalid`.

**REQ-RC-040**
Canonical JSON serialization for composition bytes MUST use `reporting_canonical_json_v1` as imported from the Reporting Subsystem NLSpec until Core adopts a shared canonical JSON owner. A composition document MUST NOT contain floating-point numbers. Object members MUST be serialized in lexicographic order by exact code point sequence. No insignificant whitespace is permitted outside strings.

**REQ-RC-041**
`composition_sha256` MUST be computed over the canonical JSON serialization of the composition document after deleting the `composition_sha256` member and leaving every other member unchanged. The stored digest MUST be exactly 64 lowercase hexadecimal characters.

**REQ-RC-042**
A server MUST verify the digest immediately after immutable version creation by recomputing REQ-RC-041 from persisted canonical bytes. A mismatch MUST fail version creation with validation code `composition_digest_mismatch` and MUST NOT persist the version.

# 8. Semantic Anchors

**REQ-RC-043**
Composition operations MUST target only semantic anchors defined in this section. Ordinal paths, generated section IDs, generated block IDs, generated slide IDs, chunk IDs, split IDs, array indexes, visible titles, visible labels, CSS selectors, and rendered DOM selectors are invalid composition targets.

**REQ-RC-044**
Every anchor object MUST carry `anchor_kind` with the exact value shown in Table 8-A and MUST NOT contain unknown members.

**Table 8-A. Anchor Schemas**

| Anchor kind | Required members | Nullable members | Required interpretation |
| --- | --- | --- | --- |
| `section_anchor` | `anchor_kind`, `template_section_decl_id`, `expansion_key` | `expansion_key` | Resolves to one emitted section instance from the template section declaration and expansion key. |
| `record_anchor` | `anchor_kind`, `record_id` | None | Resolves to one post-redaction reporting record summary by source `record_id`. |
| `block_anchor` | `anchor_kind`, `section_anchor`, `block_kind`, `record_anchor` | `record_anchor` | Resolves to one derived block by section, block kind, and optional source record. |
| `diagram_anchor` | `anchor_kind`, `decl_id` | None | Resolves to one template-owned or composition-owned diagram declaration. |

**REQ-RC-045**
For `section_anchor`, `template_section_decl_id` MUST be a template section declaration identifier. `expansion_key=null` targets a singleton emitted section. A non-null `expansion_key` MUST be the exact key emitted by the Reporting template section expansion algorithm for that declaration. Composition authoring MUST NOT invent or normalize expansion keys.

**REQ-RC-046**
For `record_anchor`, `record_id` MUST be a source record identifier. It MUST NOT be a projected row identifier, graph vertex identifier, subject token, visible row number, import unit identifier, or object storage identifier.

**REQ-RC-047**
For `block_anchor`, `block_kind` MUST be a Reporting deck block kind valid before chunking. A `record_anchor=null` block anchor is valid only when the section and block kind identify one block. If the target count is zero, the anchor is unresolved. If the target count is greater than one, the anchor is ambiguous.

**REQ-RC-048**
For `diagram_anchor`, `decl_id` MUST refer to a template-owned diagram declaration or a composition-owned `composition_diagram_decl.v1`. It MUST NOT refer to generated Mermaid node IDs, generated Mermaid edge IDs, SVG element IDs, or rendered image names.

**REQ-RC-049**
Each `composition_op.v1` MAY include `on_unresolved`. The default is `fail`. The only valid values are `fail` and `drop`. `drop` means the operation is omitted when its anchor target count is zero. `drop` MUST NOT suppress schema errors, ambiguous anchors, invalid references, invalid payload values, authored-text errors, or diagram declaration errors.

**REQ-RC-050**
`on_unresolved='drop'` is invalid for an `external_release` Reporting render attempt. Reporting MUST fail such a render with `composition_drop_invalid_for_external_release`. The authoring server MUST report the same issue when validation is requested with `release_scope='external_release'`.

# 9. Authored Presentation Text

**REQ-RC-051**
`authored_text.v1` MUST contain exactly the members in Table 9-A. Unknown members are invalid.

**Table 9-A. Authored Text Schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `authored_text_id` | identifier | Yes | No | None | Local identity referenced by operations. |
| `text_role` | string | Yes | No | None | Closed values in Table 9-B. |
| `body` | string | Yes | No | None | Presentation text subject to role limits. |
| `disclosure_partition_ref` | string | Yes | No | None | Author-declared disclosure partition; `blocked` is invalid. |

**Table 9-B. Authored Text Roles**

| `text_role` | Admission target | LF allowed | Default limit | Hard limit |
| --- | --- | ---: | ---: | ---: |
| `title_override` | Section or slide title. | No | 120 | 120 |
| `speaker_notes` | Slide speaker notes or `speaker_note` blocks. | Yes | Reporting `speaker_notes_chars_per_slide` | Reporting `speaker_notes_chars_per_slide` |
| `authored_text` | `paragraph` block with `content_class='presentation_text'`. | Yes | 2000 | 5000 |

**REQ-RC-052**
Every authored text object MUST carry a non-empty `disclosure_partition_ref` that matches a Reporting disclosure partition reference other than `blocked`. The builder MUST NOT infer a partition from the current user, incident role, selected recipient, visible section, or preview mode.

**REQ-RC-053**
`title_override` text MUST NOT contain LF and MUST NOT exceed 120 Unicode scalar values. A title-limit failure MUST use validation code `authored_title_limit_exceeded`.

**REQ-RC-054**
`authored_text` text MUST NOT exceed the configured default limit unless a server-side limit profile explicitly raises the default. It MUST NOT exceed the hard limit of 5000 Unicode scalar values. A text-limit failure MUST use validation code `authored_text_limit_exceeded`.

**REQ-RC-055**
`speaker_notes` text MUST satisfy the Reporting `speaker_notes_chars_per_slide` limit after subject placeholder substitution. A speaker-note limit failure MUST use validation code `authored_text_limit_exceeded`.

**REQ-RC-056**
Inline subject placeholders MUST use the exact form `{{subject:<stable_subject_ref>}}`. The placeholder start token is the exact ten-character sequence `{{subject:`. The placeholder end token is the exact two-character sequence `}}`.

**REQ-RC-057**
Any substring that starts with `{{subject:` and does not terminate with the next `}}`, or whose interior is empty, MUST fail with validation code `authored_subject_ref_unresolved`. Raw subject names, display tokens, email addresses, hostnames, account names, or party labels are not valid subject placeholder references.

**REQ-RC-058**
This revision does not scan authored free text for literal sensitive values outside the subject placeholder grammar. Builder linting MAY warn about possible literal sensitive values, but such linting is non-normative and MUST NOT be treated as release validation.

**REQ-RC-059**
For `external_release`, any authored text is permitted only when Reporting receives a redaction profile with `allow_authored_presentation_text=true`. This NLSpec's authoring routes MAY store authored text before that release-time decision. Omission behavior: when validation is requested without release scope, the authoring server validates schema, roles, partitions, and limits but MUST NOT claim external-release admission.

# 10. Composition Diagram Declarations

**REQ-RC-060**
`composition_diagram_decl.v1` MUST contain exactly the members in Table 10-A. Unknown members are invalid.

**Table 10-A. Composition Diagram Declaration Schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `decl_id` | identifier | Yes | No | None | Diagram declaration identity. |
| `replaces_decl_id` | identifier | Yes | Yes | `null` | `null` adds a diagram; non-null replaces one template-owned declaration. |
| `diagram_kind` | string | Yes | No | None | Must be a Reporting-supported diagram kind. |
| `selection_rule` | object | Yes | No | None | Must be one Reporting Table 15-C selection-rule object. |
| `label_overrides` | array of object | No | No | `[]` | Closed schema in Table 10-B. |

**REQ-RC-061**
When `replaces_decl_id` is non-null, it MUST equal `decl_id` and MUST resolve to exactly one template-owned diagram declaration. Replacing a composition-owned declaration from the same composition is invalid. Replacing a template declaration under a different `decl_id` is future-only.

**REQ-RC-062**
`selection_rule` MUST be a closed Reporting diagram selection-rule object. Raw Mermaid text, raw graph query text, arbitrary vertex declarations, arbitrary edge declarations, renderer syntax, and source-record mutation instructions are invalid and MUST use validation code `raw_generated_source_invalid` when detected by the authoring server.

**Table 10-B. Label Override Schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `target_kind` | string | Yes | No | None | Closed values: `vertex`, `edge`. |
| `target_ref` | string | Yes | No | None | Reporting-resolved vertex or edge reference under the selected diagram. |
| `label` | string | Yes | No | None | Single-line presentation label. |

**REQ-RC-063**
Label override `label` MUST NOT contain LF, MUST satisfy Reporting diagram label bounds, and MUST be presentation text only. A label override targeting a vertex mapped to a tokenized subject is invalid and MUST use validation code `diagram_label_override_invalid`.

# 11. Composition Operations

**REQ-RC-064**
`composition_op.v1` MUST contain exactly the members in Table 11-A. Unknown members are invalid.

**Table 11-A. Operation Object Schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `op_id` | identifier | Yes | No | None | Unique operation identity. |
| `op_kind` | string | Yes | No | None | Closed values in Table 11-B. |
| `on_unresolved` | string | No | No | `fail` | Closed values: `fail`, `drop`. |
| `payload` | object | Yes | No | None | Closed per-`op_kind` schema in Table 11-B. |

**REQ-RC-065**
Every operation payload MUST contain all required fields shown in Table 11-B and MUST NOT contain fields outside the listed payload fields for its `op_kind`.

**Table 11-B. Operation Payloads And Reporting Effects**

| `op_kind` | Required payload fields | Optional payload fields | Reporting effect owner |
| --- | --- | --- | --- |
| `exclude_section` | `section_anchor` | None | Reporting removes one emitted section before chunking; required-section behavior remains Reporting-owned. |
| `reorder_sections` | `section_anchors` | None | Reporting moves listed sections first in listed order; unlisted retained sections keep derived relative order. |
| `override_slide_layout` | `section_anchor`, `layout_id` | None | Reporting replaces layout with one template-declared layout token. |
| `override_title` | `section_anchor`, `authored_text_ref` | None | Reporting replaces section or slide title with admitted `title_override` text. |
| `set_speaker_notes` | `section_anchor`, `authored_text_ref` | None | Reporting attaches or replaces speaker notes with admitted `speaker_notes` text. |
| `insert_authored_block` | `block_anchor`, `position`, `authored_text_ref` | None | Reporting inserts one admitted `authored_text` paragraph block before or after the resolved block. |
| `exclude_block` | `block_anchor` | None | Reporting removes one derived block before chunking. |
| `override_click_profile` | `section_anchor`, `click_profile` | None | Reporting replaces click profile with `none`, `reveal_blocks`, or `reveal_list_items`. |
| `insert_diagram_slide` | `diagram_anchor`, `section_anchor`, `position` | None | Reporting emits one base slide for a composition-owned diagram declaration before chunking. |
| `exclude_diagram` | `diagram_anchor` | None | Reporting removes one template-owned diagram before diagram serialization. |
| `override_diagram_labels` | `diagram_anchor`, `label_overrides` | None | Reporting applies valid label overrides before Mermaid serialization. |

**REQ-RC-066**
`position` MUST be a closed string with value `before` or `after`. For `insert_authored_block`, the position is relative to the resolved block. For `insert_diagram_slide`, the position is relative to the resolved section.

**REQ-RC-067**
`layout_id` MUST be a template-declared layout token for the bound `template_id` and `template_version`. Unknown or undeclared layout tokens MUST fail with validation code `composition_schema_invalid`.

**REQ-RC-068**
`click_profile` MUST be one of `none`, `reveal_blocks`, or `reveal_list_items`. Unknown values MUST fail with validation code `composition_schema_invalid`.

**REQ-RC-069**
`authored_text_ref` MUST refer to an authored text object whose `text_role` is valid for the operation:

| Operation | Required `text_role` |
| --- | --- |
| `override_title` | `title_override` |
| `set_speaker_notes` | `speaker_notes` |
| `insert_authored_block` | `authored_text` |

**REQ-RC-070**
`reorder_sections.payload.section_anchors` MUST be a non-empty array of `section_anchor` objects. Duplicate anchor objects are invalid and MUST fail with validation code `composition_schema_invalid`.

**REQ-RC-071**
`override_diagram_labels.payload.label_overrides` MUST use the same schema and validation rules as `composition_diagram_decl.v1.label_overrides`.

# 12. Validation And Error Registry

**REQ-RC-072**
The validation algorithm `validate_report_composition_v1` MUST execute the stages in Table 12-A in order and MUST report the first failing stage in `cartulary.report_composition_validation_summary.v1.stage`.

**Table 12-A. Validation Stages**

| Stage | Required checks |
| --- | --- |
| `schema_validation` | Closed object schemas, required members, nullable rules, exact tokens, duplicate IDs, local references, and defaults. |
| `resource_binding_validation` | `incident_id`, `composition_id`, `template_id`, and `template_version` consistency with route resource and immutable version metadata. |
| `canonical_digest_validation` | Canonical serialization and `composition_sha256` recomputation for immutable versions. |
| `template_vocab_validation` | Template section declarations, layout tokens, click-profile tokens, and diagram declaration replacement targets. |
| `anchor_validation` | Anchor grammar and, when a render or validation context is supplied, target resolution and ambiguity checks. |
| `authored_text_validation` | Roles, LF rules, disclosure partition references, limits, placeholder grammar, and release-scope authored-text permission advisory. |
| `diagram_validation` | Diagram declaration schema, selection-rule closure, raw source rejection, and label override constraints. |
| `operation_validation` | Per-operation payload closure and operation-to-anchor/authored-text/diagram compatibility. |

**REQ-RC-073**
The request schema `cartulary.report_composition_validate_request.v1` MUST contain exactly the members in Table 12-B. Unknown members are invalid.

**Table 12-B. Validation Request**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `source_kind` | string | No | No | `draft` | Closed values: `draft`, `version`, `inline`. |
| `composition_version` | identifier | Required when `source_kind='version'` | No | None | Immutable version under the route resource. |
| `inline_composition` | `cartulary.report_composition.v1` | Required when `source_kind='inline'` | No | None | Validates supplied canonical bytes; does not persist. |
| `release_scope` | string | No | Yes | `null` | Closed non-null values: `internal_draft`, `external_release`. |
| `snapshot_id` | identifier | No | Yes | `null` | Enables snapshot-dependent anchor and subject validation. |
| `redaction_profile_id` | identifier | No | Yes | `null` | Enables release-scope authored-text permission checks. |

**REQ-RC-074**
The summary schema `cartulary.report_composition_validation_summary.v1` MUST contain exactly the members in Table 12-C. Unknown members are invalid.

**Table 12-C. Validation Summary**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `valid` | boolean | Yes | No | True only when no `error` issue exists. |
| `stage` | string | Yes | Yes | First failing stage or `null` when valid. |
| `issues` | array of `composition_issue.v1` | Yes | No | Sorted by stage order, then source array order, then code. |
| `composition_id` | identifier | Yes | Yes | Safe resource identity when attributable. |
| `composition_version` | identifier | Yes | Yes | Safe version identity when attributable. |
| `composition_sha256` | sha256_hex | Yes | Yes | Safe digest when attributable. |

**REQ-RC-075**
The issue schema `composition_issue.v1` MUST contain exactly the members in Table 12-D. Unknown members are invalid.

**Table 12-D. Validation Issue**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `severity` | string | Yes | No | Closed values: `error`, `warning`. |
| `code` | string | Yes | No | One code from Table 12-E. |
| `message_key` | string | Yes | No | Stable localization key; not release evidence. |
| `safe_details` | object | Yes | No | Only keys from Table 12-F. |

**REQ-RC-076**
Validation codes MUST use Table 12-E. A conforming implementation MAY add non-normative warning codes only under an extension namespace beginning with `x_`. It MUST NOT add new error codes without revising this table.

**Table 12-E. Validation Codes**

| Code | Severity | Stage | Required meaning |
| --- | --- | --- | --- |
| `composition_schema_invalid` | `error` | `schema_validation` | Closed schema, required member, type, nullable, token, or local-reference violation. |
| `composition_duplicate_id` | `error` | `schema_validation` | Duplicate `op_id`, `decl_id`, or `authored_text_id`. |
| `composition_incident_mismatch` | `error` | `resource_binding_validation` | Route/body/resource incident mismatch. |
| `composition_template_mismatch` | `error` | `resource_binding_validation` | Composition template binding differs from route, resource, or render tuple. |
| `composition_id_mismatch` | `error` | `resource_binding_validation` | Body/version `composition_id` differs from route resource. |
| `composition_not_found` | `error` | `resource_binding_validation` | Requested composition resource does not exist in the incident. |
| `composition_version_not_found` | `error` | `resource_binding_validation` | Requested immutable version does not exist under the resource. |
| `composition_draft_version_conflict` | `error` | `resource_binding_validation` | `base_draft_version` differs from current `draft_version`. |
| `composition_version_immutable` | `error` | `resource_binding_validation` | Request attempts to mutate immutable version bytes. |
| `composition_version_bound` | `error` | `resource_binding_validation` | Request attempts to delete or rewrite a release-bound version. |
| `composition_delete_not_supported` | `error` | `resource_binding_validation` | Delete route is intentionally unsupported by the implementation. |
| `composition_digest_mismatch` | `error` | `canonical_digest_validation` | Stored or supplied digest differs from canonical bytes. |
| `composition_anchor_invalid` | `error` | `anchor_validation` | Anchor grammar or disallowed target kind is invalid. |
| `composition_anchor_unresolved` | `error` | `anchor_validation` | Anchor resolves to zero targets and `on_unresolved='fail'` or release scope forbids drop. |
| `composition_anchor_ambiguous` | `error` | `anchor_validation` | Anchor resolves to more than one target. |
| `composition_drop_invalid_for_external_release` | `error` | `anchor_validation` | `on_unresolved='drop'` appears for an external-release validation or render. |
| `authored_text_not_permitted` | `error` | `authored_text_validation` | External release would include authored text without profile permission. |
| `authored_title_limit_exceeded` | `error` | `authored_text_validation` | `title_override` exceeds its limit or contains LF. |
| `authored_text_limit_exceeded` | `error` | `authored_text_validation` | `authored_text` or `speaker_notes` exceeds applicable limit. |
| `authored_subject_ref_unresolved` | `error` | `authored_text_validation` | Placeholder is malformed, unresolved, ambiguous, filtered, or unreleasable. |
| `raw_generated_source_invalid` | `error` | `diagram_validation` | Raw Markdown, Mermaid, HTML, renderer syntax, arbitrary node, or arbitrary edge input is present. |
| `diagram_label_override_invalid` | `error` | `diagram_validation` | Label override targets a tokenized subject vertex or violates label constraints. |
| `diagram_selection_missing_ref` | `error` | `diagram_validation` | Diagram selection references unavailable snapshot or projection data. |

**Table 12-F. Safe Detail Keys**

| Key | Type | Rule |
| --- | --- | --- |
| `composition_id` | identifier | Safe resource identity. |
| `composition_version` | identifier | Safe immutable version identity. |
| `composition_op_id` | identifier | Safe operation identity when attributable. |
| `composition_anchor_kind` | string | `section_anchor`, `record_anchor`, `block_anchor`, or `diagram_anchor`. |
| `diagram_id` | identifier | Diagram declaration identity. |
| `authored_text_id` | identifier | Authored text identity. |
| `template_id` | identifier | Template identity. |
| `template_version` | identifier | Template version. |
| `field` | string | Schema member path without raw values. |
| `limit` | finite integer | Numeric limit that was exceeded. |

# 13. Builder UI Boundary

**REQ-RC-077**
A conforming report builder UI MUST emit only this NLSpec's route requests and schema objects for composition authoring. It MUST NOT persist generated Markdown, Mermaid, rendered bytes, workbook mutations, template mutations, or snapshot mutations as a composition substitute.

**REQ-RC-078**
When a builder UI offers a control for a composition effect, the control MUST map to exactly one `op_kind`, `authored_text.v1`, `composition_diagram_decl.v1`, or validation/preview request defined by this NLSpec. A control MUST NOT emit renderer-local syntax or hidden ordinal-path targets.

**REQ-RC-079**
A conforming builder UI for this revision MUST provide authoring paths for every `op_kind` in Table 11-B, `authored_text.v1` objects, composition diagram declarations, server validation, and authoritative preview requests. Omission behavior: an implementation without a builder UI may still implement the route family, but it MUST NOT claim builder UI conformance.

**REQ-RC-080**
The builder UI MUST present server validation results from `cartulary.report_composition_validation_summary.v1` as authoritative. Client-side linting MAY run earlier, but client-side linting MUST NOT mark a composition valid for release or approval without server validation.

**REQ-RC-081**
The builder UI MAY render a local live approximation. Such an approximation is non-normative, is not reviewable or approvable output, and MUST NOT be used as evidence that Reporting bytes will pass validation.

**REQ-RC-082**
Authoritative previews MUST be requested through `POST /report-compositions/{composition_id}/preview` and MUST delegate to Reporting as an ordinary `internal_draft` render attempt. Preview output bytes remain Reporting-owned and are not immutable release bytes.

**REQ-RC-083**
The builder UI MUST serialize subject references inside authored text using the placeholder syntax from REQ-RC-056. It MUST NOT serialize raw subject display values as hidden references.

# 14. Acceptance Criteria And Fixtures

**REQ-RC-084**
Every requirement in this NLSpec MUST trace to at least one acceptance criterion in Table 14-B or fixture in Table 14-A before promotion to `adopted/current`.

**Table 14-A. Required Fixtures**

| Fixture ID | Required fixture | Pass condition |
| --- | --- | --- |
| `RC-FIX-001` | Empty valid composition with no operations, no diagrams, and no authored text. | Canonical bytes are deterministic and digest excludes only `composition_sha256`. |
| `RC-FIX-002` | Same composition serialized with non-canonical member ordering. | Server canonicalization produces the same digest as the canonical fixture. |
| `RC-FIX-003` | Duplicate `op_id`, `decl_id`, and `authored_text_id` cases. | Each duplicate fails with `composition_duplicate_id`. |
| `RC-FIX-004` | Draft update with stale `base_draft_version`. | Request fails with `composition_draft_version_conflict` and does not mutate draft state. |
| `RC-FIX-005` | Immutable version mutation and release-bound delete attempts. | Requests fail with `composition_version_immutable` or `composition_version_bound`. |
| `RC-FIX-006` | Every `op_kind` with valid payload and one closed-schema violation. | Valid payloads pass schema validation; extra or missing payload fields fail with `composition_schema_invalid`. |
| `RC-FIX-007` | Semantic anchors for resolved, unresolved, ambiguous, and generated-ID targets. | Generated IDs fail with `composition_anchor_invalid`; unresolved and ambiguous anchors use exact codes. |
| `RC-FIX-008` | `on_unresolved='drop'` internal draft and external release validation. | Internal draft may drop zero-match operations; external release fails with `composition_drop_invalid_for_external_release`. |
| `RC-FIX-009` | Authored text role, LF, limit, partition, and placeholder cases. | Each invalid case fails with the exact authored-text validation code. |
| `RC-FIX-010` | Diagram declarations with valid selection, raw Mermaid, arbitrary nodes, arbitrary edges, and tokenized-subject label override. | Valid declaration passes; invalid cases fail with `raw_generated_source_invalid` or `diagram_label_override_invalid`. |
| `RC-FIX-011` | Route authorization matrix for viewer, editor, reviewer, admin, no membership, and deployment-admin-only user. | Minimum roles in Table 6-A are enforced and deployment-admin-only access is rejected. |
| `RC-FIX-012` | Builder preview request with current draft. | Request produces a Reporting-owned `internal_draft` render attempt reference and no immutable release bytes. |

**Table 14-B. Acceptance Criteria**

| ID | Criterion | Traces to | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `RC-AC-AUTH-001` | Authority boundary is explicit. | §1 | This NLSpec owns authoring schemas and lifecycle only; Reporting owns render effects. | The document states or implies composition authoring owns render bytes or release approval. |
| `RC-AC-SCOPE-001` | Non-goals are closed. | §3 | Raw source editing, workbook mutation, template mutation, post-redaction editing, and cross-template migration are explicitly rejected or future-only. | Any non-goal is admitted as current behavior. |
| `RC-AC-LIFE-001` | Draft/version lifecycle is deterministic. | §5 | Draft updates use `base_draft_version`; freezes produce immutable `vN` versions and digest bytes atomically. | Stale writes mutate state, versions are mutable, or version identifiers are ambiguous. |
| `RC-AC-ROUTE-001` | Route interfaces are unambiguous. | §6 | Each route has method, path, role, idempotency, request, and response shape. | A route omits role, idempotency, or request/response contract. |
| `RC-AC-AUTHZ-001` | Authorization uses incident roles only. | §6 | Viewer/editor route gates match Table 6-A and deployment-admin-only access is rejected. | Composition introduces resource-specific ACLs or deployment-admin bypass. |
| `RC-AC-CANON-001` | Canonical bytes and digest are stable. | §7 | Recomputed SHA-256 over canonical bytes excluding only `composition_sha256` matches the stored digest. | Digest depends on insertion order, whitespace, storage metadata, or the digest member itself. |
| `RC-AC-ANCHOR-001` | Anchors are semantic only. | §8 | Section, record, block, and diagram anchors validate by semantic fields and reject generated IDs. | Ordinal paths or generated structural IDs are accepted. |
| `RC-AC-TEXT-001` | Authored text is bounded presentation text. | §9 | Roles, LF rules, limits, partitions, and placeholders validate exactly. | Authored text can enter as free-form facts or unpartitioned content. |
| `RC-AC-DIAGRAM-001` | Diagrams use closed selection declarations. | §10 | Raw Mermaid, arbitrary nodes, arbitrary edges, and tokenized-subject label overrides fail. | Diagram authoring bypasses Reporting selection rules. |
| `RC-AC-OPS-001` | Operation schema is closed. | §11 | Every operation accepts only its declared payload and valid local references. | Unknown payload members or wrong text-role references pass validation. |
| `RC-AC-VALID-001` | Validation output is machine-testable. | §12 | Validation summaries use declared stages, issue codes, and safe details only. | Validation emits raw sensitive values or undocumented codes. |
| `RC-AC-BUILDER-001` | Builder UI is a composition-data author. | §13 | Builder emits route requests, validation requests, and preview requests only. | Builder edits generated source, workbook records, templates, or release bytes. |
| `RC-AC-PREVIEW-001` | Preview boundary is Reporting-owned. | §13 | Authoritative preview creates an `internal_draft` Reporting attempt, not approvable release bytes. | Client preview or builder bytes are treated as reviewable output. |

# 15. Promotion Checklist

**REQ-RC-085**
This NLSpec MUST NOT be promoted to `adopted/current` until every row in Table 15-A is satisfied.

**Table 15-A. Promotion Conditions**

| Condition | Required state |
| --- | --- |
| Domain vocabulary | `docs/domain.md` includes the composition terms introduced by this NLSpec with extension-profile status. |
| Reporting import boundary | `docs/reporting-subsystem-nlspec.md` imports `cartulary.report_composition.v1` and the companion identifiers without owning authoring routes or schema. |
| Core route conventions | Core 01 route envelope, idempotency, and common error conventions are available for this route family. |
| Core authorization conventions | Core 04 incident role derivation and deployment-admin bypass prohibition are available for this route family. |
| Fixture coverage | Every `RC-FIX-*` row exists in the accepted fixture suite or is explicitly marked future-only with owner approval. |
| Traceability | Every `REQ-RC-*` maps to at least one `RC-AC-*` or `RC-FIX-*`. |
