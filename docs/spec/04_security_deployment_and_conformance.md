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
Self-service password change, self-service TOTP replacement, administrator password reset, administrator TOTP reset, account disablement, or an explicit deployment-admin revoke-all action MUST revoke all active sessions for that user immediately. Self-service display-name changes and account density-preference changes MUST NOT revoke active sessions by themselves.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-430, AC-431

**REQ-04-095**
Successful enterprise-auth binding rotate or retire actions MUST revoke all active sessions for the target user immediately. Successful enterprise-auth binding create MAY leave current sessions unchanged.
Profiles: enterprise_authentication
Verified by: AC-350, AC-351, AC-352

**REQ-04-017**
When a fresh transaction commits the loss of a user's membership in one incident, and
only after that commit succeeds, the running application process MUST terminate only
that user's WebSocket subscriptions for that incident. Those subscriptions receive
`session_revoked` with `reason_code='incident_access_revoked'` before close, and future
incident-scoped authorization checks for that incident fail closed. Exact replay,
rollback, commit failure, pre-commit failure, and rejected or no-op mutations produce
no terminal effect. The committed administrative-audit event identifier is the
internal effect identity.

Membership loss MUST NOT revoke the user's otherwise valid account session, terminate
subscriptions for another user, or terminate that user's subscriptions to another
incident. Delivery is process-local under REQ-04-145's contract for exactly one active
application process; the current profile requires neither a durable outbox nor
cross-process retry. The required process-local terminal-effect capability MUST be complete
and validated before serving begins. A socket write failure follows the existing
connection cleanup and telemetry contract and MUST NOT reinterpret the already
committed membership mutation.
Profiles: base
Verified by: AC-123, AC-131, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

### 1.2 Enterprise Authentication Extension Profile

**REQ-04-018**
If the implementation claims the **Enterprise Authentication Extension Profile**, it MUST support provider-backed identities through an `auth_providers` and `auth_identities` model equivalent to the source artifact. The current profile standardizes OIDC authorization-code flow with PKCE `S256` and `nonce`, plus SAML SP-initiated flow only.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235, AC-288, AC-289, AC-290, AC-291, AC-293

OIDC is the preferred enterprise path. SAML is the secondary enterprise path when required by the environment.

A runtime MUST NOT claim the Enterprise Authentication Extension Profile unless production OIDC and SAML verification are provided through a deployment-owned protocol-verification boundary. Deterministic provider responses, fixed OIDC codes, fixture SAML assertions, or JSON assertion shims are harness evidence only and MUST NOT be used by production route handlers or counted as production provider interoperability.

**REQ-04-019**
External identities MUST map to the same internal user identity used for attribution so that audit semantics remain unchanged. `provider_subject` MUST be the authoritative external identifier. Successful provider authentication MUST NOT auto-create a local user, auto-create incident membership, auto-create an `auth_identity`, or map provider group claims into incident roles.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235, AC-292, AC-293

**REQ-04-020**
When the Enterprise Authentication Extension Profile is implemented, successful provider authentication MUST terminate into the same server-managed session contract used by the base profile so the remaining API surface stays provider-agnostic. Completion routes MUST be single-use and replay-resistant, and provider configuration, raw assertions, and any provider tokens MUST remain server-side deployment-local state. Enterprise provider definitions MUST come from the startup-only deployment manifest defined by §12.3.4, not from browser state, public API requests, incident records, workbook rows, or generated test fixtures outside a harness-owned runtime. Cookie-based enterprise-auth browser binding MUST use an `HttpOnly` `Secure` cookie scoped to `Path=/api/v1/auth` with `SameSite=Lax`; the current profile MUST NOT require or accept `SameSite=None` for the browser-binding cookie. SAML ACS MUST use a same-origin completion hop so the browser-binding cookie is checked on a same-site request before a session is issued.
Profiles: enterprise_authentication
Verified by: AC-036, AC-235, AC-290, AC-291, AC-293, AC-433, AC-434, AC-436

**REQ-04-093**
The current profile MUST use the deployment-admin binding-management surface defined by Core 01 §20 as the only normative runtime mechanism for creating, rotating, and retiring enterprise-auth bindings. Provider-definition configuration is not a runtime capability in the current profile and is governed only by §12.3.4 startup reconciliation. A startup-only linkage artifact MAY exist only as helper tooling that produces the same binding state. It MUST NOT be a second authoritative runtime mechanism. The current profile defines no self-service account linking, no JIT provisioning, no SCIM provisioning, and no enterprise-only user-creation path.
Profiles: enterprise_authentication
Verified by: AC-348, AC-352, AC-435, AC-436

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
API routes, preview or download handle issuance and redemption, job polling, job cancellation, and WebSocket incident subscriptions MUST re-derive authorization from the caller's current scope membership and role at request time. Incident-scoped jobs MUST use current incident membership and role. If an incident-scoped job is admitted by a route that also requires `deployment_admin`, job polling MUST require both current `deployment_admin` and current incident membership, and job cancellation MUST require current `deployment_admin` plus either the submitter relationship or current incident role `admin`. Deployment-scoped jobs MUST use the owning route family's deployment-scoped authorization contract. Reference Pack deployment-scoped jobs are owned by the `/api/v1/reference-packs` route family and therefore require current `deployment_admin` for polling and cancellation; the submitting user relationship alone is not sufficient after demotion or any other loss of current `deployment_admin`. Unauthorized common job reads or cancel requests continue to use the common job visibility contract in Core 01 §3.3.9.1. Incident Bundle export jobs are incident-scoped jobs with the combined deployment-admin-and-incident-membership policy; Incident Bundle import jobs remain deployment-scoped `deployment_admin` jobs because no target incident exists before successful import.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-254, AC-255, AC-257, AC-260, AC-261, AC-427

### 2.0A Evidence route-family authorization and concealment

**REQ-04-156**
The base profile MUST authorize the six Evidence operations in Table 2-A against
the current authenticated session, active account, current incident membership,
current incident role, and current incident lifecycle. Possession of an opaque
upload target, preview handle, or download handle MUST NOT replace any current
authorization check. `deployment_admin=true` without current membership in the
addressed incident is insufficient for every operation in this table.

**Table 2-A. Evidence operation authorization matrix**

| Operation | Permitted current incident roles | Closed-incident rule | Additional current-state rechecks |
| --- | --- | --- | --- |
| `POST /api/v1/object-blobs` | `editor`, `reviewer`, `admin` | Reject the fresh write. | Incident scope decoded from the bounded request prefix; current incident visibility. |
| `PUT /api/v1/object-uploads/{upload_token}` | `editor`, `reviewer`, `admin` | Reject the upload and leave the slot for ordinary timeout and cleanup. | Issuing session, actor, incident and blob binding; accepted contract; required method and headers; declared size and optional expected hash; issue and expiry times; pending single-upload lease state. |
| `POST /api/v1/evidence-records/{record_id}/attach-blob` | `editor`, `reviewer`, `admin` | Reject the fresh write. | Evidence visibility; current Evidence and blob state; blob visibility; current Evidence row version; capability and accepted-contract binding when finalization consumes the uploaded object. |
| `POST /api/v1/evidence-records/{record_id}/preview-handle` | `viewer`, `editor`, `reviewer`, `admin` | Permit the read when every other check succeeds. | Evidence visibility; current Evidence and blob state; preview policy; current handle-issuance eligibility. |
| `POST /api/v1/evidence-records/{record_id}/download-handle` | `viewer`, `editor`, `reviewer`, `admin` | Permit the read when every other check succeeds. | Evidence visibility; current Evidence and blob state; current handle-issuance eligibility. |
| `GET /api/v1/evidence-handles/{handle_token}` | `viewer`, `editor`, `reviewer`, `admin` | Permit the read when every other check succeeds. | Issuing-session, incident, Evidence, blob, kind, filename, disposition, and preview-kind bindings; current Evidence and blob state; expiry; consumption state. |

Every request or redeem MUST re-evaluate the applicable Table 2-A state. A
state cached when a capability was issued MUST NOT authorize later use. The
route family MUST apply the failure-precedence sequence in Table 2-B from top
to bottom and MUST stop at the first failing gate.

**Table 2-B. Evidence failure precedence**

| Precedence | Gate | Required behavior |
| ---: | --- | --- |
| 1 | HTTP method and basic request framing | Reject unsupported method or unusable framing before operation-specific processing. |
| 2 | Authentication and current session validity | Authenticate before opaque-capability lookup or resource-specific disclosure. |
| 3 | Cookie-authenticated CSRF for `POST` and `PUT` | Apply the Core CSRF contract before scope, role, capability, or body diagnostics. |
| 4 | Minimal bounded scope decode | Decode only the path token or identifier, or the body-carried `incident_id` for blob-slot creation, needed to establish scope. This decode MUST NOT disclose resource-specific diagnostics to an unauthenticated or unauthorized caller. |
| 5 | Current incident or resource visibility | Conceal absent, foreign, and hidden incident resources before role, lifecycle, version, blob, or validation details. |
| 6 | Minimum current role | Apply the exact Table 2-A role set. A visible resource with an insufficient role uses the ordinary authorization-denied outcome. |
| 7 | Capability binding and expiry | Validate all Core 01 bindings, current ownership, expiry, revocation, and single-use or single-upload state. |
| 8 | Full body and header semantic validation | Apply the route-owned closed request, header, size, and hash contracts only after preceding gates succeed. |
| 9 | Incident and resource lifecycle | Apply current closed-incident, Evidence, blob, quarantine, deletion, and availability rules. |
| 10 | Row version and idempotency | Apply strict version and route-owned replay or conflict rules. |
| 11 | Object-store access | Invoke object storage only after every preceding gate succeeds. |

Unknown, foreign-incident, cross-session, and revoked opaque capabilities MUST
use the applicable Core 01 constant-disclosure not-found-or-revoked outcome.
Hidden Evidence resources MUST NOT disclose lifecycle state, row version, blob
existence, capability validity, or request-validation detail. A visible resource
with an insufficient role MUST use the Core 01 authorization-denied outcome. A
visible resource rejected by current lifecycle state MUST use the exact Core 01
route-owned unavailable or rejected outcome and safe `reason_code`.
Authentication or session failure MUST precede opaque handle lookup.

The executable authorization matrix MUST define, for every denial cell, the
exact HTTP status, `error.code`, safe `error.details.reason_code` or explicit
absence of that member, concealment posture, object-store invocation count,
handle-consumption result, and durable-effect result. Each value MUST be copied
from the applicable Core 01 route, envelope, error, and reason-code contract;
this requirement creates no error token or reason token. A denied operation
MUST make no object-store call, consume no handle, and commit no idempotency,
source, history, projection, revision, Collaboration, or other durable effect.

State changes after capability issuance MUST have the exact consequences in
Table 2-C.

**Table 2-C. Evidence capability state-change consequences**

| State change after issuance | Upload target | Preview or download handle |
| --- | --- | --- |
| Session logout, revocation, or expiry | Reject. | Reject. |
| Incident membership removed | Reject. | Reject. |
| `editor` downgraded to `viewer` | Reject upload and finalization. | Permit read use only while every other check remains valid. |
| Incident becomes closed | Reject upload and finalization. | Permit read use only while every other check remains valid. |
| Blob detached or replaced | Not applicable to the pending upload. | Invalidate the existing handle. |
| Blob becomes `pending`, `failed`, missing, or inconsistent | Reject or remain unusable under the route-owned state outcome. | Fail with the exact Core 01 unavailable outcome. |
| Blob or Evidence becomes quarantined | Reject association or finalization. | Invalidate the existing handle. |
| Evidence is deleted and later restored | Reject any use that no longer satisfies current state. | The pre-delete handle remains invalid; restore MUST NOT resurrect it. |
| Download redeem starts successful byte delivery | Not applicable. | Consume the single-use download handle. |
| Object bytes are unavailable before byte delivery | Not applicable. | Fail without consuming the handle. |
| Preview redeem succeeds | Not applicable. | Keep the preview handle reusable until expiry unless another state change invalidates it. |

Core 01 remains the sole owner of Evidence route shapes, request and response
envelopes, capability and handle bindings, lifetimes, consumption states, and
public error vocabularies. This section creates no new route, role, error token,
reason token, public resource, or object-store credential surface.
Profiles: base
Verified by: AC-543

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
The base profile MUST separate application-level deployment administration from incident-scoped data authorization. A conformant deployment MUST define one current-profile boolean capability named `deployment_admin`. The first holder of this capability MUST be provisioned only through the deployment-local bootstrap-admin manifest mechanism defined by Core 01 §3.3.5.1 and Core 04 §12.3.2, not through a public or incident-scoped surface.

For the current profile, `deployment_admin` authorization is exactly:

| Route or operation family | `deployment_admin` requirement |
| --- | --- |
| User list/get/create/patch | Required. |
| Administrative password reset, TOTP reset, revoke-all | Required. |
| Current-account profile/preferences under `/api/v1/account/*` | Not required and not sufficient for cross-user access; these routes are current authenticated user only. |
| Deployment administration browser context at `/deployment-administration` | Required for access to the context; insufficient by itself for incident data. |
| Deployment administrative-audit read at `GET /api/v1/administrative-audit-events` | Required. |
| Incident membership audit read at `GET /api/v1/incidents/{incident_id}/membership-audit-events` | Insufficient by itself; current membership in that incident with role `admin` is required. |
| Enterprise-auth binding create/rotate/retire | Required when the extension is claimed. |
| Every reference-pack list, singleton get, import, activate, disable, reverify, and refresh route in Core 01 §17.4 | Required when the extension is claimed; the current profile defines no non-admin raw reference-pack metadata view. |
| Incident-bundle import | Required when the extension is claimed. |
| Incident-bundle export | Required and current membership in the exported incident. |
| Poll/cancel a deployment-scoped job admitted under one of these families | Required. |
| Poll an incident-scoped job that also requires deployment administration | Required plus current incident membership. |
| Cancel such an incident-scoped job | Required plus submitter status or current incident role `admin`. |
| Create an incident | Not required. Any active authenticated account may create one. |
| Read or modify incident data | Insufficient by itself; ordinary incident membership and role remain required. |
| Network Flow Activity routes under `/api/v1/incidents/{incident_id}/network-flow` when the extension is claimed | Not required and insufficient by itself; current incident membership and the Network Flow route-family role matrix are required. |
| Manage incident memberships | Insufficient by itself; current incident role `admin` is required. |
| Read extension-claim discovery | Not required; any authenticated session may read it. |
| Configure enterprise providers | Not a runtime capability; startup configuration only through `enterprise_authentication.provider_manifest_path`. |
| Invoke recovery CLI | Not a runtime capability and not authorized by `deployment_admin`; local operator authorization applies. |

This matrix is exhaustive for current-profile public route families and deployment-local operator families that reference `deployment_admin`. Future granular deployment capabilities require a later profile or an explicit versioned capability registry. A current v1 implementation MUST NOT silently narrow or widen `deployment_admin` by local policy. Holding `deployment_admin` MUST continue not to disclose incident content without ordinary incident membership.

Deployment-admin route families MUST distinguish authentication failure from authorization denial. Missing, invalid, expired, revoked, or inactive sessions fail with `401` and `error.code='session_required'`. An authenticated caller whose current user lacks `deployment_admin` fails with `403` and `error.code='authorization_denied'`; the error details SHOULD identify `required_capability='deployment_admin'` when the route family is denied for that reason.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-343, AC-344, AC-345, AC-346, AC-414, AC-427, AC-432, AC-439, AC-441

**REQ-04-114**
`/api/v1/account/profile` and `/api/v1/account/preferences` MUST authorize against the current authenticated session only. Ordinary authenticated users may read and mutate only their own current-account profile and account-preference resources through those routes. `deployment_admin` is neither required for those routes nor a cross-user bypass through those routes. Incident membership, incident role, provider claim content, visible incident count, and bootstrap-token possession MUST NOT widen current-account profile/preference access. Cross-user account administration, including email or local-login-identifier changes, remains under the deployment-admin user route family.
Profiles: base
Verified by: AC-429, AC-430, AC-431, AC-432

**REQ-04-029**
Holding `deployment_admin` MUST NOT by itself grant incident read, write, preview, download, export, incident-scoped job, or incident WebSocket access. A caller who is `deployment_admin=true` but lacks current membership in incident X MUST have no incident-data or incident-scoped job access to incident X until granted ordinary incident membership.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-261, AC-414, AC-427, AC-439, AC-441

**REQ-04-127**
Inspector feature visibility is not authorization. Every inspector-backed read, route invocation, preview or download handle issuance, job action, mutation, rollback, merge, delete, restore, supersede, mention action, evidence action, related-record creation, and surface pivot MUST re-derive authorization from the caller's current incident membership and role at request time. A hidden, disabled, stale, or cached inspector control MUST NOT be sufficient to authorize or deny a server action. The inspector MUST NOT introduce record-specific ACLs, field-level ACLs, hidden sub-workspaces, generalized approval workflows, or a `deployment_admin` incident-access bypass in the base profile.

The base-profile inspector MUST NOT perform external enrichment and MUST NOT send incident-authored content, evidence bytes, filenames, indicators, hostnames, identities, investigative queries, or other incident-derived values to third-party services. Future external enrichment requires an explicit extension-profile trust boundary, user-visible disclosure, route authorization rule, and audit behavior.
Profiles: base
Verified by: AC-453

**REQ-04-126**
Client-side visibility of the Deployment administration entry is not authorization. An authenticated non-administrator who navigates directly to `/deployment-administration` MUST receive no administrative data and MUST be returned to `/`. If the current session loses `deployment_admin` while `/deployment-administration` is open, the client MUST discard loaded administrative resources, terminate pending administrative requests where possible, and navigate to `/`. A cached page, browser history entry, previously loaded response, or hidden client state MUST NOT preserve administrative access after capability loss.
Profiles: base
Verified by: AC-441

**REQ-04-106**
The Base Profile recovery operator interface is a deployment-local CLI only. It MUST run as a local process and MUST create no network listener. Invocation authority is possession of deployment-local OS execution permission plus access to the effective source deployment configuration, any required target deployment configuration, and required recovery secret references. Browser sessions, incident roles, `deployment_admin`, CSRF tokens, browser `Origin` values, cookies, bearer tokens, common-job authorization, and WebSocket authorization MUST NOT authorize the recovery CLI. CSRF and browser-origin rules are not applicable to the recovery CLI because the interface has no browser, HTTP, or WebSocket surface.

Operator output, progress records, logs, recovery journal records, and administrative-audit summaries for recovery operations MUST NOT contain credentials, secret references, raw DSNs, endpoint hosts, bucket names, object keys, raw storage paths, recovery keys, or incident content. Safe references MAY identify an operation, command, non-secret artifact schema, non-secret logical artifact reference, `backup_set_id`, `consistency_point_at`, result token, error code, or error reason code.
Profiles: base
Verified by: AC-402, AC-427, AC-428

**REQ-04-151**
The Collaboration stream-quarantine requeue interface is a deployment-local
CLI only. Invocation authority is possession of deployment-local OS execution
permission and access to the effective deployment configuration and required
Postgres secret references. Browser sessions, incident roles,
`deployment_admin`, CSRF tokens, browser `Origin`, cookies, bearer tokens,
common-job authorization, and WebSocket authorization neither authorize nor
deny this operation. The interface MUST create no listener, public route,
browser or workbook action, WebSocket action, incident action, or public
administrative-audit projection.

Explicit and discovered configuration paths MUST pass the literal absolute
path and discovery rules in Core 01 REQ-01-655. Configuration and secret
resolution MUST fail closed. The command-specific timeout covers dependency
setup and the semantic transaction. Caller cancellation and timeout MUST
remain distinguishable, and every acquired pool, transaction, row iterator,
or other resource MUST close on every terminal path. Cancellation or timeout
before commit MUST emit no success claim and leave no partial semantic effect.

The raw journal row required by Core 03 REQ-03-307 is operator-private,
append-only administrative evidence. It MUST be written through the existing
raw administrative-audit journal boundary without a corresponding public
projection or new public action-code registry entry. Its safe prior summary is
limited to the presence of quarantine, prior failure count, prior safe reason
code or JSON `null`, and prior quarantine timestamp. Its terminal summary is
limited to result token, mutation timestamp, and requeued-intent count. The row
MAY additionally carry the operation ID and incident ID in their dedicated
identity fields. Omission behavior: it contains no other before/after member.

Collaboration requeue stdout, stderr, logs, typed errors, and raw journal rows
MUST NOT contain credentials, secret references, raw DSNs, endpoint hosts,
database names, SQL, relation or constraint names, stack traces, event
payloads, record content, object keys, storage paths, or upstream error text.
Diagnostic messages are fixed safe text selected by typed code and reason, not
by interpolating a cause. The sole post-commit delivery diagnostic contains
only the operation ID and `result_delivery_failed` as allowed by Core 01.
Profiles: base
Verified by: AC-535

**REQ-04-152**
`operator object-store init` is deployment-local tooling with the same local
OS/configuration authority boundary and negative browser, HTTP, WebSocket,
session, bearer, CSRF, and common-job surfaces as REQ-04-151. Success and
failure output MUST NOT disclose an endpoint host, bucket name, object key,
storage reference, credential, secret reference, raw DSN, raw path, constraint
name, or upstream error text. Only typed configuration diagnostics and typed
Object Store adapter failures may select a specific safe reason. Every unknown
failure uses `dependency_unavailable`; error-message matching is forbidden.
Profiles: base
Verified by: AC-536

**REQ-04-113**
The deployment recovery-operation exclusion boundary is a deployment-local mutual-exclusion boundary over admitted mutating recovery operations. `backup_create`, `restore_latest`, `restore_verify_latest`, and each selected verification inside `restore_verify_due` MUST acquire this boundary before candidate allocation, source backup publication work, target database mutation, or target object-namespace mutation. `backup_inspect_latest` is non-mutating and MUST NOT require this boundary. `restore_verify_due` with no due backups returns `no_op` and need not acquire this boundary. If the boundary is unavailable, the operation MUST fail before mutation with `code='recovery_operation_in_progress'`, `reason_code='operation_lock_unavailable'`, and exit code `3` under REQ-01-595.

Before `restore_latest`, `restore_verify_latest`, or any selected verification inside `restore_verify_due` mutates a target database or target object namespace, the implementation MUST prove all target preflight conditions below:

1. source and target database bindings are distinct;
2. source and target object-store bindings are distinct;
3. the target database is fresh, meaning it contains no application-owned Cartulary data except schema or migration bookkeeping required to admit the operation;
4. the target object namespace is fresh, meaning it contains no object members reachable through the target object-store binding;
5. the operator holds the target's exclusive serving lease, proving that no
   target application HTTP or WebSocket listener is serving and preventing a
   target application process from starting listeners during mutation;
6. the target has a valid `cartulary.restore_target_marker.v2` whose purpose,
   target-generation ID, database-binding digest, object-store-binding digest,
   issuance time, and expiry bind it to this exact admitted target;
7. required recovery keys, backup artifacts, and integrity proofs for the selected backup are available.

Every application server process MUST acquire and hold the shared counterpart
of the serving lease before starting HTTP or WebSocket listeners and MUST
release it only after those listeners are closed. A recovery target admission
MUST acquire and hold the exclusive counterpart before freshness checks and
through every target mutation, validation, rebuild, journal/attestation
decision, and reset. A server startup racing an admitted restore therefore
MUST keep listener startup closed, report `recovery_serving_lease_active`, and
exit `2`. If an application process loses its shared serving lease while
starting or serving, it MUST close readiness and every admission gate
immediately, report fatal condition `recovery_serving_lease_lost`, run the
idempotent fatal drain in REQ-04-146, and exit `70`. It MUST NOT reacquire the
lease or return to serving in-process. Loss of Recovery's exclusive target
lease remains an owner-specific indeterminate mutation outcome: it leaves
readiness false, stops due-verification batching, requires target
reinitialization, and retains the existing Recovery error mapping.

The application-process lease, the application's shared recovery serving
lease, and Recovery's exclusive target lease MUST retain distinct typed
identities, modes, semantic owners, and failure vocabularies. Their
implementations MAY reuse owner-neutral session, advisory-lock, renewal, and
continuity mechanics, but MUST NOT reuse one lock identifier or translate one
owner's failure token into another's.

`restore_latest` markers use `purpose='restore_target'`.
`restore_verify_latest` and each verification admitted by
`restore_verify_due` use `purpose='restore_verification_target'`. The
database- and object-binding digests are SHA-256 over the admitted normalized
non-secret binding identities; they MUST NOT contain or be computed from raw
credentials. The marker lifetime MUST be positive and no greater than 24
hours. A version 1 marker, wrong purpose, wrong generation, wrong binding,
expired marker, missing marker, unavailable exclusive lease, or active shared
lease fails before target mutation using the existing
`unsafe_restore_target` registry: marker defects map to
`target_marker_missing` or `target_marker_invalid`, and serving-lease
contention maps to `target_serving_traffic`.

Preflight failure MUST occur before target mutation and MUST use
`code='unsafe_restore_target'` unless the failure is more specifically one of
`recovery_key_unavailable`, `backup_set_not_found`,
`backup_integrity_failed`, or `recovery_operation_in_progress` under
REQ-01-595. A timed-out, cancelled, or lease-lost restore or restore
verification MUST leave the target not-ready for application traffic, and that
target MUST be reinitialized before reuse.

Every admitted mutating recovery operation MUST append an encrypted
`cartulary.operator_recovery_journal_payload.v2` admission record and terminal
record before its terminal result is considered durable. `backup_create`,
`restore_latest`, `restore_verify_latest`, and each selected verification
inside `restore_verify_due` are mutating recovery operations.
`backup_inspect_latest` is non-mutating. A `restore_verify_due` no-op MUST
append safe admission and terminal evidence for the scheduler invocation but
need not acquire a mutating-operation lock. Journal records are append-only;
terminal state is a new row rather than an overwrite.

When an admitted restore includes the Graph participant, the implementation
MUST propagate the admitted operation ID and the target-generation ID parsed
from the validated v2 target marker. Recovery and Graph MUST NOT mint a second
operation or generation identity. Exclusive serving-lease ownership MUST be
proved continuously through Graph mutation, committed-postcondition
validation, terminal journal publication, readiness aggregation, and any
reset. Loss or uncertainty at any point after Graph mutation begins is an
indeterminate target outcome and requires target reinitialization.

The encrypted terminal evidence format MUST be versioned to retain the Graph
completion tuple and durable participant result from Graph Projection NLSpec
§11.9. Historical v2 journal payloads remain readable only through their
strict historical decoder and MUST NOT be rewritten or interpreted as Graph
completion evidence. Matching current terminal evidence MUST support response
replay without a second Graph mutation. The safe administrative-audit summary
shape below remains unchanged; it MUST NOT expose Graph digests, identifiers
beyond its existing identities, source values, configuration, SQL, database
errors, capabilities, or stack text.

The typed admission record contains exactly schema ID, record kind, operation
ID, operation token, attempt ID, started timestamp, nullable backup ID, nullable
consistency point, and sorted admitted artifact kinds. The typed terminal
record additionally contains completed timestamp, result, sorted artifact
kinds and counts, nullable error code, and nullable error reason. Open maps,
arbitrary keys, exception strings, raw artifact refs, and transport messages
are forbidden.

Once a writable database exists for the operation, the implementation MUST
also write one safe
`cartulary.operator_recovery_audit_summary.v2` derivative containing exactly
operation ID, operation token, attempt ID, result, started and completed
timestamps, nullable backup ID, nullable consistency point, sorted safe
artifact kinds and counts, and nullable error code and reason code. The
terminal encrypted journal record and its administrative-audit derivative MUST
commit in one database transaction or neither may commit. Failure of that
transaction prevents a successful terminal result and maps through the
existing `journal_write_failed` transport reason. Historical journal rows
remain readable forensic evidence and MUST NOT be rewritten. Neither record
may contain the forbidden values in REQ-04-106.
Profiles: base
Verified by: AC-428, AC-534

**REQ-04-030**
Incident membership create, role-change, and delete routes remain incident-scoped authorization decisions. In the base profile, those routes MUST require current incident role `admin`; `deployment_admin` alone MUST NOT bypass that requirement.
Profiles: base
Verified by: AC-054, AC-149, AC-178, AC-179, AC-180, AC-231, AC-427, AC-439

**REQ-04-123**
Administrative audit read authorization MUST fail closed before evaluating query filters, pagination cursor contents, event existence, or event counts. `GET /api/v1/administrative-audit-events` MUST require the current caller to hold `deployment_admin`; an authenticated caller without that capability MUST receive `403` with `error.code='authorization_denied'`. `GET /api/v1/incidents/{incident_id}/membership-audit-events` MUST require current membership in the addressed incident with role `admin`; a caller without current membership or without incident visibility MUST receive the ordinary hidden-incident `404`, and a visible current incident member whose role is not `admin` MUST receive `403` with `error.code='authorization_denied'`. Holding `deployment_admin` without current membership in the addressed incident MUST NOT grant incident membership-audit access and MUST use the same hidden-incident `404` behavior as any other non-member.
Profiles: base
Verified by: AC-438, AC-439

**REQ-04-105**
`GET /api/v1/extensions` is deployment-scoped discovery. Any authenticated session MAY read it. It is not incident-scoped and does not require `deployment_admin`. The route MUST expose only extension-claim state and reserved family roots and MUST NOT expose provider secrets, provider assertions, provider claim maps, or other deployment-local secret-bearing state.
Profiles: base
Verified by: AC-370, AC-371, AC-427

**REQ-04-105A**
When the `network_flow_activity` extension profile is claimed, every public route under `/api/v1/incidents/{incident_id}/network-flow` is incident-scoped incident data. The server MUST rederive authorization at route time from the authenticated session, current incident membership, current incident role, incident lifecycle, and current extension-claim state. `deployment_admin` is not required for these routes and is insufficient by itself to read, query, import, graph, rename, soft-delete, or bind Network Flow resources.

Unclaimed `network_flow_activity` routes MUST fail through Core 01 extension-unavailable behavior before Network Flow route authorization or resource lookup. For a claimed profile, a caller with no current visibility to the addressed incident MUST receive the ordinary hidden-incident result rather than a Network Flow-specific existence signal. A visible incident member whose role is below the route minimum MUST receive `403` with `error.code='authorization_denied'`. Route authorization MUST be evaluated before route-specific query validation, body validation, resource lookup, idempotency replay, cursor replay, graph execution, or indicator target resolution, except where Core 01 explicitly requires extension-claim dispatch before family-specific policy.

The current Network Flow route-family authorization matrix is:

| Operation family | Minimum incident role | Additional rule |
| --- | --- | --- |
| Source-profile discovery | `viewer` | Extension must be claimed. |
| Effective-limit discovery | `viewer` | Extension must be claimed. |
| Table list/read | `viewer` | Tables must be active unless a later owner defines a deleted-table inspection route. |
| Accepted row query | `viewer` | Every table in scope must be active, visible, and in the addressed incident. |
| Rejected-row diagnostic query | `viewer` | Diagnostics are incident content and require current incident visibility. |
| Graph query | `viewer` | Every table in scope must be active, visible, and in the addressed incident. |
| Graph contributor query | `viewer` | The graph selector must be reauthorized and every table in scope must remain active, visible, and in the addressed incident. |
| Import target apply through Core import | `editor` | Core import admission still governs upload/session/apply; Network Flow publication effects require current target-incident authorization at apply time. |
| Table rename | `editor` | The table must be active, visible, in the addressed incident, and at the requested version. |
| Table soft delete | `reviewer` | The table must be active, visible, in the addressed incident, and at the requested version. |
| Indicator link to existing indicator | `editor` | The Core indicator target must be visible in the same incident at initial commit and on exact idempotency replay before target details are returned. |
| Indicator create from flow value | `editor` | Core indicator ownership governs canonical create/dedupe; no Network Flow route may bypass that owner. |

Hidden or cross-incident Network Flow table, row, diagnostic, graph, contributor, cursor, and indicator-binding selectors MUST fail through Core hidden-resource behavior before disclosing Network Flow-local lifecycle, count, cursor, digest, or graph details. Network Flow v1 routes and background work admitted through those routes MUST NOT perform third-party enrichment, reputation lookup, geolocation, passive-DNS, WHOIS, vendor telemetry, LLM, external graph, or other external incident-data egress. A later external-enrichment route requires a new extension trust boundary, user-visible disclosure, route authorization rule, and audit behavior before implementation.
Profiles: network_flow_activity
Verified by: AC-370, AC-371, AC-427

**REQ-04-085**
Only `deployment_admin` may call the deployment-local credential action routes `POST /api/v1/users/{user_id}/password/reset`, `POST /api/v1/users/{user_id}/mfa/totp/reset`, and `POST /api/v1/users/{user_id}/sessions/revoke-all`. Incident membership or incident role `admin` alone MUST NOT authorize those routes.
Profiles: base
Verified by: AC-340..AC-342, AC-427

**REQ-04-094**
Only `deployment_admin` may call `POST /api/v1/users/{user_id}/auth-bindings`, `POST /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}/rotate`, and `DELETE /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}`. Incident membership or incident role `admin` alone MUST NOT authorize those routes.
Profiles: enterprise_authentication
Verified by: AC-352, AC-427

**REQ-04-150**
Every Indicator observation or lifecycle read requires a current authenticated
session and current visibility to the addressed incident. Every create,
resolve, dismiss, restore, or append requires current incident role `editor` or
higher, and cookie-authenticated mutation requires the ordinary CSRF proof.
Deployment administration alone grants no incident visibility or mutation
authority. Authentication, CSRF, path syntax, hidden-resource visibility, and
minimum role are evaluated before route-specific body validation, idempotency
lookup, child lookup, source-field reads, target resolution, or cursor
continuation. A hidden, foreign-incident, deleted, or wrong-type source,
Indicator, observation, interval, or support record MUST NOT disclose its
existence or state.

After current authority succeeds, exact idempotency replay is evaluated before
fresh optimistic concurrency and semantic transition checks. Ordinary callers
cannot select `system` or another origin, cannot supply observed source text or
an origin locator, and cannot supply actor or commit timestamps. Each successful
mutation advances every affected first-class Records envelope exactly once in
`record_id` ascending lock order, appends target and row-centric history,
refreshes affected projections, records the idempotency result, and commits
ordinary record-change Collaboration intents atomically. It MUST NOT publish a
child-specific side channel or expose source text, cursor plaintext, SQL,
constraint names, or hidden identifiers in an error. Any failure leaves all of
those effects absent.
Profiles: base
Verified by: AC-532, AC-533

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
- `redaction_profile_sha256`,
- `output_kind`,
- `output_options`,
- `release_scope`,
- `recipient_partition_refs[]`,
- `graph_projection_refs[]`,
- `composition_id`,
- `composition_version`,
- `composition_sha256`,
- `render_admitted_at`,
- `output_sha256`,
- `redaction_manifest_sha256`.
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

Approval reuse MUST also be invalidated when the selected redaction profile bytes or redaction manifest bytes change, even if the profile id and version strings remain the same.

Publication and invalidation actions in the current Snapshot and Reporting profile require current incident `admin` role on the release's incident. A caller who can see the release but lacks that role receives `authorization_denied`; a caller who lacks release visibility receives `release_not_found`.

**REQ-04-034**
For rendered-output lifecycle, the authoritative artifact state MUST be stored on the release record or an equivalent artifact-scoped record. The closed vocabulary for `release_state` is `pending_approval`, `approved`, `invalidated`, `published`, and `render_failed`.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233, AC-306

For this lifecycle, the logical output slot is the release tuple excluding `output_sha256` and including canonical `recipient_partition_refs[]`.

A release record enters `pending_approval` when bytes and `output_sha256` exist for one bound release tuple but the required approvals are not yet complete. It enters `approved` only when the approval requirements above are satisfied for that exact artifact. It enters `published` only through an explicit publish action after approval. It enters `invalidated` when a different artifact for the same logical output slot supersedes it, when its rendered bytes change, or when the implementation can no longer attest that the required approval set still applies to that exact artifact. It enters `render_failed` only for a failed-closed durable render candidate that has no output bytes and MUST NOT be approved, published, or invalidated.

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
User-account, current-account profile, account-preference, incident-membership, and successful first-deployment-admin bootstrap administration mutations MUST be captured with the same minimum actor, timestamp, source, and before/after fidelity even when they are stored outside incident `change_set` or `record_revisions` rows. For current-account profile/preference mutations, audit and idempotency records MAY record normalized old and new `display_name` or `density_mode` values, `null` density-clearing state, version tokens, route identifiers, and result status, but they MUST NOT reclassify that state as incident content. Deployment-local administrative audit records and any deployment-local idempotency substrate for user-account administration MUST NOT retain `initial_password` or any equivalent secret-bearing bootstrap or auth input in cleartext; they MAY retain only redacted placeholders or non-reversible derived metadata. Those administrative audit records and any current-account route idempotency records are deployment-local state and MUST NOT be serialized into whole-incident portability bundles.
Profiles: base, incident_portability
Verified by: AC-175, AC-231, AC-236, AC-343, AC-344, AC-346, AC-409, AC-430, AC-432, AC-440

**REQ-04-135**
When the Network Flow Activity Extension Profile is claimed, Network Flow domain audit occurrences are incident-scoped immutable audit occurrences, not deployment-local administrative audit records. Each occurrence MUST identify the incident, actor, committed timestamp, Network Flow event code, operation source, and the safe event fields selected by the Network Flow owner. Network Flow domain audit occurrence writes, idempotency-success writes, terminal result publication writes, and Network Flow resource writes that belong to one logical operation MUST participate in one Core unit of work. A conforming implementation MUST NOT commit an idempotency success, table, row, diagnostic, binding, or import terminal result, and MUST NOT return a graph-success response, while the required Network Flow domain audit occurrence for that same operation is missing, duplicated, or outside the transaction boundary. For operations that define no Network Flow domain occurrence, the implementation MUST NOT create a placeholder or synthetic success occurrence merely to satisfy a generic audit hook.
Profiles: network_flow_activity
Verified by: AC-476

**REQ-04-136**
Network Flow domain occurrence counts MUST follow this closed matrix:

| Operation outcome | Required Network Flow domain occurrence |
| --- | --- |
| Import final commit succeeds and creates one or more tables | Exactly one `network_flow_table_created` occurrence per created `network_flow_table`, in terminal result order. |
| Import preview, failed validation, all-invalid unit, no-data unit, cancellation before final commit, worker crash before final commit, or any rolled-back final commit | No Network Flow domain occurrence. |
| Rename commits a changed display name | Exactly one `network_flow_table_renamed` occurrence. |
| Rename request normalizes to the current display name | No Network Flow domain occurrence. |
| Soft delete commits | Exactly one `network_flow_table_soft_deleted` occurrence. |
| Graph query returns a successful graph result | Exactly one `network_flow_graph_query_executed` occurrence. |
| Graph query fails admission, authorization, stale-scope validation, resource-limit validation, Graph Projection execution, cancellation, or any other non-success path | No `network_flow_graph_query_executed` occurrence. |
| Indicator link inserts a new binding | Exactly one `network_flow_indicator_binding_created` occurrence. |
| Indicator link reuses an existing binding under a new `client_txn_id` | Exactly one `network_flow_indicator_binding_reused` occurrence. |
| Exact committed idempotency replay | No new Network Flow domain occurrence; return the original success and original audit correlation when the route response exposes one. |
| Same idempotency key with different normalized digest, stale version, authorization failure, hidden-resource failure, semantic validation failure, limit failure, source-change failure, or any other pre-commit rejection | No Network Flow domain occurrence. |

Profiles: network_flow_activity
Verified by: AC-476

**REQ-04-137**
Network Flow domain audit payloads MUST contain only Core stable identifiers and safe fields. They MUST NOT contain raw display names, raw filenames, raw source bytes, raw CSV cells, raw graph query scalar values, raw indicator candidates, cursor tokens, safe-digest key material, cursor key material, import-source locators, object-store paths, temporary paths, provider details, or Graph Projection provider internals. Safe digests in Network Flow audit payloads are governed by REQ-04-131 through REQ-04-134. `network_flow_graph_query_executed.truncated_example_ref_count` MUST equal the sum over returned graph edges of `example_refs_total_count - length(example_row_refs[])`; when examples are disabled, each returned edge contributes its full `example_refs_total_count`. Failed or over-limit graph queries MUST NOT emit a graph-success audit occurrence with partial counts.
Profiles: network_flow_activity
Verified by: AC-475, AC-476

**REQ-04-138**
Network Flow retry and recovery behavior MUST preserve exact audit occurrence counts. Retrying an operation before final commit MAY retry audit-outbox preparation, but only the transaction that commits the owner state MAY commit the corresponding domain occurrence. A worker crash after owner-state commit but before terminal-result delivery MUST recover and publish the already committed result and occurrence rather than rerunning the owner mutation or appending another occurrence. Exact idempotency replay MUST be satisfied from committed idempotency state before any new mutation or domain occurrence append is attempted. A failure while appending the required domain occurrence, committing idempotency success, or publishing the outbox item MUST fail closed or recover by the same committed transaction; it MUST NOT leave a visible owner resource without its required domain occurrence or a domain occurrence without its owner resource.
Profiles: network_flow_activity
Verified by: AC-476

**REQ-04-086**
Password change, password reset, TOTP begin, TOTP complete, TOTP reset, and explicit revoke-all session actions MUST be auditable deployment-local administrative events. Deployment-local audit or idempotency state for these routes MUST NOT retain `current_password`, `new_password`, `secret_base32`, `otpauth_uri`, or raw `bootstrap_token` in cleartext.
Profiles: base
Verified by: AC-336..AC-338, AC-340..AC-342, AC-440

**REQ-04-096**
Enterprise-auth binding create, rotate, and retire actions MUST be auditable deployment-local administrative events. Those audit records MUST preserve the target stable `user_id`, `provider_key`, and any old or new `provider_subject` values needed to reconstruct binding lifecycle, but they MUST NOT retain any member of this closed forbidden set: provider assertions, provider tokens, raw SAML responses, ID tokens, and access tokens.
Profiles: enterprise_authentication
Verified by: AC-352, AC-440

**REQ-04-124**
Neither stored administrative audit values nor values returned by the administrative audit read projections in Core 01 §3.3.5.1A may contain any member of this forbidden set: current passwords, new passwords, initial passwords, reset passwords, password hashes, TOTP secrets, `otpauth_uri`, bootstrap tokens, session tokens, provider assertions, provider tokens, raw SAML responses, ID tokens, access tokens, recovery keys, object-store credentials, raw DSNs, object keys, and storage secrets. When an audited field would otherwise contain a forbidden value, the audit event MUST preserve the existence of the changed field with `value_state='redacted'` and JSON `null` for both `before` and `after` under Core 01 §3.3.5.1A. Implementations MUST NOT satisfy this requirement by dropping the entire audit event, substituting masked substrings of the forbidden value, hashing the forbidden value into a returned audit value, or moving the forbidden value into an auxiliary audit field.
Profiles: base, enterprise_authentication
Verified by: AC-437, AC-440

**REQ-04-125**
Administrative audit events are immutable once committed. The current profile exposes no public delete, purge, compact, rewrite, or retention-shortening route for administrative audit events. Administrative audit retention lasts for the deployment lifetime, is included in operational backup, and remains excluded from whole-incident portability. Historical raw journal records remain available to authorized recovery and forensic workflows, but a record without an exact current action, scope, target, and nonempty secret-safe change projection MUST NOT be exposed through a public administrative audit projection.
Profiles: base
Verified by: AC-437, AC-439, AC-440

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
- future reenactment outputs, if introduced by a later adopted profile, MUST be marked `generated_presentation=true` and MUST NOT be released as `external_release` unless that later profile explicitly admits the publication boundary,
- approval and redaction checks MUST complete successfully before an `external_release` artifact is published.
Profiles: snapshot_reporting, incident_portability
Verified by: AC-031, AC-057, AC-059, AC-060, AC-061, AC-062, AC-091, AC-113, AC-114, AC-115, AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-233, AC-236, AC-333

**REQ-04-046**
In that profile, the implementation MUST support generating multiple recipient-specific artifacts from the same immutable snapshot by selecting different canonical `recipient_partition_refs[]`, versioned redaction profiles, and, when needed, different templates. If an incident involves multiple affected parties, an artifact prepared with one recipient-specific configuration MUST NOT disclose content whose `disclosure_partition_refs[]` are not allowed by the effective selected redaction profile. Manual post-render editing MAY still occur, but it MUST NOT be required for the implementation's supported recipient-specific configurations.
Profiles: snapshot_reporting, incident_portability
Verified by: AC-031, AC-057, AC-059, AC-060, AC-061, AC-062, AC-091, AC-113, AC-114, AC-115, AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-233, AC-236

**REQ-04-046a**
For the Snapshot and Reporting Extension Profile, Core 04 MUST provide the security controls consumed by the Reporting and Report Composition NLSpecs with the following defaults and omission behavior:

| Control | Default when omitted | Required boundary |
| --- | --- | --- |
| `allow_authored_presentation_text` | `false` | External releases reject composition-authored presentation text unless the selected redaction profile explicitly enables it. |
| Public Party display labels | Not permitted | A Party label is public only when Core source state marks it public-directory eligible and the selected redaction profile explicitly permits public Party labels. |
| Superseded-record external release | Excluded | A superseded record appears in external release only when both template opt-in and redaction-profile permission are present. |
| Recipient partition validation | Fail closed | `recipient_partition_refs[]` must resolve to snapshot Parties and exactly match the selected redaction profile's allowed `party:*` subset. |
| Reveal-map access and retention | Sensitive release artifact | Reveal maps require release-artifact authorization and retention controls; they MUST NOT be exposed through ordinary workbook reads. |
| Render sandbox | Non-egress | Rendering must run inside a boundary that blocks non-loopback network egress, remote package resolution, undeclared filesystem reads, arbitrary environment access, and remote runtime assets. |

These controls are release-time and render-time security boundaries. They MUST NOT alter live workbook row visibility, field visibility, evidence visibility, search results, saved views, or incident membership authorization.
Profiles: snapshot_reporting
Verified by: AC-233

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
- a backup or restore mechanism,
- telemetry exporter or receiver configuration,
- telemetry exporter headers or secret references,
- telemetry retained artifacts or diagnostics,
- telemetry redaction or attribute-governance rules,
- browser diagnostics or browser telemetry-boundary behavior,
- telemetry runtime-failure behavior or runtime-invariance evidence.
Profiles: base
Verified by: AC-048, AC-231

**REQ-04-051**
At minimum, the threat model MUST cover the following assets and abuse cases:
Profiles: base
Verified by: AC-048, AC-231

| STRIDE class | Minimum project-specific scope | Required control direction |
| --- | --- | --- |
| Spoofing | analyst sessions, provider-backed identities, explicit system-process actors, object-store upload/download capabilities, SeaweedFS S3 endpoint identity, reverse-proxy trust boundary, direct-upload target scope, telemetry exporter endpoint identity, same-origin browser diagnostics | authenticated sessions, stable internal user mapping, explicit system actors, short-lived operation-scoped object access, configured object-store endpoints only, direct-upload targets bound to one pending blob slot, explicit configured telemetry endpoints only, no environment-driven telemetry egress, no browser direct export |
| Tampering | incident records, revisions, object blobs, object overwrite or delete attempts, object metadata drift, backup manifests, backup artifacts, restore-verification extracts, reference packs, snapshots, exports, telemetry config, generated telemetry constants, telemetry source snapshot, golden telemetry corpus, retained telemetry artifacts | row-versioned writes, immutable change sets, write-once product evidence keys, blob hashes, SHA-256 backup proofs, fail-closed integrity verification, immutable published snapshots, source-snapshot validation, generated-constant drift checks, artifact integrity checks |
| Repudiation | edits, imports, rollbacks, pack activation, export generation, evidence lifecycle actions, direct-upload issuance and attach finalization, object-store dependency failures, migration copy and validation events, backup and restore verification, telemetry config changes, exporter failures | attributed append-only history with actor, timestamp, source, and reversible mutation detail, application attach audit as authoritative, object-store logs as diagnostics only, attributed deployment-config changes, retained migration and backup/restore summaries, retained telemetry conformance summaries |
| Information disclosure | evidence blobs, direct-upload targets, same-origin evidence handles, raw object keys, storage refs, bucket names, SeaweedFS admin, filer, master, volume, WebDAV, and debug surfaces, backup artifacts, restore-verification extracts, migration ledgers, exports, previews, secrets, portable runtime roots, exporter headers, incident-derived telemetry, local diagnostics, raw telemetry captures | incident-scoped authorization, same-origin preview/download handles, no raw backend identifiers in public responses, secret isolation, untrusted-content rendering rules, self-contained outputs, encrypted flyaway storage, operator-private artifact redaction, forbidden-value tests, secret redaction, admin/debug surface non-exposure by default, raw capture retention outside committed source |
| Denial of service | oversized evidence, archive bombs, pathological imports, expensive report or preview jobs, object-store prefix listing abuse, storage exhaustion, repeated range reads, startup probe cleanup failures, exporter retry loops, processor queues, telemetry self-diagnostics | size and decompression limits, background-job isolation, cancellation, bounded hot-path retrieval, no user-facing object-store listing primitive, storage quotas or operational capacity controls, bounded range behavior, deterministic probe cleanup, bounded telemetry queues, retry cutoff, non-blocking hot path, recursion guard |
| Elevation of privilege | user-controlled record or blob identifiers, destructive operations, job-worker storage access, wildcard object-store credentials, anonymous bucket access, exposed SeaweedFS admin APIs, wildcard CORS, default-service confusion, telemetry test hooks, browser telemetry configuration, environment autoconfiguration | server-side authorization derived from object ownership, role gates for destructive actions, least-privilege worker credentials, no anonymous default object-store access, no wildcard bucket credentials for product runtime, restrictive upload CORS, no admin API exposure in ordinary deployments, explicit service references, no runtime deterministic telemetry test hook, no browser telemetry config, environment containment |

When an adopted OpenTelemetry subsystem is present, the telemetry portion of this threat model also uses these explicit verification hooks. These hooks are implementation-support evidence for the telemetry boundary; they do not by themselves create a Core 05 publication claim.

| Telemetry threat surface | STRIDE coverage | Required control direction | Verification hook |
| --- | --- | --- | --- |
| Exporter endpoint configuration and no-default egress | Spoofing, elevation of privilege, denial of service | only explicit valid `telemetry.exporter.*` configuration may enable export; OTel environment variables and declarative config cannot create egress; retry is bounded and non-blocking | `contracts/otel/telemetry_config_schema.v2.json`, `contracts/otel/config_hazard_fixture_matrix.v2.json`, `internal/platform/config/config_otel_test.go`, `internal/platform/telemetry/exporter_test.go`, `make otel-conformance` |
| Exporter headers and secret references | Information disclosure, tampering | exporter headers use Core 04 `secret_ref_v1`; raw values, protocol-owned overrides, oversized headers, and leaked diagnostics fail closed before readiness | `internal/platform/config/config_otel_test.go`, `contracts/otel/telemetry_config_schema.v2.json`, `make otel-conformance` |
| Source snapshots, generated constants, normalized goldens, and retained raw captures | Tampering, repudiation, information disclosure | source and generated-constant provenance are pinned; normalized goldens remain committed; raw captures stay under the harness run root and outside committed golden directories | `contracts/otel/otel_source_snapshot.v1.json`, `contracts/otel/generated_constants_manifest.json`, `internal/testutil/golden/otel/corpus_manifest.json`, `make otel-conformance` |
| Redaction-before-recording and attribute governance | Information disclosure, tampering | forbidden value families are omitted, replaced by closed classes, or dropped before OTel calls; SDK limits are not privacy controls | `internal/platform/telemetry/privacy_test.go`, `contracts/otel/error_class_registry.json`, `internal/testutil/golden/otel/cases/OTEL-CORPUS-013`, `internal/testutil/golden/otel/cases/OTEL-CORPUS-018`, `make otel-conformance` |
| Browser telemetry boundary | Spoofing, elevation of privilege, information disclosure | browser packages and browser state cannot configure exporters, headers, endpoints, samplers, processors, metric readers, resource attributes, log bridges, or direct telemetry export | `contracts/otel/import_boundary.json`, `make build-web`, `make otel-conformance` |
| Runtime failure invariance | Denial of service, repudiation | exporter failure, timeout, queue overflow, shutdown, redaction rejection, log mapping failure, and self-diagnostic failure must not change product responses or committed state | `internal/testutil/golden/otel/cases/OTEL-CORPUS-017`, `internal/platform/telemetry/exporter_test.go`, `make otel-conformance` |

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
- **CWE-639**: Every mutation, preview or download handle issuance, preview or download handle redemption, and object-store URL issuance MUST re-derive authorization server-side from the target object's owning incident and the caller's current membership and role. View-scoped batch mutations that accept client-supplied `record_id` targets MUST validate each record target against the route's addressed incident before any row-version or conflict evaluation. Client-supplied incident identifiers, ownership metadata, or role claims MUST NOT determine access.
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
- one SeaweedFS S3-compatible object-store container or equivalent S3-compatible object store.
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

**REQ-04-112**
An on-prem stand-up package MAY provide one application image plus companion local Postgres and SeaweedFS S3-compatible object-store services as package scaffolding. The application image is the only application deployable; migration and operator tooling in that image are deployment-local operational tooling, not additional application deployables. Such a package MUST use the deployment-configuration contract and service-ref bindings defined in Core 04 §12, MUST serve packaged browser assets without a development frontend runtime, and MUST NOT be represented as disconnected-profile conformance unless the disconnected deployment profile and its acceptance criteria separately pass.
Profiles: base
Verified by: AC-404

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

**REQ-04-157**
Timed and fixture-sensitive implementation conformance MUST use the following
closed performance fixtures. `cartulary.perf.large_grid.v1` contains exactly
20,000 Timeline rows, 1,000 Host rows, 1,000 Identity rows, and 1,000 each of
deterministic tag, mention, and link associations distributed across every
twentieth Timeline row. It uses seed `20260405`, the default Timeline sort,
filter, and grouping state, ordinary security controls, and 25 authenticated
analyst sessions on one incident with presence enabled. One foreground analyst
performs the measured interaction; the other 24 sessions each commit one update
to a non-target Timeline row every five seconds, evenly staggered, for a steady
aggregate rate of 4.8 committed updates per second. Target rows MUST be excluded
from the background-update pool.

The fixture's semantic identity is defined by its owner contribution identities,
receipts, exact semantic validation, source contracts, fixture version, and seed.
An implementation MAY change an owner-internal construction batch size or
persistence strategy only when the resulting semantic identity, exact counts,
conditions, contribution order, and supported product behavior remain unchanged.
Construction timing, batch counts, batch sizes, and other build diagnostics are
non-semantic: they MUST NOT enter a contribution receipt, snapshot key, or
semantic validation digest.

`cartulary.perf.evidence_heavy.v1` contains 5,000 Timeline rows, 10,000 Evidence
records, tens of gigabytes of binary evidence in object storage, and at least
one Timeline row linked to 100 Evidence records. Throughput tests MAY stub blob
bytes, but Evidence metadata, counts, attachment state, and preview handles MUST
be real. These fixtures define product-supported shape and load independently
of any Core 05 publication environment.
Profiles: base
Verified by: AC-043, AC-044, AC-045, AC-047

**REQ-04-158**
The current implementation profile owns the following closed
`measurement_predicate_id` registry. Each predicate starts when the product
event handler accepts the named user action, not when a test driver dispatches
it. A visible stop state is satisfied only when it holds on two consecutive
animation frames, the target intersects the clipped grid viewport, and the
target remains rendered, visible, and anchored on the second frame.

| `measurement_predicate_id` | Bound criterion or criteria | Start state and initiating action | Stop predicate | Fixture and sampling |
| --- | --- | --- | --- | --- |
| `perf.timeline_paste_20x5.v1` | `AC-003` | Timeline surface loaded with default sort, filter, and grouping state; target range visible; the paste commit is accepted. | 20 committed rows are visibly painted with mapped values in five writable columns and stable `record_id` plus integer `row_version` binding. | `cartulary.perf.large_grid.v1`; single observation |
| `perf.presence_delta.rendered.v1` | `AC-008`, `AC-132` | Analyst A commits a workbook-surface, focused-row, or same-cell edit-state presence change. | Analyst B visibly renders the corresponding indicator from matching `sheet_ref`, `record_id`, and `field_key`. | `cartulary.perf.large_grid.v1`; single observation |
| `perf.rollback_or_row_restore.rendered.v1` | `AC-011` | A reviewer confirms rollback or whole-row restore. | The visible row reflects the new attributed revision and history shows the new entry. | `cartulary.perf.large_grid.v1`; single observation |
| `perf.job_progress.visible_with_cancel.v1` | `AC-016`, `AC-027`, `AC-030`, `AC-033`, `AC-046` | A user submits an owner-defined background action. | Visible progress and cancel affordances render while another grid row remains selectable and accepts text input without modal capture. | The criterion-bound Core 04 fixture; single observation |
| `perf.timeline_summary_selection_down.v1` | `AC-043` | `cartulary.view.timeline.v2` is loaded in its default state; an existing visible `timeline.activity_synopsis_text` cell is selected; keyboard targeting is on that cell; ArrowDown is accepted. | The next deterministic Timeline row is visibly selected and keyboard targeting follows the same field on that row. | `cartulary.perf.large_grid.v1`; p95 |
| `perf.timeline_summary_focus_edit.v1` | `AC-043` | The same exact Timeline summary selection state holds; Enter is accepted. | The editor for the same `record_id` and `field_key` is visibly active and focused with a collapsed caret at the rendered text end. | `cartulary.perf.large_grid.v1`; p95 |
| `perf.typing_ack.v1` | `AC-043` | An existing committed Timeline row is in active edit mode for `timeline.activity_synopsis_text`; the same row remains selected; the editor is focused; the caret is collapsed at the rendered text end; there is no selection and IME composition is inactive; the literal ASCII character `x` is accepted. | The same editor visibly renders the prior text plus one trailing `x` and remains anchored to the same `record_id` and field. Saved/sync state, badges, DOM mutation, network completion, collaboration messages, or `row_version` changes do not satisfy this predicate. | `cartulary.perf.large_grid.v1`; p95 |
| `perf.timeline_blank_row_create.v1` | `AC-043` | Timeline is loaded in its default state; a visible blank row contains one qualifying non-empty summary value; Enter commit is accepted. | The committed row is visibly painted with stable `record_id`, integer `row_version`, and the entered value in the target field. | `cartulary.perf.large_grid.v1`; p95 |
| `perf.view_change.first_useful_viewport.v1` | `AC-044` | The active surface is loaded in its default state; a sort, filter, or grouping change is accepted. | The Core 04 first useful viewport is visible. | `cartulary.perf.large_grid.v1`; p95 |
| `perf.view_change.stable_viewport.v1` | `AC-044` | The active surface is loaded in its default state; a sort, filter, or grouping change is accepted. | The Core 04 stable viewport is visible. | `cartulary.perf.large_grid.v1`; p95 |
| `perf.evidence_inspector.metadata_shell.v1` | `AC-045` | A user opens the inspector on a Timeline row linked to 100 Evidence records. | The selected-row summary, total count, and first Evidence-list window are visible with label, attachment state, and preview-handle availability for every rendered item. | `cartulary.perf.evidence_heavy.v1`; p95 |
| `perf.anchor_stability_under_live_updates.v1` | `AC-047` | The deterministic live-update trace begins while an analyst holds an edit anchored to one `record_id`. | The edit remains on that record throughout scrolling, sorting, filtering, grouping, and live updates, and viewport stabilization avoids a full-sheet rerender. | `cartulary.perf.large_grid.v1`; seeded pass/fail scenario |

One `measurement_predicate_id` MUST NOT denote interchangeable editor,
field, or harness realizations. A changed anchor, editor family, action, or stop
predicate requires a new ID. The retired generic IDs
`perf.selection_change.v1` and `perf.focus_change.v1` MUST NOT satisfy current
implementation conformance.
Profiles: base
Verified by: AC-003, AC-008, AC-011, AC-016, AC-027, AC-030, AC-033, AC-043, AC-044, AC-045, AC-046, AC-047, AC-132

**REQ-04-159**
Every p95 implementation predicate uses one discarded warm-up operation and
exactly 100 measured operations. Samples MUST be finite and non-negative; p95
is the nearest-rank sample at zero-based index `ceil(0.95 * N) - 1` after
ascending sort. Product or harness retries, percentile slack, estimator
substitution, and threshold relaxation are non-conformant. Test-driver dispatch
latency MAY be retained as a diagnostic stage but MUST NOT enter the product
threshold interval. The exact stages, when applicable, are driver dispatch,
accepted-action-to-request, request round trip, response decode, client apply,
apply-to-visible-paint, and total.
Profiles: base
Verified by: AC-043, AC-044, AC-045

**REQ-04-160**
Timeline implementation conformance MUST prove all of the following as one
owner-routed family: one immutable mutation-policy vocabulary; consumer-owned
least-capability interfaces and fixture-only bulk authority; one canonical
value and hash encoding vocabulary; one typed collection-fact reader used by
separate composition roots; one current version-identifier writer grammar;
and complete removal and verification accounting for superseded artifacts.

Evidence MUST cover exact boundary values, byte-level persisted and published
representations, caller-transaction identity, failure precedence, nil and
duplicate behavior, authorization denials without side effects, portability
when that profile is claimed, static ownership boundaries, and the authored
test catalog and generated harness topology. Conformance evidence MUST NOT
read, stat, hash, parse, or otherwise depend on Markdown or handoff files at
runtime. These rules introduce no public HTTP, OpenAPI, WebSocket, view-schema,
frontend, projection-storage, authorization, or database-schema change.
Profiles: base
Verified by: AC-545, AC-546, AC-547, AC-548, AC-549, AC-551

### 9.0 Profile claim manifests

The manifests below define implementation claim boundaries without restating requirement prose. Each manifest selects implementation requirements through the `Profiles:` trailers carried by Core 00 through Core 04 and pairs that selector with the acceptance criteria that complete the claim. Appendix F expands every selector into explicit navigation tables.

#### 9.0.1 Base claim manifest

A Base claim selects every requirement block tagged `base`.

Definition of Done:

- requirement selector: `profile:base`
- required acceptance criteria: `AC-001..AC-026`, `AC-037..AC-055`, `AC-068..AC-070`, `AC-072..AC-090`, `AC-097..AC-103`, `AC-107..AC-112`, `AC-116..AC-163`, `AC-170..AC-231`, `AC-238..AC-261`, `AC-277..AC-287`, `AC-294..AC-304`, `AC-311..AC-322`, `AC-329..AC-331`, `AC-334..AC-347`, `AC-353..AC-354`, `AC-359..AC-368`, `AC-370..AC-371`, `AC-372..AC-375`, `AC-376..AC-385`, `AC-387..AC-392`, `AC-394..AC-408`, `AC-410`, `AC-411`, `AC-412`, `AC-413`, `AC-414`, `AC-415`, `AC-416`, `AC-417`, `AC-418..AC-432`, `AC-437..AC-441`, `AC-444..AC-462`, `AC-469..AC-474`, `AC-480..AC-486`, `AC-545..AC-549`, `AC-551`, `AC-554..AC-556`, `AC-558..AC-560`
- **AC-231**: A Base claim is conformant only when every requirement selected by `profile:base` is implemented and every acceptance criterion listed in this manifest passes.
  - Verifies: `profile:base`

AC-231 through AC-236 are aggregate profile-claim gates, and PC-006 is a companion claim-publication gate. They MUST NOT be the sole verifier for any requirement that specifies substantive runtime behavior. Every such requirement family MUST also be covered by at least one non-aggregate acceptance criterion or companion criterion that exercises that behavior directly.

#### 9.0.2 Import claim manifest

An Import claim requires a passing Base claim and every requirement block tagged `import`, including the import-boundary and import-provenance requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:import`
- additional acceptance criteria: `AC-027..AC-029`, `AC-063..AC-067`, `AC-232`, `AC-262..AC-265`, `AC-323..AC-325`, `AC-393`, `AC-463..AC-468`
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
- additional acceptance criteria: `AC-033..AC-035`, `AC-092..AC-096`, `AC-234`, `AC-270..AC-272`, `AC-308..AC-310`, `AC-326`, `AC-369`, `AC-443`
- **AC-234**: A Reference Pack claim is conformant only when a Base claim passes, every requirement selected by `profile:reference_pack` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:reference_pack`

#### 9.0.5 Enterprise Authentication claim manifest

An Enterprise Authentication claim requires a passing Base claim and every requirement block tagged `enterprise_authentication`, including the provider-identity requirements in §1.2 and the provider-manifest deployment-configuration requirements in §12.3.4.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:enterprise_authentication`
- additional acceptance criteria: `AC-036`, `AC-235`, `AC-288..AC-293`, `AC-348..AC-352`, `AC-433..AC-436`
- **AC-235**: An Enterprise Authentication claim is conformant only when a Base claim passes, every requirement selected by `profile:enterprise_authentication` is implemented, and every additional acceptance criterion listed in this manifest passes.
  - Verifies: `profile:enterprise_authentication`

#### 9.0.6 Incident Portability claim manifest

An Incident Portability claim requires a passing Base claim and every requirement block tagged `incident_portability`, including the logical-bundle and import-failure requirements outside §9.

Definition of Done:

- prerequisite claim: Base
- additional requirement selector: `profile:incident_portability`
- additional acceptance criteria: `AC-164..AC-169`, `AC-236`,
  `AC-273..AC-276`, `AC-327..AC-328`, `AC-332`, `AC-386`, `AC-409`,
  `AC-440`, `AC-442`, `AC-487..AC-508`, `AC-550`, `AC-557`
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
- **AC-013**: Re-sorting, re-filtering, or re-grouping a sheet does not cause a pending edit to target a different underlying record; stale full-query results, stale live row patches, stale mutation replay responses, and stale action responses do not lower the rendered `row_version` or committed cell state for a locally accepted newer `record_id`; all mutations are sent using `record_id`, `base_row_version`, and changed fields only.
  - Verifies: REQ-01-350, REQ-03-033..REQ-03-035, REQ-03-086, REQ-03-223..REQ-03-224
- **AC-014**: Renaming a visible column header or tab label does not change filter semantics, write-back behavior, or export semantics for a built-in or system view; those behaviors are bound to `view_schema_id`.
  - Verifies: REQ-03-223..REQ-03-224

- **AC-116**: `GET /api/v1/view-schemas` returns only standardized current-profile `view_schema_id` values in `data.view_schemas[]`, ordered by `view_schema_id asc`. Each list entry and each `GET /api/v1/view-schemas/{view_schema_id}` singleton response is a `view_schema_resource_v2` containing required members `view_schema_id`, `surface_kind`, `title`, `source_record_types`, `technical_fields`, `required_reference_pack_keys`, `default_sort`, `sort_fields`, `filter_fields`, `synthetic_filter_predicates`, `grouping_fields`, and `fields`. `technical_fields` is exactly `["record_id","row_version"]`. The public discovery resource does not expose `base_projection`, storage-table names, or internal write targets. The required pack-independent floor is exactly these fourteen ids: `cartulary.view.assessments.v1`, `cartulary.view.comm_log.v1`, `cartulary.view.decisions.v1`, `cartulary.view.evidence.v1`, `cartulary.view.handoff.v1`, `cartulary.view.hosts.v1`, `cartulary.view.identities.v1`, `cartulary.view.indicators.v1`, `cartulary.view.lesson.v1`, `cartulary.view.notes.v1`, `cartulary.view.parties.v1`, `cartulary.view.status_review.v1`, `cartulary.view.task_requests.v1`, and `cartulary.view.timeline.v2`. The current implementation also exposes the implemented standardized optional artifact-backed workbook surfaces `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1`; those three surfaces are current-profile standardized workbook surfaces with `required_reference_pack_keys=[]`, not optional reference-pack overlays. Discovery MUST NOT expose any other optional or pack-dependent `view_schema_id` values unless a later owner section defines them as current-profile standardized workbook surfaces. ATT&CK, D3FEND, VERIS, and other framework-specific overlay `view_schema_id` values do not appear in the current profile. Omitting `limit` yields `meta.paging.limit=100`; terminal pages use `meta.paging.has_more=false` and `meta.paging.next_cursor=null`; invalid `limit` values or aliases such as `page`, `offset`, `page_size`, and `block_size` fail closed with `400 error.code='invalid_pagination_request'`; cursor replay against a different bound route contract fails closed; and `GET /api/v1/view-schemas/{view_schema_id}` rejects pagination members with `error.details.reason_code='pagination_not_supported'`.
  - Verifies: REQ-00-014, REQ-01-240..REQ-01-242, REQ-01-285..REQ-01-296, REQ-01-307..REQ-01-310, REQ-01-499, REQ-01-503..REQ-01-509, REQ-03-004, REQ-03-242..REQ-03-246

- **AC-410**: Core 02 §10.4.4A defines a closed tagged-variant registry for artifact-backed notes, structured coordination artifacts, and structured findings. The registry contains exactly six rows in canonical order `note`, `comm_log`, `handoff`, `status_review`, `lesson`, and `finding`. Each row declares exact `durable_discriminator`, `public_surface_ref`, `identifier_field`, `required_structured_state[]`, `optional_structured_state[]`, and `lifecycle_notes`. The `note` and `finding` rows use `identifier_field=record_id` rather than subtype-local identifiers. `finding` is the only row with `subkind_dimension='finding.kind'` and it explicitly carries the current-profile hypothesis representation. There is no separate `hypothesis` row and no `cartulary.view.hypotheses.v1`. Core 01 §7.4 and Core 03 §2 contain explicit cross-references that preserve Core 01 ownership of exhaustive field contracts and create-time policy while using the Core 02 registry as the closed family inventory.
  - Verifies: REQ-01-309, REQ-02-134, REQ-02-250..REQ-02-253, REQ-03-004, REQ-03-011

- **AC-411**: Core 01 §7.4 contains exactly one authoritative cross-layer workbook-surface mapping table for the current profile. That table contains exactly seventeen rows in canonical order: the fourteen pack-independent registry entries from REQ-01-307 followed by `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1` in the order named by REQ-01-308. Each row declares exact `view_schema_id`, `surface_kind`, `source_record_types`, canonical source discriminator or filter, `surface_status`, and `required_reference_pack_keys`; every row binds canonical workbook identity to `sheet_ref={ "kind": "view_schema", "id": <view_schema_id> }`; `surface_kind` uses only `built_in_sheet` or `system_view`; the table introduces no pack-dependent framework overlays and no new runtime discovery members; and Core 03 §2 contains a non-owner cross-reference back to that table rather than a duplicate row-mapping table.
  - Verifies: REQ-01-579, REQ-03-011

- **AC-117**: The structured contract for each of those fourteen mandatory base-profile view schemas, plus any currently exposed standardized optional artifact-backed workbook surface, includes an ordered `fields[]` list, an ordered `default_sort` tuple, explicit `sort_fields`, `filter_fields`, `synthetic_filter_predicates[]`, and the declared `grouping_fields` whitelist when grouping is supported; the final default-sort tiebreaker is `record_id`. `fields[]` are emitted in authoritative field-registry order. Set-like members, including filter `values[]` and canonicalized `full_text` token sets rendered through canonical `arg.query`, use canonical ascending order. `filter_fields` contains only keys also present in `fields[].field_key`, and filter-only synthetic predicate keys appear only in `synthetic_filter_predicates[]`; for Notes, `note.full_text` is exposed as a synthetic filter predicate rather than as a field entry.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-310, REQ-01-499..REQ-01-500, REQ-01-503..REQ-01-509, REQ-03-242..REQ-03-246

- **AC-118**: Every `fields[]` entry in those fourteen mandatory base-profile view schemas, plus any currently exposed standardized optional artifact-backed workbook surface, is `view_field_entry_v2` with required members `field_key`, `label`, `default_hidden`, `sortable`, `header_sort_field_key`, `filter_ops`, `groupable`, `read_kind`, `write_kind`, `grid_editable`, `conflict_resolution_class`, `entity_binding_mode`, `string_contract_id`, `direct_scalar_contract_id`, `direct_reference_contract_id`, `clearable`, and `enum_values`; inapplicable contract members are explicit `null`; `filter_ops=[]` means not filterable; and closed-vocabulary fields expose `enum_values` in declared token order. Every writable field declares a stable `field_key` and `conflict_resolution_class`; every entity-bearing writable field also declares `entity_binding_mode`; every human-authored writable string field or writable string-bearing action member closed by Core 01 §18 declares the correct `string_contract_id` binding; every writable direct temporal scalar field closed by Core 01 §18A declares the correct `direct_scalar_contract_id` binding and explicit clearable flag; every writable direct user reference binds `incident_member_user_ref_v1`; required scalar fields reject explicit `null` and any contract-defined normalized-empty input; clearable optional scalar text fields and clearable optional direct temporal scalar fields apply omission-versus-`null` semantics exactly as declared; `collection_review` fields reject raw arrays and raw `null` on patch in favor of `collection_actions_v1`; and create-only identifiers remain immutable after first commit. Changing only a surface `title` or field `label` is non-breaking, but changing field membership, field meaning, grid editability, `conflict_resolution_class`, `entity_binding_mode`, `sort_fields`, `filter_fields`, `synthetic_filter_predicates`, or `grouping_fields` without a new `view_schema_id` is non-conformant.
  - Verifies: REQ-00-014, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-310, REQ-01-323..REQ-01-341, REQ-01-487..REQ-01-509, REQ-02-028..REQ-02-029, REQ-02-202..REQ-02-204, REQ-03-052..REQ-03-053, REQ-03-242..REQ-03-246
- **AC-119**: In the Timeline v2 schema, each of the ten visible operational text fields from Core 01 §7.4.1 has a distinct stable `field_key` and distinct structured source-state write target. Editing one visible field does not overwrite any other visible field, inspector-side suggestion/link state, evidence links, tags, or capture metadata.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-322, REQ-03-236..REQ-03-241
- **AC-120**: A write routed to a field that the active base-profile view schema declares read-only or derived fails closed, does not mutate authoritative source state, and does not persist a misleading projection update.
  - Verifies: REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-03-236..REQ-03-241

- **AC-191**: A Timeline row created by blank-row entry, by a paste action that creates a new row, or by a screenshot-only create persists `capture_state='rough'` on first commit; `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` for `cartulary.view.timeline.v2` accepts a request whose only top-level member is `client_txn_id`, returns `201 Created` on first success, and creates exactly one Timeline row whose ten visible v2 cells default to `null` and remain eligible for later screenshot or rough-capture enrichment.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241
- **AC-192**: A client attempt to supply `timeline.capture_state` in Timeline row creation or in `PATCH /api/v1/records/{record_id}`, or to supply any other create-forbidden or unknown top-level member on Timeline row create, or to send a non-object, duplicate-member, trailing-JSON, or otherwise malformed Timeline mutation body, fails with `400 error.code='invalid_mutation_payload'`, does not mutate authoritative source state, leaves no partial row, and does not persist a misleading projection update.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110, REQ-03-236..REQ-03-241
- **AC-193**: The first later `capture-state-material` mutation to a `rough` Timeline row, including an edit to any of the ten Timeline v2 visible operational fields, an evidence attach or detach, or a row-anchored MITRE, entity, or indicator mutation, commits one visible `change_set` that also sets `capture_state='enriched'`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241
- **AC-194**: `POST /api/v1/records/{record_id}/mark-reviewed` by a `reviewer` or `admin` against a non-deleted Timeline row whose current `capture_state` is `rough` or `enriched` returns `200 OK`, increments `row_version`, appends a new `change_set`, and sets `capture_state='reviewed'`; when optional `reason` is supplied it is normalized using `reason_note_v1`, and omission, explicit `null`, and normalized-empty reason compare equal and persist as `null`; the same route by an `editor` fails with `403`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-01-487..REQ-01-488, REQ-01-496, REQ-03-102..REQ-03-110
- **AC-195**: A later `capture-state-material` mutation to a `reviewed` Timeline row commits one visible `change_set` that sets `capture_state='enriched'`; a tag-only change to the same row leaves `capture_state='reviewed'`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-196**: `POST /api/v1/records/{record_id}/supersede` by a `reviewer` or `admin` with a required `reason` bound to `reason_note_v1` against a non-deleted Timeline row whose current `capture_state` is `rough`, `enriched`, or `reviewed` returns `200 OK`, increments `row_version`, appends a new `change_set`, and sets `capture_state='superseded'`; when `replacement_record_id` is omitted, the success payload returns `replacement_record_id=null`; when `replacement_record_id` is present it is non-nullable, so explicit JSON `null` fails with `400 error.code='invalid_mutation_payload'` and `reason_code='field_not_nullable'`; when a valid `replacement_record_id` is supplied, the same committed action also persists exactly one active `record_links` row with `link_type='supersedes'`, `src_record_id=<replacement Timeline row>`, and `dst_record_id=<superseded Timeline row>`, returns the committed `replacement_record_id`, and a subsequent Timeline row query surfaces hidden `timeline.replacement_record_id` with that same value; LF canonicalization applies before validation and idempotency comparison, normalized-empty input is rejected, disallowed control characters are rejected, and values longer than 4096 Unicode scalar values fail closed.
  - Verifies: REQ-01-086, REQ-01-311..REQ-01-312, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-169, REQ-02-175, REQ-02-181, REQ-03-102..REQ-03-106
- **AC-197**: After a Timeline row reaches `superseded`, ordinary grid edits, `PATCH /api/v1/records/{record_id}` field mutations, and `POST /api/v1/records/{record_id}/mark-reviewed` fail closed; leaving `superseded` requires rollback of the superseding change.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-198**: Removing all current links or evidence from an already `enriched` Timeline row does not downgrade it to `rough`, and `timeline.has_unresolved_mentions` is `true` if and only if the current row still has at least one non-deleted unresolved mention, regardless of `capture_state`.
  - Verifies: REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110
- **AC-199**: Replaying the same normalized `POST /api/v1/records/{record_id}/mark-reviewed` or `POST /api/v1/records/{record_id}/supersede` request by the same authenticated actor with the same `(record_id, client_txn_id)` returns the originally committed success and does not create a second lifecycle transition. Normalized request comparison excludes route-key dimensions, including `client_txn_id`.
  - Verifies: REQ-03-102..REQ-03-110
- **AC-329**: A supersede request whose `replacement_record_id` equals the target row, points to a non-Timeline record, another-incident record, a soft-deleted or otherwise non-visible row, a replacement Timeline row whose current `capture_state` is `superseded`, or a target Timeline row that already has another active incoming Timeline `supersedes` link fails closed with `409`, `error.code='illegal_transition'`, and the applicable `error.details.violated_guards[]` entries drawn only from `replacement_must_be_different_timeline_record`, `replacement_must_be_visible_active_same_incident_timeline_record`, `replacement_must_not_be_superseded`, and `target_must_not_have_active_replacement`; no partial `capture_state` change or replacement link commits.
  - Verifies: REQ-01-086..REQ-01-087, REQ-02-168, REQ-03-106
- **AC-330**: Exact replay of a committed supersede-with-replacement request by the same authenticated actor with the same `(record_id, client_txn_id)` returns the original committed success and does not create a second replacement link, `change_set`, or lifecycle transition; reusing that same key for a different normalized `replacement_record_id` or normalized `reason` fails with `409 error.code='client_txn_conflict'` before stale `base_row_version` evaluation; and changing only route-key dimensions such as `client_txn_id` does not change the normalized request hash used for comparison.
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
- **AC-150**: Workbook open resolves the starting surface in this order: explicit `sheet_ref`, caller `home_sheet_ref`, incident `default_sheet_ref`, then `cartulary.view.timeline.v2`; direct selection of any pack-independent base-profile surface from `REQ-01-307`, including `cartulary.view.task_requests.v1`, `cartulary.view.decisions.v1`, `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`, `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`, succeeds when those pointers use `sheet_ref.kind='view_schema'` with the standardized `view_schema_id`, even if ordinary `GET /api/v1/incidents/{incident_id}/saved-views` returns no matching saved-view object; and if a persisted pointer is unusable, the exact invalid value is conditionally cleared and startup falls through deterministically. An effective clear advances `updated_at` once, an incident-default clear attributes `updated_by_user_id` to the authenticated repair-triggering caller, and a concurrent replacement, already-null pointer, or other no-op preserves the replacement, timestamp, and attribution and restarts selection from current state when needed.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032
- **AC-478**: The workbook startup and preference contracts accept exactly the closed three-variant `sheet_ref` union defined by Core 01: `{kind='view_schema', id}`, `{kind='saved_view', id}`, or `{kind='extension_workspace', extension_profile_id, workspace_key}`. A claimed and caller-visible `network_flow_activity/network_analysis` extension pointer round-trips through home and default preferences and startup, including when the incident has zero Network Flow tables; startup returns that exact canonical reference with `selected_view_schema_id=null` and `selected_saved_view=null`; the canonical explicit query is `sheet_ref_kind=extension_workspace&sheet_ref_id=network_analysis&extension_profile_id=network_flow_activity`; and no `workspace_key` query alias is accepted. Missing, empty, unknown, extra, and mixed-variant members fail closed without changing preferences. Claim loss, an undeclared workspace, or authorization loss conditionally clears the exact persisted pointer using the exact Core 01 cleared-pointer reason and continues the ordinary fallback chain; a comparison miss preserves the concurrent replacement and restarts selection from current state. The matching explicit selector receives the exact explicit-selector reason without changing persisted pointers. Extension-workspace collaboration presence uses the same reference and omits `record_id` and `field_key`.
  - Verifies: REQ-01-138..REQ-01-151, REQ-01-228..REQ-01-239, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032, REQ-03-092, REQ-03-094
- **AC-112**: An analyst can create a note directly in the Notes sheet by inline create from the sheet itself with no required modal; the row commits only when `note.title` or `note.body` remains non-empty after `single_line_title_v1` or `multiline_body_v1` normalization, whitespace-only values do not commit, C0/C1 control characters outside the allowed multiline set are rejected, `note.title` values longer than 512 Unicode scalar values fail closed, `note.body` values longer than 16384 Unicode scalar values fail closed, and the resulting record is stored as an artifact with `artifact_type='note'`, ordinary note attribution, and timestamps.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-01-487..REQ-01-491, REQ-02-067..REQ-02-071,
    REQ-03-004, REQ-03-242..REQ-03-246
- **AC-068**: From a selected Timeline, Host, Identity, or Evidence record, an analyst can invoke `add linked note` without leaving the workbook flow; the action may preseed the contextual link, but the note does not commit until `note.title` or `note.body` remains non-empty after `single_line_title_v1` or `multiline_body_v1` normalization, whitespace-only values do not commit, C0/C1 control characters outside the allowed multiline set are rejected, `note.title` values longer than 512 Unicode scalar values fail closed, `note.body` values longer than 16384 Unicode scalar values fail closed, and the committed record is stored with the same artifact record shape as a Notes-sheet-created note, including `artifact_type='note'`, ordinary note attribution, and timestamps. The note appears in the Notes sheet and is visible as a linked record from the source row.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-01-487..REQ-01-491, REQ-02-067..REQ-02-071
- **AC-069**: Renaming the visible Notes tab label, and any implementation-supported per-user hide/show operation for that built-in tab, does not change write-back or export semantics because behavior remains bound to `view_schema_id`.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330
- **AC-070**: Editing a lightweight free-text field on a timeline, host, identity, or evidence record does not create a standalone Notes row. Creating a standalone note does create a distinct note record that can be linked, tagged, and reviewed in history independently.
  - Verifies: REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071
- **AC-015**: An analyst can create a no-blob Evidence request record only when the first committed create includes at least one owner-defined qualifying signal: non-empty normalized `evidence.title`, external non-reserved `evidence.storage_ref`, `evidence.collector_party_text`, or `evidence.source_party_text`; explicit valid `evidence.lifecycle_state`; or non-null valid `evidence.requested_at` or `evidence.received_at`. Party IDs alone, server defaults, read-only or derived values, preseeded context, explicit clearable `null`, and text normalized to empty do not qualify. String signals are evaluated after `single_line_title_v1`, `locator_text_v1`, or `party_text_v1`, so whitespace-only or control-only input does not count; `evidence.title` values longer than 512 Unicode scalar values, `evidence.storage_ref` values longer than 1024 Unicode scalar values, and `evidence.collector_party_text` or `evidence.source_party_text` values longer than 256 Unicode scalar values fail closed. When lifecycle is omitted, the committed row defaults `evidence.lifecycle_state='requested'`; if `requested_at` is also omitted it defaults to the commit timestamp, while explicit `requested_at:null` suppresses that default. Defaults do not make a blank create valid. A same-surface create flow that reaches first commit through a successfully finalized blob attachment produces exactly one committed Evidence row. A create attempt that supplies neither a qualifying field signal nor a finalized blob returns `minimum_create_signal_missing` and commits no row. After a valid no-blob create, an analyst can later attach or replace the blob and preserve structured fields and append-only custody history.
  - Verifies: REQ-01-243..REQ-01-247, REQ-01-291..REQ-01-295, REQ-01-355..REQ-01-366, REQ-01-487..REQ-01-493,
    REQ-02-186..REQ-02-201, REQ-03-120..REQ-03-126, REQ-03-242..REQ-03-246
- **AC-521**: Every current `view_schema_resource_v2` exposes required `create_inputs[]`; every non-Evidence view returns `[]`, while `cartulary.view.evidence.v1` returns exactly optional non-null `evidence.initial_object_blob_id` bound to `object_blob_id_v1`. The generic Evidence row-create request accepts only `client_txn_id`, allowed create-writable fields, and that exact input; malformed, null, duplicate, unknown, or foreign-view inputs fail closed; the success envelope remains `data.view_schema_id`, `data.change_set_id`, and `data.row` with no sibling blob ID.
  - Verifies: REQ-01-057, REQ-01-061, REQ-01-069, REQ-01-070, REQ-01-245, REQ-01-288, REQ-01-328
- **AC-522**: Every row of REQ-01-328's minimum-signal and initial-lifecycle matrices yields the exact disposition. In particular, explicit `requested` qualifies but an omitted default does not; Party IDs alone do not qualify; reserved `object://...` user input is rejected; no-blob `available` and every initial `released` create are illegal; blob-backed `quarantined` is illegal; finalization does not auto-promote lifecycle; and the registered spelling is exactly `minimum_create_signal_missing` with no alias.
  - Verifies: REQ-01-328, REQ-02-190, REQ-03-116..REQ-03-126
- **AC-523**: A same-flow initial-blob create locks and rechecks the blob, then commits the record envelope, Evidence row, unique association, logical storage reference, custody, revision, projection, durable collaboration intent, and idempotency result atomically. Exact replay returns the original result without a second effect; divergent replay commits nothing; response or dispatcher failure does not repeat the source mutation; and two transactions racing for one blob produce at most one winner with no partial losing row. Missing, foreign, or already-associated blobs all return concealed `evidence_attach_rejected/blob_not_visible`. A database uniqueness constraint rejects a second non-null Evidence association, and an upgrade with pre-existing duplicates fails before changing the index rather than repairing custody automatically.
  - Verifies: REQ-01-070, REQ-01-245, REQ-01-328, REQ-02-190, REQ-04-053
- **AC-524**: New-record Evidence and Timeline file flows retain the provisional draft and selected file only in client-local state until generic row-create success, use slot then upload then one atomic row-create with `evidence.initial_object_blob_id`, retain the same row-create `client_txn_id` for uncertain transport retry, use the ordinary new-ID workflow after a definitive conflict, keep focus and workbook state continuous, and expose accessible pending, retry, blocked, and error feedback. Before row-create success, no draft row appears in query, collaboration, history, projection, export, backup, or portability state.
  - Verifies: REQ-01-328, REQ-03-116..REQ-03-126, REQ-04-021..REQ-04-030
- **AC-525**: Revisions bundle files for admitted version `2` accept only their exact contract-major-`1` row shapes and deterministic export order. Independent fixtures exercise every Revisions invariant, and multi-defect permutations always select the owner-defined first invariant and stable row identity. Failures expose only the closed source family and invariant IDs, leave no visible state, and never derive attribution from PostgreSQL error text or a descriptor-default invariant.
  - Verifies: REQ-01-639..REQ-01-642, REQ-02-204, REQ-02-217..REQ-02-218
- **AC-526**: Deployment config v2 requires one secure Revisions conflict-token key-ring manifest before listeners or workers. Exact manifest, path, key, secret, purpose, active/decrypt-only rotation, nonce, TTL, skew, opacity, tamper, expiry, retirement, and uniform-error fixtures pass; v1 configuration and v2 conflict tokens are rejected without an authentication-master or hard-coded fallback; and a client retains its local draft while refreshing an invalidated conflict.
  - Verifies: REQ-03-066, REQ-03-075..REQ-03-078, REQ-04-069, REQ-04-077..REQ-04-078, REQ-04-111, REQ-04-147..REQ-04-149
- **AC-527**: A Revisions import acquires transaction-scoped sequence exclusion before row application without changing the effective next value, performs the real repair only after all source-owner validation, blocks concurrent ordinary allocation, advances to at least the larger of the pre-import next value and imported maximum plus one on commit, and restores the original sequence state after a later injected failure. Runtime uses fixed hardened migration-owned functions and no `setval`.
  - Verifies: REQ-01-640..REQ-01-642
- **AC-528**: Delete, restore, rollback, and explicit conflict resolution enforce authentication, cookie CSRF, path syntax, hidden visibility, role, token where applicable, and content/body validation in owner-defined order. Unauthorized malformed requests reveal no body- or selector-specific detail and commit no idempotency, source, history, projection, or Collaboration effect; authorized valid requests preserve their public methods, paths, operation IDs, envelopes, and consequences.
  - Verifies: REQ-01-074, REQ-01-100, REQ-03-075
- **AC-529**: Revisions owns generic change-set/revision history, revision-window, conflict token/text-merge/`keep_saved`, opaque selector lifecycle, indexed association lookup, transaction/lock/idempotency/publication order, and rollback coordination behind consumer-owned ports and immutable application-composed provider catalogs. All ten current record types produce a closed `{snapshot_schema_id, record, source}` authoritative snapshot that passes its exact source-owner validator; no stored non-null row snapshot is schema-less or projection-derived. Ordinary live revisions persist transactionally atomic, revision-bound, field-keyed conflict facts derived only from explicit live-change input; those facts preserve scalar and collection conflict consequences without becoming row-history, rollback, projection, or portability authority. All fourteen current target kinds resolve exactly once to pure history semantics and generic `row` or `non_row` rollback dispatch; persisted association arrays are sorted, unique, complete, indexed, and equal owner-derived facts. Source owners retain current-state, field, collection, association, companion, revalidation, and inverse semantics. Incident Bundle version 2 retains its exact portable outer rows, recomputes association facts, omits conflict facts, and rejects schema-less snapshots. HTTP/auth/platform concerns terminate at adapters; Records construction, view-schema resolution, and process-environment capture terminate at application/server assembly; deployment-local administrative audit remains under Authentication and Administration. Static and negative-runtime boundaries reject concrete Records construction, reverse imports, ambient environment reads, global registries, projection snapshot truth/fallback, source JSON-key history predicates, target/source-type rollback switches, dynamic relation metadata, incomplete catalogs, unauthorized provider invocation, and non-atomic provider failure.
  - Verifies: REQ-00-071, REQ-01-650, REQ-01-659, REQ-02-204, REQ-02-216..REQ-02-218, REQ-02-265, REQ-03-066
- **AC-530**: Indicator observations accept and persist exactly `manual_entry`, `clipboard_paste`, `csv_import`, `xlsx_import`, `api_import`, `extraction`, and trusted-internal `system`. `interactive_cell`, empty, missing, case-folded, whitespace-padded, aliased, unknown, extension-prefixed, and ordinary caller-selected `system` values fail before the first database write. Every live producer emits its assigned token; ordinary HTTP analyst entry emits `manual_entry`. Exact tokens survive history, rollback, and Incident Bundle round trips, and repeated equal-content observations with distinct stable identities remain separate. Every invalid-origin fixture proves no observation, source or Indicator version, change set, revision, projection, Collaboration intent, idempotency success, or publication effect.
  - Verifies: REQ-01-639..REQ-01-642, REQ-02-075..REQ-02-080, REQ-02-260
- **AC-531**: Indicator source-major-`1` files for admitted bundle version `2` accept only the three exact REQ-01-640 row schemas, explicit nullable members, canonical scalar forms, and stable-identity export ordering. Interval rows accept only lifecycle state `active`, `benign`, `false_positive`, or `retired` and require unique canonical support-reference UUIDs. Independent negative fixtures exercise each of the ten Indicator invariants plus unknown, case-variant, whitespace-padded, aliased lifecycle tokens and duplicate support references; three multi-defect fixtures under different archive and row permutations always select the owner-defined lowest-precedence invariant and stable row identity. Valid v2 export/import/export is deterministic, active and tombstoned repeated observations remain distinct, and injected failure during apply, validation, or before final commit leaves no visible state. Unsupported, hostile, and malformed values expose only `source_family_id='indicators'` and the selected closed `invariant_id`; they disclose no row value, raw digest, SQL, relation, constraint, storage, path, or internal topology through responses, jobs, logs, telemetry, readiness, administrative summaries, or operator output.

- **AC-532**: The six Indicator observation and lifecycle route families, comprising eight HTTP operations, implement their exact read, create, resolve, dismiss, restore, and append contracts. Independent fixtures cover authenticated hidden-resource ordering, viewer denial, editor success, cookie CSRF, exact and divergent replay, stale base versions, every legal and illegal observation transition, source/view/field validation, ASCII and multibyte UTF-8 spans, mid-code-point and out-of-range spans, server-derived text/locator/manual origin, canonical candidate derivation, same-incident targets/support UUIDs, the four exact lifecycle tokens, canonical times, affected-record lock/version order, row-centric history, projection refresh, ordinary Collaboration publication, and failure atomicity. Source-record failures use `indicator_source_record_not_found`; requested resolution-target failures use `resolved_indicator_not_found`; addressed Indicator failures use `indicator_not_found`; unavailable prior observation dependencies remain concealed as `indicator_observation_not_found`; invalid support references remain invalid mutation input; and storage failures use the safe internal path without being rewritten as semantic 404 or 400 responses. Every such failure commits no source, envelope, history, projection, idempotency-success, or Collaboration effect and discloses no SQL, relation, constraint, driver value, or hidden identifier. Observation and interval pages are stable, actor/record-bound, tombstone-free, newest-first, OFFSET-free, and reject cursor replay under another actor, record, route, or limit. Discovery emits the four exact Core 01 Indicator feature rows, and the client resolves their complete semantic tuples before wildcard families. Indicator and Timeline Inspector handlers call the real routes, never generic record patch, preserve selection, expose accessible pending/empty/error/retry/paging states, and omit unsupported actions instead of rendering inert controls.
  - Verifies: REQ-01-615..REQ-01-617, REQ-01-652, REQ-01-654, REQ-02-263..REQ-02-264, REQ-03-306, REQ-04-150

- **AC-533**: A clean install and an upgrade with valid existing Indicator rows create and deterministically backfill exactly one `indicator_active_identities` claim per Records-authoritative active canonical identity. Concurrent create converges on one claim; delete releases it; restore fails atomically on conflict; rollback rekeys it; Incident Bundle import maintains it; recovery rebuild produces the same claims; and claims never appear in portable or backup-domain content. During expand compatibility, every writer maintains claims and mirrors atomically. After the old-writer drain gate, constraint validation and contract migration remove all Indicator envelope mirrors, legacy indexes, and mirror foreign keys; every source read obtains envelope state from Records. The schema enforces exact lifecycle tokens and unique canonical support-reference UUIDs for every writer. Empty install, valid upgrade, malformed legacy upgrade, Down/Up reconstruction, direct insert and update, delete/restore, rollback, recovery, and valid bundle v2 round trips all pass. Unknown lifecycle state, duplicate support reference, envelope drift, malformed child tuple, incompatible idempotency payload, or duplicate Records-active identity blocks the applicable migration or write atomically without guessing, silent data repair, or partial schema change.
  - Verifies: REQ-01-639..REQ-01-642, REQ-02-072..REQ-02-080, REQ-02-260
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
- **AC-047**: Under representative live-update traffic within the supported large-grid operating envelope, scrolling, sorting, filtering, grouping, stale full-query refreshes, invalidation-triggered refreshes, and live updates never retarget pending edits away from the selected `record_id`, never visibly roll a locally accepted row back to an older `row_version`, and viewport stabilization does not require a full-sheet rerender.
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
- **AC-019**: Typing or pasting `WS-023?` into a Timeline Hosts cell creates an `entity_mention` and zero host records unless the analyst explicitly resolves it to an existing visible entity or invokes a separate explicit entity-create operation when such an operation is available.
  - Verifies: REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-027, REQ-02-030..REQ-02-036, REQ-03-129..REQ-03-134
- **AC-020**: Resolving a selected unresolved mention to an existing host or identity requires explicit target identity, preserves the raw mention, resolves only the selected mention by default, creates no stub entity as a side effect of `resolve_item`, and stores manual-link provenance with `confidence=null`; any future create-from-mention operation MUST be a separate explicit operation rather than omitted-target `resolve_item` behavior.
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
- **AC-474**: Creating or resolving an `indicator_observation` and creating an `indicator_state_interval` produces target-granular history visible from every affected row-centric record history. When the server exposes a single-entry action, rollback by that record's opaque `history_entry_ref` restores exact prior resolution state or tombstones the created child without hard deletion; tombstoned children remain in retained history and incident portability but are excluded from active matching, aggregates, lifecycle selection, and projections. Whole-change-set rollback remains atomic, locks the complete canonical affected-record set before fresh-state evaluation, advances every affected first-class row version, appends inverse mutation entries and record revisions, rebuilds projections, and publishes only ordinary canonical `record_changed` events.
  - Verifies: REQ-01-089..REQ-01-111, REQ-02-205..REQ-02-218, REQ-02-238..REQ-02-242, REQ-02-260,
    REQ-03-138..REQ-03-144, REQ-03-294, REQ-04-036..REQ-04-037
- **AC-022**: Pasting into an `entity_origin` mapping upserts an existing active host or identity on a unique exact-match key and otherwise creates a stub; for hosts and identities, exact-match reuse consults the canonical field plus the full active `exact_match_reuse` identifier set for that record rather than only singular canonical slots; it never auto-merges two pre-existing entities.
  - Verifies: REQ-02-030..REQ-02-036, REQ-02-060..REQ-02-061
- **AC-023**: Merging two entities preserves loser lineage, deterministically promotes or carries loser-side `exact_match_reuse` identifiers by class, copies only `suggestion_only` aliases through the ordinary alias surface, preserves `provenance_only` values without making them reusable, repoints live mention resolutions and live links to the survivor in one change set, ensures future exact-match reuse on any preserved reusable identifier resolves to the survivor rather than creating a new stub, and does not change the survivor `record_id`.
  - Verifies: REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249
- **AC-024**: In the timeline sheet, the grouping control offers only `None`, `timeline.date_entered_sort_day`, `timeline.activity_time_pair_state`, `timeline.capture_state`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions` in the base profile, and no other grouping key. In particular, it does not offer visible Timeline v2 source-text fields, labels, `timeline.event_type`, tags, or arbitrary custom columns.
  - Verifies: REQ-03-225..REQ-03-230
- **AC-025**: A grouped timeline sheet exposes exactly one derived outline level. `expand group`, `collapse group`, `expand all`, `collapse all`, and `Group: None` work without creating editable rows, paste targets, subtotal rows, or `record_id`-bound mutation targets.
  - Verifies: REQ-03-225..REQ-03-232
- **AC-026**: While grouped, sorting, filtering, paste, autosave, conflict handling, rollback, and export flatten to underlying records only. Editing a grouped field may move the row to a different visible group, but drag-and-drop reclassification and manual row-range grouping are not available.
  - Verifies: REQ-03-225..REQ-03-233

### 9.1A Additional Base Profile criteria for direct runtime-family verification

These criteria provide direct runtime-family verification for substantive base-profile behavior that MUST NOT rely only on aggregate claim gates.

- **AC-404**: Base-profile topology is conformant only when one application deployable unit owns the browser-facing UI, the authenticated `/api/v1/*` surface, the incident-scoped `GET /ws/v1/incidents/{incident_id}` WebSocket subscription surface, and background-job execution; infrastructure MAY provision multiple instances of that unit for stopped replacement or standby, but exactly one process holds the deployment-global exclusive application-process lease and is active, ready, or serving; reverse proxies, TLS terminators, managed ingress, and load balancers MAY front the deployable unit but MUST NOT satisfy application responsibilities or permit concurrent active processes; and any deployment that splits those responsibilities across distinct application deployable units or serves active-active is non-conformant under the current contract major.
  - Verifies: REQ-01-001..REQ-01-003, REQ-04-145, EXT-REQ-213
- **AC-534**: When Recovery holds the exclusive target serving lease, an application process starts no HTTP or WebSocket listener, reports `recovery_serving_lease_active`, and exits `2`; when an application process loses its shared recovery serving lease, readiness and all admission close immediately, fatal drain runs once, `recovery_serving_lease_lost` is reported, and the process exits `70`; Recovery's exclusive target-lease loss retains its existing Recovery error mapping and does not use either application failure token.
  - Verifies: REQ-04-113, REQ-04-146
- **AC-405**: Attaching a binary evidence payload commits one evidence row whose structured incident state contains metadata and object-reference state only; preview or download of that payload succeeds only while the authoritative object-store payload remains available; loss of object-store availability after commit leaves previously committed structured incident rows intact while preview or download fails closed as unavailable; and a Postgres logical backup or logical dump of the authoritative structured data store taken after attachment exposes no inline copy of that binary evidence payload.
  - Verifies: REQ-01-002, REQ-01-278..REQ-01-280
- **AC-406**: A row created from rough input with unresolved host or account text, unstructured `details`, or unstructured `source_text` can later be normalized, resolved, or canonically linked without overwriting the original rough input; after that later normalization, the original unresolved text, `details`, or `source_text` remains recoverable through the authoritative row, mention, or history surfaces.
  - Verifies: REQ-02-024..REQ-02-025
- **AC-407**: A state-changing mutation attempted without a current authenticated session fails closed with no committed change. On a fresh deployment, successful first-deployment-admin bootstrap or any other successful non-user startup mutation records an explicit system-process actor or source identity rather than an anonymous or implicit origin.
  - Verifies: REQ-04-036
- **AC-408**: One ordinary UI mutation, one structured-ingest mutation from clipboard paste or a claimed file-import path, and one rollback mutation each append attributable history or administrative audit state that records actor or explicit system-process identity, committed timestamp, mutation source, and the exact row-field or target-family mutation evidence enumerated by REQ-04-037 at the required history granularity; the three cases MUST remain distinguishable from one another in that recorded provenance.
  - Verifies: REQ-04-037
- **AC-469**: Projection-specific subsystem NLSpec authority is accepted only when the document is explicitly adopted and listed by the project taxonomy; `docs/graph_projection_nlspec.md` is adopted only for the graph-projection subsystem; R01 through R09 evidence remain informative while research-class; and adopting or substantively revising a projections-specific NLSpec triggers a re-audit of projection Core text, trackers, descriptors, rebuild behavior, query behavior, and boundary guard tests before further projection changes are accepted.
  - Verifies: REQ-00-062
- **AC-470**: Projection descriptor validation proves that projection stores are derived, deterministic, rebuildable state; that the code-backed provider registry remains the runtime authority; that any canonical manifest is validation-only; that materialized `source_record_types` remain distinct from `source_authority_modules`; that every source-authority set is unique, uses known owner keys, and contains its `source_owner_module`; and that active provider descriptors satisfy unique provider, table-owner, view-schema-owner, schema-version, restore-participation, status, and facade-package invariants without accepting descriptor v2 or the removed `source_authorities` member.
  - Verifies: REQ-01-621, REQ-01-622
- **AC-471**: Public workbook row/query behavior is characterized at the route/viewquery boundary, including validation, authorization, normalization, filtering, sorting, grouping, pagination, saved-view query validation, error mapping, and `view_row_v1` shape; any provider split preserves that characterized behavior unless a Core Document or adopted SPEC explicitly changes it.
  - Verifies: REQ-01-623
- **AC-472**: Restore readiness evidence proves recovery owns restore orchestration, projection modules own rebuild mechanics, projection rebuilds are invoked through the recovery-owned adapter contract, required provider rebuild work succeeds or fails readiness closed according to Core defaults, and the initial adapter may delegate to the existing projection rebuild implementation without changing behavior.
  - Verifies: REQ-01-621, REQ-01-624, REQ-01-625
- **AC-473**: S-04 package-level import guard evidence distinguishes production imports from test-only imports and proves production code outside the projections subsystem imports projections only through owner-approved facades, adapters, or contracts rather than projection internals, provider internals, rebuild internals, or test fixtures.
  - Verifies: REQ-01-626
- **AC-539**: Workbook-grid Projections boundary evidence proves that the
  adopted implementation ADR is limited to package topology; that source
  owners retain authoritative source reads, typed derivation inputs, semantic
  query intent, and Reporting fact meaning; that every production read, write,
  delete, lock, cleanup, and count against each of the ten active projection
  tables executes inside compiler-contained Projections storage or query code;
  that projection writers use caller-owned transactions without beginning,
  committing, rolling back, authorizing, publishing Collaboration events, or
  mutating source/history state; and that production and test-only permissions
  remain distinct. Exact equality holds among active descriptor table IDs,
  production table-access rules, Projections recovery-state IDs, and
  schema-owned projection tables. Differential fixtures preserve public query,
  keyset bounds, complete `view_row_v1` shape, Reporting facts, typed deletion,
  rollback, deterministic rebuild order, restore failure closure, and
  pre-commit claim suppression across the package migration. The root
  `internal/modules/projections` package has no production API or production
  importer after migration, and no runtime, generator, test-routing,
  conformance, or release-evidence path consumes the ADR, tracker, Appendix I,
  or other Markdown.
  - Verifies: REQ-00-070, REQ-01-351, REQ-01-621..REQ-01-626, REQ-01-658
- **AC-540**: Artifact mutation-boundary evidence proves that Artifacts
  accepts and returns only module-native actor, operation, command, result,
  and error types; that Workbook alone maps authenticated transport state,
  route idempotency, HTTP statuses, and public payloads; that one complete,
  fallibly constructed Artifacts mutation contribution is reused for create,
  patch, conflict resolution, and contextual-note creation; and that every
  standalone current-envelope lookup and deterministic lock uses the supplied
  caller transaction through the Records-owned capability. Missing required
  capabilities fail construction, wrong-type or absent envelopes retain
  concealed-not-found behavior, deleted envelopes retain
  `record_deleted_use_restore`, and exact committed idempotency replay remains
  compatible with already stored durable response rows without duplicate
  source, history, projection, link, or Collaboration effects.
  - Verifies: REQ-01-649, REQ-01-660
- **AC-541**: Artifact source-catalog evidence proves exact set equality among
  the eight supported artifact surfaces, the 36 writable direct source
  fields, the 15 writable collection fields, and the corresponding generated
  view-schema source filters, writability, write kinds, reference contracts,
  and nullability. Generation and production construction reject missing,
  duplicate, cross-surface, read-only, unknown, or mismatched entries; direct
  SQL identifier allowlists, collection policies, conflict source keys, and
  revision source routes derive from the validated catalog; and adding a
  surface or writable field through only one projection fails closed rather
  than partially registering behavior.
  - Verifies: REQ-01-660
- **AC-542**: Production DDL Rebaseline v2 evidence proves that one pristine
  PostgreSQL 16 database applies the immutable versions `1..29` and the
  owner-approved additive versions `30..34` under lineage
  `cartulary.prod_ddl_rebaseline.v2`; v1, foreign, unmarked-nonzero, and
  contaminated states fail before v2 DDL with remediation-report schema v1,
  boundary `prod_ddl_rebaseline_v2`, reason `historical_migration_lineage`, and
  the exact reset-only hint; and no v1 SQL, migration 62, data bridge,
  export/import transition, compatibility object, dual lineage, or downgrade
  remains. Manifest and catalog evidence proves exact file, object, logical
  owner, dependency, FK coverage, routine, Recovery, SQLC, and per-purpose
  access allocation. Prerequisite matrices prove exact administrator-owned
  `pgcrypto` 1.3 and `citext` 1.6 state, contamination rejection, qualified
  DDL, transactional Up sections, and exact rollback-through-zero residue.
  Resolver, connection, and PostgreSQL catalog matrices prove three-purpose
  credential isolation, closed failure precedence and redaction, no-follow
  bounded files, fixed role attributes and memberships, per-connection
  `session_user`/`current_user` establishment including recycled connections,
  object ownership, explicit ACLs, `PUBLIC` revocation, default-deny future
  objects, runtime denial of migration and restore authority, and complete
  Recovery under `cartulary_recovery`. Recovery remains exactly 113/84 with the
  unique Revisions conflict-fact entry; evidence remains v2, remediation
  remains v1, and public product behavior and SQLC have no unexplained delta.
  - Verifies: REQ-01-661, REQ-04-153, TH-HARNESS-REQ-811

- **AC-545**: Timeline policy evidence proves that direct writable field
  membership, the 32-change patch limit, the 64-action collection limit, the
  32,768-rune visible-text limit, and allowed control characters have one
  immutable implementation owner. Root mutation, admission, import,
  clipboard, conflict, and fixture validation use that owner directly;
  vectors at 31/32/33 changes, 63/64/65 actions, and
  32,767/32,768/32,769 runes preserve exact accepted values, error codes,
  details, null behavior, and hashes; and no copied map, limit, or validator
  remains.
  - Verifies: REQ-04-160
- **AC-546**: Timeline capability evidence proves that admission, Workbook,
  and Imports depend only on their consumer-owned operation sets; Imports'
  create operation preserves caller-transaction ownership; performance
  fixture construction is exposed as a distinct fixture-only contribution;
  the ordinary Timeline facade has no fixture methods; nil and error behavior
  remains exact; and no production or test boundary allowlist expands.
  - Verifies: REQ-04-160
- **AC-547**: Timeline canonical-value evidence proves byte-exact parity for
  optional strings and UUIDs, UTC RFC3339Nano timestamps, UTC dates,
  nil/empty collection values, canonical JSON SHA-256, create/patch/action
  hashes, row cells and groups, conflicts, revisions, and Collaboration
  payloads. Row presentation has exactly one `view_row_v1` serializer and no
  superseded encoding helper remains.
  - Verifies: REQ-04-160
- **AC-548**: Timeline collection-fact evidence exercises every mention,
  link, and evidence success and failure position; proves exact source order,
  duplicate and nil/empty behavior, identical caller transaction, no read
  after failure, and byte-equivalent application facts, projection inputs,
  and rows. Projection-provider and Timeline application composition construct
  independent reader values without exchanging runtime authority, and
  descriptor fields, registries, and boundary allowlists remain exact.
  - Verifies: REQ-01-662, REQ-04-160
- **AC-549**: Timeline version evidence proves that primary row mutations and
  subordinate Entity-mention effects use the sole formatter and persist exact
  `timeline_record:<canonical-record-uuid>:<positive-row-version>` values
  derived from the target identifier and authoritative snapshot row version.
  Revisions persistence and rollback treat those values as opaque, every
  `timeline:` Timeline writer and fixture is absent, and no prefix parser
  selects authorization, dispatch, source ownership, SQL, snapshots, or
  providers.
  - Verifies: REQ-02-266, REQ-04-160
- **AC-551**: Timeline cleanup evidence proves that empty scaffolding and every
  superseded compatibility, adapter, formatter, and helper artifact is absent;
  executable tests have exact authored verification-contract and test-family
  ownership; generated topology is generator-produced and drift-free; and
  the final module inventory reconciles without an orphan or dead symbol.
  - Verifies: REQ-04-160
- **AC-552**: Backend Workbook boundary evidence proves that one private
  application facade registers the exact thirteen adopted Workbook operations;
  that one immutable, fail-closed catalog contains every active query, create,
  patch, conflict, clipboard, bulk, linked-note, and supersede capability; and
  that concrete source-owner adapters are constructed only by
  `internal/app/workbookassembly`. Source owners retain admission,
  normalization, defaults, canonical hashes, source mutation, history and
  projection inputs, and Collaboration consequences. Generic Workbook
  production code contains no source-owner command, result, error, field, or
  collection policy and no opaque admission value. Workbook maps only a closed
  safe failure vocabulary to public errors and imports no concrete source owner;
  its only cross-module production imports are the adopted incident-admission
  and Projections provider-contract capabilities. Exact route, authorization,
  replay, startup, effect-order, restore-probe, query, mutation, browser, and
  wire behavior remains owner-conformant. The former Store mutation facade,
  concrete provider constructors, redundant owner aliases, stale boundary
  sentinel, compatibility paths, and orphaned tests are absent; canonical
  boundary policy and exact authored verification accounting are generated and
  drift-free.
  - Verifies: REQ-00-072
- **AC-553**: Links-to-Reporting boundary evidence proves that Links exposes
  only ordered active source facts with exact field identity and attribution;
  Reporting owns field and support-reference provider ports, DTOs, paths,
  content classes, ordering, fallback, duplicate posture, and error mapping;
  and application composition is the only production layer importing both
  boundaries. The adapter uses the caller-owned source-boundary transaction,
  issues no SQL, performs no authorization or redaction, and returns no partial
  output on failure. Included link types, logical-target fallback, repeated
  targets, nulls, empty sets, deleted link and endpoint exclusion, paths,
  values, tags, and ordering match Reporting §7.1.2, while public snapshot,
  export-model, release, route, authorization, and generated contracts remain
  unchanged.
  - Verifies: REQ-01-664, REQ-02-169, REQ-RPT-025..REQ-RPT-026d
- **AC-554**: Links canonical record-link evidence proves that create, metadata
  patch, delete, merge, mention resolution and retargeting, supersession,
  contextual linked-note creation, rollback, and no-op paths use one
  Links-owned encoder. Every emitted value has exactly the fourteen members in
  REQ-02-267 with canonical scalar types, UUIDs, timestamps, vocabularies,
  nullability, active-state, attribution, confidence, endpoint, and transition
  invariants. Every result is returned atomically with its source write through
  the caller-owned transaction, and recursively fresh maps cannot mutate
  another result or provider value.
  - Verifies: REQ-02-267, REQ-02-269
- **AC-555**: Links canonical record-tag evidence proves that ordinary
  collection, merge, rollback, export, import, and re-export paths use one
  Links-owned encoder and values with exactly the ten members in REQ-02-268.
  Every target uses
  `record_tag:<canonical-record-uuid>:<canonical-record-tag-uuid>`; create uses
  the after record, patch and delete use the pre-mutation record, and rollback
  retains and validates the addressed original target. Bare UUID targets,
  mismatched components, `tag_id`, and alternate grammars fail closed.
  - Verifies: REQ-02-268, REQ-02-269
- **AC-556**: The application-composed Revisions catalog admits one pure
  Links-owned validator for each Links target. Local append, history
  description, inverse planning and application invoke it with the exact
  operation kind and both retained sides before querying or writing. Missing,
  extra, mistyped, noncanonical UUID or timestamp, invalid enum or nullability,
  illegal transition, compact, alias, and default-dependent values fail with a
  safe owner-neutral error and no persisted mutation, partial facts, source
  change, or invented restoration state. Canonical create, patch, delete, and
  rollback preserve every retained field exactly, including creation
  attribution and timestamps; Revisions contains no Links field or target
  grammar.
  - Verifies: REQ-02-267..REQ-02-269
- **AC-558**: Entities boundary evidence proves that every production export
  in the bounded root and its production child packages has an exact adopted
  `retain`, `privatize`, `remove`, or `replace` disposition; every retained
  export has a live production consumer, required interface method, or source-
  contribution role; runtime-excluded test support is not treated as a
  production API; and a synthetic unapproved export fails the executable
  inventory. Direct consumers and imports match the closed adopted topology;
  child packages do not import the root or application assembly to recover
  dependencies; and every caller-supplied transaction is borrowed without
  begin, commit, rollback, nesting, or detachment. Timeline retains automatic-
  resolution policy and transaction ownership while Entities supplies typed
  source facts and write operations. Workbook alone receives the complete
  Host/Identity store; Timeline and Assessments receive stateless borrowed-
  transaction source facts; Imports receives its import-create capability; and
  merge receives one immutable owner-local merge capability. Operational
  constructors accept explicit dependency structs, reject nil and typed-nil
  required dependencies without panic, return a nil capability and
  deterministic declaration-ordered error on failure, and admit no partial,
  option, `Must`, or fallback construction path. Cross-owner effects use
  consumer-owned injected ports, with concrete implementation and classified-
  error translation in application assembly. Generic application,
  frontend/grid, private cross-owner SQL/event, package-init registration,
  mutable-global registry, service-locator, fallback, alias, and dual-dispatch
  paths are absent. Exact harness accounting, focused and service-backed
  behavior, generated projections, affected frontend/browser surfaces, and the
  full repository gate remain owner-conformant.
  - Verifies: REQ-00-073
- **AC-559**: A clean install and valid upgrade create the adopted Entities
  source constraints and deterministically backfill exactly the active Host
  and Identity claims defined by Core 02. Concurrent same-tuple create and
  upsert converge on one record; multi-class matches spanning records fail
  atomically; deleted and merged records hold no claims; delete releases,
  restore reacquires or returns the safe conflict, rollback rekeys, and merge
  transfers loser claims only after complete canonical locking. Import,
  clipboard, Incident Bundle import, and recovery produce the same claim set.
  Go and SQL normalization agree for the versioned corpus and reject unknown
  classes. Every declared Entities bundle invariant has an independent
  deterministic negative fixture, valid version-2 bytes round-trip exactly,
  and claim rows appear in neither portable nor authoritative backup content.
  Invalid source rows or duplicate active claims block migration, import, or
  recovery before partial publication and without guessing or exposing SQL,
  relation, constraint, value, credential, or internal topology details.
  Existing routes, operation IDs, request and success shapes, authorization
  and concealment order, replay, history, projections, Collaboration events,
  schema IDs, and field keys remain exact.
  - Verifies: REQ-01-639..REQ-01-642, REQ-01-665..REQ-01-667, REQ-02-060..REQ-02-063, REQ-02-066, REQ-02-270

- **AC-560**: Indicators boundary evidence proves that the bounded source-owner
  root, owner-local admission and HTTP adapters, private identity, origin,
  vocabulary and source-state packages, and typed source providers match the
  adopted topology. Every root export has one reviewed disposition and role;
  the authorized 55-to-50 contraction removes only `CreateOutcome` and its
  four constants, and the Iteration 2 exchange replaces the exported
  test-convenience participant schema constant with production
  `RecordEnvelopePort` while retaining exactly 50 exports and identical
  participant schema bytes. No alias, forwarding, dual-result, or deprecation
  path remains. Only owner-local admission and HTTP
  adapters import root contracts; no child imports the root or application
  assembly to recover dependencies. Application composition constructs one
  Records adapter and injects separate narrow root transaction and HTTP reader
  capabilities. Root construction rejects nil and typed-nil Postgres,
  Revisions, Records, projections, source text, or clock dependencies; HTTP
  construction rejects nil and typed-nil owner, Records, or Postgres
  dependencies and a nil `DependencySet.Now`; neither boundary constructs
  Records or uses wall-clock fallback. Create, observation, transition,
  lifecycle, and list orchestration lives directly on Store, concern SQL uses
  named package functions, and self-referential services, empty repository
  namespaces, and Revisions forwarding adapters are absent. Records supplies one sorted
  caller-transaction locking snapshot for affected-record validation, private
  Records validation SQL is absent, storage failures remain safe internal
  failures, and the protected raw canonical-dedupe SQL literal and lock order
  remain byte-stable. Indicators derives the five deployed SHA-256 replay
  identities from their exact logical JSON preimages before normalization or
  sorting can change membership or order; callers cannot override hashes and
  existing persisted rows replay without rewrite or compatibility branches.
  Caller transactions are borrowed, cross-owner effects use typed
  injected ports, one immutable exact runtime vocabulary owns all four
  Indicator token families, and one fallibly validated immutable source-state
  catalog produces the exact three authoritative, one rebuildable, and three
  portability descriptors. Generic application, frontend/grid, private
  cross-owner SQL, package-init registration, mutable-global registry,
  service-locator, fallback, compatibility-shim, and generated-file hand-edit
  paths are absent. Exact harness accounting, focused and service-backed owner
  behavior, boundary, migration, generated, build, and full repository gates
  remain owner-conformant.
  - Verifies: REQ-00-074

### 9.1B Network Flow Activity Extension Profile criteria

- **AC-475**: When the Network Flow Activity Extension Profile is claimed, deployment startup rejects missing safe-digest key-ring configuration, duplicate safe-digest key IDs, no active safe-digest key, multiple active safe-digest keys, malformed key IDs, unresolved safe-digest `secret_ref_v1` values, unsupported key material, safe-digest key material reused for cursor tokens, invalid rotation state, or fixture-only safe-digest key material outside a harness-owned runtime before any HTTP listener, WebSocket listener, or background-job runner starts. Representative Network Flow redaction fixtures prove that every safe digest emitted to logs, telemetry, administrative audit summaries, route error details, table-name collision details, graph-query audit details, and indicator-link non-disclosure details carries the same enclosing `safe_digest_key_id`; that comparison is performed only when key IDs and value classes match; that rotation emits new digests only under the active key without rewriting old persisted digests; that no safe digest participates in authorization, deduplication, concurrency control, table identity, row identity, or cursor validation; and that raw CSV cells, source filenames, indicator candidates, graph-query scalars, fixture secrets, production secret references, and raw key material never appear in logs, telemetry, administrative audit summaries, readiness output, or public error details.
  - Verifies: REQ-04-131..REQ-04-134
- **AC-476**: When the Network Flow Activity Extension Profile is claimed, transactional audit fixtures prove that import final commit emits exactly one `network_flow_table_created` occurrence for each created table and none for failed, cancelled, preview-only, no-table, or rolled-back units; changed rename emits exactly one `network_flow_table_renamed` occurrence and normalized no-op rename emits none; soft delete emits exactly one `network_flow_table_soft_deleted` occurrence; successful graph query emits exactly one `network_flow_graph_query_executed` occurrence and failed, cancelled, stale, unauthorized, or over-limit graph queries emit none of that family; indicator-link insert emits exactly one `network_flow_indicator_binding_created` occurrence; binding reuse under a new `client_txn_id` emits exactly one `network_flow_indicator_binding_reused` occurrence; exact committed idempotency replay returns the original success and original audit correlation without a new Network Flow domain occurrence; same-key different-digest conflicts and every pre-commit failure emit none of these domain occurrences; and injected failures at audit append, idempotency success, outbox publication, or final commit leave no partial table, binding, idempotency success, or domain audit occurrence.
  - Verifies: REQ-04-135..REQ-04-138
- **AC-477**: When the Network Flow Activity Extension Profile is claimed, lifecycle and retention fixtures prove that soft delete is terminal; that no restore, hard-delete, purge, deleted-list, deleted-inspection, or deleted-query route is exposed in v1; that soft-deleted tables are retained but excluded from ordinary table, row, diagnostic, graph, contributor, indicator-link, default-list, and `all_active_tables` behavior; that visible direct references fail with `network_flow_table_not_active` while hidden references follow Core hidden-resource behavior; that soft delete invalidates affected workspace state, cursors, and ephemeral graph results without leaking hidden state; that active and retained table-count accounting matches the closed lifecycle table; that exact committed idempotency replay returns the original success while non-replay repeat delete fails as non-active; that raw source bytes, upload capabilities, staging rows, temporary paths, object-store keys, and worker paths follow Core import-job/source retention and never appear in Network Flow routes, audit payloads, diagnostics, logs, telemetry, Graph Projection input, or public fixtures; that failed, cancelled, preview-only, and rolled-back imports leave no queryable table or public table reference; that Core-governed audit retention is not shortened by Network Flow; and that no current Network Flow conformance evidence claims a whole-incident purge.
  - Verifies: REQ-04-139..REQ-04-142

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
- **AC-065**: Re-applying the same `(import_unit_id, mapping_fingerprint, incident_id)` tuple outside exact committed idempotency replay fails with `import_apply_blocked` and `reason_code='duplicate_apply_blocked'`; exact replay returns the immutable original result, and intentional re-import succeeds only through a new import session. A one-field change to `unknown_column_policy`, `source_column_ordinal`, `source_header_text`, `field_key`, `entity_binding_mode`, `transform_id`, `transform_options`, or `empty_value_policy` changes `mapping_fingerprint`, while non-semantic JSON key order does not.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-169..REQ-03-192
- **AC-066**: Selecting overlapping `import_unit` rectangles in one batch is blocked before apply with an explicit overlap diagnostic. Non-overlapping selected units apply in deterministic order, one atomic `change_set` per unit, and the session can finish in `partially_applied`.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-169..REQ-03-186
- **AC-067**: Preview or apply of a unit containing formulas, merged cells, comments, pivots, charts, external links, hidden or filtered presentation state, or workbook or sheet protection never executes workbook behavior. The unit either emits only declared `warning_code[]` values or is rejected as unsupported, formula cells without stored cached values do not enter `ready` while mapped, and encrypted or password-protected workbooks that cannot be parsed are rejected before discovery.
  - Verifies: REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-153..REQ-03-161, REQ-03-193..REQ-03-204


- **AC-262**: `POST /api/v1/import-sessions` accepts only `multipart/form-data` with a required `boundary`, exactly two leaf parts named `metadata` and `file`, and no alternate upload framing; part order is non-semantic; the `metadata` part uses `Content-Type: application/json` or `application/json; charset=utf-8`, is UTF-8 and BOM-free, parses as one JSON object, and contains required `incident_id` plus required `client_txn_id`; the `file` part media type is exactly one of `text/csv`, `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, or `application/octet-stream`; omitting `assistant_profile` defaults it to `phase2_workbook_import_v1`; duplicate or unexpected parts, malformed metadata JSON, duplicate metadata keys, a metadata JSON value that is not an object, invalid part content type, an unknown metadata member, or any envelope failure create no durable `import_session`, no idempotency commit, and no discovery job; replaying the same normalized metadata and file bytes by the same actor with the same `(incident_id, client_txn_id)` returns the original accepted `202` result even when boundary text, part order, or advisory filename differ; and when the discovery job succeeds, the terminal job summary uses `result_summary.code='import_session_discovered'` and exactly one `resource_refs[]` item `{ kind='import_session', id=<import_session_id>, route='/api/v1/import-sessions/{import_session_id}' }`.
  - Verifies: REQ-01-033, REQ-01-466..REQ-01-473, REQ-01-549..REQ-01-553
- **AC-263**: `GET /api/v1/import-sessions/{import_session_id}` returns exactly the durable `import_session resource` members; `GET /api/v1/import-sessions/{import_session_id}/units` returns `data.import_units[]` with each element using the exact durable `import_unit resource` shape; `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}` returns that same exact durable `import_unit resource`; the two singleton import reads reject pagination members; `GET /api/v1/import-sessions/{import_session_id}/units` uses the common cursor-pagination contract; `selected_unit_ids[]`, `blocking_diagnostics[]`, `nonblocking_warning_codes[]`, and `warning_codes[]` always serialize as `[]` when empty; `mapping_fingerprint` and `approved_mapping` are absent before mapping approval; serialized `header_row_ref` and `data_start_row_ref` are positive 1-based row references within `source_rect_a1`; and preview stays read-only against incident state, is not inlined into the durable unit resource, and exposes exactly the declared preview top-level members, exact `columns[]`, `preview_rows[]`, and `cells[]` shapes, the closed `cell_kind` vocabulary, preserved source order, and the first-50-row cap with `truncated` state.
  - Verifies: REQ-01-466..REQ-01-469, REQ-01-472, REQ-01-474
- **AC-264**: `PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping` accepts only the closed mapping body, persists one approved mapping plan, returns the exact durable `import_unit resource`, and derives deterministic `mapping_fingerprint`; non-contiguous ordinals, duplicate non-null target fields, invalid `unknown_column_policy`, `preserve_custom_attrs` for Timeline, `reject_if_unmapped` when any source column remains unmapped, invalid `transform_id` or `transform_options`, invalid `empty_value_policy`, or non-exhaustive `source_columns[]` fail with the import-family invalid-request surface; for Timeline mappings with `unknown_column_policy='preserve_raw_capture'`, every applied unmapped source cell is persisted as one bounded `timeline_source_provenance` row keyed by source identity hash plus row and column ordinal and carrying import identity, mapping fingerprint, source kind and content hash, parser identity, locator, source rectangle, scalar header JSON, raw value, and cell kind; after mapping approval, `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}` reflects the exact `approved_mapping` read object; `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select` and `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip` require `client_txn_id`, return `data.unit` in that same exact `import_unit resource` shape, update persisted `selected_unit_ids[]` deterministically, preserve prior approved mappings when skipped, and recompute status from prior mapping when re-selected; `POST /api/v1/import-sessions/{import_session_id}/apply` omitting `selected_unit_ids[]` uses the session's persisted selection; apply returns `202` with the common job resource; when apply succeeds, the terminal job summary uses `result_summary.code='import_session_applied'` when the durable `session_status='applied'` and `result_summary.code='import_session_partially_applied'` when the durable `session_status='partially_applied'`; in both success cases `resource_refs[]` contains exactly one `import_session` ref using `/api/v1/import-sessions/{import_session_id}` as the canonical route; and the durable import-session and import-unit resources never surface job-status tokens such as `queued` or `running` as `session_status` or `unit_status`.
- **AC-264A**: `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview` accepts only the closed analytical-extension candidate, derives all source capabilities and descriptors server-side, requires the target mapping-approval role, invokes exactly the registered owner preview facade, validates the closed owner result before returning the exact `extension_mapping_preview` wrapper, performs no durable approval, selection, target allocation, apply admission, or domain audit occurrence, and fails closed with safe import-family details for an unavailable facade, schema mismatch, owner rejection, authorization loss, or superseded source. Repeating an identical request has no additional effect, and preview responses expose no reusable capability, internal path, source hash beyond the owner-adopted result, raw unsafe value, or stack detail.
  - Verifies: REQ-01-467..REQ-01-470, REQ-01-472, REQ-01-474
- **AC-264B**: `POST /api/v1/import-sessions/{import_session_id}/units/{base_unit_id}/regions` accepts exactly `client_txn_id` and one-based inclusive `source_rect={start_row,start_column,end_row,end_column}`; the base unit is an XLSX used-range unit in the addressed session; the rectangle is non-empty, contained, and within configured limits; success creates or exactly replays one durable unselected `operator_region` unit using the ordinary import-unit response; malformed, wrong-base, non-contained, or over-limit requests fail with `invalid_import_request` and `reason_code='invalid_source_rect'`; and no mapping, selection, owner mutation, or job is created.
  - Verifies: REQ-01-472, REQ-01-474, REQ-01-620e, REQ-03-181a
- **AC-265**: Import-family routes use only `invalid_import_request`, `import_session_not_found`, `import_unit_not_found`, `import_state_conflict`, `import_source_unsupported`, `import_source_rejected`, `import_apply_blocked`, and `incident_closed`; `invalid_import_request` uses the exact REQ-01-475 registry, including analytical target/preview reasons and `invalid_source_rect`; multipart part-related failures populate the required safe part details; `import_state_conflict`, source-unsupported, and source-rejected failures use only their exact Core registries; blocked apply uses only `overlapping_units`, `duplicate_apply_blocked`, `unit_not_ready`, `target_view_schema_not_importable`, `target_kind_not_importable`, `owner_create_contract_unavailable`, `owner_apply_contract_unavailable`, `owner_create_validation_failed`, `owner_apply_validation_failed`, or `source_changed`; registered owner errors appear only as schema-validated safe nested detail; and unknown owner tokens, capabilities, paths, raw source values, package names, and stack text are never echoed.
  - Verifies: REQ-01-471, REQ-01-475, REQ-01-553
- **AC-463**: Import parser isolation is binary: no package outside `internal/modules/imports` imports XLSX/OpenXML parser packages, workbook-shape parser DTOs, or parser-specific source heuristics; owner modules may consume only stable tabular-ingest value types and owner facade requests.
  - Verifies: REQ-01-618, REQ-02-259, REQ-03-293
- **AC-464**: Dispatcher isolation is binary: `imports` owns the apply coordinator, import-owned state, apply journal, warnings, diagnostics, mapping/source integrity, durable unit outcomes, and finalization; it does not write owner source, analytical, projection, workbook, or grid state directly; every view target selects exactly one owner-create facade; every analytical target selects exactly one valid `cartulary.imports.analytical_facade_binding.v1`; and missing, duplicate, inactive, or unsupported targets fail closed without fallback.
  - Verifies: REQ-01-618, REQ-01-620, REQ-03-293
- **AC-465**: Every registry row with `import_apply_status='supported'` has exactly one callable owner create facade and at least one characterization test proving that mapped rows create or reuse the correct owner record family without using workbook DTOs or workbook store rows.
  - Verifies: REQ-01-619, REQ-02-259
- **AC-466**: File-based import of Hosts, Identities, Evidence, Task Requests, Decisions, and coordination-artifact units creates or reuses records through the target owner facade, blocks unknown-column retention where the target cannot retain unmapped values, and keeps generated row refreshes bound to `view_row_v1`.
  - Verifies: REQ-01-619, REQ-01-620, REQ-03-293
- **AC-467**: Replaying apply after a unit commit does not duplicate rows, resources, change sets, revisions, mentions, aliases, indicator observations, links, audit/outbox effects, or imported owner provenance. A successful unit commits its owner effects, apply journal, immutable owner result and resource refs, source/mapping fingerprints, idempotency success, durable unit outcome, recovery fact, and required participants atomically. Failure injection before commit leaves none; a crash after commit is recovered by an idempotent session/job finalizer that creates no owner resource.
  - Verifies: REQ-01-618, REQ-01-619, REQ-02-259, REQ-03-293
- **AC-467A**: Immediately before each unit mutation, apply rederives the current actor, incident membership and role, incident lifecycle, target visibility/importability, extension claim, facade binding, source digest, mapping fingerprint, and target-specific capability state inside the serialized unit transaction. Role revocation, membership removal, incident close, claim removal, target deletion, source change, and mapping change committed first prevent owner mutation; a valid unit commit that wins the same serialization race remains committed and is observed by the later administrative change.
  - Verifies: REQ-01-620, REQ-01-620b, REQ-01-620c, REQ-03-293
- **AC-468**: File-import row plans are independent of React Data Grid, Handsontable, or any other grid-vendor row model. Changing a source header label or a visible destination label cannot change write-back behavior except through the raw `source_header_text` captured in the approved mapping fingerprint.
  - Verifies: REQ-01-618, REQ-01-620, REQ-03-190

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
- **AC-057**: Rendering for a chosen `release_scope` fails closed if the selected redaction profile is invalid or post-redaction validation rejects the rendered export-model fields or blocks, and the rendered artifact includes a `redaction_manifest` keyed by stable export-model path and rule identifier.
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
- **AC-062**: `mermaid` and `slidev` outputs may be published as `external_release` only when every rendered block is eligible for the chosen `release_scope`, and `markdown`, `html`, and `reenactment` selectors are rejected as future-only current-profile output kinds.
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


- **AC-266**: `POST /api/v1/snapshots` accepts a JSON object with required `incident_id` and required `client_txn_id`; omitting `source_change_set_high_watermark` resolves snapshot materialization to the current committed incident head at job admission as the exact Core 02 REQ-02-143b `cartulary.source_boundary.v1:<sha256>` token; zero, one, many, and equal-timestamp change-set cases produce the specified member order, null behavior, lowercase UUIDs, UTC RFC3339Nano timestamps, `created_at DESC, change_set_id DESC` selection, exact bytes without whitespace or newline, and SHA-256 token; incident version, change-set selection, and materialization share one repeatable-read snapshot; explicit `null` fails with `400 error.code='invalid_snapshot_request'` and `reason_code='field_not_nullable'`; a supplied boundary outside the exact committed source-boundary vocabulary for that incident fails with `409 error.code='snapshot_source_boundary_conflict'`; older incident-version-only tokens are rejected rather than translated; omission and explicit transmission of the same resolved committed boundary compare equal for idempotency and replay; exact replay uses the originally committed canonical bytes and boundary rather than a later incident head; the route returns `202` with the common job resource; on success the terminal job summary uses `result_summary.code='snapshot_created'` and exactly one `resource_refs[]` item `{ kind='snapshot', id=<snapshot_id>, route='/api/v1/snapshots/{snapshot_id}' }`; and `GET /api/v1/snapshots/{snapshot_id}` rejects pagination members and returns exactly `snapshot_id`, `incident_id`, `created_by_user_id`, `created_at`, `snapshot_at`, `source_change_set_high_watermark`, `derivation_version`, and `export_model_sha256`, with the resolved `source_change_set_high_watermark` always serialized and no inlined template, redaction-profile, release-state, approval, redaction-manifest, or rendered-byte fields.
  - Verifies: REQ-01-033, REQ-01-466..REQ-01-477, REQ-02-143b
- **AC-267**: `POST /api/v1/releases` requires exact `snapshot_id`, `template_id`, `template_version`, `redaction_profile_id`, `redaction_profile_version`, `output_kind`, and `client_txn_id`; render admission resolves `redaction_profile_sha256` and `render_admitted_at` exactly once and binds both into release provenance; the only allowed `output_kind` values are `slidev` and `mermaid`; omitting `release_scope`, `recipient_partition_refs[]`, `output_options`, `graph_projection_refs[]`, and composition tuple fields materializes the declared defaults; explicit `null` fails except for all-null composition tuple members; the only allowed `release_scope` values are `internal_draft`, `internal_review`, and `external_release`; non-empty recipient partitions are valid only for `external_release`; `graph_projection_refs[]` sorts by `graph_view_id`; composition tuple members are all-null or all non-null and reject `latest`; omission and explicit defaults compare equal for idempotency and replay; requests using `latest`, `current`, or missing version selectors fail closed with `400 error.code='invalid_release_request'`; render starts as a `202` common job rather than a blocking route; and on success the terminal job summary uses `result_summary.code='release_created'` and exactly one `resource_refs[]` item `{ kind='release', id=<release_id>, route='/api/v1/releases/{release_id}' }`, with the durable `release resource` always serializing the materialized release tuple fields.
  - Verifies: REQ-01-466..REQ-01-470, REQ-01-476..REQ-01-477
- **AC-268**: `GET /api/v1/releases/{release_id}` returns exactly the durable `release resource` members including `recipient_partition_refs`, `output_options`, `graph_projection_refs`, `composition_id`, `composition_version`, `composition_sha256`, `render_admitted_at`, `redaction_profile_sha256`, `output_media_type`, `redaction_manifest_sha256`, and `render_failed_reason_code`; nullable timestamps, invalidation reason, render failure reason, and all-null composition tuple members serialize as JSON `null` when unset; output and manifest fields are non-null except on `render_failed` releases; rendered release resources use only current-profile closed-vocabulary `output_kind`, current-profile closed-vocabulary `release_scope`, and only `pending_approval`, `approved`, `invalidated`, `published`, and `render_failed` as `release_state`; `internal_draft` candidates become `approved` immediately on successful render; `approve`, `publish`, and `invalidate` each require `client_txn_id`, reject pagination members, and enforce their declared legal state transitions without introducing `queued`, `running`, or `rendering` into `release_state`; `publish` and `invalidate` additionally require current incident `admin` role; and the durable release resource never inlines approval records, full redaction manifests, rendered bytes, or worker-phase status.
  - Verifies: REQ-01-467..REQ-01-470, REQ-01-374..REQ-01-376, REQ-01-476, REQ-01-478, REQ-04-031..REQ-04-035
- **AC-269**: Snapshot and release routes use only `invalid_snapshot_request`, `snapshot_not_found`, `snapshot_source_boundary_conflict`, `invalid_release_request`, `release_not_found`, `release_state_conflict`, `release_approval_rejected`, and `release_render_failed`; hidden snapshot member resources fail as `snapshot_not_found`, hidden release member or action resources fail as `release_not_found`, and reporting routes do not expose incident-level not-found errors for hidden reporting resources; render failures use only the stable `release_render_failed` reason codes declared by the error registry; state conflicts use only stable `release_state_conflict` reason codes declared by the error registry; and approval rejections use only stable `release_approval_rejected` reason codes while the durable release state still permits an approval attempt.
  - Verifies: REQ-01-471, REQ-01-479
- **AC-305**: `POST /api/v1/releases/{release_id}/approve`, `publish`, and `invalidate` accept only a JSON object with required `client_txn_id` and optional `reason`; if present, `reason` accepts only JSON string or JSON `null` and normalizes under `reason_note_v1`; omission, explicit `null`, and normalized-empty input compare equal; unknown top-level members, non-object bodies, missing `client_txn_id`, or `null` for non-nullable members fail with `400 error.code='invalid_release_request'`; idempotency is route-scoped on `(actor_user_id, release_id, action_route, client_txn_id)`; exact replay returns the original committed success before fresh state evaluation; and same-key different-request reuse fails with `409 error.code='client_txn_conflict'`.
  - Verifies: REQ-01-469..REQ-01-470, REQ-01-478, REQ-01-496
- **AC-306**: A successful `approve` returns `200 OK` with the common success envelope and `data` equal to the same exact durable `release resource` shape returned by `GET /api/v1/releases/{release_id}`; a valid approval may leave the durable `release_state` at `pending_approval` when the full approval set is not yet satisfied.
  - Verifies: REQ-01-478, REQ-04-034
- **AC-307**: A fresh `approve` against an already `approved` release fails with `409 error.code='release_state_conflict'` and `reason_code='already_approved'`; a fresh `approve` or `publish` against an already `published` release fails with `reason_code='already_published'`; a caller or artifact tuple that cannot contribute a valid approval while the release remains `pending_approval` fails with `409 error.code='release_approval_rejected'` using the registry-declared actor or approval-role reason; and invalidated-state failures continue to use the `release_state_conflict` family.
  - Verifies: REQ-01-471, REQ-01-479

### 9.4 Reference Pack Extension Profile criteria

- **AC-033**: Reference-pack import, verification, and refresh run without blocking grid editing, and the UI shows progress and cancellation within 1 second of job start.
  - Verifies: REQ-01-399, REQ-01-407..REQ-01-413, REQ-01-452..REQ-01-454, REQ-04-040..REQ-04-043
- **AC-034**: Absent, disabled, failed, or missing optional reference packs degrade only the affected overlay labels, enrichment semantics, non-canonical analytical widgets, or snapshot/report derivations; timeline capture, entity resolution, evidence attachment, and core editing continue to function; and optional pack state does not create or remove current-profile workbook `view_schema` surfaces.
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
- **AC-369**: `POST /api/v1/reference-packs/refresh` accepts required `client_txn_id` and optional `pack_keys[]`; explicit `null` for `pack_keys[]` fails with `400 error.code='invalid_reference_pack_request'` and `reason_code='field_not_nullable'`; explicit `[]` fails with `reason_code='empty_pack_keys'`; supplied `pack_keys[]` is treated as a set-like selector so caller order is non-semantic, duplicate members coalesce by exact token equality, and canonical normalization sorts the unique `pack_key` set by `pack_key asc`; unknown, non-visible, or non-string members fail with `reason_code='invalid_pack_keys'`; omitted `pack_keys[]` resolves once at refresh-job admission to all currently imported pack keys visible to the authorized deployment-admin caller and that resolved set, not later visibility state, is the value used for idempotency and replay; and if the omitted selector resolves to zero visible imported pack keys, the refresh still succeeds as a deterministic no-op background job whose terminal success remains `reference_packs_refreshed` with zero `resource_refs[]` items.
  - Verifies: REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-481
- **AC-443**: Verify that `GET /api/v1/reference-packs` accepts only `limit`, `cursor_token`, `search`, `pack_version_state`, `verification_result`, and `active`; `search` uses `list_search_v1` over exactly `pack_key`, `pack_kind`, `pack_version`, `source_identifier`, `manifest_sha256`, `payload_sha256`, and `signer_key_id`; `pack_version_state`, `verification_result`, and `active` accept only their Core 01 exact wire tokens; repeated, unknown, empty, explicit-`null`, comma-list, array-encoded, or alternate-spelling query members fail with `400 error.code='invalid_list_query'`; authorization failure is evaluated before query validation; search and filters combine with AND against the complete authorized collection before pagination; ordering remains `pack_key asc, pack_version asc` with no relevance ranking or latest-version interpretation; and cursor replay with a changed canonical search/filter state fails with `400 error.code='invalid_pagination_request'` and `reason_code='cursor_query_mismatch'`. Verify independently that the browser controller conforms to REQ-01-610A: exact effective-state cursor binding, first-page admission after material search/filter changes, semantic `searching` status exposed accessibly as `Searching reference packs`, authorized prior-row retention, newest-generation ownership of success/error/paging/pending state, stale terminal exclusion independent of transport cancellation, immediate explicit submission, and fail-closed protected-state clearing on authentication, capability, profile, or deployment-authorization loss.
  - Verifies: REQ-00-059, REQ-01-481, REQ-01-583..REQ-01-584, REQ-01-610, REQ-01-610A
- **AC-444**: `GET /api/v1/view-schemas/cartulary.view.timeline.v2` and Timeline row-query responses expose exactly the ten visible Timeline v2 fields from Core 01 §7.4.1 in order. Every visible cell in every returned Timeline v2 row is serialized exactly as `{ "value": <string|null> }`, including create-time omitted fields as `null`.
  - Verifies: REQ-01-307, REQ-01-312, REQ-01-614, REQ-03-236..REQ-03-241
- **AC-445**: Row create, row patch, clipboard paste, and import reject any non-string non-null write to a Timeline v2 visible field with `400 error.code='invalid_mutation_payload'`; invalid UTF-8, NUL, and disallowed control characters are rejected; and rejected writes do not mutate source state or projections.
  - Verifies: REQ-01-057..REQ-01-070, REQ-01-312, REQ-01-614
- **AC-446**: Pasting or importing a table whose header row contains the exact ten Timeline v2 visible labels, case-sensitively and without aliases, maps each source column to the corresponding stable `field_key`; no-header paste targets the current v2 field order and stable field keys rather than labels.
  - Verifies: REQ-01-312, REQ-03-236..REQ-03-241
- **AC-447**: Pasting `=HYPERLINK(...)`, HTML, Markdown, URLs, or script-like text into `timeline.raw_activity_text` stores the exact decoded string and renders it as inert text in display and edit modes, with no formula evaluation, HTML/Markdown interpretation, linkification, or script execution.
  - Verifies: REQ-01-312, REQ-01-614, REQ-03-236..REQ-03-241
- **AC-448**: Editing `timeline.mitre_stage_text`, `timeline.device_object_text`, `timeline.ip_address_text`, or `timeline.data_source_text` may create inspector-side suggestions, links, observations, or derived affordances, but the original visible cell string is preserved exactly and no visible Timeline v2 cell serializes a MITRE object, entity ref, indicator object, tag collection, chip model, formula, or formula result.
  - Verifies: REQ-01-312, REQ-01-614, REQ-03-276
- **AC-449**: Incident Timeline time conversion is disabled by default; when enabled, conversion uses only the administrator-configured fixed offset, emits generated UTC as `YYYY-MM-DDTHH:MM:SSZ` and generated local time as `YYYY-MM-DDTHH:MM:SS±HH:MM`, never depends on browser, OS, or viewer time zone, preserves unparseable user text while marking conversion unavailable, and never overwrites a non-empty user-authored paired Activity Date value.
  - Verifies: REQ-01-312, REQ-01-611..REQ-01-614, REQ-03-236..REQ-03-241
- **AC-450**: Non-members, viewers without write permission, and deployment administrators without ordinary incident membership cannot create, patch, paste into, import into, delete, restore, or otherwise mutate Timeline v2 rows; authorization failure happens before row existence, validation-detail, or incident data disclosure beyond the ordinary hidden-incident contract.
  - Verifies: REQ-01-057..REQ-01-080, REQ-04-021..REQ-04-030
- **AC-451**: `GET /api/v1/incidents/{incident_id}/timeline-time-conversion-profile` requires ordinary incident membership; `PUT /api/v1/incidents/{incident_id}/timeline-time-conversion-profile` requires current incident role `admin`; non-members, non-admin members, and deployment administrators without ordinary incident membership cannot read or mutate the resource.
  - Verifies: REQ-01-611..REQ-01-613, REQ-04-021..REQ-04-030
- **AC-452**: Current Timeline tests, fixtures, visual baselines, generated contracts, and generated client artifacts contain no active Timeline v1 assumption that lets a visible Timeline v2 cell return or accept timestamp objects, collections, MITRE/entity/indicator objects, chips, formulas, or formula results.
  - Verifies: REQ-00-014, REQ-01-312, REQ-01-614, REQ-03-236..REQ-03-241
- **AC-453**: Every current-profile `view_schema_resource_v2` emitted by discovery contains a valid `inspector_config_v1` for its own `view_schema_id`; saved views persist no inspector UI state and inherit config from immutable `view_schema_id`; the inspector is closed by default before a newly active surface becomes interactive, renders `no_row_selected` without stale row data, invalidates stale row-bound inspector state, and ordinary grid create/edit/paste works without opening it; after delayed row hydration makes an inspect control actionable, one activation deterministically mounts the matching `(view_schema_id, record_id, row_version)` subject without being superseded by older lifecycle work; inspector-backed reads, mutations, evidence handles, rollback/delete/restore, supersede, merge, mention actions, record creation, and pivots reuse existing route contracts with server-side authorization re-derived from current incident membership and role; `deployment_admin` alone grants no incident inspector access; base-profile inspector behavior performs no external enrichment or third-party egress.
  - Verifies: REQ-00-061, REQ-01-615..REQ-01-617, REQ-02-258, REQ-03-291..REQ-03-292, REQ-04-127
- **AC-454**: Discovery emits exactly the Core 01 registry keys for each required surface and for each implemented optional surface. Unknown, duplicate, missing, or extra inspector feature keys fail conformance validation.
  - Verifies: REQ-01-615..REQ-01-617
- **AC-455**: Each emitted feature group has valid `panel_id`, `route_binding.kind`, `route_binding.owner`, `minimum_incident_role`, `mutates`, `requires_confirmation`, `seed_bindings[]`, `disabled_when[]`, `success_result_behavior`, and `failure_result_behavior`.
  - Verifies: REQ-01-615..REQ-01-617
- **AC-456**: Timeline, Hosts, Identities, Evidence, and Notes support create, inline edit, paste, and correction with the inspector closed.
  - Verifies: REQ-03-291..REQ-03-292
- **AC-457**: Row change, row-version change, incident close, authorization loss, delete, merge, hard refresh, and active surface switch invalidate pending inspector forms, confirmations, previews, merge plans, supersede forms, and rollback previews.
  - Verifies: REQ-03-291..REQ-03-292
- **AC-458**: At least one end-to-end path creates or links from Timeline to Task Request, Decision, Evidence, Communications Log, Handoff, Status Review, and Lesson without leaving the workbook shell.
  - Verifies: REQ-01-617, REQ-03-292
- **AC-459**: Direct route calls fail closed when the displayed feature is hidden, disabled, stale, cached, or shown under outdated role state. `deployment_admin` without incident membership grants no read, mutation, evidence handle, pivot, job, merge, rollback, restore, delete, supersede, or related-record-create access.
  - Verifies: REQ-04-021..REQ-04-030, REQ-04-127
- **AC-460**: Preview and download handles remain same-origin/application-mediated, blocked preview is explicit, and no inspector feature sends incident-derived values or evidence bytes to third parties.
  - Verifies: REQ-01-615..REQ-01-617, REQ-04-127
- **AC-461**: Findings, Investigative Queries, and Forensic Keywords emit feature groups only when the optional surface is implemented; absence of those surfaces does not fail Base Profile conformance.
  - Verifies: REQ-01-615..REQ-01-617
- **AC-462**: Switching saved views over the same `view_schema_id` does not change inspector config; switching to a different immutable `view_schema_id` selects that schema's config; saved views persist no open inspector state, active panel, form state, preview state, confirmation, rollback preview, merge plan, or stale inspector state.
  - Verifies: REQ-01-615..REQ-01-617, REQ-03-291..REQ-03-292
- **AC-310**: `reference_pack_state_conflict` uses only `already_disabled`, `not_disableable`, and `verification_pending`; disabling an already disabled version yields `already_disabled`; disabling a `staged`, `failed`, or `missing` version yields `not_disableable`; reverifying a `staged` version yields `verification_pending`; and activation rejections remain limited to `already_active` or `not_verified_available`.
  - Verifies: REQ-01-471, REQ-01-482

### 9.5 Enterprise Authentication Extension Profile criteria

- **AC-036**: Enterprise-authenticated users map to the same internal user identity used for attribution, and switching from local auth to enterprise auth does not break audit lineage for existing incidents.
  - Verifies: REQ-04-018..REQ-04-020

- **AC-288**: `GET /api/v1/auth/providers` returns only enabled interactive enterprise-auth providers derived from reconciled startup provider configuration, orders them by `display_name asc` then `provider_key asc`, exposes only `provider_key`, `provider_type`, and `display_name`, and rejects pagination members rather than silently ignoring them.
  - Verifies: REQ-01-510..REQ-01-511, REQ-04-018, REQ-04-121
- **AC-289**: `POST /api/v1/auth/providers/{provider_key}/begin` accepts only a JSON object with optional `return_to`; omitted or explicit `null` `return_to` normalizes to `/`; the default `/` target is governed by Core 01 §3.3.2.1A; an off-origin, absolute, or otherwise disallowed `return_to` fails closed with the enterprise-auth request error family; unknown top-level members are rejected; `client_txn_id` is rejected; and replaying the same structurally valid request may mint a fresh redirect target rather than replaying a durable public auth-transaction resource. Conformance MUST also verify that, after successful provider authentication, a provider-authenticated user with the same visible memberships as the local-login fixture in AC-414 reaches the same `/` destination, including the sole-membership invalidation race, and that provider claim content does not change the selected outcome.
  - Verifies: REQ-01-510..REQ-01-511, REQ-01-514..REQ-01-515, REQ-01-580, REQ-04-018
- **AC-290**: A valid OIDC callback issues the same session resource shape exposed by `GET /api/v1/auth/session`, terminates with `303 See Other` to the validated `return_to`, and preserves audit lineage for the linked internal user. When the validated `return_to` resolves to a workbook surface, startup-surface selection inside that workbook uses the Core 03 §2.4 ordered fallback rather than an enterprise-auth-specific order. When the validated `return_to` resolves to `/`, the callback hands off to Core 01 §3.3.2.1A and produces the same destination as local login for the same visible memberships; any sole-incident workbook open supplies no explicit launch `sheet_ref`. Replay, expiry, missing or invalid `state`, missing or invalid `nonce`, failed code exchange, or provider mismatch create no session and fail with the enterprise-auth transaction or provider-response error family.
  - Verifies: REQ-01-031, REQ-01-510, REQ-01-512, REQ-01-515, REQ-01-580, REQ-03-030..REQ-03-031, REQ-04-018, REQ-04-020
- **AC-291**: A valid SAML ACS verifies the provider response, stages only the verified subject and opaque completion-token hash, and redirects to the same-origin ACS completion endpoint. A valid ACS completion verifies the same browser-binding cookie, issues the same session resource shape exposed by `GET /api/v1/auth/session`, terminates with `303 See Other` to the validated `return_to`, and preserves audit lineage for the linked internal user. When the validated `return_to` resolves to a workbook surface, startup-surface selection inside that workbook uses the Core 03 §2.4 ordered fallback rather than an enterprise-auth-specific order. When the validated `return_to` resolves to `/`, completion hands off to Core 01 §3.3.2.1A and produces the same destination as local login for the same visible memberships; any sole-incident workbook open supplies no explicit launch `sheet_ref`. Replay, expiry, `RelayState` mismatch, completion-token mismatch, browser-binding mismatch, issuer mismatch, audience mismatch, signature failure, assertion expiry, or provider mismatch create no session and fail with the enterprise-auth transaction or provider-response error family.
  - Verifies: REQ-01-031, REQ-01-510, REQ-01-512, REQ-01-515, REQ-01-580, REQ-03-030..REQ-03-031, REQ-04-018, REQ-04-020
- **AC-292**: Successful enterprise authentication never auto-creates a local user, never auto-creates an `auth_identity`, never auto-creates incident membership, and never maps provider group claims into incident roles.
  - Verifies: REQ-01-513, REQ-04-019
- **AC-293**: Enterprise-auth protocol routes under `/api/v1/auth/*` use only `invalid_enterprise_auth_request`, `auth_provider_not_found`, `auth_provider_disabled`, `enterprise_auth_transaction_rejected`, `provider_response_rejected`, and `provider_identity_rejected`; their `reason_code` use is limited to the declared enterprise-auth registries; production verifier library, provider, and runtime-configuration diagnostics do not escape those registries; and provider identity binding rejects missing, ambiguous, inactive, or unlinked subjects with the correct family failure.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-513..REQ-01-514, REQ-04-018..REQ-04-020
- **AC-348**: The Enterprise Authentication Extension Profile additionally exposes `POST /api/v1/users/{user_id}/auth-bindings`, `POST /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}/rotate`, and `DELETE /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}`. Binding create targets an existing local `user_id`, accepts required `base_user_version`, `client_txn_id`, `provider_key`, and `provider_subject` plus optional `reason`, allows a reconciled configured OIDC or SAML provider even when that provider is currently disabled for interactive sign-in, returns `201 Created` on first success and `200 OK` on exact replay, and, when exercised against a user who begins with no enterprise bindings, returns a safe user resource whose `auth_bindings[]` contains exactly two active items in canonical order: first one local binding summary containing exactly `provider_type='local'`, `provider_key='local'`, `username` equal to the authoritative safe-user `email`, and `created_at` equal to the safe user resource `created_at`, while omitting `auth_binding_id`, `provider_subject`, and `last_auth_at`; then one enterprise binding summary containing exactly `auth_binding_id`, `provider_key`, `provider_type`, `provider_subject`, `created_at`, and `last_auth_at`, while omitting `username`; and no additional members appear on either item.
  - Verifies: REQ-01-033, REQ-01-116, REQ-01-537..REQ-01-538, REQ-02-234, REQ-02-236..REQ-02-237, REQ-04-093, REQ-04-121
- **AC-349**: Binding create fails closed with `404 error.code='auth_provider_not_found'` for an unknown or non-enterprise `provider_key`, with `409 error.code='auth_binding_conflict'` and `reason_code='provider_subject_in_use'` when the requested subject is already active on a different user, with `409 error.code='auth_binding_conflict'` and `reason_code='provider_already_linked_for_user'` when the addressed user already has one active binding for that provider, and with `409 error.code='client_txn_conflict'` on same-key different-request reuse. Conformance MUST also verify that `provider_subject` comparison is exact and does not trim, case-fold, Unicode-normalize, or email-normalize the supplied or returned value.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-538, REQ-01-541, REQ-02-234..REQ-02-235
- **AC-350**: Binding rotate is atomic retire-plus-create on the same provider and same `user_id`, preserves audit lineage, does not re-key the old binding in place, returns `200 OK` and does not advance `user_version` when `new_provider_subject` exactly equals the current subject, updates `last_auth_at` only on successful callbacks against the active binding, and causes future callback authentication using the old subject to fail with `409 error.code='provider_identity_rejected'` and `reason_code='no_linked_user'`.
  - Verifies: REQ-01-513..REQ-01-514, REQ-01-539..REQ-01-541, REQ-02-234..REQ-02-235, REQ-04-020, REQ-04-095
- **AC-351**: For a user with one active enterprise binding, binding retire removes only that targeted enterprise binding from active callback resolution and from active `auth_bindings[]` summaries without deleting the user or memberships, exact replay returns the original committed success, a fresh request against an already retired binding fails with `409 error.code='auth_binding_conflict'` and `reason_code='binding_not_active'`, the returned safe user resource retains exactly one local binding summary containing exactly `provider_type='local'`, `provider_key='local'`, `username` equal to the authoritative safe-user `email`, and `created_at` equal to the safe user resource `created_at`, while omitting `auth_binding_id`, `provider_subject`, and `last_auth_at`, and local email or password login remains unchanged unless separately changed through the local credential-management routes.
  - Verifies: REQ-01-116, REQ-01-234, REQ-01-238, REQ-01-540..REQ-01-541, REQ-02-234..REQ-02-237, REQ-04-095
- **AC-352**: Only `deployment_admin` can call the three enterprise-auth binding-management routes; incident role `admin` alone cannot. Create, rotate, and retire are auditable deployment-local administrative events; rotate and retire revoke all active sessions for the target user immediately; the current profile uses this audited deployment-admin binding-management surface as the only normative runtime mechanism for binding create, rotate, and retire, and that surface does not mutate provider definitions or provider secrets; the returned safe user resources and `auth_bindings[]` summaries expose only the exact member sets required by REQ-01-115 and REQ-01-116; and deployment-local audit state plus route-scoped idempotency state retain no member of the closed forbidden set in REQ-04-096.
  - Verifies: REQ-01-537..REQ-01-541, REQ-02-236, REQ-04-093..REQ-04-096, REQ-04-122
- **AC-433**: A deployment with `enterprise_authentication.claimed=true` and without `enterprise_authentication.provider_manifest_path` fails before any HTTP listener, WebSocket listener, or background-job runner starts with `invalid_deployment_config` and `reason_code='provider_manifest_path_missing'`; a deployment with `enterprise_authentication.claimed=false` or omitted fails closed when that key is present; and explicit `null`, non-string, empty, relative, `~`, shell-variable, NUL-bearing, or lexical `.`/`..` path values fail with `invalid_deployment_config` and the corresponding path or type reason.
  - Verifies: REQ-00-055, REQ-04-020, REQ-04-077..REQ-04-078, REQ-04-115..REQ-04-116
- **AC-434**: Provider-manifest validation rejects malformed UTF-8, malformed JSON, duplicate JSON object members, unknown top-level members, missing `provider_manifest_schema_id`, any schema id other than `cartulary.enterprise_auth_providers.v1`, omitted or null `providers`, provider arrays longer than `32`, duplicate `provider_key` values, malformed common fields, explicit `null` for optional `enabled` or `additional_scopes`, invalid OIDC URLs, invalid OIDC `client_id`, raw OIDC secrets, unresolved or malformed `secret_ref_v1`, duplicate or more than `16` OIDC additional scopes, OIDC `additional_scopes[]` containing `openid`, invalid SAML `subject_source`, invalid SAML certificate path arrays outside `1..8`, unreadable or non-regular certificate files, and unparseable signing certificates, all before readiness with `invalid_deployment_config`.
  - Verifies: REQ-00-055, REQ-04-077..REQ-04-078, REQ-04-116..REQ-04-120
- **AC-435**: Startup reconciliation creates new configured providers, updates mutable configuration for existing providers whose `provider_type` is unchanged, fails startup when an existing key changes `provider_type`, retains and disables a persisted provider omitted from the manifest, excludes `enabled=false` providers from `GET /api/v1/auth/providers` while preserving eligibility for deployment-admin binding create, and proves that configuration reconciliation creates no user binding, deletes no user binding, rotates no user binding, retires no user binding, deletes no local user, and deletes no incident membership.
  - Verifies: REQ-00-055, REQ-01-511, REQ-01-538, REQ-02-234..REQ-02-237, REQ-02-256, REQ-04-077..REQ-04-078, REQ-04-121..REQ-04-122
- **AC-436**: Binary closure evidence proves that changing enterprise provider definitions, provider metadata, signing certificates, redirect or ACS URL derivation inputs, enabled state, or provider secrets requires a validated deployment restart; no authenticated browser route, public API request, provider metadata upload, SAML certificate upload, OIDC metadata override, secret-management route, incident-portability input, workbook surface, or deployment-admin binding-management route can mutate provider definitions or secrets at runtime; and failed manifest validation or failed reconciliation exposes no partially updated provider set.
  - Verifies: REQ-00-055, REQ-01-510..REQ-01-511, REQ-01-537..REQ-01-541, REQ-02-236, REQ-02-256, REQ-04-020, REQ-04-093, REQ-04-115..REQ-04-122

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

- **AC-048**: The implementation maintains a STRIDE threat model for the current release that covers, at minimum, authenticated sessions, incident records and revisions, evidence blobs and previews, reference packs and import bundles, generated snapshots and exports, and portable runtime roots. When an adopted telemetry subsystem is present, the threat model also covers telemetry exporter endpoints, telemetry headers and secrets, telemetry source snapshots, generated telemetry constants, local diagnostics, retained telemetry artifacts, redaction-before-recording, browser non-export, and telemetry runtime-failure invariance. Each entry maps the threat to at least one control and one verification hook.
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
- **AC-054**: Attempting to mutate, preview, download, or issue object-store access for a `record_id`, `evidence_record_id`, `object_blob_id`, or snapshot outside the caller's incident membership is denied even when the identifier is otherwise valid. Clipboard-paste and explicit bulk-mutation batches reject a `record_id` target outside the addressed incident before row-version or same-field-conflict evaluation and commit no partial batch state.
  - Verifies: REQ-01-355..REQ-01-366, REQ-03-127..REQ-03-128, REQ-04-021..REQ-04-030, REQ-04-052..REQ-04-053
- **AC-055**: A deployment claiming flyaway or disconnected portability for use on portable hosts or removable media stores database, object, reference-pack, temporary-work-file, and export roots on encrypted storage. A deployment that does not do so is non-conformant for flyaway handling.
  - Verifies: REQ-01-455..REQ-01-456, REQ-04-052..REQ-04-055, REQ-04-058..REQ-04-059

### 9.8 Additional Base Profile criteria for promoted recurrent fields

- **AC-097**: The Hosts sheet can sort or filter on `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` from the grid surface without opening the inspector and without rereading unrelated note text or blob metadata.
  - Verifies: REQ-01-323..REQ-01-325, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246
- **AC-098**: The Identities sheet can sort or filter on `privilege_level`, `mfa_state`, and `reset_status` from the grid surface without rereading unrelated note text or evidence blobs.
  - Verifies: REQ-01-326..REQ-01-327, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246
- **AC-099**: On the implementation-supported incident metadata surface, editing `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref` after incident creation persists those values as structured incident state that survives reload, deterministic projection rebuild, and snapshot generation; they do not disappear into `custom_attrs` or extension-only JSON.
  - Verifies: REQ-01-157..REQ-01-158, REQ-01-173..REQ-01-175, REQ-01-491..REQ-01-491.1, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246
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
| `LF-01` | `timeline.capture_state` | `base` | A Timeline row is visible and persisted with `capture_state='superseded'`. | An ordinary Timeline patch that changes `timeline.activity_synopsis_text` fails with `409 error.code='illegal_transition'`. | The row remains `capture_state='superseded'`; `timeline.activity_synopsis_text` is unchanged; no new `change_set` commits. | `POST /api/v1/records/{record_id}/mark-reviewed` fails with `409 error.code='illegal_transition'`. |
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
- **AC-155**: Pending-blob timeout handling runs at least every 15 minutes; by no later than `pending_expires_at + 15 minutes`, an unfinalized slot is failed; and for any failed unattached slot, including timeout, retry exhaustion, or terminal size/hash mismatch, orphaned bytes are deleted within 1 hour of terminal failure and failed unattached slot metadata remains queryable for at least 7 days before automatic hard deletion. Service-backed fixtures prove batches bounded to 100, `SKIP LOCKED` multi-instance exclusion, restart recovery after a 5-minute lease expires, deletion outside the claim transaction under a 1-minute timeout, idempotent typed not-found completion, retry delays of 1, 5, and then repeated 15 minutes without exhaustion, claim-token and current-state completion guards, rejection of attachment/lifecycle/import/restore races, exact 45-minute eligibility, and 7-day metadata retention. A partial deployment with no private dispatcher performs no deletion. Application-lifecycle fixtures prove the private `evidence.failed_unattached_blob_cleanup.v1` dispatcher starts only after serving readiness, performs one immediate post-readiness sweep and then one sweep every 15 minutes, tolerates concurrent application instances through the durable claim boundary, stops cleanly, uses only typed object-store purpose `evidence_cleanup`, and creates no public or common Jobs surface. Telemetry conformance proves the cleanup counter, duration, overdue count, and oldest-eligible-age metrics use only closed operation/result/error-class attributes and that metrics or logs expose no incident, record, blob, object-key, filename, hash, capability, or raw-error value.
  - Verifies: REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-126
- **AC-543**: Table-driven, service-backed Evidence authorization evidence covers every combination required by REQ-04-156 across missing, expired, revoked, and valid sessions; missing, invalid, and valid cookie CSRF; no membership and every current incident role; `deployment_admin` false and true; same, foreign, and closed incidents; malformed, unknown, cross-session, cross-incident, expired, consumed, and valid capabilities; active, deleted, and restored Evidence; absent, pending, available, failed, quarantined, missing, and inconsistent blobs; successful, object-not-found, and typed transient object-store results; and role or membership changes between issuance and use. Every denial cell asserts the exact Core 01 status, `error.code`, safe `reason_code` or its required absence, concealment posture, object-store invocation count, handle-consumption result, and absence of durable effects. The suite proves the exact Table 2-A role sets, Table 2-B precedence, Table 2-C state-change consequences, current authorization at use time, and the insufficiency of `deployment_admin` without current incident membership.
  - Verifies: REQ-04-156, REQ-01-243..REQ-01-247, REQ-01-458..REQ-01-465
- **AC-104**: Rendering a report or presentation artifact creates a release record bound to one immutable release tuple, including materialized `output_options`, canonical `recipient_partition_refs[]`, canonical `graph_projection_refs[]`, all-null or all-non-null composition tuple fields, `render_admitted_at`, `redaction_profile_sha256`, `redaction_manifest_sha256`, and `output_sha256`; the record starts in `pending_approval` when approvals are required and starts in `approved` only for an internal-draft scope with no separate approval requirement.
  - Verifies: REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035
- **AC-105**: Satisfying the required approval set moves that exact release record to `approved`, and any attempted publish action on a non-approved record is rejected.
  - Verifies: REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035
- **AC-106**: Rendering a superseding artifact for the same logical output slot with a different bound tuple, different composition tuple, different graph projection binding, different `redaction_profile_sha256`, different `redaction_manifest_sha256`, or different `output_sha256` creates a distinct candidate and invalidates the prior current artifact rather than inheriting its approval or publication state.
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
- **AC-314**: For a `decision` record, a fixture whose persisted `status` and authoritative decision-to-decision `record_links` relation using `link_type='supersedes'`, `owner_user_id`, and `decided_at` do not resolve to exactly one legal machine condition is treated as inconsistent; ordinary Decision grid mutations, including direct status writes, non-status scalar writes, and relationship collection writes, plus ordinary explicit supersession actions fail closed while the inconsistency remains; and replay of the same repaired fixture is deterministic.
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
- **AC-414**: After successful local authentication, default browser navigation to `/` evaluates only the caller's current visible incident collection: with zero visible incidents it remains on `/`, renders the empty visible-incident directory, and exposes the ordinary create-incident affordance to an active authenticated account; with exactly one visible incident it opens that incident workbook without an explicit launch `sheet_ref` and then uses the Core 03 §2.4 startup fallback; with two or more visible incidents it remains on `/` and renders the visible-incident directory. If the sole visible incident loses visibility before workbook bootstrap completes, the browser returns to `/` and renders the current visible-incident directory. The same fixtures MUST prove that recency, sort order, prior visit state, client cache, and `deployment_admin` status do not select or widen incidents. A deployment administrator fixture with exactly one visible incident MUST still follow the sole-incident branch and then be able to reach `/deployment-administration` through the global administration entry. When the Enterprise Authentication Extension Profile is claimed, AC-289 through AC-291 additionally prove that provider claim content and provider authentication do not change the outcome for the same visible memberships.
  - Verifies: REQ-00-053, REQ-00-057, REQ-01-025, REQ-01-168, REQ-01-580, REQ-01-608, REQ-03-030..REQ-03-031, REQ-04-028..REQ-04-029
- **AC-427**: A current-profile route-inventory audit classifies every public route family declared by Core 01 §3.3.3, §17, and §20, the Deployment administration browser context, plus every deployment-local operator family called out by REQ-04-028, as exactly one of: `deployment_admin required`, `deployment_admin insufficient without incident authorization`, or `deployment_admin irrelevant`. The audit MUST match the matrix in REQ-04-028, MUST prove that no current public route or browser context relies on an undeclared granular deployment capability, MUST prove that `deployment_admin` alone cannot disclose incident content without current incident membership, MUST prove that the current profile exposes no deployment-wide all-incident catalog and no generic deployment-settings API, MUST prove that the recovery CLI is `deployment_admin irrelevant` because local operator authorization applies, MUST prove that reserved but unclaimed extension-family paths return `404 error.code='extension_profile_not_claimed'` before family-specific authorization or policy evaluation, MUST prove that a claimed deployment returns `403 error.code='authorization_denied'` to an authenticated non-admin for every Core 01 §17.4 reference-pack route before route-specific path, query, body, resource, idempotency, or job-admission behavior, MUST prove that a claimed Network Flow Activity route family never treats `deployment_admin` alone as incident access and enforces the role matrix in REQ-04-105A, and MUST prove that no authenticated public route mutates enterprise-auth provider definitions or secrets at runtime.
  - Verifies: REQ-00-057, REQ-01-032..REQ-01-033, REQ-01-471, REQ-01-480..REQ-01-481, REQ-01-542..REQ-01-548, REQ-01-608, REQ-04-023, REQ-04-028..REQ-04-030, REQ-04-085, REQ-04-094, REQ-04-105, REQ-04-105A, REQ-04-106, REQ-04-114, REQ-04-122, REQ-04-126
- **AC-429**: `GET /api/v1/account/profile` returns only the current authenticated user's `account_profile` resource with exactly `user_id`, `email`, `display_name`, `user_version`, `created_at`, and `updated_at`; rejects `limit`, `cursor_token`, and pagination aliases with `400 error.code='invalid_pagination_request'`; and exposes no password, TOTP, locale, time-zone, notification, theme, density, row-height, incident-membership, admin-only, or alternate login-identifier state. `PATCH /api/v1/account/profile` accepts only required non-null `base_user_version`, `client_txn_id`, and `display_name`; validates `base_user_version` as a positive integer and `display_name` under `display_name_line_v1`; rejects omitted, null, unknown, forbidden, or malformed members with `400 error.code='invalid_mutation_payload'`; leaves `email` read-only; and returns the resulting `account_profile` on success.
  - Verifies: REQ-00-054, REQ-01-032, REQ-01-597..REQ-01-599, REQ-02-255, REQ-04-114
- **AC-430**: `PATCH /api/v1/account/profile` evaluates idempotency and concurrency in the Core 01 order: exact committed replay returns the original committed result before fresh version evaluation; reuse of the same route-scoped key with different normalized input fails with `409 error.code='client_txn_conflict'`; stale `base_user_version` fails with `409 error.code='user_version_conflict'`; and a normalized no-op returns the current profile without advancing `user_version` or `updated_at`. A material display-name change advances `user_version` and `updated_at` exactly once, does not revoke active sessions, appears on later session reads and newly emitted presence payloads, and does not rewrite already emitted payloads, history, or audit snapshots.
  - Verifies: REQ-01-599, REQ-04-016, REQ-04-038, REQ-04-114
- **AC-431**: `GET /api/v1/account/preferences` returns only the current authenticated user's `account_preferences` resource with exactly `user_id`, nullable `density_mode`, `preferences_version`, `created_at`, and `updated_at`, and every existing or new user has one resource initialized to `density_mode=null` and `preferences_version=1`. `PUT /api/v1/account/preferences` accepts only required non-null `base_preferences_version`, required non-null `client_txn_id`, and required `density_mode`, where `density_mode` may be JSON `null` or exactly `compact`, `default`, or `comfortable`; omitted `density_mode`, custom tokens, custom row heights, theme values, locale, time-zone, notification settings, profile fields, and unknown members fail with `400 error.code='invalid_mutation_payload'`. Exact committed replay returns the original result, changed replay fails with `client_txn_conflict`, stale `base_preferences_version` fails with `409 error.code='preferences_version_conflict'`, no-op returns the current resource without advancing `preferences_version` or `updated_at`, and material success advances both exactly once without revoking sessions. Effective density is `compact` for Timeline when `density_mode=null`, `default` for every other workbook surface when `density_mode=null`, and the exact persisted token when non-null.
  - Verifies: REQ-01-600..REQ-01-601, REQ-02-255, REQ-03-289, REQ-04-016, REQ-04-114
- **AC-432**: Binary closure evidence proves that an ordinary authenticated user can change only their own display name and declared density override through `/api/v1/account/*`, cannot list, read, create, patch, reset, or revoke users through `/api/v1/users*`, and gains no deployment-user administration, email or login-identifier change, locale, time-zone, notification, theme-selection, global default incident, global `home_sheet_ref`, custom-density, or row-height capability. The same evidence proves that an admin still uses admin routes for email or login-identifier changes; password and TOTP behavior remains only under existing `/api/v1/auth/*` or deployment-admin user-action routes; account preferences are deployment-local normalized user state, not incident records or per-incident workbook preferences; incident portability excludes account preferences and imports do not synthesize them from bundles; and deployment-local audit/idempotency records for these routes remain outside incident portability.
  - Verifies: REQ-00-054, REQ-01-597..REQ-01-602, REQ-02-204, REQ-02-249, REQ-02-255, REQ-03-289, REQ-04-028, REQ-04-038, REQ-04-114
- **AC-437**: Administrative audit route conformance proves that the Base Profile exposes `GET /api/v1/administrative-audit-events` and `GET /api/v1/incidents/{incident_id}/membership-audit-events`; both routes return the common success envelope with `data.audit_events[]` and `meta.paging`; each event and each `changes[]` item contains exactly the member set owned by Core 01 §3.3.5.1A; closed `scope_kind`, `actor_kind`, `source`, and `value_state` vocabularies are enforced; deployment and incident `scope_id` nullability is exact; `changes[]` is nonempty, sorted by exact `field_path asc`, and has no duplicates; current action codes and target kinds follow the registry and mapping table; current servers emit only registered current codes; and clients tolerate additive future action codes without failing response parsing.
  - Verifies: REQ-00-056, REQ-01-032, REQ-01-603..REQ-01-606, REQ-02-257, REQ-04-124..REQ-04-125
- **AC-438**: Administrative audit list-query conformance proves ordering by `occurred_at desc, audit_event_id desc`; omitted `limit` defaults to `100`; `limit` is bounded to `1..500`; only `limit`, `cursor_token`, `actor_user_id`, `action_code`, `target_kind`, `target_id`, `occurred_at_gte`, and `occurred_at_lt` are accepted as query members; omitted filters mean no predicate; exact filters reject malformed, empty, repeated, comma-list, array, explicit-`null`, or out-of-registry values as specified; `search` and other unknown query members are rejected; duplicate raw query members fail before unknown-member validation with `duplicate_query_member`; timestamp filters parse and normalize under `timestamp_instant_v1`; `occurred_at_gte` is inclusive, `occurred_at_lt` is exclusive, and both present require `gte < lt`; cursor replay with a different actor, route, scope, normalized filter set, effective limit, ordering tuple, or continuation contract fails with `cursor_query_mismatch` or the most specific pagination reason; and authorization failure is evaluated before filter, count, and cursor-position disclosure.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242, REQ-01-583..REQ-01-584, REQ-01-607, REQ-04-123
- **AC-439**: Administrative audit authorization conformance proves the binary closure criterion: a current `deployment_admin` can inspect deployment/account/recovery administrative audit through `GET /api/v1/administrative-audit-events` without receiving incident membership events for incidents where they lack ordinary membership, while an incident `admin` can inspect membership changes only for that addressed incident through `GET /api/v1/incidents/{incident_id}/membership-audit-events`. An authenticated non-deployment-admin receives `403 authorization_denied` for the deployment audit route; a deployment admin without current membership in the addressed incident receives the ordinary hidden-incident `404` for the membership-audit route; a non-member receives the same hidden-incident `404`; and a visible non-admin incident member receives `403 authorization_denied`.
  - Verifies: REQ-00-056, REQ-01-603, REQ-02-257, REQ-04-028..REQ-04-030, REQ-04-123, REQ-04-125
- **AC-440**: Administrative audit retention and redaction conformance proves that neither stored nor returned administrative audit values contain current, new, initial, or reset passwords, password hashes, TOTP secrets or `otpauth_uri`, bootstrap or session tokens, provider assertions or provider tokens, raw SAML responses, ID tokens, access tokens, recovery keys, object-store credentials, raw DSNs, object keys, or storage secrets; redacted changed fields preserve field identity with `value_state='redacted'` and JSON `null` before/after values; public projections contain nonempty changes and exact current action/scope/target bindings; audit events are immutable; no current public delete or purge route exists; retention lasts for the deployment lifetime; operational backup includes raw and projected administrative audit state; whole-incident portability excludes deployment-local administrative audit state; and historical raw rows without an exact safe current projection remain recovery/forensic data rather than ambiguous public events.
  - Verifies: REQ-00-056, REQ-01-432, REQ-01-571, REQ-01-604..REQ-01-606, REQ-02-202, REQ-02-204, REQ-02-249, REQ-02-257, REQ-04-038, REQ-04-086, REQ-04-096, REQ-04-124..REQ-04-125
- **AC-441**: Deployment administration conformance proves that `/deployment-administration` is a distinct application context with canonical label `Deployment administration`, default panel `Users`, and a global shell entry reachable from both the visible-incident directory and an opened workbook; the entry is omitted for sessions without current `deployment_admin`; direct authenticated non-admin navigation returns no administrative data and navigates to `/`; losing `deployment_admin` while the context is open clears loaded administrative resources, rejects or cancels pending administrative requests where possible, and navigates to `/`; browser cache or history cannot preserve administrative access; allowed panels are limited to Users, Administrative audit, claimed Reference packs, claimed Incident import, and claimed Enterprise authentication bindings; and the UI exposes no all-incident catalog, incident metadata for incidents hidden by ordinary membership, generic General settings panel, provider-definition editor, browser recovery control, or incident membership controls authorized solely by `deployment_admin`.
  - Verifies: REQ-00-057, REQ-01-608, REQ-03-290, REQ-04-028..REQ-04-029, REQ-04-126
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
- **AC-153**: `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` read or replace only the caller's nullable `home_sheet_ref`; the `PUT` route accepts only `{ "home_sheet_ref": <sheet_ref|null> }`, creates the preference object if absent, replaces only that pointer if present, rejects unknown top-level members with `400 error.code='invalid_mutation_payload'`, succeeds for any incident member including `viewer`, and a structurally valid no-op leaves `updated_at` unchanged; `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` read or replace only the incident's nullable `default_sheet_ref`; the `PUT` route accepts only `{ "default_sheet_ref": <sheet_ref|null> }`, creates the preference object if absent, replaces only that pointer if present, rejects unknown top-level members with `400 error.code='invalid_mutation_payload'`, fails closed for non-admin incident roles, and a structurally valid no-op leaves `updated_at` and `updated_by_user_id` unchanged. Automatic startup repair conditionally clears only the exact pointer proven invalid; an effective clear advances `updated_at` once, an effective incident-default clear sets `updated_by_user_id` to the authenticated repair-triggering caller, and a comparison miss, already-null pointer, or other no-op preserves timestamp and attribution. Automatic repair grants no non-admin capability to set or replace the incident default. When the selected surface is any pack-independent base-profile surface from `REQ-01-307`, including `cartulary.view.task_requests.v1`, `cartulary.view.decisions.v1`, `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`, `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`, the persisted pointer uses only `sheet_ref.kind='view_schema'` with the standardized `view_schema_id` for the required base surface itself; a user-created saved view over the same schema persists distinctly as `sheet_ref.kind='saved_view'` with its own `saved_view_id`.
  - Verifies: REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032
- **AC-126**: Same-field conflict responses use the generic error envelope with `error.code='same_field_conflict'` and the conflict object required by Core 03 §3.3.4; a stale `conflict_token` is rejected with a fresh conflict payload rather than silently overwriting saved state.
  - Verifies: REQ-01-019, REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-03-036..REQ-03-040,
    REQ-03-063..REQ-03-076
- **AC-483**: Conflict-resolution exact replay by the same actor for the same record, `client_txn_id`, and normalized resolution returns the original successful result without duplicate source, revision, projection, idempotency-success, or Collaboration effects; reuse of that key with different normalized content returns `409 client_txn_conflict`; and reuse of the same conflict token with a new key repeats current authorization and source conflict-window validation, returning a fresh `409 same_field_conflict` when the valid token is now stale.
  - Verifies: REQ-01-019, REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-03-063..REQ-03-076
- **AC-479**: For `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve`, authentication and cookie CSRF checks dominate path, token, and body failures; path-record visibility and incident role `editor` dominate token and body failures; and token verification dominates body validation. The complete outcome matrix is `401 session_required` for a missing, invalid, or inactive session; `403 csrf_verification_failed` for missing or invalid CSRF; hidden `404 incident_not_found` for a missing record or non-member; `403 authorization_denied` for a visible record and role below `editor`; `400 invalid_mutation_payload` with `details.field='conflict_token'` for an authorized caller with a malformed, tampered, unsupported-route, path-mismatched, cross-record, or cross-incident token; the route-owned `400` for an invalid body after a valid token; `409 same_field_conflict` with a fresh conflict object for a valid stale token; and the existing success contract for a valid current token. Timeline and non-Timeline owners use this same guard. Every rejected matrix case creates no source mutation, change set, revision, projection update, idempotency commit, or collaboration event.
  - Verifies: REQ-01-019, REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-03-063..REQ-03-076, REQ-04-021..REQ-04-030
- **AC-480**: Saved-view selection, startup, create, update, duplicate, dirty comparison, and reset round-trip the exact `cartulary.layout.v1` semantic column permutation, hidden set, sparse widths, and ordered sort list; the live grid applies them; all-columns-hidden recovery is keyboard accessible; widths remain within `40..4096`; and no structural grid column or vendor coordinate enters persistence. Ordinary sort activation replaces while Ctrl/Cmd activation adds, cycles, or removes and the eight-entry maximum fails closed before query dispatch.
  - Verifies: REQ-01-627, REQ-03-295..REQ-03-296
- **AC-481**: Timeline command-gated bulk selection uses committed record IDs only, excludes group and draft rows, distinguishes active cell and inspector subject, supports pointer, keyboard, shift-range, and current-page select-all, prunes results or authorization losses, and dispatches one `multi_row_tag_assignment_v1` request with explicit current row versions. Non-adopting surfaces expose no selection control.
  - Verifies: REQ-01-628, REQ-03-297
- **AC-482**: Every discovered field exposes required `grid_editable`; owner-permitted direct fields have a typed editor and all other fields have an explicit non-editor reason. Create-only Indicators and append-only Assessments remain non-editable. Invalid local text, stale refresh, server validation, same-field conflict, read-only transition, authorization denial, and focus restoration preserve the specified semantic anchor and draft behavior. Writable user-ID fields accept only exact active same-incident member user IDs and reject labels, email addresses, inactive users, and non-members.
  - Verifies: REQ-01-627..REQ-01-628, REQ-03-298
- **AC-483**: Clipboard and fill operate on semantic rectangular ranges, contract-aware TSV values, visible field keys, and explicit versioned record targets; incompatible, hidden, non-grid-editable, grouped-fill, draft, stale, read-only, and presentation targets reject before any partial dispatch. `fill_down_v1` is limited to one compatible grid-editable non-collection field in ungrouped mode.
  - Verifies: REQ-03-297..REQ-03-299
- **AC-484**: Each closed grid data state and interaction mode has exact message posture, action availability, ARIA behavior, authorized-row and draft preservation rules, and mutation permissions. Permission loss clears protected rows; refreshing and stale errors retain prior authorized rows; closed incidents display `Closed, read-only`, permit read and copy, and dispatch no editor, paste, fill, create, or bulk mutation. State presentation uses no fake records.
  - Verifies: REQ-03-299
- **AC-485**: From an unselected committed writable scalar cell, one primary click creates and focuses exactly one contract-backed editor with a collapsed caret immediately after its existing text and no selected text. Read-only cells remain selectable without an editor, embedded controls execute only their declared action, and double-click creates no additional transition. Moving to another cell dispatches at most one commit and opens the destination only after acceptance; validation, conflict, stale-target, authorization, and other rejection outcomes retain the exact draft and original focusable editor. Navigation-mode committed scalar cells contain no editable form control. The labeled accent-token fill handle appears only for an eligible selected cell, double-click dispatches no fill, and pointer drag and `Ctrl/Cmd+D` produce the same explicit stable-ID targets through `fill_down_v1`.
  - Verifies: REQ-03-220..REQ-03-222, REQ-03-300
- **AC-486**: Browser-authored mutation identifiers are cryptographically collision-resistant across remounts and client instances, while ambiguous transport replay retains the original identifier. A definitive `client_txn_conflict` leaves no new authoritative side effect and exposes the same-surface `Retry with a new request ID` and `Discard blocked edit` actions. Retry changes only request identity, preserves FIFO intent and later work, rematerializes the latest committed row version, and routes a resulting structured `same_field_conflict` into the ordinary resolver. Discard sends no mutation, removes exactly the blocker, restores committed state plus later queued intent, and resumes FIFO work. Recovery controls and the actionable `Conflict` entry point are keyboard operable, do not steal focus on appearance, disable during transition, and disclose no raw transaction or server internals.
  - Verifies: REQ-01-070, REQ-03-301..REQ-03-302
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
- **AC-224**: The same route with a stale `base_mention_row_version` fails with `409 error.code='row_version_conflict'` and `error.details` containing `entity_mention_id`, `base_mention_row_version`, `current_mention_row_version`, and `source_record_id`; a direct `resolve_item` against a currently dismissed mention fails with `409 error.code='illegal_transition'`; a request against a soft-deleted source record fails with `409 error.code='record_deleted_use_restore'`; a caller lacking visibility to the mention receives `404 error.code='entity_mention_not_found'`; a caller who can see the mention but lacks `editor`, `reviewer`, or `admin` role receives `403`; a supplied `resolved_record_id` that does not identify a visible active target fails with `404 error.code='resolved_record_not_found'`; and missing required members, unknown top-level members, forbidden or missing `resolved_record_id`, wrong target type, target from another incident, or embedded entity-create payload fail with `400 error.code='invalid_mutation_payload'`. Authentication and state-changing CSRF dominate malformed path and body failures; malformed or hidden mention identity dominates role and body failures; visible insufficient role dominates body validation; all failures use the common error envelope and commit no source, history, projection, idempotency, or collaboration state.
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
- **AC-207**: Sending client-chosen `link_type`, direction flags, table names, storage-routing metadata, a storage `field_key` override inside an action payload, or link-local `note`, `description`, `comment`, or equivalent narrative in a base-profile relationship mutation fails with `400` and `error.code='invalid_mutation_payload'`; the server accepts only the ordinary enclosing `changes[].field_key` plus field-key-derived or action-route-derived relationship routing, and creates no partial relationship effect.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209
- **AC-208**: Repeating the same logical relationship add through supported base-profile routes, including duplicate collection actions, idempotent request replay, or a later add of the same `(incident_id, src_record_id, dst_record_id, link_type, field_key)` tuple using null-safe `field_key` equality, leaves exactly one non-deleted `record_link` for that tuple while preserving attributable history of the attempted operations. Adding the same target and link type through two different non-null canonical field keys leaves two active bindings, and removing either field's opaque item reference tombstones only that field's binding.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209
- **AC-209**: Merging a host or identity whose incoming or outgoing links collide with links already present on the survivor repoints both incoming and outgoing active links in the same `change_set`, preserves canonical direction and exact nullable `field_key`, deterministically retains at most one non-deleted row for each full `(incident_id, src_record_id, dst_record_id, link_type, field_key)` tuple using null-safe equality, and preserves otherwise-identical bindings owned by different non-null field keys.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-181..REQ-01-195, REQ-02-054..REQ-02-055, REQ-02-064..REQ-02-066,
    REQ-02-163..REQ-02-185, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249
- **AC-210**: Projection-backed linked-count fields, visible linked-record chips, and any operator-visible current-state export or report field that derives relationships from `record_links` include only links whose own row is not soft-deleted and whose source and destination records are not soft-deleted; the same inactive links remain visible through history or rollback surfaces.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-342..REQ-01-348, REQ-01-351..REQ-01-353, REQ-01-355..REQ-01-366,
    REQ-02-163..REQ-02-185, REQ-03-247..REQ-03-249
- **AC-394**: A base-profile manual Timeline relationship mutation using either `add_resolved_ref` or `resolve_item` on `timeline.host_refs` or `timeline.identity_refs` commits without any client-supplied `confidence` member and produces exactly one active `record_link` whose `link_type` is field-derived, whose `provenance` is `manual`, and whose `confidence` is `null`.
  - Verifies: REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-163..REQ-02-180, REQ-02-248, REQ-03-280
- **AC-395**: A supported base-profile collection-style relationship mutation using `add_record_ref` during assessment row creation for `assessment.support_refs`, and using the owner-declared create or existing-row collection path for `task.linked_record_ids`, `decision.support_refs`, `decision.affected_record_ids`, and at least one coordination-surface `record_ref` field such as `comm_log.decision_ids` or `handoff.open_task_ids`, commits without any client-supplied `confidence` member and creates or preserves active `record_links` whose `provenance` is `manual` and whose `confidence` is `null`. This criterion does not admit an existing-row assessment support mutation.
  - Verifies: REQ-01-311, REQ-01-333..REQ-01-340, REQ-01-503..REQ-01-506, REQ-02-163..REQ-02-180, REQ-02-248, REQ-03-280
- **AC-396**: Supplying a client-supplied `confidence` member on a base-profile relationship mutation route or relationship action payload, including assessment row creation for `assessment.support_refs`, `timeline.host_refs`, `timeline.identity_refs`, `task.linked_record_ids`, `decision.support_refs`, `decision.affected_record_ids`, and reused coordination-surface `record_ref` fields, fails with `400` and `error.code='invalid_mutation_payload'`; the server creates no durable `record_link`, persists no client `confidence` value, and leaves no misleading projection update.
  - Verifies: REQ-01-569, REQ-03-280
- **AC-397**: Authoritative stored manual-link state preserves `confidence=null`. Implementations MUST NOT normalize manual-link `confidence` to `0`, `100`, or any other sentinel in current-state projection reads or ordinary export, and when the corresponding extension profiles are claimed, snapshot-source derivation and incident-portability serialization MUST preserve the same `null` value.
  - Verifies: REQ-02-248
- **AC-127**: Potentially large list or view-query routes use opaque cursor pagination with `has_more` and `next_cursor`; replaying a cursor against a different authenticated actor, route family, route-scoping identifier, normalized list search/filter state when present, normalized view-query sort/filter/grouping contract when present, or effective `limit` is rejected rather than reinterpreted.
  - Verifies: REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-117..REQ-01-118, REQ-01-129, REQ-01-144, REQ-01-168, REQ-01-240..REQ-01-242, REQ-01-288, REQ-01-584
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
- **AC-415**: Shared `list_search_v1` conformance evidence proves that omitted `search` and a present search value that normalizes to the empty string produce and bind the same no-search predicate; invalid UTF-8, C0/C1 controls, more than `256` Unicode scalars after normalization, and a non-empty value with zero tokens fail with `400 error.code='invalid_list_query'` and `error.details.reason_code='invalid_search'`; more than `16` unique normalized tokens fails with `reason_code='search_token_count_exceeded'`; duplicate raw query members fail with `reason_code='duplicate_query_member'`; unknown non-pagination query members fail with `reason_code='unknown_query_member'`; case folding is locale-independent; diacritics remain significant; search token order and duplicate tokens are non-semantic; unsupported phrase, fuzzy, wildcard, stemming, transliteration, regex, and storage-engine syntax do not affect matching; and search never changes the owner route's declared ordering.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-240, REQ-01-581..REQ-01-583
- **AC-416**: `GET /api/v1/incidents` accepts only `limit`, `cursor_token`, `search`, and `status`; `status` accepts only the exact Core 02 `incident.status` tokens `active` or `closed`; invalid status values fail with `400 error.code='invalid_list_query'` and `error.details.reason_code='invalid_filter_value'`; a fixture where unfiltered `limit=1` would place a matching authorized incident on a later cursor page returns that incident on page one for `search=<matching-prefix>&limit=1`; `status` and `search` combine without relevance ranking; matching incidents hidden by missing current membership do not affect `data.incidents[]`, counts, `has_more`, `next_cursor`, or continuation position; and reusing a non-null incident-list cursor with a different canonical `search` or `status` fails deterministically with `400 error.code='invalid_pagination_request'` and `error.details.reason_code='cursor_query_mismatch'`.
  - Verifies: REQ-01-168, REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242, REQ-01-581..REQ-01-584, REQ-02-222..REQ-02-223
- **AC-417**: `GET /api/v1/users` remains callable only by a deployment administrator and accepts only `limit`, `cursor_token`, `search`, `is_active`, and `is_deployment_admin`; the two boolean filters accept only exact `true` or `false` wire tokens and invalid values fail with `400 error.code='invalid_list_query'` and `error.details.reason_code='invalid_filter_value'`; a fixture where unfiltered `limit=1` would place a matching user on a later cursor page returns that user on page one for `search=<matching-prefix>&limit=1`; search and the two filters combine without relevance ranking; non-admin callers cannot use search or filters to infer user-list contents; and reusing a non-null user-list cursor with a different canonical `search`, `is_active`, or `is_deployment_admin` fails deterministically with `400 error.code='invalid_pagination_request'` and `error.details.reason_code='cursor_query_mismatch'`.
  - Verifies: REQ-01-117, REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242, REQ-01-581..REQ-01-584, REQ-04-038
- **AC-418**: The base route inventory exposes `POST /api/v1/incidents/{incident_id}/close` and `POST /api/v1/incidents/{incident_id}/reopen`; each route accepts only a JSON object containing exactly `base_incident_version`, `client_txn_id`, and required `reason`; omitted, `null`, normalized-empty, longer than 4096 Unicode scalar values after normalization, rejected-control, or non-string `reason` fails with `400 error.code='invalid_incident_lifecycle_request'`; and the current profile exposes no archive, hard-delete, soft-delete, purge, tombstone, or equivalent whole-incident removal route or lifecycle token.
  - Verifies: REQ-00-035, REQ-01-032, REQ-01-585..REQ-01-586, REQ-01-234, REQ-01-238, REQ-01-496
- **AC-419**: Closing an `active` incident as a current incident `admin` with a matching `base_incident_version` returns `200 OK` with the resulting incident resource; sets `status='closed'`; sets non-null `closed_at` equal to the same commit timestamp used for `updated_at`; increments `incident_version`; sets `updated_by_user_id` to the actor; writes attributed before/after audit history with a nonzero effect identity; and, only after that fresh commit succeeds, sends every currently connected incident WebSocket a terminal `error` with `code='incident_closed'` and `retryable=false` before closing it. No socket effect occurs before commit, on rollback or commit failure, for a rejected/no-op close, or for exact idempotency replay; other incidents' sockets are unchanged. Under the single-active-process contract, process-local socket write failure follows existing cleanup and telemetry without changing the committed `200` result.
  - Verifies: REQ-01-587..REQ-01-589, REQ-01-592
- **AC-420**: Reopening a `closed` incident as a current incident `admin` with a matching `base_incident_version` returns `200 OK` with the resulting incident resource; sets `status='active'`; serializes `closed_at=null`; increments `incident_version`; sets `updated_at` and `updated_by_user_id`; and writes attributed before/after audit history.
  - Verifies: REQ-01-587..REQ-01-589
- **AC-421**: A fresh `close` against an already `closed` incident fails with `409 error.code='illegal_transition'` and `reason_code='incident_already_closed'`; a fresh `reopen` against an `active` incident fails with `409 error.code='illegal_transition'` and `reason_code='incident_not_closed'`; a stale fresh lifecycle request fails with `409 error.code='incident_version_conflict'`; exact committed replay under `(actor_user_id, incident_id, action_route, client_txn_id)` returns the original success before fresh version or transition checks; and reusing that idempotency key with different normalized `{base_incident_version, reason}` fails with `409 error.code='client_txn_conflict'`.
  - Verifies: REQ-01-587..REQ-01-589, REQ-01-234, REQ-01-238
- **AC-422**: For a `closed` incident, incident list/get, workbook queries, record history, evidence preview/download, saved views, workbook preferences, incident membership administration, extension-authorized snapshot/report/release/export, exact committed mutation replay, and `reopen` remain allowed under their ordinary authorization rules, while fresh incident metadata patch, row creation, record mutation, delete, restore, rollback, merge, supersede, conflict resolution, mention resolution, blob-slot creation, evidence attachment, import apply, and any job commit of authoritative incident source state fail or terminally fail with `409 error.code='incident_closed'`.
  - Verifies: REQ-01-590, REQ-01-234
- **AC-423**: In a race between incident closure and any authoritative incident source-state mutation, conformance evidence shows the binary outcome: either the source mutation commits before close and close observes that committed state, or close commits first and the fresh mutation fails or terminally fails with `incident_closed`; after close commits, no non-replay source mutation creates a new change set, row revision, source-state update, evidence attachment, import apply result, or collaboration event until a successful reopen.
  - Verifies: REQ-01-591
- **AC-424**: A closed incident remains visible/readable to current members in default incident listing and workbook open flows; the workbook renders a persistent `Closed, read-only` state; source-state write affordances are disabled or hidden while allowed read/report actions remain available; and queued or unsent source mutations rejected with `incident_closed` become non-authoritative local drafts that may remain copyable but do not auto-replay while closed or after a later reopen.
  - Verifies: REQ-01-168, REQ-01-585, REQ-01-590, REQ-03-287..REQ-03-288
- **AC-425**: Migration accepts only `status='active'` with `closed_at=NULL` and `status='closed'` with non-null `closed_at`; existing valid active rows with `closed_at=NULL` remain valid; unknown statuses, `active` rows with non-null `closed_at`, and `closed` rows with null `closed_at` fail migration with a row-level remediation report containing `incident_id`, field, raw value or raw value pair, reason code, and remediation hint; and migration does not silently coerce unknown lifecycle states or inconsistent state/timestamp pairs.
  - Verifies: REQ-02-222..REQ-02-223, REQ-02-254
- **AC-426**: Opening or resuming a WebSocket collaboration connection to a `closed` incident never establishes a writable collaboration subscription: it does not produce a writable `hello_ack` or `resume_ack`, it emits a terminal `error` with `code='incident_closed'` and `retryable=false`, and the client treats that signal as closed-read-only state rather than retrying queued source mutations.
  - Verifies: REQ-01-592, REQ-03-288
- **AC-243**: With grouping active and `limit=1`, the response serializes exactly one data row in `rows[]` as one full `view_row_v1`, includes no group-header pseudo-row, and still returns the full `group_values` object for that row; page-size accounting applies to serialized `rows[]` entries only.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-037
- **AC-372**: On `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, issue page 1 with a deterministic sort and a finite `limit`, then commit an intervening insert whose live sort position falls after the current continuation position. Continuing with the returned `meta.paging.next_cursor` may surface the inserted row only when it currently matches the bound route contract and current authorization, and the returned payload reflects fetch-time authoritative state.
  - Verifies: REQ-01-035, REQ-01-554, REQ-01-555, REQ-01-556
- **AC-373**: On `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, issue page 1 with a deterministic sort and a finite `limit`, then commit an intervening delete, authorization removal, restore, or sort- or filter-relevant edit affecting one or more rows that would otherwise appear later in the live result set. Continuing with the returned `meta.paging.next_cursor` does not return rows that are no longer currently visible, and any returned row uses fetch-time live payloads, including current `row_version` when present.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-555, REQ-01-556, REQ-01-557
- **AC-374**: Using the same fixture shape as `AC-372` or `AC-373`, restart the same route without `cursor_token` after the intervening mutations. The fresh request reflects current live membership, current live ordering, and current live row values; continuation also re-derives authorization but remains positioned by the server-owned cursor state rather than by client-supplied offset fields.
  - Verifies: REQ-01-035, REQ-01-036, REQ-01-554, REQ-01-555, REQ-01-557, REQ-01-558, REQ-01-560
- **AC-375**: Using a non-null `meta.paging.next_cursor` returned by a successful page-1 response from `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, tampering with the cursor payload, signature, actor binding, route binding, route scope, query contract, or effective limit fails closed with `400`, `error.code='invalid_view_query'`, and a pagination reason code such as `invalid_cursor_token` or `cursor_query_mismatch`. Reissuing the same route without `cursor_token` succeeds normally against current live state. When the Network Flow Activity Extension Profile is claimed, equivalent fixtures for pageable Network Flow routes prove the common `cursor_token` and `meta.paging.next_cursor` envelope, ASCII byte bound, opaque sealed payload, actor/session binding, route binding, incident binding, table-scope binding, normalized-query binding, full keyset continuation binding, 15-minute TTL, expiry-at-equality behavior, soft-delete and authorization-loss invalidation, rename survival, and key rotation. Malformed, tampered, too-long, expired, unknown-key, retired-key, actor-mismatched, session-mismatched, route-mismatched, incident-mismatched, table-scope-stale, query-mismatched, authorization-lost, or invalid-position Network Flow cursors fail with `400`, `error.code='network_flow_cursor_invalid'`, and safe details that do not disclose payload state, hidden table IDs, comparator values, cryptographic failure mode, raw source values, or secret material.
  - Verifies: REQ-01-035, REQ-01-238, REQ-01-554, REQ-01-558, REQ-01-559, REQ-01-559A, REQ-01-560, REQ-04-128..REQ-04-130
- **AC-128**: `POST /api/v1/object-blobs` accepts only a JSON object with required `incident_id`, `client_txn_id`, and `byte_size`, plus optional `filename_hint`, `content_type_hint`, and `sha256_hex`; rejects malformed or unknown-field input with `400 error.code='invalid_blob_create_request'`; rejects same-scope `client_txn_id` reuse with a different normalized request using `409 error.code='client_txn_conflict'`; returns `201 Created` on first success and `200 OK` on same-request replay; and returns `incident_id`, `object_blob_id`, `upload_state`, `target_expires_at`, `pending_expires_at`, `upload_target`, and `accepted_contract`, with omitted optional contract members serialized as explicit `null` inside `accepted_contract`. In the base profile `upload_target.method='PUT'`, `upload_target.href` is an opaque same-origin `/api/v1/object-uploads/{upload_token}` capability, `upload_target.headers` contains any required upload headers, and the public response does not expose raw bucket names, storage keys, presigned object-store query parameters, or object-store hostnames. `PUT /api/v1/object-uploads/{upload_token}` accepts the exact byte count bound to the pending slot, streams the body through the application to object storage, returns `204 No Content` on success, fails closed for malformed or unknown tokens, expired upload targets, byte-size mismatches, non-pending slots, or object-store dependency failures, and does not itself finalize evidence attachment. In the base profile the upload target expires 60 minutes after issuance, the pending slot expires 24 hours after creation, same-request replay after target expiry returns the same expired slot rather than refreshing it, and obtaining a fresh target requires a fresh blob slot. `POST /api/v1/evidence-records/{record_id}/attach-blob` accepts only a JSON object with exactly `object_blob_id`, `base_row_version`, and `client_txn_id`; rejects malformed input, supplied `null` for those required members, or unknown top-level members with `400 error.code='invalid_mutation_payload'`; returns `200 OK` on first success and same-request replay; rejects same-scope key reuse with a different normalized request using `409 error.code='client_txn_conflict'`; rejects stale `base_row_version` with `409 error.code='row_version_conflict'`; fails closed with `409 error.code='evidence_attach_rejected'` and a registered reason code when blob or evidence lifecycle state does not satisfy the attach bridge rules from Core 03 §8 and Core 02 §13; and on success returns `data.view_schema_id='cartulary.view.evidence.v1'`, `data.change_set_id`, `data.object_blob_id`, and `data.row=<view_row_v1>` rather than a bespoke evidence-refresh object.
  - Verifies: REQ-00-014, REQ-01-019, REQ-01-234, REQ-01-238, REQ-01-243..REQ-01-247, REQ-01-328,
    REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128,
    REQ-04-048
- **AC-129**: Long-running operations started through the public interface return `202 Accepted` with a `job_id`; `GET /api/v1/jobs/{job_id}` exposes progress and terminal result or error summary; the incident-scoped WebSocket stream authenticates with the same session contract and emits presence, `record_changed`, and `job_progress` events without broadcasting client-local drafts or grouping UI state; and when a terminal polled job resource or terminal `job_progress` message includes `result_summary.resource_refs[]`, the client treats those refs as a compact navigation summary, renders known current-profile kinds as non-modal result chips or links on the current surface, degrades unknown kinds to message-only rendering without breaking job rendering, and does not auto-follow `route` or change the active workbook surface, selection, or scroll position.
  - Verifies: REQ-00-014, REQ-01-018..REQ-01-019, REQ-01-248..REQ-01-277, REQ-01-452..REQ-01-454,
    REQ-03-092..REQ-03-098
- **AC-257**: A long-running operation started through the public interface returns `202 Accepted` with the canonical job resource in the common success envelope; `data.status` is `queued` or `running`; `data.status_route` is `/api/v1/jobs/{job_id}` for that resource; `scope.kind` is either `incident` or `deployment`; only `queued`, `running`, `cancel_requested`, `succeeded`, `failed`, and `canceled` appear as public job states; queued work commits `running` before any handler invocation; no queued job transitions directly to a terminal state; expired `running` and `cancel_requested` leases recover without public-state regression; terminal states are immutable; and polling `GET /api/v1/jobs/{job_id}` never observes a public state transition outside the legal state machine. Incident-scoped streams emit `job_progress` only for committed public-resource changes to incident-scoped jobs, and deployment-scoped jobs do not appear on incident-scoped streams. A deployment with every optional profile unclaimed retains one complete recognized-definition catalog, activates a quiescent job-dequeue component with an empty runtime selection, starts no profile handler, and remains Base Profile conformant. The immutable production attempt-operation timeout is 10 seconds and is independently test-injectable from the 10-second lease-renewal cadence. When graceful runner shutdown and a renewal tick or handler outcome are simultaneously ready before a terminal commit, shutdown wins, still-owned executions are conditionally released concurrently within handler concurrency and the earlier of the operation timeout or caller deadline, no release result increases failure count or sets a retry delay, and successful close waits for handler drain; the same renewal or handler error without runner shutdown retains its ordinary failure behavior. One or more release timeouts or operational failures make close unsuccessful with deterministic slot-ordered diagnostics containing only the operation stage, job kind, bounded attempt-slot ordinal, and closed reason; job and attempt identifiers, raw database errors, payloads, and incident content remain absent.
  - Verifies: REQ-01-248..REQ-01-249, REQ-01-268, REQ-04-023
- **AC-258**: Over the authoritative job row, `GET /api/v1/jobs/{job_id}`, and replayed or live `job_progress` messages, `progress` is always an object with non-negative integer `completed` and `total` equal to either a positive integer or `null`; `completed` never decreases for one job; an unknown total may become known once, and a known total never clears or decreases; an exact repeated update is a mutation-free and event-free success; rejected progress is mutation-free and event-free; concurrent stale progress never overwrites newer progress; when `total` is known, `completed <= total`; when `total = null`, the client renders indeterminate progress rather than a fake percent; and when `status='succeeded'` with known `total`, the terminal row, resource, and event have `completed == total`.
  - Verifies: REQ-01-249, REQ-01-268, REQ-01-453
- **AC-259**: Non-terminal job resources carry `result_summary=null` and `error_summary=null`, and any non-terminal `job_progress` message that includes those members does the same; `succeeded` and `canceled` terminal states carry only `result_summary`; `failed` carries only `error_summary`; when `status='canceled'`, `result_summary.code='job_canceled'`; terminal `succeeded` or `canceled` summaries do not carry family-specific deep payloads in place of the common summary; handler errors, recovered panic values, job IDs, payload or incident content, storage paths, secrets, and internal progress-unit IDs appear in no public error, log, telemetry item, or durable public summary; any current-profile `result_summary.resource_refs[]` items use only `incident`, `import_session`, `snapshot`, `release`, `reference_pack_version`, or `incident_bundle`; every current-profile emitted ref includes a canonical same-origin `GET` `route` with no query or fragment; `reference_pack_version.id == route`; any multi-ref current-profile terminal summary uses deterministic ordering, with `reference_packs_refreshed` sorted by `route asc`; `started_at` is `null` until work begins; `finished_at` is `null` before terminal and non-null after terminal; and `retained_until` is `null` before terminal and non-null after terminal on the HTTP job resource and, when present, on `job_progress`.
  - Verifies: REQ-01-249, REQ-01-268
- **AC-260**: `POST /api/v1/jobs/{job_id}/cancel` requires a JSON object with `client_txn_id` and accepts optional `reason`; omission and explicit JSON `null` for `reason` compare equal for normalized request comparison; first success and idempotent replay both return `200 OK` with the current authoritative job resource; same-actor replay of the same normalized request with the same `(job_id, client_txn_id)` creates no second cancel transition; reuse of that scope key with a different normalized request fails with `409 error.code='client_txn_conflict'`; and rejected cancel attempts fail with `409 error.code='job_cancel_rejected'` plus exact `reason_code` equal to `already_cancel_requested`, `already_terminal`, or `not_cancelable`.
  - Verifies: REQ-01-234, REQ-01-238, REQ-01-249, REQ-01-453, REQ-04-023
- **AC-261**: Terminal jobs expose `retained_until >= finished_at + 7 days`; `GET /api/v1/jobs/{job_id}` succeeds before expiry and may return `404 error.code='job_not_found'` after expiry; the same `404 error.code='job_not_found'` behavior is used for absent or unauthorized job reads or cancel requests, including a caller with `deployment_admin=true` but no incident membership attempting to access an incident-scoped job; and expiring the job resource does not delete or mutate durable outputs that the job produced.
  - Verifies: REQ-01-234, REQ-01-249, REQ-04-023, REQ-04-029
- **AC-130**: A cookie-authenticated state-changing request with missing or invalid CSRF proof fails closed, and the same request succeeds when a valid CSRF proof accompanies an otherwise authorized session.
  - Verifies: REQ-01-023..REQ-01-031, REQ-01-154, REQ-04-001..REQ-04-004, REQ-04-052..REQ-04-053
- **AC-131**: On `GET /ws/v1/incidents/{incident_id}`, an authorized same-origin client that sends `hello` as its first application message receives `hello_ack` containing `connection_id`, `resume_token`, `server_time`, `heartbeat_interval_ms=15000`, `presence_ttl_ms=45000`, and `resume_window_ms>=300000`, followed by `presence_snapshot` whose `presences[]` member is always present, may be empty, excludes expired presence records, contains no duplicate `connection_id`, and serializes remaining entries in canonical ascending `connection_id` order; a cookie-authenticated browser connection from an untrusted `Origin` is rejected before `hello_ack` or `presence_snapshot`, and an otherwise idle accepted connection receives application-level `ping` within 15 seconds and is closed after 45 seconds without any inbound frame or timely `pong`. Client application messages are text-frame-only UTF-8 JSON objects bounded to `32768` reassembled bytes, binary application messages close as `1003/binary_message_unsupported`, oversized messages close as `1009/message_too_large`, malformed JSON closes as `1007/invalid_json`, duplicate JSON members are rejected, invalid first semantic messages receive `invalid_websocket_handshake` before `1008/invalid_first_message`, invalid later semantic messages receive `invalid_websocket_message` before `1008/invalid_message`, and valid server messages use one text frame without an encoder-added trailing line feed. A Base claim is non-conformant if `GET /ws/v1/incidents/{incident_id}` is absent or if the implementation treats any different v1 public subscription route as equivalent to that canonical route.
  - Verifies: REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277A, REQ-03-092..REQ-03-098, REQ-04-005..REQ-04-017,
    REQ-04-052..REQ-04-053
- **AC-132**: When analyst A changes workbook surface, focused row, or same-cell edit state on an incident, analyst B subscribed to the same incident receives the corresponding `presence_delta` within 1 second, and the UI renders workbook-header, row-gutter, and same-cell indicators from matching `sheet_ref`, `record_id`, and `field_key` rather than visible labels or row numbers. Presence for direct opening of any pack-independent base-profile surface from `REQ-01-307`, including `cartulary.view.task_requests.v1`, `cartulary.view.decisions.v1`, `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`, `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`, uses `sheet_ref.kind='view_schema'` with the standardized `view_schema_id`; presence for a distinct saved view over the same schema uses `sheet_ref.kind='saved_view'` with that saved view's `saved_view_id`.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-090..REQ-03-098
- **AC-133**: After a transient disconnect, a client that reconnects to the same incident within the retained replay window can send `resume` with `resume_token` and `last_seen_stream_seq`, receives `resume_ack.status='replayed'`, then receives missed `record_changed` and incident-scoped `job_progress` messages in strict ascending `stream_seq`, followed by a fresh `presence_snapshot` whose `presences[]` member is always present, may be empty, excludes expired presence records, contains no duplicate `connection_id`, and serializes remaining entries in canonical ascending `connection_id` order.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098
- **AC-134**: If a client reconnects with an expired, unknown, malformed, or too-old `resume_token`, but the caller still has valid incident authorization, the server responds with `resume_ack.status='reset_required'`, sends a fresh `presence_snapshot` whose `presences[]` member is always present, may be empty, excludes expired presence records, contains no duplicate `connection_id`, and serializes remaining entries in canonical ascending `connection_id` order, and emits no guessed or partial replay for the missing range; the client recovers only by re-querying current workbook state through the existing HTTP view route.
  - Verifies: REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098
- **AC-135**: Replayable WebSocket messages on one incident carry monotonically increasing `stream_seq` values assigned only after the underlying record mutation or incident-scoped job-state change is committed; each participating committed live record revision contributes exactly one deterministic `record_changed` intent in its authoritative transaction, rollback contributes none, historical incident-bundle import contributes none for imported historical revisions, and that suppression does not leak into a later transaction; clients can de-duplicate duplicates by `(incident_id, stream_seq)`, reconcile row application by `record_id` plus `row_version`, must perform HTTP resynchronization rather than guessed incremental apply when a replayable sequence gap is observed, and observe identical canonical `changed_field_keys[]` and `affected_views[]` array order when a replayable `record_changed` with semantically identical array content is delivered live and via replay.
  - Verifies: REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277A, REQ-03-092..REQ-03-098
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
- **AC-162**: If a fresh transaction commits a user's membership loss in incident A while the user retains a still-valid session and access to incident B, only that user's incident-A sockets receive `session_revoked` with `reason_code='incident_access_revoked'` after commit and then close; future incident-A requests fail closed; other users' incident-A sockets are unchanged; and the same account session can still query or subscribe to incident B if otherwise authorized. Rollback, commit failure, pre-commit failure, rejected/no-op mutation, and replay produce no socket effect, and process-local write failure does not reinterpret the committed HTTP mutation.
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

- **AC-170**: An authenticated session whose internal user account is active and not disabled can `POST /api/v1/incidents` with required `client_txn_id`, `incident_key`, and `title`, plus optional nullable `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`; omitted optional fields and explicit `null` serialize as `null`; non-null `tlp` accepts and emits only the exact Core 02 `incident.tlp` tokens; on first success the caller receives `201 Created` plus `Location: /api/v1/incidents/{incident_id}`, and the response `data` includes the incident resource fields defined by Core 01 §3.3.5.3 with `status='active'`, `incident_version=1`, `closed_at=null`, `created_by_user_id` equal to the actor, `updated_by_user_id` equal to the actor, and `created_at == updated_at`.
  - Verifies: REQ-01-152..REQ-01-180, REQ-01-491..REQ-01-491.1, REQ-02-203, REQ-02-222..REQ-02-223, REQ-02-243
- **AC-171**: A successful incident create bootstraps exactly one creator membership with `role='admin'`; `GET /api/v1/incidents/{incident_id}/memberships` returns that membership for the creator; `GET /api/v1/incidents/{incident_id}/workbook-preferences/default` returns a resource with `default_sheet_ref=null`; `GET /api/v1/incidents/{incident_id}/workbook-preferences/me` for the creator returns a resource with `home_sheet_ref=null`; and `GET /api/v1/incidents` for that caller returns only incidents where the caller currently has membership in `data.incidents[]`, ordered by `updated_at desc, incident_id asc`, with `meta.paging.limit=100` when `limit` is omitted, `meta.paging.has_more=false` and `meta.paging.next_cursor=null` on terminal pages, `400 error.code='invalid_pagination_request'` for invalid `limit` values or aliases such as `page`, `offset`, `page_size`, and `block_size`, rejection of cursor replay against a different bound contract, and rejection of pagination members on `GET /api/v1/incidents/{incident_id}` with `error.details.reason_code='pagination_not_supported'`.
  - Verifies: REQ-01-152..REQ-01-180, REQ-01-240..REQ-01-242
- **AC-172**: `POST /api/v1/incidents` without authentication, with an expired or revoked session, from a disabled internal user account, or with missing or invalid CSRF protection for a cookie-authenticated request fails closed and creates no incident, membership, or workbook-preference state.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-173**: Replaying `POST /api/v1/incidents` with the same `(actor_user_id, client_txn_id)` and the same normalized request returns `200 OK`, repeats `Location: /api/v1/incidents/{incident_id}` for the originally created incident, returns the originally created incident resource, and creates no second incident; replaying with the same `(actor_user_id, client_txn_id)` and a different normalized request fails with `409` and `error.code='client_txn_conflict'`.
  - Verifies: REQ-01-152..REQ-01-180
- **AC-174**: Creating an incident with an `incident_key` whose trimmed Unicode-NFC-normalized form conflicts with an existing incident fails with `409` and `error.code='incident_key_conflict'`; `error.details.field='incident_key'` and `error.details.incident_key_canonical` are present; and no partial incident, membership, or workbook-preference bootstrap state is left behind.
  - Verifies: REQ-01-152..REQ-01-180, REQ-02-243
- **AC-219**: `POST /api/v1/incidents` whose body is not a JSON object, omits required `client_txn_id`, `incident_key`, or `title`, supplies `null` for a non-nullable field, exceeds the declared field-length limits, includes a control character in any string field bound by a create contract, attempts to set a server-managed field, supplies an empty, whitespace-only, aliased, case-variant, or otherwise non-canonical `tlp` value, includes any top-level member outside `client_txn_id`, `incident_key`, `title`, `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`, or sends `initial_memberships[]` fails with `400` and `error.code='invalid_incident_create'`; when one member is responsible, the response includes `error.details.field`, and `error.details.reason_code` uses the Core 01 registry, including `invalid_value` for malformed or non-canonical `tlp`, `unknown_field` for an undeclared top-level member, `server_managed_field` for a forbidden server-managed member, and `collaborator_seeding_not_supported` for `initial_memberships[]`.
  - Verifies: REQ-01-021, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-01-491..REQ-01-491.1, REQ-02-222..REQ-02-223
- **AC-220**: On `POST /api/v1/incidents`, idempotency comparison happens only after validation succeeds. A replay that differs only by omission versus explicit `null` for `description`, `severity`, `tlp`, `current_phase`, or `primary_external_case_ref` returns the originally created incident when those pairs normalize equivalently; normalized-empty string input compares as `null` only for nullable fields bound to clear-to-null writable-string contracts and never for `tlp`. A request that includes any undeclared top-level member fails with `400`, `error.code='invalid_incident_create'`, and `error.details.reason_code='unknown_field'` even if `client_txn_id` matches a prior successful create, and a request that differs only by an invalid TLP alias or empty TLP value fails with the applicable Core 01 reason code rather than replaying.
  - Verifies: REQ-01-021, REQ-01-152..REQ-01-180, REQ-01-491..REQ-01-491.1, REQ-02-222..REQ-02-223
- **AC-211**: `GET /api/v1/incidents/{incident_id}` returns the incident resource defined by Core 01 §3.3.5.3 and §3.3.5.3.1, including `incident_version`, `updated_at`, and `updated_by_user_id`, to any current incident member; a caller lacking visibility receives `404`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023
- **AC-212**: `PATCH /api/v1/incidents/{incident_id}` with a correct `base_incident_version` by a `reviewer` or `admin` can set, replace, clear, and no-op only `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`; explicit `null` clears any nullable mutable field; normalized-empty input clears only nullable fields bound to clear-to-null writable-string contracts; empty or whitespace-only `tlp` fails validation; success returns `200 OK` with the updated incident resource; an effective change increments `incident_version` exactly once and sets `updated_at` plus `updated_by_user_id`; a no-op patch returns `200 OK` without incrementing `incident_version`, changing `updated_at`, or changing `updated_by_user_id`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-491..REQ-01-491.1, REQ-02-014..REQ-02-023, REQ-02-222..REQ-02-223
- **AC-213**: `PATCH /api/v1/incidents/{incident_id}` with a stale `base_incident_version` fails with `409` and `error.code='incident_version_conflict'`; a caller who can see the incident but lacks sufficient role gets `403`; a caller lacking visibility gets `404`.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023
- **AC-214**: `PATCH /api/v1/incidents/{incident_id}` that attempts to mutate `incident_id`, `incident_key`, `title`, `status`, `created_by_user_id`, `created_at`, `updated_at`, `updated_by_user_id`, `incident_version`, `closed_at`, membership objects, saved-view objects, or workbook-preference objects, that sends unknown top-level members, that supplies a non-canonical or normalized-empty `tlp`, or that violates the bound string contract for `description`, `severity`, `current_phase`, or `primary_external_case_ref`, fails with `400` and `error.code='invalid_incident_patch'`, and leaves no partial incident state.
  - Verifies: REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-01-491..REQ-01-491.1, REQ-02-014..REQ-02-023, REQ-02-222..REQ-02-223
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

- **AC-181**: `DELETE /api/v1/records/{record_id}` accepts JSON `base_row_version`, `client_txn_id`, and optional `reason`; `reason` is normalized under `reason_note_v1`, and omission, explicit `null`, and normalized-empty reason compare equal for idempotency and persist as `null`; an `editor`, `reviewer`, or `admin` on the incident receives `200 OK` with `record_id`, `incident_id`, incremented `row_version`, `deleted=true`, non-null `deleted_at`, non-null `deleted_by_user_id`, and `change_set_id`; the record no longer appears in ordinary view queries; `GET /api/v1/records/{record_id}/history` still returns prior history plus a `soft_delete` entry; the collaboration stream emits `record_changed` with `changed_field_keys[]` present and empty and with `affected_views[]` present, non-empty, free of duplicate `view_schema_id` values, serialized in canonical ascending `view_schema_id` order, and using `change_kind='remove'` on each affected-view entry; a stale base version fails with `409 error.code='row_version_conflict'`; type-specific delete precondition failures fail with `409 error.code='record_delete_blocked'` and owner-defined `error.details.reason_code`; patching the deleted record fails with `409 error.code='record_deleted_use_restore'`; and replaying the same normalized delete request by the same actor with the same `(record_id, client_txn_id)` returns the originally committed success without creating a second mutation.
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
- **AC-413**: A default runtime exposes no `/api/v1/test/*` or `/ws/v1/test/*` route. When test routes are enabled without the harness-owned marker or without a valid test-route token, startup fails closed before serving those routes. In a harness-owned runtime, missing or wrong `X-Cartulary-Test-Route-Token` values return `403 error.code='test_route_forbidden'` for every test-only harness route adopted by Testing Harness NLSpec Section 12, including clock controls, runtime reset and identity, public-error fault control, Network Flow runtime controls, saved-view system fixtures, auth touch, timeline snapshot, and test WebSocket routes; session cookies, bearer sessions, bootstrap tokens, incident roles, and `deployment_admin` do not bypass that token requirement. If a harness API or browser origin is configured, wrong host, missing origin, malformed origin, or unapproved origin requests fail before any runtime-control mutation, and the response does not enable permissive CORS.
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
- **AC-187**: The same route fails closed when the two IDs are identical, the records belong to different incidents, the record types differ, the record type is not `host` or `identity`, either record is soft-deleted or already merged away, the caller lacks `reviewer` or `admin` role, either supplied base row version is stale, preserving a loser-side `exact_match_reuse` value would collide with a third active same-incident record, or an overlapping in-flight destructive operation already holds one or more required protected-set locks; stale versions fail with `409 error.code='row_version_conflict'`; an active lock fails with `409 error.code='record_locked'` and `error.retryable=true`; the third-record carry-forward collision fails with `409 error.code='merge_precondition_failed'`, `error.details.reason_code='carry_forward_identifier_collision'`, and `error.details` containing `identifier_class`, `normalized_value`, and `blocking_record_id`; same-incident semantic merge-precondition failures fail with `409 error.code='merge_precondition_failed'`; a malformed, missing, foreign-incident, or otherwise hidden survivor or loser receives `404 error.code='incident_not_found'`; no partial state is committed on any failure path; and replaying the same normalized request by the same actor with the same `(survivor_record_id, loser_record_id, client_txn_id)` returns the originally committed success without creating a second merge. Authentication and state-changing CSRF dominate malformed path and body failures; survivor visibility and the `reviewer` or `admin` role gate dominate body validation; all failures use the common error envelope and commit no source, history, projection, idempotency, or collaboration state.
  - Verifies: REQ-01-032..REQ-01-033, REQ-01-104, REQ-01-181..REQ-01-195, REQ-01-237, REQ-02-060..REQ-02-061,
    REQ-02-064..REQ-02-066, REQ-03-101, REQ-03-247..REQ-03-249

- **AC-277**: `POST /api/v1/incidents/{incident_id}/views/cartulary.view.parties.v1/rows` commits only when `party.display_name` remains non-empty after `display_name_line_v1` normalization and `party.party_kind` is present; zero-field create fails with no partial row or misleading projection update; `GET /api/v1/view-schemas` and `GET /api/v1/view-schemas/cartulary.view.parties.v1` expose the Parties schema as part of the fourteen pack-independent base-profile view contracts; and the Parties system view is incident-scoped rather than deployment-local user administration.
  - Verifies: REQ-01-296, REQ-01-343, REQ-01-497..REQ-01-501, REQ-02-003, REQ-02-006, REQ-02-009, REQ-02-022, REQ-02-202, REQ-02-222..REQ-02-225, REQ-03-005, REQ-03-266
- **AC-278**: Linking or clearing `task.requester_party_id`, `evidence.collector_party_id`, or `evidence.source_party_id` from the inspector or Parties view preserves independent text and ref semantics: text-only, ref-only, both-present, and both-null states round-trip; clearing text does not clear the ref; clearing the ref does not clear text; `owner_user_id` remains the accountable assignee when present; and `comm_log.audience` text remains required when supplemental audience or attendee party refs are added or removed.
  - Verifies: REQ-01-328, REQ-01-336, REQ-01-502, REQ-02-017, REQ-02-021..REQ-02-022, REQ-02-120, REQ-02-124..REQ-02-125, REQ-02-197..REQ-02-199, REQ-02-226..REQ-02-228, REQ-03-247, REQ-03-256, REQ-03-268..REQ-03-271
- **AC-279**: Typing requester, collector, source, or audience text into ordinary workbook fields does not auto-create or auto-link a `party` record. Explicit `Create party from text` or `Link existing party` flows from the inspector or Parties view can do so, and explicit supplemental audience or attendee party-reference actions can link existing parties without replacing source-preserving audience text; exact-match create or reuse is limited to one active same-incident party selected by normalized `primary_email` or, failing that, `external_ref`; display name, organization, role title, and phone-like text do not act as auto-upsert keys; and the resulting create or link flow remains same-surface and non-blocking.
  - Verifies: REQ-01-296, REQ-01-328, REQ-01-336, REQ-01-497, REQ-01-499..REQ-01-502, REQ-02-022, REQ-02-060..REQ-02-063, REQ-02-229..REQ-02-232, REQ-03-247, REQ-03-256, REQ-03-259, REQ-03-267..REQ-03-271
- **AC-280**: A `task.requester_party_id`, `evidence.collector_party_id`, or `evidence.source_party_id` that targets a party from another incident, a deleted party, a wrong-record-type UUID, or a deployment-local `user_id` fails closed and leaves no partial link state; deleting a currently referenced party fails closed and leaves authoritative state unchanged; and party rows and party references inherit the same incident-level authorization model as other incident data rather than any deployment-admin bypass.
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
- **AC-301**: `PATCH /api/v1/records/{record_id}` with explicit JSON `null` clears clearable timestamp fields such as `evidence.requested_at`, `evidence.received_at`, `task.due_at`, `task.completed_at`, `comm_log.next_report_at`, `handoff.acknowledged_at`, and `status_review.next_report_at` when the resulting row otherwise remains legal; omission leaves the persisted value unchanged. Timeline v2 visible Activity Date cells are not timestamp scalar cells and instead follow `timeline_visible_text_v1`.
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
- **AC-318**: Setting `task.requester_party_id`, `evidence.collector_party_id`, or `evidence.source_party_id` accepts only an exact same-incident active `party` record identifier; a wrong-type record, another-incident record, deleted or otherwise ineligible target, deployment-user UUID, email text, label text, fuzzy value, or non-scalar value fails closed with no partial authoritative state change. Clearing or setting those fields preserves independent text-plus-ref semantics: clearing the ref leaves preserved requester, collector, or source text unchanged; clearing text leaves the ref unchanged; and `Clear both` submits one patch containing both field changes and leaves both authoritative values `null` with no partial intermediate state committed.
  - Verifies: REQ-01-328, REQ-01-336, REQ-01-502, REQ-02-017, REQ-02-021..REQ-02-022, REQ-02-231, REQ-03-272, REQ-03-274
- **AC-319**: `task.decision_record_id` accepts only an exact same-incident active `decision` record identifier; a wrong-type record, another-incident record, deleted or otherwise ineligible target, or explicit clear routed through a non-direct-write shape fails closed with no partial authoritative state change; and when the task surface exposes decision-link clearing it uses the ordinary patch route with `value=null`.
  - Verifies: REQ-01-336, REQ-01-517..REQ-01-520, REQ-02-233, REQ-03-273


- **AC-370**: `GET /api/v1/extensions` succeeds for an authenticated current session, including one with `is_deployment_admin=false` and zero current incident memberships. The route returns the common success envelope with `data.extensions[]` ordered by `profile_id asc`; the exact current-profile set is `enterprise_authentication`, `import`, `incident_portability`, `network_flow_activity`, `reference_pack`, and `snapshot_reporting`; each item contains exactly `profile_id`, `claimable`, `claimed`, `contract_major`, `route_families[]`, `workspace_keys[]`, and `capabilities[]`; `network_flow_activity` uses major `4`, workspace `network_analysis`, its validated published claim state, and an empty capability array; every other capability array is empty; each route family matches Core 01 and is ordered ascending; compatible decoders ignore unknown additive members but never execute them; pagination members fail with `400 error.code='invalid_pagination_request'` and `error.details.reason_code='pagination_not_supported'`; and the route discloses no provider secrets, provider claim maps, implementation/document versions, registry digests, or live extension-family payload.
  - Verifies: REQ-00-022, REQ-01-032..REQ-01-033, REQ-01-542..REQ-01-546, REQ-04-105
- **AC-371**: When `GET /api/v1/extensions` reports one or more items with `claimed=false`, probing each such profile's declared `route_families[]` root and one descendant route under that same family returns `404 error.code='extension_profile_not_claimed'`, `error.retryable=false`, and `error.details` containing the matching `profile_id` and canonical matched `route_family`. That result is returned before family-specific authorization or policy evaluation for the unclaimed family; a path outside all reserved base and extension families does not use `extension_profile_not_claimed`; and a valid path inside a claimed extension family does not use `extension_profile_not_claimed` solely because the family currently has zero resources or because ordinary family authorization later denies access.
  - Verifies: REQ-00-022, REQ-01-033, REQ-01-234, REQ-01-544, REQ-01-546..REQ-01-548, REQ-04-105

### 9.11 Incident Portability Extension Profile criteria

- **AC-164**: Whole-incident export produces a bundle whose logical layout, manifest, checksum file, structured JSON or NDJSON files, and blob paths satisfy Core 01 §12.3, and the bundle excludes projections, search indexes, sessions, presigned URLs, locks, client-local drafts, login-capable local users, deployment-admin flags, auth-binding state, memberships, permissions, deployment-local administrative audit history including deployment and incident-membership administrative audit events, password hashes, MFA secrets, external-provider configuration, and object-store credentials.
  - Verifies: REQ-01-425..REQ-01-442, REQ-04-044..REQ-04-047, REQ-04-065
- **AC-165**: Exporting an incident and importing that bundle into an empty deployment preserves the exported `incident_id`, `record_id`, `row_version`, change-set count, revision count, record-link count, record-tag attachments, evidence custody events, entity-mention count, indicator-observation count, and blob hashes, and the imported incident opens normally after projection rebuild.
  - Verifies: REQ-01-425..REQ-01-426, REQ-01-439..REQ-01-442, REQ-01-447..REQ-01-450, REQ-04-044..REQ-04-047,
    REQ-04-065
- **AC-166**: If one required structured file or one required blob is missing, if any required checksum is corrupted, if `incident_id` already exists, or if a required capability is unsupported, import fails closed before the incident becomes visible and leaves no partially visible incident state.
  - Verifies: REQ-01-425..REQ-01-430, REQ-01-433..REQ-01-438, REQ-01-447..REQ-01-451, REQ-04-044..REQ-04-047,
    REQ-04-065
- **AC-167**: If a bundle contains optional embedded `snapshots` or `reference_packs` sections that the target deployment does not support, the importer ignores or degrades only those optional sections under their owner rules and core incident import still succeeds; any structurally valid nonempty `required_capabilities[]` instead fails with `extension_capability_not_supported` and executes no capability behavior.
  - Verifies: REQ-01-425..REQ-01-426, REQ-01-431..REQ-01-432, REQ-01-443..REQ-01-450, REQ-04-044..REQ-04-047,
    REQ-04-065
- **AC-168**: Historical actors in `actors.ndjson` that do not map to an existing local user become either inert imported actors or historical actor descriptors bound to an existing local user; neither is login-capable, neither is automatically added to incident membership, and imported history retains the source bundle actor identifier used by that history.
  - Verifies: REQ-01-425..REQ-01-426, REQ-01-443..REQ-01-446, REQ-04-044..REQ-04-047, REQ-04-065
- **AC-169**: Whole-incident export and import run as background jobs, show progress and cancellation without blocking grid editing, stage bundle contents only under the configured temporary-work root, and in flyaway or disconnected deployments keep emitted bundles and staged extracts on encrypted storage.
  - Verifies: REQ-01-425..REQ-01-430, REQ-01-447..REQ-01-450, REQ-01-452..REQ-01-456, REQ-04-044..REQ-04-047,
    REQ-04-054..REQ-04-055, REQ-04-058..REQ-04-059, REQ-04-065
- **AC-386**: Exporting and importing an incident that contains both currently reversible and currently non-reversible single-entry-addressable history items preserves the authoritative history substrate needed to materialize `GET /api/v1/records/{record_id}/history`, including change-set count, mutation-entry count, and revision count; after projection rebuild, the imported incident's `/history` surface exposes the same logical history items and rollback scopes; the importing deployment MAY emit different opaque `history_entry_ref` values than the source deployment; repeated imported-history reads keep those imported selectors stable within the importing deployment; and rollback using an imported selector succeeds or fails only under the ordinary imported-history preconditions.
  - Verifies: REQ-01-564, REQ-02-238, REQ-02-241

- **AC-332**: Exporting and importing an incident that contains a superseded Timeline row with a committed replacement relation preserves the active Timeline `supersedes` link, including its exact null `field_key`, and after projection rebuild a Timeline row query against the imported incident returns the same hidden `timeline.replacement_record_id` for that superseded row. The same round trip preserves exact non-null canonical `field_key` values for field-routed links without label translation.
  - Verifies: REQ-01-311..REQ-01-312, REQ-01-448..REQ-01-449, REQ-01-486, REQ-02-169, REQ-02-175, REQ-02-181


- **AC-273**: `POST /api/v1/incident-bundles/export` accepts a JSON object with required `incident_id` and required `client_txn_id`; omitting `reference_pack_mode` defaults it to `refs_only`; explicit `null` for `reference_pack_mode`, `optional_sections[]`, or `required_capabilities[]` fails with `400 error.code='invalid_incident_bundle_request'` and `reason_code='field_not_nullable'`; explicit `reference_pack_mode='refs_only'` compares equal to omission for idempotency and replay; `reference_pack_mode` accepts only `refs_only` or `embedded`, otherwise failing with `reason_code='invalid_reference_pack_mode'`; omitting `optional_sections[]` or `required_capabilities[]` defaults each to `[]`, and explicit `[]` compares equal to omission; current-profile `optional_sections[]` tokens are exactly `snapshots` and `reference_packs`, caller order is non-semantic, duplicates coalesce, and canonical normalization sorts them ascending; an unknown or non-string optional-section member fails with `reason_code='invalid_optional_sections'`; a non-array or non-string `required_capabilities[]` fails with `reason_code='invalid_required_capabilities'`; every structurally valid nonempty capability array, including unknown strings, fails only after authentication and membership verification with `409 error.code='extension_capability_not_supported'`, `retryable=false`, exact `details.profile_id='incident_portability'`, no caller-supplied value disclosure, and no idempotency row or job; supplying `history_mode` or `blob_mode` fails with `400 error.code='invalid_incident_bundle_request'`; export returns `202` with the common job resource; and on success the terminal job summary uses `result_summary.code='incident_bundle_exported'` and exactly one `resource_refs[]` item `{ kind='incident_bundle', id=<bundle_id>, route='/api/v1/incident-bundles/{bundle_id}' }`.
  - Verifies: REQ-01-033, REQ-01-466..REQ-01-470, REQ-01-483..REQ-01-484
- **AC-274**: A durable incident-bundle descriptor exists only after successful export; `GET /api/v1/incident-bundles/{bundle_id}` returns that descriptor, rejects pagination members with `invalid_pagination_request`, exposes fixed current-profile `history_mode='full'` and `blob_mode='full'` rather than user-tunable partial-export modes, serializes the resolved `reference_pack_mode` and canonical `optional_sections[]`, and serializes `required_capabilities=[]`; the emitted `manifest.json` uses the same values.
  - Verifies: REQ-01-467..REQ-01-469, REQ-01-483..REQ-01-484
- **AC-275**: `POST /api/v1/incident-bundles/import` accepts only `multipart/form-data` with a required `boundary`, exactly two leaf parts named `metadata` and `file`, and no alternate upload framing; part order is non-semantic; the `metadata` part uses `Content-Type: application/json` or `application/json; charset=utf-8`, is UTF-8 and BOM-free, parses as one JSON object, and contains required `client_txn_id`; the `file` part media type is exactly one of `application/zip`, `application/x-tar`, `application/gzip`, `application/x-gzip`, or `application/octet-stream`; duplicate or unexpected parts, malformed metadata JSON, duplicate metadata keys, a metadata JSON value that is not an object, invalid part content type, or any envelope failure create no idempotency commit and no import job; replaying the same normalized metadata and file bytes by the same actor with the same `client_txn_id` returns the original accepted `202` result even when boundary text, part order, or advisory filename differ; the route returns `202` with the common job resource, uses `result_summary.code='incident_bundle_imported'` plus exactly one `resource_refs[]` item `{ kind='incident', id=<incident_id>, route='/api/v1/incidents/{incident_id}' }` on terminal success, and rejects clone, merge, identifier-remap, or remote-fetch modes.
  - Verifies: REQ-01-467..REQ-01-470, REQ-01-483, REQ-01-485, REQ-01-549..REQ-01-553
- **AC-276**: Incident-bundle routes use only
  `invalid_incident_bundle_request`, `incident_bundle_not_found`,
  `extension_capability_not_supported`, `incident_bundle_export_rejected`, and
  `incident_bundle_import_rejected`; `invalid_incident_bundle_request` uses
  only `unsupported_upload_envelope`, `missing_required_part`,
  `duplicate_part`, `unexpected_part`, `invalid_part_content_type`,
  `invalid_metadata_encoding`, `malformed_metadata_json`,
  `request_not_object`, `missing_required_field`, `field_not_nullable`,
  `unknown_field`, `invalid_reference_pack_mode`,
  `invalid_optional_sections`, `invalid_required_capabilities`,
  `history_mode_not_supported`, `blob_mode_not_supported`, and
  `invalid_value`; `extension_capability_not_supported` uses status `409`,
  `retryable=false`, and exactly `details.profile_id='incident_portability'`;
  multipart part-related failures populate
  `error.details.part_name`, and `invalid_part_content_type` also includes
  `error.details.received_content_type` plus
  `error.details.allowed_content_types[]`; export rejections surface only
  `missing_required_file`, `missing_required_blob`, or
  `extension_state_not_portable`, whose details additionally contain only the
  deterministic safe blocking `profile_id`; and import rejections
  surface only `invalid_member_path`, `unsupported_member_type`,
  `checksum_mismatch`, `signature_mismatch`, `blob_hash_mismatch`,
  `duplicate_incident_id`, `remote_fetch_required`, `missing_required_file`, `missing_required_blob`,
  `malformed_manifest`, `unsupported_bundle_version`,
  `source_family_invalid`, `archive_extracted_bytes_exceeded`,
  `archive_compression_ratio_exceeded`, `archive_member_count_exceeded`, or
  `initial_admin_unavailable`.
  - Verifies: REQ-01-471, REQ-01-486, REQ-01-553
- **AC-544**: Incident Bundle export re-evaluates every recognized `blocked_when_present` profile after acquiring the incident serialization boundary in the final publication transaction. No state proceeds; active or retained soft-deleted Network Flow state fails with `incident_bundle_export_rejected` and `reason_code='extension_state_not_portable'`; concurrent state committed before the final guard is observed; multiple blockers disclose only the UTF-8-smallest safe profile ID; and failure leaves no descriptor, public object reference, or residual published object.
  - Verifies: REQ-01-486, EXT-REQ-141, EXT-REQ-198, EXT-REQ-222, NF-REQ-182
- **AC-550**: Exporting, importing into an empty deployment, and re-exporting
  current Timeline history preserves each non-null version identifier as the
  byte-exact opaque value
  `timeline_record:<canonical-record-uuid>:<positive-row-version>` and
  preserves null as null. The embedded identifier and row version match the
  target and canonical snapshot, rollback succeeds or fails without prefix
  parsing, and no retired `timeline:` fixture, import translation, alias,
  backfill, dual reader, or dual writer is accepted.
  - Verifies: REQ-01-663, REQ-02-266
- **AC-557**: Exporting current canonical Links history, importing it into a
  reset empty deployment, and re-exporting preserves record-link values,
  record-tag values, composite tag targets, operation kinds, attribution, and
  timestamps without shape or byte-significant scalar drift. Before any
  import-side mutation, an unknown or missing member, mistyped or noncanonical
  scalar, invalid enum or nullability, illegal operation/value pairing,
  `tag_id`, compact link shape, bare tag UUID target, inferred default, or
  retired target grammar is rejected through the Incident Bundle failure
  surface. Pre-cutover databases and retained bundles are discarded and
  regenerated; no migration, history rewrite, bundle-version translation,
  compatibility mode, alias, fallback, or dual reader/writer is present.
  - Verifies: REQ-02-267..REQ-02-269
- **AC-442**: Successful incident-bundle import persists the submitting internal `user_id` at job admission, creates exactly one target-local membership for the imported incident with that user as `role='admin'`, creates the incident-wide `default_sheet_ref=null` and importer `home_sheet_ref=null` workbook-preference objects, emits one attributed `membership_created` administrative audit event, and makes the incident visible only after those objects, imported source state, and projections commit atomically. Historical actors, actor match hints, provider-subject hints, email hints, saved-view owners, and source-system role information create no additional memberships. Exact replay of a successful import creates no duplicate membership, preference object, audit event, or visible incident. If the submitter is missing, inactive, or no longer a deployment administrator at final publication, the job fails terminally with `incident_bundle_import_rejected` and `reason_code='initial_admin_unavailable'`, leaves no visible incident, leaves no membership or workbook preference object, and emits no successful membership audit event.
  - Verifies: REQ-00-058, REQ-01-448..REQ-01-450, REQ-01-485..REQ-01-486, REQ-01-609, REQ-03-290

- **AC-487**: Two exports from identical authoritative incident state and
  identical normalized export inputs produce byte-identical version `2`
  structured members, canonical manifest bytes, checksum inventory, and
  archive bytes; no export path emits version `1`.
  - Verifies: REQ-01-428..REQ-01-442, REQ-01-635..REQ-01-636
- **AC-488**: Importing a valid version `2` export into an empty deployment and
  re-exporting it preserves authoritative incident and record identifiers,
  source rows, history, attribution, Timeline state and provenance, blob
  digests, and source-owner invariant results; the imported incident becomes
  visible only through the final proven commit.
  - Verifies: REQ-01-425..REQ-01-426, REQ-01-448..REQ-01-450,
    REQ-01-609, REQ-01-636, REQ-01-640..REQ-01-641
- **AC-489**: A structurally valid archive declaring retired numeric version
  `1` is admitted through the ordinary asynchronous job boundary and then
  fails with `incident_bundle_import_rejected`,
  `reason_code='unsupported_bundle_version'`, and `retryable=false` before
  source preparation, extension preparation, cross-owner transaction
  acquisition, target mutation, or publication. The admitted job,
  idempotency record, payload, terminal failure, and required staging cleanup
  remain; no incident, membership, preference, attribution, successful audit,
  final object, or success metric is created.
  - Verifies: REQ-01-635..REQ-01-636, REQ-01-640
- **AC-490**: Before source preparation, import rejects omitted, JSON `null`,
  and non-integer `bundle_version` with `malformed_manifest`; rejects an
  integer other than `2`, including retired numeric version `1`, with
  `unsupported_bundle_version`; and rejects a retired or mixed Timeline path
  set under version `2`, an unknown
  or duplicate core path, a missing required path, checksum failure, traversal,
  unsupported member, extracted-byte excess, compression-ratio excess, or
  member-count excess with the exact closed import reason and no codec
  fallback.
  - Verifies: REQ-01-428..REQ-01-430, REQ-01-433..REQ-01-438,
    REQ-01-449, REQ-01-486, REQ-01-635, REQ-01-642
- **AC-491**: Every imported Timeline row binds to one same-incident
  `timeline_event` record envelope; version `2` files accept only their closed
  row shapes; provenance composite identities are unique and non-orphaned; and
  import loses or duplicates no provenance row.
  - Verifies: REQ-01-636, REQ-01-640
- **AC-492**: When Incident Portability is claimed, its single module facade
  rejects each missing required dependency before exposing transaction
  capabilities, routes, or work. Application composition installs the
  cross-owner coordinator exactly once, registers the named handler exactly
  once, and only then publishes routes. A nil Jobs manager, nil runner,
  unconfigured runner, absent dequeue gate, runner closed before publication,
  duplicate or failed coordinator or handler registration, or unavailable
  recovery exits before Incident Bundle routes, listeners, readiness, or work
  are published; no functional-option or alternate-constructor path bypasses
  the facade.
  - Verifies: REQ-01-637
- **AC-493**: When Incident Portability is unclaimed, no Incident Bundle route
  is exposed, no Incident Bundle handler is registered or invoked, and absence
  of an Incident Portability-specific runner does not fail composition.
  - Verifies: REQ-01-637
- **AC-494**: The stable Incident Bundle handler is registered before
  recovery; new and recoverable jobs execute only through named job-ID dispatch
  and do not execute before dequeue-gate activation; no nil-receiver,
  anonymous-work, inline, or unmanaged-goroutine path executes.
  - Verifies: REQ-01-638
- **AC-495**: Named dispatch failure, recovery failure, or post-publication
  runner loss leaves the durable job queued or recoverable and closes
  readiness/admission as applicable; restart produces exactly one terminal
  result without duplicate incident, descriptor, publication, membership,
  audit, or terminal-result effects.
  - Verifies: REQ-01-450, REQ-01-638, REQ-01-641
- **AC-496**: Catalog construction rejects each closed invalid class in
  REQ-01-639, two valid builds produce the same FK-safe order, and every
  required version `2` core path has exactly one declared
  consumer or validator and one declared stable-identity invariant.
  - Verifies: REQ-01-635, REQ-01-639..REQ-01-640
- **AC-497**: For each current source-owner family, at least one fixture that is
  valid JSON and database-convertible but violates a named REQ-01-640 semantic
  invariant fails with `incident_bundle_import_rejected`,
  `reason_code='source_family_invalid'`, and the exact safe
  `source_family_id` and `invariant_id`, and commits no visible state. Missing
  and duplicate stable identities select the failing path's declared
  `<family_id>.source_identity_admitted` invariant independently of descriptor
  invariant order.
  - Verifies: REQ-01-449, REQ-01-486, REQ-01-640, REQ-01-642
- **AC-498**: Duplicate stable identities, cross-incident references,
  cross-family orphans, affected-row mismatches, and a `tags.ndjson` catalog
  unequal to the distinct imported record-tag names fail closed rather than
  being ignored or merged. Unknown paths and undeclared stable-identity
  invariants fail internally without defaulting to another invariant.
  - Verifies: REQ-01-639..REQ-01-640
- **AC-499**: Every source actor referenced by imported state has exactly one
  valid descriptor; a missing, malformed, or duplicate descriptor fails; and a
  successful import creates no login-capable user, credential, provider
  binding, deployment role, incident membership, or session from actor or
  bundle contents.
  - Verifies: REQ-01-443..REQ-01-444, REQ-01-640, REQ-01-642
- **AC-500**: Failure or cancellation before, during, or after any preparation,
  staging, owner apply, aggregate validation, revision repair, attribution,
  projection rebuild, initial-administration, audit, or publication phase
  leaves no visible incident, membership, preference, successful audit,
  projection, terminal-success result, or final object reference; staged bytes
  are absent or remain non-visible and retry-safe.
  - Verifies: REQ-01-448..REQ-01-450, REQ-01-609, REQ-01-641
- **AC-501**: Production incident import contains no descriptor-driven relation
  interpolation and no generic `ON CONFLICT DO NOTHING`; each port uses fixed
  owner-controlled persistence and affected-row equality, while shared
  Incident Portability utilities are limited to admitted bounded codec,
  canonicalization, safe-value, and actor-remap behavior.
  - Verifies: REQ-01-639
- **AC-502**: Unsupported-version and source-family failures expose only their
  exact safe reasons and allowed details, and representative failures prove
  that HTTP responses, job results, logs, telemetry, readiness,
  administrative summaries, and operator output contain none of the forbidden
  imported values or internal topology in REQ-01-642. Representative Incident
  Bundles internal failures return and retain only the exact generic internal
  error tuple and contain no injected SQL, path, storage reference, credential,
  constraint, or upstream message.
  - Verifies: REQ-01-486, REQ-01-609, REQ-01-642
- **AC-503**: Permuting archive-member order or source-row order without
  changing semantic content produces the same selected codec, catalog order,
  invariant outcome, public reason, visibility result, and deterministic
  export.
  - Verifies: REQ-01-441..REQ-01-442, REQ-01-639, REQ-01-641..REQ-01-642
- **AC-504**: The typed traceability projection contains every
  REQ-01-635..REQ-01-646-to-AC mapping and every AC-487..AC-508-to-verification
  mapping with no orphan requirement, ungrounded criterion, missing owner, or
  duplicate active verification row.
  - Verifies: REQ-01-643
- **AC-505**: Make-owned generation and drift checks prove that every affected
  generated artifact equals its authored input, carries required generator
  provenance, and satisfies generated-artifact policy with no manually edited
  generated root or dependency lockfile.
  - Verifies: REQ-01-643
- **AC-506**: Version `2` is the sole admitted Incident Bundle version. The
  active tree contains no compatibility registry, retired-version reader,
  translator, conversion utility, runtime flag, fallback codec, dual reader,
  successful-retired-version telemetry, or legacy-specific release-evidence
  machinery.
  Generic version-aware interfaces do not admit another version, and a future
  version requires a later coordinated owner, catalog, implementation, and
  verification revision.
  - Verifies: REQ-01-635..REQ-01-636
- **AC-507**: `incident.json`, `actors.ndjson`,
  `reference_pack_refs.json`, and each admitted `ext/**` payload reject their
  closed malformed, duplicate, mismatched, unknown, or unclaimed cases, invoke
  exactly their admitted consumer, and publish no unauthorized state.
  - Verifies: REQ-01-635, REQ-01-640, REQ-01-642
- **AC-508**: For admitted version `2`,
  `data/saved_views.ndjson` is always present and contains only the exact
  eleven-member REQ-01-644 row. Deterministic export covers private, shared,
  system, and zero-row incidents without relation-derived fields. Preparation
  rejects aliases, unknown, missing, duplicate, wrongly typed, null,
  noncanonical, malformed, multivalue, trailing-content, blank, and over-bound
  input with the exact safe Saved Views invariant. Successful import maps
  private/shared runtime ownership to the target submitter while preserving
  source ownership and actor attribution for re-export, preserves null system
  ownership, applies each row exactly once through fixed SQL, validates all
  seven invariants against admitted input inside the final transaction, and
  synthesizes no authorization state. Every forced failure leaves no visible
  incident, membership, preference, audit, projection, attribution, final
  object, or terminal-success artifact.
  - Verifies: REQ-01-639..REQ-01-640, REQ-01-642, REQ-01-644..REQ-01-646
- **AC-509**: The physical `records` relation and its table-local object family
  have exactly one Records owner; `change_sets`, `change_set_mutations`,
  `record_revisions`, and `record_history_entry_refs` have exactly one
  Revisions owner; recovery, Incident Portability, schema ownership, and
  authored relation identities agree without a runtime alias or historical
  migration rewrite.
  - Verifies: REQ-00-067, REQ-01-649, REQ-02-262
- **AC-510**: Direct Records evidence proves supplied and generated record
  identity, default and supplied row version, attribution, UTC normalization,
  single and batch load behavior, optional locking, delete tuple, version
  advance, and one missing-envelope outcome. Records performs no history,
  authorization, projection, publication, transport, network, or object-store
  side effect.
  - Verifies: REQ-00-067, REQ-01-649, REQ-02-262
- **AC-511**: Transactional and nontransactional record-target resolution
  return only incident ID, record type, deletion state, and row version;
  deleted rows remain resolvable to trusted internal callers; the missing-row
  outcome is the Records missing-envelope outcome; and route owners retain
  authorization, hidden outcomes, and public error mapping. No view-schema
  argument participates in target resolution.
  - Verifies: REQ-00-067, REQ-01-649
- **AC-512**: Destructive envelope locking copies, sorts, and deduplicates
  record IDs, uses caller-owned `FOR UPDATE NOWAIT`, succeeds for empty input,
  distinguishes missing from SQLSTATE `55P03` contention, identifies the
  contended record, and returns no partial protected-set result.
  - Verifies: REQ-00-067, REQ-01-649
- **AC-513**: Legal Records rows import and deterministically re-export for
  admitted version `2`; exact-shape, identity, type, version,
  timestamp, actor, delete-tuple, duplicate, incident-scope, missing-subtype,
  incompatible-subtype, duplicate-subtype, and reverse-orphan cases each fail
  with the exact closed Records invariant and safe public error, and every
  failure leaves no visible state.
  - Verifies: REQ-01-651, REQ-02-262
- **AC-514**: Revisions owns the `DeleteRestoreSource` interface and complete
  catalog; source owners construct concrete fixed-SQL adapters; application
  assembly is the join point; every current record type resolves exactly once;
  missing, duplicate, unknown, typed-nil, invalid-view, or nondeterministic
  contributions fail before publication; and adapters perform no peer
  orchestration or descriptor-driven SQL.
  - Verifies: REQ-00-067, REQ-01-650, REQ-02-262
- **AC-515**: The Records Reporting provider emits deterministic active
  current-envelope facts from a caller-supplied transaction, excludes
  soft-deleted envelopes, preserves its admitted provider key, family, path,
  value, and support-reference behavior, and leaves materialization,
  redaction, release, and publication with Reporting.
  - Verifies: REQ-01-649
- **AC-516**: Every helper formerly under Records test support has one
  source-derived semantic owner and one registered support disposition; only
  shared application support starts cross-owner services; runtime and support
  security scans retain exact coverage; selectors retain their intended tests;
  and no Records compatibility support root or forwarding alias remains.
  - Verifies: REQ-01-649
- **AC-517**: Assessment creation omits support references as an empty unordered set or admits one through 64 closed `add_record_ref` actions. Multiple heterogeneous active visible same-incident first-class targets are returned by stable record identity; duplicates yield one logical `supported_by` link from the new assessment to each target. Empty, oversized, remove, malformed, foreign, hidden, deleted, non-record, client-confidence, routing-metadata, or unknown-member input fails atomically with no assessment source, record envelope, link, revision, projection, idempotency-success, or Collaboration effect.
  - Verifies: REQ-01-311, REQ-01-332..REQ-01-334
- **AC-518**: Existing-row add and remove attempts for `assessment.support_refs`, whether submitted through direct patch, inspector, bulk, generated-client, or conflict-resolution form, both fail with `400`, `error.code='invalid_mutation_payload'`, `error.details.field='assessment.support_refs'`, and `error.details.reason_code='unsupported_field_key'`. An unknown field name retains `error.details.field='field_key'`. Every rejection leaves source, links, row version, change set, history, projection, idempotency success, and `record_changed` publication unchanged.
  - Verifies: REQ-01-310, REQ-01-335, REQ-03-254
- **AC-519**: Assessment discovery contains every retained feature exactly once, contains neither `assessment.support_refs.manage` nor `evidence.refs.manage`, and contains exactly one `create_related.assessment` with the Core 01 route, role, mutation, confirmation, seed, success, and failure contract. The follow-on workflow seeds only subject reference and type, preserves the original selection on cancel, failure, and success, keeps an editable draft on failure, appends a distinct assessment on success, copies no other field or support reference, and creates no succession relation. Timeline-only candidate discovery uses stable record IDs through the existing query route, while readable non-Timeline references remain intact.
  - Verifies: REQ-01-616..REQ-01-617, REQ-03-254, REQ-03-304..REQ-03-305
- **AC-520**: Assessment create and follow-on submission re-evaluate current incident membership, minimum `editor` role, subject eligibility, and every support target. Viewer, non-member, deployment-admin-only, capability-lost, stale-target, hidden-target, deleted-target, and cross-incident cases commit nothing, publish nothing, preserve the original selection and editable draft where applicable, and disclose no hidden target existence.
  - Verifies: REQ-01-332, REQ-01-334, REQ-03-304..REQ-03-305

- **AC-409**: Whole-incident portability preserves the boundary between incident data and deployment-local auth state: the exported bundle contains no login-capable local users, local-account credential lifecycle state such as password-hash state, active or pending TOTP state, bootstrap-token lookup state, auth bindings, bootstrap-completion markers, active sessions, active memberships, deployment-admin flags, deployment-local administrative audit state including deployment and incident-membership administrative audit events, or idempotency state; importing that bundle into another deployment does not synthesize any of those states without explicit deployment-local administrative action; and historical actors import only as inert imported actors or historical descriptors, never as login-capable principals.
  - Verifies: REQ-02-204, REQ-02-249, REQ-04-038

### 9.12 Additional Base Profile criteria for deployment configuration contract

- **AC-294**: With no selector override, the deployment loads `/etc/cartulary/config.toml`; when `CARTULARY_CONFIG_FILE` is set to an alternate absolute path, that file is selected instead; after file load, `CARTULARY__ROOTS__BACKUP_STORAGE__PATH=/srv/cartulary/backups` overrides `roots.backup_storage.path`; an unknown file key or unknown `CARTULARY__...` overlay key fails closed with `invalid_deployment_config`; and adopted subsystem keys are accepted only inside namespaces adopted by an adopted subsystem NLSpec.
  - Verifies: REQ-01-455, REQ-04-058, REQ-04-066..REQ-04-071, REQ-04-077, REQ-04-110, REQ-04-111
- **AC-295**: Required runtime-root keys are present and use the standardized binding model; `deployment_profile='disconnected'` rejects `binding_kind='managed_service'` for `roots.database_storage`, `roots.object_storage`, or `roots.backup_storage`; `roots.reference_pack_storage`, `roots.temporary_work`, and `roots.export_outputs` reject any binding kind other than `filesystem_root`; and `roots.backup_storage` accepts only `filesystem_root` in `disconnected` and only `filesystem_root` or `managed_service` in `on_prem` or `cloud`.
  - Verifies: REQ-01-455, REQ-04-058, REQ-04-069, REQ-04-071..REQ-04-073, REQ-04-077
- **AC-296**: Relative roots, `~`, shell-variable forms, empty roots, NUL, lexical `.` or `..`, overlapping configured filesystem roots after canonicalization, and non-writable startup roots fail closed with `invalid_deployment_config` and the appropriate `reason_code`. Real owner entry-point matrices prove that every filesystem effect rejects absolute or non-canonical child references, traversal, backslashes, NUL, normalization collisions, child/final symlinks, special objects, root replacement, and cross-root publication; hostile operation paths use the owning route/job error, post-ready root failure uses an owner storage/dependency failure, cleanup leaves no partial publication, managed services create no local fallback, and diagnostics disclose no raw root or hostile path.
  - Verifies: REQ-01-456, REQ-04-059, REQ-04-074..REQ-04-075, REQ-04-077
- **AC-297**: The canonical disconnected example using `/var/lib/cartulary/postgres`, `/var/lib/cartulary/object-store`, `/var/lib/cartulary/backups`, `/var/lib/cartulary/reference-packs`, `/var/lib/cartulary/tmp`, and `/var/lib/cartulary/exports` validates as a correct disconnected deployment configuration; omission of any required runtime-root key remains invalid at runtime and is not satisfied by hidden defaults.
  - Verifies: REQ-01-455, REQ-04-058, REQ-04-067, REQ-04-069, REQ-04-071..REQ-04-076
- **AC-298**: Invalid deployment configuration prevents HTTP listeners, WebSocket listeners, and background-job runners from starting; startup fails non-zero; and the surfaced error family is `invalid_deployment_config` with per-item `path`, `reason_code`, and `message`. When the Enterprise Authentication Extension Profile is claimed, provider-manifest path, encoding, schema, referenced-file, secret-reference, and reconciliation failures are deployment-configuration failures under that same pre-listener gate. When the Network Flow Activity Extension Profile is claimed, missing cursor key-ring configuration, duplicate cursor key IDs, no active cursor key, multiple active cursor keys, malformed key IDs, unresolved cursor `secret_ref_v1` values, unsupported cursor key material, invalid cursor rotation state, missing safe-digest key-ring configuration, duplicate safe-digest key IDs, no active safe-digest key, multiple active safe-digest keys, malformed safe-digest key IDs, unresolved safe-digest `secret_ref_v1` values, unsupported safe-digest key material, safe-digest key material reused for another purpose, invalid safe-digest rotation state, and fixture-only safe-digest key use outside a harness-owned runtime are deployment-configuration failures before any listener or worker starts. `/healthz` remains process liveness, while `/readyz` emits the structured readiness state required by REQ-04-078 and returns `200` only when active deployment dependencies are ready. When an adopted OpenTelemetry NLSpec is active, invalid `telemetry.*` configuration surfaces `reason_code='invalid_telemetry_config'`; before adoption, `telemetry.*` keys are unknown keys.
  - Verifies: REQ-04-066, REQ-04-077..REQ-04-078, REQ-04-110, REQ-04-111, REQ-04-115..REQ-04-122, REQ-04-128..REQ-04-134

- **AC-343**: On a fresh deployment with zero active deployment admins, no bootstrap-completion marker, `bootstrap.first_admin_manifest_path` set to a valid manifest path, and a valid `cartulary.bootstrap_admin.v1` manifest at that path, startup succeeds and listeners become available only after bootstrap completes; exactly one local user is created with `is_deployment_admin=true`, `is_active=true`, and `mfa_required=true`; no incident membership is created; and the same commit persists one deployment-local bootstrap-completion marker plus one deployment-local administrative audit event. A later valid password login for that user returns `401 error.code='mfa_setup_required'` until TOTP setup completes.
  - Verifies: REQ-01-121, REQ-01-530..REQ-01-536, REQ-02-007..REQ-02-008, REQ-02-202, REQ-02-246, REQ-04-028, REQ-04-038, REQ-04-087..REQ-04-090
- **AC-344**: On a fresh deployment with zero active deployment admins and no bootstrap-completion marker, a missing bootstrap path, unreadable bootstrap file, non-regular, symlinked, changed-during-read, or larger-than-`1048576`-byte bootstrap file, malformed JSON manifest, schema-invalid manifest, or manifest email that conflicts with an existing local user causes startup to fail before any HTTP listener, WebSocket listener, or background-job runner starts; no partial bootstrap-created user, no partial bootstrap-completion marker, and no incident membership are left behind; and the surfaced error family is `invalid_deployment_config` using the exact bootstrap-specific `reason_code`.
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
- **AC-362**: `GET /api/v1/view-schemas/{view_schema_id}` exposes `sort_fields` for every current-profile workbook surface; the Timeline v2 field entries for `timeline.activity_utc_text`, `timeline.activity_local_text`, and `timeline.date_entered_text` expose hidden sort fields rather than timestamp scalar write targets; and `record_id` does not appear in any client-sortable whitelist.
  - Verifies: REQ-01-286, REQ-01-310, REQ-01-312, REQ-01-323, REQ-01-326, REQ-01-328, REQ-01-329, REQ-01-331, REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-499, REQ-01-503..REQ-01-509
- **AC-363**: Header sort on a visible non-sortable collection field such as `note.tags` or `decision.support_refs`, or on hidden inspector-side Timeline collection fields such as `timeline.host_refs` and `timeline.tags`, does not change row order and does not synthesize a client sort.
  - Verifies: REQ-01-310, REQ-03-223
- **AC-364**: Any grouped workbook surface whose active `view_schema` declares `grouping_fields`, including `comm_log`, `handoff`, `status_review`, and `lesson`, and, when exposed, `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1`, exposes a grouping control offering exactly `Group: None` plus the surface's declared `grouping_fields` in discovery order and no undeclared key; uses derived non-record group headers keyed only by the active `group_by` and matching `group_values[group_by]`; exposes the grouped grid as a `treegrid` with exactly one outline level; renders every visible group header as a level-one parent row whose row-level expanded state matches its visible, keyboard-operable, accessible-name-bearing toggle; omits expanded state from non-parent rows; keeps mutation targets row-only while the current row sort continues to apply within each group; and uses the generic comparator order on non-Timeline grouped surfaces while Timeline alone uses the explicit override.
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

- **AC-398**: `operator backup create` can create one successful retained `backup_set` with one `backup_set_id`, one `consistency_point_at`, default `verification_state='unverified'`, `last_verified_restore_at=null`, `retained_until >= created_at + 30 days`, readable Postgres and object-store restore artifacts or anchors, and matching persisted integrity proofs; a deployment-local operator can determine structured metadata for the most recent successful retained `backup_set`, including `backup_set_id`, `consistency_point_at`, `created_at`, `retained_until`, `verification_state`, `last_verified_restore_at`, `postgres_restore_anchor`, and `object_store_restore_anchor`; `verification_state` uses only `unverified`, `verified`, or `failed`; `last_verified_restore_at` is `null` only while `verification_state='unverified'`; the latest successful retained `backup_set` has `consistency_point_at` no older than 24 hours; each successful retained `backup_set` shows `retained_until >= created_at + 30 days`; a metadata row is not represented as successful unless the required Postgres artifact, object-store artifact, and integrity proof are still readable from backup storage and match the persisted proofs; and a failed new backup candidate is not published as successful, does not change latest-successful-retained selection, and cannot replace, invalidate, or sort ahead of the prior successful retained backup while that prior backup still satisfies freshness, retention, artifact, and proof checks.
  - Verifies: REQ-01-572..REQ-01-573, REQ-01-596
- **AC-399**: Restoring the latest successful retained `backup_set` into a fresh environment restores Postgres and object-store contents from that same `backup_set`, rebuilds projections, opens at least one incident normally, executes at least one built-in workbook query when incident data is present, and preserves authoritative `incident_id`, `record_id`, `row_version`, change-set count, blob hashes, and evidence/blob lifecycle consistency.
  - Verifies: REQ-01-571, REQ-01-575, REQ-01-423..REQ-01-424, REQ-01-577
- **AC-400**: If the selected `backup_set` is missing a required Postgres artifact, required object-store artifact, or required checksum or integrity proof for the chosen backup mechanism, restore fails before the target environment becomes visible or ready.
  - Verifies: REQ-01-575..REQ-01-576
- **AC-401**: Full restore verification runs in an isolated environment at least every 7 days and after a change to the backup mechanism, `roots.database_storage`, `roots.object_storage`, or `roots.backup_storage` binding; the implementation records or derives a non-secret verification-basis digest for those mechanism and root-binding inputs; a deployment-local due-verification runner selects and executes backups due by age or basis change rather than relying only on a manual latest-verification command; a successful verification sets `verification_state='verified'` with non-null `last_verified_restore_at`; a failed verification sets `verification_state='failed'` with non-null `last_verified_restore_at`; and a failed verification is never represented as verified or ready.
  - Verifies: REQ-01-572, REQ-01-578
- **AC-402**: The public route inventory under `/api/v1/` and `/ws/v1/` exposes no backup, restore, restore-verification, or backup-inspection family; no browser, workbook, incident, common-job, HTTP, WebSocket, or session-authorized surface can create backups, inspect retained backup metadata, restore backups, or run restore verification; recovery CLI invocation is not classified as a runtime `deployment_admin` capability; and CSRF or browser-origin controls are inapplicable because the recovery CLI has no browser or HTTP surface.
  - Verifies: REQ-01-570, REQ-04-106
- **AC-428**: Operator recovery conformance evidence maps the implementation-owned executable or wrapper to exactly the five logical commands `operator backup inspect latest`, `operator backup create`, `operator restore latest`, `operator restore-verify latest`, and `operator restore-verify due`; proves `operator backup create` follows the Core 01 admission/publication algorithm, acquires the recovery-operation exclusion boundary, and maps Postgres artifact failure, object-store artifact failure, integrity-proof failure, artifact-readback failure, attestation-write failure, publication failure, journal-write failure, and timeout to the closed operator result and error vocabulary; proves failed `backup_create` results before candidate allocation emit `backup_set_id=null` and `consistency_point_at=null`; proves failed `backup_create` results after candidate allocation identify any allocated candidate only as diagnostic state and never make it selectable for restore, inspection, restore verification, or latest-successful-retained backup selection; proves `--output` omission defaults to `json` and any non-`json` value fails; proves stdout emits exactly one `cartulary.operator_recovery_result.v1` JSON object followed by LF with no extra stdout bytes; proves omitted `--progress` emits no progress records and `--progress=jsonl` emits only `cartulary.operator_recovery_progress.v1` JSONL records on stderr with operation-closed phase tokens; proves timeout defaults, allowed ranges, per-verification `restore_verify_due` behavior, `operation_timed_out`, and exit codes `0`, `2`, `3`, and `4`; proves required `--target-config-file` handling, absolute-path validation, and `restore_latest` `--confirm-backup-set-id` equality against the selected latest retained backup; proves invalid invocation, unknown flags, interactive confirmation, operator-supplied timestamp restore, unsupported backup selectors, and any operator-supplied backup scheduler flag fail closed with the required operator error codes and reason codes; proves `restore_verify_due` orders due backups by `consistency_point_at ASC, backup_set_id ASC` and returns deterministic `no_op` output when none are due; proves restore-target preflight rejects shared database bindings, shared object-store bindings, non-fresh target database or object namespace, target listeners serving traffic, missing or invalid target markers, missing recovery keys, missing artifacts, and failed integrity proofs before target mutation; proves timed-out restore targets remain not-ready and require reinitialization; proves admitted mutating recovery operations append encrypted operator recovery journal records; proves a writable database receives the safe administrative-audit summary; and proves outputs, progress, logs, journal records, and audit summaries omit credentials, secret references, raw DSNs, endpoint hosts, bucket names, object keys, raw storage paths, recovery keys, and incident content.
  - Verifies: REQ-01-593..REQ-01-595, REQ-01-596, REQ-04-106, REQ-04-113
- **AC-535**: Collaboration requeue conformance maps exactly one implementation-owned command to `operator collaboration requeue`; proves the strict long-flag grammar, canonical non-zero UUID rule, literal absolute config path rule, config discovery precedence, timeout default/range, help-only exception, exact v2 member set/order/LF, closed code/reason/exit registry, caller-cancelled versus timeout behavior, and post-commit delivery exception; proves no v1 or parser alias remains; locks and validates the quarantined cursor and every pending intent; rejects an unrepaired payload without mutation; resets only the declared retry/quarantine fields; preserves payload/event/dispatch/sequencing identity; appends one safe raw non-public operator journal row atomically; rolls back all effects on journal or transaction failure; reports commit failure as outcome unknown; permits exactly one concurrent winner; makes a second invocation reject; closes all resources; and exposes no public route, browser, workbook, WebSocket, job, incident action, public audit projection, secret, raw DSN, endpoint, SQL, constraint, payload, record content, object key, storage path, or upstream error text.
  - Verifies: REQ-00-068, REQ-01-655, REQ-03-307, REQ-04-151
- **AC-536**: Object Store initialization conformance preserves the existing logical command and successful v1 schema, creates or confirms only the configured bucket, maps known typed configuration and adapter failures deterministically, maps every untyped failure to `dependency_unavailable`, contains no error-string classifier, closes acquired resources, and emits no endpoint, host, bucket, key, storage reference, credential, secret reference, raw DSN, path, constraint, or upstream error text.
- **AC-537**: Claimed Network Flow lifecycle fixtures prove the private cleanup dispatcher starts after readiness, executes one immediate sweep, follows the five-minute non-overlapping coalesced cadence, paces `has_more` continuation by five seconds, retries transient failure without changing product behavior, treats unexpected loop loss as fatal component loss, stops before borrowed dependencies in reverse shutdown, and exposes no public route, common Job, or operator control; the unclaimed profile starts none.

- **AC-537**: Database migration conformance proves that PostgreSQL connectivity, root and service binding resolution, generic query and transaction ports, and PostgreSQL telemetry remain in the PostgreSQL adapter while immutable migration-source inspection, typed apply-to-head execution, exact migration-history and lineage classification, readiness, remediation, evidence, and migration recovery metadata exist only under the database-migrations owner; that version-targeted apply and rollback exist only behind the Testing Harness disposable migration-scratch capability; and that no compatibility alias, forwarding package, duplicate implementation, generic Goose command surface, filesystem discovery, process-global migration filesystem, process-global Goose logger, production targeted operation, legacy source or status type, or migration log-file environment path remains. Source-construction and zero-source fixtures prove rejection of nil filesystems, invalid or escaping roots, missing lineage metadata, empty or malformed catalogs, unexpected entries, invalid filenames, duplicate or non-contiguous versions, malformed markers, unsupported directives, and unbalanced statement blocks before database access, including mutation-after-construction evidence for immutable catalog bytes. Production apply fixtures prove cross-process serialization through advisory lock `4097083626`, one-second cancellation-aware acquisition intervals with a five-minute ceiling, locked initial classification, validating same-session provider classification, final locked classification, bounded detached unlock within 30 seconds on every post-acquisition exit, primary-error precedence, an invocation-local provider with discarded logging and the global Go registry disabled, and continued caller usability of the borrowed database handle. State matrices prove that malformed ledger structure precedes lineage mismatch; structurally valid nonzero histories have exactly the expected singleton lineage; pristine, behind, current, and ahead states receive their specified outcomes; nil database capabilities fail closed; and neither ledger repair nor a historical-lineage bridge occurs. Migration failures expose only the closed safe reason set, remain `errors.Is`-compatible for context cancellation, preserve the exact remediation-report-v1 JSON shape and single-object-plus-LF migrate framing, and expose no Goose or upstream error text, DSN, credential, secret reference, database root, server identity, SQL text, bind value, filesystem path, environment value, constraint, or endpoint through diagnostics, logs, telemetry, remediation, evidence, or operator output. Migration-history evidence fixtures prove that current output uses only `cartulary.migration_history_evidence.v2`, that version 1 is rejected as current evidence, that `manifest.path` and every replacement or path-derived locator are absent, and that relocating the repository or executable without changing logical inputs leaves the evidence invariant. The evidence version 2 cutover preserves the logical Operator command, authorization boundary, database resource lifecycle, every unrelated evidence member and finding rule, single-object-plus-LF stdout framing, secret-safe stderr behavior, and exit-code mapping; it provides no compatibility flag, dual emission, or version 1 translation. Recovery contribution identity, operator framing, server diagnostic mapping, and PostgreSQL telemetry privacy pass their exact owner-routed evidence.
  - Verifies: REQ-00-069, REQ-01-657
  - Verifies: REQ-01-656, REQ-04-152
- **AC-538**: Workbook application-shortcut evidence proves that `Ctrl/Cmd+K`, `Space`, and `Alt+H` are consumed only while grid navigation owns focus on an eligible committed cell or row; that link capability and inspector-group availability come from owner-declared semantic state rather than visible text; that Evidence opens explicitly and focuses its sole previewable item, list, or empty state; that History opens explicitly; and that editor, menu, popover, dialog, inspector, group-row, draft-row, and no-row cases perform no application action and do not suppress browser behavior.
  - Verifies: REQ-03-220
- **AC-403**: `roots.backup_storage` is present in the effective deployment configuration; the disconnected binding uses `binding_kind='filesystem_root'`; the on-prem or cloud binding uses only `filesystem_root` or `managed_service`; `/var/lib/cartulary/backups` is the canonical disconnected example path; `roots.export_outputs` and `roots.temporary_work` are not treated as authoritative backup roots; and backup artifacts or restore-verification extracts that carry incident data remain on encrypted storage, with the current filesystem-backed realization proving this through authenticated encrypted artifact envelopes and fail-closed missing-key behavior.
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

**REQ-04-111**
Core 04 owns the deployment-configuration artifact, discovery, environment-overlay grammar, unknown-key rejection, validation error envelope, and fail-closed startup behavior. An adopted Cartulary subsystem NLSpec MAY add one closed top-level deployment-configuration namespace only when that NLSpec defines exact keys, types, defaults, bounds, omitted behavior, explicit-`null` behavior, validation errors, secret and redaction handling, cross-key rules, and startup failure behavior for that namespace.

Before adoption, a proposed subsystem NLSpec does not alter the accepted deployment-configuration schema. Unknown keys outside Core 04 keys and adopted subsystem namespaces remain invalid. Environment overlays for adopted subsystem namespaces MUST use the Core 04 `CARTULARY__` overlay grammar and MUST remain subject to unknown-key rejection. Implementation-support guides, drafts, examples, and appendices MUST NOT widen the deployment-configuration schema.

The implementation MUST realize this ownership as an owner-neutral configuration kernel.
Core 04 mechanics parse the artifact, apply overlays, reject unknown keys, order
diagnostics, and admit an immutable snapshot. Each adopted namespace owner supplies its
closed paths, wire decoder, overlay application, semantic validation, projection, and
clone behavior through an explicit statically assembled contribution. Core 04 MUST NOT
embed owner-specific wire DTOs or semantic switches. Package initialization, mutable
global registration, runtime or source-tree scanning, reflection-based owner discovery,
and dynamic plugin loading are forbidden.

For adopted subsystem namespaces that need deployment-local secret material, Core 04 owns the reusable `secret_ref_v1` deployment-configuration primitive. A `secret_ref_v1` value MUST be an object with exactly `kind` and `name`. `kind` MUST be exactly `env` in the current profile. `name` MUST match `[A-Za-z0-9][A-Za-z0-9_.-]{0,63}`. The normalized environment suffix MUST uppercase ASCII letters, retain digits, replace every non-alphanumeric run with one underscore, trim leading and trailing underscores, and reject an empty normalized result. For `kind='env'`, the normalized suffix `<REF>` selects exactly `CARTULARY_SECRET_<REF>`. `secret_ref_v1` values are deployment-local secret references, not secret values; they MUST NOT be accepted from browser state, incident records, workbook rows, public API requests, OTel SDK environment variables, declarative config, or generated test fixtures outside a harness-owned runtime. Resolution MUST occur after deployment-configuration parsing and before readiness for any subsystem that requires the secret. Missing, empty, syntactically invalid, or unresolved secret references MUST fail closed before readiness with `invalid_deployment_config`; diagnostics MAY include the safe reference `name` and config path, but MUST NOT include raw environment variable values, transformed secret values, endpoint credentials, header values, HMAC keys, or derived secret material.
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
The deployment configuration file MUST declare `config_schema_id = "cartulary.deployment_config.v2"`, required `deployment_profile` with one of `disconnected`, `on_prem`, or `cloud`, required `application.public_origin`, and required `revisions.conflict_token_key_ring_manifest_path`. Version `1` is unsupported after this adoption and MUST NOT activate a compatibility default or authentication-master-derived conflict-token key. Keys other than those defined by the selected configuration schema version are invalid.
Profiles: base
Verified by: AC-294, AC-295, AC-297, AC-526

**REQ-04-070**
After file load, environment variables prefixed `CARTULARY__` MUST overlay nested keys by splitting segments on `__`, lowercasing each segment, and joining them with dots. Overlay keys that do not map to declared deployment-configuration keys are invalid. `CARTULARY_CONFIG_FILE` is selector-only and MUST NOT participate in that overlay mapping.
Profiles: base
Verified by: AC-294

### 12.3 Key registry and binding model


#### Application public origin

**REQ-04-110**
The stable deployment-configuration key for the browser application origin MUST be `application.public_origin`. It MUST be an absolute `http` or `https` origin and MUST NOT include userinfo, path, query, or fragment. The public WebSocket implementation MUST validate cookie-authenticated browser `Origin` values against this configured origin before joining an incident-scoped stream.
Profiles: base
Verified by: AC-131, AC-294, AC-298

#### Revisions conflict-token key ring

**REQ-04-147**
The stable base-profile configuration key is
`revisions.conflict_token_key_ring_manifest_path`. It is required and has no
default. It MUST be an absolute normalized POSIX path to one deployment-local
regular file of at most `65536` bytes. Every path component is opened without
following symbolic links, and validation uses one bounded open. The file is
UTF-8 JSON with exact top-level members `schema_id`, `algorithm`, and `keys`.
`schema_id` is exactly
`cartulary.revisions_conflict_token_key_ring.v1`; `algorithm` is exactly
`aes_256_gcm_v1`; and `keys` contains `1..8` exact key objects.

Each key contains `conflict_token_key_id`, `state`, and `secret_ref`. The key ID
matches `[A-Za-z0-9][A-Za-z0-9._-]{0,63}` and is unique. `secret_ref` is the
Core 04 `secret_ref_v1` primitive and resolves to exactly `32` bytes of key
material. Exactly one key has `state='active'` and omits rotation timestamps.
Every other key has `state='decrypt_only'`, required canonical UTC
`deactivated_at`, and required canonical UTC `retire_at` not earlier than
`deactivated_at + 31 minutes` and later than startup time. Unknown or duplicate
members, duplicate key IDs, duplicate secret references, zero or multiple
active keys, invalid state/timestamps, expired decrypt-only keys, unsupported
key material, or unresolved secrets fail deployment validation.
Profiles: base
Verified by: AC-298, AC-526

**REQ-04-148**
Conflict-token key material has secret purpose
`revisions_conflict_token`. Raw or derived material MUST NOT be reused for the
authentication master, Network Flow cursor or safe-digest keys, enterprise-auth
provider secrets, telemetry headers, storage credentials, recovery keys,
bootstrap material, or another purpose. New tokens use only the active key.
A decrypt-only key may validate only a v3 token issued before that key's
`deactivated_at`, before the token's own expiry, and before `retire_at`.
Changing material in place, reusing a key ID, extending token TTL to retain an
old key, accepting v2, falling back to a hard-coded key, or accepting fixture
material outside a harness-owned runtime is forbidden.
Profiles: base
Verified by: AC-526

**REQ-04-149**
The v3 sealed payload uses AES-256-GCM with a fresh `12`-byte CSPRNG nonce and
additional authenticated data that binds exactly
`cartulary.conflict-token.v3` and `conflict_token_key_id`. Raw claims, raw or
derived key material, secret-reference values, nonce-generation failures, and
cryptographic failure details MUST NOT appear in public APIs, browser state,
incident data, workbook rows, logs, diagnostics, readiness, telemetry,
administrative audit, or operator output. Configuration, key resolution, and
key-ring admission complete before HTTP/WebSocket listeners or background-job
runners start. Any failure exits non-zero under `invalid_deployment_config`.
Profiles: base
Verified by: AC-298, AC-526

#### Network Flow cursor key ring

**REQ-04-128**
When the Network Flow Activity Extension Profile is claimed, deployment configuration MUST provide a deployment-local Network Flow cursor key ring before readiness. The key ring MUST contain exactly one active key and MAY contain inactive decrypt-only keys retained only for bounded rotation. Each key MUST have a non-secret `cursor_key_id` that is stable for the lifetime of that key material and a `secret_ref_v1` binding for the cryptographic key material. `cursor_key_id` values MUST be ASCII, non-empty, no more than `64` bytes, and MUST match `[A-Za-z0-9][A-Za-z0-9_.-]{0,63}`. Duplicate key IDs, more than one active key, no active key, invalid key IDs, missing or unresolved key material, weak or unsupported key material, unsupported algorithms, or a decrypt-only key without a bounded retirement instant MUST fail before readiness with `invalid_deployment_config` and a Network Flow cursor key reason code from REQ-04-077. Raw key material and derived key material MUST NOT appear in public APIs, browser state, incident data, workbook rows, provider manifests, ordinary generated fixtures, logs, diagnostics, readiness output, telemetry attributes, audit events, or error details.
Profiles: network_flow_activity
Verified by: AC-298, AC-375

**REQ-04-129**
Network Flow cursor tokens MUST be sealed with an authenticated-encryption construction or an implementation-equivalent primitive that provides confidentiality and integrity for the full cursor payload. The token MUST carry the non-secret `cursor_key_id` and token-format version needed to select validation material, but all route contract, actor/session binding, timestamps, table scope, continuation position, and comparator state MUST be opaque to clients. Plaintext JSON, unsigned base64 payloads, client-readable JWT/JWS-style payloads, reversible encodings without authenticated encryption, and tokens whose unauthenticated prefix can influence authorization or continuation state are nonconformant. Unknown key ID, retired key ID, malformed token, failed authentication, failed decryption, unsupported version, payload-schema failure, binding mismatch, authorization loss, or continuation-state failure MUST collapse to the Network Flow cursor-validation error selected by Core 01 and the Network Flow owner, without exposing which cryptographic or payload check failed.
Profiles: network_flow_activity
Verified by: AC-375

**REQ-04-130**
Network Flow cursor token TTL is fixed at `15` minutes from `issued_at` in the current profile. A token is valid only when `issued_at <= now < expires_at`; equality with `expires_at` is expired. New cursor tokens MUST always be sealed with the active key at issuance time. Rotation MUST introduce a new active key ID and move the previous active key to decrypt-only status; decrypt-only keys MAY validate only tokens issued before that key's deactivation and only until the token's own `expires_at`. No decrypt-only key MAY remain necessary after `deactivated_at + 15 minutes`, and removing such a key MUST NOT invalidate any conforming unexpired cursor. Reusing a `cursor_key_id` with different key material, changing key material in place, extending the TTL to preserve old tokens, or falling back to a default hard-coded key is forbidden. A running deployment MAY reject all outstanding cursors after a fail-closed key-ring repair, but it MUST do so through the same cursor-validation error and MUST NOT publish key state or secret diagnostics.
Profiles: network_flow_activity
Verified by: AC-298, AC-375

#### Network Flow safe-digest key ring

**REQ-04-131**
When the Network Flow Activity Extension Profile is claimed, deployment configuration MUST provide a deployment-local Network Flow safe-digest key ring before readiness. The safe-digest key ring is distinct from the Network Flow cursor key ring and from every other subsystem secret purpose. The same raw key material MUST NOT be reused for cursor tokens, enterprise-auth provider secrets, telemetry headers, storage credentials, recovery keys, bootstrap material, or any non-Network Flow purpose. The key ring MUST contain exactly one active key for new safe-digest emission and MAY contain inactive historical keys retained only for explicit same-key-ID correlation. Each key MUST have a non-secret `safe_digest_key_id` that is stable for the lifetime of that key material and a `secret_ref_v1` binding for the cryptographic key material. `safe_digest_key_id` values MUST be ASCII, non-empty, no more than `64` bytes, and MUST match `[A-Za-z0-9][A-Za-z0-9_.-]{0,63}`. Secret material MUST contain at least `256` bits of CSPRNG entropy after secret-reference resolution. Duplicate key IDs, more than one active key, no active key, invalid key IDs, missing or unresolved key material, weak or unsupported key material, unsupported algorithms, key-purpose reuse, or an inactive key without a bounded retirement or retention decision MUST fail before readiness with `invalid_deployment_config` and a Network Flow safe-digest reason code from REQ-04-077.
Profiles: network_flow_activity
Verified by: AC-298, AC-475

**REQ-04-132**
The adopted Core 04 name for the secret in Network Flow safe-digest computation is Network Flow safe-digest key material. Earlier draft aliases MUST NOT become public API members, deployment-configuration key names, log fields, telemetry attributes, diagnostic members, or generated fixture fields outside a harness-owned runtime. Network Flow safe digests MUST use HMAC-SHA-256 over a Network Flow owner-defined domain separator, `value_class`, and canonical value. Core 04 owns key lifecycle, key ID handling, and disclosure rules; the Network Flow owner owns the closed `value_class` set and canonical value construction. Safe digest outputs are non-secret correlation values, but they remain incident-adjacent metadata and MUST NOT be treated as authorization tokens, idempotency keys, dedupe keys, concurrency tokens, table identity, row identity, cursor state, or proof of raw-value possession.
Profiles: network_flow_activity
Verified by: AC-475

**REQ-04-133**
Every Network Flow field that carries a safe digest in a route error detail, route response, log record, telemetry record, administrative audit summary, or domain audit payload MUST carry the producing `safe_digest_key_id` in the same enclosing object. Consumers and implementation code MUST compare safe digests only when the key IDs, digest algorithm version, and value class are equal; equality across different key IDs is undefined and MUST NOT be used for authorization, deduplication, correlation, grouping, or alerting. Safe-digest rotation establishes a new correlation epoch. Previously persisted safe digests MUST NOT be rewritten, backfilled, or re-keyed during rotation. Plain SHA-256 fields such as source-content hashes and incident-authorized rejected-row `raw_value_sha256` values are not safe digests, MUST NOT carry a safe-digest key ID, and MUST remain forbidden in logs, telemetry, administrative audit summaries, readiness output, and public error details except where an adopted owner route explicitly allows them to an incident-authorized caller.
Profiles: network_flow_activity
Verified by: AC-475

**REQ-04-134**
New Network Flow safe digests MUST always be emitted with the active safe-digest key at emission time. Rotation MUST introduce a new active key ID and key material; the previous active key becomes inactive and MUST NOT be used for new log, telemetry, route, error-detail, or audit fields. Inactive keys MAY be retained only to recompute a same-key-ID digest for explicit authorized audit or fixture comparison, and MUST NOT be used to compare against records whose `safe_digest_key_id` differs. Reusing a `safe_digest_key_id` with different key material, changing key material in place, extending an old epoch by emitting new digests with an inactive key, or falling back to a default hard-coded key is forbidden. Harness-owned conformance runtimes MAY provide deterministic fixture-only safe-digest key material and key ID for immutable transcript generation; production and ordinary development deployments MUST fail before readiness if they attempt to use fixture-only key material, fixture-only key IDs, checked-in raw key values, generated fixture secrets, or inline raw safe-digest secrets instead of `secret_ref_v1`.
Profiles: network_flow_activity
Verified by: AC-298, AC-475

#### Network Flow retention and import-source hooks

**REQ-04-139**
When the Network Flow Activity Extension Profile is claimed, the Core-visible `network_flow_table` lifecycle MUST contain only `active` and `soft_deleted` in the current profile. Soft delete is terminal in v1. The current profile defines no public, administrative, operator, import, recovery, or background route for Network Flow table restore, hard delete, compaction, deleted-table inspection, deleted-table listing, or deleted-table graph/query access. An `active` table is queryable, graphable, visible in default Network Analysis table lists, and eligible for `table_scope.mode='all_active_tables'`. A `soft_deleted` table remains retained only for provenance, binding traceability, exact committed replay, and Core-governed audit correlation; it MUST NOT be queryable, graphable, included in default lists, included in `all_active_tables`, or accepted by table, row, diagnostic, graph, contributor, or indicator-link routes. Direct references to a disclosed soft-deleted table MUST fail with the Network Flow `network_flow_table_not_active` route error; references that are not discloseable to the caller MUST use the applicable Core hidden-resource behavior instead.
Profiles: network_flow_activity
Verified by: AC-477

**REQ-04-140**
Network Flow soft delete MUST invalidate every active Network Flow workspace state, query cursor, diagnostic cursor, graph contributor cursor, and ephemeral graph result whose scope includes the deleted table. Cursor invalidation MUST use the Network Flow cursor-validation contract from Core 01 and REQ-04-128 through REQ-04-130 without disclosing whether the invalidated scope was hidden, deleted, or authorization-lost to a caller who cannot see that state. A table rename MUST NOT invalidate retained table identity, source snapshot identity, or exact committed replay, but a later soft delete of that same table MUST make all existing route scopes that require active state fail as non-active or hidden. Soft-deleted tables continue to count against `network_flow.max_retained_tables_per_incident` and no longer count against `network_flow.max_active_tables_per_incident`; preview state, import staging, failed imports, cancelled imports before final commit, and rolled-back commits count against neither limit.
Profiles: network_flow_activity
Verified by: AC-477

**REQ-04-141**
Opaque import source streams, upload capabilities, preview caches, mapping approvals, validated staging rows, temporary files, object-store keys, worker-local paths, and retry artifacts used by Network Flow import are Core import-job/source artifacts, not Network Flow incident resources. Network Flow MUST NOT expose those artifacts through Network Flow routes, workspace state, domain audit payloads, route errors, diagnostics, logs, telemetry attributes, safe digests, Graph Projection adapter inputs, or generated fixture transcripts except through harness-owned redacted evidence fields explicitly declared by an adopted owner. After final commit, the durable Network Flow incident data is limited to committed table metadata, immutable accepted rows, retained diagnostics, source snapshot identity, indicator bindings, idempotency state needed for exact replay, and Core-governed domain audit occurrences. Raw source bytes and staging artifacts MUST follow the Core import owner retention and cleanup contract and MUST NOT be retained longer because the Network Flow extension claimed a profile. Failed, cancelled, preview-only, and rolled-back imports MUST NOT leave a queryable table, a retained Network Flow diagnostic set, or a public Network Flow table reference.
Profiles: network_flow_activity
Verified by: AC-477

**REQ-04-142**
The current Network Flow Activity Extension Profile makes no whole-incident purge claim. Until a future adopted Core incident-removal or incident-purge owner exists, a conforming Network Flow implementation MUST NOT publish a route, operator command, configuration key, background job, migration promise, adoption claim, or conformance result that purges all Network Flow resources for an incident. Soft delete, import-source cleanup, cursor expiry, key rotation, and job-retention cleanup do not constitute an incident purge. If a future adopted Core incident-removal owner includes Network Flow, it MUST include Network Flow owner state, idempotency records containing Network Flow references, retained diagnostics, indicator bindings, cursor invalidation state, and import-source artifacts in one owner-approved atomic removal unit while preserving Core 04 audit-retention rules. Network Flow MUST NOT independently shorten, rewrite, compact, or purge Core-governed incident or administrative audit state.
Profiles: network_flow_activity
Verified by: AC-477

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
Each root binding MUST be a typed object. `binding_kind` MUST use the closed vocabulary `filesystem_root` or `managed_service`. When `binding_kind='filesystem_root'`, `path` is required and `service_ref` MUST NOT be present. When `binding_kind='managed_service'`, `service_ref` is required and `path` MUST NOT be present. Credentials, vendor-specific connection properties, reverse-proxy settings, and other deployment-adjacent configuration remain outside the TOML file and `CARTULARY__` overlay contract except where a managed-service reference below defines an explicit environment binding.
Profiles: base
Verified by: AC-295

**REQ-04-073**
`roots.reference_pack_storage`, `roots.temporary_work`, and `roots.export_outputs` MUST use `binding_kind='filesystem_root'`. `roots.backup_storage` is governed by §12.3.3. When `deployment_profile='disconnected'`, `roots.database_storage` and `roots.object_storage` MUST also use `binding_kind='filesystem_root'`. In `deployment_profile='on_prem'` or `deployment_profile='cloud'`, those two roots MAY use `binding_kind='managed_service'`.
Profiles: base, reference_pack
Verified by: AC-295, AC-297, AC-403

For `roots.database_storage`, the file and `CARTULARY__` overlay configuration remain backend-neutral. The only deployment-configuration keys for this root are the standardized root-binding keys above; the configuration schema MUST NOT add DSN, host, port, database, username, password, TLS, or vendor-specific database keys. When `roots.database_storage.binding_kind='filesystem_root'`, omitted `path`, explicit `path=null`, present `service_ref`, or explicit `service_ref=null` are invalid according to the root-binding contract. When `roots.database_storage.binding_kind='managed_service'`, omitted `service_ref`, explicit `service_ref=null`, present `path`, or explicit `path=null` are invalid according to the root-binding contract.

For `roots.database_storage.binding_kind='managed_service'`, `service_ref` MUST contain at least one ASCII letter or digit after normalization. Normalization for managed Postgres service environment keys MUST uppercase letters, retain digits, replace every non-alphanumeric run with one underscore, trim leading and trailing underscores, and reject an empty normalized result. The normalized reference `<REF>` selects exactly one purpose-specific key: `CARTULARY_POSTGRES_<REF>_RUNTIME_DSN`, `CARTULARY_POSTGRES_<REF>_MIGRATION_DSN`, or `CARTULARY_POSTGRES_<REF>_RECOVERY_DSN`. These values are service binding material rather than deployment-configuration keys and MUST NOT be accepted through the `CARTULARY__` overlay grammar. The former `CARTULARY_POSTGRES_<REF>_DSN` input is retired and MUST NOT provide a fallback, alias, or warning-period value.

Missing required Postgres service-binding environment variables or invalid normalized `service_ref` values MUST fail closed before readiness. Diagnostics MUST identify the config path or service binding field and a registered validation reason without exposing raw DSNs, endpoint hosts, database names, usernames, passwords, service refs as storage identity, or backend URLs.

#### 12.3.1A Production PostgreSQL purpose, identity, and privilege boundary

**REQ-04-153**
Production PostgreSQL credential purposes are the closed vocabulary `runtime`,
`migration`, and `recovery`. Purpose validation MUST occur before environment
or filesystem access. After binding validation and service-reference
normalization, resolution MUST derive the selected current locator, every
unselected current locator, and the retired locator. A present retired locator
MUST fail without reading its value; otherwise any present unselected locator
MUST fail without reading its value; otherwise the resolver MUST read and
validate only the selected locator. Presence with an empty value counts as
presence for retired and unselected rejection.

The exact managed locators and filesystem children are:

| Purpose | Managed locator | Filesystem child | Required effective role |
| --- | --- | --- | --- |
| `runtime` | `CARTULARY_POSTGRES_<REF>_RUNTIME_DSN` | `postgres.runtime.dsn` | `cartulary_runtime` |
| `migration` | `CARTULARY_POSTGRES_<REF>_MIGRATION_DSN` | `postgres.migration.dsn` | `cartulary_schema_owner` |
| `recovery` | `CARTULARY_POSTGRES_<REF>_RECOVERY_DSN` | `postgres.recovery.dsn` | `cartulary_recovery` |

The retired locators are exactly `CARTULARY_POSTGRES_<REF>_DSN` and
`postgres.dsn`. A process or binding root MUST contain only its selected
purpose credential. No fallback, alias, dual read, or compatibility warning is
permitted.

The resolver failure precedence and reason codes are closed:

1. zero or unknown purpose: `unsupported_postgres_purpose`, without locator access;
2. invalid binding kind, root, or normalized service reference: `postgres_binding_invalid`;
3. retired locator present: `retired_postgres_binding_present`, without value access;
4. unselected current locator present: `cross_purpose_postgres_binding_present`, without value access;
5. selected locator absent: `postgres_binding_missing`;
6. selected value empty, unreadable, malformed, oversized, or otherwise invalid: `postgres_binding_invalid`;
7. required role identity not established: `postgres_effective_role_mismatch`.

The repeated `postgres_binding_invalid` reason intentionally hides whether the
failure arose from deployment structure or secret content. If multiple
conditions exist, the first condition above controls. Server and Operator
startup MUST map these failures through `invalid_deployment_config`;
`cmd/migrate` and Recovery commands MUST retain their existing safe
configuration-failure exit families.

For a managed binding, the supplied environment map is the complete process
environment view; a nil map selects the operating-system environment. Only the
selected key value may be retrieved. For a filesystem binding, each selected
file MUST resolve beneath the validated database root, be opened without
following any path component or final symlink, be one regular file no larger
than 65,536 bytes, be read exactly once with a bounded read, contain valid
UTF-8 with no NUL or embedded line break, permit at most one terminal LF or
CRLF that is removed, and remain non-empty. Retired and unselected file
presence MUST be checked through the same no-follow root capability without
reading bytes.

No error, log, telemetry event, stdout, or stderr may contain credential bytes,
parsed DSN fields, host, database name, username, locator, or upstream cause.

Administrators MUST provision these fixed roles outside application DDL:

- `cartulary_schema_owner`;
- `cartulary_runtime`;
- `cartulary_recovery`.

Each fixed role MUST be `NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE
NOREPLICATION NOBYPASSRLS`. A deployment migration, runtime, or Recovery login
MUST be `LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION
NOBYPASSRLS`, have exactly one membership in its corresponding fixed role with
`INHERIT FALSE, SET TRUE, ADMIN FALSE`, and have no membership path to either
other fixed role. Fixed roles MUST NOT be members of one another.

`cartulary_schema_owner` MUST own `public`, every Cartulary-authored object,
and Goose metadata; extension-managed objects are the only ownership
exception. Runtime and Recovery MUST own no object. Every newly created or
recycled physical connection MUST execute `SET ROLE` for its required fixed
role and verify its deployment login as `session_user` and its fixed role as
`current_user` before entering a pool or becoming caller-usable. A failing
connection MUST be closed. One-time pool initialization is insufficient, and
pgx pool and `database/sql` construction MUST share one identity-establishment
implementation.

Every grantable object MUST have exactly one runtime and one Recovery access
class in the adopted object-ownership manifest. Runtime classes are:

- `schema_usage`: schema `USAGE`;
- `table_read_write`: table `SELECT`, `INSERT`, `UPDATE`, `DELETE`;
- `table_append_only`: table `SELECT`, `INSERT`;
- `table_read_only`: table `SELECT`;
- `table_no_access`: no table privilege;
- `migration_ledger_read`: migration metadata `SELECT` only;
- `sequence_use`: sequence `USAGE` only;
- `sequence_no_access`: no sequence privilege;
- `view_read_only`: view `SELECT`;
- `routine_application`: routine `EXECUTE`;
- `routine_private`: no routine privilege;
- `type_use`: type `USAGE`;
- `type_no_access`: no type privilege;
- `not_applicable`: no grantable privilege, valid only for extension, trigger,
  constraint, index, operator, operator class/family, cast, or collation.

Recovery classes are:

- `schema_usage`: schema `USAGE`;
- `table_restore`: table `SELECT`, `INSERT`, `UPDATE`, `DELETE`, `TRUNCATE`;
- `table_rebuild`: table `SELECT`, `TRUNCATE`; valid only for excluded,
  rebuildable derived state whose owner-provided recovery routine reconstructs
  every row;
- `table_read_only`: table `SELECT`;
- `table_no_access`: no table privilege;
- `migration_ledger_read`: migration metadata `SELECT` only;
- `sequence_restore`: sequence `USAGE`, `SELECT`, `UPDATE`;
- `sequence_no_access`: no sequence privilege;
- `view_read_only`: view `SELECT`;
- `routine_recovery`: routine `EXECUTE`;
- `routine_private`: no routine privilege;
- `type_use`: type `USAGE`;
- `type_no_access`: no type privilege;
- `not_applicable`: the same closed non-grantable kinds as runtime.

There is no default access class. Recovery additionally receives `SET` only on
parameter `session_replication_role`. Runtime MUST NOT receive that parameter
privilege, `TRUNCATE`, sequence update, ownership, DDL, database or schema
`CREATE`, `TEMPORARY`, `REFERENCES`, `TRIGGER`, `MAINTAIN`, role administration,
schema-owner assumption, migration-ledger mutation, or private routine
execution. Recovery MUST NOT receive ownership, DDL, schema-owner assumption,
runtime-role assumption, or migration-ledger mutation.

Administrators MUST revoke database `CONNECT` and `TEMPORARY` from `PUBLIC`,
grant `CONNECT` directly only to the three deployment logins and declared
administrator-owned operational identities, grant no `TEMPORARY`, transfer
`public` ownership to `cartulary_schema_owner`, and revoke every `PUBLIC`
privilege on `public`. Default privileges for objects created by
`cartulary_schema_owner` MUST revoke future routine `EXECUTE` and future type
`USAGE` from `PUBLIC` and MUST grant no broad table, sequence, routine, or type
access to runtime or Recovery. Each migration MUST issue explicit
manifest-matching grants. Existing ACLs are not inferred from defaults.

Every environment's administrator or harness provisioning MUST create the
roles and login memberships, database grants, `public` ownership, and exact
PostgreSQL 16 prerequisites `pgcrypto` 1.3 and `citext` 1.6 in `public`. After
extension installation it MUST revoke extension-object `PUBLIC` privileges and
grant only manifest-declared extension routine/type access. Cartulary
application DDL MUST create no role, login, credential, schema, or extension.
Profiles: base, enterprise_authentication, import, incident_portability,
network_flow_activity, reference_pack, snapshot_reporting
Verified by: AC-542

For `roots.object_storage`, the file and `CARTULARY__` overlay configuration remain backend-neutral. The only deployment-configuration keys for this root are the standardized root-binding keys above; the configuration schema MUST NOT add `seaweedfs.*`, `s3.*`, bucket, endpoint, access-key, secret-key, CORS, or vendor-specific object-store keys. When `roots.object_storage.binding_kind='filesystem_root'`, omitted `path`, explicit `path=null`, present `service_ref`, or explicit `service_ref=null` are invalid according to the root-binding contract. When `roots.object_storage.binding_kind='managed_service'`, omitted `service_ref`, explicit `service_ref=null`, present `path`, or explicit `path=null` are invalid according to the root-binding contract.

For `roots.object_storage.binding_kind='managed_service'`, `service_ref` MUST contain at least one ASCII letter or digit after normalization. Normalization for managed object-storage service environment keys MUST uppercase letters, retain digits, replace every non-alphanumeric run with one underscore, trim leading and trailing underscores, and reject an empty normalized result. The normalized reference `<REF>` selects the following environment variables, which are service binding material rather than deployment-configuration keys and MUST NOT be accepted through the `CARTULARY__` overlay grammar:

- `CARTULARY_S3_<REF>_ENDPOINT` is required.
- `CARTULARY_S3_<REF>_ACCESS_KEY_ID` is required.
- `CARTULARY_S3_<REF>_SECRET_ACCESS_KEY` is required.
- `CARTULARY_S3_<REF>_BUCKET` is required.
- `CARTULARY_S3_<REF>_SECURE` is optional and defaults to `false` when omitted; when present it MUST be exactly `true` or `false`.

Missing required service-binding environment variables, invalid normalized `service_ref`, or invalid optional `SECURE` values MUST fail closed before readiness. Diagnostics MUST identify the config path or service binding field and a registered validation reason without exposing raw endpoint hosts, bucket names, access keys, secret keys, object keys, storage refs, or backend URLs. Endpoint values are redacted as endpoint material, bucket values as bucket material, and access-key and secret-key values as secret material.

#### 12.3.2 First-admin bootstrap binding

**REQ-04-087**
The stable deployment-configuration key for the current-profile first-admin bootstrap path MUST be `bootstrap.first_admin_manifest_path`. Omission means no bootstrap manifest path is configured. This key participates in the same deployment-configuration artifact, selector, and overlay rules defined by this section.
Profiles: base
Verified by: AC-343, AC-344, AC-345, AC-346

**REQ-04-088**
When present, `bootstrap.first_admin_manifest_path` MUST be an absolute POSIX path in the runtime where interpreted. Relative paths, `~`, shell-variable expansion forms, empty strings, NUL, and lexical `.` or `..` segments are invalid. Trailing-slash normalization MUST be deterministic. The referenced path MUST identify one deployment-local regular file rather than a writable root, every path component and the final object MUST be opened without following symlinks, and the opened regular file MUST contain no more than `1048576` bytes. The runtime MUST parse and hash immutable bytes returned by that single bounded read and MUST NOT validate one object and reopen a path that may identify another object.
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

For the current filesystem-backed backup-storage realization, backup artifacts and restore-verification extracts that carry incident data MUST be written through authenticated application-level envelope encryption before they reach `roots.backup_storage`. File permissions, root separation, or an operator assertion that the underlying volume is encrypted are not sufficient implementation proof for this realization. Implementations MUST fail closed when the recovery encryption key material required to read or write those encrypted envelopes is unavailable.

For SeaweedFS S3 object-store backup and restore, operator-private object-store
manifests and restore-verification artifacts are deployment-local recovery
artifacts. Any such artifact that carries incident-derived object identity,
storage refs, raw object keys, bucket names, or blob metadata MUST remain
outside incident portability bundles and public user-facing responses.
Shareable summaries MUST redact those values before retention outside the
deployment-local recovery boundary. The current profile defines no
MinIO-source migration ledger, copy ledger, validation extract, cutover, or
rollback artifact.
Profiles: base
Verified by: AC-403

#### 12.3.4 Enterprise-auth provider manifest binding

**REQ-04-115**
The stable deployment-configuration key for current-profile Enterprise Authentication claim activation MUST be `enterprise_authentication.claimed`. Omission defaults to `false`. A deployment with `enterprise_authentication.claimed=true` claims the Enterprise Authentication Extension Profile and MUST satisfy the provider-manifest, production verifier, and provider-reconciliation requirements in this section before readiness. A deployment with `enterprise_authentication.claimed=false` or omitted does not claim Enterprise Authentication and MUST expose the enterprise route families only through the unclaimed-extension error behavior owned by Core 01.

The stable deployment-configuration key for current-profile enterprise-auth provider definitions MUST be `enterprise_authentication.provider_manifest_path`. When `enterprise_authentication.claimed=true`, this key is required. When `enterprise_authentication.claimed=false` or omitted, this key is invalid. Omission is valid only for deployments that do not claim enterprise authentication. Explicit `null`, an empty string, a non-string value, a relative path, `~`, shell-variable expansion forms, NUL, and lexical `.` or `..` segments are invalid. Both enterprise-authentication keys participate in the same deployment-configuration artifact, selector, and `CARTULARY__` overlay rules defined by this section.
Profiles: enterprise_authentication
Verified by: AC-433, AC-436

**REQ-04-116**
When present and valid for the active profile claim, `enterprise_authentication.provider_manifest_path` MUST be an absolute POSIX path in the runtime where interpreted and MUST identify one deployment-local regular file no larger than `1048576` bytes. Every path component MUST be opened without following symbolic links, and validation MUST use the bytes and metadata from that one bounded open rather than reopening the configured path. The file MUST be UTF-8 JSON without relying on charset sniffing, environment-variable expansion, object storage, browser input, incident records, workbook rows, or manual database mutation as the provider-definition source. The manifest, all `secret_ref_v1` values it contains, and all referenced files MUST validate after deployment-configuration parsing and before any HTTP listener, WebSocket listener, or background-job runner starts. Raw secret values are forbidden in the manifest. Missing, empty, syntactically invalid, oversized, or unresolved `secret_ref_v1` values MUST fail closed before readiness. Diagnostics MAY include safe config paths, provider keys, and safe secret-reference names, but MUST NOT include raw secret values, transformed secret values, endpoint credentials, assertion material, signing-certificate contents, HMAC keys, provider tokens, host-absolute paths, or derived secret material.
Profiles: enterprise_authentication
Verified by: AC-433, AC-434, AC-436

**REQ-04-117**
An enterprise-auth provider manifest MUST be one JSON object with exactly these top-level members:

- `provider_manifest_schema_id`,
- `providers`.

Unknown top-level members, duplicate top-level members, non-object top-level JSON values, omitted required members, explicit `null` for required members, and duplicate members inside any manifest object are invalid. `provider_manifest_schema_id` MUST be exactly `cartulary.enterprise_auth_providers.v1`. `providers` MUST be an array with length `0..32`; omission, explicit `null`, non-array values, and more than `32` entries are invalid. The implementation MUST reject duplicate JSON object members rather than relying on last-wins parser behavior.
Profiles: enterprise_authentication
Verified by: AC-434

**REQ-04-118**
Every provider object MUST contain exactly the common provider fields plus the fields declared for its `provider_type`; unknown provider members and duplicate provider members are invalid. The common provider fields are:

| Field | Requirement |
| --- | --- |
| `provider_key` | Required non-null string matching `[a-z][a-z0-9_-]{0,63}`; unique within the manifest by exact token equality. |
| `provider_type` | Required non-null string exactly `oidc` or `saml`. |
| `display_name` | Required non-null `display_name_line_v1`. |
| `enabled` | Optional boolean; omission defaults to `true`; explicit `null` and non-boolean values are invalid. |

Provider keys, provider types, display names, and enabled values MUST NOT be trimmed, case-folded, Unicode-normalized, or otherwise repaired after JSON decoding except where the referenced string contract explicitly says so.
Profiles: enterprise_authentication
Verified by: AC-434, AC-435

**REQ-04-119**
For `provider_type='oidc'`, the provider object MUST contain exactly the common provider fields plus required `issuer`, required `authorization_endpoint`, required `token_endpoint`, required `jwks_uri`, required `client_id`, required `client_secret_ref`, and optional `additional_scopes`. `issuer`, `authorization_endpoint`, `token_endpoint`, and `jwks_uri` MUST each be an absolute HTTPS URL with no userinfo, query, or fragment. `client_id` MUST be a non-empty JSON string, MUST contain no C0 or C1 control code point, and MUST contain at most `256` Unicode scalar values. `client_secret_ref` MUST be a `secret_ref_v1` object; raw secret strings, inline secret objects outside `secret_ref_v1`, or alternate secret mechanisms are invalid. `additional_scopes` defaults to `[]` when omitted; explicit `null`, non-array values, non-string members, duplicate exact scope tokens, more than `16` entries, and the exact token `openid` are invalid. The server MUST add `openid` to OIDC authorization requests independently of `additional_scopes`. The OIDC redirect URI is derived from `application.public_origin` and `provider_key`; any manifest member that attempts to configure a redirect URI is invalid. The authoritative OIDC subject remains `sub`; email, display-name, username, groups, roles, or similar claims MUST NOT grant incident membership, deployment administration, or incident roles.
Profiles: enterprise_authentication
Verified by: AC-434, AC-435

**REQ-04-120**
For `provider_type='saml'`, the provider object MUST contain exactly the common provider fields plus required `idp_entity_id`, required `sso_url`, required `idp_signing_certificate_paths`, required `sp_entity_id`, and required `subject_source`. `sso_url` MUST be an absolute HTTPS URL with no userinfo, query, or fragment. `idp_signing_certificate_paths` MUST be an array with length `1..8`; every entry MUST be an absolute POSIX path, MUST satisfy the no-`~`, no-variable-expansion, no-NUL, and no lexical `.` or `..` segment rules from REQ-04-115, MUST reference one deployment-local regular file no larger than `262144` bytes, and MUST contain a parseable signing certificate. Every path component MUST be opened without following symbolic links, and certificate parsing MUST use the bytes and metadata from that one bounded open rather than reopening the path. `subject_source` MUST be exactly one of `{ "kind": "name_id" }` or `{ "kind": "attribute", "attribute_name": "<non-empty value>" }`; unknown subject-source members, duplicate subject-source members, omitted required subject-source members, explicit `null`, and empty `attribute_name` are invalid. The SAML ACS URL is derived from `application.public_origin` and `provider_key`; any manifest member that attempts to configure an ACS URL is invalid. Group-to-role maps and group-to-membership maps are invalid.
Profiles: enterprise_authentication
Verified by: AC-434, AC-435

**REQ-04-121**
After manifest and referenced-file validation succeeds and before readiness, startup MUST reconcile provider definitions against deployment-local persisted provider state in `provider_key` ascending order. For each manifest provider whose key is new, startup MUST create configured provider state. For each existing key whose persisted `provider_type` is unchanged, startup MUST update mutable provider configuration, including display name, protocol configuration, secret reference binding, referenced certificate set, and enabled state. For each existing key whose manifest `provider_type` differs from persisted `provider_type`, startup MUST fail closed with `invalid_deployment_config` rather than changing the provider type in place. For each persisted enterprise provider key omitted from the manifest, startup MUST retain the provider and mark it disabled. For a manifest provider with `enabled=false`, startup MUST retain the provider and its bindings but exclude it from interactive discovery and begin behavior. Reconciliation MUST be atomic with respect to provider-definition visibility: a failed reconciliation MUST NOT expose a partially updated provider set.
Profiles: enterprise_authentication
Verified by: AC-435, AC-436

**REQ-04-122**
Provider-definition configuration MUST NOT create, delete, rotate, or retire any enterprise-auth user binding. Provider removal from the manifest, manifest omission, `enabled=false`, failed reconciliation, or provider-definition restart MUST NOT delete users, incident memberships, local credentials, enterprise bindings, binding lineage, or deployment-local audit history. Changing provider definitions requires a validated deployment restart in the current profile. The current profile defines no authenticated browser route, public API request, workbook surface, incident-portability input, provider metadata upload, SAML certificate upload, OIDC metadata override, secret-management route, or runtime provider-policy mutation route that can create, edit, delete, rotate, or retire provider definitions or provider secrets.
Profiles: enterprise_authentication
Verified by: AC-435, AC-436

### 12.4 Filesystem-root path contract

**REQ-04-074**
For `binding_kind='filesystem_root'`, `path` MUST be an absolute POSIX path in the runtime where interpreted. Relative paths, `~`, shell-variable expansion forms, empty strings, NUL, and lexical `.` or `..` segments are invalid. Trailing-slash normalization MUST be deterministic. If a configured path already exists, canonicalization MUST resolve symlinks before enforcement. After canonicalization, configured filesystem roots MUST remain distinct from one another. A configured filesystem root that must be writable for its declared role but is not writable at startup is invalid.
Profiles: base, reference_pack
Verified by: AC-296, AC-297

**REQ-04-075**
Packaged read-only resources MAY resolve from install or package locations. Any operator-owned writable or persistent location MUST derive from the deployment configuration contract. Archive extraction, imports, uploads, preview generation, report builds, and temp-file writes MUST fail if the effective target escapes the configured filesystem root.

Startup validation of a configured root is necessary but is not operation-time containment. Every operator-owned read, write, create, rename, remove, or extraction effect MUST be performed through a root-anchored capability constructed from the validated binding. The capability MUST accept only canonical relative POSIX child references; MUST reject absolute paths, empty or dot components, `..`, NUL, backslashes, and normalization collisions; MUST NOT follow child or final-target symlinks; and MUST verify the actual opened object type. Exclusive creation and owner-defined atomic replacement MUST keep temporary and final objects within one root.

Containment MUST remain race-safe when a child path or configured root is removed, replaced, or made unwritable after startup. Such an operation MUST fail without partial publication and MUST perform owner-required cleanup. A `managed_service` binding MUST NOT instantiate a local filesystem fallback. Operation diagnostics, logs, telemetry, and retained evidence MUST NOT disclose absolute roots, hostile raw names, object keys, credentials, or secrets.

Invalid root syntax, profile compatibility, overlap, canonicalization, or startup writability remains `invalid_deployment_config`. A hostile operation child/member path uses the owning route or job error family. Root loss, replacement, or writability loss after readiness uses a safe owner storage or dependency failure and MUST NOT be reclassified as deployment-configuration invalidity.
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
- `extension_config_without_claim`,
- `managed_service_ref_invalid`,
- `managed_service_env_missing`,
- `managed_service_env_invalid`,
- `path_not_writable`,
- `bootstrap_manifest_path_missing`,
- `bootstrap_manifest_not_readable`,
- `bootstrap_manifest_not_regular_file`,
- `bootstrap_manifest_parse_error`,
- `bootstrap_manifest_schema_invalid`,
- `bootstrap_email_conflict`,
- `bootstrap_recovery_not_supported`,
- `bootstrap_persist_failed`,
- `provider_manifest_path_missing`,
- `provider_manifest_not_readable`,
- `provider_manifest_not_regular_file`,
- `provider_manifest_encoding_invalid`,
- `provider_manifest_parse_error`,
- `provider_manifest_schema_invalid`,
- `provider_manifest_secret_ref_invalid`,
- `provider_manifest_secret_missing`,
- `provider_manifest_referenced_file_invalid`,
- `provider_type_change_not_supported`,
- `provider_manifest_persist_failed`,
- `network_flow_cursor_key_missing`,
- `network_flow_cursor_key_invalid`,
- `network_flow_cursor_key_id_conflict`,
- `network_flow_cursor_key_secret_missing`,
- `network_flow_cursor_rotation_invalid`,
- `network_flow_safe_digest_key_missing`,
- `network_flow_safe_digest_key_invalid`,
- `network_flow_safe_digest_key_id_conflict`,
- `network_flow_safe_digest_key_secret_missing`,
- `network_flow_safe_digest_key_purpose_conflict`,
- `network_flow_safe_digest_rotation_invalid`,
- `network_flow_fixture_safe_digest_key_forbidden`,
- `network_flow_resource_limits_invalid`,
- `revisions_conflict_token_manifest_missing`,
- `revisions_conflict_token_manifest_invalid`,
- `revisions_conflict_token_key_missing`,
- `revisions_conflict_token_key_invalid`,
- `revisions_conflict_token_key_id_conflict`,
- `revisions_conflict_token_secret_missing`,
- `revisions_conflict_token_key_purpose_conflict`,
- `revisions_conflict_token_rotation_invalid`,
- `revisions_conflict_token_fixture_key_forbidden`,
- `value_below_minimum`,
- `value_above_maximum`,
- `invalid_telemetry_config`.

`invalid_telemetry_config` is valid only when an adopted OpenTelemetry NLSpec is active. It MUST be used only for syntactically invalid adopted `telemetry.*` keys, invalid explicit `null`, invalid enum values, invalid cross-key combinations, unsafe telemetry header declarations, invalid endpoint values, and forbidden environment-passthrough attempts. These failures still surface under top-level `error.code='invalid_deployment_config'`.
Profiles: base
Verified by: AC-294, AC-295, AC-296, AC-298, AC-320, AC-433, AC-434, AC-435, AC-526

**REQ-04-078**
Validation of the effective deployment configuration MUST complete before any HTTP listener, WebSocket listener, or background-job runner starts. Invalid deployment configuration MUST cause non-zero process exit. The implementation MUST NOT partially start and then discover deployment-configuration invalidity later during analyst workflow.

Application liveness and readiness MUST distinguish process health from active dependency readiness. `/healthz` is a process-liveness surface and MUST NOT become unhealthy solely because Postgres or the object store is unreachable after the process is running. `/readyz` is the public readiness surface and MUST emit a structured state with `status` equal to one of `starting_dependency_probe`, `ready`, `degraded_dependency`, or `recovering_dependency` when Postgres or object storage participates in the active deployment. `/readyz` MUST return HTTP `200` only when `status='ready'`; it MUST return HTTP `503` for startup dependency probing, post-ready dependency degradation, and recovery probing until the required probes pass. Readiness diagnostics MUST use redacted reason data and MUST NOT expose raw endpoint hosts, database names, bucket names, access keys, secret keys, object keys, storage refs, DSNs, or backend URLs.
Profiles: base
Verified by: AC-298, AC-433, AC-434, AC-435, AC-526

## 13. Coordinated Extensions Subsystem process contracts

This section is active through the atomic adoption of the Extensions NLSpec and all companion owners. Core 04 remains the sole configuration-container, authorization, application-process lifecycle, readiness, and exit-code owner.

For the adopted Extensions companion manifest, this owner document has `owner_document_schema_id='cartulary.core04.current.v1'` and `owner_document_version='extensions-adoption-1'`.

**REQ-04-143**
Every recognized profile uses the Boolean startup-only key `<profile_id>.claimed`, omitted as `false` and rejecting explicit `null` or non-Boolean values. Claim-key recognition MUST derive from the digest-validated current Extensions descriptor and configuration catalog; a separately maintained profile switch or path list is invalid. Each profile-local configuration row MUST declare `inactive_policy` and `inactive_value_schema_id`; the schema ID is non-null exactly for `syntax_only`. An unclaimed `syntax_only` value is checked only for the closed inert structural vocabulary in EXT-REQ-207. Core 04 applies no required/default omission policy, creates no configuration view, retains no accepted value, resolves no secret/file/trust/reference, performs no DNS or egress, and invokes no profile code. An omitted key is accepted without defaulting; an accepted explicit value is discarded before the next phase.

The effective `true` values form only a typed immutable requested-claim set. They MUST
NOT be treated as claimed, published, or substituted for the resolved claim set. The
Extensions admission boundary alone produces the resolved claim set after all admission
stages succeed, and telemetry identity derives only from that resolved set.

Every other unclaimed profile-local key fails before startup with top-level `error.code='invalid_deployment_config'`. Core 04 imports the Extensions diagnostic `extension_config_without_claim`, message `Extension configuration is present while the profile is inactive.`, and safe details `profile_id` and `config_path`. It translates the finding into one deployment-config `items[]` entry whose `path` is the dotted configuration key and whose `reason_code` and `message` are the imported Extensions values. Multiple findings use the ordinary deployment-config ordering by `path`, `reason_code`, and `message`. Profile-local replacement codes and compatibility aliases are forbidden.
Profiles: base
Verified by: EXT-AC-150

**REQ-04-144**
Core 04 consumes only the generated validation-condition registry. Schema validation surfaces require complete condition annotations and procedural surfaces require mutually exclusive exhaustive decision tables; an emitted unregistered condition fails closed without exposing a library error. Owner-result precedence is invocation failure, structural invalidity, `4097+` overflow reported as bounded `actual=4097`, remaining schema invalidity including `257..4096` against a 256-bound schema, valid nonempty findings, then valid empty success.

Every extension timeout uses checked monotonic arithmetic, saturates at signed 64-bit maximum instead of wrapping, selects the earlier local/inherited deadline, selects the inherited deadline on equality, and expires at `now >= deadline`. Proven commit wins over cancellation or timeout; with proven absent commit, cancellation sampled before expiry wins and an equal cancellation/deadline sample is timeout; indeterminate mutation is fatal. Zero cancellation grace starts termination immediately after cancellation and never admits a late normal result.
Profiles: base
Verified by: EXT-AC-143, EXT-AC-147, EXT-AC-152

**REQ-04-145**
Before extension claim resolution, the process MUST acquire one crash-released, deployment-global, exclusive application-process lease through exactly `unacquired -> acquiring -> held`; orderly shutdown may release from `held`. The lease permits exactly one active application process for the deployment even when infrastructure has provisioned multiple stopped, starting, or standby instances. Loss of ownership proof transitions immediately `held -> uncertain`, closes readiness and all admission, and permits `uncertain -> held` only when the original underlying session proves continuous ownership. Confirmed loss or detection-deadline expiry transitions to irreversible `lost`; in-process reacquisition is forbidden. Initial acquisition timeout mutates no profile state, starts no listener or dequeuer, emits `extension_application_process_active`, and exits `2`. Confirmed loss invokes fatal `application_process_lease_lost` and exits `70`. The storage mechanism is non-normative but must preserve session identity. This lease is distinct from the shared recovery serving lease in REQ-04-113; typed wrappers MAY share owner-neutral mechanics but MUST retain distinct lock identifiers, modes, owners, and failure vocabularies.

Stage 6 installs one immutable six-stage publication plan while admission remains closed and opens HTTP, WebSocket, job dequeue, readiness, discovery claim state, and workspace availability through one atomic gate only after every mandatory component acknowledges readiness.

The Jobs supervisor MUST validate its immutable dependencies and exact catalog-backed handler registration before the Stage 6 job-dequeue acknowledgment. Publication activation MUST run one synchronous initial recovery scan before HTTP or readiness can be externally observed. Failure of that initial scan is a startup publication failure: no listener may become externally available and the process exits under the existing startup-failure contract. After publication, transient recovery-scan errors remain inside the live supervisor with bounded safe diagnostics, while an unexpected supervisor panic or loop exit is loss of the required `job_dequeue` component.
Profiles: base
Verified by: EXT-AC-151, EXT-AC-158

**REQ-04-146**
Core 04's fatal lifecycle admits `published_component_lost` when a required publication listener, WebSocket gate, job-dequeue gate, job-dequeue supervisor, or worker terminates unexpectedly before bind, between bind and serving, or while serving. It also admits `application_process_lease_lost` and `recovery_serving_lease_lost` for confirmed loss of their distinct application-owned leases while starting or serving. A handled request, handler failure, retryable recovery-scan error, or lease-renewal failure isolated to one job is not component loss. Fatal detection closes readiness and admission, rejects new work, drains no longer than the configured shutdown deadline, preserves committed state and durable queued jobs, emits one secret-safe diagnostic, and exits `70`; it MUST NOT restart the component, reacquire a lost lease, rebuild the plan, or return to serving. Startup admission failures, including `extension_application_process_active` and `recovery_serving_lease_active`, start no listener and exit `2`. Lease loss, component loss, registry/state integrity contradictions, and indeterminate commits use the same idempotent fatal lifecycle; ordinary isolated operation failure does not. Recovery's exclusive target-lease loss remains governed by REQ-04-113 and the Recovery owner mapping rather than `recovery_serving_lease_lost`.
Profiles: base
Verified by: AC-534, EXT-AC-104, EXT-AC-151, EXT-AC-158

**REQ-04-154**
When `network_flow_activity` is claimed, application assembly MUST include one
private Network Flow projection-result cleanup dispatcher. It MUST start only
after the atomic serving-readiness transition and then perform one immediate
sweep. Later sweeps use a five-minute base cadence, never overlap, and coalesce
ticks received while a sweep is active. A successful sweep with `has_more=true`
schedules one continuation after five seconds. A transient sweep failure is
observable through the Network Flow telemetry boundary and schedules a bounded
retry no later than the next base cadence; it is not publication-component
loss. Unexpected dispatcher loop return or panic is
`published_component_lost` and enters REQ-04-146.

The dispatcher MUST expose no public route, common Job, operator command, or
Graph Projection worker. Shutdown stops new ticks and continuations, cancels
the active sweep, waits within the application shutdown deadline, and closes
the dispatcher before the dependencies it borrows. Startup failure before the
post-readiness loop is installed leaves readiness closed. An unclaimed profile
constructs and starts no Network Flow dispatcher.
Profiles: base, network_flow_activity
Verified by: AC-537
