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

`task_request`, `decision`, and coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MUST inherit the same incident-level authorization model. The base profile MUST NOT introduce record-specific ACLs or hidden sub-workspaces for these objects.

Field-level ACLs, generalized approval workflows, and generalized record-level ACL systems are out of scope for the base profile.

If the Snapshot and Reporting Extension Profile is implemented, export redaction MUST NOT restrict live workbook views, search results, filters, saved views, row visibility, field visibility, or evidence visibility for authenticated incident participants. In the base profile, live workspace visibility is derived only from incident membership and the incident-level role model. In that profile, recipient-specific withholding MUST be implemented at snapshot, render, and release time rather than by hiding live workspace content.

### 2.1 Snapshot and Reporting Extension Profile release gate

If the implementation claims the Snapshot and Reporting Extension Profile, it MUST provide a narrow artifact-scoped release gate for rendered outputs. This release gate MUST NOT become a generalized workflow engine for routine record editing or arbitrary record approvals.

Each release approval record MUST bind, at minimum, to:

- `snapshot_id`,
- `template_id`,
- `template_version`,
- `redaction_profile_id`,
- `redaction_profile_version`,
- `output_kind`,
- `release_scope`,
- `output_sha256`.

Approval requirements are:

- `internal_draft`: no approval required,
- `internal_review`: one `reviewer` approval,
- `external_release`: two distinct approvals, one from a `reviewer` attesting evidence sufficiency or claim support and one from an `admin` attesting release posture and redaction completeness.

Any change to the bound tuple or rendered bytes MUST invalidate prior approvals automatically.

A narrow live sensitive-evidence model MAY be added in future work if repeated real-world incidents show that export-scoped withholding is insufficient. It is not a current conformance requirement.

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

Reference packs MUST record structured metadata before activation. Queryable metadata MUST include, at minimum, `pack_key`, `pack_kind`, `pack_version`, source identifier if available, `manifest_sha256`, one or more payload SHA-256 digests in deterministic member order or an equivalent canonical aggregate digest, `verification_method`, signer-key or trusted-source identifier, imported and activated actor attribution with timestamps, `previous_active_version`, and `verification_result`.

In a flyaway or disconnected deployment, reference-pack import and activation MUST operate only on locally supplied bundles rooted under the configured reference-pack storage path or an equivalent administrative upload path that writes into that root. The running application MUST NOT require a live network fetch to verify or activate a pack.

Imported pack bundles and extracted contents MUST be treated as hostile content until verification and content screening succeed.

### 4.2 Export outputs

Generated reports and presentations MUST embed or package required assets locally rather than pulling them from remote CDNs or runtime asset services.

If the Snapshot and Reporting Extension Profile is implemented:

- `external_release` outputs MUST exclude raw blob bytes and `working_material`,
- any `curated_narrative` block included in `external_release` MUST carry `support_refs[]`,
- content drawn directly from `task_request`, `decision`, `comm_log`, `handoff`, `status_review`, or `lesson` records MUST NOT be treated as inherently releasable; any `external_release` use MUST flow through the snapshot, redaction, and curation path,
- reenactment outputs MUST be marked `generated_presentation=true` and MUST NOT be released as `external_release`,
- approval and redaction checks MUST complete successfully before an `external_release` artifact is published.

In that profile, the implementation MUST support generating multiple recipient-specific artifacts from the same immutable snapshot by selecting different versioned redaction profiles and, when needed, different templates. If an incident involves multiple affected parties, an artifact prepared with one recipient-specific configuration MUST NOT disclose content whose `disclosure_partition_refs[]` are not allowed by the selected redaction profile. Manual post-render editing MAY still occur, but it MUST NOT be required for the implementation's supported recipient-specific configurations.

### 4.3 Evidence uploads

The deployment MAY use an upload-malware-scanning sidecar or equivalent adjunct service. Such a service is optional in the current core and MUST NOT break the two-step attachment semantics.

### 4.4 STRIDE threat model

