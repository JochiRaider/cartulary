# Cartulary Normative Core 03: Workbook Interaction, Collaboration, and Workflows

## 1. Interaction model

Cartulary MUST be **grid first and forms second**.

The primary interaction surface MUST be a workbook-like grid with:

- inline editing,
- keyboard navigation,
- paste,
- low-friction row creation,
- saved or system views over projections,
- a collapsible inspector for enrichment, relationships, history, and destructive actions.

The implementation MUST preserve the spreadsheet mental model at the view layer while keeping source data relational and auditable underneath.

## 2. Workbook surface

### 2.1 Built-in tabs

The workbook MUST expose these built-in tabs in the base profile:

- Timeline,
- Hosts,
- Identities,
- Evidence,
- Notes.

### 2.2 System views

The workbook MUST support additional contract-backed system views, including indicator, compromise-assessment, task-request, and decision surfaces.

The Indicators system view MUST surface canonical indicator rows and support pivots to source-bound observations and lifecycle history without leaving the workbook interaction model.

The Compromise Assessments system view MUST surface incident-scoped assessment rows and support pivots to the assessed host or identity and that subject's prior assessment history without leaving the workbook interaction model.

The Task Requests system view MUST surface `task_request` rows and support queue-oriented filtering and sorting by `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `blocked_reason`, `completed_at`, `external_ticket_ref`, and `updated_at` without leaving the workbook interaction model.

The Decisions system view MUST surface `decision` rows and support review-oriented filtering and sorting by `status`, `owner_user_id`, `decision_type`, and `decided_at` without leaving the workbook interaction model.

Structured coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MAY be exposed through contract-backed system views or saved views over artifact-backed records. They MUST NOT require additional built-in tabs in the base profile.

Such views MUST remain workbook surfaces rather than separate application modules.

### 2.3 Saved views

A saved view MUST be an incident-bound workbook configuration over exactly one `view_schema_id`.

A saved view MUST persist, at minimum:

- stable `saved_view_id`,
- owning `incident_id`,
- immutable `view_schema_id`,
- `scope` with one of `private`, `shared`, or `system`,
- `display_name`,
- normalized `query_json`,
- `layout_json`,
- `owner_user_id`,
- `created_at`,
- `updated_at`,
- `saved_view_version`.

A saved view created from another saved view MUST persist a normalized copy of that source view's `view_schema_id`, `query_json`, and `layout_json`. After creation, the new saved view MUST NOT inherit runtime behavior from the source `saved_view_id`.

`owner_user_id` MUST be present for `private` and `shared` saved views. It MAY be null only for `system` saved views.

`incident_id`, `saved_view_id`, and `view_schema_id` MUST be immutable after creation.

A `system` saved view is a saved-view configuration object with `scope='system'`. It is not the same object as a contract-backed system view identified by `view_schema_id`.

#### 2.3.1 Scope and discoverability

Saved-view scope MUST use exactly these three values:

- `private`,
- `shared`,
- `system`.

`private` means the saved-view object is visible only to its owner and incident admins. Any incident member MUST be able to create a `private` saved view for their own use.

`shared` means the saved-view object is visible to all incident members. It MUST still record one owner for accountability. The owner and incident admins MUST be able to update or delete it in place. Other incident members MAY open it and duplicate it, but MUST NOT update or delete it in place through the ordinary saved-view routes.

`system` means an implementation-owned or admin-seeded saved-view configuration still bound to one incident. It MUST be visible to all incident members, but it MUST be immutable through the ordinary saved-view write path. Users MAY duplicate a visible `system` saved view, but they MUST NOT edit or delete it in place through the ordinary saved-view routes.

Saved-view scope controls discoverability and mutability of the saved-view object only. It MUST NOT widen or narrow access to underlying incident rows, fields, search results, export redaction behavior, or evidence visibility.

#### 2.3.2 Ordinary lifecycle semantics

Ordinary saved-view listing MUST return only the saved views visible to the caller.

Ordinary saved-view creation MUST default `scope` to `private` when the request omits it. The ordinary public create path MUST reject `scope='system'`.

Ordinary saved-view update MUST allow mutation only of `display_name`, `query_json`, `layout_json`, and, when permitted by scope rules, `scope` between `private` and `shared`. It MUST reject attempted mutation of `incident_id`, `saved_view_id`, or `view_schema_id`. Every successful in-place saved-view mutation MUST advance `saved_view_version`.

Ordinary saved-view delete MUST delete only the saved-view configuration object. It MUST NOT delete or mutate underlying incident rows, links, tags, evidence records, or canonical entities.

A user who can open a visible saved view MUST be able to duplicate it into a new saved view by persisting a normalized copy of its current `view_schema_id`, `query_json`, and `layout_json`.

### 2.4 Startup and default surface selection

Saved-view scope and startup/default surface selection MUST remain separate concerns.

The per-user startup pointer MUST be `user_workbook_preferences.home_sheet_ref`. The incident-wide fallback pointer MUST be `incident_workbook_preferences.default_sheet_ref`.

Both pointers MUST use the stable `sheet_ref` shape defined by Core 01 §3.3.10.1.

Workbook open MUST select the starting surface in this order:

1. an explicit launch `sheet_ref`, if present and still valid for the caller,
2. the caller's `home_sheet_ref`, if present, still valid, and visible to the caller,
3. the incident-wide `default_sheet_ref`, if present, still valid, and visible to the caller,
4. `cartulary.view.timeline.v1`.

If a referenced saved view or view schema is missing, no longer visible to the caller, or invalid because a required optional pack is unavailable, the implementation MUST clear the invalid pointer and continue to the next step in the ordered fallback chain rather than failing workbook open.

Any incident member MUST be able to set or clear their own `home_sheet_ref`. Only incident admins MUST be able to set or clear `incident_workbook_preferences.default_sheet_ref`.

## 3. Collaboration and concurrency model

### 3.1 Concurrency strategy

The implementation MUST use **field-level optimistic concurrency on top of row versioning**.

Each visible row MUST be bound to the `record_id` and `row_version` emitted by its projection row.

Every grid write MUST include:

- `record_id`,
- the client’s `base_row_version`,
- changed fields only.

### 3.2 Server-side conflict behavior

If the current row version matches the supplied `base_row_version`, the patch MUST apply normally.

If another user has changed the row:

- when the other change touched a different field, the server MUST auto-rebase and accept the write,
- when the other change touched the same field, the server MUST reject the write with a conflict payload.

For `conflict_resolution_class='text_compare_merge'`, same-field detection MUST be keyed by `field_key`, not by textual subrange. The server MUST reject the original patch with `error.code='same_field_conflict'` whenever another committed write changed that same writable `field_key` after `base_row_version`, even if a deterministic clean text merge suggestion would be possible.

At minimum, the conflict payload MUST preserve:

- the stable `field_key`,
- the client-submitted value,
- the persisted server value,
- the client `base_row_version`,
- the current `row_version`.

A same-field conflict MUST remain unresolved until an analyst explicitly chooses a resolution. Same-field edits MUST NOT silently overwrite each other.

### 3.3 Analyst-facing same-field conflict resolution

#### 3.3.1 Resolver surface and conflict state

When a same-field conflict occurs, only the affected cell MUST enter conflict state.

The conflicted cell MUST continue to display the current saved server value plus a visible conflict marker. The client MUST retain the analyst's unsaved local value separately and MUST NOT render that local value as though it were already saved.

The analyst-facing resolver MUST open from the conflicted cell as a compare drawer or an equivalent same-surface panel. The main grid MUST remain visible while the resolver is open, and the row containing the conflicted cell MUST remain in view.

Closing the resolver without selecting a resolution MUST leave the cell in conflict state.

The workbook save-state presentation MUST remain `Conflict` until every unresolved same-field conflict local to that client has been cleared or resolved.

When same-cell presence is available before the conflict occurs, the UI SHOULD show a hint equivalent to another analyst actively editing the field.

The analyst MUST be able to continue editing other rows or cells while a same-field conflict on a different cell remains unresolved.

After an explicit same-field conflict resolution or clear action, focus MUST return to the same cell and scroll position MUST be preserved.

#### 3.3.2 Resolver contents and safety rules

The resolver MUST display:

- row context sufficient to identify the record without leaving the sheet,
- the field display label and stable `field_key`,
- the saved value plus the actor and timestamp of that saved value,
- the analyst's unsaved local value,
- for merge-capable fields, a diff summary against the common base,
- direct resolution actions in plain language.

The resolver SHOULD open with a message equivalent to: `This field changed before your edit was saved. Review the saved value and your unsaved value.`

The resolver MUST NOT default focus to a destructive action such as `Use my unsaved value`.

Initial keyboard focus MUST land on the conflict summary or an equivalent non-destructive element.

Pressing Enter while the resolver first opens MUST NOT resolve the conflict.

#### 3.3.3 Contract-declared resolution classes

For each write-back-capable `field_key`, `view_schemas.writeback_contract` MUST declare `conflict_resolution_class`.

The closed vocabulary is:

| `conflict_resolution_class` | Required use | Required resolver behavior |
| --- | --- | --- |
| `atomic_replace` | scalar fields such as timestamps, enums, numbers, single-value identifiers, and state fields | present explicit `Keep saved value` and `Use my unsaved value` actions |
| `text_compare_merge` | analyst-authored free text such as `summary`, `details`, note body, and description | present side-by-side comparison with change highlighting and an optional `Edit merged value` path |
| `collection_review` | multi-value chip or set fields such as tags, Timeline Hosts, and Timeline Identities | present base, saved, and local deltas plus a final preview before commit |

Unknown or omitted classes MUST behave as `atomic_replace`.

#### 3.3.3.1 Operational semantics for `text_compare_merge`

`text_compare_merge` is a plain-text merge-capable conflict class. It MUST treat the field value as a scalar text value and MUST NOT parse or interpret Markdown structure, HTML structure, entity chips, links, or any other rendered representation.

For merge computation only:

- `null` MUST be treated as the empty string,
- line endings MUST be normalized from `CRLF` or `CR` to `LF`,
- all other code points MUST be preserved exactly.

The authoritative merge unit is the line. The server MAY compute a deterministic three-way merge suggestion from `base_value`, `server_value`, and `client_value`, but the presence or absence of a clean suggestion MUST NOT change same-field conflict detection under §3.2.

For this class, a changed line hunk is a maximal contiguous changed line range relative to the normalized `base_value`. A clean merge suggestion exists only when the client-side and server-side changed line hunks are disjoint in base-line coordinates and the two sides do not both insert at the same base boundary. When a clean suggestion exists, the server MUST materialize `suggested_merged_value` by applying the non-overlapping hunks to the normalized `base_value` in ascending base order.

Word-level or character-level highlighting MAY be used for presentation inside changed lines, but that highlighting is non-authoritative and MUST NOT control conflict detection or merge validity.

The server MUST NOT silently auto-commit, auto-accept, or auto-retry a same-field text edit. If the deterministic line merge is clean, the conflict object MAY include `suggested_merged_value`. If the merge detects overlapping changed line hunks, or competing insertions at the same base position, the conflict object MUST omit `suggested_merged_value`.

The merge editor MUST start from the current saved value and MUST keep the analyst's unsaved local draft visible as reference. If `suggested_merged_value` is present, the UI MAY offer `Apply suggested merge` or equivalent, but MUST NOT apply it implicitly.

For short one-line fields such as `task.title`, this line-based rule means a clean suggestion will often be absent. In that case the analyst resolves the conflict explicitly using the saved value and local draft already shown in the resolver.

A later profile that needs structured Markdown, rich-text, or semantic entity-aware merge MUST define a new `conflict_resolution_class` rather than widening `text_compare_merge`.

For relationship cells that mix unresolved mention tokens and canonical entity links, the resolver MUST preserve those as different object types. It MUST NOT coerce unresolved tokens and canonical chips into plain delimited text solely for conflict presentation.

#### 3.3.4 Same-field conflict transport contract

A same-field conflict response MUST use `409 Conflict` or an equivalent explicit concurrency-conflict status in deployments that do not expose HTTP directly.

A same-field conflict response MUST use the generic error envelope defined in Core 01 §3.3.6 rather than the normal transient retry path.

The response MUST set `error.code` to `same_field_conflict`, set `error.status` to `409`, and include an `error.conflict` object.

The `error.conflict` object MUST include at least:

- `conflict_token`,
- `record_id`,
- `field_key`,
- `conflict_resolution_class`,
- `base_row_version`,
- `current_row_version`,
- `client_value`,
- `server_value`,
- `server_updated_by`,
- `server_updated_at`,
- `base_value` for merge-capable fields, or `base_revision_ref` only when a later profile or future `conflict_resolution_class` explicitly allows it.

When `conflict_resolution_class='text_compare_merge'`, the conflict object MUST include `base_value`; `base_revision_ref` alone is insufficient for the base profile. In that case `error.conflict.client_value`, `error.conflict.server_value`, and `error.conflict.base_value` MUST each be the raw text scalar for the field or `null`. The conflict object MAY additionally include `suggested_merged_value`, which, when present, MUST also be the raw text scalar for the field or `null`. The presence of `suggested_merged_value` means only that the server found a deterministic clean line merge suggestion and MUST NOT imply that the rejected write has been accepted.

When `resolution_kind='merged_value'` targets `conflict_resolution_class='text_compare_merge'`, `resolved_value` MUST be the final plain-text scalar for the field or `null`. `resolved_value` MUST NOT be a diff script, markup object, token list, AST, or field-specific merge action object. The server MUST persist that value through the field's ordinary write target, subject only to ordinary field validation and the same newline canonicalization used by ordinary writes.

When `conflict_resolution_class='collection_review'`, the conflict object MUST include `base_value`. In that case `error.conflict.client_value`, `error.conflict.server_value`, and `error.conflict.base_value` MUST each use `collection_value_v1`. The server MUST preserve distinct item kinds in those values and MUST NOT collapse unresolved mentions, resolved entity refs, tags, aliases, or linked-record references into raw string arrays or plain delimited text for conflict transport.

When `resolution_kind='merged_value'` targets `conflict_resolution_class='collection_review'`, `resolved_value` MUST use `collection_actions_v1` evaluated against the `server_value` collection carried in the same conflict object.

The client MUST place same-field conflicts into a client-local conflict queue keyed by the canonical composite `record_id:field_key`.

The client MUST keep this conflict queue separate from the transient pending-patch queue used for retryable transport failures.

A same-field conflict MUST NOT auto-retry.

When exposed over the public HTTP surface, the explicit resolution request MUST use `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve`.

An explicit resolution request MUST include:

- `conflict_token`,
- `resolution_kind` with one of `keep_saved`, `use_unsaved`, or `merged_value`,
- `resolved_value` when required by the chosen `resolution_kind`.

If the same field changes again before the analyst resolves the conflict, the server MUST reject the stale `conflict_token` and return a fresh same-field conflict payload in the same generic error envelope. The client MUST preserve the analyst's unsaved local draft and refresh the compare surface against the newest saved value.

#### 3.3.5 Local draft, history, and analytics boundary

Unresolved same-field conflict drafts MUST be treated as client-local unsaved work rather than authoritative incident state.

Until explicitly committed, they:

- MUST NOT be broadcast to other analysts,
- MUST NOT appear in history, search, exports, or snapshots,
- MUST remain memory-local in the base profile.

If a deployment later adds durable local draft persistence, that persistence MUST be explicitly configured and session-scoped.

Selecting `Use my unsaved value` or `Edit merged value` MUST create a new attributed `change_set` and a new row-field mutation against the current row version.

Selecting `Keep saved value` MUST clear the local conflict without creating a new source revision.

The implementation MAY record same-field conflict analytics, but it MUST NOT persist discarded field content solely for analytics. Conflict analytics SHOULD be limited to field identifier, timing, actors, and chosen outcome.

#### 3.3.6 Paste-time same-field conflicts

During multi-cell paste:

- non-conflicting cells MUST commit normally,
- same-field conflicts MUST move into the same-field conflict queue,
- the UI MUST present a grouped conflict tray or equivalent grouped navigator rather than one blocking modal per conflicted cell.

When a paste produces at least one same-field conflict, the committed non-conflicting portion of the paste MUST remain one visible `change_set`. Each later per-cell conflict resolution MUST create its own attributed `change_set`.

The base profile MUST NOT offer a blanket action equivalent to `Use my value for all conflicts`.

### 3.4 Client addressing rules

The client MUST NOT address or mutate a row by:

- visible row number,
- sort position,
- filtered position,
- grouped position,
- displayed cell values.

## 4. Save behavior and presence

### 4.1 Autosave

The grid MUST autosave on:

- Enter,
- Tab,
- blur,
- paste completion.

The normal workflow MUST NOT require an explicit Save button.

### 4.2 Save-state presentation

The UI MUST present a compact save state equivalent to:

- `Syncing`,
- `Saved`,
- `Conflict`.

### 4.3 Presence

The UI MUST provide all of the following presence indicators:

- workbook-header presence avatars for users on the same sheet,
- row-gutter indicators when another analyst is focused on a row,
- same-cell indicators when another analyst is actively editing the same field and such a signal is available.

Presence and live row updates MUST be driven by the bounded WebSocket message families defined in Core 01 §3.3.10 rather than by a second mutation API.

#### 4.3.1 Collaboration message application

The client MUST include its initial workbook presence in `hello` or `resume` and MUST send `presence_update` whenever any of the following changes:

- the active workbook surface changes to a different `sheet_ref`,
- the focused `record_id` changes,
- same-cell editing starts or stops for a writable `field_key`,
- the client becomes `idle` or returns from `idle`.

The client MAY coalesce rapid local cursor motion, but the server-visible presence state for any stable user-visible change MUST settle within 1 second.

Workbook-header presence avatars MUST be derived from `presence_snapshot` and `presence_delta` records whose `sheet_ref` exactly matches the active workbook surface. Row-gutter indicators MUST be derived from matching `record_id`. Same-cell indicators MUST be derived from matching `record_id` plus `field_key` with `mode = editing`. The client MUST key these indicators from `sheet_ref`, `record_id`, and `field_key`; it MUST NOT infer collaboration state from visible tab labels, row numbers, or column headers.

Presence updates are ambient last-write-wins state. They MUST NOT change the local save state, conflict queue, or pending-patch queue.

When a replayable `record_changed` message arrives, the client MUST de-duplicate it by `(incident_id, stream_seq)`. If the client detects a gap in replayable `stream_seq` or receives `resume_ack.status = reset_required`, it MUST stop incremental apply and re-query the current view through the HTTP query route before presenting the sheet as synchronized again.

When the active surface's underlying `view_schema_id` appears in `record_changed.payload.affected_views[]`, the client MUST apply the matching entry as follows:

- `patch`: update only the supplied field-key-addressable cells, replace the row's `row_version` with the authoritative value from the event, preserve selection and edit anchoring by `record_id`, and clear any matching local pending-patch entry only when the authoritative event covers that row and row version or a later committed row version,
- `invalidate`: mark the row or affected visible block dirty and refresh it through the existing HTTP view-query route rather than inventing a separate WebSocket read path,
- `remove`: remove the row from the current materialized grid or mark it absent on the next synchronized query, and if the removed row was selected, clear the selection or move it according to the normal row-removal behavior without silently rebinding the old selection to a different `record_id`.

A client that originated the mutation MUST reconcile against the same echoed `record_changed` message family as every other subscriber. Incoming collaboration messages MUST NOT surface unresolved same-field local drafts as saved state or overwrite the client-local conflict queue defined in §3.3.4 and §3.3.5.

### 4.4 Local pending queue

The client MUST maintain a small local pending-patch queue so that transient network interruptions do not lose typed data.

An authentication failure on a queued write, or a `session_revoked` event on the collaboration stream, MUST NOT discard unresolved same-field local drafts or queued unsent patches. The client MUST preserve that client-local unsaved work, prompt for re-authentication when required, and replay queued writes only after a new authenticated session is established. Replayed writes MUST still satisfy ordinary `base_row_version`, authorization, and same-field conflict checks before they become authoritative incident state.

## 5. Locking policy

Routine inline edits MUST NOT use hard record locks.

Short-lived server-side record locks MAY be used only for rare destructive operations such as:

- merging host or identity records,
- rolling back an existing change,
- restoring a soft-deleted record.

## 6. Record lifecycle

The workflow phrase:

`rough capture -> enriched -> linked -> reviewed -> superseded or rolled back`

remains a helpful summary of typical analyst progress, but for Timeline rows the current profile also defines a normative persisted lifecycle machine on `capture_state`.

For Timeline rows, `capture_state` is a system-managed persisted workflow state with the closed vocabulary `rough`, `enriched`, `reviewed`, and `superseded`.

The authoritative machine condition is determined by the persisted `capture_state` value together with the row-creation event, the explicit reviewer actions defined below, and whether the committed `change_set` contains one or more `capture-state-material` Timeline mutations as classified in §15.

The machine states mean:

- `rough`: first committed capture of a Timeline row,
- `enriched`: a Timeline row that has received at least one later `capture-state-material` mutation after creation and is not currently `reviewed` or `superseded`,
- `reviewed`: a Timeline row whose current version has been explicitly marked reviewed by an authorized reviewer action,
- `superseded`: a Timeline row that has been explicitly marked superseded for ordinary workflow by an authorized reviewer action.

Allowed lifecycle triggers are Timeline row creation, a later committed `capture-state-material` Timeline mutation, the explicit `mark-reviewed` action, the explicit `supersede` action, and rollback of a prior change.

A newly created Timeline row MUST persist with `capture_state='rough'`. This applies to blank-row creation, paste-created rows, and rows created with only an attached screenshot.

The first committed `capture-state-material` mutation to a row whose current `capture_state` is `rough` MUST set `capture_state='enriched'` in the same committed `change_set`.

The explicit review action MUST:

- be exposed only as a non-grid action from the inspector, history surface, or an equivalent reviewer-only surface,
- require current incident role `reviewer` or `admin`,
- be available only when the current `capture_state` is `rough` or `enriched`,
- set `capture_state='reviewed'` in the same committed `change_set`,
- create an attributed change that is visible through ordinary row history.

Any committed `capture-state-material` mutation to a row whose current `capture_state` is `reviewed` MUST set `capture_state='enriched'` in the same committed `change_set`. A tag-only or other non-material change MUST leave `capture_state='reviewed'` unchanged.

The explicit supersede action MUST:

- be exposed only as a non-grid action from the inspector, history surface, or an equivalent reviewer-only surface,
- require current incident role `reviewer` or `admin`,
- require a non-empty reason captured in the resulting attributed change,
- be available only when the current `capture_state` is `rough`, `enriched`, or `reviewed`,
- set `capture_state='superseded'` in the same committed `change_set`,
- create an attributed change that is visible through ordinary row history.

`superseded` is terminal for ordinary Timeline workflow. Ordinary Timeline patch, enrichment, and review actions MUST NOT mutate a row whose current `capture_state` is `superseded`. Leaving `superseded` MUST require rollback of the superseding change or an equivalent reviewer-only history reversal. The base profile defines no ordinary direct transition out of `superseded`.

`linked` is a derived milestone meaning the record has acquired one or more typed links, resolved mentions, evidence associations, or equivalent relational structure. It MUST NOT be stored as a separate `capture_state` value in the current profile.

`rolled back` is a history or reviewer action outcome, not a persisted `capture_state` value.

The implementation MUST NOT re-derive `capture_state` from current link counts, evidence counts, tags, or current unresolved-mention state alone. Removing all current links or evidence from an already `enriched` row MUST NOT revert it to `rough`. `timeline.has_unresolved_mentions` remains a separate derived field from current unresolved mention rows and MUST NOT override `capture_state`.

Normalization MUST add structure without erasing the original observed input.

## 7. Timeline creation workflow

Typing into the blank trailing timeline row MUST create a real record as soon as one non-empty user-entered value exists.

A timeline row MUST be persistable with:

- only one non-empty user-entered value, or
- only an attached screenshot.

At row creation time, the system MUST support:

- nullable `occurred_at`,
- summary text,
- raw mention tokens for host and identity references,
- raw indicator-bearing text in summary, details, source text, or other contract-declared fields without requiring dedicated IOC columns.

The system MUST NOT block row creation on missing canonical host, identity, or indicator records.

Indicator capture from supported source fields is enrichment. Row creation MUST preserve raw field text and MUST NOT require immediate indicator extraction, resolution, or canonical indicator selection.

## 8. Evidence attachment workflow

### 8.1 Two-step upload

Binary evidence attachment MUST use a two-step flow:

1. create a pending blob slot and return an upload target,
2. finalize the evidence attachment after the upload completes.

When exposed over the public HTTP surface, step 1 MUST use `POST /api/v1/object-blobs`. Step 2 MUST use `POST /api/v1/evidence-records/{record_id}/attach-blob` or the normal record-creation path that binds the returned `object_blob_id` during evidence-row creation.

In the base profile, the upload target expires 60 minutes after issuance and the pending blob slot expires 24 hours after creation. The pending slot is a single-upload lease. If the upload target expires or the upload never completes, retry MUST use a fresh `POST /api/v1/object-blobs` call. The base profile MUST NOT require same-slot upload-target refresh or resumable upload semantics.

The system MUST NOT leave fake attached evidence rows for incomplete uploads.

### 8.2 Pending evidence without blob

The evidence model MUST also support requested or pending evidence records with no blob yet attached, followed later by receipt, availability, custody, and optional blob attachment.

### 8.3 Blob and evidence lifecycle bridge

Cartulary defines two linked but separate lifecycle machines for evidence attachment:

- blob upload, authoritative on `object_blobs.upload_state`,
- evidence custody and availability, authoritative on `evidence_records.lifecycle_state` plus custody events.

The blob-upload machine uses conditions equivalent to `pending`, `available`, `failed`, and `quarantined`.

The evidence machine uses states equivalent to `requested`, `pending_receipt`, `received`, `available`, `quarantined`, and `released`.

The required bridge behavior is:

- a `pending` blob slot MUST NOT by itself create or imply an attached evidence row,
- finalize attachment MUST succeed only when the selected blob slot is `available` or `quarantined`,
- an evidence record MUST NOT surface as `available`, previewable, or released while its linked blob is `pending`, `failed`, or missing,
- if a linked blob becomes `quarantined`, normal preview and download MUST stop and the evidence surface MUST show the evidence as `quarantined` or otherwise non-available,
- requested or pending evidence with no blob remains valid and MUST continue to support later receipt, custody, and optional blob linkage.

Abandoned pending uploads MUST fail closed. A blob slot left in `pending` without successful finalization MUST NOT increment row evidence counts, MUST NOT show as attached, and MUST remain eligible only for retry, timeout handling, or administrative cleanup.

A pending blob slot that reaches `pending_expires_at` without successful finalization MUST transition to `upload_state='failed'` with `terminal_reason='pending_timeout'`. Timeout MUST NOT create a distinct blob lifecycle state.

The implementation MUST allow 3 failed explicit finalization attempts per pending blob slot. Only unsuccessful explicit finalization attempts count toward this limit; an idempotent replay after already-committed success MUST NOT consume retry budget. On the 4th failed finalization attempt, the slot MUST transition to `upload_state='failed'` with `terminal_reason='finalize_retry_exhausted'`. Later retry then requires a fresh blob slot.

Pending-blob timeout handling and orphaned unattached-blob cleanup MUST run as background cleanup work at least every 15 minutes. Any blob slot still in `pending` after `pending_expires_at` MUST be marked `failed` by the next cleanup sweep. Orphaned object bytes for a failed unattached blob slot MUST be deleted within 1 hour of terminal failure. Failed unattached slot metadata MUST remain queryable for at least 7 days, after which it MAY be hard-deleted automatically or by administrative action.

If the system detects an inconsistent blob-versus-evidence state, it MUST block preview and download and surface the row as inconsistent until explicit repair or re-finalization completes.

### 8.4 Evidence access

Evidence preview MUST open without forcing a full-page navigation away from the grid. A bottom or side preview is acceptable.

The public interface for preview or download MUST use short-lived authorization-checked preview or download handles, such as `preview-handle` or `download-handle` routes. It MUST NOT expose long-lived object-store credentials or bypass the blob-versus-evidence lifecycle checks defined above.

## 9. Mention resolution workflow

The inspector MUST support:

- resolving a selected unresolved mention to an existing entity,
- creating a host or identity from a selected mention,
- dismissing a selected mention,
- restoring a dismissed mention to unresolved state.

Resolving or creating from a selected mention MUST preserve the raw mention.

Creating from a selected mention MUST create exactly one stub entity by default and resolve only the selected mention unless the user later invokes an explicit bulk action.

Dismissing a selected mention MUST create a new attributed change, preserve the raw mention and source-bound mention identity, set `resolution_status='dismissed'`, clear any current resolution metadata, and remove or tombstone any corresponding active resolved `record_link` in the same `change_set`.

Dismissed mentions MUST NOT remain in the default unresolved resolution queue, MUST NOT contribute to `timeline.has_unresolved_mentions`, and MUST NOT appear in the active relationship collection value for the row. They MAY remain visible through history and a secondary dismissed-mentions section or toggle in the inspector.

Restoring a dismissed mention MUST create a new attributed change that returns the mention to `unresolved`, preserves the raw mention unchanged, and leaves resolution metadata empty. Ordinary restore MUST NOT silently relink to any prior resolved entity; exact pre-dismiss state recovery belongs to reviewer rollback.

### 9.1 Source-bound indicator workflow

The inspector or an equivalent same-surface enrichment flow MUST support:

- linking a selected raw source value or text span in a supported field to an existing indicator,
- creating a canonical indicator from a selected source-bound observation,
- dismissing or marking a selected source-bound observation as non-indicator,
- viewing observation provenance and current lifecycle intervals for the selected indicator.

Indicator capture from Timeline, Notes, Evidence, and other supported source fields MUST preserve the raw source field unchanged. It MUST NOT require dedicated IOC columns on non-indicator sheets.

Direct row creation on the Indicators system view MUST create canonical indicator records. Grid edits to an existing indicator row MUST be limited to the fields that the active `view_schema` declares writable for existing indicator rows. In the base profile, `cartulary.view.indicators.v1` declares no writable fields for existing indicator rows. The exact identity-defining immutable field set for an existing indicator row is `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, `indicator.normalized_value`, and, when populated and used by the canonical dedupe key, `indicator.hash_algorithm` plus `indicator.hash_value`. `indicator.stix_pattern` and `indicator.defanged_value` remain create-only in this view but MUST NOT be treated as identity-defining. Observation creation from a source field MUST remain distinct from canonical indicator creation even when both happen in one analyst action.

