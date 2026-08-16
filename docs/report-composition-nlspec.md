---
title: Cartulary Report Composition NLSpec
status: adopted/current
document_class: nlspec
profile: snapshot_reporting
document_version: 1.2.0
schema_id: cartulary.report_composition_nlspec.v1
---

# 1. Status, Scope, And Authority

This NLSpec defines the Cartulary Report Composition companion subsystem for the Snapshot and Reporting Extension Profile. It becomes implementation-conformance authority only after promotion to `status: adopted/current` and after the required Core and Reporting companion amendments named by this document are adopted.

**REQ-RC-001**
This NLSpec owns only the authoring-side composition boundary:

- `cartulary.report_composition.v1` canonical composition bytes;
- composition resource lifecycle, draft mutation, immutable versioning, and freeze behavior;
- composition authoring route family contracts;
- `composition_op.v1` operation schemas;
- semantic composition anchor grammars;
- `authored_text.v1` presentation-text objects;
- `composition_diagram_decl.v1` declaration objects and companion layout objects;
- builder-facing validation summaries, issue codes, and safe detail keys;
- report builder UI conformance boundaries.

**REQ-RC-002**
This NLSpec MUST NOT own report materialization, release tuple admission, redaction, token substitution, Mermaid source generation, manual diagram SVG serialization, Slidev source generation, deck derivation effects, render sandboxing, render bundle hashing, release approval, release publication, or Reporting fixture bytes. Those behaviors remain owned by Core 01, Core 04, and `docs/reporting-subsystem-nlspec.md`.

**REQ-RC-003**
This NLSpec is a companion to `docs/reporting-subsystem-nlspec.md`. Reporting consumes immutable composition versions by `composition_id`, `composition_version`, and `composition_sha256`. Reporting owns the observable render effects of a valid composition after Core route admission accepts a render attempt.

**REQ-RC-004**
When this NLSpec conflicts with Core 00 through Core 04 outside the composition subsystem, the conflict is a defect in this NLSpec. When this NLSpec conflicts with `docs/reporting-subsystem-nlspec.md`, the owner boundary in Table 1-A MUST decide the defect location.

**Table 1-A. Owner Boundary**

| Area | Owner | Boundary rule |
| --- | --- | --- |
| Public route envelopes, common error envelope, idempotency mechanics, job resource conventions | Core 01 | This NLSpec names route-specific request and response members only. |
| Incident roles, authorization derivation, deployment-admin limits, release approval | Core 04 | This NLSpec imports role names and MUST NOT create new incident ACL machinery. |
| Composition draft, version, schema, operation vocabulary, anchors, authored text, diagram declaration and layout authoring | This NLSpec | Reporting imports these identifiers but MUST NOT redefine their schema. |
| Composition render effect, redaction admission, `derive_deck_v2`, diagrams in render output, fixture bytes | Reporting Subsystem NLSpec | This NLSpec names the intended target and payload only. |
| Graph selection validation and graph projection lifecycle | Graph Projection NLSpec and Reporting Subsystem NLSpec | This NLSpec admits diagram declarations only through closed selection-rule and layout objects. |
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
| `composition diagram declaration` | Composition-owned diagram declaration that uses Reporting diagram selection rules, presentation labels, optional closed layout data, and no raw Mermaid. |
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
| `cartulary.report_composition_preview_view.v1` | route schema | §6 | Server view of one authoritative preview request. |
| `cartulary.report_composition_preview_source.v1` | route schema | §6 | Internal Reporting preview source descriptor for a draft or immutable version. |
| `cartulary.report_composition_validate_request.v1` | route schema | §12 | Server validation request. |
| `cartulary.report_composition_validation_context.v1` | route schema | §12 | Optional render context for snapshot-dependent validation. |
| `cartulary.report_composition_validation_summary.v1` | route schema | §12 | Server validation result. |
| `composition_issue.v1` | route schema | §12 | One validation issue. |
| `composition_op.v1` | composition schema | §11 | Closed operation object. |
| `section_anchor` | anchor grammar | §8 | Section target. |
| `record_anchor` | anchor grammar | §8 | Source record target. |
| `block_anchor` | anchor grammar | §8 | Derived block target. |
| `diagram_anchor` | anchor grammar | §8 | Diagram declaration target. |
| `authored_text.v1` | composition schema | §9 | Authored presentation text object. |
| `composition_diagram_decl.v1` | composition schema | §10 | Composition-owned diagram declaration. |
| `composition_diagram_layout.v1` | composition schema | §10 | Closed manual layout data for one composition-owned diagram declaration. |
| `composition_label_override_target.v1` | composition schema | §10 | Structured vertex or edge target for diagram label overrides. |
| `create_composition_draft_v1` | route operation | §6 | Draft resource creation. |
| `update_composition_draft_v1` | route operation | §6 | Draft mutation. |
| `freeze_composition_version_v1` | route operation | §6 | Immutable version creation. |

**REQ-RC-018**
The identifier `composition_version` MUST match `v[1-9][0-9]*`. The first immutable version for a composition resource is `v1`. Later immutable versions for the same resource MUST increment the decimal suffix by one with no gap after a successful version creation. The literal `latest` is invalid in public requests and release tuples.

**REQ-RC-019**
The identifier `composition_id` is opaque, server-assigned, and immutable. Clients MUST NOT parse it or derive incident, template, or version information from it.

**REQ-RC-019a**
Composition route requests, route responses, draft state, immutable composition documents, validation summaries, and preview-source descriptors MUST import the Reporting §8 JSON and scalar rules for UTF-8 decoding, duplicate JSON member rejection, explicit `null` handling, closed-object handling, canonical JSON serialization, `identifier`, `timestamp`, `sha256_hex`, `safe_string`, and `finite_integer` unless this NLSpec defines a narrower scalar in Table 4-C. A duplicate JSON object member at any depth MUST fail before route side effects with validation code `composition_schema_invalid`.

**Table 4-C. Composition scalar and grammar contracts**

| Contract | Definition |
| --- | --- |
| `positive_integer` | Reporting `finite_integer` with mathematical value in `[1, 9007199254740991]`. |
| `composition_id` | Reporting `identifier`; server-assigned; clients MUST NOT parse it. |
| `composition_version` | String matching `v[1-9][0-9]*`; `latest`, `draft`, `preview`, and aliases are invalid. |
| `op_id` | Reporting `identifier`; unique in `deck_ops[]`. |
| `decl_id` | Reporting `identifier`; unique in `diagram_decls[]` and used as `diagram_id` when Reporting admits a composition-owned diagram. |
| `authored_text_id` | Reporting `identifier`; unique in `authored_texts[]`. |
| `ref_segment` | JSON string containing 1 to 128 Unicode scalar values and no colon, slash, backslash, hash, Unicode whitespace, C0 controls, C1 controls, NUL, or surrogate code points. |

**REQ-RC-019b**
Every optional member with a declared default MUST materialize to that default before draft validation. Immutable `cartulary.report_composition.v1` bytes MUST include every materialized defaulted member. Omitted defaults and explicit default values MUST canonicalize to byte-identical immutable composition documents.

**REQ-RC-019c**
Canonical JSON for composition remains imported from Reporting until a Core owner adopts a shared canonical JSON profile. A later Core-owned profile MAY replace the owner document only if it preserves all v1 canonical bytes for conforming documents.

# 5. Lifecycle

**REQ-RC-020**
A composition resource MUST be scoped to exactly one `incident_id`, exactly one `template_id`, and exactly one `template_version`. These three values are fixed at resource creation and MUST NOT be changed by draft update, version creation, validation, preview, or release binding.

**REQ-RC-021**
Every composition resource MUST have at most one active draft. A draft is mutable until the resource is retired. The draft created with a resource MUST have `draft_version=1`. A draft update MUST increment `draft_version` exactly once on success.

**REQ-RC-022**
Every mutable draft MUST carry a server-emitted `draft_version` `positive_integer`. Mutating requests MUST carry `base_draft_version`. If `base_draft_version` does not equal the current server `draft_version`, the request MUST fail with `error.code='conflict'` and validation code `composition_draft_version_conflict`.

**REQ-RC-023**
Creating an immutable composition version MUST copy the current draft body, materialize all defaults, compute canonical bytes, compute `composition_sha256`, assign the next `composition_version`, and persist the resulting `cartulary.report_composition.v1` document atomically.

**REQ-RC-024**
Immutable composition versions MUST NOT be modified, reserialized with different canonical bytes, deleted while release-bound, or assigned a new digest. A request that attempts to mutate an immutable version MUST fail with validation code `composition_version_immutable`.