The implementation MUST maintain a project-local STRIDE threat model covering the current architecture, deployment profiles, and high-risk workflows.

The threat model MUST be updated before any release that adds or materially changes:

- an import path,
- an export or report surface,
- an evidence preview or rendering path,
- an external fetch capability,
- a credential-bearing integration,
- a deployment profile,
- an object-storage access pattern.

At minimum, the threat model MUST cover the following assets and abuse cases:

| STRIDE class | Minimum project-specific scope | Required control direction |
| --- | --- | --- |
| Spoofing | analyst sessions, provider-backed identities, explicit system-process actors, object-store upload/download capabilities | authenticated sessions, stable internal user mapping, explicit system actors, short-lived operation-scoped object access |
| Tampering | incident records, revisions, object blobs, reference packs, snapshots, exports | row-versioned writes, immutable change sets, blob hashes, fail-closed integrity verification, immutable published snapshots |
| Repudiation | edits, imports, rollbacks, pack activation, export generation, evidence lifecycle actions | attributed append-only history with actor, timestamp, source, and reversible mutation detail |
| Information disclosure | evidence blobs, exports, previews, secrets, portable runtime roots | incident-scoped authorization, secret isolation, untrusted-content rendering rules, self-contained outputs, encrypted flyaway storage |
| Denial of service | oversized evidence, archive bombs, pathological imports, expensive report or preview jobs | size and decompression limits, background-job isolation, cancellation, bounded hot-path retrieval |
| Elevation of privilege | user-controlled record or blob identifiers, destructive operations, job-worker storage access | server-side authorization derived from object ownership, role gates for destructive actions, least-privilege worker credentials |

### 4.5 Focused MITRE CWE constraints

The implementation MUST address the following MITRE CWE entries during architecture review, code review, and conformance testing. This list is intentionally narrow and project-specific.

- **CWE-79**: Incident-authored or imported content rendered in the browser UI or exported HTML MUST be treated as untrusted. Renderers MUST escape or sanitize by default and MUST block script execution, inline event handlers, `javascript:` URLs, and remote asset fetches sourced from incident data.
- **CWE-1236**: CSV, XLSX, and clipboard exports intended for spreadsheet consumption MUST neutralize leading formula characters before write. At minimum, values beginning with `=`, `+`, `-`, `@`, tab, or carriage return MUST be emitted with a lossless neutralizing prefix such as `'`, unless an explicit raw-forensic export mode is selected with a visible danger warning.
- **CWE-22 / CWE-73**: User-supplied filenames, archive entry names, and import paths MUST be treated as metadata, not authority. The system MUST assign storage keys, MUST reject absolute paths and parent traversal, and MUST extract archives only inside a staging root that cannot escape the declared runtime roots.
- **CWE-353**: Reference packs and any incident import bundle format, when implemented, MUST fail closed on checksum mismatch, signature mismatch, incomplete download, or missing required integrity metadata.
- **CWE-434**: Evidence, reference-pack, and workbook-import uploads MUST be treated as hostile content. The application unit MUST NOT execute uploaded content, workbook formulas, macros or VBA, workbook automation, or external links during import or preview, and preview generation MUST use only allowlisted non-executing transforms. Active-content types MUST remain download-only or isolated from the main application origin unless a dedicated isolated analysis path is explicitly implemented.
- **CWE-639**: Every mutation, preview, download, and object-store URL issuance MUST re-derive authorization server-side from the target object's owning incident and the caller's current membership and role. Client-supplied incident identifiers, ownership metadata, or role claims MUST NOT determine access.
- **CWE-312**: Deployments intended for portable or flyaway use MUST keep database storage, object storage, reference-pack storage, temporary work files that carry incident data, and export outputs on encrypted storage. Unencrypted removable media or unencrypted portable roots are non-conformant for flyaway handling.

## 5. Deployment profiles

### 5.1 Flyaway or disconnected deployment

The recommended disconnected deployment MUST consist of:

- one application container,
- one Postgres container,
- one MinIO container or equivalent S3-compatible object store.