## 10. Reviewer history and rollback workflow

### 10.1 Reviewer lens

The history experience MUST remain row-centric.

### 10.2 Minimum history presentation

In the base profile, the history panel for a selected row MUST show:

- actor,
- timestamp,
- operation,
- diff summary expanded to field, link, mention, tag, evidence-entry, and capture-state-transition units.

The history surface MUST receive enough structured metadata from `GET /api/v1/records/{record_id}/history` to render only legal reviewer actions without client-side inference from visible text. At minimum, each displayed logical history item MUST expose `change_set_id`, `reversible`, `available_rollback_actions[]`, optional `history_entry_ref` when the item maps to exactly one reversible mutation target, and `revision_no` when whole-row restore is legal.

### 10.3 Rollback granularity

The reviewer UI MUST allow rollback of a single logical history entry when that entry maps to one reversible mutation target, including:

- one scalar field edit,
- one link add or remove,
- one tag add or remove,
- one mention resolve, dismiss, or restore,
- one auto-resolution or auto-match,
- one evidence attach or detach association.

The UI MUST also expose:

- whole-row restore,
- whole-change-set rollback,

as secondary actions for destructive or multi-target changes.

Arbitrary user-selected subsets of fields from historical snapshots are not required in the base profile.

### 10.4 Rollback semantics