**REQ-RC-025**
When a release tuple binds `composition_id`, `composition_version`, and `composition_sha256`, the referenced immutable version is release-bound. Later draft edits and later immutable versions MUST NOT change the bound version or its digest.

**REQ-RC-026**
A release-bound composition version MUST NOT be deleted or rewritten. A request that attempts to delete or rewrite a release-bound version MUST fail with validation code `composition_version_bound`.

**REQ-RC-027**
Retiring a composition resource is permitted only when no immutable version under that resource is release-bound. Retirement MUST preserve immutable versions for audit and digest validation. A retired resource MUST remain readable through `GET /report-compositions/{composition_id}` and immutable-version reads. Omission behavior: if retirement is not implemented, all delete requests MUST fail with validation code `composition_delete_not_supported`.

**REQ-RC-027a**
When a resource is retired, `retired_at` MUST be the route-time timestamp assigned by the server and MUST remain immutable. A retired resource MUST reject draft update, version freeze, and preview with validation code `composition_resource_retired`. Repeating the same successful retire request through Core 01 idempotency MUST return the same logical retired resource view and MUST NOT change `retired_at`.

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
| `GET /report-compositions` | List composition resources for the incident. | `viewer` | No | Core cursor pagination query only | `composition_resources[]` with Core paging metadata |
| `POST /report-compositions` | Create draft resource. | `editor` | `client_txn_id` | Table 6-B create body | `cartulary.report_composition_resource_view.v1` |
| `GET /report-compositions/{composition_id}` | Read resource and current draft metadata. | `viewer` | No | None | `cartulary.report_composition_resource_view.v1` |
| `PATCH /report-compositions/{composition_id}` | Update active draft. | `editor` | `client_txn_id` | Table 6-C update body | `cartulary.report_composition_resource_view.v1` |
| `DELETE /report-compositions/{composition_id}` | Retire resource when allowed. | `editor` | `client_txn_id` | Table 6-D delete body | `cartulary.report_composition_resource_view.v1` |
| `POST /report-compositions/{composition_id}/versions` | Freeze immutable version. | `editor` | `client_txn_id` | Table 6-E freeze body | `cartulary.report_composition_version_view.v1` |
| `GET /report-compositions/{composition_id}/versions/{composition_version}` | Read immutable version. | `viewer` | No | None | `cartulary.report_composition_version_view.v1` |
| `POST /report-compositions/{composition_id}/validate` | Validate draft or version. | `viewer` | No | `cartulary.report_composition_validate_request.v1` | `cartulary.report_composition_validation_summary.v1` |
| `POST /report-compositions/{composition_id}/preview` | Request authoritative `internal_draft` preview through Reporting. | `viewer` | `client_txn_id` | Table 6-F preview body | `cartulary.report_composition_preview_view.v1` |

**REQ-RC-031a**
`GET /report-compositions` MUST page every composition resource visible to the caller in the route incident, including retired resources, using Core cursor pagination. The logical collection MUST sort active resources before retired resources, then by `template_id`, `template_version`, and `composition_id` using exact code point order before pagination. The route has no v1 filtering or hidden default state filter. Omitted pagination query members use the Core default limit; singleton routes in this family MUST reject pagination query members.

**REQ-RC-031b**
The algorithms in Table 6-A1 define the route-visible lifecycle effects. Route implementations MAY use any storage mechanism only when the observable state transitions, idempotent replay behavior, validation codes, and response bodies are identical.

**Table 6-A1. Route lifecycle algorithms**

| Operation | Required state transition |
| --- | --- |
| Create draft resource | Allocate a new opaque `composition_id`; set fixed incident/template binding; set `draft_version=1`; materialize omitted arrays to `[]`; set `latest_composition_version=null`, `release_bound_versions=[]`, and `retired_at=null`; validate the materialized draft before persisting. |
| Update active draft | Require active resource and matching `base_draft_version`; replace each present draft array as a whole; leave omitted draft members unchanged; materialize defaults; validate the resulting draft; persist and increment `draft_version` exactly once only after validation succeeds. |
| Retire resource | Require active resource, matching `base_draft_version`, and no release-bound immutable version; set immutable `retired_at`; preserve draft and immutable versions for audit. |
| Freeze version | Require active resource and matching `base_draft_version`; copy current draft after defaults; assign the next gapless `composition_version`; compute and verify `composition_sha256`; persist the immutable document atomically; set `latest_composition_version` to the new version. |
| Read resource | Return the current draft arrays, immutable version metadata, release-bound versions, and `retired_at`; reading a retired resource is valid. |
| Validate | Select the source by Table 12-B1; run Table 12-A stages; do not mutate resource state. |
| Preview | Select the source by Table 6-F and Table 6-J; materialize a `cartulary.report_composition_preview_source.v1`; delegate to Reporting as `internal_draft`; do not create immutable release bytes. |

**REQ-RC-032**
A route path `incident_id` MUST equal the resource `incident_id`. A body `incident_id`, when present, MUST equal the route path `incident_id`. A mismatch MUST fail with `error.code='invalid_request'` and validation code `composition_incident_mismatch`.

**REQ-RC-033**
A route path `composition_id` MUST identify a composition resource inside the route path incident. No-match requests MUST fail with `error.code='not_found'` and validation code `composition_not_found`.

**REQ-RC-033a**
Composition routes MUST use Core 01 route-envelope, authorization, idempotency, and normalized-request comparison semantics. This NLSpec defines only route-specific validation codes and response bodies. The observable route outcomes MUST follow Table 6-A2.

**Table 6-A2. Route outcome and idempotency matrix**

| Condition | HTTP status / Core error | Composition validation code | Side effect rule |
| --- | --- | --- | --- |
| Structurally invalid JSON, duplicate object member, unknown member, invalid null, invalid scalar, or closed-token violation | Core invalid-payload status and envelope | `composition_schema_invalid` when attributable | No state mutation, no idempotency success record, no Reporting invocation. |
| Authenticated actor lacks the minimum Table 6-A role, including deployment-admin-only access | Core authorization envelope | None required | No state mutation and no existence leak beyond Core 04 rules. |
| Route resource not found inside the incident | Core `not_found` envelope | `composition_not_found` | No state mutation. |
| `base_draft_version` differs from current server `draft_version` and no prior exact idempotency hit exists | `409` / Core conflict envelope | `composition_draft_version_conflict` | No state mutation. |
| Exact replay of a previously committed mutating request with the same route idempotency scope and normalized body | Core idempotent replay success status | Original validation result when one was present | Return the original committed logical response before fresh validation or conflict checks; do not mutate again. |
| Same route idempotency scope and `client_txn_id` reused with a different normalized body | `409` / `client_txn_conflict` | None required | No state mutation; this is a Core idempotency failure, not a composition validation failure. |
| First successful create, update, retire, freeze, or preview request | Route-specific success status from Core 01 | None | Persist exactly the Table 6-A1 transition, then record the idempotency result for mutating routes. |

For idempotency comparison, omitted members that materialize to defaults compare equal to explicit default values, and request bodies compare after this NLSpec's scalar validation and default materialization. A failed request MUST NOT create a committed idempotency success result.

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
| `base_draft_version` | `positive_integer` | Yes | No | None | Must equal current `draft_version`. |
| `authored_against_snapshot_id` | identifier | No | Yes | no change | Advisory only. |
| `deck_ops` | array of `composition_op.v1` | No | No | no change | Replaces the full draft array when present. |
| `diagram_decls` | array of `composition_diagram_decl.v1` | No | No | no change | Replaces the full draft array when present. |
| `authored_texts` | array of `authored_text.v1` | No | No | no change | Replaces the full draft array when present. |

**Table 6-D. Delete Body**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `client_txn_id` | string | Yes | No | None | Core 01 idempotency key. |
| `base_draft_version` | `positive_integer` | Yes | No | None | Must equal current `draft_version` unless the resource is already retired by the same idempotent request. |

**Table 6-E. Freeze Version Body**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `client_txn_id` | string | Yes | No | None | Core 01 idempotency key. |
| `base_draft_version` | `positive_integer` | Yes | No | None | Must equal current `draft_version`. |

