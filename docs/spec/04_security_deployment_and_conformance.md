# Cartulary Normative Core 04: Security, Deployment, and Conformance

## 1. Authentication model

### 1.1 Base authentication

**REQ-04-001**
The base profile MUST support:

- local user accounts stored in Postgres,
- password hashing with Argon2id,
- TOTP MFA,
- optional WebAuthn when the environment supports it.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-04-002**
The public API and WebSocket surface MUST use a server-managed session contract rather than a client-parsed identity token contract. Browser clients MUST authenticate with an `HttpOnly` `Secure` cookie carrying an opaque session token. The implementation MAY additionally accept `Authorization: Bearer <opaque_session_token>` for non-browser clients or trusted automation. The public token format MUST remain opaque and MUST NOT require clients to parse JWT claims or provider assertions.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-04-003**
State-changing HTTP requests authenticated by cookie MUST use CSRF protection that fails closed, such as a synchronizer token or an equivalent same-origin mechanism.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-04-004**
Authentication MUST work in disconnected deployments and MUST NOT depend on enterprise infrastructure for the base profile.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

#### 1.1.1 Session lifecycle boundaries

**REQ-04-005**
The base profile MUST create one server-side session record for each login-capable session.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-006**
The session credential presented by the browser or bearer client MUST identify that server-side session record and MUST be an opaque CSPRNG-generated bearer token with at least `128 bits` of unpredictability. Browser cookies carrying that token MUST be set with `HttpOnly`, `Secure`, `Path=/`, and `SameSite=Lax` or a stricter same-site policy. If bearer authentication is enabled for non-browser clients, it MUST represent the same opaque session family rather than a separate JWT or API-key contract.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-007**
The base profile MUST NOT require a separate long-lived browser refresh token.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-008**
Each session MUST persist `authenticated_at`, `last_qualifying_activity_at`, `idle_expires_at`, `absolute_expires_at`, and `session_expires_at`, where `session_expires_at` is the earlier of `idle_expires_at` and `absolute_expires_at`.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-009**
`idle_expires_at` MUST be computed as `last_qualifying_activity_at + 30 minutes`. `absolute_expires_at` MUST be computed as `authenticated_at + 12 hours`.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-010**
Qualifying activity MUST include successful authenticated user-initiated workbook or API activity. Qualifying activity MUST NOT include WebSocket `ping` or `pong`, passive server push, automatic reconnect or replay, or `GET /api/v1/auth/session`.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-011**
Qualifying activity MAY slide `idle_expires_at`, but MUST NOT extend the session beyond `absolute_expires_at`.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-012**
If a session expires, establishing a new session MUST require fresh primary authentication and any applicable MFA requirement.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-013**
The base profile MUST cap each human user at 5 concurrently active sessions. Multiple tabs sharing one browser session count as one session for this limit. Explicit system-process actors are outside this limit.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-014**
When a new login would exceed the concurrent-session cap, the server MUST revoke the least-recently-used non-current session before issuing the new session and MUST record an attributed audit event with stable reason code `concurrency_limit`.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-015**
`POST /api/v1/auth/logout` MUST revoke only the current session immediately.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-016**
Password change, MFA reset, account disablement, or an explicit deployment-admin revoke-all action MUST revoke all active sessions for that user immediately.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-017**
Loss of incident membership for one subscribed incident MUST terminate the affected incident WebSocket subscription and MUST cause future incident-scoped authorization checks for that incident to fail closed, but it MUST NOT by itself require revocation of otherwise valid account sessions for other authorized incidents.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

### 1.2 Enterprise Authentication Extension Profile

**REQ-04-018**
If the implementation claims the **Enterprise Authentication Extension Profile**, it MUST support provider-backed identities through an `auth_providers` and `auth_identities` model equivalent to the source artifact.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235

OIDC is the preferred enterprise path. SAML is the secondary enterprise path when required by the environment.

**REQ-04-019**
External identities MUST map to the same internal user identity used for attribution so that audit semantics remain unchanged.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235

**REQ-04-020**
When the Enterprise Authentication Extension Profile is implemented, successful provider authentication MUST terminate into the same server-managed session contract used by the base profile so the remaining API surface stays provider-agnostic.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235

## 2. Authorization model

**REQ-04-021**
The base profile MUST use incident-level roles equivalent to:

- `viewer`,
- `editor`,
- `reviewer`,
- `admin`.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-022**
Record access MUST inherit from incident access in the base profile.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-023**
API routes, preview or download handle issuance, job polling, and WebSocket incident subscriptions MUST re-derive authorization from the caller's current incident membership and role at request time.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-024**
`task_request`, `decision`, and coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MUST inherit the same incident-level authorization model. The base profile MUST NOT introduce record-specific ACLs or hidden sub-workspaces for these objects.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-025**
Saved-view scope controls only discoverability and mutability of the saved-view configuration object. It MUST NOT widen or narrow access to underlying incident rows, fields, search results, exports, or evidence.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-026**
Any incident member MAY create, update, or delete their own `private` saved views and set or clear their own `home_sheet_ref` even when their incident role is `viewer`, because those actions mutate personal workbook configuration rather than incident facts. Incident-wide default-surface updates and in-place mutation of another user's saved views MUST follow the scope and role rules defined by Core 03.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

Field-level ACLs, generalized approval workflows, and generalized record-level ACL systems are out of scope for the base profile.

**REQ-04-027**
If the Snapshot and Reporting Extension Profile is implemented, export redaction MUST NOT restrict live workbook views, search results, filters, saved views, row visibility, field visibility, or evidence visibility for authenticated incident participants. In the base profile, live workspace visibility is derived only from incident membership and the incident-level role model. In that profile, recipient-specific withholding MUST be implemented at snapshot, render, and release time rather than by hiding live workspace content.
Profiles: base, snapshot_reporting
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-233

**REQ-04-028**
The base profile MUST separate deployment-local account administration from incident-scoped data authorization. A conformant deployment MUST define one narrow deployment-scoped capability equivalent to `deployment_admin`. This capability authorizes only deployment-local user-account inspection and administration, including local-user creation, user patching, and any explicit revoke-all action the deployment exposes. The first holder of this capability MUST be provisioned out of band at install or bootstrap time.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-029**
Holding `deployment_admin` MUST NOT by itself grant incident read, write, preview, download, export, job, or WebSocket access. A caller who is `deployment_admin=true` but lacks current membership in incident X MUST have no incident-data access to incident X until granted ordinary incident membership.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-030**
Incident membership create, role-change, and delete routes remain incident-scoped authorization decisions. In the base profile, those routes MUST require current incident role `admin`; `deployment_admin` alone MUST NOT bypass that requirement.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

### 2.1 Snapshot and Reporting Extension Profile release gate

**REQ-04-031**
If the implementation claims the Snapshot and Reporting Extension Profile, it MUST provide a narrow artifact-scoped release gate for rendered outputs. This release gate MUST NOT become a generalized workflow engine for routine record editing or arbitrary record approvals.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233

**REQ-04-032**
Each release approval record MUST bind, at minimum, to:

- `snapshot_id`,
- `template_id`,
- `template_version`,
- `redaction_profile_id`,
- `redaction_profile_version`,
- `output_kind`,
- `release_scope`,
- `output_sha256`.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233

Approval requirements are:

- `internal_draft`: no approval required,
- `internal_review`: one `reviewer` approval,
- `external_release`: two distinct approvals, one from a `reviewer` attesting evidence sufficiency or claim support and one from an `admin` attesting release posture and redaction completeness.

**REQ-04-033**
Any change to the bound tuple or rendered bytes MUST invalidate prior approvals automatically.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233

**REQ-04-034**
For rendered-output lifecycle, the authoritative artifact state MUST be stored on the release record or an equivalent artifact-scoped record. The closed vocabulary for `release_state` is `pending_approval`, `approved`, `invalidated`, and `published`.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233

For this lifecycle, the logical output slot is the release tuple excluding `output_sha256`.

A release record enters `pending_approval` when bytes and `output_sha256` exist for one bound release tuple but the required approvals are not yet complete. It enters `approved` only when the approval requirements above are satisfied for that exact artifact. It enters `published` only through an explicit publish action after approval. It enters `invalidated` when a different artifact for the same logical output slot supersedes it, when its rendered bytes change, or when the implementation can no longer attest that the required approval set still applies to that exact artifact.

**REQ-04-035**
A new render with a different logical output slot or different `output_sha256` MUST start as `pending_approval`. It MUST NOT inherit `approved` or `published` state from an earlier artifact.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233

A narrow live sensitive-evidence model MAY be added in future work if repeated real-world incidents show that export-scoped withholding is insufficient. It is not a current conformance requirement.

## 3. Attribution and audit requirements

**REQ-04-036**
Every mutation MUST originate from an authenticated session or an explicitly identified system process.
Profiles: base
Verified by: AC-231

**REQ-04-037**
Every mutation MUST record, at minimum:

- actor user identifier,
- timestamp,
- mutation source such as `ui`, `import`, or `rollback`,
- before/after values or equivalent patch data at the required history granularity.
Profiles: base
Verified by: AC-231

**REQ-04-038**
User-account and incident-membership administration mutations MUST be captured with the same minimum actor, timestamp, source, and before/after fidelity even when they are stored outside incident `change_set` or `record_revisions` rows. Those administrative audit records are deployment-local state and MUST NOT be serialized into whole-incident portability bundles.
Profiles: base, incident_portability
Verified by: AC-231, AC-236

**REQ-04-039**
Security choices MUST NOT add friction to the primary capture path. MFA belongs at login and session establishment, not during routine row creation or evidence preview.
Profiles: base
Verified by: AC-231

## 4. Trust boundaries

### 4.1 Reference packs

**REQ-04-040**
Optional enrichment credentials MUST live in server-side configuration or secret storage. They MUST NOT live in incident records, client-side storage, or imported pack files.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-052, AC-092, AC-093, AC-094, AC-095, AC-096, AC-234

**REQ-04-041**
Reference packs MUST record structured metadata before activation. Queryable metadata MUST include, at minimum, `pack_key`, `pack_kind`, `pack_version`, source identifier if available, `manifest_sha256`, one or more payload SHA-256 digests in deterministic member order or an equivalent canonical aggregate digest, `verification_method`, signer-key or trusted-source identifier, imported and activated actor attribution with timestamps, `previous_active_version`, and `verification_result`.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-052, AC-092, AC-093, AC-094, AC-095, AC-096, AC-234

**REQ-04-042**
In a flyaway or disconnected deployment, reference-pack import and activation MUST operate only on locally supplied bundles rooted under the configured reference-pack storage path or an equivalent administrative upload path that writes into that root. The running application MUST NOT require a live network fetch to verify or activate a pack.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-052, AC-092, AC-093, AC-094, AC-095, AC-096, AC-234

**REQ-04-043**
Imported pack bundles and extracted contents MUST be treated as hostile content until verification and content screening succeed.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-052, AC-092, AC-093, AC-094, AC-095, AC-096, AC-234

### 4.2 Export outputs

**REQ-04-044**
Generated reports and presentations MUST embed or package required assets locally rather than pulling them from remote CDNs or runtime asset services.
Profiles: snapshot_reporting, incident_portability
Verified by: AC-031, AC-057, AC-059, AC-060, AC-061, AC-062, AC-091, AC-113, AC-114, AC-115, AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-233, AC-236

**REQ-04-045**
If the Snapshot and Reporting Extension Profile is implemented:

- `external_release` outputs MUST exclude raw blob bytes and `working_material`,
- any `curated_narrative` block included in `external_release` MUST carry `support_refs[]`,
- content drawn directly from `task_request`, `decision`, `comm_log`, `handoff`, `status_review`, or `lesson` records MUST NOT be treated as inherently releasable; any `external_release` use MUST flow through the snapshot, redaction, and curation path,
- reenactment outputs MUST be marked `generated_presentation=true` and MUST NOT be released as `external_release`,
- approval and redaction checks MUST complete successfully before an `external_release` artifact is published.
Profiles: snapshot_reporting, incident_portability
Verified by: AC-031, AC-057, AC-059, AC-060, AC-061, AC-062, AC-091, AC-113, AC-114, AC-115, AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-233, AC-236

**REQ-04-046**
In that profile, the implementation MUST support generating multiple recipient-specific artifacts from the same immutable snapshot by selecting different versioned redaction profiles and, when needed, different templates. If an incident involves multiple affected parties, an artifact prepared with one recipient-specific configuration MUST NOT disclose content whose `disclosure_partition_refs[]` are not allowed by the selected redaction profile. Manual post-render editing MAY still occur, but it MUST NOT be required for the implementation's supported recipient-specific configurations.
Profiles: snapshot_reporting, incident_portability
Verified by: AC-031, AC-057, AC-059, AC-060, AC-061, AC-062, AC-091, AC-113, AC-114, AC-115, AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-233, AC-236

**REQ-04-047**
If the Incident Portability Extension Profile is implemented:

- whole-incident portability bundles MUST serialize only authoritative incident source state, deterministic structured files, and referenced blob bytes, not projections or deployment-local runtime state,
- bundle import MUST stage content under the configured temporary-work root and verify required checksums before any structured incident data becomes visible,
- unsupported or missing optional embedded snapshot or reference-pack sections MUST NOT block import of the core incident state,
- portability bundles, staged extracts, and emitted artifacts for flyaway or disconnected use MUST remain on encrypted storage roots.
Profiles: snapshot_reporting, incident_portability
Verified by: AC-031, AC-057, AC-059, AC-060, AC-061, AC-062, AC-091, AC-113, AC-114, AC-115, AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-233, AC-236

### 4.3 Evidence uploads

