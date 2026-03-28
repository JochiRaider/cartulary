# Cartulary Normative Core 01: Architecture, Storage, and View Contracts

## 1. Architecture pattern

**REQ-01-001**
Cartulary MUST use a **modular monolith** architecture for the base profile.
Profiles: base
Verified by: AC-231

**REQ-01-002**
The base deployment topology MUST consist of:

- one web application deployable that contains the browser-facing UI, API surface, WebSocket hub, and background-job runners,
- one Postgres service as the authoritative structured data store,
- one S3-compatible object storage service as the authoritative binary evidence store.
Profiles: base
Verified by: AC-231

**REQ-01-003**
The application deployable MUST remain a single deployable unit even when deployed behind a reverse proxy or onto managed infrastructure.
Profiles: base
Verified by: AC-231

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
- a local pending-patch queue for transient network interruptions,
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
The HTTP surface MUST be rooted at `/api/v1/`. The live-update stream MUST be rooted at `/ws/v1/` or an equivalent versioned WebSocket path that preserves the same observable contract.
Profiles: base
Verified by: AC-124, AC-125, AC-127, AC-131, AC-135, AC-231

**REQ-01-021**
Breaking changes to route patterns, required request fields, required response fields, envelope shapes, or event semantics MUST use a new major version root. Additive response fields, additive optional request fields, and additive route families for claimed extension profiles MAY be introduced within the same major version.
Profiles: base
Verified by: AC-124, AC-125, AC-127, AC-131, AC-135, AC-231

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
- `GET /api/v1/auth/session`.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231

**REQ-01-025**
`POST /api/v1/auth/login` MUST be the base-profile local-account login route. The request body MUST be a JSON object and MUST accept:

- required `username`,
- required `password`,
- optional `second_factor`.

`username` is the deployment-local login identifier for this route. This requirement does not define provisioning rules for that identifier. `username` MUST be trimmed of leading and trailing Unicode whitespace, MUST be non-null, and MUST be non-empty after trimming. `password` MUST be a non-null non-empty string and MUST be compared exactly as supplied after JSON decoding. The server MUST NOT trim, case-fold, or Unicode-normalize `password`.

When `second_factor` is omitted, the request is a primary-credentials-only login attempt. When present, `second_factor` MUST be an object and MUST be non-null. `second_factor.kind` MUST be present and, in the base profile, MUST use the closed vocabulary `totp`. `second_factor.assertion` MUST be present, MUST be an object, and MUST be non-null. For `kind='totp'`, `second_factor.assertion` MUST use exactly this shape: `{ "code": "123456" }`. `code` MUST be a string of exactly six ASCII decimal digits with no spaces or separators.

Unknown top-level request members, unknown `second_factor` members, unknown `assertion` members for the selected `kind`, a missing required member, a type mismatch, a supplied `null` where this requirement forbids `null`, a `second_factor.kind` outside the base-profile vocabulary, or any `client_txn_id`, `id_token`, `authorization_code`, `saml_response`, `provider_assertion`, or WebAuthn ceremony field sent on this route MUST fail with `400` and `error.code = invalid_auth_request`. When the failure is attributable to one request member, `error.details.field` MUST identify that member.

If the local account does not require MFA, omission of `second_factor` MUST be accepted and a structurally valid `second_factor` MUST NOT by itself prevent login. If the local account requires MFA and `second_factor` is omitted, the server MUST fail with `401` and `error.code = mfa_required`; `error.details.required_second_factor_kinds` MUST equal `["totp"]` in the base profile. If the server is not willing to acknowledge that primary credentials were valid, including for unknown `username`, wrong `password`, inactive local account, or equivalent pre-MFA failure, it MUST fail with `401` and `error.code = invalid_credentials`. If primary credentials are valid and a structurally valid TOTP assertion is present but wrong or expired, the server MUST fail with `401` and `error.code = invalid_second_factor`.

This route MUST NOT require or interpret `client_txn_id`. On success it MUST establish the server-managed session and return the same session resource exposed by `GET /api/v1/auth/session`. On any non-success outcome, the server MUST create no session, set no session cookie, and expose no partial or pre-authenticated session state. Transport retries after an uncertain network boundary are not idempotent and MAY create a fresh session if the client repeats the request.
Profiles: base
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231, AC-244, AC-245, AC-246, AC-247, AC-248, AC-249, AC-250

Example request body:

```json
{
  "username": "analyst1",
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

**REQ-01-027**
`GET /api/v1/auth/session` MUST return, at minimum, the authenticated internal user identity, display name, provider kind, MFA state, `is_deployment_admin`, `authenticated_at`, `idle_expires_at`, `absolute_expires_at`, `session_expires_at`, and the caller's incident memberships or current incident-role context. `session_expires_at` MUST be the earlier of `idle_expires_at` and `absolute_expires_at`.
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
If the Enterprise Authentication Extension Profile is implemented, provider-backed sign-in MUST terminate into this same session contract so the remaining API routes and WebSocket stream remain provider-agnostic.
Profiles: base, enterprise_authentication
Verified by: AC-123, AC-130, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-231, AC-235, AC-250

#### 3.3.3 Route families

**REQ-01-032**
The base-profile route set MUST include stable route families for:

- incident discovery, creation, retrieval, and incident metadata mutation: `POST /api/v1/incidents`, `GET /api/v1/incidents`, `GET /api/v1/incidents/{incident_id}`, `PATCH /api/v1/incidents/{incident_id}`,
- deployment-local user account inspection and administration: `GET /api/v1/users`, `POST /api/v1/users`, `GET /api/v1/users/{user_id}`, `PATCH /api/v1/users/{user_id}`,
- incident membership inspection and administration: `GET /api/v1/incidents/{incident_id}/memberships`, `POST /api/v1/incidents/{incident_id}/memberships`, `PATCH /api/v1/incidents/{incident_id}/memberships/{user_id}`, `DELETE /api/v1/incidents/{incident_id}/memberships/{user_id}`,
- view-schema discovery: `GET /api/v1/view-schemas`, `GET /api/v1/view-schemas/{view_schema_id}`,
- saved-view discovery and persistence: `GET /api/v1/incidents/{incident_id}/saved-views`, `POST /api/v1/incidents/{incident_id}/saved-views`, `PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`, `DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`,
- workbook-preference discovery and persistence: `GET /api/v1/incidents/{incident_id}/workbook-preferences/me`, `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me`, `GET /api/v1/incidents/{incident_id}/workbook-preferences/default`, `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default`,
- workbook query and row creation: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`,
- record mutation, explicit Timeline capture-state actions, soft-delete, restore, history, rollback, and same-field conflict resolution: `PATCH /api/v1/records/{record_id}`, `POST /api/v1/records/{record_id}/mark-reviewed`, `POST /api/v1/records/{record_id}/supersede`, `DELETE /api/v1/records/{record_id}`, `POST /api/v1/records/{record_id}/restore`, `GET /api/v1/records/{record_id}/history`, `POST /api/v1/records/{record_id}/rollback`, `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve`,
- entity merge initiation: `POST /api/v1/records/{survivor_record_id}/merge`,
- entity-mention explicit action route and equivalent surface-specific single-mention actions: `POST /api/v1/entity-mentions/{entity_mention_id}/resolve`,
- blob-slot creation and evidence access: `POST /api/v1/object-blobs`, `POST /api/v1/evidence-records/{record_id}/attach-blob`, `POST /api/v1/evidence-records/{record_id}/preview-handle`, `POST /api/v1/evidence-records/{record_id}/download-handle`, `GET /api/v1/evidence-handles/{handle_token}`,
- background-job status and cancellation: `GET /api/v1/jobs/{job_id}`, `POST /api/v1/jobs/{job_id}/cancel`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-186, AC-187, AC-231, AC-251, AC-252, AC-253, AC-254, AC-255

**REQ-01-033**
Implementations that claim an extension profile MUST add that profile's route family under the same versioned root rather than overloading base workbook routes. This includes, at minimum, `/api/v1/import-sessions/*`, `/api/v1/reference-packs/*`, `/api/v1/snapshots/*` and `/api/v1/releases/*`, and `/api/v1/incident-bundles/*` for the corresponding claimed extension profiles.
Profiles: base, snapshot_reporting, reference_pack
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-186, AC-187, AC-231, AC-233, AC-234

#### 3.3.4 View-shaped read contract

**REQ-01-034**
The primary hot-path read route for workbook surfaces MUST be `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-035**
A view-query request MUST be view-shaped rather than table-shaped. It MUST accept:

- ordered `sort[]` entries keyed by `field_key`,
- `filters[]` entries keyed by `field_key`,
- zero or one `group_by` value chosen from the view's declared grouping keys,
- pagination members defined by §3.3.7.

For `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, `limit` and `cursor_token` MUST appear only as JSON-body members, not as query parameters. `limit` counts serialized `rows[]` entries only. A malformed pagination member, unsupported pagination alias, or cursor replay against a different bound view-query contract MUST fail with `400`, `error.code=invalid_view_query`, and `error.details.reason_code` equal to `invalid_limit` or `cursor_query_mismatch` from §3.3.6.2, as applicable.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-238, AC-239, AC-240, AC-243

**REQ-01-036**
A view-query response MUST return:

- `incident_id`,
- `view_schema_id`,
- `rows[]` already shaped for that view,
- `record_id` and `row_version` on every mutable row,
- field-key-addressable cell values,
- scalar `group_values` needed for client-local grouping when grouping is active,
- pagination metadata defined in §3.3.7.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-238, AC-241, AC-243

**REQ-01-037**
The server MUST NOT serialize group headers or other presentation-only grouping artifacts as writable rows.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231, AC-243

##### 3.3.4.1 Filter predicate wire contract

**REQ-01-038**
A `filters[]` entry MUST use this shape:
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

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
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

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
- an empty `prefix.value` after trimming,
- an empty `full_text.query` after trimming,
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

The canonical `invalid_view_query` `error.details.reason_code` registry is defined in §3.3.6.2.

Request order of `filters[]` is not semantically significant.

**REQ-01-046**
For saved-view persistence, cursor binding, and `meta.query`, the server MUST normalize `filters[]` into canonical order by `field_key asc`, then emit or persist the exact same wire shape defined in this subsection.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

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
For a soft-deleted record, `row_version` MUST expose the current tombstone `row_version`, or an equivalent current revision version token accepted by `POST /api/v1/records/{record_id}/restore`.
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
An `items[]` entry MUST include `history_entry_ref` if and only if that logical history item maps to exactly one reversible mutation target. The client MUST treat `history_entry_ref` as opaque. For the lifetime of retained history, the same logical history item MUST keep the same `history_entry_ref` across repeated reads.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-055**
An `items[]` entry MUST include `revision_no` if and only if whole-row restore is a legal action for that displayed logical history item.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-231

