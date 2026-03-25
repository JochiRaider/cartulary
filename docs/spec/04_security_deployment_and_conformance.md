# Cartulary Normative Core 04: Security, Deployment, and Conformance

## 1. Authentication model

### 1.1 Base authentication

The base profile MUST support:

- local user accounts stored in Postgres,
- password hashing with Argon2id,
- TOTP MFA,
- optional WebAuthn when the environment supports it.

The public API and WebSocket surface MUST use a server-managed session contract rather than a client-parsed identity token contract. Browser clients MUST authenticate with an `HttpOnly` `Secure` cookie carrying an opaque session token. The implementation MAY additionally accept `Authorization: Bearer <opaque_session_token>` for non-browser clients or trusted automation. The public token format MUST remain opaque and MUST NOT require clients to parse JWT claims or provider assertions.

State-changing HTTP requests authenticated by cookie MUST use CSRF protection that fails closed, such as a synchronizer token or an equivalent same-origin mechanism.

Authentication MUST work in disconnected deployments and MUST NOT depend on enterprise infrastructure for the base profile.

#### 1.1.1 Session lifecycle boundaries

The base profile MUST create one server-side session record for each login-capable session.

The session credential presented by the browser or bearer client MUST identify that server-side session record and MUST be an opaque CSPRNG-generated bearer token with at least `128 bits` of unpredictability. Browser cookies carrying that token MUST be set with `HttpOnly`, `Secure`, `Path=/`, and `SameSite=Lax` or a stricter same-site policy. If bearer authentication is enabled for non-browser clients, it MUST represent the same opaque session family rather than a separate JWT or API-key contract.

The base profile MUST NOT require a separate long-lived browser refresh token.

Each session MUST persist `authenticated_at`, `last_qualifying_activity_at`, `idle_expires_at`, `absolute_expires_at`, and `session_expires_at`, where `session_expires_at` is the earlier of `idle_expires_at` and `absolute_expires_at`.

`idle_expires_at` MUST be computed as `last_qualifying_activity_at + 30 minutes`. `absolute_expires_at` MUST be computed as `authenticated_at + 12 hours`.

Qualifying activity MUST include successful authenticated user-initiated workbook or API activity. Qualifying activity MUST NOT include WebSocket `ping` or `pong`, passive server push, automatic reconnect or replay, or `GET /api/v1/auth/session`.

Qualifying activity MAY slide `idle_expires_at`, but MUST NOT extend the session beyond `absolute_expires_at`.

If a session expires, establishing a new session MUST require fresh primary authentication and any applicable MFA requirement.

The base profile MUST cap each human user at 5 concurrently active sessions. Multiple tabs sharing one browser session count as one session for this limit. Explicit system-process actors are outside this limit.

When a new login would exceed the concurrent-session cap, the server MUST revoke the least-recently-used non-current session before issuing the new session and MUST record an attributed audit event with stable reason code `concurrency_limit`.

`POST /api/v1/auth/logout` MUST revoke only the current session immediately.

Password change, MFA reset, account disablement, or an explicit deployment-admin revoke-all action MUST revoke all active sessions for that user immediately.

Loss of incident membership for one subscribed incident MUST terminate the affected incident WebSocket subscription and MUST cause future incident-scoped authorization checks for that incident to fail closed, but it MUST NOT by itself require revocation of otherwise valid account sessions for other authorized incidents.

### 1.2 Enterprise Authentication Extension Profile

If the implementation claims the **Enterprise Authentication Extension Profile**, it MUST support provider-backed identities through an `auth_providers` and `auth_identities` model equivalent to the source artifact.

OIDC is the preferred enterprise path. SAML is the secondary enterprise path when required by the environment.

External identities MUST map to the same internal user identity used for attribution so that audit semantics remain unchanged.

When the Enterprise Authentication Extension Profile is implemented, successful provider authentication MUST terminate into the same server-managed session contract used by the base profile so the remaining API surface stays provider-agnostic.

## 2. Authorization model

The base profile MUST use incident-level roles equivalent to:

- `viewer`,
- `editor`,
- `reviewer`,
- `admin`.

Record access MUST inherit from incident access in the base profile.

API routes, preview or download handle issuance, job polling, and WebSocket incident subscriptions MUST re-derive authorization from the caller's current incident membership and role at request time.

`task_request`, `decision`, and coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MUST inherit the same incident-level authorization model. The base profile MUST NOT introduce record-specific ACLs or hidden sub-workspaces for these objects.

Saved-view scope controls only discoverability and mutability of the saved-view configuration object. It MUST NOT widen or narrow access to underlying incident rows, fields, search results, exports, or evidence.

Any incident member MAY create, update, or delete their own `private` saved views and set or clear their own `home_sheet_ref` even when their incident role is `viewer`, because those actions mutate personal workbook configuration rather than incident facts. Incident-wide default-surface updates and in-place mutation of another user's saved views MUST follow the scope and role rules defined by Core 03.

Field-level ACLs, generalized approval workflows, and generalized record-level ACL systems are out of scope for the base profile.

If the Snapshot and Reporting Extension Profile is implemented, export redaction MUST NOT restrict live workbook views, search results, filters, saved views, row visibility, field visibility, or evidence visibility for authenticated incident participants. In the base profile, live workspace visibility is derived only from incident membership and the incident-level role model. In that profile, recipient-specific withholding MUST be implemented at snapshot, render, and release time rather than by hiding live workspace content.