**REQ-04-048**
The deployment MAY use an upload-malware-scanning sidecar or equivalent adjunct service. Such a service is optional in the current core and MUST NOT break the two-step attachment semantics.
Profiles: base
Verified by: AC-053, AC-128, AC-231

### 4.4 STRIDE threat model

**REQ-04-049**
The implementation MUST maintain a project-local STRIDE threat model covering the current architecture, deployment profiles, and high-risk workflows.
Profiles: base
Verified by: AC-048, AC-231

**REQ-04-050**
The threat model MUST be updated before any release that adds or materially changes:

- an import path,
- an export or report surface,
- an evidence preview or rendering path,
- an external fetch capability,
- a credential-bearing integration,
- a deployment profile,
- an object-storage access pattern.
Profiles: base
Verified by: AC-048, AC-231

**REQ-04-051**
At minimum, the threat model MUST cover the following assets and abuse cases:
Profiles: base
Verified by: AC-048, AC-231

| STRIDE class | Minimum project-specific scope | Required control direction |
| --- | --- | --- |
| Spoofing | analyst sessions, provider-backed identities, explicit system-process actors, object-store upload/download capabilities | authenticated sessions, stable internal user mapping, explicit system actors, short-lived operation-scoped object access |
| Tampering | incident records, revisions, object blobs, reference packs, snapshots, exports | row-versioned writes, immutable change sets, blob hashes, fail-closed integrity verification, immutable published snapshots |
| Repudiation | edits, imports, rollbacks, pack activation, export generation, evidence lifecycle actions | attributed append-only history with actor, timestamp, source, and reversible mutation detail |
| Information disclosure | evidence blobs, exports, previews, secrets, portable runtime roots | incident-scoped authorization, secret isolation, untrusted-content rendering rules, self-contained outputs, encrypted flyaway storage |
| Denial of service | oversized evidence, archive bombs, pathological imports, expensive report or preview jobs | size and decompression limits, background-job isolation, cancellation, bounded hot-path retrieval |
| Elevation of privilege | user-controlled record or blob identifiers, destructive operations, job-worker storage access | server-side authorization derived from object ownership, role gates for destructive actions, least-privilege worker credentials |

### 4.5 Focused MITRE CWE constraints

**REQ-04-052**
The implementation MUST address the following MITRE CWE entries during architecture review, code review, and conformance testing. This list is intentionally narrow and project-specific.
Profiles: base
Verified by: AC-049, AC-050, AC-051, AC-052, AC-053, AC-054, AC-055, AC-130, AC-131, AC-231

**REQ-04-053**
- **CWE-79**: Incident-authored or imported content rendered in the browser UI or exported HTML MUST be treated as untrusted. Renderers MUST escape or sanitize by default and MUST block script execution, inline event handlers, `javascript:` URLs, and remote asset fetches sourced from incident data.
- **CWE-1236**: CSV, XLSX, and clipboard exports intended for spreadsheet consumption MUST neutralize leading formula characters before write. At minimum, values beginning with `=`, `+`, `-`, `@`, tab, or carriage return MUST be emitted with a lossless neutralizing prefix such as `'`, unless an explicit raw-forensic export mode is selected with a visible danger warning.
- **CWE-22 / CWE-73**: User-supplied filenames, archive entry names, and import paths MUST be treated as metadata, not authority. The system MUST assign storage keys, MUST reject absolute paths and parent traversal, and MUST extract archives only inside a staging root that cannot escape the declared runtime roots.
- **CWE-353**: Reference packs and any incident import bundle format, when implemented, MUST fail closed on checksum mismatch, signature mismatch, incomplete download, or missing required integrity metadata.
- **CWE-434**: Evidence, reference-pack, and workbook-import uploads MUST be treated as hostile content. The application unit MUST NOT execute uploaded content, workbook formulas, macros or VBA, workbook automation, or external links during import or preview, and preview generation MUST use only allowlisted non-executing transforms. Active-content types MUST remain download-only or isolated from the main application origin unless a dedicated isolated analysis path is explicitly implemented.
- **CWE-639**: Every mutation, preview, download, and object-store URL issuance MUST re-derive authorization server-side from the target object's owning incident and the caller's current membership and role. Client-supplied incident identifiers, ownership metadata, or role claims MUST NOT determine access.
- **CWE-352**: State-changing HTTP routes authenticated by cookie MUST require CSRF protection that fails closed. WebSocket upgrades and any incident subscription step MUST verify the authenticated session and incident authorization before joining an incident-scoped stream. Cookie-authenticated browser WebSocket connections MUST reject untrusted `Origin` values before the socket joins an incident-scoped stream.
- **CWE-312**: Deployments intended for portable or flyaway use MUST keep database storage, object storage, reference-pack storage, temporary work files that carry incident data, and export outputs on encrypted storage. Unencrypted removable media or unencrypted portable roots are non-conformant for flyaway handling.
Profiles: base, import, reference_pack
Verified by: AC-049, AC-050, AC-051, AC-052, AC-053, AC-054, AC-055, AC-130, AC-131, AC-231, AC-232, AC-234

## 5. Deployment profiles

### 5.1 Flyaway or disconnected deployment

**REQ-04-054**
The recommended disconnected deployment MUST consist of:

- one application container,
- one Postgres container,
- one MinIO container or equivalent S3-compatible object store.
Profiles: base
Verified by: AC-055, AC-092, AC-096, AC-169, AC-231

Docker Compose or Podman Compose with mounted volumes is acceptable.

**REQ-04-055**
If the deployment claims the Reference Pack Extension Profile, the smallest supported disconnected bundle MUST preinstall only the three reference packs defined in Core 01 §11.2. Framework, enrichment, template, and separately distributed view-contract packs MUST remain separately installable offline bundles in that minimum deployment.
Profiles: base, reference_pack
Verified by: AC-055, AC-092, AC-096, AC-169, AC-231, AC-234

### 5.2 On-prem deployment

On-prem deployments MAY replace the Postgres or object-store containers with centrally managed services if equivalent semantics are preserved.

**REQ-04-056**
The application deployable MUST remain a single application unit.
Profiles: base
Verified by: AC-231

### 5.3 Cloud deployment

Cloud deployments MAY run the application on ECS, Kubernetes, VMs, or equivalent platforms, with managed Postgres and native object storage.

**REQ-04-057**
The logical architecture and data contracts MUST remain unchanged.
Profiles: base
Verified by: AC-231

## 6. Runtime roots and storage paths

**REQ-04-058**
The deployment configuration MUST declare explicit persistent roots for:

- database storage,
- object storage,
- reference-pack storage,
- temporary work files,
- export outputs.
Profiles: base, reference_pack
Verified by: AC-051, AC-055, AC-169, AC-231, AC-234

**REQ-04-059**
The application MUST NOT rely on source-tree-relative paths for runtime assets or generated artifacts.
Profiles: base
Verified by: AC-051, AC-055, AC-169, AC-231

## 7. Container boundary

**REQ-04-060**
The following components MUST remain in one deployable application unit:

- web UI,
- API,
- WebSocket hub,
- background jobs.
Profiles: base
Verified by: AC-231

Postgres and object storage SHOULD remain separate services.

## 8. Required and optional supporting services

### 8.1 Required services

**REQ-04-061**
A conformant deployment MUST provide:

- Postgres,
- object storage,
- the application deployable.
Profiles: base
Verified by: AC-231

### 8.2 Optional services

A conformant deployment MAY additionally provide:

- external reverse proxy or TLS termination,
- enterprise IdP,
- evidence scanning sidecar,
- managed storage substitutes preserving the same contracts.

## 9. Conformance criteria

Each criterion below is a pass/fail requirement.

**REQ-04-062**
Within this document, subsection numbers MUST be unique and strictly increasing, acceptance-criterion identifiers MUST be unique, and a retired or superseded acceptance-criterion identifier MUST NOT be reassigned or recycled.
Profiles: base
Verified by: AC-231, AC-237

**REQ-04-063**
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
Profiles: base
Verified by: AC-231, AC-237

Unless a criterion states otherwise, `reference incident` in this section means Fixture A.

**REQ-04-064**
Latency measurements in this section MUST use end-user-observable completion time from the initiating user action to the required visible UI state. Any criterion expressed as p95 MUST be evaluated over at least 100 completed operations of the named interaction after one warm-up pass on the named fixture.
Profiles: base
Verified by: AC-231, AC-237

For this section:

- `first useful viewport` means the first rendered visible row window for the active sort, filter, and grouping state with stable `record_id` binding and working keyboard navigation, even if off-screen rows continue loading.
- `stable viewport` means the visible row window and result ordering match the final deterministic order for the active sort, filter, and grouping state and no further reorder occurs without new user or server input.
- `metadata shell` means row fields and evidence metadata needed to inspect the selected record, including counts, filenames or media-type labels, attachment state, and preview handles, but excluding binary preview bytes or full blob download.


### 9.0 Profile claim manifests and traceability

The manifests below define claim boundaries without restating requirement prose. Each manifest selects requirements through the `Profiles:` trailer defined by Core 00 §5.2 and pairs that selector with the acceptance criteria that complete the claim. Appendix F expands every selector into explicit navigation tables.

#### 9.0.1 Base claim manifest

A Base claim selects every requirement block tagged `base`.

Definition of Done:

- requirement selector: `profile:base`
- required acceptance criteria: `AC-001..AC-026`, `AC-037..AC-055`, `AC-097..AC-103`, `AC-107..AC-112`, `AC-116..AC-163`, `AC-170..AC-231`, `AC-237..AC-243`
- **AC-231**: A Base claim is conformant only when every requirement selected by `profile:base` is implemented and every acceptance criterion listed in this manifest passes.
  - Verifies: `profile:base`

#### 9.0.2 Import claim manifest

An Import claim requires a passing Base claim and every requirement block tagged `import`, including the import-boundary and import-provenance requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:import`
- additional acceptance criteria: `AC-027..AC-029`, `AC-063..AC-067`, `AC-232`, `AC-237`
- **AC-232**: An Import claim is conformant only when a Base claim passes, every requirement selected by `profile:import` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:import`

#### 9.0.3 Snapshot and Reporting claim manifest

A Snapshot and Reporting claim requires a passing Base claim and every requirement block tagged `snapshot_reporting`, including the release-gate requirements in §2.1 and any snapshot or rendering requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:snapshot_reporting`
- additional acceptance criteria: `AC-030..AC-032`, `AC-056..AC-062`, `AC-091`, `AC-104..AC-106`, `AC-113..AC-115`, `AC-233`, `AC-237`
- **AC-233**: A Snapshot and Reporting claim is conformant only when a Base claim passes, every requirement selected by `profile:snapshot_reporting` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:snapshot_reporting`

#### 9.0.4 Reference Pack claim manifest

A Reference Pack claim requires a passing Base claim and every requirement block tagged `reference_pack`, including the disconnected-pack lifecycle and attestation requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:reference_pack`
- additional acceptance criteria: `AC-033..AC-035`, `AC-092..AC-096`, `AC-234`, `AC-237`
- **AC-234**: A Reference Pack claim is conformant only when a Base claim passes, every requirement selected by `profile:reference_pack` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:reference_pack`

#### 9.0.5 Enterprise Authentication claim manifest

An Enterprise Authentication claim requires a passing Base claim and every requirement block tagged `enterprise_authentication`, including the provider-identity requirements in §1.2.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:enterprise_authentication`
- additional acceptance criteria: `AC-036`, `AC-235`, `AC-237`
- **AC-235**: An Enterprise Authentication claim is conformant only when a Base claim passes, every requirement selected by `profile:enterprise_authentication` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:enterprise_authentication`

#### 9.0.6 Incident Portability claim manifest

An Incident Portability claim requires a passing Base claim and every requirement block tagged `incident_portability`, including the logical-bundle and import-failure requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:incident_portability`
- additional acceptance criteria: `AC-164..AC-169`, `AC-236`, `AC-237`
- **AC-236**: An Incident Portability claim is conformant only when a Base claim passes, every requirement selected by `profile:incident_portability` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:incident_portability`

#### 9.0.7 Traceability integrity criterion

- **AC-237**: In Core 00 through Core 04, every atomic conformance-critical requirement block carries one `REQ-*` identifier, one `Profiles:` trailer, and one `Verified by:` trailer; every acceptance criterion carries one `Verifies:` back-reference; profile-manifest criteria use only the canonical `profile:*` selectors defined by Core 00 §5.2; and Appendix F expands every selector into deterministic `REQ -> owner section / profiles / ACs`, `AC -> REQs`, and `Profile -> required REQs + ACs` tables sorted canonically by identifier.
  - Verifies: REQ-00-009..REQ-00-012, REQ-04-062..REQ-04-064

### 9.1 Base Profile criteria

- **AC-001**: An analyst can create a new timeline row by typing into a blank grid cell and pressing Enter, with no modal and no required form, and the row is saved in under 150 ms on LAN for a single-row edit.
  - Verifies: REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-111..REQ-03-115
- **AC-002**: A timeline row can be persisted with only one non-empty user-entered value or only an attached screenshot; `recorded_at`, author, and row identity are system-generated.
  - Verifies: REQ-03-001..REQ-03-003, REQ-03-111..REQ-03-115
- **AC-003**: Pasting a 20-row by 5-column block from Excel into the timeline sheet creates rows and maps visible columns in under 2 seconds on a reference incident.
  - Verifies: REQ-01-015..REQ-01-017, REQ-03-145..REQ-03-152, REQ-03-221..REQ-03-222
- **AC-004**: Pasting an image from the clipboard or dragging a screenshot onto a selected row attaches evidence in no more than two user actions and does not navigate away from the grid.
  - Verifies: REQ-01-015..REQ-01-017, REQ-03-116..REQ-03-119
- **AC-005**: Arrow keys, Tab, Enter, Shift+Enter, and Ctrl+V work in the grid without opening side dialogs or breaking selection state.
  - Verifies: REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-217..REQ-03-220, REQ-03-263