**REQ-01-056**
The route MUST accept `limit` and `cursor_token` under §3.3.7 whenever more than one page is possible. Pagination MUST preserve the item-ordering rules in this subsection and the cursor MUST remain bound to `record_id`.
Profiles: base
Verified by: AC-124, AC-127, AC-184, AC-185, AC-215, AC-231

#### 3.3.5 Mutation contract

**REQ-01-057**
New row creation MUST use `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`. The request MUST include `client_txn_id` and the initial writable values keyed by `field_key`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-058**
Existing-row edits MUST use `PATCH /api/v1/records/{record_id}`. The request MUST include:

- `view_schema_id`,
- `base_row_version`,
- `client_txn_id`,
- `changes[]`, with each entry keyed by `field_key` and carrying the intended write value or equivalent action payload.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-059**
Each `changes[]` entry MUST contain `field_key` and exactly one of `value` or `action_payload`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-060**
A field whose active `view_schema` entry declares a direct write target MUST use `value` in `PATCH /api/v1/records/{record_id} changes[]`. A field whose active `view_schema` entry declares a write action MUST use `action_payload` in `PATCH /api/v1/records/{record_id} changes[]`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-061**
When `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` supplies initial writable values keyed directly by `field_key`, a direct-write field MUST use its direct field value as the JSON value and a write-action field MUST use the same object that a patch `changes[]` entry would carry in `action_payload`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-062**
For any field whose active `view_schema` entry declares `conflict_resolution_class=collection_review`, `action_payload.kind` MUST be `collection_actions_v1`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-063**
A `collection_actions_v1` payload MUST use this shape:
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

```json
{
  "kind": "collection_actions_v1",
  "actions": [
    { "... field-specific action object ..." }
  ]
}
```

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
The server MUST validate each requested change against the active view contract, enforce per-field writeability and `conflict_resolution_class`, and route the write to the authoritative source field or declared write action without exposing internal table layout. If mutation payload validation fails, the server MUST fail with `400 Bad Request` using the common error envelope and `error.code = invalid_mutation_payload`. At minimum, this applies to an unknown payload `kind`, an unknown action `op`, an action not allowed for the active `field_key`, an invalid or foreign `item_ref`, a malformed direct value, or an action object that contains unknown members for its declared `op`. For `collection_actions_v1`, the server MUST apply `actions[]` atomically and in request order within the parent field mutation. The public API surface MUST NOT require raw comma-delimited strings or blind full-collection replacement for `collection_review` fields.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-070**
A successful create or patch response MUST return the authoritative `record_id`, resulting `row_version`, and the committed field values needed to refresh the visible row.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-071**
Soft-delete of a user-visible first-class record MUST use `DELETE /api/v1/records/{record_id}`. Restore of a currently soft-deleted first-class record MUST use `POST /api/v1/records/{record_id}/restore`. Both routes are record-scoped, not view-scoped. They MUST NOT require `incident_id` or `view_schema_id` in the path or request body, because authorization, history, and affected projections derive from the authoritative `record_id`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-072**
The delete request MUST be JSON and MUST include `base_row_version` and `client_txn_id`. It MAY include optional `reason`. The restore request MUST be JSON and MUST include `base_row_version` and `client_txn_id`. It MAY include optional `reason`. For idempotency comparison, omission and explicit JSON `null` for `reason` MUST compare equal. Idempotency for delete and restore MUST be keyed by `(actor_user_id, record_id, client_txn_id)`. If the same authenticated actor replays the same normalized request with the same key, the server MUST return `200 OK` with the originally committed result and MUST create no second mutation. If the same actor reuses that key with a different normalized request, the server MUST fail with `409`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-073**
Because a soft-deleted record no longer appears in ordinary workbook queries, `GET /api/v1/records/{record_id}/history` MUST expose the current tombstone `row_version`, or an equivalent current revision version token accepted by `POST /api/v1/records/{record_id}/restore`, so an authorized reviewer can satisfy optimistic concurrency without requiring a separate general record-read route.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-074**
`DELETE /api/v1/records/{record_id}` MUST require current incident role `editor`, `reviewer`, or `admin`. `POST /api/v1/records/{record_id}/restore` MUST require current incident role `reviewer` or `admin`. A caller who lacks visibility to the target record MUST receive `404`. A caller who can see the record but lacks sufficient role MUST receive `403`. Soft-delete SHOULD use ordinary optimistic concurrency rather than a hard lock.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-075**
If the current `row_version` differs from `base_row_version`, either route MUST fail with `409` using the common error envelope and `error.code = row_version_conflict`. `error.details` MUST include at least `record_id`, `base_row_version`, and `current_row_version`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-076**
`PATCH /api/v1/records/{record_id}` against a currently soft-deleted record MUST fail with `409` and `error.code = record_deleted_use_restore`. `DELETE /api/v1/records/{record_id}` against an already soft-deleted record MUST fail with `409` and `error.code = record_already_deleted`, except that idempotent replay of the same normalized delete request by the same actor with the same `(record_id, client_txn_id)` MUST return the original success response. `POST /api/v1/records/{record_id}/restore` against a record that is not currently soft-deleted MUST fail with `409` and `error.code = record_not_deleted`, except that idempotent replay of the same normalized restore request by the same actor with the same `(record_id, client_txn_id)` MUST return the original success response.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-077**
The server MAY acquire a short-lived server-side record lock while applying `POST /api/v1/records/{record_id}/restore`. If such a lock is used, it MUST be internal to the restore operation and MUST NOT require a separate client-visible generic lock-acquire route in the base profile. If another destructive operation currently holds that record lock, restore MUST fail with `409`, `error.code = record_locked`, and `error.retryable = true`.
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

- require JSON `base_row_version` and `client_txn_id`; it MAY include optional `reason`,
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

- require JSON `base_row_version`, `client_txn_id`, and non-empty `reason`; it MAY include optional `replacement_record_id`,
- apply only to a visible non-deleted `timeline_event` record,
- require current incident role `reviewer` or `admin`,
- succeed only when the current persisted `capture_state` is `rough`, `enriched`, or `reviewed`,
- when `replacement_record_id` is present, require a different visible record in the same incident and, if the implementation persists that reference, store it through an authoritative typed `record_link` or an equivalent structured relation,
- in one transaction, increment `row_version`, create one attributed `change_set`, persist `capture_state='superseded'`, update derived projections, and return `200 OK` with at least `record_id`, `incident_id`, `row_version`, `capture_state`, `change_set_id`, and `reason`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-087**
If the same authenticated actor replays the same normalized request with the same `(record_id, client_txn_id)`, the server MUST return `200 OK` with the originally committed result and MUST create no second lifecycle transition. A caller who lacks visibility to the target record MUST receive `404`. A caller who can see the record but lacks sufficient role MUST receive `403`. A stale `base_row_version` MUST fail with `409` and `error.code = row_version_conflict`. A request against a non-Timeline record, a Timeline record whose current `capture_state` is already `superseded`, or a `replacement_record_id` that is invalid for the target record MUST fail with `409` and `error.code = illegal_transition`. A request against a currently soft-deleted record MUST fail with `409` and `error.code = record_deleted_use_restore`.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-01-088**
A successful call to either route MUST emit the ordinary replayable `record_changed` event for the affected record.
Profiles: base
Verified by: AC-125, AC-126, AC-181, AC-182, AC-183, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-211, AC-212, AC-213, AC-214, AC-215, AC-216, AC-217, AC-218, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

#### 3.3.5.0 Rollback contract

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

It MAY include optional `reason`.

**REQ-01-091**
Omission and explicit JSON `null` for `reason` MUST compare equal for idempotency.
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
Single-entry rollback through `history_entry_ref` MUST be available only when that history item maps to exactly one reversible mutation target.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

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
The server MAY acquire short-lived internal destructive-operation locks on all affected records while applying the rollback. If such locks are used, they MUST be acquired in canonical ascending `record_id` order and MUST NOT require a separate client-visible lock route. If another destructive operation currently holds a required lock, the server MUST fail with `409`, `error.code = record_locked`, and `error.retryable = true`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

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
`auth_bindings[]` MUST be read-only and MUST contain only safe binding summaries such as `provider_type`, `provider_key`, optional `username`, `created_at`, and `last_auth_at`. The base profile MUST NOT require clients to interpret `auth_bindings[]` in order to authorize incident data access.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-117**
`GET /api/v1/users` and `GET /api/v1/users/{user_id}` MUST fail closed unless the caller has the deployment-scoped account-administration capability defined by Core 04 §2. `GET /api/v1/users` MUST return safe user resources ordered by `user_id asc`, MUST use the common success envelope with `data.users[]` plus `meta.paging`, and MUST accept only `limit` and `cursor_token` under §3.3.7. Pagination failures on this route MUST fail with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code` from the `invalid_pagination_request` registry in §3.3.6.2. Inert imported historical actors from incident portability are not login-capable user resources and MUST NOT be returned by this route family unless they have been explicitly mapped to a local user account.
Profiles: base
Verified by: AC-127, AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-118**
`GET /api/v1/users/{user_id}` MUST return the common success envelope with `data` equal to the requested safe user resource. Because this route is singleton, it MUST reject `limit`, `cursor_token`, and pagination aliases with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code=pagination_not_supported`.
Profiles: base
Verified by: AC-127, AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-119**
`POST /api/v1/users` MUST require the same deployment-scoped account-administration capability. The route MUST accept:

- required `client_txn_id`,
- required `auth_kind`,
- required `email`,
- required `display_name`,
- required `initial_password`,
- optional `mfa_required`,
- optional `is_deployment_admin`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-120**
In the base profile, `auth_kind` MUST be `local`. `email` MUST be trimmed of leading and trailing ASCII whitespace, MUST be non-empty after trimming, and MUST be unique within the deployment after stable case-insensitive canonicalization or an equivalent deterministic uniqueness substrate. The server MUST NOT expose `initial_password` or any equivalent secret in a response or event payload.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-121**
The first deployment admin MUST be provisioned out of band at install or bootstrap time. The public API MUST NOT allow unauthenticated bootstrap of a deployment-admin user.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-122**
Idempotency for user create MUST be keyed by `(actor_user_id, client_txn_id)`. If the same authenticated actor replays the same normalized request with the same `client_txn_id`, the server MUST return `200 OK` with the originally created safe user resource and MUST create no second user. If the same actor reuses `client_txn_id` with a different normalized request, the server MUST fail with `409`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-123**
A first-time successful `POST /api/v1/users` call MUST return `201 Created` and the common success envelope with `data` equal to the created safe user resource.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-124**
`PATCH /api/v1/users/{user_id}` MUST require the same deployment-scoped account-administration capability. The route MUST accept `base_user_version` plus changed mutable fields only. In the base profile, mutable fields are `email`, `display_name`, `is_active`, `mfa_required`, and `is_deployment_admin`. The route MUST reject attempted mutation of `user_id`, `created_at`, `last_login_at`, or `auth_bindings[]`. If the current `user_version` differs from `base_user_version`, the server MUST fail with `409` and `error.code = user_version_conflict`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-125**
Changing `is_active` from `true` to `false` MUST revoke all active sessions for that user immediately and MUST NOT delete that user's incident memberships. A patch that would deactivate or demote the last active `is_deployment_admin=true` user MUST fail with `409` and `error.code = last_deployment_admin`.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