The base profile MUST separate deployment-local account administration from incident-scoped data authorization. A conformant deployment MUST define one narrow deployment-scoped capability equivalent to `deployment_admin`. This capability authorizes only deployment-local user-account inspection and administration, including local-user creation, user patching, and any explicit revoke-all action the deployment exposes. The first holder of this capability MUST be provisioned out of band at install or bootstrap time.

Holding `deployment_admin` MUST NOT by itself grant incident read, write, preview, download, export, job, or WebSocket access. A caller who is `deployment_admin=true` but lacks current membership in incident X MUST have no incident-data access to incident X until granted ordinary incident membership.

Incident membership create, role-change, and delete routes remain incident-scoped authorization decisions. In the base profile, those routes MUST require current incident role `admin`; `deployment_admin` alone MUST NOT bypass that requirement.

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

For rendered-output lifecycle, the authoritative artifact state MUST be stored on the release record or an equivalent artifact-scoped record. The closed vocabulary for `release_state` is `pending_approval`, `approved`, `invalidated`, and `published`.

For this lifecycle, the logical output slot is the release tuple excluding `output_sha256`.

A release record enters `pending_approval` when bytes and `output_sha256` exist for one bound release tuple but the required approvals are not yet complete. It enters `approved` only when the approval requirements above are satisfied for that exact artifact. It enters `published` only through an explicit publish action after approval. It enters `invalidated` when a different artifact for the same logical output slot supersedes it, when its rendered bytes change, or when the implementation can no longer attest that the required approval set still applies to that exact artifact.

A new render with a different logical output slot or different `output_sha256` MUST start as `pending_approval`. It MUST NOT inherit `approved` or `published` state from an earlier artifact.

A narrow live sensitive-evidence model MAY be added in future work if repeated real-world incidents show that export-scoped withholding is insufficient. It is not a current conformance requirement.

## 3. Attribution and audit requirements

Every mutation MUST originate from an authenticated session or an explicitly identified system process.

Every mutation MUST record, at minimum:

- actor user identifier,
- timestamp,
- mutation source such as `ui`, `import`, or `rollback`,
- before/after values or equivalent patch data at the required history granularity.

User-account and incident-membership administration mutations MUST be captured with the same minimum actor, timestamp, source, and before/after fidelity even when they are stored outside incident `change_set` or `record_revisions` rows. Those administrative audit records are deployment-local state and MUST NOT be serialized into whole-incident portability bundles.

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

If the Incident Portability Extension Profile is implemented:

- whole-incident portability bundles MUST serialize only authoritative incident source state, deterministic structured files, and referenced blob bytes, not projections or deployment-local runtime state,
- bundle import MUST stage content under the configured temporary-work root and verify required checksums before any structured incident data becomes visible,
- unsupported or missing optional embedded snapshot or reference-pack sections MUST NOT block import of the core incident state,
- portability bundles, staged extracts, and emitted artifacts for flyaway or disconnected use MUST remain on encrypted storage roots.

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
- **CWE-352**: State-changing HTTP routes authenticated by cookie MUST require CSRF protection that fails closed. WebSocket upgrades and any incident subscription step MUST verify the authenticated session and incident authorization before joining an incident-scoped stream. Cookie-authenticated browser WebSocket connections MUST reject untrusted `Origin` values before the socket joins an incident-scoped stream.
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