- **AC-006**: An analyst can resolve an unresolved host or account mention from the inspector and return focus to the original grid cell without losing scroll position or selection.
  - Verifies: REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-247..REQ-03-249
- **AC-007**: The selected row’s edit history is visible in one click or one shortcut and includes actor, timestamp, operation, changed field, link, mention, tag, or evidence entry, plus rollback actions.
  - Verifies: REQ-03-138..REQ-03-140, REQ-03-261..REQ-03-262
- **AC-008**: Two analysts on the same sheet can see each other’s presence within 1 second, including row-level presence and same-cell editing indicators when applicable.
  - Verifies: REQ-03-090..REQ-03-091
- **AC-009**: Concurrent edits to different fields on the same row auto-merge; concurrent edits to the same field never silently overwrite without a visible conflict.
  - Verifies: REQ-03-033..REQ-03-040
- **AC-010**: A reviewer can roll back one mistaken host link, tag assignment, mention resolution, or evidence association from the selected row’s history without reverting later unrelated edits on the same row.
  - Verifies: REQ-03-141..REQ-03-142
- **AC-011**: Rolling back a mistaken change or restoring a whole row creates a new attributed revision and updates the visible row in under 2 seconds on a reference incident.
  - Verifies: REQ-03-141..REQ-03-142
- **AC-012**: Whole-row restore and whole-change-set rollback are available as explicit secondary actions for multi-target or destructive changes; arbitrary field-picker rollback from historical snapshots is not required in the base profile.
  - Verifies: REQ-03-141..REQ-03-142
- **AC-013**: Re-sorting, re-filtering, or re-grouping a sheet does not cause a pending edit to target a different underlying record; all mutations are sent using `record_id`, `base_row_version`, and changed fields only.
  - Verifies: REQ-03-033..REQ-03-035, REQ-03-086, REQ-03-223..REQ-03-224
- **AC-014**: Renaming a visible column header or tab label does not change filter semantics, write-back behavior, or export semantics for a built-in or system view; those behaviors are bound to `view_schema_id`.
  - Verifies: REQ-03-223..REQ-03-224
- **AC-116**: With no optional reference packs installed or activated, the required built-in and system surfaces still resolve to exactly these nine pack-independent `view_schema_id` values: `cartulary.view.timeline.v1`, `cartulary.view.hosts.v1`, `cartulary.view.identities.v1`, `cartulary.view.evidence.v1`, `cartulary.view.notes.v1`, `cartulary.view.indicators.v1`, `cartulary.view.assessments.v1`, `cartulary.view.task_requests.v1`, and `cartulary.view.decisions.v1`.
  - Verifies: REQ-00-014, REQ-01-285..REQ-01-295, REQ-01-307..REQ-01-311, REQ-03-004, REQ-03-242..REQ-03-246
- **AC-117**: The structured contract for each of those nine base-profile view schemas includes hidden `record_id` and `row_version`, an ordered field list, an ordered default sort tuple, and the declared filter whitelist; the final default-sort tiebreaker is `record_id`.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-03-242..REQ-03-246
- **AC-118**: Every writable field in those nine base-profile view schemas declares a stable `field_key` and `conflict_resolution_class`, and every entity-bearing writable field also declares `entity_binding_mode`.
  - Verifies: REQ-00-014, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-01-323..REQ-01-341,
    REQ-02-028..REQ-02-029, REQ-02-202..REQ-02-204, REQ-03-052..REQ-03-053, REQ-03-242..REQ-03-246
- **AC-119**: In the Timeline schema, `timeline.summary`, `timeline.details`, and `timeline.source_text` are distinct field keys with distinct write targets. Editing one does not overwrite either of the other two fields.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-322, REQ-03-236..REQ-03-241
- **AC-120**: A write routed to a field that the active base-profile view schema declares read-only or derived fails closed, does not mutate authoritative source state, and does not persist a misleading projection update.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-03-236..REQ-03-241

- **AC-191**: A Timeline row created by blank-row entry, by a paste action that creates a new row, or by a screenshot-only create persists `capture_state='rough'` on first commit.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241
- **AC-192**: A client attempt to supply `timeline.capture_state` in Timeline row creation or in `PATCH /api/v1/records/{record_id}` fails closed under the ordinary read-only-field rule, does not mutate authoritative source state, and does not persist a misleading projection update.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110, REQ-03-236..REQ-03-241
- **AC-193**: The first later `capture-state-material` mutation to a `rough` Timeline row, including an edit to `timeline.occurred_at`, `timeline.summary`, `timeline.details`, or `timeline.source_text`, or the first host-ref, identity-ref, evidence-attach or evidence-detach, or row-anchored indicator mutation, commits one visible `change_set` that also sets `capture_state='enriched'`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241
- **AC-194**: `POST /api/v1/records/{record_id}/mark-reviewed` by a `reviewer` or `admin` against a non-deleted Timeline row whose current `capture_state` is `rough` or `enriched` returns `200 OK`, increments `row_version`, appends a new `change_set`, and sets `capture_state='reviewed'`; the same route by an `editor` fails with `403`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-195**: A later `capture-state-material` mutation to a `reviewed` Timeline row commits one visible `change_set` that sets `capture_state='enriched'`; a tag-only change to the same row leaves `capture_state='reviewed'`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-196**: `POST /api/v1/records/{record_id}/supersede` by a `reviewer` or `admin` with a non-empty reason against a non-deleted Timeline row whose current `capture_state` is `rough`, `enriched`, or `reviewed` returns `200 OK`, increments `row_version`, appends a new `change_set`, and sets `capture_state='superseded'`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-197**: After a Timeline row reaches `superseded`, ordinary grid edits, `PATCH /api/v1/records/{record_id}` field mutations, and `POST /api/v1/records/{record_id}/mark-reviewed` fail closed; leaving `superseded` requires rollback of the superseding change.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-198**: Removing all current links or evidence from an already `enriched` Timeline row does not downgrade it to `rough`, and `timeline.has_unresolved_mentions` is `true` if and only if the current row still has at least one non-deleted unresolved mention, regardless of `capture_state`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-199**: Replaying the same normalized `POST /api/v1/records/{record_id}/mark-reviewed` or `POST /api/v1/records/{record_id}/supersede` request by the same authenticated actor with the same `(record_id, client_txn_id)` returns the originally committed success and does not create a second lifecycle transition.
  - Verifies: REQ-03-102..REQ-03-110
- **AC-121**: The Assessments view accepts band-first creation using the canonical `NULL`, `25`, `55`, and `85` `confidence_score` mapping and rejects in-place mutation of an existing row's `subject_ref`, `subject_type`, `assessment_state`, `confidence_score`, `rationale`, `assessor`, or `assessed_at`.
  - Verifies: REQ-01-296..REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093,
    REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011, REQ-03-250..REQ-03-254
- **AC-122**: The Indicators view allows inline creation of a new canonical indicator row. For `cartulary.view.indicators.v1`, an existing indicator row exposes no writable fields. Grid-edit and API patch attempts against an existing row fail closed for every create-only field. The exact identity-defining immutable field set for that schema is always `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, and `indicator.normalized_value`, plus `indicator.hash_algorithm` and `indicator.hash_value` when populated and used by the canonical dedupe key. `indicator.stix_pattern` and `indicator.defanged_value` remain rejected under the create-only rule without being treated as identity-defining.
  - Verifies: REQ-01-296..REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-331, REQ-02-056..REQ-02-057,
    REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011
- **AC-146**: Analyst A can create a `private` saved view over one `view_schema_id`; `GET /api/v1/incidents/{incident_id}/saved-views` for analyst A returns it, the same call for analyst B on the same incident does not, and the same call for an incident admin does.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021
- **AC-147**: A `shared` saved view created in one incident is returned to all incident members; a non-owner non-admin member can open and duplicate it, but an in-place patch or delete by that member fails closed.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021
- **AC-148**: A visible `system` saved view cannot be created, patched, or deleted through the ordinary saved-view routes, but any incident member can duplicate it into a new saved view allowed by policy.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021
- **AC-149**: Changing saved-view scope, updating saved-view `query_json` or `layout_json`, or deleting a saved view never changes underlying row visibility, field visibility, evidence visibility, search results, or export-redaction behavior for incident participants whose incident membership is unchanged.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021, REQ-04-021..REQ-04-030
- **AC-150**: Workbook open resolves the starting surface in this order: explicit `sheet_ref`, caller `home_sheet_ref`, incident `default_sheet_ref`, then `cartulary.view.timeline.v1`; if a referenced saved view is deleted, hidden by scope, or invalid because a required optional pack is unavailable, the invalid pointer is cleared and the implementation falls through deterministically to the next step instead of failing workbook open.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032
- **AC-112**: An analyst can create a note directly in the Notes sheet by blank-row or equivalent grid-native entry with no required modal, and the resulting record is stored as an artifact with `artifact_type='note'`.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071, REQ-03-004,
    REQ-03-242..REQ-03-246
- **AC-068**: From a selected Timeline, Host, Identity, or Evidence record, an analyst can invoke `add linked note` without leaving the workbook flow; the resulting note is stored with the same artifact record shape as a Notes-sheet-created note, appears in the Notes sheet, and is visible as a linked record from the source row.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071
- **AC-069**: Renaming the visible Notes tab label, and any implementation-supported per-user hide/show operation for that built-in tab, does not change write-back or export semantics because behavior remains bound to `view_schema_id`.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330
- **AC-070**: Editing a lightweight free-text field on a timeline, host, identity, or evidence record does not create a standalone Notes row. Creating a standalone note does create a distinct note record that can be linked, tagged, and reviewed in history independently.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071
- **AC-015**: An analyst can create an evidence request record without uploading a blob, later attach or replace the blob, and preserve structured `requested_at`, `received_at`, `storage_ref`, `blob_hash`, `collector_party_text`, `source_party_text`, and append-only custody history across the lifecycle.
  - Verifies: REQ-01-243..REQ-01-247, REQ-01-291..REQ-01-295, REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-201,
    REQ-03-120..REQ-03-126, REQ-03-242..REQ-03-246
- **AC-016**: Evidence processing and any implemented background job start without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
  - Verifies: REQ-01-243..REQ-01-247, REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-201, REQ-03-121..REQ-03-126
- **AC-043**: On Fixture A, selection change, focus change, and typing acknowledgment each remain at or below 100 ms p95 on LAN, and blank-row creation in the timeline sheet with one non-empty user-entered value remains at or below 150 ms p95 on LAN.
  - Verifies: REQ-00-013, REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-087..REQ-03-089,
    REQ-03-217..REQ-03-219, REQ-03-263
- **AC-044**: On Fixture A, sort, filter, and grouping changes present a first useful viewport at or below 250 ms p95 and a stable viewport at or below 1.0 s p95.
  - Verifies: REQ-00-013, REQ-01-015..REQ-01-017, REQ-03-223..REQ-03-224
- **AC-045**: On Fixture B, opening the inspector for a timeline row linked to 100 evidence records shows a metadata shell at or below 300 ms p95; binary preview bytes MAY continue progressively after the inspector opens.
  - Verifies: REQ-00-013, REQ-01-015..REQ-01-017, REQ-01-355..REQ-01-366, REQ-03-242..REQ-03-246
- **AC-046**: On either Fixture A or Fixture B, imports, evidence processing, projection rebuilds, snapshot generation, report generation, and reference-pack refresh remain background jobs, show progress and cancellation within 1 second of job start, and do not block grid editing or row creation.
  - Verifies: REQ-00-013, REQ-01-004..REQ-01-014, REQ-01-018, REQ-01-248..REQ-01-249, REQ-01-342..REQ-01-348,
    REQ-01-351..REQ-01-353, REQ-01-369, REQ-01-452..REQ-01-454
- **AC-047**: On Fixture A, scrolling, sorting, filtering, grouping, and live updates never retarget pending edits away from the selected `record_id`, and viewport stabilization does not require a full-sheet rerender.
  - Verifies: REQ-01-015..REQ-01-017, REQ-03-033..REQ-03-035, REQ-03-086, REQ-03-223..REQ-03-224,
    REQ-03-233..REQ-03-235
- **AC-017**: Indicator data entered through the Indicators system view, imported through supported ingest paths, or resolved from supported source fields appears under a stable indicator system-view contract and, when export surfaces exist, a stable export contract with consistent indicator type, value kind, canonical value, normalization, deterministic dedupe key, and optional STIX-mapping fields.
  - Verifies: REQ-01-331, REQ-01-355..REQ-01-366, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082,
    REQ-02-202..REQ-02-204, REQ-03-135..REQ-03-137
- **AC-018**: Recording a new `unknown`, `suspected`, `confirmed`, `disproven`, or `cleared` assessment for a host or identity appends a new attributed assessment record; prior assessments remain visible in history and are not overwritten.
  - Verifies: REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-202..REQ-02-204, REQ-03-250..REQ-03-254
- **AC-080**: A host or identity can receive `unknown -> suspected -> confirmed -> cleared` as four separate assessment records for the same incident-scoped subject, and the reviewer can still see each earlier record in order.
  - Verifies: REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254
- **AC-081**: A host or identity can receive `unknown -> disproven` without implying prior compromise, and filtering for `assessment_state='cleared'` excludes `disproven`.
  - Verifies: REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254
- **AC-082**: Recording an operational response action such as device isolation, account disablement, credential reset, or monitoring does not mutate `assessment_state` unless a new explicit assessment record is appended.
  - Verifies: REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254
- **AC-083**: Interactive compromise-assessment entry surfaces expose confidence by default as `unset`, `low`, `medium`, or `high`; choosing those values persists `confidence_score` as `NULL`, `25`, `55`, or `85` respectively, and any supported exact-score write path accepts integers from `0` through `100`.
  - Verifies: REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254
- **AC-084**: The Compromise Assessments system view supports separate filters on `assessment_state` and derived `confidence_band`; `confidence_score=NULL` is rendered and filterable as `confidence_band='unset'` and is not treated as `0`.
  - Verifies: REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254
- **AC-085**: From a selected Notes, Timeline, Host, Identity, or Evidence context, an analyst can create a first-class `task_request` with required `task_kind`, `status`, `owner_user_id`, `priority`, `created_at`, and `updated_at`; optional structured fields including `workstream`, `requester_party_text`, `due_at`, `blocked_reason`, `completed_at`, and `external_ticket_ref` remain editable and filterable; a blocked task requires `blocked_reason`; a done task requires `completed_at`; and workbook filtering can show open tasks by owner, blocked state, due status, workstream, and external ticket reference.
  - Verifies: REQ-01-296..REQ-01-302, REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-120..REQ-02-122,
    REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260
- **AC-086**: Creating a `decision` record preserves `decision_type`, `status`, `owner_user_id`, `decided_at`, and `rationale`; reviewer history can reconstruct when the decision changed state, who owned it, and what `support_refs[]` were attached.
  - Verifies: REQ-01-296..REQ-01-302, REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-122,
    REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260
- **AC-087**: A `comm_log` artifact can record required `comm_type`, `timestamp_utc`, `audience`, `channel_or_meeting`, and `summary`, while a `status_review` artifact can record required timestamp and summary plus linked decision or task references, and a workbook surface can filter or sort those artifacts by `comm_type`, `audience`, timestamp, and next-report or next-checkpoint time without rereading unrelated note text.
  - Verifies: REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-123..REQ-02-133, REQ-03-005..REQ-03-011,
    REQ-03-255..REQ-03-260
- **AC-088**: A `handoff` artifact can capture outgoing owner, incoming owner, current state summary, open task IDs, and open decision IDs, and the incoming analyst can pivot directly from the handoff to the referenced open work from the same incident.
  - Verifies: REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-123..REQ-02-133, REQ-03-005..REQ-03-011,
    REQ-03-255..REQ-03-260
- **AC-089**: Hypothesis tracking remains artifact-backed in the current profile, using either `artifact_type='hypothesis'` or another declared structured artifact subtype; base conformance does not require or claim a first-class `hypothesis` `record_type`.
  - Verifies: REQ-01-296..REQ-01-302, REQ-02-067..REQ-02-071, REQ-02-094, REQ-02-123..REQ-02-134,
    REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260
- **AC-090**: Creating or editing timeline rows does not require owner, approver, challenge, checklist, task, or decision fields on the timeline sheet, and ordinary row edits do not enter a generalized approval workflow.
  - Verifies: REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-120..REQ-02-122, REQ-03-005..REQ-03-011,
    REQ-03-255..REQ-03-260, REQ-03-264
- **AC-019**: Typing or pasting `WS-023?` into a Timeline Hosts cell creates an `entity_mention` and zero host records unless the analyst explicitly resolves or creates an entity.
  - Verifies: REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-027, REQ-02-030..REQ-02-036, REQ-03-129..REQ-03-134
- **AC-020**: Creating a host or identity from a selected unresolved mention creates exactly one stub entity, preserves the raw mention, resolves only the selected mention by default, and stores the seed value in alias or provenance data.
  - Verifies: REQ-01-196..REQ-01-227, REQ-02-030..REQ-02-038, REQ-03-129..REQ-03-134, REQ-03-247..REQ-03-249
- **AC-021**: Repeated identical unresolved mentions across different source rows remain separate mention rows with distinct provenance and are never coalesced into a single mention record.
  - Verifies: REQ-01-196..REQ-01-227, REQ-02-042..REQ-02-044, REQ-02-058..REQ-02-059, REQ-03-129..REQ-03-134
- **AC-072**: A supported same-surface indicator-linking or extraction action on raw timeline, note, artifact, or evidence text preserves the raw field value and creates one or more source-bound `indicator_observation` rows without requiring dedicated IOC columns or rewriting the source text.
  - Verifies: REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204,
    REQ-03-135..REQ-03-137, REQ-03-247..REQ-03-249
- **AC-073**: Repeated identical indicator values observed in different source rows remain separate `indicator_observation` rows with distinct provenance and are never coalesced into a single observation record.
  - Verifies: REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204,
    REQ-03-135..REQ-03-137, REQ-03-247..REQ-03-249
- **AC-074**: Multiple source-bound `indicator_observation` rows can resolve to one canonical indicator record identified by incident scope plus deterministic indicator type and dedupe key, or an equivalent stable canonical identity.
  - Verifies: REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204,
    REQ-03-135..REQ-03-137, REQ-03-247..REQ-03-249
- **AC-075**: A canonical indicator can carry more than one attributed lifecycle interval within the same incident, and observation-derived `first_observed_at` or `last_observed_at` remain distinct from lifecycle `valid_from` or `valid_to`.
  - Verifies: REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204,
    REQ-03-135..REQ-03-137, REQ-03-247..REQ-03-249
- **AC-076**: Time-filtered indicator pivots distinguish observation time from asserted active-compromise or believed-validity time; filtering by one MUST NOT silently substitute the other.
  - Verifies: REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223,
    REQ-03-135..REQ-03-137
- **AC-077**: Retiring, clearing, or superseding an indicator does not rewrite preserved source text or delete prior `indicator_observation` rows; lifecycle changes append new structured history.
  - Verifies: REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223,
    REQ-03-135..REQ-03-137
- **AC-078**: The Indicators system view shows one row per canonical indicator record, not one row per source artifact or source observation, and remains stable across import, export, reporting, and future storage evolution.
  - Verifies: REQ-01-296..REQ-01-302, REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082,
    REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011, REQ-03-135..REQ-03-137
- **AC-079**: A hostname, MAC address, or similar value can be linked simultaneously to a host or identity record and to a canonical indicator record without forcing a shared object identity or deleting either linkage.
  - Verifies: REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223,
    REQ-03-135..REQ-03-137
- **AC-022**: Pasting into an `entity_origin` mapping upserts an existing active entity on a unique exact-match key and otherwise creates a stub; it never auto-merges two pre-existing entities.
  - Verifies: REQ-02-030..REQ-02-036, REQ-02-060..REQ-02-061
- **AC-023**: Merging two entities preserves loser lineage, repoints live mention resolutions and live links to the survivor in one change set, and does not change the survivor `record_id`.
  - Verifies: REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249
- **AC-024**: In the timeline sheet, the grouping control offers only `None`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions` in the base profile, and no other grouping key. In particular, it does not offer `timeline.event_type`, Summary, Hosts, Identities, Tags, or arbitrary custom columns.
  - Verifies: REQ-03-225..REQ-03-230
