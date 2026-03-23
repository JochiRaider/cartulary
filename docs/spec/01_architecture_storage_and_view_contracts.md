# Cartulary Normative Core 01: Architecture, Storage, and View Contracts

## 1. Architecture pattern

Cartulary MUST use a **modular monolith** architecture for the base profile.

The base deployment topology MUST consist of:

- one web application deployable that contains the browser-facing UI, API surface, WebSocket hub, and background-job runners,
- one Postgres service as the authoritative structured data store,
- one S3-compatible object storage service as the authoritative binary evidence store.

The application deployable MUST remain a single deployable unit even when deployed behind a reverse proxy or onto managed infrastructure.

Microservice decomposition is out of scope for current conformance.

## 2. Required modules and boundaries

The implementation MUST define internal boundaries equivalent to the following concerns:

- authentication and session management,
- incidents and memberships,
- timeline capture,
- entities, indicators, and observation resolution,
- evidence and object storage,
- imports and tabular ingest,
- links, tags, and analyst-work coordination,
- revisions and rollback,
- projections and search,
- reference data,
- reporting and snapshot generation,
- collaboration and presence.

These are logical module boundaries. They MUST be independently testable, but they MUST NOT require separate deployables.

File-based structured import beyond clipboard paste MUST be implemented as a dedicated internal `imports` module within the modular monolith.

The `imports` module MUST own, at minimum:

- file-based source adapters for CSV and XLSX input,
- workbook inspection, candidate-region selection, preview, and header mapping,
- import job execution with progress, cancellation, retry-safe status, and diagnostics,
- deterministic import provenance capture,
- compatibility shims for spreadsheet-specific parser behavior and workbook-shape heuristics.

Clipboard interaction remains part of the base workbook surface. When clipboard paste feeds structured ingest, it MUST use the same stable tabular-ingest contract and shared mapping engine as file-based import paths.

The workbook, timeline, entities, evidence, revisions, projections, and reporting concerns MUST depend only on the stable tabular-ingest contract and shared mapping engine for structured ingest. They MUST NOT depend directly on XLSX or OpenXML parsers, workbook-specific heuristics, or Excel-specific semantics.

## 3. Client and server responsibilities

### 3.1 Browser client

The browser client MUST provide:

- a virtualized workbook grid,
- keyboard navigation,
- paste handling,
- a detail and relationship inspector,
- evidence preview behavior,
- save/conflict state presentation,
- a local pending-patch queue for transient network interruptions,
- real-time presence and live row updates.

The virtualized workbook grid MUST render only the visible viewport plus a bounded overscan region. It MUST NOT require full-DOM rendering of every row in a 10k+ row incident.

Selection, focus, and pending-edit anchoring in the grid MUST remain bound to the selected `record_id` during live updates, sorting, filtering, and grouping.

### 3.2 Application server

The application server MUST provide:

- authenticated API endpoints,
- authoritative mutation validation,
- optimistic concurrency enforcement,
- projection maintenance,
- WebSocket-based collaboration updates,
- background-job orchestration,
- reference-pack activation verification,
- snapshot and report generation when the Snapshot and Reporting Extension Profile is implemented.

## 4. Storage boundary

### 4.1 Postgres

Postgres MUST store:

- incidents, users, and memberships,
- record envelopes and first-class records,
- canonical indicators, indicator observations, and indicator lifecycle intervals,
- entity aliases and entity mentions,
- record links and record tags,
- change sets, mutation entries, and row revisions,
- saved views and view schemas,
- projection tables,
- blob metadata,
- evidence lifecycle metadata,
- reference-pack manifests and integrity metadata,
- snapshot metadata, canonical export-model metadata, versioned template-contract metadata, versioned redaction-profile metadata, redaction manifests, and artifact release records when the Snapshot and Reporting Extension Profile is implemented.

### 4.2 Object storage

S3-compatible object storage MUST store:

- binary evidence payloads,
- optionally, generated export artifacts when the Snapshot and Reporting Extension Profile is implemented.

The object store implementation MAY be MinIO in flyaway or on-prem deployments. Cloud deployments MAY use native S3, GCS, or Azure Blob behind an equivalent abstraction.

### 4.3 Storage exclusions

Large binary evidence MUST NOT be stored inline in Postgres.

A filesystem-backed blob adapter MAY be used for development or very small laboratories. It MUST NOT replace S3-compatible storage as the default production target.

## 5. Incident data versus reference packs

Cartulary MUST distinguish **incident data** from **reference packs**.

The following are incident data:

- incident records,
- evidence envelopes,
- revisions,
- saved views,
- immutable report snapshots.

The following are reference packs:

- framework mappings,
- type and icon registries,
- evidence vocabularies,
- optional enrichment datasets,
- view contracts when distributed independently of incidents.

Reference packs MUST version independently of incidents.

Reference-pack manifests and integrity metadata MUST be stored in Postgres. Pack payloads MAY live on local disk or object storage behind the same abstraction, but their activation state and import or activation attestation MUST remain queryable from structured metadata.