Within this document, subsection numbers MUST be unique and strictly increasing, acceptance-criterion identifiers MUST be unique, and a retired or superseded acceptance-criterion identifier MUST NOT be reassigned or recycled.

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
- **AC-116**: With no optional reference packs installed or activated, the required built-in and system surfaces still resolve to exactly these nine pack-independent `view_schema_id` values: `cartulary.view.timeline.v1`, `cartulary.view.hosts.v1`, `cartulary.view.identities.v1`, `cartulary.view.evidence.v1`, `cartulary.view.notes.v1`, `cartulary.view.indicators.v1`, `cartulary.view.assessments.v1`, `cartulary.view.task_requests.v1`, and `cartulary.view.decisions.v1`.
- **AC-117**: The structured contract for each of those nine base-profile view schemas includes hidden `record_id` and `row_version`, an ordered field list, an ordered default sort tuple, and the declared filter whitelist; the final default-sort tiebreaker is `record_id`.
- **AC-118**: Every writable field in those nine base-profile view schemas declares a stable `field_key` and `conflict_resolution_class`, and every entity-bearing writable field also declares `entity_binding_mode`.
- **AC-119**: In the Timeline schema, `timeline.summary`, `timeline.details`, and `timeline.source_text` are distinct field keys with distinct write targets. Editing one does not overwrite either of the other two fields.
- **AC-120**: A write routed to a field that the active base-profile view schema declares read-only or derived fails closed, does not mutate authoritative source state, and does not persist a misleading projection update.
- **AC-121**: The Assessments view accepts band-first creation using the canonical `NULL`, `25`, `55`, and `85` `confidence_score` mapping and rejects in-place mutation of an existing row's `subject_ref`, `subject_type`, `assessment_state`, `confidence_score`, `rationale`, `assessor`, or `assessed_at`.
- **AC-122**: The Indicators view allows inline creation of a new canonical indicator row and rejects in-place mutation of an existing row's identity-defining fields, including at minimum `indicator_type`, `value_kind`, canonical display value, `normalized_value`, and any type-specific dedupe basis.
- **AC-146**: Analyst A can create a `private` saved view over one `view_schema_id`; `GET /api/v1/incidents/{incident_id}/saved-views` for analyst A returns it, the same call for analyst B on the same incident does not, and the same call for an incident admin does.
- **AC-147**: A `shared` saved view created in one incident is returned to all incident members; a non-owner non-admin member can open and duplicate it, but an in-place patch or delete by that member fails closed.
- **AC-148**: A visible `system` saved view cannot be created, patched, or deleted through the ordinary saved-view routes, but any incident member can duplicate it into a new saved view allowed by policy.
- **AC-149**: Changing saved-view scope, updating saved-view `query_json` or `layout_json`, or deleting a saved view never changes underlying row visibility, field visibility, evidence visibility, search results, or export-redaction behavior for incident participants whose incident membership is unchanged.
- **AC-150**: Workbook open resolves the starting surface in this order: explicit `sheet_ref`, caller `home_sheet_ref`, incident `default_sheet_ref`, then `cartulary.view.timeline.v1`; if a referenced saved view is deleted, hidden by scope, or invalid because a required optional pack is unavailable, the invalid pointer is cleared and the implementation falls through deterministically to the next step instead of failing workbook open.
- **AC-112**: An analyst can create a note directly in the Notes sheet by blank-row or equivalent grid-native entry with no required modal, and the resulting record is stored as an artifact with `artifact_type='note'`.
- **AC-068**: From a selected Timeline, Host, Identity, or Evidence record, an analyst can invoke `add linked note` without leaving the workbook flow; the resulting note is stored with the same artifact record shape as a Notes-sheet-created note, appears in the Notes sheet, and is visible as a linked record from the source row.
- **AC-069**: Renaming the visible Notes tab label, and any implementation-supported per-user hide/show operation for that built-in tab, does not change write-back or export semantics because behavior remains bound to `view_schema_id`.
- **AC-070**: Editing a lightweight free-text field on a timeline, host, identity, or evidence record does not create a standalone Notes row. Creating a standalone note does create a distinct note record that can be linked, tagged, and reviewed in history independently.
- **AC-015**: An analyst can create an evidence request record without uploading a blob, later attach or replace the blob, and preserve structured `requested_at`, `received_at`, `storage_ref`, `blob_hash`, `collector_party_text`, `source_party_text`, and append-only custody history across the lifecycle.
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
- **AC-085**: From a selected Notes, Timeline, Host, Identity, or Evidence context, an analyst can create a first-class `task_request` with required `task_kind`, `status`, `owner_user_id`, `priority`, `created_at`, and `updated_at`; optional structured fields including `workstream`, `requester_party_text`, `due_at`, `blocked_reason`, `completed_at`, and `external_ticket_ref` remain editable and filterable; a blocked task requires `blocked_reason`; a done task requires `completed_at`; and workbook filtering can show open tasks by owner, blocked state, due status, workstream, and external ticket reference.
- **AC-086**: Creating a `decision` record preserves `decision_type`, `status`, `owner_user_id`, `decided_at`, and `rationale`; reviewer history can reconstruct when the decision changed state, who owned it, and what `support_refs[]` were attached.
- **AC-087**: A `comm_log` artifact can record required `comm_type`, `timestamp_utc`, `audience`, `channel_or_meeting`, and `summary`, while a `status_review` artifact can record required timestamp and summary plus linked decision or task references, and a workbook surface can filter or sort those artifacts by `comm_type`, `audience`, timestamp, and next-report or next-checkpoint time without rereading unrelated note text.
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
- **AC-024**: In the timeline sheet, the grouping control offers only `None`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions` in the base profile, and no other grouping key. In particular, it does not offer `timeline.event_type`, Summary, Hosts, Identities, Tags, or arbitrary custom columns.
- **AC-025**: A grouped timeline sheet exposes exactly one derived outline level. `expand group`, `collapse group`, `expand all`, `collapse all`, and `Group: None` work without creating editable rows, paste targets, subtotal rows, or `record_id`-bound mutation targets.
- **AC-026**: While grouped, sorting, filtering, paste, autosave, conflict handling, rollback, and export flatten to underlying records only. Editing a grouped field may move the row to a different visible group, but drag-and-drop reclassification and manual row-range grouping are not available.

### 9.2 Import Extension Profile criteria

- **AC-027**: One uploaded CSV file or XLSX workbook import session starts without blocking grid editing, the UI shows progress and cancellation within 1 second of job start, and the operator can review per-unit discovery, preview, and header mapping before any apply step.
- **AC-028**: File-based import uses the same mapping engine and `entity_binding_mode` semantics as clipboard-driven structured ingest. Importing through an `entity_origin` mapping upserts an existing active entity on a unique exact-match key and otherwise creates a stub; importing through `mention_origin` preserves mentions as mentions and never auto-creates stubs or auto-merges pre-existing entities.
- **AC-029**: File-based import preserves unknown columns in raw-capture or custom-attribute storage, records `import_session_id`, `import_unit_id`, `mapping_fingerprint`, file kind, content hash, parser profile, parser version, and selected sheet or region locator as provenance, leaves imported host or account tokens unresolved unless an analyst explicitly resolves them later, and never executes workbook formulas, macros or VBA, workbook automation, or external links during import.
- **AC-063**: Unsupported workbook features are downgraded only through the closed `warning_code[]` vocabulary declared by Core 03, preserved only as raw source metadata, or rejected as unsupported. No module outside the dedicated imports module links directly against XLSX or OpenXML parsing libraries or workbook-shape heuristics.
- **AC-064**: One uploaded XLSX workbook creates one `import_session` that discovers explicit candidate `import_unit` objects for parser-resolved used ranges, Excel tables, eligible named ranges, and operator-selected regions. Each discovered unit exposes a deterministic `locator_kind`, canonical locator, inferred rectangular extent, and any nonblocking `warning_code[]` values before apply.
- **AC-065**: Re-applying the same `(import_unit_id, mapping_fingerprint, incident_id)` tuple triggers duplicate-apply detection and blocks by default until the operator explicitly chooses re-import.
- **AC-066**: Selecting overlapping `import_unit` rectangles in one batch is blocked before apply with an explicit overlap diagnostic. Non-overlapping selected units apply in deterministic order, one atomic `change_set` per unit, and the session can finish in `partially_applied`.
- **AC-067**: Preview or apply of a unit containing formulas, merged cells, comments, pivots, charts, external links, hidden or filtered presentation state, or workbook or sheet protection never executes workbook behavior. The unit either emits only declared `warning_code[]` values or is rejected as unsupported, formula cells without stored cached values do not enter `ready` while mapped, and encrypted or password-protected workbooks that cannot be parsed are rejected before discovery.

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
- **AC-113**: Selecting or changing `redaction_profile_id` and `redaction_profile_version` changes only snapshot-derived output and release state. It does not change live workbook query results, row visibility, field visibility, or evidence visibility for the same authenticated incident participant.
- **AC-114**: Given one immutable snapshot containing export-model fields or blocks tagged with `disclosure_partition_refs[]` for two different affected parties, rendering an `external_release` with a redaction profile that allows only one party excludes or redacts the other party's content and fails closed if mixed-partition content lacks an applicable rule.
- **AC-115**: Two `external_release` artifacts generated from the same immutable snapshot for two supported recipient-specific configurations require no manual post-render editing to satisfy the selected redaction profiles.

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

### 9.8 Additional Base Profile criteria for promoted recurrent fields

- **AC-097**: The Hosts sheet can sort or filter on `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` from the grid surface without opening the inspector and without rereading unrelated note text or blob metadata.
- **AC-098**: The Identities sheet can sort or filter on `privilege_level`, `mfa_state`, and `reset_status` from the grid surface without rereading unrelated note text or evidence blobs.
- **AC-099**: On the implementation-supported incident metadata surface, editing `tlp`, `current_phase`, and `primary_external_case_ref` persists those values as structured fields that survive reload, deterministic projection rebuild, and snapshot generation; they do not disappear into `custom_attrs`.
- **AC-100**: The Evidence sheet can sort or filter on `requested_at`, `received_at`, `collector_party_text`, `source_party_text`, `storage_ref`, `blob_hash`, and upload or attachment state without fetching blob bytes.
- **AC-101**: If the implementation exposes structured findings, investigative queries, or forensic keywords as workbook surfaces, their defining fields remain directly filterable and exportable as structured fields rather than JSON-only payloads.

### 9.9 Lifecycle machine criteria

- **AC-102**: A blob slot left in `pending` without successful finalization does not create or imply an attached evidence record, does not increment visible evidence counts, and by no later than the first cleanup sweep after `pending_expires_at` transitions to `upload_state='failed'` with `terminal_reason='pending_timeout'`.
- **AC-103**: If an evidence record points at a blob in `pending`, `failed`, or missing state, preview and download are blocked and the row surfaces as non-available or inconsistent until explicit repair or re-finalization completes.
- **AC-154**: A pending blob slot allows 3 failed explicit finalization attempts; the 4th failed attempt transitions the slot to `upload_state='failed'` with `terminal_reason='finalize_retry_exhausted'`; only failed explicit finalization attempts count toward this total, and an idempotent replay after already-committed success does not consume retry budget.
- **AC-155**: Pending-blob timeout handling runs at least every 15 minutes; by no later than `pending_expires_at + 15 minutes`, an unfinalized slot is failed, orphaned bytes for a failed unattached slot are deleted within 1 hour of terminal failure, and failed unattached slot metadata remains queryable for at least 7 days before automatic hard deletion.
- **AC-104**: Rendering a report or presentation artifact creates a release record in `pending_approval` bound to one immutable release tuple and one `output_sha256`.
- **AC-105**: Satisfying the required approval set moves that exact release record to `approved`, and any attempted publish action on a non-approved record is rejected.
- **AC-106**: Rendering a superseding artifact for the same logical output slot with a different bound tuple or different `output_sha256` creates a new `pending_approval` candidate and invalidates the prior current artifact rather than inheriting its approval or publication state.
- **AC-107**: For every normative lifecycle machine claimed by the implementation, the happy-path transitions succeed only in the documented order and the persisted authoritative state matches the documented machine condition after each step.
- **AC-108**: For every normative lifecycle machine claimed by the implementation, at least one terminal failure path moves to the documented failure or non-active state and blocks any forbidden follow-on action.
- **AC-109**: Illegal lifecycle transitions are rejected with no partial state advancement and no false success signal.
- **AC-110**: Retrying the last lifecycle transition after a simulated crash or duplicate delivery is idempotent or otherwise produces the documented equivalent final state without duplicating durable side effects.
- **AC-111**: Replaying the same starting state and the same ordered inputs for a normative lifecycle machine produces the same final persisted state and the same documented observable signals.
- **AC-137**: A `task_request` can transition `open -> in_progress -> blocked -> in_progress -> done`; `blocked_reason` is present only while `status='blocked'`; `completed_at` is present only while `status='done'`; and each committed transition increments `row_version`, updates the task projection row, appends one `change_set` plus mutation entries, and emits `record_changed`.
- **AC-138**: A `task_request` transition `done -> open` succeeds, clears persisted `completed_at`, preserves prior history, and returns committed row values that show the reopened task in `status='open'`.
- **AC-139**: A requested `task_request` transition `done -> canceled` or `canceled -> done` is rejected with `error.code='illegal_transition'`, `error.status=409`, `error.details.from_status`, `error.details.to_status`, and `error.details.violated_guards[]`; no status, guard field, projection row, `change_set`, or mutation entry is partially advanced.
- **AC-140**: A `task_request` transition away from `blocked` clears persisted `blocked_reason`, and a successful write that sets `status='done'` without an explicit `completed_at` stores `completed_at` at or after the commit time and not earlier than `created_at`.
- **AC-141**: A `decision` can transition `proposed -> approved -> executed`, and `approved` is treated only as an incident-coordination state rather than as a generalized approval workflow for ordinary row edits.
- **AC-142**: A direct write that sets `decision.status='superseded'` is rejected with `error.code='illegal_transition'` and `error.status=409`, and a direct write `approved -> rejected`, `rejected -> proposed`, or `executed -> approved` is likewise rejected with no partial state advancement.
- **AC-143**: An explicit supersession action from decision B to decision A succeeds only when B and A are different decisions in the same incident and B already has `status` `approved` or `executed`; the committed action persists the supersession relation, moves A from `proposed` or `approved` to `superseded`, increments `row_version`, appends one `change_set` plus mutation entries for the changed records, updates derived projections, and leaves B in its preexisting `approved` or `executed` status.
- **AC-144**: If an explicit supersession action targets a decision already in `executed`, the target record remains `executed`, the supersession relation still persists, and the decision view surfaces `decision.is_superseded=true` or an equivalent computed indicator for that target without rewriting its persisted `status`.
- **AC-145**: Replaying the same legal `task_request` status change, legal `decision` status change, or legal explicit supersession action after a simulated crash or duplicate delivery is idempotent and does not duplicate durable side effects such as extra `change_set` records, extra mutation entries, projection updates, or repeated status flips.

### 9.10 Additional Base Profile criteria for public interface surface

- **AC-123**: `GET /api/v1/auth/session` returns the authenticated internal user identity, display name, provider kind, MFA state, `is_deployment_admin`, `authenticated_at`, `idle_expires_at`, `absolute_expires_at`, and `session_expires_at`, plus the caller's incident memberships or current incident-role context, without requiring client-side token parsing; `session_expires_at` is the earlier of `idle_expires_at` and `absolute_expires_at`.
- **AC-124**: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` accepts a field-key-based sort, filter, and grouping contract and returns rows with `record_id`, `row_version`, field-key-addressable cells, and cursor metadata; group headers are not serialized as writable rows.
- **AC-125**: View-scoped row creation and record-scoped patch operate via `view_schema_id`, `client_txn_id`, `base_row_version`, and `changes[]` keyed by `field_key`; the client can mutate one writable field without resubmitting a full row snapshot.
- **AC-151**: `GET /api/v1/incidents/{incident_id}/saved-views` returns only saved views visible to the caller, and each returned resource includes `saved_view_id`, `incident_id`, `view_schema_id`, `scope`, `display_name`, `query_json`, `layout_json`, `owner_user_id`, timestamps, and `saved_view_version`; `query_json` uses stable `field_key` and grouping identifiers rather than visible labels.
- **AC-152**: `POST /api/v1/incidents/{incident_id}/saved-views` defaults omitted `scope` to `private` and rejects `scope='system'`; `PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` accepts `base_saved_view_version` and mutable fields only, and a stale base version fails with an explicit conflict status rather than silently overwriting saved-view state; `DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` deletes only the configuration object and never underlying incident records or links.
- **AC-153**: `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` read or replace only the caller's nullable `home_sheet_ref` and succeed for any incident member, including `viewer`; `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` read or replace the incident's nullable `default_sheet_ref`, and the `PUT` route fails closed for non-admin incident roles.
- **AC-126**: Same-field conflict responses use the generic error envelope with `error.code='same_field_conflict'` and the conflict object required by Core 03 §3.3.4; a stale `conflict_token` is rejected with a fresh conflict payload rather than silently overwriting saved state.
- `TODO: assign next AC id`: Patching `timeline.tags` with one `add_tag` action and one `remove_tag` action adds exactly one incident-scoped tag binding and removes exactly one binding.
- **AC-188**: Patching `timeline.host_refs` or `timeline.identity_refs` with `dismiss_item` against an unresolved or resolved `entity_mention` preserves `raw_text`, leaves the same mention row and provenance intact, sets `resolution_status='dismissed'`, clears `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`, removes or tombstones any corresponding active resolved link in the same `change_set`, and appends exactly one new `change_set` with mention-target mutation detail.
- **AC-189**: Patching `timeline.host_refs` or `timeline.identity_refs` with `revert_to_unresolved` against a dismissed `entity_mention` returns that same mention row to `resolution_status='unresolved'`, preserves `raw_text`, leaves `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method` null after commit, leaves no active derived resolved link, and appends exactly one new `change_set`; reviewer rollback of the dismissal restores the exact pre-dismiss state as a separate attributed revision rather than rewriting history.
- **AC-190**: If a row's last non-deleted unresolved host or identity mention is dismissed, the committed row computes `timeline.has_unresolved_mentions=false`, the row no longer matches unresolved-only filters or equivalent unresolved-resolution queue views, grouping by `timeline.has_unresolved_mentions` places the row in the `false` bucket, and the active `timeline.host_refs` or `timeline.identity_refs` collection value omits the dismissed mention while history or inspector affordances can still surface it.
- `TODO: assign next AC id`: Patching `timeline.host_refs` with `resolve_item` preserves `raw_text`, resolves the targeted `entity_mention`, and leaves the corresponding active host link present after commit.
- `TODO: assign next AC id`: Patching `host.aliases` or `identity.aliases` with `add_alias` plus `remove_alias` yields a committed alias collection that round-trips as `collection_value_v1`.
- `TODO: assign next AC id`: A same-field conflict payload for a `collection_review` field returns `collection_value_v1` in `client_value`, `server_value`, and `base_value` rather than a raw string array or plain delimited text.
- `TODO: assign next AC id`: Sending a raw string, blind full-collection replacement, unknown collection action `op`, unknown payload `kind`, or foreign `item_ref` to a `collection_review` field fails with `400` and `error.code='invalid_mutation_payload'`.
- **AC-127**: Potentially large list or view-query routes use opaque cursor pagination with `has_more` and `next_cursor`; replaying a cursor against a different incident, view, sort tuple, filter set, or grouping contract is rejected rather than reinterpreted.
- **AC-128**: `POST /api/v1/object-blobs` creates a pending blob slot and returns `object_blob_id`, `upload_state`, `target_expires_at`, `pending_expires_at`, and a short-lived upload target; in the base profile the upload target expires 60 minutes after issuance, the pending slot expires 24 hours after creation, retry after target expiry uses a fresh blob slot rather than same-slot target refresh, and attaching that blob to incident-visible evidence, or previewing the resulting evidence surface, fails closed until the blob or evidence lifecycle reaches the states required by Core 03 §8 and Core 02 §14.1.
- **AC-129**: Long-running operations started through the public interface return `202 Accepted` with a `job_id`; `GET /api/v1/jobs/{job_id}` exposes progress and terminal result or error summary; the incident-scoped WebSocket stream authenticates with the same session contract and emits presence, `record_changed`, and `job_progress` events without broadcasting client-local drafts or grouping UI state.
- **AC-130**: A cookie-authenticated state-changing request with missing or invalid CSRF proof fails closed, and the same request succeeds when a valid CSRF proof accompanies an otherwise authorized session.
- **AC-131**: On `GET /ws/v1/incidents/{incident_id}`, an authorized same-origin client that sends `hello` as its first application message receives `hello_ack` containing `connection_id`, `resume_token`, `server_time`, `heartbeat_interval_ms=15000`, `presence_ttl_ms=45000`, and `resume_window_ms>=300000`, followed by `presence_snapshot`; a cookie-authenticated browser connection from an untrusted `Origin` is rejected before `hello_ack` or `presence_snapshot`, and an otherwise idle accepted connection receives application-level `ping` within 15 seconds and is closed after 45 seconds without any inbound frame or timely `pong`.
- **AC-132**: When analyst A changes workbook surface, focused row, or same-cell edit state on an incident, analyst B subscribed to the same incident receives the corresponding `presence_delta` within 1 second, and the UI renders workbook-header, row-gutter, and same-cell indicators from matching `sheet_ref`, `record_id`, and `field_key` rather than visible labels or row numbers.
- **AC-133**: After a transient disconnect, a client that reconnects to the same incident within the retained replay window can send `resume` with `resume_token` and `last_seen_stream_seq`, receives `resume_ack.status='replayed'`, then receives missed `record_changed` and incident-scoped `job_progress` messages in strict ascending `stream_seq`, followed by a fresh `presence_snapshot`.
- **AC-134**: If a client reconnects with an expired, unknown, malformed, or too-old `resume_token`, but the caller still has valid incident authorization, the server responds with `resume_ack.status='reset_required'`, sends a fresh `presence_snapshot`, and emits no guessed or partial replay for the missing range; the client recovers only by re-querying current workbook state through the existing HTTP view route.
- **AC-135**: Replayable WebSocket messages on one incident carry monotonically increasing `stream_seq` values assigned only after the underlying record mutation or incident-scoped job-state change is committed; clients can de-duplicate duplicates by `(incident_id, stream_seq)`, reconcile row application by `record_id` plus `row_version`, and must perform HTTP resynchronization rather than guessed incremental apply when a replayable sequence gap is observed.
- **AC-136**: If the authenticated session expires or the caller's incident membership is revoked after connection establishment, the server sends `session_revoked`, closes the socket, and emits no further incident `presence`, `record_changed`, or `job_progress` messages to that client on that connection.
- **AC-156**: After 30 minutes with no qualifying activity, the next request on that session to `GET /api/v1/auth/session`, query a workbook view, or submit a mutation fails with `401` or an equivalent `session_expired` error, and no mutation or job start is partially applied.
- **AC-157**: Continuous qualifying activity can slide `idle_expires_at` but cannot keep a session alive past `authenticated_at + 12 hours`; after absolute expiry, the next authenticated request fails closed and any accepted WebSocket connection on that session receives `session_revoked` with `reason_code='session_expired'` before close.
- **AC-158**: `GET /api/v1/auth/session`, WebSocket `ping` or `pong`, passive server events, and successful `resume` replay do not count as qualifying activity; a session that receives only those events for 30 minutes still expires on the normal idle boundary.
- **AC-159**: A 6th concurrent human-user login revokes the least-recently-used non-current session before issuing the new session, records an attributed audit event with reason code `concurrency_limit`, and any accepted WebSocket connection on the revoked session receives `session_revoked` with `reason_code='concurrency_limit'` before close.
- **AC-160**: `POST /api/v1/auth/logout` invalidates only the current session immediately, causes any accepted WebSocket connection on that session to receive `session_revoked` with `reason_code='session_revoked'`, and leaves other still-valid sessions for that user authorized until another revoke condition occurs; password change, MFA reset, account disablement, or an explicit deployment-admin revoke-all action invalidates all active sessions immediately.
- **AC-161**: A `resume_token` is rejected as an HTTP authentication credential, becomes unusable after the earlier of replay-window expiry and underlying session expiry or revocation, and after session expiry or revocation a reconnect succeeds only after a new authenticated session is established and the client begins again with `hello` rather than relying on `resume` alone.
- **AC-162**: If a user loses membership in incident A while retaining a still-valid session and access to incident B, the connection and future requests for incident A fail closed, the incident-A socket receives `session_revoked` with `reason_code='incident_access_revoked'`, and the same session can still query or subscribe to incident B if otherwise authorized.
- **AC-163**: If session expiry or revocation occurs while the client holds queued unsent patches or unresolved same-field local drafts, the client preserves that unsaved work locally, prompts for re-authentication when required, and retries only through the normal authenticated patch or conflict-resolution paths after a new session is established; no queued draft becomes authoritative without passing the ordinary row-version, authorization, and conflict checks.