Rollback MUST create a new attributed revision. It MUST NOT rewrite prior history in place.

The reviewer client MUST choose rollback scope only from the selectors and action metadata returned by `GET /api/v1/records/{record_id}/history`. It MUST NOT infer legal rollback scope from visible labels, diff text, or storage-specific identifiers.

## 11. Clipboard paste and import workflows

### 11.1 Clipboard paste

Clipboard paste is base-profile functionality and MUST remain part of the base workbook interaction surface.

Clipboard paste alone does not satisfy the Import Extension Profile and MUST NOT be used to claim file-based structured import support.

Pasting TSV or CSV into the timeline sheet MUST create or update multiple rows starting from the selected cell.

Known columns MUST map directly.

Unknown columns MUST be preserved in `raw_capture.import_columns` or equivalent structured raw-capture storage.

Host and identity text from pasted cells MUST follow the same `entity_binding_mode` contract as interactive edits.

Successful non-conflicting writes from one paste action MUST appear as:

- one visible `change_set`,
- ordered mutation entries,
- one row revision per affected record.

Any same-field conflicts arising from that paste MUST remain outside that committed batch until explicitly resolved. Each later same-field conflict resolution MUST create its own attributed `change_set`.

### 11.2 Import Extension Profile

File-based structured import beyond clipboard paste belongs to the **Import Extension Profile**.

