# Cartulary Normative Core 04: Security, Deployment, and Conformance

## 1. Authentication model

### 1.1 Base authentication

**REQ-04-001**
The base profile MUST support:

- local user accounts stored in Postgres,
- password hashing with Argon2id,
- TOTP MFA.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231, AC-334, AC-336, AC-337, AC-341

**REQ-04-083**
The base profile MUST include a bounded public credential-management contract for local accounts. The only base-profile recovery model is `admin_assisted`. Base-profile conformance excludes email-link recovery, SMS recovery, backup codes, provider-mediated recovery, and self-service factor disablement.
Profiles: base
Verified by: AC-334..AC-342

**REQ-04-084**
The base-profile TOTP contract MUST use `SHA1`, `digits=6`, `period_seconds=30`, trusted server time, and an acceptance window limited to the current time step plus at most one adjacent step on either side. Shared secrets MUST be generated with a CSPRNG and provide at least `160 bits` of secret entropy. `secret_base32` returned on the public begin route MUST be uppercase Base32 without separators or padding. TOTP secret material MUST remain deployment-local state wrapped or encrypted at rest and MUST NOT be emitted outside `POST /api/v1/auth/mfa/totp/begin`. `bootstrap_token` is not a session, is not a bearer-token family for ordinary API use, and MUST NOT be redeemable on `/ws/v1/*`.
Profiles: base
Verified by: AC-334..AC-339

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
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-338, AC-340..AC-342

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
Qualifying activity MAY slide `idle_expires_at`, but MUST NOT extend the session beyond `absolute_expires_at`. A qualifying-activity update MUST NOT move `last_qualifying_activity_at`, `idle_expires_at`, or `session_expires_at` backward.
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
When a new login would exceed the concurrent-session cap, the server MUST revoke the least-recently-used non-current session before issuing the new session and MUST record an attributed audit event with stable reason code `concurrency_limit`. Least-recently-used ordering MUST sort by `last_qualifying_activity_at ASC`, then `authenticated_at ASC`, then stable session identifier ASC.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-015**
`POST /api/v1/auth/logout` MUST revoke only the current session immediately.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-016**
Self-service password change, self-service TOTP replacement, administrator password reset, administrator TOTP reset, account disablement, or an explicit deployment-admin revoke-all action MUST revoke all active sessions for that user immediately.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-04-095**
Successful enterprise-auth binding rotate or retire actions MUST revoke all active sessions for the target user immediately. Successful enterprise-auth binding create MAY leave current sessions unchanged.
Profiles: enterprise_authentication
Verified by: AC-350, AC-351, AC-352

**REQ-04-017**
Loss of incident membership for one subscribed incident MUST terminate the affected incident WebSocket subscription and MUST cause future incident-scoped authorization checks for that incident to fail closed, but it MUST NOT by itself require revocation of otherwise valid account sessions for other authorized incidents.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

### 1.2 Enterprise Authentication Extension Profile

**REQ-04-018**
If the implementation claims the **Enterprise Authentication Extension Profile**, it MUST support provider-backed identities through an `auth_providers` and `auth_identities` model equivalent to the source artifact. The current profile standardizes OIDC authorization-code flow with PKCE `S256` and `nonce`, plus SAML SP-initiated flow only.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235, AC-288, AC-289, AC-290, AC-291, AC-293

OIDC is the preferred enterprise path. SAML is the secondary enterprise path when required by the environment.

**REQ-04-019**
External identities MUST map to the same internal user identity used for attribution so that audit semantics remain unchanged. `provider_subject` MUST be the authoritative external identifier. Successful provider authentication MUST NOT auto-create a local user, auto-create incident membership, auto-create an `auth_identity`, or map provider group claims into incident roles.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235, AC-292, AC-293

**REQ-04-020**
When the Enterprise Authentication Extension Profile is implemented, successful provider authentication MUST terminate into the same server-managed session contract used by the base profile so the remaining API surface stays provider-agnostic. Completion routes MUST be single-use and replay-resistant, and provider configuration, raw assertions, and any provider tokens MUST remain server-side deployment-local state.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235, AC-290, AC-291, AC-293

**REQ-04-093**
The current profile MUST use the deployment-admin binding-management surface defined by Core 01 §20 as the only normative runtime mechanism for creating, rotating, and retiring enterprise-auth bindings. A startup-only linkage artifact MAY exist only as helper tooling that produces the same binding state. It MUST NOT be a second authoritative runtime mechanism. The current profile defines no self-service account linking, no JIT provisioning, no SCIM provisioning, and no enterprise-only user-creation path.
Profiles: enterprise_authentication
Verified by: AC-348, AC-352

Reference-provider compatibility targets for the current profile are Okta, Delinea, and Microsoft Entra ID.

## 2. Authorization model

**REQ-04-021**
The base profile MUST use exactly these incident-level roles:

- `viewer`,
- `editor`,
- `reviewer`,
- `admin`.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-022**
Record access MUST inherit from incident access in the base profile. Party rows and `*_party_id` references are incident data under this same rule.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-280

**REQ-04-023**
API routes, preview or download handle issuance and redemption, job polling, job cancellation, and WebSocket incident subscriptions MUST re-derive authorization from the caller's current scope membership and role at request time. Incident-scoped jobs MUST use current incident membership and role. Deployment-scoped jobs MUST use the owning route family's deployment-scoped authorization contract.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-254, AC-255, AC-257, AC-260, AC-261

**REQ-04-024**
`party`, `task_request`, `decision`, and coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MUST inherit the same incident-level authorization model. The base profile MUST NOT introduce record-specific ACLs or hidden sub-workspaces for these objects.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-280

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
The base profile MUST separate deployment-local account administration from incident-scoped data authorization. A conformant deployment MUST define one narrow deployment-scoped capability named `deployment_admin`. This capability authorizes only deployment-local user-account inspection and administration, including local-user creation, user patching, and any explicit revoke-all action the deployment exposes. The first holder of this capability MUST be provisioned only through the deployment-local bootstrap-admin manifest mechanism defined by Core 01 §3.3.5.1 and Core 04 §12.3.2, not through a public or incident-scoped surface.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-343, AC-344, AC-345, AC-346

**REQ-04-029**
Holding `deployment_admin` MUST NOT by itself grant incident read, write, preview, download, export, incident-scoped job, or incident WebSocket access. A caller who is `deployment_admin=true` but lacks current membership in incident X MUST have no incident-data or incident-scoped job access to incident X until granted ordinary incident membership.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-261

**REQ-04-106**
Any built-in backup, restore, or restore-verification control surface MUST remain deployment-local and operator-facing. In the base profile, such a surface MUST require `deployment_admin` and MUST NOT be incident-scoped.
Profiles: base
Verified by: AC-402

**REQ-04-030**
Incident membership create, role-change, and delete routes remain incident-scoped authorization decisions. In the base profile, those routes MUST require current incident role `admin`; `deployment_admin` alone MUST NOT bypass that requirement.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231

**REQ-04-105**
`GET /api/v1/extensions` is deployment-scoped discovery. Any authenticated session MAY read it. It is not incident-scoped and does not require `deployment_admin`. The route MUST expose only extension-claim state and reserved family roots and MUST NOT expose provider secrets, provider assertions, provider claim maps, or other deployment-local secret-bearing state.
Profiles: base
Verified by: AC-370, AC-371

**REQ-04-085**
Only `deployment_admin` may call the deployment-local credential action routes `POST /api/v1/users/{user_id}/password/reset`, `POST /api/v1/users/{user_id}/mfa/totp/reset`, and `POST /api/v1/users/{user_id}/sessions/revoke-all`. Incident membership or incident role `admin` alone MUST NOT authorize those routes.
Profiles: base
Verified by: AC-340..AC-342

**REQ-04-094**
Only `deployment_admin` may call `POST /api/v1/users/{user_id}/auth-bindings`, `POST /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}/rotate`, and `DELETE /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}`. Incident membership or incident role `admin` alone MUST NOT authorize those routes.
Profiles: enterprise_authentication
Verified by: AC-352

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
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233, AC-306

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
Verified by: AC-231, AC-407

**REQ-04-037**
Every mutation MUST record, at minimum:

- actor user identifier,
- timestamp,
- mutation source such as `ui`, `import`, or `rollback`,
- before/after values for row-field edits, or target-specific reversible mutation data for the exact Core 02 §14.3 target families of record links, record tags, entity mentions, indicator observations, indicator lifecycle intervals, compromise assessments, evidence associations, and merge and repoint fan-out, at the required history granularity.
Profiles: base
Verified by: AC-231, AC-408

**REQ-04-038**
User-account, incident-membership, and successful first-deployment-admin bootstrap administration mutations MUST be captured with the same minimum actor, timestamp, source, and before/after fidelity even when they are stored outside incident `change_set` or `record_revisions` rows. Deployment-local administrative audit records and any deployment-local idempotency substrate for user-account administration MUST NOT retain `initial_password` or any equivalent secret-bearing bootstrap or auth input in cleartext; they MAY retain only redacted placeholders or non-reversible derived metadata. Those administrative audit records are deployment-local state and MUST NOT be serialized into whole-incident portability bundles.
Profiles: base, incident_portability
Verified by: AC-175, AC-231, AC-236, AC-343, AC-344, AC-346, AC-409

**REQ-04-086**
Password change, password reset, TOTP begin, TOTP complete, TOTP reset, and explicit revoke-all session actions MUST be auditable deployment-local administrative events. Deployment-local audit or idempotency state for these routes MUST NOT retain `current_password`, `new_password`, `secret_base32`, `otpauth_uri`, or raw `bootstrap_token` in cleartext.
Profiles: base
Verified by: AC-336..AC-338, AC-340..AC-342

**REQ-04-096**
Enterprise-auth binding create, rotate, and retire actions MUST be auditable deployment-local administrative events. Those audit records MUST preserve the target stable `user_id`, `provider_key`, and any old or new `provider_subject` values needed to reconstruct binding lifecycle, but they MUST NOT retain any member of this closed forbidden set: provider assertions, provider tokens, raw SAML responses, ID tokens, and access tokens.
Profiles: enterprise_authentication
Verified by: AC-352

**REQ-04-039**
Security choices MUST NOT add friction to the primary capture path. MFA belongs at login and session establishment, not during routine row creation or evidence preview.
Profiles: base
Verified by: AC-231

## 4. Trust boundaries

**REQ-04-109**
Test-only runtime-control routes, including `/api/v1/test/*` and `/ws/v1/test/*`, are non-production harness surfaces. They MUST be unavailable by default and MUST fail closed unless the runtime is explicitly marked as harness-owned, test routes are explicitly enabled, and a harness-generated opaque test-route token with at least `128 bits` of entropy is configured. Ordinary session cookies, bearer sessions, bootstrap tokens, incident roles, and `deployment_admin` MUST NOT authorize test-only runtime-control routes. When a harness-owned API or browser origin is configured, test-only runtime-control routes MUST reject requests outside the configured host and origin boundary before executing any mutation.
Profiles: base
Verified by: AC-413

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
- content drawn directly from `task_request`, `decision`, structured finding rows where `finding.kind='hypothesis'`, `comm_log`, `handoff`, `status_review`, or `lesson` records MUST NOT be treated as inherently releasable; any `external_release` use MUST flow through the snapshot, redaction, and curation path,
- reenactment outputs MUST be marked `generated_presentation=true` and MUST NOT be released as `external_release`,
- approval and redaction checks MUST complete successfully before an `external_release` artifact is published.
Profiles: snapshot_reporting, incident_portability
Verified by: AC-031, AC-057, AC-059, AC-060, AC-061, AC-062, AC-091, AC-113, AC-114, AC-115, AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-233, AC-236, AC-333

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
- an object-storage access pattern,
- a backup or restore mechanism.
Profiles: base
Verified by: AC-048, AC-231

**REQ-04-051**
At minimum, the threat model MUST cover the following assets and abuse cases:
Profiles: base
Verified by: AC-048, AC-231

| STRIDE class | Minimum project-specific scope | Required control direction |
| --- | --- | --- |
| Spoofing | analyst sessions, provider-backed identities, explicit system-process actors, object-store upload/download capabilities | authenticated sessions, stable internal user mapping, explicit system actors, short-lived operation-scoped object access |
| Tampering | incident records, revisions, object blobs, backup artifacts, restore-verification extracts, reference packs, snapshots, exports | row-versioned writes, immutable change sets, blob hashes, fail-closed integrity verification, immutable published snapshots |
| Repudiation | edits, imports, rollbacks, pack activation, export generation, evidence lifecycle actions | attributed append-only history with actor, timestamp, source, and reversible mutation detail |
| Information disclosure | evidence blobs, backup artifacts, restore-verification extracts, exports, previews, secrets, portable runtime roots | incident-scoped authorization, secret isolation, untrusted-content rendering rules, self-contained outputs, encrypted flyaway storage |
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
- **CWE-22 / CWE-73**: User-supplied filenames, archive entry names, import paths, and blob-create `filename_hint` values MUST be treated as metadata, not authority. The system MUST assign storage keys, MUST reject absolute paths and parent traversal, and MUST extract archives only inside a staging root that cannot escape the declared runtime roots. `filename_hint` MUST NOT determine object-store key paths, authorization decisions, or portability layout.
- **CWE-353**: Reference packs and any incident import bundle format, when implemented, MUST fail closed on checksum mismatch, signature mismatch, incomplete download, or missing required integrity metadata.
- **CWE-434**: Evidence, reference-pack, and workbook-import uploads MUST be treated as hostile content. The application unit MUST NOT execute uploaded content, workbook formulas, macros or VBA, workbook automation, or external links during import or preview. Blob-create `content_type_hint` values are metadata only; they MUST NOT by themselves determine preview allowlisting, active-content classification, or release posture. Preview issuance MUST succeed only for allowlisted non-executing `preview_kind` values derived from current blob or evidence state and server-observed or otherwise validated media metadata, and preview or download redemption MUST fail closed when current blob or evidence state is pending, failed, missing, quarantined, or inconsistent. Active-content types MUST remain download-only or isolated from the main application origin unless a dedicated isolated analysis path is explicitly implemented.
- **CWE-639**: Every mutation, preview or download handle issuance, preview or download handle redemption, and object-store URL issuance MUST re-derive authorization server-side from the target object's owning incident and the caller's current membership and role. Client-supplied incident identifiers, ownership metadata, or role claims MUST NOT determine access.
- **CWE-352**: State-changing HTTP routes authenticated by cookie MUST require CSRF protection that fails closed. WebSocket upgrades and any incident subscription step MUST verify the authenticated session and incident authorization before joining an incident-scoped stream. Cookie-authenticated browser WebSocket connections MUST reject untrusted `Origin` values before the socket joins an incident-scoped stream.
- **CWE-312**: Deployments intended for portable or flyaway use MUST keep database storage, object storage, backup storage, reference-pack storage, temporary work files that carry incident data, restore-verification extracts, and export outputs on encrypted storage. Unencrypted removable media or unencrypted portable roots are non-conformant for flyaway handling.
Profiles: base, import, reference_pack
Verified by: AC-049, AC-050, AC-051, AC-052, AC-053, AC-054, AC-055, AC-130, AC-131, AC-231, AC-232, AC-234, AC-252, AC-253, AC-254, AC-255

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
Any on-prem substitution of centrally managed Postgres or object-storage services for the recommended disconnected containers MUST preserve the topology invariants owned by Core 01 REQ-01-001 through REQ-01-003. This section defines deployment-profile consequences only and MUST NOT be interpreted as defining an alternate deployable-unit or authoritative-service boundary.
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
- backup storage,
- reference-pack storage,
- temporary work files,
- export outputs.
Profiles: base, reference_pack
Verified by: AC-051, AC-055, AC-169, AC-231, AC-234, AC-294, AC-295, AC-297, AC-403

**REQ-04-059**
The application MUST NOT rely on source-tree-relative paths for runtime assets or generated artifacts.
Profiles: base
Verified by: AC-051, AC-055, AC-169, AC-231, AC-296

Core 04 §12 owns the operator-facing deployment configuration artifact, discovery precedence, binding keys, default disconnected-layout locations, validation contract, and fail-closed startup behavior for these runtime roots.

## 7. Container boundary

**REQ-04-060**
Container-boundary conformance MUST be evaluated against the topology invariants owned by Core 01 REQ-01-001 through REQ-01-003. This section defines no alternate deployable-unit decomposition and no alternate authoritative-service boundary.
Profiles: base
Verified by: AC-231


## 8. Required and optional supporting services

### 8.1 Required services

**REQ-04-061**
A conformant deployment MUST provide the owner-defined base topology components required by Core 01 REQ-01-002. This section defines no alternate required-service inventory.
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

For timed or fixture-sensitive implementation criteria in this section:

- `first useful viewport` means the first rendered visible row window for the active sort, filter, and grouping state with stable `record_id` binding and working keyboard navigation, even if off-screen rows continue loading.
- `stable viewport` means the visible row window and result ordering match the final deterministic order for the active sort, filter, and grouping state and no further reorder occurs without new user or server input.
- `metadata shell` means row fields and evidence metadata needed to inspect the selected record, including counts, filenames or media-type labels, attachment state, and preview handles, but excluding binary preview bytes or full blob download.

### 9.0 Profile claim manifests

The manifests below define implementation claim boundaries without restating requirement prose. Each manifest selects implementation requirements through the `Profiles:` trailers carried by Core 00 through Core 04 and pairs that selector with the acceptance criteria that complete the claim. Appendix F expands every selector into explicit navigation tables.

#### 9.0.1 Base claim manifest

A Base claim selects every requirement block tagged `base`.

Definition of Done:

- requirement selector: `profile:base`
- required acceptance criteria: `AC-001..AC-026`, `AC-037..AC-055`, `AC-068..AC-070`, `AC-072..AC-090`, `AC-097..AC-103`, `AC-107..AC-112`, `AC-116..AC-163`, `AC-170..AC-231`, `AC-238..AC-261`, `AC-277..AC-287`, `AC-294..AC-304`, `AC-311..AC-322`, `AC-329..AC-331`, `AC-334..AC-347`, `AC-353..AC-354`, `AC-359..AC-368`, `AC-370..AC-371`, `AC-372..AC-375`, `AC-376..AC-385`, `AC-387..AC-392`, `AC-394..AC-408`, `AC-410`, `AC-411`, `AC-412`, `AC-413`
- **AC-231**: A Base claim is conformant only when every requirement selected by `profile:base` is implemented and every acceptance criterion listed in this manifest passes.
  - Verifies: `profile:base`

AC-231 through AC-236 are aggregate profile-claim gates, and PC-006 is a companion claim-publication gate. They MUST NOT be the sole verifier for any requirement that specifies substantive runtime behavior. Every such requirement family MUST also be covered by at least one non-aggregate acceptance criterion or companion criterion that exercises that behavior directly.

#### 9.0.2 Import claim manifest

An Import claim requires a passing Base claim and every requirement block tagged `import`, including the import-boundary and import-provenance requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:import`
- additional acceptance criteria: `AC-027..AC-029`, `AC-063..AC-067`, `AC-232`, `AC-262..AC-265`, `AC-323..AC-325`, `AC-393`
- **AC-232**: An Import claim is conformant only when a Base claim passes, every requirement selected by `profile:import` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:import`

#### 9.0.3 Snapshot and Reporting claim manifest

A Snapshot and Reporting claim requires a passing Base claim and every requirement block tagged `snapshot_reporting`, including the release-gate requirements in §2.1 and any snapshot or rendering requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:snapshot_reporting`
- additional acceptance criteria: `AC-030..AC-032`, `AC-056..AC-062`, `AC-071`, `AC-091`, `AC-104..AC-106`, `AC-113..AC-115`, `AC-233`, `AC-266..AC-269`, `AC-305..AC-307`, `AC-333`
- **AC-233**: A Snapshot and Reporting claim is conformant only when a Base claim passes, every requirement selected by `profile:snapshot_reporting` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:snapshot_reporting`

#### 9.0.4 Reference Pack claim manifest

A Reference Pack claim requires a passing Base claim and every requirement block tagged `reference_pack`, including the disconnected-pack lifecycle and attestation requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:reference_pack`
- additional acceptance criteria: `AC-033..AC-035`, `AC-092..AC-096`, `AC-234`, `AC-270..AC-272`, `AC-308..AC-310`, `AC-326`, `AC-369`
- **AC-234**: A Reference Pack claim is conformant only when a Base claim passes, every requirement selected by `profile:reference_pack` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:reference_pack`

#### 9.0.5 Enterprise Authentication claim manifest

An Enterprise Authentication claim requires a passing Base claim and every requirement block tagged `enterprise_authentication`, including the provider-identity requirements in §1.2.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:enterprise_authentication`
- additional acceptance criteria: `AC-036`, `AC-235`, `AC-288..AC-293`, `AC-348..AC-352`
- **AC-235**: An Enterprise Authentication claim is conformant only when a Base claim passes, every requirement selected by `profile:enterprise_authentication` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:enterprise_authentication`

#### 9.0.6 Incident Portability claim manifest

An Incident Portability claim requires a passing Base claim and every requirement block tagged `incident_portability`, including the logical-bundle and import-failure requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:incident_portability`
- additional acceptance criteria: `AC-164..AC-169`, `AC-236`, `AC-273..AC-276`, `AC-327..AC-328`, `AC-332`, `AC-386`, `AC-409`
- **AC-236**: An Incident Portability claim is conformant only when a Base claim passes, every requirement selected by `profile:incident_portability` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:incident_portability`

### 9.1 Base Profile criteria

- **AC-001**: An analyst can create a new timeline row by typing into a blank grid cell and pressing Enter, with no modal and no required form.
  - Verifies: REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-111..REQ-03-115
- **AC-002**: A timeline row can be persisted with only one non-empty user-entered value or only an attached screenshot; `recorded_at`, author, and row identity are system-generated.
  - Verifies: REQ-03-001..REQ-03-003, REQ-03-111..REQ-03-115

The timed or fixture-sensitive criteria below define observable implementation outcomes only. Claim-bearing benchmark-publication policy is owned by Core 05.
- **AC-003**: Pasting a 20-row by 5-column block from Excel into the Timeline sheet creates rows and maps visible columns in under 2 seconds.
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
- **AC-011**: Rolling back a mistaken change or restoring a whole row creates a new attributed revision and updates the visible row in under 2 seconds.
  - Verifies: REQ-03-141..REQ-03-142
- **AC-012**: Whole-row restore and whole-change-set rollback are available as explicit secondary actions for multi-target or destructive changes; arbitrary field-picker rollback from historical snapshots is not required in the base profile.
  - Verifies: REQ-03-141..REQ-03-142
- **AC-013**: Re-sorting, re-filtering, or re-grouping a sheet does not cause a pending edit to target a different underlying record; all mutations are sent using `record_id`, `base_row_version`, and changed fields only.
  - Verifies: REQ-01-350, REQ-03-033..REQ-03-035, REQ-03-086, REQ-03-223..REQ-03-224
- **AC-014**: Renaming a visible column header or tab label does not change filter semantics, write-back behavior, or export semantics for a built-in or system view; those behaviors are bound to `view_schema_id`.
  - Verifies: REQ-03-223..REQ-03-224

- **AC-116**: `GET /api/v1/view-schemas` returns only standardized current-profile `view_schema_id` values in `data.view_schemas[]`, ordered by `view_schema_id asc`. Each list entry and each `GET /api/v1/view-schemas/{view_schema_id}` singleton response is a `view_schema_resource_v1` containing required members `view_schema_id`, `surface_kind`, `title`, `source_record_types`, `technical_fields`, `required_reference_pack_keys`, `default_sort`, `sort_fields`, `filter_fields`, `synthetic_filter_predicates`, `grouping_fields`, and `fields`. `technical_fields` is exactly `["record_id","row_version"]`. The public discovery resource does not expose `base_projection`, storage-table names, or internal write targets. With no optional reference packs installed or activated, the returned set is exactly these fourteen pack-independent ids: `cartulary.view.assessments.v1`, `cartulary.view.comm_log.v1`, `cartulary.view.decisions.v1`, `cartulary.view.evidence.v1`, `cartulary.view.handoff.v1`, `cartulary.view.hosts.v1`, `cartulary.view.identities.v1`, `cartulary.view.indicators.v1`, `cartulary.view.lesson.v1`, `cartulary.view.notes.v1`, `cartulary.view.parties.v1`, `cartulary.view.status_review.v1`, `cartulary.view.task_requests.v1`, and `cartulary.view.timeline.v1`. With optional reference packs installed and active, the returned current-profile set remains limited to those same fourteen ids plus, when implemented, `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1`. ATT&CK, D3FEND, VERIS, and other framework-specific overlay `view_schema_id` values do not appear in the current profile. Omitting `limit` yields `meta.paging.limit=100`; terminal pages use `meta.paging.has_more=false` and `meta.paging.next_cursor=null`; invalid `limit` values or aliases such as `page`, `offset`, `page_size`, and `block_size` fail closed with `400 error.code='invalid_pagination_request'`; cursor replay against a different bound route contract fails closed; and `GET /api/v1/view-schemas/{view_schema_id}` rejects pagination members with `error.details.reason_code='pagination_not_supported'`.
  - Verifies: REQ-00-014, REQ-01-240..REQ-01-242, REQ-01-285..REQ-01-296, REQ-01-307..REQ-01-310, REQ-01-499, REQ-01-503..REQ-01-509, REQ-03-004, REQ-03-242..REQ-03-246

- **AC-410**: Core 02 §10.4.4A defines a closed tagged-variant registry for artifact-backed notes, structured coordination artifacts, and structured findings. The registry contains exactly six rows in canonical order `note`, `comm_log`, `handoff`, `status_review`, `lesson`, and `finding`. Each row declares exact `durable_discriminator`, `public_surface_ref`, `identifier_field`, `required_structured_state[]`, `optional_structured_state[]`, and `lifecycle_notes`. The `note` and `finding` rows use `identifier_field=record_id` rather than subtype-local identifiers. `finding` is the only row with `subkind_dimension='finding.kind'` and it explicitly carries the current-profile hypothesis representation. There is no separate `hypothesis` row and no `cartulary.view.hypotheses.v1`. Core 01 §7.4 and Core 03 §2 contain explicit cross-references that preserve Core 01 ownership of exhaustive field contracts and create-time policy while using the Core 02 registry as the closed family inventory.
  - Verifies: REQ-01-309, REQ-02-134, REQ-02-250..REQ-02-253, REQ-03-004, REQ-03-011

- **AC-411**: Core 01 §7.4 contains exactly one authoritative cross-layer workbook-surface mapping table for the current profile. That table contains exactly seventeen rows in canonical order: the fourteen pack-independent registry entries from REQ-01-307 followed by `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1` in the order named by REQ-01-308. Each row declares exact `view_schema_id`, `surface_kind`, `source_record_types`, canonical source discriminator or filter, `surface_status`, and `required_reference_pack_keys`; every row binds canonical workbook identity to `sheet_ref={ "kind": "view_schema", "id": <view_schema_id> }`; `surface_kind` uses only `built_in_sheet` or `system_view`; the table introduces no pack-dependent framework overlays and no new runtime discovery members; and Core 03 §2 contains a non-owner cross-reference back to that table rather than a duplicate row-mapping table.
  - Verifies: REQ-01-579, REQ-03-011

- **AC-117**: The structured contract for each of those fourteen mandatory base-profile view schemas, plus any currently exposed standardized optional artifact-backed workbook surface, includes an ordered `fields[]` list, an ordered `default_sort` tuple, explicit `sort_fields`, `filter_fields`, `synthetic_filter_predicates[]`, and the declared `grouping_fields` whitelist when grouping is supported; the final default-sort tiebreaker is `record_id`. `fields[]` are emitted in authoritative field-registry order. Set-like members, including filter `values[]` and canonicalized `full_text` token sets rendered through canonical `arg.query`, use canonical ascending order. `filter_fields` contains only keys also present in `fields[].field_key`, and filter-only synthetic predicate keys appear only in `synthetic_filter_predicates[]`; for Notes, `note.full_text` is exposed as a synthetic filter predicate rather than as a field entry.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-310, REQ-01-499..REQ-01-500, REQ-01-503..REQ-01-509, REQ-03-242..REQ-03-246

- **AC-118**: Every `fields[]` entry in those fourteen mandatory base-profile view schemas, plus any currently exposed standardized optional artifact-backed workbook surface, is `view_field_entry_v1` with required members `field_key`, `label`, `default_hidden`, `sortable`, `header_sort_field_key`, `filter_ops`, `groupable`, `read_kind`, `write_kind`, `conflict_resolution_class`, `entity_binding_mode`, `string_contract_id`, `direct_scalar_contract_id`, `direct_reference_contract_id`, `clearable`, and `enum_values`; inapplicable contract members are explicit `null`; `filter_ops=[]` means not filterable; and closed-vocabulary fields expose `enum_values` in declared token order. Every writable field declares a stable `field_key` and `conflict_resolution_class`; every entity-bearing writable field also declares `entity_binding_mode`; every human-authored writable string field or writable string-bearing action member closed by Core 01 §18 declares the correct `string_contract_id` binding; every writable direct temporal scalar field closed by Core 01 §18A declares the correct `direct_scalar_contract_id` binding and explicit clearable flag; required scalar fields reject explicit `null` and any contract-defined normalized-empty input; clearable optional scalar text fields and clearable optional direct temporal scalar fields apply omission-versus-`null` semantics exactly as declared; `collection_review` fields reject raw arrays and raw `null` on patch in favor of `collection_actions_v1`; and create-only identifiers remain immutable after first commit. Changing only a surface `title` or field `label` is non-breaking, but changing field membership, field meaning, writeability, `conflict_resolution_class`, `entity_binding_mode`, `sort_fields`, `filter_fields`, `synthetic_filter_predicates`, or `grouping_fields` without a new `view_schema_id` is non-conformant.
  - Verifies: REQ-00-014, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-310, REQ-01-323..REQ-01-341, REQ-01-487..REQ-01-509, REQ-02-028..REQ-02-029, REQ-02-202..REQ-02-204, REQ-03-052..REQ-03-053, REQ-03-242..REQ-03-246
- **AC-119**: In the Timeline schema, `timeline.summary`, `timeline.details`, and `timeline.source_text` are distinct field keys with distinct write targets. Editing one does not overwrite either of the other two fields.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-322, REQ-03-236..REQ-03-241
- **AC-120**: A write routed to a field that the active base-profile view schema declares read-only or derived fails closed, does not mutate authoritative source state, and does not persist a misleading projection update.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-03-236..REQ-03-241