- **AC-170**: An authenticated session whose internal user account is active and not disabled can `POST /api/v1/incidents` with `client_txn_id`, `incident_key`, and `title`, receives `201 Created` plus `Location: /api/v1/incidents/{incident_id}`, and the response `data` includes the incident resource fields defined by Core 01 §3.3.5.3.
- **AC-171**: A successful incident create bootstraps exactly one creator membership with `role='admin'`, `GET /api/v1/incidents/{incident_id}/memberships` returns that membership for the creator, `GET /api/v1/incidents/{incident_id}/workbook-preferences/default` returns a resource with `default_sheet_ref=null`, and `GET /api/v1/incidents/{incident_id}/workbook-preferences/me` for the creator returns a resource with `home_sheet_ref=null`.
- **AC-172**: `POST /api/v1/incidents` without authentication, with an expired or revoked session, from a disabled internal user account, or with missing or invalid CSRF protection for a cookie-authenticated request fails closed and creates no incident, membership, or workbook-preference state.
- **AC-173**: Replaying `POST /api/v1/incidents` with the same `(actor_user_id, client_txn_id)` and the same normalized request returns `200 OK`, repeats `Location: /api/v1/incidents/{incident_id}` for the originally created incident, returns the originally created incident resource, and creates no second incident; replaying with the same `(actor_user_id, client_txn_id)` and a different normalized request fails with `409`.
- **AC-174**: Creating an incident with an `incident_key` whose trimmed Unicode-NFC-normalized form conflicts with an existing incident fails with `409` and leaves no partial incident, membership, or workbook-preference bootstrap state.
- **AC-175**: A caller with `is_deployment_admin=true` can `POST /api/v1/users` with `client_txn_id`, `auth_kind='local'`, `email`, `display_name`, and `initial_password`, receives `201 Created`, and can read the resulting safe user resource through both `GET /api/v1/users/{user_id}` and `GET /api/v1/users`; those route responses never include password hashes, TOTP secrets, WebAuthn credential material, opaque session tokens, or provider assertions.
- **AC-176**: Replaying `POST /api/v1/users` with the same `(actor_user_id, client_txn_id)` and the same normalized request returns `200 OK` and the originally created `user_id`; reusing that key with a different normalized request fails with `409`; an unauthenticated caller or an authenticated caller without `is_deployment_admin=true` cannot list, create, or patch users through the public route family; and the first deployment admin is not creatable through an unauthenticated public request.
- **AC-177**: `PATCH /api/v1/users/{user_id}` with a correct `base_user_version` can change only the mutable fields allowed by Core 01 §3.3.5.1; a stale `base_user_version` fails with `409` and `error.code = user_version_conflict`; deactivating a user revokes all active sessions immediately but leaves incident memberships intact; and demoting or deactivating the last active deployment admin fails with `409` and `error.code = last_deployment_admin`.
- **AC-178**: Any current incident member can `GET /api/v1/incidents/{incident_id}/memberships`; an incident admin can `POST /api/v1/incidents/{incident_id}/memberships` with exactly one of `user_id` or `email` and one valid role; creating a new membership returns `201 Created`; re-adding the same user with the same role returns `200 OK` and does not create a second membership row; re-adding the same user with a different role fails with `409` and `error.code = membership_exists_use_patch`; supplying a nonexistent user fails with `404` and `error.code = user_not_found`; supplying an inactive user fails with `409` and `error.code = user_inactive`; and the route never auto-creates or invites a user.
- **AC-179**: `PATCH /api/v1/incidents/{incident_id}/memberships/{user_id}` with a correct `base_membership_version` changes only `role`; a stale version fails with `409` and `error.code = membership_version_conflict`; requesting the current role returns `200 OK` without incrementing `membership_version`; and a successful role change takes effect on the next request because incident authorization is re-derived from current membership and role at request time.
- **AC-180**: `DELETE /api/v1/incidents/{incident_id}/memberships/{user_id}` removes only that incident membership and returns `204 No Content`; deleting or demoting the last incident admin fails with `409` and `error.code = last_incident_admin`; self-removal or self-demotion succeeds only when another current incident admin remains; and removing membership from incident A leaves the same still-valid session usable for incident B when the user remains authorized there.