Docker Compose or Podman Compose with mounted volumes is acceptable.

If the deployment claims the Reference Pack Extension Profile, the smallest supported disconnected bundle MUST preinstall only the three reference packs defined in Core 01 §11.2. Framework, enrichment, template, and separately distributed view-contract packs MUST remain separately installable offline bundles in that minimum deployment.

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

For performance-sensitive criteria in this section, the following reference performance fixtures apply:

- **Fixture A: large-grid incident**
  - 20,000 timeline rows
  - 1,000 host rows
  - 1,000 identity rows
  - 25 concurrently connected analyst sessions on one incident, with presence enabled and representative live row-update traffic
  - representative tags, mentions, and links, but not evidence-heavy per row
- **Fixture B: evidence-heavy incident**
  - 5,000 timeline rows
  - 10,000 evidence records
  - tens of GB of binary evidence stored in object storage
  - at least one timeline row linked to 100 evidence records
  - evidence blobs MAY be stubbed for throughput tests, but evidence metadata, counts, attachment state, and preview handles MUST be real

Unless a criterion states otherwise, `reference incident` in this section means Fixture A.

Latency measurements in this section MUST use end-user-observable completion time from the initiating user action to the required visible UI state. Any criterion expressed as p95 MUST be evaluated over at least 100 completed operations of the named interaction after one warm-up pass on the named fixture.

For this section:

- `first useful viewport` means the first rendered visible row window for the active sort, filter, and grouping state with stable `record_id` binding and working keyboard navigation, even if off-screen rows continue loading.
- `stable viewport` means the visible row window and result ordering match the final deterministic order for the active sort, filter, and grouping state and no further reorder occurs without new user or server input.
- `metadata shell` means row fields and evidence metadata needed to inspect the selected record, including counts, filenames or media-type labels, attachment state, and preview handles, but excluding binary preview bytes or full blob download.

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
- **AC-067**: An analyst can create a note directly in the Notes sheet by blank-row or equivalent grid-native entry with no required modal, and the resulting record is stored as an artifact with `artifact_type='note'`.
- **AC-068**: From a selected Timeline, Host, Identity, or Evidence record, an analyst can invoke `add linked note` without leaving the workbook flow; the resulting note is stored with the same artifact record shape as a Notes-sheet-created note, appears in the Notes sheet, and is visible as a linked record from the source row.
- **AC-069**: Renaming the visible Notes tab label, and any implementation-supported per-user hide/show operation for that built-in tab, does not change write-back or export semantics because behavior remains bound to `view_schema_id`.
- **AC-070**: Editing a lightweight free-text field on a timeline, host, identity, or evidence record does not create a standalone Notes row. Creating a standalone note does create a distinct note record that can be linked, tagged, and reviewed in history independently.
- **AC-015**: An analyst can create an evidence request record without uploading a blob, later attach or replace the blob, and preserve request, receipt, custody, and storage metadata across the lifecycle.
- **AC-016**: Evidence processing and any implemented background job start without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
- **AC-043**: On Fixture A, selection change, focus change, and typing acknowledgment each remain at or below 100 ms p95 on LAN, and blank-row creation in the timeline sheet with one non-empty user-entered value remains at or below 150 ms p95 on LAN.
- **AC-044**: On Fixture A, sort, filter, and grouping changes present a first useful viewport at or below 250 ms p95 and a stable viewport at or below 1.0 s p95.
- **AC-045**: On Fixture B, opening the inspector for a timeline row linked to 100 evidence records shows a metadata shell at or below 300 ms p95; binary preview bytes MAY continue progressively after the inspector opens.
- **AC-046**: On either Fixture A or Fixture B, imports, evidence processing, projection rebuilds, snapshot generation, report generation, and reference-pack refresh remain background jobs, show progress and cancellation within 1 second of job start, and do not block grid editing or row creation.
- **AC-047**: On Fixture A, scrolling, sorting, filtering, grouping, and live updates never retarget pending edits away from the selected `record_id`, and viewport stabilization does not require a full-sheet rerender.
- **AC-017**: Indicator data entered through the Indicators system view, imported through supported ingest paths, or resolved from supported source fields appears under a stable indicator system-view contract and, when export surfaces exist, a stable export contract with consistent indicator type, value kind, canonical value, normalization, deterministic dedupe key, and optional STIX-mapping fields.
- **AC-018**: Recording a new `unknown`, `suspected`, `confirmed`, `disproven`, or `cleared` assessment for a host or identity appends a new attributed assessment record; prior assessments remain visible in history and are not overwritten.
- **AC-080**: A host or identity can receive `unknown -> suspected -> confirmed -> cleared` as four separate assessment records for the same incident-scoped subject, and the reviewer can still see each earlier record in order.
- **AC-081**: A host or identity can receive `unknown -> disproven` without implying prior compromise, and filtering for `assessment_state='cleared'` excludes `disproven`.
- **AC-082**: Recording an operational response action such as device isolation, account disablement, credential reset, or monitoring does not mutate `assessment_state` unless a new explicit assessment record is appended.
- **AC-083**: Interactive compromise-assessment entry surfaces expose confidence by default as `unset`, `low`, `medium`, or `high`; choosing those values persists `confidence_score` as `NULL`, `25`, `55`, or `85` respectively, and any supported exact-score write path accepts integers from `0` through `100`.
- **AC-084**: The Compromise Assessments system view supports separate filters on `assessment_state` and derived `confidence_band`; `confidence_score=NULL` is rendered and filterable as `confidence_band='unset'` and is not treated as `0`.
- **AC-085**: From a selected Notes, Timeline, Host, Identity, or Evidence context, an analyst can create a first-class `task_request` with required `task_kind`, `status`, `owner_user_id`, `priority`, `created_at`, and `updated_at`; an active task cannot be saved without an owner, and workbook filtering can show open tasks by owner, blocked state, and due status.
- **AC-086**: Creating a `decision` record preserves `decision_type`, `status`, `owner_user_id`, `decided_at`, and `rationale`; reviewer history can reconstruct when the decision changed state, who owned it, and what `support_refs[]` were attached.
- **AC-087**: A `comm_log` or `status_review` artifact can record the required timestamp, summary, and linked decision or task references, and a workbook surface can filter or sort those artifacts by timestamp and next-report or next-checkpoint time without rereading unrelated note text.
- **AC-088**: A `handoff` artifact can capture outgoing owner, incoming owner, current state summary, open task IDs, and open decision IDs, and the incoming analyst can pivot directly from the handoff to the referenced open work from the same incident.
- **AC-089**: Hypothesis tracking remains artifact-backed in the current profile, using either `artifact_type='hypothesis'` or another declared structured artifact subtype; base conformance does not require or claim a first-class `hypothesis` `record_type`.
- **AC-090**: Creating or editing timeline rows does not require owner, approver, challenge, checklist, task, or decision fields on the timeline sheet, and ordinary row edits do not enter a generalized approval workflow.
- **AC-019**: Typing or pasting `WS-023?` into a Timeline Hosts cell creates an `entity_mention` and zero host records unless the analyst explicitly resolves or creates an entity.
- **AC-020**: Creating a host or identity from a selected unresolved mention creates exactly one stub entity, preserves the raw mention, resolves only the selected mention by default, and stores the seed value in alias or provenance data.
- **AC-021**: Repeated identical unresolved mentions across different source rows remain separate mention rows with distinct provenance and are never coalesced into a single mention record.
- **AC-072**: A supported same-surface indicator-linking or extraction action on raw timeline, note, artifact, or evidence text preserves the raw field value and creates one or more source-bound `indicator_observation` rows without requiring dedicated IOC columns or rewriting the source text.
- **AC-073**: Repeated identical indicator values observed in different source rows remain separate `indicator_observation` rows with distinct provenance and are never coalesced into a single observation record.
- **AC-074**: Multiple source-bound `indicator_observation` rows can resolve to one canonical indicator record identified by incident scope plus deterministic indicator type and dedupe key, or an equivalent stable canonical identity.
- **AC-075**: A canonical indicator can carry more than one attributed lifecycle interval within the same incident, and observation-derived `first_observed_at` or `last_observed_at` remain distinct from lifecycle `valid_from` or `valid_to`.
- **AC-076**: Time-filtered indicator pivots distinguish observation time from asserted active-compromise or believed-validity time; filtering by one MUST NOT silently substitute the other.
- **AC-077**: Retiring, clearing, or superseding an indicator does not rewrite preserved source text or delete prior `indicator_observation` rows; lifecycle changes append new structured history.
- **AC-078**: The Indicators system view shows one row per canonical indicator record, not one row per source artifact or source observation, and remains stable across import, export, reporting, and future storage evolution.
- **AC-079**: A hostname, MAC address, or similar value can be linked simultaneously to a host or identity record and to a canonical indicator record without forcing a shared object identity or deleting either linkage.
- **AC-022**: Pasting into an `entity_origin` mapping upserts an existing active entity on a unique exact-match key and otherwise creates a stub; it never auto-merges two pre-existing entities.
- **AC-023**: Merging two entities preserves loser lineage, repoints live mention resolutions and live links to the survivor in one change set, and does not change the survivor `record_id`.
- **AC-024**: In the timeline sheet, the grouping control offers only `None`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions` in the base profile; it does not offer grouping by Summary, Hosts, Identities, Tags, or arbitrary custom columns.
- **AC-025**: A grouped timeline sheet exposes exactly one derived outline level. `expand group`, `collapse group`, `expand all`, `collapse all`, and `Group: None` work without creating editable rows, paste targets, subtotal rows, or `record_id`-bound mutation targets.
- **AC-026**: While grouped, sorting, filtering, paste, autosave, conflict handling, rollback, and export flatten to underlying records only. Editing a grouped field may move the row to a different visible group, but drag-and-drop reclassification and manual row-range grouping are not available.

### 9.2 Import Extension Profile criteria

- **AC-027**: CSV file import and selected-sheet or selected-region XLSX import run without blocking grid editing, the UI shows progress and cancellation within 1 second of job start, and the operator can review a header-mapping preview before commit.
- **AC-028**: File-based import uses the same mapping engine and `entity_binding_mode` semantics as clipboard-driven structured ingest. Importing through an `entity_origin` mapping upserts an existing active entity on a unique exact-match key and otherwise creates a stub; importing through `mention_origin` preserves mentions as mentions and never auto-creates stubs or auto-merges pre-existing entities.
- **AC-029**: File-based import preserves unknown columns in raw-capture or custom-attribute storage, records file kind, content hash, parser version, and selected sheet or region locator as provenance, leaves imported host or account tokens unresolved unless an analyst explicitly resolves them later, and never executes workbook formulas, macros or VBA, workbook automation, or external links during import.
- **AC-063**: Unsupported workbook features such as pivot tables, charts, comments, workbook protection, and merged-cell layout semantics are downgraded with a visible warning, preserved only as raw source metadata, or rejected as unsupported. No module outside the dedicated imports module links directly against XLSX or OpenXML parsing libraries or workbook-shape heuristics.

### 9.3 Snapshot and Reporting Extension Profile criteria

- **AC-030**: Report generation and snapshot generation run without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
- **AC-031**: A generated HTML report or presentation artifact opens in a disconnected browser without fetching remote JavaScript, CSS, or fonts.
- **AC-032**: Snapshot-derived outputs preserve stable identifiers and ordering consistent with the canonical derivation layer.
- **AC-056**: Rendering from the same `snapshot_id`, `source_change_set_high_watermark` or equivalent frozen source boundary, `derivation_version`, `template_id`, `template_version`, `redaction_profile_id`, and `redaction_profile_version` produces the same `export_model_sha256` and deterministic export ordering.
- **AC-057**: Rendering for a chosen `release_scope` fails closed if any rendered export-model field or block lacks an applicable redaction rule, and the rendered artifact includes a `redaction_manifest` keyed by stable export-model path and rule identifier.
- **AC-058**: Template rendering fails closed when a template references an undeclared export-model binding or a missing required field, and no report renderer performs live workbook-table reads after the snapshot tuple has been fixed.
- **AC-059**: For `internal_review`, exactly one `reviewer` approval is sufficient. For `external_release`, two distinct approvals are required, one `reviewer` and one `admin`. `internal_draft` requires no approval.
- **AC-060**: Changing `snapshot_id`, `template_id`, `template_version`, `redaction_profile_id`, `redaction_profile_version`, `output_kind`, `release_scope`, or rendered bytes invalidates prior release approvals automatically.
- **AC-061**: An `external_release` artifact contains no raw blob bytes or `working_material`, and any included `curated_narrative` block carries at least one `support_refs[]` entry to a supporting finding, event, evidence, assessment, or query record.
- **AC-071**: Given a snapshot containing ad hoc note artifacts and a separately curated narrative block derived from one of those notes, an `external_release` artifact excludes the raw note text while allowing only the curated narrative block that satisfies `support_refs[]` and applicable redaction rules.
- **AC-091**: Given a snapshot containing direct text from `task_request`, `decision`, `comm_log`, `handoff`, `status_review`, or `lesson` records and a separately curated export-model block derived from some of that content, an `external_release` artifact excludes the raw coordination-record text while allowing only the curated block that satisfies `support_refs[]` and applicable redaction rules.
- **AC-062**: `mermaid` and `slidev` outputs may be published as `external_release` only when every rendered block is eligible for the chosen `release_scope`, and `reenactment` outputs are visibly marked `generated_presentation=true` and rejected for `external_release`.
- **AC-064**: Selecting or changing `redaction_profile_id` and `redaction_profile_version` changes only snapshot-derived output and release state. It does not change live workbook query results, row visibility, field visibility, or evidence visibility for the same authenticated incident participant.
- **AC-065**: Given one immutable snapshot containing export-model fields or blocks tagged with `disclosure_partition_refs[]` for two different affected parties, rendering an `external_release` with a redaction profile that allows only one party excludes or redacts the other party's content and fails closed if mixed-partition content lacks an applicable rule.
- **AC-066**: Two `external_release` artifacts generated from the same immutable snapshot for two supported recipient-specific configurations require no manual post-render editing to satisfy the selected redaction profiles.

### 9.4 Reference Pack Extension Profile criteria

- **AC-033**: Reference-pack import, verification, and refresh run without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
- **AC-034**: Missing optional reference packs degrade only the affected overlay views or labels; timeline capture, entity resolution, and evidence attachment continue to function.
- **AC-035**: Pack activation fails closed on checksum, signature, compatibility, safe-path, or equivalent integrity mismatch, and the previously active version remains active.
- **AC-092**: In the smallest supported disconnected bundle, the only preinstalled active reference packs are `type_registry.host`, `type_registry.evidence`, and `type_registry.indicator`; the deployment remains usable without `framework.attack`, `framework.d3fend`, `framework.veris`, or any enrichment pack.
- **AC-093**: Given an offline pack bundle placed in the configured reference-pack storage root or uploaded through the equivalent admin surface, the system stages the bundle under a temporary-work root, verifies it, records the version as `available` or equivalent non-active state on success, and does not activate it until explicit operator action.
- **AC-094**: Given a candidate pack with checksum mismatch, signature mismatch, missing required integrity metadata, incompatible `pack_contract_version`, path-traversal attempt, or disallowed active content, import or activation fails closed, the candidate version remains inactive, and the previously active version, if any, remains active.
- **AC-095**: Pack import and activation record structured attestation metadata queryable without unpacking bundle contents, including `pack_key`, `pack_kind`, `pack_version`, `manifest_sha256`, `payload_sha256`, `source_identifier`, `verification_method`, signer-key or trusted-source identifier, imported and activated actor attribution with timestamps, `previous_active_version`, and `verification_result`.
- **AC-096**: In a disconnected deployment, reference-pack import or activation succeeds without outbound network access and no supported pack-activation path performs a live internet fetch.

### 9.5 Enterprise Authentication Extension Profile criteria

- **AC-036**: Enterprise-authenticated users map to the same internal user identity used for attribution, and switching from local auth to enterprise auth does not break audit lineage for existing incidents.

### 9.6 Additional Base Profile criteria for same-field conflict resolution

- **AC-037**: When two analysts edit the same write-back-capable field concurrently, the losing client shows a conflicted cell and same-surface resolver that presents saved value, unsaved value, row context, actor, and timestamp without leaving the sheet.
- **AC-038**: Choosing `Keep saved value` clears the local conflict without creating a source revision; choosing `Use my unsaved value` or `Edit merged value` creates a new attributed change set and updates the visible row.
- **AC-039**: If the field changes again while the resolver is open, stale resolution is rejected and the latest conflict payload is shown without losing the analyst's unsaved draft.
- **AC-040**: A paste containing both non-conflicting and same-field-conflicting cells commits the non-conflicting cells immediately and groups the conflicting cells into a navigable conflict queue without per-cell modal interruption.
- **AC-041**: Unresolved local conflict drafts are not broadcast to other analysts and do not appear in search, history, exports, or snapshots unless explicitly committed.
- **AC-042**: After resolving a conflict, focus returns to the same cell and scroll position is preserved.

### 9.7 Additional Base Profile criteria for threat model and focused weakness controls

- **AC-048**: The implementation maintains a STRIDE threat model for the current release that covers, at minimum, authenticated sessions, incident records and revisions, evidence blobs and previews, reference packs and import bundles, generated snapshots and exports, and portable runtime roots, and each entry maps the threat to at least one control and one verification hook.
- **AC-049**: Rendering notes, markdown, evidence metadata, filenames, tags, and other incident-authored text in the browser UI or generated HTML does not execute script, inline event handlers, `javascript:` URLs, or remote asset fetches sourced from incident data.
- **AC-050**: CSV, XLSX, and spreadsheet-oriented clipboard exports neutralize formula-leading characters by default. A raw export mode, if implemented, requires explicit operator opt-in and a visible unsafe-export warning.
- **AC-051**: Upload, import, and archive-extraction paths reject absolute paths and parent traversal and do not write outside the configured runtime roots. Client-supplied filenames do not determine storage keys.
- **AC-052**: Reference-pack activation, and incident bundle import when implemented, fail closed on checksum mismatch, signature mismatch, incomplete download, or missing required integrity metadata.
- **AC-053**: Evidence upload and preview do not execute uploaded active content in the main application unit or browser origin. Non-previewable or active-content types remain quarantined or download-only unless an explicit isolated analysis path is configured.
- **AC-054**: Attempting to mutate, preview, download, or issue object-store access for a `record_id`, `evidence_record_id`, `object_blob_id`, or snapshot outside the caller's incident membership is denied even when the identifier is otherwise valid.
- **AC-055**: A deployment claiming flyaway or disconnected portability for use on portable hosts or removable media stores database, object, reference-pack, temporary-work-file, and export roots on encrypted storage. A deployment that does not do so is non-conformant for flyaway handling.

## 10. Non-goals preserved from the source artifact

The following remain explicit non-goals for current conformance:

- generalized field-level ACLs,
- generalized approval workflows outside the bounded artifact-release gate in the Snapshot and Reporting Extension Profile,
- full spreadsheet formula engines,
- merged cells,
- character-by-character Google-Sheets-style cell CRDT behavior,
- manual row-range grouping,
- arbitrary field-picker rollback,
- automatic entity merge based on fuzzy similarity alone.

## 11. Operational posture

A conformant Cartulary deployment MUST prefer operational simplicity over distributed architectural purity.

The smallest useful deployment is intentionally not the absolute minimum number of containers. It is the minimum that keeps binary evidence handling sane and portable while preserving collaboration, authentication, and auditable source-of-truth behavior.