Reporting template packs MAY reuse the same integrity-verification and distribution machinery as reference packs. The template selected for a specific snapshot, the selected redaction profile, approval state, and rendered output hashes remain incident data.

## 6. View contracts

Each built-in sheet or contract-backed system view MUST be declared by a **`view_schema`** contract.

A `view_schema` contract MUST define, at minimum:

- the stable `view_schema_id`,
- source record types,
- the base projection,
- computed columns,
- required reference packs, if any,
- default sort key and direction,
- filter semantics,
- write-back semantics,
- metadata needed to render the view consistently.

View behavior MUST bind to `view_schema_id`, not to the visible tab label, column header text, or any other display label.

If a visible column header or tab label changes, filter behavior, write-back behavior, and export semantics MUST remain unchanged unless the underlying `view_schema_id` changes.

## 7. Built-in sheets and system views

### 7.1 Built-in sheets

The base profile MUST provide the following built-in workbook sheets:

- Timeline,
- Hosts,
- Identities,
- Evidence,
- Notes.

The built-in sheets MUST be projections or saved/system views over common underlying source tables rather than separate storage silos.

The Hosts sheet MUST support interactive filtering and sorting over `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` without requiring synchronous scans of raw note text or evidence blobs.

The Identities sheet MUST support interactive filtering and sorting over `privilege_level`, `mfa_state`, and `reset_status` without requiring synchronous scans of raw note text or evidence blobs.

The Evidence sheet MUST support interactive filtering and sorting over `requested_at`, `received_at`, `collector_party_text`, `source_party_text`, `storage_ref`, `blob_hash`, and attachment or upload state without requiring synchronous blob access.

### 7.2 Contract-backed system views

The base profile MUST support contract-backed system views for:

- indicators,
- compromise assessments,
- task requests,
- decisions.

Framework overlays such as ATT&CK, D3FEND, or VERIS MAY also be exposed as system views when the relevant reference packs are present.

System views MUST follow the same `view_schema_id` contract discipline as built-in sheets.

The Indicators system view MUST project canonical indicator records. It MUST NOT use source artifacts or source-bound indicator observations as the primary row identity.

The Compromise Assessments system view MUST project incident-scoped assessment records. It MUST NOT collapse assessment history into a mutable static property on a host or identity row.

The Task Requests system view MUST project `task_request` records. It MUST support queue-oriented filtering and sorting over `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `blocked_reason`, `completed_at`, `external_ticket_ref`, and `updated_at` without requiring synchronous scans of raw note or artifact text.

The Decisions system view MUST project `decision` records. It MUST support filtering and sorting over `status`, `owner_user_id`, `decision_type`, and `decided_at` without requiring synchronous scans of raw note or artifact text.

Structured coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MAY be surfaced through contract-backed system views or saved views over `artifact_grid_projection`. They MUST NOT require additional built-in sheets in the base profile.

### 7.3 Notes sheet contract

The base profile MUST expose **Notes** as a built-in workbook sheet.

The built-in Notes sheet MUST:

- be declared by a stable `view_schema_id`,
- use `artifact_grid_projection` as its base projection filtered to `artifact_type='note'`,
- support blank-row or equivalent grid-native note creation from the sheet itself,
- remain backed by the shared artifact model rather than a Notes-specific storage silo.

The base profile MUST also expose contextual `add linked note` actions from Timeline, Hosts, Identities, and Evidence. All Notes entry paths MUST create the same underlying artifact record shape.

Notes behavior MUST NOT depend on the visible tab label. If the implementation allows the built-in Notes tab to be renamed or hidden per user, write-back behavior and export semantics MUST remain unchanged because they are bound to `view_schema_id`.

## 8. Projection model

### 8.1 Projection tables

Hot workbook screens MUST use projection tables rather than Postgres materialized views.

The implementation MUST define projection tables equivalent to:

- `timeline_grid_projection`,
- `host_grid_projection`,
- `identity_grid_projection`,
- `artifact_grid_projection`,
- `evidence_grid_projection`,
- `indicator_grid_projection` for the indicator system view,
- `assessment_grid_projection` for the compromise-assessment system view,
- `task_request_grid_projection` for the task-request system view,
- `decision_grid_projection` for the decision system view.

Each projection row MUST represent exactly one primary record in the base projection for that view.

For `indicator_grid_projection`, the primary record MUST be the canonical indicator record for that row.

For `assessment_grid_projection`, the primary record MUST be the assessment record for that row.

For `task_request_grid_projection`, the primary record MUST be the `task_request` record for that row.

For `decision_grid_projection`, the primary record MUST be the `decision` record for that row.

### 8.2 Projection-row identity

Every projection row exposed to the client MUST carry:

- the stable `record_id` of the underlying mutation target,
- the current `row_version` required for optimistic writes.

The client MUST NOT infer mutation targets from row position, displayed values, group headers, or transient selection state.

### 8.3 Projection maintenance

Projection rows MUST be updated transactionally with the source write that changes them.

Projection tables are derived state and MUST NOT be authoritative history.

The implementation MUST provide a deterministic rebuild command or equivalent maintenance operation that can regenerate projections from source tables.

### 8.4 Projection corruption

If a projection becomes corrupt or stale, the implementation MUST treat the projection as disposable cache state, rebuild it from authoritative source data, and preserve source-of-truth consistency.

### 8.5 Hot-path retrieval and evidence boundary

Hot workbook sheets MUST serve the visible viewport from projection rows and other small derived metadata. They MUST NOT synchronously scan source tables or evidence blobs to render the grid hot path.

Interactive retrieval for hot workbook sheets MUST use a deterministic sort tuple and stable cursor, keyset, or viewport/block retrieval. It MUST NOT rely on deep `OFFSET` pagination for large incidents.

Projection tables MUST support deterministic interactive retrieval over the default sort key and any contract-declared interactive grouping key for the view. An implementation MAY satisfy this with indexes or an equivalent mechanism that preserves the same observable latency envelope.

Projection rows for hot workbook sheets MUST carry the scalar fields required for interactive sort, filter, grouping, selection anchoring, and evidence badges in the visible viewport. For the Timeline sheet, this MUST include at least `sort_ts`, day buckets equivalent to `timeline.occurred_day` and `timeline.recorded_day`, `capture_state`, `has_evidence`, `has_unresolved_mentions`, and `evidence_count`.

For the Hosts sheet, exact lookup, sorting, filtering, and pivot counts over `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` MUST be satisfiable from `host_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text or evidence blobs.

