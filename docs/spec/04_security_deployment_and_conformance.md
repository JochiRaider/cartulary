# Cartulary Normative Core 04: Security, Deployment, and Conformance

## 1. Authentication model

### 1.1 Base authentication

The base profile MUST support:

- local user accounts stored in Postgres,
- password hashing with Argon2id,
- TOTP MFA,
- optional WebAuthn when the environment supports it.

Authentication MUST work in disconnected deployments and MUST NOT depend on enterprise infrastructure for the base profile.

### 1.2 Enterprise Authentication Extension Profile

If the implementation claims the **Enterprise Authentication Extension Profile**, it MUST support provider-backed identities through an `auth_providers` and `auth_identities` model equivalent to the source artifact.

OIDC is the preferred enterprise path. SAML is the secondary enterprise path when required by the environment.

External identities MUST map to the same internal user identity used for attribution so that audit semantics remain unchanged.

## 2. Authorization model

The base profile MUST use incident-level roles equivalent to:

- `viewer`,
- `editor`,
- `reviewer`,
- `admin`.

Record access MUST inherit from incident access in the base profile.

Field-level ACLs, approval workflows, and generalized record-level ACL systems are out of scope for the base profile.

A narrow sensitive-evidence model MAY be added in future work. It is not a current conformance requirement.

## 3. Attribution and audit requirements

Every mutation MUST originate from an authenticated session or an explicitly identified system process.

Every mutation MUST record, at minimum:

- actor user identifier,
- timestamp,
- mutation source such as `ui`, `import`, or `rollback`,
- before/after values or equivalent patch data at the required history granularity.

Security choices MUST NOT add friction to the primary capture path. MFA belongs at login and session establishment, not during routine row creation or evidence preview.

## 4. Trust boundaries

### 4.1 Reference packs

Optional enrichment credentials MUST live in server-side configuration or secret storage. They MUST NOT live in incident records, client-side storage, or imported pack files.

Reference packs MUST record version, source, and integrity metadata before activation.

### 4.2 Export outputs

Generated reports and presentations MUST embed or package required assets locally rather than pulling them from remote CDNs or runtime asset services.

### 4.3 Evidence uploads

The deployment MAY use an upload-malware-scanning sidecar or equivalent adjunct service. Such a service is optional in the current core and MUST NOT break the two-step attachment semantics.

## 5. Deployment profiles

### 5.1 Flyaway or disconnected deployment

The recommended disconnected deployment MUST consist of:

- one application container,
- one Postgres container,
- one MinIO container or equivalent S3-compatible object store.

Docker Compose or Podman Compose with mounted volumes is acceptable.

### 5.2 On-prem deployment

On-prem deployments MAY replace the Postgres or object-store containers with centrally managed services if equivalent semantics are preserved.

The application deployable MUST remain a single application unit.

### 5.3 Cloud deployment

Cloud deployments MAY run the application on ECS, Kubernetes, VMs, or equivalent platforms, with managed Postgres and native object storage.

The logical architecture and data contracts MUST remain unchanged.

## 6. Runtime roots and storage paths

The deployment configuration MUST declare explicit persistent roots for:

- database storage,
- object storage,
- reference-pack storage,
- temporary work files,
- export outputs.

The application MUST NOT rely on source-tree-relative paths for runtime assets or generated artifacts.

## 7. Container boundary

The following components MUST remain in one deployable application unit:

- web UI,
- API,
- WebSocket hub,
- background jobs.

Postgres and object storage SHOULD remain separate services.

## 8. Required and optional supporting services

### 8.1 Required services

A conformant deployment MUST provide:

- Postgres,
- object storage,
- the application deployable.

### 8.2 Optional services

A conformant deployment MAY additionally provide:

- external reverse proxy or TLS termination,
- enterprise IdP,
- evidence scanning sidecar,
- managed storage substitutes preserving the same contracts.

## 9. Conformance criteria

Each criterion below is a pass/fail requirement.

### 9.1 Base Profile criteria

