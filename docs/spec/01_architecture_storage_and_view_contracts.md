# Cartulary Normative Core 01: Architecture, Storage, and View Contracts

## 1. Architecture pattern

**REQ-01-001**
Cartulary MUST use a **modular monolith** architecture for the base profile.
Profiles: base
Verified by: AC-231, AC-404

**REQ-01-002**
The base deployment topology MUST consist of:

- one web application deployable that contains the browser-facing UI, API surface, WebSocket hub, and background-job runners,
- one Postgres service as the authoritative structured data store,
- one S3-compatible object storage service as the authoritative binary evidence store.
Profiles: base
Verified by: AC-231, AC-404, AC-405

**REQ-01-003**
The application deployable MUST remain a single deployable unit even when deployed behind a reverse proxy or onto managed infrastructure.
Profiles: base
Verified by: AC-231, AC-404

Microservice decomposition is out of scope for current conformance.

## 2. Required modules and boundaries

**REQ-01-004**
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
Profiles: base, snapshot_reporting
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-231, AC-233

**REQ-01-005**
These are logical module boundaries. They MUST be independently testable, but they MUST NOT require separate deployables.
Profiles: base
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-231

**REQ-01-006**
File-based structured import beyond clipboard paste MUST be implemented as a dedicated internal `imports` module within the modular monolith.
Profiles: base, import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-231, AC-232

**REQ-01-007**
The `imports` module MUST own, at minimum:

- file-based source adapters for CSV and XLSX input,
- workbook inspection, candidate-region selection, preview, and header mapping,
- import job execution with progress, cancellation, retry-safe status, and diagnostics,
- deterministic import provenance capture,
- compatibility shims for spreadsheet-specific parser behavior and workbook-shape heuristics.
Profiles: base, import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-231, AC-232

**REQ-01-008**
Clipboard interaction remains part of the base workbook surface. When clipboard paste feeds structured ingest, it MUST use the same stable tabular-ingest contract and shared mapping engine as file-based import paths.
Profiles: base, import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-231, AC-232

For base-profile clipboard paste, the tabular-ingest operation is ephemeral: it does not create an `import_session` or durable `import_unit`, but it MUST still derive the row plan from the active `view_schema_id`, stable `field_key` columns, source-column ordinals, and declared `entity_binding_mode`. Timeline `mention_origin` paste, Hosts `entity_origin` paste, and Identities `entity_origin` paste MUST use that same planning contract rather than surface-local parser or header heuristics. Default interactive Ctrl+V dispatch is intentionally narrower than the ingest parser: single-line comma-only `text/plain` MUST remain scalar clipboard text unless a future explicit paste-as-table command or API request declares tabular CSV intent.

**REQ-01-009**
The workbook, timeline, entities, evidence, revisions, projections, and reporting concerns MUST depend only on the stable tabular-ingest contract and shared mapping engine for structured ingest. They MUST NOT depend directly on XLSX or OpenXML parsers, workbook-specific heuristics, or Excel-specific semantics.
Profiles: base, import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-231, AC-232

### 2.1 Phase 2 Workbook Import Assistant

The **Import Extension Profile** MAY expose a **Phase 2 Workbook Import Assistant** for structured file onboarding of CSV and XLSX sources.

**REQ-01-010**
The Phase 2 Workbook Import Assistant MUST remain an internal concern of the dedicated `imports` module. It MUST NOT add workbook-specific runtime semantics to Timeline, Hosts, Identities, Evidence, Notes, projections, snapshot generation, or write-back.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-01-011**
The assistant MUST expose file-based import work as:

- one `import_session` for an uploaded source file and one operator-driven workflow,
- one or more explicit `import_unit` objects discovered from that source,
- one `mapping_fingerprint` per selected unit for the operator-approved header-to-field plan,
- zero or more closed-vocabulary `warning_code[]` values for downgraded workbook features.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-01-012**
Whole-workbook import MUST mean an orchestrated batch of explicit `import_unit` objects selected from one `import_session`. It MUST NOT preserve workbook object identity or require runtime workbook semantics outside `imports`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-01-013**
The current profile MUST limit `import_unit.locator_kind` to `csv_file`, `xlsx_used_range`, `xlsx_table`, `xlsx_named_range`, and `xlsx_region`. Workbook inspection, used-range discovery, table discovery, named-range eligibility checks, operator-selected region previewing, downgrade warnings, and spreadsheet-parser compatibility shims MUST remain inside `imports`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-01-014**
The semantic identity of an `import_unit` MUST be the tuple `source_content_sha256 + canonical_locator + parser_version`. Modules outside `imports` MUST consume only the stable tabular-ingest contract, shared mapping engine, deterministic provenance, `mapping_fingerprint`, and declared `warning_code[]` values. They MUST NOT link directly against XLSX or OpenXML parsing libraries, workbook-shape heuristics, or workbook-behavior semantics.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-046, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

## 3. Client and server responsibilities

### 3.1 Browser client

**REQ-01-015**
The browser client MUST provide:

- a virtualized workbook grid,
- keyboard navigation,
- paste handling,
- a detail and relationship inspector,
- evidence preview behavior,
- save/conflict state presentation,
- a local pending queue for transient network interruptions,
- real-time presence and live row updates.
Profiles: base
Verified by: AC-001, AC-003, AC-004, AC-005, AC-043, AC-044, AC-045, AC-047, AC-231

**REQ-01-016**
The virtualized workbook grid MUST render only the visible viewport plus a bounded overscan region. It MUST NOT require full-DOM rendering of every row in a 10k+ row incident.
Profiles: base
Verified by: AC-001, AC-003, AC-004, AC-005, AC-043, AC-044, AC-045, AC-047, AC-231

**REQ-01-017**
Selection, focus, and pending-edit anchoring in the grid MUST remain bound to the selected `record_id` during live updates, sorting, filtering, and grouping.
Profiles: base
Verified by: AC-001, AC-003, AC-004, AC-005, AC-043, AC-044, AC-045, AC-047, AC-231

### 3.2 Application server

**REQ-01-018**
The application server MUST provide:

- authenticated HTTP+JSON API endpoints,
- authoritative mutation validation,
- optimistic concurrency enforcement,
- projection maintenance,
- WebSocket-based collaboration updates,
- background-job orchestration,
- reference-pack activation verification,
- snapshot and report generation when the Snapshot and Reporting Extension Profile is implemented.
Profiles: base, snapshot_reporting, reference_pack
Verified by: AC-046, AC-129, AC-231, AC-233, AC-234

### 3.3 Public HTTP and WebSocket interface contract

**REQ-01-019**
The base-profile client/server boundary MUST be a versioned HTTP+JSON surface plus a bounded WebSocket event stream. Internal modules MAY use any internal call form, but conformance at the browser-facing boundary MUST be evaluated against this public surface.
Profiles: base
Verified by: AC-124, AC-125, AC-126, AC-127, AC-128, AC-129, AC-131, AC-135, AC-231

#### 3.3.1 Versioning and compatibility

**REQ-01-020**
The HTTP surface MUST be rooted at `/api/v1/`. The WebSocket surface MUST be rooted at `/ws/v1/`. Within `/ws/v1/`, public path identity is part of the v1 contract. The base profile MUST NOT substitute, alias, or otherwise treat another v1 public path as equivalent to the incident-scoped subscription route unless a later requirement explicitly enumerates that additional path.
Profiles: base
Verified by: AC-124, AC-125, AC-127, AC-131, AC-135, AC-231

**REQ-01-021**
Breaking changes to route patterns, required request fields, required response fields, envelope shapes, or event semantics MUST use a new major version root. Additive response fields and additive route families for claimed extension profiles MAY be introduced within the same major version. Additive optional top-level request fields within the same major version are valid only when they are explicitly declared by the owning route contract. Top-level request namespaces for mutating routes are closed by default. Unknown top-level request members MUST be rejected unless the owning route contract explicitly declares an extension container or an ignore rule.
Profiles: base
Verified by: AC-124, AC-125, AC-127, AC-131, AC-135, AC-219, AC-220, AC-231

**REQ-01-022**
All public requests and responses MUST address writable surfaces by stable identifiers. The client MUST identify the incident by `incident_id`, the active view by `view_schema_id`, record-scoped target rows by `record_id`, mention-scoped targets by `entity_mention_id`, record-scoped optimistic writes by `base_row_version`, mention-scoped optimistic writes by `base_mention_row_version`, writable cells by `field_key`, and multi-change user actions by `client_txn_id`. The public surface MUST NOT require clients to address mutations by visible row order, tab label, column label, projection-table name, or storage-table name.
Profiles: base
Verified by: AC-124, AC-125, AC-127, AC-131, AC-135, AC-231

#### 3.3.2 Session and authentication routes

**REQ-01-023**
The public authentication contract MUST be session-based. The authoritative browser-session credential MUST be a server-managed opaque session token carried in an `HttpOnly` `Secure` cookie with `Path=/` and `SameSite=Lax` or a stricter same-site policy. If bearer authentication is enabled for non-browser clients or trusted automation, the implementation MAY additionally accept `Authorization: Bearer <opaque_session_token>` from the same opaque session family. The public token format MUST remain opaque to clients. Conformance MUST NOT require a browser or API client to parse JWT claims or provider-specific assertion contents to determine actor identity, incident scope, or expiry.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-01-024**
The base route family MUST include:

- `POST /api/v1/auth/login`,
- `POST /api/v1/auth/logout`,
- `GET /api/v1/auth/session`,
- `GET /api/v1/auth/credential-state`,
- `POST /api/v1/auth/password/change`,
- `POST /api/v1/auth/mfa/totp/begin`,
- `POST /api/v1/auth/mfa/totp/complete`.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231, AC-334, AC-335, AC-336, AC-337, AC-338, AC-339

Contract tables. The tables in §3.3.2 through §3.3.2.2 are the compact owner-local contract for request shape, omission and default behavior, replay, success transport, and family-specific errors. When a table cell and surrounding prose describe the same boundary fact, the table is the quick-reference statement and the surrounding prose supplies algorithmic, lifecycle, and example detail that is not reduced to cells.

**Table 3.3.2-A. Auth route index**

| Route | Auth context | Request contract summary | Omission and default summary | Replay and idempotency | Success response or effect | Primary error codes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/auth/login` | Anonymous local-account login | `username`, `password`, optional `second_factor` | Omitted `second_factor` means primary-credentials-only attempt; `client_txn_id` and provider-protocol fields are forbidden | Intentionally non-idempotent; transport retry MAY mint a fresh session | Establishes the server-managed session and returns the same session resource exposed by `GET /api/v1/auth/session` | `invalid_auth_request`, `mfa_required`, `mfa_setup_required`, `invalid_credentials`, `invalid_second_factor` |
| `POST /api/v1/auth/logout` | Current authenticated session | No route-specific request members are declared in the current profile | No create-time defaults apply | Revokes only the current session; later use of that session fails through ordinary auth | Revokes the current session immediately and emits `session_revoked` to any accepted WebSocket on that session | Ordinary auth failures |
| `GET /api/v1/auth/session` | Current authenticated session | Singleton read; no body members | Pagination members are rejected with `invalid_pagination_request` and `reason_code=pagination_not_supported` | Read route | Returns one session resource | Ordinary auth failures; `invalid_pagination_request` |
| `GET /api/v1/auth/credential-state` | Current authenticated session | Singleton read; no body members | Pagination members are rejected; `bootstrap_token` is not allowed on this route | Read route | Returns one safe credential-state resource | Ordinary auth failures; `credential_bootstrap_rejected`; `invalid_pagination_request` |
| `POST /api/v1/auth/password/change` | Current authenticated session for the addressed current user | Required `client_txn_id`, `current_password`, `new_password`; optional `second_factor` | `second_factor` is optional only when no active TOTP credential exists; omitted or empty `reason` is not part of this route | Route-scoped idempotency within `(actor_user_id, client_txn_id)` | Updates password state, stamps `password.changed_at`, revokes all active sessions for the user, and returns safe success data including `sessions_revoked=true` | `invalid_current_password`, `invalid_second_factor`, `client_txn_conflict`, ordinary malformed-request failures |
| `POST /api/v1/auth/mfa/totp/begin` | Exactly one of current authenticated session or valid `bootstrap_token` | Required `client_txn_id`; replacement under current-session auth also requires `current_password` and `second_factor` when one active factor already exists | No factor-less replacement path exists when one active TOTP credential is present | Idempotent within the same auth scope and `client_txn_id`; replay returns the original pending enrollment and seed material while pending | Returns `enrollment_id`, `expires_at`, and `totp_setup` with the seed material and fixed TOTP parameters | `credential_bootstrap_rejected`, `invalid_second_factor`, `client_txn_conflict`, ordinary malformed-request failures |
| `POST /api/v1/auth/mfa/totp/complete` | Exactly one of current authenticated session or valid `bootstrap_token`, matching the begin route auth mode | Required `client_txn_id`, `enrollment_id`, and `code` | First-time bootstrap completion never auto-issues a session | Route-scoped idempotency uses the same auth scope discipline as begin; a stale or different replay fails rather than creating a second activation | Activates the pending TOTP secret, clears pending setup, consumes any bootstrap token used for the flow, and revokes all sessions only when replacing an existing factor | `totp_setup_not_pending`, `credential_bootstrap_rejected`, `client_txn_conflict`, ordinary malformed-request failures |

**Table 3.3.2-B. Login request members**

| Member | Type or contract | Requiredness | Allowed values | Omission and explicit-`null` behavior | Normalization and validation | Replay participation |
| --- | --- | --- | --- | --- | --- | --- |
| `username` | `string_contract_id=email_address_v1` | Required | Non-null local-account email address | Omission is invalid; explicit `null` is invalid | Normalized with the same deterministic substrate used for local-user lookup and membership-by-email resolution | Login is intentionally non-idempotent |
| `password` | JSON string | Required | Non-null, non-empty string | Omission is invalid; explicit `null` is invalid | Compared exactly after JSON decoding; the server MUST NOT trim, case-fold, or Unicode-normalize it | Login is intentionally non-idempotent |
| `second_factor` | Object | Optional | Present only when attempting MFA satisfaction on this route | Omission means primary-credentials-only attempt; explicit `null` is invalid | Unknown members are invalid | Login is intentionally non-idempotent |
| `second_factor.kind` | String | Required when `second_factor` is present | Exactly `totp` in the base profile | Omission is invalid when `second_factor` is present; explicit `null` is invalid | Closed vocabulary; any other token fails with `invalid_auth_request` | Login is intentionally non-idempotent |
| `second_factor.assertion` | Object | Required when `second_factor` is present | Exact TOTP assertion object | Omission is invalid when `second_factor` is present; explicit `null` is invalid | Unknown members are invalid | Login is intentionally non-idempotent |
| `second_factor.assertion.code` | String | Required when `kind='totp'` | Exactly six ASCII decimal digits | Omission is invalid; explicit `null` is invalid | No spaces or separators are allowed | Login is intentionally non-idempotent |

**Table 3.3.2-C. Login outcome matrix**

| Outcome | Transport | Required condition | Required response detail |
| --- | --- | --- | --- |
| Success | Ordinary successful auth response | Primary credentials valid; if MFA is required, a valid `second_factor` is present | Server-managed session created; same session resource as `GET /api/v1/auth/session` |
| `invalid_auth_request` | `400` | Request shape, types, closed vocabularies, or forbidden fields are invalid | `error.details.field` is present when one member is attributable |
| `mfa_required` | `401` | Primary credentials are valid, one active TOTP credential exists, and `second_factor` is omitted | `error.details.required_second_factor_kinds=["totp"]` |
| `mfa_setup_required` | `401` | Primary credentials are valid and no active TOTP credential is enrolled | No session is created; `bootstrap_token` and `bootstrap_expires_at` are returned; `required_setup_kinds=["totp"]` |
| `invalid_credentials` | `401` | The server is unwilling to acknowledge valid primary credentials | No session and no pre-authenticated state |
| `invalid_second_factor` | `401` | Primary credentials are valid and a structurally valid TOTP assertion is wrong or expired | No session and no partial session state |


**REQ-01-025**
`POST /api/v1/auth/login` MUST be the base-profile local-account login route. The request body MUST be a JSON object and MUST accept:

- required `username`,
- required `password`,
- optional `second_factor`.

`username` remains the v1 wire member name for this route. In the base profile, for the local-account login route, it is the user's email address. `username` MUST be non-null and MUST satisfy `string_contract_id=email_address_v1`. A supplied `username` value that normalizes to authoritative `null` under `email_address_v1`, or otherwise fails that contract, MUST fail with `400` and `error.code = invalid_auth_request` rather than `401` and `error.code = invalid_credentials`. Local-account lookup MUST use the same deterministic normalization and comparison substrate as the user-account contract and membership-by-email resolution. The base profile MUST NOT create, require, or infer a second persisted local-login identifier distinct from the authoritative user `email`. `password` MUST be a non-null non-empty string and MUST be compared exactly as supplied after JSON decoding. The server MUST NOT trim, case-fold, or Unicode-normalize `password`. For local-password verification, the server MUST encode the exact JSON-decoded `password` as UTF-8 without BOM before Argon2id verification.

When `second_factor` is omitted, the request is a primary-credentials-only login attempt. When present, `second_factor` MUST be an object and MUST be non-null. `second_factor.kind` MUST be present and, in the base profile, MUST use the closed vocabulary `totp`. `second_factor.assertion` MUST be present, MUST be an object, and MUST be non-null. For `kind='totp'`, `second_factor.assertion` MUST use exactly this shape: `{ "code": "123456" }`. `code` MUST be a string of exactly six ASCII decimal digits with no spaces or separators.

Unknown top-level request members, unknown `second_factor` members, unknown `assertion` members for the selected `kind`, a missing required member, a type mismatch, a supplied `null` where this requirement forbids `null`, a `second_factor.kind` outside the base-profile vocabulary, or any `client_txn_id`, `id_token`, `authorization_code`, `saml_response`, `provider_assertion`, or WebAuthn ceremony field sent on this route MUST fail with `400` and `error.code = invalid_auth_request`. When the failure is attributable to one request member, `error.details.field` MUST identify that member.

If the local account does not require MFA, omission of `second_factor` MUST be accepted and a structurally valid `second_factor` MUST NOT by itself prevent login. If the local account requires MFA, has one active TOTP credential, and `second_factor` is omitted, the server MUST fail with `401` and `error.code = mfa_required`; `error.details.required_second_factor_kinds` MUST equal `["totp"]` in the base profile. If the local account requires MFA, primary credentials are valid, and no active TOTP credential is currently enrolled, the server MUST fail with `401` and `error.code = mfa_setup_required`; it MUST set no session cookie; `error.details.required_setup_kinds` MUST equal `["totp"]`; and the error payload MUST also include one opaque short-lived `bootstrap_token` plus `bootstrap_expires_at`. If the server is not willing to acknowledge that primary credentials were valid, including for unknown `username`, wrong `password`, inactive local account, or equivalent pre-MFA failure, it MUST fail with `401` and `error.code = invalid_credentials`. If primary credentials are valid and a structurally valid TOTP assertion is present but wrong or expired, the server MUST fail with `401` and `error.code = invalid_second_factor`.

This route MUST NOT require or interpret `client_txn_id`. On success it MUST establish the server-managed session and return the same session resource exposed by `GET /api/v1/auth/session`. On any non-success outcome, the server MUST create no session, set no session cookie, and expose no partial or pre-authenticated session state. Transport retries after an uncertain network boundary are not idempotent and MAY create a fresh session if the client repeats the request.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231, AC-244, AC-245, AC-246, AC-247, AC-248, AC-249, AC-250, AC-311, AC-334, AC-335, AC-336, AC-337, AC-338, AC-339

Example request body:

```json
{
  "username": "analyst1@corp.example",
  "password": "correct horse battery staple",
  "second_factor": {
    "kind": "totp",
    "assertion": {
      "code": "123456"
    }
  }
}
```

##### 3.3.2.1 Session resource and expiry contract

**REQ-01-026**
The session routes MUST expose the lifecycle boundaries defined by Core 04 §1.1.1 without requiring client-side token parsing.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**Table 3.3.2.1-A. Session resource members**

| Field | Presence | Nullability and ordering | Notes |
| --- | --- | --- | --- |
| `user_id` | Required | Non-null | Stable internal user identity |
| `display_name` | Required | Non-null | Safe user-facing display name |
| `provider_type` | Required | Non-null | Closed vocabulary `local`, `oidc`, `saml` |
| `mfa_state` | Required | Non-null | Closed vocabulary `not_required`, `satisfied` |
| `is_deployment_admin` | Required | Non-null | Deployment-scoped capability summary only |
| `authenticated_at` | Required | Non-null | Session-authentication timestamp |
| `idle_expires_at` | Required | Non-null | Inspection-only route; reading it does not extend it |
| `absolute_expires_at` | Required | Non-null | Absolute session boundary |
| `session_expires_at` | Required | Non-null | Earlier of `idle_expires_at` and `absolute_expires_at` |
| `memberships[]` | Required | May be empty; ordered by `incident_id asc` | Informational bootstrap state only |
| `memberships[].incident_id` | Required when `memberships[]` item exists | Non-null | Stable incident identity |
| `memberships[].role` | Required when `memberships[]` item exists | Non-null | Closed vocabulary `viewer`, `editor`, `reviewer`, `admin` |


**REQ-01-027**
`GET /api/v1/auth/session` MUST return the common success envelope with `data`
equal to one session resource.

The session resource MUST expose, at minimum:

- `user_id`
- `display_name`
- `provider_type`
- `mfa_state`
- `is_deployment_admin`
- `authenticated_at`
- `idle_expires_at`
- `absolute_expires_at`
- `session_expires_at`
- `memberships[]`

`provider_type` MUST use the closed vocabulary `local`, `oidc`, `saml`.

`mfa_state` MUST use the closed vocabulary `not_required`, `satisfied`.

`session_expires_at` MUST be the earlier of `idle_expires_at` and
`absolute_expires_at`.

`memberships[]` MUST always be present. It MAY be empty. It MUST be ordered by
`incident_id asc`. Each `memberships[]` item MUST contain:

- `incident_id`
- `role`

`role` MUST use the closed incident-role vocabulary `viewer`, `editor`,
`reviewer`, `admin`.

`memberships[]` is informational bootstrap state only. Incident-scoped
authorization remains authoritative on incident-scoped routes, jobs, preview or
download handle issuance and redemption, and WebSocket subscriptions.

The base profile MUST NOT replace `memberships[]` with a current-incident-only
object or any other alternate response shape on this route.

Because this route is singleton, it MUST reject `limit`, `cursor_token`, and
pagination aliases with `400`, `error.code=invalid_pagination_request`, and
`error.details.reason_code=pagination_not_supported`.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-01-028**
`GET /api/v1/auth/session` is an inspection route and MUST NOT by itself extend `idle_expires_at`.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-01-029**
`POST /api/v1/auth/logout` MUST revoke the current session immediately. If that session currently owns one or more accepted WebSocket connections, the server MUST send `session_revoked` on those connections and close them.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-01-030**
When the current session has expired, the auth route family MUST fail closed and a caller can establish a new session only by completing the login flow again with any applicable MFA requirement.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-01-031**
If the Enterprise Authentication Extension Profile is implemented, provider-backed sign-in MUST terminate into this same session contract so the remaining API routes and WebSocket stream remain provider-agnostic. The public enterprise-auth route family is owned by §20.
Profiles: base, enterprise_authentication
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231, AC-235, AC-250, AC-290, AC-291

##### 3.3.2.2 Credential lifecycle and TOTP bootstrap routes

Contract tables. The tables in §3.3.2.2 compact the credential-state, password-change, and TOTP bootstrap surfaces without restating Core 04 session-lifecycle or secret-storage rules.

**Table 3.3.2.2-A. Bootstrap-token contract**

| Property | Requirement |
| --- | --- |
| Token family | `bootstrap_token` is a credential-setup token, not a session |
| Accepted routes | Only `POST /api/v1/auth/mfa/totp/begin` and `POST /api/v1/auth/mfa/totp/complete` |
| Rejected routes | `GET /api/v1/auth/session`, `GET /api/v1/auth/credential-state`, ordinary incident or record routes, and `/ws/v1/*` |
| Lifetime | Expires 10 minutes after issuance unless a later profile says otherwise |
| Single-use and supersession | Consumed by successful `totp/complete`; later bootstrap issuance for the same user, administrator password reset, administrator TOTP reset, or expiry invalidates earlier tokens |
| Rejection path | Invalid, expired, consumed, superseded, wrong-subject, or wrong-route use fails with `409`, `error.code = credential_bootstrap_rejected`, and a family `reason_code` from §3.3.6.2 |

**Table 3.3.2.2-B. Credential-state resource**

| Field | Presence | Nullability and defaults | Notes |
| --- | --- | --- | --- |
| `user_id` | Required | Non-null | Current authenticated user only |
| `auth_kind` | Required | Non-null | `local` for the base profile |
| `mfa_required` | Required | Non-null | Safe credential-policy summary |
| `recovery_model` | Required | Non-null | `admin_assisted` for the base profile |
| `password.changed_at` | Required | May be `null` | Safe timestamp only |
| `totp.state` | Required | Non-null | Closed vocabulary `not_enrolled`, `pending`, `active` |
| `totp.enrolled_at` | Required | May be `null` | Present even when null |
| `totp.pending_expires_at` | Required | May be `null` | Present even when null |
| Secret-bearing fields | Forbidden | N/A | `secret_base32`, `otpauth_uri`, password hashes, raw bootstrap tokens, TOTP secret material, and provider assertions MUST NOT appear |

**Table 3.3.2.2-C. Password-change route contract**

| Member or rule | Requirement |
| --- | --- |
| Required members | `client_txn_id`, `current_password`, `new_password` |
| Optional members | `second_factor` |
| Default and omission rules | `second_factor` is optional only when the current account has no active TOTP credential; omitted or empty `reason` is not part of this route |
| Validation | `new_password` binds to `local_password_provision_v1`; `current_password` uses the same exact verification substrate as local login |
| Idempotency | Route-scoped idempotency key is `(actor_user_id, client_txn_id)` |
| Success effect | Updates `password_hash`, stamps `password.changed_at`, revokes all active sessions for the user including the current one, and returns safe success data with `sessions_revoked=true` |
| Primary failures | `invalid_current_password`, `invalid_second_factor`, `client_txn_conflict`, plus ordinary malformed-request failures |

**Table 3.3.2.2-D. TOTP begin and complete contract**

| Route | Required members | Additional required conditions | Replay and idempotency | Success summary | Primary failures |
| --- | --- | --- | --- | --- | --- |
| `POST /api/v1/auth/mfa/totp/begin` | `client_txn_id` | Exactly one auth mode: current session or valid `bootstrap_token`; current-session replacement requires `current_password` and `second_factor` when one active TOTP credential exists | Replay within the same auth scope and `client_txn_id` returns the original pending enrollment and seed while pending | Returns `enrollment_id`, `expires_at`, and `totp_setup` with `secret_base32`, `otpauth_uri`, `algorithm='SHA1'`, `digits=6`, and `period_seconds=30` | `credential_bootstrap_rejected`, `invalid_second_factor`, `client_txn_conflict`, ordinary malformed-request failures |
| `POST /api/v1/auth/mfa/totp/complete` | `client_txn_id`, `enrollment_id`, `code` | Exactly one auth mode and it MUST match the begin route auth mode; `code` is exactly six ASCII decimal digits | Same auth-scope discipline as begin; stale or different replay does not create a second activation | Activates the pending TOTP secret, clears pending setup, consumes any bootstrap token used for the flow, and revokes all active sessions only when replacing an existing factor; first-time bootstrap completion never auto-issues a session | `totp_setup_not_pending`, `credential_bootstrap_rejected`, `client_txn_conflict`, ordinary malformed-request failures |

**Table 3.3.2.2-E. Auth-family error and reason summary**

| Condition | Transport | `error.code` | Required details |
| --- | --- | --- | --- |
| Malformed login request or forbidden login field | `400` | `invalid_auth_request` | `error.details.field` when one member is attributable |
| Primary credentials valid and active factor exists but no `second_factor` supplied | `401` | `mfa_required` | `required_second_factor_kinds=["totp"]` |
| Primary credentials valid and no active factor exists | `401` | `mfa_setup_required` | `required_setup_kinds=["totp"]`, `bootstrap_token`, and `bootstrap_expires_at` |
| Primary credentials not accepted | `401` | `invalid_credentials` | No session and no pre-authenticated state |
| Structurally valid TOTP assertion is wrong or expired | `401` | `invalid_second_factor` | No session and no partial session state |
| Wrong-route, expired, consumed, superseded, or wrong-subject bootstrap token use | `409` | `credential_bootstrap_rejected` | Family `reason_code` from §3.3.6.2 |
| Password-change current password mismatch | `409` | `invalid_current_password` | Route-local password-change failure |
| TOTP completion targets no pending enrollment or one that is expired or consumed | `409` | `totp_setup_not_pending` | Family `reason_code` from §3.3.6.2 |


**REQ-01-522**
The base profile MUST expose a bounded public credential-lifecycle contract for local accounts through the routes listed in REQ-01-024. `bootstrap_token` is a credential-setup token, not a session. It MUST be opaque, single-subject, short-lived, and accepted only by `POST /api/v1/auth/mfa/totp/begin` and `POST /api/v1/auth/mfa/totp/complete`. It MUST NOT be accepted by `GET /api/v1/auth/session`, `GET /api/v1/auth/credential-state`, ordinary incident or record routes, or `/ws/v1/*`. Unless a later profile says otherwise, `bootstrap_token` expires 10 minutes after issuance, is single-use, is consumed by successful `totp/complete`, and becomes invalid on later bootstrap issuance for the same user, administrator password reset, administrator TOTP reset, or expiry. A bootstrap token that is expired, consumed, superseded, bound to a different subject, or used on a route outside its allowed family MUST fail with `409` and `error.code = credential_bootstrap_rejected`; `error.details.reason_code` MUST use the registry in §3.3.6.2.
Profiles: base
Verified by: AC-334, AC-335, AC-336, AC-337, AC-338, AC-339

**REQ-01-523**
`GET /api/v1/auth/credential-state` MUST be session-authenticated and current-user scoped. It MUST return only safe credential state: `user_id`, `auth_kind`, `mfa_required`, `recovery_model`, nullable `password.changed_at`, and `totp` with `state`, nullable `enrolled_at`, and nullable `pending_expires_at`. For local accounts, `auth_kind` MUST be `local`, `recovery_model` MUST be `admin_assisted`, and `totp.state` MUST use the closed vocabulary `not_enrolled`, `pending`, and `active`. The response MUST NOT expose `secret_base32`, `otpauth_uri`, raw bootstrap tokens, TOTP secret material, password hashes, or provider assertions. A bootstrap token presented on this route MUST fail with `409`, `error.code = credential_bootstrap_rejected`, and `error.details.reason_code = not_allowed_for_route`.
Profiles: base
Verified by: AC-335, AC-339

**REQ-01-524**
`POST /api/v1/auth/password/change` MUST be session-authenticated and current-user scoped. The request body MUST be a JSON object and MUST accept required `client_txn_id`, required `current_password`, required `new_password`, and optional `second_factor`. `new_password` is bound to `string_contract_id=local_password_provision_v1`. `current_password` MUST be a non-null JSON string and MUST be verified using the same exact post-JSON-decoding code-point equality and UTF-8-without-BOM substrate used for local login-password verification. If the current account has one active TOTP credential, the request MUST also include `second_factor` using the same `kind='totp'` assertion shape as `POST /api/v1/auth/login`; omission or a structurally valid but wrong or expired TOTP code MUST fail with `401` and `error.code = invalid_second_factor`. A structurally valid `current_password` that does not match the current stored local password MUST fail with `409` and `error.code = invalid_current_password`. On success the server MUST update `password_hash`, stamp `password.changed_at`, revoke all active sessions for that user including the current session, and return the common success envelope with at least `user_id`, `password.changed_at`, and `sessions_revoked=true`. The route MUST use `client_txn_id` as a route-scoped idempotency key within `(actor_user_id, client_txn_id)`. Normalized replay comparison for this route MUST include exact `current_password`, exact `new_password`, and the normalized `second_factor` payload when supplied. Exact replay of a previously committed success MUST return `200 OK` with the original committed success payload before any fresh state evaluation runs. Reuse of the same route-scoped key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. Any deployment-local idempotency substrate for this route MUST NOT retain cleartext `current_password` or `new_password`.
Profiles: base
Verified by: AC-338, AC-339

**REQ-01-525**
`POST /api/v1/auth/mfa/totp/begin` MUST accept exactly one auth mode: either an authenticated current session or one valid `bootstrap_token`. The request body MUST be a JSON object and MUST accept required `client_txn_id`. When called with current-session auth and one existing active TOTP credential, it MUST also require `current_password` plus `second_factor` using the same TOTP assertion shape as the local-login route before issuing a replacement seed. On success it MUST return `enrollment_id`, `expires_at`, and a `totp_setup` object containing `secret_base32`, `otpauth_uri`, `algorithm='SHA1'`, `digits=6`, and `period_seconds=30`. `secret_base32` and `otpauth_uri` MUST appear only on this begin response and MUST NOT appear on later read routes, history payloads, WebSocket payloads, or safe user resources. For idempotency within the same auth scope and `client_txn_id`, replay of the same normalized request before enrollment expiry MUST return the original pending enrollment and the same seed material.
Profiles: base
Verified by: AC-336, AC-339

**REQ-01-526**
`POST /api/v1/auth/mfa/totp/complete` MUST accept exactly one auth mode: either an authenticated current session or one valid `bootstrap_token`, and that auth mode MUST match the auth mode used for the referenced pending enrollment. The request body MUST be a JSON object and MUST accept required `client_txn_id`, required `enrollment_id`, and required `code`, where `code` is a string of exactly six ASCII decimal digits. If no pending enrollment exists for `enrollment_id`, or the referenced pending enrollment is expired or already consumed, the server MUST fail with `409`, `error.code = totp_setup_not_pending`, and `error.details.reason_code` from §3.3.6.2. On success the server MUST atomically activate the pending TOTP secret, clear pending setup state, consume any bootstrap token used for the flow, and revoke all active sessions only when the call replaces an existing active factor. First-time bootstrap completion MUST NOT auto-issue a session; the user MUST complete ordinary login after setup. The base profile does not include email links, SMS, backup codes, provider-mediated recovery, self-service factor disablement, or forced-password-change-on-next-login state on these routes.
Profiles: base
Verified by: AC-337, AC-339

#### 3.3.3 Route families

**REQ-01-032**
The base-profile route set MUST include stable route families for:

- authentication and bounded credential-lifecycle routes: `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`, `GET /api/v1/auth/session`, `GET /api/v1/auth/credential-state`, `POST /api/v1/auth/password/change`, `POST /api/v1/auth/mfa/totp/begin`, `POST /api/v1/auth/mfa/totp/complete`,
- incident discovery, creation, retrieval, and incident metadata mutation: `POST /api/v1/incidents`, `GET /api/v1/incidents`, `GET /api/v1/incidents/{incident_id}`, `PATCH /api/v1/incidents/{incident_id}`,
- deployment-local user account inspection and administration: `GET /api/v1/users`, `POST /api/v1/users`, `GET /api/v1/users/{user_id}`, `PATCH /api/v1/users/{user_id}`, `POST /api/v1/users/{user_id}/password/reset`, `POST /api/v1/users/{user_id}/mfa/totp/reset`, `POST /api/v1/users/{user_id}/sessions/revoke-all`,
- incident membership inspection and administration: `GET /api/v1/incidents/{incident_id}/memberships`, `POST /api/v1/incidents/{incident_id}/memberships`, `PATCH /api/v1/incidents/{incident_id}/memberships/{user_id}`, `DELETE /api/v1/incidents/{incident_id}/memberships/{user_id}`,
- view-schema discovery: `GET /api/v1/view-schemas`, `GET /api/v1/view-schemas/{view_schema_id}`,
- extension-claim discovery: `GET /api/v1/extensions`,
- saved-view discovery and persistence: `GET /api/v1/incidents/{incident_id}/saved-views`, `POST /api/v1/incidents/{incident_id}/saved-views`, `PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`, `DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`,
- workbook-preference discovery and persistence plus startup selection: `GET /api/v1/incidents/{incident_id}/workbook-preferences/me`, `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me`, `GET /api/v1/incidents/{incident_id}/workbook-preferences/default`, `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default`, `GET /api/v1/incidents/{incident_id}/workbook-startup`,
- workbook query and row creation: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`,
- record mutation, explicit Timeline capture-state actions, soft-delete, restore, history, rollback, and same-field conflict resolution: `PATCH /api/v1/records/{record_id}`, `POST /api/v1/records/{record_id}/mark-reviewed`, `POST /api/v1/records/{record_id}/supersede`, `DELETE /api/v1/records/{record_id}`, `POST /api/v1/records/{record_id}/restore`, `GET /api/v1/records/{record_id}/history`, `POST /api/v1/records/{record_id}/rollback`, `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve`,
- entity merge initiation: `POST /api/v1/records/{survivor_record_id}/merge`,
- entity-mention explicit action route: `POST /api/v1/entity-mentions/{entity_mention_id}/resolve`,
- blob-slot creation and evidence access: `POST /api/v1/object-blobs`, `POST /api/v1/evidence-records/{record_id}/attach-blob`, `POST /api/v1/evidence-records/{record_id}/preview-handle`, `POST /api/v1/evidence-records/{record_id}/download-handle`, `GET /api/v1/evidence-handles/{handle_token}`,
- background-job status and cancellation: `GET /api/v1/jobs/{job_id}`, `POST /api/v1/jobs/{job_id}/cancel`.
The base profile defines no public WebAuthn or passkey routes under `/api/v1/auth/*` or `/api/v1/users/{user_id}/mfa/*`. Registration, assertion, credential enumeration, credential deletion, and reset semantics for WebAuthn or passkeys are reserved for future specification work and MUST NOT be claimed by base-profile implementations.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-186, AC-187, AC-231, AC-251, AC-252, AC-253, AC-254, AC-255, AC-340, AC-341, AC-342, AC-334, AC-335, AC-336, AC-337, AC-338, AC-339, AC-370, AC-371

**REQ-01-033**
Implementations that claim an extension profile MUST add that profile's route family under the same versioned root rather than overloading base workbook routes. This includes, at minimum, `/api/v1/import-sessions/*`, `/api/v1/reference-packs/*`, `/api/v1/snapshots/*` and `/api/v1/releases/*`, `/api/v1/incident-bundles/*`, `/api/v1/auth/providers/*`, `/api/v1/auth/oidc/*`, `/api/v1/auth/saml/*`, and `/api/v1/users/{user_id}/auth-bindings*` for the corresponding claimed extension profiles.
Profiles: base, import, snapshot_reporting, incident_portability, reference_pack, enterprise_authentication
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-186, AC-187, AC-231, AC-232, AC-233, AC-234, AC-235, AC-236, AC-370, AC-371

Core 01 §17 and §20 are the primary owners for the public route inventory, request and response defaults, omitted-versus-`null` behavior, route-scoped idempotency, family-specific error registries, and durable terminal-state representation for those extension families.

**REQ-01-570**
The current profile defines no public `/api/v1/backups*`, `/api/v1/restores*`, or `/api/v1/restore-verifications*` route family and no corresponding `/ws/v1/*` family. Backup creation, restore execution, and restore verification remain deployment-local operator-facing concerns and MUST NOT be exposed as workbook-surface routes in the current profile.
Profiles: base
Verified by: AC-402

##### 3.3.3.1 Runtime extension discovery and reserved-unclaimed extension semantics

**REQ-01-542**
`GET /api/v1/extensions` MUST be a base-profile deployment-scoped discovery route. It MUST return the common success envelope with `data.extensions[]`. `extensions[]` MUST be ordered by `profile_id asc`. Each `extensions[]` item MUST contain exactly `profile_id`, `claimed`, and `route_families[]`. This route MUST expose only deployment extension-claim state and reserved family roots. It MUST NOT expose provider secrets, provider metadata or claim maps, reference-pack version state, snapshot or release state, incident-bundle state, or other live extension-family payload.
Profiles: base
Verified by: AC-370

**REQ-01-543**
`data.extensions[]` MUST enumerate all current-profile extension identifiers, including unclaimed ones. The current-profile `profile_id` values are exactly:

- `enterprise_authentication`
- `import`
- `incident_portability`
- `reference_pack`
- `snapshot_reporting`

Clients MUST ignore unknown additive members on each item and MUST ignore unknown future `profile_id` values.
Profiles: base
Verified by: AC-370

**REQ-01-544**
`route_families[]` MUST list reserved family roots rather than full per-route inventories. `route_families[]` MUST be ordered by route-family string asc. The current-profile mapping is exactly:

- `enterprise_authentication`
  - `/api/v1/auth/oidc`
  - `/api/v1/auth/providers`
  - `/api/v1/auth/saml`
  - `/api/v1/users/{user_id}/auth-bindings`
- `import`
  - `/api/v1/import-sessions`
- `incident_portability`
  - `/api/v1/incident-bundles`
- `reference_pack`
  - `/api/v1/reference-packs`
- `snapshot_reporting`
  - `/api/v1/releases`
  - `/api/v1/snapshots`
Profiles: base
Verified by: AC-370, AC-371

**REQ-01-545**
`GET /api/v1/extensions` is a singleton discovery route. It MUST reject `limit`, `cursor_token`, and pagination aliases with `400`, `error.code = invalid_pagination_request`, and `error.details.reason_code = pagination_not_supported`.
Profiles: base
Verified by: AC-370

**REQ-01-546**
For `GET /api/v1/extensions`, `claimed=true` means the deployment currently claims that extension profile for conformance. `claimed` MUST NOT imply that the corresponding route family currently contains resources or that the current caller is authorized for every route in that family. A claimed extension family with zero current resources or a claimed family route denied by ordinary authorization or family-specific policy MUST use the ordinary claimed-family behavior rather than `extension_profile_not_claimed`.
Profiles: base
Verified by: AC-370, AC-371

**REQ-01-547**
If a request path matches one of the reserved `route_families[]` for a profile whose `claimed=false`, the server MUST return the common error envelope with `404`, `error.code = extension_profile_not_claimed`, `error.retryable = false`, and `error.details` containing at minimum `profile_id` and `route_family`. `error.details.route_family` MUST be the matched family root string from REQ-01-544 rather than the caller's requested path. A request matches a reserved family when its public path is the declared family root itself or a descendant route under that family-root template.
Profiles: base
Verified by: AC-371

**REQ-01-548**
Reserved-extension dispatch MUST use this precedence:

1. a base-profile route match dispatches first,
2. a claimed extension-family route match dispatches second and uses ordinary claimed-family behavior,
3. a reserved but unclaimed extension-family match returns `404 error.code = extension_profile_not_claimed` before family-specific authorization or policy evaluation,
4. ordinary unknown-route handling applies only to paths outside all reserved base and extension families.

For a claimed extension family, the server MUST NOT use `extension_profile_not_claimed` solely because the family has zero current resources or the caller lacks authorization for one route.
Profiles: base
Verified by: AC-371

#### 3.3.4 View-shaped read contract

**REQ-01-034**
The primary hot-path read route for workbook surfaces MUST be `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`.

This subsection owns one canonical view-row wire family reused across workbook query, row-refreshing create or patch success, and replayable collaboration patches:

- `view_row_v1`: the full row object used for query `rows[]`, successful `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`, successful `PATCH /api/v1/records/{record_id}` row refresh, and any other route that explicitly returns a full workbook row refresh.
- `view_row_patch_v1`: the sparse row object used for `record_changed.payload.affected_views[].patch_cells` when `change_kind='patch'`.

For this wire family:

- `record_id` and `row_version` MUST remain top-level members and MUST NOT be duplicated inside `cells`,
- `cells` MUST be a map keyed only by stable `field_key`,
- each base-profile cell object MUST use exactly `{ "value": <field-read-payload> }`,
- `cells[field_key].value` MUST reuse the field's existing read contract rather than retyping all fields to one scalar family,
- `group_values` MUST remain a top-level sibling of `cells` and MUST NOT be treated as a writable cell family,
- full-row field membership is governed by REQ-01-310,
- `view_row_patch_v1.cells` MUST contain only changed fields and `view_row_patch_v1.group_values`, when present, MUST contain only changed grouping scalars,
- clients MUST ignore unknown additive members inside row or cell objects unless a later profile explicitly makes them required.

Illustrative `view_row_v1` shape:

```json
{
  "record_id": "rec_01",
  "row_version": 42,
  "cells": {
    "evidence.title": { "value": "EDR package for WS-023" },
    "evidence.collector_party_text": { "value": "IR Vendor" },
    "evidence.collector_party_id": { "value": null },
    "evidence.source_party_text": { "value": "Endpoint team" },
    "evidence.source_party_id": { "value": "pty_22" }
  },
  "group_values": {
    "timeline.has_evidence": true
  }
}
```

Illustrative `view_row_patch_v1` shape:

```json
{
  "record_id": "rec_01",
  "row_version": 43,
  "cells": {
    "evidence.collector_party_id": { "value": null },
    "evidence.edited_at": { "value": "2026-04-06T11:57:00Z" }
  },
  "group_values": {
    "timeline.has_evidence": false
  }
}
```
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-366, AC-367, AC-368

Contract tables. The tables in §3.3.4 through §3.3.4.2 are the compact owner-local contract for query shape, omission and canonicalization behavior, response shape, and record-history paging. Algorithmic comparison, tokenization, and ordering rules remain in the surrounding prose.

**Table 3.3.4-A. View-query route summary**

| Route | Auth context | Request contract summary | Omission and default summary | Response summary | Primary errors |
| --- | --- | --- | --- | --- | --- |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` | Visible incident and visible view schema for the caller | Optional `sort[]`, optional `filters[]`, optional `group_by`, and pagination body members under §3.3.7 | Omitted `sort` or `sort: []` means no user sort override; omitted `filters` or `filters: []` means no filters; omitted `group_by` means grouping inactive; `group_by: null` is invalid | Returns `incident_id`, `view_schema_id`, `rows[]` as full `view_row_v1[]`, `meta.query`, and paging metadata | `invalid_view_query`, `invalid_pagination_request`, ordinary authorization failures |

**Table 3.3.4-B. Query request members**

| Member | Requiredness | Allowed shape | Omission and explicit-`null` behavior | Canonicalization and validation |
| --- | --- | --- | --- | --- |
| `sort[]` | Optional | Ordered array of `sort` entries | Omitted or `[]` means no user sort override | At most 8 raw entries; duplicate normalized `field_key` values are invalid; effective applied sort appends default-sort tail and then `record_id asc` when absent |
| `filters[]` | Optional | Array of filter objects keyed by `field_key` | Omitted or `[]` means no filters | At most 16 raw entries; request order is non-semantic; canonical persisted order is `field_key asc` |
| `group_by` | Optional | One declared grouping-key string | Omitted means grouping inactive; explicit `null` is invalid | At most one active grouping key in the current profile |
| `limit` | Optional | Pagination member under §3.3.7 | Body member only; query-parameter form is invalid | Counts serialized `rows[]` entries only |
| `cursor_token` | Optional | Pagination member under §3.3.7 | Body member only; query-parameter form is invalid | Bound to the normalized view-query contract |

**Table 3.3.4-C. `sort[]` and grouping contract**

| Item | Required members | Closed vocabulary or rule | Canonical behavior |
| --- | --- | --- | --- |
| `sort[]` entry | `field_key`, `direction` | `direction` is exactly `asc` or `desc` | `field_key` must be one declared sortable key for the active `view_schema_id`; unknown top-level members are invalid |
| Effective applied sort | N/A | User list in request order, then remaining schema `default_sort`, then `record_id asc` when absent | User sort overrides matching default-sort entries and extends rather than replaces the default tuple |
| `group_by` | N/A | One declared grouping key string | Omitted in `meta.query` when grouping is inactive |

**Table 3.3.4-D. Query response and row shape summary**

| Member | Presence | Nullability and canonicalization | Notes |
| --- | --- | --- | --- |
| `incident_id` | Required | Non-null | Echoes the addressed incident |
| `view_schema_id` | Required | Non-null | Echoes the addressed view schema |
| `rows[]` | Required | May be empty | Each entry is full `view_row_v1` |
| `rows[].record_id` | Required | Non-null | Top-level technical identifier |
| `rows[].row_version` | Required | Non-null | Top-level technical concurrency identifier |
| `rows[].cells` | Required | Includes every schema-declared non-technical field | `{ "value": null }` means authoritative null; omission of a schema-declared non-technical field is invalid |
| `rows[].group_values` | Conditional | Present only when the schema declares grouping keys | Full current grouping object for the schema |
| `meta.query.filters[]` | Required | Canonical normalized filter wire shape | Ordered by `field_key asc` |
| `meta.query.sort[]` | Required | Effective applied sort after default-tail expansion | Always present |
| `meta.query.group_by` | Conditional | Omitted when grouping inactive | Never serialized as JSON `null` |
| Paging metadata | Required when §3.3.7 requires it | Cursor contract under §3.3.7 | Transport metadata only |


**REQ-01-035**
A view-query request MUST be view-shaped rather than table-shaped. It MUST accept:

- optional ordered `sort[]` entries,
- optional `filters[]` entries keyed by `field_key`,
- optional `group_by`,
- pagination members defined by §3.3.7.

Each `sort[]` entry MUST use exactly this shape:

```json
{
  "field_key": "timeline.sort_ts",
  "direction": "asc"
}
```

For `sort[]`:

- `field_key` MUST be a stable sortable field key declared in the active `view_schema_id`'s `sort_fields`,
- `field_key` MUST NOT be a visible column label, projection-table column name, or storage-table column name,
- `direction` MUST use the exact closed vocabulary `asc` and `desc`,
- unknown top-level members in a `sort[]` entry MUST be rejected,
- duplicate normalized `field_key` entries in one request MUST be rejected,
- `sort[]` MAY be omitted or empty, but when present it MUST contain at most `8` entries,
- the count is the raw parsed array length before duplicate-field rejection, per-entry normalization, default-sort tail expansion, or cursor comparison,
- if `sort[]` exceeds `8`, the server MUST fail with `400`, `error.code = invalid_view_query`, `error.details.reason_code = sort_count_exceeded`, `error.details.field = sort`, `error.details.requested_count = <raw count>`, and `error.details.max_count = 8`,
- the server MUST NOT truncate or partially honor an oversize `sort[]`,
- omitted `sort` and `sort: []` mean no user sort override.

`group_by` is optional. Omitted `group_by` means `Group: None`. When present, `group_by` MUST be a non-null string equal to one declared grouping key for the active `view_schema_id`. `group_by: null` is invalid. The current profile allows at most one active grouping key.

The effective applied sort tuple for this route MUST be computed by taking the normalized user `sort[]` entries in request order, then appending each schema `default_sort` entry whose `field_key` does not already appear in the user list, then appending `record_id asc` if it is still absent. A user sort therefore overrides matching default-sort entries and extends rather than replaces the schema default tuple.

For `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, `limit` and `cursor_token` MUST appear only as JSON-body members, not as query parameters. `limit` counts serialized `rows[]` entries only. A malformed pagination member, unsupported pagination alias, or cursor replay against a different bound view-query contract MUST fail with `400`, `error.code=invalid_view_query`, and `error.details.reason_code` equal to `invalid_limit` or `cursor_query_mismatch` from §3.3.6.2, as applicable.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-238, AC-239, AC-240, AC-243, AC-359, AC-360, AC-361, AC-372, AC-373, AC-374, AC-375

**REQ-01-036**
A successful view-query response MUST return the common success envelope with:

- `incident_id`,
- `view_schema_id`,
- `rows[]` serialized as `view_row_v1[]`,
- `meta.query` containing the normalized applied view-query contract for this response,
- pagination metadata defined in §3.3.7.

For this route, each `rows[]` entry MUST be a full `view_row_v1`. `rows[].cells` MUST include every schema-declared non-technical field for the active `view_schema_id`, regardless of whether the field is visible, default-hidden, writable, or read-only. The only schema fields not serialized under `cells` are the hidden technical fields `record_id` and `row_version`, which remain top-level. In a full row, a schema-declared non-technical field serialized as `{ "value": null }` means authoritative null. Omission of a schema-declared non-technical field is invalid.

For a schema that declares one or more grouping keys, each full row MUST include the full current `group_values` object for that schema. When a schema declares no grouping keys, `group_values` MUST be omitted.

For this route, `meta.query` MUST always be present. `meta.query.filters[]` MUST use the canonical normalized wire shape defined in §3.3.4.1. `meta.query.sort[]` MUST serialize the effective applied sort tuple after the default-tail expansion defined by REQ-01-035. `meta.query.group_by` MUST be omitted when grouping is inactive.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-238, AC-241, AC-243, AC-361, AC-366, AC-367, AC-372, AC-373, AC-374

**REQ-01-037**
The server MUST NOT serialize group headers or other presentation-only grouping artifacts as writable rows.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-243

##### 3.3.4.1 Filter predicate wire contract


**Table 3.3.4.1-A. Filter operator and argument matrix**

| `op` | Required `arg` shape | Default field-class eligibility | Canonicalization and validation notes |
| --- | --- | --- | --- |
| `eq` | Exactly one of `{ "value": <scalar-or-null> }` or `{ "values": [<scalar>, ...] }` | Enum, boolean, timestamp, date, scalar identifier text | `values[]` is set-like, non-empty, and canonicalized to the unique normalized member set in canonical ascending order |
| `range` | One or more of `gt`, `gte`, `lt`, `lte` | Timestamp and date | At least one bound is required; `gt` and `gte` cannot coexist; `lt` and `lte` cannot coexist; contradictory normalized bounds are invalid |
| `contains_any` | `{ "values": [<scalar>, ...] }` | Multi-value collection fields | `values[]` is set-like, non-empty, and canonicalized after normalization |
| `contains_all` | `{ "values": [<scalar>, ...] }` | Multi-value collection fields | `values[]` is set-like, non-empty, and canonicalized after normalization |
| `prefix` | `{ "value": <string> }` | Scalar identifier text fields | Uses the shared query-time text-comparison substrate; match is anchored at code-point offset `0`; canonical stored form is the comparison-normalized prefix string |
| `full_text` | `{ "query": <string> }` | Declared synthetic full-text predicate fields | Token order is non-semantic; duplicates coalesce; canonical stored form is the unique normalized token set sorted ascending and joined by one ASCII space |

**Table 3.3.4.1-B. `invalid_view_query` reason summary**

| Condition | Transport | `error.code` | `reason_code` and required detail |
| --- | --- | --- | --- |
| `sort[]` raw count exceeds 8 | `400` | `invalid_view_query` | `sort_count_exceeded`; `field=sort`; `requested_count`; `max_count=8` |
| `filters[]` raw count exceeds 16 | `400` | `invalid_view_query` | `filter_count_exceeded`; `field=filters`; `requested_count`; `max_count=16` |
| Cursor replay does not match the bound normalized query | `400` | `invalid_view_query` | `cursor_query_mismatch` |
| Malformed or unsupported body-level pagination member | `400` | `invalid_view_query` | `invalid_limit` or the most specific pagination reason from §3.3.6.2 |
| Duplicate normalized `field_key`, illegal `op`, empty set-like operand, empty `prefix`, empty `full_text`, or zero full-text tokens after normalization | `400` | `invalid_view_query` | Most specific field-level reason from REQ-01-043 through REQ-01-046; `filter_index` is required and `field_key` is present when available |
| Contradictory or malformed range bounds | `400` | `invalid_view_query` | Most specific range reason from §3.3.6.2; `filter_index` is required |


**REQ-01-038**
A `filters[]` entry MUST use this shape:

`filters[]` MAY be omitted or empty. When present, it MUST contain at most `16` entries. The count is the raw parsed array length before duplicate-field rejection or operand normalization. If `filters[]` exceeds `16`, the server MUST fail with `400`, `error.code = invalid_view_query`, `error.details.reason_code = filter_count_exceeded`, `error.details.field = filters`, `error.details.requested_count = <raw count>`, and `error.details.max_count = 16`. The server MUST NOT truncate or partially honor an oversize `filters[]`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-360

```json
{
  "field_key": "task.status",
  "op": "eq",
  "arg": {
    "values": ["open", "blocked"]
  }
}
```

The closed v1 `op` vocabulary is:

- `eq`
- `range`
- `contains_any`
- `contains_all`
- `prefix`
- `full_text`

**REQ-01-039**
`field_key` MUST be a stable filterable field key, or a stable synthetic predicate key, declared by the active `view_schema_id`. It MUST NOT be a visible column label, visible tab label, SQL expression, projection-table column name, or storage-table column name.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-040**
`arg` MUST follow the operator-specific shape:

- `eq`: exactly one of:
  - `{ "value": <scalar-or-null> }`
  - `{ "values": [<scalar>, ...] }`
  `values[]` means one-of exact-match inclusion for that scalar field.
- `range`: one or more of `gt`, `gte`, `lt`, `lte`.
  At least one bound MUST be present.
  `gt` and `gte` MUST NOT both be present.
  `lt` and `lte` MUST NOT both be present.
- `contains_any`: `{ "values": [<scalar>, ...] }`
- `contains_all`: `{ "values": [<scalar>, ...] }`
- `prefix`: `{ "value": <string> }`
- `full_text`: `{ "query": <string> }`
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

Unless a view schema explicitly overrides the rule, operator eligibility is type-driven as follows:

- enum and boolean fields allow `eq`,
- timestamp and date fields allow `eq` and `range`,
- scalar identifier text fields allow `eq` and `prefix`,
- multi-value collection fields allow `contains_any` and `contains_all`,
- declared full-text predicate fields allow `full_text`.

**REQ-01-041**
A scalar operand MUST be one of JSON `string`, `number`, `boolean`, or `null`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-042**
Date operands MUST use canonical `YYYY-MM-DD`.
Timestamp operands MUST use RFC 3339.
The server MUST apply the field contract's declared normalization before comparison.

For `eq.values[]`, `contains_any.values[]`, and `contains_all.values[]`:

- `values[]` is a logical set, not an ordered list,
- caller order is non-semantic,
- duplicate members are allowed on input but MUST coalesce after operator-specific normalization,
- JSON `null` MUST NOT appear inside any set-like `values[]`; `eq.value = null` is the only null-equality encoding,
- the canonical normalized form used for `meta.query`, saved-view persistence, and cursor binding MUST serialize the unique normalized member set in canonical ascending order under the field's deterministic comparison semantics.

For `prefix`:

- comparison MUST use the shared query-time text-comparison substrate defined by REQ-01-488,
- a match exists only when the comparison-normalized field value begins at code-point offset `0` with the comparison-normalized `prefix.value`,
- `prefix` is not infix search, wildcard search, regex, or token search,
- the canonical normalized form used for `meta.query`, saved-view persistence, and cursor binding MUST store `arg.value` in that comparison-normalized form.

For `full_text`:

- the declared full-text predicate's source fields MUST be normalized using their bound field contracts, with null source fields treated as empty,
- tokenization MUST split text into maximal contiguous runs of Unicode letters or Unicode numbers, with every other code point acting as a separator that is discarded,
- comparison MUST use the same shared query-time text-comparison substrate defined by REQ-01-488,
- query token order is non-semantic and duplicate query tokens MUST coalesce after normalization,
- a row matches only when every unique normalized query token appears as an exact token in the union of the predicate's declared source fields,
- `full_text` is not phrase search, infix search inside a token, wildcard syntax, stemming, fuzzy matching, transliteration, or storage-engine query language,
- the canonical normalized form used for `meta.query`, saved-view persistence, and cursor binding MUST store `arg.query` as the unique normalized token set sorted ascending and joined by one ASCII space.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-565**
The current profile defines no public fuzzy, trigram, phrase, wildcard, stemming, accent-insensitive, transliteration, or language-aware operator on `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`. `full_text` remains the exact-token operator defined in this subsection.
Profiles: base
Verified by: AC-231, AC-387

**REQ-01-566**
For the current profile, `full_text` determines row membership only. The applied sort tuple defined by REQ-01-035 continues to determine result order. The route MUST NOT inject relevance scores, implicit ranking, thresholds, or implicit sort override into the response.
Profiles: base
Verified by: AC-231, AC-387

**REQ-01-567**
The current profile defines no generic public discovery route for alias-, mention-, similarity-, or suggestion-based candidate lookup. Any future public discovery surface requires a new operator or a new route family with its own exact contract.
Profiles: base
Verified by: AC-231, AC-387

`filters[]` combine by conjunction.

Within one filter:

- `eq` with `values[]` is disjunctive across those values,
- `contains_any` matches when any normalized member matches,
- `contains_all` matches when all normalized members match,
- `range` matches when all supplied bounds hold.

**REQ-01-043**
The server MUST reject, using the common error envelope:

- a `field_key` not declared filterable for the active view,
- an `op` not allowed for that field's filter class,
- duplicate `field_key` entries after normalization,
- an empty `values[]`,
- JSON `null` inside any set-like `values[]`,
- a set-like operand that is empty after operator-specific normalization or duplicate coalescing,
- an empty `prefix.value` after trimming,
- an empty `full_text.query` after trimming,
- a `full_text.query` that tokenizes to zero tokens after normalization,
- a `range` with no bounds,
- a `range` that specifies both `gt` and `gte`,
- a `range` that specifies both `lt` and `lte`,
- contradictory range bounds after normalization,
- unknown top-level members other than `field_key`, `op`, and `arg`,
- unknown `arg` members for the selected `op`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-044**
For these failures, the server MUST return `400` and `error.code = invalid_view_query`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-045**
`error.details` MUST identify, at minimum:

- `filter_index`,
- `field_key` when present,
- `reason_code`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

For these failures, `reason_code` MUST be `empty_values_after_normalization` when a set-like operand becomes empty only after normalization or duplicate coalescing, MUST be `empty_full_text_after_tokenization` when a non-empty trimmed `full_text.query` yields zero tokens after normalization, and otherwise MUST use the most specific code from §3.3.6.2.

The canonical `invalid_view_query` `error.details.reason_code` registry is defined in §3.3.6.2.

Request order of `filters[]` is not semantically significant.

**REQ-01-046**
For saved-view persistence, cursor binding, and `meta.query`, the server MUST normalize the view-query contract as follows:

- `filters[]` MUST use the exact wire shape defined in this subsection and MUST be ordered canonically by `field_key asc`,
- within `filters[]`, set-like `arg.values[]` MUST serialize the unique normalized member set in canonical ascending order whenever the selected `op` uses `values[]`,
- within `filters[]`, `prefix.arg.value` MUST serialize the comparison-normalized value defined by REQ-01-042,
- within `filters[]`, `full_text.arg.query` MUST serialize the canonical normalized token string defined by REQ-01-042,
- `sort[]` MUST preserve request order after per-entry normalization and duplicate rejection,
- `group_by` MUST be omitted when grouping is inactive.

For saved-view persistence, normalized `query_json.sort` stores only the normalized user sort override list and MUST use `[]` as the canonical persisted representation when no user sort override exists. Persisted `query_json.group_by` MUST be omitted when grouping is inactive. For `meta.query`, `sort[]` stores the effective applied sort tuple after the default-tail expansion defined by REQ-01-035.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-360, AC-361

**REQ-01-047**
A declared full-text predicate that spans multiple read fields MUST have its own stable synthetic `field_key`. Clients MUST address that predicate by its declared synthetic key rather than by sending raw field lists or storage-specific search syntax.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

Examples:

```json
{
  "filters": [
    {
      "field_key": "timeline.capture_state",
      "op": "eq",
      "arg": { "values": ["rough", "enriched"] }
    },
    {
      "field_key": "timeline.occurred_day",
      "op": "range",
      "arg": { "gte": "2026-03-01", "lt": "2026-04-01" }
    },
    {
      "field_key": "timeline.tags",
      "op": "contains_any",
      "arg": { "values": ["rough", "needs_followup"] }
    },
    {
      "field_key": "identity.upn",
      "op": "prefix",
      "arg": { "value": "john." }
    }
  ]
}
```

##### 3.3.4.2 Record-history read contract


**Table 3.3.4.2-A. Record-history route summary**

| Route | Auth context | Request contract summary | Ordering and paging | Success summary | Primary errors |
| --- | --- | --- | --- | --- | --- |
| `GET /api/v1/records/{record_id}/history` | Caller must be able to see the addressed record; otherwise `404` | Record-scoped singleton history read with optional `limit` and `cursor_token` under §3.3.7 | `items[]` is newest-first; pagination is transport-only, preserves order, and the cursor is bound to `record_id`; the full retained logical history remains visible in the current profile | Returns `incident_id`, `record_id`, current `row_version`, `deleted`, and `items[]`; for a soft-deleted record the returned `row_version` is the tombstone concurrency token required for restore | `invalid_pagination_request`, ordinary authorization failures |

**Table 3.3.4.2-B. History item members and rollback availability**

| Member | Presence rule | Notes |
| --- | --- | --- |
| `actor_user_id` | Required on every item | Attributed actor of the committed change |
| `committed_at` | Required on every item | Newest-first committed ordering anchor |
| `operation` | Required on every item | Displayable history operation label |
| `diff_summary` | Required on every item | Row-centric summary only |
| `change_set_id` | Required on every item | Stable change-set anchor |
| `reversible` | Required on every item | Current reversibility state, not historical omission |
| `available_rollback_actions[]` | Required on every item | Ordered only as `history_entry`, `change_set`, `row_restore`; empty when `reversible=false` |
| `history_entry_ref` | Present if and only if the logical history item maps to exactly one mutation target eligible for `target.kind='history_entry'` | Stable opaque reference for the retained-history lifetime of the record in the current deployment |
| `revision_no` | Present if and only if whole-row restore is legal for that displayed logical history item | Row-restore selector only |


**REQ-01-048**
`GET /api/v1/records/{record_id}/history` MUST return row-centric history for the addressed record. The route is record-scoped, not view-scoped. A caller who lacks visibility to `record_id` MUST receive `404`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-049**
The success-envelope `data` for this route MUST include at minimum:

- `incident_id`,
- `record_id`,
- `row_version`,
- `deleted`,
- `items[]`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-050**
For a soft-deleted record, `row_version` MUST expose the current tombstone `row_version`. That value is the required restore-concurrency token for `POST /api/v1/records/{record_id}/restore`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-051**
`items[]` MUST be serialized in newest-first committed order. If two logical history items derive from the same committed `change_set`, their relative order MUST remain deterministic and stable for identical history state.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-052**
Each `items[]` entry MUST include at minimum:

- `actor_user_id`,
- `committed_at`,
- `operation`,
- `diff_summary`,
- `change_set_id`,
- `reversible`,
- `available_rollback_actions[]`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-053**
`available_rollback_actions[]` MUST draw only from `history_entry`, `change_set`, and `row_restore`. The server MUST serialize `available_rollback_actions[]` in that canonical order with unavailable actions omitted. If `reversible=false`, `available_rollback_actions[]` MUST be empty.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-054**
An `items[]` entry MUST include `history_entry_ref` if and only if that logical history item maps to exactly one mutation target eligible for `target.kind='history_entry'` addressing. The client MUST treat `history_entry_ref` as opaque. For the retained-history lifetime of that record in the current deployment, the same logical history item MUST keep the same `history_entry_ref` across repeated reads.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-383, AC-384

**REQ-01-055**
An `items[]` entry MUST include `revision_no` if and only if whole-row restore is a legal action for that displayed logical history item.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-056**
The route MUST accept `limit` and `cursor_token` under §3.3.7 whenever more than one page is possible. Pagination MUST preserve the item-ordering rules in this subsection and the cursor MUST remain bound to `record_id`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-215, AC-231

**REQ-01-561**
In the current profile, `GET /api/v1/records/{record_id}/history` MUST return the full retained logical history for the addressed extant record in the current deployment. Pagination is transport-only and MUST NOT be used as retention-based truncation.
Profiles: base
Verified by: AC-231, AC-383

**REQ-01-562**
Incident closure MUST NOT remove previously visible history items, remove previously issued `history_entry_ref` values, or otherwise narrow the `/history` surface or rollback-route contract for extant records.
Profiles: base
Verified by: AC-231, AC-383

**REQ-01-563**
When a previously single-entry-addressable history item later becomes not currently reversible because of dependent later changes, stale target state, or other already-defined rollback-precondition reasons, that item MUST remain present in `items[]`. The route MUST express current legality through `reversible` and `available_rollback_actions[]` rather than by omitting the item or removing `history_entry_ref`.
Profiles: base
Verified by: AC-231, AC-384

#### 3.3.5 Mutation contract


Contract tables. The tables in §3.3.5 through §3.3.5.5 are the compact owner-local mutation contract. Per-field create defaults, create-time writeability, and omitted-versus-`null` behavior remain owned by the relevant field registries in §7.4 and §19 and are referenced here rather than duplicated.

**Table 3.3.5-A. Mutation route index**

| Route | Target scope | Request contract summary | Replay and idempotency | Success summary | Primary error codes |
| --- | --- | --- | --- | --- | --- |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | View-scoped row create | JSON object with required `client_txn_id` and zero or more create-time `field_key` members allowed by the addressed view | Keyed by `(actor_user_id, incident_id, view_schema_id, client_txn_id)` | First success `201 Created`; replay returns the original committed row refresh | `invalid_mutation_payload`, `client_txn_conflict` |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste` | View-scoped clipboard batch | Required `view_schema_id`, `client_txn_id`, `clipboard_text`, `start_field_key`, `columns[]`, and `targets[]`; optional `format` | Keyed by `(actor_user_id, incident_id, view_schema_id, client_txn_id)` | `200 OK` with batch result containing `view_schema_id`, optional `change_set_id`, `rows[]`, and ordered `conflicts[]` | `invalid_mutation_payload`, `client_txn_conflict`, `row_version_conflict`, entity-match conflicts where applicable |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/bulk-mutations` | View-scoped explicit bulk mutation batch | Required `view_schema_id`, `client_txn_id`, `kind`, and stable `targets[]`; command-specific fields are owned by Core 03 §13.3 | Keyed by `(actor_user_id, incident_id, view_schema_id, client_txn_id)` | `200 OK` with batch result containing `view_schema_id`, optional `change_set_id`, `rows[]`, and ordered `conflicts[]` | `invalid_mutation_payload`, `client_txn_conflict`, `row_version_conflict` |
| `PATCH /api/v1/records/{record_id}` | Record-scoped partial field mutation | Required `view_schema_id`, `base_row_version`, `client_txn_id`, and non-empty `changes[]` | Keyed by `(actor_user_id, record_id, client_txn_id)`; exact replay wins before fresh optimistic-concurrency evaluation | `200 OK` with original committed row refresh on success or exact replay | `invalid_mutation_payload`, `client_txn_conflict`, `row_version_conflict`, `same_field_conflict` |
| `POST /api/v1/records/{record_id}/mark-reviewed` | Timeline capture-state action | Required `base_row_version`, `client_txn_id`; optional `reason` | Keyed by `(actor_user_id, record_id, client_txn_id)` | `200 OK` with updated lifecycle state summary | `client_txn_conflict`, `row_version_conflict`, `illegal_transition`, `record_deleted_use_restore` |
| `POST /api/v1/records/{record_id}/supersede` | Timeline capture-state action or Decision supersession action, selected by authoritative `records.record_type` | Required `base_row_version`, `client_txn_id`, non-empty `reason`; Timeline target optional `replacement_record_id`; Decision target required `replacement_record_id` | Keyed by `(actor_user_id, record_id, client_txn_id)` | `200 OK` with either the Timeline lifecycle summary or the Decision supersession summary for the selected target type | `client_txn_conflict`, `row_version_conflict`, `illegal_transition`, `record_deleted_use_restore` |
| `DELETE /api/v1/records/{record_id}` | First-class record soft-delete | Required `base_row_version`, `client_txn_id`; optional `reason` | Keyed by `(actor_user_id, record_id, client_txn_id)` | `200 OK` with record-scoped delete summary | `client_txn_conflict`, `row_version_conflict`, `record_already_deleted`, `record_delete_blocked` |
| `POST /api/v1/records/{record_id}/restore` | First-class record restore | Required `base_row_version`, `client_txn_id`; optional `reason` | Keyed by `(actor_user_id, record_id, client_txn_id)` and participates in destructive-operation locking | `200 OK` with record-scoped restore summary | `client_txn_conflict`, `row_version_conflict`, `record_not_deleted`, `record_locked` |
| `POST /api/v1/records/{record_id}/rollback` | Record-history reversal | Required `base_row_version`, `client_txn_id`, and `target` | Keyed by `(actor_user_id, record_id, client_txn_id)` and participates in destructive-operation locking | `200 OK` with rollback summary for the selected target | `invalid_rollback_request`, `client_txn_conflict`, `row_version_conflict`, `rollback_target_not_found`, `rollback_precondition_failed`, `record_locked` |
| `POST /api/v1/records/{survivor_record_id}/merge` | Entity merge | Required `loser_record_id`, `survivor_base_row_version`, `loser_base_row_version`, `client_txn_id`; optional `reason` | Keyed by `(actor_user_id, survivor_record_id, loser_record_id, client_txn_id)` and participates in destructive-operation locking | `200 OK` with merge summary and carried-forward identifiers | `client_txn_conflict`, `row_version_conflict`, `merge_precondition_failed`, `record_locked` |
| `POST /api/v1/entity-mentions/{entity_mention_id}/resolve` | Single mention action | Required `base_mention_row_version`, `client_txn_id`, `action`; optional `resolved_record_id` and `reason` | Keyed by `(actor_user_id, entity_mention_id, client_txn_id)` | `200 OK` with `entity_mention`, `source_record`, and `change_set_id` | `invalid_mutation_payload`, `client_txn_conflict`, `row_version_conflict`, `entity_mention_not_found`, `resolved_record_not_found`, `illegal_transition`, `record_deleted_use_restore` |
| `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve` | Same-field conflict resolution | Record-scoped conflict-resolver surface | Request and result payload are owned by Core 03 §3.3 | Ordinary conflict-resolution success refresh through the addressed record | Same-field conflict and stale-token semantics are owned by Core 03 §3.3 |

Clipboard-paste and explicit bulk-mutation batch results MUST use the common success envelope on valid batch evaluation, including when every target cell becomes a same-field conflict and no row mutation commits. In that conflicts-only case, `rows[]` MUST be empty and `change_set_id` MUST be omitted. Same-field conflict entries in `conflicts[]` MUST use the Core 03 §3.3.4 conflict object. `same_field_conflict` remains the error code for single-record patch conflict responses; clipboard-paste and bulk batch conflicts are batch result members rather than a separate public error family.

**Table 3.3.5-B. Row-create request members**

| Member | Requiredness | Allowed shape | Omission and explicit-`null` behavior | Validation and replay notes |
| --- | --- | --- | --- | --- |
| `client_txn_id` | Required | Stable client transaction token | Omission is invalid; explicit `null` is invalid | Idempotency key component |
| Top-level create-time `field_key` members | Optional | Only writable create-time fields allowed by the addressed `view_schema_id` | Omission, explicit `null`, and default comparison are owned by the bound field contract in §7.4 or §19 | Unknown or forbidden top-level members are rejected before idempotency comparison |
| Zero-field create | Conditional | No create-time `field_key` members present | Legal only when the addressed view contract explicitly allows it | Otherwise `invalid_mutation_payload` |

**Table 3.3.5-C. Patch request members**

| Member | Requiredness | Allowed shape | Omission and explicit-`null` behavior | Validation and replay notes |
| --- | --- | --- | --- | --- |
| `view_schema_id` | Required | Stable `view_schema_id` | Omission is invalid; explicit `null` is invalid | Participates in normalized replay comparison |
| `base_row_version` | Required | Integer row version | Omission is invalid; explicit `null` is invalid | Participates in normalized replay comparison |
| `client_txn_id` | Required | Stable client transaction token | Omission is invalid; explicit `null` is invalid | Idempotency key component |
| `changes[]` | Required | Non-empty array of `changes[]` entries | `[]` is invalid; explicit `null` is invalid | Raw parsed length is capped at 32; duplicate `field_key` entries are invalid; outer array order is non-semantic and canonicalized by `field_key asc` |

**Table 3.3.5-D. `changes[]` entry contract**

| Member | Requiredness | Rule |
| --- | --- | --- |
| `field_key` | Required | One active writable `field_key` for the addressed view and target record |
| `value` | Conditional | Present if and only if the active field contract is a direct write target |
| `action_payload` | Conditional | Present if and only if the active field contract is a write-action target |
| Exactly-one rule | Required | Each entry contains exactly one of `value` or `action_payload` |

**Table 3.3.5-E. `collection_actions_v1` contract**

| Member | Requiredness | Rule |
| --- | --- | --- |
| `kind` | Required | Exactly `collection_actions_v1` |
| `actions[]` | Required | Raw parsed length 1 through 64 inclusive; empty is invalid |
| `actions[]` order | Required | Semantic and preserved in request order within the field-scoped mutation |
| Validation | Required | Unknown `op`, unknown members for the declared `op`, invalid or foreign `item_ref`, or a field-op mismatch fail with `invalid_mutation_payload` |

**Table 3.3.5-F. `collection_value_v1` contract**

| Member | Requiredness | Rule |
| --- | --- | --- |
| `kind` | Required | Exactly `collection_value_v1` |
| `ordered` | Required | Boolean |
| `items[]` | Required | Each item carries stable opaque `item_ref` suitable for later mutation targeting |
| Serialization order | Required | When `ordered=true`, authoritative collection order; when `ordered=false`, canonical ascending `item_ref` order |

**Table 3.3.5-G. Create and patch replay matrix**

| Route | First successful commit | Exact replay | Same key, different normalized request | No prior hit and stale base version |
| --- | --- | --- | --- | --- |
| Row create | `201 Created` with `data.view_schema_id`, `data.change_set_id`, and `data.row` as `view_row_v1` | `200 OK` with the original committed create result, not current mutable row state | `409` with `error.code = client_txn_conflict` | N/A |
| Patch | `200 OK` with `data.view_schema_id`, `data.change_set_id`, and `data.row` as `view_row_v1` | `200 OK` with the original committed patch result, not current mutable row state | `409` with `error.code = client_txn_conflict`; this wins before fresh optimistic-concurrency evaluation | Field-level Core 03 evaluation: different-field edits auto-rebase, same-field edits fail with `same_field_conflict`, and missing or unusable revision history fails closed with `row_version_conflict` |

**Table 3.3.5-H. Timeline action-route contract**

| Route | Required members | Additional members | Legal current states | Role gate | Success summary | Primary failures |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/records/{record_id}/mark-reviewed` | `base_row_version`, `client_txn_id` | Optional `reason` | `rough`, `enriched` | `reviewer` or `admin` | `200 OK` with at least `record_id`, `incident_id`, `row_version`, `capture_state`, and `change_set_id` | `client_txn_conflict`, `row_version_conflict`, `illegal_transition`, `record_deleted_use_restore` |
| `POST /api/v1/records/{record_id}/supersede` for Timeline targets | `base_row_version`, `client_txn_id`, non-empty `reason` | Optional `replacement_record_id` | `rough`, `enriched`, `reviewed` | `reviewer` or `admin` | `200 OK` with at least `record_id`, `incident_id`, `row_version`, `capture_state`, `change_set_id`, `reason`, and `replacement_record_id` | `client_txn_conflict`, `row_version_conflict`, `illegal_transition`, `record_deleted_use_restore` |
| `POST /api/v1/records/{record_id}/supersede` for Decision targets | `base_row_version`, `client_txn_id`, non-empty `reason`, `replacement_record_id` | None | Target `proposed`, `approved`, or `executed`; superseding decision `approved` or `executed` | `reviewer` or `admin` | `200 OK` with at least `view_schema_id='cartulary.view.decisions.v1'`, `change_set_id`, `target_record_id`, `superseding_record_id`, both row versions, normalized `reason`, and target status | `client_txn_conflict`, `row_version_conflict`, `illegal_transition`, `record_deleted_use_restore` |

**Table 3.3.5-I. Delete and restore summary**

| Route | Required members | Role gate | Success summary | Primary failures |
| --- | --- | --- | --- | --- |
| `DELETE /api/v1/records/{record_id}` | `base_row_version`, `client_txn_id`; optional `reason` | `editor`, `reviewer`, or `admin` | `200 OK` with at least `record_id`, `incident_id`, `row_version`, `deleted`, `deleted_at`, `deleted_by_user_id`, and `change_set_id` | `client_txn_conflict`, `row_version_conflict`, `record_already_deleted`, `record_delete_blocked`, `record_deleted_use_restore` for later patch attempts |
| `POST /api/v1/records/{record_id}/restore` | `base_row_version`, `client_txn_id`; optional `reason` | `reviewer` or `admin` | `200 OK` with at least `record_id`, `incident_id`, `row_version`, `deleted=false`, `deleted_at=null`, `deleted_by_user_id=null`, and `change_set_id` | `client_txn_conflict`, `row_version_conflict`, `record_not_deleted`, `record_locked` |


**REQ-01-057**
New row creation MUST use `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`. The request body MUST be a JSON object. It MUST include required `client_txn_id` and MAY include zero or more additional top-level members. Every additional top-level member name MUST be a writable `field_key` declared by the addressed `view_schema_id` and not otherwise forbidden for create by that view contract. The route's top-level namespace is therefore closed to `client_txn_id` plus `field_key` members allowed for create by that view. A request with no field-keyed initial values is permitted only when the addressed view contract explicitly allows zero-field create. Row-create idempotency MUST be keyed by `(actor_user_id, incident_id, view_schema_id, client_txn_id)`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-058**
Existing-row edits MUST use `PATCH /api/v1/records/{record_id}`. The request MUST include:

- `view_schema_id`,
- `base_row_version`,
- `client_txn_id`,
- required non-empty `changes[]`, with each entry keyed by `field_key` and carrying the intended write value or equivalent action payload.

`changes[]: []` is malformed request shape. It MUST NOT be treated as a no-op.

When present, `changes[]` MUST contain at most `32` entries. The count is the raw parsed array length before duplicate-`field_key` rejection, value normalization, or idempotency comparison.

Patch-route idempotency MUST be keyed by `(actor_user_id, record_id, client_txn_id)`. `view_schema_id` MUST participate in normalized replay comparison for this route, but it MUST NOT widen the idempotency-key scope beyond `(actor_user_id, record_id, client_txn_id)`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-299

**REQ-01-059**
Each `changes[]` entry MUST contain `field_key` and exactly one of `value` or `action_payload`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-060**
A field whose active `view_schema` entry declares a direct write target MUST use `value` in `PATCH /api/v1/records/{record_id} changes[]`. A field whose active `view_schema` entry declares a write action MUST use `action_payload` in `PATCH /api/v1/records/{record_id} changes[]`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-061**
When `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` supplies initial writable values keyed directly by `field_key`, a direct-write field MUST use its direct field value as the JSON value and a write-action field MUST use the same object that a patch `changes[]` entry would carry in `action_payload`. For row-create idempotency comparison, the normalized request MUST include only recognized create members after active-view validation and create-time normalization. `client_txn_id` is the idempotency key and MUST NOT be part of that normalized request comparison. For direct-write fields, comparison MUST use the create-time normalized value that would be persisted. For write-action fields, comparison MUST use the semantically validated action payload after any field-specific normalization. Unknown or forbidden top-level members are never part of normalized comparison because the route rejects them. When a field contract makes omission and explicit JSON `null` equivalent for create-time authoritative state, they MUST compare equal. When a field contract declares a create-time default, omission and explicit transmission of that same default MUST compare equal.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-062**
For any field whose active `view_schema` entry declares `conflict_resolution_class=collection_review`, `action_payload.kind` MUST be `collection_actions_v1`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-063**
A `collection_actions_v1` payload MUST use this shape:
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-299

```json
{
  "kind": "collection_actions_v1",
  "actions": [
    { "... field-specific action object ..." }
  ]
}
```

`actions[]` MUST contain at least one and at most `64` entries. The count is the raw parsed array length before semantic validation, duplicate-add coalescing, or application to the target collection. This limit applies anywhere `collection_actions_v1` is accepted, including create-time field-action values and patch-time `action_payload` values.

**REQ-01-064**
A writable `collection_review` field returned by a view query, a successful create response, a successful patch response, or a same-field conflict payload MUST use `collection_value_v1` rather than a raw string array or plain delimited text.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-065**
A `collection_value_v1` payload MUST use this shape:
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

```json
{
  "kind": "collection_value_v1",
  "ordered": true,
  "items": [
    { "... typed item object ..." }
  ]
}
```

**REQ-01-066**
Every `collection_value_v1.items[]` entry MUST carry a stable `item_ref` suitable for later mutation targeting. The client MUST treat `item_ref` as opaque. The server MUST treat `item_ref` as opaque except when validating that it belongs to the patched `record_id`, active `field_key`, and expected collection item kind. Display-only members such as `display_text` MAY aid comparison surfaces but MUST NOT be required as mutation keys.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-067**
When `ordered=true`, the server MUST serialize `items[]` in authoritative collection order. When `ordered=false`, the server MUST serialize `items[]` in canonical ascending `item_ref` order.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-068**
The public mutation surface MUST be field-key-based and partial. The client MUST NOT be required to send full projection rows, full record snapshots, or raw storage-table mutations in order to edit one field.
Profiles: base, snapshot_reporting
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-233

**REQ-01-069**
The server MUST validate each requested change against the active view contract, enforce per-field writeability and `conflict_resolution_class`, and route the write to the authoritative source field or declared write action without exposing internal table layout. For `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`, this validation MUST also reject a body that is not a JSON object, a missing required `client_txn_id`, a supplied `null` for a required non-nullable member, any top-level member other than `client_txn_id` and recognized create-time `field_key` members allowed by the addressed view contract, a field not allowed as an initial writable value for create, a malformed direct value, a malformed action payload, an attempted direct client write to a read-only or server-managed field, or a request with no field-keyed initial values when the addressed view contract does not permit zero-field create. If mutation payload validation fails, the server MUST fail with `400 Bad Request` using the common error envelope and `error.code = invalid_mutation_payload`. At minimum, this applies to an unknown payload `kind`, an unknown action `op`, an action not allowed for the active `field_key`, an invalid or foreign `item_ref`, a malformed direct value, or an action object that contains unknown members for its declared `op`. When the failure is attributable to one top-level request member, `error.details.field` MUST identify that member. For `collection_actions_v1`, the server MUST apply `actions[]` atomically and in request order within the parent field mutation. The public API surface MUST NOT require raw comma-delimited strings or blind full-collection replacement for `collection_review` fields. For `PATCH /api/v1/records/{record_id}`, normalized idempotency comparison MUST run only after request-shape validation, authorization, and target-record visibility succeed. `client_txn_id` is the idempotency lookup key and MUST NOT be part of the normalized request. The normalized request MUST include exact `view_schema_id`, exact `base_row_version`, and canonical `changes[]`. Top-level `changes[]` defines one unordered mutation set keyed by `field_key`. Top-level `changes[]` order MUST NOT be semantically significant for validation, acceptance, execution, or idempotency comparison. An empty `changes[]` array is invalid mutation payload and MUST be rejected rather than treated as a no-op. A duplicate `field_key` entry in `changes[]` is invalid mutation payload and MUST be rejected rather than normalized away. Canonicalization MUST sort outer `changes[]` by `field_key asc`; for direct-write fields comparison MUST use the authoritative normalized `value`, and for write-action fields comparison MUST use the semantically validated normalized `action_payload`. When an action payload is inherently ordered, such as `collection_actions_v1.actions[]`, that inner order MUST remain significant. If one requested patch would yield different valid outcomes depending on the client-supplied outer `changes[]` order, the request is ambiguous and MUST fail with `400 Bad Request`, the common error envelope, and `error.code = invalid_mutation_payload` rather than relying on outer-array order. Any behavior that needs ordered multi-step semantics across different writable concerns MUST be modeled either as one field-scoped `action_payload` whose own contract defines order or as a dedicated action route. For `PATCH /api/v1/records/{record_id}`, a `changes[]` array whose raw parsed length exceeds `32` MUST fail during request-shape validation with `400 Bad Request`, the common error envelope, `error.code = invalid_mutation_payload`, `error.details.reason_code = change_count_exceeded`, `error.details.field = changes`, `error.details.requested_count = <raw count>`, and `error.details.max_count = 32`. For any `collection_actions_v1` payload accepted by `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` or `PATCH /api/v1/records/{record_id}`, an empty `actions[]` array MUST fail during request-shape validation with `400 Bad Request`, the common error envelope, `error.code = invalid_mutation_payload`, `error.details.reason_code = empty_collection_actions`, `error.details.field` identifying the containing member path, and `error.details.field_key` identifying the active collection field. For any such `collection_actions_v1` payload, an `actions[]` array whose raw parsed length exceeds `64` MUST fail during request-shape validation with `400 Bad Request`, the common error envelope, `error.code = invalid_mutation_payload`, `error.details.reason_code = collection_action_count_exceeded`, `error.details.field` identifying the containing member path, `error.details.field_key` identifying the active collection field, `error.details.requested_count = <raw count>`, and `error.details.max_count = 64`. These count failures MUST be evaluated before idempotency replay comparison or write execution, so an oversize mutation never becomes replayable committed state.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-299

**REQ-01-070**
Row-refreshing success responses for `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` and `PATCH /api/v1/records/{record_id}` MUST return the common success envelope with `data.view_schema_id`, `data.change_set_id`, and `data.row`, where `data.row` is exactly `view_row_v1` for the addressed `view_schema_id`. `data.row` is the single authoritative carrier of `record_id` and `row_version` for that refresh. Those two members MUST NOT be duplicated as sibling response members outside `data.row`.

A first-time successful `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` MUST return `201 Created`. If the same authenticated actor replays the same normalized row-create request within the idempotency scope defined by REQ-01-057, the server MUST return `200 OK` with the originally committed create result rather than current mutable row state. If the same authenticated actor reuses `client_txn_id` within that same scope for a different normalized row-create request, the server MUST fail with `409` and `error.code = client_txn_conflict`; `error.details` MUST include at least `client_txn_id`. The replay response MUST therefore return the original `data.view_schema_id`, original `data.change_set_id`, and original committed `data.row`, and MUST create no second record, no second `change_set`, no second revision entry, and no second replayable collaboration event.

A first-time successful `PATCH /api/v1/records/{record_id}` MUST return `200 OK` with the originally committed patch result. If the same authenticated actor replays the same normalized patch request within the idempotency scope defined by REQ-01-058, the server MUST return `200 OK` with the originally committed patch result rather than current mutable row state. If the same authenticated actor reuses `client_txn_id` within that same scope for a different normalized patch request, the server MUST fail with `409` and `error.code = client_txn_conflict`; `error.details` MUST include at least `client_txn_id`. The patch replay response MUST therefore return the original `data.view_schema_id`, original `data.change_set_id`, and the original committed `data.row`, and MUST create no second mutation, no second `change_set`, no second revision entry, and no second replayable collaboration event. Only committed-success patch outcomes are replayable for this route. When a prior committed idempotency hit exists for the same `(actor_user_id, record_id, client_txn_id)`, the server MUST evaluate normalized-request replay before optimistic concurrency. An exact normalized match MUST replay the original success. A different normalized request MUST fail with `409` and `error.code = client_txn_conflict` even when the supplied `base_row_version` is stale. If no prior committed idempotency hit exists and the current `row_version` is greater than `base_row_version`, `PATCH /api/v1/records/{record_id}` MUST apply the Core 03 field-level optimistic-concurrency rules: committed writable field changes since `base_row_version` that do not overlap the requested fields MUST auto-rebase onto current row state, committed writable field changes that overlap the requested fields MUST fail with `409` and `error.code = same_field_conflict`, and missing or unusable revision history needed to evaluate the stale window MUST fail closed with `409` and `error.code = row_version_conflict`. Non-patch row-version routes and future-base patch requests continue to use `row_version_conflict` for stale or invalid row-version preconditions. The base profile MUST NOT require `Location` for row create and MUST NOT depend on a generic record-read route to make row-create or patch replay deterministic.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-299, AC-367

**REQ-01-071**
Soft-delete of a user-visible first-class record MUST use `DELETE /api/v1/records/{record_id}`. Restore of a currently soft-deleted first-class record MUST use `POST /api/v1/records/{record_id}/restore`. Both routes are record-scoped, not view-scoped. They MUST NOT require `incident_id` or `view_schema_id` in the path or request body, because authorization, history, and affected projections derive from the authoritative `record_id`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-072**
The delete request MUST be JSON and MUST include `base_row_version` and `client_txn_id`. It MAY include optional `reason` bound to `string_contract_id=reason_note_v1`. The restore request MUST be JSON and MUST include `base_row_version` and `client_txn_id`. It MAY include optional `reason` bound to `string_contract_id=reason_note_v1`. For idempotency comparison, omission, explicit JSON `null`, and any supplied `reason` value that normalizes to empty under `reason_note_v1` MUST compare equal. Idempotency for delete and restore MUST be keyed by `(actor_user_id, record_id, client_txn_id)`. If the same authenticated actor replays the same normalized request with the same key, the server MUST return `200 OK` with the originally committed result and MUST create no second mutation. If the same actor reuses that key with a different normalized request, the server MUST fail with `409`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-073**
Because a soft-deleted record no longer appears in ordinary workbook queries, `GET /api/v1/records/{record_id}/history` MUST expose the current tombstone `row_version` so an authorized reviewer can satisfy optimistic concurrency for `POST /api/v1/records/{record_id}/restore` without requiring a separate general record-read route.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-074**
`DELETE /api/v1/records/{record_id}` MUST require current incident role `editor`, `reviewer`, or `admin`. `POST /api/v1/records/{record_id}/restore` MUST require current incident role `reviewer` or `admin`. A caller who lacks visibility to the target record MUST receive `404`. A caller who can see the record but lacks sufficient role MUST receive `403`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-075**
If the current `row_version` differs from `base_row_version`, either route MUST fail with `409` using the common error envelope and `error.code = row_version_conflict`. `error.details` MUST include at least `record_id`, `base_row_version`, and `current_row_version`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-076**
`PATCH /api/v1/records/{record_id}` against a currently soft-deleted record MUST fail with `409` and `error.code = record_deleted_use_restore`. `DELETE /api/v1/records/{record_id}` against an already soft-deleted record MUST fail with `409` and `error.code = record_already_deleted`, except that idempotent replay of the same normalized delete request by the same actor with the same `(record_id, client_txn_id)` MUST return the original success response. `DELETE /api/v1/records/{record_id}` against an otherwise delete-eligible record whose record-type owner defines an active incoming-reference precondition MUST fail with `409`, `error.code = record_delete_blocked`, and `error.details.reason_code` from that owner-defined precondition rather than committing a partial tombstone. `POST /api/v1/records/{record_id}/restore` against a record that is not currently soft-deleted MUST fail with `409` and `error.code = record_not_deleted`, except that idempotent replay of the same normalized restore request by the same actor with the same `(record_id, client_txn_id)` MUST return the original success response.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-077**
`POST /api/v1/records/{record_id}/restore` MUST participate in the base-profile destructive-operation concurrency contract defined by REQ-01-104. For restore, the protected set is exactly the target `record_id`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-078**
On successful delete, the server MUST, in one transaction, set the record envelope soft-delete state, increment `row_version`, create one attributed `change_set`, append the required reversible mutation entry or entries, append a `record_revisions` entry with `operation = soft_delete`, and remove the record from ordinary workbook query surfaces on which it would otherwise materialize. On successful restore, the server MUST, in one transaction, clear the record envelope soft-delete state, increment `row_version`, create one attributed `change_set`, append the required reversible mutation entry or entries, append a `record_revisions` entry with `operation = restore`, and make the record eligible again for ordinary workbook query surfaces whose current view contract would otherwise materialize it. Delete and restore MUST preserve prior history in place. They MUST NOT hard-delete revisions, change sets, or blobs. A restore MUST clear only the current soft-delete state. It MUST NOT accept arbitrary historical row snapshots, target `change_set_id`, or become an implicit substitute for `POST /api/v1/records/{record_id}/rollback`.
Profiles: base, snapshot_reporting
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-233

**REQ-01-079**
A successful delete or restore MUST recompute or invalidate any surviving derived rows whose chips, counts, or linked-record summaries change because of the delete or restore. Projection rows remain derived state, not authority.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-080**
A successful delete or restore MUST return `200 OK` using the common success envelope. `data` MUST include at least `record_id`, `incident_id`, `row_version`, `deleted`, `deleted_at`, `deleted_by_user_id`, and `change_set_id`. On successful restore, `deleted` MUST be `false` and `deleted_at` plus `deleted_by_user_id` MUST be `null`. Because these routes are record-scoped rather than view-scoped, the success response MUST NOT require view-shaped row cells.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-081**
A successful delete MUST emit a replayable `record_changed` event for the affected record. For each affected base `view_schema_id`, the matching `affected_views[]` entry MUST use `change_kind = remove`. A successful restore MUST emit a replayable `record_changed` event for the affected record. For each affected base `view_schema_id`, the matching `affected_views[]` entry MUST use `change_kind = invalidate` rather than introducing a new insert-like change kind in `/ws/v1/`. Any surviving derived row whose chips, counts, or linked-record summaries change because of the delete or restore MUST be refreshed through the ordinary `patch` or `invalidate` mechanisms.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-082**
These routes MUST apply only to first-class record envelopes. They MUST NOT be reused for deleting or restoring individual `record_links`, `record_tags`, `entity_mentions`, `indicator_observations`, or other non-record mutation targets.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231


**REQ-01-083**
Explicit Timeline capture-state transitions to `reviewed` or `superseded` MUST use dedicated record-scoped action routes rather than a direct field patch.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-084**
`POST /api/v1/records/{record_id}/mark-reviewed` MUST:

- require JSON `base_row_version` and `client_txn_id`; it MAY include optional `reason` bound to `string_contract_id=reason_note_v1`,
- apply only to a visible non-deleted `timeline_event` record,
- require current incident role `reviewer` or `admin`,
- succeed only when the current persisted `capture_state` is `rough` or `enriched`,
- in one transaction, increment `row_version`, create one attributed `change_set`, persist `capture_state='reviewed'`, update derived projections, and return `200 OK` with at least `record_id`, `incident_id`, `row_version`, `capture_state`, and `change_set_id`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-085**
If the same authenticated actor replays the same normalized request with the same `(record_id, client_txn_id)`, the server MUST return `200 OK` with the originally committed result and MUST create no second lifecycle transition. A caller who lacks visibility to the target record MUST receive `404`. A caller who can see the record but lacks sufficient role MUST receive `403`. A stale `base_row_version` MUST fail with `409` and `error.code = row_version_conflict`. A request against a non-Timeline record or a Timeline record whose current `capture_state` is not eligible for review MUST fail with `409` and `error.code = illegal_transition`. A request against a currently soft-deleted record MUST fail with `409` and `error.code = record_deleted_use_restore`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-086**
`POST /api/v1/records/{record_id}/supersede` MUST:

- require JSON `base_row_version`, `client_txn_id`, and non-empty `reason` bound to `string_contract_id=reason_note_v1`; it MAY include optional `replacement_record_id`, and when `replacement_record_id` is present it MUST be a non-null stable `record_id`,
- apply only to a visible non-deleted `timeline_event` record,
- require current incident role `reviewer` or `admin`,
- succeed only when the current persisted `capture_state` is `rough`, `enriched`, or `reviewed`,
- when `replacement_record_id` is present, require a different visible non-deleted `timeline_event` record in the same incident whose current `capture_state` is not `superseded`, and require that the target row addressed by `{record_id}` does not already have another active incoming Timeline supersession relation,
- when `replacement_record_id` is present, persist exactly one authoritative non-deleted `record_links` row with `link_type='supersedes'`, `src_record_id=<replacement_record_id>`, and `dst_record_id=<record_id>` in the same committed action,
- in one transaction, increment `row_version`, create one attributed `change_set`, persist `capture_state='superseded'`, update derived projections, and return `200 OK` with at least `record_id`, `incident_id`, `row_version`, `capture_state`, `change_set_id`, `reason`, and `replacement_record_id`, where `replacement_record_id` is `null` when no replacement relation was requested and committed.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-196, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-329, AC-331

**REQ-01-087**
If the same authenticated actor replays the same normalized request with the same `(record_id, client_txn_id)`, the server MUST return `200 OK` with the originally committed result and MUST create no second lifecycle transition. For this route, normalized request comparison MUST include exact `base_row_version`, `reason` after normalization under `reason_note_v1`, and exact `replacement_record_id` when present. Reuse of the same `(record_id, client_txn_id)` with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict` before stale `base_row_version` evaluation. A caller who lacks visibility to the target record MUST receive `404`. A caller who can see the record but lacks sufficient role MUST receive `403`. A stale `base_row_version` MUST fail with `409` and `error.code = row_version_conflict`. A request against an unsupported record type, a Timeline record whose current `capture_state` is already `superseded`, or a `replacement_record_id` that fails the target constraints in REQ-01-086 MUST fail with `409` and `error.code = illegal_transition`. When failure is attributable to Timeline replacement-target validation, `error.details.violated_guards[]` MUST use only `replacement_must_be_different_timeline_record`, `replacement_must_be_visible_active_same_incident_timeline_record`, `replacement_must_not_be_superseded`, and `target_must_not_have_active_replacement` as applicable. A request against a currently soft-deleted record MUST fail with `409` and `error.code = record_deleted_use_restore`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-199, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-329, AC-330

**REQ-01-087A**
`POST /api/v1/records/{record_id}/supersede` MUST branch by authoritative `records.record_type`. For Decision targets, `{record_id}` is the superseded target Decision and `replacement_record_id` is required and is the superseding Decision. The route MUST require a different visible non-deleted Decision in the same incident, require the superseding Decision to be `approved` or `executed`, require the target Decision to be `proposed`, `approved`, or `executed`, and reject the action if either involved Decision is in an inconsistent machine state under Core 02 §10.4.2.1. The committed action MUST persist one active authoritative `record_links` row with `link_type='supersedes'`, `src_record_id=<replacement_record_id>`, and `dst_record_id=<record_id>`, MUST increment row versions and write one attributed `change_set` covering every Decision workbook row that changes, MUST refresh Decision projections, and MUST return the Decision supersession envelope containing `view_schema_id='cartulary.view.decisions.v1'`, `change_set_id`, `target_record_id`, `superseding_record_id`, both row versions, target status, and the normalized `reason`. For `proposed` or `approved` targets the persisted target status becomes `superseded`; for `executed` targets the persisted target status remains `executed` and the Decision view MUST surface `decision.is_superseded=true`. The route MUST NOT treat `reason` as a generalized approval field.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-199, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-329, AC-330

**REQ-01-088**
A successful Timeline action MUST emit the ordinary replayable `record_changed` event for the affected Timeline record. A successful Decision supersession action MUST emit ordinary replayable `record_changed` events for every Decision workbook row changed by the action.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

#### 3.3.5.0 Destructive-operation concurrency and rollback contract


**Table 3.3.5.0-A. Destructive-operation route table**

| Route | Required request members | Protected-set rule | Success summary |
| --- | --- | --- | --- |
| `POST /api/v1/records/{record_id}/restore` | `base_row_version`, `client_txn_id`; optional `reason` | Protected set is exactly the target `record_id` | `200 OK` with record-scoped restore summary |
| `POST /api/v1/records/{record_id}/rollback` | `base_row_version`, `client_txn_id`, `target`; optional `reason` | Protected set is every first-class `record_id` whose authoritative state would be mutated by the selected rollback target against current state | `200 OK` with rollback summary including `target`, `rollback_change_set_id`, and canonical `affected_record_ids[]` |
| `POST /api/v1/records/{survivor_record_id}/merge` | `loser_record_id`, `survivor_base_row_version`, `loser_base_row_version`, `client_txn_id`; optional `reason` | Protected set is the survivor, the loser, and every additional first-class `record_id` mutated directly by repointing or deterministic recreation | `200 OK` with merge summary and carried-forward identifier results |

**Table 3.3.5.0-B. Rollback target and whole-row-restore addressing**

| `target.kind` | Required selector | Scope of reversal |
| --- | --- | --- |
| `history_entry` | `history_entry_ref` | One single-entry reversible mutation target exposed by `/history` |
| `change_set` | `change_set_id` | Every reversible mutation entry in the addressed `change_set`, in reverse deterministic entry order |
| `row_restore` | `restore_to_revision_no` | Only the authoritative row-backed fields of the addressed `record_id`; non-row-backed links, tags, mentions, observations, and evidence associations are not implicitly recreated or deleted |

**Table 3.3.5.0-C. Destructive-operation error summary**

| Condition | Transport | `error.code` | Required detail |
| --- | --- | --- | --- |
| Required destructive-operation lock unavailable | `409` | `record_locked` | `error.retryable=true` |
| Current row version differs from supplied base version | `409` | `row_version_conflict` | Current and base version details for the addressed route |
| Rollback target does not resolve to a visible history target | `404` | `rollback_target_not_found` | Target selector details as available |
| Rollback target exists but cannot be reversed safely against current state | `409` | `rollback_precondition_failed` | Family `reason_code` from §3.3.6.2 |
| Merge precondition other than row-version freshness fails | `409` | `merge_precondition_failed` | Family `reason_code` from §3.3.6.2 |
| Restore against a record that is not currently soft-deleted | `409` | `record_not_deleted` | Record-scoped restore misuse |


**REQ-01-089**
`POST /api/v1/records/{record_id}/rollback` MUST execute a reviewer-visible history reversal anchored to the row-centric history of `record_id`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-090**
The route is record-scoped, not view-scoped. `record_id` identifies the row whose history surface the caller is acting from. The request body MUST be JSON. It MUST include:

- `base_row_version`,
- `client_txn_id`,
- `target`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

It MAY include optional `reason` bound to `string_contract_id=reason_note_v1`.

**REQ-01-091**
Omission, explicit JSON `null`, and any supplied `reason` value that normalizes to empty under `reason_note_v1` MUST compare equal for idempotency.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-092**
`target.kind` MUST use this closed vocabulary:

- `history_entry`,
- `change_set`,
- `row_restore`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-093**
A request `target` MUST use one of these exact shapes:

- `history_entry`: `{ "kind": "history_entry", "history_entry_ref": "<opaque string>" }`
- `change_set`: `{ "kind": "change_set", "change_set_id": "<stable identifier>" }`
- `row_restore`: `{ "kind": "row_restore", "restore_to_revision_no": <integer> }`
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-094**
Unknown top-level request members, unknown `target` members, a missing required selector for the chosen `kind`, or a selector whose JSON type does not match the declared shape MUST fail with `400` and `error.code = invalid_rollback_request`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-095**
Successful responses MAY include additive `target` members in later compatible revisions. Clients MUST ignore additive response members they do not use.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-096**
Single-entry rollback through `history_entry_ref` MUST be available only when the selected history item both has a valid `history_entry_ref` under REQ-01-054 and is currently reversible under the rollback-precondition rules of this subsection.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231, AC-384

**REQ-01-097**
Whole-`change_set` rollback through `change_set_id` MUST reverse every reversible mutation entry in that `change_set` in reverse deterministic entry order.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-098**
Whole-row restore through `restore_to_revision_no` MUST restore only the authoritative row-backed fields of `record_id` to the selected revision snapshot. It MUST NOT implicitly recreate or delete `record_links`, `record_tags`, `entity_mentions`, `indicator_observations`, or evidence associations that are not row-backed fields.
Profiles: base, snapshot_reporting
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231, AC-233

**REQ-01-099**
`POST /api/v1/records/{record_id}/rollback` MUST require current incident role `reviewer` or `admin`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-100**
A caller who lacks visibility to `record_id` MUST receive `404`. A caller who can see `record_id` but lacks sufficient role MUST receive `403`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-101**
A caller who targets a currently soft-deleted record MUST receive `409` with `error.code = record_deleted_use_restore`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-102**
Idempotency for rollback MUST be keyed by `(actor_user_id, record_id, client_txn_id)`. If the same authenticated actor replays the same normalized request with the same key, the server MUST return `200 OK` with the originally committed result and MUST create no second rollback `change_set`. If the same actor reuses that key with a different normalized request, the server MUST fail with `409`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-103**
If the current `row_version` differs from the supplied `base_row_version`, the server MUST fail with `409` and `error.code = row_version_conflict`. `error.details` MUST include at least:

- `record_id`,
- `base_row_version`,
- `current_row_version`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-104**
The base-profile destructive-operation family is exactly `POST /api/v1/records/{record_id}/restore`, `POST /api/v1/records/{record_id}/rollback`, and `POST /api/v1/records/{survivor_record_id}/merge`. Ordinary `PATCH`, ordinary `DELETE`, and ordinary soft-delete are outside this family, MUST remain on the ordinary optimistic-concurrency path, and MUST NOT acquire destructive-operation locks in the current profile. The server MUST use short-lived internal destructive-operation locks for every request in that family. The public contract includes no client-visible lock-acquire route, no lock-holder identity surface, no manual unlock route, and no server-side queueing of destructive requests. The server MUST attempt lock acquisition fail-fast and MUST NOT wait for a held lock to clear and then continue evaluating the same request. Lock acquisition MUST use canonical ascending `record_id` order across the full protected set. If any required lock is unavailable, the route MUST fail immediately with `409`, `error.code = record_locked`, and `error.retryable = true`. Only after the full protected set is acquired MAY the server re-read authoritative current state and evaluate `row_version_conflict`, restore eligibility such as `record_not_deleted`, `rollback_precondition_failed`, or `merge_precondition_failed`. Locks MUST be released on commit, rollback, or request termination. For restore, the protected set is exactly the target `record_id`. For rollback, the protected set is every first-class `record_id` whose authoritative state would be mutated if the selected rollback target were applied against current authoritative state. For merge, the protected set is the survivor record, the loser record, and every additional first-class `record_id` whose authoritative state the merge transaction mutates directly through repointing or deterministic recreation.
Profiles: base
Verified by: AC-182, AC-187, AC-215, AC-216, AC-217, AC-218, AC-231, AC-353

**REQ-01-105**
If the supplied `target` does not resolve to a visible rollback target in `GET /api/v1/records/{record_id}/history`, the server MUST fail with `404` and `error.code = rollback_target_not_found`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-106**
If the target exists but cannot be safely reversed against current authoritative state, the server MUST fail with `409` and `error.code = rollback_precondition_failed`. `error.details.reason_code` MUST use the rollback-precondition registry in §3.3.6.2.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-107**
A single-entry rollback MUST succeed when later committed changes are unrelated to the same mutation target. It MUST fail with `rollback_precondition_failed` when later committed changes touched the same mutation target or otherwise make isolated reversal ambiguous.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-108**
Reversal of an erroneous merge MUST use `target.kind = 'change_set'` against the merge `change_set_id`. The base profile defines no separate unmerge route.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-109**
On success, the server MUST, in one transaction:

- validate authorization and rollback preconditions,
- create one attributed `change_set` with `source = 'rollback'`,
- append ordered inverse mutation entries sufficient to reconstruct the reversal,
- increment `row_version` on every changed first-class record,
- append `record_revisions` entries with `operation = rollback` for every affected first-class record,
- update or invalidate affected projections before commit returns.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-110**
A successful response MUST return `200 OK` using the common success envelope. `data` MUST include at least:

- `incident_id`,
- `record_id`,
- `row_version`,
- `target`, using the same `kind` vocabulary and selector members accepted by the request,
- `target_change_set_id` when known,
- `rollback_change_set_id`,
- `affected_record_ids[]`, serialized in canonical ascending `record_id` order.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-01-111**
The initiating client and all subscribers MUST observe the rollback through the ordinary replayable `record_changed` stream for each affected record. The response itself MUST NOT require view-shaped row cells or introduce a special rollback-only event family.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

#### 3.3.5.1 Deployment-local user-account and incident-membership administration contracts

**REQ-01-112**
The base profile MUST expose two distinct mutable administrative route families:

- `/api/v1/users` for deployment-local internal user accounts,
- `/api/v1/incidents/{incident_id}/memberships` for incident-scoped membership assignments.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-113**
The user-account route family MUST manage stable internal `user_id` resources. Incident memberships, audit attribution, saved-view ownership, workbook-preference ownership, and provider-backed identities MUST bind to `user_id`, not to email address, visible labels, or provider subject.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-114**
The public user-account routes MUST expose only safe user-resource fields. They MUST NOT expose password hashes, TOTP secrets, WebAuthn credential material, opaque session tokens, provider assertions, or equivalent secret-bearing authentication state.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-115**
The user resource MUST expose, at minimum:

- `user_id`,
- `email`,
- `display_name`,
- `is_active`,
- `mfa_required`,
- `is_deployment_admin`,
- `created_at`,
- `updated_at`,
- nullable `updated_by_user_id`,
- nullable `last_login_at`,
- `user_version`,
- `auth_bindings[]`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-116**
`auth_bindings[]` MUST be read-only, MUST always be present, MUST be non-empty in the current profile, MUST contain only active safe binding summaries, and MUST be serialized in canonical order by `provider_type asc`, then `provider_key asc`, then `created_at asc`. The current profile defines `auth_bindings[]` as a closed tagged union with exactly one local binding summary and, when the Enterprise Authentication Extension Profile is implemented, zero or more enterprise binding summaries. The local binding summary MUST contain exactly `provider_type`, `provider_key`, `username`, and `created_at`; it MUST use `provider_type='local'` and `provider_key='local'`; `username` MUST equal the same authoritative `email` exposed on the safe user resource; `created_at` MUST equal the same authoritative `created_at` exposed on the safe user resource; and it MUST NOT expose `auth_binding_id`, `provider_subject`, or `last_auth_at`. If the Enterprise Authentication Extension Profile is implemented, each active enterprise binding summary with `provider_type='oidc'` or `provider_type='saml'` MUST contain exactly `auth_binding_id`, `provider_key`, `provider_type`, `provider_subject`, `created_at`, and `last_auth_at`; `last_auth_at` MAY be `null`, but the member itself MUST be present; and enterprise binding summaries MUST NOT expose `username`. No additional members are allowed on either summary variant in the current profile. In the current profile, this canonical ordering places the single local binding summary before any enterprise binding summaries. Retired enterprise bindings MUST NOT appear in `auth_bindings[]`. The base profile MUST NOT require clients to interpret `auth_bindings[]` in order to authorize incident data access.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231, AC-348, AC-351

**REQ-01-117**
`GET /api/v1/users` and `GET /api/v1/users/{user_id}` MUST fail closed unless the caller has the deployment-scoped account-administration capability defined by Core 04 §2. `GET /api/v1/users` MUST return safe user resources ordered by `user_id asc`, MUST use the common success envelope with `data.users[]` plus `meta.paging`, and MUST accept only `limit` and `cursor_token` under §3.3.7. Pagination failures on this route MUST fail with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code` from the `invalid_pagination_request` registry in §3.3.6.2. Inert imported historical actors from incident portability are not login-capable user resources and MUST NOT be returned by this route family unless they have been explicitly mapped to a local user account.
Profiles: base
Verified by: AC-127, AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-118**
`GET /api/v1/users/{user_id}` MUST return the common success envelope with `data` equal to the requested safe user resource. Because this route is singleton, it MUST reject `limit`, `cursor_token`, and pagination aliases with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code=pagination_not_supported`.
Profiles: base
Verified by: AC-127, AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-119**
`POST /api/v1/users` MUST require the same deployment-scoped account-administration capability. The request body MUST be a JSON object. The route MUST accept:

- required `client_txn_id`,
- required `auth_kind`,
- required `email`,
- required `display_name`,
- required `initial_password`,
- optional `mfa_required`,
- optional `is_deployment_admin`.

No other top-level request members are allowed. The base profile MUST NOT accept a separate local `username` create member. `display_name` is required and is bound to `string_contract_id=display_name_line_v1`. `initial_password` is required and is bound to `string_contract_id=local_password_provision_v1`. `mfa_required` and `is_deployment_admin` MUST be boolean when supplied and MUST NOT be `null`. If `mfa_required` is omitted, the server MUST default it to `true`. If `is_deployment_admin` is omitted, the server MUST default it to `false`. For `auth_kind='local'`, the created `email` becomes the only base-profile local login identifier. A missing `initial_password`, explicit JSON `null`, non-string `initial_password`, or any `initial_password` value that fails `local_password_provision_v1` MUST fail with `400` and `error.code = invalid_mutation_payload`. When the failure is attributable to `initial_password`, `error.details.field` MUST equal `initial_password`. The create contract MUST NOT accept client-supplied `is_active`, `user_id`, `created_at`, `updated_at`, `updated_by_user_id`, `last_login_at`, `user_version`, or `auth_bindings[]`; any such member or any other unknown top-level member MUST fail with `400` and `error.code = invalid_mutation_payload`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231, AC-312

**REQ-01-120**
In the base profile, `auth_kind` MUST be `local`. `email` is required and is bound to `string_contract_id=email_address_v1`. A supplied `email` value that normalizes to authoritative `null` under `email_address_v1`, or otherwise fails that contract, MUST fail create-time validation. Deployment uniqueness for local users MUST be enforced on the deterministic comparison form produced by `email_address_v1`. For `auth_kind='local'`, the created `email` becomes the only base-profile local login identifier. `display_name` MUST satisfy `display_name_line_v1` before create-time idempotency comparison or persistence. `initial_password` MUST satisfy `local_password_provision_v1` before create-time idempotency comparison or any password-hash derivation. For `auth_kind='local'`, the server MUST encode the validated `initial_password` as UTF-8 without BOM, derive `password_hash` with Argon2id from those exact bytes, persist only `password_hash`, and discard the cleartext secret after request processing. On successful create, the created user resource MUST initialize `is_active=true`. The public create contract MUST NOT permit the client to choose any different initial `is_active` state. The server MUST NOT expose `initial_password` or any equivalent secret in a response or event payload.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231, AC-312

**REQ-01-121**
The first deployment admin MUST be created only through the deployment-local bootstrap-admin manifest contract defined by REQ-01-530..REQ-01-536. The public API MUST NOT allow unauthenticated bootstrap of a deployment-admin user.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231, AC-343, AC-344, AC-345, AC-346, AC-347

**REQ-01-530**
The current profile defines exactly one conformant first-deployment-admin bootstrap mechanism: consumption of a deployment-local UTF-8 JSON manifest during startup. That manifest MUST declare `bootstrap_schema_id = "cartulary.bootstrap_admin.v1"`. Helper tooling such as an installer, CLI, Helm chart, or package script MAY materialize that manifest, but the running application MUST NOT treat those helpers as separate bootstrap mechanisms.
Profiles: base
Verified by: AC-343, AC-344, AC-345, AC-346

**REQ-01-531**
The bootstrap artifact MUST be a UTF-8 JSON file whose top-level value is one object with exactly these members:

- required `bootstrap_schema_id`,
- required `bootstrap_artifact_id`,
- required `email`,
- required `display_name`,
- required `initial_password`,
- optional `mfa_required`.
Profiles: base
Verified by: AC-343, AC-344

**REQ-01-532**
`bootstrap_schema_id` MUST equal `cartulary.bootstrap_admin.v1`. `bootstrap_artifact_id` MUST be a UUID. `email` MUST satisfy `email_address_v1`. `display_name` MUST satisfy `display_name_line_v1`. `initial_password` MUST satisfy `local_password_provision_v1`. If `mfa_required` is omitted, the bootstrap consumer MUST behave as though `true` were supplied. If `mfa_required` is supplied, it MUST be `true`; explicit `false` is invalid. Unknown top-level members are invalid. The manifest MUST NOT accept `user_id`, `auth_kind`, `is_active`, `is_deployment_admin`, incident memberships, provider bindings, TOTP seed material, session state, or any other extension field in the current profile.
Profiles: base
Verified by: AC-343, AC-344

**REQ-01-533**
After deployment-configuration validation and before any HTTP listener, WebSocket listener, or background-job runner starts, the implementation MUST query both deployment-local user state and deployment-local bootstrap-completion state. If at least one active `deployment_admin` exists, bootstrap manifest consumption MUST be skipped for that startup. If zero active `deployment_admin` users exist and no bootstrap-completion marker exists, the implementation MUST attempt bootstrap from the configured manifest path defined by Core 04 §12.3.2. If zero active `deployment_admin` users exist and a bootstrap-completion marker already exists, the current profile does not define re-bootstrap and startup MUST fail closed.
Profiles: base
Verified by: AC-343, AC-344, AC-345, AC-346

**REQ-01-534**
A valid bootstrap manifest MUST create exactly one local user in one transaction with `email` and `display_name` from the manifest, `password_hash` derived from `initial_password`, `mfa_required=true`, `is_active=true`, and `is_deployment_admin=true`. The same transaction MUST create no incident membership, no incident-scoped authorization, and no provider binding, and MUST also persist one deployment-local bootstrap-completion marker plus one deployment-local administrative audit event.
Profiles: base
Verified by: AC-343, AC-344

**REQ-01-535**
If the normalized manifest `email` already exists on any local user row, bootstrap MUST fail closed. It MUST NOT silently promote, mutate, or reuse an existing user. One-time semantics MUST be driven by the persisted bootstrap-completion marker, not by deleting, renaming, or mutating the manifest file on disk. A conformant deployment MAY therefore use a read-only secret mount for the manifest file.
Profiles: base
Verified by: AC-344, AC-345, AC-346

**REQ-01-536**
A user created by successful bootstrap enters the ordinary local credential lifecycle defined by §3.3.2.2. Because that user begins with `mfa_required=true` and no active TOTP factor, the first valid password login MUST follow the existing `mfa_setup_required` -> `bootstrap_token` -> `POST /api/v1/auth/mfa/totp/begin` -> `POST /api/v1/auth/mfa/totp/complete` flow.
Profiles: base
Verified by: AC-343, AC-347

**REQ-01-122**
Idempotency for user create MUST be keyed by `(actor_user_id, client_txn_id)`. For normalized request comparison, `email` MUST compare after `email_address_v1` normalization, omitted `mfa_required` MUST compare equal to explicit `true`, omitted `is_deployment_admin` MUST compare equal to explicit `false`, `display_name` MUST compare after `display_name_line_v1` normalization, and `initial_password` MUST compare only after `local_password_provision_v1` validation using exact post-JSON-decoding code-point equality with no trimming, Unicode NFC normalization, case-folding, or line-ending normalization. Requests rejected under REQ-01-119 are never part of normalized comparison. If the same authenticated actor replays the same normalized request with the same `client_txn_id`, the server MUST return `200 OK` with the originally created safe user resource and MUST create no second user. If the same actor reuses `client_txn_id` with a different normalized request, including a different validated `initial_password`, the server MUST fail with `409` and `error.code = client_txn_conflict`; `error.details` MUST include at least `client_txn_id`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231, AC-312

**REQ-01-123**
A first-time successful `POST /api/v1/users` call MUST return `201 Created` and the common success envelope with `data` equal to the created safe user resource.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-124**
`PATCH /api/v1/users/{user_id}` MUST require the same deployment-scoped account-administration capability. The route MUST accept `base_user_version` plus changed mutable fields only. In the base profile, mutable fields are `email`, `display_name`, `is_active`, `mfa_required`, and `is_deployment_admin`. Password reset, TOTP reset, and session revoke-all are route-owned actions and MUST NOT be expressed through this patch route. When supplied, `email` MUST satisfy `email_address_v1`, and deployment uniqueness for local users MUST continue to be enforced on the deterministic comparison form produced by that contract. When supplied, `display_name` MUST satisfy `display_name_line_v1`. The route MUST reject attempted mutation of `user_id`, `created_at`, `last_login_at`, or `auth_bindings[]`. For a local account, a successful committed change to `email` MUST change the authoritative local login identifier atomically in the same commit. After that commit, authentication MUST succeed with the new email-form `username` and MUST fail with the prior email-form `username`, unless a later profile explicitly defines login aliases. The base profile defines no such aliases and no second persisted local username namespace. If the current `user_version` differs from `base_user_version`, the server MUST fail with `409` and `error.code = user_version_conflict`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231, AC-312

**REQ-01-125**
Changing `is_active` from `true` to `false` MUST revoke all active sessions for that user immediately and MUST NOT delete that user's incident memberships. A patch that would deactivate or demote the last active `is_deployment_admin=true` user MUST fail with `409` and `error.code = last_deployment_admin`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-126**
A successful `PATCH /api/v1/users/{user_id}` call MUST return the common success envelope with `data` equal to the resulting safe user resource.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-527**
`POST /api/v1/users/{user_id}/password/reset` MUST require the same deployment-scoped account-administration capability. The request body MUST be a JSON object and MUST accept required `base_user_version`, required `client_txn_id`, required `new_password`, and optional `reason` bound to `string_contract_id=reason_note_v1`. `new_password` is bound to `string_contract_id=local_password_provision_v1`. On success the server MUST update `password_hash`, stamp `password.changed_at`, invalidate any pending TOTP enrollment or bootstrap token for that user, revoke all active sessions for that user, preserve any currently active TOTP credential state, increment `user_version`, and return the common success envelope with `data` equal to the resulting safe user resource. The route MUST use `client_txn_id` as route-scoped idempotency key within `(actor_user_id, user_id, client_txn_id)`. Normalized replay comparison for this route MUST include exact `base_user_version`, exact `new_password`, and normalized `reason`, with omitted `reason`, explicit JSON `null`, and any `reason` value that normalizes to empty under `reason_note_v1` comparing equal. Exact replay of a previously committed success MUST return `200 OK` with the original committed safe user resource before fresh `user_version_conflict` or target-state evaluation runs. Reuse of the same route-scoped key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. Any deployment-local idempotency substrate for this route MUST NOT retain cleartext `new_password`.
Profiles: base
Verified by: AC-340, AC-341

**REQ-01-528**
`POST /api/v1/users/{user_id}/mfa/totp/reset` MUST require the same deployment-scoped account-administration capability. The request body MUST be a JSON object and MUST accept required `base_user_version`, required `client_txn_id`, and optional `reason` bound to `string_contract_id=reason_note_v1`. On success the server MUST clear active and pending TOTP setup state for that user, revoke all active sessions for that user, preserve `mfa_required`, increment `user_version`, and return the common success envelope with `data` equal to the resulting safe user resource. The route MUST use `client_txn_id` as a route-scoped idempotency key within `(actor_user_id, user_id, client_txn_id)`. Normalized replay comparison for this route MUST include exact `base_user_version` and normalized `reason`, with omitted `reason`, explicit JSON `null`, and any `reason` value that normalizes to empty under `reason_note_v1` comparing equal. Exact replay of a previously committed success MUST return `200 OK` with the original committed safe user resource before fresh `user_version_conflict` or target-state evaluation runs. Reuse of the same route-scoped key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. After success, the next valid password login for that user MUST behave as follows: when `mfa_required=true`, `POST /api/v1/auth/login` returns `401 error.code='mfa_setup_required'`, includes `error.details.required_setup_kinds=["totp"]`, includes one short-lived `bootstrap_token` plus `bootstrap_expires_at`, and sets no session cookie; when `mfa_required=false`, login proceeds without requiring `second_factor` until a new factor is enrolled.
Profiles: base
Verified by: AC-341

**REQ-01-529**
`POST /api/v1/users/{user_id}/sessions/revoke-all` MUST require the same deployment-scoped account-administration capability. The request body MUST be a JSON object and MUST accept required `client_txn_id` and optional `reason` bound to `string_contract_id=reason_note_v1`. On success the server MUST revoke all active sessions for that user without changing `password_hash`, active TOTP credential state, pending TOTP enrollment state, or `mfa_required`, and MUST return the common success envelope with `data` containing at least `user_id`, `sessions_revoked=true`, and `revoked_at`. The route MUST use `client_txn_id` as a route-scoped idempotency key within `(actor_user_id, user_id, client_txn_id)`. Normalized replay comparison for this route MUST include normalized `reason`, with omitted `reason`, explicit JSON `null`, and any `reason` value that normalizes to empty under `reason_note_v1` comparing equal. Exact replay of a previously committed success MUST return `200 OK` with the original committed success payload before fresh session-state evaluation runs. Reuse of the same route-scoped key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. Incident membership, incident role, or incident-admin status alone MUST NOT authorize any of the three action routes in this subsection.
Profiles: base
Verified by: AC-342

**REQ-01-127**
The membership route family MUST persist and expose stable incident membership resources. A membership resource MUST expose, at minimum:

- `incident_id`,
- `user_id`,
- `display_name`,
- `role`,
- `joined_at`,
- `added_by_user_id`,
- `updated_at`,
- `updated_by_user_id`,
- `membership_version`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-128**
`display_name` on the membership resource is a read-only joined user summary. Membership state itself MUST remain keyed by `(incident_id, user_id)`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-129**
`GET /api/v1/incidents/{incident_id}/memberships` MUST require current membership in that incident, MUST return the common success envelope with `data.memberships[]` plus `meta.paging` ordered by `joined_at asc, user_id asc`, and MUST accept only `limit` and `cursor_token` under §3.3.7. Pagination failures on this route MUST fail with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code` from the `invalid_pagination_request` registry in §3.3.6.2.
Profiles: base
Verified by: AC-127, AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-130**
`POST /api/v1/incidents/{incident_id}/memberships` MUST require the caller's current role on that incident to be `admin`. The route MUST accept:

- required `client_txn_id`,
- exactly one of `user_id` or `email`,
- required `role`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-131**
`role` MUST use the closed vocabulary `viewer`, `editor`, `reviewer`, and `admin`. If `email` is supplied, the server MUST resolve it through the same `email_address_v1` normalization and comparison substrate used by local login and user create or update, and MUST bind stored membership state to the resolved `user_id`. The base profile MUST NOT auto-create or invite a user from this route. If the target user does not exist, the server MUST fail with `404` and `error.code = user_not_found`. If the target user exists but `is_active=false`, the server MUST fail with `409` and `error.code = user_inactive`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-132**
Idempotency for membership create MUST be keyed by `(actor_user_id, incident_id, client_txn_id)`. If no current membership exists for the resolved `(incident_id, user_id)`, the server MUST create one, MUST return `201 Created`, and MUST return the common success envelope with `data` equal to the created membership resource. If a current membership already exists with the same role, the server MUST return `200 OK` with the existing membership resource and MUST create no second membership row. If a current membership already exists with a different role, the server MUST fail with `409` and `error.code = membership_exists_use_patch`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-133**
`PATCH /api/v1/incidents/{incident_id}/memberships/{user_id}` MUST require the caller's current role on that incident to be `admin`. The route MUST accept:

- required `base_membership_version`,
- required `role`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-134**
The route MUST reject attempted mutation of `incident_id`, `user_id`, `joined_at`, or `added_by_user_id`. If the current `membership_version` differs from `base_membership_version`, the server MUST fail with `409` and `error.code = membership_version_conflict`. If the requested `role` already matches the current role, the server MUST return `200 OK` with the current membership resource and MUST NOT increment `membership_version`. A successful role change MUST take effect on the next incident-scoped authorization check for that user.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-135**
A successful `PATCH /api/v1/incidents/{incident_id}/memberships/{user_id}` call MUST return the common success envelope with `data` equal to the resulting membership resource.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-136**
`DELETE /api/v1/incidents/{incident_id}/memberships/{user_id}` MUST require the caller's current role on that incident to be `admin`. If no current membership exists for that `(incident_id, user_id)`, the server MUST fail with `404` and `error.code = membership_not_found`. A successful delete MUST remove only that incident membership and MUST return `204 No Content`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-137**
A membership create, role change, or delete that would leave the incident without any `admin` membership MUST fail with `409` and `error.code = last_incident_admin`. Self-demotion or self-removal by an incident admin MAY succeed only when another current `admin` membership remains.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

#### 3.3.5.2 View-schema, saved-view, and workbook-preference contracts

The public `view_schema` discovery-route shape is owned by REQ-01-288 in §6. This subsection owns saved-view and workbook-preference resources that bind to those schemas.

**REQ-01-138**
The saved-view route family MUST expose a saved-view resource containing, at minimum:

- `saved_view_id`,
- `incident_id`,
- `view_schema_id`,
- `scope`,
- `display_name`,
- `query_json`,
- `layout_json`,
- `owner_user_id`,
- `created_at`,
- `updated_at`,
- `saved_view_version`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-139**
`scope` MUST use the closed vocabulary `private`, `shared`, and `system`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-140**
`owner_user_id` MUST be present for `private` and `shared` saved views. It MAY be null only for `system` saved views.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-141**
`saved_view_version` MUST be monotonically increasing per `saved_view_id`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-142**
`query_json` MUST store only persisted saved-view query state and MUST use the closed top-level grammar `{ "sort": [...], "filters": [...], "group_by": "..." }`, where `group_by` is optional and MUST be omitted when grouping is inactive. Persisted `sort` and `filters` MUST always be present arrays. `[]` is the only canonical persisted representation of no user sort override or no filters. Persisted `group_by` MUST be omitted when grouping is inactive and MUST NOT be serialized as JSON `null`. No other top-level members are allowed. Each `sort[]` entry MUST use the exact shape defined in §3.3.4 and preserve declared order. Each `filters[]` entry MUST use the exact filter predicate wire shape defined in §3.3.4.1, persist the canonical normalized `arg` form defined there, and persist canonical `filters[]` ordering by `field_key asc`. Duplicate normalized `field_key` entries in either `sort[]` or `filters[]` are invalid. `sort[].field_key` MUST belong to the owning schema's `sort_fields`. `filters[].field_key` MUST belong to the union of the owning schema's `filter_fields` and `synthetic_filter_predicates[].field_key`. `group_by` MUST belong to the owning schema's `grouping_fields`. `record_id` and `row_version` MUST NOT appear anywhere in `query_json`. `query_json` MUST NOT use visible tab labels, visible column labels, presentation-only group-header text, table names, or storage-specific identifiers. Persisted `query_json.sort` MUST store only the normalized user sort override list rather than the effective default-extended sort tuple.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231, AC-360

**REQ-01-143**
`layout_json` MUST store only shared portable layout state and MUST use the closed versioned grammar:

```json
{
  "layout_schema_id": "cartulary.layout.v1",
  "column_order": ["timeline.occurred_at", "timeline.summary"],
  "hidden_field_keys": ["timeline.details"],
  "column_widths": [
    { "field_key": "timeline.summary", "width_px": 420 }
  ]
}
```

For `layout_json`:

- `layout_schema_id` MUST be present and MUST equal `cartulary.layout.v1`,
- `column_order` MUST be present and MUST be a full permutation of the active schema's non-technical `fields[].field_key` values, with each key appearing exactly once,
- `hidden_field_keys` MUST be present, MUST be unique, MUST use canonical ascending order, and MUST be a subset of `column_order`,
- `column_widths` MUST be present and MUST be a unique sparse list ordered by `field_key asc`,
- each `column_widths[]` entry MUST use exactly `field_key` and `width_px`,
- `width_px` MUST be an integer in the inclusive range `40..4096`,
- `record_id` and `row_version` MUST NOT appear in `column_order`, `hidden_field_keys`, or `column_widths`,
- unknown top-level members and unknown nested members are invalid,
- `layout_json` MUST NOT store selection, scroll position, focused cell, transient popover state, open inspector state, preview state, presence, or other per-session or per-device client state,
- `layout_json` MUST NOT be the authority for `saved_view_id`, `incident_id`, `view_schema_id`, `scope`, ownership, authorization, or startup/default surface selection.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-144**
`GET /api/v1/incidents/{incident_id}/saved-views` MUST return only the saved-view resources visible to the caller, MUST use the common success envelope with `data.saved_views[]` plus `meta.paging`, MUST order results by `updated_at desc, saved_view_id asc`, and MUST accept only `limit` and `cursor_token` under §3.3.7. This route is authoritative only for saved-view resources. Required base-profile surfaces in the authoritative `view_schema` registry are not discovered by this route unless a distinct saved-view object also exists for them. For clarity, absence of `cartulary.view.task_requests.v1` or `cartulary.view.decisions.v1` from this route MUST NOT be interpreted as absence of the required base surfaces themselves; those surfaces remain discoverable through the authoritative `view_schema` registry and directly addressable by `sheet_ref.kind='view_schema'`. Pagination failures on this route MUST fail with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code` from the `invalid_pagination_request` registry in §3.3.6.2.
Profiles: base
Verified by: AC-127, AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-145**
`POST /api/v1/incidents/{incident_id}/saved-views` MUST accept a JSON object containing required `view_schema_id`, required `display_name`, required `query_json`, optional `layout_json`, and optional `scope`. `display_name` MUST be non-null, MUST satisfy `display_name_line_v1`, and MUST be compared after that normalization for request-time equality checks. `query_json` MUST be non-null. If `scope` is omitted, the server MUST treat it as `private`. If `layout_json` is omitted or supplied as `{}`, the server MUST normalize it to the canonical schema-derived default `cartulary.layout.v1` object for `view_schema_id`. If `layout_json` is supplied and non-empty, it MUST be non-null and MUST satisfy REQ-01-143. `query_json` MUST be validated and normalized using the same sort, filter, and grouping rules as `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`; within `query_json`, omission of `sort` MUST normalize to `sort=[]`, omission of `filters` MUST normalize to `filters=[]`, and explicit `group_by=null` is invalid. During that validation, a raw parsed `query_json.sort` length greater than `8` or a raw parsed `query_json.filters` length greater than `16` MUST fail with `400`, `error.code = invalid_mutation_payload`, `error.details.reason_code` equal to `sort_count_exceeded` or `filter_count_exceeded`, `error.details.field` equal to `query_json.sort` or `query_json.filters`, `error.details.requested_count = <raw count>`, and `error.details.max_count` equal to `8` or `16` as applicable. The server MUST NOT truncate or partially honor an oversize saved-view query array. No other top-level request members are allowed. The ordinary public create route MUST reject `scope='system'`, any supplied `null` for `display_name`, `query_json`, `layout_json`, or `scope`, any `query_json` or `layout_json` field reference not declared by the addressed `view_schema_id`, any use of `record_id` or `row_version` inside `query_json` or `layout_json`, and any unknown top-level member with `400` and `error.code = invalid_mutation_payload`. When the failure is attributable to one member path, `error.details.field` MUST identify that path, including paths such as `query_json.sort[0].field_key`, `query_json.filters[1].field_key`, and `layout_json.column_widths[0].field_key`. The created saved-view resource MUST expose the normalized `query_json` and canonical `layout_json`; persisted `layout_json` MUST NOT remain `{}` after a conformant write.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231, AC-360

**REQ-01-146**
`PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` MUST accept a JSON object containing required `base_saved_view_version` plus zero or more changed mutable fields only. Mutable fields are `display_name`, `query_json`, `layout_json`, and, when permitted by scope rules, `scope`. Omission of a mutable field means unchanged. `display_name`, `query_json`, and `layout_json` MUST be non-null when supplied, and `display_name` MUST satisfy `display_name_line_v1`. When `query_json` is supplied, the server MUST validate and normalize it using REQ-01-142; within `query_json`, omission of `sort` MUST normalize to `sort=[]`, omission of `filters` MUST normalize to `filters=[]`, and explicit `group_by=null` is invalid. During that validation, a raw parsed `query_json.sort` length greater than `8` or a raw parsed `query_json.filters` length greater than `16` MUST fail with `400`, `error.code = invalid_mutation_payload`, `error.details.reason_code` equal to `sort_count_exceeded` or `filter_count_exceeded`, `error.details.field` equal to `query_json.sort` or `query_json.filters`, `error.details.requested_count = <raw count>`, and `error.details.max_count` equal to `8` or `16` as applicable. The server MUST NOT truncate or partially honor an oversize saved-view query array. When `layout_json` is supplied, the server MUST treat `{}` as a legacy-equivalent request for the canonical schema-derived default layout for that `view_schema_id` and otherwise MUST validate and normalize it using REQ-01-143. No top-level request members other than `base_saved_view_version` and those mutable fields are allowed. It MUST reject attempted mutation of `incident_id`, `saved_view_id`, or `view_schema_id`. Unknown or forbidden top-level members, unknown nested members in `query_json` or `layout_json`, invalid `query_json` or `layout_json` field references, any use of `record_id` or `row_version` inside `query_json` or `layout_json`, or explicit `null` for a non-null supplied mutable member MUST fail with `400` and `error.code = invalid_mutation_payload`. When the failure is attributable to one member path, `error.details.field` MUST identify that path. If the current saved-view version differs from `base_saved_view_version`, the server MUST reject the patch with an explicit conflict status rather than silently overwriting saved-view state. If the request is structurally valid but makes no material change after request-time normalization of `display_name`, `query_json`, and `layout_json`, the server MUST return `200 OK` with the current saved-view resource and MUST NOT change `saved_view_version` or `updated_at`. For saved-view no-op comparison, `query_json` and `layout_json` equality MUST be structural equality after normalization rather than textual JSON equality or JSONB byte equality. Any materially changed successful in-place mutation MUST advance `saved_view_version` and `updated_at` exactly once.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231, AC-360

**REQ-01-147**
`DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` MUST delete only the saved-view configuration object and MUST return the common success envelope with `data.saved_view_id` and `data.deleted=true`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-148**
The workbook-preference route family MUST expose two distinct resources:

- `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` for the authenticated caller's `user_workbook_preferences`,
- `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` for the incident-wide `incident_workbook_preferences`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-149**
Both workbook-preference resources MUST use the stable `sheet_ref` union defined in §3.3.10.1. When the target is any pack-independent base-profile registry surface listed in REQ-01-307, the stored `sheet_ref` MUST use the `view_schema` form with the standardized `view_schema_id`; the `saved_view` form remains valid only for a distinct saved-view object over that schema.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-150**
`workbook-preferences/me` MUST expose, at minimum, `incident_id`, `user_id`, `home_sheet_ref`, `created_at`, and `updated_at`. `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` MUST accept only a JSON object exactly of the form `{ "home_sheet_ref": <sheet_ref|null> }`. If the preference object does not yet exist, the route MUST create it. If it already exists, the route MUST replace only `home_sheet_ref`. Unknown top-level members MUST fail with `400` and `error.code = invalid_mutation_payload`. A structurally valid no-op update MUST return `200 OK` with the current resource and MUST NOT change `updated_at`. An effective change MUST return `200 OK` with the resulting resource and MUST update `updated_at` exactly once. The route MUST allow any current incident member to set or clear only their own home-surface preference.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-151**
`workbook-preferences/default` MUST expose, at minimum, `incident_id`, `default_sheet_ref`, `created_at`, `updated_at`, and `updated_by_user_id`. `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` MUST accept only a JSON object exactly of the form `{ "default_sheet_ref": <sheet_ref|null> }`. If the preference object does not yet exist, the route MUST create it. If it already exists, the route MUST replace only `default_sheet_ref`. Unknown top-level members MUST fail with `400` and `error.code = invalid_mutation_payload`. A structurally valid no-op update MUST return `200 OK` with the current resource and MUST NOT change `updated_at` or `updated_by_user_id`. An effective change MUST return `200 OK` with the resulting resource, MUST update `updated_at` exactly once, and MUST set `updated_by_user_id` to the current actor. The route MUST fail closed for callers whose current incident role is not `admin`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-151.1**
`GET /api/v1/incidents/{incident_id}/workbook-startup` MUST expose workbook startup selection using the ordered fallback owned by Core 03 §2.4. The route MUST accept either legacy `view_schema_id=<id>` or general `sheet_ref_kind=view_schema|saved_view&sheet_ref_id=<id>` query selectors for an explicit launch pointer; supplying both selector forms MUST fail with `400` and `error.code = invalid_startup_request`. A successful response MUST include `incident_id`, `selected_sheet_ref`, `selected_view_schema_id`, `selected_saved_view`, `source`, `cleared_pointers[]`, `home_sheet_ref`, and `default_sheet_ref`. `selected_view_schema_id` is the base schema used by workbook query routes; `selected_sheet_ref` is the selected startup identity and MAY be a distinct `saved_view` reference. The route MUST NOT treat an empty saved-view list as absence of any pack-independent base-profile surface identified by `view_schema`.
Profiles: base
Verified by: AC-150, AC-153, AC-231

#### 3.3.5.3 Incident resource and creation contract

**REQ-01-152**
This subsection owns the authoritative base-profile `/api/v1/incidents` route family and `GET /api/v1/incidents/{incident_id}` read contract. Later sections MAY reference this contract but MUST NOT redefine base-profile collection pagination, create-time required fields, server-managed initial values, or the boundary between create-time-only fields and later patchable fields.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-153**
The incident resource MUST expose, at minimum:

- `incident_id`,
- `incident_key`,
- `title`,
- `description`,
- `status`,
- `severity`,
- `tlp`,
- `current_phase`,
- `primary_external_case_ref`,
- `created_by_user_id`,
- `created_at`,
- `updated_at`,
- `updated_by_user_id`,
- `incident_version`,
- `closed_at`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

In the base profile, `description`, `severity`, `tlp`, `current_phase`, `primary_external_case_ref`, and `closed_at` are nullable incident resource fields.

**REQ-01-154**
`POST /api/v1/incidents` MUST require an authenticated session using the same session contract as the remaining API surface. Cookie-authenticated requests MUST fail closed without valid CSRF protection.
Profiles: base
Verified by: AC-130, AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-155**
The base-profile authorization gate for `POST /api/v1/incidents` MUST be an authenticated session whose internal user account is active and not disabled. This route MUST NOT require a pre-existing incident membership because the caller is creating the workspace boundary itself.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-156**
`POST /api/v1/incidents` request bodies MUST be JSON objects and MUST accept:

- required `client_txn_id`,
- required `incident_key`,
- required `title`,
- optional nullable `description`,
- optional nullable `severity`,
- optional nullable `tlp`,
- optional nullable `current_phase`,
- optional nullable `primary_external_case_ref`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-157**
`incident_key` MUST be client-supplied. The server MUST NOT auto-generate or overwrite it. `incident_key` MUST be trimmed of leading and trailing Unicode whitespace, Unicode NFC-normalized, non-empty, at most 128 UTF-8 bytes after that normalization, and unique within the deployment after that same normalization. The incident resource returned by the public API MUST serialize `incident_key` in that trimmed Unicode-NFC-normalized form. `title` MUST be trimmed of leading and trailing Unicode whitespace, non-empty, and at most 512 Unicode scalar values after trimming. The incident resource returned by the public API MUST serialize `title` in that trimmed form. If present, `description` MUST be at most 16384 Unicode scalar values. If present, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref` MUST each be at most 128 Unicode scalar values. The server MUST reject control characters in `incident_key` and `title`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-158**
If `description`, `severity`, `tlp`, `current_phase`, or `primary_external_case_ref` are omitted or explicitly `null`, the initial incident resource MUST expose that field as `null`. The base profile MUST NOT synthesize an implicit non-null default for any of those optional fields. In the base profile, `severity` is optional at create time. `status` is server-managed on create and MUST NOT be client-settable. In the base profile, `incident_key`, `title`, `description`, `status`, and `severity` are create-time-only incident fields. After create, later metadata mutation is limited to the fields defined in §3.3.5.3.1.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-159**
`POST /api/v1/incidents` uses a closed top-level request namespace. The only client-settable top-level request members are `client_txn_id`, `incident_key`, `title`, `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`. The server MUST reject any other top-level member. The server MUST also reject any attempt to set server-managed fields including `incident_id`, `status`, `created_by_user_id`, `created_at`, `updated_at`, `updated_by_user_id`, `incident_version`, `closed_at`, any membership object, any saved-view object, and any workbook-preference object.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-160**
Idempotency comparison for `POST /api/v1/incidents` MUST run only after the request passes validation and create-time normalization. The normalized request MUST include only declared request members recognized by this contract. Unknown or otherwise invalid top-level members MUST never participate in normalized-request comparison because the route rejects them before idempotency evaluation. For optional request fields declared by this contract, omission and an explicit JSON `null` MUST compare equal only when this section defines the field as optional and nullable.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-161**
If the request is not a JSON object, omits required `client_txn_id`, `incident_key`, or `title`, supplies `null` for a non-nullable field, violates a field validation rule in this section, attempts to set a server-managed field, supplies any unknown top-level member, or supplies `initial_memberships[]`, the server MUST fail with `400` and `error.code = invalid_incident_create`. When the failure is attributable to one request member, `error.details.field` MUST identify that top-level member. `error.details.reason_code` MUST use the registry in §3.3.6.2.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-162**
If the normalized `incident_key` conflicts with an existing incident, the server MUST fail with `409` and `error.code = incident_key_conflict`. `error.details` MUST include at least `field` with value `incident_key` and `incident_key_canonical`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-163**
On success, the server MUST, in one transaction:

1. insert the incident with a freshly assigned `incident_id`, `status='active'`, `incident_version=1`, `created_by_user_id` bound to the authenticated user, `updated_by_user_id` bound to that same user, `closed_at=NULL`, and one committed create timestamp used for both `created_at` and `updated_at`,
2. insert one `incident_memberships` row for the creator with `role='admin'`,
3. create the incident-wide workbook-preference object with `default_sheet_ref=NULL`,
4. create the creator's per-user workbook-preference object with `home_sheet_ref=NULL`,
5. persist attributed audit history sufficient to reconstruct the initial incident state and the bootstrap membership.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-164**
The returned incident resource for a first-time successful create and for an idempotent replay MUST therefore expose the same `incident_id`, `incident_version=1`, `status='active'`, `closed_at=null`, `created_by_user_id`, `updated_by_user_id`, and equal `created_at` and `updated_at` values.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-165**
The base profile MUST NOT accept `initial_memberships[]` on this route. Membership management beyond creator bootstrap belongs to the separate membership-write contract defined by §3.3.5.1.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-166**
Idempotency for create MUST be keyed by `(actor_user_id, client_txn_id)`. If the same authenticated user replays the same normalized request with the same `client_txn_id`, the server MUST return `200 OK`, MUST set `Location` to `/api/v1/incidents/{incident_id}` for the originally created incident, and MUST return the common success envelope with `data` equal to the originally created incident resource. If the same authenticated user reuses `client_txn_id` with a different normalized request, the server MUST fail with `409` and `error.code = client_txn_conflict`. `error.details` MUST include at least `client_txn_id`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-167**
A first-time successful create MUST return `201 Created`, MUST set `Location` to `/api/v1/incidents/{incident_id}`, and MUST return the common success envelope with `data` equal to the incident resource.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231


##### 3.3.5.3.1 Incident list, retrieval, and metadata patch contract

**REQ-01-168**
`GET /api/v1/incidents` MUST return only incidents for which the caller currently has membership, MUST use the common success envelope with `data.incidents[]` plus `meta.paging`, MUST order results by `updated_at desc, incident_id asc`, MUST accept only `limit` and `cursor_token` under §3.3.7, and MUST define no additional base-profile list filters. `GET /api/v1/incidents/{incident_id}` MUST return the common success envelope with `data` equal to the requested incident resource, any current incident member MAY call this route, and because it is singleton the route MUST reject `limit`, `cursor_token`, and pagination aliases with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code=pagination_not_supported`.
Profiles: base
Verified by: AC-127, AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-169**
`incident_version` MUST equal `1` on a first-time successful create and MUST be monotonically increasing per `incident_id`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-170**
`PATCH /api/v1/incidents/{incident_id}` MUST be the only base-profile mutation route for incident-scoped metadata. It MUST NOT reuse `PATCH /api/v1/records/{record_id}`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-171**
This route is incident-resource-scoped, not view-scoped. It MUST NOT require `view_schema_id`, `field_key`, a row envelope, or other workbook-surface routing parameters in the path or request body.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-172**
The route MUST require the caller's current role on that incident to be `reviewer` or `admin`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-173**
The request body MUST be a JSON object and MUST accept `base_incident_version` plus changed mutable fields only. In the base profile, mutable fields are `tlp`, `current_phase`, and `primary_external_case_ref`. This mutable-field list is exhaustive for post-create incident-metadata mutation in the base profile.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-174**
The server MUST apply the same normalization and validation rules defined for those fields on `POST /api/v1/incidents`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

For nullable mutable fields, omission means no change and explicit JSON `null` means clear the field.

**REQ-01-175**
The route MUST reject attempted mutation of `incident_id`, `incident_key`, `title`, `description`, `status`, `severity`, `created_by_user_id`, `created_at`, `updated_at`, `updated_by_user_id`, `incident_version`, `closed_at`, any membership object, any saved-view object, and any workbook-preference object. The route MUST reject unknown top-level request members with `400` and `error.code = invalid_incident_patch`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-176**
If the current `incident_version` differs from `base_incident_version`, the server MUST fail with `409` and `error.code = incident_version_conflict`. `error.details` MUST include at least `incident_id`, `base_incident_version`, and `current_incident_version`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-177**
If the normalized request changes no effective field value, the server MUST return `200 OK` with the current incident resource and MUST NOT increment `incident_version`, change `updated_at`, or change `updated_by_user_id`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-178**
On success, the server MUST, in one transaction:

1. update the changed structured incident fields,
2. set `updated_at` and `updated_by_user_id`,
3. increment `incident_version`,
4. persist attributed audit history sufficient to reconstruct the before and after values of each changed field.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-179**
A successful `PATCH /api/v1/incidents/{incident_id}` call MUST return `200 OK` and the common success envelope with `data` equal to the updated incident resource.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-180**
A caller who lacks visibility to the incident MUST receive `404`. A caller who can see the incident but lacks sufficient role MUST receive `403`.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

#### 3.3.5.4 Entity-merge contract


**Table 3.3.5.4-A. Merge route contract**

| Member or rule | Requirement |
| --- | --- |
| Required members | `loser_record_id`, `survivor_base_row_version`, `loser_base_row_version`, `client_txn_id` |
| Optional members | `reason`, normalized under `reason_note_v1`; omission, explicit `null`, and normalized empty compare equal |
| Valid target set | Different visible non-deleted records in the same incident, same `record_type`, and that `record_type` is exactly `host` or `identity` |
| Role gate | Current incident role `reviewer` or `admin` |
| Idempotency | `(actor_user_id, survivor_record_id, loser_record_id, client_txn_id)` |
| Success summary | `incident_id`, `record_type`, survivor and loser ids and row versions, `change_set_id`, `merged_into_record_id`, and `merge_summary` with exact-match-class counts and carry-forward counts |
| Algorithmic detail retained in prose | Deterministic carry-forward of exact-match classes, loser-state preservation, deduplication of repointed links and tags, and collaboration refresh behavior |


**REQ-01-181**
`POST /api/v1/records/{survivor_record_id}/merge` MUST initiate one explicit entity merge.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-182**
The route is record-scoped, not view-scoped. It MUST NOT require `incident_id` or `view_schema_id` in the path or request body, because authorization, history, and affected projections derive from the authoritative record identities involved.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-183**
The request MUST be JSON and MUST include:

- `loser_record_id`,
- `survivor_base_row_version`,
- `loser_base_row_version`,
- `client_txn_id`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-184**
It MAY include optional `reason` bound to `string_contract_id=reason_note_v1`. For idempotency comparison, omission, explicit JSON `null`, and any supplied `reason` value that normalizes to empty under `reason_note_v1` MUST compare equal.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

The route is valid only when all of the following are true:

- `survivor_record_id` and `loser_record_id` are different,
- both records exist and are visible to the caller,
- both records belong to the same incident,
- both records have the same `record_type`,
- that shared `record_type` is `host` or `identity`,
- neither record is soft-deleted,
- neither record is already merged away.

**REQ-01-185**
The server MUST NOT silently choose, swap, or rewrite survivor versus loser. The client MUST choose the survivor explicitly, and the survivor `record_id` MUST remain the stable anchor of the operation.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-186**
`POST /api/v1/records/{survivor_record_id}/merge` MUST require current incident role `reviewer` or `admin`. A caller who lacks visibility to either record MUST receive `404`. A caller who can see both records but lacks sufficient role MUST receive `403`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-187**
Idempotency for merge MUST be keyed by `(actor_user_id, survivor_record_id, loser_record_id, client_txn_id)`. If the same authenticated actor replays the same normalized request with the same key, the server MUST return `200 OK` with the originally committed result and MUST create no second merge `change_set`. If the same actor reuses that key with a different normalized request, the server MUST fail with `409`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-188**
If either current row version differs from the supplied base version, the route MUST fail with `409` using the common error envelope and `error.code = row_version_conflict`. `error.details` MUST include at least:

- `survivor_record_id`,
- `loser_record_id`,
- `survivor_base_row_version`,
- `loser_base_row_version`,
- `survivor_current_row_version`,
- `loser_current_row_version`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-189**
If a precondition other than row-version freshness fails, the route MUST fail with `409` and `error.code = merge_precondition_failed`. `error.details.reason_code` MUST use the canonical `merge_precondition_failed` reason-code registry defined in §3.3.6.2.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-190**
`POST /api/v1/records/{survivor_record_id}/merge` MUST participate in the base-profile destructive-operation concurrency contract defined by REQ-01-104. The merge-specific protected set described there MUST be acquired before row-version or merge-precondition evaluation proceeds.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-191**
On success, the server MUST commit the merge in one transaction. It MUST:

- preserve the survivor `record_id` unchanged,
- mark the loser as a historical row with state `merged` and `merged_into_record_id` set to the survivor,
- repoint active `entity_mentions.resolved_record_id`, active `record_links`, active assessments, and active tags from loser to survivor in the same `change_set`, or otherwise tombstone and recreate them deterministically,
- deduplicate duplicate links and tags created by repointing without losing revision history,
- preserve raw mention text unchanged,
- evaluate loser-side preserved host or identity identifiers using the same exact-match normalization and comparison substrate used for ordinary create-or-upsert reuse,
- create one attributed `change_set` plus ordered mutation entries sufficient to reconstruct the pre-merge graph, post-merge graph, and merge fan-out,
- update or invalidate affected projections before commit returns.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-192**
In accordance with Core 02 §9, the server MUST apply the deterministic carry-forward algorithm for loser-side host or identity preserved identifiers and aliases. It MUST NOT overwrite conflicting survivor canonical values, silently drop loser-side `exact_match_reuse` values, or downgrade them to `suggestion_only` behavior.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-193**
A successful response MUST return `200 OK` using the common success envelope. `data` MUST include at least:

- `incident_id`,
- `record_type`,
- `survivor_record_id`,
- `loser_record_id`,
- `survivor_row_version`,
- `loser_row_version`,
- `change_set_id`,
- `merged_into_record_id`,
- `merge_summary`.

`merge_summary` MUST include:

- `exact_match_classes[]`,
- `suggestion_aliases_copied_count`,
- `suggestion_alias_duplicate_noop_count`,
- `provenance_only_retained_count`.

`exact_match_classes[]` MUST be ordered by the exact-match precedence order for the merged `record_type` in Core 02 §8.2 and MUST contain one entry for each exact-match class for that `record_type`, even when all counts are zero. Each entry MUST include:

- `identifier_class`,
- `promoted_count`,
- `carried_count`,
- `duplicate_noop_count`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-194**
A successful merge MUST emit replayable collaboration events through the existing stream. The loser MUST leave ordinary active entity views using `change_kind = remove`. The survivor MUST refresh through `patch` or `invalidate`. Any dependent row whose chips, counts, or linked-record summaries change because of repointing MUST refresh through the ordinary collaboration mechanisms.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-195**
The base profile defines no separate unmerge route. Reversal of an erroneous merge MUST use `POST /api/v1/records/{record_id}/rollback` with `target.kind = 'change_set'` and the merge `change_set_id`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231


#### 3.3.5.5 Entity-mention action contract


**Table 3.3.5.5-A. Entity-mention action contract**

| Member or rule | Requirement |
| --- | --- |
| Required members | `base_mention_row_version`, `client_txn_id`, `action` |
| Optional members | `resolved_record_id`, `reason` |
| Closed `action` vocabulary | `resolve_item`, `dismiss_item`, `revert_to_unresolved` |
| `resolved_record_id` rule | Required if and only if `action='resolve_item'`; forbidden, including JSON `null`, for the other actions |
| Role gate | Current incident role `editor`, `reviewer`, or `admin` on the source incident |
| Idempotency | `(actor_user_id, entity_mention_id, client_txn_id)` |
| Legal transition matrix | `resolve_item`: `unresolved` or `resolved` -> `resolved`; `dismiss_item`: `unresolved` or `resolved` -> `dismissed`; `revert_to_unresolved`: `resolved` or `dismissed` -> `unresolved` |
| Success summary | `200 OK` with `incident_id`, `entity_mention`, `source_record`, and `change_set_id`; the source record row version advances and ordinary `record_changed` refresh applies |
| Primary failures | `invalid_mutation_payload`, `client_txn_conflict`, `row_version_conflict`, `entity_mention_not_found`, `resolved_record_not_found`, `illegal_transition`, `record_deleted_use_restore` |


**REQ-01-196**
`POST /api/v1/entity-mentions/{entity_mention_id}/resolve` MUST apply one explicit action to one `entity_mentions` row.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

The stable path segment `resolve` names the route family only. The request body's `action` member determines whether the committed action resolves, dismisses, or reverts the targeted mention.

**REQ-01-197**
The route is mention-scoped, not view-scoped. It MUST NOT require `incident_id`, `view_schema_id`, `field_key`, `link_type`, table names, or storage-routing metadata in the path or request body, because authorization, lifecycle state, and relationship routing derive from the authoritative `entity_mentions` row identified by `entity_mention_id`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-198**
In the base profile, the route MUST support mentions whose authoritative `source_field_key` is `timeline.host_refs` or `timeline.identity_refs`. Other source fields MAY reuse this route only when their field contracts declare the same explicit mention-action vocabulary and corresponding relationship semantics. If the targeted mention's current `source_field_key` does not declare that action vocabulary, the route MUST fail with `409` and `error.code = illegal_transition`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-199**
This route MUST be the public wire-contract owner for the inspector's explicit mention actions. The legal transition matrix in this subsection MUST also govern single-target `resolve_item`, `dismiss_item`, and `revert_to_unresolved` actions sent through `collection_actions_v1` for `timeline.host_refs` and `timeline.identity_refs`. The base profile defines no additional standalone public mention-action route family. Workbook-surface single-target mention actions for those fields use the already enumerated `PATCH /api/v1/records/{record_id}` mutation surface and MUST NOT be inventoried or described elsewhere as separate public routes.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-201, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-200**
The request MUST be a JSON object and MUST include:

- `base_mention_row_version`,
- `client_txn_id`,
- `action`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

The request MAY include:

- `resolved_record_id`,
- `reason`.

**REQ-01-201**
If present, `reason` MUST be a JSON string or JSON `null` and MUST be normalized using `string_contract_id=reason_note_v1`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-202**
`action` MUST use one of the exact tokens `resolve_item`, `dismiss_item`, or `revert_to_unresolved`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-203**
`resolved_record_id` is required if and only if `action='resolve_item'`. When `action` is `dismiss_item` or `revert_to_unresolved`, `resolved_record_id` MUST NOT be present, including JSON `null`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-204**
For idempotency comparison, omission, explicit JSON `null`, and any supplied `reason` value that normalizes to empty under `reason_note_v1` MUST compare equal.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-205**
Unknown top-level request members MUST fail with `400` and `error.code = invalid_mutation_payload`. This includes `base_row_version`, `record_id`, `incident_id`, `view_schema_id`, `field_key`, `entity_type`, `source_record_id`, `source_field_key`, `link_type`, table names, or storage-routing metadata.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-206**
This route MUST NOT create a new host or identity record. A workflow that creates a host or identity from a selected mention MUST use the ordinary entity-create or create-or-upsert contract first and then resolve the mention to the resulting `record_id`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-207**
A request example MUST use this shape:
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

```json
{
  "base_mention_row_version": 7,
  "client_txn_id": "8d6b4b48-0b7a-4d9e-9c0d-6a5d5f3f1d7a",
  "action": "resolve_item",
  "resolved_record_id": "0d7dd3d2-50b5-4b7a-a0d4-9f0d08b6c7e4",
  "reason": "optional audit note"
}
```

**REQ-01-208**
`POST /api/v1/entity-mentions/{entity_mention_id}/resolve` MUST require current incident role `editor`, `reviewer`, or `admin` on the source incident of the targeted mention. A caller who lacks visibility to the targeted mention, or whose `entity_mention_id` identifies a deleted mention row, MUST receive `404` and `error.code = entity_mention_not_found`. A caller who can see the mention but lacks sufficient role MUST receive `403`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-209**
Idempotency for this route MUST be keyed by `(actor_user_id, entity_mention_id, client_txn_id)`. If the same authenticated actor replays the same normalized request with the same key, the server MUST return `200 OK` with the originally committed result and MUST create no second `change_set`. If the same actor reuses that key with a different normalized request, the server MUST fail with `409` and `error.code = client_txn_conflict`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-210**
If the current mention `row_version` differs from `base_mention_row_version`, the route MUST fail with `409` using the common error envelope and `error.code = row_version_conflict`. `error.details` MUST include at least `entity_mention_id`, `base_mention_row_version`, `current_mention_row_version`, and `source_record_id`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-211**
If the source record of the targeted mention is currently soft-deleted, the route MUST fail with `409` and `error.code = record_deleted_use_restore`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-212**
When `action='resolve_item'`, `resolved_record_id` MUST identify a visible active target record in the same incident whose `record_type` matches the targeted mention's `entity_type`. If no such visible active target record exists, the route MUST fail with `404` and `error.code = resolved_record_not_found`. If a visible target record exists but fails the same-incident or `entity_type` compatibility check, the route MUST fail with `400` and `error.code = invalid_mutation_payload`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-213**
The legal transition matrix is:

- `resolve_item`: current `resolution_status` MUST be `unresolved` or `resolved`; the committed state MUST be `resolved`.
- `dismiss_item`: current `resolution_status` MUST be `unresolved` or `resolved`; the committed state MUST be `dismissed`.
- `revert_to_unresolved`: current `resolution_status` MUST be `resolved` or `dismissed`; the committed state MUST be `unresolved`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-214**
If the current `resolution_status` does not permit the requested action under this matrix, the route MUST fail with `409` and `error.code = illegal_transition`. `error.details.from_status`, `error.details.to_status`, and `error.details.violated_guards[]` MUST follow §3.3.6. `error.details.to_status` MUST be `resolved`, `dismissed`, or `unresolved` according to the requested action.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-215**
A successful `resolve_item` MUST:

- preserve `raw_text`,
- set `resolution_status='resolved'`,
- set `resolved_record_id`,
- set `resolved_by_user_id`,
- set `resolved_at`,
- set `resolution_method='explicit_resolve_route'`,
- create or upsert the corresponding active resolved `record_link` in the same `change_set`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-216**
If the mention was previously resolved to a different target, the old active resolved link MUST be removed or tombstoned in the same committed `change_set`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-217**
A successful `dismiss_item` MUST:

- preserve `raw_text`, stable mention identity, and provenance,
- set `resolution_status='dismissed'`,
- clear `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`,
- remove or tombstone any corresponding active resolved `record_link` in the same `change_set`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-218**
A successful `revert_to_unresolved` MUST:

- preserve `raw_text`,
- set `resolution_status='unresolved'`,
- clear `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`,
- leave no active resolved `record_link`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-219**
If the mention was previously dismissed, ordinary `revert_to_unresolved` MUST NOT silently relink any prior resolved target. Exact pre-dismiss state recovery MUST remain a rollback operation.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-220**
A successful route invocation MUST, in one transaction:

- create one attributed `change_set`,
- increment the targeted mention `row_version`,
- increment the source record `row_version`,
- update or invalidate affected projections before commit returns.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-221**
A successful route invocation MUST emit the ordinary replayable `record_changed` event for the source record and MUST NOT introduce a mention-specific collaboration event family.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-222**
A successful response MUST return `200 OK` using the common success envelope. `data` MUST include at least:

- `incident_id`,
- `entity_mention`,
- `source_record`,
- `change_set_id`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-223**
The returned `entity_mention` resource MUST include at least:

- `entity_mention_id`,
- `source_record_id`,
- `source_field_key`,
- `entity_type`,
- `raw_text`,
- `normalized_text`,
- `resolution_status`,
- `resolved_record_id`,
- `row_version`,
- `resolved_at`,
- `resolved_by_user_id`,
- `resolution_method`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-224**
When the committed state has no current resolved target, `resolved_record_id`, `resolved_at`, `resolved_by_user_id`, and `resolution_method` MUST be present and set to JSON `null`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-225**
`data.source_record` MUST include at least `record_id` and current committed `row_version`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-226**
If the committed state leaves an active resolved link, `data.active_link` MUST be present and MUST include `link_type` and `dst_record_id`. In the base profile, `link_type` MUST be derived from the targeted mention's `source_field_key` using the mapping declared in §7.4.1. If the committed state leaves no active resolved link, `data.active_link` MUST be omitted.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-227**
A success example MUST use this shape:
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

```json
{
  "data": {
    "incident_id": "3dbe2d2d-9b71-4fe5-b6ee-5dd0a5d8441d",
    "entity_mention": {
      "entity_mention_id": "6a7d3b31-4c5c-48f2-a0dc-5b41d9a9d2a1",
      "source_record_id": "2d7f0d4a-5e7e-4d2a-93b2-f0d9ce2b9ad4",
      "source_field_key": "timeline.host_refs",
      "entity_type": "host",
      "raw_text": "WS-023?",
      "normalized_text": "ws-023?",
      "resolution_status": "resolved",
      "resolved_record_id": "0d7dd3d2-50b5-4b7a-a0d4-9f0d08b6c7e4",
      "row_version": 8,
      "resolved_at": "2026-03-26T15:41:00Z",
      "resolved_by_user_id": "9b5d8f8f-2d1c-4f62-8f0b-c6b4c4b5c8d1",
      "resolution_method": "explicit_resolve_route"
    },
    "source_record": {
      "record_id": "2d7f0d4a-5e7e-4d2a-93b2-f0d9ce2b9ad4",
      "row_version": 42
    },
    "change_set_id": "4b2c9f50-cd12-4d3d-b6b8-c8c8ff6f0c01",
    "active_link": {
      "link_type": "observed_on_host",
      "dst_record_id": "0d7dd3d2-50b5-4b7a-a0d4-9f0d08b6c7e4"
    }
  },
  "meta": {
    "request_id": "req_01..."
  }
}
```

#### 3.3.6 Success and error envelopes

**REQ-01-228**
All successful JSON responses MUST use a common envelope with:

- `data` for the primary resource, row batch, or job resource returned by the route,
- `meta.request_id` for a server-generated correlation identifier,
- optional `meta.warnings[]` for machine-readable or display-safe warnings,
- optional `meta.paging` for the cursor metadata defined in §3.3.7,
- optional `meta.query` for the applied view-query contract after server normalization.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231

**REQ-01-229**
All non-success JSON responses MUST use a common error envelope with:

- `error.code` as a stable machine-readable error code,
- `error.message` as a human-readable summary,
- `error.status` as the transport status,
- `error.request_id` as a correlation identifier,
- `error.retryable` as an explicit retry hint,
- optional `error.details` for route-specific validation or state details.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231

**REQ-01-230**
Same-field conflicts MUST use this same error family with `error.code = same_field_conflict` and the additional conflict object defined by Core 03 §3.3.4.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231

**REQ-01-231**
Illegal lifecycle transitions MUST use this same error family with `error.code = illegal_transition`, `error.status = 409`, `error.details.from_status`, `error.details.to_status`, and `error.details.violated_guards[]`. `error.details.violated_guards[]` MUST be present and MAY be empty when the transition is disallowed by the legal transition matrix rather than by a failed field guard.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231

**REQ-01-232**
Clients MUST tolerate additive response members they do not use.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231

**REQ-01-233**
Public `error.code` tokens and public `reason_code` tokens defined by this core are canonicalized below. Other sections MAY require one of these tokens for a specific route or conformance criterion but MUST NOT redefine its primary meaning, required transport status, or retry hint.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231

##### 3.3.6.1 Canonical public error-code registry

**REQ-01-234**
The public API surface defined by this core MUST use the following stable `error.code` tokens for the listed conditions. A route or conformance criterion covered by this registry MUST NOT assign a second stable token to the same condition.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231, AC-239, AC-240, AC-245, AC-246, AC-247, AC-249, AC-250, AC-251, AC-252, AC-253, AC-254, AC-255, AC-260, AC-261, AC-293, AC-321, AC-323, AC-324, AC-325, AC-326, AC-328, AC-340, AC-341, AC-342, AC-334, AC-335, AC-336, AC-337, AC-338, AC-339, AC-371

| `error.code` | Required `error.status` | Required `error.retryable` | Canonical meaning | Requirement ID | Profiles | Verified by |
| --- | --- | --- | --- | --- | --- | --- |
| `invalid_view_query` | `400` | `false` | The view-query request is malformed, uses a page-size member or page-size value not allowed by the route, replays a cursor against a different bound query contract, or uses a `field_key`, filter operator, or operand shape not allowed by the active `view_schema_id`. |  |  |  |
| `invalid_pagination_request` | `400` | `false` | A non-view pageable or singleton route uses a page-size member or page-size value not allowed by the route, replays a cursor against a different bound route contract, or supplies pagination members to a route that does not support pagination. |  |  |  |
| `invalid_mutation_payload` | `400` | `false` | A mutation request body is malformed, omits a route-required member, includes an unknown or forbidden member, uses an unknown `kind`, `op`, or `action`, targets a field/action or mention-action combination that is not allowed, or carries an invalid, foreign, or type-incompatible mutation target reference. |  |  |  |
| `invalid_evidence_handle_request` | `400` | `false` | A preview-handle or download-handle issuance request is malformed, uses a non-object JSON body, omits the required JSON object wrapper, or includes an unknown top-level member. |  |  |  |
| `invalid_blob_create_request` | `400` | `false` | A blob-slot create request is malformed, omits required members, violates create-time field validation, attempts to set server-managed state, or includes an unknown top-level member. |  |  |  |
| `blob_create_rejected` | `413` | `false` | A blob-slot create request is structurally valid but exceeds the configured declared-size ceiling for `POST /api/v1/object-blobs`. `error.details.reason_code` MUST use the `blob_create_rejected` registry in §3.3.6.2. |  |  |  |
| `invalid_incident_create` | `400` | `false` | An incident-create request is malformed, omits required members, includes an unknown top-level member, violates create-time field validation, attempts to set server-managed state, or includes a rejected collaborator-seeding payload. |  |  |  |
| `invalid_incident_patch` | `400` | `false` | An incident-metadata patch request is malformed, omits required `base_incident_version`, attempts to mutate an immutable or server-managed incident field, or includes unknown top-level members. |  |  |  |
| `invalid_rollback_request` | `400` | `false` | A rollback request is malformed, uses an unknown or unsupported `target.kind`, omits the selector required for that `kind`, includes unknown request members, or supplies a selector whose JSON type does not match the declared shape. |  |  |  |
| `invalid_auth_request` | `400` | `false` | A local-account login request is malformed, omits a required member, includes an unknown or forbidden member, supplies `null` where forbidden, uses an unsupported `second_factor.kind`, or carries an invalid TOTP assertion shape. |  |  |  |
| `invalid_enterprise_auth_request` | `400` | `false` | An enterprise-auth discovery or initiation request is malformed, omits a required member, includes an unknown or forbidden member, supplies `null` where forbidden, or uses a `return_to` value not allowed by the current profile. |  |  |  |
| `extension_profile_not_claimed` | `404` | `false` | The request path matches a reserved extension route family for a profile the deployment does not currently claim. `error.details` MUST include `profile_id` and `route_family`. |  |  |  |
| `auth_provider_not_found` | `404` | `false` | The addressed enterprise-auth `provider_key` does not identify a configured enterprise-auth provider allowed by the active route. |  |  |  |
| `auth_provider_disabled` | `409` | `false` | The addressed enterprise-auth provider exists but is not currently enabled for interactive sign-in. |  |  |  |
| `enterprise_auth_transaction_rejected` | `409` | `false` | The current enterprise-auth callback or ACS request cannot use the bound server-side auth transaction because the transaction is missing, expired, already consumed, bound to a different provider, or no longer matches the browser binding context. `error.details.reason_code` MUST use the `enterprise_auth_transaction_rejected` registry in §3.3.6.2. |  |  |  |
| `provider_response_rejected` | `409` | `false` | The enterprise-auth provider response failed protocol validation or did not satisfy the bound callback contract. `error.details.reason_code` MUST use the `provider_response_rejected` registry in §3.3.6.2. |  |  |  |
| `provider_identity_rejected` | `409` | `false` | The enterprise-auth provider response completed far enough to identify or attempt to identify one provider-backed subject, but the current profile could not bind that subject to exactly one active local user. `error.details.reason_code` MUST use the `provider_identity_rejected` registry in §3.3.6.2. |  |  |  |
| `invalid_credentials` | `401` | `false` | The server is not willing to acknowledge that primary credentials were valid on the local-account login route, including for unknown login identifier, wrong password, inactive local account, or equivalent pre-MFA failure. |  |  |  |
| `mfa_required` | `401` | `false` | Primary credentials are valid for a local account that requires MFA, but the login request omitted the required second-factor assertion. On the base local-account login route, `error.details.required_second_factor_kinds` lists the accepted kinds. |  |  |  |
| `mfa_setup_required` | `401` | `false` | Primary credentials are valid for a local account with `mfa_required=true`, but no active TOTP credential is currently enrolled. The response MUST set no session cookie, MUST include `error.details.required_setup_kinds=["totp"]`, and MUST include one `bootstrap_token` plus `bootstrap_expires_at`. |  |  |  |
| `credential_bootstrap_rejected` | `409` | `false` | The supplied credential-setup bootstrap token cannot be used because it is expired, consumed, superseded, bound to a different subject, or used on a route outside its allowed family. `error.details.reason_code` MUST use the `credential_bootstrap_rejected` registry in §3.3.6.2. |  |  |  |
| `invalid_current_password` | `409` | `false` | A credential-lifecycle route that requires re-verification of the caller's current password received a structurally valid value that does not match the current stored local password. |  |  |  |
| `invalid_second_factor` | `401` | `false` | Primary credentials are valid and the local-account login request supplied a structurally valid second-factor assertion, but the asserted factor is wrong or expired. |  |  |  |
| `totp_setup_not_pending` | `409` | `false` | A TOTP-completion request referenced no pending enrollment, or the referenced pending enrollment is expired or already consumed. `error.details.reason_code` MUST use the `totp_setup_not_pending` registry in §3.3.6.2. |  |  |  |
| `client_txn_conflict` | `409` | `false` | The caller reused a `client_txn_id` within the same route-defined idempotency scope for a different normalized request. |  |  |  |
| `job_cancel_rejected` | `409` | `false` | A visible job exists, but the server will not accept cancellation because cancellation is already requested, the job is already terminal, or the current non-terminal phase is not cancelable; `error.details.reason_code` MUST use the `job_cancel_rejected` registry in §3.3.6.2. |  |  |  |
| `row_version_conflict` | `409` | `false` | The supplied `base_row_version` or `base_mention_row_version` is stale relative to authoritative current state on a strict row-version route, or an ordinary record patch could not evaluate its field-level stale window because required revision history was missing or unusable. For restore, this includes a stale tombstone `row_version`. |  |  |  |
| `incident_key_conflict` | `409` | `false` | The normalized `incident_key` conflicts with an existing incident. |  |  |  |
| `incident_version_conflict` | `409` | `false` | The supplied `base_incident_version` is stale. |  |  |  |
| `same_field_conflict` | `409` | `false` | Another committed write touched the same writable `field_key`; the response MUST include the conflict object defined by Core 03 §3.3.4. | REQ-01-235 | base | AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231 |
| `illegal_transition` | `409` | `false` | The requested lifecycle transition is not allowed for the current persisted state or guard condition. |  |  |  |
| `record_deleted_use_restore` | `409` | `false` | The caller targeted a currently soft-deleted record with an operation that requires the record to be restored first. |  |  |  |
| `record_already_deleted` | `409` | `false` | The caller attempted to soft-delete an already soft-deleted record outside an idempotent replay of the original delete. |  |  |  |
| `record_delete_blocked` | `409` | `false` | The caller attempted to soft-delete a record whose type-specific owner preconditions reject deletion while active incoming references or equivalent integrity dependencies exist. `error.details.reason_code` MUST identify the blocking precondition. |  |  |  |
| `record_not_deleted` | `409` | `false` | The caller attempted to restore a record that is not currently soft-deleted. |  |  |  |
| `record_locked` | `409` | `true` | An overlapping in-flight destructive operation already holds one or more required protected-set locks for the requested restore, rollback, or merge. |  |  |  |
| `evidence_attach_rejected` | `409` | `false` | An evidence attach request cannot commit because the supplied blob is not visible or is not attachable, the target evidence row is quarantined or inconsistent, or observed upload bytes violate the accepted blob contract. `error.details.reason_code` MUST use the `evidence_attach_rejected` registry in §3.3.6.2. |  |  |  |
| `evidence_access_unavailable` | `409` | `false` | Preview or download cannot currently proceed because the visible evidence or linked blob is unavailable, pending, failed, missing, quarantined, inconsistent, or not previewable for the requested preview contract or preview-size ceiling. |  |  |  |
| `entity_mention_not_found` | `404` | `false` | An entity-mention action route targeted no visible current entity-mention row for the supplied `entity_mention_id`. |  |  |  |
| `resolved_record_not_found` | `404` | `false` | A mention-resolve request supplied `resolved_record_id` that does not identify a visible active target record. |  |  |  |
| `rollback_target_not_found` | `404` | `false` | A rollback request targeted no visible history item, `change_set_id`, or row revision that is legal for the addressed `record_id`. |  |  |  |
| `evidence_record_not_found` | `404` | `false` | A preview-handle or download-handle issuance request targeted no visible current evidence record for the supplied `record_id`. |  |  |  |
| `handle_not_found_or_revoked` | `404` | `false` | A handle-redeem request targeted no current opaque handle token because the token is unknown, revoked, or no longer available for redeem. |  |  |  |
| `job_not_found` | `404` | `false` | No current visible job resource exists for the supplied `job_id`, including when the job has expired from retention or is outside the caller's current authorization scope. |  |  |  |
| `handle_expired` | `410` | `false` | A previously issued handle token is well-formed but no longer redeemable because its expiry time has passed. |  |  |  |
| `handle_consumed` | `410` | `false` | A single-use handle was already consumed by a prior successful redeem and cannot be redeemed again. |  |  |  |
| `rollback_precondition_failed` | `409` | `false` | A rollback target exists but cannot be safely reversed against current authoritative state; `error.details.reason_code` MUST use the rollback-precondition registry in §3.3.6.2. | REQ-01-236 | base | AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231 |
| `user_version_conflict` | `409` | `false` | The supplied `base_user_version` is stale. |  |  |  |
| `auth_binding_conflict` | `409` | `false` | A create, rotate, or retire request against an enterprise-auth binding cannot commit because current active binding state conflicts. `error.details.reason_code` MUST use the `auth_binding_conflict` registry in §3.3.6.2. |  |  |  |
| `auth_binding_not_found` | `404` | `false` | A binding-management route targeted no current enterprise-auth binding for the supplied `{user_id, auth_binding_id}` pair. |  |  |  |
| `last_deployment_admin` | `409` | `false` | The requested user mutation would leave the deployment with no active `is_deployment_admin=true` user. |  |  |  |
| `user_not_found` | `404` | `false` | A membership-create request referenced a user that does not exist in deployment-local identity state. |  |  |  |
| `user_inactive` | `409` | `false` | A membership-create request referenced a deployment-local user whose `is_active=false`. |  |  |  |
| `membership_exists_use_patch` | `409` | `false` | A membership-create request conflicts with an existing membership for the same `(incident_id, user_id)` and must be expressed as a patch instead. |  |  |  |
| `membership_version_conflict` | `409` | `false` | The supplied `base_membership_version` is stale. |  |  |  |
| `membership_not_found` | `404` | `false` | A membership route that requires an existing current membership targeted no current membership row for the identified `(incident_id, user_id)`. |  |  |  |
| `last_incident_admin` | `409` | `false` | The requested membership create, patch, or delete would leave the incident without any current `admin` membership. |  |  |  |
| `merge_precondition_failed` | `409` | `false` | An entity-merge precondition other than row-version freshness failed; `error.details.reason_code` MUST use the merge-precondition registry in §3.3.6.2. | REQ-01-237 | base | AC-126, AC-187, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231 |

| `invalid_import_request` | `400` | `false` | An import-session create, mapping, select, skip, or apply request is malformed, omits a required member, uses `null` where forbidden, supplies an out-of-range row reference, includes an unknown top-level member, or fails the shared upload-envelope contract for `POST /api/v1/import-sessions`, including unsupported framing, missing or duplicate required parts, unexpected extra parts, invalid metadata encoding or JSON, or invalid part content type. |  |  |  |
| `import_session_not_found` | `404` | `false` | No visible current import session exists for the supplied `import_session_id`. |  |  |  |
| `import_unit_not_found` | `404` | `false` | No visible current import unit exists for the supplied `import_unit_id` within the addressed import session. |  |  |  |
| `import_state_conflict` | `409` | `false` | The addressed import session or unit exists, but its current durable state does not allow the requested mapping, select, skip, or apply action. |  |  |  |
| `import_source_unsupported` | `409` | `false` | The uploaded source file or selected import unit is intentionally unsupported by the current import profile or lacks required inert source material for a mapped apply path. `error.details.reason_code` MUST use the `import_source_unsupported` registry in §3.3.6.2. |  |  |  |
| `import_source_rejected` | `413` | `false` | The uploaded source file or selected import unit is structurally valid but exceeds one or more configured source-byte, workbook-shape, extracted-bytes, compression-ratio, or member-count limits. `error.details.reason_code` MUST use the `import_source_rejected` registry in §3.3.6.2. |  |  |  |
| `import_apply_blocked` | `409` | `false` | The import apply request is structurally valid but blocked by duplicate-apply detection, overlapping selected units, or units that are not ready. `error.details.reason_code` MUST use the `import_apply_blocked` registry in §3.3.6.2. |  |  |  |
| `invalid_snapshot_request` | `400` | `false` | A snapshot-create request is malformed, omits a required member, uses `null` where forbidden, or includes an unknown top-level member. |  |  |  |
| `snapshot_not_found` | `404` | `false` | No visible snapshot exists for the supplied `snapshot_id`. |  |  |  |
| `invalid_release_request` | `400` | `false` | A release-create or release-action request is malformed, omits a required member, supplies `null` where forbidden, or attempts to rely on implicit version selection rather than exact versioned identifiers where exact versioning is required. |  |  |  |
| `release_not_found` | `404` | `false` | No visible release exists for the supplied `release_id`. |  |  |  |
| `release_state_conflict` | `409` | `false` | The addressed release exists, but its current `release_state` does not allow the requested approve, publish, or invalidate action. `error.details.reason_code` MUST use the `release_state_conflict` registry in §3.3.6.2. |  |  |  |
| `release_approval_rejected` | `409` | `false` | The addressed release exists, but the caller or current artifact tuple does not satisfy the approval requirements for the requested approval action. `error.details.reason_code` MUST use the `release_approval_rejected` registry in §3.3.6.2. |  |  |  |
| `release_render_failed` | `409` | `false` | A release render request reached the render phase but failed closed because a required template binding or redaction rule was not satisfied. `error.details.reason_code` MUST use the `release_render_failed` registry in §3.3.6.2. |  |  |  |
| `invalid_reference_pack_request` | `400` | `false` | A reference-pack import, activation, disable, reverify, or refresh request is malformed, omits a required member, uses `null` where forbidden, includes an unknown top-level member, or fails the shared upload-envelope contract for `POST /api/v1/reference-packs/import`, including unsupported framing, missing or duplicate required parts, unexpected extra parts, invalid metadata encoding or JSON, or invalid part content type. |  |  |  |
| `reference_pack_not_found` | `404` | `false` | No visible reference-pack version exists for the supplied `(pack_key, pack_version)` pair. |  |  |  |
| `reference_pack_state_conflict` | `409` | `false` | The addressed reference-pack version exists, but its current durable state does not allow the requested disable or reverify action. `error.details.reason_code` MUST use the `reference_pack_state_conflict` registry in §3.3.6.2. |  |  |  |
| `reference_pack_verification_failed` | `409` | `false` | Reference-pack import, refresh, or reverify failed closed because integrity, compatibility, or content-screening checks did not pass. `error.details.reason_code` MUST use the `reference_pack_verification_failed` registry in §3.3.6.2. |  |  |  |
| `reference_pack_activation_rejected` | `409` | `false` | Activation was rejected because the addressed version is already active or is not in a verified-available condition. `error.details.reason_code` MUST use the `reference_pack_activation_rejected` registry in §3.3.6.2. |  |  |  |
| `invalid_incident_bundle_request` | `400` | `false` | An incident-bundle export or import request is malformed, omits a required member, uses `null` where forbidden, requests unsupported partial-history or partial-blob modes, includes an unknown top-level member, or fails the shared upload-envelope contract for `POST /api/v1/incident-bundles/import`, including unsupported framing, missing or duplicate required parts, unexpected extra parts, invalid metadata encoding or JSON, or invalid part content type. |  |  |  |
| `incident_bundle_not_found` | `404` | `false` | No visible export descriptor exists for the supplied `bundle_id`. |  |  |  |
| `incident_bundle_export_rejected` | `409` | `false` | Whole-incident export could not materialize a conformant bundle because required structured files or required blobs were unavailable. `error.details.reason_code` MUST use the `incident_bundle_export_rejected` registry in §3.3.6.2. |  |  |  |
| `incident_bundle_import_rejected` | `409` | `false` | Whole-incident import failed closed because bundle-member validation, integrity validation, incident-identity collision checks, or capability checks did not pass. `error.details.reason_code` MUST use the `incident_bundle_import_rejected` registry in §3.3.6.2. |  |  |  |

##### 3.3.6.2 Canonical public reason-code registries

**REQ-01-238**
When the public API or collaboration stream uses a structured `reason_code` family listed below, it MUST use one of the exact tokens shown. A listed `reason_code` family MUST NOT define alternate tokens for the same meaning elsewhere in the core.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231, AC-239, AC-240, AC-252, AC-255, AC-260, AC-293, AC-321, AC-322, AC-323, AC-324, AC-325, AC-326, AC-327, AC-328, AC-341, AC-336, AC-337, AC-339, AC-375

`invalid_incident_create` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The request body is not a JSON object. |
| `missing_required_field` | A required incident-create field is absent. |
| `field_not_nullable` | The request supplies `null` for a non-nullable incident-create field. |
| `field_empty_after_normalization` | `incident_key` or `title` is empty after required normalization and trimming. |
| `field_too_long` | A create field exceeds its declared maximum length. |
| `control_character_not_allowed` | `incident_key` or `title` contains a rejected control character. |
| `unknown_field` | The request includes a top-level member not declared by the incident-create contract. |
| `server_managed_field` | The request attempted to set server-managed incident state. |
| `collaborator_seeding_not_supported` | The request supplied `initial_memberships[]`. |

`invalid_blob_create_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The request body is not a JSON object. |
| `missing_required_field` | A required blob-create field is absent. |
| `field_not_nullable` | The request supplies `null` for a non-nullable blob-create field. |
| `field_empty_after_normalization` | A required string field such as `client_txn_id` is empty after required normalization and trimming. |
| `invalid_byte_size` | `byte_size` is not a non-negative integer. |
| `invalid_sha256_hex` | `sha256_hex`, when present as a string, is not exactly 64 lowercase hexadecimal characters. |
| `unknown_field` | The request includes a top-level member not declared by the blob-create contract. |
| `server_managed_field` | The request attempted to set server-managed blob-slot state such as identifiers, lifecycle fields, accepted-contract echo fields, or upload-target fields. |

`blob_create_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `byte_size_exceeds_limit` | The declared `byte_size` exceeds `limits.object_blobs.max_declared_byte_size` for the current deployment. |


`evidence_attach_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `blob_not_visible` | The supplied `object_blob_id` is missing or is not visible in the target evidence record's incident. Cross-incident blob identifiers MUST use this reason rather than revealing the foreign blob. |
| `blob_pending` | The supplied blob is still pending and cannot be attached yet without a successful observed upload. |
| `blob_failed` | The supplied blob is in terminal `failed` state and cannot be attached. |
| `blob_quarantined` | The supplied blob is quarantined and cannot be attached. |
| `accepted_contract_mismatch` | Observed uploaded bytes do not match the accepted blob contract, including declared size or expected SHA-256 mismatch. |
| `evidence_quarantined` | The target evidence record is quarantined and cannot accept a new blob attachment. |
| `evidence_inconsistent` | The target evidence or blob state is inconsistent and must fail closed until repaired. |


`invalid_enterprise_auth_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The enterprise-auth request body is not a JSON object. |
| `field_not_nullable` | The request supplies `null` for a non-nullable enterprise-auth member. |
| `unknown_field` | The request includes a top-level member not declared by the enterprise-auth route contract. |
| `return_to_not_allowed` | `return_to` is not a same-origin relative-path reference allowed by the current profile. |

`enterprise_auth_transaction_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `not_found` | No current bound auth transaction exists for the callback or ACS request. |
| `expired` | The bound auth transaction exists but its expiry time has passed. |
| `already_used` | The bound auth transaction was already consumed by a prior successful completion. |
| `provider_mismatch` | The callback or ACS request targeted a different provider than the provider bound into the auth transaction. |
| `browser_binding_mismatch` | The callback or ACS request no longer matches the browser-binding context captured at initiation time. |

`provider_response_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `missing_required_field` | The provider response omitted a field required by the active protocol contract. |
| `state_mismatch` | The returned OIDC `state` does not match the bound auth transaction. |
| `relay_state_mismatch` | The returned SAML `RelayState` does not match the bound auth transaction. |
| `nonce_mismatch` | The returned OIDC identity data does not satisfy the bound `nonce`. |
| `code_exchange_failed` | OIDC code exchange or equivalent provider-token retrieval failed closed. |
| `issuer_mismatch` | The provider response issuer does not match the configured provider contract. |
| `audience_mismatch` | The provider response audience does not match the configured service-provider contract. |
| `signature_invalid` | The provider response signature or equivalent authenticity proof did not validate. |
| `assertion_expired` | The provider response or assertion expired before successful completion. |

`provider_identity_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `subject_missing` | The validated provider response did not yield one stable authoritative `provider_subject` under the configured mapping contract. |
| `no_linked_user` | No active local user is linked to the derived `(provider_id, provider_subject)` tuple. |
| `ambiguous_link` | More than one active local user mapping would satisfy the derived provider-backed identity. |
| `inactive_user` | The derived provider-backed identity maps to a local user whose account is inactive. |

`auth_binding_conflict` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `provider_subject_in_use` | Another active binding already uses the requested `(provider_id, provider_subject)` tuple. |
| `provider_already_linked_for_user` | The addressed user already has one active binding for the requested provider. |
| `binding_not_active` | The addressed enterprise binding is retired and cannot be rotated or retired except by exact idempotent replay of the original success. |

`invalid_view_query` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `unknown_filter_field` | `field_key` is not declared filterable for the active `view_schema_id`. |
| `operator_not_allowed` | `op` is not allowed for that field's declared filter class. |
| `invalid_filter_operand` | `arg` is malformed, empty after normalization, contradictory, or otherwise invalid for the selected `op`. |
| `empty_values_after_normalization` | A set-like filter operand normalizes or deduplicates to zero remaining values. |
| `empty_full_text_after_tokenization` | A `full_text.query` normalizes and tokenizes to zero remaining query tokens. |
| `duplicate_filter_field` | The request contains more than one normalized filter entry for the same `field_key`. |
| `filter_count_exceeded` | The request contains more than `16` `filters[]` entries before duplicate-field rejection or operand normalization. |
| `invalid_sort_entry` | A `sort[]` entry is not a JSON object with exactly `field_key` and `direction`, uses an invalid `direction`, or otherwise fails per-entry validation. |
| `duplicate_sort_field` | The request contains more than one normalized `sort[]` entry for the same `field_key`. |
| `sort_field_not_allowed` | `field_key` is not declared in the active `view_schema_id`'s `sort_fields`. |
| `sort_count_exceeded` | The request contains more than `8` `sort[]` entries before duplicate-field rejection, per-entry normalization, or default-sort tail expansion. |
| `invalid_group_by` | `group_by` is malformed, uses `null`, or otherwise fails scalar validation. |
| `group_by_not_allowed` | `group_by` is not one of the grouping keys declared by the active `view_schema_id`. |
| `invalid_limit` | The request supplies `limit` with a non-integer JSON type, a value less than `1`, a value greater than `500`, or an unsupported page-size alias such as `page`, `offset`, `block_size`, or `page_size`. |
| `cursor_query_mismatch` | The supplied `cursor_token` does not match the current normalized view-query contract, including the effective `limit`. |
| `cursor_snapshot_unavailable` | Reserved for future explicit snapshot-backed route families when the supplied `cursor_token` is well-formed but the bound snapshot runtime state is no longer available. Restart the route without `cursor_token` to obtain current live results. |

`invalid_pagination_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `invalid_limit` | The request supplies `limit` with a non-integer JSON type, a value less than `1`, a value greater than `500`, or an unsupported pagination alias such as `page`, `offset`, `block_size`, or `page_size`. |
| `cursor_query_mismatch` | The supplied `cursor_token` does not match the current normalized route contract, including any bound route-scoping identifier, normalized sort or filter or grouping contract when present, or the effective `limit`. |
| `cursor_snapshot_unavailable` | Reserved for future explicit snapshot-backed route families when the supplied `cursor_token` is well-formed but the bound snapshot runtime state is no longer available. Restart the route without `cursor_token` to obtain current live results for that route. |
| `pagination_not_supported` | The addressed route is not declared pageable and therefore rejects `limit`, `cursor_token`, and pagination aliases. |

`credential_bootstrap_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `expired` | The bootstrap token exists but its expiry time has passed. |
| `consumed` | The bootstrap token was already consumed by a prior successful TOTP-completion flow. |
| `superseded` | A newer bootstrap token for the same user or a later administrative reset superseded the supplied token. |
| `subject_mismatch` | The bootstrap token is bound to a different internal user than the route target or pending enrollment. |
| `not_allowed_for_route` | The bootstrap token was presented to a route outside the allowed TOTP-setup family. |

`totp_setup_not_pending` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `not_found` | No current pending TOTP enrollment exists for the supplied `enrollment_id`. |
| `expired` | The addressed pending TOTP enrollment exists but its expiry time has passed. |
| `consumed` | The addressed pending TOTP enrollment was already completed successfully and cannot be completed again. |

`merge_precondition_failed` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `same_record` | The supplied survivor and loser identify the same record. |
| `different_incident` | The two records do not belong to the same incident. |
| `record_type_mismatch` | The two records do not have the same `record_type`. |
| `unsupported_record_type` | The requested `record_type` is not mergeable through the base-profile public route. |
| `survivor_not_mergeable` | The nominated survivor is deleted, already merged away, or otherwise not eligible to survive the merge. |
| `loser_not_mergeable` | The nominated loser is deleted, already merged away, or otherwise not eligible to lose the merge. |
| `carry_forward_identifier_collision` | A loser-side `exact_match_reuse` value could not be preserved on the survivor because it would create an active exact match with a third same-incident record; `error.details` MUST include `identifier_class`, `normalized_value`, and `blocking_record_id`. |

`invalid_rollback_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The rollback request body is not a JSON object. |
| `missing_required_field` | A required rollback request field or selector member is absent. |
| `unknown_field` | The request includes a top-level or target member not declared by the rollback route contract. |
| `invalid_base_row_version` | `base_row_version` is not a positive integer. |
| `invalid_value` | A rollback request member has the wrong JSON type, is empty where non-empty text is required, or otherwise violates the declared selector shape. |
| `target_not_object` | `target` is not a JSON object. |
| `unsupported_target_kind` | `target.kind` is outside the closed rollback target vocabulary. |

`rollback_precondition_failed` `error.details.reason_code` values:

| `reason_code` | Canonical meaning | Requirement ID | Profiles | Verified by |
| --- | --- | --- | --- | --- |
| `target_not_reversible` | The selected history item exists but is not reversible through the requested rollback scope. |  |  |  |
| `entry_requires_change_set` | The selected logical history item belongs to a multi-target or destructive change that MUST be reversed as a whole `change_set`. | REQ-01-239 | base | AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231 |
| `dependent_later_changes` | Later committed changes touched the same mutation target, or otherwise make isolated reversal ambiguous. |  |  |  |
| `stale_target` | The selected historical target exists but is no longer a legal rollback point for current authoritative state because a later reversal or equivalent committed change already superseded it. |  |  |  |

`evidence_access_unavailable` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `no_visible_blob` | The evidence record is visible, but no linked blob is currently available for ordinary preview or download. |
| `blob_pending` | A linked blob slot exists, but upload finalization or availability is not complete yet. |
| `blob_failed` | The linked blob slot reached terminal `failed` state and cannot currently serve preview or download. |
| `blob_missing` | The evidence metadata points at blob content that cannot currently be located or opened. |
| `evidence_quarantined` | The evidence lifecycle or linked blob state blocks ordinary preview and download. |
| `evidence_inconsistent` | Evidence lifecycle state and linked blob state disagree in a way that intentionally fails closed until repaired. |
| `unsupported_preview` | The evidence is otherwise visible and downloadable when available, but the base-profile safe preview contract does not allow the requested preview representation. |
| `preview_payload_too_large` | The evidence is otherwise visible and downloadable when available, but the current payload exceeds the configured preview-size ceiling for the requested preview contract. |

`/ws/v1/` `session_revoked.payload.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `session_expired` | The authenticated session ended because idle or absolute expiry was reached. |
| `session_revoked` | The current session was explicitly logged out or otherwise deployment-revoked. |
| `incident_access_revoked` | The session remains otherwise valid but no longer authorizes the subscribed incident. |
| `concurrency_limit` | The session was revoked because a newer login exceeded the concurrent-session cap. |


`invalid_import_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The request metadata is not a JSON object. |
| `missing_required_field` | A required import-route field is absent. |
| `field_not_nullable` | The request supplies `null` for a non-nullable import-route field. |
| `unknown_field` | The request includes a top-level member not declared by the import-route contract. |
| `unsupported_upload_envelope` | The request is not `multipart/form-data` with a required `boundary`, or otherwise uses an unsupported upload envelope. |
| `missing_required_part` | One of the required multipart parts `metadata` or `file` is absent. |
| `duplicate_part` | The multipart envelope contains more than one `metadata` part or more than one `file` part. |
| `unexpected_part` | The multipart envelope contains a part name outside the closed two-part contract. |
| `invalid_part_content_type` | The addressed multipart part has a content type outside the allowed set for that part and route. |
| `invalid_metadata_encoding` | The `metadata` part is not valid UTF-8, includes a BOM, or declares an unsupported JSON charset. |
| `malformed_metadata_json` | The `metadata` part cannot be parsed as JSON or contains duplicate object member names. |
| `invalid_row_reference` | `header_row_ref` or `data_start_row_ref` is not a positive 1-based row coordinate within `source_rect_a1`. |
| `invalid_selected_unit_ids` | `selected_unit_ids[]` is empty, contains duplicates, or references units outside the addressed session. |
| `unsupported_assistant_profile` | `assistant_profile` is not `phase2_workbook_import_v1` in the current profile. |
| `invalid_source_columns` | `source_columns[]` is missing, empty, not exhaustive over the discovered source columns, uses duplicate or non-contiguous ordinals, or otherwise violates the per-column mapping contract. |
| `invalid_unknown_column_policy` | `unknown_column_policy` is outside the closed current-profile registry or is not legal for the addressed target view. |
| `invalid_transform` | `transform_id` or `transform_options` is outside the closed current-profile mapping-transform contract. |
| `invalid_empty_value_policy` | `empty_value_policy` is outside the closed current-profile registry or is not legal for the addressed target field. |
| `duplicate_target_field` | More than one mapped source column targets the same non-null `field_key`. |

`import_state_conflict` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `session_applying` | The addressed import session is already applying and therefore rejects the requested control-plane action. |
| `session_terminal` | The addressed import session is already in a terminal durable state and therefore rejects the requested control-plane action. |
| `unit_applying` | The addressed import unit is already applying and therefore rejects the requested control-plane action. |
| `unit_terminal` | The addressed import unit is already in a terminal durable state and therefore rejects the requested control-plane action. |

`import_source_unsupported` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `encrypted_or_unparseable_workbook` | The uploaded workbook cannot be parsed by the current profile. |
| `unsupported_named_range` | The addressed named range is dynamic, multi-area, or otherwise unsupported by the current profile. |
| `formula_cached_value_missing` | A mapped formula cell lacks an inert cached value and therefore cannot enter `ready`. |

`import_source_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `csv_source_too_large` | The uploaded CSV source bytes exceed `limits.imports.max_csv_source_bytes`. |
| `xlsx_source_too_large` | The uploaded XLSX source bytes exceed `limits.imports.max_xlsx_source_bytes`. |
| `import_rows_exceeded` | The parsed selected unit exceeds `limits.imports.max_rows`. |
| `import_columns_exceeded` | The parsed selected unit exceeds `limits.imports.max_columns`. |
| `import_cells_exceeded` | The parsed selected unit exceeds `limits.imports.max_cells`. |
| `archive_extracted_bytes_exceeded` | The extracted regular-file byte total exceeds the applicable extracted-bytes ceiling. |
| `archive_compression_ratio_exceeded` | The extracted regular-file byte total exceeds `compressed_bytes * limits.archives.max_compression_ratio`. |
| `archive_member_count_exceeded` | The extracted regular-file member count exceeds `limits.archives.max_members`. |

`import_apply_blocked` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `overlapping_units` | The selected `import_unit` rectangles overlap and therefore cannot be jointly applied. |
| `duplicate_apply_blocked` | The same `(import_unit_id, mapping_fingerprint, incident_id)` tuple was already applied and re-import was not explicitly requested. |
| `unit_not_ready` | One or more selected units are not yet in `ready` state. |

`invalid_snapshot_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The request body is not a JSON object. |
| `missing_required_field` | A required snapshot-create field is absent. |
| `field_not_nullable` | The request supplies `null` for a non-nullable snapshot-create field. |
| `unknown_field` | The request includes a top-level member not declared by the snapshot-create contract. |
| `invalid_source_boundary` | `source_change_set_high_watermark`, when supplied, is not a valid committed source-boundary reference for the addressed incident. |

`invalid_release_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The request body is not a JSON object. |
| `missing_required_field` | A required release-create field is absent. |
| `field_not_nullable` | The request supplies `null` for a non-nullable release-create field. |
| `unknown_field` | The request includes a top-level member not declared by the release-create contract. |
| `invalid_release_scope` | The supplied `release_scope` is not one of the current-profile release-scope tokens. |
| `version_selector_required` | The request omitted `template_version` or `redaction_profile_version`, or otherwise attempted implicit latest-version selection. |

`release_render_failed` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `missing_redaction_rule` | A renderable field or block lacked an applicable redaction rule for the chosen `release_scope`. |
| `undeclared_template_binding` | The selected template referenced an undeclared export-model binding. |
| `missing_required_field` | The selected template required a field not present in the frozen export model. |

`release_state_conflict` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `not_approved` | The requested publish action targeted a release that is not currently `approved`. |
| `already_approved` | The requested approve action targeted a release already in `approved` state. |
| `already_published` | The requested approve or publish action targeted a release already in `published` state. |
| `already_invalidated` | The requested approve, publish, or invalidate action targeted a release already in `invalidated` state. |

`release_approval_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `approval_requirements_not_met` | The caller or current artifact tuple cannot contribute a valid approval for the requested approve action while the release remains eligible for approval, including duplicate or otherwise non-contributing approval attempts while `release_state='pending_approval'`. |

`invalid_reference_pack_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The request metadata is not a JSON object. |
| `missing_required_field` | A required reference-pack route field is absent. |
| `field_not_nullable` | The request supplies `null` for a non-nullable reference-pack route field. |
| `unknown_field` | The request includes a top-level member not declared by the reference-pack route contract. |
| `unsupported_upload_envelope` | The request is not `multipart/form-data` with a required `boundary`, or otherwise uses an unsupported upload envelope. |
| `missing_required_part` | One of the required multipart parts `metadata` or `file` is absent. |
| `duplicate_part` | The multipart envelope contains more than one `metadata` part or more than one `file` part. |
| `unexpected_part` | The multipart envelope contains a part name outside the closed two-part contract. |
| `invalid_part_content_type` | The addressed multipart part has a content type outside the allowed set for that part and route. |
| `invalid_metadata_encoding` | The `metadata` part is not valid UTF-8, includes a BOM, or declares an unsupported JSON charset. |
| `malformed_metadata_json` | The `metadata` part cannot be parsed as JSON or contains duplicate object member names. |
| `invalid_activation_policy` | `activation_policy` is present but is not the exact current-profile scalar request form. |
| `pack_version_required` | The requested action requires an exact `pack_version` rather than implicit latest-version selection. |
| `auto_activation_not_supported` | The request attempted auto-activation instead of the staged-only current-profile import contract. |
| `invalid_pack_keys` | `pack_keys[]` is not an array of exact visible `pack_key` strings, or it contains one or more unknown, non-visible, or non-string members. |
| `empty_pack_keys` | `pack_keys[]` is present and empty. |

`reference_pack_verification_failed` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `checksum_mismatch` | One or more declared checksums did not match the supplied bundle or extracted content. |
| `signature_mismatch` | Signature verification failed for the supplied bundle. |
| `missing_integrity_metadata` | Required integrity metadata is absent from the supplied bundle. |
| `contract_incompatible` | The pack contract or schema version is not compatible with the running application. |
| `path_traversal` | One or more archive members attempt to escape the staging root. |
| `disallowed_content` | The bundle contains active or otherwise disallowed content. |
| `payload_missing` | Required pack payload content is missing at verification time. |
| `archive_extracted_bytes_exceeded` | The extracted regular-file byte total exceeds `limits.reference_packs.max_extracted_bytes`. |
| `archive_compression_ratio_exceeded` | The extracted regular-file byte total exceeds `compressed_bytes * limits.archives.max_compression_ratio`. |
| `archive_member_count_exceeded` | The extracted regular-file member count exceeds `limits.archives.max_members`. |

`reference_pack_activation_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `already_active` | The addressed pack version is already the active version for its `pack_key`. |
| `not_verified_available` | The addressed pack version is not currently in `verified_available` condition. |

`reference_pack_state_conflict` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `already_disabled` | The requested disable action targeted a version already in `disabled` condition. |
| `not_disableable` | The requested disable action targeted a version whose current durable condition does not allow disable. |
| `verification_pending` | The requested reverify action targeted a version still in `staged` condition and awaiting initial verification. |

`invalid_incident_bundle_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The request metadata is not a JSON object. |
| `missing_required_field` | A required incident-bundle route field is absent. |
| `field_not_nullable` | The request supplies `null` for a non-nullable incident-bundle route field. |
| `unknown_field` | The request includes a top-level member not declared by the incident-bundle route contract. |
| `unsupported_upload_envelope` | The request is not `multipart/form-data` with a required `boundary`, or otherwise uses an unsupported upload envelope. |
| `missing_required_part` | One of the required multipart parts `metadata` or `file` is absent. |
| `duplicate_part` | The multipart envelope contains more than one `metadata` part or more than one `file` part. |
| `unexpected_part` | The multipart envelope contains a part name outside the closed two-part contract. |
| `invalid_part_content_type` | The addressed multipart part has a content type outside the allowed set for that part and route. |
| `invalid_metadata_encoding` | The `metadata` part is not valid UTF-8, includes a BOM, or declares an unsupported JSON charset. |
| `malformed_metadata_json` | The `metadata` part cannot be parsed as JSON or contains duplicate object member names. |
| `invalid_reference_pack_mode` | The supplied `reference_pack_mode` is not one of the current-profile incident-bundle export tokens. |
| `invalid_optional_sections` | `optional_sections[]` is not an array of current-profile optional-section tokens, or it contains one or more unknown or non-string members. |
| `invalid_required_capabilities` | `required_capabilities[]` is not an array of current-profile capability tokens, or it contains one or more unknown or non-string members. |
| `history_mode_not_supported` | The request attempted to set `history_mode` to anything other than the fixed current-profile `full` behavior. |
| `blob_mode_not_supported` | The request attempted to set `blob_mode` to anything other than the fixed current-profile `full` behavior. |

`incident_bundle_export_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `missing_required_file` | One or more required structured bundle files could not be materialized. |
| `missing_required_blob` | One or more required blob bytes could not be materialized for export. |

`incident_bundle_import_rejected` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `invalid_member_path` | A bundle member path is absolute, traverses parents, or otherwise violates the logical bundle path contract. |
| `unsupported_member_type` | The bundle contains a member type outside regular files and directories. |
| `checksum_mismatch` | A declared structured-file or bundle checksum did not match the supplied bytes. |
| `signature_mismatch` | Bundle signature verification failed where supported or required. |
| `blob_hash_mismatch` | One or more blob bytes did not match the required `blobs/sha256/<sha256-lower-hex>` path digest. |
| `duplicate_incident_id` | The target deployment already contains the exported `incident_id`. |
| `unsupported_required_capability` | The bundle requires a capability the target deployment does not implement. |
| `remote_fetch_required` | Import would require a remote fetch, which the current profile forbids. |
| `archive_extracted_bytes_exceeded` | The extracted regular-file byte total exceeds `limits.incident_bundles.max_extracted_bytes`. |
| `archive_compression_ratio_exceeded` | The extracted regular-file byte total exceeds `compressed_bytes * limits.archives.max_compression_ratio`. |
| `archive_member_count_exceeded` | The extracted regular-file member count exceeds `limits.archives.max_members`. |

#### 3.3.7 Pagination and cursor contract

**REQ-01-240**
Any public list or query route that MAY exceed one response page MUST use opaque cursor pagination under this subsection. In `/api/v1/`, GET routes MUST accept only `limit` and `cursor_token` as query members and POST query routes MUST accept only `limit` and `cursor_token` as JSON-body members. `limit` MUST be an integer in the inclusive range `1..500`. When `cursor_token` is absent, omitted `limit` MUST mean `100`. When `cursor_token` is present, omitted `limit` MUST mean reuse the cursor-bound effective `limit` from the request that produced that cursor. The `/api/v1/` contract MUST reject `page`, `offset`, `page_size`, `block_size`, or any other pagination alias rather than silently accepting or translating it.
Profiles: base
Verified by: AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-238, AC-239, AC-240

**REQ-01-241**
A `cursor_token` MUST be bound to the authenticated actor, route family, every route-scoping identifier present for that route, the normalized effective `sort[]`, the normalized `filters[]`, the optional normalized `group_by`, and the effective `limit` when the route defines a view-query contract. This includes binding history cursors to `record_id`, membership and saved-view cursors to `incident_id`, and workbook-query cursors to `incident_id`, `view_schema_id`, and the normalized applied view-query contract. The server MUST reject a cursor that is replayed against a different bound route contract, including a different effective `limit`, rather than reinterpret it.
Profiles: base
Verified by: AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-239

**REQ-01-242**
The envelope for paged responses MUST include `meta.paging.limit`, `meta.paging.has_more`, and `meta.paging.next_cursor`. `meta.paging.limit` MUST equal the effective page size bound to the current page. When another page is available, `meta.paging.has_more=true` and `meta.paging.next_cursor` MUST be a non-null opaque cursor. A terminal page, including a first page with zero matching rows, MUST use `meta.paging.has_more=false` and `meta.paging.next_cursor=null`. Non-view routes that are not declared pageable in their owner section MUST reject `limit`, `cursor_token`, and any pagination alias with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code=pagination_not_supported` rather than ignoring them. Clients MUST treat `meta.paging.has_more` and `meta.paging.next_cursor` as the authoritative continuation contract and MUST NOT infer terminal state from `rows.length < meta.paging.limit` alone. Hot workbook views MUST NOT require deep `OFFSET` pagination.
Profiles: base
Verified by: AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-238, AC-239, AC-241, AC-242

**REQ-01-554**
The current base profile cursor-continuation mode for live operational list and query routes is `live_authorized_keyset`. Page 2 and every later page in a live cursor chain MUST re-derive the caller's current session validity, route authorization, and route-scoped visibility before returning rows. The cursor token MUST remain opaque to clients and MUST cryptographically protect the bound route contract and server-owned continuation position from client tampering. A cursor MUST NOT preserve access after session expiry, account revocation, loss of incident membership, loss of route visibility, or any other authorization change that would make the row or route unavailable on a fresh request.
Profiles: base
Verified by: AC-231, AC-372, AC-373, AC-374, AC-375

**REQ-01-555**
For one live cursor chain, the server MUST use a deterministic continuation position derived from the declared route ordering and route contract. Later inserts, deletes, restores, sort-key edits, grouping-key edits, filter-relevant edits, membership changes, or equivalent committed mutations MAY affect later pages because live routes re-evaluate current state. The current profile requires these observable outcomes:

| Intervening committed change after page 1 | Continuation with `meta.paging.next_cursor` from the existing chain | Fresh request without `cursor_token` |
| --- | --- | --- |
| Matching insert or restore of a row that was absent when the cursor was issued | MAY appear only if it falls after the current continuation position under the live route ordering | MUST reflect the row if the live current route contract matches it |
| Delete or authorization removal of a row that would otherwise have appeared later in the chain | MUST NOT return the row after it is no longer currently visible to the caller | MUST reflect current live visibility |
| Sort-, group-, or filter-relevant edit of a row that was present when the cursor was issued | MUST return only the current row payload if the row is still visible and still falls after the continuation position; otherwise it MAY move outside the remaining chain | MUST reflect current live order and grouping |

Profiles: base
Verified by: AC-231, AC-372, AC-373, AC-374

**REQ-01-556**
Across one complete live cursor chain, the server SHOULD avoid duplicate rows where the route's deterministic ordering and continuation position make that possible. The server MUST NOT return a row that no longer matches current authorization or route visibility merely to preserve an earlier cursor-chain membership. Hot workbook views MUST use stable keyset or viewport/block retrieval rather than unbounded full-result materialization or deep offset scans.
Profiles: base
Verified by: AC-231, AC-372, AC-373

**REQ-01-557**
When a live pageable route returns full row objects or equivalent current-row payload, the serialized row state, including `row_version` when present, MUST reflect authoritative state at fetch time after current authorization succeeds. Reading a row through a cursor does not reserve later write success; ordinary later writes still use current live optimistic-concurrency and authorization rules.
Profiles: base
Verified by: AC-231, AC-373, AC-374

**REQ-01-558**
Live collaboration and cursor continuation are intentionally compatible live views over current state. Replayable `record_changed` messages, `invalidate`, `remove`, presence updates, and other live collaboration events continue to report current live state. Cursor continuation MUST NOT use retained row payloads as an authorization cache, and clients MUST tolerate that later pages can reflect intervening authorized live changes.
Profiles: base
Verified by: AC-231, AC-374

**REQ-01-559**
The server MAY reject a cursor when its signature, version, route binding, actor binding, route-scoping identifiers, normalized query contract, effective limit, or server-owned continuation position is invalid or no longer supported. View-query routes MUST fail with `400`, `error.code='invalid_view_query'`, and the most specific pagination reason code. Other pageable public routes MUST fail with `400`, `error.code='invalid_pagination_request'`, and the most specific pagination reason code. If a future explicit snapshot-backed route family is added, unavailable snapshot runtime state MUST fail closed with `cursor_snapshot_unavailable`; live operational routes MUST NOT depend on retained in-memory row snapshots for continuation.
Profiles: base
Verified by: AC-231, AC-375

**REQ-01-560**
Immutable `snapshot_stable` continuation is reserved for explicit immutable snapshot/reporting artifacts or a future opt-in route family. It MUST NOT be used as the default continuation behavior for live operational list, workbook-query, membership, user, or history routes because those routes must re-derive authorization at request time.
Profiles: base
Verified by: AC-231, AC-372, AC-373, AC-374, AC-375

#### 3.3.8 Evidence and blob routes

**REQ-01-243**
`POST /api/v1/object-blobs` MUST accept only a JSON object request body and MUST provision exactly one incident-scoped pending blob slot for one intended upload. The request MUST accept required `incident_id`, required `client_txn_id`, and required `byte_size`. It MAY accept optional `filename_hint`, optional `content_type_hint`, and optional `sha256_hex`; each optional member MAY be omitted or set to JSON `null`. `byte_size` MUST be an integer in `0..limits.object_blobs.max_declared_byte_size`, inclusive. This route MUST create only the blob slot; it MUST NOT create or mutate evidence records, record links, preview state, release state, or workflow objects. The route MUST reject row identifiers, evidence identifiers, preview intents, release intents, workflow objects, unknown top-level members, and server-managed blob fields. If the body is not a JSON object, omits a required member, supplies `null` for a non-nullable member, violates a field validation rule in this subsection, or attempts to set server-managed state, the server MUST fail with `400` and `error.code = invalid_blob_create_request`. When `byte_size` exceeds `limits.object_blobs.max_declared_byte_size`, the server MUST fail before slot creation with `413`, `error.code = blob_create_rejected`, `error.details.reason_code = byte_size_exceeds_limit`, `error.details.requested_byte_size`, and `error.details.configured_limit_bytes`. That rejection path MUST create no `object_blob_id`, no `upload_target`, and no pending blob-slot state. When the failure is attributable to one request member, `error.details.field` MUST identify that top-level member. `error.details.reason_code` MUST use the registry in §3.3.6.2.
Profiles: base
Verified by: AC-015, AC-016, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231, AC-321

**REQ-01-244**
For idempotency comparison and response echo, the server MUST normalize the blob-create contract as follows:

- `filename_hint` and `content_type_hint` are advisory metadata only. They MUST NOT determine storage keys, authorization, portability layout, preview allowlisting, or release posture.
- omitted and explicit JSON `null` compare equal for `filename_hint`, `content_type_hint`, and `sha256_hex`,
- `filename_hint` and `content_type_hint` MUST be trimmed and an empty result MUST compare equal to omission,
- `sha256_hex`, when present as a string, MUST be exactly 64 lowercase hexadecimal characters and MUST NOT be silently case-folded or repaired,
- `byte_size` is a declared upload-contract input,
- `sha256_hex`, when present, is an expected integrity assertion for later finalization comparison and is not authoritative stored hash metadata by itself.

A successful create response MUST include `incident_id`, `object_blob_id`, `upload_state`, `target_expires_at`, `pending_expires_at`, the short-lived `upload_target`, and `accepted_contract`. `accepted_contract` MUST echo the server-accepted normalized values for `incident_id`, `byte_size`, `filename_hint`, `content_type_hint`, and `sha256_hex`. Omitted optional members MUST be serialized as explicit `null` inside `accepted_contract` rather than by field omission.
Profiles: base
Verified by: AC-015, AC-016, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-01-245**
Blob-slot creation idempotency MUST be keyed by `(actor_user_id, incident_id, client_txn_id)`. Normalized request comparison for this route MUST include `byte_size`, normalized `filename_hint`, normalized `content_type_hint`, and normalized `sha256_hex`, with omission and explicit JSON `null` treated as equal for the optional members. A first-time successful create MUST return `201 Created`. If the same authenticated actor replays the same normalized request with the same key, the server MUST return `200 OK` with the original slot payload, including the original `object_blob_id`, `target_expires_at`, `pending_expires_at`, and `accepted_contract`, and MUST create no second pending slot. If the same actor reuses that key with a different normalized request, the server MUST fail with `409` and `error.code = client_txn_conflict`. `error.details` MUST include at least `client_txn_id`. A request rejected under REQ-01-243 because `byte_size` exceeds the configured ceiling MUST create no pending slot and MUST NOT replay as a successful slot response. Blob finalization MUST occur only through an explicit follow-on call. Binding an uploaded blob to incident-visible evidence MUST either:

- attach the returned `object_blob_id` to an existing evidence record through `POST /api/v1/evidence-records/{record_id}/attach-blob`, or
- create a new evidence record through the normal view or record-creation path using that `object_blob_id` as declared input.

When finalization targets an existing evidence record through `POST /api/v1/evidence-records/{record_id}/attach-blob`, the route MUST be record-scoped rather than blob-scoped. It MUST accept only a JSON object with required `object_blob_id`, required `base_row_version`, and required `client_txn_id`. The base profile defines no optional top-level members for this route. A non-object body, a missing required member, a supplied `null` for one of those required members, or any unknown top-level member MUST fail with `400` and `error.code = invalid_mutation_payload`.

Attach idempotency for this route MUST be keyed by `(actor_user_id, record_id, client_txn_id)`. Normalized request comparison for this route MUST include exact `object_blob_id` and exact `base_row_version`. Because the base profile defines no nullable request members for this route, omission-versus-`null` equivalence does not apply. A first-time successful attach MUST return `200 OK`. If the same authenticated actor replays the same normalized attach request with the same key, the server MUST return `200 OK` with the original committed attach result and MUST create no second attach or replacement transition. If the same actor reuses that key with a different normalized attach request, the server MUST fail with `409` and `error.code = client_txn_conflict`. If the current evidence-row version differs from `base_row_version`, the route MUST fail with `409` and `error.code = row_version_conflict`.

Fresh attach requests with a new `client_txn_id` MUST still enforce the blob and evidence lifecycle bridge owned by Core 02 §13 and Core 03 §8. A pending, failed, missing, quarantined, cross-incident, contract-mismatched, or otherwise non-attachable blob MUST fail closed with `error.code='evidence_attach_rejected'` and a reason from §3.3.6.2 rather than partially mutating evidence state. Cross-incident or missing blob identifiers MUST use `reason_code='blob_not_visible'`. A successful attach response MUST use the common success envelope and return `data.view_schema_id='cartulary.view.evidence.v1'`, `data.change_set_id`, `data.object_blob_id`, and `data.row`, where `data.row` is exactly `view_row_v1` for the addressed evidence row. Dependent Timeline, Host, Identity, or other rows affected by that attach MUST refresh only through ordinary replayable `record_changed` `patch` or `invalidate` messages rather than extra inline row payloads on this route.
Profiles: base
Verified by: AC-015, AC-016, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231, AC-321

**REQ-01-246**
In the base profile, the upload target MUST expire 60 minutes after issuance and the pending blob slot MUST expire 24 hours after blob-slot creation. These timers MUST remain separate: target expiry governs upload-target use, while pending-slot expiry governs later finalization eligibility and cleanup. The base profile MUST treat a pending blob slot as a single-upload lease bound to one accepted create contract. If the upload target expires before successful upload, or if idempotent replay returns an already expired slot, obtaining a fresh target MUST use a fresh `POST /api/v1/object-blobs` call with a new `client_txn_id`. Idempotent replay of an expired slot MUST return that same expired slot rather than refreshing the original target. The base profile MUST NOT require same-slot upload-target refresh, same-slot lease renewal, or resumable upload semantics.
Profiles: base
Verified by: AC-015, AC-016, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-01-247**
Preview and download MUST be issued only through `POST /api/v1/evidence-records/{record_id}/preview-handle` and `POST /api/v1/evidence-records/{record_id}/download-handle`, and redeemed only through `GET /api/v1/evidence-handles/{handle_token}`. These routes MUST return short-lived authorization-checked handles. They MUST NOT expose long-lived object-store credentials, bypass incident membership checks, or treat a `pending` blob slot as attached evidence.
Profiles: base
Verified by: AC-015, AC-016, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231, AC-251, AC-252, AC-253, AC-254, AC-255

##### 3.3.8.1 Evidence-access handle owner pointer

This subsection declares the base route family only. Core 01 §16 is the primary owner for the preview-handle, download-handle, and redeem contract, including issuance request shape, response members, redeem semantics, lifetime, revocation, filename and disposition rules, and route-specific error use.

#### 3.3.9 Background-job routes

**REQ-01-248**
Routes that start long-running operations MUST return `202 Accepted` with the common success envelope from §3.3.6 and `data` equal to the canonical job resource defined in §3.3.9.1. The initial job resource returned by an initiating route MUST use `status` equal to `queued` or `running`, and `status_route` MUST be the same-origin path `/api/v1/jobs/{job_id}` for that resource.
Profiles: base
Verified by: AC-046, AC-129, AC-231, AC-257

**REQ-01-249**
`GET /api/v1/jobs/{job_id}` MUST return the canonical job resource defined in §3.3.9.1. `POST /api/v1/jobs/{job_id}/cancel` MUST accept only a JSON object with required `client_txn_id` and optional `reason`, with omitted `reason` and explicit JSON `null` for `reason` comparing equal for normalized request comparison. Cancel idempotency MUST be keyed by `(actor_user_id, job_id, client_txn_id)`. If the same authenticated actor replays the same normalized cancel request with the same key, the server MUST return `200 OK` with the current authoritative job resource and MUST create no second cancel transition. If the same actor reuses that key with a different normalized cancel request, the server MUST fail with `409` and `error.code = client_txn_conflict`. A fresh cancel request against an already-cancel-requested, terminal, or otherwise non-cancelable job MUST fail with `409` and `error.code = job_cancel_rejected` plus the exact `reason_code` defined in §3.3.9.1.
Profiles: base
Verified by: AC-046, AC-129, AC-231, AC-257, AC-258, AC-259, AC-260, AC-261

##### 3.3.9.1 Job resource contract

For REQ-01-248 and REQ-01-249, the canonical public job contract is:

- The canonical HTTP job resource returned by initiating routes, `GET /api/v1/jobs/{job_id}`, and successful cancel responses MUST use the common success envelope from §3.3.6 with `data` equal to one job resource object. That object MUST include `job_id`, `scope`, `status_route`, `status`, `cancelable`, `submitted_by_user_id`, `submitted_at`, `updated_at`, `progress`, `started_at`, `finished_at`, `retained_until`, `result_summary`, and `error_summary`. `started_at`, `finished_at`, `retained_until`, `result_summary`, and `error_summary` are required-but-nullable. The resource MAY include optional `message` for short operator-visible status text.
- `scope` is required and is the authorization and live-update boundary. In `/api/v1/`, the closed `scope.kind` vocabulary is `incident | deployment`. When `scope.kind = incident`, `scope.incident_id` is required. When `scope.kind = deployment`, `scope.incident_id` is forbidden.
- `status` is a closed six-token vocabulary: `queued`, `running`, `cancel_requested`, `succeeded`, `failed`, and `canceled`. The public job shell MUST NOT introduce `completed`, `done`, `warning`, or job-family-specific phase tokens as alternate `status` values.
- The allowed public state transitions are `queued -> running | cancel_requested | failed`, `running -> cancel_requested | succeeded | failed`, and `cancel_requested -> canceled | failed | succeeded`. Terminal states have no outgoing public transitions.
- `cancelable` is required. It MUST be `false` in `cancel_requested`, `succeeded`, `failed`, and `canceled`. It MAY be `false` in `queued` or `running` when the current non-terminal phase does not accept cancellation.
- `submitted_at` and `updated_at` are required timestamps. `started_at` MUST be `null` until work begins. `finished_at` MUST be `null` until the job reaches a terminal state. `retained_until` MUST be `null` until the job reaches a terminal state.
- `progress` MUST be an object of the form `{ completed, total }`, never a bare percentage. `completed` MUST be a non-negative integer and MUST be monotonically non-decreasing for one job resource. `total` MUST be either `null` or a positive integer. Once `total` becomes non-null, it MUST NOT decrease and MUST NOT change unit semantics. When `total` is non-null, `completed` MUST be less than or equal to `total`. On `succeeded`, if `total` is non-null, `completed` MUST equal `total`. Clients MAY derive `floor(100 * completed / total)` when `total` is non-null, but percent is not part of the wire contract and clients MUST render indeterminate progress when `total = null`.
- `result_summary` and `error_summary` are mutually exclusive. On non-terminal states, both MUST be `null`. On `succeeded` and `canceled`, `result_summary` is required and `error_summary` MUST be `null`. On `failed`, `error_summary` is required and `result_summary` MUST be `null`. When `status = canceled`, `result_summary.code` MUST be exactly `job_canceled`.
- `result_summary` MUST be compact and generic: `{ code, message, resource_refs? }`. `result_summary.code` is registry-backed, not opaque. When `status = succeeded`, `result_summary.code` MUST use one of the stable success codes declared by the initiating route family in this document; in the current profile, those success-code registries are defined in §17. `result_summary.message` remains operator-visible text only, and clients MUST NOT branch protocol behavior on its contents.
- `resource_refs[]` is a compact, non-exhaustive navigation summary of durable outputs or newly relevant durable resources, not a deep result payload. The current-profile closed `resource_refs[].kind` vocabulary is exactly `incident`, `import_session`, `snapshot`, `release`, `reference_pack_version`, and `incident_bundle`. Current-profile emissions MUST NOT use `job`, `blob`, `preview_handle`, `download_handle`, `saved_view`, `view_schema`, or free-form family-defined kinds.
- `resource_refs[].route` is the canonical same-origin `GET` path for the referenced durable resource. It MUST begin with `/api/v1/`, MUST use the canonical public read route for that durable resource, and MUST NOT include a query string or fragment. It MUST NOT be a UI-local route, a presigned URL, a preview handle, a download handle, or the job-status route. Clients MAY dereference `route` or resolve the target by `kind` and `id`, but they MUST treat `route` as opaque.
- For `incident`, `import_session`, `snapshot`, `release`, and `incident_bundle`, `resource_refs[].id` MUST equal the existing public identifier for that resource kind. For `reference_pack_version`, `route` is required and `id` MUST equal the exact canonical `route` string.
- Although `route` remains optional in the abstract job shell for forward compatibility, every current-profile `resource_ref` emitted by the route families in §17 MUST include `route`. If more than one current-profile ref is emitted, ordering MUST be deterministic. For `reference_packs_refreshed`, emitted refs MUST sort by `route asc`. Clients MUST ignore unknown future `kind` values rather than fail job rendering, even though current-profile servers are closed to the allowlist above.
- `error_summary` MUST be compact and generic: `{ code, message, retryable, details? }`, where `details` is an optional JSON object. The common job resource MUST NOT carry job-family-specific deep result payloads.
- `POST /api/v1/jobs/{job_id}/cancel` MUST require a JSON object containing required `client_txn_id` and optional `reason`. For idempotency comparison, omitted `reason` and explicit JSON `null` for `reason` compare equal. A cancel request body that is not a JSON object, omits required `client_txn_id`, or includes unknown top-level members MUST fail with `400` and `error.code = invalid_mutation_payload`.
- Cancel idempotency is keyed by `(actor_user_id, job_id, client_txn_id)`. If the same authenticated actor replays the same normalized cancel request with the same key, the server MUST return `200 OK` with the current authoritative job resource and MUST create no second cancel transition. If the same actor reuses that key with a different normalized cancel request, the server MUST fail with `409` and `error.code = client_txn_conflict`.
- A newly accepted cancel request MUST set `cancelable = false` immediately and MUST transition the job through `cancel_requested` before any later `canceled`, `failed`, or `succeeded` terminal state is observed. A conformant server MAY return a terminal state on the successful cancel response when the cancel lost a race to an already committed safe-stop boundary, but it MUST NOT create a second public transition.
- A fresh cancel request against `cancel_requested`, `succeeded`, `failed`, `canceled`, or any non-terminal resource whose current `cancelable = false` MUST fail with `409`, `error.code = job_cancel_rejected`, and `error.details.reason_code` equal to `already_cancel_requested`, `already_terminal`, or `not_cancelable` from §3.3.6.2. A missing, expired, or unauthorized job lookup or cancel request MUST fail with `404` and `error.code = job_not_found`.
- Terminal job resources MUST be retained for at least 7 days. After a terminal transition, `retained_until` is required and MUST be greater than or equal to `finished_at + 7 days`. After `retained_until`, `GET /api/v1/jobs/{job_id}` MAY return `404` with `error.code = job_not_found`. Expiring a terminal job resource MUST NOT delete or mutate durable outputs such as committed incident changes, imports, reports, snapshots, bundles, or blob metadata produced by that job.

#### 3.3.10 WebSocket collaboration stream

**REQ-01-250**
The base-profile public incident-scoped WebSocket subscription route MUST be `GET /ws/v1/incidents/{incident_id}`.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-251**
The live-update stream MUST use a bounded message family rather than a second mutation API. The base message set MUST include:

- session handshake messages and acknowledgements: `hello`, `resume`, `hello_ack`, `resume_ack`,
- incident-scoped presence messages: `presence_snapshot`, `presence_delta`, `presence_update`,
- `record_changed` events,
- incident-scoped `job_progress` events,
- heartbeat messages: `ping`, `pong`,
- terminal `error` or `session_revoked` events.
Profiles: base, snapshot_reporting
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-233

**REQ-01-252**
A replayable `record_changed` event MUST identify the `incident_id`, affected `record_id`, resulting `row_version`, and one or more affected `view_schema_id` entries, each with either deterministic field-key-addressable patch cells, an explicit `invalidate` signal, or an explicit `remove` signal.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-253**
The WebSocket stream MUST authenticate with the same server-managed session contract as the HTTP surface. It MUST NOT broadcast unresolved same-field local drafts, transient pending patches, expand or collapse state for grouped rows, or other client-local UI state as authoritative collaboration events.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

##### 3.3.10.1 v1 collaboration wire contract

**REQ-01-254**
A client MUST send exactly one session-establishment message as the first application message on a connection: `hello` for a fresh connection or `resume` for a reconnect. Until the server accepts one of these messages, it MUST NOT treat the socket as subscribed and MUST NOT emit replayable incident messages on that connection.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-255**
A connected socket is subscribed to exactly one incident, determined by the `{incident_id}` path parameter. The protocol MUST NOT negotiate arbitrary topic subscriptions or a second mutation surface. Authoritative record creation, record mutation, rollback, conflict resolution, blob finalization, and other write actions MUST continue to use the HTTP routes defined in §3.3.3 through §3.3.9.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-256**
Every application-level collaboration message MUST use one JSON envelope:
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

```json
{
  "type": "record_changed",
  "incident_id": "inc_...",
  "event_id": "evt_...",
  "emitted_at": "2026-03-23T20:14:31.418Z",
  "stream_seq": 1842,
  "payload": {}
}
```

The envelope contract is:

- `type`: required closed message-type token,
- `incident_id`: required stable incident identifier matching the route,
- `event_id`: required opaque server-generated event identifier on server-originated messages,
- `emitted_at`: required RFC 3339 UTC timestamp,
- `stream_seq`: required only on replayable server messages and forbidden on client-originated messages and ephemeral server messages,
- `payload`: required object whose members are specific to `type`.

**REQ-01-257**
Within `/ws/v1/`, clients MUST ignore unknown additive members. Breaking changes to required message members, required message types, or message semantics MUST use a new major version root as defined in §3.3.1.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-258**
The minimum client-to-server message set MUST be:

- `hello`, whose `payload` MUST include `client_instance_id` and current `presence`,
- `resume`, whose `payload` MUST include `client_instance_id`, `resume_token`, `last_seen_stream_seq`, and current `presence`,
- `presence_update`, whose `payload` MUST include current `presence`,
- `pong`, whose `payload` MAY be empty.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-259**
`client_instance_id` MUST remain stable for the lifetime of one browser tab or other local client instance. `resume_token` MUST be opaque to the client, MUST be a replay token rather than an authentication token, and MUST be bound by the server to the authenticated session, `incident_id`, and `client_instance_id`. A `resume_token` MUST NOT authenticate HTTP routes or establish session identity independently of the underlying authenticated session. A `resume_token` MUST expire no later than the earlier of the configured replay window and the underlying session expiry. `last_seen_stream_seq` MUST be the highest replayable `stream_seq` the client has already applied.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-260**
The `presence` object MUST be shaped as:

- `sheet_ref`: required stable workbook-surface reference,
- optional `record_id`: stable target row identifier when focus is on a materialized row,
- optional `field_key`: stable target field identifier,
- `mode`: required one of `viewing`, `editing`, or `idle`.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-261**
`sheet_ref` MUST address workbook surfaces by stable identifier rather than visible label and MUST be one of:
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

```json
{
  "kind": "view_schema",
  "id": "cartulary.view.timeline.v1"
}
```

or

```json
{
  "kind": "saved_view",
  "id": "svw_..."
}
```

**REQ-01-262**
When `sheet_ref.kind = view_schema`, `sheet_ref.id` MUST carry the `view_schema_id`. When `sheet_ref.kind = saved_view`, `sheet_ref.id` MUST carry the `saved_view_id`. For any pack-independent base-profile registry surface listed in REQ-01-307, `sheet_ref.kind = saved_view` always refers to a distinct saved-view object over that schema and MUST NOT be used as the canonical public identity of the required base surface itself. `field_key` MUST be present only when the client is focused on a concrete writable field and `mode = editing`.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-263**
The minimum server-to-client message set MUST be:

- `hello_ack`,
- `resume_ack`,
- `presence_snapshot`,
- `presence_delta`,
- `record_changed`,
- `job_progress`,
- `ping`,
- `error`,
- `session_revoked`.
Profiles: base, snapshot_reporting
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-233

**REQ-01-264**
`hello_ack.payload` MUST include `connection_id`, `resume_token`, `server_time`, `heartbeat_interval_ms`, `presence_ttl_ms`, and `resume_window_ms`. `resume_ack.payload` MUST include `status`, a fresh `resume_token`, and `server_high_water_stream_seq`. `resume_ack.payload.status` MUST be one of `replayed`, `reset_required`, or `rejected`. In the base profile, `heartbeat_interval_ms` MUST be `15000`, `presence_ttl_ms` MUST be `45000`, and `resume_window_ms` MUST be at least `300000`.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-265**
`presence_snapshot.payload.presences[]` MUST always be present. It MAY be empty. It MUST contain zero or more current presence records for the subscribed incident after expired presence records are pruned. Each presence record MUST include `connection_id`, `user_id`, `display_name`, `sheet_ref`, `mode`, `observed_at`, and `expires_at`, and MAY include `record_id` and `field_key`. `presence_snapshot.payload.presences[]` is a canonical keyed collection keyed by exact `connection_id`; it is not a recency list and not a display-order list. Within one incident, duplicate current `connection_id` values are forbidden on the wire. The server MUST serialize `presence_snapshot.payload.presences[]` in ascending lexicographic order of the exact `connection_id` string, with no locale-sensitive collation, no case folding, and no additional normalization. Array position in `presence_snapshot.payload.presences[]` has no client meaning. The same canonicalization MUST apply to the initial snapshot after `hello`, the fresh snapshot after `resume_ack.status='replayed'`, and the fresh snapshot after `resume_ack.status='reset_required'`.
Profiles: base, snapshot_reporting
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-233

**REQ-01-266**
`presence_delta.payload` MUST include `delta_kind` and `presence`. `delta_kind` MUST be one of `upsert` or `remove`. On `upsert`, `presence` MUST be a full presence record. On `remove`, `presence.connection_id` MUST identify the removed record; other presence members MAY be repeated for convenience but MUST NOT be required.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-267**
`record_changed.payload` MUST include `record_id`, `row_version`, `change_set_id`, `actor_user_id`, `changed_field_keys[]`, and `affected_views[]`. `changed_field_keys[]` MUST always be present. `changed_field_keys[]` is a canonical set of exact `field_key` identifiers. Duplicate `field_key` values are forbidden on the wire. The server MUST serialize `changed_field_keys[]` in ascending lexicographic order of the exact `field_key` string, with no locale-sensitive collation, no case folding, and no additional normalization. `changed_field_keys[]` MAY be empty only when the emitted `record_changed` reflects an effect on the addressed record for which no public `field_key` changed; in the current profile, soft-delete and restore are required examples. In every other current-profile case, `changed_field_keys[]` MUST be non-empty. `affected_views[]` MUST always be present and MUST NOT be empty. `affected_views[]` is a canonical keyed collection keyed by base `view_schema_id`. Duplicate `view_schema_id` values are forbidden on the wire. The server MUST serialize `affected_views[]` in ascending lexicographic order of the exact `view_schema_id` string, with no locale-sensitive collation, no case folding, and no additional normalization. Array position in `affected_views[]` has no client meaning; clients locate the relevant entry by `view_schema_id`. Each `affected_views[]` entry MUST include `view_schema_id` and `change_kind`. `change_kind` MUST be one of `patch`, `invalidate`, or `remove`. When `change_kind = patch`, the entry MUST include `patch_cells`, and `patch_cells` MUST be `view_row_patch_v1` for that view: top-level `record_id` and `row_version`, `cells` containing only changed fields, and optional `group_values` containing only changed grouping scalars. In `view_row_patch_v1`, omission of a cell or grouping scalar means unchanged, and an included cell with `{ "value": null }` means authoritative null when that field contract admits null. If the server cannot safely express the committed result for that view as `view_row_patch_v1`, it MUST use `change_kind = invalidate` rather than guessing a sparse patch. `affected_views[]` MUST be keyed by base `view_schema_id` values, not by visible tab labels, row order, or client-local filter state. For replayable `record_changed` messages, live emission and replay emission of semantically identical `changed_field_keys[]` and `affected_views[]` content MUST preserve identical canonical array order. A restored row that may re-enter a view MUST surface as `invalidate`; `/ws/v1/` MUST NOT introduce a distinct insert-like `change_kind` for that case within the v1 contract.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-368

**REQ-01-268**
`job_progress.payload` MUST include `job_id`, `scope`, `status`, `progress`, and `updated_at`, and MAY include `cancelable`, `message`, `result_summary`, `error_summary`, or `retained_until`. `status`, `progress`, `cancelable`, `result_summary`, `error_summary`, and `retained_until` MUST use the exact semantics defined for the HTTP job resource in §3.3.9.1. When `scope.kind = incident`, `scope.incident_id` MUST match the envelope `incident_id`. Deployment-scoped jobs MUST NOT emit on the incident-scoped stream.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-257, AC-258, AC-259

**REQ-01-269**
`error.payload` MUST include `code`, `message`, and `retryable`, and MAY include `details`. `session_revoked.payload` MUST include `reason_code` and MAY include `message`. `reason_code` MUST use the canonical session-revocation registry defined in §3.3.6.2.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-270**
The protocol MUST define two delivery classes:

- replayable ordered messages: `record_changed`, `job_progress`,
- ephemeral non-replayable messages: `hello_ack`, `resume_ack`, `presence_snapshot`, `presence_delta`, `ping`, `error`, `session_revoked`.
Profiles: base, snapshot_reporting
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-233

**REQ-01-271**
`stream_seq` MUST be monotonically increasing per `incident_id` across replayable messages. The server MUST assign `stream_seq` only after the underlying record mutation or incident-scoped job-state change is committed to authoritative server state. The server MUST NOT emit `record_changed` for uncommitted state.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-272**
The route path determines the incident subscription. `presence_update` determines only the sender's published presence scope. `record_changed` and incident-scoped `job_progress` MUST be broadcast to all currently authorized subscribers for that incident. Clients MUST determine active-view relevance locally using stable identifiers such as `view_schema_id`, `record_id`, `field_key`, and the client's current query contract.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-273**
The server MUST retain replayable messages for resume for at least 5 minutes or 10,000 replayable messages per incident, whichever is larger. After a valid `resume` whose retained replay window still covers `last_seen_stream_seq`, the server MUST send `resume_ack` with `status = replayed`, replay missed replayable messages in strict ascending `stream_seq`, and then send a fresh `presence_snapshot`.
Profiles: base, snapshot_reporting
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-233

**REQ-01-274**
If the server cannot honor incremental replay because the token is unknown, expired, malformed, or older than the retained replay window, but the caller still has valid incident authorization, the server MUST send `resume_ack` with `status = reset_required`, send a fresh `presence_snapshot`, and emit no guessed or partial replay for the missing range. The client MUST then discard incremental assumptions and re-query current workbook state through the HTTP view route before treating the socket as fully synchronized.
Profiles: base, snapshot_reporting
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-233

**REQ-01-275**
The server MAY use `resume_ack.payload.status = rejected` only when a `resume` message is syntactically invalid or the caller fails authentication or authorization checks. A rejected resume MUST be followed by `error` or `session_revoked` and immediate close.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-276**
The protocol MUST use application-level `ping` and `pong` messages inside this JSON envelope. The server MUST emit `ping` every 15 seconds when the connection is otherwise idle. The client MUST answer with `pong` within 10 seconds. The server MUST consider the connection dead after 45 seconds without any inbound frame. On clean close, the server MUST remove the corresponding presence immediately. On abrupt disconnect, presence expiry MUST follow this heartbeat timeout rather than a longer stale timeout.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-277**
For cookie-authenticated browser connections, the WebSocket upgrade and any session-establishment message MUST validate `Origin` against the configured application origin. The server MUST re-derive incident authorization on initial connect and on `resume`. WebSocket `ping` or `pong`, passive server push, and automatic reconnect or replay MUST NOT extend session idle expiry. If incident membership or session validity is revoked after connection establishment, or if the underlying session expires while the socket remains connected, the server MUST send `session_revoked` and close the socket. After session expiry or revocation, a later connection MUST establish a new authenticated session and use `hello`; `resume` alone is insufficient.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

## 4. Storage boundary

### 4.1 Postgres

**REQ-01-278**
Postgres MUST store:

- incidents, users, and memberships,
- record envelopes and first-class records,
- canonical indicators, indicator observations, and indicator lifecycle intervals,
- entity aliases and entity mentions,
- record links and record tags,
- change sets, mutation entries, and row revisions,
- saved views, workbook preference objects, and view schemas,
- projection tables,
- blob metadata,
- evidence lifecycle metadata,
- reference-pack manifests and integrity metadata,
- snapshot metadata, canonical export-model metadata, versioned template-contract metadata, versioned redaction-profile metadata, redaction manifests, and artifact release records when the Snapshot and Reporting Extension Profile is implemented.
Profiles: base, snapshot_reporting, reference_pack
Verified by: AC-231, AC-233..AC-234, AC-405

### 4.2 Object storage

**REQ-01-279**
S3-compatible object storage MUST store:

- binary evidence payloads,
- optionally, generated export artifacts when the Snapshot and Reporting Extension Profile is implemented.
Profiles: base, snapshot_reporting
Verified by: AC-231, AC-233, AC-405

The object store implementation MAY be MinIO in flyaway or on-prem deployments. Cloud deployments MAY use native S3, GCS, or Azure Blob behind an equivalent abstraction.

### 4.3 Storage exclusions

**REQ-01-280**
Large binary evidence MUST NOT be stored inline in Postgres.
Profiles: base
Verified by: AC-231, AC-405

**REQ-01-281**
A filesystem-backed blob adapter MAY be used for development or very small laboratories. It MUST NOT replace S3-compatible storage as the default production target.
Profiles: base
Verified by: AC-231

## 5. Incident data versus reference packs

**REQ-01-282**
Cartulary MUST distinguish **incident data** from **reference packs**.
Profiles: base, reference_pack
Verified by: AC-034, AC-092, AC-096, AC-231, AC-234

The following are incident data:

- incident records,
- evidence envelopes,
- revisions,
- saved views,
- workbook preference objects,
- immutable report snapshots.

The following are reference packs:

- framework mappings,
- type and icon registries,
- evidence vocabularies,
- optional enrichment datasets,
- view contracts when distributed independently of incidents.

**REQ-01-283**
Reference packs MUST version independently of incidents.
Profiles: base, reference_pack
Verified by: AC-034, AC-092, AC-096, AC-231, AC-234

**REQ-01-284**
Reference-pack manifests and integrity metadata MUST be stored in Postgres. Pack payloads MAY live on local disk or object storage behind the same abstraction, but their activation state and import or activation attestation MUST remain queryable from structured metadata.
Profiles: base, reference_pack
Verified by: AC-034, AC-092, AC-096, AC-231, AC-234

Reporting template packs MAY reuse the same integrity-verification and distribution machinery as reference packs. The template selected for a specific snapshot, the selected redaction profile, approval state, and rendered output hashes remain incident data.

## 6. View contracts

**REQ-01-285**
Each built-in sheet or contract-backed system view MUST be declared by a **`view_schema`** contract.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-231

**REQ-01-286**
A `view_schema` contract MUST define, at minimum:

- the stable `view_schema_id`,
- source record types,
- the base projection,
- an ordered field registry containing one entry per visible or writable field, each with a stable `field_key` and, when needed, optional `header_sort_field_key`,
- computed columns,
- the required hidden technical fields `record_id` and `row_version`,
- required reference packs, if any,
- an ordered default sort tuple,
- an explicit per-view `sort_fields` whitelist,
- filter semantics and any allowed grouping keys,
- per-field write-back semantics, including write target or write action, `conflict_resolution_class`, and `entity_binding_mode` where relevant,
- metadata needed to render the view consistently.
Profiles: base, reference_pack
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-231, AC-234, AC-362

**REQ-01-287**
Default sort tuples MUST be deterministic and MUST include `record_id` as the final tiebreaker unless a later profile explicitly overrides that rule.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-231

**REQ-01-288**
A base-profile implementation MUST expose the structured base-profile view-schema registry for conformance inspection through stored `view_schema` rows, the public discovery routes named in REQ-01-032, or an equivalent structured export. `GET /api/v1/view-schemas` MUST return the common success envelope with `data.view_schemas[]` plus `meta.paging`, MUST order results by `view_schema_id asc`, and MUST accept only `limit` and `cursor_token` under §3.3.7. `GET /api/v1/view-schemas/{view_schema_id}` MUST return one structured `view_schema_resource_v1` and MUST reject pagination members with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code=pagination_not_supported`. Conformance MUST NOT depend on scraping visible tab labels, column labels, or interactive UI behavior alone.

For public discovery, `view_schema_resource_v1` MUST expose the semantic workbook contract and MUST NOT require clients to infer structure from prose. The required members are:

- `view_schema_id`,
- `surface_kind`,
- `title`,
- `source_record_types`,
- `technical_fields`,
- `required_reference_pack_keys`,
- `default_sort`,
- `sort_fields`,
- `sort_null_order`,
- `filter_fields`,
- `synthetic_filter_predicates`,
- `grouping_fields`,
- `fields`.

For `view_schema_resource_v1`:

- `surface_kind` MUST use the closed vocabulary `built_in_sheet` and `system_view`,
- `title` MUST be a non-authoritative server-default-locale display hint,
- `technical_fields` MUST be the exact ordered array `["record_id","row_version"]`,
- `default_sort` MUST preserve declared tuple order,
- `sort_null_order` MUST be the exact token `last` in the current profile,
- `fields[]` MUST preserve the authoritative field-registry order for that `view_schema_id`,
- `source_record_types`, `required_reference_pack_keys`, `sort_fields`, `filter_fields`, and `grouping_fields` are set-like members and MUST use canonical ascending order,
- `filter_fields` MUST contain only keys also present in `fields[].field_key`,
- filter-only synthetic predicate keys MUST appear only in `synthetic_filter_predicates[]`,
- `synthetic_filter_predicates[]` MUST use canonical ascending `field_key` order,
- clients MUST ignore unknown additive response members they do not use.

Each `default_sort[]` entry MUST use exactly:

```json
{
  "field_key": "timeline.sort_ts",
  "direction": "asc"
}
```

Each `synthetic_filter_predicates[]` entry MUST use exactly:

```json
{
  "field_key": "note.full_text",
  "label": "Full Text",
  "filter_ops": ["full_text"]
}
```

Each `fields[]` entry MUST be `view_field_entry_v1` and MUST contain these required members:

- `field_key`,
- `label`,
- `default_hidden`,
- `sortable`,
- `header_sort_field_key`,
- `filter_ops`,
- `groupable`,
- `read_kind`,
- `write_kind`,
- `conflict_resolution_class`,
- `entity_binding_mode`,
- `string_contract_id`,
- `direct_scalar_contract_id`,
- `direct_reference_contract_id`,
- `clearable`,
- `enum_values`.

For `view_field_entry_v1`:

- `field_key` MUST be the stable field identity from the authoritative registry,
- `label` MUST be a non-authoritative server-default-locale display hint,
- `default_hidden`, `sortable`, `groupable`, and `clearable` MUST be booleans,
- `header_sort_field_key` MUST be either `null` or one key declared in `sort_fields`,
- `filter_ops` MUST use canonical operator order from §3.3.4.1 and MUST be `[]` when the field is not filterable,
- `read_kind` MUST use the closed vocabulary `text`, `number`, `boolean`, `timestamp`, `date`, `enum`, and `collection`,
- `write_kind` MUST use the closed vocabulary `read_only`, `direct_value`, and `action_payload`,
- `conflict_resolution_class` MUST be `null` when `write_kind='read_only'` and otherwise MUST use the closed vocabulary defined by Core 03 §3.3.3,
- `entity_binding_mode`, `string_contract_id`, `direct_scalar_contract_id`, and `direct_reference_contract_id` MUST be explicit `null` when not applicable,
- `enum_values` MUST be an explicit ordered array of tokens when the field is governed by a closed vocabulary and `null` otherwise.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-127, AC-231

**REQ-01-289**
View behavior MUST bind to `view_schema_id`, not to the visible tab label, column header text, or any other display label. `title` and field `label` values exposed by discovery are non-authoritative display hints only. The public discovery resource MUST describe semantic workbook behavior and MUST NOT expose `base_projection`, storage-table names, internal write targets, or other storage-realization details.
Repo-local owner contract artifacts MAY carry internal-only projection-binding metadata, including `base_projection`, when needed for conformance checks and generated implementation inputs. That metadata MUST remain outside `view_schema_resource_v1` and other runtime discovery payloads.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-231

**REQ-01-290**
If a visible column header, field label, or tab label changes, filter behavior, write-back behavior, and export semantics MUST remain unchanged unless the underlying `view_schema_id` changes. Changing only `title` or field `label` is non-breaking. Breaking changes to field membership, field meaning, writeability, `conflict_resolution_class`, `entity_binding_mode`, `sort_fields`, `filter_fields`, `synthetic_filter_predicates`, or `grouping_fields`, or any change that would reinterpret persisted `query_json`, MUST use a new `view_schema_id`. Breaking changes to shared-layout semantics MUST use a new `layout_schema_id`.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-231

## 7. Built-in sheets and system views

### 7.1 Built-in sheets

**REQ-01-291**
The base profile MUST provide the following built-in workbook sheets:

- Timeline,
- Hosts,
- Identities,
- Evidence,
- Notes.
Profiles: base
Verified by: AC-015, AC-116, AC-231

**REQ-01-292**
The built-in sheets MUST be projections or saved/system views over common underlying source tables rather than separate storage silos.
Profiles: base
Verified by: AC-015, AC-116, AC-231

**REQ-01-293**
The Hosts sheet MUST support interactive filtering and sorting over `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` without requiring synchronous scans of raw note text or evidence blobs.
Profiles: base
Verified by: AC-015, AC-116, AC-231

**REQ-01-294**
The Identities sheet MUST support interactive filtering and sorting over `privilege_level`, `mfa_state`, and `reset_status` without requiring synchronous scans of raw note text or evidence blobs.
Profiles: base
Verified by: AC-015, AC-116, AC-231

**REQ-01-295**
The Evidence sheet MUST support interactive filtering and sorting over `requested_at`, `received_at`, `collector_party_text`, `source_party_text`, `storage_ref`, `blob_hash`, and attachment or upload state without requiring synchronous blob access.
Profiles: base
Verified by: AC-015, AC-116, AC-231

### 7.2 Contract-backed system views

**REQ-01-296**
The base profile MUST support contract-backed system views for:

- indicators,
- compromise assessments,
- task requests,
- decisions,
- parties.

The Parties system view is an incident-scoped coordination-identity surface. It MUST NOT be treated as deployment-local user or account administration.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-116, AC-121, AC-122, AC-231, AC-277

**REQ-01-297**
System views MUST follow the same `view_schema_id` contract discipline as built-in sheets.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-01-298**
The Indicators system view MUST project canonical indicator records. It MUST NOT use source artifacts or source-bound indicator observations as the primary row identity.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-01-299**
The Compromise Assessments system view MUST project incident-scoped assessment records. It MUST NOT collapse assessment history into a mutable static property on a host or identity row.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-01-300**
The Task Requests system view MUST project `task_request` records. It MUST support queue-oriented filtering and sorting over `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `blocked_reason`, `completed_at`, `external_ticket_ref`, and `updated_at` without requiring synchronous scans of raw note or artifact text.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-01-301**
The Decisions system view MUST project `decision` records. It MUST support filtering and sorting over `status`, `owner_user_id`, `decision_type`, and `decided_at` without requiring synchronous scans of raw note or artifact text.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-01-302**
Structured coordination artifacts `comm_log`, `handoff`, `status_review`, and `lesson` MUST be available as workbook-native base-profile surfaces. Their canonical public identity MUST be the standardized `view_schema_id` for that surface: `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`, `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`. An implementation MAY vary the internal realization of those surfaces, including use of implementation-owned helper state or a saved-view-shaped helper object, but any such helper is a distinct implementation detail or distinct saved-view resource and MUST NOT replace the canonical public identity of the required surface. These surfaces MUST remain artifact-backed and MUST NOT require additional built-in sheets in the base profile.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231, AC-281, AC-282, AC-283, AC-284

### 7.3 Notes sheet contract

**REQ-01-303**
The base profile MUST expose **Notes** as a built-in workbook sheet.
Profiles: base
Verified by: AC-068, AC-069, AC-070, AC-112, AC-185, AC-231

**REQ-01-304**
The built-in Notes sheet MUST:

- be declared by a stable `view_schema_id`,
- use `artifact_grid_projection` as its base projection filtered to `artifact_type='note'`,
- support inline create from the sheet itself,
- remain backed by the shared artifact model rather than a Notes-specific storage silo.
Profiles: base
Verified by: AC-068, AC-069, AC-070, AC-112, AC-185, AC-231

**REQ-01-305**
The base profile MUST also expose contextual `add linked note` actions from Timeline, Hosts, Identities, and Evidence. All Notes entry paths MUST create the same underlying artifact record shape.
Profiles: base
Verified by: AC-068, AC-069, AC-070, AC-112, AC-185, AC-231

**REQ-01-306**
Notes behavior MUST NOT depend on the visible tab label. If the implementation allows the built-in Notes tab to be renamed or hidden per user, write-back behavior and export semantics MUST remain unchanged because they are bound to `view_schema_id`.
Profiles: base
Verified by: AC-068, AC-069, AC-070, AC-112, AC-185, AC-231

### 7.4 Authoritative base-profile view schema registry

**REQ-01-307**
The base profile MUST define the following fourteen pack-independent `view_schema` entries as the authoritative base-profile registry:

- `cartulary.view.timeline.v1`,
- `cartulary.view.hosts.v1`,
- `cartulary.view.identities.v1`,
- `cartulary.view.evidence.v1`,
- `cartulary.view.notes.v1`,
- `cartulary.view.indicators.v1`,
- `cartulary.view.assessments.v1`,
- `cartulary.view.task_requests.v1`,
- `cartulary.view.decisions.v1`,
- `cartulary.view.parties.v1`,
- `cartulary.view.comm_log.v1`,
- `cartulary.view.handoff.v1`,
- `cartulary.view.status_review.v1`,
- `cartulary.view.lesson.v1`.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-231, AC-281, AC-282, AC-283, AC-284

**REQ-01-308**
These fourteen entries are the authoritative required workbook surfaces of the base profile. Each pack-independent registry entry listed in REQ-01-307 is the canonical public workbook-surface identity for that surface. A saved-view object over the same schema is additive and non-canonical. In the current profile, no reference-pack-dependent framework overlay surface is a standardized workbook surface unless this core explicitly defines its `view_schema_id` and exhaustive contract. Implementations MUST NOT expose ATT&CK, D3FEND, VERIS, or other framework-specific pack overlays as workbook-discoverable `view_schema` resources in the base profile or the current Reference Pack Extension Profile. The only additional current-profile standardized workbook `view_schema_id` values beyond the fourteen pack-independent registry entries are, when implemented, `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1`. Later profiles MAY define additional standardized workbook surfaces, but they MUST NOT change the membership or semantics of the base-profile registry defined in this subsection.
Profiles: base, reference_pack
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-231, AC-234, AC-285, AC-286, AC-287

**REQ-01-579**
The base profile MUST define one authoritative cross-layer workbook-surface mapping table for every current-profile standardized workbook surface. The table is normative and identity-only. It MUST use the exact columns `surface`, `view_schema_id`, `surface_kind`, `source_record_types`, `canonical_source_discriminator_or_filter`, `surface_status`, and `required_reference_pack_keys`. `surface` is explanatory only and MUST NOT be treated as a second identifier. `surface_status` MUST use exactly `required built-in sheet`, `required system view`, and `standardized optional workbook surface`. For every row, the canonical workbook-surface identity is the `sheet_ref` object `{ "kind": "view_schema", "id": <view_schema_id> }`. The canonical row order is the REQ-01-307 registry order followed by `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1`. This table MUST NOT define or restate `base_projection`, storage-table names, internal write targets, per-field defaults, or other exhaustive field-registry content, and it MUST NOT add new members to `view_schema_resource_v1` or other runtime discovery payloads.
Profiles: base, reference_pack
Verified by: AC-411

**Table 7.4-A. Cross-layer workbook-surface identity mapping**

| `surface` | `view_schema_id` | `surface_kind` | `source_record_types` | `canonical_source_discriminator_or_filter` | `surface_status` | `required_reference_pack_keys` |
| --- | --- | --- | --- | --- | --- | --- |
| Timeline | `cartulary.view.timeline.v1` | `built_in_sheet` | `["timeline_event"]` | `record_type='timeline_event'` | required built-in sheet | `[]` |
| Hosts | `cartulary.view.hosts.v1` | `built_in_sheet` | `["host"]` | `record_type='host'` | required built-in sheet | `[]` |
| Identities | `cartulary.view.identities.v1` | `built_in_sheet` | `["identity"]` | `record_type='identity'` | required built-in sheet | `[]` |
| Evidence | `cartulary.view.evidence.v1` | `built_in_sheet` | `["evidence"]` | `record_type='evidence'` | required built-in sheet | `[]` |
| Notes | `cartulary.view.notes.v1` | `built_in_sheet` | `["artifact"]` | `artifact_type='note'` | required built-in sheet | `[]` |
| Indicators | `cartulary.view.indicators.v1` | `system_view` | `["indicator"]` | `record_type='indicator'` | required system view | `[]` |
| Compromise Assessments | `cartulary.view.assessments.v1` | `system_view` | `["assessment"]` | `record_type='assessment'` | required system view | `[]` |
| Task Requests | `cartulary.view.task_requests.v1` | `system_view` | `["task_request"]` | `record_type='task_request'` | required system view | `[]` |
| Decisions | `cartulary.view.decisions.v1` | `system_view` | `["decision"]` | `record_type='decision'` | required system view | `[]` |
| Parties | `cartulary.view.parties.v1` | `system_view` | `["party"]` | `record_type='party'` | required system view | `[]` |
| Communications Log | `cartulary.view.comm_log.v1` | `system_view` | `["artifact"]` | `artifact_type='comm_log'` | required system view | `[]` |
| Handoff | `cartulary.view.handoff.v1` | `system_view` | `["artifact"]` | `artifact_type='handoff'` | required system view | `[]` |
| Status Review | `cartulary.view.status_review.v1` | `system_view` | `["artifact"]` | `artifact_type='status_review'` | required system view | `[]` |
| Lesson | `cartulary.view.lesson.v1` | `system_view` | `["artifact"]` | `artifact_type='lesson'` | required system view | `[]` |
| Findings | `cartulary.view.findings.v1` | `system_view` | `["artifact"]` | `artifact_type='finding'`; subtype dimension `finding.kind` | standardized optional workbook surface | `[]` |
| Investigative Queries | `cartulary.view.investigative_queries.v1` | `system_view` | `["artifact"]` | implementation-declared structured investigative-query subtype; no current-profile fixed durable discriminator | standardized optional workbook surface | `[]` |
| Forensic Keywords | `cartulary.view.forensic_keywords.v1` | `system_view` | `["artifact"]` | implementation-declared structured forensic-keyword subtype; no current-profile fixed durable discriminator | standardized optional workbook surface | `[]` |

**REQ-01-309**
Each schema subsection below, together with the addenda in §19, is an exhaustive per-field registry for its `view_schema_id`, not an illustrative example. In particular, §7.4.2 through §7.4.4 close the base-profile interface contract for the built-in Hosts, Identities, and Evidence sheets, and §19 closes the Parties, coordination-artifact, and standardized optional artifact-backed surface contracts. Core 02 §10.4.4A MAY inventory the closed tagged-variant family for artifact-backed notes, coordination artifacts, and structured findings, but that registry is not a second owner for exhaustive field membership, create-time behavior, omitted-versus-`null` behavior, defaults, write targets or actions, or discovery metadata. These sections are also the sole authoritative source for populating public `view_schema_resource_v1`, `view_field_entry_v1`, and `synthetic_filter_predicates[]` discovery output. Implementations MUST NOT invent alternate base-profile or standardized optional writable `field_key` strings, write targets or actions, `conflict_resolution_class` assignments, `entity_binding_mode` values, or discovery metadata that conflicts with this registry. Surface `title` and field `label` values remain non-authoritative display hints only and MAY change without changing `view_schema_id` when field semantics do not change.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-231, AC-281, AC-282, AC-283, AC-284, AC-285, AC-286, AC-287, AC-410

**REQ-01-310**
Unless explicitly overridden below:

- `required_reference_pack_keys` MUST be `[]`,
- `record_id` and `row_version` MUST be present as hidden technical fields and are the only schema fields not serialized under `cells`,
- full `view_row_v1.cells` membership MUST be determined solely by the active schema's exhaustive field registry,
- every schema-declared non-technical field MUST appear in full `view_row_v1.cells` regardless of whether the field is visible, default-hidden, writable, or read-only,
- visibility and default-hidden state affect presentation only,
- writeability affects mutation eligibility only,
- hidden writable fields still serialize in full rows,
- in full `view_row_v1`, omission of a schema-declared non-technical field is invalid and `{ "value": null }` is the only authoritative null representation when the bound field admits null,
- the ordered default sort tuple is normative and MUST end with `record_id asc`,
- `sort_fields` MUST be an explicit whitelist of the stable `field_key` values that clients MAY use in `sort[]`; sortability MUST NOT be inferred from visibility, filterability, or writeability,
- only keys listed in `sort_fields` are client-sortable; `record_id` remains the mandatory server tiebreak and MUST NOT appear in `sort_fields`,
- hidden synthetic sort keys are allowed,
- when a visible field entry omits `header_sort_field_key`, column-header sort uses that field's own `field_key`; when `header_sort_field_key` is present, it MUST point to a key declared in `sort_fields`,
- current-profile sort comparison is type-driven and deterministic: timestamp and date fields sort chronologically, numeric fields sort numerically, boolean fields sort `false` then `true`, enum fields sort by their declared closed-vocabulary order, and text fields sort using the shared query-time text-comparison substrate defined by REQ-01-488,
- user-specified sorts place `null` values last in both directions, and public discovery MUST expose that current-profile behavior as `sort_null_order='last'`,
- `filter_fields` MUST list only schema-declared non-technical field keys from the exhaustive field registry; filter-only synthetic predicate keys are declared separately and MUST NOT appear in `fields[]`,
- filter semantics MUST be type-driven unless a schema below explicitly overrides them: enum and boolean fields use exact-match inclusion, timestamp and date fields use exact or range predicates, scalar identifier text uses exact or prefix matching on the shared query-time text-comparison substrate defined by REQ-01-488, multi-value collections use `contains_any` and `contains_all`, and declared full-text predicates use the exact-token `full_text` contract from §3.3.4.1 on that same substrate,
- for public discovery, `default_hidden` MUST equal membership in the schema's default-hidden field set, `sortable` MUST equal membership in `sort_fields`, `groupable` MUST equal membership in `grouping_fields`, `filter_ops` MUST equal the field's allowed operator set under this registry and §3.3.4.1, and `write_kind` MUST be `direct_value` for a declared write target, `action_payload` for a declared write action, and `read_only` otherwise,
- for public discovery, `read_kind` MUST be `collection` for `collection_value_v1` fields, `timestamp` or `date` for temporal scalars, `boolean` for booleans, `number` for numeric scalars, `enum` for closed-vocabulary scalars with `enum_values` in declared token order, and `text` otherwise,
- the canonical schema-derived default layout object MUST use `layout_schema_id='cartulary.layout.v1'`, `column_order` equal to the authoritative field-registry order of every schema-declared non-technical field, `hidden_field_keys` equal to the schema's default-hidden non-technical fields in canonical ascending order, and `column_widths=[]`,
- every writable field entry MUST declare `field_key`, read model, write target or write action, and `conflict_resolution_class`,
- every entity-bearing writable field entry MUST declare `entity_binding_mode`,
- every human-authored writable string field entry, and every writable string-bearing action member explicitly closed by §18, MUST declare `string_contract_id`,
- every writable direct temporal scalar field entry MUST declare `direct_scalar_contract_id` and `clearable=true` or `clearable=false`,
- required scalar fields reject explicit JSON `null` and values that normalize to empty under their bound field contract on both create and patch,
- optional scalar text fields that the bound string contract declares clearable MUST treat explicit JSON `null` or text that normalizes to empty as an authoritative clear to `null`; omission on patch means unchanged,
- optional direct temporal scalar fields with `clearable=true` MUST treat only explicit JSON `null` as an authoritative clear to `null`; direct temporal scalar fields with `clearable=false` MUST reject explicit JSON `null`; omission on patch means unchanged,
- `collection_review` fields MUST accept only `collection_actions_v1` on patch; raw arrays and raw JSON `null` MUST be rejected on patch,
- create-only identifiers such as subtype-specific row IDs remain immutable after first commit unless a later schema subsection explicitly overrides that rule,
- a create attempt that does not satisfy the active view contract's minimum create signal MUST leave no partial record row, no misleading projection row, and no misleading live-update event,
- fields not declared writable are read-only,
- per-user hide/show or reordering MAY change presentation but MUST NOT change field identity, filter semantics, or write-back semantics.
Profiles: base, reference_pack
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-184, AC-185, AC-231, AC-234, AC-281, AC-282, AC-283, AC-284, AC-285, AC-286, AC-287, AC-300, AC-301, AC-302, AC-303, AC-362, AC-363, AC-365, AC-366, AC-367, AC-368

**REQ-01-311**
Base-profile relationship mutations surfaced by these view contracts or their adjacent inspector or row-context actions MUST follow these routing rules:

- the client MUST NOT send `link_type`, direction flags, table names, or storage-specific routing metadata,
- the server MUST derive `record_links.link_type`, canonical `src_record_id` and `dst_record_id`, and storage routing from either the active `field_key` or the explicit action route,
- the base-profile mappings are:
  - `timeline.host_refs` -> `observed_on_host`, with the Timeline record as `src_record_id` and the resolved host record as `dst_record_id`,
  - `timeline.identity_refs` -> `observed_as_identity`, with the Timeline record as `src_record_id` and the resolved identity record as `dst_record_id`,
  - supported same-surface canonical-indicator linking actions on Timeline, Notes, other artifacts, or Evidence -> `references_indicator`, with the invoking source record as `src_record_id` and the canonical indicator record as `dst_record_id`; source-bound occurrences still use `indicator_observations`,
  - contextual evidence-association actions from a non-evidence record -> `attached_evidence`, with the invoking non-evidence record as `src_record_id` and the evidence record as `dst_record_id`,
  - contextual `add linked note` or equivalent artifact-association actions -> `references_artifact`, with the invoking record as `src_record_id` and the created or selected artifact record as `dst_record_id`,
  - `assessment.support_refs` -> `supported_by`, with the assessment record as `src_record_id` and the supporting record as `dst_record_id`,
  - `task.linked_record_ids` and the authoritative association represented by `task.decision_record_id` -> `references_record`, with the task-request record as `src_record_id` and the referenced record as `dst_record_id`,
  - `comm_log.decision_ids`, `comm_log.action_task_ids`, `handoff.open_task_ids`, `handoff.open_decision_ids`, `status_review.blocked_task_ids`, `status_review.pending_evidence_ids`, `status_review.open_decision_ids`, `lesson.follow_up_task_ids`, and `lesson.evidence_refs` -> `references_record`, with the owning coordination artifact as `src_record_id` and the referenced record as `dst_record_id`,
  - `decision.support_refs` -> `supported_by`, with the decision record as `src_record_id` and the supporting record as `dst_record_id`,
  - `decision.affected_record_ids` -> `references_record`, with the decision record as `src_record_id` and the affected record as `dst_record_id`,
  - explicit Timeline supersede actions that commit `replacement_record_id` -> `supersedes`, with the replacement Timeline row as `src_record_id` and the superseded Timeline row as `dst_record_id`,
  - explicit decision supersession actions -> `supersedes`, with the superseding decision as `src_record_id` and the superseded decision as `dst_record_id`.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-196, AC-231, AC-281, AC-282, AC-283, AC-284, AC-331, AC-332

**REQ-01-569**
The current profile defines no client-writable `confidence` member on any relationship mutation route, field action payload, or adjacent inspector or row-context relationship action. A request that supplies `confidence` for a relationship mutation MUST fail with `400` and `error.code = invalid_mutation_payload`. The server MUST NOT ignore, coerce, or persist that member.
Profiles: base
Verified by: AC-396

#### 7.4.1 `cartulary.view.timeline.v1`

**REQ-01-312**
- surface: built-in `Timeline` sheet
- source record types: `timeline_event`
- base projection: `timeline_grid_projection`
- `default_visible_fields`: `timeline.occurred_at`, `timeline.summary`, `timeline.host_refs`, `timeline.identity_refs`, `timeline.evidence_count`, `timeline.tags`, `timeline.edited_at`
- `default_hidden_fields`: `record_id`, `row_version`, `timeline.details`, `timeline.source_text`, `timeline.recorded_at`, `timeline.sort_ts`, `timeline.capture_state`, `timeline.replacement_record_id`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`
- `default_sort`: `timeline.sort_ts asc`, `record_id asc`. `timeline.sort_ts` MUST equal `occurred_at` when present and `recorded_at` otherwise.
- `sort_fields`: `timeline.sort_ts`, `timeline.summary`, `timeline.evidence_count`, `timeline.edited_at`, `timeline.capture_state`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`
- `filter_fields`: `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`, `timeline.tags`
- `grouping_fields`: `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`
- inline create: inline create from the sheet itself MUST create a `timeline_event` record. This view explicitly permits zero-field row creation on `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` when the request body carries only `client_txn_id`, so screenshot-only or later rough-capture flows do not require structured fields at create time
- writable fields:
  - `timeline.occurred_at`: read `occurred_at`; write target `timeline_events.occurred_at`; `header_sort_field_key=timeline.sort_ts`; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
  - `timeline.summary`: read `summary`; write target `timeline_events.summary`; `conflict_resolution_class=text_compare_merge`
  - `timeline.details`: read `details`; write target `timeline_events.details`; `conflict_resolution_class=text_compare_merge`
  - `timeline.source_text`: read `source_text`; write target `timeline_events.source_text`; `conflict_resolution_class=text_compare_merge`
  - `timeline.host_refs`: read resolved host chips plus unresolved host mentions; write action insert, update, or dismiss `entity_mentions` and resolved host `record_links` under `entity_binding_mode=mention_origin`; `conflict_resolution_class=collection_review`; `add_token.raw_text` and `add_resolved_ref.raw_text` use `string_contract_id=mention_token_text_v1`
  - `timeline.identity_refs`: read resolved identity chips plus unresolved identity mentions; write action insert, update, or dismiss `entity_mentions` and resolved identity `record_links` under `entity_binding_mode=mention_origin`; `conflict_resolution_class=collection_review`; `add_token.raw_text` and `add_resolved_ref.raw_text` use `string_contract_id=mention_token_text_v1`
  - `timeline.tags`: read `tag_names`; write action upsert tags and `record_tags`; `conflict_resolution_class=collection_review`
- read-only projection-backed or system-managed fields: `timeline.evidence_count`, `timeline.capture_state`, `timeline.replacement_record_id`, `timeline.edited_at`, `timeline.sort_ts`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`
- `timeline.capture_state` is the system-managed persisted workflow state defined by Core 03 §6. Clients MUST NOT supply `timeline.capture_state` as an initial writable value in `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` or as a `changes[].field_key` in `PATCH /api/v1/records/{record_id}`. Any attempted direct client write to `timeline.capture_state` through create or patch MUST fail closed under §3.3.5 and §3.3.6 rather than being silently ignored. Transitions to `reviewed` and `superseded` MUST use the dedicated record-scoped action routes defined in §3.3.5, and automatic transitions to `enriched` MUST be applied server-side with the committed Timeline mutation that triggered them.
- `timeline.replacement_record_id` MUST be `null` when no active incoming Timeline `supersedes` link exists for the row and otherwise MUST equal the unique replacement Timeline `record_id` derived from that active incoming link. It is hidden by default, read-only, and not part of the writable Timeline field set.
- `timeline.has_unresolved_mentions` MUST be `true` if and only if at least one non-deleted `entity_mentions` row for the source record has `resolution_status='unresolved'`; resolved or dismissed mentions MUST NOT make it `true`.
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231, AC-300, AC-301, AC-303, AC-331, AC-332

**REQ-01-313**
Collection-review wire contract for `timeline.host_refs` and `timeline.identity_refs`:

- `timeline.host_refs` and `timeline.identity_refs` MUST use `collection_value_v1` with `ordered=true`.
- The server MUST serialize `timeline.host_refs.items[]` and `timeline.identity_refs.items[]` in ascending `entity_mentions.ordinal` order and then ascending `item_ref`.
- The active `collection_value_v1.items[]` for these fields MUST include only non-deleted mentions whose `resolution_status` is `unresolved` or `resolved`; mentions with `resolution_status='dismissed'` MUST be omitted from `items[]` while remaining available through history and inspector affordances.
- Each `items[]` entry MUST use one of the following shapes:
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

```json
{
  "item_ref": "entity_mention:<entity_mention_id>",
  "item_kind": "unresolved_mention",
  "entity_type": "host",
  "display_text": "WS-023?",
  "raw_text": "WS-023?"
}
```

```json
{
  "item_ref": "entity_mention:<entity_mention_id>",
  "item_kind": "resolved_ref",
  "entity_type": "host",
  "display_text": "WS-023.corp.example",
  "raw_text": "WS-023",
  "resolved_record_id": "<record_id>"
}
```

**REQ-01-314**
- The same shapes and action vocabulary apply to identity items, with `entity_type='identity'`.
- Allowed actions for `timeline.host_refs` and `timeline.identity_refs` are `add_token`, `add_resolved_ref`, `resolve_item`, `revert_to_unresolved`, and `dismiss_item`.
- The legal transition matrix, side effects, and explicit mention-route equivalence for `resolve_item`, `dismiss_item`, and `revert_to_unresolved` are defined in §3.3.5.5 and MUST apply equally when those actions target a single mention through `collection_actions_v1`.
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

**REQ-01-315**
`add_token` MUST use:
Profiles: base
Verified by: AC-118, AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231, AC-388, AC-389, AC-390, AC-391

```json
{ "op": "add_token", "raw_text": "WS-023?" }
```

For both `timeline.host_refs` and `timeline.identity_refs`, `add_token.raw_text` MUST bind to `string_contract_id=mention_token_text_v1`.

**REQ-01-316**
`add_resolved_ref` MUST use:
Profiles: base
Verified by: AC-118, AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231, AC-388, AC-389, AC-390, AC-391

```json
{
  "op": "add_resolved_ref",
  "resolved_record_id": "<record_id>",
  "raw_text": "WS-023"
}
```

For both `timeline.host_refs` and `timeline.identity_refs`, `add_resolved_ref.raw_text` MUST bind to `string_contract_id=mention_token_text_v1`.

**REQ-01-317**
`resolve_item` MUST use:
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

```json
{
  "op": "resolve_item",
  "item_ref": "entity_mention:<entity_mention_id>",
  "resolved_record_id": "<record_id>"
}
```

**REQ-01-318**
`revert_to_unresolved` MUST use:
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

```json
{
  "op": "revert_to_unresolved",
  "item_ref": "entity_mention:<entity_mention_id>"
}
```

**REQ-01-319**
`dismiss_item` MUST use:
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

```json
{
  "op": "dismiss_item",
  "item_ref": "entity_mention:<entity_mention_id>"
}
```

The current-profile action objects above are closed to the listed members and intentionally omit any client-supplied `confidence` member.

**REQ-01-320**
For these two fields:

- the server MUST derive the target `entity_type`, any base-profile `link_type`, and storage routing from `field_key`; the client MUST NOT send `link_type`, table names, or storage-specific routing metadata,
- `add_resolved_ref` MUST create one backing `entity_mentions` row in resolved state and MUST create or upsert the corresponding active resolved `record_link` in the same `change_set`,
- `resolve_item` MUST preserve `raw_text`, set the targeted mention to resolved state, and MUST create or upsert the corresponding active resolved `record_link` in the same `change_set`,
- `revert_to_unresolved` MUST preserve `raw_text`, clear `resolved_record_id` and any current resolution metadata such as `resolved_at`, `resolved_by_user_id`, and `resolution_method`, set the targeted mention back to unresolved state whether it was previously resolved or dismissed, and MUST remove or tombstone any corresponding active resolved `record_link` in the same `change_set`,
- `dismiss_item` MUST preserve `raw_text`, stable mention identity, and provenance, MUST set `entity_mentions.resolution_status='dismissed'`, MUST clear `resolved_record_id` and any current resolution metadata such as `resolved_at`, `resolved_by_user_id`, and `resolution_method`, and MUST remove or tombstone any corresponding active resolved `record_link` in the same `change_set`,
- duplicate `add_token` actions with identical `raw_text` in the same or later request MUST create distinct mention rows rather than coalescing them.
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

**REQ-01-321**
Collection-review wire contract for `timeline.tags`:

- `timeline.tags` MUST use `collection_value_v1` with `ordered=false`.
- Each `items[]` entry MUST use this shape:
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

```json
{
  "item_ref": "record_tag:<record_id>:<tag_id>",
  "item_kind": "tag",
  "display_text": "rough",
  "tag_id": "<tag_id>"
}
```

- Allowed actions are:

```json
{ "op": "add_tag", "tag_name": "rough" }
```

```json
{ "op": "remove_tag", "item_ref": "record_tag:<record_id>:<tag_id>" }
```

**REQ-01-322**
- `tag_name` in `add_tag` MUST use `string_contract_id=tag_label_v1`.
- The server MUST normalize and compare `tag_name` using `tag_label_v1`, including trimmed Unicode NFC normalization and case-insensitive dedupe canonicalization equivalent to incident-scoped `tags.name` uniqueness. Empty tag names after normalization MUST be rejected with `400 Bad Request` and `error.code = invalid_mutation_payload`. Duplicate adds MUST coalesce to one surviving active binding.
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

#### 7.4.2 `cartulary.view.hosts.v1`

**REQ-01-323**
- surface: built-in `Hosts` sheet
- source record types: `host`
- base projection: `host_grid_projection`
- `default_visible_fields`: `host.display_name`, `host.hostname`, `host.aliases`, `host.host_state`, `host.linked_event_count`, `host.evidence_count`, `host.location`, `host.os_platform`, `host.business_owner`, `host.criticality`, `host.containment_status`, `host.edited_at`
- `default_hidden_fields`: `record_id`, `row_version`
- `default_sort`: `host.display_name asc`, `record_id asc`
- `sort_fields`: `host.display_name`, `host.hostname`, `host.host_state`, `host.linked_event_count`, `host.evidence_count`, `host.location`, `host.os_platform`, `host.business_owner`, `host.criticality`, `host.containment_status`, `host.edited_at`
- `filter_fields`: `host.host_state`, `host.business_owner`, `host.criticality`, `host.location`, `host.os_platform`, `host.containment_status`
- inline create: direct row creation or paste on the Hosts sheet MUST create or upsert a `host` record using `entity_binding_mode=entity_origin`
- create-or-upsert reuse on the Hosts sheet MUST apply the exact-match precedence in Core 02 §8.2. Suggestion-only candidates allowed by Core 02 §8.3 MUST NOT silently auto-merge or auto-resolve a host.
- writable fields:
  - `host.display_name`: read the canonical host display field; write target the canonical host display field on the underlying `host` record; `entity_binding_mode=entity_origin`; `string_contract_id=display_name_line_v1`; `conflict_resolution_class=atomic_replace`
  - `host.hostname`: read `hostname`; write target the canonical hostname field on the underlying `host` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `host.aliases`: read projected `suggestion_only` alias chips; write action upsert or remove `suggestion_only` alias values only; `entity_binding_mode=entity_origin`; `conflict_resolution_class=collection_review`
  - `host.location`: read `location`; write target the `location` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
  - `host.os_platform`: read `os_platform`; write target the `os_platform` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
  - `host.business_owner`: read `business_owner`; write target the `business_owner` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
  - `host.criticality`: read `criticality`; write target the `criticality` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
  - `host.containment_status`: read `containment_status`; write target the `containment_status` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
- read-only computed fields: `host.host_state`, `host.linked_event_count`, `host.evidence_count`, `host.edited_at`. `host.host_state` MUST be a projection-backed state that uses the exact tokens `stub` and `canonical`.
Profiles: base
Verified by: AC-097, AC-118, AC-124, AC-125, AC-231

**REQ-01-324**
Collection-review wire contract for `host.aliases`:

- `host.aliases` MUST use `collection_value_v1` with `ordered=false`.
- Each `items[]` entry MUST use this shape:
Profiles: base
Verified by: AC-097, AC-118, AC-124, AC-125, AC-231

```json
{
  "item_ref": "entity_alias:<entity_alias_id>",
  "item_kind": "alias",
  "display_text": "ws023",
  "alias_text": "ws023"
}
```

- Allowed actions are:

```json
{ "op": "add_alias", "alias_text": "ws023" }
```

```json
{ "op": "remove_alias", "item_ref": "entity_alias:<entity_alias_id>" }
```

**REQ-01-325**
- `host.aliases` MUST operate only on `suggestion_only` alias values in the base profile. It MUST NOT create, remove, or retag `exact_match_reuse` or `provenance_only` values.
- The server MUST derive any base-profile alias or reusable-identifier classification and storage routing from `field_key` and the active view contract; the client MUST NOT send table names, `alias_type`, classification flags, or storage-specific routing metadata.
- Alias rename in the base profile MUST be expressed as `remove_alias` plus `add_alias`. The public API surface MUST NOT require in-place alias-row update semantics.
- `alias_text` in `add_alias` MUST use `string_contract_id=alias_text_v1`.
- Duplicate alias adds for the same canonical record and normalized `alias_text` under `alias_text_v1` MUST coalesce to one surviving alias row.
Profiles: base
Verified by: AC-097, AC-118, AC-124, AC-125, AC-231

#### 7.4.3 `cartulary.view.identities.v1`

**REQ-01-326**
- surface: built-in `Identities` sheet
- source record types: `identity`
- base projection: `identity_grid_projection`
- `default_visible_fields`: `identity.display_name`, `identity.upn`, `identity.email`, `identity.sam_account_name`, `identity.aliases`, `identity.identity_state`, `identity.linked_event_count`, `identity.evidence_count`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status`, `identity.edited_at`
- `default_hidden_fields`: `record_id`, `row_version`
- `default_sort`: `identity.display_name asc`, `record_id asc`
- `sort_fields`: `identity.display_name`, `identity.upn`, `identity.email`, `identity.sam_account_name`, `identity.identity_state`, `identity.linked_event_count`, `identity.evidence_count`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status`, `identity.edited_at`
- `filter_fields`: `identity.identity_state`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status`
- inline create: direct row creation or paste on the Identities sheet MUST create or upsert an `identity` record using `entity_binding_mode=entity_origin`
- create-or-upsert reuse on the Identities sheet MUST apply the exact-match precedence in Core 02 §8.2. Suggestion-only candidates allowed by Core 02 §8.3 MUST NOT silently auto-merge or auto-resolve an identity.
- writable fields:
  - `identity.display_name`: read the canonical identity display field; write target the canonical identity display field on the underlying `identity` record; `entity_binding_mode=entity_origin`; `string_contract_id=display_name_line_v1`; `conflict_resolution_class=atomic_replace`
  - `identity.upn`: read `upn`; write target the canonical UPN field on the underlying `identity` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `identity.email`: read `email`; write target the canonical email field on the underlying `identity` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `identity.sam_account_name`: read `sam_account_name`; write target the canonical `sam_account_name` field on the underlying `identity` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `identity.aliases`: read projected `suggestion_only` alias chips; write action upsert or remove `suggestion_only` alias values only; `entity_binding_mode=entity_origin`; `conflict_resolution_class=collection_review`
  - `identity.privilege_level`: read `privilege_level`; write target the `privilege_level` field on the underlying `identity` record; `conflict_resolution_class=atomic_replace`
  - `identity.mfa_state`: read `mfa_state`; write target the `mfa_state` field on the underlying `identity` record; `conflict_resolution_class=atomic_replace`
  - `identity.reset_status`: read `reset_status`; write target the `reset_status` field on the underlying `identity` record; `conflict_resolution_class=atomic_replace`
- read-only computed fields: `identity.identity_state`, `identity.linked_event_count`, `identity.evidence_count`, `identity.edited_at`. `identity.identity_state` MUST be a projection-backed state that uses the exact tokens `stub` and `canonical`.
Profiles: base
Verified by: AC-098, AC-118, AC-124, AC-125, AC-231

**REQ-01-327**
`identity.aliases` MUST use the same `collection_value_v1` item shape, `collection_actions_v1` action vocabulary, and `suggestion_only`-only semantics as `host.aliases`, except the active `field_key` is `identity.aliases`.
Profiles: base
Verified by: AC-098, AC-118, AC-124, AC-125, AC-231

#### 7.4.4 `cartulary.view.evidence.v1`

**REQ-01-328**
- surface: built-in `Evidence` sheet
- source record types: `evidence`
- base projection: `evidence_grid_projection`
- `default_visible_fields`: `evidence.title`, `evidence.lifecycle_state`, `evidence.requested_at`, `evidence.received_at`, `evidence.storage_ref`, `evidence.blob_hash`, `evidence.collector_party_text`, `evidence.source_party_text`, `evidence.upload_state`, `evidence.linked_record_count`, `evidence.edited_at`
- `default_hidden_fields`: `record_id`, `row_version`, `evidence.collector_party_id`, `evidence.source_party_id`
- these default-hidden direct-reference fields remain part of every full `view_row_v1.cells` object for this schema; default-hidden affects presentation only
- `default_sort`: `evidence.requested_at desc`, `record_id asc`
- `sort_fields`: `evidence.title`, `evidence.lifecycle_state`, `evidence.requested_at`, `evidence.received_at`, `evidence.storage_ref`, `evidence.blob_hash`, `evidence.collector_party_text`, `evidence.source_party_text`, `evidence.upload_state`, `evidence.linked_record_count`, `evidence.edited_at`
- `filter_fields`: `evidence.lifecycle_state`, `evidence.upload_state`, `evidence.requested_at`, `evidence.received_at`, `evidence.collector_party_text`, `evidence.source_party_text`, `evidence.storage_ref`, `evidence.blob_hash`
- inline create: a blank zero-field create attempt MUST NOT commit. A create attempt that supplies no user-supplied non-empty writable evidence field MAY commit only when the same visible create flow successfully finalizes a blob attachment before first commit
- minimum create signal: the first committed evidence row MUST include either at least one user-supplied non-empty writable evidence field after create-time normalization or one successfully finalized blob attachment from that same visible create flow
- `evidence.collector_party_id` and `evidence.source_party_id` MAY be written through inspector or same-surface enrichment flows, but they MUST NOT by themselves satisfy the minimum create signal and MUST NOT clear preserved party text implicitly
- server-filled defaults and create-time derived fields MUST NOT satisfy the minimum create signal
- if `evidence.lifecycle_state` is omitted on create, the server MUST default it to `requested`; when that defaulted create omits `evidence.requested_at`, the server MUST fill `requested_at` from the commit timestamp
- a failed create that lacks the minimum create signal, or a create flow whose blob finalization fails before first commit, MUST leave no committed evidence row and no misleading projection update
- writable fields:
  - `evidence.title`: read the evidence title field; write target the evidence record title field; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `evidence.lifecycle_state`: read `lifecycle_state`; write target `evidence_records.lifecycle_state`; `conflict_resolution_class=atomic_replace`. State-dependent field guards and blob-bridge invariants from Core 02 §13 and Core 03 §8.3 apply.
  - `evidence.requested_at`: read `requested_at`; write target `evidence_records.requested_at`; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
  - `evidence.received_at`: read `received_at`; write target `evidence_records.received_at`; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
  - `evidence.storage_ref`: read `storage_ref`; write target `evidence_records.storage_ref`; `string_contract_id=locator_text_v1`; `conflict_resolution_class=atomic_replace`
  - `evidence.collector_party_text`: read `collector_party_text`; write target `evidence_records.collector_party_text`; `string_contract_id=party_text_v1`; `conflict_resolution_class=text_compare_merge`
  - `evidence.collector_party_id`: read the canonical collector party reference; write target `evidence_records.collector_party_id`; `direct_reference_contract_id=same_incident_party_ref_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
  - `evidence.source_party_text`: read `source_party_text`; write target `evidence_records.source_party_text`; `string_contract_id=party_text_v1`; `conflict_resolution_class=text_compare_merge`
  - `evidence.source_party_id`: read the canonical source party reference; write target `evidence_records.source_party_id`; `direct_reference_contract_id=same_incident_party_ref_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
- read-only computed fields: `evidence.blob_hash`, `evidence.upload_state`, `evidence.linked_record_count`, `evidence.edited_at`. `evidence.blob_hash` and `evidence.upload_state` are derived fields and MUST NOT satisfy the minimum create signal.
- blob attach or replacement MUST remain an explicit evidence action. It MUST NOT be modeled as a direct write to `evidence.blob_hash` or `evidence.upload_state`.
Profiles: base
Verified by: AC-100, AC-118, AC-124, AC-125, AC-128, AC-231, AC-278, AC-279, AC-280, AC-300, AC-301, AC-303, AC-315, AC-316, AC-317, AC-318

#### 7.4.5 `cartulary.view.notes.v1`

**REQ-01-329**
- surface: built-in `Notes` sheet
- source record types: `artifact`
- base projection: `artifact_grid_projection` filtered to `artifact_type='note'`
- `default_visible_fields`: `note.title`, `note.body`, `note.tags`, `note.linked_record_count`, `note.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `note.created_by_user_id`
- `default_sort`: `note.updated_at desc`, `record_id asc`
- `sort_fields`: `note.title`, `note.body`, `note.linked_record_count`, `note.updated_at`, `note.created_by_user_id`
- `filter_fields`: `note.tags`, `note.created_by_user_id`, `note.updated_at`
- `synthetic_filter_predicates`: `note.full_text`
- `note.full_text` is a filter-only synthetic predicate key declared by this view schema. It binds to the generic `full_text` operator contract in §3.3.4.1 over the union of the normalized `note.title` and `note.body` fields. It supports only `filter_ops=[full_text]`, is not a writable field, and need not be a visible column.
- inline create: zero-field create is forbidden
- minimum create signal: inline create from the sheet itself MUST commit only when at least one of `note.title` or `note.body` is non-empty after create-time normalization; whitespace-only text MUST be treated as absent
- the server MUST fill `artifact_type='note'`, timestamps, and attribution on first commit
- context-preseeded links from `add linked note` MUST remain editable context and MUST NOT by themselves satisfy the minimum create signal
- `note.tags` is optional follow-on structure and MUST NOT satisfy the minimum create signal
- writable fields:
  - `note.title`: read the note title field; write target the note artifact title field; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `note.body`: read the note body field; write target the note artifact body field; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
  - `note.tags`: read `tag_names`; write action upsert tags and `record_tags`; `conflict_resolution_class=collection_review`
- read-only computed fields: `note.linked_record_count`, `note.updated_at`, `note.created_by_user_id`
Profiles: base
Verified by: AC-068, AC-069, AC-070, AC-112, AC-118, AC-124, AC-125, AC-185, AC-231

**REQ-01-330**
`note.tags` MUST use the same `collection_value_v1` item shape and `collection_actions_v1` action vocabulary as `timeline.tags`, except the active `field_key` is `note.tags`.
Profiles: base
Verified by: AC-068, AC-069, AC-070, AC-112, AC-118, AC-124, AC-125, AC-185, AC-231

#### 7.4.6 `cartulary.view.indicators.v1`

**REQ-01-331**
- surface: contract-backed `Indicators` system view
- source record types: `indicator`
- base projection: `indicator_grid_projection`
- `default_visible_fields`: `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, `indicator.normalized_value`, `indicator.defanged_value`, `indicator.hash_algorithm`, `indicator.hash_value`, `indicator.stix_pattern`, `indicator.first_observed_at`, `indicator.last_observed_at`, `indicator.observation_count`, `indicator.lifecycle_summary`, `indicator.supporting_link_count`
- `default_hidden_fields`: `record_id`, `row_version`
- `default_sort`: `indicator.last_observed_at desc`, `indicator.display_value asc`, `record_id asc`
- `sort_fields`: `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, `indicator.normalized_value`, `indicator.defanged_value`, `indicator.hash_algorithm`, `indicator.hash_value`, `indicator.stix_pattern`, `indicator.first_observed_at`, `indicator.last_observed_at`, `indicator.observation_count`, `indicator.lifecycle_summary`, `indicator.supporting_link_count`
- `filter_fields`: `indicator.indicator_type`, `indicator.value_kind`, `indicator.hash_algorithm`, `indicator.first_observed_at`, `indicator.last_observed_at`, `indicator.lifecycle_summary`
- inline create: zero-field create is forbidden
- minimum create signal: inline create from the sheet itself MUST commit only when the request supplies enough information to determine canonical identity: `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, and `indicator.normalized_value` whenever the type-specific normalization rule applies and the server cannot derive it deterministically from the other supplied fields
- `indicator.hash_algorithm` and `indicator.hash_value` MUST be supplied together or omitted together on create
- if the canonical dedupe basis is not determinable at create time, create MUST fail with no partial indicator row
- writable fields on create only:
  - `indicator.indicator_type`: read `indicator_type`; write target the `indicator_type` field on the underlying `indicator` record; `conflict_resolution_class=atomic_replace`
  - `indicator.value_kind`: read `value_kind`; write target the `value_kind` field on the underlying `indicator` record; `conflict_resolution_class=atomic_replace`
  - `indicator.display_value`: read the canonical display value; write target the canonical display-value field on the underlying `indicator` record; `conflict_resolution_class=atomic_replace`
  - `indicator.normalized_value`: read `normalized_value`; write target the `normalized_value` field on the underlying `indicator` record; `conflict_resolution_class=atomic_replace`
  - `indicator.defanged_value`: read `defanged_value`; write target the `defanged_value` field on the underlying `indicator` record; `conflict_resolution_class=atomic_replace`
  - `indicator.hash_algorithm`: read `hash_algorithm`; write target the `hash_algorithm` field on the underlying `indicator` record; `conflict_resolution_class=atomic_replace`
  - `indicator.hash_value`: read `hash_value`; write target the `hash_value` field on the underlying `indicator` record; `conflict_resolution_class=atomic_replace`
  - `indicator.stix_pattern`: read `stix_pattern`; write target the `stix_pattern` field on the underlying `indicator` record; `conflict_resolution_class=text_compare_merge`
- read-only computed fields: `indicator.first_observed_at`, `indicator.last_observed_at`, `indicator.observation_count`, `indicator.lifecycle_summary`, `indicator.supporting_link_count`
- existing-row writable fields: none
- grid edits to an existing indicator row MUST reject writes to every field listed under `writable fields on create only`
- the exact identity-defining immutable field set for this v1 schema is:
  - always: `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, `indicator.normalized_value`
  - additionally, when populated and used by the canonical dedupe key: `indicator.hash_algorithm`, `indicator.hash_value`
- `indicator.stix_pattern` remains create-only in this view but MUST NOT be treated as identity-defining or as part of the minimum create signal
- `indicator.defanged_value` remains create-only in this view but MUST NOT be treated as identity-defining or as part of the minimum create signal
- no other additional type-specific dedupe-basis field exists in `cartulary.view.indicators.v1`
- any future additional identity-basis field requires a new explicit stable `field_key` and a new `view_schema` version
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-118, AC-122, AC-124, AC-231

#### 7.4.7 `cartulary.view.assessments.v1`

**REQ-01-332**
- surface: contract-backed `Assessments` system view
- source record types: `assessment`
- base projection: `assessment_grid_projection`
- `default_visible_fields`: `assessment.subject_ref`, `assessment.subject_type`, `assessment.assessment_state`, `assessment.confidence_band`, `assessment.confidence_score`, `assessment.rationale`, `assessment.assessor`, `assessment.assessed_at`, `assessment.supporting_link_count`
- `default_hidden_fields`: `record_id`, `row_version`, `assessment.support_refs`
- `default_sort`: `assessment.assessed_at desc`, `record_id asc`
- `sort_fields`: `assessment.subject_ref`, `assessment.subject_type`, `assessment.assessment_state`, `assessment.confidence_band`, `assessment.confidence_score`, `assessment.rationale`, `assessment.assessor`, `assessment.assessed_at`, `assessment.supporting_link_count`
- `filter_fields`: `assessment.subject_type`, `assessment.assessment_state`, `assessment.confidence_band`, `assessment.assessed_at`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `assessment.subject_ref`, `assessment.subject_type`, `assessment.assessment_state`, and non-empty `assessment.rationale` are present after create-time normalization
- context-preseeded `assessment.subject_ref` and `assessment.subject_type` MAY seed the create surface, but they MUST NOT by themselves make an otherwise empty create valid
- if omitted on create, the server MUST default `assessment.assessed_at` to the commit timestamp, `assessment.assessor` to the authenticated actor, and `assessment.confidence_score` to `NULL`
- `assessment.support_refs` MAY be included on create but MUST NOT satisfy the minimum semantic create set
- writable fields on create only:
  - `assessment.subject_ref`: read the subject record reference; write target the assessment subject reference; `conflict_resolution_class=atomic_replace`
  - `assessment.subject_type`: read the subject type; write target the assessment subject type; `conflict_resolution_class=atomic_replace`
  - `assessment.assessment_state`: read `assessment_state`; write target the `assessment_state` field on the underlying `assessment` record; `conflict_resolution_class=atomic_replace`
  - `assessment.confidence_score`: read `confidence_score`; write target the `confidence_score` field on the underlying `assessment` record; `conflict_resolution_class=atomic_replace`. A band-first editor MUST map `unset`, `low`, `medium`, and `high` to `NULL`, `25`, `55`, and `85` respectively.
  - `assessment.rationale`: read `rationale`; write target the `rationale` field on the underlying `assessment` record; `conflict_resolution_class=text_compare_merge`
  - `assessment.assessor`: read the assessor field; write target the assessor field on the underlying `assessment` record; `conflict_resolution_class=atomic_replace`
  - `assessment.assessed_at`: read `assessed_at`; write target the `assessed_at` field on the underlying `assessment` record; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=false`; `conflict_resolution_class=atomic_replace`
  - `assessment.support_refs`: read supporting record references; write action upsert supporting `record_links` or denormalized support references; `conflict_resolution_class=collection_review`
- read-only computed fields: `assessment.confidence_band`, `assessment.supporting_link_count`
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-118, AC-121, AC-124, AC-231, AC-300, AC-302, AC-303

**REQ-01-333**
Collection-review wire contract for `assessment.support_refs`:

- On read and conflict surfaces, `assessment.support_refs` MUST use `collection_value_v1` with `ordered=false`.
- When row creation initializes `assessment.support_refs`, the create-time field value MUST use `collection_actions_v1`.
- Each `items[]` entry MUST use this shape:
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-118, AC-121, AC-124, AC-231

```json
{
  "item_ref": "record_ref:<linked_record_id>",
  "item_kind": "record_ref",
  "display_text": "<linked record summary>",
  "linked_record_id": "<record_id>"
}
```

- Allowed actions are:

```json
{ "op": "add_record_ref", "linked_record_id": "<record_id>" }
```

```json
{ "op": "remove_record_ref", "item_ref": "record_ref:<linked_record_id>" }
```

The current-profile action objects above are closed to the listed members and intentionally omit any client-supplied `confidence` member. Schemas that reuse this same collection contract inherit that same omission.

**REQ-01-334**
- The server MUST derive any base-profile `link_type` and storage routing from `field_key`; the client MUST NOT send `link_type`, table names, or storage-specific routing metadata.
- Duplicate adds for the same patched `record_id`, `linked_record_id`, and field-derived link type MUST coalesce to one surviving logical reference binding.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-118, AC-121, AC-124, AC-231

**REQ-01-335**
- grid edits to an existing assessment row MUST reject writes to `assessment.subject_ref`, `assessment.subject_type`, `assessment.assessment_state`, `assessment.confidence_score`, `assessment.rationale`, `assessment.assessor`, `assessment.assessed_at`, and `assessment.support_refs`.
- append-only semantics for semantic assessment fields begin only after a valid first commit satisfying the minimum semantic create set in `REQ-01-332`.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-118, AC-121, AC-124, AC-231

#### 7.4.8 `cartulary.view.task_requests.v1`

**REQ-01-336**
- surface: required workbook-native contract-backed `Task Requests` system view with canonical public identity `cartulary.view.task_requests.v1`; any saved view over this same `view_schema_id` is a distinct saved-view object rather than the required base surface
- source record types: `task_request`
- base projection: `task_request_grid_projection`
- `default_visible_fields`: `task.title`, `task.status`, `task.owner_user_id`, `task.priority`, `task.task_kind`, `task.workstream`, `task.due_at`, `task.requester_party_text`, `task.blocked_reason`, `task.completed_at`, `task.external_ticket_ref`, `task.linked_record_count`, `task.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `task.requester_party_id`, `task.closure_summary`, `task.linked_record_ids`, `task.decision_record_id`, `task.no_owner`
- these default-hidden fields remain part of every full `view_row_v1.cells` object for this schema; default-hidden affects presentation only
- `default_sort`: `task.updated_at desc`, `record_id asc`
- `sort_fields`: `task.title`, `task.status`, `task.owner_user_id`, `task.priority`, `task.task_kind`, `task.workstream`, `task.due_at`, `task.requester_party_text`, `task.blocked_reason`, `task.completed_at`, `task.external_ticket_ref`, `task.linked_record_count`, `task.updated_at`, `task.no_owner`
- `filter_fields`: `task.status`, `task.owner_user_id`, `task.priority`, `task.task_kind`, `task.workstream`, `task.due_at`, `task.requester_party_text`, `task.blocked_reason`, `task.completed_at`, `task.external_ticket_ref`, `task.no_owner`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `task.title` is non-empty and `task.task_kind` is present after create-time normalization
- when omitted on interactive inline create from the sheet itself, the server MUST default `task.status` to `open`, `task.owner_user_id` to the authenticated actor, and `task.priority` to `normal`
- preseeded `task.linked_record_ids` or `task.decision_record_id` MAY seed the create surface, but they MUST NOT satisfy the minimum create signal
- `task.requester_party_id` MAY be written through inspector or same-surface enrichment flows, but it MUST NOT by itself satisfy the minimum create signal and MUST NOT clear preserved requester text implicitly
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `task.title`: read `title`; write target the `title` field on the underlying `task_request` record; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `task.status`: read `status`; write target the `status` field on the underlying `task_request` record; legal writes MUST be validated and any required lifecycle normalization MUST be applied under Core 02 §10.4.1.1 before commit; `conflict_resolution_class=atomic_replace`
  - `task.owner_user_id`: read `owner_user_id`; write target the `owner_user_id` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.priority`: read `priority`; write target the `priority` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.task_kind`: read `task_kind`; write target the `task_kind` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.workstream`: read `workstream`; write target the `workstream` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.due_at`: read `due_at`; write target the `due_at` field on the underlying `task_request` record; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
  - `task.requester_party_text`: read `requester_party_text`; write target the `requester_party_text` field on the underlying `task_request` record; `string_contract_id=party_text_v1`; `conflict_resolution_class=text_compare_merge`
  - `task.requester_party_id`: read the canonical requester party reference; write target the `requester_party_id` field on the underlying `task_request` record; `direct_reference_contract_id=same_incident_party_ref_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
  - `task.blocked_reason`: read `blocked_reason`; write target the `blocked_reason` field on the underlying `task_request` record; `string_contract_id=reason_note_v1`; `conflict_resolution_class=text_compare_merge`
  - `task.completed_at`: read `completed_at`; write target the `completed_at` field on the underlying `task_request` record; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
  - `task.external_ticket_ref`: read `external_ticket_ref`; write target the `external_ticket_ref` field on the underlying `task_request` record; `string_contract_id=locator_text_v1`; `conflict_resolution_class=atomic_replace`
  - `task.closure_summary`: read `closure_summary`; write target the `closure_summary` field on the underlying `task_request` record; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
  - `task.linked_record_ids`: read linked record references; write action upsert or remove linked `record_links`; `conflict_resolution_class=collection_review`
  - `task.decision_record_id`: read `decision_record_id`; write target the `decision_record_id` field on the underlying `task_request` record or an equivalent linked decision reference; `direct_reference_contract_id=same_incident_decision_ref_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
- read-only computed fields: `task.linked_record_count`, `task.updated_at`, `task.no_owner`
Profiles: base
Verified by: AC-085, AC-118, AC-124, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231, AC-278, AC-279, AC-280, AC-300, AC-301, AC-303, AC-304, AC-315, AC-316, AC-317, AC-318, AC-319

**REQ-01-337**
`task.linked_record_ids` MUST use the same `collection_value_v1` item shape and `collection_actions_v1` action vocabulary as `assessment.support_refs`, except the active `field_key` is `task.linked_record_ids` and the server derives the applicable `link_type` from that field key.
Profiles: base
Verified by: AC-085, AC-118, AC-124, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-01-338**
- task lifecycle semantics remain authoritative: any committed write set affecting `task.status`, `task.blocked_reason`, `task.completed_at`, or `task.owner_user_id` MUST produce a resulting row that satisfies Core 02 §10.4.1.1. In particular, `status='blocked'` requires `blocked_reason`, `status='done'` requires `completed_at`, active tasks MUST NOT be ownerless, a successful transition away from `blocked` or `done` MUST clear `blocked_reason` or `completed_at` respectively, and a successful write that sets `status='done'` with no explicit `completed_at` MUST fill `completed_at` from the commit timestamp.
Profiles: base
Verified by: AC-085, AC-118, AC-124, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231, AC-304

#### 7.4.9 `cartulary.view.decisions.v1`

**REQ-01-339**
- surface: required workbook-native contract-backed `Decisions` system view with canonical public identity `cartulary.view.decisions.v1`; any saved view over this same `view_schema_id` is a distinct saved-view object rather than the required base surface
- source record types: `decision`
- base projection: `decision_grid_projection`
- `default_visible_fields`: `decision.summary`, `decision.status`, `decision.owner_user_id`, `decision.decision_type`, `decision.decided_at`, `decision.rationale`, `decision.support_refs`, `decision.affected_record_count`, `decision.supersedes_record_id`, `decision.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `decision.affected_record_ids`, `decision.is_superseded`
- `default_sort`: `decision.decided_at desc`, `record_id asc`
- `sort_fields`: `decision.summary`, `decision.status`, `decision.owner_user_id`, `decision.decision_type`, `decision.decided_at`, `decision.rationale`, `decision.affected_record_count`, `decision.supersedes_record_id`, `decision.updated_at`, `decision.is_superseded`
- `filter_fields`: `decision.status`, `decision.owner_user_id`, `decision.decision_type`, `decision.decided_at`, `decision.is_superseded`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `decision.decision_type` is present and both `decision.summary` and `decision.rationale` are non-empty after create-time normalization
- when omitted on create, the server MUST default `decision.status` to `proposed`, `decision.owner_user_id` to the authenticated actor, and `decision.decided_at` to the commit timestamp
- preseeded `decision.support_refs` or `decision.affected_record_ids` MAY seed the create surface, but they MUST NOT satisfy the minimum create signal
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `decision.summary`: read `summary`; write target the `summary` field on the underlying `decision` record; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `decision.status`: read `status`; write target the `status` field on the underlying `decision` record; legal direct writes MUST be validated against Core 02 §10.4.2.1, and a direct write whose requested `status` is `superseded` MUST be rejected; `conflict_resolution_class=atomic_replace`
  - `decision.owner_user_id`: read `owner_user_id`; write target the `owner_user_id` field on the underlying `decision` record; `conflict_resolution_class=atomic_replace`
  - `decision.decision_type`: read `decision_type`; write target the `decision_type` field on the underlying `decision` record; `conflict_resolution_class=atomic_replace`
  - `decision.decided_at`: read `decided_at`; write target the `decided_at` field on the underlying `decision` record; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=false`; `conflict_resolution_class=atomic_replace`
  - `decision.rationale`: read `rationale`; write target the `rationale` field on the underlying `decision` record; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
  - `decision.support_refs`: read `support_refs`; write action upsert or remove supporting `record_links` or denormalized support references; `conflict_resolution_class=collection_review`
  - `decision.affected_record_ids`: read affected record references; write action upsert or remove affected-record `record_links`; `conflict_resolution_class=collection_review`
- read-only computed fields: `decision.affected_record_count`, `decision.supersedes_record_id`, `decision.updated_at`, `decision.is_superseded`. `decision.supersedes_record_id` MUST be a read-only computed projection of the authoritative `record_links` supersession relation with `link_type='supersedes'`
Profiles: base
Verified by: AC-086, AC-118, AC-124, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231, AC-300, AC-302, AC-303

**REQ-01-340**
`decision.support_refs` and `decision.affected_record_ids` MUST use the same `collection_value_v1` item shape and `collection_actions_v1` action vocabulary as `assessment.support_refs`, except the active `field_key` is `decision.support_refs` or `decision.affected_record_ids` and the server derives the applicable `link_type` from that field key.
Profiles: base
Verified by: AC-086, AC-118, AC-124, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-01-341**
- supersession remains an explicit decision-linking flow through typed `record_links` with `link_type='supersedes'`. It MUST NOT be modeled as a direct write to `decision.supersedes_record_id` or as a direct write that sets `decision.status='superseded'` in the base-profile grid.
- initial create with `decision.status='superseded'` remains rejected under the same lifecycle rule.
- when the explicit supersession flow succeeds, it MUST persist the authoritative `record_links` supersession relation, refresh any read-only computed `decision.supersedes_record_id` projection, and apply the status effects defined by Core 02 §10.4.2.1 atomically.
Profiles: base
Verified by: AC-086, AC-118, AC-124, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

## 8. Projection model

### 8.1 Projection tables

**REQ-01-342**
Hot workbook screens MUST use projection tables rather than Postgres materialized views.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

**REQ-01-343**
The implementation MUST define projection tables equivalent to:

- `timeline_grid_projection`,
- `host_grid_projection`,
- `identity_grid_projection`,
- `artifact_grid_projection`,
- `evidence_grid_projection`,
- `indicator_grid_projection` for the indicator system view,
- `assessment_grid_projection` for the compromise-assessment system view,
- `task_request_grid_projection` for the task-request system view,
- `decision_grid_projection` for the decision system view,
- `party_grid_projection` for the Parties system view.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231, AC-277

**REQ-01-344**
Each projection row MUST represent exactly one primary record in the base projection for that view.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

**REQ-01-345**
For `indicator_grid_projection`, the primary record MUST be the canonical indicator record for that row.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

**REQ-01-346**
For `assessment_grid_projection`, the primary record MUST be the assessment record for that row.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

**REQ-01-347**
For `task_request_grid_projection`, the primary record MUST be the `task_request` record for that row.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

**REQ-01-348**
For `decision_grid_projection`, the primary record MUST be the `decision` record for that row.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

### 8.2 Projection-row identity

**REQ-01-349**
Every projection row exposed to the client MUST carry:

- the stable `record_id` of the underlying mutation target,
- the current `row_version` required for optimistic writes.
Profiles: base
Verified by: AC-124, AC-125, AC-231

**REQ-01-350**
The client MUST NOT infer mutation targets from row position, displayed values, group headers, or transient selection state.
Profiles: base
Verified by: AC-013, AC-125, AC-231

### 8.3 Projection maintenance

**REQ-01-351**
Projection rows MUST be updated transactionally with the source write that changes them.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

**REQ-01-352**
Projection tables are derived state and MUST NOT be authoritative history.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

**REQ-01-353**
The implementation MUST provide a deterministic rebuild command or equivalent maintenance operation that can regenerate projections from source tables.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

### 8.4 Projection corruption

**REQ-01-354**
If a projection becomes corrupt or stale, the implementation MUST treat the projection as disposable cache state, rebuild it from authoritative source data, and preserve source-of-truth consistency.
Profiles: base
Verified by: AC-231

### 8.5 Hot-path retrieval and evidence boundary

**REQ-01-355**
Hot workbook sheets MUST serve the visible viewport from projection rows and other small derived metadata. They MUST NOT synchronously scan source tables or evidence blobs to render the grid hot path.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-356**
Interactive retrieval for hot workbook sheets MUST use a deterministic sort tuple and stable cursor, keyset, or viewport/block retrieval. It MUST NOT rely on deep `OFFSET` pagination for large incidents.

The current-profile workbook view-query route `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` MUST be a generic workbook-surface query route rather than a Timeline-only route. Its JSON request body MAY include `sort`, `filters`, `group_by`, `limit`, and `cursor_token`. `sort`, `filters`, and `group_by` MUST normalize under the active `view_schema_id` contract. `limit`, when supplied, MUST be an integer in the supported workbook query page-size range. `cursor_token`, when supplied, MUST be treated as an opaque continuation token bound to the authenticated actor, route family, incident, `view_schema_id`, normalized query contract, and page limit. A continuation request whose cursor binding no longer matches the request MUST fail closed rather than silently starting a new query.

Successful workbook view-query responses MUST return the common success envelope with workbook rows in `data.rows` and paging metadata in `meta.paging`. `meta.paging` MUST include the effective `limit`, a boolean `has_more`, and `next_cursor` as either an opaque string continuation token or JSON `null` on terminal pages. Invalid cursor tokens, expired cursor snapshots, unsupported pagination members, invalid limits, and cursor/query mismatches MUST use stable route-owned validation errors rather than internal errors.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-357**
Projection tables MUST support deterministic interactive retrieval over the ordered default sort tuple and any contract-declared interactive grouping key for the view. An implementation MAY satisfy this with indexes or an equivalent mechanism that preserves the same observable latency envelope.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-358**
Projection rows for hot workbook sheets and workbook-native contract-backed system views MUST carry the scalar fields required for interactive sort, filter, grouping, selection anchoring, and evidence badges in the visible viewport. For the Timeline sheet, this MUST include at least `sort_ts`, day buckets equivalent to `timeline.occurred_day` and `timeline.recorded_day`, `capture_state`, `has_evidence`, `has_unresolved_mentions`, and `evidence_count`. For the coordination-artifact and standardized optional artifact-backed surfaces defined by REQ-01-503 through REQ-01-509 when those surfaces are present, this MUST include the projection-backed scalar or derived fields needed for `comm_log.comm_type`, `comm_log.timestamp_day`, `comm_log.next_report_day`, `comm_log.audience`, `comm_log.channel_or_meeting`, `handoff.timestamp_day`, `handoff.outgoing_owner_user_id`, `handoff.incoming_owner_user_id`, `handoff.ack_state`, `status_review.timestamp_day`, `status_review.review_owner_user_id`, `status_review.next_report_day`, `lesson.closure_state`, `lesson.owner_user_id`, `lesson.timestamp_day`, `finding.kind`, `finding.state`, `finding.owner_user_id`, `finding.confidence_band`, `investigative_query.platform`, `investigative_query.created_by_user_id`, `investigative_query.created_day`, `forensic_keyword.match_mode`, `forensic_keyword.case_sensitive`, and `forensic_keyword.created_day`.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231, AC-281, AC-282, AC-283, AC-284, AC-285, AC-286, AC-287

**REQ-01-359**
For the Hosts sheet, exact lookup, sorting, filtering, and pivot counts over `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` MUST be satisfiable from `host_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text or evidence blobs.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-360**
For the Identities sheet, exact lookup, sorting, filtering, and pivot counts over `privilege_level`, `mfa_state`, and `reset_status` MUST be satisfiable from `identity_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text or evidence blobs.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-361**
For the Evidence sheet, exact lookup, sorting, filtering, and queue views over `requested_at`, `received_at`, `collector_party_text`, `source_party_text`, `storage_ref`, `blob_hash`, and upload or attachment state MUST be satisfiable from `evidence_grid_projection` and other small derived metadata. They MUST NOT require synchronous blob access.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-362**
The grid and inspector hot path MUST synchronously read only scalar fields, flags, counts, and small preview handles needed for the visible viewport or selected row. They MUST NOT synchronously fetch full attachment lists or binary blob bytes as part of grid rendering, row selection, sheet filtering, grouping, or inspector metadata open.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-363**
For the Indicators system view, exact lookup, sorting, filtering, and pivot counts over canonical indicators MUST be satisfiable from `indicator_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw timeline text, artifact text, or evidence blobs.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-364**
For the Compromise Assessments system view, exact lookup, sorting, filtering, and pivot counts over `assessment_state` and derived `confidence_band` MUST be satisfiable from `assessment_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw timeline text, artifact text, or evidence blobs.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-365**
For the Task Requests system view, exact lookup, sorting, filtering, queue counts, and stale-work views over `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `blocked_reason`, `completed_at`, `external_ticket_ref`, and `updated_at` MUST be satisfiable from `task_request_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text, communications logs, or evidence blobs.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-366**
For the Decisions system view, exact lookup, sorting, filtering, and review queues over `status`, `owner_user_id`, `decision_type`, `decided_at`, and supersession state MUST be satisfiable from `decision_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text, communications logs, or evidence blobs.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

## 9. Canonical derivation layer

**REQ-01-367**
Cartulary MUST maintain a single canonical derivation layer for all derived surfaces.
Profiles: base
Verified by: AC-032, AC-231

**REQ-01-368**
The following surfaces MUST read from the same derivation/query logic or from an explicitly versioned snapshot of that logic:

- interactive workbook views,
- report sections,
- framework rollups,
- future visualizations,
- exported artifacts.
Profiles: base, snapshot_reporting
Verified by: AC-032, AC-231, AC-233

This requirement exists to prevent drift in filtering, counts, inclusion rules, identifier stability, and ordering.

## 10. Snapshot and reporting extension profile

### 10.1 Extension boundary

**REQ-01-369**
The **Snapshot and Reporting Extension Profile** is optional for base conformance. If implemented, it MUST satisfy all requirements in this section and the corresponding criteria in Core 04.
Profiles: snapshot_reporting
Verified by: AC-030, AC-046, AC-233

### 10.2 Snapshot semantics

**REQ-01-370**
A reporting-capable implementation MUST treat report and presentation generation as a subsystem rather than direct ad hoc reads from live workbook tables.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-233

**REQ-01-371**
Each rendered artifact MUST bind to one immutable release tuple equivalent to:

- `snapshot_id`,
- `snapshot_at`,
- `source_change_set_high_watermark`,
- `derivation_version`,
- `template_id` and `template_version`,
- `redaction_profile_id` and `redaction_profile_version`,
- `export_model_sha256`.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-233

**REQ-01-372**
The implementation MUST:

- capture the `snapshot_at` boundary and frozen source boundary,
- materialize a canonical export model such as `incident_report_model.json`,
- render derivative outputs from that immutable snapshot rather than from mutable live tables.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-233

After the canonical export model has been materialized for a given release tuple, no renderer, template, or presentation builder MAY query live workbook tables or mutable projections for additional case content.

**REQ-01-373**
Re-rendering from the same release tuple MUST reproduce the same canonical export model, deterministic ordering, and `export_model_sha256`.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-233

### 10.2.1 Rendered artifact lifecycle

For Snapshot and Reporting Extension Profile outputs, Cartulary defines an artifact-scoped lifecycle over each rendered output candidate.

**REQ-01-374**
The authoritative persisted representation for this lifecycle MUST be the release record plus its bound release tuple and any bound approval records. Approval state MUST NOT bind to mutable incident rows or to template metadata outside the release record.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233

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

**REQ-01-375**
Approval invalidation MUST be an explicit lifecycle transition on the artifact record. It MUST NOT be implemented only as an implicit UI rule.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233

**REQ-01-376**
A new render with a different logical output slot or different `output_sha256` MUST start as a distinct `pending_approval` candidate. It MUST NOT inherit `approved` or `published` state from an earlier candidate.
Profiles: snapshot_reporting
Verified by: AC-059, AC-060, AC-104, AC-105, AC-106, AC-233

### 10.3 Export-model classification and release scopes

**REQ-01-377**
Every exportable field or block in the canonical export model MUST carry exactly one `content_class` with one of the following values:

- `source_evidence` for direct evidence references, hashes, timestamps, filenames or media labels, and exported excerpts or thumbnails,
- `derived_analytic` for deterministic transforms such as timelines, counts, ATT&CK rollups, graphs, and relationship summaries,
- `curated_narrative` for analyst-authored findings prose, executive summaries, recommendations, impact statements, and analyst-authored lessons-learned narrative,
- `working_material` for scratch text, unresolved notes, internal comments, and unreviewed excerpts.
Profiles: snapshot_reporting
Verified by: AC-057, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-113, AC-114, AC-115, AC-233, AC-333

The closed vocabulary for artifact release scope is:

- `internal_draft`,
- `internal_review`,
- `external_release`.

**REQ-01-378**
The snapshot and export subsystem MUST evaluate output eligibility against the chosen `release_scope` using at least the following matrix:

- `internal_draft`: any `content_class` except raw blob bytes,
- `internal_review`: any `content_class` except raw blob bytes, and any included `working_material` MUST remain visibly marked non-releasable,
- `external_release`: `derived_analytic`, `curated_narrative`, and only selected `source_evidence` excerpts or thumbnails that are eligible for the chosen `release_scope`. Raw blob bytes and `working_material` MUST NOT appear.

Eligibility filtering MUST operate on persisted `content_class`; `release_scope` MUST NOT rewrite classification.
Profiles: snapshot_reporting
Verified by: AC-057, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-113, AC-114, AC-115, AC-233, AC-333

**REQ-01-379**
For `external_release`, every `curated_narrative` block MUST carry `support_refs[]` containing one or more stable identifiers to supporting findings, events, evidence records, assessments, or query records. A narrative block lacking `support_refs[]` MUST be ineligible for `external_release`.
Profiles: snapshot_reporting
Verified by: AC-057, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-113, AC-114, AC-115, AC-233, AC-333

**REQ-01-380**
Direct-source text-bearing blocks first materialized from ad hoc note artifacts, structured finding rows where `finding.kind='hypothesis'`, `task_request` records, `decision` records, `comm_log` artifacts, `handoff` artifacts, `status_review` artifacts, and `lesson` artifacts:

- MUST default to `content_class='working_material'` when first materialized into the canonical export model,
- MUST receive that default during snapshot and export-model derivation, before template rendering,
- MUST preserve that `content_class` across `internal_draft`, `internal_review`, and `external_release`; `release_scope` filters eligibility, but it MUST NOT rewrite classification.

For this rule, copied, quoted, normalized, truncated, concatenated, or lightly reformatted text from those source families counts as direct-source text.

A block composed only of deterministic non-narrative scalars derived from those same source families, such as stable identifiers, enums, timestamps, counts, or other non-narrative scalar values, MAY use `derived_analytic`.

If one block mixes direct-source text with analytic material and there is no explicit curation boundary, that block MUST default to `working_material`.

Templates and renderers MUST consume persisted `content_class`. They MUST NOT infer a more permissive class by omission, heuristic, or template-specific convention.

Raw `lesson` record text follows this same default. A separately materialized analyst-authored lessons-learned block MAY use `curated_narrative` only if it independently satisfies the selected `content_class`, any required `support_refs[]`, and applicable redaction rules.

Such source-derived content MUST NOT appear in `external_release` unless an analyst has explicitly curated it into a separate export-model block that independently satisfies the selected `content_class`, any required `support_refs[]`, and applicable redaction rules.
Profiles: snapshot_reporting
Verified by: AC-057, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-113, AC-114, AC-115, AC-233, AC-333

### 10.4 Template packs and rendering contract

**REQ-01-381**
A reporting-capable implementation MUST treat report templates as versioned, integrity-checked local asset bundles or equivalent template packs.
Profiles: snapshot_reporting
Verified by: AC-058, AC-091, AC-233

**REQ-01-382**
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
Profiles: snapshot_reporting
Verified by: AC-058, AC-091, AC-233

**REQ-01-383**
Template bundles SHOULD reuse the same integrity-verification and activation machinery as reference packs when that machinery is available. Template activation MUST remain independent of optional reference-pack presence.
Profiles: snapshot_reporting
Verified by: AC-058, AC-091, AC-233

**REQ-01-384**
A template renderer MUST:

- read only from the immutable export model and declared local assets,
- MUST NOT query live workbook tables or mutable projections,
- MUST NOT fetch network assets,
- MUST NOT execute arbitrary user-supplied code,
- fail closed if the template references an undeclared binding or a missing required field.
Profiles: snapshot_reporting
Verified by: AC-058, AC-091, AC-233

### 10.5 Redaction profiles and manifests

**REQ-01-385**
A reporting-capable implementation MUST apply redaction to the canonical export model rather than by mutating incident records.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

**REQ-01-386**
Each versioned redaction profile MUST declare, at minimum:

- `redaction_profile_id`,
- `redaction_profile_version`,
- a default action,
- per-`content_class` rules,
- optional per-field overrides keyed by stable export-model path.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

**REQ-01-387**
When recipient-specific reporting is implemented, a versioned redaction profile MUST also be able to declare zero or more allowed stable incident-local `disclosure_partition_refs[]`.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

The closed vocabulary for redaction actions is:

- `allow`,
- `drop`,
- `mask`,
- `truncate`,
- `hash`,
- `stub`.

**REQ-01-388**
Redaction MUST run after snapshot materialization and before template rendering.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

**REQ-01-389**
If a field or block eligible for the chosen `release_scope` appears in the canonical export model without an applicable redaction rule, rendering MUST fail closed.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

**REQ-01-390**
When a field or block carries `disclosure_partition_refs[]` that are not allowed by the selected redaction profile, the renderer MUST apply the applicable redaction rule or fail closed. If a field or block contains mixed-partition content and no applicable rule can produce a conformant result, rendering MUST fail closed.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

**REQ-01-391**
Disclosure partition metadata and redaction-profile selection MUST affect only snapshot-derived rendering and release. They MUST NOT affect live workbook queries, projections, or incident authorization.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

**REQ-01-392**
Each rendered artifact MUST emit a `redaction_manifest` keyed by stable export-model path and rule identifier, recording every field or block that was dropped, masked, truncated, hashed, or stubbed.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

**REQ-01-393**
A render that changes `template_id`, `template_version`, `redaction_profile_id`, or `redaction_profile_version` relative to an earlier artifact candidate MUST create a distinct `pending_approval` candidate. It MUST NOT inherit approval or publication state from the earlier artifact.
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-113, AC-114, AC-115, AC-233

### 10.6 Output forms and generated-presentation boundary

A reporting-capable implementation MAY generate:

- Markdown reports,
- Mermaid diagram sources,
- Slidev decks,
- HTML reports,
- operator-facing reenactment outputs such as Asciinema-style walkthroughs.

**REQ-01-394**
`output_kind` MUST use a stable closed vocabulary equivalent to:

- `html`,
- `markdown`,
- `slidev`,
- `mermaid`,
- `reenactment`.
Profiles: snapshot_reporting
Verified by: AC-031, AC-061, AC-062, AC-233

Generated presentations MAY:

- rearrange and visualize snapshot facts,
- render deterministic summaries from approved export-model fields,
- generate graphs, timelines, and deck structure,
- include analyst-authored narrative blocks,
- include approved evidence excerpts with provenance.

**REQ-01-395**
Generated presentations MUST NOT:

- invent new facts, timestamps, actors, or causal chains,
- synthesize commands or terminal sessions that were not observed,
- interpolate missing steps between observed events,
- rewrite unreviewed working material into releasable claims,
- present a reenactment as externally releasable observed operator activity unless the represented steps are themselves explicit evidence.
Profiles: snapshot_reporting
Verified by: AC-031, AC-061, AC-062, AC-233

`mermaid` and `slidev` outputs MAY be `external_release` only when every rendered block satisfies the selected `release_scope`.

**REQ-01-396**
`reenactment` outputs MUST be marked `generated_presentation=true` and are limited to `internal_review` in the current profile.
Profiles: snapshot_reporting
Verified by: AC-031, AC-061, AC-062, AC-233

**REQ-01-397**
If such outputs are generated, the implementation MUST preserve the distinction between source evidence and generated presentation material.
Profiles: snapshot_reporting
Verified by: AC-031, AC-061, AC-062, AC-233

### 10.7 Self-contained outputs

**REQ-01-398**
Generated report and presentation artifacts MUST be self-contained. They MUST NOT require remote JavaScript, CSS, fonts, or runtime media assets at render time.
Profiles: snapshot_reporting
Verified by: AC-031, AC-233

## 11. Reference Pack Extension Profile

### 11.1 Extension boundary

**REQ-01-399**
The **Reference Pack Extension Profile** is optional for base conformance. If implemented, it MUST satisfy this section and the corresponding criteria in Core 04.
Profiles: reference_pack
Verified by: AC-033, AC-034, AC-035, AC-234

### 11.2 Minimum disconnected bundle

**REQ-01-400**
For the smallest supported flyaway or disconnected deployment that implements this profile, the deployment MUST preinstall and activate exactly the following three reference packs by default:

- `type_registry.host`
- `type_registry.evidence`
- `type_registry.indicator`
Profiles: reference_pack
Verified by: AC-092, AC-234

**REQ-01-401**
These three packs define the minimum disconnected bundle because base-profile host, evidence, and indicator semantics MUST come from versioned registries rather than hard-coded UI labels or workbook headers.
Profiles: reference_pack
Verified by: AC-092, AC-234

**REQ-01-402**
The smallest supported disconnected bundle MUST NOT require or preinstall framework overlay packs. Current-profile framework add-on pack keys are:

- `framework.attack`
- `framework.d3fend`
- `framework.veris`
Profiles: reference_pack
Verified by: AC-092, AC-234

**REQ-01-403**
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
Profiles: reference_pack
Verified by: AC-092, AC-234

**REQ-01-404**
Other enrichment or framework pack keys MAY exist. They MUST follow the same activation, verification, and degradation rules defined by this profile.
Profiles: reference_pack
Verified by: AC-092, AC-234

**REQ-01-405**
If the Snapshot and Reporting Extension Profile is implemented, template bundles MUST remain separately installable. They MUST NOT count toward the minimum disconnected reference-pack bundle.
Profiles: reference_pack
Verified by: AC-092, AC-234

**REQ-01-406**
Separately distributed `view_contract` packs MAY exist in larger deployments. They MUST NOT be required for the smallest disconnected bundle.
Profiles: reference_pack
Verified by: AC-092, AC-234

Larger supported disconnected bundles MAY preinstall additional packs, but the minimum disconnected bundle is fixed by this subsection.

### 11.3 Offline import, update, and activation flow

**REQ-01-407**
In a flyaway or disconnected deployment, reference-pack update MUST use an offline bundle import flow. The running application MUST NOT perform a live internet fetch as part of pack verification or activation.
Profiles: reference_pack
Verified by: AC-033, AC-093, AC-094, AC-096, AC-234

**REQ-01-408**
The import and update flow MUST satisfy all of the following:

1. the operator supplies a pack bundle either by placing it in the configured reference-pack storage root or by submitting it through `POST /api/v1/reference-packs/import`,
2. the system stages the bundle inside the configured temporary-work root,
3. the system verifies the staged bundle before any extracted content becomes active,
4. on successful verification, the system records the candidate version in durable condition `verified_available`; in storage this is realized by `reference_packs.status='available'`, `verification_result='passed'`, and the version not being the active version for its `pack_key`,
5. activation requires an explicit operator action that switches the active version pointer for the target `pack_key`.
Profiles: reference_pack
Verified by: AC-033, AC-093, AC-094, AC-096, AC-234

#### 11.3.1 Linked reference-pack lifecycle machines

For each imported reference-pack version, Cartulary defines two linked lifecycle machines:

- a verification and availability machine authoritative on `reference_packs`,
- an activation machine authoritative on `reference_pack_activation_state` and `reference_pack_attestations`.

**REQ-01-409**
These machines are linked but separate. Successful verification does not by itself activate a version, and activation MUST NOT bypass verification.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-093, AC-094, AC-095, AC-096, AC-234

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

**REQ-01-410**
A `failed` or `missing` version MUST NOT become `active` without first returning through `staged` or `verified_available` by a new import or successful re-verification path. A `disabled`, `failed`, or `missing` version MUST NOT remain or become the active version pointer for its `pack_key`.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-093, AC-094, AC-095, AC-096, AC-234

**REQ-01-411**
At most one version of a given `pack_key` MUST be active at a time.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-093, AC-094, AC-095, AC-096, AC-234

**REQ-01-412**
The implementation MUST retain the previously active version for each `pack_key` until an explicit administrative removal occurs, so operator rollback does not require incident-data changes.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-093, AC-094, AC-095, AC-096, AC-234

**REQ-01-413**
Reference-pack import, verification, and refresh MUST execute as background jobs rather than as blocking grid actions.
Profiles: reference_pack
Verified by: AC-033, AC-035, AC-093, AC-094, AC-095, AC-096, AC-234

### 11.4 Verification and attestation

**REQ-01-414**
Before a reference pack enters durable condition `verified_available` or becomes `active`, the implementation MUST verify:

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
Profiles: reference_pack
Verified by: AC-035, AC-094, AC-095, AC-234

**REQ-01-415**
Pack import or activation MUST fail closed on checksum mismatch, signature mismatch, missing required integrity metadata, contract incompatibility, incomplete download or copy, path-traversal attempt, or disallowed content.
Profiles: reference_pack
Verified by: AC-035, AC-094, AC-095, AC-234

**REQ-01-416**
If verification fails, the candidate pack version MUST remain inactive, and the previously active version, if any, MUST remain active.
Profiles: reference_pack
Verified by: AC-035, AC-094, AC-095, AC-234

**REQ-01-417**
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
Profiles: reference_pack
Verified by: AC-035, AC-094, AC-095, AC-234

**REQ-01-418**
Attestation metadata MUST remain queryable from structured metadata without unpacking bundle contents or consulting incident data.
Profiles: reference_pack
Verified by: AC-035, AC-094, AC-095, AC-234

### 11.4.1 Activation safety and observability

**REQ-01-419**
Activation MUST read only from a `verified_available` candidate and MUST emit a structured activation attestation bound to the target `pack_key` and `pack_version`.
Profiles: reference_pack
Verified by: AC-034, AC-035, AC-095, AC-234

**REQ-01-420**
If an active version is disabled, fails a later integrity or compatibility check, or becomes missing, the implementation MUST remove or replace the active pointer in `reference_pack_activation_state` before any pack-dependent operation can continue to treat that version as active.
Profiles: reference_pack
Verified by: AC-034, AC-035, AC-095, AC-234

**REQ-01-421**
Pack lifecycle state MUST be observable from structured metadata alone. At minimum, an operator and a CI conformance test MUST be able to determine the current pack-version condition, the active-version pointer, the previous active version, the last verification result, and the last import or activation attestation without unpacking bundle contents.
Profiles: reference_pack
Verified by: AC-034, AC-035, AC-095, AC-234

### 11.5 Degradation behavior

**REQ-01-422**
If optional reference packs are absent, disabled, failed, or missing, Cartulary MUST continue to support:

- timeline capture,
- entity resolution,
- evidence attachment,
- core editing.
Profiles: reference_pack
Verified by: AC-034, AC-234

Only the affected overlay labels, enrichment semantics, non-canonical analytical widgets, or snapshot/report derivations MAY degrade. This clause does not authorize additional workbook `view_schema` surfaces in the current profile.

## 12. Portability, backup, restore, and failure handling

### 12.1 Backup

Operational backup and restore remain deployment-local recovery behavior. They are distinct from whole-incident portability.

**REQ-01-571**
Each successful operational backup MUST produce one retained `backup_set` bound to exactly one declared `consistency_point_at`. A coherent restore of one `backup_set` means that, after projection rebuild, the restored deployment contains all authoritative structured source rows as of that `consistency_point_at`, all required blob bytes for any blob whose authoritative state at that point requires durable object bytes, and no evidence/blob invariant violations introduced by the restore. Projections, search indexes, sessions, presigned URLs, temporary work files, client-local drafts, export artifacts, and other deployment-local caches are not part of the authoritative backup set.
Profiles: base
Verified by: AC-398, AC-399

**REQ-01-572**
A `backup_set` counts as successful only when all of the following are durably captured for the same `consistency_point_at`:

1. one Postgres restore artifact set or restore anchor sufficient to restore authoritative structured state to that point,
2. one object-store restore artifact set or restore anchor sufficient to restore every required blob byte for that same point,
3. one durable `backup_attestation` record containing at least `backup_set_id`, `consistency_point_at`, `postgres_restore_anchor`, `object_store_restore_anchor`, `created_at`, `retained_until`, `verification_state`, and `last_verified_restore_at`.

At creation time, `verification_state` MUST be `unverified`. `verification_state` MUST use exactly `unverified`, `verified`, or `failed`. `last_verified_restore_at` MAY be `null` only while `verification_state='unverified'`.
Profiles: base
Verified by: AC-398, AC-401

**REQ-01-573**
The base profile MUST retain at least one successful retained `backup_set` whose `consistency_point_at` is no older than 24 hours, and each successful `backup_set` plus its restoreable artifacts MUST be retained for at least 30 days before disposal.
Profiles: base
Verified by: AC-398

**REQ-01-574**
Postgres base backup plus WAL archiving together with object-store bucket snapshotting or versioning is the RECOMMENDED realization. Another mechanism is conformant only if it produces a durable named `backup_set`, defines one declared `consistency_point_at` across Postgres and object storage, restores authoritative structured state and required blob bytes for that same point, preserves the same retention floor, passes the restore-verification contract in §12.2, and does not depend on projections as authoritative state. A live copy of the Postgres data directory without consistent snapshot semantics and a live object namespace without versioning, immutable snapshot semantics, or another independent restore anchor are not equivalent.
Profiles: base
Verified by: AC-398, AC-399, AC-401

### 12.2 Restore

**REQ-01-575**
A restore operation MUST select exactly one retained `backup_set` and MUST restore Postgres and object-store contents from that same `backup_set` and its declared `consistency_point_at`.
Profiles: base
Verified by: AC-399, AC-400

**REQ-01-423**
Restore MUST occur in this order for the selected `backup_set`:

1. restore Postgres,
2. restore object-store contents,
3. rebuild projections.
Profiles: base
Verified by: AC-399

**REQ-01-424**
Projection rebuild MUST be part of restore readiness when projection contents are not restored directly. Projection tables remain disposable caches and MUST NOT be required authoritative restore inputs.
Profiles: base
Verified by: AC-399

**REQ-01-576**
If the selected `backup_set` is missing a required Postgres artifact, required object-store artifact, or required checksum or integrity proof for the deployment's chosen backup mechanism, restore MUST fail before the environment is exposed as ready.
Profiles: base
Verified by: AC-400

**REQ-01-577**
The base profile MUST be able to restore the latest successful retained `backup_set`. A deployment MAY additionally support restore of an earlier retained `backup_set` or another retained `consistency_point_at`, but any such optional capability MUST restore Postgres and object-store contents to the same retained point and satisfy the same verification contract. The current profile does not require arbitrary cross-store point-in-time restore to an operator-supplied timestamp.
Profiles: base
Verified by: AC-399

**REQ-01-578**
A successful retained `backup_set` MUST undergo full restore verification in an isolated environment at least every 7 days and after any change to the backup mechanism, `roots.database_storage` binding, `roots.object_storage` binding, or `roots.backup_storage` binding. A successful verification MUST restore the selected `backup_set`, rebuild projections, satisfy authoritative evidence/blob lifecycle invariants, and, when the restored set contains incident data, successfully open at least one incident and execute at least one built-in workbook query. A successful verification MUST set `verification_state='verified'` and update `last_verified_restore_at`. A failed verification MUST set `verification_state='failed'` and update `last_verified_restore_at`.
Profiles: base
Verified by: AC-401

### 12.3 Incident portability

Operational backup and restore, `backup_set`, `backup_attestation`, restore anchors, and restore-verification state are deployment-local operational state. They are not incident-portability content.

Whole-incident export/import beyond operational backup and restore belongs to the **Incident Portability Extension Profile**.

**REQ-01-425**
If the implementation claims that profile, it MUST support full-fidelity administrative round-trip transfer of authoritative incident source state between trusted Cartulary deployments without depending on workbook-label semantics, live remote fetches, or deployment-local authentication configuration.
Profiles: incident_portability
Verified by: AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-236

**REQ-01-426**
Import into an existing incident, incident cloning with identifier remapping, and partial bundle merge semantics are out of scope for this profile. A conformant import MUST preserve the exported `incident_id`, `record_id`, `row_version`, change-set identifiers, and blob hashes.
Profiles: incident_portability
Verified by: AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-236

**REQ-01-564**
A portability bundle MUST preserve enough authoritative history substrate for the importing deployment to materialize conformant `GET /api/v1/records/{record_id}/history` results and conformant rollback behavior for imported records. Exact byte preservation of opaque `history_entry_ref` values is not part of the portability contract. The importing deployment MAY reissue `history_entry_ref` values, but once issued there they MUST be stable for the retained-history lifetime of the imported record in that deployment.
Profiles: incident_portability
Verified by: AC-236, AC-386

#### 12.3.1 Logical bundle contract

**REQ-01-427**
The canonical portability artifact MUST be a logical bundle layout. A `.zip` or `.tar` wrapper MAY be used for transport, but the normative contract is the root directory structure and file contents after extraction.
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-169, AC-236

**REQ-01-428**
At minimum, the logical bundle root MUST contain:
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-169, AC-236

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

**REQ-01-429**
Bundle member paths MUST use relative forward-slash separators. The logical bundle and any outer archive wrapper MUST reject absolute paths, `.` or `..` segments, symlinks, hard links, device nodes, and other member types outside regular files and directories.
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-169, AC-236

**REQ-01-430**
Each blob stored under `blobs/sha256/<sha256-lower-hex>` MUST contain the exact raw bytes whose SHA-256 digest matches the lowercase hex path suffix.
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-169, AC-236

#### 12.3.2 Authoritative-state boundary

**REQ-01-431**
Portability export MUST serialize authoritative incident source state and blob bytes. It MUST NOT serialize derived or deployment-local runtime state.
Profiles: incident_portability
Verified by: AC-164, AC-167, AC-236

**REQ-01-432**
A portability bundle MUST NOT include:

- backup artifacts, `backup_attestation` records, restore anchors, restore-verification state, or other deployment-local operational recovery metadata,
- projection tables or search indexes,
- live presence state,
- client-local draft queues or same-field-conflict queues,
- sessions, presigned URLs, locks, temporary caches, or other ephemeral runtime files,
- login-capable local user accounts, deployment-admin flags, auth-binding state, password hashes, MFA secrets, external provider configuration, or object-store credentials,
- incident memberships, current permissions, deployment-local administrative audit history, or other deployment-local authorization state.
Profiles: incident_portability
Verified by: AC-164, AC-167, AC-236

Reference-pack attestation metadata remains incident-external state. When the Incident Portability Extension Profile embeds reference-pack payloads, the bundle MAY include only the optional embedded-pack payloads and their bundle-local descriptors, not deployment-global activation or attestation history.

#### 12.3.3 Manifest and integrity contract

**REQ-01-433**
`manifest.json` MUST be the canonical bundle manifest and MUST include, at minimum:
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-236

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

**REQ-01-434**
`manifest.json` MUST describe one immutable export boundary. `source_change_set_high_watermark` MUST identify the frozen source boundary used to build the bundle. `history_mode` and `blob_mode` MUST each equal `full` for this profile.
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-236

**REQ-01-435**
`manifest.json.files[]` MUST enumerate every regular file in the logical bundle except `integrity/checksums.sha256` and `integrity/signature.ed25519`, sorted lexicographically by `path`. `required=true` MUST identify the files required to reconstruct the core incident state. `required=false` MAY be used only for optional embedded sections.
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-236

**REQ-01-436**
`integrity/checksums.sha256` MUST list one lowercase SHA-256 and relative path per line for every file listed in `manifest.json.files[]`, sorted lexicographically by path, using the exact file bytes carried in the bundle.
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-236

**REQ-01-437**
If `integrity/signature.ed25519` is present, `manifest.json.signing_key_id` MUST also be present. The signature MUST cover the exact bytes of `integrity/checksums.sha256`. If a deployment supports signature verification for portability bundles, signature failure MUST reject the bundle before any structured data becomes visible.
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-236

**REQ-01-438**
If `required_capabilities[]` names a capability the target deployment does not implement, import MUST fail closed. Optional embedded sections MAY be ignored when unsupported unless the corresponding capability is listed in `required_capabilities[]`.
Profiles: incident_portability
Verified by: AC-164, AC-166, AC-236

#### 12.3.4 Structured formats and deterministic serialization

**REQ-01-439**
The portability bundle MUST use JSON for singleton files and NDJSON for multi-row files. CSV, XLSX, or other workbook-shaped exports MAY exist elsewhere. They MUST NOT be the authoritative whole-incident portability format.
Profiles: incident_portability
Verified by: AC-164, AC-165, AC-236

**REQ-01-440**
All structured files in the bundle MUST use:

- UTF-8 encoding,
- LF line endings,
- no BOM,
- lexicographically sorted object keys,
- exactly one JSON object per NDJSON line,
- deterministic file-level row ordering.
Profiles: incident_portability
Verified by: AC-164, AC-165, AC-236

**REQ-01-441**
NDJSON files whose rows have a stable single-column primary key MUST sort ascending by that key. `record_tags.ndjson` MUST sort by `(record_id, tag_id)`. `change_set_mutations.ndjson` MUST sort by `(change_set_id, sequence_no)`. Files with integer primary keys such as `record_revisions.ndjson` MUST sort ascending by that integer key.
Profiles: incident_portability
Verified by: AC-164, AC-165, AC-236

**REQ-01-442**
The canonical JSON serialization for singleton JSON files and per-line NDJSON objects MUST be stable enough that exporting the same incident state twice without intervening mutations produces byte-identical structured files and identical `integrity/checksums.sha256`.
Profiles: incident_portability
Verified by: AC-164, AC-165, AC-236

#### 12.3.5 Portable actors and optional embedded sections

**REQ-01-443**
`actors.ndjson` MUST preserve historical attribution without exporting portable login material. Each actor descriptor row MUST carry, at minimum:

- stable `actor_id`,
- display name,
- non-secret match hints when available, such as normalized email or provider-subject hint.
Profiles: incident_portability
Verified by: AC-167, AC-168, AC-236

**REQ-01-444**
Import MUST materialize every actor referenced by imported history as either:

- an inert imported actor that is not login-capable and is not automatically added to incident membership, or
- a historical actor descriptor bound to an existing local user without rewriting the source bundle actor identifier used by imported history.
Profiles: incident_portability
Verified by: AC-167, AC-168, AC-236

**REQ-01-445**
`reference_pack_refs.json` MUST list the reference-pack keys and versions referenced by imported saved views, overlays, or optional embedded sections. Missing referenced packs MUST degrade only the affected overlays or saved views. They MUST NOT block import of the core incident state.
Profiles: incident_portability
Verified by: AC-167, AC-168, AC-236

**REQ-01-446**
If `reference_pack_mode='embedded'`, pack payloads MAY be embedded under `ext/reference_packs/**`. If `optional_sections` contains `snapshots`, immutable snapshot descriptors and rendered artifacts MAY be embedded under `ext/snapshots/**`. Unsupported or missing optional embedded sections MUST NOT block import of the core incident state unless the relevant capability is named in `required_capabilities[]`.
Profiles: incident_portability
Verified by: AC-167, AC-168, AC-236

#### 12.3.6 Export and import execution semantics

**REQ-01-447**
Whole-incident export and import MUST execute as background jobs rather than as blocking grid actions.
Profiles: incident_portability
Verified by: AC-165, AC-166, AC-167, AC-169, AC-236

**REQ-01-448**
A conformant import MUST execute the following phases in order:

1. stage the supplied outer archive or logical bundle under the configured temporary-work root,
2. validate the outer container, every member path, extracted regular-file byte total, extracted regular-file member count, and applicable archive compression ratio against `limits.incident_bundles.max_extracted_bytes`, `limits.archives.max_compression_ratio`, and `limits.archives.max_members`,
3. verify every required checksum and any supported signature before any structured data becomes visible,
4. stage blob bytes and verify every required blob hash,
5. import the structured incident state,
6. rebuild projections,
7. mark the imported incident visible only after the structured import and projection rebuild succeed.
Profiles: incident_portability
Verified by: AC-165, AC-166, AC-167, AC-169, AC-236, AC-327, AC-328, AC-332

**REQ-01-449**
Import MUST fail closed on any of the following:

- checksum mismatch,
- signature mismatch when signature verification is supported or required by deployment policy,
- missing required file,
- missing required blob or blob-hash mismatch,
- invalid path or unsupported bundle member type,
- extracted regular-file bytes exceeding `limits.incident_bundles.max_extracted_bytes`,
- extracted regular-file bytes exceeding `compressed_bytes * limits.archives.max_compression_ratio`,
- extracted regular-file member count exceeding `limits.archives.max_members`,
- unsupported `required_capabilities[]` entry,
- duplicate `incident_id`,
- any import path that would require a live remote fetch to complete.
Profiles: incident_portability
Verified by: AC-165, AC-166, AC-167, AC-169, AC-236, AC-327, AC-328, AC-332

**REQ-01-450**
If import fails after staging begins, the target deployment MUST leave no partially visible incident. Staged bytes MAY be retained only in a non-visible administrative quarantine or temporary-work area.
Profiles: incident_portability
Verified by: AC-165, AC-166, AC-167, AC-169, AC-236

### 12.4 Failure handling

**REQ-01-451**
The implementation MUST satisfy all of the following failure semantics:

- if the application container is unavailable, sessions MAY drop but committed data MUST remain durable,
- if Postgres is unavailable, the system MAY become unavailable,
- if object storage is unavailable, row editing MUST remain possible but evidence upload and download MUST fail clearly,
- if projections are unavailable or corrupt, the implementation MUST preserve source data integrity and rebuild projections from source state.
Profiles: base
Verified by: AC-166, AC-231

## 13. Long-running operations and background jobs

**REQ-01-452**
The implementation MUST execute the following long-running operations as background jobs rather than as blocking grid actions:

- lookups beyond trivial inline suggestion queries,
- imports,
- incident portability export and import when the Incident Portability Extension Profile is implemented,
- reference-pack import, verification, and refresh,
- snapshot generation,
- report builds,
- projection rebuilds,
- evidence processing, including blob hashing, scanning, preview generation, thumbnailing, and metadata extraction.
Profiles: base, snapshot_reporting, reference_pack
Verified by: AC-030, AC-033, AC-046, AC-129, AC-169, AC-231, AC-233, AC-234

**REQ-01-453**
Background jobs MUST expose:

- progress,
- cancellation,
- retry-safe status,
- non-blocking UI behavior.

For the public HTTP and WebSocket surface, the exact contract for `progress`, cancellation, and retry-safe status is owned by §3.3.9.1.
Profiles: base
Verified by: AC-030, AC-033, AC-046, AC-129, AC-169, AC-231, AC-258, AC-260

**REQ-01-454**
Grid editing and row creation MUST remain responsive while these jobs run. These jobs MUST NOT block row selection, sheet filtering, sorting, grouping, or inspector metadata open.
Profiles: base
Verified by: AC-030, AC-033, AC-046, AC-129, AC-169, AC-231

## 14. Runtime roots and packaging

**REQ-01-455**
The application runtime MUST obtain database storage, object storage, reference-pack storage, temporary-work, and export-output roots from the deployment configuration contract owned by Core 04 §12. Core 01 does not define the operator-facing configuration artifact, discovery precedence, key registry, default locations, or validation error contract.
Profiles: base, reference_pack
Verified by: AC-051, AC-055, AC-169, AC-231, AC-234, AC-294, AC-295, AC-297

**REQ-01-456**
Packaged read-only resources MAY resolve from install or package locations. Any operator-owned writable or persistent location MUST derive from the deployment configuration contract and MUST NOT rely on source-tree-relative or current-working-directory defaults for icons, Markdown templates, reference data, or generated artifacts.
Profiles: base
Verified by: AC-051, AC-055, AC-169, AC-231, AC-296

## 15. Architecture invariants

**REQ-01-457**
An implementation conforming to this core MUST preserve all of the following:

1. the complexity budget belongs in mutation semantics, projections, and workbook UX rather than distributed infrastructure,
2. derived surfaces MUST share a canonical derivation layer,
3. optional enrichment MUST remain off the hot capture path,
4. the object store boundary MUST remain explicit and lifecycle-aware,
5. view behavior MUST remain contract-driven,
6. projection tables MUST remain disposable derived state,
7. file-based import complexity MUST remain isolated behind the imports module and stable tabular-ingest contract rather than leaking spreadsheet-specific parser behavior into workbook-domain modules,
8. whole-incident portability, when implemented, MUST export authoritative source state and referenced blob bytes rather than projections, snapshots, or deployment-local runtime state.
Profiles: base, import, snapshot_reporting, incident_portability
Verified by: AC-231, AC-232, AC-233, AC-236


## 16. Evidence-access handle contract

**REQ-01-458**
This section owns the base-profile public contract for `POST /api/v1/evidence-records/{record_id}/preview-handle`, `POST /api/v1/evidence-records/{record_id}/download-handle`, and `GET /api/v1/evidence-handles/{handle_token}`. A successful issuance response MUST return `data.href` as an opaque same-origin redeem URL under `GET /api/v1/evidence-handles/{handle_token}`, and clients MUST NOT synthesize or parse that token. The server MAY satisfy redeem by streaming bytes itself or by performing an internal one-time redirect after redeem-time validation, but the public contract MUST NOT expose long-lived object-store credentials, bucket names, raw object keys, or storage-backend-specific identifiers.
Profiles: base
Verified by: AC-231, AC-252, AC-253, AC-254

**REQ-01-459**
Both issuance routes MUST accept only a JSON object request body. `{}` MUST be legal and means issue the default contract-defined preview or download handle for the addressed evidence. A zero-length body, `null`, any non-object JSON value, or any unknown top-level member MUST fail with `400` and `error.code='invalid_evidence_handle_request'`. The base profile defines no request members for either route yet. Accordingly, `client_txn_id` is invalid on both issuance routes and MUST fail as an unknown top-level member rather than being interpreted as an issuance idempotency key. If a future additive request member is introduced, omission MUST mean default behavior for that member, and explicit `null` MUST remain invalid unless that member explicitly allows `null`.
Profiles: base
Verified by: AC-231, AC-251, AC-255

**REQ-01-460**
A successful issuance response from either route MUST use the standard success envelope from §3.3.6 and include `data.incident_id`, `data.record_id`, `data.object_blob_id`, `data.handle_kind`, `data.href`, `data.method`, `data.expires_at`, `data.single_use`, `data.media_class`, `data.disposition`, `data.filename`, `data.content_type`, `data.size_bytes`, `data.sha256`, `data.evidence_lifecycle_state`, and `data.upload_state`. `data.method` MUST be `GET`. `data.media_class` MUST use the exact tokens owned by Core 02 §18. `data.sha256` MAY be `null`; all other listed members are required. Each successful issuance call MUST return a fresh handle. Repeating the same issuance request is not idempotent replay, and the base profile MUST NOT require or interpret `client_txn_id` or any other issuance idempotency key for these routes.
Profiles: base
Verified by: AC-231, AC-252, AC-253, AC-256

**REQ-01-461**
A successful preview-handle issuance MUST set `data.handle_kind='preview'`, `data.single_use=false`, `data.disposition='inline'`, and a non-null `data.preview_kind` that uses the exact tokens owned by Core 02 §18. In the base profile, preview issuance MUST succeed only when `data.preview_kind` is one of `image_inline`, `pdf_inline`, or `text_inline`, when `data.size_bytes <= limits.previews.max_previewable_payload_bytes`, and, for `data.preview_kind='text_inline'`, when `data.size_bytes <= limits.previews.max_text_inline_bytes`. Preview handles MUST expire exactly 5 minutes after issuance and MUST be reusable until expiry, including repeated byte-range fetches made by a browser preview surface. The server MUST NOT silently downgrade preview issuance into a download contract. When the evidence is otherwise visible but the base-profile preview allowlist does not allow a safe preview, the route MUST fail with `409`, `error.code='evidence_access_unavailable'`, and `error.details.reason_code='unsupported_preview'`. When the evidence is otherwise visible but the payload exceeds the configured preview-size ceiling for the requested preview contract, the route MUST fail with `409`, `error.code='evidence_access_unavailable'`, and `error.details.reason_code='preview_payload_too_large'`. Download-handle issuance remains legal when preview is blocked solely by preview-size limits.
Profiles: base
Verified by: AC-231, AC-252, AC-322

Example preview-handle success payload:

```json
{
  "data": {
    "incident_id": "7d4cc0c6-8081-4b52-a7a8-3c2577fe5f7e",
    "record_id": "5ad0f785-b814-4bd4-aee4-3c39769357a3",
    "object_blob_id": "d1968f09-fd8c-4ca5-b8ea-1988931b6307",
    "handle_kind": "preview",
    "href": "/api/v1/evidence-handles/hdl_01JQ8Y9AB3Q4WE7K3S8M0P6A6V",
    "method": "GET",
    "expires_at": "2026-03-27T16:00:00Z",
    "single_use": false,
    "media_class": "image",
    "preview_kind": "image_inline",
    "disposition": "inline",
    "filename": "signin.png",
    "content_type": "image/png",
    "size_bytes": 188416,
    "sha256": "c0c4f0a4e3c49f6f07a8e8ca1d0cf1ff25ed90d0f6d619fd7f3f8ea70f58de17",
    "evidence_lifecycle_state": "available",
    "upload_state": "available"
  },
  "meta": {
    "request_id": "req_01JQ8Y9AB3Q4WE7K3S8M0P6A6V"
  }
}
```

**REQ-01-462**
A successful download-handle issuance MUST set `data.handle_kind='download'`, `data.single_use=true`, and `data.disposition='attachment'`. `data.preview_kind` MUST be absent from the response. Download handles MUST expire exactly 2 minutes after issuance. The base profile MUST NOT accept caller-controlled filename or disposition overrides on this route. A single-use download handle becomes consumed on the first successful redeem that starts byte delivery through `200`, `206`, or a validated internal redirect; a failed redeem that emits no bytes MUST NOT consume the handle. Resuming an interrupted download after a successful redeem requires a fresh download handle.
Profiles: base
Verified by: AC-231, AC-253, AC-254

Example download-handle success payload:

```json
{
  "data": {
    "incident_id": "7d4cc0c6-8081-4b52-a7a8-3c2577fe5f7e",
    "record_id": "5ad0f785-b814-4bd4-aee4-3c39769357a3",
    "object_blob_id": "d1968f09-fd8c-4ca5-b8ea-1988931b6307",
    "handle_kind": "download",
    "href": "/api/v1/evidence-handles/hdl_01JQ8Y9CA9M6N8D4J1P3V5S7T9",
    "method": "GET",
    "expires_at": "2026-03-27T15:57:00Z",
    "single_use": true,
    "media_class": "image",
    "disposition": "attachment",
    "filename": "signin.png",
    "content_type": "image/png",
    "size_bytes": 188416,
    "sha256": "c0c4f0a4e3c49f6f07a8e8ca1d0cf1ff25ed90d0f6d619fd7f3f8ea70f58de17",
    "evidence_lifecycle_state": "available",
    "upload_state": "available"
  },
  "meta": {
    "request_id": "req_01JQ8Y9CA9M6N8D4J1P3V5S7T9"
  }
}
```

**REQ-01-463**
Every redeem of `GET /api/v1/evidence-handles/{handle_token}` MUST re-check current session validity, current incident membership, current evidence or blob accessibility state, and handle freshness at redeem time. A handle MUST be bound, at minimum, to the issuing session, incident, `record_id`, `object_blob_id`, `handle_kind`, resolved `filename`, and `disposition`; preview handles MUST also bind `preview_kind`. A handle issued before logout, session expiry, incident-membership loss, blob detach or replacement, evidence delete or restore, quarantine, pending or failed blob transition, or detected evidence/blob inconsistency MUST fail closed when redeemed later.
Profiles: base
Verified by: AC-231, AC-254, AC-255

**REQ-01-464**
`data.filename` and any corresponding redeem header filename parameter MUST derive from authoritative object metadata, never from caller input and never from storage keys. The server MUST sanitize `/`, `\`, NUL, carriage return, and line feed, and MUST prevent path-like segments from surviving sanitization. If the authoritative filename is empty or unusable after sanitization, the fallback MUST be deterministic and use `evidence-<record_id><canonical_extension_if_known>`. Preview redeem MUST emit `Content-Disposition: inline`; download redeem MUST emit `Content-Disposition: attachment`. The actual header SHOULD include both an ASCII-safe `filename=` parameter and a Unicode-preserving `filename*=` parameter. The JSON issuance response MUST expose only `filename` and `disposition`, not a pre-rendered header string.
Profiles: base
Verified by: AC-231, AC-256

**REQ-01-465**
Issuance MUST use `invalid_evidence_handle_request`, `evidence_record_not_found`, and `evidence_access_unavailable` from §3.3.6.1. Redemption MUST use `handle_not_found_or_revoked`, `handle_expired`, `handle_consumed`, and `evidence_access_unavailable`. Standard authentication or session failures MUST occur before handle-specific lookup and MUST use the ordinary authentication envelope rather than a handle-specific code. Whenever `evidence_access_unavailable` is used on issuance or redemption, `error.details.reason_code` MUST use the exact `evidence_access_unavailable` registry from §3.3.6.2. `preview_payload_too_large` is reserved for preview-size rejections under REQ-01-461 and MUST NOT be used for download-handle issuance.
Profiles: base
Verified by: AC-231, AC-251, AC-252, AC-253, AC-254, AC-255, AC-322

Example blocked preview response:

```json
{
  "error": {
    "status": 409,
    "code": "evidence_access_unavailable",
    "message": "Preview is not available for this evidence.",
    "retryable": false,
    "details": {
      "record_id": "5ad0f785-b814-4bd4-aee4-3c39769357a3",
      "reason_code": "unsupported_preview"
    }
  },
  "meta": {
    "request_id": "req_01JQ8Y9DA1N7R2C5V4M8K6X0P2"
  }
}
```


## 17. Extension route-family public contracts

### 17.1 Common parity rules

**REQ-01-466**
If an implementation claims the Import Extension Profile, Snapshot and Reporting Extension Profile, Reference Pack Extension Profile, or Incident Portability Extension Profile, it MUST implement that family's public route contract exactly as defined in this section in addition to the underlying model and lifecycle requirements defined elsewhere in the core.
Profiles: import, snapshot_reporting, incident_portability, reference_pack
Verified by: AC-262, AC-263, AC-264, AC-265, AC-266, AC-267, AC-268, AC-269, AC-270, AC-271, AC-272, AC-273, AC-274, AC-275, AC-276

Contract tables. The tables in §17 are the compact owner-local route-family contract for extension parity. They do not introduce new runtime behavior. They make route inventory, omission and default rules, idempotency scope, durable resource shape, and family-owned terminal results inspectable without requiring the reader to reconstruct them from long prose.

**Table 17.1-A. Extension-family parity table**

| Family | Reserved root(s) | Mutating routes require `client_txn_id` | Upload envelope | Long-running completion | Family-owned durable outputs |
| --- | --- | --- | --- | --- | --- |
| Import | `/api/v1/import-sessions` | Yes | Yes for `POST /api/v1/import-sessions` | Discovery and apply use the common job resource | `import_session` resource |
| Snapshot and Reporting | `/api/v1/snapshots`, `/api/v1/releases` | Yes | No | Snapshot create and release create use the common job resource; release approve, publish, and invalidate are synchronous | `snapshot` and `release` resources |
| Reference Pack | `/api/v1/reference-packs` | Yes | Yes for `POST /api/v1/reference-packs/import` | Import, reverify, and refresh are background jobs; activate and disable may be sync or backgrounded | `reference_pack_version` resource |
| Incident Portability | `/api/v1/incident-bundles` | Yes | Yes for `POST /api/v1/incident-bundles/import` | Export and import use the common job resource | `incident_bundle` export descriptor on export; imported `incident` on success |


**REQ-01-467**
Any extension-family action in this section that performs long-running work MUST return `202 Accepted` with the common job resource defined in §3.3.9 and §3.3.9.1. The public job-status vocabulary for those routes remains exactly `queued`, `running`, `cancel_requested`, `succeeded`, `failed`, and `canceled`. Durable family resources and durable family state fields MUST remain separate from that six-token job-status vocabulary.
Profiles: import, snapshot_reporting, incident_portability, reference_pack
Verified by: AC-262, AC-264, AC-266, AC-267, AC-268, AC-270, AC-271, AC-273, AC-274, AC-275, AC-309, AC-369

**REQ-01-468**
Extension-family list routes defined in this section MUST use the common cursor-pagination contract in §3.3.7. Extension-family singleton reads and extension-family action routes defined in this section MUST reject `limit`, `cursor_token`, and pagination aliases with `400`, `error.code = invalid_pagination_request`, and `error.details.reason_code = pagination_not_supported` rather than silently ignoring them.
Profiles: import, snapshot_reporting, incident_portability, reference_pack
Verified by: AC-263, AC-266, AC-270, AC-274

**REQ-01-469**
For the JSON request bodies and JSON metadata parts defined in this section, omission means the exact declared default only when this section explicitly declares a default. Otherwise a required member is missing and explicit JSON `null` is invalid unless this section explicitly allows `null` for that member. Every optional member declared by a route family in this section MUST also declare its omission meaning, explicit-`null` behavior, empty-array behavior when the member is an array, duplicate handling when the member is an array, whether array order is semantic, and the canonical normalization used for idempotency comparison. When omission resolves dynamically from current incident or reference-pack state, the server MUST resolve that omission once at route admission to one concrete value or one concrete value set and MUST reuse that resolved value for idempotency comparison, replay, and any durable resource or descriptor later emitted by that route family. Versioned identifiers such as `template_version`, `redaction_profile_version`, and `pack_version` MUST use exact values; extension routes in this section MUST NOT accept `latest`, `current`, display-label resolution, or equivalent implicit version selectors.
Profiles: import, snapshot_reporting, incident_portability, reference_pack
Verified by: AC-262, AC-263, AC-264, AC-266, AC-267, AC-268, AC-270, AC-271, AC-273, AC-274, AC-275, AC-305, AC-308, AC-369

**REQ-01-470**
Every mutating extension-family control-plane route defined in this section MUST require `client_txn_id` and MUST apply route-scoped idempotency keyed by the authenticated actor, the addressed family resource identity or incident scope, and the normalized request contract for that route.
Profiles: import, snapshot_reporting, incident_portability, reference_pack
Verified by: AC-262, AC-264, AC-266, AC-267, AC-268, AC-270, AC-271, AC-273, AC-275, AC-305, AC-308, AC-369

**REQ-01-471**
Extension-family routes in this section MUST use only the family-specific `error.code` tokens and `reason_code` registries added to §3.3.6.1 and §3.3.6.2 for that family. Successful terminal `result_summary.code` values for those routes are family-owned and MUST be declared in the owning family subsection below. Canceled terminal job summaries for those routes MUST use the common `job_canceled` code from §3.3.9.1 rather than family-specific or ad hoc worker strings.
Profiles: import, snapshot_reporting, incident_portability, reference_pack
Verified by: AC-265, AC-269, AC-272, AC-276, AC-307, AC-310

#### 17.1.1 Shared upload-envelope contract for upload-style extension routes


**Table 17.1.1-A. Shared upload-envelope contract**

| Envelope concern | Requirement |
| --- | --- |
| Media type | Only `multipart/form-data` with required `boundary` |
| Required parts | Exactly one `metadata` part and exactly one `file` part; part order is non-semantic |
| `metadata` part | `Content-Disposition: form-data; name="metadata"`; `Content-Type` is `application/json` or `application/json; charset=utf-8`; UTF-8 and BOM-free; parses as exactly one JSON object; duplicate JSON member names are rejected |
| `file` part | `Content-Disposition: form-data; name="file"`; exactly one uploaded payload; advisory filename has no semantic effect |
| Media-type allowlist for `file` part | Route-local and byte-validation remains authoritative; a missing or unsupported file `Content-Type` fails with the family `invalid_part_content_type` reason |
| Early-fail behavior | Missing required part, duplicate part, unexpected part, nested multipart, malformed metadata JSON, or invalid metadata encoding fail before durable resource creation, idempotency commit, or job creation |
| Shared `reason_code` subset | `unsupported_upload_envelope`, `missing_required_part`, `duplicate_part`, `unexpected_part`, `invalid_part_content_type`, `invalid_metadata_encoding`, `malformed_metadata_json` |


**REQ-01-549**
`POST /api/v1/import-sessions`, `POST /api/v1/reference-packs/import`, and `POST /api/v1/incident-bundles/import` are the only current-profile upload-style extension routes. Each of those routes MUST accept only `multipart/form-data` with a required `boundary` parameter. No alternate v1 upload framing is valid for those routes, including raw binary request bodies with metadata headers, JSON bodies containing base64 file content, JSON-only metadata bodies, or nested multipart bodies. Part order is non-semantic. Each request MUST contain exactly two leaf parts named `metadata` and `file`. `metadata` MUST appear exactly once. `file` MUST appear exactly once. Any missing required part, duplicate required part, unexpected extra part, unsupported upload envelope, or nested multipart body MUST fail closed before durable resource creation, idempotency commit, or background-job creation.
Profiles: import, incident_portability, reference_pack
Verified by: AC-262, AC-270, AC-275

**REQ-01-550**
For those routes, the `metadata` part MUST use `Content-Disposition: form-data; name="metadata"`. Its `Content-Type` MUST be `application/json` with no parameters, or `application/json` plus exactly one `charset` parameter whose value after ASCII case-folding is `utf-8`. Metadata bytes MUST be UTF-8 and BOM-free. The part MUST parse as exactly one JSON object. Duplicate JSON member names MUST be rejected. A syntactically valid JSON value that is not an object MUST fail through the existing family `request_not_object` path. An optional multipart `filename` parameter on the `metadata` part is ignored and has no semantic effect.
Profiles: import, incident_portability, reference_pack
Verified by: AC-262, AC-270, AC-275

**REQ-01-551**
For those routes, the `file` part MUST use `Content-Disposition: form-data; name="file"` and MUST carry exactly one uploaded payload. Any multipart `filename` parameter on the `file` part is advisory only. The server MAY preserve a normalized original filename where the owning route already exposes it, but multipart boundary text, part order, advisory filename, part-header order, and other non-semantic part headers or parameters MUST NOT participate in normalized idempotency comparison and MUST NOT establish file-kind trust. File-kind validation remains route-local and byte-based.
Profiles: import, incident_portability, reference_pack
Verified by: AC-262, AC-270, AC-275

**REQ-01-552**
For file-part media-type matching on those routes, the server MUST compare the `Content-Type` media type after ASCII case-folding of the type and subtype and after discarding any parameters. A missing `Content-Type`, or a media type outside the route-local allowlist below, MUST fail with the family's `invalid_part_content_type` reason.

| Route | Allowed `file` part media types |
| --- | --- |
| `POST /api/v1/import-sessions` | `text/csv`, `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, `application/octet-stream` |
| `POST /api/v1/reference-packs/import` | `application/zip`, `application/x-tar`, `application/gzip`, `application/x-gzip`, `application/octet-stream` |
| `POST /api/v1/incident-bundles/import` | `application/zip`, `application/x-tar`, `application/gzip`, `application/x-gzip`, `application/octet-stream` |
Profiles: import, incident_portability, reference_pack
Verified by: AC-262, AC-270, AC-275

**REQ-01-553**
These routes MUST use the existing family `invalid_*_request` codes for upload-envelope failures rather than introducing a new top-level error code. The shared upload-envelope `reason_code` subset is exactly `unsupported_upload_envelope`, `missing_required_part`, `duplicate_part`, `unexpected_part`, `invalid_part_content_type`, `invalid_metadata_encoding`, and `malformed_metadata_json`. When one named part is implicated, `error.details.part_name` MUST be present and MUST equal exactly `metadata` or `file`. For `invalid_part_content_type`, `error.details.received_content_type` MUST echo the received header value or JSON `null` when absent. If `part_name='metadata'`, `error.details.allowed_content_types[]` MUST equal `['application/json', 'application/json; charset=utf-8']` in that canonical order. If `part_name='file'`, `error.details.allowed_content_types[]` MUST list the exact route-local file allowlist from REQ-01-552 in canonical ascending order. After metadata JSON parsing succeeds, route-local validation continues to use existing family reasons such as `request_not_object`, `missing_required_field`, `field_not_nullable`, and `unknown_field` rather than creating upload-specific aliases.
Profiles: import, incident_portability, reference_pack
Verified by: AC-262, AC-265, AC-270, AC-272, AC-275, AC-276

### 17.2 Import Extension Profile public contract

**REQ-01-472**
The Import Extension Profile MUST expose exactly this minimum public route surface under `/api/v1/import-sessions/*`:

- `POST /api/v1/import-sessions`,
- `GET /api/v1/import-sessions/{import_session_id}`,
- `GET /api/v1/import-sessions/{import_session_id}/units`,
- `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}`,
- `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/preview`,
- `PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping`,
- `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select`,
- `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip`,
- `POST /api/v1/import-sessions/{import_session_id}/apply`.
Profiles: import
Verified by: AC-262, AC-263, AC-264

**Table 17.2-A. Import route inventory**

| Route | Request contract summary | Success resource or body | Long-running | Primary family errors |
| --- | --- | --- | --- | --- |
| `POST /api/v1/import-sessions` | Shared upload envelope with metadata `incident_id`, `client_txn_id`, and optional `assistant_profile` defaulting to `phase2_workbook_import_v1` | Common job resource; terminal success emits one `import_session` ref | Yes | `invalid_import_request`, `import_source_unsupported`, `import_source_rejected` |
| `GET /api/v1/import-sessions/{import_session_id}` | Singleton read | `import_session` resource | No | `import_session_not_found`, `invalid_pagination_request` |
| `GET /api/v1/import-sessions/{import_session_id}/units` | List read under common paging | `{ import_units[] }` plus `meta.paging` | No | `import_session_not_found` |
| `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}` | Singleton read | `import_unit` resource | No | `import_session_not_found`, `import_unit_not_found` |
| `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/preview` | Singleton read | `import_preview` resource | No | `import_session_not_found`, `import_unit_not_found` |
| `PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping` | JSON object with required `client_txn_id`, target mapping metadata, and exhaustive `source_columns[]` | `import_unit` resource | No | `invalid_import_request`, `import_state_conflict` |
| `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select` | JSON object with required `client_txn_id` | `{ import_session_id, session_status, selected_unit_ids[], unit }` | No | `import_state_conflict` |
| `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip` | JSON object with required `client_txn_id`; optional `reason` | `{ import_session_id, session_status, selected_unit_ids[], unit }` | No | `import_state_conflict` |
| `POST /api/v1/import-sessions/{import_session_id}/apply` | JSON object with required `client_txn_id` and optional `selected_unit_ids[]`; omitted `selected_unit_ids[]` means the session's persisted `selected_unit_ids[]` | Common job resource; terminal success emits one `import_session` ref | Yes | `invalid_import_request`, `import_apply_blocked`, `import_state_conflict` |

**Table 17.2-B. Import durable resources**

| Resource | Required members or properties |
| --- | --- |
| `import_session` | `import_session_id`, `incident_id`, `created_by_user_id`, `created_at`, `source_file_kind`, `original_filename`, `source_content_sha256`, `parser_profile_id`, `parser_version`, `assistant_profile`, `session_status`, `selected_unit_ids[]`, `blocking_diagnostics[]`, `nonblocking_warning_codes[]` |
| `import_unit` | `import_unit_id`, `import_session_id`, `locator_kind`, `locator`, `source_rect_a1`, `header_row_ref`, `data_start_row_ref`, `inferred_row_count`, `inferred_column_count`, `warning_codes[]`, `unit_status`, optional `mapping_fingerprint`, optional `approved_mapping` |
| `approved_mapping` | `target_view_schema_id`, `unknown_column_policy`, exhaustive ordered `source_columns[]`; `field_key = null` means intentionally unmapped |
| `import_preview` | Top-level session and unit identity plus `columns[]`, `preview_rows[]`, and `truncated`; preview returns at most the first 50 data rows in source order |

**Table 17.2-C. Import terminal results and primary error registries**

| Route or condition | Required code or registry |
| --- | --- |
| Discovery success | `result_summary.code='import_session_discovered'` and exactly one `import_session` ref |
| Apply success with `session_status='applied'` | `result_summary.code='import_session_applied'` and exactly one `import_session` ref |
| Apply success with `session_status='partially_applied'` | `result_summary.code='import_session_partially_applied'` and exactly one `import_session` ref |
| Invalid request registry | `invalid_import_request` with shared upload-envelope reasons plus the import-specific malformed-request reasons in REQ-01-475 |
| Source unsupported registry | `import_source_unsupported` with `encrypted_or_unparseable_workbook`, `unsupported_named_range`, and `formula_cached_value_missing` |
| Source rejected registry | `import_source_rejected` with size and archive-limit reasons owned by REQ-01-475 |
| Apply blocked registry | `import_apply_blocked` with `overlapping_units`, `duplicate_apply_blocked`, and `unit_not_ready` |


**REQ-01-473**
`POST /api/v1/import-sessions` MUST use the shared upload-envelope contract in §17.1.1. Within that contract, metadata MUST include required `incident_id` and required `client_txn_id`. Metadata MAY include optional `assistant_profile`, which defaults to `phase2_workbook_import_v1` when omitted and MUST use that exact value when supplied in the current profile. For this route, the `file` part media type MUST be one of the exact values declared for `POST /api/v1/import-sessions` in REQ-01-552. Those media-type values are necessary but not sufficient: the server MUST still determine CSV versus XLSX from the exact uploaded bytes and MUST enforce the route's byte-based parser and source-limit rules. Before `import_session` creation or discovery-job creation, the server MUST compare uploaded source bytes against `limits.imports.max_csv_source_bytes` for CSV and `limits.imports.max_xlsx_source_bytes` for XLSX. A CSV source that exceeds its ceiling MUST fail with `413`, `error.code='import_source_rejected'`, and `error.details.reason_code='csv_source_too_large'`. An XLSX source that exceeds its ceiling MUST fail with `413`, `error.code='import_source_rejected'`, and `error.details.reason_code='xlsx_source_too_large'`. Those rejections MUST create no durable `import_session`, no idempotency commit, and no discovery job. For an accepted source, the route MUST compute `source_content_sha256` from the exact uploaded file bytes, create or replay exactly one durable `import_session`, and start discovery as a background job. Normalized request comparison for idempotency MUST include `incident_id`, normalized `assistant_profile`, and the computed `source_content_sha256` from the exact uploaded file bytes. Multipart boundary text, part order, advisory filename, and non-semantic part headers or parameters MUST NOT affect normalized comparison.
Profiles: import
Verified by: AC-262, AC-323

**REQ-01-474**
The import route family MUST use the common success envelope and the following route-specific `data` shapes:

- `GET /api/v1/import-sessions/{import_session_id}` returns `data = <import_session resource>`.
- `GET /api/v1/import-sessions/{import_session_id}/units` returns `data = { import_units[] }` plus `meta.paging`.
- `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}` returns `data = <import_unit resource>`.
- `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/preview` returns `data = <import_preview resource>`.
- `PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping` returns `data = <import_unit resource>`.
- `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select` and `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip` return `data = { import_session_id, session_status, selected_unit_ids[], unit }`, where `unit` uses the exact `import_unit resource` shape defined here.

The `import_session resource` MUST expose exactly:

- `import_session_id`,
- `incident_id`,
- `created_by_user_id`,
- `created_at`,
- `source_file_kind`,
- `original_filename`,
- `source_content_sha256`,
- `parser_profile_id`,
- `parser_version`,
- `assistant_profile`,
- `session_status`,
- `selected_unit_ids[]`,
- `blocking_diagnostics[]`,
- `nonblocking_warning_codes[]`.

For `import_session resource` serialization:

- all top-level members above are required and non-null in the current profile;
- `selected_unit_ids[]`, `blocking_diagnostics[]`, and `nonblocking_warning_codes[]` MUST always be present and MUST default to `[]` when empty;
- each `blocking_diagnostics[]` item MUST contain exactly `code`, `reason_code`, `message`, and optional `import_unit_id`;
- `import_unit_id` MUST be absent for session-scoped blockers and MUST be present only for unit-scoped blockers.

The `import_unit resource` MUST expose exactly:

- `import_unit_id`,
- `import_session_id`,
- `locator_kind`,
- `locator`,
- `source_rect_a1`,
- `header_row_ref`,
- `data_start_row_ref`,
- `inferred_row_count`,
- `inferred_column_count`,
- `warning_codes[]`,
- `unit_status`,
- optional `mapping_fingerprint`,
- optional `approved_mapping`.

Each `data.import_units[]` entry returned by `GET /api/v1/import-sessions/{import_session_id}/units` MUST use that exact `import_unit resource` shape.

For `import_unit resource` serialization:

- all top-level members above other than `mapping_fingerprint` and `approved_mapping` are required and non-null in the current profile;
- `warning_codes[]` MUST always be present and MUST default to `[]` when empty;
- `mapping_fingerprint` and `approved_mapping` MUST both be absent until mapping approval and MUST both be present after mapping approval;
- `header_row_ref` and `data_start_row_ref` MUST be positive 1-based row references within `source_rect_a1`;
- `data_start_row_ref` MUST be greater than or equal to `header_row_ref + 1`;
- preview rows, preview columns, and `truncated` are not members of the durable `import_unit resource`.

When present, `approved_mapping` MUST expose exactly:

- `target_view_schema_id`,
- `unknown_column_policy`,
- `source_columns[]`.

Each `approved_mapping.source_columns[]` item MUST contain exactly:

- `source_column_ordinal`,
- `source_header_text`,
- `field_key`,
- `entity_binding_mode`,
- `transform_id`,
- `transform_options`,
- `empty_value_policy`.

For `approved_mapping` serialization:

- `source_columns[]` MUST be exhaustive over discovered source columns and MUST be ordered by `source_column_ordinal`;
- `transform_options` MUST always be an object;
- `field_key = null` means intentionally unmapped;
- `entity_binding_mode` MUST be `null` for unmapped columns and non-entity targets;
- `empty_value_policy` MUST be `omit_field` when `field_key = null`.

`GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/preview` is a singleton read route and MUST remain read-only against incident state. The preview resource MUST expose exactly:

- top-level fields `import_session_id`, `import_unit_id`, `locator_kind`, `locator`, `source_rect_a1`, `header_row_ref`, `data_start_row_ref`, `inferred_row_count`, `inferred_column_count`, `warning_codes[]`, `unit_status`, optional `mapping_fingerprint`, `columns[]`, `preview_rows[]`, and `truncated`;
- `columns[]` items containing exactly `source_column_ordinal` and `source_header_text`;
- `preview_rows[]` items containing exactly `source_row_ref` and `cells[]`;
- `cells[]` items containing exactly `source_column_ordinal`, `display_text`, and `cell_kind`.

For preview serialization:

- `cell_kind` MUST use exactly `blank`, `string`, `number`, `boolean`, `datetime`, `formula_cached`, and `error_literal`.
- `display_text` MUST be a string. When `cell_kind = blank`, `display_text` MUST be `""`.
- `source_row_ref` MUST use the same 1-based row coordinate system within `source_rect_a1` as `header_row_ref` and `data_start_row_ref`.
- The server MUST return at most the first 50 data rows after `data_start_row_ref`, preserve source order, and set `truncated = true` when more preview rows exist.

Before any import unit enters `ready` or `applied`, and before any imported incident data becomes visible or applicable through apply, the imports module MUST enforce the bounded ingest contract driven by Core 04 §12.3.1. For CSV, only the raw source-byte ceiling from REQ-01-473 applies. For XLSX, the route family MUST additionally enforce `limits.imports.max_rows`, `limits.imports.max_columns`, `limits.imports.max_cells`, `limits.archives.default_max_extracted_bytes`, `limits.archives.max_compression_ratio`, and `limits.archives.max_members`, treating XLSX as the ZIP-backed workbook container defined by Core 04 §12.3.1. A breach of any of those limits MUST fail the affected route or job with `413` and `error.code='import_source_rejected'`.

`PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping` MUST accept only a JSON object request body and MUST require:

- `client_txn_id`,
- `target_view_schema_id`,
- `header_row_ref`,
- `data_start_row_ref`,
- `unknown_column_policy`,
- `source_columns[]`.

`source_columns[]` MUST contain exactly one ordered entry per discovered source column. Each entry MUST contain exactly:

- `source_column_ordinal`,
- `source_header_text`,
- `field_key`,
- `entity_binding_mode`,
- `transform_id`,
- `transform_options`,
- `empty_value_policy`.

For the current profile:

- `source_column_ordinal` MUST be 1-based, unique, contiguous, and serialized in ascending order.
- `source_header_text` MUST be the raw imported header text or `null` when the header cell is empty.
- `field_key = null` means intentionally unmapped.
- Duplicate non-null `field_key` values are invalid.
- `entity_binding_mode` MUST be present and MUST be `null` for unmapped columns and non-entity targets.
- `transform_options` MUST always be an object.
- `empty_value_policy` MUST always be present and MUST be `omit_field` when `field_key = null`.
- `unknown_column_policy` MUST use exactly `preserve_raw_capture`, `preserve_custom_attrs`, or `reject_if_unmapped`.
- `transform_id` MUST use exactly `null`, `trim_v1`, `collapse_whitespace_v1`, `lowercase_v1`, or `split_delimited_v1`.
- `empty_value_policy` MUST use exactly `omit_field` or `write_null`.
- `split_delimited_v1` is the only transform that MAY use non-empty `transform_options` in the current profile. Its options object MUST contain only `delimiter`, `trim_items`, and `drop_empty_items`, and `delimiter` MUST be one of `,`, `;`, `|`, `\n`, or `\t`.
- For every current-profile transform other than `split_delimited_v1`, `transform_options` MUST be `{}`.

`POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select` and `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip` MUST each accept only a JSON object request body. Both routes MUST require `client_txn_id`. `POST /skip` MAY accept optional `reason`, bound to `string_contract_id=reason_note_v1`. Both routes are singleton action routes and MUST reject pagination members. Both routes MUST be route-scoped idempotent under §17.1. A no-op `select` against an already selected unit and a no-op `skip` against an already skipped unit MUST return the current durable state rather than fail.

`POST /api/v1/import-sessions/{import_session_id}/apply` MUST accept required `client_txn_id` and optional `selected_unit_ids[]`; omitted `selected_unit_ids[]` means use the session's persisted `selected_unit_ids[]`.

The import route family MUST preserve the durable session terminal states `applied`, `partially_applied`, `failed`, and `canceled`, and the durable unit terminal states `applied`, `skipped`, `rejected`, and `failed`; it MUST NOT serialize job-phase tokens as session or unit state.

For terminal common-job summaries produced by this family:

- `POST /api/v1/import-sessions` MUST use `result_summary.code='import_session_discovered'` and MUST emit exactly one `resource_refs[]` item `{ kind: 'import_session', id: <import_session_id>, route: '/api/v1/import-sessions/{import_session_id}' }`.
- `POST /api/v1/import-sessions/{import_session_id}/apply` MUST use `result_summary.code='import_session_applied'` when the durable `session_status='applied'` and `result_summary.code='import_session_partially_applied'` when the durable `session_status='partially_applied'`. In both success cases it MUST emit exactly one `import_session` ref using that same canonical route.
Profiles: import
Verified by: AC-263, AC-264, AC-324, AC-325

**REQ-01-475**
The import route family MUST use only `invalid_import_request`, `import_session_not_found`, `import_unit_not_found`, `import_state_conflict`, `import_source_unsupported`, `import_source_rejected`, and `import_apply_blocked`.

`invalid_import_request` MUST use only the shared upload-envelope reasons from REQ-01-553 plus:

- `request_not_object`,
- `missing_required_field`,
- `field_not_nullable`,
- `unknown_field`,
- `invalid_row_reference`,
- `invalid_selected_unit_ids`,
- `unsupported_assistant_profile`,
- `invalid_source_columns`,
- `invalid_unknown_column_policy`,
- `invalid_transform`,
- `invalid_empty_value_policy`,
- `duplicate_target_field`.

`import_state_conflict` MUST use only:

- `session_applying`,
- `session_terminal`,
- `unit_applying`,
- `unit_terminal`.

`import_source_unsupported` MUST use only `encrypted_or_unparseable_workbook`, `unsupported_named_range`, and `formula_cached_value_missing`.

`import_source_rejected` MUST use only `csv_source_too_large`, `xlsx_source_too_large`, `import_rows_exceeded`, `import_columns_exceeded`, `import_cells_exceeded`, `archive_extracted_bytes_exceeded`, `archive_compression_ratio_exceeded`, and `archive_member_count_exceeded`.

`import_apply_blocked` MUST use only `overlapping_units`, `duplicate_apply_blocked`, and `unit_not_ready`.
Profiles: import
Verified by: AC-265, AC-323, AC-324, AC-325

### 17.3 Snapshot and Reporting Extension Profile public contract

**REQ-01-476**
The Snapshot and Reporting Extension Profile MUST expose exactly this minimum public route surface under `/api/v1/snapshots/*` and `/api/v1/releases/*`:

- `POST /api/v1/snapshots`,
- `GET /api/v1/snapshots/{snapshot_id}`,
- `POST /api/v1/releases`,
- `GET /api/v1/releases/{release_id}`,
- `POST /api/v1/releases/{release_id}/approve`,
- `POST /api/v1/releases/{release_id}/publish`,
- `POST /api/v1/releases/{release_id}/invalidate`.
Profiles: snapshot_reporting
Verified by: AC-266, AC-267, AC-268

**Table 17.3-A. Snapshot and release route inventory**

| Route | Request contract summary | Success resource or body | Long-running | Primary family errors |
| --- | --- | --- | --- | --- |
| `POST /api/v1/snapshots` | JSON object with required `incident_id`, required `client_txn_id`, and optional `source_change_set_high_watermark` resolved once at job admission when omitted | Common job resource; terminal success emits one `snapshot` ref | Yes | `invalid_snapshot_request` |
| `GET /api/v1/snapshots/{snapshot_id}` | Singleton read | `snapshot` resource | No | `snapshot_not_found`, `invalid_pagination_request` |
| `POST /api/v1/releases` | JSON object with required snapshot, template, redaction-profile, output-kind, and `client_txn_id`; optional `release_scope` defaulting to `internal_draft` | Common job resource; terminal success emits one `release` ref | Yes | `invalid_release_request`, `release_render_failed` |
| `GET /api/v1/releases/{release_id}` | Singleton read | `release` resource | No | `release_not_found`, `invalid_pagination_request` |
| `POST /api/v1/releases/{release_id}/approve` | JSON object with required `client_txn_id` and optional `reason` | `200 OK` with `data.release` and `approval_progress` | No | `invalid_release_request`, `release_state_conflict`, `release_approval_rejected` |
| `POST /api/v1/releases/{release_id}/publish` | JSON object with required `client_txn_id` and optional `reason` | `200 OK` with `data.release` | No | `invalid_release_request`, `release_state_conflict` |
| `POST /api/v1/releases/{release_id}/invalidate` | JSON object with required `client_txn_id` and optional `reason` | `200 OK` with `data.release` | No | `invalid_release_request`, `release_state_conflict` |

**Table 17.3-B. Snapshot and release durable resources**

| Resource | Required members or properties |
| --- | --- |
| `snapshot` | `snapshot_id`, `incident_id`, `created_by_user_id`, `created_at`, `snapshot_at`, `source_change_set_high_watermark`, `derivation_version`, `export_model_sha256` |
| `release` | `release_id`, `incident_id`, `snapshot_id`, `snapshot_at`, `source_change_set_high_watermark`, `derivation_version`, `export_model_sha256`, `template_id`, `template_version`, `redaction_profile_id`, `redaction_profile_version`, `output_kind`, `release_scope`, `output_sha256`, `release_state`, `created_by_user_id`, `created_at`, `approved_at`, `invalidated_at`, `published_at`, `invalidation_reason` |
| `release_scope` omission rule | Omitted `release_scope` resolves to `internal_draft`; explicit `null` is invalid; omission and explicit `internal_draft` compare equal for idempotency |
| `release_state` vocabulary | Exactly `pending_approval`, `approved`, `invalidated`, `published`; worker-phase tokens are forbidden |

**Table 17.3-C. Release action summary**

| Action route | Legal current state | Success summary |
| --- | --- | --- |
| `approve` | `pending_approval` only | Records one approval for the exact immutable release tuple; success returns `approval_progress` with `approval_recorded`, `approval_requirements_satisfied`, and `resulting_release_state` |
| `publish` | `approved` only | Marks the release published and returns the post-commit `release` resource |
| `invalidate` | `pending_approval`, `approved`, or `published` | Marks the release invalidated and returns the post-commit `release` resource |

**Table 17.3-D. Snapshot and release terminal results and primary errors**

| Route or condition | Required code or registry |
| --- | --- |
| Snapshot-create success | `result_summary.code='snapshot_created'` and exactly one `snapshot` ref |
| Release-create success | `result_summary.code='release_created'` and exactly one `release` ref |
| Invalid release request registry | `invalid_release_request` for malformed create or action bodies |
| Release render failure registry | `release_render_failed` with `missing_redaction_rule`, `undeclared_template_binding`, and `missing_required_field` |
| Release state conflict registry | `release_state_conflict` with `not_approved`, `already_approved`, `already_published`, and `already_invalidated` |
| Release approval rejection registry | `release_approval_rejected` with `approval_requirements_not_met` |


**REQ-01-477**
`GET /api/v1/snapshots/{snapshot_id}` MUST return `data = <snapshot resource>`.

The `snapshot resource` MUST expose exactly:

- `snapshot_id`,
- `incident_id`,
- `created_by_user_id`,
- `created_at`,
- `snapshot_at`,
- `source_change_set_high_watermark`,
- `derivation_version`,
- `export_model_sha256`.

For `snapshot resource` serialization:

- every member above is required and non-null in the current profile;
- `source_change_set_high_watermark` MUST always serialize as the resolved committed boundary token, even when `POST /api/v1/snapshots` omitted it;
- the `snapshot resource` MUST NOT include `template_id`, `template_version`, `redaction_profile_id`, `redaction_profile_version`, `release_state`, approval data, redaction manifests, or rendered-output bytes.

`POST /api/v1/snapshots` MUST accept a JSON object with required `incident_id` and required `client_txn_id`. It MAY include optional `source_change_set_high_watermark`. For this member, omission means the current committed incident head resolved once at snapshot-job admission, explicit JSON `null` is invalid, and any supplied value MUST be one exact committed source-boundary token for the addressed incident. Omission and explicit transmission of that same resolved committed boundary MUST compare equal for idempotency and replay. Exact replay of a previously committed snapshot-create request MUST reuse the originally resolved committed boundary token rather than re-resolving a later incident head. `POST /api/v1/releases` MUST accept a JSON object with required `snapshot_id`, required `template_id`, required `template_version`, required `redaction_profile_id`, required `redaction_profile_version`, required `output_kind`, and required `client_txn_id`. It MAY include optional `release_scope`. For this member, omission means `internal_draft`, explicit JSON `null` is invalid, the allowed current-profile values are exactly `internal_draft`, `internal_review`, and `external_release`, and omission and explicit `internal_draft` MUST compare equal for idempotency and replay. The durable `release resource` MUST always serialize the resolved `release_scope`. Both routes MUST run as background jobs. The release-create route MUST fail closed if the request omits either version selector, attempts implicit latest-version resolution, or supplies a `release_scope` outside the closed current-profile vocabulary.

For terminal common-job summaries produced by this family:

- `POST /api/v1/snapshots` MUST use `result_summary.code='snapshot_created'` and MUST emit exactly one `resource_refs[]` item `{ kind: 'snapshot', id: <snapshot_id>, route: '/api/v1/snapshots/{snapshot_id}' }`.
- `POST /api/v1/releases` MUST use `result_summary.code='release_created'` and MUST emit exactly one `resource_refs[]` item `{ kind: 'release', id: <release_id>, route: '/api/v1/releases/{release_id}' }`.
Profiles: snapshot_reporting
Verified by: AC-266, AC-267

**REQ-01-478**
`GET /api/v1/releases/{release_id}` MUST return `data = <release resource>`.

The `release resource` MUST expose exactly:

- `release_id`,
- `incident_id`,
- `snapshot_id`,
- `snapshot_at`,
- `source_change_set_high_watermark`,
- `derivation_version`,
- `export_model_sha256`,
- `template_id`,
- `template_version`,
- `redaction_profile_id`,
- `redaction_profile_version`,
- `output_kind`,
- `release_scope`,
- `output_sha256`,
- `release_state`,
- `created_by_user_id`,
- `created_at`,
- `approved_at`,
- `invalidated_at`,
- `published_at`,
- `invalidation_reason`.

For `release resource` serialization:

- `approved_at`, `invalidated_at`, `published_at`, and `invalidation_reason` MUST always be present and MUST be JSON `null` when unset;
- every other member above is required and non-null in the current profile;
- `data.release` in successful `approve`, `publish`, and `invalidate` responses MUST use the exact `release resource` shape defined here;
- the `release resource` MUST NOT inline approval records, redaction manifests, rendered bytes, or worker-phase or job-status state.

`POST /api/v1/releases/{release_id}/approve`, `POST /api/v1/releases/{release_id}/publish`, and `POST /api/v1/releases/{release_id}/invalidate` MUST each accept only a JSON object with required `client_txn_id` and optional `reason`. If present, `reason` MUST be a JSON string or JSON `null` and MUST normalize under `string_contract_id=reason_note_v1`. For idempotency comparison, omission, explicit JSON `null`, and any `reason` value that normalizes to empty under `reason_note_v1` MUST compare equal. Unknown top-level members, a non-object body, missing `client_txn_id`, or `null` for a non-nullable member MUST fail with `400` and `error.code = invalid_release_request`. Route-scoped idempotency for these three action routes MUST be keyed by `(actor_user_id, release_id, action_route, client_txn_id)` and MUST compare the exact action route plus normalized `reason`. Exact replay of a previously committed success MUST return the original committed success result before any fresh state-conflict evaluation runs. Reuse of the same route-scoped key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. `approve` MUST mean recording an approval against that exact immutable release tuple. A successful `approve` MAY leave `release_state='pending_approval'` when the required approval set is not yet complete. It transitions to `approved` only when the required approval set is now satisfied for that exact release record. A successful `approve` MUST return `200 OK` with the common success envelope and `data` containing required `release` equal to the post-commit durable release resource plus required `approval_progress`, with exactly boolean `approval_recorded`, boolean `approval_requirements_satisfied`, and `resulting_release_state`, where `resulting_release_state` MUST use only `pending_approval` or `approved`. A successful `publish` MUST return `200 OK` with the common success envelope and `data.release` equal to the post-commit durable release resource. A successful `invalidate` MUST return `200 OK` with the common success envelope and `data.release` equal to the post-commit durable release resource. `approve` is legal only when the current `release_state` is `pending_approval`. `publish` is legal only when the current `release_state` is `approved`. `invalidate` is legal only when the current `release_state` is `pending_approval`, `approved`, or `published`. The public `release_state` vocabulary remains exactly `pending_approval`, `approved`, `invalidated`, and `published`. Render jobs MUST NOT introduce `queued`, `running`, `rendering`, or equivalent worker-phase tokens into `release_state`. A successful internal-draft render candidate becomes `approved` immediately because the current profile requires no separate approval action for `release_scope='internal_draft'`.
Profiles: snapshot_reporting
Verified by: AC-268, AC-305, AC-306

**REQ-01-479**
The snapshot and release route family MUST use only `invalid_snapshot_request`, `snapshot_not_found`, `invalid_release_request`, `release_not_found`, `release_state_conflict`, `release_approval_rejected`, and `release_render_failed`. `invalid_release_request` applies to malformed release-create requests and malformed approve, publish, or invalidate action bodies. `release_render_failed` MUST use only `missing_redaction_rule`, `undeclared_template_binding`, and `missing_required_field`. `release_state_conflict` MUST use only `not_approved`, `already_approved`, `already_published`, and `already_invalidated`. `release_approval_rejected` MUST use only `approval_requirements_not_met`, and that code is reserved for approval-eligibility failures while the durable release state still permits an approval attempt. Terminal-state failures MUST use `release_state_conflict`.
Profiles: snapshot_reporting
Verified by: AC-269, AC-307

### 17.4 Reference Pack Extension Profile public contract

**REQ-01-480**
The Reference Pack Extension Profile MUST expose exactly this minimum public route surface under `/api/v1/reference-packs/*`:

- `GET /api/v1/reference-packs`,
- `GET /api/v1/reference-packs/{pack_key}/{pack_version}`,
- `POST /api/v1/reference-packs/import`,
- `POST /api/v1/reference-packs/{pack_key}/{pack_version}/activate`,
- `POST /api/v1/reference-packs/{pack_key}/{pack_version}/disable`,
- `POST /api/v1/reference-packs/{pack_key}/{pack_version}/reverify`,
- `POST /api/v1/reference-packs/refresh`.
Profiles: reference_pack
Verified by: AC-270, AC-271

**Table 17.4-A. Reference-pack route inventory**

| Route | Request contract summary | Success resource or body | Long-running | Primary family errors |
| --- | --- | --- | --- | --- |
| `GET /api/v1/reference-packs` | List read under common paging | `{ pack_versions[] }` plus `meta.paging` | No | `reference_pack_not_found` only when a concrete pack version is addressed elsewhere |
| `GET /api/v1/reference-packs/{pack_key}/{pack_version}` | Singleton read | `reference_pack_version` resource | No | `reference_pack_not_found`, `invalid_pagination_request` |
| `POST /api/v1/reference-packs/import` | Shared upload envelope with required `client_txn_id`; optional `activation_policy` defaulting to `staged_only` and auto-activation forbidden | Common job resource; terminal success emits one `reference_pack_version` ref | Yes | `invalid_reference_pack_request`, `reference_pack_verification_failed` |
| `POST /api/v1/reference-packs/{pack_key}/{pack_version}/activate` | JSON object with required `client_txn_id` and optional `reason` | Either inline `200 OK` with `data.pack_version` or common job resource | Maybe | `reference_pack_activation_rejected`, `reference_pack_state_conflict` |
| `POST /api/v1/reference-packs/{pack_key}/{pack_version}/disable` | JSON object with required `client_txn_id` and optional `reason` | Either inline `200 OK` with `data.pack_version` or common job resource | Maybe | `reference_pack_state_conflict` |
| `POST /api/v1/reference-packs/{pack_key}/{pack_version}/reverify` | JSON object with required `client_txn_id` and optional `reason` | Common job resource | Yes | `reference_pack_state_conflict`, `reference_pack_verification_failed` |
| `POST /api/v1/reference-packs/refresh` | JSON object with required `client_txn_id` and optional `pack_keys[]`; omitted `pack_keys[]` resolves once at admission to all visible imported pack keys | Common job resource | Yes | `invalid_reference_pack_request`, `reference_pack_verification_failed` |

**Table 17.4-B. `reference_pack_version` resource summary**

| Member group | Requirement |
| --- | --- |
| Identity and type | `pack_key`, `pack_kind`, `pack_version` |
| Integrity and provenance | `manifest_sha256`, canonical `payload_sha256`, `verification_method`, `verification_result`, `source_identifier`, `signer_key_id` |
| Lifecycle state | `condition`, derived `active`, `previous_active_version` |
| Attribution | `imported_by_user_id`, `imported_at`, `activated_by_user_id`, `activated_at` |
| Current-profile durable condition vocabulary | Exactly `staged`, `verified_available`, `disabled`, `failed`, `missing`; `active` is derived from the activation pointer and is not an additional stored condition token |

**Table 17.4-C. Reference-pack terminal results and primary errors**

| Route or condition | Required code or registry |
| --- | --- |
| Import success | `result_summary.code='reference_pack_imported'` and exactly one `reference_pack_version` ref |
| Long-running activate success | `result_summary.code='reference_pack_activated'` and exactly one `reference_pack_version` ref |
| Long-running disable success | `result_summary.code='reference_pack_disabled'` and exactly one `reference_pack_version` ref |
| Reverify success | `result_summary.code='reference_pack_reverified'` and exactly one `reference_pack_version` ref |
| Refresh success | `result_summary.code='reference_packs_refreshed'`; `resource_refs[]` may be empty or non-exhaustive `reference_pack_version` refs sorted by `route asc` |
| Invalid request registry | `invalid_reference_pack_request` with shared upload-envelope reasons plus the request-shape and selector reasons in REQ-01-482 |
| Verification failure registry | `reference_pack_verification_failed` with checksum, signature, integrity-metadata, contract, path, content, payload, and archive-limit reasons |
| Activation rejection and state conflict registries | `reference_pack_activation_rejected` with `already_active` or `not_verified_available`; `reference_pack_state_conflict` with `already_disabled`, `not_disableable`, or `verification_pending` |


**REQ-01-481**
`GET /api/v1/reference-packs` MUST return the common success envelope with `data.pack_versions[]` plus `meta.paging` under §3.3.7. `GET /api/v1/reference-packs/{pack_key}/{pack_version}` MUST return `data = <reference_pack_version resource>`. Every item in `data.pack_versions[]` and every `data.pack_version` member returned by inline `200 OK` success from `activate` or `disable` MUST use the exact `reference_pack_version resource` shape defined here.

The `reference_pack_version resource` MUST expose exactly:

- `pack_key`,
- `pack_kind`,
- `pack_version`,
- `pack_version_state`,
- `active`,
- `source_identifier`,
- `manifest_sha256`,
- `payload_sha256`,
- `pack_contract_version`,
- `verification_method`,
- `verification_result`,
- `signer_key_id`,
- `previous_active_version`,
- `imported_by_user_id`,
- `imported_at`,
- `activated_by_user_id`,
- `activated_at`.

For `reference_pack_version resource` serialization:

- `pack_version_state` MUST use exactly `staged`, `verified_available`, `disabled`, `failed`, and `missing`;
- `verification_result` MUST use exactly `pending`, `passed`, and `failed`;
- `active` MUST always be present and MUST be the derived activation-pointer boolean for `(pack_key, pack_version)` rather than an additional durable version-state token;
- `pack_kind` MUST serialize as the exact stored metadata string and MUST NOT be narrowed to a closed v1 public enum;
- `pack_version` MUST serialize as the exact version identifier and MUST NOT imply `latest`, `current`, semantic-version ordering, or other route-local interpretation;
- `source_identifier`, `signer_key_id`, `previous_active_version`, `imported_by_user_id`, `activated_by_user_id`, and `activated_at` MUST always be present and MUST be JSON `null` when unset;
- every other member above is required and non-null in the current profile;
- `payload_sha256` is the canonical public digest field and MUST serialize as one canonical aggregate digest when storage retains more than one payload SHA-256 digest;
- object-member order is not part of the wire contract; array order is;
- the resource MUST NOT inline bundle bytes, extracted member lists, raw signatures, object-store paths, staging paths, or attestation-history arrays.

`data.pack_versions[]` MUST be exhaustive over pack versions visible to the caller for this route family and MUST sort by `pack_key asc`, then exact `pack_version asc`.

`POST /api/v1/reference-packs/import` MUST use the shared upload-envelope contract in §17.1.1. Within that contract, metadata MUST contain required `client_txn_id`. It MAY include optional `activation_policy`. For this route, the `file` part media type MUST be one of the exact values declared for `POST /api/v1/reference-packs/import` in REQ-01-552. Those media-type values are envelope gates only; bundle integrity, content screening, and archive validation remain byte-based and continue to use the route's existing verification rules. Route-scoped normalized request comparison for idempotency MUST include normalized `activation_policy` and SHA-256 of the exact uploaded file bytes. Multipart boundary text, part order, advisory filename, and non-semantic part headers or parameters MUST NOT affect normalized comparison. For this member, omission means `staged_only`, explicit JSON `null` is invalid, omission and explicit `staged_only` MUST compare equal for idempotency and replay, and the only accepted current-profile non-null token is `staged_only`. The current profile MUST reject any request that attempts auto-activation at import time. A non-null string token other than `staged_only` MUST fail with `400`, `error.code = invalid_reference_pack_request`, and `error.details.reason_code = auto_activation_not_supported`; any other malformed non-null form for `activation_policy` MUST fail with `reason_code = invalid_activation_policy`. `POST /api/v1/reference-packs/{pack_key}/{pack_version}/activate`, `disable`, and `reverify` MUST require an exact path `pack_version`; the current profile defines no implicit latest-version action route. Each of those action routes MUST accept only a JSON object with required `client_txn_id` and optional `reason`. If present, `reason` MUST be a JSON string or JSON `null` and MUST normalize under `string_contract_id=reason_note_v1`. For idempotency comparison, omission, explicit JSON `null`, and any `reason` value that normalizes to empty under `reason_note_v1` MUST compare equal. Unknown top-level members, a non-object body, missing `client_txn_id`, or `null` for a non-nullable member MUST fail with `400` and `error.code = invalid_reference_pack_request`. Route-scoped idempotency for these three action routes MUST be keyed by `(actor_user_id, pack_key, pack_version, action_route, client_txn_id)` and MUST compare the exact action route plus normalized `reason`. Exact replay of a previously committed success or accepted job MUST return the original committed result before any fresh state evaluation runs. Reuse of the same route-scoped key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. `activate` is legal only when the addressed version is in durable condition `verified_available` and is not currently active for its `pack_key`. `disable` is legal only when the addressed version is in durable condition `verified_available`; the action remains legal whether or not that version is currently active through the activation pointer. `reverify` is legal only when the addressed version is in durable condition `verified_available`, `disabled`, `failed`, or `missing`; it is not legal while still `staged`. `POST /api/v1/reference-packs/refresh` MUST accept required `client_txn_id` and optional `pack_keys[]`. For this member, omission means all currently imported `pack_key` values visible to the caller resolved once at refresh-job admission, explicit JSON `null` is invalid, explicit `[]` is invalid and MUST use `reason_code = empty_pack_keys`, and any supplied `pack_keys[]` value MUST be an array of exact visible `pack_key` strings. `pack_keys[]` is a set-like selector: caller order is non-semantic, duplicate members coalesce by exact token equality, and the canonical normalized form used for idempotency and replay is the unique exact-token set sorted by `pack_key asc`; omission MUST compare using the resolved admission-time set rather than later visibility state. If omitted `pack_keys[]` resolves to zero visible imported pack keys, refresh MUST still be admitted and MUST complete as a deterministic no-op background job rather than fail. Any non-string, unknown, or non-visible supplied `pack_key` MUST fail with `400`, `error.code = invalid_reference_pack_request`, and `reason_code = invalid_pack_keys`. Import, reverify, and refresh MUST enforce `limits.reference_packs.max_extracted_bytes`, `limits.archives.max_compression_ratio`, and `limits.archives.max_members` before a candidate version can remain or become `verified_available` or before refresh can keep or move the active pointer. A breach of any of those limits MUST fail closed using `reference_pack_verification_failed` and the exact corresponding archive-limit `reason_code`. Import, reverify, and refresh MUST run as background jobs. `activate` and `disable` MAY complete synchronously with `200 OK` using the common success envelope and `data.pack_version` equal to the post-commit durable `reference_pack_version resource`; if either action performs long-running work, it MUST return `202 Accepted` with the common job resource. `reverify` MUST always return `202 Accepted` with the common job resource. When a reverify job reaches a terminal state, its public result or error summary MUST use the same family-specific stable codes rather than ad hoc worker strings. The durable version conditions exposed to the public surface remain exactly `staged`, `verified_available`, `disabled`, `failed`, and `missing`. `active` MUST remain a derived boolean obtained from the activation pointer for `(pack_key, pack_version)`, not an additional stored version-state token.

For every `reference_pack_version` ref emitted by this family, `kind` MUST be `reference_pack_version` and both `id` and `route` MUST equal the canonical `/api/v1/reference-packs/{pack_key}/{pack_version}` path.

For terminal common-job summaries produced by this family:

- `POST /api/v1/reference-packs/import` MUST use `result_summary.code='reference_pack_imported'` and MUST emit exactly one `reference_pack_version` ref.
- A long-running `POST /api/v1/reference-packs/{pack_key}/{pack_version}/activate` MUST use `result_summary.code='reference_pack_activated'` and MUST emit exactly one `reference_pack_version` ref.
- A long-running `POST /api/v1/reference-packs/{pack_key}/{pack_version}/disable` MUST use `result_summary.code='reference_pack_disabled'` and MUST emit exactly one `reference_pack_version` ref.
- `POST /api/v1/reference-packs/{pack_key}/{pack_version}/reverify` MUST use `result_summary.code='reference_pack_reverified'` and MUST emit exactly one `reference_pack_version` ref.
- `POST /api/v1/reference-packs/refresh` MUST use `result_summary.code='reference_packs_refreshed'`. Its `resource_refs[]` MAY be empty or contain one or more `reference_pack_version` refs. When present, those refs are non-exhaustive when multiple versions changed and MUST sort by `route asc`.

These success-code rules apply only when the route completes through the common job resource. Inline `200 OK` `activate` or `disable` success continues to use `data.pack_version` equal to the exact `reference_pack_version resource` defined here.
Profiles: reference_pack
Verified by: AC-270, AC-271, AC-308, AC-309, AC-326, AC-369

**REQ-01-482**
The reference-pack route family MUST use only `invalid_reference_pack_request`, `reference_pack_not_found`, `reference_pack_state_conflict`, `reference_pack_verification_failed`, and `reference_pack_activation_rejected`. `invalid_reference_pack_request` MUST use only the shared upload-envelope reasons from REQ-01-553 plus `request_not_object`, `missing_required_field`, `field_not_nullable`, `unknown_field`, `invalid_activation_policy`, `pack_version_required`, `auto_activation_not_supported`, `invalid_pack_keys`, and `empty_pack_keys`. `reference_pack_verification_failed` MUST use only `checksum_mismatch`, `signature_mismatch`, `missing_integrity_metadata`, `contract_incompatible`, `path_traversal`, `disallowed_content`, `payload_missing`, `archive_extracted_bytes_exceeded`, `archive_compression_ratio_exceeded`, and `archive_member_count_exceeded`. `reference_pack_activation_rejected` MUST use only `already_active` and `not_verified_available`, and it is reserved for `activate`. `reference_pack_state_conflict` MUST use only `already_disabled`, `not_disableable`, and `verification_pending`.
Profiles: reference_pack
Verified by: AC-272, AC-310, AC-326

### 17.5 Incident Portability Extension Profile public contract

**REQ-01-483**
The Incident Portability Extension Profile MUST expose exactly this minimum public route surface under `/api/v1/incident-bundles/*`:

- `POST /api/v1/incident-bundles/export`,
- `GET /api/v1/incident-bundles/{bundle_id}`,
- `POST /api/v1/incident-bundles/import`.
Profiles: incident_portability
Verified by: AC-273, AC-274, AC-275

**Table 17.5-A. Incident-bundle route inventory**

| Route | Request contract summary | Success resource or body | Long-running | Primary family errors |
| --- | --- | --- | --- | --- |
| `POST /api/v1/incident-bundles/export` | JSON object with required `incident_id` and `client_txn_id`; optional `reference_pack_mode`, `optional_sections[]`, and `required_capabilities[]`; `history_mode` and `blob_mode` are forbidden user inputs | Common job resource; terminal success emits one `incident_bundle` ref | Yes | `invalid_incident_bundle_request`, `incident_bundle_export_rejected` |
| `GET /api/v1/incident-bundles/{bundle_id}` | Singleton read | Durable export descriptor | No | `incident_bundle_not_found`, `invalid_pagination_request` |
| `POST /api/v1/incident-bundles/import` | Shared upload envelope with required `client_txn_id` in metadata | Common job resource; terminal success emits one imported `incident` ref | Yes | `invalid_incident_bundle_request`, `incident_bundle_import_rejected` |

**Table 17.5-B. Export descriptor and import metadata summary**

| Contract element | Requirement |
| --- | --- |
| `reference_pack_mode` omission rule | Omitted means `refs_only`; explicit `null` is invalid; omission and explicit `refs_only` compare equal |
| `optional_sections[]` omission rule | Omitted means `[]`; explicit `null` is invalid; explicit `[]` compares equal to omission; allowed tokens are exactly `snapshots` and `reference_packs`; order is non-semantic and canonicalized ascending |
| `required_capabilities[]` omission rule | Omitted means `[]`; explicit `null` is invalid; explicit `[]` compares equal to omission; allowed tokens are exactly `snapshots` and `reference_packs`; order is non-semantic and canonicalized ascending |
| Durable export descriptor | `bundle_id`, `incident_id`, `exported_at`, `manifest_sha256`, `reference_pack_mode`, `optional_sections[]`, `required_capabilities[]`, fixed `history_mode='full'`, fixed `blob_mode='full'` |
| Import boundary | No durable import resource exists in the current profile; import is create-only into an empty incident namespace |

**Table 17.5-C. Incident-bundle terminal results and primary errors**

| Route or condition | Required code or registry |
| --- | --- |
| Export success | `result_summary.code='incident_bundle_exported'` and exactly one `incident_bundle` ref |
| Import success | `result_summary.code='incident_bundle_imported'` and exactly one imported `incident` ref |
| Invalid request registry | `invalid_incident_bundle_request` with shared upload-envelope reasons plus the export and import request-shape reasons in REQ-01-486 |
| Export rejection registry | `incident_bundle_export_rejected` with `missing_required_file` and `missing_required_blob` |
| Import rejection registry | `incident_bundle_import_rejected` with member-path, member-type, integrity, blob-hash, duplicate-incident, capability, remote-fetch, and archive-limit reasons |


**REQ-01-484**
`POST /api/v1/incident-bundles/export` MUST accept a JSON object with required `incident_id` and required `client_txn_id`. It MAY include optional `reference_pack_mode`, optional `optional_sections[]`, and optional `required_capabilities[]`. For `reference_pack_mode`, omission means `refs_only`, explicit JSON `null` is invalid, omission and explicit `refs_only` MUST compare equal for idempotency and replay, and the allowed current-profile values are exactly `refs_only` and `embedded`. For `optional_sections[]`, omission means `[]`, explicit JSON `null` is invalid, explicit `[]` compares equal to omission, the allowed current-profile tokens are exactly `snapshots` and `reference_packs`, caller order is non-semantic, duplicate members coalesce by exact token equality, and the canonical normalized form is the unique exact-token set sorted ascending. For `required_capabilities[]`, omission means `[]`, explicit JSON `null` is invalid, explicit `[]` compares equal to omission, the allowed current-profile tokens are exactly `snapshots` and `reference_packs`, caller order is non-semantic, duplicate members coalesce by exact token equality, and the canonical normalized form is the unique exact-token set sorted ascending. Unknown tokens in either array or any `reference_pack_mode` value outside the closed current-profile vocabulary are invalid and MUST NOT be silently ignored or dropped at export admission. The current profile MUST NOT expose user-tunable partial-history or partial-blob request modes; if `history_mode` or `blob_mode` is supplied, the route MUST fail closed. Export MUST run as a background job. The durable export descriptor under `GET /api/v1/incident-bundles/{bundle_id}` MUST exist only after successful export, MUST reject pagination, and MUST expose at minimum `bundle_id`, `incident_id`, `exported_at`, `manifest_sha256`, `reference_pack_mode`, `optional_sections[]`, `required_capabilities[]`, fixed `history_mode='full'`, and fixed `blob_mode='full'`. The durable export descriptor and the emitted `manifest.json` MUST both serialize the resolved `reference_pack_mode` and the canonicalized `optional_sections[]` and `required_capabilities[]` values rather than caller order. On successful export, the terminal common-job summary MUST use `result_summary.code='incident_bundle_exported'` and MUST emit exactly one `resource_refs[]` item `{ kind: 'incident_bundle', id: <bundle_id>, route: '/api/v1/incident-bundles/{bundle_id}' }`.
Profiles: incident_portability
Verified by: AC-273, AC-274

**REQ-01-485**
`POST /api/v1/incident-bundles/import` MUST use the shared upload-envelope contract in §17.1.1. Within that contract, metadata MUST contain required `client_txn_id`. For this route, the `file` part media type MUST be one of the exact values declared for `POST /api/v1/incident-bundles/import` in REQ-01-552. Those media-type values are envelope gates only; bundle-member validation, integrity verification, and archive-limit enforcement remain byte-based. Route-scoped normalized request comparison for idempotency MUST include SHA-256 of the exact uploaded file bytes. Multipart boundary text, part order, advisory filename, and non-semantic part headers or parameters MUST NOT affect normalized comparison. Import MUST run as a background job. The current profile defines no durable import resource; on success the terminal job result summary MUST use `result_summary.code='incident_bundle_imported'` and MUST emit exactly one `resource_refs[]` item `{ kind: 'incident', id: <incident_id>, route: '/api/v1/incidents/{incident_id}' }`. Import remains create-only into an empty incident namespace. The current profile MUST reject clone, merge, identifier-remap, remote-fetch, or equivalent alternative import modes.
Profiles: incident_portability
Verified by: AC-275

**REQ-01-486**
The incident-bundle route family MUST use only `invalid_incident_bundle_request`, `incident_bundle_not_found`, `incident_bundle_export_rejected`, and `incident_bundle_import_rejected`. `invalid_incident_bundle_request` MUST use only the shared upload-envelope reasons from REQ-01-553 plus `request_not_object`, `missing_required_field`, `field_not_nullable`, `unknown_field`, `invalid_reference_pack_mode`, `invalid_optional_sections`, `invalid_required_capabilities`, `history_mode_not_supported`, and `blob_mode_not_supported`. `incident_bundle_export_rejected` MUST use only `missing_required_file` and `missing_required_blob`. `incident_bundle_import_rejected` MUST use only `invalid_member_path`, `unsupported_member_type`, `checksum_mismatch`, `signature_mismatch`, `blob_hash_mismatch`, `duplicate_incident_id`, `unsupported_required_capability`, `remote_fetch_required`, `archive_extracted_bytes_exceeded`, `archive_compression_ratio_exceeded`, and `archive_member_count_exceeded`.
Profiles: incident_portability
Verified by: AC-276, AC-327, AC-328, AC-332

## 18. Writable-string contract registry

**REQ-01-487**
This section and §18A and §18B are the primary owners for the base-profile writable string, direct temporal scalar, and direct-reference scalar surface contracts used by route-scoped request members and by view-schema field or action-member bindings. A writable human-authored string surface closed by this core MUST bind to one `string_contract_id` from the closed registry in this section. A writable direct temporal scalar surface closed by this core MUST bind to one `direct_scalar_contract_id` from the closed registry in §18A rather than inventing route-local lexical, normalization, or clearability prose. A writable direct-reference scalar surface closed by this core MUST bind to one `direct_reference_contract_id` from the closed registry in §18B rather than inventing route-local lexical, normalization, clearability, or identifier-resolution prose.
Profiles: base
Verified by: AC-015, AC-068, AC-085, AC-086, AC-112, AC-118, AC-152, AC-175, AC-176, AC-181, AC-182, AC-186, AC-194, AC-196, AC-200, AC-202, AC-216, AC-221, AC-225, AC-231, AC-300, AC-301, AC-302, AC-303, AC-315, AC-316, AC-317, AC-318, AC-319

**REQ-01-488**
Unless a bound contract below explicitly says otherwise:

- normalization and equality MUST be evaluated after JSON decoding,
- create-time idempotency and structurally valid no-op detection MUST compare bound writable-surface values after contract normalization,
- on patch, omission means unchanged,
- for string contracts whose binding allows clear-to-null semantics, explicit JSON `null` and any supplied string value that normalizes to empty MUST compare equal and MUST persist as authoritative `null`,
- for string contracts whose binding is required, explicit JSON `null` and any supplied string value that normalizes to empty MUST be rejected,
- for direct temporal scalar contracts whose binding allows clear-to-null semantics, authoritative clear MUST be explicit JSON `null`,
- for direct temporal scalar contracts whose binding is non-clearable, explicit JSON `null` MUST be rejected,
- the base profile MUST NOT preserve the empty string as a distinct authoritative value for any field bound to a clear-to-null string contract,
- the shared query-time text-comparison substrate used by text sorts and by case-insensitive filter semantics MUST first apply the bound field contract's authoritative normalization and MUST then apply locale-independent Unicode case folding,
- diacritics remain significant under that substrate,
- no compatibility folding, transliteration, punctuation stripping, tokenization, or extra whitespace collapse occurs beyond the bound field contract when that substrate is used.
Profiles: base
Verified by: AC-015, AC-068, AC-085, AC-086, AC-112, AC-118, AC-152, AC-175, AC-176, AC-181, AC-182, AC-184, AC-185, AC-186, AC-194, AC-196, AC-200, AC-202, AC-216, AC-221, AC-225, AC-231, AC-300, AC-301, AC-302, AC-303, AC-315, AC-316, AC-317, AC-318, AC-319

**REQ-01-489**
`display_name_line_v1` is the required single-line display-name contract.

- apply Unicode NFC normalization,
- trim leading and trailing Unicode whitespace,
- preserve interior whitespace exactly,
- reject every C0 or C1 control code point,
- do not case-fold for authoritative storage,
- enforce a maximum length of 256 Unicode scalar values after normalization,
- when this contract is bound to a required field, `null` or normalized-empty input is invalid.
Profiles: base
Verified by: AC-118, AC-152, AC-175, AC-176, AC-231

**REQ-01-490**
`single_line_title_v1` is the bounded single-line title or summary contract.

- apply the same NFC, trim, interior-whitespace-preservation, and C0/C1-rejection rules as `display_name_line_v1`,
- enforce a maximum length of 512 Unicode scalar values after normalization,
- when the binding is optional, normalized-empty input MUST clear to authoritative `null`,
- when the binding is required, `null` or normalized-empty input is invalid.
Profiles: base
Verified by: AC-015, AC-068, AC-085, AC-086, AC-112, AC-118, AC-231

**REQ-01-491**
`multiline_body_v1` is the bounded multiline body contract.

- apply Unicode NFC normalization,
- normalize `CRLF` and bare `CR` to `LF`,
- trim leading and trailing Unicode whitespace for authoritative storage and equality,
- preserve interior whitespace and line breaks,
- allow `LF` and `TAB` but reject every other C0 or C1 control code point,
- enforce a maximum length of 16384 Unicode scalar values after normalization,
- when the binding is optional, normalized-empty input MUST clear to authoritative `null`,
- when the binding is required, `null` or normalized-empty input is invalid.
Profiles: base
Verified by: AC-068, AC-085, AC-086, AC-112, AC-118, AC-216, AC-231

**REQ-01-492**
`party_text_v1` is the optional bounded source-preserving party-label contract.

- apply the same NFC, trim, interior-whitespace-preservation, and C0/C1-rejection rules as `display_name_line_v1`,
- enforce a maximum length of 256 Unicode scalar values after normalization,
- normalized-empty input MUST clear to authoritative `null`.
Profiles: base
Verified by: AC-015, AC-085, AC-100, AC-118, AC-231

**REQ-01-493**
`locator_text_v1` is the optional bounded locator or external-reference text contract.

- apply the same NFC, trim, interior-whitespace-preservation, and C0/C1-rejection rules as `display_name_line_v1`,
- enforce a maximum length of 1024 Unicode scalar values after normalization,
- normalized-empty input MUST clear to authoritative `null`.
Profiles: base
Verified by: AC-015, AC-085, AC-100, AC-118, AC-231

**REQ-01-494**
`tag_label_v1` is the required incident-scoped tag-label contract.

- apply the same NFC, trim, and control-character rules as `display_name_line_v1`,
- enforce a maximum length of 64 Unicode scalar values after normalization,
- authoritative storage MUST remain the normalized label without case-folding,
- dedupe and uniqueness comparison MUST use the trimmed Unicode-NFC form with case-insensitive comparison,
- `null` or normalized-empty input is invalid.
Profiles: base
Verified by: AC-118, AC-200, AC-231

**REQ-01-495**
`alias_text_v1` is the required entity-alias text contract.

- apply the same NFC, trim, and control-character rules as `display_name_line_v1`,
- enforce a maximum length of 256 Unicode scalar values after normalization,
- authoritative storage MUST remain the normalized alias text without case-folding,
- dedupe comparison for one canonical record MUST use the trimmed Unicode-NFC form with case-insensitive comparison,
- `null` or normalized-empty input is invalid.
Profiles: base
Verified by: AC-118, AC-202, AC-231

**REQ-01-496**
`reason_note_v1` is the bounded multiline audit-note contract for optional or required action reasons and other reason-style note fields.

- apply the same NFC, line-ending normalization, trim, and control-character rules as `multiline_body_v1`,
- enforce a maximum length of 4096 Unicode scalar values after normalization,
- when the binding is optional, omission, explicit JSON `null`, and normalized-empty input MUST compare equal and MUST persist as authoritative `null`,
- when the binding is required, `null` or normalized-empty input is invalid.
Profiles: base
Verified by: AC-085, AC-118, AC-181, AC-182, AC-186, AC-194, AC-196, AC-216, AC-221, AC-225, AC-231, AC-305, AC-308

**REQ-01-497**
`email_address_v1` is the optional bounded email-address contract.

- apply Unicode NFC normalization,
- trim leading and trailing Unicode whitespace,
- reject every C0 or C1 control code point,
- reject interior Unicode whitespace,
- require exactly one `@` and non-empty local and domain parts after normalization,
- authoritative storage MAY preserve original letter case,
- deterministic comparison and exact-match reuse MUST use the trimmed Unicode-NFC form with case-insensitive comparison,
- enforce a maximum length of 320 Unicode scalar values after normalization,
- normalized-empty input MUST clear to authoritative `null`.

In the base profile, `email_address_v1` is the authoritative normalization and comparison substrate for local-login `username`, user-account `email`, and membership-by-email resolution.
Profiles: base
Verified by: AC-175, AC-176, AC-178, AC-231, AC-247, AC-277, AC-279, AC-311, AC-312

**REQ-01-498**
`timezone_name_v1` is the optional bounded IANA-timezone-name contract.

- apply Unicode NFC normalization,
- trim leading and trailing Unicode whitespace,
- reject every C0 or C1 control code point,
- preserve interior characters exactly,
- enforce a maximum length of 128 Unicode scalar values after normalization,
- any non-null value MUST validate as one canonical IANA timezone name known to the runtime timezone database,
- normalized-empty input MUST clear to authoritative `null`.
Profiles: base
Verified by: AC-277, AC-231

**REQ-01-521**
`local_password_provision_v1` is the required exact-string secret-input contract for local-password provisioning surfaces.

- evaluate input after JSON decoding,
- accepted non-null input MUST be a JSON string,
- MUST NOT apply Unicode NFC normalization,
- MUST NOT trim leading or trailing Unicode whitespace,
- MUST NOT case-fold,
- MUST NOT normalize line endings,
- leading and trailing whitespace are significant,
- reject input composed entirely of Unicode whitespace code points,
- reject every C0 or C1 control code point,
- enforce a minimum length of 12 Unicode scalar values after JSON decoding,
- enforce a maximum length of 1024 Unicode scalar values after JSON decoding,
- idempotency equality and structurally valid no-op comparison MUST use exact post-decoding code-point equality,
- this contract defines only transport, validation, canonical comparison, and rejection boundaries; it does not define password-composition policy beyond those boundaries,
- when the binding is required, `null`, non-string, all-whitespace, control-bearing, shorter-than-minimum, or longer-than-maximum input is invalid.
Profiles: base
Verified by: AC-175, AC-176, AC-231, AC-244, AC-245

**REQ-01-568**
`mention_token_text_v1` is the required mention-token text contract for typed Timeline host and identity tokens carried by `add_token.raw_text` and `add_resolved_ref.raw_text`.

- apply Unicode NFC normalization,
- trim leading and trailing Unicode whitespace,
- collapse each maximal run of Unicode whitespace to one ASCII space,
- reject every C0 or C1 control code point,
- preserve every other code point exactly,
- enforce a maximum length of 256 Unicode scalar values after normalization,
- `null` or normalized-empty input is invalid.

This contract owns only writable-surface normalization and validation. Suppressor grammar, forbidden rewrites, and auto-resolution eligibility remain owned by Core 03 §12.
Profiles: base
Verified by: AC-118, AC-231, AC-388, AC-389, AC-390, AC-391

## 18A. Direct-scalar timestamp contract registry

The base profile currently closes exactly one writable direct-temporal-scalar contract: `timestamp_instant_v1`.

A field entry or create-time member binding that declares `direct_scalar_contract_id=timestamp_instant_v1` uses this contract:

- accepted non-null input is a JSON string in RFC 3339 timestamp form with an explicit timezone designator, either `Z` or a numeric offset,
- timezone-less strings, date-only strings, empty strings, numbers, booleans, arrays, and objects are invalid,
- canonical normalization and equality compare one represented instant in UTC RFC 3339 form with `Z`, so offset-equivalent values compare equal after normalization,
- clear-to-null is only explicit JSON `null`, and only when the bound field entry declares `clearable=true`,
- the public API MUST NOT treat the empty string as an authoritative clear for this contract,
- this contract governs only the canonical comparable instant; preserved source text, original offset, or precision caveats remain separate source-preserving text or metadata rather than part of the scalar contract.

The base profile defines no other writable direct-temporal-scalar contract.

## 18B. Writable direct-reference-scalar contract registry

**REQ-01-516**
Any writable direct-reference scalar surface closed by this core MUST bind to one `direct_reference_contract_id` from the closed registry in this section. These surfaces remain direct-write fields that use `value` under §3.3.5. A writable direct-reference scalar surface MUST NOT be reclassified as an `action_payload` field unless a later version defines a distinct write-action contract for that surface.
Profiles: base
Verified by: AC-315, AC-317, AC-319

**REQ-01-517**
Unless a bound contract below explicitly says otherwise:

- on patch, omission means unchanged,
- on patch, authoritative clear for a clearable direct-reference scalar MUST be explicit JSON `null`,
- on create, omission and explicit JSON `null` compare equal only when the bound field is optional and the resulting authoritative state is `null`,
- when the binding is non-clearable, explicit JSON `null` MUST be rejected,
- the empty string is never a clear token and MUST be rejected,
- numbers, booleans, arrays, and objects MUST be rejected,
- non-null values MUST be exact stable identifiers only.
Profiles: base
Verified by: AC-315, AC-316, AC-317, AC-319

**REQ-01-518**
For any direct-reference contract in this section, the contract layer MUST NOT trim, case-fold, label-resolve, email-resolve, fuzzy-match, or auto-create a target from a non-null submitted value. Any target-type, same-incident, active-row, or authorization checks remain owned by the bound field's view-schema entry and the authoritative domain-model rules.
Profiles: base
Verified by: AC-317, AC-319

**REQ-01-519**
`same_incident_party_ref_v1` is the exact same-incident party-reference contract.

- accepted non-null input is a JSON string whose exact value is one stable `party_id`,
- `null` clears only when the bound field entry declares `clearable=true`,
- a submitted value that is not one exact `party_id` string is invalid,
- a submitted `party_id` MUST identify an active `party` row in the same incident as the referencing record.
Profiles: base
Verified by: AC-315, AC-316, AC-317, AC-318

**REQ-01-520**
`same_incident_decision_ref_v1` is the exact same-incident decision-reference contract.

- accepted non-null input is a JSON string whose exact value is one stable `record_id`,
- the addressed target MUST be an active same-incident `decision` record,
- `null` clears only when the bound field entry declares `clearable=true`,
- if the bound field is realized as a denormalized convenience scalar over authoritative `record_links`, set and clear operations MUST update the convenience field and the authoritative distinguished decision association atomically.
Profiles: base
Verified by: AC-315, AC-316, AC-317, AC-319

## 19. Parties system-view addendum

**REQ-01-499**
- surface: contract-backed `Parties` system view
- source record types: `party`
- base projection: `party_grid_projection`
- this surface is incident-scoped coordination identity, not deployment-local user or account administration
- `default_visible_fields`: `party.display_name`, `party.party_kind`, `party.organization_name`, `party.role_title`, `party.primary_email`, `party.timezone_name`, `party.external_ref`, `party.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `party.notes`
- `default_sort`: `party.display_name asc`, `record_id asc`
- `sort_fields`: `party.display_name`, `party.party_kind`, `party.organization_name`, `party.role_title`, `party.primary_email`, `party.timezone_name`, `party.external_ref`, `party.updated_at`
- `filter_fields`: `party.display_name`, `party.party_kind`, `party.organization_name`, `party.primary_email`, `party.external_ref`, `party.updated_at`
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-231, AC-277

**REQ-01-500**
- inline create: zero-field create is forbidden
- minimum create set: inline create from the sheet itself MUST commit only when `party.display_name` is non-empty and `party.party_kind` is present after create-time normalization
- server-managed timestamps, attribution, and `row_version` MUST NOT satisfy the minimum create set
- ordinary grid edits to `party.display_name`, `party.party_kind`, `party.organization_name`, `party.role_title`, `party.primary_email`, `party.timezone_name`, `party.external_ref`, and `party.notes` are permitted subject to their field contracts
Profiles: base
Verified by: AC-117, AC-118, AC-231, AC-277

**REQ-01-501**
- writable fields:
  - `party.display_name`: read `display_name`; write target the `display_name` field on the underlying `party` record; `string_contract_id=display_name_line_v1`; `conflict_resolution_class=text_compare_merge`
  - `party.party_kind`: read `party_kind`; write target the `party_kind` field on the underlying `party` record; `conflict_resolution_class=atomic_replace`
  - `party.organization_name`: read `organization_name`; write target the `organization_name` field on the underlying `party` record; `string_contract_id=display_name_line_v1`; `conflict_resolution_class=text_compare_merge`
  - `party.role_title`: read `role_title`; write target the `role_title` field on the underlying `party` record; `string_contract_id=display_name_line_v1`; `conflict_resolution_class=text_compare_merge`
  - `party.primary_email`: read `primary_email`; write target the `primary_email` field on the underlying `party` record; `string_contract_id=email_address_v1`; `conflict_resolution_class=atomic_replace`
  - `party.timezone_name`: read `timezone_name`; write target the `timezone_name` field on the underlying `party` record; `string_contract_id=timezone_name_v1`; `conflict_resolution_class=atomic_replace`
  - `party.external_ref`: read `external_ref`; write target the `external_ref` field on the underlying `party` record; `string_contract_id=locator_text_v1`; `conflict_resolution_class=atomic_replace`
  - `party.notes`: read `notes`; write target the `notes` field on the underlying `party` record; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
- read-only computed fields: `party.updated_at`
Profiles: base
Verified by: AC-118, AC-231, AC-277

**REQ-01-502**
Hidden writable party-link fields on other base-profile surfaces MUST remain supplemental to the visible source-preserving text fields. In particular, `task.requester_party_id`, `evidence.collector_party_id`, and `evidence.source_party_id` MAY be written through inspector or same-surface enrichment flows, but they MUST NOT replace the visible text fields or require those fields to be shown as mandatory grid columns.
Profiles: base
Verified by: AC-118, AC-231, AC-278, AC-279, AC-318

### Additional coordination and optional artifact-backed surface addenda

**REQ-01-503**
- surface: required workbook-native coordination surface with canonical public identity `cartulary.view.comm_log.v1`; any saved view over this same `view_schema_id` is a distinct saved-view object rather than the required base surface
- source record types: `artifact` filtered to `artifact_type='comm_log'`
- base projection: `artifact_grid_projection` filtered to `artifact_type='comm_log'`
- `default_visible_fields`: `comm_log.timestamp_utc`, `comm_log.comm_type`, `comm_log.audience`, `comm_log.channel_or_meeting`, `comm_log.summary`, `comm_log.next_report_at`, `comm_log.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `comm_log.comm_id`, `comm_log.privilege_tag`, `comm_log.decision_ids`, `comm_log.action_task_ids`, `comm_log.audience_party_ids`, `comm_log.attendee_party_ids`, `comm_log.timestamp_day`, `comm_log.next_report_day`
- `default_sort`: `comm_log.timestamp_utc desc`, `record_id asc`
- `sort_fields`: `comm_log.timestamp_utc`, `comm_log.comm_type`, `comm_log.audience`, `comm_log.channel_or_meeting`, `comm_log.summary`, `comm_log.next_report_at`, `comm_log.updated_at`, `comm_log.privilege_tag`, `comm_log.timestamp_day`, `comm_log.next_report_day`
- `filter_fields`: `comm_log.comm_type`, `comm_log.timestamp_day`, `comm_log.next_report_day`, `comm_log.audience`, `comm_log.channel_or_meeting`, `comm_log.privilege_tag`
- `grouping_fields`: `comm_log.comm_type`, `comm_log.timestamp_day`, `comm_log.next_report_day`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `comm_log.comm_type` is present and `comm_log.audience`, `comm_log.channel_or_meeting`, and `comm_log.summary` are non-empty after create-time normalization
- when omitted on create, the server MUST generate `comm_log.comm_id`, default `comm_log.timestamp_utc` to the commit timestamp, default `comm_log.decision_ids`, `comm_log.action_task_ids`, `comm_log.audience_party_ids`, and `comm_log.attendee_party_ids` to empty collections, and default `comm_log.next_report_at` plus `comm_log.privilege_tag` to `null`
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `comm_log.timestamp_utc`: read `timestamp_utc`; write target the `timestamp_utc` field on the underlying `comm_log` artifact subtype; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=false`; `conflict_resolution_class=atomic_replace`
  - `comm_log.comm_type`: read `comm_type`; write target the `comm_type` field on the underlying `comm_log` artifact subtype; `conflict_resolution_class=atomic_replace`
  - `comm_log.audience`: read `audience`; write target the `audience` field on the underlying `comm_log` artifact subtype; `string_contract_id=party_text_v1`; `conflict_resolution_class=text_compare_merge`
  - `comm_log.channel_or_meeting`: read `channel_or_meeting`; write target the `channel_or_meeting` field on the underlying `comm_log` artifact subtype; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `comm_log.summary`: read `summary`; write target the `summary` field on the underlying `comm_log` artifact subtype; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `comm_log.next_report_at`: read `next_report_at`; write target the `next_report_at` field on the underlying `comm_log` artifact subtype; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
  - `comm_log.privilege_tag`: read `privilege_tag`; write target the `privilege_tag` field on the underlying `comm_log` artifact subtype; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=atomic_replace`
  - `comm_log.decision_ids`: read linked decision references; write action upsert or remove same-incident `decision` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
  - `comm_log.action_task_ids`: read linked task-request references; write action upsert or remove same-incident `task_request` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
  - `comm_log.audience_party_ids`: read supplemental audience party references; write action upsert or remove same-incident `party` references using the `party_ref` family defined below; `conflict_resolution_class=collection_review`
  - `comm_log.attendee_party_ids`: read supplemental attendee party references; write action upsert or remove same-incident `party` references using the `party_ref` family defined below; `conflict_resolution_class=collection_review`
- collection-review wire contract:
  - `comm_log.decision_ids` and `comm_log.action_task_ids` MUST use the exact `collection_value_v1` item shape and `collection_actions_v1` action vocabulary defined for `assessment.support_refs` in `REQ-01-333` and `REQ-01-334`, except the active `field_key` is `comm_log.decision_ids` or `comm_log.action_task_ids` and the server derives `references_record` routing from that field key under `REQ-01-311`.
  - `comm_log.decision_ids` accepts only same-incident active `decision` targets. `comm_log.action_task_ids` accepts only same-incident active `task_request` targets.
  - For those two `record_ref` fields, duplicate adds for the same patched `record_id`, `linked_record_id`, and active `field_key` MUST coalesce to one surviving logical reference binding. Removal MUST target `item_ref` only. A target from another incident, a target of the wrong `record_type`, a soft-deleted target, or an invalid or foreign `item_ref` MUST fail with `400 Bad Request` and `error.code = invalid_mutation_payload`.
  - `comm_log.audience_party_ids` and `comm_log.attendee_party_ids` MUST use `collection_value_v1` with `ordered=false`.
  - Each `items[]` entry MUST use this shape:

```json
{
  "item_ref": "party_ref:<party_id>",
  "item_kind": "party_ref",
  "display_text": "<party display_name>",
  "party_id": "<party_id>"
}
```

  - Allowed actions are:

```json
{ "op": "add_party_ref", "party_id": "<party_id>" }
```

```json
{ "op": "remove_party_ref", "item_ref": "party_ref:<party_id>" }
```

  - `party_id` in `add_party_ref` MUST identify a same-incident active `party` record.
  - Duplicate adds for the same patched `record_id`, `party_id`, and active `field_key` MUST coalesce to one surviving logical reference binding. Removal MUST target `item_ref` only. A target from another incident, a target of the wrong `record_type`, a soft-deleted target, or an invalid or foreign `item_ref` MUST fail with `400 Bad Request` and `error.code = invalid_mutation_payload`.
- read-only computed or system-managed fields: `comm_log.comm_id`, `comm_log.timestamp_day`, `comm_log.next_report_day`, `comm_log.updated_at`
- `comm_log.audience` remains required source-preserving text even when supplemental party references are present
- the base profile defines no separate `comm_log.attendee_text` field; attendee semantics are represented by supplemental `comm_log.attendee_party_ids` references and MUST NOT replace or weaken required `comm_log.audience` text
- removing a supplemental party reference MUST NOT clear or rewrite `comm_log.audience`
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-231, AC-281, AC-300, AC-301, AC-302, AC-303

**REQ-01-504**
- surface: required workbook-native coordination surface with canonical public identity `cartulary.view.handoff.v1`; any saved view over this same `view_schema_id` is a distinct saved-view object rather than the required base surface
- source record types: `artifact` filtered to `artifact_type='handoff'`
- base projection: `artifact_grid_projection` filtered to `artifact_type='handoff'`
- `default_visible_fields`: `handoff.timestamp_utc`, `handoff.outgoing_owner_user_id`, `handoff.incoming_owner_user_id`, `handoff.current_state_summary`, `handoff.next_checks`, `handoff.acknowledged_at`, `handoff.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `handoff.handoff_id`, `handoff.open_task_ids`, `handoff.open_decision_ids`, `handoff.open_risk_refs`, `handoff.timestamp_day`, `handoff.ack_state`
- `default_sort`: `handoff.timestamp_utc desc`, `record_id asc`
- `sort_fields`: `handoff.timestamp_utc`, `handoff.outgoing_owner_user_id`, `handoff.incoming_owner_user_id`, `handoff.current_state_summary`, `handoff.next_checks`, `handoff.acknowledged_at`, `handoff.updated_at`, `handoff.timestamp_day`, `handoff.ack_state`
- `filter_fields`: `handoff.timestamp_day`, `handoff.outgoing_owner_user_id`, `handoff.incoming_owner_user_id`, `handoff.ack_state`
- `grouping_fields`: `handoff.timestamp_day`, `handoff.outgoing_owner_user_id`, `handoff.incoming_owner_user_id`, `handoff.ack_state`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `handoff.incoming_owner_user_id` is present and `handoff.current_state_summary` is non-empty after create-time normalization
- when omitted on create, the server MUST generate `handoff.handoff_id`, default `handoff.timestamp_utc` to the commit timestamp, default `handoff.outgoing_owner_user_id` to the authenticated actor, default `handoff.open_task_ids`, `handoff.open_decision_ids`, and `handoff.open_risk_refs` to empty collections, and default `handoff.next_checks` plus `handoff.acknowledged_at` to `null`
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `handoff.timestamp_utc`: read `timestamp_utc`; write target the `timestamp_utc` field on the underlying `handoff` artifact subtype; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=false`; `conflict_resolution_class=atomic_replace`
  - `handoff.outgoing_owner_user_id`: read `outgoing_owner_user_id`; write target the `outgoing_owner_user_id` field on the underlying `handoff` artifact subtype; `conflict_resolution_class=atomic_replace`
  - `handoff.incoming_owner_user_id`: read `incoming_owner_user_id`; write target the `incoming_owner_user_id` field on the underlying `handoff` artifact subtype; `conflict_resolution_class=atomic_replace`
  - `handoff.current_state_summary`: read `current_state_summary`; write target the `current_state_summary` field on the underlying `handoff` artifact subtype; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
  - `handoff.open_task_ids`: read linked task-request references; write action upsert or remove same-incident `task_request` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
  - `handoff.open_decision_ids`: read linked decision references; write action upsert or remove same-incident `decision` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
  - `handoff.open_risk_refs`: read open risk references; write action upsert or remove structured risk references using the `risk_ref` family defined below; `conflict_resolution_class=collection_review`
  - `handoff.next_checks`: read `next_checks`; write target the `next_checks` field on the underlying `handoff` artifact subtype; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
  - `handoff.acknowledged_at`: read `acknowledged_at`; write target the `acknowledged_at` field on the underlying `handoff` artifact subtype; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
- collection-review wire contract:
  - `handoff.open_task_ids` and `handoff.open_decision_ids` MUST use the exact `collection_value_v1` item shape and `collection_actions_v1` action vocabulary defined for `assessment.support_refs` in `REQ-01-333` and `REQ-01-334`, except the active `field_key` is `handoff.open_task_ids` or `handoff.open_decision_ids` and the server derives `references_record` routing from that field key under `REQ-01-311`.
  - `handoff.open_task_ids` accepts only same-incident active `task_request` targets. `handoff.open_decision_ids` accepts only same-incident active `decision` targets.
  - For those two `record_ref` fields, duplicate adds for the same patched `record_id`, `linked_record_id`, and active `field_key` MUST coalesce to one surviving logical reference binding. Removal MUST target `item_ref` only. A target from another incident, a target of the wrong `record_type`, a soft-deleted target, or an invalid or foreign `item_ref` MUST fail with `400 Bad Request` and `error.code = invalid_mutation_payload`.
  - `handoff.open_risk_refs` MUST use `collection_value_v1` with `ordered=false`.
  - Each `items[]` entry MUST use this shape:

```json
{
  "item_ref": "risk_ref:<risk_ref_id>",
  "item_kind": "risk_ref",
  "display_text": "<risk_ref_text>",
  "risk_ref_id": "<risk_ref_id>",
  "risk_ref_text": "<source-preserving risk text>"
}
```

  - Allowed actions are:

```json
{ "op": "add_risk_ref", "risk_ref_text": "<text>" }
```

```json
{ "op": "remove_risk_ref", "item_ref": "risk_ref:<risk_ref_id>" }
```

  - `risk_ref_text` in `add_risk_ref` MUST use `string_contract_id=single_line_title_v1`.
  - Duplicate adds for the same patched `record_id`, active `field_key`, and normalized `risk_ref_text` under `single_line_title_v1` MUST coalesce to one surviving active risk reference. Removal MUST target `item_ref` only.
  - `risk_ref_id` MUST be a stable server-generated child-row identifier. Clients MUST NOT derive or predict `risk_ref_id` from raw text or a public hash.
  - An invalid or foreign `item_ref` MUST fail with `400 Bad Request` and `error.code = invalid_mutation_payload`.
- read-only computed or system-managed fields: `handoff.handoff_id`, `handoff.timestamp_day`, `handoff.ack_state`, `handoff.updated_at`
- `handoff.ack_state` MUST be `acknowledged` when `handoff.acknowledged_at` is non-null and `pending` otherwise
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-231, AC-282, AC-300, AC-301, AC-302, AC-303

**REQ-01-505**
- surface: required workbook-native coordination surface with canonical public identity `cartulary.view.status_review.v1`; any saved view over this same `view_schema_id` is a distinct saved-view object rather than the required base surface
- source record types: `artifact` filtered to `artifact_type='status_review'`
- base projection: `artifact_grid_projection` filtered to `artifact_type='status_review'`
- `default_visible_fields`: `status_review.timestamp_utc`, `status_review.review_owner_user_id`, `status_review.current_state_summary`, `status_review.active_risks_summary`, `status_review.next_report_at`, `status_review.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `status_review.status_review_id`, `status_review.blocked_task_ids`, `status_review.pending_evidence_ids`, `status_review.open_decision_ids`, `status_review.timestamp_day`, `status_review.next_report_day`
- `default_sort`: `status_review.timestamp_utc desc`, `record_id asc`
- `sort_fields`: `status_review.timestamp_utc`, `status_review.review_owner_user_id`, `status_review.current_state_summary`, `status_review.active_risks_summary`, `status_review.next_report_at`, `status_review.updated_at`, `status_review.timestamp_day`, `status_review.next_report_day`
- `filter_fields`: `status_review.timestamp_day`, `status_review.review_owner_user_id`, `status_review.next_report_day`
- `grouping_fields`: `status_review.timestamp_day`, `status_review.review_owner_user_id`, `status_review.next_report_day`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `status_review.current_state_summary` is non-empty after create-time normalization
- when omitted on create, the server MUST generate `status_review.status_review_id`, default `status_review.timestamp_utc` to the commit timestamp, default `status_review.review_owner_user_id` to the authenticated actor, default `status_review.blocked_task_ids`, `status_review.pending_evidence_ids`, and `status_review.open_decision_ids` to empty collections, and default `status_review.active_risks_summary` plus `status_review.next_report_at` to `null`
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `status_review.timestamp_utc`: read `timestamp_utc`; write target the `timestamp_utc` field on the underlying `status_review` artifact subtype; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=false`; `conflict_resolution_class=atomic_replace`
  - `status_review.review_owner_user_id`: read `review_owner_user_id`; write target the `review_owner_user_id` field on the underlying `status_review` artifact subtype; `conflict_resolution_class=atomic_replace`
  - `status_review.current_state_summary`: read `current_state_summary`; write target the `current_state_summary` field on the underlying `status_review` artifact subtype; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
  - `status_review.blocked_task_ids`: read linked blocked task references; write action upsert or remove same-incident `task_request` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
  - `status_review.pending_evidence_ids`: read linked pending evidence references; write action upsert or remove same-incident `evidence` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
  - `status_review.open_decision_ids`: read linked decision references; write action upsert or remove same-incident `decision` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
  - `status_review.active_risks_summary`: read `active_risks_summary`; write target the `active_risks_summary` field on the underlying `status_review` artifact subtype; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
  - `status_review.next_report_at`: read `next_report_at`; write target the `next_report_at` field on the underlying `status_review` artifact subtype; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=true`; `conflict_resolution_class=atomic_replace`
- collection-review wire contract:
  - `status_review.blocked_task_ids`, `status_review.pending_evidence_ids`, and `status_review.open_decision_ids` MUST use the exact `collection_value_v1` item shape and `collection_actions_v1` action vocabulary defined for `assessment.support_refs` in `REQ-01-333` and `REQ-01-334`, except the active `field_key` is the patched coordination field and the server derives `references_record` routing from that field key under `REQ-01-311`.
  - Accepted targets are same-incident active `task_request` for `status_review.blocked_task_ids`, same-incident active `evidence` for `status_review.pending_evidence_ids`, and same-incident active `decision` for `status_review.open_decision_ids`.
  - Duplicate adds for the same patched `record_id`, `linked_record_id`, and active `field_key` MUST coalesce to one surviving logical reference binding. Removal MUST target `item_ref` only. A target from another incident, a target of the wrong `record_type`, a soft-deleted target, or an invalid or foreign `item_ref` MUST fail with `400 Bad Request` and `error.code = invalid_mutation_payload`.
- read-only computed or system-managed fields: `status_review.status_review_id`, `status_review.timestamp_day`, `status_review.next_report_day`, `status_review.updated_at`
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-231, AC-283, AC-300, AC-301, AC-302, AC-303

**REQ-01-506**
- surface: required workbook-native coordination surface with canonical public identity `cartulary.view.lesson.v1`; any saved view over this same `view_schema_id` is a distinct saved-view object rather than the required base surface
- source record types: `artifact` filtered to `artifact_type='lesson'`
- base projection: `artifact_grid_projection` filtered to `artifact_type='lesson'`
- `default_visible_fields`: `lesson.timestamp_utc`, `lesson.summary`, `lesson.owner_user_id`, `lesson.closure_state`, `lesson.follow_up_task_ids`, `lesson.evidence_refs`, `lesson.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `lesson.lesson_id`, `lesson.timestamp_day`
- `default_sort`: `lesson.timestamp_utc desc`, `record_id asc`
- `sort_fields`: `lesson.timestamp_utc`, `lesson.summary`, `lesson.owner_user_id`, `lesson.closure_state`, `lesson.updated_at`, `lesson.timestamp_day`
- `filter_fields`: `lesson.closure_state`, `lesson.owner_user_id`, `lesson.timestamp_day`
- `grouping_fields`: `lesson.closure_state`, `lesson.owner_user_id`, `lesson.timestamp_day`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `lesson.summary` is non-empty after create-time normalization
- when omitted on create, the server MUST generate `lesson.lesson_id`, default `lesson.timestamp_utc` to the commit timestamp, default `lesson.owner_user_id` to the authenticated actor, default `lesson.follow_up_task_ids` and `lesson.evidence_refs` to empty collections, and default `lesson.closure_state` to `open`
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `lesson.timestamp_utc`: read `timestamp_utc`; write target the `timestamp_utc` field on the underlying `lesson` artifact subtype; `direct_scalar_contract_id=timestamp_instant_v1`; `clearable=false`; `conflict_resolution_class=atomic_replace`
  - `lesson.summary`: read `summary`; write target the `summary` field on the underlying `lesson` artifact subtype; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `lesson.owner_user_id`: read `owner_user_id`; write target the `owner_user_id` field on the underlying `lesson` artifact subtype; `conflict_resolution_class=atomic_replace`
  - `lesson.closure_state`: read `closure_state`; write target the `closure_state` field on the underlying `lesson` artifact subtype; `conflict_resolution_class=atomic_replace`
  - `lesson.follow_up_task_ids`: read linked follow-up task references; write action upsert or remove same-incident `task_request` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
  - `lesson.evidence_refs`: read linked evidence references; write action upsert or remove same-incident `evidence` references using the `record_ref` family defined below; `conflict_resolution_class=collection_review`
- collection-review wire contract:
  - `lesson.follow_up_task_ids` and `lesson.evidence_refs` MUST use the exact `collection_value_v1` item shape and `collection_actions_v1` action vocabulary defined for `assessment.support_refs` in `REQ-01-333` and `REQ-01-334`, except the active `field_key` is `lesson.follow_up_task_ids` or `lesson.evidence_refs` and the server derives `references_record` routing from that field key under `REQ-01-311`.
  - Accepted targets are same-incident active `task_request` for `lesson.follow_up_task_ids` and same-incident active `evidence` for `lesson.evidence_refs`.
  - Duplicate adds for the same patched `record_id`, `linked_record_id`, and active `field_key` MUST coalesce to one surviving logical reference binding. Removal MUST target `item_ref` only. A target from another incident, a target of the wrong `record_type`, a soft-deleted target, or an invalid or foreign `item_ref` MUST fail with `400 Bad Request` and `error.code = invalid_mutation_payload`.
- read-only computed or system-managed fields: `lesson.lesson_id`, `lesson.timestamp_day`, `lesson.updated_at`
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-231, AC-284, AC-300, AC-302, AC-303

**REQ-01-507**
- surface: standardized optional workbook surface for `cartulary.view.findings.v1`; this is the only current-profile standardized workbook surface for both findings and hypotheses; the current profile defines no `cartulary.view.hypotheses.v1`; when an implementation exposes this surface it MAY do so as a contract-backed system view or as an implementation-owned `scope='system'` saved view bound to this exact `view_schema_id`
- source record types: artifact-backed structured finding rows with exact `artifact_type='finding'` governed by Core 02 §10.4.5 and §10.4.6
- base projection: `artifact_grid_projection` filtered to exact `artifact_type='finding'`
- `default_visible_fields`: `finding.statement`, `finding.kind`, `finding.state`, `finding.owner_user_id`, `finding.confidence_score`, `finding.closed_at`, `finding.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `finding.supporting_refs`, `finding.contradictory_refs`, `finding.confidence_band`
- `default_sort`: `finding.updated_at desc`, `record_id asc`
- `sort_fields`: `finding.statement`, `finding.kind`, `finding.state`, `finding.owner_user_id`, `finding.confidence_score`, `finding.closed_at`, `finding.updated_at`, `finding.confidence_band`
- `filter_fields`: `finding.kind`, `finding.state`, `finding.owner_user_id`, `finding.confidence_band`, `finding.closed_at`
- `grouping_fields`: `finding.kind`, `finding.state`, `finding.owner_user_id`, `finding.confidence_band`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `finding.statement` is non-empty after create-time normalization
- when omitted on create, the server MUST default `finding.kind` to `finding`, default `finding.state` to `open`, default `finding.owner_user_id` to the authenticated actor, and default `finding.confidence_score` plus `finding.closed_at` to `null`
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `finding.statement`: read `statement`; write target the `statement` field on the underlying structured finding row; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
  - `finding.kind`: read `kind`; write target the `kind` field on the underlying structured finding row; legal writes MUST use the exact closed vocabulary defined in Core 02 §18; `conflict_resolution_class=atomic_replace`
  - `finding.state`: read `state`; write target the `state` field on the underlying structured finding row; legal writes MUST apply the server-managed `finding.closed_at` rule from Core 02 §10.4.6 before commit; `conflict_resolution_class=atomic_replace`
  - `finding.owner_user_id`: read `owner_user_id`; write target the `owner_user_id` field on the underlying structured finding row; `conflict_resolution_class=atomic_replace`
  - `finding.confidence_score`: read `confidence_score`; write target the `confidence_score` field on the underlying structured finding row; `conflict_resolution_class=atomic_replace`
  - `finding.supporting_refs`: read supporting record references; write action upsert or remove supporting structured references; `conflict_resolution_class=collection_review`
  - `finding.contradictory_refs`: read contradictory record references; write action upsert or remove contradictory structured references; `conflict_resolution_class=collection_review`
- read-only computed or system-managed fields: `finding.closed_at`, `finding.confidence_band`, `finding.updated_at`
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-231, AC-285

**REQ-01-508**
- surface: standardized optional workbook surface for `cartulary.view.investigative_queries.v1`; when an implementation exposes this surface it MAY do so as a contract-backed system view or as an implementation-owned `scope='system'` saved view bound to this exact `view_schema_id`
- source record types: artifact-backed structured investigative-query rows governed by Core 02 §10.4.6
- base projection: `artifact_grid_projection` filtered to the implementation's declared structured investigative-query subtype
- `default_visible_fields`: `investigative_query.platform`, `investigative_query.purpose`, `investigative_query.query_text`, `investigative_query.created_by_user_id`, `investigative_query.created_at`
- `default_hidden_fields`: `record_id`, `row_version`, `investigative_query.query_id`, `investigative_query.created_day`
- `default_sort`: `investigative_query.created_at desc`, `record_id asc`
- `sort_fields`: `investigative_query.platform`, `investigative_query.purpose`, `investigative_query.query_text`, `investigative_query.created_by_user_id`, `investigative_query.created_at`, `investigative_query.created_day`
- `filter_fields`: `investigative_query.platform`, `investigative_query.purpose`, `investigative_query.created_by_user_id`, `investigative_query.created_day`
- `grouping_fields`: `investigative_query.platform`, `investigative_query.created_by_user_id`, `investigative_query.created_day`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `investigative_query.platform`, `investigative_query.purpose`, and `investigative_query.query_text` are non-empty after create-time normalization
- when omitted on create, the server MUST generate `investigative_query.query_id`, default `investigative_query.created_by_user_id` to the authenticated actor, and default `investigative_query.created_at` to the commit timestamp
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `investigative_query.platform`: read `platform`; write target the `platform` field on the underlying structured investigative-query row; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `investigative_query.purpose`: read `purpose`; write target the `purpose` field on the underlying structured investigative-query row; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `investigative_query.query_text`: read `query_text`; write target the `query_text` field on the underlying structured investigative-query row; `string_contract_id=multiline_body_v1`; `conflict_resolution_class=text_compare_merge`
- read-only computed or system-managed fields: `investigative_query.query_id`, `investigative_query.created_by_user_id`, `investigative_query.created_at`, `investigative_query.created_day`
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-231, AC-286

**REQ-01-509**
- surface: standardized optional workbook surface for `cartulary.view.forensic_keywords.v1`; when an implementation exposes this surface it MAY do so as a contract-backed system view or as an implementation-owned `scope='system'` saved view bound to this exact `view_schema_id`
- source record types: artifact-backed structured forensic-keyword rows governed by Core 02 §10.4.6
- base projection: `artifact_grid_projection` filtered to the implementation's declared structured forensic-keyword subtype
- `default_visible_fields`: `forensic_keyword.pattern`, `forensic_keyword.reason`, `forensic_keyword.match_mode`, `forensic_keyword.case_sensitive`, `forensic_keyword.created_at`
- `default_hidden_fields`: `record_id`, `row_version`, `forensic_keyword.keyword_id`, `forensic_keyword.created_day`
- `default_sort`: `forensic_keyword.created_at desc`, `record_id asc`
- `sort_fields`: `forensic_keyword.pattern`, `forensic_keyword.reason`, `forensic_keyword.match_mode`, `forensic_keyword.case_sensitive`, `forensic_keyword.created_at`, `forensic_keyword.created_day`
- `filter_fields`: `forensic_keyword.match_mode`, `forensic_keyword.case_sensitive`, `forensic_keyword.created_day`
- `grouping_fields`: `forensic_keyword.match_mode`, `forensic_keyword.case_sensitive`, `forensic_keyword.created_day`
- inline create: zero-field create is forbidden
- minimum semantic create set: inline create from the sheet itself MUST commit only when `forensic_keyword.pattern` and `forensic_keyword.reason` are non-empty after create-time normalization
- when omitted on create, the server MUST generate `forensic_keyword.keyword_id`, default `forensic_keyword.match_mode` to `literal`, default `forensic_keyword.case_sensitive` to `false`, and default `forensic_keyword.created_at` to the commit timestamp
- these defaults MUST NOT satisfy the minimum create signal
- writable fields:
  - `forensic_keyword.pattern`: read `pattern`; write target the `pattern` field on the underlying structured forensic-keyword row; `string_contract_id=single_line_title_v1`; `conflict_resolution_class=text_compare_merge`
  - `forensic_keyword.reason`: read `reason`; write target the `reason` field on the underlying structured forensic-keyword row; `string_contract_id=reason_note_v1`; `conflict_resolution_class=text_compare_merge`
  - `forensic_keyword.match_mode`: read `match_mode`; write target the `match_mode` field on the underlying structured forensic-keyword row; `conflict_resolution_class=atomic_replace`
  - `forensic_keyword.case_sensitive`: read `case_sensitive`; write target the `case_sensitive` field on the underlying structured forensic-keyword row; `conflict_resolution_class=atomic_replace`
- read-only computed or system-managed fields: `forensic_keyword.keyword_id`, `forensic_keyword.created_at`, `forensic_keyword.created_day`
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-231, AC-287


## 20. Enterprise Authentication Extension Profile public contract

**REQ-01-510**
The Enterprise Authentication Extension Profile MUST expose exactly this minimum public route surface under `/api/v1/auth/*`:

- `GET /api/v1/auth/providers`,
- `POST /api/v1/auth/providers/{provider_key}/begin`,
- `GET /api/v1/auth/oidc/{provider_key}/callback`,
- `POST /api/v1/auth/saml/{provider_key}/acs`.
Profiles: enterprise_authentication
Verified by: AC-235, AC-288, AC-290, AC-291

Contract tables. The tables in §20 compact the enterprise-auth public contract into owner-local route, request, callback, and binding summaries. The surrounding prose remains authoritative for protocol-security details, callback validation, and provider-subject semantics that do not compress well into cells.

**Table 20-A. Enterprise-auth protocol route inventory**

| Route | Request contract summary | Omission and default summary | Idempotency | Success summary | Primary error codes |
| --- | --- | --- | --- | --- | --- |
| `GET /api/v1/auth/providers` | Singleton read; no body members | Rejects pagination members | Read route | Returns `data.providers[]` sorted by `display_name asc`, then `provider_key asc` | `invalid_pagination_request` |
| `POST /api/v1/auth/providers/{provider_key}/begin` | JSON object with optional `return_to` | Omitted or explicit `null` `return_to` normalizes to `/`; unknown members, including `client_txn_id`, are invalid | Intentionally non-idempotent; no `client_txn_id` accepted | Returns `provider_key`, `provider_type`, `redirect_url`, and `expires_at`; creates no public durable auth-transaction resource | `invalid_enterprise_auth_request`, `auth_provider_not_found`, `auth_provider_disabled` |
| `GET /api/v1/auth/oidc/{provider_key}/callback` | Browser protocol endpoint | Completes only against a valid single-use auth transaction | Intentionally non-idempotent | On success issues the ordinary server-managed session and completes with `303 See Other` to validated `return_to` | `enterprise_auth_transaction_rejected`, `provider_response_rejected`, `provider_identity_rejected` |
| `POST /api/v1/auth/saml/{provider_key}/acs` | Browser protocol endpoint | Completes only against a valid single-use auth transaction | Intentionally non-idempotent | On success issues the ordinary server-managed session and completes with `303 See Other` to validated `return_to` | `enterprise_auth_transaction_rejected`, `provider_response_rejected`, `provider_identity_rejected` |

**Table 20-B. Provider discovery and begin contract**

| Member or rule | Requirement |
| --- | --- |
| `GET /api/v1/auth/providers` item shape | Exactly `provider_key`, `provider_type`, and `display_name`; only enabled interactive providers are listed |
| `provider_type` vocabulary | Exactly `oidc` or `saml` on this route |
| `begin` request members | Optional `return_to` only |
| `return_to` rules | Same-origin relative-path reference only; omitted or explicit `null` normalize to `/` |
| `/` landing semantics | If exactly one visible incident membership exists, open that incident workbook with the ordinary Core 03 startup fallback; otherwise remain on `/` and render the visible incident list |
| Protocol transaction state | Server-side single-use auth transaction bound to provider, validated `return_to`, expiry, browser binding, and protocol correlation material |

**Table 20-C. Callback and ACS transport summary**

| Concern | Requirement |
| --- | --- |
| Success transport | Same server-managed session contract as §3.3.2.1 plus `303 See Other` to the validated `return_to` |
| Callback exception to JSON API rule | OIDC callback and SAML ACS are the only Enterprise Authentication family exception to the otherwise JSON-shaped public API |
| OIDC protocol requirements | Authorization-code flow only, with PKCE `S256` and `nonce` required |
| SAML protocol requirements | SP-initiated flow only |
| `client_txn_id` | Forbidden on begin, callback, and ACS |
| Replay and expiry | Replay, expiry, provider mismatch, or browser-binding mismatch fail closed rather than minting a fresh transaction |

**Table 20-D. Enterprise-auth protocol error summary**

| Condition | Transport | `error.code` | Registry or detail note |
| --- | --- | --- | --- |
| Malformed begin request or invalid `return_to` | `400` | `invalid_enterprise_auth_request` | Uses `request_not_object`, `field_not_nullable`, `unknown_field`, and `return_to_not_allowed` |
| Provider not found or disabled | Family-defined failure | `auth_provider_not_found` or `auth_provider_disabled` | Provider lookup and enablement state |
| Transaction replay, expiry, mismatch, or browser-binding failure | Family-defined failure | `enterprise_auth_transaction_rejected` | Uses `not_found`, `expired`, `already_used`, `provider_mismatch`, and `browser_binding_mismatch` |
| Provider response verification failure | Family-defined failure | `provider_response_rejected` | Uses the callback and ACS verification reasons in REQ-01-514 |
| No linked user, ambiguous link, or inactive user | Family-defined failure | `provider_identity_rejected` | Uses `subject_missing`, `no_linked_user`, `ambiguous_link`, and `inactive_user` |


**REQ-01-511**
`GET /api/v1/auth/providers` MUST return the common success envelope with `data.providers[]`, sorted by `display_name asc`, then `provider_key asc`. Each item MUST contain exactly `provider_key`, `provider_type`, and `display_name`; `provider_type` on this route MUST be `oidc` or `saml`. The route MUST list only enabled interactive providers and MUST NOT expose provider secrets, raw metadata, claim maps, or provider-side policy. The route MUST reject `limit`, `cursor_token`, and pagination aliases with `400`, `error.code = invalid_pagination_request`, and `error.details.reason_code = pagination_not_supported`.

`POST /api/v1/auth/providers/{provider_key}/begin` MUST accept only a JSON object with optional `return_to`. Omitted or explicit JSON `null` `return_to` MUST normalize to `/`. In the Enterprise Authentication Extension Profile, `/` is the canonical authenticated post-login landing route. After successful provider authentication, navigation to `/` MUST resolve in this order:

1. if the resulting session has exactly one current visible incident membership, the browser MUST open that incident's workbook without an explicit launch `sheet_ref`;
2. once that workbook open begins, startup-surface selection inside the workbook MUST use the ordered fallback owned by Core 03 §2.4;
3. if the resulting session has zero visible incident memberships or more than one visible incident membership, the browser MUST remain on `/` and the application MUST render the caller's visible incident list;
4. if step 1 selected a sole visible incident but that incident is no longer visible by workbook bootstrap time, the browser MUST remain on `/` and the application MUST render the caller's visible incident list.

When more than one visible incident exists, the implementation MUST NOT implicitly choose one incident by recency, sort order, provider claim content, or any other heuristic. `return_to` MUST be a same-origin relative-path reference. Unknown top-level members, including `client_txn_id`, MUST fail with `400` and `error.code = invalid_enterprise_auth_request`. A successful `begin` response MUST return the common success envelope with `data.provider_key`, `data.provider_type`, `data.redirect_url`, and `data.expires_at`; it MUST create no public durable auth-transaction resource.
Profiles: enterprise_authentication
Verified by: AC-235, AC-288, AC-289

**REQ-01-512**
The OIDC callback and SAML ACS routes are browser protocol endpoints and are the only Enterprise Authentication family exception to the otherwise JSON-shaped public API. On success they MUST issue the same server-managed session family defined by §3.3.2.1 and MUST complete with `303 See Other` to the validated `return_to`. When the validated `return_to` resolves to a workbook surface, startup-surface selection inside that workbook MUST use the ordered fallback in Core 03 §2.4. The Enterprise Authentication Extension Profile MUST NOT define a separate workbook-startup fallback order. On failure they MUST create no session and MUST use the common error envelope. The current profile standardizes OIDC authorization-code flow only, with PKCE `S256` and `nonce` required. Implicit and hybrid OIDC flows are non-conformant. The current profile standardizes SAML SP-initiated flow only. IdP-initiated SAML is non-conformant.

Enterprise-auth initiation MUST create one server-side single-use auth transaction bound at minimum to `provider_key`, `provider_type`, validated `return_to`, `started_at`, `expires_at`, browser-binding context, and protocol correlation material. For OIDC, correlation material MUST include `state`, `nonce`, and PKCE verifier material. For SAML, correlation material MUST include the SP-generated request correlation data and `RelayState`. `data.expires_at` returned by `begin` MUST be no later than 10 minutes after transaction creation. Replay, expiry, provider mismatch, or browser-binding mismatch MUST fail closed rather than silently minting a fresh transaction.
Profiles: enterprise_authentication
Verified by: AC-235, AC-290, AC-291

**REQ-01-513**
The authoritative provider-to-user bind key MUST be one active unique `(provider_id, provider_subject) -> user_id` mapping in deployment-local auth state. Enterprise-auth binding in the current profile is additive to an existing local user addressed by stable `user_id`; the current profile defines no enterprise-only user-creation path. `provider_subject` is the opaque authoritative external identifier and MUST compare by exact post-JSON-decoding code-point equality with no trimming, case-folding, Unicode normalization, or email-style normalization. For OIDC, the default authoritative subject is `sub`. For SAML, each provider configuration MUST declare exactly one stable authoritative subject source. `email`, `username`, `display_name`, and similar provider claims are secondary profile attributes only. Group claims are not authorization inputs in the current profile. Successful provider authentication MUST update `last_auth_at` on the resolved active binding only. Successful provider authentication MUST NOT auto-create a local user, auto-create an `auth_identity`, auto-create incident membership, or map provider groups into incident roles.
Profiles: enterprise_authentication
Verified by: AC-235, AC-292, AC-293

**REQ-01-514**
The Enterprise Authentication protocol routes under `/api/v1/auth/*` MUST use only `invalid_enterprise_auth_request`, `auth_provider_not_found`, `auth_provider_disabled`, `enterprise_auth_transaction_rejected`, `provider_response_rejected`, and `provider_identity_rejected`. `invalid_enterprise_auth_request` MUST use only `request_not_object`, `field_not_nullable`, `unknown_field`, and `return_to_not_allowed`. `enterprise_auth_transaction_rejected` MUST use only `not_found`, `expired`, `already_used`, `provider_mismatch`, and `browser_binding_mismatch`. `provider_response_rejected` MUST use only `missing_required_field`, `state_mismatch`, `relay_state_mismatch`, `nonce_mismatch`, `code_exchange_failed`, `issuer_mismatch`, `audience_mismatch`, `signature_invalid`, and `assertion_expired`. `provider_identity_rejected` MUST use only `subject_missing`, `no_linked_user`, `ambiguous_link`, and `inactive_user`.
Profiles: enterprise_authentication
Verified by: AC-235, AC-293

**REQ-01-515**
`POST /api/v1/auth/providers/{provider_key}/begin` is intentionally non-idempotent and MUST NOT accept `client_txn_id`. `GET /api/v1/auth/oidc/{provider_key}/callback` and `POST /api/v1/auth/saml/{provider_key}/acs` are intentionally non-idempotent protocol completion endpoints and MUST NOT accept `client_txn_id`. The current profile defines no public durable auth-transaction resource and no public transaction-recovery route. Replaying a structurally valid `begin` request MAY mint a fresh transaction and redirect target. Replaying a callback or ACS request after a successful completion or after transaction expiry MUST fail with `enterprise_auth_transaction_rejected` rather than succeeding a second time.
Profiles: enterprise_authentication
Verified by: AC-235, AC-289, AC-290, AC-291

**REQ-01-537**
The Enterprise Authentication Extension Profile MUST additionally expose these deployment-admin binding-management routes outside `/api/v1/auth/*`:

- `POST /api/v1/users/{user_id}/auth-bindings`,
- `POST /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}/rotate`,
- `DELETE /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}`.

These routes manage enterprise bindings only. They MUST NOT create, rotate, or retire the derived local binding summary.
Profiles: enterprise_authentication
Verified by: AC-348, AC-352

**Table 20-E. Binding-management route inventory**

| Route | Request contract summary | Idempotency | Success summary | Primary errors |
| --- | --- | --- | --- | --- |
| `POST /api/v1/users/{user_id}/auth-bindings` | Required `base_user_version`, `client_txn_id`, `provider_key`, `provider_subject`; optional `reason` | `(actor_user_id, user_id, client_txn_id)` | First success `201 Created`; returns the resulting safe user resource | `invalid_mutation_payload`, `user_version_conflict`, `auth_provider_not_found`, `auth_binding_conflict`, `client_txn_conflict` |
| `POST /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}/rotate` | Required `base_user_version`, `client_txn_id`, `new_provider_subject`; optional `reason` | `(actor_user_id, auth_binding_id, client_txn_id)` | `200 OK`; structural no-op when the new subject exactly equals the current active subject | `invalid_mutation_payload`, `user_version_conflict`, `auth_binding_not_found`, `auth_binding_conflict`, `client_txn_conflict` |
| `DELETE /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}` | Required `base_user_version`, `client_txn_id`; optional `reason` | `(actor_user_id, auth_binding_id, client_txn_id)` | `200 OK`; returns the resulting safe user resource | `invalid_mutation_payload`, `user_version_conflict`, `auth_binding_not_found`, `auth_binding_conflict`, `client_txn_conflict` |

**Table 20-F. Binding create, rotate, and retire request rules**

| Member or rule | Create | Rotate | Retire |
| --- | --- | --- | --- |
| `base_user_version` | Required | Required | Required |
| `client_txn_id` | Required | Required | Required |
| Provider selector | Required `provider_key`; must resolve to configured `oidc` or `saml` provider | Bound by existing `auth_binding_id`; provider does not change | Bound by existing `auth_binding_id` |
| Subject member | Required `provider_subject` | Required `new_provider_subject` | No subject member |
| `reason` | Optional; omission, explicit `null`, and normalized empty compare equal | Optional; same normalization rule | Optional; same normalization rule |
| Forbidden side effects | Must not create a local user, mutate incident membership, or mutate local credential state | Must not change provider, local email, local login identifier, local credential state, or incident memberships | Must not delete the local user, incident memberships, or local credential state |
| `client_txn_id` on protocol routes | Not applicable; protocol routes do not accept it | Not applicable | Not applicable |

**Table 20-G. Binding-management success and error summary**

| Condition | Transport | `error.code` or result | Notes |
| --- | --- | --- | --- |
| First successful create | `201 Created` | Safe user resource returned | Adds one enterprise binding to an existing local user |
| Successful rotate | `200 OK` | Safe user resource returned | Retires the old active binding and creates one replacement binding atomically |
| Successful retire | `200 OK` | Safe user resource returned | Removes the binding from active callback resolution and active `auth_bindings[]` summaries |
| Exact replay of committed success | `200 OK` | Original committed result | Evaluated before fresh version or binding-state checks |
| Same key, different normalized request | `409` | `client_txn_conflict` | Route-scoped idempotency failure |
| No current binding target for `{user_id, auth_binding_id}` | `404` | `auth_binding_not_found` | Current binding target only |
| Subject already in use, provider already linked for user, or binding not active | `409` | `auth_binding_conflict` | Uses the §3.3.6.2 reason-code registry |


**REQ-01-538**
`POST /api/v1/users/{user_id}/auth-bindings` binds one enterprise-auth provider subject to the existing local user addressed by `user_id`. The route MUST accept only a JSON object with required `base_user_version`, required `client_txn_id`, required `provider_key`, required `provider_subject`, and optional `reason`. `reason`, when present, MUST be a JSON string or JSON `null` and MUST normalize under `string_contract_id=reason_note_v1`. Unknown top-level members, a non-object body, a missing required member, or `null` for a non-nullable member MUST fail with `400` and `error.code = invalid_mutation_payload`. `provider_key` MUST identify a configured enterprise-auth provider of `provider_type='oidc'` or `provider_type='saml'`; a configured provider remains eligible for this route even when it is currently disabled for interactive sign-in. `provider_subject` MUST be a non-null JSON string. A first-time successful create MUST return `201 Created` with `data` equal to the resulting safe user resource. Route-scoped idempotency MUST be keyed by `(actor_user_id, user_id, client_txn_id)` and MUST compare exact `base_user_version`, exact `provider_key`, exact `provider_subject`, and normalized `reason`, with omitted `reason`, explicit JSON `null`, and any `reason` value that normalizes to empty under `reason_note_v1` comparing equal. Exact replay of a previously committed success MUST return `200 OK` with the original committed result before fresh `user_version_conflict` or binding-state evaluation. Reuse of the same key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. If no prior committed idempotency hit exists and the current `user_version` differs from `base_user_version`, the route MUST fail with `409` and `error.code = user_version_conflict`. If no configured enterprise provider matches `provider_key`, or the matched provider is not of type `oidc` or `saml`, the route MUST fail with `404` and `error.code = auth_provider_not_found`. If another active binding already uses the same `(provider_id, provider_subject)`, the route MUST fail with `409`, `error.code = auth_binding_conflict`, and `error.details.reason_code = provider_subject_in_use`. If the addressed user already has one active binding for that same provider, the route MUST fail with `409`, `error.code = auth_binding_conflict`, and `error.details.reason_code = provider_already_linked_for_user`. This route MUST NOT create a local user, mutate incident membership, or mutate local credential state.
Profiles: enterprise_authentication
Verified by: AC-348, AC-349

**REQ-01-539**
`POST /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}/rotate` MUST accept only a JSON object with required `base_user_version`, required `client_txn_id`, required `new_provider_subject`, and optional `reason`. `reason`, when present, MUST be a JSON string or JSON `null` and MUST normalize under `string_contract_id=reason_note_v1`. Unknown top-level members, a non-object body, a missing required member, or `null` for a non-nullable member MUST fail with `400` and `error.code = invalid_mutation_payload`. `new_provider_subject` MUST be a non-null JSON string. The route is valid only for one active enterprise binding addressed by the supplied `{user_id, auth_binding_id}` pair. Rotation MUST preserve the same `user_id` and provider, MUST retire the addressed active binding and create one replacement binding atomically in one commit, MUST preserve audit lineage, and MUST NOT re-key the old binding in place. A first-time successful rotate MUST return `200 OK` with `data` equal to the resulting safe user resource. If `new_provider_subject` is exactly equal to the current active subject, the route MUST return `200 OK` as a structural no-op, MUST return the current safe user resource, and MUST NOT advance `user_version`. Route-scoped idempotency MUST be keyed by `(actor_user_id, auth_binding_id, client_txn_id)` and MUST compare exact `base_user_version`, exact `new_provider_subject`, and normalized `reason`, with omitted `reason`, explicit JSON `null`, and any `reason` value that normalizes to empty under `reason_note_v1` comparing equal. Exact replay of a previously committed success MUST return `200 OK` with the original committed result before fresh `user_version_conflict` or binding-state evaluation. Reuse of the same key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. If no prior committed idempotency hit exists and the current `user_version` differs from `base_user_version`, the route MUST fail with `409` and `error.code = user_version_conflict`. If `{user_id, auth_binding_id}` identifies no current enterprise binding target, the route MUST fail with `404` and `error.code = auth_binding_not_found`. If the addressed binding is not active, the route MUST fail with `409`, `error.code = auth_binding_conflict`, and `error.details.reason_code = binding_not_active`. If another active binding already uses the same replacement `(provider_id, new_provider_subject)`, the route MUST fail with `409`, `error.code = auth_binding_conflict`, and `error.details.reason_code = provider_subject_in_use`. Rotation MUST NOT change the authoritative local email or local login identifier, MUST NOT mutate local credential state, and MUST NOT change incident memberships.
Profiles: enterprise_authentication
Verified by: AC-350, AC-352

**REQ-01-540**
`DELETE /api/v1/users/{user_id}/auth-bindings/{auth_binding_id}` MUST accept only a JSON object with required `base_user_version`, required `client_txn_id`, and optional `reason`. `reason`, when present, MUST be a JSON string or JSON `null` and MUST normalize under `string_contract_id=reason_note_v1`. Unknown top-level members, a non-object body, a missing required member, or `null` for a non-nullable member MUST fail with `400` and `error.code = invalid_mutation_payload`. The route is valid only for one active enterprise binding addressed by the supplied `{user_id, auth_binding_id}` pair. Retirement MUST remove that binding from active callback resolution and from active `auth_bindings[]` summaries, MUST preserve deployment-local audit history, MUST NOT delete the local user, MUST NOT delete incident memberships, and MUST NOT change local credential state or the authoritative local email or login identifier. A first-time successful retire MUST return `200 OK` with `data` equal to the resulting safe user resource. Route-scoped idempotency MUST be keyed by `(actor_user_id, auth_binding_id, client_txn_id)` and MUST compare exact `base_user_version` and normalized `reason`, with omitted `reason`, explicit JSON `null`, and any `reason` value that normalizes to empty under `reason_note_v1` comparing equal. Exact replay of a previously committed success MUST return `200 OK` with the original committed result before fresh `user_version_conflict` or binding-state evaluation. Reuse of the same key with a different normalized request MUST fail with `409` and `error.code = client_txn_conflict`. If no prior committed idempotency hit exists and the current `user_version` differs from `base_user_version`, the route MUST fail with `409` and `error.code = user_version_conflict`. If `{user_id, auth_binding_id}` identifies no current enterprise binding target, the route MUST fail with `404` and `error.code = auth_binding_not_found`. If the addressed binding is not active, the route MUST fail with `409`, `error.code = auth_binding_conflict`, and `error.details.reason_code = binding_not_active`.
Profiles: enterprise_authentication
Verified by: AC-351, AC-352

**REQ-01-541**
The binding-management routes in REQ-01-537..REQ-01-540 are part of the Enterprise Authentication Extension Profile but are not part of the `/api/v1/auth/*` protocol route family. They bind only existing local users addressed by stable `user_id`. Successful provider-auth callback MUST update `last_auth_at` on the resolved active binding only. After a successful rotate or retire, the superseded or retired `provider_subject` MUST fail future callbacks with `409`, `error.code = provider_identity_rejected`, and `error.details.reason_code = no_linked_user`. Switching providers is not a rotate operation in the current profile; the supported path is create a new active binding for the second provider and, if desired, retire the prior provider binding separately. These binding-management routes MUST reuse the common success and error envelopes, MUST reuse `auth_provider_not_found`, `user_version_conflict`, and `client_txn_conflict` where applicable, MUST use `auth_binding_not_found` when `{user_id, auth_binding_id}` identifies no visible current binding target, and MUST use `auth_binding_conflict` with the reason-code registry in §3.3.6.2 for active-binding state conflicts.
Profiles: enterprise_authentication
Verified by: AC-349, AC-350, AC-351, AC-352