For the Identities sheet, exact lookup, sorting, filtering, and pivot counts over `privilege_level`, `mfa_state`, and `reset_status` MUST be satisfiable from `identity_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text or evidence blobs.

For the Evidence sheet, exact lookup, sorting, filtering, and queue views over `requested_at`, `received_at`, `collector_party_text`, `source_party_text`, `storage_ref`, `blob_hash`, and upload or attachment state MUST be satisfiable from `evidence_grid_projection` and other small derived metadata. They MUST NOT require synchronous blob access.

The grid and inspector hot path MUST synchronously read only scalar fields, flags, counts, and small preview handles needed for the visible viewport or selected row. They MUST NOT synchronously fetch full attachment lists or binary blob bytes as part of grid rendering, row selection, sheet filtering, grouping, or inspector metadata open.

For the Indicators system view, exact lookup, sorting, filtering, and pivot counts over canonical indicators MUST be satisfiable from `indicator_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw timeline text, artifact text, or evidence blobs.

For the Compromise Assessments system view, exact lookup, sorting, filtering, and pivot counts over `assessment_state` and derived `confidence_band` MUST be satisfiable from `assessment_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw timeline text, artifact text, or evidence blobs.

For the Task Requests system view, exact lookup, sorting, filtering, queue counts, and stale-work views over `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `blocked_reason`, `completed_at`, `external_ticket_ref`, and `updated_at` MUST be satisfiable from `task_request_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text, communications logs, or evidence blobs.

For the Decisions system view, exact lookup, sorting, filtering, and review queues over `status`, `owner_user_id`, `decision_type`, `decided_at`, and supersession state MUST be satisfiable from `decision_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text, communications logs, or evidence blobs.

## 9. Canonical derivation layer

Cartulary MUST maintain a single canonical derivation layer for all derived surfaces.

The following surfaces MUST read from the same derivation/query logic or from an explicitly versioned snapshot of that logic:

- interactive workbook views,
- report sections,
- framework rollups,
- future visualizations,
- exported artifacts.

This requirement exists to prevent drift in filtering, counts, inclusion rules, identifier stability, and ordering.

## 10. Snapshot and reporting extension profile

### 10.1 Extension boundary

The **Snapshot and Reporting Extension Profile** is optional for base conformance. If implemented, it MUST satisfy all requirements in this section and the corresponding criteria in Core 04.

### 10.2 Snapshot semantics

A reporting-capable implementation MUST treat report and presentation generation as a subsystem rather than direct ad hoc reads from live workbook tables.

Each rendered artifact MUST bind to one immutable release tuple equivalent to:

- `snapshot_id`,
- `snapshot_at`,
- `source_change_set_high_watermark` or equivalent frozen source boundary,
- `derivation_version`,
- `template_id` and `template_version`,
- `redaction_profile_id` and `redaction_profile_version`,
- `export_model_sha256`.

The implementation MUST:

- capture the `snapshot_at` boundary and frozen source boundary,
- materialize a canonical export model such as `incident_report_model.json`,
- render derivative outputs from that immutable snapshot rather than from mutable live tables.

After the canonical export model has been materialized for a given release tuple, no renderer, template, or presentation builder MAY query live workbook tables or mutable projections for additional case content.

Re-rendering from the same release tuple MUST reproduce the same canonical export model, deterministic ordering, and `export_model_sha256`.

### 10.2.1 Rendered artifact lifecycle

For Snapshot and Reporting Extension Profile outputs, Cartulary defines an artifact-scoped lifecycle over each rendered output candidate.

The authoritative persisted representation for this lifecycle MUST be the release record plus its bound release tuple and any bound approval records. Approval state MUST NOT bind to mutable incident rows or to template metadata outside the release record.