#### 11.2.1 Assistant boundary

If implemented, the file-based import path MUST be exposed as a **Phase 2 Workbook Import Assistant** for structured onboarding of CSV and XLSX sources.

Clipboard paste remains the base hot-path ingest surface. File-based workbook import remains an extension-profile capability.

The assistant MUST be the only application surface that knows about workbook discovery, sheet heuristics, Excel tables, named ranges, parser quirks, previewing, and workbook downgrade warnings.

The file-based import path MUST be isolated behind the dedicated `imports` module defined by Core 01.

The file-based import path MUST use the same stable tabular-ingest contract and shared mapping engine as clipboard-driven structured ingest, including the same `entity_binding_mode` contract.

Each candidate unit MUST support operator preview and header mapping before any apply step.

The assistant MUST prioritize mappings for timeline, systems or hosts, accounts or identities, indicators, evidence tracker, and VERIS-like summaries when present.

The import path MUST preserve unknown columns in `raw_capture` or `custom_attrs`.

The import path MUST determine `mention_origin` versus `entity_origin` from contracts rather than sheet labels alone.

The import path MUST avoid auto-resolution of imported host or account aliases during ingest.

#### 11.2.2 `import_session`

An `import_session` MUST be the durable, audit-visible coordinator for one uploaded source file and one operator-driven import workflow.