**REQ-01-126**
A successful `PATCH /api/v1/users/{user_id}` call MUST return the common success envelope with `data` equal to the resulting safe user resource.
Profiles: base
Verified by: AC-175, AC-176, AC-177, AC-178, AC-179, AC-180, AC-231

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
`role` MUST use the closed vocabulary `viewer`, `editor`, `reviewer`, and `admin`. If `email` is supplied, the server MUST resolve it through the same canonical email lookup used by the user-account contract and MUST bind stored membership state to the resolved `user_id`. The base profile MUST NOT auto-create or invite a user from this route. If the target user does not exist, the server MUST fail with `404` and `error.code = user_not_found`. If the target user exists but `is_active=false`, the server MUST fail with `409` and `error.code = user_inactive`.
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
`query_json` MUST be normalized against the owning `view_schema_id` and encode saved-view sort, filter, and grouping state using stable `field_key` values, ordered `sort[]`, ordered `filters[]`, optional `group_by`, and normalized scalar values. It MUST NOT use visible tab labels, visible column labels, presentation-only group-header text, or other display-only identifiers. `query_json.filters[]` MUST use the exact filter predicate wire shape defined in §3.3.4.1. Persisted normalization MUST preserve the operator-specific `arg` object, normalized scalar values, and canonical `filters[]` ordering by `field_key asc`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-143**
`layout_json` MAY encode presentation concerns such as column order, hidden fields, widths, inspector openness, and equivalent client layout state. `layout_json` MUST NOT be the authority for `saved_view_id`, `incident_id`, `view_schema_id`, `scope`, ownership, authorization, or startup/default surface selection.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-144**
`GET /api/v1/incidents/{incident_id}/saved-views` MUST return only the saved-view resources visible to the caller, MUST use the common success envelope with `data.saved_views[]` plus `meta.paging`, MUST order results by `updated_at desc, saved_view_id asc`, and MUST accept only `limit` and `cursor_token` under §3.3.7. Pagination failures on this route MUST fail with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code` from the `invalid_pagination_request` registry in §3.3.6.2.
Profiles: base
Verified by: AC-127, AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-145**
`POST /api/v1/incidents/{incident_id}/saved-views` MUST accept `view_schema_id`, optional `scope`, `display_name`, `query_json`, and `layout_json`. If `scope` is omitted, the server MUST treat it as `private`. The ordinary public create route MUST reject `scope='system'`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-146**
`PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` MUST accept `base_saved_view_version` plus changed mutable fields only. It MUST reject attempted mutation of `incident_id`, `saved_view_id`, or `view_schema_id`. If the current saved-view version differs from `base_saved_view_version`, the server MUST reject the patch with an explicit conflict status rather than silently overwriting saved-view state.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

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
Both workbook-preference resources MUST use the stable `sheet_ref` union defined in §3.3.10.1.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-150**
`workbook-preferences/me` MUST expose, at minimum, `incident_id`, `user_id`, `home_sheet_ref`, `created_at`, and `updated_at`. `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` MUST accept a nullable `home_sheet_ref` and MUST allow any current incident member to set or clear only their own home-surface preference.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-01-151**
`workbook-preferences/default` MUST expose, at minimum, `incident_id`, `default_sheet_ref`, `created_at`, `updated_at`, and `updated_by_user_id`. `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` MUST accept a nullable `default_sheet_ref` and MUST fail closed for callers whose current incident role is not `admin`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

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
The server MUST ignore unknown extension fields so additive optional request fields remain forward-compatible within the same major version. Unknown extension fields MUST be excluded from normalized-request comparison and MUST NOT affect persisted state, authorization, audit history, or response shaping. The server MUST reject any attempt to set server-managed fields including `incident_id`, `status`, `created_by_user_id`, `created_at`, `updated_at`, `updated_by_user_id`, `incident_version`, `closed_at`, any membership object, any saved-view object, and any workbook-preference object.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-160**
For idempotency comparison, the normalized request MUST include only recognized request fields after the validation and normalization rules in this section are applied. For all optional request fields in this contract, omission and an explicit JSON `null` MUST compare equal.
Profiles: base
Verified by: AC-170, AC-171, AC-172, AC-173, AC-174, AC-211, AC-212, AC-213, AC-214, AC-219, AC-220, AC-231

**REQ-01-161**
If the request is not a JSON object, omits required `client_txn_id`, `incident_key`, or `title`, supplies `null` for a non-nullable field, violates a field validation rule in this section, attempts to set a server-managed field, or supplies `initial_memberships[]` or an equivalent collaborator-seeding payload, the server MUST fail with `400` and `error.code = invalid_incident_create`. When the failure is attributable to one request member, `error.details.field` MUST identify that top-level member. `error.details.reason_code` MUST use the registry in §3.3.6.2.
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
The base profile MUST NOT accept `initial_memberships[]` or any equivalent collaborator-seeding payload on this route. Membership management beyond creator bootstrap belongs to the separate membership-write contract defined by §3.3.5.1.
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
It MAY include optional `reason`. For idempotency comparison, omission and explicit JSON `null` for `reason` MUST compare equal.
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
The server MAY acquire short-lived internal destructive-operation locks on both records while applying the merge. If such locks are used, they MUST be acquired in canonical `record_id` order and MUST NOT require a separate client-visible lock-acquire route. If either required lock is already held, the route MUST fail with `409`, `error.code = record_locked`, and `error.retryable = true`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-191**
On success, the server MUST commit the merge in one transaction. It MUST:

- preserve the survivor `record_id` unchanged,
- mark the loser as a historical row with state `merged` and `merged_into_record_id` set to the survivor,
- repoint active `entity_mentions.resolved_record_id`, active `record_links`, active assessments, and active tags from loser to survivor in the same `change_set`, or otherwise tombstone and recreate them deterministically,
- deduplicate duplicate links and tags created by repointing without losing revision history,
- preserve raw mention text unchanged,
- create one attributed `change_set` plus ordered mutation entries sufficient to reconstruct the pre-merge graph, post-merge graph, and merge fan-out,
- update or invalidate affected projections before commit returns.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-192**
In accordance with Core 02 §9, the server SHOULD copy non-conflicting aliases and seed identifiers from loser to survivor. When it does so, it MUST NOT overwrite conflicting survivor-owned values.
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
- `merged_into_record_id`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

The server SHOULD also return `merge_summary` with counts for repointed mentions, repointed links, repointed assessments, repointed tags, deduplicated links, deduplicated tags, and copied aliases.

**REQ-01-194**
A successful merge MUST emit replayable collaboration events through the existing stream. The loser MUST leave ordinary active entity views using `change_kind = remove`. The survivor MUST refresh through `patch` or `invalidate`. Any dependent row whose chips, counts, or linked-record summaries change because of repointing MUST refresh through the ordinary collaboration mechanisms.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-01-195**
The base profile defines no separate unmerge route. Reversal of an erroneous merge MUST use `POST /api/v1/records/{record_id}/rollback` with `target.kind = 'change_set'` and the merge `change_set_id`.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231


#### 3.3.5.5 Entity-mention action contract

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
This route MUST be the public wire-contract owner for the inspector's explicit mention actions. The legal transition matrix in this subsection MUST also govern single-target `resolve_item`, `dismiss_item`, and `revert_to_unresolved` actions sent through `collection_actions_v1` for `timeline.host_refs` and `timeline.identity_refs`.
Profiles: base
Verified by: AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

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
If present, `reason` MUST be a JSON string or JSON `null`.
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
For idempotency comparison, omission and explicit JSON `null` for `reason` MUST compare equal.
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
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231, AC-239, AC-240, AC-245, AC-246, AC-247, AC-249, AC-250, AC-251, AC-252, AC-253, AC-254, AC-255, AC-260, AC-261

| `error.code` | Required `error.status` | Required `error.retryable` | Canonical meaning | Requirement ID | Profiles | Verified by |
| --- | --- | --- | --- | --- | --- | --- |
| `invalid_view_query` | `400` | `false` | The view-query request is malformed, uses a page-size member or page-size value not allowed by the route, replays a cursor against a different bound query contract, or uses a `field_key`, filter operator, or operand shape not allowed by the active `view_schema_id`. |  |  |  |
| `invalid_pagination_request` | `400` | `false` | A non-view pageable or singleton route uses a page-size member or page-size value not allowed by the route, replays a cursor against a different bound route contract, or supplies pagination members to a route that does not support pagination. |  |  |  |
| `invalid_mutation_payload` | `400` | `false` | A mutation request body is malformed, omits a route-required member, includes an unknown or forbidden member, uses an unknown `kind`, `op`, or `action`, targets a field/action or mention-action combination that is not allowed, or carries an invalid, foreign, or type-incompatible mutation target reference. |  |  |  |
| `invalid_evidence_handle_request` | `400` | `false` | A preview-handle or download-handle issuance request is malformed, uses a non-object JSON body, omits the required JSON object wrapper, or includes an unknown top-level member. |  |  |  |
| `invalid_blob_create_request` | `400` | `false` | A blob-slot create request is malformed, omits required members, violates create-time field validation, attempts to set server-managed state, or includes an unknown top-level member. |  |  |  |
| `invalid_incident_create` | `400` | `false` | An incident-create request is malformed, omits required members, violates create-time field validation, attempts to set server-managed state, or includes a rejected collaborator-seeding payload. |  |  |  |
| `invalid_incident_patch` | `400` | `false` | An incident-metadata patch request is malformed, omits required `base_incident_version`, attempts to mutate an immutable or server-managed incident field, or includes unknown top-level members. |  |  |  |
| `invalid_rollback_request` | `400` | `false` | A rollback request is malformed, uses an unknown or unsupported `target.kind`, omits the selector required for that `kind`, includes unknown request members, or supplies a selector whose JSON type does not match the declared shape. |  |  |  |
| `invalid_auth_request` | `400` | `false` | A local-account login request is malformed, omits a required member, includes an unknown or forbidden member, supplies `null` where forbidden, uses an unsupported `second_factor.kind`, or carries an invalid TOTP assertion shape. |  |  |  |
| `invalid_credentials` | `401` | `false` | The server is not willing to acknowledge that primary credentials were valid on the local-account login route, including for unknown login identifier, wrong password, inactive local account, or equivalent pre-MFA failure. |  |  |  |
| `mfa_required` | `401` | `false` | Primary credentials are valid for a local account that requires MFA, but the login request omitted the required second-factor assertion. On the base local-account login route, `error.details.required_second_factor_kinds` lists the accepted kinds. |  |  |  |
| `invalid_second_factor` | `401` | `false` | Primary credentials are valid and the local-account login request supplied a structurally valid second-factor assertion, but the asserted factor is wrong or expired. |  |  |  |
| `client_txn_conflict` | `409` | `false` | The caller reused a `client_txn_id` within the same route-defined idempotency scope for a different normalized request. |  |  |  |
| `job_cancel_rejected` | `409` | `false` | A visible job exists, but the server will not accept cancellation because cancellation is already requested, the job is already terminal, or the current non-terminal phase is not cancelable; `error.details.reason_code` MUST use the `job_cancel_rejected` registry in §3.3.6.2. |  |  |  |
| `row_version_conflict` | `409` | `false` | The supplied `base_row_version`, `base_mention_row_version`, or equivalent current revision token accepted for restore is stale relative to authoritative current state. |  |  |  |
| `incident_key_conflict` | `409` | `false` | The normalized `incident_key` conflicts with an existing incident. |  |  |  |
| `incident_version_conflict` | `409` | `false` | The supplied `base_incident_version` is stale. |  |  |  |
| `same_field_conflict` | `409` | `false` | Another committed write touched the same writable `field_key`; the response MUST include the conflict object defined by Core 03 §3.3.4. | REQ-01-235 | base | AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231 |
| `illegal_transition` | `409` | `false` | The requested lifecycle transition is not allowed for the current persisted state or guard condition. |  |  |  |
| `record_deleted_use_restore` | `409` | `false` | The caller targeted a currently soft-deleted record with an operation that requires the record to be restored first. |  |  |  |
| `record_already_deleted` | `409` | `false` | The caller attempted to soft-delete an already soft-deleted record outside an idempotent replay of the original delete. |  |  |  |
| `record_not_deleted` | `409` | `false` | The caller attempted to restore a record that is not currently soft-deleted. |  |  |  |
| `record_locked` | `409` | `true` | A short-lived destructive-operation lock prevents the requested restore, rollback, or merge from proceeding at this time. |  |  |  |
| `evidence_access_unavailable` | `409` | `false` | Preview or download cannot currently proceed because the visible evidence or linked blob is unavailable, pending, failed, missing, quarantined, inconsistent, or not previewable for the requested preview contract. |  |  |  |
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
| `last_deployment_admin` | `409` | `false` | The requested user mutation would leave the deployment with no active `is_deployment_admin=true` user. |  |  |  |
| `user_not_found` | `404` | `false` | A membership-create request referenced a user that does not exist in deployment-local identity state. |  |  |  |
| `user_inactive` | `409` | `false` | A membership-create request referenced a deployment-local user whose `is_active=false`. |  |  |  |
| `membership_exists_use_patch` | `409` | `false` | A membership-create request conflicts with an existing membership for the same `(incident_id, user_id)` and must be expressed as a patch instead. |  |  |  |
| `membership_version_conflict` | `409` | `false` | The supplied `base_membership_version` is stale. |  |  |  |
| `membership_not_found` | `404` | `false` | A membership route that requires an existing current membership targeted no current membership row for the identified `(incident_id, user_id)`. |  |  |  |
| `last_incident_admin` | `409` | `false` | The requested membership create, patch, or delete would leave the incident without any current `admin` membership. |  |  |  |
| `merge_precondition_failed` | `409` | `false` | An entity-merge precondition other than row-version freshness failed; `error.details.reason_code` MUST use the merge-precondition registry in §3.3.6.2. | REQ-01-237 | base | AC-126, AC-187, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231 |

##### 3.3.6.2 Canonical public reason-code registries

**REQ-01-238**
When the public API or collaboration stream uses a structured `reason_code` family listed below, it MUST use one of the exact tokens shown. A listed `reason_code` family MUST NOT define alternate tokens for the same meaning elsewhere in the core.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-211, AC-213, AC-214, AC-218, AC-219, AC-231, AC-239, AC-240, AC-252, AC-255, AC-260

`invalid_incident_create` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `request_not_object` | The request body is not a JSON object. |
| `missing_required_field` | A required incident-create field is absent. |
| `field_not_nullable` | The request supplies `null` for a non-nullable incident-create field. |
| `field_empty_after_normalization` | `incident_key` or `title` is empty after required normalization and trimming. |
| `field_too_long` | A create field exceeds its declared maximum length. |
| `control_character_not_allowed` | `incident_key` or `title` contains a rejected control character. |
| `server_managed_field` | The request attempted to set server-managed incident state. |
| `collaborator_seeding_not_supported` | The request supplied `initial_memberships[]` or an equivalent collaborator-seeding payload. |

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

`invalid_view_query` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `unknown_filter_field` | `field_key` is not declared filterable for the active `view_schema_id`. |
| `operator_not_allowed` | `op` is not allowed for that field's declared filter class. |
| `invalid_filter_operand` | `arg` is malformed, empty after normalization, contradictory, or otherwise invalid for the selected `op`. |
| `duplicate_filter_field` | The request contains more than one normalized filter entry for the same `field_key`. |
| `invalid_limit` | The request supplies `limit` with a non-integer JSON type, a value less than `1`, a value greater than `500`, or an unsupported page-size alias such as `page`, `offset`, `block_size`, or `page_size`. |
| `cursor_query_mismatch` | The supplied `cursor_token` does not match the current normalized view-query contract, including the effective `limit`. |

`invalid_pagination_request` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `invalid_limit` | The request supplies `limit` with a non-integer JSON type, a value less than `1`, a value greater than `500`, or an unsupported pagination alias such as `page`, `offset`, `block_size`, or `page_size`. |
| `cursor_query_mismatch` | The supplied `cursor_token` does not match the current normalized route contract, including any bound route-scoping identifier, normalized sort or filter or grouping contract when present, or the effective `limit`. |
| `pagination_not_supported` | The addressed route is not declared pageable and therefore rejects `limit`, `cursor_token`, and pagination aliases. |

`merge_precondition_failed` `error.details.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `same_record` | The supplied survivor and loser identify the same record. |
| `different_incident` | The two records do not belong to the same incident. |
| `record_type_mismatch` | The two records do not have the same `record_type`. |
| `unsupported_record_type` | The requested `record_type` is not mergeable through the base-profile public route. |
| `survivor_not_mergeable` | The nominated survivor is deleted, already merged away, or otherwise not eligible to survive the merge. |
| `loser_not_mergeable` | The nominated loser is deleted, already merged away, or otherwise not eligible to lose the merge. |

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