- **AC-025**: A grouped timeline sheet exposes exactly one derived outline level. `expand group`, `collapse group`, `expand all`, `collapse all`, and `Group: None` work without creating editable rows, paste targets, subtotal rows, or `record_id`-bound mutation targets.
  - Verifies: REQ-03-225..REQ-03-232
- **AC-026**: While grouped, sorting, filtering, paste, autosave, conflict handling, rollback, and export flatten to underlying records only. Editing a grouped field may move the row to a different visible group, but drag-and-drop reclassification and manual row-range grouping are not available.
  - Verifies: REQ-03-225..REQ-03-233

### 9.2 Import Extension Profile criteria

- **AC-027**: One uploaded CSV file or XLSX workbook import session starts without blocking grid editing, the UI shows progress and cancellation within 1 second of job start, and the operator can review per-unit discovery, preview, and header mapping before any apply step.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-153..REQ-03-186
- **AC-028**: File-based import uses the same mapping engine and `entity_binding_mode` semantics as clipboard-driven structured ingest. Importing through an `entity_origin` mapping upserts an existing active entity on a unique exact-match key and otherwise creates a stub; importing through `mention_origin` preserves mentions as mentions and never auto-creates stubs or auto-merges pre-existing entities.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-030..REQ-02-038, REQ-02-045..REQ-02-053, REQ-02-058..REQ-02-063,
    REQ-03-153..REQ-03-161, REQ-03-169..REQ-03-178
- **AC-029**: File-based import preserves unknown columns in raw-capture or custom-attribute storage, records `import_session_id`, `import_unit_id`, `mapping_fingerprint`, file kind, content hash, parser profile, parser version, and selected sheet or region locator as provenance, leaves imported host or account tokens unresolved unless an analyst explicitly resolves them later, and never executes workbook formulas, macros or VBA, workbook automation, or external links during import.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-030..REQ-02-038, REQ-02-045..REQ-02-053, REQ-02-062..REQ-02-063,
    REQ-03-153..REQ-03-161, REQ-03-169..REQ-03-178
- **AC-063**: Unsupported workbook features are downgraded only through the closed `warning_code[]` vocabulary declared by Core 03, preserved only as raw source metadata, or rejected as unsupported. No module outside the dedicated imports module links directly against XLSX or OpenXML parsing libraries or workbook-shape heuristics.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-153..REQ-03-178, REQ-03-193..REQ-03-204
- **AC-064**: One uploaded XLSX workbook creates one `import_session` that discovers explicit candidate `import_unit` objects for parser-resolved used ranges, Excel tables, eligible named ranges, and operator-selected regions. Each discovered unit exposes a deterministic `locator_kind`, canonical locator, inferred rectangular extent, and any nonblocking `warning_code[]` values before apply.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-162..REQ-03-186
- **AC-065**: Re-applying the same `(import_unit_id, mapping_fingerprint, incident_id)` tuple triggers duplicate-apply detection and blocks by default until the operator explicitly chooses re-import.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-169..REQ-03-192
- **AC-066**: Selecting overlapping `import_unit` rectangles in one batch is blocked before apply with an explicit overlap diagnostic. Non-overlapping selected units apply in deterministic order, one atomic `change_set` per unit, and the session can finish in `partially_applied`.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-169..REQ-03-186
- **AC-067**: Preview or apply of a unit containing formulas, merged cells, comments, pivots, charts, external links, hidden or filtered presentation state, or workbook or sheet protection never executes workbook behavior. The unit either emits only declared `warning_code[]` values or is rejected as unsupported, formula cells without stored cached values do not enter `ready` while mapped, and encrypted or password-protected workbooks that cannot be parsed are rejected before discovery.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-153..REQ-03-161, REQ-03-193..REQ-03-204

### 9.3 Snapshot and Reporting Extension Profile criteria

- **AC-030**: Report generation and snapshot generation run without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
  - Verifies: REQ-01-369..REQ-01-373, REQ-01-452..REQ-01-454, REQ-02-139..REQ-02-146
- **AC-031**: A generated HTML report or presentation artifact opens in a disconnected browser without fetching remote JavaScript, CSS, or fonts.
  - Verifies: REQ-01-370..REQ-01-373, REQ-01-394..REQ-01-398, REQ-02-139..REQ-02-146, REQ-04-044..REQ-04-047
- **AC-032**: Snapshot-derived outputs preserve stable identifiers and ordering consistent with the canonical derivation layer.
  - Verifies: REQ-01-342..REQ-01-348, REQ-01-351..REQ-01-353, REQ-01-367..REQ-01-368, REQ-01-370..REQ-01-373,
    REQ-02-139..REQ-02-146
- **AC-056**: Rendering from the same `snapshot_id`, `source_change_set_high_watermark` or equivalent frozen source boundary, `derivation_version`, `template_id`, `template_version`, `redaction_profile_id`, and `redaction_profile_version` produces the same `export_model_sha256` and deterministic export ordering.
  - Verifies: REQ-01-370..REQ-01-373, REQ-02-139..REQ-02-146
- **AC-057**: Rendering for a chosen `release_scope` fails closed if any rendered export-model field or block lacks an applicable redaction rule, and the rendered artifact includes a `redaction_manifest` keyed by stable export-model path and rule identifier.
  - Verifies: REQ-01-370..REQ-01-373, REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146,
    REQ-02-211, REQ-04-044..REQ-04-047
- **AC-058**: Template rendering fails closed when a template references an undeclared export-model binding or a missing required field, and no report renderer performs live workbook-table reads after the snapshot tuple has been fixed.
  - Verifies: REQ-01-370..REQ-01-373, REQ-01-381..REQ-01-384, REQ-02-139..REQ-02-146
- **AC-059**: For `internal_review`, exactly one `reviewer` approval is sufficient. For `external_release`, two distinct approvals are required, one `reviewer` and one `admin`. `internal_draft` requires no approval.
  - Verifies: REQ-01-374..REQ-01-380, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035, REQ-04-044..REQ-04-047
- **AC-060**: Changing `snapshot_id`, `template_id`, `template_version`, `redaction_profile_id`, `redaction_profile_version`, `output_kind`, `release_scope`, or rendered bytes invalidates prior release approvals automatically.
  - Verifies: REQ-01-374..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211,
    REQ-04-031..REQ-04-035, REQ-04-044..REQ-04-047
- **AC-061**: An `external_release` artifact contains no raw blob bytes or `working_material`, and any included `curated_narrative` block carries at least one `support_refs[]` entry to a supporting finding, event, evidence, assessment, or query record.
  - Verifies: REQ-01-377..REQ-01-380, REQ-01-394..REQ-01-397, REQ-02-139..REQ-02-146, REQ-02-211,
    REQ-04-044..REQ-04-047
- **AC-071**: Given a snapshot containing ad hoc note artifacts and a separately curated narrative block derived from one of those notes, an `external_release` artifact excludes the raw note text while allowing only the curated narrative block that satisfies `support_refs[]` and applicable redaction rules.
  - Verifies: REQ-01-377..REQ-01-380, REQ-02-139..REQ-02-146