- **AC-181**: `DELETE /api/v1/records/{record_id}` accepts JSON `base_row_version`, `client_txn_id`, and optional `reason`; an `editor`, `reviewer`, or `admin` on the incident receives `200 OK` with `record_id`, `incident_id`, incremented `row_version`, `deleted=true`, non-null `deleted_at`, non-null `deleted_by_user_id`, and `change_set_id`; the record no longer appears in ordinary view queries; `GET /api/v1/records/{record_id}/history` still returns prior history plus a `soft_delete` entry; the collaboration stream emits `record_changed` with `change_kind='remove'`; a stale base version fails with `409 error.code='row_version_conflict'`; patching the deleted record fails with `409 error.code='record_deleted_use_restore'`; and replaying the same normalized delete request by the same actor with the same `(record_id, client_txn_id)` returns the originally committed success without creating a second mutation.
- **AC-182**: `POST /api/v1/records/{record_id}/restore` accepts JSON `base_row_version`, `client_txn_id`, and optional `reason`; a `reviewer` or `admin` on the incident can restore a currently soft-deleted record and receives `200 OK` with `deleted=false`, `deleted_at=null`, `deleted_by_user_id=null`, incremented `row_version`, and a new `change_set_id`; the record becomes eligible again for ordinary view queries; history remains append-only and adds a `restore` entry rather than rewriting the prior delete; the collaboration stream emits `record_changed` with `change_kind='invalidate'` rather than a new insert-like change kind; a restore against a non-deleted record fails with `409 error.code='record_not_deleted'`; and if the implementation uses a short-lived destructive-operation lock and the record is already locked, the route fails with `409 error.code='record_locked'` and `error.retryable=true`.
- **AC-183**: For a soft-deleted record, `GET /api/v1/records/{record_id}/history` exposes the current tombstone `row_version`, or an equivalent current revision version token accepted by `POST /api/v1/records/{record_id}/restore`; a restore request that uses that current token succeeds when otherwise authorized, while a restore request that uses an older token fails with `409 error.code='row_version_conflict'`.
- **AC-184**: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` accepts `filters[]` entries using the Core 01 §3.3.4.1 wire shape for `eq`, `range`, `contains_any`, `contains_all`, `prefix`, and `full_text`; semantically identical client orderings normalize to the same `meta.query.filters[]`; and an unknown filter field, disallowed operator, duplicate `field_key`, empty `values[]`, empty text query, or malformed range fails with `400` and `error.code = invalid_view_query`.
- **AC-185**: The Notes view exposes `note.full_text` as a stable synthetic filter key; a query using `field_key='note.full_text'` and `op='full_text'` performs the declared case-insensitive token search over `note.title` and `note.body` without requiring storage-specific search syntax in the request.
- **AC-186**: `POST /api/v1/records/{survivor_record_id}/merge` accepts JSON `loser_record_id`, `survivor_base_row_version`, `loser_base_row_version`, `client_txn_id`, and optional `reason`; a `reviewer` or `admin` on the incident can merge two visible same-incident same-type active `host` records or two visible same-incident same-type active `identity` records; on success the route returns `200 OK` with `incident_id`, `record_type`, both record IDs, updated survivor and loser row versions, and `change_set_id`; the loser becomes historical with `merged_into_record_id` set to the survivor; active mentions, links, assessments, and tags are repointed or deterministically recreated in the same `change_set`; duplicate links and tags are deduplicated without losing history; and the collaboration stream removes the loser from ordinary active entity views while invalidating or patching the survivor and affected dependent rows.
- **AC-187**: The same route fails closed when the two IDs are identical, the records belong to different incidents, the record types differ, the record type is not `host` or `identity`, either record is soft-deleted or already merged away, the caller lacks `reviewer` or `admin` role, either supplied base row version is stale, or a short-lived destructive-operation lock is already held; stale versions fail with `409 error.code='row_version_conflict'`; an active lock fails with `409 error.code='record_locked'` and `error.retryable=true`; merge-precondition failures fail with `409 error.code='merge_precondition_failed'`; a caller lacking visibility receives `404`; and replaying the same normalized request by the same actor with the same `(survivor_record_id, loser_record_id, client_txn_id)` returns the originally committed success without creating a second merge.

### 9.11 Incident Portability Extension Profile criteria

- **AC-164**: Whole-incident export produces a bundle whose logical layout, manifest, checksum file, structured JSON or NDJSON files, and blob paths satisfy Core 01 §12.3, and the bundle excludes projections, search indexes, sessions, presigned URLs, locks, client-local drafts, login-capable local users, deployment-admin flags, auth-binding state, memberships, permissions, deployment-local administrative audit history, password hashes, MFA secrets, external-provider configuration, and object-store credentials.
- **AC-165**: Exporting an incident and importing that bundle into an empty deployment preserves the exported `incident_id`, `record_id`, `row_version`, change-set count, revision count, record-link count, entity-mention count, indicator-observation count, and blob hashes, and the imported incident opens normally after projection rebuild.
- **AC-166**: If one required structured file or one required blob is missing, if any required checksum is corrupted, if `incident_id` already exists, or if a required capability is unsupported, import fails closed before the incident becomes visible and leaves no partially visible incident state.
- **AC-167**: If a bundle contains optional embedded `snapshots` or `reference_packs` sections that the target deployment does not support, the importer ignores or degrades only those optional sections unless the relevant capability is listed in `required_capabilities[]`, and core incident import still succeeds.
- **AC-168**: Historical actors in `actors.ndjson` that do not map to an existing local user become inert imported actors or equivalent historical actor descriptors, are not login-capable, are not automatically added to incident membership, and still remain visible as historical attribution in imported history.
- **AC-169**: Whole-incident export and import run as background jobs, show progress and cancellation without blocking grid editing, stage bundle contents only under the configured temporary-work root, and in flyaway or disconnected deployments keep emitted bundles and staged extracts on encrypted storage.

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