An `import_session` MUST persist, at minimum:

- `import_session_id`,
- `incident_id`,
- `created_by_user_id`,
- `created_at`,
- `source_file_kind` with closed values `csv` or `xlsx`,
- `original_filename`,
- `source_content_sha256`,
- `parser_profile_id`,
- `parser_version`,
- `assistant_profile`, fixed to `phase2_workbook_import_v1` for the current profile,
- `session_status`,
- `selected_unit_ids[]`,
- `blocking_diagnostics[]`,
- `nonblocking_warning_codes[]`.

`session_status` MUST use the closed vocabulary `created`, `discovered`, `mapped`, `ready_to_apply`, `applying`, `partially_applied`, `applied`, `failed`, and `canceled`.

`source_content_sha256` MUST be computed from the exact uploaded file bytes, not from a parsed or normalized representation.

`parser_version` MUST identify the stable behavior version of the import adapter profile. It MUST NOT be limited to a raw parser-library semantic version.

`nonblocking_warning_codes[]` MUST use the same closed vocabulary as `import_unit.warning_codes[]` and MAY aggregate source-level warnings that do not block preview or apply.

Discovery and preview MUST be read-only against incident state. Apply MUST execute as a background job and MUST NOT block ordinary workbook editing.

#### 11.2.3 `import_unit`

An `import_unit` MUST be the explicit unit of structured source material selected from an `import_session`.

An `import_unit` MUST persist, at minimum:

- `import_unit_id`,
- `import_session_id`,
- `locator_kind`,
- `locator`,
- `source_rect_a1`,
- `header_row_ref`,
- `inferred_row_count`,
- `inferred_column_count`,
- `warning_codes[]`,
- `mapping_fingerprint`,
- `unit_status`.

`locator_kind` MUST use the closed vocabulary `csv_file`, `xlsx_used_range`, `xlsx_table`, `xlsx_named_range`, and `xlsx_region`.

`unit_status` MUST use the closed vocabulary `discovered`, `selected`, `mapped`, `ready`, `applying`, `applied`, `skipped`, `rejected`, and `failed`.

The semantic identity of an `import_unit` MUST be the tuple `source_content_sha256 + canonical_locator + parser_version`. If the implementation stores one derived key for that identity, it MUST compute that key as the SHA-256 of the canonical JSON serialization of `{source_content_sha256, locator, parser_version}`.

The current profile MUST use the following canonical locator shapes:

- `csv_file`: one trivial single-unit locator bound to the uploaded file,
- `xlsx_used_range`: `{sheet_name, rect_a1}`,
- `xlsx_table`: `{sheet_name, table_name, rect_a1}`,
- `xlsx_named_range`: `{defined_name, sheet_name, rect_a1}`,
- `xlsx_region`: `{sheet_name, rect_a1}`.

The locator MUST be a canonical contract object. It MUST NOT rely on display labels or UI-only wording.

`source_rect_a1` MUST identify the deterministic rectangular extent of the imported unit. For `csv_file`, it MUST cover the full parsed rectangle of the file.

`header_row_ref` MUST be a deterministic 1-based row reference within `source_rect_a1`.

`mapping_fingerprint` MAY be absent while `unit_status` is `discovered` or `selected`. Once a mapping is approved, it MUST be persisted and used for duplicate-apply detection.

#### 11.2.4 Discovery and batch semantics

Whole-workbook import MUST mean an orchestrated batch of explicit `import_unit` objects discovered from one uploaded workbook. It MUST follow this flow:

1. upload one XLSX file and create one `import_session`,
2. discover candidate `import_unit` objects,
3. allow operator preview, mapping, selection, or skip per unit,
4. validate duplicate-apply and overlap constraints,
5. apply selected units in deterministic order,
6. record one session outcome and one apply outcome per unit.

Whole-workbook import MUST NOT mean preserving workbook object identity, formula behavior, chart behavior, pivot behavior, protection behavior, merged-cell layout behavior, or general XLSX semantics outside the assistant.

The current profile MUST support discovery of:

- parser-resolved non-empty used ranges,
- Excel tables using their current rectangular extent only,
- named ranges only when they are static, single-sheet, and single-rectangle,
- operator-selected rectangular regions from a sheet preview.

Dynamic named-range formulas, external references, and multi-area named ranges MUST be rejected as unsupported.

Table filters, sorts, styles, hidden states, and other presentation metadata MUST NOT change imported semantics.

A whole-workbook batch MUST remain unit-atomic rather than session-atomic. Each applied unit MUST yield its own `change_set`. The session MAY end in `partially_applied`.

Selected units in one batch MUST have disjoint source-cell coverage. Overlapping selected units MAY be previewed, but they MUST NOT be jointly applied until the operator reduces the selection to a disjoint set.

The default apply order for selected units MUST be deterministic: workbook sheet order, then top-left rectangle position, with operator-added regions ordered by explicit session sequence when their rectangles would otherwise compare equal.

#### 11.2.5 `mapping_fingerprint` and duplicate-apply detection

`mapping_fingerprint` MUST be the deterministic identity of the operator-approved header-to-field plan for one `import_unit`.

The fingerprint input MUST include, at minimum:

- `mapping_contract_version`,
- `target_view_schema_id`,
- `header_row_ref`,
- `data_start_row_ref`,
- `unknown_column_policy`,
- one ordered entry per source column containing `source_column_ordinal`, `source_header_text`, `field_key`, `entity_binding_mode`, and any declared `transform_id`, `transform_options`, or `empty_value_policy`.

`header_row_ref` and `data_start_row_ref` MUST use the same 1-based row coordinate system within `source_rect_a1`.

The implementation MUST serialize that mapping input canonically with lexicographically sorted object keys and source columns ordered by `source_column_ordinal`, then hash the result as lower-hex SHA-256.

Visible labels MUST NOT affect `mapping_fingerprint` except through the raw imported `source_header_text`.

The assistant MUST warn on re-applying the same `(import_unit_id, mapping_fingerprint, incident_id)` tuple and MUST default to blocking the apply until the operator explicitly chooses re-import.

#### 11.2.6 Closed warning vocabulary and workbook downgrade semantics

`warning_code[]` MUST use only the following closed vocabulary in the current profile:

- `formula_inert`,
- `formula_cached_value_missing`,
- `merged_layout_downgraded`,
- `comments_metadata_only`,
- `pivot_ignored`,
- `chart_ignored`,
- `external_links_ignored`,
- `sheet_protection_metadata_only`,
- `workbook_protection_metadata_only`,
- `filtered_or_hidden_state_ignored`.

Hard failures MUST be reported through normal job diagnostics rather than `warning_code[]`.

Formula cells and other spreadsheet active content MUST be treated as inert input. The implementation MUST NOT execute workbook formulas, macros or VBA, workbook automation, or external links as live logic during preview or apply.

Encrypted or password-protected files that cannot be parsed MUST be rejected before discovery.

If a stored cached value exists for a formula cell, preview and import MAY use that inert cached value and MUST emit `formula_inert`. If no cached value exists for a mapped formula cell, the unit MUST NOT enter `ready` until the operator excludes or remaps the affected column, and `formula_cached_value_missing` MUST be emitted.

Merged cells MUST use only the upper-left anchor value for deterministic rectangular import. Covered cells MUST otherwise be treated as empty, and the unit MUST emit `merged_layout_downgraded`.

Comments or notes MUST be preserved only as source metadata and MUST emit `comments_metadata_only` when encountered.

Pivot tables, charts, and external links MUST NOT be interpreted as executable or live workbook logic and MUST emit `pivot_ignored`, `chart_ignored`, or `external_links_ignored` when encountered.

Workbook or sheet protection that does not block parsing MUST be treated only as source metadata and MUST emit `workbook_protection_metadata_only` or `sheet_protection_metadata_only` as applicable.

Hidden rows, hidden columns, and active filters MUST NOT change imported semantics in the current profile and MUST emit `filtered_or_hidden_state_ignored` when encountered.

Unsupported workbook features MUST be downgraded with one or more declared `warning_code[]` values, preserved only as raw source metadata, or rejected as unsupported.

XLSX inputs MUST be staged and parsed as hostile archive content consistent with Core 04.

## 12. Auto-resolution policy

### 12.1 Allowed scope

The system MAY auto-resolve a typed host or identity token to an existing alias only in **interactive mention-capture flows**, and only when `auto_resolution_confidence = 100`.

### 12.2 Required confidence conditions

`auto_resolution_confidence = 100` applies only when all of the following are true:

- the edited cell determines the expected entity type and candidate matching is limited to that type within the same incident,
- the token matches exactly one existing alias after deterministic normalization limited to case-folding plus whitespace collapse,
- the raw token contains no explicit uncertainty marker such as `?`, `~`, `maybe`, or `prob`,
- the target record is not soft-deleted, merged, retired, or disabled,
- no competing candidate remains after normalization.