**Table 6-F. Preview Body**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `client_txn_id` | string | Yes | No | None | Core 01 idempotency key. |
| `source_kind` | string | No | No | `draft` | Closed values: `draft`, `version`. |
| `composition_version` | identifier | Required when `source_kind='version'` | No | None | Must name an immutable version under the resource. |
| `snapshot_id` | identifier | Yes | No | None | Passed to Reporting as preview input. |
| `derivation_version` | identifier | Yes | No | None | Reporting derivation version for the preview render. |
| `template_id` | identifier | Yes | No | None | Must equal resource `template_id`. |
| `template_version` | identifier | Yes | No | None | Must equal resource `template_version`. |
| `redaction_profile_id` | identifier | Yes | No | None | Passed to Reporting as preview input. |
| `redaction_profile_version` | identifier | Yes | No | None | Exact redaction profile version passed to Reporting. |
| `redaction_profile_sha256` | sha256_hex | Yes | No | None | Digest of the redaction profile bytes passed to Reporting. |
| `render_environment_profile_id` | identifier | Yes | No | None | Must name a template-declared Reporting render profile. |
| `output_kind` | string | Yes | No | None | Closed Reporting output kind: `mermaid` or `slidev`. |
| `output_options` | object | No | No | Reporting §7.5 defaults | Reporting output options. Omitted object materializes exactly as Reporting §7.5. |
| `recipient_partition_refs` | array of string | No | No | `[]` | Passed to Reporting as preview input. |
| `graph_projection_refs` | array of Reporting `source_projection_ref.v2` | No | No | `[]` | Exact immutable Graph Projection references available to preview diagram validation. |

Preview requests always delegate to Reporting with `release_scope='internal_draft'`. After default materialization, `recipient_partition_refs` MUST be `[]`; a non-empty value MUST fail before Reporting invocation with validation code `composition_schema_invalid`. `graph_projection_refs` MUST sort bytewise by `graph_view_id`, and duplicate `graph_view_id` values MUST fail before Reporting invocation with validation code `composition_validation_context_invalid`.

**REQ-RC-034**
The success schema `cartulary.report_composition_resource_view.v1` MUST contain the members in Table 6-G and MUST NOT contain unknown members.

**Table 6-G. Resource View**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `composition_id` | identifier | Yes | No | Resource identity. |
| `incident_id` | identifier | Yes | No | Resource incident. |
| `template_id` | identifier | Yes | No | Fixed template binding. |
| `template_version` | identifier | Yes | No | Fixed template version binding. |
| `draft_version` | `positive_integer` | Yes | No | Current mutable draft version. |
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

**REQ-RC-035a**
The success schema `cartulary.report_composition_preview_view.v1` MUST contain the members in Table 6-I and MUST NOT contain unknown members. Preview views are route responses and release-state metadata only; they are not immutable release evidence.

**Table 6-I. Preview View**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `preview_attempt_id` | identifier | Yes | No | Opaque Core job or attempt identity for this preview request. |
| `render_attempt_id` | identifier | Yes | No | Reporting-owned internal-draft render attempt identity created during successful preview admission. |
| `incident_id` | identifier | Yes | No | Route incident. |
| `composition_id` | identifier | Yes | No | Resource identity. |
| `source_kind` | string | Yes | No | `draft` or `version`. |
| `draft_version` | `positive_integer` | Yes | Yes | Current draft version when `source_kind='draft'`; otherwise `null`. |
| `composition_version` | identifier | Yes | Yes | Immutable version when `source_kind='version'`; otherwise `null`. |
| `preview_source_sha256` | sha256_hex | Yes | No | Digest of `cartulary.report_composition_preview_source.v1` canonical bytes excluding only `preview_source_sha256`. |
| `composition_sha256` | sha256_hex | Yes | Yes | Non-null only when `source_kind='version'`; digest of immutable `cartulary.report_composition.v1` bytes. |
| `snapshot_id` | identifier | Yes | No | Reporting preview snapshot. |
| `derivation_version` | identifier | Yes | No | Reporting derivation version. |
| `template_id` | identifier | Yes | No | Resource template. |
| `template_version` | identifier | Yes | No | Resource template version. |
| `redaction_profile_id` | identifier | Yes | No | Reporting redaction profile. |
| `redaction_profile_version` | identifier | Yes | No | Reporting redaction profile version. |
| `redaction_profile_sha256` | sha256_hex | Yes | No | Reporting redaction profile digest. |
| `render_environment_profile_id` | identifier | Yes | No | Reporting render environment profile. |
| `output_kind` | string | Yes | No | Reporting output kind. |
| `output_options` | object | Yes | No | Reporting output options after §7.5 defaults. |
| `recipient_partition_refs` | array of string | Yes | No | Materialized recipient partitions. |
| `graph_projection_refs` | array of Reporting `source_projection_ref.v2` | Yes | No | Materialized graph projection refs sorted by Reporting rules. |
| `release_scope` | string | Yes | No | Exact `internal_draft`. |

**REQ-RC-035b**
`cartulary.report_composition_preview_source.v1` MUST contain the members in Table 6-J and MUST NOT contain unknown members. Reporting MAY consume this descriptor only for authoritative preview. It MUST NOT appear in an external release tuple, immutable composition document, render bundle manifest, approval record, or release evidence artifact.

**Table 6-J. Preview Composition Source**

| Member | Type | Required | Nullable | Rule |
| --- | --- | ---: | ---: | --- |
| `schema_id` | string | Yes | No | Exact `cartulary.report_composition_preview_source.v1`. |
| `source_kind` | string | Yes | No | `draft` or `version`. |
| `incident_id` | identifier | Yes | No | Route incident. |
| `composition_id` | identifier | Yes | No | Resource identity. |
| `draft_version` | `positive_integer` | Yes | Yes | Present only for draft preview. |
| `composition_version` | identifier | Yes | Yes | Present only for immutable version preview. |
| `preview_source_sha256` | sha256_hex | Yes | No | Digest of canonical preview-source bytes after deleting only this member. |
| `composition_sha256` | sha256_hex | Yes | Yes | Non-null only for immutable version preview; digest of immutable `cartulary.report_composition.v1` bytes. |
| `template_id` | identifier | Yes | No | Resource template. |
| `template_version` | identifier | Yes | No | Resource template version. |
| `authored_against_snapshot_id` | identifier | Yes | Yes | Advisory authoring context copied from the selected draft or immutable version. |
| `deck_ops` | array of `composition_op.v1` | Yes | No | Materialized operation array from the selected draft or immutable version. |
| `diagram_decls` | array of `composition_diagram_decl.v1` | Yes | No | Materialized diagram declaration array from the selected draft or immutable version. |
| `authored_texts` | array of `authored_text.v1` | Yes | No | Materialized authored text array from the selected draft or immutable version. |

`preview_source_sha256` MUST be computed over the Reporting canonical JSON serialization of the Table 6-J object after deleting only the top-level `preview_source_sha256` member. For `source_kind='draft'`, `composition_sha256` MUST be `null`; the draft preview digest is `preview_source_sha256` only. For `source_kind='version'`, `composition_sha256` MUST equal the immutable version digest and `preview_source_sha256` still binds the preview descriptor bytes. A preview digest MUST NOT satisfy a release tuple `composition_sha256`.

**REQ-RC-035c**
Preview source selection MUST use the `draft` and `version` rows of Table 12-B1. `source_kind='inline'` is invalid for preview. A preview request with `source_kind='draft'` MUST omit `composition_version`; a preview request with `source_kind='version'` MUST include `composition_version`. Violations MUST fail with validation code `composition_source_invalid` before Reporting is invoked.

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
`on_unresolved='drop'` is invalid for an `external_release` Reporting render attempt. Reporting MUST fail such a render with `composition_drop_invalid_for_external_release`. The authoring server MUST report the same issue when validation is requested with `validation_context.release_scope='external_release'`.

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

**REQ-RC-051a**
Authored text bodies MUST be decoded as UTF-8 JSON strings and normalized to Unicode NFC before validation, canonical composition serialization, placeholder parsing, and limit evaluation. A body MUST NOT contain NUL, surrogate code points, C0 controls, or C1 controls. LF is valid only for roles whose Table 9-B row says LF allowed. CR and TAB are invalid for every role. A body whose NFC-normalized scalar count is `0`, or whose scalar sequence contains only Unicode whitespace, fails with validation code `authored_text_limit_exceeded` except that `title_override` uses `authored_title_limit_exceeded`.

**Table 9-B. Authored Text Roles**

| `text_role` | Admission target | LF allowed | Default limit | Hard limit |
| --- | --- | ---: | ---: | ---: |
| `title_override` | Section or slide title. | No | 120 | 120 |
| `speaker_notes` | Slide speaker notes or `speaker_note` blocks. | Yes | Reporting `speaker_notes_chars_per_slide` | Reporting `speaker_notes_chars_per_slide` |
| `authored_text` | `paragraph` block with `content_class='presentation_text'`. | Yes | 2000 | 5000 |