- **AC-091**: Given a snapshot containing direct text from `task_request`, `decision`, `comm_log`, `handoff`, `status_review`, or `lesson` records and a separately curated export-model block derived from some of that content, an `external_release` artifact excludes the raw coordination-record text while allowing only the curated block that satisfies `support_refs[]` and applicable redaction rules.
  - Verifies: REQ-01-377..REQ-01-384, REQ-02-139..REQ-02-146, REQ-04-044..REQ-04-047
- **AC-062**: `mermaid` and `slidev` outputs may be published as `external_release` only when every rendered block is eligible for the chosen `release_scope`, and `reenactment` outputs are visibly marked `generated_presentation=true` and rejected for `external_release`.
  - Verifies: REQ-01-377..REQ-01-380, REQ-01-394..REQ-01-397, REQ-02-139..REQ-02-146, REQ-02-211,
    REQ-04-044..REQ-04-047
- **AC-113**: Selecting or changing `redaction_profile_id` and `redaction_profile_version` changes only snapshot-derived output and release state. It does not change live workbook query results, row visibility, field visibility, or evidence visibility for the same authenticated incident participant.
  - Verifies: REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211,
    REQ-04-044..REQ-04-047
- **AC-114**: Given one immutable snapshot containing export-model fields or blocks tagged with `disclosure_partition_refs[]` for two different affected parties, rendering an `external_release` with a redaction profile that allows only one party excludes or redacts the other party's content and fails closed if mixed-partition content lacks an applicable rule.
  - Verifies: REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211,
    REQ-04-044..REQ-04-047
- **AC-115**: Two `external_release` artifacts generated from the same immutable snapshot for two supported recipient-specific configurations require no manual post-render editing to satisfy the selected redaction profiles.
  - Verifies: REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211,
    REQ-04-044..REQ-04-047

### 9.4 Reference Pack Extension Profile criteria

- **AC-033**: Reference-pack import, verification, and refresh run without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
  - Verifies: REQ-01-399, REQ-01-407..REQ-01-413, REQ-01-452..REQ-01-454, REQ-04-040..REQ-04-043
- **AC-034**: Missing optional reference packs degrade only the affected overlay views or labels; timeline capture, entity resolution, and evidence attachment continue to function.
  - Verifies: REQ-01-282..REQ-01-284, REQ-01-399, REQ-01-419..REQ-01-422
- **AC-035**: Pack activation fails closed on checksum, signature, compatibility, safe-path, or equivalent integrity mismatch, and the previously active version remains active.
  - Verifies: REQ-01-399, REQ-01-409..REQ-01-421, REQ-04-040..REQ-04-043
- **AC-092**: In the smallest supported disconnected bundle, the only preinstalled active reference packs are `type_registry.host`, `type_registry.evidence`, and `type_registry.indicator`; the deployment remains usable without `framework.attack`, `framework.d3fend`, `framework.veris`, or any enrichment pack.
  - Verifies: REQ-01-282..REQ-01-284, REQ-01-400..REQ-01-406, REQ-04-040..REQ-04-043, REQ-04-054..REQ-04-055
- **AC-093**: Given an offline pack bundle placed in the configured reference-pack storage root or uploaded through the equivalent admin surface, the system stages the bundle under a temporary-work root, verifies it, records the version as `available` or equivalent non-active state on success, and does not activate it until explicit operator action.
  - Verifies: REQ-01-407..REQ-01-413, REQ-04-040..REQ-04-043
- **AC-094**: Given a candidate pack with checksum mismatch, signature mismatch, missing required integrity metadata, incompatible `pack_contract_version`, path-traversal attempt, or disallowed active content, import or activation fails closed, the candidate version remains inactive, and the previously active version, if any, remains active.
  - Verifies: REQ-01-407..REQ-01-418, REQ-04-040..REQ-04-043
- **AC-095**: Pack import and activation record structured attestation metadata queryable without unpacking bundle contents, including `pack_key`, `pack_kind`, `pack_version`, `manifest_sha256`, `payload_sha256`, `source_identifier`, `verification_method`, signer-key or trusted-source identifier, imported and activated actor attribution with timestamps, `previous_active_version`, and `verification_result`.
  - Verifies: REQ-01-409..REQ-01-421, REQ-04-040..REQ-04-043
- **AC-096**: In a disconnected deployment, reference-pack import or activation succeeds without outbound network access and no supported pack-activation path performs a live internet fetch.
  - Verifies: REQ-01-282..REQ-01-284, REQ-01-407..REQ-01-413, REQ-04-040..REQ-04-043, REQ-04-054..REQ-04-055

### 9.5 Enterprise Authentication Extension Profile criteria

- **AC-036**: Enterprise-authenticated users map to the same internal user identity used for attribution, and switching from local auth to enterprise auth does not break audit lineage for existing incidents.
  - Verifies: REQ-04-018..REQ-04-020

### 9.6 Additional Base Profile criteria for same-field conflict resolution

- **AC-037**: When two analysts edit the same write-back-capable field concurrently, the losing client shows a conflicted cell and same-surface resolver that presents saved value, unsaved value, row context, actor, and timestamp without leaving the sheet.
  - Verifies: REQ-03-041..REQ-03-051
- **AC-038**: Choosing `Keep saved value` clears the local conflict without creating a source revision; choosing `Use my unsaved value` or `Edit merged value` creates a new attributed change set and updates the visible row.
  - Verifies: REQ-03-041..REQ-03-051
- **AC-039**: If the field changes again while the resolver is open, stale resolution is rejected and the latest conflict payload is shown without losing the analyst's unsaved draft.
  - Verifies: REQ-03-041..REQ-03-051
- **AC-040**: A paste containing both non-conflicting and same-field-conflicting cells commits the non-conflicting cells immediately and groups the conflicting cells into a navigable conflict queue without per-cell modal interruption.
  - Verifies: REQ-03-048..REQ-03-051, REQ-03-083..REQ-03-085, REQ-03-221..REQ-03-222
- **AC-041**: Unresolved local conflict drafts are not broadcast to other analysts and do not appear in search, history, exports, or snapshots unless explicitly committed.
  - Verifies: REQ-03-048..REQ-03-051, REQ-03-077..REQ-03-082
- **AC-042**: After resolving a conflict, focus returns to the same cell and scroll position is preserved.
  - Verifies: REQ-03-041..REQ-03-051
- **AC-226**: When two analysts concurrently edit different lines of the same `text_compare_merge` field, the losing write fails with `409` and `error.code='same_field_conflict'`; if the server can compute a deterministic clean line merge from normalized `base_value`, `server_value`, and `client_value`, the conflict payload includes `suggested_merged_value`; no write is committed until the analyst explicitly resolves the conflict.
  - Verifies: REQ-03-048..REQ-03-076
- **AC-227**: When two analysts concurrently edit the same line, or both insert at the same base position, of the same `text_compare_merge` field, the losing write fails with `409` and `error.code='same_field_conflict'`, and the conflict payload omits `suggested_merged_value`.
  - Verifies: REQ-03-048..REQ-03-076
- **AC-228**: A same-field conflict payload for `text_compare_merge` always includes `base_value`; `client_value`, `server_value`, `base_value`, and optional `suggested_merged_value` are raw text scalars or `null`, not rendered fragments, token lists, diff scripts, or field-specific merge objects.
  - Verifies: REQ-03-048..REQ-03-076
- **AC-229**: For `text_compare_merge`, `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve` with `resolution_kind='merged_value'` accepts only a final text scalar or `null` in `resolved_value`; a successful resolution creates exactly one new attributed `change_set` and MUST NOT accept a diff script, merge opcode list, AST, or field-specific merge action object.
  - Verifies: REQ-03-048..REQ-03-076
- **AC-230**: `text_compare_merge` conflict detection and suggestion generation operate on plain text after normalizing `CRLF` and `CR` to `LF` for merge computation, and Markdown syntax, HTML markup, entity-chip rendering, or link presentation in the field content do not change conflict detection or suggestion-generation outcomes.
  - Verifies: REQ-03-048..REQ-03-076

### 9.7 Additional Base Profile criteria for threat model and focused weakness controls

- **AC-048**: The implementation maintains a STRIDE threat model for the current release that covers, at minimum, authenticated sessions, incident records and revisions, evidence blobs and previews, reference packs and import bundles, generated snapshots and exports, and portable runtime roots, and each entry maps the threat to at least one control and one verification hook.
  - Verifies: REQ-04-049..REQ-04-051
- **AC-049**: Rendering notes, markdown, evidence metadata, filenames, tags, and other incident-authored text in the browser UI or generated HTML does not execute script, inline event handlers, `javascript:` URLs, or remote asset fetches sourced from incident data.
  - Verifies: REQ-04-052..REQ-04-053
- **AC-050**: CSV, XLSX, and spreadsheet-oriented clipboard exports neutralize formula-leading characters by default. A raw export mode, if implemented, requires explicit operator opt-in and a visible unsafe-export warning.
  - Verifies: REQ-04-052..REQ-04-053
- **AC-051**: Upload, import, and archive-extraction paths reject absolute paths and parent traversal and do not write outside the configured runtime roots. Client-supplied filenames do not determine storage keys.
  - Verifies: REQ-01-455..REQ-01-456, REQ-04-052..REQ-04-053, REQ-04-058..REQ-04-059
- **AC-052**: Reference-pack activation, and incident bundle import when implemented, fail closed on checksum mismatch, signature mismatch, incomplete download, or missing required integrity metadata.
  - Verifies: REQ-04-040..REQ-04-043, REQ-04-052..REQ-04-053
- **AC-053**: Evidence upload and preview do not execute uploaded active content in the main application unit or browser origin. Non-previewable or active-content types remain quarantined or download-only unless an explicit isolated analysis path is configured.
  - Verifies: REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-201, REQ-03-127..REQ-03-128, REQ-04-048,
    REQ-04-052..REQ-04-053
- **AC-054**: Attempting to mutate, preview, download, or issue object-store access for a `record_id`, `evidence_record_id`, `object_blob_id`, or snapshot outside the caller's incident membership is denied even when the identifier is otherwise valid.
  - Verifies: REQ-01-355..REQ-01-366, REQ-03-127..REQ-03-128, REQ-04-021..REQ-04-030, REQ-04-052..REQ-04-053
- **AC-055**: A deployment claiming flyaway or disconnected portability for use on portable hosts or removable media stores database, object, reference-pack, temporary-work-file, and export roots on encrypted storage. A deployment that does not do so is non-conformant for flyaway handling.
  - Verifies: REQ-01-455..REQ-01-456, REQ-04-052..REQ-04-055, REQ-04-058..REQ-04-059

### 9.8 Additional Base Profile criteria for promoted recurrent fields

- **AC-097**: The Hosts sheet can sort or filter on `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` from the grid surface without opening the inspector and without rereading unrelated note text or blob metadata.
  - Verifies: REQ-01-323..REQ-01-325, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246
- **AC-098**: The Identities sheet can sort or filter on `privilege_level`, `mfa_state`, and `reset_status` from the grid surface without rereading unrelated note text or evidence blobs.
  - Verifies: REQ-01-326..REQ-01-327, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246
- **AC-099**: On the implementation-supported incident metadata surface, editing `tlp`, `current_phase`, and `primary_external_case_ref` persists those values as structured fields that survive reload, deterministic projection rebuild, and snapshot generation; they do not disappear into `custom_attrs`.
  - Verifies: REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246
- **AC-100**: The Evidence sheet can sort or filter on `requested_at`, `received_at`, `collector_party_text`, `source_party_text`, `storage_ref`, `blob_hash`, and upload or attachment state without fetching blob bytes.
  - Verifies: REQ-01-328, REQ-01-355..REQ-01-366, REQ-02-009..REQ-02-023, REQ-02-186..REQ-02-201,
    REQ-03-242..REQ-03-246
- **AC-101**: If the implementation exposes structured findings, investigative queries, or forensic keywords as workbook surfaces, their defining fields remain directly filterable and exportable as structured fields rather than JSON-only payloads.
  - Verifies: REQ-02-009..REQ-02-023, REQ-02-135..REQ-02-138

### 9.9 Lifecycle machine criteria

- **AC-102**: A blob slot left in `pending` without successful finalization does not create or imply an attached evidence record, does not increment visible evidence counts, and by no later than the first cleanup sweep after `pending_expires_at` transitions to `upload_state='failed'` with `terminal_reason='pending_timeout'`.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-120
- **AC-103**: If an evidence record points at a blob in `pending`, `failed`, or missing state, preview and download are blocked and the row surfaces as non-available or inconsistent until explicit repair or re-finalization completes.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128
- **AC-154**: A pending blob slot allows 3 failed explicit finalization attempts; the 4th failed attempt transitions the slot to `upload_state='failed'` with `terminal_reason='finalize_retry_exhausted'`; only failed explicit finalization attempts count toward this total, and an idempotent replay after already-committed success does not consume retry budget.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-126
- **AC-155**: Pending-blob timeout handling runs at least every 15 minutes; by no later than `pending_expires_at + 15 minutes`, an unfinalized slot is failed, orphaned bytes for a failed unattached slot are deleted within 1 hour of terminal failure, and failed unattached slot metadata remains queryable for at least 7 days before automatic hard deletion.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-126
- **AC-104**: Rendering a report or presentation artifact creates a release record in `pending_approval` bound to one immutable release tuple and one `output_sha256`.
  - Verifies: REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035
- **AC-105**: Satisfying the required approval set moves that exact release record to `approved`, and any attempted publish action on a non-approved record is rejected.
  - Verifies: REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035