For this lifecycle, the logical output slot is the bound release tuple excluding `output_sha256`.

The closed vocabulary for `release_state` is:

- `pending_approval`,
- `approved`,
- `invalidated`,
- `published`.

A rendered output enters `pending_approval` when render completion has produced bytes and `output_sha256` for one immutable release tuple but the required approvals for the chosen `release_scope` are not yet satisfied.

A rendered output enters `approved` only when the approval requirements in Core 04 §2.1 are satisfied for that exact release record, logical output slot, and `output_sha256`.

A rendered output enters `published` only through an explicit publish or release action after it is already `approved`.

A rendered output enters `invalidated` when any of the following occurs after approval or publication:

- a superseding render is produced for the same logical output slot with a different `output_sha256`,
- the implementation can no longer attest that the required approval set applies to that exact artifact,
- the implementation explicitly marks the artifact as superseded by a newly rendered candidate for the same logical output slot.

Approval invalidation MUST be an explicit lifecycle transition on the artifact record. It MUST NOT be implemented only as an implicit UI rule.

A new render with a different logical output slot or different `output_sha256` MUST start as a distinct `pending_approval` candidate. It MUST NOT inherit `approved` or `published` state from an earlier candidate.

### 10.3 Export-model classification and release scopes

Every exportable field or block in the canonical export model MUST carry exactly one `content_class` with one of the following values:

- `source_evidence` for direct evidence references, hashes, timestamps, filenames or media labels, and exported excerpts or thumbnails,
- `derived_analytic` for deterministic transforms such as timelines, counts, ATT&CK rollups, graphs, and relationship summaries,
- `curated_narrative` for analyst-authored findings prose, executive summaries, recommendations, impact statements, and lessons learned,
- `working_material` for scratch text, unresolved notes, internal comments, and unreviewed excerpts.

The closed vocabulary for artifact release scope is:

- `internal_draft`,
- `internal_review`,
- `external_release`.

The snapshot and export subsystem MUST evaluate output eligibility against the chosen `release_scope` using at least the following matrix:

- `internal_draft`: any `content_class` except raw blob bytes,
- `internal_review`: any `content_class` except raw blob bytes, and any included `working_material` MUST remain visibly marked non-releasable,
- `external_release`: `derived_analytic`, `curated_narrative`, and only selected `source_evidence` excerpts or thumbnails that are eligible for the chosen `release_scope`. Raw blob bytes and `working_material` MUST NOT appear.

For `external_release`, every `curated_narrative` block MUST carry `support_refs[]` containing one or more stable identifiers to supporting findings, events, evidence records, assessments, or query records. A narrative block lacking `support_refs[]` MUST be ineligible for `external_release`.

Content derived directly from ad hoc note artifacts, `task_request` records, `decision` records, `comm_log` artifacts, `handoff` artifacts, `status_review` artifacts, and `lesson` artifacts SHOULD default to `working_material`.

Such content MUST NOT appear in `external_release` unless an analyst has explicitly curated it into a separate export-model block that independently satisfies the selected `content_class`, `support_refs[]`, and applicable redaction rules.

### 10.4 Template packs and rendering contract

A reporting-capable implementation MUST treat report templates as versioned, integrity-checked local asset bundles or equivalent template packs.

Each template contract MUST declare, at minimum:

- `template_id`,
- `template_version`,
- supported `output_kind` values,
- supported `release_scope` values,
- a local asset bundle only,
- section ordering,
- allowed export-model bindings,
- required fields,
- deterministic ordering rules,
- narrative slots that analysts MAY fill explicitly.

Template bundles SHOULD reuse the same integrity-verification and activation machinery as reference packs when that machinery is available. Template activation MUST remain independent of optional reference-pack presence.

A template renderer MUST:

- read only from the immutable export model and declared local assets,
- MUST NOT query live workbook tables or mutable projections,
- MUST NOT fetch network assets,
- MUST NOT execute arbitrary user-supplied code,
- fail closed if the template references an undeclared binding or a missing required field.

### 10.5 Redaction profiles and manifests

A reporting-capable implementation MUST apply redaction to the canonical export model rather than by mutating incident records.

Each versioned redaction profile MUST declare, at minimum:

- `redaction_profile_id`,
- `redaction_profile_version`,
- a default action,
- per-`content_class` rules,
- optional per-field overrides keyed by stable export-model path.

When recipient-specific reporting is implemented, a versioned redaction profile MUST also be able to declare zero or more allowed stable incident-local `disclosure_partition_refs[]`.

The closed vocabulary for redaction actions is:

- `allow`,
- `drop`,
- `mask`,
- `truncate`,
- `hash`,
- `stub`.

Redaction MUST run after snapshot materialization and before template rendering.

If a field or block eligible for the chosen `release_scope` appears in the canonical export model without an applicable redaction rule, rendering MUST fail closed.

When a field or block carries `disclosure_partition_refs[]` that are not allowed by the selected redaction profile, the renderer MUST apply the applicable redaction rule or fail closed. If a field or block contains mixed-partition content and no applicable rule can produce a conformant result, rendering MUST fail closed.