`/ws/v1/` `session_revoked.payload.reason_code` values:

| `reason_code` | Canonical meaning |
| --- | --- |
| `session_expired` | The authenticated session ended because idle or absolute expiry was reached. |
| `session_revoked` | The current session was explicitly logged out or otherwise deployment-revoked. |
| `incident_access_revoked` | The session remains otherwise valid but no longer authorizes the subscribed incident. |
| `concurrency_limit` | The session was revoked because a newer login exceeded the concurrent-session cap. |

#### 3.3.7 Pagination and cursor contract

**REQ-01-240**
Any public list or query route that MAY exceed one response page MUST use opaque cursor pagination under this subsection. In `/api/v1/`, GET routes MUST accept only `limit` and `cursor_token` as query members and POST query routes MUST accept only `limit` and `cursor_token` as JSON-body members. `limit` MUST be an integer in the inclusive range `1..500`. When `cursor_token` is absent, omitted `limit` MUST mean `100`. When `cursor_token` is present, omitted `limit` MUST mean reuse the cursor-bound effective `limit` from the request that produced that cursor. The `/api/v1/` contract MUST reject `page`, `offset`, `page_size`, `block_size`, or any other pagination alias rather than silently accepting or translating it.
Profiles: base
Verified by: AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-238, AC-239, AC-240

**REQ-01-241**
A `cursor_token` MUST be bound to the authenticated actor, route family, every route-scoping identifier present for that route, the normalized sort or filter or grouping contract when the route defines one, and the effective `limit`. This includes binding history cursors to `record_id`, membership and saved-view cursors to `incident_id`, and workbook-query cursors to `incident_id`, `view_schema_id`, and the normalized view-query contract. The server MUST reject a cursor that is replayed against a different bound route contract, including a different effective `limit`, rather than reinterpret it.
Profiles: base
Verified by: AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-239