- **AC-191**: A Timeline row created by blank-row entry, by a paste action that creates a new row, or by a screenshot-only create persists `capture_state='rough'` on first commit; `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` for `cartulary.view.timeline.v1` accepts a request whose only top-level member is `client_txn_id`, returns `201 Created` on first success, and creates exactly one Timeline row eligible for later screenshot or rough-capture enrichment.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241
- **AC-192**: A client attempt to supply `timeline.capture_state` in Timeline row creation or in `PATCH /api/v1/records/{record_id}`, or to supply any other create-forbidden or unknown top-level member on Timeline row create, or to send a non-object or otherwise malformed Timeline row-create body, fails with `400 error.code='invalid_mutation_payload'`, does not mutate authoritative source state, leaves no partial row, and does not persist a misleading projection update.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110, REQ-03-236..REQ-03-241
- **AC-193**: The first later `capture-state-material` mutation to a `rough` Timeline row, including an edit to `timeline.occurred_at`, `timeline.summary`, `timeline.details`, or `timeline.source_text`, or the first host-ref, identity-ref, evidence-attach or evidence-detach, or row-anchored indicator mutation, commits one visible `change_set` that also sets `capture_state='enriched'`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241
- **AC-194**: `POST /api/v1/records/{record_id}/mark-reviewed` by a `reviewer` or `admin` against a non-deleted Timeline row whose current `capture_state` is `rough` or `enriched` returns `200 OK`, increments `row_version`, appends a new `change_set`, and sets `capture_state='reviewed'`; when optional `reason` is supplied it is normalized using `reason_note_v1`, and omission, explicit `null`, and normalized-empty reason compare equal and persist as `null`; the same route by an `editor` fails with `403`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-01-487..REQ-01-488, REQ-01-496, REQ-03-102..REQ-03-110
- **AC-195**: A later `capture-state-material` mutation to a `reviewed` Timeline row commits one visible `change_set` that sets `capture_state='enriched'`; a tag-only change to the same row leaves `capture_state='reviewed'`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-196**: `POST /api/v1/records/{record_id}/supersede` by a `reviewer` or `admin` with a required `reason` bound to `reason_note_v1` against a non-deleted Timeline row whose current `capture_state` is `rough`, `enriched`, or `reviewed` returns `200 OK`, increments `row_version`, appends a new `change_set`, and sets `capture_state='superseded'`; when no replacement row is supplied, the success payload returns `replacement_record_id=null`; when a valid `replacement_record_id` is supplied, the same committed action also persists exactly one active `record_links` row with `link_type='supersedes'`, `src_record_id=<replacement Timeline row>`, and `dst_record_id=<superseded Timeline row>`, returns the committed `replacement_record_id`, and a subsequent Timeline row query surfaces hidden `timeline.replacement_record_id` with that same value; LF canonicalization applies before validation and idempotency comparison, normalized-empty input is rejected, disallowed control characters are rejected, and values longer than 4096 Unicode scalar values fail closed.
  - Verifies: REQ-01-086, REQ-01-311..REQ-01-312, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-169, REQ-02-175, REQ-02-181, REQ-03-102..REQ-03-106
- **AC-197**: After a Timeline row reaches `superseded`, ordinary grid edits, `PATCH /api/v1/records/{record_id}` field mutations, and `POST /api/v1/records/{record_id}/mark-reviewed` fail closed; leaving `superseded` requires rollback of the superseding change.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-198**: Removing all current links or evidence from an already `enriched` Timeline row does not downgrade it to `rough`, and `timeline.has_unresolved_mentions` is `true` if and only if the current row still has at least one non-deleted unresolved mention, regardless of `capture_state`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-199**: Replaying the same normalized `POST /api/v1/records/{record_id}/mark-reviewed` or `POST /api/v1/records/{record_id}/supersede` request by the same authenticated actor with the same `(record_id, client_txn_id)` returns the originally committed success and does not create a second lifecycle transition.
  - Verifies: REQ-03-102..REQ-03-110
- **AC-329**: A supersede request whose `replacement_record_id` equals the target row, points to a non-Timeline record, another-incident record, a soft-deleted or otherwise non-visible row, a replacement Timeline row whose current `capture_state` is `superseded`, or a target Timeline row that already has another active incoming Timeline `supersedes` link fails closed with `409`, `error.code='illegal_transition'`, and the applicable `error.details.violated_guards[]` entries drawn only from `replacement_must_be_different_timeline_record`, `replacement_must_be_visible_active_same_incident_timeline_record`, `replacement_must_not_be_superseded`, and `target_must_not_have_active_replacement`; no partial `capture_state` change or replacement link commits.
  - Verifies: REQ-01-086..REQ-01-087, REQ-02-168, REQ-03-106
- **AC-330**: Exact replay of a committed supersede-with-replacement request by the same authenticated actor with the same `(record_id, client_txn_id)` returns the original committed success and does not create a second replacement link, `change_set`, or lifecycle transition; reusing that same key for a different normalized `replacement_record_id` or normalized `reason` fails with `409 error.code='client_txn_conflict'` before stale `base_row_version` evaluation.
  - Verifies: REQ-01-087
- **AC-331**: Row history for a supersede action that committed `replacement_record_id` shows one committed change containing both the `capture_state='superseded'` transition and the replacement-link addition; reviewer rollback of that change removes both effects, clears hidden `timeline.replacement_record_id`, and makes the Timeline row eligible for a fresh supersede.
  - Verifies: REQ-01-086, REQ-01-311..REQ-01-312, REQ-02-181, REQ-03-106..REQ-03-107
- **AC-121**: The Assessments view accepts band-first creation using the canonical `NULL`, `25`, `55`, and `85` `confidence_score` mapping, commits a new assessment row only when `subject_ref`, `subject_type`, `assessment_state`, and non-empty `rationale` are present, and does not let preseeded subject fields alone commit an otherwise empty assessment. When omitted on create, `assessed_at`, `assessor`, and `confidence_score` default deterministically to commit timestamp, current actor, and `NULL`. The view rejects in-place mutation of an existing row's `subject_ref`, `subject_type`, `assessment_state`, `confidence_score`, `rationale`, `assessor`, or `assessed_at`.
  - Verifies: REQ-01-296..REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093,
    REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011, REQ-03-250..REQ-03-254
- **AC-122**: The Indicators view allows inline creation of a new canonical indicator row only when the create request supplies enough information to determine canonical identity. For `cartulary.view.indicators.v1`, `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, and `indicator.normalized_value` when required by type-specific normalization must be present or derivable deterministically; `indicator.hash_algorithm` and `indicator.hash_value` are pairwise; and if the canonical dedupe basis is not determinable, create fails with no partial indicator row. An existing indicator row exposes no writable fields. Grid-edit and API patch attempts against an existing row fail closed for every create-only field. The exact identity-defining immutable field set for that schema is always `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, and `indicator.normalized_value`, plus `indicator.hash_algorithm` and `indicator.hash_value` when populated and used by the canonical dedupe key. `indicator.stix_pattern` and `indicator.defanged_value` remain rejected under the create-only rule without being treated as identity-defining.
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
- **AC-150**: Workbook open resolves the starting surface in this order: explicit `sheet_ref`, caller `home_sheet_ref`, incident `default_sheet_ref`, then `cartulary.view.timeline.v1`; direct selection of any pack-independent base-profile surface from `REQ-01-307`, including `cartulary.view.task_requests.v1`, `cartulary.view.decisions.v1`, `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`, `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`, succeeds when those pointers use `sheet_ref.kind='view_schema'` with the standardized `view_schema_id`, even if ordinary `GET /api/v1/incidents/{incident_id}/saved-views` returns no matching saved-view object; and if a referenced saved view is deleted, hidden by scope, or invalid because a required optional pack is unavailable, the invalid pointer is cleared and the implementation falls through deterministically to the next step instead of failing workbook open.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032
- **AC-112**: An analyst can create a note directly in the Notes sheet by inline create from the sheet itself with no required modal; the row commits only when `note.title` or `note.body` remains non-empty after `single_line_title_v1` or `multiline_body_v1` normalization, whitespace-only values do not commit, C0/C1 control characters outside the allowed multiline set are rejected, `note.title` values longer than 512 Unicode scalar values fail closed, `note.body` values longer than 16384 Unicode scalar values fail closed, and the resulting record is stored as an artifact with `artifact_type='note'`, ordinary note attribution, and timestamps.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-01-487..REQ-01-491, REQ-02-067..REQ-02-071,
    REQ-03-004, REQ-03-242..REQ-03-246
- **AC-068**: From a selected Timeline, Host, Identity, or Evidence record, an analyst can invoke `add linked note` without leaving the workbook flow; the action may preseed the contextual link, but the note does not commit until `note.title` or `note.body` remains non-empty after `single_line_title_v1` or `multiline_body_v1` normalization, whitespace-only values do not commit, C0/C1 control characters outside the allowed multiline set are rejected, `note.title` values longer than 512 Unicode scalar values fail closed, `note.body` values longer than 16384 Unicode scalar values fail closed, and the committed record is stored with the same artifact record shape as a Notes-sheet-created note, including `artifact_type='note'`, ordinary note attribution, and timestamps. The note appears in the Notes sheet and is visible as a linked record from the source row.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-01-487..REQ-01-491, REQ-02-067..REQ-02-071
- **AC-069**: Renaming the visible Notes tab label, and any implementation-supported per-user hide/show operation for that built-in tab, does not change write-back or export semantics because behavior remains bound to `view_schema_id`.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330
- **AC-070**: Editing a lightweight free-text field on a timeline, host, identity, or evidence record does not create a standalone Notes row. Creating a standalone note does create a distinct note record that can be linked, tagged, and reviewed in history independently.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071
- **AC-015**: An analyst can create a no-blob evidence request record only when the first committed create includes at least one user-supplied non-empty writable evidence field after create-time normalization; qualifying evidence fields are evaluated after the bound `single_line_title_v1`, `locator_text_v1`, or `party_text_v1` contract, so whitespace-only or control-only input does not count; `evidence.title` values longer than 512 Unicode scalar values, `evidence.storage_ref` values longer than 1024 Unicode scalar values, and `evidence.collector_party_text` or `evidence.source_party_text` values longer than 256 Unicode scalar values fail closed. When lifecycle choice is omitted, the committed row defaults `evidence.lifecycle_state='requested'` and fills `requested_at` from the commit timestamp if omitted. A same-surface create flow that reaches first commit through a finalized blob attachment also produces exactly one committed evidence row. A create attempt that supplies neither a qualifying field-based signal nor a finalized blob commits no evidence row. After a valid no-blob create, an analyst can later attach or replace the blob and preserve structured `requested_at`, `received_at`, `storage_ref`, `blob_hash`, `collector_party_text`, `source_party_text`, and append-only custody history across the lifecycle.
  - Verifies: REQ-01-243..REQ-01-247, REQ-01-291..REQ-01-295, REQ-01-355..REQ-01-366, REQ-01-487..REQ-01-493,
    REQ-02-186..REQ-02-201, REQ-03-120..REQ-03-126, REQ-03-242..REQ-03-246
- **AC-016**: Evidence processing and any implemented background job start without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
  - Verifies: REQ-01-243..REQ-01-247, REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-201, REQ-03-121..REQ-03-126
- **AC-043**: Within the supported large-grid operating envelope, selection change, focus change, and typing acknowledgment remain at or below 100 ms p95, and Timeline blank-row creation with one non-empty user-entered value remains at or below 150 ms p95.
  - Verifies: REQ-00-013, REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-087..REQ-03-089,
    REQ-03-217..REQ-03-219, REQ-03-263
- **AC-044**: Within the supported large-grid operating envelope, sort, filter, and grouping changes present a first useful viewport at or below 250 ms p95 and reach a stable viewport at or below 1.0 s p95.
  - Verifies: REQ-00-013, REQ-01-015..REQ-01-017, REQ-03-223..REQ-03-224
- **AC-045**: In an evidence-heavy incident where a Timeline row is linked to 100 evidence records, opening the inspector shows a metadata shell at or below 300 ms p95; binary preview bytes MAY continue progressively after the inspector opens.
  - Verifies: REQ-00-013, REQ-01-015..REQ-01-017, REQ-01-355..REQ-01-366, REQ-03-242..REQ-03-246
- **AC-046**: Imports, evidence processing, projection rebuilds, snapshot generation, report generation, and reference-pack refresh remain background jobs, show progress and cancellation within 1 second of job start, and do not block grid editing or row creation.
  - Verifies: REQ-00-013, REQ-01-004..REQ-01-014, REQ-01-018, REQ-01-248..REQ-01-249, REQ-01-342..REQ-01-348,
    REQ-01-351..REQ-01-353, REQ-01-369, REQ-01-452..REQ-01-454
- **AC-047**: Under representative live-update traffic within the supported large-grid operating envelope, scrolling, sorting, filtering, grouping, and live updates never retarget pending edits away from the selected `record_id`, and viewport stabilization does not require a full-sheet rerender.
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
- **AC-085**: From a selected Notes, Timeline, Host, Identity, or Evidence context, an analyst can create a first-class `task_request` from the workbook surface without a modal only when the first commit includes non-empty `task.title` and `task.task_kind`; `task.title` is validated under `single_line_title_v1`, while `task.requester_party_text`, `task.blocked_reason`, `task.external_ticket_ref`, and `task.closure_summary` are validated under `party_text_v1`, `reason_note_v1`, `locator_text_v1`, and `multiline_body_v1` respectively; preseeded contextual links or a preseeded decision reference do not commit an otherwise empty task, and a create attempt lacking that minimum set commits no row. When omitted on interactive inline create from the sheet itself, `task.status`, `task.owner_user_id`, and `task.priority` default to `open`, current actor, and `normal`. The committed record preserves required structured fields including `created_at` and `updated_at`; optional structured fields including `workstream`, `requester_party_text`, `due_at`, `blocked_reason`, `completed_at`, and `external_ticket_ref` remain editable and filterable; a blocked task requires `blocked_reason` after `reason_note_v1` normalization; a done task requires `completed_at`; and workbook filtering can show open tasks by owner, blocked state, due status, workstream, and external ticket reference.
  - Verifies: REQ-01-296..REQ-01-302, REQ-01-336..REQ-01-338, REQ-01-487..REQ-01-493, REQ-01-496, REQ-02-094..REQ-02-109,
    REQ-02-120..REQ-02-122, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260
- **AC-086**: Creating a `decision` record commits only when `decision_type` is present and `summary` plus `rationale` are non-empty after create-time normalization; `decision.summary` is validated under `single_line_title_v1`, `decision.rationale` is validated under `multiline_body_v1`, control characters outside the allowed multiline set are rejected, `decision.summary` values longer than 512 Unicode scalar values fail closed, and `decision.rationale` values longer than 16384 Unicode scalar values fail closed; preseeded `support_refs[]` do not commit an otherwise empty decision, a create attempt lacking that minimum set commits no row, and initial create with `status='superseded'` fails closed. When omitted on create, `status`, `owner_user_id`, and `decided_at` default to `proposed`, current actor, and commit time. Reviewer history can reconstruct when the decision changed state, who owned it, and what `support_refs[]` were attached.
  - Verifies: REQ-01-296..REQ-01-302, REQ-01-339..REQ-01-341, REQ-01-487..REQ-01-491, REQ-02-094, REQ-02-110..REQ-02-122,
    REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260
- **AC-087**: A `comm_log` artifact can record required `comm_type`, `timestamp_utc`, `audience`, `channel_or_meeting`, and `summary`, while a `status_review` artifact can record required timestamp and summary plus linked decision or task references, and a workbook surface can filter or sort those artifacts by `comm_type`, `audience`, timestamp, and next-report or next-checkpoint time without rereading unrelated note text.
  - Verifies: REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-123..REQ-02-133, REQ-03-005..REQ-03-011,
    REQ-03-255..REQ-03-260
- **AC-088**: A `handoff` artifact can capture outgoing owner, incoming owner, current state summary, open task IDs, and open decision IDs, and the incoming analyst can pivot directly from the handoff to the referenced open work from the same incident.
  - Verifies: REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-123..REQ-02-133, REQ-03-005..REQ-03-011,
    REQ-03-255..REQ-03-260
- **AC-089**: Hypothesis tracking remains artifact-backed in the current profile; committed current-profile hypothesis rows use exact `artifact_type='finding'` with `finding.kind='hypothesis'`; new durable current-profile writes do not use `artifact_type='hypothesis'`; and base conformance does not require or claim a first-class `hypothesis` `record_type`.
  - Verifies: REQ-01-296..REQ-01-302, REQ-02-067..REQ-02-071, REQ-02-094, REQ-02-123..REQ-02-136,
    REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260
- **AC-090**: Creating or editing timeline rows does not require owner, approver, challenge, checklist, task, or decision fields on the timeline sheet; ordinary row edits do not enter a generalized approval workflow; and ordinary row capture or ordinary grid editing is not blocked or interrupted by required `comm_log`, `handoff`, `status_review`, or `lesson` creation.
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
- **AC-074**: Multiple source-bound `indicator_observation` rows can resolve to one canonical indicator record identified exactly by incident plus deterministic type-specific dedupe key.
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
- **AC-022**: Pasting into an `entity_origin` mapping upserts an existing active host or identity on a unique exact-match key and otherwise creates a stub; for hosts and identities, exact-match reuse consults the canonical field plus the full active `exact_match_reuse` identifier set for that record rather than only singular canonical slots; it never auto-merges two pre-existing entities.
  - Verifies: REQ-02-030..REQ-02-036, REQ-02-060..REQ-02-061
- **AC-023**: Merging two entities preserves loser lineage, deterministically promotes or carries loser-side `exact_match_reuse` identifiers by class, copies only `suggestion_only` aliases through the ordinary alias surface, preserves `provenance_only` values without making them reusable, repoints live mention resolutions and live links to the survivor in one change set, ensures future exact-match reuse on any preserved reusable identifier resolves to the survivor rather than creating a new stub, and does not change the survivor `record_id`.
  - Verifies: REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249