Disclosure partition metadata and redaction-profile selection MUST affect only snapshot-derived rendering and release. They MUST NOT affect live workbook queries, projections, or incident authorization.

Each rendered artifact MUST emit a `redaction_manifest` keyed by stable export-model path and rule identifier, recording every field or block that was dropped, masked, truncated, hashed, or stubbed.

A render that changes `template_id`, `template_version`, `redaction_profile_id`, or `redaction_profile_version` relative to an earlier artifact candidate MUST create a distinct `pending_approval` candidate. It MUST NOT inherit approval or publication state from the earlier artifact.

### 10.6 Output forms and generated-presentation boundary

A reporting-capable implementation MAY generate:

- Markdown reports,
- Mermaid diagram sources,
- Slidev decks,
- HTML reports,
- operator-facing reenactment outputs such as Asciinema-style walkthroughs.

`output_kind` MUST use a stable closed vocabulary equivalent to:

- `html`,
- `markdown`,
- `slidev`,
- `mermaid`,
- `reenactment`.

Generated presentations MAY:

- rearrange and visualize snapshot facts,
- render deterministic summaries from approved export-model fields,
- generate graphs, timelines, and deck structure,
- include analyst-authored narrative blocks,
- include approved evidence excerpts with provenance.

Generated presentations MUST NOT:

- invent new facts, timestamps, actors, or causal chains,
- synthesize commands or terminal sessions that were not observed,
- interpolate missing steps between observed events,
- rewrite unreviewed working material into releasable claims,
- present a reenactment as externally releasable observed operator activity unless the represented steps are themselves explicit evidence.

`mermaid` and `slidev` outputs MAY be `external_release` only when every rendered block satisfies the selected `release_scope`.

`reenactment` outputs MUST be marked `generated_presentation=true` and are limited to `internal_review` in the current profile.

If such outputs are generated, the implementation MUST preserve the distinction between source evidence and generated presentation material.

### 10.7 Self-contained outputs

Generated report and presentation artifacts MUST be self-contained. They MUST NOT require remote JavaScript, CSS, fonts, or runtime media assets at render time.

## 11. Reference Pack Extension Profile

### 11.1 Extension boundary

The **Reference Pack Extension Profile** is optional for base conformance. If implemented, it MUST satisfy this section and the corresponding criteria in Core 04.

### 11.2 Minimum disconnected bundle

For the smallest supported flyaway or disconnected deployment that implements this profile, the deployment MUST preinstall and activate exactly the following three reference packs by default:

- `type_registry.host`
- `type_registry.evidence`
- `type_registry.indicator`

These three packs define the minimum disconnected bundle because base-profile host, evidence, and indicator semantics MUST come from versioned registries rather than hard-coded UI labels or workbook headers.

The smallest supported disconnected bundle MUST NOT require or preinstall framework overlay packs. Current-profile framework add-on pack keys are:

- `framework.attack`
- `framework.d3fend`
- `framework.veris`

The smallest supported disconnected bundle MUST NOT require or preinstall enrichment packs. Current-profile enrichment add-on pack keys include:

- `enrichment.tor`
- `enrichment.cisa_kev`
- `enrichment.ms_portals`
- `enrichment.windows_event_ids`
- `enrichment.entra_app_ids`
- `enrichment.lolbas`
- `enrichment.loldrivers`
- `enrichment.lolesxi`
- `enrichment.hijacklibs`
- `enrichment.windows_sids`

Other enrichment or framework pack keys MAY exist. They MUST follow the same activation, verification, and degradation rules defined by this profile.

If the Snapshot and Reporting Extension Profile is implemented, template bundles MUST remain separately installable. They MUST NOT count toward the minimum disconnected reference-pack bundle.

Separately distributed `view_contract` packs MAY exist in larger deployments. They MUST NOT be required for the smallest disconnected bundle.

Larger supported disconnected bundles MAY preinstall additional packs, but the minimum disconnected bundle is fixed by this subsection.

### 11.3 Offline import, update, and activation flow

In a flyaway or disconnected deployment, reference-pack update MUST use an offline bundle import flow. The running application MUST NOT perform a live internet fetch as part of pack verification or activation.

The import and update flow MUST satisfy all of the following:

1. the operator supplies a pack bundle through the configured reference-pack storage root or an equivalent administrative upload path backed by that root,
2. the system stages the bundle inside the configured temporary-work root,
3. the system verifies the staged bundle before any extracted content becomes active,
4. on successful verification, the system records the candidate version as `available` or an equivalent non-active state,
5. activation requires an explicit operator action that switches the active version pointer for the target `pack_key`.

#### 11.3.1 Linked reference-pack lifecycle machines

For each imported reference-pack version, Cartulary defines two linked lifecycle machines:

- a verification and availability machine authoritative on `reference_packs`,
- an activation machine authoritative on `reference_pack_activation_state` and `reference_pack_attestations`.

These machines are linked but separate. Successful verification does not by itself activate a version, and activation MUST NOT bypass verification.

For the verification and availability machine, the authoritative pack-version conditions are:

- `staged`: `reference_packs.status='staged'` and `verification_result='pending'`,
- `verified_available`: `reference_packs.status='available'` and `verification_result='passed'` and the version is not the active version for its `pack_key`,
- `disabled`: `reference_packs.status='disabled'`,
- `failed`: `reference_packs.status='failed'` or `verification_result='failed'`,
- `missing`: `reference_packs.status='missing'`.

For the activation machine, a pack version is `active` only when `reference_packs.status='available'`, `verification_result='passed'`, and `reference_pack_activation_state.active_version` for the same `pack_key` equals that `pack_version`.

The allowed lifecycle transitions are:

- `staged -> verified_available` only after successful verification,
- `staged -> failed` on failed verification,
- `staged -> missing` when the staged bundle or extracted payload is no longer available before successful verification completes,
- `verified_available -> active` only through explicit activation,
- `active -> verified_available` only when another verified version for the same `pack_key` is explicitly activated,
- `verified_available -> disabled` or `active -> disabled` only through explicit administrative disablement,
- `disabled -> verified_available` only after an explicit administrative re-enable that confirms existing verification metadata still applies or after re-verification succeeds,
- `verified_available -> failed`, `active -> failed`, or `disabled -> failed` only when a later integrity, signature, or contract-compatibility check fails,
- `verified_available -> missing`, `active -> missing`, or `disabled -> missing` only when required payload content is unavailable at use time.

A `failed` or `missing` version MUST NOT become `active` without first returning through `staged` or `verified_available` by a new import or successful re-verification path. A `disabled`, `failed`, or `missing` version MUST NOT remain or become the active version pointer for its `pack_key`.

At most one version of a given `pack_key` MUST be active at a time.

The implementation MUST retain the previously active version for each `pack_key` until an explicit administrative removal occurs, so operator rollback does not require incident-data changes.

Reference-pack import, verification, and refresh MUST execute as background jobs rather than as blocking grid actions.

### 11.4 Verification and attestation

Before a reference pack becomes `available` or `active`, the implementation MUST verify:

- `pack_key`,
- `pack_kind`,
- `pack_version`,
- the source identifier, if available,
- `manifest_sha256`,
- one or more payload SHA-256 digests in deterministic member order or an equivalent canonical aggregate digest,
- signature or trusted-source metadata when available,
- reference-pack contract or schema compatibility with the running application,
- safe-path validation for archive members before extraction,
- a content allowlist that rejects executable active content at import time.

Pack import or activation MUST fail closed on checksum mismatch, signature mismatch, missing required integrity metadata, contract incompatibility, incomplete download or copy, path-traversal attempt, or disallowed content.

If verification fails, the candidate pack version MUST remain inactive, and the previously active version, if any, MUST remain active.

The implementation MUST record structured, incident-external attestation metadata for pack import and pack activation. At minimum, the attestation metadata MUST persist:

- `pack_key`,
- `pack_kind`,
- `pack_version`,
- `manifest_sha256`,
- `payload_sha256`,
- `source_identifier`,
- `verification_method`,
- `signer_key_id` or trusted-source identifier,
- `imported_by_user_id`,
- `imported_at`,
- `activated_by_user_id`,
- `activated_at`,
- `previous_active_version`,
- `verification_result`,
- optional operator note or change ticket.

Attestation metadata MUST remain queryable from structured metadata without unpacking bundle contents or consulting incident data.

### 11.4.1 Activation safety and observability

Activation MUST read only from a `verified_available` candidate and MUST emit a structured activation attestation bound to the target `pack_key` and `pack_version`.

If an active version is disabled, fails a later integrity or compatibility check, or becomes missing, the implementation MUST remove or replace the active pointer in `reference_pack_activation_state` before any pack-dependent operation can continue to treat that version as active.

Pack lifecycle state MUST be observable from structured metadata alone. At minimum, an operator and a CI conformance test MUST be able to determine the current pack-version condition, the active-version pointer, the previous active version, the last verification result, and the last import or activation attestation without unpacking bundle contents.

### 11.5 Degradation behavior

If optional reference packs are absent, disabled, failed, or missing, Cartulary MUST continue to support:

- timeline capture,
- entity resolution,
- evidence attachment,
- core editing.

Only the affected overlay views, labels, or enrichment semantics MAY degrade.

## 12. Portability, backup, restore, and failure handling

### 12.1 Backup

A deployable implementation SHOULD support:

- Postgres base backup plus WAL archiving,
- object-store bucket snapshotting or versioning.

Equivalent backup mechanisms are permitted if they preserve restore semantics.

### 12.2 Restore

Restore MUST occur in this order:

1. restore Postgres,
2. restore object-store contents,
3. rebuild projections.

Projection rebuild MUST be part of restore readiness when projection contents are not restored directly.

### 12.3 Incident portability

Whole-incident export/import beyond operational backup and restore belongs to the **Incident Portability Extension Profile**.

If the implementation claims that profile, it MUST support full-fidelity administrative round-trip transfer of authoritative incident source state between trusted Cartulary deployments without depending on workbook-label semantics, live remote fetches, or deployment-local authentication configuration.