**REQ-RC-052**
Every authored text object MUST carry a non-empty `disclosure_partition_ref` that matches a Reporting disclosure partition reference other than `blocked`. The builder MUST NOT infer a partition from the current user, incident role, selected recipient, visible section, or preview mode.

**REQ-RC-053**
`title_override` text MUST NOT contain LF and MUST NOT exceed 120 Unicode scalar values before or after subject placeholder substitution. A title-limit, empty-title, control-character, or LF failure MUST use validation code `authored_title_limit_exceeded`.

**REQ-RC-054**
`authored_text` text MUST NOT exceed the operative Reporting `composition.authored_text_chars` limit before or after subject placeholder substitution. The default is `2000` Unicode scalar values and the hard limit is `5000`. A template may declare a higher value only through Reporting `declared_limits` and only up to the hard limit. A text-limit, empty-body, control-character, or invalid-LF failure MUST use validation code `authored_text_limit_exceeded`.

**REQ-RC-055**
`speaker_notes` text MUST satisfy the Reporting `speaker_notes_chars_per_slide` limit before and after subject placeholder substitution. A speaker-note limit, empty-body, control-character, or invalid-LF failure MUST use validation code `authored_text_limit_exceeded`.

**REQ-RC-056**
Inline subject placeholders MUST use the exact form `{{subject:<stable_subject_ref>}}`. The placeholder start token is the exact ten-character sequence `{{subject:`. The placeholder end token is the exact two-character sequence `}}`.

**REQ-RC-057**
Any substring that starts with `{{subject:` and does not terminate with the next `}}`, or whose interior is empty, MUST fail with validation code `authored_subject_ref_unresolved`. Raw subject names, display tokens, email addresses, hostnames, account names, or party labels are not valid subject placeholder references.

**REQ-RC-058**
This revision does not scan authored free text for case facts, conclusions, literal sensitive values, or raw subject values outside the subject placeholder grammar. Builder linting and reviewer guidance MAY warn about possible literal-sensitive or fact-bearing prose, but such linting is non-normative and MUST NOT be treated as release validation or release conformance evidence.

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
| `diagram_source_kind` | string | Yes | No | None | `graph` or `timeline`. `template_static` is future-only for composition declarations. |
| `source_graph_view_id` | identifier | Yes | Yes | None | Required when `diagram_source_kind='graph'`; otherwise `null`. |
| `selection_rule` | object | Yes | No | None | Must be one Reporting Table 15-C selection-rule object. |
| `label_overrides` | array of object | No | No | `[]` | Closed schema in Table 10-B. |
| `layout_mode` | string | No | No | `auto` | Closed values: `auto`, `manual`. |
| `layout` | `composition_diagram_layout.v1` | No | Yes | `null` | Must be `null` when `layout_mode='auto'`; required when `layout_mode='manual'`. |

**REQ-RC-060a**
The diagram designer edits selection, labels, and closed layout data, never generated source. A composition diagram MAY define selected graph or timeline items, presentation labels, exact node positions, and manual edge routes through closed schema objects. These objects are presentation data. They MUST NOT create vertices, create edges, mutate graph projection output, edit workbook records, edit generated Mermaid, edit rendered SVG, edit Slidev source, or alter release bytes. Reporting owns deterministic rendering from the resolved post-redaction diagram model.

**REQ-RC-061**
When `replaces_decl_id` is non-null, it MUST equal `decl_id` and MUST resolve to exactly one template-owned diagram declaration. A no-match target MUST fail with validation code `composition_replacement_target_missing`. Replacing a composition-owned declaration from the same composition is invalid. Replacing a template declaration under a different `decl_id` is future-only.

**REQ-RC-062**
`selection_rule` MUST be a closed Reporting diagram selection-rule object compatible with `diagram_source_kind`. When `diagram_source_kind='graph'`, `source_graph_view_id` MUST be non-null and the selection rule MUST be `explicit_refs`, `neighborhood`, or `all_with_bounds`. When `diagram_source_kind='timeline'`, `source_graph_view_id` MUST be `null` and the selection rule MUST be `timeline_sequence`. Raw Mermaid text, raw graph query text, arbitrary vertex declarations, arbitrary edge declarations, renderer syntax, and source-record mutation instructions are invalid and MUST use validation code `raw_generated_source_invalid` when detected by the authoring server.

**Table 10-B. Label Override Schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `target` | `composition_label_override_target.v1` | Yes | No | None | Closed structured target from Table 10-C. |
| `label` | string | Yes | No | None | Single-line presentation label. |

**Table 10-C. `composition_label_override_target.v1`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `target_kind` | string | Yes | No | None | Closed values: `vertex`, `edge`. |
| `ref` | string | Yes | No | None | Exact Reporting-resolved vertex or edge reference under the selected diagram. It is compared as an opaque string and MUST NOT be parsed by delimiters. |

**REQ-RC-063**
Label override `label` MUST NOT contain LF, MUST satisfy Reporting diagram label bounds, and MUST be presentation text only. A label override targeting a vertex mapped to a tokenized subject is invalid and MUST use validation code `diagram_label_override_invalid`.

**REQ-RC-063a**
Composition label override targets MUST use `composition_label_override_target.v1`. A target object that resolves to zero or more than one selected diagram item MUST fail with validation code `diagram_selection_missing_ref`. The authoring server MUST NOT accept colon-delimited target strings, generated Mermaid IDs, SVG element IDs, rendered image IDs, or DOM selectors as label override targets.

**REQ-RC-063b**
`layout_mode='auto'` preserves Reporting's Mermaid auto-layout path and requires `layout=null`. `layout_mode='manual'` is valid only when `diagram_kind='flowchart'` and requires a non-null `composition_diagram_layout.v1`. Manual layout MAY apply to graph-derived or timeline-derived flowcharts after selection resolves retained vertex and edge refs. Manual layout MUST fail with validation code `manual_layout_not_supported_for_output_kind` when requested for an output kind whose Reporting renderer cannot honor exact positions. In this revision, `output_kind='mermaid'` cannot honor exact positions.

**Table 10-D. `composition_diagram_layout.v1` schema**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `composition_diagram_layout.v1`. |
| `coordinate_space` | object | Yes | No | None | Closed coordinate-space object from Table 10-E. |
| `node_positions` | array of object | Yes | No | `[]` | Closed node-position objects from Table 10-F, sorted bytewise ascending by `target_ref`. |
| `edge_routes` | array of object | Yes | No | `[]` | Closed edge-route objects from Table 10-G, sorted bytewise ascending by `target_ref`. |

**Table 10-E. Layout coordinate space**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `unit` | string | Yes | No | `css_px` | Exact `css_px`. |
| `origin` | string | Yes | No | `top_left` | Exact `top_left`. |
| `width` | `positive_integer` | Yes | No | None | Must be in `[1, 10000]`. |
| `height` | `positive_integer` | Yes | No | None | Must be in `[1, 10000]`. |

**Table 10-F. Layout node position**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `target_ref` | string | Yes | No | None | Exact opaque retained selected vertex ref. |
| `x` | `finite_integer` | Yes | No | None | Top-left x coordinate; must be `>=0`. |
| `y` | `finite_integer` | Yes | No | None | Top-left y coordinate; must be `>=0`. |
| `width` | `positive_integer` | Yes | No | None | Node box width. |
| `height` | `positive_integer` | Yes | No | None | Node box height. |

**Table 10-G. Layout edge route**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `target_ref` | string | Yes | No | None | Exact opaque retained selected edge ref. |
| `route_kind` | string | Yes | No | `polyline` | Exact `polyline`. |
| `waypoints` | array of object | Yes | No | `[]` | Interior waypoints from Table 10-H in authored order; maximum `32` items. |

**Table 10-H. Layout waypoint**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `x` | `finite_integer` | Yes | No | None | Coordinate-space x; must be `>=0`. |
| `y` | `finite_integer` | Yes | No | None | Coordinate-space y; must be `>=0`. |

**REQ-RC-063c**
Manual layout MUST resolve after diagram selection against the retained diagram item set. Every retained selected vertex MUST have exactly one `node_positions[]` item; omission fails with `diagram_layout_missing_node_position`. Duplicate `node_positions[].target_ref` or duplicate `edge_routes[].target_ref` values fail with `diagram_layout_duplicate_target`. A node placement for an unknown, non-retained, or non-vertex ref fails with `diagram_layout_unknown_target`. An edge route for an unknown, non-retained, or non-edge ref fails with `diagram_layout_unknown_target`. `edge_routes[]` MAY be a subset of retained selected edges; an omitted route uses Reporting's deterministic default straight routing.