**REQ-01-242**
The envelope for paged responses MUST include `meta.paging.limit`, `meta.paging.has_more`, and `meta.paging.next_cursor`. `meta.paging.limit` MUST equal the effective page size bound to the current page. When another page is available, `meta.paging.has_more=true` and `meta.paging.next_cursor` MUST be a non-null opaque cursor. A terminal page, including a first page with zero matching rows, MUST use `meta.paging.has_more=false` and `meta.paging.next_cursor=null`. Non-view routes that are not declared pageable in their owner section MUST reject `limit`, `cursor_token`, and any pagination alias with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code=pagination_not_supported` rather than ignoring them. Clients MUST treat `meta.paging.has_more` and `meta.paging.next_cursor` as the authoritative continuation contract and MUST NOT infer terminal state from `rows.length < meta.paging.limit` alone. Hot workbook views MUST NOT require deep `OFFSET` pagination.
Profiles: base
Verified by: AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-238, AC-239, AC-241, AC-242

#### 3.3.8 Evidence and blob routes

**REQ-01-243**
`POST /api/v1/object-blobs` MUST accept only a JSON object request body and MUST provision exactly one incident-scoped pending blob slot for one intended upload. The request MUST accept required `incident_id`, required `client_txn_id`, and required `byte_size`. It MAY accept optional `filename_hint`, optional `content_type_hint`, and optional `sha256_hex`; each optional member MAY be omitted or set to JSON `null`. `byte_size` MUST be a non-negative integer. This route MUST create only the blob slot; it MUST NOT create or mutate evidence records, record links, preview state, release state, or workflow objects. The route MUST reject row identifiers, evidence identifiers, preview intents, release intents, workflow objects, unknown top-level members, and server-managed blob fields. If the body is not a JSON object, omits a required member, supplies `null` for a non-nullable member, violates a field validation rule in this subsection, or attempts to set server-managed state, the server MUST fail with `400` and `error.code = invalid_blob_create_request`. When the failure is attributable to one request member, `error.details.field` MUST identify that top-level member. `error.details.reason_code` MUST use the registry in §3.3.6.2.
Profiles: base
Verified by: AC-015, AC-016, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

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
Blob-slot creation idempotency MUST be keyed by `(actor_user_id, incident_id, client_txn_id)`. Normalized request comparison for this route MUST include `byte_size`, normalized `filename_hint`, normalized `content_type_hint`, and normalized `sha256_hex`, with omission and explicit JSON `null` treated as equal for the optional members. A first-time successful create MUST return `201 Created`. If the same authenticated actor replays the same normalized request with the same key, the server MUST return `200 OK` with the original slot payload, including the original `object_blob_id`, `target_expires_at`, `pending_expires_at`, and `accepted_contract`, and MUST create no second pending slot. If the same actor reuses that key with a different normalized request, the server MUST fail with `409` and `error.code = client_txn_conflict`. `error.details` MUST include at least `client_txn_id`. Blob finalization MUST occur only through an explicit follow-on call. Binding an uploaded blob to incident-visible evidence MUST either:

- attach the returned `object_blob_id` to an existing evidence record through `POST /api/v1/evidence-records/{record_id}/attach-blob`, or
- create a new evidence record through the normal view or record-creation path using that `object_blob_id` as declared input.

When finalization targets an existing evidence record through `POST /api/v1/evidence-records/{record_id}/attach-blob`, the route MUST be record-scoped rather than blob-scoped. It MUST accept only a JSON object with required `object_blob_id`, required `base_row_version`, and required `client_txn_id`. The base profile defines no optional top-level members for this route. A non-object body, a missing required member, a supplied `null` for one of those required members, or any unknown top-level member MUST fail with `400` and `error.code = invalid_mutation_payload`.

Attach idempotency for this route MUST be keyed by `(actor_user_id, record_id, client_txn_id)`. Normalized request comparison for this route MUST include exact `object_blob_id` and exact `base_row_version`. Because the base profile defines no nullable request members for this route, omission-versus-`null` equivalence does not apply. A first-time successful attach MUST return `200 OK`. If the same authenticated actor replays the same normalized attach request with the same key, the server MUST return `200 OK` with the original committed attach result and MUST create no second attach or replacement transition. If the same actor reuses that key with a different normalized attach request, the server MUST fail with `409` and `error.code = client_txn_conflict`. If the current evidence-row version differs from `base_row_version`, the route MUST fail with `409` and `error.code = row_version_conflict`.

Fresh attach requests with a new `client_txn_id` MUST still enforce the blob and evidence lifecycle bridge owned by Core 02 §13 and Core 03 §8. A pending, failed, missing, incident-mismatched, or otherwise non-attachable blob MUST fail closed rather than partially mutating evidence state. A successful attach response MUST use the common success envelope and return the authoritative evidence-refresh payload needed to repaint the visible row, including at minimum `record_id`, `incident_id`, `row_version`, `object_blob_id`, `evidence_lifecycle_state`, `upload_state`, and `change_set_id`.
Profiles: base
Verified by: AC-015, AC-016, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-01-246**
In the base profile, the upload target MUST expire 60 minutes after issuance and the pending blob slot MUST expire 24 hours after blob-slot creation. These timers MUST remain separate: target expiry governs upload-target use, while pending-slot expiry governs later finalization eligibility and cleanup. The base profile MUST treat a pending blob slot as a single-upload lease bound to one accepted create contract. If the upload target expires before successful upload, or if idempotent replay returns an already expired slot, obtaining a fresh target MUST use a fresh `POST /api/v1/object-blobs` call with a new `client_txn_id`. Idempotent replay of an expired slot MUST return that same expired slot rather than refreshing the original target. The base profile MUST NOT require same-slot upload-target refresh, same-slot lease renewal, or resumable upload semantics.
Profiles: base
Verified by: AC-015, AC-016, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-01-247**
Preview and download routes MUST return short-lived authorization-checked handles or an equivalent mediated access contract. They MUST NOT expose long-lived object-store credentials, bypass incident membership checks, or treat a `pending` blob slot as attached evidence.
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
- `result_summary` and `error_summary` are mutually exclusive. On non-terminal states, both MUST be `null`. On `succeeded` and `canceled`, `result_summary` is required and `error_summary` MUST be `null`. On `failed`, `error_summary` is required and `result_summary` MUST be `null`.
- `result_summary` MUST be compact and generic: `{ code, message, resource_refs? }`, where `resource_refs[]` is an optional array of `{ kind, id, route? }`. `error_summary` MUST be compact and generic: `{ code, message, retryable, details? }`, where `details` is an optional JSON object. The common job resource MUST NOT carry job-family-specific deep result payloads.
- `POST /api/v1/jobs/{job_id}/cancel` MUST require a JSON object containing required `client_txn_id` and optional `reason`. For idempotency comparison, omitted `reason` and explicit JSON `null` for `reason` compare equal. A cancel request body that is not a JSON object, omits required `client_txn_id`, or includes unknown top-level members MUST fail with `400` and `error.code = invalid_mutation_payload`.
- Cancel idempotency is keyed by `(actor_user_id, job_id, client_txn_id)`. If the same authenticated actor replays the same normalized cancel request with the same key, the server MUST return `200 OK` with the current authoritative job resource and MUST create no second cancel transition. If the same actor reuses that key with a different normalized cancel request, the server MUST fail with `409` and `error.code = client_txn_conflict`.
- A newly accepted cancel request MUST set `cancelable = false` immediately and MUST transition the job through `cancel_requested` before any later `canceled`, `failed`, or `succeeded` terminal state is observed. A conformant server MAY return a terminal state on the successful cancel response when the cancel lost a race to an already committed safe-stop boundary, but it MUST NOT create a second public transition.
- A fresh cancel request against `cancel_requested`, `succeeded`, `failed`, `canceled`, or any non-terminal resource whose current `cancelable = false` MUST fail with `409`, `error.code = job_cancel_rejected`, and `error.details.reason_code` equal to `already_cancel_requested`, `already_terminal`, or `not_cancelable` from §3.3.6.2. A missing, expired, or unauthorized job lookup or cancel request MUST fail with `404` and `error.code = job_not_found`.
- Terminal job resources MUST be retained for at least 7 days. After a terminal transition, `retained_until` is required and MUST be greater than or equal to `finished_at + 7 days`. After `retained_until`, `GET /api/v1/jobs/{job_id}` MAY return `404` with `error.code = job_not_found`. Expiring a terminal job resource MUST NOT delete or mutate durable outputs such as committed incident changes, imports, reports, snapshots, bundles, or blob metadata produced by that job.

#### 3.3.10 WebSocket collaboration stream

**REQ-01-250**
The incident-scoped WebSocket route MUST be `/ws/v1/incidents/{incident_id}` or an equivalent versioned incident-scoped path.
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
When `sheet_ref.kind = view_schema`, `sheet_ref.id` MUST carry the `view_schema_id`. When `sheet_ref.kind = saved_view`, `sheet_ref.id` MUST carry the `saved_view_id`. `field_key` MUST be present only when the client is focused on a concrete writable field and `mode = editing`.
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
`presence_snapshot.payload.presences[]` MUST contain zero or more current presence records for the subscribed incident. Each presence record MUST include `connection_id`, `user_id`, `display_name`, `sheet_ref`, `mode`, `observed_at`, and `expires_at`, and MAY include `record_id` and `field_key`. At most one current presence record MAY exist per `connection_id` within one incident.
Profiles: base, snapshot_reporting
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-233

**REQ-01-266**
`presence_delta.payload` MUST include `delta_kind` and `presence`. `delta_kind` MUST be one of `upsert` or `remove`. On `upsert`, `presence` MUST be a full presence record. On `remove`, `presence.connection_id` MUST identify the removed record; other presence members MAY be repeated for convenience but MUST NOT be required.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

**REQ-01-267**
`record_changed.payload` MUST include `record_id`, `row_version`, `change_set_id`, `actor_user_id`, `changed_field_keys[]`, and `affected_views[]`. Each `affected_views[]` entry MUST include `view_schema_id` and `change_kind`. `change_kind` MUST be one of `patch`, `invalidate`, or `remove`. When `change_kind = patch`, the entry MUST include `patch_cells`, and `patch_cells` MUST be a view-shaped object containing `record_id`, `row_version`, and field-key-addressable cell values for that view. `affected_views[]` MUST be keyed by base `view_schema_id` values, not by visible tab labels, row order, or client-local filter state. For lifecycle actions such as record soft-delete or record restore, `changed_field_keys[]` MUST be present and MAY be empty. A restored row that may re-enter a view MUST surface as `invalidate`; `/ws/v1/` MUST NOT introduce a distinct insert-like `change_kind` for that case within the v1 contract.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231

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
Verified by: AC-231, AC-233, AC-234

### 4.2 Object storage

**REQ-01-279**
S3-compatible object storage MUST store:

- binary evidence payloads,
- optionally, generated export artifacts when the Snapshot and Reporting Extension Profile is implemented.
Profiles: base, snapshot_reporting
Verified by: AC-231, AC-233

The object store implementation MAY be MinIO in flyaway or on-prem deployments. Cloud deployments MAY use native S3, GCS, or Azure Blob behind an equivalent abstraction.

### 4.3 Storage exclusions

**REQ-01-280**
Large binary evidence MUST NOT be stored inline in Postgres.
Profiles: base
Verified by: AC-231

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
- an ordered field registry containing one entry per visible or writable field, each with a stable `field_key`,
- computed columns,
- the required hidden technical fields `record_id` and `row_version`,
- required reference packs, if any,
- an ordered default sort tuple,
- filter semantics and any allowed grouping keys,
- per-field write-back semantics, including write target or write action, `conflict_resolution_class`, and `entity_binding_mode` where relevant,
- metadata needed to render the view consistently.
Profiles: base, reference_pack
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-231, AC-234

**REQ-01-287**
Default sort tuples MUST be deterministic and MUST include `record_id` as the final tiebreaker unless a later profile explicitly overrides that rule.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-231

**REQ-01-288**
A base-profile implementation MUST expose the structured base-profile view-schema registry for conformance inspection through stored `view_schema` rows, the public discovery routes named in REQ-01-032, or an equivalent structured export. `GET /api/v1/view-schemas` MUST return the common success envelope with `data.view_schemas[]` plus `meta.paging`, MUST order results by `view_schema_id asc`, and MUST accept only `limit` and `cursor_token` under §3.3.7. `GET /api/v1/view-schemas/{view_schema_id}` MUST return one structured view-schema resource and MUST reject pagination members with `400`, `error.code=invalid_pagination_request`, and `error.details.reason_code=pagination_not_supported`. Conformance MUST NOT depend on scraping visible tab labels, column labels, or interactive UI behavior alone.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-127, AC-231

**REQ-01-289**
View behavior MUST bind to `view_schema_id`, not to the visible tab label, column header text, or any other display label.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-124, AC-125, AC-231

**REQ-01-290**
If a visible column header or tab label changes, filter behavior, write-back behavior, and export semantics MUST remain unchanged unless the underlying `view_schema_id` changes.
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
- decisions.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

Framework overlays such as ATT&CK, D3FEND, or VERIS MAY also be exposed as system views when the relevant reference packs are present.

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
Structured coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MAY be surfaced through contract-backed system views or saved views over `artifact_grid_projection`. They MUST NOT require additional built-in sheets in the base profile.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

### 7.3 Notes sheet contract

**REQ-01-303**
The base profile MUST expose **Notes** as a built-in workbook sheet.
Profiles: base
Verified by: AC-068, AC-069, AC-070, AC-112, AC-185, AC-231

**REQ-01-304**
The built-in Notes sheet MUST:

- be declared by a stable `view_schema_id`,
- use `artifact_grid_projection` as its base projection filtered to `artifact_type='note'`,
- support blank-row or equivalent grid-native note creation from the sheet itself,
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
The base profile MUST define the following nine pack-independent `view_schema` entries as the authoritative base-profile registry:

- `cartulary.view.timeline.v1`,
- `cartulary.view.hosts.v1`,
- `cartulary.view.identities.v1`,
- `cartulary.view.evidence.v1`,
- `cartulary.view.notes.v1`,
- `cartulary.view.indicators.v1`,
- `cartulary.view.assessments.v1`,
- `cartulary.view.task_requests.v1`,
- `cartulary.view.decisions.v1`.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-231

**REQ-01-308**
These nine entries are the contract-backed surfaces required by §7.1 and §7.2. Additional `view_schema` entries MAY exist for saved views, coordination-artifact surfaces, optional reference-pack overlays, or later profiles, but they MUST NOT change the membership or semantics of the base-profile registry defined in this subsection.
Profiles: base, reference_pack
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-231, AC-234

**REQ-01-309**
Each schema subsection below is an exhaustive per-field registry for that `view_schema_id`, not an illustrative example. In particular, §7.4.2 through §7.4.4 close the base-profile interface contract for the built-in Hosts, Identities, and Evidence sheets. Implementations MUST NOT invent alternate base-profile writable `field_key` strings, write targets or actions, `conflict_resolution_class` assignments, or `entity_binding_mode` values for those schemas.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-231

**REQ-01-310**
Unless explicitly overridden below:

- `required_reference_pack_keys` MUST be `[]`,
- `record_id` and `row_version` MUST be present as hidden technical fields,
- the ordered default sort tuple is normative and MUST end with `record_id asc`,
- filter semantics MUST be type-driven unless a schema below explicitly overrides them: enum and boolean fields use exact-match inclusion, timestamp and date fields use exact or range predicates, scalar identifier text uses case-insensitive exact or prefix matching, multi-value collections use `contains_any` and `contains_all`, and declared full-text predicates use case-insensitive token search,
- every writable field entry MUST declare `field_key`, read model, write target or write action, and `conflict_resolution_class`,
- every entity-bearing writable field entry MUST declare `entity_binding_mode`,
- fields not declared writable are read-only,
- per-user hide/show or reordering MAY change presentation but MUST NOT change field identity, filter semantics, or write-back semantics.
Profiles: base, reference_pack
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-231, AC-234

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
  - `decision.support_refs` -> `supported_by`, with the decision record as `src_record_id` and the supporting record as `dst_record_id`,
  - explicit decision supersession actions -> `supersedes`, with the superseding decision as `src_record_id` and the superseded decision as `dst_record_id`.
Profiles: base
Verified by: AC-116, AC-117, AC-118, AC-119, AC-120, AC-121, AC-122, AC-124, AC-125, AC-231

#### 7.4.1 `cartulary.view.timeline.v1`

**REQ-01-312**
- surface: built-in `Timeline` sheet
- source record types: `timeline_event`
- base projection: `timeline_grid_projection`
- `default_visible_fields`: `timeline.occurred_at`, `timeline.summary`, `timeline.host_refs`, `timeline.identity_refs`, `timeline.evidence_count`, `timeline.tags`, `timeline.edited_at`
- `default_hidden_fields`: `record_id`, `row_version`, `timeline.details`, `timeline.source_text`, `timeline.recorded_at`, `timeline.sort_ts`, `timeline.capture_state`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`
- `default_sort`: `timeline.sort_ts asc`, `record_id asc`. `timeline.sort_ts` MUST equal `occurred_at` when present and `recorded_at` otherwise.
- `filter_fields`: `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`, `timeline.tags`
- `grouping_fields`: `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`
- inline create: blank-row or equivalent grid-native creation MUST create a `timeline_event` record
- writable fields:
  - `timeline.occurred_at`: read `occurred_at`; write target `timeline_events.occurred_at`; `conflict_resolution_class=atomic_replace`
  - `timeline.summary`: read `summary`; write target `timeline_events.summary`; `conflict_resolution_class=text_compare_merge`
  - `timeline.details`: read `details`; write target `timeline_events.details`; `conflict_resolution_class=text_compare_merge`
  - `timeline.source_text`: read `source_text`; write target `timeline_events.source_text`; `conflict_resolution_class=text_compare_merge`
  - `timeline.host_refs`: read resolved host chips plus unresolved host mentions; write action insert, update, or dismiss `entity_mentions` and resolved host `record_links` under `entity_binding_mode=mention_origin`; `conflict_resolution_class=collection_review`
  - `timeline.identity_refs`: read resolved identity chips plus unresolved identity mentions; write action insert, update, or dismiss `entity_mentions` and resolved identity `record_links` under `entity_binding_mode=mention_origin`; `conflict_resolution_class=collection_review`
  - `timeline.tags`: read `tag_names`; write action upsert tags and `record_tags`; `conflict_resolution_class=collection_review`