- **AC-001**: An analyst can create a new timeline row by typing into a blank grid cell and pressing Enter, with no modal and no required form, and the row is saved in under 150 ms on LAN for a single-row edit.
- **AC-002**: A timeline row can be persisted with only one non-empty user-entered value or only an attached screenshot; `recorded_at`, author, and row identity are system-generated.
- **AC-003**: Pasting a 20-row by 5-column block from Excel into the timeline sheet creates rows and maps visible columns in under 2 seconds on a reference incident.
- **AC-004**: Pasting an image from the clipboard or dragging a screenshot onto a selected row attaches evidence in no more than two user actions and does not navigate away from the grid.
- **AC-005**: Arrow keys, Tab, Enter, Shift+Enter, and Ctrl+V work in the grid without opening side dialogs or breaking selection state.
- **AC-006**: An analyst can resolve an unresolved host or account mention from the inspector and return focus to the original grid cell without losing scroll position or selection.
- **AC-007**: The selected row’s edit history is visible in one click or one shortcut and includes actor, timestamp, operation, changed field, link, mention, tag, or evidence entry, plus rollback actions.
- **AC-008**: Two analysts on the same sheet can see each other’s presence within 1 second, including row-level presence and same-cell editing indicators when applicable.
- **AC-009**: Concurrent edits to different fields on the same row auto-merge; concurrent edits to the same field never silently overwrite without a visible conflict.
- **AC-010**: A reviewer can roll back one mistaken host link, tag assignment, mention resolution, or evidence association from the selected row’s history without reverting later unrelated edits on the same row.
- **AC-011**: Rolling back a mistaken change or restoring a whole row creates a new attributed revision and updates the visible row in under 2 seconds on a reference incident.
- **AC-012**: Whole-row restore and whole-change-set rollback are available as explicit secondary actions for multi-target or destructive changes; arbitrary field-picker rollback from historical snapshots is not required in the base profile.
- **AC-013**: Re-sorting, re-filtering, or re-grouping a sheet does not cause a pending edit to target a different underlying record; all mutations are sent using `record_id`, `base_row_version`, and changed fields only.
- **AC-014**: Renaming a visible column header or tab label does not change filter semantics, write-back behavior, or export semantics for a built-in or system view; those behaviors are bound to `view_schema_id`.
- **AC-015**: An analyst can create an evidence request record without uploading a blob, later attach or replace the blob, and preserve request, receipt, custody, and storage metadata across the lifecycle.
- **AC-016**: Evidence processing and any implemented background job start without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
- **AC-017**: Indicator values imported or entered through artifact-backed flows appear in a stable system-view contract and, when export surfaces exist, in a stable export contract with consistent indicator type, indicator value, normalization, and STIX-mapping fields.
- **AC-018**: Recording a new compromised or cleared assessment for a host or identity appends a new attributed assessment record; prior assessments remain visible in history and are not overwritten.
- **AC-019**: Typing or pasting `WS-023?` into a Timeline Hosts cell creates an `entity_mention` and zero host records unless the analyst explicitly resolves or creates an entity.
- **AC-020**: Creating a host or identity from a selected unresolved mention creates exactly one stub entity, preserves the raw mention, resolves only the selected mention by default, and stores the seed value in alias or provenance data.
- **AC-021**: Repeated identical unresolved mentions across different source rows remain separate mention rows with distinct provenance and are never coalesced into a single mention record.
- **AC-022**: Pasting into an `entity_origin` mapping upserts an existing active entity on a unique exact-match key and otherwise creates a stub; it never auto-merges two pre-existing entities.
- **AC-023**: Merging two entities preserves loser lineage, repoints live mention resolutions and live links to the survivor in one change set, and does not change the survivor `record_id`.
- **AC-024**: In the timeline sheet, the grouping control offers only `None`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions` in the base profile; it does not offer grouping by Summary, Hosts, Identities, Tags, or arbitrary custom columns.
- **AC-025**: A grouped timeline sheet exposes exactly one derived outline level. `expand group`, `collapse group`, `expand all`, `collapse all`, and `Group: None` work without creating editable rows, paste targets, subtotal rows, or `record_id`-bound mutation targets.
- **AC-026**: While grouped, sorting, filtering, paste, autosave, conflict handling, rollback, and export flatten to underlying records only. Editing a grouped field may move the row to a different visible group, but drag-and-drop reclassification and manual row-range grouping are not available.

### 9.2 Import Extension Profile criteria

- **AC-027**: Full XLSX import runs without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
- **AC-028**: Importing through an `entity_origin` mapping upserts an existing active entity on a unique exact-match key and otherwise creates a stub; it never auto-merges two pre-existing entities.
- **AC-029**: Full XLSX import preserves unknown columns in raw-capture or custom-attribute storage and leaves imported host or account tokens unresolved unless an analyst explicitly resolves them later.

### 9.3 Snapshot and Reporting Extension Profile criteria

- **AC-030**: Report generation and snapshot generation run without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
- **AC-031**: A generated HTML report or presentation artifact opens in a disconnected browser without fetching remote JavaScript, CSS, or fonts.
- **AC-032**: Snapshot-derived outputs preserve stable identifiers and ordering consistent with the canonical derivation layer.

### 9.4 Reference Pack Extension Profile criteria

- **AC-033**: Reference-pack refresh runs without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
- **AC-034**: Missing optional reference packs degrade only the affected overlay views or labels; timeline capture, entity resolution, and evidence attachment continue to function.
- **AC-035**: Pack activation fails closed on checksum, signature, or equivalent integrity mismatch.

### 9.5 Enterprise Authentication Extension Profile criteria

- **AC-036**: Enterprise-authenticated users map to the same internal user identity used for attribution, and switching from local auth to enterprise auth does not break audit lineage for existing incidents.

### 9.6 Additional Base Profile criteria for same-field conflict resolution

- **AC-037**: When two analysts edit the same write-back-capable field concurrently, the losing client shows a conflicted cell and same-surface resolver that presents saved value, unsaved value, row context, actor, and timestamp without leaving the sheet.
- **AC-038**: Choosing `Keep saved value` clears the local conflict without creating a source revision; choosing `Use my unsaved value` or `Edit merged value` creates a new attributed change set and updates the visible row.
- **AC-039**: If the field changes again while the resolver is open, stale resolution is rejected and the latest conflict payload is shown without losing the analyst's unsaved draft.
- **AC-040**: A paste containing both non-conflicting and same-field-conflicting cells commits the non-conflicting cells immediately and groups the conflicting cells into a navigable conflict queue without per-cell modal interruption.
- **AC-041**: Unresolved local conflict drafts are not broadcast to other analysts and do not appear in search, history, exports, or snapshots unless explicitly committed.
- **AC-042**: After resolving a conflict, focus returns to the same cell and scroll position is preserved.

## 10. Non-goals preserved from the source artifact

The following remain explicit non-goals for current conformance:

- generalized field-level ACLs,
- approval workflows,
- full spreadsheet formula engines,
- merged cells,
- character-by-character Google-Sheets-style cell CRDT behavior,
- manual row-range grouping,
- arbitrary field-picker rollback,
- automatic entity merge based on fuzzy similarity alone.

## 11. Operational posture

A conformant Cartulary deployment MUST prefer operational simplicity over distributed architectural purity.

The smallest useful deployment is intentionally not the absolute minimum number of containers. It is the minimum that keeps binary evidence handling sane and portable while preserving collaboration, authentication, and auditable source-of-truth behavior.
