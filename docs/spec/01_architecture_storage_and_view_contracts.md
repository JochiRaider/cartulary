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
- entities, indicators, and observation resolution,
- evidence and object storage,
- imports and tabular ingest,
- links, tags, and analyst-work coordination,
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

### 2.1 Phase 2 Workbook Import Assistant

The **Import Extension Profile** MAY expose a **Phase 2 Workbook Import Assistant** for structured file onboarding of CSV and XLSX sources.

The Phase 2 Workbook Import Assistant MUST remain an internal concern of the dedicated `imports` module. It MUST NOT add workbook-specific runtime semantics to Timeline, Hosts, Identities, Evidence, Notes, projections, snapshot generation, or write-back.

The assistant MUST expose file-based import work as:

- one `import_session` for an uploaded source file and one operator-driven workflow,
- one or more explicit `import_unit` objects discovered from that source,
- one `mapping_fingerprint` per selected unit for the operator-approved header-to-field plan,
- zero or more closed-vocabulary `warning_code[]` values for downgraded workbook features.

Whole-workbook import MUST mean an orchestrated batch of explicit `import_unit` objects selected from one `import_session`. It MUST NOT preserve workbook object identity or require runtime workbook semantics outside `imports`.

The current profile MUST limit `import_unit.locator_kind` to `csv_file`, `xlsx_used_range`, `xlsx_table`, `xlsx_named_range`, and `xlsx_region`. Workbook inspection, used-range discovery, table discovery, named-range eligibility checks, operator-selected region previewing, downgrade warnings, and spreadsheet-parser compatibility shims MUST remain inside `imports`.

The semantic identity of an `import_unit` MUST be the tuple `source_content_sha256 + canonical_locator + parser_version`. Modules outside `imports` MUST consume only the stable tabular-ingest contract, shared mapping engine, deterministic provenance, `mapping_fingerprint`, and declared `warning_code[]` values. They MUST NOT link directly against XLSX or OpenXML parsing libraries, workbook-shape heuristics, or workbook-behavior semantics.

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

- authenticated HTTP+JSON API endpoints,
- authoritative mutation validation,
- optimistic concurrency enforcement,
- projection maintenance,
- WebSocket-based collaboration updates,
- background-job orchestration,
- reference-pack activation verification,
- snapshot and report generation when the Snapshot and Reporting Extension Profile is implemented.

### 3.3 Public HTTP and WebSocket interface contract

The base-profile client/server boundary MUST be a versioned HTTP+JSON surface plus a bounded WebSocket event stream. Internal modules MAY use any internal call form, but conformance at the browser-facing boundary MUST be evaluated against this public surface.

#### 3.3.1 Versioning and compatibility

The HTTP surface MUST be rooted at `/api/v1/`. The live-update stream MUST be rooted at `/ws/v1/` or an equivalent versioned WebSocket path that preserves the same observable contract.

Breaking changes to route patterns, required request fields, required response fields, envelope shapes, or event semantics MUST use a new major version root. Additive response fields, additive optional request fields, and additive route families for claimed extension profiles MAY be introduced within the same major version.

All public requests and responses MUST address writable surfaces by stable identifiers. The client MUST identify the incident by `incident_id`, the active view by `view_schema_id`, target rows by `record_id`, optimistic writes by `base_row_version`, writable cells by `field_key`, and multi-change user actions by `client_txn_id`. The public surface MUST NOT require clients to address mutations by visible row order, tab label, column label, projection-table name, or storage-table name.

#### 3.3.2 Session and authentication routes

The public authentication contract MUST be session-based. The authoritative browser-session credential MUST be a server-managed opaque session token carried in an `HttpOnly` `Secure` cookie with `Path=/` and `SameSite=Lax` or a stricter same-site policy. If bearer authentication is enabled for non-browser clients or trusted automation, the implementation MAY additionally accept `Authorization: Bearer <opaque_session_token>` from the same opaque session family. The public token format MUST remain opaque to clients. Conformance MUST NOT require a browser or API client to parse JWT claims or provider-specific assertion contents to determine actor identity, incident scope, or expiry.

The base route family MUST include:

- `POST /api/v1/auth/login`,
- `POST /api/v1/auth/logout`,
- `GET /api/v1/auth/session`.

`POST /api/v1/auth/login` MUST accept a credentials object containing local username and password plus a second-factor assertion when MFA is required. On success it MUST establish the server-managed session and return the same session resource exposed by `GET /api/v1/auth/session`.

##### 3.3.2.1 Session resource and expiry contract

The session routes MUST expose the lifecycle boundaries defined by Core 04 §1.1.1 without requiring client-side token parsing.

`GET /api/v1/auth/session` MUST return, at minimum, the authenticated internal user identity, display name, provider kind, MFA state, `authenticated_at`, `idle_expires_at`, `absolute_expires_at`, `session_expires_at`, and the caller's incident memberships or current incident-role context. `session_expires_at` MUST be the earlier of `idle_expires_at` and `absolute_expires_at`.

`GET /api/v1/auth/session` is an inspection route and MUST NOT by itself extend `idle_expires_at`.

`POST /api/v1/auth/logout` MUST revoke the current session immediately. If that session currently owns one or more accepted WebSocket connections, the server MUST send `session_revoked` on those connections and close them.

When the current session has expired, the auth route family MUST fail closed and a caller can establish a new session only by completing the login flow again with any applicable MFA requirement.

If the Enterprise Authentication Extension Profile is implemented, provider-backed sign-in MUST terminate into this same session contract so the remaining API routes and WebSocket stream remain provider-agnostic.

#### 3.3.3 Route families

The base-profile route set MUST include stable route families for:

- incident discovery and incident metadata: `GET /api/v1/incidents`, `GET /api/v1/incidents/{incident_id}`,
- incident membership inspection: `GET /api/v1/incidents/{incident_id}/memberships`,
- view-schema discovery: `GET /api/v1/view-schemas`, `GET /api/v1/view-schemas/{view_schema_id}`,
- saved-view discovery and persistence: `GET /api/v1/incidents/{incident_id}/saved-views`, `POST /api/v1/incidents/{incident_id}/saved-views`, `PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`, `DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`,
- workbook-preference discovery and persistence: `GET /api/v1/incidents/{incident_id}/workbook-preferences/me`, `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me`, `GET /api/v1/incidents/{incident_id}/workbook-preferences/default`, `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default`,
- workbook query and row creation: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`,
- record mutation, history, rollback, and same-field conflict resolution: `PATCH /api/v1/records/{record_id}`, `GET /api/v1/records/{record_id}/history`, `POST /api/v1/records/{record_id}/rollback`, `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve`,
- entity-mention resolution and equivalent surface-specific resolve actions: `POST /api/v1/entity-mentions/{entity_mention_id}/resolve`,
- blob-slot creation and evidence access: `POST /api/v1/object-blobs`, `POST /api/v1/evidence-records/{record_id}/attach-blob`, `POST /api/v1/evidence-records/{record_id}/preview-handle`, `POST /api/v1/evidence-records/{record_id}/download-handle`,
- background-job status and cancellation: `GET /api/v1/jobs/{job_id}`, `POST /api/v1/jobs/{job_id}/cancel`.

Implementations that claim an extension profile MUST add that profile's route family under the same versioned root rather than overloading base workbook routes. This includes, at minimum, `/api/v1/import-sessions/*`, `/api/v1/reference-packs/*`, `/api/v1/snapshots/*` and `/api/v1/releases/*`, and `/api/v1/incident-bundles/*` for the corresponding claimed extension profiles.

#### 3.3.4 View-shaped read contract

The primary hot-path read route for workbook surfaces MUST be `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`.

A view-query request MUST be view-shaped rather than table-shaped. It MUST accept:

- ordered `sort[]` entries keyed by `field_key`,
- `filters[]` entries keyed by `field_key`,
- zero or one `group_by` value chosen from the view's declared grouping keys,
- a bounded result size such as `limit` or `block_size`,
- an optional opaque `cursor_token`.

A view-query response MUST return:

- `incident_id`,
- `view_schema_id`,
- `rows[]` already shaped for that view,
- `record_id` and `row_version` on every mutable row,
- field-key-addressable cell values,
- scalar `group_values` needed for client-local grouping when grouping is active,
- pagination metadata defined in §3.3.7.

The server MUST NOT serialize group headers or other presentation-only grouping artifacts as writable rows.

#### 3.3.5 Mutation contract

New row creation MUST use `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`. The request MUST include `client_txn_id` and the initial writable values keyed by `field_key`.

Existing-row edits MUST use `PATCH /api/v1/records/{record_id}`. The request MUST include:

- `view_schema_id`,
- `base_row_version`,
- `client_txn_id`,
- `changes[]`, with each entry keyed by `field_key` and carrying the intended write value or equivalent action payload.

The public mutation surface MUST be field-key-based and partial. The client MUST NOT be required to send full projection rows, full record snapshots, or raw storage-table mutations in order to edit one field.

The server MUST validate each requested change against the active view contract, enforce per-field writeability and `conflict_resolution_class`, and route the write to the authoritative source field or declared write action without exposing internal table layout.

A successful create or patch response MUST return the authoritative `record_id`, resulting `row_version`, and the committed field values needed to refresh the visible row.

#### 3.3.5.1 Saved-view and workbook-preference contracts

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

`scope` MUST use the closed vocabulary `private`, `shared`, and `system`.

`owner_user_id` MUST be present for `private` and `shared` saved views. It MAY be null only for `system` saved views.

`saved_view_version` MUST be monotonically increasing per `saved_view_id`.

`query_json` MUST be normalized against the owning `view_schema_id` and encode saved-view sort, filter, and grouping state using stable `field_key` values, ordered `sort[]`, ordered `filters[]`, optional `group_by`, and normalized scalar values. It MUST NOT use visible tab labels, visible column labels, presentation-only group-header text, or other display-only identifiers.

`layout_json` MAY encode presentation concerns such as column order, hidden fields, widths, inspector openness, and equivalent client layout state. `layout_json` MUST NOT be the authority for `saved_view_id`, `incident_id`, `view_schema_id`, `scope`, ownership, authorization, or startup/default surface selection.

`GET /api/v1/incidents/{incident_id}/saved-views` MUST return only the saved-view resources visible to the caller.

`POST /api/v1/incidents/{incident_id}/saved-views` MUST accept `view_schema_id`, optional `scope`, `display_name`, `query_json`, and `layout_json`. If `scope` is omitted, the server MUST treat it as `private`. The ordinary public create route MUST reject `scope='system'`.

`PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` MUST accept `base_saved_view_version` plus changed mutable fields only. It MUST reject attempted mutation of `incident_id`, `saved_view_id`, or `view_schema_id`. If the current saved-view version differs from `base_saved_view_version`, the server MUST reject the patch with an explicit conflict status rather than silently overwriting saved-view state.

`DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` MUST delete only the saved-view configuration object and MUST return the common success envelope with `data.saved_view_id` and `data.deleted=true`.

The workbook-preference route family MUST expose two distinct resources:

- `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` for the authenticated caller's `user_workbook_preferences`,
- `GET` and `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` for the incident-wide `incident_workbook_preferences`.

Both workbook-preference resources MUST use the stable `sheet_ref` union defined in §3.3.10.1.

`workbook-preferences/me` MUST expose, at minimum, `incident_id`, `user_id`, `home_sheet_ref`, `created_at`, and `updated_at`. `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` MUST accept a nullable `home_sheet_ref` and MUST allow any current incident member to set or clear only their own home-surface preference.

`workbook-preferences/default` MUST expose, at minimum, `incident_id`, `default_sheet_ref`, `created_at`, `updated_at`, and `updated_by_user_id`. `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` MUST accept a nullable `default_sheet_ref` and MUST fail closed for callers whose current incident role is not `admin`.

#### 3.3.6 Success and error envelopes

All successful JSON responses MUST use a common envelope with:

- `data` for the primary resource, row batch, or job resource returned by the route,
- `meta.request_id` for a server-generated correlation identifier,
- optional `meta.warnings[]` for machine-readable or display-safe warnings,
- optional `meta.paging` for the cursor metadata defined in §3.3.7,
- optional `meta.query` for the applied view-query contract after server normalization.

All non-success JSON responses MUST use a common error envelope with:

- `error.code` as a stable machine-readable error code,
- `error.message` as a human-readable summary,
- `error.status` as the transport status,
- `error.request_id` as a correlation identifier,
- `error.retryable` as an explicit retry hint,
- optional `error.details` for route-specific validation or state details.

Same-field conflicts MUST use this same error family with `error.code = same_field_conflict` and the additional conflict object defined by Core 03 §3.3.4.

Illegal lifecycle transitions MUST use this same error family with `error.code = illegal_transition`, `error.status = 409`, `error.details.from_status`, `error.details.to_status`, and `error.details.violated_guards[]`. `error.details.violated_guards[]` MUST be present and MAY be empty when the transition is disallowed by the legal transition matrix rather than by a failed field guard.

Clients MUST tolerate additive response members they do not use.

#### 3.3.7 Pagination and cursor contract

Any list or query route that MAY exceed one response page MUST use opaque cursor pagination.

A `cursor_token` MUST be bound to the route family, authenticated actor, `incident_id` when present, `view_schema_id` when present, normalized sort tuple, filter set, and grouping key. The server MUST reject a cursor that is replayed against a different query contract rather than reinterpret it.

The envelope for paged responses MUST include `meta.paging.has_more` and `meta.paging.next_cursor` when another page is available. Hot workbook views MUST NOT require deep `OFFSET` pagination.

#### 3.3.8 Evidence and blob routes

`POST /api/v1/object-blobs` MUST create a pending blob slot and return, at minimum, `object_blob_id`, `upload_state`, a short-lived upload target, `target_expires_at`, and `pending_expires_at`.

In the base profile, the upload target MUST expire 60 minutes after issuance and the pending blob slot MUST expire 24 hours after blob-slot creation. These timers MUST remain separate: target expiry governs upload-target use, while pending-slot expiry governs later finalization eligibility and cleanup.

Blob finalization MUST occur only through an explicit follow-on call. Binding an uploaded blob to incident-visible evidence MUST either:

- attach the returned `object_blob_id` to an existing evidence record through `POST /api/v1/evidence-records/{record_id}/attach-blob`, or
- create a new evidence record through the normal view or record-creation path using that `object_blob_id` as declared input.

The base profile MUST treat a pending blob slot as a single-upload lease. If the upload target expires before successful upload, retry MUST use a fresh `POST /api/v1/object-blobs` call. The base profile MUST NOT require same-slot upload-target refresh or resumable upload semantics.

Preview and download routes MUST return short-lived authorization-checked handles or an equivalent mediated access contract. They MUST NOT expose long-lived object-store credentials, bypass incident membership checks, or treat a `pending` blob slot as attached evidence.

#### 3.3.9 Background-job routes

Routes that start long-running operations MUST return `202 Accepted` with a job resource containing `job_id`, initial `status`, and a status route.

`GET /api/v1/jobs/{job_id}` MUST expose, at minimum, current job status, progress, cancelability, submitted actor, timestamps, and terminal result or error summary. `POST /api/v1/jobs/{job_id}/cancel` MUST fail closed when the job is already terminal or not cancelable.

#### 3.3.10 WebSocket collaboration stream

The incident-scoped WebSocket route MUST be `/ws/v1/incidents/{incident_id}` or an equivalent versioned incident-scoped path.

The live-update stream MUST use a bounded message family rather than a second mutation API. The base message set MUST include:

- session handshake messages and acknowledgements: `hello`, `resume`, `hello_ack`, `resume_ack`,
- incident-scoped presence messages: `presence_snapshot`, `presence_delta`, `presence_update`,
- `record_changed` events,
- incident-scoped `job_progress` events,
- heartbeat messages: `ping`, `pong`,
- terminal `error` or `session_revoked` events.

A replayable `record_changed` event MUST identify the `incident_id`, affected `record_id`, resulting `row_version`, and one or more affected `view_schema_id` entries, each with either deterministic field-key-addressable patch cells, an explicit `invalidate` signal, or an explicit `remove` signal.

The WebSocket stream MUST authenticate with the same server-managed session contract as the HTTP surface. It MUST NOT broadcast unresolved same-field local drafts, transient pending patches, expand or collapse state for grouped rows, or other client-local UI state as authoritative collaboration events.

##### 3.3.10.1 v1 collaboration wire contract

A client MUST send exactly one session-establishment message as the first application message on a connection: `hello` for a fresh connection or `resume` for a reconnect. Until the server accepts one of these messages, it MUST NOT treat the socket as subscribed and MUST NOT emit replayable incident messages on that connection.

A connected socket is subscribed to exactly one incident, determined by the `{incident_id}` path parameter. The protocol MUST NOT negotiate arbitrary topic subscriptions or a second mutation surface. Authoritative record creation, record mutation, rollback, conflict resolution, blob finalization, and other write actions MUST continue to use the HTTP routes defined in §3.3.3 through §3.3.9.

Every application-level collaboration message MUST use one JSON envelope:

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

Within `/ws/v1/`, clients MUST ignore unknown additive members. Breaking changes to required message members, required message types, or message semantics MUST use a new major version root as defined in §3.3.1.

The minimum client-to-server message set MUST be:

- `hello`, whose `payload` MUST include `client_instance_id` and current `presence`,
- `resume`, whose `payload` MUST include `client_instance_id`, `resume_token`, `last_seen_stream_seq`, and current `presence`,
- `presence_update`, whose `payload` MUST include current `presence`,
- `pong`, whose `payload` MAY be empty.

`client_instance_id` MUST remain stable for the lifetime of one browser tab or other local client instance. `resume_token` MUST be opaque to the client, MUST be a replay token rather than an authentication token, and MUST be bound by the server to the authenticated session, `incident_id`, and `client_instance_id`. A `resume_token` MUST NOT authenticate HTTP routes or establish session identity independently of the underlying authenticated session. A `resume_token` MUST expire no later than the earlier of the configured replay window and the underlying session expiry. `last_seen_stream_seq` MUST be the highest replayable `stream_seq` the client has already applied.

The `presence` object MUST be shaped as:

- `sheet_ref`: required stable workbook-surface reference,
- optional `record_id`: stable target row identifier when focus is on a materialized row,
- optional `field_key`: stable target field identifier,
- `mode`: required one of `viewing`, `editing`, or `idle`.

`sheet_ref` MUST address workbook surfaces by stable identifier rather than visible label and MUST be one of:

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

When `sheet_ref.kind = view_schema`, `sheet_ref.id` MUST carry the `view_schema_id`. When `sheet_ref.kind = saved_view`, `sheet_ref.id` MUST carry the `saved_view_id`. `field_key` MUST be present only when the client is focused on a concrete writable field and `mode = editing`.

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

`hello_ack.payload` MUST include `connection_id`, `resume_token`, `server_time`, `heartbeat_interval_ms`, `presence_ttl_ms`, and `resume_window_ms`. `resume_ack.payload` MUST include `status`, a fresh `resume_token`, and `server_high_water_stream_seq`. `resume_ack.payload.status` MUST be one of `replayed`, `reset_required`, or `rejected`. In the base profile, `heartbeat_interval_ms` MUST be `15000`, `presence_ttl_ms` MUST be `45000`, and `resume_window_ms` MUST be at least `300000`.

`presence_snapshot.payload.presences[]` MUST contain zero or more current presence records for the subscribed incident. Each presence record MUST include `connection_id`, `user_id`, `display_name`, `sheet_ref`, `mode`, `observed_at`, and `expires_at`, and MAY include `record_id` and `field_key`. At most one current presence record MAY exist per `connection_id` within one incident.

`presence_delta.payload` MUST include `delta_kind` and `presence`. `delta_kind` MUST be one of `upsert` or `remove`. On `upsert`, `presence` MUST be a full presence record. On `remove`, `presence.connection_id` MUST identify the removed record; other presence members MAY be repeated for convenience but MUST NOT be required.

`record_changed.payload` MUST include `record_id`, `row_version`, `change_set_id`, `actor_user_id`, `changed_field_keys[]`, and `affected_views[]`. Each `affected_views[]` entry MUST include `view_schema_id` and `change_kind`. `change_kind` MUST be one of `patch`, `invalidate`, or `remove`. When `change_kind = patch`, the entry MUST include `patch_cells`, and `patch_cells` MUST be a view-shaped object containing `record_id`, `row_version`, and field-key-addressable cell values for that view. `affected_views[]` MUST be keyed by base `view_schema_id` values, not by visible tab labels, row order, or client-local filter state.

`job_progress.payload` MUST include `job_id`, `status`, `progress`, and `updated_at`, and MAY include `cancelable`, `message`, `result_summary`, or `error_summary`. The incident-scoped stream MUST emit `job_progress` only for jobs whose observable resource or terminal result is scoped to the subscribed `incident_id`.

`error.payload` MUST include `code`, `message`, and `retryable`, and MAY include `details`. `session_revoked.payload` MUST include `reason_code` and MAY include `message`. The base profile MUST support, at minimum, `reason_code` values `session_expired`, `session_revoked`, `incident_access_revoked`, and `concurrency_limit`.

The protocol MUST define two delivery classes:

- replayable ordered messages: `record_changed`, `job_progress`,
- ephemeral non-replayable messages: `hello_ack`, `resume_ack`, `presence_snapshot`, `presence_delta`, `ping`, `error`, `session_revoked`.

`stream_seq` MUST be monotonically increasing per `incident_id` across replayable messages. The server MUST assign `stream_seq` only after the underlying record mutation or incident-scoped job-state change is committed to authoritative server state. The server MUST NOT emit `record_changed` for uncommitted state.

The route path determines the incident subscription. `presence_update` determines only the sender's published presence scope. `record_changed` and incident-scoped `job_progress` MUST be broadcast to all currently authorized subscribers for that incident. Clients MUST determine active-view relevance locally using stable identifiers such as `view_schema_id`, `record_id`, `field_key`, and the client's current query contract.

The server MUST retain replayable messages for resume for at least 5 minutes or 10,000 replayable messages per incident, whichever is larger. After a valid `resume` whose retained replay window still covers `last_seen_stream_seq`, the server MUST send `resume_ack` with `status = replayed`, replay missed replayable messages in strict ascending `stream_seq`, and then send a fresh `presence_snapshot`.

If the server cannot honor incremental replay because the token is unknown, expired, malformed, or older than the retained replay window, but the caller still has valid incident authorization, the server MUST send `resume_ack` with `status = reset_required`, send a fresh `presence_snapshot`, and emit no guessed or partial replay for the missing range. The client MUST then discard incremental assumptions and re-query current workbook state through the HTTP view route before treating the socket as fully synchronized.

The server MAY use `resume_ack.payload.status = rejected` only when a `resume` message is syntactically invalid or the caller fails authentication or authorization checks. A rejected resume MUST be followed by `error` or `session_revoked` and immediate close.

The protocol MUST use application-level `ping` and `pong` messages inside this JSON envelope. The server MUST emit `ping` every 15 seconds when the connection is otherwise idle. The client MUST answer with `pong` within 10 seconds. The server MUST consider the connection dead after 45 seconds without any inbound frame. On clean close, the server MUST remove the corresponding presence immediately. On abrupt disconnect, presence expiry MUST follow this heartbeat timeout rather than a longer stale timeout.

For cookie-authenticated browser connections, the WebSocket upgrade and any session-establishment message MUST validate `Origin` against the configured application origin. The server MUST re-derive incident authorization on initial connect and on `resume`. WebSocket `ping` or `pong`, passive server push, and automatic reconnect or replay MUST NOT extend session idle expiry. If incident membership or session validity is revoked after connection establishment, or if the underlying session expires while the socket remains connected, the server MUST send `session_revoked` and close the socket. After session expiry or revocation, a later connection MUST establish a new authenticated session and use `hello`; `resume` alone is insufficient.

## 4. Storage boundary

### 4.1 Postgres

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
- workbook preference objects,
- immutable report snapshots.

The following are reference packs:

- framework mappings,
- type and icon registries,
- evidence vocabularies,
- optional enrichment datasets,
- view contracts when distributed independently of incidents.

Reference packs MUST version independently of incidents.

Reference-pack manifests and integrity metadata MUST be stored in Postgres. Pack payloads MAY live on local disk or object storage behind the same abstraction, but their activation state and import or activation attestation MUST remain queryable from structured metadata.

Reporting template packs MAY reuse the same integrity-verification and distribution machinery as reference packs. The template selected for a specific snapshot, the selected redaction profile, approval state, and rendered output hashes remain incident data.

## 6. View contracts

Each built-in sheet or contract-backed system view MUST be declared by a **`view_schema`** contract.

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

Default sort tuples MUST be deterministic and MUST include `record_id` as the final tiebreaker unless a later profile explicitly overrides that rule.

A base-profile implementation MUST expose the structured base-profile view-schema registry for conformance inspection through stored `view_schema` rows or an equivalent structured export. Conformance MUST NOT depend on scraping visible tab labels, column labels, or interactive UI behavior alone.

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

The Hosts sheet MUST support interactive filtering and sorting over `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` without requiring synchronous scans of raw note text or evidence blobs.

The Identities sheet MUST support interactive filtering and sorting over `privilege_level`, `mfa_state`, and `reset_status` without requiring synchronous scans of raw note text or evidence blobs.

The Evidence sheet MUST support interactive filtering and sorting over `requested_at`, `received_at`, `collector_party_text`, `source_party_text`, `storage_ref`, `blob_hash`, and attachment or upload state without requiring synchronous blob access.

### 7.2 Contract-backed system views

The base profile MUST support contract-backed system views for:

- indicators,
- compromise assessments,
- task requests,
- decisions.

Framework overlays such as ATT&CK, D3FEND, or VERIS MAY also be exposed as system views when the relevant reference packs are present.

System views MUST follow the same `view_schema_id` contract discipline as built-in sheets.

The Indicators system view MUST project canonical indicator records. It MUST NOT use source artifacts or source-bound indicator observations as the primary row identity.

The Compromise Assessments system view MUST project incident-scoped assessment records. It MUST NOT collapse assessment history into a mutable static property on a host or identity row.

The Task Requests system view MUST project `task_request` records. It MUST support queue-oriented filtering and sorting over `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `blocked_reason`, `completed_at`, `external_ticket_ref`, and `updated_at` without requiring synchronous scans of raw note or artifact text.

The Decisions system view MUST project `decision` records. It MUST support filtering and sorting over `status`, `owner_user_id`, `decision_type`, and `decided_at` without requiring synchronous scans of raw note or artifact text.

Structured coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MAY be surfaced through contract-backed system views or saved views over `artifact_grid_projection`. They MUST NOT require additional built-in sheets in the base profile.

### 7.3 Notes sheet contract

The base profile MUST expose **Notes** as a built-in workbook sheet.

The built-in Notes sheet MUST:

- be declared by a stable `view_schema_id`,
- use `artifact_grid_projection` as its base projection filtered to `artifact_type='note'`,
- support blank-row or equivalent grid-native note creation from the sheet itself,
- remain backed by the shared artifact model rather than a Notes-specific storage silo.

The base profile MUST also expose contextual `add linked note` actions from Timeline, Hosts, Identities, and Evidence. All Notes entry paths MUST create the same underlying artifact record shape.

Notes behavior MUST NOT depend on the visible tab label. If the implementation allows the built-in Notes tab to be renamed or hidden per user, write-back behavior and export semantics MUST remain unchanged because they are bound to `view_schema_id`.

### 7.4 Authoritative base-profile view schema registry

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

These nine entries are the contract-backed surfaces required by §7.1 and §7.2. Additional `view_schema` entries MAY exist for saved views, coordination-artifact surfaces, optional reference-pack overlays, or later profiles, but they MUST NOT change the membership or semantics of the base-profile registry defined in this subsection.

Unless explicitly overridden below:

- `required_reference_pack_keys` MUST be `[]`,
- `record_id` and `row_version` MUST be present as hidden technical fields,
- the ordered default sort tuple is normative and MUST end with `record_id asc`,
- filter semantics MUST be type-driven unless a schema below explicitly overrides them: enum and boolean fields use exact-match inclusion, timestamp and date fields use exact or range predicates, scalar identifier text uses case-insensitive exact or prefix matching, multi-value collections use `contains_any` and `contains_all`, and declared full-text predicates use case-insensitive token search,
- every writable field entry MUST declare `field_key`, read model, write target or write action, and `conflict_resolution_class`,
- every entity-bearing writable field entry MUST declare `entity_binding_mode`,
- fields not declared writable are read-only,
- per-user hide/show or reordering MAY change presentation but MUST NOT change field identity, filter semantics, or write-back semantics.

#### 7.4.1 `cartulary.view.timeline.v1`

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
- read-only computed fields: `timeline.evidence_count`, `timeline.capture_state`, `timeline.edited_at`, `timeline.sort_ts`, `timeline.occurred_day`, `timeline.recorded_day`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`

#### 7.4.2 `cartulary.view.hosts.v1`

- surface: built-in `Hosts` sheet
- source record types: `host`
- base projection: `host_grid_projection`
- `default_visible_fields`: `host.display_name`, `host.hostname`, `host.aliases`, `host.host_state`, `host.linked_event_count`, `host.evidence_count`, `host.location`, `host.os_platform`, `host.business_owner`, `host.criticality`, `host.containment_status`, `host.edited_at`
- `default_hidden_fields`: `record_id`, `row_version`
- `default_sort`: `host.display_name asc`, `record_id asc`
- `filter_fields`: `host.host_state`, `host.business_owner`, `host.criticality`, `host.location`, `host.os_platform`, `host.containment_status`
- inline create: direct row creation or paste on the Hosts sheet MUST create or upsert a `host` record using `entity_binding_mode=entity_origin`
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

#### 7.4.3 `cartulary.view.identities.v1`

- surface: built-in `Identities` sheet
- source record types: `identity`
- base projection: `identity_grid_projection`
- `default_visible_fields`: `identity.display_name`, `identity.upn`, `identity.email`, `identity.sam_account_name`, `identity.aliases`, `identity.identity_state`, `identity.linked_event_count`, `identity.evidence_count`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status`, `identity.edited_at`
- `default_hidden_fields`: `record_id`, `row_version`
- `default_sort`: `identity.display_name asc`, `record_id asc`
- `filter_fields`: `identity.identity_state`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status`
- inline create: direct row creation or paste on the Identities sheet MUST create or upsert an `identity` record using `entity_binding_mode=entity_origin`
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

#### 7.4.4 `cartulary.view.evidence.v1`

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

#### 7.4.5 `cartulary.view.notes.v1`

- surface: built-in `Notes` sheet
- source record types: `artifact`
- base projection: `artifact_grid_projection` filtered to `artifact_type='note'`
- `default_visible_fields`: `note.title`, `note.body`, `note.tags`, `note.linked_record_count`, `note.updated_at`
- `default_hidden_fields`: `record_id`, `row_version`, `note.created_by_user_id`
- `default_sort`: `note.updated_at desc`, `record_id asc`
- `filter_fields`: `note.tags`, `note.created_by_user_id`, `note.updated_at`, plus a contract-declared full-text predicate over `note.title` and `note.body`
- inline create: blank-row or equivalent grid-native creation MUST create an `artifact` record with `artifact_type='note'`
- writable fields:
  - `note.title`: read the note title field; write target the note artifact title field; `conflict_resolution_class=text_compare_merge`
  - `note.body`: read the note body field; write target the note artifact body field; `conflict_resolution_class=text_compare_merge`
  - `note.tags`: read `tag_names`; write action upsert tags and `record_tags`; `conflict_resolution_class=collection_review`
- read-only computed fields: `note.linked_record_count`, `note.updated_at`, `note.created_by_user_id`

#### 7.4.6 `cartulary.view.indicators.v1`

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
- grid edits to an existing indicator row MUST reject writes to `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, `indicator.normalized_value`, and any type-specific dedupe basis.

#### 7.4.7 `cartulary.view.assessments.v1`

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
- grid edits to an existing assessment row MUST reject writes to `assessment.subject_ref`, `assessment.subject_type`, `assessment.assessment_state`, `assessment.confidence_score`, `assessment.rationale`, `assessment.assessor`, `assessment.assessed_at`, and `assessment.support_refs`.

#### 7.4.8 `cartulary.view.task_requests.v1`

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
- task lifecycle semantics remain authoritative: any committed write set affecting `task.status`, `task.blocked_reason`, `task.completed_at`, or `task.owner_user_id` MUST produce a resulting row that satisfies Core 02 §10.4.1.1. In particular, `status='blocked'` requires `blocked_reason`, `status='done'` requires `completed_at`, active tasks MUST NOT be ownerless, a successful transition away from `blocked` or `done` MUST clear `blocked_reason` or `completed_at` respectively, and a successful write that sets `status='done'` with no explicit `completed_at` MUST fill `completed_at` from the commit timestamp.

#### 7.4.9 `cartulary.view.decisions.v1`

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
- supersession remains an explicit decision-linking flow. It MUST NOT be modeled as a direct write to `decision.supersedes_record_id` or as a direct write that sets `decision.status='superseded'` in the base-profile grid.
- when the explicit supersession flow succeeds, it MUST persist the supersession relation and apply the status effects defined by Core 02 §10.4.2.1 atomically.

## 8. Projection model

### 8.1 Projection tables

Hot workbook screens MUST use projection tables rather than Postgres materialized views.

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

Each projection row MUST represent exactly one primary record in the base projection for that view.

For `indicator_grid_projection`, the primary record MUST be the canonical indicator record for that row.

For `assessment_grid_projection`, the primary record MUST be the assessment record for that row.

For `task_request_grid_projection`, the primary record MUST be the `task_request` record for that row.

For `decision_grid_projection`, the primary record MUST be the `decision` record for that row.

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

Projection tables MUST support deterministic interactive retrieval over the ordered default sort tuple and any contract-declared interactive grouping key for the view. An implementation MAY satisfy this with indexes or an equivalent mechanism that preserves the same observable latency envelope.

Projection rows for hot workbook sheets MUST carry the scalar fields required for interactive sort, filter, grouping, selection anchoring, and evidence badges in the visible viewport. For the Timeline sheet, this MUST include at least `sort_ts`, day buckets equivalent to `timeline.occurred_day` and `timeline.recorded_day`, `capture_state`, `has_evidence`, `has_unresolved_mentions`, and `evidence_count`.

For the Hosts sheet, exact lookup, sorting, filtering, and pivot counts over `business_owner`, `criticality`, `location`, `os_platform`, and `containment_status` MUST be satisfiable from `host_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text or evidence blobs.

For the Identities sheet, exact lookup, sorting, filtering, and pivot counts over `privilege_level`, `mfa_state`, and `reset_status` MUST be satisfiable from `identity_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text or evidence blobs.

For the Evidence sheet, exact lookup, sorting, filtering, and queue views over `requested_at`, `received_at`, `collector_party_text`, `source_party_text`, `storage_ref`, `blob_hash`, and upload or attachment state MUST be satisfiable from `evidence_grid_projection` and other small derived metadata. They MUST NOT require synchronous blob access.

The grid and inspector hot path MUST synchronously read only scalar fields, flags, counts, and small preview handles needed for the visible viewport or selected row. They MUST NOT synchronously fetch full attachment lists or binary blob bytes as part of grid rendering, row selection, sheet filtering, grouping, or inspector metadata open.

For the Indicators system view, exact lookup, sorting, filtering, and pivot counts over canonical indicators MUST be satisfiable from `indicator_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw timeline text, artifact text, or evidence blobs.

For the Compromise Assessments system view, exact lookup, sorting, filtering, and pivot counts over `assessment_state` and derived `confidence_band` MUST be satisfiable from `assessment_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw timeline text, artifact text, or evidence blobs.

For the Task Requests system view, exact lookup, sorting, filtering, queue counts, and stale-work views over `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `blocked_reason`, `completed_at`, `external_ticket_ref`, and `updated_at` MUST be satisfiable from `task_request_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text, communications logs, or evidence blobs.

For the Decisions system view, exact lookup, sorting, filtering, and review queues over `status`, `owner_user_id`, `decision_type`, `decided_at`, and supersession state MUST be satisfiable from `decision_grid_projection` and other small derived metadata. They MUST NOT require synchronous scans of raw note text, communications logs, or evidence blobs.

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

### 10.2.1 Rendered artifact lifecycle

For Snapshot and Reporting Extension Profile outputs, Cartulary defines an artifact-scoped lifecycle over each rendered output candidate.

The authoritative persisted representation for this lifecycle MUST be the release record plus its bound release tuple and any bound approval records. Approval state MUST NOT bind to mutable incident rows or to template metadata outside the release record.

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

Approval invalidation MUST be an explicit lifecycle transition on the artifact record. It MUST NOT be implemented only as an implicit UI rule.

A new render with a different logical output slot or different `output_sha256` MUST start as a distinct `pending_approval` candidate. It MUST NOT inherit `approved` or `published` state from an earlier candidate.

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

Content derived directly from ad hoc note artifacts, `task_request` records, `decision` records, `comm_log` artifacts, `handoff` artifacts, `status_review` artifacts, and `lesson` artifacts SHOULD default to `working_material`.

Such content MUST NOT appear in `external_release` unless an analyst has explicitly curated it into a separate export-model block that independently satisfies the selected `content_class`, `support_refs[]`, and applicable redaction rules.

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

When recipient-specific reporting is implemented, a versioned redaction profile MUST also be able to declare zero or more allowed stable incident-local `disclosure_partition_refs[]`.

The closed vocabulary for redaction actions is:

- `allow`,
- `drop`,
- `mask`,
- `truncate`,
- `hash`,
- `stub`.

Redaction MUST run after snapshot materialization and before template rendering.

If a field or block eligible for the chosen `release_scope` appears in the canonical export model without an applicable redaction rule, rendering MUST fail closed.

When a field or block carries `disclosure_partition_refs[]` that are not allowed by the selected redaction profile, the renderer MUST apply the applicable redaction rule or fail closed. If a field or block contains mixed-partition content and no applicable rule can produce a conformant result, rendering MUST fail closed.

Disclosure partition metadata and redaction-profile selection MUST affect only snapshot-derived rendering and release. They MUST NOT affect live workbook queries, projections, or incident authorization.

Each rendered artifact MUST emit a `redaction_manifest` keyed by stable export-model path and rule identifier, recording every field or block that was dropped, masked, truncated, hashed, or stubbed.

A render that changes `template_id`, `template_version`, `redaction_profile_id`, or `redaction_profile_version` relative to an earlier artifact candidate MUST create a distinct `pending_approval` candidate. It MUST NOT inherit approval or publication state from the earlier artifact.

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

### 11.2 Minimum disconnected bundle

For the smallest supported flyaway or disconnected deployment that implements this profile, the deployment MUST preinstall and activate exactly the following three reference packs by default:

- `type_registry.host`
- `type_registry.evidence`
- `type_registry.indicator`

These three packs define the minimum disconnected bundle because base-profile host, evidence, and indicator semantics MUST come from versioned registries rather than hard-coded UI labels or workbook headers.

The smallest supported disconnected bundle MUST NOT require or preinstall framework overlay packs. Current-profile framework add-on pack keys are:

- `framework.attack`
- `framework.d3fend`
- `framework.veris`

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

Other enrichment or framework pack keys MAY exist. They MUST follow the same activation, verification, and degradation rules defined by this profile.

If the Snapshot and Reporting Extension Profile is implemented, template bundles MUST remain separately installable. They MUST NOT count toward the minimum disconnected reference-pack bundle.

Separately distributed `view_contract` packs MAY exist in larger deployments. They MUST NOT be required for the smallest disconnected bundle.

Larger supported disconnected bundles MAY preinstall additional packs, but the minimum disconnected bundle is fixed by this subsection.

### 11.3 Offline import, update, and activation flow

In a flyaway or disconnected deployment, reference-pack update MUST use an offline bundle import flow. The running application MUST NOT perform a live internet fetch as part of pack verification or activation.

The import and update flow MUST satisfy all of the following:

1. the operator supplies a pack bundle through the configured reference-pack storage root or an equivalent administrative upload path backed by that root,
2. the system stages the bundle inside the configured temporary-work root,
3. the system verifies the staged bundle before any extracted content becomes active,
4. on successful verification, the system records the candidate version as `available` or an equivalent non-active state,
5. activation requires an explicit operator action that switches the active version pointer for the target `pack_key`.

#### 11.3.1 Linked reference-pack lifecycle machines

For each imported reference-pack version, Cartulary defines two linked lifecycle machines:

- a verification and availability machine authoritative on `reference_packs`,
- an activation machine authoritative on `reference_pack_activation_state` and `reference_pack_attestations`.

These machines are linked but separate. Successful verification does not by itself activate a version, and activation MUST NOT bypass verification.

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

A `failed` or `missing` version MUST NOT become `active` without first returning through `staged` or `verified_available` by a new import or successful re-verification path. A `disabled`, `failed`, or `missing` version MUST NOT remain or become the active version pointer for its `pack_key`.

At most one version of a given `pack_key` MUST be active at a time.

The implementation MUST retain the previously active version for each `pack_key` until an explicit administrative removal occurs, so operator rollback does not require incident-data changes.

Reference-pack import, verification, and refresh MUST execute as background jobs rather than as blocking grid actions.

### 11.4 Verification and attestation

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

Pack import or activation MUST fail closed on checksum mismatch, signature mismatch, missing required integrity metadata, contract incompatibility, incomplete download or copy, path-traversal attempt, or disallowed content.

If verification fails, the candidate pack version MUST remain inactive, and the previously active version, if any, MUST remain active.

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

Attestation metadata MUST remain queryable from structured metadata without unpacking bundle contents or consulting incident data.

### 11.4.1 Activation safety and observability

Activation MUST read only from a `verified_available` candidate and MUST emit a structured activation attestation bound to the target `pack_key` and `pack_version`.

If an active version is disabled, fails a later integrity or compatibility check, or becomes missing, the implementation MUST remove or replace the active pointer in `reference_pack_activation_state` before any pack-dependent operation can continue to treat that version as active.

Pack lifecycle state MUST be observable from structured metadata alone. At minimum, an operator and a CI conformance test MUST be able to determine the current pack-version condition, the active-version pointer, the previous active version, the last verification result, and the last import or activation attestation without unpacking bundle contents.

### 11.5 Degradation behavior

If optional reference packs are absent, disabled, failed, or missing, Cartulary MUST continue to support:

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

Whole-incident export/import beyond operational backup and restore belongs to the **Incident Portability Extension Profile**.

If the implementation claims that profile, it MUST support full-fidelity administrative round-trip transfer of authoritative incident source state between trusted Cartulary deployments without depending on workbook-label semantics, live remote fetches, or deployment-local authentication configuration.

Import into an existing incident, incident cloning with identifier remapping, and partial bundle merge semantics are out of scope for this profile. A conformant import MUST preserve the exported `incident_id`, `record_id`, `row_version`, change-set identifiers, and blob hashes.

#### 12.3.1 Logical bundle contract

The canonical portability artifact MUST be a logical bundle layout. A `.zip` or `.tar` wrapper MAY be used for transport, but the normative contract is the root directory structure and file contents after extraction.

At minimum, the logical bundle root MUST contain:

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

Bundle member paths MUST use relative forward-slash separators. The logical bundle and any outer archive wrapper MUST reject absolute paths, `.` or `..` segments, symlinks, hard links, device nodes, and other member types outside regular files and directories.

Each blob stored under `blobs/sha256/<sha256-lower-hex>` MUST contain the exact raw bytes whose SHA-256 digest matches the lowercase hex path suffix.

#### 12.3.2 Authoritative-state boundary

Portability export MUST serialize authoritative incident source state and blob bytes. It MUST NOT serialize derived or deployment-local runtime state.

A portability bundle MUST NOT include:

- projection tables or search indexes,
- live presence state,
- client-local draft queues or same-field-conflict queues,
- sessions, presigned URLs, locks, temporary caches, or other ephemeral runtime files,
- password hashes, MFA secrets, external provider configuration, or object-store credentials,
- incident memberships, current permissions, or other deployment-local authorization state.

Reference-pack attestation metadata remains incident-external state. When the Incident Portability Extension Profile embeds reference-pack payloads, the bundle MAY include only the optional embedded-pack payloads and their bundle-local descriptors, not deployment-global activation or attestation history.

#### 12.3.3 Manifest and integrity contract

`manifest.json` MUST be the canonical bundle manifest and MUST include, at minimum:

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

`manifest.json` MUST describe one immutable export boundary. `source_change_set_high_watermark` MUST identify the frozen source boundary used to build the bundle. `history_mode` and `blob_mode` MUST each equal `full` for this profile.

`manifest.json.files[]` MUST enumerate every regular file in the logical bundle except `integrity/checksums.sha256` and `integrity/signature.ed25519`, sorted lexicographically by `path`. `required=true` MUST identify the files required to reconstruct the core incident state. `required=false` MAY be used only for optional embedded sections.

`integrity/checksums.sha256` MUST list one lowercase SHA-256 and relative path per line for every file listed in `manifest.json.files[]`, sorted lexicographically by path, using the exact file bytes carried in the bundle.

If `integrity/signature.ed25519` is present, `manifest.json.signing_key_id` MUST also be present. The signature MUST cover the exact bytes of `integrity/checksums.sha256`. If a deployment supports signature verification for portability bundles, signature failure MUST reject the bundle before any structured data becomes visible.

If `required_capabilities[]` names a capability the target deployment does not implement, import MUST fail closed. Optional embedded sections MAY be ignored when unsupported unless the corresponding capability is listed in `required_capabilities[]`.

#### 12.3.4 Structured formats and deterministic serialization

The portability bundle MUST use JSON for singleton files and NDJSON for multi-row files. CSV, XLSX, or other workbook-shaped exports MAY exist elsewhere. They MUST NOT be the authoritative whole-incident portability format.

All structured files in the bundle MUST use:

- UTF-8 encoding,
- LF line endings,
- no BOM,
- lexicographically sorted object keys,
- exactly one JSON object per NDJSON line,
- deterministic file-level row ordering.

NDJSON files whose rows have a stable single-column primary key MUST sort ascending by that key. `record_tags.ndjson` MUST sort by `(record_id, tag_id)`. `change_set_mutations.ndjson` MUST sort by `(change_set_id, sequence_no)`. Files with integer primary keys such as `record_revisions.ndjson` MUST sort ascending by that integer key.

The canonical JSON serialization for singleton JSON files and per-line NDJSON objects MUST be stable enough that exporting the same incident state twice without intervening mutations produces byte-identical structured files and identical `integrity/checksums.sha256`.

#### 12.3.5 Portable actors and optional embedded sections

`actors.ndjson` MUST preserve historical attribution without exporting portable login material. Each actor descriptor row MUST carry, at minimum:

- stable `actor_id`,
- display name,
- non-secret match hints when available, such as normalized email or provider-subject hint.

Import MUST materialize every actor referenced by imported history as either:

- an inert imported actor that is not login-capable and is not automatically added to incident membership, or
- an equivalent historical actor descriptor bound to an existing local user without rewriting the source bundle actor identifier used by imported history.

`reference_pack_refs.json` MUST list the reference-pack keys and versions referenced by imported saved views, overlays, or optional embedded sections. Missing referenced packs MUST degrade only the affected overlays or saved views. They MUST NOT block import of the core incident state.

If `reference_pack_mode='embedded'`, pack payloads MAY be embedded under `ext/reference_packs/**`. If `optional_sections` contains `snapshots`, immutable snapshot descriptors and rendered artifacts MAY be embedded under `ext/snapshots/**`. Unsupported or missing optional embedded sections MUST NOT block import of the core incident state unless the relevant capability is named in `required_capabilities[]`.

#### 12.3.6 Export and import execution semantics

Whole-incident export and import MUST execute as background jobs rather than as blocking grid actions.

A conformant import MUST execute the following phases in order:

1. stage the supplied outer archive or logical bundle under the configured temporary-work root,
2. validate the outer container and every member path,
3. verify every required checksum and any supported signature before any structured data becomes visible,
4. stage blob bytes and verify every required blob hash,
5. import the structured incident state,
6. rebuild projections,
7. mark the imported incident visible only after the structured import and projection rebuild succeed.

Import MUST fail closed on any of the following:

- checksum mismatch,
- signature mismatch when signature verification is supported or required by deployment policy,
- missing required file,
- missing required blob or blob-hash mismatch,
- invalid path or unsupported bundle member type,
- unsupported `required_capabilities[]` entry,
- duplicate `incident_id`,
- any import path that would require a live remote fetch to complete.

If import fails after staging begins, the target deployment MUST leave no partially visible incident. Staged bytes MAY be retained only in a non-visible administrative quarantine or temporary-work area.

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
- incident portability export and import when the Incident Portability Extension Profile is implemented,
- reference-pack import, verification, and refresh,
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
7. file-based import complexity MUST remain isolated behind the imports module and stable tabular-ingest contract rather than leaking spreadsheet-specific parser behavior into workbook-domain modules,
8. whole-incident portability, when implemented, MUST export authoritative source state and referenced blob bytes rather than projections, snapshots, or deployment-local runtime state.
