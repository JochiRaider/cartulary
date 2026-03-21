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
- links and tags,
- revisions and rollback,
- projections and search,
- reference data,
- reporting and snapshot generation,
- collaboration and presence.

These are logical module boundaries. They MUST be independently testable, but they MUST NOT require separate deployables.

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
- snapshot metadata and canonical export-model metadata when the Snapshot and Reporting Extension Profile is implemented.

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

The implementation MUST:

- capture a `snapshot_at` boundary,
- materialize a canonical export model such as `incident_report_model.json`,
- render derivative outputs from that immutable snapshot rather than from mutable live tables.

### 10.3 Output forms

A reporting-capable implementation MAY generate:

- Markdown reports,
- Mermaid diagram sources,
- Slidev decks,
- HTML reports,
- operator-facing reenactment outputs such as Asciinema-style walkthroughs.

If such outputs are generated, the implementation MUST preserve the distinction between source evidence and generated presentation material.

### 10.4 Self-contained outputs

Generated report and presentation artifacts MUST be self-contained. They MUST NOT require remote JavaScript, CSS, or fonts at render time.

### 10.5 Redaction boundary

The snapshot and export subsystem MUST provide a stable boundary at which redaction rules can later be applied. The exact redaction controls remain an open design area preserved in Appendix E.

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
6. projection tables MUST remain disposable derived state.