- **AC-106**: Rendering a superseding artifact for the same logical output slot with a different bound tuple or different `output_sha256` creates a new `pending_approval` candidate and invalidates the prior current artifact rather than inheriting its approval or publication state.
  - Verifies: REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035
- **AC-107**: For every normative lifecycle machine claimed by the implementation, the happy-path transitions succeed only in the documented order and the persisted authoritative state matches the documented machine condition after each step.
  - Verifies: REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110
- **AC-108**: For every normative lifecycle machine claimed by the implementation, at least one terminal failure path moves to the documented failure or non-active state and blocks any forbidden follow-on action.
  - Verifies: REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110
- **AC-109**: Illegal lifecycle transitions are rejected with no partial state advancement and no false success signal.
  - Verifies: REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110
- **AC-110**: Retrying the last lifecycle transition after a simulated crash or duplicate delivery is idempotent or otherwise produces the documented equivalent final state without duplicating durable side effects.
  - Verifies: REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110
- **AC-111**: Replaying the same starting state and the same ordered inputs for a normative lifecycle machine produces the same final persisted state and the same documented observable signals.
  - Verifies: REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110
- **AC-137**: A `task_request` can transition `open -> in_progress -> blocked -> in_progress -> done`; `blocked_reason` is present only while `status='blocked'`; `completed_at` is present only while `status='done'`; and each committed transition increments `row_version`, updates the task projection row, appends one `change_set` plus mutation entries, and emits `record_changed`.
  - Verifies: REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110,
    REQ-03-255..REQ-03-260
- **AC-138**: A `task_request` transition `done -> open` succeeds, clears persisted `completed_at`, preserves prior history, and returns committed row values that show the reopened task in `status='open'`.
  - Verifies: REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110,
    REQ-03-255..REQ-03-260
- **AC-139**: A requested `task_request` transition `done -> canceled` or `canceled -> done` is rejected with `error.code='illegal_transition'`, `error.status=409`, `error.details.from_status`, `error.details.to_status`, and `error.details.violated_guards[]`; no status, guard field, projection row, `change_set`, or mutation entry is partially advanced.
  - Verifies: REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110,
    REQ-03-255..REQ-03-260
- **AC-140**: A `task_request` transition away from `blocked` clears persisted `blocked_reason`, and a successful write that sets `status='done'` without an explicit `completed_at` stores `completed_at` at or after the commit time and not earlier than `created_at`.
  - Verifies: REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110,
    REQ-03-255..REQ-03-260
- **AC-141**: A `decision` can transition `proposed -> approved -> executed`, and `approved` is treated only as an incident-coordination state rather than as a generalized approval workflow for ordinary row edits.
  - Verifies: REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223,
    REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260
- **AC-142**: A direct write that sets `decision.status='superseded'` is rejected with `error.code='illegal_transition'` and `error.status=409`, and a direct write `approved -> rejected`, `rejected -> proposed`, or `executed -> approved` is likewise rejected with no partial state advancement.
  - Verifies: REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223,
    REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260
- **AC-143**: An explicit supersession action from decision B to decision A succeeds only when B and A are different decisions in the same incident and B already has `status` `approved` or `executed`; the committed action persists the supersession relation, moves A from `proposed` or `approved` to `superseded`, increments `row_version`, appends one `change_set` plus mutation entries for the changed records, updates derived projections, and leaves B in its preexisting `approved` or `executed` status.
  - Verifies: REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223,
    REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260
- **AC-144**: If an explicit supersession action targets a decision already in `executed`, the target record remains `executed`, the supersession relation still persists, and the decision view surfaces `decision.is_superseded=true` or an equivalent computed indicator for that target without rewriting its persisted `status`.
  - Verifies: REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223,
    REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260
- **AC-145**: Replaying the same legal `task_request` status change, legal `decision` status change, or legal explicit supersession action after a simulated crash or duplicate delivery is idempotent and does not duplicate durable side effects such as extra `change_set` records, extra mutation entries, projection updates, or repeated status flips.
  - Verifies: REQ-01-336..REQ-01-341, REQ-02-094..REQ-02-119, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110,
    REQ-03-255..REQ-03-260

### 9.10 Additional Base Profile criteria for public interface surface

- **AC-123**: `GET /api/v1/auth/session` returns the authenticated internal user identity, display name, provider kind, MFA state, `is_deployment_admin`, `authenticated_at`, `idle_expires_at`, `absolute_expires_at`, and `session_expires_at`, plus the caller's incident memberships or current incident-role context, without requiring client-side token parsing; `session_expires_at` is the earlier of `idle_expires_at` and `absolute_expires_at`.
  - Verifies: REQ-00-014, REQ-01-023..REQ-01-031, REQ-04-001..REQ-04-017