Anything below 100 MAY drive ranking or suggestions. It MUST NOT mutate resolution state without explicit analyst selection.

### 12.3 Required write effects

When auto-resolution occurs, the system MUST still insert an `entity_mentions` row for the typed token with:

- `resolution_status='resolved'`,
- `resolved_record_id` set to the chosen record.

The corresponding `record_links` row MUST use:

- `provenance='auto_match'`,
- `confidence=100`.

### 12.4 Allowed and forbidden workflows

Auto-resolution MAY occur only in:

- inline commit of a Timeline Hosts or Identities cell,
- interactive clipboard paste into those same relationship cells when the auto-resolutions belong to the same visible `change_set`.

Auto-resolution MUST NOT occur in:

- the inspector’s explicit resolve flow,
- Hosts or Identities alias-edit cells,
- merge or dedupe workflows,
- file-based import through the Import Extension Profile,
- background jobs,
- asynchronous enrichment or cleanup,
- any workflow that would create a new canonical host or identity or edit alias rows without explicit confirmation.

### 12.5 Disclosure and undo

The UI MUST NOT silently auto-resolve.

When auto-resolution occurs, the current sheet MUST show an immediate non-modal disclosure on the same surface that includes:

- the raw token,
- the canonical target,
- the matched alias text,
- direct `Undo`,
- direct `Review`.

For batch paste, the disclosure MUST also include the number of tokens auto-resolved in that visible change set.

The resolved chip or cell MUST remain inspectably marked as auto-resolved.

Row history MUST preserve:

- raw token,
- matched alias text,
- confidence,
- mutation source.

`Undo` from the immediate disclosure MUST:

- restore the raw unresolved token,
- remove the auto-created link,
- preserve focus,
- preserve scroll position.

After the immediate disclosure expires, the user MUST still be able to choose `Revert to unresolved` from the chip context or row history in no more than two actions.

That later correction MUST be a new attributed revision.

## 13. Keyboard and editing contract

### 13.1 Grid editing

Selecting a cell and typing MUST edit it immediately.

Enter MUST commit and move vertically. Tab MUST commit and move horizontally.

Relationship cells MUST accept raw typing and MUST NOT require picker-first interaction.

### 13.2 Required keyboard actions

The base profile MUST support:

- Arrow keys to move selection,
- Enter and Shift+Enter row navigation,
- `Ctrl+V` for paste,
- `Ctrl+K` for quick link or resolve on the current cell,
- `Space` to preview linked evidence for the selected row,
- `Alt+H` to open history for the selected row,
- `Esc` to close the inspector and return focus to the prior cell.

### 13.3 Bulk editing

The implementation MUST support:

- multi-cell paste,
- fill-down,
- multi-row tag assignment.

Bulk edits are mutation batches. They MUST NOT rely on hidden macro semantics.

## 14. Sorting, filtering, and grouping

### 14.1 Sort and filter behavior

Column-header sort and inline filter chips MUST apply without leaving the sheet.

Sorting and filtering MUST apply to underlying rows before grouping is computed.

### 14.2 Timeline grouping boundary

Timeline grouping is a presentation-only transform over the current filtered result set.

Grouping MUST NOT create, delete, or mutate:

- source records,
- projection rows,
- links,
- tags.

### 14.3 Allowed grouping keys

Timeline sheets MUST support `Group: None` plus exactly one active grouping key in the base profile.

The active grouping key MUST be stored as a stable contract value in `saved_views.query_json.group_by`, not as a visible label.

The allowed grouping keys are exactly:

- `timeline.occurred_day`,
- `timeline.recorded_day`,
- `timeline.capture_state`,
- `timeline.has_evidence`,
- `timeline.has_unresolved_mentions`.

The base-profile whitelist is frozen at these five keys for the current profile. No additional grouping key is permitted in the base profile.

`timeline.event_type` MUST NOT be exposed as a base-profile grouping key.

Grouping by Summary, Hosts, Identities, Tags, arbitrary custom columns, formulas, ad hoc expressions, or visible labels is out of scope for current conformance.

### 14.4 Grouping value rules

Grouping keys MUST be scalar contract-backed values.

Group order MUST be deterministic:

- `timeline.occurred_day` and `timeline.recorded_day` sort by bucket value descending, with null buckets last,
- `timeline.capture_state` sorts `rough`, `enriched`, `reviewed`, `superseded`,
- `timeline.has_evidence` and `timeline.has_unresolved_mentions` sort `true` then `false`,
- the current row sort applies unchanged within each group.

### 14.5 Group-header behavior

Grouped timeline sheets MUST expose exactly one derived group-header level.

Group headers are derived UI rows only. They:

- MUST NOT have a `record_id`,
- MUST NOT accept inline edits,
- MUST NOT accept paste targets,
- MUST NOT appear in exports,
- MUST NOT appear in revision history,
- MUST NOT become mutation targets.

The allowed group operations are:

- expand group,
- collapse group,
- expand all,
- collapse all,
- `Group: None`.

### 14.6 Edit movement across groups

A row MAY move between visible groups only when an edit changes the grouped field value.

Dragging a row between groups MUST NOT be a write path.

### 14.7 Collaborative state boundary

Transient expand and collapse state SHOULD remain client-local and MUST NOT be broadcast as collaborative state.

Saved views MAY persist the default grouping key in `query_json.group_by`. They MUST NOT persist another user’s live open or closed state.

### 14.8 Grouping non-goals

The following are non-goals in the base profile:

- manual row-range grouping or ungrouping,
- nested outline depth greater than 1,
- subtotal, summary, or spacer rows inserted into the grid,
- pivot-style aggregation inside the timeline sheet,
- merged cells,
- indent-based hierarchy,
- parent and child tree rows.

## 15. Timeline read and write contract

The Timeline sheet MUST read from `timeline_grid_projection` or an equivalent projection.

When exposed over the public HTTP surface, Timeline reads MUST use the view-shaped query route `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`. New Timeline rows MUST use `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`. Updates to existing Timeline rows MUST use `PATCH /api/v1/records/{record_id}` with `view_schema_id`, `base_row_version`, `client_txn_id`, and `changes[]` keyed by `field_key`. Group headers remain client-local presentation state and MUST NOT appear as writable API rows.

The implementation MUST preserve the following write-back semantics:

| Timeline field or column | Read model | Required write-back behavior |
| --- | --- | --- |
| Time | `occurred_at` | update `timeline_events.occurred_at` |
| Summary | `summary` | update `timeline_events.summary` only |
| Details | `details` | update `timeline_events.details` only |
| Source Text | `source_text` | update `timeline_events.source_text` only |
| Hosts | host labels plus unresolved host tokens | if a unique exact normalized alias match qualifies for auto-resolution, insert resolved `entity_mentions` plus `record_links` with `provenance='auto_match'` and `confidence=100`; otherwise insert unresolved `entity_mentions` |
| Identities | identity labels plus unresolved identity tokens | if a unique exact normalized alias match qualifies for auto-resolution, insert resolved `entity_mentions` plus `record_links` with `provenance='auto_match'` and `confidence=100`; otherwise insert unresolved `entity_mentions` |
| Evidence | `evidence_count` | create `object_blob`, `evidence_record`, and `record_link` |
| Tags | `tag_names` | upsert tags and record-tag bindings |


