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
- entities and mention resolution,
- evidence and object storage,
- imports and tabular ingest,
- links and tags,
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

Reference-pack manifests and integrity metadata MUST be stored in Postgres. Pack payloads MAY live on local disk or object storage behind the same abstraction, but their activation state MUST remain queryable from structured metadata.

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

### 7.2 Contract-backed system views

The base profile MUST support contract-backed system views for:

- indicators,
- compromise assessments.

Framework overlays such as ATT&CK, D3FEND, or VERIS MAY also be exposed as system views when the relevant reference packs are present.

System views MUST follow the same `view_schema_id` contract discipline as built-in sheets.

### 7.3 Notes status

The current normative core treats **Notes** as a built-in sheet because the source artifact made it part of the current workbook and MVP surface. Appendix E preserves the open product question about whether this remains a built-in surface at GA. That future question does not change current conformance.

## 8. Projection model

### 8.1 Projection tables

Hot workbook screens MUST use projection tables rather than Postgres materialized views.

The implementation MUST define projection tables equivalent to:

- `timeline_grid_projection`,
- `host_grid_projection`,
- `identity_grid_projection`,
- `artifact_grid_projection`,
- `evidence_grid_projection`,
- `indicator_grid_projection` for the indicator system view.

Each projection row MUST represent exactly one primary record in the base projection for that view.

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

The grid and inspector hot path MUST synchronously read only scalar fields, flags, counts, and small preview handles needed for the visible viewport or selected row. They MUST NOT synchronously fetch full attachment lists or binary blob bytes as part of grid rendering, row selection, sheet filtering, grouping, or inspector metadata open.

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

The closed vocabulary for redaction actions is:

- `allow`,
- `drop`,
- `mask`,
- `truncate`,
- `hash`,
- `stub`.

Redaction MUST run after snapshot materialization and before template rendering.

If a field or block eligible for the chosen `release_scope` appears in the canonical export model without an applicable redaction rule, rendering MUST fail closed.

Each rendered artifact MUST emit a `redaction_manifest` keyed by stable export-model path and rule identifier, recording every field or block that was dropped, masked, truncated, hashed, or stubbed.

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

### 11.2 Activation and integrity

Before a reference pack becomes active, the implementation MUST verify:

- the pack version,
- the source identifier, if available,
- the checksum,
- signature or trusted-source metadata when available.

Pack activation or refresh MUST fail closed on checksum mismatch, signature mismatch, incomplete download, or equivalent integrity failure.

### 11.3 Degradation behavior

If optional reference packs are absent, disabled, or failed, Cartulary MUST continue to support:

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

The implementation SHOULD support whole-incident export/import as a manifest plus structured records archive plus referenced blobs archive. The import/export bundle details remain future specification work, but portability as a design goal is preserved.

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
- reference-pack refresh,
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
7. file-based import complexity MUST remain isolated behind the imports module and stable tabular-ingest contract rather than leaking spreadsheet-specific parser behavior into workbook-domain modules.