- **AC-124**: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` accepts a field-key-based sort, filter, and grouping contract and returns rows with `record_id`, `row_version`, field-key-addressable cells, and cursor metadata; group headers are not serialized as writable rows.
  - Verifies: REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-285..REQ-01-290,
    REQ-01-307..REQ-01-341, REQ-03-223..REQ-03-224, REQ-03-236..REQ-03-241
- **AC-125**: View-scoped row creation and record-scoped patch operate via `view_schema_id`, `client_txn_id`, `base_row_version`, and `changes[]` keyed by `field_key`; the client can mutate one writable field without resubmitting a full row snapshot.
  - Verifies: REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-057..REQ-01-088, REQ-01-285..REQ-01-290,
    REQ-01-307..REQ-01-330, REQ-02-208..REQ-02-209, REQ-03-086, REQ-03-111..REQ-03-115, REQ-03-236..REQ-03-241
- **AC-151**: `GET /api/v1/incidents/{incident_id}/saved-views` returns only saved views visible to the caller, and each returned resource includes `saved_view_id`, `incident_id`, `view_schema_id`, `scope`, `display_name`, `query_json`, `layout_json`, `owner_user_id`, timestamps, and `saved_view_version`; `query_json` uses stable `field_key` and grouping identifiers rather than visible labels.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021
- **AC-152**: `POST /api/v1/incidents/{incident_id}/saved-views` defaults omitted `scope` to `private` and rejects `scope='system'`; `PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` accepts `base_saved_view_version` and mutable fields only, and a stale base version fails with an explicit conflict status rather than silently overwriting saved-view state; `DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` deletes only the configuration object and never underlying incident records or links.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-016, REQ-03-022..REQ-03-026
- **AC-153**: `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` read or replace only the caller's nullable `home_sheet_ref` and succeed for any incident member, including `viewer`; `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` read or replace the incident's nullable `default_sheet_ref`, and the `PUT` route fails closed for non-admin incident roles.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032
- **AC-126**: Same-field conflict responses use the generic error envelope with `error.code='same_field_conflict'` and the conflict object required by Core 03 §3.3.4; a stale `conflict_token` is rejected with a fresh conflict payload rather than silently overwriting saved state.
  - Verifies: REQ-01-019, REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-03-036..REQ-03-040,
    REQ-03-063..REQ-03-076
- **AC-200**: Patching `timeline.tags` with one `add_tag` action and one `remove_tag` action adds exactly one incident-scoped tag binding and removes exactly one binding.
  - Verifies: REQ-01-057..REQ-01-088, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209
- **AC-188**: Patching `timeline.host_refs` or `timeline.identity_refs` with `dismiss_item` against an unresolved or resolved `entity_mention` preserves `raw_text`, leaves the same mention row and provenance intact, sets `resolution_status='dismissed'`, clears `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`, removes or tombstones any corresponding active resolved link in the same `change_set`, and appends exactly one new `change_set` with mention-target mutation detail.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-02-058..REQ-02-059, REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216,
    REQ-03-236..REQ-03-241
- **AC-189**: Patching `timeline.host_refs` or `timeline.identity_refs` with `revert_to_unresolved` against a dismissed `entity_mention` returns that same mention row to `resolution_status='unresolved'`, preserves `raw_text`, leaves `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method` null after commit, leaves no active derived resolved link, and appends exactly one new `change_set`; reviewer rollback of the dismissal restores the exact pre-dismiss state as a separate attributed revision rather than rewriting history.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241
- **AC-190**: If a row's last non-deleted unresolved host or identity mention is dismissed, the committed row computes `timeline.has_unresolved_mentions=false`, the row no longer matches unresolved-only filters or equivalent unresolved-resolution queue views, grouping by `timeline.has_unresolved_mentions` places the row in the `false` bucket, and the active `timeline.host_refs` or `timeline.identity_refs` collection value omits the dismissed mention while history or inspector affordances can still surface it.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241
- **AC-201**: Patching `timeline.host_refs` with `resolve_item` preserves `raw_text`, resolves the targeted `entity_mention`, and leaves the corresponding active host link present after commit.
  - Verifies: REQ-01-057..REQ-01-088, REQ-02-030..REQ-02-036, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209
- **AC-221**: `POST /api/v1/entity-mentions/{entity_mention_id}/resolve` with JSON `base_mention_row_version`, `client_txn_id`, `action='resolve_item'`, and `resolved_record_id` by an `editor`, `reviewer`, or `admin` on the source incident resolves one visible active host or identity mention, returns `200 OK` with `incident_id`, updated `entity_mention`, `source_record.record_id`, incremented `entity_mention.row_version`, incremented `source_record.row_version`, `change_set_id`, and `active_link`, preserves `raw_text`, sets `resolution_status='resolved'`, sets non-null `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method='explicit_resolve_route'`, and when the mention was previously resolved to a different target, removes or tombstones the old active resolved link in the same `change_set`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-03-129..REQ-03-134
- **AC-222**: The same route with JSON `action='dismiss_item'` against an unresolved or resolved visible active host or identity mention returns `200 OK`, preserves `raw_text`, stable mention identity, and provenance, sets `resolution_status='dismissed'`, returns `resolved_record_id=null`, `resolved_by_user_id=null`, `resolved_at=null`, and `resolution_method=null` on the committed `entity_mention`, omits `active_link`, and removes or tombstones any corresponding active resolved link in the same `change_set`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-03-129..REQ-03-134
- **AC-223**: The same route with JSON `action='revert_to_unresolved'` against a resolved or dismissed visible active host or identity mention returns `200 OK`, preserves `raw_text`, sets `resolution_status='unresolved'`, returns `resolved_record_id=null`, `resolved_by_user_id=null`, `resolved_at=null`, and `resolution_method=null` on the committed `entity_mention`, omits `active_link`, and when the starting state was `dismissed`, does not silently relink any prior resolved target; exact pre-dismiss recovery remains reviewer rollback.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-03-129..REQ-03-134
- **AC-224**: The same route with a stale `base_mention_row_version` fails with `409 error.code='row_version_conflict'` and `error.details` containing `entity_mention_id`, `base_mention_row_version`, `current_mention_row_version`, and `source_record_id`; a direct `resolve_item` against a currently dismissed mention fails with `409 error.code='illegal_transition'`; a request against a soft-deleted source record fails with `409 error.code='record_deleted_use_restore'`; a caller lacking visibility to the mention receives `404 error.code='entity_mention_not_found'`; a caller who can see the mention but lacks `editor`, `reviewer`, or `admin` role receives `403`; a supplied `resolved_record_id` that does not identify a visible active target fails with `404 error.code='resolved_record_not_found'`; and missing required members, unknown top-level members, forbidden or missing `resolved_record_id`, wrong target type, target from another incident, or embedded entity-create payload fail with `400 error.code='invalid_mutation_payload'`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-030..REQ-02-036, REQ-02-039..REQ-02-041,
    REQ-03-129..REQ-03-134
- **AC-225**: Replaying `POST /api/v1/entity-mentions/{entity_mention_id}/resolve` with the same normalized request by the same actor and the same `(entity_mention_id, client_txn_id)` returns the originally committed `200 OK` success without creating a second `change_set`; reusing that key with a different normalized request fails with `409 error.code='client_txn_conflict'`; and a successful explicit mention action reaches subscribers only through the ordinary `record_changed` event for the source record with no mention-specific WebSocket family.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-030..REQ-02-036, REQ-02-039..REQ-02-041,
    REQ-03-129..REQ-03-134
- **AC-202**: Patching `host.aliases` or `identity.aliases` with `add_alias` plus `remove_alias` yields a committed alias collection that round-trips as `collection_value_v1`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209
- **AC-203**: A same-field conflict payload for a `collection_review` field returns `collection_value_v1` in `client_value`, `server_value`, and `base_value` rather than a raw string array or plain delimited text.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209,
    REQ-03-048..REQ-03-053, REQ-03-063..REQ-03-076
- **AC-204**: Sending a raw string, blind full-collection replacement, unknown collection action `op`, unknown payload `kind`, or foreign `item_ref` to a `collection_review` field fails with `400` and `error.code='invalid_mutation_payload'`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209,
    REQ-03-048..REQ-03-053, REQ-03-063..REQ-03-076
- **AC-205**: When the interactive host or identity auto-resolution flow defined by Core 03 §12.3 succeeds, the committed action resolves the corresponding `entity_mention` and produces exactly one active `record_link` whose `link_type` is `observed_on_host` or `observed_as_identity` according to the mutated field, whose `provenance` is `auto_match`, and whose `confidence` is `100`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209,
    REQ-03-205..REQ-03-216
- **AC-206**: A base-profile relationship mutation that attempts to create a self-link, targets a record from a different incident, or targets a non-record mutation object such as `entity_mention` or `indicator_observation` fails closed, creates no durable `record_link`, and leaves no misleading projection update.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209
- **AC-207**: Sending client-chosen `link_type`, direction flags, table names, or storage-routing metadata in a base-profile relationship mutation fails with `400` and `error.code='invalid_mutation_payload'`; the server accepts only field-key-derived or action-route-derived relationship routing.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209
- **AC-208**: Repeating the same logical relationship add through supported base-profile routes, including duplicate collection actions, idempotent request replay, or a later add of the same `(incident_id, src_record_id, dst_record_id, link_type)` tuple, leaves exactly one non-deleted `record_link` for that tuple while preserving attributable history of the attempted operations.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209
- **AC-209**: Merging a host or identity whose incoming or outgoing links collide with links already present on the survivor repoints both incoming and outgoing active links in the same `change_set`, preserves canonical direction, and leaves at most one non-deleted tuple for each `(incident_id, src_record_id, dst_record_id, link_type)` after deduplication.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-181..REQ-01-195, REQ-02-054..REQ-02-055, REQ-02-064..REQ-02-066,
    REQ-02-163..REQ-02-185, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249
- **AC-210**: Projection-backed linked-count fields, visible linked-record chips, and any operator-visible current-state export or report field that derives relationships from `record_links` include only links whose own row is not soft-deleted and whose source and destination records are not soft-deleted; the same inactive links remain visible through history or rollback surfaces.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-342..REQ-01-348, REQ-01-351..REQ-01-353, REQ-01-355..REQ-01-366,
    REQ-02-163..REQ-02-185, REQ-03-247..REQ-03-249
- **AC-127**: Potentially large list or view-query routes use opaque cursor pagination with `has_more` and `next_cursor`; replaying a cursor against a different incident, view, sort tuple, filter set, or grouping contract is rejected rather than reinterpreted.
  - Verifies: REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-240..REQ-01-242
- **AC-238**: On `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` without `cursor_token`, omitting `limit` yields `meta.paging.limit=100`; when more rows remain, the same response uses `meta.paging.has_more=true` and a non-null `meta.paging.next_cursor`.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-242
- **AC-239**: Repeating the same route with a valid `cursor_token` and omitted `limit` reuses the cursor-bound effective `meta.paging.limit`; replaying that cursor with a different explicit `limit` fails with `400 error.code='invalid_view_query'` and `error.details.reason_code='cursor_query_mismatch'`.
  - Verifies: REQ-01-035, REQ-01-234, REQ-01-238, REQ-01-241, REQ-01-242
- **AC-240**: `limit=0`, `limit=-1`, `limit=501`, a non-integer `limit`, and use of `block_size` or `page_size` in the request each fail closed with `400 error.code='invalid_view_query'` and `error.details.reason_code='invalid_limit'`.
  - Verifies: REQ-01-035, REQ-01-234, REQ-01-238
- **AC-241**: A terminal page reached through cursor progression returns `meta.paging.limit`, `meta.paging.has_more=false`, and `meta.paging.next_cursor=null`; a zero-match first page uses the same terminal representation and returns `rows=[]`.
  - Verifies: REQ-01-036, REQ-01-242
- **AC-242**: For paged view-query responses, continuation is determined by `meta.paging.has_more` and `meta.paging.next_cursor`; conformance evidence and client continuation logic do not infer terminal state from `rows.length < meta.paging.limit` alone.
  - Verifies: REQ-01-242
- **AC-243**: With grouping active and `limit=1`, the response serializes exactly one data row in `rows[]`, includes no group-header pseudo-row, and still returns the `group_values` needed for client-local grouping; page-size accounting applies to serialized `rows[]` entries only.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-037
- **AC-128**: `POST /api/v1/object-blobs` creates a pending blob slot and returns `object_blob_id`, `upload_state`, `target_expires_at`, `pending_expires_at`, and a short-lived upload target; in the base profile the upload target expires 60 minutes after issuance, the pending slot expires 24 hours after creation, retry after target expiry uses a fresh blob slot rather than same-slot target refresh, and attaching that blob to incident-visible evidence, or previewing the resulting evidence surface, fails closed until the blob or evidence lifecycle reaches the states required by Core 03 §8 and Core 02 §14.1.
  - Verifies: REQ-00-014, REQ-01-019, REQ-01-243..REQ-01-247, REQ-01-328, REQ-01-355..REQ-01-366,
    REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128, REQ-04-048
- **AC-129**: Long-running operations started through the public interface return `202 Accepted` with a `job_id`; `GET /api/v1/jobs/{job_id}` exposes progress and terminal result or error summary; the incident-scoped WebSocket stream authenticates with the same session contract and emits presence, `record_changed`, and `job_progress` events without broadcasting client-local drafts or grouping UI state.
  - Verifies: REQ-00-014, REQ-01-018..REQ-01-019, REQ-01-248..REQ-01-277, REQ-01-452..REQ-01-454,
    REQ-03-092..REQ-03-098
- **AC-130**: A cookie-authenticated state-changing request with missing or invalid CSRF proof fails closed, and the same request succeeds when a valid CSRF proof accompanies an otherwise authorized session.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-154, REQ-04-001..REQ-04-004, REQ-04-052..REQ-04-053
- **AC-131**: On `GET /ws/v1/incidents/{incident_id}`, an authorized same-origin client that sends `hello` as its first application message receives `hello_ack` containing `connection_id`, `resume_token`, `server_time`, `heartbeat_interval_ms=15000`, `presence_ttl_ms=45000`, and `resume_window_ms>=300000`, followed by `presence_snapshot`; a cookie-authenticated browser connection from an untrusted `Origin` is rejected before `hello_ack` or `presence_snapshot`, and an otherwise idle accepted connection receives application-level `ping` within 15 seconds and is closed after 45 seconds without any inbound frame or timely `pong`.
  - Verifies: REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098, REQ-04-005..REQ-04-017,
    REQ-04-052..REQ-04-053
- **AC-132**: When analyst A changes workbook surface, focused row, or same-cell edit state on an incident, analyst B subscribed to the same incident receives the corresponding `presence_delta` within 1 second, and the UI renders workbook-header, row-gutter, and same-cell indicators from matching `sheet_ref`, `record_id`, and `field_key` rather than visible labels or row numbers.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-090..REQ-03-098
- **AC-133**: After a transient disconnect, a client that reconnects to the same incident within the retained replay window can send `resume` with `resume_token` and `last_seen_stream_seq`, receives `resume_ack.status='replayed'`, then receives missed `record_changed` and incident-scoped `job_progress` messages in strict ascending `stream_seq`, followed by a fresh `presence_snapshot`.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098
- **AC-134**: If a client reconnects with an expired, unknown, malformed, or too-old `resume_token`, but the caller still has valid incident authorization, the server responds with `resume_ack.status='reset_required'`, sends a fresh `presence_snapshot`, and emits no guessed or partial replay for the missing range; the client recovers only by re-querying current workbook state through the existing HTTP view route.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098
- **AC-135**: Replayable WebSocket messages on one incident carry monotonically increasing `stream_seq` values assigned only after the underlying record mutation or incident-scoped job-state change is committed; clients can de-duplicate duplicates by `(incident_id, stream_seq)`, reconcile row application by `record_id` plus `row_version`, and must perform HTTP resynchronization rather than guessed incremental apply when a replayable sequence gap is observed.
  - Verifies: REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098
- **AC-136**: If the authenticated session expires or the caller's incident membership is revoked after connection establishment, the server sends `session_revoked`, closes the socket, and emits no further incident `presence`, `record_changed`, or `job_progress` messages to that client on that connection.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098, REQ-04-005..REQ-04-017
- **AC-156**: After 30 minutes with no qualifying activity, the next request on that session to `GET /api/v1/auth/session`, query a workbook view, or submit a mutation fails with `401` or an equivalent `session_expired` error, and no mutation or job start is partially applied.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017
- **AC-157**: Continuous qualifying activity can slide `idle_expires_at` but cannot keep a session alive past `authenticated_at + 12 hours`; after absolute expiry, the next authenticated request fails closed and any accepted WebSocket connection on that session receives `session_revoked` with `reason_code='session_expired'` before close.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017
- **AC-158**: `GET /api/v1/auth/session`, WebSocket `ping` or `pong`, passive server events, and successful `resume` replay do not count as qualifying activity; a session that receives only those events for 30 minutes still expires on the normal idle boundary.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017
- **AC-159**: A 6th concurrent human-user login revokes the least-recently-used non-current session before issuing the new session, records an attributed audit event with reason code `concurrency_limit`, and any accepted WebSocket connection on the revoked session receives `session_revoked` with `reason_code='concurrency_limit'` before close.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017
- **AC-160**: `POST /api/v1/auth/logout` invalidates only the current session immediately, causes any accepted WebSocket connection on that session to receive `session_revoked` with `reason_code='session_revoked'`, and leaves other still-valid sessions for that user authorized until another revoke condition occurs; password change, MFA reset, account disablement, or an explicit deployment-admin revoke-all action invalidates all active sessions immediately.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017
- **AC-161**: A `resume_token` is rejected as an HTTP authentication credential, becomes unusable after the earlier of replay-window expiry and underlying session expiry or revocation, and after session expiry or revocation a reconnect succeeds only after a new authenticated session is established and the client begins again with `hello` rather than relying on `resume` alone.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017
- **AC-162**: If a user loses membership in incident A while retaining a still-valid session and access to incident B, the connection and future requests for incident A fail closed, the incident-A socket receives `session_revoked` with `reason_code='incident_access_revoked'`, and the same session can still query or subscribe to incident B if otherwise authorized.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017
- **AC-163**: If session expiry or revocation occurs while the client holds queued unsent patches or unresolved same-field local drafts, the client preserves that unsaved work locally, prompts for re-authentication when required, and retries only through the normal authenticated patch or conflict-resolution paths after a new session is established; no queued draft becomes authoritative without passing the ordinary row-version, authorization, and conflict checks.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-077..REQ-03-082, REQ-03-099..REQ-03-100, REQ-04-005..REQ-04-017

- **AC-170**: An authenticated session whose internal user account is active and not disabled can `POST /api/v1/incidents` with required `client_txn_id`, `incident_key`, and `title`, plus optional nullable `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`; on first success it receives `201 Created` plus `Location: /api/v1/incidents/{incident_id}`, and the response `data` includes the incident resource fields defined by Core 01 §3.3.5.3 with `status='active'`, `incident_version=1`, `closed_at=null`, `created_by_user_id` equal to the actor, `updated_by_user_id` equal to the actor, and `created_at == updated_at`.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-171**: A successful incident create bootstraps exactly one creator membership with `role='admin'`, `GET /api/v1/incidents/{incident_id}/memberships` returns that membership for the creator, `GET /api/v1/incidents/{incident_id}/workbook-preferences/default` returns a resource with `default_sheet_ref=null`, and `GET /api/v1/incidents/{incident_id}/workbook-preferences/me` for the creator returns a resource with `home_sheet_ref=null`.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-172**: `POST /api/v1/incidents` without authentication, with an expired or revoked session, from a disabled internal user account, or with missing or invalid CSRF protection for a cookie-authenticated request fails closed and creates no incident, membership, or workbook-preference state.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-173**: Replaying `POST /api/v1/incidents` with the same `(actor_user_id, client_txn_id)` and the same normalized request returns `200 OK`, repeats `Location: /api/v1/incidents/{incident_id}` for the originally created incident, returns the originally created incident resource, and creates no second incident; replaying with the same `(actor_user_id, client_txn_id)` and a different normalized request fails with `409` and `error.code='client_txn_conflict'`.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-174**: Creating an incident with an `incident_key` whose trimmed Unicode-NFC-normalized form conflicts with an existing incident fails with `409` and `error.code='incident_key_conflict'`; `error.details.field='incident_key'` and `error.details.incident_key_canonical` are present; and no partial incident, membership, or workbook-preference bootstrap state is left behind.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-219**: `POST /api/v1/incidents` whose body is not a JSON object, omits required `client_txn_id`, `incident_key`, or `title`, supplies `null` for a non-nullable field, exceeds the declared field-length limits, includes a control character in `incident_key` or `title`, attempts to set a server-managed field, or sends `initial_memberships[]` or an equivalent collaborator-seeding payload fails with `400` and `error.code='invalid_incident_create'`; when one member is responsible, the response includes `error.details.field`, and `error.details.reason_code` uses the Core 01 registry.
  - Verifies: REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239
- **AC-220**: Unknown extension members on `POST /api/v1/incidents` are ignored for persistence and response shaping and are excluded from normalized-request comparison; omission and explicit `null` compare equal for optional nullable create fields; therefore a replay that differs only by unknown extension members or by omission versus `null` for optional nullable create fields still returns the originally created incident rather than creating a second incident or triggering `client_txn_conflict`.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-211**: `GET /api/v1/incidents/{incident_id}` returns the incident resource defined by Core 01 §3.3.5.3 and §3.3.5.3.1, including `incident_version`, `updated_at`, and `updated_by_user_id`, to any current incident member; a caller lacking visibility receives `404`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023
- **AC-212**: `PATCH /api/v1/incidents/{incident_id}` with a correct `base_incident_version` by a `reviewer` or `admin` can change only `tlp`, `current_phase`, and `primary_external_case_ref`; `null` clears a nullable field; success returns `200 OK` with the updated incident resource; an effective change increments `incident_version` exactly once and sets `updated_at` plus `updated_by_user_id`; a no-op patch returns `200 OK` without incrementing `incident_version`, changing `updated_at`, or changing `updated_by_user_id`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-02-014..REQ-02-023
- **AC-213**: `PATCH /api/v1/incidents/{incident_id}` with a stale `base_incident_version` fails with `409` and `error.code='incident_version_conflict'`; a caller who can see the incident but lacks sufficient role gets `403`; a caller lacking visibility gets `404`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023
- **AC-214**: `PATCH /api/v1/incidents/{incident_id}` that attempts to mutate `incident_id`, `incident_key`, `title`, `description`, `status`, `severity`, `created_by_user_id`, `created_at`, `updated_at`, `updated_by_user_id`, `incident_version`, `closed_at`, membership objects, saved-view objects, or workbook-preference objects, or that sends unknown top-level members, fails with `400` and `error.code='invalid_incident_patch'`, and leaves no partial incident state.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023
- **AC-175**: A caller with `is_deployment_admin=true` can `POST /api/v1/users` with `client_txn_id`, `auth_kind='local'`, `email`, `display_name`, and `initial_password`, receives `201 Created`, and can read the resulting safe user resource through both `GET /api/v1/users/{user_id}` and `GET /api/v1/users`; those route responses never include password hashes, TOTP secrets, WebAuthn credential material, opaque session tokens, or provider assertions.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137
- **AC-176**: Replaying `POST /api/v1/users` with the same `(actor_user_id, client_txn_id)` and the same normalized request returns `200 OK` and the originally created `user_id`; reusing that key with a different normalized request fails with `409`; an unauthenticated caller or an authenticated caller without `is_deployment_admin=true` cannot list, create, or patch users through the public route family; and the first deployment admin is not creatable through an unauthenticated public request.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137
- **AC-177**: `PATCH /api/v1/users/{user_id}` with a correct `base_user_version` can change only the mutable fields allowed by Core 01 §3.3.5.1; a stale `base_user_version` fails with `409` and `error.code = user_version_conflict`; deactivating a user revokes all active sessions immediately but leaves incident memberships intact; and demoting or deactivating the last active deployment admin fails with `409` and `error.code = last_deployment_admin`.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137
- **AC-178**: Any current incident member can `GET /api/v1/incidents/{incident_id}/memberships`; an incident admin can `POST /api/v1/incidents/{incident_id}/memberships` with exactly one of `user_id` or `email` and one valid role; creating a new membership returns `201 Created`; re-adding the same user with the same role returns `200 OK` and does not create a second membership row; re-adding the same user with a different role fails with `409` and `error.code = membership_exists_use_patch`; supplying a nonexistent user fails with `404` and `error.code = user_not_found`; supplying an inactive user fails with `409` and `error.code = user_inactive`; and the route never auto-creates or invites a user.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030
- **AC-179**: `PATCH /api/v1/incidents/{incident_id}/memberships/{user_id}` with a correct `base_membership_version` changes only `role`; a stale version fails with `409` and `error.code = membership_version_conflict`; requesting the current role returns `200 OK` without incrementing `membership_version`; and a successful role change takes effect on the next request because incident authorization is re-derived from current membership and role at request time.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030
- **AC-180**: `DELETE /api/v1/incidents/{incident_id}/memberships/{user_id}` removes only that incident membership and returns `204 No Content`; deleting or demoting the last incident admin fails with `409` and `error.code = last_incident_admin`; self-removal or self-demotion succeeds only when another current incident admin remains; and removing membership from incident A leaves the same still-valid session usable for incident B when the user remains authorized there.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030

- **AC-181**: `DELETE /api/v1/records/{record_id}` accepts JSON `base_row_version`, `client_txn_id`, and optional `reason`; an `editor`, `reviewer`, or `admin` on the incident receives `200 OK` with `record_id`, `incident_id`, incremented `row_version`, `deleted=true`, non-null `deleted_at`, non-null `deleted_by_user_id`, and `change_set_id`; the record no longer appears in ordinary view queries; `GET /api/v1/records/{record_id}/history` still returns prior history plus a `soft_delete` entry; the collaboration stream emits `record_changed` with `change_kind='remove'`; a stale base version fails with `409 error.code='row_version_conflict'`; patching the deleted record fails with `409 error.code='record_deleted_use_restore'`; and replaying the same normalized delete request by the same actor with the same `(record_id, client_txn_id)` returns the originally committed success without creating a second mutation.
  - Verifies: REQ-01-057..REQ-01-088, REQ-02-210
- **AC-182**: `POST /api/v1/records/{record_id}/restore` accepts JSON `base_row_version`, `client_txn_id`, and optional `reason`; a `reviewer` or `admin` on the incident can restore a currently soft-deleted record and receives `200 OK` with `deleted=false`, `deleted_at=null`, `deleted_by_user_id=null`, incremented `row_version`, and a new `change_set_id`; the record becomes eligible again for ordinary view queries; history remains append-only and adds a `restore` entry rather than rewriting the prior delete; the collaboration stream emits `record_changed` with `change_kind='invalidate'` rather than a new insert-like change kind; a restore against a non-deleted record fails with `409 error.code='record_not_deleted'`; and if the implementation uses a short-lived destructive-operation lock and the record is already locked, the route fails with `409 error.code='record_locked'` and `error.retryable=true`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-02-210, REQ-03-101
- **AC-183**: For a soft-deleted record, `GET /api/v1/records/{record_id}/history` exposes the current tombstone `row_version`, or an equivalent current revision version token accepted by `POST /api/v1/records/{record_id}/restore`; a restore request that uses that current token succeeds when otherwise authorized, while a restore request that uses an older token fails with `409 error.code='row_version_conflict'`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-02-210
- **AC-215**: `GET /api/v1/records/{record_id}/history` returns `incident_id`, `record_id`, current `row_version`, `deleted`, and logical history `items[]`; each logical item includes `change_set_id`, `reversible`, `available_rollback_actions[]`, `history_entry_ref` only when the item maps to exactly one reversible mutation target, and `revision_no` only when whole-row restore is legal; `available_rollback_actions[]` contains only `history_entry`, `change_set`, and `row_restore`; and a caller lacking visibility receives `404`.
  - Verifies: REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-218, REQ-03-139..REQ-03-142
- **AC-216**: `POST /api/v1/records/{record_id}/rollback` with valid JSON `base_row_version`, `client_txn_id`, `target.kind='history_entry'`, and a visible `history_entry_ref` by a `reviewer` or `admin` can roll back one mistaken host link, tag assignment, mention resolution or dismissal, or evidence attach or detach from the selected row's history without reverting later unrelated edits on the same row; success returns `200 OK` with incremented `row_version`, echoed `target`, `target_change_set_id`, new `rollback_change_set_id`, and canonical `affected_record_ids[]`; and replaying the same normalized request by the same actor with the same `(record_id, client_txn_id)` returns the originally committed success without creating a second rollback `change_set`.
  - Verifies: REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-216, REQ-03-141..REQ-03-144
- **AC-217**: `POST /api/v1/records/{record_id}/rollback` with `target.kind='change_set'` and a merge `change_set_id` reverses the entire reversible `change_set` in reverse deterministic entry order and restores the pre-merge graph without any separate unmerge route; the same route with `target.kind='row_restore'` and `restore_to_revision_no` restores only row-backed fields to the selected revision and does not implicitly recreate or delete links, tags, mentions, indicator observations, or evidence associations; each success appends one new attributed rollback `change_set`, preserves prior history, increments `row_version` on every affected first-class record, and reaches subscribers only through ordinary `record_changed` events.
  - Verifies: REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-220, REQ-03-141..REQ-03-144
- **AC-218**: Malformed rollback request shape, unknown top-level request member, unknown `target.kind`, or a selector whose JSON type does not match the declared shape fails with `400 error.code='invalid_rollback_request'`; a rollback target not visible in the current record history fails with `404 error.code='rollback_target_not_found'`; a rollback against a currently soft-deleted record fails with `409 error.code='record_deleted_use_restore'`; later dependent changes, a non-individually-reversible history item, or a selector superseded by later reversals fail with `409 error.code='rollback_precondition_failed'` and `error.details.reason_code` equal to `target_not_reversible`, `entry_requires_change_set`, `dependent_later_changes`, or `stale_target`; stale `base_row_version` fails with `409 error.code='row_version_conflict'`; and an active destructive-operation lock fails with `409 error.code='record_locked'` and `error.retryable=true`.
  - Verifies: REQ-01-057..REQ-01-111, REQ-01-228..REQ-01-239, REQ-02-205..REQ-02-207, REQ-03-101,
    REQ-03-141..REQ-03-144
- **AC-184**: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` accepts `filters[]` entries using the Core 01 §3.3.4.1 wire shape for `eq`, `range`, `contains_any`, `contains_all`, `prefix`, and `full_text`; semantically identical client orderings normalize to the same `meta.query.filters[]`; and an unknown filter field, disallowed operator, duplicate `field_key`, empty `values[]`, empty text query, or malformed range fails with `400` and `error.code = invalid_view_query`.
  - Verifies: REQ-01-034..REQ-01-056, REQ-01-312..REQ-01-322, REQ-03-223..REQ-03-224