For the lifecycle machine in §6, the current Timeline write surfaces and row-anchored mutation types are classified as follows:

- `capture-state-material`: `timeline.occurred_at`, `timeline.summary`, `timeline.details`, `timeline.source_text`, `timeline.host_refs`, `timeline.identity_refs`, Timeline-row evidence attach or detach, and row-anchored source-bound indicator observation create, link, dismiss, or equivalent typed-link mutation initiated from the row or its inspector.
- not `capture-state-material`: `timeline.tags`, explicit `mark-reviewed` or `supersede` lifecycle actions, selection, focus, presence, sort, filter, grouping, projection rebuild, and idempotent no-op retry.

Any future Timeline writable `field_key`, dedicated Timeline action route, or row-anchored mutation surface MUST declare whether it is `capture-state-material` before it can claim base-profile conformance.

Core 01 §7.4.1 fixes the authoritative Timeline `view_schema_id`, ordered field set, hidden technical fields, default sort tuple, filter whitelist, and grouping whitelist for the base profile.

Reads MAY be denormalized. Write-back MUST remain intent-aware and contract-driven.

Summary, details, source text, and other contract-declared source fields remain raw source fields. When the implementation supports inline indicator capture from such a field, write-back MUST preserve the raw cell text and create or update source-bound `indicator_observation` rows separately. It MUST NOT require dedicated IOC columns or rewrite the raw field to a canonical indicator label.

## 16. Entity, evidence, and inspector surfaces

### 16.1 Entity and evidence sheets

Hosts, Identities, and Evidence sheets MUST remain workbook-shaped peer sheets backed by their own projections.

Canonical fields such as `display_name`, `hostname`, `upn`, host `location`, `os_platform`, `business_owner`, `criticality`, `containment_status`, identity `privilege_level`, `mfa_state`, `reset_status`, and evidence `title`, `requested_at`, `received_at`, `storage_ref`, `collector_party_text`, and `source_party_text` MUST be inline-editable where appropriate.

Alias cells MUST behave like chip editors.

Relationship-derived or machine-derived columns such as linked-event counts, evidence counts, blob-upload state, preview handles, and last-updated markers MUST be read-only and navigable.

Entity normalization-state fields equivalent to `stub` or `canonical` MUST be projection-backed and read-only in-grid in the base profile. A normalization transition such as `stub -> canonical` MAY be exposed through the inspector or another explicit flow, but it MUST NOT be an ordinary cell edit on the Hosts or Identities sheet.

### 16.2 Inspector

The inspector MUST:

- remain a non-blocking drawer rather than the default primary capture surface,
- expose details, relationships, evidence, and history views,
- keep the main grid visible while open,
- support mention resolution, dismissal, and restore, indicator observation linking, entity and indicator creation, indicator lifecycle inspection or editing, evidence inspection, and rollback,
- preserve the distinction between raw mention lineage and current canonical links,
- preserve the distinction between raw indicator observation lineage and current canonical indicator links and lifecycle windows.

The inspector is the enrichment surface. It MUST NOT be required for the common path of timeline row creation and editing.

For authorized users, the inspector MUST support explicit host and identity merge initiation. Merge MUST NOT be a bulk grid action in the base profile. The merge UI MUST require explicit survivor and loser selection and a final destructive-action confirmation that identifies both `record_id` values before submission.

### 16.3 Compromise-assessment surfaces

Interactive assessment-entry surfaces MUST keep `assessment_state` separate from operational-response actions such as containment, isolation, disablement, credential reset, or monitoring.

Interactive assessment-entry surfaces MUST expose confidence by default as `unset`, `low`, `medium`, or `high`. They MAY additionally expose an exact integer `confidence_score` in the range `0..100` as a secondary control.

When the band-first path is used, the implementation MUST persist canonical default scores of `25` for `low`, `55` for `medium`, and `85` for `high`. `unset` MUST persist `confidence_score=NULL`. A band-first control MUST write `confidence_score`; `confidence_band` remains derived.

Workbook filtering on compromise-assessment surfaces MUST treat `assessment_state` and `confidence_band` as separate fields.

The base-profile Assessments view is append-only for semantic assessment fields. Creating a row or equivalent entry MUST append a new assessment record. In-place grid edits to an existing assessment row MUST NOT overwrite `subject_ref`, `subject_type`, `assessment_state`, `confidence_score`, `rationale`, `assessor`, `assessed_at`, or supporting-link semantics; correction or superseding flows MUST append a new assessment record instead.

### 16.4 Analyst-work coordination surfaces

Task requests and decisions MUST remain workbook surfaces rather than separate application modules. The implementation MAY surface them as pinned system views, saved views, or equivalent workbook-native queues.

From a selected Timeline, Host, Identity, Evidence, Notes, Task Requests, or Decisions row, the analyst MUST be able to create or link a `task_request`, `decision`, or structured coordination artifact without leaving the workbook flow.

Routine timeline row creation and editing MUST NOT require task, decision, owner, approver, challenge, or checklist fields on the timeline sheet itself.

Communications logs, handoffs, status reviews, and lesson artifacts SHOULD be created at milestone, shift-change, or reporting boundaries rather than on every row edit. The workbook surface MAY assist creation from the inspector or a coordination view, but it MUST NOT interrupt ordinary grid editing.

Saved or system views over task requests, decisions, and coordination artifacts MUST support owner queues, blocked-work views, due or next-checkpoint views, workstream views, requester or audience views where applicable, external-ticket lookup, and no-owner gap detection where applicable.

Promoting recurrent coordination or evidence-management fields MUST NOT add them as mandatory Timeline columns or move them into inspector-only JSON payloads.

## 17. Authorship and attribution in the UI

The workbook MUST surface authorship with low friction.

At minimum:

- the row must expose last editor and relative update time,
- history must be reachable in one click or one shortcut,
- the system MUST preserve end-to-end attribution across current records, change sets, revisions, links, tags, blob metadata, and evidence metadata.

## 18. Excel-like feel versus intentional differences

### 18.1 Excel-like behaviors

The implementation MUST preserve the following spreadsheet-like behaviors:

- tabular grid,
- direct typing,
- paste,
- fill-down,
- keyboard navigation,
- flexible sorting and filtering.

### 18.2 Intentional differences

The implementation MUST intentionally differ from a raw spreadsheet in the following ways:

- relationship cells render canonical or unresolved chips rather than remaining raw delimited strings forever,
- evidence is attached object state rather than path text inside a cell,
- history and rollback are built in,
- formulas, macros, and merged cells are not part of the model.

## 19. Interaction invariants

A conformant implementation MUST preserve all of the following:

1. the primary capture path remains in the grid,
2. same-field concurrent edits never silently overwrite,
3. low-friction row creation survives incomplete information,
4. destructive operations stay explicit and attributed,
5. grouping remains presentation-only,
6. grouped surfaces still mutate underlying rows by `record_id` and `row_version`,
7. auto-resolution stays tightly bounded and reversible,
8. the inspector enriches work but does not replace the grid,
9. unresolved same-field conflict drafts remain client-local until explicit resolution,
10. analyst-work coordination remains workbook-native without adding required coordination fields to the routine timeline hot path.