- read-only projection-backed or system-managed fields: `timeline.evidence_count`, `timeline.capture_state`, `timeline.edited_at`, `timeline.sort_ts`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`
- `timeline.capture_state` is the system-managed persisted workflow state defined by Core 03 §6. Clients MUST NOT supply `timeline.capture_state` as an initial writable value in `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` or as a `changes[].field_key` in `PATCH /api/v1/records/{record_id}`. Any attempted direct client write to `timeline.capture_state` through create or patch MUST fail closed under §3.3.5 and §3.3.6 rather than being silently ignored. Transitions to `reviewed` and `superseded` MUST use the dedicated record-scoped action routes defined in §3.3.5, and automatic transitions to `enriched` MUST be applied server-side with the committed Timeline mutation that triggered them.
- `timeline.has_unresolved_mentions` MUST be `true` if and only if at least one non-deleted `entity_mentions` row for the source record has `resolution_status='unresolved'`; resolved or dismissed mentions MUST NOT make it `true`.
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

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
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

```json
{ "op": "add_token", "raw_text": "WS-023?" }
```

**REQ-01-316**
`add_resolved_ref` MUST use:
Profiles: base
Verified by: AC-119, AC-124, AC-125, AC-184, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-231

```json
{
  "op": "add_resolved_ref",
  "resolved_record_id": "<record_id>",
  "raw_text": "WS-023"
}
```

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
- The server MUST normalize `tag_name` by trimming leading and trailing Unicode whitespace and applying the same case-insensitive uniqueness rules as incident-scoped `tags.name`. Empty tag names after normalization MUST be rejected with `400 Bad Request` and `error.code = invalid_mutation_payload`. Duplicate adds MUST coalesce to one surviving active binding.
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
- `filter_fields`: `host.host_state`, `host.business_owner`, `host.criticality`, `host.location`, `host.os_platform`, `host.containment_status`
- inline create: direct row creation or paste on the Hosts sheet MUST create or upsert a `host` record using `entity_binding_mode=entity_origin`
- create-or-upsert reuse on the Hosts sheet MUST apply the exact-match precedence in Core 02 §8.2. Suggestion-only candidates allowed by Core 02 §8.3 MUST NOT silently auto-merge or auto-resolve a host.
- writable fields:
  - `host.display_name`: read the canonical host display field; write target the canonical host display field on the underlying `host` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `host.hostname`: read `hostname`; write target the canonical hostname field on the underlying `host` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `host.aliases`: read projected alias chips; write action upsert or remove `entity_aliases`; `entity_binding_mode=entity_origin`; `conflict_resolution_class=collection_review`
  - `host.location`: read `location`; write target the `location` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
  - `host.os_platform`: read `os_platform`; write target the `os_platform` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
  - `host.business_owner`: read `business_owner`; write target the `business_owner` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
  - `host.criticality`: read `criticality`; write target the `criticality` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
  - `host.containment_status`: read `containment_status`; write target the `containment_status` field on the underlying `host` record; `conflict_resolution_class=atomic_replace`
- read-only computed fields: `host.host_state`, `host.linked_event_count`, `host.evidence_count`, `host.edited_at`. `host.host_state` MUST be a projection-backed state equivalent to `stub` or `canonical`.
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
- The server MUST derive any base-profile alias classification and storage routing from `field_key`; the client MUST NOT send table names, `alias_type`, or storage-specific routing metadata.
- Alias rename in the base profile MUST be expressed as `remove_alias` plus `add_alias`. The public API surface MUST NOT require in-place alias-row update semantics.
- Duplicate alias adds for the same canonical record and normalized `alias_text` MUST coalesce to one surviving alias row.
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
- `filter_fields`: `identity.identity_state`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status`
- inline create: direct row creation or paste on the Identities sheet MUST create or upsert an `identity` record using `entity_binding_mode=entity_origin`
- create-or-upsert reuse on the Identities sheet MUST apply the exact-match precedence in Core 02 §8.2. Suggestion-only candidates allowed by Core 02 §8.3 MUST NOT silently auto-merge or auto-resolve an identity.
- writable fields:
  - `identity.display_name`: read the canonical identity display field; write target the canonical identity display field on the underlying `identity` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `identity.upn`: read `upn`; write target the canonical UPN field on the underlying `identity` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `identity.email`: read `email`; write target the canonical email field on the underlying `identity` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `identity.sam_account_name`: read `sam_account_name`; write target the canonical `sam_account_name` field on the underlying `identity` record; `entity_binding_mode=entity_origin`; `conflict_resolution_class=atomic_replace`
  - `identity.aliases`: read projected alias chips; write action upsert or remove `entity_aliases`; `entity_binding_mode=entity_origin`; `conflict_resolution_class=collection_review`
  - `identity.privilege_level`: read `privilege_level`; write target the `privilege_level` field on the underlying `identity` record; `conflict_resolution_class=atomic_replace`
  - `identity.mfa_state`: read `mfa_state`; write target the `mfa_state` field on the underlying `identity` record; `conflict_resolution_class=atomic_replace`
  - `identity.reset_status`: read `reset_status`; write target the `reset_status` field on the underlying `identity` record; `conflict_resolution_class=atomic_replace`
- read-only computed fields: `identity.identity_state`, `identity.linked_event_count`, `identity.evidence_count`, `identity.edited_at`. `identity.identity_state` MUST be a projection-backed state equivalent to `stub` or `canonical`.
Profiles: base
Verified by: AC-098, AC-118, AC-124, AC-125, AC-231