**REQ-RC-063d**
All manual layout coordinates and dimensions MUST be JSON integers; floating-point numbers, exponent notation, decimal notation, and `-0` are invalid under REQ-RC-019a and REQ-RC-040. A node box MUST fit inside the declared coordinate space: `x + width <= coordinate_space.width` and `y + height <= coordinate_space.height`. Every waypoint MUST fit inside the declared coordinate space: `x <= coordinate_space.width` and `y <= coordinate_space.height`. Negative coordinates are invalid. Manual edge routes are presentation routes for selected existing edges only; they MUST NOT create graph edges, remove graph edges, change source endpoints, change target endpoints, or target generated Mermaid IDs, SVG IDs, DOM IDs, labels, array indexes, React Flow IDs, or React Flow handles. A shape, coordinate, bounds, ordering, route-kind, endpoint, or layout-mode violation fails with `diagram_layout_invalid` unless a narrower layout validation code is defined in Table 12-E.

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
| `override_diagram_labels` | `diagram_anchor`, `label_overrides` | None | Reporting applies valid label overrides before diagram serialization. |

**REQ-RC-065a**
Operation payload validation MUST use Table 11-C after closed-object and required-member validation. A payload that violates the duplicate, bounds, target-ownership, or compatibility rule in the table MUST fail with validation code `composition_schema_invalid` unless a narrower validation code in Table 12-E applies.

**Table 11-C. Operation payload closure**

| `op_kind` | Additional payload rules |
| --- | --- |
| `exclude_section` | `section_anchor` MUST resolve to one emitted section. Required template sections still use Reporting required-section behavior after removal. |
| `reorder_sections` | `section_anchors[]` MUST be non-empty. Duplicate anchors are detected by canonical anchor JSON after defaults materialize and are invalid. Unlisted sections retain derived relative order. |
| `override_slide_layout` | `layout_id` MUST be a layout token declared by the bound template version. |
| `override_title` | `authored_text_ref` MUST resolve to one `authored_text.v1` with `text_role='title_override'`. |
| `set_speaker_notes` | `authored_text_ref` MUST resolve to one `authored_text.v1` with `text_role='speaker_notes'`. |
| `insert_authored_block` | `position` MUST be `before` or `after`; `authored_text_ref` MUST resolve to one `authored_text.v1` with `text_role='authored_text'`. |
| `exclude_block` | `block_anchor` MUST resolve to one derived block before chunking. |
| `override_click_profile` | `click_profile` MUST be `none`, `reveal_blocks`, or `reveal_list_items`. |
| `insert_diagram_slide` | `diagram_anchor` MUST resolve to one composition-owned diagram declaration. The referenced section controls insertion position. |
| `exclude_diagram` | `diagram_anchor` MUST resolve to one template-owned diagram declaration. Composition-owned diagrams are excluded by removing or replacing their `composition_diagram_decl.v1` draft item, not by this operation. |
| `override_diagram_labels` | `label_overrides[]` MUST be non-empty, use Table 10-B, and contain no duplicate `target` objects by canonical target JSON. |

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

**REQ-RC-071a**
`exclude_diagram` is invalid for a composition-owned diagram declaration. `insert_diagram_slide` is invalid for a template-owned diagram declaration. These ownership mismatches MUST fail with validation code `composition_schema_invalid`.

**REQ-RC-071b**
Composition operations MUST be interpreted in `deck_ops[]` array order against the current retained render model. Validation and render effects MUST use Table 11-D. A conforming implementation MUST NOT reorder operations by `op_id`, target type, creation time, UI grouping, or storage order.

**Table 11-D. Operation sequencing and conflict closure**

| Conflict or repeated operation | Required behavior |
| --- | --- |
| Repeated `exclude_section`, `exclude_block`, or `exclude_diagram` for the same target | The first operation that resolves removes the target. A later operation resolves against the current retained model; zero-match uses its own `on_unresolved` value, and ambiguous match always fails. |
| Repeated `reorder_sections` | Each reorder applies to the current retained section list. Listed retained sections move first in listed order; unlisted retained sections keep their current relative order. |
| Multiple `insert_authored_block` operations at the same `block_anchor` and `position` | All valid inserts are emitted in `deck_ops[]` order relative to that anchor. For `before`, earlier operations appear closer to the start of the block sequence; for `after`, earlier operations appear closer to the anchor. |
| Multiple `insert_diagram_slide` operations at the same `section_anchor` and `position` for different diagram declarations | All valid inserts are emitted in `deck_ops[]` order relative to that section. For `before`, earlier operations appear closer to the start of the section sequence; for `after`, earlier operations appear closer to the anchor. |
| Repeated `insert_diagram_slide` for the same composition-owned diagram declaration | Invalid; MUST fail with validation code `composition_duplicate_diagram_insert`. |
| Repeated scalar overrides for title, layout, speaker notes, click profile, or diagram label target | Later valid operation wins. Earlier overridden values MUST NOT appear in the derived deck or diagram model. |
| Duplicate targets inside one `override_diagram_labels.payload.label_overrides[]` array | Invalid under REQ-RC-065a and REQ-RC-071. |
| Same diagram-label target overridden by separate later operations | Later valid operation wins after all per-operation duplicate checks pass. |
| Any operation whose anchor resolves to more than one target | Fails with `composition_anchor_ambiguous` regardless of `on_unresolved`. |
| Any operation whose anchor resolves to zero targets | `on_unresolved='drop'` may drop the operation only for non-external validation/render; otherwise fails with `composition_anchor_unresolved` or `composition_drop_invalid_for_external_release`. |

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
| `validation_context` | `cartulary.report_composition_validation_context.v1` | No | Yes | `null` | `null` means local-only validation; non-null enables snapshot-dependent and release-scope validation. |

**REQ-RC-073a**
Validation source selection MUST use Table 12-B1. A request that violates a required or forbidden member rule in Table 12-B1 MUST fail with validation code `composition_source_invalid`.

**Table 12-B1. Validation source-kind matrix**

| `source_kind` | Required members | Forbidden members | Source selected |
| --- | --- | --- | --- |
| `draft` | None beyond defaults. | `composition_version`, `inline_composition`. | Current materialized draft for the route resource. |
| `version` | `composition_version`. | `inline_composition`. | Immutable version under the route resource. |
| `inline` | `inline_composition`. | `composition_version`. | Supplied canonical composition document after route/resource binding checks. |

**REQ-RC-073b**
For `source_kind='inline'`, the supplied `inline_composition` MUST carry `incident_id`, `composition_id`, `template_id`, and `template_version` equal to the route incident and resource. A mismatch MUST fail with the most specific binding code among `composition_incident_mismatch`, `composition_id_mismatch`, and `composition_template_mismatch`.

For `source_kind='inline'`, the supplied `inline_composition.composition_sha256` MUST equal the REQ-RC-041 digest over the supplied canonical bytes after deleting only the `composition_sha256` member. A mismatch MUST fail with validation code `composition_digest_mismatch`. Inline validation MUST NOT persist the document, allocate a version, mark a version release-bound, or let an inline digest satisfy a release tuple.

**REQ-RC-073c**
When `validation_context=null`, validation is schema, resource-binding, digest, template-vocabulary, and locally attributable validation only. The summary MUST NOT represent the composition as admissible for release. Snapshot-dependent anchor resolution, authored-subject placeholder resolution, graph projection binding, recipient partition checks, and external-release authored-text permission checks MUST be skipped rather than approximated.

A non-null validation context MUST use Table 12-B2. Unknown members are invalid. A context whose `release_scope='external_release'` MUST include every required non-null member, MUST have `recipient_partition_refs[]` non-empty, and MUST use a non-null `redaction_profile_sha256`. Omission of the context for a request that asks to validate external-release admissibility, or omission of any required external-release context member, MUST fail with validation code `composition_validation_context_missing`. A malformed context member, invalid tuple ordering, duplicate graph projection `graph_view_id`, or template/resource mismatch MUST fail with validation code `composition_validation_context_invalid`.

**Table 12-B2. `cartulary.report_composition_validation_context.v1`**