- **AC-185**: The Notes view exposes `note.full_text` as a stable synthetic filter key; a query using `field_key='note.full_text'` and `op='full_text'` performs the declared case-insensitive token search over `note.title` and `note.body` without requiring storage-specific search syntax in the request.
  - Verifies: REQ-01-034..REQ-01-056, REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071,
    REQ-03-223..REQ-03-224
- **AC-186**: `POST /api/v1/records/{survivor_record_id}/merge` accepts JSON `loser_record_id`, `survivor_base_row_version`, `loser_base_row_version`, `client_txn_id`, and optional `reason`; a `reviewer` or `admin` on the incident can merge two visible same-incident same-type active `host` records or two visible same-incident same-type active `identity` records; on success the route returns `200 OK` with `incident_id`, `record_type`, both record IDs, updated survivor and loser row versions, and `change_set_id`; the loser becomes historical with `merged_into_record_id` set to the survivor; active mentions, links, assessments, and tags are repointed or deterministically recreated in the same `change_set`; duplicate links and tags are deduplicated without losing history; and the collaboration stream removes the loser from ordinary active entity views while invalidating or patching the survivor and affected dependent rows.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-181..REQ-01-195, REQ-02-054..REQ-02-055, REQ-02-060..REQ-02-061,
    REQ-02-064..REQ-02-066, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249
- **AC-187**: The same route fails closed when the two IDs are identical, the records belong to different incidents, the record types differ, the record type is not `host` or `identity`, either record is soft-deleted or already merged away, the caller lacks `reviewer` or `admin` role, either supplied base row version is stale, or a short-lived destructive-operation lock is already held; stale versions fail with `409 error.code='row_version_conflict'`; an active lock fails with `409 error.code='record_locked'` and `error.retryable=true`; merge-precondition failures fail with `409 error.code='merge_precondition_failed'`; a caller lacking visibility receives `404`; and replaying the same normalized request by the same actor with the same `(survivor_record_id, loser_record_id, client_txn_id)` returns the originally committed success without creating a second merge.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-181..REQ-01-195, REQ-01-237, REQ-02-060..REQ-02-061,
    REQ-02-064..REQ-02-066, REQ-03-101, REQ-03-247..REQ-03-249

### 9.11 Incident Portability Extension Profile criteria

- **AC-164**: Whole-incident export produces a bundle whose logical layout, manifest, checksum file, structured JSON or NDJSON files, and blob paths satisfy Core 01 §12.3, and the bundle excludes projections, search indexes, sessions, presigned URLs, locks, client-local drafts, login-capable local users, deployment-admin flags, auth-binding state, memberships, permissions, deployment-local administrative audit history, password hashes, MFA secrets, external-provider configuration, and object-store credentials.
  - Verifies: REQ-01-425..REQ-01-442, REQ-04-044..REQ-04-047, REQ-04-065
- **AC-165**: Exporting an incident and importing that bundle into an empty deployment preserves the exported `incident_id`, `record_id`, `row_version`, change-set count, revision count, record-link count, entity-mention count, indicator-observation count, and blob hashes, and the imported incident opens normally after projection rebuild.
  - Verifies: REQ-01-425..REQ-01-426, REQ-01-439..REQ-01-442, REQ-01-447..REQ-01-450, REQ-04-044..REQ-04-047,
    REQ-04-065
- **AC-166**: If one required structured file or one required blob is missing, if any required checksum is corrupted, if `incident_id` already exists, or if a required capability is unsupported, import fails closed before the incident becomes visible and leaves no partially visible incident state.
  - Verifies: REQ-01-425..REQ-01-430, REQ-01-433..REQ-01-438, REQ-01-447..REQ-01-451, REQ-04-044..REQ-04-047,
    REQ-04-065
- **AC-167**: If a bundle contains optional embedded `snapshots` or `reference_packs` sections that the target deployment does not support, the importer ignores or degrades only those optional sections unless the relevant capability is listed in `required_capabilities[]`, and core incident import still succeeds.
  - Verifies: REQ-01-425..REQ-01-426, REQ-01-431..REQ-01-432, REQ-01-443..REQ-01-450, REQ-04-044..REQ-04-047,
    REQ-04-065
- **AC-168**: Historical actors in `actors.ndjson` that do not map to an existing local user become inert imported actors or equivalent historical actor descriptors, are not login-capable, are not automatically added to incident membership, and still remain visible as historical attribution in imported history.
  - Verifies: REQ-01-425..REQ-01-426, REQ-01-443..REQ-01-446, REQ-04-044..REQ-04-047, REQ-04-065
- **AC-169**: Whole-incident export and import run as background jobs, show progress and cancellation without blocking grid editing, stage bundle contents only under the configured temporary-work root, and in flyaway or disconnected deployments keep emitted bundles and staged extracts on encrypted storage.
  - Verifies: REQ-01-425..REQ-01-430, REQ-01-447..REQ-01-450, REQ-01-452..REQ-01-456, REQ-04-044..REQ-04-047,
    REQ-04-054..REQ-04-055, REQ-04-058..REQ-04-059, REQ-04-065

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

**REQ-04-065**
A conformant Cartulary deployment MUST prefer operational simplicity over distributed architectural purity.
Profiles: base
Verified by: AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-231

The smallest useful deployment is intentionally not the absolute minimum number of containers. It is the minimum that keeps binary evidence handling sane and portable while preserving collaboration, authentication, and auditable source-of-truth behavior.