**REQ-01-327**
`identity.aliases` MUST use the same `collection_value_v1` item shape and `collection_actions_v1` action vocabulary as `host.aliases`, except the active `field_key` is `identity.aliases`.
Profiles: base
Verified by: AC-098, AC-118, AC-124, AC-125, AC-231

#### 7.4.4 `cartulary.view.evidence.v1`

**REQ-01-328**
- surface: built-in `Evidence` sheet
- source record types: `evidence`
- base projection: `evidence_grid_projection`
- `default_visible_fields`: `evidence.title`, `evidence.lifecycle_state`, `evidence.requested_at`, `evidence.received_at`, `evidence.storage_ref`, `evidence.blob_hash`, `evidence.collector_party_text`, `evidence.source_party_text`, `evidence.upload_state`, `evidence.linked_record_count`, `evidence.edited_at`
- `default_hidden_fields`: `record_id`, `row_version`
- `default_sort`: `evidence.requested_at desc`, `record_id asc`
- `filter_fields`: `evidence.lifecycle_state`, `evidence.upload_state`, `evidence.requested_at`, `evidence.received_at`, `evidence.collector_party_text`, `evidence.source_party_text`, `evidence.storage_ref`, `evidence.blob_hash`
- inline create: blank-row or equivalent grid-native creation MUST create an evidence record even when no blob exists yet
- writable fields:
  - `evidence.title`: read the evidence title field; write target the evidence record title field; `conflict_resolution_class=text_compare_merge`
  - `evidence.lifecycle_state`: read `lifecycle_state`; write target `evidence_records.lifecycle_state`; `conflict_resolution_class=atomic_replace`. State-dependent field guards and blob-bridge invariants from Core 02 §13 and Core 03 §8.3 apply.
  - `evidence.requested_at`: read `requested_at`; write target `evidence_records.requested_at`; `conflict_resolution_class=atomic_replace`
  - `evidence.received_at`: read `received_at`; write target `evidence_records.received_at`; `conflict_resolution_class=atomic_replace`
  - `evidence.storage_ref`: read `storage_ref`; write target `evidence_records.storage_ref`; `conflict_resolution_class=atomic_replace`
  - `evidence.collector_party_text`: read `collector_party_text`; write target `evidence_records.collector_party_text`; `conflict_resolution_class=text_compare_merge`
  - `evidence.source_party_text`: read `source_party_text`; write target `evidence_records.source_party_text`; `conflict_resolution_class=text_compare_merge`
- read-only computed fields: `evidence.blob_hash`, `evidence.upload_state`, `evidence.linked_record_count`, `evidence.edited_at`
- blob attach or replacement MUST remain an explicit evidence action. It MUST NOT be modeled as a direct write to `evidence.blob_hash` or `evidence.upload_state`.
Profiles: base
Verified by: AC-100, AC-118, AC-124, AC-125, AC-128, AC-231

#### 7.4.5 `cartulary.view.notes.v1`

**REQ-01-329**
- surface: built-in `Notes` sheet
- source record types: `artifact`
- base projection: `artifact_grid_projection` filtered to `artifact_type='note'`
- `default_visible_fields`: `note.title`, `note.body`, `note.tags`, `note.linked_record_count`, `note.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `note.created_by_user_id`
- `default_sort`: `note.updated_at desc`, `record_id asc`
- `filter_fields`: `note.tags`, `note.created_by_user_id`, `note.updated_at`, `note.full_text`
- `note.full_text` is a filter-only synthetic predicate key declared by this view schema. It applies case-insensitive token search over `note.title` and `note.body`. It is not a writable field and need not be a visible column.
- inline create: blank-row or equivalent grid-native creation MUST create an `artifact` record with `artifact_type='note'`
- writable fields:
  - `note.title`: read the note title field; write target the note artifact title field; `conflict_resolution_class=text_compare_merge`
  - `note.body`: read the note body field; write target the note artifact body field; `conflict_resolution_class=text_compare_merge`
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
- `filter_fields`: `indicator.indicator_type`, `indicator.value_kind`, `indicator.hash_algorithm`, `indicator.first_observed_at`, `indicator.last_observed_at`, `indicator.lifecycle_summary`
- inline create: blank-row or equivalent grid-native creation MUST create a canonical `indicator` record
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
- `indicator.stix_pattern` remains create-only in this view but MUST NOT be treated as identity-defining
- `indicator.defanged_value` remains create-only in this view but MUST NOT be treated as identity-defining
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
- `filter_fields`: `assessment.subject_type`, `assessment.assessment_state`, `assessment.confidence_band`, `assessment.assessed_at`
- inline create: blank-row or equivalent grid-native creation MUST append a new `assessment` record
- writable fields on create only:
  - `assessment.subject_ref`: read the subject record reference; write target the assessment subject reference; `conflict_resolution_class=atomic_replace`
  - `assessment.subject_type`: read the subject type; write target the assessment subject type; `conflict_resolution_class=atomic_replace`
  - `assessment.assessment_state`: read `assessment_state`; write target the `assessment_state` field on the underlying `assessment` record; `conflict_resolution_class=atomic_replace`
  - `assessment.confidence_score`: read `confidence_score`; write target the `confidence_score` field on the underlying `assessment` record; `conflict_resolution_class=atomic_replace`. A band-first editor MUST map `unset`, `low`, `medium`, and `high` to `NULL`, `25`, `55`, and `85` respectively.
  - `assessment.rationale`: read `rationale`; write target the `rationale` field on the underlying `assessment` record; `conflict_resolution_class=text_compare_merge`
  - `assessment.assessor`: read the assessor field; write target the assessor field on the underlying `assessment` record; `conflict_resolution_class=atomic_replace`
  - `assessment.assessed_at`: read `assessed_at`; write target the `assessed_at` field on the underlying `assessment` record; `conflict_resolution_class=atomic_replace`
  - `assessment.support_refs`: read supporting record references; write action upsert supporting `record_links` or denormalized support references; `conflict_resolution_class=collection_review`
- read-only computed fields: `assessment.confidence_band`, `assessment.supporting_link_count`
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-118, AC-121, AC-124, AC-231

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

**REQ-01-334**
- The server MUST derive any base-profile `link_type` and storage routing from `field_key`; the client MUST NOT send `link_type`, table names, or storage-specific routing metadata.
- Duplicate adds for the same patched `record_id`, `linked_record_id`, and field-derived link type MUST coalesce to one surviving logical reference binding.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-118, AC-121, AC-124, AC-231

**REQ-01-335**
- grid edits to an existing assessment row MUST reject writes to `assessment.subject_ref`, `assessment.subject_type`, `assessment.assessment_state`, `assessment.confidence_score`, `assessment.rationale`, `assessment.assessor`, `assessment.assessed_at`, and `assessment.support_refs`.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-118, AC-121, AC-124, AC-231

#### 7.4.8 `cartulary.view.task_requests.v1`

**REQ-01-336**
- surface: contract-backed `Task Requests` system view
- source record types: `task_request`
- base projection: `task_request_grid_projection`
- `default_visible_fields`: `task.title`, `task.status`, `task.owner_user_id`, `task.priority`, `task.task_kind`, `task.workstream`, `task.due_at`, `task.requester_party_text`, `task.blocked_reason`, `task.completed_at`, `task.external_ticket_ref`, `task.linked_record_count`, `task.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `task.closure_summary`, `task.linked_record_ids`, `task.decision_record_id`, `task.no_owner`
- `default_sort`: `task.updated_at desc`, `record_id asc`
- `filter_fields`: `task.status`, `task.owner_user_id`, `task.priority`, `task.task_kind`, `task.workstream`, `task.due_at`, `task.requester_party_text`, `task.blocked_reason`, `task.completed_at`, `task.external_ticket_ref`, `task.no_owner`
- inline create: blank-row or equivalent grid-native creation MUST create a `task_request` record; unless the create payload explicitly chooses another allowed initial state, interactive blank-row creation SHOULD default `task.status` to `open`
- writable fields:
  - `task.title`: read `title`; write target the `title` field on the underlying `task_request` record; `conflict_resolution_class=text_compare_merge`
  - `task.status`: read `status`; write target the `status` field on the underlying `task_request` record; legal writes MUST be validated and any required lifecycle normalization MUST be applied under Core 02 §10.4.1.1 before commit; `conflict_resolution_class=atomic_replace`
  - `task.owner_user_id`: read `owner_user_id`; write target the `owner_user_id` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.priority`: read `priority`; write target the `priority` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.task_kind`: read `task_kind`; write target the `task_kind` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.workstream`: read `workstream`; write target the `workstream` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.due_at`: read `due_at`; write target the `due_at` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.requester_party_text`: read `requester_party_text`; write target the `requester_party_text` field on the underlying `task_request` record; `conflict_resolution_class=text_compare_merge`
  - `task.blocked_reason`: read `blocked_reason`; write target the `blocked_reason` field on the underlying `task_request` record; `conflict_resolution_class=text_compare_merge`
  - `task.completed_at`: read `completed_at`; write target the `completed_at` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.external_ticket_ref`: read `external_ticket_ref`; write target the `external_ticket_ref` field on the underlying `task_request` record; `conflict_resolution_class=atomic_replace`
  - `task.closure_summary`: read `closure_summary`; write target the `closure_summary` field on the underlying `task_request` record; `conflict_resolution_class=text_compare_merge`
  - `task.linked_record_ids`: read linked record references; write action upsert or remove linked `record_links`; `conflict_resolution_class=collection_review`
  - `task.decision_record_id`: read `decision_record_id`; write target the `decision_record_id` field on the underlying `task_request` record or an equivalent linked decision reference; `conflict_resolution_class=atomic_replace`
- read-only computed fields: `task.linked_record_count`, `task.updated_at`, `task.no_owner`
Profiles: base
Verified by: AC-085, AC-118, AC-124, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-01-337**
`task.linked_record_ids` MUST use the same `collection_value_v1` item shape and `collection_actions_v1` action vocabulary as `assessment.support_refs`, except the active `field_key` is `task.linked_record_ids` and the server derives the applicable `link_type` from that field key.
Profiles: base
Verified by: AC-085, AC-118, AC-124, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-01-338**
- task lifecycle semantics remain authoritative: any committed write set affecting `task.status`, `task.blocked_reason`, `task.completed_at`, or `task.owner_user_id` MUST produce a resulting row that satisfies Core 02 §10.4.1.1. In particular, `status='blocked'` requires `blocked_reason`, `status='done'` requires `completed_at`, active tasks MUST NOT be ownerless, a successful transition away from `blocked` or `done` MUST clear `blocked_reason` or `completed_at` respectively, and a successful write that sets `status='done'` with no explicit `completed_at` MUST fill `completed_at` from the commit timestamp.
Profiles: base
Verified by: AC-085, AC-118, AC-124, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