| Member | Type | Required | Nullable | Default | Rule |
| --- | --- | ---: | ---: | --- | --- |
| `schema_id` | string | Yes | No | None | Exact `cartulary.report_composition_validation_context.v1`. |
| `release_scope` | string | Yes | No | None | Closed values: `internal_draft`, `internal_review`, `external_release`. |
| `snapshot_id` | identifier | Yes | No | None | Immutable snapshot used for anchor, subject, timeline, and graph validation. |
| `derivation_version` | identifier | Yes | No | None | Reporting derivation version used for target resolution. |
| `template_id` | identifier | Yes | No | None | Must equal the route resource `template_id`. |
| `template_version` | identifier | Yes | No | None | Must equal the route resource `template_version`. |
| `redaction_profile_id` | identifier | Yes | No | None | Reporting redaction profile identity. |
| `redaction_profile_version` | identifier | Yes | No | None | Reporting redaction profile version. |
| `redaction_profile_sha256` | sha256_hex | Yes | No | None | Digest of redaction profile bytes used for authored-text permission checks. |
| `recipient_partition_refs` | array of string | Yes | No | `[]` | Must be `[]` for `internal_draft` and `internal_review`; non-empty and Reporting-valid for `external_release`. |
| `graph_projection_refs` | array of Reporting `source_projection_ref.v2` | Yes | No | `[]` | Sorted bytewise by `graph_view_id`; duplicate `graph_view_id` values invalid. |
| `output_kind` | string | Yes | No | None | Closed Reporting output kind: `mermaid` or `slidev`. |
| `output_options` | object | Yes | No | Reporting §7.5 defaults | Materialized Reporting output options. |
| `render_environment_profile_id` | identifier | Yes | No | None | Template-declared Reporting render profile used for validation. |

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

**REQ-RC-075a**
Validation issues MUST sort by Table 12-D1 for `issues[]`, `stage`, and first-failure selection. A conforming implementation MUST NOT sort issues by database order, JSON parser member order, UI creation order, map iteration order, localized message text, or renderer diagnostic order.

**Table 12-D1. Composition issue sort key**

| Key position | Sort key | Rule |
| ---: | --- | --- |
| 1 | Stage | Ordered by Table 12-A. |
| 2 | Severity | `error` before `warning`. |
| 3 | Source array | `deck_ops`, then `diagram_decls`, then `authored_texts`, then `diagram_decls.layout.node_positions`, then `diagram_decls.layout.edge_routes`, then `validation_context.graph_projection_refs`, then `validation_context.recipient_partition_refs`, then `unattributed`. |
| 4 | Source index | Zero-based index in the materialized source array; absent index sorts after all numeric indexes. |
| 5 | Stable local identity | `composition_op_id`, then `diagram_id`, then `authored_text_id`, then exact `target_ref`; absent values sort after present values. |
| 6 | Code | Exact validation-code token. |
| 7 | Safe details | Reporting canonical JSON serialization of `safe_details` after forbidden-value filtering. |

**REQ-RC-076**
Validation codes MUST use Table 12-E. A conforming implementation MAY add non-normative warning codes only under an extension namespace beginning with `x_`. It MUST NOT add new error codes without revising this table.

**REQ-RC-076a**
Report Composition consumes no generic Extensions transaction, portability,
backup/restore, Snapshot/Reporting participant, state-presence, or capability
interface. Its authoring and validation boundaries remain those defined by this
NLSpec. For an authoritative preview, Report Composition materializes and
persists the exact `cartulary.report_composition_preview_source.v1`, then
admits `snapshot_reporting.composition_preview_v1` through a narrow
Reporting-owned transactional port. Reporting alone dequeues that job, renders
it as `internal_draft`, and commits its terminal job result and proof. The
composition owner MUST NOT invoke the extension participant, renderer, or
common-job store directly. A preview source MUST never be admitted as approval,
external-release, publication, or release-bundle evidence.

Every composition schema-validation rule MUST nevertheless carry its
complete condition annotation, and every procedural composition validator MUST
publish a closed decision table for the generated Extensions
validation-condition registry. A reachable undeclared condition, an extra or
stale declaration, or an emitted condition absent from that registry is a
conformance failure. When a composition validator is invoked through a shared
validation boundary, result selection uses invocation failure, structural
invalidity, overflow at `4097+` findings, remaining schema defects, valid
findings, then valid empty result; `257..4096` is the ordinary 256-item bound
violation.

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
| `composition_source_invalid` | `error` | `schema_validation` | Validation or preview source-kind members violate the closed source-kind matrix. |
| `composition_validation_context_missing` | `error` | `resource_binding_validation` | External-release validation omits the required validation context or one of its required members. |
| `composition_validation_context_invalid` | `error` | `resource_binding_validation` | Validation or preview render context is malformed, mismatched to the resource, or has invalid tuple ordering or duplicate graph bindings. |
| `composition_resource_retired` | `error` | `resource_binding_validation` | Request attempts to mutate, freeze, or preview a retired composition resource. |
| `composition_anchor_invalid` | `error` | `anchor_validation` | Anchor grammar or disallowed target kind is invalid. |
| `composition_anchor_unresolved` | `error` | `anchor_validation` | Anchor resolves to zero targets and `on_unresolved='fail'` or release scope forbids drop. |
| `composition_anchor_ambiguous` | `error` | `anchor_validation` | Anchor resolves to more than one target. |
| `composition_drop_invalid_for_external_release` | `error` | `anchor_validation` | `on_unresolved='drop'` appears for an external-release validation or render. |
| `composition_duplicate_diagram_insert` | `error` | `operation_validation` | More than one operation inserts the same composition-owned diagram declaration. |
| `authored_text_not_permitted` | `error` | `authored_text_validation` | External release would include authored text without profile permission. |
| `authored_title_limit_exceeded` | `error` | `authored_text_validation` | `title_override` exceeds its limit or contains LF. |
| `authored_text_limit_exceeded` | `error` | `authored_text_validation` | `authored_text` or `speaker_notes` exceeds applicable limit. |
| `authored_subject_ref_unresolved` | `error` | `authored_text_validation` | Placeholder is malformed, unresolved, ambiguous, filtered, or unreleasable. |
| `raw_generated_source_invalid` | `error` | `diagram_validation` | Raw Markdown, Mermaid, HTML, renderer syntax, arbitrary node, or arbitrary edge input is present. |
| `diagram_label_override_invalid` | `error` | `diagram_validation` | Label override targets a tokenized subject vertex or violates label constraints. |
| `diagram_selection_missing_ref` | `error` | `diagram_validation` | Diagram selection references unavailable snapshot or projection data. |
| `composition_replacement_target_missing` | `error` | `diagram_validation` | A non-null `replaces_decl_id` does not resolve to exactly one template-owned diagram declaration. |
| `diagram_layout_invalid` | `error` | `diagram_validation` | Manual layout object is malformed, out of bounds, incompatible with diagram kind, or attempts to change edge semantics. |
| `diagram_layout_missing_node_position` | `error` | `diagram_validation` | Manual layout omits a required retained selected vertex placement. |
| `diagram_layout_duplicate_target` | `error` | `diagram_validation` | Manual layout contains duplicate node placement or edge route targets. |
| `diagram_layout_unknown_target` | `error` | `diagram_validation` | Manual layout target does not resolve to one retained selected vertex or edge as required by the target array. |
| `manual_layout_not_supported_for_output_kind` | `error` | `diagram_validation` | Manual layout is requested for an output kind whose Reporting renderer cannot honor exact positions. |

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
| `limit` | `finite_integer` | Numeric limit that was exceeded. |
| `source_kind` | string | `draft`, `version`, or `inline` when a source-kind rule is attributable. |
| `release_scope` | string | `internal_draft`, `internal_review`, or `external_release` when a release-scope rule is attributable. |
| `graph_view_id` | identifier | Safe graph view identity when a validation-context graph binding issue is attributable. |
| `layout_mode` | string | `auto` or `manual` when a diagram-layout rule is attributable. |
| `target_ref` | string | Safe opaque vertex or edge reference when a diagram-layout target issue is attributable. |
| `output_kind` | string | Reporting output kind when a layout/output compatibility issue is attributable. |

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

**REQ-RC-083a**
The builder UI MAY expose generated Mermaid source as read-only diagnostic text for auto-layout diagrams. The builder UI MAY offer a local, non-persisted Mermaid scratchpad only when saving requires conversion into `composition_diagram_decl.v1` and `composition_diagram_layout.v1` objects and discards the raw Mermaid text. A conforming implementation MUST NOT persist raw Mermaid text, raw Mermaid fragments, Mermaid init blocks, Mermaid comments, Mermaid styling, Mermaid click actions, generated `.mmd` bytes, or renderer-local syntax as composition source.