- **AC-024**: In the timeline sheet, the grouping control offers only `None`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions` in the base profile, and no other grouping key. In particular, it does not offer `timeline.event_type`, Summary, Hosts, Identities, Tags, or arbitrary custom columns.
  - Verifies: REQ-03-225..REQ-03-230
- **AC-025**: A grouped timeline sheet exposes exactly one derived outline level. `expand group`, `collapse group`, `expand all`, `collapse all`, and `Group: None` work without creating editable rows, paste targets, subtotal rows, or `record_id`-bound mutation targets.
  - Verifies: REQ-03-225..REQ-03-232
- **AC-026**: While grouped, sorting, filtering, paste, autosave, conflict handling, rollback, and export flatten to underlying records only. Editing a grouped field may move the row to a different visible group, but drag-and-drop reclassification and manual row-range grouping are not available.
  - Verifies: REQ-03-225..REQ-03-233

### 9.1A Additional Base Profile criteria for direct runtime-family verification

These criteria provide direct runtime-family verification for substantive base-profile behavior that MUST NOT rely only on aggregate claim gates.

- **AC-404**: Base-profile topology is conformant only when one application deployable unit, optionally replicated without responsibility split, owns the browser-facing UI, the authenticated `/api/v1/*` surface, the incident-scoped `GET /ws/v1/incidents/{incident_id}` WebSocket subscription surface, and background-job execution; reverse proxies, TLS terminators, managed ingress, and load balancers MAY front that deployable unit but MUST NOT satisfy the application responsibilities themselves; and any deployment that splits those responsibilities across distinct application deployable units is non-conformant.
  - Verifies: REQ-01-001..REQ-01-003
- **AC-405**: Attaching a binary evidence payload commits one evidence row whose structured incident state contains metadata and object-reference state only; preview or download of that payload succeeds only while the authoritative object-store payload remains available; loss of object-store availability after commit leaves previously committed structured incident rows intact while preview or download fails closed as unavailable; and a Postgres logical backup or logical dump of the authoritative structured data store taken after attachment exposes no inline copy of that binary evidence payload.
  - Verifies: REQ-01-002, REQ-01-278..REQ-01-280
- **AC-406**: A row created from rough input with unresolved host or account text, unstructured `details`, or unstructured `source_text` can later be normalized, resolved, or canonically linked without overwriting the original rough input; after that later normalization, the original unresolved text, `details`, or `source_text` remains recoverable through the authoritative row, mention, or history surfaces.
  - Verifies: REQ-02-024..REQ-02-025
- **AC-407**: A state-changing mutation attempted without a current authenticated session fails closed with no committed change. On a fresh deployment, successful first-deployment-admin bootstrap or any other successful non-user startup mutation records an explicit system-process actor or source identity rather than an anonymous or implicit origin.
  - Verifies: REQ-04-036
- **AC-408**: One ordinary UI mutation, one structured-ingest mutation from clipboard paste or a claimed file-import path, and one rollback mutation each append attributable history or administrative audit state that records actor or explicit system-process identity, committed timestamp, mutation source, and the exact row-field or target-family mutation evidence enumerated by REQ-04-037 at the required history granularity; the three cases MUST remain distinguishable from one another in that recorded provenance.
  - Verifies: REQ-04-037

### 9.2 Import Extension Profile criteria

- **AC-027**: One uploaded CSV file or XLSX workbook import session starts without blocking grid editing, the UI shows progress and cancellation within 1 second of job start, and the operator can review per-unit discovery, preview, mapping, select, and skip before any apply step.
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
- **AC-065**: Re-applying the same `(import_unit_id, mapping_fingerprint, incident_id)` tuple triggers duplicate-apply detection and blocks by default until the operator explicitly chooses re-import. A one-field change to `unknown_column_policy`, `source_column_ordinal`, `source_header_text`, `field_key`, `entity_binding_mode`, `transform_id`, `transform_options`, or `empty_value_policy` changes `mapping_fingerprint`, while non-semantic JSON key order does not.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-169..REQ-03-192
- **AC-066**: Selecting overlapping `import_unit` rectangles in one batch is blocked before apply with an explicit overlap diagnostic. Non-overlapping selected units apply in deterministic order, one atomic `change_set` per unit, and the session can finish in `partially_applied`.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-169..REQ-03-186
- **AC-067**: Preview or apply of a unit containing formulas, merged cells, comments, pivots, charts, external links, hidden or filtered presentation state, or workbook or sheet protection never executes workbook behavior. The unit either emits only declared `warning_code[]` values or is rejected as unsupported, formula cells without stored cached values do not enter `ready` while mapped, and encrypted or password-protected workbooks that cannot be parsed are rejected before discovery.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-153..REQ-03-161, REQ-03-193..REQ-03-204


- **AC-262**: `POST /api/v1/import-sessions` accepts only `multipart/form-data` with a required `boundary`, exactly two leaf parts named `metadata` and `file`, and no alternate upload framing; part order is non-semantic; the `metadata` part uses `Content-Type: application/json` or `application/json; charset=utf-8`, is UTF-8 and BOM-free, parses as one JSON object, and contains required `incident_id` plus required `client_txn_id`; the `file` part media type is exactly one of `text/csv`, `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, or `application/octet-stream`; omitting `assistant_profile` defaults it to `phase2_workbook_import_v1`; duplicate or unexpected parts, malformed metadata JSON, duplicate metadata keys, a metadata JSON value that is not an object, invalid part content type, an unknown metadata member, or any envelope failure create no durable `import_session`, no idempotency commit, and no discovery job; replaying the same normalized metadata and file bytes by the same actor with the same `(incident_id, client_txn_id)` returns the original accepted `202` result even when boundary text, part order, or advisory filename differ; and when the discovery job succeeds, the terminal job summary uses `result_summary.code='import_session_discovered'` and exactly one `resource_refs[]` item `{ kind='import_session', id=<import_session_id>, route='/api/v1/import-sessions/{import_session_id}' }`.
  - Verifies: REQ-01-033, REQ-01-466..REQ-01-473, REQ-01-549..REQ-01-553
- **AC-263**: `GET /api/v1/import-sessions/{import_session_id}` returns exactly the durable `import_session resource` members; `GET /api/v1/import-sessions/{import_session_id}/units` returns `data.import_units[]` with each element using the exact durable `import_unit resource` shape; `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}` returns that same exact durable `import_unit resource`; the two singleton import reads reject pagination members; `GET /api/v1/import-sessions/{import_session_id}/units` uses the common cursor-pagination contract; `selected_unit_ids[]`, `blocking_diagnostics[]`, `nonblocking_warning_codes[]`, and `warning_codes[]` always serialize as `[]` when empty; `mapping_fingerprint` and `approved_mapping` are absent before mapping approval; serialized `header_row_ref` and `data_start_row_ref` are positive 1-based row references within `source_rect_a1`; and preview stays read-only against incident state, is not inlined into the durable unit resource, and exposes exactly the declared preview top-level members, exact `columns[]`, `preview_rows[]`, and `cells[]` shapes, the closed `cell_kind` vocabulary, preserved source order, and the first-50-row cap with `truncated` state.
  - Verifies: REQ-01-466..REQ-01-469, REQ-01-472, REQ-01-474
- **AC-264**: `PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping` accepts only the closed mapping body, persists one approved mapping plan, returns the exact durable `import_unit resource`, and derives deterministic `mapping_fingerprint`; non-contiguous ordinals, duplicate non-null target fields, invalid `unknown_column_policy`, invalid `transform_id` or `transform_options`, invalid `empty_value_policy`, or non-exhaustive `source_columns[]` fail with the import-family invalid-request surface; after mapping approval, `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}` reflects the exact `approved_mapping` read object; `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select` and `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip` require `client_txn_id`, return `data.unit` in that same exact `import_unit resource` shape, update persisted `selected_unit_ids[]` deterministically, preserve prior approved mappings when skipped, and recompute status from prior mapping when re-selected; `POST /api/v1/import-sessions/{import_session_id}/apply` omitting `selected_unit_ids[]` uses the session's persisted selection; apply returns `202` with the common job resource; when apply succeeds, the terminal job summary uses `result_summary.code='import_session_applied'` when the durable `session_status='applied'` and `result_summary.code='import_session_partially_applied'` when the durable `session_status='partially_applied'`; in both success cases `resource_refs[]` contains exactly one `import_session` ref using `/api/v1/import-sessions/{import_session_id}` as the canonical route; and the durable import-session and import-unit resources never surface job-status tokens such as `queued` or `running` as `session_status` or `unit_status`.
  - Verifies: REQ-01-467..REQ-01-470, REQ-01-472, REQ-01-474
- **AC-265**: Import-family routes use only `invalid_import_request`, `import_session_not_found`, `import_unit_not_found`, `import_state_conflict`, `import_source_unsupported`, `import_source_rejected`, and `import_apply_blocked`; `invalid_import_request` uses only `unsupported_upload_envelope`, `missing_required_part`, `duplicate_part`, `unexpected_part`, `invalid_part_content_type`, `invalid_metadata_encoding`, `malformed_metadata_json`, `request_not_object`, `missing_required_field`, `field_not_nullable`, `unknown_field`, `invalid_row_reference`, `invalid_selected_unit_ids`, `unsupported_assistant_profile`, `invalid_source_columns`, `invalid_unknown_column_policy`, `invalid_transform`, `invalid_empty_value_policy`, and `duplicate_target_field`; multipart part-related failures populate `error.details.part_name`, and `invalid_part_content_type` also includes `error.details.received_content_type` plus `error.details.allowed_content_types[]`; `import_state_conflict` uses only `session_applying`, `session_terminal`, `unit_applying`, and `unit_terminal`; unsupported source failures surface only `encrypted_or_unparseable_workbook`, `unsupported_named_range`, or `formula_cached_value_missing`; source-rejected failures surface only `csv_source_too_large`, `xlsx_source_too_large`, `import_rows_exceeded`, `import_columns_exceeded`, `import_cells_exceeded`, `archive_extracted_bytes_exceeded`, `archive_compression_ratio_exceeded`, or `archive_member_count_exceeded`; and blocked apply requests surface only `overlapping_units`, `duplicate_apply_blocked`, or `unit_not_ready`.
  - Verifies: REQ-01-471, REQ-01-475, REQ-01-553

### 9.3 Snapshot and Reporting Extension Profile criteria

- **AC-030**: Report generation and snapshot generation run without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
  - Verifies: REQ-01-369..REQ-01-373, REQ-01-452..REQ-01-454, REQ-02-139..REQ-02-146
- **AC-031**: A generated HTML report or presentation artifact opens in a disconnected browser without fetching remote JavaScript, CSS, or fonts.
  - Verifies: REQ-01-370..REQ-01-373, REQ-01-394..REQ-01-398, REQ-02-139..REQ-02-146, REQ-04-044..REQ-04-047
- **AC-032**: Snapshot-derived outputs preserve stable identifiers and ordering consistent with the canonical derivation layer.
  - Verifies: REQ-01-342..REQ-01-348, REQ-01-351..REQ-01-353, REQ-01-367..REQ-01-368, REQ-01-370..REQ-01-373,
    REQ-02-139..REQ-02-146
- **AC-056**: Rendering from the same `snapshot_id`, `snapshot_at`, `source_change_set_high_watermark`, `derivation_version`, `template_id`, `template_version`, `redaction_profile_id`, and `redaction_profile_version` produces the same `export_model_sha256` and deterministic export ordering.
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
- **AC-333**: Given a snapshot whose canonical export model contains: (a) one or more text-bearing blocks materialized directly from each of ad hoc notes, structured finding rows where `finding.kind='hypothesis'`, `task_request`, `decision`, `comm_log`, `handoff`, `status_review`, and `lesson` source families; (b) one or more separate deterministic analytic blocks derived from those same source families using only stable identifiers, enums, timestamps, counts, or other non-narrative scalars; and (c) one separately curated narrative block with valid `support_refs[]`; the canonical export model classifies every direct-source text-bearing block as `working_material`, classifies the non-narrative analytic blocks as `derived_analytic`, preserves those `content_class` values unchanged when rendering the same snapshot for `internal_draft` and `internal_review`, and excludes the `working_material` blocks from `external_release` while allowing only the eligible `derived_analytic` and curated blocks.
  - Verifies: REQ-01-377..REQ-01-380, REQ-02-139, REQ-02-143, REQ-02-211, REQ-04-045
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


- **AC-266**: `POST /api/v1/snapshots` accepts a JSON object with required `incident_id` and required `client_txn_id`; omitting `source_change_set_high_watermark` resolves snapshot materialization to the current committed incident head at job admission; explicit `null` fails with `400 error.code='invalid_snapshot_request'` and `reason_code='field_not_nullable'`; a supplied boundary outside the exact committed source-boundary vocabulary for that incident fails with `reason_code='invalid_source_boundary'`; omission and explicit transmission of the same resolved committed boundary compare equal for idempotency and replay; exact replay uses the originally resolved boundary rather than a later incident head; the route returns `202` with the common job resource; on success the terminal job summary uses `result_summary.code='snapshot_created'` and exactly one `resource_refs[]` item `{ kind='snapshot', id=<snapshot_id>, route='/api/v1/snapshots/{snapshot_id}' }`; and `GET /api/v1/snapshots/{snapshot_id}` rejects pagination members and returns exactly `snapshot_id`, `incident_id`, `created_by_user_id`, `created_at`, `snapshot_at`, `source_change_set_high_watermark`, `derivation_version`, and `export_model_sha256`, with the resolved `source_change_set_high_watermark` always serialized and no inlined template, redaction-profile, release-state, approval, redaction-manifest, or rendered-byte fields.
  - Verifies: REQ-01-033, REQ-01-466..REQ-01-477
- **AC-267**: `POST /api/v1/releases` requires exact `snapshot_id`, `template_id`, `template_version`, `redaction_profile_id`, `redaction_profile_version`, `output_kind`, and `client_txn_id`; omitting `release_scope` defaults it to `internal_draft`; explicit `null` for `release_scope` fails with `400 error.code='invalid_release_request'` and `reason_code='field_not_nullable'`; the only allowed `release_scope` values are `internal_draft`, `internal_review`, and `external_release`; any other value fails with `reason_code='invalid_release_scope'`; omission and explicit `internal_draft` compare equal for idempotency and replay; requests using `latest`, `current`, or missing version selectors fail closed with `400 error.code='invalid_release_request'`; render starts as a `202` common job rather than a blocking route; and on success the terminal job summary uses `result_summary.code='release_created'` and exactly one `resource_refs[]` item `{ kind='release', id=<release_id>, route='/api/v1/releases/{release_id}' }`, with the durable `release resource` always serializing the resolved `release_scope`.
  - Verifies: REQ-01-466..REQ-01-470, REQ-01-476..REQ-01-477
- **AC-268**: `GET /api/v1/releases/{release_id}` returns exactly the durable `release resource` members; `approved_at`, `invalidated_at`, `published_at`, and `invalidation_reason` are always present and serialize as JSON `null` when unset; rendered release resources use only `pending_approval`, `approved`, `invalidated`, and `published` as `release_state`; `internal_draft` candidates become `approved` immediately on successful render; `approve`, `publish`, and `invalidate` each require `client_txn_id`, reject pagination members, and enforce their declared legal state transitions without introducing `queued`, `running`, or `rendering` into `release_state`; and the durable release resource never inlines approval records, redaction manifests, rendered bytes, or worker-phase status.
  - Verifies: REQ-01-467..REQ-01-470, REQ-01-374..REQ-01-376, REQ-01-476, REQ-01-478, REQ-04-031..REQ-04-035
- **AC-269**: Snapshot and release routes use only `invalid_snapshot_request`, `snapshot_not_found`, `invalid_release_request`, `release_not_found`, `release_state_conflict`, `release_approval_rejected`, and `release_render_failed`; render failures surface only `missing_redaction_rule`, `undeclared_template_binding`, or `missing_required_field`; state conflicts surface only `not_approved`, `already_approved`, `already_published`, or `already_invalidated`; and approval rejections surface only `approval_requirements_not_met` while the durable release state still permits an approval attempt.
  - Verifies: REQ-01-471, REQ-01-479
- **AC-305**: `POST /api/v1/releases/{release_id}/approve`, `publish`, and `invalidate` accept only a JSON object with required `client_txn_id` and optional `reason`; if present, `reason` accepts only JSON string or JSON `null` and normalizes under `reason_note_v1`; omission, explicit `null`, and normalized-empty input compare equal; unknown top-level members, non-object bodies, missing `client_txn_id`, or `null` for non-nullable members fail with `400 error.code='invalid_release_request'`; idempotency is route-scoped on `(actor_user_id, release_id, action_route, client_txn_id)`; exact replay returns the original committed success before fresh state evaluation; and same-key different-request reuse fails with `409 error.code='client_txn_conflict'`.
  - Verifies: REQ-01-469..REQ-01-470, REQ-01-478, REQ-01-496
- **AC-306**: A successful `approve` returns `200 OK` with the common success envelope, `data.release` using the same exact durable `release resource` shape returned by `GET /api/v1/releases/{release_id}`, and `data.approval_progress` containing boolean `approval_recorded`, boolean `approval_requirements_satisfied`, and `resulting_release_state`; a valid approval may leave both `resulting_release_state` and the durable `release_state` at `pending_approval` when the full approval set is not yet satisfied.
  - Verifies: REQ-01-478, REQ-04-034
- **AC-307**: A fresh `approve` against an already `approved` release fails with `409 error.code='release_state_conflict'` and `reason_code='already_approved'`; a fresh `approve` or `publish` against an already `published` release fails with `reason_code='already_published'`; a caller or artifact tuple that cannot contribute a valid approval while the release remains `pending_approval` fails with `409 error.code='release_approval_rejected'` and `reason_code='approval_requirements_not_met'`; and invalidated-state failures continue to use the `release_state_conflict` family.
  - Verifies: REQ-01-471, REQ-01-479

### 9.4 Reference Pack Extension Profile criteria

- **AC-033**: Reference-pack import, verification, and refresh run without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
  - Verifies: REQ-01-399, REQ-01-407..REQ-01-413, REQ-01-452..REQ-01-454, REQ-04-040..REQ-04-043
- **AC-034**: Missing optional reference packs degrade only the affected overlay labels, enrichment semantics, non-canonical analytical widgets, or snapshot/report derivations; timeline capture, entity resolution, and evidence attachment continue to function; and missing packs do not create or remove current-profile workbook `view_schema` surfaces.
  - Verifies: REQ-01-282..REQ-01-284, REQ-01-399, REQ-01-419..REQ-01-422
- **AC-035**: Pack activation fails closed with `error.code='reference_pack_verification_failed'`; `error.details.reason_code` is limited to `checksum_mismatch`, `signature_mismatch`, `missing_integrity_metadata`, `contract_incompatible`, `path_traversal`, `disallowed_content`, `payload_missing`, `archive_extracted_bytes_exceeded`, `archive_compression_ratio_exceeded`, or `archive_member_count_exceeded`; and the previously active version remains active.
  - Verifies: REQ-01-399, REQ-01-409..REQ-01-421, REQ-04-040..REQ-04-043
- **AC-092**: In the smallest supported disconnected bundle, the only preinstalled active reference packs are `type_registry.host`, `type_registry.evidence`, and `type_registry.indicator`; the deployment remains usable without `framework.attack`, `framework.d3fend`, `framework.veris`, or any enrichment pack.
  - Verifies: REQ-01-282..REQ-01-284, REQ-01-400..REQ-01-406, REQ-04-040..REQ-04-043, REQ-04-054..REQ-04-055
- **AC-093**: Given an offline pack bundle supplied either by placement in the configured reference-pack storage root or by `POST /api/v1/reference-packs/import`, the system stages the bundle under a temporary-work root, verifies it, records the candidate version in durable condition `verified_available` on success, leaves the active-version pointer unchanged, and does not activate that version until explicit operator action.
  - Verifies: REQ-01-407..REQ-01-413, REQ-02-247, REQ-04-040..REQ-04-043
- **AC-094**: Given a candidate pack with checksum mismatch, signature mismatch, missing required integrity metadata, incompatible `pack_contract_version`, path-traversal attempt, or disallowed active content, import or activation fails closed, the candidate version remains inactive, and the previously active version, if any, remains active.
  - Verifies: REQ-01-407..REQ-01-418, REQ-02-247, REQ-04-040..REQ-04-043
- **AC-095**: Pack import and activation record structured attestation metadata queryable without unpacking bundle contents, including `pack_key`, `pack_kind`, `pack_version`, `manifest_sha256`, `payload_sha256`, `source_identifier`, `verification_method`, signer-key or trusted-source identifier, imported and activated actor attribution with timestamps, `previous_active_version`, and `verification_result`.
  - Verifies: REQ-01-409..REQ-01-421, REQ-02-247, REQ-04-040..REQ-04-043
- **AC-096**: In a disconnected deployment, reference-pack import or activation succeeds without outbound network access and no supported pack-activation path performs a live internet fetch.
  - Verifies: REQ-01-282..REQ-01-284, REQ-01-407..REQ-01-413, REQ-04-040..REQ-04-043, REQ-04-054..REQ-04-055


- **AC-270**: `POST /api/v1/reference-packs/import` accepts only `multipart/form-data` with a required `boundary`, exactly two leaf parts named `metadata` and `file`, and no alternate upload framing; part order is non-semantic; the `metadata` part uses `Content-Type: application/json` or `application/json; charset=utf-8`, is UTF-8 and BOM-free, parses as one JSON object, and contains required `client_txn_id`; the `file` part media type is exactly one of `application/zip`, `application/x-tar`, `application/gzip`, `application/x-gzip`, or `application/octet-stream`; omitting `activation_policy` defaults it to `staged_only`; explicit `null` for `activation_policy` fails with `400 error.code='invalid_reference_pack_request'` and `reason_code='field_not_nullable'`; explicit `activation_policy='staged_only'` compares equal to omission for idempotency and replay; any non-null string token other than `staged_only` fails with `reason_code='auto_activation_not_supported'`; any other malformed non-null form for `activation_policy` fails with `reason_code='invalid_activation_policy'`; duplicate or unexpected parts, malformed metadata JSON, duplicate metadata keys, a metadata JSON value that is not an object, invalid part content type, or any envelope failure create no durable pack-version state, no idempotency commit, and no import job; replaying the same normalized metadata and file bytes by the same actor with the same `client_txn_id` returns the original accepted `202` result even when boundary text, part order, or advisory filename differ; the import route returns `202` with the common job resource; on success the terminal job summary uses `result_summary.code='reference_pack_imported'` and exactly one `resource_refs[]` item `{ kind='reference_pack_version', id='/api/v1/reference-packs/{pack_key}/{pack_version}', route='/api/v1/reference-packs/{pack_key}/{pack_version}' }`; `GET /api/v1/reference-packs` uses the common cursor-pagination contract and returns `data.pack_versions[]` with `meta.paging`, where every array item uses the exact `reference_pack_version resource` shape defined by Core 01 §17.4 and the array sorts by `pack_key asc` then exact `pack_version asc`; and `GET /api/v1/reference-packs/{pack_key}/{pack_version}` rejects pagination members.
  - Verifies: REQ-01-033, REQ-01-466..REQ-01-470, REQ-01-480..REQ-01-481, REQ-01-549..REQ-01-553
- **AC-271**: `POST /api/v1/reference-packs/{pack_key}/{pack_version}/activate`, `disable`, and `reverify` require an exact path `pack_version`; the public surface offers no implicit latest-version action route; `active` is derived from the activation pointer rather than stored as an additional version-state token; import, reverify, and refresh use jobs; and activate or disable preserve the durable version-state vocabulary `staged`, `verified_available`, `disabled`, `failed`, and `missing`.
  - Verifies: REQ-01-467..REQ-01-470, REQ-01-409..REQ-01-421, REQ-01-480..REQ-01-481
- **AC-272**: Reference-pack routes use only `invalid_reference_pack_request`, `reference_pack_not_found`, `reference_pack_state_conflict`, `reference_pack_verification_failed`, and `reference_pack_activation_rejected`; `invalid_reference_pack_request` uses only `unsupported_upload_envelope`, `missing_required_part`, `duplicate_part`, `unexpected_part`, `invalid_part_content_type`, `invalid_metadata_encoding`, `malformed_metadata_json`, `request_not_object`, `missing_required_field`, `field_not_nullable`, `unknown_field`, `invalid_activation_policy`, `pack_version_required`, `auto_activation_not_supported`, `invalid_pack_keys`, and `empty_pack_keys`; multipart part-related failures populate `error.details.part_name`, and `invalid_part_content_type` also includes `error.details.received_content_type` plus `error.details.allowed_content_types[]`; verification failures surface only `checksum_mismatch`, `signature_mismatch`, `missing_integrity_metadata`, `contract_incompatible`, `path_traversal`, `disallowed_content`, `payload_missing`, `archive_extracted_bytes_exceeded`, `archive_compression_ratio_exceeded`, or `archive_member_count_exceeded`; activation rejections surface only `already_active` or `not_verified_available`; and `reference_pack_state_conflict` reasons surface only `already_disabled`, `not_disableable`, or `verification_pending`.
  - Verifies: REQ-01-471, REQ-01-482, REQ-01-553
- **AC-308**: `POST /api/v1/reference-packs/{pack_key}/{pack_version}/activate`, `disable`, and `reverify` accept only a JSON object with required `client_txn_id` and optional `reason`; if present, `reason` accepts only JSON string or JSON `null` and normalizes under `reason_note_v1`; omission, explicit `null`, and normalized-empty input compare equal; unknown top-level members, non-object bodies, missing `client_txn_id`, or `null` for non-nullable members fail with `400 error.code='invalid_reference_pack_request'`; idempotency is route-scoped on `(actor_user_id, pack_key, pack_version, action_route, client_txn_id)`; exact replay returns the original committed success or accepted job before fresh state evaluation; and same-key different-request reuse fails with `409 error.code='client_txn_conflict'`.
  - Verifies: REQ-01-469..REQ-01-470, REQ-01-481, REQ-01-496
- **AC-309**: `activate` is legal only when the addressed version is in durable condition `verified_available` and is not currently active; `disable` is legal only when the addressed version is in durable condition `verified_available`, whether or not that version is currently active; `reverify` is legal only when the addressed version is in durable condition `verified_available`, `disabled`, `failed`, or `missing`; `refresh` and `reverify` always return `202` with the common job resource; `GET /api/v1/reference-packs/{pack_key}/{pack_version}` returns the exact `reference_pack_version resource` shape used by `GET /api/v1/reference-packs` `data.pack_versions[]`; inline `activate` or `disable` returns `200` with `data.pack_version` using that identical shape and identical nullability rules; a long-running `activate` or `disable` returns `202` with the common job resource; and when a reference-pack family action completes through the common job resource, `reference_pack_version` refs use the canonical `/api/v1/reference-packs/{pack_key}/{pack_version}` path as both `route` and `id`, import success uses `reference_pack_imported`, long-running `activate` success uses `reference_pack_activated`, long-running `disable` success uses `reference_pack_disabled`, `reverify` success uses `reference_pack_reverified`, and `refresh` success uses `reference_packs_refreshed` with zero or more non-exhaustive `reference_pack_version` refs sorted by `route asc`.
  - Verifies: REQ-01-467, REQ-01-481
- **AC-369**: `POST /api/v1/reference-packs/refresh` accepts required `client_txn_id` and optional `pack_keys[]`; explicit `null` for `pack_keys[]` fails with `400 error.code='invalid_reference_pack_request'` and `reason_code='field_not_nullable'`; explicit `[]` fails with `reason_code='empty_pack_keys'`; supplied `pack_keys[]` is treated as a set-like selector so caller order is non-semantic, duplicate members coalesce by exact token equality, and canonical normalization sorts the unique `pack_key` set by `pack_key asc`; unknown, non-visible, or non-string members fail with `reason_code='invalid_pack_keys'`; omitted `pack_keys[]` resolves once at refresh-job admission to all currently imported pack keys visible to the caller and that resolved set, not later visibility state, is the value used for idempotency and replay; and if the omitted selector resolves to zero visible imported pack keys, the refresh still succeeds as a deterministic no-op background job whose terminal success remains `reference_packs_refreshed` with zero `resource_refs[]` items.
  - Verifies: REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-481
- **AC-310**: `reference_pack_state_conflict` uses only `already_disabled`, `not_disableable`, and `verification_pending`; disabling an already disabled version yields `already_disabled`; disabling a `staged`, `failed`, or `missing` version yields `not_disableable`; reverifying a `staged` version yields `verification_pending`; and activation rejections remain limited to `already_active` or `not_verified_available`.
  - Verifies: REQ-01-471, REQ-01-482

### 9.5 Enterprise Authentication Extension Profile criteria

- **AC-036**: Enterprise-authenticated users map to the same internal user identity used for attribution, and switching from local auth to enterprise auth does not break audit lineage for existing incidents.
  - Verifies: REQ-04-018..REQ-04-020

- **AC-288**: `GET /api/v1/auth/providers` returns only enabled interactive enterprise-auth providers, orders them by `display_name asc` then `provider_key asc`, exposes only `provider_key`, `provider_type`, and `display_name`, and rejects pagination members rather than silently ignoring them.
  - Verifies: REQ-01-510..REQ-01-511, REQ-04-018
- **AC-289**: `POST /api/v1/auth/providers/{provider_key}/begin` accepts only a JSON object with optional `return_to`; omitted or explicit `null` `return_to` normalizes to `/`; `/` is the canonical authenticated post-login landing route; an off-origin, absolute, or otherwise disallowed `return_to` fails closed with the enterprise-auth request error family; unknown top-level members are rejected; `client_txn_id` is rejected; and replaying the same structurally valid request may mint a fresh redirect target rather than replaying a durable public auth-transaction resource. Conformance MUST also verify that, after successful provider authentication, `/` opens the sole visible incident workbook when exactly one visible incident exists, remains on `/` and renders the caller's visible incident list when zero or multiple visible incidents exist, falls back to that same list if the sole selected incident is no longer visible by workbook bootstrap time, and never implicitly chooses one incident from multiple visible incidents by recency, sort order, provider claim content, or any other heuristic.
  - Verifies: REQ-01-510..REQ-01-511, REQ-01-514..REQ-01-515, REQ-04-018
- **AC-290**: A valid OIDC callback issues the same session resource shape exposed by `GET /api/v1/auth/session`, terminates with `303 See Other` to the validated `return_to`, and preserves audit lineage for the linked internal user. When the validated `return_to` resolves to a workbook surface, startup-surface selection inside that workbook uses the Core 03 §2.4 ordered fallback rather than an enterprise-auth-specific order. For the default `/` path, the sole-visible-incident branch opens the workbook with no explicit launch `sheet_ref`, so the fallback continues through the caller's `home_sheet_ref`, then the incident-wide `default_sheet_ref`, then `cartulary.view.timeline.v1`, with invalid pointers cleared before fall-through. Replay, expiry, missing or invalid `state`, missing or invalid `nonce`, failed code exchange, or provider mismatch create no session and fail with the enterprise-auth transaction or provider-response error family.
  - Verifies: REQ-01-031, REQ-01-510, REQ-01-512, REQ-01-515, REQ-04-018, REQ-04-020
- **AC-291**: A valid SAML ACS issues the same session resource shape exposed by `GET /api/v1/auth/session`, terminates with `303 See Other` to the validated `return_to`, and preserves audit lineage for the linked internal user. When the validated `return_to` resolves to a workbook surface, startup-surface selection inside that workbook uses the Core 03 §2.4 ordered fallback rather than an enterprise-auth-specific order. For the default `/` path, the sole-visible-incident branch opens the workbook with no explicit launch `sheet_ref`, so the fallback continues through the caller's `home_sheet_ref`, then the incident-wide `default_sheet_ref`, then `cartulary.view.timeline.v1`, with invalid pointers cleared before fall-through. Replay, expiry, `RelayState` mismatch, issuer mismatch, audience mismatch, signature failure, assertion expiry, or provider mismatch create no session and fail with the enterprise-auth transaction or provider-response error family.
  - Verifies: REQ-01-031, REQ-01-510, REQ-01-512, REQ-01-515, REQ-04-018, REQ-04-020
- **AC-292**: Successful enterprise authentication never auto-creates a local user, never auto-creates an `auth_identity`, never auto-creates incident membership, and never maps provider group claims into incident roles.
  - Verifies: REQ-01-513, REQ-04-019
- **AC-293**: Enterprise-auth protocol routes under `/api/v1/auth/*` use only `invalid_enterprise_auth_request`, `auth_provider_not_found`, `auth_provider_disabled`, `enterprise_auth_transaction_rejected`, `provider_response_rejected`, and `provider_identity_rejected`; their `reason_code` use is limited to the declared enterprise-auth registries; and provider identity binding rejects missing, ambiguous, inactive, or unlinked subjects with the correct family failure.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-513..REQ-01-514, REQ-04-018..REQ-04-020
- **AC-348**: The Enterprise Authentication Extension Profile additionally exposes `POST /api/v1/users/{user_id}/auth-bindings`, `POST /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}/rotate`, and `DELETE /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}`. Binding create targets an existing local `user_id`, accepts required `base_user_version`, `client_txn_id`, `provider_key`, and `provider_subject` plus optional `reason`, allows a configured OIDC or SAML provider even when that provider is currently disabled for interactive sign-in, returns `201 Created` on first success and `200 OK` on exact replay, and, when exercised against a user who begins with no enterprise bindings, returns a safe user resource whose `auth_bindings[]` contains exactly two active items in canonical order: first one local binding summary containing exactly `provider_type='local'`, `provider_key='local'`, `username` equal to the authoritative safe-user `email`, and `created_at` equal to the safe user resource `created_at`, while omitting `auth_binding_id`, `provider_subject`, and `last_auth_at`; then one enterprise binding summary containing exactly `auth_binding_id`, `provider_key`, `provider_type`, `provider_subject`, `created_at`, and `last_auth_at`, while omitting `username`; and no additional members appear on either item.
  - Verifies: REQ-01-033, REQ-01-116, REQ-01-537..REQ-01-538, REQ-02-234, REQ-02-236..REQ-02-237, REQ-04-093
- **AC-349**: Binding create fails closed with `404 error.code='auth_provider_not_found'` for an unknown or non-enterprise `provider_key`, with `409 error.code='auth_binding_conflict'` and `reason_code='provider_subject_in_use'` when the requested subject is already active on a different user, with `409 error.code='auth_binding_conflict'` and `reason_code='provider_already_linked_for_user'` when the addressed user already has one active binding for that provider, and with `409 error.code='client_txn_conflict'` on same-key different-request reuse. Conformance MUST also verify that `provider_subject` comparison is exact and does not trim, case-fold, Unicode-normalize, or email-normalize the supplied or returned value.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-538, REQ-01-541, REQ-02-234..REQ-02-235
- **AC-350**: Binding rotate is atomic retire-plus-create on the same provider and same `user_id`, preserves audit lineage, does not re-key the old binding in place, returns `200 OK` and does not advance `user_version` when `new_provider_subject` exactly equals the current subject, updates `last_auth_at` only on successful callbacks against the active binding, and causes future callback authentication using the old subject to fail with `409 error.code='provider_identity_rejected'` and `reason_code='no_linked_user'`.
  - Verifies: REQ-01-513..REQ-01-514, REQ-01-539..REQ-01-541, REQ-02-234..REQ-02-235, REQ-04-020, REQ-04-095
- **AC-351**: For a user with one active enterprise binding, binding retire removes only that targeted enterprise binding from active callback resolution and from active `auth_bindings[]` summaries without deleting the user or memberships, exact replay returns the original committed success, a fresh request against an already retired binding fails with `409 error.code='auth_binding_conflict'` and `reason_code='binding_not_active'`, the returned safe user resource retains exactly one local binding summary containing exactly `provider_type='local'`, `provider_key='local'`, `username` equal to the authoritative safe-user `email`, and `created_at` equal to the safe user resource `created_at`, while omitting `auth_binding_id`, `provider_subject`, and `last_auth_at`, and local email or password login remains unchanged unless separately changed through the local credential-management routes.
  - Verifies: REQ-01-116, REQ-01-234, REQ-01-238, REQ-01-540..REQ-01-541, REQ-02-234..REQ-02-237, REQ-04-095
- **AC-352**: Only `deployment_admin` can call the three enterprise-auth binding-management routes; incident role `admin` alone cannot. Create, rotate, and retire are auditable deployment-local administrative events; rotate and retire revoke all active sessions for the target user immediately; the current profile uses this audited deployment-admin binding-management surface as the only normative runtime mechanism for binding create, rotate, and retire; the returned safe user resources and `auth_bindings[]` summaries expose only the exact member sets required by REQ-01-115 and REQ-01-116; and deployment-local audit state plus route-scoped idempotency state retain no member of the closed forbidden set in REQ-04-096.
  - Verifies: REQ-01-537..REQ-01-541, REQ-02-236, REQ-04-093..REQ-04-096

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
- **AC-053**: Evidence upload and preview do not execute uploaded active content in the main application unit or browser origin. Preview allowlisting and active-content classification derive from server-observed or otherwise validated media state rather than from caller-supplied `filename_hint` or `content_type_hint`. Non-previewable or active-content types remain quarantined or download-only unless an explicit isolated analysis path is configured.
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
- **AC-101**: If the implementation exposes structured findings, investigative queries, or forensic keywords as workbook surfaces, their defining fields remain directly filterable and exportable as structured fields rather than JSON-only payloads; for `cartulary.view.findings.v1`, this includes projection-backed `finding.kind` as a structured field rather than implementation-local subtype inference.
  - Verifies: REQ-01-358, REQ-01-507, REQ-02-009..REQ-02-023, REQ-02-134..REQ-02-138

### 9.9 Lifecycle machine criteria

#### 9.9.0 Lifecycle conformance fixture matrix

These matrices are normative for AC-108 and AC-110. Only rows whose `profiles` are satisfied by the active claim are required. Within this subsection, `machine_id` values are a closed allowlist.

##### Table 9.9-A. Failure and blocked-follow-on fixtures

| Fixture ID | `machine_id` | `profiles` | Starting condition | Action and expected signal | Expected committed state | Blocked follow-on action and expected signal |
| --- | --- | --- | --- | --- | --- | --- |
| `LF-01` | `timeline.capture_state` | `base` | A Timeline row is visible and persisted with `capture_state='superseded'`. | An ordinary Timeline patch that changes `timeline.summary` fails with `409 error.code='illegal_transition'`. | The row remains `capture_state='superseded'`; `timeline.summary` is unchanged; no new `change_set` commits. | `POST /api/v1/records/{record_id}/mark-reviewed` fails with `409 error.code='illegal_transition'`. |
| `LF-02` | `task_request.status` | `base` | A `task_request` row is visible and persisted with `status='done'` and non-null `completed_at`. | A direct status patch to `task.status='canceled'` fails with `409 error.code='illegal_transition'`. | The row remains `status='done'`; `completed_at` is unchanged; no new `change_set` commits. | A patch that leaves `task.status='done'` and clears `task.completed_at` fails with `409 error.code='illegal_transition'`. |
| `LF-03` | `decision.status` | `base` | A `decision` row is visible and persisted with `status='executed'`. | A direct status patch to `decision.status='approved'` fails with `409 error.code='illegal_transition'`. | The row remains `status='executed'`; no new `change_set` commits. | A direct status patch to `decision.status='superseded'` fails with `409 error.code='illegal_transition'`. |
| `LF-04` | `object_blobs.upload_state` | `base` | A blob slot is persisted with `upload_state='pending'` and accepted contract `byte_size=N`. | Explicit finalization with observed bytes whose size is not `N` fails immediately and records `terminal_reason='declared_size_mismatch'`. | The slot transitions to `upload_state='failed'` with `terminal_reason='declared_size_mismatch'`; no evidence attachment commits. | `POST /api/v1/evidence-records/{record_id}/attach-blob` with that `object_blob_id` fails closed and creates no evidence-state mutation. |
| `LF-04a` | `object_blobs.upload_state` | `base` | A blob slot is persisted with `upload_state='pending'`. | A quarantine-entry attempt using `content_inspection_quarantine` or `admin_quarantine` while the slot is still `pending` fails with `409 error.code='illegal_transition'`. | The slot remains `upload_state='pending'`; no evidence attachment commits. | `POST /api/v1/evidence-records/{record_id}/attach-blob` with that `object_blob_id` fails closed and creates no evidence-state mutation. |
| `LF-05a` | `evidence_records.lifecycle_state` | `base` | An evidence row is visible and persisted with `lifecycle_state='available'`, it has a linked blob, and that blob becomes `upload_state='quarantined'` through one legal blob-quarantine trigger. | Ordinary preview-handle issuance fails with `409 error.code='evidence_access_unavailable'` and `error.details.reason_code='evidence_quarantined'`. | The evidence row transitions to `lifecycle_state='quarantined'`; preview and download remain blocked. | Ordinary download-handle issuance fails with `409 error.code='evidence_access_unavailable'` and `error.details.reason_code='evidence_quarantined'`. |
| `LF-05b` | `evidence_records.lifecycle_state` | `base` | An evidence row is visible and persisted with `lifecycle_state='released'`, it has a linked blob, and that blob becomes `upload_state='quarantined'` through one legal blob-quarantine trigger. | Ordinary preview-handle issuance fails with `409 error.code='evidence_access_unavailable'` and `error.details.reason_code='evidence_quarantined'`. | The evidence row transitions to `lifecycle_state='quarantined'`; preview and download remain blocked. | Ordinary download-handle issuance fails with `409 error.code='evidence_access_unavailable'` and `error.details.reason_code='evidence_quarantined'`. |
| `LF-05c` | `evidence_records.lifecycle_state` | `base` | An evidence row is visible and persisted with `lifecycle_state='released'` and a linked blob in `upload_state='available'`. | A direct status patch to `evidence.lifecycle_state='requested'` fails with `409 error.code='illegal_transition'`. | The row remains `lifecycle_state='released'`; no new `change_set` commits. | A direct status patch to `evidence.lifecycle_state='received'` fails with `409 error.code='illegal_transition'`. |
| `LF-06` | `release_state` | `snapshot_reporting` | A release row is visible and persisted with `release_state='published'`. | `POST /api/v1/releases/{release_id}/publish` fails with `409 error.code='release_state_conflict'` and `reason_code='already_published'`. | The row remains `release_state='published'`; no new release-state transition commits. | `POST /api/v1/releases/{release_id}/approve` fails with `409 error.code='release_state_conflict'` and `reason_code='already_published'`. |
| `LF-07` | `reference_pack.version_condition` | `reference_pack` | The addressed pack version is visible, in durable condition `disabled`, and not currently active for its `pack_key`. | `POST /api/v1/reference-packs/{pack_key}/{pack_version}/activate` fails with `409 error.code='reference_pack_activation_rejected'` and `reason_code='not_verified_available'`. | The version remains `disabled`; the active-version pointer is unchanged. | `POST /api/v1/reference-packs/{pack_key}/{pack_version}/disable` fails with `409 error.code='reference_pack_state_conflict'` and `reason_code='already_disabled'`. |

##### Table 9.9-B. Replay and duplicate-delivery fixtures

| Fixture ID | `machine_id` | `profiles` | Starting condition | Action and replay condition | Expected committed state | Exact no-duplicate durable-side-effect rule |
| --- | --- | --- | --- | --- | --- | --- |
| `LR-01` | `timeline.capture_state` | `base` | A Timeline row is visible and persisted with `capture_state='rough'`. | `POST /api/v1/records/{record_id}/mark-reviewed` succeeds once, then the same authenticated actor replays the same normalized request with the same `(record_id, client_txn_id)`. | The row ends in `capture_state='reviewed'`. | Replay returns the originally committed result and creates no second lifecycle transition, `change_set`, revision entry, or replayable `record_changed` event. |
| `LR-02` | `task_request.status` | `base` | A `task_request` row is visible and persisted with `status='open'`. | A patch sets `task.status='in_progress'`, then the same authenticated actor replays the same normalized patch with the same `(record_id, client_txn_id)`. | The row ends in `status='in_progress'`. | Replay returns the originally committed result and creates no second lifecycle transition, `change_set`, revision entry, or replayable `record_changed` event. |
| `LR-03` | `decision.status` | `base` | A `decision` row is visible and persisted with `status='proposed'`. | A patch sets `decision.status='approved'`, then the same authenticated actor replays the same normalized patch with the same `(record_id, client_txn_id)`. | The row ends in `status='approved'`. | Replay returns the originally committed result and creates no second lifecycle transition, `change_set`, revision entry, or replayable `record_changed` event. |
| `LR-04` | `object_blobs.upload_state` | `base` | An evidence row is visible, a blob slot is already `upload_state='available'`, and the blob is attachable to that evidence row. | `POST /api/v1/evidence-records/{record_id}/attach-blob` succeeds once, then the same authenticated actor replays the same normalized request with the same `(record_id, client_txn_id)`. | The evidence row remains linked to one blob in `upload_state='available'`. | Replay returns the originally committed result, creates no second evidence attachment or `change_set`, and does not consume retry budget. |
| `LR-05` | `evidence_records.lifecycle_state` | `base` | An evidence row is visible and persisted with `lifecycle_state='received'` and a linked blob in `upload_state='available'`. | A patch sets `evidence.lifecycle_state='available'`, then the same authenticated actor replays the same normalized patch with the same `(record_id, client_txn_id)`. | The row ends in `lifecycle_state='available'`. | Replay returns the originally committed result and creates no second lifecycle transition, `change_set`, revision entry, or replayable `record_changed` event. |
| `LR-05b` | `evidence_records.lifecycle_state` | `base` | An evidence row is visible and persisted with `lifecycle_state='available'` and a linked blob in `upload_state='available'`. | A patch sets `evidence.lifecycle_state='released'`, then the same authenticated actor replays the same normalized patch with the same `(record_id, client_txn_id)`. | The row ends in `lifecycle_state='released'`. | Replay returns the originally committed result and creates no second lifecycle transition, `change_set`, revision entry, or replayable `record_changed` event. |
| `LR-06` | `release_state` | `snapshot_reporting` | A release row is visible and persisted with `release_state='approved'`. | `POST /api/v1/releases/{release_id}/publish` succeeds once, then the same authenticated actor replays the same normalized request with the same route-scoped idempotency key. | The row ends in `release_state='published'`. | Replay returns the originally committed result and creates no second publish transition, no second approval or publication timestamp mutation, and no second durable release-state change. |
| `LR-07` | `reference_pack.version_condition` | `reference_pack` | The addressed pack version is visible, in durable condition `verified_available`, and not currently active for its `pack_key`. | `POST /api/v1/reference-packs/{pack_key}/{pack_version}/activate` succeeds once, then the same authenticated actor replays the same normalized request with the same route-scoped idempotency key. | The addressed version becomes active exactly once for its `pack_key`. | Replay returns the originally committed result and creates no second activation, no second active-pointer move, and no second attestation mutation. |

- **AC-102**: A blob slot left in `pending` without successful finalization does not create or imply an attached evidence record, does not increment visible evidence counts, and by no later than the first cleanup sweep after `pending_expires_at` transitions to `upload_state='failed'` with `terminal_reason='pending_timeout'`.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-120
- **AC-103**: If an evidence record has no linked blob, preview and download handle issuance fail with `409 error.code='evidence_access_unavailable'` and `error.details.reason_code='no_visible_blob'`. If the linked blob is in `upload_state='pending'`, `upload_state='failed'`, or is missing, issuance fails with the exact `reason_code` `blob_pending`, `blob_failed`, or `blob_missing`. If the evidence or blob state is quarantined or inconsistent, issuance fails with the exact `reason_code` `evidence_quarantined` or `evidence_inconsistent`. While any of those conditions remains true, the row MUST NOT surface as `evidence.lifecycle_state='available'` or `released`, and ordinary preview and download remain blocked until repaired or re-finalized.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128
- **AC-154**: A declared-size mismatch causes explicit finalization to fail immediately, transitions the slot to `upload_state='failed'` with `terminal_reason='declared_size_mismatch'`, and creates no attached evidence; an expected-hash mismatch does the same with `terminal_reason='expected_sha256_mismatch'`. These terminal mismatches do not consume the ordinary retry budget. For other explicit finalization failures, a pending blob slot allows 3 failed attempts; the 4th such failed attempt transitions the slot to `upload_state='failed'` with `terminal_reason='finalize_retry_exhausted'`; only those non-terminal failed explicit finalization attempts count toward this total, and an idempotent replay after already-committed success does not consume retry budget.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-126
- **AC-155**: Pending-blob timeout handling runs at least every 15 minutes; by no later than `pending_expires_at + 15 minutes`, an unfinalized slot is failed; and for any failed unattached slot, including timeout, retry exhaustion, or terminal size/hash mismatch, orphaned bytes are deleted within 1 hour of terminal failure and failed unattached slot metadata remains queryable for at least 7 days before automatic hard deletion.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-126
- **AC-104**: Rendering a report or presentation artifact creates a release record in `pending_approval` bound to one immutable release tuple and one `output_sha256`.
  - Verifies: REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035
- **AC-105**: Satisfying the required approval set moves that exact release record to `approved`, and any attempted publish action on a non-approved record is rejected.
  - Verifies: REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035
- **AC-106**: Rendering a superseding artifact for the same logical output slot with a different bound tuple or different `output_sha256` creates a new `pending_approval` candidate and invalidates the prior current artifact rather than inheriting its approval or publication state.
  - Verifies: REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035
- **AC-107**: For every normative lifecycle machine claimed by the implementation, the happy-path transitions succeed only in the documented order and the persisted authoritative state matches the documented machine condition after each step.
  - Verifies: REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126
- **AC-108**: For every row in Table 9.9-A whose `profiles` are satisfied by the claim, executing the stated action from the stated starting condition yields the exact committed state and blocked follow-on result declared for that row.
  - Verifies: REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126
- **AC-109**: Illegal lifecycle transitions are rejected with no partial state advancement and no false success signal.
  - Verifies: REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126
- **AC-110**: For every row in Table 9.9-B whose `profiles` are satisfied by the claim, executing the stated successful transition and then replaying the same action under the stated replay condition yields the exact committed state and exact no-duplicate durable-side-effect rule declared for that row.
  - Verifies: REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126
- **AC-111**: Replaying the same starting state and the same ordered inputs for a normative lifecycle machine produces the same final persisted state and the same documented observable signals.
  - Verifies: REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126
- **AC-313**: For a normative lifecycle machine whose condition depends on more than one persisted source or on another lifecycle machine's result, a contradiction fixture produces the documented fail-closed outcome, leaves authoritative state unchanged except for the documented repair action, and does not allow UI-local state, background-worker-local state, or other non-contracted state to change the derived machine condition.
  - Verifies: REQ-00-016, REQ-00-019, REQ-02-189..REQ-02-196, REQ-03-121..REQ-03-126
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
- **AC-144**: If an explicit supersession action targets a decision already in `executed`, the target record remains `executed`, the supersession relation still persists, and the decision view surfaces `decision.is_superseded=true` for that target without rewriting its persisted `status`.
  - Verifies: REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223,
    REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260
- **AC-145**: Replaying the same legal `task_request` status change, legal `decision` status change, or legal explicit supersession action after a simulated crash or duplicate delivery is idempotent and does not duplicate durable side effects such as extra `change_set` records, extra mutation entries, projection updates, or repeated status flips.
  - Verifies: REQ-01-336..REQ-01-341, REQ-02-094..REQ-02-119, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110,
    REQ-03-255..REQ-03-260
- **AC-314**: For `decision.status`, a fixture whose persisted `status` and authoritative decision-to-decision `record_links` relation using `link_type='supersedes'`, `owner_user_id`, and `decided_at` do not resolve to exactly one legal machine condition is treated as inconsistent; ordinary direct status writes and ordinary explicit supersession actions fail closed while the inconsistency remains; and replay of the same repaired fixture is deterministic.
  - Verifies: REQ-02-114

### 9.10 Additional Base Profile criteria for public interface surface

- **AC-123**: `GET /api/v1/auth/session` returns the common success envelope with
  `data` equal to one session resource, without requiring client-side token
  parsing. That resource includes, at minimum, `user_id`, `display_name`,
  `provider_type`, `mfa_state`, `is_deployment_admin`, `authenticated_at`,
  `idle_expires_at`, `absolute_expires_at`, `session_expires_at`, and
  `memberships[]`; `provider_type` uses only `local`, `oidc`, or `saml`;
  `mfa_state` uses only `not_required` or `satisfied`; `session_expires_at` is
  the earlier of `idle_expires_at` and `absolute_expires_at`; `memberships[]`
  is always present, may be empty, is ordered by `incident_id asc`, and each
  item contains `incident_id` and `role` with `role` in `viewer`, `editor`,
  `reviewer`, or `admin`; the route does not return a current-incident-only
  alternate shape; and, because it is singleton, it rejects `limit`,
  `cursor_token`, and pagination aliases with `400
  error.code='invalid_pagination_request'` and
  `error.details.reason_code='pagination_not_supported'`.
  - Verifies: REQ-00-014, REQ-01-023..REQ-01-031, REQ-04-001..REQ-04-017
- **AC-244**: `POST /api/v1/auth/login` with valid email-form `username` and `password` for a non-MFA local account whose stored password includes non-ASCII characters and significant leading or trailing whitespace succeeds, returns the same session resource exposed by `GET /api/v1/auth/session`, and requires no `client_txn_id`.
  - Verifies: REQ-01-025, REQ-01-120, REQ-01-521
- **AC-245**: `POST /api/v1/auth/login` with unknown email-form `username`, wrong `password`, or an inactive local account returns `401 error.code='invalid_credentials'`, sets no session cookie, and does not expose `required_second_factor_kinds` or other evidence that primary credentials were valid; for a non-MFA local account whose stored password includes non-ASCII characters and significant leading or trailing whitespace, a `password` value that differs from the stored secret only by Unicode normalization or by trimmed surrounding whitespace also returns `401 error.code='invalid_credentials'`.
  - Verifies: REQ-01-025, REQ-01-120, REQ-01-234, REQ-01-521
- **AC-246**: `POST /api/v1/auth/login` with valid primary credentials for an MFA-required local account that already has one active TOTP credential and omitted `second_factor` returns `401 error.code='mfa_required'`, includes `error.details.required_second_factor_kinds=["totp"]`, and sets no session cookie.
  - Verifies: REQ-01-025, REQ-01-234
- **AC-247**: `POST /api/v1/auth/login` with a `username` value that fails `email_address_v1`, `second_factor=null`, an unknown top-level member, an unknown `second_factor` or `assertion` member, `second_factor.kind='webauthn'`, or a TOTP assertion whose `code` is missing or not exactly six ASCII decimal digits returns `400 error.code='invalid_auth_request'`; when exactly one member is responsible, `error.details.field` identifies that member.
  - Verifies: REQ-01-025, REQ-01-234, REQ-01-497
- **AC-248**: `POST /api/v1/auth/login` with valid email-form `username`, valid primary credentials, and valid `second_factor.kind='totp'` plus `second_factor.assertion.code` on an MFA-required local account that already has one active TOTP credential succeeds and returns the same session resource exposed by `GET /api/v1/auth/session`.
  - Verifies: REQ-01-025
- **AC-249**: `POST /api/v1/auth/login` with valid primary credentials plus a structurally valid but wrong or expired TOTP code on an MFA-required local account that already has one active TOTP credential returns `401 error.code='invalid_second_factor'` and sets no session cookie.
  - Verifies: REQ-01-025, REQ-01-234
- **AC-250**: `POST /api/v1/auth/login` with any of the forbidden top-level members `client_txn_id`, `id_token`, `authorization_code`, `saml_response`, or `provider_assertion`, or with any other unknown top-level member such as a WebAuthn ceremony field, returns `400 error.code='invalid_auth_request'` and is not interpreted as local login or provider-backed sign-in on that route.
  - Verifies: REQ-01-025, REQ-01-031, REQ-01-234
- **AC-334**: `POST /api/v1/auth/login` with valid primary credentials for a local account where `mfa_required=true` and no active TOTP credential is enrolled returns `401 error.code='mfa_setup_required'`, includes `error.details.required_setup_kinds=["totp"]`, includes one opaque `bootstrap_token` plus `bootstrap_expires_at`, and sets no session cookie.
  - Verifies: REQ-01-024..REQ-01-025, REQ-01-234, REQ-01-522, REQ-04-083, REQ-04-084
- **AC-335**: `GET /api/v1/auth/credential-state` succeeds only for an authenticated current session, returns only `user_id`, `auth_kind`, `mfa_required`, `recovery_model`, nullable `password.changed_at`, and `totp.state` plus nullable enrollment or pending-expiry timestamps, uses `auth_kind='local'`, `recovery_model='admin_assisted'`, and one of `not_enrolled`, `pending`, or `active` for `totp.state`, and exposes no TOTP seed material, `otpauth_uri`, password hash, or raw bootstrap token. A bootstrap token presented on this route fails with `409 error.code='credential_bootstrap_rejected'` and `error.details.reason_code='not_allowed_for_route'`.
  - Verifies: REQ-01-024, REQ-01-522, REQ-01-523, REQ-02-222, REQ-02-245, REQ-04-083, REQ-04-084
- **AC-336**: `POST /api/v1/auth/mfa/totp/begin` accepts exactly one auth mode: either an authenticated current session or one valid `bootstrap_token`; when the caller already has one active TOTP factor and is using session auth, the route additionally requires `current_password` and a current TOTP assertion before issuing a replacement seed; a successful response returns `enrollment_id`, `expires_at`, and `totp_setup.secret_base32`, `totp_setup.otpauth_uri`, `totp_setup.algorithm='SHA1'`, `totp_setup.digits=6`, and `totp_setup.period_seconds=30`; the seed material appears only on this begin response; and replay with the same auth scope plus `client_txn_id` before expiry returns the original pending enrollment and same seed material.
  - Verifies: REQ-01-024, REQ-01-234, REQ-01-522, REQ-01-525, REQ-02-245, REQ-04-083, REQ-04-084, REQ-04-086
- **AC-337**: `POST /api/v1/auth/mfa/totp/complete` with one valid pending `enrollment_id`, the matching auth mode, and a correct six-digit TOTP `code` activates the pending factor, clears pending setup state, and consumes any bootstrap token used for the flow; when the flow is first-time bootstrap it does not auto-issue a session and the user must log in normally afterward; when the flow replaces an existing active factor it revokes all active sessions for that user. If the addressed enrollment is missing, expired, or already consumed, the route fails with `409 error.code='totp_setup_not_pending'` and the corresponding reason code.
  - Verifies: REQ-01-024, REQ-01-234, REQ-01-238, REQ-01-522, REQ-01-526, REQ-02-245, REQ-04-083, REQ-04-084, REQ-04-086
- **AC-338**: `POST /api/v1/auth/password/change` is self-service and current-user scoped; it requires `client_txn_id`, `current_password`, and `new_password`; `new_password` follows `local_password_provision_v1`; when the current account has one active TOTP credential the route also requires a current TOTP assertion; a wrong current password fails with `409 error.code='invalid_current_password'`; and a successful change updates `password.changed_at`, revokes all active sessions including the current session, and forces ordinary re-login with the new password.
  - Verifies: REQ-01-024, REQ-01-234, REQ-01-522, REQ-01-524, REQ-02-245, REQ-04-016, REQ-04-083, REQ-04-084, REQ-04-086
- **AC-339**: A `bootstrap_token` is not accepted by `GET /api/v1/auth/session`, `GET /api/v1/auth/credential-state`, ordinary incident or record routes, or `/ws/v1/*`. Attempting to use such a token outside `POST /api/v1/auth/mfa/totp/begin` or `POST /api/v1/auth/mfa/totp/complete` fails closed with `409 error.code='credential_bootstrap_rejected'` and `error.details.reason_code='not_allowed_for_route'`; successful TOTP bootstrap or replacement consumes or supersedes the bootstrap token so it cannot be reused.
  - Verifies: REQ-01-522..REQ-01-526, REQ-01-234, REQ-01-238, REQ-02-245, REQ-04-083, REQ-04-084
- **AC-311**: After creating one local user with a specific `email`, `POST /api/v1/auth/login` succeeds when `username` differs from the stored email only by `email_address_v1` normalization, such as leading or trailing Unicode whitespace, Unicode NFC-equivalent form, or letter case; a `username` that fails `email_address_v1` returns `400 error.code='invalid_auth_request'` rather than `401 error.code='invalid_credentials'`.
  - Verifies: REQ-01-025, REQ-01-497
- **AC-312**: After creating a local user and successfully `PATCH`ing that user's `email` with the correct `base_user_version`, the returned safe user resource shows the new `email`, the stable `user_id` is unchanged, no second user or independent local-login binding is created, `POST /api/v1/auth/login` succeeds with the new email-form `username`, and the old email-form `username` fails with `401 error.code='invalid_credentials'`.
  - Verifies: REQ-01-119..REQ-01-120, REQ-01-122, REQ-01-124, REQ-01-497
- **AC-251**: `POST /api/v1/evidence-records/{record_id}/preview-handle` and `POST /api/v1/evidence-records/{record_id}/download-handle` accept `{}` as a legal request body and reject a zero-length body, `null`, any non-object JSON value, or any unknown top-level member with `400 error.code='invalid_evidence_handle_request'`. In the base profile, `client_txn_id` is one such invalid member and is rejected rather than interpreted as an issuance idempotency key.
  - Verifies: REQ-01-032, REQ-01-234, REQ-01-459, REQ-01-465
- **AC-252**: For previewable evidence, `POST /api/v1/evidence-records/{record_id}/preview-handle` returns the standard success envelope with `incident_id`, `record_id`, `object_blob_id`, `handle_kind='preview'`, an opaque same-origin `href`, `method='GET'`, `expires_at`, `single_use=false`, `media_class`, `preview_kind`, `disposition='inline'`, `filename`, `content_type`, `size_bytes`, `sha256`, `evidence_lifecycle_state`, and `upload_state`; `expires_at` is exactly 5 minutes after issuance; two back-to-back successful preview-handle issuances for the same evidence return distinct handles; redeeming the handle keeps the current workbook surface loaded, does not force full-page navigation away from the grid, and when preview is blocked the grid or inspector remains in place with inline blocked-state feedback; and when safe preview is not allowed, issuance fails with `409 error.code='evidence_access_unavailable'` and `error.details.reason_code='unsupported_preview'`.
  - Verifies: REQ-01-032, REQ-01-234, REQ-01-238, REQ-01-247, REQ-01-458, REQ-01-460..REQ-01-461, REQ-01-465, REQ-02-222..REQ-02-223, REQ-03-127..REQ-03-128, REQ-04-053
- **AC-253**: For downloadable evidence, `POST /api/v1/evidence-records/{record_id}/download-handle` returns the standard success envelope with `incident_id`, `record_id`, `object_blob_id`, `handle_kind='download'`, an opaque same-origin `href`, `method='GET'`, `expires_at`, `single_use=true`, `media_class`, `disposition='attachment'`, `filename`, `content_type`, `size_bytes`, `sha256`, `evidence_lifecycle_state`, and `upload_state`, omits `preview_kind`, sets `expires_at` exactly 2 minutes after issuance, and returns distinct handles on two back-to-back successful issuances for the same evidence.
  - Verifies: REQ-01-032, REQ-01-234, REQ-01-247, REQ-01-458, REQ-01-460, REQ-01-462, REQ-01-465, REQ-02-222..REQ-02-223, REQ-03-128
- **AC-254**: A preview handle can be redeemed multiple times before expiry, including byte-range reads; a download handle is consumed by the first successful redeem that starts byte delivery; a second redeem of that download handle fails with `410 error.code='handle_consumed'`; an expired redeem fails with `410 error.code='handle_expired'`; and a caller who loses session validity or incident membership after issuance cannot redeem the handle successfully.
  - Verifies: REQ-01-032, REQ-01-234, REQ-01-247, REQ-01-458, REQ-01-462..REQ-01-463, REQ-01-465, REQ-03-128, REQ-04-023
- **AC-255**: If a handle is issued before blob detach, pending or failed transition, missing backing object, quarantine, evidence delete or restore, or detected evidence/blob inconsistency, redeeming that same handle later fails closed with `409 error.code='evidence_access_unavailable'` and the correct `reason_code`; when preview is blocked for one of those reasons, the workbook remains in place and surfaces the blocked state inline rather than silently falling back to download.
  - Verifies: REQ-01-032, REQ-01-234, REQ-01-238, REQ-01-247, REQ-01-459, REQ-01-463, REQ-01-465, REQ-03-127, REQ-04-023, REQ-04-053
- **AC-256**: Issuance returns sanitized `filename` and `disposition`; preview redeem uses `Content-Disposition: inline`; download redeem uses `Content-Disposition: attachment`; each disposition header includes both `filename=` and `filename*=` parameters; and when the authoritative filename is empty or unusable after sanitization, the fallback name is `evidence-<record_id><canonical_extension_if_known>`.
  - Verifies: REQ-01-460, REQ-01-464
- **AC-124**: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` accepts a field-key-based sort, filter, and grouping contract and returns `rows[]` serialized as `view_row_v1` objects with top-level `record_id` and `row_version`, schema-declared non-technical fields in `cells`, full `group_values` for schemas that declare grouping keys, and cursor metadata; group headers are not serialized as writable rows.
  - Verifies: REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-285..REQ-01-290,
    REQ-01-307..REQ-01-341, REQ-01-349, REQ-03-223..REQ-03-224, REQ-03-236..REQ-03-241
- **AC-125**: View-scoped row creation and record-scoped patch operate via `view_schema_id`, `client_txn_id`, `base_row_version`, and `changes[]` keyed by `field_key`; row-create accepts only a JSON object whose top-level namespace is required `client_txn_id` plus `field_key` members allowed for create by the addressed view, and unknown top-level members fail with `400 error.code='invalid_mutation_payload'`; a first-time successful row create returns `201 Created`; replay by the same actor with the same `(incident_id, view_schema_id, client_txn_id)` and the same normalized request returns `200 OK` with the originally committed create result and no second surviving row, `change_set`, or replayable collaboration event; same-scope key reuse with a different normalized request fails with `409 error.code='client_txn_conflict'`; successful row-refreshing create and patch responses return `data.view_schema_id`, `data.change_set_id`, and `data.row=<view_row_v1>` rather than a bespoke refresh payload; replay returns the original committed `data.row`; and the client can mutate one writable field without resubmitting a full row snapshot.
  - Verifies: REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-057..REQ-01-088, REQ-01-285..REQ-01-290,
    REQ-01-307..REQ-01-330, REQ-01-349..REQ-01-350, REQ-02-208..REQ-02-209, REQ-03-086, REQ-03-111..REQ-03-115, REQ-03-236..REQ-03-241
- **AC-151**: `GET /api/v1/incidents/{incident_id}/saved-views` returns only saved views visible to the caller, and each returned resource includes `saved_view_id`, `incident_id`, `view_schema_id`, `scope`, `display_name`, `query_json`, `layout_json`, `owner_user_id`, timestamps, and `saved_view_version`; `query_json` uses stable `field_key` and grouping identifiers rather than visible labels, uses only top-level members `sort`, `filters`, and optional `group_by`, persists `sort=[]` and `filters=[]` as the canonical empty states, omits `group_by` when grouping is inactive, and never stores `record_id` or `row_version`. `layout_json` is the canonical `cartulary.layout.v1` object with required members `layout_schema_id`, `column_order`, `hidden_field_keys`, and `column_widths`; persisted `layout_json` never remains `{}`; `column_order` is a full permutation of the active schema's non-technical field keys; `hidden_field_keys` is a unique sorted subset; and `column_widths` is a unique `field_key`-sorted list. The route orders results by `updated_at desc, saved_view_id asc`; omitting `limit` yields `meta.paging.limit=100`; terminal pages use `meta.paging.has_more=false` and `meta.paging.next_cursor=null`; invalid `limit` values or aliases such as `page`, `offset`, `page_size`, and `block_size` fail closed with `400 error.code='invalid_pagination_request'`; and cursor replay against a different bound contract fails closed. Absence of `cartulary.view.task_requests.v1` or `cartulary.view.decisions.v1` from this route does not imply that the required base surfaces are unavailable, because those surfaces remain discoverable through the authoritative `view_schema` registry and directly addressable by `sheet_ref.kind='view_schema'`.
  - Verifies: REQ-01-138..REQ-01-151, REQ-01-240..REQ-01-242, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021
- **AC-152**: `POST /api/v1/incidents/{incident_id}/saved-views` accepts only a JSON object containing required `view_schema_id`, required non-null `display_name`, required non-null `query_json`, optional non-null `layout_json`, and optional `scope`; `display_name` is validated under `display_name_line_v1`, so it must remain non-empty after Unicode NFC normalization and trimming, reject C0/C1 control characters, and fail closed above 256 Unicode scalar values; omitting `scope` defaults it to `private`; omitting `layout_json` or supplying `{}` normalizes to the canonical schema-derived `cartulary.layout.v1` default object rather than returning `{}`; non-empty `layout_json` must satisfy the closed layout grammar and rejects unknown top-level or nested members; `query_json` is validated and normalized using the same sort, filter, and grouping rules as the public view-query route, omission of `query_json.sort` or `query_json.filters` normalizes to persisted `[]`, and explicit `query_json.group_by=null` fails with `400 error.code='invalid_mutation_payload'`; `scope='system'`, explicit `null` for `display_name`, `query_json`, `layout_json`, or `scope`, any `query_json` or `layout_json` field reference not declared by the addressed `view_schema_id`, any use of `record_id` or `row_version` inside those objects, and unknown top-level members fail with `400 error.code='invalid_mutation_payload'` plus `error.details.field` naming the offending path; `PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` accepts `base_saved_view_version` plus mutable fields only, treats omitted mutable fields as unchanged, rejects unknown top-level members and explicit `null` for `display_name`, `query_json`, or `layout_json`, applies the same query-json and layout-json validation and normalization rules, compares saved-view no-op semantics after normalized `display_name`, `query_json`, and `layout_json` structural equality rather than textual JSON equality, returns `200 OK` without advancing `saved_view_version` or `updated_at` for a structurally valid no-op, and a stale base version fails with an explicit conflict status rather than silently overwriting saved-view state. Changing shared-layout interpretation without changing `layout_schema_id` is non-conformant. `DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` deletes only the configuration object and never underlying incident records or links.
  - Verifies: REQ-01-138..REQ-01-151, REQ-01-487..REQ-01-489, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-016, REQ-03-022..REQ-03-026
- **AC-153**: `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` read or replace only the caller's nullable `home_sheet_ref`; the `PUT` route accepts only `{ "home_sheet_ref": <sheet_ref|null> }`, creates the preference object if absent, replaces only that pointer if present, rejects unknown top-level members with `400 error.code='invalid_mutation_payload'`, succeeds for any incident member including `viewer`, and a structurally valid no-op leaves `updated_at` unchanged; `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` read or replace only the incident's nullable `default_sheet_ref`; the `PUT` route accepts only `{ "default_sheet_ref": <sheet_ref|null> }`, creates the preference object if absent, replaces only that pointer if present, rejects unknown top-level members with `400 error.code='invalid_mutation_payload'`, fails closed for non-admin incident roles, and a structurally valid no-op leaves `updated_at` and `updated_by_user_id` unchanged. When the selected surface is any pack-independent base-profile surface from `REQ-01-307`, including `cartulary.view.task_requests.v1`, `cartulary.view.decisions.v1`, `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`, `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`, the persisted pointer uses only `sheet_ref.kind='view_schema'` with the standardized `view_schema_id` for the required base surface itself; a user-created saved view over the same schema persists distinctly as `sheet_ref.kind='saved_view'` with its own `saved_view_id`.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032
- **AC-126**: Same-field conflict responses use the generic error envelope with `error.code='same_field_conflict'` and the conflict object required by Core 03 §3.3.4; a stale `conflict_token` is rejected with a fresh conflict payload rather than silently overwriting saved state.
  - Verifies: REQ-01-019, REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-03-036..REQ-03-040,
    REQ-03-063..REQ-03-076
- **AC-200**: Patching `timeline.tags` with one `add_tag` action and one `remove_tag` action adds exactly one incident-scoped tag binding and removes exactly one binding; `add_tag.tag_name` is validated under `tag_label_v1`, duplicate adds coalesce using trimmed Unicode NFC plus case-insensitive comparison, normalized-empty values are rejected, C0/C1 control characters are rejected, and values longer than 64 Unicode scalar values fail closed.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-487..REQ-01-488, REQ-01-494, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209
- **AC-188**: Patching `timeline.host_refs` or `timeline.identity_refs` with `dismiss_item` against an unresolved or resolved `entity_mention` preserves `raw_text`, leaves the same mention row and provenance intact, sets `resolution_status='dismissed'`, clears `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`, removes or tombstones any corresponding active resolved link in the same `change_set`, and appends exactly one new `change_set` with mention-target mutation detail.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-02-058..REQ-02-059, REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216,
    REQ-03-236..REQ-03-241
- **AC-189**: Patching `timeline.host_refs` or `timeline.identity_refs` with `revert_to_unresolved` against a dismissed `entity_mention` returns that same mention row to `resolution_status='unresolved'`, preserves `raw_text`, leaves `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method` null after commit, leaves no active derived resolved link, and appends exactly one new `change_set`; reviewer rollback of the dismissal restores the exact pre-dismiss state as a separate attributed revision rather than rewriting history.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241
- **AC-190**: If a row's last non-deleted unresolved host or identity mention is dismissed, the committed row computes `timeline.has_unresolved_mentions=false`, the row no longer matches a `timeline.has_unresolved_mentions` filter with `eq` and `arg.value=true`, grouping by `timeline.has_unresolved_mentions` places the row in the `false` bucket, and the active `timeline.host_refs` or `timeline.identity_refs` collection value omits the dismissed mention while history or inspector affordances can still surface it.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241
- **AC-201**: Patching `timeline.host_refs` with `resolve_item` preserves `raw_text`, resolves the targeted `entity_mention`, and leaves the corresponding active host link present after commit.
  - Verifies: REQ-01-057..REQ-01-088, REQ-02-030..REQ-02-036, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209
- **AC-221**: `POST /api/v1/entity-mentions/{entity_mention_id}/resolve` with JSON `base_mention_row_version`, `client_txn_id`, `action='resolve_item'`, and `resolved_record_id` by an `editor`, `reviewer`, or `admin` on the source incident resolves one visible active host or identity mention, returns `200 OK` with `incident_id`, updated `entity_mention`, `source_record.record_id`, incremented `entity_mention.row_version`, incremented `source_record.row_version`, `change_set_id`, and `active_link`, preserves `raw_text`, and when optional `reason` is supplied it is normalized under `reason_note_v1` with omission, explicit `null`, and normalized-empty reason treated equivalently for idempotency and persistence; the committed mention sets `resolution_status='resolved'`, sets non-null `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method='explicit_resolve_route'`, and when the mention was previously resolved to a different target, removes or tombstones the old active resolved link in the same `change_set`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-026..REQ-02-036,
    REQ-02-039..REQ-02-044, REQ-03-129..REQ-03-134
- **AC-222**: The same route with JSON `action='dismiss_item'` against an unresolved or resolved visible active host or identity mention returns `200 OK`, preserves `raw_text`, stable mention identity, and provenance, sets `resolution_status='dismissed'`, returns `resolved_record_id=null`, `resolved_by_user_id=null`, `resolved_at=null`, and `resolution_method=null` on the committed `entity_mention`, omits `active_link`, and removes or tombstones any corresponding active resolved link in the same `change_set`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-03-129..REQ-03-134
- **AC-223**: The same route with JSON `action='revert_to_unresolved'` against a resolved or dismissed visible active host or identity mention returns `200 OK`, preserves `raw_text`, sets `resolution_status='unresolved'`, returns `resolved_record_id=null`, `resolved_by_user_id=null`, `resolved_at=null`, and `resolution_method=null` on the committed `entity_mention`, omits `active_link`, and when the starting state was `dismissed`, does not silently relink any prior resolved target; exact pre-dismiss recovery remains reviewer rollback.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044,
    REQ-03-129..REQ-03-134
- **AC-224**: The same route with a stale `base_mention_row_version` fails with `409 error.code='row_version_conflict'` and `error.details` containing `entity_mention_id`, `base_mention_row_version`, `current_mention_row_version`, and `source_record_id`; a direct `resolve_item` against a currently dismissed mention fails with `409 error.code='illegal_transition'`; a request against a soft-deleted source record fails with `409 error.code='record_deleted_use_restore'`; a caller lacking visibility to the mention receives `404 error.code='entity_mention_not_found'`; a caller who can see the mention but lacks `editor`, `reviewer`, or `admin` role receives `403`; a supplied `resolved_record_id` that does not identify a visible active target fails with `404 error.code='resolved_record_not_found'`; and missing required members, unknown top-level members, forbidden or missing `resolved_record_id`, wrong target type, target from another incident, or embedded entity-create payload fail with `400 error.code='invalid_mutation_payload'`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-030..REQ-02-036, REQ-02-039..REQ-02-041,
    REQ-03-129..REQ-03-134
- **AC-225**: Replaying `POST /api/v1/entity-mentions/{entity_mention_id}/resolve` with the same normalized request by the same actor and the same `(entity_mention_id, client_txn_id)` returns the originally committed `200 OK` success without creating a second `change_set`; for optional `reason`, omission, explicit `null`, and normalized-empty input compare equal under `reason_note_v1`; reusing that key with a different normalized request fails with `409 error.code='client_txn_conflict'`; and a successful explicit mention action reaches subscribers only through the ordinary `record_changed` event for the source record with no mention-specific WebSocket family.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-030..REQ-02-036,
    REQ-02-039..REQ-02-041, REQ-03-129..REQ-03-134
- **AC-202**: Patching `host.aliases` or `identity.aliases` with `add_alias` plus `remove_alias` yields a committed alias collection that round-trips as `collection_value_v1`; `add_alias.alias_text` is validated under `alias_text_v1`, duplicate adds coalesce per canonical record using trimmed Unicode NFC plus case-insensitive comparison, normalized-empty values are rejected, C0/C1 control characters are rejected, and values longer than 256 Unicode scalar values fail closed.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-487..REQ-01-488, REQ-01-495, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209
- **AC-203**: A same-field conflict payload for a `collection_review` field returns `collection_value_v1` in `client_value`, `server_value`, and `base_value` rather than a raw string array or plain delimited text.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209,
    REQ-03-048..REQ-03-053, REQ-03-063..REQ-03-076
- **AC-204**: Sending a raw string, blind full-collection replacement, unknown collection action `op`, unknown payload `kind`, or foreign `item_ref` to a `collection_review` field fails with `400` and `error.code='invalid_mutation_payload'`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209,
    REQ-03-048..REQ-03-053, REQ-03-063..REQ-03-076
- **AC-205**: When an interactive host or identity token on `timeline.host_refs` or `timeline.identity_refs` qualifies for auto-resolution under Core 03 §12, the committed action resolves the corresponding `entity_mention` and produces exactly one active `record_link` whose `link_type` is `observed_on_host` or `observed_as_identity` according to the mutated field, whose `provenance` is `auto_match`, and whose `confidence` is `100`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568,
    REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279
- **AC-388**: With one active host alias `VPN Gateway` and no competing candidate, interactive inline commit of submitted token ` vpn   gateway ` on `timeline.host_refs` resolves exactly one `entity_mention` to that host and creates exactly one active `record_link` with `provenance='auto_match'` and `confidence=100`, proving that auto-resolution comparison uses `mention_token_text_v1` normalization plus locale-independent Unicode case folding and no broader rewrite.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568,
    REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279
- **AC-389**: With one active host alias `WS-023` and no competing candidate, interactive entry of `WS-023?`, `WS-023??`, or `WS-023 ~` on `timeline.host_refs` creates or preserves only unresolved `entity_mentions` with `resolution_status='unresolved'`, `resolved_record_id=null`, no active resolved `record_link`, and no auto-resolution disclosure.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568,
    REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279
- **AC-390**: With one active host alias `WS-023` and no competing candidate, interactive entry of `WS-023 maybe`, `WS-023 prob`, `WS-023 probably`, `WS-023 approx`, or `WS-023 approximately` on `timeline.host_refs` creates or preserves only unresolved `entity_mentions` with `resolution_status='unresolved'`, `resolved_record_id=null`, no active resolved `record_link`, and no auto-resolution disclosure.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568,
    REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279
- **AC-391**: With one active host alias `WS-023` and no competing candidate, interactive entry of `WS-023 likely` or `(WS-023)` on `timeline.host_refs` creates or preserves only unresolved `entity_mentions` with `resolution_status='unresolved'`, `resolved_record_id=null`, and no active resolved `record_link`, proving that extra-word stripping, token deletion, punctuation stripping, parenthetical stripping, and duplicate-punctuation collapse are non-conformant.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568,
    REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279
- **AC-392**: Even when one active alias exactly matches the submitted token, background jobs, async enrichment, and cleanup do not auto-resolve host or identity tokens and create no `provenance='auto_match'` link through those workflows.
  - Verifies: REQ-02-163..REQ-02-185, REQ-03-208, REQ-03-276
- **AC-393**: In the Import Extension Profile, file-based import of exact-match Timeline Hosts or Identities values preserves unresolved mentions only and creates no `provenance='auto_match'` links, even when the same submitted token would auto-resolve in an eligible interactive Timeline workflow.
  - Verifies: REQ-01-010..REQ-01-014, REQ-02-032, REQ-03-208, REQ-03-276
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
- **AC-394**: A base-profile manual Timeline relationship mutation using either `add_resolved_ref` or `resolve_item` on `timeline.host_refs` or `timeline.identity_refs` commits without any client-supplied `confidence` member and produces exactly one active `record_link` whose `link_type` is field-derived, whose `provenance` is `manual`, and whose `confidence` is `null`.
  - Verifies: REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-163..REQ-02-180, REQ-02-248, REQ-03-280
- **AC-395**: A supported base-profile collection-style relationship mutation using `add_record_ref` on `assessment.support_refs`, `task.linked_record_ids`, `decision.support_refs`, and at least one coordination-surface `record_ref` field such as `comm_log.decision_ids` or `handoff.open_task_ids` commits without any client-supplied `confidence` member and creates or preserves active `record_links` whose `provenance` is `manual` and whose `confidence` is `null`.
  - Verifies: REQ-01-311, REQ-01-333..REQ-01-340, REQ-01-503..REQ-01-506, REQ-02-163..REQ-02-180, REQ-02-248, REQ-03-280
- **AC-396**: Supplying a client-supplied `confidence` member on a base-profile relationship mutation route or relationship action payload, including `timeline.host_refs`, `timeline.identity_refs`, `assessment.support_refs`, `task.linked_record_ids`, `decision.support_refs`, and reused coordination-surface `record_ref` fields, fails with `400` and `error.code='invalid_mutation_payload'`; the server creates no durable `record_link`, persists no client `confidence` value, and leaves no misleading projection update.
  - Verifies: REQ-01-569, REQ-03-280
- **AC-397**: Authoritative stored manual-link state preserves `confidence=null`. Implementations MUST NOT normalize manual-link `confidence` to `0`, `100`, or any other sentinel in current-state projection reads or ordinary export, and when the corresponding extension profiles are claimed, snapshot-source derivation and incident-portability serialization MUST preserve the same `null` value.
  - Verifies: REQ-02-248
- **AC-127**: Potentially large list or view-query routes use opaque cursor pagination with `has_more` and `next_cursor`; replaying a cursor against a different authenticated actor, route family, route-scoping identifier, normalized sort or filter or grouping contract when present, or effective `limit` is rejected rather than reinterpreted.
  - Verifies: REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-117..REQ-01-118, REQ-01-129, REQ-01-144, REQ-01-168, REQ-01-240..REQ-01-242, REQ-01-288
- **AC-238**: On `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` without `cursor_token`, omitting `limit` yields `meta.paging.limit=100`; when more rows remain, the same response uses `meta.paging.has_more=true` and a non-null `meta.paging.next_cursor`.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-242
- **AC-239**: Repeating the same route with a valid `cursor_token` and omitted `limit` reuses the cursor-bound effective `meta.paging.limit`; replaying that cursor with a different explicit `limit`, a different normalized effective `sort[]`, or a different normalized `group_by` fails with `400 error.code='invalid_view_query'` and `error.details.reason_code='cursor_query_mismatch'`.
  - Verifies: REQ-01-035, REQ-01-234, REQ-01-238, REQ-01-241, REQ-01-242
- **AC-240**: `limit=0`, `limit=-1`, `limit=501`, a non-integer `limit`, and use of `page`, `offset`, `block_size`, or `page_size` in the request each fail closed with `400 error.code='invalid_view_query'` and `error.details.reason_code='invalid_limit'`.
  - Verifies: REQ-01-035, REQ-01-234, REQ-01-238
- **AC-241**: A terminal page reached through cursor progression returns `meta.paging.limit`, `meta.paging.has_more=false`, and `meta.paging.next_cursor=null`; a zero-match first page uses the same terminal representation and returns `rows=[]`.
  - Verifies: REQ-01-036, REQ-01-242
- **AC-242**: For paged view-query responses, continuation is determined by `meta.paging.has_more` and `meta.paging.next_cursor`; conformance evidence and client continuation logic do not infer terminal state from `rows.length < meta.paging.limit` alone.
  - Verifies: REQ-01-242
- **AC-243**: With grouping active and `limit=1`, the response serializes exactly one data row in `rows[]` as one full `view_row_v1`, includes no group-header pseudo-row, and still returns the full `group_values` object for that row; page-size accounting applies to serialized `rows[]` entries only.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-037
- **AC-372**: On `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, issue page 1 with a deterministic sort and a finite `limit`, then commit an intervening insert whose live sort position falls before or inside the not-yet-read remainder of the result set. Continuing with the returned `meta.paging.next_cursor` does not surface the inserted row in that existing cursor chain, and the original snapshot rows still appear exactly once with no skip.
  - Verifies: REQ-01-035, REQ-01-554, REQ-01-555, REQ-01-556
- **AC-373**: On `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, issue page 1 with a deterministic sort and a finite `limit`, then commit an intervening delete, restore, or sort- or filter-relevant edit affecting one or more rows that would otherwise appear later in the live result set. Continuing with the returned `meta.paging.next_cursor` preserves the original snapshot membership and snapshot order, does not duplicate rows, and returns later rows using snapshot-state payloads, including the snapshot `row_version` when present, rather than fetch-time live state.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-555, REQ-01-556, REQ-01-557
- **AC-374**: Using the same fixture shape as `AC-372` or `AC-373`, restart the same route without `cursor_token` after the intervening mutations. The fresh request reflects current live membership, current live ordering, and current live row values, demonstrating that continuing an existing cursor chain and starting a fresh query are intentionally different operations.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-554, REQ-01-555, REQ-01-557, REQ-01-558, REQ-01-560
- **AC-375**: Using a non-null `meta.paging.next_cursor` returned by a successful page-1 response from `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, a continuation request issued before that cursor chain exceeds the REQ-01-559 inactivity window succeeds normally. A continuation request issued after that same chain exceeds the REQ-01-559 10-minute minimum inactivity window, with the caller still satisfying current session and authorization checks for the same route contract and with an otherwise unchanged well-formed cursor, fails with `400`, `error.code='invalid_view_query'`, and `error.details.reason_code='cursor_snapshot_unavailable'`. Reissuing the same route without `cursor_token` succeeds normally against current live state.
  - Verifies: REQ-01-035, REQ-01-238, REQ-01-554, REQ-01-558, REQ-01-559, REQ-01-560
- **AC-128**: `POST /api/v1/object-blobs` accepts only a JSON object with required `incident_id`, `client_txn_id`, and `byte_size`, plus optional `filename_hint`, `content_type_hint`, and `sha256_hex`; rejects malformed or unknown-field input with `400 error.code='invalid_blob_create_request'`; rejects same-scope `client_txn_id` reuse with a different normalized request using `409 error.code='client_txn_conflict'`; returns `201 Created` on first success and `200 OK` on same-request replay; and returns `incident_id`, `object_blob_id`, `upload_state`, `target_expires_at`, `pending_expires_at`, `upload_target`, and `accepted_contract`, with omitted optional contract members serialized as explicit `null` inside `accepted_contract`. In the base profile the upload target expires 60 minutes after issuance, the pending slot expires 24 hours after creation, same-request replay after target expiry returns the same expired slot rather than refreshing it, and obtaining a fresh target requires a fresh blob slot. `POST /api/v1/evidence-records/{record_id}/attach-blob` accepts only a JSON object with exactly `object_blob_id`, `base_row_version`, and `client_txn_id`; rejects malformed input, supplied `null` for those required members, or unknown top-level members with `400 error.code='invalid_mutation_payload'`; returns `200 OK` on first success and same-request replay; rejects same-scope key reuse with a different normalized request using `409 error.code='client_txn_conflict'`; rejects stale `base_row_version` with `409 error.code='row_version_conflict'`; fails closed with `409 error.code='evidence_attach_rejected'` and a registered reason code when blob or evidence lifecycle state does not satisfy the attach bridge rules from Core 03 §8 and Core 02 §13; and on success returns `data.view_schema_id='cartulary.view.evidence.v1'`, `data.change_set_id`, `data.object_blob_id`, and `data.row=<view_row_v1>` rather than a bespoke evidence-refresh object.
  - Verifies: REQ-00-014, REQ-01-019, REQ-01-234, REQ-01-238, REQ-01-243..REQ-01-247, REQ-01-328,
    REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128,
    REQ-04-048
- **AC-129**: Long-running operations started through the public interface return `202 Accepted` with a `job_id`; `GET /api/v1/jobs/{job_id}` exposes progress and terminal result or error summary; the incident-scoped WebSocket stream authenticates with the same session contract and emits presence, `record_changed`, and `job_progress` events without broadcasting client-local drafts or grouping UI state; and when a terminal polled job resource or terminal `job_progress` message includes `result_summary.resource_refs[]`, the client treats those refs as a compact navigation summary, renders known current-profile kinds as non-modal result chips or links on the current surface, degrades unknown kinds to message-only rendering without breaking job rendering, and does not auto-follow `route` or change the active workbook surface, selection, or scroll position.
  - Verifies: REQ-00-014, REQ-01-018..REQ-01-019, REQ-01-248..REQ-01-277, REQ-01-452..REQ-01-454,
    REQ-03-092..REQ-03-098
- **AC-257**: A long-running operation started through the public interface returns `202 Accepted` with the canonical job resource in the common success envelope; `data.status` is `queued` or `running`; `data.status_route` is `/api/v1/jobs/{job_id}` for that resource; `scope.kind` is either `incident` or `deployment`; only `queued`, `running`, `cancel_requested`, `succeeded`, `failed`, and `canceled` appear as public job states; and polling `GET /api/v1/jobs/{job_id}` never observes a public state transition outside the legal state machine. Incident-scoped streams emit `job_progress` only for incident-scoped jobs, and deployment-scoped jobs do not appear on incident-scoped streams.
  - Verifies: REQ-01-248..REQ-01-249, REQ-01-268, REQ-04-023
- **AC-258**: Over both `GET /api/v1/jobs/{job_id}` and replayed or live `job_progress` messages, `progress` is always an object with non-negative integer `completed` and `total` equal to either a positive integer or `null`; `completed` never decreases for one job; when `total` is known, `completed <= total`; when `total = null`, the client renders indeterminate progress rather than a fake percent; and when `status='succeeded'` with known `total`, the terminal resource or event has `completed == total`.
  - Verifies: REQ-01-249, REQ-01-268, REQ-01-453
- **AC-259**: Non-terminal job resources carry `result_summary=null` and `error_summary=null`, and any non-terminal `job_progress` message that includes those members does the same; `succeeded` and `canceled` terminal states carry only `result_summary`; `failed` carries only `error_summary`; when `status='canceled'`, `result_summary.code='job_canceled'`; terminal `succeeded` or `canceled` summaries do not carry family-specific deep payloads in place of the common summary; any current-profile `result_summary.resource_refs[]` items use only `incident`, `import_session`, `snapshot`, `release`, `reference_pack_version`, or `incident_bundle`; every current-profile emitted ref includes a canonical same-origin `GET` `route` with no query or fragment; `reference_pack_version.id == route`; any multi-ref current-profile terminal summary uses deterministic ordering, with `reference_packs_refreshed` sorted by `route asc`; `started_at` is `null` until work begins; `finished_at` is `null` before terminal and non-null after terminal; and `retained_until` is `null` before terminal and non-null after terminal on the HTTP job resource and, when present, on `job_progress`.
  - Verifies: REQ-01-249, REQ-01-268
- **AC-260**: `POST /api/v1/jobs/{job_id}/cancel` requires a JSON object with `client_txn_id` and accepts optional `reason`; omission and explicit JSON `null` for `reason` compare equal for normalized request comparison; first success and idempotent replay both return `200 OK` with the current authoritative job resource; same-actor replay of the same normalized request with the same `(job_id, client_txn_id)` creates no second cancel transition; reuse of that scope key with a different normalized request fails with `409 error.code='client_txn_conflict'`; and rejected cancel attempts fail with `409 error.code='job_cancel_rejected'` plus exact `reason_code` equal to `already_cancel_requested`, `already_terminal`, or `not_cancelable`.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-249, REQ-01-453, REQ-04-023
- **AC-261**: Terminal jobs expose `retained_until >= finished_at + 7 days`; `GET /api/v1/jobs/{job_id}` succeeds before expiry and may return `404 error.code='job_not_found'` after expiry; the same `404 error.code='job_not_found'` behavior is used for absent or unauthorized job reads or cancel requests, including a caller with `deployment_admin=true` but no incident membership attempting to access an incident-scoped job; and expiring the job resource does not delete or mutate durable outputs that the job produced.
  - Verifies: REQ-01-234, REQ-01-249, REQ-04-023, REQ-04-029
- **AC-130**: A cookie-authenticated state-changing request with missing or invalid CSRF proof fails closed, and the same request succeeds when a valid CSRF proof accompanies an otherwise authorized session.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-154, REQ-04-001..REQ-04-004, REQ-04-052..REQ-04-053
- **AC-131**: On `GET /ws/v1/incidents/{incident_id}`, an authorized same-origin client that sends `hello` as its first application message receives `hello_ack` containing `connection_id`, `resume_token`, `server_time`, `heartbeat_interval_ms=15000`, `presence_ttl_ms=45000`, and `resume_window_ms>=300000`, followed by `presence_snapshot` whose `presences[]` member is always present, may be empty, excludes expired presence records, contains no duplicate `connection_id`, and serializes remaining entries in canonical ascending `connection_id` order; a cookie-authenticated browser connection from an untrusted `Origin` is rejected before `hello_ack` or `presence_snapshot`, and an otherwise idle accepted connection receives application-level `ping` within 15 seconds and is closed after 45 seconds without any inbound frame or timely `pong`. A Base claim is non-conformant if `GET /ws/v1/incidents/{incident_id}` is absent or if the implementation treats any different v1 public subscription route as equivalent to that canonical route.
  - Verifies: REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098, REQ-04-005..REQ-04-017,
    REQ-04-052..REQ-04-053
- **AC-132**: When analyst A changes workbook surface, focused row, or same-cell edit state on an incident, analyst B subscribed to the same incident receives the corresponding `presence_delta` within 1 second, and the UI renders workbook-header, row-gutter, and same-cell indicators from matching `sheet_ref`, `record_id`, and `field_key` rather than visible labels or row numbers. Presence for direct opening of any pack-independent base-profile surface from `REQ-01-307`, including `cartulary.view.task_requests.v1`, `cartulary.view.decisions.v1`, `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`, `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`, uses `sheet_ref.kind='view_schema'` with the standardized `view_schema_id`; presence for a distinct saved view over the same schema uses `sheet_ref.kind='saved_view'` with that saved view's `saved_view_id`.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-090..REQ-03-098
- **AC-133**: After a transient disconnect, a client that reconnects to the same incident within the retained replay window can send `resume` with `resume_token` and `last_seen_stream_seq`, receives `resume_ack.status='replayed'`, then receives missed `record_changed` and incident-scoped `job_progress` messages in strict ascending `stream_seq`, followed by a fresh `presence_snapshot` whose `presences[]` member is always present, may be empty, excludes expired presence records, contains no duplicate `connection_id`, and serializes remaining entries in canonical ascending `connection_id` order.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098
- **AC-134**: If a client reconnects with an expired, unknown, malformed, or too-old `resume_token`, but the caller still has valid incident authorization, the server responds with `resume_ack.status='reset_required'`, sends a fresh `presence_snapshot` whose `presences[]` member is always present, may be empty, excludes expired presence records, contains no duplicate `connection_id`, and serializes remaining entries in canonical ascending `connection_id` order, and emits no guessed or partial replay for the missing range; the client recovers only by re-querying current workbook state through the existing HTTP view route.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098
- **AC-135**: Replayable WebSocket messages on one incident carry monotonically increasing `stream_seq` values assigned only after the underlying record mutation or incident-scoped job-state change is committed; clients can de-duplicate duplicates by `(incident_id, stream_seq)`, reconcile row application by `record_id` plus `row_version`, must perform HTTP resynchronization rather than guessed incremental apply when a replayable sequence gap is observed, and observe identical canonical `changed_field_keys[]` and `affected_views[]` array order when a replayable `record_changed` with semantically identical array content is delivered live and via replay.
  - Verifies: REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098
- **AC-136**: If the authenticated session expires or the caller's incident membership is revoked after connection establishment, the server sends `session_revoked`, closes the socket, and emits no further incident `presence`, `record_changed`, or `job_progress` messages to that client on that connection.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098, REQ-04-005..REQ-04-017
- **AC-156**: After 30 minutes with no qualifying activity, the next request on that session to `GET /api/v1/auth/session`, query a workbook view, or submit a mutation fails with `401`, and no mutation or job start is partially applied.
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
- **AC-163**: If session expiry or revocation occurs while the client holds queued unsent writes or unresolved same-field local drafts, the client preserves that unsaved work locally within the same browser runtime, prompts for re-authentication when required, and retries only through the normal authenticated create, patch, or conflict-resolution paths after a new session is established; no queued draft becomes authoritative without passing the ordinary row-version, authorization, and conflict checks.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-077..REQ-03-082, REQ-03-099..REQ-03-100, REQ-04-005..REQ-04-017
- **AC-376**: The workbook save-state label uses only `Syncing`, `Saved`, and `Conflict`; `Syncing` is shown when a workbook mutation is in flight or the local pending queue is non-empty, including while replay waits for connectivity recovery, re-authentication, or an HTTP re-query required by `REQ-03-096`; `Saved` is shown only when no workbook mutation is in flight, the local pending queue is empty, and there are no unresolved same-field local drafts; `Conflict` is shown when unresolved same-field local drafts exist, queue overflow has refused admission of a new replay unit, or replay is halted on a non-retryable failure; and presence updates alone do not change the save-state label.
  - Verifies: REQ-03-089, REQ-03-095, REQ-03-099..REQ-03-100
- **AC-377**: In one browser runtime, if queued hot-path writes are blocked by transient disconnect, HTTP auth failure, or `session_revoked`, the client preserves the local pending queue, re-establishes a valid authenticated session, completes any required HTTP re-query, and replays queued writes in FIFO order on return without reordering by visible row order, sort order, or record type.
  - Verifies: REQ-03-099..REQ-03-100
- **AC-378**: For one `(incident_id, client_instance_id)`, the client can hold exactly `64` non-coalescible local pending replay units; when a 65th non-coalescible unit would be admitted, the client refuses that new unit, preserves every already queued unit, preserves the current visible edit as unsaved local work, sets the save state to `Conflict`, shows a same-surface non-modal overflow message, and silently evicts nothing.
  - Verifies: REQ-03-099..REQ-03-100
- **AC-379**: When a still-uncommitted local row has one queued create plus later unsent edits to that same row, reconnect or re-auth replay commits one create unit carrying the final local values rather than a queued create followed by a pre-create patch sequence.
  - Verifies: REQ-03-099..REQ-03-100
- **AC-380**: For an existing authoritative row, same-record unsent patch units coalesce only within one contiguous same-record run in queue order; an interleaving such as `A1, B1, A2` does not replay as a reordered `A1+A2, B1` sequence.
  - Verifies: REQ-03-099..REQ-03-100
- **AC-381**: Replay halts at the first blocking non-retryable failure; if that failure is a same-field conflict, the blocked replay unit leaves the local pending queue and enters the client-local same-field conflict queue keyed by `record_id:field_key`, and later queued units remain queued and unapplied behind it rather than being applied out of order.
  - Verifies: REQ-03-072, REQ-03-099..REQ-03-100
- **AC-382**: A full reload or recreated page instance does not restore or silently replay the base-profile local pending queue. Any later durable local persistence would require a separate explicitly configured feature or profile.
  - Verifies: REQ-03-099..REQ-03-100

- **AC-170**: An authenticated session whose internal user account is active and not disabled can `POST /api/v1/incidents` with required `client_txn_id`, `incident_key`, and `title`, plus optional nullable `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`; on first success it receives `201 Created` plus `Location: /api/v1/incidents/{incident_id}`, and the response `data` includes the incident resource fields defined by Core 01 §3.3.5.3 with `status='active'`, `incident_version=1`, `closed_at=null`, `created_by_user_id` equal to the actor, `updated_by_user_id` equal to the actor, and `created_at == updated_at`.
  - Verifies: REQ-01-152..REQ-01-180, REQ-02-203, REQ-02-243
- **AC-171**: A successful incident create bootstraps exactly one creator membership with `role='admin'`; `GET /api/v1/incidents/{incident_id}/memberships` returns that membership for the creator; `GET /api/v1/incidents/{incident_id}/workbook-preferences/default` returns a resource with `default_sheet_ref=null`; `GET /api/v1/incidents/{incident_id}/workbook-preferences/me` for the creator returns a resource with `home_sheet_ref=null`; and `GET /api/v1/incidents` for that caller returns only incidents where the caller currently has membership in `data.incidents[]`, ordered by `updated_at desc, incident_id asc`, with `meta.paging.limit=100` when `limit` is omitted, `meta.paging.has_more=false` and `meta.paging.next_cursor=null` on terminal pages, `400 error.code='invalid_pagination_request'` for invalid `limit` values or aliases such as `page`, `offset`, `page_size`, and `block_size`, rejection of cursor replay against a different bound contract, and rejection of pagination members on `GET /api/v1/incidents/{incident_id}` with `error.details.reason_code='pagination_not_supported'`.
  - Verifies: REQ-01-152..REQ-01-180, REQ-01-240..REQ-01-242
- **AC-172**: `POST /api/v1/incidents` without authentication, with an expired or revoked session, from a disabled internal user account, or with missing or invalid CSRF protection for a cookie-authenticated request fails closed and creates no incident, membership, or workbook-preference state.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-173**: Replaying `POST /api/v1/incidents` with the same `(actor_user_id, client_txn_id)` and the same normalized request returns `200 OK`, repeats `Location: /api/v1/incidents/{incident_id}` for the originally created incident, returns the originally created incident resource, and creates no second incident; replaying with the same `(actor_user_id, client_txn_id)` and a different normalized request fails with `409` and `error.code='client_txn_conflict'`.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-174**: Creating an incident with an `incident_key` whose trimmed Unicode-NFC-normalized form conflicts with an existing incident fails with `409` and `error.code='incident_key_conflict'`; `error.details.field='incident_key'` and `error.details.incident_key_canonical` are present; and no partial incident, membership, or workbook-preference bootstrap state is left behind.
  - Verifies: REQ-01-152..REQ-01-180, REQ-02-243
- **AC-219**: `POST /api/v1/incidents` whose body is not a JSON object, omits required `client_txn_id`, `incident_key`, or `title`, supplies `null` for a non-nullable field, exceeds the declared field-length limits, includes a control character in `incident_key` or `title`, attempts to set a server-managed field, includes any top-level member outside `client_txn_id`, `incident_key`, `title`, `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`, or sends `initial_memberships[]` fails with `400` and `error.code='invalid_incident_create'`; when one member is responsible, the response includes `error.details.field`, and `error.details.reason_code` uses the Core 01 registry, including `unknown_field` for an undeclared top-level member, `server_managed_field` for a forbidden server-managed member, and `collaborator_seeding_not_supported` for `initial_memberships[]`.
  - Verifies: REQ-01-021, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239
- **AC-220**: On `POST /api/v1/incidents`, idempotency comparison happens only after validation succeeds. A replay that differs only by omission versus explicit `null` for `description`, `severity`, `tlp`, `current_phase`, or `primary_external_case_ref` returns the originally created incident when those pairs normalize equivalently. A request that includes any undeclared top-level member fails with `400`, `error.code='invalid_incident_create'`, and `error.details.reason_code='unknown_field'` even if `client_txn_id` matches a prior successful create.
  - Verifies: REQ-01-021, REQ-01-152..REQ-01-180
- **AC-211**: `GET /api/v1/incidents/{incident_id}` returns the incident resource defined by Core 01 §3.3.5.3 and §3.3.5.3.1, including `incident_version`, `updated_at`, and `updated_by_user_id`, to any current incident member; a caller lacking visibility receives `404`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023
- **AC-212**: `PATCH /api/v1/incidents/{incident_id}` with a correct `base_incident_version` by a `reviewer` or `admin` can change only `tlp`, `current_phase`, and `primary_external_case_ref`; `null` clears a nullable field; success returns `200 OK` with the updated incident resource; an effective change increments `incident_version` exactly once and sets `updated_at` plus `updated_by_user_id`; a no-op patch returns `200 OK` without incrementing `incident_version`, changing `updated_at`, or changing `updated_by_user_id`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-02-014..REQ-02-023
- **AC-213**: `PATCH /api/v1/incidents/{incident_id}` with a stale `base_incident_version` fails with `409` and `error.code='incident_version_conflict'`; a caller who can see the incident but lacks sufficient role gets `403`; a caller lacking visibility gets `404`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023
- **AC-214**: `PATCH /api/v1/incidents/{incident_id}` that attempts to mutate `incident_id`, `incident_key`, `title`, `description`, `status`, `severity`, `created_by_user_id`, `created_at`, `updated_at`, `updated_by_user_id`, `incident_version`, `closed_at`, membership objects, saved-view objects, or workbook-preference objects, or that sends unknown top-level members, fails with `400` and `error.code='invalid_incident_patch'`, and leaves no partial incident state.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023
- **AC-175**: A caller with `is_deployment_admin=true` can `POST /api/v1/users` with `client_txn_id`, `auth_kind='local'`, `email`, `display_name`, and `initial_password`; `email` is validated under `email_address_v1`; the resulting safe user resource shows `mfa_required=true`, `is_deployment_admin=false`, and `is_active=true` when those optional create fields are omitted, shows `email` as the only base-profile local login identifier, exposes no top-level `username`, and returns `auth_bindings[]` with exactly one local binding summary containing exactly `provider_type='local'`, `provider_key='local'`, `username` equal to that same authoritative `email`, and `created_at` equal to the safe user resource `created_at`, while omitting `auth_binding_id`, `provider_subject`, and `last_auth_at`; `display_name` is validated under `display_name_line_v1`, so it must remain non-empty after Unicode NFC normalization and trimming, reject C0/C1 control characters, and fail closed above 256 Unicode scalar values; `initial_password` is validated under `local_password_provision_v1`, so it must be a non-null JSON string, preserve exact post-decoding code points without Unicode NFC normalization, trimming, case-folding, or line-ending normalization, reject input composed entirely of Unicode whitespace code points, reject any C0/C1 control code point, and fail closed below 12 or above 1024 Unicode scalar values; explicit `mfa_required` or `is_deployment_admin` values override only those fields; explicit `null` for either optional boolean, any client-supplied `is_active`, any malformed `email`, `display_name`, or `initial_password`, and any unknown top-level member fail with `400 error.code='invalid_mutation_payload'`; when the failure is attributable to `initial_password`, `error.details.field='initial_password'`; the created user can be read through both `GET /api/v1/users/{user_id}` and `GET /api/v1/users`; `GET /api/v1/users` returns safe user resources in `data.users[]` ordered by `user_id asc`, uses `meta.paging.limit=100` when `limit` is omitted, uses `meta.paging.has_more=false` and `meta.paging.next_cursor=null` on terminal pages, fails closed with `400 error.code='invalid_pagination_request'` for invalid `limit` values or aliases such as `page`, `offset`, `page_size`, and `block_size`, rejects cursor replay against a different bound contract, and `GET /api/v1/users/{user_id}` rejects pagination members with `error.details.reason_code='pagination_not_supported'`; those route responses never include password hashes, TOTP secrets, WebAuthn credential material, opaque session tokens, or provider assertions; the server encodes the validated `initial_password` as UTF-8 without BOM for Argon2id derivation, persists only `password_hash`, and neither the safe user resource nor the deployment-local administrative audit or idempotency state created by this operation contains cleartext `initial_password`.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-01-240..REQ-01-242, REQ-01-487..REQ-01-489, REQ-01-497, REQ-01-521, REQ-02-237, REQ-02-244..REQ-02-245, REQ-04-038
- **AC-176**: Replaying `POST /api/v1/users` with the same `(actor_user_id, client_txn_id)` and the same normalized request returns `200 OK` and the originally created `user_id`; for this normalization, canonically equivalent `email` values compare after `email_address_v1` normalization, canonically equivalent `display_name` values compare after `display_name_line_v1` normalization, omitted `mfa_required` compares equal to explicit `true`, omitted `is_deployment_admin` compares equal to explicit `false`, and `initial_password` compares only after `local_password_provision_v1` validation using exact post-JSON-decoding code-point equality with no trimming, Unicode NFC normalization, case-folding, or line-ending normalization; reusing that key with a different normalized request, including an `initial_password` value that differs only by case, surrounding whitespace, or Unicode composition, fails with `409 error.code='client_txn_conflict'` and `error.details.client_txn_id`; an unauthenticated caller or an authenticated caller without `is_deployment_admin=true` cannot list, create, or patch users through the public route family; and the first deployment admin is not creatable through an unauthenticated public request.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-01-487..REQ-01-489, REQ-01-497, REQ-01-521, REQ-02-244
- **AC-177**: `PATCH /api/v1/users/{user_id}` with a correct `base_user_version` can change only the mutable fields allowed by Core 01 §3.3.5.1 through the user patch route; password reset, TOTP reset, and session revoke-all remain distinct action routes; a stale `base_user_version` fails with `409` and `error.code = user_version_conflict`; deactivating a user revokes all active sessions immediately but leaves incident memberships intact; and demoting or deactivating the last active deployment admin fails with `409` and `error.code = last_deployment_admin`.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-02-245
- **AC-340**: `POST /api/v1/users/{user_id}/password/reset` is callable only by a caller with `is_deployment_admin=true`; incident membership or incident role `admin` alone is insufficient; the request accepts `base_user_version`, `client_txn_id`, `new_password`, and optional `reason`; a successful reset updates the user's password, increments `user_version`, invalidates pending bootstrap or pending TOTP-enrollment state, revokes all active sessions immediately, preserves any active TOTP factor, returns only the resulting safe user resource, and never exposes cleartext password material in the response.
  - Verifies: REQ-01-032, REQ-01-527, REQ-02-245, REQ-04-016, REQ-04-083, REQ-04-085, REQ-04-086
- **AC-341**: `POST /api/v1/users/{user_id}/mfa/totp/reset` is callable only by a caller with `is_deployment_admin=true`; incident membership or incident role `admin` alone is insufficient; the request accepts `base_user_version`, `client_txn_id`, and optional `reason`; a successful reset clears active and pending TOTP state, increments `user_version`, revokes all active sessions immediately, preserves `mfa_required`, returns only the resulting safe user resource, and causes the next valid password login for that user to return `401 error.code='mfa_setup_required'`, `error.details.required_setup_kinds=["totp"]`, one `bootstrap_token`, and `bootstrap_expires_at` when `mfa_required=true`.
  - Verifies: REQ-01-025, REQ-01-032, REQ-01-234, REQ-01-238, REQ-01-528, REQ-02-245, REQ-04-001, REQ-04-016, REQ-04-083, REQ-04-084, REQ-04-085, REQ-04-086
- **AC-342**: `POST /api/v1/users/{user_id}/sessions/revoke-all` is callable only by a caller with `is_deployment_admin=true`; incident membership or incident role `admin` alone is insufficient; a successful call revokes all active sessions for the addressed user immediately, returns `user_id`, `sessions_revoked=true`, and `revoked_at`, and does not mutate password hash, active TOTP credential state, pending TOTP-enrollment state, or `mfa_required`.
  - Verifies: REQ-01-032, REQ-01-529, REQ-02-245, REQ-04-016, REQ-04-083, REQ-04-085, REQ-04-086
- **AC-178**: Any current incident member can `GET /api/v1/incidents/{incident_id}/memberships`; the list route returns `data.memberships[]` ordered by `joined_at asc, user_id asc`, uses `meta.paging.limit=100` when `limit` is omitted, uses `meta.paging.has_more=false` and `meta.paging.next_cursor=null` on terminal pages, fails closed with `400 error.code='invalid_pagination_request'` for invalid `limit` values or aliases such as `page`, `offset`, `page_size`, and `block_size`, and rejects cursor replay against a different bound contract; an incident admin can `POST /api/v1/incidents/{incident_id}/memberships` with exactly one of `user_id` or `email` and one valid role; creating a new membership returns `201 Created`; re-adding the same user with the same role returns `200 OK` and does not create a second membership row; re-adding the same user with a different role fails with `409` and `error.code = membership_exists_use_patch`; supplying a nonexistent user fails with `404` and `error.code = user_not_found`; supplying an inactive user fails with `409` and `error.code = user_inactive`; the route never auto-creates or invites a user; and when `email` is supplied, resolution uses the same `email_address_v1` normalization and comparison substrate as local login and user create or update, while stored membership state binds to resolved `user_id` rather than raw email text.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-01-240..REQ-01-242, REQ-01-497, REQ-02-244, REQ-04-021..REQ-04-030
- **AC-179**: `PATCH /api/v1/incidents/{incident_id}/memberships/{user_id}` with a correct `base_membership_version` changes only `role`; a stale version fails with `409` and `error.code = membership_version_conflict`; requesting the current role returns `200 OK` without incrementing `membership_version`; and a successful role change takes effect on the next request because incident authorization is re-derived from current membership and role at request time.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030
- **AC-180**: `DELETE /api/v1/incidents/{incident_id}/memberships/{user_id}` removes only that incident membership and returns `204 No Content`; deleting or demoting the last incident admin fails with `409` and `error.code = last_incident_admin`; self-removal or self-demotion succeeds only when another current incident admin remains; and removing membership from incident A leaves the same still-valid session usable for incident B when the user remains authorized there.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030

- **AC-181**: `DELETE /api/v1/records/{record_id}` accepts JSON `base_row_version`, `client_txn_id`, and optional `reason`; `reason` is normalized under `reason_note_v1`, and omission, explicit `null`, and normalized-empty reason compare equal for idempotency and persist as `null`; an `editor`, `reviewer`, or `admin` on the incident receives `200 OK` with `record_id`, `incident_id`, incremented `row_version`, `deleted=true`, non-null `deleted_at`, non-null `deleted_by_user_id`, and `change_set_id`; the record no longer appears in ordinary view queries; `GET /api/v1/records/{record_id}/history` still returns prior history plus a `soft_delete` entry; the collaboration stream emits `record_changed` with `changed_field_keys[]` present and empty and with `affected_views[]` present, non-empty, free of duplicate `view_schema_id` values, serialized in canonical ascending `view_schema_id` order, and using `change_kind='remove'` on each affected-view entry; a stale base version fails with `409 error.code='row_version_conflict'`; patching the deleted record fails with `409 error.code='record_deleted_use_restore'`; and replaying the same normalized delete request by the same actor with the same `(record_id, client_txn_id)` returns the originally committed success without creating a second mutation.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-210
- **AC-182**: `POST /api/v1/records/{record_id}/restore` accepts JSON `base_row_version`, `client_txn_id`, and optional `reason`; `reason` is normalized under `reason_note_v1`, and omission, explicit `null`, and normalized-empty reason compare equal for idempotency and persist as `null`; a `reviewer` or `admin` on the incident can restore a currently soft-deleted record and receives `200 OK` with `deleted=false`, `deleted_at=null`, `deleted_by_user_id=null`, incremented `row_version`, and a new `change_set_id`; the record becomes eligible again for ordinary view queries; history remains append-only and adds a `restore` entry rather than rewriting the prior delete; the collaboration stream emits `record_changed` with `changed_field_keys[]` present and empty and with `affected_views[]` present, non-empty, free of duplicate `view_schema_id` values, serialized in canonical ascending `view_schema_id` order, and using `change_kind='invalidate'` on each affected-view entry rather than a new insert-like change kind; a restore against a non-deleted record fails with `409 error.code='record_not_deleted'`; and if an overlapping in-flight destructive operation already holds the required protected-set lock for the target record, the route fails immediately with `409 error.code='record_locked'` and `error.retryable=true`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-104, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-210, REQ-03-101
- **AC-183**: For a soft-deleted record, `GET /api/v1/records/{record_id}/history` exposes the current tombstone `row_version`; a restore request that uses that current tombstone `row_version` succeeds when otherwise authorized, while a restore request that uses an older `row_version` fails with `409 error.code='row_version_conflict'`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-02-210
- **AC-299**: `PATCH /api/v1/records/{record_id}` closes route-scoped idempotency and mutation-list semantics: a first successful patch by an authorized actor with valid JSON `view_schema_id`, `base_row_version`, `client_txn_id`, and non-empty `changes[]` returns `200 OK` with `record_id`, incremented `row_version`, `change_set_id`, and authoritative committed row fields; replay by the same actor with the same `(record_id, client_txn_id)` and the same normalized patch request, including any outer-`changes[]` permutation that canonicalizes to the same field-key-sorted change set, returns `200 OK` with the original committed patch result rather than the row's later current state; exactly `32` `changes[]` entries succeed when otherwise valid; `33` `changes[]` entries fail with `400 error.code='invalid_mutation_payload'`, `error.details.reason_code='change_count_exceeded'`, `error.details.field='changes'`, `error.details.requested_count=33`, and `error.details.max_count=32`; duplicate `field_key` entries in `changes[]` and `changes[]: []` each fail with `400 error.code='invalid_mutation_payload'`; a `collection_actions_v1.actions[]` array with exactly `64` entries succeeds when otherwise valid; an empty `collection_actions_v1.actions[]` array fails with `400 error.code='invalid_mutation_payload'`, `error.details.reason_code='empty_collection_actions'`, and `error.details.field` plus `error.details.field_key` identifying the containing collection member; a `collection_actions_v1.actions[]` array with `65` entries fails with `400 error.code='invalid_mutation_payload'`, `error.details.reason_code='collection_action_count_exceeded'`, `error.details.field` plus `error.details.field_key` identifying the containing collection member, `error.details.requested_count=65`, and `error.details.max_count=64`; same-key reuse with a different normalized patch request fails with `409 error.code='client_txn_conflict'` even when the supplied `base_row_version` is stale; with no prior committed idempotency hit, a stale `base_row_version` on ordinary record `PATCH` follows the Core 03 field-level rules: non-overlapping writable fields auto-rebase, overlapping writable fields fail with `409 error.code='same_field_conflict'`, and missing or unusable revision history fails closed with `409 error.code='row_version_conflict'`; replay of a prior success emits no second mutation, no second `change_set`, no second revision entry, and no second replayable `record_changed` event; and any request whose outcome would depend on client-supplied outer `changes[]` order fails with `400 error.code='invalid_mutation_payload'`. For this criterion, normalized patch comparison occurs only after request-shape validation, authorization, and target-record visibility succeed, and compares exact `view_schema_id`, exact `base_row_version`, and canonical `changes[]` sorted by `field_key asc`, with direct-write fields compared by authoritative normalized `value`, write-action fields compared by semantically validated normalized `action_payload`, inherently ordered payload members such as `collection_actions_v1.actions[]` left in declared order, and count-limit failures evaluated before replay or commit so an oversize mutation never becomes replayable committed state.
  - Verifies: REQ-01-058, REQ-01-063, REQ-01-069, REQ-01-070
- **AC-215**: `GET /api/v1/records/{record_id}/history` returns `incident_id`, `record_id`, current `row_version`, `deleted`, and logical history `items[]`; each logical item includes `change_set_id`, `reversible`, `available_rollback_actions[]`, `history_entry_ref` only when the item maps to exactly one reversible mutation target, and `revision_no` only when whole-row restore is legal; `available_rollback_actions[]` contains only `history_entry`, `change_set`, and `row_restore`; a caller lacking visibility receives `404`; when more than one page is possible, the route accepts only `limit` and `cursor_token`, omitting `limit` yields `meta.paging.limit=100`, terminal pages use `meta.paging.has_more=false` and `meta.paging.next_cursor=null`, invalid `limit` values or aliases such as `page`, `offset`, `page_size`, and `block_size` fail closed with `400 error.code='invalid_pagination_request'`, and cursor replay against a different `record_id` or other bound contract fails closed.
  - Verifies: REQ-01-056, REQ-01-057..REQ-01-111, REQ-01-240..REQ-01-242, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-218, REQ-03-139..REQ-03-142
- **AC-216**: `POST /api/v1/records/{record_id}/rollback` with valid JSON `base_row_version`, `client_txn_id`, `target.kind='history_entry'`, and a visible `history_entry_ref` by a `reviewer` or `admin` can roll back one mistaken host link, tag assignment, mention resolution or dismissal, or evidence attach or detach from the selected row's history without reverting later unrelated edits on the same row; when optional `reason` is supplied it is normalized under `reason_note_v1`, and omission, explicit `null`, and normalized-empty reason compare equal for idempotency and persistence; success returns `200 OK` with incremented `row_version`, echoed `target`, `target_change_set_id`, new `rollback_change_set_id`, and canonical `affected_record_ids[]`; and replaying the same normalized request by the same actor with the same `(record_id, client_txn_id)` returns the originally committed success without creating a second rollback `change_set`.
  - Verifies: REQ-01-057..REQ-01-111, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-216,
    REQ-03-141..REQ-03-144
- **AC-217**: `POST /api/v1/records/{record_id}/rollback` with `target.kind='change_set'` and a non-merge reversible `change_set_id` reverses the entire reversible `change_set` in reverse deterministic entry order; the same route with `target.kind='row_restore'` and `restore_to_revision_no` restores only row-backed fields to the selected revision and does not implicitly recreate or delete links, tags, mentions, indicator observations, or evidence associations; each success appends one new attributed rollback `change_set`, preserves prior history, increments `row_version` on every affected first-class record, and reaches subscribers only through ordinary `record_changed` events.
  - Verifies: REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-218, REQ-03-141..REQ-03-144
- **AC-412**: `POST /api/v1/records/{record_id}/rollback` with `target.kind='change_set'` and a merge `change_set_id` reverses a merge-aware `change_set` in reverse deterministic entry order through the existing rollback route without introducing a separate unmerge route. A successful merge rollback restores survivor record fields, loser merged or supersession state, carried aliases and preserved identifiers, repointed or deduped links, tags, mentions, assessments, and evidence links, projection state, and canonical `affected_record_ids[]`; it appends exactly one attributed rollback `change_set`, preserves the original merge history, increments `row_version` on every affected first-class record, emits only ordinary `record_changed` invalidations, and replays idempotently for the same normalized request without creating a second rollback `change_set`. Merge rollback fails closed before partial graph reversal when the merge substrate lacks complete reversible before/after data, any target in the protected set is locked, the supplied `base_row_version` is stale, the merge was already rolled back, or any affected survivor, loser, or dependent target has a later dependent change; incomplete historical merge data returns `409 error.code='rollback_precondition_failed'` with `error.details.reason_code='target_not_reversible'`.
  - Verifies: REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-220, REQ-03-141..REQ-03-144
- **AC-413**: A default runtime exposes no `/api/v1/test/*` or `/ws/v1/test/*` route. When test routes are enabled without the harness-owned marker or without a valid test-route token, startup fails closed before serving those routes. In a harness-owned runtime, missing or wrong `X-Cartulary-Test-Route-Token` values return `403 error.code='test_route_forbidden'` for test clock, runtime reset, runtime identity, auth touch, timeline snapshot, and test WebSocket routes; session cookies, bearer sessions, bootstrap tokens, incident roles, and `deployment_admin` do not bypass that token requirement. If a harness API or browser origin is configured, wrong host, missing origin, malformed origin, or unapproved origin requests fail before any runtime-control mutation, and the response does not enable permissive CORS.
  - Verifies: REQ-04-109
- **AC-218**: Malformed rollback request shape, unknown top-level request member, unknown `target.kind`, or a selector whose JSON type does not match the declared shape fails with `400 error.code='invalid_rollback_request'`; a rollback target not visible in the current record history fails with `404 error.code='rollback_target_not_found'`; a rollback against a currently soft-deleted record fails with `409 error.code='record_deleted_use_restore'`; later dependent changes, a non-individually-reversible history item, or a selector superseded by later reversals fail with `409 error.code='rollback_precondition_failed'` and `error.details.reason_code` equal to `target_not_reversible`, `entry_requires_change_set`, `dependent_later_changes`, or `stale_target`; stale `base_row_version` fails with `409 error.code='row_version_conflict'`; and if an overlapping in-flight destructive operation already holds one or more required protected-set locks for the rollback, the route fails immediately with `409 error.code='record_locked'` and `error.retryable=true`.
  - Verifies: REQ-01-057..REQ-01-111, REQ-01-228..REQ-01-239, REQ-02-205..REQ-02-207, REQ-03-101,
    REQ-03-141..REQ-03-144
- **AC-383**: For an extant record with paginated history, repeated `GET /api/v1/records/{record_id}/history` before incident closure, after incident closure, after a successful delete or restore cycle on that same record, after a successful rollback on that same record, after ordinary background-job execution, and after service restart returns the same pre-existing logical history items in the same newest-first committed order, plus the newly appended delete, restore, and rollback items; previously issued `history_entry_ref` values for older single-entry-addressable items remain unchanged; and pagination continues to traverse the full retained set rather than a retention-truncated subset.
  - Verifies: REQ-01-054, REQ-01-561..REQ-01-562, REQ-02-238..REQ-02-240
- **AC-384**: If a later dependent change or stale target state makes an earlier single-entry-addressable history item not currently reversible, that earlier item still appears in `GET /api/v1/records/{record_id}/history` with the same `history_entry_ref`, `reversible=false`, and `available_rollback_actions[]=[]`; invoking rollback with that selector fails through `409 error.code='rollback_precondition_failed'` using `error.details.reason_code='target_not_reversible'`, `dependent_later_changes`, or `stale_target` as applicable, rather than through omission of the item or disappearance of the selector.
  - Verifies: REQ-01-054, REQ-01-096, REQ-01-563, REQ-02-216, REQ-02-241
- **AC-385**: The current-profile public route inventory and deployment-configuration contract expose no history-purge route and no history-retention-horizon setting for extant incidents. A conformant base-profile deployment provides no operator path that narrows retained record history by incident age, incident closure, or per-record purge while the incident record remains extant.
  - Verifies: REQ-02-239
- **AC-353**: If request A is already holding the destructive-operation locks required by REQ-01-104 for a restore, rollback, or merge protected set, an overlapping request B against that same protected set fails immediately with `409 error.code='record_locked'` and `error.retryable=true` even when B also supplies a stale `base_row_version` or would otherwise fail a route-specific precondition. After A commits or rolls back and releases its locks, replaying B without refreshed inputs fails through the ordinary downstream path for the route, such as `row_version_conflict`, `record_not_deleted`, `rollback_precondition_failed`, or `merge_precondition_failed`, rather than `record_locked`.
  - Verifies: REQ-01-104, REQ-03-101
- **AC-184**: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` accepts `filters[]` entries using the Core 01 §3.3.4.1 wire shape for `eq`, `range`, `contains_any`, `contains_all`, `prefix`, and `full_text`; semantically identical client orderings normalize to the same `meta.query.filters[]`; reordered or duplicated set-like `values[]` normalize to the same canonical value set in canonical ascending order; case-only variants of `prefix.value` normalize to the same canonical filter value and match only at code-point offset `0` of the comparison-normalized field value; and equivalent normalized filter contracts bind the same cursor contract. An unknown filter field, disallowed operator, duplicate `field_key`, JSON `null` inside a set-like `values[]`, a set-like operand that becomes empty after normalization or duplicate coalescing, empty `prefix.value`, empty `full_text.query`, a `full_text.query` that tokenizes to zero tokens after normalization, malformed range, malformed sort entry, duplicate sort `field_key`, disallowed sort key, malformed `group_by`, or disallowed `group_by` fails with `400` and `error.code = invalid_view_query`; the post-normalization empty-set case uses `reason_code='empty_values_after_normalization'` and the zero-token full-text case uses `reason_code='empty_full_text_after_tokenization'`.
  - Verifies: REQ-01-034..REQ-01-056, REQ-01-312..REQ-01-322, REQ-03-223..REQ-03-224
- **AC-185**: The Notes view exposes `note.full_text` as a stable synthetic filter key; a query using `field_key='note.full_text'` and `op='full_text'` tokenizes the union of `note.title` and `note.body` into maximal contiguous Unicode letter-or-number runs, treats all other code points as separators, treats null source fields as empty, applies the shared query-time text-comparison substrate with diacritics still significant, coalesces duplicate query tokens, treats query-token order as non-semantic, matches only when every unique normalized query token appears as an exact token across that union, and rejects a zero-token query after normalization with `400 error.code='invalid_view_query'` and `error.details.reason_code='empty_full_text_after_tokenization'`.
  - Verifies: REQ-01-034..REQ-01-056, REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071,
    REQ-03-223..REQ-03-224
- **AC-387**: A row or note that would match only under fuzzy similarity, trigram similarity, stemming, accent-insensitive comparison, phrase semantics, wildcard semantics, or transliteration does not match `prefix` or `full_text` unless it also satisfies the exact Core 01 operator contract. When a valid `full_text` filter is present, `rows[]` still follow the applied sort tuple reported in `meta.query.sort[]`; the route does not inject score ordering or hidden relevance tie-breaks.
  - Verifies: REQ-01-035..REQ-01-047, REQ-01-565..REQ-01-567, REQ-03-223..REQ-03-224
- **AC-186**: `POST /api/v1/records/{survivor_record_id}/merge` accepts JSON `loser_record_id`, `survivor_base_row_version`, `loser_base_row_version`, `client_txn_id`, and optional `reason`; `reason` is normalized under `reason_note_v1`, and omission, explicit `null`, and normalized-empty reason compare equal for idempotency and persistence; a `reviewer` or `admin` on the incident can merge two visible same-incident same-type active `host` records or two visible same-incident same-type active `identity` records; on success the route returns `200 OK` with `incident_id`, `record_type`, both record IDs, updated survivor and loser row versions, `change_set_id`, and required `merge_summary`; `merge_summary.exact_match_classes[]` is ordered by the merged type's exact-match precedence and each entry includes `identifier_class`, `promoted_count`, `carried_count`, and `duplicate_noop_count`, while top-level summary counts include `suggestion_aliases_copied_count`, `suggestion_alias_duplicate_noop_count`, and `provenance_only_retained_count`; the loser becomes historical with `merged_into_record_id` set to the survivor; active mentions, links, assessments, and tags are repointed or deterministically recreated in the same `change_set`; duplicate links and tags are deduplicated without losing history; a loser-side reusable exact-match value missing on the survivor is promoted or carried so later exact-match reuse using that value resolves to the survivor; when survivor and loser hold different reusable exact-match values of the same class, later exact-match reuse using either value resolves to the survivor; and the collaboration stream removes the loser from ordinary active entity views while invalidating or patching the survivor and affected dependent rows.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-181..REQ-01-195, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-054..REQ-02-055,
    REQ-02-060..REQ-02-061, REQ-02-064..REQ-02-066, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249
- **AC-187**: The same route fails closed when the two IDs are identical, the records belong to different incidents, the record types differ, the record type is not `host` or `identity`, either record is soft-deleted or already merged away, the caller lacks `reviewer` or `admin` role, either supplied base row version is stale, preserving a loser-side `exact_match_reuse` value would collide with a third active same-incident record, or an overlapping in-flight destructive operation already holds one or more required protected-set locks; stale versions fail with `409 error.code='row_version_conflict'`; an active lock fails with `409 error.code='record_locked'` and `error.retryable=true`; the third-record carry-forward collision fails with `409 error.code='merge_precondition_failed'`, `error.details.reason_code='carry_forward_identifier_collision'`, and `error.details` containing `identifier_class`, `normalized_value`, and `blocking_record_id`; other merge-precondition failures fail with `409 error.code='merge_precondition_failed'`; a caller lacking visibility receives `404`; no partial state is committed on any failure path; and replaying the same normalized request by the same actor with the same `(survivor_record_id, loser_record_id, client_txn_id)` returns the originally committed success without creating a second merge.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-104, REQ-01-181..REQ-01-195, REQ-01-237, REQ-02-060..REQ-02-061,
    REQ-02-064..REQ-02-066, REQ-03-101, REQ-03-247..REQ-03-249

- **AC-277**: `POST /api/v1/incidents/{incident_id}/views/cartulary.view.parties.v1/rows` commits only when `party.display_name` remains non-empty after `display_name_line_v1` normalization and `party.party_kind` is present; zero-field create fails with no partial row or misleading projection update; `GET /api/v1/view-schemas` and `GET /api/v1/view-schemas/cartulary.view.parties.v1` expose the Parties schema as part of the fourteen pack-independent base-profile view contracts; and the Parties system view is incident-scoped rather than deployment-local user administration.
  - Verifies: REQ-01-296, REQ-01-343, REQ-01-497..REQ-01-501, REQ-02-003, REQ-02-006, REQ-02-009, REQ-02-022, REQ-02-202, REQ-02-222..REQ-02-225, REQ-03-005, REQ-03-266
- **AC-278**: Linking or clearing `task.requester_party_id`, `evidence.collector_party_id`, or `evidence.source_party_id` from the inspector or Parties view preserves independent text and ref semantics: text-only, ref-only, both-present, and both-null states round-trip; clearing text does not clear the ref; clearing the ref does not clear text; `owner_user_id` remains the accountable assignee when present; and `comm_log.audience` text remains required even when supplemental audience or attendee party refs are also present.
  - Verifies: REQ-01-328, REQ-01-336, REQ-01-502, REQ-02-017, REQ-02-021..REQ-02-022, REQ-02-120, REQ-02-124..REQ-02-125, REQ-02-197..REQ-02-199, REQ-02-226..REQ-02-228, REQ-03-247, REQ-03-256, REQ-03-268..REQ-03-271
- **AC-279**: Typing requester, collector, source, audience, or attendee text into ordinary workbook fields does not auto-create or auto-link a `party` record. Explicit `Create party from text` or `Link existing party` flows from the inspector or Parties view can do so; exact-match create or reuse is limited to one active same-incident party selected by normalized `primary_email` or, failing that, `external_ref`; display name, organization, role title, and phone-like text do not act as auto-upsert keys; and the resulting create or link flow remains same-surface and non-blocking.
  - Verifies: REQ-01-296, REQ-01-328, REQ-01-336, REQ-01-497, REQ-01-499..REQ-01-502, REQ-02-022, REQ-02-060..REQ-02-063, REQ-02-229..REQ-02-232, REQ-03-247, REQ-03-256, REQ-03-259, REQ-03-267..REQ-03-271
- **AC-280**: A `task.requester_party_id`, `evidence.collector_party_id`, or `evidence.source_party_id` that targets a party from another incident or a deleted party fails closed and leaves no partial link state; deleting a currently referenced party fails closed and leaves authoritative state unchanged; and party rows and party references inherit the same incident-level authorization model as other incident data rather than any deployment-admin bypass.
  - Verifies: REQ-01-328, REQ-01-336, REQ-02-021..REQ-02-022, REQ-02-198, REQ-02-226, REQ-02-231..REQ-02-232, REQ-04-022..REQ-04-024


- **AC-281**: `POST /api/v1/incidents/{incident_id}/views/cartulary.view.comm_log.v1/rows` commits only when `comm_log.comm_type` is present and `comm_log.audience`, `comm_log.channel_or_meeting`, and `comm_log.summary` remain non-empty after create-time normalization; omitted `comm_log.timestamp_utc` defaults to the commit timestamp; `comm_log.decision_ids`, `comm_log.action_task_ids`, `comm_log.audience_party_ids`, and `comm_log.attendee_party_ids` default to empty collections; `comm_log.next_report_at` and `comm_log.privilege_tag` default to `null`; existing-row writes to `comm_log.comm_id` fail closed; `comm_log.audience` text remains required and source-preserving even when supplemental party refs are present; `comm_log.decision_ids` and `comm_log.action_task_ids` round-trip only as `collection_value_v1` `record_ref` items using `item_ref="record_ref:<linked_record_id>"`, `item_kind="record_ref"`, `linked_record_id`, and `ordered=false`; create or patch accepts only `collection_actions_v1` `add_record_ref` and `remove_record_ref` actions for those fields; target validation accepts only same-incident active `decision` and `task_request` records respectively; duplicate adds coalesce by `(record_id, linked_record_id, field_key)`; `comm_log.audience_party_ids` and `comm_log.attendee_party_ids` round-trip only as `collection_value_v1` `party_ref` items using `item_ref="party_ref:<party_id>"`, `item_kind="party_ref"`, `display_text`, `party_id`, and `ordered=false`; create or patch accepts only `collection_actions_v1` `add_party_ref` and `remove_party_ref` actions for those fields; target validation accepts only same-incident active `party` records; duplicate adds coalesce by `(record_id, party_id, field_key)`; same-field conflict payloads for all four collection fields use typed `collection_value_v1` rather than raw arrays; invalid or foreign `item_ref`, wrong target type, foreign-incident target, or soft-deleted target fails with `400 error.code='invalid_mutation_payload'`; removing a party ref leaves `comm_log.audience` unchanged; filter, sort, and grouping behavior over `comm_log.comm_type`, `comm_log.timestamp_day`, and `comm_log.next_report_day` is satisfied from projection-backed state; clearable optional scalar fields honor the declared omission-versus-`null` contract; raw arrays and raw `null` are rejected for collection patches; and any rejected create leaves no partial record, no projection row, and no misleading live event.
  - Verifies: REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-503, REQ-02-123..REQ-02-125, REQ-02-132..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265
- **AC-282**: `POST /api/v1/incidents/{incident_id}/views/cartulary.view.handoff.v1/rows` commits only when `handoff.incoming_owner_user_id` is present and `handoff.current_state_summary` remains non-empty after create-time normalization; omitted `handoff.timestamp_utc` defaults to the commit timestamp; omitted `handoff.outgoing_owner_user_id` defaults to the current actor; `handoff.open_task_ids`, `handoff.open_decision_ids`, and `handoff.open_risk_refs` default to empty collections; `handoff.next_checks` and `handoff.acknowledged_at` default to `null`; existing-row writes to `handoff.handoff_id` fail closed; `handoff.open_task_ids` and `handoff.open_decision_ids` round-trip only as `collection_value_v1` `record_ref` items using `item_ref="record_ref:<linked_record_id>"`, `item_kind="record_ref"`, `linked_record_id`, and `ordered=false`; create or patch accepts only `collection_actions_v1` `add_record_ref` and `remove_record_ref` actions for those fields; target validation accepts only same-incident active `task_request` and `decision` records respectively; duplicate adds coalesce by `(record_id, linked_record_id, field_key)`; `handoff.open_risk_refs` round-trip only as `collection_value_v1` `risk_ref` items using `item_ref="risk_ref:<risk_ref_id>"`, `item_kind="risk_ref"`, `display_text`, `risk_ref_id`, `risk_ref_text`, and `ordered=false`; create or patch accepts only `collection_actions_v1` `add_risk_ref` and `remove_risk_ref` actions for that field; `add_risk_ref.risk_ref_text` normalizes under `single_line_title_v1`; duplicate adds coalesce by normalized `risk_ref_text` within the same `handoff` record; remove targets the stable server-issued `item_ref` rather than raw text; creating the same normalized `risk_ref_text` on two different `handoff` records yields distinct `risk_ref_id` values, confirming that duplicate-coalescing scope is one parent `handoff` rather than an incident-wide or deployment-wide reusable risk object namespace; same-field conflict payloads for all three collection fields use typed `collection_value_v1` rather than raw arrays; invalid or foreign `item_ref`, wrong target type, foreign-incident target, or soft-deleted target fails with `400 error.code='invalid_mutation_payload'`; filter, sort, and grouping behavior over `handoff.timestamp_day`, `handoff.outgoing_owner_user_id`, `handoff.incoming_owner_user_id`, and derived `handoff.ack_state` is satisfied from projection-backed state; clearable optional scalar fields honor the declared omission-versus-`null` contract; raw arrays and raw `null` are rejected for collection patches; and any rejected create leaves no partial record, no projection row, and no misleading live event.
  - Verifies: REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-504, REQ-02-123, REQ-02-126..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265
- **AC-283**: `POST /api/v1/incidents/{incident_id}/views/cartulary.view.status_review.v1/rows` commits only when `status_review.current_state_summary` remains non-empty after create-time normalization; omitted `status_review.timestamp_utc` defaults to the commit timestamp; omitted `status_review.review_owner_user_id` defaults to the current actor; `status_review.blocked_task_ids`, `status_review.pending_evidence_ids`, and `status_review.open_decision_ids` default to empty collections; `status_review.active_risks_summary` and `status_review.next_report_at` default to `null`; existing-row writes to `status_review.status_review_id` fail closed; all three collection fields round-trip only as `collection_value_v1` `record_ref` items using `item_ref="record_ref:<linked_record_id>"`, `item_kind="record_ref"`, `linked_record_id`, and `ordered=false`; create or patch accepts only `collection_actions_v1` `add_record_ref` and `remove_record_ref` actions for those fields; target validation accepts only same-incident active `task_request` for `status_review.blocked_task_ids`, same-incident active `evidence` for `status_review.pending_evidence_ids`, and same-incident active `decision` for `status_review.open_decision_ids`; duplicate adds coalesce by `(record_id, linked_record_id, field_key)`; same-field conflict payloads for all three collection fields use typed `collection_value_v1` rather than raw arrays; invalid or foreign `item_ref`, wrong target type, foreign-incident target, or soft-deleted target fails with `400 error.code='invalid_mutation_payload'`; filter, sort, and grouping behavior over `status_review.timestamp_day`, `status_review.review_owner_user_id`, and `status_review.next_report_day` is satisfied from projection-backed state; clearable optional scalar fields honor the declared omission-versus-`null` contract; raw arrays and raw `null` are rejected for collection patches; and any rejected create leaves no partial record, no projection row, and no misleading live event.
  - Verifies: REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-505, REQ-02-123, REQ-02-128..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265
- **AC-284**: `POST /api/v1/incidents/{incident_id}/views/cartulary.view.lesson.v1/rows` commits only when `lesson.summary` remains non-empty after create-time normalization; omitted `lesson.timestamp_utc` defaults to the commit timestamp; omitted `lesson.owner_user_id` defaults to the current actor; `lesson.follow_up_task_ids` and `lesson.evidence_refs` default to empty collections; `lesson.closure_state` defaults to `open`; existing-row writes to `lesson.lesson_id` fail closed; both collection fields round-trip only as `collection_value_v1` `record_ref` items using `item_ref="record_ref:<linked_record_id>"`, `item_kind="record_ref"`, `linked_record_id`, and `ordered=false`; create or patch accepts only `collection_actions_v1` `add_record_ref` and `remove_record_ref` actions for those fields; target validation accepts only same-incident active `task_request` for `lesson.follow_up_task_ids` and same-incident active `evidence` for `lesson.evidence_refs`; duplicate adds coalesce by `(record_id, linked_record_id, field_key)`; same-field conflict payloads for both collection fields use typed `collection_value_v1` rather than raw arrays; invalid or foreign `item_ref`, wrong target type, foreign-incident target, or soft-deleted target fails with `400 error.code='invalid_mutation_payload'`; filter, sort, and grouping behavior over `lesson.closure_state`, `lesson.owner_user_id`, and `lesson.timestamp_day` is satisfied from projection-backed state; clearable optional scalar fields honor the declared omission-versus-`null` contract; raw arrays and raw `null` are rejected for collection patches; and any rejected create leaves no partial record, no projection row, and no misleading live event.
  - Verifies: REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-506, REQ-02-123, REQ-02-130..REQ-02-133, REQ-02-222, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265
- **AC-285**: If the implementation exposes `cartulary.view.findings.v1`, `POST /api/v1/incidents/{incident_id}/views/cartulary.view.findings.v1/rows` commits only when `finding.statement` remains non-empty after create-time normalization; `finding.kind` defaults to `finding` when omitted on create and uses the exact closed vocabulary `finding | hypothesis`; `finding.state` defaults to `open`; `finding.owner_user_id` defaults to the current actor; `finding.confidence_score` and `finding.closed_at` default to `null`; existing-row writes reject explicit `null` or normalized-empty input for required finding fields; `finding.closed_at` is server-managed and is set only when `finding.state` transitions to `closed`; supporting and contradictory references mutate through collection actions rather than JSON blobs; filter, sort, and grouping behavior over `finding.kind`, `finding.state`, `finding.owner_user_id`, and derived `finding.confidence_band` is satisfied from projection-backed state; and any rejected create leaves no partial record, no projection row, and no misleading live event.
  - Verifies: REQ-01-308..REQ-01-310, REQ-01-358, REQ-01-507, REQ-02-134..REQ-02-136, REQ-03-259, REQ-03-265
- **AC-286**: If the implementation exposes `cartulary.view.investigative_queries.v1`, `POST /api/v1/incidents/{incident_id}/views/cartulary.view.investigative_queries.v1/rows` commits only when `investigative_query.platform`, `investigative_query.purpose`, and `investigative_query.query_text` remain non-empty after create-time normalization; `investigative_query.created_by_user_id` defaults to the current actor; `investigative_query.query_id` is server-generated and immutable after first commit; `investigative_query.created_by_user_id` is read-only after first commit; existing-row writes reject explicit `null` or normalized-empty input for required investigative-query fields; filter, sort, and grouping behavior over `investigative_query.platform`, `investigative_query.created_by_user_id`, and `investigative_query.created_day` is satisfied from projection-backed state; and any rejected create leaves no partial record, no projection row, and no misleading live event.
  - Verifies: REQ-01-308..REQ-01-310, REQ-01-358, REQ-01-508, REQ-02-135, REQ-02-137, REQ-03-259, REQ-03-265
- **AC-287**: If the implementation exposes `cartulary.view.forensic_keywords.v1`, `POST /api/v1/incidents/{incident_id}/views/cartulary.view.forensic_keywords.v1/rows` commits only when `forensic_keyword.pattern` and `forensic_keyword.reason` remain non-empty after create-time normalization; `forensic_keyword.match_mode` defaults to `literal`; `forensic_keyword.case_sensitive` defaults to `false`; `forensic_keyword.keyword_id` is server-generated and immutable after first commit; existing-row writes reject explicit `null` or normalized-empty input for required forensic-keyword fields; filter, sort, and grouping behavior over `forensic_keyword.match_mode`, `forensic_keyword.case_sensitive`, and `forensic_keyword.created_day` is satisfied from projection-backed state; and any rejected create leaves no partial record, no projection row, and no misleading live event.
  - Verifies: REQ-01-308..REQ-01-310, REQ-01-358, REQ-01-509, REQ-02-135, REQ-02-138, REQ-02-222, REQ-03-259, REQ-03-265

- **AC-300**: Any writable field bound to `direct_scalar_contract_id=timestamp_instant_v1` accepts RFC 3339 timestamps only when the JSON value is a string with an explicit timezone designator, and offset-equivalent forms such as `2026-03-31T10:00:00-04:00` and `2026-03-31T14:00:00Z` compare equal after normalization for normalized equality, create-time idempotency, and structurally valid no-op detection.
  - Verifies: REQ-01-310, REQ-01-312, REQ-01-328, REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-487..REQ-01-488, REQ-01-503..REQ-01-506
- **AC-301**: `PATCH /api/v1/records/{record_id}` with explicit JSON `null` clears `timeline.occurred_at`, `evidence.requested_at`, `evidence.received_at`, `task.due_at`, `task.completed_at`, `comm_log.next_report_at`, `handoff.acknowledged_at`, and `status_review.next_report_at` when the resulting row otherwise remains legal; omission leaves the persisted value unchanged.
  - Verifies: REQ-01-310, REQ-01-312, REQ-01-328, REQ-01-336, REQ-01-488, REQ-01-503..REQ-01-505
- **AC-302**: Explicit JSON `null` is rejected for non-clearable fields bound to `timestamp_instant_v1`, including `assessment.assessed_at`, `decision.decided_at`, `comm_log.timestamp_utc`, `handoff.timestamp_utc`, `status_review.timestamp_utc`, and `lesson.timestamp_utc`; where the owning view contract defines create-time defaults, omitting the field still yields the declared commit-time default.
  - Verifies: REQ-01-310, REQ-01-332, REQ-01-339, REQ-01-488, REQ-01-503..REQ-01-506
- **AC-303**: For any field bound to `direct_scalar_contract_id=timestamp_instant_v1`, timezone-less strings, date-only strings, empty strings, numeric JSON values, booleans, arrays, and objects fail closed as invalid mutation payload; clearing a clearable timestamp therefore succeeds only when the client sends explicit JSON `null`.
  - Verifies: REQ-01-310, REQ-01-312, REQ-01-328, REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-487..REQ-01-488, REQ-01-503..REQ-01-506
- **AC-354**: Any writable field bound to `direct_scalar_contract_id=timestamp_instant_v1` preserves an invalid timestamp draft as client-local unsaved editor state: editing a clearable timestamp field with a timezone-less string causes no commit, no authoritative row refresh, and no misleading persisted-value update; the active editor retains the typed text until the analyst corrects it, discards it, or performs an explicit JSON `null` clear; explicit JSON `null` succeeds only on fields whose binding declares `clearable=true`; and attempting that clear on a non-clearable timestamp field fails closed while leaving the local state unsaved.
  - Verifies: REQ-03-281
- **AC-304**: `task.completed_at` remains governed by task lifecycle semantics after the timestamp contract binding is added: a successful transition away from `status='done'` clears `completed_at`; a write that would leave `status='done'` with `completed_at=null` fails closed; and a successful write that sets `status='done'` without an explicit `completed_at` value fills `completed_at` from the commit timestamp.
  - Verifies: REQ-01-336, REQ-01-338
- **AC-315**: `PATCH /api/v1/records/{record_id}` clears `task.requester_party_id`, `evidence.collector_party_id`, `evidence.source_party_id`, and `task.decision_record_id` only when the client sends the corresponding direct-write `field_key` with `value=null`; omission leaves the persisted reference unchanged; and attempting the same clear through `action_payload` fails closed for these direct-write fields.
  - Verifies: REQ-01-059..REQ-01-060, REQ-01-328, REQ-01-336, REQ-01-516..REQ-01-517, REQ-01-519..REQ-01-520, REQ-03-272..REQ-03-274
- **AC-316**: For optional fields bound to `direct_reference_contract_id`, create-time omission and explicit JSON `null` compare equal for normalized replay and result in authoritative `null` when no non-null reference is supplied; this applies at minimum to `task.requester_party_id`, `evidence.collector_party_id`, `evidence.source_party_id`, and `task.decision_record_id`.
  - Verifies: REQ-01-061, REQ-01-328, REQ-01-336, REQ-01-517, REQ-01-519..REQ-01-520
- **AC-317**: For any field bound to `direct_reference_contract_id`, empty strings, whitespace-padded identifier strings, numeric JSON values, booleans, arrays, and objects fail closed as `invalid_mutation_payload`; accepted non-null values are exact stable identifiers only and are not trimmed, label-resolved, email-resolved, fuzzy-matched, or auto-created by the direct-reference contract layer.
  - Verifies: REQ-01-328, REQ-01-336, REQ-01-517..REQ-01-520
- **AC-318**: Clearing or setting `task.requester_party_id`, `evidence.collector_party_id`, or `evidence.source_party_id` preserves independent text-plus-ref semantics: clearing the ref leaves preserved requester, collector, or source text unchanged; clearing text leaves the ref unchanged; and `Clear both` submits one patch containing both field changes and leaves both authoritative values `null` with no partial intermediate state committed.
  - Verifies: REQ-01-328, REQ-01-336, REQ-01-502, REQ-02-017, REQ-02-021..REQ-02-022, REQ-03-272, REQ-03-274
- **AC-319**: `task.decision_record_id` accepts only an exact same-incident active `decision` record identifier; a wrong-type record, another-incident record, deleted or otherwise ineligible target, or explicit clear routed through a non-direct-write shape fails closed with no partial authoritative state change; and when the task surface exposes decision-link clearing it uses the ordinary patch route with `value=null`.
  - Verifies: REQ-01-336, REQ-01-517..REQ-01-520, REQ-02-233, REQ-03-273


- **AC-370**: `GET /api/v1/extensions` succeeds for an authenticated current session, including one with `is_deployment_admin=false` and zero current incident memberships. The route returns the common success envelope with `data.extensions[]` ordered by `profile_id asc`; the exact current-profile set is `enterprise_authentication`, `import`, `incident_portability`, `reference_pack`, and `snapshot_reporting`; each item contains exactly `profile_id`, `claimed`, and `route_families[]`; each `route_families[]` list matches the canonical reserved family roots in Core 01 and is ordered by route-family string asc; pagination members fail with `400 error.code='invalid_pagination_request'` and `error.details.reason_code='pagination_not_supported'`; and the route discloses no provider secrets, provider claim maps, or live extension-family payload.
  - Verifies: REQ-00-022, REQ-01-032..REQ-01-033, REQ-01-542..REQ-01-546, REQ-04-105
- **AC-371**: When `GET /api/v1/extensions` reports one or more items with `claimed=false`, probing each such profile's declared `route_families[]` root and one descendant route under that same family returns `404 error.code='extension_profile_not_claimed'`, `error.retryable=false`, and `error.details` containing the matching `profile_id` and canonical matched `route_family`. That result is returned before family-specific authorization or policy evaluation for the unclaimed family; a path outside all reserved base and extension families does not use `extension_profile_not_claimed`; and a valid path inside a claimed extension family does not use `extension_profile_not_claimed` solely because the family currently has zero resources or because ordinary family authorization later denies access.
  - Verifies: REQ-00-022, REQ-01-033, REQ-01-234, REQ-01-544, REQ-01-546..REQ-01-548, REQ-04-105

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
- **AC-168**: Historical actors in `actors.ndjson` that do not map to an existing local user become either inert imported actors or historical actor descriptors bound to an existing local user; neither is login-capable, neither is automatically added to incident membership, and imported history retains the source bundle actor identifier used by that history.
  - Verifies: REQ-01-425..REQ-01-426, REQ-01-443..REQ-01-446, REQ-04-044..REQ-04-047, REQ-04-065
- **AC-169**: Whole-incident export and import run as background jobs, show progress and cancellation without blocking grid editing, stage bundle contents only under the configured temporary-work root, and in flyaway or disconnected deployments keep emitted bundles and staged extracts on encrypted storage.
  - Verifies: REQ-01-425..REQ-01-430, REQ-01-447..REQ-01-450, REQ-01-452..REQ-01-456, REQ-04-044..REQ-04-047,
    REQ-04-054..REQ-04-055, REQ-04-058..REQ-04-059, REQ-04-065
- **AC-386**: Exporting and importing an incident that contains both currently reversible and currently non-reversible single-entry-addressable history items preserves the authoritative history substrate needed to materialize `GET /api/v1/records/{record_id}/history`, including change-set count, mutation-entry count, and revision count; after projection rebuild, the imported incident's `/history` surface exposes the same logical history items and rollback scopes; the importing deployment MAY emit different opaque `history_entry_ref` values than the source deployment; repeated imported-history reads keep those imported selectors stable within the importing deployment; and rollback using an imported selector succeeds or fails only under the ordinary imported-history preconditions.
  - Verifies: REQ-01-564, REQ-02-238, REQ-02-241

- **AC-332**: Exporting and importing an incident that contains a superseded Timeline row with a committed replacement relation preserves the active Timeline `supersedes` link, and after projection rebuild a Timeline row query against the imported incident returns the same hidden `timeline.replacement_record_id` for that superseded row.
  - Verifies: REQ-01-311..REQ-01-312, REQ-01-448..REQ-01-449, REQ-01-486, REQ-02-169, REQ-02-175, REQ-02-181


- **AC-273**: `POST /api/v1/incident-bundles/export` accepts a JSON object with required `incident_id` and required `client_txn_id`; omitting `reference_pack_mode` defaults it to `refs_only`; explicit `null` for `reference_pack_mode`, `optional_sections[]`, or `required_capabilities[]` fails with `400 error.code='invalid_incident_bundle_request'` and `reason_code='field_not_nullable'`; explicit `reference_pack_mode='refs_only'` compares equal to omission for idempotency and replay; `reference_pack_mode` accepts only `refs_only` or `embedded`, otherwise failing with `reason_code='invalid_reference_pack_mode'`; omitting `optional_sections[]` or `required_capabilities[]` defaults each to `[]`, and explicit `[]` compares equal to omission; the current-profile tokens for both arrays are exactly `snapshots` and `reference_packs`; caller order is non-semantic, duplicates coalesce by exact token equality, canonical normalization sorts exact tokens ascending, and unknown or non-string members fail with `reason_code='invalid_optional_sections'` or `reason_code='invalid_required_capabilities'` as applicable; supplying `history_mode` or `blob_mode` fails with `400 error.code='invalid_incident_bundle_request'`; export returns `202` with the common job resource; and on success the terminal job summary uses `result_summary.code='incident_bundle_exported'` and exactly one `resource_refs[]` item `{ kind='incident_bundle', id=<bundle_id>, route='/api/v1/incident-bundles/{bundle_id}' }`.
  - Verifies: REQ-01-033, REQ-01-466..REQ-01-470, REQ-01-483..REQ-01-484
- **AC-274**: A durable incident-bundle descriptor exists only after successful export; `GET /api/v1/incident-bundles/{bundle_id}` returns that descriptor, rejects pagination members, exposes fixed current-profile `history_mode='full'` and `blob_mode='full'` rather than user-tunable partial-export modes, serializes the resolved `reference_pack_mode`, and serializes `optional_sections[]` plus `required_capabilities[]` in canonical exact-token ascending order after duplicate coalescing; the emitted `manifest.json` uses that same canonical serialization rather than caller order.
  - Verifies: REQ-01-467..REQ-01-469, REQ-01-483..REQ-01-484
- **AC-275**: `POST /api/v1/incident-bundles/import` accepts only `multipart/form-data` with a required `boundary`, exactly two leaf parts named `metadata` and `file`, and no alternate upload framing; part order is non-semantic; the `metadata` part uses `Content-Type: application/json` or `application/json; charset=utf-8`, is UTF-8 and BOM-free, parses as one JSON object, and contains required `client_txn_id`; the `file` part media type is exactly one of `application/zip`, `application/x-tar`, `application/gzip`, `application/x-gzip`, or `application/octet-stream`; duplicate or unexpected parts, malformed metadata JSON, duplicate metadata keys, a metadata JSON value that is not an object, invalid part content type, or any envelope failure create no idempotency commit and no import job; replaying the same normalized metadata and file bytes by the same actor with the same `client_txn_id` returns the original accepted `202` result even when boundary text, part order, or advisory filename differ; the route returns `202` with the common job resource, uses `result_summary.code='incident_bundle_imported'` plus exactly one `resource_refs[]` item `{ kind='incident', id=<incident_id>, route='/api/v1/incidents/{incident_id}' }` on terminal success, and rejects clone, merge, identifier-remap, or remote-fetch modes.
  - Verifies: REQ-01-467..REQ-01-470, REQ-01-483, REQ-01-485, REQ-01-549..REQ-01-553
- **AC-276**: Incident-bundle routes use only `invalid_incident_bundle_request`, `incident_bundle_not_found`, `incident_bundle_export_rejected`, and `incident_bundle_import_rejected`; `invalid_incident_bundle_request` uses only `unsupported_upload_envelope`, `missing_required_part`, `duplicate_part`, `unexpected_part`, `invalid_part_content_type`, `invalid_metadata_encoding`, `malformed_metadata_json`, `request_not_object`, `missing_required_field`, `field_not_nullable`, `unknown_field`, `invalid_reference_pack_mode`, `invalid_optional_sections`, `invalid_required_capabilities`, `history_mode_not_supported`, and `blob_mode_not_supported`; multipart part-related failures populate `error.details.part_name`, and `invalid_part_content_type` also includes `error.details.received_content_type` plus `error.details.allowed_content_types[]`; export rejections surface only `missing_required_file` or `missing_required_blob`; and import rejections surface only `invalid_member_path`, `unsupported_member_type`, `checksum_mismatch`, `signature_mismatch`, `blob_hash_mismatch`, `duplicate_incident_id`, `unsupported_required_capability`, `remote_fetch_required`, `archive_extracted_bytes_exceeded`, `archive_compression_ratio_exceeded`, or `archive_member_count_exceeded`.
  - Verifies: REQ-01-471, REQ-01-486, REQ-01-553

- **AC-409**: Whole-incident portability preserves the boundary between incident data and deployment-local auth state: the exported bundle contains no login-capable local users, local-account credential lifecycle state such as password-hash state, active or pending TOTP state, bootstrap-token lookup state, auth bindings, bootstrap-completion markers, active sessions, active memberships, deployment-admin flags, or deployment-local administrative audit or idempotency state; importing that bundle into another deployment does not synthesize any of those states without explicit deployment-local administrative action; and historical actors import only as inert imported actors or historical descriptors, never as login-capable principals.
  - Verifies: REQ-02-204, REQ-02-249, REQ-04-038

### 9.12 Additional Base Profile criteria for deployment configuration contract

- **AC-294**: With no selector override, the deployment loads `/etc/cartulary/config.toml`; when `CARTULARY_CONFIG_FILE` is set to an alternate absolute path, that file is selected instead; after file load, `CARTULARY__ROOTS__BACKUP_STORAGE__PATH=/srv/cartulary/backups` overrides `roots.backup_storage.path`; and an unknown file key or unknown `CARTULARY__...` overlay key fails closed with `invalid_deployment_config`.
  - Verifies: REQ-01-455, REQ-04-058, REQ-04-066..REQ-04-071, REQ-04-077, REQ-04-109
- **AC-295**: Required runtime-root keys are present and use the standardized binding model; `deployment_profile='disconnected'` rejects `binding_kind='managed_service'` for `roots.database_storage`, `roots.object_storage`, or `roots.backup_storage`; `roots.reference_pack_storage`, `roots.temporary_work`, and `roots.export_outputs` reject any binding kind other than `filesystem_root`; and `roots.backup_storage` accepts only `filesystem_root` in `disconnected` and only `filesystem_root` or `managed_service` in `on_prem` or `cloud`.
  - Verifies: REQ-01-455, REQ-04-058, REQ-04-069, REQ-04-071..REQ-04-073, REQ-04-077
- **AC-296**: Relative paths, `~`, shell-variable forms, empty strings, NUL, lexical `.` or `..`, overlapping configured filesystem roots after canonicalization, non-writable filesystem roots, and effective writes or extracts that escape a configured root all fail closed with `invalid_deployment_config` and the appropriate `reason_code`.
  - Verifies: REQ-01-456, REQ-04-059, REQ-04-074..REQ-04-075, REQ-04-077
- **AC-297**: The canonical disconnected example using `/var/lib/cartulary/postgres`, `/var/lib/cartulary/object-store`, `/var/lib/cartulary/backups`, `/var/lib/cartulary/reference-packs`, `/var/lib/cartulary/tmp`, and `/var/lib/cartulary/exports` validates as a correct disconnected deployment configuration; omission of any required runtime-root key remains invalid at runtime and is not satisfied by hidden defaults.
  - Verifies: REQ-01-455, REQ-04-058, REQ-04-067, REQ-04-069, REQ-04-071..REQ-04-076
- **AC-298**: Invalid deployment configuration prevents HTTP listeners, WebSocket listeners, and background-job runners from starting; startup fails non-zero; and the surfaced error family is `invalid_deployment_config` with per-item `path`, `reason_code`, and `message`.
  - Verifies: REQ-04-066, REQ-04-077..REQ-04-078, REQ-04-109

- **AC-343**: On a fresh deployment with zero active deployment admins, no bootstrap-completion marker, `bootstrap.first_admin_manifest_path` set to a valid manifest path, and a valid `cartulary.bootstrap_admin.v1` manifest at that path, startup succeeds and listeners become available only after bootstrap completes; exactly one local user is created with `is_deployment_admin=true`, `is_active=true`, and `mfa_required=true`; no incident membership is created; and the same commit persists one deployment-local bootstrap-completion marker plus one deployment-local administrative audit event. A later valid password login for that user returns `401 error.code='mfa_setup_required'` until TOTP setup completes.
  - Verifies: REQ-01-121, REQ-01-530..REQ-01-536, REQ-02-007..REQ-02-008, REQ-02-202, REQ-02-246, REQ-04-028, REQ-04-038, REQ-04-087..REQ-04-090
- **AC-344**: On a fresh deployment with zero active deployment admins and no bootstrap-completion marker, a missing bootstrap path, unreadable bootstrap file, non-regular-file bootstrap path, malformed JSON manifest, schema-invalid manifest, or manifest email that conflicts with an existing local user causes startup to fail before any HTTP listener, WebSocket listener, or background-job runner starts; no partial bootstrap-created user, no partial bootstrap-completion marker, and no incident membership are left behind; and the surfaced error family is `invalid_deployment_config` using the exact bootstrap-specific `reason_code`.
  - Verifies: REQ-01-121, REQ-01-530..REQ-01-535, REQ-02-246, REQ-04-028, REQ-04-038, REQ-04-087..REQ-04-092
- **AC-345**: When at least one active deployment admin already exists, startup skips bootstrap preflight even if `bootstrap.first_admin_manifest_path` is present; no second deployment admin is created; and a stale, missing, or invalid bootstrap manifest at that path does not block ordinary startup in this state.
  - Verifies: REQ-01-121, REQ-01-530, REQ-01-533, REQ-01-535, REQ-04-028, REQ-04-089..REQ-04-091
- **AC-346**: When zero active deployment admins exist but a bootstrap-completion marker already exists, startup fails before listeners start with `invalid_deployment_config` and `reason_code='bootstrap_recovery_not_supported'`; the manifest is not consumed again; no replacement deployment admin is created implicitly; and deployment-local administrative audit coverage remains intact.
  - Verifies: REQ-01-121, REQ-01-530, REQ-01-533..REQ-01-535, REQ-02-007..REQ-02-008, REQ-02-202, REQ-02-246, REQ-04-028, REQ-04-038, REQ-04-090, REQ-04-092
- **AC-347**: For a bootstrap-created first deployment admin, the first successful password-authenticated credential flow reuses the ordinary local TOTP bootstrap contract exactly: `POST /api/v1/auth/login` returns `401 error.code='mfa_setup_required'` plus one `bootstrap_token` and `bootstrap_expires_at`; `POST /api/v1/auth/mfa/totp/begin` returns the pending enrollment; `POST /api/v1/auth/mfa/totp/complete` activates the factor; and first-time TOTP completion alone issues no session.
  - Verifies: REQ-01-121, REQ-01-536, REQ-04-084


### 9.13 Additional Base Profile criteria for resource-limit registry

- **AC-320**: With every declared `limits.*` key omitted, the effective deployment configuration resolves to the exact defaults declared in Core 04 §12.3.1; an explicit integer override on one or more declared keys replaces only that key; `limits.archives.max_compression_ratio=0` fails closed with `invalid_deployment_config` and `reason_code='value_below_minimum'`; `limits.archives.max_compression_ratio=1001` fails closed with `invalid_deployment_config` and `reason_code='value_above_maximum'`; string forms such as `limits.previews.max_previewable_payload_bytes='32MB'` fail closed with `invalid_deployment_config` and `reason_code='type_mismatch'`; undeclared keys such as `limits.view_query.max_sort_entries`, `limits.view_query.max_filter_entries`, `limits.records.max_changes`, or `limits.records.max_collection_actions` fail closed with `invalid_deployment_config` and `reason_code='unknown_key'`; and valid overrides to declared registry keys do not alter the fixed public ceilings `sort[]<=8`, `filters[]<=16`, `changes[]<=32`, or `collection_actions_v1.actions[]<=64`.
  - Verifies: REQ-04-066, REQ-04-077, REQ-04-079..REQ-04-081
- **AC-321**: `POST /api/v1/object-blobs` with `byte_size=536870912` succeeds, while `byte_size=536870913` fails before slot creation with `413`, `error.code='blob_create_rejected'`, `error.details.reason_code='byte_size_exceeds_limit'`, `error.details.requested_byte_size=536870913`, and `error.details.configured_limit_bytes=536870912`; the rejection creates no `object_blob_id`, upload target, or pending slot state.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-243..REQ-01-245, REQ-04-079..REQ-04-080
- **AC-322**: Preview-handle issuance for an otherwise available `text_inline` evidence record backed by a `104857600`-byte blob fails with `409`, `error.code='evidence_access_unavailable'`, and `reason_code='preview_payload_too_large'`, while download-handle issuance for that same evidence succeeds; a previewable payload at exactly `limits.previews.max_previewable_payload_bytes` succeeds when media class and preview kind are otherwise allowed.
  - Verifies: REQ-01-238, REQ-01-461, REQ-01-465, REQ-04-079..REQ-04-080
- **AC-323**: `POST /api/v1/import-sessions` rejects a CSV source of `33554433` bytes with `413 error.code='import_source_rejected' reason_code='csv_source_too_large'` and rejects an XLSX source of `67108865` bytes with `413 error.code='import_source_rejected' reason_code='xlsx_source_too_large'`; neither rejection creates an `import_session` or discovery job.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-473..REQ-01-475, REQ-04-079..REQ-04-081
- **AC-324**: An XLSX source below the raw source-byte ceiling but with `100001` rows, `257` columns, or `5000001` cells fails before apply with `413 error.code='import_source_rejected'` and the exact corresponding `reason_code` `import_rows_exceeded`, `import_columns_exceeded`, or `import_cells_exceeded`.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-474..REQ-01-475, REQ-04-079..REQ-04-080
- **AC-325**: Structured import of an archive-backed or XLSX-backed source that stays under raw source-byte ceilings but exceeds `2147483648` extracted bytes, `limits.archives.max_compression_ratio=100`, or `limits.archives.max_members=10000` fails before imported incident data becomes visible with `413 error.code='import_source_rejected'` and the exact corresponding `reason_code` `archive_extracted_bytes_exceeded`, `archive_compression_ratio_exceeded`, or `archive_member_count_exceeded`.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-474..REQ-01-475, REQ-04-079..REQ-04-080
- **AC-326**: Reference-pack import, reverify, or refresh of a bundle that exceeds `limits.reference_packs.max_extracted_bytes=536870912`, `limits.archives.max_compression_ratio=100`, or `limits.archives.max_members=10000` fails closed with `409 error.code='reference_pack_verification_failed'` and the exact corresponding `reason_code`; a bundle with `42949672960` extracted bytes is rejected under the same family because the reference-pack override is narrower than the incident-bundle override.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-481..REQ-01-482, REQ-04-079..REQ-04-080
- **AC-327**: Whole-incident import of a bundle with `42949672960` extracted bytes succeeds when all other bundle validations pass, because `limits.incident_bundles.max_extracted_bytes=68719476736` is the applicable extracted-bytes ceiling rather than the default archive ceiling.
  - Verifies: REQ-01-238, REQ-01-448..REQ-01-449, REQ-01-486, REQ-04-079..REQ-04-080
- **AC-328**: Whole-incident import of a bundle that exceeds `limits.incident_bundles.max_extracted_bytes=68719476736`, `limits.archives.max_compression_ratio=100`, or `limits.archives.max_members=10000` fails closed before structured incident state becomes visible with `409 error.code='incident_bundle_import_rejected'` and the exact corresponding `reason_code`.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-448..REQ-01-449, REQ-01-486, REQ-04-079..REQ-04-080



- **AC-359**: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` accepts only `sort[]` entries shaped as `{ "field_key": <stable key>, "direction": "asc"|"desc" }`; `sort[]` MAY be omitted or empty; a request with exactly `8` `sort[]` entries succeeds when otherwise valid; a request with `9` `sort[]` entries fails with `400 error.code='invalid_view_query'`, `error.details.reason_code='sort_count_exceeded'`, `error.details.field='sort'`, `error.details.requested_count=9`, and `error.details.max_count=8`; malformed entries, unknown members, duplicate normalized sort fields, or disallowed sort keys fail with `400 error.code='invalid_view_query'` and `error.details.reason_code` equal to `invalid_sort_entry`, `duplicate_sort_field`, or `sort_field_not_allowed` as applicable; and the server neither truncates nor partially honors an oversize `sort[]`.
  - Verifies: REQ-01-035, REQ-01-286, REQ-01-310
- **AC-360**: On the view-query route, omitted `sort` and `sort: []` are equivalent `no user sort override`; a request with exactly `16` `filters[]` entries succeeds when otherwise valid; a request with `17` `filters[]` entries fails with `400 error.code='invalid_view_query'`, `error.details.reason_code='filter_count_exceeded'`, `error.details.field='filters'`, `error.details.requested_count=17`, and `error.details.max_count=16`; on saved-view create or patch, omission of `query_json.sort` normalizes to persisted `[]`, omission of `query_json.filters` normalizes to persisted `[]`, `query_json.filters[]` persist in canonical `field_key asc` order, set-like filter operands persist only as unique normalized values in canonical ascending order, `prefix` filter values persist only in comparison-normalized form, and `full_text` filter queries persist only as the canonical unique normalized token string; saved-view create or patch with `9` `query_json.sort` entries or `17` `query_json.filters` entries fails with `400 error.code='invalid_mutation_payload'`, `error.details.reason_code` equal to `sort_count_exceeded` or `filter_count_exceeded`, `error.details.field` equal to `query_json.sort` or `query_json.filters`, and matching `requested_count` plus `max_count`; on both routes, explicit `group_by: null` fails closed, and omission is the only `Group: None` representation.
  - Verifies: REQ-01-035, REQ-01-038, REQ-01-046, REQ-01-142, REQ-01-145, REQ-01-146, REQ-02-010, REQ-02-155, REQ-03-224, REQ-03-227
- **AC-361**: For a view with a declared `default_sort`, successful `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` responses always include `meta.query`; `meta.query.sort` returns the effective tuple formed by user override first, then undeclared default-tail entries, then `record_id asc` if still absent; and `meta.query.group_by` is omitted when grouping is inactive.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-046
- **AC-362**: `GET /api/v1/view-schemas/{view_schema_id}` exposes `sort_fields` for every current-profile workbook surface; the Timeline field entry for `timeline.occurred_at` exposes `header_sort_field_key='timeline.sort_ts'`; and `record_id` does not appear in any client-sortable whitelist.
  - Verifies: REQ-01-286, REQ-01-310, REQ-01-312, REQ-01-323, REQ-01-326, REQ-01-328, REQ-01-329, REQ-01-331, REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-499, REQ-01-503..REQ-01-509
- **AC-363**: Header sort on a visible non-sortable collection field such as `timeline.host_refs`, `timeline.tags`, `note.tags`, or `decision.support_refs` does not change row order and does not synthesize a client sort.
  - Verifies: REQ-01-310, REQ-03-223
- **AC-364**: Any grouped workbook surface whose active `view_schema` declares `grouping_fields`, including `comm_log`, `handoff`, `status_review`, and `lesson`, and, when exposed, `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1`, exposes a grouping control offering exactly `Group: None` plus the surface's declared `grouping_fields` in discovery order and no undeclared key; uses derived non-record group headers keyed only by the active `group_by` and matching `group_values[group_by]`; exposes exactly one outline level; keeps mutation targets row-only while the current row sort continues to apply within each group; and uses the generic comparator order on non-Timeline grouped surfaces while Timeline alone uses the explicit override.
  - Verifies: REQ-03-225..REQ-03-235
- **AC-365**: User-specified sorts place `null` values last in both directions for timestamp, date, text, enum, boolean, and numeric sortable fields in the current profile.
  - Verifies: REQ-01-310
- **AC-366**: Querying `cartulary.view.evidence.v1` returns full `view_row_v1` rows whose `cells` include `evidence.collector_party_id` and `evidence.source_party_id` even though both fields are default-hidden in the grid, and querying `cartulary.view.task_requests.v1` returns full `view_row_v1` rows whose `cells` include `task.requester_party_id` even though that field is default-hidden. Hidden writable fields appear in the full row payload and do not require a second read path.
  - Verifies: REQ-01-034, REQ-01-036, REQ-01-310, REQ-01-328, REQ-01-336, REQ-03-247
- **AC-367**: Successful `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` and `PATCH /api/v1/records/{record_id}` responses return `data.view_schema_id`, `data.change_set_id`, and `data.row`, where `data.row` is a full `view_row_v1`; and when a clearable direct-reference field such as `task.requester_party_id`, `evidence.collector_party_id`, or `evidence.source_party_id` is authoritatively null after commit, the field remains present inside `data.row.cells` as `{ "value": null }` rather than being omitted.
  - Verifies: REQ-01-034, REQ-01-036, REQ-01-070, REQ-01-310, REQ-01-328, REQ-01-336
- **AC-368**: For a replayable `record_changed` event with one or more `affected_views[]` entries using `change_kind='patch'`, `changed_field_keys[]` is present, non-empty, free of duplicate `field_key` values, and serialized in canonical ascending `field_key` order; `affected_views[]` is present, non-empty, free of duplicate `view_schema_id` values, and serialized in canonical ascending `view_schema_id` order; and each such `affected_views[].patch_cells` is a sparse `view_row_patch_v1`: `record_id` and authoritative `row_version` are top-level, only changed fields appear in `patch_cells.cells`, only changed grouping scalars appear in `patch_cells.group_values`, an included cleared field such as `evidence.collector_party_id` is serialized as `{ "value": null }`, omitted sibling fields remain unchanged client-side, and when the server cannot safely express the committed result as such a sparse row patch it emits `change_kind='invalidate'` rather than guessing a partial patch.
  - Verifies: REQ-01-034, REQ-01-267, REQ-01-310, REQ-03-097


### 9.14 Additional Base Profile criteria for backup and restore contract

- **AC-398**: A deployment-local operator can determine structured metadata for the most recent successful retained `backup_set`, including `backup_set_id`, `consistency_point_at`, `created_at`, `retained_until`, `verification_state`, `last_verified_restore_at`, `postgres_restore_anchor`, and `object_store_restore_anchor`; `verification_state` uses only `unverified`, `verified`, or `failed`; `last_verified_restore_at` is `null` only while `verification_state='unverified'`; the latest successful retained `backup_set` has `consistency_point_at` no older than 24 hours; and each successful retained `backup_set` shows `retained_until >= created_at + 30 days`.
  - Verifies: REQ-01-572..REQ-01-573
- **AC-399**: Restoring the latest successful retained `backup_set` into a fresh environment restores Postgres and object-store contents from that same `backup_set`, rebuilds projections, opens at least one incident normally, executes at least one built-in workbook query when incident data is present, and preserves authoritative `incident_id`, `record_id`, `row_version`, change-set count, blob hashes, and evidence/blob lifecycle consistency.
  - Verifies: REQ-01-571, REQ-01-575, REQ-01-423..REQ-01-424, REQ-01-577
- **AC-400**: If the selected `backup_set` is missing a required Postgres artifact, required object-store artifact, or required checksum or integrity proof for the chosen backup mechanism, restore fails before the target environment becomes visible or ready.
  - Verifies: REQ-01-575..REQ-01-576
- **AC-401**: Full restore verification runs in an isolated environment at least every 7 days and after a change to the backup mechanism, `roots.database_storage`, `roots.object_storage`, or `roots.backup_storage` binding; a successful verification sets `verification_state='verified'` with non-null `last_verified_restore_at`; a failed verification sets `verification_state='failed'` with non-null `last_verified_restore_at`; and a failed verification is never represented as verified or ready.
  - Verifies: REQ-01-572, REQ-01-578
- **AC-402**: The public route inventory under `/api/v1/` and `/ws/v1/` exposes no backup, restore, or restore-verification family; and any built-in operator-facing backup or restore control surface is deployment-local, requires `deployment_admin`, and is not incident-scoped.
  - Verifies: REQ-01-570, REQ-04-106
- **AC-403**: `roots.backup_storage` is present in the effective deployment configuration; the disconnected binding uses `binding_kind='filesystem_root'`; the on-prem or cloud binding uses only `filesystem_root` or `managed_service`; `/var/lib/cartulary/backups` is the canonical disconnected example path; `roots.export_outputs` and `roots.temporary_work` are not treated as authoritative backup roots; and backup artifacts or restore-verification extracts that carry incident data remain on encrypted storage.
  - Verifies: REQ-04-053, REQ-04-058, REQ-04-071..REQ-04-073, REQ-04-076, REQ-04-107..REQ-04-108

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

## 12. Deployment configuration contract

### 12.1 Scope and owner

**REQ-04-066**
This section owns the operator-facing deployment configuration surface for application public origin, runtime roots, resource limits, and startup validation.
Profiles: base
Verified by: AC-294, AC-298, AC-320

### 12.2 Canonical artifact and discovery

**REQ-04-067**
By default, the deployment configuration MUST be loaded from the canonical TOML artifact at `/etc/cartulary/config.toml`.
Profiles: base
Verified by: AC-294, AC-297

**REQ-04-068**
`CARTULARY_CONFIG_FILE` MAY override the config file path. When present, it MUST itself be an absolute path in the runtime where interpreted.
Profiles: base
Verified by: AC-294

**REQ-04-069**
The deployment configuration file MUST declare `config_schema_id = "cartulary.deployment_config.v1"`, required `deployment_profile` with one of `disconnected`, `on_prem`, or `cloud`, and required `application.public_origin`. Keys other than those defined by the selected configuration schema version are invalid.
Profiles: base
Verified by: AC-294, AC-295, AC-297

**REQ-04-070**
After file load, environment variables prefixed `CARTULARY__` MUST overlay nested keys by splitting segments on `__`, lowercasing each segment, and joining them with dots. Overlay keys that do not map to declared deployment-configuration keys are invalid. `CARTULARY_CONFIG_FILE` is selector-only and MUST NOT participate in that overlay mapping.
Profiles: base
Verified by: AC-294

### 12.3 Key registry and binding model


#### Application public origin

**REQ-04-109**
The stable deployment-configuration key for the browser application origin MUST be `application.public_origin`. It MUST be an absolute `http` or `https` origin and MUST NOT include userinfo, path, query, or fragment. The public WebSocket implementation MUST validate cookie-authenticated browser `Origin` values against this configured origin before joining an incident-scoped stream.
Profiles: base
Verified by: AC-131, AC-298

#### 12.3.1 Resource-limit registry

**REQ-04-079**
The stable deployment-configuration keys for the resource-limit registry MUST be the following, and omitted keys MUST resolve to exactly the declared defaults.
Profiles: base
Verified by: AC-320, AC-321, AC-322, AC-323, AC-324, AC-325, AC-326, AC-327, AC-328

| Key | Default | Applies to |
| --- | --- | --- |
| `limits.object_blobs.max_declared_byte_size` | `536870912` | Declared `byte_size` ceiling for `POST /api/v1/object-blobs`. |
| `limits.imports.max_csv_source_bytes` | `33554432` | Raw CSV source-byte ceiling for `POST /api/v1/import-sessions`. |
| `limits.imports.max_xlsx_source_bytes` | `67108864` | Raw XLSX source-byte ceiling for `POST /api/v1/import-sessions`. |
| `limits.imports.max_rows` | `100000` | Structured-import row ceiling per selected unit after parsing. |
| `limits.imports.max_columns` | `256` | Structured-import column ceiling per selected unit after parsing. |
| `limits.imports.max_cells` | `5000000` | Structured-import cell ceiling per selected unit after parsing. |
| `limits.archives.default_max_extracted_bytes` | `2147483648` | Default extracted-bytes ceiling for structured import and other archive-backed workflows that do not declare a narrower or wider family override. |
| `limits.archives.max_compression_ratio` | `100` | Maximum allowed extracted-bytes to compressed-bytes multiplier for archive-backed workflows. |
| `limits.archives.max_members` | `10000` | Maximum allowed extracted regular-file member count for archive-backed workflows. |
| `limits.reference_packs.max_extracted_bytes` | `536870912` | Extracted-bytes ceiling for reference-pack import, reverify, and refresh. |
| `limits.incident_bundles.max_extracted_bytes` | `68719476736` | Extracted-bytes ceiling for whole-incident bundle import. |
| `limits.previews.max_previewable_payload_bytes` | `33554432` | Maximum payload size eligible for any ordinary inline preview issuance. |
| `limits.previews.max_text_inline_bytes` | `1048576` | Maximum payload size eligible for `text_inline` preview issuance. |

**REQ-04-080**
All resource-limit keys above MUST use integer values only. Human-readable unit strings such as `32MB` or `2GiB` are invalid. For this registry:

- byte-count keys are bytes,
- count keys are integer counts,
- `extracted_bytes` means the sum of extracted regular-file byte lengths after archive or workbook expansion,
- `compressed_bytes` means the byte length of the uploaded or staged container before extraction,
- a workflow breaches the extracted-bytes ceiling when `extracted_bytes > configured_limit`, breaches the compression-ratio ceiling when `extracted_bytes > compressed_bytes * limits.archives.max_compression_ratio`, and breaches the member ceiling when extracted regular-file member count exceeds `limits.archives.max_members`,
- structured import MUST use `limits.archives.default_max_extracted_bytes`, reference-pack flows MUST use `limits.reference_packs.max_extracted_bytes`, and whole-incident bundle import MUST use `limits.incident_bundles.max_extracted_bytes`,
- XLSX MUST be treated as a ZIP-backed workbook input for extracted-bytes, compression-ratio, and member-count checks,
- preview ceilings are independent from storage ceilings, so a blob MAY be storable or downloadable while still non-previewable,
- this registry does not include and MUST NOT be used to override the fixed public-contract ceilings on `sort[]`, `filters[]`, `changes[]`, or `collection_actions_v1.actions[]`; those ceilings are owned by Core 01 and remain deployment-invariant in the current profile.
Profiles: base
Verified by: AC-320, AC-322, AC-323, AC-324, AC-325, AC-326, AC-327, AC-328

**REQ-04-081**
Every resource-limit key above MUST be present in the effective configuration either explicitly or by omitted-key default resolution. Each configured value MUST be an integer within the closed deployment-configuration domain for that key family: byte-count and count limits MUST be in `1..9223372036854775807`, and `limits.archives.max_compression_ratio` MUST be in `1..1000`. Values below the minimum MUST fail with `invalid_deployment_config` and `reason_code='value_below_minimum'`. Values above the maximum MUST fail with `invalid_deployment_config` and `reason_code='value_above_maximum'`.
Profiles: base
Verified by: AC-320

**REQ-04-071**
The required stable deployment-configuration keys for runtime roots MUST be:

- `roots.database_storage`,
- `roots.object_storage`,
- `roots.backup_storage`,
- `roots.reference_pack_storage`,
- `roots.temporary_work`,
- `roots.export_outputs`.
Profiles: base, reference_pack
Verified by: AC-294, AC-295, AC-297, AC-403

**REQ-04-072**
Each root binding MUST be a typed object. `binding_kind` MUST use the closed vocabulary `filesystem_root` or `managed_service`. When `binding_kind='filesystem_root'`, `path` is required and `service_ref` MUST NOT be present. When `binding_kind='managed_service'`, `service_ref` is required and `path` MUST NOT be present. Credentials, vendor-specific connection properties, reverse-proxy settings, and other deployment-adjacent configuration remain outside this contract.
Profiles: base
Verified by: AC-295

**REQ-04-073**
`roots.reference_pack_storage`, `roots.temporary_work`, and `roots.export_outputs` MUST use `binding_kind='filesystem_root'`. `roots.backup_storage` is governed by §12.3.3. When `deployment_profile='disconnected'`, `roots.database_storage` and `roots.object_storage` MUST also use `binding_kind='filesystem_root'`. In `deployment_profile='on_prem'` or `deployment_profile='cloud'`, those two roots MAY use `binding_kind='managed_service'`.
Profiles: base, reference_pack
Verified by: AC-295, AC-297, AC-403

#### 12.3.2 First-admin bootstrap binding

**REQ-04-087**
The stable deployment-configuration key for the current-profile first-admin bootstrap path MUST be `bootstrap.first_admin_manifest_path`. Omission means no bootstrap manifest path is configured. This key participates in the same deployment-configuration artifact, selector, and overlay rules defined by this section.
Profiles: base
Verified by: AC-343, AC-344, AC-345, AC-346

**REQ-04-088**
When present, `bootstrap.first_admin_manifest_path` MUST be an absolute POSIX path in the runtime where interpreted. Relative paths, `~`, shell-variable expansion forms, empty strings, NUL, and lexical `.` or `..` segments are invalid. Trailing-slash normalization MUST be deterministic. The referenced path MUST identify one deployment-local regular file rather than a writable root.
Profiles: base
Verified by: AC-343, AC-344

**REQ-04-089**
The running application MUST consume first-admin bootstrap material only from the file at `bootstrap.first_admin_manifest_path`. It MUST NOT create the first deployment admin directly from environment-variable values, browser UI input, unauthenticated HTTP input, object storage, incident bundles, incident records, or manual database mutation. Helper tooling MAY write the manifest file, but the runtime contract remains one deployment-local file.
Profiles: base
Verified by: AC-343, AC-344, AC-345

**REQ-04-090**
After deployment-configuration validation and before listeners start, startup MUST perform first-admin bootstrap preflight. If zero active `deployment_admin` users exist and no bootstrap-completion marker exists, a configured bootstrap manifest path and a valid manifest file are required or startup MUST fail closed. If at least one active `deployment_admin` exists, bootstrap preflight is skipped. If zero active `deployment_admin` users exist and a bootstrap-completion marker already exists, startup MUST fail closed because the current profile does not define lost-admin recovery.
Profiles: base
Verified by: AC-343, AC-344, AC-345, AC-346

**REQ-04-091**
Manifest presence and content MUST be validated only when bootstrap preconditions require them. A stale, missing, or invalid file at the configured path MUST NOT block ordinary startup when at least one active `deployment_admin` already exists.
Profiles: base
Verified by: AC-345

**REQ-04-092**
Bootstrap-preflight failures MUST surface through `invalid_deployment_config` and MUST use the bootstrap-specific `reason_code` tokens defined in REQ-04-077.
Profiles: base
Verified by: AC-344, AC-346

#### 12.3.3 Backup storage binding

**REQ-04-107**
The stable deployment-configuration key for operational backup artifacts and restore-verification extracts MUST be `roots.backup_storage`.
Profiles: base
Verified by: AC-403

**REQ-04-108**
`roots.backup_storage` is a persistent runtime root. When `deployment_profile='disconnected'`, it MUST use `binding_kind='filesystem_root'`. In `deployment_profile='on_prem'` or `deployment_profile='cloud'`, it MAY use `binding_kind='filesystem_root'` or `binding_kind='managed_service'`. `roots.export_outputs` and `roots.temporary_work` MUST NOT be treated as authoritative backup roots or used to satisfy the retained-backup floor.
Profiles: base
Verified by: AC-403

### 12.4 Filesystem-root path contract

**REQ-04-074**
For `binding_kind='filesystem_root'`, `path` MUST be an absolute POSIX path in the runtime where interpreted. Relative paths, `~`, shell-variable expansion forms, empty strings, NUL, and lexical `.` or `..` segments are invalid. Trailing-slash normalization MUST be deterministic. If a configured path already exists, canonicalization MUST resolve symlinks before enforcement. After canonicalization, configured filesystem roots MUST remain distinct from one another. A configured filesystem root that must be writable for its declared role but is not writable at startup is invalid.
Profiles: base, reference_pack
Verified by: AC-296, AC-297

**REQ-04-075**
Packaged read-only resources MAY resolve from install or package locations. Any operator-owned writable or persistent location MUST derive from the deployment configuration contract. Archive extraction, imports, uploads, preview generation, report builds, and temp-file writes MUST fail if the effective target escapes the configured filesystem root.
Profiles: base, reference_pack
Verified by: AC-296

### 12.5 Canonical disconnected-layout defaults

**REQ-04-076**
For official disconnected-deployment examples and scaffolding, the canonical filesystem-root paths are:

- `/var/lib/cartulary/postgres`,
- `/var/lib/cartulary/object-store`,
- `/var/lib/cartulary/backups`,
- `/var/lib/cartulary/reference-packs`,
- `/var/lib/cartulary/tmp`,
- `/var/lib/cartulary/exports`.

These canonical paths are example and scaffolding defaults only. They MUST NOT become hidden runtime fallbacks. If any required root key is missing from the effective deployment configuration, startup MUST fail closed.
Profiles: base, reference_pack
Verified by: AC-297, AC-403

### 12.6 Validation error contract and startup behavior

**REQ-04-077**
Deployment-configuration validation failures MUST surface the top-level error code `invalid_deployment_config`. Unknown file keys and unknown overlay keys MUST fail validation rather than being ignored. The structured error details MUST include one or more items with `path`, `reason_code`, and `message`. The minimum `reason_code` registry is:

- `config_file_not_found`,
- `config_parse_error`,
- `unsupported_config_schema_id`,
- `missing_required_key`,
- `unknown_key`,
- `type_mismatch`,
- `invalid_enum`,
- `invalid_origin`,
- `path_not_absolute`,
- `path_forbidden_segment`,
- `path_overlap`,
- `profile_incompatible_binding`,
- `path_not_writable`,
- `bootstrap_manifest_path_missing`,
- `bootstrap_manifest_not_readable`,
- `bootstrap_manifest_not_regular_file`,
- `bootstrap_manifest_parse_error`,
- `bootstrap_manifest_schema_invalid`,
- `bootstrap_email_conflict`,
- `bootstrap_recovery_not_supported`,
- `bootstrap_persist_failed`,
- `value_below_minimum`,
- `value_above_maximum`.
Profiles: base
Verified by: AC-294, AC-295, AC-296, AC-298, AC-320

**REQ-04-078**
Validation of the effective deployment configuration MUST complete before any HTTP listener, WebSocket listener, or background-job runner starts. Invalid deployment configuration MUST cause non-zero process exit. The implementation MUST NOT partially start and then discover deployment-configuration invalidity later during analyst workflow.
Profiles: base
Verified by: AC-298