#### 7.4.9 `cartulary.view.decisions.v1`

**REQ-01-339**
- surface: contract-backed `Decisions` system view
- source record types: `decision`
- base projection: `decision_grid_projection`
- `default_visible_fields`: `decision.summary`, `decision.status`, `decision.owner_user_id`, `decision.decision_type`, `decision.decided_at`, `decision.rationale`, `decision.support_refs`, `decision.affected_record_count`, `decision.supersedes_record_id`, `decision.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `decision.is_superseded`
- `default_sort`: `decision.decided_at desc`, `record_id asc`
- `filter_fields`: `decision.status`, `decision.owner_user_id`, `decision.decision_type`, `decision.decided_at`, `decision.is_superseded`
- inline create: blank-row or equivalent grid-native creation MUST create a `decision` record
- writable fields:
  - `decision.summary`: read `summary`; write target the `summary` field on the underlying `decision` record; `conflict_resolution_class=text_compare_merge`
  - `decision.status`: read `status`; write target the `status` field on the underlying `decision` record; legal direct writes MUST be validated against Core 02 §10.4.2.1, and a direct write whose requested `status` is `superseded` MUST be rejected; `conflict_resolution_class=atomic_replace`
  - `decision.owner_user_id`: read `owner_user_id`; write target the `owner_user_id` field on the underlying `decision` record; `conflict_resolution_class=atomic_replace`
  - `decision.decision_type`: read `decision_type`; write target the `decision_type` field on the underlying `decision` record; `conflict_resolution_class=atomic_replace`
  - `decision.decided_at`: read `decided_at`; write target the `decided_at` field on the underlying `decision` record; `conflict_resolution_class=atomic_replace`
  - `decision.rationale`: read `rationale`; write target the `rationale` field on the underlying `decision` record; `conflict_resolution_class=text_compare_merge`
  - `decision.support_refs`: read `support_refs`; write action upsert or remove supporting `record_links` or denormalized support references; `conflict_resolution_class=collection_review`
- read-only computed fields: `decision.affected_record_count`, `decision.supersedes_record_id`, `decision.updated_at`, `decision.is_superseded`
Profiles: base
Verified by: AC-086, AC-118, AC-124, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-01-340**
`decision.support_refs` MUST use the same `collection_value_v1` item shape and `collection_actions_v1` action vocabulary as `assessment.support_refs`, except the active `field_key` is `decision.support_refs` and the server derives the applicable `link_type` from that field key.
Profiles: base
Verified by: AC-086, AC-118, AC-124, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-01-341**
- supersession remains an explicit decision-linking flow. It MUST NOT be modeled as a direct write to `decision.supersedes_record_id` or as a direct write that sets `decision.status='superseded'` in the base-profile grid.
- when the explicit supersession flow succeeds, it MUST persist the supersession relation and apply the status effects defined by Core 02 §10.4.2.1 atomically.
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
- `decision_grid_projection` for the decision system view.
Profiles: base
Verified by: AC-032, AC-046, AC-210, AC-231

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
Verified by: AC-231

**REQ-01-350**
The client MUST NOT infer mutation targets from row position, displayed values, group headers, or transient selection state.
Profiles: base
Verified by: AC-231

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
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-357**
Projection tables MUST support deterministic interactive retrieval over the ordered default sort tuple and any contract-declared interactive grouping key for the view. An implementation MAY satisfy this with indexes or an equivalent mechanism that preserves the same observable latency envelope.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

**REQ-01-358**
Projection rows for hot workbook sheets MUST carry the scalar fields required for interactive sort, filter, grouping, selection anchoring, and evidence badges in the visible viewport. For the Timeline sheet, this MUST include at least `sort_ts`, day buckets equivalent to `timeline.occurred_day` and `timeline.recorded_day`, `capture_state`, `has_evidence`, `has_unresolved_mentions`, and `evidence_count`.
Profiles: base
Verified by: AC-015, AC-016, AC-017, AC-045, AC-053, AC-054, AC-100, AC-128, AC-210, AC-231

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
- `source_change_set_high_watermark` or equivalent frozen source boundary,
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
- `curated_narrative` for analyst-authored findings prose, executive summaries, recommendations, impact statements, and lessons learned,
- `working_material` for scratch text, unresolved notes, internal comments, and unreviewed excerpts.
Profiles: snapshot_reporting
Verified by: AC-057, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-113, AC-114, AC-115, AC-233

The closed vocabulary for artifact release scope is:

- `internal_draft`,
- `internal_review`,
- `external_release`.

**REQ-01-378**
The snapshot and export subsystem MUST evaluate output eligibility against the chosen `release_scope` using at least the following matrix:

- `internal_draft`: any `content_class` except raw blob bytes,
- `internal_review`: any `content_class` except raw blob bytes, and any included `working_material` MUST remain visibly marked non-releasable,
- `external_release`: `derived_analytic`, `curated_narrative`, and only selected `source_evidence` excerpts or thumbnails that are eligible for the chosen `release_scope`. Raw blob bytes and `working_material` MUST NOT appear.
Profiles: snapshot_reporting
Verified by: AC-057, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-113, AC-114, AC-115, AC-233

**REQ-01-379**
For `external_release`, every `curated_narrative` block MUST carry `support_refs[]` containing one or more stable identifiers to supporting findings, events, evidence records, assessments, or query records. A narrative block lacking `support_refs[]` MUST be ineligible for `external_release`.
Profiles: snapshot_reporting
Verified by: AC-057, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-113, AC-114, AC-115, AC-233

Content derived directly from ad hoc note artifacts, `task_request` records, `decision` records, `comm_log` artifacts, `handoff` artifacts, `status_review` artifacts, and `lesson` artifacts SHOULD default to `working_material`.

**REQ-01-380**
Such content MUST NOT appear in `external_release` unless an analyst has explicitly curated it into a separate export-model block that independently satisfies the selected `content_class`, `support_refs[]`, and applicable redaction rules.
Profiles: snapshot_reporting
Verified by: AC-057, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-113, AC-114, AC-115, AC-233

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

1. the operator supplies a pack bundle through the configured reference-pack storage root or an equivalent administrative upload path backed by that root,
2. the system stages the bundle inside the configured temporary-work root,
3. the system verifies the staged bundle before any extracted content becomes active,
4. on successful verification, the system records the candidate version as `available` or an equivalent non-active state,
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

Only the affected overlay views, labels, or enrichment semantics MAY degrade.

## 12. Portability, backup, restore, and failure handling

### 12.1 Backup

A deployable implementation SHOULD support:

- Postgres base backup plus WAL archiving,
- object-store bucket snapshotting or versioning.

Equivalent backup mechanisms are permitted if they preserve restore semantics.

### 12.2 Restore

**REQ-01-423**
Restore MUST occur in this order:

1. restore Postgres,
2. restore object-store contents,
3. rebuild projections.
Profiles: base
Verified by: AC-231

**REQ-01-424**
Projection rebuild MUST be part of restore readiness when projection contents are not restored directly.
Profiles: base
Verified by: AC-231

### 12.3 Incident portability

Whole-incident export/import beyond operational backup and restore belongs to the **Incident Portability Extension Profile**.

**REQ-01-425**
If the implementation claims that profile, it MUST support full-fidelity administrative round-trip transfer of authoritative incident source state between trusted Cartulary deployments without depending on workbook-label semantics, live remote fetches, or deployment-local authentication configuration.
Profiles: incident_portability
Verified by: AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-236

**REQ-01-426**
Import into an existing incident, incident cloning with identifier remapping, and partial bundle merge semantics are out of scope for this profile. A conformant import MUST preserve the exported `incident_id`, `record_id`, `row_version`, change-set identifiers, and blob hashes.
Profiles: incident_portability
Verified by: AC-164, AC-165, AC-166, AC-167, AC-168, AC-169, AC-236

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
- an equivalent historical actor descriptor bound to an existing local user without rewriting the source bundle actor identifier used by imported history.
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
2. validate the outer container and every member path,
3. verify every required checksum and any supported signature before any structured data becomes visible,
4. stage blob bytes and verify every required blob hash,
5. import the structured incident state,
6. rebuild projections,
7. mark the imported incident visible only after the structured import and projection rebuild succeed.
Profiles: incident_portability
Verified by: AC-165, AC-166, AC-167, AC-169, AC-236

**REQ-01-449**
Import MUST fail closed on any of the following:

- checksum mismatch,
- signature mismatch when signature verification is supported or required by deployment policy,
- missing required file,
- missing required blob or blob-hash mismatch,
- invalid path or unsupported bundle member type,
- unsupported `required_capabilities[]` entry,
- duplicate `incident_id`,
- any import path that would require a live remote fetch to complete.
Profiles: incident_portability
Verified by: AC-165, AC-166, AC-167, AC-169, AC-236

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
The deployment configuration MUST name explicit persistent roots for:

- database storage,
- object storage,
- reference-pack storage,
- temporary work files,
- export outputs.
Profiles: base, reference_pack
Verified by: AC-051, AC-055, AC-169, AC-231, AC-234

**REQ-01-456**
The implementation MUST NOT rely on source-tree-relative defaults for icons, Markdown templates, reference data, or generated artifacts.
Profiles: base
Verified by: AC-051, AC-055, AC-169, AC-231

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
A successful preview-handle issuance MUST set `data.handle_kind='preview'`, `data.single_use=false`, `data.disposition='inline'`, and a non-null `data.preview_kind` that uses the exact tokens owned by Core 02 §18. In the base profile, preview issuance MUST succeed only when `data.preview_kind` is one of `image_inline`, `pdf_inline`, or `text_inline`. Preview handles MUST expire exactly 5 minutes after issuance and MUST be reusable until expiry, including repeated byte-range fetches made by a browser preview surface. The server MUST NOT silently downgrade preview issuance into a download contract. When the evidence is otherwise visible but the base-profile preview allowlist does not allow a safe preview, the route MUST fail with `409`, `error.code='evidence_access_unavailable'`, and `error.details.reason_code='unsupported_preview'`.
Profiles: base
Verified by: AC-231, AC-252

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
Issuance MUST use `invalid_evidence_handle_request`, `evidence_record_not_found`, and `evidence_access_unavailable` from §3.3.6.1. Redemption MUST use `handle_not_found_or_revoked`, `handle_expired`, `handle_consumed`, and `evidence_access_unavailable`. Standard authentication or session failures MUST occur before handle-specific lookup and MUST use the ordinary authentication envelope rather than a handle-specific code. Whenever `evidence_access_unavailable` is used on issuance or redemption, `error.details.reason_code` MUST use the exact `evidence_access_unavailable` registry from §3.3.6.2.
Profiles: base
Verified by: AC-231, AC-251, AC-252, AC-253, AC-254, AC-255

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