# 14. Behavioral Acceptance Scenarios

**REQ-RC-084**
Tables 14-A and 14-B summarize observable behavior for implementation and
review. They do not require a fixture corpus, one test per row, stable test
identities, or a requirement-to-test completeness map. Current behavior without
observable evidence MUST be implemented and tested, or this NLSpec MUST be
revised before that behavior is removed or moved to a future profile.

**Table 14-A. Behavior Scenarios**

| Scenario label | Scenario | Pass condition |
| --- | --- | --- |
| `RC-FIX-001` | Empty valid composition with no operations, no diagrams, and no authored text. | Canonical bytes are deterministic and digest excludes only `composition_sha256`. |
| `RC-FIX-002` | Same composition serialized with non-canonical member ordering and with omitted defaults versus explicit defaults. | Server canonicalization produces the same digest as the canonical fixture. |
| `RC-FIX-003` | Duplicate `op_id`, `decl_id`, and `authored_text_id` cases. | Each duplicate fails with `composition_duplicate_id`. |
| `RC-FIX-004` | Draft update with stale `base_draft_version`. | Request fails with `composition_draft_version_conflict` and does not mutate draft state. |
| `RC-FIX-005` | Immutable version mutation and release-bound delete attempts. | Requests fail with `composition_version_immutable` or `composition_version_bound`. |
| `RC-FIX-006` | Every `op_kind` with valid payload and one closed-schema violation. | Valid payloads pass schema validation; extra or missing payload fields fail with `composition_schema_invalid`. |
| `RC-FIX-007` | Semantic anchors for resolved, unresolved, ambiguous, and generated-ID targets. | Generated IDs fail with `composition_anchor_invalid`; unresolved and ambiguous anchors use exact codes. |
| `RC-FIX-008` | `on_unresolved='drop'` internal draft and external release validation. | Internal draft may drop zero-match operations; external release fails with `composition_drop_invalid_for_external_release`. |
| `RC-FIX-009` | Authored text role, NFC normalization, control-character, LF, before/after substitution limit, partition, and placeholder cases. | Each invalid case fails with the exact authored-text validation code. |
| `RC-FIX-010` | Diagram declarations with valid selection, raw Mermaid, arbitrary nodes, arbitrary edges, and tokenized-subject label override. | Valid declaration passes; invalid cases fail with `raw_generated_source_invalid` or `diagram_label_override_invalid`. |
| `RC-FIX-011` | Route authorization matrix for viewer, editor, reviewer, admin, no membership, and deployment-admin-only user. | Minimum roles in Table 6-A are enforced and deployment-admin-only access is rejected. |
| `RC-FIX-012` | Builder preview request with current draft. | Request produces `cartulary.report_composition_preview_view.v1`, materializes `cartulary.report_composition_preview_source.v1`, and produces no immutable release bytes. |
| `RC-FIX-013` | Route request and canonical composition with duplicate JSON members, explicit invalid nulls, and invalid `positive_integer` values. | Duplicate members and invalid nulls fail with `composition_schema_invalid`; invalid `positive_integer` values fail before side effects. |
| `RC-FIX-014` | Retire resource, then read, update, freeze, and preview the retired resource. | Read succeeds; update, freeze, and preview fail with `composition_resource_retired`; idempotent retire replay preserves `retired_at`. |
| `RC-FIX-015` | Validation source-kind matrix covering `draft`, `version`, `inline`, forbidden member combinations, null validation context, and external-release context omission. | Valid combinations select the expected source; forbidden combinations fail with `composition_source_invalid`; null context produces local-only validation; missing external context fails with `composition_validation_context_missing`. |
| `RC-FIX-016` | Diagram label overrides with structured targets, duplicate targets, raw colon-delimited strings, generated Mermaid IDs, and missing selected refs. | Structured targets pass when resolved once; ambiguous, duplicate, generated, raw string, and missing targets fail with exact validation codes. |
| `RC-FIX-017` | Composition diagram declarations for graph, timeline, replacement targets, and future-only `template_static`. | Graph requires `source_graph_view_id` and graph-compatible selection; timeline requires null graph view and `timeline_sequence`; replacement targets must resolve to one template-owned declaration or fail with `composition_replacement_target_missing`; `template_static` fails in v1. |
| `RC-FIX-018` | Operation payload matrix covering wrong text role, duplicate canonical anchors, template/composition diagram ownership mismatch, and empty label overrides. | Each invalid case fails with `composition_schema_invalid` or the narrower declared code. |
| `RC-FIX-019` | Route idempotency matrix covering exact replay, same-key different body, stale `base_draft_version`, schema failure, not found, authorization failure, and successful mutation. | Outcomes match Table 6-A2; failures have no side effects; exact replay returns the original committed response; same-key different body fails with Core `client_txn_conflict`. |
| `RC-FIX-020` | Full validation context with `internal_draft`, `internal_review`, and `external_release` scopes, recipient partitions, redaction profile digest, graph projection refs, output options, and render profile. | Valid context enables snapshot-dependent checks; null context remains local-only; internal scopes require empty recipient partitions; malformed or duplicate graph bindings fail with `composition_validation_context_invalid`; omitted external context fails with `composition_validation_context_missing`. |
| `RC-FIX-021` | Draft preview and immutable-version preview over the same resource. | Draft preview emits `preview_source_sha256` and `composition_sha256=null`; immutable preview emits both `preview_source_sha256` and immutable `composition_sha256`; preview digest cannot satisfy release binding. |
| `RC-FIX-022` | Operation sequencing fixture with repeated excludes, repeated reorders, same-anchor inserts, later scalar overrides, repeated label overrides, ambiguous anchors, and duplicate diagram insertion. | Operation effects match Table 11-D exactly; duplicate diagram insertion fails with `composition_duplicate_diagram_insert`. |
| `RC-FIX-023` | Validation-issue ordering across source arrays and nested layout entries. | Issue ordering follows Table 12-D1 and is independent of traversal or storage order. |
| `RC-FIX-024` | Manual-layout flowchart with valid coordinate space and node placement for every retained selected vertex. | Layout validates, node placement targets exact retained vertex refs, coordinates fit inside the declared coordinate space, and canonical bytes contain no generated IDs. |
| `RC-FIX-025` | Manual layout with missing node placement, duplicate targets, unknown targets, and non-retained targets. | Cases fail with `diagram_layout_missing_node_position`, `diagram_layout_duplicate_target`, or `diagram_layout_unknown_target` as attributable. |
| `RC-FIX-026` | Manual edge-route subset with endpoint-preserving routes and invalid endpoint-changing attempts. | Valid route subset passes; attempts to create edges, remove edges, change endpoints, or target non-selected refs fail with `diagram_layout_invalid` or `diagram_layout_unknown_target`. |
| `RC-FIX-027` | Generated Mermaid read-only diagnostic and scratchpad save attempt. | Read-only diagnostic persists nothing; scratchpad save persists only closed composition diagram and layout objects; raw Mermaid text or fragments fail with `raw_generated_source_invalid`. |

**Table 14-B. Acceptance Criteria**