Import into an existing incident, incident cloning with identifier remapping, and partial bundle merge semantics are out of scope for this profile. A conformant import MUST preserve the exported `incident_id`, `record_id`, `row_version`, change-set identifiers, and blob hashes.

#### 12.3.1 Logical bundle contract

The canonical portability artifact MUST be a logical bundle layout. A `.zip` or `.tar` wrapper MAY be used for transport, but the normative contract is the root directory structure and file contents after extraction.

At minimum, the logical bundle root MUST contain:

```text
/
  manifest.json
  data/
    incident.json
    actors.ndjson
    records.ndjson
    timeline_events.ndjson
    hosts.ndjson
    identities.ndjson
    entity_aliases.ndjson
    indicators.ndjson
    indicator_observations.ndjson
    indicator_state_intervals.ndjson
    artifacts.ndjson
    task_requests.ndjson
    decisions.ndjson
    evidence_records.ndjson
    evidence_custody_events.ndjson
    object_blobs.ndjson
    entity_mentions.ndjson
    compromise_assessments.ndjson
    record_links.ndjson
    tags.ndjson
    record_tags.ndjson
    change_sets.ndjson
    change_set_mutations.ndjson
    record_revisions.ndjson
    saved_views.ndjson
    reference_pack_refs.json
  blobs/sha256/<sha256-lower-hex>
  integrity/checksums.sha256
  integrity/signature.ed25519        # optional
  ext/snapshots/**                   # optional
  ext/reference_packs/**             # optional
```

Bundle member paths MUST use relative forward-slash separators. The logical bundle and any outer archive wrapper MUST reject absolute paths, `.` or `..` segments, symlinks, hard links, device nodes, and other member types outside regular files and directories.

Each blob stored under `blobs/sha256/<sha256-lower-hex>` MUST contain the exact raw bytes whose SHA-256 digest matches the lowercase hex path suffix.

#### 12.3.2 Authoritative-state boundary

Portability export MUST serialize authoritative incident source state and blob bytes. It MUST NOT serialize derived or deployment-local runtime state.

A portability bundle MUST NOT include:

- projection tables or search indexes,
- live presence state,
- client-local draft queues or same-field-conflict queues,
- sessions, presigned URLs, locks, temporary caches, or other ephemeral runtime files,
- password hashes, MFA secrets, external provider configuration, or object-store credentials,
- incident memberships, current permissions, or other deployment-local authorization state.

Reference-pack attestation metadata remains incident-external state. When the Incident Portability Extension Profile embeds reference-pack payloads, the bundle MAY include only the optional embedded-pack payloads and their bundle-local descriptors, not deployment-global activation or attestation history.

#### 12.3.3 Manifest and integrity contract

`manifest.json` MUST be the canonical bundle manifest and MUST include, at minimum:

```json
{
  "bundle_format": "cartulary.incident_bundle",
  "bundle_version": 1,
  "bundle_id": "uuid",
  "incident_id": "uuid",
  "incident_key": "string",
  "exported_at": "RFC3339 timestamp",
  "source_change_set_high_watermark": "uuid-or-sequence",
  "history_mode": "full",
  "blob_mode": "full",
  "reference_pack_mode": "refs_only | embedded",
  "optional_sections": ["snapshots", "reference_packs"],
  "required_capabilities": [],
  "signing_key_id": "optional key identifier",
  "files": [
    {"path": "data/incident.json", "sha256": "sha256:...", "bytes": 123, "required": true}
  ]
}
```

`manifest.json` MUST describe one immutable export boundary. `source_change_set_high_watermark` MUST identify the frozen source boundary used to build the bundle. `history_mode` and `blob_mode` MUST each equal `full` for this profile.

`manifest.json.files[]` MUST enumerate every regular file in the logical bundle except `integrity/checksums.sha256` and `integrity/signature.ed25519`, sorted lexicographically by `path`. `required=true` MUST identify the files required to reconstruct the core incident state. `required=false` MAY be used only for optional embedded sections.

`integrity/checksums.sha256` MUST list one lowercase SHA-256 and relative path per line for every file listed in `manifest.json.files[]`, sorted lexicographically by path, using the exact file bytes carried in the bundle.

If `integrity/signature.ed25519` is present, `manifest.json.signing_key_id` MUST also be present. The signature MUST cover the exact bytes of `integrity/checksums.sha256`. If a deployment supports signature verification for portability bundles, signature failure MUST reject the bundle before any structured data becomes visible.

If `required_capabilities[]` names a capability the target deployment does not implement, import MUST fail closed. Optional embedded sections MAY be ignored when unsupported unless the corresponding capability is listed in `required_capabilities[]`.

#### 12.3.4 Structured formats and deterministic serialization

The portability bundle MUST use JSON for singleton files and NDJSON for multi-row files. CSV, XLSX, or other workbook-shaped exports MAY exist elsewhere. They MUST NOT be the authoritative whole-incident portability format.

All structured files in the bundle MUST use:

- UTF-8 encoding,
- LF line endings,
- no BOM,
- lexicographically sorted object keys,
- exactly one JSON object per NDJSON line,
- deterministic file-level row ordering.

NDJSON files whose rows have a stable single-column primary key MUST sort ascending by that key. `record_tags.ndjson` MUST sort by `(record_id, tag_id)`. `change_set_mutations.ndjson` MUST sort by `(change_set_id, sequence_no)`. Files with integer primary keys such as `record_revisions.ndjson` MUST sort ascending by that integer key.

The canonical JSON serialization for singleton JSON files and per-line NDJSON objects MUST be stable enough that exporting the same incident state twice without intervening mutations produces byte-identical structured files and identical `integrity/checksums.sha256`.

#### 12.3.5 Portable actors and optional embedded sections

`actors.ndjson` MUST preserve historical attribution without exporting portable login material. Each actor descriptor row MUST carry, at minimum:

- stable `actor_id`,
- display name,
- non-secret match hints when available, such as normalized email or provider-subject hint.

Import MUST materialize every actor referenced by imported history as either:

- an inert imported actor that is not login-capable and is not automatically added to incident membership, or
- an equivalent historical actor descriptor bound to an existing local user without rewriting the source bundle actor identifier used by imported history.

`reference_pack_refs.json` MUST list the reference-pack keys and versions referenced by imported saved views, overlays, or optional embedded sections. Missing referenced packs MUST degrade only the affected overlays or saved views. They MUST NOT block import of the core incident state.

If `reference_pack_mode='embedded'`, pack payloads MAY be embedded under `ext/reference_packs/**`. If `optional_sections` contains `snapshots`, immutable snapshot descriptors and rendered artifacts MAY be embedded under `ext/snapshots/**`. Unsupported or missing optional embedded sections MUST NOT block import of the core incident state unless the relevant capability is named in `required_capabilities[]`.

#### 12.3.6 Export and import execution semantics

Whole-incident export and import MUST execute as background jobs rather than as blocking grid actions.

A conformant import MUST execute the following phases in order:

1. stage the supplied outer archive or logical bundle under the configured temporary-work root,
2. validate the outer container and every member path,
3. verify every required checksum and any supported signature before any structured data becomes visible,
4. stage blob bytes and verify every required blob hash,
5. import the structured incident state,
6. rebuild projections,
7. mark the imported incident visible only after the structured import and projection rebuild succeed.

Import MUST fail closed on any of the following:

- checksum mismatch,
- signature mismatch when signature verification is supported or required by deployment policy,
- missing required file,
- missing required blob or blob-hash mismatch,
- invalid path or unsupported bundle member type,
- unsupported `required_capabilities[]` entry,
- duplicate `incident_id`,
- any import path that would require a live remote fetch to complete.

If import fails after staging begins, the target deployment MUST leave no partially visible incident. Staged bytes MAY be retained only in a non-visible administrative quarantine or temporary-work area.

### 12.4 Failure handling

The implementation MUST satisfy all of the following failure semantics:

- if the application container is unavailable, sessions MAY drop but committed data MUST remain durable,
- if Postgres is unavailable, the system MAY become unavailable,
- if object storage is unavailable, row editing MUST remain possible but evidence upload and download MUST fail clearly,
- if projections are unavailable or corrupt, the implementation MUST preserve source data integrity and rebuild projections from source state.

## 13. Long-running operations and background jobs

The implementation MUST execute the following long-running operations as background jobs rather than as blocking grid actions:

- lookups beyond trivial inline suggestion queries,
- imports,
- incident portability export and import when the Incident Portability Extension Profile is implemented,
- reference-pack import, verification, and refresh,
- snapshot generation,
- report builds,
- projection rebuilds,
- evidence processing, including blob hashing, scanning, preview generation, thumbnailing, and metadata extraction.

Background jobs MUST expose:

- progress,
- cancellation,
- retry-safe status,
- non-blocking UI behavior.

Grid editing and row creation MUST remain responsive while these jobs run. These jobs MUST NOT block row selection, sheet filtering, sorting, grouping, or inspector metadata open.

## 14. Runtime roots and packaging

The deployment configuration MUST name explicit persistent roots for:

- database storage,
- object storage,
- reference-pack storage,
- temporary work files,
- export outputs.

The implementation MUST NOT rely on source-tree-relative defaults for icons, Markdown templates, reference data, or generated artifacts.

## 15. Architecture invariants

An implementation conforming to this core MUST preserve all of the following:

1. the complexity budget belongs in mutation semantics, projections, and workbook UX rather than distributed infrastructure,
2. derived surfaces MUST share a canonical derivation layer,
3. optional enrichment MUST remain off the hot capture path,
4. the object store boundary MUST remain explicit and lifecycle-aware,
5. view behavior MUST remain contract-driven,
6. projection tables MUST remain disposable derived state,
7. file-based import complexity MUST remain isolated behind the imports module and stable tabular-ingest contract rather than leaking spreadsheet-specific parser behavior into workbook-domain modules,
8. whole-incident portability, when implemented, MUST export authoritative source state and referenced blob bytes rather than projections, snapshots, or deployment-local runtime state.