| ID | Criterion | Owner section | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `RC-AC-AUTH-001` | Authority boundary is explicit. | §1 | This NLSpec owns authoring schemas and lifecycle only; Reporting owns render effects. | The document states or implies composition authoring owns render bytes or release approval. |
| `RC-AC-SCOPE-001` | Non-goals are closed. | §3 | Raw source editing, workbook mutation, template mutation, post-redaction editing, and cross-template migration are explicitly rejected or future-only. | Any non-goal is admitted as current behavior. |
| `RC-AC-LIFE-001` | Draft/version lifecycle is deterministic. | §5 | Draft updates use `base_draft_version`; freezes produce immutable `vN` versions and digest bytes atomically. | Stale writes mutate state, versions are mutable, or version identifiers are ambiguous. |
| `RC-AC-ROUTE-001` | Route interfaces are unambiguous. | §6 | Each route has method, path, role, idempotency, request, and response shape. | A route omits role, idempotency, or request/response contract. |
| `RC-AC-ROUTE-002` | Route outcomes and idempotency are deterministic. | §6 | Schema failure, authorization failure, not found, stale version, exact replay, idempotency conflict, and successful mutation follow Table 6-A2. | Failed requests mutate state, exact replay revalidates against current state, or same-key different-body requests are accepted. |
| `RC-AC-SCALAR-001` | Scalar and default handling is closed. | §§4, 7 | Duplicate JSON members, invalid nulls, invalid positive integers, unknown members, and omitted defaults behave exactly as specified. | Parser overwrite behavior, implementation-specific integer bounds, or non-materialized defaults affect state or bytes. |
| `RC-AC-RETIRE-001` | Retirement behavior is deterministic. | §§5, 6 | Retired resources remain readable and reject mutation, freeze, and preview with `composition_resource_retired`. | Retired resources disappear from audit reads or accept later changes. |
| `RC-AC-AUTHZ-001` | Authorization uses incident roles only. | §6 | Viewer/editor route gates match Table 6-A and deployment-admin-only access is rejected. | Composition introduces resource-specific ACLs or deployment-admin bypass. |
| `RC-AC-CANON-001` | Canonical bytes and digest are stable. | §7 | Recomputed SHA-256 over canonical bytes excluding only `composition_sha256` matches the stored digest. | Digest depends on insertion order, whitespace, storage metadata, or the digest member itself. |
| `RC-AC-ANCHOR-001` | Anchors are semantic only. | §8 | Section, record, block, and diagram anchors validate by semantic fields and reject generated IDs. | Ordinal paths or generated structural IDs are accepted. |
| `RC-AC-TEXT-001` | Authored text is bounded presentation text. | §9 | Roles, LF rules, limits, partitions, and placeholders validate exactly. | Authored text can enter as free-form facts or unpartitioned content. |
| `RC-AC-DIAGRAM-001` | Diagrams use closed selection declarations. | §10 | Raw Mermaid, arbitrary nodes, arbitrary edges, and tokenized-subject label overrides fail. | Diagram authoring bypasses Reporting selection rules. |
| `RC-AC-DIAGRAM-LAYOUT-001` | Manual node placement is closed composition data. | §10 | A manual-layout diagram persists one schema-valid node placement per retained selected vertex and no generated IDs. | Coordinates persist as React Flow state, DOM IDs, SVG IDs, Mermaid IDs, labels, or array indexes. |
| `RC-AC-DIAGRAM-LAYOUT-002` | Manual edge routing cannot invent facts. | §10 | Edge routes target only selected retained edge refs and cannot change semantic endpoints. | A route creates an edge, removes an edge, changes endpoints, or targets non-selected refs. |
| `RC-AC-DIAGRAM-LAYOUT-004` | Direct Mermaid editing is not a composition source. | §§10, 13 | Raw Mermaid can be read-only or local scratch only; persisted composition contains only closed schema objects. | Raw `.mmd` text or fragments persist into composition bytes or release-bound input. |
| `RC-AC-OPS-001` | Operation schema is closed. | §11 | Every operation accepts only its declared payload and valid local references. | Unknown payload members or wrong text-role references pass validation. |
| `RC-AC-OPS-002` | Operation conflicts are closed. | §11 | Repeated operations, inserts, scalar overrides, duplicate label targets, ambiguous anchors, and duplicate diagram inserts follow Table 11-D. | Operation results depend on storage order, UI grouping, or implementation-local conflict rules. |
| `RC-AC-VALID-001` | Validation output is machine-testable. | §12 | Validation summaries use declared stages, issue codes, and safe details only. | Validation emits raw sensitive values or undocumented codes. |
| `RC-AC-VALID-002` | Validation source semantics are closed. | §12 | `draft`, `version`, and `inline` validation requests follow Table 12-B1 and external-release validation requires a full validation context. | A validation request silently ignores forbidden members or claims release validity without required context. |
| `RC-AC-VALID-003` | Validation context and issue ordering are closed. | §12 | `cartulary.report_composition_validation_context.v1` carries the full render context, null means local-only validation, `internal_review` behaves as an internal scope, malformed context fails with exact codes, and issues sort by Table 12-D1. | Implementations infer snapshot, redaction, graph, output, or render-profile context from live state or partial request fields, or first failure depends on traversal order. |
| `RC-AC-VALID-004` | Extension validation-condition accounting is complete without adding a generic participant. | §12 | Every reachable schema/procedural condition is declared and registered, result precedence and finding bounds are exact, and all generic participant surfaces remain unconsumed. | A validator emits an undeclared condition, classifies an ambiguous result differently, or Composition acquires an implicit generic participant. |
| `RC-AC-BUILDER-001` | Builder UI is a composition-data author. | §13 | Builder emits route requests, validation requests, and preview requests only. | Builder edits generated source, workbook records, templates, or release bytes. |
| `RC-AC-PREVIEW-001` | Preview boundary is Reporting-owned. | §13 | Authoritative preview creates an `internal_draft` Reporting attempt, not approvable release bytes. | Client preview or builder bytes are treated as reviewable output. |
| `RC-AC-PREVIEW-002` | Preview source digest bytes are closed. | §6 | `cartulary.report_composition_preview_source.v1` contains materialized draft/version arrays and binds by `preview_source_sha256`; immutable preview keeps `composition_sha256` separate. | Draft previews reuse release `composition_sha256` or omit materialized source arrays from the digest input. |
| `RC-AC-BEHAVIOR-001` | Current behavior has durable evidence or a prior specification disposition. | §14 | Production-facing tests or public targets prove retained behavior without relying on IDs, counts, or status rows. | A current behavior is represented only by accounting metadata or an executable placeholder. |

# 15. Promotion Checklist

**REQ-RC-085**
This NLSpec MUST NOT be promoted to `adopted/current` until every row in Table 15-A is satisfied.

**Table 15-A. Promotion Conditions**

| Condition | Required state |
| --- | --- |
| Domain vocabulary | `docs/domain.md` includes the composition terms introduced by this NLSpec with extension-profile status. |
| Reporting import boundary | `docs/reporting-subsystem-nlspec.md` imports `cartulary.report_composition.v1` and the companion identifiers without owning authoring routes or schema. |
| Core route conventions | Core 01 route envelope, idempotency, and common error conventions are available for this route family. |
| Core preview job convention | Core 01 exposes the job or attempt identity fields used by `cartulary.report_composition_preview_view.v1`. |
| Core authorization conventions | Core 04 incident role derivation and deployment-admin bypass prohibition are available for this route family. |
| Core redaction profile convention | Core 04 exposes `allow_authored_presentation_text` and release-scope redaction profile identity, version, and digest for Reporting validation. |
| Reporting preview import | `docs/reporting-subsystem-nlspec.md` accepts `cartulary.report_composition_preview_source.v1` only for `internal_draft` preview and never for external release. |
| Graph Projection dependency | Core 01 and the adopted Graph Projection owner expose completed digest-bound `graph_projection_refs[]` required by graph composition diagrams. |
| Behavioral disposition | Every current behavior is retained and routed, implemented and tested, or revised before being pruned or moved to a future profile. |

# 16. Non-Normative Source Notes

This section is non-normative. It records supporting inputs used to shape this revision. It does not create requirements, defaults, exceptions, or conformance evidence beyond the normative sections above.

**Table 16-A. Supporting research notes**

| Source | Material use in this NLSpec revision | Boundary |
| --- | --- | --- |
| `docs/research/R01-aurora_incident_response_report.md` and `docs/research/R03-Kanvas_technical_research_report.md` | Supported keeping preview/render behavior bound to Reporting-owned output paths rather than builder-local generated source or live report views. | Normative closure remains in §§6, 12, and the Reporting Subsystem NLSpec. |
| `docs/research/R04-responsive_browser_spreadsheet_ui_research_memo.md` and `docs/research/R05-responsive-interface-design-report.cr.md` | Supported explicit local-only versus release-context validation states, server-authoritative preview, and stale-result avoidance. | UI guidance remains non-normative; conformance is only the route, validation, and preview contract. |
| `docs/research/R06-spreadsheet_of_doom_dfir_research_report.md` and `docs/research/R07-spreadsheet-of-doom-sod-report.cr.md` | Supported keeping composition as presentation data over immutable snapshots rather than raw evidence, case facts, or tracker source-of-truth. | Raw evidence and durable case facts remain outside composition source. |
| `docs/research/R02-cartulary_crm_tem_dfir_research_report.md` | Supported separating product behavior from training cadence, report-preparation playbooks, and operator coaching. | Playbooks and SOPs do not create composition conformance behavior. |
| `docs/research/R08-handsontable-react-research-report.md` and `docs/research/R09-react-data-grid-research-report.md` | Reviewed for builder/UI boundary risks around row identity, controlled state, and framework wrappers. | UI implementation technology remains outside this NLSpec unless it affects emitted route contracts. |
